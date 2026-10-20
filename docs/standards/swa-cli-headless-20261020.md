# Swarm CLI隔离后台、身份、控制规则与真实测试计划

对应[Worker任务](../../tasks/swa-cli-headless-20261020.md)，接管基线main `e149510762f89bed06e7c8c766917554a58294e4`，后续普通快进至 `4e6e8ed4fa52521c2232dd770eb98297d8713785`。当前为 **设计提案和待审核测试计划**；没有程序实现、安装、构建、模型或功能验收结果。

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

拟定Windows安装布局为`<InstallRoot>/bin/swa.exe`、`<InstallRoot>/lib/swa-runtime/`和`<InstallRoot>/runtimes/node-<原生版本标识>/node.exe`；可写数据显式使用`<DataRoot>`，不得位于本机Orca根、其父目录、账号根或现有项目根。安装器使用真实路径、reparse point及ACL核查拒绝重合/嵌套共享，不以路径字符串不同推断隔离。目录及名字属于本任务实施细节，当前为待审核设计。

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

保全检查比较明确配置文件和认证文件的字节、ACL及持有进程，不遍历私人会话正文。运行中原Orca可以合法写入自身日志/状态，不能要求整棵活跃目录逐字节不变；需要记录自然写入并证明产品的实际文件I/O只在获准范围。原客户端自然刷新认证时，还须从实际writer及进程身份归因，证明产品无写入/刷新；无法归因即证据不足，不能仅凭差异将本机状态改回旧值。动态时使用本项目进程的Windows文件/进程追踪，认证只记录比较结果，不把原件放进报告；缺少可用追踪权限/工具时保留相应隔离证据未验，不自行提权或安装系统驱动。

### 3.2 从普通终端或原有Agent取得真实控制身份

原生`runCreate`要求stable pane/session，普通外部终端不能凭空取得该身份。方案增加一个**实际运行的控制helper**，只承担已有命令的传输和身份绑定，无模型或自动调度循环：

1. `swa controller open`通过现有项目scoped token向Go App请求，重查项目、批准及允许动作。后台在私有runtime中创建真实后台PTY，启动随包`node swa-controller.js`；保存原生terminal/pane/process incarnation及实际PID/创建时间。该内部创建传入真实`launchConfig`，其`agentCommand/agentArgs/agentEnv`仅描述随包helper命令，不传`startupAgent/launchAgent`。`O:src/main/agent-launch/agent-launch-token.ts:agentLaunchTokenFor`仅在launchConfig存在时mint token，`terminal/launch/terminal-create.ts`负责环境注入和PTY commitment；普通空壳terminal没有这项token，不能当控制身份。此helper不登记成模型Agent，不伪造hook状态。
2. helper从自己原生launch环境读取宿主产生的真实token/handle和产品metadata，经RuntimeClient发起请求。用户命令不接受`--from`、pane、consumer、actor或launch token；控制helper只以自己的实际身份调用runCreate/runUse和原生任务/Worker方法。
3. 外部终端和已有Agent使用同一`controller`选择器及原有项目token。后台在每次写入核准主体、项目、批准、root Task、controller的完整活身份后，经原生terminal输入通道交付命令；helper只接收版本化JSON动作，由成熟JSON解析器解析，不执行调用者传入的shell字符串。命令结果使用原生request回执及私有终端输出返回；helper传输自身不另建持久队列或任务表。
4. 严格产品admission调用原生精确launch校验，不能因上游核验返回null就沿宽松declared-handle分支放行。helper的输入接受与请求执行分别记录；传输未知时按原request ID查询，不再次terminal.send。真实收到的correlation ID用于定位，不充当权限凭据。
5. 本机Orca内的原有Agent调用swa时，其ORCA环境属于另一个宿主。产品入口只用产品controller，不把该环境转换成产品Worker或human；产品子进程清除外部Orca身份、pairing、account及remote选择变量，再注入自己的真实环境。父进程环境不改动。
6. runtime重启后，能核实原生恢复证明的原helper才继续；证明不足返回`controller_unverifiable`，只读查询仍可用。确认原helper真实退出后才允许创建新helper并由它调用原生runUse推进consumer generation。未知时不重绑、不创建竞争控制端。

