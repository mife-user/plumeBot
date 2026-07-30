package handler

import (
	"context"

	"plumebot/internal/domain/entity"
	"plumebot/internal/service/event"
)

// MessageHandler 是消息事件的 HTTP/WebSocket 无关处理入口，
// 将消息事件透传给 EventService。
type MessageHandler struct {
	event *event.EventService
}

// NewMessageHandler 创建 MessageHandler，注入 EventService。
func NewMessageHandler(svc *event.EventService) *MessageHandler {
	return &MessageHandler{event: svc}
}

// Handle 将消息透传给 EventService.HandleMessage。第一阶段无业务逻辑。
func (h *MessageHandler) Handle(ctx context.Context, msg entity.Message) error {
	return h.event.HandleMessage(ctx, msg)
}
