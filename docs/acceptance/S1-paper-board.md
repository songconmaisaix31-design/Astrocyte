# S1 论文检索、进度与前端 DTO 交接（本轮 W3，已消费 W0/W1 最终单一权威契约）

本轮（`run_8c1696bb815a`，W3，两个新 Dispatch）消费 W0 **`e39f8f1`/`662887e`/`cc211ce`**（单一权威 paper/progress DTO、`paper_snapshot` 进 adapter 枚举、progress infer 去 body `operation_id`）与 W1 **`0b189e5`**（`paper_snapshot` 离线摄取 adapter），替换上一 Worker 对齐 `b8dfe67`/`d89daa1` 的过期假 DTO，不再 mock 自想 DTO 或旧 `id/source_type/availability` 双轨：

- **论文检索**：`GET /api/v1/papers/search?q=&provider=&limit=`（W0 契约），返回 `PaperSearchHitV1`（`source_key/provider/title/authors/year/venue/arxiv_id/doi/locator/abstract/content_state/pdf_urls`），不再有 `site`/`id`/`source_type`/`availability`/`already_imported`。检索只返回公开元数据（`content_state=abstract_only`）；原文可得性在勾选入库后由作业确定。勾选纯人工、可取消、批量入库由人显式提交且不全量。
- **批量入库**：逐条走既有 `POST /materials/imports`（`ImportMaterialRequestV1`，`kind=paper`，`adapter=arxiv|paper_url`，`source_key` 去重、新 revision），**每项用 `crypto.randomUUID()` 幂等键**（不再自建 FNV Hash），后端 source_key/内容去重负责重复工作；`Promise.allSettled` 只确认「提交被接受」而非「入库成功」，失败汇总为单一可操作错误，成功作业在处理队列可查；不新增 scheduler/cache 或双轨 `/papers/import`。
- **插件快照复核**：消费 W1 `paper_snapshot` adapter（`export_text`=整份已复核快照 JSON，`source_locator`=来源 URL，服务按 DOI/arXiv/URL 重新推导 `source_key`，不做网络抓取）；`pdf_urls` 映射到 `paper_pdf` adapter（SSRF 安全公开 PDF 抓取，非 arxiv-only）。人类明确选择「HTML 正文 / 公共 PDF」，快照正文保留不再让 `paper_url` 重新抓网页；摘要/截断/付费墙不当正文，无个人浏览器自动读取。
- **项目进度**：消费 W0 `ProjectProgressV1`（`status/summary/percent/source/evidence[]/native_id/model/observed_at/warning/revision`，`source`/`observed_at` 可空，无 `operations`/`pending_operation` 内部泄漏），`ProgressEvidenceV1`（`source_path/kind/version/excerpt`，无 `freshness`）。`status=unknown` 呈现「未知」，`percent` 按 source 显示「人工记录 / Agent 估计 / 未知」，绝不伪造、不把模型百分比冒充人工输入。UI 接 `GET /local-projects/{id}/progress` 与 `POST .../progress/infer`（Idempotency-Key 即操作身份 + `files` 为**人类输入的相对路径**，默认 `STATUS.md`、1–8 行，受后端已批准目录校验；无 body `operation_id`）。
- CLI 原生操作沿用现有 `ManagedProjectsPanel`/`NativeProjectPanel`（start/read/resume/send/stop/observe/context_handoff），权限判断 `nativePermission.ts` 不变，不以看板即自动授权。CLI 完整覆盖与记忆隔离/全局/仅规划仍 PENDING（用户未答），未代选、未建 UI 伪工具。

## 分支与源码

