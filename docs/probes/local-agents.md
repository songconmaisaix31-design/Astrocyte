# 本地 Agent 与工具清点

## 2026-10-10 W2 原生续作（当前，源码d5ee4a1）

用户现已要求全套本地操作，人类选择固定项目/CLI并分别授权 A/B/C；旧“仅库存、待答”章节保留为历史证据。实现与验证见 [W2 当前报告](../../tasks/S1-sync-W2.md)。全部原生 CLI 仅协作限制、关闭工具，不保证进程不能读其自有全局配置；宿主不读取认证文件或自动接管既有会话。

| CLI | 已实现原生控制 | 独立真实结果 | 未验证/拒绝 |
|---|---|---|---|
| Codex0.162.0 | app-server JSON-RPC initialize/thread start+resume/turn send+interrupt/events/owned stop；完整context用thread/read校验ID/cwd后分页thread/items/list | 非推理握手与两真实公开marker turns，同原nativeID恢复且owned exit确认；实际gpt-6.1-sol/openai；完整context分页仅离线协议检查 | 用户历史根未给；外部session占用交接/reconcile未做；完整context安装协议验收NOT_RUN，停止进程不隐式重启 |
| Claude2.1.238 | 原生stream-json SDK控制initialize/interrupt、指定UUID start/同UUID resume、user消息、assistant/result事件、owned stop | 非推理握手与两真实公开marker turns、同UUID恢复、owned exit确认；实际qwen3.7-max/provider端点未核实 | 外部历史schema及pre-paid当前model观测缺失，generic selected-text明确EvidenceMissing；不将workspace成功扩展成distillation验收 |
| Pi1.0.1 | RPC get_state/prompt/abort/clear_queue、原sessionFile resume、events、owned stop | 非推理握手通过；真实首turn definite失败，单次授权诊断authentication_rejected、stop确认 | 原生real resume尚未执行，模型失败不得算PASS；不迁移认证、不重试 |
| OpenCode/Grok/Kimi/Qwen/Cursor | 固定安装库存 | 仅version/help证据 | registry无已验证driver，native API明确unsupported，能力不因安装而提升；不声称官方协议不存在 |

协议/公开成熟接口调查仅用于实现路线，不能替代安装版本与task-live：

