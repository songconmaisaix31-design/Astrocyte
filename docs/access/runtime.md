# Works Relay runtime

## 2026-10-15 同实例可信登记与 L 最小事务闭包

当前 Task `task_f082d8b36471` / Dispatch `ctx_f17d901de131` 使用原工作树、原分支和独占写入范围。实际客户端 Codex 0.162.0，终端配置 GPT-6.1-Sol / xhigh / fast，未覆盖模型配置。已读取模块当前管理文件、v1.4 唯一规格和 Release 任务；本阶段没有原生初始化、模型输入或新的三分钟登记起点。

登记 SOURCE `4d93a52028e5ea72c705c5c6e7f30d0cb7f016bb` 提供内部可选端口 `RelayTrustedAgentRegistrar.RegisterTrustedAgents(context.Context, []domain.AgentSpec) error`。原 `RelayRuntime` 方法集合保持兼容，H 的包装器须显式转发该可选端口。受信组装将固定 `agents.json` 经现有成熟 JSON 解码后交完整快照；方法校验全部条目及重复名称、drive / input mode、实际 launch 字段，复制 slice、map 和时间指针，然后在原服务实例原子发布。ListAgents 与内部 spec 读取使用同一锁并返回独立复制；读取不启动或探测控制器。

原 store 事务核未结束任务、工作区占用、当前 owned 会话及 pending / delivery_unknown 收据。已保存的 ready intent 同样保留其登记策略；移除或更改所用登记返回 VersionConflict。同一完整登记可以重载，不涉及在途身份的新增登记可以发布。失败不部分发布，不改动作收据、任务、controller 或当前权限；调用者必须按返回错误记录本次登记失败。实际文件写入起点、写入完成、解析 / reload 和激活时间由 H 沿唯一 t0 记录，当前没有开始该窗口。

依赖消费经模块正式答复 `msg_dbff4cd21c8a` 授权，独立提交 `4b5312f` 原样登记 SOURCE `aa445adf834c3a397a0817bec0593fcf1039420a` 的十一文件：SQLite `workspace_approvals.go`、`workspace_context_packets.go`、`workspace_goals.go`、`workspace_proposals.go`、`workspace_session_bindings.go`，app `proposal_ports.go`，domain `approvals.go`、`goals.go`、`proposal_context.go`、`proposals.go`、`session_bindings.go`。逐项 Git blob 与原 SOURCE 相等，没有消费其他 S2 路径。B SOURCE `ce281d003c093ec26e19dfec78e912f20a014787` 的两文件由 `0ce81b7` 原样消费，019 迁移 SOURCE `da79452d1735f8fceca103aaf9d10bb53182f1f5` 由 `b9cb066` 原样消费；对应内容独立 diff 全等。L 最小闭包不需要 020 / 021 / 022。

