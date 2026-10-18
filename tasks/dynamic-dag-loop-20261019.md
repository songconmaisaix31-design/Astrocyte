# 动态 DAG 与独立效果验收

## 目标、基线与建议

2026-10-19 用户确认动态变化的 DAG、机制下逐渐消亡的 Worker，参考 Morphogenesis；开发规范改为主控建议、唯一 Worker 实现、独立 Reviewer 验证并决定是否接受，退回原 Worker 继续。主控管理基线 `validation/v1-20261012` / `d468e3d71106bd06f9979b213c499c394ce8e89e`；本轮产品基线和同一 Draft PR #2 为 `songconmaisaix31-design/frontend-observatory-20261018` / `657ca1464517a171ed50ddc8fcc787b237c44ad9`。产品集成仍为 `f17a491c`，本轮不合并或部署。

主控建议：以现有 Mission/WorkItem 的有向依赖组成任务 DAG，Agent 在既有批准范围内新增工作项推动结构生长；真实认领、执行、交接和复核事件附着相应节点。Worker 的参与状态依实际执行权威和退出记录变化，保留历史任务、退出原因及事件来源；租约失效或观察中断不能直接宣称进程死亡。参考 Morphogenesis 的有限能量/空闲/预算停止与弱连接裁剪，区分可衰减偏好和永久事实，不新建调度器、Attempt 或完成证明系统。任务依赖保持无环；可能往返的协作关系单独显示，不能为了画成 DAG 丢弃或伪造关系。

Morphogenesis 只读基线 `ddd2849afe330bf221d405dc2b91948c2416bd47`，未追踪 `.workbuddy-ai/` 保持；参考 `swarm/worker_loop.py:1121`、`swarm/task_ledger.py`、`topology/engine.py:151`、`docs/tracks/topology.md` 及当前 README。有限循环按显式 policy 停止，prune 同时满足低权重、空闲窗口和最少连接，required 连接保留；不迁入 fixture/mock 或其 Python 服务。具体端口及生命周期策略由原 Worker 读取现有实现后提交完整设计，主控与 Reviewer 核查再实施；不自选用户尚未授权的原生工具、后端切换或公开任务执行范围。

## 所有权与循环

- Worker 继续使用原产品工作树和分支，同一 PR #2。首先读取本文件、两份交接、SPEC §10/16/17 和前轮报告，提交设计及准确写域 Handoff。最初允许产品 `AGENTS.md` 同步最新规范、`docs/standards/dynamic-dag-20261019.md`；代码写域经设计核查后追加。实现、测试、文档、普通 push 和全部返修由此唯一 Worker 负责。
- Reviewer 为独立 Codex 会话、独立工作树和分支，基于准确657源码建立基线；只写 `docs/reviews/dynamic-dag-20261019.md` 与本树忽略的 `.verification/`，读取全仓及 Worker 的准确提交。先静态核查设计、已有真实接口和适用验收，再按主控分配的串行窗口实际验证候选。不得修改业务、默认 CI、断言或替 Worker 修复。
- 主控维护本文件、AGENTS、QUESTIONS、HANDOFF、STATUS，提出下一项有实际依据的改进，决定资源窗口；Reviewer 对可用性、效果及接受与否负责。未接受时原 Worker 返修，Reviewer 复验；存在真实阻塞时继续独立工作并明确未验范围，不以重复自评证明优化。

## 验收、资源与停止条件

Reviewer 比较基线与候选在相同实际来源下的可用能力：从禁用拓扑到可交互动态 DAG，实际新增工作项后的节点/依赖变化、Worker 退出证据与历史保留、重连/过期/空态、跨项目权限、环与缺端点拒绝、键盘和两视口使用。事件 ID 逐项回查真实事件库；实际生成、进程退出与页面投影分别核查，未取得真实链路不得接受为全功能完成。既有前端、OpenViking可选、审批与作品功能不得回归。性能只记录实际采样与前后差异，不编造提升比例或任意成功线。Reviewer 明确输出接受、退回或证据不足及准确 SHA、命令和限制。

重操作只有一个串行窗口：CPU<70%、可用内存≥4GiB，间隔2–5秒两次采样；BelowNormal；Go GOMAXPROCS=1 / GOFLAGS=-p=1 / GOMEMLIMIT=256MiB；Node max-old-space-size=512、浏览器单 worker、最多一个验证服务。不足时只做静态检查；不关闭用户应用。所有编辑使用文件编辑工具，无 mock/伪 API/伪 CLI/伪原生/新增 Skip；不使用视觉功能或 `/tmp`。临时产物在各自已忽略 `.verification/`，依赖复用现有缓存。

原生文件/命令工具关闭、同 Mission 后端独占、S3 未答、实际作品和自由工具边界保持；没有可验证的真实来源时记录证据不足，不伪造生命周期。提交普通 push、保留清晰阶段、同一产品 PR，完整默认 CI 仍单列未通过。正式 worker_done 后主控立即安排原终端下一任务或官方释放，保留 transcript；不自动合并、部署、打包或发布。

## 实际派发

沿 Run `run_73578b70c449` / generation3 / 主控 `term_d7578194-fd20-42b4-9668-b0b0e8683571`。Worker：Task `task_50ff3b1212c2`、Dispatch `ctx_e5e98fad4345`、终端 `term_fbfa5b1f-a75a-4f4f-a164-e41a0d0450c2`，复用原产品树/分支。Reviewer：Task `task_40fbc675246c`、Dispatch `ctx_5a4a6eff08a2`、终端 `term_df8b920b-ed76-480f-8fe6-b995a71a0652`，独立树 `C:/Users/DW/orca/workspaces/Astrocyte/dynamic-dag-review-20261019`、分支 `songconmaisaix31-design/dynamic-dag-review-20261019`。两者实际 Codex0.162.0 / GPT-6.1-Sol xhigh fast，继承已有设置，无模型覆盖。启动回执仅观察到 input_accepted；文本检查逐一确认任务停留 composer 后，各补一次裸 Enter，随后逐一观察到 Working 和实际读取，没有重复任务或进程。
