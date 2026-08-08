package memory

import (
	"context"
	"os"
	"testing"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
	"plumebot/pkg/logger"
)

// TestMain 初始化全局 logger（error 级，避免画像加载失败日志触发 nil 指针）。
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "plumebot-test-logs")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	logger.Init(logger.Config{Level: "error", Dir: dir})
	os.Exit(m.Run())
}

// fakeStorage 嵌入 domain.Storage 接口，实现 PersistMessage / ProfileCache 用到的几个方法。
// 其余方法嵌入 nil 接口，调用即 panic（但本测试不会调用）。
type fakeStorage struct {
	domain.Storage
	saved      []entity.Message
	memberProf map[string]*entity.MemberProfile // key = groupID + "|" + userID
	groupProf  map[string]*entity.GroupProfile
	memberGets int // 个人画像查询次数（断言不重复查库）
	groupGets  int // 群画像查询次数
}

func (f *fakeStorage) SaveMessage(_ context.Context, msg entity.Message) error {
	f.saved = append(f.saved, msg)
	return nil
}

func (f *fakeStorage) GetMemberProfile(_ context.Context, groupID, userID string) (*entity.MemberProfile, error) {
	f.memberGets++
	if p, ok := f.memberProf[groupID+"|"+userID]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}

func (f *fakeStorage) GetGroupProfile(_ context.Context, groupID string) (*entity.GroupProfile, error) {
	f.groupGets++
	if p, ok := f.groupProf[groupID]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}

func TestPersistMessageWritesWindowAndStorage(t *testing.T) {
	store := &fakeStorage{}
	svc := NewMemoryService(NewWindow(), store)
	msg := groupMsg("g1", "hi")

	full, err := svc.PersistMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("持久化失败: %v", err)
	}
	if full {
		t.Error("首条消息不应触发压缩")
	}
	if len(store.saved) != 1 || store.saved[0].Content != "hi" {
		t.Errorf("SQLite 未收到消息: %+v", store.saved)
	}
	got, _ := svc.GetWindow(context.Background(), "g1")
	if len(got) != 1 {
		t.Errorf("窗口应包含消息: %+v", got)
	}
	// 持久化同时触达画像缓存：群画像被缓存（无画像也缓存为已查询）。
	if _, ok := svc.GetGroupProfile("g1"); !ok {
		t.Error("持久化应触达群画像缓存")
	}
}

func TestPersistMessageSignalsCompressionAtCap(t *testing.T) {
	store := &fakeStorage{}
	svc := NewMemoryService(NewWindow(), store)
	for i := 0; i < WindowCap; i++ {
		full, err := svc.PersistMessage(context.Background(), groupMsg("g1", "m"))
		if err != nil {
			t.Fatalf("持久化失败: %v", err)
		}
		if i == WindowCap-1 && !full {
			t.Error("达到上限时应触发压缩信号")
		}
	}
	if len(store.saved) != WindowCap {
		t.Errorf("SQLite 应持久化 %d 条, 实际 %d", WindowCap, len(store.saved))
	}
}

func TestPersistMessageLoadsMemberProfile(t *testing.T) {
	store := &fakeStorage{}
	store.memberProf = map[string]*entity.MemberProfile{
		"g1|u1": {GroupID: "g1", UserID: "u1", Activity: 0.9},
	}
	svc := NewMemoryService(NewWindow(), store)
	msg := entity.Message{GroupID: "g1", UserID: "u1", MessageType: "group", Content: "hi"}

	if _, err := svc.PersistMessage(context.Background(), msg); err != nil {
		t.Fatalf("持久化失败: %v", err)
	}
	prof, ok := svc.GetMemberProfile("g1", "u1")
	if !ok || prof == nil || prof.Activity != 0.9 {
		t.Errorf("成员画像未加载到缓存: ok=%v prof=%+v", ok, prof)
	}
}
