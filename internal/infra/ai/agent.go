package ai

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// 编译期校验：AgentStub 实现 domain.Agent。
var _ domain.Agent = (*AgentStub)(nil)

// AgentStub 是 domain.Agent 的空壳实现，阶段 4 接入 eino。
type AgentStub struct{}

// Generate 返回空响应，无错误。
func (s *AgentStub) Generate(_ context.Context, _ []entity.ChatMessage) (string, error) {
	return "", nil
}
