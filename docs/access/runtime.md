# Works Relay runtime

## 最新工程交接：v1.3 L1 owned-session 读侧（2026-10-14）

事实源仅用 `git show 913db499233720c369f7f5bf24965d9bc6a8ba54:tasks/L1-session-capture-plan.md` 及其认知栈 §2／§7 读取，未合 validation 父业务。原分支 `songconmaisaix31-design/works-relay-runtime` 普通 no-ff 消费 I 基线 `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`；当前 L1 SOURCE `62c7d99f1d6b180ef19f90a2112e801702d9fffb` 已 push，REPORT 单独提交，精确 SHA 通过本 Dispatch 的 Orca Handoff 给主控。相对该 I 基线仅本 R 写域 11 个 Go 路径与本报告变化，未向 v1.2 候选插入 L1 源码。

实际客户端仍为 Codex，`orca terminal show` 核实原 pane 为 `GPT-6.1-Sol high fast`。本 Dispatch `task_3303613a7ff7` / `ctx_98f86f7facfd`，原 owner、terminal、worktree、branch 保持；H 是唯一完整 Session 采集 owner。

早期接口 SOURCE `29a400b725ca301ead31d189b353758c992ecfc0` 提前 push 给 H；后续实现 `b640ad506042df37a6074d68a64d36c20b628eee`、架构修复 `aa8899fdd985671dbf3528d045b455048dfeda9c`、ownership 修复 SOURCE `686f104a7550d6dab81604d90e797fe3ace277c7`／阶段 REPORT `9550e526ff26b357369043bbcada88b4f5003055`、provider SOURCE `7e282225f7e4b7277ed36c9f2f77c64501fa56ed`／REPORT `b482ac357dbc56c4855e4a8f1d08cada6800f80e` 保留，完成源码为上述 `62c7d99`。Root 的 Orca ask 明确批准在现有 RelayState 中保留每 task 的最小 owned 引用及现有 JSON 私有 wire，并追加实际 metadata provider 和 Resume 失效工程；不新增表、schema、migration、Attempt 或证明系统，不复制归档原件。

已实现的接缝：

