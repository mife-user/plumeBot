package entity

// BotState 表示 bot 在某个群的状态，以 JSON blob 存储。
// 每个群一条记录。
type BotState struct {
	GroupID string // 群 ID
	State   string // 状态 JSON（精力值、连续回复计数、冷却时间等）
}
