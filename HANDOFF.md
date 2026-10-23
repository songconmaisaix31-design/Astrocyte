# 接续说明

先读 [STATUS](STATUS.md)、[本轮任务](tasks/swa-evolution-20261023.md)和 [AGENTS](AGENTS.md)。当前工作目录 `/workspace/swarm-anything`，分支 `integration/swa-task-cards`；用户要求持续完成产品拆分、完整 Run 入口、能力状态、角色/交接与真实 N>3 测试。

用户已确定 Astrocyte 和 Attention Mirror 分别成为独立仓库。两仓已有独立构建与实际本地入口，准确提交和验证范围见 STATUS。SWA 保留 Workspace/Swarm/原生执行及必要显式集成；旧混合源码按准确保存点移出，来源、许可、研究数据与公开测试输入保全。真实用户数据库不自动搬迁。最终同提交安装和真实 N4/N6 仍须完成。

当前负责人及互斥写域：

- separation：Attention Mirror 独立仓、显式提取器配置与资料适配器测试，当前已独立审查接受；任务 `swa-product-separation-20261023.md`。
- capabilities：能力投影、角色/讨论 GUI、SWA 专用 Go server 与 HTTP/SQLite 闭包，已完成局部验证及依赖核对。
- run_cli：CLI `run/wait/export`、同 Run GUI、独立前端及默认检查、构建和启动。
- collaboration_model：原生角色/讨论持久化、交接、递归 Run 观察与原 journal 完成/资源回收；Astrocyte 显式事件导入已完成独立交付。
- release_audit：已审查的 N4/N6 驱动、候选构建/独立安装与平台验收，当前负责真实无模型RPC/TTY局部验证及可恢复副本空间计划；其Windows负责人独占Windows诊断工作树。
- n46_execution：当前正式安装的N4/N6实际执行、原件保全和准确余留资源正常收尾。
- acceptance_audit：独立计划、源码及实际结果审查，记录 `swa-evolution-review-20261023.md`。
- Root：Go 组合提交及内部效果前准入、公共 command/invocation 契约、生成物、Astrocyte 工作台、规格/状态、串行提交和整合；独立审查非本人实现的 Attention Mirror。

共享工作树已有用户授权，禁止覆盖其他负责人的未提交改动。按准确路径分批提交，普通 push；不要 `git add .`、重写历史或修改原安装。前端/后端生成类型由 Root 串行维护。

默认检查、拆分清理及此前Linux安装均已通过；长报文、工作树参数、Pi选择/结构化模式、执行文件探测、输入绑定和停止关闭路径的修复及各轮原失败保留在STATUS与对应任务。旧请求不重放，已结束服务不复用，正式安装不改作开发副本。每次失败先向用户交代最小复现、根因证据及局部结果，再启动下一次完整构建。

pi09已通过实际单Worker输出、活动停止、完整退出、资源与租约释放及Controller关闭。后继正式来源 `74aa5aa84bb64cccade786973a8d0c7b402b67fc` 的Linux完整构建、严格安装、装内D12（12通过）和组合服务正常退出均已独立接受，正式安装位于 `/workspace/swa-linux-installs/swa-linux-combined-74aa5aa8-n46-04`。

S74aa新N4原件位于 `/workspace/swa-pi-n4-packet-74aa5aa8-20261023-01/` 及对应owned目录。前三Worker ready/input_accepted，第四个consult孙任务已创建但父请求因helper串行处理在120秒外层超时；默认wait另外把正常执行pending误判unknown并过早退出6。随后Grant撤销引起的403是尾部后果。原执行FAIL、正常EOF后serve6/observer1及完整树UNKNOWN均保留。两个余留daemon531361/helper532517已通过原认证接口正常EOF及shutdownIfIdle结束，双pidfd两次退出间隔50.130200ms，补充结果已独立接受；不改数据库资源/租约或原全树UNKNOWN。工具首次模块加载失败及同安装NODE_PATH局部修正均保留。不得重放旧请求或改写原租约。

