package onebot

import (
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"

	"plumebot/internal/domain/entity"
)

// toMessage 将 ZeroBot 消息事件转换为领域层 entity.Message。
// 仅接受群聊/私聊消息事件，其余返回 false。
func toMessage(ev *zero.Event) (entity.Message, bool) {
	if ev.PostType != "message" {
		return entity.Message{}, false
	}
	if ev.MessageType != "group" && ev.MessageType != "private" {
		return entity.Message{}, false
	}
	return entity.Message{
		MessageID:   formatID(ev.MessageID),
		GroupID:     formatID(ev.GroupID),
		UserID:      formatID(ev.UserID),
		Content:     ev.Message.ExtractPlainText(),
		Timestamp:   ev.Time,
		MessageType: ev.MessageType,
	}, true
}

// toEvent 将 ZeroBot 通知/请求/元事件转换为领域层 entity.Event。
func toEvent(ev *zero.Event) (entity.Event, bool) {
	var t entity.EventType
	switch ev.PostType {
	case "notice":
		t = entity.EventNotice
	case "request":
		t = entity.EventRequest
	case "meta_event":
		t = entity.EventMeta
	default:
		return entity.Event{}, false
	}
	return entity.Event{
		Type:      t,
		SubType:   ev.DetailType,
		GroupID:   formatID(ev.GroupID),
		UserID:    formatID(ev.UserID),
		Timestamp: ev.Time,
		RawJSON:   ev.RawEvent.Raw,
	}, true
}

// formatID 将 ZeroBot 的 int64/string 类型 ID 统一转为字符串。
// 零值（0/空）归一化为空字符串，保证非群/无 ID 场景下实体字段为空。
func formatID(v any) string {
	switch id := v.(type) {
	case int64:
		if id == 0 {
			return ""
		}
		return strconv.FormatInt(id, 10)
	case string:
		return id
	default:
		return ""
	}
}
