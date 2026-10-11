# Attention Engine W3 · Sleep Consolidation 流水线（M3）

## 最新结果（2026-10-11 22:48 Asia/Shanghai）：W3 实现及适用验收 PASS

原工作树和分支保持；实际客户端 Orca Codex / gpt-6.1-sol / YOLO。
范围为 W3 拥有的完整 app/store 流水线。验收输入为 **contract_local 合成事件**，
经实际 W1 Record、W2 Dynamics、W4 Knowledge、正式迁移与文件 SQLite 执行；
冷启动使用新的 Go 测试进程，未触碰个人库。不是七天真实使用或全产品验收。

### 实现与接口 Handoff

- 固定日周期入口：`app.NewSleepService(sqlite.NewSleepRepository(db), dynamicsEngine, SleepOptions).Run(ctx, day)`。
  `day.Location()` 明确当日时区；调用方保持同一个个人时区，日期键为 `YYYY-MM-DD`。
  读取午夜至 `min(次日午夜-1ms, 本次Clock)` 的已存事件/信号，不建设调度器。
- 全链同一个 `BEGIN IMMEDIATE`：检查已完成日周期 → mapped 当日 Event → topic 邻接聚类 → 去噪
  → W2 birth/已存 Signal 应用/关联更新 → W4 fluid 更新 → W2 生命周期/重浮现 → pending 冲突队列
  → 四张 Sleep 表与 metrics 一起提交。中途失败完整回滚；再次运行或冷进程读取原 receipt。
- `SleepTx` 复用 `NewEventTx`、`NewDynamicsTx`、`BindKnowledgeTx`、`NewSleepRunTx`。
  W3 不直接写 Event/Signal/图/Knowledge/Conflict 表，不开嵌套事务。
- 既有节点的信号由 W2 `h.Applied` 去重；新节点用最早本主题证据时间创建，再应用真实当日 Signal。
  最晚真实事件决定 LastActive，不用 Sleep 的执行时间冒充用户活动。
- 同簇相关 topic 保持独立；新主题使用同 scope 的邻接簇支持，节点只挂本 topic 证据。
  真合并调用 W2 `ObserveAssociationInTx`，每个已有关联 pair 日周期一次，不在 W3 判定阈值。
  合并参数未配置时保存 `association_pending / merge_policy_undecided`，不代选持续时长。
- Knowledge 按 Event scope 隔离；共享检索结果先检查 OwnerProject，不写其他项目的共享条目。
  通过 W4 `RewriteKnowledgeFluidTx` 更新 fluid/sources/consolidated_at；保留 confirmed、crystallized、版本及审计。
  同 key 字面声明冲突时保留原 fluid，W2 pending Conflict + Sleep 候选持有两侧声明/原始引用，交人类裁决。
- `SleepFactProjector` 是显式确定性源格式规则，无产品默认。测试仅解析 `raw.facts` 的字面 key/value；
  未配置或没有关键事实时显示不可计算，不把 EventID 留存冒充事实召回。不推断普通文本的事实或真假。
- LLM 解释器只在事务提交后接收复制的 cluster topic 视图，返回解释/错误给调用者，不落库、不反馈决策。
  每日 replay 不重调解释器。无模型请求发生，LLM 开关测试使用本地恶意解释器。

### 可测指标及计量范围

保留精度按去重的 `(owner_project, topic, literal key, literal value)` 召回；
保留集包括当前 Knowledge 及隔离候选中的字面声明，不代表声明已被证实或人工批准。
`RetentionMeasurable` / `StorageMeasurable`、原始计数与 `Representation` 全部持久化。

字节来自实际 raw BLOB 长度与 Knowledge `data` 的 UTF-8 字节长度（SQLite `CAST AS BLOB`），
加上实际持久化的本周期候选 JSON。计量为：mapped 原始载荷 + 涉及 Knowledge 工作表示 →
涉及 Knowledge 工作表示 + 本周期候选；原始 Event 保留，排除图、审计副本与 SQLite 页。
该值不代表整个 SQLite 文件缩小。缩减率 `(before-after)/before` 可为负，无达标线。

