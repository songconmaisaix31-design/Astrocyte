# 2026-10-18 DDD 开发与审查整改

用户要求按照开发规范继续代码开发，采用既有 DDD 分层。基线：管理分支 `validation/v1-20261012` / `734c9b0d0ed357c159bd2127bb63b79e4c6667a9`；产品集成分支 `songconmaisaix31-design/works-relay-integration` / `8c8510426b828359f4c161f93c1edf1083fd5e88`，均干净。本轮 review 原件在根 `.verification/review-20261018/review.md`。当前 64505 无监听，旧服务运行记录按历史保留。正式实验与未答决定依根 QUESTIONS。

用户要求继续开发并避免占满电脑资源。当前执行限制为一个开发 Worker、`GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`、BelowNormal，编译、测试和实际服务串行，临时 HTTP 服务最多一个。每项重操作前两次核 CPU 低于 70%、可用内存至少 4 GiB；不满足时继续静态审查和文档。Go 内存目标属于软限制，实际峰值与可用资源另行观察；不重做完整模型包或并行安装依赖。资源守卫曾拒绝启动的检查按未执行记录保留。

17:17 为在资源守卫生效期间继续实施，将尚无修改的 T/P 写域合并到当前 D Worker、`ddd-domain-20261018` 工作树和同名分支；D 成为本轮唯一开发写入者，同一 Worker 继续代码、测试、报告和返修。新增互斥写域仅为原 T/P 表中已授权路径：`.github/workflows/**`、`scripts/check*.mjs`、`release/**`、`docs/standards/ddd-validation-20261018.md`、`docs/release/windows.md`、`docs/release/ddd-release-20261018.md`。原 T/P Task 记录为已合并处理，不启动其他开发 Agent，原空工作树保留。资源不足时可继续静态代码和文档修改，全部编译、真实服务与验证等待资源恢复；完成后仍由唯一集成 Agent 合并并由主控复验。当前 D Dispatch 为 `ctx_5242bb05a538`，终端 `term_e0771644-b770-46ef-b28a-89cfc660acd1`。

本轮原三个隔离工作树均从准确产品基线创建。Orca 重启后原 Dispatch 记为 failed / terminal_missing 并已官方 release；D 的五件未提交修改保留并继续完成，T/P 当时没有修改。当前下表三类写域均由同一 D Worker 在 `ddd-domain-20261018` 同树同分支处理；历史开发树、未接受提交和原 T/P 空树保留。主 Agent 维护计划、状态、决定及验收，开发完成后由唯一集成 Agent 在原产品集成树普通合并，主 Agent 最终复验、提交与普通 push。

| 轨道 | 固定工作树 / 分支 | write_paths 与结果 |
|---|---|---|
| D 领域与接入 | ddd-domain-20261018 / ddd-domain-20261018 | `internal/swarm/**`、`internal/workspace/{app,domain}/relay*.go`、`internal/adapters/sqlite/swarm*.go`、`internal/adapters/agents/relay*.go`、`internal/adapters/events/relay*.go`、`cmd/server/relay*.go`、`docs/standards/ddd-domain-20261018.md`。沿既有 prepared input 与成员端口修正首次 Claim；检查 archive 可信完整性来源和传递，只有来源可证时实施相应接线。消费方端口归 app，外部事实与持久化归 adapters，启动只做组装。未知事实和未答预算不补默认值。 |
| T 检查入口与真实测试 | 并入 ddd-domain-20261018 | `.github/workflows/**`、`scripts/check*.mjs`、`docs/standards/ddd-validation-20261018.md`；需要修改其他测试先交准确 Handoff。使 DDD 检查覆盖当前业务上下文，核默认检查所需真实前置；修正可独立验证的工程问题，保持全部原断言、来源范围与验收要求。Native 和模型来源不足时如实失败，不加入 skip、mock 或伪造素材。 |
| P 发布配置 | 并入 ddd-domain-20261018 | `release/**`、`docs/release/windows.md`、`docs/release/ddd-release-20261018.md`。标准启动入口接通既有 Interpreter deadline 与固定公开上下文配置，成组校验并保持启动时权限和来源检查；使用真实程序核参数拒绝、实际启动及正常关闭。不得启动原生控制器或模型，不重新打包整个模型资产。 |

