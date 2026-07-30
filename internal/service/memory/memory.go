package memory

import (
	"context"

	"plumebot/internal/domain"
)

// MemoryService 负责上下文窗口维护与画像缓存编排。
type MemoryService struct {
	memory domain.Memory
	store  domain.Storage
}

// NewMemoryService 创建 MemoryService，注入 domain.Memory 和 domain.Storage。
func NewMemoryService(memory domain.Memory, store domain.Storage) *MemoryService {
	return &MemoryService{memory: memory, store: store}
}

// BuildContext 组装上下文窗口内容。第一阶段返回空。
func (s *MemoryService) BuildContext(_ context.Context, _ string) (string, error) {
	return "", nil
}
