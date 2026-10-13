# Works Relay runtime

事实源：`Works-Relay-v1.2.md`；基线 `6ebf523ad4abf4f80d78ee4207fa30168f299ba6`。独占路径为新 relay 前缀 Go 文件、relayworktree 目录及本报告。实际客户端 Codex；Orca terminal show 核实面板模型 `GPT-6.1-Sol high fast`。

早期接缝：`app.RelayRuntime` 提供 CreateIntent（不启动）与 ConfirmIntent（人类身份、权确认及当前项目 revision 必经），其余 List/Get/Send/Read/Stop/Resume/Transition 都由接入 Service 维护唯一状态。`RelayEventSink.EmitRelayEvent` 只传 derived 活动，H-http 通过现有 EventService 装配，事件 mapper 和认知引擎保持冻结。

主控已通过 Orca ask 授权新 `internal/adapters/sqlite/relay_store.go` 中的独立 `relay_state` / `relay_receipts` 两表，文件固定在 LOCALAPPDATA/astrocyte-workbench/v0.2-access/relay.sqlite，不改任何既有迁移，不写认知源或副本。A/B/C 是现有 ProjectSettings/ProjectGrant 的确认摘要，不另造授权系统。

实码调查：现有 NativeAdapter 提供 Start/Send/Observe/Resume/Stop，Codex/Claude/Pi 的既有实现禁用文件与命令工具。主控确认：需要写代码的接力不可把无工具会话当成功；安全确定 structured 不可用时才正常降级 terminal，未确定的 delivery_unknown 保留占用且禁止第二写者。未找到运行中的 generic Git watcher；已有 registeredReadCommand 的 fixed Git 元数据观察可复用为接入端口。

用户尚未回答：Today 实时与快照接缝；未提交/未跟踪及非 Git 单文件中性化范围；活性窗配置。接口不设隐含 60 秒窗口，准备策略保持显式配置，不将未回答变成最终产品决定。

## 工程交接

分支 `songconmaisaix31-design/works-relay-runtime`；SOURCE `a3f47921f4627a8b3e48e44c7cabe0d2b021b617` 已 push。本报告另行 REPORT commit + push，精确 REPORT SHA 在 Orca Handoff 中给出。当前源码相对基线只修改本轨 14 个授权路径，没有既有业务文件、认知 schema、migrations、Go 模块、依赖锁或快照脚本改动。

已实现：

