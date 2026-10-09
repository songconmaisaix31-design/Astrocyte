# Astrocyte 主控交接

2026-10-10。项目目录：`C:\Users\DW\orca\Astrocyte`。

正在执行视频修复切片，见 [当前计划](tasks/S1-video-fix-plan.md)。用户“允许修改”已明确授权所选B站DNS例外：01:07实际热加载成功，上游网络检查PASS，真实音轨已下载，唯一一次本地转写正在运行。W0修复Windows下直接启动的项目CLI真实路径并调整既有默认作业期限；W3准备独立临时库正向页面验收，尚未运行。下文假IP阻断与默认300秒为前一轮结果，不代表当前状态；视频正文、持久化成功和浏览器正向仍待本轮实际结果。公开账号/收藏夹筛选的关键产品决定仍待用户，未冻结新契约。

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

- 保持 Go/TS/SQLite、无 Python 平台与现有目录。领域在 `internal/{attention,workspace,swarm}/{domain,app}`，适配器在 `internal/adapters`，组装在 `cmd/server`，页面在 `web/src/pages`。
- 修改 API 从 `contracts/openapi.yaml` 开始，运行 `pnpm generate`；不手改生成类型。契约、迁移序列、依赖锁和入口各设单一所有者。
- 主控负责计划、决策、最终 check、提交与 push；业务代码和返修由对应 Worker 负责。使用 Orca CLI，保持一轨一 Agent、一 worktree、一 branch；难任务强模型，简单任务便宜模型。依赖/模型下载已授权，费用和调用不限。
- S0 轨道已收口，工作树保留在 `C:\Users\DW\orca\workspaces\Astrocyte\s0-*`。后续并行时新建任务与 Dispatch；保留原有 charybdis 工作树及旧项目。
- 保留 Playwright 的 `gracefulShutdown`：Linux 测试退出依靠它关闭 API/Vite 子进程。安装浏览器前设 `PLAYWRIGHT_SKIP_BROWSER_GC=1`，保留其他项目使用的版本。
- 只执行能发现问题或支持下一步开发的检查。关键歧义先问用户；示例数据与真实接入分别记录。
