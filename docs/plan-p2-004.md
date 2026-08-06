# P2-004 eino Agent 接入 — 实施计划

> 状态：**实施中**（阶段 1~4 已完成并全绿，待阶段 5 main 接入收尾）
> 关联任务：roadmap P2-004 / B-001
> 依据文档：docs/architecture.md（§2 技术选型、§6 Prompt 组装顺序、§14 分层与启动流程）、AGENTS.md
> 调研基准：eino v0.8.13 / eino-ext components/model/openai v0.1.13（源码级核实，2026-08）

---

## 1. 任务目标与验收

### 1.1 目标（roadmap P2-004 原文）

实现 domain.Agent 接口，封装 eino ChatModelAgent，支持 tool calling。
验收：传入简单 prompt 可收到 LLM 文本回复。

### 1.2 扩展验收（本计划补充）

- **多模态**：图片（URL / base64 两路）可转换为 eino 多模态输入并送达 LLM（链路单测验证；真实 API 可选）。
- **tool calling**：机制就绪（注册表 + ToolInfo + 自动循环），内置工具为空，不提前实现任何业务工具。
- **LLM 注册中心**：config 按 `llm.provider` 分发；当前仅实现 `openai`（OpenAI 兼容端点，可指向 DeepSeek/Ollama/OpenAI 等）；未知 provider 启动报错。
- **质量**：`go build ./... && go vet ./... && go test ./...` 全绿；main 启动即注入真实 Agent，无 panic。

---

## 2. eino 调研结论（事实速查，源码级核实）

### 2.1 版本现状

| 项 | 值 |
|---|---|
| eino 最新稳定版 | v0.9.13（2026-07 发布）——`ChatModelAgent` 迁入新 `adk` 包，API 大重构（Runner + AsyncIterator 事件流） |
| eino v0.8 线 | v0.8.13（本计划**锁定版本**，0.8 线共 14 个 tag，成熟） |
| eino-ext openai | v0.1.13（最新）；其 go.mod require `eino v0.7.13`——**eino-ext 生态滞后于 eino 主线** |

### 2.2 多模态消息字段（v0.8.13 已存在，非废弃）

```go
schema.Message{
    Content                  string
    Role                     schema.RoleType
    ToolCalls                []schema.ToolCall
    ToolCallID               string
    ToolName                 string
    UserInputMultiContent    []schema.MessageInputPart    // ← 多模态输入走这里
    AssistantGenMultiContent []schema.MessageOutputPart
    MultiContent             []schema.ChatMessagePart     // Deprecated，不用
    ReasoningContent         string
    ResponseMeta             *schema.ResponseMeta
    Extra                    map[string]any
}

schema.MessageInputPart{
    Type  schema.MessagePartType   // ChatMessagePartTypeText / ImageURL / AudioURL / VideoURL / FileURL
    Text  string
    Image *schema.MessageInputImage
    Audio *schema.MessageInputAudio
    Video *schema.MessageInputVideo
    File  *schema.MessageInputFile
}

schema.MessageInputImage{
    schema.MessagePartCommon   // URL *string / Base64Data *string（base64 字符串）/ MIMEType string
    Detail schema.ImageURLDetail // auto / low / high
}
```

官方图片输入形态（eino-ext examples/generate_with_image 已核实）：

```go
msg := &schema.Message{Role: schema.User, UserInputMultiContent: []schema.MessageInputPart{
    {Type: schema.ChatMessagePartTypeText, Text: "看图说话"},
    {Type: schema.ChatMessagePartTypeImageURL, Image: &schema.MessageInputImage{
        MessagePartCommon: schema.MessagePartCommon{URL: &imgURL},
        Detail:            schema.ImageURLDetailAuto,
    }},
}}
```

### 2.3 ChatModelAgent 与 tool（v0.8.13 实测形态，阶段 1 spike 已核实）

- **v0.8.13 中 ChatModelAgent 位于 `adk` 包**（`adk/chatmodel.go`；`flow/agent` 无 ChatModelAgent）。调用形态：
  - `adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{Name, Description, Instruction, Model, ToolsConfig, MaxIterations, ...})`——`Instruction` 留空 = 不注入系统提示词（本项目 system 由消息列表携带）
  - `agent.Run(ctx, &adk.AgentInput{Messages}) *adk.AsyncIterator[*adk.AgentEvent]`——迭代 `ev.Output.MessageOutput.Message`（非流式）取最终答案；`ev.Err` 报错
  - `MaxIterations` 为 tool 循环上限（默认 20），对应原计划中"MaxStep 兜底"
