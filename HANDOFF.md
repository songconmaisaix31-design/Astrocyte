# Astrocyte 主控交接

## 当前交付与下一步（2026-10-10）

先读 [STATUS 顶部](STATUS.md)、[用户决定](docs/QUESTIONS.md) 和 [四轨计划](tasks/S1-sync-local-agent-plan.md)。主控分支 `s1/attention-materials-20261009`，当前代码组装 `3e3ee8cc5c51233935d8665f3a411371169c7ff7`；四轨沿原工作树普通合并，主控没有写业务代码。check/build通过；主控默认浏览器首场152 PASS/2 FAIL/8 SKIP，两个旧授权提示断言已由原W3修复，单独两尺寸复验2 PASS/32.1秒。首场失败保留，没有宣称整套重跑变绿。

S1 资料核心的 AT01–04 已通过真实浏览器续验，源码 `cd721692122c5e8f7820e34f25181058dd648cf0`，1 PASS、测试1.5分钟/整场1.7分钟。指定论文真实既有 JSON 与指定视频现网 URL、固定版本和两轮实际模型整理、候选依据/下一步/延期、授权 Agent 三次读取不增人类关注，以及保存模型结果后真实发布失败、冷重启、人工重试不重复模型均通过。主控只读核对同库四个模型作业成功、五条沉淀、候选 deferred/later 和对象原文，并目视1280/1920截图。旧未知作业保留、未重发；额外论文2501.12948v1/v2只用于真实版本验证。

账号/公开收藏夹绑定、启动一次同步100条、元数据建议后人工选择、单视频 summarize、项目登记/A默认BC可选/按项目选择 CLI 与身份授权已组装。本机 Codex 原生 start/resume/send/read_context/observe/stop 已完成真实浏览器；Claude 实际两轮和同一原生 ID 接续通过，Pi 认证失败。已安装不等于已配置或可用。自动派发仍关闭，个人目录和历史根尚未提供。

真实非空账号批量选择尚未验收：指定 B站 UID 匿名投稿412/-352、收藏列表0条；抖音 self 地址需要公开身份，不能读取登录凭据代替。arXiv 新 URL 当前假IP阻断，论文使用真实既有导出。第二自动整理 CLI 的配置选择、双源完整模型输入容量及自动 Agent 控制限度待用户，详见 QUESTIONS；不据此扩大权限或改变输入上限。

