# W3 论文检索、插件快照、进度与前端契约对齐

Branch `s1-sync-ui-1010`；Task `task_88f06e405600` / Dispatch `ctx_986d9118db5e`，Run `run_8c1696bb815a`。固定工作树与分支，客户端 OpenCode（当前会话实际模型由运行环境配置，未覆盖模型参数，不声称继承 Codex 模型）。独占写域 `web/src/`（除 `api/`）、`web/public/`、`web/e2e/`、`docs/acceptance/S1-paper-board.md` 与本文件。

## 基线与边界

- 干净工作树普通合入 ROOT 精确基线 `8277667`，随后快进至 `origin/s1/attention-materials-20261009`（保留原 owner 历史，无 squash/rebase/reset/force，未越写域编辑）。
- 已读 HANDOFF、STATUS、QUESTIONS、本轮计划与 W0 `b8dfe67` 契约、W1 `11791ce` 扩展/paper_url、W2 `90c18d4` 进度收敛。三项已答（论文搜索＋正文、允许批量读取、Agent 从获准 TASK/STATUS 推断进度）可落实；新两项（CLI 完整覆盖、记忆隔离/全局/仅规划）仍 PENDING，未代选、未建 UI。
- 未改契约/OpenAPI/生成客户端/迁移/锁/入口；`web/src/api/` 由 W0 单一 owner，检索/进度包装方法待 W0 集成生成后替换本地 seam。

## 完成内容

本场对齐 W0 已发布契约与 W1 真实输出，替换上一 Worker 本地猜测的假 DTO：

- **论文检索**：`POST /api/v1/paper/search`（`paperSearchClient.searchPapers`，CSRF + `Idempotency-Key`，稳定键按 query 派生）。结果按 W0 `PaperMetadataV1` 呈现（site/source_key/doi/arxiv_id/title/authors/abstract/published_at/locator/content_state/license/warning）；检索只返回公开元数据，原文可得性在勾选入库后由作业确定。勾选纯人工、可取消、批量入库由人显式提交（`importPaperBatch` 逐条 `importMaterial`，各带幂等键，失败汇总为单一可操作错误且成功作业可在处理队列查），不全量。
- **插件快照复核**：`parsePluginSnapshot` 对齐 W1 扩展真实输出 `source_url/content_state/host_family/source_key/doi/arxiv_id/observed_version/pdf_urls/warning/truncated/provenance`；`importAdapterFor` arXiv→`arxiv`、其余公开 HTTPS→`paper_url`（不再仅 arxiv）；`canImportSnapshot` 付费墙/受限无绕过。快照摘要永不当作正文。
- **项目进度**：`progressPresentation.ts` 对齐 W0 `ProjectProgressV1`（status/summary/evidence/inferred/inferred_at/processor/model/warning），契约无百分比字段，`status=unknown` 呈现「未知」不伪造；`progressClient.ts`（`GET /local-projects/{id}/progress`、`POST .../progress/infer`）+ `ProjectProgressPanel.tsx` 挂入 `ManagedProjectsPanel`，人类可读缓存进度并可明确发起一次推断（需项目已许可模型客户端）。
- CLI 原生操作沿用既有 `ManagedProjectsPanel`/`NativeProjectPanel`，权限判断 `nativePermission.ts` 不变，看板不据发现授权。记忆/工具层仅规划，未建 UI 伪工具。

## DTO 交接（供 W0/W1/W2）

见 `docs/acceptance/S1-paper-board.md` 顶部「DTO 交接」：检索 `POST /paper/search`、入库沿既有 `ImportMaterial`（`adapter` 按站点，**需 W0 将 `paper_url` 加入 `ImportMaterialRequestV1.adapter` 枚举**）、进度 `GET/POST /local-projects/{id}/progress[/infer]`、记忆（待答后）。本地形状已在 `paperSearch.ts`/`progressPresentation.ts` 精确镜像 W0 生成类型；集成时由 W0 用生成 client 替换本地 seam，UI 结构不变。

## 验证与首失败

| 命令/行为 | 结果 |
|---|---|
| `pnpm --dir web typecheck` | PASS/exit0 |
| `pnpm --dir web test` | 77 PASS（论文检索 6 + 插件快照 6 + 进度 4；既有其余保持） |
| 定向 `eslint`（新增/改动文件） | PASS/exit0 |
| `pnpm --dir web build` | PASS/exit0（101 模块） |
| `playwright test paper-search.spec.ts plugin-snapshot.spec.ts progress.spec.ts` | **24 PASS**（1920/1280；含 390 移动无横溢、真实契约载荷 mock 对接、501 诚实「检索未完成」、0 写请求边界、付费墙无绕过、进度 GET/POST infer） |
| `playwright test navigation.spec.ts` | 104 PASS（无回归） |

首失败保留：本轮之前未提交的插件快照 e2e 首场 **6 FAIL**（fixture 表单误隐藏致 `.fill` 超时 + `getByText` 严格模式冲突）与进度 e2e 首场 **2 FAIL**（`getByText(/tasks\/S1\.md/)` 命中两条）；分别改表单两模式渲染、断言用 `getByRole('heading')` 后通过，原 RED 不覆盖。`importAdapterFor` 原非法 URL 直接 `new URL` 抛错已修并补单测。

## 未完成与真实限制

- 检索/入库/进度真实后端尚未由 W0 统一组装：UI 已对齐真实契约并直接消费真实路由，但端到端真实检索→入库→SQLite/objects→新 API 进程闭环待 W0 合并 W1/W2 后端后联合验收；本轮 e2e 用真实契约形状 mock 验证 UI 对接，未伪造真实结果。
- `paper_url` 未进 W0 契约 adapter 枚举，需 W0 集成确认；进度契约无百分比（W0 权威），UI 以 status/evidence 呈现，不伪造数字。
- CLI 完整覆盖、记忆范围仍 PENDING，未代选；记忆与工具层仅规划。
- 未运行个人 5173/8787、未安装插件、未读个人 Chrome、未合 main/远程 CI；真实数据三尺寸截图待端到端后由 root 复核。
- SOURCE 为 UI/测试精确提交，REPORT 为本文件与接受文档的随后提交；最终集成与主控验收由 ROOT 承担。
