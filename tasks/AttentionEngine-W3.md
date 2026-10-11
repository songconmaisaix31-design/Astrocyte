# Attention Engine W3 · Sleep Consolidation 流水线（M3）

事实源：任务书 §6、§5.4/§5.5、§4.4；主计划 `tasks/AttentionEngine-plan.md` W3 行。
基线分支 `attention-engine/v1-20261011`（当前 `77712f3`，仅计划/派发/恢复记录，M0 契约类型仍未并入）。

## 阶段状态（本次 dispatch 波 A）

- 波 A 已交付并验证（见下）；阶段状态 `merge_ready`，等 M0 精确 SHA 后接 `app/sleep_*.go` 与
  `sleep_store*.go`（波 B）。本阶段不提前 worker_done。
- 预审修正：`ResolveClusters` 多 topic 直接 `NodeMerge` 违反 §5.4 —— 邻接同簇只说明相关，
  不得自动合并既有 topic。真正合并须关联持续 >0.8 且证据重叠，由 W2 生命周期规则判定；
  未达条件保留关联。已改：多 topic 集群拆为逐 topic `NodeCreate`，不产生合并决策。

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
