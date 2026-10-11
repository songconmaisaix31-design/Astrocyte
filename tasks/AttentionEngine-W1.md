# AttentionEngine-W1 · 事件流与归一化（M1）

事实源：任务书 `tabbit-code (2).md` §4.1/§4.5/§5.2/§8 M1；`tasks/AttentionEngine-plan.md` W1 行。
分支：`songconmaisaix31-design/attention-engine-w1-events`，基线 `dab8fea`。

## 状态（2026-10-11）

波 A 完成：只读事件源勘察 + 无损映射表 + Event→Signal 归一化设计 + §5.2 权重表 +
Claude Code/Codex 会话回放格式勘察。编码阻塞于 M0 契约类型未发布（集成分支
`attention-engine/v1-20261011` 当前仅文档 `0e14c7d`，W0 分支未 push）。主控发
merge_ready 后执行 `git fetch origin && git merge origin/attention-engine/v1-20261011`
再开码（波 B）。

### 阶段收束（partial，2026-10-11）

主控通知用户将切换 `codex --yolo`，本 Task/Dispatch 按当前真实状态收束并 settle，
不再等待 M0 开展新阶段。

- 已完成（真实）：六类事件源只读勘察、§二无损映射表、§三归一化设计 + §5.2 权重表、
  Claude Code/Codex JSONL 格式勘察、§五验收脚手架设计、M0 契约 Handoff 需求（§四）。
- 未完成（真实 partial）：M1 编码未开工（Event→Signal 归一化、权重表落码、event_store
  持久化、真实会话回放），全部阻塞于 M0 契约类型（Event/Signal/迁移/引擎接口）未发布，
  无任何一行 M1 业务代码产出，仅文档。
- 下一步（接续入口）：M0 契约合入 `attention-engine/v1-20261011` 后，沿本分支/worktree/
  写域新建 Task/Dispatch，执行 `git fetch origin && git merge origin/attention-engine/v1-20261011`，
  按 §三/§五 开码：`internal/attention/domain/{event,signal}*.go`、
  `internal/attention/app/{event,signal}*.go`、`internal/adapters/events/**`、
  `internal/adapters/sqlite/event_store*.go`、`tests/attention-engine/M1*`。

## 目标与验收（任务书原文，不降级）

- §4.1 Event（source/session/ts/kind/tokens/scope/topic/raw）作为归一化目标；现有
  watcher 事件无损映射（`raw` 保留原始载荷，不丢任何事件，计数可对账）。
- §5.2 权重表硬编码（不在线学习，不做自适应调参）。
- 归一化：Event → Signal（信号强度 s_i + 权重 w_i + 信号类型）。
- 验收：真实 Claude Code/Codex 会话回放 → Event/Signal；权重标注抽检正确率 +
  无静默丢失（映射前后计数对比）。

## 一、事件源勘察结果（只读）

现有 s1 产出六类，均为 SQLite 已持久化的既有数据；本轨只经适配器消费，不改 s1。

### 1.1 attention_outbox（领域事件，最完整的事件源）

表 `attention_outbox`（migration 003）。行字段：`id, type, aggregate_id,
aggregate_version, occurred_at, correlation_id, causation_id, payload(json)`。

`type` = `"attention." + kind`，`payload` 为事件专属 JSON。全部 24 种 type 已枚举
（见 §二映射表）。写入口 `internal/attention/app/attention.go:213` `Service.event`，
事务内与聚合/回执同写（失败回滚），`occurred_at` 由注入 Clock 生成。这是唯一自带
"每事务一条、可计数对账"语义的源。

### 1.2 material Usage（人类/Agent 使用记录）

存于 `attention_materials.data` 的 `MaterialDetail.Uses []Usage`（`internal/attention/app/contracts.go:94`）。
`Usage{id, material_id, actor_id, actor_kind(human|agent), action, occurred_at}`。
action 词表：`read/reread/annotate/adopt/classify/mention/project_reuse`（人类，
`spaces.go:253`）+ `read`（Agent，`materials.go:14`）。同时 material 上维护
`HumanUsageCount/AgentUsageCount` 计数器。Usage 在 material 聚合内追加，不单独出 outbox 行，
但 `material_used` 事件会同步入 outbox（`spaces.go:262`、`materials.go:20`）。

### 1.3 attention_source_items（收藏条目）

表 `attention_source_items`（migration 006）：`source_id, external_id, revision,
data(json)`。`data` 为 `SourceItem`（`contracts.go:582`）：`source_id, external_id,
revision, metadata(ListingMetadata), stale, selected, import_job_id, material_id,
recommendation`。选中条目（`selected=true`）即"收藏入 Inbox"，对应 kind=`collect`。

