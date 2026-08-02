package entity

// EventType 枚举 OneBot 事件大类。
type EventType string

const (
	EventMessage EventType = "message" // 消息事件
	EventNotice  EventType = "notice"  // 通知事件（群增减、戳一戳等）
	EventRequest EventType = "request" // 请求事件（加好友、加群）
	EventMeta    EventType = "meta"    // 元事件（心跳、生命周期）
)

// Event 表示一条 OneBot 事件的通用结构。
type Event struct {
	Type      EventType // 事件大类
	SubType   string    // 事件子类型，如 group_message、private_message
	GroupID   string    // 来源群 ID（非群事件时为空）
	UserID    string    // 来源用户 QQ 号
	Timestamp int64     // Unix 时间戳（秒）
	RawJSON   string    // 原始事件 JSON，供通知规则/插件订阅使用
}
