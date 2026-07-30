package domain

import (
	"context"

	"plumebot/internal/domain/entity"
)

// Storage 持久化消息。
type Storage interface {
	SaveMessage(ctx context.Context, msg entity.Message) error
}
