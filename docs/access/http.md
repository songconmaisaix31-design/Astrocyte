# Works Relay HTTP 与 L1 会话采集

H 分支：`songconmaisaix31-design/works-relay-http`。当前非原生依赖 SOURCE 为 `4f45a3ab50e6a2863b940915874374f58e65f1da`，生产 listener 接口为 `1aadeebaba2330bb8b786989f577415da5208fd5`；前阶段原生验证准备 SOURCE 为 `b6c67e91555522b3d7515d405c4c69a93d0e0bc8`。主装配保护源接线为 `b247dfb12b938739f09acc73e6b7582376951a0c`，真实 SQLite 配置测试为 `78903848199b915a21e95e42f52cbed64c6d606c`。R 的人工完成权限 SOURCE `d14b08dc50014baf9ee636543564b040112cf3b5` 与 REPORT `fae304b80cb3288b9edd08f10500c6e5af7e4210` 已普通合并。已关闭旧服务的业务版本为 `e5714bb868a6519ec70520bca1c0a1b6c4678fb8`。客户端为 Codex API，工作会话没有暴露精确部署模型名称；旧实际原生控制器为 Codex `0.162.0`，原 `turn_context.model` 为 `gpt-6.1-sol`。

## 端口 0 与精确配置交接

本轮 Task `task_abf1f8b1e12e` / Dispatch `ctx_d7d8c2828fba` 只完成非原生依赖。`ASTROCYTE_PORT=0` 由 Go 一次调用 `net.Listen("tcp", "127.0.0.1:0")` 原子选择端口，主装配一直持有该 listener，并以实际端口构建 HTTP 服务。`httpapi.Server.Serve(net.Listener)` 直接使用同一个已绑定的 loopback TCP listener，关闭时返回 `http.ErrServerClosed`。无第二次绑定、端口探测后释放、换端口或重启 fallback。

主装配在路由构建完成、listener 已绑定且仍被持有时，向 stdout 编码一行 JSON：`{"type":"http_listening","api_url":"http://127.0.0.1:<actual-port>"}`。仅此安全地址记录进入该流；它表示绑定完成，HTTP 响应 readiness 仍由消费者对该 URL 的真实 health GET 确认。消费者不得把公告先于 Serve 当成 HTTP 已可响应。每次进程启动重新消费其实际公告，正常关闭不补发地址。B 的 `tests/s1/server.mjs` 消费此接口，API 不再通过 Node 先占用、释放再交 Go 重绑；配置的 Web Origin 仍由既有 Origin 白名单处理，API 同源以真实请求地址核对。

`cmd/server/workbench_config.go` 已核对与精确 SOURCE `24b6390bcad60a2a20b671dcb60725e77eb3a7a4` 无 diff，并在 B 复制期间冻结。B ACK SOURCE `4a6076c402e0b5d31982154a31a905bee6127477`：内容逐行一致、实际 SQLite 两项测试和 build / vet 通过，文件写入范围已归还。H 本轮没有修改 B 的 `workbench_config_test.go`、`config.go`、helper 或依赖。listener 三文件向 B 定向交付精确 `1aadeeb` changedblocks 与新 helper；消费者不复制额外 H / Relay 历史。

| 本轮命令或真实检查 | 结果 |
|---|---|
| `go test -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json ./cmd/server -run=^TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown$ -v` | PASS，0.11 秒；真实 Go HTTP、实际 SQLite 迁移和项目读写，无原生会话 |
| `go build -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json -o dist/http-listener/server.exe ./cmd/server` | PASS；只读 overlay 为 P `579a1cede98b47b08a3126f5d3974da938348929` 的五个精确 Reader 文件 |
| `go vet -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json ./cmd/server ./internal/adapters/httpapi` | PASS |
| H 单树 `go build -mod=readonly -p=1 ./cmd/server` | FAIL：旧单参数 Reader 与主装配可变参数接线未组合；没有整体合入 B 分支 |
| `gofmt -l` 四个本轮 Go 文件、`git diff --check` | PASS |
| 原子绑定、实际地址公告、同 socket 持有 | PASS，公告实际端口非 0，重复绑定该自有端口立即失败，服务地址保持 |
| Origin / CSRF | 实际 API bootstrap 与 CSRF 写入 PASS；配置的独立 Web Origin 读写 PASS；未授权 session、端口 0 Origin 和缺 CSRF 均 403，拒绝后 SQLite 项目数仍 0 |
| 实际 SQLite 与关闭 | 真实项目创建及设置 revision 增加可通过仓储核对；独立 query_only SQL 的 native session 数为 0；正常 Shutdown 后 Serve 返回关闭结果，原地址连接被拒绝 |

