# 动态 DAG 与独立效果验收

## 目标、基线与建议

2026-10-19 用户确认动态变化的 DAG、机制下逐渐消亡的 Worker，参考 Morphogenesis；开发规范改为主控建议、唯一 Worker 实现、独立 Reviewer 验证并决定是否接受，退回原 Worker 继续。主控管理基线 `validation/v1-20261012` / `d468e3d71106bd06f9979b213c499c394ce8e89e`；本轮产品基线和同一 Draft PR #2 为 `songconmaisaix31-design/frontend-observatory-20261018` / `657ca1464517a171ed50ddc8fcc787b237c44ad9`。产品集成仍为 `f17a491c`，本轮不合并或部署。

主控建议：以现有 Mission/WorkItem 的有向依赖组成任务 DAG，Agent 在既有批准范围内新增工作项推动结构生长；真实认领、执行、交接和复核事件附着相应节点。Worker 的参与状态依实际执行权威和退出记录变化，保留历史任务、退出原因及事件来源；租约失效或观察中断不能直接宣称进程死亡。参考 Morphogenesis 的有限能量/空闲/预算停止与弱连接裁剪，区分可衰减偏好和永久事实，不新建调度器、Attempt 或完成证明系统。任务依赖保持无环；可能往返的协作关系单独显示，不能为了画成 DAG 丢弃或伪造关系。

Morphogenesis 只读基线 `ddd2849afe330bf221d405dc2b91948c2416bd47`，未追踪 `.workbuddy-ai/` 保持；参考 `swarm/worker_loop.py:1121`、`swarm/task_ledger.py`、`topology/engine.py:151`、`docs/tracks/topology.md` 及当前 README。有限循环按显式 policy 停止，prune 同时满足低权重、空闲窗口和最少连接，required 连接保留；不迁入 fixture/mock 或其 Python 服务。具体端口及生命周期策略由原 Worker 读取现有实现后提交完整设计，主控与 Reviewer 核查再实施；不自选用户尚未授权的原生工具、后端切换或公开任务执行范围。

## 所有权与循环

用户追加确认：各模块 Agent 均可通过 Orca CLI 调用 subAgent，禁止 Aliyun Agent；每次 Reviewer 决策后以 hook 唤起主控继续方案，持续至规划内容验收完成。当前动态 DAG 模块仍由同一 Worker 长期负责；需要子 Agent 时使用互斥 write_paths，逐项登记生命周期，重操作全局串行。Reviewer 的 hook 由 Worker 在 `scripts/review-decision-hook.mjs` 实现，使用项目现有 Node.js 调用 `orca orchestration send` 持久投递和原生唤醒，不新增调度器；主控收到并回复才确认实际唤起。

- Worker 继续使用原产品工作树和分支，同一 PR #2。首先读取本文件、两份交接、SPEC §10/16/17 和前轮报告，提交设计及准确写域 Handoff。最初允许产品 `AGENTS.md` 同步最新规范、`docs/standards/dynamic-dag-20261019.md`；代码写域经设计核查后追加。实现、测试、文档、普通 push 和全部返修由此唯一 Worker 负责。
- Reviewer 为独立 Codex 会话、独立工作树和分支，基于准确657源码建立基线；只写 `docs/reviews/dynamic-dag-20261019.md` 与本树忽略的 `.verification/`，读取全仓及 Worker 的准确提交。先静态核查设计、已有真实接口和适用验收，再按主控分配的串行窗口实际验证候选。不得修改业务、默认 CI、断言或替 Worker 修复。
- 主控维护本文件、AGENTS、QUESTIONS、HANDOFF、STATUS，提出下一项有实际依据的改进，决定资源窗口；Reviewer 对可用性、效果及接受与否负责。未接受时原 Worker 返修，Reviewer 复验；存在真实阻塞时继续独立工作并明确未验范围，不以重复自评证明优化。

## 验收、资源与停止条件

Reviewer 比较基线与候选在相同实际来源下的可用能力：从禁用拓扑到可交互动态 DAG，实际新增工作项后的节点/依赖变化、Worker 退出证据与历史保留、重连/过期/空态、跨项目权限、环与缺端点拒绝、键盘和两视口使用。事件 ID 逐项回查真实事件库；实际生成、进程退出与页面投影分别核查，未取得真实链路不得接受为全功能完成。既有前端、OpenViking可选、审批与作品功能不得回归。性能只记录实际采样与前后差异，不编造提升比例或任意成功线。Reviewer 明确输出接受、退回或证据不足及准确 SHA、命令和限制。