最后实测（合成输入，真实持久化字节，独立 SQL 核对）：

| 例 | 关键事实 before/retained | 召回 | UTF-8 字节 before/after | 缩减率 |
|---|---|---|---|---|
| 重复事实 + 一次孤立噪声 | 2 / 1 | 0.5 | 4465 / 592 | 0.867413 |
| 冲突保留原 confirmed 快照并隔离新声明 | 2 / 2 | 1.0 | 实际读取 | 不表示批准 |
| 紧凑输入生成工作表示 | 1 / 1 | 1.0 | 38 / 584 | -14.368421 |
| 空日 | 0 / 0 | 不可计算 | 0 / 0 | 不可计算 |

### 验证与首次失败

环境均为 `GOFLAGS=-p=1`、`GOMAXPROCS=2`，串行单包检查。

| 命令 | 结果 |
|---|---|
| `go test ./tests/attention-engine -run '^TestM3' -count=1 -v` | **9 PASS**；顶层 child helper **1 SKIP**，在两项父测试派生的新进程中实际执行 PASS |
| domain 核心定向测试（下方命令） | **12 PASS** |
| `go vet ./internal/attention/app` | PASS |
| `go vet ./internal/adapters/sqlite` | PASS |
| `go build ./internal/attention/app` | PASS |
| `go build ./internal/adapters/sqlite` | PASS |
| `git diff --check` | PASS |

```powershell
go test ./internal/attention/domain -run 'Cluster|Denoise|Resolve|Retention|Storage|ComputeMetrics|Interpreter|Consolidate' -count=1 -v
```

覆盖：当日边界/pending 无 Signal；新主题 A=.17、日末 M=.01638（显式测试政策 decay=0、strength=.2）；
真实 LastActive；孤立去噪；单事件邻接 birth 与不复制邻居证据；强关联无重叠不合并；
未决定合并政策显式 pending；图/fluid/日记录原子回滚；每日 replay；
新进程图/fluid/候选/metrics/pending Conflict 一致；Agent 裁决拒绝与人工 action 持久化；
LLM off 与恶意解释器的 run/node/knowledge 持久化 JSON 逐字一致；增长与空分母可见。

首次 RED 保留，不用后续定向 PASS 覆盖：

1. 波 A 入边去噪丢失、重复事实误计召回；修正见 `881f7fd`。
2. 增长 100→200 被夹到 0；修正见 `a203c5f`。
3. 初始 M3 fixture 缺 Observation.Provenance / pending SourceKind，测试尚未进入 Sleep；同时出现 TempDir cleanup 文件占用。
   补全来源字段与关闭连接后不再出现，未放宽存储验证。
4. 真实 `aa4a64e` 新主题 birth 把 UpdatedAt 设成日末，导致 `New signal predates persisted updated_at`。
   W2 `23bc4d7` 修正最早 own TS；正式消费后此路径通过。
5. 初始手算 M=.0546 漏了日末零增量步，应再乘 alpha=.3，最终 .01638；仅修测试参考值。
6. W4 新 API 的 GetKnowledge 参数顺序为 `(id, project)`；初始测试旧顺序产生 not_found，改调用后通过。

### 精确依赖、边界及下一步

普通合并的关键依赖：M0 `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`；
shared metrics `6973c171c0e05ea4b432bc3a62a3ac172abeb910`；
W1 `e9bb3809eadcbf48cf017b6e09ae0ef4112b1a90`；
W2 final `2c148c1034ec28b5d787309f1320198706d97e4a`（含 chronology `23bc4d7`）；
W4 final `82e4340e50943d355ecd81a2a47d65b7e3929f5a`（含 018 ConsolidatedAt）；
W7 harness fix `a3eab2c6d0df026e056a5a873a0d63bf58282688`。未经发布的原 working changes 未消费。

