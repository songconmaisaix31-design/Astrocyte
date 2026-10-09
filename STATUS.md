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
| S0 | Windows 主路径基座验收 PASS；Linux CI 浏览器阶段待结果 |
| S1–S6、AT01–AT16 | NOT_RUN |

开发使用 Codex、Claude Code、Pi、OpenCode 四种本地 Agent；CLI 名称与实际配置的模型分别记录，不以客户端名称推断模型。

最终集成源：`856698000fba3271d25ff68434ef4ced08e2d9fc`。主分支普通合并提交：`fa7b97f659e6773d83d7dd864ddd0b0835ceac77`；与集成源的代码树一致。

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

CI 单独记录：集成源 SOURCE_SHA `856698000fba3271d25ff68434ef4ced08e2d9fc`，[运行 37897830295](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37897830295)。Windows-2025 全部适用步骤 PASS；Ubuntu-24.04 检查、race 与构建 PASS，浏览器阶段仍在运行，W0 正在诊断异常耗时。未将其计为通过。

开发预览由 Orca 终端 `term_b2100c22-c321-4344-9578-8b9f42601ce2` 启动 `pnpm dev`；实际就绪状态待检查。GitHub main 尚未合入。

## 后续适配准备

S0 按 SPEC §19.1 的基座与契约验收收口。开发规划 P0 的适配准备不随基座通过自动变为 PASS：

- P0-05：Agent 八项原生能力实测 NOT_RUN；当前报告是版本与帮助清点。
- P0-06：无 Python 向量后端选型与实测 NOT_RUN，尚未决定后端。
- P0-07：summarize 已安装；真实导出样本与字段映射 NOT_RUN。用户要求先搭环境，案例后定。
- P0-08：AOCI/CodeGraph 所选版本、许可和工作树隔离实测 NOT_RUN。
- P0-09：Laya 可直接加载的 ONNX 模型包与断网推理 NOT_RUN；平台未启用该适配器。

以上后端、资产与控制模式的关键决定仍待相关任务明确；arXiv 与 summarize 导入方向已由用户确认。
