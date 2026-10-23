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

pi09已通过实际单Worker输出、活动停止、完整退出、资源与租约释放及Controller关闭。当前正式来源 `0177b95140dc28e8c4f3a25a7e9c49f62e841504` 的Linux完整构建、严格安装、装内D12（12通过）和组合服务正常退出均已独立接受，正式安装位于 `/workspace/swa-linux-installs/swa-linux-combined-0177b951-n46-07`。旧正式来源和失败仍保存于STATUS及对应任务；不复用旧Run或请求。

S74aa新N4原件位于 `/workspace/swa-pi-n4-packet-74aa5aa8-20261023-01/` 及对应owned目录。前三Worker ready/input_accepted，第四个consult孙任务已创建但父请求因helper串行处理在120秒外层超时；默认wait另外把正常执行pending误判unknown并过早退出6。随后Grant撤销引起的403是尾部后果。原执行FAIL、正常EOF后serve6/observer1及完整树UNKNOWN均保留。两个余留daemon531361/helper532517已通过原认证接口正常EOF及shutdownIfIdle结束，双pidfd两次退出间隔50.130200ms，补充结果已独立接受；不改数据库资源/租约或原全树UNKNOWN。工具首次模块加载失败及同安装NODE_PATH局部修正均保留。不得重放旧请求或改写原租约。

产品修复 `26f5db17` 的11项真实规则、strict、架构和源码审查通过；提前启动driver已随 `6190b773` 完成新正式Linux构建、安装、D12（12通过）及实际服务检查。I为 `/workspace/swa-linux-installs/swa-linux-combined-6190b773-n46-05`。服务smoke产品记录和原脚本退出0；工具曾把最外层显式exit回传为1，无服务Node对照已复现该采集边界，原两个值均保留，后继使用独立脚本或子shell。唯一无模型RPC的在途EOF仍未验。

S6190新N4四Worker均ready/input_accepted、默认wait在真实pending期间持续active。原packet `/workspace/swa-pi-n4-packet-6190b773-20261023-01` 和owned目录保持FAIL。notify误传text被公开Go拒绝且原R absent；现两消费者统一使用获准input_file和invocation.input.text。另slot3接受用户journal实际存在且为threadscope，现reader按原turn.userItemId唯一关联。两项源码、直接消费者、Go/SQLite和关联规则、合并strict已全部独立接受。该slot原native 403 No approved upstream IPv4 address发生在撤Grant前，其余3有真实输出；当前一次无认证GET返回401仅证明当前端点可达，不解释原403。原driver未到正式overlap、计划停止、业务审阅或GUI。正常EOF退出后serve6/observer1、整树UNKNOWN保持；daemon571091/helper571741已通过原认证正常EOF/idle收尾，双pidfd退出50.142113ms/helper缺席/daemon为退出僵尸，补证获独立接受。原DB四资源与三live租约未改，旧请求不重放。

Saba新N4四个原生启动均完成；通知完整4200字节输入已实际保存且未增Task/Worker。首错是活动trace在增长时被完整文件版本门拒绝，另有投影快照变化；三个成功Pi最终消息中的reasoning项被旧reader误要求assistant。第四Worker原生网络403早于撤Grant，原路径已停止释放，不改判成功。原driver FAIL、owner退出1/serve6/observer1、全树UNKNOWN保留。准确daemon589283/helper589954已经原认证正常EOF/idle结束，双pidfd退出50.110446ms，helper缺席/daemon僵尸；原数据库未改，补证已独立接受。

活动观察改为同FD完整前缀、身份固定和有限只读重采，完整结果继续全版本检查；Pi最终块按原type/index严格对应reasoning或assistant；完整成果预算为单源12MiB、结果96MiB、完整响应128MiB，普通响应16MiB。两组各4项规则、实际四trace和六成功close回归、Go三包及架构通过；首次strict类型收窄错误修正后r2退出0。该批S0177源码及n46-07正式构建、安装、D12和Root服务检查均已独立接受。

