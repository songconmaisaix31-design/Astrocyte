# V1R：原集成owner组合树验收

原唯一集成owner继续`C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-final-integration`、分支`songconmaisaix31-design/validation-v1-integration`，Codex --yolo。先读HANDOFF/STATUS及M6-V1R-replay.md；root只状态和独立验收，不写业务。W1正式SOURCE/REPORT发布并由root核对后派发，不将准备SOURCE5657/REPORT4fc当真实交付。

集成write_paths仅`tasks/M6-integration.md`。先普通merge本轮精确root文档基线，再普通merge派发指定W1 REPORT；保留全部SOURCE/REPORT祖先，不force/squash/rebase。业务问题交W1，冻结引擎/schema/参数/权重表零改动。本轮预计胶水0，不自行扩大写域。

核对W1文件与精确SOURCE相同、所有冻结目录与旧V1 SOURCE12cc4dd零差异。运行changed packages、原M2/EX-01/MCP暴露测试、Python ingest、kind_map生成检查、vet及git diff --check，PASS/条件SKIP分开。真实V1R已由W1执行、root独立复核；集成不另抓源、不覆盖运行目录，不把合成测试替代真实回放或用户验收。

新证据在外部`v1r-w1/`及`root-v3/`，raw/subject不进stdout、Git或网络。首RED、原capture-001浮点窗口及旧replay全部保留。普通push组合SOURCE后单独REPORT commit+push，回执给出精确SHA、时间、测试数、日志位置。

用户轨迹与新50权重核对仍PENDING；V1完整关账未完成，EXPERIMENT未锁、七日V4未开。不能代填用户结论、创建终局报告或动main/个人实例。