本机忽略目录 `dist/http-listener` 保存命令日志与每次独立 `owned-*` SQLite、公开项目目录和 `safe-result.json`。`actual-first.log` 保留首 SQL 表名检查失败；`actual-final.log` 为当前通过结果，不覆盖首失败。测试没有替代 API、driver、会话或模型，没有修改 schema / engine / mapper / weights、V4、日程和 17321。完整默认浏览器与 B helper 的实际组合消费由 B / F 继续验证，本轮通过范围不替代旧 L1 失败或新 v1.4 原生验收。

## 当前实现

当前 v1.4 验证准备 SOURCE 为 `261fb73921d2a1b929f40bf4d714de5356b9024d`。实际服务测试与归档 / 通知证据检查都消费本次 listener 和 `service-public.json` 保存的真实 URL，通过 `net/url` / `encoding/json` 校验明确 IPv4 loopback、已绑定端口、无凭据或额外 URL 内容；不再以固定端口连接其他服务。实际验证服务使用既有 `Server.Serve(listener)`；`ASTROCYTE_L1_API_PORT=0` 要求显式 `ASTROCYTE_L1_WEB_PORT`，两端口必须独立。公共记录解析拒绝尾随 JSON。

薄配置 SOURCE `a293225afa70db386013f25b6796d3d8361a0891` 使已有受保护目录、应用库、来源与快照的异常身份 IO 立即返回固定安全错误；首次尚未创建的目标保持原初始化语义。快照通过 `os.ReadDir` 枚举实际出版条目，显式源、默认空配置、samefile 和目录保护保持。三项真实 SQLite / 目录 / snapshot hardlink 检查分别 PASS 0.08 秒，精确 Reader overlay vet 与格式检查通过；没有编辑 B 配置测试或读取正式材料。

上述准备通过真实 HTTP / SQLite 的 `TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown`（0.10 秒），精确 P579 Reader overlay build / vet 与格式检查通过。首正常关闭失败保留 `dist/http-listener/v14-address-first-failure.log`，当前通过日志为 `v14-address-second.log`；本次客户端主动释放自己的 idle 连接后正常关闭，没有修改原生状态或重试业务请求。新的实际会话、模型和 Today 窗口尚未启动。R v1.4 精确接口交付前不消费其 WIP、不冻结待答状态机，不把这次验证准备计为语义或 L1 完成。

所有 Relay HTTP 使用既有 Go 服务、loopback、human session、Origin、CSRF 与 ErrorV1。GET 读取已持有任务、派生事件和原收据，不启动 CLI、不发送输入、不恢复会话、不读取原文归档。人工创建 intent 后必须明确确认权限；运行时 Service 是 task、terminal、worktree 与收据的唯一写入者。未知输入交付结果保持 `delivery_unknown`，不能自动重发。

`agents.json` 使用 `encoding/json` 解析受信任的 `AgentSpec` 数组，在启动时加载。条目包含 `name/launch_cmd/capabilities/drive_mode`，可选 `input_mode` 为 `line` 或 `stdin_eof`。新增条目后正常关闭并重启同一服务，调度按注册数据选择既有原生适配器；注册本身不授予启动或工具权限。

`cmd/server/relay_assembly.go` 复用实际 Works、既有项目权限与 NativeRegistry、独立 relay.sqlite、既有 PendingEventService 和 EventRepository。observer 只观察人工已接受的任务，取消并 join 后关闭自有进程和存储。静默重读不更新活动时间或持久化未变化的游标。

