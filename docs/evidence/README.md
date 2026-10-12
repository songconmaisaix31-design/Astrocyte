# V2：真实 V1R 库的完整生产 MCP 证据

SOURCE `93030768b10549542546178a006507615cdd8ba8`，精确 root `fbad72cf5257aebabe171af67098d41c3739c6d6` 已普通 merge；回放 SOURCE `7d57e06eaeaa1ddc8ddf132af47840acf8f777cc`。V1/V1R 的用户终审已于 2026-10-13 PASS 并关账，本资产不重新要求确认。

生产导出时间：2026-10-13 00:19:20.704–00:19:26.278 +08:00，5.574 秒。截止 ts 为 `1791734399999`（2026-10-11 23:59:59.999 +08:00），读取持久状态，没有按导出时钟重衰减。

| 完整资产 | 证据数 | kind 计数 | 关联节点 |
| --- | ---: | --- | ---: |
| [Morphogenesis](morphogenesis.md) | 887 | commit 887 | 0 |
| [Astrocyte/Attention Engine](astrocyte-attention-engine.md) | 1096 | commit 1006 / edit_file 74 / search 16 | 0 |
| [McPet](mcpet.md) | 23 | commit 23 | 0 |

三份表全部按 ts/event_id 排序，无长度截断。A/M 与全部证据来自生产 `cmd/mcp` 的 `get_attention_state`、`explain_focus`。现有两个共享响应均无 confidence，协调者已批准仅导出层以原库 `PRAGMA query_only=ON` 读取持久节点 confidence，三项目均为 0.5，资产逐项标注 SQLite 来源。生产 MCP/app/domain/store/schema/公式/权重/参数/UI 零改。

Agent Swarm、Agent Memory、MCP 也实际调用了生产 explain_focus，证据均 0，真实信号均 0；保持原观察。849 无标事件与 132 已认可结构残余保持原样。

原库 `C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v1r-w1/replay-002/state.sqlite` 只以 mode=ro/query_only 读取，用 SQLite backup API 创建全库副本。生产入口只打开外部独占目录内的副本；全部持久表、attention/history/boost、原件文件与 WAL 在调用前后均一致，副本启动/调用前后全部持久表也一致。原库 SHA-256 始终为 `279ffd78ad8cf083b1e279fb8c4ed32ae53a662b52d8a0f220f96d98c9816d4e`，导出脚本核对此冻结库身份，不接受 fixture 作为这份真实回放。

每项 event_id 与原库非空原始 BLOB、六字段 derived、kind/ts/weight 及完整应用历史核对；原始载荷和 traceback 索引仅保存在本机外部独占目录。完整实际证据没有用合成信号、截断列表或重新计算公式替换。

[安全审计结果](audit-safe.json)：全库 2855 BLOB，11956 个逐事件 raw_unique 字符串、7611 个不同字符串；18 个实际 MCP 响应含 7 个成功工具调用、10 个错误响应及 1 个 initialize，三份完整 Markdown，违规 0。审计沿原口径收集 UTF-8 长度 ≥8 的原始字符串，逐事件只扣除精确六字段合法 derived scalar；完整 BLOB base64 另查。完整响应、递归解码 text/JSON 的字符串与键、Markdown 均检查，不豁免同词、schema 名或冲突。错误探针覆盖五工具未知参数，以及真实缺失节点、跨 scope、agent veto、无效 report 和未知 method；没有提交有效写入。控制性后端错误与提交回执错误由原 M5/隐私回归另测，不声称在此库自然发生。

实际构建生产 `cmd/mcp` 成功；新增 Python 导出/投影/traceback 19 项 PASS；原 M5 6 顶层 PASS/1 明确条件 SKIP；MCP 13、cmd/mcp 1、privacy 11 顶层 PASS，privacy 6 个显式 opt-in SKIP。首新增 Windows SQLite 清理失败日志保留，显式关闭测试连接后复测通过。

最终另审计三资产、此入口与完整 W0 报告共 5 份 Markdown：[报告文本安全复审](report-markdown-audit-safe.json)。首次扩大审计遇一处仅历史报告的同词字面碰撞，首 RED 和匹配值保留本机；仅重述该报告字段，未放宽审计口径，历史原报告/Git 与日志保留。

可在接受这份 SOURCE 的 checkout 独立复跑：

```powershell
python scripts/evidence/export_mcp.py --database C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v1r-w1/replay-002/state.sqlite --local-root C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v2-w0 --confidence-source sqlite-readonly --publish
python -B -m unittest discover -s tests/validation/evidence -v
go test -mod=readonly ./tests/attention-engine -run MCP -count=1 -v
go test -mod=readonly ./internal/adapters/mcp ./cmd/mcp ./tests/validation/privacy -count=1 -v
```

脚本每次创建新独占 run 目录，raw 不打印、不进 Git。正式证据与本地 raw 引用目录：`C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v2-w0/run-mn895jl6/`；回归日志与首 RED 在 `run-rin8d921/`；首次探索输出 `run-mg7dbnve/` 保留。

本资产证明这份历史真实库的生产 stdio MCP 证据链和有界曝光审计。root 的普通集成与独立验收另执行；V3 工具链、协议锁定、未来七日 V4、D+3 终局结果不在本轨交付中。没有新增 UI、官方客户端验收、网络模型调用、个人运行库升级或 main 合并。
