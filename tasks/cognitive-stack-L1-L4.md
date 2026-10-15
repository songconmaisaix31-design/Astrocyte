# Astrocyte 认知栈计划（L1–L4）：从接入闭环到 Agent Swarm 的纵向完成

事实源登记：validation/v1-20261012 纯文档。本文取代口头路线讨论，
成为 v0.2–v0.4 的纵向基准；各层实施任务书从本文派生。

## 0. 唯一验收命题

**经过 Astrocyte 的 Agent，工作强于未经过的 Agent。**

产品完成度 = 已拿出该证据的层数。四层各一层证据，同一方法论
（预注册、三权分立评分、真实任务、宏平均统计），四层合起来构成
一条完整证据链：接入 → 认知 → 组织 → 群体。

功能不进入本计划，除非它贡献某一层的证据。

## 0.1 证据架构（V4 独立）

| 槽位 | 层级 | 状态 |
|---|---|---|
| 一 · 会话级 | V4 七日三臂 | **空位**。进行中，本计划不引用、不等待、
不以任何 V4 结果为前提。V4 落地后按真实结果填入——胜，填迁移
证据；负，填差距定位。两种结果都不改变本计划任何开工门禁 |
| 二 · 工作级 | L2 纵向对照 | **本计划主证据生成器**（§3.4） |
| 三 · 组织级 | L3 对照 | 多 Agent 组织化收益（§4.4） |
| 四 · 群体级 | L4 对照 | Swarm 认知增益（§5.3） |

## 1. 学术基线（每层每个机制一个来源，均已核验）

| 层 | 来源 | 取用机制 |
|---|---|---|
| L1 | Orca v1.4 源码架构 | intent 驱动/单一权威 store/三值活性裁决/
类型化注册表/中性 worktree/降级即正常路径 |
| L1 | abtop、AgentSight | 观测指标体系参照（token 速率/上下文占用/
异常检测），采集内核自研 |
| L2 | Generative Agents (Park et al. 2023) | recency×importance×relevance 三因子
（Attention 层既定基线） |
| L2 | SRD / LEAFE (arXiv 2610.08077 等) | 失败轨迹是信息密度最高的经验来源 |
| L2 | AgeMem (ACL 2026, arXiv 2601.01885) | 记忆操作工具化：store/retrieve/
update/summarize/delete 为显式操作 |
| L2 | Human-Inspired Memory (Microsoft Research) | 睡眠巩固、干扰遗忘、再巩固、
engram 成熟；巩固验收双指标（保留精度 × 存储缩减） |
| L2 | MAGMA (ACL 2026) | 语义/时间/因果/实体多图正交——经验关联结构 |
| L2 | Cognitive Scaffold (ACL 2026) | fluid/crystallized 两级存储——
经验到知识的固化路径 |
| L2 | Mem0 | 事实冲突消解——经验/知识冲突入待裁决队列 |
| L3 | AgentEvolver (NTU/SMU) | 八类资产统一版本化生命周期；
**采纳架构，拒绝验收**（其 82.08% vs 79.07% 净增量不可归因，
反面教材写入验收纪律） |
| L3 | Homogenization (McGill/Mila/UW/Apple) | anti-crowding 评分项；
异构 harness ≠ 观点异构（显式设计约束） |
| L3 | MeshHeal | 双时标信任：眼前任务同行审查 + 跨任务
累积信誉分离 |
| L3 | Agensh (Microsoft Research) | 自组织发生在共享设施之上——
公告板即共享设施的独立实验背书 |
| L3 | MAScope (北航) | 拓扑是诊断变量——scope 共现恢复协作图 |
| L4 | Waggle | 承诺三态（维持/修改/撤销）→ 租约语义；
v0.4 远期：匿名局部规则蒸馏 |
| L4 | Adversarial Influence (Princeton) | 对抗影响随占比而非人数缩放——
韧性来自边界（scope/预算/审计三道硬门），不来自节点数 |
| L4 | Context Language Models (UW/Meta/MIT) | 定位对比：CLM 内生
自由编辑上下文（审计链可破坏）vs Astrocyte 外部不可篡改认知层
（event_id 全可溯、crystallized 只追加）——互补路线，叙事坐标 |
| L4 | Morphogenesis（自有） | stigmergy 账本/信息素/fencing 租约/
经验晋级链——L4 的执行底座 |

