# Workspace S2

当前开发入口为 `product-video-20261014`，Task `task_6846cdb90c46`，Dispatch `ctx_f32b6d92119a`。实际客户端 Codex，模块已核模型 `gpt-6.1-sol` / high fast。基础精确源码 `391c6176145044b099c13a05a926aef2096db170` 已普通合并。

## 批准与事务

`app.NewS2Service(repository, files, references)` 提供版本化 Goal、Proposal、human-only 批准和撤销。Proposal 的 `version` 是每次写入的 CAS 版本，`revision` 是内容版本；更新任何内容生成 successor revision，旧 revision 保留。状态为 draft、in_review、approved、declined、superseded。Grant epoch 独立递增，状态为 active、revoked、expired、completed。

批准命令固定 `expected_version` 和 `expected_revision`，保存 Grant、提案状态、foundation idempotency 成功回执和 foundation outbox 事件在同一 BEGIN IMMEDIATE 事务内完成。HTTP 不启动 Agent。相同 caller / command / key 的请求使用完整 JSON 输入摘要；不同输入返回 version_conflict。

`GrantAuthorizer.QueryGrant(ctx,id)` 是供受信任跨上下文桥接读取的端口，HTTP 使用另一个带身份检查的 ReadApproval。`WithAuthorizedGrant(ctx,domain.GrantCheck,callback)` 在当前事务内重查 epoch、固定提案 revision、状态、过期、项目目录权限、真实 canonical root、动作、工具、外发目标、并发及累计资源。callback 只执行数据库动作。S3 SQLite adapter 使用 `sqlite.WorkspaceTransaction(ctx,db)` 复用同一连接，不能另行 Begin。

调用、token 和 money_minor 预算分别使用可空整数，钱款有 currency；unknown 保持 null。需要至少一项显式资源上限；有上限的每个维度必须提供包含本次预留的累计用量。停止条件含 deadline、conditions、max_work_items、max_attempts、no_progress_seconds；并发由 scope.max_concurrency 显式提供。

事件使用 `workspace.proposal_approved`，payload 为 proposal_id、proposal_revision、grant_id、epoch。`workspace.approval_revoked` 含 grant_id、new_epoch、reason。`ReadWorkspaceEvents(ctx,afterEventID,limit)` 按 foundation outbox rowid 提供稳定续读；不会删除或消费事件。消费者在自己的事务中幂等记录 event_id，并使用当前授权检查。

## 交接

B 独占 migrations、共享 service.go/contracts.go 和依赖；已收到 019 Workspace 六类聚合表及 foundation_idempotency.result_json 的准确 DDL。H 独占 server.go/main/OpenAPI，调用 `h.registerWorkspaceS2(mux,services.WorkspaceS2)` 并移除已被覆盖的 GET proposals、POST submit/approvals/revoke S0 route。应用装配使用 `sqlite.NewWorkspaceRepository(db)`、既有 ProjectFiles 与 ProjectReferences，不启动 native controller。前端从 H 的 OpenAPI 生成类型。

## 当前验证

`go test ./internal/workspace/app ./internal/workspace/domain ./internal/adapters/sqlite -run '^$'` PASS，仅证明编译。真实 SQLite、HTTP、重启、事务回滚与批准权限验证目前 NOT_RUN，等待 B 的 019 和 H 的真实主服务装配。ContextPacket 持久 successor 和 SessionBinding 继续实施；原生 owned 安全接口及用户 working 条件仍 UNKNOWN，不调用其他 CLI 补证。

## 第二阶段实现与当前业务覆盖

B `da79452d1735f8fceca103aaf9d10bb53182f1f5` 已普通消费。新增 immutable ContextPacket / successor、固定 Goal 与材料引用、对象存储发布、项目快照验证，以及只接受当前 owned 原生事实的 SessionBinding；构造后通过 ConfigureContext 装配 SnapshotProvider / ObjectStore / OwnedSessionProvider。R/H 的 owned 提供者和 S4 的真实快照仍待准确实现；这两条实际成功保持 NOT_RUN，不制造原生会话或项目快照。

`go test ./tests/s2 -run TestActualSQLiteApprovalRestartAndRollback -timeout 45s -v` PASS：实际新 SQLite、既有生产项目注册/设置/空间入口、Goal/Proposal、相同输入幂等/异输入冲突、Agent 拒绝、旧版本冲突、人类批准、foundation outbox 重读、真实 SQLite 唯一约束失败整事务回滚、关闭重开及撤销后拒绝。最后一次库为 `.verification/s2-first/actual-700047605/state.sqlite`，日志 `sqlite-third.log`。该检查使用应用与存储实际实现；HTTP 验收尚 NOT_RUN。

首次真实路径校验 FAIL 保留在 `.verification/s2-first/sqlite-first.log`；现通过 filepath.Rel 向既有 ValidateProjectPath 传相对路径，并复核现行 AllowedSubdirs。首次 Domain import FAIL 保留在 `architecture-first.log`；现在领域自有 WorkspaceRuleError，应用 NormalizeWorkspaceError 映射统一错误，架构检查 32 包 PASS。

Proposal 读 DTO 增加 grant_id，刷新页面可读取批准与撤销入口。scope.settings_revision 固定当前项目授权版本；保护动作持有可信 Caller，当前项目版本变更阻止新动作。scope.write_paths 明确相对根目录的写入范围。stops.max_no_progress_rounds 实现 SPEC 连续无产出轮数约束；no_progress_seconds 可额外显式配置，不取代轮数。

## 第二阶段独立 REPORT

SOURCE `5d0ba735533f68259b91128dc404c8e4c2f2a535` 普通 push 成功，独立 remote 查询相同 SHA。实际 SQLite 测试与架构检查通过；HTTP、资料访问拒绝、上下文真实快照与 native binding 的完整成功覆盖仍 NOT_RUN。Grant/成员/Goal 端口、ContextSnapshot 类型、提供者装配和准确 UI 字段已分别交 S3、B、H、S4 与 UI；下一项依赖为 B 的纯当前策略 `4725f0b1c6907ddefb3b1e1dac0f49840031db59` 和 H 的主 HTTP 装配。原架构与路径首次 RED 均保留。

## 第一阶段独立 REPORT

SOURCE `4eec31ddfde39fa3f548691ec6e7f2f34470351a` 普通 push 成功；独立 `git ls-remote origin refs/heads/songconmaisaix31-design/product-video-20261014` 返回相同 SHA。四个包 `go test -run '^$'` PASS，`git diff --check` PASS；两项均不代表业务验收。准确 SOURCE 已交 S3、H 和 UI，B 的 019 已交 `da79452d1735f8fceca103aaf9d10bb53182f1f5`，下一步消费该迁移并运行真实 SQLite/HTTP。第一次 B SOURCE 通知出现 PowerShell 单引号参数错误，CLI 明确 invalid_argument，未投递；改正参数后继续交接，不将失败视为成功。
