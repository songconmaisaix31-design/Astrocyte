# S1 W2 收尾：OpenCode 会话驱动、获准 Agent 进度返修与工具/记忆规划

分支 `s1-local-agents-1010`，普通合入 ROOT 精确基线 `8277667`（origin/s1/attention-materials-20261009，fast-forward，无 squash/rebase/reset/force）。本文件为 REPORT，业务 SOURCE 提交随本报告一并 push。

## 1. OpenCode 1.18.35 loopback Basic-auth 会话驱动（正式接入，替代 unsupported）

前轮结论「OpenCode 全局 instructions 不可关闭，只能等下一版本」作废。按 root 返修落地正式 driver：

- **隔离机制（同版本源码）**：`@opencode-ai/core/global.ts` `make().config = Flag.OPENCODE_CONFIG_DIR ?? Path.config`（`Path.config = xdgConfig/opencode`）；`session/instruction.ts` `globalFiles = [path.join(global.config,"AGENTS.md"), path.join(global.home,".claude","CLAUDE.md")]`。把 `OPENCODE_CONFIG_DIR` 指向控制器自有空目录即可让用户全局 `AGENTS.md`、MCP、plugins、instructions 不加载；`~/.claude/CLAUDE.md` 由 `OPENCODE_DISABLE_CLAUDE_CODE_PROMPT=1` 关闭、项目层由 `OPENCODE_DISABLE_PROJECT_CONFIG=1` 关闭。**如实限制**：`config.ts` 的 `ConfigManaged.managedConfigDir()`（Windows `%ProgramData%\opencode`）与 auth wellknown/account 远程 config 仍会加载、可能声明 MCP/plugins，本 driver 不宣称压制这些来源；本机 `%ProgramData%\opencode` 不存在，且 `permission:deny` 阻断一切工具使用。
- **deny-by-default（最终 overlay）**：自有空 config 目录写 `opencode.json` 显式逐键 `permission:deny`（read/edit/bash/glob/grep/webfetch/websearch/task/skill/lsp/question/external_directory/doom_loop 全 deny），并设 `OPENCODE_PERMISSION` 相同对象，`config.ts` 在 wellknown/account/managed 之后 `mergeDeep` 它作最终覆盖；不传 `--auto`。**如实限制**：仅 `{"*":"deny"}` 通配不能清除继承的 specific read/bash allow，故逐键 deny；agent-specific permission（agent rules 优先）与 managed/wellknown 注入的 instructions/MCP/plugins 仍不保证压制，本 driver 不宣称残留已隔离，fail-closed 保持 `Configured=unknown`。`--pure` + `OPENCODE_DISABLE_DEFAULT_PLUGINS=1` 关闭外部/默认插件（`--pure` 不保证 MCP/plugins 不初始化）。
- **auth 由 CLI 原生读取**：不复制/读取 provider 凭据；`OPENCODE_SERVER_USERNAME`/`OPENCODE_SERVER_PASSWORD` 每会话随机，仅供 loopback Basic auth，不落库不打印；`opencodeIsolationEnv` 剥离继承的 `OPENCODE_*`/`XDG_CONFIG_HOME`，其余环境（含 CLI 自读的 provider 凭据）保留。
- **协议面**：`opencode serve --hostname 127.0.0.1 --port <N>`；`POST /session`→native ID、`POST /session/:id/message`（同步等待，`parts[]`）、`GET /session/:id`、`GET /session/:id/message`、`GET /session/status`、`GET /global/health`、`POST /session/:id/abort`、`POST /instance/dispose`。

`internal/adapters/agents/opencode.go`：`opencodeNative` 独立会话驱动（与 stdio `*Native` 并列注册），owned subprocess + Job Object 正向退出、未知投递不重放。**同 nativeID 恢复**：`Resume` 经 `GET /session/:id` 加载原 SID 并续发（不 `POST /session` 建新会话，不换 nativeID），跨会话 handoff 才新 nativeID；模型按真实 `AssistantMessage` 的 `modelID`/`providerID` 解析（不再把 `info.model` 当 string）。`Registry.List()` 现为 `codex/pi/claude/opencode`；`installedCommand` 增 opencode 原生 `node_modules/opencode-ai/bin/opencode.exe` 入口；`textSession`/`ConfigurationID` 对非 `*Native` 适配器返回干净 `unsupported_capability`。

能力映射：start/send/read_context/resume/stop/observe 已实现；discover 无已验历史头协议，保持 `unsupported`；reconcile/context_handoff 沿用应用层既有 NativeOperation/模式语义。selected-text/distillation 路径仍依赖 `*Native` 的模型观测，OpenCode `Configured=unknown`，不冒充模型就绪。

## 2. 获准 Agent 进度返修（root 四项）

`internal/workspace/domain/progress.go` + `internal/workspace/app/progress.go`：