- tool 定义（v0.8.13，描述与执行分离）：
  - `schema.ToolInfo{Name, Desc, Extra, ParamsOneOf}`——**只有描述**（v0.9 移除的 `Bound/InvokableRun` 在 v0.8 的 schema.ToolInfo 上同样不存在）
  - 执行接口 `tool.InvokableTool.InvokableRun(ctx, argumentsInJSON string, opts...) (string, error)`；创建用 `toolutils.NewTool(desc, fn)` / `toolutils.InferTool[T,D](name, desc, fn)`
  - `ParamsOneOf` 构建：`toolutils.GoStruct2ParamsOneOf[T]()`（struct + `jsonschema` tag）或 `schema.NewParamsOneOfByParams/NewParamsOneOfByJSONSchema`
- 详细代码与字段表见 `docs/eino-notes.md`（阶段 1 产出）。

### 2.4 eino-ext openai ChatModel（字段已核实）

```go
openai.ChatModelConfig{
    APIKey              string
    Timeout             time.Duration
    BaseURL             string
    Model               string
    MaxTokens           *int          // Deprecated → 用 MaxCompletionTokens
    MaxCompletionTokens *int
    Temperature         *float32
    TopP                *float32
    Stop                []string
    PresencePenalty     *float32
    FrequencyPenalty    *float32
    Seed                *int
    ResponseFormat      *ChatCompletionResponseFormat
    ReasoningEffort     ReasoningEffortLevel
    // 注意：无 Headers 字段；自定义 header 用 openai.WithExtraHeader(map[string]string) 选项
}
func NewChatModel(ctx context.Context, config *ChatModelConfig) (*ChatModel, error)
```

- 图片输入自动转 OpenAI `image_url` part（URL 直传或 base64 data: URL），官方示例已证实。
- ChatModel 接口（供测试 fake 实现）：以 spike 核实的 v0.8 定义为准。

---

## 3. 版本决策：为什么不用 eino 最新版 v0.9.13

1. **生态不同步**：eino-ext（模型供应商实现）main 分支仍 require `eino v0.7.13`；v0.9 的 schema 大改（`ToolInfo` 移除 `Bound/InvokableRun`、agent 迁移 adk），与最新 eino-ext 组合存在编译/行为不兼容风险。
2. **API 形态复杂**：v0.9 取最终文本需 `Runner + AsyncIterator[*AgentEvent]` 事件循环（`TypedAgentEvent{Output, Action, Err}` 多态结构）；v0.8 是 `Generate(ctx, msgs) (*schema.Message, error)` 一次调用。本需求只是"消息进、文本出"的薄门面，选简单形态。
3. **文档与示例成熟度**：v0.9 发布太新，官方示例仓库与社区资料几乎全是 flow/agent API，踩坑成本高。
4. **功能无缺失**：本项目需要的多模态字段（`UserInputMultiContent/MessageInputImage`）、tool 自动循环、SystemPrompt 在 v0.8.13 全部可用且非废弃。
5. **迁移成本可控**：`domain.Agent` 接口是门面，infra/ai 内部换实现不影响上层；待 eino-ext 跟进 v0.9 且 adk 稳定后评估迁移（记入遗留事项 B-008）。
6. **回退策略**：阶段 1 spike 若发现 v0.8.13 flow/agent 不可用或与预期不符，回退评估 v0.9 adk，更新本文件 §5.1 后再继续。

---

## 4. 架构决策：prompt 模板 / memory / 人格 的边界

### 4.1 概念区分

| 概念 | 本质 | 归属 |
|---|---|---|
| 模板（eino ChatTemplate / Go 拼接） | 纯函数，"怎么拼"，无状态 | 组装层（service） |
| memory 层 | 有状态数据源，"拼什么" | service/memory（P3） |
| 人格 | 动态数据（每群不同、可演化），本质是 system 消息的内容 | service/persona（P4）生产，service/agent 消费 |

### 4.2 结论（本计划采用）

