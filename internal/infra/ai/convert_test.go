package ai

import (
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"

	"plumebot/internal/domain/entity"
)

func strPtr(s string) *string { return &s }

func textPart(text string) entity.ContentPart {
	return entity.ContentPart{Type: entity.PartTypeText, Text: text}
}

func TestToSchemaText(t *testing.T) {
	msgs := []entity.ChatMessage{
		{Role: entity.RoleSystem, Parts: []entity.ContentPart{textPart("你是 PlumeBot。")}},
		{Role: entity.RoleUser, Parts: []entity.ContentPart{textPart("你好")}},
	}

	got, err := ToSchema(msgs)
	if err != nil {
		t.Fatalf("ToSchema 失败: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("消息数 = %d，期望 2", len(got))
	}
	if got[0].Role != schema.System || got[1].Role != schema.User {
		t.Errorf("角色映射错误: %q / %q", got[0].Role, got[1].Role)
	}
	// 单文本 part 走 Content（UserInputMultiContent 仅限 user/tool 角色，eino-ext 转换器限制）
	if got[0].Content != "你是 PlumeBot。" {
		t.Errorf("system 消息 Content = %q，期望透传文本", got[0].Content)
	}
	if got[1].Content != "你好" {
		t.Errorf("user 消息 Content = %q，期望透传文本", got[1].Content)
	}
	if len(got[0].UserInputMultiContent) != 0 || len(got[1].UserInputMultiContent) != 0 {
		t.Errorf("纯文本消息不应携带 UserInputMultiContent: %+v", got)
	}
}

func TestToSchemaImageURL(t *testing.T) {
	msgs := []entity.ChatMessage{{
		Role: entity.RoleUser,
		Parts: []entity.ContentPart{
			textPart("看图"),
			{Type: entity.PartTypeImage, URL: "https://example.com/cat.png", MIMEType: "image/png"},
		},
	}}

	got, err := ToSchema(msgs)
	if err != nil {
		t.Fatalf("ToSchema 失败: %v", err)
	}
	parts := got[0].UserInputMultiContent
	if len(parts) != 2 {
		t.Fatalf("part 数 = %d，期望 2", len(parts))
	}
	if parts[0].Type != schema.ChatMessagePartTypeText {
		t.Errorf("part[0].Type = %q，期望 text", parts[0].Type)
	}
	img := parts[1]
	if img.Type != schema.ChatMessagePartTypeImageURL || img.Image == nil {
		t.Fatalf("part[1] = %+v，期望 ImageURL part", img)
	}
	if img.Image.URL == nil || *img.Image.URL != "https://example.com/cat.png" {
		t.Errorf("Image.URL = %v，期望图片 URL", img.Image.URL)
	}
	if img.Image.Detail != schema.ImageURLDetailAuto {
		t.Errorf("Image.Detail = %q，期望 auto", img.Image.Detail)
	}
}

func TestToSchemaImageBase64(t *testing.T) {
	b64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
	msgs := []entity.ChatMessage{{
		Role: entity.RoleUser,
		Parts: []entity.ContentPart{
			{Type: entity.PartTypeImage, Base64: b64, MIMEType: "image/png"},
		},
	}}

	got, err := ToSchema(msgs)
	if err != nil {
		t.Fatalf("ToSchema 失败: %v", err)
	}
	img := got[0].UserInputMultiContent[0]
	if img.Image == nil || img.Image.Base64Data == nil || *img.Image.Base64Data != b64 {
		t.Fatalf("Image.Base64Data = %v，期望透传 base64", img.Image)
	}
	if img.Image.MIMEType != "image/png" {
		t.Errorf("Image.MIMEType = %q，期望 image/png", img.Image.MIMEType)
	}
	if img.Image.URL != nil {
		t.Errorf("Image.URL 应为 nil，实际 %v", *img.Image.URL)
	}
}

func TestToSchemaAudioVideoFile(t *testing.T) {
	msgs := []entity.ChatMessage{{
		Role: entity.RoleUser,
		Parts: []entity.ContentPart{
			{Type: entity.PartTypeAudio, URL: "https://example.com/a.wav"},
			{Type: entity.PartTypeVideo, Base64: "dmlkZW8=", MIMEType: "video/mp4"},
			{Type: entity.PartTypeFile, URL: "https://example.com/doc.pdf", MIMEType: "application/pdf"},
		},
	}}

	got, err := ToSchema(msgs)
	if err != nil {
		t.Fatalf("ToSchema 失败: %v", err)
	}
	parts := got[0].UserInputMultiContent
	if len(parts) != 3 {
		t.Fatalf("part 数 = %d，期望 3", len(parts))
	}

	audio := parts[0]
	if audio.Type != schema.ChatMessagePartTypeAudioURL || audio.Audio == nil ||
		audio.Audio.URL == nil || *audio.Audio.URL != "https://example.com/a.wav" {
		t.Errorf("audio part = %+v", audio)
	}

	video := parts[1]
	if video.Type != schema.ChatMessagePartTypeVideoURL || video.Video == nil ||
		video.Video.Base64Data == nil || *video.Video.Base64Data != "dmlkZW8=" ||
		video.Video.MIMEType != "video/mp4" {
		t.Errorf("video part = %+v", video)
	}

	file := parts[2]
	if file.Type != schema.ChatMessagePartTypeFileURL || file.File == nil ||
		file.File.URL == nil || *file.File.URL != "https://example.com/doc.pdf" {
		t.Errorf("file part = %+v", file)
	}
}

// 空 Parts 的消息应整体跳过；全空时返回空列表。
func TestToSchemaSkipsEmptyParts(t *testing.T) {
	msgs := []entity.ChatMessage{
		{Role: entity.RoleSystem, Parts: []entity.ContentPart{textPart("人设")}},
		{Role: entity.RoleUser}, // 空 Parts
		{Role: entity.RoleAssistant, Parts: []entity.ContentPart{}},
	}

	got, err := ToSchema(msgs)
	if err != nil {
		t.Fatalf("ToSchema 失败: %v", err)
	}
	if len(got) != 1 || got[0].Role != schema.System {
		t.Fatalf("消息列表 = %+v，期望仅保留 system", got)
	}

	empty, err := ToSchema(nil)
	if err != nil {
		t.Fatalf("ToSchema(nil) 失败: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("ToSchema(nil) = %d 条，期望 0", len(empty))
	}
}

// 错误分支：URL/Base64 双空、双设、未知 PartType、未知 Role。
func TestToSchemaErrors(t *testing.T) {
	cases := []struct {
		name    string
		msg     entity.ChatMessage
		wantSub string
	}{
		{
			name:    "双空",
			msg:     entity.ChatMessage{Role: entity.RoleUser, Parts: []entity.ContentPart{{Type: entity.PartTypeImage}}},
			wantSub: "URL 与 Base64 均为空",
		},
		{
			name: "双设",
			msg: entity.ChatMessage{Role: entity.RoleUser, Parts: []entity.ContentPart{
				{Type: entity.PartTypeImage, URL: "https://a.com/1.png", Base64: "YQ=="},
			}},
			wantSub: "不能同时设置",
		},
		{
			name: "未知 PartType",
			msg: entity.ChatMessage{Role: entity.RoleUser, Parts: []entity.ContentPart{
				{Type: entity.PartType("hologram"), URL: "https://a.com/h"},
			}},
			wantSub: "未知 part 类型",
		},
		{
			name:    "未知 Role",
			msg:     entity.ChatMessage{Role: entity.Role("boss"), Parts: []entity.ContentPart{textPart("hi")}},
			wantSub: "未知角色",
		},
		{
			name: "非文本 part 非 user 角色",
			msg: entity.ChatMessage{Role: entity.RoleSystem, Parts: []entity.ContentPart{
				{Type: entity.PartTypeImage, URL: "https://a.com/1.png"},
			}},
			wantSub: "非文本 part 仅支持 user 角色",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ToSchema([]entity.ChatMessage{tc.msg})
			if err == nil {
				t.Fatal("期望报错，实际无错误")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("错误 = %q，期望包含 %q", err.Error(), tc.wantSub)
			}
		})
	}
}

// 错误信息应带消息序号与 part 序号定位。
func TestToSchemaErrorLocatesIndex(t *testing.T) {
	msgs := []entity.ChatMessage{
		{Role: entity.RoleSystem, Parts: []entity.ContentPart{textPart("ok")}},
		{Role: entity.RoleUser, Parts: []entity.ContentPart{textPart("ok"), {Type: entity.PartTypeImage}}},
	}

	_, err := ToSchema(msgs)
	if err == nil {
		t.Fatal("期望报错，实际无错误")
	}
	if !strings.Contains(err.Error(), "消息 1 part 1") {
		t.Errorf("错误信息应定位到 消息 1 part 1，实际 %q", err.Error())
	}
}

func TestFromSchema(t *testing.T) {
	if got := FromSchema(nil); got != "" {
		t.Errorf("FromSchema(nil) = %q，期望空串", got)
	}
	if got := FromSchema(schema.AssistantMessage("你好，PlumeBot。", nil)); got != "你好，PlumeBot。" {
		t.Errorf("FromSchema = %q，期望透传 Content", got)
	}
	if got := FromSchema(&schema.Message{Role: schema.Assistant}); got != "" {
		t.Errorf("空 Content = %q，期望空串", got)
	}
}
