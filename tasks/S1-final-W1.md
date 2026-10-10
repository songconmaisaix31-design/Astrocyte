# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件

Branch `s1-sync-attention-1010`。普通合入 ROOT 精确基线 `8277667`（`origin/s1/attention-materials-20261009` fast-forward 至 `59e6199`），保留原 owner 历史，未 squash/rebase/reset/force。写域遵循 `tasks/S1-final-paper-cli-memory-plan.md`；OpenAPI/迁移/入口/锁由 W0 单一 owner，本轨仅 `internal/attention/{domain,app}` 除 contracts.go、`internal/adapters/{importers,distillers,objects}`、`internal/adapters/sqlite` 除 local_agents*.go、`extensions/paper` 除锁、`docs/acceptance/S1-paper.md`、`tasks/S1-final-W1.md`。开发客户端 OpenCode / DeepSeek V4 Pro（实际观察，不声称 Codex 模型）。

## 交付

- `importers/scholar.go`：Crossref/EuropePMC/arXiv credential-free 检索；官方 HTTPS host allowlist + redirect 再检；命中永远 `abstract_only`；网络/解析错误保留 `required_action`（`retry_paper_request_or_check_network` / `check_paper_search_provider_response`）。
- `importers/paper_web.go` + `paper_url` 适配器：`PaperWeb` 走受控代理但先校验目标公开（IP 字面量 + 解析，拒绝 loopback/private/link-local/CGNAT/multicast/unique-local/IPv4-mapped）；HTTPS-only、无凭据/显式端口、16MiB 上限、redirect 再检。`paper_url` 复用 `InspectPaperHTML`，全文仅当公开完整正文；`abstract_only` 只入明确元数据；`paywall`/`restricted` 拒绝、无绕过；原文 HTML 附件入 objects，sourceKey/digest/revision 真实，经既有 job 幂等去重。
- `attention/domain/paper_search.go` + `attention/app/paper_search.go`：`PaperSearcher` 端口 + `Service.SearchPapers`（human-only、不持久化），`importers.Scholar` 实现端口。
- `extensions/paper/`：可加载 MV3（manifest/background/popup/extract + vendored Readability），`activeTab` 逐次点击、显式复制快照、无 host_permissions/daemon/新锁。content_state 全链对齐 `readable_fulltext|abstract_only|paywall|restricted`。

## 真实观察与首次失败

- 检索三 provider 各 5 hits PASS（Crossref 首条命中图题 doi:10.7717/peerj-cs.3829/fig-3，为真实 Crossref 数据非本轨缺陷）。
- paper_url live：PLOS 全文 66379B `public_html_fulltext`、ACL 摘要 1859B `public_html_abstract_only`。
- 首失败保留：EuropePMC `pubYear` 字符串反序列化失败（改 `europePMCInt`）；`paper_url` 首跑本地代理被 `publicOnlyDialContext` 误拒（`proxyconnect ... non-public address: 127.0.0.1`，改 `paperProxy` 先校验目标）；初审 popup.js `InjectionResult[]` 解构得 undefined（改 `injection.result`）。均未改写为成功。

## 验证

`go test ./internal/adapters/importers ./internal/attention/...` PASS；`node --test paper-dom.test.mjs` 7 PASS；`node extensions/paper/build.mjs` + `node --check` PASS；`GOFLAGS=-p=1 pnpm check` PASS/exit0（19 包、237 契约、14 API 合成、61 前端、TS/lint）。live 检索/paper_url 均 opt-in PASS，无模型/媒体调用，旧 UNKNOWN 不重发。

## 剩余 / Handoff

搜索端点 `POST /api/v1/paper/search` 与 contracts.go 接口由 W0 定稿（本轨签名就绪，按 W0 微调）；插件→应用同源审阅/入库 transport 归 W0/W3（W3 Dispatch 现 failed，DTO 已发 W0 与 ROOT）。插件真实安装/加载→读取→粘贴审阅→API 持久→重启闭环属 W0/W3 集成验收。PDF 全文（官方 arXiv 除外）不自动获取；`license`/`published_at`(现 `year` int) 待契约定稿。多站点“大部分 PDF 验收”不声明。
