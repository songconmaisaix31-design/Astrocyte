# Attention Engine W0 · M0 契约锁定 · 迁移 · 存储基础 · HTTP 关口

## Windows CI 限定返修（2026-10-11，新 Dispatch）

Task `task_18154894914d` / Dispatch `ctx_84bdbc33a991`；原 W0 Codex --yolo 接续，无新 Agent。分支 `songconmaisaix31-design/attention-engine-w0-core`，先普通 fast-forward merge root 精确最终树 `0e49301734549a02b9b1b9078aaef96fd6422742`，SOURCE `425d965f82bb3d8b573604476134d466838be416` 已 commit+push。此次只改 `internal/adapters/agents/**` 与本报告；root 合入带来的其他 owner 文件没有再修改。

首 RED 保留：真实 Windows CI `38150022874` 日志 `C:/Users/DW/AppData/Local/Temp/astrocyte-ci-38150022874-first-failed.log` 未改。runner 的 TEMP 使用 `RUNNER~1`，`EvalSymlinks` 输出长名 `runneradmin`，原字符串比较误拒绝同一授权目录：native header 源被标成 redirected、批准文件 scope_denied、registered root/parent 断言不匹配；scanner/Works 的相关失败另归 W7。本机新真实 8.3 fixture 首场也 RED（`real 8.3 alias is treated as a different approved directory`、header unavailable），后续 junction 首场 RED（末级 junction 被 Go 保留为 irregular、metadata 源进入扫描）；以下 GREEN 是返修复验，不覆盖首错。

- agents 包内增加 Windows `GetLongPathName` 拼写展开与非 Windows 原行为 helper；rootKey、inside、显式 history 项目匹配共用同目录比较。缺失 checkout 只展开现存祖先，缺失 suffix 仍为注册关联元数据，不猜测替代根或授予读取。此 API 的长名称/失败语义参见 [Microsoft 文档](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getlongpathnamew)。
- 未将 `EvalSymlinks` 解析结果当成原路径名称：metadata 源和项目 cwd 仍拒绝 root/祖先重定向，Windows 另检查 reparse 属性，覆盖 Go 末级 junction 未解析的情形。批准相对路径逐组件检查仍在；短名称 `.codex` / `credentials.json` 先展开再执行原 secret 限制，真实 `CODEX~1` / `CREDEN~1.JSO` 拒绝通过。绝对路径、越界、其他未批准子目录、输出预算、默认拒绝与撤销行为保持。
- 原三个测试的 root/parent 断言改为同目录 key 比较，仍检验数量、来源、缺失 main 失败、状态和未授权目录。新增 Windows 测试通过实际 `GetShortPathName` 获取 fixture 的 8.3 名称，不修改 TEMP、不造假别名；无可用别名时明确 `NOT_RUN`。仅合成既有形状的临时 metadata header 和目录，不读取私人真实记录或扩大 metadata 权限。

所有 Go 命令串行使用 `GOFLAGS=-p=1`、`GOMAXPROCS=2`：

| 命令 / 范围 | 结果 |
|---|---|
| `go test ./internal/adapters/agents -run '^(TestWindowsActualShortPathScopes\|TestWindowsRedirectedPathsRemainDenied\|TestNativeMetadataIdentifiesStandaloneProjectWithoutReadGrant\|TestNativeMetadataSessionSampleDoesNotHideLaterProjectRoots\|TestProjectMetadataHeaderOnlyScopeDuplicateAndUnknown\|TestProjectScopeDefaultDenyAndDirectoryGranularity\|TestRegisteredDiscoveryScopesDedupAndMetadata\|TestRegisteredMissingMainStillAllowsExplicitLocalWorktree\|TestProjectSymlinkEscapeAndOutputBound\|TestHistoryDiscoveryRequiresChosenRootAndFiltersBeforeBody)$' -count=1 -v` | PASS：六个 CI 指定测试、真实 8.3 的 scope/源/cwd/重复/缺失 parent、secret alias、真实 junction root 与祖先拒绝、history 选择及 traversal/output bound 均执行；日志 `C:/Users/DW/AppData/Local/Temp/astrocyte-W0-ci-38150022874-targeted.log` |
| `go test ./internal/adapters/agents -count=1` | PASS，3.761s；live opt-in 测试按原规则不执行，不是私人 native 验收 |
| `go vet ./internal/adapters/agents` | PASS |
| `GOOS=linux go test -c -o C:/Users/DW/AppData/Local/Temp/astrocyte-W0-agents-linux.test ./internal/adapters/agents` | PASS，仅 Linux 编译，未运行 Linux 测试 |
| 对修改的 agents Go 文件执行 `gofmt -w`，`git diff --check` | PASS |
| `git merge-base --is-ancestor 0e49301734549a02b9b1b9078aaef96fd6422742 HEAD` | EXIT0，保持精确 root 普通历史 |

