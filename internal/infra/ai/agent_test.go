package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"

	"plumebot/internal/domain/entity"
	"plumebot/pkg/config"
)

// agentCfg 构造测试用 AgentConfig（name/description 固定，system_prompt 按用例传）。
func agentCfg(system string) config.AgentConfig {
	return config.AgentConfig{Name: "test-agent", Description: "测试 agent", SystemPrompt: system}
}

func sysMsg(text string) entity.ChatMessage {
	return entity.ChatMessage{Role: entity.RoleSystem, Parts: []entity.ContentPart{{Type: entity.PartTypeText, Text: text}}}
}

func usrMsg(text string) entity.ChatMessage {
	return entity.ChatMessage{Role: entity.RoleUser, Parts: []entity.ContentPart{{Type: entity.PartTypeText, Text: text}}}
}

// 文本往返：业务消息 → schema 消息 → 模型 → 最终文本；system+user 原样透传。
func TestEinoAgentTextRoundtrip(t *testing.T) {
	fake := &fakeChatModel{script: []*schema.Message{
		schema.AssistantMessage("你好，我是 PlumeBot。", nil),
	}}

	agent, err := NewEinoAgent(context.Background(), fake, nil, agentCfg(""))
	if err != nil {
		t.Fatalf("NewEinoAgent 失败: %v", err)
	}

	got, err := agent.Generate(context.Background(), []entity.ChatMessage{sysMsg("你是 PlumeBot。"), usrMsg("你好")})
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if got != "你好，我是 PlumeBot。" {
		t.Errorf("回复 = %q，期望「你好，我是 PlumeBot。」", got)
	}

	in := fake.inputs[0]
	if len(in) != 2 {
		t.Fatalf("模型收到 %d 条消息，期望 2", len(in))
	}
	if in[0].Role != schema.System || in[1].Role != schema.User {
		t.Errorf("角色不符: %q / %q", in[0].Role, in[1].Role)
	}
	if in[0].Content != "你是 PlumeBot。" || in[1].Content != "你好" {
		t.Errorf("文本未透传: system=%q user=%q", in[0].Content, in[1].Content)
	}
}

// Instruction 注入：固定人设由 AgentConfig.SystemPrompt 提供，adk 在每次 Run 时前置为 system 消息
// （defaultGenModelInput），业务消息只传 user。
func TestEinoAgentInstructionInjected(t *testing.T) {
	fake := &fakeChatModel{script: []*schema.Message{
		schema.AssistantMessage("好的", nil),
	}}

	agent, err := NewEinoAgent(context.Background(), fake, nil, agentCfg("你是 PlumeBot，一个赛博群友。"))
	if err != nil {
		t.Fatalf("NewEinoAgent 失败: %v", err)
	}

	got, err := agent.Generate(context.Background(), []entity.ChatMessage{usrMsg("你好")})
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if got != "好的" {
		t.Errorf("回复 = %q", got)
	}

	in := fake.inputs[0]
	if len(in) != 2 {
		t.Fatalf("模型收到 %d 条消息，期望 2（instruction system + user）", len(in))
	}
	if in[0].Role != schema.System || in[0].Content != "你是 PlumeBot，一个赛博群友。" || in[1].Role != schema.User {
		t.Errorf("消息列表不符: %+v，期望 [system(instruction), user]", in)
	}
}

// name/description 透传：NewEinoAgent 把 AgentConfig 元数据写入 adk 配置（构造成功即被接受）。
func TestEinoAgentNameDescription(t *testing.T) {
	fake := &fakeChatModel{script: []*schema.Message{
		schema.AssistantMessage("ok", nil),
	}}
	agent, err := NewEinoAgent(context.Background(), fake, nil, agentCfg(""))
	if err != nil {
		t.Fatalf("NewEinoAgent 失败: %v", err)
	}
	if _, err := agent.Generate(context.Background(), []entity.ChatMessage{usrMsg("hi")}); err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
}

// 模型错误（脚本用尽）应透传，adk 会包装错误信息。
func TestEinoAgentPropagatesModelError(t *testing.T) {
	fake := &fakeChatModel{} // 空脚本 → 首次调用即报错
	agent, err := NewEinoAgent(context.Background(), fake, nil, agentCfg(""))
	if err != nil {
		t.Fatalf("NewEinoAgent 失败: %v", err)
	}

	_, err = agent.Generate(context.Background(), []entity.ChatMessage{usrMsg("hi")})
	if err == nil {
		t.Fatal("期望报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "脚本用尽") {
		t.Errorf("错误应包含模型原始错误，实际 %q", err.Error())
	}
}

// echoArgs 是 echo 工具的参数结构（jsonschema 由 struct 推断）。
type echoArgs struct {
	Text string `json:"text" jsonschema:"description=要回显的文本"`
}

// tool 自动循环：模型先回 tool_call → echo 执行 → 结果回填 tool 消息 → 模型回最终答案。
func TestEinoAgentToolLoop(t *testing.T) {
	var echoCalled []string
	params, err := toolutils.GoStruct2ParamsOneOf[echoArgs]()
	if err != nil {
		t.Fatalf("GoStruct2ParamsOneOf 失败: %v", err)
	}
	echo := toolutils.NewTool(&schema.ToolInfo{Name: "echo", Desc: "回显文本", ParamsOneOf: params},
		func(_ context.Context, args echoArgs) (string, error) {
			echoCalled = append(echoCalled, args.Text)
			return args.Text, nil
		})

	fake := &fakeChatModel{script: []*schema.Message{
		schema.AssistantMessage("", []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "echo", Arguments: `{"text":"hi"}`}},
		}),
		schema.AssistantMessage("已回显：hi", nil),
	}}

	agent, err := NewEinoAgent(context.Background(), fake, []tool.BaseTool{echo}, agentCfg(""))
	if err != nil {
		t.Fatalf("NewEinoAgent 失败: %v", err)
	}

	got, err := agent.Generate(context.Background(), []entity.ChatMessage{usrMsg("echo hi")})
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if got != "已回显：hi" {
		t.Errorf("回复 = %q，期望「已回显：hi」", got)
	}
	if len(echoCalled) != 1 || echoCalled[0] != "hi" {
		t.Errorf("echo 执行记录 = %v，期望 [hi]", echoCalled)
	}

	// 第二轮输入应包含 tool 角色消息（工具结果回填）
	second := fake.inputs[1]
	foundToolMsg := false
	for _, m := range second {
		if m.Role == schema.Tool && m.ToolCallID == "call-1" && m.Content == "hi" {
			foundToolMsg = true
		}
	}
	if !foundToolMsg {
		t.Errorf("第二轮输入应含 tool 结果消息，实际 %+v", second)
	}
}

// 模型返回 nil 消息：adk 或 EinoAgent 应报错，不得返回空串静默成功。
func TestEinoAgentNilModelOutput(t *testing.T) {
	fake := &fakeChatModel{script: []*schema.Message{nil}}
	agent, err := NewEinoAgent(context.Background(), fake, nil, agentCfg(""))
	if err != nil {
		t.Fatalf("NewEinoAgent 失败: %v", err)
	}

	_, err = agent.Generate(context.Background(), []entity.ChatMessage{usrMsg("hi")})
	if err == nil {
		t.Fatal("模型返回 nil 消息应报错")
	}
	if !errors.Is(err, errNoModelOutput) && !strings.Contains(err.Error(), "nil") {
		t.Errorf("错误应来自 errNoModelOutput 或 adk nil 校验，实际 %q", err.Error())
	}
}
