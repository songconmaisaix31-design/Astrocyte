# 动态任务 DAG 与 Worker 生命周期

## 身份、范围与基线

实际客户端为 Orca Codex 0.162.0，模型为 GPT-6.1-Sol xhigh fast，继承已有启动设置。Task 为 `task_50ff3b1212c2`，Dispatch 为 `ctx_e5e98fad4345`，Run 为 `run_73578b70c449`。唯一实现工作树为 `frontend-observatory-20261018`，分支为 `songconmaisaix31-design/frontend-observatory-20261018`，起始 HEAD 为 `657ca1464517a171ed50ddc8fcc787b237c44ad9`，接管时干净。

本轮继续 Draft PR #2，不合并、部署、打包或发布。独立 Reviewer 使用 `dynamic-dag-review-20261019` 工作树，只审查业务并记录自己的验收结果。主控管理任务和资源窗口，同一实现 Worker 负责全部返修。

## 已核查的生产构件

| 构件 | 当前源码与行为 | 对本轮的约束 |
|---|---|---|
| 生产 Swarm 装配 | `cmd/server/relay_swarm_assembly.go` 将同一 SQLite、Workspace 批准服务、Registry、prepared input reader 和对象存储注入 `NewSwarmService` | 账本和真实授权已装配；是否能执行仍由每次能力和权限检查判定 |
| 工作项生长 | `internal/swarm/app/commands.go:AddWorkItem` 允许 human/agent 经 `manage_work_items`、当前 Grant、Mission 版本及预算限制创建工作项 | 固定来源必须来自批准；写域不可扩大；依赖必须是同 Mission 的已有工作项；新 ID 由服务生成 |
| 当前任务查询 | `ReadMission` 经 `WithMissionRead → ReadApproval → projectAccess` 核当前项目及Agent的read_context，返回工作项、当前租约、产物、执行历史和真实 outbox 时间线；该链路没有核read_swarm | 新图查询增加用途明确的read_swarm读取授权，在同一事务核当前Settings和成员Grant；不直接复用旧检查宣称权限已足够 |
| 执行权威 | `SwarmNativeBridge` 使用原 Registry 的精确 session/input operation，核查原绑定的 controller/startup/native 元组 | 执行结果、原控制器持有、进程退出分别处理；无观察不能宣称死亡 |
| 已有停止 | `PauseMission`、`CancelMission`、`RecoverMission` 使用 `ExecutionOwnership.Stop/Reconcile`，只有确认退出和可结算用量才能释放原租约 | 停止失败或未知用量保留原预留和写域；旧 fence 失效；不创建替代执行者 |
| 已有停止限制 | Mission 具有截止时间、总工作项、尝试、调用/Token/金额及无进展限制；原生控制器具有实际 deadline | 不引入第二份 energy 或调用预算；无进展和预算只能使用实际已保存事实 |
| 原生限制 | portable member evidence 当前 `ToolsKnown/TargetsKnown/CallsKnown/TokensKnown` 保持 false；当前 Codex 文件工具关闭条件未满足 | S3 Claim、真实模型规划及真实 Worker 执行正向验收保持阻塞；不替换后端或伪造能力 |
| 前端 | `SwarmPage` 保留已有任务和时间线；协作入口仍为禁用的 `DesignGraph`；统一 `useReadApi` 支持 refresh、abort、stale | 本轮启用真实 DAG，继续使用统一查询与现有任务详情入口 |

以上属于源码核查；本轮尚未运行产品服务或模型。历史报告中的成功不作为本轮正向生成或生命周期证据。

## Morphogenesis 参考

只读参考 HEAD 为 `ddd2849afe330bf221d405dc2b91948c2416bd47`。已读取 `WorkerConfig`、`Worker.run()`、`TaskLedger.enqueue`、`TopologyEngine.advance_idle_window/prune`、`docs/tracks/topology.md` 及 README。

Worker 的 cycle/claim energy、sense 次数、实际运行时限、连续空闲和 budget sleeping 均有明确停止条件；未知发送先保留，不能用本地能量耗尽掩盖。任务账本在同一事务中固定依赖，要求依赖已经存在，不允许改写既有任务的身份与依赖。

连接裁剪同时要求低权重、明确推进的空闲窗口和最少活动连接，required 连接保留。该规则属于已配置的路由偏好；任务依赖、提交、复核和执行事件作为事实保留。Astrocyte 当前没有该类路由权重及反馈契约，本轮不会将任务依赖改成可衰减权重，也不会因页面隐藏节点删除事实。

