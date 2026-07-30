package persona

import (
	"context"

	"plumebot/internal/domain"
)

// PersonaService 负责人格 extend 链合并与缓存编排。
type PersonaService struct {
	persona domain.Persona
}

// NewPersonaService 创建 PersonaService，注入 domain.Persona 实现。
func NewPersonaService(persona domain.Persona) *PersonaService {
	return &PersonaService{persona: persona}
}

// GetEffectivePersona 获取 extend 链合并后的完整人格文本。第一阶段返回空。
func (s *PersonaService) GetEffectivePersona(_ context.Context, _, _ int64) (string, error) {
	return "", nil
}