所有轨道先读根 HANDOFF / STATUS / QUESTIONS、本计划、产品 AGENTS、SPEC §3/§7/§8/§12/§14/§17 和相关原任务。现有 Go/React/SQLite、目录及消费方接口优先复用；禁止新调度器、跨 app 依赖、传输层访问 SQLite、手写成熟格式 parser、程序化编辑代码和视觉操作。原生文件/命令工具关闭要求保持；没有获准真实来源时不运行原生正向测试。Agent 创建的公开项目及实际自有 SQLite/HTTP 可用于其他检查。

每轨提交 SOURCE、普通 push 并记录实际检查及剩余缺口。新增依赖、DDL、HTTP 契约或跨轨修改先交 Handoff，由主 Agent 核既有授权和唯一写入者；不得借工程接线决定未答权限、archive 证明语义或业务默认。输出使用各树已忽略 `.verification/ddd-20261018/`；可复用 D 现有 Go 缓存，禁止读写 `/tmp`、私人会话和正式实验数据。

验收为精确合并源码的 Go vet/build、架构检查、各变更实际 SQLite/HTTP/发布入口验证及适用前端检查。完整默认 CI、L1/L2 正向、最终包和正式 Release 的限制分别登记。当前已确认 native 工具边界、S3 预算与完整能力证据、具体经验批准等保持未答；本轮不以子集通过宣称整体交付完成。

Orca Run：`run_73578b70c449`。三个工作树均位于 `C:/Users/DW/orca/workspaces/Astrocyte/`，准确初始提交均为 `8c8510426b828359f4c161f93c1edf1083fd5e88`。实际客户端为 Orca 1.4.220 启动 Codex 0.162.0，终端显示有效模型 `GPT-6-Astra xhigh fast`，沿已有配置启动，没有覆盖模型设置。

| 轨道 | Task / Dispatch | Agent 终端 |
|---|---|---|
| D | `task_b44aecc75724` / `ctx_51069eb84f7b` | `term_c1a27e21-fa6a-4c57-87cf-73e03476c5f4` |
| T | `task_7f0e39dc8f3a` / `ctx_5eef71b9239f` | `term_84a6036e-dce4-4c4a-bb03-407934d673c0` |
| P | `task_bc2b4bd63d10` / `ctx_cb685eba3000` | `term_f63eb2bc-16b1-444e-9207-c6c02d9fb4a4` |

主控复核生产准备依赖：`cmd/server/main.go` 的 LocalProject 服务未注入 `WithNativeMessageObjects` / `WithOwnedNativeSessions`；直接注入会同时启用 `SendNativeMessage`，该发送路径尚未核 S3 租约。D 本轮只修复 prepared 证据读取及首次 Claim 的身份校验，生产 Send 保持现状，不增加 main.go 写域，不据此宣称 Prepare → Claim → Send 已跑通。后续须补发送前的当前租约与绑定校验，并保留 S3 两项未答规则；archive 没有可信 PrefixComplete / AppendOnly 正事实，继续保留拒绝。

## 开发结果与集成范围

开发分支 `ddd-domain-20261018` 的报告 HEAD 为 `f7007d2d132d3b0585ebe9777b2a377174e78d7c`，已普通 push，工作树干净。D SOURCE 为 `0b585160541cc8f8e29eaffea699c29690b50d03`，T SOURCE 为 `0ce1161de0de2a344a6f33a8442b9874ca5768bd`，P SOURCE 为 `b8d89fa2f9c4cdf725a740873db424bbe249cb51`。主控已逐文件审查代码、三个报告及作者真实验证原件。

