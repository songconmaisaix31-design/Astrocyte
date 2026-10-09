# 个人研究与创造工作台 — 开发时间规划与并行 TODO 清单

版本：v0.1　日期：2026-10-09　对应规格：Workbench_DDD_Spec_v0.1
用途：把规格第 19 节的切片 S0–S6 展开为可交给多个 Agent 并行执行的任务清单。

---

## 0 规划前提与并行原则

**依据规格：** 18.2（S0 由集成负责人先冻结契约，之后一队做领域/Service/存储、另一队做界面或独立适配器）、19.1（S0 先固定数据契约，随后并行）、D08（每切片界面+Service+存储+检查一起交付）。

**2026-10-09 用户确认：** 首个交付节点为 S0。主 Agent 负责最终 check 和提交；开发并发不设数量上限，可使用本地其他 Agent。各阶段日期和 Agent 数为原估算，不是已确认的交付承诺。并行边界是：

1. 一个任务 = 一条分支 / 一个 worktree = **一个不重叠写域**；
2. 跨上下文只通过被发布的 Service、版本化事件、ID+DTO 交互，不写对方数据表；
3. 公共文件（OpenAPI、公共 Go/TS 类型、迁移序列、依赖锁、程序入口）由**单一集成负责人**持有，其他任务申请变更。

满足以上三条后按任务需要安排多工作流，不为达到某个 Agent 数拆分任务。

**三类不可并行的串行门：**
- G1：S0 契约冻结（全部业务任务的前置）。
- G2：公共文件与跨上下文桥接的合并（任何时刻只有一人改）。
- G3：每个切片的合并收口与独立复核（由另一会话完成，不能用 mock 顶替）。

**通用纪律（来自规格）：**
- 任务固定五字段契约：`目标｜可改路径｜依赖｜验收｜停止点`。
- 状态仅用 `PASS / FAIL / NOT_RUN / BLOCKED`；mock 不能替代 AT09、AT14 与完整样例验收。
- 遇到范围、费用、权限或共享接口变化才停下来请求决定；否则先查现有资料。
- 禁止：`git add .`、reset/clean 清空、覆盖已有 AGENTS、移动旧项目、push/发布未经授权。

---

## 1 工作流划分（可并行的纵向轨道）

| ID | 工作流 | 主要写域 | 依赖 | 建议 Agent 数 |
|---|---|---|---|---|
| W0 | 集成与契约（集成负责人） | `cmd/server`、`contracts/openapi.yaml`、公共 DTO、`dependencies.lock.json`、根入口 | — | 1（串行，兼合并守门） |
| W1 | Attention 上下文 | `internal/attention/{domain,app}`、`internal/adapters/importers`、attention 表、资料沉淀页 | G1 | 2–3 |
| W2 | Workspace 上下文 | `internal/workspace/{domain,app}`、项目/会话/提案/批准 HTTP、workspace 表、共同工作区页 | G1 | 2 |
| W3 | Swarm 上下文 | `internal/swarm/{domain,app}`、mission/workitem/lease、swarm 表、蜂群执行页 | G1、P2 批准边界 | 2 |
| W4 | Judgment | `internal/judgment`、Jev 适配器、Laya worker、路由与缓存 | G1 | 1–2 |
| W5 | 仓库上下文 | RepositoryMap(AOCI)/CodeSearch/ImpactQuery(CodeGraph) 适配器、SemanticIndex | G1 | 1–2 |
| W6 | Agent 适配器与 MCP | `internal/adapters/agent`、`internal/adapters/mcp`、共享工具服务 | G1、P0 探测 | 1–2 |
| W7 | 存储与基础设施 | `internal/adapters/sqlite`、对象存储、作业队列、outbox、迁移 | G1 | 1–2 |
| W8 | 前端与设计系统 | `web/src`（导航/查询层/设计变量/三页） | G1、OpenAPI | 2（可按下沉/工作区/蜂群分页） |
| W9 | 质量与验收 | Playwright、AT 场景自动化、fixture、验收目录 | 各切片 | 1 |

---

## 2 阶段时间线与并发度

> **估计假设：** 1 迭代 = 1 周；Agent 辅助编码 + 每日人工复核；S0 的探测结论可能调整 S3/S4 的验收与工期。日期以 2026-10-12（周一）为起点，仅供参考。

