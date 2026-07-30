package domain

import (
	"context"

	"plumebot/internal/domain/entity"
)

// Control 判断 bot 在当前事件下是否应该回复。
type Control interface {
	ShouldReply(ctx context.Context, event entity.Event) (bool, error)
}
