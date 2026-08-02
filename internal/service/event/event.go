package event

import (
	"context"

	"plumebot/internal/domain/entity"
	"plumebot/internal/service/agent"
	"plumebot/internal/service/control"
	"plumebot/internal/service/memory"
	"plumebot/internal/service/persona"
	"plumebot/internal/service/plugin"
	"plumebot/pkg/config"
)

// EventService 是顶层事件调度编排，组合所有子 service。
type EventService struct {
	agent    *agent.AgentService
	memory   *memory.MemoryService
	persona  *persona.PersonaService
	plugin   *plugin.PluginService
	control  *control.ControlService
	msgChain Handler
}

// NewEventService 创建 EventService，注入全部子 service 与中间件配置。
// 消息管线顺序固定为：日志 → 限流 → 敏感词 → 末端处理。
func NewEventService(
	agent *agent.AgentService,
	memory *memory.MemoryService,
	persona *persona.PersonaService,
	plugin *plugin.PluginService,
	control *control.ControlService,
	mwCfg config.MiddlewareConfig,
) *EventService {
	s := &EventService{
		agent:   agent,
		memory:  memory,
		persona: persona,
		plugin:  plugin,
		control: control,
	}
	mws := []Middleware{
		logMiddleware,
		rateLimitMiddleware(newRateLimiter(mwCfg.RateLimit)),
		sensitiveWordMiddleware,
	}
	s.msgChain = chain(mws, tailHandler)
	return s
}

// HandleMessage 处理消息事件：走「日志 → 限流 → 敏感词 → 末端」管线。
func (s *EventService) HandleMessage(ctx context.Context, msg entity.Message) error {
	return s.msgChain(ctx, msg)
}

// HandleNotice 处理通知事件。第一阶段返回 nil。
func (s *EventService) HandleNotice(_ context.Context, _ entity.Event) error {
	return nil
}
