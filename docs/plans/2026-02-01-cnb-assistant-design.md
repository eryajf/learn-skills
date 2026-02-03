# CNB Assistant 设计文档

**创建日期**: 2026-02-01
**目标**: 创建一个基于 Go 语言的 CNB Skill Demo，用于学习 Skill 机制和 CNB 平台操作

## 项目概述

CNB Assistant 是一个命令行 AI 助手，帮助用户通过自然语言与 CNB 平台交互。项目包含两个核心部分：
1. **CNB Skill 文件** - 定义 CNB 平台的能力和使用方法
2. **Go CLI 程序** - 提供命令行交互界面，集成 LLM 和 MCP

### 核心功能

- **知识库查询**: 调用 CNB 知识库 API，实现 RAG 检索
- **仓库操作**: 通过 MCP 管理代码仓库（查看仓库、分支、提交等）
- **流水线管理**: 查看和触发 CI/CD 流水线，查询构建状态

### 技术特点

- 支持 OpenAI 格式 API，兼容国内大模型（DeepSeek、通义千问、智谱等）
- 使用 CNB 官方 MCP Server (https://mcp.cnb.cool/sse)
- 支持交互式对话和单次命令两种运行模式
- 配置文件 + 环境变量的灵活配置方式

## 架构设计

### 整体架构

```
用户输入
   ↓
CLI 解析（交互式/单次命令）
   ↓
加载 CNB Skill 到 System Prompt
   ↓
调用 LLM API（OpenAI 格式）
   ↓
LLM 理解任务并调用 MCP 工具
   ↓
CNB 平台执行操作
   ↓
返回结果给用户
```

### 目录结构

```
cnb-assistant/
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖校验
├── config.yaml.example     # 配置文件示例
├── config.yaml             # 实际配置（gitignore）
├── README.md               # 项目说明文档
│
├── skills/
│   └── cnb-skill.md        # CNB Skill 定义
│
├── internal/
│   ├── config/
│   │   └── config.go       # 配置加载和管理
│   │
│   ├── llm/
│   │   ├── client.go       # LLM 客户端封装
│   │   └── types.go        # 请求/响应类型定义
│   │
│   ├── mcp/
│   │   ├── client.go       # MCP 客户端
│   │   ├── tools.go        # MCP 工具转 Function 定义
│   │   └── cnb.go          # CNB MCP 具体实现
│   │
│   └── cli/
│       ├── interactive.go  # 交互式模式
│       └── oneshot.go      # 单次命令模式
│
└── .gitignore
```

## 核心模块设计

### 1. CNB Skill 文件 (skills/cnb-skill.md)

**设计原则**:
- 使用通用自然语言，兼容各种大模型
- 清晰描述 CNB 平台的能力
- 提供具体的工作流指导
- 包含三个主要功能模块的使用说明

**内容结构**:
```markdown
# CNB Skill

## CNB 平台简介
[CNB 是什么、主要功能]

## 使用前提
[需要的配置、Token、MCP Server]

## 功能模块

### 1. 知识库查询
[如何调用知识库 API、RAG 流程]

### 2. 仓库操作
[如何通过 MCP 操作仓库、可用的工具]

### 3. 流水线管理
[如何查看和触发流水线、状态查询]

## 工作流程
[处理用户请求的步骤]
```

### 2. 配置管理 (internal/config)

**配置项**:
```yaml
llm:
  api_key: "sk-..."          # LLM API Key
  base_url: "https://..."    # API Base URL
  model: "deepseek-chat"     # 模型名称

cnb:
  token: "your-token"        # CNB 访问令牌
  mcp_url: "https://mcp.cnb.cool/sse"
  api_base: "https://api.cnb.cool"
```

**实现要点**:
- 使用 viper 库处理配置
- 支持配置文件（config.yaml）
- 支持环境变量覆盖（OPENAI_API_KEY、CNB_TOKEN 等）
- 环境变量优先级高于配置文件

### 3. MCP 客户端 (internal/mcp)

**职责**:
- 连接 CNB MCP Server (SSE 方式)
- 获取可用的 MCP 工具列表
- 将 MCP 工具转换为 LLM Function Calling 格式
- 执行 MCP 工具调用，返回结果

**关键接口**:
```go
type MCPClient interface {
    Connect() error
    ListTools() ([]Tool, error)
    ExecuteTool(name string, args map[string]interface{}) (interface{}, error)
    Close() error
}
```

**CNB MCP 工具示例**:
- `cnb_list_repos` - 列出仓库
- `cnb_get_repo_info` - 获取仓库详情
- `cnb_list_branches` - 列出分支
- `cnb_trigger_pipeline` - 触发流水线
- `cnb_get_pipeline_status` - 获取流水线状态
- `cnb_query_knowledge` - 查询知识库

### 4. LLM 客户端 (internal/llm)

**职责**:
- 封装 OpenAI 格式的 API 调用
- 支持 Function Calling / Tool Use
- 处理流式响应（可选）
- 管理对话上下文

**关键类型**:
```go
type Message struct {
    Role    string      `json:"role"`
    Content string      `json:"content"`
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
    Model    string     `json:"model"`
    Messages []Message  `json:"messages"`
    Tools    []Tool     `json:"tools,omitempty"`
}

type ChatResponse struct {
    Choices []Choice `json:"choices"`
}
```

**实现要点**:
- 在 System Prompt 中加载 CNB Skill 内容
- 将 MCP 工具作为 Functions 传递给 LLM
- 处理 LLM 返回的 tool_calls，调用 MCP 执行
- 将执行结果返回给 LLM 继续生成

### 5. CLI 交互 (internal/cli)

**交互式模式** (interactive.go):
- 启动后进入对话循环
- 显示提示符 `CNB Assistant> `
- 支持多轮对话，维护上下文
- 特殊命令：`exit`、`clear`、`help`

**单次命令模式** (oneshot.go):
- 接收命令行参数作为问题
- 执行一次对话
- 打印结果后退出

**实现要点**:
- 交互式模式使用 `bufio.Scanner` 读取输入
- 单次模式直接处理 `os.Args`
- 统一的对话处理逻辑，两种模式复用

### 6. 主程序 (main.go)

**流程**:
1. 解析命令行参数
2. 加载配置（文件 + 环境变量）
3. 初始化 MCP 客户端，连接 CNB
4. 初始化 LLM 客户端
5. 加载 CNB Skill 内容
6. 根据参数选择运行模式：
   - 无参数 → 交互式模式
   - 有参数 → 单次命令模式
7. 启动对应的 CLI 模式

## 数据流设计

### 用户请求处理流程

```
1. 用户输入: "查询我的仓库列表"
   ↓
2. CLI 接收输入
   ↓
3. 构建对话请求:
   - System: [CNB Skill 内容]
   - User: "查询我的仓库列表"
   - Tools: [MCP 工具列表]
   ↓
4. 调用 LLM API
   ↓
5. LLM 理解任务，决定调用 cnb_list_repos
   ↓
6. 返回 tool_calls: [{name: "cnb_list_repos", args: {...}}]
   ↓
7. 执行 MCP 工具调用
   ↓
8. 获取仓库列表数据
   ↓
9. 将结果作为 tool 消息追加到对话
   ↓
10. 再次调用 LLM 生成最终回复
    ↓
11. 返回格式化的仓库列表给用户
```

### 知识库查询流程

```
1. 用户: "CNB 如何配置自定义按钮？"
   ↓
2. LLM 识别为知识库查询
   ↓
3. 调用 cnb_query_knowledge
   - args: {query: "配置自定义按钮"}
   ↓
4. CNB 知识库返回相关文档片段
   ↓
5. LLM 基于检索结果生成回答
   ↓
6. 返回给用户
```

## 错误处理

### 配置错误
- 缺少必要配置项 → 友好提示并退出
- Token 无效 → 提示检查 CNB_TOKEN

### 网络错误
- MCP 连接失败 → 重试机制 + 错误提示
- LLM API 调用失败 → 显示错误信息

### MCP 工具调用错误
- 工具不存在 → LLM 会收到错误，自行调整
- 参数错误 → 返回错误信息给 LLM

## 技术栈

### Go 依赖
- `github.com/spf13/viper` - 配置管理
- `github.com/spf13/cobra` (可选) - CLI 框架
- HTTP 客户端 - 标准库 `net/http`
- JSON 处理 - 标准库 `encoding/json`

### MCP 协议
- 传输方式: SSE (Server-Sent Events)
- 认证: Bearer Token in Headers

### LLM API
- 格式: OpenAI Compatible
- 功能: Chat Completions + Function Calling

## 示例使用场景

### 场景 1: 查询仓库
```bash
$ go run main.go "列出我的所有仓库"
正在查询您的 CNB 仓库...

找到 5 个仓库：
1. cnb/demo-app (主分支: main, 最后更新: 2 天前)
2. cnb/backend-service (主分支: master, 最后更新: 1 周前)
...
```

### 场景 2: 触发流水线
```bash
$ go run main.go
CNB Assistant> 触发 demo-app 的主分支构建
正在触发流水线...
✓ 流水线已触发
  - 构建 ID: #123
  - 状态: running
  - 查看详情: https://cnb.cool/cnb/demo-app/builds/123

CNB Assistant>
```

### 场景 3: 知识库查询
```bash
$ go run main.go "CNB 如何配置 webhook？"
根据 CNB 文档，配置 webhook 的步骤如下：

1. 进入项目设置页面
2. 选择"集成与插件" → "Webhook"
3. 点击"添加 Webhook"
4. 填写 URL 和触发事件
...

参考文档: https://docs.cnb.cool/zh/guide/webhook
```

## 扩展性考虑

### 未来可扩展功能
- 支持更多 CNB 功能（Issue 管理、制品查询等）
- 添加本地缓存，提高响应速度
- 支持多个 CNB 账户切换
- 集成更多 MCP Server
- Web UI 界面

### 代码扩展点
- `internal/mcp/` - 添加新的 MCP Server 支持
- `skills/` - 添加更多 Skill 文件
- `internal/cli/` - 添加新的交互模式

## 学习目标

通过这个项目，可以学习：

1. **Skill 机制**
   - Skill 文件的结构和编写方法
   - 如何通过自然语言指导 AI 行为
   - Skill 与 MCP 的协作方式

2. **MCP 协议**
   - MCP Server 的连接和认证
   - 工具定义和调用
   - MCP 与 LLM Function Calling 的转换

3. **LLM 集成**
   - OpenAI API 的使用
   - Function Calling 机制
   - 多轮对话管理

4. **Go 工程实践**
   - 项目结构组织
   - 配置管理
   - CLI 程序开发

## 实施计划

### Phase 1: 基础框架
1. 创建项目结构
2. 实现配置管理
3. 实现基础 CLI（不含 MCP）

### Phase 2: MCP 集成
1. 实现 MCP 客户端
2. 连接 CNB MCP Server
3. 工具列表获取和转换

### Phase 3: LLM 集成
1. 实现 LLM 客户端
2. 集成 Function Calling
3. 实现完整对话流程

### Phase 4: Skill 开发
1. 编写 CNB Skill 文件
2. 测试各个功能模块
3. 优化提示词

### Phase 5: 完善和文档
1. 错误处理完善
2. 编写 README
3. 添加使用示例

## 总结

CNB Assistant 是一个实用的学习项目，通过实现一个完整的 AI Agent，深入理解 Skill、MCP 和 LLM 的协作机制。项目设计考虑了可扩展性和实用性，既可以作为学习材料，也可以在实际工作中使用。
