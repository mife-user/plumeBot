package entity

// MemberProfile 表示某个用户在某个群里的个人画像。
// 按 (GroupID, UserID) 唯一。
// 事实记忆和兴趣话题通过 member_facts 表查询，不在此处冗余存储。
type MemberProfile struct {
	GroupID  string  // 群 ID
	UserID   string  // 用户 QQ 号
	Activity float64 // 活跃度统计（0~1）
	Intimacy float64 // 与 bot 的亲密度（0~1）
}

// GroupProfile 表示一个群的群聊画像，每个群一条。
// 黑话词典通过 group_jargon 表查询，不在此处冗余存储。
type GroupProfile struct {
	GroupID     string   // 群 ID
	Culture     string   // 群文化特征描述
	Topics      []string // 主流话题标签
	ActiveHours string   // 活跃时段描述（如"晚上8-11点"）
	Rules       []string // 群规摘要
	Atmosphere  []string // 氛围标签（如"轻松"、"技术向"）
}

// Persona 表示一条人格定义，支持 extends 继承链。
// groupid=0 为全局默认人格（父级），groupid!=0 为群专属人格（子级）。
type Persona struct {
	ID      int64    // 人格 ID
	UserID  int64    // 所属用户 ID
	GroupID int64    // 群 ID，0 表示全局默认
	Extend  int64    // 继承的父级 Persona.ID，0 表示无继承
	Traits  []string // 性格标签（如"幽默"、"毒舌"、"话少"）
}
