# S1 论文、CLI、进度与记忆收尾 W0

当前 Dispatch `ctx_a785bfe63b44` / Task `task_3097e734e4d0`；开发 Worker 为 OpenCode，执行环境观察到的模型标识 `deepseek/deepseek-v4-pro`（DeepSeek V4 Pro），与 root 计划中记录的 TUI 观察一致。独占共享契约、HTTP、入口、迁移 013+、生成 API、脚本、根依赖与所有锁；其他领域由原 owner 交付，最后普通合并精确已 push 提交。本轮已完成契约收敛与真实组装，统一集成在 W1/W2/W3 交付最终 SOURCE/REPORT 后进行。

## 基线与边界

- 工作树 `s1-sync-contract-1010` 已在 HEAD `1b8dfd1` 普通合入 root 精确基线 `8277667`，随后普通合入 W1 `11791ce`、W2 `f1ace50`、W3 `97090d5` 三条部分基线（保留原 owner 历史），未 squash/rebase/reset/force，未删除工作树贡献。
- 已读 AGENTS/HANDOFF/STATUS、`S1-final-paper-cli-memory-plan.md` 与 QUESTIONS 最新段，并消费 root 恢复协调消息 `msg_8ba020674a0a` 与 W3 DTO 交接（`docs/acceptance/S1-paper-board.md`）。
- 上一 W0 因 runtime 恢复遗留的未提交契约与迁移 013 已保存，但其中 `PaperSearchService`/`ProgressService` 接口与 W1/W2 实际领域签名分叉；本轮以 W1/W2 领域为权威收敛，未重写其 domain。
- 论文搜索＋正文、显式批量、Agent 据获准 TASK/STATUS 推断进度已确认；CLI 完整覆盖与记忆可见范围两项仍 PENDING，未代选、不据此降低验收。复用同源 HttpOnly/SameSiteStrict 会话与 CSRF，不扩 CORS、不装 daemon、不改个人 5173/8787、不读个人 Chrome/凭据。

## 契约收敛与真实组装（本轮交付）

上一契约接口（`POST /paper/search`、`PaperMetadata`/`inferred` 字段、`ProgressService`）与 W1 领域 `PaperSearchService(SearchPapers(ctx,Principal,domain.PaperSearchQuery) []domain.PaperSearchHit)`、W2 领域 `ProjectProgressService(Get/Set/Infer)` 分叉。本轮从 contracts.go 删除这些分叉声明，以 W1 `paper_search.go`、W2 `progress.go` 的接口与 domain 类型为唯一权威；迁移 013 与 W2 Handoff 完全一致（`local_project_progress`、workspace schema 5→6），保留。

- **OpenAPI（W0 单写）**：`POST /paper/search` 改为 `GET /papers/search?q=&provider=&limit=`（provider 缺省 crossref），新增 `POST /papers/import`，进度为 `GET`/`PUT`/`POST .../infer`；schema 按 W3 DTO 与 W2 domain 收敛为 `PaperSearchResultV1/PaperSearchHitV1/PaperAvailabilityV1`、`PaperImportRequestV1/ResultV1`、`ProjectProgressResultV1/ProjectProgressV1/ProgressEvidenceV1/ProgressOperationV1/SetProgressRequestV1/InferProgressRequestV1`。`pnpm generate` 重新生成 `web/src/api/schema.d.ts`。
- **HTTP（W0 单写）**：`paper.go` 检索为 human-only GET，`content_state`→`availability.status`（readable_fulltext→full_text、abstract_only→metadata_only、paywall/restricted→restricted、空→unknown）映射，`availability.detail` 保留原始值；`already_imported`/`import_material_id` 字段已纳入契约但去重查询仍属 W1 未完成。批量导入 `POST /papers/import` 领域未组装前为显式 501，不伪装成功。`progress.go` 直接透传 W2 domain（evidence/observed_at/source/revision/operations 幂等回执），缺失服务为 501 非空成功。
- **组装（cmd/server，W0 单写）**：`attention.ServiceOptions` 增 `ScholarSearcher: importers.NewScholar()`，`Services.PaperSearch = attention`（真实启用，非 501）；`Services.Progress = workspaceapp.NewProjectProgressService(db, db, agents.ProjectFiles{}, registry, registry)`。`paper_url` 正文提取沿既有 `Reader`（`NewReader` 已含 `Web: NewPaperWeb()`）与 `ImportMaterial` 路径，无需新组装。

## 验证

| 命令 | 结果 |
|---|---|
| `GOFLAGS=-p=1 pnpm check` | **PASS/exit0**：Go fmt/vet/test/mod、依赖、架构、进程停止、redocly lint、示例校验、生成客户端一致、14 项真实 HTTP 验收场、TS/lint、72 前端单测、diff；既有 EventV1 未使用警告保持 |
| `GOFLAGS=-p=1 pnpm build` | **PASS/exit0**：Go 可执行程序与 TS/Vite 生产构建 |

定向 `go test ./internal/adapters/httpapi ./internal/attention/... ./internal/workspace/... ./internal/adapters/sqlite ./cmd/server` 均 PASS（含 W1 scholar/paper_url、W2 progress 领域测试）。本轮只做组装与运输，未重发旧 joint 成功/UNKNOWN、未启用 paid/media opt-in、未读个人 Chrome；论文真实多站点检索、正文与插件加载→应用持久化的真实闭环仍属 W1 与最终集成。

## 给领域 owner 的交接（Handoff）

- **W1**：`GET /papers/search` 契约已定稿（provider 缺省 crossref、limit 1–50、availability 词汇）；剩余为批量导入 `POST /papers/import` 的领域实现（沿既有 `ImportMaterial`/job/source_key 去重）、`already_imported`/`import_material_id` 去重查询，以及插件→应用同源审阅/入库 transport 的真实闭环。W1 `paper_search.go` 接口已作为唯一权威保留，无需改签名。
- **W2**：`GET/PUT/POST infer` 进度 HTTP 契约已按 W2 domain 定稿，cmd/server 已用 `NewProjectProgressService` 组装。OpenCode 原生 selected-context 仍 unsupported、真实进度错误/成功恢复与记忆接口（待答可见范围）由 W2 继续；新迁移 014+ 仍经本轨单写。
- **W3**：OpenAPI 已发布并生成 client；`availability` 词汇与 W3 `paperSearchClient.ts` 对齐，进度 wire 用 W2 domain（`status/evidence/observed_at/source`），集成时由 W3 将本地 `stage/source_refs/freshness` 视图映射到生成类型，UI 结构不变。

## 剩余限制与未执行操作

- 批量导入、检索去重、OpenCode 原生接入与记忆接口未完成；CLI 覆盖范围与记忆可见范围仍 PENDING，未代选。统一集成与最终宽 check/build/browser 在 W1/W2/W3 最终 SOURCE/REPORT 后由本轨一次进行，领域缺陷退原 owner。
- 未标全 CLI/大部分付费全文已支持；未运行个人 5173/8787、未安装插件、未合 main/远程 CI。
