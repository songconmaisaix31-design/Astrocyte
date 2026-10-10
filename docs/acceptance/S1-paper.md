# S1 W1 论文与容量：独立切片交付

2026-10-10。整体论文 Search / 独立插件目标 **未完成**；本报告只接纳已获准的容量和解析核心。用户尚未回答搜索/当前页/两者及插件访问/动作范围，未冻结相关 API、权限、manifest、配对或自动读取行为。主控要求先发布此可独立集成切片，W1 继续作为后续返修 owner。

## 完成内容

- Selected-text 输入改为完整 **512KiB UTF-8**，计算正文、schema、指令、引用 JSON；W2 完整 native authority/context 包装再次预检。超限拒绝、无截断、进处理前为已知失败；输出仍 **128KiB**，历史及项目许可没有随容量扩大。
- 直接 Codex 整理也检查指令/输入/schema 的聚合容量；结构化输出和推荐输出单独拒绝超过128KiB。
- 新增独立 `InspectPaperHTML`：对明确提供的 HTML 字节使用现有 summarize0.25.1 / core0.25.1 的 Readability0.6.0 和 linkedom；不抓网页、不调用模型、不导入、不迁移数据库。泛用 citation/DC/JSON-LD 元数据、DOI/arXiv身份、来源版本、作者、摘要、PDF候选保留。
- Readability 克隆提取；全文证据只来自实际保留的正文，排除摘要章节、导航、aside、hidden/script/相关区域；公开正文旁的登录机构控件不能让公开全文变为受限。只有标题/摘要/PDF链接的页面不返回正文；受限与未知明确。
- 独立浏览器解析资产可构建，复用现有依赖，无新增锁文件。由于产品决定未答，`extensions/paper/dist` **不是可安装插件**；最终 MV3权限/action/应用接入和实际加载未执行。

## 上游来源与许可

