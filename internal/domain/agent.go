package domain

import (
	"context"

	"plumebot/internal/domain/entity"
)

// Agent 定义 AI 推理能力——接收组装好的完整消息列表（含 system role），返回文本回复。
// 消息组装（prompt 构建、历史对话、人格注入）在 service 层完成，infra 只执行推理。
type Agent interface {
	Generate(ctx context.Context, messages []entity.ChatMessage) (string, error)
}
