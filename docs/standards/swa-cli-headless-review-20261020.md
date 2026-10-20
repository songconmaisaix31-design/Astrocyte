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

这些独立条件没有预选公共命令名、持久字段、供应方或具体任务输入。产品自带抽取后台及不妨碍本机Orca是当前确认方向。后续方案对照及返修结论见下文；新增条件按实际需求及已核源码判断，旧提案的实现选择不自动成为当前要求。

## 准确上游的隔离与身份核查

依据主控`msg_f20e7ba11db0`的来源线索，在方案到达前独立读取准确b287的实际调用路径；再次核实本地上游HEAD为b287且工作区干净。已完整读取`src/main/orcad/orcad-entry.ts`、相关路径/launcher/凭据准备与迁移/托管hooks/账号组装/daemon监督模块，以及原生身份和headless PTY注册入口。以下为源码事实，实际闭包及运行均未验证。

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

这些核查说明产品组装需要封闭的全局消费者和必须保留的能力边界，没有形成全部provider依赖闭包，也没有给出可运行或隔离通过结论。具体源码清单、产品composition和可执行测试计划的对照见后文。

主控`msg_00f28da43226`允许Worker提出产品私有Orca原生Run/Task/Dispatch为唯一执行账本的方案，Workspace仅复用适用项目/批准/固定上下文，不双写旧Mission/Claim/Relay；此项仍需准确方案审核及相应SPEC更新。Claude当前经Aliyun、当前Codex工具关闭拒绝及Pi尚无正向认证是主控报告的核查起点，本Reviewer没有调用这些模型。允许的只读认证/支持格式核查不代表已批准配置、转写协议、迁入秘密或调用模型；最终按真实配置与所选路径审核。

## Pi SDK的独立静态核查

主控`msg_c3ead5926d05`要求独立追踪SDK源码。已只读核本机公开`pi.ps1`入口和安装包，实际`@earendil-works/pi-coding-agent`及其`@earendil-works/pi-ai`均为1.0.1。源码根为`C:/Users/DW/AppData/Roaming/npm/node_modules/@earendil-works/pi-coding-agent/`；以下条件与H18均先于新Worker计划形成，没有读取真实认证原文、构造认证、配置SDK或执行任何网络/模型请求。

| 真实源码 | 已核事实及评审判据 |
| --- | --- |
| pi-ai `dist/providers/openai-codex.js`及原模型目录 | 该包原生提供openai-codex，固定ChatGPT backend-api来源，原目录包含gpt-6.1-sol。支持声明可以核查；实际认证、模型返回、Windows子进程和产品接入仍未通过，不能称原Codex控制器已经可用 |
| pi-ai `dist/auth/resolve.js`、`dist/models.js`及CredentialStore | 请求认证默认至少提前5分钟判为短期，刷新在credentials.modify接收的回调内部执行；目录网络刷新另有过期后进入modify的路径。readonly modify若入口立即拒绝且不执行回调，会阻止这两条源码路径调用refresh；回调执行后再拒绝持久化会留下远端凭据已旋转的可能。expiry的类型/有限值/单位及当前时间需要真实成熟解析与校验，不能用未知值绕过比较 |
| coding-agent `dist/core/model-runtime.js`、`runtime-credentials.js`及`pi-ai/dist/auth/resolve.js` | ModelRuntime可注入CredentialStore；默认可创建AuthStorage并读默认models配置。RuntimeCredentials将modify/delete转发原store。已有stored认证解析失败不会静默回到env；仍须核私有配置覆盖、明确model选择、目录网络/创建刷新和每次请求路径，不能把只读文件当成已证明无远端刷新 |
| coding-agent `dist/core/sdk.js`、`agent-session.js` | noTools=all或tools=[]建立固定空allowlist，工具注册及激活按该表过滤。未提供resourceLoader时会建立并reload默认加载器；未提供model时可按历史或默认寻找模型。计划需要固定真实模型与资源来源，不能从参数名推断完整隔离 |
| coding-agent `dist/core/resource-loader.js` | noExtensions仍保留显式additional extension路径，inline factories有自己的加载流程；extensionsOverride在扩展实际加载后应用。noContextFiles关闭loadProjectContextFiles路径，SYSTEM/APPEND仍可自动发现。skills/templates也有显式附加路径。私有agentDir需与明确空附加资源及实际加载规则一起核查；只把加载结果清空不能阻止已执行的扩展 |
| coding-agent `dist/core/agent-session-runtime.js`及`dist/modes/rpc/rpc-mode.js` | createAgentSessionRuntime调用真实运行时factory并供原生RPC使用。原runRpcMode另有bash命令直接调用session.executeBash，不经过模型工具选择。产品在ToolsOff路径下须只发送获准的原生操作并把文本按原协议编码，不能向用户暴露该直接命令能力或将输入拼成RPC；无需伪造或改写Pi原生协议 |
| coding-agent `settings-manager.js`、`cache-warmer.js`及sdk调用 | cache warming默认streaming，其实现具备按条件重新发送模型请求的能力；provider重试参数也来自设置。不能推断当前模型一定触发了预热或重试，本轮没有运行。计划需核选定配置和真实额外请求边界，以满足未知输入不重发及准确模型调用记录的要求 |

本轮只确认原SDK具备相关公开接口及需约束的调用顺序，没有认定Pi已配置、认证通过、模型调用成功或产品方案可实施。H18用于对照后续准确方案及实际支持/失败路径，没有将Pi选型写成最终已接受决定。

## 当前状态

结论：**接受准确`396ed69b30890185bdade3ae23f8206cf18b9b82`的完整方案与T00—T15测试计划**。已独立核对5523→3162→396e的全部差异及相关真实源码，F01/F02的方案缺口已补，F03的内核采集前置已删除。后继准确6168接受F04—F07静态修复及Job方案/测试补充，退回F08关闭竞争。接受限于所列静态范围；程序实现、模型认证和动态验收均未接受，原S12800运行且没有动态窗口。

本方案首次归属为`task_4501441a2f59 / ctx_dc2fe5ba5526`，工作树/分支为`swa-cli-headless-20261020`。主控在`msg_d2e2d0b64a50`确认该静态Task已结算，返修沿原终端/工作树/分支由`task_9f062888e786 / ctx_0d069056baff`负责；仍为同一个原Worker。正式意见通过`msg_e632a2126807`送该Dispatch。计划接受后，主控准确f7c97已授予逐文件程序写域及028迁移，Worker继续静态实现；后继两项依赖专用窗口已交回，原S同源恢复为PID12800，当前没有构建、模型或产品服务窗口。旧ctx_c9的启动未知及开发会话恢复不作为产品运行证据。本Reviewer没有修改程序、安装、运行模型、进程测试、HTTP、SQLite或服务；后续动态操作仍需主控安排。

## 准确5523的独立对照

已完整读取准确8964的两份文档及8964→5523全部后继差异，合计原方案391行；后继只补实际工具/权限查询及采集前置。真实远端`swa-cli-headless-20261020`为5523。H01—H18在本次读取前已形成，SDK和原生身份事实均独立追踪，不以Worker说明代替源码判断。

| 条件 | 覆盖及独立判断 |
| --- | --- |
| H01/H02 来源与无界面 | §2/§3/§5及T01/T02保留准确b287、Node24.21.0、真实Store/SQLite、必要native/children和构建metafile；源码族及动态闭包检查方向成立。尚未构建不表示已通过，也不单独构成计划缺陷 |
| H03 控制身份 | §3.2/T03/T12提出真实PTY helper、当前launch证明和严格拒绝未验证调用者。已核agentLaunchTokenFor仅在launchConfig存在时mint，terminal-create实际注入token并保存原生归属，currentRuntimeLaunchSufficient仅适用精确当前绑定。没有发现伪造human的设计；helper创建的请求幂等和恢复映射仍有F02缺口 |
| H04 唯一状态 | §1保留私有Orca原生Run/Task/Dispatch为唯一执行账本，Workspace复用适用项目/批准/上下文，旧Mission/Claim/Relay不双写；符合主控已允许的提案边界。SPEC和程序写域仍待主控接续 |
| H05 CLI消费 | §6给出命令、退出、JSON及未知结果，T03/T06/T07/T14对应真实参数、权限和错误。控制端open/close等组合动作与底层请求的映射尚需按F02补清，不能仅靠统一envelope描述幂等 |
| H06 重复与部分失败 | T05/T06/T10覆盖Worker首输入、重复R、断开和unknown不重发；原生workerStart本身支持同次请求内的spec/parent任务创建。尚未覆盖controller open的同键并发、同键异正文以及PTY已创建而Run绑定未完成的边界，见F02 |
| H07/H14 同候选并发 | §7.3/T13明确同一安装、同一Run/根Task、真实A/B模型执行区间及双方流事件重叠，停止时A仍在执行、B原session继续；开发会话、空闲PID、串行输入均不算通过。覆盖本次用户要求 |
| H08 停止保全 | T09/T10/T11/T13要求精确PID/创建身份、实际全树退出和非目标保全，unknown不释放。普通Node生命周期fixture仅代表进程，真实Agent停止单独验证，范围适当 |
| H09 真实成果 | 已另读准确da3bdddb的README/go.mod/stats.go/stats_test.go：现有Mean真实实现和Median要求与固定输入吻合。T13目标明确为设计及检查说明，保留A实际停止事实和B真实输出，不宣称代码已实现Median或旧S3通过；输出来源、游标和未知用量有覆盖 |
| H10/H11 接续及资源 | T08—T12区分真实竞争、接管、重启、断开、退出及释放；未命中竞争保持INCONCLUSIVE，未验证structured存活时不声明恢复不中断。quota/双Worker资源窗口不扩大宿主额度 |
| H12/H18 Agent和认证 | §4、T04明确Pi SDK身份、原生provider、只读原账户、空allowlist、原生RPC白名单、关闭cache/retry/compaction；关键方向成立。SDK服务工厂仍触发全provider刷新，F01使固定provider与认证来源范围尚不成立 |
| H13 安装闭包 | §5/T01/T02保留Node/native/SDK及资源，不依赖全局node_modules或本机Orca；未验平台不扩写。最终准确metafile、产物装载和安装效果按计划动态检验，当前全部NOT_RUN |
| H15/H16 测试与范围 | 原生SQLite/HTTP/SDK/Agent分别验证；纯fixture未冒充Agent；Go/Node/契约/架构检查与本轮新增消费方相关。完整旧S3、MCP、网页、历史分支清理均未追加为前置。F03指出的自选内核采集前置需要删除，其余未发现应删的无关检查 |
| H17 本机Orca保全 | §3明确全局凭据迁移/hooks/账号等不组装、私有根/端点/进程身份、原Orca公开查询和配置保全。实际写入边界需按已组装消费者验证；全量WPR/Kernel-File并非本条件指定的方法，现计划将其变成T02/T15采集执行者前置，需按F03修订 |

### F01：SDK服务工厂仍遍历全部provider

严重程度：较高。影响H12/H17/H18及方案§4.2—§4.3/T04。

准确Pi1.0.1 `dist/core/agent-session-services.js:createAgentSessionServices`在资源加载后无条件执行`await modelRuntime.refresh({allowNetwork:false})`。`model-runtime.js:refresh`在未限定providers时重建所有provider，并调用queueAvailabilityRefresh；后者对所有provider执行getAvailable/checkAuth及credentials.list。`refreshOnCreate:false`只抑制ModelRuntime.create中的初次刷新，不会抑制这条后续调用。因此现选服务工厂不能兑现“不枚举其他provider环境/认证”的承诺。这里尚未执行，未认定已有秘密读取或额外网络请求。

原Worker需选择真实SDK公开接口形成只触及固定provider的完整组装顺序，明确资源加载、auth snapshot、session/runtime factory和原生RPC消费方。原ModelRuntime支持带providers集合的refresh，但createAgentSessionServices未传该限定；不能仅重复创建参数或将所有provider的读取结果填成undefined来宣称无枚举。T04须核实际调用的provider集合与真实模型/认证来源，在首次prompt前确认没有未批准来源探测；保留原生SDK及只读modify/delete的刷新前拒绝。

### F02：controller创建的幂等与恢复映射缺失

严重程度：较高。影响H03/H05/H06，方案§3.2/§6及T03/T06/T12。

方案把全部写命令的request-id归入原生orchestration receipt，但controller open先创建真实PTY/helper，之后才能用该helper创建或绑定Run。准确b287的`shared/orchestration-rpc-contract.ts:isDurableMutation`没有将terminal.create列入orchestration durable mutation；`rpc/orchestration-mutation-executor.ts`对不适用方法直接invoke。原生另有`terminal-create-deduplication.ts:dedupeTerminalCreate(clientIdentity,worktreeSelector,clientMutationId,reconcileExisting,...)`，需要明确调用、命名空间和现有reconcile证明，不能由correlation ID自动获得。

需要补充controller open/close的实际原生调用顺序、公开R到原生请求/创建身份的稳定映射及保存位置，说明同键异正文如何冲突、PTY已创建但响应或Run绑定丢失时如何查询和保留占用，以及后台重启后从何处恢复。复用已有原生操作及状态，不新增第二任务账本。T06补同键并发open、同键异项目/目录及真实创建后的响应丢失；确认只有同一个可解释的helper/Run，未证不存在时不能再创建。当前测试只检查原生Task/Dispatch/输入去重，不能覆盖bootstrap阶段。

### F03：删除自选内核采集前置，采用足够的定向证据

这是测试方法调整，不要求新增提权、性能组或驱动。H17规定隔离与本机Orca保全结果，没有指定WPR/全量Kernel-File采集。已只读核本进程非提升，权限列表没有SeSystemProfilePrivilege，wpr/logman文件存在，PATH没有Procmon入口；这些只说明当前观察，不能推导全机没有其他工具，也不能证明可采集。

5523将T02/T15绑定到另行安排的有权限采集执行者，属于自选方法增加的前置，应移除。当前固定候选及声明路径的最小充分证据可以由以下组合取得，无须强制全量内核追踪：

1. 对实际组装的每个可写消费者核来源和私有根：profile/SQLite、metadata/lock/key、daemon/history、launcher、session/journal及SDK状态；源码调用闭包关闭全局迁移、hooks和账号写入，必要端口用真实路径/reparse/所属验证及写入准入约束。模型认证另按H18验证，不以本机文件未变单项证明没有远端刷新。
2. 同一安装候选实际运行，核真实profile/DB文件、端点、CLI解析、runtime/daemon/helper/Worker的PID、创建时间和所属；以原Orca根、重合/reparse目录、占用端点及非所属进程作为真实拒绝输入，确认在写入/终止前拒绝。对已审消费者使用现权限可执行的定向观察，具体工具/API、路径、范围及判据写入计划。
3. 在安装前后、A/B并发、停止A后及最终退出时，用本机原CLI公开查询并核原runtime/daemon/会话归属，定向比较受保护配置/认证文件的内容与ACL；结合产品实际写入准入和源路径证明其保全。原客户端自然写入或认证变化需要按实际来源解释，无法解释的具体范围保留证据不足，不回写旧状态。
4. 只关闭本产品已核所属的进程与资源，错误或unknown保留实际占用；原S服务按主控窗口处理。不得为观测安装驱动、修改全局审计策略、取消其他trace或终止非所属进程。

这些要求覆盖本次实际风险；没有要求整棵活跃Orca目录逐字节不变或恶意同用户隔离证明。原Worker需给出当前可执行的具体检查并更新T02/T15。缺少某一必要事实时明确该具体未验范围，不把自选工具的缺失改成全体实现必须等待提权。其他源码和接口准备可继续。

## 实际静态验证及返修边界

- `.verification/check-revision.ps1 -Previous 4e6e8ed4fa52521c2232dd770eb98297d8713785 -Candidate 8964f0df4569c82aa90d8f4fed85c1f8928e7e8a` EXIT0，两份Markdown、4本地链接、Git空白通过。
- `.verification/check-revision.ps1 -Previous 8964f0df4569c82aa90d8f4fed85c1f8928e7e8a -Candidate 5523564f2d76ee992b8ffe00d590dbc3b24a4f85` EXIT0，一份变动Markdown、1本地链接通过；未变任务沿前项结果。实际差异只有工具与权限记录，没有修订SDK组装和controller请求映射。
- 已安装marked18.1.0、BelowNormal、Node512MiB。真实`git -c http.sslBackend=schannel ls-remote origin refs/heads/swa-cli-headless-20261020`为准确5523；没有修改全局TLS配置。原默认TLS失败由Worker/主控保留，本Reviewer本次远端查询成功。
- SDK服务工厂、auth/refresh、工具与资源加载、RPC直接bash、SSE maxRetries路径、原生launch证明/receipt/terminal去重及公开固定项目四份原文均已按实际源码只读核查。没有运行SDK、创建认证、调用模型或执行测试脚本。

返修交原Worker沿原会话、工作树及分支完成。F01/F02影响依赖的模型/控制路径，修订后重新审核准确方案；F03不新增运行权限前置。原M0、八份文档集成及M1调查的接受范围保持。程序实现、安装、实际隔离、ToolsOff认证、同候选双Worker及全部动态测试仍NOT_RUN，当前不作通过结论。

准确5523退回hook为`msg_b66fab57fbe3`；主控在`msg_7f066521739c`明确完整读取H01—H18对照、F01/F02的实际源码与必要返修、F03定向证据组合及需删除的内核采集前置，确认未批准程序或动态操作、原接受范围保持，并回复确认本次唤起。原Worker已在新Dispatch读取正式意见，下一准确提交按变更范围复验。

## 准确396e的方案返修核查

本次准确来源为`396ed69b30890185bdade3ae23f8206cf18b9b82`，包含前序`3162f9db37a3ea2d7799d058f4cdc6436e33fb49`。实际远端该分支为396e、工作树干净。已读取两次完整差异，没有只审核3162而遗漏SDK过滤前的home消费者；未变内容复用前述H01—H18独立对照。

| 条件与原问题 | 返修后的事实及结论 |
| --- | --- |
| H12/H17/H18、F01 | §4.2—§4.3明确不调用无条件全量refresh的createAgentSessionServices，使用公开createAgentSessionRuntime(factory)、DefaultResourceLoader、显式createAgentSession及真实services。已核原runtime factory确实直接调用该factory；sdk.js在传入真实modelRuntime/resourceLoader/model时不创建默认来源、不reload、不寻找默认模型。model-runtime.js的providers限定分支只对选定provider重组、刷新和更新availability；T04核真实provider操作、snapshot、errors和auth来源。静态内置catalog与实际认证访问分别陈述，不再以创建参数代替后续调用约束。方案缺口已补，实际认证仍未运行 |
| H17/H18、过滤前资源发现 | 已独立核package-manager.js的resolve/addAutoDiscoveredResources：noSkills过滤前仍读取用户.agents/skills，来源为process.env.HOME或homedir；projectTrusted=false关闭项目祖先发现。396e在SDK真实子进程启动前指定私有HOME/USERPROFILE、agent/session目录及随包PI_PACKAGE_DIR，原认证保留唯一显式只读绝对路径，父进程和用户环境保持原值。T04增加真实解析路径与过滤前调用闭包检查；解析到本机用户资源根即失败。不能以最后资源列表为空单项判定隔离 |
| H03/H05/H06、F02 | §3.3区分controller open与task create，并把四项待实现composition方法明确接入已有OrchestrationMutationExecutor及mutation_receipts；公开R、固定子请求、canonical payload、PTY预分配handle和实际身份均有保存位置与副作用顺序。原生terminal创建去重只合并in-flight，真实reconcile在找不到时可创建，方案准确区分这些能力；查询不调用有创建副作用的reconcile。副作用后异常转为operation_unknown以保留pending，原executor对这类pending不再invoke，符合已核源码。T06补同键并发、异正文、PTY/Run/根Task边界、关闭后丢响应及重启只读查询，未命中窗口保持INCONCLUSIVE。方案缺口已补，没有另建任务账本 |
| H17/H15、F03 | §3.1/T02/T15采用实际写入消费者闭包及写端准入、规范路径/reparse、真实SQLite/端点/PID创建身份、本机Orca公开查询、定向配置/认证字节与ACL及目录通知的组合。明确FileSystemWatcher没有writer PID、可能溢出，快照及无通知均不能单独证明隔离；自然变化无法解释时保留具体证据不足。删除WPR、提权、性能组及驱动前置，范围符合原条件 |
| H01/H02/H04/H13、模块及安装 | §5.3逐项申请现有文件和必要新增目录，Go CLI/HTTP只消费App端口，Node适配使用真正RuntimeClient，原生状态唯一；SQLite仅保存项目/批准关联。第三方来源副本与产品composition分开，原上游只读，许可证和真实依赖闭包保留。028迁移序号仍由主控核准；所列文件未获写入授权，不代表Reviewer已经授权实现 |
| H07—H11/H14/H16、其余边界 | 同一安装候选、同一任务两个真实模型执行区间与流事件重叠、A实际停止全树退出、B保全并继续、分别查询和真实成果仍完整保留。失败/unknown不释放、不换键重发；开发会话、普通进程fixture、空闲PID和旧S3均未替代真实Agent验收。没有发现本次返修新增的无关测试 |

原生checkpoint使用已有数据库接口，composition handler及精确方法分类属于后续产品实现；这些内容需要按已审计划实际编码和验证。本次源码核对只能说明方案所用接口和调用顺序有真实来源，不能证明新增handler、隔离写端、安装闭包或失败恢复已经工作。

实际验证：`.verification/check-revision.ps1 -Previous 5523564f2d76ee992b8ffe00d590dbc3b24a4f85 -Candidate 396ed69b30890185bdade3ae23f8206cf18b9b82` EXIT0；两份Markdown、4个本地链接、无锚点/HTML或空白错误。marked18.1.0、BelowNormal、Node512MiB。新增Microsoft外部文档链接未重新联网读取，该链接不作为此次接受的独立运行证据。真实远端查询使用单次schannel配置，未修改全局配置。

主控准确`b26a2fab9b5cad853c777793dd17593fb4a509a2`已将独立后台及唯一执行状态写入SPEC §21.1，并更新当前Task归属；独立管理核查见主线报告。Worker396e任务明确尚未纳入该后继，开始实现前需由主控安排普通纳入及明确写域。M0、M1调查、八文档集成和原来源调查的接受范围保持。

## 原S服务窗口脚本的只读核查

主控`msg_f426537127f9`指定只读审查`C:/Users/DW/orca/Astrocyte/.verification/swa-cli-v1-control/service-window.py`。已完整读取234行；该忽略文件没有源码Commit，本次准确SHA256为`8084f0c34e913285b357695e4682e9586ef9656ce2ad141394004baf1c306ee8`。没有运行脚本、访问原SQLite、发送HTTP或停止/创建进程。该核查不作为396e程序或动态通过证据。

结论：**该脚本当前版本退回窗口执行**，需主控补充异常接续办法。其正常路径按原PID52040及准确创建时间、真实exe/优先级/唯一loopback监听核所属，保存实际环境和原Q/media/CodeGraph来源，恢复使用同一S、原库、原命令及环境，并核健康和原Web；这些正常路径约束有具体实现。

1. `stop`先terminate/wait并保存stopped.json，之后才运行resources及backup。若后两项失败，原S可以已经停止而current.sqlite缺失或不完整；`start`强制读取并比较该备份，缺少已停止但备份未完成时的恢复入口。停止超时或stopped.json写入失败也需按准确原身份接续，不能因没有成功标记推断原进程仍运行。
2. `start`在Popen之后依次读取新PID、保存launch.json、等待监听、核环境/健康/Web及写最终running指针。后续任一步失败可以留下新S运行或状态未知；重跑start会因端口占用或证据目录已存在拒绝，脚本没有只读核同一launch身份并完成恢复记录的入口。Popen之后、launch.json写入之前的间隔也需要明确保存和恢复所属的办法。
3. 应在停止前具备必要保全资料及可用备份，并保留原库变化的核对条件；对于实际已启动的新S，按已有准确身份只读检查并接续记录。异常时保留实际状态，不自动再次启动、不终止未知进程、不用备份覆盖原库。原PID被复用时安全拒绝是合理行为，不能据此宣称原S仍活。

准确问题通过`msg_510546b1c82f`交主控；修订及执行均归主控。该限制只影响后续服务窗口交接，已接受396e静态方案及其他独立准备可以继续。原S仍按主控记录运行PID52040，本Reviewer没有重新验证其当前运行状态。

### 主控返修后的窗口脚本

已按`msg_e7da135158c3`完整只读返修后的345行，准确SHA256为`28433cfdd297814a39211d6f5926cd3c342216bd3c2862eb42eff1a211b27b29`。**接受该版本作为主控手动保全/恢复的静态方案**，原两处正常恢复入口缺口已补：

- prepare在原服务仍运行时保存实际身份、环境、来源和配置，先核额度并完成一致备份/完整性，再写prepared；stop接受该基线，持久化准确stop-intent后才终止，正向退出同时输出并记录，停止后另备份stopped.sqlite，最后写window-ready。
- 正常start要求window-ready，比较停止后备份与当前原库，再以同一S、原环境和原库启动。新增recover-original仅在窗口尚未ready时可用，明确停止区间比较不可用，禁止在ready后绕过比较；该入口不覆盖原库。
- Popen后输出并记录准确新进程身份，新增reconcile-start只核已有launch及真实环境/健康/Web，再完成running元数据，不再创建或终止进程。旧PID若被复用保全其新所属者。

停止退出或launch身份的正向记录都缺失时仍为未知。主控已明确根据实际stdout/系统身份检查后显式接续；脚本不自动从文件缺失推断未发生副作用，也不保证任意中断后自动恢复。静态接受不代表实际备份、停止、窗口就绪、启动或恢复成功；所有这些步骤仍由主控在获准窗口执行和核实。本Reviewer没有执行该脚本，也没有独立复跑其syntax检查。

