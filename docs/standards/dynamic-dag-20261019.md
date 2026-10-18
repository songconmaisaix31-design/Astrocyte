# 动态任务 DAG 与 Worker 生命周期

## 身份、范围与基线

实际客户端为 Orca Codex 0.162.0，模型为 GPT-6.1-Sol xhigh fast，继承已有启动设置。Task 为 `task_50ff3b1212c2`，Dispatch 为 `ctx_e5e98fad4345`，Run 为 `run_73578b70c449`。唯一实现工作树为 `frontend-observatory-20261018`，分支为 `songconmaisaix31-design/frontend-observatory-20261018`，起始 HEAD 为 `657ca1464517a171ed50ddc8fcc787b237c44ad9`，接管时干净。

本轮继续 Draft PR #2，不合并、部署、打包或发布。独立 Reviewer 使用 `dynamic-dag-review-20261019` 工作树，只审查业务并记录自己的验收结果。主控管理任务和资源窗口，同一实现 Worker 负责全部返修。

动态DAG源码提交为 `e369537bfcf3e3dad8c6ee76cfc5778055ab1f08`，报告候选为 `b0a21a2cd03c074ba4097012be8f7e7748bb1f3d`。独立Reviewer已接受该候选的人工依赖图、增长、来源、权限、恢复和无Worker停止构件，完整原生链路仍为证据不足；对应[CR 5564261919](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5564261919)由Reviewer本人发布。主控随后批准同一PR中的局部按需加载优化，实际范围与结果见下文。

## 起始基线的生产构件

| 构件 | 657基线源码与行为 | 本轮处理 |
|---|---|---|
| 生产 Swarm 装配 | `cmd/server/relay_swarm_assembly.go` 将同一 SQLite、Workspace 批准服务、Registry、prepared input reader 和对象存储注入 `NewSwarmService` | 账本和真实授权已装配；是否能执行仍由每次能力和权限检查判定 |
| 工作项生长 | `internal/swarm/app/commands.go:AddWorkItem` 允许 human/agent 经 `manage_work_items`、当前 Grant、Mission 版本及预算限制创建工作项 | 固定来源必须来自批准；写域不可扩大；依赖必须是同 Mission 的已有工作项；新 ID 由服务生成 |
| 当前任务查询 | `ReadMission` 经 `WithMissionRead → ReadApproval → projectAccess` 核当前项目及Agent的read_context，返回工作项、当前租约、产物、执行历史和真实 outbox 时间线；该链路没有核read_swarm | 新图查询增加用途明确的read_swarm读取授权，在同一事务核当前Settings和成员Grant；不直接复用旧检查宣称权限已足够 |
| 执行权威 | `SwarmNativeBridge` 使用原 Registry 的精确 session/input operation，核查原绑定的 controller/startup/native 元组 | 执行结果、原控制器持有、进程退出分别处理；无观察不能宣称死亡 |
| 已有停止 | `PauseMission`、`CancelMission`、`RecoverMission` 使用 `ExecutionOwnership.Stop/Reconcile`，只有确认退出和可结算用量才能释放原租约 | 停止失败或未知用量保留原预留和写域；旧 fence 失效；不创建替代执行者 |
| 已有停止限制 | Mission 具有截止时间、总工作项、尝试、调用/Token/金额及无进展限制；原生控制器具有实际 deadline | 不引入第二份 energy 或调用预算；无进展和预算只能使用实际已保存事实 |
| 原生限制 | portable member evidence 当前 `ToolsKnown/TargetsKnown/CallsKnown/TokensKnown` 保持 false；当前 Codex 文件工具关闭条件未满足 | S3 Claim、真实模型规划及真实 Worker 执行正向验收保持阻塞；不替换后端或伪造能力 |
| 前端 | `SwarmPage` 保留已有任务和时间线；协作入口仍为禁用的 `DesignGraph`；统一 `useReadApi` 支持 refresh、abort、stale | 本轮启用真实 DAG，继续使用统一查询与现有任务详情入口 |

该表记录657基线的源码核查，后续实际服务与测试结果见本文验收记录。历史报告中的成功不作为本轮正向模型生成或Worker生命周期证据。

## Morphogenesis 参考

只读参考 HEAD 为 `ddd2849afe330bf221d405dc2b91948c2416bd47`。已读取 `WorkerConfig`、`Worker.run()`、`TaskLedger.enqueue`、`TopologyEngine.advance_idle_window/prune`、`docs/tracks/topology.md` 及 README。

