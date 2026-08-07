package onebot

import (
	"testing"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"plumebot/internal/domain/entity"
)

func TestFormatID(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"int64 非零", int64(123), "123"},
		{"int64 零值", int64(0), ""},
		{"string", "abc", "abc"},
		{"空字符串", "", ""},
		{"其它类型", 3.14, ""},
		{"nil", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatID(tc.in); got != tc.want {
				t.Errorf("formatID(%#v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestToMessageGroup(t *testing.T) {
	ev := &zero.Event{
		PostType:    "message",
		MessageType: "group",
		MessageID:   int64(1001),
		GroupID:     int64(2002),
		UserID:      int64(3003),
		Time:        1700000000,
		Message:     message.Message{message.Text("你好呀")},
	}
	m, ok := toMessage(ev)
	if !ok {
		t.Fatal("群聊消息应被接受")
	}
	if m.MessageID != "1001" || m.GroupID != "2002" || m.UserID != "3003" {
		t.Errorf("ID 转换错误: %+v", m)
	}
	if m.Content != "你好呀" {
		t.Errorf("Content = %q, want %q", m.Content, "你好呀")
	}
	if m.Timestamp != 1700000000 || m.MessageType != "group" {
		t.Errorf("Timestamp/MessageType 错误: %+v", m)
	}
}

func TestToMessagePrivateGroupIDEmpty(t *testing.T) {
	ev := &zero.Event{
		PostType:    "message",
		MessageType: "private",
		MessageID:   int64(1001),
		UserID:      int64(3003),
		Message:     message.Message{message.Text("hi")},
	}
	m, ok := toMessage(ev)
	if !ok {
		t.Fatal("私聊消息应被接受")
	}
	if m.GroupID != "" {
		t.Errorf("私聊 GroupID 应为空，实际 %q", m.GroupID)
	}
}

func TestToMessageRejects(t *testing.T) {
	// 非 message 事件
	if _, ok := toMessage(&zero.Event{PostType: "notice"}); ok {
		t.Error("notice 事件不应被 toMessage 接受")
	}
	// 不支持的 message_type
	if _, ok := toMessage(&zero.Event{PostType: "message", MessageType: "guild"}); ok {
		t.Error("guild 消息不应被接受")
	}
}

func TestToEventNotice(t *testing.T) {
	raw := `{"notice_type":"group_increase"}`
	ev := &zero.Event{
		PostType:   "notice",
		DetailType: "group_increase",
		GroupID:    int64(1),
		UserID:     int64(2),
		Time:       1700000001,
		RawEvent:   gjson.Result{Raw: raw},
	}
	e, ok := toEvent(ev)
	if !ok {
		t.Fatal("notice 事件应被接受")
	}
	if e.Type != entity.EventNotice {
		t.Errorf("Type = %v, want EventNotice", e.Type)
	}
	if e.SubType != "group_increase" || e.GroupID != "1" || e.UserID != "2" || e.Timestamp != 1700000001 {
		t.Errorf("字段转换错误: %+v", e)
	}
	if e.RawJSON != raw {
		t.Errorf("RawJSON = %q, want %q", e.RawJSON, raw)
	}
}

func TestToEventTypesAndReject(t *testing.T) {
	if e, ok := toEvent(&zero.Event{PostType: "request"}); !ok || e.Type != entity.EventRequest {
		t.Errorf("request 事件应转为 EventRequest，got %+v", e)
	}
	if e, ok := toEvent(&zero.Event{PostType: "meta_event"}); !ok || e.Type != entity.EventMeta {
		t.Errorf("meta_event 事件应转为 EventMeta，got %+v", e)
	}
	if _, ok := toEvent(&zero.Event{PostType: "message"}); ok {
		t.Error("message 事件不应被 toEvent 接受")
	}
}
