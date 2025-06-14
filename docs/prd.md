# 产品需求文档：智能命令行 Agent

## 1. 引言

### 1.1 产品概述

本项目旨在开发一款**智能命令行 Agent**。它能理解用户的自然语言指令，并通过与大型语言模型（LLM）API（如 DeepSeek）进行**迭代的"思考（Reasoning）"和"行动（Acting）"循环（ReAct 范式）**来完成复杂任务。Agent 的核心是利用 LLM 的 **Function Calling（函数调用）**能力，将用户意图转化为一系列可执行的系统命令。Agent 将在命令行中执行这些命令，并将结果（**观察 Observation**）反馈给 LLM，促使 LLM 持续规划，最终达成用户目标并返回结果。这款 Agent 旨在革新命令行交互模式，显著提升用户操作效率和便捷性。

### 1.2 目标用户

* **高级开发者/SRE/DevOps 工程师**：需要处理复杂、多步骤的系统诊断或自动化任务，希望通过自然语言实现高效、容错的操作。
* **数据科学家/分析师**：频繁进行文件查找、数据处理（如日志分析）等操作，希望通过智能 Agent 简化工作流。
* **教育与研究人员**：用于探索 LLM 工具调用和复杂任务规划能力，作为教学或实验平台。

### 1.3 产品愿景

构建一个能进行多步推理和行动的智能命令行 Agent，将命令行操作从单次命令执行转变为基于意图的复杂任务自动化，从而降低命令行使用门槛，实现真正的"对话式操作"。

---

## 2. 功能需求

### 2.1 核心功能：ReAct 循环

Agent 的核心功能围绕 ReAct 循环展开，这是实现复杂任务处理的关键：

* **用户指令接收**：Agent 提供命令行界面，接收用户输入的自然语言文本指令（例如："帮我找出所有大于 1GB 的日志文件，并显示它们的详细信息"）。
* **LLM 推理与行动规划 (Thought & Action)**：
    * Agent 将用户指令和当前的**对话历史**（包含 LLM 的思考、行动和观察结果）作为输入发送给 LLM API。
    * **思考 (Thought)**：LLM 根据用户意图、当前状态及对话历史进行内部思考，规划解决问题的步骤，并**以文本形式输出其思考过程**。
    * **行动 (Action)**：根据思考结果，LLM 利用 **Function Calling 机制**，从 Agent 提供的"工具"（系统命令）列表中选择最合适的工具，并生成其参数，以结构化格式（`tool_calls`）输出。
* **命令执行 (Execution)**：
    * Agent **解析 LLM 返回的 `tool_calls`**。
    * Agent 在底层操作系统中执行相应的系统命令。
    * **初始支持的命令**：`find`、`grep`、`wget`、`ps`、`ss`、`lsof`。
* **结果观察 (Observation)**：
    * Agent 捕获命令的执行结果（**标准输出、标准错误或执行状态**）。
    * 这些结果被视为**观察（Observation）**，并以结构化或非结构化文本形式记录。
* **循环迭代**：Agent 将 `Observation` 作为新的消息（`role: tool`）添加到对话历史中，并标记与哪个 `tool_call_id` 相关联。更新后的对话历史会再次发送给 LLM，促使其根据新的观察结果进行下一轮的思考和行动，直到任务完成或 LLM 判断无法继续。
* **最终结果输出**：当 LLM 判断任务完成（例如，通过 `finish_reason` 为 `stop` 或 LLM 生成一个总结性回答），Agent 会将最终结果或总结清晰地展示给用户。
* **退出机制**：用户可以输入特定命令（例如 `exit`）来退出 Agent。

### 2.2 LLM 交互设计：Prompt 的核心

Prompt 的设计是 Agent 智能性的核心。

#### 2.2.1 System Prompt (系统提示)

**目的**：定义 LLM 的角色、能力边界、ReAct 工作流，以及期望的输出格式。这是指导 LLM 进行推理和行动的基石。

