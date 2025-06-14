package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DoraZa/mini-agent/internal/config"
	"github.com/DoraZa/mini-agent/internal/history"
	"github.com/DoraZa/mini-agent/internal/llm"
	"github.com/DoraZa/mini-agent/internal/tools"
)

// systemPrompt 定义了代理应该如何行为的系统级指令
const systemPrompt = `你是一个高度智能且负责任的命令行 Agent，能够理解复杂的自然语言指令，并通过一系列"思考(Thought)"、"行动(Action)"和"观察(Observation)"的迭代循环来完成任务。你的目标是作为一名专家，精确地将用户的意图转化为一系列可执行的系统命令，并在收到执行结果后，根据结果继续推理或提供最终答案。

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
    - 你必须使用 JSON 格式的 tool_calls 来指定要调用的工具及其参数。
    - 如果任务需要，你可以同时生成多个 tool_calls。
    - **Function Calling 严格要求**：确保为选定的工具提供所有必要且准确的参数。参数值必须符合工具定义的 JSON Schema。如果用户输入不足以形成完整参数，你应在 Thought 中解释原因，并可能请求更多信息，而不是生成不完整的 Action。
- **Observation (观察)**：Agent 将执行你指定的 Action，并将执行结果作为 Observation 反馈给你。你将收到一个 role: tool 的消息，其 content 字段包含命令的实际输出。你必须仔细分析这些 Observation 来进行下一步的 Thought。
- **Task Completion (任务完成)**：当任务成功完成，或者你判断无法通过现有工具继续，或者需要用户提供更多信息时，你可以生成一个最终的总结性回答（非 Tool Call）。这个回答应该清晰、直接，并说明任务的结果或你的限制。

**可用工具和使用指南：**
你只能使用通过 'tools' 参数提供给你的工具。请仔细阅读它们的描述和参数，不要臆造不存在的工具或参数。

**输出格式要求：**
- 如果你决定进行 Thought 和 Action，你的输出应该严格遵循 "Thought: <你的思考过程>" 的格式，紧接着是 tool_calls 的JSON结构。
- 如果你认为任务已经完成、无法继续，或者需要用户提供更多信息，则直接输出最终的总结性回答（非 Tool Call）。
- 你的回复不应包含任何额外的寒暄或不必要的文本，保持简洁、专业。`

func main() {
	fmt.Println("智能命令行 Agent 启动... (输入 'exit' 或 'quit' 退出)")

	// 1. 加载配置
	// 从环境变量或 .env 文件中加载配置，如 API Key, Base URL, Model 等。
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. 初始化 LLM 客户端和工具定义
	// 基于配置创建一个与 OpenAI API 兼容的客户端。
	var llmClient llm.LLM = llm.NewOpenAIClient(cfg.APIKey, cfg.BaseURL)
	// 获取所有已定义的工具的 JSON Schema。
	toolDefs := tools.GetToolDefinitions()

	// 3. 主交互循环
	// 将历史记录管理器移到循环外部，以实现多轮对话记忆。
	histManager := history.NewHistoryManager()
	histManager.AddSystemMessage(systemPrompt)

	// 使用 bufio.Scanner 来读取用户的多行输入。
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// 3.1. 获取用户输入
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break // 如果读取失败（例如 EOF），则退出循环。
		}
		userInput := scanner.Text()
		trimmedInput := strings.TrimSpace(userInput)

		// 检查退出命令
		if strings.ToLower(trimmedInput) == "exit" || strings.ToLower(trimmedInput) == "quit" {
			break
		}

		// 如果输入为空，则跳过本次循环，不与 LLM 交互
		if trimmedInput == "" {
			continue
		}

		histManager.AddUserMessage(trimmedInput)

		// 3.2. 内部 ReAct 循环
		// 这个循环会一直持续，直到 LLM 认为任务完成（不返回工具调用）或发生错误。
		for {
			// 准备发送给 LLM 的请求
			request := llm.ChatRequest{
				Model:    cfg.Model,
				Messages: histManager.GetHistory(),
				Tools:    toolDefs,
			}

			fmt.Println("🤔 Thinking...")
			// 调用 LLM 获取响应
			response, err := llmClient.ChatCompletion(context.Background(), request)
			if err != nil {
				log.Printf("Error from LLM: %v", err)
				break // 出现错误时，中断内部循环，等待用户新指令。
			}

			if len(response.Choices) == 0 {
				log.Println("Received no choices from LLM.")
				break
			}

			// 提取助手的消息并将其添加到历史记录中
			assistantMessage := response.Choices[0].Message
			histManager.AddAssistantMessage(assistantMessage)

			// 3.3. 判断是思考+行动，还是最终答案
			hasToolCalls := len(assistantMessage.ToolCalls) > 0
			hasContent := assistantMessage.Content != nil && *assistantMessage.Content != ""

			if hasToolCalls {
				// ReAct 循环中的一步：思考 -> 行动
				if hasContent {
					fmt.Printf("📝 Thought: %s\n", *assistantMessage.Content)
				}

				// 如果 LLM 返回了工具调用，则执行它们
				for _, toolCall := range assistantMessage.ToolCalls {
					fmt.Printf("🔧 Executing tool: %s(%s)\n", toolCall.Function.Name, toolCall.Function.Arguments)

					// 交互式确认：在执行前请求用户批准
					fmt.Print("Do you want to execute this command? [Y/n]: ")
					if !scanner.Scan() {
						log.Println("Scanner failed, cancelling execution.")
						break
					}
					confirmation := strings.ToLower(strings.TrimSpace(scanner.Text()))

					var observation string
					var err error

					if confirmation == "n" || confirmation == "no" {
						// 用户取消了执行
						fmt.Println("❌ Execution cancelled by user.")
						observation = "User cancelled the execution of this tool."
					} else {
						// 用户批准执行（输入 "y" 或直接按 Enter）
						observation, err = tools.ExecuteTool(toolCall, cfg.AllowedTools, cfg.DeniedTools)
						if err != nil {
							// 如果工具执行失败，将错误信息作为观察结果。
							// 这允许 LLM "看到"错误并据此决定下一步行动。
							fmt.Printf("❌ Error executing tool '%s': %v\n", toolCall.Function.Name, err)
							observation = fmt.Sprintf("Error: %v", err)
						}
					}

					// 确保 observation 永不为空
					if observation == "" {
						observation = "(No output)" // 使用英文以保持日志一致性
					}

					// 将工具执行的观察结果添加到历史记录中
					histManager.AddToolObservation(toolCall.ID, observation)
					// 打印观察结果，让用户了解发生了什么
					fmt.Printf("🔭 Observation: %s\n", observation)
				}
				// 工具执行完毕，继续内部 ReAct 循环
				continue
			} else {
				// ReAct 循环的结束：最终答案
				if hasContent {
					fmt.Println("\n✅ Final Answer:")
					// 清理模型可能返回的不必要的前缀
					content := *assistantMessage.Content
					content = strings.TrimSpace(content)
					content = strings.TrimPrefix(content, "Thought:")
					content = strings.TrimPrefix(content, "Final Answer:")
					fmt.Println(strings.TrimSpace(content))
				}
				// 结束当前任务的 ReAct 循环
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from stdin: %v", err)
	}

	fmt.Println("\nAgent session ended.")
}
