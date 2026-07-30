package agent

import (
	"context"

	"plumebot/internal/domain"
)

// AgentService 负责 prompt 组装与 Agent 推理编排。
type AgentService struct {
	agent domain.Agent
}

// NewAgentService 创建 AgentService，注入 domain.Agent 实现。
func NewAgentService(agent domain.Agent) *AgentService {
	return &AgentService{agent: agent}
}

// GenerateReply 组装 prompt 并调用 Agent 生成回复。第一阶段返回空。
func (s *AgentService) GenerateReply(_ context.Context, _ string) (string, error) {
	return "", nil
}
