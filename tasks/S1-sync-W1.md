# S1 sync W1 — 选定浏览器与公开列表同步

## 当前交付：选定求职收藏夹后端（2026-10-10，阶段15）

分支/工作树 `s1-sync-attention-1010`；业务 SOURCE **`38f4f4dff0e9884c77daf1c0ca333cd96d93f1e5`**，已push，包含阶段13 `2993ad0` 与阶段14 `3268943`。先普通合入主控计划 `6e84a92`，再普通消费 W0 `fde698d` / `2d4f759` 的契约与010/011迁移；没有reset/rebase/force，也没有编辑共享契约、入口、迁移、UI、原AT01–04或队友写域。实际客户端按主控回执为 Codex 0.162.0 / GPT-6.1-Sol high fast。

- 已交付限定浏览器元数据桥、显式目录发现与SQLite缓存、真实身份/版本去重、目录GET无执行、默认公开与显式browser_selected分离，以及UNKNOWN同步不重发。`owner_id=self`只查宿主配置，不扫描账号；人工discover回传稳定owner/folder，只有该已发现收藏夹可绑定，当前账号变化拒绝新读取。固定同步不调用Agent，人工推荐/勾选后的正文链继续复用原有实现，没有新增Mission或框架。
- 实际限定抖音会话 `astrocyte-douyin` / profile `terpzafx`，已选求职folder `7694962768730068771`。独立第二场真实桥→Service→SQLite **PASS**（测试12.44秒/Go包22.486秒）：3条真实标题/简介，两个source_sync均attempts1，repeat所有metadata revision1且无重复，真正SQLite close/reopen逐字段一致；catalog固定身份保留，材料/沉淀/候选/非同步作业均0。只返回选定元数据，无媒体/模型/评论/认证字段；并非应用页面或OS进程重启验收。
- 现有startup hook仍一次、默认100观察上限；最新UNKNOWN补丁拒绝同source的启动同步和换新request key重发，原job JSON不变。定向 `TestUnknownSourceRead|TestStartupResumes` PASS（包4.123秒），初次测试误用了不存在的helper名而编译FAIL，已改用既有errorCode。最终 `go test ./...` PASS；`go vet ./...`、`go build ./...`、`node scripts/check-architecture.mjs`（19包）与`git diff --check` PASS。默认Go套件跳过显式外部live测试，不能把该绿结果称为浏览器/付费调用通过。
- 当前bridge只接受提供方确认完整、数量一致的已加载选定收藏夹，分块最多100；不调用页面加载回调、猜签名API或跨文件夹抓取。未完整加载的收藏夹明确EvidenceMissing并保留旧缓存；不宣称任意账号/大夹主动分页已支持。未连接、换账号、未打开选定夹与UNKNOWN均保留真实限制，错误不伪装成功空列表。科研0只是原页面目录观察，没有读取其正文或宣称其列表实测完成。

原件：`C:/Users/DW/AppData/Local/Temp/astrocyte-W1-selected-ctx_54629f75a8e9-second/` 内实际SQLite、catalog及repeat/restart JSON；`...-second.log`。同目录 `selected-browser-config.json` 仅含node_path=`C:/Program Files/nodejs/node.exe`、cli_main=`C:/Users/DW/AppData/Roaming/npm/node_modules/@jackwener/opencli/dist/src/main.js`、profile/session、已验证owner和folder_ids，已交root/W0，root负责验收后持久化及个人预览备份/升级。本轨未改个人库或integration目录。

首个整页state越界输出事实、首场实际metadata SQL审计FAIL与DB/log均保留，具体见阶段14；不复制侧栏私密文本。后续固定产品表达式和实际SQLite只保存所选元数据。原B站99+2、旧AT01–04和旧模型UNKNOWN均沿用，没有重跑付费/媒体。另在root提出消除重复验收前，本轨已经启动重复HTTP harness；其遗漏Foundation配置导致health panic，尚未进入任何新浏览器读取。收到root纠正后终止、删除仅本轨新增且未提交的测试文件，原`...-api.log`保留为停止的harness失败，确认该测试Go进程已退出；没有业务HTTP改动，不把它当作应用失败或通过。

**跨轨剩余操作：**W0组装宿主配置/能力/入口；W3独占真实应用discover→人工选夹→绑定→metadata sync/repeat/cache与OS进程冷重启/界面验收，root已授权该隔离场景，不调用推荐模型/媒体。本轨Service/SQLite与领域检查不替代它。root明确要求结束重复HTTP验收，只完成本轨领域修复与报告；应用返修仍交原owner。bound Chrome未关闭或解绑，按root指令继续供应用固定读取，W3不直接操作个人Chrome；需要新的个人标签操作先通过root协调。

## 继续阶段 13：选定抖音浏览器收藏夹（2026-10-10，历史开发记录）

本轮按主控完整互斥写域计划先普通合入 `6e84a92acadc7633c8d905ade246818b265df550`，保留原历史。主控实际核实本终端客户端 Codex 0.162.0 / GPT-6.1-Sol high fast；没有启动额外 Worker。主控负责 Chrome 插件安装，尚未显式交接浏览器，本轨没有读取个人 Chrome、Cookie、历史或其他标签。

