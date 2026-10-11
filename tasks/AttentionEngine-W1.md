# AttentionEngine-W1 · 事件流与归一化（M1）

事实源：`C:/Users/DW/Downloads/tabbit-code (2).md` §4.1/§4.5/§5.2/§8 M1。
分支：`songconmaisaix31-design/attention-engine-w1-events`，基线 `dab8fea`。
当前 Task `task_13ac51983837` / Dispatch `ctx_62837e7857f0`；客户端 Codex，沿启动配置执行，没有启动其他 Agent。
协调者确认运行配置为 Codex / gpt-6.1-sol / YOLO。

## 本 Dispatch 最终交接（2026-10-11 22:46，Asia/Shanghai）

root 在外部权限边界明确要求收束为 `worker_done outcome=failed`，不是成功验收或实现回滚。
业务 SOURCE `f3460ef785cdea0ecaf443628b82f573d05b9155`；报告之前 HEAD `b51c036fb1033cbeb637a848031459cf33d60869`；本报告提交也会 push。
已完成：Event→Signal 固定权重、未知观测 pending、原 raw BLOB 与稳定回执 ID、SQLite 原子/冷重开/重复回放、既有 S1 应用真实事件路径、
提交后 W5 capture + typed committed IDs、W3 同事务读侧、W6 入账/错误 Handoff、pending-only WorksEventSink。
受影响五包测试及构建 PASS；M1 本地 9 PASS / 1 SKIP，§5.2 11 条权重样本全对。
所用命令 `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/adapters/events ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite ./tests/attention-engine`，
构建 `go build ./internal/adapters/events ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite`；未执行 Node/浏览器/外部模型请求。
真实选定缓存第一次 RED：两客户端 eligible metadata 均 0；root 另行确认该 exact 项目 contributors 为空、cacheStatus=partial；没有扩大读取或重跑已确认空场。
真实两客户端工具轨迹回放 NOT_DONE，精确原始元数据范围/样本待用户答复；private message/prompt/file body/credential 读取未执行。
M1 总体未验收，产品强度仍未决定；已有 map constructor 只接受显式 scale，pending-only ctor 不指定 scale。
后续由 root 在用户回答后沿原 owner/worktree 新 Dispatch 接续真实样本验收；不从局部绿推断整体 Attention Engine/M1 完成。
W0 最终应用组装、root 综合验收/合并仍归协调者与原 owner，本轨没有跨域修改。

## 当前状态（2026-10-11）

M1 NOT_DONE；独立解析阶段已 push `bd33fdaf2810ce07a378b2b4d6e133af89d79c16`。
该阶段落 `internal/adapters/events/native_record.go`、`native_reader.go`、`s1_record.go`，
仅适配器内部观测解析，不定义跨轨共享 Event/Signal，不消费未发布 M0。
新增单项目 S1 缓存元数据读侧 `s1_metadata.go`，不启动 discovery、不读消息正文。
新增 canonical outbox 读侧 `s1_reader.go`：每页最多 128 条，稳定 ID 游标，完整 envelope 字段和原 payload 字节；不修改 s1 回执。
有界 snapshot 分页以 ID 排序，不把游标当增量时间水位；下一次完整回放从空游标开始并靠 event_id 去重。

M0 已由 root 放行，并普通合入精确 source `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`；此前未提交代码只读，没有复制。
旧波 A 映射文档有错，本文件已重写；旧 SHA `1321b23`、`05065f3` 保留于普通 Git 历史，不能作为实现规格。

## M1 业务实现阶段（2026-10-11）

