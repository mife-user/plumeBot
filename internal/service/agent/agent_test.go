package agent

import (
	"context"
	"testing"

	"plumebot/internal/domain/entity"
	"plumebot/pkg/config"
)

// fakeAgent 记录收到的消息列表，返回预设回复，用于验证组装逻辑。
type fakeAgent struct {
	messages []entity.ChatMessage
	reply    string
	err      error
}

func (f *fakeAgent) Generate(_ context.Context, msgs []entity.ChatMessage) (string, error) {
	f.messages = append([]entity.ChatMessage(nil), msgs...)
	return f.reply, f.err
}

func userMsg(text string) entity.ChatMessage {
	return entity.ChatMessage{Role: entity.RoleUser, Parts: []entity.ContentPart{{Type: entity.PartTypeText, Text: text}}}
}

// GenerateReply 应将 system 提示词前置为第一条消息，用户消息原样追加，回复透传。
func TestGenerateReplyPrependsSystem(t *testing.T) {
	fake := &fakeAgent{reply: "测试回复"}
	svc := NewAgentService(fake, "系统人设：你好")

	got, err := svc.GenerateReply(context.Background(), []entity.ChatMessage{userMsg("第一条"), userMsg("第二条")})
	if err != nil {
		t.Fatalf("GenerateReply 失败: %v", err)
	}
	if got != "测试回复" {
		t.Errorf("回复 = %q，期望透传 %q", got, "测试回复")
	}

	if len(fake.messages) != 3 {
		t.Fatalf("agent 应收到 3 条消息（system+2 user），实际 %d 条", len(fake.messages))
	}
	if fake.messages[0].Role != entity.RoleSystem {
		t.Errorf("第 1 条角色 = %q，期望 system", fake.messages[0].Role)
	}
	if len(fake.messages[0].Parts) != 1 || fake.messages[0].Parts[0].Text != "系统人设：你好" {
		t.Errorf("第 1 条 system 内容 = %+v，期望文本「系统人设：你好」", fake.messages[0].Parts)
	}
	for i, want := range []entity.ChatMessage{userMsg("第一条"), userMsg("第二条")} {
		gotMsg := fake.messages[i+1]
		if gotMsg.Role != want.Role || len(gotMsg.Parts) != 1 || gotMsg.Parts[0].Text != want.Parts[0].Text {
			t.Errorf("第 %d 条用户消息 = %+v，期望 %+v（原样透传）", i+1, gotMsg, want)
		}
	}
}

// systemPrompt 为空时，应兜底使用 config.DefaultSystemPrompt。
func TestGenerateReplyFallsBackToDefaultSystem(t *testing.T) {
	fake := &fakeAgent{}
	svc := NewAgentService(fake, "")

	if _, err := svc.GenerateReply(context.Background(), nil); err != nil {
		t.Fatalf("GenerateReply 失败: %v", err)
	}
	if len(fake.messages) != 1 {
		t.Fatalf("应仅收到 1 条 system 消息，实际 %d 条", len(fake.messages))
	}
	if fake.messages[0].Role != entity.RoleSystem {
		t.Errorf("角色 = %q，期望 system", fake.messages[0].Role)
	}
	if len(fake.messages[0].Parts) != 1 || fake.messages[0].Parts[0].Text != config.DefaultSystemPrompt {
		t.Errorf("system 内容 = %+v，期望 DefaultSystemPrompt", fake.messages[0].Parts)
	}
}

// 无用户消息时，仍应组装出仅含 system 的消息列表并调用 agent。
func TestGenerateReplyEmptyMessages(t *testing.T) {
	fake := &fakeAgent{}
	svc := NewAgentService(fake, "人设")

	if _, err := svc.GenerateReply(context.Background(), nil); err != nil {
		t.Fatalf("GenerateReply 失败: %v", err)
	}
	if len(fake.messages) != 1 || fake.messages[0].Role != entity.RoleSystem {
		t.Fatalf("消息列表 = %+v，期望仅 1 条 system", fake.messages)
	}
}

// agent 返回错误时应原样透传。
func TestGenerateReplyPropagatesAgentError(t *testing.T) {
	fake := &fakeAgent{err: context.DeadlineExceeded}
	svc := NewAgentService(fake, "人设")

	if _, err := svc.GenerateReply(context.Background(), nil); err != context.DeadlineExceeded {
		t.Errorf("错误 = %v，期望透传 DeadlineExceeded", err)
	}
}
