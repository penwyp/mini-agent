## 智能命令行 Agent 开发计划

本计划旨在将复杂的 Agent 开发过程分解为一系列小而可控的、可验证的步骤。我们将采用 Go 语言作为示例，因为它非常适合构建命令行工具。

### 准备工作：环境设置

在开始之前，请确保你已具备：

1.  **Go 语言环境**：版本 1.18 或更高。
2.  **DeepSeek API Key**：你需要一个有效的 API Key 来与 LLM 进行通信。
3.  **一个IDE或编辑器**：如 VS Code, GoLand, Cursor 等。

-----

### **阶段 0：项目初始化与基础架构搭建**

**目标**：创建项目结构、管理依赖并搭建一个最基础的可运行的空壳。

  * **步骤 0.1：初始化项目**

      * **任务**：创建一个新的项目目录，并使用 Go Modules 初始化项目。
      * **指令**：
        ```bash
        mkdir smart-cli-agent
        cd smart-cli-agent
        go mod init github.com/your-username/smart-cli-agent
        ```

  * **步骤 0.2：创建目录结构**

      * **任务**：建立清晰的模块化目录结构，为后续功能开发做准备。
      * **指令**：
        ```bash
        mkdir -p cmd/agent internal/llm internal/tools internal/config internal/history
        ```
      * **目录说明**：
          * `cmd/agent`: 主程序入口。
          * `internal/config`: 负责加载配置（如 API Key）。
          * `internal/llm`: 封装与 LLM API 的所有交互。
          * `internal/tools`: 负责工具的定义和执行。
          * `internal/history`: 管理对话历史。

  * **步骤 0.3：创建主程序入口**

      * **任务**：在 `cmd/agent/main.go` 中编写一个简单的 "Hello, Agent\!" 程序。
      * **代码位置**：`cmd/agent/main.go`
      * **内容**：
        ```go
        package main

        import "fmt"

        func main() {
            fmt.Println("智能命令行 Agent 启动...")
        }
        ```
      * **验证**：在项目根目录运行 `go run cmd/agent/main.go`，应该能看到 "智能命令行 Agent 启动..." 的输出。

-----

### **阶段 1：实现核心的 LLM API 通信**

**目标**：让 Agent 能够通过 API Key 与 DeepSeek LLM 进行最基础的文本问答。

  * **步骤 1.1：配置管理**

      * **任务**：创建 `internal/config/config.go`，用于从环境变量 `AGENT_API_KEY` 或配置文件加载 API Key（只支持 AGENT_API_KEY）。
      * **验证**：编写一个简单的测试来确认函数可以成功读取环境变量。

  * **步骤 1.2：定义 API 数据结构**

      * **任务**：在 `internal/llm/types.go` 中，根据 DeepSeek API 文档定义请求和响应的核心结构体。至少需要 `Message`（包含 `role` 和 `content`），以及请求体和响应体。
      * **验证**：代码能够编译通过。

  * **步骤 1.3：创建 LLM 客户端**

      * **任务**：创建 `internal/llm/client.go`。实现一个 `Client` 结构体，并提供一个 `NewClient` 函数来初始化它（包含 API Key 和 `http.Client`）。
      * **验证**：代码能够编译通过。

  * **步骤 1.4：实现简单的聊天功能**

      * **任务**：在 `LLM Client` 中实现一个方法 `Chat(messages []Message) (string, error)`。该方法向 LLM 发送一个简单的请求（不含函数调用），并返回纯文本响应。
      * **任务**：修改 `main.go`，使其：
        1.  加载 API Key。
        2.  初始化 LLM 客户端。
        3.  发送一个硬编码的消息（例如 `{"role": "user", "content": "你好"}`）。
        4.  打印 LLM 的回复。
      * **验证**：运行 `go run cmd/agent/main.go`，你应该能看到来自 DeepSeek 的问候语。**这标志着与 LLM 的通信链路已打通。**

-----

### **阶段 2：实现第一个工具的完整调用流程 (ReAct 单次循环)**

