# Agent 角色说明

四份角色说明分别来自用户提供的 `tabbit-code (1).md` 至 `tabbit-code (4).md`，正文保留用户原文。

| 角色 | 文件 | 用户提供的原件 |
| --- | --- | --- |
| 主 Agent | [main.md](main.md) | `tabbit-code (1).md` |
| 模块 Worker | [worker.md](worker.md) | `tabbit-code (2).md` |
| 独立 Reviewer | [reviewer.md](reviewer.md) | `tabbit-code (3).md` |
| 科研 Agent | [research.md](research.md) | `tabbit-code (4).md` |

启动相应角色时，先加载根 [AGENTS.md](../../AGENTS.md)，再加载该角色文件及其中要求的事实源。派发模块任务时填写 `{模块名}`；科研任务填写 `{论文目录}`，两者均以具体任务获准范围为准。这些文件供任务引用，运行时是否自动加载须由实际客户端确认。

执行规则统一维护在 AGENTS，代码规则见 [CODING](../CODING.md)，资源参数见 [RESOURCES](../RESOURCES.md)。角色说明中的开发流程用于对应开发任务；本次管理文档清理的范围见 [当前任务](../../tasks/documentation-cleanup-20261020.md)。