准确396e的方案接受hook为`msg_f3ddbb00f6f1`，包含初版窗口脚本退回的准确hash和分别适用范围；主控在`msg_09941fdfc009`明确完整读取条件对照、真实factory/provider/home、原生receipt/unknown及F03定向证据，确认F01—F03关闭、仅方案/测试接受及动态全部NOT_RUN，实际读取回复确认本次唤起。准确b26管理接受hook及实际回复分别为`msg_a32ea5a916a4`、`msg_d14013f15d20`。返修窗口脚本的后继静态结论另随本报告回传原Run。

返修脚本的后继hook为`msg_b33ac8e678e8`；主控在`msg_4c58b53912fc`明确完整读取345行脚本静态结论、准确28433c及prepare/stop/恢复/reconcile边界，确认缺正向记录仍未知及仅静态接受，实际读取回复确认本次唤起。其后`msg_6546abe4db8b`报告prepare实际EXIT0、备份完整性/外键及原环境/配置复核通过，原件在主控`.verification/swa-cli-v1-control/prepare-01`，原PID52040仍运行；未stop、launch或恢复，窗口未交付。本Reviewer已读取该消息，没有执行或独立重跑prepare，不把该保全报告计作新CLI或动态隔离通过。

## 依赖专用窗口交回后的范围

主控`msg_75b31dd84dfc`报告原S52040正向退出，只向原Worker交付锁生成和冻结安装两项ignore-scripts命令；`msg_c06fa2e8159a`报告两项EXIT0后收回窗口并完成restore-01，同一S恢复PID12800。构建、模型、产品运行及T00—T15功能验收没有据此通过，原Worker继续静态实现，准确代码Commit/PR尚未就绪。

本Reviewer只读核非秘密stop-01的stopped/window-ready和restore-01的database-check，以及restored中的source/PID/创建时间/路径/health字段：52040实际退出15、窗口ready；停止区间数据库比较equal、integrity ok、foreign_keys空；恢复为同一ed543c12程序/Web和原state.sqlite，PID12800、创建UTC09:43:52.279923、health ok。未读取实际环境或SQLite原文，未重跑备份、服务、HTTP或安装；实际完整环境/Q/media/CodeGraph与原Orca保全为主控按已审脚本取得的记录，当前只接受记录范围。

最新管理来源为main准确e4df5158，核查见主线报告。候选安装、SDK真实认证和ToolsOff、双真实Worker并发、指定停止及成果仍待后续准确程序和实际结果审核。原M0、M1及S3范围保持。

主控`msg_9d3fef7a4e0e`仅将窗口脚本的停前期望改为12800及创建UTC09:43:52.279923，准确SHA256为`0742a46eb64d399ed8171b5a9f426a0a5bfb53e1561442428d7484b46df22424`。已只读核两个常量，并在内存中仅替换回先前两个值，所得全部文件字节的SHA256与已审28433c完全相同；没有写入文件或执行脚本。**接受这两处身份常量的静态同步**，其余流程沿用已审范围。该同步不代表下一次prepare、停止或窗口交付已发生；执行前仍须核实际PID、创建时间、路径、环境、所属及唯一监听。

本次两常量同步hook为`msg_992fef6a0e30`；主控在`msg_9c7ecc38bfe7`明确完整读取准确0742a46e与两处变化，确认其余流程保持、下次重核及未作动态通过，实际读取回复确认本次唤起。

## 准确847ec阶段代码的独立静态CR

结论：**退回准确`847ec2687857b06430284aecdf81a6887f2f996f`中F04—F07所列已成形代码**。整体功能与动态结果仍证据不足；CLI、成果导出、安装、停止保护和真实测试是原Worker明确声明的阶段待办，不因阶段尚未完成单独判定失败。396e方案接受和此前所有接受范围保持。

依据主控`msg_14fa15fafb20`、`msg_b3ec74ef5115`和`msg_a47d578c243e`只读准确Git对象，其父提交为已审f7c97；没有读取正在修改的工作文件代替该提交，没有运行Go/Node功能检查、SDK、模型或服务。847ec已普通push，但没有正式PR，当前不发布GitHub COMMENT review；待真实PR及准确后继再按相应范围CR。

### F04：同一controller的两项task.create可共同通过唯一性检查

严重程度：较高，影响H03/H04/H06。准确`runtime/swa-orca/src/native-methods.ts:163`的taskCreate在第167行检查已有task receipt，随后第168行await exactController，直到第169行才保存带controller的checkpoint。两个不同R可在await期间都通过检查，然后各自向helper交付task.create。原生mutation executor只按caller/R去重，没有按controller合并不同R；helper串行也没有重新检查单任务约束。准确上游`rpc/methods/orchestration/runs/runs.ts`的runCreate每次调用都会createRun并更新该调用者当前Run，因此存在创建两个Run/root Task及改变绑定的执行交错。

原Worker需在首次可能创建Run之前完成同一controller的原子准入，复用已有原生持久状态；pending/unknown继续保留占用。T06补不同R、同一controller的真实竞争及重启未知情况，只有同一Run/root Task可被接受。这里是源码推导的可行交错，没有宣称动态复现。

### F05：固定两Worker上限没有原子占位

严重程度：较高，影响H07/H11及T08。准确`native-methods.ts:178`的controllerDeliver在await身份观察之后调用validateTaskAction；第228行只计算已经创建的子Task数量，随后异步向helper发送真正workerStart。多个不同R可以在子Task尚未创建时都看到少于2而通过，helper以后逐项调用真实workerStart即可超过上限。准确原生workerStart/local-worker-start没有该SWA root Task的max_workers=2约束；原Coordinator的maxConcurrent属于另一条自动协调入口，此处没有使用该入口。

需要把同一root的创建接受与配额占用合并为原子准入，unknown请求保持额度，不以等待helper或同R去重代替异R竞争保护。按已审T08使用真实请求/原生SQLite检查竞争，禁止模拟接收器；未命中竞争保留INCONCLUSIVE。

### F06：释放判据读取了不存在的外层字段

严重程度：较高，影响H08/H09及T11/T13。准确`native-methods.ts:213`的requireReleased在第216行读取row.release_state。准确上游`orchestration/db/worker-terminal/worker-terminal-listing.ts:listWorkerTerminalResources`返回的外层行含resource，release_state位于真实WorkerTerminalResourceRow内；外层没有该字段。有任意Worker行时，该表达式即使对已released资源也得到undefined并拒绝controller.close/task.complete，静态类型也不匹配。

原Worker需按实际原生资源/退出状态结构判断，resource缺失或unknown不能默认released。T11/T13核真实已释放资源允许完成/关闭、未释放或未知仍拒绝。未运行类型检查，不将源码定位写成已获得编译器错误。

### F07：合法人类安全停止被执行批准与旧上下文阻断

严重程度：较高，影响原权限规则、H08及T03/T10。准确`internal/workspace/app/orca_execution.go:207`的authorize在合法human通过AuthorizeProjectActionCaller后仍无条件调用ValidateProjectActionScope，第279—290行还对所有controller动作拒绝旧context settings_revision。因此关闭AllowAgentControl、撤销AllowedActions或更新设置后，合法已认证人类也不能停止仍属自己的Worker。

既有`internal/workspace/app/projects.go:66`对合法human只核身份返回，StopNativeSession在第742行核实际会话所属后调用stopOwned；`project_policy.go:91`也明确人类安全停止不依赖执行批准。新消费方应保留该规则，安全停止仍须严格核真实native所属、PID/创建时间及退出证据；Agent的过期/撤销准入保持拒绝，不构造human凭据。使用已有真实人类认证核撤销执行批准及设置变化后的自有停止，不扩大到其他主体或未知进程。

### 原始来源与产品适配的边界

已用原生Git树跨只读对象库比较准确b287与847ec中的来源子树，原件为自有`.verification/orca-source-847ec268.json`：10处内容适配、1个新增swa-sdk-launch、31个100755→100644的文件模式变化；另有314个原始已跟踪路径未进入该子树，范围为docs/audits及其他文档/审查工具和7个tests/tools。LICENSE及根package/锁等对象与原来源相同。全部10处内容适配及新增SDK launch均已读取；31个模式变化不计为产品行为适配。

这些缺路径尚未发现运行时源码缺失，不能据此认定Windows功能失败；原Worker应说明实际抽取及模式边界，不能把该子树称为全部原始源码逐项相同。安装闭包仍按计划待验，不追加运行全部上游fixture或修改原始终端文本的要求。真实完整来源空白检查的EXIT1由Worker记录保留，本Reviewer没有重复该检查。

实际检查：产品路径`git diff --check f7c97d36 847ec268 -- . ':!third_party/**'`通过；已安装marked18.1.0只检查本次三份产品Markdown，6本地链接通过；BelowNormal、Node512MiB。只读来源比较使用Git原生diff及两个已有对象库，不执行上游程序，不安装或编译。F04—F06已通过`msg_0e341cd05d66`交原Worker，F07及来源边界通过`msg_5aceb2713a1d`补充；修复均由原Worker完成。

准确847ec退回hook为`msg_dacb32984985`；主控在`msg_528a75bf20b2`明确完整读取F04—F07的行号/原生依据、来源及模式边界、实际轻量结果和J01—J06，确认仅退回已成形缺陷、阶段待办不单独判失败、396e方案接受保留及没有动态窗口，实际读取回复确认本次唤起。原Worker已收到具体意见，下一准确方案及返修仍待交付。

## 停止强化方案的独立条件

主控`msg_461d9f9b6f74`要求复核后继Windows Job方案。以下J01—J06在收到准确方案版本前形成，沿H08/H10/H17及T09/T13，不替代原测试或扩大为全机隔离验收。

Microsoft明确kill-on-close在最后一个Job句柄关闭时作用于其成员；Job查询必须核API成功返回，空句柄会查询调用者当前Job。上述细节使句柄所属和见证路径成为必要条件。依据为[Job Objects](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects)、[QueryInformationJobObject](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-queryinformationjobobject)和[TerminateJobObject](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-terminatejobobject)，已完整读取。这些是API事实，以下为根据本任务自行推导的验收条件。

| 条件 | 必须满足的行为、边界与失败路径 |
| --- | --- |
| J01 模型前准入 | 实际随包native能力、Job创建/限制/分配及SDK真实root绑定在首次模型请求和可生成后代的加载之前成功；缺API、原生装载错误、分配拒绝或身份不明时立即失败，不只报告能力标志或切回宽松启动 |
| J02 精确所属 | 每个Worker使用可解释的自有Job、PID/创建时间和session/Dispatch；不得把原Orca、IDE、daemon共享Job或B当A的终止对象。说明句柄创建者、持有者、继承/复制/关闭和重启后如何核所属；空句柄或当前Job不自动代表自有 |
| J03 实际后代范围 | 核已采用启动调用中的嵌套Job、breakaway和句柄继承，实际模型子进程及其已用后代创建路径都进入获准范围；适用环境拒绝嵌套时失败。只审本产品实际路径，不要求证明恶意同用户或全部Windows进程创建接口 |
| J04 生命周期与见证 | 分别说明正常stdin关闭、abort、强制停止、root异常退出及宿主失联/重启。kill-on-close依赖最后句柄，额外见证句柄可能延后其生效；必须给出真实句柄生命周期和独立退出观察，不能以设置flag成功、root退出、API调用成功或一次空枚举判全树退出 |
| J05 未知保全 | 只有真实root和完整目标树退出可写停止确认/释放；Job查询失败、原生证据缺失、身份改变、句柄丢失及残留保持unknown和占用，不关闭未知或非目标资源，不重发任务。重启不能把无记录当无后代 |
| J06 真实检查 | T09使用真实自有Node进程/后代及实际安装的native能力检查正常与异常关闭，fixture不冒充Agent；T13仍用同一安装候选的真实A/B并发，停止A时B原进程/session继续执行并取得成果。补丁和能力必须属于该候选；不改本机Orca，不新增管理员、驱动或全机追踪前置 |

上述条件已先于读取后继准确6168方案写入80ed37f6并普通push；该版本及新的范围结论见下文。当前原S12800运行、没有动态窗口；后继仅按实际来源和相应变动复审。

## 准确6168的四项返修及Job方案复核

准确来源为`6168b4ea19e3b5468403e71a77a279b4bc18adf3`，已普通push，真实远端同SHA。已读取847→6168全部五路径差异；来源副本和原有依赖没有变化，沿用前次来源核查，没有使用工作目录中仍在编写的保护、CLI或打包代码代替此提交。

结论：**接受F04—F07的静态修复及§5.4/T03/T06/T08/T09/T12/T13的方案与测试补充；退回当前已成形代码的F08关闭竞争**。整体产品、native Job实现和全部动态结果仍证据不足。上述两项接受与该代码退回分别适用，不阻止原Worker继续已审方案的独立实现。

| 原问题 | 静态复核结果与接受范围 |
| --- | --- |
| F04 controller单任务 | native-methods新增reserve，在原生BEGIN IMMEDIATE事务中重新读取controller原生状态、拒绝既有任务/占用并保存同一receipt checkpoint，提交后才向helper发送。没有事务内await；公开R与原生R须一致。两个不同R的单任务准入缺口已补；仍需T06实际竞争与重启结果 |
| F05 两Worker配额 | 同一事务合并已创建子Task和pending worker.start receipt；借真实子请求、原生accepted/完成receipt及Dispatch识别已物化任务，避免重复计数。未知未物化继续占用，第三名在副作用前拒绝。已核原生getPendingWorkerStartRecovery与完成receipt的dispatchId来源，配额缺口已补；T08仍NOT_RUN |
| F06 释放字段 | 按row.resource.release_state读取，缺resource仍拒绝，不默认released。对应原生返回结构一致；已释放与未知的实际分支尚未运行 |
| F07 人类安全停止 | 只有已通过真实传输身份检查的human且permission=stop跳过执行批准与旧context settings_revision；项目所属继续核验。HTTP将真实human的project scope取自已解析路径，不接收正文提供的human身份。Agent撤销/过期门禁及其他动作保持原检查；T03真实人类停止仍待执行 |

### Job方案对J01—J06的覆盖

§5.4选择父runtime为每个Worker的唯一Job句柄持有者，SDK只继承成员资格。模型前原生Job限制与root绑定/成员查询成功，再通过Node原生IPC准入SDK加载；root退出后父仍持有确切句柄，可查询及终止实际残留。这给出了J01/J02/J04要求的所属和见证路径，没有把SDK自身退出当最后句柄关闭的证明。

