package memory

import (
	"context"
	"strconv"
	"testing"

	"plumebot/internal/domain/entity"
)

func groupMsg(groupID, content string) entity.Message {
	return entity.Message{GroupID: groupID, MessageType: "group", Content: content}
}

func TestWindowAppendAndGet(t *testing.T) {
	w := NewWindow()
	full, err := w.AppendMessage(context.Background(), groupMsg("g1", "hi"))
	if err != nil {
		t.Fatalf("追加失败: %v", err)
	}
	if full {
		t.Error("首条消息不应触发压缩信号")
	}
	got, err := w.GetWindow(context.Background(), "g1")
	if err != nil {
		t.Fatalf("读取窗口失败: %v", err)
	}
	if len(got) != 1 || got[0].Content != "hi" {
		t.Errorf("窗口内容错误: %+v", got)
	}
}

func TestWindowPerGroupIsolation(t *testing.T) {
	w := NewWindow()
	w.AppendMessage(context.Background(), groupMsg("g1", "a"))
	w.AppendMessage(context.Background(), groupMsg("g2", "b"))

	g1, _ := w.GetWindow(context.Background(), "g1")
	g2, _ := w.GetWindow(context.Background(), "g2")
	if len(g1) != 1 || g1[0].Content != "a" {
		t.Errorf("g1 窗口错误: %+v", g1)
	}
	if len(g2) != 1 || g2[0].Content != "b" {
		t.Errorf("g2 窗口错误: %+v", g2)
	}
}

func TestWindowPrivateSessionKeyIsolated(t *testing.T) {
	w := NewWindow()
	// 私聊消息 GroupID 为空，按用户隔离，避免互相串窗。
	w.AppendMessage(context.Background(), entity.Message{UserID: "u1", MessageType: "private", Content: "a"})
	w.AppendMessage(context.Background(), entity.Message{UserID: "u2", MessageType: "private", Content: "b"})

	u1, _ := w.GetWindow(context.Background(), "private:u1")
	u2, _ := w.GetWindow(context.Background(), "private:u2")
	if len(u1) != 1 || u1[0].Content != "a" {
		t.Errorf("u1 私聊窗口错误: %+v", u1)
	}
	if len(u2) != 1 || u2[0].Content != "b" {
		t.Errorf("u2 私聊窗口错误: %+v", u2)
	}
}

func TestWindowReachesCapSignalsCompression(t *testing.T) {
	w := NewWindow()
	for i := 0; i < WindowCap-1; i++ {
		full, err := w.AppendMessage(context.Background(), groupMsg("g1", "m"))
		if err != nil {
			t.Fatalf("追加失败: %v", err)
		}
		if full {
			t.Fatalf("第 %d 条不应触发压缩信号", i+1)
		}
	}
	full, _ := w.AppendMessage(context.Background(), groupMsg("g1", "cap"))
	if !full {
		t.Error("达到上限时应触发压缩信号")
	}
}

func TestWindowEvictsOldestBeyondCap(t *testing.T) {
	w := NewWindow()
	for i := 1; i <= WindowCap+5; i++ {
		m := groupMsg("g1", "m")
		m.MessageID = strconv.Itoa(i)
		_, _ = w.AppendMessage(context.Background(), m)
	}
	got, _ := w.GetWindow(context.Background(), "g1")
	if len(got) != WindowCap {
		t.Fatalf("窗口长度 = %d, want %d", len(got), WindowCap)
	}
	// 最旧的 5 条被淘汰，窗口保留最新 WindowCap 条（MessageID 从 6 开始）。
	if got[0].MessageID != "6" || got[WindowCap-1].MessageID != strconv.Itoa(WindowCap+5) {
		t.Errorf("淘汰最旧消息错误: 首条 %q 末条 %q", got[0].MessageID, got[WindowCap-1].MessageID)
	}
}

func TestWindowGetWindowReturnsCopy(t *testing.T) {
	w := NewWindow()
	w.AppendMessage(context.Background(), groupMsg("g1", "a"))

	got, _ := w.GetWindow(context.Background(), "g1")
	got[0].Content = "mutated"
	again, _ := w.GetWindow(context.Background(), "g1")
	if again[0].Content != "a" {
		t.Error("GetWindow 应返回副本，外部修改不应影响窗口内部数据")
	}
}