- 分支 `s1-sync-ui-1010`，普通合入 ROOT 精确基线，保留原 owner 历史，无 squash/rebase/reset/force。
- 本轮 UI/测试源码为一次精确提交（SOURCE），REPORT 为随后仅更新 `docs/acceptance/S1-paper-board.md` 与本任务报告的提交，精确 SHA 通过 Orca 交接。
- 独占写域遵守 `web/src/`（除 `api/`）、`web/public/`、`web/e2e/`、本文件与 `tasks/S1-final-W3.md`；未改契约、生成客户端、迁移、锁或入口（`web/src/api/` 由 W0 单一 owner 生成 `searchPapers`/`getProjectProgress`/`setProjectProgress`/`inferProjectProgress` 包装后替换本地 seam）。

## DTO 交接（供 W0 集成阶段替换本地 seam）

前端已按 W0 `e39f8f1` 契约形状实现本地视图（`paperSearch.ts` 的 `PaperSearchHit`/`PaperSearchResult` 精确镜像 `PaperSearchHitV1`/`PaperSearchResultV1`；`progressPresentation.ts` 的 `ProjectProgress`/`ProgressEvidence` 精确镜像 `ProjectProgressV1`/`ProgressEvidenceV1`），并直接消费真实路由。W0 生成客户端后只需：

1. **论文检索**：`GET /api/v1/papers/search`（query `q`/`provider`/`limit`）→ `PaperSearchResultV1`。前端 `paperSearchClient.searchPapers` 已按此调用；W0 可替换为 `attentionApi.searchPapers(...)`。
2. **论文入库**：沿用既有 `POST /materials/imports`（`ImportMaterialRequestV1`，`kind=paper`，`adapter` 按站点 `arxiv` 或 `paper_url`，`source_key`=hit.source_key 去重、新 revision）。`paper_url`/`paper_pdf`/`paper_snapshot` 均已进 W0 生成 adapter 枚举（`662887e`），前端 `PluginSnapshotReview.tsx`/`paperSearchClient.ts` 的过期 cast 与「未进枚举」注释已删除。
3. **进度依据**：`GET /api/v1/local-projects/{id}/progress` → `ProjectProgressResultV1`；`POST .../progress/infer`（body `InferProgressRequestV1`{`files`}，Idempotency-Key 为操作身份）→ 同。前端 `progressClient.getProjectProgress`/`inferProjectProgress` 已按此调用；W0 可替换为 `localProjectsApi.*`。
4. **记忆（W2，待答决定后）**：列表/搜索 → 时间线 → 详情/按项目 scoped 配置/遗忘；本轮未建 UI、未默认全局或原对话采集。

## 验证

| 验证 | 结果 |
|---|---|
| `pnpm --dir web typecheck` | PASS/exit0 |
| `pnpm --dir web test` | 82 PASS（论文检索 6 + 插件快照 10 + 进度 5；既有其余保持） |
| 定向 `eslint`（新增/改动文件） | PASS/exit0 |
| `pnpm --dir web build` | PASS/exit0（101 模块） |
| `playwright test paper-search.spec.ts plugin-snapshot.spec.ts progress.spec.ts`（1920/1280） | **28 PASS**：检索 GET 契约载荷渲染、示例检索选择/取消 0 写、501 诚实「检索未完成」、移动 390 无横溢；插件快照正文保留+`paper_snapshot`、非 arxiv `paper_pdf`、付费墙无绕过、**同 URL 正文变化产生新幂等键（无 409）回归**；进度 GET/POST infer 带 Idempotency-Key+`files` + 未知不伪造百分比 |
| `playwright test --config e2e/paper-snapshot.config.ts`（`ASTROCYTE_TEST_PAPER_SNAPSHOT=1`，真实浏览器闭环） | **1 PASS**：真实 fulltext 快照粘贴→`paper_snapshot` 正文入库（revision1 browser_snapshot_fulltext，source_key=doi，正文>10000 字符，`paper-snapshot.json` 附件对象级等于界面原始提交的快照）→ 同 body 去重（revisions 仍 1）→ 同 URL 改正文生成 revision2 → fresh API 后 content_digest 一致 → 1920/1280/390 无横溢 + 键盘 Enter 开/关 |

