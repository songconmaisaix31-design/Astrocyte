# Works Relay v1.4 语义修改与真实验收

2026-10-15 当前状态：IMPLEMENTATION_AUTHORIZED_OPEN_GATES。总负责人消息 `msg_0ba76f2d26c3` 授权修改语义层、真实 CLI、今日采集并推进 L/S；`msg_c0b5fd86176f`、`msg_2cf72d3f5ad9` 和正式答复 `msg_3b809b065867` 更新完成通道及验收顺序。当前顺序为四树普通集成与组合验证、集成版真实 L1、L2 经验管道及六步端到端首跑，目标 10-22 并入包。Calls 单位及权限完整覆盖仍待答，不开始登记 t0 或模型输入。L/S 的其他模块和正式 V4 日程由总负责人协调，本模块只交真实安全派生结果。

本文是本次四项语义修改的唯一具体规格。业务基线继承 [Works Relay v1.2](Works-Relay-v1.2.md)、[SPEC](../docs/SPEC.md)、[发布任务](Release-v0.1-20261014.md) 和 [REA研究报告](../docs/research/orca-rea-20261014.md)；最新用户决定通过总负责人正式消息更新本文件。模块分工、精确 SHA、运行参数和验收事实记录在 [模块管理文档](../docs/modules/relay.md)。

## 已授权与待答

已经授权实施：分别记录执行权、输入接受、回合开始、人工目标完成；记录真实控制器及正选原生来源的归属；区分初始化、会话确认、输入通道提交与客户端消费；保存 new / resume / fork / handoff 的真实关系，旧缺证据保持 unknown。全部沿既有 store、operation、binding、收据与事件实现。

总负责人已向用户询问以下关键决定，答案未到时保留当前状态机，独立工作继续：

| 待答项 | 当前约束 | 答案影响 |
|---|---|---|
| working 是否严格要求明确开始证据 | 不据 stdout、Git、PID、pipe 写入或完成结果产生新的开始证明；不冻结依赖该决定的接口 | 哪些真实客户端证据允许 accepted → working，以及 UI 的具体条件 |
| Calls 的计数单位及 Known 覆盖 | 保持 CallsKnown=false；completed Responses 只作私有独立观察量 | 区分已完成 response 与全部尝试 / 失败 / 重试的分母和预算语义 |
| 工具与外部目标的完整集合门禁 | 保持 ToolsKnown / TargetsKnown=false 及当前 Claim 拒绝；不以 argv、tool0 或空数组填已知 | 已授权文件 / 命令关闭、外发策略与完整能力集合各自需要的证明范围 |

接口的公共字段由原 R/H 根据已授权证据语义和用户答案在现有契约中共同维护；确定后交原 F 消费。本文中的证据含义明确，未决定的字段名称、枚举及迁移规则不作为冻结契约。

人工完成已由总负责人正式答复 `msg_3b809b065867`：权限成立且操作者明确确认目标完成时，accepted 或 working 均可进入 done，使用 `user_confirmed_complete`，缺少 started 继续 unknown。`verified_complete` 须关联本任务合法 owned 输出及固定目标实际检查通过的证据，不能由归属或原生回合结束单独产生。user_confirmed 任务正常进入经验管道并保留通道和确认出处；资源释放独立核验，仍沿 done → released，不新增 accepted / working → released。

总负责人正式答复 `msg_9e32e1f32ba5` 确认后两项已向用户提问，当前未决定。答复前不开始三分钟登记窗口，不绕过原 Claim / 预算发送输入，也不扩大工具权限；原 R 继续适配与独立检查，H 继续已有 Context / S3 / S4 契约、生产装配和无模型负例，F / UI 继续已交接口。真实完成 response 可以私有记录其实际身份和观测数量，不能转换为全部尝试次数；六字段 session-events 不变。

## 1. 投递阶段与人工完成

每次动作保留既有 operation 身份和幂等规则。store 首先原子取得该动作执行权并记录 pending；只有实际持有执行权的调用进入 driver。同 operation 的重复或并发请求按既有收据读取和互斥规则处理，不能再次提交输入。

输入通道提交、客户端接受输入、owned 回合明确开始、人工确认目标完成分别记录事实、实际时间、证据来源和确认程度。未观察到的事实保持 unknown；提交到 pipe 或完成 EOF 关闭只记录该通道事实，不能产生客户端消费确认。原生 RPC 回复、原生通知和 controller 自产档案按各自实测语义归类。

