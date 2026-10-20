# 无界面 Swarm CLI 独立审核

## 范围与准确需求

本报告由用户继续原Reviewer Dispatch的指令授权，写入范围为本文件及自有忽略目录。Reviewer为`task_1cd8014750f9 / ctx_980a3282c339`，原Run为`run_62ada397a104`。实际客户端为Orca/Codex，实际会话projection为`gpt-6.1-sol`、终端预览为`GPT-6.1-Sol xhigh fast`；没有子Agent。

H01—H16先依据准确main `81812c32a6d345362655c6e569f680aab8fca877`的首版任务和QUESTIONS形成，并在`846d67ecd3684b2c6e3b4922fe4714056461371d`提交。当前事实源接续准确main `e149510762f89bed06e7c8c766917554a58294e4`的`tasks/swa-cli-v1-20261020.md`及`docs/QUESTIONS.md`，并沿现行SPEC、META_PLAN和RESOURCES。已完整读取最新首版任务及相对818的全部决定变化；按用户不妨碍本机Orca并集成产品的最新要求更新H02/H13/H16、追加H17，均在读取新Worker任何计划前完成。

用户已确认五阶段：固定抽取范围；接入Orca无界面执行和控制端身份并定义CLI控制规则；完成多Worker流程；验证失败与资源处理；验证安装包。最终验收使用安装后的同一候选，完成一次至少两个真实Worker实际并发的任务，并验证指定停止、其他Worker保全及真实成果获取。

MCP、网页重构、完整旧S3及全部历史分支清理属于后续整合。旧S3规格、未答决定及失败/未验结果继续保留，按实际使用路径适用；旧Mission/Claim提案及其两项未答决定不自动成为当前首版的整体前置。当前按产品自带且隔离运行的抽取后台推进，实际命令、profile、配置/状态、端点、身份及所属进程隔离需要核实和验证；本机Orca二进制、PATH解析、账号、自启、现有后台与会话保持。适用的真实工具、项目、模型、资料和发布授权保持。

原CLI30b报告继续保留当时已接受的来源调查、设计提案及E01—E16覆盖框架；M1 b112仅接受调查和本次实际运行记录，历史故障原因与补丁有效性仍证据不足。本报告接续最新目标，不扩大这两项接受范围。

## 读取Worker计划前形成的条件

