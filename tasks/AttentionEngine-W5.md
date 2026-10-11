# AttentionEngine W5 · Experience Memory（M5 经验侧）

- 轨道分支：`songconmaisaix31-design/attention-engine-w5-experience`
- 事实源：任务书 `tabbit-code (2).md` §3/§4.3/§4.5/§7/§8 M5；主计划 `tasks/AttentionEngine-plan.md` W5 行
- 独占写域：`internal/attention/domain/experience*.go`、`internal/attention/app/experience_*.go`、`internal/adapters/sqlite/experience_store*.go`、`tests/attention-engine/M5-experience*`、本文件
- 状态：限定 W5 实现与实际本地集成已验收，主控已确认结算。SOURCE `b43d2a4cddd8a3f4a367676f9baaa6d3f3414ade` 已 push；8 个 SQLite 经验用例通过，W0 HTTP 与 W6 stdio 复验已回传。原始输入为 contract_local fixtures，真实 native 失败回放 NOT_RUN，整体 M5 未完成。

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
- 已有失败事件 `attention.job_failed`，payload `{job_id, operation_id, delivery_unknown, attempts, error, cause, cause_detail}`（`jobs.go:592`）。这是导入/蒸馏作业证据，不能直接映射成 `agent_task` 或触发经验捕获。原交接将其称为 agent_task 失败轨迹的判断已纠正。

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

Wave B：
- [x] 普通 merge W0 精确 `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`，复用 domain Experience / EventRef / app StoredEvent、StoredExperience 和 `014/015` 迁移。
- [x] `app/experience_service.go`：CaptureFailure / CaptureFailures / 人工 approve、demote、share / approved-only GetExperiences / 原始证据解析。
- [x] `adapters/sqlite/experience_store.go`：复用 EngineRepository 的 BEGIN IMMEDIATE；原始事件唯一键捕获幂等、乐观版本、审计原子写入。
- [x] `tests/attention-engine/M5-experience*`：8 个 contract_local 源事件的真实 SQLite 验收，含冷连接重开、跨连接并发捕获、项目范围、人工审计及不可变命令结果回放。
- [x] 消费 W1 精确 `e9bb380` / `6730531` 生产 EventReader 与提交后自动失败捕获 hook；实际服务入口和重启回放通过。真实 native 源权限未答，仍为独立 NOT_RUN，不用 fixtures 替代。
- [x] 本轨限定范围验收通过，主控确认结算；最终报告提交 push 后发送一次 worker_done succeeded，整体 M5 与 native 回放另记。

## 9. 阶段收束记录（2026-10-11）

真实 partial 状态（本轮交付，commit `fe4f8c6`）：
- 已交付并验证：`internal/attention/domain/experience.go` 纯规则 + `experience_test.go` 单测（`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain/` PASS）；`tasks/AttentionEngine-W5.md` 设计与 Handoff。
- 未交付（真实剩余）：`app/experience_*.go` 服务层、`adapters/sqlite/experience_store*.go` 存储层、`tests/attention-engine/M5-experience*` 验收测试（含 SQLite 持久化 + 冷重启）。M5 未通过，不算完成。

下一步（新阶段，切 codex --yolo 后接续）：M0 契约 SHA 发布后 fetch+merge 集成分支，实现 app/store/验收测试，完成「失败事件候选 + approved + project scope + 人工分享 SQLite 持久化及冷重启」串行验收，再 worker_done succeeded。

## 10. Wave B 接续与依赖 Handoff（2026-10-11）

- 实际客户端：Codex，当前派发 task `task_01d6c1323a06` / dispatch `ctx_64ccc9c2ca1a`；不启动并发 Agent。
- 基线 `61497e7d0b83a7a1c2f54e402e6b9b06d57dd79f`，工作树干净；`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain/` PASS（cached）。这是纯规则基线，不是持久化验收。
- 捕获只接受实际持久化 `agent_task` 且明确失败的事件；导入、蒸馏、未知执行效果和私人原对话不能生成失败模式。证据引用原始 Event，不创建单独证据对象。
- 只读 W0 schema 观察：当前 `attention_engine_experiences` 无 project scope / 原始事件去重字段，尚无跨项目分享审计表。已通过 Orca Handoff 请求 W0 提供 owner project、原始 event 外键与幂等键、人工分享 who/when/where 持久化结构；迁移仍只由 W0 修改。
- 人工 approve/demote 与跨项目分享经受信 human 身份；Agent 检索仅返回其受信项目可见的 approved 经验。W0 HTTP 和 W6 MCP 接线通过 Handoff，不在本轨越界修改。
- 失败 outcome 的归一化字段、规则结构化输入、受信项目上下文等待协调者与共享 owner 确认；不据未提交文件冻结接口。收到精确 SHA 后普通合并，完成 SQLite 捕获、重启、幂等、人工关口和检索范围验收。
- 协调者已确认：M0 尚无可消费稳定 SHA，继续等待；项目身份来自应用 Principal/启动绑定，不能从 Event.Scope 自授；失败 outcome 只取明确原始元数据。依赖等待不发送 worker_done。
- 等待期间新增 `StructureFailure` 纯规则：pattern 为已观察到的 `agent_task`，verdict 为 `condition`，condition 为 `outcome=failed`；不把单次失败推出为 avoid 建议，不读原对话文本。导入/蒸馏/未知效果拒绝进入此规则。`GOFLAGS=-p=1 GOMAXPROCS=2 go test -count=1 ./internal/attention/domain/` PASS；仍未替代 SQLite 验收。

