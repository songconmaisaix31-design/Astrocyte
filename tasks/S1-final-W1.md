# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件

Branch `s1-sync-attention-1010`。在 11791ce 基础上按主控返修完成 SSRF 边界修正、`paper_pdf` 公共 PDF 正文适配器、隔离 Chrome 插件真实加载快照。写域遵循 `tasks/S1-final-paper-cli-memory-plan.md`；OpenAPI/迁移/入口/锁/契约由 W0 单一 owner，本轨仅 `internal/attention/{domain,app}` 除 contracts.go、`internal/adapters/{importers,distillers,objects}`、`internal/adapters/sqlite` 除 local_agents*.go、`extensions/paper` 除锁、`docs/acceptance/S1-paper.md`、本文件。开发客户端 OpenCode / DeepSeek V4 Pro。

## 交付

- **SSRF 修正** `paper_web.go`：`isPublicIP` 拒 CGNAT 100.64/10、fakeIP 198.18/15、文档段、0/8、240/4、广播、IPv4-mapped；`validatePublicTarget` DNS 失败返回 retryable `required_action`；`paperProxy`/`publicPaperRedirect` 全目标校验；`paperDialContext` 直连 pin 公网 IP；移除前轮对任意域放行 fakeIP 的 `isRoutableNonPublic`。
- **paper_pdf 适配器**：`FetchPDF`（%PDF 魔数 + 64MiB）→ `ExtractPDF`（summarize0.25.1 uvx/markitdown 隔离子环境，无 LLM/OCR）→ `paperPDFSource`（PDF 原始附件 + 抽取全文，pdf_url 绝不当元数据全文，ACL 推导 `doi:10.18653/v1/<id>`）。
- **契约对齐 handoff**：本轨 `PaperSearchService`/`PaperSearcher`/`domain.PaperSearchQuery/Hit` 与 W0 b8dfe67 `PaperSearchService`/`PaperSearchCommand`/`PaperMetadata` 冲突，详见 `docs/acceptance/S1-paper.md`；W0 缺 `pdf_urls` 字段需补，否则断 PDF 候选链路。

## 真实观察与验证

- `go test ./internal/adapters/importers ./internal/attention/...` PASS；`node --test paper-dom.test.mjs` 7 PASS；`go build`/`gofmt`/`go vet` PASS。
- `TestPaperPDFExtractLive`：真实 ACL `2024.acl-long.1.pdf` 经隔离子环境 markitdown 抽取 143428 字符，original 147552 字节，36.95s。
- 隔离 Chromium（临时 profile）加载 `extensions/paper/dist` 成功 `hjbbigkdgghbcjfdjoganabagkfkfgca`；PLOS fulltext 63934 字符 `readable_fulltext`、ACL `abstract_only`，source_key/pdf_urls 正确。

## 剩余 / 阻塞

- `paper_url`/`paper_pdf` **真实公网下载 vertical 被 fakeIP DNS 阻断**（aclanthology/mlr.press/thecvf 解析 198.18.x.x，公网校验正确拒绝）。已向 root 提 A/B/C 三方案待答；B（受控官方域 allowlist + configured proxy + HTTPS/redirect 限制）实现提案备审未启用。
- W0 契约合并后由 W0 做胶水（删重复接口、`Service.SearchPapers` 签名、Scholar→PaperMetadata 映射、补 `pdf_urls`）。
- 个人 Chrome 未装插件；插件→应用同源审阅/入库 transport 归 W0/W3。多站点 PDF 全量不声明，仅 ACL 一篇 vertical 已验（抽取），下载 vertical 待 DNS 授权。