实际资料库 `%LOCALAPPDATA%/Temp/astrocyte-s1-4pmMWO`、原生库 `astrocyte-s1-pznwxr` 和 W3 原始 Playwright 结果保留，所属验收进程已退出。个人库升级前备份 `%APPDATA%/astrocyte/backups/pre-s1-sync-20261010-1640/state.sqlite`，原资料/作业均0，未灌入验收数据。最新预览已在16:54启动：[资料沉淀](http://127.0.0.1:5173/attention)、[本地项目](http://127.0.0.1:5173/workspace)，summarize启用、全局自动模型处理关闭，按项目许可操作。最终进程与检查见 STATUS 顶部，下文旧 PID 不可复用。

## 前轮公共部分与历史交接

当前轮次：2026-10-10 账号同步、本地 Agent 接入与完整 S1；先读 [STATUS 顶部](STATUS.md)、[四轨计划](tasks/S1-sync-local-agent-plan.md)、[待答决定](docs/QUESTIONS.md)。公共列表适配、缓存 CLI 清单/保护 API、局部参考图界面已合入集成轨，主控 check/build PASS、完整浏览器152 PASS/4 SKIP。账号绑定/同步入库、项目原生操作与完整 AT01–04 尚未完成，不能沿前轮单视频成功或公共部分绿标记 S1 完成。

四轨固定工作树/分支 `s1-sync-contract-1010`、`s1-sync-attention-1010`、`s1-local-agents-1010`、`s1-sync-ui-1010`，本机 Orca Codex / gpt-6.1-sol。W0 是契约/入口/迁移/依赖及集成唯一 owner；W1/W2/W3 原任务未完成结算，工作树与已push源码保留。用户答复后沿原轨、原写域重新分派，不另建框架，不扫描用户目录或推导权限，不重发旧 UNKNOWN。

最新 CLI 探测只核版本/help：10个已知客户端、8个安装，配置/可启动/原生均未知。官方抖音列表解析不是登录绑定；B站实际公开空收藏夹只验证空结果与错误区分，不替代用户非空收藏夹同步验收。新增平台目录保持待接入。当前业务与最终预览来源见 STATUS 最新记录；下文预览 PID 和结论仅为历史，操作前重新核实。

本轮最终业务合并 `e131feeeb63b203ae1a0a2332912d876f923f87a`（W0 `b67921f`）。完整套件源 `bbf99ac` 后仅补平台名称，另经类型/构建与主控实际两尺寸页面检查。四轨已结算释放，工作树保留；主控仍在 `s1/attention-materials-20261009`，未合 main。03:30:34 原预览已重启到该版本，summarize启用、Codex关闭，真实客户端10/安装8、平台占位14；具体进程与限制见 STATUS。

2026-10-10。项目目录：`C:\Users\DW\orca\Astrocyte`。

指定单视频导入已修好，见 [当前计划](tasks/S1-video-fix-plan.md) 与 [当前结果](STATUS.md)。用户“允许修改”已授权B站DNS例外，01:07热加载与上游检查PASS。W0修正Windows项目CLI路径及既有Reader/Service默认期限1800秒；Codex独立180秒不变。W2首次真实Service成功730.29秒；W3从空库实际页面提交指定BV成功，作业275.98秒、attempts1，保存32908bytes正文和三附件，真正停止/重启服务后API、SQLite/对象逐字节一致，同一结果resize两尺寸通过。主控复核JSON及截图，独立check/build PASS、完整E2E146 PASS/4 SKIP（4.7分钟、exit0），真实正向另为1 PASS。业务/验收源017c133ec4322a6d1011874eee25e148d814fd9c，普通最终报告合并a1a2d5213d2a17ede83c04ecbce95e308d687080只有报告差异。三轨已push、验收释放，工作树保留。

01:35:52预览已重启到017c133业务源，5173/8787均200，原个人库材料/作业仍0、Codex关闭。当前准确API PID86716、Vite56584，父64976，目录`%LOCALAPPDATA%/Temp/astrocyte-dev-QtmjWR`；停止前必须重新核对所属关系/创建时间/监听，不能依赖此历史PID。验收服务/媒体/浏览器已退出，测试端口无监听。实际原件和截图位置见 [W3](tasks/S1-W3.md)，W2首次原件保留。W3默认dev临时`astrocyte-dev-HuaPRH`被自动审批拒绝删除（仅blocked by policy）；主控默认dev`astrocyte-dev-x8dhVu`也保留且无服务，人工清理即可，不绕过审查。

剩余范围：中文ASR误识、无上游时间片段；转写中途取消未实测；普通新表单视频复用/主动刷新语义待答。完整S1的最新自动多轮/摘要重试、授权Agent正向读取仍待；旧Codex UNKNOWN未重发，arXiv当前DNS未调整。账号/公开收藏夹同步、先标题简介反馈还是先字幕、同步频率与实际来源尚待用户，不冻结接口或全量自动入库。Clash持久文件/运行DNS已改，但设置UI缓存未刷新，后续设置操作可能覆盖B站例外。下文是前轮交接快照，不将其阻断或未运行结论覆盖本轮单视频成功。

## summarize 接入前轮交接（历史）

最新任务：用户要求正式接入 steipete/summarize，随后明确 **允许Python**，原无Python门槛不再适用。项目已锁定并组装 summarize 0.25.1、公开视频 URL 入口和上游媒体薄适配；独立 Python yt-dlp、ffmpeg、whisper.cpp 与多语言模型已经安装并检查。当前真实视频正文仍被本机假IP DNS阻断，未获得字幕/转写/首次模型整理，完整S1未通过。当前源码、最终检查与限制以 [STATUS](STATUS.md) 顶部和 [接入计划](tasks/S1-summarize-plan.md) 为准。普通视频导入复用/主动刷新、公开来源DNS调整、登录和扩展仍待用户；不改变网络检查或授权边界来绕过问题。

S1 主控分支 `s1/attention-materials-20261009` 当前业务合并源 `81e92ae521fadf6def68b771e005bd77cc3f7ed3`。当前主控结果见STATUS顶部，后续收口只更新交接文档。历史论文第二步源 `556165c338519e7f35818d9ea05113cc3644f862` 的 check/build/E2E PASS（API14、单测40、浏览器144，59.9秒）仍保留，不能代替新媒体链路通过。原写域与集成记录见 [S1 计划](tasks/S1-plan.md)；当前四轨已结算释放、工作树保留。用户确认“域”、算法排序、人工分类、项目空间 @文件建立引用且原分类保留、整理先接本机 Codex CLI。视频复用锁定summarize及上游媒体实现，不新建下载/转录框架。

第二步论文纵向链路已通过：用户允许公开多版本论文，实际页面提交 `2501.12948v1/v2`，SQLite保存同一资料两个版本、对象保留两套原文/附件；不同新表单及固定v1的PDF/ID别名复用同一job、attempts=1。实际后端停止/重启后，详情、两版正文和8个附件经API/存储核对一致，浏览器两尺寸重载可看旧v1，关注0、Mission0。首次指定论文重复提交曾产生两次下载，保留该失败与后续修复区别。旧库Refresh键不迁移，首次新规则命令可能再建job；原回执/历史记录不删除。该额外论文只授权导入/版本验证，不扩大模型授权。

真实材料为 arXiv `2504.16054` / B 站 `BV1PReT6EEqR`。论文全文、人工内容/主题待查记录、候选 later 已由浏览器/API/SQLite/重载核对，历史 AT04 PASS；完整 AT01、真实最新摘要重试和授权 Agent 正向 AT03 尚未通过。旧 summarize 网页文字被拒绝、最新 Codex 调用超时未知均保留，未重放。新版本真实B站URL提交已到达服务并持久化网络失败，原命令重放/SQLite实际重启保留原失败且无重复提取；不能替代成功视频与新表单同源去重。Agent范围、私有外发、候选门槛、A→B→A默认版本与新模型作业仍待用户，不得从热度或令牌推导权限。宿主ACL探针副作用未凭猜测回滚，见 [W2](tasks/S1-W2.md)。

当前预览在2026-10-10 00:32重启到最新集成源：http://127.0.0.1:5173/attention，API8787；两端HTTP200。继续使用原个人SQLite，迁移前备份 `%APPDATA%/astrocyte/backups/pre-s1-20261009-2230/state.sqlite` 保留。启动自动解析项目 summarize 0.25.1 与独立媒体工具，已清除该预览终端的旧全局CLI覆盖；Codex自动处理关闭，验收资料未灌入个人库。原预览Windows清理报告子进程已退出竞态，旧PID/监听实际消失后才新启动。准确进程/遗留测试目录见STATUS。下文 S1 前界面仅为历史，不是当前写入能力。

历史交付：`ui/preview-replacement-20261009`；源码 `5c1517d652afef8273cf14dab5c28021db648cfa` 已 push，尚未合入 main。S0 交接基线为 `649d4297d5bd34ebd849944379cb07bf909d7acb`。以下内容描述 S1 前的界面基线，当前开发与决定以顶部链接为准。

## 界面基线状态（S1 前）

- 界面已采用用户提供的 `Astrocyte-preview.html` 设计迁入 React/TypeScript：白绿布局、搜索、导航、三页标签、封面/会话卡、右栏、插画、详情抽屉与键盘路径。现有客户端、查询层、DTO 和后端保持原边界，开发记录见 [前端交接](web/UI-preview-report.md)。
- 主 Agent 在当前源码执行 `pnpm check`、`pnpm build`、`pnpm test:e2e` 均 PASS；37 项单元测试、112 项浏览器测试，另有两尺寸三页真实/示例截图复核。本轮 CI 仍在运行；此前 S0 的 Windows/Linux CI PASS 属于历史基线。记录见 [STATUS.md](STATUS.md)。
- S0 完成：Go + React/TypeScript 基座、OpenAPI 客户端、SQLite 迁移与备份、开发/检查/构建入口。
- 当前 API 提供健康、基座与空集合查询；写操作、SSE、原生接续返回 501。页面默认读真实 API；`?fixture=1` 进入示例模式。
- S1–S6 尚未实施。8 个本地 Agent CLI 已清点；原生能力未实测。summarize 已安装，真实导出与 arXiv 导入未实现。
- 搜索限当前页已加载数据；收藏、研究路线和事件接口尚未接入，真实拓扑禁用。HTML 的固定动态与图仅在显式示例模式展示。改为默认示例尚未决定，沿用 S0 默认真实 API 的约定。

## 阅读入口

| 文件 | 用途 |
|---|---|
| [AGENTS.md](AGENTS.md) | 执行协议及用户要求 |
| [docs/SPEC.md](docs/SPEC.md) | 业务规格；下一切片先读 §19、对应领域及验收章节 |
| [docs/TASKS.md](docs/TASKS.md) | 切片任务；日期与 Agent 数是参考估算 |
| [docs/QUESTIONS.md](docs/QUESTIONS.md) | 已确认与待定事项 |
| [README.md](README.md) | 安装、启动、检查、备份 |
| [tasks/S0-plan.md](tasks/S0-plan.md) | S0 轨道、负责人、模型与写域 |
| [docs/probes/local-agents.md](docs/probes/local-agents.md) | 本地工具入口与版本清单 |

## 本地运行

Go 1.27.2、Node 24.16.0、pnpm 11.27.0 已安装。Go 位于 `%LOCALAPPDATA%\Programs\go\bin\go.exe`，项目脚本可自动找到。

- 预览正在运行：http://127.0.0.1:5173/workspace；设计示例入口：http://127.0.0.1:5173/workspace?fixture=1；API：`http://127.0.0.1:8787/api/v1/health`。
- 新预览终端：`term_af3ba99d-1301-41be-be0c-860189357aaa`。旧 S0 终端已失效。重启时在新终端按 Ctrl+C，再在项目根运行 `pnpm dev`；已有服务时不重复启动。
- 数据库：`C:\Users\DW\AppData\Roaming\astrocyte\state.sqlite`。重启沿用该目录；Windows/WSL 各用独立数据库。

```powershell
pnpm check
pnpm test:e2e
pnpm build
```

## 接下来

1. 继续收口 S1 的指定视频 summarize 真输出、授权 Agent 读取正路径及最新 Codex 自动整理；用户未回答的决定保持 pending，不扩大权限，不重发未知效果调用。
2. 按确认后的范围恢复原 Worker 返修，沿原互斥写域与分支，不重新搭调度框架；若允许新模型作业，保留原失败与新作业区别，同时核对队列与 native 超时上限。
3. 完整 AT01–AT04 通过前不宣称 S1 完成，也不进入 S2 批准/执行。P0 剩余适配仅在下一切片确需时处理。

## 开发约束

- 保持 Go/TS/SQLite 与现有目录；用户已允许 Python 作为媒体依赖。领域在 `internal/{attention,workspace,swarm}/{domain,app}`，适配器在 `internal/adapters`，组装在 `cmd/server`，页面在 `web/src/pages`。
- 修改 API 从 `contracts/openapi.yaml` 开始，运行 `pnpm generate`；不手改生成类型。契约、迁移序列、依赖锁和入口各设单一所有者。
- 主控负责计划、决策、最终 check、提交与 push；业务代码和返修由对应 Worker 负责。使用 Orca CLI，保持一轨一 Agent、一 worktree、一 branch；难任务强模型，简单任务便宜模型。依赖/模型下载已授权，费用和调用不限。
- S0 轨道已收口，工作树保留在 `C:\Users\DW\orca\workspaces\Astrocyte\s0-*`。后续并行时新建任务与 Dispatch；保留原有 charybdis 工作树及旧项目。
- 保留 Playwright 的 `gracefulShutdown`：Linux 测试退出依靠它关闭 API/Vite 子进程。安装浏览器前设 `PLAYWRIGHT_SKIP_BROWSER_GC=1`，保留其他项目使用的版本。
- 只执行能发现问题或支持下一步开发的检查。关键歧义先问用户；示例数据与真实接入分别记录。