1. **模板不替代 memory**：memory 的核心职责是数据生命周期——窗口滚动（20→100 轮）、画像按需加载/延迟淘汰、压缩触发与摘要 FIFO 融合。这些是模板（纯渲染函数）无法承载的。模板即使消费历史数据，也只是 memory 数据的一个**只读视图**。若想用模板"省掉"memory 而每次全量查 SQLite 渲染，就丢掉了 P3 的全部设计。
2. **memory 存结构化数据，不做渲染**：窗口存 `entity.ChatMessage`，渲染发生在组装层。同一份窗口数据未来要喂给三种消费（对话 prompt、摘要压缩 prompt、检索），一个数据源 + 多个渲染目标 ⇒ 渲染必须在外部。
3. **对话 prompt 组装放 service/agent**（对齐架构 §6 五段顺序、§14.4 的 context.Build 职责）：用 Go 原生拼接，**不引入 eino ChatTemplate**——eino 模板类型是 infra 依赖，service 层不能 import（分层规则 5.2）；把组装逻辑塞进 infra 又违反"业务编排在 service"（规则 5.3）。eino 的 `StateModifier/ChatTemplate` 记为"已知但不用"，机制最小化。
4. **人格不进 agent 层**：`domain.Agent.Generate` 收**完整消息列表（含 system role）**，人格文本由 service/agent 在组装时从 persona service 获取（P4-001 后）；P2-004 用配置 `prompt.system` 静态占位。infra 构造参数不绑 SystemPrompt，保证 P4 人格动态化时 **infra 零改动**。人格的"更新"（update_persona tool）与"注入"（system 内容）是两条路：注入走 service 组装，更新走 P4 的 tool 机制。
5. **模板跟随数据所有者**：摘要压缩 prompt 的数据只来自窗口 ⇒ 归 service/memory 内部（P3-003）；对话 prompt 的数据横跨 persona/memory 两个 service ⇒ 归 service/agent 编排（P6-001 完整实现，P2-004 最小实现）。

### 4.3 备选方案（不采纳）

- **B：eino StateModifier/ChatTemplate 在 infra 内动态注入** —— 人格/画像/摘要数据在 service 层，需把数据管道穿透到 infra 或每次调用闭包注入，耦合 infra 与业务组装，且难单测。
- **C：模板放 memory 层** —— memory 变成"数据 + 渲染"双职责；且五段组装的数据横跨 persona/memory 两个 service，memory 拼不了人格。

### 4.4 对 P2-004 的落点

- `prompt.system` 配置 = **机器人自身系统提示词**（人设占位文案，静态）；`GenerateReply` 最小组装 = `[system] + 用户消息`；完整五段组装属 P6-001。
- 不引入 eino ChatTemplate / StateModifier / MessagesTemplate。

---

## 5. 范围边界（明确不做）

- 不实现任何业务工具（记忆 tool → P3-004；人格 tool → P4-001）；`tools.enabled` 默认空。
- 不接事件管线尾端：`tailHandler` 保持 stub；端到端验证 = agent 直调（群聊闭环 → P6-002）。
- 不改 `entity.Message`（OneBot 层消息模型不动，避免波及 P2-002/003 已验证链路）；OneBot 图片段 → ChatMessage 映射留到 P6 接线（遗留 B-009）。
- 中间件链 / SQLite / onebot 均不动；不引入 eino ChatTemplate、StateModifier。
- domain 层保持零第三方 import。

---

## 6. 拆分规划（5 阶段，按序执行）

### 阶段 1：eino 依赖接入与 API 冒烟验证（spike）

**目标**：锁定版本可编译；文本/图片/tool 三路冒烟跑通；产出 API 速查笔记。

