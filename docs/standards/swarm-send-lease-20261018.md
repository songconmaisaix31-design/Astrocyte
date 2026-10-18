# Swarm 发送前租约校验

## 实现与方案 review

产品基线 `f17a491cdd6392829094c95a3e5efb6011362bc0`，分支 `swarm-send-lease-20261018`。Run `run_73578b70c449`，Task `task_1795bc720ae3`，当前 Dispatch `ctx_3bdd3825f01c`，终端 `term_6b6909b0-97b7-4e35-a45a-601f887c15d3`。工作树 `C:/Users/DW/orca/workspaces/Astrocyte/swarm-send-lease-20261018`。实际客户端 Codex 0.162.0，终端显示 GPT-6.1-Sol xhigh fast，本轮没有覆盖启动配置。主控于 18:16 review 并批准以下方案；同一开发 Agent 在恢复后的原工作树继续实现、检查和提交，PR 保持 draft，代码 review 与实际原生验收分别记录。

业务 SOURCE `0642efdf7748a64c36e07e466b5f507b9c04a004` 已普通 push，前置实现为 `3abe39be8f09689012ddd818e19141a1d480ea25`。[Draft PR #1](https://github.com/songconmaisaix31-design/Astrocyte/pull/1) 目标分支为 `songconmaisaix31-design/works-relay-integration`；后继报告提交不改变该业务源码。主控已完成 `0642efdf` 的独立静态复核，当前没有追加业务代码修改要求；GitHub review 记录由主控在最终 HEAD 发布，实际原生验收保持未通过。

### 准备的用途与原租约身份

在 Workspace app 的私有 `NativeMessagePreparation` 中新增 `Scope` 与 `Lease`。`Scope.Kind` 明确为 `project` 或 `swarm`，Swarm 同时固定 `BindingID`；`Lease` 固定 `WorkItemID`、`MissionID`、`Fence`。这两项只进入原 session JSON 的 `private_native_messages`，公开 DTO 不变。旧记录缺少明确用途时拒绝，不设置默认。

既有 `PrepareNativeMessage` 明确创建 project 准备；新增 `PrepareSwarmNativeMessage(ctx, caller, projectID, sessionID, bindingID, message)` 创建 Swarm 准备。Swarm 准备核对当前 S2 绑定、Grant、成员和固定包，但尚不要求存在租约。两条入口均使用新增 app 消费端口 `NativeMessageAuthorizer`，缺少依赖立即拒绝。生产 main.go 不连接这些依赖。

S3 Claim 的准备读取改为必需的 Swarm 准备读取端口，project 准备不能认领。增加 app 消费端口 `PreparedMemberLeaseBinder`，由 agents/sqlite 组合 adapter 实现；Claim 在保存租约的同一 SQLite 事务中，把确切租约引用写入原准备。绑定失败则整个认领事务失败。绑定只能首次发生，不能让同一原准备改绑新 fence。

project 准备和发送的最终事务同时拒绝该 session 的 Swarm 准备标记或任意尚存 Swarm 租约。Swarm 标记在租约删除或过期后仍保留，换 caller、换 operation、删除当前租约均不能把原 Swarm 会话改成 project 用途。用途从明确的私有准备产生，缺少租约只导致拒绝。独立 Works Relay 使用既有 RelayRuntime/NativeDriver 路径，继续原权限流程。

最终 project 事务还读取该 session 的全部 S2 绑定，不按绑定状态或批准状态过滤。若固定批准含 `claim_work_item` / `submit_artifact`，或该绑定的 Grant 已建立 Swarm Mission，则要求显式 Swarm 准备及当前租约。普通 Workspace 绑定的已知非 Swarm 动作且无 Mission 时，继续 project 规则；只读 `read_swarm` 本身不构成执行用途。主控于 21:03 批准此规则，它依据现有批准动作与 Mission 归属，不新增用途政策。

S2 的 `ChangeApproval` 只改变状态、epoch、原因和变更时间；`ReviseProposal` 保留原 revision 的 Content，并创建新 revision。用途读取同时核对存储键、绑定版本、原 packet 的 epoch、当前批准和原 proposal revision；三个 Scope 必须相同。原提案按其批准时间核内容完整性，未知动作、缺失批准/提案/包、损坏 JSON 或不一致 Scope 立即失败。历史撤销或提案后继不将原 Swarm 会话改成 project 用途。上述读取均由 SQLite adapter 在同一最终事务内执行，没有外部观察。

### 精确接口与事务边界

Workspace app 新增：

```go
type NativeMessageScope struct { Kind, BindingID string }
type NativeMessageLease struct { WorkItemID, MissionID string; Fence int64 }
type NativeMessageAuthorization struct {
    Caller domain.Caller
    Preparation NativeMessagePreparation
}
type NativeMessageAuthorizer interface {
    WithNativeMessagePreparation(context.Context, NativeMessageAuthorization, func(context.Context) error) error
    WithNativeMessageSend(context.Context, NativeMessageAuthorization, func(context.Context) error) error
}
```

新增 `WithNativeMessageAuthorizer` 与 `ConfigureNativeMessageAuthorizer` 组装方法。首次准备把完成的候选记录交 authorizer，授权回调内执行既有保存；重复准备重新核当前授权和原记录。发送阶段读取原对象、原 Registry、固定准备后，把原记录交 authorizer，授权回调内重新读取当前项目/session/preparation，核对所有原值并 CAS 为 `send_attempted`。对象读取、Git 观察和原生 Send 均在数据库事务外。

Swarm app 新增 `NativeSendCheck`，仅包含成员、原准备、固定租约身份，不含 Workspace 类型；新增 `Service.WithAuthorizedNativeSend(ctx, check, fn)`。SQLite `swarm_native_message_authorization.go` 实现 Workspace 消费端口并做上下文之间的类型转换，将跨上下文拒绝映射到既有 `ServiceError`。S3 应用方法通过已有 GrantAuthorizer 的两阶段成员观察，在最终同库事务中核对最新 Grant/epoch、成员绑定/version、固定 packet、Mission 可执行状态、WorkItem 为 leased 且 fence 相等、原 Lease 的完整 controller/startup/native/session/operation/root、有效期及原预算预留。未知执行能力继续拒绝。最终回调只进行主库读取和 CAS，不执行外部调用。

发送授权的线性化点为最终授权事务提交 `send_attempted`。在该提交之前完成的撤销、停止、换绑定、租约删除或 fence 更新，均被当前事实拒绝。提交之后的撤销属于已认领的在途操作；数据库提交与原生写入之间无法建立跨系统原子事务，本任务不宣称消除这段间隔。发送 context 保留请求取消，并取请求、租约、批准到期、批准执行截止、无进展截止和成员授权到期中的最早绝对时间。原生发送前复核该 context；失败仍保留 `send_attempted`，不自动重发。已受理操作仅允许核原输入后观察，未知或在途操作拒绝重发。此在途边界已由主控依 SPEC §17.3 批准，实际竞态仍待合法原生来源验证。

### 写入范围与验证

修改 `internal/workspace/app/native_messages.go`，新增 `native_message_*.go`；修改 Swarm Claim 并新增发送授权 app/domain 文件；修改 `internal/adapters/sqlite/local_native_messages.go` 与 `swarm*.go`，新增同前缀 adapter；修改 `internal/adapters/agents/relay_prepared_member.go` 与 `cmd/server/relay_swarm_assembly.go` 及允许范围内真实测试；同步产品 AGENTS.md 和本报告。消费方端口属于 app，SQLite 与 Registry 组合属于 adapter，入口只负责组装。`projects.go`、Workspace `contracts.go`、HTTP/OpenAPI、DDL 和 `main.go` 没有变化。

验证使用真实 SQLite 和实际 HTTP，覆盖缺来源、当前事务/跨库拒绝、取消、权限拒绝、重开数据库；纯规则检查只作为参数规则记录。Go 限制为 `GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`、BelowNormal，全部重操作串行，最多一个临时 HTTP listener。每项前两次采样 CPU <70%、可用内存 >=4 GiB，采样之间等待 2–3 秒；不满足时在调用 Go 前退出 75。Go 缓存复用 `D:/Astrocyte-agent-assets/20261015/go-build/backend`，软内存目标不作为硬峰值保证。临时编译、SQLite、对象、资源采样及日志保存在本树已忽略的 `.verification/`。

## 验证命令与实际结果

执行入口 `.verification/run-sw-send.ps1 -Phase rules|sqlite|agents|http|vet|build|architecture`。下表按最终源码的适用检查填写；完整默认 CI 未执行，仍未通过。源码与报告提交使用 `[skip ci]`，保持现有 workflow、Skip 和断言。

| 命令 | 当前结果 |
|---|---|
| `go test -mod=readonly -p=1 ./internal/workspace/app -run '^TestNativeMessageExplicit' -count=1 -timeout=120s -v` | EXIT 0；最终源码 4 项参数规则检查，通过；原件 `rules-20261018T130624.log` |
| `go test -mod=readonly -p=1 ./internal/adapters/agents -run '^TestRelayPreparedMemberRejectsMissingOriginalSources$' -count=1 -timeout=120s -v` | EXIT 0；1 PASS、0 SKIP；原件 `agents-20261018T124356.log`，对应 agents 实现与最终源码相同 |
| `go test -mod=readonly -p=1 ./internal/adapters/sqlite -run '^(TestNativeMessageActualTransactionBoundaries\|TestPreparedNativeMessageMissingOriginalReaderDoesNotPublish\|TestNativeMessageExplicitJSONRules\|TestSwarmNativeMessageActual)' -count=1 -timeout=120s -v` | EXIT 0；最终源码 4 PASS、0 SKIP；原件 `sqlite-20261018T130652.log` |
| `go test -mod=readonly -p=1 ./cmd/server -run '^TestRelaySwarmActualPreparedAssemblySQLiteHTTP$' -count=1 -timeout=120s -v` | EXIT 0；含生产发送缺依赖拒绝的最终 1 PASS、0 SKIP；原件 `http-20261018T130723.log` |
| `go vet -mod=readonly -p=1 ./...` | EXIT 0；最终源码，实际终端 `PHASE=vet EXIT=0`；Go 标准输出为空，未产生 stdout 日志文件，资源原件 `vet-20261018T131035-resources.json` |
| `go build -mod=readonly -p=1 ./...` | EXIT 0；最终源码；实际退出记录 `build-20261018T131246.log`，Go 标准输出为空 |
| `node scripts/check-architecture.mjs` | EXIT 0；最终源码；42 包；原件 `architecture-20261018T131517.log` |
| `gofmt -l` 本次 Go 文件；`git diff --check` | EXIT 0，无输出 |

首次 SQLite 运行 `sqlite-20261018T124718.log` 为 EXIT 1；后续通过结果不覆盖原件。HTTP 新断言首次运行在 Go 调用前返回 `RESOURCE_GUARD_REFUSED`：两次 CPU 63% / 79%，可用内存 4,470,108,160 / 4,497,633,280 bytes，没有启动测试服务。

最终 vet 的前两次运行在 Go 调用前分别因 CPU 72% / 100% 拒绝，资源原件为 `vet-20261018T130810-resources.json`、`vet-20261018T130900-resources.json`；后续实际通过单独记录。前置 `3abe39be` 的 vet/build 标准输出同样为空，没有 stdout 日志文件，实际终端退出记录与对应资源采样保留；这些结果不替代本表最终源码的复验。

真实检查只创建本公开 Git 仓库对应的项目空间、项目注册、成员 Grant 和真实令牌。取消的成员授权事务没有留下 Grant；实际撤销状态重开 SQLite 后仍拒绝进入回调。缺少 session、原准备或租约时不产生执行身份。原生会话、Swarm Mission/WorkItem/Lease 和成员绑定表均保持零行，输入对象库为空，原生目录不存在。UUID 仅作为不存在的查询键，未作为原生来源写入。私有用途、租约和授权类型的 JSON 规则检查不计为原生链路验收。

## 实际验收与剩余限制

主控静态 review 已完成，GitHub review 记录由主控在最终 HEAD 发布；实际原生验收未通过。生产入口只安装 authorizer，仍没有配置准备/发送所需的对象及原 Registry 读取选项；实际 HTTP 发送继续按缺少依赖拒绝。

正向 Prepare → Claim → 同 operation Send、真实停止/撤销/换绑定/fence 与发送的竞态、原租约绑定的正向 CAS、租约消失后 Swarm 标记的真实持久化拒绝、已有真实 S2 绑定的用途分流及未知原生交付对账仍未验。当前没有符合用户文件/命令工具关闭要求的合法 Codex 原生来源；没有进行原生初始化、模型输入、会话造数、mock、API/CLI 替身或新增 Skip。新增绑定 SQL 的真实空查询、单连接复用、事务外拒绝及跨库拒绝均在本轮 SQLite 检查中通过，动作分类仅计参数规则。

S3 调用预算和能力证据范围、archive 可信完整性与具体经验批准继续等待用户决定，没有设置默认值、`Known=true` 或隐式授权。完整默认 CI、L1/L2 和正式发布未执行；本任务没有共享产品部署、打包、正式 Release 或正式 V4 操作。
