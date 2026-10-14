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

## 第一阶段独立 REPORT

SOURCE `4eec31ddfde39fa3f548691ec6e7f2f34470351a` 普通 push 成功；独立 `git ls-remote origin refs/heads/songconmaisaix31-design/product-video-20261014` 返回相同 SHA。四个包 `go test -run '^$'` PASS，`git diff --check` PASS；两项均不代表业务验收。准确 SOURCE 已交 S3、H 和 UI，B 的 019 已交 `da79452d1735f8fceca103aaf9d10bb53182f1f5`，下一步消费该迁移并运行真实 SQLite/HTTP。第一次 B SOURCE 通知出现 PowerShell 单引号参数错误，CLI 明确 invalid_argument，未投递；改正参数后继续交接，不将失败视为成功。