- 独立 `app.RelayOwnedSessionReader.ListOwnedRelaySessions(ctx)`，不扩 RelayRuntime 的 HTTP 接口／fakes。返回内部 `domain.RelayOwnedSession`：绑定 task/terminal/worktree/project/agent、binding_id、实际 drive mode、真实 ExitedAt 与来源状态；WorkspaceRoot／SourceRoot／controller SessionID／NativeID／Ownership／ArchivePath／ArchiveRoot／ArchiveFormat 均 `json:-`。
- `RelayState.OwnedSessions[task_id]` 由唯一 Service 原事务写。new task 使用新 binding／新 native session，同未完成 task Resume 保留绑定并不重放首输入；同 terminal 新任务删除当前 session 时不删除已退出旧 task 引用。不同 Agent 实码创建不同 terminal，旧 terminal.Agent 不被新 Agent 覆盖。当前 scope／revision／Agent 授权变化时旧 closed 引用仍保留，但读侧返回拒权 gap。
- 现有独立 relay.sqlite 两表不变，只扩 JSON private wire 持久上述引用和现有 Session.HistoryPath。HistoryPath 本身从未被旧 Native.Start 赋值且不提供 root／format／CWD 授权证据；单独存在仍 unknown，不猜文件名／家目录／PID／Agent 原生格式。旧 v1.2 状态未记录绑定时只读返回 `binding_not_recorded` gap，不在 GET 补数据或制造 UUID。
- 每次 read 重新核现有 grant、settings revision、observe action 与当前 Agent、source／prepared worktree／terminal／task 对应关系，active 绑定还核当前 controller session 的 ID／NativeID／Ownership／ProjectID／CLI。拒权输出 safe gap，清私有根／归档／原生 ID；不调用 driver／provider、不读 archive、不写 receipt／Store.Update、不更新时间。
- H 装配 `RelayServiceOptions.OwnedRoots = agents.RelayOwnedRoots{}`。窄口复用既有 `CanonicalRoot` 与 redirect 检查，只验证已登记源根与 neutral 根的目录，归档文件身份／CWD／完整读取仍由 H 校验。无该口安全返回 `workspace_unavailable` gap；不改变 launch policy 或 neutral source 决定。
- `RelayObservation.Archive` 仅来自 trusted 内部 `RelaySessionArchive`；可选 `app.RelayOwnedArchiveProvider.OwnedRelayArchive(ctx, NativeSession)` 为原生适配提供精确 source。来源 IDs、CWD、absolute Path／Root 包含关系、非空 Format 必须一致；source mismatch 或 ownership 拒绝不被后续 Archive 字段覆盖。默认无 provider 为 unknown；source available 只说明已正选来源，不等于格式支持、事件覆盖完整或 Today 可见。
- 新 `agents/relay_owned.go` 给既有 `*Native` 实现实际 provider：Codex 先 `n.get` 核已 owned controller 和原 native ID，仅经现有 `p.call` 查询 `thread/read {threadId: original, includeTurns:false}`。严格回验原 ID／已捕获 CWD，null／relative／非 JSONL 路径不选择。正 metadata Path 的直接 parent 是该 exact file 的 containment，不能据此扫描／授权旁边会话；格式标签为当前 Codex 协议的 `codex`，文件存在、格式／CWD／事件覆盖仍由 H 读取前验证。只解码 ID／CWD／Path，不复制 preview／turn 内容；不调用 thread/list、turns/items/list，不读原档／home，不启动 app-server／模型或第二写者。
- 本机 Codex 0.162 的公开 `app-server generate-json-schema --out` 成功，输出仅工作树外 `l1-codex-schema-0162`；`ThreadReadParams.includeTurns` 与 `Thread.path`（nullable／unstable）及 captured CWD 已核。公开 schema generation 不等于真实 controller 验收。Claude／Pi 的现有 typed adapter 尚无已核 metadata-only path＋CWD 组合，返回具体 unsupported；本机 Pi 公开 get_state 实码含 sessionFile／sessionId 而没有 CWD，不能调用 bash 获取 CWD。OpenCode 的现有 bounded HTTP context 也不冒充文件归档来源。
- `ArchiveFinalized` 与 `ArchiveAppendOnly` 为独立 unknown／confirmed：只有 trusted source 的正 metadata 才 confirmed，实际 Wait／Stop 的 ExitedAt 不推导归档 flush 或 append-only。accepted／delivery_unknown 的 Send 或 Resume 使旧 current claims 失效，只有本次新的正 provider metadata 才能恢复；known rejected 无输入影响不冒称重新收尾。新的纯 metadata／安全 ArchiveGap 证据能在既有 observer 原事务保存，不依赖制造活动事件；不变／相同拒绝证据不重复写状态。具体 gap 区分 metadata unsupported／unverifiable／identity_unverified／path_unavailable／unavailable，不含私有路径或原始错误。
- Send／Resume 尚在 pending receipt 阶段时，读侧也立即派生 closure／append-only 为 unknown，避免动作等待时沿用旧收尾确证；只改返回投影，不额外写 store 或 receipt。公开阻塞 Resume 的竞态验证零 Update。

适用验证（Windows / PowerShell、Go `-mod=readonly`）：