Worker 的 cycle/claim energy、sense 次数、实际运行时限、连续空闲和 budget sleeping 均有明确停止条件；未知发送先保留，不能用本地能量耗尽掩盖。任务账本在同一事务中固定依赖，要求依赖已经存在，不允许改写既有任务的身份与依赖。

连接裁剪同时要求低权重、明确推进的空闲窗口和最少活动连接，required 连接保留。该规则属于已配置的路由偏好；任务依赖、提交、复核和执行事件作为事实保留。Astrocyte 当前没有该类路由权重及反馈契约，本轮不会将任务依赖改成可衰减权重，也不会因页面隐藏节点删除事实。

## 已批准的实现设计

### 任务 DAG 与来源

增加只读 `GET /api/v1/missions/{id}/dag`，在当前项目和read_swarm授权下复用 `ReadMission` 的一致事务。DTO 包含 `schema_version`、`mission_id`、`mission_version`、服务端 `observed_at`、工作项节点、依赖边、Worker 参与记录及单独的执行/协作事件列表。

节点为真实 WorkItem ID，边方向为 dependency → dependent。新工作项保存实际创建主体、主体类型和创建时间，随 `swarm.work_item_created` 保存到同一 outbox 事务。旧记录缺少创建主体时明确显示来源未记录；不能推断为模型生成。Agent 凭据创建只能证明 Agent 主体，真实模型生成还需原生输入/输出及会话来源，本轮报告分别计数。

Domain 验证拒绝重复节点、自依赖、重复依赖、缺失端点和环。AddWorkItem 在保存前验证合并后的完整图；只读 DTO 也验证历史图，损坏时立即报错，不跳过边。既有“只依赖已有项”的写入规则继续保留。所有依赖为 required 事实；协作往返放在另一个列表，不加入拓扑排序。

每个节点和依赖给出实际 `swarm.work_item_created` 的 event_id、aggregate_version 和发生时间。执行/复核列表继续返回原 operation/session/binding 及真实事件；图详情保留原任务、上下文和产物入口。无来源时明确缺失，不补造事件 ID。

### Worker 参与与退出

Worker 身份使用真实 session 与 binding，operation 为参与记录的执行身份；相同 Worker 的多次执行保留各次来源。当前 lease 的存在、到期、原生观察、回合完成、确认退出和资源释放分别显示。租约到期仅表示不能持有当前写权限；submitted/accepted 不代表进程已退出；服务重启丢失内存权威时显示无法核实。

补充用途明确的只读参与观察端口，由生产 `SwarmNativeBridge` 读取现有 Registry evidence。它只访问已由当前 Mission 执行历史确定的 tuple，并核对 project/session/operation；不扫描配置或会话目录，不启动或接续原生，不进行模型调用。明确缺失的原控制器返回 `unverifiable` 与能力原因；身份不符、数据库错误等立即失败。

持久退出记录只能来自实际wait/stop确认。新观察独立保存 `execution_exit_observed` 及原tuple，实际结果与用量另行验证；现有 `execution_reconciled` 保留其退出和释放语义。原代码写对账事件前已要求StopConfirmed，历史负载缺少实际退出时间，旧记录时间只标为对账事件时间。当前Registry的ExitObservedAt没有实际赋值，不用读取或记录时间代替进程退出时间。

停止机制继续使用当前 Mission stop limits、原生 deadline 和原有 Stop/Reconcile。主控在 `msg_d4d7958cd78a` 明确采用服务生命周期内单一有界检查，具体行为如下。

