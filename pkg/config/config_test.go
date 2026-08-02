package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// 缺失/空值配置应回落到默认值：rate=2、burst=20、max_wait=10。
func TestLoadRateLimitDefaults(t *testing.T) {
	cfg, err := Load(writeTempConfig(t, "middleware:\n  rate_limit:\n"))
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if got := cfg.Middleware.RateLimit.Rate; got != 2 {
		t.Errorf("Rate 默认值 = %v，期望 2", got)
	}
	if got := cfg.Middleware.RateLimit.Burst; got != 20 {
		t.Errorf("Burst 默认值 = %v，期望 20", got)
	}
	if got := cfg.Middleware.RateLimit.MaxWaitSeconds; got != 10 {
		t.Errorf("MaxWaitSeconds 默认值 = %v，期望 10", got)
	}
}

// 零值字段（0 被视作未配置）应回落默认值。
func TestLoadRateLimitZeroFallsBack(t *testing.T) {
	cfg, err := Load(writeTempConfig(t,
		"middleware:\n  rate_limit:\n    rate: 0\n    burst: 0\n    max_wait_seconds: 0\n"))
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if got := cfg.Middleware.RateLimit.Rate; got != 2 {
		t.Errorf("Rate 零值应回落为 2，实际 %v", got)
	}
	if got := cfg.Middleware.RateLimit.Burst; got != 20 {
		t.Errorf("Burst 零值应回落为 20，实际 %v", got)
	}
	if got := cfg.Middleware.RateLimit.MaxWaitSeconds; got != 10 {
		t.Errorf("MaxWaitSeconds 零值应回落为 10，实际 %v", got)
	}
}

// 合法配置值不应被改写。
func TestLoadRateLimitPreservesValues(t *testing.T) {
	cfg, err := Load(writeTempConfig(t,
		"middleware:\n  rate_limit:\n    rate: 5\n    burst: 30\n    max_wait_seconds: 15\n"))
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if got := cfg.Middleware.RateLimit.Rate; got != 5 {
		t.Errorf("Rate = %v，期望 5", got)
	}
	if got := cfg.Middleware.RateLimit.Burst; got != 30 {
		t.Errorf("Burst = %v，期望 30", got)
	}
	if got := cfg.Middleware.RateLimit.MaxWaitSeconds; got != 15 {
		t.Errorf("MaxWaitSeconds = %v，期望 15", got)
	}
}
