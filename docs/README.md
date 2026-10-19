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
| 文档清理 | [本次任务](../tasks/documentation-cleanup-20261020.md) | [清理前原文](archive/20261020/README.md) |
| 产品集成 | [收敛计划](../tasks/integration-convergence-20261020.md) | [准确版本、Task 与独有成果](integration-audit-20261020.md) |
| 产品与科研续行 | [重启任务](../tasks/restart-product-research-20261019.md) | [十轮改进](../tasks/s3-paper-ten-cycles-20261019.md) |
| 科研 | [预印本任务](../tasks/research-preprint-20261019.md) | [研究与产品决定](research/research-product-decisions.md) |
| 正式实验 | [V4 日运行](../tasks/M6-V4-daily-operation.md) | [锁定协议](EXPERIMENT.md)、[操作要求](experiment/OPERATIONS.md) |

## 维护方式

AGENTS 只保存长期规则；资源参数集中在 RESOURCES；代码约定集中在 CODING。HANDOFF 只保留当前执行所需内容，STATUS 只保留当前事实和实际限制。具体命令、失败记录、审查和授权依据写入对应任务或报告，再由当前入口引用。

历史任务继续保留原文件与来源。已被后续决定取代的文字按其历史日期理解；决定是否仍有效先查 QUESTIONS 和当前任务。历史状态原文见 [归档说明](archive/20261020/README.md)。
