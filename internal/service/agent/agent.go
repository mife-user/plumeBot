package agent

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// AgentService 负责消息组装与 Agent 推理编排。
// 系统提示词已下沉至 infra 层（ChatModelAgentConfig.Instruction，由 provider 工厂
// 注入 cfg.Agent.SystemPrompt，空值兜底 config.DefaultSystemPrompt），本层只透传业务消息。
type AgentService struct {
	agent domain.Agent
}

// NewAgentService 创建 AgentService，注入 domain.Agent 实现。
func NewAgentService(agent domain.Agent) *AgentService {
	return &AgentService{agent: agent}
}

// GenerateReply 透传消息列表调用 Agent 生成回复。
func (s *AgentService) GenerateReply(ctx context.Context, msgs []entity.ChatMessage) (string, error) {
	return s.agent.Generate(ctx, msgs)
}
