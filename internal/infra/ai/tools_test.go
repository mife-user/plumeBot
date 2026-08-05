package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// emptyArgs 是 fakeTool 的空参数结构。
type emptyArgs struct{}

// newFakeTool 创建仅用于验证注册/过滤机制的工具（无真实逻辑）。
func newFakeTool(name string) tool.BaseTool {
	params, err := toolutils.GoStruct2ParamsOneOf[emptyArgs]()
	if err != nil {
		panic(err)
	}
	return toolutils.NewTool(&schema.ToolInfo{Name: name, Desc: "工具 " + name, ParamsOneOf: params},
		func(context.Context, emptyArgs) (string, error) { return "", nil })
}

func TestToolsRegistryRegisterDuplicate(t *testing.T) {
	tr := NewToolsRegistry()
	if err := tr.Register("echo", newFakeTool("echo")); err != nil {
		t.Fatalf("首次注册失败: %v", err)
	}
	err := tr.Register("echo", newFakeTool("echo"))
	if err == nil || !strings.Contains(err.Error(), "已注册") {
		t.Errorf("重名注册应报错，实际 %v", err)
	}
}

func TestToolsRegistryEnabledFilter(t *testing.T) {
	tr := NewToolsRegistry()
	for _, name := range []string{"a", "b", "c"} {
		if err := tr.Register(name, newFakeTool(name)); err != nil {
			t.Fatalf("注册 %s 失败: %v", name, err)
		}
	}

	got, err := tr.EnabledTools([]string{"c", "a"})
	if err != nil {
		t.Fatalf("EnabledTools 失败: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("工具数 = %d，期望 2", len(got))
	}
	info, _ := got[0].Info(context.Background())
	if info.Name != "c" {
		t.Errorf("got[0] = %q，期望 c（保持启用列表顺序）", info.Name)
	}
	info2, _ := got[1].Info(context.Background())
	if info2.Name != "a" {
		t.Errorf("got[1] = %q，期望 a", info2.Name)
	}
}

func TestToolsRegistryEnabledUnknown(t *testing.T) {
	tr := NewToolsRegistry()
	if err := tr.Register("a", newFakeTool("a")); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	_, err := tr.EnabledTools([]string{"nope"})
	if err == nil {
		t.Fatal("未注册工具应报错")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "已注册：a") {
		t.Errorf("错误应包含工具名与已注册列表，实际 %q", err.Error())
	}
}

func TestToolsRegistryEnabledEmpty(t *testing.T) {
	tr := NewToolsRegistry()

	got, err := tr.EnabledTools(nil)
	if err != nil {
		t.Fatalf("空列表不应报错: %v", err)
	}
	if got != nil {
		t.Errorf("空列表应返回 nil，实际 %v", got)
	}
}

func TestToolsRegistryEnabledDedupe(t *testing.T) {
	tr := NewToolsRegistry()
	if err := tr.Register("a", newFakeTool("a")); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	got, err := tr.EnabledTools([]string{"a", "a", "a"})
	if err != nil {
		t.Fatalf("EnabledTools 失败: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("重复名应去重，实际 %d 个", len(got))
	}
}