1. 使用现有 `workerCtx`，启动时检查一次，之后每5秒检查一次；每轮串行读取现有 Mission。仅处理 ready/running，核其已批准 deadline、已有无进展轮数/时限和能准确判定的已结算资源耗尽。资源缺少已知值时保持未知；不能把全部预留等于上限当作正在执行的任务已耗尽，也不能用响应次数替代调用预算。
2. 对每个触发 Mission，在同一 SQLite 写事务内重新核 Mission 版本、状态和条件，并取得该 Mission 的当前全部原租约；先将 Mission 设为 blocked，保存准确停止原因，给当前工作项增加 fence 并保存 blocked，写入 `swarm.mission_stop_requested` 与按原 operation/session/fence 关联的 `swarm.execution_stop_requested`。事件来源固定为内部 `system:swarm-lifecycle`，不借用 human/agent，也不授予新授权。
3. 事务提交后串行调用原 `ExecutionOwnership.Stop`；不持有写事务等待外部进程。每次结算重新核对原 operation/session/lease fence，复用 `settleExecutor`。确认进程退出与可结算用量同时满足后才释放原租约；退出与资源释放分别投影。触发时无租约只保存 Mission 停止事实，不制造 Worker。
4. 明确控制器/input缺失、匹配原tuple但停止未确认/结果未知，或已知部分合法的未知用量，在blocked与预留成功保存后使用专用保留类型。不会重复Stop或释放原预留，下一个tick不重试。结构缺失、已知超预留、身份/结果矛盾、数据库或未分类错误立即返回主服务，触发已有失败/关闭流程。每轮没有派发、认领、模型调用或后端切换。
5. 服务关闭时取消计时及后续检查，等待唯一检查任务结束。事务已经固定停止意图后发生取消或崩溃，Mission 仍为 blocked，原租约保留；下次启动不会重放 Stop。原生控制器仍由已有关闭路径处理，不宣称关闭已成功。用户通过现有显式 Recover 对账原 operation，核当前授权和版本；停止条件仍成立时保持 blocked。
6. 内部检查只在 Go 应用服务向生产装配开放，不注册可被 HTTP/token调用的系统身份入口。GET DAG 仅读取，不推进窗口、改变状态或终止进程。停止前后的实际事件和历史节点永久保留。

当前没有已启用的自主 Worker 规划循环，也没有可据以裁剪的路由偏好，完整“模型规划→生长→逐步退出”的真实正向链路仍受原生和 S3 门禁约束；此范围不得由只读图验收替代。主控已在 `msg_efc179794ddc`、`msg_7b09c2ce3101` 核准代码范围，并在 `msg_87d3f2ef04be` 追加明确观察缺失的三处来源路径。

### API、前端与动态更新

新增 DAG DTO 与端点写入现有 OpenAPI，使用官方生成脚本更新 schema。现有精确Agent路径分类在 `relay_scoped_workspace.go`；只增加 `GET .../dag`，不放宽其他方法。

新增 `MissionDAGReadAuthorizer.WithMissionDAGRead` 端口，由现有 `SwarmApprovalBridge` 在同一Workspace/Swarm事务实现：先读取实际Approval以确定项目，保留现有projectAccess；再加载当前Project，以 `ValidateProjectActionScope` 核当前Settings的read_swarm，对Agent加载当前ProjectGrant并使用 `AuthorizeProjectActionCaller` 核同项目、read_swarm、过期和撤销。human使用原认证会话并核当前项目Settings，读取过期Mission批准的历史不要求执行授权重新生效。该读取不调用执行资源预留检查，也不因为历史预算耗尽拒绝读取；不授予新写权限。添加工作项继续调用原端点。

前端在原蜂群页提供“任务依赖”视图和 Mission 选择，初始仍可看现有任务与时间线。采用 `@dagrejs/dagre` 进行布局，布局不计算业务状态；依赖安装前读取其当前官方包资料。方向切换、缩放、节点键盘选择和依赖详情均只改变显示。保持节点 DOM 身份，更新后保留可用选中项；所选节点不存在时明确说明。

复用 `useReadApi` 对选中 Mission 每5秒读取，变更提交后立即刷新；取消已过时请求，离开页面停止计时。服务端时间和任务版本可见；请求失败保留旧图并明确 stale，提供重试；重连后重新取得完整快照。不将轮询当作 Worker 心跳或死亡判定。空图显示真实空态；权限撤销后不继续展示旧项目内容。图与 Worker 历史可独立选择，退出记录不会从历史消失。

### 已批准的精确 write_paths