| 步骤 | 内容 |
|---|---|
| 1.1 | `go get github.com/cloudwego/eino@v0.8.13` |
| 1.2 | `go get github.com/cloudwego/eino-ext/components/model/openai@latest`；`go mod tidy`；`go build ./...` 确认依赖解析与编译 |
| 1.3 | 核实 v0.8.13 flow/agent 确切 API：`NewChatModelAgent` 签名、`AgentConfig` 字段（Model/SystemPrompt/ToolsConfig/StateModifier/MaxStep）、`Generate` 返回、`ParamsOneOf` 构建方式、ChatModel 接口定义位置（供 fake 实现） |
| 1.4 | 临时 spike 测试 `internal/infra/ai/spike_test.go`（验证后沉淀为阶段 4 正式单测）：<br>① 文本冒烟：fake ChatModel → NewChatModelAgent → Generate → 断言最终文本<br>② 图片冒烟：构造 `UserInputMultiContent`（URL 与 base64 两路）→ 断言转换结构<br>③ tool 冒烟：注册临时 tool → fake 模型先回 tool_call 再回最终答案 → 断言自动循环执行<br>④ 真实 API 冒烟（env 门控 `PLUMEBOT_TEST_LLM=1`，可选）：文本 + 图片 URL 各一次 |
| 1.5 | 产出 `docs/eino-notes.md`：多模态字段速查、构造/调用最小代码、openai 配置字段、参考 URL（防止后续阶段重新调研） |

**验收**：`go build ./...` 通过；冒烟测试绿；eino-notes.md 落盘。
**涉及文件**：go.mod、go.sum、docs/eino-notes.md（新增）、internal/infra/ai/spike_test.go（临时）。

### 阶段 2：配置扩展（llm / tools / prompt 三段）

**目标**：config.yaml 支持 LLM 注册中心、工具开关、系统提示词；模板与根配置双处同步（B-006 流程）。

| 步骤 | 内容 |
|---|---|
| 2.1 | `pkg/config/config.go` 新增结构（mapstructure tag + 注释）：<br>`LLMConfig{Provider string, TimeoutSeconds int, OpenAI OpenAICompatConfig}`<br>`OpenAICompatConfig{BaseURL string, APIKey string, Model string, Temperature *float64, MaxTokens int}`<br>`ToolsConfig{Enabled []string}`<br>`PromptConfig{System string}` |
| 2.2 | `Config` 根结构增加 `LLM LLMConfig`、`Tools ToolsConfig`、`Prompt PromptConfig` |
| 2.3 | 默认常量下沉 `pkg/config`（消费方兜底原则，config 层不改写字段）：<br>`DefaultLLMProvider = "openai"`；`DefaultOpenAIBaseURL = "https://api.openai.com/v1"`；`DefaultLLMTimeoutSeconds = 60`；`DefaultSystemPrompt`（机器人人设占位文案）<br>兜底语义：provider 空 → openai；base_url 空 → 默认；api_key 空 → 透传（本地 Ollama 场景）；timeout_seconds ≤0 → 60；temperature nil → 不传（模型默认）；max_tokens ≤0 → 不传；model 空 → **构造期报错**（不可猜）；prompt.system 空 → DefaultSystemPrompt |
| 2.4 | `config.default.yaml` 与根 `config.yaml` **双处同步**：llm 段示例 `base_url: https://api.deepseek.com/v1`、`model: deepseek-chat`、`api_key: ""`；`tools.enabled: []`；`prompt.system: <默认人设文案>` |
| 2.5 | 单测 `pkg/config/config_test.go`：三段解析、空值兜底（provider/base_url/timeout/system）、temperature 指针语义、未知字段忽略 |

**验收**：`go test ./pkg/config/` 通过；`go build ./...` 通过。
**涉及文件**：pkg/config/config.go、pkg/config/config.default.yaml、config.yaml（根）、pkg/config/config_test.go（新增）。

### 阶段 3：domain 实体与接口升级（多模态消息结构）

**目标**：LLM 会话消息实体 + Agent 接口改造，落实"组装在 service、执行在 infra"。

| 步骤 | 内容 |
|---|---|
| 3.1 | 新增 `internal/domain/entity/content.go`（纯数据结构，零依赖）：<br>`Role` 常量：system / user / assistant<br>`PartType` 常量：text / image / audio / video / file<br>`ContentPart{Type PartType, Text string, URL string, Base64 string, MIMEType string}`（URL 与 Base64 二选一）<br>`ChatMessage{Role Role, Parts []ContentPart}` |
| 3.2 | 改造 `internal/domain/agent.go`：`Generate(ctx context.Context, messages []entity.ChatMessage) (string, error)`——完整消息列表（含 system role）由调用方组装，infra 只执行 |
| 3.3 | 同步签名：`infra/ai/agent.go`（AgentStub）；`service/agent/agent.go`：<br>`GenerateReply(ctx, msgs []entity.ChatMessage) (string, error)`，内部最小组装 = system（cfg.Prompt.System 或默认常量）→ 追加用户消息 → agent.Generate；构造函数增加 systemPrompt 参数（main 注入） |
| 3.4 | 单测：service/agent 用 fake domain.Agent 验证 system 前置与消息透传 |