## 未完成与真实限制

- **真实浏览器闭环已通过**（见上表最后一行，保留库 `%TEMP%/astrocyte-s1-HgzEDq` 供 root 只读）；但元数据真实搜索（`GET /papers/search` 返回 provider 元数据）与 W1 DOI 精确查找仍由 W1 补，不冒充所有站全文可入库。
- `paper_pdf`（abstract 快照 PDF 抓取）为在线抓取，本轮未触发；本地已取 PDF ≠ 在线抓取，不据此宣称在线 PDF 全文。
- CLI 完整覆盖范围与记忆隔离/全局/仅规划两项仍 PENDING，未代选、未建对应 UI。
- 未合 main/远程 CI、未更新个人 5173/8787、未安装插件/读个人 Chrome；无模型调用/旧 UNKNOWN 重放。真实数据三尺寸截图在本工作树 `web/test-results/`。

---

# S1 W3 项目总览与人类流程验收

本报告只覆盖本轮已授权、已实现的项目看板与现有入口。论文搜索、当前页提取、插件权限与进度推断仍为 **ASKED PENDING**；没有据此选择产品方案、冻结接口或宣称完整 S1 交付。

## 分支与源码

- 分支 `s1-sync-ui-1010`，最终 UI/测试提交 `2aa57e264aa56a7d7f271c5e087bc540f3f862ab`，消费 W2 返修后的最终 SOURCE `90cee4eda3c5f069325751064ecc944a303f6f77`，均已 push。REPORT 为提交本文件及 W3 任务记录的随后报告提交，精确 SHA 通过 Orca 交接。
- 精确根基线 `087127d30fc505f8fe25c52142a174f5842e2bef` 普通合并；W0 契约/HTTP/helper 与 W2 聚合/存储按发布 SHA 普通消费，最后已消费 W2 `be3cbfdbf536053f196ae5c48afad5fbf48138c2`、W0 组装 `967ed09eb50c988fbd9e04c99f48e91a4c023f47`。没有 rebase/squash/reset/force push 或越域业务编辑。
- 固定开发客户端为本 Orca Codex 会话，无额外 Agent。未自行切换模型；会话未暴露可核实的精确模型标识，不推测记录。

## 完成内容

工作区实际入口改为真实项目总览。后端 canonical ID 决定卡片身份，同一项目多客户端/多目录不增加重复卡。项目、来源文件夹、已记录客户端、已知活动时间与归档分别统计；支持平铺、平台组合、分组、时间线，按名称/目录/备注、自由人工意愿、客户端、组、来源、活动状态与归档筛选。页顶与看板搜索取交集，“清除筛选”同时清除两处输入，搜索文案按当前页实际数据范围显示。

项目详情优先呈现备注/下一步、复盘、分组、自由人工意愿与可恢复归档。SQLite 人工记录使用真实 API 与 revision；保存后采用服务器版本，外部更新保留未保存输入并明确提示。意愿不转成完成率，来源活动不转成人类意愿，完成度保持未知。根目录、来源、创建时间、真实记录活动、贡献客户端和仓库观察可回查；权限/CLI/原生/源码仓库工作流仍在“项目接入、权限与原生操作”入口，未增加自动控制或授权。

默认无 fixture 回填；加载、空、错误、旧缓存与 partial 来源明确显示。布局复用现有 Go/React/SQLite/OpenAPI、查询层、抽屉与官方图标，实际查看了用户参考图片；手机详情长目录换行有界，键盘 Enter 打开、Escape 关闭并恢复卡片焦点。

## 验证与范围

