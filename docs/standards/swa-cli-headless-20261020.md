# Swarm CLI隔离后台、身份、控制规则与真实测试计划

对应[Worker任务](../../tasks/swa-cli-headless-20261020.md)，接管基线main `e149510762f89bed06e7c8c766917554a58294e4`，后续普通快进至4e6e及`f7c97d36af03309868269a7ee3d8fa791a570974`。准确396e方案与T00—T15已获原Reviewer独立接受，主控已授予§5.3逐文件写域，当前正在实现。两条依赖命令实际EXIT0；产品构建、模型、安装和功能验收仍未运行。

## 1. 来源与唯一状态归属

`O:` 表示只读上游 `C:/Users/DW/orca/Astrocyte/.verification/orca-source-v1.4.230`，准确 `b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`、tag `v1.4.230`；本次HEAD和干净状态均已核。`P:` 表示已安装 `C:/Users/DW/AppData/Roaming/npm/node_modules/@earendil-works/pi-coding-agent`，package版本 `1.0.1`。`S:` 表示本产品。来源源码描述可支持静态判断；动态能力仍须按§7验证。

最新用户要求采用产品自带且隔离的后台，保全本机Orca。主控在本次 `ask` 回复批准本方案的状态边界：**私有Orca原生Run/Task/Dispatch拥有首版执行状态**；Workspace复用适用项目、授权与固定上下文；旧Mission/Claim/Relay保留原归属，同一首版任务不双写。主控在准确方案审核后更新SPEC，当前实现未变。

| 对象 | 唯一拥有者与产品边界 |
| --- | --- |
| 一次用户任务 | 私有Orca中一个Run及一个根Task；根Task没有Worker Dispatch，A/B为该根的两个子Task，各自具有真实Dispatch。汇总视图读取原生记录，不生成第二Task |
| Worker生命周期、消息、请求回执、终端资源、输出归档 | 原生 `orchestration.db`；沿原表、事务和生命周期方法，不另建Attempt、调度器或任务完成账本 |
| 项目、权限、批准、固定上下文 | 既有Workspace服务及其主SQLite；首版只保存指向原生Run/根Task的关联，不复制执行状态 |
| provider会话与进程 | 私有runtime的真实session record/journal、Pi自产会话、实际ChildProcess/daemon；这些是执行观察来源 |
| 已有认知、网页、Mission/Claim/Relay | 保持既有产品及其独立任务路径。本首版任务不调用Claim、不创建同义Mission或RelayTask；相应旧准入继续生效 |

Go/TypeScript/SQLite继续保留。Go CLI及已有HTTP/App负责协议和Workspace准入；抽取的Node runtime负责原生执行。跨进程边界来自复用真实Node执行宿主的必要条件；没有拆分既有Go模块。

## 2. 已读取的真实调用链

### 2.1 CLI、Node组装与传输

已完整读取 `O:src/cli/index.ts`、`handlers/core.ts`、`specs/serve.ts`、`runtime/launch.ts`、`runtime/serve-orcad-launch.ts`、`runtime/client.ts`、`runtime/metadata.ts`、`runtime/transport.ts`。

1. `index.ts`先解析并验证参数，`handlers/core.ts:serve`调用`serveOrcaApp`。`runtime/launch.ts`选择实际宿主；`resolveLocalServeRuntime`失败时原实现可以选择Electron。产品固定使用自带Node入口，选择失败立即退出，不调用该兼容选择分支。
2. `serveWithOrcad`向子进程传入`ORCA_USER_DATA`和实际入口；`orcad-entry.ts:startOrcad`安装Node AppEnvironment，设置`ORCA_USER_DATA_PATH`并取得实例锁。CLI读取后者，Node宿主读取前者，两者必须由产品启动器显式指向同一私有根。
3. 已完整读取`orcad-entry.ts`、`orcad-host-adapters.ts`、`orcad-profile-state-startup.ts`、`orcad-app-paths.ts`、`orcad-daemon-supervision.ts`。真实Node组装仍包括Store/SQLite、agent hook server、daemon、账户服务、structured session host、恢复和RPC；不能用process-host包代替完整组装。
4. `createOrcadProfileStateStartup`经`initOrcaProfilePaths/ensureActiveOrcaProfile`打开`profiles/<profileId>/profile-state.db`及对应authority，真实Store必须传给`registerHeadlessPtyRuntime`。Orchestration另由`workspaces/persisted-surfaces/automation-operations.ts:orchestrationDbPath`定位`<userData>/orchestration.db`；两库分别拥有配置与执行记录，不声称跨库原子事务。
5. `RuntimeClient`读取私有`orca-runtime.json`、协商contract capability并发送RPC。`transport.ts`使用原生named pipe/Unix socket，校验响应ID及runtimeId；mutation request ID与传输response ID分别保留。远端配对/环境选择在本首版入口不启用。
6. 已完整读取`runtime-rpc-state.ts`、`runtime-rpc-request-admission.ts`、`runtime-rpc-socket-metadata.ts`。runtime每次生成24字节随机owner token；metadata记录runtimeId、PID、startedAt和endpoint；Windows endpoint含PID与runtime后缀。token是本地runtime凭据，不能输出给Worker、写进Git或混称human批准。产品RPC只注册本首版必要方法，`enableWebSocket=false`；现有Go HTTP固定loopback与显式端口，冲突即失败。

### 2.2 Run/Worker、输入、观察、停止及释放

本节表中`worker/`相对`O:src/main/runtime/rpc/methods/orchestration/`；`runs/`、`messaging/`同属该目录，`db/`相对`O:src/main/runtime/orchestration/`。

| 行为 | 准确来源与实际语义 |
| --- | --- |
| Run/Task | `rpc/methods/orchestration/runs/runs.ts`、`runs/run-scope.ts`、`messaging/message-methods.ts`；Run绑定真实pane或Orca session，Task变更要求当前consumer |
| Worker创建 | `worker/local-worker-start.ts`经`createStartingWorkerDispatch`持久化后执行placement、readiness、`prepareStartingWorkerAuthority`和首输入；子Task/Dispatch即为实际成员身份 |
| 实际放置 | `worker/worker-topology.ts`和`worker-start-agent-placement.ts`；structured路径为`orchestration-structured-worker-session.ts:createStructuredWorkerSession`，创建真实session record后才取得host/fence |
| 首提示 | `deliver-worker-dispatch-preamble.ts`、`worker-start-readiness-settlement.ts`；native `sendAgentTurn`区分accepted、pending、rejected及unknown，未观察到turn保持`start_unknown/outcome_unknown` |
| 身份观察 | `worker/worker-observation.ts`同时核Dispatch、pane/process incarnation、host和resource；PTY live、Agent fleet liveness、结构化session是否可寻址分别输出 |
| 输出 | `worker-control.ts:workerRead`及`worker-output.ts`读取前后重核身份；structured journal及原provider transcript是成果来源。`source=auto`的终端尾部退路不用于产品结果 |
| 停止 | `worker-stop.ts`先beginWorkerStop，再核exact identity和owned resource，执行structured close或精确terminal close；未知保持`stop_unknown`，context-only停止不能证明进程退出 |
| 释放 | `worker-release.ts`、`worker-release-completion.ts`先核身份、冻结输出、持久归档，重核release lease后关闭并结算；保留`retained/release_pending/release_unknown` |
| 完成 | `db/tasks/task-status-transition.ts`明确禁止活跃supervised Worker下把Task改completed。产品ToolsOff Worker不发送CLI worker_done：控制端取得真实settled输出后显式停止并证明退出，再以自己的当前consumer调用原生taskUpdate；记录完成来源为控制端确认及真实结果引用 |

`workerStop`、provider turn结束、根进程退出、全部后代退出、归档和释放是不同事实。产品公开输出保留这些字段；停止A后不能因根退出就释放仍有后代的资源。未确认退出的原Worker不启动替代者。原生释放代码中的重试/恢复也只能对同一已记录动作作观察；产品入口不自动重发未知输入。

### 2.3 capability创建、恢复和校验的准确区别

已完整读取`orchestration-session-caller.ts`、`orchestration-caller-identity.ts`、`terminal-authority/verify-orchestration-compatibility-caller.ts`、`agent-session-claim-identity.ts`、`db/worker-dispatch/worker-dispatch-authority.ts`、`rpc-caller-scope.ts`、`orchestration-mutation-executor.ts`和contract fence。

- owner token认证本地RPC持有者；本地owner scope并不证明操作者是human。产品不调用headerless bootstrap，不签发human cookie，不传入consumer自称或`--actor human`。
- Run consumer通过真实pane/Orca session和`consumer_generation`绑定。Worker准备在事务中保存真实assignee、pane、process incarnation、host、launch-token commitment并推进generation；停止、结算及接管使旧绑定失效。
- PTY的真实launch token由宿主产生并注入所属子进程。`verifyOrchestrationCompatibilityCaller`结合token、host、process incarnation和实际hook attestation；`currentRuntimeLaunchSufficient=true`只在精确当前runtime的真实launch上允许免hook核验。恢复路径要求原有恢复证明，不能把token字符串直接当恢复通过。
- structured session的HMAC claim使用私有根的`agent-session-authority.key`，绑定authority domain、machine/principal/container/providerRoot、原生session及worktree scope；原key不复制，损坏拒绝。它与普通`ORCA_AGENT_SESSION_ID`有别；源码明确后者是同用户下的定位信息，本身不构成凭据。
- 当前源码保留`orchestrationCapability`传输字段、`capability_hash/capability_revoked_at`兼容列；本次源码搜索没有找到当前Worker启动为该header签发并校验新capability的路径。实际链路使用上述launch/session/consumer证明。不能把遗留字段写成现有安全令牌系统。
- 上游`@orca/process-host/src/process-tree-kill-gate.ts`默认允许、gate抛出仍允许；产品在树终止调用前执行fail-fast所有权检查。单独注册会抛异常的gate无法承担准入保护。

## 3. 产品隔离与控制端身份方案

### 3.1 安装、状态和全局消费者

