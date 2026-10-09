# Astrocyte 主控交接

2026-10-09。项目目录：`C:\Users\DW\orca\Astrocyte`。

交接基线：`s0/workbench-foundation` / `549c78928bd93ed7fec9ef607b7a9aa48096dc74`，已 push，尚未合入 main。

## 当前状态

- S0 完成：Go + React/TypeScript 基座、三页界面、OpenAPI 客户端、SQLite 迁移与备份、开发/检查/构建入口。
- `pnpm check`、`pnpm build`、`pnpm test:e2e` 通过；37 项单元测试、88 项浏览器测试。Windows/Linux CI 通过，含 Linux race。记录见 [STATUS.md](STATUS.md)。
- 当前 API 提供健康、基座与空集合查询；写操作、SSE、原生接续返回 501。页面默认读真实 API；`?fixture=1` 进入示例模式。
- S1–S6 尚未实施。8 个本地 Agent CLI 已清点；原生能力未实测。summarize 已安装，真实导出与 arXiv 导入未实现。

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

- 预览正在运行：http://127.0.0.1:5173；API：`http://127.0.0.1:8787/api/v1/health`。
- 预览终端：`term_b2100c22-c321-4344-9578-8b9f42601ce2`。重启时在该 Orca 终端按 Ctrl+C，再在项目根运行 `pnpm dev`；已有服务时不重复启动。
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
