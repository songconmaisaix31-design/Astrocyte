# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件（续）

## 续三：检索精确 ID 路由（DOI / arXiv ID）

2026-10-11。root 独立真实 `GET /api/v1/papers/search?q=10.1371%2Fjournal.pdig.0000514&provider=crossref&limit=3` 返回 3 条，但首条是 `Algorithm1:PDIG`（DOI `10.7717/peerj-cs.3663/table-101`）——原因是 `scholar.crossref` 把完整 DOI 当 `query` 关键词。UI 承诺输入 DOI/arxivID，默认 Crossref 也同样把 arXiv 编号当关键词。本续在 `internal/adapters/importers/scholar.go` 加**精确 ID 路由**，不改公开 DTO/HTTP/迁移/锁/入口，不新增框架/代理 DNS 例外，不读个人资料/不调模型；正文/PDF/paper_snapshot 提取不动。

- **完整 DOI / doi.org URL**：`parseDOIQuery` 复用 `normalizeDOIRaw`，识别裸 DOI、`doi:` 前缀、`doi.org`/`dx.doi.org` URL（含 `http(s)`），要求完整 `10.<注册商>/<后缀>`（`10.1234` 这种不完整片段不路由）。命中后走 Crossref 官方 `GET /works/{doi}`（`url.PathEscape` 单条查找，参考 `api.crossref.org/swagger-ui` 当前接口）。**404 保持明确 `not_found`（HTTP 404），绝不退化成不相关关键词**。
- **合法 arXiv 编号/URL（含版本）**：复用 `NormalizeArxivID`（严格 regex，含 `vN` 版本），走 arXiv API `id_list` 精确版本（参考 info.arxiv.org user-manual 的 id_list 章节：base 返回最新版，`vN` 返回精确版本）。请求版本与返回不符、或 0/≠1 条，均显式 `not_found`。
- **默认 UI/provider crossref 亦按精确 ID 路由**：精确 ID 判定先于 provider 关键词搜索，与所选 provider 无关（DOI→Crossref、arXiv→id_list）；普通标题/作者仍走既有 provider 关键词（crossref `query=`/europepmc/arxiv `all:`）。
- 单条命中沿用既有 `abstract_only`、`SourceKey`（`doi:`/`arxiv:`）与 `PDFURLs`，无需契约字段变更；`getStatus` 分离 HTTP 状态使 404 与临时失败可区分。

### 验证

- 协议回归（无网络，RoundTripper 断言端点）：`TestParseDOIQuery`、`TestScholarExactDOIRouting`（路径 `/works/10.1371%2Fjournal.pdig.0000514`）、`TestScholarUnknownDOIExplicitNotFound`（404→NotFound）、`TestScholarExactArxivRouting`（`id_list=2504.16054v2`）、`TestScholarExactArxivVersionMismatch`、`TestScholarKeywordQueryUsesProviderSearch` 均 PASS。
- 真实元数据验证（各一次，无正文/付费）：`TestScholarExactIDLive` —— Crossref DOI `10.1371/journal.pdig.0000514` → 单条，标题「Frameworks for procurement…」，year 2024；arXiv `2504.16054` → 单条 `2504.16054v1`，标题「$π_{0.5}$: a Vision-Language-Action Model…」。
- `go test ./internal/adapters/importers ./internal/attention/...` PASS；`go build ./...`、`go vet`、`gofmt` PASS；`node --test paper-dom.test.mjs` 7 PASS。

### 给 W0 / W3 的 handoff（不改契约）

- **W0 `ctx_601d4fe7909f`**：无契约/DTO/HTTP/迁移改动；`not_found` 已由 `httpapi/attention.go` 映射 HTTP 404，无需变更。检索 HTTP 入口沿用 `PaperSearchCommand{Query,Limit,Site}`。
- **W3 `ctx_55f6713c28e7`**：搜索框输入 DOI/arxivID 现已精确命中，UI 无需改字段；建议对未知 DOI 的 404 在搜索面板给明确「未找到该 DOI」提示（而非空结果/不相关命中）。

网络阻断（aclanthology/mlr.press/thecvf 解析 198.18.x.x）与本机 fakeIP DNS 限制保持，未新增例外；以下续二及此前交付为历史。

---

# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件（续）

2026-10-11。本轨（`s1-sync-attention-1010`）续 11791ce，按 `tasks/S1-final-paper-cli-memory-plan.md` 与主控返修完成 SSRF 网络边界修正、`paper_pdf` 公共 PDF 正文适配器，并实际隔离 Chrome 加载插件取得真实公开页快照；本续再补 `paper_snapshot` 插件快照离线入库（无联网、无 model）与 `isPublicIP` IPv6 文档/丢弃段收尾。开发客户端 OpenCode / DeepSeek V4 Pro。检索/当前页/批量已确认，CLI 全覆盖与记忆范围仍 PENDING；公开论文 host 的 DNS/fakeIP 例外已向 root 提问待答，未代选。

## 续二：`paper_snapshot` 插件快照离线入库 + IPv6 SSRF 收尾

目标：真实插件快照已含 `text/source_url/source_key/content_state/provenance`，但 W3 若只再走 `paper_url` 联网会被本机 fakeIP 挡住且丢正文。为此新增**零联网**的 `paper_snapshot` 导入适配器：人类粘贴审阅后的插件快照 JSON → `ImportMaterial` → objects/jobs/revisions/去重，真正闭环而不依赖待答的代理策略。

### `paper_snapshot` 适配器（`internal/adapters/importers/paper_snapshot.go` + `reader.go` case）

