# Works Relay runtime

## 2026-10-14 Native 内部证据口单文件消费

本阶段 Task `task_d480ceaf0e15` / Dispatch `ctx_1f88d9e6f4f4`，仍在原 R 工作树和 `songconmaisaix31-design/works-relay-runtime` 分支。SOURCE `0de4e464e3dab342a06fdb37d76ea58a8d5f1f32` 已普通 push；REPORT 独立普通提交并由当前 Dispatch 回传精确 SHA。前阶段 `87a54b7` / `63b55d7` 及全部普通历史保留。当前客户端 Codex，未覆盖模型配置，无第四层人员。

只读核实 B 的 `standards-backend` 基线 `15c738f88822e553eb228c12c0fdc71ad07b7531`：没有 Relay app / domain / capture 类型；`nativeProcess` 没有 R 的 `relayCapture` / projectID / startedAt / deadlineAt / permissionPolicy / firstProtocolFailure。因此此前六文件清单不是 B 的闭合依赖，不能复制一半、临时注释 R 逻辑或把整个 R / H / frozen 树并入 B。本阶段没有修改 B、H、T 或根计划。

### 唯一生产消费文件与调用约束

B 通过文件编辑工具从上述 immutable SOURCE 原样消费唯一文件：

```text
internal/adapters/agents/relay_execution.go
```

该文件只依赖双方已有的 `agents.Native` / `nativeProcess` 的 `id`、`get`、`mu`、`nativeID`、`cmd`、`done`，已有 `installedCommand` / `nativeError` / `codexDisabledFeatures`，标准库 context / reflect / time 和既有 apierrors / domain.NativeSession。没有 app.RelayLaunch、RelayStore、Relay capture、HTTP、认知 schema、配置或依赖变更。B 不消费 native.go / native_process.go / relay_native.go / relay_terminal_policy.go / relay_policy_test.go；不得编改 R 逻辑来适配。

调用接口保持 `n.OwnedRelayExecution(ctx, exactSession)`。调用方必须持有原 Registry 返回的同一个 `*agents.Native`，并保持原 `Start` 返回的完整 `domain.NativeSession`：ID、NativeID、ProjectID、CLI、Ownership 和 Version。实际 controller map 只按原 session ID 找已经启动的 owned process，再核原 native ID；没有进程 / 缺少 ID 或项目身份返回 DeliveryUnknown，external / 错 CLI 返回 ScopeDenied。不会凭 NativeID 创建 session、查私人 history、发 RPC、启动模型或推断进程已经死亡。B 的 JobID / OperationID 到原 session / controller / 明确来源的映射仍归其普通 text.go / bridge，T 只写其测试。

原进程返回真实 PID、实际 done 退出事实、实际 argv 对固定关闭工具配置的匹配、CLI / 原 session Version。原 argv / 环境 / 原档 / 模型文本 / 凭据不返回。`effective_tool_set` 始终 unknown，argv 配置、只读 sandbox 或 tool_calls 0 不单独证明完整有效工具集合。

R 的原 project binding / 时间 / permission response / 首 RPC 摘要通过同包可选 `relayExecutionMetadataProvider` 供给，具体实现移到 R-only `relay_execution_metadata.go`。R 保留原项目身份核查、实际时间、通知 writer 参数和 first failure，不减少原 R 证据。B 仅消费公共文件时没有该可选实现，`process_metadata_status=unknown`、权限 unknown、开始 / deadline 为未记录零值、first_protocol_failure 缺省；不能把这些补为当前时间、任务时间、成功 Start 证明或新工具能力。B 基线没有进程绑定的 project 元数据，公共证据不声称独立验证了它；来源 / 业务权限仍由其原调用链核实，不能以该内部证据口新增授权。

R 新增 `relay_execution_test.go` 是可移植纯拒绝测试，只检查缺原身份、无 controller、external、错 CLI 和 canceled read，没有伪造已持有进程或成功状态。B 的生产消费清单不包含该测试；本轮只在只读 overlay 中编译执行它，T 后续独立测试沿自己的写域。