**目标**：实现从用户意图到 LLM 生成工具调用，再到 Agent 执行该工具并获得观察结果的完整闭环。我们选择最简单的 `ps` 命令作为起点。

  * **步骤 2.1：扩展 LLM 类型定义**

      * **任务**：回到 `internal/llm/types.go`，根据 PRD 补充 `Tool`、`Function` 和 `ToolCall` 等与 Function Calling 相关的完整数据结构。
      * **验证**：代码能够编译通过。

  * **步骤 2.2：定义工具 Schema**

      * **任务**：创建 `internal/tools/definitions.go`。编写一个函数，该函数返回一个包含所有工具定义的 JSON 字符串或 Go 结构体数组，**严格按照 PRD 中的 `System Prompt` 部分提供的 JSON 格式**。
      * **验证**：确保生成的 JSON 格式与 PRD 中的完全一致。

  * **步骤 2.3：创建工具执行器**

      * **任务**：创建 `internal/tools/executor.go`。定义一个 `ExecuteTool` 函数，它接收一个 `llm.ToolCall` 对象，并返回执行结果（字符串）和错误。
      * **任务**：在 `ExecuteTool` 中，使用 `switch` 语句根据 `toolCall.Function.Name` 来分发任务。先实现 `ps` 的 `case`。
      * **任务**：在 `ps` 的实现中，使用 `os/exec` 包来执行系统中的 `ps` 命令，并捕获其标准输出和标准错误。
      * **验证**：为 `ExecuteTool` 编写一个单元测试，手动构造一个 `ps` 的 `ToolCall`，断言函数返回了预期的进程信息。

  * **步骤 2.4：集成 Function Calling 到主流程**

      * **任务**：修改 `main.go` 和 `llm.Client`。在调用 LLM API 时，将步骤 2.2 中定义的工具 Schema 一同发送。
      * **任务**：修改 `main.go` 的逻辑：
        1.  接收 LLM 的响应。
        2.  检查响应中是否包含 `tool_calls`。
        3.  如果包含，提取第一个 `tool_call`。
        4.  调用 `tools.ExecuteTool` 执行它。
        5.  打印 `Thought` 和执行后的 `Observation`（即命令输出）。
      * **验证**：运行 Agent，输入指令："**列出当前所有进程**"。程序应能正确打印出 LLM 的思考过程，然后执行 `ps` 命令，并显示进程列表。**这是项目的核心里程碑。**

-----

### **阶段 3：构建完整的 ReAct 迭代循环**

**目标**：将单次循环扩展为多步迭代循环，直到任务完成。

  * **步骤 3.1：实现消息历史管理器**

      * **任务**：创建 `internal/history/manager.go`。定义一个 `HistoryManager`，内部维护一个 `[]llm.Message` 切片。
      * **任务**：提供方法，如 `AddUserMessage`, `AddAssistantMessage(thought string, toolCalls []llm.ToolCall)`, 和 `AddToolObservation(toolCallID string, content string)`。
      * **验证**：为 `HistoryManager` 编写单元测试，验证消息能被正确添加和格式化。

  * **步骤 3.2：实现主循环**

      * **任务**：重构 `main.go`，建立一个 `for` 循环作为 Agent 的主生命周期。
      * **循环逻辑**：
        1.  **初始化**：在循环外，用 PRD 中的 `System Prompt` 初始化消息历史。
        2.  **用户输入**：在循环开始时，接收用户的自然语言指令，并添加到消息历史中。
        3.  **调用 LLM**：将完整的消息历史发送给 LLM。
        4.  **添加 LLM 回复**：将 LLM 的回复（`Thought` + `ToolCalls`）添加到历史记录中。
        5.  **判断与行动**：
              * **如果 LLM 返回 `tool_calls`**：遍历并执行所有 `tool_calls`，将每个执行结果（`Observation`）作为 `role: tool` 的消息（包含 `tool_call_id`）添加回历史记录。然后 `continue` 进入下一次循环。
              * **如果 LLM 返回最终答案**（没有 `tool_calls`）：打印最终答案，然后 `break` 退出循环。
      * **验证**：给出一个需要两步才能完成的任务，例如："**先找到当前目录下所有叫 'go.mod' 的文件，然后告诉我它的内容**"（假设 `grep` 或 `cat` 工具已实现）。观察 Agent 是否能成功完成两次循环并给出最终答案。

-----

### **阶段 4：扩展并完善工具集**

