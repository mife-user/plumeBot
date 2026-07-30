package domain

import "context"

// Plugin 执行命名插件命令。
type Plugin interface {
	Execute(ctx context.Context, name string, args []string) (string, error)
}
