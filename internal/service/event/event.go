package event

import (
	"context"

	"plumebot/internal/domain/entity"
	"plumebot/internal/service/agent"
	"plumebot/internal/service/control"
	"plumebot/internal/service/memory"
	"plumebot/internal/service/persona"
	"plumebot/internal/service/plugin"
)

// EventService 是顶层事件调度编排，组合所有子 service。
type EventService struct {
	agent   *agent.AgentService
	memory  *memory.MemoryService
	persona *persona.PersonaService
	plugin  *plugin.PluginService
	control *control.ControlService
}

// NewEventService 创建 EventService，注入全部子 service。
func NewEventService(
	agent *agent.AgentService,
	memory *memory.MemoryService,
	persona *persona.PersonaService,
	plugin *plugin.PluginService,
	control *control.ControlService,
) *EventService {
	return &EventService{
		agent:   agent,
		memory:  memory,
		persona: persona,
		plugin:  plugin,
		control: control,
	}
}

// HandleMessage 处理消息事件。第一阶段返回 nil。
func (s *EventService) HandleMessage(_ context.Context, _ entity.Message) error {
	return nil
}

// HandleNotice 处理通知事件。第一阶段返回 nil。
func (s *EventService) HandleNotice(_ context.Context, _ entity.Event) error {
	return nil
}
