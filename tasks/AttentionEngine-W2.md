# AttentionEngine-W2 · 动力学引擎里程碑设计（M2）

事实源：任务书 `tabbit-code (2).md` §5.1–5.5、§8 M2；`tasks/AttentionEngine-plan.md` W2 行。
分支：`songconmaisaix31-design/attention-engine-w2-engine`，基线 `dab8fea`。

## 状态（2026-10-11）

### M2 持久化接续（当前 Dispatch）

- 正式 M0 source `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0` 已由 root 放行并普通合并。本轨采用其 `StoredAttentionNode`/`StoredEvent`/解释包装与 migration 015，不消费 W0 草稿。
- app/store 第一阶段落码：`NewDynamicsEngine`、`NewDynamicsRepository`、`NewDynamicsTx`，事务内 ApplySignals + read-only AttentionState/AttentionNode/ExplainFocus + pending Conflict/人工裁决。参数 nil 保持未决定并明确返回配置错误。
- `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine -run '^TestM2' -count=1` PASS：原纯 golden、真实 app+SQLite 30 天逐点回放（day 10/20 关闭/重开库）、去重/读取不写节点/失败只降 confidence/整批回滚。输入是任务书规定的伪造流，非真实会话。
- 首次编译 RED：误用不存在的 `newID`，改为复用既有 `crypto/rand.Text` 后编译与上述测试 PASS。生命周期集成仍在开发，尚未 M2 完成。

- 接续基线 `3ed950250172d2efcf73a171e29e149e1a4916ea`，工作树干净；客户端 Codex（用户已应用 `codex --yolo`），不启动并发 Agent。
- 只读检查 W0 尚未发布的 M0 草稿；未复制或消费草稿类型。等待协调者给出精确可消费 SHA。
- M0 Handoff：`updated_at` 与 `last_active` 必须独立持久化；解释需原始 Event、强度、有效贡献/失败 confidence 路径和带时间戳的关联。生命周期及持续合并观察可在 W2 节点 JSON 内保存，迁移/共享类型仍由 W0 决定。
- 实现次序：消费获准 M0 → 事务内幂等 ApplySignals/读侧 → 持久化生命周期与冲突 → 30 天实际应用回放/冷重启 → commit + push。
- 验收保留现有独立 golden；额外回放必须运行生产 app + 实际 SQLite，输出逐点实际 A/M/confidence/updated_at/last_active 供 root 第二实现复算。读侧前后库状态相同；失败任务只给 confidence 负贡献；归档不删，重浮现从 floor 唤醒。
- 参数 λ、初始 confidence、成功 confidence 增量、分裂/密度/confidence 阈值、持续合并时长未被用户锁定，必须由显式配置输入；下文旧 A1/A4/A5 不作为运行默认值。
- 基线验证：`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain ./tests/attention-engine` PASS（缓存）；仅纯规则证据，非 M2 完成。
- 协调者回复：明确等待 M0 稳定 SHA；可继续自足生命周期参数化与 app 接口草案，不消费未提交类型，不重复询问等待条件。
- 自足持续合并规则 `ObserveMerge`：连续强关联且重叠、满足显式配置的时长与观测数才许可合并；相同观测时间不累计，中断清空。`go test ./internal/attention/domain -run TestSustainedMerge -count=1` PASS；尚无 SQLite 合并证据。

#### 待 M0 发布的 app/store 实现草案（不冻结共享接口）

1. Store 使用 W0 `EngineRepository.WithTx` 串行化读改写；节点 JSON 展平共享字段并保存 W2 生命周期、历史 A 最大值、已使用信号与关联观测。证据去重和节点更新同事务，不新增自研调度或完成证明。
2. `ApplySignals` 消费可追溯已落库 Event/Signal，按时间分组求和，一组更新一次 M；空批只推进尚未应用的时间区间。新增过期信号明确拒绝，已应用信号重放不重复贡献。`last_active` 只随真实行为更新，空批与重启不刷新活动时间。
3. `AttentionState` / `AttentionNode` / `ExplainFocus` 仅查询持久状态；读取不推进衰减。解释返回每条事件/信号的时间、权重、强度及失败任务的 confidence 路径；关联来自持久化项目/Agent/节点事实。
4. 手动种子校验 human principal；同 scope 密度诞生按显式阈值、两簇分裂保留证据与关联；持续合并取历史 attention 最大，源节点保留；archive 不删除、重浮现仅抽取触底非 rising 候选、抽中 cold 从 floor 唤醒。
5. confidence 骤降或事实矛盾新增 pending Conflict；裁决只通过 human principal，不自动覆盖冲突证据。
6. 实际 SQLite 验收使用任务书授权的伪造事件流。应用回放、查询、关闭/重新打开库、重放/闲置推进输出单独保存，与真实会话 M1、HTTP 人工关口和全产品验收分开。

