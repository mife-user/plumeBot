package memory

import (
	"context"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// ---------------------------------------------------------------------------
// MemoryService
// ---------------------------------------------------------------------------

// MemoryService 负责上下文窗口维护与画像缓存编排。
type MemoryService struct {
	memory   domain.Memory
	store    domain.Storage
	profiles *ProfileCache
}

// NewMemoryService 创建 MemoryService，注入 domain.Memory（窗口实现）和 domain.Storage。
func NewMemoryService(memory domain.Memory, store domain.Storage) *MemoryService {
	return &MemoryService{memory: memory, store: store, profiles: NewProfileCache(store)}
}

// PersistMessage 持久化一条消息：写入上下文窗口（内存 ring buffer）+ SQLite messages 表，
// 并触达画像缓存（窗口内成员按需加载）。返回窗口是否达到压缩阈值（P3-003 消费该信号触发一级压缩）。
func (s *MemoryService) PersistMessage(ctx context.Context, msg entity.Message) (bool, error) {
	full, err := s.memory.AppendMessage(ctx, msg)
	if err != nil {
		return false, err
	}
	if err := s.store.SaveMessage(ctx, msg); err != nil {
		return false, err
	}
	s.profiles.TouchMessage(ctx, msg)
	return full, nil
}

// GetWindow 返回会话窗口消息（会话键：群聊=GroupID，私聊="private:"+UserID）。
func (s *MemoryService) GetWindow(ctx context.Context, sessionID string) ([]entity.Message, error) {
	return s.memory.GetWindow(ctx, sessionID)
}

// GetMemberProfile 返回缓存的成员画像（P6 prompt 组装时使用）。
// ok=false 表示未在窗口出现/未加载；ok=true 时 profile 可能为 nil（确认无画像）。
func (s *MemoryService) GetMemberProfile(groupID, userID string) (*entity.MemberProfile, bool) {
	return s.profiles.GetMemberProfile(groupID, userID)
}

// GetGroupProfile 返回缓存的群画像。
func (s *MemoryService) GetGroupProfile(groupID string) (*entity.GroupProfile, bool) {
	return s.profiles.GetGroupProfile(groupID)
}

// BuildContext 组装上下文窗口内容。第一阶段返回空。
func (s *MemoryService) BuildContext(_ context.Context, _ string) (string, error) {
	return "", nil
}
