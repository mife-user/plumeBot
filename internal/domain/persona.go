package domain

import (
	"context"

	"plumebot/internal/domain/entity"
)

// Persona 管理人格系统，返回 extend 链合并后的完整人格。
type Persona interface {
	Get(ctx context.Context, userID, groupID int64) (*entity.Persona, error)
}