## 11. SQLite 核心阶段与接线 Handoff（2026-10-11）

- 实际客户端/模型：Codex / gpt-6.1-sol（协调者运行态确认）；当前 YOLO 模式，不启动其他 Agent。
- 服务构造：`sqlite.NewExperienceRepository(db)` 与 `app.NewExperienceService(repo, EventReader, ExperienceOptions{ProjectID: trustedStartupProject, Clock: clock})`。绑定必须由受信应用配置提供，不能从事件/请求正文初始化；事件 Scope 仅用于核对该绑定。
- W0/W1 自动捕获接线：已持久化原始事件后调用 `CaptureFailure(ctx,eventID)`；重启补消费用 `CaptureFailures(ctx,from,to)`。服务重读原始 StoredEvent，pending/unknown/non-agent_task/import/distillation/job_failed 不捕获；原始 EventRef 只引用事件，未复制私人对话进入 Experience。
- W6 读侧：`GetExperiences(ctx,Principal,projectID,pattern)` 返回 approved 且本项目持有/人工分享的 `[]StoredExperience`；请求 projectID 必须等于可信绑定。Agent 不可读取本项目 quarantine；人类可通过 `GetExperience` 审阅本项目候选。分享接收项目无法审批、降级或再分享源项目经验。
- 人工动作：`ApproveExperience` / `DemoteExperience` / `ShareExperience` 实现 M0 `ExperienceGovernanceService`，命令要求 schema=1、RequestID、ExpectedVersion；版本冲突和非法迁移不修改记录。共享后降回 quarantine 会立即隐藏检索，保留原分享和动作审计。
- 首次 app/sqlite 编译检查 RED：误用了不存在的 `newID`，已改为现有栈 `crypto/rand.Text`；定向复验通过，原 RED 不覆盖。早期定向 app/sqlite 输出 `[no tests to run]` 仅证明编译，不算行为验收。
- 实际验收：`GOFLAGS=-p=1 GOMAXPROCS=2 go test -count=1 ./tests/attention-engine -run TestM5Experience -v` 4 PASS；四个输入来源明确为 contract_local fixtures，存储和重启使用真实 modernc SQLite。
- 适用检查：`go test -count=1 ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite ./tests/attention-engine` 四包 PASS；同四包 `go vet` PASS；`go build ./cmd/...` PASS；`git diff --check` PASS。均设置 `GOFLAGS=-p=1 GOMAXPROCS=2`。
- 仍待：W1 生产 EventReader/明确来源语义和自动失败捕获接线；W0 HTTP/cmd 人类入口组装；W6 真正 SQLite+stdio 消费与完整 explain_focus。此阶段不宣称真实会话或整体 M5 完成。

## 12. 自动入口与命令收据返修（2026-10-11）

