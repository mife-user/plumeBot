// PlumeBot 唯一入口。第一阶段仅完成依赖注入组装，不连接任何外部服务。
package main

import (
	"context"
	"log"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
	"plumebot/internal/handler"
	"plumebot/internal/infra/ai"
	"plumebot/internal/infra/plugin_exe"
	"plumebot/internal/infra/plugin_so"
	"plumebot/internal/infra/sqlite"
	"plumebot/internal/service/agent"
	"plumebot/internal/service/control"
	"plumebot/internal/service/event"
	"plumebot/internal/service/memory"
	"plumebot/internal/service/persona"
	"plumebot/internal/service/plugin"
	"plumebot/pkg/config"
)

// ---------------------------------------------------------------------------
// 第一阶段临时 stub —— Memory / Persona / Control 由 service 层编排，
// infra 层不做独立实现，在此直接定义空壳。
// ---------------------------------------------------------------------------

type memoryStub struct{}

var _ domain.Memory = (*memoryStub)(nil)

func (m *memoryStub) AppendMessage(_ context.Context, _ entity.Message) error      { return nil }
func (m *memoryStub) GetWindow(_ context.Context, _ string) ([]entity.Message, error) { return nil, nil }

type personaStub struct{}

var _ domain.Persona = (*personaStub)(nil)

func (p *personaStub) Get(_ context.Context, _, _ int64) (*entity.Persona, error) { return nil, nil }

type controlStub struct{}

var _ domain.Control = (*controlStub)(nil)

func (c *controlStub) ShouldReply(_ context.Context, _ entity.Event) (bool, error) { return false, nil }

func main() {
	// 1. 加载配置
	_, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建 infra（空壳实现）
	agentInfra := &ai.AgentStub{}
	storageInfra := &sqlite.StorageStub{}
	pluginSOInfra := &plugin_so.PluginSOStub{}
	pluginEXEInfra := &plugin_exe.PluginEXEStub{}
	memStub := &memoryStub{}
	perStub := &personaStub{}
	ctrlStub := &controlStub{}

	// 3. 注入 service
	agentSvc := agent.NewAgentService(agentInfra)
	memorySvc := memory.NewMemoryService(memStub, storageInfra)
	personaSvc := persona.NewPersonaService(perStub)
	pluginSvc := plugin.NewPluginService(pluginSOInfra, pluginEXEInfra)
	controlSvc := control.NewControlService(ctrlStub)
	eventSvc := event.NewEventService(agentSvc, memorySvc, personaSvc, pluginSvc, controlSvc)

	// 4. 注入 handler
	_ = handler.NewMessageHandler(eventSvc)
	_ = handler.NewNoticeHandler(eventSvc)

	// 5. 启动
	log.Println("PlumeBot 已启动（第一阶段骨架）")
	select {}
}
