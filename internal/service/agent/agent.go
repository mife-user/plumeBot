package agent

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
	"plumebot/pkg/config"
)

// AgentService 负责 prompt 组装与 Agent 推理编排。
type AgentService struct {
	agent        domain.Agent
	systemPrompt string // 机器人系统提示词（main 注入 cfg.Prompt.System；空则兜底默认常量）
}

// NewAgentService 创建 AgentService，注入 domain.Agent 实现与系统提示词。
func NewAgentService(agent domain.Agent, systemPrompt string) *AgentService {
	return &AgentService{agent: agent, systemPrompt: systemPrompt}
}

// GenerateReply 组装完整消息列表（system 前置 + 追加消息）并调用 Agent 生成回复。
func (s *AgentService) GenerateReply(ctx context.Context, msgs []entity.ChatMessage) (string, error) {
	system := s.systemPrompt
	if system == "" {
		system = config.DefaultSystemPrompt
	}

	all := make([]entity.ChatMessage, 0, len(msgs)+1)
	all = append(all, entity.ChatMessage{
		Role:  entity.RoleSystem,
		Parts: []entity.ContentPart{{Type: entity.PartTypeText, Text: system}},
	})
	all = append(all, msgs...)

	return s.agent.Generate(ctx, all)
}