重操作只有一个串行窗口：CPU<70%、可用内存≥4GiB，间隔2–5秒两次采样；BelowNormal；Go GOMAXPROCS=1 / GOFLAGS=-p=1 / GOMEMLIMIT=256MiB；Node max-old-space-size=512、浏览器单 worker、最多一个验证服务。不足时只做静态检查；不关闭用户应用。所有编辑使用文件编辑工具，无 mock/伪 API/伪 CLI/伪原生/新增 Skip；不使用视觉功能或 `/tmp`。临时产物在各自已忽略 `.verification/`，依赖复用现有缓存。

原生文件/命令工具关闭、同 Mission 后端独占、S3 未答、实际作品和自由工具边界保持；没有可验证的真实来源时记录证据不足，不伪造生命周期。提交普通 push、保留清晰阶段、同一产品 PR，完整默认 CI 仍单列未通过。正式 worker_done 后主控立即安排原终端下一任务或官方释放，保留 transcript；不自动合并、部署、打包或发布。

## 实际派发

沿 Run `run_73578b70c449` / generation3 / 主控 `term_d7578194-fd20-42b4-9668-b0b0e8683571`。Worker：Task `task_50ff3b1212c2`、Dispatch `ctx_e5e98fad4345`、终端 `term_fbfa5b1f-a75a-4f4f-a164-e41a0d0450c2`，复用原产品树/分支。Reviewer：Task `task_40fbc675246c`、Dispatch `ctx_5a4a6eff08a2`、终端 `term_df8b920b-ed76-480f-8fe6-b995a71a0652`，独立树 `C:/Users/DW/orca/workspaces/Astrocyte/dynamic-dag-review-20261019`、分支 `songconmaisaix31-design/dynamic-dag-review-20261019`。两者实际 Codex0.162.0 / GPT-6.1-Sol xhigh fast，继承已有设置，无模型覆盖。启动回执仅观察到 input_accepted；文本检查逐一确认任务停留 composer 后，各补一次裸 Enter，随后逐一观察到 Working 和实际读取，没有重复任务或进程。

## 设计核查与代码写域批准

Worker Handoff `msg_3c0b2fa427bc` 的完整设计位于产品树 `docs/standards/dynamic-dag-20261019.md`，主控已完整读取。Reviewer 初查发现 `ReadMission` 当前只核项目访问，不能直接宣称已核 `read_swarm`；新 DAG 对 Agent 须核当前项目、动作和 Grant，人类有权限的历史读取保留。该问题交原 Worker 实施并复验。

主控通过准确问题 `msg_450db3c253c5` 批准以下代码写域：`internal/swarm/domain/{dag.go,dag_test.go,mission.go}`；`internal/swarm/app/{dag.go,dag_test.go,contracts.go,commands.go,lifecycle.go}`；`internal/adapters/sqlite/{swarm_history.go,swarm_native.go,swarm_dag_test.go}`；`internal/adapters/httpapi/{swarm_missions.go,swarm_dag_test.go,relay_scoped_workspace.go}`；`cmd/server/{relay_swarm_assembly.go,swarm_lifecycle.go,swarm_lifecycle_test.go,main.go}`；`contracts/openapi.yaml` 与必要 `contracts/examples/swarm-*.json`；`web/src/api/{schema.d.ts,swarm.ts}`、`web/src/hooks/useReadApi.ts`、`web/src/pages/swarm/dag/**`、`web/src/pages/swarm/SwarmPage.tsx`、`web/package.json`、`pnpm-lock.yaml`、必要 `dependencies.lock.json`；`web/e2e/dynamic-dag.check.ts`、`web/dynamic-dag.config.ts`；`scripts/review-decision-hook.mjs`；产品 `AGENTS.md`、`docs/SPEC.md` §18.2、`docs/TASKS.md` 入口、上述标准报告。不得修改其他业务文件；新增依赖先核官方资料。

停止触发批准为既有 workerCtx 内启动时及每5秒的单一有界核查，仅针对已有 ready/running Mission 原已批准、可准确判定的截止时间和无进展等条件。内部系统来源，无假 human；事务固定停止意图和 CAS，事务外原 Stop/Reconcile，精确原 tuple/fence 后结算。未知结果固定 blocked，保留租约与预留，显式 Recover 前不自动重发；数据库/身份错误立即上抛，关闭与重启保留原记录。不增加 energy、调用预算单位、认领、规划调度、后端切换或新原生权限。

