package plugin_so

import (
	"context"

	"plumebot/internal/domain"
)

// 编译期校验：PluginSOStub 实现 domain.Plugin。
var _ domain.Plugin = (*PluginSOStub)(nil)

// PluginSOStub 是 domain.Plugin 的 .so 加载版本空壳，第一阶段不实现。
type PluginSOStub struct{}

// Execute 返回空响应，无错误。
func (s *PluginSOStub) Execute(_ context.Context, _ string, _ []string) (string, error) {
	return "", nil
}