| 编号 | 正确实现必须满足的条件 | 边界、失败路径与验收判据 | 来源 |
| --- | --- | --- | --- |
| H01 准确抽取范围 | 固定真实Orca来源、必要文件及后台、直接复用/适配边界、运行时与依赖和MIT许可；从CLI调用追踪到实际Worker启动、输入、查询、输出、停止及回收 | 已知准确来源为v1.4.230/b2879bd5；不从某个小包、类型或RPC外观推导完整可运行。未知provider/native闭包及平台条件保留未知 | 阶段1、上游来源决定 |
| H02 无界面执行 | 用真实无界面入口和产品随附抽取后台完成声明的执行能力，说明产品后台发现、连接、启动与已有实例处理；控制无需借助桌面交互 | 无后台、版本/配置不符、端点不可信或连接失败在原位置失败，不静默转到桌面或本机现有Orca后台。候选CLI实际连接的产品端点与状态来源可核实 | 阶段2、e149产品隔离方向 |
| H03 真实控制端身份 | 启动、绑定、输入、查询和停止均取得可验证的控制端、Run/Task、Worker及会话归属，目标选择使用稳定身份和当前授权 | 不能通过CLI参数或持久文件伪造主体、批准或提升权限；跨任务/项目、失效凭据、过期绑定或非所属目标拒绝。复用实际Orca身份与所属边界，不以标题、PID或本机可访问单独证明所有权 | 阶段2、适用权限 |
| H04 唯一状态及产品边界 | 明确Orca后台与现有SWA服务各自权威事实、映射和直接消费者，现有产品历史与数据保留，任务/Worker状态只有可解释的来源 | 不建立第二套重复任务账本或自研调度器，不由CLI缓存覆盖后台事实。是否经过原Mission/Claim按实际设计判断；使用时遵守其原门禁，未使用时不自动追加旧S3整体验收 | 阶段1/2、唯一状态决定 |
| H05 CLI控制契约 | 在实现前定义用户/原有Agent实际可用的命令、参数、身份与目标选择、正常和失败退出、stdout/stderr及结构化结果，查询可还原同一执行 | 真实parser检查未知/缺失参数、路径和空格、换行/JSON/长输入及输出边界；输入不变成shell语句。命令回执、输入接受、实际执行、停止、释放和任务完成分别表达，未知不可伪装为成功或失败 | 阶段2、终端与Agent消费 |
| H06 启动输入与重复请求 | 真实Agent/工作目录/会话可追溯，启动及输入操作有稳定身份与可查询效果；同键同内容不重复执行，异内容冲突 | 启动部分成功、投递响应丢失、超时和客户端断开后查询原操作，不新建重复Worker、不自动重发未知输入。记录已经发生的效果、存储与资源归属 | 阶段2/4 |
| H07 同任务真实并发 | 至少两个真实Worker属于同一任务且实际并发工作；分别保留控制身份、进程/原生会话、输入和有归属的输出 | 用真实运行事实证明实际工作区间重叠。两个PID同时存活或先后各完成一轮不足以证明并发工作；测试进程、固定成功输出或本Reviewer开发会话不能代替候选Worker | 阶段3、最终明确验收 |
| H08 指定停止及其他Worker保全 | 精确选择其中一个Worker，区分停止请求与真实退出；只控制当前所属进程/会话及必要后代，其他Worker继续所属任务 | 核PID/创建身份与会话，防止旧绑定或PID复用控制其他进程。停止未知/失败或后代未确认退出时保留资源状态；其他Worker身份、上下文、输出及继续工作的能力保持，无全局终止 | 阶段3/4、最终明确验收 |
| H09 输出及成果获取 | 从安装候选按任务和Worker分别取得真实输出、实际状态和任务要求的成果，记录来源与读取范围 | 停止Worker的已发生输出及部分成果保留实际状态；继续Worker产出的成果能由控制端取得。CLI退出、摘要和日志片段不能独自宣布任务完成；截断、游标变化、产物缺失/越界及未知用量如实返回 | 阶段3、最终明确验收 |
| H10 持久性及接续 | 控制端退出/重连后可观察同一任务，后台或Worker中断后的可恢复事实、未知效果和接续方式明确 | 原生resume与新会话交接分别验证；宿主重启或响应丢失不抹除旧效果，不能凭联系丢失宣布退出。旧所有者失效后拒绝继续写入或控制，不能让新旧实例未经隔离写同一目录 | 阶段2/4、停止与接续要求 |
| H11 失败与资源处理 | 启动、输入、查询、输出、停止及回收的失败在实际边界保留，资源占用、退出与释放分别确认；额度来自当前任务和宿主事实 | 未知用量或费用不填零，未确认停止不释放为已空闲；部分启动及后代遗留有明确所属与处理。复用既有生命周期和资源控制，不通过自动重试或换Agent/模型隐藏失败 | 阶段4、真实失败要求 |
| H12 实际Agent能力与资料范围 | 使用真实已安装并配置的Agent，分别证明可启动、所需模型/工具、会话及本任务能力；固定获准公开项目、输入和成果要求 | 安装及历史成功不证明当前受限能力。受限Codex的原工具关闭要求按适用路径保持；其他Agent也核实际工具及外发边界。资料读取、写入、执行和模型调用分别按已有授权核实，不从本开发会话推导候选权限 | 五阶段实际验证、QUESTIONS |
| H13 可安装候选与依赖 | 从准确实现源码形成可安装产物，包含产品抽取后台，在独立位置通过正常入口运行；明确后台归属、配置、原生依赖、平台及原许可 | 不依赖开发树、开发node_modules、本机Orca后台或偶然端口/桌面状态。必要源码和依赖闭包及未随包的运行前提可验证；已有数据和配置保全，未验平台不声明支持 | 阶段5、e149产品隔离方向 |
| H14 同一安装候选完整验收 | 在安装后的同一准确候选中，从原有Agent或普通终端完成同一任务的两个真实Worker并发、分别查询/输出、指定停止及成果获取 | 真实可执行路径、候选版本/源码、后台与Worker身份保持可追溯；上述行为不能由不同包、开发树或多份候选的局部结果拼接。任务完成及停止Worker的终态按已定义规则核实 | 最终明确验收 |
| H15 验证计划与资源窗口 | 测试由需求及失败条件推导，使用真实CLI、后台、存储、Agent和公开任务，独立审核准确方案及实现；检查范围按实际改动选择 | 原生协议、CLI和业务状态不得伪造；OS子进程生命周期检查与真实Agent验收区分。安装、编译、服务和动态检查先取得唯一窗口并实际采样；单次验收内部双Worker并发按已确认目标设计，宿主额度不自行提高 | AGENTS、RESOURCES、用户验收 |
| H16 范围与未决定事项 | 按五阶段和产品隔离方向推进，真实能力未知继续核实；其他未答事项按实际使用路径判断，继续不依赖答案的来源和完整控制设计 | 不用完整旧S3、MCP、网页或全部历史分支清理阻塞当前首版；适用权限继续保持。调查、计划接受、代码完成、安装验收与正式发布分别记录；无证据不填通过 | 最新任务、QUESTIONS |
| H17 本机Orca保全 | 产品使用独立命令和安装目录、私有profile/配置/状态、端口/IPC、控制身份和所属进程；安装前后及双Worker并发、指定停止后通过公开CLI核本机Orca正常查询和归属保全 | 核实际命令解析及RPC/profile来源，不能只比较目录名或产品调用成功。禁止覆盖本机orca、修改PATH解析/账号/自启、停止或接管现有后台/会话；冲突端点、锁或遗留状态在产品范围内失败，不借用本机Orca完成验收，不终止非所属进程。检查既有配置及数据的适用保全，不要求读取秘密原文 | e149最新用户要求、主控msg_f572017d3cc2 |
| H18 凭据保全与真实模型路径 | 所选原生Agent/SDK及provider/model身份明确；复用原凭据时保持原账户可用，短期/过期/未知有效性在任何刷新HTTP前拒绝，原认证及配置只读 | 使用成熟库或原生接口核有效性；只读store的modify/delete在调用传入回调前拒绝，不能在刷新后仅拒绝写回。私有agentDir、model配置及资源加载不能静默改provider、引入未批准上下文、重启工具或额外重放输入；需要的支持与拒绝路径按真实实现验证 | H12/H17、主控msg_c3ead5926d05、独立SDK源码 |

