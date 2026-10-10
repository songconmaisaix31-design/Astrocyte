# 本地 Agent 与工具清点

## 2026-10-11 OpenCode 1.18.35 loopback Basic-auth 会话驱动（隔离配置，正式接入）

本机 npm `opencode-ai/bin/opencode.exe` 仍为 `1.18.35`，`--version`/`--help` exit 0。按用户要求不再以「等下一版本」维持 `unsupported`，改为落地正式 loopback HTTP 会话驱动。关键纠正来自同版本官方源码：

- **`OPENCODE_CONFIG_DIR` 真正隔离用户全局 config**：`@opencode-ai/core/global.ts` 的 `make().config = Flag.OPENCODE_CONFIG_DIR ?? Path.config`，`Path.config = xdgConfig/opencode`；`session/instruction.ts` 的 `globalFiles = [path.join(global.config,"AGENTS.md"), ...]` 读的是 `global.config`，**不是**不可变用户目录。因此把 `OPENCODE_CONFIG_DIR` 指向控制器自有的空目录，用户全局 `AGENTS.md`、MCP、plugins、instructions 不加载（`~/.claude/CLAUDE.md` 由 `OPENCODE_DISABLE_CLAUDE_CODE_PROMPT` 关闭、项目层由 `OPENCODE_DISABLE_PROJECT_CONFIG` 关闭）。前轮「全局 instructions 不可关闭」的结论因此不成立，已作废。**注意**：这不等同于全机隔离——`config.ts` 的 `ConfigManaged.managedConfigDir()`（Windows `%ProgramData%\opencode`）与 auth wellknown/account 远程 config 仍会被加载，可能声明 MCP/plugins；本机 `%ProgramData%\opencode` 不存在，且 `permission:deny` 仍阻断一切工具使用，不据此宣称这些来源已被压制。
- **deny-by-default 落地（最终 overlay）**：自有空 config 目录写 `opencode.json` 显式逐键 `permission:deny`（read/edit/bash/glob/grep/webfetch/websearch/task/skill/lsp/question/external_directory/doom_loop 全 deny），并设 `OPENCODE_PERMISSION` 相同对象，`config.ts` 在 wellknown/account/managed 之后 `mergeDeep` 它作为最终覆盖；不传 `--auto`。**如实限制**：仅 `{"*":"deny"}` 通配不能清除继承的 specific read/bash allow，故逐键 deny；但 agent-specific permission（agent rules 优先）与 managed/wellknown 注入的 instructions/MCP/plugins 仍不保证被压制，本 driver 不宣称这些残留已隔离，fail-closed 保持 `Configured=unknown`。`--pure` + `OPENCODE_DISABLE_DEFAULT_PLUGINS=1` 关闭外部/默认插件（`--pure` 本身不保证 MCP/plugins 不初始化）。
- **auth 由 CLI 原生读取**：不复制/读取 provider 凭据；`OPENCODE_SERVER_USERNAME`/`OPENCODE_SERVER_PASSWORD` 每次会话随机生成，仅供本 loopback 服务 Basic auth，不落库、不打印。
- **正式接口面**：`opencode serve --hostname 127.0.0.1 --port <N>`，`POST /session`（返回会话 id 作 native ID）、`POST /session/:id/message`（同步等待，`parts[]` 返回文本）、`GET /session/:id/message`（读会话）、`GET /session/status`、`GET /global/health`、`POST /session/:id/abort`、`POST /instance/dispose`。driver 实现 start/send/read_context/resume/stop/observe；`discover` 无已验历史头作用域协议，保持 `unsupported`；`reconcile`/`context_handoff` 沿用应用层既有 NativeOperation/模式语义，不新增方法。

已交付 `internal/adapters/agents/opencode.go`：`opencodeNative` 独立会话驱动（与 stdio `*Native` 并列注册），owned subprocess + Job Object 正向退出、未知投递不重放、**同 nativeID 恢复**（resume 经 `GET /session/:id` 加载原 SID 续发，不新建会话）与**新 nativeID 交接**分开；模型按真实 `AssistantMessage` 的 `modelID`/`providerID` 解析；`Registry` 现 `List()` 为 `codex/pi/claude/opencode`。httptest 假 server 覆盖 create/send/read/verify 协议与 Basic auth、modelID/providerID、隔离 env、`permission:deny` 配置与注册快照。