已审Windows安装布局为`<InstallRoot>/bin/swa.exe`、`<InstallRoot>/lib/swa-runtime/`和`<InstallRoot>/runtimes/node-<原生版本标识>/node.exe`；可写数据显式使用`<DataRoot>`，不得位于本机Orca根、其父目录、账号根或现有项目根。安装器使用真实路径、reparse point及ACL核查拒绝重合/嵌套共享，不以路径字符串不同推断隔离。

| 原生消费者 | 当前源码行为 | 产品composition处理 |
| --- | --- | --- |
| `orcad-app-paths.ts` | 未设置时回到home/.orca | 必须显式私有根，禁止默认home分支；CLI和宿主同根 |
| `profile-index-store.ts`及`profile-storage-paths.ts` | profile index、默认profile与legacy迁移均按userData | 只在新私有根初始化，不读入本机index或legacy数据 |
| `orcad-plaintext-credential-migration.ts` | `orcadCredentialRoot()`固定home/.orca，无视私有profile；各integration store还读取Linear/Jira/Bitbucket/speech等凭据 | 不组装该迁移及这些integration入口；secret store如需初始化，keyDir显式私有根 |
| `orcad-managed-agent-hooks.ts` | `startCodexHooks`、`adoptLoginShellAgentHomes`、startup hook reconcile、OpenCode plugin refresh可能接触全局Agent homes | 不启动全局reconcile。Worker采用structured Pi native events提供真实状态；控制端使用真实当前launch证明。必要hook服务只监听私有endpoint，禁止改写系统Agent配置 |
| `orcad-account-services.ts`及codex account homes | 准备/迁移运行home、账户选择、认证和session resume | 不启动账户管理/迁移。仅显式只读访问获准的现有Codex认证来源，使用§4只读CredentialStore；不把整个home复制成可写账户 |
| `orcad-cli-launcher.ts` | 在profile下生成orca launcher；进程可将该目录放到child PATH前 | 产品入口名固定swa，直接调用随包Node和bundle绝对路径；原生控制helper使用产品内RuntimeClient。无orca同名shim，不改进程外PATH |
| daemon与history | `daemon-launch-paths.ts`按AppEnvironment解析daemon/history；daemon detached、runtime关闭只disconnect | 所有路径归私有根；只接管经PID/创建时间、token、协议及安装根核实的本产品daemon，未知拒绝；不关闭本机Orca daemon |
| runtime metadata / instance lock / signer | userData下metadata、`orcad.lock`、HMAC key；Windows runtime pipe含PID | 保留真实锁与随机runtime身份，核创建时间；原Orca metadata不读作产品目标，不扫除其他根的锁/socket |
| automation/session search/push/state-rule更新 | 原orcad启动会主动安装这些服务，部分涉及网络和会话发现 | 不在首版composition启动。结构化journal、已有特定会话读写继续可用，无全局会话扫描 |
| browser/desktop/mobile/remote | 原runtime具备相关类型、注册和资源 | 产品方法表不注册相应入口，不安装visuals插件或浏览器资源；保留原上游源码的必要静态依赖，不把未注册能力写作产品功能 |

保全采用以下当前用户权限可执行的组合，实施脚本使用系统PowerShell/.NET和随包Node，不需要内核采集、提权、性能组或驱动：

| 对象与具体观察方法 | 准入、实际结果与判据 |
| --- | --- |
| 每个实际可写消费者 | 按上表及构建metafile核真实调用闭包；为profile/SQLite、metadata/lock/key、daemon/history、launcher、session/journal、Pi settings/cache逐项记录注入位置、允许根和实际文件。写端在打开/创建前核规范路径及所属，所有可写路径限定DataRoot或本项安装输出；禁止只审顶层环境变量。实际SQLite以`node:sqlite`的`PRAGMA database_list`核主文件位置，并查真实Run/Task/Dispatch/session记录 |
| 路径与端点 | 安装器及写端用Go/Node真实路径解析、Windows文件属性/reparse核验及既有windows-security读取ACL；拒绝本机Orca/账号根、祖先/后代重合和重定向到它们的junction。用本项自有junction指向受保护根作真实拒绝输入，不向目标试写。`Get-NetTCPConnection`核loopback端口及OwningProcess；真实RuntimeClient核私有metadata中的runtimeId/endpoint/PID/startedAt，私有锁冲突和占用端口均在使用前拒绝 |
| 进程与关闭目标 | `Get-CimInstance Win32_Process`定向读取已知runtime/daemon/helper/Worker的ProcessId、ParentProcessId、CreationDate、ExecutablePath；与原生host/incarnation及精确所属记录互核。只保留非秘密字段；创建时间或所属失配时在调用树终止前拒绝。原生退出观察与后代检查证明实际全树退出，不能用一次进程枚举的缺项推断退出 |
| 原Orca配置和认证 | T00按真实路径解析确定原Orca安装、profile/config、Agent配置及本次选定Codex认证文件清单；用`.NET FileStream(FileMode.Open,FileAccess.Read,FileShare.ReadWrite\|FileShare.Delete)`读入内存比较字节，用`Get-Acl`比较owner/SDDL，仅保存路径类别、相等布尔值和变化时间，不保存秘密字节。读取保持原客户端共享权限，及时关闭读取句柄。T02、并发期间、停止A后和T15重复；同时用原绝对CLI的`status --json`核原runtime/daemon/会话归属及查询可用 |
| 定向目录变更通知 | 在上述明确配置/认证父目录使用`.NET FileSystemWatcher`，按FileName/DirectoryName/LastWrite/Size/Security通知记录创建、修改、删除、重命名及Error；范围只含已审目录，不读取会话正文。先在本项自有目录实际验证通知，再观察受保护目录。该接口不提供writer PID，不能独自归因或证明没有写入；溢出/不可读或自然变更无法解释时保留相应范围证据不足，不把无事件当完整隔离证明 |