- 单一持久 RelayStore：独立 `relay_state` / `relay_receipts`，复用既有 SQLite driver/连接；BEGIN IMMEDIATE 保护跨 service/store 的互斥。主进程、PID 和 UI 都不另写状态；terminal_id 与 native session 分开，重载不生成替代身份。
- 持久 ready → accepted → working → done → released，创建 ready 也保存初始 edge。每次迁移记录 task/from/to/ts/operator；done 必须人类明确确认，release 必须有真实退出证据。退出不自动完成或重派，未知启动保留占用。
- 人类 CreateIntent 无工作区／进程副作用；ConfirmIntent 必须权限确认与当前 settings revision，默认 A，权限摘要保留 A/B/C 原义。各动作重查现有项目许可与已有 grant；Transition 使用 reconcile。没有自动扩大 scope 或新增 grant。
- 每个 dispatch/send/read/stop 动作先持久 receipt，精确 operation/input 幂等；首 rejected / delivery_unknown 不覆盖，不自动重发。先 neutral worktree 后挂 Agent，同源工作区最多一个 active task，拒绝重复时提示现有 task。
- 数据 AgentSpec 与统一 RelayDriver。structured 复用现有 NativeAdapter，不改 native.go；只有明确 NativePrelaunchFailure + unsupported 才可降级 terminal，未知交付不降级。terminal 命令为可信注册 argv，无平台专用 dispatch 分支，默认执行策略未配置时拒绝启动。
- 首输入保留完整手册与任务。真实 native 事件与 CLI stdout 只保留 derived 活动，丢弃文本；Git HEAD/status/有界 tracked diff 只在独立本机 state 内比较，公开事件不含路径或代码内容。Git 与 turn_completed 都不等于 task done。
- RelayLifecycle.ObserveOwned 只观察已有任务，entrypoint 负责可取消的调用；Shutdown 停止实际 owned process，写 service:shutdown stop receipt 并持久退出。无自动 launch/send/resume。未知进程所有权不靠 PID 推断。
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
| ObserveOwned / Shutdown | PASS：只观察已启动任务；即使旧 grant 撤销，shutdown 可 containment stop，保留 stop receipt、退出和未完成 writer 占用 |
| `node scripts/check-architecture.mjs` | PASS，29 package 导入边界 |
| `go vet -mod=readonly ./internal/workspace/domain ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents ./internal/adapters/relayworktree` | PASS |
| `gofmt -l $relayFiles.FullName` / `git diff --check` | PASS；relayFiles 为本轨 domain/app/sqlite/agents 的 relay*.go 与 relayworktree/*.go |
| `go build -mod=readonly -o (Join-Path $evidenceDir 'runtime-server.exe') ./cmd/server` | PASS；evidenceDir 为 LOCALAPPDATA/astrocyte-workbench/v0.2-access/evidence/runtime，产物在工作树外，HTTP 装配由 H-http 独占 |
| race detector | NOT_RUN，本机 CGO_ENABLED=0；跨 store 真实并发互斥测试 PASS 不冒充 `-race` |

首失败仍保留：早期 `3430302` Domain → apierrors 导入违规由主控静态独验发现，`d2cf9d3` 修复；`bcc60c0` 的写连接先于 DB 归属检查及吞掉 sink 错误由主控静态独验指出，`cf11b1a` 修复。公共 CLI 贯穿首轮在 10 秒 ctx 耗尽时报固定 Git 读失败（RED，10.34s）；同一源码单项诊断随后 PASS，未放宽超时；最新完整 suite PASS 不覆盖首 RED。

外部证据目录：`LOCALAPPDATA/astrocyte-workbench/v0.2-access/evidence/runtime`。首 RED：`controlled-cli-runtime-test.log`；单项诊断：`controlled-cli-runtime-diagnostic.log`；最新完整测试：`final-full-go-test.log`；此前包测试：`first-service-test.log`、`first-process-worktree-test.log`、`scope-publication-lifecycle-test.log`。报告只保留 derived 结果，原日志和 DB 不进 Git。

## 真实剩余门

- 用户未答的标准 CLI worktree 读写／执行确认保持未授权，LaunchPolicy 默认 nil；A/B/C 资料许可不等于 CLI 自由工具权限。需要写代码的 structured 不可因无工具会话假称接力成功。
- neutral source 策略默认未配置；测试仅显式批准 clean committed Git 子集。未提交／未跟踪来源、非 Git 单文件范围等待用户；既有 prepared worktree 的 dirty/untracked relay 状态复用保留。
- 活性窗不设隐含 60 秒值，零表示无 configured live window。CLI 注册数据可以追加但不凭 PATH 宣称 configured/startable；公共 helper 预验不等于真实 Agent 可用性。
- 同名 untracked 文件内容的第二次变化不被当前 Git status 检测；tracked diff 能观察持续改动。当前事件不声称完整采集全部文件活动，也无模型结果推断。
- terminal 当前是标准进程 stdin/stdout 管道，不提供 PTY/TUI 提交证明；要求 TTY 的 CLI 未验证，terminal-only 原生 Resume 显式 unsupported。服务进程重启后旧进程所有权不可证时保持 unknown，不接管或重派。
- Today 的实时本机事件与快照语义仍等用户裁决。H-http 负责 EventService bridge 与 observer/server 生命周期装配；pending event 写入不等于已更新认知状态。
- 真实 Codex/Claude 原生接力与同 worktree 换手、真实 terminal-only coding、完整 UI/E2E、三分钟单 URL、部署、Procmon 正式零写审计、00:30 两日快照验收均 NOT_RUN，需主控与原 owner 在用户决定后协作。没有向其他渠道发消息或修改 validation 工作区。