- `domain.NormalizeSignal(eventID,event,outcome,depth,strength)` 按 §5.2 硬编码归一化；未知 outcome/depth 不产生对应信号；failure/failed 只标 feedback_failure/confidence-only。
- `app.NewEventService(repo,strength).Record(ctx,[]StoredEvent)` 统一事务入账；显式强度，不按 tokens 缩放，不冻结产品默认。
- 可选第三参 `func(context.Context,eventID string) error` 在提交后捕获明确 agent_task failure/failed；原始回执先提交，capture 错误返回已提交 IDs + error，不回滚或隐去证据。
- 已按 root 授权普通合入 W5 精确 `d8e73334cb75c46a8c38460f4764cfe950ce603e` 作为编译依赖；没有修改 W5 写域。
- `sqlite.NewEventRepository(db)` 存储 Event/Observation/pending 与 Signal，OriginKey 唯一，ID 首次生成后持久；原字节 raw_bytes BLOB 同行保存。
- source 重放不能覆写 raw/observation；Signal 重放不能改 weight/strength；失败批次回滚 Event 和 Signal。
- `events.MapS1Outbox/MapCachedMetadata/MapNativeRecord/ReplayNativeFile` 消费共享 StoredEvent，不定义第二事件/证据表。
- `EventReader.GetEvent/ListEvents` 和 `ListSignals(ctx,fromMS,untilMS)` 提供读侧；`sqlite.NewEventTx(EngineTx).ListEvents/ListSignals(fromMS,untilMS)` 复用调用者事务，范围为 inclusive。
- Event.Tokens 是 §4 必需 int 投影；真实已知性一律看 Observation.Tokens，nil 明确未知，不能把投影占位 0 当实际 token 计数。
- topic 不从工具文本猜测；pending 元数据没有 Event，也没有 Signal；映射行为可留下空 topic 待后续已观测关联。

业务阶段本地验证：五包 `go test ./internal/adapters/events ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite ./tests/attention-engine` PASS；
M1 项目验收 4 PASS / 1 SKIP（真实元数据 opt-in）；source 独立检查 5 PASS。
§5.2 11 条样本权重全对；SQLite 重启后的 ID/原字节/未知字段/Signal 一致、重放不重复；同事务读侧可消费；pending SQL 触发器不准 Signal。
`go build ./internal/adapters/events ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite` PASS。
均以 `GOFLAGS=-p=1 GOMAXPROCS=2` 执行；没有 Node 操作。
这些通过不覆盖未获授权的真实两客户端工具轨迹。首次 real cache RED 保持，M1 整体仍 NOT_DONE，继续等待真实样本范围。
后续两客户端格式 fixture→实际 EventService/store 的有界/原字节/ID/权重对账 PASS；这是 contract_local，不是真实来源。
后续提交后 W5 实际 repo 捕获、同 ID 重放候选不重复、capture 错误保留原回执 PASS；这是显式失败 fixture，不是原生执行失败。

## 作品源与提交后错误接线阶段

按 root 授权普通合入精确 P2 `5d356370c4157e74d96388afd098d5fea14dc799` 作为编译依赖，没有修改 W7/W0 写域。
`events.NewWorksEventSink(eventService,clock)` 实现既有 `workspaceapp.WorksEventSink`；
原 WorkEvent envelope 与 Payload 原字节入 RawBytes，SourceKind/Source/Scope/At 保留；TS=At.UnixMilli。
源 key 为 work_id + kind + 原 at + 已观测 revision，不创建新计数/任务语义。
Works 原词表尚未映射为核心行为，全部 pending；尤其未知 topic 的 human metadata 和 machine sync/verify 没有 Signal。
已向 W7 请求确认稳定源 identity/time/revision 语义；W0 负责 ctor 组装。

`app.EventCapturePendingError` 携带 ReceiptIDs/Cause/Unwrap：写入已经提交，捕获还没完成。
Record 保留 rows+error，调用者不能丢弃 ID 或返回无原写入的 Accepted:false；已交接 W0/root。
向旧 W6 Dispatch 的消息被拒（已 completed），没有假称送达；root 已收到需 fresh Dispatch 接续的纠正。

`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine -run 'TestM1Works|TestM1CommittedFailure' -v`：
2 PASS / 1 FAIL。提交后捕获 typed IDs 和 Works 正常 Service→新 store→冷重开 PASS。
首 RED `TestM1WorksSavedButSinkFailureIsExplicit`：关闭事件库、保留作品库开放，SetWorkGroup 保存 metadata 后隐去 sink 错误返回成功。
原因在 W7 既有 saveMetadata/RefreshWorks 忽略 emit 错误；已经发 owner Handoff，不跨域修复、不重放未知动作。
首 RED 保留。随后 root/owner 放行并普通合入修复 `b1668480bed5c153e24a946ce1912f50d8ded4e5`，没有改 W7 文件。
同一 targeted 命令后续 3 PASS：capture typed IDs、Works happy、Works sink-failure 传播；绿色复验不覆盖首次 RED。
W7 确认 OriginKey 中原 At/revision 的真实语义，重复已接受请求不重新 emit；来源不推断 topic。
Works sink + typed error 发布在精确 `d618b0bfe6d70b49fad2d43b1cf1b497ffc6a1c4`，已通过 orchestration 交接 W0/W7。
root 已启用同一 W6 终端的新 Dispatch `ctx_7f6b4a0be462`，typed-error 精确源码和签名已送达；未启动新 Agent。