## 提交主控核查的实现设计

### 任务 DAG 与来源

增加只读 `GET /api/v1/missions/{id}/dag`，在当前项目和read_swarm授权下复用 `ReadMission` 的一致事务。DTO 包含 `schema_version`、`mission_id`、`mission_version`、服务端 `observed_at`、工作项节点、依赖边、Worker 参与记录及单独的执行/协作事件列表。

节点为真实 WorkItem ID，边方向为 dependency → dependent。新工作项保存实际创建主体、主体类型和创建时间，随 `swarm.work_item_created` 保存到同一 outbox 事务。旧记录缺少创建主体时明确显示来源未记录；不能推断为模型生成。Agent 凭据创建只能证明 Agent 主体，真实模型生成还需原生输入/输出及会话来源，本轮报告分别计数。

Domain 验证拒绝重复节点、自依赖、重复依赖、缺失端点和环。AddWorkItem 在保存前验证合并后的完整图；只读 DTO 也验证历史图，损坏时立即报错，不跳过边。既有“只依赖已有项”的写入规则继续保留。所有依赖为 required 事实；协作往返放在另一个列表，不加入拓扑排序。

每个节点和依赖给出实际 `swarm.work_item_created` 的 event_id、aggregate_version 和发生时间。执行/复核列表继续返回原 operation/session/binding 及真实事件；图详情保留原任务、上下文和产物入口。无来源时明确缺失，不补造事件 ID。

### Worker 参与与退出

Worker 身份使用真实 session 与 binding，operation 为参与记录的执行身份；相同 Worker 的多次执行保留各次来源。当前 lease 的存在、到期、原生观察、回合完成、确认退出和资源释放分别显示。租约到期仅表示不能持有当前写权限；submitted/accepted 不代表进程已退出；服务重启丢失内存权威时显示无法核实。

补充用途明确的只读参与观察端口，由生产 `SwarmNativeBridge` 读取现有 Registry evidence。它只访问已由当前 Mission 执行历史确定的 tuple，并核对 project/session/operation；不扫描配置或会话目录，不启动或接续原生，不进行模型调用。明确缺失的原控制器返回 `unverifiable` 与能力原因；身份不符、数据库错误等立即失败。

持久退出记录只能来自实际 wait/stop 确认。现有 `execution_reconciled` 可证明当时的确认退出；应在原事件负载中保留退出确认、实际观察时间及原有安全原因，并投影到 Worker 历史。原 `execution_reconciled` 代码已要求 StopConfirmed，但历史负载缺少实际退出时间；旧记录时间只标为对账事件时间。当前 Registry 的 `ExitObservedAt` 没有实际赋值，不用读取时间代替进程退出时间。

停止机制继续使用当前 Mission stop limits、原生 deadline 和原有 Stop/Reconcile。主控在 `msg_d4d7958cd78a` 明确采用服务生命周期内单一有界检查，具体行为如下。

1. 使用现有 `workerCtx`，启动时检查一次，之后每5秒检查一次；每轮串行读取现有 Mission。仅处理 ready/running，核其已批准 deadline、已有无进展轮数/时限和能准确判定的已结算资源耗尽。资源缺少已知值时保持未知；不能把全部预留等于上限当作正在执行的任务已耗尽，也不能用响应次数替代调用预算。
2. 对每个触发 Mission，在同一 SQLite 写事务内重新核 Mission 版本、状态和条件，并取得该 Mission 的当前全部原租约；先将 Mission 设为 blocked，保存准确停止原因，给当前工作项增加 fence 并保存 blocked，写入 `swarm.mission_stop_requested` 与按原 operation/session/fence 关联的 `swarm.execution_stop_requested`。事件来源固定为内部 `system:swarm-lifecycle`，不借用 human/agent，也不授予新授权。
3. 事务提交后串行调用原 `ExecutionOwnership.Stop`；不持有写事务等待外部进程。每次结算重新核对原 operation/session/lease fence，复用 `settleExecutor`。确认进程退出与可结算用量同时满足后才释放原租约；退出与资源释放分别投影。触发时无租约只保存 Mission 停止事实，不制造 Worker。
4. 明确的原生能力缺失、交付未知和用量证据不足作为该 Mission 的 blocked 事实保留；不重复 Stop，不释放原预留，也不在下一个 tick 重试。数据库错误、tuple不符或其他未分类错误立即返回主服务，触发已有失败/关闭流程；不记录日志后继续循环。每轮无派发、无认领、无模型调用、无后端切换。
5. 服务关闭时取消计时及后续检查，等待唯一检查任务结束。事务已经固定停止意图后发生取消或崩溃，Mission 仍为 blocked，原租约保留；下次启动不会重放 Stop。原生控制器仍由已有关闭路径处理，不宣称关闭已成功。用户通过现有显式 Recover 对账原 operation，核当前授权和版本；停止条件仍成立时保持 blocked。
6. 内部检查只在 Go 应用服务向生产装配开放，不注册可被 HTTP/token调用的系统身份入口。GET DAG 仅读取，不推进窗口、改变状态或终止进程。停止前后的实际事件和历史节点永久保留。