**验收**：`go build ./...`；`go test ./internal/service/agent/` 通过。
**涉及文件**：internal/domain/entity/content.go（新增）、internal/domain/agent.go、internal/infra/ai/agent.go、internal/service/agent/agent.go（+测试）。

### 阶段 4：infra/ai：eino Agent 实现

**目标**：注册中心 + 多模态转换 + ChatModelAgent 封装 + tool 机制，全部可单测（fake ChatModel，不发网络）。

| 步骤 | 内容 |
|---|---|
| 4.1 | `provider.go` 注册中心：`type Factory func(ctx, cfg config.Config) (domain.Agent, error)`；`Register(name, f)`；`NewAgent(ctx, cfg)` 按 Provider 分发，未知 provider 返回明确错误（含已注册列表）。注册 `openai`：`openai.NewChatModel`（映射 BaseURL/APIKey/Model/Temperature/MaxTokens/Timeout，`*float64 → *float32`，空值按 §阶段 2 兜底）→ 包装为 EinoAgent。<br>**⚠ 已确认偏差（用户拍板）**：① Registry 为注入式实例（`NewRegistry()`，非包级全局单例），依赖经 main 组装注入；② Factory 接收**完整 `config.Config`**（非仅 LLMConfig），工具启用列表在 Tools 段、LLM 配置在 LLM 段，二者组装点即工厂；③ 工具表经 `NewOpenAIFactory(tr *ToolsRegistry)` 闭包注入（Factory 签名固定，tools 解析依赖用闭包捕获） |
| 4.2 | `convert.go` 纯函数：`ToSchema([]entity.ChatMessage) []*schema.Message`——text → `{Type: Text, Text}`；image → `{Type: ImageURL, Image: {URL 或 Base64Data, MIMEType, Detail: auto}}`；audio/video/file 同构；空 Parts 跳过；未知 Type 报错。<br>`FromSchema(*schema.Message) string`——取最终文本 Content（tool 循环后为最终 assistant 消息；空则返回空串，P6 再议） |
| 4.3 | `agent.go`：EinoAgent 实现 domain.Agent——构造 `adk.NewChatModelAgent`（Instruction 留空，消息列表已含 system；ToolsConfig 注入启用工具；MaxIterations 兜底，默认 20）；Generate = ToSchema → `agent.Run` 事件迭代 → FromSchema；错误透传（含超时、循环超限） |
| 4.4 | `tools.go` 工具机制：内置注册表 `map[string]tool.BaseTool`（本期空）；`RegisterTool(name, t)`（供 P3-004/P4-001）；`EnabledTools(enabled []string) []tool.BaseTool`——按 cfg.Tools.Enabled 过滤，enabled 含未注册名 → 构造期报错；执行形态为 `toolutils.NewTool(desc, fn)` 包装 |
| 4.5 | 单测（fake ChatModel）：convert 全分支（文本/图片 URL/图片 base64/混合/空 Parts/未知 Type）；agent 文本往返；tool 自动循环（fake 先回 tool_call → 断言 InvokableRun 被调 → 再回最终答案）；错误透传；provider 注册与未知 provider 报错 |

**验收**：`go build ./... && go vet ./... && go test ./...` 全绿。
**涉及文件**：internal/infra/ai/{provider.go, convert.go, agent.go, tools.go}（新增/重写）+ *_test.go。

### 阶段 5：main 组装 + 端到端验收 + 台账收尾

**目标**：真实 LLM 链路打通，文档与台账同步。

