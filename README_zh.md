<div align="center">

# llmview

**LLM API 的 Chrome DevTools。**

本地反向代理，拦截、检查、重放、预算控制你的 LLM API 调用 — 配有实时仪表盘。
零代码改动。单文件运行。交互式调试。

[English](README.md) | [中文](README_zh.md)

<!-- TODO: 替换为实际截图/GIF -->
<!--
<img src="docs/screenshot.png" alt="llmview 仪表盘" width="720" />
-->

[![CI](https://github.com/uucz/llmview/actions/workflows/ci.yml/badge.svg)](https://github.com/uucz/llmview/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/uucz/llmview)](https://goreportcard.com/report/github.com/uucz/llmview)

</div>

---

## 快速开始

```bash
# 安装（选一种方式）
go install github.com/uucz/llmview@latest
# 或者：从 https://github.com/uucz/llmview/releases 下载二进制文件
# 或者：docker run -p 4700:4700 ghcr.io/uucz/llmview

# 运行
llmview

# 将 AI 工具指向 llmview
export OPENAI_BASE_URL=http://localhost:4700/proxy/openai/v1
export ANTHROPIC_BASE_URL=http://localhost:4700/proxy/anthropic
```

打开 **http://localhost:4700** — 所有 LLM 调用都会实时出现在仪表盘上。

## 为什么做这个

上周我花了 **$47 调试一个 AI Agent**。完全盲调。不知道哪些调用在烧 token，哪些 prompt 太臃肿，也不知道 Agent 在哪里死循环。

现有的工具要么：
- **云端托管**（Helicone、AgentOps）— 你的 prompt 会离开你的电脑
- **需要改代码**（Langfuse、Phoenix）— 需要集成 SDK，绑定框架
- **纯命令行**（llm-interceptor）— 没有 UI，只有原始日志

llmview 坐在你的 AI 工具和 API 之间。**你只需要改一个环境变量。** 就这样。你的 prompt 永远不会离开你的电脑。所有东西都会出现在实时仪表盘上。

## 功能特性

| 特性 | 说明 |
|------|------|
| **实时时间线** | 看着 API 调用实时进来 |
| **流式 token 展示** | 一个 token 一个 token 地看到响应生成 |
| **请求重放** | 一键重发任意调用，支持修改参数或切换模型 |
| **预算控制** | 设置会话费用上限 — 超支时代理返回 402 |
| **费用追踪** | 每次调用和会话总计费用，按模型定价 |
| **多供应商** | OpenAI、Anthropic、Ollama — 一个仪表盘搞定 |
| **零代码改动** | 只需设置一个环境变量 |
| **单文件** | 约 10MB 文件，无需安装数据库，无需 Docker |
| **本地隐私** | SQLite 存储，数据不离开你的电脑 |
| **暗黑主题** | 因为你大概率是凌晨 2 点在用这个 |

## 支持的供应商

| 供应商 | 环境变量 | 支持模型 |
|--------|----------|----------|
| OpenAI | `OPENAI_BASE_URL=http://localhost:4700/proxy/openai/v1` | GPT-4o、o1、o3 等所有 OpenAI 模型 |
| Anthropic | `ANTHROPIC_BASE_URL=http://localhost:4700/proxy/anthropic` | Claude Opus、Sonnet、Haiku |
| Ollama | `OLLAMA_HOST=http://localhost:4700/proxy/ollama` | Llama、Mistral、Qwen 等所有本地模型 |

兼容**所有使用这些 SDK 的工具**：Claude Code、Cursor、Aider、LangChain、CrewAI、OpenAI Python/Node SDK、Anthropic SDK 等。

## 工作原理

```
你的 Agent / IDE / 脚本
         │
         ▼  （只需改一个环境变量）
    ┌─────────┐
    │ llmview │ ← 拦截、记录、计算费用
    └────┬────┘
         │
         ▼  （转发到真实 API）
   OpenAI / Anthropic / Ollama
```

llmview 是一个**反向代理**。它接收请求，记录下来，转发到真实 API，记录响应，计算费用，然后通过 WebSocket 把所有信息推送到仪表盘。流式响应逐块转发，零额外延迟。

## 配置

```bash
# 自定义端口（默认：4700）
llmview --port 8080

# 自定义数据库路径（默认：~/.llmview/llmview.db）
llmview --db /path/to/data.db

# 设置预算上限（超支时代理返回 402）
llmview --budget 5.00
```

### 模型定价

llmview 内置了热门模型的定价（GPT-4o、Claude Sonnet 等）。本地模型（Ollama）按免费计算。定价随新版本更新。

## REST API

llmview 提供 JSON API 用于程序化访问：

```bash
# 当前会话统计
curl http://localhost:4700/api/session

# 列出最近的调用
curl http://localhost:4700/api/calls?limit=20&offset=0

# 获取单次调用的完整请求/响应
curl http://localhost:4700/api/calls/{id}

# 重放调用（可覆盖参数）
curl -X POST http://localhost:4700/api/replay \
  -H 'Content-Type: application/json' \
  -d '{"call_id":"abc123","overrides":{"model":"gpt-4o-mini"}}'

# 获取配置（预算信息）
curl http://localhost:4700/api/config

# 健康检查
curl http://localhost:4700/api/health
```

## 从源码构建

```bash
git clone https://github.com/uucz/llmview.git
cd llmview
make build    # 构建 UI + Go 二进制文件
make test     # 运行所有测试
```

需要：Go 1.25+、Node.js 18+（用于 UI 构建）。不需要 C 编译器 — 纯 Go 实现。

## 常见问题

<details>
<summary><b>macOS："permission denied"（权限不足）</b></summary>

下载的二进制文件默认没有执行权限。解决方法：

```bash
chmod +x llmview-darwin-arm64
./llmview-darwin-arm64
```

**原因**：macOS/Linux 中，新下载的文件被视为普通文件。`chmod +x` 赋予它「可执行」权限。
</details>

<details>
<summary><b>macOS："无法打开，因为无法验证开发者"</b></summary>

macOS Gatekeeper 会阻止未签名的二进制文件。两种解决方法：

**方法一**：前往 **系统设置 > 隐私与安全性**，拉到最下方，点击 **"仍要打开"**。

**方法二**：使用命令移除隔离标记：

```bash
xattr -d com.apple.quarantine llmview-darwin-arm64
```
</details>

<details>
<summary><b>"command not found: llmview"（找不到命令）</b></summary>

在 macOS/Linux 终端中，直接输入文件名只会在系统 PATH 路径中查找，不会查找当前文件夹。

```bash
# 正确：用 ./ 前缀运行当前目录下的程序
./llmview-darwin-arm64

# 或者：移动到系统路径中，以后直接用 llmview 命令
sudo mv llmview-darwin-arm64 /usr/local/bin/llmview
llmview
```

如果你用 `go install` 安装，确保 `~/go/bin` 在你的 PATH 中：

```bash
export PATH="$HOME/go/bin:$PATH"
```

可以把上面这行加到 `~/.zshrc` 或 `~/.bashrc` 中永久生效。
</details>

<details>
<summary><b>常见错误命令</b></summary>

```bash
# 错误："go" 是编译 Go 源码的工具，不能直接运行二进制文件
go llmview

# 错误："run" 不是系统命令
run llmview

# 错误：不加 ./ 前缀，系统找不到当前目录的文件
llmview-darwin-arm64

# 正确：
./llmview-darwin-arm64
```
</details>

<details>
<summary><b>快速设置别名（可选）</b></summary>

如果你不想每次都输入完整路径，可以设置一个别名：

```bash
# 把下面这行加到 ~/.zshrc 或 ~/.bashrc
alias llmview='/usr/local/bin/llmview'

# 或者直接移动二进制文件到 PATH 路径
sudo mv llmview-darwin-arm64 /usr/local/bin/llmview
```
</details>

## 路线图

- [x] 实时代理 + 仪表盘
- [x] Token 费用追踪
- [x] 多供应商支持
- [x] 请求/响应详情查看器（支持消息线程解析）
- [x] 导出会话为 JSON/CSV
- [x] 按供应商、状态、模型过滤和搜索
- [x] 请求重放（支持参数覆盖）
- [x] 预算控制（实时进度条）
- [x] Prompt 差异对比器
- [ ] VS Code 扩展
- [ ] 断点调试（在危险操作前暂停）
- [ ] 历史会话浏览器

## 许可证

MIT

---

<div align="center">

**如果 llmview 帮你避免了一笔意外的 API 账单，考虑给它一个 Star。**

</div>
