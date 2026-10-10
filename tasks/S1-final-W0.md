# S1 论文、CLI、进度与记忆收尾 W0（续：单一权威契约发布）

当前 Dispatch `ctx_a57bbaeba6b1` / Task `task_6a2d3346d6dd`；开发 Worker OpenCode / `deepseek/deepseek-v4-pro`。独占共享契约、HTTP、入口、迁移 013+、生成 API、脚本、根依赖与所有锁。本轮按 root 协调 `msg_00bd90978c65` / `msg_580e5db6f626` 收敛为**单一最小权威 GET/DTO**，通知 W3 同 owner 对齐客户端，并只做 enum/schema/glue，未写 W1/W2/W3 领域业务。上一 W0 里程碑 `d89daa1`/`dcdc5db` 仅契约里程碑，非最终验收。

## 基线与合并

- 工作树 `s1-sync-contract-1010` 沿原分支，普通合并 W1 本轮精确 SOURCE `91b57578ccf1bc797e5adf0026578c139642b1a9`（SSRF CGNAT/fakeIP/保留段收紧 + `paper_pdf` 适配器 + `UVXPath`）与 REPORT `2d315c23ac41418df84de56bf18da0721c3e45c9`，保留 W1 owner 历史，未 squash/rebase/reset/force。
- W1 SOURCE 从 REPORT 父提交查得精确 SHA；W1 新增 `paper_pdf`（真实本地 ACL PDF 抽取 143428 字符）与隔离 Chrome 插件 PLOS/ACL 快照，真实网络下载仍被本机 fakeIP `198.18.x` 阻断、公网校验正确拒绝；用户代理策略未答，未放行、未改 host/DNS。

## 契约单一发布（W0 唯一 owner）

- **`ImportMaterialRequestV1.adapter`** 增 `paper_url`/`paper_pdf`，枚举顺序 `arxiv/paper_url/paper_pdf/summarize_url/summarize/summarize_json/summarize_markdown/manual`。
- **`PaperSearchHitV1`** 收敛为 W1 canonical domain 的字段：`source_key`（替换旧 `id`）、`provider`（替换旧 `source_type`）、`content_state`（原始 `readable_fulltext|abstract_only|paywall|restricted|unknown`，替换旧 `availability.status/detail`）、新增 `pdf_urls`；**删除**伪字段 `already_imported`/`import_material_id`（无真实源键去重查询前不伪造 false，去重在 `ImportMaterial` 作业解析）。删除 `PaperAvailabilityV1`。
- **`/papers/import`** 整条路径及 `PaperImportItemV1`/`PaperImportRequestV1`/`PaperImportResultV1` 删除：批选由 W3 客户端逐项复用既有 `ImportMaterial`，不再保留死 501 完整功能入口。
- **`ProjectProgressV1`** 删除内部幂等回执 `operations`/`pending_operation`（不泄内部 job receipt）；`observed_at` 改 nullable（缺失时 null，非 `0001-01-01` 伪记录）；`source` 改 nullable（缺失时 null）；`status` 缺失时 wire 显式 `unknown`。删除 `ProgressOperationV1`。示例改 `status=unknown/source=null/observed_at=null`。

## 组装与运输（W0 唯一 owner）

- `internal/adapters/httpapi/paper.go`：`paperSearchHitWire` 改为 `source_key/provider/content_state/pdf_urls`，`mapPaperSearchHit` 直接映射 W1 `domain.PaperSearchHit`（`PDFURLs→pdf_urls` 逐字透传）；移除 `availabilityStatus` 与 `POST /papers/import` 501 注册。
- `internal/adapters/httpapi/progress.go`：新增 `progressWire`/`mapProgress`，`observed_at` 零值→null、`status` 空→`unknown`、`source` 空→null、`evidence` 恒 `[]`；不暴露 `operations`/`pending_operation`。权限识别与错误透传仍由 W2 `ProjectProgressService`（`progressProject` human/agent-read-grant + 错误映射）负责，transport 只投影。
- `cmd/server/config.go`：`SummarizeOptions.UVXPath` 接 `ASTROCYTE_UVX_PATH`（空则构造器 `LookPath("uvx")` 尽力解析，缺失如实报不可用）。检索/进度组装沿用上一轮（`ScholarSearcher: importers.NewScholar()`、`PaperSearch: attention`、`Progress: NewProjectProgressService(...)`），未重复定义 W1/W2 端口签名。
- `pnpm generate` 重新生成 `web/src/api/schema.d.ts`（W0 单写）。

