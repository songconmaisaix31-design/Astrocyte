# Astrocyte 当前状态

日期：2026-10-09。当前任务：[S1 Attention 计划](tasks/S1-plan.md)，开发进行中。用户决定见 [docs/QUESTIONS.md](docs/QUESTIONS.md)。下面界面替换与 S0 验收为历史基线，不代表 S1 通过。

## S1 当前进度

主控分支 `s1/attention-materials-20261009`，基线 `d6c1bf2`。Orca Codex 五条互斥写域轨持续开发与返修，实际模型 `gpt-6.1-sol`；契约、迁移、锁和入口只有 W0 写入，主控只布置、维护决定及最终验收。

已发布阶段：W1 用例 `f7bf7fdb1c6f311e19a64c54cc1334d247b735ab`、W2 存储/导入 `7d27ccdde41e6ef57d8b82faa6d81e74e8fea5db`、W3 页面 `eeb269ec24fe53d5e45210870a83ffc75b23a26c`、W0 可运行后端 `d89f3567ebba07d6e7bc022ae37cbfb75dbeca01`。后续修复和扩展持续普通合并到 `s1-contract-1009`，尚未最终合入主控；阶段绿灯不能沿用到未验的新源码。

实际结果：W0 在 `d89f356` 的 `pnpm check` / `pnpm build` PASS。W4 已运行独立真实 API、SQLite、对象与浏览器场景，首轮 4/8 API 通过，发现数组为 null、topic/theme 阶段不一致、反馈 HTTP 状态等问题并交原轨修复；各轮结果保留于 Worker 报告，S1 最终验收未通过。真实 arXiv `2504.16054v1` 已通过官方导入并取得 16,213,397 字节 PDF；指定 B 站视频 summarize 仅取得推荐网页文字，没有字幕或真实总结，不能作为视频验收成功。

当前新增工作：人类分类的域、算法排序、项目空间 @文件引用（原分类与原文保留）、本机 Codex CLI 自动整理及原生读取隔离。Agent 默认不能读取全库；是否可读取纳入文件的间接引用仍待用户。AT01–AT04 正在验收，自动 CLI、多轮真实视频及授权 Agent 正向阅读尚未通过；S2–S6 未实施。已有 5173/8787 预览仍属于下方 UI 基线，新源码用隔离端口和临时数据库验收，不覆盖现有个人数据。

## 历史交付：界面替换

分支 `ui/preview-replacement-20261009`。源码提交 `5c1517d652afef8273cf14dab5c28021db648cfa`，已 push；S0 交接基线 `649d4297d5bd34ebd849944379cb07bf909d7acb`。

已把用户提供的 `C:\Users\DW\Downloads\Astrocyte-preview.html` 的白绿布局、品牌、顶部搜索、侧栏、三页标签、封面卡、会话卡、右栏、网络插画和详情抽屉迁入现有 React/TypeScript。保持现有 API、查询层、DTO、依赖与 Go/SQLite。开发使用 Orca Codex / `gpt-6.1-sol`，单轨独占 `web/`，主 Agent 最终检查、提交与 push。开发记录见 [前端交接](web/UI-preview-report.md)。

主 Agent 在上述源码提交、Windows/PowerShell 实际验证：

| 命令或操作 | 结果 |
|---|---|
| `pnpm check` | PASS；Go 格式/vet/测试、依赖、13 包边界、226 契约示例、生成漂移、TS、lint、37/37 单测；保留既有 `EventV1` 未引用警告 |
| `pnpm build` | PASS；`dist/astrocyte.exe` 与 `web/dist` |
| `pnpm test:e2e` | PASS，112/112，51.0 秒；1920×1080、1280×720，独立临时数据库与端口 |
| 三页 × 真实/示例 × 两尺寸预览 | PASS，12 张截图，标题、横向溢出与页面运行错误检查；目视复核布局、长标题、右栏、抽屉与拓扑示例 |

本轮 [CI 37928464780](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37928464780) 针对源码 `5c1517d` 已启动，记录时仍在运行，不沿用下方 S0 的 CI PASS。后续交付记录只改文档，不重复触发应用检查。

开发阶段 E2E 依次为 86/88、108/110、110/112，失败分别是旧标题断言重复匹配、新测试选错预算样本、空预算详情缺少明确未知标记；修正后 Worker 与主 Agent 各自另行取得 112/112。前轮失败保留在前端交接，不改记成功。

真实限制：默认继续读取真实 API，`?fixture=1` 显式示例；改为默认示例尚未决定。搜索只筛选当前页已加载数据。收藏、研究路线与事件接口尚未接入；固定研究路线/动态/拓扑只在显式示例模式出现。写入、批准、执行、原生接续与采用仍未实现，相关按钮禁用；S1–S6、AT01–AT16 未因界面替换变成 PASS。未合入 main。

当前预览：http://127.0.0.1:5173/workspace；显式示例：http://127.0.0.1:5173/workspace?fixture=1。新预览终端 `term_af3ba99d-1301-41be-be0c-860189357aaa`；API 仍为 8787，沿用仓库外数据库。停止预览在该终端按 Ctrl+C。

## S0 基线与历史验收

- 起始代码：只有 AGPL v3 LICENSE，初始提交 3a02ce38481d1885adee2e051a1b524d78376897。
- 文档基线：de433528d360413305a4a65c997e11ef7fbf4841，已 push baseline/workbench-v0.1。该提交的 SPEC/TASKS 与用户下载文件逐字节一致。
- S0 交付分支：s0/workbench-foundation。SPEC/TASKS 后续修订记录用户确认，原文保留于上述基线提交。
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

S0 当时由 Orca 终端 `term_b2100c22-c321-4344-9578-8b9f42601ce2` 启动 `pnpm dev`，就绪检查 PASS：网页、代理后的 health/foundation 均为 200。该旧终端已失效且服务已停止；当前预览终端见本页顶部。

- 网页：http://127.0.0.1:5173；三页显式示例入口在 [README.md](README.md)。
- API：http://127.0.0.1:8787/api/v1/health。
- 数据：`C:\Users\DW\AppData\Roaming\astrocyte\state.sqlite`，位于仓库外。
- 旧预览已停止；当前预览的停止方法见本页顶部。

最终验收记录为纯文档变更，执行本地链接与 `git diff --check` 后提交；不重复触发应用全套 CI。交付分支为 `s0/workbench-foundation`，已 push；GitHub main 尚未合入。

## 后续适配准备

S0 按 SPEC §19.1 的基座与契约验收收口。开发规划 P0 的适配准备不随基座通过自动变为 PASS：

- P0-05：Agent 八项原生能力实测 NOT_RUN；当前报告是版本与帮助清点。
- P0-06：无 Python 向量后端选型与实测 NOT_RUN，尚未决定后端。
- P0-07：summarize 已安装；真实导出样本与字段映射 NOT_RUN。用户要求先搭环境，案例后定。
- P0-08：AOCI/CodeGraph 所选版本、许可和工作树隔离实测 NOT_RUN。
- P0-09：Laya 可直接加载的 ONNX 模型包与断网推理 NOT_RUN；平台未启用该适配器。

以上后端、资产与控制模式的关键决定仍待相关任务明确；arXiv 与 summarize 导入方向已由用户确认。