W0 负责把上述构造器接入现有固定入口、实际 HTTP 人工关口和最终集成；W3 不改公共契约/迁移/HTTP/cmd。
参数（strength/decay/初始 confidence、merge 持续定义等）及事实投影的产品选择仍未决定，测试参数不是产品默认。
晚到且早于既有 UpdatedAt 的新 Signal 由 W2 显式拒绝，raw 原件保留，不改时间或静默重放。
缺失原始冲突引用保存 `conflict_pending_evidence` 候选，不伪造完整 Conflict 证据。
未跑主控全仓检查/最终可执行程序构建、真实生产事件、真实 LLM、人工交互或七天科学对比；这些不由本轨 PASS 推导。

## 以下为阶段历史，当前结论以上文为准

事实源：任务书 §6、§5.4/§5.5、§4.4；主计划 `tasks/AttentionEngine-plan.md` W3 行。
基线分支 `attention-engine/v1-20261011`（当前 `77712f3`，仅计划/派发/恢复记录，M0 契约类型仍未并入）。

## 波 B 进行中：已消费 M0，日周期存储可用

- 正式消费 W0 `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`，普通合并为本轨
  `c861b747d6f2dfe6d07b4e5878f2128c5c8e38ce`；不消费未放行的 partial 基线。
- `sqlite.NewSleepRunTx(EngineTx)` 提供 `LoadSleepRun(day)` 与 `SaveSleepRun(run)`，
  复用已有 `BEGIN IMMEDIATE`，只写 migration 015 的 Sleep 四表，不开嵌套事务。
- `TestM3DailyRunStoreRollbackAndReopen` **PASS**：缺失 Conflict 引用使整笔回滚，
  daily UNIQUE 拒绝重复，关闭重开后 run/candidate/metrics 一致，负缩减值保存。
  该测试的 100→200 是存储类型测试值，不是实际流水线压缩观测。
- W1/W2/W4 已收到同事务读写 Handoff：当日 mapped Event/Signal；既有图/seed/lifecycle/
  pending Conflict；按项目隔离的 Knowledge fluid 更新且保留 crystallized/状态。

命令：`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine -run TestM3DailyRunStoreRollbackAndReopen -count=1 -v`。
当前完整 M3 **未完成**：app 编排、实际表示字节与事实召回、LLM 开关结果一致以及完整图/fluid/冲突
冷启动回归还未执行；入口/HTTP 人工关口归 W0，当前共享 HTTP 仍为 501。

## 2026-10-11 Codex 接续：波 A 修正，M3 仍未完成

当前工作树/分支保持 `attention-engine-w3-sleep` /
`songconmaisaix31-design/attention-engine-w3-sleep`，接续起点 `5399e31`。
本次实际客户端 Orca Codex / gpt-6.1-sol / YOLO（主控 runtime 通知确认），未自行覆盖。

- 原 `3f6cc6f` 已取消多 topic 自动合并，本次保留该规则；关联持续 >0.8 且证据重叠的真正合并只消费 W2 生命周期判定。
- 修正去噪只认出边的问题：既有节点指向当日单事件主题的入边也算关联；空邻居和自指不算关联。聚类和去噪采用同一 topic 空白归一化口径。
- 双指标的关键事实按去重集合计数；重复事实不增加召回率或事实数。按主控预审修正，存储缩减为 `(before-after)/before`，增长返回负数，不再夹到 [0,1]；原始字节计数保留。`RetentionMeasurable` / `StorageMeasurable` 显示可计算性，空 before 事实不再返回 vacuous 1，数值 0 且不可计算。后续以实际持久表示的序列化字节衡量，不冒充 SQLite 文件总量减少。
- 新增 `tests/attention-engine/M3-golden.json`，由 `TestConsolidateGoldenInputIsOrderIndependent` 实际读取并验证决策和双指标，仅为 `contract_local` 合成输入；不属于实际 Event→图→Knowledge→人工裁决全流程证据。

