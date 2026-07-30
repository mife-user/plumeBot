# PlumeBot 开发阶段规划

> 单人项目，按架构分层逐步构建。每个阶段完成后进入下一阶段，不跨阶段开发。

---

## 总览

| 阶段 | 内容 | 产出 |
|------|------|------|
| 第一阶段 | 项目骨架 | 可编译、零外部依赖的空框架 |
| 第二阶段 | 基础设施接入 | ZeroBot 连通 NapCat + eino 可用 + SQLite 落盘 |
| 第三阶段 | 记忆系统 | 上下文窗口 + 画像 + 压缩摘要 |
| 第四阶段 | 人格与插件 | 人格 extend 链 + .so 插件加载 |
| 第五阶段 | 触发控制 | mention/auto 模式 + 状态规则 |
| 第六阶段 | 联调验收 | 完整消息链路跑通，bot 可对话 |

---

## 第一阶段｜项目骨架

目标：Go DDD 分层工程骨架，可编译，不接入任何外部服务。

| 任务编号 | 任务名 | 内容 | 优先级 | 涉及模块 | 启动条件 | 验收标准 | 状态 |
|---|---------|------|:---:|------|------|------|:--:|
| P1-001 | Go module 初始化 | `go mod init plumebot`，创建目录骨架（cmd/、internal/domain/、internal/service/、internal/handler/、internal/infra/、pkg/），写入 .gitignore (Go 标准) | P0 | 根工程 | 无 | `go mod tidy` 无报错；目录结构匹配 AGENTS.md | ✅ |
| P1-002 | 配置加载 (pkg/config) | 定义 Config 结构体（Bot、Log、Control），使用 viper 从 config.yaml 加载，Go 侧仅兜底默认值 | P0 | pkg/config | P1-001 完成 | `go build ./pkg/config/` 通过；加载 config.yaml 不 panic | ✅ |
| P1-003 | domain 接口定义 | 定义 domain.Agent、domain.Memory、domain.Persona、domain.Plugin、domain.Storage、domain.Control 六个核心接口，零外部 import | P0 | internal/domain | P1-001 完成 | 每个接口 1-3 个方法签名，仅依赖标准库和 domain/entity；`go build ./internal/domain/` 通过 | ✅ |
| P1-004 | domain 实体定义 | 定义 entity.Message、entity.Event、entity.MemberProfile、entity.GroupProfile、entity.Persona 等公共结构体 | P0 | internal/domain/entity | P1-001 完成 | 实体字段对齐架构文档 4、7、9 节；纯 struct 无方法限制 | ✅ |
| P1-005 | infra 空壳 | 为每个 domain 接口创建 stub 实现：返回 nil 或 error，放置于 internal/infra/{ai,sqlite,onebot,plugin_so,plugin_exe}/ | P0 | internal/infra | P1-003 完成 | 每个包至少一个文件实现对应接口；`go build ./internal/infra/...` 通过 | ⬜ |
| P1-006 | service 编排框架 | 创建各 service 包的构造函数，接收 domain 接口依赖并持有；方法体返回 nil 或 error stub | P0 | internal/service | P1-003、P1-005 完成 | 每个 service 构造函数可注入 stub 实现；`go build ./internal/service/...` 通过 | ⬜ |
| P1-007 | handler 空壳 | 创建 message.go 和 notice.go，持有 service 引用，方法 stub | P0 | internal/handler | P1-006 完成 | `go build ./internal/handler/` 通过 | ⬜ |
| P1-008 | main.go 组装 | 在 cmd/bot/main.go 中完成：加载配置 → 创建 infra stub → 注入 service → 注入 handler → `select{}` 阻塞（不实际连接任何服务） | P0 | cmd/bot | P1-001～007 全部完成 | `go build ./cmd/bot/` 通过；`go vet ./...` 无警告；可编译出空骨架 | ⬜ |

---

## 第二阶段｜基础设施接入

目标：ZeroBot 连通 NapCat 可收发消息，eino Agent 可调用 LLM，SQLite 可读写。

| 任务编号 | 任务名 | 内容 | 优先级 | 涉及模块 | 启动条件 | 验收标准 |
|---|---------|------|:---:|------|------|------|
| P2-001 | SQLite 存储层 | 实现 domain.Storage 接口：建表（8 张表：messages、group_profile、group_jargon、member_profile、member_facts、persona、bot_state、plugin_config）、CRUD；启动时自动建表 + 插入默认人格 | P0 | infra/sqlite | P1 全部完成 | 可写入/查询 messages；默认人格 groupid=0 自动插入 |
| P2-002 | ZeroBot 连接层 | 实现 OneBot WebSocket 客户端，连接 NapCat，接收原始事件 → 转为 domain.Event → 交给 handler | P0 | infra/onebot | P1 全部完成 | bot 启动后连上 NapCat，能收到群消息事件并打印日志 |
| P2-003 | 消息管线中间件 | 实现中间件链（日志 → 限流 → 敏感词过滤），在 event service 中编织 | P0 | service/event | P2-002 完成 | 每条消息有日志输出；限流超限丢弃；敏感词命中拦截且记录 |
| P2-004 | eino Agent 接入 | 实现 domain.Agent 接口，封装 eino ChatModelAgent，支持 tool calling | P0 | infra/ai | P1 全部完成 | 传入简单 prompt 可收到 LLM 文本回复 |

---

## 第三阶段｜记忆系统

