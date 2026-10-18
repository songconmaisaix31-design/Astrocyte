# D 轨：首次 Claim 的准备身份读取

## 来源与实施范围

- 日期：2026-10-18。
- 基线：`8c8510426b828359f4c161f93c1edf1083fd5e88`。
- 分支：`ddd-domain-20261018`。
- SOURCE：`0b585160541cc8f8e29eaffea699c29690b50d03`。
- 本报告后的文档提交不改变上述 SOURCE 的代码与测试。
- 任务：根 `tasks/ddd-review-remediation-20261018.md` 的 D 轨。
- 客户端：Orca 1.4.220 / Codex；根任务计划记录模型为 `GPT-6-Astra xhigh fast`，本轮没有覆盖启动配置。
- 本阶段只修改 D 轨允许的 Swarm app、agents/sqlite adapter、`cmd/server/relay*.go` 和本报告；同分支后继 T/P 阶段按根 17:17 合并授权实施并分别报告。

`SwarmNativeBridge` 在原 Registry 身份校验后查询同 operation/session 租约。已有租约继续核当前 binding、member version、holder 和原生身份；首次查无租约时读取 `PreparedMemberInputReader`，核对当前授权回调中的成员和准备身份，再继续原能力及预算规则。

## 消费方端口与生产组装

`internal/swarm/app/prepared_member.go` 定义 `PreparedMemberInputReader` 和私有身份值。`ValidatePreparedMemberInput` 要求当前 S2 授权绑定、`claim_work_item` 用途、相同 authenticated actor、项目、固定根、session、native session、adapter、controller、startup operation、input operation，以及非空 adapter version 和 preparation time。startup operation 不能替代 input operation。私有字段全部禁止 JSON 序列化。

`internal/adapters/agents/relay_prepared_member.go` 的既有 `RelayPreparedMemberEvidenceSource` 实现新端口。它复用 `ReadPreparedNativeSession`，先后核原 Registry 持有的尚未尝试输入，并读取 Workspace 原准备记录、当前权限、待发送回执与固定上下文。映射不携带输入正文，也不产生能力或预算事实。

`internal/adapters/sqlite/swarm_native.go` 仅对真实 `not_found` 租约查询执行准备身份读取；其余数据库错误直接返回。准备中的 adapter version 必须等于原 Registry 版本。`CheckMember` 和 `DescribeExecution` 共用此校验；`ToolsKnown`、`TargetsKnown`、调用与用量未知状态和原预算限制保持。新准备端口不授权发送。

`cmd/server/relay_swarm_assembly.go` 从实际 LocalProjects 服务取得既有 `RelayPreparedInputReader`，与同一个 Registry 组装 adapter，再注入同一个 SQLite bridge。缺少实际准备读取端口立即失败。

生产 `cmd/server/main.go` 的 LocalProject 服务仍未注入 `WithNativeMessageObjects` 和 `WithOwnedNativeSessions`；该准备/发送路径继续返回 `unsupported_capability`。主控确认本轮不增加该文件写入范围。启用这些选项会同时启用尚未核 S3 当前租约的发送路径，因此后续需要原负责方完成发送前租约与绑定校验，并处理根 QUESTIONS 中两项 S3 待答规则。

## 实际验证

两项测试与 vet 在上述 SOURCE 执行，使用 `GOMAXPROCS=2`、`GOFLAGS=-p=1`、BelowNormal。全仓 build 和扩展架构检查在同分支后继 `b8d89fa2f9c4cdf725a740873db424bbe249cb51` 执行，其 Go 代码与 D SOURCE 相同，另含 T/P 启动脚本与架构检查变更；使用主控 17:27 确认的 `GOMAXPROCS=1`、`GOMEMLIMIT=256MiB`、`GOFLAGS=-p=1`、BelowNormal，每项前两次核 CPU 低于 70% 且可用内存至少 4 GiB。全部重操作串行，软内存目标不作为硬峰值保证。

Go 缓存使用 `D:/Astrocyte-agent-assets/20261015/go-build/backend`；临时目录和实际 SQLite 位于本树 `.verification/ddd-20261018/`，已由既有 `.gitignore` 忽略。此前资源条件未满足时运行器在调用 Go 前拒绝，相关终端记录保留。