### 1.4 local_agent_sessions（原生会话）

表 `local_agent_sessions`（migration 007）：`id, project_id, data(json)`。
`data` 为 `domain.NativeSession`（`internal/workspace/domain/projects.go:109`）：
`id, project_id, cli(codex|pi|claude|opencode), native_id, version, mode,
status, stop_confirmed, context_packet, operations(map[string]NativeOperation),
limitations, updated_at`。`Operations` 内每条 `NativeOperation{action, status,
created_at, finished_at}` 是一次原生命令（start/send/stop 等）的持久化回执。

### 1.5 local_project_discovery（项目活动）

表 `local_project_discovery`（migration 008）：单行 `id='registered', data(json)`。
`data` 为 `ProjectDiscoverySnapshot`（`domain/registered_projects.go:43`），含
`Projects []RegisteredProject`，每个带 `Git{status, branch, head, last_commit_at,
dirty}` 与 `Activity{status, source, last_activity_at}`，及 `Contributors
[]ProjectContributor{cli, source, root, session_id, observed_at, created_at,
activity_at}`。这是 commit（git last_commit/head）与原生会话头关联的活动快照。

### 1.6 原生会话元数据（Claude Code / Codex 会话记录格式，只读勘察）

真实会话存于本机 JSONL，只读头部/有界样本，不读凭据、不整盘扫描。格式已由
`internal/adapters/agents/{project_metadata,history}.go` 证实：

| CLI | 目录 | 首记录（会话头） | 后续记录 | token/工具信号 |
|---|---|---|---|---|
| Claude Code | `~/.claude/projects/<encoded-cwd>/<sessionId>.jsonl` | `{type:"user"\|"assistant"\|"system", sessionId, cwd, timestamp}` | `type=user/assistant/system`，assistant 带 `message.content`（含 `tool_use` 块：Edit/Bash/Grep/Read/WebSearch/WebFetch）与 `message.usage{input_tokens,output_tokens}` | usage 字段 + tool_use.name |
| Codex | `~/.codex/sessions/<YYYY>/<sessionId>.jsonl` | `{type:"session_meta", payload:{id,cwd,timestamp}}` | `{type:"response_item", payload:{type:"message"\|"function_call"\|"function_call_output", role, content:[{type:"input_text"\|"output_text",text}]}}` | payload 内 tool 调用（shell/edit/…） |

回放只允许经人类注册的 history root（`ProjectSettings.HistoryRoots[cli]`）读取，
沿用 `discoverHistory`/`readHistory` 的 64KiB 头/1MiB 全文有界与 symlink 拒绝约束，
不扫描 `~/.claude` 全盘。Pi/opencode 已存在适配器但 M1 回放范围只覆盖 Claude Code/Codex。

## 二、无损映射表（现有 watcher 事件 → 新 Event）

映射原则：每条现有事件生成一条新 `Event`，不丢、不合并、不重排；`raw` 完整保留
原始记录（含 `type/aggregate_id/aggregate_version/correlation_id/causation_id/payload`
或原生 JSONL 原文），保证审计可回溯。类型化字段是语义投影。

- `source` 派生规则：原生会话回放 → `claude_code`/`codex`；outbox 中 `actor_kind=agent`
  的 `material_used` → `mcp`；其余 s1 面板/人类行为 → `manual`。
- `ts` = 源时间戳（outbox `occurred_at`、Usage `occurred_at`、会话 `timestamp`、
  git `last_commit_at`、贡献者 `activity_at/created_at`）转 Unix 秒。
- `scope` = 项目/路径（NativeSession `project_id`、RegisteredProject `root`、
  outbox `aggregate_id` 或 payload 内 project 字段）。
- `topic` = 空串，由 Attention Graph / sleep 聚类后回填（§4.1 新增字段）。
- `tokens` = 源 payload 中可得的 token 数（Claude/Codex usage 字段），否则 0。
- `session` = 会话标识（NativeSession `native_id`、原生 JSONL `sessionId/id`、
  贡献者 `session_id`）；outbox 无会话语义时 = `correlation_id`（请求关联）。

### 2.1 attention_outbox（source=manual/mcp）

