# Astrocyte 主控交接

2026-10-09。项目目录：`C:\Users\DW\orca\Astrocyte`。

当前任务为 S1 开发，主控分支 `s1/attention-materials-20261009`，五条 Orca 写域轨与集成进度见 [S1 计划](tasks/S1-plan.md) 和 [STATUS.md](STATUS.md)。用户确认“域”、算法排序、人工分类、在项目空间 @文件建立引用且原分类保留、自动整理先接本机 Codex CLI。Agent 读取范围的关键决定仍待用户，不得从热度或令牌推导全库权限。真实材料为 arXiv `2504.16054` / B 站 `BV1PReT6EEqR`，默认链路无 Python；最终 AT01–AT04 未通过。

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

1. 与用户确认下一阶段范围及真实案例。规划的下一切片是 S1「资料到候选」；论文方向已定 arXiv，视频总结工具已定 summarize，具体材料、目标仓库和改进任务待用户选择。
2. 按下一阶段需要补 P0-05～09：Agent 原生能力、向量索引、总结导出字段、AOCI/CodeGraph、Laya ONNX。后端与模型选型、执行控制模式、外发范围及 S3 接续降级方案，在相关切片开始前询问用户。
3. 先写一页计划，任务固定「目标、可改路径、依赖、验收、停止点」；公共契约先发布，再按互斥文件路径并行。

## 开发约束

- 保持 Go/TS/SQLite、无 Python 平台与现有目录。领域在 `internal/{attention,workspace,swarm}/{domain,app}`，适配器在 `internal/adapters`，组装在 `cmd/server`，页面在 `web/src/pages`。
- 修改 API 从 `contracts/openapi.yaml` 开始，运行 `pnpm generate`；不手改生成类型。契约、迁移序列、依赖锁和入口各设单一所有者。
- 主控负责计划、决策、最终 check、提交与 push；业务代码和返修由对应 Worker 负责。使用 Orca CLI，保持一轨一 Agent、一 worktree、一 branch；难任务强模型，简单任务便宜模型。依赖/模型下载已授权，费用和调用不限。
- S0 轨道已收口，工作树保留在 `C:\Users\DW\orca\workspaces\Astrocyte\s0-*`。后续并行时新建任务与 Dispatch；保留原有 charybdis 工作树及旧项目。
- 保留 Playwright 的 `gracefulShutdown`：Linux 测试退出依靠它关闭 API/Vite 子进程。安装浏览器前设 `PLAYWRIGHT_SKIP_BROWSER_GC=1`，保留其他项目使用的版本。
- 只执行能发现问题或支持下一步开发的检查。关键歧义先问用户；示例数据与真实接入分别记录。