最终动作收据不可重写。后到的开始、结束和观察结果沿既有事件和 operation 关系追加，不将第一次失败、`delivery_unknown` 或 provider 错误改为成功。旧输入不自动重发；未知交付继续占用工作区，原进程和 pending 未核实之前不得释放或由第二执行者接管。

task 保留 ready → accepted → working 的执行状态；权限成立且明确人工目标确认时，accepted 和 working 均可进入 done。accepted 表示既有人工派发动作已被接受，不能代替模型开始或固定目标完成。working 的严格证据门禁按上表保持待决定，后到 done 不补造 started。

done 分别保存 `verified_complete` 与 `user_confirmed_complete`。前者保存合法 owned 输出关联及固定目标实际检查通过的证据；后者保存有权限操作者的明确目标完成确认。只有 CompletionConfirmed 而无实际目标检查事实时归入人工通道。沿既有 RelayService Transition、收据、事件和人工确认端口增加必要证据接缝，不以原生 turn 完成、进程退出或 Git 活动产生目标通过。released 继续从 done 进入并独立核真实退出、pending 为零、对象所有权和权限；中性工作区仍可以 Prepared=true，但 ActiveTaskID 必须解除。失败或取消沿现有终止清理路径保留审计，不通过伪造 done 清理资源。

保存 intent 时绑定真实项目 settings revision 和 permission level，确认及后续权限敏感动作重新核版本、A/B/C、操作者、对象所有权与 grant 状态。纯人工 done / released 继承已确认的 observe 授权规则；真正 publication reconciliation 继续使用 reconcile。

## 2. 原生会话与正选来源归属

私有归属证据同时关联当前 task、project、worktree、Agent、持久 terminal、binding、真实控制器或采集连接、provider 类型、原生会话身份、CWD 以及本次启动 / 恢复 operation。持久 terminal 身份与 PID 分离；PID存在、近期活动或持久化 NativeID 均不能单独证明当前控制器所有权。

真实控制器只使用自身提供的正选来源和授权根，继续核文件身份、实际 CWD、原生身份与 task/binding 关系。未授权根、身份冲突、旧连接、路径别名、多个同文件名称及来源歧义应在实际失败位置拒绝，返回安全固定原因并保留 unknown / partial；不能扫描 home、改用其他候选文件或绕过单链接保护。

当前 nlink2 的旧原档继续保留来源 FAIL。两个系统名称指向同一文件的事实不能证明另一会话、另一控制器或链接创建者；未经授权不得删除未知链接。本次新 scope 同样必须通过真实来源检查，若仍失败如实报告并停止依赖该来源的操作。

来源活性三值、会话归属和档案完整性分别记录。process 退出不能推导 ArchiveAppendOnly、PrefixComplete 或 SessionComplete。档案完整行覆盖、unknown 行、错误数、当前 snapshot 完整性与整个 session 完整性分别有真实分母；恢复原始缺口保持原结果。

每个发布的 event_id 逐一通过真实 EventRepository / SQLite 定位；游标仅在真实事件成功提交后推进。重复档案或通知重放不得刷新 live。通知与档案独立核来源和覆盖，不能互相替代验收。

## 3. 输入就绪、提交与消费

注册数据的 installed、configured、launchable、native capability 各自核实。DriveMode 和 InputMode 表达可信执行方式与 framing，不直接授予权限或证明当前 ready。

structured 接入必须取得本次 owned 控制器的初始化和正确会话建立确认，再按对应真实协议提交输入。GET只读已观察的状态，不因页面刷新启动、probe 或恢复控制器。每个客户端的真实协议边界保留来源证据，不能把另一个 provider 的信号迁移成相同的原生身份。

2026-10-15 总负责人正式答复 `msg_a1839d29aa46` 授权原 R 补齐 Relay 同一控制器的两阶段端口，优先于 L Interpreter 工程适配。本轮沿既有 Relay service、store、operation、Registry 和真实控制器完成初始化与首次输入的衔接。H 维护唯一生产组装与公共契约，F 继续唯一实际界面操作；接口名称与准确签名由 R/H 随精确 SOURCE 交接。