| 写域 | 用途 |
|---|---|
| `AGENTS.md`、`docs/standards/dynamic-dag-20261019.md`、`.gitignore` | 当前已授权；同步最新第3节、设计和实际验证；`.verification/` 已被忽略，无需重复修改 |
| `docs/SPEC.md` §18.2、`docs/TASKS.md` 当前开发流程入口 | 同步主控已确认的唯一Worker、独立Reviewer和hook规范 |
| `internal/swarm/domain/dag.go`、`dag_test.go`、`mission.go` | DAG 完整性、真实创建来源及当前限制原因 |
| `internal/swarm/app/dag.go`、`dag_test.go`、`contracts.go`、`commands.go`、`lifecycle.go` | DTO、原授权下的快照、写前图验证、参与观察端口与原生命周期事件证据 |
| `internal/adapters/sqlite/swarm_history.go`、`swarm_native.go`、`swarm_approvals.go`、`swarm_dag_test.go` | 从真实 outbox 和 Registry 读取来源及参与，同一事务核当前read_swarm权限，实际 SQLite 验收 |
| `internal/workspace/domain/relay_member_evidence.go`、`internal/adapters/agents/relay_facts.go`、`relay_input_operation.go` | 仅明确缺少原控制器或未接受原输入时使用带原错误的类型；身份冲突原样失败 |
| `internal/adapters/httpapi/swarm_missions.go`、`swarm_dag_test.go`、`relay_scoped_workspace.go` | DAG GET 注册与精确 Agent GET 分类 |
| `cmd/server/relay_swarm_assembly.go`、`swarm_lifecycle.go`、`swarm_lifecycle_test.go`、必要 `main.go` 接缝 | 同一生产端口装配及经批准的停止检查；不提供新执行后端 |
| `contracts/openapi.yaml`、必要 `contracts/examples/swarm-*.json` | 新 DTO 和端点 |
| `web/src/api/schema.d.ts`、`web/src/api/swarm.ts` | 生成 schema 与统一客户端 |
| `web/src/hooks/useReadApi.ts` | 增加可选的错误保留规则，DAG在401/403/404时清除已撤销或不再存在的数据；其他调用保持原行为 |
| `web/src/pages/swarm/dag/**`、`web/src/pages/swarm/SwarmPage.tsx` | 布局、真实查询、键盘和详情 |
| `web/package.json`、`pnpm-lock.yaml` | Dagre 锁定依赖 |
| `web/e2e/dynamic-dag.check.ts`、`web/dynamic-dag.config.ts` | 实际 HTTP / SQLite 双视口、断连及键盘验收，独立显式配置；默认 CI 不改变 |

代码修改限定在已批准路径。实际 Agent 生成的执行窗口仍受原生权限门禁约束；不把未回答的权限当作可用。

## Reviewer 决策通知 hook

主控追加授权的唯一入口为 `scripts/review-decision-hook.mjs`，使用项目现有Node.js。Reviewer 每次实际形成静态或运行决策后，使用自身当前 Run/Task/Dispatch/发信终端、准确40位候选 SHA、`accept` / `revise` / `insufficient_evidence`、实际非空报告绝对路径和简短依据调用脚本。

脚本通过标准parseArgs验证必填参数，通过真实文件状态核非空报告，再用spawnSync(shell:false, windowsHide:true)和参数数组调用本机已安装orca.exe。单次发送type=handoff、subject=ReviewDecision到指定run:主控邮箱，使用typed Task/Dispatch/report-path参数。正文保留候选、决策、审查身份、报告与依据，输出并解析原生JSON回执。调用最多等待60秒，失败或未知发送立即返回且不自动重试。Windows默认使用本轮已核实的安装路径，明确要求orca.exe；macOS/Linux必须传入当前会话已选CLI绝对路径，Linux拒绝裸orca名称；未知平台立即失败。没有守护进程、定时任务或另一份调度状态。

发送成功只表示持久入队及原生唤醒请求；主控收到并回复后才记录实际唤起。Worker 不替 Reviewer 发送接受或退回决定。本轮 hook 的正向通知由独立 Reviewer 使用真实决定执行。

历史验证证据：原PowerShell入口的AST语法与缺参数/缺报告拒绝检查通过，原件为本树 `.verification/hook-missing-parameter.log` 和 `hook-missing-report.log`；Reviewer真实调用被本机签名策略以UnauthorizedAccess拒绝，未调用Orca。Node入口由主控在 `msg_33fccc4f5eed` 明确批准，Reviewer已在16dd0ed8使用真实静态决定发送 `msg_5e5595bbaf68`，主控实际收到并回复，确认见 `msg_67b07d0e3b25`。没有修改执行策略或代发审查决定；最终候选继续由Reviewer实际运行hook。

## 可运行验收与阻塞

编译、测试与服务使用主控独占串行窗口，双采样 CPU<70%、可用内存≥4GiB，间隔2–5秒；BelowNormal；Go `GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`；Node 512MiB；浏览器单 worker，最多一个自有服务。仅复用 D 盘已有缓存。原件放本树已忽略 `.verification/`。