| 命令／范围 | 结果 |
|---|---|
| `go test -mod=readonly ./internal/workspace/app -run TestRelayOwned -count=1 -v -timeout=60s` | PASS：真实独立 SQLite 重载、读侧零 Update／receipt／provider 调用、JSON／现有 task transport 无归档或 native 标识泄漏、closed 尾部／快速同 terminal 换 task、新 session、原 Resume 无重派 |
| `go test -mod=readonly ./internal/workspace/app ./internal/adapters/agents -run 'TestRelayOwned\|TestRelayControlledEOF' -count=1 -timeout=60s` | PASS：拒权与错 binding 在目录口前返回、源／CWD／session／root／format 组合拒绝、来源未知／legacy gap、目录验证、真实公共 EOF process 无 native source 明确 gap |
| `go test -mod=readonly ./internal/workspace/app -run 'TestRelayOwnedArchiveCannot\|TestRelayOwnedReadChecks\|TestRelayOwnedDifferentAgent' -count=1 -v -timeout=60s` | PASS：错 native ProjectID／CLI 即使 Archive IDs/CWD 匹配仍持久拒绝；当前 controller 后续换 ProjectID／CLI 也拒绝；换 Agent 保留旧 closed 身份但遵守当前权限 |
| `go test -mod=readonly ./internal/adapters/agents -run 'TestRelayOwnedCodex\|TestRelayOwnedMetadata' -count=1 -v -timeout=60s` | PASS：actual nativeProcess 受控 RPC 请求只原 ID／includeTurns false；source 只 metadata，无文件打开；ID／CWD／null／relative／format 错配拒绝，unowned／stopped／unsupported 零 RPC |
| `go test -mod=readonly ./internal/adapters/agents ./internal/workspace/app -run 'TestRelayNativeDriverReadsActual\|TestRelayOwnedResumeAndAppend' -count=1 -v -timeout=60s` | PASS：真实 RelayNativeDriver 消费新 *Native provider，metadata 不造活动／exit；confirmed → same task Resume → 实际 public fixture append → Stop，未重新证明 closure 时保持 unknown |
| `go test -mod=readonly ./internal/workspace/app -run 'TestRelayOwnedPendingResume\|TestRelayOwnedResumeAndAppend' -count=1 -v -timeout=60s` | PASS：pending Resume 尚未结束时即不返回旧 closure，读侧零 Update；结束及实际 append／Stop 后仍 unknown，绑定不重派 |
| `go test -mod=readonly ./...` | 最新 SOURCE PASS，20 个有测试 package、9 个无测试 package；原环境门控 live tests 不因此成为真实 Agent 验收 |
| `node scripts/check-architecture.mjs` | PASS，29 package，无 app I/O／跨层导入；首 RED 保留 |
| `go vet -mod=readonly ./internal/workspace/domain ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents` | PASS |
| `gofmt -l $relayFormatFiles.FullName`／`git diff --check` | PASS；变量只包含本轨 relay Go 文件 |
| `go build -mod=readonly -o "$env:LOCALAPPDATA/astrocyte-workbench/v0.2-access/evidence/runtime/l1-runtime-server.exe" ./cmd/server` | PASS，产物在工作树外，未改 H 入口 |

首 RED 不覆盖：L1 最初直接在 app stat root 导致架构拒绝 `internal/workspace/app imports I/O: os`，已移到本 R 新 relay adapter，保留 `l1-first-architecture-red.log` 的首次诊断记录。Root 对 `b640ad5` 静态发现 ownership gap 被 Archive 覆盖的缺陷（Handoff `msg_1d3f33e90e95`）；`686f104` 修复并增加上述组合负例。Root 另指出 Resume 只清 exit 而保留旧 closure claims（`msg_a41b0c07fb70`），`7e28222` 修正；不能把阶段 fixture PASS 当作两项缺陷已不存在。RPC fixture 首轮漏 Cancel cleanup callback，`l1-codex-metadata-first-test.log` 为 cleanup panic RED；桥接测例漏既有 Registry 的 capability map 初始化，`l1-native-provider-bridge-test.log` RED。两项都只修 fixture 初始化／join，原 Native 共享实现未变，后续 corrected 日志 PASS。

最新完整测试为 `l1-provider-pending-full-go-test.log`；前阶段完整为 `l1-provider-full-go-test.log`；metadata 正反例为 `l1-codex-metadata-corrected-test.log`，实际 Driver bridge／closure 回归为 `l1-native-provider-bridge-corrected-test.log`、`l1-metadata-gap-resume-test.log`，pending 竞态为 `l1-pending-resume-closure-test.log`。此前 `l1-ownership-complete-full-go-test.log`、`l1-owned-first-test.log`、`l1-source-gaps-test.log`、`l1-adapter-root-test.log`、`l1-final-ownership-test.log` 均保留在工作树外。所有报告只有 derived 汇总。

真实限制／未执行：Codex metadata provider 已有实码，但真实原生 controller 未运行，path nullable／unstable、null 时不可选择，未知真实文件格式和 flush 仍不宣称已可用。CWD 比较仅接受 normalized spelling（Windows 大小写兼容），不可证明的短／长别名差异保守拒绝，不额外探测未授权 CWD。Claude／Pi／OpenCode 的正路径 metadata 仍具体 unsupported／未实现，不是只归为用户未答。terminal-only 当前只有 owned 标准管道和启动 CWD／Git 观察，stdout 没有可验证的 native archive Path；缺少原 controller 的 metadata 口，不允许自动起第二 app-server、扫 home 或以 thread ID 猜路径。Root／H 已收到该具体差距 Handoff。