产品修复 `26f5db17` 的11项真实规则、strict、架构和源码审查通过；提前启动driver已随 `6190b773` 完成新正式Linux构建、安装、D12（12通过）及实际服务检查。I为 `/workspace/swa-linux-installs/swa-linux-combined-6190b773-n46-05`。服务smoke产品记录和原脚本退出0；工具曾把最外层显式exit回传为1，无服务Node对照已复现该采集边界，原两个值均保留，后继使用独立脚本或子shell。唯一无模型RPC的在途EOF仍未验。

S6190新N4四Worker均ready/input_accepted、默认wait在真实pending期间持续active。原packet `/workspace/swa-pi-n4-packet-6190b773-20261023-01` 和owned目录保持FAIL。notify误传text被公开Go拒绝且原R absent；现两消费者统一使用获准input_file和invocation.input.text。另slot3接受用户journal实际存在且为threadscope，现reader按原turn.userItemId唯一关联。两项源码、直接消费者、Go/SQLite和关联规则、合并strict已全部独立接受。该slot原native 403 No approved upstream IPv4 address发生在撤Grant前，其余3有真实输出；当前一次无认证GET返回401仅证明当前端点可达，不解释原403。原driver未到正式overlap、计划停止、业务审阅或GUI。正常EOF退出后serve6/observer1、整树UNKNOWN保持；daemon571091/helper571741已通过原认证正常EOF/idle收尾，双pidfd退出50.142113ms/helper缺席/daemon为退出僵尸，补证获独立接受。原DB四资源与三live租约未改，旧请求不重放。

S74aa候选与解包两副本、开发副本与六份官方Node解包输入已按独立完整恢复核对清理。S6190 candidate/package及旧I50/I7ff四根也经逐成员恢复/引用/check-only独审后执行唯一删除，operation-01记录0/四根已删、remaining空，净增加3403153408字节；I4c和当前I6190、所有tar/备份/数据/证据保留。下一步固定已接受源码，按原阶段空间门进行新n46-06正式构建、安装、D12与Root服务检查，再以新packet进行N4；N4完整通过后才N6。

Windows D10 `1b8828d6` / run `39952528075` 仍为FAIL/unknown649；`3ffa2774`只读取证成功，不能清除原UNKNOWN。D11 `eb7a1746` / run `39963000754` 的SDK7项、TTY/pipe各三次ready及自然EOF、真实Controller打开和关闭均通过；整体fixture6、serve0、关闭latch=true/unknown5，完整树仍未通过。`79639b18`、evidence-04 tag的只读run `39967710170` 已SUCCESS，原5项全部OpenProcess 87、创建时间UNKNOWN，分为2项当前root身份未验证和3项父身份未验证；此次root行存在，原predates排除条件不适用。原实际pure overlay两SHA均匹配，不按错误码猜进程角色或退出原因。Windows owner继续仅基于原件审查最小收集方案，未重跑SDK/A/B/模型；06子独审。[草稿PR #6](https://github.com/songconmaisaix31-design/swarm-anything/pull/6)保持草稿。

自动完成仅表示执行结束，验收状态仍为 `unreviewed`。父任务恢复须证明原生输入被实际接受、父 Agent 继续产生输出；持久邮件已排队不代替实际恢复。N6 保留 lead、停止另一 Worker，并通过真实孙任务成果触发父任务继续。所有子任务、未决输入和完整资源释放均纳入最终 Run 状态。

工具链先 `source .verification/cloud-setup/activate.sh`。临时结果放项目忽略目录，浏览器使用短 TMPDIR。资源限制已由用户解除，按 [RESOURCES](docs/RESOURCES.md)执行，不等待旧单窗口规则。密钥仅供实际原生消费者读取，不打印、散列或提交。StepFun、公开四文件项目、ToolsOff、Windows/Linux 与不限预算的当前决定见 [QUESTIONS](docs/QUESTIONS.md)。

旧N=3持有服务已正常关闭，所有旧 N=3 包已失效。新 N>3 请求已授权同范围必要测试，不再索取逐来源 SHA 的人类确认。当前平台基线和未验限制只查 STATUS；更多历史保留在[原交接文件](https://github.com/songconmaisaix31-design/swarm-anything/blob/11619d72f08c163f159a4f4befe270806b1b67ea/HANDOFF.md)。
