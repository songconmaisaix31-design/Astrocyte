# S1 W2 收尾：OpenCode 协议复核、获准 Agent 进度与工具/记忆规划

分支 `s1-local-agents-1010`，普通合入 ROOT 精确基线 `8277667`（origin/s1/attention-materials-20261009，fast-forward，无 squash/rebase/reset/force）。本文件为 REPORT，业务 SOURCE 提交随本报告一并 push。

## 1. OpenCode 1.18.35 正式接口复核

按用户要求重查官方 [Server](https://docs.opencode.ai/docs/server/)、[Config](https://docs.opencode.ai/docs/config/)、[Permissions](https://docs.opencode.ai/docs/permissions/) 与同版本源码（`v1.18.35` 的 `config.ts`、`session/instruction.ts`、`mcp/index.ts`），本机 `--version` 仍为 `1.18.35`、`--help` 均 exit 0。结论未变，本次补齐可执行侧证据：

- **正式原生接口存在**：`opencode serve --hostname 127.0.0.1 --port <N>` 暴露 headless HTTP + OpenAPI 3.1（`/doc`），含会话 `POST/GET /session[/:id]`、`/session/:id/message`（`noReply`/`system`/`tools` 字段）、`/prompt_async`、`/abort`、`/permissions/:permissionID`、`/instance/dispose`；`OPENCODE_SERVER_PASSWORD`/`OPENCODE_SERVER_USERNAME` 提供 loopback HTTP Basic Auth。属“HTTP loopback owned server + auth + provided prompt”可行协议面。
- **deny-by-default 工具可达成**：`permission` 配置（v1.1.1 起合并旧 `tools` 布尔），`Flag.OPENCODE_PERMISSION` 用 `mergeDeep` 按进程注入；不传 `--auto` 并对权限请求回 `deny` 即可拦下 read/edit/bash/webfetch/mcp 等。这是按进程环境变量，不是覆盖用户配置。
- **无法安全关闭全局 instructions**：`config.ts` 的 `mergeConfigConcatArrays` 对 `instructions` 做数组拼接 + Set 去重，晚到的 `instructions: []` 不能清掉已继承项；`instruction.ts` 的 `globalFiles` 恒加载 `~/.config/opencode/AGENTS.md` 与 `~/.claude/CLAUDE.md`（仅 `disableClaudeCodePrompt` 关后者，AGENTS.md 无开关），并按 `findUp` 自动读项目 `AGENTS.md/CLAUDE.md/CONTEXT.md`，再叠加拼接后的 `config.instructions`。`OPENCODE_DISABLE_PROJECT_CONFIG` 只关项目层，不关全局。
- **无法全局关闭 MCP**：`mcp/index.ts` 遍历合并后的 `cfg.mcp`，对每个 `enabled !== false` 条目立即 connect（stdio spawn / 远程 HTTP/SSE），无 pure 级全局抑制。

因此 OpenCode 1.18.35 的 selected-context 边界（只交付获准正文、不泄漏全局指令、不连接用户 MCP）仍不能在不覆盖用户 CLI 配置的前提下成立。原生适配对 OpenCode 保持 `unsupported`；安装/版本/help 分别记录，不制造 PTY/daemon 包装、不改全局配置、不读 provider 认证。详见 [local-agents.md](../docs/probes/local-agents.md)。后续版本若有干净的 instructions/MCP 抑制开关，即可沿上述 HTTP 服务协议落地 driver。

## 2. 获准 Agent 进度

新增 per-project 进度记录，来源区分 `human` 与 `agent_inferred`，含 stage、可选 percent、依据 refs（读到的固定文件 path+version）、native_id/model 与 `observed_at` 新鲜度。

- `internal/workspace/domain/progress.go`：`ProjectProgress`、`ProgressBasis`、`ProgressInput`、`ProgressCommand` 与 stage/percent 校验。
- `internal/workspace/app/progress.go`：`ProjectProgressRepository` 与 `ProjectProgressService`（Get/Set/Infer）。Get 只读持久缓存、不启动 Agent；Set 仅人类且校验 stage/percent；Infer 仅人类发起，要求项目 `ExternalModelCLI` 同意 + adapter 存在，经既有 `ProjectFiles.ReadProjectFiles` 读固定文件（范围 B 有界，≤8 文件），经既有 `TextProcessor.ProcessSelectedText`（selectedtext native + 512KiB 边界）产出结构化 `{"stage","percent"}`，解析失败返回 `evidence_missing`，不伪造成功。依据 refs 来自实际读到的文件，不从模型输出取信。
- `internal/adapters/sqlite/local_agents_progress.go`：`LoadProjectProgress`/`SaveProjectProgress`，与 project metadata 相同的 revision CAS；缺失项目返回零值，不伪造 stage；不创建 grant/project/session。
- 测试：`app/progress_test.go`（人/Agent/匿名写入拒绝、超界 stage/percent、模型同意前置、依据 refs、缓存 GET 不启动 processor、Agent 读需 grant 且撤销后拒绝、非法模型输出拒绝）；`sqlite/local_agents_progress_test.go`（round-trip + CAS + 不触碰 projects/grants）。

进度的人类阶段 vs Agent 推断两者并存且来源可区分，不据活动推完成、不自授予权限；HTTP/迁移/契约由 W0 唯一 owner 落（见 §5 Handoff）。

## 3. 工具层与长期记忆规划

新增 [swarm-tools-memory.md](../docs/architecture/swarm-tools-memory.md)：共享工具服务（SPEC §8.4/§13.2）按官方 Go MCP SDK 单一 Service + stdio 薄桥接、`ToolProfile`（工具引用/配置覆盖/`secret_ref`/版本，不落明文密钥）、按任务最小挂载、批准/提权不暴露为工具、不新建调度框架。长期记忆参考 [claude-mem](https://github.com/thedotmack/claude-mem)（Apache-2.0，`search→timeline→get_observations` 三层渐进披露，会话观察→摘要），只参考思路、不读 Codex 平台 memories 或用户 CLI 全部历史当产品记忆。**记忆存储/可见范围（项目隔离/全局/仅规划）仍未获用户答复，保持 PENDING**：不落地迁移表/接口/FTS5 索引，不自动采集；答复后由原 owner 落 workspace 项目记忆表+服务+既有 SQLite FTS5，或仅规划。

## 4. 验证

| 命令 | 结果 |
|---|---|
| `go build ./...` | PASS |
| `go test ./...`（含 app/sqlite 新进度测试） | PASS，全部 ok |
| `GOFLAGS=-p=1 pnpm check` | Go gofmt/vet/test/mod、19 包边界、依赖、stop-process、redocly（1 个既有 EventV1 警告）、237 契约例、generate-client、14 验收例、tsc 全部 PASS；eslint 首次进程崩溃（0xC0000409，与 Go 改动无关），`pnpm --dir web lint` 单独重跑 PASS（0 error/0 warn） |
| `pnpm --dir web test`（vitest） | 61 PASS |
| `git diff --check` | PASS |

## 5. Handoff 提议（交 W0/ROOT，非本轨落地）

- **迁移 `013_project_progress`**：`CREATE TABLE local_project_progress (project_id TEXT PRIMARY KEY, data TEXT NOT NULL, revision INTEGER NOT NULL CHECK (revision >= 1));` + `UPDATE foundation_schema SET version = <next> WHERE context='workspace';` + `INSERT OR IGNORE INTO _migrations(id) VALUES ('013_project_progress');`（沿用 `012_project_metadata` 模式）。
- **HTTP/契约**：`GET /projects/{id}/progress`（缓存读，human 或带 `read_context` grant 的 Agent）、`PUT /projects/{id}/progress`（human 写 stage/percent）、`POST /projects/{id}/progress/infer`（human 发起，body `{files:[...]}`，返回 202 或同步结果）；OpenAPI 契约与生成类型由 W0 落。
- **组装**：`cmd/server` 注入 `NewProjectProgressService(projects, progressRepo, files, processor, registry)`，`TextProcessor` 复用现有 `agents.Registry`。

## 6. 剩余限制

- OpenCode 原生 selected-context 保持 `unsupported`；CLI 完整覆盖范围仍待用户，不能称全机完整接入。
- 进度功能域/app/sqlite 已实现并通过测试，但 HTTP/迁移/契约/UI 未落地（W0/W3 归属），尚未经新 API 进程 query 验收。
- Codex/Claude 既有同 nativeID 成功与跨会话 context_handoff 代码路径保留，本轮未重跑付费原生 turn（旧成功与 UNKNOWN 不重发）；真正跨会话交接的真实浏览器验收仍待协调者安排一次付费运行。
- 记忆存储/可见范围 PENDING，不冻结接口、不自动采集、不降低验收。
- 未合 main、未跑远程 CI。