`GET /api/v1/relay/session-events` 按当前 owned binding 与权限读取已经提交的真实事件。事件只有 `event_id/kind/topic/tokens/scope/ts` 六字段，时间倒序，ID 升序，分页最多 100 条。未知 tokens 保持 null。任务级事件使用仓储提交 ID，归档记录使用稳定 origin 查得的提交 ID；公开查询没有 prompt、reasoning 原文、工具参数、stdout、归档路径或 SQL 错误。

## 实际来源与覆盖

`ASTROCYTE_RELAY_SESSION_SOURCE` 未配置或为 `archive` 时采集控制器提供的准确归档路径。H 通过 `github.com/olivere/ndjson` 读取该实际 JSONL，校验 owned task、terminal、session、原生 ID、原 CWD 和文件身份；增量位置只在事件 admission 成功后持久化。未知记录仍计入覆盖，格式错误立即失败，未完成尾行保留待追加范围。

选择 `native_notifications` 时显式启用 R 的原生观察端口，从控制器持有的通知序号派生事件。其覆盖标为 `source=native_notifications`；归档覆盖标为 `source=archive`。通知和归档分别保留来源语义，通知的前缀或关闭证据不证明归档文件永久追加。

`snapshot_complete` 只表示当前选定范围的完整行已覆盖。归档的完整前缀、永久追加性和关闭证据不足时，`session_complete=false`，公开状态保留 partial 与具体 gap。R 的真实原生结果保留 `ArchiveAppendOnly=unknown`：当前接口没有永久 append-only 保证，原生会话允许 rollback/revert/压缩。恢复后的 raw 通知订阅缺口为 `native_resume_raw_events_unavailable`。进程退出、输入接受、任务完成和模型成功分别解释，不互相替代。

## 本机配置

接入状态位于 `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access`，注册文件为其中 `agents.json`，应用状态为 `app-state/state.sqlite`。启用接入模式用 `ASTROCYTE_RELAY=1`。`ASTROCYTE_RELAY_LIVE_SECONDS` 没有默认值；未配置时活动三值保持 unverifiable，不自行决定活性窗口。

`ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS` 是启动时的可信配置，显式值范围为 1–1800 秒，无值时传递 NativeDriver 原默认策略。配置错误在创建 Relay store 前立即返回。已授权独立验证使用 900 秒；该寿命不会改变活动三值或既有任务权限。实际验证程序通过 `ASTROCYTE_L1_API_PORT` 选择独占端口，新建项目的初始 actions 包含 `reconcile`。原项目和旧收据身份保持。

`ASTROCYTE_RELAY_PUBLIC_GIT_ROOT` 只用于明确获准的 Agent 自建公开 Git 验证项目。目录必须位于独立接入根的 `verification` 内，每次 Prepare 都以 `os.SameFile` 核对该唯一目录；其他目录立即拒绝。未配置时公开项目启动权限保持关闭。既有中性 worktree 仍核对干净、已提交、正确 Git 根和原身份，不修改用户作品或开放自由工具。实际 Codex 使用现有控制器关闭文件、命令、MCP、skills、浏览器、视觉和其他工具，只提供固定公开上下文。

实际验收服务使用生产 `assembleRelay`、真实 SQLite、真实 HTTP 和本次控制器自己产生的会话文件。API 为 `http://127.0.0.1:17342`，F 网页为 `http://127.0.0.1:17343`。本次数据根为 `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/h-production-current-20261014`；项目、worktree、数据库和进程均由本次验证持有。所有脚本、中间文件和安全聚合位于 H 工作目录内被忽略的 `dist/l1-live/current-20261014`。原件不进入 Git、HTTP、stdout 或 release 包。官方 CLI 自行鉴权，不读取或复制凭据；归档只读取控制器给出的本次 thread 路径，不枚举历史。

## 验证命令与当前结果