| 阶段 | 覆盖切片 | 起止（估） | 时长 | 可安全并发工作流 | 建议峰值 Agent |
|---|---|---|---|---|---|
| P0 契约冻结与探测 | S0 | 10-12 → 10-23 | 2 周 | W0 串行门 + W1/W4/W5/W6/W8 探测并行 | 6–8 |
| P1 资料到候选 | S1 | 10-26 → 11-20 | 4 周 | W1 + W7 + W8 + W9 | 6–9 |
| P2 候选到批准 | S2 | 11-23 → 12-18 | 4 周 | W2 + W1 + W4 + W5 + W8 + W9 | 8–12 |
| P3 两 Agent 闭环 | S3 | 12-21 → 01-22 | 5 周（含缓冲） | W3 + W6 + W0 + W8 + W9 | 8–12 |
| P4 仓库上下文 + 判断委托 | S4 + S5 | 01-25 → 02-19 | 4 周 | W5 + W4 + W1 + W9 | 6–10 |
| P5 局部协作与继承 | S6 | 02-22 → 03-19 | 4 周 | W3 + W0 + W8 + W9 | 6–8 |
| P6 加固与迁移收口 | — | 03-22 → 04-09 | 3 周 | W7 + W0 + W9 | 4–6 |

**总计约 26 周（≈6 个月）**，其中 P3 为全规格技术密度最高的关键路径。压缩工期的唯一安全方式是削减 1.3 "首版不做"范围或接受更高集成风险，不能靠堆 Agent 压缩串行门。

---

## 3 逐阶段 TODO（含五字段任务契约）

### P0 契约冻结与探测（2 周）

- **P0-01 项目基座** ｜W0
  - 目标：独立 Go+TS 项目、两层包结构、可启动可构建
  - 可改路径：`cmd/server`、`internal/`（空包）、`web/`、`contracts/`、根配置
  - 依赖：无　｜　停止点：依赖未锁定或引入 Python 即停
  - 验收：`pnpm install --frozen-lockfile`、`pnpm dev`、`pnpm check`、`pnpm build` 全通过；包布局符合 §3.1
- **P0-02 OpenAPI 与契约冻结** ｜W0（G1）
  - 目标：固定 `/api/v1` 全部契约并生成 TS 客户端
  - 可改路径：`contracts/`、`web/src/api`
  - 验收：类型生成通过；前后端不各写一套相同协议类型
- **P0-03 SQLite 迁移与架构检查** ｜W7
  - 目标：`state.sqlite` 布局、按上下文前缀建表、依赖检查入口
  - 可改路径：`internal/adapters/sqlite`、`migrations/`
  - 验收：Domain 无 I/O、上下文无循环、传输层不访问 SQL
- **P0-04 三页 fixture** ｜W8
  - 目标：资料沉淀页/共同工作区页/蜂群执行页骨架 + fixture
  - 可改路径：`web/src`
  - 验收：1920×1080、1280×720、键盘路径；loading/empty/error/stale/disabled 齐备
- **P0-05 探测：Agent 原生能力矩阵** ｜W6
  - 目标：实测 Codex、Claude Code 八项能力（`discover/read_context/start/resume/send/stop/observe/reconcile`），填 §8.1 矩阵
  - 可改路径：`docs/`（探测报告），**先不改平台代码**
  - 验收：每个适配器附版本、原生会话协议证据、resume 可行性结论
- **P0-06 探测：无 Python 向量索引选型** ｜W5
  - 目标：验证 TS/ONNX 或 Go 侧 ANN 满足无 Python 且版本可追溯
  - 验收：选型报告 + 索引 generation/digest 方案
- **P0-07 探测：视频总结工具导出样本** ｜W1
  - 目标：取得真实导出样本，锁定格式/来源/时间戳字段
  - 验收：样本 + 字段映射；不足则先建 fixture，不冒充真实导入
- **P0-08 探测：AOCI/CodeGraph 版本与许可** ｜W5
  - 目标：固定版本、核对 FSL-1.1-MIT 分发条件、验证 worktree 隔离与增量更新
- **P0-09 探测：Laya ONNX 模型包** ｜W4
  - 目标：取得可加载 ONNX + tokenizer + config + revision + digest
  - 验收：断网推理成功；无资产则 Laya 置 `BLOCKED`，不改变无 Python 约束