作者结果：全 Go vet/build EXIT 0；架构 42 包、7 个 Domain/app、1 个共享 judgment；两项 D 实际检查 2 PASS / 0 SKIP；PowerShell 5.1 语法检查通过；Windows 18 项拒绝、PowerShell/CMD/default 三种实际启动及 HTTP/SQLite 检查通过。全部服务已关闭。CMD/default 服务退出码为 0；PowerShell 服务正常关闭日志、进程退出与端口释放已核，但未取得服务退出码。正向启动后的宿主环境恢复没有取得返回标记，自动项目发现仍有警告；这些项目没有登记通过。准确命令、来源与原件见产品三个 DDD 报告。

唯一集成 Agent 使用现有 `works-relay-integration` 工作树及原产品分支，仅普通合并上述准确报告 HEAD，并编写 `docs/standards/ddd-integration-20261018.md`。不主动改变业务代码；冲突或领域问题交回原作者。其提交及普通 push 完成、终端释放后，主控运行根 `.verification/ddd-20261018/verify-integration.ps1`，依次执行全 Go vet/build、扩展架构与八项现有真实 SQLite/HTTP 组合检查。集成期间不启动第二个 Worker 或并行编译，不重复发布入口已通过的来源验证。

D 已正式 `worker_done`，Task `task_b44aecc75724` / Dispatch `ctx_5242bb05a538` 为 succeeded；官方 release 返回 `retained / user_takeover`，保留该用户所有终端，不关闭或再派开发。集成使用 Task `task_ce7309910dab`，首个 Dispatch `ctx_4557b5a82f12` 在 readiness 阶段失败、未运行任务，已官方 release。相同 Task 在相同产品树接续为 `ctx_f87e16719af4` / `term_371e4c8c-edb2-4736-af26-294592fa6829`；实际客户端 Codex 0.162.0，终端显示 `GPT-6.1-Sol xhigh fast`，没有覆盖模型配置。

产品普通合并提交为 `1674075a803225e82f69436a6dc363d7b8d570ab`，集成报告提交为 `11326d98ca0ebf6f5b1ed8f217d1ca3949fa88bb`，已普通 push，工作树干净。合并无冲突，来源文件内容保持。集成 Agent 已正式交付并官方释放终端，transcript 已捕获。主控首次启动最终复验时采样 CPU 59% / 70%、可用内存约 5.57 GiB，守卫在调用 Go 前拒绝；该次没有执行功能检查，等待满足资源条件后续行。

主控之后在 CPU 65% / 62%、可用内存约 5.1 GiB 时执行准确 `11326d98` 的全仓 `go vet -mod=readonly -p=1 ./...`，实际 EXIT 0，根 `.verification/ddd-20261018/go-vet.log` 无错误输出。进入构建前 CPU 为 70%；后继启动采样为 50% / 72%、90%、73%、94%，守卫均在构建前拒绝，可用内存约 4.75–5.72 GiB。主控 build、架构及八项真实组合没有执行，不能计为通过；作者准确源码的已通过结果仍按三个报告保留。

当前没有本轮编译或验证服务进程。主控只读 CPU 采样显示其他桌面应用正在占用资源，保持这些进程不变。资源恢复后使用同一入口依次执行 `-Phase build`、`-Phase architecture`、`-Phase tests`，不重复 vet；各 phase 继续独立进行资源检查。当前报告只提交已实施及实际已验范围，最终组合验收、原生正向、完整 CI/L1/L2、新包和正式 Release 继续未完成。

产品验收报告后继为 `f17a491cdd6392829094c95a3e5efb6011362bc0`，已普通 push、工作树干净，只更新 `docs/standards/ddd-integration-20261018.md`。该提交后再续行时实际 CPU 99%、可用内存约 7.07 GiB，守卫在构建前拒绝，本轮不再尝试重操作。当前实现和作者验证已交付，主控最终组合验证尚未完成；后续仅继续上述未执行 phase，并如实更新报告。