控制helper是需要实现并真实验证的适配部分，不能由当前开发终端的身份代替。本地owner token与同一Windows用户的权限边界仍属协作式控制；不宣称阻止恶意同用户读取其可读凭据。应用层项目/批准检查不因这项边界而省略。

## 4. 当前Agent与非Aliyun受限正向方案

### 4.1 实际静态观察

| Agent | 本次观察 | 首版适用状态 |
| --- | --- | --- |
| Codex | 已安装0.162.0，当前开发确实运行gpt-6.1-sol。`~/.codex`及当前Orca account home的TOML/JSON白名单解析均为默认OpenAI、chatgpt OAuth、有tokens、无API key | 当前公开原生控制器不能可靠关闭全部原生文件工具；维持受限拒绝。开发会话不充当ToolsOff证据 |
| Claude | 已安装2.1.238；JSON白名单观察到qwen3.7-max与dashscope来源，且含hooks/statusLine | 主控明确禁止调用该配置；不修改全局设置、不自动选择其他模型 |
| Pi | 已安装1.0.1，系统settings只记录版本，系统auth文件为2字节。真实Pi包包含openai-codex provider、chatgpt.com/backend-api、gpt-6.1-sol和原生RPC/SDK | 系统Pi当前认证成功未知。产品私有Pi SDK路径具静态兼容依据，待真实验证 |

只读检查没有输出认证值、账号ID或私有对话。当前Orca终端`terminal show`仅证明实际开发客户端；没有把它作为产品执行或工具准入结果。

### 4.2 复用Pi原生provider和只读OAuth

准确接口来自`P:dist/index.js`导出的`ModelRuntime`、`createAgentSessionServices`、`createAgentSessionFromServices`、`createAgentSessionRuntime`、`SessionManager`、`SettingsManager`、`runRpcMode`，以及对应`.d.ts`、`docs/sdk.md`、`examples/sdk/13-session-runtime.ts`。采用产品`Pi SDK 1.0.1`入口，调用真正的`runRpcMode`；不仿造pi CLI、不手写Pi原生协议或模型HTTP。

`P:node_modules/@earendil-works/pi-ai/dist/providers/openai-codex.js`使用原生`openai-codex-responses` API及`https://chatgpt.com/backend-api`；其catalog文件`providers/data/openai-codex.json`包含`gpt-6.1-sol`。OAuth来源`auth/oauth/openai-codex.js`具有真实OAuth flow与toAuth，使用access bearer；产品不变更endpoint、provider、账号或模型，未通过就返回具体错误。

只读CredentialStore是该SDK明确允许注入的基础设施端口，执行以下规则：

1. 仅从明确选择的当前Codex account home读取原JSON，并使用成熟`jose 6.2.3`解析JWT有效期和必要标准声明，禁止手写split/base64 parser；该版本和integrity已在准确Orca的`pnpm-lock.yaml`找到，复用其pin并列为直接实施依赖。TOML使用`smol-toml`，JSON使用标准解析器。JWT解码只解释已获准原件中的过期字段，不替代提供方的签名/认证判断。
2. `read/list`仅向同一`openai-codex` provider返回真实OAuth字段及非秘密元数据。原access/refresh/account字段保持来源，expires来自真实标准声明；不伪造有效期，不把OAuth改称API key，不创建可写认证副本。
3. 每次read先核真实有效期足以覆盖本次最长请求和SDK五分钟刷新界限；不足或无法解释即在模型请求前拒绝。read按当前原件取快照，不刷新、不撤销、不写回本机登录。
4. `modify/delete`立即抛出明确只读错误，**不执行传入的callback**。原生`auth/resolve.js:resolveStoredOAuth`只在`credentials.modify`的callback中调用`oauth.refresh`；`models.js:resolveRefreshCredential`亦如此。两条路径都在网络刷新之前被拒绝；不能先refresh再阻止保存。
5. `ModelRuntime.create({credentials, modelsPath:null, allowModelNetwork:false, refreshOnCreate:false})`使用真实已安装catalog；禁用启动时所有provider的catalog/auth枚举及网络刷新。显式`getModel('openai-codex','gpt-6.1-sol')`取得原生对象，随后仅对该模型调用getAuth；不枚举其他provider的环境/云账户。产品Pi options适配读取真实get_state及这个固定catalog项，不调用原实现全量get_available_models。没有有效模型/认证时立即失败，拒绝SDK默认选其他model的结果。
6. 原凭据自然被原有Codex更新时，下次read校验同一来源与当前授权。产品不写原auth，不复制刷新权；解析错误保持失败。后续产品长期登录能力另行设计，本次可通过重新读取仍由原客户端正常管理的当前有效access继续获准固定任务。

