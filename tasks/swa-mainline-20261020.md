# SWA 主线重构（2026-10-20）

## 目标与完成条件

用户明确要求“以 re.md 为准重构项目主线”，随后明确首版目标为抽取 Orca CLI 对应执行能力，交付从原有 Agent 或终端调用的可控 Swarm CLI。首版以 [CLI 任务](swa-cli-v1-20261020.md) 为准。规划原文已保存为 [META_PLAN](../docs/META_PLAN.md) 2.1，M0—M6 保留总体整合方向；MCP 扩展和完整原生 S3 验收分别接续。普通 main 集成遵循 §9.1、§10。各工作包使用连续任务记录，当前执行按最新首版目标排序。

## 基线与当前事实

- 接管工作树：`C:/Users/DW/orca/Astrocyte`，`validation/v1-20261012@c8df672770ddfd5e9f53ea64fdcd932d2e4bff76`，接管时干净。
- 2026-10-20 真实 GitHub API：仓库 ID `1410836101`，规范名称 `songconmaisaix31-design/swarm-anything`，默认分支 `main`。
- 真实远端：`main@3a02ce38481d1885adee2e051a1b524d78376897`；产品分支与前端分支均为 `32ab1eec808a5848aabad40f11133735a1214e55`；主控分支为上述 c8df。产品代码作为重构起点。
- 复用 [既有集成清单](../docs/integration-audit-20261020.md)、[交付候选](delivery-candidate-20261020.md) 与独立审查；历史来源保持。主控分支包含独有决定、实验及研究记录，按内容处理，禁止用旧程序覆盖产品代码。
- 运行源码、程序及原 SQLite 依照交付候选记录重新核查。服务停止、串行验证与恢复已有授权见 QUESTIONS 的“当前交付候选与真实产品开发主线”；资源额度引用 [RESOURCES](../docs/RESOURCES.md)。

## 执行轨道与文件归属

M0 文档整合与 M1 契约故障的只读调查可以并行，分别由一名 Worker 负责，合用一名独立 Reviewer；其余实现按实际依赖安排。两条轨道分别修改治理文档和契约工具，任务文件互斥；共享状态由主 Agent 维护。主 Agent 管理范围、状态、决定、资源与最终检查，保持任务分支在交付后结束。

| 轨道 | 写入范围 | 结果与验收 | 依赖与停止条件 |
| --- | --- | --- | --- |
| 主 Agent | 本任务、主控的 META_PLAN、QUESTIONS、STATUS、HANDOFF、TASKS；忽略目录 `.verification/` | 维护事实、派发、实际读取并回复 review hook、最终检查和集成 | 不编写业务代码；不替代独立接受 |
| M0 Worker | 独立产品基线工作树；AGENTS、CODING、SPEC、角色文档、README/CONTRIBUTING、文档目录、M0 记录、整合清单、当前状态与决定；需要迁入的主控文档及原件 | 修订 §7.5、保留有效决定和历史来源、确定 main 纳入范围；文档与本地链接检查 | 先交测试计划独立审核；只修改文档；不改变程序、冻结实验、数据或发布状态 |
| Reviewer | 独立工作树；`docs/standards/swa-mainline-review-20261020.md`、相关独立报告；`.verification/` | 自行推导验收、审核计划、对准确提交发布 CR，执行已有 review-decision hook | 不修改被审实现；发现领域问题交原 Worker |
| M1 Worker（M0 后） | 原产品 Worker 工作树；根 package 命令、契约工具及必要依赖、对应任务和报告，具体路径由根因方案确定 | 原始 `pnpm check:contracts` 正常完成；保留失败原件、实际错误语义和生成漂移检查 | 根因方案与测试计划通过后实施；主 Agent 授予唯一重操作窗口 |
| 后续工作包 | M2、M3、M4、M5、M6 的范围按 META_PLAN 依赖派生 | 公共契约、真实 CLI/MCP、安装后候选与既定 S3 验收 | 共享契约先确定；不依赖未决项的工作继续 |

整合顺序：M0 文档及有效主控内容 → 已验证产品历史进入 main → M1 契约修复 → M2/M3 对应入口 → M4 已授权原生路径 → M5 候选包 → M6 真实任务。每次只集成一个分支；不 force push，不删除原分支、工作树及未提交内容。

## 验证、授权与停止条件

- M0 按文档变化检查来源、链接、规则一致性和 `git diff --check`；通过现有成熟 Markdown 库解析，资源按 RESOURCES。文档变更复用有效程序验证，不重复构建。
- M1 先审核根因与测试计划；真实运行原始契约命令、相关生成/适配器检查；不采用替身或绕过错误的 workaround。
- main 纳入产品历史及有效当前决定；现有数据库、运行目录、Go module 和旧配置兼容性独立处理，不从仓库改名推导数据迁移。
- S3 预算单位、权限证据覆盖、受限 Codex 工具关闭、资料/经验准入及真实用户采用沿 QUESTIONS，尚未回答的条目保持未决定。新入口不得伪造人类主体或批准。
- 发布和论文上传仍按各自原条件；未验结果保持 FAIL / NOT_RUN / BLOCKED。所有可执行范围完成后如实登记实际阻塞。

## 当前执行信息

- Orca Run：`run_62ada397a104`，主控 handle：`term_3209681a-faa5-4161-a2f3-9563e1451a33`。
- 主 Agent 客户端：当前 Codex；模型由当前会话提供。Worker/Reviewer 通过 Orca 启动，实际模型与工作树在启动后登记，不从请求参数推断。
- 当前状态：规划入仓，M0 基线已核，准备提交测试计划与独立审核；尚未改变 main 或程序。

| 执行者 | 工作树与分支 | Task / Dispatch | 实际客户端与模型 |
| --- | --- | --- | --- |
| M0 Worker | `swa-m0-20261020`，由产品32ab创建的任务分支 | `task_76639dc0a8b8` / `ctx_3d6edcaf3e8b` | Orca / Codex，实际状态报告 `gpt-6.1-sol` |
| 独立 Reviewer | `swa-mainline-review-20261020`，独立审查分支 | `task_1cd8014750f9` / `ctx_980a3282c339` | Orca / Codex，实际状态报告 `gpt-6.1-sol` |
| M1 契约 Worker | 原 `frontend-observatory-20261018` 工作树及分支 | `task_261b4dab6d89` / `ctx_958a2bf0ae27` | Orca / Codex，实际模型待会话状态报告 |

根主控已将 Git remote 更新为经实际仓库 ID 核实的规范地址 `https://github.com/songconmaisaix31-design/swarm-anything.git`。没有改变 Go module、数据目录、协议标识或已有客户端配置。