**目标**：逐一实现 PRD 中要求的所有工具，并确保参数映射正确。

  * **步骤 4.1 - 4.5：逐一实现工具**
      * **任务**：对于 `find`, `grep`, `wget`, `ss`, `lsof` 中的每一个工具：
        1.  在 `internal/tools/executor.go` 的 `switch` 语句中添加一个新的 `case`。
        2.  实现该工具的逻辑。**关键在于将 LLM 返回的 JSON 参数（例如 `{"ignore_case": true, "pattern": "error"}`）正确地转换为实际的命令行标志和参数（例如 `grep -i "error"`）**。
        3.  仔细处理每个工具的输出和潜在错误。
      * **验证**：每完成一个工具，就设计一个具体的指令来测试它：
          * **find**: "在当前目录和子目录中查找所有名为 'README.md' 的文件"
          * **grep**: "在 go.mod 文件中搜索包含 'smart' 的行，忽略大小写"
          * **wget**: "帮我下载 [https://www.google.com/robots.txt](https://www.google.com/robots.txt)"
          * **ss**: "显示所有正在监听的 TCP 连接"
          * **lsof**: "查看是哪个进程占用了 8080 端口"

-----

### **阶段 5：提升健壮性与用户体验 (非功能性需求)**

**目标**：根据 PRD 实现安全、稳定和易用的特性。

  * **步骤 5.1：安全加固**

      * **任务**：在 `tools.ExecuteTool` 中，执行 `os/exec` 之前，添加一个**命令白名单验证**。只允许执行 PRD 中列出的命令，如果 LLM 生成了其他命令（如 `rm`），则立即拒绝并返回错误。
      * **验证**：给出一个恶意指令 "删除我的家目录"，Agent 应拒绝执行并报错。

  * **步骤 5.2：用户体验优化**

      * **任务**：在 `main.go` 的主循环中，添加清晰的状态反馈。
          * 在调用 LLM 前，打印 `> 思考中...`
          * 在执行工具前，打印 `> 正在执行: find . -name "*.log"`
      * **任务 (可选)**：添加一个交互式确认。在执行命令前，询问用户 `是否执行此命令? [Y/n]`，增加操作的透明度和可控性。
      * **验证**：运行 Agent，观察输出是否清晰、易于理解。

  * **步骤 5.3：完善错误处理**

      * **任务**：确保 `tools.ExecuteTool` 在命令执行失败时（例如文件不存在、权限不足），能够捕获 `stderr` 和 Go 的 `error`，并按照 PRD 的要求格式化为 `[ERROR]: <错误信息>` 返回。
      * **验证**：给出一个必然失败的指令，例如 "在不存在的文件 'abc.txt' 中搜索内容"，观察 Agent 能否将错误信息传递给 LLM，并观察 LLM 是否能根据错误信息进行下一步判断（例如，报告文件不存在）。

  * **步骤 5.4：实现退出机制**

      * **任务**：在接收用户输入的逻辑中，增加对 `exit` 或 `quit` 命令的判断，如果匹配则优雅地退出程序。
      * **验证**：在 Agent 提示符下输入 `exit`，程序应正常终止。

-----

### **阶段 6：最终润色与文档化**

**目标**：完成项目，使其可以被其他人理解和使用。

  * **步骤 6.1：代码审查与重构**

      * **任务**：通读所有代码，确保命名规范、注释清晰、逻辑简洁。移除不必要的测试代码。

  * **步骤 6.2：编写 README 文档**

      * **任务**：在项目根目录创建 `README.md` 文件。
      * **内容**：
          * 项目简介。
          * 如何配置（设置 `AGENT_API_KEY` 环境变量或在 configs/config.yaml 里配置 api_key 字段）。
          * 如何编译和运行。
          * 提供几个使用示例。

  * **步骤 6.3：进行端到端测试**

      * **任务**：使用 PRD 中提到的复杂任务进行测试，例如："**帮我找出所有大于 1GB 的日志文件，并显示它们的详细信息**"。虽然这个任务可能超出现有工具的组合能力（需要`find`配合`ls`或`du`），但正好可以测试 Agent 的边界处理能力——它应该能利用 `find` 找到文件，然后向用户报告它无法直接获取文件大小，或者尝试使用现有工具进行近似分析。
      * **验证**：Agent 能够逻辑连贯地处理复杂指令，并在能力不足时给出合理解释，而不是崩溃或陷入死循环。

