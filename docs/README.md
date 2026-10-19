# 项目文档

接管时先读根目录的 [AGENTS.md](../AGENTS.md)、[HANDOFF.md](../HANDOFF.md) 和 [STATUS.md](../STATUS.md)，再按任务读取具体规格。文档中的历史检查保留原时间和来源，当前状态由根 STATUS 统一维护。

| 需要了解的内容 | 文档 |
| --- | --- |
| 产品用途、开发环境和启动命令 | [README](../README.md) |
| 长期执行规则 | [AGENTS](../AGENTS.md) |
| 当前任务、下一步与停止条件 | [HANDOFF](../HANDOFF.md) |
| 当前版本、已验范围与阻塞 | [STATUS](../STATUS.md) |
| 业务需求、模块与接口契约 | [SPEC](SPEC.md) |
| 开发规划与当前任务入口 | [TASKS](TASKS.md) |
| 用户授权、已确认决定与待答问题 | [QUESTIONS](QUESTIONS.md) |
| 代码组织、命名、接口与错误处理 | [CODING](CODING.md) |
| 资源额度、采样及静态检查范围 | [RESOURCES](RESOURCES.md) |
| 主 Agent、Worker、Reviewer、科研 Agent 的角色说明 | [角色目录](agents/README.md) |

## 当前任务与依据

| 范围 | 执行文件 | 相关依据 |
| --- | --- | --- |
| 当前交付候选与产品主线 | [交付任务](../tasks/delivery-candidate-20261020.md) | [准确32ab产品报告](https://github.com/songconmaisaix31-design/Astrocyte/blob/32ab1eec808a5848aabad40f11133735a1214e55/docs/standards/delivery-candidate-20261020.md)、[独立CR](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5577111605) |
| 全仓文档检查与归档（已完成） | [任务](../tasks/documentation-conflicts-20261020.md) | [逐份处置报告](documentation-audit-20261020.md)、[旧执行入口归档](archive/superseded-20261020/README.md) |
| 接管时的全量分支与任务 | [接管计划](../tasks/integration-convergence-20261020.md) | [原版本、Task与独有成果](integration-audit-20261020.md) |
| 产品与科研续行 | [重启任务](../tasks/restart-product-research-20261019.md) | [十轮改进](../tasks/s3-paper-ten-cycles-20261019.md) |
| 科研 | [预印本任务](../tasks/research-preprint-20261019.md) | [研究与产品决定](research/research-product-decisions.md) |
| 正式实验 | [V4 日运行](../tasks/M6-V4-daily-operation.md) | [实验目录与锁说明](experiment/README.md)、[原锁定协议](EXPERIMENT.md) |

## 维护方式

AGENTS 只保存长期规则；资源参数集中在 RESOURCES；代码约定集中在 CODING。HANDOFF 只保留当前执行所需内容，STATUS 只保留当前事实和实际限制。具体命令、失败记录、审查和授权依据写入对应任务或报告，再由当前入口引用。

历史任务按 [逐份处置报告](documentation-audit-20261020.md) 保留独有正文或转入归档，旧路径保留可用入口。决定是否仍有效先查 QUESTIONS 和当前任务，原日期的人员、额度和操作不作为当前派发。首次规则安装记录见 [原任务](../tasks/documentation-cleanup-20261020.md)，历史状态原文见 [归档说明](archive/20261020/README.md)。