当前没有已启用的自主 Worker 规划循环，也没有可据以裁剪的路由偏好，完整“模型规划→生长→逐步退出”的真实正向链路仍受原生和 S3 门禁约束；此范围不得由只读图验收替代。精确代码范围仍等待主控批准。

### API、前端与动态更新

新增 DAG DTO 与端点写入现有 OpenAPI，使用官方生成脚本更新 schema。现有精确Agent路径分类在 `relay_scoped_workspace.go`；只增加 `GET .../dag`，不放宽其他方法。

新增 `MissionDAGReadAuthorizer.WithMissionDAGRead` 端口，由现有 `SwarmApprovalBridge` 在同一Workspace/Swarm事务实现：先读取实际Approval以确定项目，保留现有projectAccess；再加载当前Project，以 `ValidateProjectActionScope` 核当前Settings的read_swarm，对Agent加载当前ProjectGrant并使用 `AuthorizeProjectActionCaller` 核同项目、read_swarm、过期和撤销。human使用原认证会话并核当前项目Settings，读取过期Mission批准的历史不要求执行授权重新生效。该读取不调用执行资源预留检查，也不因为历史预算耗尽拒绝读取；不授予新写权限。添加工作项继续调用原端点。

前端在原蜂群页提供“任务依赖”视图和 Mission 选择，初始仍可看现有任务与时间线。采用 `@dagrejs/dagre` 进行布局，布局不计算业务状态；依赖安装前读取其当前官方包资料。方向切换、缩放、节点键盘选择和依赖详情均只改变显示。保持节点 DOM 身份，更新后保留可用选中项；所选节点不存在时明确说明。

复用 `useReadApi` 对选中 Mission 每5秒读取，变更提交后立即刷新；取消已过时请求，离开页面停止计时。服务端时间和任务版本可见；请求失败保留旧图并明确 stale，提供重试；重连后重新取得完整快照。不将轮询当作 Worker 心跳或死亡判定。空图显示真实空态；权限撤销后不继续展示旧项目内容。图与 Worker 历史可独立选择，退出记录不会从历史消失。

### 申请的精确 write_paths

| 写域 | 用途 |
|---|---|
| `AGENTS.md`、`docs/standards/dynamic-dag-20261019.md`、`.gitignore` | 当前已授权；同步最新第3节、设计和实际验证；`.verification/` 已被忽略，无需重复修改 |
| `docs/SPEC.md` §18.2、`docs/TASKS.md` 当前开发流程入口 | 仅同步主控已确认的唯一 Worker / 独立 Reviewer 规范，取得批准后执行 |
| `internal/swarm/domain/dag.go`、`dag_test.go`、`mission.go` | DAG 完整性、真实创建来源及当前限制原因 |
| `internal/swarm/app/dag.go`、`dag_test.go`、`contracts.go`、`commands.go`、`lifecycle.go` | DTO、原授权下的快照、写前图验证、参与观察端口与原生命周期事件证据 |
| `internal/adapters/sqlite/swarm_history.go`、`swarm_native.go`、`swarm_approvals.go`、`swarm_dag_test.go` | 从真实 outbox 和 Registry 读取来源及参与，同一事务核当前read_swarm权限，实际 SQLite 验收 |
| `internal/adapters/httpapi/swarm_missions.go`、`swarm_dag_test.go`、`relay_scoped_workspace.go` | DAG GET 注册与精确 Agent GET 分类 |
| `cmd/server/relay_swarm_assembly.go`、`swarm_lifecycle.go`、`swarm_lifecycle_test.go`、必要 `main.go` 接缝 | 同一生产端口装配及经批准的停止检查；不提供新执行后端 |
| `contracts/openapi.yaml`、必要 `contracts/examples/swarm-*.json` | 新 DTO 和端点 |
| `web/src/api/schema.d.ts`、`web/src/api/swarm.ts` | 生成 schema 与统一客户端 |
| `web/src/hooks/useReadApi.ts` | 增加可选的错误保留规则，DAG在401/403/404时清除已撤销或不再存在的数据；其他调用保持原行为 |
| `web/src/pages/swarm/dag/**`、`web/src/pages/swarm/SwarmPage.tsx` | 布局、真实查询、键盘和详情 |
| `web/package.json`、`pnpm-lock.yaml` | Dagre 锁定依赖 |
| `web/e2e/dynamic-dag.check.ts`、`web/dynamic-dag.config.ts` | 实际 HTTP / SQLite 双视口、断连及键盘验收，独立显式配置；默认 CI 不改变 |

