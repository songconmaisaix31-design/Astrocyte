# M1 Windows 契约检查：只读根因调查与测试计划

本报告支持[本轨任务](../../tasks/swa-m1-contracts-20261020.md)的独立计划审核。当前结论为可检验的根因判断，运行验证尚未执行。

## 需求与准确来源

主控工作树为 `C:/Users/DW/orca/Astrocyte`；其 `docs/META_PLAN.md` §5.1 明确原有三个串行阶段，§9.1 M1 要求原始命令完整成功，§12 A03 要求契约一致。主控 `tasks/swa-mainline-20261020.md` 要求 M0 后实施，本轮允许先独立只读调查。现产品 [TASKS](../TASKS.md)、[交付报告](delivery-candidate-20261020.md)和主控 `tasks/delivery-candidate-20261020.md` 共同规定保留真实退出和失败原件，先审核方案与测试计划。

源码基线为 `32ab1eec808a5848aabad40f11133735a1214e55`；原实际运行源码 S 为 `ed543c12106c2dbb3e181cb9bc1647b626455179`。当前 Node 路径 `C:/Program Files/nodejs/node.exe` 的文件版本为 `24.16.0.0`；package 要求 Node `24.16.0`、pnpm `11.27.0`。PATH 上 `pnpm.ps1` 经 Corepack 入口调用 Node；其本地已缓存 `pnpm/11.27.0/package.json` 确认为 `11.27.0`。本轮没有启动 Node 或 pnpm 以获取运行结果，实际运行版本将在获准窗口中记录。

当前实际依赖文件及锁文件均为 `@redocly/cli@2.60.0`；其 engines 允许当前 Node 范围。官方 npm 元数据返回 `gitHead=d7177a33a7dde5a4058744dfdfd0000a592998dc`，包完整性与现有锁文件记录一致：`sha512-/xzlR+FGbo2pblWlgIZ92nrLwHgnpE6Ri8kCQm+LaizgijdwMYetly/pv4vdeZ4Hv8S6Zdw/+2fRBlhucbC9qw==`。没有预选或安装其他依赖版本。

## 原始失败与已知成功的边界

原件根 `E0=.verification/delivery-candidate-20261020/`；本轮只读取，未覆盖任何原件。

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

官方 Redocly 的 wrapper、lint、telemetry 三份对应源码已完整读取并保存到自有忽略目录 `E=.verification/swa-m1-contracts-20261020/sources/`；其他官方来源通过只读 HTTP 核查，下表分别说明用途：

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

判断仍有两项未知：本机具体异步句柄的创建方、只改变该退出操作是否足以消除当前崩溃。测试将固定其他条件，仅修改该依赖包的这一处操作。如果相同条件下改前成功、改后仍崩溃或出现其他退出异常，当前因果判断保留为证据不足或被否定，先更新计划后继续。没有把网络失败、pnpm、资源或契约内容直接认定为根因。

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

## 待审核测试计划

本轮所有条目均为 **NOT_RUN**。测试使用真实原契约、原 CLI 和真实历史生成文件，不构造协议、CLI 替身或业务数据。修改依赖退出行为的相关边界由真实子进程检查；不为这一修改添加只复述实现的纯函数测试。原命令集成与干净副本安装分别覆盖项目调用和交付后的依赖复现。