## 2. L1 · 接入闭环（本周，冻结兼容）

**命题：Astrocyte 认识一个新 Agent，并看见它的 Session。**

现状：agents 域注册与发现、events 域三源采集、Works Relay v1.2
（人工派发 + 任务级事件回流）均已在库。

唯一缺口：Works Relay 持有终端的完整原生 Session 尚未进入采集
视野。补法——接入层自有采集管道：

- Works Relay 终端的 session 目录纳入接入层采集，按统一 Event
  模型归一化，写入 **workbench 独立库**（V4 正式库采集根零改动，
  10-23 后两库合并）
- 新 Agent 注册（注册表加一条 AgentSpec）→ 派发 → 其 Session
  全事件自动出现在 Today 视图，全程无人工采集动作
- 三值活性裁决（live/unverifiable/exited）覆盖全部接入终端

**L1 验收（演示脚本）**：一个从未见过的 CLI，从注册到"Today 视图
看见它在想什么"，三分钟。

实施载体：Works Relay v1.3 增补条款，随 D1–D3 执行轨落地。

## 3. L2 · 认知三级（10-16 起，本计划核心）

**命题：Attention → Experience → Memory Consolidation 从三个
平行模块变成一条纵向流水线，每层有明确输入输出与交接。**

### 3.1 层级语义

```
Attention（已验收，V1R/V4 在用）
  我关注什么 —— A(t)/M(t) 状态，决定"什么重要"
  输出：接力首条输入的认知投影（A 臂格式原样复用）
        ↓
Experience（骨架在库，approved=0，从未闭环）
  做事学到什么 —— 失败与成功轨迹的结构化沉淀
  输出：可检索、可注入的 approved 经验
        ↓
Memory Consolidation（Sleep 已跑 32 次，知识沉淀薄）
  固化成什么 —— 经验+事件经 Sleep 固化为 KnowledgeEntry
  fluid 可重写 / crystallized 只追加；检索触发的再巩固
  更新置信（MSR 机制），冲突入待裁决队列（Mem0）
```

### 3.2 Experience 管道（唯一真实开发量）

```
Works Relay 任务完成收据（成败/产物/轨迹引用）
  → 经验抽取适配器：Interpreter 起草候选
     {pattern, verdict: avoid|prefer|condition, condition,
      evidence: []event_id}
  → quarantine（M4/M5 既有接口，workbench 侧调用）
  → 用户批准（人工关口既有实现）
  → approved：get_experiences 可检索 + 注入接力首条输入
```

裁决一（经验抽取的 LLM 角色）：Interpreter 只起草、不决策，
沿"提交后解释"既有纪律的延伸——候选永不自动晋级，批准权在人。

经验优先捕获失败轨迹（SRD）；同 pattern 多次验证提升置信（engram
成熟）；经验间关联按 MAGMA 正交维度挂接。

### 3.3 Consolidation 增强

Sleep 流水线既有结构不动，输入侧并入经验事件；巩固验收学 MSR
双指标（保留精度 × 存储缩减），第一版只要求可测。

### 3.4 L2 纵向对照实验（主证据生成器，不依赖 V4）

三臂工作级，真实接力任务：

| 臂 | 首条输入 |
|---|---|
| 认知组 | 交接手册 v2 = 手册 + 认知投影 + approved 经验 |
| 传统组 | 项目相关聊天记录 + 文档塞满上下文（最强版本） |
| 冷启动组 | 仅任务描述 |

