# OpenCode DeepSeek V4.1 Flash 实际可用性

根授权 `msg_22459dbf3717`，2026-10-15。本检查独占此文档；临时公开请求与回执只放本树已忽略 `.verification/opencode-ready-20261015/`。模块实际协调客户端为 Codex / gpt-6.1-sol，开发模型检查对象为本机 OpenCode。原 V4 的模型、预算、协议、锁与日程保持。

## 执行计划与已核元数据

本模块直接读取公开 CLI，使用最少执行者；不派新开发人员、不修改全局默认或认证。实际 `opencode --version` 为 `1.18.35`，launcher为 `C:/Users/DW/AppData/Roaming/npm/opencode.ps1`。实际读取主 help、run / models / debug help；`opencode models` 共105个ID，严格匹配用户指定模型为 `alibaba-cn/deepseek-v4.1-flash`。模型列表存在尚未证明实际调用通过。

一次固定公开请求要求只返回 `READY`。用原CLI `--pure run --model alibaba-cn/deepseek-v4.1-flash --format json`，不继续旧会话；独立空cwd、进程内临时权限deny / MCP关闭 / share disabled，沿CLI已有鉴权，不复制或输出密钥。读取真实事件与指定本次会话的标准JSON导出，核终态、精确provider/model元数据、实际READY、工具计数与usage。失败保留原件、不自动重发；没有真实成功时不把该模型配置给简单开发任务。

权限与进程内临时配置依据 [OpenCode官方Permissions](https://opencode.ai/docs/permissions/) 和 [Config](https://opencode.ai/docs/config/)；`OPENCODE_CONFIG_CONTENT`、禁Claude外部提示/skills、禁默认插件与CLI参数依据 [官方CLI](https://opencode.ai/docs/cli/) 及本机1.18.35 help。只使用已核支持参数，不将空MCP对象视为清除继承；当前resolved config有1个MCP名称，临时逐个关闭。

官方 `orca orchestration worker-start` 对 opencode 不支持 `--model`。实际可用后，后续简单任务需原工作树上的 `orca terminal create --command 'opencode --model alibaba-cn/deepseek-v4.1-flash'`，核TUI ready后用合法新Task / `worker-start --terminal`接管；不得自建PTY或凭请求参数声称实际模型。

## 实际调用结果

公开固定请求实际EXIT0、9.5964618秒；原JSON事件为step_start1 / text1 / step_finish1、error0 / tool_use0，唯一text实际为READY。原会话 `ses_ec4a6b6d9ffelPpqUUTqjTyI51` 的原生 `opencode export … --sanitize` 标准JSON包含2条消息；assistant info实际providerID=`alibaba-cn`、modelID=`deepseek-v4.1-flash`、agent / mode=`availability_probe`、finish=`stop`，created / completed均存在。usage为input2159、output3、reasoning23、total2185、cache read/write0；工具parts为0。sanitize导出会隐藏正文，READY由同次原run事件独立核查，没有再次调用模型。

本次证明该原CLI精确配置请求实际完成；provider/model来自原生assistant元数据，未独立抓取provider wire或模型服务回声。当前模型可用于后续简单任务，具体任务的工具、写域和验收仍按其独立授权；没有修改全局默认、账号、认证或正式V4实验。

首前置检查把OpenCode正常resolved permission对象当字符串，model IO前实际FAIL、run输出尚不存在、模型请求0。读取实际resolved schema后，使用`permission.*=deny`检查修正；一次实际READY请求随后完成，未重发已开始的请求。继承的1个MCP逐项临时关闭，进程内权限和禁插件变量在finally恢复；没有写入全局配置。原件仅本树忽略目录 `run.jsonl`、`stderr-local.log`、`session-local.json` 与 `export-stderr-local.log`。

## 独立REPORT

### 2026-10-15 当前许可任务实际启动

同作者两项追加均已合法 succeeded：Task `task_f0b1dca0e6fc` / `ctx_16eab9ed9e91` 的许可补充 SOURCE `da5bbf28eaa4776b6fa267c22bbf3bdf309df224` / REPORT `2341b353cdc29ebeae71aa996b5fc1c8d31f859c`，Task `task_b18a862cfa25` / `ctx_a2f72a4c0bc3` 的准确 adapter 路径 SOURCE `e5d52481fa4a66d66d25497a9c80cc26a82441f1` / REPORT `61cad05f13d56d82a1ac916a5f0027b64c429554`。均复用同 Flash 终端、原工作树和分支，只修改组件文档，远端完整 SHA、单文件与 clean 已核。两项首轮均明确 API socket 关闭且无本任务工具写入，同终端 / 同模型各续接一次后实际读取官方原件或精确 P 源码，未重复 READY 或创建重复 worker，未切换默认模型。最终官方 release 返回 retained / external_terminal / processAction none；完成后的该终端保持空闲，未宣称关闭或归档。

组件许可首次显式 Flash Task `task_08d2ca49a825` 为 Dispatch `ctx_c9d28891657d` / 终端 `term_63d7be11-6caf-4ad6-9206-8ecab071a370`，在原 `release-research-20261014` 工作树与原分支执行，只新增 `docs/release/components-20261015.md`。本轮先用公开 `opencode --help` 核 `--model` / `--auto`，再通过 `orca terminal create --command 'opencode --model alibaba-cn/deepseek-v4.1-flash --auto'` 显式启动，TUI idle satisfied；有界文字显示实际 `Build auto · DeepSeek V4.1 Flash Alibaba (China)` / OpenCode 1.18.35，之后以原 Task / `worker-start --terminal` 合法接管。input_accepted 已核，provider turn-start observation unsupported；后续真实工具执行与文档交付已核，TUI 身份仍不冒称 provider wire 实证。本轮未重复 READY 请求。

首次默认 `--agent opencode` 启动的 `ctx_84d744210f8a` 实际界面为 DeepSeek V4 Pro，产生两个入口文件读取与 `Content Exists Risk` 错误，没有修改授权文件。已停止该配置：首次官方 worker-stop 为 stop_unknown / terminal_tab_close_timeout；正常 `/exit` 后准确进程 53332 不存在且终端返回 PowerShell，随后官方 abandoned / failed，资源 retained / identity_unproven。首失败与外部 shell 保留，当前唯一文档执行者为上面的显式 Flash Dispatch；未修改全局模型或 auth。

本模型可用性SOURCE为 `d895ffcd0f4e58389a570a6c8a9197195b826bdd`，已普通push；本文最后一节是该SOURCE后独立报告，只修改此文档，完整REPORT SHA由实际Git回执提供。D1独立REPORT为 `5a1a4b0474013de9f39eebb910166e7c5af4b822`，范围分别登记，不执行额外模型或日采集。

实际 `opencode --version / --help / models / run --help / --pure debug config`、单次固定 `--pure run -m alibaba-cn/deepseek-v4.1-flash --agent availability_probe --format json`、同次原会话 `export --sanitize` 和标准JSON安全字段核查均通过；最后finish=stop、READY1、error0、工具0、真实模型元数据及usage见上文。`git check-ignore`确认临时目录不入Git，`git diff --check`通过。未安装新模型客户端、修改全局默认/认证、发起第二次probe或启动简单开发任务；后续任务的实际TUI配置与业务检查须由其当前owner核实。
