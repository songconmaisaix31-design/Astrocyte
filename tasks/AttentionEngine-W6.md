# W6 · MCP 认知接口（M5 接口侧）

分支：`attention-engine/w6-mcp`。写域：`internal/adapters/mcp/**`、`internal/attention/app/mcp_*.go`、`tests/attention-engine/M5-mcp*`、本文件。

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
