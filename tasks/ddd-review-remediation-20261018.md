# 2026-10-18 DDD 开发与审查整改

用户要求按照开发规范继续代码开发，采用既有 DDD 分层。基线：管理分支 `validation/v1-20261012` / `734c9b0d0ed357c159bd2127bb63b79e4c6667a9`；产品集成分支 `songconmaisaix31-design/works-relay-integration` / `8c8510426b828359f4c161f93c1edf1083fd5e88`，均干净。本轮 review 原件在根 `.verification/review-20261018/review.md`。当前 64505 无监听，旧服务运行记录按历史保留。正式实验与未答决定依根 QUESTIONS。

17:06 用户要求继续开发并避免占满电脑资源。当前 32 个逻辑处理器，CPU 约 11%，总内存 31.3 GiB、可用 16.2 GiB。Orca 重启后原三个 Dispatch 已记为 failed / terminal_missing；D 留下五件未提交修改，T、P 尚无修改。保留原工作树和 Task，逐个恢复，同一时间只运行一个开发 Worker；Go 设置 `GOMAXPROCS=2`、`-p=1`，编译、测试、架构检查串行，使用 BelowNormal 优先级，临时 HTTP 服务最多一个。可用内存低于 8 GiB 或 CPU 持续超过 60% 时停止启动新的重操作并重新检查；不重做完整包和模型推理，不并行安装依赖。

17:17 为在资源守卫生效期间继续实施，将尚无修改的 T/P 写域合并到当前 D Worker、`ddd-domain-20261018` 工作树和同名分支；D 成为本轮唯一开发写入者，同一 Worker 继续代码、测试、报告和返修。新增互斥写域仅为原 T/P 表中已授权路径：`.github/workflows/**`、`scripts/check*.mjs`、`release/**`、`docs/standards/ddd-validation-20261018.md`、`docs/release/windows.md`、`docs/release/ddd-release-20261018.md`。原 T/P Task 记录为已合并处理，不启动其他开发 Agent，原空工作树保留。资源不足时可继续静态代码和文档修改，全部编译、真实服务与验证等待资源恢复；完成后仍由唯一集成 Agent 合并并由主控复验。当前 D Dispatch 为 `ctx_5242bb05a538`，终端 `term_e0771644-b770-46ef-b28a-89cfc660acd1`。

本轮使用三个互斥开发轨道。历史开发树和未接受提交保留；为避免历史未结算责任与分支差异影响本次源码，在准确产品基线上创建三个独立工作树。每轨固定同一 Agent 完成实现、真实验证、报告与返修。主 Agent 只维护本计划、状态、决定及验收；开发完成后由唯一集成 Agent 在原产品集成树普通合并，主 Agent 最终检查、提交与普通 push。

| 轨道 | 固定工作树 / 分支 | write_paths 与结果 |
|---|---|---|
| D 领域与接入 | ddd-domain-20261018 / ddd-domain-20261018 | `internal/swarm/**`、`internal/workspace/{app,domain}/relay*.go`、`internal/adapters/sqlite/swarm*.go`、`internal/adapters/agents/relay*.go`、`internal/adapters/events/relay*.go`、`cmd/server/relay*.go`、`docs/standards/ddd-domain-20261018.md`。沿既有 prepared input 与成员端口修正首次 Claim；检查 archive 可信完整性来源和传递，只有来源可证时实施相应接线。消费方端口归 app，外部事实与持久化归 adapters，启动只做组装。未知事实和未答预算不补默认值。 |
| T 检查入口与真实测试 | ddd-validation-20261018 / ddd-validation-20261018 | `.github/workflows/**`、`scripts/check*.mjs`、`docs/standards/ddd-validation-20261018.md`；需要修改其他测试先交准确 Handoff。使 DDD 检查覆盖当前业务上下文，核默认检查所需真实前置；修正可独立验证的工程问题，保持全部原断言、来源范围与验收要求。Native 和模型来源不足时如实失败，不加入 skip、mock 或伪造素材。 |
| P 发布配置 | ddd-release-20261018 / ddd-release-20261018 | `release/**`、`docs/release/windows.md`、`docs/release/ddd-release-20261018.md`。标准启动入口接通既有 Interpreter deadline 与固定公开上下文配置，成组校验并保持启动时权限和来源检查；使用真实程序核参数拒绝、实际启动及正常关闭。不得启动原生控制器或模型，不重新打包整个模型资产。 |

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