- **错误分支不再骗 HTTP 成功**：`settle` 拆为 `failSettle`（返回原 typed `*apierrors.ServiceError`）与 `publishAccepted`（返回 `(record, error)`）。错误分支语义错误恒为非 nil；持久化收据失败仍返回语义错误，且 pre-paid 的 pending 收据已阻断重放。
- **Set 不再清空 Operations**：人类 `SetProjectProgress` 保留 `current.Operations` 与 `PendingOperation`，旧 request 重试命中已 accepted 收据，不会重新付费。
- **accepted retry 返回原始快照**：`ProgressOperation` 增 `Result *ProgressResult` 不可变快照；accepted 重试 `restoreAccepted` 从快照还原，而非返回被后续人类写覆盖的 live 字段。
- **成功结果先 durable 再 publish（两步持久化，非仅循环冒充恢复）**：模型成功且合法输出后，先 `persistComputed` 把结果独立写入 `computed` 收据（durable），再 `publishAccepted` 复核 `result.SettingsRevision` 与模型许可后落 `accepted` + live 字段。进程在 response 后/保存前死亡，或 publish 一直冲突时，同 operation 的 `computed` retry 只重新 publish 已持久结果、不重跑模型；许可 revision 变更则落 `unknown` 丢弃结果。存储真失败无 result 时保持 `unknown` 且不重发。无新框架/新 Attempt。

**W0 对齐**：`ProgressCommand` 仅 `{files}`，操作身份来自传输层 `Idempotency-Key` header（`Caller.OperationID`，`json:"-"`），与 `NativeCommand` 一致；迁移 `013_project_progress` 仍由 W0 唯一 owner。

测试：`app/progress_test.go` 更新为 header 身份，新增回归（Set 保留收据 + accepted retry 返回原始、computed 结果在 publish 失败/冲突后重试恢复且不重跑模型、发布失败不报成功）；`sqlite/local_agents_progress_test.go` 不变仍通过。

## 3. 工具层与长期记忆规划

保留 [swarm-tools-memory.md](../docs/architecture/swarm-tools-memory.md)：共享工具服务按官方 Go MCP SDK 单一 Service + stdio 薄桥、`ToolProfile`（`secret_ref` 不落明文）、按任务最小挂载、批准/提权不暴露为工具、不新建调度框架。**读写分开授权**：`swarm_claim`/`swarm_submit` 是写操作，复用 Service 获准写入路径、逐次按 action/project 授权，不是只读旁路。长期记忆参考 claude-mem 三层渐进披露，只摘要不原文归档。**用户已明确要实现长期记忆（无「仅规划」选项）**，尚未回答的是存储/可见范围（项目隔离/个人全局），未落地迁移表/服务/FTS5 索引，未自动采集；答复后由原 owner 立即落地 workspace 项目记忆表 + 既有 SQLite FTS5 或个人全局记忆。

## 4. 验证

| 命令 | 结果 |
|---|---|
| `go build ./...` | PASS |
| `go vet ./internal/adapters/agents/` | PASS |
| `go test ./internal/workspace/... ./internal/adapters/agents/ ./internal/adapters/sqlite/`（`GOFLAGS=-p=1`） | PASS，含 app 进度回归、agents OpenCode 协议/隔离 env/快照、sqlite 进度 CAS |
| `git diff --check` | PASS |
| `ASTROCYTE_TEST_OPENCODE_LIVE=1 go test ./internal/adapters/agents -run TestOpencodeLiveSessionChain` | **PASS，~12s**：真实 `opencode serve` 隔离启动、显式 `Send` 正向 turn、`ReadContext` 双向、`model=deepseek-flash provider=deepseek` 真实解析、同 nativeID resume 保持原 ID、owned stop 正向退出、**跨会话 context_handoff**（新 nativeID + 携带原 selected context marker 并被新会话引用） |

真实 `opencode serve` 会话链已由原 owner 运行通过；中间一轮旧 `waitHealth` 阻塞 180s（客户端超时）的首次失败保留，已改每请求 3s/整体 20s 有界超时。旧成功媒体/模型与 UNKNOWN 不重发。

## 5. Handoff 提议（交 W0/ROOT，非本轨落地）

- **迁移 `013_project_progress`**（W0 落）：`CREATE TABLE local_project_progress (project_id TEXT PRIMARY KEY, data TEXT NOT NULL, revision INTEGER NOT NULL CHECK (revision >= 1));` + `UPDATE foundation_schema SET version = 6 WHERE context='workspace';` + `INSERT OR IGNORE INTO _migrations(id) VALUES ('013_project_progress');`（沿用 `012_project_metadata` 模式）。
- **HTTP/契约**（W0 落）：`GET /api/v1/local-projects/{id}/progress`（缓存读）、`PUT .../progress`（human 写）、`POST .../progress/infer`（human 发起，`Idempotency-Key` header 提供操作身份，body `{files:[...]}`）；`ProgressCommand` 已收敛为 files-only，domain DTO（`ProgressEvidence`/`ProjectProgress`/`ProgressOperation{Result}`/`ProgressResult`）由本轨落地。
- **组装**：`cmd/server` 注入 `NewProjectProgressService(projects, progressRepo, files, processor, registry)`；OpenCode 已随 `NewRegistry` 注册，无需额外组装。

## 6. 剩余限制

- OpenCode 会话驱动已实现并通过 httptest + 真实会话链（start/send/read/resume 同 ID/stop/handoff 新 ID，model=deepseek-flash/deepseek）；discover 保持 unsupported；selected-text 路径仍 `*Native` 专属，OpenCode `Configured=unknown`；全局 state 目录 flock 与用户 TUI 潜在争用保留为已知限制。
- 进度功能域/app/sqlite 已实现并通过测试，HTTP/迁移/契约/UI 未落地（W0/W3 归属），尚未经新 API 进程 query 验收。
- 记忆：用户已确认要实现，存储/可见范围（项目隔离/个人全局）仍待答，不冻结接口、不自动采集、不降低验收；CLI 完整覆盖范围（三个 vs 全已安装）仍 pending，不代选。
- 未合 main、未跑远程 CI。
