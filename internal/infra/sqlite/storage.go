package sqlite

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// 编译期校验：StorageStub 实现 domain.Storage。
var _ domain.Storage = (*StorageStub)(nil)

// StorageStub 是 domain.Storage 的空壳实现，第一阶段不接入 SQLite。
type StorageStub struct{}

// SaveMessage 返回 nil，无实际操作。
func (s *StorageStub) SaveMessage(_ context.Context, _ entity.Message) error {
	return nil
}
