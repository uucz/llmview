<div align="center">

# llmview

**See what your AI is thinking. See what it costs.**

A local proxy that intercepts LLM API calls and shows them in a real-time dashboard.
Zero code changes. One binary. Dark mode.

[English](README.md) | [中文](README_zh.md)

<!-- TODO: Replace with actual screenshot/GIF -->
<!--
<img src="docs/screenshot.png" alt="llmview dashboard" width="720" />
-->

[![CI](https://github.com/uucz/llmview/actions/workflows/ci.yml/badge.svg)](https://github.com/uucz/llmview/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/uucz/llmview)](https://goreportcard.com/report/github.com/uucz/llmview)

</div>

---

## Quick Start

```bash
# Install (pick one)
go install github.com/uucz/llmview@latest
# or: download from https://github.com/uucz/llmview/releases
# or: docker run -p 4700:4700 ghcr.io/uucz/llmview

# Run
llmview

# Point your AI tools at it
export OPENAI_BASE_URL=http://localhost:4700/proxy/openai/v1
export ANTHROPIC_BASE_URL=http://localhost:4700/proxy/anthropic
```

Open **http://localhost:4700** — every LLM call now streams through your dashboard.

## Why

I spent **$47 debugging an AI agent last week**. Blind. No idea which calls were burning tokens, which prompts were bloated, or where the agent was looping.

Existing tools are either:
- **Cloud-hosted** (Helicone, AgentOps) — your prompts leave your machine
- **SDK-based** (Langfuse, Phoenix) — requires code changes and framework lock-in
- **CLI-only** (llm-interceptor) — no UI, just raw logs

llmview sits between your AI tools and the API. **You change one environment variable.** That's it. Your prompts never leave your machine. Everything shows up in a real-time dashboard.

## Features

| Feature | Description |
|---------|-------------|
| **Real-time timeline** | Watch API calls stream in as they happen |
| **Live token streaming** | See the response being generated token-by-token |
| **Request/response viewer** | Parsed message threads + syntax-highlighted JSON bodies |
| **Filter & search** | Filter by provider, status code, model. Search across all calls. |
| **Export** | Download session data as JSON or CSV |
| **Cost tracking** | Per-call and session-total cost with per-model pricing |
| **Multi-provider** | OpenAI, Anthropic, Ollama — all through one dashboard |
| **Zero code changes** | Just set an environment variable |
| **Single binary** | One ~10MB file. No database to install. No Docker required. |
| **Local & private** | SQLite storage. Nothing leaves your machine. |
| **Dark theme** | Because you're probably running this at 2am |

## Supported Providers

| Provider | Environment Variable | Works With |
|----------|---------------------|------------|
| OpenAI | `OPENAI_BASE_URL=http://localhost:4700/proxy/openai/v1` | GPT-4o, o1, o3, any OpenAI model |
| Anthropic | `ANTHROPIC_BASE_URL=http://localhost:4700/proxy/anthropic` | Claude Opus, Sonnet, Haiku |
| Ollama | `OLLAMA_HOST=http://localhost:4700/proxy/ollama` | Llama, Mistral, Qwen, any local model |

Works with **any tool** that uses these SDKs: Claude Code, Cursor, Aider, LangChain, CrewAI, OpenAI Python/Node SDK, Anthropic SDK, and more.

## How It Works

```
Your Agent / IDE / Script
         │
         ▼  (just an env var change)
    ┌─────────┐
    │ llmview │ ← intercepts, logs, calculates cost
    └────┬────┘
         │
         ▼  (forwards to real API)
   OpenAI / Anthropic / Ollama
```

llmview is a **reverse proxy**. It receives the request, records it, forwards it to the real API, records the response, calculates the cost, and pushes everything to the dashboard via WebSocket. Streaming responses are forwarded chunk-by-chunk with zero added latency.

## Configuration

```bash
# Custom port (default: 4700)
llmview --port 8080

# Custom database path (default: ~/.llmview/llmview.db)
llmview --db /path/to/data.db
```

### Model Pricing

llmview ships with built-in pricing for popular models (GPT-4o, Claude Sonnet, etc.). Local models (Ollama) are tracked as free. Pricing updates with new releases.

## REST API

llmview exposes a JSON API for programmatic access:

```bash
# Current session stats
curl http://localhost:4700/api/session

# List recent calls
curl http://localhost:4700/api/calls?limit=20&offset=0

# Get full request/response for a specific call
curl http://localhost:4700/api/calls/{id}

# Health check
curl http://localhost:4700/api/health
```

## Building from Source

```bash
git clone https://github.com/uucz/llmview.git
cd llmview
make build    # builds UI + Go binary
make test     # runs all tests
```

Requires: Go 1.25+, Node.js 18+ (for UI build). No C compiler needed — pure Go.

## Troubleshooting

<details>
<summary><b>macOS: "permission denied" when running the binary</b></summary>

Downloaded binaries don't have execute permission by default. Fix:

```bash
chmod +x llmview-darwin-arm64
./llmview-darwin-arm64
```
</details>

<details>
<summary><b>macOS: "cannot be opened because the developer cannot be verified"</b></summary>

macOS Gatekeeper blocks unsigned binaries. Go to **System Settings > Privacy & Security**, scroll down, and click **"Open Anyway"**. Or remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine llmview-darwin-arm64
```
</details>

<details>
<summary><b>"command not found: llmview"</b></summary>

If you downloaded the binary, you need to use the full path or `./` prefix:

```bash
# Run from current directory
./llmview-darwin-arm64

# Or move to a directory in your PATH
sudo mv llmview-darwin-arm64 /usr/local/bin/llmview
llmview
```

If you used `go install`, make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"
```
</details>

<details>
<summary><b>Common mistakes</b></summary>

```bash
# WRONG: "go" is for compiling source code, not running binaries
go llmview

# WRONG: "run" is not a system command
run llmview

# WRONG: just the filename doesn't work without PATH setup
llmview-darwin-arm64

# CORRECT:
./llmview-darwin-arm64
```
</details>

## Roadmap

- [x] Real-time proxy + dashboard
- [x] Token cost tracking
- [x] Multi-provider support
- [x] Request/response detail viewer with message thread parsing
- [x] Export sessions to JSON/CSV
- [x] Filter & search calls by provider, status, model
- [ ] VS Code extension
- [ ] Breakpoints (pause before dangerous operations)
- [ ] Prompt diff viewer (compare request variations)
- [ ] Session history browser

## License

MIT

---

<div align="center">

**If llmview saved you from a surprise API bill, consider giving it a star.**

</div>
