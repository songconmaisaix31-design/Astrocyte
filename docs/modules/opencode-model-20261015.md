# OpenCode DeepSeek V4.1 Flash 实际可用性

> 文档用途：需求与历史记录。保留本文件独有业务条款、来源、实际检查和未验限制；原日期的人员分工、资源额度与操作安排只适用于当时阶段。后续执行读取 [AGENTS](../../AGENTS.md)、[RESOURCES](../RESOURCES.md) 和 [当前任务目录](../TASKS.md)，独立 Reviewer 按准确来源验收；归档与入口调整不改变功能完成状态。

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

本模型可用性SOURCE为 `d895ffcd0f4e58389a570a6c8a9197195b826bdd`，已普通push；本文最后一节是该SOURCE后独立报告，只修改此文档，完整REPORT SHA由实际Git回执提供。D1独立REPORT为 `5a1a4b0474013de9f39eebb910166e7c5af4b822`，范围分别登记，不执行额外模型或日采集。

实际 `opencode --version / --help / models / run --help / --pure debug config`、单次固定 `--pure run -m alibaba-cn/deepseek-v4.1-flash --agent availability_probe --format json`、同次原会话 `export --sanitize` 和标准JSON安全字段核查均通过；最后finish=stop、READY1、error0、工具0、真实模型元数据及usage见上文。`git check-ignore`确认临时目录不入Git，`git diff --check`通过。未安装新模型客户端、修改全局默认/认证、发起第二次probe或启动简单开发任务；后续任务的实际TUI配置与业务检查须由其当前owner核实。
