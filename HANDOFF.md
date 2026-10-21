# SWA 主线接续

准确1209aee399a2c0b0677d7a4dbedc5f87880d14ab已普通push，strict06/07和四native真实前置检查通过，build09实际EXIT0，首次形成完整candidate08，VERSION相同；包内闭包、原生装载和空Job创建/查询/关闭通过。T02安装后装载因找不到picomatch失败，原pnpm Junction被展开为普通目录，依赖解析图改变；原Worker集中修复制语义，旧安装和日志保留，Orca保全通过。认证、后代停止、精简依赖及同安装候选双Worker仍未验。用户已解除资源限制，普通修复直接累计实施复验。原S65980因UTC2026-10-21T12:44:59.500系统重启退出，原库/配置及旧running.json保留；本机Orca29972/runtime8362保持ready。原Run generation2由term_84161ef8-54be-49cd-b699-145ed804f645继续，原Worker ctx_0cf6ac28fd13与Reviewer ctx_066b3ab5f10c保持同Task/树/分支。

先读 [STATUS](STATUS.md)、[AGENTS](AGENTS.md)、[代码规范](docs/CODING.md) 和[资源限制](docs/RESOURCES.md)，再读 [META_PLAN](docs/META_PLAN.md)、[总任务](tasks/swa-mainline-20261020.md)、[M0](tasks/swa-m0-20261020.md)、具体规格与[用户决定](docs/QUESTIONS.md)。开发者从 [CONTRIBUTING](CONTRIBUTING.md) 进入，完整导航见[文档目录](docs/README.md)。

用户以 re.md v2.1 重构主线。M0 从产品 `32ab1eec808a5848aabad40f11133735a1214e55` 接续，仅整合文档；主控独有决定、研究、实验入口和归档来自准确 `cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`。程序来源 S、已恢复服务和 M1 实际结果统一见 STATUS。处理运行服务前重新核对真实 PID、路径、创建时间、所属终端和唯一监听，遵守原库保全及授权。

M0 准确 `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765` 已获独立接受，PR #4 已合入产品分支；主控已普通快进并 push 默认 main。后续开发从 `origin/main` 接续，Orca 默认基准已相应更新。当前 main 工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/main-20261020`；原根目录 validation 分支及实验材料保留。准确来源和逐项处置见[整合清单](docs/integration-audit-20261020.md)及 [M0 报告](docs/standards/swa-mainline-m0-20261020.md)，当前状态由 STATUS 维护。

准确 `fac75c7f963d16c7f112e8e90ad30e36510d4738` 的八份旧调查与独立报告也已接受并普通快进/push main。来源为 [CLI 调查](docs/standards/swa-cli-execution-20261020.md)、[M1 实际记录](docs/standards/swa-m1-contracts-20261020.md)及各自独立报告；保留其当时设计与限定接受范围。当前无界面方案按最新首版任务继续，不以旧提案替代新范围。

首版目标为抽取 Orca CLI 对应执行能力，交付从原有 Agent 或终端调用的可控 Swarm CLI。准确396e完整方案及测试计划已接受，接口按SPEC §21.1，由同一Worker实现。M1调查已结束：原入口本次通过，历史故障未复现，补丁未实施，修复因果证据不足。原v2.1 M2—M6保留后续规划，MCP和完整S3按各自任务处理，入口见 [TASKS](docs/TASKS.md)。默认CI、受限Codex、S3预算与权限证据、经验准入及实际采用继续按原范围验收，未答项不推断授权。

用户进一步明确首版必须发现本机注册的Agent，并用Swarm CLI调用这些真实Agent；使用REA逆向抽取Orca CLI能力，减少运行依赖并保持隔离环境。独立发现/稳定选择、真实输入/模式及控制端恢复按SPEC §21.1.1—§21.1.5定稿，由同一Worker在原写域继续实现。REA当前6a9交付物已有两个限定静态分析，覆盖限制和原件统一见STATUS及首版任务；旧原件保留原版本，不能混用。当前Orca运行及原服务已退出的事实见本页接续段，动态窗口单独交付。

独立发现与稳定AgentSelection按准确740/c3fca056接受，发现和选择不依赖原Orca正在运行；准确d441的N04/N05输入与模式技术计划已获73946fd2接受，准确079的T12计划已获7c6d3973接受。主控完整读取相应准确差异和报告并回复原hooks，技术契约定稿于SPEC §21.1.1—§21.1.3，同一Worker继续实现。工具范围待用户回答，技术计划接受不授予实际启动或模型窗口。主控独立9c四包Go编译及三个传输前函数已实际通过，原Worker未运行旧撤销指令；该检查不代替Worker自测或安装验收。当前原S及控制脚本身份、唯一窗口和实际限制统一见STATUS。

用户要求继续构建、参考REA仓库格式并直接复用Orca CLI源码。[CLI参考](docs/standards/swa-orca-cli-reference-20261021.md)保留实际PE/Rust及20JS的REA结果，原生函数未反编译及部分覆盖保持；[仓库格式](docs/standards/swa-repository-format-20261021.md)记录官方来源及现有路径对应，具体格式范围待答。准确1209的原生短源码编译、完整strict及首个完整包已通过，T12实际覆盖、通用终端及同安装候选双Worker仍未验。Go扩大的旧案例测试因专用真实输入缺失而失败，原件保持；适用编译和三个传输前函数的结果各按真实范围记录。原Reviewer已读真实bundle和动态消费者，八项纯JS仅是待验证精简候选，新增运行数据和开发类型的inventory同步遗漏已交原Worker；后继普通修复继续累计处理。