上述文件变更通知的能力和缓冲区溢出限制依据[Microsoft FileSystemWatcher文档](https://learn.microsoft.com/en-us/dotnet/api/system.io.filesystemwatcher?view=net-9.0)。实际隔离结论依赖写入消费者准入、精确运行事实和前后保全的共同结果。原客户端可正常写自身日志/状态或更新认证；发生变化时核原客户端实际记录及产品写端和SDK来源，无法解释的具体文件保留证据不足，不回写原值。不要求活跃Orca整棵目录逐字节不变，不宣称恶意同用户隔离或系统调用级全量追踪通过。

### 3.2 从普通终端或原有Agent取得真实控制身份

原生`runCreate`要求stable pane/session，普通外部终端不能凭空取得该身份。方案增加一个**实际运行的控制helper**，只承担已有命令的传输和身份绑定，无模型或自动调度循环：

1. `swa controller open`通过现有项目scoped token向Go App请求，重查项目、批准及允许动作。后台在私有runtime中创建真实后台PTY，启动随包`node swa-controller.js`；保存原生terminal/pane/process incarnation及实际PID/创建时间。该内部创建传入真实`launchConfig`，其`agentCommand/agentArgs/agentEnv`仅描述随包helper命令，不传`startupAgent/launchAgent`。`O:src/main/agent-launch/agent-launch-token.ts:agentLaunchTokenFor`仅在launchConfig存在时mint token，`terminal/launch/terminal-create.ts`负责环境注入和PTY commitment；普通空壳terminal没有这项token，不能当控制身份。此helper不登记成模型Agent，不伪造hook状态。
2. helper从自己原生launch环境读取宿主产生的真实token/handle和产品metadata，经RuntimeClient发起请求。open只创建并核实helper，不创建Run；每个controller只服务一次用户任务，`task create`随后由它创建一个Run及根Task。用户命令不接受`--from`、pane、consumer、actor或launch token；控制helper只以自己的实际身份调用runCreate/runUse和原生任务/Worker方法。
3. 外部终端和已有Agent使用同一`controller`选择器及原有项目token。后台在每次写入核准主体、项目、批准、root Task、controller的完整活身份后，经原生terminal输入通道交付命令；helper只接收版本化JSON动作，由成熟JSON解析器解析，不执行调用者传入的shell字符串。命令结果使用原生request回执及私有终端输出返回；helper传输自身不另建持久队列或任务表。
4. 严格产品admission调用原生精确launch校验，不能因上游核验返回null就沿宽松declared-handle分支放行。helper的输入接受与请求执行分别记录；传输未知时按原request ID查询，不再次terminal.send。真实收到的correlation ID用于定位，不充当权限凭据。
5. 本机Orca内的原有Agent调用swa时，其ORCA环境属于另一个宿主。产品入口只用产品controller，不把该环境转换成产品Worker或human；产品子进程清除外部Orca身份、pairing、account及remote选择变量，再注入自己的真实环境。父进程环境不改动。
6. runtime重启后，能核实原生恢复证明的原helper才继续；证明不足返回`controller_unverifiable`，只读查询仍可用。确认原helper真实退出后才允许创建新helper并由它调用原生runUse推进consumer generation。未知时不重绑、不创建竞争控制端。

控制helper是需要实现并真实验证的适配部分，不能由当前开发终端的身份代替。本地owner token与同一Windows用户的权限边界仍属协作式控制；不宣称阻止恶意同用户读取其可读凭据。应用层项目/批准检查不因这项边界而省略。

### 3.3 controller请求、原生去重与恢复

准确上游`shared/orchestration-rpc-contract.ts:isDurableMutation`未包含`terminal.create/close`。`terminal-lifecycle-methods.ts`将create交给`dedupeTerminalCreate(clientIdentity,worktreeSelector,clientMutationId,reconcileExisting,run)`；`terminal-create-deduplication.ts`解析真实workspace，以原生`deriveRemoteRuntimeTerminalCreateHandle`从clientIdentity、workspace ID和clientMutationId得到确定handle，并传入`preAllocatedHandle`。`RemoteRuntimeTerminalCreateIdempotency`只合并in-flight，成功或失败均移除；`reconcileExisting`找不到时会执行新创建，不能承担未知操作的只读恢复。

产品需要在**同一原生mutation executor及orchestration.db的mutation_receipts**中登记准确方法名：`swa.controllerOpen`、`swa.controllerClose`、`swa.taskCreate`、`swa.controllerDeliver`。这些为待实施的产品composition方法，非声称上游已有；将它们加入现有durable方法分类并复用`OrchestrationMutationExecutor.run`、原生canonical payload比较、begin/checkpoint/complete receipt及`orchestration.requestShow`。没有新增表、执行状态或payload hash实现。Go App先核真实项目主体/批准，再由受保护私有RuntimeClient发起；入口不允许外部原始RPC，普通CLI输入不能选择caller fingerprint。

| 公开动作与请求R | 精确原生映射、保存位置及副作用顺序 |
| --- | --- |
| `controller open` | Go App规范化并核实际project、workdir、批准与主体，把这些非秘密字段全部放入`method=swa.controllerOpen`的原生payload，`orchestrationRequestId=R`，使用私有runtime真实local caller fingerprint。原生receipt在任何PTY副作用前pending；其checkpoint保存真实workspace ID、固定clientIdentity=`swa-controller-local`、clientMutationId=`R/controller-pty`、调用原生derive函数所得expected handle及stage。产品handler直接复用原生`runtime.dedupeTerminalCreate`，明确传上述clientIdentity、`worktree=id:<真实ID>`、clientMutationId、`reconcileExisting=false`；callback调用真实`runtime.createTerminal`，传原生preAllocatedHandle及§3.2真实launchConfig。这与原terminal.create handler使用相同的实际原生操作，不依赖RuntimeClient未提供的clientId参数。创建后保存实际pane/host/incarnation/PID/创建时间；真实helper ready和launch核验满足后才complete receipt，controller selector为该原生handle，runId此时为空 |
| `task create` | 在同一原生executor先登记`method=swa.taskCreate,R`及精确controller绑定，checkpoint记录子请求`R/run-create`、`R/root-task`。只向真实helper输入一次结构化命令；helper以自身真实身份执行`orchestration.runCreate`，header请求为`R/run-create`，随后以得到的实际runId执行`orchestration.taskCreate`、请求`R/root-task`。两项均有各自原生receipt；返回真实Run/根Task后完成外层receipt。失败保留已创建Run/根Task，禁止用第二个controller/Run替代；controller已有根Task时拒绝创建另一用户任务 |
| `controller close` | 原生`method=swa.controllerClose,R`的payload固定controller handle及当前open receipt中的精确incarnation/host/进程身份，先重核授权和所有所属Worker已退出/归档/释放。checkpoint在关闭前保存exact identity及stage，然后调用原生精确`terminal.close`和全树退出观察；没有Run状态删除动作。全树退出明确时complete，未知保持pending/EXIT6及占用，不能借close清除残留 |
| 其他经helper的写操作 | `swa.controllerDeliver`以公开R登记原生outer receipt，payload含真实主体/project/controller/动作及完整输入；在一次terminal输入前checkpoint保存固定子请求`R/native`、精确controller绑定及stage。helper仅一次执行对应原生方法，native请求为`R/native`；`workerStart(spec,parent=<根Task>)`在原生接受内创建子Task/Dispatch，无先行独立taskCreate。send定位真实input operation。外层unknown不再次terminal.send，native及外层receipt均可由公开R查询；只有原生动作明确完成才完成outer receipt |
| 无controller的查询 | 增加`swa request show --project P --request-id R --json`，用于open回执丢失而没有controller selector时；与`--controller C`互斥。Go App核原主体/project，再读取原生requestShow和相关native状态；不启动helper、不发送输入、不根据absent创建资源。返回open/close checkpoint、原始副作用身份和真实观察；项目不匹配在输出前拒绝 |

同一公开R的所有controller/task组合步骤仅使用上述固定子请求名称，映射在原生receipt中保存，不放在CLI缓存。native caller fingerprint为同一私有runtime的真实local身份，因此同R更改project、主体、workdir、controller或正文会由原生payload比较得到`request_mismatch`，不能换clientIdentity回避冲突。并发同R由原生executor合并，已完成同R只回原receipt；PTY的in-flight去重另外使用原生clientMutationId，不把transport response ID当去重键。

open在PTY请求开始后抛出或失联时，产品handler将该已可能发生的效果返回`operation_unknown`，保留原生pending及checkpoint；不能让原executor的普通异常清理把它变成可重建。close及task create也在首次外部副作用前checkpoint，之后的异常保持相同unknown规则。此处错误转换保留具体cause/stage和已有身份，不捕获后继续执行。原executor对pending非release动作本就拒绝再次invoke，产品不增加自动恢复执行分支。

响应丢失或runtime重启后，查询按原checkpoint只读核精确handle、真实daemon `listProcesses`、持久PTY/session归属及原生Run coordinator/子请求receipt。已创建PTY而尚无Run的状态保留该helper；Run已创建而根Task未确认时返回原Run及pending子请求。原生reconcile可核exact handle的既有PTY，但产品查询不得调用会在无匹配时spawn的`dedupeTerminalCreate(...,reconcileExisting=true,...)`。未知、无inventory、缺handle或多重匹配均不创建、不重发、不释放；单独absent不证明无副作用。需要新controller时先获得原精确helper和后代已退出的正向证明，再由新的明确请求创建；绑定已有Run仍按§3.2真实runUse及generation规则。

## 4. 当前Agent与非Aliyun受限正向方案

### 4.1 实际静态观察

| Agent | 本次观察 | 首版适用状态 |
| --- | --- | --- |
| Codex | 已安装0.162.0，当前开发确实运行gpt-6.1-sol。`~/.codex`及当前Orca account home的TOML/JSON白名单解析均为默认OpenAI、chatgpt OAuth、有tokens、无API key | 当前公开原生控制器不能可靠关闭全部原生文件工具；维持受限拒绝。开发会话不充当ToolsOff证据 |
| Claude | 已安装2.1.238；JSON白名单观察到qwen3.7-max与dashscope来源，且含hooks/statusLine | 主控明确禁止调用该配置；不修改全局设置、不自动选择其他模型 |
| Pi | 已安装1.0.1，系统settings只记录版本，系统auth文件为2字节。真实Pi包包含openai-codex provider、chatgpt.com/backend-api、gpt-6.1-sol和原生RPC/SDK | 系统Pi当前认证成功未知。产品私有Pi SDK路径具静态兼容依据，待真实验证 |

只读检查没有输出认证值、账号ID或私有对话。当前Orca终端`terminal show`仅证明实际开发客户端；没有把它作为产品执行或工具准入结果。

### 4.2 复用Pi原生provider和只读OAuth

准确接口来自`P:dist/index.js`导出的`ModelRuntime`、`DefaultResourceLoader`、`createAgentSession`、`createAgentSessionRuntime`、`SessionManager`、`SettingsManager`、`runRpcMode`，以及对应`.d.ts`、`docs/sdk.md`、`examples/sdk/13-session-runtime.ts`。采用产品`Pi SDK 1.0.1`入口，调用真正的`runRpcMode`；不仿造pi CLI、不手写Pi原生协议或模型HTTP。

`P:node_modules/@earendil-works/pi-ai/dist/providers/openai-codex.js`使用原生`openai-codex-responses` API及`https://chatgpt.com/backend-api`；其catalog文件`providers/data/openai-codex.json`包含`gpt-6.1-sol`。OAuth来源`auth/oauth/openai-codex.js`具有真实OAuth flow与toAuth，使用access bearer；产品不变更endpoint、provider、账号或模型，未通过就返回具体错误。

只读CredentialStore是该SDK明确允许注入的基础设施端口，执行以下规则：

1. 仅从明确选择的当前Codex account home读取原JSON，并使用成熟`jose 6.2.3`解析JWT有效期和必要标准声明，禁止手写split/base64 parser；该版本和integrity已在准确Orca的`pnpm-lock.yaml`找到，复用其pin并列为直接实施依赖。TOML使用`smol-toml`，JSON使用标准解析器。JWT解码只解释已获准原件中的过期字段，不替代提供方的签名/认证判断。
2. `read/list`仅向同一`openai-codex` provider返回真实OAuth字段及非秘密元数据。原access/refresh/account字段保持来源，expires来自真实标准声明；不伪造有效期，不把OAuth改称API key，不创建可写认证副本。
3. 每次read先核真实有效期足以覆盖本次最长请求和SDK五分钟刷新界限；不足或无法解释即在模型请求前拒绝。read按当前原件取快照，不刷新、不撤销、不写回本机登录。
4. `modify/delete`立即抛出明确只读错误，**不执行传入的callback**。原生`auth/resolve.js:resolveStoredOAuth`只在`credentials.modify`的callback中调用`oauth.refresh`；`models.js:resolveRefreshCredential`亦如此。两条路径都在网络刷新之前被拒绝；不能先refresh再阻止保存。
5. `ModelRuntime.create({credentials, modelsPath:null, allowModelNetwork:false, refreshOnCreate:false})`注册SDK自带的静态catalog对象，不执行初次catalog/auth refresh；这些随包元数据的注册仍涉及全部内置provider，不宣称已经移除它们。后续只调用真实`refresh({providers:['openai-codex'],allowNetwork:false})`：源码明确分支只更新该provider的catalog及`refreshProviderAvailability`，其getAvailable/checkAuth/credentials.read均限定该ID；不会进入全量queueAvailabilityRefresh/credentials.list。检查aborted、errors、getError，任何错误立即失败。再显式`getModel('openai-codex','gpt-6.1-sol')`及该模型getAuth，核实际OAuth来源和模型身份。产品不探测其他provider的环境/云账户，不调用全量get_available_models，也不把静态catalog注册说成实际认证通过。
6. 原凭据自然被原有Codex更新时，下次read校验同一来源与当前授权。产品不写原auth，不复制刷新权；解析错误保持失败。后续产品长期登录能力另行设计，本次可通过重新读取仍由原客户端正常管理的当前有效access继续获准固定任务。

### 4.3 工具、扩展、输入和实际完成

使用公开`createAgentSessionRuntime(factory,{cwd,agentDir,sessionManager})`。该函数真实调用factory并用返回的session/services构造原生AgentSessionRuntime。factory按以下顺序执行，所有路径均显式私有，不调用`createAgentSessionServices`：

1. 宿主先在SDK子进程的独立launch env中把`HOME/USERPROFILE`设为该Worker的`<DataRoot>/workers/<实际Dispatch>/home`，`PI_CODING_AGENT_DIR/PI_CODING_AGENT_SESSION_DIR`分别指向其私有agent/session目录，`PI_PACKAGE_DIR`固定随包SDK根；拒绝从调用者继承这些选择值。只修改传给真实子进程的env map，不给当前PowerShell的`$HOME`赋值，不修改父进程、用户或系统环境。原Codex认证通过已获准的唯一显式绝对路径只读访问，不依赖这些私有home去查认证。
2. 用上述显式选项创建真实ModelRuntime；`SettingsManager.inMemory`提供固定设置，packages和各资源路径为空、`defaultProjectTrust:'never'`且项目trusted=false。
3. 直接`new DefaultResourceLoader`并`await reload()`；`noExtensions/noSkills/noPromptTemplates/noContextFiles/noThemes`均为true，additional资源和extensionFactories为空。非空固定systemPrompt及`appendSystemPrompt:[]`在原生discovery之前阻止SYSTEM/APPEND发现；不用加载后的Override过滤器承担准入。核实际扩展和pending provider/native/virtual registrations为空，非空立即失败。
4. 仅执行§4.2的provider定向refresh并检查真实返回的errors/aborted/auth snapshot，取得并验证固定model及真实getAuth；没有选定OAuth或来源变化则在首次prompt前失败。原生`isUsingOAuth/hasConfiguredAuth`来自该次真实定向snapshot，不填假状态。
5. 构造符合公开`AgentSessionServices`接口的对象，字段为实际cwd/agentDir/modelRuntime/settingsManager/resourceLoader/真实diagnostics；调用`createAgentSession`并显式传全部真实对象及固定model/thinkingLevel、sessionManager、sessionStartEvent和空工具参数。`sdk.js`在已有modelRuntime/resourceLoader/model时不创建默认来源、不reload、不寻找默认model；modelFallbackMessage或诊断错误立即失败。factory返回真实session结果、同一services和diagnostics，由原生runtime接收后运行真正`runRpcMode`。

`P:dist/core/package-manager.js:resolve/addAutoDiscoveredResources`在resource-loader的noSkills等过滤前执行，其中用户`.agents/skills`来自`getHomeDir()=process.env.HOME||homedir()`；因此空settings和noSkills本身不表示没有发生目录发现。上述真实子进程env把该消费者也导向空的私有home；`config.js:getAgentDir/getPackageDir`同样使用明确的私有目录与随包资源。T04在首次prompt前核子进程这些非秘密解析路径、真实资源来源列表及对应source调用闭包，任一解析到本机用户资源根立即失败。

该组装没有改写依赖或替换refresh；定向refresh执行真实SDK逻辑，services包含真实已加载对象。后续允许的prompt继续走同一modelRuntime及只读store；会话替换等非白名单RPC不开放，factory本身仍保持相同限制。只读公开项目输入由Workspace固定提供，不继承项目本地配置。

`createAgentSession`传入`noTools:'all'`、`tools:[]`、`customTools:[]`。源码`dist/core/sdk.js`把此组合转换为空allowlist及空active tools；`DefaultResourceLoader`不加载扩展/MCP。启动与每次输入前检查原生`session.getActiveToolNames()`为空、注册工具集合及resource列表为空，并从实际请求前原生session状态核对。仅零工具调用次数不能通过该检查。

禁用agent retry、provider retry、autoCompaction、cacheWarming，transport固定`sse`，拒绝modelFallbackMessage及任何模型身份变化；这些来自已读取`docs/settings.md`的真实设置及对应SettingsManager。原Pi API的默认transport auto与retry设置不沿用。内部RPC只允许`get_state/get_messages/get_session_stats/get_last_assistant_text/prompt/abort`六项；不向外暴露`bash`、全provider枚举、任意工具启用、任意model切换或session目录选择。原Orca Pi options查询依§4.2改为仅当前真实state，不能冒充已进行全provider动态catalog核查。

`P:dist/modes/rpc/rpc-mode.js`的`bash`可直接执行命令，不经过模型tools。因此六项白名单必须在产品父进程持有的真实`JsonlRpcAgentConnection.request/send`写入边界校验；SDK原生stdin/stdout仅交给该owned父子管道，不绑定网络端口、不接受CLI原始RPC。外部输入只能进入严格参数schema的`prompt.message`，绝不把任务正文解释为RPC envelope。违规type/extra字段在向子进程写入前失败；不合成一个看似Pi返回的拒绝帧。

已读取原生`rpc-session.ts`和`rpc-session-adapter.ts`，必要消费者逐项适配为：

| 消费者 | 产品允许的原生调用与处理 |
| --- | --- |
| `PiRpcSession.start/readOptions`、`rpc-options.ts` | 保留get_state，核实际provider/model/sessionFile；移除全量get_available_models/get_commands及保存选项的set_model/set_thinking_level恢复，初始模型/effort由SDK显式固定。变化立即拒绝，不吞错写skipped |
| `rpc-turns.ts`、`rpc-context-usage.ts`、`rpc-idle-check.ts` | 保留prompt、get_state、get_session_stats的真实发送/事件对应和idle检查；结果读取可用get_messages/get_last_assistant_text，均经同一白名单 |
| `cancelTurn/awaitStoppedRequestEnd/closeSession` | 保留abort、get_state及真实ChildProcess root/tree close证明；不增加bash式停止命令 |
| `rpc-compaction.ts`、dialog callbacks、option写入口 | 首版明确不提供compact、工具/模型变更或extension UI回复能力；native capability中返回不支持，unexpected扩展事件立即失败。保留已有turn、journal与退出处理，不重写原生RPC |

SDK的`getActiveToolNames()`及工具allowlist为空，与传输动作白名单分别验证；其中任一缺失都不能记ToolsOff通过。

隐式模型请求采用原生真实设置显式关闭，memory settings包含：

```json
{
  "cacheWarming": "off",
  "compaction": { "enabled": false },
  "retry": { "enabled": false, "maxRetries": 0, "provider": { "maxRetries": 0 } },
  "transport": "sse"
}
```

已核实际调用顺序：`SettingsManager.getCacheWarmingMode`未设置默认streaming，`sdk.js:streamFn`每次调用CacheWarmer.start；`cache-warmer.js:start`在mode=off时于创建run/timer之前返回，因此不安排一token暖缓存重放。`agent-session.js:_willRetryAfterAgentEnd`在retry.enabled=false时返回false，`_checkCompaction`在compaction.enabled=false时于overflow恢复和agent.continue之前返回false。`sdk.js:buildRequestOptions`把provider.maxRetries=0传入`api/simple-options.js:buildBaseOptions`，再到`openai-codex-responses.js`的SSE循环，只有attempt=0一次HTTP调用；流异常保持真实error/unknown。transport=sse在进入WebSocket分支前即绕开其缺失continuation重发及WS/SSE切换，不依赖默认auto。所有新会话和原生接续均使用同一memory设置，不从session或用户文件恢复这些开关。

T04必须读回上述实际设置和`session.cacheWarmingStatus`的inactive/disabled状态；T06/T14核断开/失败后的真实请求及provider事件，没有auto_retry/compaction/warm再次请求。无法观察请求次数时相应证据保持不足，不能仅凭最终一条assistant消息声称没有重放。

Orca的原Pi launch仅允许fullAccess；产品Pi SDK composition以真实noTools状态替换该准入，明确标注适配差异，不能将fullAccess布尔值置true来通过原检查。保留原生Orca structured session/journal/fence及Pi session文件，双方记录provider session ID与原native事件。

工具关闭的Worker只输出任务结果，不执行Orca命令。产品preamble明确结果由宿主读取，不要求worker_done。`agent_settled`证明本轮原生执行结束；结构化结果及固定事实检查通过后，实际控制helper调用workerStop并核全树退出、原生归档/释放，再用自己的consumer执行taskUpdate。任意缺失不生成假worker_done、假hook、零费用或人类批准。

## 5. 固定抽取与完整安装输入

### 5.1 源码、入口及传递依赖

固定采用准确上游的源码作为构建输入，保留原路径和完整MIT/Lovecast许可。来源快照只作为构建输入；随包runtime由以下入口及其真实传递依赖打包，禁止只复制process-host或把RPC类型当执行实现：

| 部分 | 必须保留的真实源码/构建输入 | 产品适配 |
| --- | --- | --- |
| Node组装与生命周期 | `src/main/orcad/`中的host、profile startup、instance lock、process-start-time、lifecycle、stop request、runtime lifetime；`src/shared/app-environment.ts`及secret-store端口 | 新产品composition按§3显式组装；不直接调用有全局副作用的startOrcadRuntime |
| Store/SQLite | `src/main/persistence/`、`src/main/sqlite/`、`src/main/orca-profiles/`及shared profile storage/contracts | 保留真实迁移、authority、writer/backup；SQLite由Node24的builtin提供，不以JSON假存储替代 |
| 原生协作状态 | `src/main/runtime/orchestration/`、相关`runtime/rpc/methods/orchestration/`、runtime workspaces/terminal/session实现及shared契约 | Run/Task/Dispatch、messages、mutation receipts、worker resources/archives原样复用；按项目准入包装必要命令 |
| 会话host和Pi | `src/main/native-chat/agent-session-wire/`、runtime structured-session/worker实现、`src/main/pi/`、provider-process及native journal/record store | Pi launch改为明确SDK入口；保留原生RPC connection/translator/exit机制；移除fullAccess单一准入，改核真实noTools |
| 进程/PTY/Windows | `src/packages/process-host/`、`src/main/daemon/`、`src/main/ipc/pty/`、provider-process、`pty-descendant-exit-verification.ts`、`src/main/windows/`及`native/windows-process-tree`、`native/windows-security`、`native/windows-cli-launcher` | 真实控制helper需要PTY/daemon；核PID+创建时间与owned范围后才树终止；Windows安全原生扩展用于私有文件ACL |
| CLI与RPC | `src/cli/runtime/`的client/metadata/transport/source fence、必要parser/schema/output；runtime-rpc及必要shared rpc contracts | 公共入口为Go swa；helper调用真实RuntimeClient，不调安装中的orca、不暴露全量RPC |
| 打包 | `config/scripts/orcad-entry-build.mjs`、原生stage/prebuild/runtime-pin脚本及它们的实际imports，`src/shared/orcad-artifacts.ts`、`node-runtime-pin.ts`、package/workspace/lock/tsconfig及被引用资源 | 产品入口替代原完整build脚本；输出目录、缓存和所有TEMP/TMP仅在获准忽略目录，不直接执行上游含smoke/tmpdir的完整build |

runtime类跨多个继承模块并静态引用工作区、布局、账户及其他功能；这些Node可达支持文件与shared/types不能按目录名称随意删除。为保证上述构建输入完整，实施保留准确upstream源码快照及其lock作为本地来源，使用既有esbuild解析真实入口的传递import，生成普通构建metafile供审查；最终输出只包含实际入口可达代码和明确native资源。未注册/未启动的能力不成为产品入口。metafile不参与任务状态、请求幂等或验收权威。

静态调查已经固定来源、入口、必要native/children和原生依赖族；**当前没有运行构建，尚不能证明最终可达文件集合或可安装闭包已通过**。实施的T01必须核全部产物、动态import和fork子入口；发现表外依赖或全局副作用先补充来源与审核，不能以缺文件时调用系统Orca兜底。该限制如实保留，不将来源快照等同于功能完成。

### 5.2 必要npm、runtime及子进程文件

- 原生Node固定`24.21.0`，来自`O:src/shared/node-runtime-pin.ts`；其Windows x64/arm64实际asset pin及原生N-API要求沿原文件。本轮验收针对当前Windows x64；不宣布其他平台已验。host当前Node24.16.0的历史观察不能替代随包runtime。
- 上游root runtime依赖声明为：`@anthropic-ai/claude-agent-sdk 0.3.284`、`@electron-toolkit/utils ^4.0.0`、`@floating-ui/dom 1.8.0`、`@linear/sdk ^97.0.0`、`@orca/process-host workspace:*`、`@parcel/watcher ^2.5.6`、`@xterm/addon-serialize 0.15.0-beta.301`、`@xterm/headless 6.1.0-beta.303`、`agent-browser ~0.27.0`、`electron-updater ^6.8.10`、`fflate 0.8.3`、`i18next ^26.4.2`、`jsonc-parser ^3.3.1`、`node-pty ^1.1.0`、`posthog-node ^5.52.5`、`proper-lockfile 4.1.2`、`qrcode ^1.5.4`、`react-i18next 17.0.15`、`serve-sim 0.1.47`、`smol-toml 1.9.1`、`ssh2 ^1.17.0`、`stream-chain 4.2.6`、`stream-json 3.7.0`、`tldts 7.4.16`、`tweetnacl ^1.0.3`、`ws ^8.22.0`、`yaml ^2.9.1`、`zod ~4.6.5`。准确解析版本以该pin的`pnpm-lock.yaml`为准，禁止按这些范围升级取新版本。此列表说明原输入，不表示全部都应随最终bundle独立安装。
- 构建复用上游esbuild `^0.28.2`及锁定工具；product Pi SDK及`pi-ai`明确pin `1.0.1`，带其真实production依赖/资源/许可。SDK公开入口还会加载其TUI/theme、session及native图像支持模块，不能只放一份rpc-mode.js。实现期安装需锁定实际解析树，不依赖用户全局node_modules。新增直接依赖`jose 6.2.3`沿准确upstream lock中的版本/integrity固定，当前未安装；不会把仅源码锁中存在写成当前可import。
- Windows node-pty除JS loader外必须携带`conpty.node`、`conpty_console_list.node`、`conpty/conpty.dll`、`conpty/OpenConsole.exe`；JS文件列表完整沿`ORCAD_NODE_PTY_JS_ARTIFACTS`。watcher携带真实`node_modules/@parcel/watcher/index.js`及`watcher.node`；Windows tree/security/CLI launcher使用对应架构原件，不用PowerShell命令拼接替代多行argv。
- 原`ORCAD_CHILD_ENTRY_POINTS`九个入口为：`parcel-watcher-process-entry`、`daemon-entry`、`profile-state-writer-worker-entry`、`profile-state-backup-worker-entry`、`foreign-sqlite-reader-entry`、`port-scan-command-worker-entry`、`session-scanner-service-entry`、`worker-thread-child-process-entry`、`claude-profile-setup-worker-entry`。产品必须保留实际启动的daemon/writer/backup/watcher/worker child；未组装功能对应child不得仍被动态调用。若原runtime的实际可达调用仍需要其余child，就随包保留并核本任务专属路径，不用缺省空处理器满足加载。
- 原native资源定位以bundle所在目录为根，不能依赖shell cwd。保留需要的emoji、ripgrep及其第三方许可等实际引用资源；未启用的browser/mobile/visuals不进入安装功能。Electron静态import不得进入最终Node入口；Node内建SQLite的动态加载仍须真实通过。

### 5.3 实施写域建议及安装行为

以下准确实施写域已由主控f7的首版任务授予。新增模块以必要目录列出，既有文件逐文件列出，不据此扩大为整个Workspace、HTTP或其他产品目录：

| 路径 | 必要修改与复用位置 |
| --- | --- |
| 新增`cmd/swa/**` | 真实Go CLI、参数/退出/JSON处理、serve启动及相应真实检查；协议边界只消费App或已装Go HTTP，不复制执行状态 |
| `cmd/server/main.go`；新增`cmd/server/orca_runtime.go`、`cmd/server/orca_runtime_test.go` | 现有服务composition注入Orca执行端口、私有配置/生命周期和实际组装检查；已有业务实现保持各自归属 |
| 新增`internal/workspace/app/orca_execution.go`、`internal/workspace/app/orca_execution_test.go` | Workspace的Orca消费方端口、DTO及用例；复用`LocalProjectRepository/ProjectFiles`、项目Grant和`ValidateProjectActionScope/AuthorizeProjectActionCaller`，每次重查主体与适用动作；固定上下文仍走既有显式服务 |
| 新增`internal/adapters/orca/**` | App端口的真实适配、安装路径/metadata/进程所属检查、原生RuntimeClient调用及相应真实测试；传输使用随包Node入口导入准确原生RuntimeClient，Go侧仅以标准JSON与其自有stdin/stdout通信，避免重写Orca framing/协议或手写named-pipe parser |
| 新增`internal/adapters/sqlite/orca_execution.go`、`internal/adapters/sqlite/orca_execution_test.go`、`migrations/028_orca_execution_links.sql` | 只保存Workspace项目/批准到真实native Run/root Task的关联；不保存另一套Task/Worker状态。当前迁移最大027，序号须由主控与公共迁移归属核准后使用。现有`migrations/embed.go`已按`*.sql`嵌入，无需改动 |
| `internal/adapters/httpapi/server.go`、`internal/adapters/httpapi/session.go`；新增`internal/adapters/httpapi/orca_cli.go`、`internal/adapters/httpapi/orca_cli_test.go` | Services/路由注册、scopedAgentRoute精确放行与既有sessionGuard/项目token校验，新增HTTP处理只调用App，不直接读数据库，不调用human bootstrap代替批准 |
| 新增`third_party/orca-v1.4.230/**` | 准确b287来源及原MIT/Lovecast许可，保留原路径和锁；只在此产品来源副本实施已审适配，原O目录保持只读。源码目录不作为独立运行候选 |
| 新增`runtime/swa-orca/**` | 产品Nodecomposition、controller/helper、Go消费的真实RuntimeClient入口及Pi SDK入口；限定必要native methods/资源、来源适配和真实测试。构建产物进入既有忽略dist/verification，不把状态或凭据放入源码 |
| `package.json`、`pnpm-workspace.yaml`、`pnpm-lock.yaml`、`dependencies.lock.json` | 声明新runtime workspace、准确SDK/JOSE/构建依赖、实际锁及许可清单；原产品Node24.16.0工具链与随包Orca Node24.21.0分别记录。Go所需端口/JSON/Windows基础依赖已在go.mod内，本方案不申请Go依赖升级 |
| `contracts/openapi.yaml`、`web/src/api/schema.d.ts`；新增`contracts/swa-cli.schema.json` | 首版HTTP及CLI契约、真实状态/错误/unknown和生成类型；schema.d.ts沿既有generate-client生成，不手改。无网页功能重构 |
| `scripts/check-dependencies.mjs`；新增`scripts/build-swa-runtime.mjs`、`scripts/build-swa-package.mjs`、`scripts/check-swa-runtime.mjs`、`scripts/install-swa.ps1`、`scripts/verify-swa-cli-headless.ps1`及`tests/swa-cli-headless/**` | 精确runtime依赖检查、原生入口/资源闭包、候选打包/安装和T00—T15真实检查。Node架构/全局写端检查直接覆盖本新增runtime；既有Go架构检查已覆盖Workspace App及HTTP边界，无需为数量另加规则 |
| 新增`THIRD_PARTY_SWA.md`；本任务两份文档及自有`.verification/swa-cli-headless-20261020/**` | 上游Orca MIT/Lovecast、Pi/Node/native/实际传递依赖许可和来源说明；候选实际结果/限制/命令保留。原LICENSE不替换；HANDOFF/STATUS/SPEC/QUESTIONS继续由主控维护 |

现有`.gitignore`已经覆盖node_modules、dist及.verification；不申请其他公共文件。实施发现该清单以外的真实必要消费者时，先交主控具体路径及原因，不擅自写入。

构建使用`go build`现有产品及新增swa、锁定Node构建工具和真实native；安装只写明确InstallRoot/DataRoot，不替换系统orca或Agent，不注册自动启动、不改变用户/系统PATH，不下载私有profile。普通终端通过swa绝对路径调用；如用户另行安装PATH入口，仅新增swa且需专门授权，不影响orca解析。

现有Go HTTP及私有Node/daemon作为同一产品服务组被主控安排在唯一动态窗口内，固定所属关系。原S服务只能由主控保全、停止并在测试后恢复；本Worker本次未停止任何已有服务。构建和安装完成后才启动服务组，过程中不并行运行第二组构建或服务。

## 6. 最少完整CLI规则

以下为已审精确命令契约，当前实现中，尚无运行通过结论。CLI/HTTP/App共用同一准入语义；所有写命令要求显式`--request-id <UUID>`，按§3.3映射原生outer/child mutation receipt，公开键及每一步真实身份均保留。上下文通过UTF-8文件传入；禁止在命令行正文中放凭据。

| 命令 | 输入与实际结果 / 复用 |
| --- | --- |
| `swa serve --data-dir D --port N --json` | 必须明确私有根和端口；打印真实readiness/版本/路径/身份的非秘密部分。缺依赖、锁冲突或端口占用失败，禁止改用其他runtime/端口 |
| `swa status --json`、`swa capabilities --json` | 原生runtimeId/startedAt、数据库authority、daemon/provider支持/当前证据与限制；读操作不调用模型、不隐式启动后台 |
| `swa controller open --project P --workdir W --request-id R --json` | 创建§3真实helper；验证当前项目/批准/写域，回controller selector及真实native身份，此时没有Run。输入的P必须是实际Workspace ID |
| `swa controller show --controller C --json`、`close ... --request-id R` | 精确观察/关闭自有helper；close需所属全部Worker已退出/释放，否则拒绝。不得用terminal全量关闭 |
| `swa task create --controller C --spec-file F --request-id R --json` | 依§3.3创建真实Run及根Task；每controller一次用户任务。F固定目标、公开来源版本、scope与成功条件，返回runId/root taskId及已有批准引用 |
| `swa task show --controller C --task T --json`、`list --controller C` | 原生根/子Task/Dispatch及只读汇总；原请求未决、启动/停止未知和资源占用分别展示 |
| `swa worker start --controller C --task T --name A --agent pi --model openai-codex/gpt-6.1-sol --input-file F --request-id R --json` | 新建T的子Task及原生Worker Dispatch，以真实Pi SDK执行固定公开输入；ready只在真实turn开始证明满足时返回，接受但未证明开始保持unknown |
| `swa worker show --controller C --dispatch X --json`、`list --controller C --task T` | 精确Dispatch、子Task、runtime/session/fence、PID/创建时间、Agent实际状态和root/tree exit；不只输出PTY是否live |
| `swa worker read --controller C --dispatch X --source transcript --limit 100 --json` | 读取native journal/provider自产会话的实际消息、真实cursor/source identity；无原件即缺证。后续分页必须使用原cursor，来源变化拒绝 |
| `swa worker send --controller C --dispatch X --input-file F --request-id R --json` | 验完整当前绑定及授权后向实际原session发送；保留input accepted、turn started与settled；不把写入stdin当开始 |
| `swa worker stop --controller C --dispatch X --request-id R --json` | 精确停止该owned Worker，返回stop request、实际root/tree exit及保留占用；PID/创建时间或ownership失配时不终止 |
| `swa worker release --controller C --dispatch X --request-id R --json` | 只对原生允许结算状态执行冻结输出与释放；未知、人工接管、缺归档或后代未退出时不释放。重复同键只回原receipt |
| `swa task result --controller C --task T --output-dir O --json` | 提取根及各Worker真实输出、来源、检查引用、版本和已知usage；输出路径须不存在或由本命令精确拥有，禁止覆盖其他文件 |
| `swa task complete --controller C --task T --result-file F --request-id R --json` | 当前控制端确认固定成功条件和真实结果，全部相关Worker退出/归档后调用原生taskUpdate；失败或unknown不能完成；完成不表示用户采用或Git发布 |
| `swa request show --controller C --request-id R --json`或`--project P` | 两种scope互斥；project形式支持open未回selector时查询。按§3.3查outer/child原生receipt、checkpoint及原run/task/dispatch/input operation，不生成新请求。无receipt仍不能单独证明未发生外部输入 |

CLI stdout在`--json`下只包含一个完整JSON envelope；诊断写stderr。envelope统一携带`ok/request_id/runtime_id/controller/run_id/task_id/dispatch_id/result/error`中本操作存在的字段，保留原生`code/message/retryable/required_action`和未知语义；不输出token或原账户路径。文字模式显示相同状态与明确下一动作；不通过匹配自然语言决定状态。

| 退出码 | 语义 |
| --- | --- |
| 0 | 查询成功，或写动作取得满足该命令要求的明确结果；查询到failed任务仍可EXIT0，result保留failed |
| 2 | 命令/字段/编码/路径校验失败，尚未发送 |
| 3 | 身份、批准、ownership或scope拒绝，无授权扩大 |
| 4 | 当前Agent/依赖/能力不满足，尚未输入 |
| 5 | 启动前资源/实例锁/端口拒绝 |
| 6 | 请求可能已发生，输入/开始/停止/释放等效果未知；保留原ID，禁止自动重发 |
| 7 | 明确执行或存储失败，按原生记录返回已知影响 |
| 130 | 调用端等待被用户终止；后台任务状态不据此改为停止。若已获知请求身份，stderr给出原request查询命令 |

同键同正文使用原生receipt回放，同键不同正文`request_mismatch`；不自建另一份payload hash系统。未知请求不能换键重发、换Agent/模型/目录或降级为另一路径。指定source缺失不回到terminal tail；失联不视为退出；root exited而tree unverifiable继续保留占用。既有usage缺失显示unknown，subscription目录价格不充当实际费用。

## 7. 从最新目标独立推导的完整真实测试计划

### 7.1 样例、资源与命令约定

已核存在Agent创建的公开验证项目原件：`C:/Users/DW/orca/workspaces/Astrocyte/product-swarm-20261015/.verification/s3-public-stats`，本次HEAD为 `da3bdddb085808863090a05f60cc1c64415039cf`，origin为`https://github.com/songconmaisaix31-design/astrocyte-s3-public-stats.git`。本次完整读取该准确提交的README、go.mod、stats.go、stats_test.go四文件，来源记录见已有`docs/standards/swarm-s3.md`的公开建库记录；未读用户作品。动态T00仍需匿名核公开可读及同一准确提交，不把历史公开记录当当前网络可用保证。

固定业务目标为**基于这四份公开原文，交付Median设计及验证说明**，成果采用真实执行输出，无程序修改：A分析实现、溢出/输入保持等要求，B独立核查Mean与Median边界、生成最终检查说明。两者先各自分析相同固定原文，实际模型执行必须重叠；A被指定停止后B继续，并在同一真实session接受一次包含A已取得输出及原来源的明确后续输入。B说明必须覆盖奇偶长度、负数/重复、空值及NaN/±Inf、有限大中间值溢出、输入顺序保持及原Mean不变，引用准确文件位置。按原文人工或固定程序断言检查这些明确事实，不用模型自评作判据，不把它写成已经实现Median或旧S3通过。

根Task的固定成功条件明确允许按要求中断A，并要求B形成最终说明；A的子Task保留原生stopped/failed事实，不改写为成功。根Task只有在该预先声明的流程和真实成果均满足后才能由控制端显式确认completed，查询始终展示A的实际终止原因。

以下命令中`$swaBin/$installRoot/$dataRoot/$caseRoot/$controllerId/$rootTask/$dispatchA/$dispatchB/$projectId`均由真实安装/准入回执设置，不能填虚构值。`$caseRoot`限定本任务忽略目录；`task.json/a.txt/b.txt/b-followup.txt`由T00从上述准确公开Git内容及固定指令生成，文件不含凭据。`$projectId`来自现有Workspace的实际登记与现有human批准；没有有效批准时先保留拒绝，主控安排当前用户批准操作，产品CLI不自取human身份。

拟新增`pwsh -File scripts/verify-swa-cli-headless.ps1 -Case Txx -SwaPath $swaBin -InstallRoot $installRoot -DataRoot $dataRoot -CaseRoot $caseRoot`作为真实检查入口，`-Case`仅接受下表编号。脚本是后续实施内容，当前不存在、未运行；其职责是执行表内真实CLI/HTTP/进程/SQLite检查并保留实际失败，不模拟provider/原生协议或填入成功事件。测试脚本不增加产品调度状态。

每项独立运行在主控授予的窗口；构建、安装和服务不并行。真实服务组及双Worker是同一次任务窗口；使用RESOURCES的BelowNormal、Go `GOMAXPROCS=4/-p=4/GOMEMLIMIT=512MiB`、Node各进程heap≤1024MiB。每次重操作前实际CPU/可用内存双采样满足CPU<85%、可用≥4GiB，单调间隔2—5秒；Node heap非总RSS上限，同时记录两个Worker及服务组实际RSS和进程数。不修改全局限额或终止其他进程。

### 7.2 T00—T15及可判断结果

全部当前为`NOT_RUN`，下表为实现前审核对象。必要控制脚本或fixture缺失直接失败；不能以未发生的竞争窗口登记通过。

| 编号 | 命令与真实动作 | 通过条件及失败证据 |
| --- | --- | --- |
| T00 来源、基线与授权 | 原件执行`git rev-parse HEAD`、`git show <准确SHA>:<四文件>`；匿名GET公开repo及准确commit；`Get-Command orca -All`、`orca status --json`和只读进程/监听查询；核现有Workspace真实项目/Grant | 固定公开原件、允许输入和成功条件可定位；记录原Orca二进制解析、metadata/runtimeId、PID/创建时间/daemon、配置字节及ACL比较基线；不读取全体会话正文；缺批准保持明确拒绝 |
| T01 构建及闭包 | 获写域/窗口后执行`pnpm install --frozen-lockfile`、产品`pnpm build:swa-runtime`、`go build -p=4 ./cmd/swa ./cmd/server`；逐个加载实际随包entry/children及真实SQLite，检查构建metafile | bundle/Node/native架构匹配、Node24.21.0、无Electron执行依赖、所有实际fork/dynamic依赖在包内；所有TEMP/TMP归本任务忽略目录。不能运行上游mock测试当产品证据 |
| T02 安装与私有隔离 | 使用T01同一候选执行产品`install.ps1 -InstallRoot ... -DataRoot ...`；绝对路径运行`$swaBin --version`和`serve --data-dir ... --port ... --json`；按§3.1逐一核实际可写消费者、`PRAGMA database_list`、metadata/pipe/锁、CIM PID/创建时间及监听；对已占用真实端口、原Orca根及本项junction指向原根提供真实拒绝输入 | 指定端口冲突EXIT5且无另选端口；路径在写入前拒绝，不能试写受保护根；私有profile-state.db/orchestration.db/pipe/daemon/key及所有实际session/cache根正确。§3.1定向配置/认证字节/ACL、目录通知和原绝对CLI status一起核保全；全局迁移/hooks/账号消费者源级关闭。当前权限执行，无内核采集前置；具体不可解释变化保持证据不足 |
| T03 控制身份及权限 | `controller open/show`及`task create`；使用真实过期/撤销或缺失项目token、另一个真实项目/Run/旧controller访问；尝试CLI `--from`/human参数 | 仅真实helper及当前consumer可写；未知flag EXIT2，无效主体EXIT3；无headerless bootstrap调用、human cookie或伪consumer；Controller PID/创建时间和真实native pane可查；旧代请求拒绝 |
| T04 Pi认证及ToolsOff | `capabilities`；无prompt准备真实Pi SDK并按§4.3 factory顺序加载。首次prompt前检查真实定向refresh结果、auth snapshot及getAuth来源，CredentialStore在实际read/list/modify/delete入口记录非秘密provider ID和操作，未知ID立即拒绝；核真实子进程home/agent/session/package解析路径和过滤前的package-manager来源，从源码调用闭包核没有全量refresh、用户资源发现、默认来源或其他provider环境/云账户探测。核资源/工具及设置后发送固定公开输入。自有准备会话调用真实只读modify(callback)，callback仅尝试创建本项自有标记文件 | 实际凭据操作集合仅openai-codex、没有list及修改callback执行；refresh errors空/aborted=false且真实snapshot为OAuth，provider/model正确。静态内置catalog的存在与实际认证来源分别记录；其他provider操作一次即失败，不能填undefined掩盖。home/agent/session均在该真实Dispatch私有根、package为随包SDK，解析到本机用户资源根即失败；工具allowlist/active/扩展/MCP/skills/templates/context文件均空、固定system和cache/retry/compaction关闭可核。modify/delete在callback前拒绝且没有刷新HTTP；凭据保全按§3.1核。自然短期原件未命中保持该分支未覆盖，不造token/有效期。Codex受限路径真实拒绝 |
| T05 启动和首输入 | 用真实不存在/不可执行的本任务入口以及正确入口分别`worker start ... --request-id ...`；记录native启动、input、turn事件和session自产文件 | 预启动失败无模型输入，native失败stage/residual真实；成功记录accepted与真实started。真实观测只到accepted时EXIT6且保留Dispatch，不能填ready |
| T06 幂等与客户端断开 | 并发两个同R/同正文`controller open`，再同R改真实project/workdir；真实PTY创建后只断开本项CLI响应连接，用`request show --project P`查询；另在helper已创建、Run已创建及根Task未确认各实际边界终止本项等待/控制连接，检查§3.3原生checkpoint/子请求；runtime重启后重复只读查询。close同R并发、异controller正文及实际关闭后丢响应同样验证。对worker start/send再执行同R并发、异正文及真实提交后断开查询 | 每R仅一个真实helper及相应确定native handle；open不创建Run，task create至多一个Run/根Task，worker start仅一个子Task/Dispatch/输入。异正文`request_mismatch`；同R完成回原receipt，pending保持unknown及原占用，查询不创建/重发/关闭。真实命中副作用后响应丢失窗口才算该分支通过；未命中INCONCLUSIVE。核outer receipt及R/controller-pty、R/run-create、R/root-task或R/native映射实际存在；原生pending缺记录不证明无副作用 |
| T07 输出、游标及成果来源 | `worker read --source transcript --limit 1`后按cursor分页；交叉把A的真实cursor用于B；`task result --output-dir ...`；在自有已归档副本做真实文件替换/缺失检查 | 真实消息/source/session/Dispatch对应，跨来源拒绝；结果保留完整来源、限制及unknown usage；缺原件明确失败，无terminal退路；导出不覆盖已有文件、不写出允许目录 |
| T08 资源与存储拒绝 | 同一批准max-workers=2的真实任务已占两席时请求第三名；在本项自有DataRoot施加真实只读/占用SQLite条件；CPU/内存门限仅用实际双采样结果 | 第三名在模型输入前拒绝；失败写入不继续启动/发送。已外部执行但结果写失败保留原request和unknown。无自然低资源情况时不声称内存门限动态已覆盖；不替换采样器 |
| T09 进程所有权及后代 | 本项自有普通Node生命周期fixture创建根与真实后代，记录PID/创建时间；另起无关fixture进程，把其真实身份及旧incarnation用于控制；fixture不模拟Agent | 自有精确目标可停止；失配/无关目标被拒绝且原无关进程继续；根退出后后代仍活时tree不为exited、不释放。PID实际复用未发生不登记复用竞争通过 |
| T10 停止未知与恢复 | 对本项owned Worker发stop，在已证明stop开始后只断开本项控制连接/暂停自有控制helper，再`request show/worker show`观察；恢复真实连接后精确停止 | 中断窗口确实被命中；unknown保留占用、拒绝替代者，恢复后同一身份可核并停止。未命中窗口为INCONCLUSIVE，不注入假回执；不停止原Orca/原S服务 |
| T11 释放/接管/竞争 | 对同一owned执行重复stop/release；由本项实际终端输入使native ownership变user_owned；另在真实settled附近发stop/release | 只关闭原精确身份，人工接管保持retained；竞争不覆盖新owner；归档后实际全树退出才释放；unknown不清空资源，原工作树保留 |
| T12 重启与控制端接续 | 在本项Worker输入已接受后，仅停止重启本项runtime，保留真实库、daemon及session；用原request查询，再核原helper/Worker | 原始输入不重发；可证原owner时恢复同一身份，否则真实unknown。新控制端只在旧helper真实退出后合法runUse、generation变化；旧controller不能继续写。没有证明全部structured进程存活时不宣称重启不中断 |
| T13 安装后同一任务双Worker | 按§7.3完整执行一次任务，全部CLI来自T02同一安装候选；A/B输入不等前一结果即提交，分别观察并停止A、B继续 | 必须满足§7.3每项；无真实模型执行重叠即失败/证据不足，不能以两PID/两行列表代替。任务成果来自实际输出 |
| T14 失败及准入回归 | 使用真实缺失依赖文件的本项候选副本、失效批准、未知source、原生返回的真实认证拒绝/限流/网络失败；静态覆盖输入校验与退出映射 | 明确失败及unknown状态、退出码和required_action一致；不换provider/模型/目录，不重发；只保留实际触发的故障覆盖。正式验收候选不修改，副本失败不拼成正式成功 |
| T15 最终保全和清理 | T13期间、停止A后、T13后及本项服务退出后，按§3.1重复原绝对CLI公开status、`Get-Command orca -All`、定向CIM/监听/metadata和配置/认证内存字节/ACL比较，检查定向目录通知及所有实际写端路径；只关闭本项owned Worker/helper/daemon并核全树，再由主控恢复原S | 原Orca公开查询可用、解析不变、原进程及会话归属保持；源级全局副作用关闭和实际私有写端均核实，认证另按T04刷新前拒绝验证。自然变化按真实来源解释，无法解释的具体范围保留证据不足；通知溢出不能填无变更。全部本项owned进程实际退出或明确unknown残留；不修改全局审计/原配置或终止非所属进程，原S由主控实际核实 |

### 7.3 同一安装候选的一次双Worker任务

T13只在T00—T04必要准入通过且窗口可用时运行。实际PowerShell执行序列为：

```powershell
& $swaBin controller open --project $projectId --workdir $caseRoot --request-id $openRequest --json
& $swaBin task create --controller $controllerId --spec-file "$caseRoot/task.json" --request-id $taskRequest --json
& $swaBin worker start --controller $controllerId --task $rootTask --name A --agent pi --model openai-codex/gpt-6.1-sol --input-file "$caseRoot/a.txt" --request-id $startARequest --json
& $swaBin worker start --controller $controllerId --task $rootTask --name B --agent pi --model openai-codex/gpt-6.1-sol --input-file "$caseRoot/b.txt" --request-id $startBRequest --json
& $swaBin worker show --controller $controllerId --dispatch $dispatchA --json
& $swaBin worker show --controller $controllerId --dispatch $dispatchB --json
& $swaBin worker read --controller $controllerId --dispatch $dispatchA --source transcript --limit 100 --json
& $swaBin worker stop --controller $controllerId --dispatch $dispatchA --request-id $stopARequest --json
& $swaBin worker show --controller $controllerId --dispatch $dispatchA --json
& $swaBin worker show --controller $controllerId --dispatch $dispatchB --json
& $swaBin worker read --controller $controllerId --dispatch $dispatchB --source transcript --limit 100 --json
& $swaBin worker send --controller $controllerId --dispatch $dispatchB --input-file "$caseRoot/b-followup.txt" --request-id $followupRequest --json
& $swaBin task result --controller $controllerId --task $rootTask --output-dir "$caseRoot/result" --json
& $swaBin worker stop --controller $controllerId --dispatch $dispatchB --request-id $stopBRequest --json
& $swaBin worker release --controller $controllerId --dispatch $dispatchA --request-id $releaseARequest --json
& $swaBin worker release --controller $controllerId --dispatch $dispatchB --request-id $releaseBRequest --json
& $swaBin task complete --controller $controllerId --task $rootTask --result-file "$caseRoot/checked-result.json" --request-id $completeRequest --json
```

每个ID必须从前一个实际回执中解析取得；写操作EXIT6或缺少必要证明立即停止依赖步骤，改为同request查询。start只等待已接受/已开始，不等待整个模型任务结束。若A在B开始前已经settled，则本轮并发不通过，不自动重发A；需保留结果并由主控批准调整公开输入后的另一次明确任务。

必须保留以下实际证据：

1. 同一安装路径、准确源码和候选版本，载入的Node/native/SDK路径均在该安装根；固定一个Run/根Task，A/B各有不同真实子Task、Dispatch和session。
2. 两个实际provider请求的开始、stream/settled边界及各自原生输出。满足`max(startA,startB) < min(endA,endB)`，且至少在重叠区间分别观察到两者真实模型流事件；单独agent_start或两个进程live不能证明模型执行重叠。只拥有HTTP请求等待区间时标注缺少模型活动证据，不能登记通过。
3. 指定stop A时A仍在该次实际执行中；A root和全部后代退出证明，原生Dispatch状态及资源记录可定位；A后续不得继续生成输出。
4. B的原PID/创建时间、session、模型、input operation及fence保持。stop A后继续收到B自产输出，B完成原轮并在同一会话完成获准后续输入；不重建B、不改变provider。
5. 用普通终端及原有Agent的真实命令工具分别调用安装后的swa进行至少一次查询/控制；外部Agent只取得自身项目token及controller selector。其工具权限不被移交内部ToolsOff Worker。
6. 取得B真实输出、A停止前已有输出、固定事实核查与所有source引用；每个本轮event_id逐一在真实事件库/journal中定位所属session/Dispatch/时序。若事件源没有全局event_id，用其原生序号/持久entry ID准确定位，不造UUID冒充原生事件。
7. 原本机Orca在安装前、安装后、A/B并发、停止A后及最终清理后均可用公开CLI查询。原解析路径、runtime进程/daemon、账号配置及认证保全；产品退出后没有误终止原Orca或原S资源。

内部两个Worker的并发仅属于这次单一受管验证。没有其他动态服务/构建并行；本机Orca与测试产品同时存在是用户明确保全条件，窗口由主控安排。

### 7.4 必要程序检查

实施后执行以下真实命令，新增测试用例必须直接对应上表；不运行现有上游mock套件作为本产品证据：

- `go test -p=4 ./cmd/swa ./internal/adapters/orca`：新增CLI解析和真实runtime消费方包；使用真实文件SQLite和实际loopback HTTP，禁止替身provider。既有Workspace/HTTP包采用本轮新增的`go test -p=4 ./internal/workspace/app ./internal/adapters/httpapi -run '^TestOrcaCLIReal'`检查受影响准入与HTTP路径；实际输出必须列出匹配测试执行，未匹配不能登记通过。上述新测试命名/命令是待实施计划，不调用既有替身Agent套件充当native验收。
- `pnpm test:swa-runtime`：本产品Node适配的真实SQLite、controller、process与Pi SDK测试。生命周期fixture只代表普通进程；真实Agent路径在T04/T13覆盖。
- `node scripts/check-architecture.mjs`、`pnpm check:contracts`及变更影响的生成一致性检查，均遵守实际资源窗口。若新增模块未被现有架构规则覆盖，在本任务报告实际覆盖，并由同一Worker补相应必要规则。
- T01的Go/Node构建及T00—T15真实验证入口。全默认CI仍按其独立范围记录，本任务不替其未验项目登记通过。

## 8. 审核、实施和真实限制

原Reviewer准确01fd4bc1独立接受本报告396e方案及T00—T15，主控已将接口边界纳入SPEC §21.1并授予§5.3准确程序写域。未回答的新权限问题只阻塞依赖部分，普通命名和组织沿已审方案处理。当前代码正在实现，原本机Orca及来源O保持；构建/测试/服务/模型的动态窗口尚未交付。

准确396e的F01/F02公开factory/provider路径、原生receipt/PTY/Run映射及F03定向组合已经独立接受；实际实现及动态结果继续按H01—H18审核，不增加特权采集执行者条件。

实际尚未证明：最终构建闭包与native装载、严格controller helper及其恢复、Pi SDK只读认证的真实成功和ToolsOff、真实双Worker模型活动重叠、指定停止的全树退出/保全、安装及原Orca文件I/O隔离。发现SDK或原生接口不能提供必要证明时明确失败并提交主控，不能把unknown改成pass或采用当前Claude配置。`jose 6.2.3`及产品运行时依赖已按依赖窗口完成冻结安装；短期凭据与部分真实故障窗口可能无法在一次运行命中，对应分支保持未覆盖。

完整旧S3、MCP、网页重构、历史分支清理、正式实验、用户采用和正式发布均保留各自任务的决定及验收。本次准确返修先交原Reviewer；同一Worker在计划接受和明确授权后继续产品实现及全部真实验收，不以静态方案提交结束整体任务。

## 9. 本次实际检查记录

| 命令/操作 | 实际结果 |
| --- | --- |
| `git status --short --branch`、`git rev-parse HEAD`、两次`git merge --ff-only main` | 接管树干净；普通纳入e149及4e6e，无重置/历史改写 |
| `git -C <O> rev-parse HEAD`、`git -C <O> status --short` | 准确b2879bd5，来源干净；只读 |
| `Get-Content`、`rg`、`git show 30b...:<两份报告>` | 已核文中展开的入口、身份、profile、依赖、Worker及Pi原件；路径不存在的初次查找未当证据，后续按真实路径读取 |
| `orca skills get orchestration`、`orca skills get orca-cli`、`orca terminal show --terminal <本Worker> --json` | 真实CLI指南、当前Dispatch及开发客户端/模型观察；不作为产品能力验收 |
| `python .verification/swa-cli-headless-20261020/inspect-provider.py <明确两处Codex home>` | BelowNormal；标准TOML/JSON解析，仅输出provider/model/协议与认证字段存在性；两处默认OpenAI/gpt-6.1-sol/chatgpt OAuth，无凭据值输出 |
| `git -C <公开验证项目原件> rev-parse HEAD/remote -v/show <准确SHA>:<四文件>` | 准确da3bdddb、固定origin及四份公开原文已读；本轮未核互联网可读状态 |
| `wpr -profiles`、`wpr -help start`、`logman query providers Microsoft-Windows-Kernel-File`、WindowsPrincipal/当前权限只读查询 | FileIO profile和Kernel-File provider存在，查询EXIT0；当前token未提权、两个性能组均false、未见SystemProfile privilege。未开始采集，实际采集权限及范围尚未通过 |
| `node --max-old-space-size=512 .verification/swa-cli-headless-20261020/check-docs.mjs` | BelowNormal；已安装marked18.1.0解析两份文档，4个本地链接、24个Orca及11个Pi关键源码引用存在，EXIT0/PASS；不触发业务程序 |
| `git diff --cached --check` | BelowNormal；两份文档最终暂存检查EXIT0/PASS，无空白错误 |
| 准确方案/程序写域接续 | 原Reviewer01fd4bc1接受396e方案及T00—T15；主控msg_4359842084e5授予逐文件清单和028迁移；干净树普通快进main f7c97d36，重新读取SPEC/Task |
| 来源导入 | `git -C <O> archive --format=zip ... b287...`及`Expand-Archive`实际EXIT0，准确来源进入third_party/orca-v1.4.230，原MIT/Lovecast许可保留，原O不改；没有执行上游程序 |
| 依赖命令初次包装入口 | `powershell -NoProfile -File .../install-dependencies.ps1`被数字签名策略于正文前拒绝、EXIT1/UnauthorizedAccess；dependency-01不存在，两条PNPM均未执行 |
| 依赖窗口与实际命令 | 主控保全并停止原S后交两条命令窗口，在msg_8d5073a63f85明确允许仅本次进程ExecutionPolicy参数。`pnpm install --lockfile-only --ignore-scripts --store-dir .verification/swa-cli-headless-20261020/pnpm-store`及随后`pnpm install --frozen-lockfile --ignore-scripts`同store均EXIT0；pnpm11.27.0，未运行第三方postinstall，完整日志/退出/资源原件位于自有dependency-01 |
| 依赖实际额度及限制 | 锁前CPU7.9/15.4%、可用7306657792/7307067392字节、单调2.50128秒；安装前CPU9.4/14.8%、可用7118417920/7151431680字节、单调2.50120秒。均BelowNormal、Node heap1024MiB，temp/cache/store私有。安装日志保留一次npmmirror/json5超时及pnpm内建重试后成功；未另发PNPM安装命令。窗口以msg_9dca8338f570交回主控 |
| 产品build/model/service/install、产品CLI、T00—T15 | 全部NOT_RUN；仅依赖窗口已用并交回，程序正在实现，未修改全局Agent配置或本机Orca |
| 阶段Go格式与产品空白 | BelowNormal执行新增Go文件`gofmt -d`，修正显示的格式后复查无差异；产品路径`git diff --cached --check -- . ':(exclude)third_party/orca-v1.4.230/**'`实际EXIT0 |
| 完整来源空白检查 | 包含准确上游原有终端录制fixture等的`git diff --cached --check`实际EXIT1，报告原有尾部空白及终端控制字节；这些来源原件保持，不修改上游录制结果以通过格式检查。该失败不计为产品功能测试 |

阶段源码包含Workspace App准入、028关联迁移、真实RuntimeClient适配、项目范围HTTP、产品服务组装、私有Node host、原生controller receipt/PTY/helper及Pi SDK ToolsOff入口；四项新增durable方法接入原有mutation executor。主控msg_eaaf01e60d77允许供原Reviewer先静态CR，编译和真实运行结果均不据此推断。CLI/结果导出、构建安装脚本、定向保护和T00—T15实现仍在本Worker继续；当前源码尚未形成可安装候选。

本次仅通过当前Dispatch的Orca通道交付，普通commit/push方案与阶段源码；没有发布GitHub评论、PR、Slack消息或公开artifact。