这些条件没有冻结公共命令名、持久字段、供应方或具体任务输入。产品自带抽取后台及不妨碍本机Orca是当前确认方向，其实际机制仍须核实。收到准确Worker计划后，依据实际提案判断覆盖、缺口与多余检查；需要新增条件时先说明其需求依据，避免将旧提案的实现选择当作当前要求。

## 准确上游的隔离与身份核查

依据主控`msg_f20e7ba11db0`的来源线索，独立读取准确b287的实际调用路径；再次核实本地上游HEAD为b287且工作区干净。已完整读取`src/main/orcad/orcad-entry.ts`、相关路径/launcher/凭据准备与迁移/托管hooks/账号组装/daemon监督模块，以及原生身份和headless PTY注册入口。以下为源码事实，产品方案尚未到达，实际闭包及运行均未验证。

| 来源与调用关系 | 独立核实的事实 | 对H17和必要能力的影响 |
| --- | --- | --- |
| `orcad-app-paths.ts`→`orcad-entry.ts` | ORCA_USER_DATA只决定userData；home仍为真实homedir，appData按操作系统目录解析。启动将ORCA_USER_DATA_PATH改为当前userData，并组装完整OrcaRuntimeService | 私有profile不能独自封闭其他消费者；实际产品composition、CLI和子进程需要使用一致的产品根与端点 |
| `prepareOrcadSecretStoreAndMigrate`→`orcad-plaintext-credential-migration.ts` | orcadCredentialRoot固定为homedir/.orca；默认迁移目标来自Linear/Jira/Bitbucket/speech/MiniMax/Zcode/OpenCode凭据访问器。已另核Linear访问器也固定homedir/.orca，迁移会读取、收紧文件权限并重封符合条件的原文件 | 原入口不能直接按私有ORCA_USER_DATA启动后宣称隔离。产品所需秘密存储的根及迁移目标分别需要私有来源；无关集成入口可不组装，不能访问或重写本机Orca凭据 |
| `orcad-secret-store.ts`→`orcad-secret-key-resolution.ts`/`orcad-secret-key-record.ts` | prepare在resolveState之前调用sweepStaleRecordStaging；后者按传入keyDir枚举并删除符合名称/失活PID/所属条件的staging文件。临时profile的plaintext分支发生在此调用之后 | 临时profile分支只解释resolveState的部分行为，不能证明没有访问全局凭据目录；判断依据必须包括完整准备顺序和实际传入根 |
| `orcad-managed-agent-hooks.ts`→`managed-agent-hook-startup.ts`/`managed-agent-hook-controls.ts` | 启动调用Codex hooks、托管hooks协调及已安装OpenCode插件刷新；hooks开关未明确false时为开启。安装流程先刷新已有脚本，再做本机Agent存在检查；源码明确用户级Claude/Cursor配置与其他profile共用。关闭的启动分支选择skip，显式移除会删除共用hooks | 避免副作用的规则必须覆盖启动、刷新、设置变化和退出的实际消费者，不能通过调用全局remove清理“本产品hooks”。产品必要状态/身份能力仍需保留真实来源 |
| `orcad-cli-launcher.ts` | launcher使用userData/cli/bin，Windows会写orca.exe和orca-launcher.cfg并可能重命名旧映像；它本身没有注册系统PATH | 必须区分私有launcher和公开产品命令，核实际安装路径和调用解析。该模块的私有路径不证明账号、hooks或其他模块已经隔离，也不能据此宣称它必然改了系统PATH |
| `orcad-account-services.ts`→`provider-account-services.ts`及`orcad-codex-account-homes.ts` | 无界面入口仍组装Claude/Codex账号、运行home准备、限额/配置来源和resume准备；客户端CODEX_HOME选择可能协调该home的hook。账号服务未start轮询不等于没有构造和按需消费者 | 需要按选定provider核原始配置来源、产品私有运行home和实际认证路径；不能仅删账号组装后把所需启动/认证/resume写为支持，也不要求迁入全部账号界面 |
| `agent-session-claim-identity.ts`、`register-headless-runtime.ts`及入口恢复顺序 | 持久签名key以明确profile目录和authorityDomainId构造；headless PTY注册接收真实store，随后恢复orchestration authority、协调旧Worker、发布headless graph后才开放RPC。daemon目录与历史目录来自AppEnvironment.userData | 产品私有根需要贯穿身份、持久状态、provider和daemon。原启动/持久/停止路径声明使用PTY或daemon时，须保留对应真实注册及恢复；不能只让RPC可监听 |
| `orchestration-session-caller.ts`、CLI terminal/session caller helpers及`rpc-caller-identity.ts` | session调用者核真实本机记录、租约及声明绑定；session ID本身不构成凭据。传输身份区分local-cli、desktop和paired-device，声明字段与会话身份分别处理；原无身份structured调用会拒绝猜测其他pane | 控制端身份方案需要准确说明信任范围、真实登记和任务归属。不能从当前开发terminal复制handle充当产品控制端，也不能将同机用户模型夸大成操作系统账户隔离 |
| `orcad-daemon-supervision.ts`及daemon路径解析 | 原实现含启动重试及不可用时转用进程内PTY的行为，能力按真实状态声明；stopOrcadDaemon只断开，daemon可继续存在。Windows端点还须核命名管道来源 | 如果方案声明持久性、退出或资源已回收，必须提供选定路径的真实事实；不能把断开当退出或把降级当原能力成功。当前未启动任何上游后台或检查实际端点 |

