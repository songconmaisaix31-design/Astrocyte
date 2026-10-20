# Agent 角色说明

四份角色说明基于用户提供的 `tabbit-code (1).md` 至 `tabbit-code (4).md`，按 [META_PLAN §7.5](../META_PLAN.md) 维护当前职责。修订前原文可从产品 `32ab1eec808a5848aabad40f11133735a1214e55` 的同名 Git 路径读取，决定来源见 [QUESTIONS](../QUESTIONS.md)。

| 角色 | 文件 | 用户提供的原件 |
| --- | --- | --- |
| 主 Agent | [main.md](main.md) | `tabbit-code (1).md` |
| 模块 Worker | [worker.md](worker.md) | `tabbit-code (2).md` |
| 独立 Reviewer | [reviewer.md](reviewer.md) | `tabbit-code (3).md` |
| 科研 Agent | [research.md](research.md) | `tabbit-code (4).md` |

启动相应角色时，按根 [AGENTS.md](../../AGENTS.md) 的阅读顺序加载当前事实源，再加载角色文件及任务需要的规格。派发模块任务时填写 `{模块名}`；科研任务填写 `{论文目录}`，均以获准范围为准。运行时是否自动加载由实际客户端确认。

执行规则统一维护在 AGENTS，代码规则见 [CODING](../CODING.md)，资源参数见 [RESOURCES](../RESOURCES.md)。角色说明用于对应开发任务，具体范围与验收由 [当前任务目录](../TASKS.md) 引用的任务确定。
