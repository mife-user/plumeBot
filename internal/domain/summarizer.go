package domain

import "context"

// Summarizer 将一段长文本（对话历史）压缩为摘要。
// 与 Agent 分离：不带人设、不注册工具，是单次原始模型调用。
// 摘要 prompt 由调用方（service 层）组装，infra 只执行推理。
type Summarizer interface {
	// Summarize 执行一次摘要推理，返回模型输出文本（结构化解析由调用方负责）。
	Summarize(ctx context.Context, system string, user string) (string, error)
}
