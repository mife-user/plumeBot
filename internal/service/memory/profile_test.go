package memory

import (
	"context"
	"testing"

	"plumebot/internal/domain/entity"
)

// groupMsgBy 构造带发送者的群聊消息。
func groupMsgBy(groupID, userID, content string) entity.Message {
	return entity.Message{GroupID: groupID, UserID: userID, MessageType: "group", Content: content}
}

func TestProfileCacheLoadsOnFirstAppearance(t *testing.T) {
	store := &fakeStorage{}
	store.memberProf = map[string]*entity.MemberProfile{
		"g1|u1": {GroupID: "g1", UserID: "u1", Activity: 0.8, Intimacy: 0.5},
	}
	store.groupProf = map[string]*entity.GroupProfile{
		"g1": {GroupID: "g1", Culture: "轻松"},
	}
	p := NewProfileCache(store)

	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "hi"))

	prof, ok := p.GetMemberProfile("g1", "u1")
	if !ok || prof == nil || prof.Activity != 0.8 {
		t.Errorf("成员画像未加载: ok=%v prof=%+v", ok, prof)
	}
	gp, ok := p.GetGroupProfile("g1")
	if !ok || gp == nil || gp.Culture != "轻松" {
		t.Errorf("群画像未加载: ok=%v gp=%+v", ok, gp)
	}
}

func TestProfileCacheNoReloadOnReappear(t *testing.T) {
	store := &fakeStorage{}
	store.memberProf = map[string]*entity.MemberProfile{"g1|u1": {GroupID: "g1", UserID: "u1"}}
	store.groupProf = map[string]*entity.GroupProfile{"g1": {GroupID: "g1"}}
	p := NewProfileCache(store)

	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "a"))
	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "b"))

	if store.memberGets != 1 {
		t.Errorf("成员第二次出现不应重复查库，实际查询 %d 次", store.memberGets)
	}
	if store.groupGets != 1 {
		t.Errorf("群画像应只查一次，实际查询 %d 次", store.groupGets)
	}
}

func TestProfileCacheCachesAbsentMember(t *testing.T) {
	store := &fakeStorage{} // 无任何画像
	p := NewProfileCache(store)

	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "a"))
	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "b"))

	if store.memberGets != 1 {
		t.Errorf("确认无画像后不应重复查库，实际查询 %d 次", store.memberGets)
	}
	prof, ok := p.GetMemberProfile("g1", "u1")
	if !ok || prof != nil {
		t.Errorf("应缓存为「无画像」: ok=%v prof=%+v", ok, prof)
	}
}

func TestProfileCachePrivateMessageIgnored(t *testing.T) {
	store := &fakeStorage{}
	p := NewProfileCache(store)

	p.TouchMessage(context.Background(), entity.Message{UserID: "u1", MessageType: "private", Content: "hi"})

	if store.memberGets != 0 || store.groupGets != 0 {
		t.Errorf("私聊消息不应触发画像加载: member=%d group=%d", store.memberGets, store.groupGets)
	}
	if _, ok := p.GetMemberProfile("", "u1"); ok {
		t.Error("私聊不应缓存成员画像")
	}
}

func TestProfileCacheDelayedEviction(t *testing.T) {
	store := &fakeStorage{}
	store.memberProf = map[string]*entity.MemberProfile{"g1|u1": {GroupID: "g1", UserID: "u1"}}
	p := NewProfileCache(store)

	// u1 出现一次（seq=1），之后不再说话。
	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "hi"))

	// 其他用户刷 WindowCap + delay - 1 条：u1 仍在「窗口 + 延迟保护」内，不应淘汰。
	for i := 0; i < WindowCap+profileEvictDelay-1; i++ {
		p.TouchMessage(context.Background(), groupMsgBy("g1", "noise", "m"))
	}
	if _, ok := p.GetMemberProfile("g1", "u1"); !ok {
		t.Fatal("延迟轮数未满不应淘汰")
	}

	// 再来 1 条：达到淘汰条件。
	p.TouchMessage(context.Background(), groupMsgBy("g1", "noise", "m"))
	if _, ok := p.GetMemberProfile("g1", "u1"); ok {
		t.Error("延迟轮数满后应移出缓存")
	}
}

func TestProfileCacheReappearResetsEviction(t *testing.T) {
	store := &fakeStorage{}
	store.memberProf = map[string]*entity.MemberProfile{"g1|u1": {GroupID: "g1", UserID: "u1"}}
	p := NewProfileCache(store)

	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "hi"))
	// 逼近淘汰阈值但未满。
	for i := 0; i < WindowCap+profileEvictDelay-1; i++ {
		p.TouchMessage(context.Background(), groupMsgBy("g1", "noise", "m"))
	}
	// u1 再次出现，lastSeen 刷新。
	p.TouchMessage(context.Background(), groupMsgBy("g1", "u1", "again"))

	// 再刷满一窗口，u1 因 lastSeen 已刷新仍应保留。
	for i := 0; i < WindowCap+profileEvictDelay-1; i++ {
		p.TouchMessage(context.Background(), groupMsgBy("g1", "noise", "m"))
	}
	if _, ok := p.GetMemberProfile("g1", "u1"); !ok {
		t.Error("刷新 lastSeen 后不应被误淘汰")
	}
}