已读取安装的 OpenCLI 适配器和官方 [collections.js](https://github.com/jackwener/OpenCLI/blob/main/clis/douyin/collections.js)、[browser-fetch.js](https://github.com/jackwener/OpenCLI/blob/main/clis/douyin/_shared/browser-fetch.js)。`douyin collections` 是 creator.douyin.com 的作品合集 mix_list，不是选定用户收藏夹；user-videos 会扩展下载地址/评论，也不用于本轮。主控随后将 OpenCLI 更新为 1.8.8，实际连接状态仍由主控交接后核实，不用旧 1.8.6/未连接观察冒充当前实测。

本轮先实现 `DouyinBrowser` / `SelectedFolderBridge` 的限定范围、严格元数据解析及 PublicListing 分派：宿主明确配置稳定账号 ID 和选定收藏夹 ID，HTTP 输入不能扩大允许集合；浏览器必须验证同一账号/收藏夹身份，缺字段、挑战、分页不前进、重复 ID 冲突和漏选文件夹均返回错误，零条只接受已核实成功空收藏夹。视频 ID 字符串保持精确；去重不改变实际观察行计数，没有正文/媒体/模型调用。该适配层的真实 OpenCLI transport 与实际网页结构仍待主控浏览器交接，当前不能宣称真实收藏同步完成。

`go test ./internal/adapters/importers ./internal/attention/app ./internal/adapters/sqlite` PASS（3.849s / cached / 6.160s），属于 contract_local 与真实 SQLite 回归，不是实际浏览器验收。W0 已接受 `access_mode=public/browser_selected`、稳定 owner-aware identity 和目录 GET 只读缓存，正在独占发布 contracts/迁移与 POST discover；W1 负责消费、缓存与后续同轨返修。原 B站99+2观察、原 AT01–04和旧 UNKNOWN 不重跑或改写。

### 阶段14：实际限定桥与缓存目录（继续中）

普通消费 W0 `fde698dfe924f0509f391228fd47cda7234d84f8` 和 `2d4f759e0ebf517fa3ffd55020c26d57c23b6415` 的 AccessMode、owner-aware identity、010/011迁移和人工 POST discover。旧 AccessMode 缺值读作 public；身份不可变；浏览器 binding 必须使用显式已发现的同账号/文件夹缓存，URL tracking 参数不造成新绑定。目录 GET 只读 SQLite，尚未发现返回 EvidenceMissing；`owner_id=self` 仅使用主控配置的本地 accessor 解析，缓存与回执保留真实 owner ID。人工 discover 在外部读取前查原回执，不持有 SQLite 写锁调用浏览器；失败保留原缓存。Agent不能触发发现，目录动作不产生资料、模型或媒体工作。

主控于11:32 UTC明确交接 profile `terpzafx` / bound session `astrocyte-douyin`。**第一条 OpenCLI state 返回整页并包含隐藏的非授权侧栏文本，超出所选收藏夹范围**；已立即停止整页state/extract/networkdump并向主控 escalation，主控确认保留错误事实，不复制私密内容。本报告不保存该输出。后续只针对选定DOM节点读取，稳定账号 `MS4wLjABAAAAA9cYVtOn_y6ULxB6xOkam6HvqsfhxDjBbXju2jnbkEJEoUtWCuO0fTbGpPLy5Fdt`、当前求职收藏夹 `7694962768730068771`、active true/count3已核实。

具体 `NewOpenCLISelectedFolderBridge(nodePath,cliMainPath,profile,session)` 只通过绝对 native Node 与 OpenCLI 主入口执行固定嵌入表达式；页面输入不能提供代码/命令。只读 user-info secUid、collection-navigation 指定ID的名称/数量、user-detail中匹配同folderID的scroll-list元数据；不读取全局store、Cookie、历史、私信或其他标签，不调用会修改页面state的fetchMoreData。不返回下载、评论、模型或原始认证字段，外部stderr/表达式不写公开错误。当前完整已加载列表支持有界100与精确字符串ID；部分/未加载完列表返回EvidenceMissing，不把未知分页伪造为完成。主动分页的固定只读API仍待实际观察，这是当前功能限制。

消费后 `go test ./internal/adapters/importers ./internal/attention/app ./internal/adapters/sqlite` PASS（9.492s/4.917s/19.556s）。`TestSelectedCatalogExplicitDiscoveryCacheRestartAndBindingScope` 实际SQLite/cache GET不执行桥、POST回执复用、绑定alias复用、未发现拒绝、断开后缓存保留、冷重启和Agent拒绝 PASS（测试0.29s/包8.659s）。首次 opt-in `TestSelectedDouyinBrowserLiveSyncRepeatColdRestart` 已实际通过当前folder3条读取、两次sync无重复/无revision变化、SQLite close/reopen逐字段一致，最后零额外作业审计测试误查不存在kind列，整场FAIL（12.06s/包20.596s）。首场DB和仅选定元数据保留 `%LOCALAPPDATA%/Temp/astrocyte-W1-selected-ctx_54629f75a8e9-first/`，日志同名first.log；已修正按既有JSON data.kind读取，独立复验待执行，不改写首FAIL。

W0负责持久非凭据主控配置/入口组装，W3负责最终真实API/新UI/进程重启验收；本阶段未做应用正文/模型请求，旧UNKNOWN不重放。Chrome未关闭/解绑，仍由本轨限定持有，后续现场使用需要主控协调。

## 历史交接（2026-10-10，继续阶段 12 后）

分支/工作树 `s1-sync-attention-1010`；当前业务 SOURCE `90c9e0419f6ef5e9b808fbd9fff6f6b734970ffe`，已push（真实metadata验证 `8c7292c`、失效行修复 `e1061eb` 均包含）。以下是已实现范围；后文按时间保留原阶段记录、首败及当时限制，不能把历史“尚未接线”当作当前状态，也不能把本地检查当作真实 S1 验收。

Root/W3 于 2026-10-10 08:46 UTC 交接真实集成续验 SOURCE `cd721692122c5e8f7820e34f25181058dd648cf0` 的 AT01–04 PASS，root 已独立复核同库；该 SOURCE 包含上述 W1 业务提交。四个新模型目标成功、共五条沉淀；视频已保存 Result 经真实本地 OS 发布失败、冷重启和人工 retry（attempts2）恢复，无重复模型调用；机器读取3次、人工读取不增加 Agent heat，later 保留。此为 root/W3 的实际任务与浏览器证据，区别于本轨 contract_local 检查；原 UNKNOWN 未重放，非空账号全链未通过。Root 默认 E2E 此时仍在进行，最终交接尚未接纳。

- Attention/Storage 已接应用和 SQLite：候选需同一固定资料版本的实际沉淀，满足来源/用途/下一步，并有关联主题/项目或明确开放问题；旧版本反馈不能替代当前版本。普通 URL 新表单复用既有 job，人工显式 refresh 才重新提取；固定导出换正文可新建 revision，A→B→A 复用旧 A 并保留 B head。官方 arXiv 论文导出的身份与现有 arXiv 导入归一，原件及 provenance 保留。
- 公开账号绑定区分真实平台身份、uploads/favorites、多账号与显式收藏夹；沿用上游 yt-dlp 的 B站公开 UID 列表薄桥及官方收藏夹接口。同步最多消费 100 条提供方元数据，持久保存 cursor/计数/部分缓存/失败；挑战不是空列表，失效行不导致整页失败。启动只排一次有界 metadata sync，无定时器；推荐使用选定已许可项目/CLI，先保存实际建议，人工选择后才排选中正文的既有 summarize 导入，不自动全量入库。
- 本轮真实公开收藏目录返回默认夹2356677875/count452与小夹3501892975/count2。默认夹实际匿名HTTP→Service→SQLite两轮each100observations/99unique条目/cursor6/has_moretrue，source URL别名绑定复用、未变metadata revision不变、真实冷重启逐字段一致；没有建议/选取/正文/模型作业。小夹真实attr9失效行有BV也不可选，原状态与identity保留。固定来源/人类/project/CLI/config/metadata-rev的相同推荐复用原job与结果，反序/newkey不能重复调用或绕过旧UNKNOWN；每请求当前权限检查不省略。
- 固定版本空间读取及 C 一跳关联受当前范围/权限约束；人工读取不增加 Agent heat。已知安全失败使用既有 job 的有界重试、期限及实际失败事件；旧 UNKNOWN 不自动重放。已返回并持久保存的 Result 遇到本地对象发布故障时，保留 OS 原因、提示修复存储并人工重试，不自动消耗修复等待期的次数；真 SQLite 重启后使用同一 Result，选定 CLI 的配置观察冷缓存也无需再调用模型，撤销许可/空间变化/引用移除仍拒绝发布。
- 通用选定 CLI 适配器通过现有 consumer port，实际新处理仍核验配置与权限；无虚构评分或 mock 反馈。组装完整 UTF-8 输入及输出各限 128KiB，超限明确已知失败、不截断；处理限时沿用有界 job，视频上限 30 分钟。全局状态不会仅因装配了工厂就声称 CLI 可用。跨来源扩大输入上限的决定仍待主控/用户，当前未放宽。

阶段12 SOURCE 已通过 `go test ./...`、`go vet ./...`、`go build ./...`、`node scripts/check-architecture.mjs`（19包）和 `git diff --check`；固定推荐输入/UNKNOWN/撤权/变化隔离的SQLite冷重启回归PASS。阶段11 opt-in真实HTTP同步/repeat/restart单独PASS，原件/两场DB保留 `C:/Users/DW/AppData/Local/Temp/astrocyte-W1-public-ctx_514362ae7294/`，首场误写100unique断言失败保留。默认Go套件的contract_local/真实存储测试不是paid/native/browser；本轨没有新的媒体/模型调用。W3独占实际metadata模型推荐/人工选择原BV与旧正文复用的浏览器现场，当前等候其结果及root最终接纳；领域返修仍归本Worker。

阶段 10 仅修复新 `attention.job_failed.cause` 丢失原适配器服务错误说明的问题：公开 job.Error/UNKNOWN/禁止重试不变，旧事件不补写/改写，不写入来源正文或凭据。W3 保留的首个真实选定项目 job `4GQ7GKDXHZIND7CAB7GIZU3W4V` 的旧 UNKNOWN/Result:null 已按其自有数据库只读确认，不能套已保存 Result 的恢复；不能凭 18 秒或 external_started 推导模型已接受，新目标成功不改写这个旧未知。

真实外部限制：指定 B站 UID `3494358764489275` 公开uploads的历史HTTP412 / -352 challenge未改写；历史收藏目录0也保留，但本轮两夹非空及metadata实际持久化已独立通过，完整推荐/人工选择链等待W3。Douyin `/user/self?showTab=favorite_collection` 需要登录，仍待真实公开身份且没有实际可用的公开profile列表transport；未读取cookies或绕过挑战。Root 报告 arXiv URL 获取受当前fake-IP/DNS环境限制；使用真实历史summarize论文导出时保留existing_json_export等实际provenance，不冒充本轮官方下载。原视频既有summarize_url真实receipt可复用，没有新媒体；额外 `2501.12948v1/v2` 原件仅获导入/版本检查授权，不能调用模型。原UNKNOWN、首败与未知效果保留；本轨不宣称全S1或完整账号验收完成。

共享contracts/migrations/HTTP/入口由W0独占，native CLI/权限桥由W2负责，UI/真实浏览器由W3负责；当前继续使用现有ProjectSpaceID的scope-only恢复路径，不增加第二套optional port或恢复接口。源码已按root指令补阶段11真实失效行与阶段12推荐API固定输入复用，现场领域问题返回本Worker修复，最终接纳由root决定。

## 历史公共适配阶段（保留首败与当时限制）

2026-10-10。工作树/分支 `s1-sync-attention-1010`，基线 `20437708e43201e352d6c6926902e1363fd2ad3e`。本机 Orca Codex；`orca orchestration worker-show --dispatch ctx_3dec09df71bd --json` 实际观察 projection.provider.model=`gpt-6.1-sol`，原生终端显示 high fast。仅修改本轨授权写域，没有应用模型调用、真实媒体转写、浏览器验收、凭据读取或宿主网络变更。

## 本阶段完成

- B站公开收藏夹 URL 严格识别；显式 UID 的公开收藏夹目录及显式 fid 分页视频元数据 HTTP 适配。无 cookie jar，不跟随重定向，固定官方 HTTPS API，限制响应大小；上下文取消保留。仅列出元数据，没有自动入库、正文提取或建议生成。
- 真实成功空收藏夹（`code=0`、`medias=null`、`media_count=0`、`has_more=false`）与验证页、登录拒绝、非零平台错误、缺字段分别处理。原始上游 ID 使用 `json.Number`，不经过浮点数；同一页面/跨页面相同条目折叠，冲突元数据拒绝任意覆盖。
- 抖音官方授权账号 `video.list` 响应解析：保留 opaque `item_id`、独立 `video_id`、精确 cursor 和 provider status。**未接授权 transport；不宣称抖音公开账号/收藏夹已可读。**
- 纯领域列表比较：完整成功列表区分 new/updated/unchanged；相同元数据不增加 revision；部分或冲突列表不替换旧缓存。完整列表的缺项只返回 observed missing，旧资料保留，不自行决定删除政策。该规则尚未接 SQLite、API 或同步作业。
- 主控审查后修正：已失效或非视频条目有 provider ID 时保留 unavailable reason 与空 locator；没有可信 ID 的条目返回明确 warning。可用视频继续列出，不伪造无效视频 URL，不因单条失效封锁整页。视频身份优先保留精确 type+numeric provider id，BVID 是正文来源 locator；平台缺 numeric id 时只使用有效 BVID。暂无公开 DTO 冻结。

## 实际公开诊断与来源

| 只读诊断 | 实际结果 | 解释 |
|---|---|---|
| `Invoke-WebRequest https://api.bilibili.com/x/v3/fav/resource/list?media_id=1103407912&pn=1&ps=20`（上游公开测试 URL 的 fid） | HTTP200/code0；media_count0/medias=null/has_more=false；标题“【V2】（旧）” | 真实成功空列表；不是非空账号同步验收 |
| 上游 UID3985676 `/x/space/wbi/arc/search?mid=3985676&pn=1&ps=30` | HTTP412 | 未签名公开账号上传列表被拒绝；不能推导登录后结果 |
| 上游抖音公开 user URL `MS4wLjABAAAAEKnfa654JAJ_N5lgZDQluwsxmY0lhfmEYNQBBkwGG98` | HTTP200/72914bytes；验证脚本，无列表数据 | challenge，不是空列表 |
| 同一公开用户 `/aweme/v1/web/aweme/post/`，count20/max_cursor0 | HTTP403 | 未绕过验证或提取 cookie；停止反复探测 |
| 指定 BV1PReT6EEqR `/x/web-interface/view` | code0，公开 owner UID1625561074 | 仅协调者允许的已有视频公开元数据诊断 |
| 上述 owner `/x/v3/fav/folder/created/list-all?up_mid=1625561074` | code0，公开 folder count0 | 没有可供真实非空收藏夹验收的示例；未绑定该账号 |
| 独立安装的 `%LOCALAPPDATA%/Astrocyte/media/yt-dlp-2026.08.19/Scripts/yt-dlp.exe --version` | 2026.08.19 | 只读工具启动，不下载媒体 |
| 同环境 Python `inspect.getsource(BilibiliFavoritesListIE._real_extract / BilibiliSpaceVideoIE._real_extract / DouyinIE._real_extract)` | 有B站收藏夹/空间提取；DouyinIE仅单视频，明确 fresh cookies challenge；无 Douyin user-list 类 | 不把单视频能力推导为账号同步 |

本轮阅读一手来源：

- [B站开放平台签名/OAuth 文档](https://open.bilibili.com/doc/4/8673959e-f7bb-56e6-6e68-d225f971b81b)：官方应用接口与公众网页接口有不同授权要求。
- [yt-dlp B站主源码](https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/extractor/bilibili.py)：BilibiliFavoritesListIE 的公开测试、`resource/list` / `resource/ids`、私有收藏夹 -403；空间上传列表委托 WBI 及分页面判断。
- [yt-dlp TikTok/Douyin 主源码](https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/extractor/tiktok.py)：DouyinIE 仅单视频。
- [抖音官方授权账号视频列表](https://open.douyin.com/platform/resource/docs/openapi/video-management/douyin/search-video/account-video-list/)：`video.list` 用户授权且需申请权限；公开账号视频元数据也不等于免授权收藏夹。
- [抖音官方企业账号授权说明](https://partner.open-douyin.com/docs/resource/zh-CN/mini-app/open-capacity/operation/business-account/bizaccount)：绑定、账号确认及能力授权是独立步骤。
- [OpenCLI B站原生适配说明](https://github.com/jackwener/OpenCLI/blob/main/docs/adapters/browser/bilibili.md) 和 [favorite.js](https://github.com/jackwener/OpenCLI/blob/main/clis/bilibili/favorite.js)：浏览器+已登录 Chrome+Browser Bridge 是明确前提；未选择或调用此组件。
- 现有 `summarize-media.mjs` 薄桥仍锁定项目 summarize/core 0.25.1，并调用上游 `fetchTranscriptWithYtDlp`。本轨尚未安装本工作树 npm 依赖（没有 node_modules），没有运行新版/全局 summarize 替代锁定依赖。

## 验证

已实际运行：

```powershell
& "$env:LOCALAPPDATA/Programs/go/bin/gofmt.exe" -w internal/adapters/importers/discovery.go internal/adapters/importers/discovery_test.go internal/attention/domain/listing.go internal/attention/domain/listing_test.go
& "$env:LOCALAPPDATA/Programs/go/bin/go.exe" test ./internal/adapters/importers ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite
```

四包首次 PASS（1.230s / 0.377s / 0.526s / 1.794s）；修正失效条目后四包再次 PASS（1.049s / 0.301s / cached / 1.621s）。默认真实媒体/模型例未启用。这些测试为 contract_local，不能替代账号、模型、浏览器或真实 S1 验收。

已实际追加运行 `go test ./...` PASS（无用户账号或新模型调用）；`go build ./...` PASS；`git diff --check` PASS。根 `pnpm check/build` 与真实 E2E 尚未在本轨运行；主控/W0 负责最终集成。

公开 adapter opt-in `ASTROCYTE_TEST_PUBLIC_DISCOVERY=1 go test ./internal/adapters/importers -run '^TestPublicBilibiliLiveMetadata$' -count=1 -v` 实际 PASS（0.14s），标题“【V2】（旧）”、items0/has_morefalse/next_cursor空。这是 adapter 的真实公共元数据诊断，仍不是用户非空账号同步/人工入库验收。

## S1 缺口与待协调接口（proposal，未冻结）

- AT01：现有真实单视频/论文历史不含本轮新多轮模型验收；需主控单次授权调度新 job，保留旧 Codex UNKNOWN、不自动重发。
- AT02：`app/jobs.go` ImportMaterial 对未固定版本 URL 使用新的 `IdempotencyKey` 作为 refresh 身份。新表单必然重新 Reader 提取，正文 digest 去重发生在工作已经完成之后；同命令 replay 不等于新表单复用。最小修正建议是由用户确认 ordinary reuse / explicit refresh，再由 W0 添加显式契约，W1 在 job 建立前用规范单视频身份复用/显式刷新；当前没有改变默认。
- AT03：`app/attention.go` authorize 对 agent 无条件拒绝，现有 machineRead 不是可到达的已授权正路径。需要用户明确资料/项目读取范围后，W0 transport 与 W1 app 同步对授权对象判断；machineRead 只记录 agent use，人类热度公式保持无 agent 事件。没有开放目录或令牌即全部可读。
- AT04：existing later 处理及历史正向保留，不以本阶段元数据解析测试宣称本轮 live 验收完成。
- 跨轨持久化建议：source 唯一 `(provider,source_kind,external_id)`，source item 唯一 `(source_id,provider_item_id)`；元数据 revision 与 material content revision 分离。selection / import job linkage 独立保存，人工勾选复用 ImportMaterial 和现有 jobs/CommandMeta/receipt/outbox 事务，不创建 Mission。
- 同步只在整套分页成功时事务提交可比较列表；失败、challenge、cancel、部分页不得清旧缓存或推进成功 cursor。同步恢复沿现有 job payload/claim/recovery；真实外部模型 unknown 保持不可 retry。W0 已同意该不变量并保留 migration006+ 单一写权；没有编辑迁移或 contracts.go。
- 未答：公开/登录绑定和实际 URL、标题简介建议或字幕优先、用户选取顺序、普通视频复用/刷新；因此尚未发布 sync/source/item/selection API、SQLite schema、抖音授权组件或默认同步频率。主流其他平台 registry 占位待 W0 公共定义，不伪造连接。

本阶段源码 SHA 在 commit/push 后通过 Orca handoff 报给主控。账号同步+人工入库及完整 S1 仍未完成；必须继续沿本轨返修与接线，不能把本阶段适配器调查结算为整个任务 succeeded。

## 继续阶段 1（2026-10-10）

普通合并主控 `ca7ba6e80ab256a33465c5bdd959400375569a0e`，本轮已回答的决定以 QUESTIONS 顶部为准。候选来源、用途、下一步齐全时，还必须有针对相同资料固定版本的成功沉淀记录提供主题关联或明确待查问题，才进入 ready_for_review；仅正文摘要、无关联记录或旧版本问题仍 incubating。保留原候选历史，没有创建 Mission。

`go test ./internal/attention/app ./internal/attention/domain` PASS（0.546s / cached）。新增候选门槛测试覆盖缺关联、仅摘要、明确问题正向和旧版本问题不满足新版本。真实模型/媒体/浏览器本轮未执行；等待主控时隙。同步持久化与 selected CLI 消费端口草案已发 W0，共享契约/迁移仍由 W0 唯一写入。

## 继续阶段 2（2026-10-10）

W0 公共 ABI/006 普通合入；普通新表单复用已有来源作业（包括历史旧规则作业），只有 refresh=true 显式刷新才再次提取，失败/UNKNOWN 不因换表单伪装成新调用。固定论文别名继续复用；视频 canonical source 共用领域函数，B站 tracking/p=1 别名复用，p=2 保持独立。显式 A→B→A 提取回到旧 digest 后复用旧 A revision，B head 不回退。

首次更新语义后的 app 测试 FAIL：旧 FixedArxivBoundary/MutableImport/CanonicalContract 测试仍要求新表单自动重新提取。按用户已确认普通复用+显式刷新调整其意图后，`go test ./internal/attention/app ./internal/adapters/importers ./internal/attention/domain ./internal/adapters/sqlite` PASS（0.461s/1.164s/cached/1.720s）。PowerShell glob 直接传给 gofmt 首次报 CreateFile，改显式路径后通过；不是业务失败。

真实指定 UID 3494358764489275 公开只读：锁定 yt-dlp2026.08.19 flat-list exited1 HTTP412，stdout null；薄 Python bridge 复用相同版本 BilibiliSpaceVideoIE 的原生 WBI/指纹/网络函数，单页 rawvlist 返回 code -352 验证失败。首失败保持；未换登录、读取 cookie、绕过挑战或下载未选媒体。公开收藏夹目录实际 code0/count0/listnull：仅公开目录为空，不证明上传列表成功。桥限单页30条/60秒/4MiB输出，严格 pin；测试保留精确 aid/UID、owner匹配、标题简介与不可用行。上游参考仍为 yt-dlp/extractor/bilibili.py；不新建 signing/downloader。

## 继续阶段 3（2026-10-10）

实现 W0 发布的 TrackingService/SourceCollectionService、SQLite006 CRUD/CAS：官方公开身份归一、多账号/显式收藏夹、metadata revision 与正文 revision 分离、逐页持久化 job payload 游标/上限、失败 partial cache + stale、成功/缺项中立保留、推荐与人工选取固定元数据版本。推荐需选项目+CLI，真实 caller 进入消费端口，不默认替换供应方。选取前必须当前 metadata recommendation succeeded；选取只建立现有 summarize_url import job/receipt，outbox 保存所选完整元数据/推荐依据。没有自动入库、Mission、定时或自启。

Service.Run 启动时一次元数据同步（system startup 队列，不伪装人类操作）。既有 job MaxAttempts/deadline/operation 保持，已知 retryable 安全失败由原队列有限重试，持久 UpdatedAt 控制间隔；每次真实失败 error/attempts 保留 outbox，旧 UNKNOWN 不自动重排。只读 source_sync 中断可恢复原 payload；返回的模型结果先保存 payload，再发布推荐。未知原生开始保持 UNKNOWN。

AttentionProjectReferences 已实现：W2 授权后 trusted agent caller 才可读 selected space 的固定版本；可选 C 只展开来自当前纳入根的成功沉淀 RelatedRefs 一跳，拒绝撤销/不在范围/其他 revision，读取前后都重新查成员。Agent use 只增加机器计数，不加人类关注。普通全库 Agent API 仍拒绝。

有意测试通过：应用3项（来源identity/100 metadata+推荐门槛+人工选取/部分缓存安全resume+UNKNOWN不重放）；固定旧版本连续3次 scoped Agent 读保持 human heat；真实SQLite停止/重开保持部分缓存、描述、revision、失败与游标，恢复只读第2页、operation/deadline不重置、materials0。SQLite测试首次最后查询误用 event_type 列 FAIL，查实际 schema 后改 type，独立重跑 PASS，首次失败保留。

`go test ./...` PASS；`go build ./...` PASS（本轮未运行模型/媒体/真实UI）。原公开来源 412/-352/公开folder0 仍有效限制。原 generic selected CLI 推荐/自动沉淀 factory 待下一具体阶段；实际100 provider rows（包括缺ID跳过行）的消费端口 Observed 字段已请 W0 添加，并准备调整 native page size 避免最后页多抓 metadata，不声称当前已通过该细节。W0 入口/HTTP/OpenAPI 与 W2 native 功能仍由原轨集成。

## 继续阶段 4（2026-10-10）

通用 selected CLI 推荐/自动沉淀适配完成：SelectedTextFactory/ListingRecommender 只消费 W0 SelectedTextProcessor，不跨应用依赖；real caller/project/CLI 每次 config/process 由组装桥校验模型许可。自动请求保存 ProjectID/CLI、实际配置与 SpaceID，读取 objects 前检查相同固定版本已纳入选定空间；撤销或空间改变不退回 legacy Codex。只有明确不选 project/CLI 的既有获准公共来源路保留原规则。选定项目可传空 processing_config，实际原生 config 由后端冻结；不猜模型名或评分。

推荐仅传选定公开 metadata，严格要求每个真实 ID 一条 text/reason，不接受编造 ID、score字段或尾随内容。自动输出复用已有 schema/引用校验；实际 CLI provenance 来自 text result，未知 model 保持空。新处理默认最高1800秒且服从已有 job context，legacy NewCodex 默认也改为1800秒，原 UNKNOWN 不重发。

actual100 cap：PublicListingReader 用剩余上限选择 native ps/pn，最后10条请求为 pn10/ps10，对 skipped missing-ID 行仍计 Observed；resume cursor 基于真实位置，API不会抓120条仅显示100。HTTP transport contract_local 测试实测请求总100行（含1条跳过ID），显示99条，5请求；上传末页10条解析测试PASS，没有再次真实网络探测挑战。

W0 发现 human A scope 空洞后本轨修正：trusted human/agent 都保留自身 identity 读固定 scope；普通human context查询无 implicitheat/机器使用，agent才计机器读。测试覆盖human连续2次、agent连续3次旧版本。Root另发现 bounded prefix 误标未观察缓存 stale，本轨修正为 only initialcursor empty AND !hasMore 时标缺项；针对 first100+explicitolder10+freshprefix100 的110缓存测试PASS。首发现记录保留。

app project factory测试覆盖未纳入拒绝、纳入后选定body处理、撤销模型许可无fallback；adapter推荐测试保留实际caller/配置/版本与未知模型、拒绝额外评分/ID/尾随。上述为 contract_local，没有模型/媒体/浏览器调用。`go test ./...`、`go vet ./...`、`go build ./...`、`git diff --check` 全部PASS。

本轮账户 live 仍未通过：给定UID公开uploads412/-352、公开收藏夹目录0，Douyin self不属公开来源且无公开profile transport。新论文DNS/rootlive时隙待主控；不改宿主代理，不用样本代替真实内容。全S1 AT01–04和当前模型/浏览器正向不能由上述本地测试宣称完成，Worker继续跟随root验收与原轨返修。

## 继续阶段 5（2026-10-10）

Root 首先发现通用文本请求2MiB，而真实W2 Registry只接受128KiB，之前的stub测试未覆盖这项原生边界；保持此真实集成失败，不把先前全绿改写成原生通过。本轨把自动沉淀和metadata推荐都改为128KiB输出。普通合入W0 `866b124b11033aa0f96945daa915b356e25a72c8` 后，新跨适配器测试把两条实际构造请求送到真实Registry.ProcessSelectedText；权限配置为contract_local，取消上下文仅用于在原生参数检查后、配置观察前停止，结果必须context.Canceled。没有启动CLI、访问私有配置或付费调用；此证据是原生接口边界，不是任务live模型结果。原生输入含schema/正文整体也受128KiB限制，超限明确失败，未截断正文或偷偷分块。

启动恢复补两项：已持久化最终page Done payload可在reader未配置时完成，不重抓；startup只保留当前安全失败的resume，已有更新完成任务时旧失败不挡住本次启动同步。旧失败记录没有重写。新增测试分别验证两种启动分支与保存最终页后完成；可选C一跳测试验证未开启拒绝、显式关联可读、撤销根后再次拒绝，没有递归扩展。

`go test ./internal/adapters/distillers ./internal/attention/app` PASS；`go test ./...`、`go vet ./...`、`go build ./...`、`node scripts/check-architecture.mjs`、`git diff --check` PASS。Root/W0后续fresh native model、媒体与W3浏览器验收未执行；真实账号公开挑战及Douyin transport限制仍保留，等待主控分配/验收，不宣称全S1完成。

## 协调者要求的历史真实导出入口（2026-10-10）

Root接受阶段5 owner implementation交接，未接受全S1/账号；明确保持同一Worker返修且暂不发worker_done、不启动本轨model/media/browser。按要求只核对仓库历史报告明确列出的公开资料原件，无个人磁盘搜索、fixture或新验证框架；路径/JSON关键字段现时存在已核对，后续导入仍需W3实际调用。

- 论文真实原件目录：`C:/Users/DW/AppData/Local/Temp/Astrocyte-S1-W2-ctx_00e6f40e6fe7/full-arxiv-import/`。`summarize-original.json`196347bytes，input/extracted.url均为`https://arxiv.org/html/2504.16054v1`，extracted.content94496字符，summary/llm=null；同目录原HTML279261bytes、PDF16213397bytes、Atom4613bytes、arxiv-imported-text.md95531bytes。API现有POST `/api/v1/materials/imports`可用`adapter=summarize_json,kind=paper,source_locator=该HTML地址,export_text=原JSON`、空source_key/content_digest和新的真实CommandMeta/key，不复用历史arxiv-text-command.json的旧request/digest。新provenance为`existing_json_export/0.21.8-format`，不冒称fresh official arXiv fetch；原PDF等不自动随这条导出路径成为附件。当前UI summarize选项仍kind=video，不能冒称已支持UI论文导出；fresh arxiv模式仍需实际网络。
- 视频真实原件目录：`C:/Users/DW/orca/workspaces/Astrocyte/s1-attention-ui-1009/web/test-results/attention-video-import-sel-d52ae--an-export-or-model-request-chromium-1280/evidence/`。`actual-video-content.json`33434bytes含原text11746字符与`processor=summarize,version=0.25.1; whisper.cpp,mode=upstream_media_transcript,source=https://www.bilibili.com/video/BV1PReT6EEqR/`；`actual-video-detail.json`保存原题“秋招企业级Agent项目，从架构讲到追问”。现有UI summarize纯文本/Markdown入口可粘贴exact text与原URL/title，不造时间位置；新provenance保持`existing_markdown_export/0.21.8-format`，原0.25.1转写provenance仍从真实content/report审阅，不伪装fresh视频URL或模型调用。
- `summarize-upstream-media.json`33213bytes仅text/provider/error/notes/segments，provider=whisper.cpp/error=null，不是带input.url的CLI export，不能直接按JSON导入；`summarize-cli-original.json`9973bytes为旧page-only输出，不能代替正文。三附件与实际job/restart报告留原目录，不转换成fixture或修改历史记录。上述细节已向root/W3发交接。

## 继续阶段 6：组装提示调用前大小校验（2026-10-10）

Root批准在两条selected适配器完整组装后按UTF-8实际bytes检查128KiB，超限返回已有ValidationFailed/review_selected_text_size；不截断、不分块、不调用ProcessSelectedText。Registry保留独立原生防御。首次新增回归RED：自动/推荐都实际进入processor stub1次并返回下游格式错误；修复后相同多字节正文（字符数未超上限但UTF-8超限）两模式均known preflight failure/0processcalls。启动后BudgetExhausted/DeliveryUnknown的保守分类没有改变，原UNKNOWN没有重放。

真实历史0.25.1 CLI page-only原件通过既有opt-in TestBilibiliPageOnlyIsNotVideoEvidence核验：PASS，明确evidence_missing，未运行媒体或网络。普通 `go test ./...`、`go vet ./...`、`go build ./...`、架构/diff检查仍PASS；不作为新模型任务live验收。

## 继续阶段 7：真实能力投影与跨适配论文身份（2026-10-10）

W0实际API第一套12PASS/2FAIL中，changed supplied export新job的失败断言由W0确认是错把fixed inputbytes当作普通URL复用，原失败保留后由W0移除过约束；本轨确认新bytes可形成新job/revision。另一项真实缺陷是注册ProjectDistillers即global available=true：本轨改false，保留selected-project-cli与select_permitted_project_cli说明，未选项目/客户端/许可/配置不声明连接可用；实际选定POST仍独立验证。

Root另指出官方arxiv Reader的arxiv:baseID与现有summarize_json paper导出的HTML URL key分裂资料。本轨仅actual kind=paper+既有NormalizeArxivID接受的官方论文URL归一到相同arxiv:baseID；summarize/manual既有导出两路适用，video/其他域名/无效ID不折叠。原始固定版本URL、真实导出内容/附件和existing_*_export/manual provenance原样保存，不声明官方fresh fetch或相同bytes。未迁移/重写已存在历史重复材料，也未改共享契约。

测试覆盖现代/旧arxiv ID、官方多URL版本、非paper及其他host不误并。跨adapter→Service→实际SQLite/objects→真close/reopen测试验证官方A与不同exportB同一Material/新revision2、exact B新表单job复用、旧A digest复用revision1但B head不回退、restart后唯一material及原JSON附件保留。HTTP为明确contract_local，不是真实arxiv重新抓取。该测试首次编译误用Job.Revision，随后查询误写materials表失败；按现有MaterialRevision字段/attention_materials表更正后单独PASS，首失败保留。Global status回归PASS。全Go test/vet/build、架构/diff检查PASS；实际fresh模型/媒体/browser仍由root安排。

## 导入专用2501.12948两版真实原件交接（2026-10-10）

Root要求AT02使用用户已授权IMPORT/version verification的2501.12948v1/v2，不授权模型。仅沿tasks/S1-plan末尾明确保留的目录查询：`C:/Users/DW/AppData/Local/Temp/astrocyte-s1-OHU8NF/data/state.sqlite`以mode=ro查历史material `JFKF5J6VWD43TN4W3LO4SVRD77`；r1/r2为同arxiv:2501.12948，原provenance是arxiv+summarize/2501.12948v1或v2;summarize0.21.8/official_atom_pdf_and_html_text，对应官方HTML。没有全盘/个人history搜索，没有修改历史库或文件。

原件根 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-OHU8NF/data/objects/` 下：v1原summarize JSON文件`159560bf4f2e4f2f82d4a1a03facd615d8a923345864d6cf30514a8249d237c1`（124424bytes、59375字符/59654UTF8bytes正文、input/extracted.url=https://arxiv.org/html/2501.12948v1）；v2文件`75a089219c3dbeeb5ccef83ee8166ec63c4f956d9ff6e34fb8386f624584aae8`（390015bytes、190458字符/191455UTF8bytes正文、对应v2HTML URL）。两原正文现时直接比较不同，summary/llm均null；可将exact原JSON直接粘贴现有summarize导出+kindpaper，source_locator用各自HTML版本地址。新导出provenance仍existing_json_export，不假称本轮官方下载；W1阶段7归一材料身份，differentbytes可新revision。无需手改body、制作fixture或复制整个数据库。v2超过128KiB原生输入，但这里只导入，不能模型处理这两论文。

root/W3已收到上述exact路径；原正文对象r1=`041548e2951f314ede1bfd128bbd477c86765021fec0feba0e482606c15abed5`、r2=`a65e3235482675bc6d13cae1b7e6b324db2690dbebe2f6b44f465781355daeb9`含官方Reader包装，和原JSON extracted.content形式不同，不混称同bytes。本交接不启动真实服务、browser或模型；root预告的已返回模型Result后真实对象发布失败场景保持待W3首RED，未预先改领域行为。

## 继续阶段 8：已返回结果的真实本地发布恢复（2026-10-10）

W3代码复核发现cached Result后普通OS Publish错误映射InternalError/Retryable=false，UI因此不能人工恢复。Root随后明确允许先在本轨自有临时目录复现/修复，不运行已知有缺陷的付费路径只为保存RED。普通合入W0 `dfd3949` 的当前W2 scope/native及W3选定模型/论文导出组装；没有编辑跨轨业务文件。

真实OS回归使用contract_local Reader/processor+实际SQLite/objects：在本地processor返回时只把自有objects目录移到同一TempDir的preserved-objects并在原目录位置写普通文件，造成真实OS发布失败；原对象保留。首次测试漏SourceRef.Locator在任何处理前失败，补实际固定URL后取得真实RED：job failed、InternalError不可retry、payload.Result已保存、processor calls1，不是假称付费模型失败。修复后失败为Retryable=true，中文说明已保存Result/修复存储后重试不会再次调用模型、action=repair_storage_then_retry_cached_result；底层原OS错误保存在已有attention.job_failed事件cause_detail。只有Result已成功提交后的本地Publish失败适用，Result持久化gap/native未知保持原UNKNOWN。

Root另确认存储修复必须先发生，因此此action不进入自动循环耗尽次数；其它known safe retries保持原机制。人工Retry仍受原MaxAttempts/deadline限制，不重置operation/payload/预算。回归恢复原目录、真close/reopen SQLite且processor=nil后，原job人工retry成功、attempts2、operation/deadline不变、calls仍1。应用测试验证等待修复不自动重排、重复本地失败耗尽2次后禁重试、processor始终只调用一次；未知禁重放既有测试继续通过。事件查询首次用未加attention前缀的type无行，按实际事件名更正后PASS；该测试错误保留，不改事件或历史来迎合断言。

`go test ./...`、`go vet ./...`、`go build ./...`、架构/diff检查PASS。本轮没有paid/media/browser调用；实际新模型返回后页面故障/恢复由W3继续，不能将本地processor+真实OS/SQLite覆盖写成已通过付费/浏览器验收。

## 继续阶段 9：选定项目冷缓存下保存已返回结果（2026-10-10）

Root复核指出阶段8真实OS/SQLite测试仅走legacy，实际SelectedTextFactory.Resolve此前预先ConfigurationID，selected project_id/cli在API重启native缓存unknown时仍不能保存Result。按root明确指令，Resolve改为只验证trusted human/caller和现有ProjectSpaceID（组装桥检查项目范围/外发CLI许可，不检查native配置或启动CLI）；新请求/新处理仍沿应用已有独立ConfigurationID检查。应用cached分支原resolve/checkScope流程不变，没有增加可选接口、共享fields或entry。

真实SelectedTextFactory+project_id/cli选定固定空间、contract_local权限/text port+实际SQLite/objects的OS故障测试扩展为legacy、native_unknown、revoked、space_changed、reference_removed。返回Result后恢复目录并真正close/reopen SQLite，配置标unknown、配置观察计数归零：selected native_unknown可完成原Result、配置观察0/processing仍1；撤销许可/改变项目空间/移除固定引用三者都ScopeDenied、无沉淀发布，配置观察0/processing1，UNKNOWN没有误改。native_unknown恢复后另一个新processing请求仍以EvidenceMissing失败且仅观察ConfigurationID1次，不调用处理器。全部场景PASS；此证明实际工厂和持久恢复，不把受控权限/text port说成真实付费native/完整API验收。

`go test ./...`、`go vet ./...`、`go build ./...`、架构/diff检查PASS。阶段8已accepted的legacy测试仍独立保留；本轮无网络/media/model/browser调用，W0/W3继续实际已返回付费Result的故障/恢复验收。

## 继续阶段 10：UNKNOWN 保留原服务错误原因（2026-10-10）

W3 的 SOURCE3660b0b 首次真实选定项目模型 job `4GQ7GKDXHZIND7CAB7GIZU3W4V` 18秒后 failed/delivery_unknown，Result:null。仅按其精确提供的 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-4pmMWO/data/state.sqlite` mode=ro 核对固定 job 及失败 event，没有改数据库或重发。首次 SQL 误用 job_id，真实 schema 是 id/data；按 schema 修正后读取 succeeded，outbox 实际表 attention_outbox，确认 cause/error 都被 genericUNKNOWN 覆盖、没有 cause_detail。external_started 是 app 调用前保守边界，并非原生 turn accepted；单独 probe session 不能作为本次模型会话证据。原未知保持、W2 查原生有限 method/reason，W3 新 README 验证由 root 另行授权。

Root 明确许可原写域最小日志修复：failJob 先从原 cause 映射复制已有 ServiceError，写已有新 event.cause；公开 job.Error 的 genericUNKNOWN、状态及重试拒绝不变，已知失败的预算耗尽行为不变，既有 FailureDetail OS 细节路径继续保留。没有新增字段系统、接口或日志框架，不把原始正文、凭据或任意原生输出写入事件，不改旧 event。新回归首次 RED 正确复现 cause 丢失有限 native turn/start deadline 说明；修复后同测试 PASS，原说明和 RequiredAction 保留，公共 Error 仍通用 UNKNOWN，旧未知 event 原 bytes 不变，显式 Retry 拒绝、Recover/Process 不调用第二次，operation/deadline/payload/attempts1 不变且无沉淀。

针对性 `TestAutomaticUnknownKeepsOriginalServiceCauseWithoutReplay`、既有 unknown/已知失败 retry 回归 PASS；`go test ./internal/attention/app ./internal/adapters/sqlite`、`go vet ./internal/attention/app`、`go build ./...`、`git diff --check` PASS。SOURCE `f1a44b269fc2d10a78ab3b64f3b765f3eaca5332`。本阶段无 paid/model/media/browser 调用；这是实际诊断缺失的功能修复，不将旧 UNKNOWN 改成已恢复或推定其原生原因，不宣称全 S1 验收。

## 继续阶段 11：公开收藏夹重查与真实失效行（2026-10-10）

本 Dispatch `ctx_514362ae7294` 普通合入主控 `41f3fab`，主控将任务收敛为公开收藏夹续验。指定 UID3494358764489275 匿名目录现在 HTTP200/code0，真实两夹：默认收藏夹2356677875/count452，王大葱的叠3501892975/count2。两夹首metadata页分别20行/has_more=true和2行/false；默认第5行是原指定BV1PReT6EEqR，external_id=2:117284634891766。原始响应独立保存于 `C:/Users/DW/AppData/Local/Temp/astrocyte-W1-public-ctx_514362ae7294/` 的 `bilibili-public-folders-first.json` 与 `bilibili-folder-{id}-page1.json`；历史目录0和上传412/-352仍保留，不改写。

小夹一行是70p长视频BV11t411C7Lk，另一行BV15NQrYGEuA明确title已失效视频/attr9但仍保留BV。现解析器忽略attr导致错误可选，新增回归首RED实际输出该行Locator非空/ProviderStatus nil/UnavailableReason空。最小修复复用已有provider_status字段保留attr，非零状态留清单/计入observed但不生成可导入URL并给unavailable_reason；0保持可选、缺字段保持nil、未知非零状态保守不可用，不让坏行使整页失败。语义参考[上游开源资源列表](https://github.com/bilibili-plugins/bilibili-api-collect/blob/master/docs/fav/list.md)中的medias.attr（0正常、9UP删除、1其他删除），与本轮匿名原件一致；没有变更transport或读取cookie。

`go test ./internal/adapters/importers -run 'TestFavoriteAvailability|TestStaleCollectionEntry|TestPublicListingBounds' -count=1` 修复后PASS；`git diff --check` PASS。主控把browser/model槽交W3，默认夹只推荐并人工选择原BV、复用既有正文；本轨继续纯metadata100cap/SQLite/重复sync核查，没有模型/媒体调用，真实选择入库正向仍等W3，不处理小夹长视频或失效视频。

阶段11功能源 `e1061eba28aec014e3f0742434fd8c42ce256697` 已push并交W0/W3。随后真实匿名PublicListingReader→应用Service→独立SQLite的 opt-in `TestPublicFavoritesLiveSyncReuseAndRestart` 首场失败：测试误约束100唯一行，上游每轮实际观察100行、分页缺行补位有重叠，所以保存99个唯一条目；实际job succeeded/has_moretrue/cursor6，业务上限和去重正确。首场DB及12份原响应保留 `.../sync-first/`，未覆盖。改测试按真实observations上限、唯一条目去重以及未变metadata revision检查，独立新目录 `.../sync-second/` 两轮都provider_rows100/items99/cursor6/has_moretrue，原指定BV为revision1可读；两种官方folder URL绑定同一source ID、不产生重复绑定，未变metadata不升revision。真正close/reopen后完整source/items逐字段一致；materials/distillations/opportunities为0、无建议/选取/import状态。**PASS**，测试2.79秒/Go包11.841秒，exit0。原响应24份、真实DB与 `actual-source-after-repeat.json` 留sync-second，原 `live-sync-second.log` 留父目录。没有fixture/provider mock，也没有浏览器或模型/媒体调用。

仅在主控指定的 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-4pmMWO/data/state.sqlite` mode=ro查看固定import jobs：真实URL原receipt `26MPWE7S3S4FLEUE62BZLKC54T` succeeded/summarize_url/video，canonical URL是BV1PReT6EEqR，无source_key/content_digest/local_file_ref，原Refreshtrue。现有普通enqueueImport忽略old.Refresh并复用latest同URL/adapter输入，source selection同样summarize_url/空key/digest，因此可直接复用，不需新媒体。原旧summarize导出单独存在；没有改库、重发UNKNOWN或用导出来冒充本次URL提取。初次只读脚本误从data读payload报KeyError，改用现有独立payload列后成功；该诊断失败不是业务处理失败。

当前含opt-in真实验证源码：`go test ./...`、`go vet ./...`、`go build ./...`、`node scripts/check-architecture.mjs`（19包）和 `git diff --check` 均PASS；默认Go全套跳过外部live测试，以上真实HTTP/SQLite单独明确执行并记录。浏览器实际推荐/人工选择/旧正文复用与冷重启仍由W3唯一槽验收，不能把本轨纯metadata PASS代替它。Douyin self公开身份与匿名transport仍未取得，上传挑战历史不变，旧UNKNOWN与首失败保留。

阶段11真实metadata验证源码/报告 `8c7292c407526903f956d903ec0bcfd0f028f7f3` 已push。仅原独立验证库只读查作业也确认两个source_sync/succeeded、sources1/items99，无媒体/模型作业。

## 继续阶段 12：推荐 API 固定输入复用与 UNKNOWN 防绕过（2026-10-10）

主控实际源码复核发现RecommendSourceItems仅按新幂等键排工作，UI禁重做不能保护API。新增应用回归首RED：相同metadata/project/CLI换新key使processor calls2，UNKNOWN换新key返回nil。沿已有事务、ListJobs/payload与digest增加固定输入匹配：source ID、trusted human身份、project/CLI、当前configuration、选中条目的完整metadata/revision；选取顺序排序，不把pending/success建议、selected/importjob、request/operation等临时状态当作新输入。直接扫描原payload兼容已经queued/succeeded/UNKNOWN的旧key，没有迁移/新表/调度系统。

同输入成功/queued/known failure返回原job与原固定结果，避免新key重设budget/operation；相同UNKNOWN明确DeliveryUnknown拒绝，并保留原记录。不同metadata/revision/config/project/CLI/human独立工作。每次请求仍先沿配置端口检查现时项目权限，撤权不因缓存成功跳过。Source级active作业仍阻止不同工作并发；重复同queued工作复用。完成结果仅投影原匹配结果供本次响应，不把另一个项目的建议称为本次结果，也不重写旧job/payload。未改共享契约、入口或UI。

应用两原RED回归修复后PASS。`TestRecommendationSemanticReuseSQLiteRestartAndScope` 使用contract_local权限/处理端口+实际SQLite/close/reopen：pending新key反序选取同job、成功冷重启两新key处理器保持calls1；UNKNOWN冷重启两新key都DeliveryUnknown/calls1，原jobdata/payload逐字节不变；撤权ScopeDenied。成功组随后逐项改变config/project/CLI/human/metadata(title+rev2)，各只新增一次正确处理、最终calls6，原完成job不变，材料0。不是付费/native或browser验收，没有执行额外模型/媒体。`go test ./...`、`go vet ./...`、`go build ./...`、19包架构和diff检查PASS；测试日志简化后独立SQLite场景另复验，不重复真实HTTP或付费链路。等待W0普通合入、W3在唯一现场实际推荐后新key复用断言与主控最终接纳。
