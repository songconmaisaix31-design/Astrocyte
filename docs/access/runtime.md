# Works Relay runtime

2026-10-14 当前交付：分支 `songconmaisaix31-design/works-relay-runtime`；SOURCE `fbfaa65c2c01094305424ac40d2d39d271eab51b`，已 push 且 `git ls-remote` 返回同一 SHA。REPORT 单独提交，精确 SHA 通过当前 Orca Dispatch 发送。原 SOURCE `62c7d99f1d6b180ef19f90a2112e801702d9fffb`、REPORT `0b4ec501721361e8104a4e6e00781d73656157a4` 和可选通知端口 `139ce8090b868ee2764a13b0517fec6dee079424` 的普通历史保留。

事实源为认知规划原文 `cognitive-stack-L1-L4.md` §2 / §7、`L1-session-capture-plan.md`、当前根 AGENTS 和 `tasks/Release-v0.1-20261014.md`。本轨只负责 runtime、owned provider、可选 Native 观察端口和本报告；H 负责完整采集、HTTP 和实际服务，B 独占依赖，F 负责实际 DOM，最终整合和 release 验收由主控安排。本次 Task `task_a51fab762c01` / Dispatch `ctx_b4fcb141ca00`，执行客户端为 Codex，`orca terminal show` 核实面板模型 `GPT-6.1-Sol high fast`。

## 当前实现

- 单一 RelayService 和独立 SQLite RelayStore 管理 task、terminal、neutral worktree、native session、receipt 和 owned binding。沿现有 `relay_state / relay_receipts` 两表及 private JSON wire 持久引用和派生通知，没有新增 migration、认知 schema 或证明系统。私有 archive 路径、native ID 和原始消息不进入 HTTP DTO。
- 人类 CreateIntent 不启动；ConfirmIntent 必须人工权限确认和当前 settings revision。每次操作重新核现有 ProjectSettings / ProjectGrant，沿 A/B/C 原语义。来源 workspace 同时只有一个 active task；同 Agent 新任务可复用 terminal ID，使用新 native session，同未完成任务 Resume 保持原 task / binding / native ID 且不重放首输入。
- ready → accepted → working → done → released 有持久 edge；done 需要人类确认，release 需要真实退出。操作先存 receipt；原 accepted / rejected / delivery_unknown 不重写，不自动重发未知输入、不创建第二 writer。派生 sink 失败保持 pending/error，只有显式 reconciliation 可重新核查发布。
- `RelayLifecycle.ObserveOwned` 只观察已接受且仍持有的任务；无变化轮询不创建 receipt、不写状态、不刷新活性。首采集错误立即返回并保存，后续后台不自动 probe。Shutdown 对实际 owned process 做 containment stop，不将退出推断为任务完成。
- terminal 使用可信 argv 和已确认的 line / stdin_eof 策略。stdin_eof 首输入后关闭管道，后续 Send 明确拒绝。默认 LaunchPolicy 和 neutral source 授权保持关闭；实际公开项目只能由 H 的显式单根配置组装，不开放用户作品或自由工具。terminal-only 没有原生 archive metadata 端口，保持具体 gap。
- `RelayOwnedSessionReader.ListOwnedRelaySessions` 是内部只读端口，逐次核 task / terminal / worktree / project / agent、当前权限、目录和 controller session 身份。拒权清除私有引用且不调用 provider；closed task 引用可保留，但仍遵守当前权限。GET 不补绑定、不启动进程、不扫描家目录或其他会话。
- `Native.OwnedRelayArchive` 从同一已持有控制器取得来源。未启用通知时仅调用原 thread 的 `thread/read {threadId, includeTurns:false}`，解码 ID / CWD / nullable Path；实际错误立即返回。单独 HistoryPath 不证明文件、CWD 或格式。Claude / Pi / OpenCode 的必要 metadata 组合尚无已验证端口，明确 unsupported。
- `RelayNativeDriverOptions.CaptureNativeNotifications` 默认 false。启用时，`StartRelay` 在原 app-server 接收器启动前安装观察器，原 `thread/start` 开启官方 `experimentalRawEvents`，接收并核实原 thread 身份，仅保存 kind / tokens / sequence / 实际接收时间和覆盖状态。`ListOwnedRelayNativeSessions` 给 H 提供这些派生通知；不保存 prompt、思考原文、工具参数或 stdout。默认 Native 行为保持。
- 观察器使用既有 JSON 解码对象，没有新增字符串或字节 parser。未知通知语义、身份错误、解码失败、截断和容量上限立即失败并保留首 gap。来源前缀只在实际新建、非 fork、ID/CWD 一致且没有采集错误时 confirmed；已有历史或缺失订阅不制造前缀。
- `StopRelay` 在同一控制器调用 `thread/unsubscribe`，等待相同 thread 的 `thread/closed`，然后停止 owned process 并等待通知 reader 结束。只有这条真实关闭链才能确认当前 writer / 通知来源结束；仅 Wait / Stop 的进程退出保持 unknown。`ArchiveFinalized` 表示这次 writer 已结束并完成官方 flush 路径，不表示文件永久不可修改、磁盘掉电安全或全事件已入档。
- Resume 重新打开原 thread，旧 finalized / append-only claims 在 pending 时即失效。Windows 恢复响应出现不同路径拼写时，只 stat 前后两个控制器正选文件，要求 regular 且 `os.SameFile`，保持原引用；不同文件或无法核实立即拒绝。不会从别名扩大来源目录权限。
- 官方 Resume 没有本实现可用的 raw-response 订阅开关，恢复后的通知保存 `native_resume_raw_events_unavailable` 和 Unknown，不能宣称完整恢复覆盖。当前的真实通知活性更新使用新 sequence 的实际接收时间，缓存重读不刷新；三值裁决保持 live / unverifiable / exited。
- 归档 `ArchiveAppendOnly` 始终 unknown，真实停止不赋 confirmed。文件模式的 PrefixComplete 没有独立已实现证据口，H 的 `relayCaptureBinding` 不得自行补 true。文件当前完整行范围与原生通知覆盖分别报告。