真实限制/未执行：本机创建真正 symlink 缺 Windows privilege，该分支明确 **NOT_RUN**；实际无需该权限的 junction 已创建并验证拒绝。原 symlink 用例保留，需有权限的 CI/Linux 环境执行；真实 8.3 本机已执行通过，不是 NOT_RUN。远端 CI 复跑、scanner/Works 返修集成、全量 check/build/UI/media 不在本次执行，交由原 W7 与原集成 Agent/root；本地 agents GREEN 不宣称 CI 或整体交付 GREEN。原未答产品/私人 native 权限仍未决定。

## 当前 Codex W0 shared/M6 交付（2026-10-11）

Task `task_168bacd9d56d` / Dispatch `ctx_b16e623a7b11`；当前 Orca Codex（codex --yolo），未启动其他 Agent，精确模型标识未由本会话接口提供。分支 `songconmaisaix31-design/attention-engine-w0-core`；共享基础 SOURCE `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`，正式 HTTP/MCP/Works startup `396aec1d6bee777dd380ff06fc5d821d0f7f5b3f`，单日 Sleep/Windows 检查 `363fcca1d515b6c373a8aba285e5ded4605727d5`，强化 M6 SOURCE `f4bf0257b3f2e93e2ce67197bfbb0764a68332a3`。各阶段均 commit+push，原 owner 代码通过授权的精确普通 merge 消费；下文旧波 A 状态保留为历史，不代表当前缺少服务组装。

- 完成共享持久化契约/迁移015–018、稳定原始 Event/幂等 origin/精确 BLOB、Signal 防重、独立节点更新时间与生命周期、项目记忆及人工审计、版本只追加、Sleep 日周期/指标/候选/冲突；没有独立 Evidence 表或证明系统。
- 正式 HTTP 组装三个原 owner 人工关口、分享/降级/引用/读取，以及七个 Works 端点和 registered-root 缓存边界；真实原 W1 Works sink 只写 pending/noSignals。独立 stdio MCP 绑定可信启动身份/项目/会话，原始已提交 ID 和 capture_pending/dynamics_pending 分开保留。
- `pnpm check` 最终 EXIT0：gofmt、Go vet/全 Go 测试、modverify、依赖、27 包架构、进程清理2项、275 OpenAPI examples/生成漂移、真实 S1 API14项、TypeScript/ESLint、Vitest13文件90项、diff检查均通过。Redocly 保留 EventV1 未使用组件 warning1。使用 GOFLAGS=-p=1、GOMAXPROCS=2、Node heap512；前面 paging/ENAMETOOLONG/版本探测超时及首次零推荐失败不被此次 PASS 覆盖。
- `pnpm build` 最终 PASS，产物 `dist/astrocyte.exe`、`dist/astrocyte-mcp.exe`、`dist/astrocyte-sleep.exe` 和 web。实际构建 MCP 子进程/SQLite/重启、实际 HTTP 幂等审计/重启、实际单日 Sleep 命令/日幂等/两种显式事实投影均通过；M6 通过共享正式 factory 注入合成时钟，未新增历史 CLI 回放参数。
- M6 强化验收与 root 独立复算 PASS：七天焦点 research→implementation，30 原始观察/7 日周期/1 冲突/1 人工批准经验；implementation A=.9552011491162884、score1.0480013210594767，research A=.33674673723047327、score.3350055591708292，均读取真实图并独立计算非零分数。精确 BLOB、新进程冷重启、人工确认/分享/引用、只追加版本、失败仅降 confidence、query_only 推荐均验证。最新报告 `C:/Users/DW/AppData/Local/Temp/astrocyte-M6-W0-20261011/M6-week-result.json`；first-zero 原件单独保留。

真实剩余限制：M1 私人 metadata 范围未答、真实 native replay 未执行；官方 MCP client/native Agent 互操作、七日真实用户比较、远端 CI、最终全轨集成/main merge 未在本任务执行。引擎参数、检索年龄/衰减、版本确认策略的产品默认仍未决定，测试显式配置不替用户选值；sqlite-vec 依既定 spike CUT。最终跨轨集成、root 复验及最终 Git 由协调者后续执行；本任务不宣称整体产品或真实七日验证完成。

