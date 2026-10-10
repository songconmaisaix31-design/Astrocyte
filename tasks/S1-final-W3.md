# W3 论文检索、插件快照、进度与前端契约对齐（最终权威契约）

Branch `s1-sync-ui-1010`；Task `task_fd085e390fa7` / Dispatch `ctx_95591c220e99`，Run `run_8c1696bb815a`。固定工作树与分支，客户端 OpenCode（当前会话实际模型由运行环境配置，未覆盖模型参数，不声称继承 Codex 模型）。独占写域 `web/src/`（除 `api/`）、`web/public/`、`web/e2e/`、`docs/acceptance/S1-paper-board.md` 与本文件。

## 基线与边界

- 干净工作树普通合入 ROOT 基线，保留原 owner 历史，无 squash/rebase/reset/force，未越写域编辑。
- 消费 W0 **`e39f8f1`**（单一权威 paper/progress GET DTO，`source_key/provider/content_state/pdf_urls`，去掉 `/papers/import` 死路由与 `already_imported` 假字段）与 W1 **`0b189e5`**（`paper_snapshot` 离线摄取 adapter + IPv6 SSRF 加固）。**不再回写 `d89daa1` 的 `id/source_type/availability` 旧 shape**。
- 未改契约/OpenAPI/生成客户端/迁移/锁/入口；`web/src/api/` 由 W0 单一 owner，检索/进度包装方法待 W0 集成生成后替换本地 seam。`paper_snapshot` 仍未进 W0 契约 `adapter` 枚举（W1 reader 已实现），前端在单一 seam 处 cast。

## 完成内容

- **论文检索**：`paperSearchClient.searchPapers` 改为 `GET /api/v1/papers/search?q=&provider=&limit=`，结果按 `PaperSearchHitV1`（`source_key/provider/title/authors/year/venue/arxiv_id/doi/locator/abstract/content_state/pdf_urls`）呈现。`paperSearchKey`=source_key，`canSelectPaper` 仅排除 paywall/restricted，勾选纯人工、可取消。
- **批量入库**：`importPaperBatch` 逐条 `attentionApi.importMaterial`（`kind=paper`，`adapter=arxiv|paper_url`，`source_key`=hit.source_key 去重），**每项确定性幂等键** `stableKey(['import',adapter,source_key,locator])`（替换原 `crypto.randomUUID()`），`Promise.allSettled` 只确认「提交被接受」而非「入库成功」，失败汇总单一可操作错误且成功作业在处理队列可查，不新增 scheduler/cache。
- **插件快照复核**：`parsePluginSnapshot` 保留 `text` 与 `pdf_urls`；`snapshotImportOptions` 提供人类明确选择——`readable_fulltext` 且未截断且正文非空 → `paper_snapshot`（`export_text`=整份快照 JSON，`source_locator`=来源 URL，服务重推导 source_key，不重抓网页）；每个 `pdf_urls` → `paper_pdf`（非 arxiv-only）；arxiv 无正文/PDF 回退 `arxiv`；paywall/restricted/摘要/截断无绕过、不当全文。
- **项目进度**：`progressPresentation.ts` 对齐 `ProjectProgressV1`（`source`/`observed_at` 可空、`percent` 仅人工、`revision`，无 `operations`/`pending_operation`），`ProgressEvidenceV1`（无 `freshness`）；`progressClient.ts` 接 `GET/POST .../progress[/infer]`，infer body 含 `operation_id`(uuid)+`files:[TASK.md,STATUS.md]`；`ProjectProgressPanel.tsx` 呈现 source/percent/observed_at/evidence，`status=unknown` 如实显示「未知」不伪造。
- CLI 原生操作沿用既有 `ManagedProjectsPanel`/`NativeProjectPanel`，权限判断 `nativePermission.ts` 不变。CLI 完整覆盖与记忆范围仍 PENDING，未代选、未建 UI 伪工具。

## DTO 交接（供 W0/W1/W2）

见 `docs/acceptance/S1-paper-board.md` 顶部「DTO 交接」：检索 `GET /papers/search`、入库沿既有 `ImportMaterial`（`adapter=arxiv|paper_url`，`paper_snapshot` 待 W0 加入枚举）、进度 `GET/POST .../progress[/infer]`、记忆（待答后）。本地形状已在 `paperSearch.ts`/`progressPresentation.ts` 精确镜像 W0 生成类型；集成时由 W0 用生成 client 替换本地 seam，UI 结构不变。

## 验证

| 命令/行为 | 结果 |
|---|---|
| `pnpm --dir web typecheck` | PASS/exit0 |
| `pnpm --dir web test` | 82 PASS（论文检索 6 + 插件快照 10 + 进度 5；既有其余保持） |
| 定向 `eslint`（新增/改动文件） | PASS/exit0 |
| `pnpm --dir web build` | PASS/exit0（101 模块） |
| `playwright test paper-search.spec.ts plugin-snapshot.spec.ts progress.spec.ts` | **26 PASS**（1920/1280：检索 GET 契约载荷、示例选择/取消 0 写、501 诚实、移动 390；快照正文保留+`paper_snapshot`、非 arxiv `paper_pdf`、付费墙无绕过；进度 infer 带 `operation_id`/`files` + 未知不伪造） |

本轮 e2e 均为真实契约形状的 mock 对接，**不冒充真实入库**。

## 未完成与真实限制

- 真实端到端闭环（真实检索→勾选→逐条 ImportMaterial→SQLite/objects→新 API 进程去重→三尺寸/keyboard/空/未知错误/撤销边界）待 W0 合并 W1/W2 后端组装后联合验收；本轮未运行个人 5173/8787、未安装插件、未读个人 Chrome、未合 main/远程 CI。
- `paper_snapshot` 未进 W0 契约 `adapter` 枚举，需 W0 集成确认；进度 `percent` 仅人工记录，UI 不伪造数字。
- CLI 完整覆盖、记忆范围仍 PENDING，未代选；记忆与工具层仅规划。
- SOURCE 为 UI/测试精确提交，REPORT 为本文件与接受文档的随后提交；最终集成与主控验收由 ROOT 承担。
