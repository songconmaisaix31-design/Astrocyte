# Works Relay v1.4 语义修改与真实验收

2026-10-15 当前状态：IMPLEMENTATION_AUTHORIZED_DECISIONS_PENDING。总负责人消息 `msg_0ba76f2d26c3` 转交用户最新明确要求：现在修改语义层，通过后执行真实 CLI、今日采集并推进 L/S。此前停止新 Relay 原生与 Today 的时间安排已被本次授权替代；执行顺序仍为语义修改和验证、真实 CLI 端到端、三分钟 Today、L1 最终验收。L/S 的其他模块和正式 V4 日程由总负责人协调，本模块只交真实安全派生结果。

本文是本次四项语义修改的唯一具体规格。业务基线继承 [Works Relay v1.2](Works-Relay-v1.2.md)、[SPEC](../docs/SPEC.md)、[发布任务](Release-v0.1-20261014.md) 和 [REA研究报告](../docs/research/orca-rea-20261014.md)；最新用户决定通过总负责人正式消息更新本文件。模块分工、精确 SHA、运行参数和验收事实记录在 [模块管理文档](../docs/modules/relay.md)。

## 已授权与待答

已经授权实施：分别记录执行权、输入接受、回合开始、人工目标完成；记录真实控制器及正选原生来源的归属；区分初始化、会话确认、输入通道提交与客户端消费；保存 new / resume / fork / handoff 的真实关系，旧缺证据保持 unknown。全部沿既有 store、operation、binding、收据与事件实现。

总负责人已向用户询问以下关键决定，答案未到时保留当前状态机，独立工作继续：

| 待答项 | 当前约束 | 答案影响 |
|---|---|---|
| working 是否严格要求明确开始证据 | 不据 stdout、Git、PID、pipe 写入或完成结果产生新的开始证明；不冻结依赖该决定的接口 | 哪些真实客户端证据允许 accepted → working，以及 UI 的具体条件 |
| 缺少独立 started 信号的客户端如何人工完成 | 保留既有 working → done → released；不自行增加 accepted → done / released | Claude、非交互 CLI 和交互终端的人工完成入口及相应验收 |

接口的公共字段由原 R/H 根据已授权证据语义和用户答案在现有契约中共同维护；确定后交原 F 消费。本文中的证据含义明确，未决定的字段名称、枚举及迁移规则不作为冻结契约。

## 1. 投递阶段与人工完成

每次动作保留既有 operation 身份和幂等规则。store 首先原子取得该动作执行权并记录 pending；只有实际持有执行权的调用进入 driver。同 operation 的重复或并发请求按既有收据读取和互斥规则处理，不能再次提交输入。

输入通道提交、客户端接受输入、owned 回合明确开始、人工确认目标完成分别记录事实、实际时间、证据来源和确认程度。未观察到的事实保持 unknown；提交到 pipe 或完成 EOF 关闭只记录该通道事实，不能产生客户端消费确认。原生 RPC 回复、原生通知和 controller 自产档案按各自实测语义归类。

最终动作收据不可重写。后到的开始、结束和观察结果沿既有事件和 operation 关系追加，不将第一次失败、`delivery_unknown` 或 provider 错误改为成功。旧输入不自动重发；未知交付继续占用工作区，原进程和 pending 未核实之前不得释放或由第二执行者接管。

当前 task 状态机保持 ready → accepted → working → done → released。accepted 表示既有人工派发动作已被接受，不能代替模型已经开始或固定目标已完成。working 的严格证据门禁及无开始信号客户端的完成路径按上表等待用户决定。

done 必须由原产品界面的人工确认执行。固定目标的实际输出先通过合法 owned 来源核查，再允许操作者确认；原生 turn 完成、进程退出、Git 活动均不产生人工目标通过。released 保留真实退出、pending 为零、对象所有权和权限校验；中性工作区仍可以 Prepared=true，但 ActiveTaskID 必须解除。

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

非交互 stdin 接入必须符合可信启动策略、实际 argv、原工作区和当前 owned 输入通道。pipe 可写、写入完成、EOF 提交完成、客户端接受与客户端消费是不同事实。stdin_eof 只有一次首输入；关闭后的 send 明确拒绝，不改变 framing 重新投递。实际输入消费没有协议证据时保持 unknown。

交互 TUI 的实际就绪仍须由相应客户端提供已验证信号。当前没有通用 TUI / paste-ready 观察端口；缺少证据时不得声称 ready 或利用视觉判断补齐。不新增通用终端组件或自研调度器。

新真实链路必须使用已安装、已配置的实际控制器与 Agent 创建的公开 Git 项目。原生文件与命令工具保持关闭；核本次真实 owned PID、argv / 启动策略及 controller 自产权限证据。tool_calls=0 只证明本次未调用，不能代替关闭工具的配置事实。

## 4. new / resume / fork / handoff 关系

首次创建记录真实 new 会话及启动 operation。resume 保持原 task/binding/session/native 身份根和原工作区，使用空首输入，并由实际 provider 恢复结果确认；持久 binding 本身不能重建当前 owned process。

