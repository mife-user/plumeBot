package event

import (
	"context"
	"os"
	"testing"

	"plumebot/internal/domain/entity"
	"plumebot/pkg/logger"
)

// TestMain 初始化全局 logger（error 级，避免测试日志噪音与 nil 指针）。
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "plumebot-test-logs")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	logger.Init(logger.Config{Level: "error", Dir: dir})
	os.Exit(m.Run())
}

// chainOrder 记录中间件执行顺序。
type chainOrder struct {
	order []string
}

func (c *chainOrder) record(name string) {
	c.order = append(c.order, name)
}

func TestChainFixedOrder(t *testing.T) {
	rec := &chainOrder{}
	mk := func(name string) Middleware {
		return func(next Handler) Handler {
			return func(ctx context.Context, msg entity.Message) error {
				rec.record(name)
				return next(ctx, msg)
			}
		}
	}
	h := chain([]Middleware{mk("log"), mk("limit"), mk("sensitive")}, func(ctx context.Context, msg entity.Message) error {
		rec.record("tail")
		return nil
	})
	if err := h(context.Background(), groupMsg("m1")); err != nil {
		t.Fatalf("链执行失败: %v", err)
	}
	want := []string{"log", "limit", "sensitive", "tail"}
	if len(rec.order) != len(want) {
		t.Fatalf("执行顺序 %v，期望 %v", rec.order, want)
	}
	for i := range want {
		if rec.order[i] != want[i] {
			t.Fatalf("执行顺序 %v，期望 %v", rec.order, want)
		}
	}
}

// 敏感词 stub 当前永远放行：消息应到达末端 handler。
func TestSensitiveWordStubPassesThrough(t *testing.T) {
	reached := false
	h := sensitiveWordMiddleware(func(ctx context.Context, msg entity.Message) error {
		reached = true
		return nil
	})
	if err := h(context.Background(), groupMsg("m1")); err != nil {
		t.Fatalf("stub 应放行: %v", err)
	}
	if !reached {
		t.Fatal("敏感词 stub 应把消息传给下一个 handler")
	}
}

// 日志中间件不应阻断消息。
func TestLogMiddlewareDoesNotBlock(t *testing.T) {
	reached := false
	h := logMiddleware(func(ctx context.Context, msg entity.Message) error {
		reached = true
		return nil
	})
	if err := h(context.Background(), groupMsg("m1")); err != nil {
		t.Fatalf("日志中间件不应报错: %v", err)
	}
	if !reached {
		t.Fatal("日志中间件不应阻断消息")
	}
}
