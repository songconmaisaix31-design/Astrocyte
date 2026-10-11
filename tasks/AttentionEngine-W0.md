# Attention Engine W0 · M0 契约锁定 · 迁移 · 存储基础 · HTTP 关口

Task `task_0544bbb70579` / Dispatch `ctx_64948eee5483`；开发 Worker OpenCode / `deepseek/deepseek-v4-pro`。独占契约、迁移 014+、引擎接口、HTTP、入口与依赖锁；跨轨只 Handoff，不写各域 `*_store.go`。基线 `4d200df`（`attention-engine/v1-20261011` 的 plan 提交），集成分支上的 `94c6b5d`/`0e14c7d` 为 root 文档提交，非本轨写域。

## 交付内容

- **M0 契约类型**（`internal/attention/domain/engine_contract.go`）：§4 全部 schema 落为纯 Go 值类型——`Event`（source/session/ts/kind/tokens/scope/topic/raw）、`Signal` + §5.2 权重表（`SignalClass` 七类 + `SignalWeight`，失败信号 `-0.1` 仅降 confidence 语义在注释锁定）、`AttentionNode`（含 `related_topics` MAGMA 式邻接）、`Experience`（avoid|prefer|condition，quarantine|approved）、`KnowledgeEntry`（fluid/crystallized 两级）、`EvidenceRef`（event_id/weight_at_use/ts，不建独立对象）、`Conflict`（§5.5）。类型与字段值域（kind/source/verdict/status/decision）均以常量关闭。
- **引擎接口**（`internal/attention/app/engine_contract.go`）：`Engine` 边界（`ApplySignals`/`AttentionState`/`AttentionNode`，§5 状态演化与读侧分离）、`FocusExplanation` 证据链（§7 explain_focus）、三个人工关口服务接口 `KnowledgeGateService`/`ExperienceGateService`/`ConflictGateService` + 命令/结果 DTO + `EngineGates` 聚合。
- **迁移 014**（`migrations/014_attention_engine.sql`）：`attention_engine_events`（topic/source/kind 索引）、`attention_engine_signals`（event_id 外键）、`attention_engine_nodes`（topic UNIQUE）、`attention_engine_experiences`、`attention_engine_knowledge`、`attention_engine_conflicts`；status 均 `CHECK` 关闭值域；`foundation_schema` attention 版本 6→7。纯 schema，无 store。
- **存储基础**（`internal/adapters/sqlite/engine_common.go`）：`EngineRepository`（BEGIN IMMEDIATE 串行化 + `Conn`）+ `EngineTx`（context 绑定 Exec/Query/QueryRow）+ 通用 `engineLoadJSON`/`engineListJSON`。连接/迁移执行沿用既有 `adapter.go`。
- **HTTP 关口端点**：`contracts/openapi.yaml` 新增三路径 `/knowledge/{id}/confirm`、`/experiences/{id}/approve`、`/conflicts/{id}/adjudications`（+ `attention-engine` tag + 8 个 schema）；`internal/adapters/httpapi/engine_gates.go` 注册路由，`Services.Gates` 缺省为 nil 时返回 501 `unsupported_capability`（M0 未组装，非伪造转换）。`web/src/api/schema.d.ts` 已重新生成。

## 验证

| 验证 | 结果 |
|---|---|
| `GOFLAGS=-p=1 go test ./...` | PASS/exit0（全绿；含 `tests/attention-engine` 新增 10 例） |
| `GOFLAGS=-p=1 pnpm check` | PASS/exit0：gofmt/vet/mod/架构/依赖/进程退出、Go 全测试、14 项 S1 acceptance、251 契约例、生成一致、TS/lint、83 前端单测、diff |
| `tests/attention-engine` 契约测试 | PASS：§5.2 权重表精确值、Event/AttentionNode/Experience/KnowledgeEntry/Conflict JSON 形状与 EvidenceRef 三元组 |
| `tests/attention-engine` 迁移测试 | PASS：014 全新临时库应用、6 表存在、attention 版本=7、status CHECK 拒绝非法值、signal→event 外键、幂等重跑 |
| `tests/attention-engine` watcher 映射测试 | PASS：attention_outbox / material Usage / attention_source_items / local_agent_session 四类 fixture 无损映射到新 Event（raw 逐字保留原文 + 规范化字段正确投影） |

## sqlite-vec spike 结论（接/砍）

**砍。** 证据：

1. 当前唯一驱动 `modernc.org/sqlite v1.60.1` 是纯 Go 转译（c2go/ccgo），`CGO_ENABLED=0`，`go.mod` 无任何 cgo 驱动（无 `mattn/go-sqlite3`）。
2. 源码核查（`modernc.org/sqlite@v1.60.1/lib/sqlite.go:34703`）确认 `_sqlite3OsDlOpen` 走 VFS `xDlOpen`（Windows 映射 `libc.XLoadLibraryW`），但 SQLite 核心是 Go 内存对象，**不导出 C ABI**；sqlite-vec 是原生 C 扩展（`.dll`/`.so`，官方 Go 绑定 `sqlite-vec-go-bindings` 要求 CGO），其 `sqlite3_vec_init` 对真实 C `sqlite3*` 操作，无法注册进纯 Go 核心。
3. 实测 `SELECT load_extension('vec0')` 返回 `not authorized`（`ENABLE_LOAD_EXTENSION=0`，默认关闭）；即使打开，上面第 2 点的 C ABI 断裂仍在。
4. 接上即须引入 CGO + C 编译器 + mattn 驱动，违背"纯 Go 无 CGO"既有约束，且远超红线"1 天接完"。

**降级方案（记入偏差）**：§5.6 检索排序的 `relevance` 因子退化为确定性回退——topic 邻接 + 精确/关键词命中打分，仍保留 `A(topic)` 与 `e^{-λ_r Δt}` 衰减；检索侧重排的"不与状态更新混写"不变。红线 §9"不自研向量检索"不违反（未自研，是砍）。

## 给 owner 的 Handoff

- **W1**：watcher→Event 生产适配器写 `internal/adapters/events/**`，沿用本轨 `tests/attention-engine/watcher_mapping_test.go` 的映射约定（规范化字段投影 + `raw` 逐字保留）；`attention_engine_events` 表已就绪。
- **W4**：`knowledge_store` 写 `internal/adapters/sqlite/knowledge_store*.go`，`ConfirmKnowledge`（inbox→confirmed）在 `EngineRepository.WithTx` 上实现后组装 `EngineGates.KnowledgeGateService`。
- **W5**：`experience_store` 同理实现 `ApproveExperience`（quarantine→approved）。
- **W2**：`engine_store`/`conflict` 实现 `AdjudicateConflict`（pending→resolved_*）。
- 三个关口端点当前 501；集成轨（波 B/C）组装 `httpapi.Services.Gates` 后即启用，无需再改 openapi 契约（契约已锁定）。

## 剩余限制

- 关口端点仅在 `Services.Gates` 组装后返回 200；M0 内返回 501 是诚实状态（未写 W4/W5/W2 store）。
- 未合 main/未跑远程 CI；本轨分支 `songconmaisaix31-design/attention-engine-w0-core` 待 push。
