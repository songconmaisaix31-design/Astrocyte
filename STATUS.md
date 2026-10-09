# Astrocyte 当前基线

日期：2026-10-09。当前任务：[B0 文档基线](tasks/B0-baseline.md)。

## 仓库起点

- 工作目录：`C:\Users\DW\orca\Astrocyte`。
- 起始分支：`main`；起始 HEAD：`3a02ce38481d1885adee2e051a1b524d78376897`。
- 远端：`https://github.com/songconmaisaix31-design/Astrocyte`。读取远端 `main` 得到同一提交。
- 初始工作树干净；唯一已跟踪文件为 `LICENSE`，内容标明 GNU AGPL v3。
- 已有另一工作树：`C:\Users\DW\orca\workspaces\Astrocyte\charybdis`，分支 `songconmaisaix31-design/charybdis`。本任务只修改当前工作树。
- 本次基线分支：`baseline/workbench-v0.1`。提交以该分支 Git HEAD 为准。

## 业务事实源

| 仓库文件 | 用户提供的原文件 | 版本 |
|---|---|---|
| [docs/SPEC.md](docs/SPEC.md) | `C:\Users\DW\Downloads\Workbench_DDD_Spec_v0.1.md` | v0.1，2026-10-09 |
| [docs/TASKS.md](docs/TASKS.md) | `C:\Users\DW\Downloads\Workbench_开发时间规划_TODO_v0.1.md` | v0.1，2026-10-09 |

两份文件按原文复制；本次未修改其产品规则、日期、切片或验收。后续规格维护入口为 `docs/SPEC.md`，规划维护入口为 `docs/TASKS.md`。

规格已有：Go 模块化单体、React + TypeScript + Vite、SQLite；Attention / Workspace / Swarm 三个业务上下文与 Judgment 共享能力；资料准入和执行批准分离；新平台构建、安装和运行无 Python；Windows 主路径，Linux/WSL 分开验证。

以上是原规格内容。开发组织、产品账本实现边界及交付节点仍有冲突，见 [待确认问题](docs/QUESTIONS.md)。26 周和各阶段日期保留为原文估算，尚未成为用户确认的排期。

## 本机工具检查

| 检查 | 实际结果 |
|---|---|
| `git --version` | `2.47.0.windows.1` |
| `node --version` | `v24.16.0` |
| `pnpm --version` | `11.27.0` |
| `codex --version` | `codex-cli 0.160.0` |
| `claude --version` | `2.1.238 (Claude Code)` |
| `Get-Command go -ErrorAction SilentlyContinue` | 当前 PATH 未找到 Go；未安装或调整环境 |

上述为当前可调用版本，不是项目依赖锁定版本。Agent 原生能力和会话控制尚未探测。

## 验证与切片状态

| 项目 | 状态 | 结果或下一步 |
|---|---|---|
| 原文复制与来源核对 | PASS | 两份仓库文件与下载原文件逐字节一致 |
| 文档相对链接/路径 | PASS | 新增文档链接及规格中的本地入口引用均可解析 |
| `git diff --cached --check` | PASS | 无空白错误 |
| S0 基座与契约 | NOT_RUN | 尚无代码、依赖锁、OpenAPI、迁移或启动入口 |
| S1–S6、AT01–AT16 | NOT_RUN | 尚未实现或执行验收 |
| 构建、测试、UI 和独立复核 | NOT_RUN | 文档基线未建立这些命令入口 |

下一步：用户回答 Q1–Q3 后，更新相应规格/规划并确定 S0 任务；Q4–Q6 的回答用于安排真实适配探测和样例。Go 环境准备在安装授权明确后执行。
