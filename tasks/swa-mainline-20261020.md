# SWA 主线重构（2026-10-20）

## 目标与完成条件

用户明确要求“以 re.md 为准重构项目主线”，随后明确首版目标为抽取 Orca CLI 对应执行能力，交付从原有 Agent 或终端调用的可控 Swarm CLI。首版以 [CLI 任务](swa-cli-v1-20261020.md) 为准，按固定抽取范围、无界面执行与控制端身份/CLI规则、多Worker、失败资源、安装包推进；安装后的同一候选须完成一次至少两个真实Worker的任务并验证并发、指定停止和成果获取。用户要求不妨碍本机Orca，产品按独立后台与运行隔离集成。M0已接受进入main；MCP、网页重构、完整旧S3和全部历史分支清理后置，旧S3保留独立验收记录，原M2建议保留此前调查。

原“以 re.md 为准重构项目主线”决定及 [META_PLAN](../docs/META_PLAN.md) v2.1 正文保留。普通 main 集成沿 §9.1、§10；原 M2—M6 及 §12 A01—A10 按各自后续范围处理，不自动扩大最新首版前置。当前任务是总入口，既有任务覆盖的内容继续使用原记录。

## 基线与当前事实

- 接管工作树：`C:/Users/DW/orca/Astrocyte`，`validation/v1-20261012@c8df672770ddfd5e9f53ea64fdcd932d2e4bff76`，接管时干净。
- 2026-10-20 真实 GitHub API：仓库 ID `1410836101`，规范名称 `songconmaisaix31-design/swarm-anything`，默认分支 `main`。
- 接管远端快照：`main@3a02ce38481d1885adee2e051a1b524d78376897`；产品分支与前端分支均为 `32ab1eec808a5848aabad40f11133735a1214e55`；主控分支为上述 c8df。产品代码作为重构起点。
- 主线集成：M0 准确 `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765` 获独立 COMMENT review `5581860208` 接受，原 hook 已由主控读取并回复。产品分支及默认 main 已普通快进并 push 到该来源，PR #4 实际已合入；没有程序变更。后续开发使用 `origin/main`，原实验根分支保留。
- 复用 [既有集成清单](../docs/integration-audit-20261020.md)、[交付候选](delivery-candidate-20261020.md) 与独立审查；历史来源保持。主控分支包含独有决定、实验及研究记录，按内容处理，禁止用旧程序覆盖产品代码。
- 运行源码、程序及原 SQLite 依照交付候选记录重新核查。服务停止、串行验证与恢复已有授权见 QUESTIONS 的“当前交付候选与真实产品开发主线”；资源额度引用 [RESOURCES](../docs/RESOURCES.md)。

## 执行轨道与文件归属

M0 已完成文档交付；首版 CLI 无界面方案与测试计划已独立接受，实施写域见首版任务，程序由同一Worker实现并交原Reviewer审查。M1 原命令本次未补丁即完整通过，历史崩溃未复现，已停止补丁实施、恢复原服务并完成限定调查验收；具体归属见下表及首版任务。共享状态由主 Agent 维护，主 Agent 管理范围、决定、资源与最终检查。

| 轨道 | 写入范围 | 结果与验收 | 依赖与停止条件 |
| --- | --- | --- | --- |
| 主 Agent | 本任务、主控的 META_PLAN、QUESTIONS、STATUS、HANDOFF、TASKS；忽略目录 `.verification/` | 维护事实、派发、实际读取并回复 review hook、最终检查和集成 | 不编写业务代码；不替代独立接受 |
| M0 Worker | 独立产品基线工作树；AGENTS、CODING、SPEC、角色文档、README/CONTRIBUTING、文档目录、M0 记录、整合清单、当前状态与决定；需要迁入的主控文档及原件 | 修订 §7.5、保留有效决定和历史来源、确定 main 纳入范围；文档与本地链接检查 | 先交测试计划独立审核；只修改文档；不改变程序、冻结实验、数据或发布状态 |
| Reviewer | 独立工作树；`docs/standards/swa-mainline-review-20261020.md`、相关独立报告；`.verification/` | 自行推导验收、审核计划、对准确提交发布 CR，执行已有 review-decision hook | 不修改被审实现；发现领域问题交原 Worker |
| M1 只读阶段 | 原产品 Worker 工作树的两份M1文档及自有忽略目录；已普通 push `80d4f8d1e323982ef73d880ce9927f4bc5c58340` | 根因与测试计划已独立接受，实际修复尚未实施；原 FAIL 保留 | Task 已结算、终端已释放；实际修复需重新授予程序写入范围及唯一窗口 |
| M1 接续 Worker（已结算） | 独立工作树；既定程序写域未实施，只更新两份 M1 文档及自有忽略目录 | 原冻结安装及未补丁完整命令均 EXIT0；准确b112调查记录已独立接受，修复因果证据不足，T02—T08未执行 | 唯一窗口已交还，主控 T09 恢复同一 S 与原库成功；Task结算、终端released |
| 首版 CLI 来源调查 Worker（已结算） | [首版任务](swa-cli-v1-20261020.md) 登记的两份来源方案/报告及忽略目录 | 准确30b的调查、设计提案及覆盖框架已独立接受 | 此前Mission/Claim提案及其待决项按适用任务保留；当前实现沿已审无界面方案接续 |
| 无界面执行 Worker | 独立 `swa-cli-headless-20261020` 工作树/分支；首版任务授权的逐文件程序写域、任务/报告及自有忽略目录 | 准确396e方案及T00—T15已接受；53ad程序源码正在CR，实际构建/模型/安装及双Worker验收均未通过 | 已结束首轮依赖窗口并恢复原S12800；第二次prepare因可用内存不足在停服前拒绝，无新动态窗口，继续静态工作 |
| 后续工作包 | M2、M3、M4、M5、M6 的范围按 META_PLAN 依赖派生 | 公共契约、真实 CLI/MCP、安装后候选与既定 S3 验收 | 共享契约先确定；不依赖未决项的工作继续 |

