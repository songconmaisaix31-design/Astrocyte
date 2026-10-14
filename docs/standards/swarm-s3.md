# Swarm S3 / S6 实施与实际验证

## 当前阶段

开发 Task `task_da9aebf03f83` / Dispatch `ctx_d4bf65ca06bf`，模块 Task `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa`。使用当前 Orca Codex 客户端和会话所配置模型；本阶段未调用外部模型，也未根据客户端名称推断模型。独立分支 `songconmaisaix31-design/product-swarm-20261015`，接管源码 `391c6176145044b099c13a05a926aef2096db170`。

第一阶段 SOURCE `b220ab27c4e57d0bfdfaef224a9d03c1ae92366a` 已普通 push，`git ls-remote` 核实远端同值。通过普通合并消费 B 的五表迁移 SOURCE `da48a114d91181075bd32b62996beadd9c2c0a25` 和 S2 的 Grant 公开端口 SOURCE `4eec31ddfde39fa3f548691ec6e7f2f34470351a`。全部直接编辑位于本 Task 独占路径；共享迁移、依赖、Workspace 业务与 Web 由各原负责人维护。

## 实现与接口

`internal/swarm/domain/mission.go` 定义 Mission、WorkItem、Lease、Artifact、固定来源版本、写域、资源与停止条件。货币使用明确币种和最小单位，空费用保持未知；预算上限与已知资源数量分开，空 calls/tokens 上限保留原值。

`internal/swarm/app/contracts.go` 和 `commands.go` 提供 CreateMission、AddWorkItem、ClaimWorkItem、SubmitArtifact、AcceptArtifact。Command 的 schema_version、request_id、expected_version 位于 JSON 顶层，Idempotency-Key 来自受信任入口。Principal 不能从命令正文反序列化。采用要求人类身份，通过真实 ExperienceRecorder 返回经验候选引用；经验候选的后续批准与分享继续由原治理接口处理。

认领在授权写事务内重新检查成员、依赖、当前任务版本、项目活跃执行者上限、写域占用与预算，然后保存工作项、单调 fence、租约、资源预留、幂等回执和 outbox。过期租约继续占用资源，删除租约必须使用匹配 fence。提交保存独立 Artifact；提交与采用使用不同命令。Git merge、push 和 release 未作为采用副作用。

`lifecycle.go` 提供暂停、取消和显式恢复：先持久化停止新认领的状态，再调用原执行者的停止或对账端口。原 operation/session 或用量未知时保留租约、预算和写域，标记 blocked；恢复只核对原身份，不创建执行者。新认领仍要求显式请求和当前批准。

`sqlite.NewSwarmRepository(db)` 使用真实 SQLite BEGIN IMMEDIATE；`NewSwarmApprovalBridge(db, actualWorkspaceGrantAuthorizer)` 消费 S2 的 `WorkspaceTransaction(ctx, db)`，复用同一数据库连接。桥接不读取 Workspace 内部表。`NewSwarmGitEvidence(objects)` 核验独立 Git worktree、共同仓库、固定 commit、真实 diff 对象和固定 git diff --check 记录；路径判定由 Git 完成，无自写 Git 格式 parser。

## 已运行验证

| 命令 | 实际结果与覆盖 |
|---|---|
| `go test ./internal/swarm/... ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1` | PASS；三个真实 SQLite 组件测试。两个独立 DB 连接并发仅一个认领成功；失败事务没有半预算；版本/fence/尝试计数、回执与 outbox 一致；关闭重开后过期未知租约保留；grant 唯一；消费事件去重；旧 fence 无法删除租约 |
| `go test ./internal/adapters/sqlite -run '^$'` | PASS；适配层编译验证 |
| `node scripts/check-architecture.mjs` | 在消费 S2 前 PASS，31 包；消费 `4eec31d` 后 FAIL，Workspace Domain 引入 apierrors，已交 S2 原负责人返修；Swarm Domain 自有 RuleError，未越层引用 |
| `git diff --check` | PASS |
| `git push` / `git ls-remote` | 第一 SOURCE push 成功，远端完整 SHA 一致 |

SQLite 测试中的明确参数仅验证真实数据库与领域规则，未声称存在真实批准、原生会话或 Agent。该组件结果不等于 AT08 两真实执行者完整通过。

首失败保留：初次 SQLite 测试编译误用了不存在的 `DB.Migrate`，实际方法为 `RunMigrations`，修改测试调用后上述三项通过；首次 PowerShell gofmt 命令把 `swarm_*.go` 作为字面路径传入而失败，随后使用准确文件路径完成格式化；首次元数据补丁因当前格式与预期文本不同拒绝应用，读取当前文件后重新应用。没有降低断言或新增 skip。

## 验收与依赖

| 场景 | 当前结论 |
|---|---|
| AT07 真实 S2 批准、重复启动、撤销启动拒绝 | NOT_RUN；S2 019 迁移与实际业务验证尚待原负责人 SOURCE。当前只有同 grant 最多一个 Mission 的真实 SQL 约束组件通过 |
| AT08 完整授权/成员/两个实际 Agent/HTTP 认领与成果 | NOT_RUN；SQLite 两连接 CAS、无半预算与旧 fence 组件已通过，真实成员与原生原 ownership 端口尚待接入 |
| AT09 原会话真实 resume 与一次显式交接 | NOT_RUN；R SOURCE `79d93132967de20fb1fda3d8d74b4edff386178c` 的安全 current ownership 端口已通知，实际适配与合法原生条件待完成 |
| AT13 在途崩溃与原 operation 对账 | NOT_RUN；未知占用保留的 SQLite 重启组件通过，原控制器实际在途场景待完成 |
| S6 LocalPolicy、无进展停止、第二实际任务引用 approved 经验 | NOT_RUN；本 Task 继续负责实现，经验治理/查询走原公开接口，不修改冻结经验内部 |

当前应用服务尚未接入 H 的实际 HTTP 主路由/鉴权/装配、F/UI 页面或 R 的真实成员与生命周期实现。Git/object 验证器已编译，真实 Git 集成测试仍待运行；Artifact 采用需要真实经验候选管道。提交重放及生命周期幂等的完整服务测试仍待 S2/R 接入后实施。完整核心路径、默认检查、构建和真实两 Agent 改进尚未通过，不将此阶段登记为完成，也不发送 worker_done；继续同一 Task 返修和集成。
