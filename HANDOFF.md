# SWA 主线接续

先读 [STATUS](STATUS.md)、[AGENTS](AGENTS.md)、[代码规范](docs/CODING.md) 和[资源限制](docs/RESOURCES.md)，再读 [META_PLAN](docs/META_PLAN.md)、[总任务](tasks/swa-mainline-20261020.md)、[M0](tasks/swa-m0-20261020.md)、具体规格与[用户决定](docs/QUESTIONS.md)。开发者从 [CONTRIBUTING](CONTRIBUTING.md) 进入，完整导航见[文档目录](docs/README.md)。

用户以 re.md v2.1 重构主线。M0 从产品 `32ab1eec808a5848aabad40f11133735a1214e55` 接续，仅整合文档；主控独有决定、研究、实验入口和归档来自准确 `cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`。程序来源 S、已恢复服务和 M1 实际结果统一见 STATUS。处理运行服务前重新核对真实 PID、路径、创建时间、所属终端和唯一监听，遵守原库保全及授权。

M0 准确 `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765` 已获独立接受，PR #4 已合入产品分支；主控已普通快进并 push 默认 main。后续开发从 `origin/main` 接续，Orca 默认基准已相应更新。当前 main 工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/main-20261020`；原根目录 validation 分支及实验材料保留。准确来源和逐项处置见[整合清单](docs/integration-audit-20261020.md)及 [M0 报告](docs/standards/swa-mainline-m0-20261020.md)，当前状态由 STATUS 维护。

准确 `fac75c7f963d16c7f112e8e90ad30e36510d4738` 的八份旧调查与独立报告也已接受并普通快进/push main。来源为 [CLI 调查](docs/standards/swa-cli-execution-20261020.md)、[M1 实际记录](docs/standards/swa-m1-contracts-20261020.md)及各自独立报告；保留其当时设计与限定接受范围。当前无界面方案按最新首版任务继续，不以旧提案替代新范围。

首版目标为抽取 Orca CLI 对应执行能力，交付从原有 Agent 或终端调用的可控 Swarm CLI。准确396e完整方案及测试计划已接受，接口按SPEC §21.1，由同一Worker实现。M1调查已结束：原入口本次通过，历史故障未复现，补丁未实施，修复因果证据不足。原v2.1 M2—M6保留后续规划，MCP和完整S3按各自任务处理，入口见 [TASKS](docs/TASKS.md)。默认CI、受限Codex、S3预算与权限证据、经验准入及实际采用继续按原范围验收，未答项不推断授权。

首版事实源为 [CLI 任务](tasks/swa-cli-v1-20261020.md)。用户已确认固定抽取范围、Orca 无界面执行与控制端身份/CLI规则、多 Worker、失败与资源、安装包的顺序；最终使用安装后的同一候选完成一次至少两个真实 Worker 的任务，验证并发、指定停止及成果获取。MCP、网页重构、完整旧 S3 和全部历史分支清理进入后续任务，旧 S3 保留独立记录。用户要求不妨碍本机Orca CLI并集成产品，产品采用独立后台，准确396e方案及测试计划已接受，SPEC §21.1明确接口与隔离边界。准确30b来源事实及此前提案保留，其Mission/Claim旧待决项按适用任务处理。程序已开始静态实现，CLI动态验收未通过。

M1 由 `task_ceb3b3bb76be / ctx_8d6a6e1c0145` 在 `swa-m1-contracts-implementation-20261020` 工作树及同名分支接续，实际客户端为 Orca / Codex `gpt-6.1-sol`。准确 `987811f417e92cd6f15961e60c279f616e3b21a8` 原冻结安装及未补丁的完整契约命令均 EXIT0，历史崩溃本次未复现。依已接受计划停止补丁实施；程序/依赖未改，T02—T08 未执行，因果证据不足。主控已完成 T09并曾恢复同一 S、原库和原配置，记录PID52040；后继首版窗口及当前服务见STATUS。最终记录 `b112ab9e03d9ea82247e73b5cb578529ca529f9d` 已独立接受并普通push，原hook由主控实际读取回复；本次调查Task已结算、终端released，完整修复保持未接受。

无界面执行由 `task_9f062888e786 / ctx_0d069056baff` 沿同一真实Codex会话、`swa-cli-headless-20261020` 工作树/分支继续负责。准确396e方案及T00—T15已独立接受并普通纳入main，原hook已实际读取回复；SPEC §21.1及首版任务明确接口边界、逐文件程序写域和迁移028。阶段源码、远端回执及当前检查以STATUS和Worker报告为准，准确代码由原Reviewer接续CR。原静态Task450/ctxdc资源原生retain以保留正在使用的同一终端，整体任务结算后明确释放。首次依赖窗口及prepare-01/stop-01/restore-01原件保留。最近prepare-03/stop-03及Worker的dependency-02/generate-01均实际通过，准备提交`f08b52111d4860b5ef162987cce2f43f4a4e6ce0`仅同步锁与实际HTTP类型，Worker报告普通push EXIT0。主控已收回窗口并完成restore-03：原S同源恢复PID43068、创建UTC2026-10-20T13:41:49.053571Z，原库/实际环境/配置/停止区间内容及health/Web通过，本机Orca公开CLI仍ready。当前唯一窗口归原S；CLI的§7.5退出处理和T12接续契约继续静态工作，T11范围待用户回答，构建/功能/安装/模型及双Worker验收未执行。下次窗口须先取得控制脚本两项真实进程身份更新的独立接受，再核实际归属和新资源准入，使用对应次保全原件。

已有产品原件位于原 Worker 的 `.verification/delivery-candidate-20261020/`；Windows 目录为 `dist/delivery-candidate-20261020/`。原库一致备份为该原件根 `original-01/original.sqlite`；保全、恢复条件与实际命令在[交付报告](docs/standards/delivery-candidate-20261020.md)。备份包含真实数据，仅用于本地获准核查，不能进入 Git 或发布包。

最新程序准确`c0ab15e50d230a93a0aca17ea12e1f5c132ccb5c`已交静态CR；C01—C09及T12接续契约继续由同一Worker负责。T12来源核查确认已审strict helper caller准入尚未实施，须在原生去重执行前核当前及恢复证明，后继完整契约/计划待主控定稿和独立审核。独立Reviewer同一Task已从实际退出的旧ctx980经原生retry-of恢复到`ctx_6339d8beeb73 / term_9e9193d2-733a-4ab1-8220-22a5dc640c0f`，保持原审查树、分支和两报告写域；新preamble实际开始，后续邮件只发新Dispatch。完整当前接续brief在根忽略目录`.verification/swa-review-recovery-20261020.md`，旧准确报告及恢复事实保持；尚未取得新审查结论或控制脚本接受。

科研沿[预印本任务](tasks/research-preprint-20261019.md)、[重启任务](tasks/restart-product-research-20261019.md)和[研究反馈](docs/research/research-product-decisions.md)接续；V4 沿[日运行任务](tasks/M6-V4-daily-operation.md)及[实验目录](docs/experiment/README.md)接续。外部权威锁、原件、正式研究数据、TeX 技术交付和作者许可各按原条件处理。历史管理记录见[归档说明](docs/archive/20261020/README.md)，准确 S 的交接保留在 [Git 原文](https://github.com/songconmaisaix31-design/swarm-anything/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/HANDOFF.md)。
