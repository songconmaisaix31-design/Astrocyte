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

## 准确76e9dafd的授权标识、安装入口与SDK检查源码

结论：**接受准确`76e9dafd12f4ff430a3082ff4a65ec543d1e02f8`的相应静态修复及检查源码范围**。已完整读取相对f295的全部七路径差异；准确76e9是d7的父提交。完整成果方案与实现分别沿上一节和后继处理，整体产品、真实SDK与安装结果继续证据不足。

Runtime.Target改为现有注册标识pi。已核同一准确代码：SetProjectSettings通过NativeRegistry.Adapter校验ExternalModelCLI，Registry实际注册pi，ValidateProjectActionScope随后要求target与批准值相同。原绝对PiEntry路径无法通过该设置准入，新标识使已有真实批准路径一致；它没有证明Pi已安装/可用，也没有签发Grant。执行仍使用配置的同候选Node/SDK绝对路径及固定provider/model，未从PATH挑选其他执行目标，原human/Agent权限规则保持。

构建复制真实install-swa.ps1和USAGE.md到候选根，闭包检查要求这些交付文件存在；CLI help列出实际必需的auth/protected参数及stdin生命周期。USAGE说明新目录、保护根、既有human项目批准/token、pi标识、固定输入和原生request查询、停止及释放，不把服务启动当授权签发。源码形成了随候选安装和调用入口，实际构建、安装、私有HTTP项目授权和命令可用性尚未验证。

新增SDK检查使用同候选Node、真实只读OAuth适配、真实受限SDK和实际Job；代码核准备会话的tools/cache/retry/compaction/认证操作、非法写回调不执行、真实get_state，以及仅允许get_state/abort时禁止prompt。已核真实JsonlRpcPeer成功响应解出data、Pi1.0.1 get_state字段与断言对应。该源码明确isStreaming=false、messageCount=0、model_calls=0，其abort为**空闲SDK的真实RPC关闭检查**，不能证明执行中模型取消、正向模型调用或双Worker停止；T13实际A/B仍须单独通过。

verify入口在普通进程fixture后重新资源准入并调用该SDK检查，保留两个命令的实际退出及剩余分支；没有在源码阶段执行这些命令。`.verification/check-revision.ps1 -Previous f295f7c9360a31e6c1ac4ade2ab9c090c455ddae -Candidate 76e9dafd12f4ff430a3082ff4a65ec543d1e02f8` EXIT0，七路径Git空白及USAGE成熟Markdown解析通过，无本地链接/锚点/HTML；marked18.1.0、BelowNormal、Node512MiB。当前无正式PR或动态窗口，所有测试、构建、安装、模型及服务仍NOT_RUN。
