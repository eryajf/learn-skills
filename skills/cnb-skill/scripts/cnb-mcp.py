#!/usr/bin/env python3
"""
CNB MCP 调用脚本
通过 HTTP API 调用 CNB MCP server (https://mcp.cnb.cool/mcp)
支持从环境变量或 .env 文件读取 CNB_TOKEN
"""

import sys
import json
import os
from pathlib import Path

def load_env_file():
    """
    加载 .env 文件（如果存在）
    查找顺序：
    1. 当前工作目录的 .env
    2. 脚本所在目录的 .env
    3. 项目根目录的 .env

    返回: (成功加载, 配置字典)
    """
    env_paths = [
        Path.cwd() / '.env',  # 当前工作目录
        Path(__file__).parent / '.env',  # 脚本目录
        Path(__file__).parent.parent / '.env',  # 项目根目录
    ]

    env_vars = {}
    for env_path in env_paths:
        if env_path.exists():
            try:
                with open(env_path, 'r', encoding='utf-8') as f:
                    for line in f:
                        line = line.strip()
                        # 跳过空行和注释
                        if not line or line.startswith('#'):
                            continue
                        # 解析 KEY=VALUE 格式
                        if '=' in line:
                            key, value = line.split('=', 1)
                            key = key.strip()
                            value = value.strip()
                            # 移除引号
                            if value.startswith('"') and value.endswith('"'):
                                value = value[1:-1]
                            elif value.startswith("'") and value.endswith("'"):
                                value = value[1:-1]
                            if key and value:
                                env_vars[key] = value
                return True, env_vars
            except Exception as e:
                # 读取失败不影响程序继续运行
                print(f"警告: 读取 .env 文件失败: {e}", file=sys.stderr)
                pass

    return False, {}

def call_mcp(method, params=None):
    """调用 MCP server"""
    # 先尝试加载 .env 文件
    _, env_vars = load_env_file()

    # 获取 CNB token（优先从 .env 文件读取,如果为空则从系统环境变量获取）
    token = env_vars.get('CNB_TOKEN') or os.getenv('CNB_TOKEN')

    if not token:
        print(json.dumps({
            "error": "CNB_TOKEN 未设置",
            "hint": "请设置 CNB_TOKEN 环境变量或在 .env 文件中配置 CNB_TOKEN=your_token"
        }, ensure_ascii=False), file=sys.stderr)
        sys.exit(1)

    # 构建 JSON-RPC 请求
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": method,
    }
    if params:
        request["params"] = params

    # MCP endpoint 地址（固定值,一般不需要修改）
    mcp_url = 'https://mcp.cnb.cool/mcp'

    import subprocess

    try:
        result = subprocess.run(
            [
                'curl', '-s', '-X', 'POST',
                mcp_url,
                '-H', 'Content-Type: application/json',
                '-H', 'Accept: application/json, text/event-stream',
                '-H', f'Authorization: Bearer {token}',
                '-d', json.dumps(request)
            ],
            capture_output=True,
            text=True,
            timeout=30
        )

        if result.returncode != 0:
            print(json.dumps({"error": f"curl 执行失败: {result.stderr}"}), file=sys.stderr)
            sys.exit(1)

        # 解析响应（可能是 SSE 格式）
        try:
            output = result.stdout.strip()

            # 检查是否是 SSE 格式 (event: message\ndata: {...})
            if output.startswith('event:'):
                # 解析 SSE 格式
                lines = output.split('\n')
                for line in lines:
                    if line.startswith('data:'):
                        json_str = line[5:].strip()  # 移除 "data:" 前缀
                        response = json.loads(json_str)
                        break
                else:
                    print(json.dumps({"error": "SSE 格式中未找到 data 行"}), file=sys.stderr)
                    sys.exit(1)
            else:
                # 普通 JSON 响应
                response = json.loads(output)

            # 检查 JSON-RPC 错误
            if "error" in response:
                print(json.dumps({"error": response["error"]}), file=sys.stderr)
                sys.exit(1)

            return response.get('result', response)
        except json.JSONDecodeError as e:
            print(json.dumps({
                "error": f"JSON 解析失败: {str(e)}",
                "output": result.stdout
            }), file=sys.stderr)
            sys.exit(1)

    except subprocess.TimeoutExpired:
        print(json.dumps({"error": "MCP 调用超时"}), file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)

def main():
    if len(sys.argv) < 2:
        print("用法: cnb-mcp.py <operation> [args...]")
        print("支持的操作:")
        print("  list-tools          - 列出所有可用工具")
        print("  call <tool> <args>  - 调用指定工具")
        sys.exit(1)

    operation = sys.argv[1]

    if operation == "list-tools":
        result = call_mcp("tools/list")
        print(json.dumps(result, ensure_ascii=False, indent=2))

    elif operation == "call":
        if len(sys.argv) < 3:
            print("错误: 需要指定工具名称")
            sys.exit(1)

        tool_name = sys.argv[2]
        tool_args = {}

        # 解析参数 (格式: key=value)
        for arg in sys.argv[3:]:
            if '=' in arg:
                key, value = arg.split('=', 1)
                # 尝试解析为 JSON，否则作为字符串
                try:
                    tool_args[key] = json.loads(value)
                except:
                    tool_args[key] = value

        result = call_mcp("tools/call", {
            "name": tool_name,
            "arguments": tool_args
        })
        print(json.dumps(result, ensure_ascii=False, indent=2))

    else:
        print(f"未知操作: {operation}")
        sys.exit(1)

if __name__ == "__main__":
    main()