目标：上下文窗口滚动、画像按需加载缓存、摘要压缩流水线。

| 任务编号 | 任务名 | 内容 | 优先级 | 涉及模块 | 启动条件 | 验收标准 |
|---|---------|------|:---:|------|------|------|
| P3-001 | 上下文窗口 | 实现 ring buffer 窗口：初始 20 轮、上限 100 轮；消息自动追加；窗口满触发压缩信号 | P0 | service/memory | P2-001、P2-002 | 群聊消息持续追加窗口；达到 100 轮时返回压缩触发信号 |
| P3-002 | 画像加载与缓存 | 实现群聊个人画像和群画像的按需加载 + 内存缓存 + 延迟淘汰（N 轮后移除）；非窗口人物不加载 | P0 | service/memory + infra/sqlite | P3-001、P2-001 | 窗口中人物画像在首次出现时加载到缓存；消失后延迟 N 轮淘汰 |
| P3-003 | 窗口压缩策略 | 实现一级压缩（LLM 摘要 + 关键词提取）、二级压缩（多摘要融合）、摘要淘汰（FIFO）；摘要存内存 | P0 | service/memory | P3-001、P2-004 | 100 轮触发摘要生成；多摘要达到上限后二次融合；摘要总数有上限 |
| P3-004 | 记忆更新闭环 | 实现 eino Tool：store_fact、learn_jargon 等，Agent 对话中自行调用更新 SQLite | P1 | infra/ai/tools | P3-002、P2-004 | Agent 在对话中能通过 tool calling 存储事实；黑话学习标记待确认状态 |

---

## 第四阶段｜人格与插件

目标：多群人设隔离、.so 插件动态加载运行。

| 任务编号 | 任务名 | 内容 | 优先级 | 涉及模块 | 启动条件 | 验收标准 |
|---|---------|------|:---:|------|------|------|
| P4-001 | 人格系统 | 实现 persona 表的 extend 链查询与合并；Agent 通过 update_persona tool 驱动人格更新（内存生效 + 异步写 SQLite） | P0 | service/persona + infra/sqlite | P2-001、P2-004 | 按 userid+groupid 查询可正确合并父级人格；Agent 调 tool 可更新并生效 |
| P4-002 | .so 插件加载 | 实现 domain.Plugin 接口的 Go plugin 版本：扫描 plugins/ 目录 → plugin.Open() → 注册命令路由 | P0 | infra/plugin_so | P1 全部 | 编写测试 .so 插件，启动后自动扫描加载，命令匹配可执行 |
| P4-003 | exe 插件加载（补充） | 实现 exec 子进程版本：启动 exe → stdin JSON → 读 stdout 响应；以可添加方案实现 | P2 | infra/plugin_exe | P4-002 完成 | 测试 exe 插件可正常调用并返回结果 |

---

## 第五阶段｜触发控制

目标：mention/auto 双模式切换，精力/冷却/连续回复规则生效。

| 任务编号 | 任务名 | 内容 | 优先级 | 涉及模块 | 启动条件 | 验收标准 |
|---|---------|------|:---:|------|------|------|
| P5-001 | 触发模式 | 实现 mention（仅 @/私聊回复）和 auto（AI 自主判断）模式，每个群可独立配置 | P0 | service/control | P2-002、P2-004 | mention 模式只有被 @ 才回复；auto 模式可通过预检后自主回复 |
| P5-002 | 状态规则 | 实现精力值（消耗/恢复）、冷却时间、连续回复上限、时段控制、短消息忽略；纯规则层，不调 LLM | P0 | service/control | P5-001 完成 | 精力低于阈值不主动说话；连续 N 句后强制冷却；深夜静默 |

---

## 第六阶段｜联调验收

目标：完整消息链路跑通，bot 能在群里正常对话。

| 任务编号 | 任务名 | 内容 | 优先级 | 涉及模块 | 启动条件 | 验收标准 |
|---|---------|------|:---:|------|------|------|
| P6-001 | Prompt 组装联调 | 确认人格 → 群画像 → 压缩摘要 → 窗口 → 当前消息的 prompt 顺序正确 | P0 | service/agent + service/memory + service/persona | P3-003、P4-001 | 生成 prompt 格式符合架构文档第 6 节 |
| P6-002 | 完整消息链路 | 端到端：收到群消息 → 中间件 → 触发判断 → 拼 prompt → Agent 推理 → 回复 → 记忆更新 → 窗口追加 | P0 | 全部 | 前五阶段全部完成 | bot 在群聊中被 @ 能正常回复；记忆正常更新；摘要正常生成 |
| P6-003 | 稳定性验证 | 连续运行数小时，检查内存泄漏、goroutine 泄漏、SQLite 文件增长、API 调用频率 | P1 | 全部 | P6-002 完成 | 内存不持续增长；goroutine 不泄漏；API 调用不超过限制 |

---

## 禁止提前实现清单

以下功能在对应阶段到达前明确不得实现：

- **第一阶段禁**: ZeroBot 连接、eino 调用、SQLite 读写、业务逻辑、中间件
- **第二阶段禁**: 窗口管理、画像缓存、压缩摘要、记忆 Tool、人格系统、插件系统、触发控制
- **第三阶段禁**: 人格系统、插件系统、触发模式
- **第四阶段禁**: 触发模式、完整联调
- **第五阶段禁**: 端到端压测

---

> 每个任务完成后应执行 `go build ./... && go vet ./...` 验证。