**真实 `opencode serve` 会话链已运行通过**（`ASTROCYTE_TEST_OPENCODE_LIVE=1`，`opencode_live_test.go`）：`opencode serve --hostname 127.0.0.1 --port <N> --pure` 隔离启动；显式 `Send` 正向 turn 回显 selected marker；`read_context` 含 user/assistant 双向；`model=deepseek-flash provider=deepseek` 从真实 reply 解析；同 nativeID resume 保持原 ID；owned stop 正向退出；**跨会话 context_handoff**：新 nativeID + 携带原 selected context marker 并被新会话引用（非仅新建会话）。首次失败保留：一轮旧 `waitHealth` 阻塞 180s（客户端超时）导致 serve 启动竞态，已改为每请求 3s/整体 20s 有界超时，非改写为成功。旧成功媒体/模型与 UNKNOWN 不重发。selected-text/distillation 路径仍依赖 `*Native` 的模型观测（`Configured=unknown`），不冒充模型就绪。

## 2026-10-10 OpenCode 1.18.35 正式接口复核（serve/config/permissions + 本机版本源码）——结论已被上节更正

本机 npm 原生 `bin/opencode.exe` 仍为 `1.18.35`，`--version`/`--help` 均 exit 0。按用户要求重查官方 [Server](https://docs.opencode.ai/docs/server/)、[Config](https://docs.opencode.ai/docs/config/)、[Permissions](https://docs.opencode.ai/docs/permissions/) 与同版本源码（`v1.18.35` 的 `config.ts`、`session/instruction.ts`、`mcp/index.ts`）。结论未变，且本次补充了可执行侧的证据：

- **正式原生接口存在**：`opencode serve --hostname 127.0.0.1 --port <N>` 起一个 headless HTTP 服务，暴露 OpenAPI 3.1（`/doc`）；会话 `POST/GET /session[/:id]`、`POST /session/:id/message`（含 `noReply`/`system`/`tools` 字段）、`/prompt_async`、`/abort`、`/permissions/:permissionID`（对权限请求回 `response`）与 `/instance/dispose`。可用 `OPENCODE_SERVER_PASSWORD`/`OPENCODE_SERVER_USERNAME` 做 loopback HTTP Basic Auth，属“HTTP loopback owned server + auth + provided prompt”的可行协议面。
- **deny-by-default 工具可达成**：`permission` 配置（`v1.1.1` 起合并旧 `tools` 布尔），且 `Flag.OPENCODE_PERMISSION` 用 `mergeDeep` 按进程注入权限规则；不传 `--auto`，并对每个权限请求回 `deny` 即可把 read/edit/bash/webfetch/mcp 等全部拦下。这是按进程环境变量，不是覆盖用户配置。
- **无法安全关闭全局 instructions**：`config.ts` 的 `mergeConfigConcatArrays` 对 `instructions` 做**数组拼接 + Set 去重**，晚到的 `instructions: []` 不会清掉已继承项；`instruction.ts` 的 `globalFiles` 恒加载 `~/.config/opencode/AGENTS.md` 与 `~/.claude/CLAUDE.md`（仅 `disableClaudeCodePrompt` 关后者，AGENTS.md 无开关），并按 `findUp` 自动读项目 `AGENTS.md/CLAUDE.md/CONTEXT.md`，再叠加拼接后的 `config.instructions`。`OPENCODE_DISABLE_PROJECT_CONFIG` 只关项目层发现，不关全局。
- **无法全局关闭 MCP**：`mcp/index.ts` 遍历合并后的 `cfg.mcp`，对每个 `enabled !== false` 的条目立即 `create`/`connect`（stdio spawn 或远程 HTTP/SSE），无观察到 pure 级全局抑制；按 key 设 `enabled:false` 需知道并覆盖用户所有 key，而覆盖用户配置被禁止。

因此 OpenCode 1.18.35 的 selected-context 边界（只交付获准正文、不泄漏全局指令、不连接用户 MCP）**仍不能在不覆盖用户 CLI 配置的前提下成立**：仅凭 tools/permission deny 不能把仍在加载的全局 instructions 与已连接的 MCP 称为 selected-context 范围。原生适配对 OpenCode 保持 `unsupported`，安装/版本/help 仍分别记录为 `installed`/`version_help_passed`；不制造 PTY/daemon 包装、不改全局配置、不读取 provider 认证。若后续版本提供干净的 instructions/MCP 抑制开关，即可沿上述 HTTP 服务协议落地原生 driver。不把“正式接口存在”冒充“已接入”。

## 2026-10-10 项目汇总、会话元数据与独立人工记录

本轮已授权本地项目/历史识别。沿 `RegisteredProjects`：Orca 公共登记根/工作树 + Codex、Pi、Claude 的明确原生元数据目录；CLI 安装、配置和执行权限仍独立。新增 CLI 元数据指出的目录只核 canonical 绝对路径及 `.git`/`go.mod`/`package.json`/`pyproject.toml`/`Cargo.toml`/`CMakeLists.txt` 的文件标记，不读取内容、不递归搜索父目录，不登记执行项目。排除重定向路径、home 根、系统安装目录及凭据/原生状态目录。

Codex 只读 `CODEX_HOME/sessions` 或 `~/.codex/sessions` 日期目录的首条 `session_meta` 中 id/cwd/timestamp；Pi 只读 `PI_CODING_AGENT_DIR/sessions` 或 `~/.pi/agent/sessions` 一层项目目录的首条 session 元数据。每源最多 2048 目录项、64KiB 首记录；不读 SQLite/认证/全文对话。字段依据 [Codex 官方 rollout 接口](https://github.com/openai/codex/blob/main/codex-rs/core/src/rollout.rs) 和 [Pi 官方 SessionHeader](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/src/core/session-manager.ts)。

Claude 路径来自 [官方会话文档](https://code.claude.com/docs/en/sessions#where-transcripts-are-stored)及 [官方 SDK 元数据读取实现](https://github.com/anthropics/claude-agent-sdk-python/blob/main/src/claude_agent_sdk/_internal/sessions.py)。本机 8 个项目目录没有 `sessions-index.json`，因此使用 `CLAUDE_CONFIG_DIR/projects` 或 `~/.claude/projects`：每文件总共最多64KiB、16条初始记录，仅解析顶层 cwd/sessionId/timestamp，舍弃 message/prompt/tool/summary；不调用 SDK 的摘要/首提示抽取、不读取尾部。异常或内部格式变化保持 partial。

目录由 Orca 登记身份或实际 Git common-dir 聚合；名称、相同提交或创建客户端不是身份/贡献者证据。独立克隆及嵌套项目保持独立。原生 timestamp 记录为 `created_at`，不作为最新活动；Git 最后提交、Orca 工作树活动和原生历史关联分别保留。每根/客户端最多保留4条会话关联样本、每源最多256条关联；`entries_examined`、`headers_examined`、`matched_headers`、`matched_roots`、`retained_associations` 分开，均不是完成率或全量会话数。

本轮实际有界采集曾观察111目录→71项目组，含57个原生元数据独立根、4个多目录组、5个 Codex+Claude 关联组；保留125条关联样本，来源仍 partial。Pi 已知源不存在；OpenCode/Grok/Kimi/Qwen/Cursor/Gemini/Cursor Agent 尚无可靠元数据适配，逐客户端 unknown，不从安装情况推导项目覆盖。

notes/review/group/intent/archived 保存于独立 `local_project_metadata` 表，人工身份/CSRF + revision CAS 写入；updated_at 由服务生成。反复刷新保留人工字段，归档可逆，不删除代码、不改变 grants。GET 只读持久缓存、在内存聚合，不扫描或写入。人工阶段/Agent推断决定尚待答复，intent 为人工自由文字。具体验证、首 RED 和剩余验收见 [W2 报告](../../tasks/S1-paper-W2.md)。

原生 selected-text 总输入上限512KiB UTF-8，包括 `packetPrompt` 固定指令和上下文信封；空/非法 UTF-8/超限在 ConfigurationID 和原生启动前拒绝，无截断。`ValidateSelectedTextPrompt` 和 `SelectedTextPromptBytes` 使用同一信封。实际发送前只记录输入字节数/输入上限/输出上限，不记录内容或配置。输出上限和原生历史128KiB保持独立；已有项目上下文/普通消息/权限限制未随之放宽。

## 2026-10-10 公开 GitHub 元数据与人工蜂群纳入

本轮用户已确认“先信息、人工纳入顶层蜂群空间才克隆”。复用 Go HTTP、SQLite、`LocalProject` 与既有 `ProjectSpace` 桥；`GET /github-repositories` 只查缓存，人工同步接受公开账号名、`owner/repo` 或 `https://github.com/...`。按 [GitHub Repository REST API](https://docs.github.com/en/rest/repos/repos) 匿名查询，账号仅第一页最多100项，不声称完整账号库存；不读认证、不查私有仓库、不在元数据阶段创建代码目录。固定稳定 GitHub ID 去重，观察 revision 与 metadata_revision 分开，失败保留旧数据/时间并标 stale。

仓库输入遇到明确匿名限流（403 且 `X-RateLimit-Remaining: 0`）时，至多再读取同一官方公开仓库 HTML 页。必须同时核验稳定 repository_id、repository_nwo 与 repository_public=true，DOM变化/缺证据保持失败；账号、认证403和重定向不走回退。`metadata_source` 区分 `github_rest`、`github_public_html`、`unknown`；HTML未观察的 description/default_branch/language/stars/archived/updated_at/pushed_at 在 `unknown_fields` 明确列出，不能把0/false/零时间显示为真实事实。REST缺失/null字段同样显式未知，必须给出private=false证明；没有 provenance 的旧缓存不默认变成 REST。

人工+CSRF placement 经入口桥校验已存在的人选顶层空间后，固定程序执行 [Git clone](https://git-scm.com/docs/git-clone) 的 shallow/single-branch/no-tags/no-recurse-submodules/empty-template，仅 canonical public HTTPS；不执行源码、hooks、脚本、LFS、子模块或推送，不用凭据 helper。继承 Git/SSH 重定向清除，global/system Git config 禁用，HOME指向自有临时目录。应用管理根内新 staging，Windows MoveFile/Linux RENAME_NOREPLACE 原子发布，任何已有目标都不覆盖；已有完整 checkout 仅核本地 origin/HEAD，拒绝 config include/外部存储/符号链接。其他平台原子发布明确不可用，不静默降级覆盖。

实际 root/HEAD 持久化后才登记 A 默认的 LocalProject；B/C、模型、control/actions/grants 仍独立，人类重复 placement 不改权限、不启动 Agent/Mission。clone最长90秒、整场120秒含排队、最多3次人工尝试；GET/冷启动不自动重试 cloning/interrupted，人工明确恢复可复用已发布 checkout。失败 staging保留，未覆盖或递归删除；耗尽后需人工检查。没有 pull/update 或修改既有代码目录。

本轮实际匿名 REST→SQLite→clone→冷重启/重复纳入 PASS，公开上游 `steipete/summarize` 的实际ID1118209243、HEAD `560197cd4b580554cccf648744c592e867b43bb5`。首403、单独Git验证与后续REST200分别保留；HTML回退的目标测试不冒充真实REST场。详细命令、原始日志和限制见 [W2 当前报告](../../tasks/S1-sync-W2.md)。下文“GitHub尚未实现”为前轮历史。

## 2026-10-10 登记项目发现与 GitHub 调查

最新授权见 QUESTIONS 顶部：可读本地项目，先消费 `orca repo list --json` 的登记根与 `orca worktree list --limit 256 --json` 的本机登记工作树。固定程序只读取公开登记元数据和选定根的目录标记；不调用模型，不读取 Orca 私有库、认证原件、客户端历史正文或含认证的 remote。不把登记许可扩成 Agent B/C、写入、启动或外发许可。缺失根记录失败；主根缺失不抹掉独立登记且可访问的本机工作树，不搜索替代目录。GET 只读持久化缓存，冷重启可查；整个来源失联时保留旧时间和数据，状态 `stale`。

Git 观察复用原生程序，固定 `rev-parse HEAD`、`symbolic-ref`、单条提交时间和 tracked-only porcelain status。[Git status 官方文档](https://git-scm.com/docs/git-status)描述机器可读格式及后台刷新；使用[Git 全局选项](https://git-scm.com/docs/git) `--no-optional-locks`、`GIT_OPTIONAL_LOCKS=0`，关闭 fsmonitor、hooks、untrackedCache、签名显示和子模块递归，清除继承的 `GIT_*` 路由/trace 环境。无 fetch、checkout、clone、凭据请求或补 safe.directory。所有命令使用固定二进制 argv 与既有 owned process helper，10秒/命令、120秒/场，Orca输出2MiB/Git64KiB、256候选、每登记根目录深度3/2000项；超限、取消和失联保留未知。仅检查已跟踪文件是否改变，不展示 diff/文件列表、不声称未跟踪文件也已检查。

活动只来自 Orca `lastActivityAt` / `createdWithAgent`；创建客户端不是正在运行的客户端，活动不是原生会话历史，Git提交时间单列。不用 addedAt、mtime、安装库存推断 AI 最近修改。首次真实观测的明确 partial 与冷重启/身份测试见 W2 报告。

GitHub 仅完成官方可复用能力调查，用户未确定同步方式、首批账号/仓库和私有权限，当前未新增 API 或下载入口。[Repository API](https://docs.github.com/en/rest/repos/repos)、[README/contents API](https://docs.github.com/en/rest/repos/contents)及[commits API](https://docs.github.com/en/rest/commits/commits)允许公开资源匿名查询；若用户选信息同步，可沿现有 Go HTTP 与 SQLite 有界快照，而无需让 Agent 执行同步。固定程序应遵循[官方最佳实践](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api)条件请求、速率限制及有界分页；文档返回的仓库文字仍不成为命令授权。[GitHub 入门文档](https://docs.github.com/en/rest/using-the-rest-api/getting-started-with-the-rest-api)说明 `gh api` 依赖登录，因此不读取现有 token 来冒充匿名。若用户选 clone/update，需要另确认写入目标、分支、冲突及凭据使用方式；不将普通只读发现扩成 fetch/pull 或全账号下载。

## 2026-10-10 W2 原生续作（当前，源码a797324）

用户现已要求全套本地操作，人类选择固定项目/CLI并分别授权 A/B/C；旧“仅库存、待答”章节保留为历史证据。实现与验证见 [W2 当前报告](../../tasks/S1-sync-W2.md)。全部原生 CLI 仅协作限制、关闭工具，不保证进程不能读其自有全局配置；宿主不读取认证文件或自动接管既有会话。

| CLI | 已实现原生控制 | 独立真实结果 | 未验证/拒绝 |
|---|---|---|---|
| Codex0.162.0 | app-server JSON-RPC initialize/thread start+resume/turn send+interrupt/events/owned stop；完整context用thread/read校验ID/cwd后分页thread/items/list | 非推理握手与两真实公开marker turns，同原nativeID恢复且owned exit确认；实际gpt-6.1-sol/openai；后续W3真实同线程完整context200及same-ID恢复/第三send/stop通过 | 用户历史根未给；外部session占用交接/reconcile未做；新增opt-in完整context断言自身尚未运行，停止进程不隐式重启 |
| Claude2.1.238 | 原生stream-json SDK控制initialize/interrupt、指定UUID start/同UUID resume、user消息、assistant/result事件、owned stop | 非推理握手与两真实公开marker turns、同UUID恢复、owned exit确认；实际qwen3.7-max/provider端点未核实 | 外部历史schema及pre-paid当前model观测缺失，generic selected-text明确EvidenceMissing；不将workspace成功扩展成distillation验收 |
| Pi1.0.1 | RPC get_state/prompt/abort/clear_queue、原sessionFile resume、events、owned stop | 非推理握手通过；真实首turn definite失败，单次授权诊断authentication_rejected、stop确认 | 原生real resume尚未执行，模型失败不得算PASS；不迁移认证、不重试 |
| OpenCode1.18.35 | loopback Basic-auth serve 会话 driver（start/send/read_context/resume/stop/observe + context_handoff，配置隔离 `OPENCODE_CONFIG_DIR`+逐键 `permission:deny`+`OPENCODE_PERMISSION`+`--pure`） | httptest 假 server 协议/隔离 env/注册快照通过；真实会话链（显式 Send/ReadContext/同 ID resume/context_handoff 新 ID 携带旧上下文/owned stop，model=deepseek-flash）通过 | discover 无已验历史头协议 unsupported；selected-text 需 `*Native` 模型观测，OpenCode `Configured=unknown`；agent-specific permission 与 managed/wellknown 残留不保证压制，不冒充模型就绪/全机隔离 |
| Grok/Kimi/Qwen/Cursor | 固定安装库存 | 仅version/help证据 | registry无已验证driver，native API明确unsupported，能力不因安装而提升；不声称官方协议不存在 |

协议/公开成熟接口调查仅用于实现路线，不能替代安装版本与task-live：

- [Codex app-server官方文档](https://learn.chatgpt.com/docs/app-server)及[原生源码](https://github.com/openai/codex/tree/main/codex-rs/app-server)：复用JSON-RPC线程、turn、审批和事件；安装CLI公开生成schema确认thread `sandbox=read-only`，turn policy `type=readOnly`，两者不要混用。
- [Claude官方CLI参考](https://code.claude.com/docs/en/cli-reference)、[官方stream模式](https://code.claude.com/docs/en/agent-sdk/streaming-vs-single-mode)、[官方SDK query实现](https://github.com/anthropics/claude-agent-sdk-python/blob/main/src/claude_agent_sdk/_internal/query.py)：复用stream-json控制请求关联、SDK initialize、deny tool、原UUIDresume；仅调用现有native包，不安装seat或读取CLI认证。
- [Pi官方RPC/Session文档](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/docs)：对照本机公开包rpc.md/session-format.md复用get_state/prompt/abort/sessionFile；所有read/command、extensions、skills/templates关闭。
- [OpenCode官方server接口](https://opencode.ai/docs/server/)具有session create/message/prompt_async/abort和SSE；这是后续真实driver候选，尚未实现，本轮不连接任意现有server、不调用provider/auth或全局project/session清点。
- [coder/agentapi](https://github.com/coder/agentapi)的HTTP消息、状态、SSE展示了协作控制方式，但terminal稳定状态不能证明原生session ID恢复或OS隔离；未部署包装服务。[Happy](https://github.com/slopus/happy)和[Happy providers案例](https://github.com/slopus/happy-agent/blob/main/packages/happy-providers/EXAMPLES.md)用于参考原生session/中断接口，不另建远程同步、scheduler、Attempt/Manifest或证明系统。

Context packet是controller选择范围的固定版本封装，不是完整原生history；external_observed packet的created_at只是发现范围封装时间，native未知updated_at为null。正向stop需拥有的进程树退出，单个abort响应/kill请求不能冒充StopConfirmed；UNKNOWN投递不重放。Codex完整context需initialize协商experimentalApi，方法不支持时返回unsupported。确认未启动的resume失败保留原owned/nativeID及停止事实，仅新操作回执failed；不是旧会话unstarted。此前真实turn/resume结果不覆盖后加的完整context断言，新增原生断言及跨session context_handoff仍NOT_RUN。当前所有真实slot已归还，个人项目/历史根待人类提供。

后续W3独立 SOURCE `0d2319e`复用原会话的真实API验收：active完整context200同时含原user_text/assistant marker、same-ID resume/第三send/positive stop通过，保留首次错误URL matcher RED；不外推Pi或跨session handoff。首paper selected-text job约18秒UNKNOWN、Result=null，实际原因未保留，禁止重放。协调者随后指出流式delta计数缺陷，离线1000个小delta首次RED在765字节触及256事件；现只合并同一消息连续增量，不跨控制/消息边界，原128KiB与256事件上限仍强制。返修及针对回归通过、未新发模型，精确提交见W2 Handoff；不能把这个离线缺陷改写为首paper已确定原因。

资料流程后续冻结应用SOURCE `cd721692122c5e8f7820e34f25181058dd648cf0`/W3 REPORT `56e9c856`：四独立新整理目标真实Codex处理成功、资料AT四项通过，原UNKNOWN保持；本地publication故障恢复沿已保存Result，不重发模型。W2只读指定资料库核对四新job succeeded/unknown=false，完整范围见W2报告。自动派发仍关闭；个人项目/历史根未给，Pi认证失败、Claude自动模型选择/许可待答、OpenCode范围阻塞保持，不能称完整全面接入。默认E2E及最终集成验收仍待主控结算。

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
