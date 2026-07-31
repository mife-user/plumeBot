// PlumeBot 唯一入口。第一阶段仅完成依赖注入组装，不连接任何外部服务。
package main

import (
	"log"

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

func main() {
	// 1. 加载配置
	_, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建 infra 实现
	agentInfra := &ai.AgentStub{}
	storageInfra, err := sqlite.Open("data")
	if err != nil {
		log.Fatalf("打开 SQLite 失败: %v", err)
	}
	defer storageInfra.Close()
	soPlugin := &plugin_so.PluginSOStub{}
	exePlugin := &plugin_exe.PluginEXEStub{}

	// 3. 注入 service
	agentSvc := agent.NewAgentService(agentInfra)
	memorySvc := memory.NewMemoryService(memory.Nop(), storageInfra)
	personaSvc := persona.NewPersonaService(persona.Nop())
	pluginSvc := plugin.NewPluginService(soPlugin, exePlugin)
	controlSvc := control.NewControlService(control.Nop())
	eventSvc := event.NewEventService(agentSvc, memorySvc, personaSvc, pluginSvc, controlSvc)

	// 4. 注入 handler
	_ = handler.NewMessageHandler(eventSvc)
	_ = handler.NewNoticeHandler(eventSvc)

	// 5. 启动
	log.Println("PlumeBot 已启动（第一阶段骨架）")
	select {}
}
