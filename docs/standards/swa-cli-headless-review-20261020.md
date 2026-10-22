# 无界面 Swarm CLI 独立审核

## 范围与准确需求

本报告接续首版CLI独立审核；早期Reviewer身份为`task_1cd8014750f9 / ctx_980a3282c339`，当前为`task_9ae5a1337e1d / ctx_5f4446e340a0`，原Run为`run_62ada397a104`。当前写入范围为本文件、既有swa-mainline-review报告及自有忽略目录，业务实现由原Worker负责。实际客户端为Orca/Codex，实际会话projection为`gpt-6.1-sol`、终端预览为`GPT-6.1-Sol xhigh fast`；没有子Agent。

H01—H16先依据准确main `81812c32a6d345362655c6e569f680aab8fca877`的首版任务和QUESTIONS形成，并在`846d67ecd3684b2c6e3b4922fe4714056461371d`提交。该阶段接续准确main `e149510762f89bed06e7c8c766917554a58294e4`的首版任务及QUESTIONS，按不妨碍本机Orca并集成产品的要求更新H02/H13/H16、追加H17，均在读取对应新Worker计划前完成。本轮采用现行SPEC §21.1.5—§21.1.6、F18C/F803及用户最新授权；具体准确版本、已验范围和真实限制随文末审核记录接续，旧阶段事实不替代当前结论。

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

e549、e54d和main72f三个hook分别为`msg_f1a9fe13d2e4`、`msg_0e985fb2b815`、`msg_393ef3c9a2d5`，已获主控`msg_d21b77c8148c`、`msg_71fc61e466cc`、`msg_1db6699ae7e7`完整读取准确c5f3两报告新增及实际差异的逐项回复，实际唤起确认。回复前本地参数拒绝的尝试没有计为成功回执，后继动态结果独立记录。

## 准确e54d的build-03失败及restore-12

结论：**退回准确`e54d45489f54819cccb2d408474e6f885aac6970`的可构建交付范围，真实缺失依赖由同一原Worker修复；既有构建顺序静态接受保持。** 已完整读取build-03的13545字节日志、资源/退出、实际运行和两次退出观察，核准确上游导入及pin、当前runner摘要和部分候选目录；没有重新执行构建。

原Run只交付一次`next-window.ps1 -Stage build -EvidenceName build-03`，实际exec76466退出1，资源退出0。runner SHA256为77d93dab38c8517c4280e03b4396ef21acb538ccf5eb2c9a890d9ca3c49352e4，与交付记录一致。UTC17:54:01.542113/04.043874的CPU5.3%/8.4%、可用10962219008/10999603200字节、单调2.501772秒符合准入；原Node1024MiB、Go4、私有缓存/TEMP及BelowNormal按实际入口保持。pnpm原命令内建ETIMEDOUT重试和慢下载警告保留，最终prod deploy完成后才进入esbuild。

首host bundle真实错误为headless-emulator.ts的@xterm/addon-unicode11和acp-tool-timeline.ts的diff未解析；准确e54d上游package分别pin为0.10.0-beta.301、8.0.4，源文件实际import存在。首bundle失败即终止后续流程，候选02只有runtime及部署文件，没有host/metafile、Node目录、native/Go产物或VERSION，不能安装。需要按实际消费者补必要声明或收窄真实闭包，不能仅external后留下未安装模块。host-entry、dispatcher和RuntimeRpcState的完整ALL_RPC_METHODS及旧SDK消费者仍待处理，尚无完整闭包结论。

实际运行观察包含所属pwsh、conhost及Node；UTC18:00:05.6984119/08.0245030两次所属进程为空记录来自Worker原件，主控查询范围较窄，分别保留覆盖。prepare-12/stop-12完整原件及一致备份记录已读，原S22036于UTC17:52:04.334684明确退出15；prepare资源CPU9.7%/12.9%、可用10.284/10.386GiB、间隔3.002321秒，stop资源CPU7.8%/7.7%、可用10.387/10.352GiB、间隔3.009414秒。原备份integrity ok/foreign_keys空，后继恢复使用原库。

restore-12实际资源CPU16.0%/8.9%、可用10.413/10.462GiB、单调3.002029秒；launch/start-intent/source/database/restored及空stderr已核，原库记录ok/fk空/equal。`.verification/inspect-restored-owner.ps1 -RestoreName restore-12`于UTC18:02:27实际EXIT0：current/restored整份JSON一致，实际S56896/父52616、原ed543程序、创建UTC18:00:13.3731609/epoch1792519213.3731608相差0.0001毫秒、BelowNormal及唯一127.0.0.1:64505监听符合。完整原环境/配置/health/Web及原Orca38652/runtime49da沿主控实际来源保留；本Reviewer未操作SQLite/HTTP或目标进程。

下一prepare取restore-12/restored.json，c8源不改，当前窗口归原S；build03授权已消费，不授新重操作。两个旧构建失败及部分candidate01/02、restore11失败和诊断限制继续保留，全部真实Agent、输入/恢复、Job/完整成果、安装及同候选双Worker保持未通过。主控`msg_4a9851052cd9`只读默认Pi auth条目0及指定父环境未设置，仅限定该来源/时点；其他profile/Agent未知，本Reviewer没有读取认证内容，N08正式契约仍未定稿。

build03退回hook `msg_7627dfe9b893`与main6883管理接受hook `msg_d2bdb411ad3c`已分别获主控`msg_43118ea4073a`、`msg_6dd65c489f89`完整读取de4准确新增全节、真实导入/原件及管理差异的回复，实际唤起确认。

## 准确8ebefec7的必要依赖与私有store修正

结论：**局部静态接受准确`8ebefec700781e9883a3bb3f905233e07f03e671`的两项必要依赖声明和构建私有store复用，N08技术计划另行审核。** 已完整读取相对e54d的五文件差异、§13正式计划及对应实际消费者。runtime/package与dependencies.lock.json同步声明Unicode11=0.10.0-beta.301、diff=8.0.4，与实际缺失导入及上游devDependencies准确pin一致，没有添加external绕过。pnpm-lock未手写，真实生成/安装仍待单独窗口。

构建store改为同任务已有`.verification/swa-cli-headless-20261020/pnpm-store`，mkdir后核最终lstat和整条realpath与实际仓库根下指定路径一致，重定向失败；仍由完整pnpm prod deploy建立新候选，candidate/temp独占创建，原pin/校验/后续VERSION条件保持。只读实际已装pnpm11.27.0的bin shim和dist/pnpm.mjs：verify-store-integrity为Boolean选项、普通配置默认true，store消费者传递该值；构建显式启用，没有复制node_modules或改变全局store。本Reviewer没有执行pnpm或重操作。

`check-revision.ps1 -Previous e54d45489f54819cccb2d408474e6f885aac6970 -Candidate 8ebefec700781e9883a3bb3f905233e07f03e671`实际EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。成熟解析额外识别§13未包代码的`Promise<AgentSessionAccountHome>`为HTML，该计划显示问题交同一Worker修正；它不改变两项构建修复的局部接受。N08尚未技术接受或冻结，不以本次决定授权实施新增native_launch；依赖安装、bundle、完整闭包、native/Go/TS、安装和全部真实调用继续未验。

准确8ebe必要构建修复hook `msg_dc1694b21392`已获主控`msg_c98acfafff42`完整读取891准确新增、五文件差异及实际消费者的回复，实际唤起确认。主控随后仅交付dependencies/dependency03，其他tracked/Git冻结；本次N08审查不要求Worker在窗口内修改冻结文件。

## 准确8ebefec7的N08正式技术计划审核

结论：**退回准确`8ebefec700781e9883a3bb3f905233e07f03e671`的§13技术计划，补齐原生配置准入及有效期语义后再审。** 独立条件已在本报告“N08独立条件”列出；本轮完整读正式§13后，另核本机Pi1.0.1实际args/main/settings/sdk、auth-storage/runtime-credentials、pi-ai/auth/resolve/openai-codex provider、provider-composer/resolve-config-value和RPC分支。没有读取认证内容、运行CLI、制造配置或调用模型，必要构建修复的局部接受保持。

| 条件 | 已覆盖范围及必须补充内容 |
| --- | --- |
| 原账号保全及真实认证 | 计划不交OAuth/refresh、不改变凭据类型，只允许原生provider实际支持的既有literal API key；默认Pi无条目限定准确。实际auth/resolve对stored类型和provider handler成对匹配，openai-codex只声明OAuth，与限制相符。私有API key格式和auth check ready仍不代表远程模型有效，当前没有可执行账号结论 |
| N11 必要models和auth条目的原生投影不完整 | §13.3只写必要models声明、拒绝命令/未证环境形式认证，没有具体规定provider/model headers、models apiKey/oauth、modelOverrides及credential.env的逐类准入。实际provider-composer会在已有stored credential时继续解析provider headers，并对模型headers调用resolveHeadersOrThrow；resolve-config-value支持以!执行shell和环境插值，ReadOnlyAuthStorage也保留credential.env。空extraArgs和no-tools不能封闭这些消费者。须明确成熟结构解析后的允许/拒绝字段和原生literal判定、明确外发目标绑定、私有投影及每次输入前复核，并增加上述真实失败路径的计划；不能把任意原models块复制后交原生自行验证 |
| N12 无TTL API key语义有歧义 | §13.5先写原生API key无TTL时保留unknown，后句又写未知或到期拒绝，没有说明后句适用于有期限的来源还是全部key。须明确无TTL来源的准入、当前批准和实际远程验证边界，有明确期限时的执行时限/安全余量以及过期拒绝。不能伪造有效期，也不能以原生格式没有TTL字段自动判定全部既有API key不能调用 |
| 原生参数和设置 | 固定完整argv、空extraArgs、环境白名单、offline/no-approve和禁资源覆盖必要方向。实际args支持这些开关；no-approve映射projectTrusted=false，project scope返回空设置；私有compaction/retry/provider retry/cache warming的字段和sdk消费者相符，真实回执及autoCompactionEnabled可核。最终实现仍须完整比较实际part和私有文件，不能只看版本字符串 |
| 原生状态与原身份 | 实际get_state提供model/sessionFile/sessionId和autoCompactionEnabled，没有fence或工具字段。§13.5应分别由真实Pi状态证明模型/会话、由原Native R/Dispatch记录及current consumer证明fence/归属，避免要求不存在的原生字段。首次和后续效果前当前批准、来源、停止/释放、Job、异步等待后重核及unknown保全均已覆盖 |
| 私有账号、并发与测试 | 私有canonical根/ACL、原accountHome生产者、原R/checkpoint、同项目同来源不可变复用和未知副本拒绝覆盖必要条件。六项真实计划涵盖明确模型、人工批准、真实产出和工具效果、撤权/来源/ACL及同候选双Worker，没有要求伪造协议或无关自证；还须补N11逐类消费者边界和N12明确有效期场景，不能把未执行写成失败或通过 |

§13的`Promise<AgentSessionAccountHome>`需要代码格式以正确显示。上述缺口已以`msg_13289c6f4c9f`交原Worker，修正由同一Worker在当前依赖窗口归还且冻结解除后提交准确版本。新增native_launch与accountHome签名本轮不冻结；主控SPEC定稿、全部直接消费者实施及真实运行仍按原顺序，工具/T11和原Codex/S3限制保持。

准确8ebe N08退回hook `msg_d3cda884fcdb`已获主控`msg_f9c663e6cf86`完整读取67b准确新增及实际消费者依据的回复，实际唤起确认。纯dependencies窗口独立继续，N08不得先行实施。

## dependency-03实际子阶段结果与restore-13

已完整读取准确8ebe一次授权下dependency-03的lock/install日志、各自资源和四个退出文件：实际pnpm11.27.0的lock-only为36.4秒、frozen-install为4.8秒，两个执行及两个资源退出均0。原peer及弃用警告保留，生成锁的准确源码审查待原Worker提交；没有执行generate/build/测试或模型，不能由依赖安装推出可安装候选。

lock资源UTC18:14:44.580157/47.086692，CPU62.0%/46.9%、可用9000308736/9069203456字节、单调2.506547秒；install资源UTC18:15:25.575968/28.077254，CPU33.4%/21.0%、可用8904908800/8882749440字节、间隔2.501296秒，符合当前门限。UTC18:15:35.9239104的空running快照已经在阶段结束后，不能证明运行期完整子树；UTC18:18:16.6197410/19.0326939两次root60728及所属为空另按实际记录保留。外层exec27025最终回执未取回、后继poll为Unknown process id沿主控/Worker真实来源保留，**未取得外层EXIT0**，不因四个子阶段退出值或进程为空补造该回执。

prepare-13/stop-13及一致备份原件已读，原库integrity ok/foreign_keys空；原S56896旧epoch1792519213.3731608于UTC18:12:35.908098明确退出15。prepare资源CPU30.4%/16.0%、可用8.283/8.382GiB、单调3.002656秒；stop资源CPU26.4%/19.1%、可用8.503/8.613GiB、单调3.003143秒。restore-13使用同一ed543、原库及该次prepare/stop，资源CPU15.8%/19.8%、可用8.231/8.440GiB、间隔3.002350秒，数据库保全记录ok/fk空/equal。

`.verification/inspect-restored-owner.ps1 -RestoreName restore-13`于UTC18:20:09实际EXIT0：current/restored整份记录一致，实际S56896/父9948、原程序、创建UTC18:17:05.6516341/epoch1792520225.651634相差0.0002毫秒、BelowNormal及唯一127.0.0.1:64505监听符合。该PID数值与旧者相同，创建身份不同；下一prepare取restore-13/restored.json，c8不变。完整原环境/配置/health/Web及原Orca38652/runtime49da沿主控实际来源保留，本Reviewer没有执行SQLite/HTTP、资源采样、控制或服务。当前窗口归原S，新build/模型未授，全部原生执行及安装双Worker仍未通过。

## 准确449184b8的真实锁与后继构建准备

结论：**局部接受准确`449184b826ae64de99d73476147d64967ef31667`的真实pnpm锁及后继构建静态准备，N08修后技术契约另审。** 已完整读取相对8ebe的三文件差异和修后计划，程序保持已审8ebe。锁差异仅含两项必要importer及Unicode11的resolution/snapshot；`.verification/review-generated-lock.mjs`以成熟yaml2.9.1/JSON解析准确Git对象，核全部runtime声明与importer相符、两依赖的实际resolution/snapshot存在，Node512MiB/BelowNormal实际EXIT0，没有执行pnpm。真实lock/install来源及外层最终回执缺失继续按上一节保留。

原build03 runner已保存在该次next-window-executed.ps1，SHA77d93dab38c8517c4280e03b4396ef21acb538ccf5eb2c9a890d9ca3c49352e4；后继runner实际SHA7755b770b1c1a7b25393aa21da166f01c6738513e73ac3a7d8cc65ddb704be37。独立全文比较仅candidate02改candidate03，PowerShell Parser.ParseFile语法错误0、body_executed=false；新candidate03和build-04当前均不存在。原pin/完整性/Stage/资源条件保持，下一正式一次build04须主控独立交付，不重复dependencies/generate或修改旧失败原件。

`check-revision.ps1 -Previous 8ebefec700781e9883a3bb3f905233e07f03e671 -Candidate 449184b826ae64de99d73476147d64967ef31667`实际EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB；原Promise泛型的HTML问题已消除。新增model_target及N11/N12投影/期限/分证计划尚未在本次局部结论中技术接受或冻结，没有真实认证/工具/模型结果，完整闭包和安装双Worker继续未通过。

准确449锁/入口局部接受hook `msg_de865a336ab4`已获主控`msg_1d9c6ba948d7`完整读取70f准确新增和三文件差异的回复，实际唤起确认。下一build窗口仍由主控控制，N08技术决定独立继续。

## 准确449184b8的N08返修及实际请求覆盖

结论：**退回准确`449184b826ae64de99d73476147d64967ef31667`的N08技术契约，补齐N13实际请求参数准入后再审；真实锁与构建入口局部接受保持。** 已完整读修后§13和三文件差异，并独立继续追踪本机Pi1.0.1的ModelConfig、provider-composer、ModelRuntime及pi-ai真实请求生成分支。没有执行模型、伪造请求或修改实现，新稳定字段未冻结。

N11已明确auth仅原生literal type/key、拒绝非空credential.env及models apiKey/oauth，三层headers逐项命令/环境/认证覆盖拒绝；使用同版本原生分类和转义，成熟结构解析、只投影准确provider/model及按原顺序合成。新增model_target绑定实际model/API/baseUrl，每次输入前重新解析来源、读取私有auth/models/settings比较及核当前批准。N12已明确无TTL只记录unknown、在其余批准条件满足后真实请求，有明确期限时每次按真实时间/本次批准上限加30秒判断；没有伪造有效期或将全部API key拒绝。Pi真实模型/会话与原Native fence分别取证、泛型显示及新增真实场景均已修正，与对应独立条件相符。

**N13：计划仍直接保留model和modelOverrides的samplingParams，实际请求可覆盖已核目标和工具。** 原生ModelConfig将该字段定义为Record<string, unknown>，provider-composer的modelFromJson保留原字段，applyModelOverride合并它。真实pi-ai的api/openai-responses.js第286行、api/openai-completions.js第748行和api/azure-openai-responses.js第259行在最终请求参数生成末尾执行`Object.assign(params, model.samplingParams, options?.samplingParams)`，原源码明确后置自定义字段覆盖已命名字段。

因此该路径支持覆盖实际model、input/messages、tools及其他请求字段；get_state中的model_target一致和原生no-tools空工具集合不能封闭最终参数覆盖。此结论来自实际消费者，未声称本机模型配置具有此类值或已经发生调用。技术计划须明确按真实API消费者准入/拒绝及最终请求条件，覆盖来源model、override、合成内置model和每请求options，不能只核表面selection，也不能删除原字段改变来源后继续。补充首次与后续真实输入的对应失败/保全场景，未命中仍按INCONCLUSIVE记录。

N13已以`msg_0b85f81894d2`交主控、`msg_580bfac00d99`交原Worker；修正由同一Worker在当前build冻结解除后提交准确计划。原无refresh、私有ACL/accountHome/原R及并发、固定argv/env/原生设置、效果前批准、unknown保全和同候选双Worker条件保持。必要构建不依赖本次技术接受，本Reviewer没有增加运行阶段；实际原生账号/工具/调用和安装双Worker继续未通过。

准确449 N08退回hook `msg_551561658c55`已获主控`msg_6e93b9e548ee`完整读取b009准确新增及修后§13的回复，实际唤起确认。N11/N12覆盖保留，N13交同Worker，不冻结新字段，也不作为已定稿输入/恢复/Job/成果代码的停止条件。

## 准确449的实际bundle、TLS失败及restore-14

结论：**准确`449184b826ae64de99d73476147d64967ef31667`的完整构建、安装和动态运行证据不足；实际bundle静态结果已确认，候选03不能安装。** 已完整读取build-04的7802字节日志、内外退出文件、资源、运行及两次退出观察，核实际21组bundle/metafile和部分输出目录，没有执行候选或重新构建。原一次授权与准确source/grant/根8800身份在实际命令记录中一致，实际exec86069同步退出1沿Worker/主控来源保留，build-exit和outer-exit均1。

build资源UTC18:28:32.431690/34.933581，CPU22.1%/19.5%、可用9068683264/9124237312字节、单调2.501891秒，通过当前门限；原Node1024MiB/Go4/私有TEMP和BelowNormal保持。完整prod deploy及供应链核查后21输出已生成，随后首个固定Node归档fetch在TLS建立前ECONNRESET/fetch failed；没有据该日志确定具体网络根因或重试。candidate03保留runtime和空node，native/Go/VERSION/swa.exe未完成；原慢请求及模块类型警告保留，四次构建失败和全部部分候选继续保存。

`.verification/inspect-build-metafiles.mjs`使用成熟JSON逐项读取21份原metafile并比较每一实际输出尺寸，BelowNormal/Node512MiB实际EXIT0，当前结果保存在`.verification/build04-metafiles-review-02.json`。实际host为22483367字节，扫描6136输入，其中6120个在输出中bytesInOutput大于0。输出明确包含runtime/rpc/methods下browser-core/browser-profile/Linear/Jira/SSH/mobile-web-bundle/updater等代码；这是实际静态输出来源，不能当作执行结果，也不能仅按名称删除原生必需消费者。

已完整读Worker标准JSON分析脚本及实际来源/限制/五个registry importer路径，再以真实host输入核其存在。除host-entry、dispatcher及RuntimeRpcState，图中还有SSH legacy-relay-cli-server和legacy-relay-cli-call-limits的两个实际全方法表导入；准确源码直接import已核，最短静态路径经headless PTY的kill消费者到legacy-relay registry。完整方法过滤、旧SDK和native/dynamic/assets仍须同Worker按实际消费者处理，本轮没有接受运行依赖闭包；静态输出或21metafile存在不等于全部原生ABI和子入口可用。

运行记录含根8800及conhost36768、pwsh40088、Node61684/62384，均保留实际创建身份。UTC18:31:09.3250354/12.6708658两次所属及此前身份为空记录已读。prepare-14/stop-14原件和一致备份记录通过，原库integrity ok/foreign_keys空；原S56896/epoch1792520225.651634于UTC18:27:27.516680明确退出15。prepare资源CPU19.4%/14.9%、可用8.501/8.496GiB、间隔3.001809秒，stop资源CPU54.5%/52.8%、可用8.583/8.602GiB、间隔3.010869秒。

restore-14使用同一ed543、原库及该次prepare/stop，资源CPU45.5%/48.4%、可用8.697/8.688GiB、单调3.011225秒。`.verification/inspect-restored-owner.ps1 -RestoreName restore-14`于UTC18:35:18实际EXIT0：current/restored整份JSON相同，实际S55148/父61516、原程序、创建UTC18:33:26.8747532/epoch1792521206.8747532相差0毫秒、BelowNormal及唯一127.0.0.1:64505监听符合；原库记录ok/fk空/equal。完整原环境/配置/health/Web及原Orca38652/runtime49da沿主控实际来源，本Reviewer没有访问SQLite/HTTP、采样或运行服务。

后继prepare取restore-14/restored.json并新鲜准入，c8不改，全部已消费授权收回，当前窗口归原S；没有新Stage/模型授权。原Worker继续真实闭包、已定稿输入/恢复/Job/成果及N13返修；账号、原生工具、完整成果、安装和同候选双Worker保持未通过。

准确449构建证据不足hook `msg_469085a11abd`与main646接受hook `msg_0f72418da2c6`已分别获主控`msg_5e31833c4e22`、`msg_e72a4374d103`完整读取9ff两报告新增、实际输出及准确管理差异的回复，实际唤起确认。

## 准确9e46579d的RPC静态修正与两份技术计划

结论：**局部静态接受准确`9e46579df4a073d1d67fc74c198274e0178e4147`的五条RPC导入边修正，技术接受§13修后N08限定契约及§14下载计划；实现、效果和完整闭包仍须后继验证。** 完整读取相对449的十文件差异、全部修后计划，追踪原声明分组、现有构造调用/继承、SSH调度及回执消费者，以及实际Pi请求生产链。没有执行类型检查、bundle、网络、服务或模型。

RPC产品host直接使用原RuntimeRpcShutdown，声明16项allowlist并保留方法数量与唯一性检查；公共OrcaRuntimeRpcServer负责原完整默认选择。dispatcher/RuntimeRpcState改为显式methods，原桌面唯一隐式构造已补ALL，既有IPC/bridge及其他生产调用保留原明确方法表；中间继承链没有新增构造器丢失该字段。SSH的server和call-limits共享原五项声明，dispatcher参数、callerScope、原schema/handler、owner admission及receipt规则保持。当前改动消除已定位的五条全表静态入口，尚无后继metafile或真实运行，不能宣布所有浏览器/账户/旧SDK及native/dynamic/assets依赖已经移除或回归通过。

N13计划现要求原model定义、override、内置和合成model的samplingParams仅缺省或空对象，非空整源拒绝，禁止删掉后继续。限制只覆盖已完整核过最终参数生成的openai-responses、openai-completions及azure-openai-responses，其他实际API保持unknown。Pi真实RPC返回完整模型，首次及后续输入前复核模型/API/baseUrl和samplingParams；产品逐命令strict DTO封闭每请求samplingParams、provider/options/onPayload透传。已核本机RPC prompt仅把message/images/streamingBehavior交session.prompt、Agent.createLoopConfig显式字段不含samplingParams、agent-loop到sdk/ModelRuntime的实际链；无扩展、固定私有设置及资源封闭用户请求变换生产者，源码条件无法证明时拒绝。新增真实场景覆盖四类模型来源和每请求透传、首次/后续输入及unknown保全，原N11/N12、无refresh、原R/ACL/并发和同候选双Worker条件保持，没有多余自证用例。

§14在阅读前已有独立条件清单`msg_37c751eb51fc`。计划固定官方HTTPS和原pin三项资产，独立参数数组调用固定绝对pwsh/script，仅私有canonical工作目录/新文件，逐级拒绝重定向并CreateNew流式写入；单一600秒token覆盖SendAsync和正文复制，无应用重试/换源，非2xx与redirect失败，原Node仍做archive/header/lib/exe大小及SHA/version验证，失败不能进入后续native/Go/VERSION。现有系统代理及默认TLS仅作用于子进程，不修改全局代理、证书或账号，代理/凭据及原始网络异常不输出，失败部分文件与安全类别保留。

