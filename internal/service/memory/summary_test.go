package memory

import (
	"context"
	"testing"

	"plumebot/internal/domain/entity"
)

func TestSummaryAppendAssignsSeq(t *testing.T) {
	store := &fakeStorage{}
	ss := NewSummaryStore(store)

	s1 := ss.Append(context.Background(), "g1", "第一段", nil, nil)
	s2 := ss.Append(context.Background(), "g1", "第二段", nil, nil)
	if s1.Seq != 1 || s2.Seq != 2 {
		t.Errorf("seq 应递增, 实际 %d, %d", s1.Seq, s2.Seq)
	}

	got := ss.GetAll(context.Background(), "g1")
	if len(got) != 2 || got[0].Text != "第一段" || got[1].Text != "第二段" {
		t.Errorf("热链应按旧→新保存, 实际: %+v", got)
	}
	if len(store.archived) != 0 {
		t.Errorf("新追加的摘要不应立即归档, 实际 %d 条", len(store.archived))
	}
}

func TestSummaryStorePerChatIsolation(t *testing.T) {
	ss := NewSummaryStore(&fakeStorage{})
	ss.Append(context.Background(), "g1", "g1 内容", nil, nil)
	ss.Append(context.Background(), "g2", "g2 内容", nil, nil)

	g1 := ss.GetAll(context.Background(), "g1")
	g2 := ss.GetAll(context.Background(), "g2")
	if len(g1) != 1 || len(g2) != 1 {
		t.Errorf("会话热链应互相隔离, g1=%d g2=%d", len(g1), len(g2))
	}
}

func TestSummaryReplaceWithFusedArchivesOld(t *testing.T) {
	store := &fakeStorage{}
	ss := NewSummaryStore(store)
	ss.Append(context.Background(), "g1", "a", nil, nil)
	ss.Append(context.Background(), "g1", "b", nil, nil)

	fused := ss.ReplaceWithFused(context.Background(), "g1", "融合", []string{"k"}, []string{"d"})
	if fused.Seq != 3 || fused.Text != "融合" {
		t.Errorf("融合摘要 seq 应续号=3, 实际 %+v", fused)
	}
	got := ss.GetAll(context.Background(), "g1")
	if len(got) != 1 || got[0].Text != "融合" {
		t.Errorf("热链应只剩融合摘要, 实际: %+v", got)
	}
	if len(store.archived) != 2 {
		t.Errorf("被覆盖的 2 条摘要应归档, 实际 %d 条", len(store.archived))
	}
}

func TestSummaryEvictOldestArchives(t *testing.T) {
	store := &fakeStorage{}
	ss := NewSummaryStore(store)
	ss.Append(context.Background(), "g1", "a", nil, nil)
	ss.Append(context.Background(), "g1", "b", nil, nil)

	evicted := ss.EvictOldest(context.Background(), "g1")
	if evicted.Text != "a" {
		t.Errorf("应淘汰最旧的 a, 实际 %+v", evicted)
	}
	got := ss.GetAll(context.Background(), "g1")
	if len(got) != 1 || got[0].Text != "b" {
		t.Errorf("热链应只剩 b, 实际: %+v", got)
	}
	if len(store.archived) != 1 || store.archived[0].Text != "a" {
		t.Errorf("被淘汰的 a 应归档, 实际: %+v", store.archived)
	}
}

func TestSummaryStoreSeedsFromArchive(t *testing.T) {
	store := &fakeStorage{}
	// 模拟重启前的归档：g1 已有 seq 3、4 两条摘要（1、2 已在更早融合中被覆盖后淘汰）。
	store.archived = []entity.Summary{
		{ChatID: "g1", Seq: 3, Text: "旧三"},
		{ChatID: "g1", Seq: 4, Text: "旧四"},
	}
	ss := NewSummaryStore(store)

	got := ss.GetAll(context.Background(), "g1")
	if len(got) != 2 || got[0].Seq != 3 || got[1].Seq != 4 {
		t.Errorf("应回灌归档最新 2 条作热链底, 实际: %+v", got)
	}

	// 重启后的新摘要应续号到最大 seq+1，不重复。
	appended := ss.Append(context.Background(), "g1", "重启后", nil, nil)
	if appended.Seq != 5 {
		t.Errorf("重启后 seq 应从 5 续号, 实际 %d", appended.Seq)
	}
}
