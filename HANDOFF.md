# Astrocyte 主控交接

## 当前工作

用户于 2026-10-20 要求明确一个当前交付候选，使集成代码、验证记录和实际运行版本对应，沿真实产品开发主线收敛已有成果。当前执行见 [交付候选任务](tasks/delivery-candidate-20261020.md)，产品仍沿原 PR #2 和工作树接续，原 `works-relay-integration` 为产品开发主线。

[全仓文档检查](tasks/documentation-conflicts-20261020.md) 已完成，逐份处置见 [检查报告](docs/documentation-audit-20261020.md)，旧执行正文见 [归档目录](docs/archive/superseded-20261020/README.md)。新版协议和四份角色原文保持，资源与代码规范分别由 RESOURCES、CODING 统一维护。

当前版本与验收状态统一见 [STATUS](STATUS.md)，用户决定与待答项见 [QUESTIONS](docs/QUESTIONS.md)。文档用途和任务入口见 [文档目录](docs/README.md)。本文件仅保留后续执行所需信息，历史原文见 [归档](docs/archive/20261020/README.md)。

## 产品接续

按 [集成收敛计划](tasks/integration-convergence-20261020.md) 优先处理原产品候选：

- 原分支 `songconmaisaix31-design/frontend-observatory-20261018@5725c1840a760105518d273eaaa28d7367f369ed`。
- 原集成目标 `songconmaisaix31-design/works-relay-integration@f17a491cdd6392829094c95a3e5efb6011362bc0`。
- 保留原 Worker 的树、分支、PR 及写入范围；先核准确 CI 失败与 Source22 已修检查前置，业务修改交原 Worker，独立 Reviewer 按准确 Commit 复验。
- 发送租约校验已移入产品候选，后续接续前查 [集成清单](docs/integration-audit-20261020.md)，避免重复引入。
- 当前主控分支 `validation/v1-20261012` 保留管理与正式实验记录，产品开发沿原产品分支进行。

本次冻结构建源码为 `ed543c12106c2dbb3e181cb9bc1647b626455179`，同源程序与Web已构建并获独立可部署范围接受。主控已将其部署到原运行根 `versions/ed543c12` 并使用原真实库启动64505：PID45956，创建UTC `2026-10-19T19:47:04.6572166Z`，BelowNormal、唯一监听、health正常。原业务记录保全，准确027迁移及唯一启动观察时间增量已核；原库备份及c5/172旧产物保留。当前独立Reviewer正在唯一服务内只读核真实ID、权限与静态资源，随后Worker提交最终治理文档及报告。准确节点、Task/Dispatch与证据入口见 [当前任务](tasks/delivery-candidate-20261020.md)。

## 科研接续

沿原分支 `songconmaisaix31-design/research-preprint-20261019@fcb1e0a70541e15de1f52507a43edadb9801e368` 和原论文写入范围继续。科学正文仍为 a664233a 已接受的方法与统计提案；技术候选的完整 TeX 安装、新 PDF、来源 ZIP 及独立重建尚未完成。

先读 [科研任务](tasks/research-preprint-20261019.md)、[重启续行任务](tasks/restart-product-research-20261019.md) 和 [研究与产品决定](docs/research/research-product-decisions.md)，实际使用项目科研 skill。科研 Agent 只修改获准论文、来源和研究反馈；产品建议由主 Agent 转成具体任务。作者、机构、许可和正式研究数据缺失保持未知。

## 正式实验接续

V4 与产品任务分别管理。依据 [V4 日运行任务](tasks/M6-V4-daily-operation.md) 和 [协议](docs/EXPERIMENT.md)，先核原权威锁、day0、路由及授权。最近核查中外部原根、锁和正式 SQLite 缺失；没有恢复事实前不补采、不重建替代锁或正式库、不请求模型。

原 automation 为 enabled，已登记的日程窗口止于 2026-10-23；后续按原任务处理最后一天及窗口外禁用要求。本次没有查询或修改 automation。原模型路由、参数、漏日、预测和人评记录保留在任务及归档中。

## 原任务与停止条件

上一轮产品、科研和 Reviewer 属于 Run `run_29c7e66b4136`，三个 Task 均为 failed、对应 Dispatch 已 released。当前产品收敛使用 Run `run_1187287582ea`，原产品工作树 Worker 与独立 Reviewer 已通过 Orca 启动，准确 Task/Dispatch 及验收顺序见 [交付候选任务](tasks/delivery-candidate-20261020.md)。本轮没有接续科研和正式实验模型调用。旧未结记录继续见集成清单，不据历史标签判断当前活性。

完整默认 CI、原生/S3、两 Agent 闭环、实际采用和正式 Release 未通过；十轮仍为 01/02/03/04/06/08 限定接受，05/07/09/10 未通过。S3 预算单位、权限证据范围、受限 Codex、经验准入与采用、首发范围、安装与服务共存、Laya 语言及作者许可等事项的待答状态按 QUESTIONS 保留。

已接受的结果在原条件内复用。未满足资源、授权或真实证据的路径保持未执行或阻塞；等待期间继续完全不依赖这些条件的工作。产品合并、部署、发布和 arXiv 上传分别遵守已有授权与验收，不由本次文档清理触发。