### 4.3 工具、扩展、输入和实际完成

创建服务时显式私有agentDir；`SettingsManager.inMemory`提供固定设置，packages和各资源路径为空、`defaultProjectTrust:'never'`且项目trusted=false；`resourceLoaderOptions`将`noExtensions/noSkills/noPromptTemplates/noContextFiles/noThemes`均设为true，未传入任何additional资源或extensionFactories。直接提供非空固定systemPrompt及`appendSystemPrompt:[]`，在原生discovery分支之前阻止SYSTEM/APPEND文件发现；不使用加载后才执行的extensionsOverride等过滤器承担准入。只读公开项目输入由Workspace固定提供，不继承项目本地配置。

`createAgentSessionFromServices`传入`noTools:'all'`、`tools:[]`、`customTools:[]`。源码`dist/core/sdk.js`把此组合转换为空allowlist及空active tools；`DefaultResourceLoader`不加载扩展/MCP。启动与每次输入前检查原生`session.getActiveToolNames()`为空、注册工具集合及resource列表为空，并从实际请求前原生session状态核对。仅零工具调用次数不能通过该检查。

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

建议后续由同一Worker修改`cmd/swa/**`、Workspace消费方端口/现有HTTP组装的必要文件、`internal/adapters/orca/**`、明确的Orca来源目录与产品Nodecomposition、对应契约、构建/验证脚本和本任务文档。主控须在计划接受后授予准确文件写域，当前未获授权。

构建使用`go build`现有产品及新增swa、锁定Node构建工具和真实native；安装只写明确InstallRoot/DataRoot，不替换系统orca或Agent，不注册自动启动、不改变用户/系统PATH，不下载私有profile。普通终端通过swa绝对路径调用；如用户另行安装PATH入口，仅新增swa且需专门授权，不影响orca解析。

现有Go HTTP及私有Node/daemon作为同一产品服务组被主控安排在唯一动态窗口内，固定所属关系。原S服务只能由主控保全、停止并在测试后恢复；本Worker本次未停止任何已有服务。构建和安装完成后才启动服务组，过程中不并行运行第二组构建或服务。

## 6. 最少完整CLI规则

以下为待审核的精确拟定命令，不是当前可执行功能。变更后CLI/HTTP/App共用同一准入语义；所有写命令要求显式`--request-id <UUID>`，原生mutation receipt保留该键。上下文通过UTF-8文件传入；禁止在命令行正文中放凭据。

