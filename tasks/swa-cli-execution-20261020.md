# 首版 Swarm CLI 执行模块：来源、方案与测试计划

## 本次派发

| 项目 | 事实 |
| --- | --- |
| 范围 | 真实来源调查、完整方案与实施前测试计划；程序实现及动态检查未获本次授权 |
| 首版事实源 | 主控树 `tasks/swa-cli-v1-20261020.md`；`docs/QUESTIONS.md@9140cf8aa618c832b3a099222a46991b7c2b3ddd` |
| 总体规划 | 主控树 `docs/META_PLAN.md@044a6dd688389a9354e23e3030534ac354e97f12` §0、§3、§5—8；首版任务覆盖相应旧完成前置 |
| Run / Task / Dispatch | `run_62ada397a104` / `task_92402bd9f0c7` / `ctx_974917191668` |
| Worker / 主控 | `term_4b891ad8-d0df-4e7b-bed4-670800745519` / `term_3209681a-faa5-4161-a2f3-9563e1451a33` |
| 客户端与模型 | Orca 1.4.230 中的 Codex CLI；当前终端信息显示 `GPT-6.1-Sol xhigh fast`，此为本次实际执行客户端观察 |
| 工作树 | `C:/Users/DW/orca/workspaces/Astrocyte/frontend-observatory-20261018` |
| 分支 / 产品基线 | `swa-cli-execution-20261020` / `32ab1eec808a5848aabad40f11133735a1214e55` |
| write_paths | 本文件、`docs/standards/swa-cli-execution-20261020.md`、`.verification/swa-cli-execution-20261020/**` |

接管已完整读取本树 HANDOFF、STATUS、AGENTS、CODING、RESOURCES，并核对主控当前首版任务、QUESTIONS 和 META_PLAN 的相关章节。旧分支接管时 `git status --short` 为空，HEAD 为 `80d4f8d1e323982ef73d880ce9927f4bc5c58340`；已执行普通 `git push origin HEAD`，远端旧分支由 `32ab1eec` 推进到 `80d4f8d1`。随后使用 `git switch -c swa-cli-execution-20261020 32ab1eec808a5848aabad40f11133735a1214e55` 创建新任务分支，没有改写历史。M1 静态材料保存在旧分支准确 Commit；原 Task/Dispatch 已结算，本次不复用其生命周期身份、不更新旧 PR #2。

主控随后通知默认 main 已普通推进到 `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765`。本次已只读其最新 AGENTS、CODING、SPEC 变更和首版任务；按当前 Domain/App/Adapters/Composition 与变更边界检查规则继续。主控在实施前安排普通合入该 main，本次不修改治理文档或自行改变任务基线。

## 问题与建议

上游准确来源为 `stablyai/orca@b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`（`v1.4.230`）。本地只读源码目录为 `C:/Users/DW/orca/Astrocyte/.verification/orca-source-v1.4.230`，本次重新核对 HEAD 与干净工作树。MIT 版权归属为 `Copyright (c) 2026 Lovecast Inc.`。

建议采用一个业务架构：`swa` CLI 调用现有 Go 后台；Workspace 负责项目、批准、上下文与会话绑定；Swarm 继续独占 Mission、WorkItem、Lease、结果和资源预留；Orca 来源代码或公开接口只提供所属进程与会话的执行事实。完整依据、执行链和测试计划见[分析报告](../docs/standards/swa-cli-execution-20261020.md)。

上游已有纯 Node `orcad`，可以脱离 Electron 运行，但它会组装完整 `OrcaRuntimeService`、profile SQLite、Agent hooks、账户服务、PTY daemon 和 RPC。`@orca/process-host` 的真实位置为 `src/packages/process-host`，它具有独立包边界；完整 Agent 执行模块尚在该包以外。`orchestration.workerStart` 会创建 Orca Run/Task/Dispatch；不能把它与现有 Swarm 任务各自作为同一业务的权威。

部署边界仍待用户回答：独立运行需要抽取必要 Node 执行模块与宿主依赖；依赖已安装 Orca 则使用经能力协商的低层会话/进程接口。两者对业务状态归属、权限和验收的共同约束已明确；依赖该答案的执行协议、安装产物集合与程序 write_paths 保持未定稿。推荐独立 Node 执行宿主作为源码抽取方向，此推荐不构成用户选择。

## 预期用户结果与验收