波 A 已落码：纯数学状态更新 + 参数化函数 + 30 天 golden 数据（不依赖 M0 类型，不自建共享
Event/Node/Store）。域文件 `dynamics/antimatthew/lifecycle/topic/conflict` 已实现，单测与
`tests/attention-engine/M2_golden_test.go` 通过（GOFLAGS=-p=1 GOMAXPROCS=2）。HEAD `acb4561`。

### 真实 partial（截至本次 settle）

- 已交付：§5.1 A(t)/M(t)/confidence 纯函数、§5.3 A_floor/θ/种子化重浮现、§5.4 生命周期
  状态机（归档/唤醒/合并/分裂/密度诞生）、§5.5 confidence 骤降冲突判定；固定常量按任务书
  （α=.3、floor=.05、θ=.15、合并>.8、90 天归档），λ/分裂/密度/confidence 阈值保持参数化未冻结。
- 已交付：30 天 golden 复算流 `tests/attention-engine/M2_golden_stream.json`（Unix 毫秒 ts +
  逐步 A/M/confidence，独立 Python 参照生成，非被测公式），`M2_golden_test.go` 用生产域函数
  逐点回放断言；root 可用独立 JS 公式复算。
- 未完成：波 B（app 读侧 `engine_*.go` 的 get_attention_state/explain_focus、`engine_store*.go`
  持久化）阻塞于 M0 契约类型（Event/Signal/AttentionNode/EvidenceRef/引擎接口）未发布。

### 下一步

1. 主控发 M0 merge_ready 后，`git fetch origin && git merge origin/attention-engine/v1-20261011`
   取得契约类型，再开波 B：引擎接口实现 + app 读侧证据链组装 + engine_store 持久化。
2. 波 B 衰减推进用 `updated_at`（独立于行为时间 `last_active`），ts 为 Unix 毫秒、Δt=ms/86400000。
3. 后续开发按用户要求切换 `codex --yolo`；参数问题（λ/强度/分裂/密度/confidence 阈值）待答，
   保持显式配置输入不冻结默认值。

## 目标与验收（任务书原文，不降级）

- §5.1：A(t) = A(t-1)·e^(−λ·Δt) + Σ s_i·w_i；M(t) = α·M(t-1) + (1−α)·(A(t)−A(t-1))，α=0.3。
- §5.3：A_floor=0.05；每日随机重浮现 3 节点；momentum 门限 θ=0.15。
- §5.4：topic 生命周期（诞生/分裂/合并/消亡→cold archive 不删可唤醒）。
- §5.5：冲突检测（不覆盖、不静默合并，生成 Conflict 记录）。
- 读侧：get_attention_state、explain_focus 证据链组装（供 W6 调用）。
- 验收：给定 30 天伪造事件流，A(t)/M(t) 序列与手工复算逐点一致（golden 测试，固定时钟注入，
  不造假时间）；A_floor/重浮现/momentum 门限可验证。

## M0 契约依赖（Handoff 需求清单，不冻结形状）

W2 编码依赖以下由 W0 锁定的契约类型与接口（全部 Handoff，不在本轨定义）：

1. `Event`（§4.1）：source/session/ts/kind/tokens/scope/topic/raw。
2. `Signal` + 权重表（§5.2 硬编码）：M1 归一化产物，含信号强度与权重。W2 只消费。
3. `AttentionNode`（§4.2）：attention/momentum/confidence/decay(λ)/evidence/
   related_topics/last_active/emotional_weight。
4. `EvidenceRef = {event_id, weight_at_use, ts}`（§4.5）。
5. 引擎接口（W0 "引擎接口定义"）：W2 实现，不自行定义对外接口形状。
6. 存储：`engine_store` 持久化 AttentionNode 与 Conflict，W2 依赖 W0 的 sqlite 连接/迁移基础设施。

## 确定性规则（§5.1–5.5，可手工复算）

时间单位：天（浮点）。λ 单位 1/天。Δt = (t − t_prev) / 86400 秒。

### 5.1 状态更新

```
A_decayed = A(t-1) · exp(−λ · Δt)
A(t) = clamp(A_decayed + Σ s_i·w_i, A_floor, 1.0)
M(t) = α · M(t-1) + (1 − α) · (A(t) − A(t-1))     // α = 0.3
```

- 上限 1.0 来自 §4.2 attention ∈ [0,1]；下限 A_floor=0.05 来自 §5.3。
- 失败信号特例（§5.2）：Agent 任务失败 w=−0.1 只降 confidence，不进 A(t) 求和项。
- λ 默认值：假设 A1，λ=0.1（半衰期 ≈ 6.93 天），全主题初始统一。

