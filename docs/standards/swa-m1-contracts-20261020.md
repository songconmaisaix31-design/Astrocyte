# M1 Windows 契约检查：根因调查、接受计划与实施记录

本报告支持[本轨任务](../../tasks/swa-m1-contracts-20261020.md)的实施与独立审核。准确 `5cd8dbdfa59a0cabe8191e9343ca43122c8afa51` 的方案与 T00—T09（含 T05A）已接受；本树原依赖安装与 T01 实际退出 0，未复现历史崩溃，补丁因果判断为证据不足。

## 需求与准确来源

主控工作树为 `C:/Users/DW/orca/Astrocyte`；其 `docs/META_PLAN.md` §5.1 明确原有三个串行阶段，§9.1 M1 要求原始命令完整成功，§12 A03 要求契约一致。主控 `tasks/swa-mainline-20261020.md` 要求 M0 后实施，本轮允许先独立只读调查。现产品 [TASKS](../TASKS.md)、[交付报告](delivery-candidate-20261020.md)和主控 `tasks/delivery-candidate-20261020.md` 共同规定保留真实退出和失败原件，先审核方案与测试计划。

源码基线为 `32ab1eec808a5848aabad40f11133735a1214e55`；原实际运行源码 S 为 `ed543c12106c2dbb3e181cb9bc1647b626455179`。Node 路径 `C:/Program Files/nodejs/node.exe` 的文件版本为 `24.16.0.0`；package 要求 Node `24.16.0`、pnpm `11.27.0`。PATH 上 `pnpm.ps1` 经 Corepack 入口调用 Node；其本地已缓存 `pnpm/11.27.0/package.json` 确认为 `11.27.0`。原调查阶段只读取文件，本次实际运行版本已在窗口记录中核实。

当前实际依赖文件及锁文件均为 `@redocly/cli@2.60.0`；其 engines 允许当前 Node 范围。官方 npm 元数据返回 `gitHead=d7177a33a7dde5a4058744dfdfd0000a592998dc`，包完整性与现有锁文件记录一致：`sha512-/xzlR+FGbo2pblWlgIZ92nrLwHgnpE6Ri8kCQm+LaizgijdwMYetly/pv4vdeZ4Hv8S6Zdw/+2fRBlhucbC9qw==`。没有预选或安装其他依赖版本。

## 原始失败与已知成功的边界

原件根 `E0=C:/Users/DW/orca/workspaces/Astrocyte/frontend-observatory-20261018/.verification/delivery-candidate-20261020/`；本次只读取，未覆盖任何原件。

| 原件 | 实际命令及事实 |
| --- | --- |
| `E0/run-contracts/{run.log,result.json,resources.json}` | `pnpm check:contracts`；Redocly 输出 valid、13 项 no-unused-components 警告后断言；退出 `-1073740791`，UTC `2026-10-19T19:30:02.8460016+00:00`；后两个 `&&` 阶段未执行 |
| `E0/run-redocly/{run.log,result.json,resources.json}` | `node node_modules/@redocly/cli/bin/cli.js lint contracts/openapi.yaml --config contracts/redocly.yaml`；相同断言和退出；UTC `2026-10-19T19:32:25.6647030+00:00` |
| `E0/contracts-diagnosis.md`、交付报告 | 独立诊断范围已审，随后独立示例检查 264 项和生成一致性各自退出 0；这些结果未将原入口改为通过 |

两个失败均使用 BelowNormal、Node1024MiB；原入口 CPU 样本 15.4% / 9.7%、可用内存约 11.25 / 11.24 GiB，直调 CPU 9.0% / 7.6%、约 11.28 / 11.32 GiB，单调间隔均约 3 秒。样本只说明启动时满足额度，未证明运行峰值，也不能确定崩溃的内部句柄。

## 本机实现与官方原始来源

已静态读取以下实际调用关系：