| 验证 | 精确范围与结果 |
|---|---|
| `pnpm --dir web typecheck` | 首次展示层 PASS；最终布局 `7373922` PASS；`acbfc43` build 含 tsc PASS |
| `pnpm --dir web test` | `c07fb11` 前端 61 PASS，其中项目展示模型 3 例验证去重、活动/意愿分离与筛选非破坏性；后续仅布局及导航测试变更，没有重复不变单测 |
| `pnpm --dir web build` | `c07fb11` 与最终生产布局 `acbfc43` PASS；`acbfc43..2aa57e2 -- web/src web/public` 无差异 |
| 定向 `eslint` | 已修改 UI、呈现模型、表单与浏览器测试 PASS；最后测试修正是标题选择器的精确修复 |
| 实际浏览器 `--config=e2e/project-board.config.ts` | **固定 `c07fb112ffd2497d4a913540f6836793e60a13db` 1 PASS，47.4秒（测试46.2秒）**。真实本机来源、真实 API、SQLite 保存/复盘/分组/意愿/归档/恢复、刷新后保留、同库新进程冷重启、四视图无重复、键盘焦点与 1920/1280/390 无横溢；没有模型/媒体/克隆或原生运行 |
| 最终布局实际预览 | **固定 `acbfc432aceec14fcafe3ad37d3845b994f450ef` 1 PASS，24.2秒**，1920×1080、1280×720、390×844；实际项目 API、非 mock，仅 preview GET，未重放人工字段/模型任务。首卡顶部桌面604.7px、手机769.6px；桌面明确断言首卡顶部至少80px处于初始视口。来源详情真实 partial，无虚构完成率 |
| 定向导航/错误传播 | **固定 `2aa57e2` 4 PASS，22.2秒**，1920/1280。工作区项目总览→接入设置→Agent会话；错误注入 `missions` 后任务、工作项、产物引用各自在实际面板/标签页显示错误。此项是受控接口失败回归，不冒充实际服务故障或 task_live |