初始化阶段使用空 Message，不调用模型，记录真实初始化、原 native 身份、原控制器及本次准备 operation 关系。初始化不产生输入已发送、模型已开始、目标完成或人工 done 事实。正选 metadata 尚未可用时如实表示 NotYetAvailable，不另起探针、不更换控制器或档案补齐。初始化失败、来源拒绝、进程退出与恢复后的所有权不足分别保留实际结果；未核实的初始化不能进入首次输入。

首次输入必须复用上述原控制器和原 native 会话，并通过当前权限 revision、CAS、预算、准确 input operation、来源 guard、实际工具关闭及固定上下文检查。空初始化后的 generic Send 同样必须核查首输入来源，并将实际批准的固定上下文交给本次输入；仅配置回调不能作为运行证据。错误在模型请求前明确拒绝时保留该拒绝；已受理后未知继续占用原 operation，不重复发送、不创建替代会话。重复或并发首次输入不能产生第二次提交。

总负责人正式答复 `msg_e8f75be9692e` 确认实际单一 Registry 时序：F 在同 URL 确认空 agents 与 observer ready → H 唯一真实登记并保存当刻 t0 → F CreateIntent / PrepareIntent 以空 Message 初始化同一控制器 → H 核实际正选来源与工具关闭 → F 原 Confirm 提交首次输入。初始化、来源检查和准备耗时全部计入同一个三分钟，t0 不重记；不隐藏登记、不预热其他控制器、不创建第二 Registry 或 generic 替代会话。source guard 必须先于任何模型输入。当前固定目标为公开 Go 项目 baseline `da3bdddb085808863090a05f60cc1c64415039cf` 的四个批准文件，保留 Mean 并实现 Median。初始化与首输入之间的身份、授权和来源变化均须重新核查；后继第二 native / handoff 独立记录，不混入首个三分钟窗口。

总负责人正式答复 `msg_c818d5461146` 授权原 R 在原 RelayService / Store / Receipts 与原写域内实现内部可信 `RegisterTrustedAgents`，原 H 复用 `relayConfiguration` 的成熟 JSON 解析和固定可信 `agents.json`，提供同一服务实例的实际重载路径。保持唯一 Service / Registry / store / controller；不重启或新建实例替代空登记阶段，不新增 schema、调度器或接受 HTTP 命令 / 路径 / argv 的入口。完整登记快照须先校验和复制，再原子发布并保护并发读取；错误立即返回，不部分激活或沿旧配置声称新登记成功。重登记不得改变在途 operation / intent / controller 的身份、执行策略或权限，冲突沿现有版本和占用规则拒绝。

唯一 t0 为 F 在同 URL 确认空 agents 且 observer ready 后，H 首次实际写入本次目标 Agent 完整可信登记的开始时刻。另保留实际写入完成、解析 / reload 与首次激活时刻，全部耗时计入原 180 秒，失败和重载不重记 t0。空文件创建、文件 mtime、预先写好的登记或激活后新取时间均不能作为起点。R 交准确 SOURCE / 独立 REPORT，H 消费后验证真实 SQLite / HTTP、同实例更新、失败与并发，再交 F 准确契约。Calls / Known 两项用户待答约束保持；当前只开发与独立验证，不提前开始登记窗口或发送模型输入。

非交互 stdin 接入必须符合可信启动策略、实际 argv、原工作区和当前 owned 输入通道。pipe 可写、写入完成、EOF 提交完成、客户端接受与客户端消费是不同事实。stdin_eof 只有一次首输入；关闭后的 send 明确拒绝，不改变 framing 重新投递。实际输入消费没有协议证据时保持 unknown。

交互 TUI 的实际就绪仍须由相应客户端提供已验证信号。当前没有通用 TUI / paste-ready 观察端口；缺少证据时不得声称 ready 或利用视觉判断补齐。不新增通用终端组件或自研调度器。

新真实链路必须使用已安装、已配置的实际控制器与 Agent 创建的公开 Git 项目。原生文件与命令工具保持关闭；核本次真实 owned PID、argv / 启动策略及 controller 自产权限证据。tool_calls=0 只证明本次未调用，不能代替关闭工具的配置事实。

## 4. new / resume / fork / handoff 关系

首次创建记录真实 new 会话及启动 operation。resume 保持原 task/binding/session/native 身份根和原工作区，使用空首输入，并由实际 provider 恢复结果确认；持久 binding 本身不能重建当前 owned process。

