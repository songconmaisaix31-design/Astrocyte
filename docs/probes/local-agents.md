# 本地 Agent 与工具清点（W3）

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
