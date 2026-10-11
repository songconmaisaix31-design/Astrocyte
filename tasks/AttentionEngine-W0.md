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

## 枚举单一来源（协调预审，避免同名重复定义）

本轨 `engine_contract.go` 是 §4 类型与值域词汇的单一 owner。与已提交波 A 域文件的冲突及裁决（集成轨按此收敛，各轨 drop 重复、引用本契约）：

- **W4 `domain/knowledge.go`**：本轨已采纳其 `KnowledgeStatus` 类型 + `KnowledgeInbox`/`KnowledgeConfirmed`（`KnowledgeEntry.Status KnowledgeStatus`）。W4 波 A 文件须删除同名 `KnowledgeStatus`/`KnowledgeInbox`/`KnowledgeConfirmed` 重复定义，仅保留 `ConfirmKnowledge`/`ReferencableKnowledge`/`AppendCrystallized`/`KnowledgeVisible` 函数与 `CrystallizedVersion`（版本接口，本轨未重复定义）。
- **W5 `domain/experience.go`**：其 `VerdictAvoid/VerdictPrefer/VerdictCondition`、`StatusQuarantine/StatusApproved`、`KindAgentTask/OutcomeFailed/OutcomeSucceeded` 与本轨 `ExperienceVerdict*`/`ExperienceStatus*`/`EventKindAgentTask` 语义重复。集成时以本轨 `domain.Experience`（verdict/status）与 `domain.Event`（kind）词汇为单一来源，W5 删除重复 value 常量（或反向改名，二选一，勿并存）。
- **W2 `domain/{conflict,dynamics,lifecycle,topic,antimatthew}.go`**：仅函数与阈值常量（`MergeRelatednessThreshold=0.8`、`MomentumThreshold=0.15`、`ColdArchiveInactivityDays=90`），无类型与本轨 `Conflict`/`AttentionNode`/`Signal` 冲突；`Conflict` 结构体以本轨 `domain.Conflict` 为源。
- **W3 `domain/sleep.go`**：`EventTopic`/`Cluster`/`NodeDecision` 等为 sleep 流水线内部类型，与本轨 §4 类型无冲突。

## 预审接口要求（本轨未完成，列为下一步；Codex 接续时落实）

1. **StoredEventID**：`Event`（§4 原始形状）无 id；稳定 `event_id` 由存储层（`attention_engine_events.id`）分配，`EvidenceRef`/`EventRef` 经 `GetEvent` 回溯原始 `raw`。本轨未建独立 Evidence 表（正确），但应用/存储包装的 `GetEvent(id)→Event` 接口与稳定 id 生成（内容寻址或单调 id）待定义。
2. **TS 单位**：`ts`/`last_active` 均为毫秒；W2 动力学 `Δt` 天换算须用 `86400000`，勿按秒。已通知 W2。
3. **updated_at / 生命周期**：`AttentionNode.last_active` 保持 §4"最后行为"语义不变；独立"状态更新时间戳/生命周期状态"（避免重复衰减与冷归档丢失）在应用/存储包装另加字段，勿改 `last_active` 原义。
4. **project_id / 分享审计**：§4 的 `KnowledgeEntry`/`Experience` 未含 project；默认项目隔离 + 显式跨项目分享（带审计）在应用/存储包装保留 `project_id` 与分享审计字段。`migrations/014` 表未加 `project_id` 列，接续时在对应 store 的 `data` JSON 或加列承载；MCP scope 来自受信启动配置，不信任请求 project 参数。
5. **topic 待归类**：原始事件未带显式主题时 `topic` 保持空/待归类，由 sleep 规则生成，不凭空编标签。

## 剩余限制 / 未完成清单

- **已完成**：§4 契约类型 + 引擎接口 + 三个人工关口接口、迁移 014、存储基础 `engine_common.go`、HTTP 关口路由（501 未组装）、watcher 映射 fixture 测试、spike 结论。SHA `1b3b2ef5a4fd2b96a947da43de81ffdd9b878f24`。
- **未完成（Handoff）**：`GetEvent(id)` 稳定 id 包装接口；`project_id`+分享审计落库；`updated_at`/生命周期字段；§4 与 W4/W5 波 A 枚举最终改名合并（见上节）；三个关口 store（W4/W5/W2）。
- **FAIL/UNKNOWN**：无 FAIL；sqlite-vec 为已裁决"砍"（非 UNKNOWN，证据见上）。关口端点 501 为 M0 诚实未组装状态，非失败。
- 未合 main/未跑远程 CI；本轨分支 `songconmaisaix31-design/attention-engine-w0-core` 已 push `1b3b2ef`。

## Codex PRIORITY1 persistent foundation (2026-10-11)

Client: current Orca Codex session; no other Agent launched. Baseline `b6a5794a4bf6ed282b60cace4bf81d779383df87`.