新格式的 parser／完整行增量／coverage／游标与 Today 消费归 H，R 没有读取模型原件或扩大旧 bounded ReadContext。provider 的正 metadata 是可信主机端口边界，H 仍需读取前核 exact file identity／CWD 与格式；目录检查不是防 TOCTOU 的 OS 隔离保证。归档 finalized／append-only 未证实不宣称完整，旧引用被 current scope 撤回时不绕过权限读取尾部。

真实 Native／terminal-only coding、完整 Session→Today、此前未见 CLI 的三分钟演示、模型／自由 tools、dirty／untracked／nonGit source、原件留存、活性窗口、Procmon 与真实两日门保持 NOT_RUN／未决定。仅 public fixtures 的工程接缝 PASS，本轮未修改 V4／validation 工作区、旧 Native 文件、mapper、认知引擎、权重、正式 schema／库或快照管道。v1.2 候选 CI 属独立 I 基线，本轨检查不覆盖或替代它。

事实源：`Works-Relay-v1.2.md`；基线 `6ebf523ad4abf4f80d78ee4207fa30168f299ba6`。独占路径为新 relay 前缀 Go 文件、relayworktree 目录及本报告。实际客户端 Codex；Orca terminal show 核实面板模型 `GPT-6.1-Sol high fast`。

早期接缝：`app.RelayRuntime` 提供 CreateIntent（不启动）与 ConfirmIntent（人类身份、权确认及当前项目 revision 必经），其余 List/Get/Send/Read/Stop/Resume/Transition 都由接入 Service 维护唯一状态。`RelayEventSink.EmitRelayEvent` 只传 derived 活动，H-http 通过现有 EventService 装配，事件 mapper 和认知引擎保持冻结。

主控已通过 Orca ask 授权新 `internal/adapters/sqlite/relay_store.go` 中的独立 `relay_state` / `relay_receipts` 两表，文件固定在 LOCALAPPDATA/astrocyte-workbench/v0.2-access/relay.sqlite，不改任何既有迁移，不写认知源或副本。A/B/C 是现有 ProjectSettings/ProjectGrant 的确认摘要，不另造授权系统。

实码调查：现有 NativeAdapter 提供 Start/Send/Observe/Resume/Stop，Codex/Claude/Pi 的既有实现禁用文件与命令工具。主控确认：需要写代码的接力不可把无工具会话当成功；安全确定 structured 不可用时才正常降级 terminal，未确定的 delivery_unknown 保留占用且禁止第二写者。未找到运行中的 generic Git watcher；已有 registeredReadCommand 的 fixed Git 元数据观察可复用为接入端口。

用户尚未回答：Today 实时与快照接缝；未提交/未跟踪及非 Git 单文件中性化范围；活性窗配置。接口不设隐含 60 秒窗口，准备策略保持显式配置，不将未回答变成最终产品决定。

## 工程交接

分支 `songconmaisaix31-design/works-relay-runtime`；SOURCE `b19d83635fe10cde9d42642d7444f3771d02f513` 已 push。本报告另行 REPORT commit + push，精确 REPORT SHA 在 Orca Handoff 中给出。当前源码相对基线只修改本轨 15 个授权路径，没有既有业务文件、认知 schema、migrations、Go 模块、依赖锁或快照脚本改动。阶段 SOURCE `16a506b0ceffa10549e3925c6dca66c1a39f34aa`／REPORT `99b7f04eb33aee681c8db740c60e71f52f28eaa9` 保留，输入模式早期接口 `556c6f1645778286a171ea02668ed0456cde3a98` 提前供 H-http 装配。

已实现：