- 普通消费 W1 `e9bb3809eadcbf48cf017b6e09ae0ef4112b1a90`、`673053186a12377c3a1dded4c0b987daa63f2dd5`；上述 §11 对 W1 的等待已解除。输入原始字节、Observation 与失败状态由生产 EventRepository 读取，不再使用测试读取双。
- W1 正式自动接线：`app.NewEventService(eventRepo, selectedStrength, experience.CaptureCommittedFailure)`；该 callback 在 receipt/signal 提交之后重读原 ID。W5 支持明确 `failed` / `failure`，不从 native operation accepted、header、工具报错或私人消息推断任务闭环。
- 项目入口薄方法：`experience.RecordEvents(ctx,eventService,receipts)` 先核对可信绑定，再提交事件和捕获候选。捕获失败保留 W1 返回的已提交事件 ID + error；只通过原 ID 显式核对/重放，`CaptureFailures(ctx,from,to)` 可重启补消费。没有自动重新执行 Agent 任务。
- W0 实际 HTTP 审查发现 d8e 阶段人工命令缺结果回放，原版本仅 CAS + actions 为 partial。按 owner Handoff 复用现有 `attention_receipts`，操作名隔离为 `Experience/{action}/{project}`；收据、聚合头、分享和动作审计同 EngineTx 提交，不修改 S1 命令或迁移。
- actor + project + operation + IdempotencyKey 绑定输入；复用既有 CommandMeta 校验与 semanticDigest（RequestID 是跟踪字段，不改变语义）。已接受同输入先于 CAS 回放不可变完整结果，改输入返回 version_conflict；新请求仍先检查当前版本。unknown/incomplete 收据只返回 DeliveryUnknown，不自动重做人工动作。
- 最终本轨定向验证：`GOFLAGS=-p=1 GOMAXPROCS=2 go test -count=1 ./tests/attention-engine -run TestM5Experience -v` 8 PASS，包括自动入口/捕获中断保留 ID、审批/分享/降级冷重启历史结果回放、审计一次及 unknown 待核对。
- 受影响四包整包 `go test -count=1`、同四包 `go vet`、`go build -o <本次临时目录> ./cmd/...`、`git diff --check` PASS；没有跑媒体模型或读取未授权 native 原对话。此前首次 newID 编译 RED 保留为独立历史。
- 协调者确认：本轨实现与真实 SQLite 核心可通过明确 contract_local 生产服务输入验收；原始 native 失败 task_live 回放由用户权限单独阻塞，不能据此称整体 M5/M1 已通过。当前保留活跃，等待 W0 HTTP 重复命令与 W6 stdio 集成复验；领域返修仍由本轨负责。

## 13. 最终限定范围交付（2026-10-11）

- 分支：`songconmaisaix31-design/attention-engine-w5-experience`；源码 SOURCE `b43d2a4cddd8a3f4a367676f9baaa6d3f3414ade`。阶段 `ecf3549` / `fdaaa70` / `d8e7333` / `b43d2a4` 均已 push；M0 与 W1 精确依赖均普通合并，无重写历史。
- 完成：仅明确失败 agent_task 的 quarantine 候选，规则结构 pattern/condition/evidence；证据指向并解析原始持久 Event，不复制私人对话形成模式。生产 W1 post-commit 自动捕获与原 ID 重放、SQLite 重启/幂等、可信项目 approved 检索、人类 approve/demote/share 和同事务审计/结果收据已实现。
- W5 自身检查：`GOFLAGS=-p=1 GOMAXPROCS=2 go test -count=1 ./tests/attention-engine -run TestM5Experience -v` 8 PASS；四包 domain/app/sqlite/attention-engine 整包 test、vet、cmd build、diff 检查 PASS。首次 newID 编译 RED 与 d8e 缺人工结果回放的 partial 保留，后续修复为独立证据。
- W6 owner 回传 SOURCE `2314b6036f3b57788398adb420e8d8e283f639cb`：真实 stdin/stdout 与冷进程验收 PASS，包含实际 W1/W5 hook 故障、capture_pending 原 ID 返回、冷进程补捕获一条 quarantine、重复补消费不变和原始行/字节不变。采用明确可信 contract_local 失败输入；公开 MCP 自称 failed 仍为 unknown。此项为 owner 回传，本轨未重复运行 stdio。
- 主控转交 W0 SOURCE `271d18cc4d7853bd1a9bc96875b8875733889837`：实际 cmd/server factory + W5 + SQLite + HTTP session/CSRF approve/share 验收 PASS；不同 trace/RequestID 同 header key、冷重启返回原结果，changed ExpectedVersion 409，foreign/Agent/缺 CSRF 403。审批/分享审计各一次、目标 share 一条、最终 version=3；无 W5 领域返修。此项为 W0/主控回传，本轨未重复运行 HTTP。
- 主控已明确允许按 bounded W5 实现及实际本地集成结算 succeeded；不据本轨结果宣称整体 M5 完成。
- 剩余限制与未执行：模式仅描述观察到的任务种类和明确失败条件，不从私人语义生成泛化建议；真实 native 失败 task_live 回放 NOT_RUN（原始 source scope 权限待答），完整 M1/M5/M6、最终总体集成与远程 CI 由主控继续。未启动其他 Agent，未改 S1 分支族或个人运行库，未新增产品前端。
