# 本地 Agent 与工具清点（W3）

日期：2026-10-09。只读探测：`Get-Command` + `--version` / `--help`（带超时）。未读取密钥、完整会话记录或私有文件，未扫描整盘，未安装/升级任何 Agent，未运行任何付费能力会话或真实 summarize。

> 重要约定：本次只核对「可调用入口 + 版本 + 接入提示」。`--version`/`--help` 能证明 CLI 已安装、可调用，**不能**证明真实会话能力（start/resume/send/stop 的实际效果）。凡未实际运行会话验证的能力一律记为 `unknown` / NOT_RUN。

## 1. 环境基线（实际命令结果）

| 工具 | 版本 | 路径 | 命令 |
|---|---|---|---|
| Node | v24.16.0 | `C:\Program Files\nodejs\node.exe` | `node -v` |
| pnpm | 11.27.0 | `C:\Users\DW\.local\bin\pnpm.ps1` | `pnpm -v` |
| Go | go1.27.2 windows/amd64 | `C:\Users\DW\AppData\Local\Programs\go\bin\go.exe` | `go version` |
| orca | 1.4.220 | `...\orca\resources\bin\orca.exe` | `orca --version` |
| bl（百炼） | 1.1.2 | npm 全局 | `bl --version` |
| dws（钉钉） | v1.0.60 | npm 全局 | `dws --version` |

> 注意：STATUS.md 记录「Go 正在安装便携工具链」，现已完成并进入 PATH（go1.27.2）。`go` 已可被 `Get-Command` 解析到，不再是缺失项。

## 2. Agent CLI 清点（已安装、可调用）

| Agent | 版本 | 入口 | 说明 |
|---|---|---|---|
| Codex | 0.160.0 | `codex.ps1`（npm 全局） | `codex --version` |
| Claude Code | 2.1.238 | `claude.ps1`（npm 全局） | `claude --version` |
| OpenCode | 1.18.35 | `opencode.ps1`（npm 全局） | `opencode --version` |
| Pi | 1.0.1 | `pi.ps1`（npm 全局） | `pi --version` |
| Grok | 1.0.34 (3736acbc8658) | `C:\Users\DW\.grok\bin\grok.exe` | `grok --version` |
| Kimi Code | 24.15.0 | `C:\Users\DW\.kimi-code\bin\kimi.EXE` | `kimi -V`（`--version` 无输出，用 `-V`） |
| Qwen Code | 0.24.6 | `qwen.ps1`（npm 全局） | `qwen --version` |
| Cursor | 3.23.23 (2dac2428…) | `cursor.cmd`（GUI 应用 CLI 包装） | `cursor --version` |

未发现（`Get-Command` 无结果）：`opc`、`cc`、`copilot`、`gemini`、`ollama`、`firecrawl`、`windsurf`、`aider`、`deno`、`bun`、`laya`/`laya-ts`、`jev`。

## 3. 能力矩阵（按 SPEC 8.1 的 discover/read_context/start/resume/send/stop/observe/reconcile）

取值：`supported`=帮助文本明确提供对应子命令/参数；`unknown`=帮助未体现或未实测。**均为 CLI 表面能力，未跑真实会话。**

| 能力 | Codex | Claude | OpenCode | Pi | Grok | Kimi | Qwen | Cursor |
|---|---|---|---|---|---|---|---|---|
| discover | `codex agents` | `claude doctor`/`agents --json` | `session list`/`attach` | `pi list`/`--list-models` | `sessions list/search`/`dashboard`/`doctor` | `session list`/`doctor` | `sessions list/ps` | unknown |
| read_context | exec `--json`/`--output-last-message` | `-p` `--output-format stream-json` | `export [id] --sanitize` | `--export <html>` | `export <id>` (Markdown) | `export [id] -o zip` | `--output-format json/stream-json` | unknown |
| start | `codex [p]`/`exec`/`--worktree`/`-C` | `claude [p]`/`-p`/`--bg`/`-w` | `run [msg..]`/`[project]` | `pi [msg]`/`-p` | `grok [p]`/`-p`/`--output-format` | `kimi`/`-p` | `qwen [query]`/`-p`/`-i` | unknown |
| resume | `resume [id]`/`--last`/`exec resume` | `-r [id]`/`-c`/`--session-id`/`--fork-session` | `-c`/`-s [id]`/`--fork` | `--continue`/`--resume`/`--session`/`--fork` | `-c`/`-r [id]`/`--fork-session` | `-S [id]`/`-c` | `-c`/`-r [id]`/`--session-id`/`--fork-session` | unknown |
| send | `queue --thread <id> --message` | `--bg` + `claude agents` | `serve` + `attach` | `--mode rpc` | `agent`/`--output-format streaming-json` | `web`/`rc` 本地服务 | `serve` daemon / `--input-file` | unknown |
| stop | unknown | `claude agents`（后台管理，暗示） | unknown | unknown | `leader kill` | unknown | serve 会话 reaping | unknown |
| observe | `agents`/`app-server daemon` | `agents --json` | `serve`/`attach`/`stats` | `--export`/`list` | `leader list/info`/`usage`/`dashboard` | `vis [id]`/`session list` | `serve`（SSE `/events`）/`sessions ps` | unknown |
| reconcile | unknown | unknown | unknown | unknown | unknown | unknown | unknown | unknown |
| ACP/MCP | 无 ACP（`app-server`/`remote-control` 实验） | 无 ACP（`mcp` 子命令） | `acp` ✓ | 无（`--mode rpc`） | `--output-format streaming-json` + `agent stdio/serve/headless` ✓ | `acp` ✓ | `--acp` ✓ + `serve --http-bridge` | unknown |

