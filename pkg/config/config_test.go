package config

import (
	"os"
	"path/filepath"
	"strings"
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

// 配置文件不存在时：写入嵌入的默认配置，且解析出的值来自默认 yaml。
func TestLoadMissingFileWritesDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 默认配置已写入磁盘
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("默认配置未写入: %v", err)
	}
	if string(content) != string(defaultConfigYAML) {
		t.Fatal("写入的默认配置与嵌入模板不一致")
	}

	// 解析结果应包含默认值（来自 yaml，而非 Go 侧兜底）
	if cfg.Bot.Name != "PlumeBot" {
		t.Errorf("Bot.Name = %q，期望 %q", cfg.Bot.Name, "PlumeBot")
	}
	if cfg.Onebot.WsURL != "ws://127.0.0.1:3001" {
		t.Errorf("Onebot.WsURL = %q，期望默认地址", cfg.Onebot.WsURL)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level = %q，期望 %q", cfg.Log.Level, "info")
	}
	if cfg.Control.Mode != "mention" {
		t.Errorf("Control.Mode = %q，期望 %q", cfg.Control.Mode, "mention")
	}
	if cfg.Middleware.RateLimit.Rate != 2 || cfg.Middleware.RateLimit.Burst != 20 ||
		cfg.Middleware.RateLimit.MaxWaitSeconds != 10 {
		t.Errorf("限流默认值 = %+v，期望 rate=2/burst=20/max_wait=10",
			cfg.Middleware.RateLimit)
	}
}

// 配置文件已存在时：不做 Go 侧兜底，空值/非法值原样保留（由消费包处理）。
func TestLoadExistingFileNoFallback(t *testing.T) {
	path := writeTempConfig(t,
		"bot:\n  name: \"\"\n"+
			"onebot:\n  ws_url: \"\"\n"+
			"middleware:\n  rate_limit:\n    rate: 0\n    burst: 0\n    max_wait_seconds: 0\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Bot.Name != "" {
		t.Errorf("Bot.Name 空值不应被改写，实际 %q", cfg.Bot.Name)
	}
	if cfg.Onebot.WsURL != "" {
		t.Errorf("Onebot.WsURL 空值不应被改写，实际 %q", cfg.Onebot.WsURL)
	}
	if got := cfg.Middleware.RateLimit.Rate; got != 0 {
		t.Errorf("Rate 零值不应被改写，实际 %v", got)
	}
	if got := cfg.Middleware.RateLimit.Burst; got != 0 {
		t.Errorf("Burst 零值不应被改写，实际 %v", got)
	}
	if got := cfg.Middleware.RateLimit.MaxWaitSeconds; got != 0 {
		t.Errorf("MaxWaitSeconds 零值不应被改写，实际 %v", got)
	}
}

// 配置文件已存在时：合法值原样保留。
func TestLoadExistingFilePreservesValues(t *testing.T) {
	path := writeTempConfig(t,
		"bot:\n  name: TestBot\n"+
			"middleware:\n  rate_limit:\n    rate: 5\n    burst: 30\n    max_wait_seconds: 15\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Bot.Name != "TestBot" {
		t.Errorf("Bot.Name = %q，期望 %q", cfg.Bot.Name, "TestBot")
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

// 路径父目录不存在时，写入默认配置应自动创建目录。
func TestLoadMissingFileCreatesParentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "nested")
	path := filepath.Join(dir, "config.yaml")

	if _, err := Load(path); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("默认配置应写入嵌套目录: %v", err)
	}
}

// 嵌入的默认配置应包含全部配置节（与 Config 结构对应）。
func TestDefaultYAMLHasAllSections(t *testing.T) {
	content := string(defaultConfigYAML)
	for _, section := range []string{"bot:", "onebot:", "log:", "control:", "middleware:", "rate_limit:"} {
		if !strings.Contains(content, section) {
			t.Errorf("默认配置缺少 %q 节", section)
		}
	}
}
