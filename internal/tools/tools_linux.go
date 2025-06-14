//go:build linux

package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/DoraZa/mini-agent/internal/llm"
)

// executePs 执行 'ps' 命令的 Linux 版本。
// 它根据 JSON 参数构建命令，可以直接使用 'ps' 的 `-u` 等标志进行高效过滤。
func executePs(toolCall llm.ToolCall) (string, error) {
	var args struct {
		User    string   `json:"user"`
		Name    string   `json:"name"`
		PID     string   `json:"pid"`
		Options []string `json:"options"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'ps' arguments: %w", err)
	}

	cmdArgs := []string{}
	if len(args.Options) > 0 {
		cmdArgs = append(cmdArgs, args.Options...)
	} else {
		// Linux 下 'aux' 是一个常见的默认值
		cmdArgs = append(cmdArgs, "aux")
	}

	// 在 Linux 上，我们可以直接使用 pgrep 或 ps -u 进行过滤，比管道更可靠
	if args.User != "" {
		cmd := exec.Command("ps", "-u", args.User)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("command 'ps -u %s' failed: %w, output: %s", args.User, err, string(output))
		}
		return string(output), nil
	}

	// 按名称过滤
	if args.Name != "" {
		cmd := exec.Command("pgrep", "-af", args.Name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return "No matching processes found by name.", nil
			}
			return "", fmt.Errorf("command 'pgrep -af %s' failed: %w, output: %s", args.Name, err, string(output))
		}
		return string(output), nil
	}

	// 按 PID 过滤
	if args.PID != "" {
		cmd := exec.Command("ps", "-p", args.PID)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("command 'ps -p %s' failed: %w, output: %s", args.PID, err, string(output))
		}
		return string(output), nil
	}

	// 如果没有特定过滤，则执行带选项的 ps 命令
	cmd := exec.Command("ps", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command 'ps %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	return string(output), nil
}

// executeFind 执行 'find' 命令。
// 它将 JSON 参数（如 path, name, type, maxdepth）转换为 `find` 命令的命令行标志。
func executeFind(toolCall llm.ToolCall) (string, error) {
	var args struct {
		Path     string `json:"path"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		MaxDepth int    `json:"maxdepth"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'find' arguments: %w", err)
	}
	if args.Path == "" {
		args.Path = "."
	}

	cmdArgs := []string{args.Path}
	if args.MaxDepth > 0 {
		cmdArgs = append(cmdArgs, "-maxdepth", strconv.Itoa(args.MaxDepth))
	}
	if args.Name != "" {
		cmdArgs = append(cmdArgs, "-name", args.Name)
	}
	if args.Type != "" {
		cmdArgs = append(cmdArgs, "-type", args.Type)
	}
	if len(cmdArgs) == 1 {
		return "", fmt.Errorf("at least one of 'name' or 'type' must be provided for find")
	}

	cmd := exec.Command("find", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command 'find %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	return string(output), nil
}

// executeGrep 执行 'grep' 命令。
// 它处理 pattern, file, 和布尔标志（如 ignore_case, recursive, count_only）。
func executeGrep(toolCall llm.ToolCall) (string, error) {
	var args struct {
		Pattern    string `json:"pattern"`
		File       string `json:"file"`
		Recursive  bool   `json:"recursive"`
		IgnoreCase bool   `json:"ignore_case"`
		CountOnly  bool   `json:"count_only"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'grep' arguments: %w", err)
	}
	if args.Pattern == "" || args.File == "" {
		return "", fmt.Errorf("the 'pattern' and 'file' arguments are required for grep")
	}

	cmdArgs := []string{}
	if args.Recursive {
		cmdArgs = append(cmdArgs, "-r")
	}
	if args.IgnoreCase {
		cmdArgs = append(cmdArgs, "-i")
	}
	if args.CountOnly {
		cmdArgs = append(cmdArgs, "-c")
	}
	cmdArgs = append(cmdArgs, args.Pattern)
	// 在 file 参数周围加上引号可能有助于处理带空格的文件名，但这里我们直接传递
	cmdArgs = append(cmdArgs, strings.Fields(args.File)...)

	cmd := exec.Command("grep", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matching lines found.", nil
		}
		return "", fmt.Errorf("command 'grep %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	return string(output), nil
}

// executeWget 执行 'wget' 命令。
// 它处理 URL 和可选的 output_file 参数。
func executeWget(toolCall llm.ToolCall) (string, error) {
	var args struct {
		URL        string `json:"url"`
		OutputFile string `json:"output_file"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'wget' arguments: %w", err)
	}
	if args.URL == "" {
		return "", fmt.Errorf("the 'url' argument is required for wget")
	}

	cmdArgs := []string{}
	if args.OutputFile != "" {
		cmdArgs = append(cmdArgs, "-O", args.OutputFile)
	}
	cmdArgs = append(cmdArgs, args.URL)

	cmd := exec.Command("wget", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command 'wget %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	if args.OutputFile != "" {
		return fmt.Sprintf("Successfully downloaded from %s to %s.", args.URL, args.OutputFile), nil
	}
	return fmt.Sprintf("Successfully downloaded from %s.", args.URL), nil
}

// executeSs 执行 'ss' 命令的 Linux 版本。
// 它将 JSON 参数转换为命令行标志，支持复杂的网络套接字查询。
func executeSs(toolCall llm.ToolCall) (string, error) {
	var args struct {
		Options  []string `json:"options"`
		Port     int      `json:"port"`
		Protocol string   `json:"protocol"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'ss' arguments: %w", err)
	}

	cmdArgs := []string{}
	if len(args.Options) > 0 {
		cmdArgs = append(cmdArgs, args.Options...)
	} else {
		cmdArgs = append(cmdArgs, "-tan") // 通用默认值
	}

	filterExpression := []string{}
	if args.Protocol != "" {
		filterExpression = append(filterExpression, args.Protocol)
	}
	if args.Port > 0 {
		// ss 使用 'sport' 和 'dport' 进行过滤
		filterExpression = append(filterExpression, fmt.Sprintf("( sport = :%d or dport = :%d )", args.Port, args.Port))
	}

	if len(filterExpression) > 0 {
		// 在 Linux 上，过滤器是 ss 命令的直接参数
		cmdArgs = append(cmdArgs, "state", "all", "and", strings.Join(filterExpression, " and "))
	}

	cmd := exec.Command("ss", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command 'ss %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	return string(output), nil
}

// executeLsof 执行 'lsof' 命令。
// 它支持按端口、文件路径或用户进行查询。
func executeLsof(toolCall llm.ToolCall) (string, error) {
	var args struct {
		Path    string   `json:"path"`
		Port    int      `json:"port"`
		User    string   `json:"user"`
		Options []string `json:"options"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'lsof' arguments: %w", err)
	}
	if args.Port == 0 && args.Path == "" && args.User == "" && len(args.Options) == 0 {
		return "", fmt.Errorf("at least one of 'port', 'path', 'user', or 'options' must be provided for lsof")
	}

	cmdArgs := []string{}
	if len(args.Options) > 0 {
		cmdArgs = append(cmdArgs, args.Options...)
	}
	if args.Port > 0 {
		cmdArgs = append(cmdArgs, "-i", fmt.Sprintf(":%d", args.Port))
	}
	if args.User != "" {
		cmdArgs = append(cmdArgs, "-u", args.User)
	}
	if args.Path != "" {
		cmdArgs = append(cmdArgs, args.Path)
	}

	cmd := exec.Command("lsof", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matching processes found.", nil
		}
		return "", fmt.Errorf("command 'lsof %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	return string(output), nil
}