首轮定向测试：`TestConsolidateKeepsInboundRelatedTopic` **FAIL**（入边关联丢失）；
`TestComputeMetricsCountsDistinctKeyFacts` **FAIL**（预期 0.5、实得 0.6667）；
`TestConsolidateGoldenInputIsOrderIndependent` **PASS**。修正后的 domain 包 test/vet/build 均 **PASS**。
新增测试共 3 项，Sleep 核心共 12 项；单包执行，未执行全量检查。

命令（环境 `GOFLAGS=-p=1`、`GOMAXPROCS=2`）：

```powershell
go test ./internal/attention/domain -count=1
go vet ./internal/attention/domain
go build ./internal/attention/domain
git diff --check
```

波 B **待依赖**：主控尚未发布 M0 精确 SHA。已向主控请求日周期/冲突表、
共享接口、W1 当日 Event 读取、W2 生命周期及 Conflict、W4 fluid 写入 Handoff；
公共契约/迁移/HTTP/固定日周期入口仍归 W0。尚未写 app/store，
每日一次、事务失败回滚、冷启动一致及 LLM 开关全流程一致均 **NOT_RUN**。
阶段使用 `status/merge_ready`，不发送最终 `worker_done`。

主控已明确回复：当前无稳定 M0 SHA，保持完整 M3 进行中；SHA 发布后主动通知并由
W1/W2/W4 提供接口。波 A 源码 `881f7fd5411f6655f3a46599fd45d17de7d96cc6` 已 push，
后续仅补金标准数据的实际读取；定向 `go test ./internal/attention/domain -run TestConsolidateGoldenInputIsOrderIndependent -count=1 -v` **PASS**。

预审补充首次 RED：`TestStorageReductionPreservesGrowth` 输入 100→200 时预期 -1，旧代码实得 0。
修正后 `go test ./internal/attention/domain -run 'Retention|Storage|ComputeMetrics|ConsolidateGolden' -count=1 -v`
**PASS（5 项）**，domain `go vet` 与 `git diff --check` **PASS**。
该结果仍是纯函数测试；100/70 只用于合成金标准，完整流水线必须观测实际 before/after 表示字节。

## 交接（阶段结果 partial，后续 Codex --yolo 承接原写域与 worktree）

阶段结果：**partial**。波 A 确定性核心已交付并验证；完整 M3（app/store 编排、每日持久化、
断 LLM 全流程验收）未完成，不冒充完整 M 已完成。

### 已完成
- `internal/attention/domain/sleep.go`：`EventTopic`、`Cluster`、`ClusterTopics`（union-find topic
  邻接聚类）、`Denoise`（孤立信号去噪）、`NodeDecision`/`ResolveClusters`（逐 topic `NodeCreate`，
  多 topic 不合并）、`Consolidate`（cluster→denoise→resolve 确定性入口）、
  `ClusterInterpreter`/`ClusterExplanation`（LLM 只解释）、`ConflictCandidate`。
- `internal/attention/domain/sleep_metrics.go`：双指标 `RetentionPrecision`/`StorageReduction`/`ComputeMetrics`。
- 修正预审缺陷：多 topic 不再自动 `NodeMerge`（合并归 W2 §5.4 生命周期规则）。
- 测试 9 项全绿（`GOFLAGS=-p=1 GOMAXPROCS=2` 下 `go test`/`go vet`/`go build` 均 PASS）。
- 验收金标准示例已写入本文档（§验收示例）。

### 未完成（依赖 M0/W1/W2/W4 契约类型，未开展）
- `internal/attention/app/sleep_*.go` 编排（当日 Event→Consolidate→图更新 + KnowledgeEntry fluid 更新）。
- `internal/adapters/sqlite/sleep_store*.go`（每日一次持久化 + 冷启动一致；迁移 014+ 属 W0）。
- `tests/attention-engine/M3*` 验收脚手架（断 LLM 跑 pipeline、双指标、冲突记录、无 LLM 决策入库）。

### FAIL 或 UNKNOWN
- 无 FAIL。
- UNKNOWN：M0 契约类型（Event/Signal/AttentionNode/KnowledgeEntry/Evidence/Conflict/引擎接口）的
  精确 SHA 与未完成清单尚未由 W0 发布；`tests/attention-engine/` 共享 harness 尚未存在。