指标：完成质量（人工评分 + 测试通过率）、上下文 token 消耗、
迭代轮次、撞坑命中率（approved 含避坑条目时是否命中——经验复用
专项测量）。评分三权分立：质量归人工 + 客观测试，过程指标归
收据机判，撞坑归事件库回溯。

裁决二（协议时点）：现在起草、冻结解除后锁 SHA 开跑——设计先行
节奏，与 v0.2 四轨同时开工；协议含 V4 响应条款：V4 胜，验证
迁移；V4 负，定位差距层级；两种结果下 L2 都跑。

## 4. L3 · Agent Organization（10-23 后，v0.3 能力板）

**命题：组织化的 Agent 群体优于无组织的群体。**

v0.3 任务书草案已就绪（研究情报登记任务在充实），本文定三处
验收锚：

- 公告板：任务自选（过滤→评分→租约），评分含 A(topic) 注意力项
  + anti-crowding 项（Homogenization）；无中心分配者，共享设施
  即公告板（Agensh 背书）
- 双时标信任：高影响任务同行审查（MeshHeal 局部时标）+ 晋级链
  既有（全局时标）
- 资产版本化：八类资产统一生命周期，逐事件验收，净增量叙事禁入
  （AgentEvolver 反面教材）

**L3 验收（组织级对照）**：同一批真实任务，公告板组织 vs 随机
认领——分配质量、重复劳动率、**一条 approved 经验被第二个
Agent 复用（它没踩第一个 Agent 踩过的坑）**。经验跨成员复用是
本层招牌证据。

前置输入：L1 注册表 + L2 真实 approved 经验。

## 5. L4 · Multi-Agent Swarm（11 月末起）

**命题：带认知层的 Swarm 优于裸 Swarm。**

Morphogenesis 四通道接通（v1.1 §7 预留接口零改动）：Worker
认领前查认知状态（scope 冲突过滤 + 任务排序锚定注意力）、
Astrocyte 经验供体入 quarantine、Swarm 项目热度回流注意力池、
认知检索五操作对 Worker 开放。

边界纪律（Wu）：三道硬门——scope 权限、预算熔断、全审计——
不放权不跳门；租约承诺三态参考 Waggle 语义（维持/修改/撤销）。
CLM 对比定位进入对外叙事：内生上下文自治与外部不可篡改认知层
互补。

**L4 验收（群体级对照）**：同任务流，Swarm+认知 vs 裸 Swarm
——任务选择品味（与注意力分布的重合度）、经验复用率、
重复劳动率。

## 6. 时间表

| 时段 | 事件 |
|---|---|
| 10-14 ~ 10-17 | Works Relay D1–D3 + L1 session capture 双库管道 |
| 10-17 ~ 10-22 | Dogfood 真实接力；L2 经验管道积累 quarantine 候选；
V4 闭日独立运行 |
| 10-23 | V4 终局报告（独立线）；冻结解除 |
| 10-23 ~ 11 月初 | v0.2 四轨 + L2 管道并入正式库 + **L2 工作级实验**
（证据槽位二） |
| 11 月 | L3 能力板 + 组织级对照（证据槽位三） |
| 11 月末 ~ 12 月 | L4 四通道 + 群体级对照（证据槽位四） |

## 7. 红线

- V4 不作为任何层的前置条件；证据槽位一保持空位直到 V4 落地
- 冻结期内一切业务实现走 workbench 线，validation/V4/引擎/
  schema 零触碰；L2 实验与 L3/L4 待冻结解除
- 经验永不自动晋级；人工关口不可绕过
- 接入层不写引擎、不写认知状态——只产生事件
- 每层实验预注册、协议锁 SHA 后不可改；胜负都入档
- 不新增调度/证明系统；净增量叙事禁入验收

## 8. 文风纪律

存活标准：推进实现或传递信息。以下句式出现即删：
"为了稳妥起见"、"这在某些情况下可能会有问题"、"仅供参考"、
任何不改变下一步动作的免责声明。