### P1 资料到候选（4 周｜W1+W7+W8+W9）

- **P1-01 Attention Domain 聚合与状态机** ｜W1
  - 目标：Material/Distillation/Opportunity/TasteProfile 及值对象、§4.2 状态机
  - 可改路径：`internal/attention/domain`
  - 验收：4.2 状态转换、4.3 沉淀完成条件单测；Domain 无 I/O
- **P1-02 ImportMaterial 用例 + 导入器** ｜W1
  - 可改路径：`internal/attention/app`、`internal/adapters/importers`
  - 验收：source key + content digest 去重；同源内容变化生成 revision
- **P1-03 RecordDistillation 三层沉淀** ｜W1
  - 验收：内容/主题/项目三层；相同输入无新增问题不重复付费调用
- **P1-04 活跃度与四维排序** ｜W1
  - 验收：5.1 衰减公式区分 human/agent；5.2 信息不全用 unknown 不填 0
- **P1-05 资料沉淀页（真实 API）** ｜W8
  - 验收：候选卡显示四维+下一步+缺失依据；机器使用与人类关注分列
- **P1-06 作业队列与可见作业状态** ｜W7
  - 验收：17.4 持久化队列、并发上限、去重键、可取消
- **P1-07 验收：AT01/AT02/AT03/AT04** ｜W9
  - 验收：自动化可重复执行

### P2 候选到批准（4 周｜W2+W1+W4+W5+W8+W9）

- **P2-01 Workspace Domain 聚合** ｜W2
  - 目标：Project/Goal/Proposal/ApprovalGrant/SessionBinding/ContextPacket/ToolProfile
  - 可改路径：`internal/workspace/domain`
- **P2-02 项目注册与会话绑定** ｜W2
  - 目标：RegisterProject/BindSession、RepoRef/WorktreeRef/ProjectSnapshot（§6.2）
- **P2-03 提案与批准（两道门）** ｜W2
  - 目标：SubmitProposal/ApproveProposal；Grant、幂等回执、outbox 同事务（§12.5）
  - 验收：模型无批准入口；version_conflict 正确处理
- **P2-04 撤销与上下文包** ｜W2
  - 目标：RevokeApproval/BuildContextPacket；撤销后阻止新动作
- **P2-05 跨上下文桥接层** ｜W0（G2）
  - 目标：消费 `opportunity_admitted` 幂等形成草稿 Proposal；不创建执行目录、不拉起 Agent
- **P2-06 共同工作区页：批准面板** ｜W8
  - 验收：集中展示范围/成本调用限制/产出/停止条件；费用 unknown 明确标注
- **P2-07 Attention 消费 `artifact_accepted`（占位）** ｜W1
- **P2-08 验收：AT05/AT06/AT07/AT12** ｜W9
- **P2-09 适配准备（不改平台）** ｜W4/W5
  - 目标：Jev Go HTTP 适配契约确认、CodeGraph 增量更新与隔离验证；不得绕过 S2 批准边界

### P3 两 Agent 工作闭环（5 周｜关键路径）

- **P3-01 Swarm Domain 聚合与状态机** ｜W3
  - 目标：Mission/WorkItem/Lease/Attempt/Artifact/Experience（§10.1）
- **P3-02 CreateMission** ｜W3
  - 目标：消费 `proposal_approved`，再查有效 Grant，按 grant_id 幂等建 Mission
  - 验收：重复投递不重复启动；撤销后拒绝启动
- **P3-03 ClaimWorkItem** ｜W3
  - 目标：CAS + 单调 fence + 预算预留同事务；失败不留半份预留
  - 验收：AT08 两 Agent 同时认领仅一个有效租约
- **P3-04 SubmitArtifact/AcceptArtifact** ｜W3
  - 目标：核对租约/Attempt/写域/快照/digest；失败转 revision_needed/blocked
  - 验收：提交≠通过；仅验收命令转 accepted
- **P3-05 上下文交付与两种接续** ｜W6
  - 目标：原会话续（验证原生 ID/项目/占用/版本）+ 显式交接（不可变 ContextPacket）
  - 验收：AT09 分别保留原生身份或 successor 链；不支持时显示模式已改变
