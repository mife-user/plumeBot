package ai

import (
	"context"
	"strings"
	"testing"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
	"plumebot/pkg/config"
)

// recordFactory 记录收到的配置并返回预设 agent，用于验证分发逻辑。
func recordFactory(agentName string, gotCfg *config.Config) Factory {
	return func(_ context.Context, cfg config.Config) (domain.Agent, error) {
		*gotCfg = cfg
		return &stubAgent{name: agentName}, nil
	}
}

// stubAgent 是最小 domain.Agent 实现（测试替身）。
type stubAgent struct {
	name string
}

func (s *stubAgent) Name() string { return s.name }

func (s *stubAgent) Generate(_ context.Context, _ []entity.ChatMessage) (string, error) {
	return s.name, nil
}

func TestRegistryDispatch(t *testing.T) {
	r := NewRegistry()
	var got config.Config
	if err := r.Register("fake_a", recordFactory("A", &got)); err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	if err := r.Register("fake_b", recordFactory("B", &got)); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	cfg := config.Config{LLM: config.LLMConfig{Provider: "fake_b"}}
	agent, err := r.NewAgent(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewAgent 失败: %v", err)
	}
	if agent.(*stubAgent).Name() != "B" {
		t.Errorf("分发的 agent = %q，期望 fake_b 的工厂产物", agent.(*stubAgent).Name())
	}
	if got.LLM.Provider != "fake_b" {
		t.Errorf("工厂收到的 cfg.LLM.Provider = %q，期望原样透传", got.LLM.Provider)
	}
}

func TestRegistryUnknownProvider(t *testing.T) {
	r := NewRegistry()
	if err := r.Register("fake", recordFactory("F", &config.Config{})); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	_, err := r.NewAgent(context.Background(), config.Config{LLM: config.LLMConfig{Provider: "nope"}})
	if err == nil {
		t.Fatal("未知 provider 应报错")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "已注册：fake") {
		t.Errorf("错误应含 provider 名与已注册列表，实际 %q", err.Error())
	}
}

// provider 为空 → 兜底 DefaultLLMProvider（openai）。
func TestRegistryEmptyProviderDefaults(t *testing.T) {
	r := NewRegistry()
	var got config.Config
	if err := r.Register(config.DefaultLLMProvider, recordFactory("O", &got)); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	cfg := config.Config{LLM: config.LLMConfig{Provider: ""}}
	agent, err := r.NewAgent(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewAgent 失败: %v", err)
	}
	if agent.(*stubAgent).Name() != "O" {
		t.Errorf("空 provider 应分发到 openai 工厂，实际 %q", agent.(*stubAgent).Name())
	}
}

func TestRegistryDuplicateRegister(t *testing.T) {
	r := NewRegistry()
	if err := r.Register("fake", recordFactory("F", &config.Config{})); err != nil {
		t.Fatalf("首次注册失败: %v", err)
	}
	err := r.Register("fake", recordFactory("F2", &config.Config{}))
	if err == nil || !strings.Contains(err.Error(), "已注册") {
		t.Errorf("重名注册应报错，实际 %v", err)
	}
}

func fullLLMCfg() config.Config {
	temp := 0.7
	return config.Config{
		LLM: config.LLMConfig{
			Provider:       "openai",
			TimeoutSeconds: 30,
			OpenAI: config.OpenAICompatConfig{
				BaseURL:     "https://api.deepseek.com/v1",
				APIKey:      "sk-test",
				Model:       "deepseek-chat",
				Temperature: &temp,
				MaxTokens:   512,
			},
		},
	}
}

// openai 工厂构造：完整配置 → NewChatModel + NewEinoAgent（不发网络）。
func TestOpenAIFactoryConstruct(t *testing.T) {
	f := NewOpenAIFactory(NewToolsRegistry())

	agent, err := f(context.Background(), fullLLMCfg())
	if err != nil {
		t.Fatalf("工厂执行失败: %v", err)
	}
	if _, ok := agent.(*EinoAgent); !ok {
		t.Errorf("工厂应产出 *EinoAgent，实际 %T", agent)
	}
}

// model 为空 → 构造期报错（不可猜测）。
func TestOpenAIFactoryModelEmpty(t *testing.T) {
	f := NewOpenAIFactory(NewToolsRegistry())
	cfg := fullLLMCfg()
	cfg.LLM.OpenAI.Model = ""

	_, err := f(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "llm.openai.model") {
		t.Errorf("model 空应报错并提示配置项，实际 %v", err)
	}
}

// base_url 为空 → 兜底默认端点（构造成功即证明走了默认值路径）。
func TestOpenAIFactoryBaseURLFallback(t *testing.T) {
	f := NewOpenAIFactory(NewToolsRegistry())
	cfg := fullLLMCfg()
	cfg.LLM.OpenAI.BaseURL = ""

	if _, err := f(context.Background(), cfg); err != nil {
		t.Fatalf("base_url 空应兜底默认值，实际报错: %v", err)
	}
}

// tools.enabled 含未注册工具 → 构造期报错（含已注册列表）。
func TestOpenAIFactoryUnknownTool(t *testing.T) {
	f := NewOpenAIFactory(NewToolsRegistry())
	cfg := fullLLMCfg()
	cfg.Tools = config.ToolsConfig{Enabled: []string{"nope"}}

	_, err := f(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "未注册") {
		t.Errorf("未注册工具应报错，实际 %v", err)
	}
}
