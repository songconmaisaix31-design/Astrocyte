# W0 V2：v3.0 三主题真实库证据资产

依赖V1R root复核及用户终审关账；用户2026-10-13已PASS并关账，基线3e1f6dae，本单已放行。复用原W0 `C:/Users/DW/orca/workspaces/Astrocyte/validation-v1-boundary` / `songconmaisaix31-design/validation-v1-boundary`，Codex`--yolo`。本单取代旧ECharts新增UI方案，v3.0只沿用轨迹，不新增UI。

## 固定写域

`internal/adapters/mcp/**`（仅既有exposure边界必要修复）、`scripts/evidence/**`、`tests/validation/evidence/**`、`docs/evidence/**`、`tasks/M6-W0.md`。引擎/domain/app/store/SQLite/schema/contracts/cmd/validation/前端全部冻结。若现有MCP能力已满足，MCP生产代码零改；不把导出工具变成新认知能力。其他必需路径先Handoff。

## 交付与验收

1. 新版真实V1R SQLite经生产MCP接口，查询Morphogenesis、Astrocyte/Attention Engine、McPet。现有explain_focus的attention/momentum及get_attention_state的confidence在导出层组合，不能改共享schema或公式。
2. 各主题Markdown存`docs/evidence/`：精确源码SHA/回放身份/截止ts、attention/momentum/confidence、全部证据按ts/event_id排列，每项ts/kind/w/event_id及分kind计数、关联节点与强度。仅既有derived投影，消息全文/工具参数/base64/session元数据不进Git。相关节点为空时如实写0，不推断或人工补关联。
3. 每个event_id与真实库只读审计匹配，raw本地可回溯；证据条数须与实际接口一致，不能因长度只截一段。raw_unique字符串审计生产响应及Markdown，包括错误路径；不能为通过审计弱化字段投影。
4. Agent Swarm/Agent Memory/MCP三零信号主题继续零信号，不移用三项目证据、不补论文/agent任务/用户标记；v3.0验收主题已取代旧Agent Swarm示例，原零信号仍是历史观察。
5. 只运行契约必要测试/生产真实库导出；新增脚本适用测试、原M5/真实raw审计及check/build按改动面执行。日志/raw审计只落外部`v2-w0/`新独占目录，首RED保留。SOURCE/REPORT分别commit+push，原集成owner普通合，root独立核对后登记V2。

V3题集/评分表/成功线锁SHA必须在V1用户关账之后，真实七日及D+3另验。