- **P3-06 共享 MCP 工具服务** ｜W0/W6
  - 目标：`attention_search/workspace_context/code_search/code_impact/swarm_*/experience_search`；批准/提权不暴露为 Agent 工具
- **P3-07 蜂群执行页** ｜W8
  - 验收：任务时间线+工作项详情+暂停/取消/diff/采用/退回入口
- **P3-08 首个真实样例端到端** ｜W9（§19.2）
  - 目标：一篇论文 + 一份视频总结 + 一个可改项目 + 一个小型改进
  - 验收：可用产物、相关检查、来源链、一次交接、一次采用反馈
- **P3-09 验收：AT08/AT09/AT13** ｜W9

### P4 仓库上下文 + 判断委托（4 周｜W5+W4+W1+W9）

- **P4-01 RepositoryMap 适配器（AOCI）** ｜W5
- **P4-02 CodeSearch/ImpactQuery 适配器（CodeGraph）+ SemanticIndex** ｜W5
- **P4-03 索引 generation 与快照校验** ｜W5
  - 验收：`index_stale` 正确返回；两 worktree 不串结果
- **P4-04 修改前/修改后影响证据** ｜W5
  - 验收：未索引动态调用或缺失测试用例时输出缺口，不产出"没有影响"
- **P4-05 Judgment 契约与路由** ｜W4
  - 目标：JudgmentRequest/Result、路由顺序、缓存键（§11.2/11.3）
- **P4-06 Jev Go HTTP 适配器** ｜W4
- **P4-07 Laya TS/ONNX 常驻 worker** ｜W4
  - 验收：断网推理与重启可用
- **P4-08 逐题型委托 + 反馈回流** ｜W1/W4
  - 验收：AT11 拒答/截断/外发未授权不自动准入，不静默切云端
- **P4-09 验收：AT10/AT11/AT12 + S4/S5 出口** ｜W9

### P5 局部协作与继承（4 周｜W3+W0+W8+W9）

- **P5-01 局部候选选择 + LocalPolicy** ｜W3
- **P5-02 并行工作项与写域串行化** ｜W3
  - 验收：默认最多两个活跃执行者；写域重叠串行或人工调整
- **P5-03 无进展停止与预算状态** ｜W3
  - 验收：预算耗尽/无进展产生明确状态与原因；订阅制费用记 unknown 不记 0
- **P5-04 Experience 继承检索** ｜W3
  - 验收：AT16 第二任务实际引用第一次成果并记录适用性
- **P5-05 拓扑视图** ｜W8
  - 验收：节点/边对应真实执行者与事件，不绘制无事件支撑的装饰
- **P5-06 验收：AT14/AT15/AT16 + S6 出口** ｜W9

### P6 加固与迁移收口（3 周｜W7+W0+W9）

- **P6-01 重启恢复与对账** ｜W7
  - 目标：未完成作业/租约/operation 恢复；`delivery_unknown` 先对账；过期租约递增 fence
  - 验收：AT13、AT15
- **P6-02 无 Python 全链路验证** ｜W0/W9
  - 验收：AT14 默认构建、启动、导入、判断路径可运行
- **P6-03 迁移一致性与失败停止** ｜W7
  - 验收：迁移前备份 + schema 版本；失败停止启动，不边运行边用半升级表
- **P6-04 全量 Playwright + 截图 + 日志** ｜W9
  - 验收：§16.5 关键页面截图与功能验证放同一验收目录
- **P6-05 依赖锁与许可证清单** ｜W0
  - 验收：`dependencies.lock.json` 定版；AOCI/Laya/Jev 许可随版本记录
- **P6-06 文档收口** ｜W0
  - 目标：`AGENTS.md`/`STATUS.md`/`docs/SPEC.md`/`docs/TASKS.md` 单一来源

---

## 4 依赖图与关键路径

```
P0-01 基座 → P0-02 契约冻结(G1) ─┬→ P1-01..04 (Attention) → P1-05/07
                                 ├→ P0-03 存储 → (各上下文 Repository/表)
                                 ├→ P0-04 三页 fixture → W8 各页
                                 └→ W2-01 Workspace Domain → P2-03 提案批准
P0-05 能力探测 ─────────────────────────→ P3-05 两种接续
P2-03 批准 → P2-05 桥接(G2) → P3-02 CreateMission → P3-03 Claim → P3-04 Submit
                                              → P3-05 接续 → P3-06 MCP → P3-08 真实样例
P2-03 批准边界 ─────────────────────────→ W5/W4 适配器不得绕过
P3 基础账本 → P4/P5 扩展
```