1. 根 `package.json` 原命令为 `redocly lint contracts/openapi.yaml --config contracts/redocly.yaml && node scripts/validate-examples.mjs && node scripts/generate-client.mjs --check`。
2. 实际 `.bin/redocly.CMD` 使用 Node 调用同一包的 `bin/cli.js`；该入口启用编译缓存后导入 `lib/index.js`。包直调已有同类失败，故单独更换 pnpm 调用外壳不足以解释或修复现有事实。
3. 本机 `lib/index.js` 的命令包装器 `Q` 对应官方 `commandWrapper`：初值退出码 2，加载配置后设为 1，实际命令成功后设为 0；finally 调用原 telemetry，随后登记 `process.once("beforeExit",()=>{process.exit(i)})`。
4. 实际 lint chunk `DJWI7SF7.js` 对应官方 `handleLint`：运行原 lint、聚合原 errors/warnings、打印结果；存在 error 时抛出 `AbortFlowError`，包装器保留退出码 1。原 13 个 warning 按现规则允许成功。
5. telemetry 原实现包含网络可达性、授权查询和 OpenTelemetry；其异步工作可以影响退出时序。原失败没有 native 调用信息，当前无法把发生错误的句柄确定为 telemetry、fetch、V8 后台任务或编译缓存所创建。
6. `validate-examples.mjs` 使用现有 yaml、Ajv2020 和 ajv-formats，遍历原组件及请求/响应示例并检查 supported capability 边界。`generate-client.mjs --check` 使用原 openapi-typescript 生成结果，与提交类型逐字比较，不一致抛出既有 drift 错误。
7. 已安装 pnpm 的 `verifyPatches` 对未应用补丁默认抛出 `UNUSED_PATCH`，`applyPatchToDir` 对缺失、无效及应用失败分别抛出 `PATCH_NOT_FOUND` / `INVALID_PATCH` / `PATCH_FAILED`。方案使用既有失败语义，不开启容错配置。

原调查已完整读取官方 Redocly 的 wrapper、lint、telemetry 三份对应源码，保存在原产品树 `.verification/swa-m1-contracts-20261020/sources/`；本次仍从原路径只读使用。其他官方来源由原调查通过只读 HTTP 核查，下表分别说明用途：

