# W6 · MCP 认知接口（M5 接口侧）

> 文档用途：需求与历史记录。保留本文件独有业务条款、来源、实际检查和未验限制；原日期的人员分工、资源额度与操作安排只适用于当时阶段。后续执行读取 [AGENTS](../AGENTS.md)、[RESOURCES](../docs/RESOURCES.md) 和 [当前任务目录](../docs/TASKS.md)，独立 Reviewer 按准确来源验收；归档与入口调整不改变功能完成状态。

实际分支：`songconmaisaix31-design/attention-engine-w6-mcp`。写域：`internal/adapters/mcp/**`、`internal/attention/app/mcp_*.go`、`tests/attention-engine/M5-mcp*`、本文件。

## 2026-10-11：提交后 capture 错误回执修复（当前最新）

最新 SOURCE `1f185a494ef05203977b33b3ff5816bbb033d75c` 已 commit + push；初始 capture 修复 SOURCE `2314b6036f3b57788398adb420e8d8e283f639cb`、其报告 `a65b493` 均保留在普通历史中。本次最终 REPORT SHA 为最新 SOURCE 后紧邻的文档提交（准确 SHA 随 Orca worker_done 投递）。普通精确合并 W1 `673053186a12377c3a1dded4c0b987daa63f2dd5`、后续 typed error `d618b0bfe6d70b49fad2d43b1cf1b497ffc6a1c4`，W5 `b43d2a4cddd8a3f4a367676f9baaa6d3f3414ade`，以及 W2 生命周期/seed floor 修正 `aa4a64e194d7710d3cd39badb71bcc6803d7108d`。外轨文件变更仅来自这些依赖的普通 merge；本轨代码修改仍严格限于上述写域。

原问题：W1 已提交 receipt/signal 后捕获经验失败，会返回持久 rows + error；原 MCP 后端遇到 error 丢掉 rows/稳定 ID，transport 又只输出 error 文本，使调用者可能误判原写入不存在。现在按本次生成的 origin_key 匹配自身 receipt，保留 event_id 与全部 receipt_ids；通过 `errors.As` 消费 W1 `EventCapturePendingError`，不匹配错误字符串。

- 已知提交：`delivery_state="committed"`、`accepted=true` 或 `acknowledged=true` 表示原 receipt 已提交；typed error 同时返回 `capture_pending=true`、`isError=true`，绝不将回执提交当成 capture 已完成。
- 未取得持久回执：`delivery_state="delivery_unknown"`，保留 origin_key，省略 accepted/acknowledged/event_id，避免无 ID 的 `accepted=false` 暗示原写入不存在。error/recovery 明确要求核对原 origin/receipt，禁止换新 origin 盲重发；后端不自动重试。
- 错误结果的 JSON 在 `content[0].text` 中包含上述全部字段及 error/recovery，并复制到 structuredContent。客户端即使在 isError 时丢弃 structuredContent，也能从错误正文取得稳定 ID/未知状态。
- 主控追加的 dynamics 待处理要求已修复：本轨导出 `app.MCPDynamicsPendingError{Cause error}`（Error/Unwrap），供 W0 recorder 在原 receipt 提交后，signal 读取/ApplySignals/缺少配置时返回 **原 rows + typed error**。后端独立标记 `dynamics_pending=true`，不误称 capture_pending；稳定 ID/正文/原 origin 的保留规则相同。W6 不隐式调用动力学或自动 birth。
- 外部 MCP observation outcome 继续为 unknown，raw 自称 failed/success 不成为可信结果；task veto 继续 pending receipt，任务→主题关联未知，无假主题/停止任务行为。

