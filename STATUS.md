# 当前状态

当前任务是[产品拆分与完整任务交付](tasks/swa-evolution-20261023.md)。用户已确认 Astrocyte、Attention Mirror 拆为两个独立项目/仓库，SWA 只保留必要集成；本轮实际并发目标为 N=4、N=6。开发分支 `integration/swa-task-cards`，已建立[草稿 PR #6](https://github.com/songconmaisaix31-design/swarm-anything/pull/6)。当前 Linux 安装来源为 `7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f`，完整构建、安装及装内 D12 已通过。实际 N4 完成 Controller 和根任务创建后，在 Worker 启动请求失败并保留 pending/UNKNOWN。后继开发单 Worker 已修复模式、命令来源、严格准入报文和原生选项读回后的批准时序；最新 pi05 在 `45ce26f0` 实际启动为 ready/structured，但读取状态时拒绝了真实 orca 输入回执。该投影修复已通过同原件前后对照和局部检查，完整输入绑定、模型输出和资源释放仍待新请求。Windows 权限修复已通过原生 SDK 六项及真实 Controller.open/close；D3服务退出0但17项未知观测使完整关闭未证。后继D4短控制中止、记录1项创建时间不可用，原因继续核查，不回填D3历史。N6 未运行。

## 当前实现与验证

| 范围 | 已取得的结果 | 剩余验收 |
| --- | --- | --- |
| Workspace 与 SWA 服务 | 唯一空间存储责任、原表/ID/资料引用/同连接事务保留；独立 swa-server、swahttp/swastore/projectfiles 闭包通过；真实 SQLite/HTTP、原生产服务生成的公开旧库兼容、开发进程生命周期及 Windows 交编通过；S7ff Linux 完整安装与 D12 通过 | 当前 Worker 启动失败修复后的完整验收 |
| Astrocyte | 独立本地仓库 `/workspace/astrocyte`，`426723b5525ec36b1e0c93dd4b6d9eb6e42287f1`；认知 MCP/sleep、快照/只读 HTTP、GUI 和显式单文件事件导入均已独立审查。Go 测试、Linux 构建、真实 MCP→快照→七 API→浏览器 DOM/原库不变/正常退出通过；真实 Git→实际导入二进制→SQLite 幂等与冲突回滚通过；Windows x64 交编通过 | 自动来源和外部完成证据仍为显式集成边界；Windows 原生运行未验 |
| Attention Mirror | 独立仓库 `/workspace/attention-mirror`，`58bb6e4454e03db135cabe6de2dec8dfc481aee6`；独立资料服务/GUI、显式提取器配置、资产根限制及原适配器测试已独立审查。真实公开原件导入/重放/人工沉淀/Taste/撤回保全、Go 检查、前端 67 项规则、DOM 9 项规则、实际安装启动通过；Windows 交编通过 | 公开 ACL URL 实测因 DNS 解析失败；模型、PDF/媒体、OpenViking 当前空间权限集成与 Windows 原生仍未验 |
| run/wait/export | CLI 与 GUI 接同一 Run；稳定请求 ID、递归 strict、分离退出码、完整导出已实现；CLI 局部测试、Web 类型检查/ESLint、观察规则 9 项、真实无原生运行时 DOM 2 项通过 | 完整真实 Run、断线恢复、实际递归失败和模型结果 |
| Go 组合提交 | 原 controller/task/worker 三阶段、输入完整内容绑定、已完成任务关联恢复；真实文件同大小同时间变更拒绝测试通过 | 本轮安装中的原生幂等恢复 |
| 能力与角色 | 实际观察层级、送达/恢复状态、角色/讨论界面已接入；能力、角色和实际 SQLite CRUD/CAS 规则通过，源码接受 | 实际模型及账号归属、父任务恢复 |
| 原生 Run 生命周期 | 原 Task/Dispatch、journal、成果和原资源释放路径已实现并获独立源码接受；完成邮件仅经明确 worker.send 恢复，notify 只存记录，成果保持 unreviewed。严格 TS 与实际 SQLite 规则检查通过 | 实际子成果唤醒父任务、完整资源回收、最终递归聚合 |
| N4/N6 | 同 Run/角色复用/实际孙任务/父恢复/停止隔离/自动释放驱动已审查；S7ff 安装已完成真实 Controller/根任务创建，Worker 启动被上游工作树参数规则拒绝，原请求保留 UNKNOWN | 修复并局部验证后继续 N4/N6；并发、停止隔离、继续输出与完整业务结算未通过 |

旧混合 Go/SQL 与配套资源已按独立审查清单移出，公开 testdata 和研究数据保留；完整保存点与归档见 [LEGACY](docs/LEGACY.md)。清理后的 `pnpm check`、`pnpm check:race`、`pnpm build` 与实际默认启动检查均退出0，同产物页面/JS/CSS、独立 SQLite、健康/会话/项目及正常退出通过。Go 依赖清理后双平台含测试图、17包架构、空 tidy 差分与模块校验通过。

S50 Linux 组合包构建、归档、解包、严格安装及装内 D12（12通过、0失败、0跳过）均通过，安装根为 `/workspace/swa-linux-installs/swa-linux-combined-50d099dd-n46-01`。首次 N4 在空间 POST 201 后被工具误套 UUID 条件拒绝；原数据库只有该空间，项目和 Grant 均为零，原服务正常退出。空间 ID 在拆分前就为不透明字符串；工具修复的真实 SQLite/HTTP 创建、查询和项目登记局部测试通过，原请求没有重放。

第二次 N4 使用新的独立服务。原 Run 请求 `52573151-edbc-409c-9210-21efdad053f0` 在 `pty-spawn` 阶段停为 pending：SWA 将 `/bin/bash` 传给仅支持 Windows 的 `shellOverride`，上游明确拒绝创建终端。项目 Grant 撤销已确认，无 Worker 成果。关闭时服务因未决原生请求退出6；观察器退出0，连续两次观察所属进程为空。资源退出与业务 UNKNOWN 分别保留，不能将该次判为成功。准确原件在 `/workspace/swa-pi-n4-packet-50d099dd-20261023-02/evidence` 和对应 `swa-pi-owned-50d099dd-n4-20261023-02/evidence`。

同 S50 Windows run `39900064207` 在 build 阶段退出1。两次只读诊断 `39901968797`、`39902560190` 均完成，固定原产物证明冻结安装、Go 下载与校验均退出0；原冻结安装日志仍筛选旧工作区名称并报告 `No projects matched the filters "astrocyte, @astrocyte/web"`，Web 命令随后因 `'tsc' is not recognized as an internal or external command` 失败。修复 `67c3d63de12ce87f65ea14b429a8845116f5a288` 仅改为根目录与 Web 路径选择器；固定 pnpm 实际选择从2恢复为4个工作区，固定 Node 下实际 Web 工具版本和 PowerShell Parser 检查均通过，已独立审查，未重试 Windows 完整构建。

Linux 控制器改为省略仅适用 Windows 的 shell override。严格 TypeScript 和真实本地 PTY→原 helper 两次同 PID 识别→正常 EOF/进程消失的局部检查通过，细节见[修复记录](tasks/swa-controller-linux-startup-20261023.md)。该局部检查不代替完整 RPC、Run 或模型验收。每次失败均先交代最小复现、根因证据和局部验证，再启动后继完整构建。

S4c Linux 组合包构建、归档、解包和严格安装均退出0，安装根为 `/workspace/swa-linux-installs/swa-linux-combined-4c007f75-n46-02`。装内 D12 首次因共享树已有 Windows 修复而在干净源码前置检查被拒绝，未启动原生测试；随后在独立同提交检出 `/workspace/swa-s4c-verification` 执行，12通过、0失败、0跳过。实际结果为 `.verification/swa-linux/evidence/4c007f756d6a67f0c9c7fcdcd87f7a81855276a5-combined-n46-02/`。

S4c Windows run `39903915697` / job `120251402016` 已完成前六项 setup，但在 Web 类型检查遇到 `SwaRoles.tsx` 与 `swaRoles.ts` 的大小写文件名冲突，build 退出1，未安装。修复 `06495931` 将组件原字节改名为 `SwaRolesPanel.tsx`，只更新唯一 import；全 Web 类型检查、实际 manifest 构建及引用核查通过，156个 TypeScript 源文件无大小写 stem 冲突，已独立审查。该局部验证不代表 Windows 完整构建通过。

S4c N4 使用新的独立服务，原 Run 请求 `fd652447-4a93-48aa-8628-78d0f1ee2602` 的控制器创建已完成；子请求 `654666e4-58a1-599c-825a-8746ac4ba8cf` 的 `swa.taskCreate` 停在 `helper-input` pending，实际 PTY 在 05:17:46.955Z 退出1，Worker 数量为零。任务消息为322,411字节。隔离真实 PTY 使用原公开 invocation 和新 identify 请求，短报文成功，322,408字节长报文触发 helper 自然退出1；其 JSON 解析错误与同一报文前4095字节的解析错误完全相同，证实该传输路径存在长行截断。此复现只执行身份请求，未代替真正任务创建往返。原 runtime 未注册 `terminal.read`，原 daemon 对已退出终端返回 snapshot null，旧现场错误正文未能取回。授权撤销已确认；Root 正常 EOF 后服务因未决业务退出6，观察器退出0，连续两次所属进程观察为空且最终 complete 为 true。业务 UNKNOWN 与资源退出分别保留，原请求未重放。原件在 `/workspace/swa-pi-n4-packet-4c007f75-20261023-01/evidence`、对应 owned 目录及 `.verification/controller-taskcreate-diagnosis-20261023/`。

两份独立仓库另有可恢复完整 main 历史的 `/workspace/astrocyte-426723b5.bundle` 和 `/workspace/attention-mirror-58bb6e4.bundle`，均通过 `git bundle verify`。它们不包含 ignored 数据或账号；远端仓库尚未创建。

helper 初步改为标准 readline 的 `terminal: process.stdin.isTTY`。真实 TTY 的短消息、322,408字节长消息和后续短消息均由同一进程响应，Ctrl-D 自然退出0、进程消失；同一公开 invocation 的 pipe 对照也响应成功并在 stdin EOF 后退出0，未强制终止。准确原件为 `.verification/controller-taskcreate-diagnosis-20261023/long-WNtUBF/` 与 `pipe-uY69to/`。该结果覆盖 LF 输入对照，尚未覆盖生产 writer 的分段输入与 CR 提交。

新开发副本 `/workspace/swa-tcrpc-r1-N9ptBE` 使用原 S4c Go/Node/runtime 与局部重编 helper，不能作为同来源完整安装。原 controller.open 请求 `93b72058-4ee3-46e5-9538-88f82b22bb39` 返回 `operation_unknown`，SQLite 保持 `pending/pty-spawn`；未调用 task.create，Worker/model 为零。实际 PTY 日志证明 helper 已启动并返回 `ready=true`，但回显的输入 JSON 与回复 JSON 相邻且没有换行，原正式 NDJSON 读取器无法识别回复；正式 writer 分别写入文本和 CR。Grant 已撤销，原服务退出6、观察器退出1，原报告的完整进程退出 UNKNOWN 保留。补充收尾通过原认证 daemon 接口向准确会话发送 Ctrl-D，helper 自然退出；随后多余的会话删除得到 SessionNotFound，该工具退出1保留。独立只读 `cleanup-final-01.json` 连续两次证明原 daemon 的准确 pidfd 已退出、helper 的 /proc 记录缺席；这两个已观察资源均已结束，daemon 仅剩 PID1 尚未回收的 zombie，未发送 TERM/KILL。该补充证明不覆盖未观察过的后代。原 `result.json` 的 `controller_closed:NOT_CREATED` 只反映 CLI 未返回 selector，不能证明未创建；以 pending 请求和实际 PTY 为准，业务 FAIL 不变。未重放该请求、未启动下一次完整构建。

正式 writer 的局部最小复现 `writer-27bsPU` 重现同一行回显与回复，原读取器记录零条回复并超时。helper 补上标准 readline 的 `output:process.stdout` 后，`writer-N4J60x` 经原分段 writer/CR 提交完成短→322,408字节长→短的三次响应、同一进程及 Ctrl-D 自然退出0；实际 ANSI 输出323,608字节，未超限。`pipe-RTTy6Q` 的真实 pipe 响应、stdin EOF 和进程退出也通过。准确源码与实际验证查[完整任务输入修复](tasks/swa-controller-task-input-20261023.md)。

后继 r2 在复用开发副本的预检阶段因已退出 esbuild zombie 的 `/proc/100013/fd` 不可读而停止，未替 helper、启动服务或发送业务请求。独立确认状态 Z 后，预检仅对明确退出的记录保留未枚举计数。r3 使用同一开发副本及全新数据、项目、请求完成真实 RPC 验证，原件为 `/workspace/swa-tcrpc-r3-ZM9XQg/evidence/result.json`：请求 `c8eff090-40a1-4179-8fdc-06d4b985d870` 及两个子操作 completed，SQLite 唯一根任务和完整 spec/context 一致，原只读查询前后回执总数不变，后续 helper observe 正常。真实回执按原序列化规则得到6105字节报文，超过4095边界；该数字不是独立 wire 抓取。controller close、Grant revoke、serve/observer 退出0和两次完整所属进程为空通过，原安装关键 SHA 不变。该结果证明开发集成，Worker/model 为零。永久回归入口随后消费这次 completed invocation，TTY/pipe 各完成三次响应、6102字节长输入、自然 EOF 退出0、实际进程缺席及来源首末不变检查，未强制清理；原件为 `.verification/controller-taskcreate-diagnosis-20261023/formal-r3-01/`。这些结果未替代新安装、N4/N6 或 Windows 动态验证。

S7ff Linux 组合构建、归档、解包及独立严格安装均退出0，安装根为 `/workspace/swa-linux-installs/swa-linux-combined-7ff3a7d7-n46-03`。同一干净提交执行装内 `linux-provider-tree.test.mjs`，12通过、0失败、0跳过，原件在 `.verification/swa-linux/evidence/7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f-combined-n46-03/`。

该安装的 N4 原请求 `18752a1a-e72f-4f43-97c3-bfdff1d07cf7` 已完成 Controller 创建；任务请求 `6fb4e04d-e4a7-53c1-a7d4-0639b1db5b7e` 及两个子操作均 completed。Worker 请求 `c853b7e2-b789-5652-a375-2d13783c083d` 停在 pending/helper-input，实际返回 `Creation and setup options apply only to new-child or new-top-level worktrees`。只读 SQLite 有唯一 ready 根任务 `task_bdedb368b00d` / Run `run_ba4946db2c34`，Dispatch、Worker 与资源表均为零，驱动无 Worker 输出；这些记录不能证明原生端完全没有副作用。源码将 Worker 名称同时作为 `worktree:'current'` 的原生 `name` 参数，触发上游创建选项校验。修复 `3f110ee430387c66ff9e3b253d893e066bed0e86` 已推送，发送端和精确准入端各移除该参数，保留任务标题、输入与实际模型；strict TypeScript 退出0且源码独立审查接受，后继开发单 Worker 验证见下。Grant 撤销已确认；Root 正常 EOF 后服务退出6、观察器退出1，原完整所属进程关闭 UNKNOWN。旧请求不重放。原件在 `/workspace/swa-pi-n4-packet-7ff3a7d7-20261023-01/evidence`、对应 owned 目录与 `.verification/controller-taskcreate-diagnosis-20261023/n4-7ff3a7d7-failure-01/`。

补充只读核查确认原 daemon351430 与 helper351866 仍保持本次内核身份。正常收尾工具准备时先因根目录不能解析 esbuild 退出1，随后因缺少正式源码解析条件与 process-host 插件退出1；两次均未连接 daemon。恢复正式构建的依赖入口与解析配置后，第三次局部编译退出0、56个生产输入逐字等于 S7ff，语法检查及独立审查通过。实际收尾经原认证 daemon 接口向唯一原会话发送一次 Ctrl-D，helper 自然退出、session 清空，shutdownIfIdle 返回 ACCEPTED；两个原 pidfd 连续两次报告已退出，间隔50.128084ms，观察器退出0，helper 的 /proc 记录缺席。daemon 仅剩 PID1 未回收的 zombie，未发送 OS 信号。执行退出0，原件为 `.verification/controller-taskcreate-diagnosis-20261023/n4-7ff3a7d7-cleanup-02/result.json`。该补充只证明两个准确资源已结束，原业务 FAIL、Worker pending 和原完整树 UNKNOWN 均保留。

沿同一启动路径另发现 Pi 的显式模型选择缺少启动能力目录，原 `checkAgentLaunchModelPreferences` 会拒绝当前固定模型。[Pi 模型启动任务](tasks/swa-pi-worker-model-launch-20261023.md)已完成官方 1.0.1 参数与原消费者的无网络复现及修复，5项规则和11项原参数处理测试通过，runtime strict 退出0，源码与结果独立审查接受。两次原测试配置装配失败均未进入测试，原件保留，默认检查已接入 Pi 规则。后继开发验证见下；新的完整安装与 N4/N6 尚未执行。

随后以 `0bf2c3e534adb02e8a5f5789e586cbd9e6b18274` 重编独立开发副本的 host/helper，运行一次真实单 Worker 检查，原件为 `/workspace/swa-worker-pi-01-Vg8TlX/evidence/`。Controller 和根任务创建通过；Worker 原请求 `cb57657e-89a6-4e67-88f9-c5eff888951e` 及 `/native` 都 completed，但 CLI 退出4，业务回执为 failed/dispatch_input/agent_not_in_foreground。实际新建 Task、Dispatch 和终端；准入选择为 structured/ToolsOff，实际 mode 为 terminal/user_default。源码确认私有 runtime 未覆盖上游 `experimentalNativeChat=false`，因此没有进入结构化启动准入。该次仅记录一次启动命令尝试，没有模型 delta 证据；原服务退出6、观察器退出1，业务 FAIL 和完整树 UNKNOWN 保留。Grant 撤销确认，I4c/I7ff 原安装关键文件不变。

本次补充收尾经原认证 daemon 接口向准确 Agent shell、Controller helper 分别发送一次 EOF，两个 session 清空后 shutdownIfIdle 返回 ACCEPTED。两组原 pidfd 观察分别确认 daemon/helper 和 daemon/shell 已退出，两次观察间隔均超过50ms；helper 和 shell 已缺席，daemon 仅剩 PID1 未回收的 zombie，执行退出0、未发 OS 信号。原件为 `.verification/controller-taskcreate-diagnosis-20261023/worker-pi-01-cleanup-01/result.json`；仅证明这三个准确资源，原失败和数据库未释放状态未改写。后继[执行模式修复](tasks/swa-worker-structured-mode-20261023.md)启用私有结构化设置，并在首条 Dispatch 创建前核当前权限、实际模式及原请求；4项直接规则、strict 和架构检查已通过，准确源码和局部结果已获独立审查接受，提交 `381c9159b2ac58ca32b2062dfbde9a0c6b6844db` 已推送。

pi-02 使用该提交重编开发 host/helper，原件为 `/workspace/swa-worker-pi-02-ZP6Rmz/evidence/`。Controller 和根任务创建通过；Worker 原请求 `20360144-5b05-4c37-b98d-7323b063d0e5` 的 CLI 退出6，实际模式检查拒绝 terminal，原请求保持 pending/helper-input，`/native` 子行缺席。只读数据库确认唯一 ready 根任务、0 Dispatch/Worker/resource 和5条回执。实际私有配置已为 `experimentalNativeChat=true`、`nativeChatProviders={}`；服务原日志记录 `pi --version` 的 `spawn pi ENOENT`，随后 installed-agent 检查拒绝，说明启动支持探测未使用发现时选中的绝对命令。没有结构化成功或模型 delta 证据；Grant 撤销和 dispatcher 关闭确认，原 serve 退出6、observer 退出1、完整树 UNKNOWN 保留。补充正常 EOF 经准确原会话结束 helper390568，session 清空后 shutdownIfIdle 返回 ACCEPTED；daemon390500/helper390568 的原 pidfd 连续两次观察退出，间隔50.181318ms，helper 已缺席，daemon 仅余 PID1 zombie。该收尾退出0、没有 OS 信号，只证明两个准确资源；原件为 `.verification/controller-taskcreate-diagnosis-20261023/worker-pi-02-cleanup-01/result.json`。命令来源修复与无模型局部验证继续进行，尚未重试完整构建。

相同 S7ff 的 Windows run `39912010769` / full job `120276917189` 已最终 FAILURE：前六项 setup 成功，driver 在 `service-smoke` 报 `The actual service/Job shutdown check failed or remains unknown`。原 evidence artifact `12483830115` 已上传；本机通过正常 `gh` 下载及读取日志时，对象存储返回403。后继只读取证 run `39916407634` 成功：原 producer 停在 `controller-open`，cli10 退出6、双流 EOF，没有超时，未进入 task.create；内部错误因原 artifact 未收集该 CLI 内容仍为 UNKNOWN。Grant 撤销及 HTTP dispatcher 关闭确认，原 serve/helper 退出6、未决 mutation 和两项未知进程观察保留。原生退出观察两次剩余为0，CIM的已知身份退出观察通过，仍不能替代完整服务关闭证明。正在准备无需完整构建的 Windows 最小复现；该只读取证的其他 native job 和 S0 均 Skipped。

[命令来源修复](tasks/swa-worker-launch-probe-20261023.md)已贯通 Worker 模式选择和 session intent 的两次支持探测。无模型复现的首轮在探测前因工具误要求所有 Agent 默认环境为空而停止；收窄为核 Pi 配置后，实际 login-shell 捕获证明 host PATH 不含 Pi，裸命令 ENOENT/不支持，同一环境使用原批准绝对命令则真实返回官方版本1.0.1/支持，退出0。修后原4项模式规则通过，runtime strict 退出0；原件为 `.verification/worker-launch-probe-20261023/`。源码与局部结果已获独立审查接受，后继 pi-03 实测结果如下。

pi-03 使用已推送的 `063adec6e802e733f47f407df80069b14fdedaf3` 和全新请求，原件为 `/workspace/swa-worker-pi-03-YYQ8lz/evidence/`。Controller/根任务创建通过，Worker 请求 `ebc8f902-e405-49f6-95fe-537c3e746db1` 的 CLI 退出6，错误为 `invalid original delivery invocation`；原请求 pending/helper-input，native 子行缺席，实际1 ready根任务、0 Dispatch/Worker/resource、5回执。离线用正式 Go Decoder 读取原 invocation 及真实 TypeScript schema 解析后的 JSON，两者都以 `json: unknown field "handoff"` 拒绝；正文284,632字节，未超2MiB。原生 producer 将交接/角色元数据加入 `invocation.spec`，Go 的严格 `OrcaTaskSpec` 不包含这些字段，错误发生在实际权限检查之前。该 handler 和 Go 类型自 I4c 未变，不能归因于混合开发副本过旧。修复将明确传递效果前准入所需的8个原字段，保留 `submission` 供现有交接恢复拒绝规则使用，完整原生任务记录与输入绑定继续保留。

该次 Grant 撤销、dispatcher 关闭确认，原服务退出6、observer退出1、完整树 UNKNOWN 保留。补充正常 EOF 结束唯一 helper407835，session 空后 shutdownIfIdle 返回 ACCEPTED；原 daemon407752/helper407835 的两个 pidfd 连续两次退出，间隔50.131274ms，helper 已缺席，daemon 仅余 PID1 zombie。收尾退出0且独立审查接受，原件为 `.verification/controller-taskcreate-diagnosis-20261023/worker-pi-03-cleanup-01/result.json`；只证明两个准确资源，原业务 FAIL/pending 和完整树 UNKNOWN 未改。

[准入投影修复](tasks/swa-delivery-admission-projection-20261023.md)仅在现有发送文件选择8个原字段，完整保留 Scope.Context 和 Submission，原 native spec、账本及输入绑定不变。同一284632字节原报文仍被正式严格 Decoder 拒绝，282957字节投影解码成功且8值与原对象相等。既有真实 SQLite 六组、50个子测试通过，包括 submission 与 handoff 冲突拒绝；TypeScript strict 退出0。准确源码与局部结果已独立接受，提交为 `5170d9bcb36ae29527dc7c943350641d48c5a9ce`；后继实际单 Worker 结果如下。

pi-04 使用该提交和全新请求，原件在 `/workspace/swa-worker-pi-04-zo0osP/evidence/`。Controller、根任务创建和只读查询成功；Worker CLI05 退出6，CLI06 查询原请求退出0。原请求 `6a91ab20-58ad-4f57-9ea2-80e2aa95f9f0` 及 `/native` 均 completed，业务状态为 outcome_unknown/dispatch_input，Worker 为 start_unknown。实际 mode 与 preferredMode 均为 structured；错误为 `The dispatch preamble was submitted but not acknowledged (unknown): The native launch identity changed during current approval.` 继续定位同步身份检查的具体失败条件，不将原请求 completed 解释为输入送达。数据库有2 Task（ready 根任务 `task_8dcff3c1f59f`、blocked 子任务 `task_67a921c7101e`）、1 Dispatch `ctx_0694ee721c2f`、1 Worker 和1 owned/not_requested资源 `wtr_45dfbb6215f8`；完整 binding-proof 和模型 delta 未确认。Grant 撤销和 dispatcher 关闭确认，serve退出6、observer退出1、controller关闭及完整树 UNKNOWN。原 custody 中 host/bootstrap/native CLI 当前缺席，不能据此把原 lease 或资源标为 released；准确 daemon/helper 的补充正常收尾另行记录。I4c/I7ff 未改，未启动完整构建。

pi04 补充收尾经原认证接口向唯一 helper427585 发送一次正常 EOF，session 清空后 shutdownIfIdle 返回 ACCEPTED；daemon427514/helper427585 的原 pidfd 连续两次退出，间隔50.167381ms，helper 缺席，daemon 仅余 PID1 zombie，执行退出0、stderr为空。原件为 `.verification/controller-taskcreate-diagnosis-20261023/worker-pi-04-cleanup-01/result.json`。仅这两个准确资源得到补证；原 Worker start_unknown、owned/not_requested 资源、live lease 和完整树 UNKNOWN 均保留。

Windows 最小复现只组合保留1d原生组件和固定7ff生产源码，不创建任务或 Worker。观察提交 `f589b2b7e49c8ab3a8ac22e663c96a8c2f38dbb0` 的 run `39924868969` / job `120318666615` 在短消息检查失败，未进入 controller-open：诊断 checker 将原 `import('node-pty')` 改写为裸 Windows 绝对路径，ESM loader 拒绝 `d:` 协议。服务正常退出0、无未知进程观察、两次原生观察为空、CIM确认退出；模型和 Worker 为零。局部修复仅对 dynamic-import 使用标准文件 URL，实际两个 bundle 比较确认3处动态导入改变、8处静态 require 不变，真实模块 URL 加载和语法检查通过；两次局部检查数量假设错误的原件保留。

后继观察提交 `75e69fcf84ba6853aeb2f24b69ed01800658c618` 的 run `39926780628` / job `120325012614` 已完成同一最小诊断，TTY/pipe 各三次短消息成功并自然退出0，验证了动态导入修复。Controller.open 实际返回 `operation_unknown/native_codex_private_security_changed`，原请求 `99a68634-7047-4435-8592-5881bbf403cd` 停在 pending/workspace-registration，尚无子任务。公开 ACL 观察显示新 data 根 owner 与当前用户不符、包含额外 Administrators ACE；新 Pi/OpenCode 项目目录满足 User/System 严格权限，Codex 目录缺席。继续追踪 data 创建过程。Grant 撤销及 dispatcher 关闭确认；服务退出6、8项未知进程观察、完整关闭 UNKNOWN，原生两次剩余为0和 CIM 已知身份退出通过不覆盖该限制。原件为 `.verification/cloud-setup/windows-controller-open-actual-75e69fcf-02/`，实际结果已独立接受；公开 annotation 尾部被 API 截断，未补造缺失字段。两轮其他六个原生 job 及各自 S0 均 Skipped。原7ff历史目录权限和内部原因仍 UNKNOWN，当前开发复现不能替代完整候选。

创建路径已定位到产品 `privateDataDirectory` 的 `os.MkdirAll(0700)`：实际 Go1.27.2 在 Windows 调用 `CreateDirectory(path,nil)`，不指定 owner/DACL，无法保证后继 Codex 严格检查要求的安全形状。未读取历史 TokenOwner，不声称确定原历史目录的 SID 或继承来源。修复 `447ccb059d8d4ae3e40afec86d276b9fc268f1de` 仅在 Go 平台函数中为新目录显式设置 TokenUser owner、受保护的唯一 User/System 权限；已有目录只读强验，原路径保护与原生严格检查保留。Linux `go test ./cmd/swa` 退出0（1.213s），Windows `go test -c ./cmd/swa` 交叉编译退出0；五源前后与准确清单一致，独立源码审查接受。原件为 `.verification/cloud-setup/windows-data-root-source-r1/`；后继实际结果如下，完整候选仍未重建。

Windows D3 `f5b004565d8e292ddf403d991de0d12bc8e5d997` 的 run `39932389781` / job `120343569017` 已执行原生 SDK 六项测试，全部通过且无跳过；TTY/pipe 各三次短响应与自然退出0通过，Controller.open/close 的原回执均 completed，原 owner 和关闭验证通过。新 data 及三个 Agent 目录均为当前用户 owner、受保护的两项权限。Controller 关闭、Grant 撤销、dispatcher 关闭和 serve退出0确认，但 fixture退出6，closure latch=true、unknown_process_observations=17；两次原生剩余0及 CIM 已知身份退出不能消除未知项。原件为 `.verification/cloud-setup/windows-controller-open-actual-f5b00456-03/`，源码与实际结果独立接受；Go 使用 D3，runtime 固定7ff，原生组件固定1d，不代表同来源候选或模型验收。原 artifact 未收集17项具体原因，后继 D4 `7af48f8de3f0ed5ade18018392ca19663190fb25` 只增加有界原因与数字身份投影，保留全部关闭判定；run `39934353177` 已结束，最终结果另见下文，旧六个原生 job 和 S0 均跳过。

[批准时序修复](tasks/swa-worker-acquired-options-20261023.md)已提交并推送为 `45ce26f0df82598bd1c6d454376efcfde9e41c3c`。pi04 原生 get_state 返回 thinkingLevel=off，原选项持久化将显式 model 扩展为 model/effort=off，但同步批准闭包仍捕获取得原生会话前的选项。用实际记录和原 SQLite 更新路径最小复现该拒绝后，修复在 owner 持久化之后、输入前，按原生读回和完整身份取得一次当前批准；后续选项和身份变化仍被拒绝。3项新规则、4项模式规则及 strict 通过，源码和局部结果独立接受；首次测试因测试对象共享嵌套引用失败的原件保留，修正测试隔离后通过。

pi05 使用该提交和全新请求，原件为 `/workspace/swa-worker-pi-05-obRHpk/evidence/`。Controller/根任务创建通过，Worker 请求 `293f1f84-fd8c-4708-9259-f6ac4ec81a5c` 的启动 CLI 实际退出0，ready/structured；随后只读 worker.show 退出7：`The accepted input has no matching native provider identity`。原请求及/native 均 completed，数据库2 Task、1 Dispatch `ctx_2dd3636acea5`、1 ready/input_accepted Worker、1 owned/not_requested资源 `wtr_e47d70336c6c`；lease仍live。只读原 SQLite/replayJournal 的最小复现确认 accepted submission 和原回执的 providerItemId 一致，值为 orca 消息 ID，原 user item 存在；当前能力投影只接受 codex/claude/legacy 分支而抛错，修复仍须保留原会话、消息和 fence 绑定。未完成完整输入绑定及模型 delta 检查，不推断模型调用次数。Grant 撤销、dispatcher关闭确认，serve退出6、observer退出1，原完整树及Controller关闭 UNKNOWN；准确 daemon/helper 的正常收尾另行记录，原请求不重放。

pi05 补充收尾经原认证接口向唯一 helper451847 发送一次正常 EOF，session 清空后 shutdownIfIdle 返回 ACCEPTED；daemon451779/helper451847 的原 pidfd 连续两次退出，间隔50.168720ms，helper 缺席，daemon 仅余 PID1 zombie，执行退出0、stderr为空，实际结果已独立接受。原件为 `.verification/controller-taskcreate-diagnosis-20261023/worker-pi-05-cleanup-01/result.json`。仅这两个准确资源得到补证；原业务 FAIL、ready/input_accepted Worker、owned/not_requested 资源、live lease 和完整树 UNKNOWN 保留。

[Pi 输入观察修复](tasks/swa-pi-input-observation-20261023.md)只增加已验证 per-session/jsonl-rpc 合同的同消息 orca 身份分支，原请求、Dispatch、session、fence及accepted settlement检查不变。用同一pi05真实输入调用保留的修前和修后生产投影，修前仍抛原错误，修后为sent/accepted/control available，result与authenticated仍unknown；输入对象未改，原写入/全文用户回显/接受事件关联已核。8项规则、strict及该原件对照退出0，准确源码和局部结果独立接受。未重放旧请求、未重新执行模型；新单 Worker实际验证仍待运行。

Windows D4 run `39934353177` / job `120350113249` 最终FAIL，原件为 `.verification/cloud-setup/windows-controller-open-actual-7af48f8d-04/`。SDK六子项继续全部通过；短控制返回 request_aborted，transport未观察、B producer未启动，不声称Controller.open回归。新投影记录唯一 creation_time_unavailable：PID1368/PPID8500、创建时间UNKNOWN；父8500的记录只是历史身份，不补同一快照的归属链。serve退出0、fixture退出6、closure latch=true；两次原生剩余0和CIM已知身份退出不能消除未知项。两条安全摘要完整，结果已独立核对；D3原17项具体原因仍UNKNOWN。该轮无模型或Worker，旧六个原生job及S0均跳过。继续先核短控制中止和创建时间读取路径，完整构建未重试。

## 保留的安装基线

- Linux 组合安装来源 `aefe6c4be30ec3ba0447fed9163eeff5a0fb04ad`，根为 `/workspace/swa-linux-installs/swa-linux-combined-aefe6c4b-pi-01`。构建、归档、严格安装、装内 D12 与组合服务正常 EOF 检查通过。旧 held 服务已经正常关闭，serve/observer 均退出 0，最终两次完整所属进程观察为空。
- Windows 标准 CLI 来源 `1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da`，CI run `39858402174` / job `120102466532`，构建、ZIP、严格安装、readiness-eof 服务及正常退出通过。该结果不包含真实模型或 N4/N6，也不代验本轮源码。
- 两平台基线来源不同。过期 N=3 执行包永久失效，不得执行或借用旧身份。
- 已清理四份可从保留归档恢复的旧解包副本，实际释放 3,967,369,216 字节；原归档、SHA、独立安装、数据库、账号与证据均保留。记录见 `.verification/cloud-setup/release-cleanup-20261023-01/operation-01/`。
- 本轮另核对并清理 `559c621b-02`、`f71eec93-03`、`aefe6c4b-pi-01`、`50d099dd-n46-01` 的四份 packages 解包副本，实际可用空间增加 3,957,583,872 字节。完整成员内容与硬链接核验、当前用户进程引用核查、删除后原归档校验通过；dist 候选、独立安装及所有数据和证据保留。系统 Docker/containerd 进程资源不可读的范围单独保留，未宣称核查所有系统进程。原件为 `.verification/cloud-setup/release-cleanup-n46-20261023-02/operation-01/`。
- 后续清理八份已有完整保留归档的旧 dist 候选，准确固定清单见 `.verification/cloud-setup/dist-cleanup-n46-20261023-03/PLAN.md`；实际操作退出0、八份全部移除，可用空间从1,191,329,792增至4,540,567,552字节。共享文件的树外硬链接未计入回收量，也未修改内容或权限；八份原 tar 的官方 SHA 校验前后均通过。当前 S4c、所有独立安装、账号、源码、数据、证据及两独立仓 bundle 均保留。实际原件为同目录 `operation-01/`，未来需要旧候选时从对应保留归档恢复。
- 随后只移除 S4c 的 packages 解包副本，完整归档、dist 候选和独立安装均保留，操作退出0。可用空间从3,513,548,800增至4,494,946,304字节；完整归档比较、当前用户进程无引用及归档 SHA 前后检查通过。原件为 `.verification/cloud-setup/package-cleanup-s4c-20261023-04/operation-01/`；下一次构建前重新检查实际空间。
- S7ff 验收失败后，按新的准确范围先移除 S7ff packages 副本和 S4c dist 副本，再独立移除 S7ff dist 副本，两次操作均退出0。完整归档比较、链接与当前用户进程引用核查、保留归档前后校验通过；全部独立安装、数据、源码、账号和证据保留。原件为 `.verification/cloud-setup/package-cleanup-s7ff-20261023-05/operation-01/` 与 `operation-02/`。两窗口净增加空间分别为1,398,685,696和413,003,776字节；后一窗口末实际可用3,272,003,584字节，期间其他写入另计。未清 Go 缓存，下一阶段重新读取实际空间。
- 随后仅清理 S50/S4c/S7ff 成功 compiler_work 下六份 `headers`/`node-extract` 官方输入副本，唯一操作退出0。六树完整比对保留官方归档、当前用户无引用、保留真实编译输出及原构建证据不变、两份官方和三份发行归档前后校验均通过。三个父目录及 pty-source/tmp/gypi、所有安装、源码和数据保留，系统进程不可读限制单列。原件为 `.verification/cloud-setup/space-budget-063adec6-20261023-06/operation-01/`；结果时点可用空间从2,885,894,144增至3,471,818,752字节，净增585,924,608字节。未清 Go 缓存；预算已计入下一 compiler_work，完整构建和各安装阶段仍须刷新实际空间。

## 当前边界与下一步

当前仅连接 Linux 环境；Windows x64 可用原生 CI，但真实模型验收尚无连接环境。真实模型继续使用已配置 StepFun 来源、固定公开四文件项目及 ToolsOff；密钥不进入仓库、日志或导出。此次明确的 N>3 请求授权其必要验证，不逐 SHA 重复请求同范围批准。

正式科研实验的原证据根与权威锁尚未恢复，本轮未运行依赖这些原件的实验；既有科研与用户数据保全。

下一步验证修复后的真实 Worker 输入观察、模型输出和资源回收，并定位 Windows 完整关闭的未知进程观测。每项失败先报告最小复现、根因证据及局部结果，再启动后继完整构建，随后继续 N4/N6、同 Run GUI 和 Windows 适用检查。正式发布、远端新仓库创建和 main 合并尚未执行。云环境配置草稿 revision 9 已保存并读回确认，安装选择器与启动说明已同步；仍需在环境设置保存并发布才生效。草稿仓库成员仍只有 SWA，两份没有远端的独立本地仓库在新任务中的恢复尚未验证。

历史阶段原件保留在[准确旧状态](https://github.com/songconmaisaix31-design/swarm-anything/blob/11619d72f08c163f159a4f4befe270806b1b67ea/STATUS.md)，其过期服务、请求和待答说明不作为当前执行条件。
