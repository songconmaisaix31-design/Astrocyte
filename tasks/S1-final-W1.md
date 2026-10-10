# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件

## 续四：README 与实际实现对齐返修（本轮）

Branch `s1-sync-attention-1010`。root 已构建当前 MV3 `extensions/paper/dist`，但 `extensions/paper/README.md` 仍留三处旧表述——`text` 仅 display-only、authoritative import 走 `paper_url` 联网重取、应用侧 transport 刻意未接线。实际实现已落地人类复核的 `paper_snapshot` 零网络入库（存原提交 JSON + 正文/版本，provenance 明确 browser_snapshot、非 publisher 在线原件）。本续只改 `extensions/paper/README.md` 与本文件，写域不变，不改代码/接口/锁/权限，不重跑模型/抓网页/造测试。

- README 改为描述真实路径：安装（`extensions/paper/dist` 作 unpacked MV3 加载）→ 点击「读取当前页并提取」→「复制快照」→ 粘贴进 Astrocyte Attention 表单人工审阅确认 → `POST /materials/imports` `adapter=paper_snapshot` 零联网入库；权限如实写 `activeTab`/`scripting`/`clipboardWrite`，无 storage/cookies/history/tabs 读权限。
- 分清 `paper_snapshot`（离线、人类复核、不验 publisher 原件，存 `paper-snapshot.json` + 正文/版本，`browser_snapshot_*` 三态）与可选在线 `paper_url`（抓正文）/`paper_pdf`（抓 PDF 正文）两条独立联网路径；后者依赖公网源访问、当前被本机 fakeIP DNS 阻断，与快照路径无关。
- 说明个人 Chrome 独立插件未安装；隔离 Chromium 临时 profile 加载与真实浏览器粘贴快照入库已通过，不据此宣称个人浏览器已装或多数站点 PDF 全量。
- 应用侧 transport 仍无 `externally_connectable` 消息交接，实际交接是显式「复制→粘贴→人工确认」；`background.js` 只信插件自身 popup、外部 sender 一律拒绝。

验证：`extensions/paper` 源码（`popup.js`/`popup.html`/`manifest.json`/`background.js`/`extract.js`）与 `paper_snapshot.go` 逐项核对一致（快照字段、content_state 三态、source_key 服务端重推导、provenance 落库），`git diff --check` 干净；纯文档改动，无接口/锁/权限变更，未重跑模型/抓网页/造测试。

---

## 续三：检索精确 ID 路由返修（本轮）

Branch `s1-sync-attention-1010`。主控返修：真实 `GET /api/v1/papers/search?q=10.1371%2Fjournal.pdig.0000514&provider=crossref` 把完整 DOI 当关键词，首条返回无关的 `Algorithm1:PDIG`。本续在 `importers/scholar.go` 加精确 ID 路由，写域不变，仅本文件 + `docs/acceptance/S1-paper.md` + `scholar.go`/`scholar_test.go`。

- 完整 DOI/doi.org URL → Crossref 官方 `GET /works/{doi}` 单条，404 显式 `not_found`，不退关键词。
- 合法 arXiv 编号/URL（含 `vN` 版本）→ arXiv `id_list` 精确版本，版本不符/非单条显式 `not_found`。
- 精确 ID 判定先于 provider 关键词、与 provider 无关；默认 crossref 请求也按精确 ID 路由；普通标题/作者仍既有 provider 关键词。
- 不改公开 DTO/HTTP/迁移/锁/入口，无新框架/代理 DNS 例外，不读个人资料/不调模型，正文/PDF/paper_snapshot 提取不动。

验证：协议回归 `TestParseDOIQuery`/`TestScholarExactDOIRouting`/`TestScholarUnknownDOIExplicitNotFound`/`TestScholarExactArxivRouting`/`TestScholarExactArxivVersionMismatch`/`TestScholarKeywordQueryUsesProviderSearch` PASS；真实元数据验证各一次 `TestScholarExactIDLive`（DOI `10.1371/journal.pdig.0000514`、arXiv `2504.16054`→v1）。`go test ./internal/adapters/importers ./internal/attention/...`、`go build ./...`、`go vet`、`gofmt` PASS；`paper-dom.test.mjs` 7 PASS。Handoff：W0 `ctx_601d4fe7909f`（契约/HTTP 无需改，404 已映射）、W3 `ctx_55f6713c28e7`（UI 无需改字段，建议对未知 DOI 提示明确 404）。网络阻断保持，未新增例外。

---

# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件

Branch `s1-sync-attention-1010`。在 11791ce 基础上按主控返修完成 SSRF 边界修正、`paper_pdf` 公共 PDF 正文适配器、隔离 Chrome 插件真实加载快照；本续再补 `paper_snapshot` 插件快照离线入库（零联网/零 model）与 `isPublicIP` IPv6 文档/丢弃段收尾。写域遵循 `tasks/S1-final-paper-cli-memory-plan.md`；OpenAPI/迁移/入口/锁/契约由 W0 单一 owner，本轨仅 `internal/attention/{domain,app}` 除 contracts.go、`internal/adapters/{importers,distillers,objects}`、`internal/adapters/sqlite` 除 local_agents*.go、`extensions/paper` 除锁、`docs/acceptance/S1-paper.md`、本文件。开发客户端 OpenCode / DeepSeek V4 Pro。

