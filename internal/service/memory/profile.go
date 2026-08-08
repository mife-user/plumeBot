package memory

import (
	"context"
	"errors"
	"sync"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
	"plumebot/pkg/logger"
)

// profileEvictDelay 是画像延迟淘汰的轮数（架构 §4.2「延迟 N 轮」）。
// 成员从窗口消失（其消息滚出 WindowCap 窗口）后，再经过这么多轮消息仍未出现则移出缓存。
const profileEvictDelay = 20

// cachedMember 是缓存中的个人画像条目。
type cachedMember struct {
	profile     *entity.MemberProfile // nil = 已查询但该群此人无画像（缓存"查过不存在"，避免重复查库）
	lastSeenSeq int64                 // 最后一次出现在消息中的会话序号
}

// ProfileCache 管理群聊个人画像与群画像的按需加载 + 内存缓存 + 延迟淘汰（架构 §4.2/§4.3）。
//
// 加载策略：
//   - 只加载窗口中出现过的成员（消息驱动）：成员首次发消息时从 SQLite 加载，之后走内存不重复查库；
//   - 群画像在群首次出现时加载一次；
//   - 成员从窗口消失（消息滚出窗口）后再延迟 profileEvictDelay 轮淘汰，避免频繁加载/卸载。
//
// 私聊消息不参与画像缓存（画像按群组织）。
type ProfileCache struct {
	store domain.Storage

	mu      sync.Mutex
	members map[string]map[string]*cachedMember // groupID -> (userID -> entry)
	groups  map[string]*entity.GroupProfile     // groupID -> 群画像（nil = 查询过但无画像）
	seqs    map[string]int64                    // groupID -> 消息计数（用于判断成员是否仍在窗口）
}

// NewProfileCache 创建画像缓存，注入 domain.Storage 用于按需加载。
func NewProfileCache(store domain.Storage) *ProfileCache {
	return &ProfileCache{
		store:   store,
		members: make(map[string]map[string]*cachedMember),
		groups:  make(map[string]*entity.GroupProfile),
		seqs:    make(map[string]int64),
	}
}

// TouchMessage 在每条群聊消息到达时更新画像缓存。私聊消息直接忽略。
func (p *ProfileCache) TouchMessage(ctx context.Context, msg entity.Message) {
	if msg.MessageType != "group" {
		return
	}
	gid := msg.GroupID

	p.mu.Lock()
	defer p.mu.Unlock()

	p.seqs[gid]++
	seq := p.seqs[gid]

	// 群画像：群首次出现时加载一次；加载失败（非 NotFound）不缓存，下条消息重试。
	if _, ok := p.groups[gid]; !ok {
		if prof, cached := p.loadGroupProfile(ctx, gid); cached {
			p.groups[gid] = prof
		}
	}

	// 成员画像：首次出现时加载，之后只刷新最近出现序号。
	ms, ok := p.members[gid]
	if !ok {
		ms = make(map[string]*cachedMember)
		p.members[gid] = ms
	}
	m, ok := ms[msg.UserID]
	if !ok {
		if prof, cached := p.loadMemberProfile(ctx, gid, msg.UserID); cached {
			m = &cachedMember{profile: prof}
			ms[msg.UserID] = m
		}
	}
	if m != nil {
		m.lastSeenSeq = seq
	}

	// 延迟淘汰：从窗口消失（lastSeen 早于「窗口保持范围 + 延迟轮数」）的成员移出缓存。
	threshold := seq - WindowCap - profileEvictDelay
	for uid, c := range ms {
		if c.lastSeenSeq <= threshold {
			delete(ms, uid)
		}
	}
}

// GetMemberProfile 返回缓存的个人画像；ok=false 表示尚未缓存（未在窗口出现或加载失败）。
// ok=true 时 profile 可能为 nil，表示已确认该群此人无画像。
func (p *ProfileCache) GetMemberProfile(groupID, userID string) (*entity.MemberProfile, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if ms, ok := p.members[groupID]; ok {
		if m, ok := ms[userID]; ok {
			return m.profile, true
		}
	}
	return nil, false
}

// GetGroupProfile 返回缓存的群画像；ok=false 表示尚未缓存。ok=true 时可能为 nil（已确认无画像）。
func (p *ProfileCache) GetGroupProfile(groupID string) (*entity.GroupProfile, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	prof, ok := p.groups[groupID]
	return prof, ok
}

// loadMemberProfile 查询个人画像；确认无画像（ErrNotFound）返回 nil 并标记可缓存。
// 查询失败（非 ErrNotFound）不缓存，返回 cached=false 以便下条消息重试。
func (p *ProfileCache) loadMemberProfile(ctx context.Context, groupID, userID string) (*entity.MemberProfile, bool) {
	prof, err := p.store.GetMemberProfile(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, true
		}
		logger.Warn("加载个人画像失败",
			logger.S("group_id", groupID), logger.S("user_id", userID), logger.Err(err))
		return nil, false
	}
	return prof, true
}

// loadGroupProfile 查询群画像；确认无画像（ErrNotFound）返回 nil 并标记可缓存。
// 查询失败（非 ErrNotFound）不缓存，返回 cached=false 以便下条消息重试。
func (p *ProfileCache) loadGroupProfile(ctx context.Context, groupID string) (*entity.GroupProfile, bool) {
	prof, err := p.store.GetGroupProfile(ctx, groupID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, true
		}
		logger.Warn("加载群画像失败",
			logger.S("group_id", groupID), logger.Err(err))
		return nil, false
	}
	return prof, true
}