1. Domain 参数测试核查重复、环、缺端点和一致节点来源，不使用伪执行者。
2. 独立 SQLite、生产 HTTP 服务、Agent 创建的批准公开 Git 项目，使用真实人类会话创建批准/Mission/工作项；核查新增节点/边、幂等、范围、跨项目拒绝与真实 SQL event_id。人工创建明确标为人工。
3. 双视口1920×1080、1280×720，在同一真实服务核查更新、方向/缩放、详情、键盘、空态、长标题、HTTP断线与重连、stale和权限拒绝。禁止截图、视频、trace或像素读取。
4. 退出与资源释放分别核对实际 Registry/真实事件；没有合法执行来源时只验空与拒绝路径，不构造会话、租约或退出事实。
5. 主控批准且原生和 S3 门禁满足后，以同一单后端运行真实模型生成、认领及停止，再逐项回查真实会话与事件；门禁未满足时完整正向链路保留未验。
6. 前端 typecheck/lint/build、契约生成一致性、架构检查及适用 Go 测试/vet/build；原前端、可选 OpenViking、批准与作品入口定向回归。完整默认 CI 单列，既有失败、警告和未执行项保留。

基线与候选使用相同公开来源和资源条件，记录真实节点/边、请求次数、响应时间和实际资源；未测不写性能比例。Reviewer独立给出接受、退回或证据不足。本轮实际检查结果列于下节；真实模型生成、Worker退出及资源释放正向链路未验。

## 已落实的证据与动态读取规则

Worker退出确认固定关联首个真实 `swarm.execution_exit_observed` 或 `swarm.execution_reconciled` 事件，不从后来验收事件取时间。真实StopConfirmed先独立保存，回合Outcome和用量另行验证；使用原认领事件ID和现有Consume在同事务去重，核原operation/session/mission/project/binding/holder/fence，重复Recover保持首个来源且不改变原Mission版本增量。旧schema v1对账事件按原已强制StopConfirmed的语义保留退出确认；显式false/null/类型异常立即失败。实际退出观察时间缺失时为null，确认记录时间和本次读取时间分别展示。产物提交的原事务已释放租约，投影据该真实事件记录资源释放，不据产物状态推断进程退出。

停止原因与StopEvent只取原 `execution_stop_requested`，后续blocked说明独立保存为ReconciliationReason；原出处缺失时显示未知。停止检查仅在blocked及预留保留已成功持久化后产生专用保留类型；明确控制器/input缺失由adapter映射至Swarm消费方类型，结构、身份、结果矛盾、已知超预留及数据库错误不包装，立即返回。实际退出确认、原生输入、资源释放和模型生成分别保留。

`RelayObservationUnavailable` 只在原Registry明确未持有控制器或原input尚未受理时产生，Unwrap保留外部错误码。身份、tuple、startup与退出观察矛盾仍返回原错误。参与投影同时保留当前原生input标记、当前/过期/保留租约和退出/释放事实，不将控制器持有记录提升为完整Agent活性证明。

动态读取每5秒检查一次，在途请求未结束时保留该请求；显式刷新仍可取消旧读取，离开组件清理计时及请求。401/403/404清除旧图，临时断线保留明确stale，重连后取完整快照。图中事实不因空闲衰减或过滤删除。

