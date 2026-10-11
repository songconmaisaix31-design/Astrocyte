# W6 · MCP 认知接口（M5 接口侧）

实际分支：`songconmaisaix31-design/attention-engine-w6-mcp`。写域：`internal/adapters/mcp/**`、`internal/attention/app/mcp_*.go`、`tests/attention-engine/M5-mcp*`、本文件。

## 2026-10-11 Codex 接续：stdio 边界阶段，M5 尚未完成

SOURCE `a57d67c77fd44513dbef3611bbb091dea3378937`，协议阶段 SOURCE `058a67baff3748933415ab14103e0171e838ffc6`，均已 push。实际客户端 Codex / GPT-6，未使用子 Agent；`GOFLAGS=-p=1`、`GOMAXPROCS=2`。

- 接管原 `daa3f6d`，它的 `memoryBackend` 和 `os.Pipe` 只算接口局部证据，不证明领域、持久化或实际 MCP 子进程完成。
- 初始化验证 `protocolVersion`、`capabilities`、`clientInfo`，支持 2024-11-05 / 2025-03-26 / 2025-06-18 / 2025-11-25 协商；初始化前工具调用拒绝。参数错误返回 JSON-RPC -32602，业务拒绝返回工具 `isError`。
- 启动必须绑定既有 `app.Principal`、项目 `Scope`、`Session`；`CanVeto` 只在受信人类启动通道开启。工具参数和 `raw` 不授予权限。
- `get_experiences` 未传 scope 仍使用启动项目，显式外项目 scope 拒绝；`report_event({event:{...}})` 强制 `source=mcp`、scope/session=启动绑定，拒绝外项目和 `user_approve` / `user_veto`。raw 只保留审计，不赋予批准、提权或成功闭环语义。
- `veto({topic|task,reason})` 必须二选一，普通 Agent 调用拒绝；人类回流具体落库仍待 M0/W1/W2 接线。

验证（2026-10-11 22:21 +08，局部范围）：

```powershell
$env:GOFLAGS='-p=1'; $env:GOMAXPROCS='2'
go test ./internal/adapters/mcp ./tests/attention-engine -count=1 -v
go vet ./internal/adapters/mcp ./tests/attention-engine
go build ./internal/adapters/mcp ./internal/attention/app
node scripts/check-architecture.mjs
git diff --check
```

全部 exit 0；MCP 12 顶层测试 + M5 3 顶层测试 PASS，版本协商含 5 子例；21 包架构通过。测试后端证据不能替代 `approved` 真查询、完整解释权重路径、真实子进程 stdio 或 SQLite 冷重启。

W0 Handoff（已通过 Orca 投递）：

```go
binding := app.MCPBinding{
    Principal: app.Principal{ID: trustedID, Kind: trustedKind},
    Scope: trustedProject, Session: trustedSession,
    CanVeto: trustedHumanChannel,
}
// W6 将在 M0/W2/W5 发布后提供真实 backend 构造器。
server, err := mcp.NewAstrocyteServer(backend, binding)
// 检查 err 后：server.Serve(ctx, os.Stdin, os.Stdout)
```

`cmd/`、锁文件、迁移仍仅 W0 可写；stdout 仅协议，错误写 stderr；不要从工具入参/客户端 capabilities 推断 principal、项目或人类授权。需 W0/M0 统一 StoredEvent / metadata 与事件入口，W2 排序/解释接口，W5 approved/scope 检索接口；主控通知稳定 SHA 后普通 fetch+merge，再完成真实 backend 和 stdio+SQLite/重启验收。task veto 需通过真实任务引用确定 topic，unknown task 不凭空创建主题。

协议参考：[MCP stdio](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)、[MCP lifecycle](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle)、[MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)。以下波 A 描述保留为历史，最新 API 和权限语义以上文为准。

## 里程碑与验收（任务书 §7 / §8 M5）

五操作：`astrocyte.get_attention_state`、`astrocyte.explain_focus`、`astrocyte.get_experiences`、`astrocyte.report_event`、`astrocyte.veto`。
验收：真实 stdio 调用五操作；`explain_focus` 返回完整证据链（事件引用 + 权重 + 关联节点，每条带时间戳）；`get_experiences` 项目 scope 过滤正确。

## 波 A 交付（自足骨架，M0 发布前）

- `internal/adapters/mcp`：最小 stdio JSON-RPC 2.0 服务器（`initialize`/`ping`/`tools/list`/`tools/call`，换行分隔），不 import persistence、不 import 领域。
  - `protocol.go` 消息与错误码；`server.go` 分帧与分发；`tool.go` `Tool`/`Handler`；`astrocyte.go` 五工具注册与 JSON Schema。
- `internal/attention/app/mcp_backend.go`：`MCPBackend` 接口 + 五工具 DTO（§4/§7 定型），是 stdio 适配器与 W2/W5 读侧用例的接缝。
- 测试：`internal/adapters/mcp/*_test.go`（协议/注册/分发，`os.Pipe` 真实 stdio 分帧）；`tests/attention-engine/M5-mcp_test.go`（五操作全链路，测试双后端演示证据链与 scope 过滤的 DTO 往返）。

依赖方向：`adapters/mcp → attention/app`（与 `httpapi` 同向）；`attention/app` 不 import 适配器、不 import I/O。架构检查 `check-architecture.mjs` 已把 `mcp` 纳入 transport 禁持久化约束。

## 波 B 待接（M0 发布后，契约合入集成分支再编码）

- `internal/attention/app` 新增 `MCPBackend` 具体实现，调用 W2 读侧（注意力排序/解释）与 W5 读侧（approved 经验检索 + 项目 scope 过滤），把领域类型映射到本节 DTO。
- `cmd/` 由 W0/集成轨新增 `astrocyte-mcp` 入口，`mcp.Serve(ctx, os.Stdin, os.Stdout)` 绑定真实 backend（入口与 `go.mod` 属 W0 写域，本轨不写）。
- `explain_focus` 证据链与 `get_experiences` scope 过滤的正确性由 W2/W5 读侧提供；本轨测试双后端只验证 DTO 往返与传输契约。

## 停止点

- 批准/提权不暴露为 Agent 工具（五工具无 approve/promote）。MCP 只是分界线，不做 Agent 编排/调度。
- 未加 `go.mod` 依赖：最小 stdio JSON-RPC 用标准库实现，官方 SDK 需动依赖锁（W0 独占），不在本轨引入。
