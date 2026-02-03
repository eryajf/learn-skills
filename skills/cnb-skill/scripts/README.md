# CNB MCP 脚本使用指南

这个脚本用于通过 HTTP API 调用 CNB MCP server。

## 配置

### 方式 1: 使用 .env 文件 (推荐)

1. 复制示例文件：
   ```bash
   cp .env.example .env
   ```

2. 编辑 `.env` 文件，填入你的 CNB Access Token：
   ```bash
   CNB_TOKEN=your_actual_token_here
   ```

3. `.env` 文件会被自动加载，脚本会在以下位置查找：
   - 当前工作目录
   - 脚本所在目录
   - 项目根目录

### 方式 2: 使用环境变量

```bash
export CNB_TOKEN=your_actual_token_here
export CNB_MCP_URL=https://mcp.cnb.cool/mcp  # 可选
```

**注意**：环境变量优先级高于 .env 文件。

## 获取 CNB Access Token

访问 [CNB Access Token 文档](https://docs.cnb.cool/zh/guide/access-token.html) 了解如何获取你的访问令牌。

## 使用示例

### 列出所有可用工具

```bash
python3 cnb-mcp.py list-tools
```

### 调用工具

```bash
# 查询知识库
python3 cnb-mcp.py call cnb_queryKnowledgeBase query="如何配置webhook"

# 获取仓库信息
python3 cnb-mcp.py call cnb_get_repository repo="demo-app"

# 触发流水线
python3 cnb-mcp.py call cnb_startBuild repo="demo-app" branch="main"
```

## 安全提示

- **永远不要**将包含真实 token 的 `.env` 文件提交到 Git 仓库
- `.env` 已被添加到 `.gitignore`，确保不会意外提交
- 如需分享配置示例，请使用 `.env.example` 文件