## 历史波 A 与阶段记录

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

W4 requested independent retrieval age basis: nullable StoredKnowledge.ConsolidatedAt and migration018 consolidated_at, without fabricated legacy backfill. Human governance updated_at must not substitute for consolidation time; W4 owns content writes and explicit retrieval policy. Named shared GetKnowledge(ctx,p,project,id) and HTTP call aligned to actual W4 signature. Targeted M0 checks PASS; regenerated client includes nullable consolidated_at.

Consumed W7 receipt source b166848 through ordinary merge. Works-only transport now requires request_id equal Idempotency-Key, expected_version=1 (independent manual expected_revision); real HTTP accepted-result replay, mismatched-input conflict, no rescan on same key, and replay across SQLite cold restart PASS. Public Server.Handler exposes actual session/CSRF transport for owner integration tests. Root canonicalized GetKnowledge(ctx,p,id,project) after simultaneous same-string changes; named shared signature and HTTP follow root, W4 owner latest implementation must be consumed before read assembly. This replaces prior project,id note without changing W4 files.

W4 first query_only retrieval RED retained: EngineRepository.WithTx used BEGIN IMMEDIATE before SELECT. Added shared WithReadTx(ctx,fn(EngineTx)) with deferred BEGIN/consistent snapshot and mutation Exec rejection; W4 owns routing its reads. Actual PRAGMA query_only=ON snapshot test PASS. No full M4/retrieval claim follows from this shared fix.

Standalone optional engine services now activate the existing auth/session/CSRF guard without dummy ancillary services. Actual Knowledge/Experience/Works service session regression checks PASS. HTTP Experience factory uses cached registered/local project IDs or roots, rejects unknown/Agent context and unknown share target, then invokes the original project-bound W5 service; actual owner-mismatch/human share checks PASS. First assembly compile RED (`NewKnowledgeRepository` expects EngineRepository, not DB) preserved; corrected cmd injection, targeted cmd/httpapi green is a separate retest.

Production startup assembly stage (2026-10-11): consumed exact owner sources W1 b73bf657, W2 2c148c10, W4 73ba73b2, W5 b43d2a4c, W6 2314b603 and W7 a3eab2c6 through authorized ordinary merges; owner business files remain untouched. cmd/server now wires the actual W1 NewPendingEventService -> WorksEventSink, actual W4 Knowledge governance, project-bound W5 factory and W2 ConflictGate. Works metadata remains pending observations with exact raw BLOB and no Signals; no undecided signal-strength default is supplied.

Added separate cmd/mcp production stdio executable, emitted by pnpm build alongside HTTP server. Explicit startup project/session/principal and optional human veto bind MCP authority. Signal strength requires explicit configuration; optional Dynamics policies remain nil until supplied. The glue invokes original EventService, W5 committed-failure callback and W2 ApplySignals for persisted signals. It never auto-seeds agent topics: original human seed/Sleep birth rules remain authoritative. Known receipt IDs survive downstream errors through original W1/W6 committed-delivery transport.

First built-command test RED: unseeded fixture-topic receipt committed, then ApplySignals returned not_found attention_topic. This was a test assumption about topic birth, not grounds for granting agent human seed authority. Separate corrected fixture explicitly seeds through original W2 human gate; actual built executable stdio reports mapped/pending events, rejects Agent veto/cross-scope, preserves forged-source raw bytes while trusted observation binding/outcome unknown, explains exact Event ID and strength, and reopens the same SQLite state. Additional unseeded report verifies isError=true with accepted=true, committed Event ID and capture_pending=true, original BLOB retained and no extra topic born. Targeted cmd/mcp+cmd/server checks PASS; pnpm build PASS under serial Go/Node heap512. These are contract_local inputs and process integration, not official MCP client/task_live acceptance.

Remaining: latest W3 Sleep publication and shared mixed-week M6 assembly/restart checks, repeated human HTTP governance idempotency, W7 browser validation of this actual sink, final broad check. User default numerical policy/native metadata consent remain undecided; no model/private text capture performed. Original owner duplicate-DDL/paging failures remain recorded separately from their targeted fixes.

Production startup source `396aec1d6bee777dd380ff06fc5d821d0f7f5b3f` published with actual dist/astrocyte.exe and dist/astrocyte-mcp.exe build entries. Subsequently consumed W6 typed seam `1f185a494ef05203977b33b3ff5816bbb033d75c`; recorder now returns original rows plus MCPDynamicsPendingError for signal-list/fold errors. Actual built executable verifies dynamics_pending=true and capture_pending=false for prebirth topic, independently of W5 capture pending. No new topic is auto-created.

