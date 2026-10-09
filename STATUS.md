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
| Go | 正在从官方发布安装便携工具链 |
| S0 | NOT_RUN；已拆分 W0–W3，开始搭建 |
| S1–S6、AT01–AT16 | NOT_RUN |

下一步：发布公共契约和命令约定，后端/界面并行实现；最终检查结果在此更新。