| 命令 | 实际结果 |
|---|---|
| `go test -mod=readonly -p=1 ./internal/adapters/agents -run '^TestRelayPreparedMemberRejectsMissingOriginalSources$' -v -count=1` | EXIT 0；1 PASS，0 SKIP；包 0.978s |
| `go test -mod=readonly -p=1 ./cmd/server -run '^TestRelaySwarmActualPreparedAssemblySQLiteHTTP$' -v -count=1` | EXIT 0；1 PASS，0 SKIP；测试 0.14s，包 1.845s |
| `go vet -mod=readonly -p=1 ./...` | EXIT 0 |
| `go build -mod=readonly -p=1 ./...` | EXIT 0；SOURCE `b8d89fa2` |
| `node scripts/check-architecture.mjs` | EXIT 0；SOURCE `b8d89fa2`；42 包，7 个 Domain/app 与 1 个共享 judgment |
| `gofmt -l` 本次七件 Go 文件 | EXIT 0，无输出 |
| `git diff --check` | EXIT 0，无输出 |

执行入口为 `.verification/ddd-20261018/run.ps1 -Phase tests|vet|build|architecture`，来源、命令和退出结果保存在 `tests.log`、`vet.log`、`build.log`、`architecture.log`，测试 stdout 保留在本次 Orca 终端记录中。架构输出同时保存在 `architecture.stdout.log`。

两项测试的准确范围：

1. 缺少原准备来源或 context 已取消时，新消费方端口返回错误及空身份。
2. 真实迁移后的 SQLite、实际 LocalProject/S2 服务、原 Registry 和对象库完成 Swarm 生产组装；缺少 LocalProjects 或准备端口时拒绝组装。
3. 单连接 Workspace 授权事务内，实际查询缺失的租约和准备记录，均为 `not_found`；空 Registry 拒绝产生准备身份。查询使用随机 UUID 作为缺失查询键，没有登记原生会话或准备输入。
4. 实际 HTTP listener 和生产路由验证未认证读取 403、人工会话建立、空 Mission 列表 200、human Claim 返回 `scope_denied / bound agent identity required`。
5. 缺少租约时原生对账和停止均返回 `not_found`；缺少当前授权绑定时不产生执行能力。
6. 检查后 `swarm_missions`、`swarm_work_items`、`swarm_leases`、`foundation_idempotency`、`foundation_outbox`、`local_agent_sessions`、`workspace_session_bindings` 均为零行，原生会话目录不存在；HTTP listener 和数据库正常关闭。

## archive 完整性检查

本轮只进行源码检查，没有打开会话原档或启动控制器。

- `internal/workspace/domain/relay_owned.go` 的 archive 端口没有 `PrefixComplete` 来源字段。
- `internal/adapters/agents/relay_owned.go` 的 `OwnedRelayArchive` 不提供 archive `PrefixComplete` 或 `AppendOnly=true`。其已有 `Finalized` 来自原通知采集状态，不能补充缺失的 archive 前缀与追加写入事实。
- `cmd/server/relay_session.go` 的 `relayCaptureBinding` 不填 `PrefixComplete`，覆盖状态保留 `append_only_unverified` / `prefix_unverified` 拒绝。
- `internal/adapters/events/relay_session_reader.go` 仅在快照、退出、归档完成、可信前缀、无未知及无错误均满足时返回 `SessionComplete`。

没有既有可信来源能证明本轮所需 archive 完整性，故没有新增事实字段、持久化或传递。进程退出、读取到 EOF 和 native notification 的完整性均不作为 archive 前缀完整性的证明。

## 未完成与后续操作

首次真实 Claim 正向验收和 Prepare → Claim → 同 operation Send 仍未完成。当前没有合法可运行的原生准备来源：Codex 文件/命令工具关闭条件仍不满足；生产准备/发送依赖保持未启用；S3 调用预算计数与能力清单的业务选择仍待用户回答。新缺租约分支中的实际正向身份匹配、真实租约创建和已有租约的原生重放没有执行。

archive `PrefixComplete` / `AppendOnly` 仍缺可信正事实，archive 完整会话验收未完成。此次两项拒绝测试不替代 L1/L2、两个实际 Agent 执行、成果采用及后续经验复用。

没有运行原生初始化、模型、私有会话读取、正式 V4、旧替身测试、完整默认 CI、完整浏览器检查或发布打包；没有新增 mock、伪造原生证据或 Skip。后续由集成负责人普通合入本分支并执行组合验收；原生正向验证须先满足上述真实前置，具体用户决定继续以根 QUESTIONS 为准。
