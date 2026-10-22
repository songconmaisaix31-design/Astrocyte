# CLI 与 GUI 的共享执行合同

本合同以[准确基线](00-context.md)及现有 `OrcaCommand`、`OrcaEnvelope`、原生记录为输入。后续源码以各卡准确提交为准；表中的接口位置不表示动态验收通过。

## 单一服务与授权

公共入口为 `POST /api/v1/local-projects/{id}/orca/commands`，定义于 [OpenAPI](../../../contracts/openapi.yaml)，动作和返回外形定义于 [JSON Schema](../../../contracts/swa-cli.schema.json)。HTTP 适配器调用 Workspace App `ExecuteOrca`；CLI、GUI 不直接读写原生 SQLite。执行权威仍为产品私有 Orca Run/Task/Dispatch，Workspace 仅保存项目/Grant 与原生关联，不增加 Mission/Claim 或第二任务状态。

CLI 使用明确回环 `SWA_ENDPOINT`、`SWA_PROJECT_ID` 和当前 `SWA_PROJECT_TOKEN`。浏览器复用现有人类 session/CSRF；服务将 path project 与真实调用者绑定，再核当前项目动作、目录、模型和控制权限。启动服务不创建授权。注册、安装、配置、可启动和 start/send/read/stop/resume/tools 控制能力分别使用真实 observation，`available` 不表示当前提供商已经实测。

人类向CLI Agent创建的任务追加`worker.send`时，独立核验当前人类会话/CSRF、同项目当前settings、send动作、输入目录/固定上下文、所选原生CLI/模型与外发批准。人类不冒领原Agent Grant，也不以创建者Grant仍有效作为其自身权限来源；原link中的创建者、Run/root Task和原request继续保留实际来源。该例外只适用于`worker.send`，不开放跨创建者`worker.start`。人类自己创建的controller/task/start沿原当前项目权限；Agent请求继续严格核创建者Grant/主体及原有效期。观察与人类自有安全停止分别沿原动作核权。

新增权限规则不改变原Dispatch/session、精确AgentSelection、模型、当前settings revision、固定输入/上下文版本、controller consumer generation、runtime fence或原请求幂等。投递前再次核权；人类当前权限撤销/过期、跨项目、来源变化或原生身份变化均不得产生新输入。用户已确认此跨入口追加输入范围；后端实际实现及正反向测试按02/04/06的准确源码和结果登记，不从文案或按钮显示推导已经可执行。

| 动作组 | App 权限 | 范围与副作用 |
| --- | --- | --- |
| status、capabilities、controller.show、task.show/list/result、worker.show/list/read、request.show | observe | 读当前项目关联；task.result 还核导出目录 |
| controller.open/recover、task.create、worker.start | start | 固定项目、目录、上下文与原生身份 |
| worker.send、task.complete | send | 原 Dispatch 的追加输入或根任务完成 |
| controller.close、worker.stop/release | stop | 指定原生所属；人类自有安全停止沿现规则 |

## 身份与来源

| 对象/真实字段 | 来源与绑定 | 客户端约束 |
| --- | --- | --- |
| project / project_id | Workspace 注册项目与当前批准 | body/path/caller 必须一致；不能切换项目复用 token |
| controller | 原生 PTY helper selector 与不可伪造调用者身份 | 不是展示角色或用户输入的进程号 |
| runtime_id、authority_domain | 实际运行时与本地 mutation authority | 重启/代次变化后重新核对；不凭旧 PID 接管 |
| run_id、root task / task_id | controller 的持久 task.create 回执和原生 Run/root Task | child Task 的 parent_id 必须指向同 root |
| dispatch_id | 原生 Worker Dispatch | start 回执和后续查询指向同 Task/Run，停止仅作用于该 Dispatch |
| session_id、operation_id、runtime_fence、consumer_generation | structured 原生会话及输入绑定 | 只能由实际宿主记录证明；不是浏览器生成的 Worker 身份 |
| request_id | 调用者提供的规范 UUID，绑定原动作、body 和回执 | 变更操作及 request.show 必需；Idempotency-Key 与 body 一致；效果未知查询原键 |
| settings_revision / context | 当前项目配置与固定版本上下文 | 输入前重新准入；变化或撤销依原服务拒绝 |

展示名称和职责来自已有注册/配置。持久角色、执行引擎、原生会话与一次 Dispatch 是不同对象；未有角色后端时不增加可执行身份。CLI 与 GUI 都保留服务返回的真实字段，`result` 是原生对象，按动作校验后展示，不能把所有返回改造成未经服务支持的统一终态。

## 状态、回执与恢复

