# W6 · MCP 认知接口（M5 接口侧）

实际分支：`songconmaisaix31-design/attention-engine-w6-mcp`。写域：`internal/adapters/mcp/**`、`internal/attention/app/mcp_*.go`、`tests/attention-engine/M5-mcp*`、本文件。

## 2026-10-11 Codex 接续：stdio 边界阶段，M5 尚未完成

薄后端 SOURCE `0ac5d3a700c3090d4bb532d1fc763539de804326` 已 push，消费三个 owner 公布的签名：

```go
app.NewEngineMCPBackend(binding, attention, experiences, events, app.MCPBackendOptions{})
// attention: AttentionState(ctx) / ExplainFocus(ctx, topic) — W2
// experiences: GetExperiences(ctx, Principal, project, pattern) — W5
// events: Record(ctx, []StoredEvent) ([]StoredEvent, error) — W1
```

report 保存收到的 nested event 原始 JSON 为 `RawBytes`（含被覆盖前 source/session 供审计），Observation 强制启动项目/session/source=mcp/provenance=mcp_unverified:actor，outcome/depth=unknown；raw 标签不参与推断，空 topic 生成 `Event=nil` pending。人类 veto 对已有 topic 生成 manual/user_veto receipt；task 因无持久解析显式 unsupported，不虚构 topic。W1 Record 负责 receipt+signals 原子保存，W2 流消费更新不在 MCP 内重写。22:22 +08 两包 test、app/MCP vet+build、架构检查通过；真实 W1/W2/W5 实现仍待 SOURCE，未运行实际 SQLite 后端或进程重启。

M0 放行后已普通 merge 精确 `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`。共享 DTO 阶段 SOURCE `d9cbdd15ae79a69818ab65e39427d4220c3d8e19` 已 push：`ExplainFocusResult=FocusExplanation`、`EvidenceItem=ExplanationSignal`、`RelatedTopic=RelatedNode`、`ExperienceItem=StoredExperience`；权重路径、原始 receipt/raw bytes、confidence-only、关联时间/Agent、经验事件引用与时间戳全部保留。`ReportEventInput.Tokens` 为 `*int`，缺失仍为未知。

首失败保留：消费 M0 后 `go test ./internal/adapters/mcp ./tests/attention-engine -run MCP -count=1` build FAIL，原因是 W6 的测试函数 `roundTrip` 与 M0 `engine_contract_test.go` 同名；只在本轨改为 `mcpRoundTrip`，并让管道读写并行，完整解释返回不等待输出缓冲区清空才读取。随后两包 `go test -count=1`、相应 vet、MCP/app build、21 包架构、diff 全部 exit 0；未知 token 补断言后 `go test ./tests/attention-engine -run MCP -count=1` PASS。它们仍是 transport/test-backend 证据，不是 W2/W5 实际应用与 SQLite 验收。

已向 W1/W2/W5 原 owner 投递读写接口 Handoff，等待实际应用签名/精确 SOURCE；backend DTO mapping 由本轨 `app/mcp_*.go` 完成，W0 只做 cmd startup 及依赖注入，禁止交叉写域。

主控复核修正 SOURCE `8a496786885f4181263863f270485f13c00ee303`：`report_event` 接受空/缺失 topic，保留待归类观测，不要求伪造标签；2026-10-11 22:16 +08 受影响 MCP/M5 两包重测 PASS、diff 检查 exit 0。后续将调用 W1 observation/pending 机制；当前仍未接真实入库。

SOURCE `a57d67c77fd44513dbef3611bbb091dea3378937`，协议阶段 SOURCE `058a67baff3748933415ab14103e0171e838ffc6`，均已 push。客户端 Codex，主控 runtime 记录 `gpt-6.1-sol/YOLO`，未使用子 Agent；`GOFLAGS=-p=1`、`GOMAXPROCS=2`。

- 接管原 `daa3f6d`，它的 `memoryBackend` 和 `os.Pipe` 只算接口局部证据，不证明领域、持久化或实际 MCP 子进程完成。
- 初始化验证 `protocolVersion`、`capabilities`、`clientInfo`，支持 2024-11-05 / 2025-03-26 / 2025-06-18 / 2025-11-25 协商；初始化前工具调用拒绝。参数错误返回 JSON-RPC -32602，业务拒绝返回工具 `isError`。
- 启动必须绑定既有 `app.Principal`、项目 `Scope`、`Session`；`CanVeto` 只在受信人类启动通道开启。工具参数和 `raw` 不授予权限。
- `get_experiences` 未传 scope 仍使用启动项目，显式外项目 scope 拒绝；`report_event({event:{...}})` 强制 `source=mcp`、scope/session=启动绑定，拒绝外项目和 `user_approve` / `user_veto`。raw 只保留审计，不赋予批准、提权或成功闭环语义。
- `veto({topic|task,reason})` 必须二选一，普通 Agent 调用拒绝；人类回流具体落库仍待 M0/W1/W2 接线。

验证（2026-10-11 22:12–22:13 +08，局部范围）：

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

`cmd/`、锁文件、迁移仍仅 W0 可写；stdout 仅协议，错误写 stderr；不要从工具入参/客户端 capabilities 推断 principal、项目或人类授权。M0 StoredEvent / metadata 已消费，W1 真实入口、W2 排序/解释、W5 approved/scope 应用仍待发布；继续普通 fetch+merge 后完成真实 backend 和 stdio+SQLite/重启验收。task veto 需通过真实任务引用确定 topic，unknown task 不凭空创建主题；没有实际解析则显式 unsupported。

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