### 对 B 真实基线的只读编译验证

使用本机 `go help build` 核实的标准 `-overlay`；忽略目录的 `dist/relay-l1/portable-native-b-overlay.json` 只映射 B 的两个虚拟文件到本阶段 R 的公共文件和纯拒绝测试。它没有替换 B Native、contracts、依赖、H 或 frozen 文件；不在 B 磁盘创建或编辑文件，不执行其服务或 CLI。验证前后 B HEAD 均为上面的 15c738f，工作树 clean。

| 命令 / 工作目录 | 本阶段结果 |
|---|---|
| R：`go test -mod=readonly ./internal/adapters/agents -run 'TestRelayExecutionPortable\|TestRelayOwnedExecution\|TestRelayProtocolFirstFailure' -count=1 -v -timeout=90s` | 四项 PASS，无 Skip / helper / fake 状态；日志 `dist/relay-l1/portable-native-r-first.log` |
| B：`go test -mod=readonly -overlay <本机R绝对overlay路径> ./internal/adapters/agents -run TestRelayExecutionPortable -count=1 -v -timeout=90s` | 两项 PASS，实际 B 基线编译闭合；日志 `dist/relay-l1/portable-native-b-first.log` |
| B：`go vet -mod=readonly -overlay <同路径> ./internal/adapters/agents` | EXIT 0 |
| B：`go build -mod=readonly -overlay <同路径> -o <R忽略证据目录>/portable-native-b-server.exe ./cmd/server` | EXIT 0；只构建，不运行服务 |
| R：`go test -mod=readonly ./...` | PASS，20 个有测试 package / 9 个无测试 package，既有 native opt-in 未开启；日志 `dist/relay-l1/portable-native-r-full.log` |
| R：`go vet -mod=readonly ./internal/adapters/agents` 和 `go build -mod=readonly -o dist/relay-l1/portable-native-r-server.exe ./cmd/server` | EXIT 0 |
| R：`gofmt` / `git diff --check` / `node scripts/check-architecture.mjs` | PASS，29 package 边界 |

本阶段没有新的 RED，没有 CLI / 模型 / 视觉操作或 H/F 原 task 动作；未将 overlay 编译或纯拒绝测试计为当前 actual owned PID / Send / 取消 / 工具集合成功。B 编辑工具实际消费后的 compile / test / commit ACK 仍须由 B 发回，才能视为消费完成。

### 安全 checkpoint 与剩余限制

本阶段 SOURCE / 独立 REPORT 发布后，R 暂停写唯一消费文件 `internal/adapters/agents/relay_execution.go`，直到 B 通过编辑工具原样消费、普通 commit 并明确 ACK 归还。R-only metadata / 本报告不是 B 消费文件。精确内容差异或所有 R 逻辑冲突退回原 R，B 仅实现普通 text / bridge 的映射，不改此公共逻辑；T 不手改 Native / WIP。

最终 I 仍普通合 R / B 原 SOURCE 保留作者和 DAG；临时单文件消费不授权删除 / squash 原作者历史，也不等于最终集成或 ZIP 验收。先前旧 task 一次模型与逐事件核查、首 503 accepted + thread/read -32603、首 registry 三分钟 FAIL、旧 PID / 工具证据不足原样保留。当前旧任务 `reconcile` 权限门禁由模块向最高层询问，裁决前本阶段不改 Transition / revision / A-B-C / grant / ownership、不手写 DB、不由退出推断 done。

具体单文件 closure 已直接交 M-product / B / T / H 与父模块。raw 源 / credentials / 模型输出不进 HTTP、Git、包或 stdout；V4 / schema / engine / mapper / weights / schedule / 17321 冻结不动。实际模型和取消路径验证、H/F 真实操作窗口及最后 ZIP 仍由原 owner / 模块继续，本阶段只交付最小 Native 内部证据依赖。

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
