package history

import "github.com/DoraZa/mini-agent/internal/llm"

// HistoryManager 管理对话历史记录
type HistoryManager struct {
	Messages []llm.Message
}

// NewHistoryManager 创建一个新的历史记录管理器
func NewHistoryManager() *HistoryManager {
	return &HistoryManager{
		Messages: make([]llm.Message, 0),
	}
}

// AddUserMessage 添加一条用户消息到历史记录
func (h *HistoryManager) AddUserMessage(content string) {
	h.Messages = append(h.Messages, llm.Message{
		Role:    "user",
		Content: &content,
	})
}

// AddSystemMessage 添加一条系统消息到历史记录
func (h *HistoryManager) AddSystemMessage(content string) {
	h.Messages = append(h.Messages, llm.Message{
		Role:    "system",
		Content: &content,
	})
}

// AddAssistantMessage 添加一条完整的助手消息到历史记录
// 这条消息可以包含思考过程（content）和工具调用
func (h *HistoryManager) AddAssistantMessage(message llm.Message) {
	// 确保角色是 assistant，以防万一
	message.Role = "assistant"
	h.Messages = append(h.Messages, message)
}

// AddToolObservation 添加一条工具执行结果（观察）到历史记录
func (h *HistoryManager) AddToolObservation(toolCallID, content string) {
	h.Messages = append(h.Messages, llm.Message{
		Role:       "tool",
		ToolCallID: toolCallID,
		Content:    &content,
	})
}

// GetHistory 返回当前所有的消息历史
func (h *HistoryManager) GetHistory() []llm.Message {
	return h.Messages
}
