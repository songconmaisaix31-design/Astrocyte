# M1 Windows 契约检查独立审核

## 范围与原始证据

主控消息 `msg_41e06b11d9d1` 授权本轮在 M0 实施等待期间独立准备 M1 条件，并允许本报告路径。Reviewer 沿用 `task_1cd8014750f9 / ctx_980a3282c339` 和原 Run `run_62ada397a104`；M0 实现审核优先。实际会话为 Orca / Codex，预览为 `GPT-6.1-Sol xhigh fast`；没有子 Agent。

依据为 META_PLAN v2.1 M1、§5.1、§8、当前 TASKS、交付候选报告及其真实原件。产品基线 `32ab1eec808a5848aabad40f11133735a1214e55`，程序来源 S `ed543c12106c2dbb3e181cb9bc1647b626455179`；根命令顺序为 Redocly lint → 示例与能力边界验证 → TypeScript 生成一致性。

已读取原产品树 `.verification/delivery-candidate-20261020/run-contracts/{result.json,check.log}` 和 `run-redocly/{result.json,check.log}`：UTC `2026-10-19T19:30:02.8460016Z`、`19:32:25.6647030Z` 两次均退出 `-1073740791`，Redocly 已输出 valid 与 13 项 no-unused-components 警告，随后出现 `src/win/async.c:94` 的 `UV_HANDLE_CLOSING` 断言。原 pnpm 链的后两阶段未执行；其独立直调通过只支持各自范围。已读取原始 package、Redocly 配置、示例和生成脚本、check.mjs、依赖核对脚本及默认 workflow。没有运行复现、安装或业务检查。

## 读取 Worker 计划前确定的条件

| 编号 | 正确实现与必要覆盖 | 边界及失败判定 | 需求来源 |
| --- | --- | --- | --- |
| C01 | 根因方案解释真实 Windows 原入口和 Redocly 直接入口均崩溃的共同边界，区分日志证据、源码推断和待验证假设 | valid 文案和警告数量不构成正常退出；不得将全部原因归为 pnpm 或规范无效；不预选依赖版本 | M1、原始两次失败 |
| C02 | 正常路径由原 `pnpm check:contracts` 实际完成全部三阶段并 EXIT0，保留完整 stdout/stderr、准确源码、实际工具和环境 | 单独组成检查、复制旧成功或吞掉退出码不能替代原入口；执行次数和未运行阶段如实记录 | M1、§5.1、TASKS |
| C03 | 原 Redocly 推荐规则和显式配置保持，当前 13 项警告继续可观察；真实无效规范/引用/显式 error 规则产生非零退出 | 不关闭规则、屏蔽报错、删除实际契约、跳过 lint、把崩溃映射为成功或强制正常退出 | M1、原配置、fast-fail |
| C04 | 原示例及 supported capability 边界检查保持并实际验证当前 264 项；不满足 Schema 或未来操作缺失 unsupported_capability 时非零退出 | 不减少遍历范围或豁免类型；错误立即终止后继阶段；使用真实校验器和公开确定性输入，不伪造 CLI/API | §5.1、validate-examples.mjs |
| C05 | 生成检查继续由原 OpenAPI 生成类型并逐字比较；匹配时 EXIT0，实际差异时非零退出且不改写已跟踪生成文件 | 不将 --check 改为自动生成，不仅检查文件存在，不缩减生成范围 | §5.1、generate-client.mjs |
| C06 | 原链的每一失败阶段保留真实失败，之后阶段不运行；外层调用能获得该结果 | 覆盖 lint 失败、示例失败、生成漂移；不使用替身命令、mock、固定成功值、重试到通过或 fallback | TASKS、原 && 语义 |
| C07 | 方案在负责故障的项目入口或必要工具依赖边界修复，同步直接消费者 | 根 `check:contracts` 与 `scripts/check.mjs` 的实际 Redocly调用均需解释；同一故障不在不同入口残留或复制不同补丁；保持原公开命令 | §7.2、M1、当前源码 |
| C08 | 如变更依赖或工具，准确 pin、pnpm lock、依赖清单及实际安装版本一致，且来源和兼容依据可查 | 未修改的依赖无须重新安装；不改系统全局配置、项目业务/API/S3门禁或其他运行时；安装与运行需主控窗口 | SPEC、依赖脚本、总任务 |
| C09 | 按实际变更验证适用静态检查和直接消费者，不把完整 CI、浏览器和所有业务测试设为此次修复的额外前置 | 原全仓 CI 的旧替身/来源/权限失败继续明确，契约通过不宣称全仓通过；未验证的平台不写已通过 | §8、产品集成优先 |
| C10 | 正负向输入与故障注入使用真实项目工具，在隔离目录保护现有公开源码；所有改动通过文件编辑工具 | 测试用输入只检验规范和错误传播，不冒充业务或原生结果；不得修改用户数据、运行服务、正式实验或原始失败记录 | AGENTS、§8.1、总任务 |
| C11 | 主控授予唯一重操作窗口后，按 RESOURCES 真实采样、BelowNormal、Node额度串行运行；Reviewer 独立复验正常与失败退出 | 未获窗口不启动命令复现或安装；记录资源拒绝、实际启动、退出、后继未执行和原件；不自动重试 | RESOURCES、TASKS |
| C12 | 准确提交、PR与报告的结果相符，普通 push 后供主控集成；main收拢、发布、S3分别记录 | 修复契约不解答原预算/权限/资料/采用问题，不宣称 S3 或 A01 已通过；实现变更后按变化复验 | M1、§10、A01、QUESTIONS |
| C13 | 主控安排服务窗口时，先核当前running.json、实际PID/路径/创建时间/监听与运行S；只停止获准且归属已核的本项目服务 | 不复用旧c5程序限制或过期PID；不终止其他用户进程；实际库和Q/媒体配置保持原归属；停止未确认时不开始重操作 | 既有同源服务授权、RESOURCES、主控新增窗口范围 |
| C14 | 验证窗口结束后按已核当前S产物及原SQLite/配置恢复唯一64505服务，重新核进程与实际来源、健康和必要只读数据一致性 | 验证失败仍需按主控安排恢复并保留失败；恢复未知或失败立即报告，不宣称可用、不重发业务、不回退旧源码；不为工具补丁重复构建产品 | 交付候选、RESOURCES、主控新增窗口范围 |