### 5.3 反马太

- A_floor = 0.05：节点首个信号后 A 永不低于 0.05（旧主题不归零）。
- 每日重浮现：sleep cycle（W3）触发，W2 提供纯函数 `SelectResurface(nodes, rng, k=3)`
  （种子化 rng，golden 测试可复算）。重浮现把沉底节点重新注入候选，唤醒后从 A_floor 起步。
- momentum 门限 θ=0.15：`M(t) > θ` 判定 rising。读侧排序不对 rising 新主题按低 A 惩罚；
  生命周期上 rising 节点不进入 cold archive。

### 5.4 生命周期状态机

状态：`born → active → cold(archive)`；cold 可被重浮现唤醒 → active（A 从 A_floor 起步）。

| 迁移 | 触发（确定性规则） | 归属 |
|---|---|---|
| 诞生 | 手动种子 `SeedTopic`；同一 scope 事件簇密度 > 阈值 τ_density | W2 |
| 诞生（聚类） | sleep cycle 聚类生成 | W3（Handoff 输入） |
| 分裂 | 内部证据呈两簇且相关性 < τ_split → 拆两节点，关联保留 | W2 |
| 合并 | 关联强度持续 > 0.8 且证据重叠 → 合并，历史 attention 取 max | W2 |
| 消亡 | A ≤ A_floor 且 90 天无活动 → cold archive（不删） | W2 |

- 0.8 为任务书固定值。τ_split、τ_density、关联强度/证据簇的输入来源为 W3/W4 Handoff。

### 5.5 冲突检测

```
Conflict {
  id, topic, kind, new_evidence EvidenceRef, existing EvidenceRef,
  created_at, status(pending|resolved), resolution
}
kind ∈ { confidence_drop, fact_contradiction }
```

- confidence_drop：单步 confidence 降幅 > δ_conf → 生成 Conflict（δ_conf 假设 A5，初值 0.3）。
- fact_contradiction：W4 检测 crystallized 事实矛盾，调用 W2 生成 Conflict 记录。
- 不覆盖、不静默合并；记录进 sleep cycle 裁决队列（W3）。

## 读侧用例

- `get_attention_state()` → [(topic, attention, momentum)] 按 attention 降序（W6 MCP 出口）。
- `explain_focus(topic)` → 证据链：事件引用（带 ts）+ 权重路径（signal→w）+ 关联节点，
  每条带时间戳（§7 存在理由）。

## Golden 测试设计（验收脚手架）

测试文件：`tests/attention-engine/M2_*_test.go` + 域内 `internal/attention/domain/*_test.go`。

1. 固定时钟注入：定义 `Clock` 接口（Now() time.Time），测试注入固定时间序列，不用 `time.Sleep`
   或改系统时钟（不造假时间）。
2. 种子化 RNG：重浮现用可注入 seed 的 RNG，golden 用固定 seed 逐点复算。
3. 30 天伪造事件流：确定性构造（固定 topic、kind、ts 序列），逐日推进引擎。
4. 手工复算参照：golden 表为独立手算的 A(t)/M(t) 逐点期望值，测试逐点比对（非自我引用）。
5. 门限/下限用例：A_floor 触底不归零；重浮现命中可唤醒且从 A_floor 起步；M(t)>θ 判定 rising；
   合并取历史 attention 最大；冲突生成记录且不覆盖原值。

## 独占写域（本轨可改，其余只读）

- `internal/attention/domain/{dynamics,lifecycle,topic,antimatthew,conflict}*.go`
- `internal/attention/app/engine_*.go`
- `internal/adapters/sqlite/engine_store*.go`
- `tests/attention-engine/M2*`
- `tasks/AttentionEngine-W2.md`

## 开放问题（待用户/主控答复，不据此冻结接口）

- Q1：λ 按主题分化的确定性规则（初始统一，后续如何分化）。
- Q2：信号强度 s_i 的取值（v0.1 假设每事件 =1；或由 tokens 缩放）。
- Q3：分裂阈值 τ_split、诞生密度阈值 τ_density 初值。
- Q4：confidence 骤降阈值 δ_conf 初值。
- Q5：momentum 门限 θ 是否仅读侧+生命周期信号（假设 A3），还是修改 A(t) 更新。

## 假设记录（开工即记，不冻结）

- A1：λ 默认 0.1/天，全主题初始统一（Q1 待答）。
- A2：s_i = 1（每信号事件贡献一次权重），tokens 暂不参与（Q2 待答）。
- A3：θ 为读侧排序 + 生命周期保护信号，不改 A(t) 公式（Q5 待答）。
- A4：τ_split 初值 0.3、τ_density 初值待定（Q3 待答）。
- A5：δ_conf 初值 0.3（Q4 待答）。
