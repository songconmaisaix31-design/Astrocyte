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

默认检查、清理及 Linux S50/S4c 安装和 D12 已通过。S4c 的实际 N4 在长任务消息处失败，未启动 Worker；原请求及关闭结果见 STATUS。修复 `36fb24a9` 已通过正式分段 writer 的短/322KB长/短消息、Ctrl-D、pipe EOF 及 strict 检查。后继 r3 使用新数据和新请求，通过真实 Go/RPC/SQLite 任务创建、同 helper 后续响应、关闭与授权撤销、serve/observer 退出0及两次完整所属进程为空；永久回归入口消费该 completed invocation，TTY/pipe 的三次响应和自然 EOF 均通过。旧失败和 UNKNOWN 原件保留，开发集成不代替新安装验收。Windows 文件名修复 `06495931` 及真实任务输入接线 `ab689087` 已局部验证、独立审查并推送。下一步固定源码构建新的 Linux/Windows 完整候选，再继续 N4/N6、同 Run GUI 和逐平台核验。原请求不重放，已结束服务不复用。每次失败先向用户交代最小复现、根因证据及局部结果，再启动下一次完整构建。

自动完成仅表示执行结束，验收状态仍为 `unreviewed`。父任务恢复须证明原生输入被实际接受、父 Agent 继续产生输出；持久邮件已排队不代替实际恢复。N6 保留 lead、停止另一 Worker，并通过真实孙任务成果触发父任务继续。所有子任务、未决输入和完整资源释放均纳入最终 Run 状态。

工具链先 `source .verification/cloud-setup/activate.sh`。临时结果放项目忽略目录，浏览器使用短 TMPDIR。资源限制已由用户解除，按 [RESOURCES](docs/RESOURCES.md)执行，不等待旧单窗口规则。密钥仅供实际原生消费者读取，不打印、散列或提交。StepFun、公开四文件项目、ToolsOff、Windows/Linux 与不限预算的当前决定见 [QUESTIONS](docs/QUESTIONS.md)。

旧持有服务已正常关闭，所有旧 N=3 包已失效。新 N>3 请求已授权同范围必要测试，不再索取逐来源 SHA 的人类确认。当前平台基线和未验限制只查 STATUS；更多历史保留在[原交接文件](https://github.com/songconmaisaix31-design/swarm-anything/blob/11619d72f08c163f159a4f4befe270806b1b67ea/HANDOFF.md)。