| 层次 | 已有来源或词语 | 解释 |
| --- | --- | --- |
| HTTP / envelope | HTTP status、ok、error | 传输及当前动作结果；HTTP 200 不证明任务完成 |
| 变更回执 | 原 mutation_receipts 的 pending/completed 与子请求 | pending/unknown 保留占用；不能换 request 重发输入 |
| start / send | 原生 ready；journaled/queued/accepted/handed-to-terminal | 启动或投递阶段，不推导模型完成 |
| task / worker | 原生 task status、worker state、原生消息与 operation | 按实际字段展示运行、等待、失败、停止；不存在的审批状态保持未知 |
| 退出 / 释放 | 对应平台的Job或Linux所属证明、PTY、原生资源release_state与新鲜进程观察 | 服务级退出和每Worker完整退出分别核验；root退出不足以确认整树退出，无法确认保持未知 |
| 成果 | 原生 journal 与所选 Agent 自产 transcript/SQLite 输出 | complete/incomplete/unavailable/unknown 分开，坏源或来源变化失败 |

`worker.read --source transcript` 是有界预览，具有原来源 cursor 和 limit（1..1000）。`worker.list` 当前请求原生 page（limit=100），客户端必须尊重 `hasMore`。`task.result` 需要完整 inventory；发现更多页必须明确失败，不把一页当全集。完整成果保持既有总量 12MiB、来源身份、读取一致性及秘密检查，不能截断后标成功。

现有 API 没有承诺与 OpenMausBot 等价的 job 长轮询或统一 SSE。GUI 刷新读取同一 Run/Task/Dispatch 的真实快照；预览续读只能用该来源 cursor。展示时使用原生事件/entry ID 和原有顺序/代次，重复可去重；跨来源无序、缺口或过期游标应重新读快照并保留缺失提示。授权失效后停止受保护读取与变更，不自动换账号或再授权；人类安全停止继续由后端原有规则裁定。

## 平台进程身份与关闭

服务内部`SWA_SERVICE_JOB`与daemon身份沿真实平台联合传递。Windows保持`Local\SWAService<UUID>`、`serve_created_at_ms/daemon_created_at_ms`及原Job核验，原wire不增加Linux字段。Linux使用`platform:'linux'`、`SWAService<UUID>`、`serve_start_ticks/daemon_start_ticks`、`boot_id/owner_uid`和原`serve_pid/daemon_pid/daemon_session/data_root/install_root/source_revision`；ticks和UID为十进制字符串。它们来自实际内核proc身份，不换算为Windows创建时间。Node和Go还核同安装程序、VERSION、原daemon记录及实际祖先归属，旧活daemon不自动接管。

`serve`的`resource_samples`在Linux为一次实际`{platform:'linux',logical_cpus,available_bytes,sampled_at}`观察，未提供CPU百分比；Windows保留原实际阻塞CPU采样数组。该区别不改变Worker自己的资源准入，也不把瞬时内存或CPU数量称为容量承诺。Linux `e14800c1`同安装的服务/daemon正常启动、health和stdin关闭后的服务树退出已验；强杀serve的完整回收仍未知，不能推广为每Worker退出通过。

每Worker Linux完整回收D为已独立接受、正在实现且尚待真实验收的内部合同。最小进程身份为`{platform:'linux',pid,boot_id,start_ticks,owner_uid}`，live身份另核实际程序`name`；`readObservation()`按平台联合返回，Linux不补`creationTimeMs/rootCreationTimeMs`。每Worker独占Node引导器在Agent加载前以真实`PR_SET/PR_GET_CHILD_SUBREAPER`确认，内部IPC绑定随机`custody_id`、原owner/root/native CLI身份及本次`stop_id`；native API版本为`providerLinuxIsolationVersion:1`。

D需要运行内核实际支持subreaper、`pidfd_open/pidfd_send_signal`及`waitid(P_PIDFD)`。glibc/GLIBCXX/CXXABI符号基线或发行版标签不能证明这些内核能力；缺失时拒绝或保持unknown，不降级为裸PID发信号。

D要求原生直接child先由libuv完成close/回收，再仅精确回收已确认收养的直接child。需要两次独立内核`ECHILD`、递增单调时刻、原身份匹配的完整退出proof、父核对后ACK，以及原引导器实际退出，才能判定该Worker完整退出。`WNOHANG=0`、根退出、IPC中断、进程组消失或普通进程表为空均不足。native未创建与引导器从未创建分别记录；已有引导器时仍需完整证明，不能冒充processless。身份、内核查询或父IPC失效保持unknown与占用，不先杀引导器再声称释放。

这些是现有provider内部生命周期接口，不新增HTTP/CLI命令或公共Task字段。普通系统进程的内核测试、实际安装服务、Agent NativeLaunch/ToolsOff、原生模型与成果及N=3场景分别验收；各Agent的Linux来源和成果政策尚未接入时，模型输入前保持不支持。

## Worker 额度

任务文件沿用 `goal/source_repo/source_revision/success_criteria/max_workers/context`。`max_workers` 是本任务获准创建 Worker 的总额度，取安全正整数 1..9007199254740991；数值上界是跨 Go/JavaScript 的整数精度边界，不是已验证资源容量。模型、项目权限、机器能力与当前运行资源仍独立准入。