## P2/W7 合并后的受影响范围复验

当前发布 `3afc664b8aad6da25ccd17c97207bc001f7af83b`。
五包复验：events/domain/app/tests-attention-engine PASS；sqlite FAIL。
新首 RED `TestWorksRequestReceiptsPersistAndPreventUnknownReplay`，`works_store_requests_test.go:52`：
production migration017 已创建 works_requests，W7 测试随后再 CREATE 同名表导致 table already exists。
此测试基线在 W7 旧 migrations016 下通过；冲突是实际 P2+W7 依赖组合，已交 owner 修复，不改其写域。
M1 局部 8 PASS / 1 SKIP（真实元数据未重跑）；受影响四包 `go build` PASS。
不把局部绿或构建绿写成当前五包全绿；等待精确 owner 测试修复 SHA 后定向复验。
随后普通合入 owner 精确 `a3eab2c6d0df026e056a5a873a0d63bf58282688`：仅调整 W7 接口验收使用实际 production migration，未改领域行为。
原 receipt fixture 定向 `-count=1` PASS；首 duplicate-DDL RED 保持。后续受影响五包是否全绿另记，不由该定向通过推断。
在该修复之后重新执行受影响五包，全部 PASS；受影响 domain/app/events/sqlite 构建 PASS，发布 `f3460ef785cdea0ecaf443628b82f573d05b9155`。
此后无扩大源读取、无再次跑确认空的真实 cache；real 两客户端回放仍保持首次 RED/授权待答。

## 未决定参数下的 pending-only 入口

root 指定作品源不能为未答参数猜 strength/NaN。
新增 `app.NewPendingEventService(repo)`：strength 为 nil；只接收 pending 回执，mapped 输入显式报错且事务回滚。
已有 `NewEventService(repo,explicitStrength,optionalCapture)` 签名保持兼容；已配置时才归一化。
W0 作品接线可用 `events.NewWorksEventSink(app.NewPendingEventService(sqlite.NewEventRepository(db)),clock)`。
`TestM1PendingRecorderDoesNotChooseSignalScale` 和以真实 pending-only service 运行的 Works 正常/失败、capture 四项 targeted PASS。
这条入口保存事实，不选择产品默认，也不把 pending 降为最终验收。
缓存元数据 source key 包含原 observed_at 区分真实缓存 refresh 与同条 replay；source TS 仍为 header created_at，不把 refresh 当 session activity。

## 当前有效映射边界

- 每条源记录保留原载荷；映射未知的记录留 adapter pending，不能创建 core `unknown` kind 或流入 Dynamics。
- 导入/蒸馏 job 保留原 source kind/outcome；无论 succeeded/failed，都不是 agent_task 闭环或失败经验。
- NativeOperation accepted 只说明命令回执，不证明 Agent 任务成功；unknown 不等于 failure。
- source/session 时间为 Unix 毫秒；correlation_id 是请求关联，不冒充原生会话。
- token 源字段缺失/null 保持未知，不猜 0；普通 Read 不证明长时阅读，视频类型不证明暂停/回放/倍速。
- topic 未观测留空；缺少 project_id 不用 aggregate_id 猜项目 scope。
- later/deferred 不当 veto；用户明确 reject/revise/withdrawn 才可映射否决。
- material_imported/source_items_selected/project_space_referenced 可映射 collect。
- material_updated 的 pinned=true 可映射用户标记；原生明确 Edit/Write/MultiEdit/apply_patch 可映射 edit_file。
- Grep/Glob/WebSearch/WebFetch 可映射 search；单个明确 git commit 命令可映射 commit。
- 单条原生记录有多种工具行为时保留 pending，不猜单一类型。

