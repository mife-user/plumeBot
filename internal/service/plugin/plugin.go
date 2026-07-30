package plugin

import (
	"context"

	"plumebot/internal/domain"
)

// PluginService 负责插件发现、加载与路由编排。
type PluginService struct {
	pluginSO  domain.Plugin
	pluginEXE domain.Plugin
}

// NewPluginService 创建 PluginService，注入两种 Plugin 实现（.so / exe）。
func NewPluginService(pluginSO, pluginEXE domain.Plugin) *PluginService {
	return &PluginService{pluginSO: pluginSO, pluginEXE: pluginEXE}
}

// Dispatch 根据命令名路由到对应插件并执行。第一阶段返回空。
func (s *PluginService) Dispatch(_ context.Context, _ string, _ []string) (string, error) {
	return "", nil
}
