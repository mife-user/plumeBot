// PlumeBot 唯一入口。第一阶段仅完成依赖注入组装，不连接任何外部服务。
package main

import (
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
	"plumebot/pkg/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Fatal("加载配置失败", logger.Err(err))
	}

	// 初始化日志
	logger.Init(logger.Config{Level: cfg.Log.Level})
	defer logger.Sync()

	// 2. 创建 infra 实现
	agentInfra := &ai.AgentStub{}
	storageInfra, err := sqlite.Open("data")
	if err != nil {
		logger.Fatal("打开 SQLite 失败", logger.Err(err))
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
	logger.Info("PlumeBot 已启动（第一阶段骨架）")

	logger.Info("bot 已启动",
		logger.S("name", cfg.Bot.Name),
	)

	select {}
}