| 命令 | 输入与实际结果 / 复用 |
| --- | --- |
| `swa serve --data-dir D --port N --json` | 必须明确私有根和端口；打印真实readiness/版本/路径/身份的非秘密部分。缺依赖、锁冲突或端口占用失败，禁止改用其他runtime/端口 |
| `swa status --json`、`swa capabilities --json` | 原生runtimeId/startedAt、数据库authority、daemon/provider支持/当前证据与限制；读操作不调用模型、不隐式启动后台 |
| `swa controller open --project P --workdir W --request-id R --json` | 创建§3真实helper；验证当前项目/批准/写域，回controller selector及真实native身份。输入的P必须是实际Workspace ID |
| `swa controller show --controller C --json`、`close ... --request-id R` | 精确观察/关闭自有helper；close需所属全部Worker已退出/释放，否则拒绝。不得用terminal全量关闭 |
| `swa task create --controller C --spec-file F --request-id R --json` | 在真实Run中创建根Task；F固定目标、公开来源版本、scope与成功条件，返回runId/root taskId及已有批准引用 |
| `swa task show --controller C --task T --json`、`list --controller C` | 原生根/子Task/Dispatch及只读汇总；原请求未决、启动/停止未知和资源占用分别展示 |
| `swa worker start --controller C --task T --name A --agent pi --model openai-codex/gpt-6.1-sol --input-file F --request-id R --json` | 新建T的子Task及原生Worker Dispatch，以真实Pi SDK执行固定公开输入；ready只在真实turn开始证明满足时返回，接受但未证明开始保持unknown |
| `swa worker show --controller C --dispatch X --json`、`list --controller C --task T` | 精确Dispatch、子Task、runtime/session/fence、PID/创建时间、Agent实际状态和root/tree exit；不只输出PTY是否live |
| `swa worker read --controller C --dispatch X --source transcript --limit 100 --json` | 读取native journal/provider自产会话的实际消息、真实cursor/source identity；无原件即缺证。后续分页必须使用原cursor，来源变化拒绝 |
| `swa worker send --controller C --dispatch X --input-file F --request-id R --json` | 验完整当前绑定及授权后向实际原session发送；保留input accepted、turn started与settled；不把写入stdin当开始 |
| `swa worker stop --controller C --dispatch X --request-id R --json` | 精确停止该owned Worker，返回stop request、实际root/tree exit及保留占用；PID/创建时间或ownership失配时不终止 |
| `swa worker release --controller C --dispatch X --request-id R --json` | 只对原生允许结算状态执行冻结输出与释放；未知、人工接管、缺归档或后代未退出时不释放。重复同键只回原receipt |
| `swa task result --controller C --task T --output-dir O --json` | 提取根及各Worker真实输出、来源、检查引用、版本和已知usage；输出路径须不存在或由本命令精确拥有，禁止覆盖其他文件 |
| `swa task complete --controller C --task T --result-file F --request-id R --json` | 当前控制端确认固定成功条件和真实结果，全部相关Worker退出/归档后调用原生taskUpdate；失败或unknown不能完成；完成不表示用户采用或Git发布 |
| `swa request show --controller C --request-id R --json` | 查原生mutation receipt及原run/task/dispatch/input operation，不生成新请求。无receipt仍不能单独证明未发生外部输入 |

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
| T02 安装与私有隔离 | 使用T01同一候选执行产品`install.ps1 -InstallRoot ... -DataRoot ...`；绝对路径运行`$swaBin --version`和`serve --data-dir ... --port ... --json`；另对已占用真实端口/原Orca根尝试启动 | 指定端口冲突EXIT5且无另选端口；重合/reparse共享路径拒绝；私有profile-state.db/orchestration.db/pipe/daemon/key正确。安装前后原orca解析/进程/配置及公开status可用；不启动Electron/自动启动或全局hook修改 |
| T03 控制身份及权限 | `controller open/show`及`task create`；使用真实过期/撤销或缺失项目token、另一个真实项目/Run/旧controller访问；尝试CLI `--from`/human参数 | 仅真实helper及当前consumer可写；未知flag EXIT2，无效主体EXIT3；无headerless bootstrap调用、human cookie或伪consumer；Controller PID/创建时间和真实native pane可查；旧代请求拒绝 |
| T04 Pi认证及ToolsOff | `capabilities`；启动无用户prompt的Pi SDK准备，读真实model/资源/工具状态；随后一个固定公开输入。使用本项自有SDK准备会话调用真实只读CredentialStore.modify(callback)，callback仅尝试创建本项自有标记文件；必须在callback执行前拒绝，原生调用顺序审查另核refresh只在该callback内部 | provider/model为原生openai-codex/gpt-6.1-sol、工具allowlist/active/扩展/MCP/skills/templates/context文件均空、固定system输入可核。modify/delete在callback执行前拒绝且没有刷新HTTP；实际凭据字节未变。短期原凭据只能等待自然进入实际界限，未命中则该分支未覆盖；不造token/有效期，不因测试有缺陷触发真实refresh。Codex受限路径真实拒绝 |
| T05 启动和首输入 | 用真实不存在/不可执行的本任务入口以及正确入口分别`worker start ... --request-id ...`；记录native启动、input、turn事件和session自产文件 | 预启动失败无模型输入，native失败stage/residual真实；成功记录accepted与真实started。真实观测只到accepted时EXIT6且保留Dispatch，不能填ready |
| T06 幂等与客户端断开 | 同时发两个相同R/正文的真实请求，再同R改正文；实际提交后只终止本项CLI等待进程，`request show`/`worker show/read`查询 | 只有一份原生Task/Dispatch/对应输入；不同正文冲突。CLI退出后同一原生会话继续。没有receipt不能作为安全重发证据；未知不换键/不第二次输入 |
| T07 输出、游标及成果来源 | `worker read --source transcript --limit 1`后按cursor分页；交叉把A的真实cursor用于B；`task result --output-dir ...`；在自有已归档副本做真实文件替换/缺失检查 | 真实消息/source/session/Dispatch对应，跨来源拒绝；结果保留完整来源、限制及unknown usage；缺原件明确失败，无terminal退路；导出不覆盖已有文件、不写出允许目录 |
| T08 资源与存储拒绝 | 同一批准max-workers=2的真实任务已占两席时请求第三名；在本项自有DataRoot施加真实只读/占用SQLite条件；CPU/内存门限仅用实际双采样结果 | 第三名在模型输入前拒绝；失败写入不继续启动/发送。已外部执行但结果写失败保留原request和unknown。无自然低资源情况时不声称内存门限动态已覆盖；不替换采样器 |
| T09 进程所有权及后代 | 本项自有普通Node生命周期fixture创建根与真实后代，记录PID/创建时间；另起无关fixture进程，把其真实身份及旧incarnation用于控制；fixture不模拟Agent | 自有精确目标可停止；失配/无关目标被拒绝且原无关进程继续；根退出后后代仍活时tree不为exited、不释放。PID实际复用未发生不登记复用竞争通过 |
| T10 停止未知与恢复 | 对本项owned Worker发stop，在已证明stop开始后只断开本项控制连接/暂停自有控制helper，再`request show/worker show`观察；恢复真实连接后精确停止 | 中断窗口确实被命中；unknown保留占用、拒绝替代者，恢复后同一身份可核并停止。未命中窗口为INCONCLUSIVE，不注入假回执；不停止原Orca/原S服务 |
| T11 释放/接管/竞争 | 对同一owned执行重复stop/release；由本项实际终端输入使native ownership变user_owned；另在真实settled附近发stop/release | 只关闭原精确身份，人工接管保持retained；竞争不覆盖新owner；归档后实际全树退出才释放；unknown不清空资源，原工作树保留 |
| T12 重启与控制端接续 | 在本项Worker输入已接受后，仅停止重启本项runtime，保留真实库、daemon及session；用原request查询，再核原helper/Worker | 原始输入不重发；可证原owner时恢复同一身份，否则真实unknown。新控制端只在旧helper真实退出后合法runUse、generation变化；旧controller不能继续写。没有证明全部structured进程存活时不宣称重启不中断 |
| T13 安装后同一任务双Worker | 按§7.3完整执行一次任务，全部CLI来自T02同一安装候选；A/B输入不等前一结果即提交，分别观察并停止A、B继续 | 必须满足§7.3每项；无真实模型执行重叠即失败/证据不足，不能以两PID/两行列表代替。任务成果来自实际输出 |
| T14 失败及准入回归 | 使用真实缺失依赖文件的本项候选副本、失效批准、未知source、原生返回的真实认证拒绝/限流/网络失败；静态覆盖输入校验与退出映射 | 明确失败及unknown状态、退出码和required_action一致；不换provider/模型/目录，不重发；只保留实际触发的故障覆盖。正式验收候选不修改，副本失败不拼成正式成功 |
| T15 最终保全和清理 | T13后及本项服务退出后重复T00原Orca公开CLI/解析/配置/认证/进程/监听观察；关闭本项owned Worker/helper/daemon，核全树，再由主控恢复原S | 原Orca仍可公开CLI查询、原进程归属和配置保持；产品无全局写入/认证刷新。全部本项owned进程实际退出或明确报告unknown残留；不删除未知文件/进程，原S恢复由主控实际核实 |

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

