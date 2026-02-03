# MCP 工具调用可见性功能设计

## 概述

在调试和使用 CNB MCP 助手时，需要清晰地了解何时调用了 MCP 工具以及调用的详细信息。本设计为所有 MCP 工具调用添加自动的调用前提示和调用后摘要。

## 目标

1. **透明性**：用户清楚地知道何时通过 MCP 获取数据
2. **调试性**：显示工具名称、参数和执行时间，便于调试
3. **通用性**：作为通用 skill，可以在其他项目中重用
4. **自动化**：无需依赖 LLM，在代码层面强制显示

## 需求

### 功能需求

当系统调用 CNB MCP 工具时（通过 `cnb-mcp.py call` 命令），需要：

1. **调用前输出**：
   - 显示正在调用的工具名称
   - 显示传入的参数（JSON 格式）

2. **调用后输出**：
   - 显示数据来源（工具名称）
   - 显示完整参数
   - 显示执行耗时

3. **非侵入性**：
   - 不影响工具返回的实际数据
   - 不干扰 LLM 解析工具结果

### 非功能需求

1. **性能**：添加的输出格式化开销可忽略不计
2. **可维护性**：代码清晰，易于理解和修改
3. **健壮性**：解析失败时优雅降级，不影响原有功能

## 设计方案

### 架构决策

**实现方式**：代码层自动添加（选项 1）

**理由**：
- 保证每次 MCP 调用都显示信息
- 不依赖 LLM 的判断，更可靠
- 便于调试和通用化

**实现位置**：`internal/cli/executor.go`

### 输出格式

#### 调用前格式

```
📡 正在调用 MCP 工具：<tool_name>
   参数：<JSON formatted arguments>
```

示例：
```
📡 正在调用 MCP 工具：list_organizations
   参数：{}

📡 正在调用 MCP 工具：get_repository
   参数：{
     "repo": "demo-app"
   }
```

#### 调用后格式

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ️ 数据来源：<tool_name>
   参数：<JSON formatted arguments>
   耗时：<duration>s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

示例：
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ️ 数据来源：list_organizations
   参数：{}
   耗时：1.45s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 核心组件

#### 1. 数据结构

```go
// MCPToolInfo 存储 MCP 工具调用的信息
type MCPToolInfo struct {
    ToolName   string                 // 工具名称，如 "list_organizations"
    Arguments  map[string]interface{} // 参数键值对
    StartTime  time.Time              // 开始时间
    EndTime    time.Time              // 结束时间
}

// Duration 返回执行耗时
func (m *MCPToolInfo) Duration() time.Duration {
    return m.EndTime.Sub(m.StartTime)
}
```

#### 2. 命令解析

```go
// parseMCPCommand 解析 MCP 工具调用命令
// 返回：(工具名, 参数map, 是否为MCP调用)
func parseMCPCommand(command string) (string, map[string]interface{}, bool)
```

**解析逻辑**：

1. 检查命令是否包含 `cnb-mcp.py call`
2. 使用正则表达式或字符串分割提取：
   - 工具名称：`call` 后的第一个参数
   - 参数：后续所有 `key=value` 格式的参数
3. 解析参数值：
   - 尝试作为 JSON 解析（处理引号字符串、数字、布尔值）
   - 失败则作为普通字符串

**命令格式示例**：
```bash
python3 skills/scripts/cnb-mcp.py call list_organizations
python3 skills/scripts/cnb-mcp.py call get_repository repo=demo-app
python3 skills/scripts/cnb-mcp.py call query_knowledge query="CI/CD配置"
```

#### 3. 格式化输出

```go
// formatMCPCallStart 格式化工具调用开始的输出
func formatMCPCallStart(toolName string, args map[string]interface{}) string

// formatMCPCallEnd 格式化工具调用结束的输出
func formatMCPCallEnd(info MCPToolInfo) string
```

**实现细节**：
- 使用 `json.MarshalIndent` 格式化参数
- 空参数显示为 `{}`
- 时间格式化为秒，保留 2 位小数

#### 4. ExecuteTool 修改

```go
func (a *Assistant) ExecuteTool(toolName string, argumentsJSON string) (string, error) {
    switch toolName {
    case "execute_bash":
        var args struct {
            Command string `json:"command"`
        }
        if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
            return "", fmt.Errorf("failed to parse arguments: %w", err)
        }

        // 解析 MCP 命令
        mcpToolName, mcpArgs, isMCP := parseMCPCommand(args.Command)

        if isMCP {
            // 输出调用前信息
            fmt.Println(formatMCPCallStart(mcpToolName, mcpArgs))

            // 记录开始时间并执行
            startTime := time.Now()
            result, err := executeBashCommand(args.Command)

            // 输出调用后信息
            info := MCPToolInfo{
                ToolName:  mcpToolName,
                Arguments: mcpArgs,
                StartTime: startTime,
                EndTime:   time.Now(),
            }
            fmt.Println(formatMCPCallEnd(info))

            return result, err
        }

        // 非 MCP 命令，正常执行
        return executeBashCommand(args.Command)
    default:
        return "", fmt.Errorf("unknown tool: %s", toolName)
    }
}
```

