# 当前交付候选与实际运行收敛（2026-10-20）

| 项目 | 执行定义 |
| --- | --- |
| 用户目标 | 总控接手，明确一个当前交付候选，使集成代码、验证记录和实际运行版本对应，收敛已有成果，建立真实产品开发主线 |
| 接管基线 | 主控 `validation/v1-20261012@3484d106e0cc4456bbae51a6488630b9facf7a89`；产品候选 `songconmaisaix31-design/frontend-observatory-20261018@5725c1840a760105518d273eaaa28d7367f369ed`；原集成 `songconmaisaix31-design/works-relay-integration@f17a491cdd6392829094c95a3e5efb6011362bc0`，三处工作区干净 |
| 问题 | 最新产品源码、Source15 验包与 c5 程序 / 172bb445 Web 分别来自不同版本；历史限定接受不能用作最新源码的完整验收 |
| 建议与预期效果 | 沿原产品 Worker 和 PR #2 收尾，独立审核验证计划及准确源码；原集成分支成为唯一产品开发主线，保存准确候选源码、构建产物、SQLite 与 HTTP 运行结果的对应关系 |
| 主控写入范围 | 本文件、HANDOFF.md、STATUS.md、docs/TASKS.md、docs/QUESTIONS.md、docs/integration-audit-20261020.md；核查中间文件仅放 `.verification/delivery-candidate-20261020/` |
| 产品 Worker | 沿原 `frontend-observatory-20261018` 工作树及分支；延续原产品模块写入范围，新增本轨测试计划、交付报告和必要验证入口；先提交计划接受独立审核，获得重操作窗口后才编译和运行 |
| Reviewer | 沿原 `research-preprint-review-20261019` 审查工作树及分支，限定本次产品审查报告与忽略目录；独立推导覆盖要求、审查计划和准确源码，按真实证据记录接受 / 退回 / 证据不足，提交 CR 并执行既有 hook |
| 并发与顺序 | 仅产品 Worker 与独立 Reviewer；产品文件和审查报告互斥。静态审核可并行，重操作全局串行。计划审核 → Worker 自测及构建 → Reviewer 复验 → 主控推进原集成分支并普通 push → 同来源程序启动及核实。出现业务问题交回原 Worker |
| 服务窗口 | 用户本次要求实际运行版本对应，加上 2026-10-16 普通集成与启动授权，允许为本项目候选切换安排原 64505 服务的停止、验证和恢复窗口。仅操作已核实归属的本项目程序；保留真实数据和原件，仍执行 RESOURCES 的串行、采样和优先级要求 |
| 验收 | 明确一个准确候选源码与一个集成分支；默认检查和必要专项检查均保存真实命令 / 状态 / 限制；程序与 Web 从同一源码构建；实际服务 PID、路径、版本、地址及原真实数据库可核查；独立 Reviewer 对本次范围作出结论；主控记录及集成分支普通 push |
| 停止条件 | 以上对应关系可核查，或者准确记录无法继续的真实资源 / 授权 / 功能阻塞。完整 CI、原生 / S3、两 Agent 闭环与实际采用继续按原标准判断；未答业务选择保持未决定 |
| 使用的客户端与模型 | 主控使用当前 Codex / GPT-6；Worker 和 Reviewer 均为 Codex 0.162.0 / gpt-6.1-sol，沿用户默认配置，无模型覆盖；Orca projection 已核实实际模型 |

## 合并顺序

1. 原 PR #2 产品分支：计划、必要返修、自测与独立审查。
2. 原 `works-relay-integration`：仅接入本次获准准确产品候选，不并入正式 V4 或论文分支。
3. 主控管理记录：写入准确结果和下一项真实产品任务。

接管核查已确认原服务 PID `45676`，监听 `127.0.0.1:64505`，程序路径 `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/versions/c5b3b571/astrocyte.exe`。运行窗口开始前重新核实进程归属及数据库路径。

## 本次协作

Run `run_1187287582ea`，主控 `term_136af66c-7d90-4921-9568-520883163990`。

| 角色 | Task / Dispatch | 原工作树与当前状态 |
| --- | --- | --- |
| 产品 Worker | `task_6f09f37378a8` / `ctx_1e0467a3f24a` | `frontend-observatory-20261018`，静态编写验证计划 |
| 独立 Reviewer | `task_8149678f3fac` / `ctx_3406e529baa6` | `research-preprint-review-20261019`，静态推导验收条件；首次启动未观察到提交，读取终端确认任务留在输入区后仅补一次 Enter，随后 projection 确认 live，没有重复派发 |

已核原 c5 与候选差异包含新增 `027_openviking.sql`。服务切换须通过成熟 SQLite backup API 保存原库，并分别核查既有业务表保全、准确迁移增量及新增表；旧恢复脚本的全部表计数相等条件无法直接用于该迁移。
