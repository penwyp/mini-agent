package llm

// Message 定义了聊天消息的通用结构。
// 它可以代表用户、助手、系统或工具等多种角色。
type Message struct {
	Role       string     `json:"role"`
	Content    *string    `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// Tool 表示模型可以调用的工具的通用定义。
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function 定义了工具函数的通用结构。
type Function struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters"`
}

// ToolCall 代表模型请求调用一个具体的工具。
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 包含了要调用的函数名和参数。
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // 参数是一个 JSON 字符串
}

// ChatRequest 定义了通用的聊天请求结构。
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

// ChatResponse 定义了通用的聊天响应结构。
type ChatResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// Choice 是 API 返回的响应选项之一。
type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"` // 模型生成的消息
}