首版事实源为 [CLI 任务](tasks/swa-cli-v1-20261020.md)。用户已确认固定抽取范围、Orca 无界面执行与控制端身份/CLI规则、多 Worker、失败与资源、安装包的顺序；最终使用安装后的同一候选完成一次至少两个真实 Worker 的任务，验证并发、指定停止及成果获取。MCP、网页重构、完整旧 S3 和全部历史分支清理进入后续任务，旧 S3 保留独立记录。用户要求不妨碍本机Orca CLI并集成产品，产品采用独立后台，准确396e方案及测试计划已接受，SPEC §21.1明确接口与隔离边界。准确30b来源事实及此前提案保留，其Mission/Claim旧待决项按适用任务处理。程序已开始静态实现，CLI动态验收未通过。

M1 由 `task_ceb3b3bb76be / ctx_8d6a6e1c0145` 在 `swa-m1-contracts-implementation-20261020` 工作树及同名分支接续，实际客户端为 Orca / Codex `gpt-6.1-sol`。准确 `987811f417e92cd6f15961e60c279f616e3b21a8` 原冻结安装及未补丁的完整契约命令均 EXIT0，历史崩溃本次未复现。依已接受计划停止补丁实施；程序/依赖未改，T02—T08 未执行，因果证据不足。主控已完成 T09并曾恢复同一 S、原库和原配置，记录PID52040；后继首版窗口及当前服务见STATUS。最终记录 `b112ab9e03d9ea82247e73b5cb578529ca529f9d` 已独立接受并普通push，原hook由主控实际读取回复；本次调查Task已结算、终端released，完整修复保持未接受。

无界面执行由 `task_9f062888e786 / ctx_0cf6ac28fd13` 沿原`swa-cli-headless-20261020` 工作树/分支继续负责，旧Codex进程的退出由系统boot证明。准确396e方案及T00—T15已独立接受并纳入main，原hook已实际读取回复；SPEC §21.1及首版任务明确接口、逐文件写域和迁移028。初始Task的两文档写域由这些后继完整实施授权接续，新Worker须按首版任务执行代码、测试、文档及返修。阶段源码、远端回执及窗口见STATUS，准确代码由原Reviewer Task接续。原静态Task450/ctxdc在重启后两次原生release均为release_unknown，保留未知记录；旧进程退出有boot证明，新Worker资源独立，未关闭其他终端。原件和真实数据库不进入Git或安装包。CLI退出处理、真实驱动及T12由同一Worker负责，T11待答；首版功能及安装双Worker仍未通过。

已有产品原件位于原 Worker 的 `.verification/delivery-candidate-20261020/`；Windows 目录为 `dist/delivery-candidate-20261020/`。原库一致备份为该原件根 `original-01/original.sqlite`；保全、恢复条件与实际命令在[交付报告](docs/standards/delivery-candidate-20261020.md)。备份包含真实数据，仅用于本地获准核查，不能进入 Git 或发布包。

当前执行源码、构建修正和技术计划的准确提交及审查范围见STATUS与首版任务。C01—C09/helper、T12及F16/F17真实验证由同一Worker负责，已定稿接口按SPEC执行。独立Reviewer当前为`ctx_066b3ab5f10c / term_7f2903a4-3597-4407-9905-d65fc541c60e`，保持原审查树、分支和两报告写域，后续邮件只发新Dispatch。本次brief为根忽略目录`.verification/swa-reuse-recovery-20261021.md`；旧review-recovery记录及全部准确报告保留。主控须读取实际源码与正式报告后确认接受范围，重启前局部结论不代替新源码或真实运行验收。

构建接续：build01—08及部分candidate01—07全部保留，build09已形成准确1209完整candidate08。原生编译采用每包私有短源码副本及原ABI/跟踪/分析，产品只复制真实产物。T01通过；T02首次数组参数失败修原验证入口后，真实安装已发现整张pnpm链接图的复制问题，继续局部修复与独立安装，不逐个提升缺包到根目录。必要同模块copy helper写域已授，最终候选须带入实际安装代码及准确来源。首装T02/T03须完成真实human session/CSRF、项目设置、Grant/token流程，完整USAGE已局部静态接受，不代用户批准。原库和原环境/配置继续保全。

科研沿[预印本任务](tasks/research-preprint-20261019.md)、[重启任务](tasks/restart-product-research-20261019.md)和[研究反馈](docs/research/research-product-decisions.md)接续；V4 沿[日运行任务](tasks/M6-V4-daily-operation.md)及[实验目录](docs/experiment/README.md)接续。外部权威锁、原件、正式研究数据、TeX 技术交付和作者许可各按原条件处理。历史管理记录见[归档说明](docs/archive/20261020/README.md)，准确 S 的交接保留在 [Git 原文](https://github.com/songconmaisaix31-design/swarm-anything/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/HANDOFF.md)。