实际浏览器执行设置：`ASTROCYTE_TEST_PROJECT_BOARD=1`，`ASTROCYTE_BOARD_REUSE_OWNED_TEMP` 与 `ASTROCYTE_BOARD_REUSE_APPROVED_ROOT` 均为自有 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-7wY4sg`；API18787/Vite15173，helper 只启动/停止自有进程。预览另设 `ASTROCYTE_BOARD_PREVIEW_ONLY=1`；导航重验使用 `ASTROCYTE_E2E_API_PORT=18787`、`ASTROCYTE_E2E_WEB_PORT=15173` 和 `playwright test navigation.spec.ts --grep 'navigates to 共同工作区 page|swarm propagates mission error'`。个人 Chrome、5173/8787、后台 daemon、全局配置与 Provider 凭据均未操作。

固定实际场次看到 **111来源目录 / 71 canonical 项目 / 2已记录客户端 / 43来源读取限制，status=partial**。这不是全盘覆盖、全客户端覆盖或真实项目完成度。实际测试还断言 `attention_jobs`、`local_agent_grants` 和受控项目权限记录在本次人工字段操作前后不变。

### 保留原失败与原件

1. 接入初次 TS RED：Windows 下 `ProjectBoardView.tsx` 与 `projectBoardView.ts` 大小写邻近冲突；后改为 `projectBoardPresentation.ts`。W0 helper 首缺 apiPort/webPort 声明造成另一次 TS RED，交唯一 owner 修正后 PASS；不改写原 RED。
2. `s1-board-live-20261010` 首实际场次 **FAIL / 4.5分钟**：可访问名称混入提示/选项，HMR 修正后继续保存、刷新、冷重启，最终390长目录详情按钮横溢。该场不能算固定源码验收；原 SQLite、日志、截图、`error-context.md`、`trace.zip` 均保留。独立固定 `c07fb11` 的 PASS 作为后续证据。
3. W0 `967ed09` 默认回归 **158 PASS / 4 FAIL / 22 SKIP / 7.7分钟**（W0交接，不冒充 W3 全套执行）：旧导航标题和折叠任务错误面板断言。W3 首修后定向 **2 PASS / 2 FAIL / 13.5秒**，暴露原数3遗漏产物专用标签；改按三处语义检查后的第二场 **2 PASS / 2 FAIL / 23.2秒**，产物标题选择器漏“引用”字样。原失败及 traces 均保留；最终4PASS不覆盖历史、不称整套重新通过。

本工作树本地原件：

- `web/test-results/s1-board-live-20261010/`：首失败 trace/截图/上下文；日志 `%TEMP%/astrocyte-w3-board-live-20261010.log`。
- `web/test-results/s1-board-final-live-20261010/s1-project-board-actual-pr-f19e0-resh-and-owned-cold-restart-chromium-owned-board/`：固定c07实际 `actual-snapshot.json`、`actual-after-restart.json`、三尺寸看板及详情 PNG；日志 `%TEMP%/astrocyte-w3-board-final-live-20261010.log`。
- `web/test-results/s1-board-final-compact-20261010/s1-project-board-actual-pr-f19e0-resh-and-owned-cold-restart-chromium-owned-board/`：最终布局实际来源快照、1920/1280/390截图；日志 `%TEMP%/astrocyte-w3-board-final-compact-20261010.log`。
- `web/test-results/s1-board-navigation-{retest,final,verified}-20261010/` 与 `%TEMP%/astrocyte-w3-board-navigation-{retest,final,verified}-20261010.log`：两次 RED 与最终定向4PASS原件分开。

## 未完成与交接

- 根发现 W2 汇总未纳入已观察 Git 提交时间，native-only 项目活动被错误标未知；上述快照7/71已知活动是修复前观察。W2 精确 SOURCE `cd201962997e69bb4772d73a6d8752874085c3ef` / REPORT `9319374f4a55cf88901e2ed77b439f45aaea1b89` 已普通消费为 `90cee4e`，修复为取已观察 Git 提交和既有 Orca 活动的最大时间，会话创建仍独立。已读 W2 原报告与5行业务修复：domain/app定向测试及 Go build PASS；真实既有缓存JSON通过实际应用服务重投影111根/71组，已知时间7→17、其余54未知，source调用0/write调用0、原JSON不变，**metadata adapter是模拟，不宣称真实人类字段/API/浏览器**。UI既有Git来源匹配/标签无需变更；`acbfc43..90cee4e -- web/src web/public`无差异。
- 根2026-10-10 22:12已实际查看并接受最终1280/390布局；22:17接受 W2 独占修复与 W3 4项定向浏览器通过，并明确 **W0负责最终集成的一次真实缓存HTTP GET，验证17已知活动及人类记录不变且不发现/不调用模型；W3可结算独立UI范围**。W3不冒充已执行该后续HTTP检查，也不再启动会必做启动发现的服务来重复证明；该最终全局检查仍由W0/root归属。
- 根已正式回复协调问题 `msg_de7eb50d3762`：同意W3不再启动服务，将已有真实UI与W2缓存只读重投影分开记录并结算独立UI；root负责必要个人预览升级后的实际GET与活动显示确认，论文Search/插件/进度选择继续待答。该协调问题已关闭，W3没有据时间经过替代答复。
- 论文搜索→来源→选提取/入库→继续整理、独立插件真实加载/入库：依用户对搜索/当前页/插件权限的待答决定，**NOT_RUN/PENDING**。W3未自行启用，不把 W1 DOM提取核心视作可安装插件。
- 进度推断依据仍未决定，完成度未知；自由人类意愿仅为已批准的人工记录。
- 冷重启为停止自有 API 进程并使用同库启动新进程，未执行操作系统重启。最终小布局变更没有重跑不变真实人工字段流程；最终产物对应固定版本检查分别列明。
- 未合 main、未运行远程 CI、未更新个人预览/安装插件；根主控负责最终普通集成、全局检查与人类视觉接受。本报告不是完整S1或所有待答目标的接受证明。
