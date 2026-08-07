package memory

import (
	"context"
	"testing"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// fakeStorage 嵌入 domain.Storage 接口，仅实现 SaveMessage 记录
// （PersistMessage 只用到这一个方法，其余方法永不调用）。
type fakeStorage struct {
	domain.Storage
	saved []entity.Message
}

func (f *fakeStorage) SaveMessage(_ context.Context, msg entity.Message) error {
	f.saved = append(f.saved, msg)
	return nil
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