```
你是一个高度智能且负责任的命令行 Agent，能够理解复杂的自然语言指令，并通过一系列"思考(Thought)"、"行动(Action)"和"观察(Observation)"的迭代循环来完成任务。你的目标是作为一名专家，精确地将用户的意图转化为一系列可执行的系统命令，并在收到执行结果后，根据结果继续推理或提供最终答案。

以下是你必须严格遵守的规则和工作流程：

**工作模式：ReAct 循环**
你将严格遵循 Thought -> Action -> Observation 的循环模式，直到任务成功完成。
- **Thought (思考)**：在采取任何行动之前，你必须先阐述你的思考过程。这包括：
    - 你对用户意图的清晰理解。
    - 你当前正在解决的问题。
    - 你计划采取的下一步行动是什么，以及为什么选择这个行动。
    - 你预期这个行动会带来什么结果，并评估其潜在风险。
    - 如果是多步任务，你还需要思考后续的步骤、依赖关系和整体策略。
    - 思考过程必须清晰、有逻辑，以便于 Agent 和用户理解。
- **Action (行动)**：根据你的思考，调用你被授予的工具。
    - 你必须使用 JSON 格式的 `tool_calls` 来指定要调用的工具及其参数。
    - 如果任务需要，你可以同时生成多个 `tool_calls`。
    - **Function Calling 严格要求**：确保为选定的工具提供所有必要且准确的参数。参数值必须符合工具定义的 JSON Schema。如果用户输入不足以形成完整参数，你应在 Thought 中解释原因，并可能请求更多信息，而不是生成不完整的 Action。
- **Observation (观察)**：Agent 将执行你指定的 Action，并将执行结果作为 Observation 反馈给你。你将收到一个 `role: tool` 的消息，其 `content` 字段包含命令的实际输出。你必须仔细分析这些 Observation 来进行下一步的 Thought。
- **Task Completion (任务完成)**：当任务成功完成，或者你判断无法通过现有工具继续，或者需要用户提供更多信息时，你可以生成一个最终的总结性回答（非 Tool Call）。这个回答应该清晰、直接，并说明任务的结果或你的限制。

**可用工具和使用指南：**
以下是你被授予的工具列表，请仔细阅读它们的描述和参数。你只能使用这些工具，且不能臆造不存在的工具或参数。

```json
[
  {
    "type": "function",
    "function": {
      "name": "find",
      "description": "在文件系统中查找文件或目录。返回所有匹配项的路径。当不指定路径时，默认在当前目录及子目录中查找。",
      "parameters": {
        "type": "object",
        "properties": {
          "name": { "type": "string", "description": "要查找的文件或目录的名称，支持通配符。" },
          "path": { "type": "string", "description": "查找的起始路径，例如 '.' (当前目录) 或 '/home/user/'。如果未指定，默认在当前目录递归查找。" },
          "type": { "type": "string", "description": "查找类型，'f' 表示文件，'d' 表示目录。", "enum": ["f", "d"] },
          "maxdepth": { "type": "integer", "description": "查找的最大深度，例如 1 表示只在当前目录查找，不进入子目录。" }
        },
        "required": ["name"]
      }
    }
  },
  {
    "type": "function",
    "function": {
      "name": "grep",
      "description": "在文件中搜索匹配指定模式的行。返回所有匹配的行。",
      "parameters": {
        "type": "object",
        "properties": {
          "pattern": { "type": "string", "description": "要搜索的正则表达式或字符串模式。" },
          "file": { "type": "string", "description": "要搜索的文件名或路径。如果是多个文件，可以是以空格分隔的字符串。" },
          "recursive": { "type": "boolean", "description": "如果为true，递归搜索目录下的文件，相当于 grep -r。" },
          "ignore_case": { "type": "boolean", "description": "如果为true，忽略大小写，相当于 grep -i。" },
          "count_only": { "type": "boolean", "description": "如果为true，只返回匹配行的数量，相当于 grep -c。" }
        },
        "required": ["pattern", "file"]
      }
    }
  },
  {
    "type": "function",
    "function": {
      "name": "wget",
      "description": "从互联网下载文件。将文件保存到当前目录或指定路径。",
      "parameters": {
        "type": "object",
        "properties": {
          "url": { "type": "string", "description": "要下载文件的 URL。" },
          "output_file": { "type": "string", "description": "可选。下载后文件的保存名称或路径。" }
        },
        "required": ["url"]
      }
    }
  },
  {
    "type": "function",
    "function": {
      "name": "ps",
      "description": "列出当前运行的进程。可以根据用户、进程名或 PID 过滤。注意：在 macOS 上，按用户或名称过滤是通过 'grep' 管道实现的，这可能与 Linux 上的行为略有不同。",
      "parameters": {
        "type": "object",
        "properties": {
          "user": { "type": "string", "description": "按用户名过滤进程。" },
          "name": { "type": "string", "description": "按进程名过滤进程。" },
          "pid": { "type": "string", "description": "按进程 ID 过滤进程。" },
          "options": { "type": "array", "items": { "type": "string" }, "description": "ps 命令的额外选项，例如 '-ef' 或 '-aux'。" }
        },
        "required": []
      }
    }
  },
  {
    "type": "function",
    "function": {
      "name": "ss",
      "description": "显示套接字统计信息，用于查看网络连接。可以过滤特定端口或连接状态。注意：'ss' 命令在 macOS 上不可用。当按端口查询时，此工具将自动使用 'lsof' 作为替代方案。",
      "parameters": {
        "type": "object",
        "properties": {
          "options": { "type": "array", "items": { "type": "string" }, "description": "ss 命令的选项，例如 '-l' (监听), '-t' (TCP), '-u' (UDP), '-n' (数字显示), '-p' (显示进程), '-a' (所有套接字)。" },
          "port": { "type": "integer", "description": "过滤指定端口的连接。" },
          "protocol": { "type": "string", "description": "过滤指定协议，例如 'tcp' 或 'udp'。" }
        },
        "required": []
      }
    }
  },
  {
    "type": "function",
    "function": {
      "name": "lsof",
      "description": "列出打开的文件。可以查看进程打开的文件，或端口被哪个进程占用。",
      "parameters": {
        "type": "object",
        "properties": {
          "path": { "type": "string", "description": "查找打开指定文件的进程。" },
          "port": { "type": "integer", "description": "查找占用指定端口的进程。" },
          "user": { "type": "string", "description": "查找某个用户打开的文件。" },
          "options": { "type": "array", "items": { "type": "string" }, "description": "lsof 命令的额外选项，例如 '-i' (列出所有网络文件)。" }
        },
        "required": []
      }
    }
  }
]
```

**输出格式要求：**
- 如果你决定进行 Thought 和 Action，你的输出应该严格遵循以下结构，并且 `Thought` 必须在前，紧接着是 `tool_calls` 的 JSON 结构：
  ```
  Thought: <你的思考过程>
  <然后立即是 tool_calls 的JSON结构>
  ```
- 如果你认为任务已经完成、无法继续，或者需要用户提供更多信息，则直接输出最终的总结性回答（非 Tool Call）。
- 你的回复不应包含任何额外的寒暄或不必要的文本，保持简洁、专业。
```