## 续二交付

- **`paper_snapshot` 适配器**（`paper_snapshot.go` + `reader.go` case）：人类粘贴审阅后的插件快照 JSON 经 `ImportMaterial(adapter=paper_snapshot)` 离线入库正文，零联网/零 model/无 paywall 绕过/无新协议缓存 daemon；复用既有 objects/jobs/revisions/dedup。`source_key` 由服务端从 `arxiv_id`→ACL URL 内嵌 id（与 `paperPDFSourceKey` 共享 `aclAnthologySourceKey`，含落地页尾斜杠）→`doi:`+`normalizeDOIRaw`→canonical URL 重推导，**绝不信任快照自报 source_key**；调用方显式 `source_key` 不符即拒。provenance `Processor=paper_snapshot`、`Mode=browser_snapshot_fulltext|abstract_only|truncated`；`truncated`/`abstract_only` 绝不冒充全文；`paywall|restricted` 拒 `EvidenceMissing`；original snapshot JSON 存附件 `paper-snapshot.json`。
- **`isPublicIP` IPv6 收尾**：补拒 `2001:db8::/32`（RFC 3849）与 `100::/64`（RFC 6666），普通 global unicast IPv6 不误阻；`TestIsPublicIP` 增例。不改 host DNS、不启用官方域 proxy 策略（用户未选）。
- `aclAnthologySourceKey` 从 `paperPDFSourceKey` 抽取共享，落地页尾斜杠与 `.pdf` 同源归一键；`paper_pdf` 行为不变（既有测试保持）。

## 交付

- **SSRF 修正** `paper_web.go`：`isPublicIP` 拒 CGNAT 100.64/10、fakeIP 198.18/15、文档段、0/8、240/4、广播、IPv4-mapped；`validatePublicTarget` DNS 失败返回 retryable `required_action`；`paperProxy`/`publicPaperRedirect` 全目标校验；`paperDialContext` 直连 pin 公网 IP；移除前轮对任意域放行 fakeIP 的 `isRoutableNonPublic`。
- **paper_pdf 适配器**：`FetchPDF`（%PDF 魔数 + 64MiB）→ `ExtractPDF`（summarize0.25.1 uvx/markitdown 隔离子环境，无 LLM/OCR）→ `paperPDFSource`（PDF 原始附件 + 抽取全文，pdf_url 绝不当元数据全文，ACL 推导 `doi:10.18653/v1/<id>`）。
- **契约对齐 handoff**：本轨 `PaperSearchService`/`PaperSearcher`/`domain.PaperSearchQuery/Hit` 与 W0 b8dfe67 `PaperSearchService`/`PaperSearchCommand`/`PaperMetadata` 冲突，详见 `docs/acceptance/S1-paper.md`；W0 缺 `pdf_urls` 字段需补，否则断 PDF 候选链路。

## 真实观察与验证

- `go test ./internal/adapters/importers ./internal/attention/...` PASS；`node --test paper-dom.test.mjs` 7 PASS；`go build`/`gofmt`/`go vet` PASS。
- `TestPaperPDFExtractLive`：真实 ACL `2024.acl-long.1.pdf` 经隔离子环境 markitdown 抽取 143428 字符，original 147552 字节，36.95s。
- 隔离 Chromium（临时 profile）加载 `extensions/paper/dist` 成功 `hjbbigkdgghbcjfdjoganabagkfkfgca`；PLOS fulltext 63934 字符 `readable_fulltext`、ACL `abstract_only`，source_key/pdf_urls 正确。

## 剩余 / 阻塞

- `paper_url`/`paper_pdf` 的**真实公网下载 vertical 被本机 fakeIP DNS 阻断**（aclanthology/mlr.press/thecvf 解析 198.18.x.x，公网校验正确拒绝）。已向 root 提 A/B/C 三方案待答；B（受控官方域 allowlist + configured proxy + HTTPS/redirect 限制）实现提案备审未启用。用户答复后跑真实 ACL/PMLR/CVF PDF 下载 vertical。
- W0 契约对齐（app 契约 W0 唯一 owner，本轨只 handoff，**不再依据过期 b8dfe67 要求 W0 改 app 签名**，W0 已采用 canonical ports/domain）：`importers.Scholar.SearchPapers` 已实现 `app.PaperSearcher`，`domain.PaperSearchQuery/Hit`（含 `PDFURLs`）就绪，最终接口/枚举/胶水由 W0 定。
- W3 人工审阅 UI 应走 `paper_snapshot` 离线入库正文（不再 `paper_url` 联网）；`source_locator` 传精确 `source_url`、`source_key` 留空。插件→应用 transport 归 W0/W3。个人 Chrome 未装插件；多站点 PDF 全量不声明，仅 ACL 一篇 vertical 已验（抽取），下载 vertical 待 DNS 授权。