### 错误处理

#### 1. 解析失败

如果无法解析 MCP 命令格式：
- 不添加额外输出
- 正常执行命令
- 避免干扰原有流程

#### 2. 参数解析失败

如果参数格式不规范：
- 显示原始字符串
- 例如：`参数：无法解析 (原始命令片段)`

#### 3. 工具执行失败

即使工具调用失败：
- 仍然显示完整的调用前后信息
- 让用户知道调用了什么工具及其参数
- 便于诊断问题

#### 4. 输出位置

- 所有额外信息输出到 stdout
- 确保不干扰 LLM 接收工具返回的 JSON 结果

### 边缘情况

#### 1. 多行命令

```bash
python3 skills/scripts/cnb-mcp.py call query_knowledge \
  query="如何配置"
```

**处理**：使用 `strings.ReplaceAll` 移除换行符和多余空格后再解析

#### 2. 参数包含特殊字符

```bash
python3 skills/scripts/cnb-mcp.py call query_knowledge query="CI/CD 配置"
```

**处理**：正确识别和保留引号内的内容

#### 3. 无参数调用

```bash
python3 skills/scripts/cnb-mcp.py call list_organizations
```

**处理**：显示 `参数：{}`

#### 4. 非 call 命令

```bash
python3 skills/scripts/cnb-mcp.py list-tools
```

**处理**：不添加 MCP 调用信息（只对 `call` 命令生效）

## 实现计划

### 阶段 1：核心解析功能

1. 实现 `MCPToolInfo` 结构
2. 实现 `parseMCPCommand` 函数
3. 编写单元测试验证解析逻辑

### 阶段 2：格式化输出

1. 实现 `formatMCPCallStart` 函数
2. 实现 `formatMCPCallEnd` 函数
3. 测试各种参数格式的输出

### 阶段 3：集成到 ExecuteTool

1. 修改 `ExecuteTool` 函数
2. 添加时间记录
3. 确保不影响原有功能

### 阶段 4：测试和验证

1. 测试成功场景
2. 测试失败场景
3. 测试边缘情况
4. 验证输出格式

## 预期输出示例

### 成功场景

```
CNB Assistant> 帮我列出所有的组织

📡 正在调用 MCP 工具：list_organizations
   参数：{}

找到 10 个组织，以下是详细信息：

## 📋 你的组织列表

### 1. **k_k**
- **路径**: `k_k`
- **权限**: Owner
- **仓库**: 2个
- **成员**: 1人
- **描述**: 暂无描述
- **创建时间**: 2024-11-13

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ️ 数据来源：list_organizations
   参数：{}
   耗时：1.45s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 带参数的调用

```
CNB Assistant> 查看 demo-app 仓库详情

📡 正在调用 MCP 工具：get_repository
   参数：{
     "repo": "demo-app"
   }

仓库详情：
...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ️ 数据来源：get_repository
   参数：{
     "repo": "demo-app"
   }
   耗时：0.82s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 失败场景

```
CNB Assistant> 查看不存在的仓库

📡 正在调用 MCP 工具：get_repository
   参数：{
     "repo": "non-existent"
   }

错误：仓库未找到

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ️ 数据来源：get_repository
   参数：{
     "repo": "non-existent"
   }
   耗时：0.65s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## 影响评估

### 对现有功能的影响

- **LLM 交互**：不影响，工具结果照常返回
- **性能**：可忽略不计（仅增加字符串格式化和输出）
- **错误处理**：更容易调试，能看到确切的调用参数

### 可扩展性

这个设计可以轻松扩展到：
- 其他 MCP 服务器
- 其他类型的工具调用
- 添加更多调试信息（如 token 使用量等）

## 测试策略

### 单元测试

- `parseMCPCommand` 的各种输入格式
- 参数解析边缘情况
- 格式化输出的正确性

### 集成测试

- 完整的工具调用流程
- 与 LLM 交互的兼容性
- 错误场景的处理

### 手动测试

- 实际运行助手并触发各种 MCP 调用
- 验证输出格式和可读性
- 确认不影响用户体验

## 未来改进

### 可配置性

添加配置选项控制输出详细程度：
- `verbose`: 完整信息（默认）
- `simple`: 仅工具名称
- `off`: 关闭额外输出

### 日志记录

将调用信息同时写入日志文件：
- 便于事后分析
- 支持调试和审计

### 统计信息

累计统计：
- 总调用次数
- 平均响应时间
- 失败率

## 结论

此设计通过在代码层自动添加 MCP 工具调用信息，提供了透明、可调试的工具调用体验。实现简单、健壮，且不影响现有功能，适合作为通用 skill 在其他项目中使用。