## 测试计划与实施状态

当前准确计划为 `5cd8dbdfa59a0cabe8191e9343ca43122c8afa51` 的 `tasks/swa-m1-contracts-20261020.md` 和 `docs/standards/swa-m1-contracts-20261020.md`，通过主控 Orca ask 回复及原Worker消息`msg_a27d20dcf6d4`取得。结论：**接受修订测试计划与待验证的精确补丁方案**；当前还不能判断运行修复有效。

| 独立条件 | 计划覆盖与结论 |
| --- | --- |
| C01 根因及边界 | 已完整读取本地保存的准确 gitHead wrapper.ts 与 lint.ts；包装器确实按配置前2、配置后1、成功0计算，并在 beforeExit 强制退出。报告明确句柄创建方及补丁因果尚未知，T01/T03用于验证；符合范围 |
| C02 正常三阶段 | T01、T03、T07记录原入口、输出方式及完整三阶段，符合 |
| C03 原规则与真实失败 | T04以真实现存 unused components 提升为error，检查1及链式终止；T05检查配置错误2，符合 |
| C04 示例和能力边界 | T03、T07覆盖正常264项；T05A覆盖原supported capability检查拒绝，脚本和链式命令保持原文，符合 |
| C05 生成逐字检查 | T06使用真实历史文件并核不被覆盖，符合 |
| C06 原链fast-fail | T04覆盖首阶段，T05A覆盖第二阶段及生成阶段未执行，T06覆盖末阶段；诊断、真实退出、原&&语义和完整输出共同判定，符合 |
| C07 直接消费者 | 独立读取check.mjs及process.mjs，runTool('redocly')确实启动同一根node_modules包的bin/cli.js。T04/T05直调覆盖该入口；补丁登记与安装需在实施审核继续核实，无需运行完整pnpm check |
| C08 依赖一致 | T02、T07、T08与精确2.60.0补丁、锁文件和冻结安装相符；无预选升级，符合 |
| C09 验证边界 | T08保留依赖清单、差异、Markdown检查，架构扫描已移除；未启动完整默认集合，符合 |
| C10 隔离与真实输入 | T04真实配置和T06真实历史schema有来源，原生产文件不修改，符合；F01同样须使用真实检查器和有明确用途的公开确定性输入 |
| C11 资源与独立性 | 主控窗口、采样、BelowNormal和真实退出原件均明确；实际执行与Reviewer独立复验仍未开始 |
| C12 交付及权限 | 普通集成、S3、发布分开，准确实现尚未产生，符合当前阶段 |
| C13 服务窗口准入 | T00要求实时核running.json、PID、路径、创建时间、S、唯一监听和当前环境；身份或环境不可用时保留原服务，主控确认停止后才交窗口，符合 |
| C14 同源恢复 | T09按T00当前环境恢复准确S、同目录Web、原库及Q/媒体/CodeGraph配置；记录新身份、健康、唯一监听和真实观察变化；检查失败后也交主控恢复，符合计划范围 |

F01已关闭：T05A在隔离副本将准确32ab契约GET /events的403完整example原样用于同操作501。独立读取原契约确认x-s0-supported=false、无x-s1-supported，ErrorV1允许scope_denied；原validate-examples会在Schema校验之后检查unsupported_capability。实际仍须取得Redocly先成功、第二阶段原诊断、整体退出1、生成终态未出现及文件未写的证据；如果错误提前发生，按计划不计覆盖。