这些核查说明产品组装需要封闭的全局消费者和必须保留的能力边界，没有形成全部provider依赖闭包，也没有给出可运行或隔离通过结论。后续将对照Worker的具体源码清单、产品composition和可执行测试计划判断覆盖。

主控`msg_00f28da43226`允许Worker提出产品私有Orca原生Run/Task/Dispatch为唯一执行账本的方案，Workspace仅复用适用项目/批准/固定上下文，不双写旧Mission/Claim/Relay；此项仍需准确方案审核及相应SPEC更新。Claude当前经Aliyun、当前Codex工具关闭拒绝及Pi尚无正向认证是主控报告的核查起点，本Reviewer没有调用这些模型。允许的只读认证/支持格式核查不代表已批准配置、转写协议、迁入秘密或调用模型；最终按真实配置与所选路径审核。

## Pi SDK的独立静态核查

主控`msg_c3ead5926d05`要求独立追踪SDK源码。已只读核本机公开`pi.ps1`入口和安装包，实际`@earendil-works/pi-coding-agent`及其`@earendil-works/pi-ai`均为1.0.1。源码根为`C:/Users/DW/AppData/Roaming/npm/node_modules/@earendil-works/pi-coding-agent/`；以下条件与H18均先于新Worker计划形成，没有读取真实认证原文、构造认证、配置SDK或执行任何网络/模型请求。