新增 `M5-mcp_delivery_test.go` 的 4 子例（report_event/veto × capture/unknown）启动 **8 个独立 stdio 子进程**，实际 SQLite 迁移/W1 Record/W5 CaptureCommittedFailure/CaptureFailures，无 memoryBackend。capture 子例通过测试专用 recorder 追加一条明确的独立 `contract_local` 可信合成失败，真实 W1 批次提交后实际 W5 callback 在经验事务边界注入不可用；不从公网 Agent raw 提升可信 outcome。独立父进程仅解析错误 content 文本即可看到原 MCP receipt ID、两条 typed receipt IDs、capture_pending，随后重开数据库确认原写入；冷 stdio 进程通过健康 W5 catchup 恢复恰好 1 条 quarantine 候选，其 Evidence/OriginEventID 指向独立可信失败，MCP approved-only 查询仍为空。再次 catchup 不复制候选，原 receipt/raw/Observation 完全不变。unknown 子例在实际提交后丢失 ack，正文仅承认 delivery_unknown；冷进程核对原 origin 对应唯一持久 receipt，没有新 origin/重写/虚构失败候选。

追加 `TestMCPRealStdioDynamicsPendingKeepsReceipt/unborn_topic` 启动 **2 个独立 stdio 进程**：实际 W1/SQLite 提交 commit receipt/signal，实际 W2 对未 birth 主题 ApplySignals 返回错误；测试专用 recorder 保留 rows 并包装正式 typed seam，客户端正文保留 committed ID 与 dynamics_pending，capture_pending=false。冷进程读注意力仍无主题，数据库中原单条 receipt/signal 不变，无自动创建/重发。此用例不声明未确定参数的全部分支或生产 cmd 已验证。

原 `M5-mcp_sqlite_test.go` **5 个独立进程、全部五操作**验收继续 PASS（32→34 完整解释、2 条项目范围 approved 经验、原 raw BLOB、边界拒绝与冷启动），helper 现使用实际 W5 postcommit callback。新旧三项真实 stdio 验收总计 15 次独立进程；所有输入均 `contract_local` 合成数据，不声明 native/task_live。

验证命令（`GOFLAGS=-p=1`、`GOMAXPROCS=2`）及结果：

```powershell
go test ./tests/attention-engine -run 'MCPRealStdioPostcommit' -count=1 -v
# PASS：4 子例，8 个实际 stdio 进程。
go test ./tests/attention-engine -run 'MCP|Experience|M1CommittedFailure' -count=1 -v
# PASS：16 顶层 PASS / 1 helper SKIP；capture/unknown/dynamics 与原五操作真实 stdio 均通过。
go test ./internal/adapters/mcp ./internal/attention/app -count=1
# 两包 PASS（adapter 全量运行，未套过滤条件）。
go vet ./internal/attention/app ./internal/adapters/mcp ./tests/attention-engine
go build ./...
node scripts/check-architecture.mjs
git diff --check
# 全部 exit 0；23 包架构边界通过。
```

本阶段首 RED 保留：定向回归的旧五进程测试在治理 approve fixture 报 `validation_failed: Command metadata is incomplete`（新增 IdempotencyKey 契约）；本轨只为 approve/share 测试命令补齐独立稳定 key，随后相同定向套件 PASS，未降低生产校验。首次合并执行的 adapter 命令误用了只匹配 MCP/Experience 名字的过滤器，得到 `[no tests to run]`，不算 adapter 验收；随后如上单独全量 adapter/app PASS。历史首失败与原成功记录保留在下文。

W0 Handoff 已投递精确 SOURCE/constructor：`NewEngineMCPBackend` 签名不变，event recorder 使用 `NewEventService(events, explicitStrength, experiences.CaptureCommittedFailure)`（或保留 rows+error 的可信项目入口），startup principal/scope/session 与人类 veto 门保持原要求。本轨不写 cmd/HTTP/迁移/契约。**生产命令实际启动、本机 MCP Agent/官方 client 互操作、真实可信 Agent 失败以及 task_live 仍未执行**；W0 消费最终后端与生产接线验收独立于本阶段合成 stdio 检查。真实业务参数/批准提权不由本轨冻结或暴露，M2 全部算法验收/M6 与整体项目完成仍由主控分别判断。

## 2026-10-11 22:28 +08：W6 接口与真实 stdio/SQLite 局部验收完成

SOURCE `e057f100cab38e80b7a081c6f5a3639d023e2f63` 已 push；真实薄后端 `0ac5d3a`，task pending 修正 `ae252c6`。精确普通消费 M0 `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`、W2 `e1222e75048c6cb83222c2307be874e6ae38126a`、W5 `d8e73334cb75c46a8c38460f4764cfe950ce603e`、W1 `e9bb3809eadcbf48cf017b6e09ae0ef4112b1a90`，未修改其他 owner 文件。

