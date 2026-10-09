# Astrocyte 当前状态

日期：2026-10-09。当前任务：[S0 计划](tasks/S0-plan.md)。

- 起始代码：只有 AGPL v3 LICENSE，初始提交 3a02ce38481d1885adee2e051a1b524d78376897。
- 文档基线：de433528d360413305a4a65c997e11ef7fbf4841，已 push baseline/workbench-v0.1。该提交的 SPEC/TASKS 与用户下载文件逐字节一致。
- 当前分支：s0/workbench-foundation。SPEC/TASKS 后续修订记录用户确认，原文保留于上述基线提交。
- 远端：https://github.com/songconmaisaix31-design/Astrocyte。
- 主工作树：C:\Users\DW\orca\Astrocyte；原有 charybdis 工作树保持原位。

用户决定见 [docs/QUESTIONS.md](docs/QUESTIONS.md)：先交付 S0；主 Agent 最终 check/提交；并发无数量上限；允许依赖/模型下载；模型按任务难度选择；summarize 和 arXiv 为导入方向，案例后定。

## 环境与进度

| 项目 | 当前结果 |
|---|---|
| Node / pnpm | v24.16.0 / 11.27.0 |
| Codex / Claude Code | 0.160.0 / 2.1.238，原生能力尚未验收 |
| Go | go1.27.2 windows/amd64，已安装；纯 Go SQLite，无 CGO 依赖 |
| W0 契约与入口 | 已合入 a1530cd；锁定依赖、生成客户端、开发/检查/构建脚本及 CI |
| W1 Go/SQLite | 已合入 72409eb；含保留记录的迁移、一致性备份、完整错误响应 |
| W2 三页 | 已合入 96cb7f5；含刷新失败保留数据、焦点与请求清理修正 |
| W3 本地工具清点 | 已合入 59fd965；8 个 Agent CLI、summarize 版本与帮助入口已核实 |
| S0 | PASS；主 Agent 本地检查、构建、浏览器与实际启动通过，最终分支 Windows/Linux CI 通过 |
| S1–S6、AT01–AT16 | NOT_RUN |

开发使用 Codex、Claude Code、Pi、OpenCode 四种本地 Agent；CLI 名称与实际配置的模型分别记录，不以客户端名称推断模型。

业务集成源：`856698000fba3271d25ff68434ef4ced08e2d9fc`；Linux 浏览器测试退出修复：`0169ce2c075c94e5c850408c6f651fb0a1718cf5`。最终源码为主分支普通合并提交 `0a0f49c25f4c2e4a7714e80387f98c48dd7441e3`。

首次合并检查：`pnpm check` 在 gofmt 阶段 FAIL（`cmd/server/main.go`）；独立 `pnpm --dir web lint` FAIL（动态 Hook 依赖，另有三项警告）。原负责人修正后，主 Agent 在上述最终合并源码运行 `pnpm check` PASS。OpenAPI 保留一项 `EventV1` 未引用警告；ESLint 零错误、零警告，Vitest 37/37。

## 主 Agent 最终验证

环境：Windows / PowerShell，SOURCE_SHA `fa7b97f659e6773d83d7dd864ddd0b0835ceac77`。

| 实际操作 | 结果 |
|---|---|
| `go mod download` | PASS |
| `pnpm install --frozen-lockfile` | PASS |
| `pnpm check` | PASS；Go 格式/vet/测试、依赖、13 包导入边界、226 契约示例、生成漂移、TS、lint、Vitest |
| `pnpm build` | PASS；`dist/astrocyte.exe` 与 `web/dist` |
| `pnpm test:e2e` | PASS，88/88；1920×1080、1280×720 |
| 两尺寸三页与详情截图检查 | PASS；标题、徽标、布局、示例标记和禁用操作可辨认 |
| 构建产物在仓库外目录启动，按 OpenAPI 验证实际响应 | PASS；24 操作、空集合、501、Origin 拒绝、错误 DTO、4 网页路由 |

退出修复合入后的主 Agent 复验：SOURCE_SHA `0a0f49c25f4c2e4a7714e80387f98c48dd7441e3`，`pnpm --dir web typecheck`、`git diff --check`、`pnpm test:e2e` 均 PASS（88/88）。应用代码未因该三行测试服务配置修改而改变。

## CI 与退出问题

最终分支 SOURCE_SHA `0a0f49c25f4c2e4a7714e80387f98c48dd7441e3`，[运行 37901441478](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37901441478)，整体 success：

| 环境 | 实际命令与结果 |
|---|---|
| Windows-2025 | `go mod download`、冻结锁安装、`pnpm check`、`pnpm build`、Chromium 安装、`pnpm test:e2e` 全部 PASS |
| Ubuntu-24.04 | 上述命令全部 PASS；另 `go test -mod=readonly -race ./...` PASS |

初次集成源的 [运行 37897830295](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37897830295) 保留为 cancelled：Windows PASS；Linux 的 88 项断言成功后服务退出卡住，取得日志时主动取消。退出卡住由 Playwright 默认强制关闭监督进程、留下持有输出管道的独立 API/Vite 进程组造成。原 W0 在 `0169ce2` 加入标准 `gracefulShutdown`（SIGTERM，15 秒），让现有开发脚本关闭自己的子进程组；未改变浏览器断言。修复前主分支 [运行 37900190606](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37900190606) 的 Windows PASS，Linux 仍卡住后取消；修复源 [运行 37901041699](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37901041699) 两平台 PASS，Linux 显示服务正常关闭，浏览器阶段约 50 秒完成。最终分支上述运行另行通过。

## 本地交付

开发预览由 Orca 终端 `term_b2100c22-c321-4344-9578-8b9f42601ce2` 启动 `pnpm dev`，就绪检查 PASS：网页、代理后的 health/foundation 均为 200。

- 网页：http://127.0.0.1:5173；三页显式示例入口在 [README.md](README.md)。
- API：http://127.0.0.1:8787/api/v1/health。
- 数据：`C:\Users\DW\AppData\Roaming\astrocyte\state.sqlite`，位于仓库外。
- 结束预览：在该 Orca 终端按 Ctrl+C。

最终验收记录为纯文档变更，执行本地链接与 `git diff --check` 后提交；不重复触发应用全套 CI。交付分支为 `s0/workbench-foundation`，已 push；GitHub main 尚未合入。

## 后续适配准备

S0 按 SPEC §19.1 的基座与契约验收收口。开发规划 P0 的适配准备不随基座通过自动变为 PASS：

- P0-05：Agent 八项原生能力实测 NOT_RUN；当前报告是版本与帮助清点。
- P0-06：无 Python 向量后端选型与实测 NOT_RUN，尚未决定后端。
- P0-07：summarize 已安装；真实导出样本与字段映射 NOT_RUN。用户要求先搭环境，案例后定。
- P0-08：AOCI/CodeGraph 所选版本、许可和工作树隔离实测 NOT_RUN。
- P0-09：Laya 可直接加载的 ONNX 模型包与断网推理 NOT_RUN；平台未启用该适配器。

以上后端、资产与控制模式的关键决定仍待相关任务明确；arXiv 与 summarize 导入方向已由用户确认。