已实际读取 [官方扩展 README](https://github.com/steipete/summarize/blob/main/apps/chrome-extension/README.md)、[官方扩展文档](https://github.com/steipete/summarize/blob/main/docs/chrome-extension.md)，并检查 actual `extract.content.ts`、`page-readability.ts` 及 LICENSE；研究 pin `560197cd4b580554cccf648744c592e867b43bb5`。复用 DOM clone / Readability片段保留 MIT Peter Steinberger2026原文；Readability Arc90/Mozilla Apache版权说明及完整Apache许可均保留，见 `extensions/paper/THIRD_PARTY.md`。研究 main pin 与实际安装0.25.1 pin 分别记录，没有把两者混称同一版本。

## 实际公开样本观察

以下是本轮明确 URL 的研究快照检查，未经浏览器资料选择/入库，**不是浏览器到应用验收**。原HTML和JSON在 `%LOCALAPPDATA%/Temp/astrocyte-paper-W1-20261010`，未放个人资料库。

| 官方来源 | 实际提取 | 限制 |
|---|---|---|
| [ACL](https://aclanthology.org/2024.acl-long.1/) | HTTP200，标题/作者/DOI；摘要1685bytes，1PDF候选 | 页面是摘要与PDF入口，无正文 |
| [PMLR](https://proceedings.mlr.press/v202/dettmers23a.html) | HTTP200，标题/作者；摘要1298bytes，1PDF候选 | DOI未知；无正文 |
| [CVF](https://openaccess.thecvf.com/content_cvpr_2016/html/He_Deep_Residual_Learning_CVPR_2016_paper.html) | HTTP200，标题；摘要1288bytes，2PDF候选 | DOI未知；正文PDF/补充PDF未获取 |
| [PMC](https://pmc.ncbi.nlm.nih.gov/articles/PMC11135672/) | HTTP200，摘要2089bytes，正文66915bytes | 仅当前公开HTML正文，图/方程/PDF完整保真未验 |
| [PLOS](https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514) | HTTP200，摘要2089bytes，正文64008bytes | 仅当前公开HTML正文，图/方程/PDF完整保真未验 |

PMC/PLOS共享 `doi:10.1371/journal.pdig.0000514` 身份，原URL与各自正文保留。代码识别 arXiv、PMC、bioRxiv/medRxiv、PLOS、Nature、Springer、Elsevier、IEEE、ACM、Wiley、Science、MDPI、Frontiers、OpenReview、ACL、CVF、PMLR及DOI等家族；**家族识别不是所有站点已适配或大部分网站验收通过**。其余族未实测，代理DNS/付费墙/挑战页面/仅PDF等限制未绕过。

## 容量实际源复用

仅只读既有 `%LOCALAPPDATA%/Temp/astrocyte-s1-4pmMWO/data`，没有重新提取媒体、下载旧论文、执行旧模型或重发 UNKNOWN。

- `arxiv:2504.16054` 当前rev1，**95314bytes**，SHA256 `05acdc763e444fa7b9762df6dc2c879867d6a2e38d42884dcf3b80e07cf89d64`。
- `https://www.bilibili.com/video/BV1PReT6EEqR/` 当前rev2，**32908bytes**，SHA256 `eb1106b97e3db4ddb6465ae066d1efe29affb809aa4841d52014ede3ab45a3fc`。
- 实际 distiller 组装调用者 prompt **135051 UTF-8 bytes**（含原文/ref/schema/指令）；逐份反序列化核正文、digest、引用未变。最终 native envelope预检PASS；输出131072bytes不变。该测试 capture 后取消，**没有模型调用，不是联合模型结果验收**。恰好512KiB及+1byte/多字节、包装额外开销分别有单测。
- 原模型 UNKNOWN不重发；主控明确仅授权 W0 在集成后做 **1次新增联合模型作业**，W1不调用模型；W0必须复核当前项目/CLI/来源许可且记录实际API job、最终请求及结果。

## 验证与保留失败

- `node --test internal/adapters/importers/paper-dom.test.mjs`：最终 **7 PASS**。最初 **3 PASS/1 FAIL**；根复核后的第一新回归 **5 PASS/1 FAIL**；随后修复的green单独记录，见 `tasks/S1-paper-W1.md`，不改写首失败。
- `go test ./internal/adapters/importers -run TestPaperDOM -v`（设置 `ASTROCYTE_PAPER_CAPTURE_ROOT`）：安装pin/隔离、3真实元数据样本、2真实全文/跨host同DOI **PASS**。
- `go test ./internal/adapters/distillers -run 'TestCombinedPrompt|TestActualKnownPaperVideo' -v`（设置 `ASTROCYTE_CAPACITY_SOURCE_ROOT`）：合成212533bytes和真实135051bytes完整组装 **PASS**；未调用模型。
- `go test ./internal/adapters/importers ./internal/adapters/distillers`：PASS；实际源/capture检查默认opt-in，不能把未设环境时的SKIP记为实测。
- `node extensions/paper/build.mjs`：PASS，仅解析资产，不是插件加载。
- `pnpm check`：首轮PASS；最终设置上述两个opt-in根的检查也 **PASS/exit0**，Go test/vet/mod、19包边界、237契约例/生成一致、14API合成用例、TS/lint、58前端单测及diff。API合成不混入论文browser验收。日志 `%LOCALAPPDATA%/Temp/astrocyte-paper-W1-check-{final-}20261010.log`；首场与最终分别保存。
- `pnpm build`：PASS，Go后端及React/TS/Vite生产构建；日志 `%LOCALAPPDATA%/Temp/astrocyte-paper-W1-build-20261010.log`。

## 真实剩余与未执行

Search provider/检索选择；最终可安装插件、权限动作范围、浏览器真实加载及同源人工审阅/入库；多站点选取/全文PDF提取；SQLite资料库冷启动和版本/别名去重；联合真实模型结果均未完成本轨验收。现有快照DOI identity不替代SQLite去重。个人Chrome/插件未变；不保存cookie/token/message/globalstore，不自动全库导出。W3仍拥有18787/15173隔离app-browser位，后续需owner handshake；W0拥有共享契约/迁移/路由/入口/锁。