| 现有 type | Event.kind | §5.2 信号 | w |
|---|---|---|---|
| material_imported | collect | 隐式轻度 | 0.3 |
| material_used · human · read/reread | read_paper / watch_video（按 material.kind） | 隐式深度 | 1.0 |
| material_used · agent · read | read_paper / watch_video（source=mcp） | 隐式深度 | 1.0 |
| material_used · human · annotate/adopt/classify/mention/project_reuse | user_approve | 显式正向 | 1.5 |
| material_updated · pinned=true | user_approve | 显式正向 | 1.5 |
| material_updated · lifecycle=withdrawn | user_veto | 显式负向 | -0.8 |
| job_queued / job_retried / job_cancelled / job_recovered | agent_task（运行中，无信号） | — | 0 |
| job_succeeded | agent_task（成功） | 环境反馈 | 0.8 |
| job_failed | agent_task（失败） | 环境反馈 | -0.1（只降 confidence） |
| distillation_recorded | agent_task（成功） | 环境反馈 | 0.8 |
| opportunity_created / opportunity_revised | agent_task（运行中） | — | 0 |
| opportunity_reviewed · adopt | user_approve | 显式正向 | 1.5 |
| opportunity_reviewed · reject/revise/later | user_veto | 显式负向 | -0.8 |
| ranking_profile_updated | user_approve | 显式正向 | 1.5 |
| material_classified | user_approve | 显式正向 | 1.5 |
| project_space_referenced | collect | 隐式轻度 | 0.3 |
| project_space_reference_removed | user_veto | 显式负向 | -0.8 |
| tracking_source_bound | collect | 隐式轻度 | 0.3 |
| source_items_selected | collect | 隐式轻度 | 0.3 |
| source_sync_progress | agent_task（运行中） | — | 0 |
| source_catalog_discovered | collect | 隐式轻度 | 0.3 |
| domain_created / domain_revised / project_space_created | collect（结构事件，w=0 不产生 Signal） | — | 0 |

### 2.2 material Usage（source=manual/mcp，kind 同上按 action/material.kind）

Usage 已由 `material_used` outbox 覆盖；独立枚举保证"Usage 表内记录"也逐条映射，
避免仅依赖 outbox 造成静默丢失。计数对账以 outbox + Usage 双源合计。

### 2.3 attention_source_items（source=manual，kind=collect，w=0.3）

`selected=true` 的 `SourceItem` → `collect`；`recommendation` 结果（status=ready/deferred）
保留在 raw，不单独产生信号。

### 2.4 local_agent_sessions → NativeSession.Operations（source=claude_code/codex/…）

每条 `NativeOperation` → `agent_task`（action=start/send/stop，status 定成败：
accepted→成功 0.8，failed/unknown→失败 -0.1 只降 confidence）。`mode=native`、
`stop_confirmed`、`context_packet` 入 raw。

### 2.5 local_project_discovery（source=manual）

| 现有观测 | Event.kind | §5.2 信号 | w |
|---|---|---|---|
| Git.last_commit_at / head（有提交活动） | commit | 隐式常规 | 0.6 |
| Activity.last_activity_at（来源=session） | agent_task | — | 0 |
| Contributors 每条（source=native_session_header） | agent_task（会话头关联） | — | 0 |

### 2.6 原生会话回放（Claude Code / Codex JSONL，source=claude_code/codex）

| 原生记录 | Event.kind | §5.2 信号 | w |
|---|---|---|---|
| tool_use.name ∈ {Edit, Write, MultiEdit, Bash(编辑类)} | edit_file | 隐式常规 | 0.6 |
| tool_use.name ∈ {Grep, Glob, WebSearch, WebFetch, Task(搜索)} | search | 隐式常规 | 0.6 |
| tool_use.name ∈ {Read, ReadPaper, WebFetch(论文)} | read_paper | 隐式深度 | 1.0 |
| Bash 内含 git commit | commit | 隐式常规 | 0.6 |
| 视频/媒体相关 Read/Open | watch_video | 隐式深度 | 1.0 |
| 会话头（session_meta / sessionId） | agent_task（会话建立，无信号） | — | 0 |
| assistant 有 usage 字段 | （tokens 计入 Event.tokens） | — | — |

`kind` 由工具名 + 内容启发式判定，判定规则确定性、可复算；无法归类的工具调用保留
`raw` 且 kind 回退 `agent_task`（不丢事件）。`tokens` 取 assistant `usage` 累计。

## 三、Event→Signal 归一化 + §5.2 权重表（M1 编码目标）

```go
// Signal：归一化产物，W2 引擎消费（形状待 M0 契约锁定，此处为设计示意）
type Signal struct {
    EventID    string
    Source     string   // claude_code | codex | manual | mcp
    Topic      string
    SignalType string   // implicit_deep | implicit_regular | implicit_light |
                        // explicit_positive | explicit_negative | environment
    Strength   float64  // s_i，初值每事件=1（tokens 暂不缩放，待 Q2）
    Weight     float64  // w_i，§5.2 硬编码
    Ts         int64
}
```

§5.2 权重表硬编码（`internal/attention/domain/signal*.go`，常量，无在线学习）：

