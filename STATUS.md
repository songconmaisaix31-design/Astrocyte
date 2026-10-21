# Astrocyte 当前状态

## 本轮主线重构

用户已明确首版为抽取 Orca CLI 执行能力、从原有 Agent 或终端调用的可控 Swarm CLI，见[首版任务](tasks/swa-cli-v1-20261020.md)。M0 文档整合正在检查来源和链接；M1 契约修复测试计划已获独立接受，动态检查尚未开始。当前核实对应 Orca 源码，是否脱离桌面独立运行的边界待用户回答；本轮尚未修改 main、程序或数据。整体整合规划仍见 [META_PLAN](docs/META_PLAN.md)，执行信息见[总任务](tasks/swa-mainline-20261020.md)。

## V4 日运行（2026-10-22）

本次仅核查 Asia/Shanghai 昨日2026-10-21，automation run 10 `be8c8b7e-313a-4d67-86a5-e37abad0568d` 已实际派发。协议全文 SHA256 与登记 `5aa238599b6dd59e28d9ca30f71b6b9d547822b46d8dadc477092a46c09ab1b7` 一致，Codex CLI `0.162.0`，相关运行源码相对 `d749ecb` 无差异。

原外部 `C:/Users/DW/AppData/Local/astrocyte-validation/`、`v1-20261012` 根、权威锁和正式 SQLite 的 `Test-Path` 均为 false。真实锁、day0、路由及授权无法重新核验；按历史锁日昨日对应 D8，本次采集、Sleep、成熟预测及其他实验命令均未启动，新增事件、Sleep、回答、judge、人评及实验模型请求均0。历史累计数无法复核，没有实际可用用户评分表路径，终局门禁未满足。

阻塞 Handoff `msg_6383905de04a` 已沿原线程排入原 W1 所属 Run，关联原 Task `task_e538790026ad`；原 owner 已读及原根恢复未确认。10-21 run 9 为 `dispatch_failed`，不计为已采集。automation 仍 enabled，下次2026-10-23 00:20；最后 D9 处理结束或窗口外唤醒须禁用。仅更新 root HANDOFF/STATUS，普通 `[skip ci]` 提交与 push；执行细节见 [HANDOFF](HANDOFF.md)。

## 状态来源

本页依据 2026-10-20 [交付候选与实际运行收敛](tasks/delivery-candidate-20261020.md) 更新。当前交付候选已完成源码、验证记录和实际运行对应，并普通推送到产品主线。原全量分支与旧任务事实见 [接管核查](docs/integration-audit-20261020.md)，接续步骤见 [HANDOFF](HANDOFF.md)。

## 准确版本

产品、科研和审查分支的完整前缀为 `songconmaisaix31-design/`；主控分支与 `main` 使用表中的完整名称。

| 用途 | 分支或位置 | 记录中的准确来源 |
| --- | --- | --- |
| 当前交付候选与唯一产品主线 | `works-relay-integration` | **`32ab1eec808a5848aabad40f11133735a1214e55`**；本地、远端、原PR #2合入来源一致，普通快进及push完成 |
| 构建、功能验证与实际运行源码 S | 原产品工作树 `frontend-observatory-20261018` | `ed543c12106c2dbb3e181cb9bc1647b626455179`；S至候选HEAD仅15份治理/报告Markdown，业务、Web、契约、迁移、依赖与构建输入相同；独立运行和最终文档审查限定接受 |
| 当前普通交付目录 | 原产品工作树 `dist/delivery-candidate-20261020/` | 三个Go程序、Web及启动入口；Go版本元数据为完整S、modified=false，实际运行文件内容对应 |
| 历史完整便携包 | Source15 | `15d74da3edcee26550a71120370fda9c86230e04`；原130项验包保持该来源，本次没有重建完整便携模型包 |
| 当前实际服务 | `http://127.0.0.1:64505/` | 程序与Web均为S，版本目录`versions/ed543c12`；PID45956，创建UTC `2026-10-19T19:47:04.6572166Z`，BelowNormal、唯一loopback监听、health正常；原库业务保全及027迁移已核 |
| 科研候选 | `research-preprint-20261019`，原 PR #3 | `fcb1e0a70541e15de1f52507a43edadb9801e368`；科学正文仍为 a664233a |
| 本次独立审查与结算 | `research-preprint-review-20261019` | 最终报告 `43d4f132c67d115cc0fabb27dd180a38447c80f9`；准确32ab的[COMMENT review 5577111605](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5577111605)已核；结算回执见当前任务 |
| 主控与正式实验记录 | `validation/v1-20261012` | 本次接管3484d106，运行节点记录42570f89；产品代码通过上述产品主线维护 |
| 默认分支 | `main` | `3a02ce38481d1885adee2e051a1b524d78376897`；产品尚未合入 |