当前待独立审核的事项是本报告的准确方案与T00—T15；主控已批准唯一原生账本及Pi静态调查方向，未授予程序写入或动态窗口。未回答的新权限问题只阻塞依赖部分，普通命名和组织沿此方案处理。没有把后台安装归属旧选项保留成整体阻塞。

主控要求提前核查文件追踪的现成条件，本次只读结果为：系统自带`C:/Windows/system32/wpr.exe`、`logman.exe`、`wevtutil.exe`存在；`wpr -profiles`列出FileIO，`logman query providers Microsoft-Windows-Kernel-File`返回真实provider及WRITE/DELETE/RENAME等keywords、EXIT0。当前Windows token的ElevatedAdministrator=false，PerformanceLogUsers/PerformanceMonitorUsers均false，当前权限列表未发现SeSystemProfilePrivilege；未发现Procmon可调用入口或进程。工具存在和provider查询成功不能证明当前token可以采集，未启动WPR/ETW，未提权或安装驱动。

因此T02/T15所需隔离证据在动态开始前仍有明确准入条件：主控须安排获准且实际具备权限的采集执行者，并先核采集范围、现有trace会话及产品PID/创建时间过滤；当前Worker token没有实际采集证明，不能直接把全机FileIO默认采集作为获准范围。使用WPR时需要独立instance、专属recordtempto和输出根，只停止自己创建的trace，不能取消其他会话。没有满足这些条件时T02/T15必要隔离验收保持阻塞，不能豁免、降低标准或以配置快照单项代替全部证据；主控可继续独立审核及其他不依赖该条件的工作。