- `app/persistence_contract.go` preserves task-book domain shapes through `StoredEvent`/`Observation`, `StoredAttentionNode`, `StoredKnowledge`, `StoredExperience`, share/action audit and daily sleep wrappers. `EventReader.GetEvent` resolves original receipt identity; pending Event is nil; unknown tokens remain nil; exact source bytes live in `raw_bytes` BLOB, not JSON serialization.
- Migration015 adds origin-key uniqueness, replay signal uniqueness `(event_id,topic,class)`, rejection of pending observation signals, separate node `updated_at`/`last_active`, lifecycle/version, project-owned knowledge/experience, append-only timestamped knowledge versions, human share/action rows, origin-event uniqueness, sleep daily run/metrics/candidate/conflict rows.
- Upgrade from migration014 leaves missing legacy origin/raw/project fields unknown; app owners must reject empty owner scope rather than expose it globally. Existing node updated_at initializes from last_active only during migration.
- Shared governance signatures require Principal, project filtering, CommandMeta idempotency/version. W4/W5 own implementations and must validate human authority, persisted ownership/shares and source references; W2 owns conflict/engine. W4 deletes duplicate KnowledgeStatus enums and references W0 definitions.
- Exported FocusExplanation includes original StoredEvent, strength/weight, elapsed days/decay factor, separate attention/confidence contribution, confidence-only flag and project/agent/topic relationships.
- Removed old fixture claims: correlation_id is not native session; aggregate/material/external ID is not project scope; source human/read are not core enum values; stop_confirmed is not task success; publication time is not selection time. Remaining watcher fixture tests only receipt unknown preservation; production source mapping and live replay remain W1 responsibility.
- sqlite-vec CUT follows original spike and pure-Go stack; no embedding/vector implementation added.

Focused verification: `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine ./internal/adapters/sqlite ./internal/adapters/httpapi` PASS (actual migration upgrade/restart/constraints; business app integration not yet assembled). Broad check/build pending in this stage. HTTP three gates remain 501 unsupported until actual owner implementations are assembled; W7 endpoints/CLI/M6 are subsequent work, not completed by M0.

PRIORITY1 source committed+pushed `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`; root consumed integration `2268f4920973128068f46a182dfddfb040a1dbc2`. `pnpm check` PASS (all Go/vet, 14 S1 acceptance, 251 contract examples/drift, TS/lint, 83 frontend tests); `pnpm build` PASS under GOFLAGS=-p=1/GOMAXPROCS=2/Node heap512.

Preserved first failure: original W0 Dispatch `ctx_64948eee5483`, terminal `term_398bda4c`, default-parallel `go build ./...` on Windows observed `paging file too small`; root baseline document retains the observation. Full original stderr was not archived, so exact temporary paths/error lines remain unavailable. Later serial PASS is separate evidence and does not erase the first failure.

PRIORITY2 contract stage: governance get/share/demote/reference routes and strict project/version DTOs; seven Works endpoint schemas/generated types. `pnpm check:contracts` PASS, 275 examples, generated drift PASS. `go test ./internal/adapters/httpapi` PASS. Actual services remain optional/unsupported until owner assembly. Root authorized exact ordinary W7 dependency merge `ca0c255b94ee4374cad9fa610b895d6f87d7dabf`; this is compile-time source consumption, not W7 acceptance.

PRIORITY2 HTTP/Works assembly stage:
- Consumed W7 original published sources `ca0c255b94ee4374cad9fa610b895d6f87d7dabf` then `0e38f94fa516f5b9f8ec4f48f6f76fbad425e5e1` through authorized ordinary merges. No owner domain/app/store edits.
- Added seven Works HTTP handlers and actual `cmd/server` NewWorksService assembly; GET is cache-only, refresh selects only exact cached registered metadata roots, omitted roots resolve from that cache. Human-session/CSRF/strict JSON/Common CommandMeta validation reuse existing transport; manual metadata uses independent expected_revision.
- Migration016 creates works_album/works_metadata plus owner-requested works_requests exact BLOB receipt rows keyed (caller_id,command,request_key), status pending/accepted/unknown/rejected, nullable result/error BLOB. W7 owns Begin/Finish request semantics and sink error propagation; header validation alone is not durable idempotency evidence.
- Migration017 adds append-only project knowledge references with pinned knowledge/crystallized versions, nullable composite version FK and human audit. Upgrade/restart/append-only/FK checks PASS.
- Actual HTTP through W7 production service+scanner+SQLite/migrations: cache unknown before scan; missing session/CSRF and Agent writes rejected; parent root denied; default registered-root scan; attitude/group/handoff/hidden mutations; JSON delivery remains unverified; cold reopen retains all metadata. This test used temporary authorized fixture roots; no real-user filesystem or chain1 claim.
- `pnpm check:contracts` PASS (275 examples and generated drift); `go test ./internal/adapters/httpapi ./cmd/server ./tests/attention-engine -count=1` PASS; `pnpm build` PASS. W4/W5/W2 actual gate persistence, W7 durable receipts/sink wiring, separate stdio startup and M6 remain pending owner APIs. Current Works event sink is nil and does not claim Event ingestion.

W3 requested metric clarification: shared EngineSleepMetrics now includes retention_measurable/storage_measurable and representation. Denominator-zero is explicitly unmeasurable; representation describes measured before/after payloads, not SQLite file reduction. Fields persist inside EngineSleepRun data JSON; W3 owns calculation. Targeted M0 test PASS; exact source reported to W3/root.