| 编号 / 覆盖要求 | 具体输入与操作 | 必须观察的结果 |
| --- | --- | --- |
| T01 改前对照 | 获批窗口中以干净基线与当前真实依赖运行一次原 `pnpm check:contracts`，沿原 pipe 输出方式；先记录 `process.version`、`process.versions`、execPath、pnpm 版本及环境额度 | 记录完整原退出和三个阶段实际发生情况；发生原崩溃则保持 FAIL；没有复现时不得宣告已完成因果对照 |
| T02 安装修复 | E 内明确 edit-dir 准备精确 2.60.0 包副本，经文件编辑工具仅改该退出操作；使用当前 pnpm 登记补丁和锁文件，再冻结安装 | 补丁被实际应用；锁文件及实际包中仅出现预期变动；依赖仍为同版，原命令及检查输入相同。不得设置允许未用补丁或忽略补丁失败 |
| T03 原入口成功 | `pnpm check:contracts`，原契约、原配置和未改脚本，保持与 T01 同一输出方式及额度 | EXIT0；Redocly 原规则输出有效与 13 警告，`Validated 264 OpenAPI examples and supported capability boundaries`、`OpenAPI generated types match source` 均实际出现；没有 native 断言、遗漏阶段或残留本次子进程 |
| T04 CLI 规则失败 | E 内只复制真实 `contracts/redocly.yaml`，通过文件编辑工具把真实现存 13 项 unused component 对应规则 `no-unused-components` 的级别指定为 error；对原 `contracts/openapi.yaml` 调用真实 `node node_modules/@redocly/cli/bin/cli.js lint ... --config <E内配置>` | 实际规则错误、退出 1、诊断完整且没有 native 崩溃。随后在 E 的原文件副本中用该配置运行原 package 命令，确认两个后续阶段没有输出；保留这项故意拒绝结果 |
| T05 CLI 配置错误 | 同一真实 CLI、原契约，`--config` 指向 E 内实际不存在的 yaml | 原错误 `Please provide a valid path to the configuration file.`，退出 2；没有成功信息或 native 崩溃 |
| T06 类型漂移拒绝 | 在 E 的完整真实源码副本中，仅将 `web/src/api/schema.d.ts` 换成 Git 中真实 `f48122ba5eb4f469bf13cb0e27539942ceee3841` 的该文件；当前历史比较确认其与基线存在 100 行差异。使用成熟归档工具取得 Git 原文件，禁止改写生产文件；运行原 `pnpm check:contracts` | 前两阶段实际通过；原 `generate-client --check` 抛出 `Generated client schema drift: run pnpm generate and commit schema.d.ts`，整体非零退出；历史 schema 完整字节保持，证明 --check 未写文件 |
| T07 干净冻结安装 | 用最终准确 Commit 的真实归档创建 E 内独立副本，包含原 package、workspace、lock、patch、contracts、scripts 和生成类型；副本初始无 node_modules。在唯一窗口内运行原 `pnpm install --frozen-lockfile`，随后原 `pnpm check:contracts`；不使用业务/CLI 替身 | 安装及命令均退出 0；三个阶段实际完成；同版精确依赖补丁可重建，锁文件无自动漂移。该步骤为 Windows 本地证据；未执行 Linux 时不扩写跨平台通过 |
| T08 适用静态检查 | 准确最终差异核对；`git diff --check`、`node scripts/check-architecture.mjs`、`node scripts/check-dependencies.mjs`，按各自额度串行；两份 Markdown 通过已安装成熟解析器检查本地链接 | 命令退出 0；原规则和生成物未发生修改，依赖差异限定为该精确包的项目补丁；不触发完整默认 CI 中其他受阻路径 |

T04 / T06 隔离副本使用真实原文件与原 package 命令；可以复用同一实际安装的 node_modules 链接，必须记录解析的实际 CLI 路径，T07 则独立安装。副本生成、历史文件取得、命令启动及原件保存脚本均只位于 E 内；脚本在独立接受后才编写并执行。测试执行不写真实 SQLite、不访问业务 HTTP、不产生模型调用。

T04 的规则级别变化只在负向检查副本中使用，不进入生产配置。T06 的旧文件有真实 Git 来源；当前产品 schema 保持。T03 和 T07 以退出码及三个实际阶段输出共同判定通过，不能仅匹配 success 文案。原 CLI / pnpm / 检查器产生的非零结果逐项保存，测试报告明确区分预期拒绝与功能失败。

## 资源、审核与接续

主控继续管理原 64505 服务；当前只读调查没有取得重操作窗口。每个实际安装、测试或命令验收前，先取得主控授予的独占窗口，确认服务已由主控按授权处理；随后两次真实 CPU/可用内存采样，UTC 与单调计时记录齐全，间隔 2—5 秒、CPU 均小于 85%、内存均至少 4 GiB。进程采用 BelowNormal、Node heap 不超过 1024 MiB；临时目录和缓存限定 E。已有窗口或采样不复用为下一重操作的额度。静态审查与文档检查遵守现有 RESOURCES 专用范围。

各项前后记录准确 Commit、工具版本、输入、命令、stdout/stderr、实际退出、采样、时长及实际残留进程。改后与改前同条件比较，没有以成功结果覆盖失败。出现资源不足、非预期包变动、不同警告/规则集合、崩溃或错误码变化时立即报告，未验范围继续保留；不自动恢复数据或停止其他进程。

本轮已执行的操作为文件和 Git 只读核查、Orca 状态/消息核查及官方原文 HTTP GET；未运行任何复现、依赖安装、编译、测试或架构/依赖业务脚本。当前只新增任务和报告，外部原始源码保存在 E/sources。计划待主控通过 Orca 转交独立 Reviewer；M0 完成、方案获接受和实现写域确定后，同一 Worker 继续实现与自测。PR、普通 push、准确 Commit 审查及主线集成按主控安排；当前未执行。
