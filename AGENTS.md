# AGENTS.md
# PlumeBot — AI 开发执行规范

## 0. 快速参考

| 项 | 值 |
|----|-----|
| Go 版本 | 1.26.4 (go.mod: `go 1.26.4`) |
| 模块名 | `plumebot` |
| 入口 | `cmd/bot/main.go` |
| 当前阶段 | 第一阶段：项目骨架（stub 可编译，无外部服务） |

```bash
# 编译
go build -o bot.exe ./cmd/bot/

# 全量编译检查
go build ./...

# 静态分析
go vet ./...

# 运行（当前仅打印启动日志后阻塞）
./bot.exe
```

测试：暂无测试文件（`**/*_test.go` 不存在），标准库 `testing` 留待后续。

---

## 1. 文档用途

本文件是本仓库中所有 AI 开发任务的长期执行规范。

AI 助手在开始任何任务前，必须先阅读本文件，并严格按照本文件中的：

- 项目目标；
- 技术栈；
- 模块边界；
- 目录规范；
- 代码限制；
- 分层规则；
- 测试要求；
- 任务验收标准；

执行工作。

用户每次只会下达一个小任务，例如：

```text
完成 domain 层 storage 接口定义
```

AI 助手必须根据本文件找到对应任务，只完成该任务，不顺带开发后续任务，也不实现任何未明确要求的业务功能。

---

## 2. 项目基本信息

项目名称：

```text
PlumeBot
```

定位：

```text
基于 OneBot 协议的 QQ 机器人，对接 NapCat，AI 驱动的赛博群友。
```

架构设计详见：`docs/architecture.md`

当前阶段：

```text
第一阶段：项目骨架
```

本阶段目标：

```text
建立 Go DDD 分层工程骨架 —— domain 接口定义 + infra 空实现 + service 编排框架，
可编译、不接入外部服务、不实现业务逻辑。
```

禁止提前实现：

- ZeroBot 连接与消息收发；
- eino Agent 推理；
- SQLite 建表与数据操作；
- 上下文窗口管理；
- 群聊画像加载；
- 人格系统；
- 插件系统；
- 记忆更新闭环；
- 触发控制/状态规则；
- 中间件链；
- 任何具体业务逻辑。

---

## 3. 技术栈

必须使用：

- Go 1.21+（当前环境 go 1.26.4）；
- ZeroBot（OneBot v11 连接层）；
- eino / CloudWeGo（AI Agent 引擎）；
- `turso/sqlite`（SQLite 驱动，纯 Go 无 cgo）；
- `uber/zap`（结构化日志）；
- `github.com/spf13/viper`（配置文件解析）；
- Go `plugin` 包（插件动态加载）；
- Go 标准库 `testing`（测试）。

当前明确不使用：

- MySQL / PostgreSQL；
- Redis；
- 消息队列；
- gRPC / 微服务；
- 任何外部向量数据库（Milvus / Qdrant）；
- `stretchr/testify`（用标准库 testing）；
- cgo 依赖的库。

可选能力（默认关闭）：

- LLM API embedding（向量检索开关，需用户自行配置 API）。

第一阶段还明确不使用：

- ZeroBot；
- eino；
- `turso/sqlite`；
- Go `plugin` 包。

---

## 4. 工程目录

```text
plumebot/
├── cmd/
│   └── bot/
│       └── main.go                 # 唯一入口，组装依赖注入
├── internal/
│   ├── domain/                     # 领域层：纯接口 + 实体，零外部依赖
│   │   ├── entity/                 #   公共实体（Message, Event, Profile...）
│   │   ├── agent.go                #   Agent 接口
│   │   ├── memory.go               #   Memory 接口
│   │   ├── persona.go              #   Persona 接口
│   │   ├── plugin.go               #   Plugin 接口
│   │   ├── storage.go              #   Storage 接口
│   │   └── control.go              #   Control 接口
│   ├── service/                    # 业务编排层，依赖 domain 接口
│   │   ├── agent/                  #   prompt 组装 → Agent 推理
│   │   ├── memory/                 #   窗口 + 画像缓存 + 摘要流程
│   │   ├── persona/                #   extend 链 + 缓存
│   │   ├── plugin/                 #   发现、加载、路由
│   │   ├── control/                #   触发判断 + 状态规则
│   │   └── event/                  #   中间件链 + 分流编排
│   ├── handler/                    # 事件处理入口
│   │   ├── message.go              #   消息事件 → event service
│   │   └── notice.go               #   通知事件 → 规则处理
│   └── infra/                      # 基础设施，实现 domain 接口
│       ├── onebot/                 #   ZeroBot 封装
│       ├── ai/                     #   eino Agent 实现
│       ├── sqlite/                 #   SQLite 存储实现 (+ queries.go)
│       │   └── migrations/        #     版本化 DDL 迁移文件
│       ├── plugin_so/              #   plugin.Open() 实现
│       └── plugin_exe/             #   exec 子进程实现
├── pkg/                            # 可复用工具
│   └── config/                     #   配置加载
├── plugins/                        # .so 文件目录
├── data/                           # SQLite 自动生成（运行时创建）
├── config.yaml                     # 配置文件
├── docs/
│   └── architecture.md             # 架构设计文档
├── AGENTS.md                       # 本文件
├── README.md
├── go.mod
└── go.sum
```

