package llm

import "context"

// LLM 定义了与语言模型交互的客户端应遵循的通用接口。
// 这使得底层实现可以轻松切换（例如，从一个 API 提供商切换到另一个），而无需更改业务逻辑代码。
type LLM interface {
	// ChatCompletion 是与 LLM 进行对话的核心方法。
	// 它接收一个上下文和一个包含消息历史与可用工具的请求，
	// 然后返回 LLM 的响应。
	ChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error)
}