| 命令 | 结果 |
|---|---|
| 当前 H 原样 `go build -mod=readonly -p=1 ./cmd/server` | FAIL：本分支 Reader 仍为单参数，main 接线需组合消费 B 精确可变参数 Reader |
| 精确发布组合 `go test -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json -c ./cmd/server` | PASS，实际服务测试二进制已编译；使用下面登记的精确五文件 overlay |
| 精确发布组合 `go vet -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json ./cmd/server` | PASS |
| `node scripts/check-architecture.mjs` | PASS，29 个 Go 包边界 |
| `pnpm exec redocly lint contracts/openapi.yaml --config contracts/redocly.yaml` | PASS，保留既有 EventV1 未引用警告 |
| `node scripts/validate-examples.mjs` | PASS，291 个契约示例 |
| `git diff --check` | PASS |
| 实际服务无 Codex 注册基线 | PASS，17342 空注册表，无模型派发 |
| F 注册前真实浏览器基线 | F 于 `2026-10-14T12:53:45Z` 报告两尺寸 session 200、空 registry 与浏览器 ready |
| 实际注册前 HTTP 权限检查 | PASS，无 session、跨 Origin、未授权 bearer、缺 CSRF 均 403 |
| 原生派发前 SQLite 独立只读计数 | 实际 events=3、signals=0、nodes=0；项目 / Works 观测 |
| 同 URL 追加登记与正常重启 | 实际执行一次，原始 `registered_at=2026-10-14T13:23:34.0366105Z` |
| 注册起点 → 原生事件首次 DOM | FAIL：F 两尺寸首显示 `13:33:25.152Z/.153Z`，591116/591117 ms，大于 180000 ms |
| 原 dispatch 收据创建 → 原生事件首次 DOM | PASS：创建 `13:33:10.6363218Z`，约 14.516/14.517 秒；独立于注册起点 FAIL |
| 首次实际原件与提交 ID | 首次检查 PASS：15 个完整行、105564 字节、15 个原档 origin 对应提交 ID；全部 18 个 HTTP 事件 GetEvent 一致 |
| 首次实际文件范围 | `snapshot_complete=true`、尾行 / 剩余字节 0、unknown 5、errors 0、`session_complete=false` |
| 当前原件身份与采集范围 | FAIL：实际 nlink=2，`source_identity_unverifiable_or_aliased`、`snapshot_complete=false`；现有已提交 ID 保留 |
| 实际回答与结束轮次 | `gpt-6.1-sol`，assistant 1、task_complete 1、已报总 tokens 8827；没有重复原输入 |
| 实际权限、分页与 GET 只读 | PASS：四项 HTTP 403、五项实际文件身份 / 权限拒绝、全分页、canary 排除、静默 GET 事件数不变、持久游标相符 |
| 实际原件中的工具与策略 | 当前文件工具调用 0、`never/read-only`；旧 owned PID / 启动参数未当时采集，effective tool set 保持 unknown，不能仅用 tool 0 填禁用证据 PASS |
| 连续追加 / 明确 Stop 后范围 | 未执行，原控制器已在默认 180 秒后实际退出，旧输入没有自动重发 |
| 当前 Foundation / Projects | 实际服务 GET 均 200；运行源码包含 R `d14b08dc50014baf9ee636543564b040112cf3b5` |
| 旧任务人工 done | 实际 200，operation `32c7f5a0-0cb0-4de2-91aa-3e7b9bc0ad07`，收据 accepted；目标验收 FAIL |
| done 后实际 read | 首次 409，operation `5f99133e-390f-4987-b8f2-2889716921ff`，收据 `delivery_unknown`；没有重试 |
| 旧任务单次合法 release | 实际 200 / accepted，operation `ccf54824-6194-4161-81c0-0d6985ee5171`；独立 SQL 核 `released`、`ActiveTaskID` 空、Prepared 保持 true、pending action / receipt 0 |
| 旧实际服务正常关闭 | PASS，1909.08 秒 / exit 0；原真实进程退出、read unknown 和收据身份保持 |
| 主装配显式保护源 | SOURCE `b247dfb12b938739f09acc73e6b7582376951a0c`；解析在应用目录和迁移前执行，参数传给 Reader |
| 精确发布组合构建 | `go build -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json -o dist/l1-live/reader-overlay/server.exe ./cmd/server` PASS；overlay 为 P `579a1cede98b47b08a3126f5d3974da938348929` 的五个精确 Reader 文件 |
| 保护源与寿命配置目标测试 | 四项 PASS：真实 SQLite 显式来源、受保护目录、实际 snapshot hardlink、启动寿命边界；原编译首 FAIL 保留本机 `first-test-compile-failure.log` |
| 新独立服务 17344 | 实际启动和正常关闭 PASS，60.32 秒 / exit 0；真实 GET agents=0、tasks=0，无注册文件、无原生会话文件 |
| 新 scope 原生 / Today | 用户明确今晚停止；注册、模型派发、浏览器与三分钟验收均 NOT_RUN，明日按新窗口继续 |

