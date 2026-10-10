# S1 收尾 W1：论文检索、正文提取与独立 MV3 插件

2026-10-10。本轨（`s1-sync-attention-1010`）按 `tasks/S1-final-paper-cli-memory-plan.md` 完成论文公开学术检索、通用正文提取和可安装独立插件，普通合入 ROOT 精确基线 `8277667`（`origin/s1/attention-materials-20261009`，fast-forward 至 `59e6199`，保留原 owner 历史）。开发客户端 OpenCode / DeepSeek V4 Pro（按当前客户端实际配置观察，不声称 Codex 模型）。搜索/当前页/批量决定已由用户确认，记忆/CLI 全覆盖两项仍 PENDING，不代选、不据此降低验收。

## 完成内容

- **公开学术检索** `internal/adapters/importers/scholar.go`：Crossref、EuropePMC、arXiv 三个 credential-free 官方元数据索引。仅公开 HTTPS、无凭据/无 cookie、严格官方 host allowlist + redirect 再检、分页/查询/响应大小上限。命中永远 `abstract_only`，绝不把搜索条目当正文入库。`Scholar.SearchPapers` 实现应用端口 `PaperSearcher`；`Service.SearchPapers`（human-only、不持久化）已组装。
- **通用正文提取 + 导入** `internal/adapters/importers/paper_web.go` + `paper_url` 读取适配器：`PaperWeb.FetchHTML` 走受控代理但先校验目标为公开地址（IP 字面量 + 解析地址，拒绝 loopback/private/link-local/CGNAT/multicast/unique-local，IPv4-mapped 亦覆盖），HTTPS-only、无凭据/无显式端口、redirect 再检、16MiB 上限。`paper_url` 复用既有 `InspectPaperHTML`（summarize0.25.1 的 Readability0.6.0 + linkedom），全文仅当页面公开且具备完整学术正文结构；`abstract_only` 只入明确标注的元数据；`paywall`/`restricted` 拒绝为 `evidence_missing`，无绕过。sourceKey 用 `doi:`/`arxiv:`/canonical URL，经既有 `ImportMaterial`/job 幂等去重与 `NextMaterialRevision`，原文 HTML 作为附件入 objects。
- **content_state 对齐 W0 契约**：`readable_fulltext | abstract_only | paywall | restricted`，Node DOM 核心、浏览器 extract、Go 适配器与搜索命中统一。
- **独立可安装 MV3 插件** `extensions/paper/`：`manifest.json`（`activeTab`+`scripting`+`clipboardWrite`，无 host_permissions、无 daemon）、`background.js`（自校验发送者，不信任外部）、`popup.html/js`（逐次人点击当前页→注入 vendored Readability+extract→展示标题/摘要/身份/状态→显式复制快照）、`extract.js`（浏览器原生 DOM + vendored Readability）。`node extensions/paper/build.mjs` 产出可加载 `dist/`，复用已复制的 Readability 与 MIT/Apache 许可，无新框架/新依赖/新锁。应用侧同源审阅/入库 transport 归 W0（externally_connectable + CSRF），插件 token 不当人类批准，插件不自动读任意站点/批量。

## 实际公开样本观察（本轮新功能必要动作）

| 来源 | 实际结果 |
|---|---|
| Crossref `attention mechanism transformer` | 5 hits，首条 `doi:10.7717/peerj-cs.3829/fig-3`（query 命中图题，真实 Crossref 数据） |
| EuropePMC 同 query | 5 hits，首条 `doi:10.7717/peerj-cs.1928` |
| arXiv 同 query | 5 hits，首条 `arxiv:2206.03003` |
| PLOS `10.1371/journal.pdig.0000514`（paper_url） | 全文 66379 bytes，`public_html_fulltext`，source_key `doi:10.1371/journal.pdig.0000514` |
| ACL `2024.acl-long.1`（paper_url） | 摘要 1859 bytes，`public_html_abstract_only`，source_key `doi:10.18653/v1/2024.acl-long.1` |