fork 需要实际 provider 支持、明确父关系和适用授权。当前不支持的 fork 明确拒绝，不用新的普通会话填入成功 fork 关系。跨 Agent handoff 在已安全释放的中性工作区创建新 session，记录原会话与人工上下文交接关系，显示实际模式和身份变化。

handoff 沿原 CreateIntent 增加可选 `HandoffFromTaskID` 和 `HandoffContextConfirmed`，表示操作者明确声明原任务和已经核查的人工上下文交接。两者成对校验：缺原任务却确认、给原任务却未确认、原任务不存在 / 未 released / pending非零 / 未核退出 / 跨project或worktree / Agent未变化均立即拒绝。原intent继续核human、当前权限与settings revision，前任务仅限当前操作者允许观察的同域历史。省略两者保持new；不能仅凭同工作区上一released任务推导confirmed handoff。

人工声明与新执行者真实归属分别确认。handoff关系绑定准确原task、本次实际dispatch operation及新owned provider session结果；缺少session确认时执行证据仍unknown。F通过原界面明确取得人工交接确认，不自动勾选；H负责准确DTO / OpenAPI，R交精确领域SOURCE后消费。上述衔接属于已经授权的身份关系实现，working 门禁继续待答，人工 done 采用第 1 节已确认双通道。

每次新的控制器或采集连接关联自身启动 / 恢复 operation。新的序号空间不能冒充旧游标的继续位置；旧连接事件不得更新新会话 live。服务重启、原生状态丢失和身份冲突分别保留真实未确认状态，不自动创建第二执行者。

旧记录缺少的关系、归属、初始化或消费证据保持 unknown。读取旧数据不能补造成功时间、控制器身份或原生开始事件。必要存储兼容只作用于 Relay 自身 store，冻结的认知 schema / engine / mapper / weights 不变。

## 5. HTTP、前端与证据边界

沿现有 Relay HTTP、OpenAPI、生成类型和前端动作入口增加明确已决定的安全派生内容。session-events 继续精确 `event_id / kind / topic / tokens / scope / ts` 六字段；不追加原生文件名、CWD、原生 payload 或凭据。内部私有证据与公开 DTO 分别映射，公共响应只含允许的 task/session 引用、安全状态、时间和固定原因。

HTTP / UI 不执行额外模型调用来获取 ready 或修补未知结果。F 继续作为唯一真实产品界面操作者，通过现有 HTTP 执行 confirm / send / read / stop / done / release / 409；H 只核原档、SQLite、HTTP契约和来源覆盖，不另起 CLI 或复现 F 的任务。

H 的非原生启动依赖 SOURCE `1aadeebaba2330bb8b786989f577415da5208fd5` 已提供：Go 真实绑定127.0.0.1:0一次并持有同 listener，stdout仅公告 `type=http_listening` 与实际 `api_url`，客户端继续通过 health 确认可响应。B 负责原 helper 消费，Origin / CSRF 由真实 API 与独立 web 地址验证；该交付与 v1.4 语义验收分别记录。

原始源、原始输出、credentials、原生历史和私人路径留在本机受控位置，不进入 HTTP、Git、release 包或 stdout。安全报告只给聚合结果和本机证据指针，包不携带真实 DB、会话或鉴权资料。

## 6. 固定分工与阶段交付

| 人员 | 固定工作树 / 分支 | 本轮责任 |
|---|---|---|
| 模块负责人 | module-relay-20261014 / songconmaisaix31-design/module-relay-20261014 | 只写本文与 docs/modules/relay.md，管理开发、语义决定和实际验收 |
| 原 R | works-relay-runtime / songconmaisaix31-design/works-relay-runtime | 原 relay 领域、应用、store、native / terminal adapters 的证据和关系；当前 task_f082d8b36471 / ctx_f17d901de131 |
| 原 H | works-relay-http / songconmaisaix31-design/works-relay-http | 当前 task_3004b636a3ab / ctx_7eb5801f4450 处理来源、HTTP、OpenAPI 与实际采集；历史阶段与恢复记录见模块管理文档 |
| 原 F / B | 由产品负责人持续管理原树 / 分支 | F 消费准确契约并独占 UI 操作；B 独占依赖、config.go、普通后端与 tests/s1 helper |