追加批准 `internal/adapters/sqlite/swarm_approvals.go`，新 DAG 读取由当前实际 Project/Settings/ProjectGrant 核查动作，保留 human 已授权历史读取。Reviewer 可在自己的准确源码工作树使用现有忽略的 node_modules、web/dist、tsbuildinfo 与 pnpm 缓存，不另导出源码副本；业务及锁文件保持只读，实际数据库和日志放自身 `.verification/`。

Worker `msg_47e067af6328` 确认当前没有重操作或自有验证服务，明确交出窗口。主控已通过 `msg_c42cfdde9971` 将准确657基线的独占检查窗口授予 Reviewer；Worker 继续静态实现，安装/生成/编译/测试/服务须等窗口明确回交。完整模型生成/执行/退出仍受原门禁约束，图和拒绝路径验证不替代真实链路。

Worker 已普通 push hook/规范/设计提交 `9097404a1e751a5dfd60eb042005efb2071e5704`。首次 Reviewer 实际执行被 Windows 签名策略拒绝，未进入 Orca 发送；主控已将入口改用现有 Node.js 的返修交回原 Worker，授权以文件编辑工具移除该 Worker 本轮的 PowerShell 入口，保留实际失败原件，不改变执行策略。Node 入口使用标准参数/文件/进程库及参数数组，无 shell 或 eval；收到真实决策并回复前不登记唤起完成。Reviewer 的基线安装首采样 CPU86%，在 pnpm 前拒绝，未提高门槛。当前没有产品业务接受结论。

## 独立基线与首个接受决定

Reviewer 准确657基线独立验证已完成，install/build/实际HTTP SQLite和顺序双视口九项通过，人工公开README固定来源，两工作项一条依赖，三个event_id逐项查真实outbox；同库重启历史一致，CSRF/Origin/不存在依赖拒绝无部分写入。基线 DAG GET404、拓扑入口禁用；原生会话和租约均0，不证明Agent生成或退出。原件为Reviewer树 `.verification/dynamic-dag/actual-657ca1464517-1792344688827/result.json` 及 `review-notes.md`，服务和浏览器已停止、49941无监听、tracked树干净。主控通过 `msg_fff1bc28094a` 将窗口回交Worker，Reviewer继续静态。

Node hook 的准确源码为 `16dd0ed82f64f82d1b7e66cbfb6731db002846fb`。Reviewer实际657不足决定以 `msg_5e5595bbaf68` 入队，主控实际读取报告并以 `msg_0a59dd232fa7` 回复，同时交原Worker继续权限、停止事件时间和慢轮询改进。Reviewer随后通过hook给出仅通信构件的接受决定 `msg_da0b5fae07e9`，主控已核收；不再为回执重复生成新的证明决定。动态DAG与真实生成/退出尚未接受，继续原规划。

Reviewer 的WIP发现 `msg_bfee437a23f6` 与 `msg_b56f4c77f3b7` 已交原Worker：旧execution_reconciled缺少新字段时保留其原有退出确认语义，新显式false/null异常拒绝；明确缺失观察与controller/startup/退出矛盾分开。主控批准追加 `internal/workspace/domain/relay_member_evidence.go`、`internal/adapters/agents/relay_facts.go`、`relay_input_operation.go`，只为真实controller未持有/原input未受理包装明确 `RelayObservationUnavailable`，Unwrap原错误，外部错误码和权限不变。身份与矛盾不得吞成普通未知，不解析错误文字，不打开原生工具。当前仍Worker独占重操作、Reviewer静态检查，WIP不作正式接受判断。

交付文档追加允许产品 `docs/standards/frontend-observatory-20261018.md` 同步当前DAG已确认方向与新报告入口，保持657基线记录。Worker发现退出确认不能因用量未知而只留blocked；主控在原lifecycle/history范围批准 `swarm.execution_exit_observed` 保存真实原tuple的独立退出观察，实际退出时间无来源保持null，再按原Outcome/用量/预留规则结算，未释放保持false。新增观察必须维持原Recover版本/CAS/fence和幂等，不能重发Stop、伪造事件或绕过门禁。Reviewer已要求DAG读取和Source同时接受准确退出观察事件，正向真实退出仍未验。

Reviewer后续WIP发现已回原Worker：首次退出确认用真实原claim event_id及现有Consume在同事务去重，后续观察不能改写首次来源；原StopReason和后续ReconciliationReason按真实事件分开，无法追溯时未知；停止检查只能识别已成功持久保留的专用结果，不按通用EvidenceMissing/DeliveryUnknown吞结构或身份错误。Worker已采用消费方明确缺失类型及私有retainedExecutionStop，外部原错误Unwrap保持，受影响检查与最终页面正在串行复验，尚无动态DAG接受决定。