真实 Claude/Codex 回放只允许本轮获准 Astrocyte 项目或原批准 scope 的既有 metadata-only 产出。
外部原生对话记录选取范围仍待用户答复；没有读取消息正文、提示词、文件正文、凭据。
原生选定文件 reader 已实现最多 128 条/1MiB、头部项目/session 匹配、符号链接/越界拒绝，返回截断状态；
此 reader 尚未用于读取私密原文件，不把本地 fixture 当真实回放。

## §5.2 固定权重与强度

| 信号 | 权重 |
|---|---:|
| 深度行为有直接证据 | 1.0 |
| edit_file/commit/search | 0.6 |
| bookmark/collect | 0.3 |
| 用户标记重要 | 1.5 |
| 用户否决/打回 | -0.8 |
| 明确 Agent 任务成功闭环 | 0.8 |
| 明确 Agent 任务失败 | -0.1，仅 confidence |

权重不在线学习。强度产品初值尚未决定；保持显式参数，测试 fixture 可显式传 1，不冻结默认、不按 tokens 缩放。
缺少 outcome/depth 的 observation 不推断对应信号。共享 Signal 失败路径由 feedback_failure 分类承载，仅供 confidence。

## M0 Handoff（已由 47f47f3 统一，历史需求）

1. 稳定发布 SHA 与 Event/Signal 最终代码。
2. §4 Event 字段保持原规格；另统一 adapter observation wrapper：稳定 receipt/event_id、源 kind、可空 token、unknown outcome/depth、可选映射 Event。
3. event 同一行存 rawbytes，避免 encoding/json.Marshal 压缩 RawMessage 丢原字节；不新增 evidence 表。
4. pending 未映射观测能入库，不能造新核心 kind；signal 去重键支持 replay 幂等。
5. 现有 s1 outbox 为 canonical 事件入口；不重复把 Usage/outbox 同一行为算两次。
6. 明确 replay 元数据来源与 provenance；缺失真实来源保留 NOT_RUN/RED，不造样本。

## 验证与首失败

- `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/adapters/events` PASS，3 项：未知语义、逐条原字节与有界回放、s1 job/deferred 不造 Agent 结果。
- 后续独立 source 阶段同命令 PASS，5 项；新增缓存项目范围/截断和 outbox 原 payload 字节/分页收据对账。
- `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine -run TestM1ExistingS1ActionOutputPath -v` PASS：既有 s1 Service.UpdateMaterial→事务 outbox→新读侧→SQLite 冷重开，ID/envelope/payload 字节一致；新 Engine Event/Signal 段仍待 P1。
- `ASTROCYTE_M1_METADATA_DB=C:/Users/DW/AppData/Roaming/astrocyte/state.sqlite ASTROCYTE_M1_PROJECT_ROOT=C:\Users\DW\orca\Astrocyte GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine -run TestM1CachedNativeMetadataLive -v` 首次 FAIL：claude=0/codex=0，partial=false。
- 该失败只表示精确选定缓存/项目没有可用来源，不表示本机总会话为 0；原库只读，不重新发现或扩大范围。
- root 另行只读确认该项目缓存存在、contributors 为空、cacheStatus=partial；初次输出 partial=false 只描述本次 record-bound 没有截断，不代表全局缓存完整。
- Windows 范围键现在规范 slash/case；Unix 仍按路径大小写精确比较。不重跑已确认空场。
- contract_local 已验证 Event→Signal 与 SQLite 冷重启、幂等回放、既有 s1 应用完整事件路径；task_live 真实两客户端工具回放仍待授权/来源。
- M1 不宣称完成；等待不是 worker_done。

## 后续验收

稳定 M0 合入后：Event→Signal 落码、event_store、s1 outbox 经真实应用路径消费，
原始 rawbytes 冷重启一致、稳定 ID 重放不重复、真实 Claude/Codex 元数据回放 provenance 和未知字段保留，
§5.2 各类固定权重抽检及 failure confidence-only。

## 独占写域

`internal/attention/domain/{event,signal}*.go` 与 `internal/attention/app/{event,signal}*.go`，排除所有 `*_contract*.go`；
`internal/adapters/events/**`；`internal/adapters/sqlite/event_store*.go`；`tests/attention-engine/M1*`；本文件。
契约/迁移/入口交 W0；状态、规划、最终集成由 root 管理。