交付：stdio 五操作、完整解释共享 DTO、approved/项目范围经验读侧、未验证 Agent 上报、人类 topic/task 否决 receipt。全部通过 `app.NewEngineMCPBackend` 调用原 W1/W2/W5 服务，不 import persistence 到 MCP transport，不实现调度。

`tests/attention-engine/M5-mcp_sqlite_test.go` 启动 **5 次独立 stdio 子进程**，子进程打开真实 SQLite 并组装原 W1/W2/W5 应用，无 mockBackend；输入为 `contract_local` 合成数据，不是本机真人 Agent 会话或 `task_live`：

- 真迁移 + W1 Record 生成 receipt/Signal，W2 实际 ApplySignals 形成注意力；解释保留 18 篇论文、12 次编辑、1 次标记、1 次失败的 32 条原事件/时间/strength/weight/elapsed decay/confidence-only、2 项目/2 Agent、1 条带时间戳关联。关联在测试中通过实际 repository 明确种入，不声明 lifecycle 算法验收。
- W5 实际失败捕获与人类 approve/share：stdio 返回本项目 approved + 人工共享 approved 共 2 条，排除 quarantine 与未共享外项目；工具传外 scope 拒绝。
- 两次独立进程读结果完全一致，node/experience/event 快照未改变；未知 focus 读不新增主题。
- report_event 原 nested JSON 字节与 SQLite BLOB 完全一致；source/session/scope 为启动绑定，未知 tokens=nil，raw 声称 success/human 不改变 observation outcome=unknown、不产生成功信号；未归类 topic 保存 pending 且不生成 signal。
- 人类 topic veto 保存 manual/user_veto 与 -0.8 信号；**task veto 保存 pending receipt，返回 `pending=true`，无可信任务→主题关联时不生成主题/信号，不执行或停止任务**。普通 Agent veto / human-kind 上报 / 外项目写入拒绝，持久状态未改变。
- 明确在 MCP 外经 W2 消费 commit/veto 两条持久信号，再冷进程解释返回 34 条原事件与否决权重；MCP 本身不隐式调用动力学更新，也不建后台任务。

实际验证命令（`GOFLAGS=-p=1`、`GOMAXPROCS=2`）：

```powershell
go test ./internal/adapters/mcp -count=1 -v
go test ./tests/attention-engine -run 'MCP|Experience' -count=1 -v
go vet ./internal/attention/app ./internal/adapters/mcp ./tests/attention-engine
go build ./internal/attention/app ./internal/adapters/mcp
node scripts/check-architecture.mjs
git diff --check
```

全部 exit 0。MCP adapter 12 顶层 PASS（版本协商 5 子例）；MCP/Experience 选定套件 9 顶层 PASS / 1 SKIP，SKIP 仅辅助子进程入口，实际子进程由真实验收测试执行；22 包架构通过。W6 自有验收 4 顶层 PASS，其中真实 SQLite/stdio 1 条、测试后端 3 条，分别记录，不混为 4 条真实后端。

首失败保留：共享测试函数名冲突见历史段；真实 stdio harness 首运行 `no such column: history`（history 存在 data JSON 内），修正查询后第二次 `sql: database is closed`（重开后旧 repository），本轨修正连接绑定，随后相同真实验收与以上选定套件通过。未修改迁移、放宽断言或重复真实媒体/模型请求。

剩余接线：W0/集成 owner 在 `cmd/` 组装相同 startup binding、显式业务 policy/strength、原 W1/W2/W5；本轨已通过 Orca 投递精确 constructor Handoff。**生产命令启动、官方 MCP SDK/client 或本机 Agent 互操作未执行**；没有调用模型、私人源扫描、真实会话或个人数据库。合成验收用 strength=.01、decay=.1、initial confidence=.5、confidence-drop threshold=.05 只是显式 fixture 值，不冻结待用户决定的产品参数。M5 Experience 自动捕获接线、完整 M2/M6 与整体项目验收仍属主控/原 owner。

## 历史阶段记录（以下不覆盖最新结果）

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