fork 需要实际 provider 支持、明确父关系和适用授权。当前不支持的 fork 明确拒绝，不用新的普通会话填入成功 fork 关系。跨 Agent handoff 在已安全释放的中性工作区创建新 session，记录原会话与人工上下文交接关系，显示实际模式和身份变化。

handoff 沿原 CreateIntent 增加可选 `HandoffFromTaskID` 和 `HandoffContextConfirmed`，表示操作者明确声明原任务和已经核查的人工上下文交接。两者成对校验：缺原任务却确认、给原任务却未确认、原任务不存在 / 未 released / pending非零 / 未核退出 / 跨project或worktree / Agent未变化均立即拒绝。原intent继续核human、当前权限与settings revision，前任务仅限当前操作者允许观察的同域历史。省略两者保持new；不能仅凭同工作区上一released任务推导confirmed handoff。

人工声明与新执行者真实归属分别确认。handoff关系绑定准确原task、本次实际dispatch operation及新owned provider session结果；缺少session确认时执行证据仍unknown。F通过原界面明确取得人工交接确认，不自动勾选；H负责准确DTO / OpenAPI，R交精确领域SOURCE后消费。上述衔接属于已经授权的身份关系实现，不决定待答working或人工done迁移。

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
| 原 R | works-relay-runtime / songconmaisaix31-design/works-relay-runtime | 原 relay 领域、应用、store、native / terminal adapters 的证据和关系；当前 task_f082d8b36471 / ctx_aa83ef2b5087 |
| 原 H | works-relay-http / songconmaisaix31-design/works-relay-http | port0阶段 task_abf1f8b1e12e / ctx_d7d8c2828fba 已 succeeded 结算；当前同终端新 task_3004b636a3ab / ctx_e1a8ebeeea83 处理来源、HTTP、OpenAPI 与实际采集 |
| 原 F / B | 由产品负责人持续管理原树 / 分支 | F 消费准确契约并独占 UI 操作；B 独占依赖、config.go、普通后端与 tests/s1 helper |

原 H 的 port0 阶段先完成真实验证和 SOURCE / 独立 REPORT；既有活跃人员不中断，不安排第二 H。同 owner 连续任务或合法结算交接后的新 Task 仍保持原树、分支、终端及写域。源码接口未精确交付前，消费者不采用 WIP；领域问题交原 owner，集成人员只补必要薄组装。

每阶段普通 commit / push SOURCE，随后独立 REPORT，核远端 exact SHA。各第一 RED / UNKNOWN / NOT_RUN 保留。禁止 mock、伪造 API / 原生协议 / 模型成功、新增 skip、降低原断言、程序化编辑代码、force / rebase / reset / squash；禁止第四层 Agent。

## 7. 验收顺序与当前状态

| 阶段 | 实际检查 | 当前状态 |
|---|---|---|
| 语义领域实现 | 唯一执行权与重复操作；独立输入 / 开始 / 人工完成事实；初始化 / 消费分别记录；真实归属和关系拒绝；旧记录 unknown；权限版本与撤销 | IMPLEMENTING，待 R 精确 SOURCE / REPORT |
| HTTP 与 UI 组合 | 正确生成契约、六字段、安全错误；真实 SQLite / HTTP 与权限；F 展示和操作；port0地址与Origin / CSRF / 关闭 | port0真实SQLite / HTTP / CSRF / 关闭独立通过，B/F实际helper消费继续；v1.4组合 NOT_RUN |
| 单次真实 CLI | 新公开 scope、准确 owned 控制器、自有 DB、工具关闭、一次明确输入；初始化、接受、开始、结束与人工目标分开；源码通过后执行 | NOT_RUN |
| 三分钟 Today | F observer 先 ready，无登记基线；H一次真实登记并保存原始起点；每 native event_id 逐 SQLite 核；实际六字段 HTTP 与两尺寸 firstDOM 时间 | 新窗口 NOT_RUN，旧591116 / 591117毫秒 FAIL保留 |
| 生命周期全部核心路径 | 实际 EOF / structured、互斥 / 重复 / 通用冲突、working、unknown send、合法目标完成与释放、alternate / previous、publication error；真实拒绝不制造假状态 | 未满足项 NOT_RUN / FAIL，旧默认203PASS / 41FAIL / 32SKIP / 4NOT_RUN保留 |
| 最终 Windows x64 包 | 原集成人员消费准确 SOURCE / REPORT，在实际 packaged server 自有端口 / DB / scope 核全部适用功能及原生路径 | v1.4最终包 NOT_RUN |

语义与组合通过之前只准备公开固定任务、可核查的目标判定、启动策略和 observer，不启动新模型补证。单次真实执行出现未知交付后停止重发；每项真实不足独立保留。

固定目标必须能从本次合法来源独立核查。不能用预置短回复、旧成功输出、静态页面、targeted Go通过或 replay 代替实链路、完整默认浏览器或实际发布包。来源拒绝使目标无法核查时明确失败，不先勾选人工目标完成。

源码、报告、实际参数和独立结果交总负责人及产品留存，由合法新集成发布人员普通合并。最终 GitHub Release 仅由总负责人接受具体结果后发布。旧正式 V4、锁定协议、自动任务日程及17321服务继续不变。
