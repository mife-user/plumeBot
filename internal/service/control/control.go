package control

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// ---------------------------------------------------------------------------
// Nop 实现
// ---------------------------------------------------------------------------

type nopControl struct{}

var _ domain.Control = (*nopControl)(nil)

func (c *nopControl) ShouldReply(_ context.Context, _ entity.Event) (bool, error) { return false, nil }

// Nop 返回 domain.Control 的无操作实现。
func Nop() domain.Control { return &nopControl{} }

// ---------------------------------------------------------------------------
// ControlService
// ---------------------------------------------------------------------------

// ControlService 负责触发判断与状态规则编排。
type ControlService struct {
	control domain.Control
}

// NewControlService 创建 ControlService，注入 domain.Control 实现。
func NewControlService(control domain.Control) *ControlService {
	return &ControlService{control: control}
}

// ShouldReply 判断 bot 在当前事件下是否应该回复。第一阶段返回 false。
func (s *ControlService) ShouldReply(_ context.Context, _ entity.Event) (bool, error) {
	return false, nil
}