- 单一持久 RelayStore：独立 `relay_state` / `relay_receipts`，复用既有 SQLite driver/连接；BEGIN IMMEDIATE 保护跨 service/store 的互斥。主进程、PID 和 UI 都不另写状态；terminal_id 与 native session 分开，重载不生成替代身份。同 Agent 下一任务复用 terminal_id 时清除已 released 的 native session，新 Launch 使用新 SessionID／空 NativeID；原未完成任务 Resume 保留原 session 且不重放首输入。
- 持久 ready → accepted → working → done → released，创建 ready 也保存初始 edge。每次迁移记录 task/from/to/ts/operator；done 必须人类明确确认，release 必须有真实退出证据。退出不自动完成或重派，未知启动保留占用。
- 人类 CreateIntent 无工作区／进程副作用；ConfirmIntent 必须权限确认与当前 settings revision，默认 A，权限摘要保留 A/B/C 原义。各动作重查现有项目许可与已有 grant；Transition 使用 reconcile。没有自动扩大 scope 或新增 grant。
- 每个 dispatch/send/read/stop 动作先持久 receipt，精确 operation/target/input 幂等；首 rejected / delivery_unknown 不覆盖，不自动重发。TaskID 原已封入比较输入，本次补显式 receipt.TaskID 比较与跨任务重载负例，不记作原功能 RED。先 neutral worktree 后挂 Agent，同源工作区最多一个 active task，拒绝重复时提示现有 task。
- 数据 AgentSpec 与统一 RelayDriver。drive_mode 是默认路径，structured 与 both 均先复用现有 NativeAdapter，不改 native.go；只有明确 NativePrelaunchFailure + unsupported 才可降级 terminal，未知交付不降级。terminal 命令为可信注册 argv，无平台专用 dispatch 分支，默认执行策略未配置时拒绝启动。
- 首输入保留完整手册与任务。真实 native 事件与 CLI stdout 只保留 derived 活动，丢弃文本；Git HEAD/status/有界 tracked diff 只在独立本机 state 内比较，公开事件不含路径或代码内容。新 Git delta 形成登记事件并更新 LastEventAt；不变轮询不刷新，跨来源时间单调而原生 sequence 独立，真实 exited 优先。Git 与 turn_completed 都不等于 task done。
- RelayLifecycle.ObserveOwned 只观察已有任务，entrypoint 负责可取消的调用；无变化的后台 cache／Git 轮询不创建 receipt、不调用 Store.Update，不写完整状态。真正新证据／退出／首采集错误在一次事务持久；提交前复核 task、session、writer、PendingActions，首错误保留后不自动 probe。页面显式 Read 仍逐 action receipt；Shutdown 停止实际 owned process，写 service:shutdown stop receipt 并持久退出。无自动 launch/send/resume。未知进程所有权不靠 PID 推断。
- 主控明确批准 optional `AgentSpec.input_mode`，空／line 保持标准持续 stdin，stdin_eof 只在 Launch 写完整首输入并 close。模式固定在 owned process；即使后来的注册改成 line，Send 也明确 unsupported／rejected，Read／Stop 仍有效。未知模式在任何进程启动前拒绝；nil LaunchPolicy 不因 EOF 获得权限。owned stdin deadline 显式 delivery_unknown，不重写或起第二 writer；EOF 退出不推断 done。
- TaskView.event_publication 明确 none/pending/published/error 与安全计数。sink 失败保留动作 accepted，持久 pending/error；普通后续动作不重发失败 sink，显式人工 ReconcileEvents 才按同一 event origin 复核。
- 在任何写连接／WAL PRAGMA 前以 mode=ro/immutable=1/query_only 预检既有 relay.sqlite；foreign DB、误置认知 DB 与模拟硬链接先拒绝。正式源／副本未打开、迁移或写入。

## 验证

Windows / PowerShell，Go `-mod=readonly`；实际客户端模型见本报告开头。