1. 用户在普通终端或已有 Agent 的命令工具中使用相同 `swa` 入口，提交获准项目目标、查询真实阻塞、取得结果，并控制已有执行。
2. CLI 退出与断开不终止后台任务；重复请求保持既有业务与执行身份，未知效果不自动重发。
3. 启动接受、输入接受、原生回合开始、成果提交、用户完成确认、进程退出、资源释放分别保留真实事实。
4. 停止、原生接续和上下文交接核验完整所有权；旧执行者、身份变化和停止未知时拒绝有冲突的新动作。
5. 使用真实已安装 Agent、Agent 创建的固定公开 Git 项目、真实 SQLite、实际 HTTP 与实际子进程验收；输出包含真实未知量及未执行检查。
6. 至少两名真实执行者分别可查询和读取输出；停止准确指定的执行者后，另一执行者保持原身份、输出和继续工作能力。该场景在唯一受管服务与资源窗口内安排，不能用单一进程启动替代。

首版目标并未授予 Agent 冒充 human 的权限。当前 Relay 控制入口整体要求 human；Swarm 创建/取消 Mission 也有明确主体限制；实际 `proposal_approved` 消费者尚未在生产调用中找到。报告列出沿现有批准规则补齐消费者与控制入口的具体方案，以及需要主控处理的权限差异。Codex 受限原生文件工具准入仍拒绝；源码抽取不能移除该条件。

本报告推荐 Mission/Claim 归属会直接依赖现有调用预算单位和执行权限证据两项待答。正式答复前 Claim 拒绝、Known=false 保持，所以该方案的首版正向认领/执行当前受阻；不能以全 nil 预算、独立 Relay 替代或新账本绕过。完整旧 S3 的成果采用和经验没有加入本次验收。Claude/Pi/OpenCode 的真实候选证据和各自限制分别写在报告，不把 Codex 单项拒绝推广为所有 Agent 不可执行。

## 执行顺序与停止条件

| 阶段 | 当前状态 / 下一动作 |
| --- | --- |
| 来源与依赖 | 本次完成源码读取及静态分析；保留功能可用性未验 |
| 方案 / 测试计划 | 提交准确 Commit 给同一 Run 的主控，由独立 Reviewer 推导条件后审核 |
| 决策与写域 | 主控取得执行边界答案并处理报告中的权限、工具/预算适用范围后，另行批准程序 write_paths |
| 实现 / 自测 | 本次未执行；后续由同一模块 Worker 负责代码、测试、文档和返修 |
| PR | 实现并完成适用自测后新建首版 PR；本次只普通 commit/push 调查材料 |

本次禁止安装、编译、模型调用、运行上游源码、启动/停止服务及功能测试；没有使用视觉功能或另起 Agent。资源窗口不由本次调查占用。未来真实验证须由主控串行安排，按 RESOURCES 执行 BelowNormal、CPU/内存双采样及 Go/Node 限额。任何未决授权、真实来源不满足、不能确认原进程退出、当前 Grant 失效或预算/能力证据不足，都停止依赖动作并保留未知与占用。

## 本次检查与交付记录

- `git status --short`、`git log -1`、`git branch --show-current`：确认接管基线、新分支与干净状态。
- `git push origin HEAD`：旧 M1 分支普通推送成功；新报告沿普通 push 方式交付，准确 SHA 通过本 Dispatch 回报。
- `git -C <上游目录> rev-parse HEAD` / `status --short`：准确来源为 `b2879bd5...`，源码树干净。
- `orca skills get orca-cli` / `orca skills get orchestration`：读取当前安装版本指南；`status --json` 与当前 `terminal show` 为本次只读环境观察。
- `Get-Content` / `rg`：读取报告列明的完整核心文件和现有产品对应实现；没有执行来源代码或其构建脚本。
- `node --max-old-space-size=512 .verification/swa-cli-execution-20261020/check-doc-links.mjs`：BelowNormal，使用已安装 marked 18.1.0 解析两份 Markdown，3 个本地链接全部可访问，PASS；没有安装依赖。
- `git diff --cached --check`：BelowNormal，提交前对两份完整已暂存文档检查，无错误，PASS。
- 安装、构建、功能测试、模型调用、候选安装包验收及 M1 动态检查：全部 `NOT_RUN`。

本文件与分析报告是本次可审阅交付物。独立 Reviewer 接受前不把方案、测试计划或首版功能登记为已验收；准确交付 SHA 通过同一 Dispatch 的 Orca `ask` 和最终 `worker_done` 回报。