实际证据检查逐原档 origin 查询 `attention_engine_events` 并通过 `EventRepository.GetEvent` 定位提交 ID，独立比较 HTTP 六字段、分页和 canary 排除；核对重读及持久游标、追加与停止后的真实文件范围。实际 HTTP 权限检查包含无 session、跨 Origin、未授权 bearer 和缺 CSRF；静默 GET 前后事件数量必须一致。模型名称与工具调用数量仅从选定实际文件生成安全聚合。

## 剩余限制

本报告为实际已执行范围的阶段报告，H 保持同一活动 Task 继续真实闭环。F 是主任务唯一产品 UI 操作人；H 负责注册文件与正常重启，并只读核对原档、SQLite 和 HTTP。原 task 为 `0c6522bc-b232-4040-8407-745885644791`，仅确认一次。HTTP 503 与 dispatch `accepted + provider_unavailable` 原收据保留；实际安全原因是 `thread/read / -32603 / thread_rejected`，更具体的服务内部原因未知。之后原 observer 实际取得正选归档，原件回答成功不覆盖首 503，不授权重发。

F 首 DOM 三个 ID 均在真实 SQLite 定位，其中 `MBV3V2HETK5753LFVJY4PGVJWL` 和 `HFZSCX6TFXYYMOU5VBUNG65QIV` 分别是原档 `task_started` 与 `native_session_header`，第三个是 task 级 `turn_started`。安全证据为本地忽略目录的 `committed-public.json`、`initial-actual-evidence-derived.json`、`initial-actual-model-derived.json` 和 `first-dom-sql-derived.json`。14 个原档记录的 tokens 为真实 null，未填值。原 controller 文件路径只保存在本次 relay.sqlite 的私有绑定中。

人工 done/release 使用用户确认的 `observe` 权限，实际事件出版重核使用 `reconcile`。当前旧 task 已 released，独立 SQL 检查确认进程真实退出、Prepared 工作区和 terminal 归属一致、pending receipt / action 均为 0，工作区 `ActiveTaskID` 为空，已解除占用。read 的 `delivery_unknown` 收据与任务条件保持；工作区清理不表示目标或来源验收通过。原生回答的目标内容未获合法核实，目标验收 FAIL；当前身份守卫拒绝的文件不绕过读取。F 首 release 脚本对 Prepared 的错误期望保留首 FAIL，本次 SQL 结果核对现有正确释放语义。

R 通过实际 Windows 文件名称接口与 `samefile` 核对两个名称，分别位于默认 Codex sessions 与 Orca 账户 sessions，均不在 H / R 验证目录。创建者未知，镜像解释仅为推断，没有删除链接或放宽守卫。完整路径仅在本机私有证据中，安全聚合为 `source-identity-derived.json` 和 R 的 `owned-archive-link-safe-aggregate.json`。

主装配通过显式 `ASTROCYTE_WORKBENCH_PROTECTED_SOURCES` JSON 数组传递保护来源，默认空数组，不隐式选择私人来源。Reader 使用发布候选精确文件的只读 overlay 验证，未将 B 冻结域历史整体合入。组合构建和适用配置检查通过不替代当前 H 分支完整测试；测试 SOURCE `7890384` 的原域消费由 B 确认。后续 R 寿命、EOF 与 owned 执行策略端口不反填旧进程证据。

新独立验证目录为本机 `verification/h-production-next-20261014-2310`，中间文件位于 H 被忽略的 `dist/l1-live/next-20261014-2310`。用户停止指令收到时未注册或派发，随后正常关闭。安全计数与关闭结果保存在 `closure-derived.json`，不会替代明日实际会话验证。内部执行策略采集复用原 Registry 的 `OwnedRelayExecution` 与 SQLite 实际 session，保存安全派生 PID、启动策略、寿命和首 RPC 摘要，不返回 HTTP；新 scope 没有实际模型，执行元数据采集仍 NOT_RUN，effective tool set 不推定为已验证。

