package memory

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// ---------------------------------------------------------------------------
// Nop 实现 —— 无操作默认实现，启动装配或测试时使用。
// ---------------------------------------------------------------------------

type nopMemory struct{}

var _ domain.Memory = (*nopMemory)(nil)

func (m *nopMemory) AppendMessage(_ context.Context, _ entity.Message) error        { return nil }
func (m *nopMemory) GetWindow(_ context.Context, _ string) ([]entity.Message, error) { return nil, nil }

// Nop 返回 domain.Memory 的无操作实现。
func Nop() domain.Memory { return &nopMemory{} }

// ---------------------------------------------------------------------------
// MemoryService
// ---------------------------------------------------------------------------

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
