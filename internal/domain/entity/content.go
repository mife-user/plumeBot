package entity

// Role 表示对话消息的发送者角色。
type Role string

const (
	RoleSystem    Role = "system"    // 系统提示词
	RoleUser      Role = "user"      // 用户输入
	RoleAssistant Role = "assistant" // 模型回复（历史对话回传时使用）
)

// PartType 表示消息内容片段的类型。
type PartType string

const (
	PartTypeText  PartType = "text"
	PartTypeImage PartType = "image"
	PartTypeAudio PartType = "audio"
	PartTypeVideo PartType = "video"
	PartTypeFile  PartType = "file"
)

// ContentPart 是 ChatMessage 的一个内容片段。
// URL 与 Base64 二选一（按 Type 决定语义）；MIMEType 用于图片等二进制片段（如 image/png）。
type ContentPart struct {
	Type     PartType
	Text     string // Type == PartTypeText 时的文本内容
	URL      string // 远程资源地址（如图片 URL）
	Base64   string // 二进制内容（base64 编码）
	MIMEType string // 二进制片段媒体类型
}

// ChatMessage 是传给 LLM 的一条会话消息（多模态：Parts 可含文本、图片等片段）。
// 完整消息列表（含 system role）由调用方（service 层）组装，infra 层只执行。
type ChatMessage struct {
	Role  Role
	Parts []ContentPart
}