#### 2.2.2 Messages 历史管理

`messages` 数组是 LLM 交互的核心，它包含了 Agent 和 LLM 之间的完整对话历史，用于提供上下文并实现 ReAct 循环。

* **初始化**：首次调用 LLM 时，`messages` 数组包含 `System Prompt` 和用户的第一个 `User Message`。
* **迭代过程中的消息添加**：
    * **LLM 生成 Thought 和 Action**：Agent 接收到 LLM 的响应后，将 LLM 的 `Thought` 文本和 `tool_calls` 结构体解析，并封装成一个 `role: assistant` 的消息添加到 `messages` 历史中。
        * 示例：`{"role": "assistant", "content": "Thought: 用户想查找文件...Action:", "tool_calls": [...]}`
    * **Agent 执行 Action 并生成 Observation**：命令执行的输出（标准输出、标准错误）被包装成一个 `role: tool` 的消息添加到 `messages` 历史中。
        * **`tool_call_id` 是关键**：它必须与 LLM 返回的 `tool_calls` 中的 ID 匹配，以将 `Observation` 与特定 `Action` 关联起来。
        * 示例：`{"role": "tool", "tool_call_id": "call_123", "content": "<命令的实际输出>"}`
* **`messages` 序列示例**：
    ```json
    [
      {"role": "system", "content": "你是一个智能命令行 Agent... (System Prompt)"},
      {"role": "user", "content": "找找当前目录下有没有叫做'report.pdf'的文件，然后告诉我它的大小。"},
      {"role": "assistant", "content": "Thought: 用户想查找文件，并获取其大小。我首先需要使用 `find` 命令找到文件。Action:", "tool_calls": [
        {"id": "call_123", "type": "function", "function": {"name": "find", "arguments": "{\"name\":\"report.pdf\", \"path\": \".\", \"type\": \"f\"}"}}
      ]},
      {"role": "tool", "tool_call_id": "call_123", "content": "/home/user/documents/report.pdf\n"}, // find 命令的 Observation
      {"role": "assistant", "content": "Thought: 我已经找到了文件 'report.pdf' 的路径。现在我需要获取它的大小。我可以使用 `ls -lh` 命令来获取文件信息。Action:", "tool_calls": [
        {"id": "call_456", "type": "function", "function": {"name": "ls", "arguments": "{\"path\":\"/home/user/documents/report.pdf\", \"options\":[\"-lh\"]}"}} // 假设 'ls' 是可用工具
      ]},
      {"role": "tool", "tool_call_id": "call_456", "content": "-rw-r--r-- 1 user user 2.5M May 10 14:30 /home/user/documents/report.pdf\n"}, // ls 命令的 Observation
      {"role": "assistant", "content": "文件 'report.pdf' 存在于 `/home/user/documents/`，其大小为 2.5MB。"} // LLM 最终总结
    ]
    ```