---

## 5. 分层规则

### 5.1 domain — 领域层

职责：

- 定义所有业务接口；
- 定义公共实体（Entity）；
- 零外部依赖（不依赖任何 infra 实现、不依赖第三方库）。

禁止：

- 包含任何 `import` 第三方库；
- 包含接口实现；
- 包含数据库操作；
- 包含网络调用。

### 5.2 service — 业务编排层

职责：

- 实现业务流程编排；
- 依赖 domain 接口（不依赖 infra 实现）；
- 通过构造函数注入依赖。

禁止：

- 直接 import infra 包；
- 直接操作数据库；
- 直接发起网络请求。

### 5.3 infra — 基础设施层

职责：

- 实现 domain 层定义的接口；
- 封装第三方库（ZeroBot、eino、SQLite 驱动等）。

依赖方向：

```text
infra → domain（实现接口）
```

### 5.4 handler — 事件处理入口

职责：

- 接收外部事件；
- 调用 service 层；
- 不包含业务逻辑。

### 5.5 依赖方向总图

```text
cmd ──→ handler ──→ service ──→ domain（接口）
                      │
                      └──→ infra（编译时注入）

infra ──→ domain（实现接口）
domain 零依赖
```

上层依赖接口，底层实现接口。编译时通过 main.go 组装依赖关系。

---

## 6. 模块说明

### 6.1 cmd/bot/main.go

职责：

- 唯一 main 入口；
- 加载配置；
- 手动依赖注入（组装 service + infra）；
- 启动运行。

### 6.2 internal/domain/

职责：

- 全部业务接口定义；
- 公共实体定义。

禁止：

- 任何外部依赖 import。

### 6.3 internal/service/

职责：

- 各模块的业务流程编排；
- 持有 domain 接口引用。

禁止：

- import infra 包；
- 包含数据库、网络等具体实现。

### 6.4 internal/handler/

职责：

- 消息事件、通知事件的处理入口；
- 调用 event service 进行分发。

### 6.5 internal/infra/

职责：

- domain 接口的具体实现；
- 封装 ZeroBot、eino、SQLite 等第三方库。

每个 infra 包必须：

- 若包含 SQL 操作：`queries.go` — 所有 DML 语句以包级 `const` 存放，方法体内不内嵌 SQL 字面量。

哨兵错误定义在 `internal/domain/errors.go`，infra 层引用 `domain.ErrXxx` 返回，供上层 service 通过 `errors.Is` 判断，避免上层耦合数据库驱动。

第一阶段 infra 下各包只需空文件或返回错误的 stub，保证编译通过即可。

### 6.6 pkg/config/

职责：

- 配置文件加载与解析。

---

## 7. 代码规则

1. domain 层不 import 任何第三方库（标准库除外）。
2. service 层不 import infra 包。
3. 接口在 domain 定义，实现在 infra。
4. 依赖通过构造函数注入，不使用全局变量。
5. 第一阶段所有 infra 实现为 stub（返回 nil 或 error），保持可编译。
6. 不提前实现任何业务逻辑。
7. 不接入任何外部服务（ZeroBot、eino、SQLite 驱动等）。
8. 不为了"架构完整"创建大量空类和方法，只创建架构文档中明确列出的模块。
10. 哨兵错误（`domain.ErrNotFound`、`domain.ErrConflict`、`domain.ErrClosed`）定义在 `internal/domain/errors.go`，infra 层引用并返回，service 层通过 `errors.Is` 判断，禁止在 infra 包内自定义哨兵。
11. infra 包中 SQL DML 语句不得内嵌在方法体内，必须提取到 `queries.go` 作为包级 `const`；DDL 语句存放于 `migrations/*.sql` 通过 `//go:embed` 加载。
12. 代码保持直接、易读，不引入不必要的抽象层。

---

## 8. AI 助手每次任务的执行流程

接到任务后必须执行以下步骤。

### 第一步：阅读

阅读：

1. 本文件 `AGENTS.md`；
2. `docs/architecture.md`；
3. 已有代码和目录结构。

### 第二步：确认范围

在修改前输出：

- 本次任务目标；
- 计划修改/创建的文件；
- 明确不修改的内容；
- 遵守的分层规则。

不得把多个任务合并执行。

### 第三步：实现

要求：

- 优先最小改动；
- 不提前实现后续任务；
- 不接入外部服务；
- 不加入未要求的框架或库；
- 代码保持直接、易读。

### 第四步：验证

至少执行：

```bash
go build ./...
go vet ./...
```

如果环境无法执行，必须如实说明实际错误。

不得伪造测试通过。

### 第五步：交付

完成后必须输出：

- 修改文件列表；
- 关键实现说明；
- 实际执行命令；
- 编译/测试结果；
- 未完成事项。

完成后停止，不自动进入下一个任务。

---

## 9. 最重要的执行原则

1. 一次只完成一个任务。
2. 不提前开发下一任务。
3. 不扩大范围。
4. 不实现业务逻辑。
5. 不接入外部服务。
6. 不引入未声明的依赖。
7. 不在 domain 层 import 第三方库。
8. 不在 service 层 import infra 包。
9. 不把代码塞进一个文件。
10. 不为了"架构完整"过度设计。
11. 不伪造编译结果。
12. 代码优先简单、直接、易读。
13. 完成任务后停止，等待审核。