S0177新N4四Worker均成功结算，原轨迹存在约2.61秒共同输出；首次来源核验时driver失败，旧catch未保留具体断言，精确历史首错UNKNOWN。四次release晚于首catch，不得归因初次观察时资源已经released。已用原件复现确定拒绝条件：custody.owner是runtime609743，driver却比较daemon609757；生产来源与独立ready/live-proof均证明两者不同。托管者修复2c0fa0f0八源及真实ready/普通Node进程链、严格类型与适用局部检查已独立接受并推送；它未运行新SWA/model。n46负责人继续已独审的时序计划c3b966d0，只改原driver及必要边界规则：逐启动完成观察、固定槽位、收齐在途操作、复用完整envelope、提前完整quota、保最新活动stop检查、notify移后。未解锁新完整构建或模型，停止隔离、业务审阅和GUI未完成，N6未运行。原owner退出1/serve6/observer1、全树UNKNOWN不变；准确daemon609757/helper610275已原认证正常EOF/idle结束，双pidfd50.112896ms、helper缺席/daemon僵尸，补证已独立接受。原packet/owned/数据库均保留，详见STATUS。安全失败位置记录35d45a5c九源及两版Node各4项、Go authorizer/依赖检查已独审接受，尚未用于新模型；后继authorizer需重新编译并记录新来源。GUI报告流式读取工具源码01895af及7项局部检查已独审通过，实际浏览器仍未运行。

可恢复副本清理记录均在STATUS。最近实际移除Saba package及旧I6190/I74aa三根，净增2989182976字节并获独立接受。后继准确四根（Saba candidate、S0177 package、Iaba、I4c）的恢复核验/check-only已接受，但删除被自动审批在进程创建前拒绝，等待用户明确确认准确路径；四根仍在、实际释放0，不重试或换入口。所有tar/备份/数据/认证/证据及当前I0177保留。检查会重建Go缓存，下一正式构建各阶段必须重新过实际空间门，不据旧剩余量推定可用。

Windows D10原FAIL/unknown649与D11的FAIL/unknown5保留。D11 SDK7项、TTY/pipe各三次ready及自然EOF、真实Controller打开和关闭通过；只读取证证明原5项全部OpenProcess87且创建时间UNKNOWN，不能猜角色或原因。D12诊断DLL编译失败，原日志未保具体MSVC首错；D13实际取得psutil头三项C2371。局部include适配及D14单组件编译已独立接受，准确DLL SHA和限制见STATUS。D15来源a11a2126的ABI两源/local已独审，唯一tag已推送，run39990878968/job120541997009执行中；仅必要单组件编译→原self→真实rootchild/正常EOF/native/CIM后停。Windows owner收集原结果并交原Reviewer，不自动重试或开启SDK/Go/A/B/Controller/服务/model，不升级D11未知。Root串行决定新CI。[草稿PR #6](https://github.com/songconmaisaix31-design/swarm-anything/pull/6)保持草稿。

自动完成仅表示执行结束，验收状态仍为 `unreviewed`。父任务恢复须证明原生输入被实际接受、父 Agent 继续产生输出；持久邮件已排队不代替实际恢复。N6 保留 lead、停止另一 Worker，并通过真实孙任务成果触发父任务继续。所有子任务、未决输入和完整资源释放均纳入最终 Run 状态。

工具链先 `source .verification/cloud-setup/activate.sh`。临时结果放项目忽略目录，浏览器使用短 TMPDIR。资源限制已由用户解除，按 [RESOURCES](docs/RESOURCES.md)执行，不等待旧单窗口规则。密钥仅供实际原生消费者读取，不打印、散列或提交。StepFun、公开四文件项目、ToolsOff、Windows/Linux 与不限预算的当前决定见 [QUESTIONS](docs/QUESTIONS.md)。

旧N=3持有服务已正常关闭，所有旧 N=3 包已失效。新 N>3 请求已授权同范围必要测试，不再索取逐来源 SHA 的人类确认。当前平台基线和未验限制只查 STATUS；更多历史保留在[原交接文件](https://github.com/songconmaisaix31-design/swarm-anything/blob/11619d72f08c163f159a4f4befe270806b1b67ea/HANDOFF.md)。
