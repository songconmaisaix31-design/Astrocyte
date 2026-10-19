# Astrocyte 产品接续

先读 [STATUS](STATUS.md)、[AGENTS](AGENTS.md)、[代码规范](docs/CODING.md)、[资源限制](docs/RESOURCES.md) 和[交付报告](docs/standards/delivery-candidate-20261020.md)，再读具体任务、规格与[待决定项](docs/QUESTIONS.md)。

当前运行源码S为 `ed543c12106c2dbb3e181cb9bc1647b626455179`，Go程序与Web均从同一干净Git源码普通构建。原64505已由主控切换为同一来源，PID45956；原SQLite、Q与媒体配置保留。服务由独立终端 `term_c043b5a9-d4f4-41d0-b387-866ed544ca6d` 保持，Worker结算不关闭该服务。后续操作前重新核真实PID、路径、创建时间和唯一监听。

产品开发主线为 `songconmaisaix31-design/works-relay-integration`。本次候选来源为原 `frontend-observatory-20261018` 工作树、分支和 PR #2；准确CR与普通合入分别查[原PR审查记录](https://github.com/songconmaisaix31-design/Astrocyte/pull/2)和[主控交付记录](https://github.com/songconmaisaix31-design/Astrocyte/blob/validation/v1-20261012/tasks/delivery-candidate-20261020.md)，科研和正式实验分别沿各自任务。S以后的报告与治理文档不改变构建输入，后继Commit无需重复构建。

后续产品任务为修复Windows契约检查进程崩溃，入口见 [TASKS](docs/TASKS.md)。默认CI旧替身/真实来源/权限问题继续保留。受限Codex、S3预算与权限证据、经验准入及采用尚未决定的部分不得据一般费用或集成授权开放。

本次真实原件位于 `.verification/delivery-candidate-20261020/`；Windows目录为 `dist/delivery-candidate-20261020/`。原库一致备份为该原件根 `original-01/original.sqlite`，主控原服务切换保全、恢复条件与实际命令在交付报告。备份包含真实数据，仅用于本地获准核查，不能进入Git或发布包。

科研、V4和历史管理任务由[主控交接](https://github.com/songconmaisaix31-design/Astrocyte/blob/validation/v1-20261012/HANDOFF.md)接续；本产品分支不改正式实验或论文。原交接历史保留在[准确S的Git原文](https://github.com/songconmaisaix31-design/Astrocyte/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/HANDOFF.md)。