Firecrawl CLI在本会话不可用，没有安装；已通过只读网页工具完整读取计划的三份Microsoft官方文档。Windows的默认代理会优先读取代理环境，缺省时使用用户代理设置，与本方案的默认选择相符；现有元数据仍不能证明旧TLS失败原因。[DefaultProxy官方说明](https://learn.microsoft.com/en-us/dotnet/api/system.net.http.httpclient.defaultproxy?view=net-10.0)。关闭自动重定向会把3xx交回应用，符合固定URL失败规则。[AllowAutoRedirect官方说明](https://learn.microsoft.com/en-us/dotnet/api/system.net.http.httpclienthandler.allowautoredirect?view=net-10.0)。ResponseHeadersRead的HttpClient时限只覆盖收到头部前，正文必须另传取消条件，计划全流使用同一token与此一致。[HttpCompletionOption官方说明](https://learn.microsoft.com/en-us/dotnet/api/system.net.http.httpcompletionoption?view=net-10.0)。

D01—D05与独立条件相符：真实目录/已有输入在网络前拒绝，单次获准窗口中的三项官方资产及原哈希、当前系统代理/TLS、不伪造HTTP的真实失败/取消保全及完整候选顺序；未命中为INCONCLUSIVE，未执行保持NOT_RUN。下载脚本尚未实施，后继准确CR和实际窗口不可省略。N08所有新字段须主控SPEC定稿后同Worker实施，限定技术接受不代表当前Pi账号、全部Agent或模型可用；其他已定稿功能继续。

`check-revision.ps1 -Previous 449184b826ae64de99d73476147d64967ef31667 -Candidate 9e46579df4a073d1d67fc74c198274e0178e4147`实际EXIT0，两Markdown/5本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。Worker八TS语法/来源检查保持原归属；本Reviewer只读官方网页和准确源码，没有安装工具、执行下载或把静态语法结果记为真实功能通过。

准确9e局部静态/限定技术接受hook `msg_3a0f11e22c89`及main72ee管理hook `msg_d2806d65a148`已获主控`msg_f9f37a5900c6`、`msg_7b039abdede4`完整读取f5dc两报告新增、十文件差异及全部计划的逐项回复，实际唤起确认。主控后继0055已将N08定稿于SPEC21.1.5，技术接受与实际程序/账号未验分别保持。

## 准确2c391e27的Windows下载实施审核

结论：**局部静态接受准确`2c391e2718ceef2f44721b727c7471918fecdb4e`的下载helper、原构建消费者及新候选准备，真实下载与完整构建未验。** 已完整读取相对9e的四文件差异和103行脚本，逐项对照已审§14及D01—D05，原Node所有archive/header/lib SHA和exe大小/SHA/version、原解包/native/Go/VERSION顺序保持，没有以脚本退出0跳过校验。

helper只准原24.21.0的三个官方HTTPS路径、443、无账号/查询/fragment及原文件名；使用成熟Uri/Path，WorkDirectory必须为本仓私有build下既有直接candidate工作目录，目标必须在其内部。逐级真实目录和祖先ReparsePoint拒绝、已有目标拒绝、CreateNew/Write/Share.None及异步流满足已审文件条件。build使用固定安装ProgramFiles/PowerShell/7/pwsh.exe和脚本绝对路径，execFileSync独立参数数组，无shell字符串拼接。

HttpClientHandler仅在子进程读实际DefaultProxy，默认TLS、无Cookies、禁止redirect；非2xx立即进入失败。600秒CancellationToken覆盖SendAsync/ReadAsStreamAsync/CopyToAsync/FlushAsync，client无另一个较短时限，外层调用仍有600秒硬上限。逆序finally清理所有实际IDisposable，失败或清理异常均保留非零退出及部分文件，输出只含阶段/状态/异常类别，不记录原网络异常全文或代理凭据；清理后才输出成功，Node等待退出后读取真实文件并校验。未加入应用重试、换源或全局代理/证书写入。

准确helper以PowerShell Parser.ParseInput、后继runner以Parser.ParseFile均语法错误0；准确MJS经UTF-8 stdin交Node24.16.0 `--check --input-type=module`实际EXIT0，均body_executed=false，BelowNormal/Node512MiB。旧runner7755已保存在build-04原件，后继实际SHA为b7b5289cf7d312dea37320386de356c19156cb0e94b5634c17892f8e8a519237，全文比较只有candidate03改candidate04；build-05和新candidate04当前均不存在。`check-revision.ps1`以9e→2c实际EXIT0，两Markdown/5本地链接及Git空白通过。

当前只是静态代码符合技术计划，实际D01—D05、网络/哈希/ABI/Go/类型/安装结果均未由本Reviewer执行；外部TLS问题尚未证明修复，完整依赖仍待新metafile。新build05须主控另行交付一次窗口，旧失败及原S55148/restore14/c8保全保持，全部真实账号/工具/输入恢复/Job/完整成果和同候选双Worker继续未通过。

准确2c下载静态接受hook `msg_a1adae3ed2c8`及main0055定稿hook `msg_7c48f3fdf795`已分别获主控`msg_3af1a167b16a`、`msg_d1abbccd3082`完整读取d1ba两报告新增、103行helper和四文档规格的回复，实际唤起确认。后继实际结果分别按下载、RPC图及完整构建范围记录。

## 准确2c的build-05实际下载、RPC图及Python失败

结论：**接受准确`2c391e2718ceef2f44721b727c7471918fecdb4e`本次D02三项实际下载/原校验及已定位完整RPC入口的静态收窄结果；完整可构建交付退回Python选择修复，候选04不可安装。** 已完整读取16968字节build-05日志、内外退出/资源/运行及两次退出观察，实际exec63388和build-exit/outer-exit均1，没有重跑任何Stage。D01未运行、D04未命中分支及D05完整候选未通过，不能由本次200推断所有网络条件或旧TLS根因。

本次准确2c、根37144/创建UTC18:56:09.2623720及一次msg_51ec窗口来源一致。资源UTC18:56:13.312041/15.815249，CPU37.5%/29.3%、可用7926308864/7770779648字节、单调2.503214秒，通过当前门限。实际prod deploy与21组bundle/metafile后，三个官方输入HTTP200，zip/headers/node.lib分别37618919/10291794/2988566字节；日志和原未改消费者连续进入实际Node24.21.0/node-gyp12.3.0，证明原三输入SHA、exe大小/SHA/version条件已通过。本次仅说明默认系统代理/TLS政策下的实际请求成功，没有读取或披露具体代理路由。原VoidTaskResult类型文本保留，安全但尚未验证后继抑制输出。

随后首次node-pty配置因node-gyp自动发现未能取得Python而失败，尚未完成native、其他ABI、Go或VERSION。已完整读实际node-gyp12.3.0/lib/find-python.js：当前版本要求>=3.6.0，默认Windows位置只列311/310/39/38；查找失败不证明本机没有3.13。NODE_GYP_FORCE_PYTHON存在时只有一个检查来源，错误后失败，不转其他默认位置。修复交同一Worker，不能通过文件版本元数据替代实际解释器运行。

同一只读JSON/尺寸检查器对candidate04的21组真实输出实际EXIT0，结果在`.verification/candidate04-metafiles-review.json`。相对build04同一esbuild配置/入口，host实际22483367→20957928字节、输入6136→5666、正bytesInOutput输入6120→5650，完整runtime/rpc/methods/index.ts不在新图中；原browser/Linear/Jira/SSH/updater方法来源已退出该host方法图，仍有通用mobile wiring等真实消费者。本次仅接受已定位入口的静态效果，不证明原生调用、完整动态/native/assets闭包或运行回归。

UTC18:59:04.9625541/07.4104430两次所属及此前身份为空已读。prepare-15/stop-15及一致备份通过，原库integrity ok/foreign_keys空；原S55148/epoch1792521206.8747532于UTC18:55:05.264133明确退出15。prepare资源CPU67.4%/57.7%、可用7.060/7.301GiB、间隔3.014431秒，stop资源CPU65.5%/55.2%、可用7.546/7.608GiB、间隔3.011040秒。restore-15使用同一ed543与原库，资源CPU15.3%/17.2%、可用7.368/7.345GiB、间隔3.002247秒；数据库保全记录ok/fk空/equal。

`.verification/inspect-restored-owner.ps1 -RestoreName restore-15`于UTC19:03:14实际EXIT0：current/restored整份JSON一致，实际S68128/父63436、原程序、创建UTC19:00:53.0194234/epoch1792522853.0194235相差0毫秒、BelowNormal及唯一127.0.0.1:64505监听符合。完整原环境/配置/health/Web及原Orca沿主控实际来源，本Reviewer没有执行SQLite/HTTP、解释器预检、构建、下载或模型。下一prepare取restore-15/restored.json，c8不变，当前窗口归原S；五次构建失败及所有部分候选保留，完整原生/成果/安装双Worker仍未通过。

## 准确b75cf8ff的已有Python预检修正

结论：**局部静态接受准确`b75cf8ffeedffd244934dfac078b4926e865eb15`的Python选择/实际预检、VoidTaskResult抑制及新构建准备。** 完整读取相对2c的四文件差异，并与真实node-gyp configure/find-python消费者对照。构建选择本机已有C:/Python313/python.exe的realpath，在prod deploy/下载前以独立参数数组执行短标准库JSON命令；成熟JSON读取sys.executable和三段整数版本，canonical身份不同、不可执行或不满足原>=3.6条件立即失败。仅本构建子进程environment设置force，四项gyp的--python与force同一来源，不修改系统PATH/全局配置或安装Python；当前没有执行该新预检或声明3.13已由它验证。

下载器两处异步返回值使用void抑制，实际请求、token、文件、失败清理及全部原Node校验保持。准确MJS Node `--check --input-type=module`、准确PS Parser.ParseInput及runner ParseFile均EXIT0/语法错误0，body_executed=false，BelowNormal/Node512MiB。旧b7b5 runner已存build-05原件；后继SHA bae62ff4c1eee443f0efafabf87f814f4a47f6177b9e0812a0ed9622cd087522，全文只candidate04改candidate05，新candidate05/build-06当前均不存在。`check-revision.ps1`以2c→b75实际EXIT0，两Markdown/5本地链接及Git空白通过。

本次静态接受不等于Python预检、ABI/类型/完整构建或输出修正实际通过；build06仍须主控新鲜准入和一次明确窗口，原失败不能覆盖或以重试替代准确修复。N08与原已定稿输入/恢复/Job/成果继续由同一Worker实施，全部真实验收保持未通过。

准确b75静态接受hook `msg_e8be86ce5dd4`、准确2c完整构建退回hook `msg_2b1cf6550bf9`已分别获主控`msg_9c02fba1520f`、`msg_94d7206befc5`完整读取4b99/36d两报告新增及准确差异的回复，实际唤起确认。后继实际结果按各自阶段另行记录。

## 准确b75的build-06实际Python通过与native编译失败

结论：**接受准确`b75cf8ffeedffd244934dfac078b4926e865eb15`本次已有Python预检/实际选择、VS2022发现及下载输出修正的实际结果；完整可构建交付退回首项C++编译修复，candidate05不可安装。** 已完整读取build-06全部303行/44386字节日志、内外退出、资源、三次运行观察及两次后继退出原件，未重跑任何Stage。实际exec65865同步EXIT1沿Worker/主控实际来源保留，build-exit/outer-exit均1、资源退出0；三官方输入和原pin/exe/version校验成功不能补作native、Go、VERSION或安装通过。

原一次msg_5b7d83fdfb5a窗口、准确b75/runner bae62和根55820/创建UTC19:18:44.0288700一致。资源UTC19:18:46.515495/49.017364，CPU15.4%/15.8%、可用7777550336/7801974784字节、单调2.5018756秒，通过当前门限。部署前真实JSON预检取得C:/Python313/python.exe 3.13.13，node-gyp12.3.0随后实际使用同一解释器并发现C:/BuildTools/VS2022的17.14.37710.0；已解决build05的自动发现阻塞。324项prod deploy、21份真实metafile、三项官方HTTP200及未改Node校验完成，下载正文37618919/10291794/2988566字节，旧VoidTaskResult文本本次已不再输出。

首node-pty/conpty编译出现swa_provider_job.h的C4819/代码页936及第17行C2059/C2143，后继SwaProviderJob未声明等连锁诊断和原警告完整保留。实际候选头文件与准确b75的Git blob同为`7d9701d7a8939d58ea7043e103135bc02ac2d7f3`，中文注释和struct完整，未发现候选复制截断；以成熟.NET XML读取实际conpty.vcxproj，两个配置的AdditionalOptions没有显式UTF-8字符集参数。以上支持编码配置修复方向，仍不足以宣称已验证根因或全部编译错误将消失。

运行原件包含根55820及conhost43460、pwsh61744、Node67284/63864、esbuild46104和下载器64368等各自实际创建身份。UTC19:20:34.5682573/36.9947290两次新鲜所属及此前身份均空；较早UTC19:20:27空观察单独保留。prepare-16资源CPU14.4%/13.7%、可用7.188/7.255GiB、单调3.002397秒，stop-16资源CPU22.2%/17.3%、可用7.193/7.229GiB、单调3.009458秒。原S68128/epoch1792522853.0194235于UTC19:13:16.473849退出15，准确c8控制及原库一致备份保持。

restore-16实际exec77870 EXIT0沿主控来源，资源CPU18.5%/8.9%、可用7.190/7.201GiB、单调3.001681秒，原库记录ok/fk空/停止区间equal。`.verification/inspect-restored-owner.ps1 -RestoreName restore-16`于UTC19:24:24实际EXIT0：当前/restored整份JSON相同，实际S60656/父66988、原ed543程序、创建UTC19:21:16.7506354/epoch1792524076.7506354相差0.0001毫秒、BelowNormal及唯一127.0.0.1:64505监听符合。完整原环境/配置/health/Web和原Orca38652/runtime49da沿主控实际来源，本Reviewer仅只读核进程/监听和记录，没有访问SQLite/HTTP或运行控制脚本。

已消费授权收回、tracked/Git冻结解除，下一prepare取restore-16/restored.json且c8不改，当前窗口归原S。六次失败、所有部分候选及candidate-IHFNTU保留；同一Worker继续真实编译、N08/输入恢复/Job/完整成果，原账号/工具及安装同候选双Worker验收仍未通过。

## 准确653cfff7的UTF-8构建修复

结论：**局部静态接受准确`653cfff7bc8c23e467a9f233fcb8f332aaae082b`相对b75的native编译参数修复及实际结果记录。** 已完整读取三文件差异，并核实际候选头文件和生成的MSBuild项目。四行代码仅在原四项native构建前，从本构建environment删除大小写形式的CL/_CL_并明确CL=/utf-8，日志只输出固定非秘密参数；没有修改process.env、系统代码页、源注释、Node pin或原编译/退出条件。Go后续沿原environment调用，CL的MSVC含义没有据此扩大为新运行权限。

已通过只读网页工具完整读取两份Microsoft官方文档：无BOM源默认使用当前用户代码页，显式/utf-8同时指定源及执行字符集；这与当前UTF-8源码和936诊断相符。[UTF-8官方说明](https://learn.microsoft.com/en-us/cpp/build/reference/utf-8-set-source-and-executable-character-sets-to-utf-8?view=msvc-170)。CL参数在命令行前处理、_CL_在后处理，清除继承冲突后明确本构建参数符合该消费者规则。[CL环境变量官方说明](https://learn.microsoft.com/en-us/cpp/build/reference/cl-environment-variables?view=msvc-170)。Firecrawl CLI仍不可用，没有安装。实际完整编译尚未运行，生成项目未显式指定字符集也不能证明外部属性表或真实编译器的全部有效参数，后继结果必须以真实日志判定。

`check-revision.ps1 -Previous b75cf8ffeedffd244934dfac078b4926e865eb15 -Candidate 653cfff7bc8c23e467a9f233fcb8f332aaae082b`实际EXIT0，两Markdown/5本地链接及Git空白通过；准确MJS管道交Node `--check --input-type=module`实际EXIT0，未执行模块正文。全部检查BelowNormal、marked18.1.0、Node512MiB。Worker首次普通push被远端拒绝、确认远端仍b75后普通push成功沿其来源保留，本Reviewer没有将首次失败记为成功或重复远端操作。

本次只接受准确静态修复，没有新的构建窗口或后继编译结果，不能宣布C++问题已消失。N08与其他已定稿实现继续，全部真实原生、完整成果、安装及同候选双Worker保持未通过。

准确b75完整退回、653c静态接受及main bc66管理接受的三个hook `msg_1684846da095`、`msg_9564668427e8`、`msg_0815b6efb659`，已分别获主控`msg_dd7f181dafe6`、`msg_3c3a70abe497`、`msg_66ffe80bd8f3`完整读取cfbd两报告新增和准确差异的回复，实际唤起确认。

## 准确3590bf11的原生Pi及结构化输入阶段

结论：**退回准确`3590bf11043a2e95cb00b319d4a44a44c2e5c784`的本阶段源码；局部静态接受后继准确`1b782b49f530a32c255a9a042e3299cb799fe5b5`的N14/N15返修，整体仍须完成N16/N17/N18。** 已读取相对653c的阶段差异、报告§15、SPEC21.1.2/21.1.4/21.1.5，以及原生配置、父子receipt、host/journal、Pi RPC、真实stdio队列、Job与成果的实际消费者。继续同一Worker返修，未提交T12修改不作为准确实现。没有运行产品、制造协议或调用模型。

本阶段确实移除产品SDK会话入口，改由父Job准入后的bootstrap启动已选本机Pi命令，独立记录实际CLI PID/创建身份及Job成员。NativeLaunchPolicy在Go/Node/Zod/JSON Schema/OpenAPI声明同步；Windows安全模块新增owner SID读取，私有账号使用成熟DACL API和原格式投影。严格RPC参数、工具/资源关闭、真实model/API/baseUrl与空samplingParams检查、原host及stdio队列异步准入、原R与R/native绑定保留均有对应源码。源码存在不证明这些功能已运行；HTTP生成类型仍待生成，完整闭包、类型、Node/native ABI、认证、模型及安装双Worker未通过。

**N14，严重程度中：原models不存在时仍制造私有models.json。** 3590的inspectNativePiPolicy始终返回models对象，materialize与verify均无条件处理该文件；与已接受§13.3的来源不存在则私有不存在条件不符。建议动作：将不存在状态贯穿投影、独占写入及每次实际核查，拒绝出现、消失和重定向。准确1b已用null表示不存在、跳过创建并以lstat要求缺失；存在来源仍要求实际投影相同，本项源码修正已确认，真实分支未运行。

**N15，严重程度高：Go等待后继续使用先前读取的来源。** 3590的beforeInput先policySnapshot，再await当前Go批准，之后只核私有文件及原生身份；最终stdio beforeWrite也沿此序列。原auth/models/CLI在Go等待期间变化时，最后写入没有重读该来源版本。建议动作：等待返回后重新核原selection/完整来源，在后续异步私有检查结束后核准确来源及不存在状态。准确1b已在materialize、environment、tools、Job launcher、reserveInput及beforeInput的Go等待后重读policySnapshot，并在效果前同步核canonical路径/文件身份/版本和models缺失；没有在writer回调递归提交本队列RPC。本项源码修正已确认，真实变化场景仍未运行。

**N16，严重程度高：未固定并复核实际Pi会话身份。** rpc-session.ts的start仅返回首次state.sessionFile，checkPolicyState只要求sessionId非空及compaction关闭；后继实际get_state没有比较首次真实sessionId/sessionFile，也没有将这两个实际值与原provider link作持续绑定。已有Native session/fence和Job检查不能补作Pi自身会话相同的证明。建议动作：保存首次真实原生身份，后续实际状态逐项比较并核原关联，失配在效果前拒绝，可能输入后保全原operation及unknown。此项属于已交付N08声称的真实会话身份范围，未将T12恢复列为本阶段完成条件；没有观察到实际错误会话运行。

**N17，严重程度高：host缺少实际workerList定义，会被现有完整性检查拒绝。** 原Worker先报告遗漏，本Reviewer独立核准确源：orchestration.workerList唯一声明为worker-list-method.ts的ORCHESTRATION_WORKER_LIST_METHOD；host只拼接的ORCHESTRATION_WORKER_METHODS包含start/control/stop/release四组，没有该独立定义。allowlist要求它存在，方法集合无法满足原数量/唯一性条件。3590及1b均保持此缺口，当前不能启动该host。建议动作：明确加入原生单独方法，保留现有完整性检查和MESSAGE组的实际Task方法，后继准确提交再审；没有执行host或采用完整方法表替代。

**N18，严重程度高：异步等待结束后缺少真实Agent Grant到期检查。** Go的checkOriginalOrcaGrant在回包前使用真实时间检查原ExpiresAt，Node的grant_expires_at目前仅用于回包值相同的比较。beforeInput随后还会等待来源/私有检查，最后写入前没有按当前实际时间拒绝已经到期的Grant；1b增加的来源复核也不包含该时限条件。建议动作：在最后异步条件完成后的实际效果入口同步核原Agent批准期限，适用时同样覆盖账号复制与启动；不伪造时钟、增加账本或递归RPC。该期限来自真实项目Grant，与原生API key没有TTL而允许其余条件满足后请求的政策分别处理。当前只确认源码边界缺失，未声称实际命中过期场景。

N14/N15已通过`msg_26a78f9cd9bc`、`msg_76bf311f86cc`交主控/原Worker，主控`msg_c9cbeff4ac68`独立核准确代码并回复；其余三项已通过`msg_6978c87325db`、`msg_eb39950544c7`交回。§13.6对应的缺失/出现、等待期间来源变化、原生身份、真实期限及原operation保全检查须保持，未命中的真实分支按INCONCLUSIVE，未执行按NOT_RUN。T13继续要求真实双Worker模型活动重叠、停止A、B同会话后续输入及完整成果；T09复用同一实际原件，没有增加模型调用，普通进程fixture的范围没有扩大。

本Reviewer的两次check-revision分别核653c→3590及3590→1b，实际EXIT0；前者两Markdown/6本地链接，后者一Markdown/1本地链接，Git空白均通过，marked18.1.0、BelowNormal、Node512MiB。§15本地锚点与实际标题一致。Worker报告的39份语法、JSON/YAML、五Go格式、三PS及来源路径检查保持原执行归属；本Reviewer未重跑这些入口。全部实际HTTP准入、私有ACL、原生身份、来源变化、期限、模型及完整功能检查仍未完成，T12严格caller/recover/后台Job和通用TUI继续同Worker实施。

## 准确3590的主控独立Go检查与restore-17

结论：**接受准确3590本次四包Go编译及三个传输前函数的限定实际结果，源码阶段退回保持。** 已只读核SOURCE、四包产物、完整执行日志、三函数逐项PASS/零SKIP、五项资源、运行及两次退出原件；结束HEAD与SOURCE均为完整3590。主控实际exec2414最终EXIT0，未将该结果归为Worker自测。原已审入口及resource-gate相对准确16ad5ef6无差异。实际cwd为`C:/Users/DW/orca/Astrocyte/.verification/swa-go-3590-01`，核心命令为：

```powershell
& 'C:/Program Files/PowerShell/7/pwsh.exe' -NoProfile -File tests/swa-cli-headless/compile-go-tests.ps1 -OutputRoot C:/Users/DW/orca/Astrocyte/.verification/swa-go-3590-01/.verification/swa-cli-headless-20261020/go-source-01
```

外层实际设置BelowNormal并Tee至control/go-3590-execution-01.log后传播LASTEXITCODE。四包为cmd/swa、internal/adapters/orca、internal/workspace/app、internal/adapters/httpapi；实际运行仅TestOrcaCLIRejectsAuthorityArgumentsBeforeTransport、TestOrcaCLIRejectsUnknownActionAndIrrelevantInput、TestOrcaCLIMissingCurrentGrantIsDenied。tests.json列出的四个live/runtime函数没有执行，模型调用0。五项CPU10.1%—21.5%、可用7550619648—7849086976字节、单调2.501329—2.501747秒，passed均true，Go4/-p4/512MiB及私有缓存沿该实际入口保持。

运行记录保留根21508/创建UTC20:15:08.809810及子65804，UTC20:17:29.9163235/46.1954802两次所属为空。prepare-17资源CPU17.5%/18.8%、可用7.043/6.915GiB、单调3.001633秒，stop-17资源CPU10.0%/8.7%、可用7.216/7.311GiB、单调3.002028秒；原S60656/epoch1792524076.7506354于UTC20:14:51.179275退出15。该窗口不冻结Worker树，没有给Reviewer或Worker新增运行权限。

restore-17实际exec92929 EXIT0沿主控来源，资源CPU12.7%/11.8%、可用7.286/7.329GiB、单调3.002327秒，原库记录ok/fk空/equal。`.verification/inspect-restored-owner.ps1 -RestoreName restore-17`于UTC20:20:38实际EXIT0：整份current/restored相同，实际S36744/父58404、原ed543程序、创建UTC20:17:53.6968189/epoch1792527473.6968188相差0.0001毫秒、BelowNormal及唯一127.0.0.1:64505监听符合。完整原环境/配置/health/Web和原Orca38652/runtime49da沿主控实际来源；本Reviewer没有执行Go、SQLite、HTTP、控制脚本、服务或模型。下一prepare取restore-17/restored.json、c8不改，当前窗口归原S，六次失败和全部部分候选保留。

准确3590退回hook `msg_281c2a76979a`、1b局部接受hook `msg_ed987f6379c8`及main7832接受hook `msg_d5dbdcd67a7a`，已分别获主控`msg_ec69c3f1cbfe`、`msg_da18278169cc`、`msg_6d938539cbff`完整读取bc970新增和准确差异的回复，实际唤起确认。

## 准确319272c8的会话、writer及共享类型消费者

结论：**准确`319272c8852d7992923fa235206aec22335e1d42`已修正Pi路径的N16/N17/N18对应源码；共享接口修改退回N19，阶段运行验收继续未通过。** 已完整读取相对1b的差异和报告§15.2，逐层核对返回值的真实消费者，以及实际host图中的Claude直接调用。没有运行类型检查器、host、模型或自然到期场景。

N16现由首次真实get_state固定UUID sessionId及绝对sessionFile，每次实际get_state检查准确相同值；adapter用同一返回文件生成原Pi provider link并调用bindProviderLink关联。异步等待后还比较record的accountHome、options和providerHandleChain，原session/fence/consumer及Job检查保持。本次确认这些源码条件，真实Pi身份变化拒绝和同会话运行仍待验证。

N17显式导入并加入唯一原生ORCHESTRATION_WORKER_LIST_METHOD，保持MESSAGE组Task消费者及原数量/唯一性检查，补齐此前已定位的遗漏；没有重引完整方法表。尚未实际启动host，不能据本次源码改变登记入口运行通过。

N18将异步准入得到的同步finalAdmission沿structured turns、adapter、Pi turns/prompt delivery、connection和peer传入原stdio writer。writer在output.write前调用它，中间没有await；来源、原身份/consumer及实际Grant期限由该函数再次读取，Pi外层同步核Job成员及连接状态。源/身份/期限拒绝在尚未调用output.write时单独包装ProviderWriteAdmissionError；实际write及其callback失败继续走普通错误/unknown，未被并入无效果拒绝。账号材料/launch各阶段也补真实期限判断。API-key无TTL政策保持，不在占用中的writer递归提交RPC，不增加队列或执行账本。

**N19，严重程度高：共享beforeDispatch返回类型扩大，Claude直接消费者仍使用旧声明。** structured-agent-session-adapter.ts:297现在声明回调返回`Promise<void | (() => void)>`。claude-structured-session-adapter.ts:166—167按该共享dispatch类型取得input，直接将input.beforeDispatch传给dispatchClaudeTurn；后者claude-structured-dispatch.ts:119仍只接收返回`Promise<void>`的回调。这两个实际参数声明没有同步，现有parser与esbuild不能检验这条语义类型关系。Claude路径在实际host依赖图中，属于变更接口的直接消费者。建议动作：同Worker同步必要的声明和实际返回值消费，保留同步guard；不能以cast、假模块/声明、丢弃guard或放宽错误代替修正。本Reviewer未取得实际类型检查错误码，也未声称该旧路径发生了模型输入。

ACP/Codex仍有旧的显式beforeDispatch声明，需按真实调用与返回值使用核对；不据名称增加无关UI或历史测试范围。原package提供typecheck:tsc:node及config/tsconfig.node.json，该原配置还包含非产品入口，不能直接把整个范围当成本任务要求；runtime/swa-orca本身也须由实际受检入口覆盖。必要类型检查应覆盖产品源码、修改的共享接口、所有实际直接消费者和真实依赖声明，保持既定严格程度。缺依赖或未执行均如实记录，不能把13份parser检查改称类型检查。此范围已通过`msg_9900739dff10`、`msg_151bc0023353`交主控和原Worker，主控已要求同Worker准备具体真实入口，当前没有新运行授权。

`check-revision.ps1 -Previous 1b782b49f530a32c255a9a042e3299cb799fe5b5 -Candidate 319272c8852d7992923fa235206aec22335e1d42`实际EXIT0，一Markdown/1本地链接及Git空白通过，marked18.1.0、BelowNormal、Node512MiB。原Go owner capability在清除环境变量前已保存至adapter Config，runtime.go再显式传给Node，此准确源码链保持；没有发现本次清除导致传递丢失。N14/N15局部接受保持，N16/N17/N18的实际分支、类型/生成/Node/native构建、所有模型及安装双Worker继续未验；T12/recover/后台Job与通用TUI仍由同一Worker实施。

准确3192退回hook `msg_44446c835081`已获主控`msg_7628e94e93fb`完整读取1db70两报告新增、准确十四文件和共享类型问题的回复，实际唤起确认。后继源码与测试入口分别按以下准确提交审核。

## 准确56ffc3a8的T01类型检查入口与声明计划

结论：**接受准确`56ffc3a86d74ae4b4d84043d061afd4b9605f1d8`的T01具体类型检查技术计划及七项真实开发声明来源；未执行类型检查正文，未取得任何运行窗口。** 完整读取相对3192的三文件差异、实际上游配置和process-host导出，以及Worker的配置解析入口和原件。`pnpm typecheck:swa`实际展开为Node1024MiB、已安装TypeScript6.0.3对runtime/swa-orca/tsconfig.json运行tsc；Node声明为实际25.9.5。配置覆盖全部产品src/**/*.ts和真实传递imports、原build-constants声明，strict/noEmit开启、incremental/skipLibCheck关闭，保留所有诊断并按非零失败。原桌面/历史测试/科研/web的独立roots未加入，真实imports的消费者和声明仍在范围内。

process-host沿真实orca-source入口及子路径映射，两个Windows addon沿实际index.d.ts。通配paths只优先查找产品node_modules；缺失匹配仍可能进入TypeScript普通模块查找，不能据该配置宣称隔绝其他祖先node_modules或证明运行依赖闭包。此准确边界已通过`msg_e8726522f1d9`交原Worker修正文档措辞；后继真实解析、完整诊断及必要依赖来源继续按原要求检查，不增加自研resolver、假模块、any、断言或诊断过滤。

本Reviewer实际运行`node --max-old-space-size=512 .verification/review-type-plan.mjs 56ffc3a86d74ae4b4d84043d061afd4b9605f1d8`，EXIT0；成熟YAML读取上游锁的两个文档，在第二文档唯一实际根importer与同一package.json逐项比较，核得electron43.7.8、@types/proper-lockfile4.1.4、@types/qrcode1.5.6、@types/react19.2.18、@types/react-dom19.2.7、@types/ssh2 1.15.6及@types/ws8.18.1，保留原peer限定。七项真实声明尚未加入或安装；后继按已接受必要依赖规则显式记录devDependencies与清单，取得主控独占dependency窗口后串行锁定/ignore-scripts安装。Electron声明用途不运行其二进制安装脚本，prod deploy保持原规则。

Worker的19个实际root配置解析结果只证明配置可读及入口存在，没有创建Program。本Reviewer同样没有创建Program或运行tsc、生成器、安装、构建或模型。`check-revision.ps1`准确3192→56ffc3实际EXIT0，一Markdown/1本地链接及Git空白通过。入口接受不等于语义类型、HTTP生成声明、完整构建或功能接受；后继真实缺失项按实际来源处理，不能放宽严格程度。

## 准确7a0058fe的N19共享回调消费修复

结论：**局部静态接受准确`7a0058fefb8499ffb37ffebc47e24866f0218ff7`的N19源码修复，整体运行验收未通过。** 完整读取相对56ffc3的九份TS及报告差异，核对共享签名、实际队列、转交位置、拒绝与晚到结果消费者。三条路径均保留回调返回的同步检查函数，未增加其他Agent的账号、模型、工具或执行权限。

Claude的dispatch、connection及SDK queue同步为Promise返回void或同步函数；authorize在实际arm后传回该函数，输入队列等待批准后、iterator交出消息前执行。此时拒绝包装原unwritten类型，dispatch释放对应waiter；已经交出消息后的写入失败仍保留unknown，没有以新包装覆盖该路径。该源码检查止于真实SDK消费边界，没有取得实际SDK或原生消息写入结果。

ACP在等待附件及异步批准后，将同步函数放入原Send队列项；steer等前一prompt/cancel结束后，start在开启本次prompt前执行。拒绝以neverReachedAgent结算本项并继续原队列，未调用本次prompt。实际host的journal pending先于dispatch，admitted分支不追加覆盖结算，原late settlement消费者保留。真实取消、拒绝及并发队列未运行。

Codex把同步函数沿dispatch/start及两次原steer分支传至connection.request的beforeWrite。等待turn打开及原生拒绝后的后续分支仍使用同一函数，实际sendLine前同步执行；只有该函数抛出才包装ProviderWriteAdmissionError并解除该echo。实际sendLine、超时或断开继续原传输错误处理，保留可能已经写入的关联和unknown。没有将原生拒绝后的既有分支改造成新重试能力，真实协议运行未执行。

Worker的n19-static-01/syntax.log及syntax-exit.txt已完整读取，九份成熟TS parser检查为PASS/EXIT0，入口没有执行类型检查或产品正文；此结果保持Worker归属。本Reviewer实际`check-revision.ps1`准确56ffc3→7a0058为EXIT0，一Markdown/1本地链接及Git空白通过；全部只读检查BelowNormal、Node512MiB。N19已定位声明和消费的源码修复确认，不能据此宣称实际类型检查通过。N14—N18的真实分支、类型/生成/Node/native构建、全部模型、安装同候选双Worker仍未验，T12/recover/后台Job与通用TUI由同一Worker继续，当前窗口仍归restore-17的原S36744。

## 准确68bc4d8c的必要开发声明配置

结论：**局部静态接受准确`68bc4d8c5d182575fe6fe679a436033e8cdc6b36`相对7a0058的三文件差异。** 七项已审精确版本仅加入runtime devDependencies及既有依赖清单，原prod dependencies、scripts和其他包参数保持；文档正确记录产品paths优先匹配及TypeScript普通祖先查找，显式声明与尚未安装分别记录。pnpm权威锁没有修改，本次不具备frozen安装或实际类型通过结论，后继仅由主控安排原串行dependency Stage。

本Reviewer完整读取三文件差异及两个实际JSON，以成熟JSON/YAML实际运行`node --max-old-space-size=512 .verification/review-type-plan.mjs 68bc4d8c5d182575fe6fe679a436033e8cdc6b36 7a0058fefb8499ffb37ffebc47e24866f0218ff7`，EXIT0：上游原锁版本、七项精确新增、完整manifest其余字段及清单56项逐项一致，变更集合只含这三文件。`check-revision.ps1`准确7a0058→68bc4d8c实际EXIT0，一Markdown/1本地链接及Git空白通过，BelowNormal/Node512MiB。没有安装依赖、创建Program或运行产品正文；T12/recover/后台Job仍为同Worker未完成范围，整体真实验收保持未通过。

## 准确7e090587的控制端恢复阶段

结论：**退回准确`7e09058776a06253c7134c60cc88868be27eefbd`的N20/N21，整体T12保持未验。** 以已接受main783203c3的SPEC21.1.3和079bd已审§3.4/§7.6为准，读取相对68bc4d8c的四十文件差异及实际消费者。先从定稿推导：真实caller校验先于legacy/executor/完成重放；事务先保留同一Run/root/context和旧generation；旧helper须两次新鲜退出证明；首次PTY效果前持久checkpoint；pending只观察原R；原命名Job以实际成员、来源及authenticated hello证明复开；任何未知继续原占用，完整关闭及R01—R10另由实际证据验收。

源码将recover接入公开字段/动作、当前start准入、写动作与输出分类、Node enum/私有durable方法。dispatcher的executionAdmission确在legacy route和executor前执行；12阶段分支核原生launch/owner/PTY/进程与父R参数，restored仍进入原生receipt和attestation分支。恢复使用原SQLite BEGIN IMMEDIATE及R/controller-pty、R/run-use，持久pty-spawn后才创建新helper；保留原Run/root/context及原pending Worker.start席位/关闭屏障，历史读取沿原方法handler。命名Job、当前用户/SYSTEM DACL、PID record、authenticated hello及真实成员交叉核查均有对应Go/TS/C++源码，native尚未编译或运行。

**N20，严重程度高：缺失checkpoint的pending恢复会进入删除原receipt的错误路径。** orchestration-mutation-executor.ts:232为pending recover传入JSON.parse(receipt ?? 'null')；原生begin完成、首次checkpoint前发生真实中断时，原receipt可为null。native-methods.ts:253—254先调用nativeObject，contracts.ts:89的成熟Zod record.parse对null抛出ZodError，尚未进入后面的effect包装。executor:240—246对该非prompt、非operation_unknown异常调用discardPendingMutationReceipt。此源码路径将原pending变成缺失，违反SPEC21.1.3及§3.4的checkpoint前未知保留条件；当前没有运行该故障或观察到第二个helper。建议动作：缺失或无法完整解释的原pending保持operation_unknown及原R，只有明确信息支持的无效果分支使用原生effect-free结算，不删除未知原件。

**N21，严重程度高：故障停止入口缺少同一目标句柄上的完整归属核查。** controller-recovery.mjs先按PID调用readServiceJobMember，再启动stop-owned-controller.ps1。后者持有Process.SafeHandle并核创建时间、程序文件名及保护PID，随后Kill；未在所持目标句柄上核原私有Job成员，也没有计划要求的新鲜PTY/incarnation、安装来源及父链复核。父进程先前按PID取得的观察不能补作该持有句柄的全部证明。建议动作：沿获准现有Windows能力，在真正终止前对同一实际句柄完成上述证据并保留结果，缺项立即拒绝；不在服务期间编译或按新PID寻找替代进程。此入口未运行，本Reviewer没有终止任何进程。

两项已通过`msg_83a96a758e65`交回原Worker，其`msg_3d7fed665f00`确认读取并持续返修。其余源码的存在和限定静态检查不代替下列真实覆盖：

| 已审条件 | 本次真实入口的覆盖范围 |
| --- | --- |
| R01 | 已有完成run-create子R的无helper调用拒绝；细pending/bootstrap、其他方法及restored分支仍未覆盖 |
| R02 | 另一真实项目/Run、批准撤销及settings变化仍需实际入口与条件 |
| R03/R04 | 活旧helper拒绝/原receipt解除、同lineage结束helper后恢复；故障停止须先修N21，未涉及活模型Worker |
| R05 | 已完成同R重放及旧C/第二任务拒绝；不同payload与两阶段不同R竞争仍需覆盖 |
| R06/R07/R08 | 同库重启、原handle关闭后复开、各真实checkpoint断开及pending观察仍需覆盖，缺失checkpoint先修N20 |
| R09 | helper失效后的历史task读取及新owner close；完整成果、全部未知资源和最终服务退出仍需覆盖 |
| R10 | recover的真实连接失败、处理后断开、取消130和stdout失败仍需覆盖 |

该入口remainingCases非空时外围返回6，未将核心路径登记为完整T12通过；所有上述正文均NOT_RUN，没有新增模型轮次。本Reviewer实际`check-revision.ps1`准确68bc4d8c→7e090587 EXIT0，三Markdown/7本地链接及Git空白通过，§15/§16两个锚点与实际标题一致。Worker的t12-source-static-01日志、两PS parser结果及六项退出0原件已读，24份语法、JSON/YAML、8Go格式、native JS和来源检查保持其执行归属。全部检查BelowNormal；本Reviewer仅使用成熟Markdown读取/静态源码核查，未创建Program、编译、安装、访问目标数据库/HTTP、控制进程或调用模型。T12余项、正文结算及通用TUI继续同Worker，类型/生成/完整构建和同安装候选双Worker全部未验，当前没有新重操作窗口。

## 准确efa31191的正文持久结算

结论：**局部静态接受准确`efa31191025e45ac65063993adb3d1e0d49d3f94`相对7e090587的七文件差异，适用当前structured输入路径。** 依据SPEC21.1.2及已审D01—D05，读取完整差异、host send消费者和journal字段。handler在第一次host调用前，将有正向无输入证明的prepare拒绝通过原recordMutationReceipt完成child；callback缺失或保存失败保持operation_unknown。parent/child在同一原SQLite事务保存operation及完整正文关联，事务不能确认、host调用后的异常、refused或无确定submission继续unknown；实际host确定结果先持久完成child再交父请求，不只依靠markEffectPossible。

共用严格binding增加agent/structured模式，实际handover逐项比较Agent、session、operation、fence、body及原consumer，继续已有当前Go批准和最后同步guard。request.show只读原parent/child的相同binding及原journal submission，核fence/operation；handoverRecorded存在且尚无handedOverAt的pending区分queued，其余pending为journaled，与实际journal类型和fold相符。返回已确定的请求、Dispatch、Agent、模式、session及operation，缺少真实turn不制造值。

同一T13驱动增加原send/重放binding、完整公开输入文件与text block、真实只读SQLite中child completed/binding、B第二轮结束后原request.show accepted/session/operation的检查；复用既有两Worker及轮次，正文与所有模型均未运行。效果前拒绝、host后不确定和真实存储失败仍须自然条件验证，terminal正文继续原Worker实施。实际`check-revision.ps1`准确7e090587→efa31191 EXIT0，两Markdown/7本地链接及Git空白通过。Worker的input-settlement-static-01日志及三项退出0已读，5份parser/JSON/YAML/文档检查保持原执行归属，类型、生成、构建及实际队列未验；此限定接受不解除7e的N20/N21。

## 准确8894055a的pending checkpoint返修

结论：**准确`8894055a6fea1cabcd3753381c7cb1571310bdaf`已修正N20的JSON/null入口；N20仍退回scope-reserved后续异常路径，N21继续未完成。** 完整读取相对efa的两份TS及报告差异。executor将原JSON解码失败转为operation_unknown，handler将null和字段解释失败同样保留unknown，并比较原invocation、owner/scope、旧generation及准确children/R/run-use；这些源码修正已确认。

scope-reserved的后续解析及db.getRun仍位于新增try/catch之外。共享scopeSchema为兼容尚无Run的阶段允许空run_id，随后该分支的min(1)解析可直接抛出ZodError；真实Run读取异常同样缺少unknown包装，原非prompt executor仍可能删除该pending。建议动作：完整pending观察中的解析、读取及证明错误保持operation_unknown，只有完整正向无效果证明支持的controller_unverifiable/effectsApplied=false按原生规则结算。此剩余路径已通过`msg_f5164ac46155`交原Worker，没有操作业务数据库或制造故障。

实际`check-revision.ps1`准确efa31191→8894055a EXIT0，一Markdown/1本地链接及Git空白通过。Worker的n20-static-01两TS/JSON/YAML、文档日志及两项退出0已读；全部只读检查BelowNormal、Node512MiB。真实begin后/首checkpoint前中断、其他pending阶段、N21停止、Program/生成/构建/安装及全部模型均NOT_RUN，同Worker持续返修，不以局部源码修正登记T12完成。

## 准确c962bea7的完整pending观察保护

结论：**局部静态接受准确`c962bea7ca39b9ca2190ee60380a80cc08f2ed42`的N20源码返修。** 完整读取相对889的handler和报告差异，全部pending字段解释、scope-reserved后继解析、真实Run读取、退出证明及完成观察均进入同一effect包装，未知异常统一operation_unknown。只有完整原记录、同一原consumer及两次成功新鲜观察支持的明确无效果结果作为内部返回值退出包装，才在外层抛出带effectsApplied=false的原生拒绝；继续使用原executor结算，未增加删除未知原件的路径或重放效果。

实际`check-revision.ps1`准确8894055a→c962bea7 EXIT0，一Markdown/1本地链接及Git空白通过。n20-static-02的实际单handler成熟parser、文档日志及两个退出0已读，Program未创建、正文未执行。N20已定位源码路径修正确认，真实中断、存储失败和effect-free解除的实际结果仍NOT_RUN，不据此登记恢复、安装或T12通过。

## 准确3c4e0364的同句柄故障归属与身份消费者

结论：**准确`3c4e0364642a01c68c3b75a072d81573d6d6f0ed`已修正N21对应源码；控制端路径退回N22，全部真实运行未验。** 完整读取相对c962的十文件差异、401行Go控制程序、TS只读检查及实际身份producer。普通控制程序在原候选构建的Go串行阶段预编译，Node检查同样预构建并进入strict/noEmit实际类型范围；服务期间没有编译或新故障钩子。

Go持续持有目标root/helper和daemon句柄，真实同句柄核Job/DACL、创建时间及完整程序来源，保护PID直接拒绝；持有之后才调用原DaemonClient的authenticated hello/listSessions，并核完成owner receipt与私有SQLite原件。真实Toolhelp snapshot建立helper→root→daemon父链，持有链内句柄并核创建顺序、安装或允许的系统程序来源，终止前再次核活身份、成员和观察时效。只有原helper/root句柄取得终止权限并被TerminateProcess，随后等待这些相同句柄；daemon、父链及整个Job均不作为终止对象。父链缺失、观察过期或错误立即失败，部分终止及关闭失败返回6。上述源码修正确认，实际Go编译、链路和停止均未运行。

**N22，严重程度高：createController比较两个语义不同的原生incarnation字段。** 当前native-methods.ts:133仍将live.processIncarnation与binding.processIncarnation直接比较。实际orca-runtime-agent-prompt-request-correlation.ts:13—31的getTerminalPromptRequestBinding将daemon原始incarnationId放入binding；terminal-interactive-wait.ts:172—192的getTerminalProcessIncarnation在原始值存在时返回ptyId加冒号及原始incarnation，terminal-orchestration-authority.ts:137用此完整值生成live authority。正常daemon提供原始值时，两者必不相等，控制端已创建PTY后会被该条件拒绝，open与recover均受影响；当前只确认源码路径，没有实际启动结果。建议动作：按原生接口各自语义核原始PTY incarnation和完整caller process incarnation，保留真实handle/PTY/launch/current证明，不伪造身份或放宽检查。此项已通过`msg_66685719f000`交原Worker。N21检查程序比较daemon原始session.incarnation与binding原始值符合这两个接口的含义，不将其改为完整caller字段。

实际`check-revision.ps1`准确c962bea7→3c4e0364 EXIT0，两Markdown/7本地链接及Git空白通过，原§15/§16锚点保持。Worker的n21-static-01语法、文档、类型配置日志和五项退出0已完整读取，5份parser、Go格式、配置20个真实根文件名及来源检查保持Worker归属，没有创建Program或执行probe。全部只读检查BelowNormal、Node512MiB；类型/生成/Go-C++构建、Job及停止、全部T12余项、模型和同安装候选双Worker继续未验，修复由同一Worker完成，当前没有新重操作窗口。

## 准确5110a888的原生incarnation消费修复

结论：**局部静态接受准确`5110a888d0586a91ebd07cfecba7486261799d3f`的N22源码返修。** 完整读取相对3c4的四文件差异，并与三个原生producer逐项对照。createController沿真实controller handle调用原getTerminalProcessIncarnation取得完整值，与live authority及同一ptyId核对；原始binding仍由实际daemon inventory单独核对，owner分别保存两个实际接口原值。exactController另核当前完整值与原caller_authority相同，继续原始PTY/binding和真实root/helper创建身份条件。没有手动拼接、截取身份或放宽launch证明，controller-inspection的raw字段比较保持。

实际`check-revision.ps1`准确3c4e0364→5110a888 EXIT0，两Markdown/7本地链接及Git空白通过。Worker的n22-static-01两TS成熟parser、文档来源日志和两个退出0已读，保持其执行归属；本Reviewer只读检查使用BelowNormal及Node512MiB，没有创建Program、运行正常helper启动、recover、故障停止或模型。N22已定位源码条件修正确认，N20/N21的限定源码结论保持，全部实际分支、T12余项、通用terminal、完整构建及同安装候选双Worker继续未验，原Worker持续实施。

## 准确26cf55a5的私有daemon失联政策

结论：**局部静态接受准确`26cf55a56a25a56f338973da88b83f35564d7282`的N23对应源码修复。** 原Worker先以msg_59da8950370f报告自动respawn来源，本Reviewer独立读取准确5110的inventory、recovery、init及host：listProcesses经withDaemonRetry，在daemon gone且存在respawnFn时doRespawn，原init提供resetHandle/ensureRunning回调。该读路径可能创建替代daemon，与pending只观察及原未知占用规则不符；未运行或观察到真实替代行为。

完整读取本次五文件差异及实际constructor后确认：仅SWA host显式automaticRespawn=false，原可选policy在init处不给adapter respawn，实际constructor将其保存为null；withDaemonRetry的无回调分支抛原错误，后台reconnectAfterWriteFailure的无回调guard直接返回，相关自动替换分支同样以回调为前提。原初次显式初始化及其他宿主默认保持，不增加恢复服务或状态。Worker报告§21关于后台函数直接抛出错误的表述已通过`msg_6fff8f5f2869`交原Worker修正，实际写失败仍沿其原调用链结算。

实际`check-revision.ps1`准确5110a888→26cf55a5 EXIT0，两Markdown/7本地链接及Git空白通过。daemon-failure-policy-static-01的三TS parser、文档来源日志及两个退出0已读，保持Worker归属；本Reviewer没有创建Program、启动或故障停止daemon。N23已定位源码修正确认，真实失联下原R保留、无替代和无重发仍须原R06/R08及所属条件验证，没有新增故障目标或重操作权限。

## 准确e328d739的R05真实竞争入口

结论：**退回准确`e328d739113c06b6ac13ba0c5eaf87fc208f863c`的N24阶段证据判定，全部新增正文未运行。** 完整读取相对26cf的四文件差异。驱动使用真实CLI与只读SQLite，在原一次恢复期间每阶段最多提交一次不同R，等待已启动请求结算；核真实operation_in_progress、竞争receipt按无效果规则解除、无竞争run-use及Run/Task数量保持。完成后增加同R不同真实controller payload的request_mismatch、原receipt保持和完整daemon inventory保存，未制造响应或修改业务数据库。

**N24，严重程度中：发出竞争请求前的阶段不能证明实际准入时仍处于该阶段。** observeCompetition先读取原R的scope-reserved/pty-spawn，再await真实CLI；随后仅凭operation_in_progress把先前阶段登记OBSERVED。原R可在CLI准入前推进到后续阶段，即使这次确实命中共同竞争屏障，现有记录仍不足以证明两个指定阶段分别命中。建议动作：同时保存竞争返回后的原R真实checkpoint，仅在状态/阶段证据足以证明指定阶段时登记OBSERVED，发生推进、完成或缺证时保留INCONCLUSIVE及remainingCases，不增加暂停/延迟钩子、重试或业务库写入。此项已通过`msg_43dacadda1e1`交原Worker。

当前PTY断言比较故障前后存活sessionId集合，只能证明指定旧标识移除、实际新标识加入及其他标识集合保持；报告对其他原件全部不变的表述须据该实际范围修正，原进程/会话身份保全继续相应真实用例。实际`check-revision.ps1`准确26cf55a5→e328d739 EXIT0，两Markdown/7本地链接及Git空白通过；t12-r05-static-01两TS/MJS/JSON/YAML、文档日志及两个退出0已读。所有检查BelowNormal、Node512MiB，Program、竞争CLI、SQLite正文、PTY故障、服务及模型均NOT_RUN；R05实际命中、其余T12、通用terminal、类型/生成/构建和同候选安装双Worker继续未验。

## 准确4ceff1ad的N24返修与后继Stage入口

结论：**局部静态接受准确`4ceff1aded43eca345c9d14729d51697baf23069`的N24源码判定、文档修正及所述新窗口技术入口。** 完整读取相对e328的三文件差异。竞争返回后另存同一原R的实际checkpoint，OBSERVED同时要求真实operation_in_progress、原R前后均pending、相同原payload及同一指定stage；推进、完成、缺失或缺证继续INCONCLUSIVE并加入remainingCases。原每阶段一次竞争、原生无效果结算、无竞争run-use及数量检查保持，不增加控制产品进度的钩子或重试。文档准确描述存活sessionId集合，§21后台void恢复函数的返回语义也已修正。N24已定位判定修正确认，实际两个阶段是否自然命中仍NOT_RUN。

已完整读取Worker自有next-window-04.ps1，并仅运行PowerShell Parser.ParseFile及Get-FileHash，未执行脚本。新入口SHA256实际为`05b0bf24bca39ba075a683063ab26e19f50ac09d39974516b00705a06ff420a1`，Parser错误0；原next-window.ps1和build-06/next-window-executed.ps1均仍为`bae62ff4c1eee443f0efafabf87f814f4a47f6177b9e0812a0ed9622cd087522`。新candidate06本次只读观察不存在，旧构建入口与失败候选未改变。

新入口以正式邮件ExpectedSource核准确工作树、分支及干净HEAD，唯一新EvidenceName保留原件；dependencies/build逐级核私有CAS至本树无重定向。每条重命令前重新使用原资源采样，dependencies串行lock-only和frozen ignore-scripts/verify-store-integrity，typecheck使用实际TS6严格入口，generate使用原生成器/Node512，build只指向新candidate06并沿原构建器独占创建；其余Node1024、Go4/-p4/512MiB及BelowNormal保持。入口每次只执行一个选定Stage，不能据技术接受取得窗口；主控仍须新鲜核原S/所属、停止和资源，正式给定准确来源及单次Stage，结束后沿原保全流程恢复。当前代码继续开发，尚未冻结或执行任何Stage。

实际`check-revision.ps1`准确e328d739→4ceff1ad EXIT0，两Markdown/7本地链接及Git空白通过，原锚点保持。window04-n24-static-01的一MJS/JSON/YAML、文档日志及两项退出0已读，保持Worker归属；本Reviewer的Parser及摘要核查同为只读。Program、依赖锁/安装、生成、构建、竞争/恢复/停止、服务及模型均未执行或仍未通过，全部T12余项、通用terminal和最终同安装候选双Worker继续同Worker，主控实际读取回执仍待答。

## 准确616eab31的R10真实CLI故障入口

结论：**退回准确`616eab312cfb0741fc89cf8ae65027aceac2748b`的N25输出契约检查及N26原控制端诊断，全部新增正文NOT_RUN。** 根据main783203c3的SPEC21.1.3及079bd计划C05/C09/R10，先独立推导实际取消须130及单一失败envelope、stdout关闭须6及已知原R查询诊断、原请求与唯一PTY/runUse保持、缺证不登记命中，再完整读取相对4ce的九文件差异和实际消费者。没有增加模型调用、故障目标或运行窗口。

原构建入口在服务前用实际Go预编译cli-fault-control，闭包检查要求该文件。控制程序分两次创建隐藏console及真实CLI独立group，在所持句柄的创建身份、完整程序、保护PID和两个实际console成员核实后定向发送CTRL_BREAK；stdout关闭使用真实StdoutPipe。驱动要求实际130或6、同一句柄退出、无控制错误及非超时，未命中保留INCONCLUSIVE。其原始流、真实request.show、私有SQLite及authenticated daemon库存均有明确来源；初次及重放后比较全部PTY六身份字段，原Run/root/spec、全部runUse集合及唯一原R子请求另行核查。pending观察和网络两项继续列入remainingCases，驱动和外围保持6；两项草稿覆盖缺口在本次准确源码中已补充。

**N25，严重程度中：输出缺证仍可登记取消或stdout失败命中。** cli-failure.mjs先计算hit/status，再解析实际stdout；空流保持cliReply=null，坏JSON或多份JSON仅记录parseError，取消130仍可能登记OBSERVED，未验证C05明确要求的单一失败envelope。Go控制程序将实际管道读取/保存错误写入actual_pipe_read_error，该取消判定同样未检查。stdout-close允许真实部分输出，但驱动没有读取或断言实际CLI stderr中的已知原R及查询诊断，退出6单项不足以覆盖C09。建议动作：取消场景核完整保存的单一真实JSON、失败结果及原关联；stdout关闭按真实部分输出和stderr核原R及查询信息，缺证不能OBSERVED，不要求其stdout为完整JSON。使用成熟JSON解析和真实流，不新增Writer、响应替身或时序钩子。本项已通过`msg_2fd8345dcf4f`交原Worker。

**N26，严重程度中：恢复成功回执后的stdout失败没有保留原C_old诊断。** 独立继续读取准确616e的cmd/swa/main.go和output.go及079bd§3.4.1。实际响应解析后main.go以reply覆盖output.identity；恢复成功时其controller为真实C_new，随后publish失败只进入stderr。output.go:83仅输出当前identity.Controller和requested_task/dispatch，未输出仍保存在selected.Controller的原C_old；projectID存在时查询命令也只含project和R。完整stdout写入失败可以没有任何envelope字节，当前诊断不能同时保留原C_old与实际C_new，不满足全部失败路径保留原R/C_old及实际已知新身份的定稿要求。建议动作：使用已验证command中保留的原选择器，连同实际响应已知C_new/Run/root核实际stderr；不从expected checkpoint补造身份，不增加公开接口。原project加R仍可查询原件，未声称恢复效果或原receipt丢失。本项已通过`msg_bead5052222b`交原Worker。

实际`check-revision.ps1 -Previous 4ceff1aded43eca345c9d14729d51697baf23069 -Candidate 616eab312cfb0741fc89cf8ae65027aceac2748b` EXIT0，两Markdown/7本地链接及Git空白通过，既有§15/§16锚点保持。Worker的r10-source-static-01五份TS/MJS语法、JSON/YAML、Go格式、PS Parser、文档来源日志及四项退出0已完整读取，保持Worker归属。全部只读检查BelowNormal、Node512MiB；本Reviewer未执行Program、Go构建、安装、SQLite/HTTP正文、console、故障或模型。类型/生成/完整构建、R06/R08与其余T12、通用terminal及同安装候选双Worker继续未验；后继Stage须主控指定本次之后的准确源码，原Run实际读取回复仍待答。

## 准确7c812a1a的N25及N26返修

结论：**局部静态接受准确`7c812a1a972ed914bae3b9cd864702e5426b9492`的N25/N26对应源码修复。** 完整读取相对616e的五文件差异，并重新对照真实Go输出producer和OrcaNativeError字段。取消场景在真实130窗口成立后，严格UTF-8解码和成熟JSON.parse须得到单一完整失败envelope，核原R/C_old、实际delivery_unknown及inspect_original_request和项目查询诊断；读取、保存或未分类复制错误不能作为取消证据。stdout关闭允许真实部分输出，实际6之外还要求真实pipe读取错误、实际stderr输出失败、已知原关联及原请求查询信息，缺证加入remainingCases和INCONCLUSIVE。

Go控制程序以标准os.PathError区分原读取、文件保存及未分类错误，后两者使控制程序失败；stdout关闭的预期读取错误继续保存实际原件。outputFailure使用已验证selected.Controller输出requested_controller，当前controller/Run/root仍取已收到的真实回复，原C_old与已知C_new分别保留。驱动核原选择器，C_new只接受真实completed receipt关联；实际诊断已知C_new时另核同一Run/root，没有使用pending expected身份。原PTY/runUse唯一性、同R重放六身份和全部动态限制保持。

实际`check-revision.ps1 -Previous 616eab312cfb0741fc89cf8ae65027aceac2748b -Candidate 7c812a1a972ed914bae3b9cd864702e5426b9492` EXIT0，两Markdown/7本地链接及Git空白通过，既有锚点保持。n25-n26-static-01全部初次及final日志、七项退出0和summary已完整读取，一MJS/JSON/YAML、两Go格式与文档来源保持Worker执行归属。全部检查BelowNormal、Node512MiB，本Reviewer未运行Program、编译、安装、console、故障、SQLite/HTTP正文或模型。真实130/6、全部R10/R06/R08及T12、通用terminal、完整构建和同安装候选双Worker仍未验；技术接受不授予重操作窗口。

## 固定 b287 源码直接复用的接续核查

结论：**现有原生核心继续直接复用；产品授权、隔离和效果结算的组装保持。整体运行验收继续证据不足。** 已完整读取主线SPEC21.1.1—21.1.5、当前CLI/REA参考及主控恢复记录，再对照真实 `b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e` 源码、安装CLI的RuntimeClient JS和准确7c消费者。后继3ba只改锁与文档，该源码范围未变化。本次不重复REA分析，也不将来源对照作为后继类型或构建的新增前置。

下表原路径均位于 `third_party/orca-v1.4.230/`；本轮逐项读取所列核心与产品消费者，静态关系只支持其对应范围。

| 原生实现及当前直接消费者 | 核实结果与具体接续建议 |
| --- | --- |
| `src/cli/runtime/client.ts`、`transport.ts` → `runtime/swa-orca/src/client.ts`、`controller-helper.ts` | 两个产品入口已直接实例化同一RuntimeClient并显式给定私有data及产品executable。与b287的准确文件差异显示，client只增加已审owner capability及SWA兼容调用，transport只增加该本地包络字段；继续保留原协议解析、runtime/source fence和未知请求诊断，不另写客户端或socket协议 |
| `src/main/orcad/orcad-lifecycle.ts`、`OrcaRuntimeService`、`RuntimeRpcShutdown` → `runtime/swa-orca/src/host-entry.ts` | 原profile准入、instance lock、cleanup顺序及运行服务已复用；lifecycle差异为可关闭browser provider的已审可选参数。产品须保留私有路径、秘密存储、只读owner、精确方法白名单及真实空闲关闭条件，不能直接使用原全局host组装 |
| `rpc/methods/orchestration/runs/runs.ts`、`worker/worker-methods.ts`、`worker-list-method.ts`、`messaging/message-methods.ts`、`runs/mutation-request-show.ts` → host方法表、helper与`nativeHistory` | 状态查询和写动作都调用原handler；helper固定原R子请求和真实controller，保留原Run/Task/Dispatch归属。原CLI `worker-launch-handler.ts`提供完整flags及from解析，其直接照搬会引入未批准参数；当前产品只映射已审动作和选择，无需再复制一层原handler |
| `rpc/orchestration-mutation-executor.ts`、原mutation receipt存储 → `native-methods.ts`及`controller-admission.ts` | 产品继续使用原executor及同一SQLite原receipt/checkpoint；严格准入在legacy和executor之前，pending恢复观察和效果后unknown属于已审产品补充。保持原存储和请求身份，不另建执行或恢复账本 |
| `runtime/orchestration/send-agent-turn.ts`、`native-chat/agent-session-wire/structured-agent-session-task-queue.ts`及原host → `worker-input.ts` | structured正文已进入原host.send及waitForSendSettlement，原session串行队列与b287相同。当前批准、原operation/fence和最后写入前同步检查仍必要；继续沿原队列及其真实adapter消费者。通用terminal正文仍须完成SPEC21.1.2规定的普通sendTerminalAgentPrompt路径，实际结果没有通过 |

Go App的项目/Grant/单CLI批准、产品输出与退出语义、私有账号投影及工具政策、真实helper身份、所属Job和完整自产成果均为已确认产品行为。它们继续由现有App/adapter/composition承担。MIT原许可和Lovecast归属保留。原核心源码与整个生产依赖表的保留范围分别判断；完整运行闭包继续依据实际metafile、动态子入口、native和独立安装结果，静态未连边不能支持删除依赖。

独立读取也发现helper中task.list/show/result及worker.show/list/read聚合与`nativeHistory`存在相近逻辑。当前公开只读调用已经沿client默认分支进入swa.observe；本次不以此为由改动稳定模块或阻断构建。原Worker应先确认完整真实消费者，只有证实没有调用的helper分支才在原写域中移除，保留准确差异及适用回归。该建议不增加新的框架、公开命令或审批。

## 准确3ba76a85的真实依赖锁与安装结果

结论：**限定接受准确 `3ba76a85cc968224f283315b64a4be4c27f1fcce` 相对7c的真实锁与两份记录，未接受类型、构建或运行结果。** 完整读取三文件差异和dependency-04全部命令日志、SOURCE、五项退出、两组资源及运行/退出观察原件。原runner05b0以准确7c开始，原pnpm11.27.0先lock-only、后frozen ignore-scripts/verify-store-integrity，7.9秒和2.1秒均成功；两次资源准入与外层最终退出均0。资源CPU11.3/8.0%及14.6/23.2%，可用内存均超过4GiB，单调间隔2.5012007及2.5016213秒。deprecated/peer提示和两次所属为空记录保持，不将其改写为无警告运行或全部依赖兼容。

本Reviewer使用已安装成熟YAML/JSON独立核真实锁、清单和已安装package.json：56项声明逐项匹配，manifest与dependencies.lock.json相对7c保持；其他importer、原patch、override及设置保持。七项已审开发依赖已实际锁定并在产品node_modules中核得相同版本，Node类型25.9.5和TypeScript6.0.3同样对应。pnpm自然peer变化只有runtime importer的Electron toolkit→43.7.8和serve-sim→React类型19.2.18，以及相应snapshot；原声明没有升级。该读取没有执行包正文、安装脚本或创建TypeScript Program。

实际命令及结果：

- `pwsh -NoProfile -File .verification/check-revision.ps1 -Previous 7c812a1a972ed914bae3b9cd864702e5426b9492 -Candidate 3ba76a85cc968224f283315b64a4be4c27f1fcce` EXIT0：Git空白及marked18.1.0检查通过，两Markdown、7本地链接，原§15/§16锚点保持。
- `node --max-old-space-size=512 .verification/review-runtime-lock.mjs 3ba76a85cc968224f283315b64a4be4c27f1fcce 7c812a1a972ed914bae3b9cd864702e5426b9492` EXIT0：成熟YAML核锁和真实已安装声明全部通过；BelowNormal、Node512MiB，没有重复依赖命令。

主控已收回dependencies04窗口，后继strict类型、生成及完整native/Go构建另授。原全部故障、T12、通用terminal、模型及同安装候选双Worker均保持未验。当前没有正式PR，本轮不发布GitHub review，准确交付后的COMMENT review保持待执行。

原Worker随后提交忽略目录 `reuse-comparison-20261021.md`。本Reviewer完整读取并独立核全部helperCall、taskCreate的精确action条件、controllerDeliver五项mutation白名单及swa.observe六项只读入口：helper中六项历史只读分支与collectTaskResult没有当前产品调用者。限定清理范围和复用原T01/T07/T12/T13验证成立；应在已交类型窗口的源码冻结解除后实施，只删除这些分支及因此无用的zod import，保留identify/observe/recover、全部写动作、nativeHistory和完整成果/授权消费者，准确后继另审。该范围已通过 `msg_440ca4668039` 交原Worker，不增加模型轮次、运行授权或构建前置。

主控 `msg_9315b3a0edac` 已明确完整读取准确3ba三文件差异、dependency04原件及af605/c6fb两报告新增，原hook `msg_f210e2f1fbe6` 的实际唤起确认。六helper分支清理由原Worker在构建窗口归还后实施，当前已取得的源码和实际检查不因此重写。

## 准确c42a695f的真实生成与类型失败记录

结论：**局部接受准确 `c42a695f756886c2afcc8f36f0af6339dae9ec29` 相对3ba的生成产物及两份记录；严格类型检查保持FAIL。** 已完整读取三文件差异、原生成脚本及对应OpenAPI/JSON Schema。生成产物增加controller.recover、nativeConfigurationSource、SwaNativeLaunchPolicyV1及AgentSelection.native_launch，保持已审账号来源、三个API、空工具政策和必填字段语义；没有变更原契约、程序实现、依赖或权限。

generate03的准确SOURCE为3ba，原runner05b0只运行既有generate-client.mjs，实际生成、资源及外层退出均0。资源CPU9.5/9.4%、可用18586574848/18619158528字节、单调2.5010712秒，Node512及BelowNormal保持；两份UTC13:12:28.7262117/13:12:41.1910594所属为空原件已读。主控已收回生成窗口，本Reviewer没有重跑生成器或字节再生成检查；本轮只确认真实生成来源、完整差异和其与已审契约的字段对应。

typecheck01同样以准确3ba运行，完整日志明确为Reached heap limit，约1022.7MB时OOM，TS子退出134、外层1、资源0；没有任何语义类型诊断。CPU5.5/5.0%、可用18849153024/18853732352字节、单调2.5010102秒；实际四进程身份及UTC13:09:07.9578263/13:09:21.5806254两次所属为空原件已读。真实主机可用内存不能替代已批准Node1024 heap额度；新额度待答，未扩大额度、关闭strict/skipLibCheck规则或重跑。本次失败不能证明源码语义通过，也不支持推断源码已经出现某项TS诊断。

`pwsh -NoProfile -File .verification/check-revision.ps1 -Previous 3ba76a85cc968224f283315b64a4be4c27f1fcce -Candidate c42a695f` 实际EXIT0，唯一前缀解析为上述完整SHA，两Markdown/7本地链接及Git空白通过，原§15/§16锚点保持；marked18.1.0、BelowNormal、Node512MiB。本Reviewer未执行TypeScript Program、编译、生成、服务、数据库/HTTP正文、模型或安装，当前没有正式PR，COMMENT review仍待准确交付。全部实际故障、T12、通用terminal、完整构建及同安装候选双Worker继续未验；候选06/build07是否执行及结果由主控后继窗口另行记录。

原生成接受hook `msg_86b971cb67b6` 已获主控 `msg_ca043545c824` 明确完整读取准确三文件、30c258两报告新增及真实原件的回复，实际唤起确认。

## 准确c42的build07失败与分析日志目录

结论：**准确c42的完整包构建FAIL，交原Worker修复；每个native包独占短分析日志目录的窄技术范围成立，准确实现和真实复编另审。** 本Reviewer完整读取build07日志、SOURCE、资源、内外退出、两次运行观察及UTC13:18:42.8148453/13:19:13.0583412两次所属为空原件。资源CPU6.5/5.0%、可用18707197952/18706276352字节、单调2.5010402秒通过；构建与外层均1。主控已收回本次窗口，candidate06及全部失败原件保持。

324项prod deploy、363项供应链核查、原三个固定Node输入200、Python3.13.13、VS2022及CL=/utf-8实际进入conpty.cc。日志仍有C4267警告，首fatal为C1258不能保存conpty.nativecodeanalysis.xml；没有将这些局部阶段计为完整native、Go或VERSION完成。旧头文件17行错误在本日志未出现，具体C1258根因尚未验证。

独立用成熟.NET XML读取实际conpty.vcxproj，两个配置EnablePREfast均true，IntDir为原相对对象目录；真实源包含conpty.cc、path_util.cc及node-gyp自动加入的win_delay_load_hook.cc。失败XML完整路径269字符，当前父目录存在、文件缺失；CL.command.1.tlog仅2字节，不能作为完整实际CL命令证据。长度观察用于定位，未单独判定根因。

已完整读取指定[Microsoft分析选项](https://learn.microsoft.com/en-us/cpp/build/reference/analyze-code-analysis?view=msvc-170)：日志参数指向已存在目录且保留末尾反斜线时采用目录模式；文件模式合并日志。按[CL环境变量](https://learn.microsoft.com/en-us/cpp/build/reference/cl-environment-variables?view=msvc-170)，CL选项前置、_CL_选项后置；带引号路径必须按[实际参数规则](https://learn.microsoft.com/en-us/cpp/c-language/parsing-c-command-line-arguments?view=msvc-170)保留尾部反斜线与闭引号。原生实际消费仍须本次真实编译证明。

原Worker `msg_1e020216d2eb` 的具体范围为：沿原四包native循环，在现有私有work中分别独占创建a0—a3并核canonical所属，保留/utf-8与全部原分析/编译选项，只追加日志目录参数；不同包串行、包内MSBuild4保持，并核同名源码的日志归属。该范围不变更源码、ABI、headers、分析规则或全局配置，必要验证由原完整构建覆盖，已通过 `msg_10b6e8a59ba9` 交原Worker。实际MSBuild后继参数覆盖、引号处理和日志文件结果在准确实现及复编中核查，不因静态方向成立宣告问题已解决。

本机没有Firecrawl命令，未安装依赖；网页工具初次无view直读失败，随后通过同一Microsoft官方view=msvc-170全文读取。未执行候选、编译、分析引擎或模型，没有新增运行窗口。严格类型仍为OOM失败；全部恢复、通用terminal、安装及同候选双Worker验收继续未通过。

上述build07退回及窄技术范围hook `msg_6067337645ec` 获主控 `msg_ae1bd22a9e49` 明确完整读取c20cd8ba新增和实际原件的回复，实际唤起确认。下一构建须使用独占candidate07/build08和新runner文件/hash，原05b0硬编码的candidate06已消费，不能覆盖旧原件；当前没有新运行窗口。

## 准确e7fd948a的日志目录与参数覆盖

结论：**准确 `e7fd948ad9214ffee033eb4701adad776bda27a4` 的短目录、转义及新runner范围静态检查通过；日志选项最终覆盖证据不足，后继准确修正另审。** 已完整读取相对c42的四文件差异。原native循环为四包分别创建现有work/a0—a3，独占mkdir及realpath/lstat核查均在原gyp调用前；1024字符限制、实际非空ConPTY日志检查和Go前恢复编译环境保持fast-fail，没有关闭分析或修改ABI、依赖、源码及全局配置。

成熟TypeScript6.0.3 AST独立解析原脚本，日志参数字符串末尾的真实字符为 `[92,92,34]`，即两个反斜线及闭引号；该检查未创建Program或执行脚本。实际node-pty项目含三源，其他三个已部署包的binding已读，process-tree六源及两个安全模块各自addon.cc与附加hook在各包内没有同名；每包独立日志目录保留包间归属。实际产生日志仍待复编。

`CL`在命令前置，当前安装 `Microsoft.CppCommon.targets` 的721/865/1052行把PREfastLog传给CL任务，`2052/cl.xml`也映射该日志选项。旧tlog仅BOM、项目未显式指定该属性，尚不能证明后继CL任务不会给出日志路径。[Microsoft选项顺序](https://learn.microsoft.com/en-us/cpp/build/reference/order-of-cl-options?view=msvc-170)规定冲突采用最右选项；结合已读CL/_CL_顺序，当前前置CL中的日志选项不能保证最终覆盖。这是静态可见的覆盖缺证，没有宣称已观察到真实覆盖失败。已通过 `msg_c25d3d942674` 交原Worker，仅将日志参数移到产品自有后置_CL_、保留CL=/utf-8及原宿主变量清理，并在Go前删除该后置变量；原引号、目录和真实日志检查继续保留。

新 `next-window-05.ps1` 已完整读取，实际SHA256为 `7a375ffec6f7ff532abbdde3926736f5d460233f37568ae08fea853876443131`，PowerShell Parser错误0。与当前旧04相比只改变candidate06→candidate07常量；旧04与build07执行snapshot当前均为原05b0。candidate07及build08本次只读观察尚不存在。来源、干净分支、独占证据、资源门限、私有CAS及各Stage额度保持，不执行脚本；后继正式窗口须使用准确新源码和本次新runner。

实际 `node --max-old-space-size=512 .verification/check-analysis-option.mjs e7fd948ad9214ffee033eb4701adad776bda27a4` EXIT0；`check-revision.ps1`准确c42→e7 EXIT0，三Markdown/7本地链接及Git空白通过，原锚点保持；全部BelowNormal、Node512MiB。未执行CL、Program、编译、生成、候选、模型或安装，C1258及严格类型失败均未关闭。

e7证据不足hook `msg_7a36811bcbbc` 已获主控 `msg_4e67eeb7e521` 明确完整读取5d859ad4新增、准确e7四文件和实际安装消费者的回复，实际唤起确认。

## 准确ada4089d的原生后置日志参数

结论：**局部静态接受准确 `ada4089de6a9fcaf72f3a91c0b0331ced11343d9` 相对e7的两文件修正及累计四文件范围，真实编译效果继续未验。** 已完整读取builder和报告差异，日志参数使用产品自有_CL_后置，CL继续为原/utf-8；每包独占目录、原宿主CL/_CL_先行清理、1024长度检查、尾部转义、实际非空ConPTY日志验证及全部原native编译参数保持。完成native循环后删除_CL_再进入原Go编译，不改变其他进程或全局环境。

该后置位置依已读原生CL规则解决本次前置覆盖缺证；未宣称MSVC实际消费、分析文件生成或C1258修复已通过。新runner05的7a375、旧04及snapshot的05b0、仅candidate路径差异和Parser0等未变化范围复用前次准确检查。仍需主控单独授予准确ada和新runner的全新candidate07/build08，旧候选与失败原件保留。

实际 `node --max-old-space-size=512 .verification/check-analysis-option.mjs ada4089de6a9fcaf72f3a91c0b0331ced11343d9 environment._CL_` EXIT0，成熟TypeScript6.0.3 AST确认后置变量和真实尾部 `[92,92,34]`，未创建Program；`check-revision.ps1`准确e7→ada EXIT0，一Markdown/1本地链接及Git空白通过，BelowNormal、Node512MiB。累计范围只改原builder和三文档，没有新依赖或程序权限。严格类型OOM、全部真实功能及同安装候选双Worker继续未通过，本Reviewer没有执行编译、候选、服务或模型。

## 用户解除资源限制后的累计审查

主控 `msg_5d03182be32c` 明确转达用户解除资源限制、提高普通修复效率的决定；准确main `1ca10457b95fe3e2ed3ea5a3813ec66295b035ec` 已更新六份管理文档，本Reviewer完整读取准确差异并接受其限定范围。后继strict可用8192MiB，原Worker集中检查全部类型、工具链与四个native模块，复用原固定校验输入，分模块修复后执行最终完整包；本Reviewer直接累计反馈，不再逐小改等待主控资源窗口或回执。历史检查保留实际参数和失败结果，原账号、工具、资料、批准及最终同安装候选验收要求保持。ada静态接受hook `msg_6f0b4c91eee7` 已获主控 `msg_f68c9a2feb53` 明确读取准确源码和d1d9ca31报告后回复。

**build08完整包仍为FAIL。** 已读准确SOURCE=ada、完整build.log、资源、内外退出及UTC13:38:59.1407035/13:39:15.3641194两次所属为空原件。资源样本CPU8.5/8.1%、available18029510656/17906249728字节、间隔2.501628秒；child1/resource0/outer1。日志记录Python3.13.13、VS2022及固定Node24.21.0，ConPTY的conpty.cc、path_util.cc和win_delay_load_hook.cc均进入编译并创建lib/exp；三份真实分析XML各59字节，成熟.NET XML解析确认DEFECTS根节点和零子元素。C1258在本次日志未再出现，原C4267警告保持。后续FileTracker FTK1011无法创建原长目录中的link-rc、link-cvtres及link读写tlog，未完成四个native模块、Go、VERSION或swa.exe，candidate07保留不可安装。

只读本机v170真实文件确认：Microsoft.BuildSteps.targets:219默认TLogLocation引用IntDir及项目名；Microsoft.Cl.Common.props:113、Microsoft.Link.Common.props:24及其他RC/Manifest/Lib消费者使用该目录。对象文件与编译器PDB依赖IntDir，链接器PDB依赖OutDir；路径及项目归属需要统一处理。node-gyp12.3.0的lib/build.js会将Windows收到的额外argv改写为/t:，MSBuild属性须经真实入口和最终求值核实。已通过 `msg_164f46834c6c` 直接交原Worker；保留代码分析、跟踪、ABI和原编译规则，实际修复与四模块复编另审，没有关闭FileTracker或修改全局配置。

已只读typecheck02的SOURCE=ada、实际tsc-exit=2及日志首尾，确认取得语义诊断；native-methods.ts:160缺少nativeError导入、边界值可能undefined、真实类型路径和Windows安全契约等已通过 `msg_7ad118b74966` 交原Worker累计修复。完整入口参数、全部诊断处理及后继类型结果以原Worker准确交付另审，本轮未将该失败视为通过，也未启动第二份类型检查。

### candidate07的实际执行依赖

以build08生成的21份实际metafile和JS为对象，`node --max-old-space-size=512 .verification/review-external-dependencies.mjs` 使用JSON.parse读取全部outputs.imports，`node --max-old-space-size=2048 .verification/review-dynamic-consumers.mjs` 使用已安装TypeScript6.0.3逐份解析成品AST，均实际EXIT0；未创建Program或执行候选。全部metafile去重6048个inputs，host为5769个；这些输入数量只用于定位消费者，不作为进度。

直接非Node内置external包括node-pty及其两个内部入口、@orca/windows-security、@parcel/watcher、ws的可选bufferutil/utf-8-validate和ssh2的两个可选native相对路径。动态createRequire还保留以下真实消费者，单看inputs或outputs静态边均不充分：

| 实际成品位置 | 消费内容 | 当前结论 |
| --- | --- | --- |
| host.cjs:26942、27735 | @orca/windows-registry、@vscode/windows-process-tree | 保留四个原生模块及其真实JS装载入口 |
| host.cjs:179274 | @linear/sdk | 真实动态加载，保留当前安装依赖 |
| host.cjs:111867 | emojibase-data/en/shortcodes/emojibase.json | 实际worktree名称处理加载器存在，candidate07的对应包目录缺失；实际可达范围与补齐方式交原Worker核查 |
| agent-connection、swa-provider-process-tree、host等成品 | node-pty/lib/utils、真实ConPTY文件与支持资源 | 保留包内文件、原slotSourceFiles结果及原ABI |
| profile-state-writer、claude-profile-setup等后台成品 | Windows安全加载器与进程树 | 子入口与主入口一并核查 |

provider-job-bootstrap.mjs虽然使用packages=external，实际成品仅导出父IPC准入函数、只导入Node内置模块。真实Pi政策从已发现并验证的原生CLI所在安装目录按绝对文件URL导入配置/provider/catalog模块；零metafile输入不能证明原生Agent依赖可以删除。agent-browser二进制、ripgrep、Claude SDK资源和其他动态资源的完整必要范围继续未知，尚无删除或独立安装通过结论。

明确可供原Worker验证的重复纯JS安装依赖候选为zod、yaml、diff、smol-toml、jsonc-parser、@streamparser/json、tweetnacl及proper-lockfile；实际成品已包含对应190/72/19/10/6/9/2/4个输入，21输出没有对应external，已识别的动态createRequire没有读取这些包。可保留构建依赖并检查prod deploy的重复副本是否必要，最终结论必须结合传递消费者、资源与独立安装/真实任务。现有THIRD_PARTY依赖node_modules原包携带LICENSE/notices，删除重复文件前须继续保留内联代码对应的原许可和归属。这些具体保留、缺失及有条件候选已通过 `msg_b1b8858f3b76`、`msg_3a3b957f1439` 交原Worker，没有修改实现或把本调查设为四模块修复的前置。

## 准确1209aee3的累计源码、完整构建及安装失败

结论：**退回准确 `1209aee399a2c0b0677d7a4dbedc5f87880d14ab` 的安装复制与直接依赖清单，已核类型、原生编译及完整构建结果按限定范围成立。** 本轮累计读取相对ada的49文件差异；修复继续由原Worker完成，后继准确提交另审。普通修复和已授权检查继续执行，本结论不构成新的资源窗口或等待回执条件。

`build-swa-native.mjs`只从本次部署根解析原包，在独占n0—n3短目录保留原源码及真实node-addon-api；原MSBuild跟踪、代码分析、Python3.13.13、VS2022、node-gyp12.3与固定Node24.21.0 ABI保持。产物只复制真实.node及node-pty原slotSourceFiles所列支持文件，中间文件留在work。build脚本显式选择复用输入时仍按原三个固定SHA及真实exe校验，没有下载失败后的自动替代路径。原六个helper查询分支及无消费者collectTaskResult已删除，既有读取继续由实际native handlers承担。

deny/fail改为保留原throw的函数声明；缺失nativeError使用原contracts导出；IPC send/disconnect先绑定真实方法并检查存在，Pi源header按实际已解析对象校验。Windows安全包装贯通原native的readServiceJobMember，其余上游差异主要为明确原类型和既有身份校验。journal拒绝原因继续由agentSessionFailureWords提供，符合原ResolveDispatchInput的规范失败类型，没有扩展原生协议接受任意错误句子。strict/noEmit/skipLibCheck=false和原类型编译器保持，没有虚假ambient模块或关闭严格检查。

资源解除同步去除原CPU/内存门限、强制优先级和固定Go/Node额度，原Job所有权、权限、私有目录、实际身份及退出条件保持。另发现Go采样删除原64处理器保护后缺少观测范围说明：[Microsoft GetSystemTimes](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getsystemtimes)明确超过64处理器时只汇总调用线程所属primary processor group。已将这一准确性问题交原Worker，当前工作区的cpu_scope补充尚未纳入本准确提交；本轮没有运行多processor-group机器验证。

已读取Worker实际原件和当前准确源码，限定结果如下：

| 检查 | 实际结果与范围 |
| --- | --- |
| strict06/strict07 | 两份完整日志均为空，退出均0；采用原TypeScript6.0.3、8192MiB及原strict配置。strict03—05和此前OOM/语义诊断失败保留 |
| native-check01 | 退出0，四个真实模块编译及require成功。其执行helper快照仍复制整个Release，当前helper只复制产物；差别由后继完整包另验 |
| build09/candidate08 | 退出0，最终source_revision与VERSION均为准确1209；四模块实际编译完成，最大中间路径188/227/229/229，各8个tracking文件。C4267及C6387警告保留，Go输出和VERSION完成 |
| candidate08闭包 | 退出0，记录34010个闭包路径、真实Node/native API与空Job观察；认证、后代Job及功能明确NOT_RUN |
| Go prerequisites01/03 | cmd/swa、internal/swarm/app和domain通过，两个真实故障控制程序包完成编译且没有测试用例 |
| Go prerequisites02 | FAIL；workspace与httpapi的实际测试缺少明确publication、SQL期待、独占输出或真实服务输入，没有登记为全部Go测试通过 |
| 准确1209四测试包编译与传输前检查 | 实际四份测试二进制及各自编译日志已读；原聚合退出2保留，首次-test参数错误保持。后继正确参数的三个传输前用例实际PASS/0，不扩大为真实已安装服务用例通过 |
| architecture01 | 实际EXIT0，47个包通过原依赖方向检查 |

本Reviewer实际执行 `check-revision.ps1 -Previous ada4089de6a9fcaf72f3a91c0b0331ced11343d9 -Candidate 1209aee399a2c0b0677d7a4dbedc5f87880d14ab` EXIT0，成熟marked18.1.0解析三份变化Markdown、6个本地链接，原§15锚点保持，Git空白通过；没有重复无变化文档或重跑Program、编译及安装。

### 退回项及准确来源

新增emojibase-data17.0.0和@types/express5.0.6已写入真实package/pnpm锁并已安装，其他importer、overrides和patch政策保持。`node --max-old-space-size=512 .verification/check-runtime-inputs.mjs 1209aee399a2c0b0677d7a4dbedc5f87880d14ab ada4089de6a9fcaf72f3a91c0b0331ced11343d9` 使用成熟YAML/JSON及实际已安装package，确认58项声明与锁/安装一致，但dependencies.lock.json的swa_runtime缺这两项，实际EXIT1；既有check-dependencies的同项检查会拒绝该差异。已通过 `msg_1d787202d146` 交原Worker补齐，不为清单修复重复已完成的原生编译或完整包。

T02首次真实调用因独立pwsh不能按原写法传递ProtectedRoots数组而失败；后继在原PowerShell中直接调用包内脚本，复制及版本输出成功，随后真实安装闭包在@parcel/watcher处报Cannot find module picomatch，install08-exit仍1。独立Get-Item确认candidate08中的@parcel/watcher为指向自身.pnpm的Junction，installed08对应路径已变成普通目录；原installer的dereference=true使模块真实位置和就近依赖图改变。准确[Node24.21.0 cp选项](https://nodejs.org/download/release/v24.21.0/docs/api/fs.html#fspromisescpsrc-dest-options)说明dereference解引用，verbatimSymlinks只跳过链接路径解析；原绝对Junction目标还须随新根重定位。安装须集中保留合法包内链接图并拒绝包外目标，不能逐个增加顶层依赖代替闭包修复。原包、失败安装及日志保留，原Worker负责实现和真实新安装验证。

工作区新服务故障工具的实际readiness/metadata/PID record、Job句柄、image/创建身份/父链及原DaemonClient观察已读，尚无准确提交或真实执行接受。新工具保持独立来源，禁止覆盖candidate08产物后仍按1209记账。全部实际模型、T12故障矩阵、完整成果、精简依赖、人工批准及同安装候选双Worker验收继续未通过；本轮没有正式PR可发布COMMENT review，也没有集成或发布main。

## 准确b624b454的独立安装与两项真实失败

结论：**退回准确 `b624b45418279499887fe0c1583bd7c73c0979e2` 的T02目录保护及Job后代验收；hoisted安装复制与inventory修正按实际范围接受。** 本次Reviewer沿原Task、工作树和两报告接续，当前Dispatch为`ctx_40bf84d10b8c`，终端为`term_6b1b14a1-e6e8-4f57-b394-8bcd79ec1e3e`。实际公开terminal show预览为`GPT-6.1-Sol xhigh fast`，没有子Agent。已完整读取主控重启接续文件，资源解除与既有权限继续适用，普通修复直接交原Worker累计处理。

相对1209的十一文件差异已完整读取。prod deploy只在本次子调用设置node-linker=hoisted，安装复制前新增拒绝候选内reparse point，未改开发目录和全局配置；原Node固定输入、原生编译和许可保持。inventory补齐两项真实依赖。独立`check-runtime-inputs.mjs b624b45418279499887fe0c1583bd7c73c0979e2 1209aee399a2c0b0677d7a4dbedc5f87880d14ab`实际EXIT0，成熟YAML/JSON核58项声明、锁及实际安装一致，其他importer和政策保持，inventory无差异。`cpu_scope=primary_processor_group`明确原GetSystemTimes观察范围，没有增加CPU准入门限。

build10退出原件为0，最终实际输出列出全部四native及准确b624，candidate09/VERSION相同。install09实际完成独立installed09，版本一致；真实Node24.21.0、native API1、空Job及28659路径闭包通过。candidate09-native-origins实际EXIT0，输出的CJS/native来源均在该候选，包含watcher、picomatch及其平台native。上述证据只接受路径、实际模块装载和空Job范围；全部入口及子进程、SQLite和动态消费者的运行闭包继续未验。

**T02目录保护，严重程度高。** install09先实际拒绝占用端口与直接受保护根，随后junction用例返回`storage_failed`并尝试`owned-protected-root-junction/home`，完整日志中的Access is denied及测试异常保持。`privateDataDirectory`在检查重叠时复用`agents.CanonicalRoot`的EvalSymlinks；同一既有projects.go已说明Windows最终junction可能保持原路径。当前路径随后进入MkdirAll，没有在受保护目标任何写入前拒绝。修复须在写入前检查数据根及现存全部父级的真实reparse属性，覆盖根junction、父级junction和不存在叶目录，保持`scope_denied/3`及目标无新增写入。install09没有外层退出文件，重启前异常不推断外层退出码。

**Job后代测试，严重程度高。** candidate09-provider-job实际EXIT1，普通进程fixture在provider-job.test.mjs:81断言成员数时得到5，预期3，不能计为后代退出验证通过。原日志保留进程缺失、成员未确认及两个实际准入观察；它没有保存失败时额外两个成员的完整身份，尚不能判定额外成员来源。原Worker须先记录全部实际成员PID、创建时间、映像及父链，依据真实来源解释与修复；直接放宽计数不能证明全树所属及退出。normal、forced、crash、late-child、parent-loss、两次完整空Job观察及无关所属Job继续保全均保留验收要求。已通过`msg_74688d72d9d4`直接交当前原Worker。

准确提交已包含R06的service_windows.go和runtime-restart.mjs；两份正文、原controller-fault共用函数和service-request入口已读。控制程序核实际readiness/metadata、原daemon PID record与具名Job、持有句柄的安装映像/创建身份/成员及host/server父链，故障只选择serve/host/server并保留daemon。restart驱动按原真实journal、Run/root/Dispatch/session/operation及同库记录观察不重发和资源保留，并明确remainingCases/退出6。该部分当前仅确认源码和既有编译范围，实际故障、恢复证明及清理仍未运行。

`compile-b624.log`与原编译入口已读，四份test-c产物及SOURCE存在，三个传输前函数实际PASS，不扩大为四个安装服务用例通过。Reviewer实际`check-revision.ps1 -Previous 1209aee399a2c0b0677d7a4dbedc5f87880d14ab -Candidate b624b45418279499887fe0c1583bd7c73c0979e2` EXIT0，marked18.1.0核三Markdown、6本地链接，原§15锚点保持，Git空白通过。当前`gh pr list --head swa-cli-headless-20261020`实际返回空列表，没有可提交COMMENT review的正式PR；准确交付后继续执行。认证、精简依赖、T12、普通终端后续输入及同安装候选真实双Worker并发/指定停止/成果仍未通过，本Reviewer未修改实现或启动候选。

## 准确5070073c的目录、Job与SQLite修复

结论：**限定接受 `5070073c904aac6498dc821d6e775130deeb4257` 的十三路径源码及下列局部验证；安装后服务仍待复验。** Win32 GetFileAttributes在数据根创建前检查全部现存父级、home/temp及保护根；daemon-entry原有Job字段经startDaemon传给DaemonServer，严格hello/PID record/内核身份核验保持；SyncDatabase的打开和备份边界使用标准toNamespacedPath。

已完整读取junction-local-01/02脚本、四输出和退出原件。01的home/temp提前命中安装根重叠，已通过`msg_05f9d4897a7a`交原Worker补验。02保持同一本地编译程序，数据/保护目标与程序目录互相独立，四项逐一命中指定reparse路径、scope_denied/3，目标sentinel及拒绝前目录保持，实际EXIT0。准确新候选的T02继续执行。

provider-job11实际额外成员为子/孙Node派生的两个conhost，PID/创建时间/父链与Job对应。新fixture逐项核全部成员，没有从产品Job中过滤console。provider-job12实际EXIT0；本Reviewer用成熟stream-json与JSON独立核四模式的成员、旧身份退出及两次空Job，间隔分别约59/63/56/63毫秒，无关Job root保持。late-child新增Node6020未进入已捕获的native身份表，只确认完整空Job及已记录身份退出；parent-loss旧Job仍unknown，native abort剩余保持。独立命令`node .verification/review-provider-job12.mjs`最终EXIT0，原件为`provider-job12-independent-03.log`。

service09-diagnostic实际host退出1、serve退出6，readiness为delivery_unknown；SQLite备份错误和Job对象校验失败均保留。源码确认daemon-main曾遗漏转交Job字段，507已补齐。backup09-probe以真实profile数据库只读为源，289字符普通路径失败，namespace备份成功；backup09-source-check实际302字符目标经修改后的真实适配器备份并重新只读打开，integrity_check=ok，两命令退出均0。新完整候选仍须验证服务、实际数据库Worker入口和关闭。

实际`check-revision.ps1 -Previous b624b45418279499887fe0c1583bd7c73c0979e2 -Candidate 5070073c904aac6498dc821d6e775130deeb4257` EXIT0：三Markdown、6本地链接、原§15锚点及Git空白通过。原2cc1f64b两hook已由主控`msg_f81a8d207fa9`及`msg_5ee89b7fa1a7`明确完整读取回复，本次唤起确认。其余验收继续按本报告H01—H18及当前任务执行。

### candidate10实际启动：退回重复RPC注册

build11实际EXIT0，candidate10/VERSION为准确507。service10-diagnostic实际host退出1、serve退出6，readiness为delivery_unknown；stack定位原生RPC完整性检查。`node .verification/review-rpc-registration.mjs 5070073c904aac6498dc821d6e775130deeb4257`使用TypeScript AST沿实际七组及跨文件spread展开，得到17项、无缺失，唯一重复为workerList：原WORKER_METHODS已经经release集合包含该方法，host再次显式加入。**退回507的服务启动范围**，通过`msg_b5f23ff6d649`交原Worker移除重复来源，保留原16项白名单和完整性检查；其他局部接受范围保持。

主控新增native显式复用审查，具体五项条件已交原Worker并获`msg_c01841838b60`确认：固定成功源及原产物，逐模块核实际源码/补丁/addon-api/Node输入与编译规则，复制原产物和slotSourceFiles支持资产，新候选实际装载；已知变化只编译受影响模块，未知/失配立即失败，沿既有记录。当前build11未改变，后继准确实现及实际检查另审。原507限定接受hook已获`msg_f4570bfd562e`完整读取回复。

## 准确26036164的RPC去重与native显式复用

**限定接受 `26036164e57acc8267d3727f5ed35a97a74f0de5` 的五文件源码、实际复用及装载范围。** 完整差异已读。独立TypeScript AST按host的实际导入及跨文件spread展开，准确16项、missing/duplicates均空，白名单和严格拒绝保持。复用以candidate10/507、build11原件及原工作目录为明确来源，先检查来源与原字节，再选择已知变更模块；原注入头文件单独比较，仅排除实际生成物，源目标重叠在创建目标前拒绝。原编译记录保留，复制没有增加当前VS发现前置。

已完整读取native-reuse07/08正文、结果及退出：07四包实际复制/四native装载/空Job及两项来源拒绝为0；08真实registry源码副本变化，仅registry重新编译，其余三包复制并实际装载为0。独立检查实际addon-api来源为7.1.1/7.1.0/8.9.2/8.9.2，候选、复制目录和原编译目录一致。早期诊断0及失败分别沿原记录，未计为四包通过。

新candidate11/VERSION为260。Reviewer用该候选Node执行`.verification/review-native-copy-26036164.mjs`实际EXIT0：四个二进制与原n0—n3、源candidate10字节一致，三件ConPTY支持文件一致；四个实际.node均从新根装载，真实空Job创建/查询/关闭通过。build12最终退出原件为0，完整包列明四模块复制、零模块重新编译及准确260，实际原生/28659路径闭包通过。此前独立candidate10原生检查及真实数据库/自产备份只读integrity/fk检查均0，原件在Reviewer忽略目录。

实际`review-rpc-registration.mjs 26036164e57acc8267d3727f5ed35a97a74f0de5`及`check-revision.ps1 -Previous 5070073c904aac6498dc821d6e775130deeb4257 -Candidate 26036164e57acc8267d3727f5ed35a97a74f0de5`均0；marked检查两Markdown、6本地链接及原§15锚点、Git空白通过。typecheck09实际退出0且完整日志为空。bd9e4457两hook已由主控`msg_153bd7a0b360`、`msg_7218be552c8c`完整读取回复。

### 260实际服务与仓库外安装

**限定接受260的完整构建、仓库内实际服务启动/SQLite/关闭及仓库外安装闭包；真实任务继续未验。** service11-diagnostic实际readiness.ok=true，runtime为5720a822-9d0f-40be-be29-9bd767ff4444，host/server为30160/37504，HTTP health实际200。五个实际数据库含profile备份均integrity_check=ok、foreign_key_check空；service与外层退出均0，关闭的两次Job观察仅余serve自身，间隔约101毫秒。运行中和关闭后六个受保护文件的bytes/ACL及公开Orca身份保持；目录通知没有writer PID，不能将观察到的agent-hooks通知归因于本次服务。

install11完成复制与准确版本输出后，原native子检查30秒ETIMEDOUT，实际退出1，原件保留。主控明确批准将该同一安装移到 `C:/Users/DW/AppData/Local/SwarmCli/candidates/26036164`；已完整读取固定路径、现存父级reparse、旧版本、所属进程及ACL检查脚本。实际移位退出0，原源消失、私有ACL保持，新根版本准确260；该根为普通目录，其全部祖先均没有node_modules。新位置原检查实际退出0，四native实际调用、空Job及28659路径闭包通过，没有把第一次超时原因推断为依赖缺失或修改检查时限。仓库外服务已产生实际就绪、health及SQLite原件，关闭与完整功能继续复验。

原260限定接受hook `msg_3931dedc11b0` 已获主控 `msg_ff8a9e31d6e4` 明确读取046fffb5报告、准确源码及原件的回复。认证、CLI完整读写、T12故障恢复、普通终端后续输入及同一最终安装的真实双Worker并发/指定停止/成果仍按H01—H18保留，没有把启动或闭包检查扩大为这些功能通过。

### 八个直接依赖的具体调整条件

主控要求按准确260重新审查原八项纯JavaScript候选。Reviewer实际运行原pnpm11.27.0的`list --prod --depth 100 --json`，并以已安装TypeScript6解析candidate11全部21份实际metafile/成品及createRequire别名；两命令退出0。动态参数逐类检查覆盖相对process-tree、Windows security候选、Pi实际安装配置/catalog和Claude平台CLI；保留真实native、@linear/sdk、emoji JSON、node-pty支持文件、watcher及原生Agent资源。没有仅凭静态输入缺少作删除结论。

| 可从runtime dependencies移至devDependencies的精确版本 | 删除八条直接生产边后仍须由pnpm保留的实际间接版本 | 内联许可 |
| --- | --- | --- |
| zod4.6.5 | 4.6.5，现有Anthropic/OpenAI/MCP等SDK | MIT |
| yaml2.9.1 | Pi coding agent的2.9.0 | ISC |
| diff8.0.4 | Pi coding agent的8.0.4 | BSD-3-Clause |
| smol-toml1.9.1 | 当前生产图没有其他来源 | BSD-3-Clause |
| jsonc-parser3.3.1 | 当前生产图没有其他来源 | MIT |
| @streamparser/json0.0.26 | 当前生产图没有其他来源 | MIT |
| tweetnacl1.0.3 | ssh2→bcrypt-pbkdf的0.14.5 | Unlicense |
| proper-lockfile4.1.2 | Pi coding agent的4.1.2 | MIT |

这八项保留原精确版本，pnpm生成锁中的依赖类别随manifest移动，其他importer、overrides、patches保持。现有dependencies.lock.json合并dependencies/devDependencies，八项inventory版本保持。间接依赖由成熟pnpm正常部署，不手动删除传递包，不推断八个目录都会消失或具体体积减少。

产物须从实际内联输入所属包保存完整版权/LICENSE/NOTICE，并更新现有THIRD_PARTY_SWA指向；覆盖九个内联版本，特别tweetnacl1.0.3及0.14.5，jsonc-parser实际许可文件为LICENSE.md。可直接沿已有构建输入和包元数据复制原文件，安装继续携带全部许可。实际迁移后须核准确manifest/锁/inventory、原frozen-lockfile/依赖检查、新成品及许可，再用新独占候选安装到批准的仓库外位置，核动态来源、实际服务/数据库/故障及同安装真实双Worker成果。当前260产物不覆盖，原native输入未变时沿已接受的显式复用规则处理。详细实施条件及原始图/调用证据在Reviewer自有忽略目录，已通过`msg_07d7ac1bea50`、`msg_2fb8ddaf9f26`交主控与原Worker；本节没有宣布后继实现或精简效果通过。

## 260仓库外动态装载与本人Observe入口

**接受准确260安装的下列实际局部结果，以及本人Observe入口的静态/API范围；人工授权和真实任务结果继续未验。** Reviewer使用该仓库外安装自带Node实际执行`.verification/review-external-install-26036164.mjs`，EXIT0：四个原native及watcher平台native、动态@linear/sdk和emoji JSON全部从该根装载，实际21个装载文件均在安装根内，祖先没有node_modules；空Job创建、查询及关闭通过，模型调用0。原件为`external-install-26036164-independent.log`。已完整读取Worker的installed11-native-stages、refusals及scope实际脚本和原件：实际装载、端口占用/直接保护根拒绝及根、父级加不存在叶、home、temp四类reparse检查均0，保持scope_denied/3、拒绝前目标内容及目录；该范围不代表完整T02或任务通过。

本人入口和配套消费脚本在Worker自有忽略目录中，准确版本如下。两份正文完整读取后，Reviewer仅调用成熟PowerShell Parser，均错误0，没有执行正文。

| 文件 | 字节数 | SHA256 |
| --- | --- | --- |
| human-authorize-installed11.ps1 | 9686 | B9A06E352B131273B1053F5ECEF5766766B4B9127824473B4C789DCA2B2928C4 |
| use-installed11-credential.ps1 | 4578 | FF1D61DA9958DD5F788E024FE49D02E888E51838D39BFBCE2660DAA0C3E00C9B |

流程仅支持Observe。固定公开stats项目根、准确da3bdddb来源及caller=swa-cli-acceptance，真实新项目默认关闭目录和Agent控制；settings与Grant只增加observe。已沿实际sessionGuard、classification/spaces、local_projects及projects代码核cookie/CSRF、UUID/Idempotency-Key、schema/expected_version、settings CAS和space/project/grant/credential字段，沿既有USAGE/API操作。服务独立观察为serve30160、创建UTC2026-10-21T16:32:44.1021696Z、准确安装映像，HTTP39960父链和64506监听归属一致，health实际200；人工确认后bootstrap、每个写请求以及两脚本每次发送凭据的CLI前均重新核上述身份、版本及health。

cookie/CSRF仅驻本人窗口，token使用CurrentUser DPAPI、固定来源entropy、私有DataRoot及当前用户所有权/ACL保存，CreateNew拒绝覆盖；消费侧核来源、根、caller、唯一observe、时效及reparse，token只进入调用进程环境并在finally清理。输出只有实际请求标识、公开结果和凭据文件路径。当前静态接受已通过`msg_c9de61d348a6`、`msg_14302c355962`交主控及原Worker，可由主控请求本人执行；Agent没有创建human session、Grant或token，未把该批准扩大为执行、目录、工具或原生Agent的实际通过。

主控准确3b46b293的四文档及后继3cdf29ce的依赖计划已完整读取，限定接受其当前记录与实施条件；原超时和未知保持。Reviewer实际`check-revision.ps1`分别核0ae50787→3b46b293、3b46b293→3cdf29ce，marked18.1.0检查四文件/46本地链接及一文件/5本地链接，Git空白均0。compile260实际SOURCE=260、四个测试包编译日志为空、三个传输前用例PASS/外层0已读，只接受该范围，不表示四包全部运行通过。原ac950358 hook已获主控`msg_50db746ddf2f`明确完整读取报告和原件回复；八项实施条件已获`msg_4f401ee97ec6`采用。后继依赖实现及本人实际批准继续累计核查。

## 准确780a35e9的依赖分类与完整许可

**限定接受 `780a35e9ec996494f4188f44229d2568bf38b9a2` 的六文件源码、build13完整包及九版本许可检查；仓库外新安装与真实任务继续复验。** 已完整读取全部差异。八个原精确版本移入devDependencies，原bundler条件、external及各入口保持；生产间接依赖继续由pnpm部署。原buildSWARuntime收集实际21份metafile，按真实包名/版本核内联来源，复制全部原许可/notice至安装根third-party-licenses；缺失来源或全文立即失败，原安装复制覆盖该目录。

Reviewer实际`review-dependency-move-780a.mjs`使用成熟YAML/JSON，确认manifest及完整锁仅发生八项预定移动，全部其他字段相同，inventory字节不变；`check-runtime-inputs.mjs 780a35e9ec996494f4188f44229d2568bf38b9a2 26036164e57acc8267d3727f5ed35a97a74f0de5`核58项声明/实际安装/锁及inventory一致，两命令均0。dependencies13实际lock-only、frozen install、原check-dependencies三份完整日志和退出0已读。`check-revision.ps1`准确260→780实际0，marked核三Markdown、6本地链接及原§15锚点，Git空白通过。

build13实际SOURCE与candidate12/VERSION均为准确780，最终退出0；四native沿原candidate10/build11/ZZTfh4复制，built为空，真实Node/native/空Job及28233路径闭包0。Reviewer以candidate12自带Node执行`review-inline-licenses-780a.mjs`实际0：独立沿实际metafile输入确认九个包版本，全部原许可与输出字节一致、文件为普通文件；21个实际JS成品及其输入路径与已审260逐项相同。原件为`inline-licenses-780a-independent.log`。该结果没有改变原安装超时，也没有证明新生产部署的动态路径、实际服务或精简效果。

原260服务和准确本人Observe入口保持，用户执行尚无实际回执；执行、工具、T12及同一最终安装的真实双Worker成果不计通过。07b12530的260及main两hook已获主控`msg_0819f42de142`、`msg_6baaf9b421e2`完整读取回复，本次唤起确认。

### 780实际新安装失败与260普通进程补验

**退回780的install12复制范围，严重程度高：新安装尚不可用。** 已读实际install12入口、完整日志及outer-exit=1。目标为批准的仓库外独占`SwarmCli/candidates/780a35e9`，私有ACL设置后，固定Node24.21.0的原cpSync在该目标根返回EIO/Access is denied，installer第57行抛出；后继闭包及T02分支没有完成。Reviewer只读观察源/目标为普通目录，目标所有者DW、DACL仅当前用户FullControl，根内尚无条目；没有据此推断失败原因。失败目录和全部原件保留，修复交原Worker，不能把此前780源码/build/许可的限定接受扩大为安装通过。

原Run实际问询`msg_b953a635e20e`获`msg_c1e1351069ae`给出准确原件位置。260的provider-job-installed11/result.json、test.log、test-exit.txt已完整读取，真实退出0。Reviewer使用成熟stream-json逐条独立核实际身份/父链、两次完整空Job与两次旧身份退出、无关Job保持，命令`review-provider-job12.mjs <provider-job-installed11> <准确260> <仓库外260安装根>`实际0，原件`provider-job-installed11-independent.log`。四种模式的空Job间隔约98.51/94.30/88.17/80.86毫秒；late-child本次旧身份观察含第六PID44952。parent-loss旧Job保持unknown，原生Agent实际abort仍未验，模型调用0。

dpapi-probe01的完整正文、日志和退出0已读：输入为该固定公开项目真实stats.go字节，CurrentUser实际加密/文件/解密往返及当前用户所有权/ACL通过；没有伪造项目token或调用人工入口。限定接受该公开文件与保护API范围，实际授权凭据交接仍待本人执行。cb0491e7的780限定接受hook已获主控`msg_8cf94974d080`完整读取回复，原Worker已通过`msg_e6a6376c9645`收到安装失败及保留要求。

## 780正常发行安装、实际关闭及5e917bc4局部修复

**限定接受准确780的正常ZIP发行、install13安装闭包、仓库外实际服务及下列关闭结果；限定接受准确 `5e917bc461e62835d384d553756229dbf6f1452e` 的累计源码和局部检查。** 原install12失败及退出1保持。已完整读取TokenFacts实际Win32查询、普通复制探针及发行/安装原件：工作区Node和其子PowerShell实际为Low完整性，原shell为Medium；同一固定Node字节普通复制到仓库外后为Medium，原cpSync实际成功。没有修改二进制、系统完整性设置或全局配置。

正常发行使用成熟.NET ZipFile.CreateFromDirectory/ExtractToDirectory，源为准确780的完整candidate12，ZIP实际410042444字节，目标均预先不存在。提取根为`C:/Users/DW/AppData/Local/SwarmCli/candidates/780a35e9-package13`，原安装脚本将完整包安装到同父目录的`780a35e9-install13`；两根均检查现存父级reparse及私有ACL，版本保持准确780。install13复制、版本、真实native/空Job/28233路径闭包均0；端口占用返回address_in_use/5，直接保护根与四类reparse返回scope_denied/3，拒绝前目标内容与ACL保持。**安装检查的内层及外层退出为6，T03/T13实际消费者仍未验证，不能登记为完整T02通过。** 正常提取和新安装保持独立来源，没有覆盖旧失败目录或将后继源码写入780安装。

Reviewer以新安装自带Node实际执行以下独立检查，均EXIT0。`review-external-install-26036164.mjs <install13> <准确780>`核五个实际native、动态@linear/sdk及emoji JSON，21个实际装载文件均来自新根，祖先没有node_modules，实际空Job创建/查询/关闭通过；`review-inline-licenses-780a.mjs <install13>`核21份实际JS及输入路径与已审260相同，九个版本完整原许可字节一致。原件分别为`external-install13-independent.log`与`installed-inline13-independent.log`，没有用脚本中的NOT_RUN字段推断服务结果。

### 实际服务与所属资源关闭

run-service13的完整入口、stdout/stderr及实际退出0已读。同一780安装使用独立仓库外DataRoot，实际serve/host/server为34392/46628/45740，三者Medium，readiness的runtime为`64c1ceae-d4e5-4b5a-a0ae-f9dd66a12fea`，HTTP health实际200。journal、orchestration、profile、profile备份和workspace五个真实SQLite均integrity_check=ok、foreign_key_check空。EOF后serve实际0，完整Job两次观察仅余serve自身，间隔约100毫秒。运行中和退出后六个受保护文件的字节/ACL及原Orca公开身份保持。安装根ACL只允许当前用户；普通DataRoot的ACL含当前用户、SYSTEM和Administrators，未将其记为凭据专用目录。目录通知没有writer PID，不用于推断写入进程。

原lifecycle14实际来源为780，四种情况及完整原始输出已读；Reviewer执行`review-lifecycle14.mjs`实际0，原件`lifecycle14-independent.log`。立即EOF、启动中EOF、就绪后EOF、启动中stdout读端关闭分别返回130/130/0/6，记录2/10/11/11个实际身份，两次旧身份退出观察间隔264/271/276/279毫秒且剩余均空。非立即退出情况保留两次完整Job观察，间隔至少100毫秒；立即EOF在准入前取消。**该旧stdout关闭输出仍含未知关联字段的空值，只接受实际关闭与资源范围，C09诊断格式继续未接受。** 后继加强断言的driver没有作为这次780执行来源。

provider-job-installed13原件和真实退出0已读，Reviewer用成熟stream-json独立执行`review-provider-job12.mjs <provider-job-installed13> <准确780> <install13>`实际0。普通所属进程fixture的normal/forced/crash/late-child两次空Job间隔约69.80/78.25/75.09/71.92毫秒，初始Node/conhost成员身份与父链、两次已记录身份退出及无关Job保持均通过。late-child额外fixture PID46432没有进入native旧身份表，保留该逐身份观察限制；全Job为空证据独立成立。parent-loss旧Job仍unknown，实际原生Agent abort仍NOT_RUN。本组模型调用0，没有执行本人授权入口。

### 准确f46afc49至5e917bc4的源码范围

已完整读取 `f46afc49e0e23cbb3d4e915ce0fc8c5ff78a8854` 和准确5e的差异。USAGE给出正常ZIP提取和同一PowerShell直接传递ProtectedRoots数组的原安装方式。T03沿原实际Go测试目录及SOURCE比较准确提交的cmd、internal、migrations、contracts、go.mod、go.sum和web/src/api/schema.d.ts，源码或比较失败立即拒绝，保留原编译来源。5e修改cmd后，Worker重新编译全部四个真实测试包、CLI及故障工具，实际Go为go1.27.2 windows/amd64；SOURCE在编译完成和源码复核后才记为5e。完整compile5e917入口、日志、空编译日志及外层0已读，三个传输前用例实际PASS/0，没有扩大为四个测试包全部运行通过。

stop-http保持实际安装映像、创建身份、父链、同Job、真实IPC及持有句柄核验，只向HTTP server句柄取得终止权限。5e在server实际退出后对原daemon同句柄记录alive/exited，unknown立即失败；原stop-service继续要求daemon存活。该后置条件符合原serve关闭Node stdin和空闲daemon退出的实际代码路径，原Worker收到`msg_b1aca3528002`后完成修复。此处接受准确源码和编译，实际HTTP故障结果另验。

outputFailure只输出实际已知且非空的关联字段，保留原R、query提示及动作退出6/7；cli-failure检查拒绝空关联字段。真实cli-local-pipe15使用准确5e编译CLI，在参数错误前关闭实际stdout读端，两例均退出7：有效request-id用例仅输出该已知request_id，无request-id用例没有关联行，均无stream错误。完整脚本、原始输出及外层0已读。**只接受此传输前失败边界，尚未接受同一最终安装完整serve的C09。**

Reviewer实际`check-revision.ps1 -Previous 780a35e9ec996494f4188f44229d2568bf38b9a2 -Candidate 5e917bc461e62835d384d553756229dbf6f1452e` EXIT0，成熟marked18.1.0核三Markdown、6本地链接、原§15锚点及Git空白。

### 历史所属与全新空闲服务检查计划

独立条件清单为Reviewer忽略目录的`http-fault-review-conditions14.md`。完整读取Worker的`http-fault-variants14.md`后，**接受SHA256为 `33488A255DAF210FB300899828DEA86C0174D57F781E90374D82DB53C95DAC8B` 的准确计划实施**，已通过`msg_eebefb5fffa5`交原Worker，通过`msg_452ab9a2e23f`交主控；不据此接受后继实现或故障运行。

计划保留默认live owner和纯inventory，显式历史检查只接受原真实成功envelope、原request/caller_fingerprint/receipt、完整owner及实际同源daemon/Job。实际inventory中的原session须完整身份匹配且isAlive=false，消失则如实记absent，仍存活或不明立即拒绝；历史结果没有当前调用者或恢复授权含义。owner-file、request-show-file、inventory-only、idle-only互斥，实际database必须是已核DataRoot/orchestration.db。

idle只允许driver启动前不存在的独占DataRoot，实际inventory为空且coordinator_runs、tasks、dispatch_contexts、worker_dispatches、worker_terminal_resources、mutation_receipts六表均0。runs只允许原迁移的唯一 `ORCHESTRATION_LEGACY_RUN_ID=run_legacy_local` 默认行：objective为`Legacy orchestration state (inspect only)`、home_database=this_database、consumer_generation=0、legacy=1，四项coordinator绑定均null。已独立读取原常量、contract-constants与migrate-v2-v12的实际插入，并核service13真实数据库同一默认行及六表0；其他行、绑定、缺表或缺字段均拒绝。Go inspection_mode显式区分live/historical-owner/idle，历史及idle仅配stop-http，输出状态逐项核对；纯inventory不能代替idle。C08须记录实际HTTP目标故障、准确7/6及原因、唯一readiness、完整Job与旧身份两次退出和原Orca保全，任意非零不能代替通过。C04/R10仍须真实当前Grant及原lineage退出条件，处理窗口未命中保留INCONCLUSIVE。

主控准确ef49ebf3、732f92c3及 `4dbbb564025734de12331ea024bd1ee1757cc362` 的变化已完整读取，限定接受正常发行安装和当前/历史来源分列、局部结果及未知范围的记录。Reviewer实际`check-revision.ps1 -Previous ef49ebf3896dc1d9c5bbf2ee3ff822d0c790a5c5 -Candidate 4dbbb564025734de12331ea024bd1ee1757cc362` EXIT0，成熟marked18.1.0核三Markdown、44本地链接及Git空白。933c63f5的install12退回及260补验两hook已获主控`msg_cd7e1bff73fe`、`msg_9fbc7e271bdf`明确读取回复。当前正式PR列表为空，尚无COMMENT review对象；没有集成或发布main。

本人Observe尚无实际执行回执；T11、真实执行Grant、模型与工具范围、完整T12及同一最终安装的真实双Worker并发/指定停止/另一Worker继续/成果仍未通过，原S3和真实用户作品授权保持未决定。后继准确63eb7091及新candidate13已收到通知，本节没有登记该实现或新运行通过。

## 准确63eb7091与95efcc28的故障和退出观察

**退回准确63eb的C08退出分类；接受其检查模式源码、安装闭包及下述故障/资源和C09实际范围。** 已完整读取 `63eb7091dbf2b846f5e1984221e229ce18528bf0` 的四文件差异。历史所属与idle显式分支符合准确33488计划，原live、inventory和stop-service条件保持，没有增加生产授权。http-variants16的实际命令记录、完整空日志及两个退出0已读：完整TS6 strict检查和Go工具编译在5e HEAD加两份已审测试源码变更时执行，相同测试正文随后纳入63eb；不将它们称为63eb完整包构建。准确build14/SOURCE及candidate13/VERSION为63eb，完整构建退出0，四native沿507原成功编译显式复制，实际28233路径闭包及空Job0。

主控已批准中间完整普通复制。prepare-package17的完整脚本、原件及退出0已读，外部package17实际28232子项、25485文件与源字节相同，Node和子PowerShell为当前用户Medium。原installer将其安装到`C:/Users/DW/AppData/Local/SwarmCli/candidates/63eb7091-install17`，安装及原生闭包、端口/保护根/四reparse拒绝范围通过，原聚合退出6保留T03/T13。该准备没有完成最终ZIP发行。Reviewer以该安装自带Node执行动态来源和内联许可检查均0，原件`external-install17-independent.log`、`installed-inline17-independent.log`：五native、动态SDK/emoji和空Job通过，九版本完整原许可相同；21份JS仅controller-inspection正文与输入变化，其余20份与260相同。

### C08的准确失败分类

http-idle17完整入口、三项拒绝、readiness、五个SQLite、Token、所持句柄、完整流及最终保全已读。实际serve33496、host37096、server44256、daemon29952属于同一具名Job与安装，runtime为`b515a84e-0bf2-4630-b56c-86245b5df874`。两份实际idle观察均为六表0、唯一未绑定原生legacy默认行及空PTY。故障工具只终止所持server44256句柄，工具退出0；daemon当时alive，随后全部四旧身份退出，两次完整Job只余serve相隔约101毫秒，两次旧身份空观察相隔约798毫秒。三个服务及故障工具均为当前用户Medium；六个保护文件字节/ACL、原Orca公开身份保持。Reviewer实际`review-http-idle17.mjs` EXIT0，原件`http-idle17-independent.log`，仅核这些实际事实。

**该次serve实际为delivery_unknown/6，不能接受其分类。** 已独立核SPEC §21.1引用的方案§6及§7.5：明确失败为7，未知效果为6。准确63eb在server.done后closeOwner取得标准os.ErrClosed即将code改为6，随后await已知exit1无法覆盖6。已读取本机实际Go1.27.2的Wait与StdinPipe源码：Wait结束后自动关闭该parent pipe。原日志只有此重复关闭、已知server失败及成功所属退出，没有其他原生或Job未知日志；该准确失败交原Worker，主控原返修为`msg_3ecbaafbd839`，Reviewer独立条件为`msg_051ee36c1686`。后继已知C08须精确返回runtime_error/7，宽收7或6不能证明分类正确。

### 未知创建时间的完整查询条件

原lifecycle17和c09-only18均因实际行缺creationTimeMs退出1，原件保持。17为root50796/host46128下的powershell43304，18为root50512/host50036下的node48788；两个旧集合中有此前创建时间，缺项样本没有补填该值。18实际已关闭stdout读端，并输出唯一非空runtime_id，但该次驱动没有完成最终退出和Job证据，不能登记完整C09通过。

Reviewer独立读取原native的process.cc和process_worker.cc，确认CreationTime查询失败时该字段不输出，快照或首行查询失败还可能返回空数组。原rows非空判断不能证明完整OS查询成功。Worker补充计划完整版本的SHA256为 `6E3F057AA3A32C65B27F2731F9A457B7E35A3C9542159F1DE18A056B349D164F`，已完整重读并经成熟marked解析0，准确接受通知为`msg_7284cb381877`、`msg_e4f139163990`。保留实际未知行/时间/同帧父链，禁止填旧创建时间或向未知PID发信号；原native身份及完整Job条件保持。最后两次独立成熟CIM查询须成功、完整且非空，全部返回行PID存在并类型有效，本项命中身份字段完整，至少间隔50毫秒。原已知身份必须退出，曾经unknown的PID必须完全不存在；存在任何创建时间的同PID也保持unknown/6。仅记录所需身份字段及查询数量、状态和时间。

### 准确95ef源码与真实局部分支

已完整读取 `95efcc288e32ad8d72d4ef902f32ec130f97fea3` 五文件差异。serve先取得closeError并await；只有标准os.ErrClosed、实际Wait完成、同PID ProcessState确认退出且等待结果明确时才接受重复关闭。非ExitError的Wait错误、实际退出状态未确认、其他关闭错误及原生/Job未知继续为6。该准确实现的Go编译与新安装C08实际结果见下节。

新增observe-process-exits.ps1和serve-lifecycle.mjs保持上述准确计划：完整CIM每行PID在筛选前核验，缺字段使结果保持未知；查询异常保存实际失败后退出6。驱动保留两次native观察，遇缺项单独记录unknown；本项root当前创建身份未核实时也记录未知，不用该行获得新控制权限。独立CIM限15秒，超时只处理实际创建的helper并最多再等待5秒，结果始终未知；所选、未选及所选尚未完成的Case分别记录，任一退出证明不足后停止后续Case。Reviewer实际MJS语法及成熟PowerShell Parser均0，未执行故障注入来伪造CIM失败。

cim19的完整真实脚本、四个输入/输出、进程来源和原句柄Wait原件已读。实际自有pwsh44160/1792608160074存活时返回6，关闭本项stdin并Wait确认0后返回0；实际自有pwsh47180/1792608162686在unknown请求中只携带真实PID，CIM创建时间正常可用，存活返回6、实际退出后返回0。旧known在后两例已实际退出，独立区分了两个拒绝分支。四次两份完整CIM成功、字段完整，间隔约505.40/510.29/522.25/469.32毫秒；未知输入没有冒称原生查询失败。实际CIM查询故障和超时路径继续未运行。

lifecycle19在原63eb安装、新仓库外数据和64508端口完成四种实际情况，原驱动退出0；立即EOF、启动中EOF、就绪后EOF、启动中关闭stdout读端分别130/130/0/6。Reviewer独立`review-lifecycle14.mjs <lifecycle19-outer> <准确63eb> diagnostic cim` EXIT0，原件`lifecycle19-independent.log`：实际2/10/11/11个已知身份两次native退出间隔263/280/264/264毫秒，四组独立CIM间隔约471.97/520.48/444.01/450.38毫秒，查询均成功、全清单非空、匹配行均空。最后一种实际conhost25892缺创建时间样本原样保留，两次CIM均未见该PID，没有声称该样本曾获得完整创建身份。三个有子进程的情况保留两次完整Job只余serve，stdout关闭实际输出失败、唯一非空runtime关联和省略未知字段符合已审C09；六个保护文件及原Orca身份保持。接受这一准确63eb安装的四种关闭及C09诊断范围。

独立`review-source95ef.mjs` EXIT0，原件`source95ef-independent.log`：lifecycle19实际执行的MJS与PS快照，经换行形式规范后与准确95ef的对应源码全文相同，四份CIM原件与真实来源/请求/匹配身份/退出逐项一致。此结果证明测试源码来源和既有63eb局部结果，不表示95ef核心二进制、C04/R10真实接续或原生Agent通过。

主控准确ac4e059ba7ea4f52e8cde4d4dec6cdefa645e1bf、59c7a652a6a16814c0cd198a2b99d9f28650c819的三文档变化已完整读取，按各准确来源限定接受中间安装、实际缺口和6E3F计划记录。Reviewer实际`check-revision.ps1`依次核4db→ac4、ac4→59c7、5e→63eb及63eb→95ef均0；两组管理文档各44本地链接，实施文档各6本地链接及原§15锚点，均由成熟marked18.1.0解析并通过Git空白。110e567a三个hook已获主控`msg_57ad2b4cc887`、`msg_1d987461b0e4`、`msg_493ffdc3792d`明确完整读取回复。

### 95ef真实ZIP安装与C08/C09复验

**限定接受准确95ef的源码、完整构建、正常发行安装、C08已知失败7及四种生命周期/C09实际范围。** compile20完整入口、原件及退出0已读：Go1.27.2在准确干净95ef重新编译四份原测试二进制，四编译日志为空，三个传输前函数实际PASS；SOURCE在完成和干净源码复核后才记95ef。四个真实安装服务消费者继续未运行。build15完整日志和退出0、SOURCE与candidate14/VERSION均为95ef，四native按原507/build11/ZZTfh4成功输入复制，built为空，实际Node/native/空Job及28233路径闭包0。

extract-package20完整入口及实际0已读：成熟.NET ZipFile生成实际410059449字节ZIP，再向预先不存在的普通私有目录正常解包；原固定Node字节/版本保持，实际Node及子进程均当前用户Medium。原installer形成独立`C:/Users/DW/AppData/Local/SwarmCli/candidates/95efcc28-install20`，实际版本95ef；install20日志包含真实native/空Job/闭包0、占用端口拒绝、直接保护根及四reparse拒绝，聚合退出文件为6。该结果没有扩大为完整T02或实际任务通过。Reviewer以这个安装自带Node独立核五native、动态SDK/emoji及实际装载来源，九版本完整原许可和21份成品均符合原范围，命令退出均0；原件`external-install20-independent.log`、`installed-inline20-independent.log`。

http-idle20沿同一实际安装及全新仓库外数据运行，外层退出文件为0，真实serve42212、host40280、server33444、daemon37620，runtime为`ef5610a1-804f-4b24-beae-0b855c1032c6`。完整源码及原件核唯一成功readiness、三项实际模式/错库拒绝、六表0与唯一原生默认行、空PTY、五个真实SQLite及四项Medium Token。故障工具只定向终止所属server33444的同一句柄，实际目标已退出；daemon在该次观察alive，最终全部四个旧身份两次缺席，原Job两次只余serve。实际serve精确返回runtime_error/7，没有重复关闭stdin的未知错误。Reviewer显式要求7执行`review-http-idle17.mjs <http-idle20> <准确95ef> 7 <http-idle20-exit.txt>` EXIT0，原件`http-idle20-independent-02.log`：完整Job观察间隔约101毫秒，旧身份观察约810毫秒，六个保护文件字节/ACL及原Orca公开身份保持。接受本次C08准确分类和所属收尾；原63eb/17实际6失败保持原记录。

lifecycle20在同一个95ef安装及独立仓库外数据完成立即EOF、启动中EOF、就绪后EOF、启动中关闭stdout读端，实际退出130/130/0/6，驱动0，四项OBSERVED，remainingCases空。Reviewer使用该安装Node独立运行`review-lifecycle14.mjs <lifecycle20-outer> <准确95ef> diagnostic cim` EXIT0：已知身份2/10/11/11，两次native退出间隔276/274/267/265毫秒，四组独立完整CIM成功且匹配行空，间隔约411.39/474.93/409.16/537.54毫秒。本次没有创建时间缺项，原19真实缺项及其两次缺PID证明保持。完整stdout/stderr已读：两个取消为唯一失败envelope，正常EOF为唯一成功readiness，stdout关闭为实际输出失败、唯一非空runtime_id、未知字段省略及6；三个有子进程的情况均保留两次完整Job只余serve。原六个保护文件及Orca公开身份保持，实际执行的MJS/PS快照与准确95ef相同。原件`lifecycle20-independent.log`、`source95ef-lifecycle20-independent.log`，均0。

provider-job-installed20的完整真实入口、结果及退出0已读，Reviewer用成熟stream-json独立`review-provider-job12.mjs <provider-job-installed20> <准确95ef> <install20>` EXIT0。普通自有进程normal/forced/crash/late-child完整空Job两次间隔约66.50/63.34/74.04/71.84毫秒，实际初始成员含两个conhost，所属父链和已捕获身份观察、无关Job保持符合原范围。late-child额外PID604未进入已捕获的native创建身份表，保留该逐身份观察限制；parent-loss旧Job仍unknown，实际原生Agent abort仍NOT_RUN。原件`provider-job-installed20-independent.log`。本节实际模型调用0，没有创建本人session、Grant或token。

主控后继准确 `f0fb667e4820a4cd486226b7f6c7d58c01affbfb` 三文档差异已完整读取，接受其95ef源码、原63eb生命周期及当时待验新C08的来源记录。Reviewer实际59c7→f0fb的成熟marked/Git空白检查0，三文件44本地链接，首装和当前审查记录按实际更新。最新正式PR查询为空，尚无可发布COMMENT review的对象。

本人Observe、T11、真实执行Grant/模型/工具、完整T03/T12与同一最终发行安装的真实双Worker并发/指定停止/另一Worker继续/完整成果仍未通过。CIM查询故障及超时没有实际触发，本轮只接受已描述的源码和真实运行范围。原S3和真实用户作品授权保持未决定，没有合并或发布main。

## 同一95ef安装的本人Observe入口

**原22的公开身份/API核查结果保留，两份人工/消费版本已撤销；当前本人入口为后文准确23版本。** 本轮只执行成熟PowerShell Parser、文件摘要及共享函数的只读`Assert-SwaService`，退出0；没有执行人工授权正文、创建human session、Grant或token。下表两份原人工/消费正文已分别保留为Worker忽略目录`date-fix23/human-authorize-before.ps1`、`date-fix23/use-credential-before.ps1`，共享身份函数保持；三份全文均已完整读取并独立核摘要。

| 文件 | 字节 | SHA256 |
| --- | ---: | --- |
| human-authorize-installed22.ps1 | 8007 | E0CF0C3CC91648B00E14D4442FCEC03AD718D8F6B09FE48495ACBFD49CF72E67 |
| use-installed22-credential.ps1 | 2924 | B9792BE393874B1F16EC7CF882D5A2A90911521808C47DF1EE31FF1E20E0BC3C |
| assert-installed22-service.ps1 | 3044 | FD5FA88CC2179AEF2BB97E86F28D76E4AAFBC44488540BF5A458C10E6659CD0D |

服务复用准确95ef的同一个`95efcc28-install20`，DataRoot为仓库外`95efcc28-human22`，endpoint为回环64509。实际serve37764创建UTC2026-10-21T18:52:06.3683428Z，host44492创建UTC18:52:07.3998766，server48980创建UTC18:52:08.6552096，两者父PID均37764；runtime为`cc80c07c-c8bf-47ff-965e-6177bd2a7c69`，metadata.startedAt为1792608728394。run-service22完整入口、readiness、实际身份与三项Medium Token已读，五份真实SQLite均integrity_check=ok/foreign_key_check空，启动前及运行中六文件字节/ACL和原Orca公开身份保持。服务维持所属stdin，尚未执行收尾，不登记其最终退出。原human21脚本路径比较失败及原服务正常关闭的记录由Worker保留。

两个入口共用准确3044字节身份函数，在人工提示之后、每个实际mutation及每次status/capabilities前重新核安装/数据父级reparse、三个完整映像路径与精确创建时间、实际父链、私有runtime/startedAt、唯一监听所属、版本和health。Reviewer独立实际执行该只读函数0；固定公开项目的origin及HEAD实际为原astrocyte-s3-public-stats和准确da3bdddb085808863090a05f60cc1c64415039cf，API/App来源保持已审范围。

人工入口仅接受Stage=Observe，要求本人明确输入APPROVE OBSERVE；实际session cookie/CSRF、固定公开项目登记、仅observe设置/Grant/token均由本人发起。原request_id/Idempotency-Key、设置revision和未知创建保全保持。凭据只以CurrentUser DPAPI写入该DataRoot下的当前用户所有且仅当前用户允许访问的目录，CreateNew拒绝覆盖，token不输出；消费侧核固定来源/endpoint/DataRoot/project root/caller、唯一observe和到期时间，再进行实际只读CLI观察。三份准确入口已通过`msg_99552e5115da`、`msg_f4d1abea92ab`交主控及原Worker；原260入口与服务保持独立。

Worker的缺凭据消费原件实际退出1，错误停在文件不存在检查，未调用CLI；Reviewer实际观察human-approval目录仍不存在。三份Parser均错误0，真实人工授权及新凭据消费尚未运行，原公开源码DPAPI往返结果不能替代本人批准。T11、Execute、工具、真实模型及最终同安装双Worker/完整T12仍保持未决定或未验。

ec225632的三份hook已获主控`msg_26fdd5701816`、`msg_eca4144b06b0`、`msg_e2a90496c96e`明确完整读取报告、准确源码及原件的回复，本次唤起确认。

### 到期时间类型边界的独立复核

**旧E0CF0C3C与B9792BE3两份本人入口的日期处理已退回并撤销；FD5FA88C公开身份函数及95ef程序的既有接受范围保持。** 旧260关闭补充检查的实际输入揭示PowerShell日期边界：成熟ConvertFrom-Json将原UTC日期转为System.DateTime/KindUtc，随后DateTimeOffset.Parse接收该对象时先隐式转为不带UTC和毫秒的字符串。本机+08环境实际serve时间应为1792600364102，却得到1792571564000；owner应为1792600340610，却得到1792571540000。原始JSON的完整UTC字符串经成熟InvariantCulture解析，与直接DateTimeOffset类型转换所得毫秒逐项相同。Reviewer实际`pwsh -NoProfile -File .verification/review-public-utc.ps1` EXIT0，原件`public-utc-review-02.log`，输入均来自真实已保存的公开进程时间，没有生成凭据或人类会话。

本人消费脚本对ConvertFrom-Json得到的expires_at使用同一Parse写法；原App的IssueProjectAgentToken明确签发一小时有效期。在此实际JSON日期对象类型下，到期判断会提前八小时失效；不能将原三Parser/公开身份检查替代该类型边界验证。两份入口交原Worker按成熟日期值与完整字符串的正确转换返修，新版本需重新核准确SHA及实际日期来源；人工HTTP回执和真实新凭据消费仍未运行。原入口及首次关闭补充原件保留，本轮没有重操作已关闭服务。主控已通过`msg_1825382d6257`及原Run的`msg_e3564a45456f`收到需替换本人待办的通知。

## 准确23本人入口、UTC补验与旧260收尾

**接受下列准确23两文件的仅Observe静态/API、日期类型和实际公开身份范围；尚无真实人工签发及DPAPI消费的验证结果。** 原Worker同Task以新Dispatch `ctx_4c001fda773c`接续，准确补充计划SHA256为 `7DF6314A9686702CE01C5A9D4E8A9E35E6729322AA6567808A9800D252215BC0`，已完整读取并经成熟marked解析0，实施接受通知为`msg_f995d676d2b6`。新两份全文已完整读取，对照完整保留的E0CF/B979原件，各仅修改一处expires_at为显式DateTimeOffset值转换；原API、CurrentUser DPAPI、私有ACL、真实项目/caller/唯一observe和精确服务检查保持。

| 当前文件 | 字节 | SHA256 |
| --- | ---: | --- |
| human-authorize-installed23.ps1 | 8000 | A427BE2BAD35E7AF37D823419B599C3E02931F4EF36749247E0F7E47EC177C1B |
| use-installed23-credential.ps1 | 2917 | 05C9F245B6E8DF0FC6B33C5431BE99307C4E7436D13AB1C8EF56407B5712F055 |
| assert-installed22-service.ps1 | 3044 | FD5FA88CC2179AEF2BB97E86F28D76E4AAFBC44488540BF5A458C10E6659CD0D |

旧22两个原路径现均为235字节的参数声明后立即throw，真实执行分别拒绝1；原8007/2924字节全文及原完整摘要保持在date-fix23，不覆盖首份CIM请求或结果。新23消费在缺少真实凭据时实际拒绝1，停在文件检查，human-approval前后均不存在。verify-entries23完整入口和原件已读，五个文件Parser通过，FD5不变，前后仅执行实际只读身份检查0，原服务5187、64509、runtime及三个进程创建身份保持；没有执行新人工正文。

verify-types24完整脚本及全部原件已读：使用原serve/owner完整UTC，以及同一次真实时钟的+08:00和UTC两种序列化，共四份实际来源；每份核原字符串、JSON实际DateTime(Utc或Local)、DateTimeOffset三种输入。12项完整UTC ticks与Unix毫秒逐项一致，保留类型、Kind和offset。Reviewer独立`pwsh -NoProfile -File .verification/review-human23.ps1` EXIT0，采用成熟JsonDocument/ConvertFrom-Json和Int64核全部100纳秒ticks，避免大整数精度损失；同时核新SHA、保留原件、Parser、三个真实拒绝及当前实际服务，原件`human23-independent.log`。没有构造token、expiry、human或模型回执，也没有把这些日期检查计作真实授权消费。

旧260收尾的完整precheck、原owner记录、service-exit=0、service-stderr及最终Orca保全已读。主控准确授权`msg_493e4d2f080c`仅在无human目录且原serve/owner身份核实后，向原工具59063的Read-Host输入Enter；沿原stdin及原持有进程Wait关闭。实际两次完整Job只余serve间隔约100毫秒，六个保护文件字节/ACL及原Orca身份保持，旧数据和原11两个入口摘要不变。最终旧服务精确补充使用date-fix23/v2请求：serve30160/1792600364102、owner33432/1792600340610，host44232/server39960只按真实PID缺席核验，未冒称取得它们的创建时间。成熟观察器两次完整CIM各410项，全部PID有效，匹配行均空，间隔约469.64毫秒，实际退出0；独立同脚本逐项核0。接受该真实旧服务收尾范围，首次时间输入与原结果保留，未再次控制已关闭服务；新human22公开身份独立检查保持0。

产品准确 `c851e615d062e39be60753fc02d11a86ca7241a3` 的两份文档变化已完整读取，覆盖95ef后继8c60、57fe及当前日期/收尾记录。相对95ef只有两份文档变化，程序及安装保持；成熟Markdown与Git空白检查分别95ef→57fe、57fe→c851均0，各6本地链接及原§15锚点保持。主控f0fb→4aee四文档及4aee→d5fca的暂停/恢复记录已完整读取；两组成熟检查均0，各89本地链接，原本人待执行、T11和执行权限保持未决定或未验。暂停期的记录以其准确提交为范围，当前23恢复待办由主控维护。

69d6b062的退回hook已获主控`msg_54866872354d`明确完整读取回复。新23技术接受已通过`msg_783d990b16b4`、`msg_d1e7a303fcc4`交主控和原Worker。当前目录与凭据仍不存在，实际本人回执、观察消费、T11/执行Grant、真实Agent/模型/工具、完整T03/T12和同一最终安装的真实双Worker/指定停止/另一Worker继续/完整成果保持未验；未合并或发布产品main。

### 本阶段最终记录

**限定接受产品记录 `c6d95473796290f96c213513a70d49e3b546f904` 与主控记录 `ebb16d18d0fbfa627ed169c59cbf3a379442db79`。** 两份产品文档和四份管理文档的准确差异均完整读取：当前命令为准确human-authorize-installed23.ps1 -Stage Observe，旧22已撤销；产品二进制仍为95ef同一install20，human22/64509与工具5187保留。新23日期验证、准确v2和旧服务收尾按实际范围记录，原Worker本阶段因人工阻塞failed/retained，后续同Task/树/终端使用新Dispatch接续。QUESTIONS保留本人回执及旧命令反馈未确认、T11和执行范围未定，没有把S3或真实用户作品权限改为通过。

Reviewer实际`check-revision.ps1`核c851→c6d及d5→ebb均EXIT0；成熟marked18.1.0分别解析两文档/6本地链接及四文档/89本地链接，Git空白通过。准确23全文、真实UTC类型/时钟、精确CIM及公开服务既有独立结果保持，没有重做构建或模型检查。最新GitHub正式CLI PR查询为空，本阶段无可提交COMMENT review的对象；M0原正式审查及既有合并范围保持。

beadc233三份hook已获主控`msg_77bc2ab06d08`、`msg_1668d99d70a9`、`msg_1f695caf66e8`明确完整读取回复。主控已再次发送新23本人请求，仍没有实际成功回执或凭据；原旧命令是否执行过保持待反馈。本阶段独立审查记录完成，程序整体及真实Agent/模型、工具、完整T03/T12、同一最终安装双Worker及成果继续未验，产品main未发布。

## 会话批准后的仅Observe委托复验

本阶段Reviewer为`task_9ae5a1337e1d / ctx_5f4446e340a0`，沿原Orca/Codex会话和本报告接续，没有新模型或子Agent。此前`ctx_40bf84d10b8c`已结算，独立只读咨询`msg_1b2765afea53`没有使用其生命周期；失败的readiness尝试也不承担本阶段结果。当前Worker为同模块原树的`ctx_fb2c13c28d40`。

用户已在当前会话明确`APPROVE OBSERVE`，禁止Aliyun key、Aliyun Agent和百炼代理，允许阶跃星辰及其他Agent测试。接管时重读主控QUESTIONS首节、HANDOFF和STATUS，该决定当时位于主控未提交工作树、HEAD为`ebb16d18d0fbfa627ed169c59cbf3a379442db79`；后继准确记录见本阶段末节。Observe批准已确认，真实凭据生成与消费按下文结果验收；执行Grant、工具范围、T11及原S3未决项分别保持。

独立条件在读取本轮Worker计划前已由上述咨询根据真实源码形成。实际`session.go`的bootstrap接受普通本机无Origin请求，服务生成随机cookie/CSRF并固定返回`local-human/human`；middleware依据实际cookie、有效期和CSRF建立Principal，未校验本人TTY。项目设置、Grant和token仍由App的human检查处理，正文不能声明Actor，Agent Bearer不能调用管理路由。现有human会话具有12小时管理能力，没有服务端仅Observe委托会话；本次固定准备流程和最终一小时项目token/Grant限定获准范围。记录须区分真实用户会话批准、实际Agent执行者和服务返回主体，不能将其登记为本人TTY或T11通过。

| 条件 | 正确行为及必要验证 |
| --- | --- |
| O01 授权与执行者 | 原用户消息及主控真实转达是批准来源；新入口明确本次会话委托和原Worker身份，原23保持原件，不调用Read-Host、不注入TTY、不填写Actor或自造cookie/token。 |
| O02 准确安装与公开项目 | 使用同一95ef/install20、human22/64509、原5187及FD5精确身份；每次效果前核三进程创建身份、父子/监听/runtime，固定公开仓库HEAD和origin，原Orca、账号及保护文件保全。 |
| O03 真实准备与持久范围 | 真实session/cookie/CSRF只用于固定空间、项目、仅observe设置和Grant及token五项准备；核真实响应及SQLite的project/root/caller/revision、唯一observe、无执行/工具/目录及外发扩大，不能仅凭HTTP回显宣布持久成功。 |
| O04 凭据与日期 | 服务真实签发token并保存摘要/实际到期；成熟DateTimeOffset转换、CurrentUser DPAPI、私有ACL、无reparse、CreateNew及finally清理保持，秘密不进入参数、日志、提交、消息或模型输入。 |
| O05 成功消费 | 新入口及原consumer23分别用实际项目token取得同一候选status/capabilities，真实退出和来源完整；不创建模型Worker或调用模型，发现/配置/安装/可启动与实际调用分开判断。 |
| O06 真实拒绝与效果 | 用真实Observe token核控制动作、人类写端点拒绝；缺cookie、缺CSRF和跨源请求实际拒绝。对照实际持久任务/Worker、原生会话及所属进程，确认拒绝先于模型和Worker效果，不伪造有效凭据或故障。 |
| O07 部分效果与重复 | 原目录/凭据/部分状态存在时停止；写前保存非秘密request_id及路径，未知立即核原请求和实际状态，不盲重试。generic commandHandler只核元字段，RegisterProject每次生成UUID，不能由Idempotency-Key推断持久去重。 |
| O08 Agent与未验范围 | 只读真实配置并使用成熟格式库，输出来源类型及非秘密模型/目标信息；Aliyun调用为零，StepFun或其他Agent仍按实际注册和既定NativeLaunch合同验证。工具、执行权限及缺失原生调用不得由Observe或配置存在替代。 |

准确`EFD2DCD35C8B051A91B90F4226529B07E6574DA0084E0AB113EA3E69D78E7C56`的`delegated-observe24-plan.md`已完整读取。O01/O02/O04/O05/O07及仅Observe边界具备，O03允许API结果替代持久核验，O06未列实际拒绝及无副作用检查；本版本计划退回这些验证范围，已通过`msg_6a4fcd0750f3`交原Worker、`msg_c729e58ab00a`交主控。未要求构建、额外模型窗口或伪造错误；补充版本按准确摘要重新审核，客户端正文及真实授权尚未执行。

**接受完整后继计划`4B5279939675B4C8E062A8854F1E803DFF9A89F4957AA6EC47A437A2C92F43B9`的仅Observe技术及必要验证范围。** 逐项对照O01—O08：五项真实403明确为无cookie读取、真实cookie缺CSRF写入、不同loopback Origin读取、真实Observe token的合法controller.open和人类Grant端点拒绝；后两项沿实际公开root和真实request_id，不造原生对象。最终真实SQLite须核root、settings r2、唯一observe、控制/目录false、工具/子目录为空、外发CLI空，以及同project/caller的Grant动作和实际expiry；准备及拒绝前后核space/project/Grant、原生session、关联、journal和全部工作表，只允许已批准三种记录新增。12小时宽cookie、未知停止、准备写无持久去重、原账号/Orca保全和秘密范围均明确，没有多余的构建或模型检查。完整后继及准确摘要已读取，成熟marked18.1.0解析EXIT0；主控`msg_58a2d16539f2`明确采用补充后连续实施范围，Worker实施通知为`msg_700b286bdd14`。本接受只覆盖计划，客户端准确源码、真实授权与消费结果仍须独立复验。

### 准备检查失败与已知零效果恢复

两次客户端及helper全文和实际原件已独立读取。24客户端`D74E3C73E7BBE6BE7EC28E3BAF9B56684DA4DECC54BD9C0059D8CD397ED182A5`、helper`BCA50FC85BB4B35252522909BCEAC12EEECA0CB2506BB77F2F15C0066F366BBF`实际退出1于预检，误用不存在的coordinator_tasks；未进入会话或准备写入。25客户端`BE4E9586474FA4437C534050A398356C3C9911F92ACBECEB43EA8CBB42158D39`、helper`2E2492A8ECE7F9E2364E9B7F58DEF119E0CCCEB711CB236E910599B23ACB0738`改用真实tasks后，取得真实会话并通过前三项403，在首个空间准备请求`1182992b-8b77-4d47-adbe-05e53eb70676`停止1；其原始HTTP状态未单独保存。实际failure、原请求和两份源码快照保留，两次Orca保全均通过；这两次运行未接受为授权消费成功。

Reviewer实际以安装Node24.21.0和成熟`node:sqlite`的readOnly逐项读取三库schema及计数，`read-observe-schema.mjs`实际0；原件`observe-schema-independent.log`及首次准备写后再次读取的`observe-schema-post25.log`。真实表为tasks；空间、项目、Grant、session和执行关联均0，orchestration保留原runs1及原ledger1、任务/Worker/receipts0，journal全部工作表0。此范围只读，没有取得human会话或读取凭据。当前human目录实际0项，保护ACL的owner及唯一Allow主体均为当前DW。

Worker的`diagnose-observe25.ps1`全文和实际退出0原件已读：同一真实服务签发诊断cookie，仅GET重现WebRequestSession保留Origin64510，跨源及后续无Origin参数请求分别403/403，显式删除该Header后200；没有准备写或模型。其真实三库计数对比与独立零效果观察一致。该诊断支持客户端清除负向测试Header；没有重发原写请求或以诊断成功替代完整授权。

**接受恢复补充计划`328571126C4116B32572EA23DF2E30205451C9086D12FEC2389BEE609901B578`。** 完整新增条款及保留的准确4B原件已读，成熟marked解析0；26须在每次效果前重新核FD5、25 register失败及唯一原R、当前三库无准备/工作效果、原目录空且私有、无真实凭据。保留原目录和全部失败原件，禁止首次重建或覆盖，任何已有或未知效果立即停止；新独占证据目录、真实新请求和cookie沿同一批准执行完整五项拒绝、持久范围、DPAPI和保全。请求保存实际HTTP状态或网络未知及安全错误码，秘密保持私有。Worker接受通知为`msg_724ce658f6f2`，26准确源码和实际结果见下节。e5a0a429两份计划hook已获主控`msg_ae872f18a18c`、`msg_662535eb1edb`明确完整读取回复。

### 真实Observe26与独立消费结果

**限定接受准确恢复客户端`655799C350501C6D0659B2BF343C4FDE0F3F30FF239A4E97F9F0267D31765697`及helper`2E2492A8ECE7F9E2364E9B7F58DEF119E0CCCEB711CB236E910599B23ACB0738`的真实Observe26范围。** 两份完整执行源码及快照已读；恢复先核原25唯一失败请求、真实三库零效果和原空私有目录，清除真实跨源测试Origin后才准备写。原目录、全部24/25原件及准确4B计划保留，原23人工/consumer及FD5全文摘要分别保持A427BE2B、05C9F245、FD5FA88C；没有修改95ef程序或安装包。

真实26退出0，五个唯一准备request_id的响应依次201/200/200/200/200，无自动重试。实际服务返回`local-human/human`，批准源为原用户会话`msg_a536f02a7294`、执行者为原Worker Agent；新空间`MNHPWL2X74QSIZMIYBTAYL3RLO`、项目`46ceef7b-0be3-4a4b-8e74-575fd327b31c`属于原固定公开仓库。真实SQLite设置r2仅observe、目录/控制false、子目录/工具为空、引用扩展false、history_roots空、外发CLI空；唯一该caller Grant仅observe、未撤销，服务真实expiry为`2026-10-21T20:51:16.4468291Z`，私有token摘要存在且未输出。真实CurrentUser DPAPI凭据写入原独占文件，原consumer23实际消费退出0；没有把服务端12小时会话登记为已撤销。

五项真实HTTP均403/scope_denied，原件逐项核无cookie、缺CSRF、不同Origin、真实Observe token的合法controller.open及人类Grant路由拒绝。workspace指定space/project/Grant各增加1，session/关联不变；原生工作表保留原runs1，其余任务/Dispatch/Worker/receipts0，journal全工作表不变，未创建模型Worker。Orca前后公开身份及六保护文件字节/ACL均通过；目录通知继续只证明对应文件夹观察，不能推导写入者PID。

Reviewer实际独立运行原`use-installed23-credential.ps1 -Stage Observe`，EXIT0，原件`observe26-consumer-independent.log`；没有取得管理cookie或签发Grant。再以安装Node24.21.0执行自有`review-observe26.mjs`，直接readOnly核实际项目/Grant/空间、完整设置及真实到期，并核准确源码摘要、全部五请求状态、五拒绝、独立及原消费的完整JSON结果和保全原件，EXIT0，原件`observe26-independent-03.log`；JSON流使用已安装成熟stream-json解析。两次消费均指向同一95ef、runtime `cc80c07c-c8bf-47ff-965e-6177bd2a7c69`及原authority；这一范围真实授权和观察消费通过，Worker通知为`msg_7b85c8d32daa`。

真实capabilities返回七个注册Agent；五项实际命令/版本检测为available，另外两项unavailable，全部launchable仍unknown。Pi配置未知；Codex和Claude的configured字段分别显示gpt-6.1-sol/openai、qwen3.7-max/anthropic。该字段没有证明实际供应端点许可，Claude实际调用前须核真实来源并排除Aliyun。产品`agent-discovery.ts`的选择路径、NativeLaunchPolicy及launch admission目前只接受Pi1.0.1的已审无工具/API key合同；其他Agent具备注册或配置观察不表示真实调用已实现。独立读取的agent-sources25原件及成熟JSON/TOML reader仅覆盖其明确来源，未发现匹配StepFun条目，不能推断整机没有StepFun配置。真实Agent/模型、执行Grant、T11、完整T03/T12、同候选双Worker和成果仍未验，原工具及S3范围保持。45661024两份hook已获主控`msg_cb1904c65a6c`、`msg_bae6fd3b382d`完整读取回复。

### 读取custom28计划前形成的条件

依据H12/H18、SPEC §21.1.5及实际安装的Pi1.0.1，独立读取当前`native-pi-policy.ts`、原ModelConfig、provider-composer和ModelRegistry。原composer允许base为空，以models.json构造模型，stored credential走原API key解析；没有inherited或models key且没有OAuth时仍生成API key handler。实际模型的api/baseUrl沿模型、provider、默认值顺序合成，modelOverrides最后应用，调用按有效api进入原getApiProvider。当前产品在composer之前要求唯一builtin handler，阻断这一真实原生custom路径。以下条件形成后才读取Worker的custom28方案。

| 条件 | 必要行为与验证边界 |
| --- | --- |
| P01 原生组合来源 | builtin与custom明确区分，保留builtin原行为；custom使用实际1.0.1 ModelConfig及原composer、明确空base和无extension，唯一所选provider/model须来自真实源，不增加虚构注册或另一SDK。 |
| P02 真实目标 | 完整组合后的provider/model/API/baseUrl与批准一致；新增custom范围限定直接StepFun HTTPS及已审API，拒绝任意目标、用户信息、查询/片段、别名网关、认证Header、fallback/routing和非空sampling，严禁Aliyun调用。 |
| P03 真实认证及保全 | 使用原stored literal api_key、空credential env及既有原生literal分类；无models key/OAuth/refresh/命令或环境解引用。原账号只读，私有投影只含所选源，真实源和私有文件版本在每次效果前复核，秘密不进入结果或模型上下文。 |
| P04 生产消费者 | 原NativeLaunchPolicy、实际AgentSelection、ToolsOff、私有账号/DACL、Job、首次/后续输入准入及完整成果消费者均保持一致；当前项目Grant和工具/目录/外发批准仍逐次核实，Observe不能替代执行Grant。 |
| P05 拒绝与回归 | 核缺源/空认证/重复或缺少模型、未审字段/API/目标、headers/routing/sampling和源变更的真实边界；不伪造key、成功CLI或故障协议，真实故障副本只证明其命中的拒绝，原builtin回归按现有真实来源记录。 |
| P06 分层证据 | 语法/类型、原composer来源和真实空认证拒绝分别记录；没有真实获准key时custom正向保持未验，不能从handler或配置对象生成推导可调用。取得实际源后核真实握手/输入/响应/目标及原账号保全，现有未知不改为通过。 |
| P07 同候选验收 | 计划覆盖适用构建、准确产物及正常安装；不混改已验95ef安装。最终仍在同一候选完成两个真实Worker并发、指定停止、另一Worker继续、完整成果和所属资源条件，纯验证客户端改动复用有效原结果。 |

**限定技术接受`native-pi-custom28-plan.md`准确SHA256 `4E698ACC7EF42E9D667CD2D31187F0BDED00F7A460D4F5CD95B84DF323E07514`，并纳入`msg_48f93338842d`的实际API源码身份补充。** 完整计划对照P01—P07覆盖唯一真实custom model、原composer空base/无extension、stored literal key及空env、有效openai-completions与StepFun准确HTTPS目标、既有字段/认证/routing/sampling拒绝、全部私有账号/ToolsOff/Job/输入和成果消费者。原DTO及生产权限不改变；原默认空auth可先验证，真实来源及正向模型缺失保持NOT_RUN。故障副本只验证命中拒绝，最终新候选及同安装双Worker完整验收仍保留，没有多余的文档重构建。计划成熟marked18.1.0解析0，Worker接受通知为`msg_d0aa56dfb57a`；实现及真实调用尚待准确复验。

已独立读取实际1.0.1 provider-composer及ModelConfig，完整文件SHA256分别为DBA201A60EC5B0832DC52CF3D9736C4D470169150B16FCAF989E43427B9FFD5B、68C447FF2EF729331713E044DAE41B7416DBAFBF96C83773308E2EB9C3F86107；原ReadOnlyAuthStorage/private账户消费者保持前述读取范围。Worker `read-native-api28.mjs`全文及实际0原件已读：同版compat取得原openai-completions handler，stream/streamSimple存在，未调用stream、auth resolve或模型；三份实际compat/lazy/API实现的身份前后相同。这些生产来源须纳入效果前重核，handler存在不表示实际调用通过。

独立读取[StepFun全球官方首页](https://platform.stepfun.ai/)，其chat.completions示例采用api.stepfun.ai/v1；[国内官方检索内容](https://platform.stepfun.com/)采用api.stepfun.com/v1，国内首页及JSON Mode页面的直接读取均超时。计划须按实际账号及明确来源再次核对应文档，不能推断跨地区key或模型可用。Firecrawl命令未安装，本次改用网页读取工具，未安装依赖或调用模型。既有私有环境的系统proxy白名单保持；本次目标检查指供应端点，未承诺关闭或修改宿主网络代理。

### 本轮准确记录范围

**限定接受产品`8cd402ed84d55bb10a620895962b0d4d1c9f73a3`两份文档和主线`b5303ee5b5db9caefd726573542a70de762b071e`四份管理文档的记录范围。** 准确差异均已完整读取，记录真实Observe批准、实际Agent执行者/服务主体、26和独立消费、准确expiry、24/25失败、七注册与五安装、全部launchable未知及NativeLaunch缺口；未将执行Grant、T11、S3或完整成果改为通过。产品相对c6仅文档变化，原95ef程序和安装保持；主线记录已完成实际授权消费，原23作为保留的人工入口。

Reviewer实际`check-revision.ps1`分别核c6→8cd及ebb→b53，EXIT0；成熟marked分别解析两份/6本地链接和四份/89本地链接，原§15锚点保持、Git空白通过。OpenCode/StepFun来源条目按所述读取范围使用，没有以局部匹配为空推断整机无配置；后续具体配置路径仍待用户提供。bd55febd真实Observe hook已获主控`msg_0e5f304fabb5`明确完整读取回复，不重复实际凭据消费或完整构建。

### 准确f2ab4fd9静态实现及真实空认证边界

**限定接受`f2ab4fd9155f4f4bb85c74845f3f3ef5eb607caa`的静态实现、相关记录及实际空认证拒绝范围；custom完整正向证据不足。** 相对8cd的三份准确差异已完整读取，唯一生产修改为native-pi-policy。逐项核P01—P07：builtin重复及无API key handler拒绝保持；custom须有真实models源和唯一所选model，沿原composer的空base/无extension组成有效原生API key provider，modelOverrides不能补足缺失模型。最终custom API限定openai-completions，URL成熟解析后核准确StepFun HTTPS hostname、默认端口、/v1或/v1/及原userinfo/query/fragment拒绝；原字段/Headers/compat/routing/sampling和账号literal/空env检查保持。原DTO、私有账号消费者、工具关闭、Job及输入/成果代码未改。

三份实际compat/lazy/openai-completions实现加入sources，读取前后及assertNativePiSnapshotCurrent均逐项核真实路径与stat版本；custom取得实际registry handler后核api及两stream函数，未新增自有handler或SDK。AgentSelection的原configuration_version包含完整sources，新来源仍进入原选择比较；没有用新注册观察替代旧Dispatch绑定或当前Grant。

完整`check-custom28.ps1`及SOURCE/程序diff、真实退出和日志已读：来源为8cd加与准确f2一致的程序差异。TypeScript实际6.0.3、8192MiB、原tsconfig strict/noEmit/skipLibCheck=false，真实strict退出0且完整日志为空；architecture退出0，完整日志确认47个Go package的现有边界。Reviewer没有重复完整构建。实际`check-revision.ps1`核8cd→f2退出0，成熟Markdown两份/6本地链接、原§15锚点及Git空白通过。

empty-native-account28在所读默认配置找不到非禁provider的单独model定义，实际退出1，未调用policy；原脚本和日志保留。29完整入口及实际0原件已独立读：从真实同版catalog取得openai/gpt-4，明确只作缺credential的显式负向输入，不声明其已配置或已被选中执行。真实auth为{}，实际policy返回native_api_key_source_unreviewed；原三源stat保持、实际项目的private账户前后不存在，没有启动CLI、Worker或模型。29在认证入口结束，未覆盖新增custom组合、目标、正向认证或模型请求；这些范围及真实源故障副本、builtin正向回归、新候选和同安装双Worker继续未验。Worker限定接受通知为`msg_cfe3b755c511`。

已读Grok来源reader及grok-source28非秘密结果：成熟TOML/JSON、三文件前后stat保持，配置无默认模型；唯一所读账号的expiry为2026-10-21T18:52:47.570220800Z，已早于本轮检查，没有刷新或模型请求。该观察不能提供可用的替代账号。Claude禁止供应目标、Codex完整ToolsOff及其他Agent NativeLaunch缺口继续按实际范围保留，StepFun真实源仍待明确。17e440ea的计划/记录两份hook已获主控`msg_ebf576095746`、`msg_acf22bddd79d`明确完整读取回复；后继构建由原Worker继续，不结算整体CLI通过。

### 准确f2ab发行安装29与实际故障范围

**限定接受程序`f2ab4fd9155f4f4bb85c74845f3f3ef5eb607caa`的build16、正常ZIP提取、安装29已验子项和本轮实际生命周期/C08/C09；接受后继`01b23b452a7949f0468859044ff622c0be327398`两份文档的准确记录范围。** 完整构建、解压、安装及检查入口和原始日志已独立读取。build16实际0，来源为干净f2ab，形成dist/swa-cli-install-20261020-15；实际Node24.21.0、pnpm11.27及Python3.13.13，21份成品、28233路径闭包通过。四份不变native沿实际源码、依赖、Node及编译条件核对原build11证据，从准确5070073c904aac6498dc821d6e775130deeb4257候选普通复制，未宣称重编。

extract-package29实际0，成熟.NET ZipFile正常生成410059866字节完整ZIP并提取至仓库外f2ab4fd9-package29；原Node pin保持，实际父子进程为同一用户、Medium完整性8192。原安装器生成仓库外`C:/Users/DW/AppData/Local/SwarmCli/candidates/f2ab4fd9-install29`，**install29内外实际退出均为6，完整安装验收保持未通过**。已验子项覆盖真实复制/准确版本、native/空Job、完整闭包、端口占用退出5、保护根及四种junction退出3、目标字节/ACL和原Orca保全。T03/T13真实消费者仍未运行；安装原件位于case29/.verification/swa-cli-headless/t02-65bca04b-93d3-4f21-a217-01f70ee1b07c。

Reviewer用该安装的固定Node实际执行`review-external-install-26036164.mjs <install29> <完整f2ab SHA>`，退出0，原件external-install29-independent.log。五个实际.node分别加载conpty、windows-process-tree、windows-security、windows-registry和watcher；原Job API1创建、读取空Job并关闭，creationTime/SID/registry只读和动态SDK/数据加载成立，21个实际加载文件均位于安装目录。实际`review-inline-licenses-780a.mjs <install29> <完整f2ab SHA>`退出0，原件installed-inline29-independent.log，逐项核21份成品meta及九个准确版本许可证。实际swa.exe --version返回完整f2ab且退出0；这些独立检查没有启动服务、Worker或模型。

check-installed29入口及执行快照已完整读取，实际0。准确同一安装沿原run-lifecycle20和run-http-idle16，在新独占根完成本轮检查，没有替换原95ef服务。Reviewer实际`review-lifecycle14.mjs <lifecycle29-outer> <完整f2ab SHA> diagnostic cim`和`review-lifecycle29-source.mjs`均退出0，原件lifecycle29-independent.log及lifecycle29-source-independent.log。四种模式全部OBSERVED、remainingCases为空，实际退出130/130/0/6；真实native两次旧身份观察间隔分别278/278/269/278毫秒，完整CIM两次间隔分别461.2836/467.37/421.8026/485.5947毫秒，每次416行、所有PID有效。准确执行快照与f2ab的serve-lifecycle和observe-process-exits原文件一致；已启动模式完整Job两次仅余serve，C09为唯一关联非空runtime的精确诊断。

真实创建时间缺项共四条：readiness-eof的swa37240/父33304与conhost32444/父37240，startup-stdout-close的swa48696/父33304与conhost33604/父48696。原样本及creation_time_unavailable原因保留；没有补写出生时间或向未知身份发送终止信号。各模式两次成功完整CIM均确认这些PID缺席，known身份退出和unknown PID缺席分别成立。该结论限本轮实际数据，未扩大为实际Agent中止、父进程丢失、CIM查询失败或超时场景通过。

Reviewer实际`review-http-idle17.mjs <http-idle29> <完整f2ab SHA> 7 <http-idle29-outer-exit.txt>`退出0，原件http-idle29-independent.log。实际runtime为0b2287db-21b2-499b-ae7a-336ab889e9d0，HTTP故障只终止经原持有句柄/完整身份确认的14304，daemon故障时存活；实际serve退出7/runtime_error，readiness唯一且没有尾部消息。完整Job两次仅余serve间隔100毫秒，旧身份两次退出间隔929毫秒；五真实库完整性及外键、三项真实拒绝、六保护文件和原Orca保全通过。没有human session、模型或实际任务/恢复验收。

原FD5在Worker检查前/中/后及Reviewer本次只读重核均0，human22/64509、runtime cc80c07c-c8bf-47ff-965e-6177bd2a7c69及原工具5187保持。新检查服务均已结束。原human23/consumer23及已接受Observe不重复执行，凭据实际expiry保持原2026-10-21T20:51:16.4468291Z，不能由保留文件推断后续仍有效。准确f2ab→01b两份文档差异及worker-handoff29全文已读，实际check-revision退出0，成熟marked18.1.0核两份/6本地链接、原§15锚点及Git空白通过；限定接受已通过`msg_c88f4297a8b9`交原Worker。788db525两份hook已获主控`msg_b70e9d35f340`、`msg_8873ae8b776b`明确完整读取回复。

本轮模型调用为0。真实StepFun源、custom/builtin完整正向及真实故障副本、执行Grant、实际Agent调用、T03/T07/T09/T11/T12/T13剩余场景及同一最终安装双Worker重叠/指定停止A/B继续/完整成果保持未验。主控最新`msg_a6d3cbe3ec5b`要求继续独立核已有OpenCode/DeepSeek真实来源及原生配置/工具关闭/隔离合同；原Worker继续同模块实现，当前不以StepFun局部源缺失结算整体阻塞。未使用Aliyun，未改变原权限、S3或产品main发布状态。

### 读取完整OpenCode计划前形成的条件

实际安装package为@opencode/cli2.0.31，原Orca准确v1.4.230的ACP launch spec、account binding、OpenCode dialect及环境处理已读：2.x ACP沿原生私有serve --stdio和stdin所有权工作，原通用ACP本身不保证全部工具关闭。独立git ls-remote取得上游v2.0.31准确`22728d67461260d56bb9796b9b8f36e1de47202e`。完整读取相关config/discovery、credential及legacy migration、permission/Tool.snapshot、Agent/provider配置、plugin发现/supervisor、instruction、warming/title、CLI/server及ACP账户消费者；自有review-opencode31-public-source实际0，原件opencode31-public-source-independent-02.log。25份公开文件中12份与上游Git blob逐字节一致，其余13份Worker只读副本只增加文件末尾空行，核完整正文及原blob通过；该差异未登记为逐字节相同。

实际core Credential.Service从SQLite credential表读取integration/value并按active/time_created/id选择。原生migration把legacy api条目导入type=key，已有该integration时跳过导入；因此auth.json存在不能替代最终私有库选择验证。config discovery还包括系统目录、全局与项目向上目录及.agents/.claude；CLI实际提供project/Claude相关开关，系统/全局隔离须沿真实环境和私有目录核验。Tool.snapshot按最终permission ruleset最后匹配规则移除定义，并单独过滤execute；Agent及session后置权限均可能影响该结果。plugin supervisor支持原生移除/启用操作，guarded内建插件继续保留；移除插件不代表其发现期间从未读取来源。

| 条件 | 必要行为及失败边界 |
| --- | --- |
| D01 原生版本与账号 | 精确使用实际2.0.31及原生认证/协议，真实来源只读；沿原生legacy导入或已有credential选择核私有SQLite实际integration、凭据类型与所选身份，秘密不输出，原账号及Orca保持。 |
| D02 每Worker隔离 | 私有HOME/USERPROFILE、XDG、ProgramData、TEMP、DB及配置目录的实际消费者均须核实；关闭项目、兼容规则和自动来源，核最终plugins/MCP/指令及来源，两个Worker无账户、临时目录、DB或所属资源混用。 |
| D03 工具关闭 | 核原生最终全局、Agent及session权限和实际工具定义，覆盖文件、命令、网络、subagent及execute/code mode；不能只凭空plugin/MCP对象或客户端不处理permission请求推导关闭。 |
| D04 准确模型与外发 | 真实选择固定DeepSeek provider/model及原builtin目标、API和原native消费者；查明title、warming、compaction、默认模型、fallback和provider插件行为，只允许已审来源与目标，无Aliyun或隐含外部插件请求。 |
| D05 原生生命周期 | 复用准确Orca ACP及原协议、私有serve/stdin所有权；真实握手、native session和正文事件可定位，全部所属进程/Job/监听可关联，原95ef服务与其他Worker保全。 |
| D06 生产准入 | 沿原AgentSelection/configuration_version、真实model_target、当前项目设置及Grant校验启动和每次输入；Observe与执行分别验收，未签发执行Grant时产品控制拒绝保持，不改变原工具与T11未决项。 |
| D07 错误与未知 | 缺失/过期或不支持认证、源变化、模型不符、配置扩大、协议/进程失败和未知效果立即停止；保留实际原件及部分状态，不制造key、成功CLI、Actor或TTY，不盲重试未知请求。 |
| D08 完整真实验收 | 原生成功、产品闭环和安装验收分别记录；同一最终安装须完成两真实Worker执行重叠、指定停止A、B继续、完整正文/成果及所属资源退出，真实库/事件逐项定位，未验范围保持未验。 |

该表形成于读取Worker完整OpenCode计划之前，依据当前明确需求和上述实际消费者。新增接口仍须沿主控规格流程确认，表内条件不代替尚未作出的权限决定。

### 单次真实DeepSeek模型列表检查

在读取窄计划前独立形成五项要求：真实literal api来源和无覆盖的准确builtin目标；单次GET且禁止redirect、30秒期限和无重试；原源前后stat/字节保持；秘密及原错误正文不输出；HTTP200与所选模型存在分别判断，不能计作原生账户或推理成功。完整`deepseek-models31-plan.md`准确SHA256 `35F8B8BD52972D4F885B4FED40B11FAF3D0F4DE78503C3D6E2C35A822921D548`逐项覆盖，未要求多余原生启动或模型；**接受该计划**，实施通知为`msg_229457546952`。Reviewer没有发送重复HTTP请求。

Reviewer实际以安装Node24.21.0执行自有review-opencode31-account.mjs，退出0，原件opencode31-account-independent-02.log。成熟JSON/已安装jsonc-parser确认实际auth.deepseek仅type/key、type=api且key非空，默认deepseek/deepseek-v4-pro、该provider无覆盖；实际公开snapshot的Git blob为67261ea5f5641232554e25258f600c0b6da3ce6c，builtin目标为https://api.deepseek.com、package为@ai-sdk/openai-compatible且包含所选模型。原auth/config和公开catalog前后stat及字节保持，未启动原账号CLI、迁移数据库或调用模型。[DeepSeek官方文档](https://api-docs.deepseek.com/api/list-models/)已完整读取，其模型列表接口和返回结构支持本项只读验证。

**接受准确客户端`B3F40C728C7CCD57915579C2CD1950F4561AC24D22CF781D3CFD0105708AD765`的实际GET31范围。** 完整4382字节客户端及原始log、退出码和result已独立读取；实际固定Node下于UTC2026-10-21T20:28:05.162Z只发出一次GET https://api.deepseek.com/models，返回200，所选deepseek-v4-pro存在，实际退出0。redirect:error、30秒signal、无重试，认证仅在内存HTTP Header，保存内容限安全元数据；原auth/config stat及字节保持，没有源写入、原生session、推理请求或Grant。原件deepseek-models31/result.json与log一致；该接受限本次真实key的模型列表读取，不证明余额、实际推理或2.0.31最终credential选择。

32f254c5安装范围hook已获主控`msg_794a229cd5fb`明确完整读取回复。完整OpenCode计划、私有账户实际导入/最终权限、原生模型及产品执行仍待原Worker提交，不依据本次200扩大接受范围。

### 准确OpenCode33完整技术计划

**接受`native-opencode33-plan.md`准确SHA256 `F18C0ED603668BE5A3FA1E49C8C9F2DF476D08C4824E17BBEC6C8BAD4AA63BD9`的技术方案、接口及O01—O10测试计划范围。** 当前正文已完整读取，独立D01—D08形成在先。D01对应O01/O03的实际源和原生credential选择，D02对应O03/O04/O09的每Worker根/DB及拒绝，D03对应O03—O05的最终权限/空工具和实际模型请求，D04对应O03/O05的固定逻辑模型/来源及初始目标和官方retry，D05对应O03/O06/O09/O10的完整所属Job，D06/D07对应O04/O07/O09/O10的当前批准、原输入、错误和恢复，D08对应O05及O07—O10的完整成果、同安装双Worker和停止。必要覆盖齐全，未发现多余用例；不重复既有GET/Observe或为文档重编不变native/Go。定向入口以实际文件和准确摘要交审，计划中的拟用命令未登记为执行。

新NativeLaunch account严格联合保留Pi，增加固定实际2.0.31/DeepSeek原生政策与非秘密来源和credential身份；每个实际Orca session有自己的private locator、HOME/USERPROFILE、XDG、ProgramData、TEMP和DB。原node-script与真实executable的IPC分支分别处理，原ACP request/response/update沿准确原协议进入trace，没有合成Pi事件。完整结果增加实际opencode-sqlite source联合，按所属session只读事务取得全部原始消息、完整文本和12MiB限额；检查输入/模型/状态/版本及原journal，跨库一致性分别观察，秘密出现时整体拒绝导出并保留私有原文。Go/TS/HTTP/schema及原launch/account/resume/input/result/build/install直接消费者均列入同一Worker范围。

官方policy插件就绪接口已具体为同一私有serve的plugin.updated事件→官方plugin.list实际inventory→原stderr管道，原Standalone在OPENCODE_PRINT_LOGS=1时继承该管道。严格状态关联每Worker值、private root、单调序号和实际server PID；host沿原Job取得真实创建身份和成员，不依据文件或setup日志准入。缺失/failed/pending/inactive或身份/来源变化拒绝首笔及后续输入，原生hook逐次核inventory/source；缺插件真实用例不得产生ready或prompt。清除原工具/来源和最终global/Agent/session deny共同关闭工具，实际context和HTTP请求核工具0，primary模型/初始URL固定，辅助/WS意外通道拒绝，原retry hook设false。插件不注册工具/provider，不改写model/prompt/请求正文；O03先以真实私有实例无推理验证生产与消费，随后进入真实正文和最终整包。

独立完整读取原Promise session hook与PluginHooks/retry及model-request相关消费者，callback沿Effect.promise逐项执行、handler之前的失败传播具备实际源码依据。另读准确AISDKNative和factory，确认该实际DeepSeek原生映射。opencode-credential-db32完整reader/result/退出0已读：node:sqlite readOnly/query_only只查询原credential表，唯一DeepSeek key、active=null，原生排序选中值与legacy相同；只输出非秘密身份和等值布尔值，原auth字节/stat及DB stat保持，没有原native CLI或模型。此源观察与后继每Worker真实导入/选择分别验收，Reviewer没有重复读取原会话内容或调用HTTP。

HTTP范围依据主控`msg_1861239c4c1a`：保留原CLI默认follow，产品检查初始URL及固定逻辑provider/model/source，不声称控制全部网络跳转。独立读取准确core/util平台绑定与RequestExecutor、同版effect4.0.1官方npm包；成熟JSON/.NET SHA512核dist.integrity一致，标准tar读取原FetchHttpClient与HttpClientRequest，实际fromWeb不保留redirect，RequestInit只提供timeout:false。该事实限定传输声明，没有新增禁止redirect能力、验收或阻塞；GET31的独立redirect:error保持原范围。既定原生账户/权限、失败停止和真实完整执行条件保持。

完整计划接受通知为`msg_47ae44cc7e18`，主控接口定稿请求为`msg_51e0753e5b74`。计划第4节的CACHE_HOME/CONFIG_HOME/STATE_HOME按原Global roots使用完整XDG_CACHE_HOME/XDG_CONFIG_HOME/XDG_STATE_HOME，并在O03直接核实际根；该明确含义已交原Worker和主控。主控写入准确SPEC后同一Worker连续实施O02/O03，再完成全部受影响消费者及真实正文/同安装双Worker，不在计划接受处结束任务。Reviewer不签Grant，原Observe按实际时效自然到期，执行Grant由主控按真实批准路径安排。

**同时限定接受主线`d21aee9088a5f1e4a568cf09dd99025eb741f310`五份事实源的记录范围。** 相对b530的准确差异已完整读取，Pi custom原合同、f2ab已验安装/故障范围、95ef Observe、GET31及继续OpenCode分别记录；完整安装6、实际模型/双Worker、T11/S3和执行Grant未验或待定保持。实际check-revision退出0，成熟marked18.1.0解析五份/95本地链接，Git空白通过，不为管理文档另开功能阶段。5e038466的窄计划/GET hook已获主控`msg_07afe97ef05d`明确完整读取回复；当前没有将OpenCode真实运行或完整CLI记为通过。

### 真实私有实例34—38及正式原生初始化补充

**OpenCode正向原生执行仍为证据不足；接受`native-opencode39-import-plan.md`准确SHA256 `F803EF3DB6697F921C0692D29FCB2E6D012F9694585CA837ADD108AFDC37D78E`的必要初始化修订。** 该修订仅替换全新私有SQLite的初始化路径，保留F18C的其余合同和O01—O10。原生账户使用同版正式`opencode.exe auth import --standalone`，随后另启ACP实例验证真实凭据和空工具。主控`msg_be11d514447e`已将准确修订写入其SPEC §21.1.6工作树并授权接受后直接实施；Reviewer已通过`msg_008aa6e7bf58`通知原Worker继续O03，主控通知为`msg_6c08f53a7d5d`。

原Worker准确程序基线为`3bbfc1389a63535d0b3c13d01a86e105a110a3fd`，含三个OpenCode模块和真实exe IPC修改；36—38各自SOURCE均为该SHA，source.diff及七个定向产物保留当时的工作树差异。Reviewer完整读取了初始五份程序、定向build/native/PowerShell入口，以及后继修复和实际原件；本次没有把这些工作树产物登记为准确Commit的完整实现通过。34/35的实际定向构建均0，探针均1：34在jsonc-parser捆绑加载阶段失败，35在来源准入阶段失败；35 strict实际2及原类型错误原样保留。这些失败没有启动模型。后继定向build36—38均0，真实native ACP/private serve已启动，实际政策准入均失败、没有prompt或模型输入；构建0和进程启动分别记录。

36/37的原log和cleanup已完整核对：36为root43660、实际native48604、conhost20600、private serve20820；37为root15148、native38928、conhost35172、private serve620，均有实际创建时间和同一所属Job。36两次完整Job空查询间隔约78毫秒，两次独立四身份均已退出；37对应间隔约79毫秒，两次四身份均已退出。两者root/tree均exited、forced=false，政策状态failed。38的真实root46288/native50360/conhost47108/private serve27168亦有创建时间，Job owner49944、limitFlags8192；两次完整Job空查询约1404/1482毫秒、两次四身份退出观察约1418/1496毫秒，forced=false。38安全失败分类为connection_resolve/TypeError，源原件未导出秘密或原生未知stderr正文。

实际38只读检查器及`opencode38/actual-private-db.json`已完整读取：私有原生schema存在，credential实际为空。准确上游`core/database/migration.ts`的fresh分支执行schema.up后登记所有migration完成，既有库才逐项执行migration.up；因此全新库不能依赖legacy auth文件自动取得凭据。原`integration.connection.active`取实际排序后第一项，空凭据得到undefined，后续resolve访问connection.type，符合38的真实失败。该范围支持修订初始化步骤，尚未证明后继正式导入或模型成功。

依据D01/D02/D05/D07，本新增路径必须满足以下条件；F803逐项覆盖，没有多余检查：

| 条件 | 初始化与验证边界 |
| --- | --- |
| I01 真实原生来源 | 只读取得实际选定Entry的原id/integrationID/label/value，active对应原生实际选定身份；同版正式CLI和原credential.list/create写入，不改写数据库或migration记录。 |
| I02 私有且无推理 | 每Worker新独占根、原已审HOME/XDG/DB/配置/DACL，导入前实际库为空；仅所选真实DeepSeek条目，导入过程无session/prompt/model，不启动原账户managed服务。 |
| I03 秘密与退出 | Entry JSON仅在内存经stdin，秘密不进入argv、环境、日志、断言或错误正文；实际命令退出和完整所属Job两次收尾确认后才进入只读核库。 |
| I04 实际持久结果 | 只读核唯一DeepSeek、原id、真实key类型/值及selection；失败或效果未知先查询原私有库，禁止自动重发或把unknown当空，后继独占案例保留原反证。 |
| I05 独立ACP准入 | 后启ACP/private serve有自己的Job，沿原同实例plugin.updated/list active、真实connection.active/resolve及工具0重新验证，不复用导入进程的就绪状态；缺失/失败无prompt。 |

已全文核准确227的migration、Credential.Service/schema、auth import/shared/export和ServerConnection；正式import先解析标准Entry数组，createClient的standalone分支只启动私有原生服务，再逐项credential.list/create，真实原命令处理重复id并输出计数。原auth export直接输出同一Entry，实际选择沿active/time_created/id和原connection排序；本次没有运行原账户export、CLI或模型。[原生import源码](https://github.com/anomalyco/opencode/blob/22728d67461260d56bb9796b9b8f36e1de47202e/packages/cli/src/commands/handlers/auth/import.ts)与[原生服务选择](https://github.com/anomalyco/opencode/blob/22728d67461260d56bb9796b9b8f36e1de47202e/packages/cli/src/services/server-connection.ts)支持这一路径。网页工具的两份raw读取失败后使用实际HTTP取得公开源码，无依赖安装。

Reviewer以安装Node24.21.0实际运行`.verification/review-opencode39-source.mjs`，退出0，原件`opencode39-source-independent.log`；成熟JSON和Git核九份公开源的准确blob，五份逐字节一致，四份Worker副本仅文件末尾空行差异，未将后者记为逐字节相同。真实36—38原件及38只读库检查支持失败范围，不重复模型列表GET、Observe或原native探针。已读后继源码中保护根双向检查使用真实config/data/Orca/安装目录，凭据比较仅固定安全错误，plugin.list消费官方Location.response.data；这些修复交原Worker，后继准确实现及正向仍须复验。

主线d21→`c3fcd86a76b9477fb2da77ae8928b439e0f80a25`五份准确差异已完整读取，实际check-revision退出0、marked18.1.0核五份/95本地链接及Git空白；其接口记录与F18C对应，当前初始化合同以主控已核定的F803补充为准，后继准确记录随实现累计复验。828ca336的计划hook及d21记录hook分别获主控`msg_b284d3bc9087`、`msg_324627f50a0e`明确完整读取回复。本次继续同Task/Dispatch审核真实初始化、正文和全部产品消费者；原95ef/5187保持，Observe实际已到期不续签，执行Grant、完整安装/双Worker、T11及S3保持各自真实状态，未使用Aliyun。

### 实际42私有导入与无推理ACP

**限定接受42的真实私有原生导入、ACP握手/设置及所属收尾证据。** 准确基线为`3bbfc1389a63535d0b3c13d01a86e105a110a3fd`加`opencode41-build/source.diff`，差异SHA256 `D4074E8C790B049124CDE45DC200C85D1661F5879523F506387EA6BC473663FC`；已读探针SHA256 `2BEFFFEB6A20A987592548C3359737BE28D469199BB9889AD236E45150039140`，Reviewer自有native-opencode42-reviewed.mjs及opencode41-reviewed-source.diff保留逐字节副本。八个定向产物由41构建，42入口核原差异不变后复用，没有重编不变成品。该轮使用固定公开项目和随机私有探针根，没有创建产品Orca Dispatch，接受范围不包括正式产品执行、首次prompt、最终session工具及完整安装。

原`opencode42.log`、真实退出0、import-exit/import-result/result/cleanup及入口已完整读取。正式原生auth import实际退出0，原真实credential id进入新私有库，计数1、stdoutBytes0、六条其他stderr仅计数且不导出正文；导入Job真实root12984/native15120/conhost44700的完整Job及三身份两次退出。后启ACP使用另一所属Job，root5480/native43524/conhost33336/private serve43600均有实际创建时间；原plugin.updated/list得到准确七个active插件、tool.list为0，状态PID43600属于实际Job，根/目录/数据库与各XDG路径匹配。ACP initialize实际报告OpenCode2.0.31，原生session为`ses_e9fb3dcaaffeJ5r1F6PF21cuWv`。

新session最初回应中的model为opencode免费模型，实际setMode/build及setConfigOption/model后，原生SQLite持久agent=build、model=deepseek/deepseek-v4-pro；没有以配置文件或最初回应推断最终模型。原session_message的两条记录实际为agent-switched及model-switched，没有prompt/模型正文。该轮未调用模型，效果前最终session/model/context/HTTP工具政策由后续O04/O05和生产消费者另行验证。39的原生导入已真实退出0且库中有唯一凭据，但Windows CRLF计数断言失败；40对两条原生设置记录的数量断言失败，41对同一canonical目录的斜线形式断言失败，原log/私有库及实际非零原样保留。42使用新独占案例，没有重发原导入或未知模型请求。

Reviewer实际使用安装Node24.21.0执行自有`review-opencode42.mjs`，退出0，原件`opencode42-independent-03.log`。node:sqlite readOnly/query_only直接核全credential表只有一行、原id/integration/type/key和active=1，key仅在内存与private auth作安全等值比较；核原session和两条真实有序记录，数据库dev/ino/birth/mtime/size保持。安装原生windows-security只读核12个私有目录的protected DACL、当前用户所有者、仅当前用户/SYSTEM及权限；成熟stream-json JSONL核导入/ACP两个不同Job各自两次完整空查询、两次已观测身份退出且间隔均超过50毫秒，forced=false。Reviewer首次检查因本树未编native而失败，第二次因JSONL适配入口错误失败，实际改用已有安装native和成熟jsonl.asStream后完整检查0；没有构建native或重复原生/HTTP/模型。原FD5身份guard实际0，95ef/5187保持。限定接受通知为`msg_ee7604d98544`，原Worker继续O04/O05及所有后继消费者。

**同时限定接受准确主线`c05b341def7c67181991703fd6f717897c1c948c`的管理记录及F803初始化合同范围。** c3→c05四份准确差异已完整读取，正式import/原Entry/内存stdin/空库与两Job/未知先查写入SPEC，开发基线3bb和已验安装f2分别记录；未将34—38失败、首次模型、完整安装/双Worker、执行Grant、T11或S3改为通过。实际check-revision退出0，marked18.1.0核四份/50本地链接及Git空白，前序c3五份准确差异及对应检查已在本报告记录。1e8f8d84的补充接受及正向证据不足两个hook分别获主控`msg_57d899dacc41`、`msg_b17aa9b9cedb`明确完整读取回复，本Task/Dispatch继续。

### 准确fe822原生实现与46真实模型

**限定接受`fe822783c92bd6a303b4f83157351e6b159e68f4`的原生程序、43—45已验子项及46首次真实模型范围；完整产品仍为证据不足。** 累计核3bb的真实exe/node-script IPC生产与Job消费，以及fe822的四份准确程序差异，相关全文和此前原件均已读取。正式导入等待真实exit、stdio close和完整所属Job后核计数/持久结果，内存stdin清理、失败与未知停止保持；私有config携带准确源/WAL存在状态供每hook复核。原plugin/agent/session接口按准确官方形状消费，实际只保留build及末尾wildcard deny；session/context/model/HTTP逐次核固定选择和工具0，原retry=false及辅助通道拒绝保持。原Pi及其他产品消费者未因本次原生接受自动通过。

46-build的准确3bb基线及source.diff SHA256 `0AE295D9A194291EED7208BC4A203EBEB288CA0641581B3E7145C318AC7BEF51`已读，47-checks同基线及source.diff SHA256 `698621EC2E8930A32F84EFCC796DDE857D4ABCDC2825DB72EA545D9835B00344`已读。Reviewer自有check-opencode46-source.mjs实际0，原件opencode46-source-commit-independent-02.log：真实Git在Reviewer忽略目录的独立index中读取3bb、解析原差异并与准确fe822比较，两份程序来源均相同，未修改任何工作树或真实index。47的PowerShell记录为CRLF序列化，初次直接Git解析失败；后继只规范化传输CRLF至LF再交Git解析，原件保持、其他字符不变。47完整入口、严格tsconfig、SOURCE及实际日志/退出已核，TypeScript6.0.3 strict/noEmit/skipLibCheck=false实际0，architecture实际0并确认47个既有Go package边界；没有重复完整编译。实际check-revision核3bb→fe822为0，成熟marked18.1.0两份/6本地链接及Git空白通过，原§15锚点保持。

43实际准入核官方agent.list仅build、最终通配deny及工具0；44在原真实catalog选择deepseek-flash后，原生prompt被session guard拒绝，实际AcpAgentError及rejected/native_session_selection_changed，原库只有两次model-switched和一次agent-switched，没有user/assistant记录或HTTP请求。45的新私有库有真实key，移除该案例policy后原ACP无active；原private verifier明确configuration_changed，探针未发送prompt。Reviewer自有review-opencode43-boundaries.mjs实际0，原件opencode43-boundaries-independent.log，直接readOnly核三真实库、原消息ID/类型和配置、库stat保持；成熟JSONL核三个案例的六个完整所属Job均两次空查询、两次已观察身份退出且间隔超过50毫秒。该接受限上述真实边界，不宣称O04所有故障和权限分支均已验。

46实际使用原OpenCode2.0.31、固定公开项目和a.txt原文，唯一已观察模型请求为初始POST `https://api.deepseek.com/chat/completions`，实际model=deepseek-v4-pro、工具0、HTTP200；原native默认follow保持。准确session为`ses_e9faef009ffeNz0OBEXx0psmUh`，实际ACP end_turn，原SQLite idle_outcome=succeeded。Reviewer自有review-opencode46.mjs退出0，原件opencode46-independent-02.log：node:sqlite readOnly/query_only事务逐一核完整session和全部五条session_message，原user.text与公开a.txt逐字相同、files为空；实际类型为agent-switched/model-switched/user/assistant/idle，seq为1/2/6/7/14，按原真实递增顺序保留。所有ID均实际定位，未把原生稀疏序号改成连续序号。

478条真实ACP更新关联同一session，114正文块和362思考块分别与完整原生content逐字一致，另两条为commands/usage更新；完整正文4619个UTF-16代码单元。首次独立检查对思考标识的假设失败；准确[原生ACP转换](https://github.com/anomalyco/opencode/blob/22728d67461260d56bb9796b9b8f36e1de47202e/packages/cli/src/acp/translate.ts)明确思考ID为原assistantMessageID加reasoning ordinal，公开blob `c6960b4cc29641c06cc3d72da1fac915cd9ce3ea`已逐字节核对。后继按该原标识及实际唯一ordinal0核关联后完整检查0，没有生成替代事件或重跑模型。实际秘密值在导出session/消息/更新/policy中未出现，源版本、WAL存在状态及私有DB stat保持；导入/ACP两个真实Job及全部已观察身份均两次退出，forced=false。Reviewer没有调用额外模型。

**退回fe822文档§45的单项计量事实修复：原assistant.data为26912个UTF-16代码单元、30622个UTF-8字节。** 原件直接按两种编码计算确认，修复已通过`msg_8176dd2f15de`交原Worker；当前Worker工作树已采用30622个UTF-8字节，后继准确Commit待复验，不重跑模型、不另开功能阶段。准确程序及独立原生范围接受通知为`msg_483092407308`，原Worker继续产品选择/Go与HTTP合同、当前Grant、实际Dispatch/输入/trace/完整成果及最终同安装双Worker。新增kind触及原node-script，后续适用回归须核真实生产/消费和Job身份，旧于3bb的结果不直接覆盖该变化；无法运行的真实范围保持未验，不伪造key或协议。

本次没有创建产品Dispatch，没有签发执行Grant，完整产品、O06/O07—O10剩余范围、完整安装/双Worker、T11和S3保持各自真实状态，禁止Aliyun。6636的42与c05两个hook分别获主控`msg_da4904100106`、`msg_294dd17e2635`明确完整读取回复；继续本Task/Dispatch审核原Worker后继实现。

### 最终执行Grant计划的先行条件与27FA审核

主控`msg_60bb918a808d`要求沿既有真实Agent测试许可准备最终候选的受托执行客户端。Reviewer在阅读计划正文前形成G01—G09，自有原件为execute-grant-independent-conditions.md；已重新读取当前QUESTIONS、原23/26客户端及实际HTTP session/项目路由、Grant/token、项目权限和CLI动作/上下文消费者。比较main031dc与准确fe822，服务签发human cookie/CSRF、Grant和token的既有主体逻辑保持；产品增加私有owner准入、orca/commands项目路由及实际注册CLI校验。以下独立条件以这些实际消费者及现行SPEC为依据，Worker通知摘要及解释未作为用例存在的依据。

| 条件 | 正确实现与验证边界 |
| --- | --- |
| G01 真实委托身份 | 保留用户原文与主控具体测试范围，分别记录用户批准、Agent执行者、实际服务human/local-human主体及最终scoped Agent；不编造本人TTY、新批准口令或Actor/token，保留原23本人入口。 |
| G02 最终候选与选择 | 准备按实际可见阶段绑定准确最终源码/安装、新DataRoot、服务PID/创建身份/监听、endpoint/runtime、现存authority及固定公开Git。Grant前共用生产读取函数核当前OpenCode/DeepSeek-pro来源和工具关闭合同；token后核现有status/capabilities实际字段。完整选择仍经模型前原Go/native准入及首次实际原R/receipt/journal逐项核验，不要求新增公开selection接口；来源变化拒绝。 |
| G03 原生HTTP认证 | 真实auth/session、cookie、CSRF及Origin边界；Bearer不能回退至human cookie，scoped Agent不取得管理批准能力。12小时服务端session与客户端释放引用分别记录，内部owner能力不用于绕过批准。 |
| G04 必要动作与范围 | 按真实命令映射仅给必要动作；task.create及效果前重读Context实际要求read_context，controller.recover按start，resume须有实际恢复消费者或已核适用范围。single CLI、空tools、公开目录/原文本及当前Agent control设置逐项匹配，不扩大旧Mission或管理范围。 |
| G05 当前批准准入 | 每笔expected_revision/响应revision、完整settings及Grant归属/动作/撤销状态可核；模型效果前仍核当前Grant、selection、R/Dispatch/session与fence。准备成功不改变产品权限、T11或S3待答项。 |
| G06 未知效果停止 | 每笔请求前保留真实request_id/幂等header和安全阶段，不假设准备接口持久去重；token签发会更新digest，未知响应后先查原持久效果，不能自动再签、改SQL或清目录重试，不改写95ef旧Observe。 |
| G07 DPAPI与真实时效 | CurrentUser DPAPI、准确owner/私有DACL/无reparse、CreateNew及候选/DataRoot/endpoint/runtime/authority/project/caller/actions/选择绑定；真实服务token固定一小时，使用实际UTC/expiry，错配或过期在请求前拒绝，秘密不进入参数、日志、异常或提交。 |
| G08 实际消费与拒绝 | 真正解密消费并以同最终安装运行status/capabilities及固定后继验收；适用HTTP负向核缺cookie/CSRF、跨源、scoped Agent不能管理、错项目/主体/时效和来源变化。未改auth原证据可复用，新范围及消费者需实际证据，不伪造协议、凭据或时钟。 |
| G09 持久事实与保全 | 响应、实际库及CLI分别核；准备space/project/settings/Grant/token不计作已运行Worker/模型。原23/26、95ef/5187、旧Observe实际到期及本机Orca/账号保持，只控制本次已证实所属资源，未知不释放，禁止Aliyun。 |

实际源码明确：GrantProjectAgent仅接受服务认证human，IssueProjectAgentToken保存随机token的digest并返回当前UTC加一小时；再次签发同caller token会替换digest。AuthenticateAgentToken每次核当前Grant/expiry/digest，项目动作/目录/工具/外发范围另经App准入。CLI的controller.recover属于start，task.create及输入前上下文检查实际调用ReadProjectContext；据此可以推导必要动作，不按名称猜测。主控`msg_c6a500cc1957`已确认恢复消费者均映射start时将本次Grant收紧为实际五项，恢复功能验收保持，不需重复通用许可。

**退回`delegated-execute50-plan.md`准确SHA256 `27FAAE443540DA9B8B1CEEB3440E5DCF731CB7951F04F7B0896F8F7822C1D3A6`两项集中修订。** 正文已完整读取：第一，移除本次CLI没有独立消费者的resume，settings/Grant/DPAPI及验证统一为observe/start/send/stop/read_context五项；第二，在最终描述、DPAPI及写前guard中明确actual authority_domain与当前selection的非秘密来源/配置版本及provider/model/native contract，并说明签发前使用的既有真实只读生产者。当前正文主要固定二进制及agent/model，capabilities放在签发后，尚未具体覆盖签发前稳定选择/来源条件。主控`msg_8a3b2e56addc`采用该修订并明确分阶段：项目尚未生成时只核可观测的真实候选/来源，生成后补真实project/root绑定，Grant/token效果前完整匹配，不构造项目或selection记录。不为此新增业务API或权限、重跑模型或另开整体方案阶段；原Worker修改计划后给准确新摘要，客户端随后集中审核，O07原已接受实现继续。

27FA的其余E01—E08覆盖真实session/三个HTTP拒绝、实际独占新项目、真实返回主体/Agent执行者、UTC一小时token、CurrentUser DPAPI、未知写请求停止、真实库与CLI消费及原23/26/95ef保全；未发现需新增模型自我证明检查。准备前后执行表保持基线的检查支持确认没有Controller/Task/Worker或模型效果，未触发的网络未知分支保持未验。原件摘要及计划尚未作为客户端或签发接受，最终候选身份/selection尚未冻结，本次没有签发Grant或启动后继模型。修订通知为`msg_abc56a9a198c`、主控为`msg_84e5b2e484af`。

主控工具等待异常跨UTC21:41至03:33后，原Task/Dispatch于UTC03:36收到恢复通知并继续；没有重启或结算。原FD5只读身份guard恢复后实际0，旧Observe到期不续签。42d70d4e的原生接受和计量修复两个hook分别获主控`msg_9fdd638b6b7b`、`msg_e12907e497de`明确完整读取回复；准确main031dc三份后继记录已读，STATUS范围措辞普通修复由主控继续，当前未借时间流逝新增验收结论。

`delegated-execute50-plan.md`准确SHA256 `7DFE168251FD546960CAEA6AED4E6552B135E2AF565844101E377EA9A52AA125`中的五动作、真实HTTP主体、E01—E08及G03—G09条件保持。controller.recover实际映射start，公开目录、单CLI、空工具、原文字段及原T11/S3边界保持；真实12小时cookie与一小时token分别记录，原23/26/95ef旧Observe不提升，未知准备写入停止且先查原状态，实际响应/库/CLI分别验收。其签发前直接RPC读取方式已经退回，并由下述384D准确补充替代。

Worker `msg_bc8abd6bf59e`发现来源读取也受owner门禁约束，Reviewer随后完整读取controller-admission.ts和client.ts并核RuntimeClient：所有swa.*读取要求isProductOwner，RuntimeClient只从SWA_OWNER_CAPABILITY传入该值，无owner的registered_agent/selection会被拒绝。handler中的只读分支不能单独证明客户端可访问，7DFE这部分没有执行通过；不读取内部owner secret、不放宽准入、不新增免鉴权接口。819d97ed原计划hook `msg_ea29f427b587`已获主控`msg_530802abb40d`完整读取回复，此前接受不构成该RPC方式的实际验证。

**接受`delegated-execute54-source-plan.md`准确SHA256 `384D188FD31F477500401FE61D36F6C757F2EC7B4C35096C28BFF9D40BB7F066`的来源技术修订和S01—S05测试范围。** 正文已全文读取，依据先行G01—G09及真实生产准入判断。Grant前只读取准确候选VERSION/host产物、实际服务进程/runtime、真实Orca库现存authority及原生来源；没有实际项目时不生成NativeLaunch或项目记录。真实项目登记后，共用同版生产readAgentSettings和安装inspectNativeOpenCodePolicy，绑定原exe/version/config/auth/SQLite/credential/WAL、真实project/root及关闭工具合同，输出仅包含非秘密source_snapshot。每笔准备写入前复核，实际authority尚缺时保持未知，Grant/token前必须取得并固定。helper导入源码和metafile须与准确候选匹配，不能复制选择算法或启动第二个runtime。

token后及任何Worker/模型前，通过正常HTTP/CLI的status/capabilities核其真实公开字段；完整selection由首次真实Worker返回或原持久R取得，与预先绑定来源逐项比对，并以CreateNew的非秘密观测文件保留，原DPAPI不覆盖。原Go/native效果前的当前Grant、选择、R/Dispatch/session/fence检查继续。主控`msg_08260890ed3f`、`msg_7d7525da57e7`、`msg_ca057d7b791c`采用该分阶段方式，并明确不为取得测试返回而扩大capabilities接口；G02已按真实生产消费链修订，不存在公开selection查询时按真实字段与首次实际准入/原R核。384D没有把source_snapshot冒充服务selection或registry，没有增加模型自我证明检查。

384D通知为`msg_8dd6a3e6be9d`。最终准确客户端、实际前置描述和候选稳定身份仍须集中审核，Grant/token尚未签发，产品/双Worker继续未验；Reviewer不签Grant、不重复模型。03d7ac6f的27FA退回hook已获主控`msg_22d3680f757b`明确完整读取回复；Ajv外部ref消费者在既有窄写域随O07累计审核。

## 准确e9fa97e3产品消费者与完整Job退出

**退回准确程序Commit `e9fa97e33e1235577cfef7342942398eb804f1f5`的ACP完整Job退出处理。** 已按fe822至该准确Commit审查新增选择联合、Go序列化与来源验证、每Worker私有账户、初始化/ACP接入、输入准入、完整SQLite结果及直接测试/构建消费者。JSON Schema、TS与Go按Pi/OpenCode account kind分别保留字段，原23和产品权限没有变化；原生46的assistant数据计量已准确修为30622个UTF-8字节。完整产品、安装双Worker及故障验收尚未接受。

| 缺口 | 准确源码与影响 | 原Worker返修和验证要求 |
| --- | --- | --- |
| ACP未知Job退出被缓存 | acp-agent-connection.ts调用spawnManagedProviderProcess时未提供完整Job的acceptClose，原默认acceptProviderRootExit只要求根退出。RetryableProcessExitProof会保留root=exited、tree=live/unverifiable的结果，后续close无法重新查询并证明原Job。 | 复用原JsonlRpcAgentConnection的完整Job/真实processless接受条件；完整退出未知时保留原连接和资源，后续关闭继续核同一Job。 |
| 根退出提前发布会话结束 | 新onExit(listener)过滤不能保护managed.onExit内直接调用的options.onExit；后者经acquire的onExit进入adapter.finish并删除session，发生时只证明managed root退出，私有serve和其他Job成员可能仍存在。 | 同一已证明完整退出路径统一、一次发布订阅和options回调；相应exited/rootVerdict反映真实完整Job，未知保持所属连接可查询/停止，普通无Job ACP语义保持。 |

缺口由实际生产调用关系推导：managed-provider-process使用RetryableProcessExitProof和默认root-only判据；adapter.finish会endAcpStructuredSession并删除原session，后续stopSession对不存在的session直接返回true。原Pi连接已经对owned Job采用严格acceptClose和完整退出后发布，可复用该消费者模式。通知`msg_a77bde6c9899`交回原Worker，补真实所属ACP根/私有serve故障及Job退出证据，不用伪进程、mock或假退出结果。当前未执行破坏性故障或新模型调用。

新SQLite读取代码在真实只读事务中读取指定session的全部消息，检查严格递增但可有间隔的seq、消息身份/时间、fatal UTF-8、实际账户/模型/内容合同、完整最终正文及12MiB边界；返回前复核DB/WAL版本，整体task结果再核journal epoch/sequence和来源身份。它保留实际用量和账单费用未知，超限或不完整失败，不剪裁成果。新增备份故障消费者基于真实native DB的SQLite backup，仅改所属副本；这些实现静态检查通过不等于该消费者已实际执行。

已读取49/51/52原始入口与日志：49 strict和架构0，广包Go实际1，app的Relay Git/交付用例失败、httpapi的Workbench缺明确publication/SQL输入，未作同条件基线比较，不归类为既有失败；51 strict、gofmt/MJS语法与三个CLI传输前Go用例0，四项同安装真实TestOrca未运行。52同pin Node24.21完整check:contracts实际0，Redocly20项警告、264样例及生成一致性通过；准确Ajv修改注册既有外部SWA Schema并补示例基URI，约束和样例保留。build17入口要求准确干净e9fa，原SOURCE及日志指向同Commit、Node24.21和复用的四个未变native输入，build-exit为0；完整功能仍标NOT_RUN。Reviewer没有重复完整构建。

Reviewer实际运行`check-revision.ps1 -Previous fe822783c92bd6a303b4f83157351e6b159e68f4 -Candidate e9fa97e33e1235577cfef7342942398eb804f1f5`为0：成熟marked18.1.0读取该提交三份Markdown，六个本地链接无缺失，旧§15锚点仍存在；本报告WORKING的同库检查和Git空白检查均0。上述检查不替代缺陷返修或最终实际路径。

f5393fb4两项hook均获主控明确完整读取回复：e9fa退回`msg_baf565d18fb3`对应`msg_256464bbc5f1`，384D限定接受`msg_1fab0d024d13`对应`msg_af83ff50b903`，本次唤起已确认。原Worker已在工作树实现严格acceptClose、统一完整退出后发布和真实processless处理，准确返修Commit及实际ACP所属故障仍待核，未提前登记程序接受。

**接受54来源补充的准确后继SHA256 `64DD11A6A2C87CE7C1230AB3F249F7676D0E748CB69FEF629106307DB30EB010`，范围沿384D。** 全文及摘要已读取；唯一来源入口调整为同生产agent-settings.cjs随当前必要ACP返修整包导出，metafile/准确源码一致仍核。它取代自有目录另行esbuild导入方式，不复制parser或selection规则，不增加公开API/权限或独立runtime。通知为`msg_cf1100ecd2ff`，客户端与最终实际描述仍待集中审核。

**接受e9fa安装53的实际Pi node-script IPC无推理回归56。** 原入口run-pi55.ps1和native-pi-startup.mjs已完整读取，实际候选Node24.21启动真实pi.cmd及其原Node解释器/cli.js，私有空账户仅发送get_state，offline、原生工具及扩展/上下文入口关闭。实际session为`01a161c6-53ea-766d-ad75-5b66b01d8a16`，模型id为unknown，messageCount/pendingMessageCount均0，未复制原凭据、未进行推理。55首次原件退出1，真实Job已收尾，失败来自检查入口错误的manifest路径；后继入口改用Node原findPackageJSON，56实际退出0，55原件保留。

Reviewer自有`review-pi56.mjs`实际0，原件`pi56-independent.log`：成熟stream-json读取全部56 JSONL，按实际root 51724/出生1792641749666和native node-script 50768/出生1792641749770关联成员，核两次完整Job为空、两次全部已观察出生身份缺席，均forced=false；真实result与安装VERSION一致，before/after原Orca保全passed。Reviewer没有重新启动Pi或模型。53之后FD5来源摘要核对及只读Assert-SwaService独立执行0，原95ef/5187保持。Pi正向模型仍未验，完整产品/e9fa ACP返修/最终同候选安装及双Worker仍未接受；该回归不替代这些范围。通知为`msg_75d7101ae0dd`。

01386648的hook `msg_ca42d330e69a`获主控`msg_fe10c2c79d8c`明确完整读取报告、准确64DD、55/56原件和独立消费者后回复，限定范围已实际读回。当前Task继续。

## e9fa后续权限与完整导出调用检查

**准确e9fa继续退回以下既定F18C合同缺口。** 这些发现来自继续追踪真实生产消费者，没有发现当前真实模型结果含实际key，也没有构造native错误或请求。

| 既定条件 | 准确调用路径及缺口 | 原Worker处理与验证边界 |
| --- | --- | --- |
| 受限ACP权限请求立即取消 | acp-structured-acquire的onPermission在active turn进入AcpStructuredPrompts.handle；OPENCODE_ACP_DIALECT仅normalizeToolUpdate，通用translateAcpRequest会为session/request_permission生成presentation，handle写request.open并等待人工回答。 | 本次launch.swaOpenCodePolicy路径立即返回原cancelled，普通ACP保持。实际tools0不能证明未出现的incoming权限分支已执行；不能伪造native协议来登记通过。 |
| 完整最终成果检查实际凭据 | readNativeOpenCodeSessionSource仅对native SQLite value序列化检查snapshot.key，completeTaskResult之后还加入原journal rows/snapshot及其他完整字段，没有对合并的最终返回执行同一实际key检查。 | 将同一实际选定key检查覆盖完整返回，命中整体失败，私有原记录保持；不得通过删改成果后冒充完整内容。 |
| 原生错误仅输出安全分类 | AcpAgentConnection将super.prompt的AcpAgentError原样交AcpStructuredTurns，再进入timeline的promptFailed/promptRefused；acpPromptErrorDetail优先取dialect转换，再取原data字符串/data.details，最后message。OpenCode通用dialect没有受限转换，详细内容进入person diagnostic/journal；acquire的onDiagnostic另直接传native message给logger。 | 仅收紧本次受限路径的安全code/status/session等输出，保持真实送达、unknown和AuthRequired语义，普通ACP保持。未出现的上游异常不登记为实际已验，不把已经送达错误改成neverReached。 |

权限缺口通知`msg_98368a0df934`，完整导出及错误范围通知`msg_fdf73c269297`，具体data/details调用证据补充为`msg_8e6fa3fa670c`。与前两项完整Job缺口均交原Worker在原域集中返修，准确程序及必要实际证据形成后复验；没有新增功能、扩大权限或重复旧46模型。

委托客户端仍在编写，未固定准确原件，尚未接受或执行。Reviewer已读取其真实session/CSRF、五项scope、写前来源/authority检查、原请求未知停止、CurrentUser DPAPI及消费入口，working反馈`msg_bb767bb06960`要求核已有capabilities的enabled/provider/model/account_source，并指出RestartObserve的旧runtime guard、必需readiness/新进程输入和真实预期退出6。Worker回复`msg_bdd18c62540a`称已补专用的真实重启观察绑定，原DPAPI不修改/续签、普通模式仍拒绝runtime替换；此修订仍待准确源码及实际描述集中审查，未把准备成功计为R08完成。

主控管理Commit `eee780b9f20d9023c1216ee6013df3bba65c62ce`四份准确差异已全文读取：累计实际e9fa退回、build17及49失败、53局部安装6、56/64DD限定接受、五动作真实测试授权和原95ef保全，未将未知S3/T11或最终双Worker改成通过。此前STATUS模型范围措辞已限定原安装/产品Worker及780历史阶段。Reviewer运行031dc至eee780的Git/成熟marked18.1.0检查为0，四份Markdown的89个本地链接无缺失。只读使用.NET ZipFile核原e9fa归档411419795字节、25494条归档条目，VERSION/Node/policy成品各唯一；归档条目数与含目录的28243闭包路径分别记录，没有重新解压或构建。这些管理事实随产品累计，不构成e9fa程序接受或新操作窗口。

088f0fee的e9fa累计退回和eee780限定记录hook分别获主控`msg_85936106d5e4`、`msg_e2ea12a18ab6`明确完整读取回复，两项唤起已确认。

## 准确3bcf0c71返修及真实根退出62

**接受`3bcf0c71ba86e607cd63aef3b423d7cc6f5788eb`对前述五项缺口的源码返修和真实原生根退出62的限定范围。** e9fa至该准确Commit的13份差异均已完整读取。完整Job未知时acceptClose不再缓存成功，rootVerdict/exited保持unverifiable；统一verifiedExit在完整Job或真实processless后一次发布，原连接/tree可继续关闭。初始化的processless只表示没有实际创建进程，后续实际import成功条件仍要求原exit0和真实导入结果。

受限ACP的权限请求直接cancelled，其他客户端工具请求直接拒绝，普通ACP保持原流程。完整结果合并之后运行每个真实OpenCode来源的validateExport闭包，检查覆盖summary/native session/journal原rows与snapshot，命中整体失败。受限错误保留AcpAgentError、AuthRequired、InvalidResponse、ConnectionClosed、Timeout等实际类别和安全numeric code/status/session，原data/details/body/cause不进入通用消费者；admitNotification在全部journal/扩展订阅分发前执行真实trace及凭据检查，失败关闭同一Job并阻止该条外送。普通连接的两个基础钩子分别返回原错误和允许通知，未替换普通ACP协议。

自有`check-opencode59-source.mjs`实际0，原件`opencode59-source-independent.log`：在自有忽略目录的独立Git index以真实Git读取e9fa、应用59的完整source.diff，对全部程序/构建/测试文件与准确3bcf比较一致，原diff仅在内存规范化CRLF传输，原件保留，实际工作树和真实index未修改。59原入口、strict-exit0及语法日志已读，准确程序与该严格检查对应；客户端位于忽略目录，后续仍须按冻结摘要分别审核。两份准确Markdown的成熟marked18.1.0/Git检查0，六个本地链接无缺失，旧§15锚点保留。build18入口要求准确干净3bcf，实际exit0，Node24.21、29份bundle/metafile及28251路径闭包，四native继续来自原507/build11准确输入；Reviewer没有重复构建。

62使用同一仓库外3bcf-install60、真实原CLI及官方auth import、新独占私有账户和固定公开目录。原native session为`ses_e9e19e0f8ffeLNzuncnmOEsKuw`，实际build/DeepSeek-pro、七项active工具0；没有prompt。故障入口在核managed root的实际程序与创建身份后终止该根，真实退出通知仅一次，发布时exited/rootVerdict均exited、processTreeUnproven=false且lastCloseResult为完整Job退出，再次close返回true。此驱动采用普通ACP加owned Job，未传产品swaOpenCodePolicy对象，没有产品Dispatch或Grant。

Reviewer自有`review-opencode62.mjs`实际0，原件`opencode62-independent.log`：真实只读SQLite逐项核两条原生消息`msg_161e61f7f0015LtoGv1BuZa7jG`、`msg_161e61fbe0016Z7Fbeq0f2t84e`及session，与原导出一致，消息仅agent-switched/model-switched，实际唯一active DeepSeek凭据与私有auth等值且结果不含该值；SQLite mtime保持。成熟stream-json核官方import及ACP两个完整Job，各有两次空清单与全部已观察出生身份两次缺席；实际ACP root为51304/1792643832430，native为12500/1792643832536。通知发布时的完整退出状态、驱动另两次实际身份观察及原Orca before/after保全均核。Reviewer没有重跑原生进程或模型。

该接受包括已检查的静态整改及上述真实根退出范围；未知到已知关闭没有实际出现，原件明确NOT_RUN。受限incoming权限、秘密错误/上游异常没有实际触发，保留静态保障及未验范围；62不能代替这些分支、真实产品输入或双Worker。安装60汇总实际6、Grant客户端/最终实际描述及最终同安装双Worker仍待验，原95ef/5187保持，整体产品没有接受。

## pnpm CI窄配置计划与参考来源

在读取Worker计划正文前，Reviewer依据实际CI安装失败形成C01—C03：六项准确包/版本及脚本/真实消费者有来源；明确允许或不运行有必要依据，保留严格未知拒绝和原受控native能力；真实frozen结果及未执行步骤按平台记录，最终同安装验证保持。**接受`pnpm-ci64-plan.md`准确SHA256 `45FE97A216B14C4CF3FBEF04B4790C8AEE5B28FFB48C13EC2F8BE8AE25B9702C`的技术与必要检查范围，实际配置提交和安装/CI结果尚未接受。**

实际job119639202522日志的错误尾段已读：pnpm11.27 frozen install以ERR_PNPM_IGNORED_BUILDS退出1，后续检查/测试/构建没有运行。计划只为watcher2.5.6、windows-process-tree0.8.0、agent-browser0.27.0、cpu-features0.0.10、node-pty1.1.0、ssh2 1.17.0补明确false。Reviewer核实际pnpm createAllowBuildFunction的版本规则和false优先分支、准确3bcf原workspace策略、各包实际元数据、watcher/ssh2安装源和cpu-features的原可选消费者、agent-browser下载/全局shim分支。watcher使用锁内平台预编译包，原check-swa-runtime真实require已通过；四native继续原准确受控构建和成品装载，没有批准隐式替代编译或浏览器下载，没有删依赖或改未知脚本策略。真实Windows归档frozen install、私有store完整性、依赖检查/锁保持及原Linux CI复验结果继续由原Worker提供；不能用本机结果替代Linux，不能将计划接受计为实际CI通过。通知为`msg_479dc758e59d`。

用户新增OpenMausBot参考由主控固定为`dbb73a73ae2cba48fdf608e9e22d1d5d91608532`。Reviewer只读Git对象，完整读取README、LICENSING、headless-cli/custom-engines文档及headless/output.ts，没有安装、运行或复制代码。其统一result/error/progress秘密处理可用于核对既定输出边界；其深度redaction会改写正文，本项目完整成果仍按实际凭据整体拒绝。该参考的本地owner、原生工具、重试及分页导出约定与本次合同分别核，不作为本项目Grant、ToolsOff、unknown或完整成果证据。当前没有必要直接复制建议，未来具体建议再核准确文件及Apache2/enterprise许可归属，不为参考库增加整体方案阶段；通知主控为`msg_888ca40623e2`。

## 68准确委托客户端与67真实服务

**接受execute68-review-input/actual-clients.json冻结的四份客户端源码及67真实前置描述，范围为既定G01—G09准备和后继消费者入口。** 四份正文均已完整读取并独立核实际SHA256；该结论允许原Worker依已确认五动作范围执行实际准备，Reviewer没有签发Grant、运行准备客户端或调用模型。实际HTTP响应、库中Grant、DPAPI消费、产品Worker和最终同安装双Worker仍须后续审核。

| 原件 | 准确SHA256 |
| --- | --- |
| execute50-guard.ps1 | DEA336E12B020A63D27D2F8B0B81ECA826FA82B4CF9ABD136C43710D5F565825 |
| delegated-execute50.ps1 | 380A7C2F44CF39D7362D763E08465238499F11C01B9CC5B03D2600D943B4CDA2 |
| use-execute50-credential.ps1 | 3D4B0940C45E3EC74428C618AD4B8F8312400831385AA17A0653FF8FEC0F0DE1 |
| observe-execute50.mjs | F0DDAFC5D525C5700FA7B2A2877AB0991DAD44E5F21DF925180B36309DD010E6 |

准备客户端使用真实auth/session、Web cookie和CSRF，分别记录服务返回的human/local-human主体与当前Agent Task/Dispatch/terminal执行者，未伪造本人TTY、Actor或token，原23入口保留。缺cookie、缺CSRF、跨Origin以及scoped Agent管理接口拒绝均有真实消费者；正向请求前移除负向Origin。space/project/settings通过现有API取得真实ID和revision，五动作准确为observe/start/send/stop/read_context，固定公开Git、单OpenCode CLI、空tools及当前范围保持。Grant/token之前须取得真实项目、authority和同生产native来源，缺失或变化立即停止；没有复制选择算法或读取owner能力。

每笔写入前保存当前UTC、原request_id和阶段，客户端没有将请求header当成持久去重保证；未知响应立即停止，保留已经发生的效果，token不自动重签。CurrentUser DPAPI使用原源码/DataRoot/runtime/authority熵及完整项目、caller、actions、来源观察和实际expiry绑定，私有DACL、owner、无reparse、CreateNew及最终路径均核。消费者普通模式拒绝运行身份替换，status/capabilities核实际enabled/installed/configured、2.0.31、DeepSeek-pro、account_source、registry产物。RestartObserve另要求原真实状态与新真实描述、原三进程出生身份已不存在、同来源/项目/数据/authority、新runtime及实际readiness/current-serve-process；旧DPAPI仍用原熵解密，不覆盖或续签。原驱动退出6明确为incomplete，未登记恢复通过。

Reviewer独立`review-execute68.mjs`实际0，原件`execute68-independent.log`：四摘要、三个同生产bundle及metafile摘要和入口一致，相关真实源码与准确`3bcf0c71ba86e607cd63aef3b423d7cc6f5788eb`相同；68原两次source JSON逐字节一致。真实只读SQLite核workspace五表、orchestration七表及journal十四表与原观察一致，DB/WAL的mtime和大小保持；现存唯一Run为生产迁移生成的`run_legacy_local`、legacy=1、generation=0及空coordinator。其余业务记录均0，local_authenticated_transport主体尚不存在，authority_domain/source_snapshot保持null；最终delegated-execute/version-home尚未创建。源migrate-v2-v12.ts的真实inspect-only bootstrap已核，未要求伪造全零基线或提前生成项目。

Reviewer独立`review-execute67-service.ps1`实际0，原件`execute67-service-independent.log`：私有描述、实际PID/创建时间/程序/父进程、监听归属、metadata、VERSION、CLI版本、host产物和health全部匹配。准确候选仍为3bcf-install60，数据为3bcf-service67，runtime为`9ea24e54-4f25-44d1-ae2c-72462e92e130`，serve55300、host53008、http46428，创建时间分别UTC04:39:46.262484、04:39:47.309058、04:39:48.482722，端点64508。Reviewer没有写入或轮询原服务工具17950/5187；原95ef服务及已到期Observe保持，未调用Aliyun。

2b4b9d22的3bcf/62与CI计划两个hook分别获主控`msg_d51e98d35289`、`msg_cd18a2caf6d6`明确完整读取回复，实际唤起已确认。本轮客户端接受不结算当前Task/Dispatch；CI准确配置与真实安装/基线比较、Grant实际准备及产品验收继续累计，尚未接受完整CI或产品。

## d829依赖策略与3fadb迁移测试前提

**接受准确`d82958a0dd48ab87ea6e77a6d69d6305bf3d11b9`的六项安装脚本策略及实际Windows普通frozen安装范围，接受准确`3fadb0bc3aa69a62b1c9b3c772cb518ec57d768b`的迁移测试前提修订和真实SQLite升级用例。** 两提交全部差异已读。六项准确版本false与45FE已接受计划一致，原严格未知拒绝、esbuild规则、依赖/锁和受控native保持；3fadb仅将真实迁移文件前提增加到28/028，原25项旧版本、两Taste表、备份、完整ID、幂等和无human profile断言全部保留。两提交没有改产品权限、原S1/Relay接口、现有3bcf程序或native输入，没有要求重复未变成品和模型。

64实际Git完整归档后因Windows的SOURCE文件与source目录冲突，在安装前停止；69普通安装因两个workspace native的MSBuild长tlog路径失败，两份原错误保留。72改用独占短目录c72和同一真实Git归档/私有store，实际命令为`pnpm install --frozen-lockfile --verify-store-integrity --store-dir <实际私有store>`，没有ignore-scripts或关闭strictDepBuilds；实际退出0，随后check-dependencies实际0。源码归档117429167字节，准确依赖输入逐字节相符，锁前后SHA256均为`F7A1BC8DAE51770233894041E0F3A415A3C68AEA57BF06A235159D5E3498D3EA`。Reviewer完整读取64/72入口、72实际配置和日志，独立`review-ci72.mjs`实际0，原件`ci72-independent.log`；成熟yaml解析核准确变化仅为六规则，真实归档摘要、四份依赖输入、锁、退出和后继日志均核。

Linux job119643340844实际检出的PR合并Commit为`89ff401aecc725ade76a4e97132e6165d4d990a1`，由d829与main `50c2aaff28a54310b1d290190859d6cc6cc1faec`组成，pnpm11.27 frozen install真实成功。后续pnpm check内Go仍退出1；Reviewer读取其完整日志并与同平台main job119638841884的全部失败名称集合比较，基线68项、候选69项，唯一新增为TestTasteProfileActualMigrationUpgrade，移除0项。该比较支持定位新增测试前提，不代表原68项失败已修复、整体CI通过或全部失败原因相同；后继build/浏览器等未执行范围保持，3fadb的Linux后继结果尚待原件。

73使用真实`go test -mod=readonly ./internal/adapters/sqlite -run '^TestTasteProfileActualMigrationUpgrade$' -count=1 -v`实际0，SOURCE=d829及完整source.diff与准确3fadb一致。Reviewer独立`review-migration73.mjs`实际0，原件`migration73-independent.log`：实际保存目录taste-profile-actual-303335430中旧25份SQL逐份与准确源码一致；只读实际state.sqlite含28个准确迁移ID、两Taste表且human profile为0，唯一真实备份仍为旧25个准确ID且无Taste表，两库mtime保持。没有重跑迁移或构造数据库。准确d829三份、3fadb两份Markdown的Git/成熟marked检查分别0，本地链接各六个无缺失，旧§15锚点保持。

主控`6e4ba81cf853dfd945335226dff362a6a7f75bf9`相对eee780的三份准确差异已完整读取，原3bcf/62/67、draft PR、CI窄域和OpenMausBot参考仅按真实限定范围记录；QUESTIONS/S3/T11及最终同安装双Worker未提升。独立Git/marked检查0，44个本地链接无缺失，这些管理事实随本轮限定接受。4aa077c4的68客户端hook获主控`msg_81ff405f9296`明确全文读取四份摘要、真实67描述及两个独立结果后回复，唤起已确认；原Worker实际准备继续，不新增管理等待或用户批准。

## 75实际前置失败与后继故障消费者

75一次实际运行准确准备客户端，New-OrcaPreservation的目录Created通知自检抛出异常，退出1，发生在ReadSource、HTTP和主try之前。原execute75-outer.log及唯一notification-probe实件保留，已经创建的delegated-execute/version-home不删除、不覆盖或盲重跑。Reviewer独立`review-preflight75.mjs`实际0，原件`preflight75-independent.log`：真实三库全部原业务计数保持、全部认证身份行0，没有项目、Grant、token、会话或Worker效果；私有目录仅有空version-home、没有DPAPI，原SQLite mtime保持。另独立调用原只读Private guard核两个现存目录的实际owner/DACL/无reparse为0。Grant实际执行继续证据不足，未运行HTTP或模型的事实不能登记准备成功。

通知自检的同一异常兼含没有事件和实际FullPath字符串不等两种条件，当前原outer没有记录实际事件，尚不能区分原因。原Worker负责以真实.NET事件诊断并提供已证preflight状态的准确恢复入口；恢复前须核原服务身份、空私有目录、真实无业务效果，后续保持完整来源、真实HTTP、CreateNew凭据及未知效果停止，不伪造事件或批准。通知Worker为`msg_404e9d09fb57`，主控为`msg_5a576a67dd75`。

**退回准确3fadb中三个既有故障消费者在当前3bcf成品上的来源前提。** controller-recovery.mjs:19—20、runtime-restart.mjs:24—25、cli-failure.mjs:25—26均将实际测试仓库HEAD直接要求等于候选VERSION并要求整树clean；当前准确HEAD3fadb与保持的3bcf程序不同，这三条验收入口必在故障效果前停止。应由原Worker在原测试域分别记录真实测试版本与候选版本，继续核必要产品/原生输入未变、成品及服务身份，保留全部真实故障、R、fence及所有权断言。固定公开项目HEAD的检查继续，installed-flow没有该前提；该返修不阻断75前置诊断及实际准备，也不要求为纯文档/策略/测试后继重编成品。通知为`msg_e12f1a63f929`、`msg_c53c3a75df67`；准确返修及实际故障结果待核，d829/72和3fadb/73的限定接受保持。

## 79真实路径修订与75显式恢复入口

**接受watch79-normalized/actual-inputs.json冻结的准备客户端后继`F72D42D2E36BDF33F6752B45EFC71018A99D4F6E06F27A60486D1D6EBD82E046`及保全模块`CBABA9130FA99410AEB54E48009EAD152EC7D300698311C3CAAB9927384D92E9`，范围为实际67上75已证preflight部分状态的恢复。** guard、消费者及来源helper三个摘要与68相同，四份实际摘要均独立核过；新准备客户端正文完整读取，保全模块准确差异完整读取。该模块当前按此SHA256冻结，不宣称上述工作树修订已经包含在3fadb Commit中，后继准确提交继续核。

77实际.NET Created记录约5毫秒、78按75相同forward路径输入复现：Created确实存在，事件FullPath与Join-Path结果原字符串不等，系统GetFullPath规范后相等。模块现仅规范ProbeDirectory及实际两路径比较，并区分没有事件和路径不符；订阅、实际写入通知文件、Wait-Event、finally清理、原文件/ACL/进程/监听保全条件全部保持。79用同forward输入执行原模块，Created及原Orca/FD5/67身份检查实际0。Reviewer没有另造事件或重跑保全写入。

恢复入口必须显式提供PreflightResumeEvidence，准确限定原execute75-preparation及其唯一普通probe文件；原delegated-execute和version-home须同SID私有DACL、无reparse，前者仅有后者且后者为空，保留所有原件。新证据目录仍CreateNew；第一笔实际ReadSource在任何session/HTTP之前要求authority/source_snapshot为null、唯一inspect-only bootstrap及全部业务表0，之后才按原真实cookie/CSRF、五动作、完整来源/authority、未知写停止及CurrentUser DPAPI/CreateNew流程继续。没有删除、移动、覆盖75状态或重签未知token，没有放宽产品权限。

独立`review-watch79.ps1`实际0，原件`watch79-independent.log`：冻结四客户端/模块摘要、真实78原两路径规范相等、79实际Created文件和原preservation通过、77实际来源JSON与68前置逐字节相同及新客户端Parser均核。此前独立75只读计数和空私有ACL观察共同提供该恢复的实际前提。真实Grant、DPAPI消费、产品Worker/T13仍未发生或接受，三个故障消费者来源前提返修继续，Reviewer不执行准备或签Grant。

## 准确5c75001e的故障消费者来源修订

**接受准确`5c75001e4507d1e68751bead2b1ded5742e27852`对三个故障消费者来源前提和79路径规范化的修订。** 七份准确差异已完整读取。三个消费者共用同测试域installed-product-source.mjs，分别记录候选3bcf和实际测试5c750版本；工作树仍须干净，Git比较全部必要程序、构建及原生输入，限定排除已审文档/许可及明确测试路径。pnpm-workspace只允许六个准确版本安装脚本false，成熟YAML比较全部其他字段保持；实际成品metafile另核被排除测试没有进入产品。原故障和业务断言、固定公开Git检查全部保持，原19/20类缺口已修，未将此静态及来源接受登记为实际故障通过。

81原真实检查0，Reviewer从同准确工作树调用该检查，只将结果写入自有忽略目录，实际0：`source81-independent.json`和`source81-independent.log`分别记录candidate_source_revision为3bcf、test_source_revision为5c750，全部必要产品源码不变及六规则例外。没有启动服务、原生Agent、编译或模型。准确两份Markdown的Git/marked检查0，六个本地链接及旧§15锚点保持。保全模块的准确Commit差异与此前CBABA冻结修订一致，79实际通知与独立观察继续适用。

实际80前置观察另记录原两个目录创建时间、私有DACL/无reparse、空version-home和无DPAPI，来源完整等于68；原Worker收到79接受后已在同67开始显式恢复82，不新增批准等待。准确生产源码亦核实settings的ExternalModelCLI校验经既有ValidateExternalCLI→registered_agent桥→swa.registeredAgent，在真实DB取得或创建local_authenticated_transport主体；authority须在该真实效果后取得，未手动写身份。实际82响应和持久效果仍待验。

a4d00884的79限定接受hook获主控`msg_30d1e0fa7235`明确全文读取报告、原件和独立结果后回复，唤起已确认；5ee4dd6c的故障消费者退回hook此前获`msg_2f239fc7cf0d`全文读回。本Task继续原产品和最终同安装双Worker审核。

## 82实际准备及83原请求未知

**接受82真实委托准备及同安装Observe消费的实际范围。** 原F72恢复入口沿同3bcf-install60/67服务，真实五笔space、project、settings、Grant、token请求均有独立原request_id、当前UTC及2xx回执，无自动重试；六项缺cookie、缺CSRF、跨Origin和Agent管理Grant/token/settings实际403。实际服务主体为human/local-human，执行者明确为原Agent Task/Dispatch，tty_input_used=false；服务cookie12小时且未撤销与本次一小时token分别记录，未伪造用户输入。

真实项目为`6d8cf620-f26c-4ee3-8554-1c9609075736`，固定公开目录，settings revision2、single OpenCode CLI、空tools、五动作observe/start/send/stop/read_context及当前目录/引用边界保持。实际authority为`e63abe4b00a394a833acc1e804e1a084775f62958a93a16015219c915fcf650c`，在原设置校验后由生产链生成，Grant/token前真实同生产来源完整取得；原scope没有手工写入或伪造。token实际到期UTC`2026-10-22T06:04:06.7927994Z`，与原签发请求时间加一小时匹配，未重签或续旧95ef Observe。

Reviewer独立`review-execute82.ps1`及其mjs实际0，原件`execute82-independent.log`：只读CurrentUser DPAPI并在内存核完整绑定和token，对实际库内唯一Grant的private_token_digest进行真实SHA256等值检查，秘密不进入命令行或输出；文件/父目录私有owner/DACL和无reparse核。真实project设置及Grant/actions/expiry/revoked_at、现存authority均匹配原回执；准备各阶段原三库观察只增加space/project/Grant各一，orchestration/journal业务状态保持，没有Worker或模型。当前83的后继业务状态与准备阶段分别审核，没有把后继记录强称为准备时零状态。

同安装真实status/capabilities及CurrentUser DPAPI Observe消费实际0，runtime/authority/registry/原生版本、provider/model/account_source逐项匹配，credential_renewed=false。实际安装发现Pi1.0.1、Codex0.162.0、Claude2.1.238、Grok1.0.34和OpenCode2.0.31；capabilities中实际launchable仍为unknown，发现或配置状态不登记实际执行。Reviewer在内存读取原选定DeepSeek凭据，核62份准备/消费者导出和prepared.json均不含该实际key或实际token，原native DB mtime保持；原Orca before/after保全passed，Reviewer没有发HTTP、签Grant、原生调用或模型。

**83产品流程当前证据不足。** 真实第二条CLI命令controller.open返回HTTP409、exit6、operation_unknown，原请求`3bed9c88-9c99-4cdd-83cf-a523e5383374`不可重发。Reviewer独立只读`inspect-flow83.mjs`实际0，原件`flow83-original-request-independent.log`：真实原R为swa.controllerOpen/pending、stage=pty-spawn，没有已登记controller/process/helper或子请求；真实Run仅原bootstrap，tasks/dispatch/worker/resource表仍0。这些记录没有证明PTY未创建或完整资源退出，未知继续保留，由原Worker查询同一原请求和准确所属资源并诊断。未重开Controller、重跑Flow或重签token，产品Worker、模型和最终双Worker均未接受；通知为`msg_12180190c336`。

## 83控制端启动的准确调用缺口

**退回准确`3bcf0c71ba86e607cd63aef3b423d7cc6f5788eb`的createController启动参数。** 独立沿准确源码核native-methods.ts、runtime-agent-terminal-startup.ts、terminal-create.ts、agent-launch-token.ts及daemon terminal-host-session-create.ts：本次调用只传launchConfig.agentCommand/agentArgs/agentEnv，没有实际command和env。无launchAgent时原选项解析直接返回opts，原PTY spawn只取launchOpts.command及launchOpts.env，不从launchConfig生成它们；daemon实际hasCommand也取opts.command，故83的无启动命令现象与源码一致。

launchConfig仍用于保存启动配置并触发原agentLaunchTokenFor生成真实启动身份。原Worker须在原模块同时提供按实际shell正确构造的command/env并保留launchConfig、真实token/pane/创建身份；原helper实际启动与身份握手需要真实验证。引用Node/入口路径按既有shell构造规则处理，不冒充注册Agent、用户输入或人工Actor。主控`msg_e96e80027fb4`和Worker报告原pwsh50820仍存在、helper未启动；这份资源观察的准确原件及完整所属退出继续核，原R保持pending，不手工写SQL完成或据表内缺少Worker推断无进程。

该缺口属于必要程序修复，准确新Commit改变安装输入后须新构建/安装并在同一新候选完成真实验收；已核未变native可继续复用。旧3bcf不能代表修复后的最终程序，82准备的限定接受及原95ef保全继续保持，不自动消费不匹配的旧DPAPI或重签未知请求。独立源码结论已送原Worker`msg_039e094ef4dd`、主控`msg_31b5696cd5b0`，全部修复由原Worker负责，当前Task继续。

## controller83准确窄修计划

读取计划正文前，Reviewer按真实83及准确原调用链形成C83-01—C83-06：真实shell command/env与原launch身份、真实helper握手及祖先/出生身份、未知R及完整所属资源保留、新准确候选和Grant/DPAPI来源、同候选最终双Worker与完整成果、必要失败条件及未验范围。**接受controller83-repair-plan.md准确SHA256 `79A74EC13A21FFE64B2106FDABD336D4F2177EF9BB70ED8ECDCA46F7483B4978`的返修方向及必要验证范围。** 正文完整读取；明确原command/env/shellOverride与PowerShell7、原quotePowerShellNativeArgument及EncodedCommand规则，保留launchConfig与原launch token/commitment、identify、同controller及严格Caller，未新增公开协议或权限。

计划沿既有controller/T13真实消费者，准确新程序构建/安装后核helper实际启动和R、再执行双Worker/指定停止/完整成果；未变四native/Node可按准确来源复用，没有新增自我证明模型。原83 R和3bcf/82独立保留，原67未知PTY先按现有inventory及出生身份观察，只处置同67整个所属服务/Job，取得两次完整成员和已观察身份退出证据后才安排后继来源。该范围不授权清除未知原R或伪造完成，具体实际退出与新来源/凭据仍独立核；Reviewer不关闭服务、签Grant或修改实现。

此前fde3a91b的82限定接受获主控`msg_46d6f90182f2`完整读回；83证据不足获`msg_3f82ad7241a6`完整读回；a3d37243启动缺口退回获`msg_c254bc57f789`完整读回，三项实际唤起均已确认。新准确实现、旧67所属退出、新候选及实际产品验收继续，不结算Task。

## 后继CI按实际平台核查

Ubuntu job119649240667实际检出合并源`356905dd1317d7006cb6cd53ad23932dbf5ab2ab`，父版本为main6e4ba81c及Worker5c75001e，frozen install成功、后续Go仍退出1。Reviewer独立`review-ci5c.mjs`实际0，原件`ci5c-independent.log`：与main50c2同平台原日志比较，失败名称同为68项，无新增/删除，Taste迁移前提失败已消失。此结果限于已读日志，不登记完整CI、build或e2e通过。

**Windows job119649240350的新增迁移前置失败原因证据不足。** 原完整日志与main50c2的Windows job119638842033逐项比较，失败名称81对79，新增为同一父用例TestRelayRepositoryCleanupActualSQLiteLifecycle及其actual_database_failure子例，不能沿用Linux的无新增结论。实际同一合并源356905dd的普通frozen成功，Go整体1；子例11.12秒，在cmd/server/relay_repository_cleanup_test.go:35—36的RunMigrations返回非nil处分支停止，尚未创建cleanup provider或主动关闭数据库。

Reviewer取得该准确Git合并对象并只读核源码：原清理测试、startRelayRepositoryCleanup和SQLite adapter.go与main50c2逐字节相同；迁移目录唯一增加028_orca_execution_links.sql，原27份保持。测试在TempDir/Open/全部迁移之前建立10秒context，RunMigrations本来返回带具体迁移及原cause的错误，但测试丢弃该错误，只输出固定失败文本。时限耗尽是与时长及源码相符的推断，原日志没有ctx.Err或具体迁移错误，不能据此认定028 SQL有错、清理行为回归或简单延长时限即可修复。

独立`review-ciwindows.mjs`实际0，原件`ciwindows-independent.log`，核完整日志、准确合并对象和上述相同源码/迁移差异；Reviewer没有编译、迁移或启动服务。已向主控`msg_ca0924d99290`报告必要复验范围：原Worker在控制端启动检查之后串行对准确合并源与基线运行原真实SQLite两子例，保留原10秒及业务断言；需要归因时只补实际返回err、ctx.Err及耗时证据后按原条件复验，不mock、改SQL或扩大旧S3。实际复验仍未执行或接受。

主控b0306af5的四份管理差异及QUESTIONS已完整读取，Git/成熟marked检查0、89个本地链接无缺失；旧Observe、原工具/S3/T11及82实际范围分别保留。其83父R与/native子请求表述已通知主控精确化：独立实际查询仅父R pending，所有后缀子请求为空。该文案随后继准确管理提交核，不阻断原产品修复。b159计划接受hook获主控`msg_4a29fd7e4cf8`全文读回确认，当前Task继续。

## 原67的EOF及所属资源观察

原Worker按已确认范围结束67的owner生命周期，实际退出6，原stderr保留delivery_unknown及完整所属后代未退出的错误；没有登记正常关闭通过。service67-close85a原件在EOF前记录六个实际出生身份和完整Job成员，之后两次均仍有daemon46744、OpenConsole55220、pwsh50820；原serve/host/http已退出。83父R/pty-spawn继续pending，缺少Task/Worker不能作为资源退出依据。

主控`msg_965a24b63bee`明确允许原Worker对该67具名Job再次核准确来源、CurrentUser+SYSTEM DACL和全部出生身份后使用TerminateJobObject，随后两次完整空清单及全部已观察身份退出证明；仅针对该所属Job，原95ef/5187和Orca保持。Reviewer完整读取close67-job86.ps1与ServiceJobObservation.cs的原Kernel API调用、完整成员查询和权限/身份检查，没有执行终止；86实际结果仍待审核，强制收尾将与原EOF6分别记录，不冒充正常关闭。

## 准确893fb885源码与67后继证据边界

**接受准确`893fb8852e15109ca9a99f1bd51d83ea249a37b8`的createController参数源码返修及84严格检查限定范围。** 三份准确差异已完整读取，唯一产品变化为原native-methods.ts六行：成熟PowerShell参数引用组成实际调用，按真实本地workspace预核短EncodedCommand投递，显式command/env/shellOverride/fast，保留原launchConfig与token/identify/祖先/出生/pane/strict caller。实际引用模块的PowerShell分支及长度限制、环境传递、helper和原RPC准入调用链已完整核对。源码修订不等于实际helper握手或最终产品通过。

Reviewer独立`review-controller84.mjs`实际0，原件`controller84-independent.log`：SOURCE=5c75001e及完整source.diff与准确893fb885程序差异一致，strict-exit0，实际工作树相关源码等于准确Commit；没有重复TypeScript编译。准确两份Markdown的Git/marked检查0，六个本地链接及旧§15锚点保持。build19原入口要求准确干净Commit，已读取SOURCE及真实构建入口，输出指向独立dist/swa-cli-install-20261020-18，未变Node/四native继续原准确来源；实际构建/安装/握手结果尚待完成原件。

**原67完整Job退出继续证据不足。** Worker及主控报告86重新打开原具名Job为ERROR_FILE_NOT_FOUND，未执行TerminateJobObject，也未创建替代Job；不能把不可查询改记为完整空清单。原83父R仍pending，85a的三残留与EOF6保留。主控`msg_587f117df8b1`明确限定后继为同67真实daemon的listSessions、原session/incarnation/cwd/出生来源核对、一次原kill及原空闲退休，再核实际退出；原Job不可查询的限制仍保留。Reviewer完整读取cleanup67-daemon87.ts及其真实源码打包入口：连接原PID/launchNonce及服务归属的原DaemonClient，没有自动创建runtime；两次实际inventory及出生身份检查后只杀指定原session，空清单后才请求shutdownIfIdle，未改R或SQL。实际原生清理及全部已观察身份退出结果仍待核，该有限证明不替代已缺失的完整Job证明。

原67未知与新候选分别保留来源和资源事实；准确新构建可继续，实际后继服务和Grant须满足主控限定的原所属清理及独立新DataRoot/runtime/来源条件，不错配旧82 DPAPI，不宣称旧EOF恢复为正常退出。73835fe0的Windows证据不足hook已获主控`msg_4e8b24f4a033`全文读取准确报告、原CI和独立结果后回复，实际唤起已确认。Windows定向SQLite复验由原Worker在控制端检查后串行执行，未改原时限或扩大范围，当前Task继续。

## 87e原生清理的有限退出证明

**接受87e对原67指定session及全部已观察出生身份的实际清理范围。** 原DaemonClient的kill和shutdownIfIdle及真实路由已核。87b在bundle的ESM格式处失败，87c在真实inventory的cwd=null比较处停止，87d原getCwd也返回null并停止，均未到kill-intent；原件保留。主控`msg_b5063d6bdb78`明确以真实daemon PID/nonce/原服务归属、原terminal/incarnation及全出生路径/两次一致库存判定所属，cwd缺失如实保留，允许只对原session一次原kill及空闲退休，没有伪造路径、协议或Actor。

87e实际清理退出0，终端为`term_afd4ccc0b217bd82cd4c843cbf8ed114`、incarnation为`1dd634e4-7c23-4b3d-92db-b0651cdbc8d5`、PID50820；其原session/所属与原83 expected_controller匹配。真实kill返回空对象，后继原native inventory为空，shutdownIfIdle返回retiring=true；原daemon46744、OpenConsole55220和pwsh50820的实际持有句柄退出码分别0、0、1，均已退出。两次完整CIM查询原六身份均缺席，保留original_job_query=unavailable；实际cwd仍null，不将sessionId中的路径称为观测cwd。

Reviewer独立`review-cleanup87e.mjs`实际0，原件`cleanup87e-independent.log`：真实inventory、动作意图、原kill/退休回执、三句柄及两次六身份观察与85a出生记录一致，原Orca前后保全passed；只读真实旧库中的83父R仍为pending，receipt全部字段保持，没有SQL补结算。独立`review-cleanup87e-current.ps1`实际0，原件`cleanup87e-current-independent.log`于UTC05:48:52再次核原六出生身份均已不存在，Reviewer没有发清理调用或终止进程。原FD5由Worker前后实际检查，原95ef/5187保持；主控`msg_01e2506e06b3`已明确该已观察范围可承接独立新服务。

该接受不补足原具名Job的全成员退出证明，旧EOF6、86不可查询及83父R pending继续保留；不把原正常关闭或历史资源处理全部登记为通过。新候选仍须独立准确安装、服务/runtime、真实项目/authority和Grant/DPAPI，不复用或覆盖旧82绑定。

准确893fb885的build19实际退出0，SOURCE、完整入口及日志已读，Node24.21与四个原507/build11 native均按原证据复制，没有重编native；实际闭包28251路径，原检查明确native认证和功能验收NOT_RUN。普通归档/提取安装88及服务89入口已读，安装和真实helper结果仍待完整原件，不将构建成功计为T13通过。

5eab3922的893源码限定接受与旧67退出证据不足hook分别获主控`msg_be0efa8926f7`、`msg_064436444de9`完整读取回复，两项唤起已确认。本Task继续新候选的真实验收。

## 893安装88、服务89与准确来源90

**接受execute90-review-input/actual-inputs.json冻结的准确893来源、四份既有委托客户端和89真实服务前置，范围为既定新91准备及后续原调用链入口。** 四份客户端摘要与79已接受正文相同，准备客户端为F72D42D2；guard、消费者、observer分别仍为DEA336E1、3D4B0940、F0DDAFC5。没有新增权限或人工确认条件。新数据为`893fb885-service89`，新runtime为`6aec5c49-4077-4376-8e4d-8dc5b83acc77`，新安装为`893fb885-install88`；旧82的项目、authority、cookie和DPAPI不作为本次新绑定消费。Reviewer没有运行准备客户端、创建session、签发Grant或调用模型。

Reviewer独立`review-execute90.mjs`实际0，原件`execute90-independent.log`：四客户端及三同生产bundle/metafile实际摘要、入口和准确893相关干净源码均核；实际source前后逐字节一致，authority_domain/source_snapshot仍null，最终delegated-execute/version-home尚不存在。真实只读三库的业务计数与90原观察相同，workspace五表全0，orchestration仅原inspect-only bootstrap一行，其余六表0，journal十四表全0，认证身份0；DB/WAL大小和mtime保持。没有把真实迁移bootstrap要求为全零或提前生成项目。

Reviewer独立`review-execute89-service.ps1`实际0，原件`execute89-service-independent.log`：当前serve6224、host33976、http55368的创建时间分别为UTC05:51:07.2867800、05:51:08.3431430、05:51:11.6592370，父进程分别13636、6224、6224，路径均属于准确新安装；64508监听、metadata、VERSION/CLI版本、health、私有owner/DACL/无reparse、host SHA256 `2ba1146a5662b3aef2fac0ef7040d7308aec28785769c729fdb366819c012493`及三份同用户Medium原身份记录匹配。原服务工具32198与只读Job句柄21806保留，Reviewer不写入或结束它们；原95ef/5187保持。hold89-job.ps1完整读取，仅持有QUERY和READ_CONTROL句柄供所属服务结束后两次完整成员观察，当前尚未登记退出通过。

88普通.NET归档/提取实际0，归档412666860字节、25502条目，原archive SHA256为`74BA57A4C40D5614D2421607029C9046DAFFAC430ADF829F171E7AB28B509745`；独立ZIP元数据检查核关键唯一入口及准确VERSION。原安装T02聚合退出6，真实T03/T13尚未完成，不能把包提取和服务存活登记为完整安装验收。93准确干净893的四个Go消费者test binary编译原入口、SOURCE、测试列表、outer及退出0已读；此处只登记实际编译，真实四用例结果尚待原件，Reviewer没有重复编译。

本轮独立前置接受已先送原Worker`msg_3974ee636733`、主控`msg_c65a966ab1e3`，允许按原授权连续91及无模型controller/helper/Idempotency，再核同安装双Worker。真实Grant、DPAPI消费、helper握手、产品Worker和最终T13仍待实际证据；没有将原83未知父R、旧EOF6或缺失Job证明改为通过。2197a441的87e限定接受hook已获主控`msg_3191bfa6192c`完整读取回复，唤起已确认，本Task继续。

## 准确893的后继双平台CI

Reviewer完整读取Ubuntu job119658661466及Windows job119658661639原日志，实际合并源码为`fce64fb1533f8224f2617d8313d351a90657939e`，父版本为main b0306af5及Worker893fb885。普通frozen install两平台成功，后续Go整体均退出1，完整CI仍失败。独立`review-ci893.mjs`实际0，原件`ci893-independent.log`：同平台main50c2基线比较，Ubuntu失败名称同为68项，Windows同为79项，新增与删除均0；先前5c750的Relay父例及actual_database_failure子例本轮没有出现。

实际两个合并版本356905dd与fce64fb1的cmd、internal、migrations、go.mod/go.sum及pnpm-workspace.yaml逐字节差异为空。这支持本轮没有观察到该新增失败，旧5c750具体原因仍未知，不能称为已修复、全部失败同因或完整CI通过。主控`msg_ca10be70ec49`及`msg_3191bfa6192c`已取消尚未开始的额外两源码SQLite复验和诊断改写；当前继续产品关键调用链，不增加重复测试或管理等待，历史失败原件保留。

## 91真实委托准备与Observe消费

**接受准确893安装88/89服务的91新委托准备及实际Observe消费范围。** 原入口实际0，五笔space、project、settings、Grant、token写入各一次，真实请求ID、执行时间和2xx回执完整；六项缺cookie、缺CSRF、跨Origin及Agent管理Grant/token/settings均实际403。服务返回主体仍为human/local-human，实际执行者为原Worker Task/Dispatch/terminal且tty_input_used=false；12小时服务cookie未撤销与本次一小时token分别记录，没有伪造本人输入或身份。

新项目`5f596c58-e7ce-429c-b505-124821055cee`属于原固定公开目录，settings revision2、五动作、single OpenCode CLI、原生tools为空、目录/引用边界保持。真实authority为`6b276bc4777128a42992eb317e79a594403d1d1760f7b891b82d1ea6b0264b31`，在生产身份桥后取得；token实际到期UTC`2026-10-22T06:57:49.3951688Z`，与本次真实签发时间加一小时匹配。该准备使用893/新DataRoot/新runtime完整绑定，没有复用82或续签95ef旧Observe。

Reviewer独立`review-execute91.ps1`及mjs实际0，原件`execute91-independent.log`：只读CurrentUser DPAPI并在内存核实际token与真实Grant private_token_digest等值、完整来源/项目/actions/expiry绑定及私有owner/DACL/无reparse；真实库中project、唯一Grant、authority与原阶段回执相符。91各阶段三库记录只增加真实space/project/Grant各一，orchestration和journal业务状态保持；后继92与91阶段分别记录。实际status/capabilities和Observe消费0，credential_renewed=false，registry、2.0.31、DeepSeek-pro及account_source一致，实际launchable仍unknown。实际发现Pi、Codex、Claude、Grok、OpenCode安装，不据此登记全部Agent已调用。

Reviewer只读实际选定DeepSeek凭据并在内存核62份导出及prepared.json均不含实际key/token，原生库mtime保持，Orca前后保全passed；没有调用HTTP、签Grant、原生Agent或模型。限定接受已送原Worker`msg_710bed268243`，继续原无模型调用链及T13，不增加确认。5481497d的89/90来源hook获主控`msg_43d37b92aa26`完整读取准确报告和两个独立日志后回复，唤起已确认。主控准确64451493的HANDOFF、STATUS及QUESTIONS差异已完整读，授权及旧限制保持。

## 92实际helper链与消费者Run表缺口

**接受92已经完成的真实controller/helper及Run/root Task限定范围；退回准确893测试消费者的Run查询，完整幂等验证尚未通过。** 原92驱动实际1，在native-idempotency.mjs:217以coordinator_runs计数0比较预期1处停止，尚未执行后续关闭；原失败输出、请求与资源保留。准确生产run-create.ts写入runs，coordinator-run-store.ts是另一条独立旧协调状态，不能为错误测试修改产品写表。

Reviewer独立`inspect-idempotency92.mjs`实际0，原件`idempotency92-binding-independent.log`：真实原open `febc0b31-069e-4b08-8972-2afb86597935`、task `c1504651-6a55-4fc8-8fb5-0bc48116cf19`及其/run-create、/root-task四回执均completed。Controller为`term_e86fd2e2c7caf2a0383ec5c3e39aaaa4`，实际PTY53564/出生1792648746942、helper56260/出生1792648748228、真实binding/incarnation/pane与回执相同；实际runs唯一`run_66c11beca641`绑定该controller、generation1、legacy0，root `task_0613700f941d`恰一行且parent_id=null。并发及完成重放回到同controller/Run/Task，四项实际workdir/spec/context变化拒绝request_mismatch；Worker记录为0，本次没有模型。该限定观察不替代后续完整关闭、全部幂等失败路径或T13。

同一准确测试域另有native-idempotency的storage计数、controller-recovery计数、runtime-restart的计数和原Run读取、cli-failure的计数和原Run读取误用coordinator_runs。CLI消费者随后访问beforeRun.coordinator_handle，restart恢复随后访问run.id，真实缺行将抛错停止；这些入口不能执行既定验收，应由原Worker集中修正真实runs查询并保留记录存在、身份/代际、唯一root及全部原失败断言。controller-inspection.ts已另有SELECT * FROM runs及唯一inspect-only bootstrap/legacy/generation0校验，其旧协调表零计数仍适用，无须因此修改。准确范围已送原Worker`msg_6472dfb88468`及主控`msg_f9bb5d6efe20`。

已完整读取并独立核新增ControllerClose消费者SHA256 `F7B9C7F430B02B694BDB3330592787381E7434A6A314529D29184012AA0062F7`，限定接受原92真实open响应、completed原R及当前同controller完整owner/出生/pane核实后的单新close R，原DPAPI/来源/stop授权保持，未知响应不重发。前置intent须保留当前UTC、原Task/Dispatch和阶段，原Worker在原客户端补齐后可连续执行，最终准确SHA和实际关闭结果另核；原89服务及只读Job/95ef保持。当前没有登记该关闭或完整幂等通过；纯测试修订不要求重编未变893程序和native。

## 准确f35消费者修订与94a原92关闭

**接受准确`f35beaaedf4598ddce413c59da6792341e9eb0aa`的四项真实Run消费者及来源检查修订。** 七份完整差异已读：四mjs现读取runs且legacy=0，分别核实际controller/pane/consumer_generation及唯一root，记录存在立即检查；旧coordinator_runs作为独立状态继续计数，原并发、幂等、context变化、恢复和故障断言保留。来源检查仅将未进入实际安装metafile的native-idempotency加入既有明确测试排除域；controller-inspection及全部程序/Go/构建/原生输入与准确893保持。

Reviewer独立`review-f35-source.mjs`实际0，原件`f35-source-actual-independent.log`与`f35-source-independent.json`：准确干净f35、七条差异路径、实际相关源码、全部必要产品输入相同、成熟YAML原六项策略及实际metafile排除均核，没有编译或原生调用。准确两份Markdown的Git/marked检查0，六个本地链接及旧§15锚点保持；主控64451493准确四份事实源差异已全部读取，89个本地链接无缺失。

**接受94a按既有stop权限完成原92 controller及已观察helper出生身份退出的限定范围。** 最终消费者SHA256为`D646D5AC562ACE274EF8D6E9D84840C2CA7B1FA639D1CC2F8F41CF906FFA6FB3`，已核实际摘要和Parser0，close-intent含真实UTC、原Task/Dispatch和阶段。94实际1在原stdout的Protected DACL前置停止，尚无close-intent或关闭请求；原文件owner为CurrentUser、仅同用户继承权限，未设置Protected。原92字节和DACL保持，94a以CreateNew保存同字节私有副本并设保护DACL，摘要两者均为`5805128B044A46687EB1DBAF571EF5B1D7A61D348C069CA6D45C9E555FEB85DE`，原件和94失败保留。

94a单次新R为`29d672ec-e3cf-4e71-a53a-5d35356c1c41`，实际关闭0、原request.show为completed；当前owner与原open的完整binding、两出生身份、pane及项目相符。Reviewer独立`review-close94a.ps1`及mjs实际0，原件`close94a-independent.log`：私有副本/原继承ACL、字节和摘要、原open及真实库内close回执完整核，两个实际PID缺席样本及Reviewer当前出生身份缺席通过。关闭前后来源、原项目/Grant及其他业务状态保持，仅新增一个close receipt；没有签Grant、重发未知R或调用模型。该结论不代替完整Service Job或后续Worker资源退出证明。

## 95同安装非模型幂等实际范围

**接受95在同893-install88/89服务上的既定非模型幂等检查。** 准确f35消费者使用原91绑定，实际驱动0，全部24条原CLI输出、实际退出/时间/request_id已独立核；并发open和完成重放同一controller，并发task与完成重放同一Run/root，close并发及完成重放完成，另一controller也经实际关闭。五项workdir、spec、真实context/版本及其他controller payload变化拒绝request_mismatch，foreign controller查询原任务拒绝scope_denied；原请求和两级真实子回执保持。

Reviewer独立`review-idempotency95.ps1`及mjs实际0，原件`idempotency95-complete-independent.log`：真实SQLite七个原请求全部completed，实际Run按controller/pane/generation唯一、root恰一、固定context与真实版本变化拒绝相符；真实workspace执行关联准确指向当前项目、Grant caller、settings2及同Run/root。95阶段前后仅新增Run、root及该执行关联各一、receipt七，其余三库计数和来源保持；这些阶段原件与后继Flow96的当前业务状态分别审核。两controller各有两次实际进程出生身份退出样本，58份导出不含真实DeepSeek key或本次token，原生库mtime保持。Reviewer没有重跑95产品调用。

本轮接受限于原三个无模型幂等案例。实际逐效果边界响应丢失、runtime重启pending查询、Worker start/send并发与断连重放、不同R的close/create/start交错四项仍未验，不以24条命令或编译结果扩大接受。原92失败、94前置失败及全部原件保留，Flow96沿同安装进行真实双Worker验收，模型、重叠、指定停止和完整成果仍待实际证据。原95ef/5187及89服务/只读Job工具保持，未调用Aliyun；2b64fcfb两个hook已分别获主控`msg_12e573547662`、`msg_f4c1f132120d`全文读取回复，唤起均已确认，本Task继续。

## 96真实JSON错误及并发完整输入缺口

**退回准确893程序的helperCall同PTY并发完整输入边界；Flow96的产品Worker、模型及T13尚未通过。** 原96驱动实际1，CLI5的worker.start实际6/operation_unknown。两个原父R分别为`9c6159d2-c507-42e7-88fa-c53b0092543f`和`6bb88711-c5ea-4743-8d48-c92fb041ff84`，均为swa.controllerDeliver/pending/helper-input，两个对应/native子请求尚不存在，原件及未知效果保持，不能重发或SQL补完成。

Reviewer完整核原inspect96-pty.ts的真实DaemonClient连接、原PID/nonce/Job归属及session/incarnation/PTY匹配。UTC06:13:45.172的实际文字快照包含helper49128的identify、observe和task回复，随后明确记录JSON错误：Unexpected non-whitespace character after JSON at position 12431。独立`review-flow96.mjs`实际0，原件`flow96-independent.log`：原controller `term_d420f723d67bd4a40e7ac17b83cabcab`、PTY50264/出生1792649388051、helper49128/出生1792649388479及原binding一致；当前真实库核上述两原R和两个缺失子请求，worker_dispatches、worker_terminal_resources、dispatch_contexts计数均0。这些观察不证明全部系统资源缺席或完整Job退出；实际OpenCode执行和模型尚无证据。

准确helperCall使用普通sendTerminal(text, enter=true)；原terminal-input-dispatch和RuntimeTerminalWriter完整读取，前者没有对普通发送使用现有Agent prompt串行机制，后者将text分块、异步submit delay和Enter后缀分开执行。两个调用可在首个Enter前写入第二个JSON，或在多chunk间交错，缺少同PTY完整JSON请求的互斥。真实解析错误与该缺口一致；没有捕获的每字节调度顺序继续未知。helper中JSON.parse位于单条execute的try之外，原外层错误路径设置exitCode1，不能靠吞错继续消费或改串行驱动替代产品修订。

读Worker计划前已独立形成C96-01—C96-07，覆盖完整JSON/chunks/后缀互斥、原真实身份及成熟机制、拒绝/部分交付/unknown及生命周期、原96所属资源、真实同controller双start及完整T13、新准确程序/安装/绑定、实际输入大小及失败窗口。准确计划`60F009037C32CD695931C6529A903097C84207A1A9CA2A1A65421FA71105ABC0`全文及原serializeAgentPromptSubmission实现已读：复用现有perPTY/generation队列的方案方向成立，范围维持三个原文件及现有协议/身份。

**退回60F009计划的两项覆盖缺口。** C96-02仍将实际队列覆盖留到最终T13，须补主控已要求的最终归档前针对性真实服务/SQLite/PTY/原调用链检查，记录96两输入实际UTF8长度、完整原上下文、同/跨真实chunk并发及各自JSON/request_id完整到达后再正式包装；不能mock、伪造身份或截短输入。原writer只在suffix delay检查signal，writeChunks不查；须明确队首和每chunk/后缀之前核实际signal、handle/PTY/generation，未写超时请求不晚发，已部分交付保持unknown，订阅/定时器/排队Promise有确定结束条件。未命中的真实失败窗口保持未验；没有要求重复未变模型/native或新增调度系统。

上述准确返修范围已送原Worker`msg_9748a38e6e22`及主控`msg_a2c5414c736a`，原Worker负责补计划和全部实现，不修改893服务或错用91到新来源。原89服务工具32198及完整Job只读句柄21806保持，具体所属处置由主控安排；原95ef/5187及Orca保全。ed972b49限定接受hook获主控`msg_c5abee194210`全文读取准确报告及三独立原件后回复，唤起已确认，当前Task继续。

## 96准确修订计划及归档前真实验证

**接受准确`44DDD0E5B03039D95A1D36711E54F9BCBDEA6A4FD32128E3C9D79CC529273DB2`的controller96-write-plan.md修订。** 全文及实际摘要已核，对照先行七项独立条件，补齐成熟perPTY/generation机制覆盖整个writeAction、队首及每chunk/后缀的signal和当前身份复核、异步原钩子前后复核、同时持有真实写入settlement和回复的Promise、原helper解析失败后的readline/stdin收尾。原请求/NDJSON/schema/身份和unknown语义保持，未命中的取消、部分交付及生命周期窗口登记静态范围与NOT_RUN，不能冒充动态通过。

严格及架构检查后，先以准确未压缩成品经原install脚本进行独占component安装和真实新来源准备，再在正式归档前运行真实传输检查：同controller两次show并发、超过真实16KiB chunk的公开a/b正文selected_context之task.create与show并发；核各自原R/真实回复及完整SQLite context正文、Run/root/两child、所属退出和保全。按实际序列化输入长度判断跨chunk，没有达到时立即停止且不计通过。主控随后`msg_e7676f1f9ab2`明确要求开发候选真实两Worker诊断也先于正式ZIP；该最新顺序已送原Worker`msg_23c625c250c0`继续执行。开发结果分别记录，最终ZIP/普通安装仍须同候选真实双start、模型重叠、指定停止及完整成果，不以component结果替代T13。

原analyze96.mjs及flow96-diagnostic.json已完整读取，Reviewer补充只读核`review-flow96.mjs`实际0，原件`flow96-length-independent.log`：按真实两pending receipt内invocation及原JSON序列化，两帧分别12431/12480字符、12719/12822 UTF8字节，均低于原16KiB chunk。首帧字符数与真实JSON错误位置12431相同；这是原存储输入与错误位置的关联，实际每字节写入顺序仍未捕获。没有据此登记跨chunk已经复现，也没有构造替代请求或重发原R。

主控已明确原89准确所属处置授权：原stop下可单新close R处理原96 controller；原owner EOF若残留，仅在重核原893来源、serve6224出生、私有Job DACL和全部成员出生/路径后处置该具名Job。21806保持至两次完整空清单及全部旧身份退出，原95ef/Orca保持，原未知R不变；Reviewer只审核实际原件，不执行关闭、终止或新Grant。准确新程序及测试结果继续由原Worker提供，当前计划接受不登记程序修复、实际component或T13通过。

## 准确6778源码与101静态检查

**接受准确`6778a66faf6d1ca7b50658f23966db3e406f38c6`的三项产品源码返修、两项必要消费者及101严格/架构检查限定范围。** 七份完整Commit差异已读。普通sendTerminal仅在driving且requireWriteSettlement路径进入原perPTY/generation队列，覆盖全部text/chunks/后缀及实际write settlement；队首和原beforeWrite异步钩子前后检查signal、handle/PTY和generation。回复等待位于该队列之外，实际模型并发仍须真实区间验证；其他输入路径及原chunk算法保持。

helperCall以局部AbortController接原120秒deadline，Promise.all同时持有写入和实际回复，finally取消写入并释放订阅/timer；未写排队项到队首失败，部分交付仍unknown，不重发旧R。helper保留原JSON/schema及错误边界，在finally关闭readline/stdin；原正确引用的PowerShell启动命令追加exit原LASTEXITCODE，处理原-NoExit下Node结束后的shell生命周期。原launch token/identify/祖先/出生/pane、Grant/工具和真实固定来源保持，没有引入新调度器或协议。

controller-fault新增显式helper-only模式，沿原实际owner、Job、父链及持有句柄验证，仅给helper终止权限，root句柄只读等待，分别记录实际终止PID和两目标退出，默认控制端/服务模式保持。Flow后续输入以该次实际send R命名并CreateNew，完整正文、请求及原业务断言保持，支持component和最终安装分别保留原件。这些属于主控NoExit生命周期及两阶段真实验证要求；动态helper-only结果尚未取得。

Reviewer独立`review-controller101.mjs`实际0，原件`controller101-source-independent.log`：101 SOURCE=f35及原三产品source.diff与准确6778一致，工作树干净，五份产品/测试源码等于准确Commit；严格退出0、47包架构退出0。101入口及原日志已读，目录内实际fault binary存在，原Worker报告Go构建/格式和Flow语法0；此处未将其存在或报告当作实际故障通过。Reviewer没有重复编译。准确两份Markdown的Git/成熟marked检查0，六个本地链接和旧§15锚点保持。

计划准确细化`752D733059F1DED177635D0E188CD03DC9D08EC540B543A14F2F64745F49F1E0`全文及摘要已核，41734C八个真实公开文件具体化和helper-only/Flow原件保全均在先行七条件范围内。component真实长输入/并发、开发候选真实双Worker、正式ZIP/最终安装同候选T13顺序保持；新程序须新准确来源及绑定，未变native可复用，未命中的取消/部分写/parser窗口仍NOT_RUN。源码限定接受已送原Worker`msg_0862612fff91`，允许连续原计划，不增加通用批准等待。

## 原89的明确原生观察关闭入口及102拒绝

**接受93C7D5CF4F2CC38C8B0A1B6362E82E47E89C95F55E6A481FB08D28AAFB4380A8消费者的明确native_inspection关闭分支。** 新辅助文件assert-controller-inspection.mjs准确SHA256为`6D85A960158637BC58231D9D50F372B48D23FFD818B00FB17CB4822477F3595D`，两个正文及原同生产inspection、Job schema均完整读取，实际摘要与独立Parser0已核。分支显式指定真实原生观察，核原DB、source、DataRoot/安装、原响应路径、60秒新鲜度、七个完整owner字段和实际session/incarnation；再用当前CIM核两出生、准确image及父链，沿原open completed和DPAPI/Grant范围单次新close R。默认94a入口保持，没有fallback或修改产品准入。

Reviewer只读当前50264/49128的真实CIM，父进程分别3824/50264，原出生及实际pwsh/893安装Node路径相符；没有发关闭或修改ACL。原89 Job为`Local\SWAService95379a87-5c72-4127-acd5-deea6645637c`，100原完整成员七个及出生/路径/CurrentUser+SYSTEM保护DACL已读，21806继续持有。该客户端限定接受已送原Worker`msg_4095845b3298`，实际关闭按独立原件审核。

**102没有证明原89 controller或服务关闭，后继退出证据仍不足。** 新close R `bb82920e-2b18-41f6-8178-8a5356bbb39f`真实返回operation_in_progress，消息为原Task或Worker创建仍持有reservation；驱动1保留。独立`review-close102.mjs`实际0，原件`close102-independent.log`：真实close receipt缺席，原96两start父R仍pending/helper-input，当前原生owner观察此前通过；不把该拒绝称为成功、未知R重试或权限放宽。原Worker按主控已授owner EOF及必要同Job核实后处置继续，完整空清单与全部旧身份退出另验。原95ef/5187、Orca和全部原R保持；246ac0cf计划hook获主控`msg_12e909d2894b`全文读取回复，5aa退回hook获`msg_6222198ed5b6`全文读取回复，唤起均已确认，本Task继续。

## 103原89具名Job退出及当前通信限制

**接受103对原89准确具名Job和全部七个已观察出生身份的实际退出范围。** 正常owner EOF实际6保持；之后按主控msg_99db5f427819原授权，重核893安装/VERSION、原serve6224出生及daemon所属记录、CurrentUser+SYSTEM保护DACL、完整成员/出生/父链/路径和实际持有句柄，仅对`Local\SWAService95379a87-5c72-4127-acd5-deea6645637c`调用TerminateJobObject。该次实际调用为UTC06:39:08.7837289，后继原件退出0；被终止的daemon3824、OpenConsole8948、PTY50264、helper49128持有句柄均实际退出1。

Reviewer完整读取close89-job103.ps1及原100/103原件，独立`review-service89-close103.mjs`和`review-service89-close103-current.ps1`均实际0，原件`service89-close103-actual-independent.log`、`service89-close103-birth-independent.log`：两次完整Job清单均为空，全部七身份两次缺席，原21806独立只读观察的两份完整空清单同时核对；终止前出生时间用.NET DateTimeOffset按实际ticks核对，当前CIM亦确认原七出生身份缺席。103阶段Orca前后检查passed，Worker原FD5前后检查保留；Reviewer没有执行关闭、终止或新Grant。真实库内原96两start仍pending/helper-input，102 close receipt仍缺席，没有把强制资源退出登记为正常EOF或未知R完成。

**当前原Orca可用性与后续真实验收证据不足。** UTC07:08:25只读观察原Orca主PID6228缺席，原子进程12576及daemon32280仍存在，原`C:/Users/DW/AppData/Roaming/orca/orca-runtime.json`缺失；原生heartbeat、check及escalation均返回runtime_unavailable，未确认消息投递。103时段的Orca保全结果与当前故障分别记录，当前故障原因及发生过程未知，不能由103历史通过推断当前可用。Reviewer没有重启原Orca、终止其残留进程、读取或修改用户作品。

现存component104-checks只记录SOURCE=6778及工作树差异的严格/语法检查退出0；其入口和退出原件已读，两个后继测试文件尚未提交，未据此接受准确新测试源码。新候选构建、实际component传输/helper-only、开发双Worker及最终ZIP同安装T13仍未完成独立验收，新来源/Grant绑定不能由原91替代。须恢复原Orca运行时及主控原生通信后继续这些依赖条件；当前不签Grant、不运行模型或处置其他进程。f50ef850的6778限定接受及102证据不足hook此前分别获主控`msg_4010b45320f0`、`msg_a8ca5119cb89`全文读取回复，实际唤起已确认；本段103结论的原生投递受当前运行时故障限制。
