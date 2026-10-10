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