布局依赖锁定为 `@dagrejs/dagre 3.1.1`，MIT许可，自带类型；仅作布局。安装前读取了[官方仓库](https://github.com/dagrejs/dagre)、[官方使用文档](https://github.com/dagrejs/dagre/wiki)、[官方类型入口](https://github.com/dagrejs/dagre/blob/master/index.ts)及npm registry实际发布元数据。本机发布元数据原件为 `.verification/dagre-package.json`。

## 实际验收记录

全部命令由本树 `.verification/run-dag-check.ps1` 分phase串行执行，原件保留在同一目录。每次启动前两次采样，间隔3秒；已记录的启动采样最高CPU为39%，最低可用内存约7.23GiB。这些数值属于启动采样。BelowNormal、Go单并发/256MiB、Node512MiB、浏览器单worker持续按脚本设置；复用D盘Go缓存和既有依赖缓存。

| 实际命令 | 结果与范围 | 原件 |
|---|---|---|
| `pnpm install --no-frozen-lockfile --prefer-offline --network-concurrency=1 --child-concurrency=1` | EXIT0，7分52秒；复用353、下载并新增2件；镜像manifest慢请求和既有deprecated提示保留 | `dag-install-20261019-013407-579.log` |
| `pnpm generate` | PASS，生成OpenAPI类型 | `dag-generate-20261019-015535-898.log` |
| `pnpm --dir web typecheck` | PASS；最终build再次包含完整类型检查 | `dag-typecheck-20261019-014443-841.log` |
| `pnpm --dir web lint` | 最终PASS，零诊断 | `dag-lint-20261019-020713-704.log` |
| `pnpm --dir web build` | 最终PASS；JS774.32kB/gzip218.13kB；保留大于500kB的分块提示 | `dag-build-20261019-015756-863.log` |
| `go test -mod=readonly ./internal/swarm/domain ./internal/swarm/app ./internal/adapters/sqlite ./internal/adapters/httpapi -run '^TestDynamicDAG' -count=1 -timeout=2m` | PASS，四包均有实际匹配测试，共9个测试函数；包含图规则、合法未知用量边界/错误保留类型、真实Registry缺失分类及原错误Unwrap、确认字段参数，以及真实HTTP/SQLite增长、权限和停止历史 | `dag-go-dag-20261019-020125-101.log` |
| `gofmt -d` 已批准的20个Go源文件/测试文件 | 最终PASS，无差异，未用formatter直接改写源码 | `dag-format-20261019-020112-827.log` |
| `pnpm check:contracts` | PASS，264件既有契约示例与生成一致；13项既有未引用组件警告保留 | `dag-contracts-20261019-015703-102.log` |
| `node scripts/check-architecture.mjs` | PASS，43包满足Domain/app/上下文/传输边界 | `dag-architecture-20261019-014906-953.log` |
| `go vet -mod=readonly ./...` | 最终EXIT0，无诊断 | `dag-vet-20261019-020556-270.log` |
| `go build -mod=readonly ./...` | 最终EXIT0，无诊断 | `dag-go-build-20261019-020846-851.log` |
| `pnpm --dir web exec playwright test --config dynamic-dag.config.ts --workers=1` | 2/2 PASS，1920×1080与1280×720；场景分别51.6秒、46.1秒 | `dag-browser-20261019-020155-785.log` |
| `pnpm --dir web exec playwright test --config observatory.config.ts --workers=1 e2e/observatory.check.ts` | 8/8 PASS，25.2秒；既有页面、批准与作品入口、Ctrl K/清除搜索、可选OpenViking未配置和26件品牌HTTP/decode | `dag-regression-20261019-020434-657.log` |
| `node --check scripts/review-decision-hook.mjs` | EXIT0；正向决策发送由独立Reviewer执行 | `dag-hook-20261019-020955-119.log` |

真实浏览器使用同一生产Go程序、生产静态资源和独立SQLite，工作对象为获准的Agent创建公开Astrocyte项目；工作项由验收脚本经真实human会话创建。实际节点从0增长至2，依赖从0增长至1；两视口均验证了事件ID逐项SQL回查、创建主体、键盘、方向/缩放、长标题与页面横向范围、真实HTTP断线保留stale、同库重启后的相同节点、自动无进展停止、403清图和离页停止轮询。自动停止的任务没有Worker，未产生任何Worker退出事实。

| 视口 | 实际数据目录 | 生长Mission | 自动停止Mission |
|---|---|---|---|
| 1920×1080 | `.verification/astrocyte-s1-Gs4VVK` | `a141dc35-6d4a-43d1-b3c3-e196fc412d97` | `45ef3b18-338e-40af-aa6a-e9412967c665` |
| 1280×720 | `.verification/astrocyte-s1-HOhRgZ` | `68187815-cf98-487f-ba53-6dfa59ee102f` | `ea6b14c8-8bf2-4903-9b29-84b79e427d47` |

每目录保存真实 `dag-before-disconnect.json`、`dag-actual-idle-stop.json`、`data/state.sqlite` 和两次实际服务进程日志。生长快照各2节点/1依赖/0Worker、3个Swarm事件；停止快照各0节点/0依赖/0Worker、2个Swarm事件，state=blocked、reason=no_progress_time_limit。所有返回的event_id均在测试中逐一核对相应真实outbox记录。临时服务均已关闭；结束后的OS检查显示本工作树 `.verification` 下没有存活的server.exe。

两次独立单次DAG HTTP读取分别为27.3973ms与27.6019ms，仅记录当前样本。657首轮JS实测717.36kB，本轮774.32kB；没有性能提升比例、运行期资源峰值或完整性能结论。

首轮typecheck缺少Point参数类型、格式检查差异和首轮双视口批准有效期超出实际停止deadline的失败原件均保留。对应文件为 `dag-typecheck-20261019-014413-526.log`、各次 `dag-format-*.log` 以及 `dag-browser-20261019-015322-928.log`；后者两个实际服务目录为KcHZN6/Lh69gT。后续检查使用正确的实际提案deadline，没有修改业务门禁、断言或添加Skip。默认Playwright配置与CI未改变。

## 真实剩余范围

- 原生文件/命令工具关闭和S3调用预算/能力证据门禁保持。没有调用新模型、启动实际Worker或选择另一执行后端；真实Agent规划、Claim/Send、进程退出及资源释放的完整正向链路未验。authenticated Agent主体与模型生成继续分开。
- 确认字段编码属于参数规则测试；没有构造旧会话、claim、lease、退出或outbox。真实带Worker的过期租约、暂时失联、Stop/Recover并发、未知用量和首个退出来源的系统正向验收缺少合法原生来源，保留未验。
- OpenViking未配置正向服务，真实远端提交/任务完成/未知副作用/检索/分层读取保持首轮未验；本轮通过其未配置回归。完整默认CI未执行、未通过。
- 同一Draft PR继续普通push；没有合并、共享部署、打包或发布。独立Reviewer对准确候选的接受、退回或证据不足另行记录。

## 局部按需加载优化

主控在 `msg_d6a7d9dc26a5` 和 `msg_f3601a2c3d99` 批准按需加载，依据是b0单一JS资源774329bytes，全部页面静态导入DAG/Dagre。前端写域限定SwarmPage、dag目录、现有动态浏览器检查、AGENTS与本报告；必要静态资源修复另在 `msg_c962e8c17fc0` 精确批准 `internal/adapters/httpapi/server.go` 的spaHandler和相邻说明。契约、依赖、锁文件和其他Go保持。Worker维护PR，独立Reviewer发布针对准确Commit的CR。

轻量 `MissionDAGRegion` 在模块顶层声明React.lazy，动态导入原 `MissionDAGView`，仅Swarm挂载该区域时加载；Suspense显示加载状态。外层局部ErrorBoundary保存并显示真实错误，人工“重新加载页面”调用浏览器整页reload，清除已缓存的失败模块Promise。没有自动重试或替代图。lazy身份及区域保持稳定，父组件刷新不会反复重挂载原图，原轮询、选择、权限和离页清理继续由原组件实现。设计已读取[React lazy文档](https://react.dev/reference/react/lazy)及[Error Boundary文档](https://react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary)。

准确b0构建复用实际构建日志 `dag-build-20261019-015756-863.log` 和本树 `web/dist`，在源码修改前复制至已忽略 `.verification/dag-loading-b0-dist`，保留完整产物。比较固定使用657提交的公开README原文与精确locator；两个视口分别为attention和workspace创建全新context，收集全部实际JS响应body的bytes、编码、文件清单，再记录进入Swarm后的新增JS。全构建未压缩与gzip总量分别计算。候选使用同一测量脚本、来源及服务配置，原件在 `.verification/dag-loading-*/actual-js-measurement.json`。

真实缺失分块验收仅移动本树已忽略dist中精确核查的DAG文件，finally原样恢复；不劫持响应。静态服务仅对 `os.ErrNotExist` 且 `/assets/` 前缀的缺失文件返回404和 `Cache-Control:no-store`，已有文件仍由标准FileServer处理，实际目录、其他页面路径与API行为保持。局部区域显示实际导入错误，恢复后由用户在同一context键盘触发普通整页reload，核真实200/JavaScript响应与依赖视图。验收同时检查实际 `/assets`、`/assets/` 和深层页面返回准确构建index正文、API继续返回JSON、非Swarm首次页面、其他导航及无自动重试。直接进入Swarm的原动态增长/选择/权限/断线检查继续执行。

两次最初测量因h1的CSS伪元素影响完整可访问名称而超时，实际JS200、pageErrors为0。失败原件及实际DOM文本保留；后续使用可见h1及准确textContent同时核验。早期HEAD来源样本单列保留；首组固定657来源的b0比较目录为 `.verification/dag-loading-b0a21a2c-1792348172803`，四个全新context均实际加载774329bytes、进入Swarm追加0bytes。

首次真实恢复验收原件 `dag-browser-20261019-023120-033.log` 中，两视口原动态功能均通过，分块缺失检查在恢复后均取得text/html而失败。实际8vcdz2/m0kJ6J目录保存首次缺失200/HTML562bytes和完整服务日志；恢复后的reload没有再次请求该分块，浏览器继续使用HTTP缓存。该实测触发上述经批准的no-store/404修复。没有清除缓存、更换context、修改mtime或删除恢复断言。

最终测量前后均使用最终收窄的Go服务源码，串行提供准确b0静态产物和候选静态产物。固定来源、视口、全新context和资源守卫一致；实际HTTP编码均为identity。以下gzip使用同一 `node:zlib.gzipSync` 默认参数单独计算，Vite构建输出的gzip提示原样保留，两种口径不交叉比较。

| 全量资源测量 | b0静态产物 | 按需加载产物 |
|---|---:|---:|
| 非Swarm首屏全部JS body bytes，每个视口的attention/workspace均一致 | 774329 | 719784 |
| 进入Swarm追加JS body bytes | 0 | 57129 |
| 完整构建全部JS body bytes | 774329 | 776913 |
| 完整构建全部JS gzip bytes，同参数 | 216760 | 217799 |

b0文件为 `assets/index-DpvBzN0-.js`；候选为 `assets/index-DyOZ-u9Q.js` 和 `assets/MissionDAGView-DJuVNnnW.js`。非Swarm首屏减少54545bytes，完整构建增加2584bytes，gzip总量增加1039bytes；没有全站总量下降或运行期性能结论。每个context的全部实际JS响应清单、HTTP状态、编码、长度及构建文件清单均在最终原件 `.verification/dag-loading-b0a21a2c-1792348856045/actual-js-measurement.json` 与 `.verification/dag-loading-lazy-candidate-1792348879495/actual-js-measurement.json`。

| 本项实际检查 | 结果及对应源码 | 原件 |
|---|---|---|
| `pnpm --dir web build` | PASS，业务前端最终源码；719.78kB入口及57.12kB按需分块；原500kB提示保留 | `dag-build-20261019-023028-475.log` |
| `pnpm --dir web typecheck`、`pnpm --dir web lint` | 最终PASS，零诊断 | `dag-typecheck-20261019-024016-677.log`、`dag-lint-20261019-024031-015.log` |
| `gofmt -d internal/adapters/httpapi/server.go`、`go vet -mod=readonly ./internal/adapters/httpapi` | 最终收窄源码PASS，无差异/诊断 | `dag-static-format-20261019-023825-160.log`、`dag-http-vet-20261019-023831-294.log` |
| `pnpm --dir web exec playwright test --config dynamic-dag.config.ts --workers=1` | 中间源码4/4 PASS，两个视口各实际增长与缺失恢复；随后仅收窄目录条件并增加原行为覆盖 | `dag-browser-20261019-023603-868.log` |
| 同上增加 `--grep '任务依赖按需加载'` | 最终收窄源码2/2 PASS，33.8秒；同context真实移走/还原/普通人工reload、目录/深层原index和真实授权API404 JSON；内含实际生产 `go build -mod=readonly -o … ./cmd/server` | `dag-browser-loading-20261019-023932-219.log` |
| `pnpm --dir web exec playwright test --config observatory.config.ts --workers=1 e2e/observatory.check.ts` | 按需前端8/8 PASS，25.9秒；后续只修改缺失assets条件，已受影响路径由上述最终真实检查覆盖 | `dag-regression-20261019-023427-915.log` |
| `.verification/measure-dag-loading.mjs`，`--dist .verification/dag-loading-b0-dist` / `--dist web/dist` | 最终Go服务、相同固定来源；四个新context前后全量比较PASS | `dag-measure-baseline-20261019-024050-953.log`、`dag-measure-candidate-20261019-024114-443.log` |

最终缺失/恢复原件在 `.verification/astrocyte-s1-N9LTJS` 与 `.verification/astrocyte-s1-1FvXPx`，分别含 `actual-static-routes.json`、`actual-missing-dag-chunk.json`、`actual-recovered-dag-chunk.json`、真实SQLite与进程日志。恢复实际响应为200 / text/javascript / 57129bytes，目录和深层路由与真实构建index正文逐字节相等。所有临时分块已原样恢复，结束时OS检查本树验证server存活0；累计112个启动采样最高CPU39%、最低可用内存约7.23GiB。

既有 `TestStaticServing_*` 使用手写静态fixture，未执行、未改动；最终生产HTTP覆盖了实际静态文件、目录、深层页面和API。窄化后的第一轮额外API检查缺少真实human会话而实际403，原件 `dag-browser-loading-20261019-023844-139.log` 保留；随后使用实际签发的同一human cookie，保留404/JSON断言并通过。最终候选的独立CR尚待Reviewer发布，完整原生与OpenViking范围、默认CI未验状态保持。