原 H 的 port0 阶段先完成真实验证和 SOURCE / 独立 REPORT；既有活跃人员不中断，不安排第二 H。同 owner 连续任务或合法结算交接后的新 Task 仍保持原树、分支、终端及写域。源码接口未精确交付前，消费者不采用 WIP；领域问题交原 owner，集成人员只补必要薄组装。

每阶段普通 commit / push SOURCE，随后独立 REPORT，核远端 exact SHA。各第一 RED / UNKNOWN / NOT_RUN 保留。禁止 mock、伪造 API / 原生协议 / 模型成功、新增 skip、降低原断言、程序化编辑代码、force / rebase / reset / squash；禁止第四层 Agent。

## 7. 验收顺序与当前状态

| 阶段 | 实际检查 | 当前状态 |
|---|---|---|
| 语义领域实现 | 唯一执行权与重复操作；独立输入 / 开始 / 双通道完成事实；初始化 / 消费分别记录；真实归属和关系拒绝；旧记录 unknown；权限版本与撤销 | 独立事实 / 人工handoff拒绝阶段实际检查PASS；双完成通道已明确，R/H实施与实际验收继续；working门禁及Calls/权限覆盖仍待答 |
| HTTP 与 UI 组合 | 正确生成契约、六字段、安全错误；真实 SQLite / HTTP 与权限；F 展示和操作；port0地址与Origin / CSRF / 关闭 | port0真实SQLite / HTTP / CSRF / 关闭独立通过，B/F实际helper消费继续；v1.4组合 NOT_RUN |
| Relay 两阶段初始化与首输入 | 同控制器空 Message 零模型初始化；准确身份 / operation；来源未就绪及拒绝；首输入重新核权限 revision / CAS / 预算 / tools / source / 固定上下文；并发与未知禁止重发 | SOURCE `cc326be57a9a5f3bdd45687b0ac23b972adb8cef` / REPORT `41b4d71452f711f972d3c9836e0484fa94fdb854` 已交；模块精确快照五项实际边界检查 PASS，原 B SQLite 冷 Registry 断言 RED 保留；真实初始化和首输入 NOT_RUN |
| 单次真实 CLI | 新公开 scope、准确 owned 控制器、自有 DB、工具关闭、一次明确输入；初始化、接受、开始、结束与人工目标分开；两阶段源码通过后执行 | NOT_RUN |
| 三分钟 Today | F observer 先 ready，无登记基线；H一次真实登记并保存原始起点；每 native event_id 逐 SQLite 核；实际六字段 HTTP 与两尺寸 firstDOM 时间 | 新窗口 NOT_RUN，旧591116 / 591117毫秒 FAIL保留 |
| 生命周期全部核心路径 | 实际 EOF / structured、互斥 / 重复 / 通用冲突、working、unknown send、合法目标完成与释放、alternate / previous、publication error；真实拒绝不制造假状态 | 未满足项 NOT_RUN / FAIL，旧默认203PASS / 41FAIL / 32SKIP / 4NOT_RUN保留 |
| 最终 Windows x64 包 | 原集成人员消费准确 SOURCE / REPORT，在实际 packaged server 自有端口 / DB / scope 核全部适用功能及原生路径 | v1.4最终包 NOT_RUN |

总负责人 `msg_d807dc7e94d0` 曾授权实际 Codex turn-start 正向链路，后续 `msg_9e32e1f32ba5` 明确 Calls / Known 用户待答期间不开始 t0 或输入，当前继续保持该门禁。允许实际执行且准确源码组合通过后，按第 3 节时序由原 F 先取得空登记 observer ready，H 唯一登记并保存 t0，F 空初始化同一控制器；实际工具关闭与合法来源核实后，经产品 HTTP 单次输入，H 核原档与 SQLite。人工 done 如实保留实际完成通道；verified 须有目标核验事实。未知交付停止重发；旧失败、来源拒绝和每项真实不足独立保留。

verified 固定目标必须能从本次合法来源独立核查，来源拒绝时保留未通过。user_confirmed 保存操作者明确确认及其出处，不冒记独立目标核验。不能用预置短回复、旧成功输出、静态页面、targeted Go通过或 replay 代替实链路、完整默认浏览器或实际发布包。

源码、报告、实际参数和独立结果交总负责人及产品留存，由合法新集成发布人员普通合并。最终 GitHub Release 仅由总负责人接受具体结果后发布。旧正式 V4、锁定协议、自动任务日程及17321服务继续不变。
