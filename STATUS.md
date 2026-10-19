# Astrocyte 当前状态

## 状态来源

本页依据 2026-10-20 [交付候选与实际运行收敛](tasks/delivery-candidate-20261020.md) 更新。原全量分支与旧任务事实见 [集成核查](docs/integration-audit-20261020.md)。本次正在沿原产品主线执行普通构建和必要真实检查，科研与正式实验未重新启动；接续步骤见 [HANDOFF](HANDOFF.md)。

## 准确版本

产品、科研和审查分支的完整前缀为 `songconmaisaix31-design/`；主控分支与 `main` 使用表中的完整名称。

| 用途 | 分支或位置 | 记录中的准确来源 |
| --- | --- | --- |
| 当前交付候选的构建来源 S | `frontend-observatory-20261018`，原 PR #2 | `ed543c12106c2dbb3e181cb9bc1647b626455179`；三个Go程序及Web已同源构建，独立可部署复验接受；实际运行已启动，独立运行复验进行中 |
| 产品集成目标 | `works-relay-integration` | `f17a491cdd6392829094c95a3e5efb6011362bc0` |
| 独立验包来源 | Source15 | `15d74da3edcee26550a71120370fda9c86230e04`；Source22 尚未重建包 |
| 当前实际服务 | `http://127.0.0.1:64505/` | 程序与Web均为S，版本目录`versions/ed543c12`；PID45956，创建UTC `2026-10-19T19:47:04.6572166Z`，BelowNormal、唯一loopback监听、health正常；原库业务保全及027迁移已核 |
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

当前 Run 为 `run_1187287582ea`。产品 Worker `task_6f09f37378a8 / ctx_1e0467a3f24a` 已完成自测并交还窗口，保持静态等待最终运行结果；独立 Reviewer `task_8149678f3fac / ctx_3406e529baa6` 正在唯一实际服务内只读复验。均为真实 Orca 启动的 Codex / gpt-6.1-sol。实际服务由独立普通终端 `term_c043b5a9-d4f4-41d0-b387-866ed544ca6d` 保持。上一轮failed/released及旧未结项继续按集成清单保留。

S3 预算单位、权限证据范围与受限 Codex 选择仍未回答，Claim 拒绝及 Known=false 保持。具体资料准入、Agent 读取完成证据、经验写入交接的方式、Laya 语言、首发范围、安装与服务共存和作者许可继续见 [QUESTIONS](docs/QUESTIONS.md)。实际用户作品、自由工具授权和正式实验缺失字段保持未知。

## 文档维护

新版 [AGENTS](AGENTS.md) 与四份 [角色说明](docs/agents/README.md) 保持用户原文，CODING、RESOURCES 提供统一规范。全仓 141 份原文档的处置见 [报告](docs/documentation-audit-20261020.md)：14 份旧执行正文归档，旧路径保留入口；17 份归档文本包含修订前 README、SPEC 和 TASKS。独有需求、冻结参数与真实证据继续保留。[docs/STATUS.md](docs/STATUS.md) 只保留本页入口；历史状态原文见 [首次归档](docs/archive/20261020/README.md)。本轮文档检查不增加功能验收结果。
