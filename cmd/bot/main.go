// PlumeBot 唯一入口。第二阶段：接入 ZeroBot 连接 NapCat，接收事件并分发。
package main

import (
	"context"
	"strconv"

	"plumebot/internal/handler"
	"plumebot/internal/infra/ai"
	"plumebot/internal/infra/onebot"
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
	ctx := context.Background()
	// LLM 注册中心：本期仅注册 openai（OpenAI 兼容接口）provider；
	// 工具注册表为空（仅机制，P3-004/P4-001 起注册具体工具）。
	toolsRegistry := ai.NewToolsRegistry()
	llmRegistry := ai.NewRegistry()
	if err := llmRegistry.Register(config.DefaultLLMProvider, ai.NewOpenAIFactory(toolsRegistry)); err != nil {
		logger.Fatal("注册 LLM provider 失败", logger.Err(err))
	}
	agentInfra, err := llmRegistry.NewAgent(ctx, *cfg)
	if err != nil {
		logger.Fatal("初始化 LLM Agent 失败", logger.Err(err))
	}
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
	eventSvc := event.NewEventService(agentSvc, memorySvc, personaSvc, pluginSvc, controlSvc, cfg.Middleware)

	// 4. 注入 handler
	msgHandler := handler.NewMessageHandler(eventSvc)
	noticeHandler := handler.NewNoticeHandler(eventSvc)

	// 5. 启动 onebot 连接（阻塞，ZeroBot 底层自动重连）
	client := onebot.New(cfg.Onebot, cfg.Log.Level, msgHandler, noticeHandler)
	if cfg.Bot.Name == "" {
		cfg.Bot.Name = config.DefaultBotName // 展示用兜底
	}
	if cfg.Onebot.WsURL == "" {
		cfg.Onebot.WsURL = config.DefaultWsURL // 与 onebot.New 内部兜底保持一致，保证日志显示真实连接地址
	}
	logger.Info("PlumeBot 启动，正在连接 NapCat",
		logger.S("name", cfg.Bot.Name),
		logger.S("ws_url", cfg.Onebot.WsURL),
		logger.S("llm_provider", llmProviderName(cfg)),
		logger.S("llm_model", cfg.LLM.OpenAI.Model),
		logger.S("llm_base_url", llmBaseURL(cfg)),
		logger.S("llm_api_key_set", strconv.FormatBool(cfg.LLM.OpenAI.APIKey != "")), // 只标记是否配置，绝不打印密钥本身
	)
	client.Run()
}

// llmProviderName 返回实际生效的 provider 名（与 Registry.NewAgent 的兜底一致）。
func llmProviderName(cfg *config.Config) string {
	if cfg.LLM.Provider == "" {
		return config.DefaultLLMProvider
	}
	return cfg.LLM.Provider
}

// llmBaseURL 返回实际生效的 base_url（与 openai 工厂的兜底一致，保证日志显示真实端点）。
func llmBaseURL(cfg *config.Config) string {
	if cfg.LLM.OpenAI.BaseURL == "" {
		return config.DefaultOpenAIBaseURL
	}
	return cfg.LLM.OpenAI.BaseURL
}