- 输入：`adapter=paper_snapshot`、`kind=paper`、`source_locator=source_url`、`export_text=快照 JSON`。无网络读取、无 model、无 paywall 绕过、无新增协议/缓存/daemon/框架；复用既有 `ImportMaterial`/objects/jobs/revisions/dedup 全链路。
- 解析并校验：`schema_version==1`、`source_url==source_locator`（精确）、`source_url` 为无凭据无端口 public HTTPS；`content_state` 三态。
- **来源身份重推导，绝不信任快照自报 `source_key`**（防伪复核/异文合并）：`arxiv_id`（`NormalizeArxivID` 去版本）→ ACL 站 URL 内嵌 id（`aclAnthologySourceKey`，与 `paperPDFSourceKey` 共享，含落地页尾斜杠）→ `doi:` + `normalizeDOIRaw` → `canonicalWebKey(source_url)`。若调用方显式 `source_key` 与重推导不符，沿用既有 `ReadSource` 尾检拒绝。
- provenance：`Processor=paper_snapshot`，`Mode=browser_snapshot_fulltext|browser_snapshot_abstract_only|browser_snapshot_truncated`（明确 browser_snapshot/user_provided，而非网络 publisher 已验证）；`Version` 携带插件 provenance.version 及 `observed <observed_version>`。
- 正文边界：`readable_fulltext` 要求正文非空，`truncated` 显式标 `browser_snapshot_truncated` 并在正文头注「truncated」；`abstract_only` 只保留摘要（丢弃未确立全文结构的正文，绝不冒充全文）；`paywall`/`restricted` 拒绝 `EvidenceMissing`（`choose_accessible_public_paper_or_provide_existing_export`），无 entitlement 绕过。
- 保留 original snapshot JSON 为附件 `paper-snapshot.json`（来源真相原样入 objects），`SourceSpan=browser snapshot: <source_url>`。

### `isPublicIP` IPv6 收尾（`internal/adapters/importers/paper_web.go`）

- 补拒 RFC 3849 文档段 `2001:db8::/32` 与 RFC 6666 丢弃段 `100::/64`（与 IPv4 reserved 同策略）；普通 global unicast IPv6（如 `2606:4700:4700::1111`、`2a00:1450:4001::`）不受误阻。用例见 `TestIsPublicIP`。
- 不改 host DNS、不启用官方域 proxy 策略（用户未选）；已下载 ACL PDF 正文结果沿用，不重复在线下载/模型/旧 UNKNOWN。

## 本续交付（前次）

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
- **W0 契约对齐**（`app`/契约由 W0 唯一 owner，本轨只 handoff，且**不再依据过期 `b8dfe67` 要求 W0 改 app 签名**——W0 已采用 canonical ports/domain）：本轨 `importers.Scholar.SearchPapers` 已实现 `app.PaperSearcher` 端口，`domain.PaperSearchQuery/Hit`（含 `PDFURLs`）与 `app.SearchPapers` 已就绪；最终接口形态、枚举与胶水由 W0 定。若 W0 契约仍需 `pdf_urls` 字段，本轨 `ScholarHit.PDFURLs` 已产出可映射。
- 与 W3 对齐：插件快照字段 `source_url/source_key/content_state/pdf_urls/doi/arxiv_id/authors/title/abstract/warning/truncated/provenance` 已与 Go `PaperSnapshot` 一致。**W3 人工审阅 UI 现应走 `ImportMaterial(adapter=paper_snapshot, export_text=<粘贴的快照 JSON>)` 离线入库正文**，不要再走 `paper_url`（会再联网、被 fakeIP 挡、且丢已提取正文）；`source_locator` 传精确 `source_url`，`source_key` 留空由服务端从 DOI/arxiv/URL 重推导。`paper_url`/`paper_pdf` 仅在 W3 需要服务端联网提取时保留。

## 给 W0 的 ports/data 建议（延续前轮，增补）

- `PaperMetadata` 建议补 `pdf_urls []string`（否则 PDF 候选无法从搜索命中传递）；`Site` 即本轨 provider（crossref/europepmc/arxiv），`PublishedAt` 建议 ISO 字符串或 null（本轨 `Year int`）。
- 检索 `POST /api/v1/paper/search`（W0 已定 `PaperSearchCommand{Query,Limit,Site}`），命中 `content_state=abstract_only`。
- 导入沿用 `POST /materials/imports`：`kind=paper`、`adapter=paper_url`（HTML 正文，联网）或 `paper_pdf`（公共 PDF，联网）或 `paper_snapshot`（插件快照，零联网）；`source_locator` 为正文/PDF/快照 `source_url`，`source_key` 去重新 revision。`paper_pdf` 依赖 summarize0.25.1 + uvx(markitdown)，需在 `cmd/server` 配 `SummarizeOptions.UVXPath`（本轨已加字段，W0 接线）；`paper_snapshot` 不依赖任何联网/uvx/LLM，仅解析粘贴 JSON。
- 其余来源 PDF/landing 若无法与 DOI/arxiv 身份 correlate，`paperSnapshotSourceKey` 回退 canonical URL，**不伪造 source_key、不异文合并**；显式由 provenance 与附件保留原快照。

## 验证与命令

- `go test ./internal/adapters/importers ./internal/attention/...` PASS（含 `TestPaperSnapshot*`、`TestIsPublicIP` IPv6 段、`paper_pdf` 源键/附件/魔数、检索解析、paperSource 三态）；`node --test paper-dom.test.mjs` 7 PASS。
- `ASTROCYTE_PAPER_PDF_LIVE=1 go test -run TestPaperPDFExtractLive` PASS（36.95s）。
- 隔离 Chromium 插件加载 + 快照脚本 PASS（见上表，脚本在本机 TEMP 诊断目录，不入库）。
- `go build ./...`、`gofmt -l`、`go vet` 均 PASS。
