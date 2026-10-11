# AttentionEngine-W2 · 动力学引擎里程碑设计（M2）

事实源：任务书 `tabbit-code (2).md` §5.1–5.5、§8 M2；`tasks/AttentionEngine-plan.md` W2 行。
分支：`songconmaisaix31-design/attention-engine-w2-engine`，基线 `dab8fea`。

## 状态（2026-10-11）

波 A 已落码：纯数学状态更新 + 参数化函数 + 30 天 golden 数据（不依赖 M0 类型，不自建共享
Event/Node/Store）。域文件 `dynamics/antimatthew/lifecycle/topic/conflict` 已实现，单测与
`tests/attention-engine/M2_golden_test.go` 通过（GOFLAGS=-p=1 GOMAXPROCS=2）。

波 B 阻塞于 M0 契约类型未发布：app 层 `engine_*.go`（get_attention_state/explain_focus）与
`engine_store*.go` 依赖 M0 的 Event/Signal/AttentionNode/EvidenceRef 类型。主控发 merge_ready
后执行 `git fetch origin && git merge origin/attention-engine/v1-20261011` 再开码。

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
