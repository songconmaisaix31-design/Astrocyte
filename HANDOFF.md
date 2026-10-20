# SWA 主线接续

先读 [STATUS](STATUS.md)、[AGENTS](AGENTS.md)、[代码规范](docs/CODING.md) 和[资源限制](docs/RESOURCES.md)，再读 [META_PLAN](docs/META_PLAN.md)、[总任务](tasks/swa-mainline-20261020.md)、[M0](tasks/swa-m0-20261020.md)、具体规格与[用户决定](docs/QUESTIONS.md)。开发者从 [CONTRIBUTING](CONTRIBUTING.md) 进入，完整导航见[文档目录](docs/README.md)。

用户以 re.md v2.1 重构主线。M0 从产品 `32ab1eec808a5848aabad40f11133735a1214e55` 接续，仅整合文档；主控独有决定、研究、实验入口和归档来自准确 `cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`。程序来源 S 及最近运行观察统一见 STATUS；本轮没有重新运行程序或查询数据库。处理运行服务前重新核对真实 PID、路径、创建时间、所属终端和唯一监听，遵守原库保全及授权。

M0 PR 目标为 `songconmaisaix31-design/works-relay-integration`，独立接受后由主控安排产品历史和有效文档进入默认 main。仓库实际名称为 `songconmaisaix31-design/swarm-anything`。准确来源和逐项处置见[整合清单](docs/integration-audit-20261020.md)及 [M0 报告](docs/standards/swa-mainline-m0-20261020.md)。本任务不执行 main 合并、部署或发布。

用户最新明确首版为抽取 Orca CLI 对应执行能力、交付从原有 Agent 或终端调用的可控 Swarm CLI，主控继续核实可抽取范围，本任务不冻结新接口。原 v2.1 M2—M6 保留为后续规划；MCP 和完整 S3 不自动成为首版 CLI 的前置条件。原 M1 测试计划已接受、尚未运行，后继任务入口见 [TASKS](docs/TASKS.md)。默认 CI、受限 Codex、S3 预算与权限证据、经验准入及实际采用继续按原范围验收，未答项不推断授权。

首版事实源为 [CLI 任务](tasks/swa-cli-v1-20261020.md)。独立运行或依赖已安装 Orca 的执行边界已询问用户，尚未回答；源码、许可、能力与依赖调查继续，依赖该答案的执行接口保持未决定。最新接续来源为主控 `9140cf8aa618c832b3a099222a46991b7c2b3ddd` 的六份变更文档，整合过程保留当前有效规则和导航。

已有产品原件位于原 Worker 的 `.verification/delivery-candidate-20261020/`；Windows 目录为 `dist/delivery-candidate-20261020/`。原库一致备份为该原件根 `original-01/original.sqlite`；保全、恢复条件与实际命令在[交付报告](docs/standards/delivery-candidate-20261020.md)。备份包含真实数据，仅用于本地获准核查，不能进入 Git 或发布包。

科研沿[预印本任务](tasks/research-preprint-20261019.md)、[重启任务](tasks/restart-product-research-20261019.md)和[研究反馈](docs/research/research-product-decisions.md)接续；V4 沿[日运行任务](tasks/M6-V4-daily-operation.md)及[实验目录](docs/experiment/README.md)接续。外部权威锁、原件、正式研究数据、TeX 技术交付和作者许可各按原条件处理。历史管理记录见[归档说明](docs/archive/20261020/README.md)，准确 S 的交接保留在 [Git 原文](https://github.com/songconmaisaix31-design/swarm-anything/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/HANDOFF.md)。