F02已关闭：T08已删除架构扫描，保留与精确补丁直接相关的依赖、差异、文档及实际安装/退出检查。

独立核实的[Node 24.16.0退出说明](https://nodejs.org/download/release/v24.16.0/docs/api/process.html#processexitcode)支持使用exitCode保留退出状态并允许待处理I/O完成；[pnpm原生补丁说明](https://pnpm.io/cli/patch)支持精确版本、显式edit-dir和patch-commit登记。它们支持方案可检验性，不能证明本机句柄归属或修复已经有效。

独立静态读取当前运行根running.json，其source为准确S，记录PID45956、同版本程序与Web、64505及原数据库路径。该读取只核计划采用的来源，不代表重新观察进程活性、数据库或实际恢复。T00/T09的新控制工具限自有目录，实施后先交主控审阅再由主控执行；本轮没有执行服务控制。

本次接受允许主控在M0依赖完成后安排具体写域与独占窗口。实际依赖修复、安装、正常/失败退出与独立复验仍未执行；M0准确实现到达后优先审核。本轮没有发布M1代码CR。

首次准确5fc计划的退回hook为`msg_7c39ff86af0c`，主控在`msg_6f71eac94aa1`明确完整读取并交原Worker，唤起已确认。C13/C14在读取扩展T00/T09计划前补充。当前5cd修订已关闭该次F01/F02，接受只限测试计划及方案边界。

修订接受hook为`msg_38a94c047d3d`，主控在`msg_298a8710f923`明确完整读取C01—C14及F01/F02关闭依据，唤起已确认。Worker先准备自有忽略目录控制脚本交主控静态审阅；正式依赖修改和重操作仍等待M0接受及主控窗口。

## 接续窗口及准确最终记录

主控授权在M0/A01接受后接续本轮，原Worker的新归属为`task_ceb3b3bb76be / ctx_8d6a6e1c0145`，工作树与分支均为`swa-m1-contracts-implementation-20261020`，基线main为`4132f28ecda414f2a1bd040d9eef50bbe1087cb6`。已完整读取准确`987811f417e92cd6f15961e60c279f616e3b21a8`的两份文档并与准确5cd比较，T00—T09及T05A输入和断言没有变化；程序、依赖和契约与S相同。主控将唯一窗口交该Worker，本Reviewer保持只读。

结论：**接受准确`b112ab9e03d9ea82247e73b5cb578529ca529f9d`的调查及实际运行记录**。本次原入口完整通过已核实；历史故障原因及拟议补丁有效性为**证据不足**，没有接受M1缺陷修复。该范围不要求继续实施尚无因果对照的补丁或重复原检查。没有M1程序PR，本轮不发布程序CR。

完整读取了准确fa536e80的任务和报告，以及其准确61e9、b112后继差异。61e9明确列出旧新运行条件，b112区分原PID45956与恢复后PID52040、已停止的实施范围及检查源码9878与文档后继，均符合原件。相对main4132只增加两份M1文档，未修改程序或依赖。原Worker普通保留调查历史，继续保留原失败和未运行项目。

### 实际原件对照

本树原件根为`C:/Users/DW/orca/workspaces/Astrocyte/swa-m1-contracts-implementation-20261020/.verification/swa-m1-contracts-20261020/`。已完整读取`baseline-install-01`及`t01-baseline-01`的命令日志、退出、资源、实际版本、调用记录及进程观察，并重新完整读取原产品树`delivery-candidate-20261020/run-contracts`和`run-redocly`的原始日志、退出、资源及原检查入口。

| 项目 | 独立核实的事实 |
| --- | --- |
| 当前准确运行来源 | 987811f4，普通文档后继为b112；Git比较S→987的package、workspace、lock、contracts、scripts及schema均无差异 |
| 原冻结安装 | 原`pnpm install --frozen-lockfile` EXIT0，UTC07:42:00.1772990—07:42:10.9296330，10.749932秒；355包复用、无下载；源和lock未修改 |
| 原契约命令 | 原`pnpm check:contracts` EXIT0，UTC07:42:29.3471493—07:42:48.4786604，19.1292555秒；真实13警告、264示例和类型逐字一致三阶段均实际完成 |
| 真实工具 | Node v24.16.0、pnpm11.27.0、Redocly2.60.0；完整Node versions及本树真实CLI路径已记录。Node priority10、PowerShell BelowNormal，heap_limit=1056964608字节，即1008MiB |
| 实际资源 | 安装CPU4.6%/12.3%、可用内存10.9533/10.9142GiB，间隔3.0022342秒；契约CPU6.6%/7.1%、10.5914/10.5616GiB，间隔3.0021054秒；均符合启动额度 |
| 原包未补丁 | 使用`fc.exe /b`比较两树实际`bin/cli.js`与`lib/index.js`，两项逐字节相同；原强制退出操作仍在，未登记或实施补丁 |
| 历史失败 | 原S原入口与Redocly直调仍各为-1073740791，valid/13警告后UV_HANDLE_CLOSING断言；这两次原链后继未运行的事实保持 |
| 进程与未运行范围 | 原件保存入口PID7840、观测到的pnpm Node PID18540及其创建/命令；后续对应查询为空。该范围只支持已记录进程观察，不能推导所有历史后代。T02—T08含T05A均NOT_RUN，无补丁、负向或独立动态复验结果 |

历史脚本采用`NODE_OPTIONS=--max-old-space-size=1024`，只记录old-space参数，没有当次完整heap上限；新脚本为`--max-old-space-size=960 --max-semi-space-size=16`，实际上限1008MiB。此外，新工作树使用原锁重新安装，临时及编译缓存目录也不同。原命令、受检源码和包入口相同，尚不构成完全同条件因果比较。没有逐项控制这些差异，不能把未复现归因某项参数、目录、安装、缓存、telemetry或异步句柄。61e9已明确保留这些限制；无需为改变额度重跑。

### 服务记录与当前管理入口

主控报告T00于UTC07:40:57停止原PID45956并保全当前原库和私有环境，T09于UTC07:44:04恢复原S，新PID52040。已只读核原产品树本轮`service-before-01/{backup.json,stopped.json,running-before.json}`及`service-after-01/{restored.json,startup-resources.json}`：备份integrity为ok、停止身份吻合，恢复记录为S、原binary/Web/database/port及health ok；旧新记录中的Q reader/generation/完整pointer和媒体配置相同。启动CPU8.2%/7.0%、可用10.8258/10.8416GiB、单调间隔约3.001秒，符合额度。

真实环境相同、启动前SQLite iterdump相同、主页字节及唯一监听由主控实际执行并通过消息`msg_92d779257991`报告；本Reviewer没有重新执行HTTP、SQLite或服务操作。恢复记录支持本轮已执行结果，后续服务操作仍需重新核实实际身份。私有环境值及数据库未进入Git。

准确main管理提交`bb51d3eee8d638ade0c12e5be3c7ba2acb2dc2d4`已逐项读取，与4132仅五份状态/入口文档不同：限定接受CLI30b调查、M1原入口本次EXIT0、历史FAIL、补丁未实施及原S恢复均与实际来源一致，完整修复继续未接受，原S3未决项保持。真实远端main为bb51。

实际只读验证：

- `.verification/check-revision.ps1 -Previous 4132f28ecda414f2a1bd040d9eef50bbe1087cb6 -Candidate fa536e8048ac4c58636712bf0e3fd4d4929dc88f` EXIT0，两份准确Markdown、5本地链接、无锚点、Git空白通过；61e9差异已逐项读取，仅补充原件已有的运行条件限制。
- `.verification/check-revision.ps1 -Previous fa536e8048ac4c58636712bf0e3fd4d4929dc88f -Candidate 61e9cf63ae7acd838e20e33393030ae8566921e8` EXIT0，两份Markdown、5本地链接、无锚点、Git空白通过；后继`.verification/check-revision.ps1 -Previous 61e9cf63ae7acd838e20e33393030ae8566921e8 -Candidate b112ab9e03d9ea82247e73b5cb578529ca529f9d` EXIT0，一份变动Markdown、2本地链接通过，未变报告复用前项结果。
- `.verification/check-revision.ps1 -Previous 4132f28ecda414f2a1bd040d9eef50bbe1087cb6 -Candidate bb51d3eee8d638ade0c12e5be3c7ba2acb2dc2d4` EXIT0，五份Markdown、76本地链接、无锚点、Git空白通过。
- 已安装marked@18.1.0、BelowNormal、Node512MiB；真实Git来源、原件及包字节比较均为只读。没有重跑原入口、安装、构建、负向测试、模型或服务。

当前可接受的是完整命令本次通过及停止实施、同源恢复的如实记录。补丁尚未实施，原故障内部句柄与发生条件仍未知，未验证负向退出、跨平台及故障修复效果；这些范围保持未接受。当前没有形成可区分原因的新证据，因此没有提出盲目重试或额外补丁计划。

准确b112的限定接受已通过原review-decision hook投递原Run，消息为`msg_4fa60d1d821f`；主控在`msg_5b91a0502f28`明确完整读取本报告，确认本次原入口EXIT0及同源恢复的记录获接受，历史故障原因和补丁有效性继续证据不足，未接受缺陷修复。本次唤起已实际读取并回复。