## 官方协议和归档语义

本机实际安装并启动的 Codex 为 `0.162.0`。官方源码固定为 `rust-v0.162.0` 对应 `c1382380de69521303b416720a52f42d51af6248`；完整公开文件和官方 CLI 生成的普通 / experimental JSON schema 保留于本工作树忽略目录 `dist/relay-l1`，不进入产品或 Git。当前读取并核查的实现：

- [Thread start / resume 协议](https://github.com/openai/codex/blob/c1382380de69521303b416720a52f42d51af6248/codex-rs/app-server-protocol/src/protocol/v2/thread.rs)：新建线程可开启 raw events；metadata path nullable / unstable；恢复按原 thread ID，metadata-only 使用 excludeTurns。
- [订阅结束与 ThreadClosed](https://github.com/openai/codex/blob/c1382380de69521303b416720a52f42d51af6248/codex-rs/app-server/src/request_processors/thread_lifecycle.rs)：unsubscribe 只移除连接订阅；无订阅且 idle 时等待 `shutdown_and_wait` 成功、移除匹配 runtime 后才发送 thread/closed。
- [Recorder](https://github.com/openai/codex/blob/c1382380de69521303b416720a52f42d51af6248/codex-rs/rollout/src/recorder.rs)：新建先给出路径，文件可延迟 materialize；恢复按 append 打开；shutdown 排空 pending items、flush 并等待 writer task 退出。普通文件 flush 不等于 sync_all 或永久封存。
- [持久保留策略](https://github.com/openai/codex/blob/c1382380de69521303b416720a52f42d51af6248/codex-rs/rollout/src/policy.rs)：多种 transient 通知不持久化，包括 delta、warning/error、raw-response 通知和 ShutdownComplete。归档包含 durable retained records，不等于 app-server 全通知流。

官方源码还支持原 thread 的后续恢复、元数据追加、revert / rollback 和压缩。单次实际追加观察不能签署永久 append-only；模型完成、文件大小稳定、进程退出都不替代关闭确认。

## 实际运行与验证

所有本轮中间产物均在当前已忽略目录 `dist/relay-l1`；每次真实测试创建新的、Agent 持有的公开 Git 项目，只发送固定公开任务 `PUBLIC_L1_FIRST / SECOND / RESUMED`，原生文件、命令、插件、应用、MCP、网页、图像和其他工具关闭。认证仅由实际 CLI 使用自身现有配置，主机没有读取或复制 auth.json / 用户 config。仅访问自身 controller 返回的精确新 thread 文件及测试 SQLite，未打开正式原档。

真实测试命令的共同环境：

```powershell
$env:ASTROCYTE_RELAY_L1_NATIVE = '1'
$env:ASTROCYTE_RELAY_L1_EVIDENCE = (Resolve-Path 'dist/relay-l1').Path
```

| 命令 | 结果与证据 |
|---|---|
| `go test -mod=readonly ./internal/workspace/app -run TestRelayOwnedActualCodexLifecycleAndSQLite -count=1 -v -timeout=210s` | PASS，48.16s；实际 3 次模型任务，同文件增大、真实 Stop / Shutdown、原 native ID / binding Resume、cold SQLite reload；新线程通知前缀和关闭链 confirmed，恢复 raw gap 保持，最后 85 条派生通知。日志 `current-actual-lifecycle-identity.log` |
| `go test -mod=readonly ./internal/adapters/agents -run TestRelayActualCodexProtocolSourceStart -count=1 -v -timeout=90s` | PASS，5.26s；真实 Codex 0.162.0 新 thread 启动、8 条通知、零模型调用、owned 强制退出；缺少 thread/closed 时 stream / archive finalized 均 unknown。日志 `current-actual-forced-exit.log` |
| `go test -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree -run 'TestRelay\|TestNeutral' -count=1 -timeout=90s` | 4 package PASS；日志 `current-scoped-tests.log`。既有 contract fixtures 不计为本轮真实控制器证据 |
| `go test -mod=readonly ./...` | 最终完整重验 PASS，20 个有测试 package / 9 个无测试 package；日志 `current-full-go-test-repeat.log`。未设置真实环境开关的默认测试仍按原门控 SKIP，真实命令另列 |
| `go test -mod=readonly ./internal/adapters/agents -run TestWindowsShimSpacesAndDescendantCleanup -count=1 -v -timeout=60s` | 同源码独立 PASS，1.43s；日志 `current-inventory-diagnostic.log` |
| `go vet -mod=readonly ./internal/workspace/domain ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents ./internal/adapters/relayworktree` | EXIT 0 |
| `go build -mod=readonly -o dist/relay-l1/current-runtime-server.exe ./cmd/server` | EXIT 0；产物忽略，不改 H 入口 |
| `node scripts/check-architecture.mjs` | PASS，29 package 边界 |
| `gofmt` / `git diff --check` | PASS，本次 relay 写域 |
| `git -c http.version=HTTP/1.1 push origin songconmaisaix31-design/works-relay-runtime` / `ls-remote` | SOURCE 远端精确一致；REPORT 另发 |

安装、配置、可启动和原生能力分别核实：真实 installed entrypoint / version 为 Codex 0.162.0；Registry Configured=available 来自实际 thread 配置；Startable=available 来自原生 start；3 次模型 turn 完成另证真实模型可响应。Observe / Send / Stop / Resume 本次均实际执行，不据客户端名称推断其他能力。活性实测使用显式 10 秒测试窗口：真实新活动为 live，11 秒静默后 unverifiable 且 LastEventAt 不变，真实退出为 exited。该测试窗口不成为产品默认值。

首失败原件保持：`actual-native-first.log`、`actual-notification-first.log`、`actual-protocol-start-diagnostic.log`、`scoped-after-fast-fail.log` 是接管前的真实失败；本轮 `current-actual-lifecycle-first.log` 和 `current-actual-lifecycle-diagnostic.log` 保存恢复绑定 RED，`current-full-go-test.log` 保存库存 shim 清理 RED。完整重验通过不覆盖这些记录，库存领域已 Handoff B。无自动模型重发，各次实际测试使用独立新公开项目；原 UNKNOWN / 未结算状态不改。

## 真实限制与交接

- R 的真实 controller / provider / SQLite 生命周期已验证。H 的实际 HTTP 服务、逐原生 event_id 的数据库回查、F 首次 DOM、从未见 CLI 的三分钟演示和组合 release 验收仍由对应 owner / 主控独立执行。本报告不将 R 测试提升为整条 L1 完成。
- H 已收到精确 SOURCE 和具体限制：原档 retained records、通知完整范围、current-byte-range、writer shutdown 和永久 append-only 分列；H 当前文件模式的 PrefixComplete / SessionComplete 保持未知，通知模式可单独启用。生产获准单个公开 Git 根的组装由 H 负责，默认来源许可不扩大。
- B 收到普通库存测试首 RED 和真实进度上下文 Handoff。R 生命周期公开任务只证明 Session 行为；作品进度 infer 需要 B / H 使用实际服务同一公开 README / 手册和既有领域端口，不能用本测试的固定回复冒充进度验收。
- 观察器最多保留 16384 条派生通知；继承的 Native 原生流 / 事件边界仍有容量限制，达到限制立即返回错误，不能宣称无限流或所有 CLI 原生能力已覆盖。新格式库依赖由 B 独占，R 无 go.mod / go.sum 变更。
- 当前进程结束的 writer 证据不跨服务重启证明仍有原控制器所有权。持久引用和通知可重载；不能据 PID、terminal pane 或数据库记录接管旧 writer、自动 Resume 或补历史。
- 实际用户作品、dirty / untracked / non-Git 来源、自由工具、原件留存和产品活性窗口仍未决定。终端 TTY / terminal-only coding、其他 CLI 完整 Session、race detector、正式 Procmon 零写和真实两日 00:30 验收仍未执行。
- 正式 V4 材料 / 协议 / engine / mapper / 认知 schema / migration / 权重 / 快照流程未修改或读取正式原档。当前部署 17321 保持独立，未合 main、未发布 ZIP、未向外部沟通渠道发送消息。