| 来源 | 支持的判断 |
| --- | --- |
| [Redocly 2.60.0 npm 元数据](https://registry.npmjs.org/@redocly%2fcli/2.60.0) | 发布版本、gitHead 与包完整性 |
| [准确 gitHead 的 wrapper.ts](https://raw.githubusercontent.com/Redocly/redocly-cli/d7177a33a7dde5a4058744dfdfd0000a592998dc/packages/cli/src/wrapper.ts) | 退出码 0/1/2、错误处理及 `beforeExit` 强制退出；本机 minified 实现相符 |
| [准确 gitHead 的 lint.ts](https://raw.githubusercontent.com/Redocly/redocly-cli/d7177a33a7dde5a4058744dfdfd0000a592998dc/packages/cli/src/commands/lint.ts) | 实际校验、警告和规则失败路径 |
| [准确 gitHead 的 telemetry.ts](https://raw.githubusercontent.com/Redocly/redocly-cli/d7177a33a7dde5a4058744dfdfd0000a592998dc/packages/cli/src/utils/telemetry.ts) | 当前命令存在的异步外部操作，无法单凭其存在判定肇因 |
| [Node 24.16.0 process 原文](https://raw.githubusercontent.com/nodejs/node/v24.16.0/doc/api/process.md) | process.exit 同步终止；推荐以 exitCode 指定自然退出状态并完成待处理 I/O |
| [Node 24.16.0 Windows async.c](https://raw.githubusercontent.com/nodejs/node/v24.16.0/deps/uv/src/win/async.c) | uv_async_send 对 closing 句柄有断言；这只是句柄状态解释 |
| [Node 官方问题 56645](https://github.com/nodejs/node/issues/56645) | Windows 强制退出存在同类崩溃；上游记录只能辅助形成假设 |
| [pnpm patch 官方说明](https://pnpm.io/cli/patch) | 精确包补丁、显式 edit-dir、patch-commit 登记；具体行为同时以本机 11.27.0 实现核实 |

Node 原件的断言行号为 94，本次取得的对应 tag 源码断言位于 87；当前缺少当次二进制来源和 native 调用信息，保留该差异，不以相同行号或上游 issue 状态证明本机因果。官方 Redocly `v2.60.0` tag URL 未取得，已改按 npm 元数据中的准确 gitHead 读取，未猜测其他版本。

## 可检验的根因判断

现有事实把故障定位到“校验已经完成后的进程退出阶段”。最有依据的修改边界是 Redocly 命令包装器的强制退出：原规则通过后得出 code=0，但它在自然退出事件里再次调用同步 `process.exit`。官方 Node 文档说明可以保留 code 并自然退出，Windows 上游记录与当前断言相符。

判断仍有两项未知：本机具体异步句柄的创建方、只改变该退出操作是否足以消除崩溃。测试将固定其他条件，仅修改该依赖包的这一处操作。如果相同条件下改前成功、改后仍崩溃或出现其他退出异常，当前因果判断保留为证据不足或被否定，先更新计划后继续。本次 T01 改前成功，已停止补丁实施并交原 Reviewer。没有把网络失败、pnpm、资源或契约内容直接认定为根因。

## 一个修复方案与具体影响

通过 pnpm 原生精确依赖补丁，把 `@redocly/cli@2.60.0` 的命令包装器 `beforeExit` 回调中的 `process.exit(i)` 改为 `process.exitCode=i`。回调仍只执行一次，计算 code 的原控制流及错误处理保持。补丁只作用于项目安装的该包，版本和所有规则保持。补丁登记写入 `pnpm-workspace.yaml`，锁文件由当前 pnpm 更新，补丁存入 `patches/@redocly__cli@2.60.0.patch`；实际实现写域以主控答复为准。

仅编辑已授权的包工作副本，使用文件编辑工具；安装后的 package 由 pnpm 应用补丁。使用 pnpm 明确的 `--edit-dir` 将工作副本放在 E 内。原 `package.json` 的命令字符串、原 Redocly config、契约、示例脚本及生成器均继续执行。没有新增等待、重试、Skip、错误拦截、成功码替换、预加载改写 process 方法或系统环境改动。原有外部依赖错误处理不在本次扩大修改。

| 可观察行为 | 改前 | 预期改后 |
| --- | --- | --- |
| 原入口 | Redocly 校验有效后 native 崩溃；后两阶段未执行 | 保留 13 警告；依次实际输出 264 项验证和类型一致，退出 0 |
| 原 CLI 的真实 lint error | 包装器应保留 1；Windows 最终退出尚待比较 | 输出实际错误和退出 1，后续链式阶段不执行 |
| 实际不存在的 config 路径 | 包装器加载前错误，code=2 | 仍为真实配置错误和退出 2 |
| 真实历史 schema 与当前契约不匹配 | 原生成脚本具有逐字漂移拒绝 | 仍抛出原 drift 错误并非零退出；原生成文件不被覆盖 |
| 冻结安装 | 原 CLI 未带项目补丁 | 从项目补丁与锁文件安装同版真实 CLI，再执行原入口 |

## 已接受测试计划

下表保留准确 `5cd8` 接受的输入和断言，实际状态单列于本树实施记录。测试使用真实原契约、原 CLI 和真实历史生成文件，不构造协议、CLI 替身或业务数据。修改依赖退出行为的相关边界由真实子进程检查；不为这一修改添加只复述实现的纯函数测试。原命令集成与干净副本安装分别覆盖项目调用和交付后的依赖复现。

| 编号 / 覆盖要求 | 具体输入与操作 | 必须观察的结果 |
| --- | --- | --- |
| T00 当前服务停止与保全（主控） | 重新读取原运行根 `running.json`，核 PID、精确创建时间、程序路径、S、BelowNormal 及唯一 `127.0.0.1:64505` 监听；保全当前 SQLite、完整 Q 指针、媒体配置和 CodeGraph 配置，取得真实进程的当前环境用于同源恢复。主控核实后停止已确认进程并等待退出 | 记录实际原状态和当前配置来源；只有该进程停止且端口无监听才交付重操作窗口。读取环境失败、身份变化或存在其他占用时立即报告，保持原服务。敏感环境只保存在私有忽略目录，不打印值、不进入 Git |
| T01 改前对照 | 获批窗口中以干净基线与当前真实依赖运行一次原 `pnpm check:contracts`，沿原 pipe 输出方式；先记录 `process.version`、`process.versions`、execPath、pnpm 版本及环境额度 | 记录完整原退出和三个阶段实际发生情况；发生原崩溃则保持 FAIL；没有复现时不得宣告已完成因果对照 |
| T02 安装修复 | E 内明确 edit-dir 准备精确 2.60.0 包副本，经文件编辑工具仅改该退出操作；使用当前 pnpm 登记补丁和锁文件，再冻结安装 | 补丁被实际应用；锁文件及实际包中仅出现预期变动；依赖仍为同版，原命令及检查输入相同。不得设置允许未用补丁或忽略补丁失败 |
| T03 原入口成功 | `pnpm check:contracts`，原契约、原配置和未改脚本，保持与 T01 同一输出方式及额度 | EXIT0；Redocly 原规则输出有效与 13 警告，`Validated 264 OpenAPI examples and supported capability boundaries`、`OpenAPI generated types match source` 均实际出现；没有 native 断言、遗漏阶段或残留本次子进程 |
| T04 CLI 规则失败 | E 内只复制真实 `contracts/redocly.yaml`，通过文件编辑工具把真实现存 13 项 unused component 对应规则 `no-unused-components` 的级别指定为 error；对原 `contracts/openapi.yaml` 调用真实 `node node_modules/@redocly/cli/bin/cli.js lint ... --config <E内配置>` | 实际规则错误、退出 1、诊断完整且没有 native 崩溃。随后在 E 的原文件副本中用该配置运行原 package 命令，确认两个后续阶段没有输出；保留这项故意拒绝结果 |
| T05 CLI 配置错误 | 同一真实 CLI、原契约，`--config` 指向 E 内实际不存在的 yaml | 原错误 `Please provide a valid path to the configuration file.`，退出 2；没有成功信息或 native 崩溃 |
| T05A 示例与能力边界失败 | E 内真实源码副本中，通过文件编辑工具把原 `GET /events` 的 403 响应完整 example 原样复制到该操作的 501 响应 example；其余契约、原配置和三个脚本及 package 命令保持。输入来自公开准确基线 `32ab1eec808a5848aabad40f11133735a1214e55:contracts/openapi.yaml` 的该 403 example，code 为既有 `scope_denied`，仍满足原 ErrorV1 Schema；不生成新值或调用业务服务。运行原 `pnpm check:contracts` | Redocly 完成原规则校验、退出成功，原示例验证器随后抛出 `Future operation get /events must document unsupported_capability`，原链整体退出 1。完整 stdout/stderr 中没有 `Validated 264 ...` 终态，也没有生成类型一致或 drift 终态；结合原样保留的 `&&` 命令及第二阶段实际未捕获错误，判定生成阶段未运行。原生成文件字节保持。故意拒绝结果单独保存 |
| T06 类型漂移拒绝 | 在 E 的完整真实源码副本中，仅将 `web/src/api/schema.d.ts` 换成 Git 中真实 `f48122ba5eb4f469bf13cb0e27539942ceee3841` 的该文件；当前历史比较确认其与基线存在 100 行差异。使用成熟归档工具取得 Git 原文件，禁止改写生产文件；运行原 `pnpm check:contracts` | 前两阶段实际通过；原 `generate-client --check` 抛出 `Generated client schema drift: run pnpm generate and commit schema.d.ts`，整体非零退出；历史 schema 完整字节保持，证明 --check 未写文件 |
| T07 干净冻结安装 | 用最终准确 Commit 的真实归档创建 E 内独立副本，包含原 package、workspace、lock、patch、contracts、scripts 和生成类型；副本初始无 node_modules。在唯一窗口内运行原 `pnpm install --frozen-lockfile`，随后原 `pnpm check:contracts`；不使用业务/CLI 替身 | 安装及命令均退出 0；三个阶段实际完成；同版精确依赖补丁可重建，锁文件无自动漂移。该步骤为 Windows 本地证据；未执行 Linux 时不扩写跨平台通过 |
| T08 适用静态检查 | 准确最终差异核对；`git diff --check`、`node scripts/check-dependencies.mjs`，按各自额度串行；两份 Markdown 通过已安装成熟解析器检查本地链接 | 命令退出 0；原规则和生成物未发生修改，依赖差异限定为该精确包的项目补丁；不触发完整默认 CI 中其他受阻路径 |
| T09 同一 S 和原库恢复（主控） | Worker 确认所有本次子进程退出、交还窗口；主控按 T00 取得的当前 `running.json`、真实环境与配置，重新采样后启动原 `versions/ed543c12/astrocyte.exe`、同目录 Web 和当前原 SQLite，保持原 Q/媒体/CodeGraph 配置及 BelowNormal/额度；核实际健康和唯一监听，登记新的 PID/创建时间 | 程序仍为准确 S、Web/原库路径保持、真实配置可追溯；只有新进程身份、健康和唯一监听实际通过才更新运行记录并报告恢复。启动失败保存原件并报告；没有自动切换 c5、恢复旧数据库或重试 |

T04 / T06 隔离副本使用真实原文件与原 package 命令；可以复用同一实际安装的 node_modules 链接，必须记录解析的实际 CLI 路径，T07 则独立安装。副本生成、历史文件取得、命令启动及原件保存脚本均只位于 E 内；脚本在独立接受后才编写并执行。测试执行不写真实 SQLite、不访问业务 HTTP、不产生模型调用。

T04 的规则级别变化只在负向检查副本中使用，不进入生产配置。T06 的旧文件有真实 Git 来源；当前产品 schema 保持。T03 和 T07 以退出码及三个实际阶段输出共同判定通过，不能仅匹配 success 文案。原 CLI / pnpm / 检查器产生的非零结果逐项保存，测试报告明确区分预期拒绝与功能失败。

T05A 的公开输入为[准确基线契约](https://github.com/songconmaisaix31-design/swarm-anything/blob/32ab1eec808a5848aabad40f11133735a1214e55/contracts/openapi.yaml)中 `#/paths/~1events/get/responses/403/content/application~1json/example`，只在隔离副本复制到同操作 `501` 的 example。原 `GET /events` 标记 `x-s0-supported: false`，没有 `x-s1-supported: true`；原 ErrorV1.enum 明确允许 `scope_denied`，故 Schema 检查可继续，随后原 supported capability 检查会拒绝该未来操作的 501 example。这个负向检查验证原脚本的真实契约边界，不表示业务接口曾返回该结果。验收同时保存两个 example 的来源、隔离文件实际差异、原脚本/命令未改的比较结果、整体实际退出及完整输出；若先在 Redocly 阶段拒绝或出现其他诊断，则 T05A 未覆盖目标，须报告并修订计划，不能按预期补写第二阶段已执行。

## 本树实施记录

本次固定 Worktree 为 `C:/Users/DW/orca/workspaces/Astrocyte/swa-m1-contracts-implementation-20261020`，Branch 同名，Task `task_ceb3b3bb76be` / Dispatch `ctx_8d6a6e1c0145`，仍属原 Run `run_62ada397a104`。接管 HEAD `4132f28ecda414f2a1bd040d9eef50bbe1087cb6` 干净，主控确认 M0/A01 已独立接受；普通合并 `4f49d6ea3c0cfcfa6bdf4143d576239b158fe878` 保留 `80d4f8d1e323982ef73d880ce9927f4bc5c58340` 的两份 M1 文档及其原历史。

实际 Git 比较确认 `80d4` 对 `32ab` 仅增加本任务和报告，`5cd8` 至 `80d4` 未修改测试表。本树 package、pnpm workspace/lock、contracts、scripts 和生成类型与 `32ab`、S 均无差异；已接受计划可用于当前源。已完整读取本树入口、规范和资源文件，根主控当前首版 CLI 任务及 QUESTIONS，准确两份 Git 原文、独立审核报告、原两次失败的结果与日志、原 `run-check.ps1` / `runtime-info.mjs` / `CONTROL.md` 和资源脚本。

本次实际会话为 Orca 1.4.230 / Codex，`worker-show --dispatch ctx_8d6a6e1c0145` 的实际 projection 为 `gpt-6.1-sol`，终端实际预览为 `GPT-6.1-Sol xhigh fast`；请求中的模型字段为 null，没有据此推断模型，没有子 Agent。首版是否独立运行仍未回答；本修复只支持既有契约命令。

新树 `E=.verification/swa-m1-contracts-20261020/` 已被现有 `.gitignore` 忽略。通过文件编辑工具准备本树 `run-check.ps1`、`runtime-info.mjs`、`resources.py` 和 `CONTROL.md`：沿原脚本的真实命令、合并 pipe 输出、准确源码、资源采样和真实退出设计；资源入口移到本 E，补充实际 PID/优先级和耗时。原产品树未修改，服务控制只由主控使用原已审脚本执行。本树初始没有 node_modules，先在获准窗口冻结安装原锁文件，随后 T01；补丁登记后保存准确程序 Commit，后续冻结安装与 T07 归档均使用该 Commit。

| 阶段 | 本树实际命令 / 输入 | 独立输出目录（E 内） | 当前结果 |
| --- | --- | --- | --- |
| T00 | 主控执行原树已审服务控制脚本；停止当前同 S 服务并保全原库与环境 | 原树 E/service-before-01 | 主控报告 PASS，消息 msg_17851dde26c8 |
| 准备 | `pnpm install --frozen-lockfile`，原 workspace/lock | `baseline-install-01` | EXIT0 |
| T01 | 原 `pnpm check:contracts`，未补丁依赖 | `t01-baseline-01` | EXIT0；本次未复现历史崩溃，补丁因果证据不足 |
| T02 | `pnpm patch @redocly/cli@2.60.0 --edit-dir <本E/package-edit>`；文件编辑工具修改准确包；`pnpm patch-commit <本E/package-edit>`；`pnpm install --frozen-lockfile` | `t02-prepare-01`、`t02-register-01`、`t02-install-01` | NOT_RUN |
| T03 | 原 `pnpm check:contracts` | `t03-contracts-01` | NOT_RUN |
| T04 | 真实 CLI + 原规范 + E 内 error 配置；隔离副本中的原 `pnpm check:contracts` | `t04-cli-01`、`t04-chain-01` | NOT_RUN |
| T05 | 真实 CLI + 原规范 + E 内不存在的配置路径 | `t05-config-01` | NOT_RUN |
| T05A | 隔离副本复制原 GET /events 403 example 至同操作 501；原 `pnpm check:contracts` | `t05a-capability-01` | NOT_RUN |
| T06 | 隔离副本采用准确历史 schema；原 `pnpm check:contracts` | `t06-drift-01` | NOT_RUN |
| T07 | 准确程序 Commit 的独立归档；原冻结安装及原契约命令 | `t07-install-01`、`t07-contracts-01` | NOT_RUN |
| T08 | `git diff --check`、`node scripts/check-dependencies.mjs`、成熟 Markdown 解析器本地链接检查 | `t08-dependencies-01`、文档检查原件 | NOT_RUN |
| T09 | 本次命令结束后交还窗口；主控恢复同 S、Web、原库和配置 | 原树 E/service-after-01 | 主控报告 PASS，消息 msg_7d11251e5e36 |

每次以独立 PowerShell 进程调用本 E 的 run-check，填写准确完整 ExpectedSource、新 OutputName 和明确窗口标记；副本使用 CheckRoot，CLI 负向配置使用 ConfigPath。原 `pnpm check:contracts` 字符串保持。实际 stdout/stderr 合并保存到 check.log，另存 run.log、invocation.json、resources.json、runtime.json、pnpm-version.txt、result.json；实际负向退出保留，不覆盖为通过。每个安装/检查启动前重新采样，原服务恢复期间不运行检查。

### 首次实际窗口结果

主控 `msg_17851dde26c8` 明确 T00 于 `2026-10-20T07:40:57Z` 完成，原 PID45956 已退出、64505 空闲；当前原库一致备份和环境/Q/媒体配置保全成功。窗口只用于已接受计划；本 Worker 没有停止或恢复服务。

实际执行命令如下，均由本 E 的 run-check 原样调用目标命令：

```powershell
pwsh -NoLogo -NoProfile -File .verification/swa-m1-contracts-20261020/run-check.ps1 -Phase install -ExpectedSource 987811f417e92cd6f15961e60c279f616e3b21a8 -OutputName baseline-install-01 -CoordinatorWindow
pwsh -NoLogo -NoProfile -File .verification/swa-m1-contracts-20261020/run-check.ps1 -Phase contracts -ExpectedSource 987811f417e92cd6f15961e60c279f616e3b21a8 -OutputName t01-baseline-01 -CoordinatorWindow
```

| 项目 | 原依赖冻结安装 | T01 原命令 |
| --- | --- | --- |
| 实际退出与耗时 | 0，10.749932 秒 | 0，19.1292555 秒 |
| 目标命令 UTC | 07:42:00.1772990 至 07:42:10.9296330 | 07:42:29.3471493 至 07:42:48.4786604 |
| 资源样本 UTC | 07:41:56.388392 / 07:41:59.390621 | 07:42:25.801555 / 07:42:28.803656 |
| CPU / 可用内存 | 4.6% / 10.9533 GiB；12.3% / 10.9142 GiB | 6.6% / 10.5914 GiB；7.1% / 10.5616 GiB |
| 实际单调间隔 | 3.0022342 秒 | 3.0021054 秒 |
| 结果 | 355 个包复用，无下载；原 lock 无变化 | Redocly valid、13 warnings；`Validated 264 OpenAPI examples and supported capability boundaries`；`OpenAPI generated types match source` |

两次均为 Node `v24.16.0`、pnpm `11.27.0`、实际 CLI `2.60.0`，Node 路径 `C:/Program Files/nodejs/node.exe`；实际 heap 为 `1056964608` 字节（1008 MiB），Node priority=10 与 PowerShell BelowNormal 相符，资源采样进程 priority=16384。T01 的实际 CLI 路径为本树 `node_modules/.pnpm/@redocly+cli@2.60.0/node_modules/@redocly/cli/bin/cli.js`。完整 versions 保存在 runtime.json。调用保留 pipe 输出和原 telemetry 行为，没有改动 package 命令、规范、规则、脚本、生成类型、workspace 或 lock。

T01 入口 PID7840，采样中观察到真实子进程 Node PID18540、创建本地时间 `2026-10-20 15:42:29`；实际命令行保存在 children-snapshot.json。命令返回后实际进程查询确认这两个 PID 及其当前直接子进程均不存在，查询结果保存在 residual-processes.json。原服务仍由主控管理，未对业务 HTTP 或 SQLite 运行检查。

T01 未复现原失败，已接受计划中“改前成功则因果证据不足，先更新计划”的条件触发。没有实施 T02、没有生成补丁、没有自动重复原命令；T02—T08 保持 NOT_RUN。已通过 `msg_596d46c36d91` 报告并交还唯一窗口，主控已将实际结果交原 Reviewer。两次历史 FAIL 与本次 EXIT0 分别保留；本次 EXIT0 不能证明退出补丁有效。

主控 `msg_7d11251e5e36` 报告 T09 已完成：新 PID52040、创建 UTC `2026-10-20T07:44:04.137854Z`，原 S 程序、同目录 Web、原 cwd、完整环境、原库及 Q/媒体/CodeGraph 配置保持。主控核实停止期间当前原库与 T00 备份 iterdump 逐字相同，启动资源合法、健康成功、主页字节与同源 Web 一致，唯一 loopback 监听、实际进程身份/优先级/环境通过，running.json 已按新身份更新。原件在原产品树 `E/service-after-01`；本 Worker 未执行这些服务/数据库检查。原服务重新占用窗口，后续动态操作停止；本次只提交、push 真实调查文档，不创建修复 PR。

## 资源、审核与接续

主控继续管理原 64505 服务；本次窗口已交还并由主控恢复服务。每个实际安装、测试或命令验收前，先取得主控授予的独占窗口，确认服务已由主控按授权处理；随后两次真实 CPU/可用内存采样，UTC 与单调计时记录齐全，间隔 2—5 秒、CPU 均小于 85%、内存均至少 4 GiB。进程采用 BelowNormal、Node heap 不超过 1024 MiB；本次检查的临时目录和缓存限定 E，原服务恢复沿真实配置。已有窗口或采样不复用为下一重操作的额度。静态审查与文档检查遵守现有 RESOURCES 专用范围。

原调查期间主控 `msg_656f36f44f3e` 核实的服务身份为 PID45956、创建 UTC `2026-10-19T19:47:04.6572166Z`、S 路径和唯一 loopback 监听；该进程在本次 T00 已停止。原只读 running.json 的 source 为准确 S；旧 `candidate-service.py` 的 original_identity 和 start 前置仍绑定 c5，与旧恢复脚本均不用于本次同源恢复。当前新身份以本节 T09 实际记录为准。

T00/T09 若需要新控制工具，限 `E/service-window.py` 等自有文件，先审核后编写，再交主控审阅执行。工具依当前实际记录限定程序、Web、原库、端口及进程身份，成熟 psutil 读取真实进程环境，成熟 SQLite backup API 保全原库，不直接写业务表；缺少权限或配置立即失败。保全 Q 的完整指针及原媒体/CodeGraph 配置，禁止输出秘密值。停止与恢复均由主控执行；保存失败或资源拒绝时保留原状态，不使用旧库覆盖当前业务，不编造 stage.json，不修改旧已验脚本。服务自身正常启动的既有观察更新时间按真实启动区间保留；出现业务值变化交主控核查。

各项前后记录准确 Commit、工具版本、输入、命令、stdout/stderr、实际退出、采样、时长及实际残留进程。改后与改前同条件比较，没有以成功结果覆盖失败。出现资源不足、非预期包变动、不同警告/规则集合、崩溃或错误码变化时立即报告，未验范围继续保留；不自动恢复数据或停止其他进程。

独立 Reviewer 已接受准确 `5cd8dbdfa59a0cabe8191e9343ca43122c8afa51` 的方案和全部 T00—T09（含 T05A）；独立报告为 `swa-mainline-review-20261020/docs/standards/swa-m1-contracts-review-20261020.md`，hook `msg_38a94c047d3d` 已由主控完整读取并回复 `msg_298a8710f923`。本次按其停止条件保留 T01 未复现结果；补丁方案需要原 Reviewer 处理证据不足后才能继续。

原调查 Task `task_261b4dab6d89 / ctx_958a2bf0ae27` 已正式结算，原脚本和证据仍在原产品树。主控已完整静态审阅该树 `service-window.py`、`CONTROL.md` 和检查入口，并核旧 PowerShell Parser 通过；服务控制只由主控执行。原当前 S、Web、原库与环境的停止/恢复仍按 T00/T09，不沿用旧 c5 控制脚本。

本次静态准备中的 PowerShell `Language.Parser.ParseFile` 只读解析通过，`git diff --check` 退出 0，本地 resources.py 与原资源入口逐字比较无差异，逐项 BelowNormal 执行。随后在明确窗口中运行一次原冻结安装与一次 T01，实际结果如上，没有实施依赖补丁、编译、其他功能检查或系统环境修改。文档静态检查使用已安装 `marked@18.1.0`，`node .verification/swa-m1-contracts-20261020/check-docs.mjs` 退出 0，核两份文档 5 个本地链接和已接受测试表原文保持，实际 heap `520093696` 字节（496 MiB）、priority=10；原件为 E/doc-check-01.log。该检查只覆盖文档，T08 的依赖程序检查仍未执行。主控已要求本次准确调查文档普通提交、push 后回传 SHA，等待原独立 Reviewer `task_1cd8014750f9 / ctx_980a3282c339` 的范围结论；不创建修复 PR。

当前程序修复有效性为 NOT_RUN；两次原始崩溃保持 FAIL，本次准确源原契约命令 EXIT0 单列保留。没有创建或更新 PR，不宣称历史故障已修复、首版 CLI、完整 CI 或 S3 已通过。原 Reviewer 处理补丁因果证据不足，主控负责最终检查与集成。
