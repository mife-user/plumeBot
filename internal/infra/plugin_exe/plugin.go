package plugin_exe

import (
	"context"

	"plumebot/internal/domain"
)

// 编译期校验：PluginEXEStub 实现 domain.Plugin。
var _ domain.Plugin = (*PluginEXEStub)(nil)

// PluginEXEStub 是 domain.Plugin 的 exe 子进程版本空壳，第一阶段不实现。
type PluginEXEStub struct{}

// Execute 返回空响应，无错误。
func (s *PluginEXEStub) Execute(_ context.Context, _ string, _ []string) (string, error) {
	return "", nil
}
