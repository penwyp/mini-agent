//go:build !linux && !darwin && !windows

package tools

import (
	"fmt"
	"runtime"

	"github.com/DoraZa/mini-agent/internal/llm"
)

var errUnsupportedOS = fmt.Errorf("command not supported on operating system: %s", runtime.GOOS)

func executePs(toolCall llm.ToolCall) (string, error) {
	return "", errUnsupportedOS
}

func executeFind(toolCall llm.ToolCall) (string, error) {
	return "", errUnsupportedOS
}

func executeGrep(toolCall llm.ToolCall) (string, error) {
	return "", errUnsupportedOS
}

func executeWget(toolCall llm.ToolCall) (string, error) {
	return "", errUnsupportedOS
}

func executeSs(toolCall llm.ToolCall) (string, error) {
	return "", errUnsupportedOS
}

func executeLsof(toolCall llm.ToolCall) (string, error) {
	return "", errUnsupportedOS
}
