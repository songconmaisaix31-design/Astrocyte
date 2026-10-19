# 当前交付候选与实际运行收敛（2026-10-20）

| 项目 | 执行定义 |
| --- | --- |
| 用户目标 | 总控接手，明确一个当前交付候选，使集成代码、验证记录和实际运行版本对应，收敛已有成果，建立真实产品开发主线 |
| 接管基线 | 主控 `validation/v1-20261012@3484d106e0cc4456bbae51a6488630b9facf7a89`；产品候选 `songconmaisaix31-design/frontend-observatory-20261018@5725c1840a760105518d273eaaa28d7367f369ed`；原集成 `songconmaisaix31-design/works-relay-integration@f17a491cdd6392829094c95a3e5efb6011362bc0`，三处工作区干净 |
| 问题 | 最新产品源码、Source15 验包与 c5 程序 / 172bb445 Web 分别来自不同版本；历史限定接受不能用作最新源码的完整验收 |
| 建议与预期效果 | 沿原产品 Worker 和 PR #2 收尾，独立审核验证计划及准确源码；原集成分支成为唯一产品开发主线，保存准确候选源码、构建产物、SQLite 与 HTTP 运行结果的对应关系 |
| 主控写入范围 | 本文件、README.md、HANDOFF.md、STATUS.md、docs/README.md、docs/TASKS.md、docs/QUESTIONS.md、docs/integration-audit-20261020.md；核查中间文件仅放 `.verification/delivery-candidate-20261020/`；已审服务脚本的运行证据位于产品树本次original-01目录，部署产物限原运行根新版本目录 |
| 产品 Worker | 沿原 `frontend-observatory-20261018` 工作树及分支；延续原产品模块写入范围，新增本轨测试计划、交付报告和必要验证入口；先提交计划接受独立审核，获得重操作窗口后才编译和运行 |
| Reviewer | 沿原 `research-preprint-review-20261019` 审查工作树及分支，限定本次产品审查报告与忽略目录；独立推导覆盖要求、审查计划和准确源码，按真实证据记录接受 / 退回 / 证据不足，提交 CR 并执行既有 hook |
| 并发与顺序 | 仅产品 Worker 与独立 Reviewer；产品文件和审查报告互斥。静态审核可并行，重操作全局串行。计划审核 → Worker 自测及构建 → Reviewer 可部署范围复验 → 主控同来源程序启动及核实 → Worker 最终报告与准确 CR → 主控一次推进原集成分支并普通 push。出现业务问题交回原 Worker |
| 服务窗口 | 用户本次要求实际运行版本对应，加上 2026-10-16 普通集成与启动授权，允许为本项目候选切换安排原 64505 服务的停止、验证和恢复窗口。仅操作已核实归属的本项目程序；保留真实数据和原件，仍执行 RESOURCES 的串行、采样和优先级要求 |
| 验收 | 明确一个准确候选源码与一个集成分支；默认检查和必要专项检查均保存真实命令 / 状态 / 限制；程序与 Web 从同一源码构建；实际服务 PID、路径、版本、地址及原真实数据库可核查；独立 Reviewer 对本次范围作出结论；主控记录及集成分支普通 push |
| 停止条件 | 以上对应关系可核查，或者准确记录无法继续的真实资源 / 授权 / 功能阻塞。完整 CI、原生 / S3、两 Agent 闭环与实际采用继续按原标准判断；未答业务选择保持未决定 |
| 使用的客户端与模型 | 主控、Worker 和 Reviewer 均使用 Codex 0.162.0 / GPT-6.1-Sol xhigh fast，沿用户配置，无模型覆盖；主控终端及 Worker / Reviewer 的 Orca projection 已核实实际模型 |

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

## 已执行节点

