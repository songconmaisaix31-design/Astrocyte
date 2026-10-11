# AttentionEngine W5 · Experience Memory（M5 经验侧）

- 轨道分支：`songconmaisaix31-design/attention-engine-w5-experience`
- 事实源：任务书 `tabbit-code (2).md` §3/§4.3/§4.5/§7/§8 M5；主计划 `tasks/AttentionEngine-plan.md` W5 行
- 独占写域：`internal/attention/domain/experience*.go`、`internal/attention/app/experience_*.go`、`internal/adapters/sqlite/experience_store*.go`、`tests/attention-engine/M5-experience*`、本文件
- 状态：Wave A 阶段已收束（partial）——勘察 + 里程碑设计 + 纯本轨经验模式规则已实现并单测通过；app/service/store 与验收测试未实现，待 M0 契约 SHA 发布后在下一阶段（codex --yolo）继续。

## 1. 目标（§4.3 Experience）

Experience 是「Agent 做事经验」，不是知识。模型：

```
Experience {
  id        string
  pattern   string          // 如 "centralized scheduler"
  verdict   avoid | prefer | condition
  condition string          // 触发条件，如 ">100 agents"；verdict=condition 时必有
  evidence  []EventRef      // 失败/成功的原始轨迹引用（EventRef = {event_id, weight_at_use, ts}，§4.5）
  status    quarantine | approved
}
```

职责边界（本轨只做经验侧）：

1. 失败轨迹自动捕获为候选：`agent_task` 失败事件 → 生成 `status=quarantine` 的 Experience 候选，`evidence` 指向失败原始轨迹。
2. `quarantine → approved` 晋级链：只经人工关口（HTTP 端点，W0 承载），不暴露为 Agent 工具（W6 不提供 approve）。
3. 默认项目隔离 + 人工显式跨项目分享（带审计：谁、何时、分享到哪）。
4. 失败信号只降 confidence、不重罚 attention：这是 M1/M2 的权重表语义（§5.2 `Agent 任务失败 w=-0.1`），本轨不实现 attention 更新，只负责把失败轨迹沉淀为候选经验供后续检索。

## 2. Wave A 勘察结论（Wave B 复用现有模式，不新建框架）

### 2.1 领域层 `internal/attention/domain/`
- 纯值 + 纯函数规则，无 I/O、无时钟依赖（见 `rules.go`、`job.go`）。
- 惯例：枚举以字符串常量/切片表达，非法值返回 `domain.ErrInvalid`；用 `slices.Contains` 校验；验证函数返回 `error`。

### 2.2 应用层 `internal/attention/app/`
- 契约类型集中在 `contracts.go`（带 `json` tag 的 DTO + 命令结构）。
- 服务走 `AttentionService` 接口 + `Repository`/`AttentionTx` 回调事务。
- 事件：`OutboxEvent{ID,Type,AggregateID,AggregateVersion,OccurredAt,CorrelationID,CausationID,Payload}`，经 `AttentionTx.AppendEvent` 落 `attention_outbox` 表。
- 已有失败事件 `attention.job_failed`，payload `{job_id, operation_id, delivery_unknown, attempts, error, cause, cause_detail}`（`jobs.go:592`）。这是「agent_task 失败轨迹」的现成原始证据。

### 2.3 存储层 `internal/adapters/sqlite/`
- `AttentionRepository` + `attentionTx` 实现 `app.AttentionTx`；`BEGIN IMMEDIATE` 串行化；JSON-in-TEXT 列；`saveHead` 做乐观版本写入；`listJSON`/`load` 通用读写。
- 迁移：`migrations/NNN_*.sql` 内 `INSERT OR IGNORE INTO _migrations(id)` + `foundation_schema` 版本更新；`embed.go` 自动嵌入。`014+` 归 W0。

### 2.4 项目隔离键
- 项目标识来自 `internal/workspace/domain` 的 `LocalProject.ID`/`ProjectID`；应用侧传输身份为 `app.Principal{ID,Kind}`（由受信传输注入，不从命令体解码）。
- Experience 的 scope 用 `project_id` 表达，默认隔离、分享后可见。

## 3. M5 经验侧设计

### 3.1 领域规则（Wave B 落 `domain/experience.go`）
- 枚举：`verdict ∈ {avoid, prefer, condition}`；`status ∈ {quarantine, approved}`。
- 校验：`pattern` 非空；`verdict=condition` 时 `condition` 非空；`evidence` 非空（候选必须有原始轨迹，§4.5 证据链必须可回溯）。
- 晋级规则：`quarantine → approved` 只能由人工 approve 触发；`approved → quarantine` 允许人工降回；其余迁移非法。
- 项目作用域规则：读取/检索只在 `project_id` 匹配（或已分享）时可见。

### 3.2 应用操作（Wave B 落 `app/experience_*.go`，服务接口挂 `app` 契约）
- `CaptureFailure(ctx, Principal, Event)`：`agent_task` 失败事件 → 生成 quarantine 候选（含 evidence 指向失败 Event）。
- `ApproveExperience(ctx, Principal, id)`：人工关口，quarantine → approved（Principal.Kind 必须为 human）。
- `ShareExperience(ctx, Principal, id, targetProjectID)`：显式跨项目分享，写审计记录（who/when/where）。
- `ListExperiences(ctx, Principal, projectID)`：按项目作用域过滤（默认隔离；分享后可见）。
- 检索入口 `get_experiences(pattern)`（§7）归 W6 MCP；本轨提供按 pattern 查询的只读方法供 W6 调用。

