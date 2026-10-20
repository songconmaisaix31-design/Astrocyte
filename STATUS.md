# SWA 当前状态

本页统一维护当前版本、已验范围和实际限制。资源参数只在 [RESOURCES](docs/RESOURCES.md) 维护，授权和待答项只在 [QUESTIONS](docs/QUESTIONS.md) 维护，接管操作见 [HANDOFF](HANDOFF.md)。

## 当前主线与 M0

| 项目 | 当前事实 |
| --- | --- |
| 用户首版目标 | 抽取 Orca CLI 对应执行能力，交付用户从原有 Agent 或终端调用的可控 Swarm CLI；[首版任务](tasks/swa-cli-v1-20261020.md)记录具体范围与验收，尚未实现 |
| 首版执行边界 | 是否脱离 Orca 桌面独立运行已询问、未答；执行接口保持未决定，继续真实源码及依赖调查 |
| 规范仓库 / 默认分支 | GitHub 仓库 ID `1410836101`，`songconmaisaix31-design/swarm-anything` / `main` |
| main 的产品与治理集成来源 | `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765`；已从初始 main 普通快进并 push，完整包含产品32ab历史与接受后的M0文档；后继当前事实记录沿 main 维护 |
| 产品基线 | `songconmaisaix31-design/works-relay-integration@32ab1eec808a5848aabad40f11133735a1214e55`；原 PR #2 已普通合入 |
| M0 任务分支 | `swa-m0-20261020@6eb7c7f640a9cf1b05ea365b3afdf6883cd01765`；独立 [COMMENT review 5581860208](https://github.com/songconmaisaix31-design/swarm-anything/pull/4#pullrequestreview-5581860208) 接受，原hook已由主控读取并回复；[PR #4](https://github.com/songconmaisaix31-design/swarm-anything/pull/4) 已实际合入产品分支，同源快进 main |
| 主控内容来源 | `validation/v1-20261012@cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`；独有决定、归档、任务、实验入口与研究文档按[整合清单](docs/integration-audit-20261020.md)纳入 |
| 后继工作 | 首版来源调查准确30b已独立接受，实施边界与可执行计划待定，归属见[首版任务](tasks/swa-cli-v1-20261020.md)；M1 修复 Worker 已在独立工作树接续已接受计划，动态窗口尚未交付；后续排序见 [TASKS](docs/TASKS.md) |

当前 main 工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/main-20261020`；原根目录的 `validation/v1-20261012` 与正式实验材料保留。Orca 的新工作树默认基准已实际改为 `origin/main`。本轮没有停止服务、修改程序或数据库，也没有将 M0 的接受计为首版 CLI 功能通过。

## 已有产品验证与最近运行观察

程序和 Web 的已验源码 S 为 `ed543c12106c2dbb3e181cb9bc1647b626455179`。以下运行信息继承 2026-10-20 交付记录，本次纯文档工作没有重新查询进程、HTTP 或数据库；后续操作前须重新核实。原件、准确命令与独立接受范围见[产品交付报告](docs/standards/delivery-candidate-20261020.md)及[交付任务](tasks/delivery-candidate-20261020.md)。

| 范围 | 记录的结果 |
| --- | --- |
| 最近实际应用观察 | `http://127.0.0.1:64505/`，PID45956，启动 UTC `2026-10-19T19:47:04.657217Z`，BelowNormal；所属终端 `term_c043b5a9-d4f4-41d0-b387-866ed544ca6d` |
| 同源产物 | 程序和 Web 同一 S；`D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/versions/ed543c12/` |
| 原库与配置 | `local/astrocyte-workbench/v0.2-access/app-state/state.sqlite`；原库、Q、媒体配置保全，准确027迁移与原启动观察分别核查 |
| Windows 普通交付 | `dist/delivery-candidate-20261020/`，三个 Go 程序、Web 与启动入口；完整新便携模型包未验 |
| 已通过的限定检查 | 构建、DAG9函数、OpenViking4函数、前端53项、DAG双视口4项、页面8项、架构/依赖、264项契约示例及生成一致性 |
| 独立运行核验 | 准确 S 的49项 GET（含5项403）、真实资料/项目 ID 与原 SQLite 对应、33件 Web 字节、原业务和 Q 保全 |
| 原始失败与未执行 | 完整默认 CI：NOT_RUN；原 `pnpm check:contracts` 与 Redocly 直调：进程崩溃 FAIL；lint/vet 的实际串行偏差和退出保留在原报告 |

完整默认浏览器、真实原生/S3、带 Worker 停止退出、未知用量对账和释放、两 Agent 闭环、成果采用、OpenViking 真实远端正向均未通过完整验收。正式 Release 未创建。原生控制器工具关闭、真实事件逐项定位、预算与权限证据门禁保持。

## 科研与正式实验

| 范围 | 已有结果及限制 |
| --- | --- |
| 科研预印本 | 独立分支 `songconmaisaix31-design/research-preprint-20261019@fcb1e0a70541e15de1f52507a43edadb9801e368`；方法稿已有接受范围，完整 TeX 安装/构建/验证、新引擎产物及独立重建未完成；正式实验结果、作者/机构/许可未确认，原 [PR #3](https://github.com/songconmaisaix31-design/swarm-anything/pull/3) 的证据不足范围保持 |
| 研究与产品边界 | [研究反馈](docs/research/research-product-decisions.md)与[科研 Skill](.agents/skills/astrocyte-research/SKILL.md)已保留；研究运行时、正式实验代码和论文原件仍在原位置，不作为产品执行权威 |
| 正式 V4 | 上次接管核查中 `C:/Users/DW/AppData/Local/astrocyte-validation/` 的原证据根、外部权威锁和正式 SQLite 不存在，无法核实锁/day0/路由/授权；本轮没有重新核查或补采。实际人评与终局未完成，按[日运行任务](tasks/M6-V4-daily-operation.md)和[实验入口](docs/experiment/README.md)接续 |

完整历史状态见[20261020归档](docs/archive/20261020/README.md)，本轮状态入口迁移的产品原文见[M0归档](docs/archive/swa-m0-20261020/README.md)。历史任务中的旧进程、派发和额度按原核查时间阅读。