实际尚未证明：最终构建闭包与native装载、严格controller helper及其恢复、Pi SDK只读认证的真实成功和ToolsOff、真实双Worker模型活动重叠、指定停止的全树退出/保全、安装及原Orca文件I/O隔离。发现SDK或原生接口不能提供必要证明时明确失败并提交主控，不能把unknown改成pass或采用当前Claude配置。`jose 6.2.3`来源pin已找到，产品依赖尚未安装；短期凭据与部分真实故障窗口可能无法在一次运行命中，对应分支保持未覆盖。

完整旧S3、MCP、网页重构、历史分支清理、正式实验、用户采用和正式发布均保留各自任务的决定及验收。本报告只交付本首版静态方案与真实测试计划。

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
| `node --max-old-space-size=512 .verification/swa-cli-headless-20261020/check-docs.mjs` | BelowNormal；已安装marked18.1.0解析两份文档，4个本地链接及16个关键源码引用存在，EXIT0/PASS；不触发业务程序 |
| `git diff --cached --check` | BelowNormal；两份文档最终暂存检查EXIT0/PASS，无空白错误 |
| install/build/model/service、产品CLI、T00—T15 | 全部NOT_RUN；未取得窗口，未修改程序/依赖/全局Agent配置 |

本次仅通过当前Dispatch的Orca通道交付，普通commit/push两份文档；没有发布GitHub评论、PR、Slack消息或公开artifact。
