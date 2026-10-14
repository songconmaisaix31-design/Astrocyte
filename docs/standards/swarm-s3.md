# Swarm S3 / S6 实施与实际验证

## 第二阶段当前结果

开发仍为 `task_da9aebf03f83` / `ctx_d4bf65ca06bf`，模块仍为 `task_b0197ca7ebdf` / `ctx_b1cd2e0a39aa`。第二阶段 SOURCE `3b78fd04dbf758e745e9880297ee1c2d6e58bfc2` 已普通 push，远端同值；准确消费共享 policy SOURCE `4725f0b1c6907ddefb3b1e1dac0f49840031db59`、S2 SOURCE `5d0ba735533f68259b91128dc404c8e4c2f2a535` 和 `507be52f4e31759dc8c3765aeea964f1f2efe013`。以下第一阶段记录保留当时结果。

真实 S2 → Swarm 链已通过：Attention 项目空间、真实 LocalProject、Goal、Proposal、人工 ApproveProposal，经两个独立 SQLite 连接并发 CreateMission，得到同一 Mission；同键不同输入冲突、关闭重开后的回执重放与撤销后拒绝启动均通过。未启动原生执行者，没有使用批准 reader 替代实现。首轮因 S2 的零并发限制返回 scope_denied；原 S2 修改允许真实零活跃数后，第二轮发现回调替换 context 导致另取同 DB 连接等待，已终止该次实际测试。Swarm bridge 在相同连接重新附加自身事务 context 后回归通过；两次失败保持原含义。

授权 bridge 现在在公开 Workspace 写事务中查询自身 Swarm 账本，传递实际活跃租约数、当前操作与累计 Used/Reserved/本次资源、可信 Caller、实际工具和外发目标。Claim 和 Submit 需要准确 `binding_id`，同事务消费 S2 的 `WithAuthorizedSessionBindings`，并核对真实成员；未知执行策略拒绝认领。调用和 tokens 使用独立 known 标记，没有完整实际用量证据时不释放占用。原 R 明确当前真实有效工具集合与目标 endpoint 仍 unknown，未以请求关闭配置或零 tool_calls 填为已核实。

`httpapi/swarm_missions.go` 和 `swarm_work_items.go` 提供 `registerSwarmS3(mux, *swarmapp.Service)`，Principal 从 `localCaller` 注入。Domain RuleError 在本域 HTTP 层转换为既有 apierrors，scope/approval 拒绝为 403，version/fence/budget/unknown 为 409。路径与正文 ID 冲突拒绝；外部请求不能写 trusted consumer 的 event_id。

公开路由为 GET missions、missions/{id}、missions/{id}/work-items、missions/{id}/timeline、artifacts/{id}?mission_id=；POST missions、missions/{id}/work-items、missions/{id}/pause、cancel、recover、work-items/{id}/claims、work-items/{id}/artifacts、artifacts/{id}/acceptance，均使用 `/api/v1/` 前缀。Mission 详情返回 schema_version、mission、work_items、leases、artifacts、timeline；时间线仅返回真实 outbox 的 event_id、type、aggregate_id、aggregate_version、occurred_at，不返回原始模型或控制器内容。H 负责主服务路由替换、鉴权规则、OpenAPI 与装配；本阶段没有修改这些共享文件。

`NewSwarmLedger(repo, actualGrantAuthorizer)` 可提供真实账本与批准操作，执行能力缺失时明确返回 unsupported_capability。完整 `NewSwarmService` 需要真实成员、原执行者、成果检查和经验管道。Git/object 实际集成测试已通过；Git diff --check 的结果单独保存，Submit 同时要求实际项目适用检查的 `ApplicableChecksVerified`。缺少真实项目检查时无法进入可采用提交。固定策略 `swarm.fixed.v1` 先筛当前权限、依赖、能力、占用与停止条件，再按相同固定输入匹配排序；策略版本保存在 Lease。

| 第二阶段验证命令 | 实际结果 |
|---|---|
| `go test ./internal/swarm/... ./tests/s3 -count=1` | PASS；固定 LocalPolicy 安全筛选与真实批准/两连接 Mission/幂等/重启/撤销链 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1` | PASS，4 项；新增真实 Git worktree、commit、diff/check 对象正路径及无关仓库/脏工作树/错误对象/遗漏路径拒绝 |
| `go test ./internal/adapters/httpapi -run '^$'` | PASS；真实领域路由编译，尚未当作实际 HTTP 验收 |
| `node scripts/check-architecture.mjs` | PASS，33 包；原 S2 Domain 越层首 RED 由其原负责人修复 |
| `git diff --check` / push / ls-remote | PASS；第二 SOURCE 远端完整 SHA 一致 |

新增首编译失败记录：Recover 的重放检查增加局部 err 后重复声明，编译拒绝；删除该重复声明后通过。第二阶段没有新增 skip、替代原生或外部 API 实现，也没有减少断言。

当前剩余：H 主 HTTP/鉴权实际服务尚未装配验收；真实成员绑定的原 full Start session / 同 Registry / controller / startup 与 input operation resolver 待 R/S2 原负责人闭合；真实有效工具集合、目标与分项 usage 缺证保持 unknown。实际公开小改进项目、获准固定测试/构建由模块向根确认，项目原生自由工具权限保持关闭。实际两 Agent、一次显式交接、人工采用、经验候选真实录入和后任务 approved 经验实际使用均 NOT_RUN。AT07 的真实服务批准、并发唯一 Mission、重启/撤销组件通过；完整 AT08/09/13、S6 实际第二任务、最终组合 check/build 仍未通过，本 Task 继续执行。

## 第一阶段记录

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