`paywall`/`restricted` 分类由 `paper-dom.test.mjs` 7 项回归覆盖（含“access restriction never produces full paper evidence”），未实时抓取付费墙页面（不绕过付费墙）；PDF 候选只作为链接展示，官方 arXiv PDF 走既有 `arxiv` 适配器（旧成功不重发）。

## 首次失败保留

- 首次 live EuropePMC：`pubYear` 实际为字符串，`json: cannot unmarshal string into int`；改为 `europePMCInt` 兼容数字/字符串，随后 5 hits PASS。
- 首次 `paper_url` live：受控本地代理被 `publicOnlyDialContext` 当作目标拒绝，`proxyconnect tcp: paper page host resolves to a non-public address: 127.0.0.1`；改为 `paperProxy` 先校验目标再 `ProxyFromEnvironment`，真实全文/摘要复验 PASS。首失败日志与旧错误保留，未改写。
- 初审 `popup.js` 用 `const [{ snapshot }]` 解构 `chrome.scripting.executeScript` 返回的 `InjectionResult[]` 会得 `undefined`；改为 `const [injection] = ...; return injection.result`。manifest description 原写 “Metadata only” 与真实正文矛盾，已改为明确“公开正文可选、默认仅元数据”。

## 验证与命令

- `go test ./internal/adapters/importers ./internal/attention/...`：PASS（检索解析/校验、paperSource 三类状态、SearchPapers 授权、IP/URL 校验）。
- `node --test internal/adapters/importers/paper-dom.test.mjs`：**7 PASS**。
- live 检索（`ASTROCYTE_PAPER_SEARCH_LIVE=1`）与 live paper_url（`ASTROCYTE_PAPER_URL_LIVE=1`）分别 PASS；未调用模型/媒体，旧 UNKNOWN 不重发。
- `node extensions/paper/build.mjs`：PASS，`dist/` 为可加载 MV3（manifest+脚本+Readability+许可），`node --check` 三个脚本通过。
- `GOFLAGS=-p=1 pnpm check`：**PASS/exit0**（gofmt/vet/test、19 包边界、237 契约例/生成一致、14 真实 API 合成、TS/lint、61 前端、diff；EventV1 既有未用警告保留）。整仓 `pnpm check`/`build` 由 W0 最终一次集成，本轨不并发全仓（主机 pagefile/GC 受限）。

## 真实剩余与未执行

搜索 HTTP 端点 `POST /api/v1/paper/search` 及 contracts.go 接口由 W0 最终定稿（已回执 handoff，本轨 `Service.SearchPapers`/`PaperSearcher`/`importers.Scholar` 已就绪，签名按 W0 定稿微调）；插件→应用同源审阅/入库 transport 由 W0/W3（W3 当前 Dispatch failed，DTO 已附 W0 handoff 与 ROOT）。个人 Chrome 未装插件、未读取个人标签/凭据；插件实际加载需隔离浏览器真实安装→读取公开论文→粘贴/审阅→API 持久→进程重启，属 W0/W3 集成验收。PDF 全文（官方 arXiv 除外）未自动获取，仅候选链接；`license` 字段未采集；`published_at` 现以 `year`(int) 表示。多站点“大部分 PDF 实际验收”不声明；家族识别仅目标提示，非站点全覆盖。

## 给 W0 的 ports/data 建议

- 检索端点建议 `POST /api/v1/paper/search`（human session+CSRF，Idempotency-Key），返回 `PaperSearchResult{schema_version, items[], warnings[]}`，`items[]` 即本轨 `PaperSearchHit` 字段，`content_state=abstract_only`。
- 导入沿用 `POST /materials/imports`，`kind=paper`，`adapter=paper_url`（或 `arxiv`），`source_locator=source_url`，服务端重新抓取+提取，`source_key` 去重、新 revision、16MiB HTML 上限。
- 插件 `dist/` 无新依赖/新锁（复用已复制 Readability0.6.0 + 浏览器原生 DOM）；如 W0 需要把插件 dist 纳入发布，`extensions/paper/dist` 目前 gitignored，是否随仓库发布由 W0 决定，无需改动根 lock。