先完成 M0 文档及有效主控内容，由主控安排已验证产品历史进入 main。最新首版 CLI 按主控核实后的实际任务推进；原 M1 契约修复及 M2—M6 分别沿各自依赖与授权处理，M2/M4 原独立工作不依赖 M1 修复完成。每次只集成一个分支；不 force push，不删除原分支、工作树及未提交内容。

## 验证、授权与停止条件

- M0 按文档变化检查来源、链接、规则一致性和 `git diff --check`；通过现有成熟 Markdown 库解析，资源按 RESOURCES。文档变更复用有效程序验证，不重复构建。
- M1 先审核根因与测试计划；真实运行原始契约命令、相关生成/适配器检查；不采用替身或绕过错误的 workaround。
- main 纳入产品历史及有效当前决定；现有数据库、运行目录、Go module 和旧配置兼容性独立处理，不从仓库改名推导数据迁移。
- S3 预算单位、权限证据覆盖、受限 Codex 工具关闭、资料/经验准入及真实用户采用沿 QUESTIONS，尚未回答的条目保持未决定。新入口不得伪造人类主体或批准。
- 发布和论文上传仍按各自原条件；未验结果保持 FAIL / NOT_RUN / BLOCKED。所有可执行范围完成后如实登记实际阻塞。

## 当前执行信息

- Orca Run：`run_62ada397a104`，主控 handle：`term_3209681a-faa5-4161-a2f3-9563e1451a33`。
- 主 Agent 客户端：当前 Codex；模型由当前会话提供。Worker/Reviewer 通过 Orca 启动，实际模型与工作树在启动后登记，不从请求参数推断。
- M0 准确文档已独立接受并普通快进进入 main，原 PR 已合入；当前版本与已验范围统一见根 STATUS。首版执行能力由同一Worker实现并交原Reviewer审查，动态验收仍未通过。

| 执行者 | 工作树与分支 | Task / Dispatch | 实际客户端与模型 |
| --- | --- | --- | --- |
| M0 Worker | `swa-m0-20261020`，由产品32ab创建的任务分支 | `task_76639dc0a8b8` / `ctx_3d6edcaf3e8b` | Orca / Codex，实际状态报告 `gpt-6.1-sol` |
| 独立 Reviewer | `swa-mainline-review-20261020`，独立审查分支 | `task_1cd8014750f9` / `ctx_980a3282c339` | Orca / Codex，实际状态报告 `gpt-6.1-sol` |
| M1 只读 Worker（已结算） | 原 `frontend-observatory-20261018` 工作树及分支，材料已普通 push | `task_261b4dab6d89` / `ctx_958a2bf0ae27`，终端 released | Orca / Codex，实际模型 `gpt-6.1-sol` |
| M1 修复 Worker | `swa-m1-contracts-implementation-20261020` 工作树及同名分支，起点main4132 | `task_ceb3b3bb76be` / `ctx_8d6a6e1c0145` | Orca / Codex，实际模型 `gpt-6.1-sol` |
| 首版 CLI 执行 Worker | 同一产品工作树，新任务分支 `swa-cli-execution-20261020` | `task_92402bd9f0c7` / `ctx_974917191668` | Orca / Codex，实际模型 `gpt-6.1-sol` |
| 无界面执行 Worker | `swa-cli-headless-20261020` 工作树及同名分支 | 当前`task_9f062888e786 / ctx_0d069056baff`；原静态Task450/ctxdc已结算，资源retain以保留同一活动终端 | Orca / Codex，实际终端 `GPT-6.1-Sol xhigh fast`；原启动观察保留，当前归属见首版任务 |

根主控已将 Git remote 更新为经实际仓库 ID 核实的规范地址 `https://github.com/songconmaisaix31-design/swarm-anything.git`。没有改变 Go module、数据目录、协议标识或已有客户端配置。
