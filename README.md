# 智能命令行 Agent (Mini-Agent)

本项目是一个使用 Go 语言开发的智能命令行 Agent。它能理解用户的自然语言指令，通过与大型语言模型（LLM）的函数调用（Function Calling）功能进行交互，将用户意图转化为一系列可执行的系统命令。Agent 遵循 **ReAct (Reasoning and Acting)** 的模式，通过"思考 -> 行动 -> 观察"的循环来完成复杂任务。

## 功能特性

- **自然语言交互**：直接用自然语言下达命令，无需记住复杂的命令行语法。
- **ReAct 循环**：通过持续的"思考、行动、观察"循环，处理多步骤的复杂任务。
- **工具集**：内置了一系列常用的系统命令作为工具（如 `find`, `grep`, `ps`, `lsof` 等）。
- **可扩展**：可以方便地添加新的工具或更换 LLM 模型。
- **安全**：内置命令白名单机制，防止执行危险命令。

## 环境准备

在开始之前，请确保你已具备：

1.  **Go 语言环境**：版本 1.18 或更高。
2.  **一个兼容 OpenAI 的 LLM API Key**：如 DeepSeek、Moonshot、OpenAI。
3.  **(可选) `wget` 命令**: 如果你想使用 `wget` 工具，请确保你的系统已安装该命令。

## 快速开始

### 1. 配置 API Key

Agent 需要通过环境变量或配置文件获取你的 LLM API Key 和其他可选配置。

**方式一：环境变量（推荐）**

在你的 shell 配置文件（如 `.zshrc`, `.bash_profile`）中添加：

```bash
# 必须：你的 LLM API Key
export AGENT_API_KEY="sk-your-api-key"

# (可选) 自定义 API Endpoint
# export AGENT_BASE_URL="https://api.deepseek.com/v1"

# (可选) 指定要使用的模型名称
# export AGENT_MODEL="deepseek-coder"
```

**方式二：配置文件**

你也可以在 `configs/config.yaml` 中设置：

```yaml
api_key: "sk-your-api-key"
model: "deepseek-coder"
base_url: "https://api.deepseek.com/v1"
allowed_tools:
  - "ps"
  - "find"
  - "grep"
  - "wget"
  - "ss"
  - "lsof"
denied_tools: []
```

**注意**：环境变量优先生效，其次为配置文件，最后为默认值。

### 2. 编译和运行

你可以使用 `go run` 直接运行 Agent，或使用 `Makefile` 进行编译。

**方式一：直接运行**

在项目根目录执行：
```bash
go run cmd/agent/main.go
```

**方式二：使用 Makefile 编译**

首先，编译项目：
```bash
make build
```
这会在 `bin/` 目录下生成一个名为 `mini-agent` 的可执行文件。

然后，运行 Agent：
```bash
./bin/mini-agent
```

## 使用示例

启动 Agent 后，你将看到一个提示符 `>`。现在你可以输入你的指令了。

**示例 1: 查找文件**

> 帮我找到当前目录下所有的 go.mod 文件。

**示例 2: 查看进程**

> 列出所有正在运行的 go 进程。

**示例 3: 查看端口占用**

> 哪个进程占用了 8080 端口？

**示例 4: 退出**

在提示符后输入 `exit` 或 `quit` 即可退出 Agent。

## 技术架构

- **主循环**: `cmd/agent/main.go` 包含了核心的 ReAct 循环逻辑。
- **LLM 通信**: `internal/llm/` 负责与 LLM API 进行交互。
- **工具定义与执行**: `internal/tools/` 定义了所有可用工具的 Schema，并负责执行这些工具。
- **历史管理**: `internal/history/` 负责管理对话历史，为 LLM 提供上下文。
- **配置**: `internal/config/` 负责加载环境变量。 