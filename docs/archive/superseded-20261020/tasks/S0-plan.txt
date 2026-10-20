# S0 一页执行计划

目标：交付独立 Go + React/TypeScript 项目、浅分层包、三页 fixture、OpenAPI 与生成客户端、SQLite 迁移、开发/检查/构建入口。

起点：`de433528d360413305a4a65c997e11ef7fbf4841`。主分支：`s0/workbench-foundation`。每轨由一个 Orca Worker 在一个独立 worktree/branch 持续负责开发、测试和返修。主 Agent 维护决定、最终检查与提交。

| 轨道 | 目标与可改路径 | 依赖 | 验收 | 停止点 |
|---|---|---|---|---|
| W0 基座/契约 | 根 package/pnpm/go 配置和锁、`contracts/`、`scripts/`、`web/src/api/`、`web/package.json`、web 工具配置、`README.md`、`dependencies.lock.json`、`.github/` | 已确认 S0 | 依赖锁定、OpenAPI 校验、客户端生成；发布接入约定；最后负责少量集成胶水 | 产品规则变化先交主 Agent；不进入 S1–S6 实现 |
| W1 Go/SQLite | `cmd/`、`internal/`、`migrations/` | W0 公共契约、Go 工具链 | Go 测试/vet/build；迁移和重启；健康与空集合 API；未实现写接口明确拒绝 | 不改根依赖/契约/前端；变更需求交 W0 |
| W2 三页界面 | `web/index.html`、`web/src/`（排除 `api/`）、`web/public/`、`web/e2e/` | W0 客户端和前端配置 | 三页、显式 fixture、默认真实空态；loading/empty/error/stale/disabled；两尺寸、键盘、Playwright | 不改公共 API/锁/工具配置；不把 fixture 当研究或执行事实 |
| W3 本地工具清点 | `docs/probes/local-agents.md` | 本机现有工具 | 只读记录可调用版本和接入线索，覆盖本地 Agent、summarize；不读取密钥 | 不启动真实研究、不改 Agent 全局配置、不将版本命令当能力验收 |

顺序：W0 先发布公共契约；W3 可同时清点。随后 W1/W2 并行。集成由 W0 接回少量胶水，领域缺陷退回对应 Worker。主 Agent 在合并结果上运行最终检查并 push。

本轮客户端与实际模型：W0 使用 Codex / gpt-6.1-sol（high）；W1 使用 Claude Code / qwen3.7-max；W2 使用 Pi / qwen3.7-max；W3 使用 OpenCode / DeepSeek V4 Pro。每轨保持原负责人处理返修。

已发布并合入的轨道源码：W0 集成 `856698000fba3271d25ff68434ef4ced08e2d9fc`、W1 `72409eb259dc5c4d231c9cb56be658aae6f5faf7`、W2 `96cb7f513e9358c90e83e1cc47027d8f37a3ad1d`、W3 `59fd965ba95ba998ee957418e4981e7a010dac3a`。最终验证结果见 [STATUS.md](../STATUS.md)。

收口：原 W0 发布 Linux 测试退出修复 `0169ce2c075c94e5c850408c6f651fb0a1718cf5`，主 Agent 合入最终源码 `0a0f49c25f4c2e4a7714e80387f98c48dd7441e3`。本地检查、构建、88 项浏览器测试和独立启动通过；最终分支 Windows/Linux CI 全部通过。四轨及返修已结束，S0 PASS。

S0 验收入口：`go mod download`、`pnpm install --frozen-lockfile`、`pnpm check`、`pnpm test:e2e`、`pnpm build`、`pnpm dev`。

此次不选择真实案例，也不降低 S3 原生接续验收。Laya/Jev/索引后端的实际选型和执行控制模式留在相关适配任务；S0 不实现任务执行账本。移除重复证明检查，保留能发现实际错误的检查。