已独立核Node24.21.0的stdio/IPC文档及准确上游process-spec、spawn-resolution、run-process：真实PipedProcessSpec允许附加IPC，实际spawn将stdio和serialization传入nodeSpawn，未使用自写协议。Node send的成功回调可以早于子进程接收；方案由bootstrap实际收到准入后才加载SDK，后续实现不能以发送成功代替SDK已准入或模型已开始。[Node stdio/IPC](https://nodejs.org/download/release/v24.21.0/docs/api/child_process.html#optionsstdio)

原生句柄不可由JSON构造、不可继承、不查询NULL/current Job，A/B独立且不触碰原Orca/IDE/daemon；分配及嵌套失败拒绝，所用CreateProcess路径禁止breakaway。已完整读取并核[AssignProcessToJobObject](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-assignprocesstojobobject)的分配和嵌套限制，实际环境仍须检查，不能从类型或文档直接判支持。J02/J03覆盖成立。

正常、强制、root故障和宿主失联分别陈述：真实root exit、两次成功完整Job空查询与实际旧进程身份退出观察共同满足后才返回tree exited及关闭句柄；API返回成功单项不结算。旧句柄在重启后丢失仍unknown，不复原数值、不自动重发或启动替代者，满足J04/J05。测试T09以实际同候选Node/native和真实后代覆盖这些路径，T13仍为真实A/B并发、停止A且B原session继续；未发生的PID复用/嵌套拒绝保持未覆盖，fixture不冒充Agent。未发现本次补充应删除的无关测试，J06覆盖成立。

这是方案和测试计划接受。6168没有相应native Job、bootstrap或生命周期实现，更没有二进制/安装/模型/全树退出结果；仍须以后继准确代码及实际T09/T13审核。没有重新运行已审来源、功能测试、SDK或服务。

### F08：controller关闭与创建准入没有共同屏障

严重程度：较高，影响H03/H06/H08及T06/T11。准确`native-methods.ts:193`的controllerClose只通过requireReleased检查已物化Worker行，忽略已由reserve保留但尚未生成Dispatch的task.create/worker.start。第199行保存关闭checkpoint后还会await身份观察；reserve在第220行也没有检查该controller的pending/completed关闭记录。

存在这一交错：Worker创建已保留或正准备保留，close观察到零资源并通过；另一不同R的创建在实际PTY关闭前继续获得准入/创建Dispatch；close随后关闭helper并只观察其PTY/root/helper退出。准确上游`session-tabs/explicit-terminal-close-stops.ts:133`只停止所选PTY及相关surface，没有等待该Run全部独立Worker退出。结果可能为controller关闭成功而新Worker仍运行或启动未知，违反“所属全部Worker退出/归档/释放后才能关闭”。这是源码推导，没有动态复现声明。

原Worker需将close的持久关闭标记、未决创建占用与新建准入放入同一原生事务屏障：创建先保留时关闭拒绝或等待明确结算，关闭先保留时后续创建拒绝，unknown保留屏障及实际资源。复用现有原生receipt，不新增执行状态账本。T06/T11补不同R的close对task.create/worker.start的真实竞争；分别核两种顺序，不把关闭helper当停止全部Worker。已通过`msg_08a9c85e2399`交原Worker。

实际轻量检查：`.verification/check-revision.ps1 -Previous 847ec2687857b06430284aecdf81a6887f2f996f -Candidate 6168b4ea19e3b5468403e71a77a279b4bc18adf3` EXIT0；五路径Git空白、两份Markdown、5本地链接通过，无锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。新增官方API/Node相关章节已读取，未进行动态验证、编译或安装。

准确6168的决策hook为`msg_18d91715fa96`；主控在`msg_ef7d11edb358`明确完整读取四项静态修复、J01—J06/§5.4及测试补充的接受范围，并读取F08原生关闭作用域、交错及共同屏障要求，确认返修交原Worker、整体产品和动态未接受，实际读取回复确认本次唤起。

## 准确f0ba与53ad的程序及真实fixture静态CR

主控`msg_d76b0a4ae070`交付准确`f0ba550352ea28b63b008c12542e83dc47adbdf3`，其两个父提交分别为程序`10deb049cac61dd14646f2164bd291a1fc14de09`和已审main`13974f10041d2f8e846b951418172433751798c8`。10deb到f0只有五份主控文档变化，这五份文档与13974对应对象完全相同；程序审查依据为准确10deb/f0，未将main记录纳入程序修复。随后按`msg_548c85fec44e`读取准确`53ad6e1479ef5cf3099372d2855087cf002513bb`全部差异，真实远端分支同53ad。

结论：**接受f0的F08静态修复；退回f0及53ad继承的F09未绑定Job退出结算；整体产品、安装、SDK认证和动态结果仍证据不足。** 53ad的共享字段校验、普通进程fixture及测试命令可以继续实施，尚未运行或尚未补齐的阶段项目保持待办。此前M0、M1记录、接口方案、集成计划和6168各项接受范围保留。

### F08共同关闭屏障的静态修复

f0的controllerClose通过与task.create/worker.start相同的reserve进入原生BEGIN IMMEDIATE事务。事务重新读取所属和scope，拒绝该controller既有关闭receipt；关闭分支核同一controller的pending taskCreate和worker.start，再核真实资源已释放，最后保存关闭checkpoint并提交。创建先保留时关闭拒绝；关闭先保留时后续创建及写动作被持久关闭记录拒绝，pending/unknown仍占用。现有原生同R回放在handler外处理，未新增执行状态账本。

这补上6168所列不同R竞争缺口。真实T06/T11两种顺序竞争仍为NOT_RUN；静态接受不表示已经观察到竞争、全部Worker退出或关闭成功。

### 父runtime持有Job与模型前准入的实际代码范围

已读取新增native头文件、完整SwaProviderProcessTree、bootstrap及原生连接/关闭消费方。每Worker创建未命名且不可继承的真实Job；N-API external按当前environment/进程和自有集合核所属，查询实际限制、完整成员及PID/创建时间。分配前核root实际创建时间、父进程和空Job；父runtime保留Job与root进程句柄，SDK不取得句柄。

实际子进程使用三条pipe和Node IPC；bootstrap准入前只加载Node内置生命周期代码，父确认成员后发送一次准入，bootstrap以自身PID/父PID回显收到的准入，之后才动态加载SDK。真实Pi请求等待该回显；原生连接在root退出后继续查询/清理Job，只有root和tree共同exited才发布已验证退出及Pi结算。重启旧runtime_epoch拒绝send/stop/release，未知句柄继续保留原资源。上述为源码行为，native编译、Windows实际分配和模型前无副作用仍待验证。

### F09：已创建root的Job分配失败仍可成为tree exited

严重程度：较高，影响J01/J04/J05及T03/T09。准确f0的`third_party/orca-v1.4.230/native/swa-provider-job.h:116`先保存真实root进程句柄、PID和创建时间，第119行AssignProcessToJobObject拒绝时抛出，assigned仍false。TypeScript的`swa-provider-process-tree.ts:76`在调用前已保存this.root；第55行捕获准入失败并关闭stdin，bootstrap可以实际退出。

该状态下，reap的第140/146行只组合root exit、root句柄已signaled、空members和旧身份缺失，两次满足后第150—152行即可关闭Job并返回tree exited；成功条件没有要求该root曾成功分配并被实际Job成员查询观察。native close也只核成员为空及root已退出。准确`pi/rpc-session-adapter.ts:111`随后将root/tree exited转为AgentSessionAcquisitionExitProvenError，跳过第114行的retainFailed。

因此一个**已创建root、从未成功加入目标Job**的分配失败可沿成功全树证明路径结算，违反已审§5.4“尚未绑定的空Job均不构成通过”及缺证保全。53ad只增加跨自有Job重复绑定拒绝，没有修改这一reap条件，问题仍在。这是源码推导的失败路径；当前没有动态复现，也没有据此声称已发生后代泄漏或模型调用。

原Worker需保留实际分配及真实成员见证这一必要条件；已创建root的分配/绑定缺证保持unknown及原资源，不能借未绑定空Job获得全树通过。真正spawn未创建任何进程的processless预输入失败，应按实际无进程事实单独处理，不混入已绑定Job的全树证明。沿T03/T09检查相应实际失败，不增加模拟回执、管理员或全机追踪要求。具体调用链、行号及范围已通过`msg_331f32dfa9f5`交原Worker。

### CLI、成果、构建安装及53ad测试差异

已读取新CLI/serve/Windows资源采样、服务组装、RuntimeClient消费、真实成果收集、构建/检查/安装脚本及契约/依赖声明。查询使用明确回环HTTP端点及现有项目token，不隐式启动后台；App仍核当前Grant、目录和固定context。成果收集沿真实原生Task/Worker/transcript分页，缺原件或超限失败，usage未暴露时记录unknown。私有服务使用明确安装/数据根、生命周期stdin和实际readiness；构建导出独立依赖闭包，使用固定官方Node及对应native编译输入，安装只复制到不存在的新目录并设置当前用户ACL。

这些入口目前没有构建或运行结果，不能据源码宣称安装闭包可用。锁同步、OpenAPI生成类型、全部T00—T15仍按Worker声明保持未完成；§3.1旧安装布局与当前flat源码布局的文档一致性也已交原Worker同步，没有替其修改实现或规格。

53ad让CLI在HTTP传输前调用App同一ValidateOrcaCommand，区分worker.read默认100和显式无效0，参数解析诊断不混入JSON；新增测试只使用无效输入核预传输拒绝。普通进程fixture使用同候选Node、真实bootstrap、ManagedProviderProcess和SwaProviderProcessTree，实际创建root/child/grandchild及无关自有Job，编写正常、强制、root异常和最后派生等路径，以及真实句柄/创建身份/重复绑定负向检查。它没有冒充真实Agent；仍须实际执行后才能判断这些断言是否成立。

53ad明确记录父宿主失联、实际SDK abort后关闭及缺native预输入三项未完成；验证入口遇剩余项返回6，不把局部fixture成功写成T09整体通过。当前没有正式PR，尚不能发布针对对应PR准确Commit的GitHub COMMENT review；该步骤随原Worker正式交付继续。

### 实际检查及剩余窗口

产品差异`git diff --check 6168b4ea f0ba5503 -- . ':!third_party/**'`及`git diff --check f0ba5503 53ad6e14`通过。已安装marked18.1.0分别检查f0三份产品Markdown、5本地链接，以及53ad两份Markdown、5本地链接，均无断链、锚点或HTML待审项；Node512MiB、BelowNormal。f0合入的五份主控文档与准确13974对象比较通过。

主控`msg_548c85fec44e`记录prepare-02因当时可用内存3.529GiB在停服前拒绝，原S12800继续运行，未交依赖/生成/构建窗口。本Reviewer没有运行prepare、资源采样、原生fixture、Go测试、构建、安装、SDK、模型、HTTP或服务；所有功能、并发、停止和成果动态验收保持NOT_RUN。

本轮报告`f911c36c12a6eab0415a973d453b307ae89b7873`已普通push；f0和53ad的决策hook分别为`msg_ce3651740f1a`及`msg_2b097b9e4660`。主控在`msg_577f0cbfcd21`明确完整读取新增准确报告、F08接受/F09退回的源码因果及静态与动态范围，确认修复交原Worker、processless分别处理、原S和无窗口保持；`msg_d5ed76bd51bf`另对53ad hook确认完整读取同一报告及九路径范围，两次实际读取回复确认本轮唤起。

## 完整成果读取的后继复核条件

主控`msg_f946e03389dc`转交原Worker发现的原生预览限制，要求随F09后继复核。以下沿H09及既定完整成果目标形成，不新增模型任务或动态窗口，尚未收到对应准确修复提交。

已独立读取准确f0的`worker-transcript-payload.ts`：每条消息最多六个block，文字block最多1200字符，裁剪使limited为true。`worker-transcript-read.ts`合并分页限制与payload裁剪；实际structured路径`orchestration-structured-worker-lifecycle.ts`的readStructuredWorkerJournal先裁剪，pageMessages把该limited继续带入每页，消息已读完也不清除。普通消息cursor不能恢复同一block已裁去的文字。当前task.result循环仅见limited时继续分页，后续无消息或cursor不前进即报不完整，不能由此获得长消息的完整原文。Worker对此静态问题的发现成立，当前没有动态复现。

| 条件 | 后继必须保持的结果与失败路径 |
| --- | --- |
| 原件所属 | 从真实Run/root Task/Dispatch和provider session绑定取得原件；不接受正文任意路径或另一会话。实际路径、reparse point、固定上下文与私有根仍核验，缺绑定或来源改变立即失败 |
| 完整内容 | 保留完整消息及原件来源，区分消息分页与单消息裁剪。空文件、损坏/未完整写入、缺原件或超出已声明读取/输出上限时明确失败或缺证，不能靠非空预览宣称完整 |
| 只读与格式 | 复用成熟库或准确SDK公开解析接口；读取原件不迁移、修补、创建会话、调用模型或刷新认证。解析库若跳过坏记录，该行为不能自动代表原件完整；不得为规避依赖手写JSONL parser |
| 真实结算 | 导出保留A被停止等实际状态及B的原会话成果，不将文本内容当成功状态或用量。T13仍为同一安装候选、真实A/B重叠、停止A后B继续及完整成果；当前不登记通过 |

已只读核已安装Pi SDK1.0.1的公开index及session-manager实现：SessionManager.open最终使用persist=true，旧版本可迁移并重写原文件；loadEntriesFromFile在尾部pending非空时会追加换行，且缺文件或坏行有宽松处理；公开parseSessionEntries只解析传入内容，但同样跳过损坏行。接口名称不足以证明只读或完整。上述来源事实与条件已通过`msg_2ec478e8ce85`及`msg_698e72686772`交原Worker，具体方案及实现继续由原Worker完成并交准确版本审核。

主控在`msg_5a47a744a9c1`明确完整读取累计`f75bb5959bb409d116208c437600d4d8ce6f401a`两报告差异、真实预览及SDK来源和上述四项条件，确认仅静态依据、返修交原Worker、没有新接口/模型授权或动态窗口，原接受及F09退回范围保留。

## 准确f8c69161的F09返修静态复核

结论：**接受准确`f8c69161a6a3bb030c730cdc0c9ff6229b8c95e6`对F09的静态修复**。按主控`msg_866051f4a56c`完整读取相对53ad的全部十二路径差异，真实远端分支同SHA；没有用原Worker继续编写的完整成果或测试文件替代这一准确提交。F08及此前接受范围保持，整体产品和全部动态结果仍证据不足。

SwaProviderProcessTree只在native assign成功、实际查询assigned/root身份一致且成员含真实root后设置assignmentWitness。已创建root但无该见证时，reap直接返回unverifiable，保留外部对象及Job，后续空Job/root退出不能改变该结论；正常退出成功条件还同时要求当前assigned和已保存见证。JsonlRpcAgentConnection对这种已创建进程继续屏蔽单独root exit，Pi acquisition仍retainFailed并返回ExitUnproven，修复已覆盖原F09的调用及消费因果。

真实processless沿原生spawn error加无PID的close事件证明从未创建进程。releaseProcessless再核无root/成员/既往身份、Job未分配后才清理空Job；ManagedProviderProcess返回root exited/tree null。连接与Pi只在自身实际processless时接受该无进程结算，未将其转换为已绑定Job全树证明；已创建root的分配失败不能使用该分支。

文档§3.1现与CLI/adapter/构建一致使用平铺安装路径及真实controller-helper.cjs入口；安装器在复制前建立独占新目录、设置当前用户私有ACL，再复制随包内容。新增构建entry服务实际连接/认证检查，未改原上游、本机Orca、执行账本或权限规则。上述是静态一致性核查，安装ACL和可用性尚未实际验证。

普通Node fixture新增：已真实退出再绑定的无成员见证保持unverifiable；不存在程序导致实际无PID/processless且tree为null；父宿主实际退出后两个新鲜进程身份快照及无关Job保全；缺conpty.node的真实私有候选副本在spawn前拒绝。父宿主失联记录明确保留oldJobTreeVerdict unknown，未以父退出代替旧Job证明。这些源码对应既定J/T09条件，未发现本次新增必须删除的无关检查，尚未运行，不能据断言内容登记通过。

完整成果读取、真实SDK abort后关闭、其余T00/T02/T13/T15准备及全部构建/安装/模型结果均保持原Worker待办；阶段尚未完成不单独判失败。当前无正式PR，GitHub COMMENT review仍待正式PR及准确交付。

`.verification/check-revision.ps1 -Previous 53ad6e1479ef5cf3099372d2855087cf002513bb -Candidate f8c69161a6a3bb030c730cdc0c9ff6229b8c95e6` EXIT0，十二路径Git空白、两份Markdown和5本地链接通过，无锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。没有编译、加载native、执行fixture/Go测试、启动SDK/模型/产品服务或安装；原S12800与无窗口保持。

准确f8的正确报告路径hook为`msg_4dfd4ffda591`，补正前通知`msg_cfda7e4b89c7`的报告路径。主控在`msg_afc9334326f4`逐项确认完整读取本报告cc4d相对f75的全部新增范围，并在`msg_9ee6e0885665`再次确认实际读取的正是此文件及正确路径hook；F09仅静态接受、阶段待办和无窗口保持，实际读取回复确认本次唤起。修复接受已通过`msg_0b3890b3b33c`送回原Worker。

## 准确f295的完整成果补充计划审核

准确来源为`f295f7c9360a31e6c1ac4ade2ab9c090c455ddae`，按原Worker`msg_31f127304b2b`及主控`msg_fab0d6773558`读取相对f8的全部两份文档差异；新增§5.5与T07/T13补充尚未实现。独立条件先见已普通push的f75报告，未使用Worker解释替代需求判断。

结论：**退回本次补充计划中T13计划停止A的未持久化输出分支**，其余只读完整读取方法有据，可以按修订继续；原f8 F09静态接受保持。当前尚不授予新增读取路径实施接受，原方案其他独立实现及静态工作可继续。

§5.5覆盖真实controller/Run/root/Dispatch/structured incarnation到host record/provider link，再到私有SDK原件的来源；header/cwd、路径和读取前后身份/fence一致，严格UTF-8/JSONL、格式/父entry、大小/数量门限及原生journal完整性都有对应失败要求。已独立读取实际安装stream-json3.7.0及stream-chain的JSONL factory：正常严格路径直接传播JSON解析错误，ignoreErrors或errorIndicator会改变拒绝语义，实施须保持严格配置；库本身不替代计划声明的UTF-8及尾部完整性检查。

已核准确原生host deps公开store/journalDatabase和workspace解析端口，replayJournal可报告缺journal、damage/newer，renderJournalState及原生row读取保留真实epoch/seq/ts。方案复用这些既有接口，SDK不使用会补写或迁移的读取入口，符合四项独立条件。对应T07对真实会话长消息及自有归档副本损坏、截断、非法编码、替换和超限的检查可从H09直接推导，没有新增模型任务或无关检查；全部NOT_RUN。

缺口来自既定T13：A可在真实并发后按预定条件被停止，其SDK尚未持久化完整assistant输出属于允许的流程，B仍需在原会话继续并交付完整成果。f295一方面要求“缺原件……均明确失败”“不输出部分成功”，另一方面只说保留A的journal和SDK持久化区别，尚未规定该正常分支的返回及根Task判据。按全体源必有完整内容实现会阻止取得B成果，追加A完成或持久化完整输出要求；把空或缺失A源当完整又会产生错误通过。

原Worker需明确：A保持真实停止状态及实际journal已有部分，SDK没有已持久化输出或不可用时保留可核事实和未知原因，同时允许读取同任务B的完整成果；根Task只按预先允许A停止的既定条件确认。B要求的成果缺失/损坏/截断仍失败，A已有坏源、错误归属或来源改变也不能归类为正常未持久化，不能自动换源、重放或生成占位成果。T07/T13须直接检查该正常A分支及相应错误边界。具体缺口已通过`msg_b1c6dc6b3c54`交原Worker，主控本轮要求与此边界一致。

`.verification/check-revision.ps1 -Previous f8c69161a6a3bb030c730cdc0c9ff6229b8c95e6 -Candidate f295f7c9360a31e6c1ac4ade2ab9c090c455ddae` EXIT0，两份Markdown、5本地链接及Git空白通过；marked18.1.0、BelowNormal、Node512MiB。库和原生接口只读源码核查，没有执行parser、SDK、SQLite、模型、fixture、安装或服务；无新动态窗口。

旧准确f295的退回hook为`msg_d583c0bbc776`，明确只记录旧源，后继d7另行复审；主控在`msg_8d860ec22854`完整读取faaa报告的缺口、原生/成熟库来源及实际范围，确认未将旧源退回套用到新源，实际读取回复确认本次唤起。

## 准确d7f21242的完整成果补充计划复核

结论：**接受准确`d7f21242259739cfb49a4f68a19a1a8f28bcec32`的§5.5及T07/T13补充计划**。已按`msg_caba0e4d1b44`和`msg_d03579e2e05f`读取准确76e9→d7的唯一文档差异及修订后的完整§5.5；真实远端分支同d7。新增读取路径尚未实现，这一接受只允许原Worker按已审补充实施，尚无功能或动态接受。

d7明确A真实计划停止时分别返回实际stopped/failed原因、Job/资源状态、journal已有部分和SDK可用性；SDK无完整assistant条目为incomplete，缺文件保留实际ENOENT及unavailable，未知原因继续unknown，不声称从未持久化。A这些限制可与B完整成果一起导出，B缺失、损坏、截断或不完整仍使最终成果验收失败；A已存在坏源、错误归属、读取期间消失及已观察来源改变也继续报错。它补上f295所列正常分支，没有新增A完成条件或占位完整成果。

根Task仍要求实际计划停止A、B完整成果和固定业务判据，以及全部相关资源退出/归档/释放。T07/T13明确实际A未完整持久化时B仍可取得，以及坏源/来源变化不被归为正常未持久化；严格只读、成熟JSONL、原件绑定及12MiB完整导出门限保持。对应独立条件没有剩余计划缺口，也未发现应删除的多余用例；将来仍须按真实原件和准确代码审核，不能由计划接受登记通过。

`.verification/check-revision.ps1 -Previous 76e9dafd12f4ff430a3082ff4a65ec543d1e02f8 -Candidate d7f21242259739cfb49a4f68a19a1a8f28bcec32` EXIT0，唯一Markdown、1本地链接、Git空白通过，无锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。未重跑已审库或产品检查，无动态窗口。

准确d7计划接受hook为`msg_2348de1f3c82`；主控在`msg_071202f7e0e4`确认完整读取99d084新增报告，逐项核A限制/B完整成果、错误分支、原生结算及严格只读来源，确认只接受补充计划，仍无动态结果或窗口；实际读取回复确认本次唤起。原Worker收到`msg_90d477914ad6`接续实施通知。

## 准确76e9dafd的授权标识、安装入口与SDK检查源码

结论：**接受准确`76e9dafd12f4ff430a3082ff4a65ec543d1e02f8`的相应静态修复及检查源码范围**。已完整读取相对f295的全部七路径差异；准确76e9是d7的父提交。完整成果方案与实现分别沿上一节和后继处理，整体产品、真实SDK与安装结果继续证据不足。

Runtime.Target改为现有注册标识pi。已核同一准确代码：SetProjectSettings通过NativeRegistry.Adapter校验ExternalModelCLI，Registry实际注册pi，ValidateProjectActionScope随后要求target与批准值相同。原绝对PiEntry路径无法通过该设置准入，新标识使已有真实批准路径一致；它没有证明Pi已安装/可用，也没有签发Grant。执行仍使用配置的同候选Node/SDK绝对路径及固定provider/model，未从PATH挑选其他执行目标，原human/Agent权限规则保持。

构建复制真实install-swa.ps1和USAGE.md到候选根，闭包检查要求这些交付文件存在；CLI help列出实际必需的auth/protected参数及stdin生命周期。USAGE说明新目录、保护根、既有human项目批准/token、pi标识、固定输入和原生request查询、停止及释放，不把服务启动当授权签发。源码形成了随候选安装和调用入口，实际构建、安装、私有HTTP项目授权和命令可用性尚未验证。

新增SDK检查使用同候选Node、真实只读OAuth适配、真实受限SDK和实际Job；代码核准备会话的tools/cache/retry/compaction/认证操作、非法写回调不执行、真实get_state，以及仅允许get_state/abort时禁止prompt。已核真实JsonlRpcPeer成功响应解出data、Pi1.0.1 get_state字段与断言对应。该源码明确isStreaming=false、messageCount=0、model_calls=0，其abort为**空闲SDK的真实RPC关闭检查**，不能证明执行中模型取消、正向模型调用或双Worker停止；T13实际A/B仍须单独通过。

verify入口在普通进程fixture后重新资源准入并调用该SDK检查，保留两个命令的实际退出及剩余分支；没有在源码阶段执行这些命令。`.verification/check-revision.ps1 -Previous f295f7c9360a31e6c1ac4ade2ab9c090c455ddae -Candidate 76e9dafd12f4ff430a3082ff4a65ec543d1e02f8` EXIT0，七路径Git空白及USAGE成熟Markdown解析通过，无本地链接/锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。当前无正式PR或动态窗口，所有测试、构建、安装、模型及服务仍NOT_RUN。

准确76e9接受hook为`msg_4a9a79825778`；主控在`msg_fe6bcb71dc35`确认完整读取同一99d084报告的七路径、target及安装入口和空闲SDK检查范围，确认只接受静态修复与检查源码、未宣布SDK或安装可用；实际读取回复确认本次唤起。

## 准确0c6c701d/f2bcba97的完整读取及保全代码CR

准确程序来源为`0c6c701daa0ab29995b34d657c463fb2c33d137c`，其后普通merge为`f2bcba97b475b348bbb60b47dc3dd605a3f40739`，父提交分别为0c6及已审main4b。按原Worker`msg_1a88202db278`和主控`msg_bdb5b576d13f`完整读取0c6相对d7的全部十一路径；0c6→f2只有四份主控文档变化，对应blob与准确4b完全相同，程序和其余路径相同。真实远端Worker分支为f2。

结论：**退回准确f2中来源0c6的F10—F12**。d7完整成果补充计划、F04—F09静态修复和76e9已接受范围保持。整体产品及实际结果仍证据不足；双Worker驱动、故障副本用例及其他阶段待办没有被单独判定为失败。当前无正式PR，GitHub COMMENT review继续等待真实PR及准确交付。

### F10：保全观察器没有保存实际事件订阅

严重程度：较高，影响H17及T00/T02/T13/T15。准确`tests/swa-cli-headless/OrcaPreservation.psm1:43`把Register-ObjectEvent的返回放入subscriptions，第44行没有传Action。官方文档明确默认无返回值，使用Action时才返回PSEventJob；本轮已读取该[官方接口文档](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/register-objectevent?view=powershell-7.5)，仅查询本机cmdlet元数据，没有注册或执行观察器。

由此可推导subscriptions为空：第62行不会读取定向事件或Error，第85—86行也不会注销实际注册；第75行仍可能在末态字节/ACL和进程相同的情况下给出passed。监控溢出或通知缺失因而不能正确成为缺证，这不符合已审“目录通知为补充证据、Error不能忽略”的条件。没有动态复现或原Orca文件变化声明。

原Worker需按真实SourceIdentifier保存/取得实际订阅及事件信息，确实读取Changed/Created/Deleted/Renamed/Error、在监控错误时保留证据不足，并清理本项订阅；不增加内核追踪、写入PID推断或全机监控。具体依据和行号已通过`msg_7d52ae511577`交原Worker。

### F11：来源观察的局部process阻断真实正向分支

严重程度：较高，影响H09及真实Worker来源观察。准确`runtime/swa-orca/src/task-result.ts:110`访问process.env.SWA_DATA_DIR，同一observeWorkerSource作用域第116行却声明const process并赋值record.lease.ownerProcess。该局部绑定遮蔽Node全局process，真实record/provider head成立后，第110行会在局部初始化之前访问它而抛出ReferenceError。

因此worker.show新增source观察的正向路径无法返回；缺record的早返回不能覆盖该缺陷。原Worker应使用不遮蔽全局的owner进程身份变量，保留真实PID/创建时间/runtime字段及所属检查。此为源码判断，未运行TS/Go编译或产品查询，已同F10经`msg_7d52ae511577`交原Worker。

### F12：宽松消息结构可把坏SDK源标成complete

严重程度：较高，影响§5.5未知格式/损坏拒绝和T07。准确task-result.ts第24行messageSchema将所有role的content共同放宽为string或未知元素数组；第73—80行仅进一步核assistant的provider/model并排除toolCall，没有校验内容块结构。第89行只看最后role与stopReason即可标complete。

已独立读取实际安装Pi AI类型定义：AssistantMessage.content为Text/Thinking/ToolCall数组，TextContent.text为string等。当前assistant使用普通字符串、未知block，或text块的text不是字符串，都可通过现有宽松条件；若最后stopReason为stop，还会成为availability complete。这里指出的是JSON语法合法但SDK结构损坏的来源，未生成伪造模型输出或运行损坏副本测试，也没有据此宣称根Task已错误完成。

原Worker需依实际受限SDK消息和entry分支校验必要结构，保留真实允许格式，未知或坏结构明确拒绝；继续使用成熟JSONL及schema库，不能手写格式parser或只凭终止字段认定源有效。沿既定T07的自有真实归档副本覆盖坏字段/未知块，A正常计划停止的可用性范围保持。已通过`msg_ad371a15339d`交原Worker。

### 其余已读取范围及真实限制

完整读取代码从原生所属记录取得SDK路径，采用已固定stream-json严格解析、标准fatal UTF-8检查、header/entry父链及前后文件/record/fence/epoch/seq核查，保留真实journal rows与usage并最终检查12MiB；helper不再把有界预览拼成完整成果。源码返回停止Worker的SDK可用性事实，查询成功不代表B已满足最终成果或根Task完成；后继T13仍须核预定A和B完整成果及固定业务条件。本轮未将这些源码描述为已通过完整功能验收。

worker.start增加真实CPU/可用内存双采样，在原生reserve及模型输入前失败；reserve仍重新核原生配额和关闭屏障。实际SDK事件日志只在私有Worker根独占创建，保存SDK原事件、接收时间及Job root身份，写失败沿原RPC错误关闭，不作为执行账本。实际模型重叠、停止A后B原身份继续须由真实T13验证，当前没有流式事件结果。

T00固定公开来源及四原文、T02同候选安装、T13私有服务/缺Grant拒绝和T15旧进程/Orca保全检查都仅形成源码；完整驱动及故障用例继续由原Worker实施。认证配置原件只由将来的实际受限验证定向读取，本Reviewer没有读取原token、环境或SQLite秘密原件，没有执行这些脚本。

`.verification/check-revision.ps1 -Previous d7f21242259739cfb49a4f68a19a1a8f28bcec32 -Candidate 0c6c701daa0ab29995b34d657c463fb2c33d137c` EXIT0，十一路径Git空白、两份Markdown及5本地链接通过，无锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。f2两个父来源、四份主控原blob及其余程序未变已核；没有编译、安装、执行JSONL读取/PS观察器/fixture、模型或服务。原S12800与无动态窗口保持。

准确f2退回hook为`msg_4329ac9d3bc1`；主控在`msg_2903075f2500`完整读取bc273ad无界面报告，逐项确认准确来源/F10—F12因果、官方及Pi类型依据、其余范围和全部动态未运行，确认返修仍交同一Worker，d7/F04—F09/76原接受保留；实际读取回复确认本次唤起。

## 准确abdb07d3的返修及实际检查源码CR

准确来源为`abdb07d31954b9e72f12bdaafdfe56406deb9004`，父提交f2。按原Worker`msg_20ee3dc6b88f`和主控`msg_80224e5e2ce8`完整读取十四路径；未提交的四个Go检查不纳入本次审查。此前两次TLS推送失败保持原记录，后继Worker`msg_0681a26676c2`及主控`msg_4f5e4e394ca4`报告普通push和远端核对成功，本Reviewer另以真实ls-remote核远端Worker分支同一准确SHA。

结论：**接受F10—F12的静态修复；退回T13的F13及T07的F14检查缺陷**。整体程序、构建、安装、真实SDK/模型和双Worker功能仍证据不足，全部动态NOT_RUN。d7正常停止A/取得B完整成果的计划、F04—F09及76原有静态接受保持；尚未完成的其余阶段没有被单独判为失败。修复继续交原Worker，当前无正式PR，准确Commit的GitHub COMMENT review等待正式交付。

### F10—F12静态返修接受范围

F10现按实际SourceIdentifier取得Get-EventSubscriber返回的PSEventSubscriber，核对五个EventName和SubscriptionId、启用状态及活订阅，逐一读取Changed/Created/Deleted/Renamed/Error；监控Error失败，最后注销准确订阅、清理事件并释放watcher。另在自有探测目录创建真实文件并等待实际通知，检查观察机制可用。本轮只读核[Get-EventSubscriber官方文档](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/get-eventsubscriber?view=powershell-7.5)，返回类型和按标识读取方式相符；没有注册观察器或运行探测。

F11把局部process改为ownerProcess，移除同作用域访问process.env前的遮蔽，保留真实owner PID/创建时间/runtime来源。原正向分支的ReferenceError源码因果已消除，尚未执行该查询。

F12按实际Pi AI1.0.1类型区分user/system/assistant及允许entry分支，核必要timestamp、usage、provider/model和内容块字段；assistant字符串、未知block、非字符串text及未知entry不再通过原宽松schema。完整成果还要求实际末条assistant正常stop、没有错误及非空text。工具调用仍拒绝，实际SDK usage保留原件来源，billed_cost明确未知。原反例在源码上已被拒绝，真实合法SDK成果和故障副本尚未运行。

### F13：T13不能证明原生处理停止时A仍在执行

严重程度：较高，影响首版同一安装候选的真实并发与执行中停止验收。准确`tests/swa-cli-headless/installed-flow.mjs:144`先检查A尚无agent_end，第146行发起CLI停止，第151行只要求事后agent_end接收时间不早于客户端命令started；空数组的every也返回true。

存在可从源码推导的顺序：客户端停止命令已启动，A随后自然结束，原生停止随后才处理并关闭空闲SDK。现有条件全部可成立，最终仍能写functional_flow passed。准确原生StructuredWorkerStopOutcome的stopped证明会话关闭及lease消失，空闲Worker同样可以满足，不能补足执行中取消的因果。这是静态可行分支，尚未动态复现。

原Worker需关联实际A目标turn及原生处理停止/SDK中止的真实证据，使自然结束先于原生停止的分支保持失败或证据不足；复用已有journal/SDK事实，不新建执行账本或重放替代任务。已读取实际JournalStopAcceptor：原生接收时在writer事务内保存live turn及epoch/sequence，可用于实际关联。d7允许A没有完整持久化输出，不能因此追加A完成条件；缺少取消证据时继续未知。B仍须保留同一身份、继续执行和交付完整成果。

### F14：T07八个结构损坏用例首先命中尾部截断

严重程度：较高，影响H09、§5.5及F12返修的实际回归覆盖。准确`tests/swa-cli-headless/result-sources.test.mjs:104`对八个结构用例调用无选项stringer.asStream重写真实归档副本，第105行只作无错误谓词的assert.rejects。

已读取实际安装stream-json到stream-chain的stringer调用及`src/jsonl/stringerStream.js`：默认separator为换行，suffix为空，末条后不输出换行。产品readSDKSessionSource先检查尾字节必须为10，然后才解析JSONL及schema。这八项unknown_content_block、non_string_text、assistant_string_content、missing_message_timestamp、missing_usage_tokens、unknown_entry_type、missing_model_change_field及empty_final_content均会先被统一尾部错误拒绝，无法据现有断言证明各目标结构错误被检查。

原Worker需用成熟stringer配置形成完整尾部，保留真实合法原件对照，并断言各目标语义错误或准确错误分支，避免截断等无关错误满足断言。上述八项本来直接来自已审H09，没有新增无关用例；本轮未生成模型原件、运行损坏副本或执行parser。F13/F14及三项返修接受依据已通过`msg_c2f04f4bd2b2`交原Worker。

### 其余已读取范围和验证结果

T13驱动固定同一候选Node/SDK/version、真实公开原文、原生项目批准和Run/root/Dispatch；实际SDK delta重叠、真实owner根身份、停止A后B同PID/创建时间/fence继续、真实journal逐事件定位及完整B导出均形成检查源码。业务成果仍须原Worker实际执行后由独立Reviewer核语义与普通终端证据，不能凭程序的文本门槛或模型自评登记接受。实际A限制与B完整成果、根Task及资源最终状态均保持原计划要求。

只读JSONL辅助使用成熟parser、fatal UTF-8及读前后文件身份/大小/时间和完整边界，原件故障用例基于实际T13归档的自有副本。构建复用上游slotSourceFiles复制候选所需ConPTY文件，conpty.node保留本次实际构建来源；闭包检查增加对应真实文件，尚无安装可用性结果。实际SDK准备及agent_start检查固定xhigh。PS入口保存自有进程树身份和EOF后两次实际缺席观察，未终止原S或Orca。

`pwsh -NoProfile -File .verification/check-revision.ps1 -Previous f2bcba97b475b348bbb60b47dc3dd605a3f40739 -Candidate abdb07d31954b9e72f12bdaafdfe56406deb9004` EXIT0，十四路径Git空白、两份Markdown及5本地链接通过，无锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。此前误用Windows PowerShell的两次辅助脚本调用分别被执行策略及UTF-8脚本编码阻断，未执行产品检查；随后改用已安装PowerShell7，没有更改系统配置。

原S12800及无窗口保持。本轮没有编译、安装、运行Go/Node测试、SDK/模型、JSONL产品读取、HTTP/SQLite或服务，也未读取原认证秘密、环境或SQLite原件。所有适用动态检查仍待原Worker在主控授予真实窗口后执行。

准确abdb退回hook为`msg_16d3b555e2f2`；主控在`msg_723b52740787`完整读取a211b66相对629e的四十行新增，逐项确认三项静态修复接受、F13/F14的源码及库因果、其余已读取范围和宿主调用失败/成功及动态未运行，确认问题交原Worker、既有接受及A正常分支保持。实际读取回复确认本次唤起，没有新增动态窗口。

## 四包Go检查补充的独立条件

收到准确39cc交付后，先于读取其§7.4及新增Go函数，沿H03—H06/H10—H15推导本次检查条件：

- 实际测试针对本轮Go消费方及真实授权、所属、请求状态和资源生命周期，直接调用真实实现或同一候选服务；适用拒绝分支须核具体效果，不能由统一任意错误满足。未使用的旧S3/MCP/网页路径不新增为前置。
- 无服务编译只代表编译结果；四个准确包、实际测试名称及匹配过滤可核，零匹配、SKIP及未运行不能计通过。真实执行使用同一安装候选和私有状态、实际HTTP/SQLite，不能借用原S/本机Orca或伪造服务/协议。
- 本次补充不替代T13两个真实Worker并发、执行中停止和完整成果；纯Go调用、已创建PID、关闭回执和任意非空输出都不能代替对应事实。F13/F14仍按已写明的准确因果复核，A允许没有完整持久化输出。
- 编译和实际服务检查按各自唯一窗口及既定资源门槛串行；实际身份/批准及错误保留，原件与其他进程保全。缺窗口/前提时保持NOT_RUN，计划不能通过静默跳过或降低门槛宣称通过。

## 准确39cc967d的F13/F14及检查补充复核

结论：**接受准确`39cc967df6a71b2c0c22e3daa520ca08331db17e`的F13/F14静态返修、四包Go补充计划及相应检查源码范围**。按原Worker`msg_941f578c27bc`及主控`msg_7a78cac25840`完整读取相对abdb的十七路径；本Reviewer真实ls-remote核远端同一准确SHA。以下为源码结论，整体产品与全部动态验收仍证据不足，原有接受范围保持。

F13新增实际A journal中的唯一evict Stop行，要求原生turnId及客户端命令开始后的实际Stop时间，按同候选原生startJournalRowFold还原标记前全部行，要求无damage/newer且目标turn仍running。实际SDK agent_end若在该标记前已被观察即拒绝；最终要求该同一turn为原生interrupted且无其他outcome，或真实cancellation并有对应时间后的SDK assistant aborted事件。最终回执保存Stop行、前后turn及实际SDK中止证据，空结束集合本身不再满足接受条件。

已独立读取原生host-lifetime、host-stop-event、journal Stop行构建、fold及turn-end规则：host先尝试在一秒内排空已收到的生命周期事件，失败/超时仍按当前工作事实判断，保存实际活动turn，行的at来自原生writer时间；新检查另核标记前的实际fold及SDK结束事件，空闲关闭本身不满足条件。evict保留原生中断事实，不自动改写为人的cancellation；提供方已有完成判据保持。新检查同时核标记前running及标记后中断事实，拒绝原F13“原生处理前已自然结束”的分支，没有把所有evict都称为模型取消，也不要求A生成完整SDK成果。B先结束原轮再在原身份接收后续输入，原PID/session/fence保全继续检查。实际A/B运行仍未发生。

F14现以成熟stringer的suffix换行形成完整尾部，核尾字节及成熟parser往返；新增合法round_trip_valid必须取得complete并与真实原件entries相同。空最终内容要求准确不完整错误，其余七类要求ZodError及对应type/modelId/message timestamp/usage totalTokens/content text等路径。统一尾部拒绝或任意错误不能满足这些断言，原缺陷的库到产品检查因果已补；全部十九个实际归档副本场景仍未运行，原件来源及保全沿既定T07要求。

四包方案与前节独立条件相符。安装CLI检查真实VERSION、controller/runtime；adapter检查真实原生ResolveScope及保护根重合拒绝；App检查当前真实项目批准取得同一runtime及私有根的实际PRAGMA database_list；HTTP检查缺凭据和禁止自行传human/runtime字段。没有伪造服务、provider、human或批准，没有将四函数扩大为全部T03，撤销后的安全停止、另一真实项目/Run和旧代仍明确未验。

compile-go-tests在干净准确源码与候选VERSION一致时，以Go四并发/512MiB逐包test -c，temp/cache/module cache放在本项新目录；仅传输前的三个真实CLI拒绝函数在编译窗口执行。服务窗口由同一候选启动实际私有HTTP/原生后台，以真实批准创建controller后串行运行四个预编译二进制，逐项要求准确PASS、拒绝SKIP/零匹配，最后关闭该controller。无实际环境时普通Go的Skip明确不计验收；T03本入口完成局部范围后仍返回证据不足。计划没有编译/服务并发或凭编译结果宣布功能通过的缺口，未发现需要删除的无关检查。

完整成果新增只读SQLite连接复用上游SyncDatabase，已核该适配器把readonly传给Node DatabaseSync.readOnly、fileMustExist先检查存在；不再访问可能修复/初始化的host db getter。原生replay及读取最终epoch/seq仍核，SDK读取新增当前路径文件身份检查。实时事件辅助仅取实际append日志在读取开始时的有界完整前缀，核当前/句柄身份、没有缩短及历史已观察行未变；完整SDK成果保持原严格读前后大小/时间条件。这里没有打开SQLite或运行JSONL产品读取。

T02实际端口占用、保护目录及本项junction拒绝只形成源码：占用端口在serve的资源准入及子进程启动前被真实bind拒绝；保护根由真实privateDataDirectory在创建home/temp前校验。本项junction清理核本项绝对路径及ReparsePoint，仅删除该链接。安装后尚未验证的服务消费者继续标requires_T03_T13并返回证据不足，不能独自登记安装验收。资源采样器现在随源码跟踪，阻塞psutil CPU和可用内存双采样、单调2—5秒、85%/4GiB及局部资源参数符合RESOURCES；未采样或执行正文，测试Python不成为产品安装依赖。

`pwsh -NoProfile -File .verification/check-revision.ps1 -Previous abdb07d31954b9e72f12bdaafdfe56406deb9004 -Candidate 39cc967df6a71b2c0c22e3daa520ca08331db17e` EXIT0，十七路径Git空白、两Markdown及5本地链接通过，无锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。仅执行Git/文档检查和只读源码追踪，没有运行Go/Node产品正文、编译、安装、SQLite、SDK/模型、服务或资源采样，原S12800和无窗口保持；无正式PR，COMMENT review仍待准确交付。

准确39cc接受hook为`msg_fd2dbcd8953b`；主控在`msg_768ba04e2f3b`完整读取2bd相对db704无界面报告二十九行新增，确认上述静态接受、独立条件及实际来源、未执行限制，9741及后继与新补充计划另审；实际读取回复确认唤起。

## CLI输出及owner生命周期补充的独立条件

先重读已审§6和实际main/serve/传输前测试，再于读取准确b3的§7.5前，按H05/H06/H08/H10/H11/H15/H17形成以下条件。主控和Worker报告的新缺陷仅作线索，后继具体方案尚未接受。

- JSON调用在stdout仅发布一个完整、可解析且含实际ok/身份/结果或错误的envelope，stderr承载诊断；本地参数/编码/路径/缺批准、真实HTTP应用拒绝和原生结果均覆盖。保留原生错误字段及失败任务查询EXIT0的区别，不能把ErrorV1透传为另一种顶层契约。
- 已知合法原R在失败和取消中保留并可用于原请求查询，未获知或坏R如实未知；HTTP trace只表示HTTP来源，不冒充mutation R。解析在任意阶段失败、读操作没有R或不同scope时，不猜测后台身份、不生成替代请求、不自动重发。
- 实际网络响应不完整/断开、真实用户取消130及stdout管道关闭分别检查实际退出和未知效果。以真实连接和本项子进程制造条件，未命中保持证据不足；输出管道已坏时不能要求仍成功写JSON，也不能把写失败或客户端退出当后台已停止。
- serve实际ready前失败只发布失败，ready已发布后不再写第二个stdout结果；其后诊断及实际退出仍须准确。readiness证明服务当时就绪，后继退出/错误和资源结算不能被既有ok覆盖。
- 实际owner stdin EOF在启动前、启动中及ready后均受处理，明确正常关闭与用户取消/失败退出语义；已发生的自有node/server/后代须逐项证明退出。只有root退出、没有创建身份、清理超时或后代未知时继续占用，不终止原S/Orca/IDE或以其他进程补足证据。
- Windows取消须先证明本项独立console、真实新组及成员PID/创建身份，定向到已证实的本项子组；仅CREATE_NEW_PROCESS_GROUP不足以证明console隔离。禁止组0广播、附着或修改原console；隔离、真实批准或窗口缺失时保持未运行，普通CLI取消不代替T13真实模型停止保全。

## 准确b3f0e229的§7.5补充计划审核

结论：**接受准确`b3f0e22919fbbfab155af0fa9990602f50922117`的CLI输出、原R及owner生命周期补充计划**。按原Worker`msg_2ead377afef8`及主控`msg_93123453bc57`完整读取相对8220的两文档差异和全部§7.5，独立条件先见前节；程序没有随本提交修改，8220及其他后继修复仍分别准确CR，不能由本次计划接受算作已实现。

C01/C02/C03覆盖真实本地解析/缺批准、HTTP ErrorV1和原生结果。当前命令已解析且校验的R与真实HTTP trace分开，未知身份省略；既有CLI字段承载结果，完整ErrorV1保存在error.data，没有把HTTP trace用于mutation关联。已读当前apierrors结构的retryable默认false；计划仍保留真实HTTP原件字段及原生结果来源，实施时不得用统一错误覆盖真实code/result/data。成熟Decoder核恰好一个stdout对象及EOF，诊断留stderr，与§6相符。

C04/C05/C09分别采用实际产品HTTP处理中的断开、本项定向CTRL_BREAK和真实stdout读管道关闭，不伪造HTTP响应、抛错Writer或协议。实际窗口未命中为INCONCLUSIVE，不自动重试；正常取消130保留原R及原查询方式，不改变后台Worker/资源状态。stdout确实不可写时不再要求成功输出JSON，保留已知关联及实际影响到stderr、非零退出，避免第二次发布或重发。

C06—C08明确ready前/启动中owner EOF的取消启动130、未就绪失败envelope及未知6；ready之后仅保留原一次readiness，失败/关闭另记stderr与实际退出。空闲且真实Worker/helper已按原生流程退出/归档/释放后，owner EOF只有全部所属及已观察后代退出才为0，已知子进程失败7、未知6，残留记录和占用保持。阶段必须实际命中，不能把serve根退出或另一阶段当覆盖；未验的其他运行/重启边界仍沿原计划，不能由这组CLI用例代替T13模型保全。

已完整只读本机Go1.27.2的VERSION、signal_windows_test.go及doc.go Windows节；真实官方测试用新组和目标PID发送CTRL_BREAK，Notify转为os.Interrupt，原测试没有建立独立console。本Reviewer另实际读取[GenerateConsoleCtrlEvent](https://learn.microsoft.com/en-us/windows/console/generateconsolectrlevent)、[GetConsoleProcessList](https://learn.microsoft.com/en-us/windows/console/getconsoleprocesslist)及[Process Creation Flags](https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags)官方正文：信号目标须与调用者共享console，组0影响该console全部进程；GetConsoleProcessList返回零或缓冲不足不能当完整成员清单；CREATE_NEW_PROCESS_GROUP与CREATE_NEW_CONSOLE同次指定时前者被忽略，CREATE_NO_WINDOW也不代表存在可发信号的console。

计划已区分本项隐藏console宿主与其中新组CLI，要求实际成员、创建身份及组的创建来源核实后才发送，符合上述API边界；实施不能仅凭创建flag或Go示例宣布隔离。缺console/成员/身份或原批准、窗口继续不发送，不能强制终止代替130。Worker此前网页失败仍保留，本Reviewer这次官方网页读取独立记录，没有把它改写为Worker已成功读取或实际信号通过。

覆盖对应独立六条件，未发现本次计划剩余缺口或应删除的多余用例。所有C项仍NOT_RUN，Go helper仅可先在编译窗口串行形成同源二进制，真实服务期间不编译；不增加命令/公开字段、权限、模型任务或执行账本。当前无动态窗口，本Reviewer没有运行信号、进程、编译、HTTP/SQLite、模型或计划正文。

`pwsh -NoProfile -File .verification/check-revision.ps1 -Previous 8220d851dd998639a0ae3fbf4eff0212dd20ed7c -Candidate b3f0e22919fbbfab155af0fa9990602f50922117` EXIT0，两Markdown、5本地链接及Git空白通过，无锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。本报告新增独立条件、对照及官方链接的成熟Markdown检查也通过。普通push/远端同SHA采用Worker准确交付回执，本轮主控未另查远端的事实保持。

准确b3计划接受hook为`msg_52e19c678dc9`；主控在`msg_9a257fbab657`完整读取d7f相对2bd的三十一行无界面报告，逐项确认六条件、C01—C09、实际官方来源和全部未运行限制，安排原Worker按原写域静态实施；实际读取回复确认唤起，没有授予动态窗口。

## 准确9741至8220的幂等、错误与观察静态CR

按H03/H05/H06/H10/H11分别核准确`9741fccb8f705b057b913e617d9ae522536c1e88`八路径、`2c007de5ec54ce1cdb11a2f6052a116d1dc958c4`六路径、`72c7b8961a5677c7291967448507d3fc64340a4b`四路径及`8220d851dd998639a0ae3fbf4eff0212dd20ed7c`五路径，已完整读取四段差异。来源为Worker及主控实际交付消息`msg_23cf13c9882a/msg_14a13515a505`、`msg_35d491c02287/msg_35ae6f311ebc`、`msg_6065c2fe4f9e/msg_f7e6725f7263`和`msg_43f41e21991d/msg_1547e60aa401`；普通push/远端同SHA采用各自Worker成功回执，主控9741远端连接失败及其余未另查保持原事实。后继WIP和已审b3计划没有混入这些准确程序结论。

### 9741：比较规范化静态接受，F15检查缺陷退回

**接受task.create规范化及实际receipt补充的静态修复，退回F15**。原App每次ReadProjectContext确实生成新UUID/created_at，scope也在首次创建后增加Run/root/context；原executor把它们放入既有hashCanonical比较会使同R误判不同。新分支只作用于swa.taskCreate，复制invocation/context后删除两项采样字段，scope比较保留project/controller/authority；实际command/caller/project及上下文的其余内容、版本和路径保留，传给handler的原参数没有被修改，其他原生方法仍走原分支。

已独立读取实际App双次当前授权、ReadProjectContext、原生controllerScope和controller-helper：每次task.create仍检查真实项目、controller、authority、Run/根Task关联和固定context settings，helper只经既定同controller单任务事务创建Run/root，完成receipt保存其实际身份和子请求。被剔除的scope内容来自这些关联；原件首次UUID/时间仍写入实际根Task spec。缺/坏invocation由实际wrapper的完整invocationSchema及dispatcher成熟schema在executor之前拒绝，task所需context缺失也仍由原handler失败，没有把无类型输入按规范化结果执行。

完成receipt增加实际open的client identity/mutation ID、task两个真实子请求、deliver子请求及close的binding/root/helper身份，支持原请求只读观察，没有另建账本。T06源码保持实际CLI/PTY/SQLite，T13重放原start/stop/send/release并核B恰好两轮，第三席要求明确resource_limit且inventory/实际SDK没有第三项；这些均只形成检查源码，其余竞争、断开/重启及动态范围继续未验。

**F15严重程度：中等，阻断真实T06检查。** 准确9741的`tests/swa-cli-headless/native-idempotency.mjs:126`在改写contextPath、writeFile恢复原字节后，又调用原taskArgs期待EXIT0；准确8220第209行仍保留该断言。真实`internal/adapters/agents/projects.go:154`把ContextFile.Version定义为mtime纳秒加size，ReadProjectContext读取并传入该版本，规范化继续保留files.version。恢复字节没有恢复真实mtime，版本已经变化时原R应得到request_mismatch，现断言与正确语义冲突。

原Worker需调整合法未变来源的成功重放控制与版本变化用例的顺序/预期，保留首次固定context原件检查；不能删除真实version、回填旧mtime或重放替代任务来通过。该静态因果已通过`msg_dd12c01c25ef`交原Worker，没有运行写文件、版本读取测试或动态复现。此缺陷沿2c/72/822继承，其后续局部静态接受不会撤销F15退回。

### 2c：产品拒绝与未知退出映射静态接受

**接受本段错误映射及T08检查补充的静态范围**。dispatcher只对swa.*请求中的真实OrchestrationError保留code/message/data；已核实际admission先解析注册方法及参数，未知方法和schema失败沿原处理，普通SQLite异常及其他原生方法仍进入原来源映射。CLI把operation_in_progress/controller_closing映为6，保留未知效果；没有把它们改为已失败/已退出或授权成功。

T08只在真实私有Orca库持有BEGIN IMMEDIATE锁，实际controller.open要求存储失败7/runtime_error、Run/Task和实际liveTab/liveLeaf不增、该R无receipt，释放自有锁后只查原R。已核getStatus确实提供相应数量字段，executor确实先begin receipt再invoke；这组组合证据针对输入前存储拒绝，单独absent没有承担无副作用证明。输入后存储失败及自然低资源仍未覆盖，Case保持6，F15及整体动态限制保持。

### 72：一次真实所属CLI断开检查源码接受

**接受本段检查源码范围**。仅从真实已完成open receipt和实际PTY root/helper观察副作用，保存本项CLI创建身份，仍同一身份且stdout尚为空才经其ChildProcess句柄终止这个客户端。实际终止成功、非正常退出且最终stdout为空才记命中，不能由发出kill请求独自成立。已返回输出、错过窗口或正常结束保持INCONCLUSIVE，不重复任务、不注入延迟或provider故障。

后续只查原R并比较实际receipt/controller/binding，以新明确关闭请求清理已证实helper、两次新鲜身份观察确认退出。未到可核效果且客户端仍未结算时保存原R和本项身份后失败，继续保留占用；没有杀控制helper、Worker或其他进程来制造窗口。这一真实断开尝试不表示130，也不代替其他Run/输入/重启边界；F15与§7.5未实施限制保持。

### 8220：原请求查询绕过派生pending scope静态接受

**接受本段request.show必要修复及局部检查源码**。App仅该动作跳过ResolveScope，其前后真实project/caller/observe授权及revision检查仍在；输入与其他动作的scope校验未动。client先读取原生原R，再核实际receipt.project_id及给定controller，错配用既有RuntimeClientError返回scope_denied；不启动helper或重发。

absent只返回原生absent，不能推断无副作用；已checkpoint的pending和完成记录都先核所属再输出原件及子receipt。pending尚无可核receipt时，成熟schema拒绝并沿既有nativeError保留operation_unknown，不输出未核业务内容；没有把这种情况登记为成功查询。源码用例仅补完成后project/controller两selector相同和另一真实controller拒绝，pending/重启实际命中仍未验，F15继承及b3计划尚未实施范围保持。

四段`git diff --check`逐项EXIT0；`node --max-old-space-size=512 .verification/check-markdown.mjs 8220d851dd998639a0ae3fbf4eff0212dd20ed7c 39cc967df6a71b2c0c22e3daa520ca08331db17e`在BelowNormal下EXIT0，累计两Markdown、5本地链接及无锚点/HTML待审通过，marked18.1.0。没有类型/功能、编译、SQLite、CLI/进程、SDK/模型或服务检查，所有动态仍证据不足，当前无正式PR及动态窗口，原S12800/Orca保持。

四个准确决策hook分别为`msg_34fa3d66dcb3/msg_282494a27c66/msg_18b6a95d5f58/msg_648e4acd75df`。主控在`msg_5192424ceb6e/msg_68719f80a5ba/msg_190a9123fe49/msg_9ce998174d1e`逐项完整读取a9相对d7f的三十八行新增，确认每项局部静态范围、F15沿各准确版本继承退回、原生/实际来源及全部动态未运行，c35另审。四次实际读取回复确认本轮唤起，没有授予新窗口。

## 准确c35ec0d9的F15检查返修复核

结论：**接受准确`c35ec0d9003d0052f0a1ee9e5da5d1b8f9efb279`对F15的静态返修**。按原Worker`msg_c3b0bcb67a99`及主控`msg_87b05dc293bb`完整读取相对b3的全部三路径，只有测试及两份说明；产品规范化、真实ContextFile.Version及原handler均未改。

成功重放控制现放在contextPath实际改写之前；改写并恢复字节后，真实stat核size及原字节相同、mtimeNs确已变化，再要求原R得到request_mismatch，实际前后mtime记录进入结果。没有观察到mtime变化时该断言失败，不能计版本变化通过。首次原生固定context和唯一Run/root仍检查，不回填旧mtime、不删除version或生成替代任务，原F15的预期矛盾已消除。运行结果仍NOT_RUN，原有局部静态接受和整体动态证据不足保持。

本段`git diff --check b3f0e22919fbbfab155af0fa9990602f50922117 c35ec0d9003d0052f0a1ee9e5da5d1b8f9efb279`在BelowNormal下EXIT0。说明仅更新接受归属/事实及原件版本检查，没有新增或修改链接，沿用两文档成熟Markdown结果，没有运行测试或资源采样；推送及远端同SHA采用Worker成功回执，主控本轮未另核远端保持。§7.5仍由原Worker静态实施；T11新增真实人类入口待用户范围回答，T12后继契约及测试计划待准确交审，均未被本次接受登记通过或冻结接口。

## 恢复会话与准确8e066控制脚本身份复核

本任务经原生retry-of恢复为`task_1cd8014750f9 / ctx_6339d8beeb73`，当前terminal为`term_9e9193d2-733a-4ab1-8220-22a5dc640c0f`，实际客户端Orca / Codex、会话模型`gpt-6.1-sol`、effort xhigh，来源为主控当前任务记录。完整读取根忽略目录`swa-review-recovery-20261020.md`及main当前事实源；原M0与既有准确审查保持。已通过`msg_6ae48a37e0fd`报告实际读取，主控`msg_e24b6fe726d9`回复确认恢复唤起；c29/f08/c0及累计管理记录分别接续审核。

先独立确定本次控制脚本接受条件：准确全文件只有两项真实身份常量改变；恢复launch、restored及当前running记录同源同身份；当前进程程序、创建时间及唯一监听相符；原库保全原件保留；既有准备、精确停止和恢复条件保持，下次窗口仍须新鲜身份核查与资源准入。

结论：**接受准确SHA256 `8e06639aa6c0bf635b037fb82a8b745540ed0dd3a754b15e8220d098652b606d`的两项身份常量更新，可用于主控安排下次窗口的准备；本次不授予执行窗口。** 已完整读取根忽略目录`swa-cli-v1-control/service-window.py`。仅在内存把PID43068及创建UTC13:41:49.053571恢复为旧PID12800及UTC09:43:52.279923后，全文件SHA256精确等于已接受`0742a46eb64d399ed8171b5a9f426a0a5bfb53e1561442428d7484b46df22424`，其他流程字节没有变化。

UTC2026-10-20T14:08:18的独立只读查询确认：restore-03/launch.json、restored.json及当前running.json均为原S `ed543c12106c2dbb3e181cb9bc1647b626455179`、PID43068、创建UTC2026-10-20T13:41:49.053571Z；真实程序为原`versions/ed543c12/astrocyte.exe`，父PID51152，唯一监听127.0.0.1:64505属于该进程。restore-03/database-check.json记录integrity ok、foreign_keys空及stopped_interval_comparison equal；这是已执行恢复的原件核对，本Reviewer没有重新访问SQLite或HTTP。

`pwsh -NoProfile -File .verification/verify-window-8e066.ps1` EXIT0，BelowNormal，原件为`.verification/window-8e066-readonly.json`。没有执行控制脚本、资源采样、停止、启动、编译、安装、信号或模型。下次prepare仍须重新核对真实进程与环境、使用新保全目录及真实资源检查；当前唯一窗口继续属于原S。首版程序、T00—T15、C01—C09及安装后同候选双Worker动态范围继续未验，T11待答及T03/T12 strict helper准入缺口保持。

准确8e066接受hook为`msg_da4f4cbe6ff9`；主控`msg_3abd95e5ac8f`完整读取准确报告ebd相对27f的全部新增，确认内存字节对比、当前真实身份及数据库原件、全部未执行限制，实际读取回复确认唤起。后继主控安排prepare-04/stop-04和仅构建窗口，不改变本次接受范围。

## c29、f08与c0的独立条件及准确静态CR

本会话依据main的首版目标、SPEC §21.1、既有H05/H06/H08/H10/H11/H15/H17及已审b3 §7.5核定条件，再沿准确Git对象逐段追踪：普通CLI只发布一次完整JSON，原R和实际选择器与HTTP trace分开，错误原件和原生结果保持；网络、取消、输出失败不重发；serve启动前后EOF和失败分别结算，ready后的诊断不追加结果；直接子进程完成可重复观察，全体所属后代明确退出且原生资源空闲才允许正常0；准备提交须与已声明依赖、补丁、实际生成来源相符，准备成功不扩大为功能通过。后继工作树WIP没有用于这些结论。

### c29：普通CLI输出首段静态接受

结论：**接受准确`c29e5af7dd8fa1c578ba7499564365f692dd8a9d`相对c35五路径的局部静态范围。** 已完整读取三份Go及两份说明差异，并追踪实际ValidateOrcaCommand、OrcaEnvelope与ErrorV1结构。合法原R仅从实际已解析的canonical UUID取得，选择器在App校验后保存；未知flag导致解析提前终止时，不从尚未处理的后继参数补取身份。三个传输前函数现检查一次对象及EOF，分别保留合法R先解析、尚未解析和非法R的实际次序，未生成runtime/authority/Run。

单次发布在实际写入前设置attempted，短写与部分写入失败不再输出第二个结果，stderr保留已知及请求的关联；有副作用的命令返回未知6，其余输出失败7。成熟Decoder的UseNumber保留结果数字、DisallowUnknownFields及恰好一个对象检查明确native和HTTP ErrorV1来源；原生结果/error.data及实际关联保持，HTTP ErrorV1原件进入既有error.data，HTTP header trace只写独立诊断。真实连接未建立映射4，其他网络/响应不完整保持6，实际context取消为130；没有重发分支。失败任务的成功查询仍按真实envelope退出0。

这段没有实施serve EOF、全后代收尾或C01—C09驱动。`check-revision.ps1 -Previous c35ec0d9003d0052f0a1ee9e5da5d1b8f9efb279 -Candidate c29e5af7dd8fa1c578ba7499564365f692dd8a9d` EXIT0，Git空白、两Markdown及5本地链接通过。没有运行Go函数、HTTP或取消/管道正文；实际功能继续证据不足。c29首次TLS push失败与随后Worker普通push/ls-remote成功均按主控准确任务记录保留，本Reviewer没有重新推送Worker分支。

### f08：锁与真实类型生成准备范围接受

结论：**接受准确`f08b52111d4860b5ef162987cce2f43f4a4e6ce0`相对c29的两路径准备范围。** 完整读取锁与类型差异、准确workspace/runtime包声明及generate-client.mjs。成熟YAML2.9.1/JSON读取准确Git对象，核43项runtime依赖与4项开发依赖specifier均对应锁，injectWorkspacePackages与现有workspace声明一致；node-gyp12.3.0及其新增闭包来自既有包声明。准确node-pty补丁原字节SHA256为`e8bc0de94dc02036833f92059fc6e000fea00e113811193382164deab7c7ea60`，锁的patch值、importer及snapshot相同，其他补丁没有改动。

类型只增加既有contracts的Orca命令路径、18项动作及对应command/nativeError/envelope/operation，没有修改原公开字段或生成器。已读取Worker dependency-02的lock/install原日志、四项实际退出与两组资源原件，以及generate-01日志、两项退出与新双样本：均EXIT0；真实镜像超时的命令内建重试、peer和弃用警告保留。next-window脚本明确BelowNormal、私有TEMP/cache/store、安装Node1024MiB及生成512MiB，各项重新采样且串行，实际日志与准入门限相符。生成日志为实际成熟openapi-typescript链写入结果，本Reviewer没有重复安装或生成。

`node --max-old-space-size=512 .verification/review-f08-source.mjs` EXIT0，原件`.verification/f08-source-review.json`；两路径`git diff --check c29e5af7dd8fa1c578ba7499564365f692dd8a9d f08b52111d4860b5ef162987cce2f43f4a4e6ce0` EXIT0，BelowNormal。仅支持锁同步及实际类型生成准备接受，native编译、类型检查、完整包、安装或功能均不能由这次结果判定通过。

### c0：所属EOF、Job与原生空闲关闭局部静态接受

结论：**接受准确`c0ab15e50d230a93a0aca17ea12e1f5c132ccb5c`相对f08八路径的局部静态范围，完整§7.5和首版功能继续证据不足。** 已完整读取全部差异，额外追踪准确原生orcad-lifecycle、OrcadRuntimeLifetime、RPC stop、Worker资源查询与daemon代际清单。serve消费自己的真实stdin，通过带原因的context区分EOF、读取失败及取消；readiness前取消输出单一失败，readiness后只保留原结果并用stderr及实际退出表示关闭。子进程的done关闭后可重复观察，stdout先读至EOF再Wait，不再重复消费单一Wait值。

serve在创建子进程前加入新独立Job；初始核自身成员及active=1，直接子进程经真实句柄核成员/创建时间，关闭后两次相隔100毫秒的新accounting必须只剩serve自身。没有kill-on-close、breakaway、原Job配置或广域终止；成员/查询/句柄失败、残留和清理超时保留6，子进程失败保留相应非零。实际嵌套Job、ConPTY及全部退出观察仍须动态验证，静态代码不能证明本机已经支持该路径。

原生stop先停止RPC，再检查in-flight、全部pending receipts、无分页上限的实际Worker资源、每个controller准确close及所有daemon代际权威PTY清单。原有cleanup按逆序运行，失败会聚合抛出且不释放相应fence；只有全部成功才写一次私有关闭确认，Go核runtime/data与readiness一致并等到EOF。缺确认或未知后代不能取得正常0；没有把直接root退出当全树完成。内部生命周期记录没有扩展公开命令/字段或执行状态账本。

App任务规格使用成熟json.Valid先核完整单JSON，再用原DisallowUnknownFields Decoder和原业务字段校验；连续JSON或坏结构不能通过，移除io没有降低该校验。已读static-eof-01实际架构失败及static-eof-02的45包通过原件，检查规则没有改动。`check-revision.ps1 -Previous f08b52111d4860b5ef162987cce2f43f4a4e6ce0 -Candidate c0ab15e50d230a93a0aca17ea12e1f5c132ccb5c` EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。

C01—C09安装驱动/console helper尚未交付；T03/T12已审strict helper caller准入尚未实施，当前host的方法白名单不能证明真实caller，原生verify=null仍允许declared-handle的边界保持待修。强核验必须位于native mutation executor前并覆盖已完成R重放；恢复证明须单独核实，缺attestation拒绝或保留unknown，不能沿当前launch豁免。T12完整契约/计划须由主控定稿并独立审核，T11真实人类入口范围待答；本轮不冻结新参数，不计这些范围通过。没有正式PR，本轮未发布GitHub review；仅通过原Run交付。当前仅原Worker获主控的一次build窗口，本Reviewer没有编译、产品测试、安装、服务、SQLite、信号或模型调用；后继实际构建和动态结果另审。

三个准确hook为`msg_e1e98224f02a/msg_9da61336bef0/msg_d41f49127830`，主控在`msg_c227f4c799b8/msg_8e0f825991a9/msg_ec1406eec1e5`逐项完整读取报告6d693df7并回复，分别确认局部接受、实际来源及全部未验限制；本轮实际读取唤起已确认。

## build-01实际失败、bc217返修及准确6d7控制脚本

已完整读取准确c0的构建入口和build-01实际log/exit/processes-after原件：资源检查EXIT0，构建EXIT1，Node48860在第33行mkdir候选目录时因dist父目录不存在发生ENOENT；原MODULE_TYPELESS警告保留。该退出发生在部署、Node下载及native/TS/Go编译之前，没有候选VERSION。主控已核实际退出和所属进程后收回窗口，restore-04真实结果及恢复身份见下文；本Reviewer没有重新执行构建。

先核定返修条件：入口自身建立获准仓库中的dist父目录，现存重定向或越界仍拒绝；候选目录继续独占创建；不手工补目录后绕过原入口，不改变上游或把失败计作编译通过。结论：**接受准确本地提交`bc217de88f4bb4b8af1ab63f3c281f4c2f649aea`相对c0三路径的局部静态返修。** 完整读取七行脚本及两份说明；mkdir仅为dist父目录递归创建，lstat/realpath要求它是实际仓库的dist，候选的实际parent必须仍位于其中，候选自身继续非递归mkdir，已有路径会失败。未改变后续部署/编译行为；修复效果及完整构建仍未验。本次只取得主控本地准确干净提交交付，尚未取得该提交普通push成功回执，不补写远端成功。

`check-revision.ps1 -Previous c0ab15e50d230a93a0aca17ea12e1f5c132ccb5c -Candidate bc217de88f4bb4b8af1ab63f3c281f4c2f649aea` EXIT0，两Markdown/5本地链接与Git空白通过，marked18.1.0、BelowNormal、Node512MiB。没有运行脚本或创建业务目录；原Worker负责后继实际验证。

结论：**接受控制脚本准确SHA256 `6d7bed3576ade7b9522738ba741a55eade309f9c8b9de93a869b3150562d6870`的两项恢复身份常量，可供主控下次prepare使用，仍须新鲜身份/环境/保全和资源准入。** 按已审8e066的相同独立条件，仅在内存逆置PID45948和UTC14:14:47.360557为原43068/13:41:49.053571，得到全文件准确8e066，其他流程字节未变。已核restore-04 launch/restored/current running及原S、原库保全JSON，实际integrity ok、foreign_keys空、停止区间equal；恢复双样本CPU4.8%/9.8%、可用10.709/10.650GiB、BelowNormal/3.001706秒，符合既有门限。

UTC14:20:26的独立只读查询确认原程序、PID45948、父PID55664和唯一127.0.0.1:64505监听；Get-Process原生创建时间为UTC14:14:47.3605568，与原件created_seconds换算精确相同。原件datetime四舍五入为360557，CIM截断为360556；只读检查先前按字符串比较及DateTime隐式转字符串丢失小数两次EXIT1，已按实际epoch数值核对，沿被审owner原有小于1毫秒判据，当前差为0，没有修改控制脚本或原件。`pwsh -NoProfile -File .verification/verify-window-6d7.ps1`最终EXIT0，证据`.verification/window-6d7-readonly.json`保留真实精度与来源。没有执行控制脚本、SQLite、HTTP、资源采样或重操作；当前窗口继续归原S，后继尝试由主控安排。

## T12补充计划读取前的独立条件

以下条件依据已审§3.2/3.3及T03/T12、SPEC §21.1和H03/H06/H10/H11/H17形成；尚未读取后继T12草案，不冻结新参数或公开接口。

- 接续明确区分原helper仍可证、原helper已真实退出、原helper归属或退出未知。只有前两类能分别进入已核恢复或新helper路径；未知保持占用且不创建竞争控制端。runtime/host变化不单独证明旧进程退出。
- 读写前核当前project主体、适用Grant及动作权限，核原Run/root Task、controller、固定context、原authority和实际持久关联。另一真实项目/Run、撤销/过期批准、旧context与缺关联各自拒绝；可信人类自有安全停止仍沿原适用例外。
- 产品strict admission必须在原生mutation executor及完成R重放之前执行。无hook/verify=null、declared handle、外传PID/actor/correlation或只在handler内核验均不能证明当前helper；当前launch和重启恢复证明分别核实。
- 新helper须有自身真实PTY/进程创建身份和原生准入，再由真实runUse取得原Run并推进consumer generation。旧helper/旧generation即使持原R、已完成receipt或旧环境也不能继续写入；原Run/root/context不复制为新任务。
- 原请求、checkpoint和固定子请求仍是效果来源；接续或超时不重发未知输入。原receipt/SDK/Run原件保留，跨generation对重复R、不同正文及pending的处理须明确，不能换fingerprint或request ID重复副作用，不能另建恢复账本。
- surviving Worker、daemon、session及工具关闭范围分别核原身份；缺原生进程句柄或恢复证明保持unknown及占用。新helper不能仅凭可查询就重建原Worker或宣布原SDK连续存活，旧会话的状态和部分结果仍可按授权读取。
- CLI/HTTP/App/helper端到端接口、错误/退出、幂等及状态消费者须完整对应；只读原请求查询在不调用helper、不创建或发送的路径上继续可用，输出前核project/controller所属，缺证与拒绝不泄漏其他项目原件。
- 测试必须使用同一真实候选/私有runtime、原生SQLite/PTY及自产session，实际命中重启、恢复/新helper、generation变化和旧代拒绝。权限、缺恢复证明、未知副作用及重复请求均按原要求保留；无真实材料或未命中分支继续未验。当前不授权模型、重启或服务；T11人工入口待答不由T12代答。

## F16：原生runShow参数错配阻断完整成果

结论：**退回准确`c0ab15e50d230a93a0aca17ea12e1f5c132ccb5c`所含task.result完整成果路径，严重程度高。** 此为等待T12准确计划确认期间，独立追踪既有原生契约发现的继承缺陷；c29/f08/c0先前明确限定的新增局部静态接受保持，整体功能从未登记通过。

准确`runtime/swa-orca/src/controller-helper.ts:74`的collectTaskResult调用`orchestration.runShow`传入`{ run }`。同一准确第三方`src/shared/rpc-contract/orchestration-runs-params.ts:23`将RunShowParams定义为必填`id`及可选`from`；`src/main/runtime/rpc/methods/orchestration/runs/runs.ts:144`处理器直接读取params.id，RuntimeClient.call将params原样送入transport，没有run到id的转换。实际task.result在前序task/fleet读取后必然缺id，原生schema拒绝使helper返回错误，无法进入native-methods中的completeTaskResult或CLI成果导出。bc217与897531只改脚本/文档，继承这条调用。

已分别通过`msg_298441a32b13`及`msg_31b270bc84ec`交原Worker和主控：按真实契约修正现有helper消费者，保持上游schema、真实所属与完整成果要求；后继准确提交独立复核，T07/T13的真实task.result及完整成果验证继续必需。本Reviewer没有运行原生schema、HTTP、模型或产品正文，没有使用mock或伪造原生调用；静态因果来自准确双方契约，实际运行结果仍未验。

主控`msg_c817596bd488`完整读取准确abc8fc98相对97a八行新增并实际核双方契约，回复F16 hook `msg_d0ebbd3cc707`，确认退回范围及原Worker后继返修，实际读取唤起已确认；当时Worker仅执行已授build-02，源码冻结，未运行成果获取。

## 准确897531d5的T12提案审核

结论：**退回准确`897531d5d2978e3a51a043063a26e254b3d0f766`的完整T12提案，补齐P01—P03后复审。** 本会话八条独立条件已在97a写入并普通push，随后通过原Orca ask取得主控对准确版本、唯一recover方向及尚未定稿/不授权实施的确认，再完整读取相对bc217两文档的§3.4/7.6及R01—R09。主控`msg_fe106620a3f2`明确新增动作需覆盖CLI副作用分类及真实C04/C05/C09、helperless bridge不能放行任何原生写入。bc217/897的普通push成功采用Worker后继实际回执，主控`msg_b78b1ff1e625`已确认收到；先前未取得回执的事实范围保持。

已覆盖的需求：R01/R02核真实caller与当前权限，R03/R04核活旧helper拒绝及真实同Run接续，R05/R08核竞争、重复R及checkpoint后只观察，R06/R07核同库重启和原Job复开，R09核历史只读、完整原件及最终关闭。真实runUse改变原consumer_generation，原Run/root/context不复制；缺恢复attestation、SDK父Job丢失、未知输入和资源不被写成连续运行或已释放。没有发现应删除的无关检查，全部实际场景仍NOT_RUN。

已独立读取准确原生run-binding、strict launch验证、legacy路由、RuntimeClient及mutation executor，确认currentRuntimeLaunchSufficient只适用于真实当前launch commitment，restored仍需要原生恢复证明；完成R在executor中可能直接返回，前置准入位置必须保留。命名Job复开方向有[Job Objects](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects)、[CreateJobObjectW](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-createjobobjectw)及[OpenJobObjectW](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-openjobobjectw)官方正文依据：旧对象有存活关联进程时可继续存在，Open只取得既有对象句柄，名字及成功打开不能证明实际成员或合法接管。该API依据不表示本机嵌套/重开已通过，提案的kernel成员、实际daemon来源与握手共同核验仍必需。

| 缺口 | 来源与具体影响 | 原Worker需补充的方案及真实覆盖 |
| --- | --- | --- |
| P01 recover的CLI失败边界 | H05/H06/H11、SPEC §21.1及主控明确要求。现有outputFailure按硬编码动作分类；新recover创建helper和改变consumer，不能沿默认只读输出失败7，也不能只用普通新动作正常退出检查代替 | 明确在CLI所有副作用/取消/输出失败处归入recover；复用C04/C05/C09的真实连接、受控130及stdout关闭条件，核原R、实际已知旧/新selector、6/130、一次结果和无第二副作用，未命中保留INCONCLUSIVE |
| P02 严格准入的合法bootstrap | H03/H06和既有controller open/task create功能。提案要求全部helper写入/带caller查询同时比较原receipt和当前Run consumer，但pending open的首次runCurrent/首runCreate尚无Run，首次taskCreate及新helper ready/runUse亦处于不同阶段；尚无按方法/阶段的完整判据 | 列出这些真实bootstrap、后续current consumer及已完成R重放的准入矩阵，明确每项真实launch及准确parent/子请求证据；恢复缺证继续拒绝。helperless只读白名单不能推导任何orchestration写入豁免。R01必须经真实open、taskCreate和recover分别覆盖，不以修改上游宽松分支代替 |
| P03 无副作用拒绝与unknown结算 | H06/H10/H11。提案先在事务内保留lineage再异步观察旧helper，同R pending禁止create/send/runUse、不同R被屏障阻止；尚未明确已证旧helper仍活等未产生外部副作用的拒绝，如何结算reservation并允许以后显式合法接续。将其全部保留为unknown会使正常R03拒绝影响后继R04 | 定稿内部保留范围与首次可能外部副作用的分界。仅明确证明无外部效果的拒绝按原生结算；可能已创建/发送/绑定、证明不足及中断继续保留原R和占用。补同一真实lineage的连续场景：R03拒绝后，旧helper随后真实退出，再以明确请求完成合法R04；同时保留R05竞争与R08未知不能重发的反向条件 |

P03为方案未规定的状态分支，不宣称现有源码已经产生永久阻塞。原生executor会按错误类型等条件结算pending；补充方案须明确如何使用其真实规则，不能在实现时猜测。P02同样不预先指定新的公开字段或额外身份系统，要求原已审能力的真实启动过程能够通过严格验证。

已通过`msg_6dc0bb0e151b`交原Worker三项补充；接口仍为提案，由主控定稿后才实施，T11待答不阻断本项，F16另行准确返修。`check-revision.ps1 -Previous bc217de88f4bb4b8af1ab63f3c281f4c2f649aea -Candidate 897531d5d2978e3a51a043063a26e254b3d0f766` EXIT0，两Markdown/5本地链接和Git空白通过，marked18.1.0、BelowNormal、Node512MiB；没有编译、原生调用、重启、服务、Job创建/复开、模型或计划正文执行。

准确897531退回hook为`msg_7f667155d33a`；主控`msg_43bc1c4befbe`完整读取准确bad4d7c5相对abc8二十行新增并回复，确认三项方案缺口及全部未运行限制，接口继续提案。原Worker在build-02/Git冻结期间只准备忽略目录草案，窗口归还后正式提交复审；原hook实际读取唤起已确认。

## F17：worker.send消息类型不符合真实原生枚举

结论：**退回准确`c0ab15e50d230a93a0aca17ea12e1f5c132ccb5c`所含worker.send路径，严重程度高。** 对F16同一helper的直接消费者继续核查，准确controller-helper.ts的worker.send分支把`type: 'message'`送入`orchestration.send`；同准确第三方`src/main/runtime/rpc/methods/orchestration/schemas.ts:96`的SendParams.type使用`z.enum(MESSAGE_TYPES)`，其`src/main/runtime/orchestration/types.ts:3`定义的真实枚举没有message。RuntimeClient不转换这个字段，合法项目/Worker/输入也会在原生schema处被拒绝，T13要求的同session后续输入不能由当前实现完成。bc217/897531继承；没有运行产品或模型复现。

已通过`msg_1dacccd77c30`及`msg_518dd0cfda7f`交原Worker与主控，要求在当前build-02结束解除冻结后，与F16按原helper写域依据真实协议返修；不修改在途准确897531、不扩充上游枚举、不伪造消息或以新Worker代替后续一轮。后继准确提交独立复核，原真实send、B的同session后续输入及无重复效果验证继续必需。既有局部静态接受保持，完整输入/成果与首版动态验收从未登记通过。

F17正式hook `msg_b216a6dae3f1`已获主控`msg_8e18be67c949`完整读取回复，主控核实际枚举并要求原Worker核全部既有RPC消费者及真实成功语义，实际读取唤起已确认。原Worker后继`msg_778f4556e8f4`报告合法原生消息仍只投递mailbox pointer，ToolsOff Pi无法自行读取正文，正另行形成正文投递方案；本Reviewer尚未读取该新方案，其接受范围仍未知，枚举修复不能单独证明后续输入可用。

## build-02实际失败及准确03c64恢复身份

已读build-02最终日志、build-exit=1和processes-after-01/02两次实际空进程表原件：PNPM部署已返回，随后固定Node首归档下载命中原10分钟AbortSignal TimeoutError，尚无headers/native/TS/Go/VERSION。部分candidate01及失败/警告保留，没有重复执行本阶段；本Reviewer没有下载或执行构建。主控实际确认原Node18536/9272及所属执行退出后收回窗口并解除源码冻结，同一Worker负责后续参数/语义和计划返修。

结论：**接受控制脚本准确SHA256 `03c64ef6922af3e4ce00763159377c26a03d8d065eb36e0b32e20b65ba577ceb`的两项恢复身份常量，其他流程沿已审6d7不变。** 只在内存把PID18944/UTC14:49:56.491641恢复到旧45948/14:14:47.360557，全文件精确得到已审`6d7bed3576ade7b9522738ba741a55eade309f9c8b9de93a869b3150562d6870`。沿既有两字段/真实恢复/当前所属独立条件核查，无额外流程修订。

UTC14:52:33独立只读确认restore-05 launch/restored/current running、原程序、PID18944/父50264、Get-Process创建UTC14:49:56.4916412和唯一127.0.0.1:64505监听；实际创建epoch与原件1792507796.4916413数值对应，沿既有owner小于1毫秒判据差为0。database-check原件为integrity ok、foreign_keys空、stopped_interval_comparison equal；恢复启动资源CPU15.3%/11.4%、可用10.364/10.406GiB、BelowNormal、单调3.002259秒，符合门限。当前原S已恢复并占用唯一窗口，主控原配置/health/Web及原Orca状态观察保留其实际来源。

`.verification/verify-restored-owner.ps1`以本次准确hash、前后PID/UTC和restore-05参数运行EXIT0，原件`.verification/window-03c64-readonly.json`；标准JSON明确保持日期字符串，并以原件epoch核真实进程创建时间，没有隐式日期格式比较。该检查只读文件、进程和监听，未执行控制脚本、SQLite、HTTP或资源采样。准确脚本可供主控下次prepare使用，仍须重新核实际身份/环境、新保全目录和资源准入；本次不授予新重操作窗口或功能接受。

## F17正文语义的来源核查与计划读取前条件

已独立读取准确c0的structured-mailbox-pointer-delivery、formatter、structured-mail-source及Pi rpc-prompt实际调用链：原生队列把formatMessagePointer的check提示写入AgentJournalMessageItem.blocks；source只携带mailbox/Dispatch/message ID和sender引用，未携带正文；preparePiRpcPrompt只把text blocks连接为SDK的prompt.message，没有按source回查MessageRow。关闭工具的Pi不能自行执行check取得正文。因此把非法message换成合法枚举只能消除参数拒绝，现有正文语义仍不足以证明worker.send完成；该继承限制并入F17，不以重复问题另计结果。

在读取后继正文投递方案前，依据用户控制目标、H03/H05/H06/H08/H09/H11/H12/H14及真实T13形成以下条件：

- 实际发送必须从准确原生MessageRow取得已授权正文及来源，进入同一目标Worker的真实SDK输入；仅通知check、ID来源或回执不能代表正文已到。调用者正文保持文本，不成为shell、原始SDK RPC或新的权限参数。
- 只在明确属于本产品及当前Run/root/Dispatch/session的Pi路径接入；其他原生宿主和本机Orca保持原行为。固定provider/model、ToolsOff、六项RPC和资源关闭条件继续覆盖后续输入，不能为读取mail开放命令工具。
- 保持当前项目主体/Grant、controller consumer及原生session/runtime fence；发送、排队和实际消费各按其原生边界核验。停止、释放、user_owned、旧代或目标变化不能被正文投递绕过，不向替代Worker发送来完成原请求。
- 继续使用同一原生消息、delivery/operation及公开R的持久来源，说明正文快照、排序和批次对应关系；重复R/重连不得重建消息、改读后来正文或产生第二prompt，不建设另一队列/请求账本。
- 区分写前明确拒绝、原生队列已接受、实际SDK输入及副作用未知。断开、超时、busy/held、重启与输入后存储失败保留真实状态；未知继续查原R及原operation，不重发或用新R/新模型隐藏原效果。
- 按声明的编码、大小和批次边界真实处理换行、引号、Unicode及类似JSON的文本；超限或坏源在实际边界失败，不截断为成功或拼接其他项目邮件。完整来源及未知用量按当前原件返回。
- 测试须在最终同一安装候选的T13中证明B接收实际后续正文、同一PID/session/Dispatch及正确原生操作/turn来源，按固定业务判据产出真实后续结果，A及其他Worker保全。核自产SDK输入与准确MessageRow正文对应，并覆盖真实重复请求/断开和必要拒绝；schema接受、另一会话、仅消息ID或模型自述均不满足这些条件。

这些是现有首版输入能力的必要条件，尚未选择或接受新接口。后继完整方案及测试计划由原Worker准确提交、主控定稿并独立审核后实施；当前无新的运行窗口，本Reviewer没有运行任何消息、SDK、模型或业务正文。

## 准确079bd958的T12补充计划复审

结论：**接受准确`079bd958260ac3edb9fda6c13bb9e28a0a08cd56`的T12完整补充计划，P01—P03已覆盖。** 按原Worker`msg_fadace455b7c`完整读取相对897531的两文档差异，重读§3.4与§7.6及实际原生executor规则，再对照97a的八条独立条件和bad4的三项缺口。此提交只改文档；F16/enum参数WIP及正文语义新方案没有混入。实际普通push回执尚待补报，本次接受基于准确本地Git对象，不声明远端成功。

P01现明确将recover纳入App/adapter/CLI outputFailure、client/native enum、durable方法及HTTP/CLI契约的写动作消费者；输入可能进入handler后网络或stdout失败为6，真实取消为130，保留原R/C_old及实际取得的C_new/Run/root，一次输出和原请求查询保持。R10复用真实C04/C05/C09条件并核原receipt/PTY/runUse与无第二副作用，没有伪Writer、假响应或延迟钩子，未命中仍INCONCLUSIVE。

P02的默认拒绝矩阵分别规定pending open真实runCurrent/ready可为current=null、open完成后的只读、准确pending taskCreate的首次runCreate/root子请求、pending recover的ready/runUse、当前consumer写入、restored身份及完成Native R重放。合法未绑定阶段不预设已是consumer，完整launch/PTY/pane/host/incarnation与原parent/子R仍必需；worker.stop/release从实际evidence核caller，无参数from不构成豁免。helperless bridge仅白名单只读，任何原生写入/legacy takeover或隐式Run绑定拒绝；公开历史outer receipt不声明活owner。R01相应覆盖各真实启动阶段及无launch普通进程，缺restored attestation继续拒绝。

P03已区分scope-reserved内部持久保留与pty-spawn前首次可能外部效果：前者不调用PTY/send/runUse，也不mark effect；两次成功观察确认旧helper仍活且原生effect-free结算成功后才解除reservation。读取/结算未知保持原R/占用。pty-spawn之前先checkpoint expected身份再mark effect，后继异常必须使用operation_unknown保留pending；准确原生executor对非prompt普通异常可能discard的实际行为已明确，单独mark effect不承担保全。pending同R只观察、不再次create/send/runUse；R03明确无效果且结算成功后，才允许旧helper真实退出及调用者显式新R的连续R04，R05同时保留竞争和未知屏障反向条件。

原Run/root/context、真实generation、原Job复开与daemon来源/握手、历史只读和最终全部资源退出的既有条件保持，原R01—R09加R10没有发现剩余覆盖缺口或多余检查。此处接受方案及计划，源码仍须由同一Worker实施并准确CR，主控先将接口写入SPEC后再授权；T11待答、F16完整成果、F17真实正文语义及全部动态结果各自保持。没有执行controller、Job、数据库、服务、信号、SDK或模型。

`check-revision.ps1 -Previous 897531d5d2978e3a51a043063a26e254b3d0f766 -Candidate 079bd958260ac3edb9fda6c13bb9e28a0a08cd56` EXIT0，两Markdown/5本地链接、无锚点/HTML或Git空白错误，marked18.1.0、BelowNormal、Node512MiB。普通源码格式/编译、安装和R01—R10正文均未执行，当前没有新重操作窗口。

## 准确ebf55ac0的F16及F17参数返修

结论：**接受准确`ebf55ac024a618d1ad32a741eb4eefcfc5dae77b`相对079bd三路径的局部静态返修。** 完整读取helper两行及两份说明差异，F16调用现传`runShow({id: run})`，对应已独立读取的RunShowParams必填id及实际handler；F17改为真实枚举包含的status，原生schema与RuntimeClient没有修改。原R及固定子R、实际Dispatch收件人、原项目与上下文没有改变。

Worker逐调用核查表保留实际来源；本Reviewer另核原生WorkerStartParams的spec/task互斥、from/model/effort/timeout字段，WorkerListParams的显式run/paginate/limit、真实分页上限100，workerRead的transcript及workerShow投影assigneeHandle，与现有直接消费者一致。已有hasMore仍使不完整成果清单失败，没有通过扩大上游字段、枚举或忽略分页解决问题。

F16参数错配及F17非法枚举的静态原因已消除；真实task.result、完整成果及T07/T13正文仍未验。合法status当前仍只投递check提示，F17完整正文能力继续退回，七项独立条件及其新完整方案另审；本次没有把参数接受扩大为权限、功能或模型成功。原Worker`msg_e63aea03cb87`后继报告079bd普通push EXIT0，`msg_f11c208bde90`报告ebf普通push EXIT0，按这两项真实回执记录，本Reviewer未另查其远端。

`check-revision.ps1 -Previous 079bd958260ac3edb9fda6c13bb9e28a0a08cd56 -Candidate ebf55ac024a618d1ad32a741eb4eefcfc5dae77b` EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB；没有运行原生schema、SDK、HTTP、消息、模型、编译或产品测试。

## 独立Go源码检查调整的计划读取前条件

按主控`msg_d7173cc26dcb`的既有四包检查顺序要求，先完整读取准确079bd的compile-go-tests.ps1，并依据H05/H13/H15、已审39cc四包计划及原T03/T13形成条件，尚未读取后继c2d4b62d计划正文：

- 编译前取得真实准确干净HEAD，只在本项忽略目录的全新私有绝对输出目录编译；移除Candidate前置时，不创建假安装目录、伪VERSION或替代候选。SOURCE仅记录实际编译源码。
- 四个真实Go package及已有目标live函数不变，逐包串行test -c；此阶段只运行已审三项真实传输前CLI拒绝函数，逐项实际PASS，零匹配、SKIP或失败不能计通过。不在编译窗口启动HTTP/原生服务或模型。
- 按RESOURCES每次真实资源准入、BelowNormal、Go4/-p4/512MiB及私有TEMP/cache/modcache执行，原S窗口由主控另行交付；无窗口保持未执行，不更改全局配置或安装后的用户目录。
- 编译及允许的检查全部成功后，再核同一干净源码未变化，才写既有SOURCE/tests.json；任一失败保留真实日志/退出与部分输出，不发布完整成功，也不伪造未执行包的结果。
- 后继安装服务T03仍强制候选VERSION、编译SOURCE及当前准确源码一致，并运行四个真实live函数；零匹配/SKIP和仅编译不代替真实结果。源码变化后旧二进制不得冒充新候选测试。
- 独立源码检查只帮助发现真实Go错误，T12/正文方案、原生/SDK/权限/隔离、最终同一安装候选双Worker及全部适用失败路径继续各自验收，原失败和未知不被执行顺序调整覆盖。

这是实施前覆盖条件，未修改或运行编译入口，也没有新增动态授权。

## 准确c2d4b62d的独立Go源码检查计划

结论：**接受准确`c2d4b62d64eef73eb3c83064a8354696b6e623c7`的既有四包Go检查顺序调整计划。** 六条独立条件先在da2e7268形成并普通push，随后完整读取相对ebf的两文档差异及§7.4，并核原编译入口和现有真实服务T03的SOURCE/VERSION消费者；本次没有改动脚本或产品。

新入口移除Candidate参数及VERSION前置，只读取实际40位干净HEAD并使用本项全新私有绝对输出，不伪造候选。四个原Go package逐包test -c及三个真实传输前函数保持，逐函数PASS、零匹配/SKIP拒绝；各重操作新资源准入、BelowNormal、Go4/-p4/512MiB及私有temp/cache/modcache、串行和失败即退出明确。只有全部实际成功且结束时仍为同一干净源码，才写既有SOURCE/tests.json，不把部分编译或未执行函数登记为通过。

最终真实安装T03仍强制Candidate VERSION=编译SOURCE=准确源码，运行四个原live函数并拒绝零匹配/SKIP；源码变化时旧二进制不能用于新候选验收。当前调整只允许先发现Go源码错误，没有放宽HTTP、当前批准、原生SQLite/控制端、ToolsOff、模型、隔离或同候选T13双Worker条件，T03其他未覆盖范围继续保留。未发现六条件的剩余缺口或新增多余检查。

`check-revision.ps1 -Previous ebf55ac024a618d1ad32a741eb4eefcfc5dae77b -Candidate c2d4b62d64eef73eb3c83064a8354696b6e623c7` EXIT0，两Markdown/5本地链接、无锚点/HTML或Git空白错误，marked18.1.0、BelowNormal、Node512MiB。本轮准确c2只取得本地已提交交付，普通push回执仍待补；计划接受后由原Worker修改入口并交准确CR，编译/函数/实际服务仍等待主控独占窗口，本Reviewer没有执行任何编译、业务测试或模型。

准确079bd的接受hook `msg_bb2f7ef94f96`已获主控`msg_314b61388dd4`完整读取7c6d报告回复，准确ebf的接受hook `msg_4638289c5248`已获主控`msg_9d609304dd77`完整读取da2报告回复，实际唤起均确认；T12先写SPEC再实施、F16/F17仅参数静态接受的范围保持。准确c2计划接受hook `msg_4cf6aa73441e`已获主控`msg_6c1a7618ade4`完整读取483及准确计划回复，实际唤起确认，仅可先改入口准确CR，无运行窗口。原Worker`msg_f508bb720505`补报c2及2f普通push均实际TLS失败EXIT1，准确本地Git对象保留，不宣称远端更新或重试成功。

## 准确2f1eceaf的完整正文投递计划审核

局部发现：**准确`2f1eceaf3d5ad872339a2efba7fdbaa1178dbc78`的正文投递计划有B01/B02两项覆盖缺口，接口定稿暂停。** 在1b295先形成七条独立条件，完整读取本提交相对c2的两文档§4.4/§7.7差异，并核准确c0的真实structured host、sendPlan、performSend、delivery loop、operation store及消息来源。主控`msg_fc6fe552ac5b`转来用户首版须发现本机注册Agent并通过Swarm CLI调用的新要求，明确暂停2f冻结结论；本次只保留局部发现，不发送本提案正式接受或定稿决定，不宣称尚未实施的策略已经动态失败。

方案已覆盖真实批准全文及MessageRow逐项关联、仅产品Pi策略、ToolsOff及固定模型、同目标/fence、单封原SQL顺序、原样文本与成熟JSON边界、未知不mint、原R重复、同安装候选B的真实正文及业务结果。D01—D05分别覆盖来源一致、队列/去重、授权/归属、断开/未知、编码/大小，没有发现这些业务条件的多余检查；普通SDK存在、schema接受及模型自述继续不足以验收。

| 缺口 | 独立来源及影响 | 返修要求 |
| --- | --- | --- |
| B01 实际handover前的异步授权 | §4.4把host beforeRun称为实际消费边界。真实send-agent-turn、host-mutations及sendPlan的beforeRun为同步`function(): void`；mutation-plans.ts:163在performSend之前调用，performSend只写journal。后续delivery-loop.ts的独立serialized handOver及turns.ts的handOverSubmission才调用adapter.dispatch，等待启动和队列后没有再次执行原beforeRun。因此现有callback位置不能证明SDK实际输入前仍有当前授权，异步HTTP也不能依同步调用获得等待保证 | 明确定稿被await的真实handover准入接口及其消费者，等待返回后重新核原child/fence/consumer/归属及ToolsOff，保留原同步held和无产品策略宿主行为。D03必须明确覆盖journal已接收、SDK尚未输入的真实窗口和拒绝后原状态，未命中仍未验 |
| B02 原operation关联及公开状态契约 | §4.4把必要结果字段留待后继明确。真实原生send子receipt在nudge前完成，正文operation后分配；现有request.show只返回原R。structured_pointer_operations以mailbox为唯一键，put会覆盖，原delivery在queued/sent后标记消息并删除指针。仅声明保留原operation不足以定稿多封后续邮件和原R已完成时的查询来源 | 明确message/session到原operation的唯一持久关联、各queued/pending/accepted/unknown状态的保留或原journal恢复来源、后续消息不能覆盖旧未知的规则，以及worker.send/request.show准确结果字段和权威读取路径。复用既有R/journal/operation，不另建账本或把完成R解释为重发许可。D02/D04补原R已完成而正文仍排队、指针移除或后续消息到达时仍查询同一原operation及真实状态 |

B01已通过`msg_0d0472dcf593`、B02通过`msg_f7a1c12ab6a5`交原Worker按原写域补计划；已以`msg_f959d9bd7555`回复主控暂停定稿，新内部HTTP路由及结果消费者保持未定稿，后继须结合注册Agent要求完整审核。此处不指定另一队列或新权限系统。T12独立接受和Go源码检查顺序可按各自范围继续，F17完整功能、T11与所有动态结果保持未验。

`check-revision.ps1 -Previous c2d4b62d64eef73eb3c83064a8354696b6e623c7 -Candidate 2f1eceaf3d5ad872339a2efba7fdbaa1178dbc78` EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。只读真实源码和准确Git对象，没有调用新路由、操作数据库、创建消息、启动服务、SDK、模型、编译或执行D项；普通push实际失败的范围保持。

## 注册Agent要求的后继方案读取前条件

依据主控`msg_fc6fe552ac5b`转达的最新用户要求、首版任务当前前五阶段、原隔离/工具授权及最终同候选双Worker条件，先独立形成下表，再读取Worker后继完整方案。注册来源和具体支持范围尚未定稿，不从固定Pi路径推导其他本机Agent可用。

已读取reverse-engineer-anything的SKILL及连接指引。本会话没有REA MCP，公开`rea`与`rea-agents`解析到已安装npm入口；没有安装、修复配置或启动新分析。根`.verification/rea-orca/evidence-summary.json`是既有两份完整Evidence的派生摘要：focused `ev_c9cc8212af2b3b0ceb94199bac8fcec4775269e7f04272593e2e4d321d6ada8c`只分析12个所选JS、1963文件未选，ASAR边界`ev_0159c976d48ea721c101a8d95acbb071c4b0ad25a39e2b09e755f06729e5c43e`未选4280文件。已完整读取startup trace `ev_15961175dc22bbd52cec2f8d701c88f6c232b155b78282b0bdeac817167c9ba1`，只有一个startup IPC seed、无后继边；其自身保留partial及执行未知。未读取43MiB/182MiB完整应用Evidence，不声称摘要已经独立证明注册发现、完整调用闭包或无依赖；后继应复用准确目标Evidence并针对缺口补充。

准确2f源码的TUI定义和detection commands列出支持的命令/别名及运行平台；agent-command-detection核PATH存在及安装目录，尚未验证账号配置或实际启动。StructuredAgentRegistry从真实definition/adapter组成并核声明能力所需方法；当前产品STRUCTURED_AGENT_RUNTIME_CATALOG只有Pi。structured launch support另按settings/账号home/实际命令和version检查；完整读取的composition前半部显示provider/session/account及执行宿主须匹配。上述是普通Git源码观察，不能标为REA产物或本机动态调用结果。

| 条件 | 正确实现及必要真实验证 |
| --- | --- |
| G01 准确交付物及REA来源 | 固定用户本机Orca交付物的实际版本/路径/摘要及对应准确源码，REA从真实注册/发现入口沿CLI命令到原生启动、身份、输入、输出、停止及资源消费者追踪。保留选取范围、动态边和unknown；上游同版本源码或partial静态图不能证明交付物执行及依赖不存在 |
| G02 注册与本机状态分别可查 | 定义并定稿注册事实源、发现范围及刷新语义，分别报告定义/注册、命令实际安装、配置、可启动和原生控制能力；未知与已拒绝分开。命令别名/override、准确可执行路径、平台和版本依据真实来源，命令存在不代表登录或调用成功。发现不触发未授权WSL/远端、安装、账号设置或配置迁移，不输出凭据 |
| G03 选择结果进入真实调用 | Swarm CLI明确选择的已发现Agent须沿其真实命令和原生adapter启动，实际argv/cwd/env、账号来源、provider/model、Run/root/Dispatch/session和原生文件可对应。不得默换Pi、另一个Agent/账号/模型或执行宿主；未注册、缺程序、缺配置及不支持的真实能力立即拒绝。实际Agent自产会话/SQLite/事件逐项可定位，SDK模型可调用不代表该本机Agent已经被调用 |
| G04 每Agent的权限及能力 | 每条实际启动/后续输入/停止/恢复路径重新核当前项目、Grant、设置版本、目录和外发目标；当前工具限制继续适用。分别证明该Agent能关闭原生文件/命令工具及扩展/MCP，或报告真实不支持并保留依赖决定；不为满足发现要求默认开放工具。Pi原六项RPC仅约束该路径，其他原生协议/模型的准确允许范围先定稿；原strict caller/generation/fence与已完成R重放检查保持 |
| G05 完整依赖与独立安装 | 按准确执行闭包说明保留及裁剪的程序、动态import、native ABI、PTY/daemon、必要worker/sidecar、资源和许可，成熟库实际保留。不能以静态未连边判定可删，也不自行设依赖数量目标。最终安装目录在独立位置完整运行；对本机Agent程序及允许账号读取的外部依赖明确来源，不隐式依赖原开发树、原Orca资源、桌面IPC或原后台 |
| G06 原Orca及账号隔离 | 产品profile、配置/SQLite、缓存、端口/IPC、controller、Job与后代身份独立；发现可以按已定范围读取必要本机来源，不能覆盖原Orca/Agent、PATH、自动启动、账号/原配置或接管现有会话。安装前后、真实双Worker并发与指定停止后核原Orca公开CLI及实际进程/数据归属，秘密不进入日志、receipt或发布物；只改目录名不足以验收 |
| G07 失败、异步与未知 | 覆盖真实发现刷新/命令变化、无配置/权限撤销、错误项目/旧consumer、启动部分失败、busy/held、断开/超时/取消、输出失败、停止失败/未知和资源未释放。沿原R/operation保存已有效果及真实拒绝，未知不重发、不换目标或结束其他进程。B01真实handover授权和B02持久关联在适用的后续输入路径继续检查，注册发现不取消其业务条件 |
| G08 最终同候选任务 | 每条声称可调用的适用Agent路径有真实安装/配置/启动/控制证据，未验Agent明确保留状态。最终仍用安装后的同一准确候选完成原任务的至少两个真实Worker并发、分别查询身份/输出、只停止指定一个并核实际退出及另一Worker保全，取得符合固定业务判据的完整成果；覆盖选择的真实Agent及原件来源，不按两个模型回复或替代Worker通过。是否要求不同Agent类型不自行增加，按用户明确目标与后继定稿范围验收 |

这些条件直接来自新发现/调用目标及已有权限、隔离、输入和最终验收，不增加自我证明检查。注册的具体含义、可支持Agent集合、账号读取和各Agent工具/协议能力在准确方案定稿前保持未决定；不冻结选择或降低原验收。编译、安装、实际probe/Agent/服务/模型与新REA分析均未执行，没有取得重操作窗口。

## 准确c4d1c140的Go源码检查入口静态审核

结论：**接受准确`c4d1c140d8d231058c42c455bd8433de0fc256c2`相对2f的四路径局部静态修改。** 按已接受c2计划及先前六条件完整读取差异、整个compile-go-tests.ps1及实际T03消费者。原入口移除Candidate/VERSION前置，只接收本项全新绝对私有输出；真实40位干净HEAD在创建输出前核查，四包逐一`go test -c -p=4`、私有TEMP/Go缓存、Go4/512MiB及BelowNormal保持。三项传输前真实函数前补资源准入，逐函数PASS和任意SKIP拒绝；任一命令失败立即退出，最终再次核同一干净HEAD后才写原SOURCE/tests.json。

T03在创建验证目录、资源检查或实际服务之前，核同一安装路径、真实Node、绝对Go目录及实际`VERSION=SOURCE=准确干净源码`。原四个live函数的逐项PASS、零匹配/SKIP拒绝及原controller关闭保持，没有用源码编译代替安装服务。原T03缺失的人类撤销安全停止、不同项目/Run及旧代范围、T12/正文和新增注册Agent要求仍各自未验。未发现六条已审条件的剩余静态缺口；此修改不接受继承2f的新接口或产品功能。

对准确Git对象的两个PowerShell脚本使用已安装`System.Management.Automation.Language.Parser.ParseInput`在内存解析，均PARSE_PASS/EXIT0，没有执行脚本。`check-revision.ps1 -Previous 2f1eceaf3d5ad872339a2efba7fdbaa1178dbc78 -Candidate c4d1c140d8d231058c42c455bd8433de0fc256c2` EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。原Worker`msg_8bbc73a957da`报告按主控要求暂不另跑网络，c4仅本地提交、尚未push；没有编译、函数、probe、服务、Agent或模型结果，本Reviewer没有取得运行窗口。

准确c4接受hook `msg_32ef943e003c`已获主控`msg_1733606dc435`完整读取af666报告及准确四路径/整个编译入口回复，实际唤起确认。主控明确普通Git push授权保持，先前不另跑网络针对REA/HTTP探测；原Worker`msg_69bb54bdc6ad`后继报告一次普通push实际EXIT0，远端由ebf到bf，包含c2/2f/c4/bf；原c2/2f TLS失败事实保留，未将后继成功追改为首轮成功。

## 准确1dc4cb5e的恢复身份及当前REA原件

结论：**接受控制脚本准确SHA256 `1dc4cb5e5d35555a474792540535fcb081021cfe562b77427c07f870413290ab`的两项恢复身份常量。** 在内存把57400/UTC15:33:20.748677恢复为18944/UTC14:49:56.491641，全文件精确得到已审`03c64ef6922af3e4ce00763159377c26a03d8d065eb36e0b32e20b65ba577ceb`。沿既有真实恢复/所属/两常量判据，没有其他流程修订。

UTC15:36:13独立只读确认restore-06 launch/restored/current running一致、准确原程序、PID57400/父32056、Get-Process创建UTC15:33:20.7486771、原件epoch1792510400.7486773，差0.0001毫秒符合既有小于1毫秒判据，唯一127.0.0.1:64505监听。database-check原件integrity ok、foreign_keys空、stopped_interval_comparison equal；启动资源CPU6.2%/5.7%、可用11.967/11.985GiB、BelowNormal、间隔3.001819秒符合门限。prepare-06/stop-06的prepared/stopped/window-ready及真实backup资源原件已读，原S18944精确退出15后恢复同一原库；未打开SQLite或执行HTTP。

`.verification/verify-restored-owner.ps1`以本次准确hash和前后身份运行EXIT0，原件为`.verification/window-1dc4-readonly.json`。该结论只允许主控将准确脚本用于后继新的prepare，仍须新鲜身份/环境/备份及资源准入；没有授予新的动态窗口，当前窗口归原S57400。health/Web、完整环境与本机Orca38652/runtime49da公开状态保持主控实际来源，本Reviewer没有补写动态通过。

已读主控限定inventory identity/分析exit/资源及doctor exit；成熟asar记录当前Orca1.4.230和REA6.5.0、目标6a9bd961，分析实际EXIT0。使用已安装stream-json3.7.0的parser/filter/streamValues对96945279字节完整JSON流进行语法读取，只组装Evidence身份、所选来源/统计、graph coverage、semantic coverage的status/truncated/limits及limitations，输出`.verification/rea-current-static-review-02.json`，BelowNormal/Node512MiB/EXIT0。未组装或逐项审核完整图、unknown_ids或全部AST事实，不声称已完整核查其语义。

准确Evidence `ev_7003a444cd2970c2d3ace2427c883c461f6a1ec67b9dfdbe9be6528e775726d3`绑定当前6a9完整摘要，7来源/6JS、33407文本字节、解析失败0、5051文件未选，graph和semantic均partial/truncated=false。其limitations明确未选关系及动态路径未知，静态语法不证明执行。doctor原件healthy=false、Codex registration为config_drift、实际EXIT1；CLI分析独立成功，没有修复MCP、启动新REA分析或执行Orca目标。主控首次派生摘要失败及后继修正保留其原记录，没有覆盖为应用分析失败或重复应用分析。

## 准确bf996c34的注册Agent方案审核

结论：**退回准确`bf996c3465a8a4efb06e9e2a598dd6aab3ec6c2f`的具体注册发现/调用方案，保留G01—G08覆盖框架及全部真实未知。** 在2593先形成独立条件后，完整读取相对c4两文档§4.5/§7.8，并核真实注册handler、启动设置投影、worker配置预检和local-worker-start消费者；普通Git源码观察与当前REA限定证据分别记录。原Worker后继普通push成功已取得，不把本地可审误写为未提交。

| 问题 | 独立依据与影响 | 原Worker需要补齐 |
| --- | --- | --- |
| N01 发现依赖原Orca后台 | §4.5.2每次发现及启动前都要求原Orca runtime的agents/settings响应，不可达即unknown；§4.5.4又要求安装运行不依赖原桌面IPC。必要原服务前置与独立抽取边界冲突。主控`msg_f513a3e06a00`确认新用户目标包含独立发现和调用，原Orca无须正在运行 | 使用产品复用的真实registry/adapter和产品本机命令发现，明确必要原配置的只读事实源及敏感筛选，原服务只能作可选对照。补原后台未作为发现前提的独立安装检查，同时按已有范围保全原Orca，不为测试终止它 |
| N02 选择与调用契约未完整 | 框架提出capabilities五状态及agent/model选择，但准确输出schema、选择快照/当前配置变化绑定、模型/账号定位、失败/unknown字段及各CLI/HTTP/App/helper/native消费者仍未定。真实agentSession.agents受协商和client策略筛选，返回build声明；settings.getAgentLaunch只返回只读投影；worker别名与TUI路径的实际launcher不同。只列方法名不足以冻结原生caller、consumer及调用边界 | 提交准确发现/选择/执行及结果契约和消费者清单，说明原生声明与本机实际观察的对应、启动前变化拒绝、原controller/Run/root/Dispatch/session/generation，以及B01/B02在实际输入路径如何实现；不得在实现时另选协议或字段 |
| N03 真实可调用范围证据不足 | 方案逐Agent列出工具/账号/供应来源未知，却尚无实际支持矩阵及可执行的原生launch/send/read/stop/resume路径。当前ev7003只选配置/检测规则和两个CLI入口，未覆盖注册handler到worker控制及动态闭包。发现后全部拒绝不能满足实际调用和最终双Worker目标 | 按准确交付物补必要链路/依赖来源和逐Agent程序、原生协议、账号读取/写入边界、产品能力及本轮实际验证权限矩阵；必须用户决定的范围保持未决定，由主控依据原决定处理。当前工具限制按其适用验证范围核实，不能扩大为所有产品能力的禁令，也不能推断开放。未具备依据的实现保持暂停，已审独立Go准备可继续 |

G01—G08及既有D/T覆盖了来源、状态区分、原生调用、权限、独立依赖、原Orca保全、失败未知及同候选双Worker任务；未发现无关多余检查。它们是后继完整方案的必要验收条件，不等于当前未知接口或Agent范围已接受。N01—N03已通过`msg_9816edcff224`交原Worker，主控另已核相同后台依赖问题；修复全部交原Worker，本Reviewer没有修改实现。

`check-revision.ps1 -Previous c4d1c140d8d231058c42c455bd8433de0fc256c2 -Candidate bf996c3465a8a4efb06e9e2a598dd6aab3ec6c2f` EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。所有G/D/T12真实Agent/协议、注册查询、安装、模型与新接口仍未执行；没有重跑REA、采样、服务、控制端或网络探测。

准确1dc4接受hook `msg_eb869aeebe1c`已获主控`msg_0c5569d36751`完整读取60b无界面报告新增回复，实际唤起确认；主控仅在接受及新鲜身份/环境依据下启动prepare-07，尚未返回或交付Go窗口的实际时点保持。准确bf退回hook为`msg_b30e747ca85f`，该回复确认N01—N03已完整读取并由原Worker返修；正式bf线程回复尚待取得，不把脚本接受扩大为方案或动态接受。

准确bf退回hook已获主控`msg_bc3f2574082a`完整读取60b该节并正式回复，实际唤起确认。主线1b管理接受hook `msg_a6e45f5a75cc`亦获主控`msg_8a98651ebe6c`完整读取a58报告回复，累计c9管理范围及各自时点保持。

## 准确9c0b62a3来源及独立源码检查目录

完整读取准确`9c0b62a39b04cdd3fc3d210ced200277824a4e75`相对bf两文档的5增1删；当前REA6a9/ev7003、选定6JS/5051未选/partial及后继普通push记录，与已独立读到的准确原件一致。库存chunk只标后继核查候选，没有当已分析或实际调用通过。c4至9c的scripts/tests/cmd/runtime/internal/go.mod/go.sum差异为空；已审入口及程序未变，bf的具体方案退回继续有效。对应两Markdown/5本地链接及Git空白EXIT0。

原Worker连接错误后，主控`msg_9000137ae158`明确撤回其全部Go/重操作授权，保留原树两文档WIP；没有实际运行回执或输出目录，不把网络错误当程序退出。主控另在根忽略目录`.verification/swa-go-9c-01`创建独立detached准确9c，仅执行同一已审四包/三函数检查，不修改程序或冒充Worker自测。初次只读看到locked initializing和检出中间状态，已以`msg_5c27d5bbf73d`通知等实际检出完成，没有删除、解锁或判源码失败。

收到主控`msg_2e25586799de`的checkout EXIT0回执后，使用`git --no-optional-locks`独立重核该目录HEAD为准确9c、status条目0、两脚本相对已审c4无差异，命令EXIT0；结果已通过`msg_d2e867af5ca6`回复。实际Go资源/下载尚在途采用主控来源，没有运行第二套编译或将原Worker未运行登记为通过。原S57400按stop-07停止，本机Orca保全，执行结束后由主控恢复同次原库与环境。

## 准确740ccdf9的独立发现及选择契约审核

结论：**接受准确`740ccdf9a9dfa33901274fd8dccd37c4beacf1fd`所含独立发现、明确设置源、capabilities DTO和AgentSelection/现行外发批准的技术方案；新增输入及普通TUI路由仍退回补齐。** 完整读取5fc相对9c及740相对5fc的两文档差异，按G01—G08、N01—N03和B01/B02核查，并重读原Profile只读API、原mutation executor及既有Go外发消费者。实际Agent支持/工具适用范围和全部动态验证保持未决定或未验，不由本次技术方案接受推断执行授权。

N01已解决：注册来自产品实际definition/adapter构造的StructuredAgentRegistry，本机命令与配置在产品进程检测，原Orca不构成发现/选择/调用前提。标准native源和显式可选`--agent-profile`分别规定；明确源缺失/损坏即失败。已核classifyProfileStateStorage只读分类，openProfileStateDatabaseReadOnly使用readonly/fileMustExist且不迁移，readProfileStateDomains在同一只读snapshot返回settings/revision并区分unreadable；legacy并存时仍须核原acceptance，禁止Store/migrate/marker写入。JSON-only使用成熟JSON流筛选，不读取或复制原会话。

capabilities经既有HTTP/App observe、只读Capabilities端口和私有swa.capabilities取得实际注册、安装、配置、可启动及各控制能力的独立状态，缺来源保持unknown；不调用模型刷新或接受HTTP身份/授权声明。740补齐ResolveSelection及稳定AgentSelection，把真实CLI ID、程序/配置版本、model/provider及external_target绑定至原R；App先核observe再按实际单CLI批准和当前Grant核start/目录/工具/外发，send沿原Dispatch选择复核。既有Target=pi及适配器/真实检查的固定断言明确纳入同一更新范围，未扩大ExternalModelCLI单CLI语义。原controller/父子R、strict caller、native队列及完整成果消费者列明，B01实际handover异步复核与B02固定原receipt/journal关联已有明确方向。

| 仍需补齐 | 准确原生依据及要求 |
| --- | --- |
| N04 新swa.workerSend的异常结算 | 新方法是非prompt durable mutation。原executor在普通异常且非operation_unknown时会discard pending，markEffectPossible和checkpoint本身不能保留该receipt。方案须明确首次可能host输入前后的边界：效果前已证拒绝如何结算，checkpoint后发送/输入结算/存储异常或证明不足如何保留同一R/native、session/operation和占用。不能只由外层parent的unknown代替子receipt，也不能在子receipt丢失后再次mint；G07须覆盖该真实边界，未具备条件保持未验 |
| N05 结构化与普通TUI的路由 | 当前worker.send表统一Dispatch→私有session→StructuredHost，但其余TUI表描述原生PTY且可能没有该session；StructuredAgentRegistry本身也只登记真实definition/structured adapter。须按execution_mode明确普通TUI的真实注册来源及start/send/read/stop/resume路由、父子R和输入结算/未知语义，或者准确标明哪些能力未实现；不能给无session的目标套StructuredHost，也不能用TUI常量补造注册或把终端尾部当完整成果 |

N04/N05已以`msg_e527bd2d50de`交原Worker，主控已获`msg_d2e867af5ca6`说明独立发现可单独接受；新输入方法及模式适配不得据此提前实施。readonly发现和稳定选择契约不依赖工具范围待答，可由主控先写SPEC并授权相应原写域；实际启动/模型、原生工具及账号写入条件仍由当前决定和准确Agent能力核实。G/D/T验收没有删除，未发现额外无关检查。

`check-revision.ps1 -Previous 9c0b62a39b04cdd3fc3d210ced200277824a4e75 -Candidate 5fc799e1dd9beda8f2978d4a04f1a1f8f9f2ffd3`及以5fc为基线、740为Candidate的检查均EXIT0，各两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。本Reviewer没有访问原profile数据库、实际注册配置、启动Agent/服务/模型、编译或运行新REA。

## 准确9c的实际Go检查结果与fea0恢复身份

结论：**接受准确`9c0b62a39b04cdd3fc3d210ced200277824a4e75`的主控独立四包源码编译及三项传输前检查结果。** 主控在准确干净detached目录仅执行一次已审`compile-go-tests.ps1 -OutputRoot C:/Users/DW/orca/Astrocyte/.verification/swa-go-9c-01/.verification/swa-cli-headless-20261020/go-source-01`，实际会话81611 EXIT0。已独立完整读取根control/go-9c-execution-01.log、输出SOURCE/tests.json、三函数日志及五份资源原件，并核四个实际测试二进制存在；SOURCE为准确9c、末尾成功摘要明确actual_runtime_tests_executed=false/model_calls=0。

原四包test-c通过，实际运行的三个函数为TestOrcaCLIRejectsAuthorityArgumentsBeforeTransport、TestOrcaCLIRejectsUnknownActionAndIrrelevantInput、TestOrcaCLIMissingCurrentGrantIsDenied，均有真实RUN/PASS且无SKIP。四次编译和一次传输前检查资源各passed=true、单调间隔约2.50—2.51秒，CPU最高60.4%、可用内存最低12051972096字节，符合当前门限；原入口Go4/-p4/512MiB、私有缓存及BelowNormal保持。主控已确认所属Go/compile/link/PWSH两次新鲜CIM无自有执行，本Reviewer没有运行第二次编译或二进制。

此结果只覆盖主控该独立源码检查，不登记为Worker自测，也不证明四个live函数、服务、真实当前项目批准、原生controller/SQLite、Node/native包、安装、Agent/模型或同候选双Worker通过。旧两次构建失败、N04/N05及待答工具范围保持；源码改变后该SOURCE不能作为新候选测试。后继官方HEAD观察保留主控实际来源，不从HEAD200推断完整下载或原TimeoutError原因。

结论：**接受控制脚本准确SHA256 `fea0faa051e6fe89674b2a4dfe3c53abb27990f959d7d0d6aa17cd4454bff57c`的两项恢复身份常量。** 内存逆置21300/UTC16:05:49.906991到57400/UTC15:33:20.748677，全文件精确得到已审1dc4。UTC16:08:08只读核restore-07 launch/restored/current与实际PID21300/父45632、原程序、创建UTC16:05:49.9069910及epoch1792512349.906991完全对应、差0毫秒，唯一127.0.0.1:64505监听。

原库记录integrity ok、foreign_keys空、stopped_interval_comparison equal；启动资源CPU9.2%/11.7%、可用11.634/11.714GiB、BelowNormal、单调3.001725秒合规。`.verification/verify-restored-owner.ps1`以本次准确参数运行EXIT0，原件`.verification/window-fea0-readonly.json`；未执行控制脚本、SQLite、HTTP或采样。health/Web、完整原环境/配置及原Orca38652/runtime49da状态保留主控来源；当前窗口归原S21300，下次prepare仍须新鲜身份/环境/保全和资源，脚本接受不授予新窗口。

准确740的独立发现接受hook `msg_ac171707a1d8`、输入分支退回hook `msg_d5756727a7cb`已分别获主控`msg_78bd22dfc244`和`msg_5c0ab07b0334`完整读取c3报告及准确差异回复；独立发现已按SPEC §21.1.1授同一Worker静态实施，实际读取唤起确认。fea0接受hook `msg_358ed21f4cca`及9c源码检查接受hook `msg_9e886939fb1a`已分别获主控`msg_1987104c919f`和`msg_0163867d64b0`完整读取bb5准确版本回复，实际唤起确认，各自局部结果及全部动态未验保持。

## 准确d441703a的N04/N05技术计划复审

结论：**接受准确`d441703a620f2e1bcca73edd50c7edc66dc92351`所含新workerSend异常结算及structured/terminal分流技术计划，N04/N05已覆盖。** 完整读取763相对740及d441相对763的两组准确差异，再对照c3的两项独立缺口、B01/B02与G07。主控通知d441已普通push；程序WIP只属于已授独立发现实现，本次只读取准确Git对象，不以WIP替代方案或功能结果。

N04现明确child R/native自身先持久checkpoint实际模式、操作和正文来源，在input-submit/mark之后才进入一次真实host或terminal writer；该边界之后异常、失联、超时及journal/结算失败均使用operation_unknown保留原child/parent及占用。效果前只有正向证明尚未输入才持久完成原child的真实rejected/effects_applied=false，再由parent保存同一拒绝；存储失败继续unknown，不删除或重发。已核RpcContext真实recordMutationReceipt、原executor的recordReceipt完成路径、mutation-receipt-store的complete可更新准确已有记录且discard仅删除pending；handler先持久完成结果可以保留外层重复完成失败时的原件，单独mark不能承担该保证。

N05现按固定execution_mode明确来源及全过程：structured来自真实registry/definition/adapter，以实际Dispatch/session/fence及原native operation进入原host队列；terminal只有真实TUI配置与已接入PTY launcher/controller共同绑定才登记，以原handle/PTY/incarnation/创建身份、正向Agent识别和readiness进入普通sendTerminalAgentPrompt。后续正文不用仅dispatch-preamble的terminal arm；实际mode变化不换模式继续输入。终端read明确原生transcript与显式有界terminal来源的区别，stop/release使用真实PTY custody及精确树退出，resume缺原生证据保持unknown/unavailable，不创建替代会话。

已核真实terminal-input-dispatch的串行generation校验、terminal-agent-prompt及RuntimeTerminalWriter在实际写入/各chunk/提交前await beforeWrite的来源；计划要求产品异步授权在实际队列出队和写入前执行，等待后重新核身份/ownership，并保留原身份guard。structured继续在实际serialized handOver到adapter.dispatch前await授权，复核相同child/consumer/fence/工具；同步beforeRun不承担实际输入授权。DTO分别使用真实session或terminal关联，handed-to-terminal只表示PTY接受，不能当模型accepted/turn started。原receipt及实际模式原件承担持久关联和只读查询，不添加另一队列或请求账本。

G07补效果前明确拒绝、效果后子receipt保全、真实存储失败、逐模式来源及mode冲突；未实际命中继续未验，没有注入假异常或伪造事件。成熟marked AST核763时G表缺独立G08行，d441已恢复G01—G08同一表，补说明在表外，原最终同候选双Worker条件完整保留。另以`msg_94aa1086713a`交原Worker在常规文档更新中同步查询摘要的逐模式措辞，准确DTO及详细查询契约已经分别规定实际来源，不阻塞本次技术结论。

`check-revision.ps1 -Previous 740ccdf9a9dfa33901274fd8dccd37c4beacf1fd -Candidate d441703a620f2e1bcca73edd50c7edc66dc92351` EXIT0，两Markdown/5本地链接及Git空白通过；`.verification/inspect-plan-tables.mjs`使用marked18.1.0读取准确763/d441，确认d441八行同表，BelowNormal/Node512MiB。未发现剩余N04/N05计划缺口或新增多余检查；本次只接受技术方案，由主控定稿SPEC再授权实施，原生Agent/模型/工具范围、T11/T12及全部真实输入/恢复/安装/同候选双Worker结果继续未验，没有新增运行授权。

## 准确092b2613的恢复身份及注册链补充原件

结论：**接受根控制脚本准确SHA256 `092b2613de7f6be3bd13356968029a9125777d446778304edf8bc7e303862433`的两项恢复身份常量。** 内存逆置41264/UTC16:23:39.495409到21300/UTC16:05:49.906991，全文件精确得到已审fea0。UTC16:33:29只读核restore-08 launch/restored/current及实际PID41264/父54292、原程序、创建UTC16:23:39.4954090/epoch1792513419.495409差0毫秒、唯一127.0.0.1:64505监听；原库记录integrity ok、foreign_keys空、stopped_interval_comparison equal。

restore-08启动资源CPU6.0%/10.1%、可用11.679/11.662GiB、BelowNormal、单调3.001391秒符合当前门限。`.verification/verify-restored-owner.ps1`运行EXIT0，原件`.verification/window-092b-readonly.json`；没有执行控制脚本、SQLite、HTTP或采样。完整原环境/配置、health/Web及原Orca38652/runtime49da公开状态采用主控实际来源；当前窗口归原S41264，准确脚本供后继新prepare使用时仍须新鲜身份、环境、保全和资源，本次不授予新窗口。

已独立读取当前注册链REA的exit/resources/summary及101314219字节原JSON。`.verification/read-rea-fields.mjs`使用stream-json3.7.0 parser/filter/streamValues在BelowNormal/Node512MiB下只组装身份、选取来源、统计、coverage/limits及limitations，输出`.verification/rea-registration-static-review.json`，EXIT0；未组装完整图/unknown_ids或逐项审核所有事实，没有重新执行REA或目标程序。

准确Evidence `ev_4e2d7df01d23ced0410fc4f5f74584e6df4467159c03d01d19d5ba5efe62b7f9`绑定当前6a9交付物，实际CLI EXIT0。选取10个JS共896699字节，包括真实注册chunk、TUI配置、registered-agents/capabilities/account-home、launch command/override/defaults及两个RPC参数，解析失败0、5048文件未选。graph为partial/truncated=false，semantic为partial/truncated=true；原件明确注册chunk语义图超过parent剩余结果预算，static/module事实保留。不能由静态保留、声明能力或sender检查推断完整调用、授权执行、当前配置、依赖闭包或模型成功，缺失语义保持未知。

原host Node24.16.0默认global fetch只对固定官方Node24.21.0归档执行一次30秒HEAD，原result为HTTP200/841.8795毫秒、retries=0、proxy_environment_present空，实际exit0。资源CPU8.2%/9.4%、可用11.850/11.821GiB、BelowNormal、约3.002秒通过；原MODULE_TYPELESS警告保留，没有据警告修改上游配置。本次仅证明该时点原传输HEAD返回200；完整下载及旧TimeoutError因果仍未验，没有新建或安装候选、启动产品服务或调用模型。

主控`msg_3d3c842efa23`提出的首装人工授权入口检查纳入既有T02/T03及USAGE：后继准确候选须给出新DataRoot下实际可用的首次入口、真实session/CSRF与项目设置/Grant/token完整操作，必要原资源可复用，不能依赖用户猜正文、开发树或原服务；不因包列表就断言没有路径，不代human批准、不另开Web重构模块。先核当前真实实现，只有具体缺口交回原Worker；该检查不阻断已接受的独立发现代码。

准确d441技术计划hook `msg_885e2ea87b6b`及092b控制脚本hook `msg_b607d65e257b`分别获主控`msg_fbb1dd4189d6`和`msg_084e2c28f2d0`完整读取739/5b准确报告及差异的回复，实际唤起确认。后继main a6f已将对应技术契约定稿，原Task继续，不因一次静态决策结束。

## 准确61a17226的发现与选择实现审核

结论：**退回准确`61a172264d7d3784c63ab8a100af6382ebe78c85`，修正实际启动消费者、可证配置选择及来源/策略准入后再交准确CR。** 已完整读取相对d441的31文件差异，并分别追踪原生factory、launch resolver、adapter、Pi builder、账户home、Go项目批准与设置入口。只读发现的来源组织及当前外发目标调整已有实现；该提交尚不能接受为完整稳定选择与原生调用实现。全部问题交同一原Worker，没有修改程序或执行Agent。

| 问题 | 准确源码与影响 | 返修及必要验证 |
| --- | --- | --- |
| N06 Pi factory与实际builder不兼容 | `rpc-runtime-registration.ts`改用`piRpcLaunchPart({})`，resolver返回无sdkEntry且fullAccess=true。`PiRpcSessionAdapter.acquire`仍消费导出的`buildPiRpcLaunch`，该函数明确要求sdkEntry且fullAccess=false，否则抛出固定SDK入口错误。未导出的另一个builder没有调用方；当前原生Pi路径在spawn前必拒绝 | 修实际消费者及原生argv/策略，核全部相关调用方，保持未获批准的工具路径拒绝；后继真实Pi检查必须以该本机CLI及实际自产会话证明，不能计SDK会话 |
| N07 配置已读分支仍无法形成选择 | `configuration('claude')`识别settings中的API key或auth token并设置account，但provider只有OAuth token存在才赋值；无OAuth的该分支永远不能通过model/provider/account全有的条件。Codex仅取显式model_provider，未核原生已可证默认来源；其他registry项全部固定native-configuration-unobserved，`selectAgent`统一拒绝 | 按具体Agent的真实配置/默认及明确模型选择核事实，不填造provider或伪装配置成功。明确已支持和仍缺来源的路径，必要真实Agent正向选择不能由清单加拒绝替代；当前没有读取本机秘密或断言每个Agent已配置 |
| N08 发现来源未绑定实际运行账户与工具策略 | `agent-discovery`从原home或可选profile读取配置；实际host仍使用私有Store/环境，Pi的PI_CODING_AGENT_DIR固定私有DataRoot/pi。launch admission只返回所选命令并核程序、原session/model及重读发现结果，没有比较实际record.accountHome或实际launch账户/策略与获准来源。Go authorize构造的ProjectActionScope.Tools为空，capabilities中的allowed_tools和tools unknown仅为显示，不能证明启动前的工具准入 | 继续只读发现可独立进行。实际调用必须使用同一已批准来源及已证隔离，按当前工具限制在首次输入前正向核准；来源、策略或写原home风险未证时明确拒绝。不能通过把原账户home设成可写运行目录或把旧工具限制推断开放来补齐 |
| N09 原生注册与人工外发批准入口仍不一致 | 实际`SetProjectSettings`仍以旧`s.registry.Adapter`校验ExternalModelCLI，无法为仅在新Native registry中的Agent保存现有单CLI批准。此入口还明确拒绝非空AllowedTools，实际语义必须与未答工具范围共同保留 | 主控处理必要窄接口/写域，同一Worker复用真实注册验证并保持旧S3职责；禁止造空Adapter来放行，禁止自行改变工具批准。新增批准入口与具体Agent实际调用均须后继验证 |

N06已通过`msg_dba72e4e9ac0`通知主控、`msg_0c9771247145`交原Worker；原Worker`msg_101f553d3180`实际读到并核原件，返修继续。其余发现独立依据上述源码，主控关于默认来源、旧注册批准及账户运行环境的来信仅作为检查线索。没有编译或启动假协议来证明这些分支。

正向静态范围：serve不再以固定Pi SDK入口/OAuth文件作为启动前提，原本机home/AppData仅作为显式发现定位，实际子进程私有home保持；显式profile使用成熟流JSON或原read-only SQLite/snapshot和acceptance读取，不创建原Store或迁移。注册来自实际StructuredAgentRegistry，禁用项和真实version probe各有来源，启动/完整读取/恢复/工具缺证保持unknown。App先observe解析实际选择，再按当前Grant、revision和真实单CLI外发目标核动作，稳定选择进入父R并在实际launch前重读；上述方向不消除N06—N09。

`check-revision.ps1 -Previous d441703a620f2e1bcca73edd50c7edc66dc92351 -Candidate 61a172264d7d3784c63ab8a100af6382ebe78c85`实际EXIT0，三个Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。Worker所报gofmt/TS语法/架构检查只作为其静态结果，不当作本Reviewer执行或功能通过。生成、类型检查、编译、HTTP/SQLite/Native正文、安装及同候选双Worker均未执行；新输入/T12准确技术计划已接受，程序按各自准确提交另审。

首装实际入口另已核：`server.go`在LocalProjects服务存在时注册`GET /api/v1/auth/session`，与SPA资产无依赖；session guard在无Authorization且Origin合法时建立真实cookie/CSRF，写路由继续验证它们。准确61a的USAGE只列项目/设置/Grant/token路径，缺初始session获取及完整合法正文、实际响应ID/revision的人工步骤；已交原Worker补在既有T02/T03及说明，无需由包内缺Web推断无授权入口。本Reviewer未执行human bootstrap、签发token或代用户批准。

61a退回hook `msg_b118bbad9a61`已获主控`msg_9c01fff3163f`完整读取be1准确全节及实际消费者源码的回复，实际唤起确认。N06—N09交原Worker继续，独立发现技术方向和既有计划保持，程序/实际调用未接受。

## 准确c8fdf318的恢复原件身份准入

结论：**接受控制脚本准确SHA256 `c8fdf3188ea97f1334b22bd261e578db2bf50a2f5039ddf944613753b42d1176`的实际恢复原件参数准入。** 已完整读取18384字节当前脚本并核准确摘要。prepare必须显式提供`--expected-owner`，路径严格存在，resolve后父目录的父目录须为脚本所属control根，文件名须为restored.json；省略、越界、错误文件名或其他phase携带该参数均在控制正文前拒绝。prepare读取当前RUNNING并与所选既有恢复原件整份JSON相等后，才调用实际owner、环境/配置、源码、资源及一致备份检查，不从输入文件声明跳过真实进程。

原PID/创建时间固定常量已由该实际原件替代。owner仍核准确S、程序/Web/数据库实际路径、创建时间差小于1毫秒、BelowNormal及唯一127.0.0.1:64505监听；prepare重复核环境/配置及身份，stop仍要求实际prepared基线，恢复仍核正向退出、端口、原库完整性/外键/停止区间内容及同次原环境。start/reconcile的原有创建与只读观察区别、恢复元数据冲突拒绝保持，没有新增状态账本或自动重试执行。

静态`ast.parse`当前Python源码EXIT0，只解析语法，没有运行脚本正文；Node512MiB以成熟JSON和isDeepStrictEqual读取restore-08/restored.json及当前running.json，整份相等，实际记录仍为S41264/epoch1792513419.495409。上述只读查询没有重复资源采样、访问SQLite/HTTP或操作服务。该次源码接受允许主控按新参数准备后继具体窗口；实际prepare仍须新鲜身份、环境、资源及保全全部通过，当前窗口仍归原S。

准确c8接受hook `msg_72ea71c99c18`已获主控`msg_f7638e45b5c4`完整读取aed4准确全节及当前源码回复，实际唤起确认。主控继续安排准确CLI交付物的限定REA，原app.asar分析只作为对应后台来源；本Reviewer没有据新任务自行运行REA或目标程序。

## 准确16ad5ef6的N06局部返修、首装说明及N09技术计划

结论：**接受准确`16ad5ef6fdbef17b54d8ce04224f0dca3abfafc3`中N06实际Pi builder局部静态返修、随包首装人工流程和N09外发批准注册端口技术计划；整体发现/选择实现及真实调用仍未接受。** 完整读取7f相对61a的五文件和16ad相对7f的两文件差异，重读实际Pi adapter/permission枚举、session guard、空间/项目/设置/Grant/token源码及真实StructuredAgentRegistry；USAGE在16ad相对7f无差异。

N06当前导出buildPiRpcLaunch由实际PiRpcSessionAdapter消费，构造真实本机Pi的`--mode rpc`及原生session/fork参数，明确拒绝sdkEntry。resolver使用原isAgentChatPermissionMode校验basis或record的实际值，仅合法full-access且原resolveFullAccess允许时形成true；缺失或未知值保持拒绝。此局部消除固定SDK builder与原生factory的不兼容；实际账户/工具策略、Job、原生握手及成果继续按N07/N08和现行验收核实，未运行Pi或模型。

首装说明使用同一安装服务GET auth/session的真实cookie/CSRF，实际创建空间并取space.id、登记公开目录并取project.id/settings.revision，保持完整settings与expected_revision，Grant及token分别校验实际归属。初始只有observe，之后实际status/capabilities供人类审阅；执行批准明确等待真实Agent能力和当前授权，单CLI批准、目录/空工具和旧注册限制如实说明。实际一小时token、重新签发或Grant更新失效、human会话重取及revision冲突处理均与源码一致；调用端ID与原生Agent ID分别保留。没有自动批准或要求开发树/原服务/Web资产。

适用T02/T03计划覆盖同安装/新私有根的真实人工session→space→project/settings→Grant/token以及状态/发现，实际库逐项定位身份和revision；拒绝Agent bearer管理、缺CSRF、过期会话、旧revision及旧token。人工请求由用户本人执行，未命中或没有具体候选时保持未验；当前没有申请用户提前批准或执行这些正文。`.verification/check-usage-snippets.mjs`用marked18.1.0读准确7f的14个PowerShell片段，再以原生Parser.ParseInput逐项仅解析语法，EXIT0、body_executed=false，BelowNormal/Node512MiB。文档数据以UTF-8/Base64交给解析器，没有执行其中命令。

N09先独立提出新项目批准不能循环依赖既有Grant/CLI批准、真实注册来源、失败无设置写入及旧HistoryRoots/S3职责等条件，并以`msg_034c86ed6be0`交Worker；之后对照7f/16ad方案。只读ExternalCLIRegistry.ValidateExternalCLI由同一真实orca adapter在HTTP开放前一次注入；nil/重复配置失败，无Orca配置才保持原旧registry路径，已绑定端口失败不能换源。只改变非空ExternalModelCLI验证，原HistoryRoots、旧NativeAdapter和AllowedTools限制保持，不造空Adapter。

16ad已明确strict registered_agent请求/Native params及RegisteredAgent/OrcaEnvelope字段：真实ID/adapter/mode及成对source.kind、registered=true、schema、同安装VERSION/host.cjs来源、私有runtime/authority/data/install根。Native取实际registration，Go逐项核与请求、真实RPC包络及Config根一致，全部通过才返回nil；缺失/错误/未知来源失败。新项目人工校验不依赖已有批准，不执行完整发现、命令版本、会话或模型。安装来源摘要复用已有字段，不另建执行账本。

对应计划覆盖旧四CLI以外真实注册ID的正向人工批准、真实未知ID/未绑定模式/服务不可达失败且设置不变、另一选择不匹配时模型前拒绝、无Orca旧行为、HistoryRoots/S3及工具/身份/revision回归，与独立条件相符，没有多余自证用例。技术计划由主控定稿SPEC及必要窄写域后，同一Worker实施；它只证明拟校验注册，未证明该Agent安装/账户/工具/可启动。N07/N08剩余、N09实现和首版真实安装/双Worker全部未验。

`check-revision.ps1`以61a→7f及7f→16ad分别检查EXIT0，前者三Markdown/5本地链接、后者一Markdown/1本地链接及Git空白通过；成熟marked18.1.0、BelowNormal、Node512MiB。没有运行编译、类型检查、HTTP/SQLite正文、实际Agent或模型；两个既有构建失败及当前权限未答保持。

准确16ad局部接受hook `msg_56c9462ad896`已获主控`msg_92795a7f47c5`完整读取b64新增全节及7f/16ad两组完整差异回复，实际唤起确认。主控随后按SPEC §21.1.4定稿N09技术契约并授原Worker必要窄写域；本次接受没有改变N07/N08或真实运行状态。

## 当前CLI交付物限定证据与restore-09

已完整读取当前CLI PE清单及Rust元数据原JSON，Evidence分别为`ev_134bd75b1aba20bc6e6587377854b27f2468f2ba137024a215868bcde78bdbc4`和`ev_46808fa0cc8dba796125e3d035c82acdc40be7d51c7ab41d93e60559878c861a`。二者绑定实际389384字节、SHA256 `612b2669808fe1e53ed7e8afd9e2b5c11ae59b7c2eeb324e2e0086e429332dab`的orca.exe；分别只证明根文件/PE及amd64/64位/little endian等静态事实。原生函数/类型/执行未恢复，未找到.dep-v0不能说明依赖为空，rustc路径hash只是推断。

当前命令目录分析Evidence `ev_c6bf5b3fd2db044912040136dd5195895e3a4bc111b9341396587d2f327a9e7e`绑定目录144e9caf549ab54f0756e2108b3584cdd8a721cde2970777941de7945e285f49，242文件清单、20JS/108537字节/11401 AST节点/解析失败0，222文件未选；graph与semantic均partial/truncated=false。独立`read-rea-fields.mjs`以成熟stream-json读取16937686字节原件，仅组装身份、来源选择、统计及coverage/limitations，精确核Evidence及目标摘要，EXIT0，输出`.verification/rea-cli-js-static-review.json`；没有重跑REA或组装完整关系图。

已读取实际index、Worker启动/观察/停止/释放、消息及请求查询等相关JS和RuntimeClient/transport/兼容包络对应分支，核当前文件先校验参数再加载客户端、实际Agent/model/from与原RPC、transcript/terminal来源、未知停止/释放非零及原请求查询。另完整读取准确b287的Rust启动器源码，它提供桌面Electron与旁置服务配置两种布局参考，不能认证当前PE每个函数。对主控分析输入记录中的CLI及20个选取JS逐项只读核当前长度/内容摘要，21项均相同，没有执行目标脚本或操作本机Orca。

三项实际REA exit均0，原resources及summary资源已读，BelowNormal、间隔约3.002—3.009秒，CPU最高19.4%、可用内存最低约11.369GiB；静态关系、未选取关系与完整动态闭包各保留其限制。prepare-09/stop-09实际原件表明S41264在UTC16:57:54.831364明确退出15后才分析；恢复资源CPU5.6%/12.0%、可用11.327/11.404GiB、间隔3.001976秒，通过当前门限。

`.verification/inspect-restored-owner.ps1 -RestoreName restore-09`于UTC17:05:43实际EXIT0：恢复/current整份JSON深比较相同，实际S47228/父54112、原程序、创建UTC16:59:21.6222126及epoch1792515561.6222126差0毫秒、BelowNormal及唯一127.0.0.1:64505监听符合；数据库记录integrity ok、foreign_keys空、stopped_interval_comparison equal。控制脚本仍为已审c8，无恢复常量修改。该命令只读记录和进程/监听，没有执行SQLite、HTTP、资源采样或控制正文；完整环境/配置、health/Web及原Orca38652/runtime49da保留主控实际来源。下次prepare取实际restore-09/restored.json并重新准入，当前窗口归原S，无新Worker运行授权。

## 准确16ad的独立Go结果与restore-10

结论：**接受准确`16ad5ef6fdbef17b54d8ce04224f0dca3abfafc3`的主控独立四包Go源码编译和三个传输前函数实际结果。** 主控新建detached目录`.verification/swa-go-16ad-01`，使用已审且相对c4无差异的原compile入口；本Reviewer独立核该目录准确HEAD、status条目0、实际SOURCE/tests.json、四个非空测试二进制、完整执行日志、函数及五份资源日志。主控实际session49790 EXIT0，本Reviewer没有重复执行编译或二进制。

实际命令为`pwsh -NoProfile -File tests/swa-cli-headless/compile-go-tests.ps1 -OutputRoot C:/Users/DW/orca/Astrocyte/.verification/swa-go-16ad-01/.verification/swa-cli-headless-20261020/go-source-01`。原四包test-c完成；TestOrcaCLIRejectsAuthorityArgumentsBeforeTransport、TestOrcaCLIRejectsUnknownActionAndIrrelevantInput、TestOrcaCLIMissingCurrentGrantIsDenied逐项真实RUN/PASS、无SKIP。成功后SOURCE完整等于16ad、同一源码仍干净；实际摘要保留runtime tests executed=false、model_calls=0。Go4/-p4/512MiB、私有缓存/TEMP及原依赖下载日志保留；五阶段passed=true、CPU最高18.9%、可用内存最低11586949120字节、单调间隔2.500929—2.502002秒符合当前门限。

prepare-10/stop-10实际原件均通过，S47228于UTC17:11:16.413051明确退出15后才运行检查。restore-10启动资源CPU11.1%/10.8%、可用11.029/11.013GiB、单调3.002088秒、BelowNormal。`.verification/inspect-restored-owner.ps1 -RestoreName restore-10`于UTC17:17:27 EXIT0，恢复/current整份记录相同，实际S58620/父27260、原程序、创建UTC17:15:00.6021515/记录epoch1792516500.6021516相差0.0001毫秒，唯一127.0.0.1:64505监听；原库记录integrity ok、foreign_keys空、停止区间equal。

控制脚本仍为准确已审c8，后继prepare取实际restore-10/restored.json并新鲜准入。完整原环境/配置、health/Web、两次所属执行退出观察及原Orca38652/runtime49da保持主控实际来源；本Reviewer只读原件和进程/监听，没有访问SQLite/HTTP或采样。该结果属于主控独立检查，不能登记Worker自测；N07/N08、N09程序、新输入/恢复、四个runtime/live函数、Node/native/TS完整构建、安装及同候选双Worker均保持未验。

N08另已独立完整读取本机Pi1.0.1的docs/cli.md及实际main/sdk/agent-session/RPC相关消费者：no-tools映射noTools=all，allowedToolNames和初始active为空，实际tool registry刷新按该允许集合过滤；显式tools可覆盖，资源禁用开关仍允许各自显式路径。真实get_state没有tools/activeTools/ToolsOff字段，不能沿旧SDK字段声明原生握手已核工具空集合。认证文件backend的锁/写入与只读发现类分别有消费者，源文件读取不等于运行不刷新。上述来源线索已交原Worker；没有读取本机认证文件内容，也没有启动Pi、会话或模型，N07/N08技术方案及效果仍待准确审核。

准确16ad独立Go结果hook `msg_0ea2de6ec592`已获主控`msg_634d579c87f1`完整读取f367准确全节及实际原件回复，实际唤起确认。该次接受及后继源代码审查分别按自己的准确范围进行。

## 准确33c5af4f的注册批准实现审核

结论：**退回准确`33c5af4f941bb1252aa382623656612bb2e7c3cd`中的新真实检查及状态说明，修正后再审N09。** 已完整读取相对16ad的27文件差异，追踪App/Composition、Go bridge/严格来源、Node注册/owner、原dispatcher/transport及新增真实检查。N09主体实现与已定稿§21.1.4的静态结构相符；检查存在确定的请求准入问题，当前不能据此接受完整功能，全部动态仍未运行。

| 问题 | 实际来源、影响与返修要求 |
| --- | --- |
| N10 新真实检查在读取前后项目时即被拒绝 | TestExternalCLIRegistryRealHumanApproval的readProject使用Agent bearer访问GET /api/v1/local-projects。实际session.scopedAgentRoute在目标项目路径分支之外返回false，relayScopedWorkspaceRoute也未放行该全局列表，因此在进入ListProjects之前返回403；增加discover Grant不能解决。原Worker须复用真实私有SQLite或已有合法项目内只读来源，保留当前设置/身份比较，不为检查开放新管理接口或获取human会话。缺失与错误身份应检查实际明确拒绝原因，不能将任意依赖错误计作权限保护通过 |
| §12状态说明不一致 | 该节首段仍写只提案、未修改新写域或冻结接口，与后面的真实实施及已定稿SPEC21.1.4冲突；同一Worker同步为当前实际状态 |

N10已以`msg_b4d0f6632c93`交原Worker，并通过`msg_fb5df4d5a6ef`通知主控。该问题不影响按原授权对准确源码进行生成或编译准备，也不授予新运行窗口；实际检查修复后仍须用户完成具体候选的真实人工批准。

正向静态范围已核对：同一真实adapter在Composition中一次绑定，只替换非空ExternalModelCLI注册验证，原HistoryRoots/旧S3/工具约束保持；端口错误在SaveProject前返回。Native只取实际registration，strict请求/结果使用原envelope；Go核实际ID、schema、mode/source、同安装VERSION/host.cjs、内外runtime/authority及canonical根，缺失或不匹配失败，不进行完整发现/version probe或模型。新项目校验不消费已有Grant，未配置Orca保留原路径。

owner由serve的crypto/rand生成，Go本地子进程环境传给Node host/client；host在启动其他组件前删除继承变量并在内存持有。真实RuntimeClient/transport包络传递该凭据，远程目标拒绝携带；PTY和structured子环境删除该值。新增dispatcher callback在legacy handling及mutation executor/完成重放前核SWA请求的本地caller与恒定时间secret比较，其他宿主无callback时保持原行为。已追踪原caller identity及本地request JSON接收字段；它不等于T12完整helper/current-consumer矩阵已经实现，后续workerSend仅有合同/兼容字段准备，实际方法仍按原任务继续。

N07明确模型已送入configuration，Pi复用原parsePiModelOptionId及对应provider认证条目，OAuth格式要求access/refresh/有限expires；没有默认模型时发现仍如实unknown，明确选择可继续核配置。Codex仅对有相应来源的默认openai、Claude对已读账号格式提供当前静态来源；真实账号有效性、实际运行账户、外发目标/原禁令和工具仍按N08核准，ACP等未适配项没有得到可启动结论。该局部避免以缺默认模型直接判无账号，未证明实际模型或来源绑定成功。

依赖方面，准确源除host-entry外，rpc/dispatcher及RuntimeRpcState也静态引用完整ALL_RPC_METHODS。主控已要求按真实消费者收窄并检查metafile/安装；本Reviewer将dispatcher的准确第二入口以`msg_d01dfdd755f1`和`msg_f755847b56ad`交主控/Worker。只过滤注册方法不能证明模块已经从bundle移除，尚未接受完整运行依赖闭包。

`check-revision.ps1 -Previous 16ad5ef6fdbef17b54d8ce04224f0dca3abfafc3 -Candidate 33c5af4f941bb1252aa382623656612bb2e7c3cd`实际EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。生成类型、编译、测试正文、Native/HTTP服务、安装和模型均未由本Reviewer执行；Worker所报语法/格式/架构结果保留其原归属，不能替代本次未通过的真实检查或功能验收。

准确33c5退回hook `msg_a0ef3d3dc524`已获主控`msg_b1076cff80e7`完整读取82d准确新增全节及真实请求准入的回复，实际唤起确认。原Worker按问题提交准确e549后继；技术实现的局部静态范围与整体未验状态分别保持。

## 准确e5492b18的真实检查修正与类型生成

结论：**接受准确`e5492b1898758d041bacadfabc5e0310fb87ae09`的N10检查修正、当前状态说明及已实际生成的类型同步；接受范围为静态代码和生成结果。** 已完整读取相对33c5的六文件差异和修后完整检查，核实际session/middleware、SQLite项目读写及私有服务数据库生产者。N09整体功能、权限真实正文及模型调用保持未验。

修后检查使用成熟modernc SQLite库，以canonical私有DataRoot/workspace/state.sqlite构造file URI和mode=ro，只查询实际local_agent_projects及成熟JSON解码；该路径与runOrcaProduct实际打开路径一致，未为检查增加HTTP路由。合法项目内capabilities要求实际runtime、schema及注册ID，Agent设置写入要求403、schema_version=1及实际scope_denied/明确操作权限消息，前后实际项目深比较相同。无owner的直接Native检查同样要求准确OrcaNativeError/code/message，其他依赖错误不能计作权限通过。未知/无服务/旧行为及真实人工批准等原计划仍须运行，修后正文没有执行。

报告§12与USAGE现已描述已实施的真实注册验证和其适用范围，不再将定稿接口写成提案。真实注册批准仍只核注册来源，账号、工具、可启动及不同执行模式不能据此通过；原HistoryRoots/S3、非空工具限制和人工批准职责保持。

独立读取Worker generate-02的完整实际入口、generate日志、资源/退出文件及所属退出观察：实际日志只记录生成web/src/api/schema.d.ts，两个退出值均0；UTC17:28:02.235392/04.744076、CPU49.6%/55.0%、可用11254743040/11278381056字节、单调2.508680秒，符合当前门限。实际generate分支明确Node512MiB；入口中的其他Stage没有因本次授权运行。生成结果新增三个SWA类型及明确Agent/model字段，与准确33c5已定稿来源同步；UTC17:29:34.2865424/37.8930265两次所属进程为空记录保留。准确33c5干净前置及执行session/授权归属采用原Run实际消息，没有由本Reviewer重跑生成或编译。

`check-revision.ps1 -Previous 33c5af4f941bb1252aa382623656612bb2e7c3cd -Candidate e5492b1898758d041bacadfabc5e0310fb87ae09`实际EXIT0，三Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。完整TypeScript、Go编译、真实测试正文、安装、Native/HTTP调用及模型均未由本Reviewer运行；两个既有构建失败、N07/N08、N04/N05、T12、所属Job和同候选双Worker继续按各自准确状态处理。

## N08独立条件与未定稿草案反馈

对照草案前，独立正确性条件包括：同一本机原生CLI和准确版本/模型；原账号来源只读、私有根canonical/互斥/受保护；不得从文件副本推断服务端账号保全；原生工具和全部资源/参数/环境关闭路径；真实accountHome生产者及原R/Dispatch/session关联；效果前当前Go批准、来源、TTL、consumer/session/fence重核；没有新增凭据、身份或执行账本；未知Agent与原禁令保持；同候选真实双Worker、准确停止A及B原会话后续/完整成果。真实get_state没有工具字段，静态方案和未执行场景分别保留。

已完整读取原Worker互斥忽略目录`n08-native-account-tools-plan.md`，它尚无准确Git版本，**未接受或冻结新增native_launch契约**。正式计划须纳入原已授文档，提供准确提交，经独立审核和主控SPEC定稿后实施。当前反馈如下，全部交同一原Worker，不修改程序。

| 条件 | 草案覆盖及缺口 |
| --- | --- |
| 真实账号及原账号保全 | 草案正确识别普通RPC使用可写FileAuthStorageBackend、auth check才有noRefresh只读分支，但第5项仍允许复制原refresh并实际刷新。当前安装的pi-ai/auth/oauth/openai-codex.js明确向token端点提交grant_type=refresh_token，并接收新的access/refresh/expiry；静态副本不能证明服务端原凭据保持有效。需要真实原生可成立的无刷新路径及过期边界，或由主控登记必要授权待定；禁止伪造refresh/协议。该项已以msg_cb77604f68c5交原Worker、msg_fe8d26f9d24a交主控 |
| 私有根及并发 | canonical/ACL、明确record.accountHome和原R/checkpoint生产者方向覆盖必要条件；同项目共享副本还须明确来源变更、原生并发锁/生命周期、旧Worker未释放及已有未知副本的拒绝条件，不能覆盖原生刷新结果或暗中替换账号 |
| 原生工具和资源 | 固定no-tools及资源禁用、拒绝冲突extraArgs和秘密环境、真实握手不造工具字段方向成立；须补准确offline、retry/compaction默认消费者和完整允许参数/环境集合，防止其他实际消费者改换资源或外发行为。版本字符串和缺少工具事件都不能单独证明关闭 |
| 效果前准入 | 草案列出Go当前批准、异步serialized handOver后重核、原selection/current consumer/session/fence；须把来源/凭据有效期及安全余量明确到首次和每次真实输入前，unknown后仍保全原R/operation/Job，不重选或重放 |
| 真实验证 | 同本机CLI、真实provider/model/session及自产JSONL、实际文件/命令效果检查和双Worker覆盖必要方向；不得通过修改原配置、假凭据/时钟/CLI制造失败。未命中与无可用安全原生路径如实记录NOT_RUN/INCONCLUSIVE，Pi局部结果不扩展成全部Agent通过 |

上述条件来自当前需求和实际原生消费者，没有要求重复编译、另造协议或执行自证用例。当前只读反馈未运行Pi、认证、会话或模型，没有读取本机认证文件内容；通用工具、T11及原Codex/S3的各自限制保持。

## restore-11失败、限定诊断与同源恢复

首次restore-11实际PID43188在监听前退出，原stderr只记录CodeGraph SDK exports命令exit status 1，目录没有restored.json。准确ed543原config源码使用15秒ContextTimeout，错误消息不能单独说明原因；本次失败保留。主控通报诊断01因Windows环境键大小写KeyError在目标启动前失败，目录为空；该执行错误沿主控原输出保留，没有写成目标失败。

独立完整读取诊断02脚本、launch/result/resources及空stdout/stderr：实际相同Node/SDK导出命令，私有原cwd、显式环境、禁下载/遥测、BelowNormal和Node512MiB，一次45秒诊断，实际PID14344、8.470091秒、exit0/timeout=false。资源CPU52.0%/45.6%、可用约10.498/10.387GiB、间隔3.012440秒。诊断未改变原服务15秒、原依赖或配置，也不能据后继成功认定首次失败原因。

新restore-11-recovery-01使用同prepare-11/stop-11及原ed543，原件launch/start-intent/source/database/resources/restored已读。资源CPU45.0%/52.0%、可用约10.473/10.495GiB、单调3.012373秒；原库记录integrity ok、foreign_keys空、停止区间equal。`.verification/inspect-restored-owner.ps1 -RestoreName restore-11-recovery-01`于UTC17:45:08实际EXIT0：current/restored整份记录深比较相同，真实S22036/父23492、创建UTC17:38:10.2284094/epoch1792517890.2284093相差0.0002毫秒，原程序、BelowNormal和唯一127.0.0.1:64505监听符合。

控制脚本仍为已审c8fdf3188ea97f1334b22bd261e578db2bf50a2f5039ddf944613753b42d1176，下一prepare取实际restore-11-recovery-01/restored.json并重新准入。完整原环境/health/Web、两次旧所属进程退出及原Orca38652/runtime49da按主控实际来源保留。本Reviewer只读原件及进程/监听，没有执行诊断、控制、SQLite、HTTP、服务或模型；当前窗口归原S，没有新Worker重操作授权。

## 准确e54d4548的构建顺序调整

结论：**接受准确`e54d45489f54819cccb2d408474e6f885aac6970`相对e549的构建顺序静态调整及两份对应记录。** 唯一程序差异把同一个await buildSWARuntime(runtime)移至prod deploy和准确干净源码检查后、固定Node下载/native编译前。已完整读取build-swa-package及build-swa-runtime实际消费者：esbuild来自source已安装依赖、输出使用已部署package.json、native项external，当前调用没有待下载Node/已编译ABI的运行前置。没有改变后续Node pin/校验/超时、native/Go构建或最终VERSION写入条件。

前移可在后续窗口先获得实际bundle/metafile或明确导入失败，尚未执行构建，也没有证明依赖已收窄或全部动态资源可用。真实dispatcher/RuntimeRpcState的ALL_RPC_METHODS及其他闭包问题保持，不能以成功生成metafile替代导入、native/dynamic/assets和安装验证。忽略目录next-window入口的候选02输出与新目录规则一致，build03仍须主控另行交付一次窗口，旧候选和失败保留。

`check-revision.ps1 -Previous e5492b1898758d041bacadfabc5e0310fb87ae09 -Candidate e54d45489f54819cccb2d408474e6f885aac6970`实际EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。本Reviewer没有运行构建、安装、服务或模型。原Worker`msg_115a6219f739`已收到N08账号问题并继续形成无refresh权的正式计划；修后忽略文件没有被本次接受，完整原生可用路径和新增契约仍待准确提交及主控定稿。
