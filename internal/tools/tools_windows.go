//go:build windows

package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/DoraZa/mini-agent/internal/llm"
)

// For Windows, many commands behave differently. We provide equivalents where possible.

func executePs(toolCall llm.ToolCall) (string, error) {
	// tasklist is the Windows equivalent of ps.
	// It doesn't have flags as flexible as ps, so we ignore them and use /FO CSV for parsable output.
	cmd := exec.Command("tasklist", "/FO", "CSV")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command 'tasklist' failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

func executeFind(toolCall llm.ToolCall) (string, error) {
	// 'dir /s /b' is a rough equivalent for 'find'.
	var args struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'find' arguments: %w", err)
	}
	if args.Path == "" {
		args.Path = "."
	}
	// The 'name' parameter is used as a pattern for 'dir'.
	searchPath := fmt.Sprintf("%s\\%s", args.Path, args.Name)
	if args.Name == "" {
		searchPath = args.Path
	}

	cmd := exec.Command("cmd", "/c", "dir", "/s", "/b", searchPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 'dir' can return exit code 1 if no files are found. This is not a fatal error.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No files found.", nil
		}
		return "", fmt.Errorf("command 'dir' failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

func executeGrep(toolCall llm.ToolCall) (string, error) {
	// 'findstr' is the Windows equivalent of grep.
	var args struct {
		Pattern    string `json:"pattern"`
		File       string `json:"file"`
		IgnoreCase bool   `json:"ignore_case"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'grep' arguments: %w", err)
	}
	if args.Pattern == "" || args.File == "" {
		return "", fmt.Errorf("the 'pattern' and 'file' arguments are required for grep/findstr")
	}

	cmdArgs := []string{}
	if args.IgnoreCase {
		cmdArgs = append(cmdArgs, "/I")
	}
	cmdArgs = append(cmdArgs, args.Pattern, args.File)

	cmd := exec.Command("findstr", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matching lines found.", nil
		}
		return "", fmt.Errorf("command 'findstr' failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

func executeWget(toolCall llm.ToolCall) (string, error) {
	// PowerShell's Invoke-WebRequest is a good equivalent for wget.
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

	psCommand := fmt.Sprintf("Invoke-WebRequest -Uri %s", args.URL)
	if args.OutputFile != "" {
		psCommand += fmt.Sprintf(" -OutFile %s", args.OutputFile)
	}

	cmd := exec.Command("powershell", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command 'Invoke-WebRequest' failed: %w, output: %s", err, string(output))
	}
	return fmt.Sprintf("Successfully downloaded from %s.", args.URL), nil
}

func executeLsof(toolCall llm.ToolCall) (string, error) {
	// 'netstat -ano' is the way to find processes by port on Windows.
	var args struct {
		Port string `json:"port"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("error decoding 'lsof' arguments: %w", err)
	}
	if args.Port == "" {
		return "", fmt.Errorf("the 'port' argument is required for lsof on Windows")
	}

	// Find the PID using the specified port
	// The output format is tricky, so we'll do our best.
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command 'netstat -ano' failed: %w, output: %s", err, string(output))
	}

	// Filter the output to find the relevant line.
	lines := strings.Split(string(output), "\n")
	result := "Proto  Local Address          Foreign Address        State           PID\n"
	found := false
	for _, line := range lines {
		if strings.Contains(line, ":"+args.Port) {
			result += strings.TrimSpace(line) + "\n"
			found = true
		}
	}

	if !found {
		return "No process found using the specified port.", nil
	}
	return result, nil
}

// 'ss' does not have a direct, simple equivalent on Windows.
// 'netstat' is the closest, but its functionality is more aligned with lsof.
// We will consider 'ss' unsupported on Windows for now.
func executeSs(toolCall llm.ToolCall) (string, error) {
	return "", fmt.Errorf("'ss' command is not supported on Windows. Use 'lsof' with a port instead")
}
