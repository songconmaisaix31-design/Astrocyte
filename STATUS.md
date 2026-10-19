# Astrocyte 产品状态

当前产品源码为 `ed543c12106c2dbb3e181cb9bc1647b626455179`（S）。程序、Web、真实检查和运行参数见[交付报告](docs/standards/delivery-candidate-20261020.md)。报告及治理文档后继通过 Git 差异明确只含文档，不为文档重复构建。

| 项目 | 实际状态 |
| --- | --- |
| 产品开发主线 | `songconmaisaix31-design/works-relay-integration`；本次候选沿原 PR #2，由主控普通推进 |
| 候选来源分支 / 原PR | `songconmaisaix31-design/frontend-observatory-20261018` / [PR #2](https://github.com/songconmaisaix31-design/Astrocyte/pull/2) |
| 实际应用 | `http://127.0.0.1:64505/`，PID45956，启动UTC `2026-10-19T19:47:04.657217Z`，BelowNormal |
| 程序与 Web | 同一S，`D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/versions/ed543c12/` |
| SQLite | 原运行根 `local/astrocyte-workbench/v0.2-access/app-state/state.sqlite`；原业务数据保全，准确027迁移与既有启动观察时间分别核查 |
| Windows 普通交付 | `dist/delivery-candidate-20261020/`；三个Go程序、Web及启动入口；完整新便携模型包未验 |
| 必要自测 | 构建、DAG9函数、OpenViking4函数、前端53项、DAG双视口4项及页面8项、架构/依赖、264项契约示例与生成一致性通过；命令和原件见报告 |
| 真实限制 | 完整默认CI NOT_RUN；原 `pnpm check:contracts` 及Redocly直调崩溃FAIL；lint/vet发生串行偏差，实际退出与额度保留 |

独立Reviewer已接受准确S的同源产物、迁移副本和当前原库运行对应：实际49项GET（含5项403）、真实资料/项目ID与原SQLite对应、33件Web字节及原业务/Q保全均通过。准确Commit的CR见[原PR #2审查记录](https://github.com/songconmaisaix31-design/Astrocyte/pull/2)，产品主线的普通合入见[主控交付记录](https://github.com/songconmaisaix31-design/Astrocyte/blob/validation/v1-20261012/tasks/delivery-candidate-20261020.md)。完整默认浏览器、真实原生/S3、带Worker停止退出/未知用量释放、两Agent闭环、成果采用及OpenViking真实远端正向继续未验；正式Release未创建。

下一项产品任务为[修复Windows契约检查进程崩溃](docs/TASKS.md)，须先独立审核根因方案与测试计划。本次不开始该实现，不代答原生、预算、资料和权限选择。

主控的合入结果、科研和正式实验事实源见[主控状态](https://github.com/songconmaisaix31-design/Astrocyte/blob/validation/v1-20261012/STATUS.md)。原历史状态保留在[准确S的Git原文](https://github.com/songconmaisaix31-design/Astrocyte/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/STATUS.md)。