| 命令 / 场景 | 结果与范围 |
|---|---|
| `go test -mod=readonly ./...` | 最新 SOURCE PASS：20 个有测试 package，9 个无测试 package。既有环境门控 live tests 不因此变成真实 Agent 验收 |
| `go test -mod=readonly ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents ./internal/adapters/relayworktree -run 'TestRelay\|TestNeutral' -count=1 -timeout=60s` | PASS：真实 SQLite 重载、operation 去重、跨 store 竞态、权限拒绝／撤销、持久首 RED、unknown send/launch、仅确定 prelaunch 降级、手册输入、worktree 换手与显式状态机 |
| 公共真实 CLI → Relay Service → SQLite → neutral Git → derived sink | 最新全套 PASS；真实 CLI 改动 worktree，源文件保持 baseline，Git event 入 sink，进程退出保留 terminal_exited，未伪造 done |
| 公共测试 CLI 活跃／静默／真实退出 | PASS：实际输出 = live，静默 = unverifiable，真实 wait/stop = exited；raw stdout 不进入事件；prelaunch 未启动 = unverifiable |
| 既有 DB 误置与模拟硬链接 | PASS：拒绝前后文件字节和 mtime 一致，未产生 WAL/SHM/journal；不涉及正式库 |
| sink failure → 重载 → 显式 reconciliation | PASS：pending/error 可见；不重发启动或普通失败发布；人工复核保留稳定 origin；原 action accepted 不改 |
| `go test -mod=readonly ./internal/workspace/app -run 'TestRelayOperationCannotChangeTask\|TestRelaySameAgentNewTask\|TestRelayGitLiveness' -count=1 -v` | PASS：真实 RelayNativeDriver 的 NativeRequest recording 接缝；同 Agent 新任务清原会话，原任务 Resume 无重放；六类同 operation 跨任务重载均拒绝且原 accepted／unknown／rejected receipt 不变；Git 单调时间／静默／退出优先 |
| `go test -mod=readonly ./internal/workspace/app -run 'TestRelayKnownUnsupported\|TestRelayFallbackRequires\|TestRelayUnknown' -count=1 -timeout=60s` | PASS：structured 与 both 均安全降级；其他 prelaunch 错误与未知启动不降级 |
| 实际 Windows 8.3 路径与 nested repository root | PASS：目录以 os.SameFile 判同一根，不依赖短／长名拼写；真实 8.3 test 有别名时不 skip，无别名明确 NOT_RUN；不同与嵌套根拒绝。H-http 独验真实 8.3 PASS；修复后的 remote Windows CI 尚未在本轨验证 |
| ObserveOwned / Shutdown | PASS：只观察已启动任务；即使旧 grant 撤销，shutdown 可 containment stop，保留 stop receipt、退出和未完成 writer 占用 |
| `go test -mod=readonly ./internal/workspace/app ./internal/adapters/agents -run 'TestRelayBackground\|TestRelayControlledEOF\|TestRelayTerminal\|TestRelayLifecycle' -count=1 -timeout=60s` | PASS：16 次 silent＋16 次 cached event polls 完整状态／receipt 不增，实际新事件一次 Update；首 capture RED／unknown 不覆盖不重试；pending action 不竞争；真实 EOF process 的 8 次 cache polls 零 Update |
| `go test -mod=readonly ./internal/adapters/agents -run TestRelayTerminalEOFUnknown -count=1 -v` | PASS：真正不读 stdin 的公共进程、输入 deadline unknown、同 terminal 拒绝第二进程、EOF Send 零 write、真实 owned stop／exit |
| stdin_eof 公共实进程贯穿 Service／SQLite | PASS：EOF 后才执行，完整手册＋任务逐字保存，Send rejected 首 receipt 重载保持；Stop accepted、exited、未完成 active writer 保留。未调用真实 Codex／Claude |
| `node scripts/check-architecture.mjs` | PASS，29 package 导入边界 |
| `go vet -mod=readonly ./internal/workspace/domain ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents ./internal/adapters/relayworktree` | PASS |
| `gofmt -l $relayFiles.FullName` / `git diff --check` | PASS；relayFiles 为本轨 domain/app/sqlite/agents 的 relay*.go 与 relayworktree/*.go |
| `go build -mod=readonly -o (Join-Path $evidenceDir 'runtime-server.exe') ./cmd/server` | PASS；evidenceDir 为 LOCALAPPDATA/astrocyte-workbench/v0.2-access/evidence/runtime，产物在工作树外，HTTP 装配由 H-http 独占 |
| race detector | NOT_RUN，本机 CGO_ENABLED=0；跨 store 真实并发互斥测试 PASS 不冒充 `-race` |

首失败仍保留：早期 `3430302` Domain → apierrors 导入违规由主控静态独验发现，`d2cf9d3` 修复；`bcc60c0` 的写连接先于 DB 归属检查及吞掉 sink 错误由主控静态独验指出，`cf11b1a` 修复。公共 CLI 贯穿首轮在 10 秒 ctx 耗尽时报固定 Git 读失败（RED，10.34s）；同一源码单项诊断随后 PASS，未放宽超时；最新完整 suite PASS 不覆盖首 RED。

随后 H-http remote Windows CI `38457889686`（H `8d353e3`）及 `38458172696`（H `c3533d5`）在 neutral root 的短／长路径拼写上 RED；`af6e76d1083c73b3fb3d28d0e3c915e99ee014cb` 改为真实目录身份，仍拒绝不同和嵌套仓库根。H 合并后本机真实 8.3 独验 PASS；此 targeted PASS 不覆盖 remote 首 RED。

`post-review-full-go-test.log` 又记录公共 active helper 仅启动输出一次、在 1 秒窗过期后被测试错误要求 live 的 RED（9.75s）。`16a506b` 将 active 场景改为持续真实输出，静默实进程负例保留，源 Git fixture 准备在操作计时前；真实 Service 的 neutral Prepare 仍在 Confirm 内。未放宽活性窗或 10 秒操作超时；子进程与输出 goroutine 随 owned shutdown／真实退出结束。适用单测与完整 suite PASS，不覆盖该 RED。`native-new-task-session-test.log` 记录新增 recording 测例最初使用固定钟早于 native 桥真实事件而拒绝迁移的测试 fixture RED；该测例改用真实钟后 PASS，未改变生产未来事件拒绝规则。

EOF 首轮 `terminal-eof-test.log` 为新增测试遗漏两个 import 的编译 RED，补 import 后 `terminal-eof-corrected-test.log` PASS。`eof-unknown-write-test.log` 保留已创建进程在启动耗时后返回原始 context error 的 RED，以及测试过晚注册 cleanup 引起的临时目录清理失败；修为 owned submit 明确 unknown、测试启动前登记 owned Stop cleanup 后 `eof-unknown-write-corrected-test.log` PASS，不放宽 100ms write deadline。父测试进程持有的既有 owned Job 随其退出关闭，后续测例显式 Stop 作为主要 containment；没有按 PID 猜测／终止其他进程。

外部证据目录：`LOCALAPPDATA/astrocyte-workbench/v0.2-access/evidence/runtime`。首 RED：`controlled-cli-runtime-test.log`；单项诊断：`controlled-cli-runtime-diagnostic.log`；最新完整测试：`eof-noop-full-go-test.log`；前阶段完整测试：`review-complete-full-go-test.log`；复核负例：`identity-target-git-test.log`；默认路径降级：`default-mode-fallback-test.log`；活跃 fixture：`controlled-cli-active-fixture-test.log`；EOF 与 noop：`noop-eof-combined-test.log`；此前包测试：`first-service-test.log`、`first-process-worktree-test.log`、`scope-publication-lifecycle-test.log`。报告只保留 derived 结果，原日志和 DB 不进 Git。

## 公开 CLI 输入接缝调查（只读）

仅运行 public `--version`／`--help`，限定 PATH 与已知安装入口 stat；未读取配置、凭证、历史或会话，未启动模型或自由工具。`installed` 可确认不代表 `configured`／`startable`／workspace tool 能力；后三者保持 unknown，真实 runtime NOT_RUN。这里的协议描述来自既有 NativeAdapter 实码和公开 help，不是本轮真实 Agent 调用验收。

| 客户端／版本 | 已核实可信直接 argv 前缀 | 原生复用与 terminal 候选限制 |
|---|---|---|
| Codex 0.162.0 | `["node","C:/Users/DW/AppData/Roaming/npm/node_modules/@openai/codex/bin/codex.js"]` | existing app-server NDJSON；`exec -` 声明标准输入 prompt，候选一次性 EOF。help 未验证 EOF／keep-open 增量输入实际边界 |
| Claude 2.1.238 | `["C:/Users/DW/AppData/Roaming/npm/node_modules/@anthropic-ai/claude-code/bin/claude.exe"]` | existing print stream-json NDJSON；`--print --input-format text` 非交互候选，EOF 实际未知。旧 cli.js 不存在；print 跳过 workspace trust 不等于获准工具 |
| Pi 1.0.1 | `["node","C:/Users/DW/AppData/Roaming/npm/node_modules/@earendil-works/pi-coding-agent/dist/cli.js"]` | existing `--mode rpc` JSON；`--print` 一次性候选，但 help 未声明普通文本持续 stdin 契约 |
| OpenCode 1.18.35 | `["C:/Users/DW/AppData/Roaming/npm/node_modules/opencode-ai/bin/opencode.exe"]` | existing serve loopback HTTP；`run --format json` 是输出模式，不能推断 JSON 输入；增量 stdin 未声明 |
| Qwen 0.24.6 | PATH 入口为 ps1，直接入口未验证 | help 声明 text／stream-json 输入；无现有 NativeRegistry adapter，当前 RelayTerminal 拒绝 ps1，不能登记为已可用 |
| Gemini | PATH 未发现 | 未扩大本机私有搜索，installed／configured／startable 不推断 |

前四项 direct version/help 皆 PASS；日志为 `<cli>-direct-version.log`、`<cli>-direct-help.log`，其余公开 help 按 `<cli>-public-<index>.log` 保存工作树外。所有客户端的实际 EOF／TTY／持续 prompt 帧边界仍未知，普通 helper 的 keep-open stdin PASS 不替代真实 CLI。

主控已批准并在 SOURCE 实施可信 AgentSpec 可选 `input_mode=line|stdin_eof` 数据，空维持 line；拒绝未知模式，不按客户端名写 dispatch 分支。Codex 前缀追加 `exec -` 与 Claude 前缀追加 `--print --input-format text` 是 stdin_eof 候选。Launch 写完整手册＋任务后关闭 stdin；Send 明确 unsupported，Read／Stop 仍有 receipt，Resume unsupported；退出仍不 done、不重派、不起第二进程。只在完整输入＋EOF 后执行的公开 helper 已证明适配器；真实客户端仍依赖用户工具授权与实测。capability 的 supported 必须来自实际运行，不能用 help 冒充；可选字段是登记输入策略，真实能力依旧 unknown。

如需交互 TTY，现有 Orca `terminal create/send/read/wait/close` 是可复用的 PTY 资源口，不必新增 PTY 库或 scheduler。最小差距：create 使用 shell 文本而非 argv，可信参数需正确编码；neutral worktree 路径能否挂既有 Orca 托管上下文尚未确认；opaque terminal 所有权／真实 exit 须经 RelayStore 保持，raw read 在边界丢弃。只凭 PATH 或终端 pane 存在不能证明这些接缝已实现。

## 真实剩余门

- 用户未答的标准 CLI worktree 读写／执行确认保持未授权，LaunchPolicy 默认 nil；A/B/C 资料许可不等于 CLI 自由工具权限。需要写代码的 structured 不可因无工具会话假称接力成功。
- neutral source 策略默认未配置；测试仅显式批准 clean committed Git 子集。未提交／未跟踪来源、非 Git 单文件范围等待用户；既有 prepared worktree 的 dirty/untracked relay 状态复用保留。
- 活性窗不设隐含 60 秒值，零表示无 configured live window。CLI 注册数据可以追加但不凭 PATH 宣称 configured/startable；公共 helper 预验不等于真实 Agent 可用性。
- 同名 untracked 文件内容的第二次变化不被当前 Git status 检测；tracked diff 能观察持续改动。当前事件不声称完整采集全部文件活动，也无模型结果推断。
- terminal 当前是标准进程 stdin/stdout 管道，支持可信 line／stdin_eof framing，不提供 PTY/TUI 提交证明；要求 TTY 的 CLI 未验证，terminal-only 原生 Resume 显式 unsupported。服务进程重启后旧进程所有权不可证时保持 unknown，不接管或重派。后台首采集错误后保留记录并停止自动 probe，人工明确 Read／containment Stop 仍可执行；没有声称长期常驻／恢复门已全验收。
- Today 的实时本机事件与快照语义仍等用户裁决。H-http 负责 EventService bridge 与 observer/server 生命周期装配；pending event 写入不等于已更新认知状态。
- 真实 Codex/Claude 原生接力与同 worktree 换手、真实 terminal-only coding、完整 UI/E2E、三分钟单 URL、部署、Procmon 正式零写审计、00:30 两日快照验收均 NOT_RUN，需主控与原 owner 在用户决定后协作。没有向其他渠道发消息或修改 validation 工作区。