### 2.3 Observation 设计

**目的**：将命令执行的实际结果，以清晰、可被 LLM 理解的格式反馈给 LLM，作为其下一步推理的依据。准确的 Observation 对于 LLM 的 ReAct 循环至关重要。

* **内容**：
    * **标准输出 (stdout)**：命令成功执行时的主要信息流。
    * **标准错误 (stderr)**：命令执行失败或警告信息。
    * **统一格式**：将 `stdout` 和 `stderr` 合并为一个字符串作为 `content`。如果 `stderr` 不为空，应在 `content` 中清晰指出其为错误信息，例如 `"[ERROR]: <stderr 内容>\n<stdout 内容>"`。
* **处理空输出**：如果命令执行没有 `stdout` 或 `stderr`（例如 `touch` 命令成功执行），`content` 应为空字符串。
* **处理长输出**：对于某些命令（如 `ps -ef`），输出可能非常长。Agent 应考虑：
    * **截断**：设置最大 `content` 长度，截断过长的输出，并在末尾添加提示（例如 `"... [输出已截断]"`)。
    * **概括性提示**：对于极长的输出，LLM 可能难以处理。Agent 可以只返回一个概括性提示（例如 `"命令执行成功，但输出过长。"`），如果 LLM 需要详细信息，再通过特定的工具请求。
* **错误信息**：如果命令执行失败（例如，命令不存在，参数错误，权限不足），确保将完整的错误信息（包括 `stderr` 和 Go 语言 `exec` 包返回的错误）清晰地包含在 `content` 中，以便 LLM 能够诊断问题。
    * 示例 `Observation` for `wget` 失败：
        `{"role": "tool", "tool_call_id": "call_789", "content": "[ERROR]: wget: unable to resolve host address 'nonexistent.url'\n"}`
* **LLM 的消息结构**：
    * Observation 必须封装为 `{"role": "tool", "tool_call_id": "..." , "content": "<命令的 stdout/stderr>"}`。
    * `tool_call_id` 是唯一的标识符，它将特定的 `Observation` 关联到 LLM 之前生成的那次 `tool_call`。这是 ReAct 循环中信息关联的关键。

---

## 3. 非功能需求