此表为申请，代码路径在批准前不修改。停止检查和实际 Agent 生成的执行窗口分别等待主控裁决；不把未回答的权限当作可用。

## Reviewer 决策通知 hook

主控在 `msg_580f42fd9146` 追加授权 `scripts/review-decision-hook.ps1`。Reviewer 每次实际形成静态或运行决策后，使用自身当前 Run/Task/Dispatch/发信终端、准确40位候选 SHA、`accept` / `revise` / `insufficient_evidence`、实际非空报告绝对路径和简短依据调用脚本。

脚本先验证必填参数与报告文件，再通过本机已安装 `orca.exe orchestration send` 单次发送 `type=handoff`、`subject=ReviewDecision` 到指定 `run:` 主控邮箱，使用 typed Task/Dispatch/report-path 参数。正文保留候选、决策、审查身份、报告与依据；原生 JSON 回执直接输出。失败立即返回，未知发送不自动重试。没有守护进程、定时任务或另一份调度状态。

发送成功只表示持久入队及原生唤醒请求；主控收到并回复后才记录实际唤起。Worker 不替 Reviewer 发送接受或退回决定。本轮 hook 的正向通知由独立 Reviewer 使用真实决定执行。

已执行 PowerShell AST语法检查及两次实际脚本拒绝检查：未提供参数立即以非零退出；使用本轮真实身份和候选但缺少报告文件时在调用Orca前非零退出。两项原件分别为本树 `.verification/hook-missing-parameter.log` 和 `hook-missing-report.log`。`git diff --check` 通过。没有发送审查决定，也没有将参数检查计为实际唤起。

## 可运行验收与阻塞

编译、测试与服务使用主控独占串行窗口，双采样 CPU<70%、可用内存≥4GiB，间隔2–5秒；BelowNormal；Go `GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`；Node 512MiB；浏览器单 worker，最多一个自有服务。仅复用 D 盘已有缓存。原件放本树已忽略 `.verification/`。

1. Domain 参数测试核查重复、环、缺端点和一致节点来源，不使用伪执行者。
2. 独立 SQLite、生产 HTTP 服务、Agent 创建的批准公开 Git 项目，使用真实人类会话创建批准/Mission/工作项；核查新增节点/边、幂等、范围、跨项目拒绝与真实 SQL event_id。人工创建明确标为人工。
3. 双视口1920×1080、1280×720，在同一真实服务核查更新、方向/缩放、详情、键盘、空态、长标题、HTTP断线与重连、stale和权限拒绝。禁止截图、视频、trace或像素读取。
4. 退出与资源释放分别核对实际 Registry/真实事件；没有合法执行来源时只验空与拒绝路径，不构造会话、租约或退出事实。
5. 主控批准且原生和 S3 门禁满足后，以同一单后端运行真实模型生成、认领及停止，再逐项回查真实会话与事件；门禁未满足时完整正向链路保留未验。
6. 前端 typecheck/lint/build、契约生成一致性、架构检查及适用 Go 测试/vet/build；原前端、可选 OpenViking、批准与作品入口定向回归。完整默认 CI 单列，既有失败、警告和未执行项保留。

基线与候选使用相同公开来源和资源条件，记录真实节点/边、请求次数、响应时间和实际资源；未测不写性能比例。Reviewer 独立给出接受、退回或证据不足。本轮当前只有静态读取及规范同步，未运行重操作，无真实模型或退出验收结果。