完整默认 check/build/E2E、合并前端、Windows ZIP 解压启动和正式发布由对应 owner 与集成验收负责。H 的静态构建和契约检查不能代替实际 Today DOM、完整会话、终端编码或正式部署。

## v1.4 技术差距草案

已全文读取 `docs/research/orca-rea-20261014.md`。以下为交模块的设计输入，四项关键语义、字段名称及旧记录处理由用户确认；本轮没有修改 v1.4 业务或冻结接口。

| 方向 | 当前 H 接缝与差距 | 待确认的必要安全信息 | 真实验收 |
|---|---|---|---|
| 投递阶段 | `relay_session_codex.go` 可识别带原 turn_id 的开始 / 完成记录；`relay_session_native.go` 保留通知序号。六字段事件不关联某次 input 的 operation，采集层不判定 task working | R 提供某次操作的输入接受、原生回合开始证据和未知状态；原 turn_id 仅在内部关联，公开 task / receipt 用既有 operation_id 和安全状态表达。人工完成保持独立 | 实际原 controller 的接受收据与开始事件逐一关联 SQL ID；无开始证据不从普通活动、Git、stdout 推定开始；未知收据保留占用且不重发 |
| 原生会话归属 | `RelaySessionBinding` 绑定 task、terminal、session、worktree、project、native ID、format；私有 cursor 绑定该身份及文件位置或通知序号。当前 source 与 gap 已表达 alias FAIL，缺少可持久的原 controller 代次关系 | 内部持久原会话与执行所有者关系及原因；公开沿现有 source、status、gap 和公开 task / terminal 身份表达，未经必要性确认不扩大事件六字段 | 只读本次准确 controller 路径，逐 ID 回查真实事件库；冷重启保持历史绑定且不重建进程所有权，文件身份冲突立即失败、不删未知链接 |
| 输入就绪 | 注册已有 drive_mode / input_mode；H 仅解析登记并组装 driver，没有 TUI 或 structured 的输入就绪检测端口，CLI 可执行和注册不证明 ready | R 供给输入方式、就绪三值、实际就绪证据原因和观测时间；命令、原生输出及工具参数保持私有。GET 只读缓存 | 实际 installed controller 的初始化 / ready 正证据；stdin 写入、换行或 EOF 不推定回合提交，缺端口分别保留 unknown / unsupported；不制作伪 CLI 或 ready 输出 |
| 恢复身份关系 | archive cursor 以 BindingID / FileIdentity / Offset 去重，native cursor 以 BindingID / Sequence 去重。通知恢复缺口仍为 `native_resume_raw_events_unavailable`；现有 coverage 没有持久的恢复 / 换 Agent 关系序列 | R 明确同 session 恢复、新 session、换 Agent 与不支持的 fork 关系；H 消费内部原 / 新绑定关系和 replay 范围。公开只使用获准的 opaque 身份与安全关系状态 | 同 task 恢复核原 native ID / CWD，换 Agent 在合法释放后新 session；原输入不重发，旧 cursor 不错误套到新 stream，重复 ID 不新增或刷新 live，恢复缺口保持 partial |

归档与通知的完整性分别验收，回合完成不封存整个 Session。明日真实 CLI → 三分钟 Today → L1 最终验收需要新独立 scope、F 提前 ready、实际目标内容在人工 done 前核实和真实禁止工具的启动证据；今晚没有新增该窗口。

实际用户作品、自由工具授权、默认活性窗口仍未决定。原生文件的完整前缀、永久追加性与恢复 raw 通知接口不足时，保持具体 unknown/partial。原始 Native 共享源码、mapper、engine、schema/迁移、七类权重及快照流程没有 H 修改。V4 协议、库和运行资料没有读取或修改。管理员 Procmon、两次实际 00:30、正式认知验收和现有 17321 部署均未由本次 H 验收执行。
