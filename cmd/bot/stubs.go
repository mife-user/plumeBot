package main

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// ---------------------------------------------------------------------------
// 第一阶段临时 stub —— Memory / Persona / Control 由 service 层编排，
// infra 层不做独立实现，main 包内联定义用于启动装配。
// ---------------------------------------------------------------------------

type memoryStub struct{}

var _ domain.Memory = (*memoryStub)(nil)

func (m *memoryStub) AppendMessage(_ context.Context, _ entity.Message) error        { return nil }
func (m *memoryStub) GetWindow(_ context.Context, _ string) ([]entity.Message, error) { return nil, nil }

type personaStub struct{}

var _ domain.Persona = (*personaStub)(nil)

func (p *personaStub) Get(_ context.Context, _, _ int64) (*entity.Persona, error) { return nil, nil }

type controlStub struct{}

var _ domain.Control = (*controlStub)(nil)

func (c *controlStub) ShouldReply(_ context.Context, _ entity.Event) (bool, error) { return false, nil }