- 并行 `go test ./...` 触发本机 Windows 内存 panic（已知，`-p=1` 可过），与本次改动无关。

### 下一步接口要求（给后续 Codex / 各轨）
- **W0**：发布 M0 契约/迁移精确 SHA + `tests/attention-engine/` 共享 harness + 冲突裁决 HTTP 关口。
- **W1**：Event 类型与当日事件读取入口（`EventTopic` 视图映射来源）。
- **W2**：AttentionNode（`related_topics` 作聚类邻接输入）、Conflict 记录生成（§5.5）、§5.4 合并规则。
- **W4**：KnowledgeEntry fluid 重写 / crystallized 追加的写入口。
- **波 B 接续**：app 编排以 `domain.Consolidate` 为唯一确定性入口；interpreter 只装饰解释、不改决策。

## 目标与验收

每日一次的 Sleep 流水线：raw events（当日）→ 按 topic 邻接聚类 → 去噪（孤立信号不生成节点）→
生成/合并概念（LLM 只做聚类解释，不做入库决定）→ 更新 attention graph + KnowledgeEntry fluid 层 →
冲突裁决候选 → 用户关口。

验收（root 独立复验）：断 LLM 运行 pipeline；双指标输出可测；冲突记录生成；无 LLM 决策入库。
第一版双指标只要求可测，不设达标线。

## 流水线步骤 → 代码落点

| §6 步骤 | 责任轨 | W3 写域落点 | 状态 |
|---|---|---|---|
| raw events（当日） | W1（Event） | 输入视图 `EventTopic` | 已定义 |
| cluster（topic 邻接） | W3 | `domain/sleep.go` `ClusterTopics` | 已实现 |
| 去噪（孤立信号不生成节点） | W3 | `domain/sleep.go` `Denoise` | 已实现 |
| 生成/合并概念 | W3 | `domain/sleep.go` `ResolveClusters` | 已实现 |
| update graph + fluid 层 | W2/W4 + W3 编排 | `app/sleep_*.go`（波 B） | 待 M0/W2/W4 类型 |
| 冲突裁决候选 → 用户关口 | W3 队列 + W0 HTTP 关口 | `app/sleep_*.go`（波 B） | 待 W2 Conflict + W0 端点 |
| 双指标 | W3 | `domain/sleep_metrics.go` | 已实现 |

## 本波（波 A）已交付的确定性核心

`internal/attention/domain/sleep.go` / `sleep_metrics.go`（纯函数、无 I/O、不依赖 M0 契约类型）：

- `ClusterTopics(events []EventTopic, related map[string][]string) []Cluster` — union-find 按 topic
  邻接聚类，输出按首 topic 排序（确定性）。
- `Denoise(clusters, related) []Cluster` — 丢弃「单事件且无邻居」的孤立信号，不生成节点。
- `ResolveClusters(clusters) []NodeDecision` — 逐 topic 生成 `NodeCreate` 候选概念；多 topic 集群
  只表示相关、**不合并**（合并是 W2 生命周期规则）。**函数签名不含 interpreter**，LLM 输出永远
  无法改变决策。
- `Consolidate(events, related) Consolidation` — cluster→denoise→resolve 的单一确定性入口，
  返回存活集群 + 逐 topic 决策；波 B app 层在此接入 interpreter（只装饰解释，不改决策）。
- `ClusterInterpreter` 接口 + `ClusterExplanation{Label, Note}` — LLM 唯一角色：解释集群为何聚合。
  输出不含任何决策字段；nil（断 LLM）不影响任何决策。
- `ConflictCandidate{Topic, Summary}` — sleep 周期的待裁决队列项，只记录冲突、不裁决。
- `SleepMetrics` / `ComputeMetrics` / `RetentionPrecision` / `StorageReduction` — 双指标：
  保留精度 = |before∩after|/|before|（空 before 记 1）；存储缩减比 = (before-after)/before，夹到 [0,1]。

