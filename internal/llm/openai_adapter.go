package llm

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient 是一个适配器，它实现了我们内部的 LLM 接口，
// 但在底层使用 go-openai SDK 来与兼容 OpenAI 的 API（如 DeepSeek）通信。
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient 创建一个新的 OpenAI 适配器客户端。
func NewOpenAIClient(apiKey, baseURL string) *OpenAIClient {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL // 使用传入的 BaseURL
	return &OpenAIClient{
		client: openai.NewClientWithConfig(config),
	}
}

// ChatCompletion 实现了 LLM 接口。它负责类型转换和调用底层 SDK。
func (c *OpenAIClient) ChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	// 1. 将内部请求类型转换为 go-openai 的请求类型
	sdkRequest := toOpenAIRequest(request)

	// 2. 调用 SDK
	sdkResponse, err := c.client.CreateChatCompletion(ctx, sdkRequest)
	if err != nil {
		return nil, err
	}

	// 3. 将 go-openai 的响应类型转换回内部响应类型
	response := fromOpenAIResponse(sdkResponse)
	return &response, nil
}

// toOpenAIRequest 将我们的内部 ChatRequest 转换为 go-openai 的类型。
func toOpenAIRequest(req ChatRequest) openai.ChatCompletionRequest {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		var content string
		if msg.Content != nil {
			content = *msg.Content
		}
		messages[i] = openai.ChatCompletionMessage{
			Role:       msg.Role,
			Content:    content,
			ToolCalls:  toOpenAIToolCalls(msg.ToolCalls),
			ToolCallID: msg.ToolCallID,
		}
	}

	tools := make([]openai.Tool, len(req.Tools))
	for i, t := range req.Tools {
		tools[i] = openai.Tool{
			Type: openai.ToolType(t.Type),
			Function: &openai.FunctionDefinition{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		}
	}

	return openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Tools:    tools,
	}
}

func toOpenAIToolCalls(calls []ToolCall) []openai.ToolCall {
	if calls == nil {
		return nil
	}
	sdkCalls := make([]openai.ToolCall, len(calls))
	for i, call := range calls {
		sdkCalls[i] = openai.ToolCall{
			ID:   call.ID,
			Type: openai.ToolType(call.Type),
			Function: openai.FunctionCall{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		}
	}
	return sdkCalls
}

// fromOpenAIResponse 将 go-openai 的响应转换为我们的内部 ChatResponse。
func fromOpenAIResponse(resp openai.ChatCompletionResponse) ChatResponse {
	choices := make([]Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		var content *string
		if choice.Message.Content != "" {
			content = &choice.Message.Content
		}

		choices[i] = Choice{
			Index: choice.Index,
			Message: Message{
				Role:      choice.Message.Role,
				Content:   content,
				ToolCalls: fromOpenAIToolCalls(choice.Message.ToolCalls),
			},
		}
	}

	return ChatResponse{
		ID:      resp.ID,
		Model:   resp.Model,
		Choices: choices,
	}
}

func fromOpenAIToolCalls(sdkCalls []openai.ToolCall) []ToolCall {
	if sdkCalls == nil {
		return nil
	}
	calls := make([]ToolCall, len(sdkCalls))
	for i, sdkCall := range sdkCalls {
		calls[i] = ToolCall{
			ID:   sdkCall.ID,
			Type: string(sdkCall.Type),
			Function: FunctionCall{
				Name:      sdkCall.Function.Name,
				Arguments: sdkCall.Function.Arguments,
			},
		}
	}
	return calls
}
