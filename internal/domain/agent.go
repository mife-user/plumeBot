package domain

import "context"

// Agent 定义 AI 推理能力——接收组装好的 prompt，返回文本回复。
type Agent interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
