package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"plumebot/internal/domain"
	"plumebot/internal/domain/entity"
)

// defaultMaxIterations 是 tool 自动循环上限（adk 默认同为 20，显式设置防止未来默认值变化）。
const defaultMaxIterations = 20

// errNoModelOutput 表示推理结束但未收到任何模型输出消息。
var errNoModelOutput = errors.New("模型未返回任何消息")

// 编译期校验：EinoAgent 实现 domain.Agent。
var _ domain.Agent = (*EinoAgent)(nil)

// EinoAgent 是基于 eino ChatModelAgent 的 domain.Agent 实现。
// 系统提示词由消息列表携带（Instruction 留空，组装在 service 层），工具按需注入。
type EinoAgent struct {
	agent *adk.ChatModelAgent
}

// NewEinoAgent 组装 eino ChatModelAgent。
// tools 为空时不注入 ToolsConfig（空 ToolsNodeConfig 行为未验证，不冒险）；
// MaxIterations 显式兜底 defaultMaxIterations。
func NewEinoAgent(ctx context.Context, cm model.BaseChatModel, tools []tool.BaseTool) (*EinoAgent, error) {
	cfg := &adk.ChatModelAgentConfig{
		Model:         cm,
		MaxIterations: defaultMaxIterations,
	}
	if len(tools) > 0 {
		cfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		}
	}

	agent, err := adk.NewChatModelAgent(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("组装 ChatModelAgent 失败: %w", err)
	}
	return &EinoAgent{agent: agent}, nil
}

// Generate 将业务消息转换为 eino 消息并执行推理，返回最终文本回复。
// 事件迭代中第一个错误（含超时、tool 循环超限、工具执行失败）直接透传；
// 无任何模型输出消息时报错；有输出但 Content 为空 → 返回空串（P6 再议）。
func (e *EinoAgent) Generate(ctx context.Context, msgs []entity.ChatMessage) (string, error) {
	schemaMsgs, err := ToSchema(msgs)
	if err != nil {
		return "", err
	}

	iter := e.agent.Run(ctx, &adk.AgentInput{Messages: schemaMsgs})

	var last *schema.Message
	for {
		ev, ok := iter.Next()
		if !ok {
			break
		}
		if ev.Err != nil {
			return "", ev.Err
		}
		if ev.Output != nil && ev.Output.MessageOutput != nil && !ev.Output.MessageOutput.IsStreaming {
			last = ev.Output.MessageOutput.Message
		}
	}
	if last == nil {
		return "", errNoModelOutput
	}
	return FromSchema(last), nil
}
