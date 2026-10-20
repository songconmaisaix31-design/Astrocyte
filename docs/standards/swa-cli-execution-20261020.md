# 首版 Swarm CLI：真实执行来源、抽取方案和测试计划

本报告对应[当前 Worker 任务](../../tasks/swa-cli-execution-20261020.md)。状态为 **SOURCE_REVIEW_COMPLETE / DESIGN_PROPOSED / TEST_PLAN_PENDING_REVIEW / IMPLEMENTATION_NOT_STARTED**。所有动态验证均为 `NOT_RUN`；现有源码中的能力与上游构建检查，均未转换为本机验收结果。

## 1. 范围、版本与证据强度

当前首版目标来自主控 `tasks/swa-cli-v1-20261020.md` 与 `docs/QUESTIONS.md@9140cf8aa618c832b3a099222a46991b7c2b3ddd`：抽取 Orca CLI 对应的执行能力，让用户从已有 Agent 或终端调用可控 Swarm CLI。MCP、完整旧 S3、科研和正式发布各自保留完成条件。META_PLAN §0、§3、§5—8 提供共同业务架构；依赖首版新选择的部分以当前任务为准。

| 对象 | 本次实际证据 | 可支持的结论 |
| --- | --- | --- |
| 产品基线 | `32ab1eec808a5848aabad40f11133735a1214e55`，本树源码读取 | 现有服务、权限与接缝的静态状态 |
| 治理接续 | main `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765` 的最新 AGENTS/CODING/SPEC 已只读核查 | 程序仍沿产品基线；实施前由主控安排普通合入 |
| 旧 M1 材料 | 旧分支 `80d4f8d1e323982ef73d880ce9927f4bc5c58340` 已普通 push | 材料和历史保留；动态检查未执行 |
| 安装的 Orca | `orca status --json` 返回 `appVersion=1.4.230`，runtime ready | 当前宿主可读；未验证本首版能力 |
| 本次 Worker | Codex CLI；终端信息为 `GPT-6.1-Sol xhigh fast` | 实际开发客户端与模型观察 |
| 上游 | `stablyai/orca` tag `v1.4.230`，准确 `b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`，本次重新核对 HEAD/干净树 | 下列来源分析绑定该版本 |
| 不存在的复用依据 | 原 Agent-Graph 无有效本地 HEAD，OrcaKernelLab 登记路径不存在 | 不把它们列作已有实现 |