| SignalType | §5.2 原文 | w |
|---|---|---|
| implicit_deep | 视频暂停/回放/倍速、长时阅读 | 1.0 |
| implicit_regular | 编辑文件、commit、搜索 | 0.6 |
| implicit_light | 收藏、加入 Inbox | 0.3 |
| explicit_positive | 用户标记重要 | 1.5 |
| explicit_negative | 用户否决/打回 | -0.8 |
| environment_success | Agent 任务成功闭环 | 0.8 |
| environment_failure | Agent 任务失败 | -0.1（只降 confidence，不重罚 attention） |

归一化规则（确定性）：`Event.kind → SignalType` 查表；`environment_success/failure`
由 outbox `job_succeeded`/`job_failed` 或 `NativeOperation.status` 判定；失败信号
w=-0.1 只进 confidence 路径（W2 消费，本轨只标注），不进 A(t) 求和。

## 四、M0 契约依赖（Handoff 需求清单，不冻结形状）

W1 编码依赖 W0 锁定，全部 Handoff 不在本轨定义：

1. `Event`（§4.1）：source/session/ts/kind/tokens/scope/topic/raw——本轨归一化目标。
2. `Signal` + 权重表常量（§5.2）——本轨实现归一化，W2 消费；Signal 形状需 W0 或
   本轨与 W2 对齐（当前 W2 报告标注"Signal+权重表为 M1 产物，W2 只消费"）。
3. 存储：`event_store` 持久化 Event/Signal 行，依赖 W0 的 sqlite 连接/迁移基础设施
   （`migrations/014+`）。
4. 引擎接口：W0 "引擎接口定义"，本轨只调用，不自行定义对外接口。

Handoff 需求：① Event 最终字段/JSON tag；② Signal 归属（W1 domain vs W0 contract
统一）；③ event_store 表结构（W0 迁移所有者的建表字段）；④ Clock 接口与随机种子注入
约定（供 golden 测试）。

## 五、验收测试设计（脚手架）

测试文件：`tests/attention-engine/M1_*_test.go` + `internal/attention/domain/signal*_test.go`。

1. 无损映射：固定 outbox/Usage/SourceItem/NativeSession/JSONL fixture 全量 → 映射后
   `len(events)` == 源记录数；`raw` 与源逐字节一致（计数对账 = 无静默丢失）。
2. 权重标注：fixture 已知 kind/status → 断言 `SignalType` 与 `w` 逐条等于 §5.2 表
   （抽检正确率 = 全对）。
3. 真实回放：取人类注册 history root 下一条真实 Claude Code JSONL + 一条 Codex JSONL，
   回放为 Event 序列；断言 tool_use→kind 判定与 usage→tokens 非负、事件数 > 0。
4. 失败信号特例：`job_failed`/`status=failed` → w=-0.1 且只标 confidence 路径。

## 六、独占写域（本轨可改，其余只读）

- `internal/attention/domain/{event,signal}*.go`
- `internal/attention/app/{event,signal}*.go`
- `internal/adapters/events/**`（新包，消费现有 s1 产出）
- `internal/adapters/sqlite/event_store*.go`
- `tests/attention-engine/M1*`
- `tasks/AttentionEngine-W1.md`

## 七、开放问题（待用户/主控答复，不据此冻结接口）

- Q1：Signal 类型归属——放 `internal/attention/domain`（本轨）还是 W0 契约文件统一
  定义？W2 已按"M1 产物"标注，需 W0 确认。
- Q2：信号强度 s_i 取值——v0.1 假设每事件=1，还是按 tokens 缩放（W2 同 Q2）。
- Q3：原生回放真实数据来源——M1 验收需要人类注册至少一个含 `HistoryRoots[codex]` 与
  `HistoryRoots[claude]` 的项目，还是允许有界 fixture + 一条真实样本混合。
- Q4：`watch_video`（隐式深度）在现有六类源中无直接来源（视频暂停/倍速不在 s1 产出），
  v0.1 是否仅由原生回放/未来 W7 事件源供给。

## 八、假设记录（开工即记，不冻结）

- A1：outbox `material_used` 与 Usage 表双源计数，去重以 outbox 为准、Usage 兜底，
  对账目标"任一源记录都有对应 Event"。
- A2：结构事件（domain_created/project_space_created/ranking_profile_updated）映射
  kind 但 w=0，不产生 Signal，仅 Event 无损入账。
- A3：`source` 四值中 `manual` 覆盖 s1 面板人类行为；Agent 凭证读 = `mcp`；原生会话
  = `claude_code`/`codex`。
- A4：kind 由工具名/内容启发式判定为确定性规则（不做 LLM 判定）；未命中回退
  `agent_task`，事件不丢。
- A5：s_i = 1（每信号事件贡献一次权重），tokens 暂不缩放（Q2 待答）。
