# Learn Skills - CNB 助手

我的skill学习之旅的示例demo，一个用于与 CNB（Cloud Native Build，云原生构建）平台交互的命令行 AI 助手。

**项目地址**: [https://cnb.cool/znb/learn-skills](https://cnb.cool/znb/learn-skills)


## 特性

- 🤖 **智能辅助**：使用自然语言与 CNB 平台交互
- 🔧 **MCP 集成**：通过 CNB 官方 MCP 服务器进行平台操作
- 📚 **知识库**：通过 RAG 查询 CNB 文档
- 🗂️ **仓库管理**：列出、查看和管理代码仓库
- 🚀 **流水线操作**：触发构建并检查流水线状态
- 🌐 **模型无关**：支持任何 OpenAI 兼容的 API（DeepSeek、通义千问、智谱 GLM 等）
- 💬 **双模式**：支持交互式聊天或单次命令执行

### 组件

- **Skills**（[skills/cnb-skill/SKILL.md](skills/cnb-skill/SKILL.md)）：为 LLM 提供自然语言指导，说明如何使用工具
- **CNB MCP 脚本**（[skills/cnb-skill/scripts/cnb-mcp.py](skills/cnb-skill/scripts/cnb-mcp.py)）：Python 脚本，调用 CNB 官方 MCP HTTP API
- **工具定义**（[internal/cli/tools.go](internal/cli/tools.go)）：定义 execute_bash 工具供 LLM 调用
- **工具执行器**（[internal/cli/executor.go](internal/cli/executor.go)）：执行工具调用并返回结果
- **LLM 客户端**（[internal/llm](internal/llm)）：OpenAI 兼容 API 包装器，支持 function calling
- **CLI**（[internal/cli](internal/cli)）：交互式和单次命令模式，实现工具调用循环
- **配置**（[internal/config](internal/config)）：配置管理

### 设计理念

这个项目采用了简洁的架构设计：

1. **Go 程序负责 LLM 调用和 CLI** - 保持代码简单
2. **通过 Skill 指导 LLM** - 使用自然语言描述如何操作
3. **Function Calling 机制** - LLM 通过 OpenAI function calling 调用工具
4. **Python 脚本调用 CNB MCP** - 通过 HTTP API 访问 CNB 平台
5. **完整的工具循环** - 自动执行工具并将结果返回给 LLM


## 快速开始

### 前置要求

- Go 1.25 或更高版本
- Python 3.6 或更高版本
- curl 命令（用于调用 CNB MCP HTTP API）
- CNB 账户和 API Token
- OpenAI 兼容的 LLM API 访问权限（OpenAI、DeepSeek、通义千问等）

### 从源码安装

```bash
git clone https://cnb.cool/znb/learn-skills.git
cd learn-skills
go build -o learn-skills
```

安装完成后，将二进制文件移动到 PATH 中（可选）：

```bash
# macOS/Linux
sudo mv learn-skills /usr/local/bin/

# 或者添加到用户目录
mv learn-skills ~/.local/bin/
```

## 配置

Learn Skills 支持两种配置方式：配置文件和环境变量。

### 方式 1：配置文件（推荐）

在项目根目录创建 `config.yaml`：

```yaml
llm:
  api_key: "sk-..."                       # LLM API 密钥
  base_url: "https://api.openai.com/v1"  # API 端点
  model: "gpt-4"                          # 模型名称

cnb:
  token: "your-cnb-token"                 # CNB 访问令牌
```

### 方式 2：环境变量

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_BASE_URL="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-4"
export CNB_TOKEN="your-cnb-token"
```

### 常见 LLM 提供商配置示例

<details>
<summary>OpenAI</summary>

```yaml
llm:
  api_key: "sk-..."
  base_url: "https://api.openai.com/v1"
  model: "gpt-4"
```
</details>

<details>
<summary>DeepSeek</summary>

```yaml
llm:
  api_key: "sk-..."
  base_url: "https://api.deepseek.com/v1"
  model: "deepseek-chat"
```
</details>

<details>
<summary>阿里通义千问</summary>

```yaml
llm:
  api_key: "sk-..."
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
  model: "qwen-plus"
```
</details>

<details>
<summary>智谱 AI</summary>

```yaml
llm:
  api_key: "..."
  base_url: "https://open.bigmodel.cn/api/paas/v4"
  model: "glm-4.7"
```
</details>

### 获取 CNB Token

1. 访问 [CNB 平台](https://cnb.cool)
2. 进入 设置 → 访问令牌
3. 创建新令牌并赋予所需权限：
   - `repo-code:r`（读取仓库代码）
   - 根据需要添加其他权限

### 测试 CNB MCP 脚本

在运行助手之前，你可以先测试 CNB MCP 脚本是否正常工作：

```bash
# 设置 CNB Token
export CNB_TOKEN="your-cnb-token"

# 列出可用工具
python3 skills/cnb-skill/scripts/cnb-mcp.py list-tools

# 测试调用（示例）
python3 skills/cnb-skill/scripts/cnb-mcp.py call query_knowledge query="CI/CD"
```

## 使用方法

### 交互式模式

启动对话会话：

```bash
./learn-skills
```

示例交互：
```
Learn Skills> 列出我的仓库
找到 3 个仓库：
1. demo/web-app（main，2 天前更新）
2. demo/backend（master，1 周前更新）
3. demo/mobile（main，3 周前更新）

Learn Skills> 如何配置 webhook？
根据 CNB 文档：

要在 CNB 中配置 webhook：
1. 进入项目设置页面
2. 选择"集成与插件" → "Webhook"
3. 点击"添加 Webhook"
4. 输入 webhook URL 并选择触发事件
...

参考文档：https://docs.cnb.cool/zh/guide/webhook

Learn Skills> exit
再见！
```

### 单次命令模式

执行单个命令：

```bash
./learn-skills "列出我的仓库"
./learn-skills "触发 demo-app 主分支的构建"
./learn-skills "如何设置 CI/CD 流水线？"
```

### 特殊命令（交互模式）

- `exit` 或 `quit` - 退出助手
- `clear` - 清除对话历史
- `help` - 显示帮助信息

## 示例查询

### 仓库操作
- "列出我的所有仓库"
- "显示 demo-app 仓库中的分支"
- "获取 backend-service 仓库的详细信息"

### 流水线管理
- "触发 demo-app 主分支的构建"
- "显示构建 #123 的状态"
- "列出最近的构建记录"

### 知识库
- "如何配置 webhook？"
- "CNB 的 CI/CD 功能有哪些？"
- "如何设置远程开发工作空间？"

### 工作流程

```
用户输入
   ↓
CLI 解析（交互式/单次命令）
   ↓
加载 CNB Skill → 系统提示词
   ↓
LLM API 调用（附带 execute_bash 工具定义）
   ↓
LLM 决定调用 execute_bash 工具
   ↓
Go 程序执行 Bash 命令 (skills/cnb-skill/scripts/cnb-mcp.py)
   ↓
Python 脚本调用 CNB MCP HTTP API → CNB 平台
   ↓
返回结果给 Go 程序
   ↓
Go 程序将结果返回给 LLM
   ↓
LLM 生成最终响应
   ↓
显示给用户
```

> 💡 **详细架构说明**：想了解完整的架构设计和实现细节，请阅读 [SKILL_ARCHITECTURE.md](docs/SKILL_ARCHITECTURE.md)


## 开发指南

### 项目结构

```
learn-skills/
├── main.go                               # 入口点
├── go.mod                                # Go 模块
├── config.yaml.example                   # 配置示例文件
├── skills/
│   └── cnb-skill/                        # CNB Skill 定义
│       ├── SKILL.md                      # Skill 描述文档
│       └── scripts/
│           └── cnb-mcp.py                # MCP 客户端脚本
├── internal/
│   ├── config/                           # 配置管理
│   ├── llm/                              # LLM 客户端
│   └── cli/                              # CLI 模式
└── docs/
    └── plans/                            # 设计和实施文档
```


## 故障排除

### "需要 LLM API key"
设置 `OPENAI_API_KEY` 环境变量或在 `config.yaml` 中添加 `llm.api_key`。

### "需要 CNB token"
设置 `CNB_TOKEN` 环境变量或在 `config.yaml` 中添加 `cnb.token`。

### "无法连接到 MCP"
- 检查你的 CNB token 是否有效
- 验证到 `https://mcp.cnb.cool` 的网络连接
- 确保 token 具有所需的权限

### 工具调用不工作
- 验证你的 LLM 模型支持函数调用
- 检查启动时是否列出了 MCP 工具
- 查看 skill 文件语法

## 相关资源

- [CNB 平台](https://cnb.cool) - 云原生构建平台
- [CNB MCP Server](https://cnb.cool/cnb/tools/cnb-mcp-server) - 官方 MCP 实现
- [CNB 文档](https://docs.cnb.cool) - 平台使用文档
- [MCP 协议](https://modelcontextprotocol.io) - 模型上下文协议标准
- [Claude Code](https://claude.com/claude-code) - Skill 机制来源

## 贡献

欢迎贡献!你可以:
- 报告 Bug 或提出建议
- 改进文档和示例
- 添加新的工具集成
- 优化 Skill 定义

## 许可证

参见 [LICENSE](LICENSE) 文件。

## 后记

想了半天如何写简介，忽然脑海冒出一句：`🙇 与其 awesome skills，不如 learn a skill`。

当今AI领域，各种概念更迭不休，保持学习的精神，抛却表面的焦虑，学习一个具体的知识，就能让自己变得充实且快乐。