源码根为 `C:/Users/DW/orca/Astrocyte/.verification/orca-source-v1.4.230`，只读。以下 `O:` 路径均相对该根；上游永久入口为[准确源码树](https://github.com/stablyai/orca/tree/b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e)。`S:` 路径相对本产品根，绑定上述产品基线。源码读取覆盖完整核心文件，较长文件分段连续读取；对支持模块仅引用实际展开的函数，不宣称已执行完整依赖图分析。

## 2. 完整执行链的真实边界

### 2.1 命令到 RPC

`O:src/cli/index.ts` 先通过 `args` 解析/验证命令与 flags，再由 `src/cli/dispatch.ts` 的 handler manifest 选择处理器。CLI 不持有工作 Agent 的执行进程。

| 用户命令 | 真实 RPC / 后续实现 | 不能从回执推导的结果 |
| --- | --- | --- |
| `orchestration worker-start` | `src/cli/handlers/orchestration/worker-launch-handler.ts` → `orchestration.workerStart` → `runtime/rpc/methods/orchestration/worker/workers.ts` | CLI 命令返回不自动证明任务完成 |
| `agent launch` | `src/cli/agent-launch-request.ts` → `agent.launchReplay`，携带 operationId；旧路径为 `agent.launch` | 创建表面、journaled、input accepted 各有不同证据 |
| `worker-show` / `worker-read` | `orchestration.workerShow` / `workerRead` → `worker-control.ts` | PTY 存活不能单独证明 Agent 存活或回合开始 |
| `worker-stop` | `orchestration.workerStop` → `worker-stop.ts` | context-only 停止不关闭终端；未知停止不等于退出 |
| `worker-release` | `orchestration.workerRelease` → `worker-release.ts` / `worker-release-completion.ts` | 已结算、归档和资源释放分别记录 |

`O:src/cli/runtime/client.ts` 协商能力、保存 mutation request ID、检查原宿主来源，并为可能交付的错误附恢复信息。`runtime/metadata.ts` 读取 runtime metadata 和 token；`runtime/transport.ts` 通过 Unix socket / Windows named pipe 发送带 token 的 JSON RPC，校验响应 ID 和 runtimeId。远端另走 WebSocket/配对。metadata 中的 authToken 属于凭据，不进入产品任务输出、日志、Git 或公开报告。

`agent.launchReplay` 对 operationId 与 fingerprint 做持久准入和回放；`agent.launch` 不携带 ID 的旧行为仍存在。`orchestration.workerStart` 要求当前 Run 的 coordinator 绑定身份，从普通终端直接包装该命令不能绕过此要求。能力缺失、旧协议、跨宿主和结果未知应直接返回真实拒绝。SWA 不照搬 CLI 的兼容重试与模式降级作为自己的执行策略。

### 2.2 Node 后台实际组装

完整读取的关键来源：

- [orcad-entry.ts](https://github.com/stablyai/orca/blob/b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e/src/main/orcad/orcad-entry.ts)：`startOrcad` 安装 Node host adapters、选择 userData，并通过 `startOrcadWithHost` 保持实例锁与清理顺序；214 行启动 daemon，237 行创建同一 `OrcaRuntimeService`，309 行注册 headless PTY，325 行恢复 orchestration authority，346 行创建 RPC server。
- `O:src/main/orcad/orcad-host-adapters.ts`：替代 AppEnvironment 和 secret store；这是真实 Node 适配，并非 Electron API 的空实现。
- `O:src/main/orcad/orcad-profile-state-startup.ts`：读取 profile、建立真实 profile SQLite、迁移并取得 storage authority。`register-headless-runtime.ts` 仍必须传入 Store，后续 terminal.create 才能持久化。
- `O:src/main/orcad/orcad-daemon-supervision.ts`：daemon 是独立、detached 终端持有者；停止 orcad 时仅 disconnect。daemon 不可用/降级时上游仍可服务进程内终端，这些新终端不具备相同的重启持久性。
- `O:src/main/orcad/orcad-lifecycle.ts`：先停止 RPC/运行资源，再处理 profile flush、host 清理和实例锁；SIGINT/SIGTERM 有有界关闭，SIGHUP 不结束 headless host。

`OrcaRuntimeService` 的最终类还继承多层 runtime 模块，拥有工作区、布局、账户、会话、Orchestration、来源检查和 UI 投影。Node 启动仍会组装 Agent hooks、account homes、session search、恢复、自动化和推送。纯 Node 可运行这一事实，不证明把该类整体复制就只抽取了执行能力。

### 2.3 启动与首输入

`worker/workers.ts` 校验 coordinator、Run、Task、prompt budget 和 placement。`local-worker-start.ts` 先解析配置 Agent 与模型，再 `createStartingWorkerDispatch` 持久记录 Task/Dispatch，随后经 `worker-start-agent-placement.ts` 调用共享 `src/main/agent-launch/agent-launch-executor.ts`。新工作区 setup、Agent readiness、终端/结构化会话创建、身份准备和 preamble delivery 各自保留阶段与 effects。

`worker-topology.ts` 的 `requireWorkerAuthority` 要求 paneKey + processIncarnation，结构化身份另有 session/fence。`deliver-worker-dispatch-preamble.ts` 通过 `sendAgentTurn` 或结构化发送投递真实任务。`worker-start-readiness-settlement.ts` 单独观察回合开始；可观察却未观察到时保留 `start_unknown` / `outcome_unknown` 及残留资源。源码还兼容部分无法证明回合开始的旧 provider，SWA 需如实公布该覆盖缺口。

`agent-launch-prompt-delivery.ts` 的 `journaled` 表示已写入 transcript，`handed-to-terminal` 表示交给终端，未确认的历史回放为 `unconfirmed`。它们都不能直接映射为业务 done。`agent-launch-executor.ts` 会根据宿主设置选择 structured/terminal，并可在明确拒绝后改变模式；SWA 获准的执行方式必须明确固定，不能借这一分支改变工具或权限边界。

### 2.4 进程、会话身份与状态

`worker-observation.ts` 按 Dispatch 的 paneKey、processIncarnation、host scope 检查当前对象；旧 handle 可在证明同一 incarnation 后重新取得。结果包括 `unattached / missing / identity_changed / live / exited / unverifiable`。`workerShow` 在 runtime epoch 改变后将未完成的 starting/stopping 记录为未知。只有匹配身份的观察才可归属于本次 Worker。

结构化 Codex 路径位于 `O:src/main/codex/`，具有真实 app-server、acquisition、thread open/resume、journal、turn dispatch、prompts 和 teardown。已完整读取 `codex-structured-session-adapter.ts`、`codex-structured-session-acquire.ts`、`codex-structured-session-close.ts`、`codex-structured-prompt-ownership.ts` 与 `codex-app-server-session.ts`。它们依赖 `native-chat/agent-session-wire` 接口、host event sink、launch resolver、账户来源、进程起始身份及 journal 翻译；全部不属于 `@orca/process-host` 独立包。

`managed-provider-process.ts` 经 `provider-process-supervisor.ts` 启动真实子进程。POSIX supervisor 持有专用进程组并在 owner 消失时执行有界终止；Windows 直接启动 provider，没有该 POSIX supervisor。`provider-process-close.ts` 分别返回 root 与 tree verdict，默认接受 root exit 的策略不能充当全部后代退出证明。SWA 接入须检查实际树状态，保留 `tree=null/unverifiable`。

### 2.5 输出与结果

`worker-control.ts` 先核身份再读输出，结束后再次核验。`worker-output.ts` 优先读取精确 provider transcript，将 cursor 绑定 Dispatch、源、进程与来源 fingerprint；来源变化返回 `source_changed`。上游 `source=auto` 可退到终端尾部，并显式标记不完整。SWA 的结果路径要求精确原生来源；来源缺失即报告缺口，终端尾部仅作为显式选择的诊断输出。

结构化低层 RPC 实际为 `agentSession.create / send / history / subscribe / cancel / close` 等，来源 `O:src/main/runtime/rpc/methods/structured-agent-session.ts`。`cancel` 针对回合；`close` 保留会话记录与 journal、处理 provider child。停止是否已经完成还须核 host 与 provider 的退出事实。公开 `terminal.close` 返回的 close 证明同样需要检查，不能只读 RPC 的 ok。

业务结果继续来自 SWA 的 Artifact、固定快照、检查和完成证据。Agent 自报 worker_done 或提供方 turn-completed 是执行观察；用户采用与 Git push/merge 均沿现有独立规则。

### 2.6 中止、退出、所有权与释放

`worker-stop.ts` 先 `beginWorkerStop`，再核 exact identity、owned resource 与原宿主；结构化和 PTY 分别关闭。无法确认关闭时写 `stop_unknown`，成功仅在真实退出/`ptyKilled` 证明后结算。context-only 的 stop 可仅终止 assignment，保留实际进程。

`worker-release-completion.ts` 重核 Worker 与 resource、实际 handle、release lease 和归档，提交归档后再次检查所有权，才关闭准确终端；未知/断联保留 `release_pending / release_unknown`，人工接管或身份无法证明时 retained。主控此次 Run 观察到的旧终端复用失败和启动初期 identity_changed 仅为真实环境补充信息，不作为本首版通过证据。

`O:src/packages/process-host/src/process-tree-kill-gate.ts` 的默认 gate 为 null，默认允许树终止；注册 gate 抛出也允许。`own-chromium-tree-kill-guard.ts` 在宿主层安装保护；包外 `provider-process-teardown.ts`、`windows-process-tree-kill.ts`、`pty-descendant-exit-verification.ts` 负责额外树与身份验证。仅抽取包导出名不能取得这些保护。

抽取时须在调用树终止前，以当前持有 child handle、原 PID/创建时间、host/session/operation/fence 和所属关系验证目标；无法证实时拒绝按 PID 遍历终止，仍允许对确实持有的进程句柄请求关闭，结果保留未知。单独注册一个会抛出的 gate 不构成 fail-closed：上游会捕获并允许。强制调用前验证须置于产品适配层，不能依赖诊断 hook。PID 回收、无关进程、人工接管、并发 stop/release 和根退出后仍存活后代均须真实测试。

### 2.7 接续与交接

上游结构化 adapter 保存 providerHandle/thread ID 与 acquisition generation；会话接续须重查该 Agent 的原生能力、账户来源、cwd 和原记录。PTY daemon 持有的是终端进程与尾部，不自动等于原生会话恢复。CLI handle 是 runtime scoped，也不能直接充当跨重启永久身份。

SWA 已有 Workspace 的 SessionBinding、ContextPacket 与原生接续/交接语义，以及 Swarm fencing。沿 SPEC §8：保存成果和上下文、阻止旧动作、确认原执行退出、释放原占用、建立新绑定与 fence，再输入。原执行停止未知时保持 blocked。`S:internal/workspace/app/relay_handoff.go` 的现有人工交接还要求同项目同作品、不同 Agent、原 Task released、原终端退出且无 pending action；不能直接包装成任意 Agent 自动交接。

## 3. Node 构建、依赖与可抽取范围

已完整阅读 `O:config/scripts/build-orcad-node.mjs`、`build-orcad.mjs` 与 `orcad-entry-build.mjs`。

`build-orcad-node.mjs` 先取得/核验对应目标的 node-pty prebuild，为本机目标准备固定 Node，再运行 `build-orcad.mjs`。后者分别打包 server、CLI、launcher 和 fork children，阶段化放置 node-pty、@parcel/watcher、Windows process-tree/CLI/security、emoji data、Linear SDK、ripgrep、native visuals 与 mobile web。`orcad-entry-build.mjs` 明确列出 daemon、watcher、profile writer/backup、foreign SQLite reader、session scanner、port scan 等子入口；这些 children 的真实文件位置和运行时解析不能遗漏。

构建脚本会检查全部 metafile 中对 Electron 和 `node:sqlite` 的直接 import，并执行 plain Node load、provider、daemon 和 worker smoke。该脚本的存在不等于本次运行通过。`src/main/sqlite/sync-database.ts` 使用 `process.getBuiltinModule('node:sqlite')` 动态加载，所以“零 node:sqlite import”也不表示无需 SQLite。root engines 为 Node 24；`node-runtime-pin.ts` 固定 Node 24.21.0，esbuild target=node18 只描述输出语法。完整宿主不可只按 process-host 的 Node >=18 engines 交付。

| 范围 | 现有模块性质 | 本次建议及必要适配 |
| --- | --- | --- |
| `src/packages/process-host` | 唯一已建立的独立 workspace Node 包；private，非已证明可从 npm 安装的公开产品；runtime 无 npm/native/Electron 依赖 | 按准确源码复用包与 public exports；保持单实例注册状态；保留 Windows 参数编码、限额和退出语义 |
| `src/main/provider-process/*` | Node 子进程管理模块，依赖包外后代核验、Windows 与诊断模块 | 与实际采用 provider 一并抽取必要闭包，适配所属进程身份与日志；root-only close 不映射为全树释放 |
| `src/main/codex/*` 与 `native-chat/agent-session-wire/*` | Node 可达的真实结构化 provider；尚无独立包边界 | 使用真实原生控制器和事件 sink；保留 thread/turn/acquisition；按所选 Agent 读取其完整必要依赖，不能用空 provider |
| `src/main/daemon/*`、`ipc/pty/*` | Node daemon + 原生 node-pty/ConPTY + Store/配置/终端状态 | 只有获准的 PTY 路径才纳入；保持 daemon 存活、socket/token、reconnect、owned close；daemon 降级即拒绝宣称持久执行 |
| `src/main/orcad/*` | 已有 Node host 组装 seam | 作为抽取宿主生命周期、路径、持久准备和打包边界的来源；不能原样调用后忽略其账户迁移、自动化、推送和浏览器安装行为 |
| `src/main/runtime/orchestration/*` 与 worker RPC | 完整业务协作库，含真实 Run/Task/Dispatch/消息/释放记录 | 不复制成 SWA 的第二任务服务；复用所需规则与执行事实，业务权威保留 Go |
| `src/cli/*` | 参数、能力协商、transport、回执展示 | 采用已证实的请求身份和未知语义；公共 `swa` 入口继续 Go 服务协议，不调用 shell 拼接字符串 |
| 桌面/布局/账户/remote/browser | 部分为 Node host 抽象，部分依赖 Electron、原生扩展或已有 profile | 仅为获准执行所需的配置读取；不复制用户 profile，不自动安装 hooks/MCP、不开放远端/配对/浏览器/视觉 |

MIT [LICENSE](https://github.com/stablyai/orca/blob/b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e/LICENSE) 已完整读取，版权 `Lovecast Inc.`。后续复制源码及有实质部分的发布产物必须保留完整许可证与归属，记录原仓/准确 Commit/原路径/本地修改。native/第三方依赖另带其许可；没有根据 root MIT 推断全部依赖同许可。本次没有复制程序或依赖。

## 4. 对照现有产品的唯一业务归属

| 对象 | 唯一拥有者 | 执行适配层可保存的内容 |
| --- | --- | --- |
| 项目、用户/Agent 权限、批准与当前 epoch | Workspace Go 服务及现有主 SQLite | 获准的固定引用和执行所需快照，不签发或扩大批准 |
| Mission / WorkItem / Lease / 预留 / Artifact | Swarm Go 服务及现有主 SQLite | 关联 ID、当前 fence；没有第二份任务状态机 |
| SessionBinding / ContextPacket / Native operation | Workspace 原服务及存储 | 原生 provider ID、controller、启动 operation、host、进程起始身份、观察事实 |
| Orca terminal / provider thread / launch request | 所属执行宿主 | 自身执行事实、幂等回执和 transcript；不决定 Mission、验收或预算 |
| 既有独立 Works Relay Task | 当前 Workspace RelayRuntime 与 relay store | 保留既有用户流程；新 CLI Mission 不自动复制一份 RelayTask |

`S:cmd/server/relay_swarm_assembly.go` 已组装同一主库、WorkspaceS2、Registry、object store、SwarmApprovalBridge、PreparedMemberEvidence、NativeBridge 和 send authorizer。`internal/swarm/app/commands.go` 已处理幂等、版本、批准、认领及资源限制；`lifecycle.go` 在停止/费用未知时保留原租约和写域。不能只复用其类型后跳过真实授权回调。

现有接缝有明确前提：`PreparedMemberInput` 必须来自原 Workspace 准备，含真实 controller/startupOperation/session/项目/root；Claim 仅接受绑定 Agent；首次输入前还要把真实 Lease 绑定回原 preparation。当前实现依赖 retained owned controller。接入 Orca 的新执行适配器需要在这些消费方端口提供同样可核实的证据，不能生成一份虚假的 RelayTask 来满足查找，也不能靠不同身份代发。适配器程序边界需在执行部署答案后定稿。

`S:internal/adapters/sqlite/relay_store.go` 明确独立 relay.sqlite 与其独占目录，不能称其与主库天然同一事务。本首版新任务状态放在现有 Swarm 主库；现有 Relay 独立流程保留自己的真实归属，不作为同步镜像。Orca 的 `tasks/dispatch_contexts` 也属于上游协作对象；选择低层 `agent.launchReplay` / `agentSession.*` / terminal 接口可避免为 SWA 业务再创建 Orca Task。

## 5. 推荐方案与未决部署影响

### 5.1 一个业务架构

`cmd/swa` 作为薄 CLI，向现有 `cmd/server` 后台发送版本化 HTTP；复用 OpenAPI、现有主体认证、幂等键、版本和安全错误。Go App 服务管理授权、任务、租约、上下文、产物、停止与对账；执行适配器只接收本次获准的会话和输入。执行成员从现有可执行工作项中选择并原子认领；外层 Agent 负责表达目标和传达结果，不另建中央 LLM 调度服务。

推荐源码抽取方向为独立的必要 Node 执行宿主：从 process-host、实际 provider、身份/终止模块及 orcad host seam 取得执行能力；由 Go 服务持有或连接，不加载 Orca 任务调度、产品资料库或默认自动化。执行宿主拥有真实 child/session，Go 保留唯一业务状态及既有操作收据。执行宿主只向 Go 返回有来源的事实，不能自主创建后继业务任务或取得新权限。

当前交付为待审核设计，抽取模块尚未实现。纯 Node 构建说明独立路径有真实上游基础；所选 provider 的打包闭包、Windows 后代退出能力、干净配置下的账户来源和工具关闭仍须本机真实验证。没有以结构化 provider 具备功能为由预先判定它符合当前工具限制。尚未取得构建图或逐 provider 完整依赖闭包；本报告列出的模块与调用边界不构成精确 vendor 文件清单或可构建保证。

### 5.2 用户未答的部署选择

| 用户选择 | 具体实现影响 | 必须验收的差异 |
| --- | --- | --- |
| 脱离 Orca 桌面独立运行 | SWA 发布必要 Node runtime/执行 bundle 与相应原生依赖；专用 SWA 数据目录和凭据引用；使用抽取的真实 provider/进程模块 | 未安装/未运行 Orca 的独立安装目录中完成真实链路；不读写原 Orca profile；后台重启和残留进程可核查 |
| 依赖已安装 Orca 的公开接口 | Go 适配器通过当前 CLI/RPC 的能力协商连接显式原宿主；优先低层 launch/session 操作；记录 runtimeId、版本、会话/终端身份与 launch operationId | 断联/旧版本/宿主变化明确拒绝；同一个请求不能切到其他 Orca；原桌面关闭、重启、人工接管的实际状态可见 |

依赖安装 Orca 时，`agent.launchReplay` 的创建+prompt 一体化不能绕过 SWA 的准备、Claim 与发送准入。必须先完成无输入会话准备、取得真实身份、绑定并认领，再使用有幂等保障的实际输入；公开接口不能证明原生工具关闭或退出时，此能力保持 unavailable。公开 API 对 SWA 所需精确退出/所有权证据是否完整仍待真实核实，不能默认 `agentSession.close {ok:true}` 已满足资源释放。

首版执行协议、Node/PTY 是否进入安装包、所选 Agent 能力集合和程序 write_paths 依赖用户答案，保持未冻结。两个部署选择都继续同一业务架构，不同时实现两套后台或自动互相降级。

### 5.3 Agent 与用户控制的合法路径

已核实 `S:internal/adapters/httpapi/session.go`、`relay_scoped_workspace.go`、`internal/workspace/app/project_policy.go`、`internal/swarm/app/commands.go` 和 `lifecycle.go`：

1. Agent 项目 token 可以经现有 scoped routes 准备/提交 Proposal、读 Mission，以及在相应 Grant 下管理工作项、认领、提交等；服务仍逐次重查当前 Grant、项目和能力。
2. Agent 不能创建 Mission、批准、采用、拒绝、暂停、取消或恢复 Mission。RelayRuntime 的读写还整体要求 human，不能直接作为 Agent 任务入口。
3. `ApproveProposal` 已保存 `proposal_approved` 事件；生产源码对 `CreateMission(` 的调用仅找到服务定义与 HTTP handler，尚未找到实际批准事件消费者。必须实现真实可信消费：只消费真实 human 批准事件，重读当前 Grant，按 event/grant 幂等创建 Mission。消费方身份由服务组装提供，HTTP/CLI 参数不能自称 consumer。
4. 新 CLI 的 Agent submit 可创建/提交 Proposal；批准未完成时返回 `approval_required` 与固定 proposal revision。已有有效批准继续原规则，无须重复询问。创建 Mission 的动作通过上述可信消费者完成。
5. Agent `task stop` 只请求停止自己持有、当前项目 Grant 允许 `stop` 的原执行会话，保留业务 Mission 与 Lease 的真实状态，交现有生命周期对账；human 的 `mission cancel` 继续现有入口。主控已确认这一划分符合现有权限语义，无需重复询问用户；新增聚合 API 契约仍交 Reviewer 审核。现有 native stop 权限不能被扩大为控制别人的 Worker。
6. 人类批准、取消 Mission、完成确认与采用沿既有用户控制面。终端用户先在该入口选择项目并获得原项目 scoped token，随后终端与 Agent 均用真实 Agent 主体调用已批准范围内的 CLI；已有有效批准直接继续执行。批准缺失时 CLI 返回固定 Proposal/revision 和现有用户批准入口的 required_action，批准后重用原任务/请求查询接续。此路径足以让终端用户执行及控制获准任务，无须设计额外认证系统；`--actor human` 不构成凭据，CLI 不代用户批准。

现有 `sessionGuard.bootstrap` 检查 Authorization 与 HTTP origin，然后签发 local-human cookie；它没有证明实际操作者为人的外部凭据。带 Agent bearer 的请求不能转换为 human，但拥有通用 shell 的调用者去掉 Authorization 后主动请求本机 bootstrap，不在该服务目前已证明的隔离范围内。因此 CLI 必须保留 Agent 身份、不调用此 bootstrap 冒充人，也不能把一项 bearer 拒绝测试表述为操作系统层的人机隔离已通过。该既有边界如实保留，本次不扩大为新认证系统任务。

工具/外发/预算的适用条件保持原决定。当前 Codex 0.162.0 受限路径在 `S:internal/adapters/agents/relay_native_tools.go` 立即拒绝，其公开 app-server 无已证实的 native file tools 排除能力。抽取 Orca 的 provider 不改变 Codex 注册工具。选择别的真实 Agent 也要逐项核当前任务能力与授权；不会自动换 Agent/模型/供应方。

可用于固定公开文本正向核查的本机候选具有不同的现有证据，不能从 Codex 拒绝推导全部 Agent 不可执行：

| 候选 | 本仓真实来源与已有观察 | 本首版仍需确认的条件 |
| --- | --- | --- |
| Claude Code | `S:internal/adapters/agents/native.go` 已用 stream-json、`--safe-mode --tools=`、空 strict MCP、显式 session UUID/resume；`docs/probes/local-agents.md` 的 W2 记录 2.1.238 两次真实公开文本 turn、同 UUID 接续和 owned stop | 历史结果未重新动态核；通用 selected-text 的调用前当前 model/实际供应端点仍缺证据，不能沿用旧输出猜测；抽取 Orca Claude SDK 路径也不能默认具有同一关闭配置 |
| Pi | 同一 Go adapter 使用 `--mode rpc --no-tools --no-extensions --no-skills --no-prompt-templates`，会话文件限定自有目录；已有本机 1.0.1 握手，历史模型调用明确 authentication_rejected | 当前正常配置/认证未核，不能重发原失败或未知操作；上游 `O:src/main/pi/rpc-launch.ts` 仅接受 fullAccess，`rpc-launch-resolution.ts` 另要求版本与 Windows 创建时间证明，需明确适配受限参数后再验证 |
| OpenCode | 本仓 loopback driver 与 W2 记录已有同 session、输出、停止及新 session 交接的真实观察 | 所选上下文的 managed/wellknown instructions/MCP 边界仍有缺口，不能因 permission deny 就宣称完整工具集合及外发已知 |

这些为源码和本仓历史证据支持的候选；本次没有启动它们或确认当前登录。推荐在用户明确选择的现有配置内优先核 Claude 固定文本路径，Pi 正常配置后也可核其受限 RPC；具体模型/供应方和外发许可不由自动选择改变。若当前批准不允许相应客户端，就返回具体缺失条件并保留其他候选能力事实。

### 5.4 可审阅的用户调用路径

以下为建议的用户命令意图，最终参数与共享契约在待决项处理后定稿；本次没有实现这些命令。

| 命令意图 | 用户可观察结果 / 共同服务 |
| --- | --- |
| `swa capabilities --json` | 软件支持、当前连接、Agent/版本、start/send/stop/resume 与证据缺口分别显示 |
| `swa task submit` | 带固定 project、goal、scope、Agent 和稳定幂等键；返回已有 Proposal/Mission/WorkItem 标识或当前所缺批准 |
| `swa task status` / `list` | 查询同一任务、依赖/权限/资源阻塞、原会话与当前执行身份；默认不启动 provider |
| `swa task result` | 结果摘要、Artifact/diff/检查引用、完成通道、已知用量与 unknown；分段读取精确输出 |
| `swa task stop` | 请求停止有权控制的原执行，返回 request/stop/exit/release 各自事实；不声称 Mission 已取消 |
| `swa task resume` | 核原生 session、账户/项目、原退出、当前 Grant 与占用后原生接续；不支持时明确拒绝 |
| `swa task handoff` | 固定 ContextPacket、新目标会话和新 fence；原执行未确认停止时 blocked |
| `swa task release` | 已结算执行的归档、进程退出和占用核对；保留原工作树/产物，不删除用户文件 |
| `swa mission cancel` | 只走既有 human 权限；Agent 返回明确拒绝和用户动作入口 |
| `swa serve` / 诊断 | 复用现有后台生命周期，显示已选择的执行部署方式与缺失条件；不隐式安装或更新 Agent |

CLI 的 task 视图使用原 Mission ID，工作项使用原 WorkItem ID；尚待批准的提交返回原 Proposal ID/revision。针对执行者的操作再明确 WorkItem、binding/session 和当前 operation，控制前由后台解析精确所有权；不从列表位置或终端标题选择执行者。一次任务可以包含多个独立工作项/执行成员。用户对每个成员分别查询状态、取回输出、停止、接续或交接；停止成员 A 时，成员 B 的 session/operation/fence 和输出不改变。只读任务汇总与对单个执行者的控制使用同一 Go 状态和上述身份关系，不新建 Worker 任务账本。

默认给终端简洁文本，`--json` 为稳定 envelope；stdout 只写结果，诊断写 stderr。成功查询的退出码与查询到的失败任务状态分别表达；拒绝/参数错误/未知效果使用非零退出码并附 `request_id`、原 operation 和 required_action，禁止提示不带原身份的盲目重试。新错误名与码只在契约审核后加入；优先复用 `scope_denied / approval_required / approval_revoked / unsupported_capability / provider_unavailable / delivery_unknown / context_stale / version_conflict / lease_lost / budget_exhausted / evidence_missing`。

### 5.5 状态、重复与后台故障

- 所有写请求在客户端调用前确定稳定幂等身份；同键同内容回放既有结果，同键不同内容冲突。业务 receipt 与 provider operation 分别关联，避免将 HTTP request ID 当原生回合 ID。
- 事务先保存待执行动作与当前绑定，再对原宿主执行 I/O，最后保存观察；不用数据库事务包住长时间模型调用。任何“外部已执行、结果尚未持久”的窗口保留原身份并对账，不重新创建 Agent。
- 用户 CLI 断开只终止等待。Go 后台和执行宿主继续持有原动作；新 CLI 使用同一任务与 operation 查询。
- Go/执行宿主重启时，从当前持有者和真实 native/provider 记录核对身份；structured 子进程不能无依据重连，daemon PTY 持久性也不能推广到 structured provider。实际执行无法证实则标未知、保留预留与写域；恢复仅在原退出确认和现有授权条件满足后执行。
- 任一新受保护动作重查 Grant epoch/expiry/撤销、binding revision、fence、预算、来源版本与工具能力。旧执行者即使晚到完成回报也不能覆盖当前任务或提交权。
- 取消请求、原生回合结束、根进程退出、后代退出、用量结算、占用释放分别记录。费用未知不填零；需用量才能释放的现有路径保持其规则。

### 5.6 资源与安装产物

产品的工作项、尝试、截止时间、调用/费用与无进展限制由现有批准与 Swarm 规则处理；项目活跃上限沿当前构造参数，当前默认两名。此次开发验证仍遵守全局唯一重操作窗口，不能以产品允许两个执行者为由并行启动两组检查。需要两成员的同一场景在一项受管验证中进行，其他构建/测试/服务操作串行。

具体预算适用点分开核验：CLI 查询/结果读取不提交模型输入；独立会话执行由当前 ProjectGrant 的允许动作、模型/外发、期限与批准输入约束。本报告推荐的 Mission/Claim 路径直接依赖真实 reservation、`ExecutionDescriptor.Known` 和现有 S3 预算/工具证据门禁；QUESTIONS 的两项正式答复前要求为 Claim 拒绝、Known=false，因此当前该方案的正向认领与执行保持阻塞。`domain/mission.go` 的 `Budget.Validate` 还明确拒绝 calls/tokens/money 全 nil，不能填空预算绕过；`usage.go` 对缺证据的受限维度不允许结算释放。停止已经持有的原执行仍允许按当前安全停止规则请求，不以费用未知阻止关闭，但释放是否允许须遵守原预留规则。完整旧 S3 的成果采用、经验与科研不加入本首版验收；当前选用接缝所需的准入条件也不以首版名义绕过。

已有独立 Relay 的 CreateIntent/Prepare/Confirm/ObserveOwned 路径已获其原范围授权，明确不调用 Claim、不建立 Mission/WorkItem/Member 或 S3 预留；当前生产 Relay human guard 及实际 Agent 工具/资料范围仍逐项适用。其人工正向范围可沿原任务继续，当前首版 Agent 写入口不能借此自动开放。该独立路径不作为本推荐 Mission 执行的隐式降级，也不复制新的任务账本。主控取得相关答案后再决定推荐方案能否进入完整正向验证，当前可以继续来源/契约设计、实际可用范围的计划与真实拒绝场景。

部署答案决定附带 runtime 集合；共同产物为 `swa` Go CLI、同源后台、对应版本契约/用户使用说明/许可证与已验安装说明。独立执行方式另含必要 Node bundle、固定 runtime 和实际采用的 native/PTY 资源；依赖方式明确最低能力、安装的 Orca 版本及连接要求。安装不扫描/复制原 profile、凭据或用户会话；配置仅写 SWA 拥有的目录，冲突立即拒绝，重复安装不重复注册。

不先把上游完整 `build:orcad` 当作最小产品构建：它会下载/编译、使用系统临时目录、打包额外功能并运行 smoke。后续构建入口需按所选闭包保留必要 load/实际执行验证，将可控中间目录放在本仓已忽略目录；缺依赖立即失败。本次没有执行这些脚本。

## 6. 从需求推导的测试计划

### 6.1 前提与真实样例

全部用例当前 `NOT_RUN`。实施前由独立 Reviewer 按当前首版需求、SPEC §6—8/§10、META_PLAN §5—8 自行推导条件，再审核本计划；本报告不代替接受决定。部署答案或权限范围改变时先修订受影响用例。

真实模型场景使用本项目 Agent 创建的公开 Git 验证项目，放在 `.verification/swa-cli-execution-20261020/` 的隔离工作树；登记实际 repo URL、准确 HEAD、当前项目 ID、授权输入版本与成功条件，不读取用户作品或无关会话。不得把本机工作目录当已确认的公开项目。若尚无满足要求的现成固定项目，由主控按原公开验证范围指定/安排创建后再执行；样例身份未知时不能登记通过。

固定目标为：从明确给出的公开 README/许可证段落完成三项事实核查，输出包含原文位置的结果；由另一个真实会话依据固定 ContextPacket 检查遗漏并产出最终结果。所需输入以批准的固定文本提供，原生文件/命令工具保持关闭。结果由测试者按固定事实与来源核对，不用模型自评分作判据。若首版还必须在工作树修改文件，原工具限制与该验收冲突由主控提交用户决定，未答时这部分不执行、不缩减为展示数据通过。

原生 Agent 使用部署决策下已安装并实际配置的客户端。记录实际 executable/version/model、配置来源及允许工具/外发证据；Codex 当前限制下只能验证真实准入拒绝，不能用本开发终端拥有工具权限来替代产品受限运行时。原生接续用同一真实 provider session；交接用另一真实 session，按批准范围选择 Agent，不伪造 provider ID。

所有持久测试使用真实文件 SQLite，集成使用实际 HTTP socket，执行测试使用实际子进程。基础进程测试可用真实 Node/OS 子进程做等待、输出和创建后代等行为，它们只测试进程生命周期，不扮演 Agent CLI。禁止 stub provider、伪造 API、虚构费用或静态填充成功事件。

### 6.2 覆盖表

| 编号 / 层次 | 需求与场景 | 实际操作与判定 |
| --- | --- | --- |
| U1 单元 | 参数/输入在执行前拒绝 | 用真实 parser/Schema 检查缺 project、越界路径、非法幂等键、过长输入、未知命令；观察无执行调用；有效参数保持原意 |
| U2 单元 | 身份、租约与结果语义 | 使用纯领域函数校验完整绑定、版本/fence、过期与未知结果映射；不得从 input accepted 推导 started/done，不将 tree=null 当 exited |
| U3 单元 | 输出游标 | 对实际编码器检查原 task/source/version 绑定、合法边界和来源变化拒绝；不复制实现内部的每个分支作为独立用例 |
| I1 集成 | 唯一任务与真实授权 | 真实 SQLite/HTTP：Agent 提交 Proposal，真实 human 批准后可信消费者创建同一 Mission；重复事件/请求不重复；撤销发生于消费前时拒绝，无第二 Orca Task/RelayTask |
| I2 集成 | 从 Agent 调用不冒充 human | 使用真实项目 token 调用允许读写；跨项目、撤权/过期、伪造 actor、直接 Create/CancelMission、带 bearer 请求 human bootstrap 均被拒绝；产品 CLI 不请求人类 cookie，数据库无越权状态；§5.3 所列 headerless 本机 bootstrap 限制单独保留 |
| I3 集成 | 已有批准与受控输入 | 原生空输入准备→真实身份绑定→实际 Claim/预留→原 preparation 绑定 Lease→首输入；对失效 binding/Grant/fence/能力缺失逐一拒绝；输入前失败无模型提交 |
| I4 集成 | 稳定重复及断开 | 并发发送同键同请求，另发同键改正文；实际 HTTP 客户端在响应前断开，再用原键查询/回放；只存在一次业务动作和一次对应 provider 输入，未知禁止重发 |
| I5 集成 | 生命周期与原件 | 查询每个本轮 event_id 在真实事件库的所属/时间/操作；核 provider 自产 session 文件、原生 ID、SQLite 绑定和结果引用；文件与事件缺失保持 evidence_missing |
| I6 集成 | 持久存储失败 | 专用实际 SQLite 在只读/不可写目录等真实条件下提交动作，立即失败；启动/发送前写失败不能执行，执行后结果写失败保留 unknown/可对账身份，不能重发 |
| I7 集成 | 上限与冲突 | 真实库的项目活跃上限、重叠写域、未就绪依赖、超时/无进展/预算耗尽；拒绝新认领并保持原预留，固定目标已完成后不继续派发 |
| P1 进程 | 正常执行/后代退出 | 实际已持有子进程与其后代，记录 PID/创建时间；正常退出与停止分别观察 root/tree，stdout/stderr 超限明确截断并有界返回 |
| P2 进程 | 所有权与 PID 重用防护 | 使用另一个真实无关进程的身份、旧 incarnation/已退出句柄发起控制；必须拒绝且无关进程继续；不要求强行制造操作系统复用特定 PID，不能发生的竞争保留未覆盖 |
| P3 进程 | 停止未知与残留 | 在本项专用宿主已持有真实执行后，终止测试自己持有的控制连接/宿主，保存中断前后的身份和实际进程表；不能确认退出时返回 unknown、保留占用，拒绝替代执行；恢复观察后按真实证据结算 |
| P4 进程 | stop/完成/release 竞争 | 对同一实际运行任务发起重复 stop，及在回合结束时发起 stop/release；持久请求合并或回放，归档保留；不能重复释放、覆盖完成通道或终止新所有者 |
| E1 端到端 | 终端用户路径 | 从实际安装目录运行 capabilities、submit、status、result；完成上述固定公开事实核查，文字/JSON 对应同一真实任务；不依赖开发树中的隐含 PATH/模块 |
| E2 端到端 | 已有 Agent 用户路径 | 原有真实 Agent 在获准项目中用其命令工具调用安装后的 swa，经项目 token 查询/提交/停止；产品内部执行者是实际子进程；没有传递 human cookie 或全局凭据 |
| E3 端到端 | 客户端断开后继续 | 提交后退出原终端/断开调用端，保持产品后台；另一实际终端按原任务查询并取回最终结果，无第二次输入；不终止用户或主控终端 |
| E4 端到端 | 后台重启与未知恢复 | 只对本项专用后台做停止/重启，保留原 SQLite；核实际 process/session 是否仍在，精确重新绑定或诚实 unknown；未知不启动替代，不把 PID 不在列表当必然退出 |
| E5 端到端 | 原生接续 | 原进程真实退出/占用释放后，对原 provider session 发送第二个固定输入；证明 native ID 连续、项目/账户一致、ContextPacket 版本和结果可回查；不支持时记失败/未验，不能登记接续通过 |
| E6 端到端 | 上下文交接及旧执行者 | 原会话保全结果、实际停止、释放；第二会话读取获准 ContextPacket 接续核查；旧 fence 提交被拒绝，原停止未知时新写入被拒绝 |
| E7 端到端 | 安装与依赖 | 独立方式在无桌面依赖的隔离安装目录启动；依赖方式核真实 Orca 的 capability/runtimeId 并覆盖断联/版本不支持；缺 Node/native/Agent 直接失败，不自动改执行方式 |
| E8 端到端 | 准入与脱敏 | 当前 Codex 文件工具无法关闭时真实拒绝启动；日志/JSON/错误不泄漏 token、账户文件、无关路径或原生正文；私有资料和未经批准外发不能进入上下文 |
| E9 端到端 | 多执行者精确控制 | 在一个受管服务/窗口内建立两个有真实原生进程的独立工作项 A/B，模型输入串行；分别用原有 Agent 和普通终端列出、查询、读取 A/B，核各自自产输出。停止 A 并核 A 的实际退出/占用后，B 仍为原 PID/创建时间/session/operation/fence，历史输出完整，且能接受获准的后续固定输入并返回真实结果；再清理 B。身份不得仅靠两个界面行，资源不足时记录本项未验 |

P3 的故障必须来自真实连接或本项专用进程失联。若故障发生前没有证明动作已准入，记录该次未覆盖“停止中断”窗口，不能伪造回执补齐。没有权限制造特定宿主故障、缺少能力或资源窗口时保持 `NOT_RUN / INCONCLUSIVE`，不得改成 mock 或绕过门禁。残留清理只操作本项证明拥有的进程；所属不明的进程交主控处理。

### 6.3 检查命令与证据要求

未来程序写域批准后，使用仓库既有 Go 测试入口检查受影响的 workspace/swarm/agents/sqlite/httpapi 包，以及 `cmd/swa` / `cmd/server` 的新增真实路径；Node 部分运行实际抽取包和宿主的相关测试。命令范围与新文件对应，不执行旧替身测试来充当本首版证据。既有 `scripts/check-architecture.mjs` 的适用检查、公共契约及生成一致性需要实际通过；M1 原计划已接受但尚未动态执行，若相关契约命令依旧崩溃，保留该失败并由原负责人处理。

当前尚未存在 `cmd/swa`、抽取 bundle 或对应测试入口，不能给出假装已可执行的命令。实施时将具体 test name/脚本和 CLI 参数登记到本任务，Reviewer 核对其与本表的映射后执行。每项保存命令、源码 SHA、配置与主体、实际退出码、UTC、两次资源采样、真实 SQLite/HTTP/进程/会话引用、预期与实际差异；敏感原件留在已忽略的验证目录。测试报告逐个解释通过、拒绝、未知和未执行的覆盖范围。

资源参数沿 [RESOURCES](../RESOURCES.md)：唯一窗口、BelowNormal；CPU 两次均低于 85%、可用内存均至少 4 GiB，采样单调间隔 2—5 秒；Go `GOMAXPROCS=4 / -p=4 / GOMEMLIMIT=512MiB`，Node heap 不超过 1024 MiB。来源/文档静态检查不占用动态窗口，按其单独范围执行。当前服务不能由本 Worker 擅自停用来取得窗口。

完成必须同时具有真实用户调用、多个执行者的分别观察与精确停止、结果核对、停止/退出/释放、接续/交接及关键失败证据。模型输入串行不免除两个原生进程同时存活时的资源限制，取得当前额度后才运行 E9。全部模型调用、受限工具、资料准入和停止未知门禁按当前决定处理；必要正向条件未满足时首版保持未验。

## 7. 主控需处理的明确差异与 Handoff

| 事项 | 原依据 / 需要的决定或工作 |
| --- | --- |
| 桌面独立性 | QUESTIONS 首版条目已询问未答；决定独立 Node 产物或安装 Orca 连接方式，不能代选 |
| Agent 控制与 human 边界 | SPEC §7.4，project_policy 与 Relay human guard；主控已确认 §5.3 的自己获准会话停止/人工 Mission cancel 划分，新 API 交独立审核；不重复请求既有权限授权 |
| 工具、调用预算与资料 | QUESTIONS 受限 Codex/S3 两项待答；固定公开核查与工作树写入需要的工具不同，选择具体正向执行范围前先核适用授权 |
| 可信批准事件消费者 | 当前只有批准写事件与人工 HTTP CreateMission；需实际补齐现有规则所需生产消费者，防重复、撤权及崩溃恢复，不能写成已经接通 |
| 执行接缝与写域 | Workspace 原准备和 Swarm Claim 要真实 controller/member 证据；部署明确后定稿端口和程序 write_paths，不跨写本次文档范围 |
| 测试对象与资源窗口 | 固定 Agent 创建公开项目、真实 Agent/配置、唯一动态窗口由主控协调；未经审核和窗口不执行 |

本次修改只限任务与报告，未修改 SPEC/QUESTIONS/契约/程序、M0 治理材料、原服务、数据或上游来源。建议主控将适用决策写入其拥有的事实源，并把准确报告 Commit 交独立 Reviewer。测试计划接受后仍须取得依赖答案和程序写域；接受调查不代表程序实现、首版验收或发布已经完成。
