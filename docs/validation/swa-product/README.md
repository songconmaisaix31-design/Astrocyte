# SWA 产品验收

本目录由第06卡负责。开发基线为 `6195f4cfb6b32aedd43845443c029ecc11d5578d`，具体计划见[本轮计划](plan-20261022.md)。历史安装和两 Worker 结果保留其原范围。

`tests/swa-product-acceptance/preflight.mjs` 只读观察当前平台、明确选择的安装根、目录、Agent 文件及变量存在状态。支持观察 Windows x64 和 Linux x64 glibc；新候选的 `PLATFORM` 为 `win32-x64` 或 `linux-x64-glibc`，缺失时仅兼容原 Windows 候选，不能将其当成 Linux 安装。它不申请 Grant，不读取凭据文件，不调用模型，不执行 Worker，也不确认产品验收通过。`SWA_PROJECT_TOKEN` 已设置仅证明变量存在。

在仓库根目录运行：

```sh
node tests/swa-product-acceptance/preflight.mjs --workers 3
```

当前平台具有实际候选和合法当前授权后，可以指定全部路径。`--agent` 是明确选择的原生可执行文件绝对路径，`--model` 是对应已批准模型名称。私有数据根、安装根和公开测试项目必须分离；Linux 数据根还要求当前 uid 所有、0700。凭据通过受保护运行环境注入，不放入命令参数、报告或 Git。

```powershell
node tests/swa-product-acceptance/preflight.mjs --candidate $InstallRoot --data-root $DataRoot --project $PublicProject --agent $AgentExecutable --model $ApprovedModel --workers 3
```

退出码 `2` 表示参数无效；`1` 表示明确选定的实际输入检查失败；`3` 表示验收尚未运行或前置未完整验证。该脚本始终保留真实产品场景为 `NOT_RUN`，不运行其他平台的候选。主机平台已观察不代表已有对应候选。最小文件及版本观察不能代替 `check:swa` 闭包、权限、实际完整所属树和模型验证。

## 固定公开输入准备

用户已确认本轮真实模型测试费用不设上限。ToolsOff、固定公开输入、当前项目/模型/外发批准、原请求去重及资源退出要求继续适用；预算决定不会创建账号或项目 token，也不扩大私有资料范围。跨创建者的新许可仅为人类独立当前权限下的 `worker.send`，不包括 `worker.start`。

准备器通过现有 HTTPS Git 代理读取固定公开 stats 项目的准确提交，核对实际四文件，不执行 hook、子模块或项目脚本。选择仓库外尚不存在的专用目录：

```sh
node tests/swa-product-acceptance/prepare-public-case.mjs --root /workspace/swa-public-case-20261022 --workers 3
```

实际生成任务规格、N份不同职责的公开输入、同会话后续输入及来源记录；这些均是验收输入，不是原生身份、事件、状态或模型结果。目录已存在时拒绝覆盖。此准备器按当前完整成果100项观察边界接受 N=3..100；这不是产品 `max_workers` 的硬上限。更大的完整验收需要先解决已有分页/完整库存限制。

当前实际公开案例为 `/workspace/swa-public-case-20261022/public-project`，提交 `da3bdddb085808863090a05f60cc1c64415039cf`。2026-10-22实际准备退出0，已有目录拒绝退出1，N=2/3.5/101拒绝退出2，未调用模型。Linux后继前置观察退出3，主机平台匹配但没有据此声明安装候选或里程碑通过。日志在 `.verification/swa-product-acceptance/`。

2026-10-22后续实际CLI04候选的源码为 `e14800c1b207a8c063531fedf014aa0cd2374558`，仓库外安装位于 `/workspace/swa-linux-installs/swa-linux-cli-e14800c1-04`。构建、归档校验和私有安装退出0，固定Node24.21.0的真实PTY输入/回显/退出、watcher创建事件及28253项闭包检查通过；安装根0700，属于当前uid。真实服务在18793完成同Node/daemon/HTTP readiness、health200及正常stdin关闭退出0，内部完整所属观察与外部实际进程身份均确认退出。证据分别位于 `.verification/swa-linux/evidence/e14800c1b207a8c063531fedf014aa0cd2374558-cli-04/` 和 `/workspace/swa-cli-smoke-e14800c1-1792661325429/smoke-evidence/`。

