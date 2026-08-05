package ai

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"

	"plumebot/internal/domain"
	"plumebot/pkg/config"
)

// Factory 根据完整配置构建 domain.Agent（LLM 映射 + 工具组装 + Agent 封装）。
// 接收完整 config.Config：LLM 配置在 LLM 段，工具启用列表在 Tools 段，二者组装点在此。
type Factory func(ctx context.Context, cfg config.Config) (domain.Agent, error)

// Registry 是 LLM provider 注册中心（注入式实例，无全局状态）。
type Registry struct {
	factories map[string]Factory
}

// NewRegistry 创建空注册中心。
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register 注册 provider 工厂；重名返回错误（防止静默覆盖）。
func (r *Registry) Register(name string, f Factory) error {
	if _, ok := r.factories[name]; ok {
		return fmt.Errorf("provider %q 已注册", name)
	}
	r.factories[name] = f
	return nil
}

// NewAgent 按 cfg.LLM.Provider 分发构建 Agent。
// provider 空 → DefaultLLMProvider；未注册 → 错误（错误信息含已注册列表）。
func (r *Registry) NewAgent(ctx context.Context, cfg config.Config) (domain.Agent, error) {
	name := cfg.LLM.Provider
	if name == "" {
		name = config.DefaultLLMProvider
	}

	f, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("未知 LLM provider %q（已注册：%s）", name, registeredProviders(r.factories))
	}
	return f(ctx, cfg)
}

// registeredProviders 返回已注册 provider 名的排序、逗号分隔列表（用于错误提示）。
func registeredProviders(m map[string]Factory) string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// NewOpenAIFactory 创建 openai（OpenAI 兼容接口）provider 工厂。
// 依赖工具注册表，通过闭包注入：按 cfg.Tools.Enabled 过滤启用工具。
// 兜底语义（消费方）：base_url 空 → DefaultOpenAIBaseURL；model 空 → 构造期报错（不可猜测）；
// timeout_seconds ≤0 → DefaultLLMTimeoutSeconds；api_key 空透传；Temperature 转 *float32；
// MaxTokens >0 → MaxCompletionTokens（*int，不用已废弃的 MaxTokens 字段）。
func NewOpenAIFactory(tr *ToolsRegistry) Factory {
	return func(ctx context.Context, cfg config.Config) (domain.Agent, error) {
		o := cfg.LLM.OpenAI

		baseURL := o.BaseURL
		if baseURL == "" {
			baseURL = config.DefaultOpenAIBaseURL
		}
		if o.Model == "" {
			return nil, errors.New("llm.openai.model 未配置（空值不可猜测，请在 config.yaml 中填写模型名）")
		}
		timeout := cfg.LLM.TimeoutSeconds
		if timeout <= 0 {
			timeout = config.DefaultLLMTimeoutSeconds
		}

		var temperature *float32
		if o.Temperature != nil {
			t := float32(*o.Temperature)
			temperature = &t
		}
		var maxTokens *int
		if o.MaxTokens > 0 {
			mt := o.MaxTokens
			maxTokens = &mt
		}

		cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			BaseURL:             baseURL,
			APIKey:              o.APIKey,
			Model:               o.Model,
			Timeout:             time.Duration(timeout) * time.Second,
			Temperature:         temperature,
			MaxCompletionTokens: maxTokens,
		})
		if err != nil {
			return nil, fmt.Errorf("构造 openai ChatModel 失败: %w", err)
		}

		tools, err := tr.EnabledTools(cfg.Tools.Enabled)
		if err != nil {
			return nil, err
		}
		return NewEinoAgent(ctx, cm, tools)
	}
}
