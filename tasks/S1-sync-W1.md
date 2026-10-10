# S1 sync W1 — 公开列表调查与独立同步规则

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