### 3.3 存储（Wave B 落 `adapters/sqlite/experience_store*.go` + W0 的 `014+` 迁移）
- 表：`attention_experiences`（聚合头：id、project_id、pattern、verdict、condition、status、version、data）与审计表 `attention_experience_shares`（experience_id、shared_to_project、shared_by、shared_at）。
- evidence 存 `[]EventRef` JSON，引用 `attention_outbox`/新 Event 表的事件 id（依赖 W0/W1 事件落库，跨轨 Handoff）。

## 4. M0 契约依赖（待 W0 发布后 fetch 集成分支再编码）

1. `Experience` schema 类型（W0 的 `*_contract*.go`，含 id/pattern/verdict/condition/evidence/status）。
2. `EventRef`/`EvidenceRef` 类型（§4.5 `{event_id, weight_at_use, ts}`）。
3. `Event` 类型（§4.1，含 `kind=agent_task`）与事件落库表/读取接口。
4. 引擎接口（若失败事件经引擎侧下发，需确认消费入口；否则经 W1 归一化后的 Signal/Event 流）。

## 5. 跨轨 Handoff

- 向 W1：本轨消费「agent_task 失败事件」的归一化产出；请 W1 保证失败事件保留 `job_id/operation_id/attempts/error/cause` 原始轨迹字段，供 evidence 回溯。
- 向 W0：Experience 契约类型、`014+` 迁移、HTTP 人工关口端点（approve/share 走 HTTP，Principal.Kind=human）。
- 向 W6：MCP 只读暴露 `get_experiences`；**不得**暴露 approve/share（批准/提权只经人工关口）。

## 6. 验收标准（M5 经验侧）

1. 失败轨迹自动捕获为候选：给定 `agent_task` 失败事件，生成一条 quarantine Experience 候选，evidence 指向该失败事件原始轨迹，可回溯。
2. 晋级链：quarantine 只能经人工 approve 转 approved；无任何 Agent 工具可触发 approve。
3. 项目作用域过滤正确：默认隔离（项目 A 的 Experience 对项目 B 不可见）；人工显式分享后目标项目可见；分享审计记录 who/when/where。

## 7. 测试计划（Wave B 落 `tests/attention-engine/M5-experience*`）

- `M5-experience_capture_test.go`：失败事件 → 候选捕获 + evidence 回溯断言。
- `M5-experience_promotion_test.go`：quarantine→approved 人工关口；Agent 无法 approve；非法迁移拒绝。
- `M5-experience_scope_test.go`：默认隔离 + 分享后可见 + 分享审计字段断言。
- 依赖 W0 共享 harness 与 M0 契约类型；验收由 root 独立复验。

## 8. Wave A 已实现 / Wave B 编码清单

Wave A（已完成，`domain/experience.go` + `domain/experience_test.go`，纯规则不依赖 M0 共享类型）：
- [x] 枚举 + 校验：`ValidVerdict`/`ValidStatus`/`ValidateCapture`（pattern/verdict/condition/evidence）
- [x] 晋级纯规则：`Promote`（quarantine↔approved，仅 approve/demote 合法）
- [x] 作用域 + 分享审计：`ScopedForProject`（默认隔离）/`ShareRecord` + `ValidateShare`（谁/何时/到哪）
- [x] 失败来源分类：`CaptureEligible`（仅 `agent_task`+`failed`；`job_failed`/`import`/`distillation` 不被伪装成 agent_task 失败）
- [x] 单包验证：`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain/` PASS

Wave B（待 M0 契约 SHA 发布后编码）：
- [ ] fetch + merge 集成分支取 M0 契约类型（Experience schema / EventRef / Event）
- [ ] `app/experience_*.go`：CaptureFailure / ApproveExperience / ShareExperience / ListExperiences + 服务接口（人工关口按 Principal.Kind）
- [ ] `adapters/sqlite/experience_store*.go`：仓库 + Tx 接口 + 迁移表读写（迁移 `014+` 由 W0 负责，跨轨 Handoff）
- [ ] `tests/attention-engine/M5-experience*`：三组验收测试（失败候选捕获 / 晋级关口 / 作用域分享），含 SQLite 持久化 + 冷重启
- [ ] 串行验收通过后 commit + push，发 worker_done succeeded

## 9. 阶段收束记录（2026-10-11）

真实 partial 状态（本轮交付，commit `fe4f8c6`）：
- 已交付并验证：`internal/attention/domain/experience.go` 纯规则 + `experience_test.go` 单测（`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain/` PASS）；`tasks/AttentionEngine-W5.md` 设计与 Handoff。
- 未交付（真实剩余）：`app/experience_*.go` 服务层、`adapters/sqlite/experience_store*.go` 存储层、`tests/attention-engine/M5-experience*` 验收测试（含 SQLite 持久化 + 冷重启）。M5 未通过，不算完成。

下一步（新阶段，切 codex --yolo 后接续）：M0 契约 SHA 发布后 fetch+merge 集成分支，实现 app/store/验收测试，完成「失败事件候选 + approved + project scope + 人工分享 SQLite 持久化及冷重启」串行验收，再 worker_done succeeded。
