# SWA 主线接续

先读 [STATUS](STATUS.md)、[AGENTS](AGENTS.md)、[代码规范](docs/CODING.md) 和[资源限制](docs/RESOURCES.md)，再读 [META_PLAN](docs/META_PLAN.md)、[总任务](tasks/swa-mainline-20261020.md)、[M0](tasks/swa-m0-20261020.md)、具体规格与[用户决定](docs/QUESTIONS.md)。开发者从 [CONTRIBUTING](CONTRIBUTING.md) 进入，完整导航见[文档目录](docs/README.md)。

用户以 re.md v2.1 重构主线。M0 从产品 `32ab1eec808a5848aabad40f11133735a1214e55` 接续，仅整合文档；主控独有决定、研究、实验入口和归档来自准确 `cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`。程序来源 S、当前服务暂停和 M1 窗口统一见 STATUS。处理运行服务前重新核对真实 PID、路径、创建时间、所属终端和唯一监听，遵守原库保全及授权。

M0 准确 `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765` 已获独立接受，PR #4 已合入产品分支；主控已普通快进并 push 默认 main。后续开发从 `origin/main` 接续，Orca 默认基准已相应更新。当前 main 工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/main-20261020`；原根目录 validation 分支及实验材料保留。准确来源和逐项处置见[整合清单](docs/integration-audit-20261020.md)及 [M0 报告](docs/standards/swa-mainline-m0-20261020.md)，当前状态由 STATUS 维护。

用户最新明确首版为抽取 Orca CLI 对应执行能力、交付从原有 Agent 或终端调用的可控 Swarm CLI，主控继续核实可抽取范围，本任务不冻结新接口。原 v2.1 M2—M6 保留为后续规划；MCP 和完整 S3 不自动成为首版 CLI 的前置条件。M1 沿已接受计划实施，后继任务入口见 [TASKS](docs/TASKS.md)。默认 CI、受限 Codex、S3 预算与权限证据、经验准入及实际采用继续按原范围验收，未答项不推断授权。

首版事实源为 [CLI 任务](tasks/swa-cli-v1-20261020.md)。准确 Orca v1.4.230 来源调查及设计提案 `30b14177ef00a4e1b52f8e18c45d2375013cbd4d` 已独立接受，范围限来源材料和 E01—E16 测试覆盖框架。独立运行或依赖已安装 Orca 的选择及 Mission/Claim 两项决定未答；正向 Agent 当前受限能力、provider 闭包、精确接口和可执行计划尚未定稿，程序实现及动态验收未开始。

M1 修复由 `task_ceb3b3bb76be / ctx_8d6a6e1c0145` 在 `swa-m1-contracts-implementation-20261020` 工作树及同名分支接续，起点为 main `4132f28ecda414f2a1bd040d9eef50bbe1087cb6`，实际客户端为 Orca / Codex `gpt-6.1-sol`。沿用已独立接受的 T00—T09（含 T05A），仅修改 pnpm 配置/锁文件、精确 Redocly 补丁和两份 M1 文档；准确静态准备为 `987811f417e92cd6f15961e60c279f616e3b21a8`。T00 已完成，原服务暂停，唯一动态窗口已通过 `msg_cee9b540397e` 交付该 Worker；主控负责 T09 恢复同一 S、原 SQLite 和原配置。

已有产品原件位于原 Worker 的 `.verification/delivery-candidate-20261020/`；Windows 目录为 `dist/delivery-candidate-20261020/`。原库一致备份为该原件根 `original-01/original.sqlite`；保全、恢复条件与实际命令在[交付报告](docs/standards/delivery-candidate-20261020.md)。备份包含真实数据，仅用于本地获准核查，不能进入 Git 或发布包。

科研沿[预印本任务](tasks/research-preprint-20261019.md)、[重启任务](tasks/restart-product-research-20261019.md)和[研究反馈](docs/research/research-product-decisions.md)接续；V4 沿[日运行任务](tasks/M6-V4-daily-operation.md)及[实验目录](docs/experiment/README.md)接续。外部权威锁、原件、正式研究数据、TeX 技术交付和作者许可各按原条件处理。历史管理记录见[归档说明](docs/archive/20261020/README.md)，准确 S 的交接保留在 [Git 原文](https://github.com/songconmaisaix31-design/swarm-anything/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/HANDOFF.md)。