**关键路径（决定总工期）：**
`P0-01 → P0-02 → P1-01 → P2-01 → P2-03 → P2-05 → P3-01 → P3-03 → P3-04 → P3-05 → P3-06 → P3-08 → P3-09`
其余工作流均为可并行支线，只要写域不重叠。

---

## 5 并行并发度与调度说明

- **同写域内严禁并行**：如 `contracts/openapi.yaml`、`dependencies.lock.json`、迁移序列、公共 Go/TS 类型——任何时刻只有 W0 一人改；其他任务走变更申请。
- **跨工作流并行安全条件**：只经 Service/版本化事件交互，不写对方表（§2.3、§3.2）。
- **前端可再分**：W8 可拆成"资料沉淀页 / 共同工作区页 / 蜂群执行页"三个 Agent，共享设计变量与查询层由一人持有。
- **复核隔离**：每个切片收口由**另一会话**复核（G3），复核 Agent 不计入实现并发。
- **回滚保护**：新平台从独立目录/新 worktree 开始，保留旧 Morphogenesis；同一执行任务不能由新旧两个账本控制。

---

## 6 风险缓冲与触发条件

| 风险 | 触发 | 应对 / 缓冲 |
|---|---|---|
| Agent 原生接续（resume）能力不足 | P0-05 探测结论 | 降级为仅显式交接；调整 S3/AT09 验收；**加 1 周缓冲** |
| Jev 地区不可用 | P4-06 接入时 | 本地 Laya / 开源复现为默认后端；云端 Jev 仅显式授权启用 |
| 无 Python 向量选型受限 | P0-06 结论 | SemanticIndex 端口兜底；S4 前置 spike；不引入 Python |
| Laya 模型包缺失 | P0-09 | Laya 置 BLOCKED，其余模块继续；不用随机回答替代验收 |
| S3 集成风险 | P3 中段 | 预留 1 周缓冲；AT08/AT09 尽早自动化 |

---

## 7 里程碑验收映射（S0–S6 ↔ AT 场景）

| 阶段 | 切片出口 | 对应 AT |
|---|---|---|
| P0 | 可启动、可构建、接口类型生成与依赖检查通过 | — |
| P1 | 收藏不启动任务；重复导入与继续沉淀可操作 | AT01/02/03/04 |
| P2 | 模型无批准入口；版本冲突与撤销正确处理 | AT05/06/07/12 |
| P3 | 一个真实改进完成，结果与来源可回查 | AT08/09/13 |
| P4 | 两 worktree 不串结果，过期索引可恢复；本地模型断网可用 | AT10/11/12 |
| P5 | 第二次任务实际引用第一次成果且记录适用性 | AT16 |
| P6 | 记录一致，旧权限不能执行新动作；无 Python 全链路 | AT14/15 |

---

## 8 Agent 任务交接模板

每个任务下发时使用固定格式（直接可粘贴给 Agent）：

```text
任务ID: <P?-??>
目标: <一句话可验收的目标>
可改路径: <精确文件/目录白名单>
依赖: <前置任务ID / 已冻结契约 / 已发布 Service>
验收: <可执行的命令或可观察的结果>
停止点: <遇到范围/费用/权限/共享接口变化即停，不得越界>
状态: <初始 NOT_RUN>
记录: SOURCE_SHA / 环境 / 实际命令 / 结果
```

**交接注意：**
- 只给当前任务需要的工具与资料，不广播全部工具定义（§8.4）。
- 未安装真实适配器时保留 `NOT_RUN / BLOCKED`，不用 mock 顶替（§18.5）。
- 失败写下一条可执行处理动作，不扩写自我说明。

---

## 9 看板字段建议（用于多 Agent 跟踪）

`任务ID | 阶段 | 工作流 | 目标 | 可改路径 | 依赖 | 验收 | 停止点 | 负责Agent | 分支/worktree | 预估 | 起止 | 并行组 | 状态(PASS/FAIL/NOT_RUN/BLOCKED)`
