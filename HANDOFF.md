# Astrocyte 主控交接

## 当前工作

用户于 2026-10-20 要求检查全部文档与新规范的冲突，并归档已被覆盖的旧文档。本次范围与结果见 [全仓检查任务](tasks/documentation-conflicts-20261020.md)，逐份处置见 [检查报告](docs/documentation-audit-20261020.md)，旧执行正文见 [归档目录](docs/archive/superseded-20261020/README.md)。新版协议和四份角色原文保持，资源与代码规范分别由 RESOURCES、CODING 统一维护。

当前版本与验收状态统一见 [STATUS](STATUS.md)，用户决定与待答项见 [QUESTIONS](docs/QUESTIONS.md)。文档用途和任务入口见 [文档目录](docs/README.md)。本文件仅保留后续执行所需信息，历史原文见 [归档](docs/archive/20261020/README.md)。

## 产品接续

按 [集成收敛计划](tasks/integration-convergence-20261020.md) 优先处理原产品候选：

- 原分支 `songconmaisaix31-design/frontend-observatory-20261018@5725c1840a760105518d273eaaa28d7367f369ed`。
- 原集成目标 `songconmaisaix31-design/works-relay-integration@f17a491cdd6392829094c95a3e5efb6011362bc0`。
- 保留原 Worker 的树、分支、PR 及写入范围；先核准确 CI 失败与 Source22 已修检查前置，业务修改交原 Worker，独立 Reviewer 按准确 Commit 复验。
- 发送租约校验已移入产品候选，后续接续前查 [集成清单](docs/integration-audit-20261020.md)，避免重复引入。
- 当前主控分支 `validation/v1-20261012` 保留管理与正式实验记录，产品开发沿原产品分支进行。

上次实际核查的应用为 `http://127.0.0.1:64505/`，程序 c5、Web 172bb445，早于产品候选。本次文档清理未重新探测服务、切换程序或验证数据库。原服务终端和运行目录见集成清单；重操作窗口与服务共存问题按 RESOURCES 和原任务处理。

## 科研接续

沿原分支 `songconmaisaix31-design/research-preprint-20261019@fcb1e0a70541e15de1f52507a43edadb9801e368` 和原论文写入范围继续。科学正文仍为 a664233a 已接受的方法与统计提案；技术候选的完整 TeX 安装、新 PDF、来源 ZIP 及独立重建尚未完成。

先读 [科研任务](tasks/research-preprint-20261019.md)、[重启续行任务](tasks/restart-product-research-20261019.md) 和 [研究与产品决定](docs/research/research-product-decisions.md)，实际使用项目科研 skill。科研 Agent 只修改获准论文、来源和研究反馈；产品建议由主 Agent 转成具体任务。作者、机构、许可和正式研究数据缺失保持未知。

## 正式实验接续

V4 与产品任务分别管理。依据 [V4 日运行任务](tasks/M6-V4-daily-operation.md) 和 [协议](docs/EXPERIMENT.md)，先核原权威锁、day0、路由及授权。最近核查中外部原根、锁和正式 SQLite 缺失；没有恢复事实前不补采、不重建替代锁或正式库、不请求模型。

原 automation 为 enabled，已登记的日程窗口止于 2026-10-23；后续按原任务处理最后一天及窗口外禁用要求。本次没有查询或修改 automation。原模型路由、参数、漏日、预测和人评记录保留在任务及归档中。

## 原任务与停止条件

最近核查的产品、科研和 Reviewer 属于 Run `run_29c7e66b4136`，三个 Task 均为 failed、对应 Dispatch 已 released；没有可据此视为正在工作的 Worker。准确 Task/Dispatch、独立审查与旧未结记录见集成清单，恢复前重新查询运行时，不据历史标签启动并行工作。

完整默认 CI、原生/S3、两 Agent 闭环、实际采用和正式 Release 未通过；十轮仍为 01/02/03/04/06/08 限定接受，05/07/09/10 未通过。S3 预算单位、权限证据范围、受限 Codex、经验准入与采用、首发范围、安装与服务共存、Laya 语言及作者许可等事项的待答状态按 QUESTIONS 保留。

已接受的结果在原条件内复用。未满足资源、授权或真实证据的路径保持未执行或阻塞；等待期间继续完全不依赖这些条件的工作。产品合并、部署、发布和 arXiv 上传分别遵守已有授权与验收，不由本次文档清理触发。