- 修订计划准确来源 `ed543c12106c2dbb3e181cb9bc1647b626455179`，仅产品报告后继，作为本次冻结构建源码 S。独立 Reviewer 接受计划的 hook 为 `msg_4c93b946d1de`，主控实际读取并回复 `msg_5320964d4f66`；源码和运行结果尚待验收。
- 必要交付使用现有 `pnpm build` 生成普通 Go / Web 产物及本树交付目录；Source15 完整便携包保持历史来源，本次不以完整模型包或重复130项验包作为前置。
- UTC `2026-10-19T19:25:06.2015728Z`，主控核准确路径、创建时间和唯一监听后停止原 PID45676，确认退出及64505端口空闲。原库和配置保留；实际记录在 `.verification/delivery-candidate-20261020/stop-original-service.json`、`running-before.json`。
- 主控以 `msg_165526d39873` 将唯一阶段 B 窗口交产品 Worker；Reviewer 保持静态。原真实库只读备份到 Worker 自有副本可执行，外部运行根写入和64505启动仍由主控控制。
- 静态核实产品树仍包含旧版 AGENTS、历史 STATUS/HANDOFF 和已过期 README 入口，缺少当前 CODING、RESOURCES 与角色文件。最终报告阶段追加这些治理文档的产品树写入范围：`AGENTS.md`、`README.md`、`HANDOFF.md`、`STATUS.md`、`docs/CODING.md`、`docs/RESOURCES.md`、`docs/QUESTIONS.md`、`docs/TASKS.md`、`docs/README.md`、`docs/agents/**`。同步用户已确认的规则和实际候选入口，保留历史 Git 来源及独有使用说明；不修改业务规格和构建输入，不重复构建。主控管理树仍由主控维护。
- 产品主线最终普通 push 后，将 Orca 的后续新工作树基准设为 `origin/songconmaisaix31-design/works-relay-integration`；已存在的树和分支归属保持原记录。

## 当前真实验证结果

原件根为产品树 `.verification/delivery-candidate-20261020/`，普通产物为 `dist/delivery-candidate-20261020/`。源码 S 的三个 Go 程序均带 `vcs.revision=ed543c12106c2dbb3e181cb9bc1647b626455179` 和 `vcs.modified=false`；38项构建复制内容逐字节一致。

| 命令或检查 | 已核实际结果 |
| --- | --- |
| `pnpm build` | EXIT0，三个 Go 程序、TypeScript / Vite 完成；保留主分块提示 |
| `go test ... -run '^TestDynamicDAG'`、`go test ... -run '^TestOpenViking'` | 原计划的9函数、4函数通过，实际 HTTP / SQLite 原件保留 |
| 三文件 Vitest、两视口 Dynamic DAG / Observatory Playwright | 53项、4项、8项通过；非默认完整浏览器矩阵 |
| 架构、依赖、`validate-examples`、`generate-client --check` | EXIT0；架构实际 Go1 / Node512，示例264项通过 |
| `pnpm check:contracts` 及补充 Redocly 同参数直调 | 两次分别记录实际 `UV_HANDLE_CLOSING` 断言崩溃，退出 `-1073740791`；不能记为通过。后两条既有契约脚本分别通过 |
| lint / vet | 各 EXIT0，实际运行重叠，不满足全局串行条件；原件及偏差保留，不为改变记录重复运行 |
| 原库一致备份 | SQLite backup API 与全列 / 类型 / 重复数比较通过，68表、26迁移；原库尚未写入候选迁移 |
| 副本迁移预演 | 同源实际程序及44次 HTTP / 全静态字节核查完成，进程已退出；首次全值比较因既有启动刷新更新 `local_project_discovery.data.observed_at` 而 FAIL。仅该字段和准确027记录变化，其他原数据一致 |

副本观察时间的补充计划已由独立 Reviewer 以 `msg_80e2d54c6380` 接受，主控实际读取并回复 `msg_03ff818849eb`，以 `msg_d44def0f154a` 通知原 Worker。仅允许持真实 launch 核唯一 registered 行时间的实际区间及其余字段完全保全，只读复核原副本并另存结果，首次 FAIL 保留，不重启服务或重跑 HTTP。实际迁移及主控原库启动尚待复验。

## 主控实际运行

独立可部署复验的 hook `msg_884a33f8ba50` 已实际读取回复 `msg_4bf8a8b91c47`，Reviewer以 `msg_533f7d4bb061` 交还唯一窗口。主控执行已审 `candidate-service.py stage` 和 `start --execute-original` 均 EXIT0；产物及原库备份、launch/http/migration、Q指针和启动采样保存在产品树 `.verification/delivery-candidate-20261020/original-01/`。

- 运行源码与 Web：`ed543c12106c2dbb3e181cb9bc1647b626455179`，部署目录 `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/versions/ed543c12/`。
- 服务：`http://127.0.0.1:64505/`，PID45956，创建UTC `2026-10-19T19:47:04.6572166Z`，BelowNormal；唯一loopback监听及启动工具退出后health正常已核。
- 原数据库路径保持，69表、27迁移；原业务值 / 类型 / 重复数保全，新增OpenViking表为空，唯一registered观察时间为实际启动观察；Q完整指针及全部静态资源字节一致。
- 独立普通服务终端：`term_c043b5a9-d4f4-41d0-b387-866ed544ca6d`；该终端与Worker释放分开管理，保留实际服务运行。
- 主控以 `msg_fa5e8f215495` 将唯一服务内只读复验窗口交Reviewer，Worker保持tracked干净；独立运行、最终治理文档与准确CR尚待结算。