- [Codex app-server官方文档](https://learn.chatgpt.com/docs/app-server)及[原生源码](https://github.com/openai/codex/tree/main/codex-rs/app-server)：复用JSON-RPC线程、turn、审批和事件；安装CLI公开生成schema确认thread `sandbox=read-only`，turn policy `type=readOnly`，两者不要混用。
- [Claude官方CLI参考](https://code.claude.com/docs/en/cli-reference)、[官方stream模式](https://code.claude.com/docs/en/agent-sdk/streaming-vs-single-mode)、[官方SDK query实现](https://github.com/anthropics/claude-agent-sdk-python/blob/main/src/claude_agent_sdk/_internal/query.py)：复用stream-json控制请求关联、SDK initialize、deny tool、原UUIDresume；仅调用现有native包，不安装seat或读取CLI认证。
- [Pi官方RPC/Session文档](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/docs)：对照本机公开包rpc.md/session-format.md复用get_state/prompt/abort/sessionFile；所有read/command、extensions、skills/templates关闭。
- [OpenCode官方server接口](https://opencode.ai/docs/server/)具有session create/message/prompt_async/abort和SSE；这是后续真实driver候选，尚未实现，本轮不连接任意现有server、不调用provider/auth或全局project/session清点。
- [coder/agentapi](https://github.com/coder/agentapi)的HTTP消息、状态、SSE展示了协作控制方式，但terminal稳定状态不能证明原生session ID恢复或OS隔离；未部署包装服务。[Happy](https://github.com/slopus/happy)和[Happy providers案例](https://github.com/slopus/happy-agent/blob/main/packages/happy-providers/EXAMPLES.md)用于参考原生session/中断接口，不另建远程同步、scheduler、Attempt/Manifest或证明系统。

Context packet是controller选择范围的固定版本封装，不是完整原生history；external_observed packet的created_at只是发现范围封装时间，native未知updated_at为null。正向stop需拥有的进程树退出，单个abort响应/kill请求不能冒充StopConfirmed；UNKNOWN投递不重放。Codex完整context需initialize协商experimentalApi，方法不支持时返回unsupported。确认未启动的resume失败保留原owned/nativeID及停止事实，仅新操作回执failed；不是旧会话unstarted。此前真实turn/resume结果不覆盖后加的完整context断言，新增原生断言及跨session context_handoff仍NOT_RUN。当前所有真实slot已归还，个人项目/历史根待人类提供。

后续W3独立 SOURCE `0d2319e`复用原会话的真实API验收：active完整context200同时含原user_text/assistant marker、same-ID resume/第三send/positive stop通过，保留首次错误URL matcher RED；不外推Pi或跨session handoff。首paper selected-text job约18秒UNKNOWN、Result=null，实际原因未保留，禁止重放。协调者随后指出流式delta计数缺陷，离线1000个小delta首次RED在765字节触及256事件；现只合并同一消息连续增量，不跨控制/消息边界，原128KiB与256事件上限仍强制。返修及针对回归通过、未新发模型，精确提交见W2 Handoff；不能把这个离线缺陷改写为首paper已确定原因。

### OpenCode1.18.35 精确范围阻塞（零native/model启动）

本机公开npm wrapper/package确认native `bin/opencode.exe`、版本1.18.35；未查私有设置。按同版本官方源核实：[Config合并](https://github.com/anomalyco/opencode/blob/v1.18.35/packages/opencode/src/config/config.ts)对instructions拼接数组，runtime `[]`无法清除继承输入；[Instruction系统读取](https://github.com/anomalyco/opencode/blob/v1.18.35/packages/opencode/src/session/instruction.ts)仍读取globalFiles与已配置绝对/远程instructions；[MCP初始化](https://github.com/anomalyco/opencode/blob/v1.18.35/packages/opencode/src/mcp/index.ts)对合并后enabled连接，无观察到pure全局抑制。不能仅凭tools deny把这些输入/连接称为selected-context范围。

[Plugin源](https://github.com/anomalyco/opencode/blob/v1.18.35/packages/opencode/src/plugin/index.ts)的OPENCODE_PURE抑制外部插件，内置native认证插件保留；[Provider.defaultModel](https://github.com/anomalyco/opencode/blob/v1.18.35/packages/opencode/src/provider/provider.ts)采用cfg.model或native recent state/default推导，/config/providers的排序目录default并非唯一当前model证据。/config返回完整配置可能带敏感provider options，本轮未请求，不打印或保存配置。root说明普通native runtime cache/dependency初始化属正常CLI授权；真正未解决的是未选择instructions/MCP范围与可信model元数据。保持unimplemented/unknown，未制造PTY包装或改全局配置；第二generic backend决定由root向人类确认。

## 2026-10-10 W2 库存探测（先前阶段）

基线 `20437708e43201e352d6c6926902e1363fd2ad3e`；开发工作树 `s1-local-agents-1010`。本轮只对明确 CLI 名称做 PATH 查找、固定 `--version` / `--help`，并读取安装包内公开协议文档；没有私有会话发现、配置/认证文件读取、整盘扫描、登录、模型请求、媒体转换或既有会话控制。`installed=available` 表示 PATH 入口存在，版本/帮助启动成功也不证明模型调用可用。

| Agent | 本轮版本（version/help 均 exit 0） | PATH 入口 | configured | startable | 八项原生能力 |
|---|---|---|---|---|---|
| Codex | 0.162.0 | npm `codex.cmd` | unknown | unknown | 全部 unknown / NOT_RUN |
| Claude Code | 2.1.238 | npm `claude.cmd` | unknown | unknown | 全部 unknown / NOT_RUN |
| OpenCode | 1.18.35 | npm `opencode.cmd` | unknown | unknown | 全部 unknown / NOT_RUN |
| Pi | 1.0.1 | npm `pi.cmd` | unknown | unknown | 全部 unknown / NOT_RUN |
| Grok | 1.0.34 (3736acbc8658) | `.grok/bin/grok.exe` | unknown | unknown | 全部 unknown / NOT_RUN |
| Kimi Code | 2.1.1 | `.kimi-code/bin/kimi.exe` | unknown | unknown | 全部 unknown / NOT_RUN |
| Qwen Code | 0.24.6 | npm `qwen.cmd` | unknown | unknown | 全部 unknown / NOT_RUN |
| Cursor | 3.23.23 (2dac2428994fe34f12658d9ecad1541b98db2c00, x64) | GUI `cursor.cmd` | unknown | unknown | 全部 unknown / NOT_RUN |

额外固定名称查找：`gemini`、`cursor-agent` 未在 PATH 找到；库存记录 `installed=unavailable, reason=not_found_on_path`，不声称未安装在其他目录。PATH 中 `agent.exe --version/--help` 返回 Grok 1.0.34，是 Grok 别名，不作为独立 Cursor Agent。Cursor GUI help 没有原生 headless 会话入口，仅凭此不能证明本机所有 Cursor Agent 能力均 unsupported。

八项为 `discover/read_context/start/resume/send/stop/observe/reconcile`，状态词固定 `supported/unsupported/unknown`。库存安装发现与原生会话 discover 是不同事实；安装检查不把 native discover 标记 supported。configured 保持 `private_configuration_not_inspected`，startable 保持 `native_start_not_authorized_or_tested`。能力 unknown 不开放控制按钮；unsupported 仅用于实际确认不支持的接口，当前没有这样的原生运行时结果。

### 实际命令与首个失败

- `Get-Command codex,claude,opencode,pi,grok,kimi,qwen,cursor,gemini,cursor-agent,agent -ErrorAction SilentlyContinue`。
- Python `shutil.which(name+'.cmd') or shutil.which(name)` 定位预期命令；原生 EXE 用 `subprocess.run([path, flag], timeout=25)`，CMD 用 `subprocess.run(['cmd.exe','/d','/c',path,flag], timeout=25)`，仅固定 version/help。
- 首次 Python CMD 参数额外套引号，6 个包装 CLI 返回 exit 1（命令未被识别）；Grok/Kimi 原生 EXE 成功。修正后 Codex version 成功，但打印 help 遇 `UnicodeEncodeError: 'gbk' codec can't encode character U+2011`，探测脚本 exit 1；设置 `sys.stdout.reconfigure(encoding='utf-8')` 后，上表全部 version/help exit 0。首个错误是探测工具问题，未改写为原生 Agent 成功或失败。
- `codex app-server --help` exit 0；确认本机 `stdio://` 默认、`--stdio`、实验 daemon/proxy/schema 入口。没有执行 daemon、proxy、thread list 或任何原生 RPC。
- `ASTROCYTE_TEST_LOCAL_AGENT_CLI=1 go test ./internal/adapters/agents -run '^TestInventoryLiveCLI$' -count=1 -v`：本机适配器实际 version/help PASS，8 个版本可解析，2 个额外入口未找到。这里只测 nonsecret CLI，不测模型或原生会话。

### 可直接复用的原生接口与前置条件

| 路线 | 已核实证据 | 后续用途 | 必须先满足 |
|---|---|---|---|
| Codex app-server | 0.162.0 本机 help；[官方接口文档](https://learn.chatgpt.com/docs/app-server) | stdio JSON-RPC；initialize/initialized 握手；thread/start、thread/resume、turn/start、turn/steer；通知流和 turn/interrupt；带 cwd 的 thread/list | 明确项目/读取根、会话 ID 与所有权交接、批准工具与外发规则；用户允许会话操作且协调者安排一次运行；保存本机版本对应 schema，不由最新在线文档反推本机支持 |
| Pi RPC | 1.0.1 包元数据、包内 `docs/rpc.md` / `rpc-commands.md`；[上游 RPC](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/rpc.md)、[命令参考](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/rpc-commands.md) | `--mode rpc` JSONL，按 id 关联响应；get_state、prompt/steer、switch_session；事件流与关闭 stdin | 显式批准的隔离 session-dir、工具/读取根、provider 配置可用性、现有会话占用交接；不能用默认 session 目录做全量发现 |

Codex 文档中的 `turn/interrupt` 取消当前 turn，必须观察最终事件，不能把 `{}` 回执等同所有外部工具已停或所有客户端退出。Pi prompt 的 success 是接收/排队状态，需观察 `agent_settled`；`switch_session` 可能被扩展取消；Pi `abort` 后仍可能处理排队消息，停止新工作应先 `clear_queue`，并确认流结束。Pi 包内文档与上游一致，但本轮没有启动 RPC，故能力仍 unknown。这里提供原生薄适配路线，不新增调度器或把旧日志读取当 resume。

其余帮助线索：Claude 支持 print/stream-json、continue/resume 和后台 agents；OpenCode 有 acp/serve/attach；Grok 有 agent 与 sessions；Kimi 有 acp、session 与 stream-json；Qwen 有 acp、serve 与 sessions。全部只是候选入口，不能作为已配置、可启动或八项能力已验收的证据。

### 已实施边界与待定范围

`agents.NewInventory()` 只进行 PATH 查找；`RefreshCLI(ctx)` 是显式版本/帮助探测，每次子命令最多 5 秒，输出保留最多 64KiB 并持续排空。Windows CMD 路径经过元字符拒绝、使用明确 CmdLine，原生子进程归属本次 Job Object，超时关闭本次子进程树；Unix 使用所属进程组。只返回结构化原因与解析后的版本，不向 API 暴露原始帮助、诊断、可执行路径或私有状态。`Snapshot` 只读缓存并复制可变字段，不启动 CLI。`workspaceapp.NewLocalAgentService(provider)` 通过 W0 已发布端口读取缓存；HTTP/组装由 W0 拥有。建议启动总期限 10–15 秒；上下文取消保留部分结果，未运行项标记 `cli_entry_found_probe_not_run`，只完成版本项标记 `cli_entry_found_probe_incomplete`，两探测都通过才标记 `cli_entry_found_version_help_passed`；配置和原生能力始终 unknown。

测试实际拒绝 exec/resume/sessions/doctor 等非 version/help 参数和 shell 元字符路径，并核实拥有的临时测试进程超时退出、Windows 空格路径包装与后代清理；不是伪造原生会话。配置、历史、全盘目录和 Agent 模型没有测试授权，故没有探测。用户尚未选择“完整原生操作”或“先库存”，Agent 可读项目根也未定；本轮不冻结项目/活动绑定契约或 SQLite 迁移，不赋予目录或资料权限，不创建 Mission/Swarm。项目 @材料引用现有 Attention 语义继续保留，热度不授权。

## 2026-10-09 W3 清点（历史，以下原记录保留）

日期：2026-10-09。只读探测：`Get-Command` + `--version` / `--help`（带超时）。未读取密钥、会话记录或私有文件，未扫描整盘，未安装/升级任何 Agent，未运行任何付费能力会话或真实 summarize。

> 约定：本文只记录「可调用入口 + 版本 + `--help` 表面提示」。`--version`/`--help` 只证明 CLI 已安装、可调用，**不能**证明真实会话能力。SPEC 8.1 的 8 项运行时能力（discover/read_context/start/resume/send/stop/observe/reconcile）对全部 8 个 Agent 一律为 `unknown` / NOT_RUN，待真实会话测试后更新。

## 1. 平台依赖基线（S0 验收相关，实际命令结果）

| 工具 | 版本 | 路径 |
|---|---|---|
| Node | v24.16.0 | `C:\Program Files\nodejs\node.exe` |
| pnpm | 11.27.0 | `C:\Users\DW\.local\bin\pnpm.ps1` |
| Go | go1.27.2 windows/amd64 | `C:\Users\DW\AppData\Local\Programs\go\bin\go.exe` |

> STATUS.md 记「Go 正在安装便携工具链」，现已完成并进入 PATH（go1.27.2）。

## 2. Agent CLI（8 个，已安装、可调用、版本已核实）

| Agent | 版本 | 入口 |
|---|---|---|
| Codex | 0.160.0 | `codex.ps1`（npm 全局） |
| Claude Code | 2.1.238 | `claude.ps1`（npm 全局） |
| OpenCode | 1.18.35 | `opencode.ps1`（npm 全局） |
| Pi | 1.0.1 | `pi.ps1`（npm 全局） |
| Grok | 1.0.34 (3736acbc8658) | `C:\Users\DW\.grok\bin\grok.exe` |
| Kimi Code | 24.15.0 | `C:\Users\DW\.kimi-code\bin\kimi.EXE` |
| Qwen Code | 0.24.6 | `qwen.ps1`（npm 全局） |
| Cursor | 3.23.23 (2dac2428…) | `cursor.cmd`（GUI 应用 CLI 包装） |

## 3. `--help` 提示矩阵（仅帮助文本表面线索，非能力验收）

| 能力 | Codex | Claude | OpenCode | Pi | Grok | Kimi | Qwen | Cursor |
|---|---|---|---|---|---|---|---|---|
| discover | `agents` | `doctor`/`agents --json` | `session list`/`attach` | `list`/`--list-models` | `sessions list/search`/`dashboard` | `session list`/`doctor` | `sessions list/ps` | 未见于帮助 |
| read_context | exec `--json`/`--output-last-message` | `-p --output-format stream-json` | `export [id] --sanitize` | `--export <html>` | `export <id>` (md) | `export [id] -o zip` | `--output-format json/stream-json` | 未见于帮助 |
| start | `[p]`/`exec`/`--worktree`/`-C` | `[p]`/`-p`/`--bg`/`-w` | `run [msg]`/`[project]` | `[msg]`/`-p` | `[p]`/`-p` | `[p]`/`-p` | `[query]`/`-p`/`-i` | 未见于帮助 |
| resume | `resume [id]`/`--last` | `-r [id]`/`-c`/`--session-id`/`--fork-session` | `-c`/`-s [id]`/`--fork` | `--continue`/`--resume`/`--fork` | `-c`/`-r [id]`/`--fork-session` | `-S [id]`/`-c` | `-c`/`-r [id]`/`--fork-session` | 未见于帮助 |
| send | `queue --thread --message` | `--bg` + `agents` | `serve` + `attach` | `--mode rpc` | `agent` | `web`/`rc` | `serve` / `--input-file` | 未见于帮助 |
| stop | 未见于帮助 | `agents`（后台管理） | 未见于帮助 | 未见于帮助 | `leader kill` | 未见于帮助 | serve 会话回收 | 未见于帮助 |
| observe | `agents`/`app-server daemon` | `agents --json` | `serve`/`attach`/`stats` | `--export`/`list` | `leader list/info`/`usage` | `vis [id]`/`session list` | `serve`(SSE)/`sessions ps` | 未见于帮助 |
| reconcile | 未见于帮助 | 未见于帮助 | 未见于帮助 | 未见于帮助 | 未见于帮助 | 未见于帮助 | 未见于帮助 | 未见于帮助 |

> 上表全部为 `--help` 中出现的子命令/参数提示；不代表能力已验证。8 项运行时能力对所有 Agent 保持 `unknown`/NOT_RUN。

### 接入提示（nonsecret，摘自 `--help`）

- **Codex**：`exec [PROMPT] --json -o out.txt`；`resume <SESSION_ID> --last`；`queue --thread <UUID> --message <text>`；`app-server daemon`（实验）；`-C <dir>`、`--worktree`。
- **Claude Code**：`-p "…" --output-format stream-json`；`--bg` + `agents --json`；`-r <id>`/`-c`；`--session-id`/`--fork-session`；`--remote-control [name]`。
- **OpenCode**：`run`；`session list`；`export <id> --sanitize`；`serve`/`attach`；`acp`；`stats`。
- **Pi**：`-p`；`--continue`/`--resume`；`--export <file>`（HTML）；`--mode rpc`；`--tools read,grep,find,ls` 只读。
- **Grok**：`-p "…" --output-format json|streaming-json`；`-r [id]`/`-c`；`sessions list`；`export <SESSION_ID> [OUT]`（Markdown）；`leader list|kill`；`agent stdio|serve|headless`。
- **Kimi**：`-p "…" --output-format stream-json`；`-S [id]`/`-c`；`session list`；`export [id] -o out.zip`；`acp`；`web`/`rc`。
- **Qwen**：`-p`/`--acp`；`-r [id]`/`-c`/`--fork-session`；`sessions list/ps`；`serve --port 4170`（HTTP daemon：`POST /session`、`GET /session/:id/events` SSE、`--input-file` JSONL）；`--approval-mode plan|default|auto-edit|auto|yolo`。
- **Cursor**：仅 GUI 应用 CLI 包装，`--version` 可用；未见 headless 会话/接续子命令。

### ACP 观察（仅本次 `--help` 表面）

- 帮助中**明确出现** `acp` 子命令/`--acp` 标志：OpenCode（`acp`）、Kimi（`acp`）、Qwen（`--acp`）。
- 其余 Agent 未在本帮助中观察到 ACP 入口（Codex 为 `app-server`/`remote-control` 实验通道；Claude 为 `mcp`/`--bg`；Pi 为 `--mode rpc`；Grok 有 `--output-format streaming-json` 与 `agent stdio/serve`，但 streaming-json 仅说明输出格式，**不等于** ACP）。未实测定为非 ACP，仅记「本帮助未观察」。

## 4. summarize 导出提示（视频导入方向，P10）

| 工具 | 版本 | 导出/接入提示 |
|---|---|---|
| summarize | 0.21.8 | `--extract --format md --markdown-mode llm` 导出结构化转写；`--youtube auto|web|yt-dlp|apify`；`--diarize`；`--slides`；`--json`；`--model cli/<provider>`（经 `CLAUDE_PATH`/`CODEX_PATH` 等定位本地 Agent） |

> 未运行真实 summarize；真实样本格式未选定。这是「视频导入」方向最接近现有 S1 适配输入的 CLI。

## 5. 结论与限制

- 已核实 8 个 Agent 与 summarize 的版本和入口；SPEC 1.2 要求的优先适配器 Codex、Claude Code 均具备 CLI。
- 全部矩阵内容为 `--help` 表面提示；8 项运行时能力对 8 个 Agent 一律 `unknown`/NOT_RUN，`stop`/`reconcile` 普遍未在帮助中暴露。版本/帮助不得当作能力验收。