本卡对该实际安装物运行preflight退出3，版本和互不包含目录观察通过；当前项目Grant、Agent NativeLaunch/ToolsOff、真实多Worker和CLI/GUI里程碑仍为 `NOT_RUN`，没有模型调用。没有自动取得human会话、签发Grant或申请token。service级完整退出证据不等同每Worker完整资源证明。

现有真实验收入口为 `tests/swa-cli-headless/installed-flow.mjs`、`public-case.mjs` 和 `scripts/verify-swa-cli-headless.ps1`。它们当前具有 Windows 专用路径和两 Worker 条件，完整多 Worker 驱动须在公共契约及实际候选完成后由本卡接续，不能把现有脚本直接记为 N=3 通过。

## N Worker 安装验收驱动

`tests/swa-product-acceptance/run-installed.mjs` 实现同一Run/root Task下N项启动、全体真实delta公共交集、活动停止、其余原session继续、同会话后续输入、额度拒绝、原request幂等、完整原件和journal核对，以及真实Agent控制端业务审阅后的完成。它只附着明确服务，不自动启动第二runtime、取得human会话或签发Grant。

先用同安装物的固定Node查看 `--help`。完整调用中所有变量必须来自本次准确安装、当前合法项目授权及固定公开案例；`serviceLog`必须是本次私有dataRoot内实际serve日志，`newEvidence`为尚不存在的目录：

```powershell
& $installedNode tests/swa-product-acceptance/run-installed.mjs --candidate $installRoot --data-root $dataRoot --project $publicProject --project-id $projectId --endpoint $endpoint --agent $agent --model $model --output $newEvidence --service-log $serviceLog --workers 3
```

当前实现只消费已有Windows Pi/OpenCode原生事件和完整成果形状；Linux及Codex的完整原生来源/进程union消费者接入前，驱动在任何worker.start之前明确 `NOT_RUN`。当前无合法Grant时也不发送模型。该边界将随02实际来源合同和独立审查更新，不能靠移除门或填Windows创建时间运行Linux。

准备器和驱动共用固定公开定义；模型输入前逐字核对四份真实Git原文、完整任务规格（含空references）、各职责正文和后续正文。每条命令在启动前以0600独占保存原参数和request并同步磁盘；未知stop/release只查询该Dispatch已有的原request，不换键重新停止。已完成请求的回放同时核原回执、input operation和真实原生执行次数；继续输出的证据须晚于实际原生停止标记。报告记录安装物实际EDITION；缺少装内正式checker时保持NOT_RUN，不使用开发树脚本替换。

服务UUID与进程owner incarnation分别验证：CLI回执及实际Worker的`runtimeEpoch`必须属于开始时的服务；`source.process.runtime_id`则绑定本Session真实journal数据库记录中的`lease.ownerProcess.runtime`。首次保存进程身份前还核原start request的实际input binding及provider head原fence；每次事件读取前后重新核同一普通journal文件的必要record/head/lease。停止后只在原ownerFence的真实deathEvidence、已保存进程身份及原完整Job退出证明一致时读取已释放来源。驱动当前只消费Windows Pi/OpenCode已有neutral handles；Linux/Codex消费者仍须另审准确来源形状。

实际成果齐备后，驱动输出原结果路径，等待真实Reviewer/Agent阅读后输入一行JSON，字段为 `reviewed_result` 和 `checked_result`；只把真实审阅记录写入完成正文，不自动生成通过，也不取得human身份。未知请求只查询原request；缺少实际分块观察、未命中的效果丢失分支、过期/跨项目权限与GUI行为分别保留未验。

本轮实际 `node --check`、固定安装Node的 `--help` 退出0；workers=2/3.5、相对候选路径、未知参数各退出2。共享公开定义修改后，重新从固定真实来源准备独立案例退出0，并核既有案例与新案例的完整正文、任务和清单逐字一致，证据为 `.verification/swa-product-acceptance/review-driver-checks.json`。尚无本轮当前项目ID/Grant与可用Linux原生来源，完整N驱动正向未运行；修改前落盘、未知stop保键及原生回放次数的动态路径仍等待真实授权执行，不用替身验证。参数验证和语法通过不代表N=3通过。退出0仅表示当次已执行断言及成果交接完成，最终里程碑由独立审查决定。
