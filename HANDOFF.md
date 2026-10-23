# 接续说明

先读 [STATUS](STATUS.md)、[本轮任务](tasks/swa-evolution-20261023.md)和 [AGENTS](AGENTS.md)。当前工作目录 `/workspace/swarm-anything`，分支 `integration/swa-task-cards`；用户要求持续完成产品拆分、完整 Run 入口、能力状态、角色/交接与真实 N>3 测试。

用户已确定 Astrocyte 和 Attention Mirror 分别成为独立仓库。两仓已有独立构建与实际本地入口，准确提交和验证范围见 STATUS。SWA 保留 Workspace/Swarm/原生执行及必要显式集成；旧混合源码按准确保存点移出，来源、许可、研究数据与公开测试输入保全。真实用户数据库不自动搬迁。最终同提交安装和真实 N4/N6 仍须完成。

当前负责人及互斥写域：

- separation：Attention Mirror 独立仓、显式提取器配置与资料适配器测试，当前已独立审查接受；任务 `swa-product-separation-20261023.md`。
- capabilities：能力投影、角色/讨论 GUI、SWA 专用 Go server 与 HTTP/SQLite 闭包，当前负责最后依赖核对。
- run_cli：CLI `run/wait/export`、同 Run GUI、独立前端及默认检查、构建和启动。
- collaboration_model：原生角色/讨论持久化、交接、递归 Run 观察与原 journal 完成/资源回收；Astrocyte 显式事件导入已完成独立交付。
- release_audit：已审查的 N4/N6 驱动、新候选构建/独立安装与平台验收，当前准备同提交 Linux 和 Windows 组合包。
- acceptance_audit：独立计划、源码及实际结果审查，记录 `swa-evolution-review-20261023.md`。
- Root：Go 组合提交及内部效果前准入、公共 command/invocation 契约、生成物、Astrocyte 工作台、规格/状态、串行提交和整合；独立审查非本人实现的 Attention Mirror。

共享工作树已有用户授权，禁止覆盖其他负责人的未提交改动。按准确路径分批提交，普通 push；不要 `git add .`、重写历史或修改原安装。前端/后端生成类型由 Root 串行维护。

默认检查、清理及 Linux S50/S4c/S7ff 安装和 D12 已通过。当前完整来源 `7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f` 的 N4 已完成真实 Controller 与根任务创建，但 Worker 启动因 `worktree:'current'` 同时携带仅用于新工作树的 `name` 被上游拒绝。原请求保留 pending/UNKNOWN，尚无 Worker 输出；Grant 已撤销，正常 EOF 后服务退出6、观察器退出1，原完整树关闭 UNKNOWN 保留。补充正常收尾已通过原认证接口结束两个准确的 daemon/helper 资源，双 pidfd 退出、helper 缺席、执行退出0，未发送 OS 信号。该补充不替代原完整树观察。原件和 ID 见 STATUS。

参数修复 `3f110ee430387c66ff9e3b253d893e066bed0e86` 和 Pi 显式模型启动修复 `0bf2c3e534adb02e8a5f5789e586cbd9e6b18274` 已推送，局部检查与独立源码审查通过。pi-01 开发单 Worker 已完成 Controller/根任务创建，但实际 terminal 模式与准入 structured 不符，输入派发失败；原请求及子操作 completed，真实 Task/Dispatch/终端已创建，无模型输出证据。Grant 已撤销，原业务 FAIL/完整树 UNKNOWN 保留；补充正常 EOF 和双 pidfd 观察确认本次三个准确资源已退出，未改原数据库释放状态。collaboration_model 的模式修复已过4项规则、strict 和架构检查，准确源码与局部结果已获独立审查接受；release_audit 负责下一次全新请求的单 Worker 检查。正式 I4c/I7ff 不修改，混合开发检查不代表同提交完整安装或 N4/N6。三份可恢复构建副本已清理，保留归档、安装和原件，后续每阶段重读实际空间。

相同 S7ff 的 Windows run `39912010769` / job `120276917189` 在 service-smoke 失败。只读取证 run `39916407634` 已成功，确认原 controller-open CLI 退出6，未进入 task.create；原 artifact 未保留该 CLI 内容，内部错误仍 UNKNOWN。release_audit 的 Windows 负责人准备无需完整构建的最小复现。此前长报文修复 `36fb24a9` 的真实 writer、r3 Go/RPC/SQLite 与永久 TTY/pipe 验证，以及 Windows 文件名修复 `06495931`、真实任务输入接线 `ab689087` 的局部证据继续保留。完成当前失败的根因和局部验证后，再固定新来源构建并继续 N4/N6、同 Run GUI 和逐平台核验。旧请求不重放，已结束服务不复用；每次失败先向用户交代最小复现、根因证据及局部结果，再启动下一次完整构建。

自动完成仅表示执行结束，验收状态仍为 `unreviewed`。父任务恢复须证明原生输入被实际接受、父 Agent 继续产生输出；持久邮件已排队不代替实际恢复。N6 保留 lead、停止另一 Worker，并通过真实孙任务成果触发父任务继续。所有子任务、未决输入和完整资源释放均纳入最终 Run 状态。

工具链先 `source .verification/cloud-setup/activate.sh`。临时结果放项目忽略目录，浏览器使用短 TMPDIR。资源限制已由用户解除，按 [RESOURCES](docs/RESOURCES.md)执行，不等待旧单窗口规则。密钥仅供实际原生消费者读取，不打印、散列或提交。StepFun、公开四文件项目、ToolsOff、Windows/Linux 与不限预算的当前决定见 [QUESTIONS](docs/QUESTIONS.md)。

旧持有服务已正常关闭，所有旧 N=3 包已失效。新 N>3 请求已授权同范围必要测试，不再索取逐来源 SHA 的人类确认。当前平台基线和未验限制只查 STATUS；更多历史保留在[原交接文件](https://github.com/songconmaisaix31-design/swarm-anything/blob/11619d72f08c163f159a4f4befe270806b1b67ea/HANDOFF.md)。
