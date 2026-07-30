package handler

import (
	"context"

	"plumebot/internal/domain/entity"
	"plumebot/internal/service/event"
)

// NoticeHandler 是通知事件的处理入口，将通知事件透传给 EventService。
type NoticeHandler struct {
	event *event.EventService
}

// NewNoticeHandler 创建 NoticeHandler，注入 EventService。
func NewNoticeHandler(svc *event.EventService) *NoticeHandler {
	return &NoticeHandler{event: svc}
}

// Handle 将通知透传给 EventService.HandleNotice。第一阶段无业务逻辑。
func (h *NoticeHandler) Handle(ctx context.Context, evt entity.Event) error {
	return h.event.HandleNotice(ctx, evt)
}