* **性能**：LLM 响应和命令执行应尽可能快速。ReAct 循环会增加 API 调用次数，需关注整体延迟和吞吐量。
* **安全性**：
    * **API Key 管理**：LLM API Key 等敏感信息应通过环境变量或安全配置管理，绝不硬编码。
    * **命令沙箱/白名单**：尽管 LLM 会在 `System Prompt` 限制下生成命令，但在 Agent 端执行前，必须对所有 LLM 返回的命令及其参数进行严格的**白名单校验和参数消毒**，以防止潜在的命令注入、恶意命令执行（如 `rm -rf /`）或资源滥用。
    * **最小权限原则**：Agent 进程应以最小必要的系统权限运行。
* **可扩展性**：设计应允许未来轻松添加更多支持的系统命令、自定义工具或切换到其他 LLM 模型。工具定义应易于管理和更新。
* **稳定性与健壮性**：Agent 应能稳定运行，处理各种用户输入、LLM 响应和命令执行结果，避免崩溃。应具备对异常情况（如网络中断、LLM API 错误、命令执行失败）的容错和恢复机制。
* **可维护性**：代码结构清晰，模块化程度高，易于理解、测试和维护。
* **用户体验**：
    * **清晰的提示符**：Agent 应提供明确的命令行提示符（例如 `Agent >`）。
    * **实时反馈**：在 LLM 处理和命令执行过程中，提供适当的加载或等待提示，并显示当前处于 ReAct 循环的哪个阶段（例如 "Thought...", "Executing Action...", "Observing result...")。
    * **友好错误信息**：当无法理解用户指令、LLM 未返回有效命令或命令执行失败时，提供清晰、易懂的错误或提示信息。
    * **命令展示**：Agent 可以选择性地在执行 LLM 生成的命令之前，向用户展示即将执行的命令，以增加透明度和用户控制（例如，询问"是否执行此命令？[Y/N]"）。

---

## 4. 技术设计（高层）

* **模块化架构**：
    * **用户接口模块 (User Interface)**：负责命令行输入输出，管理用户会话。
    * **LLM 交互模块 (LLM Interaction)**：封装与 DeepSeek API 的通信逻辑，处理 API 请求的构建和响应的接收。
    * **消息历史管理模块 (Message History Manager)**：负责维护和更新 `messages` 数组，确保上下文的正确性。
    * **LLM 响应解析模块 (LLM Response Parser)**：专门用于解析 LLM 返回的 `Thought` 文本和 `tool_calls` 结构体。
    * **工具执行模块 (Tool Executor)**：负责将 LLM 生成的抽象工具调用转换为具体的系统命令（包括参数转换，如 `ignore_case=true` 转换为 `-i`），执行命令，并捕获 `Observation`。此模块需包含命令白名单和参数校验。
    * **工具定义模块 (Tool Definitions)**：集中管理 LLM 可用的工具的 JSON Schema 定义。
* **并发处理**：考虑到 LLM API 调用可能存在延迟，可以考虑异步或并发地处理 LLM 请求，但目前 Demo 仍以串行处理为主。
* **环境变量配置**：所有敏感信息和可配置参数通过环境变量或配置文件管理。

---

## 5. 未来展望

* **更复杂的错误处理和恢复**：LLM 能够基于失败的 `Observation` 进行更智能的错误诊断和重试策略，甚至可以尝试自我修正错误。
* **多代理协作**：引入多个 Agent，每个 Agent 专注于不同类型的任务，并通过内部通信机制相互协作完成复杂目标。
* **学习与适应能力**：Agent 可以从成功的案例中学习并优化其规划策略，甚至根据用户的使用习惯进行个性化调整。
* **持久化会话**：保存会话历史和 Agent 状态，允许用户在不同时间点继续之前的任务。
* **集成更多外部 API**：不限于系统命令，可以调用外部 Web 服务 API、自定义脚本、数据库查询等。
* **上下文增强**：集成更多关于当前系统环境的信息（如当前目录、用户信息、环境变量），作为 LLM 额外上下文输入，提高推理准确性。