| 真实源码 | 已核事实及评审判据 |
| --- | --- |
| pi-ai `dist/providers/openai-codex.js`及原模型目录 | 该包原生提供openai-codex，固定ChatGPT backend-api来源，原目录包含gpt-6.1-sol。支持声明可以核查；实际认证、模型返回、Windows子进程和产品接入仍未通过，不能称原Codex控制器已经可用 |
| pi-ai `dist/auth/resolve.js`、`dist/models.js`及CredentialStore | 请求认证默认至少提前5分钟判为短期，刷新在credentials.modify接收的回调内部执行；目录网络刷新另有过期后进入modify的路径。readonly modify若入口立即拒绝且不执行回调，会阻止这两条源码路径调用refresh；回调执行后再拒绝持久化会留下远端凭据已旋转的可能。expiry的类型/有限值/单位及当前时间需要真实成熟解析与校验，不能用未知值绕过比较 |
| coding-agent `dist/core/model-runtime.js`、`runtime-credentials.js`及`pi-ai/dist/auth/resolve.js` | ModelRuntime可注入CredentialStore；默认可创建AuthStorage并读默认models配置。RuntimeCredentials将modify/delete转发原store。已有stored认证解析失败不会静默回到env；仍须核私有配置覆盖、明确model选择、目录网络/创建刷新和每次请求路径，不能把只读文件当成已证明无远端刷新 |
| coding-agent `dist/core/sdk.js`、`agent-session.js` | noTools=all或tools=[]建立固定空allowlist，工具注册及激活按该表过滤。未提供resourceLoader时会建立并reload默认加载器；未提供model时可按历史或默认寻找模型。计划需要固定真实模型与资源来源，不能从参数名推断完整隔离 |
| coding-agent `dist/core/resource-loader.js` | noExtensions仍保留显式additional extension路径，inline factories有自己的加载流程；extensionsOverride在扩展实际加载后应用。noContextFiles只关闭AGENTS上下文，SYSTEM/APPEND仍可自动发现。skills/templates也有显式附加路径。私有agentDir需与明确空附加资源及实际加载规则一起核查；只把加载结果清空不能阻止已执行的扩展 |
| coding-agent `dist/core/agent-session-runtime.js`及`dist/modes/rpc/rpc-mode.js` | createAgentSessionRuntime调用真实运行时factory并供原生RPC使用。原runRpcMode另有bash命令直接调用session.executeBash，不经过模型工具选择。产品在ToolsOff路径下须只发送获准的原生操作并把文本按原协议编码，不能向用户暴露该直接命令能力或将输入拼成RPC；无需伪造或改写Pi原生协议 |
| coding-agent `settings-manager.js`、`cache-warmer.js`及sdk调用 | cache warming默认streaming，其实现具备按条件重新发送模型请求的能力；provider重试参数也来自设置。不能推断当前模型一定触发了预热或重试，本轮没有运行。计划需核选定配置和真实额外请求边界，以满足未知输入不重发及准确模型调用记录的要求 |

本轮只确认原SDK具备相关公开接口及需约束的调用顺序，没有认定Pi已配置、认证通过、模型调用成功或产品方案可实施。H18用于对照后续准确方案及实际支持/失败路径，没有将Pi选型写成最终已接受决定。

## 当前状态

独立条件H01—H18已形成，尚未读取新Worker计划，没有方案或实施接受结论。实际新Worker为`task_4501441a2f59 / ctx_dc2fe5ba5526`，工作树/分支为`swa-cli-headless-20261020`，依据主控消息及Orca ask回复；旧ctx_c9的启动未知及开发会话恢复不当作产品运行证据。本轮没有修改程序、安装、运行模型、进程测试、HTTP、SQLite或服务。现有S服务恢复记录为PID52040，运行窗口归该服务；后续动态操作仍需主控安排。