## 已验范围与真实限制

| 范围 | 已有证据 | 未通过或未完成 |
| --- | --- | --- |
| 当前普通候选与实际运行 | S普通构建、真实DAG/OpenViking专项、53项显示规则、双视口12项；独立产物/迁移/实际运行、49项GET含5项403、原资料2及项目1的HTTP/SQLite ID对应、原业务/Q保全 | 完整默认CI和默认浏览器NOT_RUN；`pnpm check:contracts`及Redocly直调崩溃FAIL；lint/vet实际重叠不符合串行条件；完整新便携模型包未验 |
| 动态 DAG | 人工任务依赖、增量、来源、撤权清图与按需加载获限定接受 | 模型生成、原生执行与带 Worker 生命周期尚缺合法正向证据 |
| 科研 | 方法与统计提案、14 条引用、已有 PDF 流程和真实来源审查获限定接受 | 固定 TeX 完整安装、新 PDF/来源 ZIP 及独立编译未完成；作者/机构/许可、正式研究数据及 arXiv 就绪缺失 |
| 十轮改进 | 01/02/03/04/06/08 限定接受，详见 [任务](tasks/s3-paper-ten-cycles-20261019.md) | 05/07/09/10 未通过；限定接受不代表 S3 或投稿通过 |
| V4 正式实验 | 历史采集、评分和冻结记录保留在原任务及归档 | 最近核查原证据根、权威锁与正式 SQLite 缺失；无法复核历史累计数，未补采或启动模型；人评及终局未完成 |
| 发布 | 当前准确候选、实际运行及本次治理文档已获限定接受；用户已授权满足原门禁后的首版GitHub Release | 完整CI、原生/S3、两Agent闭环及采用等原门禁未通过；预览首发范围未答，没有tag/Release或arXiv上传 |

完整 CI 没有通过证据；GitHub 的可合并状态及空的 head 检查不能代替 CI 结果。PR、远端 HEAD 和运行时状态随时间变化，实际接续前按原任务重新核查。

## 任务和待答项

本次Run `run_1187287582ea` 的产品Worker与独立Reviewer均已正式报告succeeded，两个Task均completed，两个Dispatch终端均released并保存转录。本次没有保留live开发Worker；实际服务由独立普通终端 `term_c043b5a9-d4f4-41d0-b387-866ed544ca6d` 保持。Orca后续新工作树默认基准已设为 `origin/songconmaisaix31-design/works-relay-integration`。准确结算见当前任务，旧未结项保留在接管清单。

下一项产品任务为修复Windows契约检查进程崩溃，使原 `pnpm check:contracts` 完成；原Worker先提交根因方案和测试计划，经独立审核后实施。本次没有在集成工作树另行安装依赖或重复构建，开发前按锁文件及资源窗口准备环境。

S3 预算单位、权限证据范围与受限 Codex 选择仍未回答，Claim 拒绝及 Known=false 保持。具体资料准入、Agent 读取完成证据、经验写入交接的方式、Laya 语言、首发范围、安装与服务共存和作者许可继续见 [QUESTIONS](docs/QUESTIONS.md)。实际用户作品、自由工具授权和正式实验缺失字段保持未知。

## 文档维护

新版 [AGENTS](AGENTS.md) 与四份 [角色说明](docs/agents/README.md) 保持用户原文，CODING、RESOURCES提供统一规范；上述规则及当前产品入口已带入32ab产品主线。全仓文档的前轮处置见 [报告](docs/documentation-audit-20261020.md)，独有需求、冻结参数与真实证据继续保留。[docs/STATUS.md](docs/STATUS.md) 只保留本页入口；历史状态原文见 [归档](docs/archive/20261020/README.md)。