所有 Go 命令使用 `GOMAXPROCS=2`、`-p=1 -mod=readonly`，共享 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`，GOTMPDIR / TEMP / TMP 仅本进程使用 R 的 D 盘独占目录。安全日志位于 `D:/Astrocyte-agent-assets/20261015/relay/R/registration/`。

| 命令 / 检查 | 实际结果 |
|---|---|
| `go test -p=1 -mod=readonly ./internal/workspace/app -run '^TestRelayTrustedRegistration' -count=1 -v -timeout=180s` | PASS，2 项：原服务空登记激活、完整失败校验、复制隔离、ready intent 保护、独立新增，以及 16 个并发读写调用者和真实 intent 竞争；包 2.421 秒，`registration-first.log` |
| `go test -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree -count=1 -timeout=180s` | app 13.680 秒、agents 5.630 秒、relayworktree 4.451 秒 PASS；SQLite 原 B `TestProvenUnstartedResumePreservesOwnedOriginalSession` 第 300 行冷 Registry 恢复期望 FAIL，`scoped-packages-first.log` 保留，原 owner 处理 |
| `go test -p=1 -mod=readonly ./internal/adapters/sqlite -run '^Test(AttentionUsesActualWorkspaceTransaction\|AttentionRejectsActualWorkspaceTransactionFromAnotherDB\|AcceptedExperienceActualTransactionLifecycle\|AcceptedExperienceRejectsActualWrongDatabaseAndClosedStore\|AcceptedExperienceActualCancellationCreatesNoEvidence)$' -count=1 -v -timeout=120s` | PASS，5 个顶层、7 个叶场景；包 2.462 秒，实际同事务 commit / constraint rollback、跨库拒绝、关闭与取消检查，`l-minimal-transaction.log` |
| SQLite Relay 来源与不可变收据检查 | PASS，含独立库、真实 hardlink 拒绝和首 RED 保留，`sqlite-relay-transaction.log` |
| `go vet -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree` | EXIT 0，`vet.log` |
| `go build -p=1 -mod=readonly -o D:/Astrocyte-agent-assets/20261015/relay/R/registration/registration-server.exe ./cmd/server` | EXIT 0，`build-first.log`；产物未启动 |
| `git diff --check` | EXIT 0 |

SOURCE 首次 push 和之后独立 ls-remote 首次查询均遇 `SSL_ERROR_SYSCALL`；同 SHA 普通 push 重试成功，没有更改 TLS、代理或公共历史。远端精确 SHA 后续独立核实。上述编译和事务结果使用已经准确登记的最小闭包，没有把越域 WIP 作为源码交付。

H 的固定可信文件实际 reload、生产 HTTP / SQLite 同实例检查和 F observer 窗口继续由原 owner 完成。双完成通道 `verified_complete` / `user_confirmed_complete` 已收到最新用户决定，当前登记 SOURCE 尚未实施该后继变更；开始证据保持独立，完成不能补造 started。L 完整模型输出、quarantine → 人工批准 → 后续检索、实际 native / Today 和最终包仍未通过，Calls / Known 的未决定项继续保持原 Claim 拒绝。旧来源 nlink2 FAIL、raw gap、ArchiveAppendOnly / PrefixComplete / SessionComplete unknown 和全部第一 RED / UNKNOWN / NOT_RUN 独立保留。

## 2026-10-15 准确 turn 关系、原 Start resolver 与 Read 前置

追加 SOURCE `014692e644b08b3acb1f76d696afa728362f4591` 已普通 push，独立 ls-remote 同值；此前 `79d9313` / `dbb0fe7` 保留普通祖先。本阶段继续原 Task / Dispatch，没有另起 CLI 或模型。模块已授权准备共同成立的 Codex 真实 started 正向，H 唯一登记、F 唯一实际 UI 操作者；无 started 客户端的迁移规则仍待用户决定。

`RelayOwnedTurnBindingReader.ListOwnedTurnBindings(ctx, taskID)` 从原隔离 store 私有 controller / history 读取 OperationID、原 session / native / controller / startup operation、实际 turn ID、输入受理与开始时间。RPC turn/start 返回 ID 后保存准确关系；turn/started 只匹配同 ID，通知先到时保留其实际时间。接受关系不能改写，开始时间只能由该真实通知补入。旧缺字段返回空 gap，不按相邻时间或最近输入推定。H 只能在正选来源 guard 接纳后，用该 tuple 核对应 turn 的合法输出；不读取被拒绝 alias，不以通知通过覆盖 archive FAIL。全部 tuple 字段 JSON 隐藏，原 six-field events 不变。

`RelayOwnedNativeSessionResolver.ReadOwnedNativeSession(ctx, platformSessionID)` 由当前原 `agents.Registry` 或 `Native` 实现。它只返回实际 n.start 保存的原完整 Start 会话及原 controller 证据，没有从 DB scalar / NativeID 拼会话，没有 RPC 或启动。普通接入从请求中已存在的 PendingOperation / Operations[pending] 的真实 start / resume pending 收据取 operation；start 核 LastOperationID，resume 的 LastOperationID 允许仍为上次完成动作。命令显式 operation 不一致时拒绝；未记录的旧对象仍 unknown。普通冷 Registry 恢复也必须持有原退出控制器，不能只据持久 NativeID 建立 owned。

structured NativeRead 的显式动作核当前准确 controller、无 unresolved delivery、无 pending 和无确认退出；driver 在任何 RPC 前再读原 done 通道。当前 ownership 仍可表示持有已退出对象，safe reason `controller_exited` 明确该事实，不能推 RPC-ready 或 live。terminal Read 保留缓存进程 / Git 观察，EOF 关闭后可记录真实退出；普通 GET 继续只读历史。Send 的 EOF / running / blocked / done / delivery_unknown / 退出拒绝与人工 release 原独立门禁保持。

使用 `GOMAXPROCS=2` / Go `-p=1 -mod=readonly`：app / agents 全包最终 PASS（`v14-turn-resolver-final.log`），agents 在最后普通 resume 原持有检查后再验 PASS（`v14-resolver-native-final.log`）；SQLite Relay 范围 PASS（`v14-turn-store-final.log`）。vet、server build、diff check 均 EXIT 0，产物 `dist/relay-l1/v14-turn-resolver-server.exe` 未启动。空实际 Registry resolver 拒绝、旧无 turn gap、私有 JSON 隐藏与生产 SQLite / HTTP handoff 拒绝通过；实际有 controller 的 resolver / turn 对应仍等待唯一真实窗口。

本阶段两项首失败保留：`v14-turn-read-gates-first.log` 的旧 controlled CLI 用例在派发时 delivery_unknown，耗时 14.91 秒，随后临时目录清理因占用失败；该用例孤立原断言 PASS（`v14-controlled-cli-isolated.log`），R 将清理登记移到 Confirm 之前，原 10 秒期限和成功断言不变，最终 app 全包通过。`v14-turn-resolver-scoped.log` 的 B `local_agents_narrowing_test.go:300` 恢复用例没有本次原 held process，首次错误保留并交 B 原域迁移；R 没有改 B 测试。原 H reconciliation 缺许可的全 Go 403 FAIL 也保留，当前不宣称全套通过。

S3 的 ExternalTargets 批准维按 B / S2 契约是准确 native CLI ID；已核原 Session.CLI / Provider 可提供该维的真实 ID。私有 endpoint 仍未核，与 CLI 维分开。有效工具集合的实际证据仍不足，保持 unknown，不能固定空 Tools；usage 无完整原 operation 实源保持 nil，S3 保留 lease / 预算。S2 的原人类绑定、实际 Agent 身份、grant、绑定版本与私有字段由原 S2 维护，R 不伪造 RelayTask 映射、不将 Agent 改成人类、不修改 Swarm 业务。

当前共同正向的共享公开 Go 基线及 brief 由 S3 交 H/F，R 只提供上述原域接口。实际 tools 关闭、正选来源、模型 started、合法输出目标核查、人工 done / release、Today 与最终包仍须唯一窗口逐项实际核验；所有旧失败和未知不覆盖。完整 raw、原生路径、tuple、credentials 与模型输出不进入 HTTP / Git / 包 / stdout。

## 2026-10-15 当前所有权纯缓存读取

追加 SOURCE `79d93132967de20fb1fda3d8d74b4edff386178c` 已普通 push，独立远端同值；事实阶段 `1a4f120beff6a9a8d6033c0eed9c0a748f7f3f7e` 和独立 REPORT `a455b7f0e39f606576dbc4c859d1862622fc8f0f` 保留普通祖先。当前 Task / Dispatch 继续活跃，working / 人工完成待答规则未改变。

新增 `RelayCurrentOwnershipReader.GetCurrentOwnership(ctx, humanCaller, taskID)` 返回安全 `RelayCurrentOwnership{Status, Reason}`。读取先核当前 observe、项目 revision / A-B-C / grant 和工作区身份；store 中的历史 ControllerID 不产生当前 confirmed。服务把原 `Sessions` 保存的完整 Start 返回会话交原 driver 的 `CurrentControllerEvidence`，由当前 Registry 的原 owned 对象核准确 controller / startup operation / session / native identity / provider / CWD。缺原 map、旧未记录启动关系或缺完整 ID / project / version / 实际 owned process 均保持 unknown；身份冲突在实际位置拒绝。

该方法只读缓存与原持有对象，没有调用 Native.Observe、thread/read 或其他 RPC，没有 CLI probe、启动、恢复、模型调用、store 写入或 live 时间刷新。所有权与活性分开；持有已退出进程的原控制器对象不表示终端 live。私有证据不进入返回 DTO；没有从历史 scalar 重建 NativeSession。

H 对公共 DTO 使用上述 safe status / 固定 reason，并独立展示已持久化历史事实的时间和来源。实际工具配置核查沿原 RelayStore Snapshot 的完整 session 与同 Registry `Native.OwnedRelayExecution`，只交安全本机证据；history reader 的 session / native 标量不能拼成新的 owned session。raw argv、原生 ID、CWD、原档和 credentials 继续不公开。

`GOMAXPROCS=2` / Go `-p=1 -mod=readonly` 下，app / agents 全包 PASS（`v14-current-ownership-first.log`），追加身份缺失 guard 后权限 / handoff / 原 metadata 回归与空 Registry 当前归属拒绝 PASS（`v14-current-ownership-final.log`）。生产 SQLite 检查确认 ready 任务只能得到 controller_binding_unrecorded / unknown，Agent 读取被 ScopeDenied，读取前后 store 全等；没有生成 controller 或开始事实。vet、server build 和 diff check 均 EXIT 0，产物 `dist/relay-l1/v14-current-ownership-server.exe` 未启动。

带真实原生 session 的正向当前所有权、实际服务重启后的冷 Registry 读取、当前来源 / 工具关闭及所有新 CLI / Today / L1 仍 NOT_RUN，等待语义与 H/F 组合门禁。全套 Go 的原 H reconciliation 许可首 FAIL 保留，未因定向检查覆盖；本阶段没有扩大来源权限或重复旧输入。

## 2026-10-15 v1.4 独立事实阶段

当前 Task `task_f082d8b36471` / Dispatch `ctx_aa83ef2b5087` 延续原 R 工作树、分支及写入范围。已读取模块唯一规格 `tasks/Works-Relay-v1.4.md`，初稿 SOURCE `297d00420f4478dd24805a9ad70ced47fb5b711f` 与人工 handoff 增补 SOURCE `61977021f7babe86205cd23f1b83c23a6627af6c`；此前停止窗口安排已被当前实现授权替代。本阶段 SOURCE `1a4f120beff6a9a8d6033c0eed9c0a748f7f3f7e` 已普通 push，独立 ls-remote 同值。客户端 Codex，模型设置未修改，没有派第四层人员。

现有 RelayStore JSON 事务记录动作执行权、最终动作结果、人工目标确认和控制器产生的独立事实。最终 receipt 保持不可变；事实、关系和人工交接声明不得删除或改写。Confirm / send / read / stop / resume 的原 operation 取得执行权后才进入 driver，动作取得执行权时重新核工作区占用、当前会话与控制器身份。观察事实被拒绝时，保存最终 delivery_unknown、安全拒绝原因及工作区占用，匹配原任务的准确 NativeSession 保留用于停止；后续 send 被阻断。没有从坏证据补造所有权。

原 Native 进程实际启动后创建私有 controller 身份，绑定原 session / provider / CWD 及真实 startup operation。初始化、会话身份确认、请求 pipe 写入、原生 RPC 输入受理、明确开始分别追加事实。Codex turn/started 必须匹配原 thread 和实际 turn/start 返回的 turn；通知先到时保留其实际时间，ACK 到后按相同 turn 关联，并去重。Claude 的 user message 写入仅记录 input_written，system/init 另记录身份确认；Pi 保留自己的 get_state / prompt / agent_start 来源，不创建 Codex turn 身份。没有明确消费确认时单独记录 input_consumption=unknown。terminal 只记录实际 pipe 建立、完整写入与 EOF 关闭，短写立即失败，关闭后 send 仍拒绝；未增加 TUI / paste-ready 协议。

原 `RelayOwnedSession` 与 `RelayNativeSessionCapture` 增加私有 `ControllerID` / `StartupOperationID`，JSON 隐藏，隔离 store 私有 wire 保留。恢复必须从当前 Registry 原持有且已退出的准确 session / CWD / operation 验证，首 input 为空；新 capture 从自身 sequence 0 开始，prefix / finalized 保持 unknown，原 raw gap 和错误计数保留。旧连接的身份与事实保留在私有历史，只能用于已提交事件的归属核查。旧记录缺启动 operation 的可选证据口返回空未记录，不产生新 confirmed 事实；实际 SendOperation 缺原控制器证据仍拒绝为 delivery_unknown。Source guard、单链接限制及正选路径规则未修改。

经模块及根明确确认，原 CreateIntent 支持可选 `handoff_from_task_id` / `handoff_context_confirmed`，两者成对校验。前任务必须同项目和作品、Agent 实际不同、released、真实退出、无 pending、工作区已准备且解除占用；确认时锁定前任务准确 worktree_id。读取只限当前人工操作者可观察的同域安全历史，继续核当前项目 observe 许可和 grant，不读取旧原生上下文。人工声明独立记录，实际新 provider session 确认前执行关系保持 unknown；省略声明保持 new，不能凭相邻任务推导 handoff。fork 明确 UnsupportedCapability，没有创建普通会话代替。

### H / F 消费接口

| 原 R 接口 | 含义与消费要求 |
|---|---|
| `RelayRuntimeFactsReader.ListRuntimeFacts(ctx, humanCaller, taskID)` | 只读已持久化安全事实与关系，核原 observe / revision / grant；不启动控制器、发 RPC 或刷新 live。旧缺事实为空，含义为未记录 / unknown |
| `RelayControllerHistoryReader.ListOwnedControllerHistory(ctx, taskID)` | 内部只读已保存 controller 历史及当前持久指针，核当前权限和原工作区根；JSON 全部私有。只核已提交 ID 的历史归属，不能授予当前进程所有权 |
| `Native.OwnedRelayControllerEvidence(ctx, exactSession)` | 只访问原 Registry 当前实际持有的准确 NativeSession；返回私有连接身份和原进程事实，不读原档、不发 RPC。原完整 Start session 必须由原调用链保存，不能从 NativeID 重建 |
| `RelayOwnedSession.ControllerID / StartupOperationID` 与 capture 同名字段 | H 将新的通知游标按实际连接分开，并保留旧连接游标 / 已提交 IDs。当前 capture 与原启动操作必须一致；新 sequence 不延续旧 stream |
| 原 `RelayIntentCommand` 的两个可选 handoff 字段 | H 在原 OpenAPI / DTO 衔接，F 用生成类型消费；确认框默认未勾选，项目或 Agent 改变后清空。R 完整核实际前置条件 |

本阶段历史 reader 不表达服务重启后的当前所有权。冷 Registry 没有原控制器时当前归属必须 unknown，不能用持久 ControllerID、PID 或历史 confirmed 事实替代。当前所有权的安全纯缓存读取口随后由 R 精确追加 SOURCE 交 H；H 不从私有历史自行拼新 session。公共 session-events 仍精确六字段，HTTP / OpenAPI 由 H 原域维护。

### 实际验证与第一失败

所有 Go 命令使用 `GOMAXPROCS=2`、`-p=1`、`-mod=readonly`。日志均在本机忽略目录 `dist/relay-l1`，没有原始模型数据。

| 命令 / 检查 | 结果 |
|---|---|
| `go test -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree -count=1 -timeout=180s` | 最终四包 PASS，日志 `v14-scoped-final.log`；最后工作区及观察失败检查后 app 全包再验 PASS，`v14-app-handoff-final.log` |
| `go test -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents -run 'TestRelayHumanTransitionProductionPermissionGates\|TestRelayFactsRejectMissingOriginalControllerAndFork' -count=1 -v -timeout=120s` | PASS，`v14-handoff-first.log`；实际生产 HTTP / SQLite 拒绝五类 handoff 声明，省略声明只创建普通 ready；空实际 Registry 拒绝缺控制器 / fork，没有制造成功进程 |
| 实际两个 SQLite 连接、16 个同 operation 并发拒绝 | 一个最终拒绝 receipt、一个结果事实；事实删除 / 改写与最终 receipt 改写全部拒绝，跨连接只读与原状态全等，没有 controller / started / 人工 done / 关系成功 |
| `go vet -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree ./internal/adapters/httpapi` | EXIT 0 |
| `go build -p=1 -mod=readonly -o dist/relay-l1/v14-stage-runtime-server.exe ./cmd/server` | EXIT 0；服务未启动 |
| `node scripts/check-architecture.mjs`，Node memory 512 MB | 29 包 PASS；`git diff --check` EXIT 0 |
| `go test -p=1 -mod=readonly ./... -count=1 -timeout=180s` | 首 FAIL，唯一失败为 H 原 `TestRelayPersistentRealHTTPReceiptsMutexAndRestart`：缺 reconcile 许可却期待 200，实际 403 approval_required；`v14-full-go-first.log` 保留，交 H 原域处理 |

新增生产测试最初因 helper 类型名 `WorksService` 不存在而编译失败，`v14-sqlite-facts-first.log` 保留；改为现有 `Works` 后实际通过。首次四包回归的既有 metadata fixture 缺新 startup operation，新增端口 ScopeDenied 导致 agents FAIL，`v14-scoped-first.log` 保留；旧缺证据按规格返回空 unknown 后原成功断言保持并回归通过，没有新增 skip 或成功协议替身。既有 fixture 回归不计入真实控制器验收。

working 的严格证据与无 started 客户端的人工完成规则仍待用户答复，当前状态机未改变。真实初始化、输入受理与开始、恢复新连接、合法 handoff、观察失败后的实际停止、工具关闭及新来源文件身份正向验证均 NOT_RUN；没有新 native / 模型或旧输入重发。H/F 组合、三分钟 Today、L1 和最终 Windows 包仍待精确语义组合通过后的唯一实际窗口。旧 nlink2 来源 FAIL、原注册三分钟 FAIL、503 accepted 及全部首失败独立保留。

## 2026-10-14 v1.4 技术建议，等待语义确认

本节按模块 `msg_17d0345e078d` 的范围，只提供现有实现、拟改内容与真实验收条件。已完整读取根 `docs/research/orca-rea-20261014.md`、`tasks/Works-Relay-v1.2.md` 和最新 `tasks/Release-v0.1-20261014.md`。四个方向已经确认，具体字段、迁移规则与就绪协议尚未确认；唯一规格草案由模块在 `docs/modules/relay.md` 维护。本节没有改变业务代码或 HTTP 契约，今晚没有启动新的 CLI、模型或 Today 窗口。

### 投递阶段与执行开始

现有 `RelayReceipt.Status` 持久化 pending / accepted / rejected / delivery_unknown，`RelayTask.State` 持久化 ready / accepted / working / done / released。`relay.go` 的 ConfirmIntent 和 finish 保留原 OperationID、原输入比较及首次错误；输入已经成功、后续归档观察失败时仍保存 accepted。`RelayArchiveObservationFailure` 不允许重发原输入。人工完成继续要求现有 Transition 的 human caller、observe、原 revision / A-B-C、所属检查和 completion_confirmed，释放继续要求真实退出与无 pending。

差距在 `relay.go` 的 observe：每个新 `RelayObservation.Events` 都可使 accepted 进入 working，`relay_lifecycle.go` 的 recordGit 也有同样迁移。`Native.Send` 在发送前设置内部 status=running，这个内部值不证明回合已经开始；Codex `turn/start` 成功响应中的 turn ID 证明请求已被原生端受理，实际 `turn/started` 通知另由 `native_process.go` 记录为 turn_started。Relay 目前保留 kind / sequence，但转换时间使用本机观察时间，事件也未公开关联到触发输入的 receipt / OperationID。完成通知、模型文本、stdout、Git 改动、PID 和近期活动都不能互相替代这些事实。

建议在既有 store 内分别保留输入动作结果、执行开始证据、人工完成记录，不新增调度系统。执行开始的内部事实至少需要原 task / terminal / session 身份、对应动作或明确无法关联、证据种类、来源 sequence 和实际观察时间；关联不明时保留 unknown。Codex 只以同 owned thread 的真实 turn_started 进入已确认开始，Pi 的真实 agent_start 可以记录为该 CLI 的开始证据，不能生成不存在的 Codex turn ID。原线程启动或恢复的握手成功只建立会话，不建立新回合开始。

关键待确认项：working 是否仅表示明确执行开始，以及 terminal-only / Claude 缺少独立开始协议时，如何满足 v1.2 的完整接力与人工完成。当前 Claude 记录 native_session_initialized、text 和 turn_completed，没有独立 turn_started；terminal 只记录 cli_activity 与 Git。若采用明确开始规则，它们应显示输入接受、活动存在、开始 unknown，现有仅 working→done 的合法迁移将不能完成这些任务。必须确认完成门禁与该状态组合的完整规则，再修改代码；不能从 stdout、Git 或 result 倒造 turn_started，也不能自行增加 accepted→done / released。

真实验收应逐项比较动作 receipt、原生开始事件与人工 Transition，验证输入被接受但尚无开始事件时仍不显示已确认开始，Git / stdout 活动不会替代原生开始，回合完成不会自动 done。真实 unknown 保留原动作且不重试。旧 task 的 state、首次错误、registered_at、first DOM 及 Transition 历史全部保留；新增事实字段缺失保持 unknown，不能追溯补成 confirmed。

### 原生会话归属

`relay_owned.go` 的 attachOwnedSession / observeOwnedSession 已绑定 task、terminal、worktree、project、agent、SessionID 和 NativeID；`RelayState.Sessions` 保存原完整 NativeSession。`Native.get` 只接受当前 Native 实例持有的实际进程，公共 `OwnedRelayExecution(ctx, exactSession)` 只读取该原控制器和完整 Start 返回身份。`relay_native_capture.go` 的 bind 校验 provider thread ID / CWD，新会话拒绝 forkedFromId；恢复校验原 thread，并在路径别名出现时核两个正选文件的 os.SameFile。H 的来源文件单链接保护继续生效。

差距在持久绑定与当前进程所有权的表达。重启后 SQLite 仍有会话 binding，Registry 的进程 map 已为空；binding 本身不证明控制器仍持有进程。建议分别表示持久会话身份、当前控制器所有权、来源覆盖。持久 binding 可以用于历史派生查询；实际 Send / Stop / Resume 必须核当前原 controller 与原完整 session，并按原生恢复流程验证结果。缺进程、缺 ID、缺正选来源的字段保持 unknown，不能以 NativeID 重新建立所有权或扫描历史。

HTTP 由 H 提供安全派生状态、明确 gap、不可读路径的关联标识；本机进程与原档证据通过现有内部端口交接。argv、环境、凭据、archive 路径、账户目录、原始输出均不投影到 HTTP。两个硬链接名称属于同一文件，不能计为两个 session 或两次控制器启动；当前 nlink=2 的 guard FAIL 保留，路径类别不能证明链接创建者。

真实验收包括原 exact session 成功读取当前进程证据、错误 project / CLI / session 身份拒绝、服务重启后有 binding 但无 owned process 保持 unknown、正选文件被替换或存在 alias 时来源拒绝，以及每个实际公开 event_id 在对应真实事件库中定位。当前 portable 文件在 B 已消费，后续扩展只能由 R 修改后交精确 SOURCE，不由 B 改 R 逻辑。

### 输入方式与就绪条件

`AgentSpec.InputMode` 现有 line / stdin_eof，`DriveMode` 现有 structured / terminal / both；它们是可信注册信息。`RelayTerminalDriver.Launch` 核实际可执行入口、固定 LaunchPolicy、原公开根、原工作区与 owned process，然后建立 stdin pipe。submit 的成功只证明写入完成；stdin_eof 另证明输入已关闭，后续 send 被明确拒绝。实际进程存在、CLI 已安装与工具关闭策略分别记录，均不证明客户端已经消费输入。terminal 目前没有 PTY、TUI ready 或 bracketed paste 实现。

已有 structured 就绪步骤可以具体登记：Codex 的 initialize 成功、initialized 写入、thread/start 或 thread/resume 返回原身份与 CWD；Claude 控制 initialize 成功，但会话 init 身份在后续真实 system/init 校验；Pi get_state 验证实际 sessionFile 位于自身隔离根并在 resume 保持原文件。协议握手事实与允许下一次输入还须区分；会话已占用、已停止或 blocked 时原 Native.Send 会拒绝。当前未持久化每个握手阶段的安全就绪证据。

建议注册项声明实际支持的 transport / framing / readiness 条件，运行时记录该条件的真实观测结果及时间。当前 EOF Claude 固定入口可声明“pipe 可写并完成 EOF 提交”，开始消费仍 unknown；不声明 TUI ready。需要 TUI / paste 的条目，在存在公开、可验证的就绪协议和实际 PTY 支持前返回 unsupported，不开放自由命令或用延时、进程存活、stdout 关键词推断就绪。新的就绪字段及允许状态需明天确认；不改变已决定的工具权限。

真实验收应分别核未安装、已安装但未启动、握手未成功、就绪、已占用与已退出；EOF 一次提交与后续 send 拒绝使用真实已装且关闭工具的公开固定任务。公开 help 只能确认参数表，不能证明输入就绪或有效工具集合。今晚未执行这些原生窗口。

### 恢复、fork 与换 Agent 的关系

`RelayNativeDriver.Resume` 使用原任务绑定与原完整 session，首 input 清空；`Native.Resume` 要求 StopConfirmed 与原 NativeID，实际 thread/resume 返回 ID / CWD 必须一致。原 terminal_id 不因 PID 变化重新生成，工作区在 released 后可复用；换 Agent 的现有流程是新 intent / task / session，保留同一 worktree。terminal Resume 当前不支持。当前没有持久化完整的会话启动、恢复和换手关系记录，也没有受支持的 fork 动作。

建议在原 RelayState 中增加经真实动作产生的关系记录，不改变原任务、会话身份或历史 receipt。原会话 resume 保持 task / terminal / session / provider identity，记录前后实际控制器执行关联及原 OperationID，恢复本身不重发输入、不宣称新回合开始。新任务与换 Agent 各建立新 task / session 身份，显式关联前任务与同 worktree；关联成立要求前任务已合法 release、无 pending、原工作区所属成立。provider fork 保留为未支持，只有未来获得明确授权并由真实 provider 返回新 ID 与 parent ID 时才能记录；不把 resume 或换 Agent 当成 fork。

需确认的关系字段只建议包含操作种类、源 / 目标的安全身份标识、worktree_id、receipt / OperationID、实际时间及 confirmed / unknown。缺旧字段、重启后丢失进程、缺 parent 返回均保持 unknown；不根据历史时间或类似模型输出补关系。关系持久化复用 RelayStore 的现有 JSON 和事务，不修改冻结 cognition schema / engine / mapper，也不新增 scheduler。H 决定公开 DTO 与契约，F 展示已确认含义，B 只维护普通 JobID / OperationID 到原 exactsession / 来源的映射。

真实验收需比较恢复前后相同原生会话与新 owned process、原 input 未重复、恢复后独立新输入的 receipt，以及 Codex→另一已支持 Agent 的同 worktree 新任务关系。服务重启后的只读历史关系仍在，实际控制器缺失保持 unknown。fork 仍未执行，不能填成功。

本阶段静态核查完成，四项拟改规则没有当作功能通过。2026-10-15 先由根结合模块草案确认关键语义，再由 R/H/F 原负责人实现、测试并在唯一操作窗口完成真实 CLI→三分钟 Today→L1；最终包验收仍依赖组合 SOURCE。本报告源码基线为 SOURCE `d14b08dc50014baf9ee636543564b040112cf3b5`，portable SOURCE `0de4e464e3dab342a06fdb37d76ea58a8d5f1f32` 保持普通祖先。

## 2026-10-14 本次归档链接元数据核查

当前 Task `task_d480ceaf0e15` / Dispatch `ctx_1f88d9e6f4f4` 接收最高层授权，只核原 task `0c6522bc-b232-4040-8407-745885644791` 的已获准归档链接元数据。SOURCE 仍为 `d14b08dc50014baf9ee636543564b040112cf3b5`，当前追加只更新独立 REPORT。

R 从 H 明确给出的本机 locator 读取 `app_root`，以 SQLite `mode=ro` / `query_only=ON` 读取 `relay_state` 的单个原任务 binding，再与 H 提供的本次私有 pointer 做同文件身份比较。两者指向同一正选文件，native identity 已存在；归档为 regular，105564 bytes，当前 nlink=2。原档正文、其他 history 和 credentials 没有读取。

实际 `fsutil hardlink list` EXIT 0，完整名称仅保留本机忽略文件。通过 Windows 官方 [FindFirstFileNameW](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findfirstfilenamew) / [FindNextFileNameW](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findnextfilenamew) 查询该文件的两个硬链接名称，使用 `os.path.samefile` 确认两者为同一文件身份。未解析 fsutil 文本。

安全聚合结果：两个名称有相同 basename，一个与原正选文件同 parent，另一个 parent 不同。原名称位于默认 Codex 目录的 sessions，另一个位于 Orca codex-accounts 目录的 sessions；都不在 R / H 工作树或 H 隔离 app_root，也不位于 client_sessions / archived_sessions。路径类别与账户镜像相符属于推断，创建者尚无命令、进程或运行时创建记录确证。系统当前元数据不能恢复创建者历史。

R 当前任务和前一支持任务没有归档链接创建动作。R 调用边界没有 thread/archive / thread/unarchive 请求或硬链接创建；已保存的官方 Codex 0.162.0 归档实现使用 rename。这些是已核源码事实，不能据此确定此次第二链接的实际创建者。

完整原档和链接路径仅在 R 本机 `dist/relay-l1/owned-archive-link-pointer-private.json`、`owned-archive-hardlinks-private.txt` 和 `owned-archive-native-targets-private.json`，安全结果在 `owned-archive-link-safe-aggregate.json`。私有信息不进入 Git、HTTP、包或 stdout。查询定位的第一次失败和 PowerShell Target 查询失败保留在执行记录；随后系统 API 查询和同文件比较通过。

当前来源的 `source_identity_unverifiable_or_aliased` / nlink=2 保持 FAIL，nlink=1 守卫保持。两路径的创建者未确证，R 没有删除或移动它们。先前 15 行 / 18 event IDs 的成功保持当时结果，不能代替当前来源验证。最高层明确：旧 done 状态保留，目标验收未通过；H 核实 done / exited / pending=0 / 对象归属后，由 F 执行清理释放，再准备独立公开新任务。来源调查不阻塞 H 的必要组装与新任务准备，旧模型不重发。

## 2026-10-14 人工完成与释放的权限

当前 Task `task_d480ceaf0e15` / Dispatch `ctx_1f88d9e6f4f4` 的后续授权来自最高层决定 `msg_ac65fadbbe5e` / 回复 `msg_0c1795a84006`。SOURCE `d14b08dc50014baf9ee636543564b040112cf3b5` 已普通 push，独立 REPORT 通过当前 Dispatch 交接。此前单文件证据接口 SOURCE `0de4e464e3dab342a06fdb37d76ea58a8d5f1f32` 保留普通祖先，`relay_execution.go` 内容保持一致。

人工 `working→done` 和 `done→released` 检查当前项目已有的 `observe` 许可。入口在写入 receipt 前通过现有 human 门禁拒绝 Agent 调用；原 intent 的 revision / A-B-C、当前 CLI、项目权限和 grant 撤销继续核实。任务必须持有 prepared workspace，workspace 的 ActiveTaskID / project ID 和 terminal 的 WorktreeID / Agent 必须匹配该任务。完成需要 `completion_confirmed=true`；释放需要 done、持久化的真实 ExitedAt、没有 pending action 或 pending receipt。释放清除该任务的 workspace 占用。

`ReconcileEvents` 明确检查 `reconcile` 许可。它的权限版本和 grant 检查继续使用原任务绑定；人工完成和释放的授权不会扩大事件发布 reconciliation 的权限。

新增 `TestRelayHumanTransitionProductionPermissionGates` 在 R 当前忽略目录内创建实际公开 README Git 项目，使用现有 LocalProjectService、Works 导入、两个实际 SQLite store、NativeRegistry 和实际 HTTP handler。A/B/C 三组都验证：只有 observe 的任务通过权限检查后仍受真实 ready 状态限制；缺少 reconcile 的发布请求返回 ApprovalRequired；Agent transition 返回 ScopeDenied 且没有新 receipt；grant 缺少 observe、grant 撤销和权限版本变化阻断 transition，原 intent 的 state / revision / A-B-C 保持。测试只创建 ready 意图，完成 / 退出 / released 状态由 H/F 的真实原任务验证。

| 验证命令 | 结果 |
|---|---|
| `go test -mod=readonly ./internal/workspace/app -run TestRelayHumanTransitionProductionPermissionGates -count=1 -v -timeout=90s` | A/B/C 全部 PASS；日志 `dist/relay-l1/human-transition-production-first.log` |
| `go test -mod=readonly ./internal/workspace/app -run TestRelay -count=1 -timeout=90s` | PASS；既有回归日志 `dist/relay-l1/human-transition-existing-first.log` |
| `go vet -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/httpapi` | EXIT 0 |
| `go build -mod=readonly -o dist/relay-l1/human-transition-runtime-server.exe ./cmd/server` | EXIT 0，服务未由 R 启动 |
| `gofmt` / `git diff --check` / `node scripts/check-architecture.mjs` | PASS，29 package 边界 |
| `go test -mod=readonly ./...` | 本阶段首次 FAIL，唯一失败在 H 的 `cmd/server/relay_runtime_http_test.go:297`：缺 reconcile 许可的事件 reconciliation 请求实际 403 ApprovalRequired，测试期待 200；原件 `dist/relay-l1/human-transition-full-go-test.log` 保留 |

该全套失败已交 H 在原测试写入范围处理，成功场景需要真实许可和适用生产验证。R 的生产权限测试和构建通过不替代最终全套通过。最高层后续资源决定为 Go 测试使用 `-p=1` / `GOMAXPROCS=2`；已完成的目标检查不重复执行。

H 负责串行部署此 SOURCE，保留原 DB、权限 revision 2、原输入、receipt 和第一次 503；F 是旧 task `0c6522bc-b232-4040-8407-745885644791` 完成与释放的唯一 UI / HTTP 操作人。F 已确认其人工控件使用 observe、事件 reconciliation 控件使用 reconcile，当前真实完成 / 释放结果待其回传。R 没有操作该任务或调用新模型。旧进程退出仍以 H 已确认的实际结果为依据。

单文件消费已收到 B 和产品模块的明确 ACK：B SOURCE `3384a89aac3be6d5670d756eb2d44263784d0cb8` 已普通 push，R 只读比较两工作树的 `relay_execution.go` 内容一致。B 已归还该文件写入范围；B 普通 text / bridge 的映射和 T 的真实取消验证继续由原负责人执行。

## 2026-10-14 Native 内部证据口单文件消费

本阶段 Task `task_d480ceaf0e15` / Dispatch `ctx_1f88d9e6f4f4`，仍在原 R 工作树和 `songconmaisaix31-design/works-relay-runtime` 分支。SOURCE `0de4e464e3dab342a06fdb37d76ea58a8d5f1f32` 已普通 push；REPORT 独立普通提交并由当前 Dispatch 回传精确 SHA。前阶段 `87a54b7` / `63b55d7` 及全部普通历史保留。当前客户端 Codex，未覆盖模型配置，无第四层人员。

只读核实 B 的 `standards-backend` 基线 `15c738f88822e553eb228c12c0fdc71ad07b7531`。公共证据文件使用 B 已有的 Native 类型和字段，R 的采集、项目身份、时间、权限和第一次 RPC 错误记录通过可选接口提供。本阶段的生产消费范围为一个文件。

### 唯一生产消费文件与调用约束

B 通过文件编辑工具从上述 immutable SOURCE 原样消费唯一文件：

```text
internal/adapters/agents/relay_execution.go
```

该文件依赖双方已有的 `agents.Native` / `nativeProcess` 的 `id`、`get`、`mu`、`nativeID`、`cmd`、`done`，已有 `installedCommand` / `nativeError` / `codexDisabledFeatures`，标准库 context / reflect / time 和既有 apierrors / domain.NativeSession。B 使用文件编辑工具原样消费此文件；涉及该文件逻辑的修改由 R 负责。

调用接口为 `n.OwnedRelayExecution(ctx, exactSession)`。调用方必须持有原 Registry 返回的同一个 `*agents.Native`，并保持原 `Start` 返回的完整 `domain.NativeSession`：ID、NativeID、ProjectID、CLI、Ownership 和 Version。实际 controller map 按原 session ID 查找已经启动的 owned process，再次检查原 native ID；没有进程 / 缺少 ID 或项目身份返回 DeliveryUnknown，external / 错 CLI 返回 ScopeDenied。B 在普通 text.go / bridge 中维护 JobID / OperationID 到原 session、controller 和明确来源的映射，T 在自己的写入范围验证它们。

原进程返回真实 PID、实际 done 退出事实、实际 argv 对固定关闭工具配置的匹配、CLI / 原 session Version。原 argv / 环境 / 原档 / 模型文本 / 凭据不返回。`effective_tool_set` 始终 unknown，argv 配置、只读 sandbox 或 tool_calls 0 不单独证明完整有效工具集合。

R 的 project binding、时间、permission response 和第一次 RPC 摘要通过同包可选 `relayExecutionMetadataProvider` 提供，实现在 R 专用的 `relay_execution_metadata.go`。R 继续检查原项目身份并提供实际时间、通知 writer 参数和第一次错误。B 消费公共文件后，`process_metadata_status=unknown`、权限 unknown、开始 / deadline 为未记录零值、first_protocol_failure 缺省；这些字段保持未记录状态。B 的来源、业务权限和项目身份仍由原调用链核实。

R 新增 `relay_execution_test.go` 是可移植纯拒绝测试，只检查缺原身份、无 controller、external、错 CLI 和 canceled read，没有伪造已持有进程或成功状态。B 的生产消费清单不包含该测试；本轮只在只读 overlay 中编译执行它，T 后续独立测试沿自己的写域。

### 对 B 真实基线的只读编译验证

使用本机 `go help build` 核实的标准 `-overlay`。忽略目录的 `dist/relay-l1/portable-native-b-overlay.json` 将公共文件和纯拒绝测试提供给 B 的编译操作。验证前后 B HEAD 均为上面的 15c738f，工作树 clean。B 命令使用以下本机路径：

```powershell
$relayEvidenceOverlay = 'C:/Users/DW/orca/workspaces/Astrocyte/works-relay-runtime/dist/relay-l1/portable-native-b-overlay.json'
```

| 命令 / 工作目录 | 本阶段结果 |
|---|---|
| R：`go test -mod=readonly ./internal/adapters/agents -run 'TestRelayExecutionPortable\|TestRelayOwnedExecution\|TestRelayProtocolFirstFailure' -count=1 -v -timeout=90s` | 四项 PASS，无 Skip / helper / fake 状态；日志 `dist/relay-l1/portable-native-r-first.log` |
| B：`go test -mod=readonly -overlay $relayEvidenceOverlay ./internal/adapters/agents -run TestRelayExecutionPortable -count=1 -v -timeout=90s` | 两项 PASS，B 实际基线编译通过；日志 `dist/relay-l1/portable-native-b-first.log` |
| B：`go vet -mod=readonly -overlay $relayEvidenceOverlay ./internal/adapters/agents` | EXIT 0 |
| B：`go build -mod=readonly -overlay $relayEvidenceOverlay -o C:/Users/DW/orca/workspaces/Astrocyte/works-relay-runtime/dist/relay-l1/portable-native-b-server.exe ./cmd/server` | EXIT 0；服务运行未执行 |
| R：`go test -mod=readonly ./...` | PASS，20 个有测试 package / 9 个无测试 package，既有 native opt-in 未开启；日志 `dist/relay-l1/portable-native-r-full.log` |
| R：`go vet -mod=readonly ./internal/adapters/agents` 和 `go build -mod=readonly -o dist/relay-l1/portable-native-r-server.exe ./cmd/server` | EXIT 0 |
| R：`gofmt` / `git diff --check` / `node scripts/check-architecture.mjs` | PASS，29 package 边界 |

单文件依赖验证阶段没有新的 RED。B 已通过编辑工具消费并提交 SOURCE `3384a89aac3be6d5670d756eb2d44263784d0cb8`，产品模块和 B 均已 ACK，R 只读比较确认公共文件内容一致。实际 owned PID / Send / 取消 / 有效工具集合验证仍按各原负责人的真实任务独立执行。

### 安全 checkpoint 与剩余限制

本阶段 SOURCE / 独立 REPORT 发布后，R 曾暂停写消费文件 `internal/adapters/agents/relay_execution.go`；B 原样消费、普通 commit 和明确 ACK 已完成，该文件写入范围已归还 R。B 继续在普通 text / bridge 中实现映射，涉及公共文件逻辑的修改由 R 负责；T 在自己的测试范围工作。

最终 I 通过普通合并 R / B 原 SOURCE 保留作者和 DAG。旧 task 一次模型与逐事件核查、第一次 503 accepted + thread/read -32603、首次 registry 三分钟 FAIL、旧 PID / 工具证据不足按任务要求保留。人工完成 / 释放的 observe 权限与事件 reconciliation 的 reconcile 权限按本报告当前章节执行。最后集成和 ZIP 验收仍由对应负责人完成。

单文件接口、精确 SOURCE 和验证结果已直接交 M-product / B / T / H 与父模块。raw 源 / credentials / 模型输出不进 HTTP、Git、包或 stdout；V4 / schema / engine / mapper / weights / schedule / 17321 保持冻结。实际模型和取消路径验证、H/F 真实操作窗口及最后 ZIP 仍由原负责人和模块继续执行。

## 2026-10-14 默认真实 Relay 场景支持

本阶段 Task `task_6586f574ed0b` / Dispatch `ctx_1210a540a63b`，原分支 `songconmaisaix31-design/works-relay-runtime`；最终 SOURCE `87a54b762aef0c6a175eab1470508e76aa6ace45` 已普通 push。首 SOURCE `6ed9ae1301dc2e91f0d17a335300697fae3e253f`、通知参数修复 `055a57a18c732ffbbc55d49c605f65c655b8edea`、无新增 Skip 的测试返修 `5c37c261ae1f98fbdab6926415712ef4a5181581` 和阶段 REPORT `26925b92a62babce52e96862719eb6f7269bf950` 均保留普通祖先。最终 REPORT 独立普通提交，精确 SHA 经当前 Dispatch 回传。此前 `d6d4da6` / `07449a6` 没有重写历史。客户端为当前 Codex，未覆盖实际模型设置，没有派第四层人员。

H 仍独占 17342 的公开 Git、DB、控制器及服务组装，F 独占 17343 的实际产品 UI / HTTP 操作。本阶段没有 create / confirm / send / stop H/F 的任务，没有另启 CLI / 模型，未读取凭据或私人历史。协调端报告的首注册三分钟 FAIL 和模型调用 0 保持原结论，不重记 `registered_at` / 首 DOM。只运行本机已装 `codex exec --help`、`claude --help`、`pi --help` 的公开选项探测；安装不代表配置、可启动或模型能力通过。

### 可消费的生产组装接口

原 Relay Native 请求没有传入显式寿命，既有 Native 默认进程 180 秒后退出，影响顺序人工操作窗口。新增 `RelayNativeDriverOptions.SessionDeadlineSeconds` 传给同一个 `NativeRequest.Command.DeadlineSeconds`：零保留既有默认，1 到 1800 秒显式配置；负数或超过 1800 在 Launch / Resume 调用 adapter 前作为确定未启动拒绝。没有延迟伪造退出、续租、自动 Resume 或重发。H 在自己的组装里选择明确验证窗口，例如：

```go
structured := agents.NewRelayNativeDriver(registry, agents.RelayNativeDriverOptions{
    CaptureNativeNotifications: true,
    SessionDeadlineSeconds: 900,
})
```

900 秒只是 H 可显式选用的公开操作窗口示例，未变为默认值或活性窗口。实际可用窗口需以同一进程的 `started_at` / `deadline_at` 和实际退出为准；进程可能提前失败。此选项不是现成环境变量，R 没有修改 config / H 组装；最终 actualserver 参数由 H 交 P / M-release。

H 原 production terminal 组装的 `LaunchPolicy` 为空，EOF_TERMINAL 会明确 unsupported。新增 `agents.RelayPublicEOFTerminalSpec()` 解析已安装 Claude 的直接入口，Windows npm wrapper 复用既有 `installedCommand` 选择实际 native exe 或 `node + cli.js`，不调用 shell、不启动 CLI、不读取 auth / config。固定 argv 为：

```text
-p --input-format text --output-format stream-json --verbose
--safe-mode --tools= --disallowedTools mcp__*
--strict-mcp-config --mcp-config {"mcpServers":{}} --no-chrome
```

本机公开 help 明确 `--tools` 空集合、safe mode、strict MCP 和 stdin text 的语义。该命令没有验证本机 Claude 鉴权或 paid turn，也不承诺操作系统读取隔离。它使用 spec 名称 `claude`、`drive_mode=terminal`、`input_mode=stdin_eof`，首条手册和任务写入后真实关闭 stdin；不能后续 send / native resume。Claude 自然完成会退出，没有为测试保持假进程，F 必须在新确认后的实际窗口操作并读取真实状态。

`agents.NewRelayPublicEOFTerminalPolicy(exactApprovedPublicRoot)` 返回可交 `RelayTerminalOptions.LaunchPolicy` 的门禁。入口拒绝相对或 redirected 根，每次启动对 `Project.Root` / `Workspace.SourceRoot` 核同一目录文件身份，要求完全相同的已装 argv、terminal EOF 和空 `AllowedTools`；附加 Read / shell / 任意脚本或另一个来源都会拒绝。H 先沿现有 public workspace 组装核该唯一根确在独立 verification 下，再显式组装，不能开放任意用户作品：

```go
spec, err := agents.RelayPublicEOFTerminalSpec() // 注册保存由 H 独占
if err != nil { return err }
policy, err := agents.NewRelayPublicEOFTerminalPolicy(exactApprovedPublicRoot)
if err != nil { return err }
terminal := agents.NewRelayTerminalDriver(agents.RelayTerminalOptions{LaunchPolicy: policy})
```

以上是受信组装片段，H 仍负责注册和服务消费；R 不改 registry 起点、已有进程或权限版本，不制造成功状态。

### 工具关闭的本机证据口

新增 `(*agents.Native).OwnedRelayExecution(ctx, exactOwnedSession)`：只核当前 Registry 已持有的 session ID / native ID / project ID / CLI，返回实际 owned PID、开始时间、deadline、真实退出、CLI / version、安全启动策略与禁用 feature 名称。Codex 启动参数必须与本实例固定策略完全一致；通知模式额外核现有 `thread_unload_delay_secs=0` 的 writer 关闭参数。不返回原 argv、环境、工作目录、原档、模型文本或凭据，不发 RPC、不读文件或扫描 history。PID 是本次已持有进程证据，不能据 PID 接管重启前的 controller。

原 `thread/start` / `thread/resume` 响应现读取官方 schema 的 `approvalPolicy` 和 `sandbox.type`；实际明确不符合 `never` / `readOnly` 时立即失败。两项均实际返回符合时内部证据为 `permission_policy=never_read_only`；缺项保持 unknown。已识别字段错型仍沿既有标准 JSON 解码 fast fail，不增加 parser。

`launch_policy=selected_context_tools_disabled_v1` 表示实际 owned 命令使用该固定禁用参数；`effective_tool_set` 始终 unknown。完整有效工具清单仍需 H 只对本次 controller 正选自产来源聚合，不能用 `tool_calls=0`、源码静态 feature 或只读 sandbox 独自证明所有工具关闭。本轮只验证拒绝路径，未将该新证据口声明为已实际执行成功。

首次实际来源错误出现后，新增 `first_protocol_failure` 可选派生字段保留同一进程第一条 RPC method / 数值 protocol_code / 既有白名单 reason / 实际时间。原始 message 不保留，后续失败不覆盖，日志关闭也可由原 owner 的内部证据口核查。错误 reply / success 解码失败、明确 false、reply deadline 或 reply 前进程退出同样给出固定摘要，没有新增重试、fallback、HTTP 或持久证明系统。它不能补回旧二进制没捕获的内容。

旧运行中的二进制没有此新增口，不通过重启新 Registry 反填旧 controller 所有权或新策略。H 对自己的已知 server PID 可读取精确 ParentProcessId 子链，本机保留 CommandLine、输出只含固定禁用项 bool / PID / 时间，再对本次正选 `turn_context` 做安全聚合；不得输出命令全文、全机扫描或重发旧成功补证。R 不代操作 H 当前服务。

B 的旧成功 native ID 只有 ID / model / README mtime / size，不足以正选原档。`Native.get` 只认当前实例内 exact owned session；`ProcessSelectedText` 最后实际 Stop，若已退出且先前没有 capture metadata，`OwnedRelayArchive` 直接 EvidenceMissing，不发新的 thread/read。重启后 map 空保持 DeliveryUnknown，不能从 native ID 建新 ownership / history fallback。该旧证据不足已直接 Handoff B 与 M-product，旧成功 inference 不重复调用。

### F 的每个真实场景前置条件

以下是源码确认的实际生产路径和缺口，未运行的 H / F 默认浏览器场景保持 NOT_RUN。所有写操作沿既有 session / origin、UUID `Idempotency-Key` 和真实 DTO；F 是唯一执行人，H 不与 F 对同 task 双写。每个尺寸 1920 / 1280 的可变动作使用独立真实任务和明确 controller 窗口；同来源 workspace 同时仅一个 active task，需要实际完成退出释放后顺序推进。

| F 场景 / 环境变量片段 | 实际前置条件与动作 | 支持与限制 |
|---|---|---|
| `EOF_TERMINAL` | H 消费上述单根 policy / 已装 Claude EOF spec，F 实际确认后立即进入 read / stop 窗口 | EOF 一次性输入、read 收据与真实 stop 已有实现；实际 Claude 配置、启动及自然完成窗口尚未验收 |
| `EOF_STRUCTURED` | 注册 codex structured + input_mode stdin_eof，实际确认；等首 turn 完成再交 F send | InputMode 只约束 terminal，structured 仍通过官方 turn/start 发送，不伪造 EOF |
| `CONFLICT_INITIAL` / `CONFLICT_REOPENED` | 保持实际 working A；同来源 ready B 经一次真实 confirm 拒绝，持久 rejected 收据 `view_existing_task:A`；F 只 GET | 已有真实 workspace 互斥与收据持久化，不能手写 ready / working / receipt |
| `DUPLICATE` | 同来源 A 仍 working，交 F 新 ready C，由 F 唯一一次 confirm 得 409 | task C 保持原意图；HTTP 现有 required_action 导向人工查看 A，不自动重派 |
| `GENERIC_CONFLICT` | 非互斥 VersionConflict 的 HTTP action 是 `refresh_relay_state`；直接错误 revision 的 confirm 可触发 409 | 正常 UI 使用当前只读 DTO 且旧意图权限变化优先 403，尚无稳定真实 UI 前置触发；不能故意改状态、手改 DB 或伪造响应凑断言 |
| `WORKING` | 首 confirm 接受后由实际 native / Git activity 进入 working；保留原 controller | model turn complete 不表示 task done；GET / reload 不刷新 LastEventAt，live / unverifiable / exited 保持三值 |
| `UNKNOWN_SEND` | 新 working structured 任务，F browser offline 阻断 send，保留 sessionStorage 未知输入；恢复网络只 GET 后 F stop | 不制造 native delivery_unknown，不重发 send，不以浏览器 unknown 写服务原生状态；实际退出仍由 StopConfirmed |
| `COMPLETE_RELEASE` | working，F 人工确认 `working→done`，read，stop 实际退出，再 `done→released` | `completion_confirmed=true` 仅完成阶段；release 必须真实 exited；原 UI 取消确认不会写服务 |
| `ALTERNATE` / `ALTERNATE_PREVIOUS` | 前任务实际 released；该公开项目显式改 external_model_cli 为另一真实已注册 CLI，创建新 ready，再 F confirm | prepared worktree 的 ID / root / branch / created_at 复用；第二 CLI 配置/启动/tool 禁用尚未核实，不能别名装成第二 Agent；项目设置变更不迁移旧意图 |
| 用户描述 `EVENT_PUBLICATION`，F 实际变量 `PUBLICATION_ERROR` | 真生产 sink 出错得到 error，之后 F 明确 `/events/reconcile` 并 GET | 持久 pending/error、显式 reconcile 已有；没有获准真实生产失败触发条件，不新增 fake sink、不手写 error、不删 success assertion 或加 skip |

F 固定命名为 `ASTROCYTE_TEST_RELAY_<SCENARIO>_<1920或1280>_TASK_ID`，另有 `ASTROCYTE_TEST_RELAY_PUBLIC_URL`；ALTERNATE_PREVIOUS 同格式，publication 的准确场景名为 PUBLICATION_ERROR。H 交每个真实 task ID 与状态、前任务及开始 / deadline 窗口，F 使用自己的正常浏览器 helper；R 不填测试环境的 task ID，也不制造任务。

既有实际 production 操作保持：`POST /api/v1/relay/intents` 创建 ready；`POST /api/v1/relay/intents/{id}/confirm` 使用当前 permission_level / project_settings_revision / permission_confirmed；任务 `send` / `read` / `stop` / `transition` 与 `/events/reconcile` 由 F 正常产品调用。完成 JSON 为 `{"from":"working","to":"done","completion_confirmed":true}`；退出后的释放为 `{"from":"done","to":"released","completion_confirmed":false}`。GET `/tasks/{id}` / `/receipts` / `/events` 只是查询，没有确认/补绑定/重发。当前 DTO `task.state/condition`、`terminal.drive_mode/liveness/exited_at`、`worktree` 身份、`receipt.status/error.required_action`、`event_publication.status` 足够表达这些状态，本阶段没有改 HTTP / OpenAPI。

### 本阶段验证和交接

| 命令 | 本阶段结果 |
|---|---|
| `go test -mod=readonly ./internal/adapters/agents -run 'TestRelayNativeDeadline\|TestRelayPublicEOFPolicy\|TestRelayOwnedExecution\|TestRelayProtocolFirstFailure' -count=1 -v -timeout=90s` | 最终四项和真实 PATH 缺入口拒绝子例 PASS，无 Skip；实际已装入口解析 / 目录身份门禁、纯请求拒绝、首错误保留与 raw canary 不泄漏，零 CLI 启动 / 模型调用；日志 `dist/relay-l1/default-policy-protocol-final.log` |
| `go test -mod=readonly ./...` | 最终 SOURCE PASS，20 个有测试 package / 9 个无测试 package；日志 `dist/relay-l1/default-protocol-full-go-test.log`，未设置 opt-in 的既有真实 native 测试仍 SKIP |
| `go vet -mod=readonly ./internal/workspace/domain ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents ./internal/adapters/relayworktree` | EXIT 0 |
| `go build -mod=readonly -o dist/relay-l1/default-protocol-runtime-server.exe ./cmd/server` | 最终 SOURCE EXIT 0，忽略产物，不代 H 当前服务 |
| `node scripts/check-architecture.mjs` / `gofmt` / `git diff --check` | PASS，29 package 边界 |
| `git -c http.version=HTTP/1.1 push origin songconmaisaix31-design/works-relay-runtime` / `ls-remote` | SOURCE 普通 push 与精确远端核实 PASS；REPORT 独立普通提交 |

没有新的 Go RED；旧 Native / Intent 第一失败日志全部保留。模块审阅首 SOURCE 发现新增 `t.Skip` 违反本轮要求，最终 SOURCE 已普通追加移除，改为实际缺入口时明确检查 ProviderUnavailable / 无可启动命令，并继续全部拒绝检查；额外 PATH 缺入口子例在已装环境也实际运行，没有 return / helper / 假模式跳过。新增测试没有创建 controller / helper process、合成成功协议或伪 lifecycle，实际 H HTTP / F DOM / 两视口默认 suite / 第二 CLI / publication error / ZIP 验收均没有由上述 Go 通过填绿。

模块于 13:37 / 13:38 UTC 交来首次实际 UI task `0c6522bc-b232-4040-8407-745885644791` 的查询：确认 operation `906dd04e-6956-4539-af2b-694d33f286c0` 返回 503，dispatch receipt 是 accepted 并有 provider_unavailable，随后任务 working / terminal_exited，事件 published。该组合沿 `RelayArchiveObservationFailure` 表示原生操作已经接受、之后来源读取出错，不能解释成首输入未交付或重发许可。标准解码失败是 DeliveryUnknown，与这次 provider_unavailable 不同。H 于 13:49:50 UTC 补核原日志：上海 21:33:19 `method=thread/read` / `protocol_code=-32603` / `reason=thread_rejected`，本次实际 mode archive、notifications 未 arm；安全摘要不含具体内部错误 message，不能推断确切上游根因或模型失败。

H 明确报告原任务没有重发，后来原 observer 获得同一 owned source，当前 source available / 实际 exit；本次原档 15 完整行 / 105564 bytes，15 origin→SQL ID 匹配、HTTP 18 event IDs / 分页 / 权限 / GET 只读核查通过，assistant / task_complete 各 1。H 自产安全证据保留在原工作树忽略目录 `dist/l1-live/current-20261014/initial-actual-*.json` / `committed-public.json`；私有正选指针只在其本机 RelayStore，不输出到 Git / HTTP / 包。snapshot complete 只描述当前范围，SessionComplete、prefix、finalized 和 append-only 仍未确证；tool_calls 0 不证明工具关闭。以上是 H 的现场 Handoff，R 没有读取其 raw 或重演该任务。新 lifetime / first failure / PID 口不反填这个旧 180 秒进程；首 503 保持 RED，根因未确证，不以之后来源读取或本阶段 Go 通过覆盖。H 自查 Foundation / 旧 Projects 端口未组装产生的 GET nil panic 属于 H 自身薄组装返修，待 F 释放旧任务后串行修，不归 R 改跨域文件。

接口和 SOURCE 已直接交 H `ctx_312d983772a9`、F `ctx_5bac511ada08`、B `ctx_7217777c7bf5` 与模块父 `ctx_e6328a370d56`。T `ctx_8674c456f49f` 已收到 NewRegistry / TextProcessor / LocalProjectService 真实入口、配置时效和真正 Observe / StopConfirmed / 取消 unknown 的要求，不复用 H DB 或重发 B 成功任务。P / M-release 只消费 H 最终 actualserver 参数及精确集成 SOURCE，R 不改包、不复制 raw 配置或凭据。V4 / schema / engine / mapper / weights / schedule / 17321 冻结不动，archive append-only 和恢复 raw gap 的限制原样保留。

## 2026-10-14 同阶段 Native 解码追加返修

最终 SOURCE 更新为 `d6d4da6ac00a196e6bbcf592cfdc0a2a2731dbe6`，同分支普通 push 完成，命令级 HTTP/1.1 `ls-remote` 精确一致。原意图修复 SOURCE `9719e6970927caee3e0b08d4020708e383881880` 与独立 REPORT `392bd81e822148cb2a7cb8f7bd7a83699d0f7d44` 保留普通祖先。本 Dispatch 尚未结算时，模块转交 M-product 的 Native 解码缺口，R 在原写域持续返修；没有启动另一人员或重复真实任务。

`native.go` 原三处 `thread/start` / `thread/resume` result、Pi `get_state` data、Codex `turn/start` result 忽略 `json.Unmarshal` 错误；现均立即返回固定安全 `delivery_unknown`，不以零值身份检查覆盖首解码失败。缺少 turn ID 也拒绝记录发送成功。`native_process.go` 同类 protocol error / success、envelope、Claude control response / request 和当前已支持事件字段解码全部检查返回值；缺失、null、类型错误或截断立即失败，读取器保留 blocked / 原生 capture gap。已识别的 turn 完成缺少 ID / status 时不会发布完成事件；本机 Codex 0.162.0 生成的 experimental schema 确认这些字段 required。数值 JSON-RPC ID 保留标准支持，非法 ID 不转换成任意字符串。拒绝工具请求的实际写入失败也立即停止读取。仍复用标准 JSON 解码，没有新增 parser / 依赖。

新增 `relay_native_decode_test.go` 只有纯 JSON 错误输入：缺失 / null / 错型 / 截断 / trailing data 和错误布尔类型必须返回不可自动重试的安全 `delivery_unknown`，不泄漏输入 canary。没有创建控制器、进程或合成成功协议。完整既有测试中的 contract fixture 仍只作为回归检查，不作为实际原生能力验收。

| 追加 SOURCE 验证命令 | 结果 |
|---|---|
| `go test -mod=readonly ./internal/adapters/agents -run TestRelayNativeDecode -count=1 -v -timeout=30s` | PASS，纯解析失败输入，不产生原生 effect |
| `go test -mod=readonly ./internal/adapters/agents -count=1 -timeout=90s` | PASS，日志 `dist/relay-l1/native-decode-final-adapters.log` |
| `go test -mod=readonly ./...` | PASS，日志 `dist/relay-l1/native-decode-full-go-test.log`；默认 native opt-in 仍 SKIP |
| `go vet -mod=readonly ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents ./internal/adapters/relayworktree ./internal/adapters/httpapi` | EXIT 0 |
| `go build -mod=readonly -o dist/relay-l1/native-decode-runtime-server.exe ./cmd/server` | EXIT 0，产物忽略 |
| `node scripts/check-architecture.mjs` / `gofmt` / `git diff --check` | PASS，29 包边界 |

首编译 RED 为移除最后一个 Native JSON 调用后残留未使用 import，保留 `native-decode-adapters-first.log`；修正后适配器和完整 Go 通过，不覆盖首结果。新源码 / 边界已 Handoff 模块、H 和 B；H/F 当前实际成功派发及逐事件 / DOM 仍独立完成，本阶段没有另启模型或重发成功任务。没有新增 archive finalized / append-only / prefix / session complete 结论，原 raw 恢复 gap 和正式冻结全部保留。最终 REPORT 再独立普通提交，精确 SHA 通过当前 Dispatch 回传。

## 2026-10-14 子 Run 意图权限返修

当前分支 `songconmaisaix31-design/works-relay-runtime`，最终 SOURCE `9719e6970927caee3e0b08d4020708e383881880`，普通 push 完成，命令级 `git -c http.version=HTTP/1.1 ls-remote` 精确一致。本阶段 Task `task_b833b3a06c1e` / Dispatch `ctx_216bf731aaf8` / 子 Run `run_8bbda1c3974c`；模块负责人父 Dispatch `ctx_e6328a370d56`。客户端 Codex，未覆盖模型设置。本段为当前返修记录；下方旧实际 Codex 生命周期结果保留原 SOURCE 归属，不冒称本阶段再次运行。

CreateIntent 原先丢弃 `s.work` 返回的真实项目设置，写入固定权限 A 和 revision 0；现从既有 LocalProjectRepository 取得并持久绑定 `settings.revision` 与现有 A/B/C 语义。ConfirmIntent 先比较原意图与当前项目设置，不同即 `approval_revoked`；请求中的当前版本 / 权限比较仍保留 `version_conflict`。确认不再用请求 revision 覆盖原意图，权限变化后即使更新请求版本也不能绕过旧绑定。原 revision 0 的旧意图保持过期，需要人类创建新意图，不迁移或自动重发。

普通 SOURCE `ce3a519efa5317e642d7488a776b850eebca1093` 保留业务修复和首次测试历史；负责人拒绝新增测试中的替身成功和手改权限后，`9719e69` 仅普通追加替换 `relay_intent_test.go`，业务逻辑与 `ce3a519` 相同。最终新增测试使用自建且实际提交 README 的公开 Git 根、生产 Works 导入 / ProjectFiles / NativeRegistry / relayworktree / LocalProjectService / 两个 SQLite store 和真实 HTTP guard，没有新增替身。实际设置变更、A/B/C 保存、cold RelayStore 重载、当前 cmd 不能绕过旧意图、HTTP 新意图显示真实版本 / 权限、旧意图 HTTP 403 `approval_revoked` 全部通过；store 中无 workspace / terminal / session / owned binding / pending action，磁盘没有创建 neutral worktree。成功 Confirm 不在本测试中制造，仍由 H/F 独占实际服务和控制器执行。本阶段零 CLI 启动、零模型调用、零视觉操作。

| 本阶段命令 | 结果 |
|---|---|
| `go test -mod=readonly ./internal/workspace/app -run TestRelayIntent -count=1 -v -timeout=90s` | PASS；A/B/C 三场生产适配器 SQLite / HTTP 门禁，日志 `dist/relay-l1/intent-production-first.log` |
| `go test -mod=readonly ./...` | 最终 SOURCE 全套 PASS；20 个有测试 package / 9 个无测试 package，默认 native opt-in 仍 SKIP；日志 `dist/relay-l1/intent-production-full-go-test.log` |
| `go vet -mod=readonly ./internal/workspace/app ./internal/adapters/sqlite ./internal/adapters/agents ./internal/adapters/relayworktree ./internal/adapters/httpapi` | 最终 SOURCE EXIT 0 |
| `go build -mod=readonly -o dist/relay-l1/intent-runtime-server.exe ./cmd/server` | EXIT 0，业务树 `ce3a519`；最终 SOURCE 只替换测试，业务树不变 |
| `node scripts/check-architecture.mjs` / `gofmt` / `git diff --check` | PASS；29 包边界，当前所有 Go 变更 |
| `git -c http.version=HTTP/1.1 push origin songconmaisaix31-design/works-relay-runtime` 与同配置 `ls-remote` | SOURCE 普通 push / 精确远端一致；REPORT 独立提交 |

本阶段首 RED 保留：新增测试两次未组装 ProjectFiles / Registry 的 nil pointer 在当前终端执行记录中；`intent-http-sqlite-first.log` 保存 HTTP DTO 错按顶层 task 解码的失败，修正读取真实 `view.task` 后最终生产测试通过。首次 plain `ls-remote` 的 `SSL_ERROR_SYSCALL` 保留在当前执行记录，命令级 HTTP/1.1 精确查询成功；没有修改 TLS、代理或全局 Git。旧失败 / UNKNOWN / 归档缺口不覆盖。

H / F 已收到两个精确 SOURCE，最终业务与首个修复相同；本阶段生产 HTTP 门禁通过不代表 H 服务真实 native 成功派发、逐原生事件采集、F DOM 或 release 接受。H 继续持有 17342 的公开 Git / DB 与唯一 controller，F 为 17343 唯一 UI 操作人。B 已收到生产入口和旧 `local_agents*_test.go` 迁移建议：既有 `agents.NewRegistry`、LocalProjectService、公开 Git / 显式设置 / UUID 操作、StartNativeSession / ObserveNativeSession / StopNativeSession / ResumeNativeSession；completed 只取实际 Observe，stopped 只取 StopConfirmed，不自动重发 unknown。真实 progress infer 已由模块通知 B 独立成功，本阶段没有重复调用。`dist/relay-l1/actualpublicsources.json` 沿原本机指针交接，不进 Git、HTTP 或发布包。

原 `ArchiveAppendOnly=unknown`、文件 PrefixComplete / SessionComplete 未确证、恢复 raw gap 的限制全部保持。原始来源 / 凭据 / 模型输出不进入 HTTP、Git、包或 stdout；没有改依赖、H 组装、前端、正式 V4 / schema / engine / mapper / 权重 / 日程 / 17321。整体产品 / release 验收由对应 owner 和主控继续，不能由本阶段 Go 通过填绿。

## 前阶段实际生命周期记录

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

## v1.4 固定上下文与首输入前来源检查阶段

本阶段 SOURCE 为 `ce0302835cb7477eb606107351a9ab46dd8d4714`，分支 `songconmaisaix31-design/works-relay-runtime`。本阶段沿原 Task `task_f082d8b36471` / Dispatch `ctx_aa83ef2b5087`，没有启动新的原生进程或模型，没有操作旧任务。

`RelayServiceOptions.LaunchContext` 是受信后台组装的可选私有端口。`relayworktree.NewFixedContext` 接收原 `LocalProjects.ReadProjectContext` 实例和明确的 `FixedContextSelection`：实际 ProjectID、批准的绝对项目根、完整 baseline commit、settings revision 和明确相对文件列表。选择在构造时复制，重复项目、重复文件、相对根、不完整 commit 和越界路径立即拒绝。

Confirm 在原唯一 action claim 后核当前 `read_context` 权限和原 Agent grant，再读取一次原项目授权端口。读取前后分别核实际 Git 根身份、准确 HEAD、干净状态和指定文件已被跟踪；packet 的项目、revision、数量、文件名与版本分别核对。读取完成后重新核 `start` 权限，才交原 driver。未配置该端口保留既有任务手册输入；配置后的未选项目拒绝，不能自动改选根或文件。

原项目根是固定源码的读取来源，中性工作区是实际控制器 CWD，两者分别保存。首次公开 Go 候选可使用批准 primary root 的四个固定文件。Swarm 的 actual worktree snapshot / BindSession / Claim 顺序仍由 S2 / S4 / S3 沿原授权端口处理；当前端口不将 primary root 当作新的 worktree snapshot，也不授予自由读取或文件工具。

`RelayLaunch.Packet` 私有传入原 `NativeRequest.Packet`，structured 输入继续使用既有 `packetPrompt`，terminal 首输入使用相同 fixed context framing。正文只在实际输入和原控制器内存中使用；Relay store 保存会话绑定时剔除 `ContextFile.Text` 和 `ContextMaterial.Text`，保留 packet、版本和身份。新增端口没有 HTTP 正文或路径参数，session-events 六字段保持。

H 可在原 `RelayNativeDriverOptions` 同时设置 `CaptureNativeNotifications=true` 与 `BeforeFirstInput`。该回调位于同一个实际 Codex 控制器完成 initialize / thread-start 之后、任何首 turn/start 写入之前。`StartRelayOperationWithInputGuard` 从 armed capture 读取同次官方 Thread.path，不发额外 RPC，不发现其他档案，不另起 CLI。回调收到实际完整 NativeSession、私有正选 archive 和安全 gap；官方尚未提供 path 时传 `nil` 与 `archive_not_yet_available`。提供的 path 非绝对 JSONL 或与 ephemeral 状态冲突时立即拒绝，不能转换为可用路径。

H 在该回调中沿原 source guard 核正选文件的身份和单链接条件。指针已经给出但文件尚未 materialize、实际来源拒绝或身份变化分别保留实际结果；不得删除未知链接、扫描 home 或改用旧档案。回调拒绝会停止原 held controller，保留已经发生的初始化 / session facts、实际退出和 unknown 收据，不发送首输入。未 armed 的控制器、不支持的 provider 和缺失 operation 在进程启动前拒绝该回调。回调通过不证明工具已经有效关闭、整个 session 完整或目标已经完成。

普通 `Native.Send` 现在核原完整会话的 `PendingOperation` 和 `Operations[operation]`，只将实际 `send / pending` operation 传入原发送路径。缺记录返回 `delivery_unknown`，动作或状态不匹配返回 `scope_denied`；不从 SessionID、startup operation 或时间邻近推导新的输入身份。S3 的原 generic Start 空 Message 仍仅初始化 owned / idle，不发送模型首输入；实际 worktree 授权与绑定后 Claim 新 input operation，再沿同一个 Registry / controller Send。

验证共同使用 `GOMAXPROCS=2`、`go -p=1 -mod=readonly`，本机证据均在忽略目录 `dist/relay-l1`：

| 命令 | 实际结果 |
|---|---|
| `go test -p=1 -mod=readonly ./internal/adapters/relayworktree ./internal/adapters/agents ./internal/workspace/app` | 首次 context 阶段 PASS：5.242 / 7.204 / 11.495 秒；`v14-fixed-context-first.log` |
| 同三个 package 的 checkpoint 回归 | PASS：agents 6.938 / app 18.666 / worktree 9.706 秒；`v14-context-checkpoint-final.log` |
| 三个 package，限定 `TestRelayStoredSession / TestRelayFirstInputCheckpoint / TestRelayFactsRejectMissing / TestFixedRelayContext`，`-count=1` | PASS：app 2.269 / agents 1.971 / worktree 3.482 秒；`v14-context-boundaries-final.log`。真实 Git / SQLite / 原 ProjectFiles 读取、源根变化、未提交修改和 baseline 变化拒绝均实际运行；拒绝和隐私测试不构造成功模型或 owned 进程 |
| `go test -p=1 -mod=readonly ./internal/adapters/agents ./internal/adapters/sqlite -run '^TestRelay' -count=1` | PASS：3.636 / 2.482 秒；`v14-context-relay-final.log` |
| `go test -p=1 -mod=readonly ./internal/adapters/agents ./internal/workspace/app ./internal/adapters/relayworktree` | 最终源码 PASS：6.497 / 11.039 秒，worktree 使用同源码 cache；`v14-context-source-final.log` |
| `go vet -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/relayworktree ./internal/adapters/sqlite` | EXIT 0；`v14-context-source-vet.log` |
| `go build -p=1 -mod=readonly -o dist/relay-l1/v14-context-source-server.exe ./cmd/server` | EXIT 0；产物未启动，`v14-context-source-build.log` |
| `node scripts/check-architecture.mjs` / `git diff --check` | PASS，29 package 边界 / EXIT 0 |

本阶段第一次普通 SOURCE push 遇到 `SSL_ERROR_SYSCALL`，保留该首失败；同一普通 push 再执行成功，独立 `ls-remote` 精确核得 `ce0302835cb7477eb606107351a9ab46dd8d4714`。此前完整 Go 的 H 权限断言 RED、控制器测试首 RED、旧 nlink2 来源 FAIL、旧未知交付、旧三分钟 firstDOM FAIL 均保留，没有用本阶段 targeted PASS 改写完整验收。

H / F 仍须消费准确 SOURCE 后组装当前新公开项目和 observer，核实际 argv / 自产权限 / 合法来源，唯一登记保存真实计时起点，再执行已授权的共同 Codex started 正向。首输入前 callback、当前新档案单链接、实际上下文消费、模型候选、准确 turn mapping 的 HTTP / SQLite 回查、Today 两尺寸、人工目标完成 / release 和新 provider handoff 仍未实测。working 严格门禁与无开始客户端人工 done 的待答规则保持现有状态机；Swarm usage / actual toolset 不补造。最终 ZIP、L / S 完整默认验收仍由原 owner 持续处理。

## v1.4 同控制器两阶段输入

本阶段 SOURCE 为 `cc326be57a9a5f3bdd45687b0ac23b972adb8cef`，分支 `songconmaisaix31-design/works-relay-runtime`，当前 Task `task_f082d8b36471` / Dispatch `ctx_99b96d053325`。依据模块准确规格 `77bcde45a8981d7c42fbb1020285df176148bad5` 第3节实施。执行客户端为 Codex；本阶段没有运行新的原生进程、模型、旧输入或原档探针。

`RelayIntentPreparer.PrepareIntent(ctx, caller, taskID, RelayConfirmCommand)` 沿原 Relay service、store、operation、工作区、terminal 和 owned binding 保存 prepare 收据。准备动作核人工身份、确认、当前 settings revision、A/B/C 和 start 权限；驱动必须提供实际两阶段初始化能力。Codex 初始化通过原 armed capture 完成 initialize / thread-start，Message 为空，任务保持 ready。初始化不生成输入写入、输入接受、turn_started、人工 done 或 accepted 迁移。原初始化的 controller / native ID / CWD / startup operation 和实际 metadata 沿原私有记录保存。

原 ConfirmIntent 对已有准备会话使用新的 dispatch operation，重新核当前授权、固定上下文版本、原控制器、原 prepare 收据、无既有输入、无 pending 以及当前工作区归属。事务取得动作执行权后，只调用原 driver.Send；成功受理后进入原 accepted 路径。初始化和首输入不能复用同一个 operation ID；不同并发动作受原 PendingActions / CAS 保护。ready 任务的普通 Send 被拒绝，首次输入须经原确认入口。普通一次性 Confirm 的既有行为保留。

RelayNativeDriver.Initialize 需要 Codex、通知观察器及受信 BeforeFirstInput。初始化只观察同次控制器给出的 metadata；路径尚未提供或文件未 materialize，来源状态仍须由 H 如实标记 NotYetAvailable。实际首输入在原 operation 锁内执行来源 callback，H 消费自己的 events.VerifyRelaySessionSource 核正选文件、授权根、CWD、单链接和实际工具关闭。缺少正选 archive 不能由 callback 返回 nil 放行。拒绝后停止本次原 owned controller，不提交模型输入；只有确认停止且输入写入前的明确拒绝使用 RelayInputNotSubmitted，其他不确定结果继续保留 unknown 和占用。

固定正文取自原控制器内存中的完整 ContextPacket，真实首 Send 使用既有 packetPrompt；SQLite 绑定仍剔除正文。最终输入检查 UTF-8 和512KiB上限。受信 NewRelayGuardedRegistry(stateRoot, guard) 同时提供 generic Start / Send 的相同保护：空 Start 自动 armed，同控制器后续 Send 核原 pending send operation 并携带原固定上下文。该 generic 入口不能替代 Relay task / Today 绑定。冷 Registry 无原控制器时拒绝，持久标识不能创建所有权。

总负责人正式答复 msg_e8f75be9692e 已确认单 Registry 顺序：F 在同 URL 读取空 agents 并确认 observer ready，H 唯一真实登记并保存当刻 t0，F CreateIntent / PrepareIntent 空初始化，H 核同控制器正选来源及工具，F 原 Confirm 首次提交。初始化、来源检查和准备耗时计入同一三分钟，t0 不重记；不隐藏登记或更换控制器。当前没有该新窗口的登记或三分钟结果。

进程环境为 GOMAXPROCS=2、go -p=1 -mod=readonly；后续缓存按模块指令使用 D:/Astrocyte-agent-assets/20261015/go-build/backend，临时目录为独占 go-temp/R。本机证据位于 D:/Astrocyte-agent-assets/20261015/relay/R：

| 命令 | 结果 |
|---|---|
| `go test -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree -run 'TestRelay\|TestNeutral\|TestFixedRelayContext' -count=1 -timeout=150s` | PASS，10.102 / 2.195 / 1.674 / 4.277 秒；two-stage-scoped-final.log |
| `go test -p=1 -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/relayworktree -count=1 -timeout=150s` | PASS，8.693 / 5.222 / 4.130 秒；two-stage-full-owned-packages.log |
| `go test -p=1 -mod=readonly ./internal/workspace/app -run 'TestRelayIntentBindsSQLiteSettingsAcrossReload\|TestRelayStoredSession\|TestRelayHumanTransitionProductionPermissionGates' -count=1 -timeout=90s` | 最终源码 PASS，4.497 秒；two-stage-final-app-boundaries.log。真实 Git / SQLite 的 A/B/C、版本变化、无确认、缺来源 guard、原拒绝收据重读、无输入事实通过 |
| `go vet -p=1 -mod=readonly ./internal/workspace/domain ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree` | 最终源码 EXIT0 |
| `go build -p=1 -mod=readonly -o D:/Astrocyte-agent-assets/20261015/relay/R/two-stage-final-server.exe ./cmd/server` | EXIT0；产物未启动 |
| `node scripts/check-architecture.mjs` / `git diff --check` | PASS，29 package / EXIT0 |
| 普通 push 与独立 ls-remote | SOURCE cc326be57a9a5f3bdd45687b0ac23b972adb8cef 远端精确一致 |

首四包完整检查 guarded-registry-components.log 的 SQLite TestProvenUnstartedResumePreservesOwnedOriginalSession 保留 RED：缺原 controller 的恢复实际返回 delivery_unknown，该 B 测试要求 prelaunch。已交 B / 模块核现有断言与所有权约束，没有修改 B 测试或放宽 cold Registry。随后定向通过不改写该完整检查结果。模块读取活动 WIP 中间类型组合的首编译 FAIL 也独立保留，不作为准确已提交 SOURCE 的结果。

H 须准确消费 SOURCE 后增加准备入口及公共契约，F 须通过同一实际服务操作；空初始化正向、来源和工具关闭、首次提交、逐 event_id SQL、真实目标输出、两尺寸 Today、人工 done / release 与实际发布包均尚未执行。原 nlink2、首 UNKNOWN、raw gap、ArchiveAppendOnly / PrefixComplete / SessionComplete unknown 保留；working / 无 started 客户端人工完成规则仍待用户。L Interpreter 后继接口为 B SOURCE c715342cff4ccfa1b3198275f981385296fe6641，依赖消费方式和实际模型窗口由模块协调。冻结的 V4 / schema / engine / mapper / weights / schedule / 17321 均未修改。

## 原输入 operation 私有证据端口

本阶段 SOURCE `9807d2b6d7f0042bcd59d36ded0b332d74333dc2` 只新增 agents/relay_input_operation.go 与同名测试；仍属同一 R 工作树、分支和当前 Dispatch。`Registry.ReadOwnedInputOperation(ctx, sessionID, operationID, purpose)` 仅从该 Registry 原持有控制器读取完整原 NativeSession、controller evidence、owned execution 和准确 TurnBinding。purpose 只接受 claim_work_item / submit_artifact，operation 必须为 UUID；不从 SQLite scalar、PID 或调用方重建原会话。

Claim 需要原控制器仍有真实进程持有记录且未退出；其他输入正在运行或已 blocked 时拒绝。Claim 可发生在实际 Send 之前，未接受输入保持 InputAccepted=false，不生成虚构回合。Submit 允许原持有控制器已退出，但必须存在相同原 input operation 的真实 TurnBinding，准确 session / native / controller / startup operation 全部相同。

`StopOwnedInputOperation` 只停止原已接受输入对应的原控制器。停止前在同一 operation 锁内核当前 inputOperationID，已经转为其他输入时拒绝；不停止其他任务。已确认退出直接返回现有证据，不重新发送或恢复进程。该端口不替代 S2 的 binding / epoch / caller / grant 检查。

全部返回字段的 JSON 投影为空对象。ToolsKnown / TargetsKnown / CallsKnown / TokensKnown 保持 false；启动策略、输入接受和进程退出均不补造有效工具、外部目标、完整 usage 或任务成果。此阶段不实现 Swarm 的权限和绑定领域；H / S2 / S3 按各自准确端口作必要映射，缺少的实际执行描述继续未知。

从 git archive 的准确 SOURCE 在 D:/Astrocyte-agent-assets/20261015/relay/R/exact-input-source 独立验证，未包含本树后继 L 的未提交修改。共同环境为 GOMAXPROCS=2、共享 backend GOCACHE、独占 go-temp/R：`go test -p=1 -mod=readonly ./internal/adapters/agents -count=1 -timeout=120s` PASS19.212秒，`go vet -p=1 -mod=readonly ./internal/adapters/agents` 和 `go build -p=1 -mod=readonly ./internal/adapters/agents` 均 EXIT0；日志 exact-9807-input-full-agents.log。冷 Registry、错误 purpose、未知停止和私有 JSON 边界实际运行，没有启动原生进程或构造成功协议。

独立 L 依赖按模块明确授权通过 cherry-pick -x 消费 B c715342 的五路径，保作者与来源，提交为37c3b9d。该依赖实际需要尚未提供的 S2 WorkspaceRepository / WorkspaceTx / WorkspaceTransaction / workspaceCAS，组合 Go 检查首 FAIL 保留在 completion-components-first.log，已交原 B / 模块；R 未修改这些领域。B 后继准确 SOURCE f76e25b84b29495994f2e5877b586d4c22669d3f / REPORT ce5a86b0cdc27ea9567d1041a7fccc89795d9395 已由独立远端和普通祖先核实，窄消费为6b7ba3ad14209ace91a89cf2f63ecafafbba390d。模型标识来源区分 configured_route / provider_reported，独立版本缺证据时固定 unknown；不以别名或 CLI 版本补造。

当前原输入端口准确源码可单独消费，后继 L 源码仍待 B 事务前置和适用组合验证后交付。原控制器正向、S3 Member / Ownership 实际组装、真实 tools / targets / usage、模型候选和第二任务经验引用仍未验证；本阶段不宣称 service build、完整 Swarm / L 或最终 release 已通过。

## L Completion Interpreter 包级源码

SOURCE 为 `d7d47d68c80a1ee3f0007de9f9b08ccfd654d158`，本阶段为适配器源码和包级检查。普通 push 后独立 ls-remote 得到同一 SHA。整体服务构建仍有下述 S2 事务前置缺口；真实模型、完整 L 和 Swarm 均未验证。

`agents.NewCompletionExperienceInterpreter(registry, CompletionExperienceOptions)` 直接实现 B 已提交的 CompletionExperienceInterpreter。调用使用 B CompletionExperienceRequest 的 OperationID / Evidence / ResumeOnly，不创建替代请求身份。受信组装必须提供 ReadContext、BeforeFirstInput 和明确 DeadlineSeconds。ReadContext 返回当前批准项目和固定 ContextPacket；首输入 callback 必须包含当前来源、工具、授权及实际模型窗口门禁。它不替代 S2 / S3 的 Claim、预算、绑定或权限服务。模型版本读取 callback 可省略，省略时保持 unknown。

实际调用只选择同一 Registry 的 Codex native。空输入初始化先建立原 controller，确认其配置路由为 gpt-6.1-sol，再在相同原 operation 提交一次固定上下文和 JSON 草稿要求。原 receiver 只接受同 thread / turn 的完整 final_answer item 和成功 turn/completed，不拼接 delta 作为完成结果；额外工具 item、模型路由变化、截断、错误或身份冲突立即失败。outputSchema 只约束 pattern / verdict / condition / exact evidence / refused，JSON 库拒绝外来字段、额外输出、拒绝回答和缺少原 evidence 的草稿。产物只作为 quarantine 候选的输入，不批准、评分或决定业务成果。

原生响应正文仅留在原进程内存。错误返回固定安全消息，配置路径和原生正文不进入 stdout / HTTP / Git / 包。Provenance.Model 使用实际 thread/start 的配置路由，ModelIdentitySource=configured_route；独立 ModelVersion / ModelVersionSource 均为 unknown。只有受信来源确实报告独立版本时才可记录 provider_reported；模型别名和 CLI 版本不能充当独立版本。PolicyVersion=completion-experience-fixed-json-v1 标识本源码实际实现的固定草稿指令和 JSON 约束，不能替代项目授权政策。

同一 Native 的原操作检查串行进行。重复调用核对原固定 Evidence 后只读取原结果；ResumeOnly 在冷 Registry、缺原进程或缺原 evidence 时返回 delivery_unknown，不启动、恢复或重发。已经提交后未知仍保留原 operation，后续仅核对相同原控制器。B 的现有 attention_receipts 保留同 owner / event 唯一操作，候选事务仍由 B 保存和复核。

以下命令使用 GOMAXPROCS=2、共享 backend GOCACHE、独占 go-temp/R；最终包检查另将 TEMP / TMP 指向同一独占 D 盘目录：

| 命令 | 结果 |
|---|---|
| `go test -p=1 -mod=readonly ./internal/adapters/agents -count=1 -timeout=150s` | PASS32.960秒；completion-adapter-final-package.log |
| `go test -p=1 -mod=readonly ./internal/adapters/agents ./internal/attention/app -run 'TestCompletionExperience\|Test.*Interpretation\|TestRelayInputOperation' -count=1 -v -timeout=120s` | PASS8.616 / 3.052秒；completion-final-model-source-rules.log。实际缺前置、冷重放、非法 purpose、未知停止、拒绝/不完整输出、正文隐私和 unknown 模型版本规则通过 |
| `go vet -p=1 -mod=readonly ./internal/adapters/agents` / `go build -p=1 -mod=readonly ./internal/adapters/agents` | EXIT0 |
| `node scripts/check-architecture.mjs` / `git diff --check` | PASS，29 package / EXIT0 |

组合首 FAIL 原件 completion-components-first.log 继续保留：B c715 的 SQLite 适配依赖尚未消费的 S2 NewWorkspaceRepository / WorkspaceTx / WorkspaceTransaction / workspaceCAS。S2 的最小准确十一文件闭包由原 owner 验证并经模块交付后再消费；R 不改 B / S2 领域或整合其无关历史。此包级 SOURCE 不将该服务组合标为通过。先前原输入端口 REPORT c569 首普通 push 的 TLS128 保存在 input-report-push-first-failure.log；同一提交普通重试后远端精确同值。

当前执行门禁按 ROOT msg_9e32e1f32ba5：Calls 计数单位及 Known 完整覆盖两项等待用户，现 Calls / Tools / Targets Known=false 和 Claim 拒绝保持；不登记新 t0、不绕 Claim / 预算提交输入、不扩大工具权限。新登记、controller 和 model 均为0。rawResponse 仅交私有观察设计，尚未扩展 store / schema 或执行额外 RPC，已完成 Responses 的数量和 nullable usage 不充当全部调用/用量。H / B 精确装配、服务组合构建、真实完整非拒绝结果、人工批准和第二任务引用仍未完成。

## 独立 portable Member 契约与原 Registry 映射

契约 SOURCE `b2461140be791596242e9a8d13aba49e32d098ee` 只新增 workspace/domain/relay_member_evidence.go 与同名测试，依赖仅为标准库 context / time。S3 可单独消费该两文件提交，不需要复制 Registry、NativeSession、完整 Relay 状态或其他历史。实际映射 SOURCE `00dbd9c63dce2311518a1c47002947180d770a2f` 只新增 agents/relay_member_evidence.go 与同名测试；H 在原实际服务将同一个 Registry 注入为 RelayMemberEvidenceSource。普通 push 后独立远端核实为00dbd9c。

准确签名为 ReadRelayMemberEvidence(ctx, sessionID, inputOperationID, purpose) 和 StopRelayMemberEvidence(ctx, sessionID, inputOperationID)，均返回 domain.RelayMemberEvidence / error。该接口提案已通过 Orca 直接发送原 S3（msg_9011e5b87393、msg_31111c404faf），消费签名仍待其确认；不据此冻结待答的 Known 语义。

事实包含私有 session / native / controller / startup / input / accepted turn、ProjectID / CWD / CLI / CLIVersion / purpose，以及原控制器持有、当前输入和真实退出确认。接受时间、明确开始时间和控制器启动时间只复制原事实；缺少实际退出观察时间时保留 nil。ControllerHeld 表示此 Registry 当前仍持有原对象，单独持有记录不证明客户端当前可响应，也不刷新活性时间。Claim / Submit 继续使用9807的原完整会话与精确输入检查；S2 / S3 自行检查 binding / epoch / holder / grant，不能由该私有事实获得权限。

Capabilities / Tools / ExternalTargets 均为 nil，ToolsKnown / TargetsKnown / CallsKnown / TokensKnown 均为 false。nullable reported usage 和 completed response 观察量当前均为 nil，未把未知值改为0或空的已知集合。全部 JSON 字段排除，非空私有 canary 的投影仍为 `{}`。读取和停止期间原 controller / startup 身份或退出事实发生变化时立即拒绝，不补造新的所有权。没有新 store、schema、RPC、控制器或模型。

共同环境沿用独占 D 盘临时目录及共享 backend cache，以下检查通过：

| 命令 | 结果 |
|---|---|
| `go test -p=1 -mod=readonly internal/workspace/domain/relay_member_evidence.go internal/workspace/domain/relay_member_evidence_test.go -count=1 -v` | 独立两文件 PASS0.919秒；portable-contract-standalone-first.log |
| `go test -p=1 -mod=readonly ./internal/adapters/agents -run 'TestRelayPortableMember\|TestRelayInputOperation\|TestRelayFactsRejectMissing' -count=1 -v -timeout=90s` | PASS1.205秒；portable-member-refusals-first.log，真实 cold Registry / 原输入缺失与未知停止拒绝 |
| `go test -p=1 -mod=readonly ./internal/adapters/agents -count=1 -timeout=150s` | PASS7.172秒；portable-member-final-package.log |
| `go vet -p=1 -mod=readonly ./internal/adapters/agents` / `go build -p=1 -mod=readonly ./internal/adapters/agents` | EXIT0 |
| `node scripts/check-architecture.mjs` / `git diff --check` | PASS，29 package / EXIT0 |

S3 自有桥接的准确组合检查和实际 Claim / Submit / Stop 正向尚未执行；Known 和 Calls 用户待答维持原拒绝。H 另报告原 Options.Agents 仅构造时读取，没有同 Service / Registry 的运行期登记端口。可信一次性登记设计已交模块与 H，登记锚点、显式可信触发方式等待根准确决定；未新增 HTTP 命令入口、未登记或开始计时。L 服务构建仍等待 S2 十一文件事务闭包，原首 FAIL 与全部0实际原生窗口保持。
