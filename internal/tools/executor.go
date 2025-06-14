package tools

import (
	"fmt"
	"slices"

	"github.com/DoraZa/mini-agent/internal/llm"
)

// ExecuteTool 执行一个工具调用，并应用安全策略。
// 它首先检查调用的工具是否被允许，然后分发给相应的执行函数。
func ExecuteTool(toolCall llm.ToolCall, allowedTools, deniedTools []string) (string, error) {
	toolName := toolCall.Function.Name

	// 安全策略：
	// 1. 如果白名单不为空（默认情况），工具必须在白名单中。
	// 2. 工具决不能在黑名单中。
	isDenied := slices.Contains(deniedTools, toolName)
	if isDenied {
		return "", fmt.Errorf("tool '%s' is in the configured blacklist", toolName)
	}

	// 默认情况下，allowedTools 包含所有支持的工具。
	// 如果用户特意配置了白名单，则以用户的配置为准。
	if len(allowedTools) > 0 {
		isAllowed := slices.Contains(allowedTools, toolName)
		if !isAllowed {
			return "", fmt.Errorf("tool '%s' is not in the configured whitelist", toolName)
		}
	}

	// 根据工具名称分发到具体的实现函数
	switch toolName {
	case "ps":
		return executePs(toolCall)
	case "find":
		return executeFind(toolCall)
	case "grep":
		return executeGrep(toolCall)
	case "wget":
		return executeWget(toolCall)
	case "ss":
		return executeSs(toolCall)
	case "lsof":
		return executeLsof(toolCall)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}
