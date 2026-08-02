package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
	"plumebot/pkg/config"
)

func groupMsg(id string) entity.Message {
	return entity.Message{MessageID: id, GroupID: "g1", UserID: "u1", MessageType: "group"}
}

func TestRateLimiterBurstPassThrough(t *testing.T) {
	rl := newRateLimiter(config.RateLimitConfig{Rate: 1, Burst: 2, MaxWaitSeconds: 10})
	for i := 0; i < 2; i++ {
		if err := rl.wait(context.Background(), groupMsg("m1")); err != nil {
			t.Fatalf("突发消息 #%d 应直通，实际报错: %v", i+1, err)
		}
	}
}

func TestRateLimiterQueueThenPass(t *testing.T) {
	rl := newRateLimiter(config.RateLimitConfig{Rate: 100, Burst: 1, MaxWaitSeconds: 10})
	if err := rl.wait(context.Background(), groupMsg("m1")); err != nil {
		t.Fatalf("第 1 条应直通: %v", err)
	}
	// 第 2 条需等待令牌：rate=100/s 下 10ms 内应放行
	start := time.Now()
	if err := rl.wait(context.Background(), groupMsg("m2")); err != nil {
		t.Fatalf("排队消息应放行: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 5*time.Millisecond {
		t.Fatalf("排队消息应等待令牌，实际仅耗时 %v", elapsed)
	}
}

func TestRateLimiterTimeoutReturnsSentinel(t *testing.T) {
	rl := newRateLimiter(config.RateLimitConfig{Rate: 0.001, Burst: 1, MaxWaitSeconds: 1})
	if err := rl.wait(context.Background(), groupMsg("m1")); err != nil {
		t.Fatalf("第 1 条应直通: %v", err)
	}
	err := rl.wait(context.Background(), groupMsg("m2"))
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("超时应返回 domain.ErrRateLimited，实际: %v", err)
	}
}

func TestRateLimiterKeyIsolation(t *testing.T) {
	rl := newRateLimiter(config.RateLimitConfig{Rate: 0.001, Burst: 1, MaxWaitSeconds: 1})
	if err := rl.wait(context.Background(), groupMsg("m1")); err != nil {
		t.Fatalf("群 g1 第 1 条应直通: %v", err)
	}
	// 另一群不受影响
	other := entity.Message{MessageID: "m1", GroupID: "g2", UserID: "u1", MessageType: "group"}
	if err := rl.wait(context.Background(), other); err != nil {
		t.Fatalf("群 g2 应独立放行: %v", err)
	}
	// 私聊按用户独立
	private := entity.Message{MessageID: "m1", GroupID: "", UserID: "u1", MessageType: "private"}
	if err := rl.wait(context.Background(), private); err != nil {
		t.Fatalf("私聊应独立放行: %v", err)
	}
}

func TestRateLimitMiddlewareStopsChainOnTimeout(t *testing.T) {
	rl := newRateLimiter(config.RateLimitConfig{Rate: 0.001, Burst: 1, MaxWaitSeconds: 1})
	rl.wait(context.Background(), groupMsg("m1")) // 耗尽令牌

	called := false
	h := rateLimitMiddleware(rl)(func(ctx context.Context, msg entity.Message) error {
		called = true
		return nil
	})
	err := h(context.Background(), groupMsg("m2"))
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("应返回 ErrRateLimited，实际: %v", err)
	}
	if called {
		t.Fatal("超时丢弃时不应调用后续 handler")
	}
}

// 上游 context 取消不应被归因为限流超时，应原样上抛。
func TestRateLimiterUpstreamCancelNotRateLimited(t *testing.T) {
	rl := newRateLimiter(config.RateLimitConfig{Rate: 0.001, Burst: 1, MaxWaitSeconds: 10})
	rl.wait(context.Background(), groupMsg("m1")) // 耗尽令牌

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 先取消，模拟上游终止

	err := rl.wait(ctx, groupMsg("m2"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("上游取消应返回 context.Canceled，实际: %v", err)
	}
}
