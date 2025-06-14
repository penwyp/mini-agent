package tools

import (
	"github.com/DoraZa/mini-agent/internal/llm"
)

// GetToolDefinitions 返回所有可用工具的定义。
// 这些定义严格遵循 PRD 文档中的 JSON Schema，以确保 LLM 能够理解和正确使用。
func GetToolDefinitions() []llm.Tool {
	return []llm.Tool{
		{
			Type: "function",
			Function: llm.Function{
				Name:        "ps",
				Description: "列出当前运行的进程。可以根据用户、进程名或 PID 过滤，类似 'ps' 命令。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"user": map[string]any{
							"type":        "string",
							"description": "按用户名过滤进程。",
						},
						"name": map[string]any{
							"type":        "string",
							"description": "按进程名过滤进程。",
						},
						"pid": map[string]any{
							"type":        "string",
							"description": "按进程 ID 过滤进程。",
						},
						"options": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "ps 命令的额外选项，例如 '-ef' 或 '-aux'。",
						},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "find",
				Description: "在文件系统中查找文件或目录。返回所有匹配项的路径。当不指定路径时，默认在当前目录及子目录中查找。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "查找的起始路径，例如 '.' (当前目录) 或 '/home/user/'。如果未指定，默认在当前目录递归查找。",
						},
						"name": map[string]any{
							"type":        "string",
							"description": "要查找的文件或目录的名称，支持通配符。",
						},
						"type": map[string]any{
							"type":        "string",
							"description": "查找类型，'f' 表示文件，'d' 表示目录。",
							"enum":        []string{"f", "d"},
						},
						"maxdepth": map[string]any{
							"type":        "integer",
							"description": "查找的最大深度，例如 1 表示只在当前目录查找，不进入子目录。",
						},
					},
					"required": []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "grep",
				Description: "在文件中搜索匹配指定模式的行。返回所有匹配的行。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pattern": map[string]any{
							"type":        "string",
							"description": "要搜索的正则表达式或字符串模式。",
						},
						"file": map[string]any{
							"type":        "string",
							"description": "要搜索的文件名或路径。如果是多个文件，可以是以空格分隔的字符串。",
						},
						"recursive": map[string]any{
							"type":        "boolean",
							"description": "如果为true，递归搜索目录下的文件，相当于 grep -r。",
						},
						"ignore_case": map[string]any{
							"type":        "boolean",
							"description": "如果为true，忽略大小写，相当于 grep -i。",
						},
						"count_only": map[string]any{
							"type":        "boolean",
							"description": "如果为true，只返回匹配行的数量，相当于 grep -c。",
						},
					},
					"required": []string{"pattern", "file"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "wget",
				Description: "从互联网下载文件。将文件保存到当前目录或指定路径。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"url": map[string]any{
							"type":        "string",
							"description": "要下载文件的 URL。",
						},
						"output_file": map[string]any{
							"type":        "string",
							"description": "可选。下载后文件的保存名称或路径。",
						},
					},
					"required": []string{"url"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "ss",
				Description: "显示套接字统计信息，用于查看网络连接。可以过滤特定端口或连接状态。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"options": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "ss 命令的选项，例如 '-l' (监听), '-t' (TCP), '-u' (UDP), '-n' (数字显示), '-p' (显示进程), '-a' (所有套接字)。",
						},
						"port": map[string]any{
							"type":        "integer",
							"description": "过滤指定端口的连接。",
						},
						"protocol": map[string]any{
							"type":        "string",
							"description": "过滤指定协议，例如 'tcp' 或 'udp'。",
						},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "lsof",
				Description: "列出打开的文件。可以查看进程打开的文件，或端口被哪个进程占用。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "查找打开指定文件的进程。",
						},
						"port": map[string]any{
							"type":        "integer",
							"description": "查找占用指定端口的进程。",
						},
						"user": map[string]any{
							"type":        "string",
							"description": "查找某个用户打开的文件。",
						},
						"options": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "lsof 命令的额外选项，例如 '-i' (列出所有网络文件)。",
						},
					},
					"required": []string{},
				},
			},
		},
	}
}
