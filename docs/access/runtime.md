# Works Relay runtime

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
