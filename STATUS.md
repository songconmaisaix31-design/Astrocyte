# Astrocyte 当前状态

## 状态来源

本页依据 2026-10-20 [集成核查](docs/integration-audit-20261020.md) 和 [重启续行任务](tasks/restart-product-research-20261019.md) 汇总。文档清理没有重新运行产品、科研或正式实验；表中的运行状态保留原核查时间，不代表本次实时探测。当前管理任务见 [文档清理](tasks/documentation-cleanup-20261020.md)，接续步骤见 [HANDOFF](HANDOFF.md)。

## 准确版本

产品、科研和审查分支的完整前缀为 `songconmaisaix31-design/`；主控分支与 `main` 使用表中的完整名称。

| 用途 | 分支或位置 | 记录中的准确来源 |
| --- | --- | --- |
| 优先产品候选 | `frontend-observatory-20261018`，原 PR #2 | `5725c1840a760105518d273eaaa28d7367f369ed`；业务检查源码 `22cde9efc540553c433463af5b9e955790d57a99`，后继为文档 |
| 产品集成目标 | `works-relay-integration` | `f17a491cdd6392829094c95a3e5efb6011362bc0` |
| 独立验包来源 | Source15 | `15d74da3edcee26550a71120370fda9c86230e04`；Source22 尚未重建包 |
| 最近核查的运行应用 | `http://127.0.0.1:64505/` | 程序 `c5b3b5715f2e93b82365720914be24c73642bd1e`、Web `172bb445597193d55f1563ee5d21ac2ec881798d`；PID 45676，创建 UTC `2026-10-19T12:09:30.7340265Z` |
| 科研候选 | `research-preprint-20261019`，原 PR #3 | `fcb1e0a70541e15de1f52507a43edadb9801e368`；科学正文仍为 a664233a |
| 最新独立审查报告 | `research-preprint-review-20261019` | `13bba19f5fa4c7fc2980ebc840c70ea9765bd4b8` |
| 主控与正式实验 | `validation/v1-20261012` | 文档清理基线 `a56dcbe3080404953a80b1f6fd518be8d4820508` |
| 默认分支 | `main` | `3a02ce38481d1885adee2e051a1b524d78376897`；产品尚未合入 |

## 已验范围与真实限制

| 范围 | 已有证据 | 未通过或未完成 |
| --- | --- | --- |
| 产品与 Windows 包 | Source15 的构建、独立 130 项运行检查和双视口人工 DOM；Source22 的三项真实 HTTP/SQLite 检查；原 c5 同库恢复及只读独立审查 | 完整默认 CI、默认浏览器矩阵、真实原生/S3、带 Worker 停止与未知用量对账及释放、两 Agent 闭环、成果实际采用 |
| 动态 DAG | 人工任务依赖、增量、来源、撤权清图与按需加载获限定接受 | 模型生成、原生执行与带 Worker 生命周期尚缺合法正向证据 |
| 科研 | 方法与统计提案、14 条引用、已有 PDF 流程和真实来源审查获限定接受 | 固定 TeX 完整安装、新 PDF/来源 ZIP 及独立编译未完成；作者/机构/许可、正式研究数据及 arXiv 就绪缺失 |
| 十轮改进 | 01/02/03/04/06/08 限定接受，详见 [任务](tasks/s3-paper-ten-cycles-20261019.md) | 05/07/09/10 未通过；限定接受不代表 S3 或投稿通过 |
| V4 正式实验 | 历史采集、评分和冻结记录保留在原任务及归档 | 最近核查原证据根、权威锁与正式 SQLite 缺失；无法复核历史累计数，未补采或启动模型；人评及终局未完成 |
| 发布 | 用户已授权满足原门禁后的首版 GitHub Release | 准确最终源码交付及对应独立验收尚未满足；预览首发范围未答，没有 tag/Release 或 arXiv 上传 |

完整 CI 没有通过证据；GitHub 的可合并状态及空的 head 检查不能代替 CI 结果。PR、远端 HEAD 和运行时状态随时间变化，实际接续前按原任务重新核查。

## 任务和待答项

最近集成核查中没有运行时报告的 live Worker；产品、科研和 Reviewer 的最新 Task 均为 failed，对应 Dispatch 已 released。原普通服务终端独立保留，旧未结 Task 与独有分支逐项见集成清单。本次没有更改 Task、Dispatch、服务、工作树或分支归属。

S3 预算单位、权限证据范围与受限 Codex 选择仍未回答，Claim 拒绝及 Known=false 保持。具体资料准入、Agent 读取完成证据、经验写入交接的方式、Laya 语言、首发范围、安装与服务共存和作者许可继续见 [QUESTIONS](docs/QUESTIONS.md)。实际用户作品、自由工具授权和正式实验缺失字段保持未知。

## 文档维护

新版 [AGENTS](AGENTS.md) 使用用户原文，四份 [角色说明](docs/agents/README.md) 分别保存；[CODING](docs/CODING.md) 和 [RESOURCES](docs/RESOURCES.md) 提供统一规范。[docs/STATUS.md](docs/STATUS.md) 只保留本页入口。全部历史状态、授权和验证原文见 [清理前归档](docs/archive/20261020/README.md)，不作为新增验收结果。
