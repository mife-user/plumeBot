package ai

import (
	"context"

	"plumebot/internal/domain"
)

// 编译期校验：AgentStub 实现 domain.Agent。
var _ domain.Agent = (*AgentStub)(nil)

// AgentStub 是 domain.Agent 的空壳实现，第一阶段不接入 eino。
type AgentStub struct{}

// Generate 返回空响应，无错误。
func (s *AgentStub) Generate(_ context.Context, _ string) (string, error) {
	return "", nil
}
