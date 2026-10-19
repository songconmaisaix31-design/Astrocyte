# Astrocyte 主控交接

## 当前工作

用户于2026-10-20要求明确一个当前交付候选，使集成代码、验证记录和实际运行版本对应。当前候选为 `songconmaisaix31-design/works-relay-integration@32ab1eec808a5848aabad40f11133735a1214e55`，已普通push，原PR #2实际状态为MERGED。运行源码S为ed543c12，S至候选仅治理/报告文档；完整对应关系见 [交付候选任务](tasks/delivery-candidate-20261020.md)。

[全仓文档检查](tasks/documentation-conflicts-20261020.md) 已完成，逐份处置见 [检查报告](docs/documentation-audit-20261020.md)，旧执行正文见 [归档目录](docs/archive/superseded-20261020/README.md)。新版协议和四份角色原文保持，资源与代码规范分别由 RESOURCES、CODING 统一维护。

当前版本与验收状态统一见 [STATUS](STATUS.md)，用户决定与待答项见 [QUESTIONS](docs/QUESTIONS.md)。文档用途和任务入口见 [文档目录](docs/README.md)。本文件仅保留后续执行所需信息，历史原文见 [归档](docs/archive/20261020/README.md)。

## 产品接续

产品开发入口固定为 `C:/Users/DW/orca/workspaces/Astrocyte/works-relay-integration`：

- 唯一产品主线为上述32ab，原来源分支 `songconmaisaix31-design/frontend-observatory-20261018` 同为32ab，保留原Worker的树和模块写入范围。
- 下一项真实产品任务为修复Windows契约检查进程崩溃，使原 `pnpm check:contracts` 完成；依据本次run-contracts/run-redocly真实失败，先交根因方案、具体写域和测试计划独立审核。本次只登记后续任务，没有开始新实现。
- 生产构建与功能检查实际在原产品树的S执行；集成树没有另行安装依赖或重复构建。后续依赖安装、编译与功能检查按RESOURCES安排单一服务窗口。
- Orca后续新树默认基准为 `origin/songconmaisaix31-design/works-relay-integration`，既有树及分支归属保留。
- 发送租约校验已移入产品候选，后续接续前查 [集成清单](docs/integration-audit-20261020.md)，避免重复引入。
- `validation/v1-20261012` 保留主控和正式实验记录，原分支与独有成果见 [接管清单](docs/integration-audit-20261020.md)。

实际服务为 `http://127.0.0.1:64505/`：S程序与Web同在原运行根 `versions/ed543c12`，PID45956，创建UTC `2026-10-19T19:47:04.6572166Z`，BelowNormal及唯一监听。独立Reviewer已核真实HTTP/SQLite ID、权限拒绝、全部Web字节、原业务/027/Q指针，并对准确32ab发布 [限定接受CR](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5577111605)。原库备份和c5/172旧产物保留；独立普通服务终端保持，开发Worker结算不关闭服务。

## 科研接续

沿原分支 `songconmaisaix31-design/research-preprint-20261019@fcb1e0a70541e15de1f52507a43edadb9801e368` 和原论文写入范围继续。科学正文仍为 a664233a 已接受的方法与统计提案；技术候选的完整 TeX 安装、新 PDF、来源 ZIP 及独立重建尚未完成。

先读 [科研任务](tasks/research-preprint-20261019.md)、[重启续行任务](tasks/restart-product-research-20261019.md) 和 [研究与产品决定](docs/research/research-product-decisions.md)，实际使用项目科研 skill。科研 Agent 只修改获准论文、来源和研究反馈；产品建议由主 Agent 转成具体任务。作者、机构、许可和正式研究数据缺失保持未知。

## 正式实验接续

V4 与产品任务分别管理。依据 [V4 日运行任务](tasks/M6-V4-daily-operation.md) 和 [协议](docs/EXPERIMENT.md)，先核原权威锁、day0、路由及授权。最近核查中外部原根、锁和正式 SQLite 缺失；没有恢复事实前不补采、不重建替代锁或正式库、不请求模型。

原 automation 为 enabled，已登记的日程窗口止于 2026-10-23；后续按原任务处理最后一天及窗口外禁用要求。本次没有查询或修改 automation。原模型路由、参数、漏日、预测和人评记录保留在任务及归档中。

## 原任务与停止条件

上一轮Run `run_29c7e66b4136` 的failed/released记录保留。本次Run `run_1187287582ea` 已完成候选交付、独立CR及普通集成，Task/Dispatch正式结算见当前交付任务；本轮没有接续科研或正式实验模型调用。旧未结记录继续见接管清单，不据历史标签判断当前活性。

完整默认 CI、原生/S3、两 Agent 闭环、实际采用和正式 Release 未通过；十轮仍为 01/02/03/04/06/08 限定接受，05/07/09/10 未通过。S3 预算单位、权限证据范围、受限 Codex、经验准入与采用、首发范围、安装与服务共存、Laya 语言及作者许可等事项的待答状态按 QUESTIONS 保留。

已接受的结果在原条件内复用。当前候选普通集成与运行已完成，完整CI/S3及发布门禁继续保持；契约工具失败和lint/vet串行偏差均保留。新任务按当前实际源码、资源、授权与证据安排，正式Release和arXiv上传分别遵守原条件。