测试：`internal/attention/domain/sleep_test.go`（package domain，沿既有 `rules_test.go` 风格），
覆盖聚类邻接/空白 topic、去噪、逐 topic 决策不合并、interpreter 不改决策、断 LLM 确定性、
双指标边界。

## 无 LLM 决策入库的保证

入库决定权 = 确定性规则 + 用户关口。代码层面三条硬约束：

1. `ResolveClusters` 不含 `ClusterInterpreter` 参数，决策只由集群形状决定；多 topic 集群不合并，
   合并判定归 W2 生命周期规则（§5.4）。
2. `ClusterInterpreter.Explain` 返回值 `ClusterExplanation` 只有 `Label/Note`，无决策字段。
3. `ConflictCandidate` 不裁决，只进队列等用户关口。
4. `Consolidate` 是断 LLM 也能跑的确定性概念阶段（无 interpreter 参与）。

验收测试（波 B，`tests/attention-engine/M3*`）会以 nil interpreter 跑完整 pipeline 断言决策不变。

## 验收示例（波 B 断 LLM 回归的金标准输入）

给定某日事件与邻接 `related = {paper-reading: [agent-swarm]}`：

| topic | 事件数 | 邻接 | 结果 |
|---|---|---|---|
| agent-swarm | 3 | — | 集群 {agent-swarm, paper-reading}，`NodeCreate(agent-swarm)` |
| paper-reading | 1 | agent-swarm | 同集群，`NodeCreate(paper-reading)`，不合并 |
| one-off-noise | 1 | 无 | 去噪丢弃，不生成节点 |

- 决策：2 个 `NodeCreate`（agent-swarm、paper-reading），**无合并**；关联保留供 W2 判定。
- 断 LLM 一致性：`Consolidate` 无 interpreter 参与，决策与上表逐点一致。
- 双指标示例：before facts {a,b} → after facts {a}，保留精度 0.5；beforeBytes 100 → afterBytes 70，
  存储缩减比 0.3（`ComputeMetrics` 已覆盖）。

## 波 B 待办（依赖 M0/W1/W2/W4 契约类型）

1. `app/sleep_*.go` 编排：从 event store 读当日 Event → `Consolidate`（cluster/denoise/resolve）
   → 应用 `NodeCreate` 决策到 attention graph（W2 AttentionNode）+ 写 KnowledgeEntry fluid 层（W4）
   → 多 topic 集群的关联（related_topics）只保留不合并，合并候选交 W2 生命周期。
2. 冲突队列：消费 W2 的 Conflict 记录 → 生成 `ConflictCandidate` → 落 sleep store → 用户关口
   （W0 的 HTTP 端点负责人工裁决）。
3. `internal/adapters/sqlite/sleep_store*.go`：迁移 014+ 下 sleep cycle run、conflict candidate、
   metrics 记录的持久化；每日一次 + 冷启动一致。
4. `tests/attention-engine/M3*` 验收脚手架：断 LLM 跑 pipeline、双指标可测、冲突记录生成、
   无 LLM 决策入库（决策与 interpreter 输出无关）。

## Handoff 需求（跨轨，只读不写对方域）

- **W0**：M0 契约类型（Event/Signal/AttentionNode/KnowledgeEntry/Evidence/Conflict/引擎接口）并入
  `attention-engine/v1-20261011` 后通知；`tests/attention-engine/` 共享 harness；冲突裁决 HTTP 关口。
- **W1**：Event 类型与当日事件读取入口（`EventTopic` 视图的映射来源）。
- **W2**：AttentionNode（topic 邻接 `related_topics` 作为聚类输入）、Conflict 记录生成（§5.5）。
- **W4**：KnowledgeEntry 两级存储，fluid 层重写与 crystallized 追加的写入口。

## 验证命令

```
go test ./internal/attention/domain/ -run 'Cluster|Denoise|Resolve|Retention|Storage|ComputeMetrics|Interpreter|Consolidate' -v
go vet ./internal/attention/domain/
```
