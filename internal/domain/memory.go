package domain

import (
	"context"

	"plumebot/internal/domain/entity"
)

// Memory 管理短期上下文窗口：追加消息、获取窗口。
type Memory interface {
	AppendMessage(ctx context.Context, msg entity.Message) error
	GetWindow(ctx context.Context, groupID string) ([]entity.Message, error)
}