| 步骤 | 内容 |
|---|---|
| 5.1 | `cmd/bot/main.go`：`agentInfra, err := ai.NewAgent(ctx, cfg.LLM)`（失败 → logger.Fatal 退出）；`agentSvc := agent.NewAgentService(agentInfra, cfg.Prompt.System)`；删除 AgentStub 引用 |
| 5.2 | 启动日志：provider / model / base_url（**不打印 api_key**） |
| 5.3 | 端到端验收（真实 API，需在 config.yaml 填 api_key 与 model）：临时验证入口（测试 env 门控）传 `[system 默认人设, user 你好]` → 收到 LLM 文本回复；可选：图片 URL 输入验证多模态链路 |
| 5.4 | 文档收尾：roadmap P2-004 ✅、删除 B-001、新增遗留事项候选 B-008（eino 版本升级观察项：eino-ext 跟进 v0.9 后评估 adk 迁移）、B-009（OneBot 图片段 → ChatMessage 映射，P6 接线实现；NapCat 图片 URL 可达性需验证）；AGENTS.md 更新第二阶段状态（P2-004 完成，阶段禁令随 P3 调整） |
| 5.5 | 回归：`go build ./... && go vet ./... && go test ./...` |

**验收**：真实 API 返回 LLM 回复；全量构建/测试绿；台账更新。
**涉及文件**：cmd/bot/main.go、docs/roadmap.md、AGENTS.md、（可选）config.yaml。

---

## 7. 涉及文件总表

| 文件 | 变更类型 | 阶段 |
|---|---|---|
| go.mod / go.sum | 新增依赖（eino v0.8.13、eino-ext openai） | 1 |
| docs/eino-notes.md | 新增（调研笔记） | 1 |
| pkg/config/config.go | 扩展（LLM/Tools/Prompt 配置） | 2 |
| pkg/config/config.default.yaml | 扩展（双处同步） | 2 |
| config.yaml（根） | 扩展（双处同步） | 2 |
| pkg/config/config_test.go | 新增 | 2 |
| internal/domain/entity/content.go | 新增（ChatMessage/ContentPart） | 3 |
| internal/domain/agent.go | 改造（Generate 签名） | 3 |
| internal/infra/ai/agent.go | 改造（stub 签名 → 阶段 4 重写） | 3/4 |
| internal/infra/ai/provider.go / convert.go / tools.go | 新增 | 4 |
| internal/service/agent/agent.go | 改造（GenerateReply + system 组装） | 3 |
| cmd/bot/main.go | 改造（真实注入） | 5 |
| docs/roadmap.md / AGENTS.md | 更新（P2-004 ✅、B 表调整） | 5 |

---

## 8. 风险与回退

| 风险 | 概率 | 应对 |
|---|---|---|
| v0.8.13 flow/agent API 与预期不符 | 低 | **已触发并解决**：v0.8.13 的 ChatModelAgent 位于 adk 包（非 flow/agent），调用为 Runner 事件迭代；已按实测形态修正 §2.3 与阶段 4，eino-ext v0.1.13 与 eino v0.8.13 编译、冒烟全部通过，未触发 v0.9 回退 |
| eino-ext openai 与 eino v0.8.13 编译不兼容 | 低 | 依赖解析后立即 build 验证；必要时在 v0.8.x 内微调版本 |
| fake ChatModel 接口签名与 v0.8 实际不符 | 中 | 以 spike 核实为准；测试替身只依赖接口 |
| 真实 API 图片链路（NapCat 图片 URL 可达性） | 中 | P2-004 用 base64/本地 URL 验证转换层；NapCat 侧问题记入 B-009 |

---

## 9. 评审结论（2026-08 已确认）

- [x] **版本策略：锁定 eino v0.8.13**（理由见 §3；v0.9 adk 列为 B-008 观察项）
- [x] **prompt 架构：方案 A（实施期修订为 A'）**——系统提示词（机器人固定人设）经 `ChatModelAgentConfig.Instruction` 注入，由 provider 工厂传 `cfg.Prompt.System`（空值兜底 `config.DefaultSystemPrompt`）；service/agent 只透传业务消息（移除 system 前置逻辑）。修订理由（用户拍板 2026-08）：机器人系统提示词全局固定，与 P4 人格系统（群聊成员画像，走消息列表）无耦合，Instruction 是 eino 原生机制；Name/Description 同时补齐（multi-agent 铺路）。eino ChatTemplate/StateModifier 仍未引入
- [x] **temperature / max_tokens 未配置语义：不传（模型默认）**
- [x] **新增遗留事项 B-008、B-009 采纳**（阶段 5 写入 roadmap）
