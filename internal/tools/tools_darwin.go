//go:build darwin

package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/DoraZa/mini-agent/internal/llm"
)

// executePs 执行 'ps' 命令。
// 它根据 JSON 参数构建命令，支持按用户、名称、PID 和其他 ps 选项进行过滤。
// 注意：在 Darwin (macOS) 上，没有直接的方法可以像在 Linux 上那样通过 `ps -u <user>` 来按用户名过滤，
// 并且 `pgrep` 的行为也可能有所不同。这里的实现是一个简化版本，主要依赖 `ps` 的 flags 和 `grep` 管道。
// 一个更健壮的实现可能需要更复杂的逻辑或不同的工具。
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

	// 基础命令为 ps aux
	cmdArgs := []string{"aux"}
	if len(args.Options) > 0 {
		// 如果用户提供了 options，则使用用户的 options，否则使用 'aux' 作为默认值。
		cmdArgs = args.Options
	}

	// 如果需要按名称或用户过滤，则使用管道连接 grep
	// 注意：这是一个简化的实现。
	if args.Name != "" || args.User != "" {
		var grepPattern string
		if args.Name != "" {
			grepPattern = args.Name
		} else {
			grepPattern = args.User
		}

		// 管道链: ps aux | grep "pattern"
		psCmd := exec.Command("ps", cmdArgs...)
		grepCmd := exec.Command("grep", grepPattern)

		pipe, err := psCmd.StdoutPipe()
		if err != nil {
			return "", fmt.Errorf("failed to create pipe for ps|grep: %w", err)
		}
		grepCmd.Stdin = pipe

		if err := psCmd.Start(); err != nil {
			return "", fmt.Errorf("failed to start ps command for ps|grep: %w", err)
		}

		output, err := grepCmd.CombinedOutput()
		if err != nil {
			// grep 在没有匹配项时会返回退出码 1，这不应被视为致命错误。
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return "No matching processes found.", nil
			}
			return "", fmt.Errorf("ps|grep pipeline failed: %w, output: %s", err, string(output))
		}
		return string(output), nil
	}

	// 如果没有过滤，直接执行 ps 命令
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
		MaxDepth int    `json:"maxdepth"` // 注意：JSON 数字会自动解析为 Go 的数值类型
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
	if len(cmdArgs) == 1 { // 只有路径，没有其他参数
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
	cmdArgs = append(cmdArgs, args.Pattern, args.File)

	cmd := exec.Command("grep", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// grep 在没有找到匹配项时返回退出码 1，这对于 Agent 来说是有效信息，不应视为错误。
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matching lines found.", nil
		}
		// 其他错误（如文件不存在）应被报告。
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
		// wget 的错误通常包含在 stderr 中，由 CombinedOutput() 捕获
		return "", fmt.Errorf("command 'wget %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	// wget 成功时，stdout 通常为空。返回一条确认信息。
	if args.OutputFile != "" {
		return fmt.Sprintf("Successfully downloaded from %s to %s.", args.URL, args.OutputFile), nil
	}
	return fmt.Sprintf("Successfully downloaded from %s.", args.URL), nil
}

// executeSs 执行 'ss' 命令。
// 在 macOS 上 'ss' 命令不可用，此函数会尝试使用 'netstat' 作为替代方案来查找端口信息。
func executeSs(toolCall llm.ToolCall) (string, error) {
	var args struct {
		Options  []string `json:"options"`
		Port     int      `json:"port"`
		Protocol string   `json:"protocol"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'ss' arguments: %w", err)
	}

	// macOS 上没有 'ss' 命令，因此我们直接检查是否可以转为 'netstat' 或 'lsof'
	if args.Port > 0 {
		// 如果目标是查询端口，lsof 是一个更好的选择
		return executeLsofForPort(args.Port)
	}

	return "", fmt.Errorf("'ss' command is not available on macOS. Try using 'lsof' to check for a specific port")
}

// executeLsofForPort 是一个辅助函数，使用 lsof -i:<port> 来模拟 'ss' 或 'netstat' 的端口查询功能。
func executeLsofForPort(port int) (string, error) {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return fmt.Sprintf("No processes found using port %d.", port), nil
		}
		return "", fmt.Errorf("lsof command failed for port %d: %w, output: %s", port, err, string(output))
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
	hasNonPathArgs := false

	if len(args.Options) > 0 {
		cmdArgs = append(cmdArgs, args.Options...)
		hasNonPathArgs = true
	}
	if args.Port > 0 {
		// 使用 -i TCP:<port> 来查找特定端口的 TCP 连接
		cmdArgs = append(cmdArgs, "-i", fmt.Sprintf("TCP:%d", args.Port))
		hasNonPathArgs = true
	}
	if args.User != "" {
		cmdArgs = append(cmdArgs, "-u", args.User)
		hasNonPathArgs = true
	}

	// lsof 可以将文件路径作为主要参数，而不是标志参数
	if args.Path != "" {
		// 如果同时有其他参数，最好将它们放在前面
		if hasNonPathArgs {
			finalArgs := append(cmdArgs, "--", args.Path)
			cmd := exec.Command("lsof", finalArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
					return "No matching processes found.", nil
				}
				return "", fmt.Errorf("command 'lsof %s' failed: %w, output: %s", strings.Join(finalArgs, " "), err, string(output))
			}
			return string(output), nil
		}
		// 只有路径
		cmd := exec.Command("lsof", args.Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return "No matching processes found.", nil
			}
			return "", fmt.Errorf("command 'lsof %s' failed: %w, output: %s", args.Path, err, string(output))
		}
		return string(output), nil
	}

	cmd := exec.Command("lsof", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// lsof 在没有找到匹配项时返回退出码 1，这不应视为致命错误。
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matching processes found.", nil
		}
		return "", fmt.Errorf("command 'lsof %s' failed: %w, output: %s", strings.Join(cmdArgs, " "), err, string(output))
	}
	return string(output), nil
}
