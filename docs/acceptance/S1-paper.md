# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件（续）

2026-10-11。本轨（`s1-sync-attention-1010`）续 11791ce，按 `tasks/S1-final-paper-cli-memory-plan.md` 与主控返修完成 SSRF 网络边界修正、`paper_pdf` 公共 PDF 正文适配器，并实际隔离 Chrome 加载插件取得真实公开页快照。开发客户端 OpenCode / DeepSeek V4 Pro。检索/当前页/批量已确认，CLI 全覆盖与记忆范围仍 PENDING；公开论文 host 的 DNS/fakeIP 例外已向 root 提问待答，未代选。

## 本续交付

### SSRF / 网络边界修正（`internal/adapters/importers/paper_web.go`）

- `isPublicIP` 现拒绝 CGNAT `100.64.0.0/10`、benchmark/fakeIP `198.18.0.0/15`、文档段 `192.0.2.0/24`/`198.51.100.0/24`/`203.0.113.0/24`、`192.0.0.0/24`、`0.0.0.0/8`、`240.0.0.0/4`、广播，以及 IPv4-mapped IPv6（`::ffff:10.0.0.1` 归一为内嵌 IPv4 再判）。
- `validatePublicTarget` 对 DNS 解析失败不再返回 nil，而是 `ProviderUnavailable`（retryable，`required_action=configure_public_source_network_or_retry`）；`publicPaperRedirect` 与 `paperProxy` 均走完整目标校验。
- `paperDialContext`：直连（无代理）hostname 目标解析一次、要求公网地址并 pin 到该 IP，防重绑定；IP 字面量（代理地址）直接拨。本地代理（127.0.0.1）不作为 target 拒绝，target 仍须公网。移除前轮对任意域放行 fakeIP 的错误 `isRoutableNonPublic` 例外。

### 公共论文 PDF 正文（`paper_pdf` 适配器）

- `PaperWeb.FetchPDF`：受控公网下载 + `%PDF-` 魔数校验 + 64MiB 上限；登录墙/HTML 错误页绝不当作 PDF。
- `SummarizeExtractor.ExtractPDF`：复用 summarize0.25.1 本地 PDF→markdown（uvx+markitdown），经生产隔离子环境（`isolatedEnv` + `UVX_PATH` + 独立 `UV_CACHE_DIR`），无 LLM、无 OCR 模型调用；空正文返回 `EvidenceMissing`。
- `reader.go` case `paper_pdf` + `paper.go` `paperPDFSource`：PDF 作为原始附件（`<name>.pdf`）入 objects，抽取 markdown 为全文，`pdf_url` 绝不当元数据全文；`paperPDFSourceKey` 对 ACL Anthology 从 URL 推导 `doi:10.18653/v1/<id>`，其余 canonical URL。source_key 经既有 `ImportMaterial`/job 幂等去重与 `NextMaterialRevision`。
- `SummarizeOptions` 增 `UVXPath`（缺省时构造器 `LookPath("uvx")` 尽力解析）。

## 实际观察（本轮新功能必要动作，未调用模型）

| 动作 | 结果 |
|---|---|
| `go test ./internal/adapters/importers ./internal/attention/...` | PASS（SSRF 段、`paper_pdf` 源键/附件/魔数/校验、检索解析、paperSource 三类状态） |
| `node --test paper-dom.test.mjs` | 7 PASS |
| `TestPaperPDFExtractLive`（隔离子环境，已下载 `aclanthology.org/2024.acl-long.1.pdf`） | PASS：真实 markitdown 抽取 143428 字符，original JSON 147552 字节 |
| 隔离 Chromium 加载 `extensions/paper/dist`（临时 profile，非个人） | 加载成功 `hjbbigkdgghbcjfdjoganabagkfkfgca`；PLOS `readable_fulltext` 63934 字符、ACL `abstract_only`（source_key `doi:10.18653/v1/2024.acl-long.1`、各 1 条 `pdf_urls`） |

## 剩余 / 阻塞

- `paper_url`/`paper_pdf` 的**真实公网下载 vertical 被本机 fakeIP DNS 阻断**：aclanthology.org/proceedings.mlr.press/openaccess.thecvf.com 均解析为 198.18.x.x，收紧后公网校验正确拒绝。已向 root 提三方案（A Clash DNS 例外 / B 受控公开源 allowlist / C 暂不跑），未答；B 的「明确官方域 + configured proxy + 正常 HTTPS/redirect 限制」实现提案已备审，未启用。用户答复后跑真实 ACL/PMLR/CVF PDF 下载 vertical。
- **W0 契约对齐**（`app` 契约由 W0 唯一 owner，本轨只 handoff）：本轨 `app/paper_search.go` 的 `PaperSearchService`/`PaperSearcher` 与 W0 `b8dfe67` contracts.go 的 `PaperSearchService` 同名重复；本轨 `domain.PaperSearchQuery/Hit` 与 W0 `PaperSearchCommand/PaperMetadata` 签名不符。`importers.Scholar.Search` 已就绪，映射见下。W0 合并后由 W0 做胶水：删本轨重复接口、`Service.SearchPapers` 改 `(ctx, Principal, PaperSearchCommand) (PaperSearchResult, error)`、Scholar 命中映射 `PaperMetadata`。**W0 `PaperMetadata` 缺 `pdf_urls` 字段**，会断「搜索→PDF 候选→paper_pdf 导入」链路，需 W0 补 `pdf_urls`（本轨 `ScholarHit.PDFURLs` 已产出）。
- 与 W3 对齐：插件快照字段 `source_url/source_key/content_state/pdf_urls/doi/arxiv_id/authors/title/abstract/warning` 已与 Go `PaperSnapshot` 一致；W3 人工审阅 UI 取 `source_url`+`content_state`+`pdf_urls` 后走 `ImportMaterial(adapter=paper_url|paper_pdf)`。

## 给 W0 的 ports/data 建议（延续前轮，增补）

- `PaperMetadata` 建议补 `pdf_urls []string`（否则 PDF 候选无法从搜索命中传递）；`Site` 即本轨 provider（crossref/europepmc/arxiv），`PublishedAt` 建议 ISO 字符串或 null（本轨 `Year int`）。
- 检索 `POST /api/v1/paper/search`（W0 已定 `PaperSearchCommand{Query,Limit,Site}`），命中 `content_state=abstract_only`。
- 导入沿用 `POST /materials/imports`：`kind=paper`、`adapter=paper_url`（HTML 正文）或 `paper_pdf`（公共 PDF）、`source_locator` 为正文/PDF URL、`source_key` 去重新 revision。`paper_pdf` 依赖 summarize0.25.1 + uvx(markitdown)，需在 `cmd/server` 配 `SummarizeOptions.UVXPath`（本轨已加字段，W0 接线）。

## 验证与命令

- `go test ./internal/adapters/importers ./internal/attention/...` PASS；`node --test paper-dom.test.mjs` 7 PASS。
- `ASTROCYTE_PAPER_PDF_LIVE=1 go test -run TestPaperPDFExtractLive` PASS（36.95s）。
- 隔离 Chromium 插件加载 + 快照脚本 PASS（见上表，脚本在本机 TEMP 诊断目录，不入库）。
- `go build ./...`、`gofmt -l`、`go vet` 均 PASS。