## 验证

| 命令 | 结果 |
|---|---|
| `go build ./...` | PASS |
| `go test ./internal/adapters/httpapi ./internal/adapters/importers ./internal/attention/... ./internal/workspace/... ./cmd/server` | PASS |
| `pnpm generate` + `pnpm check:contracts` | PASS：redocly lint 有效（仅既有 EventV1 未用警告）、241 契约示例校验、生成客户端一致 |
| `GOFLAGS=-p=1 pnpm check` | **PASS/exit0**：gofmt/vet/mod、19 包架构、依赖、进程退出、Go 全测试、14 项 S1 acceptance、TS/lint、72 前端单测、git diff --check |
| `GOFLAGS=-p=1 pnpm build` | **PASS/exit0**：Go 程序与 Vite 生产构建 |

未重发旧 joint 成功/UNKNOWN、未启用 paid/media、未读个人 Chrome/5173/8787、未安装插件。真实搜索→人工选择→正文/PDF/source 版本→SQLite/objects/新 API 进程闭环与插件字段复核属 W1/W3 最终集成，未在本轨伪造在线抓取成功。

## 给 owner 的 Handoff（已发 W3 ctx_95591c220e99、W1 ctx_3e887faf41c6）

- **W3**：`GET /api/v1/papers/search`（q/provider/limit，缺省 crossref）返回 `PaperSearchHitV1{schema_version,query,items[],next_cursor,has_more,warnings}`，`items[]` 字段为 `source_key/provider/title/authors/year/venue/arxiv_id/doi/locator/abstract/content_state/pdf_urls`；**不再有** `id/source_type/availability/already_imported/import_material_id`。请把 `paperSearchClient.ts` 的 `RawHit/mapHit` 改读这些字段（`site←provider`、`published_at/license/warning` 如为本地展示需要请按可空本地视图映射，勿当契约字段），`importPapers`/`PAPER_IMPORT_PATH` 移除、批选改逐项 `attentionApi.importMaterial(adapter=paper_url|paper_pdf)`。进度 wire 为 `status/summary/percent/source(nullable)/evidence[]/native_id/model/observed_at(nullable)/warning/revision`（`source_path/kind/version/excerpt`，无 `freshness`），`status=unknown`/`observed_at=null`/`source=null` 表示未观测；`progressPresentation.ts/progressClient.ts` 的 `inferred/inferred_at/processor/freshness` 改为 `source/observed_at`。`web/src/api/` 已生成 `searchPapers/getProjectProgress/setProjectProgress/inferProjectProgress` 包装，可替换本地 seam。
- **W1**：契约 adapter 枚举已含 `paper_url`/`paper_pdf`，`source_key`/`content_state`/`pdf_urls` 单一定稿，port 签名未回改。`paper_snapshot`（插件完整公开页 JSON→既有 Service）沿用既有 `InspectPaperHTML`/`paperSource` 与 `ImportMaterial`，**保留 snapshot.text 正文真实性与明确 provenance**，不要丢弃正文把全部插件导入退成 `paper_url` 重新联网（本机 fakeIP 会阻断）；直接摄取已准当前页正文复用既有 export/manual 属 W1 owner 适配。真实 ACL/PMLR/CVF PDF 下载 vertical 待 DNS 授权，已下载真实 PDF/Chrome page body 可测可持久化，不标在线抓取成功。

## 剩余限制与未执行

- 统一集成（合并 W2/W3 最终 SOURCE/REPORT 后真实 search→选择→正文/PDF 持久化→插件复核→冷重启闭环、真实浏览器三尺寸）待 W2/W3 发布精确 REPORT，由本轨做最后路由/导入/config/type 胶水，领域返修退 owner。CLI 完整覆盖与记忆可见范围仍 PENDING，未代选、未默认权限。未合 main/远程 CI、未升级个人预览。
