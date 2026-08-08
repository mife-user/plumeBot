package entity

// Summary 表示一段对话历史的压缩摘要（P3-003 窗口压缩产出）。
// ChatID 为会话键（群聊=GroupID，私聊="private:"+UserID），与窗口一致。
// Seq 为会话内递增序号，既是排序依据，也是归档落库的幂等键（(chat_id, seq) 唯一）。
type Summary struct {
	ChatID    string   // 会话键
	Seq       int64    // 会话内递增序号
	Text      string   // 摘要文本
	Keywords  []string // 关键词标签（索引/召回用）
	Decisions []string // 关键决定/共识
	CreatedAt int64    // 生成时间（Unix 秒）
}
