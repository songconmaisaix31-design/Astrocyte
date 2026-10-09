# B0 文档基线

- 目标：将用户两份 v0.1 文档纳入当前仓库，建立开发入口、实际起点与待确认问题。
- 可改路径：`AGENTS.md`、`STATUS.md`、`docs/SPEC.md`、`docs/TASKS.md`、`docs/QUESTIONS.md`、`tasks/B0-baseline.md`。
- 依赖：用户提供的两份文件、通用执行协议与本次基线请求。
- 验收：两份源文档逐字节一致；文档本地链接和路径可解析；`git diff --cached --check` 通过；按白名单提交并 push 基线分支，远端提交与本地一致。
- 停止点：关键节点及歧义先向用户提问；依赖这些回答的实现等待决定。

SOURCE_SHA：`3a02ce38481d1885adee2e051a1b524d78376897`。

执行环境与检查结果统一记录于 [STATUS.md](../STATUS.md)。此任务交付文档基线；S0 的可启动、可构建和契约生成验收保持 NOT_RUN。