### 关键接入提示（nonsecret，摘自 `--help`）

- **Codex**：`codex exec [PROMPT] --json -o out.txt` 非交互执行；`codex resume <SESSION_ID> --last`；`codex queue --thread <UUID> --message <text>` 向已存在会话投递；`codex app-server daemon`（实验）提供控制 socket；`-C <dir>` 指定工作根；`--worktree` 隔离工作树。
- **Claude Code**：`claude -p "prompt" --output-format stream-json` 非交互流式输出；`claude --bg` 后台 agent + `claude agents --json` 观察；`-r <id>`/`-c` 接续；`--session-id`/`--fork-session` 指定/分叉会话；`--remote-control [name]` 受控远程；无 ACP。
- **OpenCode**：`opencode run`、`opencode session list`、`opencode export <id> --sanitize`、`opencode serve`/`attach`、`opencode acp`（ACP 服务）、`opencode stats`（用量）。
- **Pi**：`pi -p` 非交互；`pi --continue`/`--resume`；`pi --export <file>`（HTML）；`pi --mode rpc`；`pi --tools read,grep,find,ls` 只读模式。
- **Grok**：`grok -p "p" --output-format json|streaming-json`（streaming-json 即 ACP 原生格式）；`grok -r [id]`/`-c`；`grok sessions list`；`grok export <SESSION_ID> [OUT]`（Markdown）；`grok leader list|kill`（进程级停止）；`grok agent stdio|serve|headless`。
- **Kimi**：`kimi -p "p" --output-format stream-json`；`kimi -S [id]`/`-c`；`kimi session list`；`kimi export [id] -o out.zip`；`kimi acp`（ACP stdio）；`kimi web`/`kimi rc` 本地服务。
- **Qwen**：`qwen -p`/`--acp`；`qwen -r [id]`/`-c`/`--fork-session`；`qwen sessions list/ps`；`qwen serve --port 4170`（HTTP daemon：`POST /session`、`GET /session/:id/events` SSE，支持 `--input-file` JSONL 双向输入）；`--approval-mode plan|default|auto-edit|auto|yolo`。
- **Cursor**：仅 GUI 应用的 CLI 包装，`cursor --version` 可用；未见 headless 会话/接续子命令 → 适配价值待定（unknown）。

## 4. summarize / 导入方向（P10「沿用现有视频总结工具」）

| 工具 | 版本 | 说明 |
|---|---|---|
| summarize | 0.21.8 | npm 全局；网页 + YouTube 总结，`--extract` 纯抽取，`--youtube auto|web|yt-dlp|apify`，`--diarize`，`--slides`，`--model cli/<provider>`（claude/gemini/codex/opencode/pi…，经 `CLAUDE_PATH`/`CODEX_PATH` 等定位） |
| whisper | python3.13 Scripts | 本地转写（`--transcriber whisper`） |
| yt-dlp | python3.13 Scripts | YouTube 音轨抽取 |
| ffmpeg | ShadowBot 自带 | 音视频处理 |

> summarize 支持 `--extract --format md --markdown-mode llm` 导出结构化转写，是「视频导入方向」最接近现有 S1 适配输入的 CLId；真实样本格式仍未选定，未运行真实 summarize。

## 5. 结论与限制

- 已安装、可调用、版本已核实的 Agent：Codex、Claude Code、OpenCode、Pi、Grok、Kimi Code、Qwen Code、Cursor（共 8 个）。SPEC 1.2 要求的至少两个真实适配器（优先 Codex、Claude Code）均已具备 CLI 入口。
- 具备 ACP 入口：OpenCode、Grok（streaming-json/agent）、Kimi、Qwen；Codex/Claude 无 ACP，走 `exec`/`-p` + 各自 daemon/`--bg` 通道。
- 所有能力均为 `--help` 表面证据；真实会话的 start/resume/send/stop/observe 全部 NOT_RUN，`stop`/`reconcile` 对绝大多数 Agent 仍是 `unknown`。不得把版本命令当作能力验收。

## 6. 下一步建议探针（不阻塞 S0，真实接入时执行）

1. `codex exec "…" --json`、`codex queue` 最小非交互往返，确认会话 ID 与输出结构。
2. `claude -p "…" --output-format stream-json` 确认 stream-json 事件可解析。
3. `opencode serve` / `qwen serve --http-bridge` / `kimi acp` / `grok agent stdio` 各起一次本地会话，验证 ACP/HTTP 的 start/observe/send/stop 真实行为。
4. 用一个真实视频样本跑 `summarize <url> --extract --format md` 锁定导出格式（S1 前提）。
5. 核实 Kimi/Qwen 的 `stop`（会话终止）与各 Agent 的 `reconcile`（对账）是否有隐藏子命令（`<cmd> help`）。