Actual HTTP cmd/server Experience factory integration verifies W1 committed failure callback -> W5 candidate, human session/CSRF approval and share, Agent rejection, same header Idempotency-Key with changed RequestID/trace returning original results before CAS, changed expected-version conflict, registered foreign owner denial, and cold SQLite restart preserving version3, one share and exactly two human actions. First test expectations RED (foreign owner expected409 but correctly403; human action count expected3 but correctly2) retained; corrected targeted test PASS. Full cmd/mcp+cmd/server tests PASS. W4 final report and W7 actual UI source consumed by ordinary merges; not claimed as W0 live-browser acceptance.

Single-date Sleep production entry: cmd/sleep assembles original W3/W2 via shared cmd/attentionruntime factory, emitted as dist/astrocyte-sleep.exe by pnpm build. Explicit --day YYYY-MM-DD, --timezone IANA, --decay, --initial-confidence and --fact-projection none|facts-map-v1 required; --data-dir and feedback-policy flags optional. No scheduler/model/policy defaults. facts-map-v1 directly selects original W3 LiteralSleepFactsMap (SOURCE61a196ba, REPORT39cb8fec); none leaves retention_measurable=false. Actual built command verifies daily idempotent run/restart plus explicit literal projection recall, targeted cmd/sleep PASS; pnpm build PASS (server, MCP, Sleep and web). Historical mixed-week CLI clock boundary is a coordinator question; no wall-time backdating presumed.

Broad check first new RED: Node spawnSync gofmt.exe ENAMETOOLONG from all 320 absolute filenames in one Windows argv. Shared process.mjs now checks every file in bounded 40-file relative-path batches; gofmt PASS separately. Subsequent assembled broad run Go vet/all Go tests/modverify PASS, then existing summarize0.25.1 --version 60-second probe ETIMEDOUT/SIGTERM with no output, so remaining checks NOT_RUN in that invocation. Targeted same dependency script serial measured2.718s/direct PASS; timeout/version unchanged. Full check must be rerun after M6 owner repair.

M6 first insufficient green is withdrawn: original actual services did migrate focus, exercise human gates and preserve restart state, but the initial test only asserted recommendation order. Root independently found actual recommendation input Attention=0/score=0 despite positive original nodes; stronger M6 oracle reproduced RED implementation storedA=0.955201149116/inputA=0/score=0. W4 KnowledgeAttention reads topic correctly but unmarshals StoredAttentionNode JSON as a plain AttentionNode; original W4 owner repair required, no W0 domain edit. First artifact retained at C:/Users/DW/AppData/Local/Temp/astrocyte-M6-W0-20261011/M6-first-zero-ranking.json and root copy M6-week-first-zero-attention.json. Future green must match original topic A, positive score and independent relevance * exp(-lambda * age_seconds) * A * audited-human-boost calculation. Separate first cold-comparison RED was report JSON RawMessage whitespace compaction; public JSON equality plus exact raw BLOB comparisons verifies transport and byte preservation independently.

M6 repaired targeted result: consumed original W4 SOURCEabff94dbb8b614e4489e467e9095935fd0815ba9, then strengthened mixed-week test PASS. Exact artifact C:/Users/DW/AppData/Local/Temp/astrocyte-M6-W0-20261011/M6-week-result.json; first-zero artifacts remain separate. Seven synthetic days through the same formal Sleep assembly factory and original W1/W2/W3/W4/W5/W7 owners yield research focus days1-4, implementation days5-7; final implementation A=0.955201149116. Recommendations require original per-topic A, positive score and independent formula (explicit lambda1e-8/created-age/human-reference boost0.1), not alphabetical order. Actual HTTP three human gates reject Agents; scoped share/reference, two crystallized versions with explicit fixture retain-confirmation policy, quarantine exclusion then human-approved experience, confidence-only failure explanation, seven atomic daily replays, nil interpreter, raw BLOB preservation and cold child-process state all checked.

Coordinator answered the historical CLI boundary: formal factory seven-day synthetic Clock plus actual built CLI single-day/restart acceptance satisfies M6; no replay-as-of product flag added and old-signal chronological rejection retained. Fact rule is original owner LiteralSleepFactsMap opt-in, never generic private-text inference. Final full check/build still pending at this report stage. Actual real M1 replay remains permission-blocked; official MCP client/native interop and user seven-day comparison not executed, user numerical/retrieval/version policy defaults remain undecided.