公共Schema描述现有任务文件，生成类型不将文件读取变为inline HTTP接口。`context`可省略或为null；引用与文件仍由原项目上下文读取器逐项核权。既有文本校验存在层次差异：Go拒绝全空白goal但仅检查成功判据数组非空，原TypeScript还要求每条判据字符串非空。当前合同补充不改变这些业务文本规则，不能把单个Schema通过等同于所有后端准入通过。

默认验收 N=3 时使用 `max_workers >= N` 的任务，三个独立 Dispatch/session 必须有真实模型执行区间重叠。N 不写成产品固定上限，也不把 `max_workers=2` 用例的第三席拒绝改为成功。pending/unknown start 继续占用额度，并发预留沿原持久事务核查。资源边界用例明确设置`max_workers=N`，占满后第N+1席应被拒绝；若`max_workers>N`，则第N+1席仍按真实剩余额度及其他准入规则处理。当前没有排队替代拒绝的公共合同，不能偷偷减少 N。

当前额度计数包含该root已创建的全部child Task及尚未落Task的pending预留；child完成或其资源释放不会自动归还创建名额。这是固定任务可创建/预留的Worker总席位上限，不能解释成无限顺序补位的并发池。改变回收语义须另行修订准入及测试。

`max_workers` 为上限，允许少于上限的任务；根任务完成仍要求预定 success_criteria 和所有相关 Worker/资源明确结算。被指定停止项保留 stopped/cancelled 及已有部分；其他 Worker 的完整成果和资源退出独立核查。分页或完整成果容量不足时报告明确限制。

## 命令兼容与退出

| 用户操作 | SWA 现有命令 | OpenMausBot 仅作参考 |
| --- | --- | --- |
| 发现和选择 | capabilities；worker start 的 agent/model 经稳定选择 | bots/models/engine；不导入其账号服务 |
| 提交与查询 | controller open；task create/show/list；worker start/show/list | run 的使用路径；不增加第二调度器 |
| 观察与追加 | worker read/send；request show | messages/send/events；不声称已有相同事件协议 |
| 选择性停止 | worker stop/release；controller close 另核全域空闲 | interrupt；不能默认 stop 所有任务 |
| 完整导出 | task result --output-dir；task complete | export；不采用预览截断为完整成果 |
| 等待 | 当前查询组合，新增 wait 必须先固定原生终态/分页语义 | wait 只参考，当前不新增未经支持的命令 |

SWA 退出码保留：0 当前动作成功；2 参数/请求不匹配；3 权限/批准/身份拒绝；4 不可用/不支持；5 资源拒绝；6 效果或退出未知；7 其他失败；130 当前 CLI 等待取消。OpenMausBot 的6用于 team diff，不能照用。CLI 等待的 Ctrl-C 不停止 Worker。`serve` 的 Ctrl-C/EOF 管理其所属进程，分别按 readiness、已知失败及完整退出证据保留0/130/7/6，不等同于普通查询取消。

业务 `--json` stdout 仅一个 envelope，进度/诊断写 stderr；帮助和版本是文本元数据。GUI 关闭页面只解除观察，不能关闭其附着的 CLI 服务。组合包可配置安装根内静态 Web 目录，沿同 HTTP 后台提供资产，缺资产明确失败，不自动创建第二运行时。

## 工作目录与未验范围

当前 `controller.open` 使用获准目录登记 folder，Worker 使用 `worktree: 'current'`。已见 session/账号私有根和完整所属资源机制；没有每 Worker 自动 Git worktree 隔离的现成产品证明。开发云 checkout 不构成产品写入隔离验收，工具关闭范围不扩展为代码写入能力。

产品目标并列Windows x64与Linux，Agent来源和逐引擎实际能力见[适配矩阵](02-agent-matrix.md)。45个TUI模板和7个结构化引擎是不同层次，先复用7个原生adapter并逐项评估其余来源；“大部分”的验收分母已询问、尚未决定，不能将模板探测或某个局部比例称为已完成。上游明确的平台限制单列，不用另一个终端程序冒充同一原生引擎。

本轮真实模型测试预算不设费用上限。模型账号与当前许可仍须实际存在；固定公开输入、ToolsOff、禁止Aliyun/百炼、超时、原始request/未知效果及机器资源/并发准入继续有效。Windows目标已明确为x64，当前仅Linux可访问；Windows主机、安装物及真实链路未取得证据时保持NOT_RUN。

正式动态检查由02/03/04/05/06按[负责人表](ownership.md)接续。合同与生成类型通过不代表 Windows安装、真实 Agent 调用、并发模型、GUI跨入口或最终 N=3 已通过。

执行失败后先记录最小复现、根因证据和局部验证结果，再开始下一次完整构建；原失败与未知原因保留。普通局部检查及不相关工作继续，不因等待完整构建而冻结。
