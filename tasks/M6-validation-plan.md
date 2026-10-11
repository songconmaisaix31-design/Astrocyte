# M6 验证期执行计划

事实源：[用户任务书v2.0](M6-validation-v2.md)。基线 `6b2b91086d87e2173122c3c9f53a36332acffff2`，交付 `validation/v1-20261012`。所有开发用Codex --yolo，模型沿客户端配置，观察启动后记录实际值。root只计划/状态/决策/验收/最终Git；领域开发由固定owner提交推送，普通合并，保留原工作树。

| 轨 | 固定Agent/Worktree/Branch | write_paths | 交付与顺序 |
|---|---|---|---|
| W1 接入与回放 | terminal `term_fcb1eb9e-a1a8-49bc-af85-992ff4a5eda0`；worktree `C:/Users/DW/orca/workspaces/Astrocyte/validation-v1-ingest`；branch `songconmaisaix31-design/validation-v1-ingest` | `internal/adapters/events/**`、`events/**`、`cmd/validation/**`、`scripts/validation/**`、`tests/validation/ingest/**`、`tasks/M6-W1.md` | 先在30分钟内诊断真实50+50，不改映射遮盖首错；再kind/种子/unassigned，V0后按ts逐条正式引擎回放30天/30Sleep/JSON/50权重抽检 |
| W0 RED边界与证据展示 | terminal `term_94d760ff-7468-4e9e-b42c-44acfd4cbdb8`；worktree `C:/Users/DW/orca/workspaces/Astrocyte/validation-v1-boundary`；branch `songconmaisaix31-design/validation-v1-boundary` | `internal/adapters/sqlite/event_store*.go`、`internal/adapters/sqlite/validation*.go`、`internal/adapters/mcp/**`、`migrations/019_*`、`tests/validation/privacy/**`、`web/src/pages/validation/**`、`web/src/router.ts`、`web/src/App.tsx`、`web/package.json`、`pnpm-lock.yaml`、`tasks/M6-W0.md` | 先复现raw泄漏/接口与store边界，schema豁免由root逐条登记后才实施；V0 RED独立审计；V1后explain_focus和一张轨迹图 |
| W2 实验工具与协议 | validation-v1-experiment；V1产物后启动 | `scripts/experiment/**`、`tests/validation/experiment/**`、`docs/EXPERIMENT.md`、`docs/EXPERIMENT_REPORT.md`、`docs/experiment/**`、`tasks/M6-W2.md` | V1分布确定题库，三臂收集/确定性事实审计/独立评分表/预测D+3，V3锁SHA，真实七日与末题D+3完成才V4 |

本次P0两轨结算后的唯一集成owner沿用Codex终端`term_28cadb7a-4147-4dec-9f7a-87ecc545576d`及干净工作树`C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-final-integration`，新阶段分支`songconmaisaix31-design/validation-v1-integration`从验证交付HEAD创建，原工程分支保留。只普通合W1 REPORT a227ca5、W0 REPORT 9f654e6与必要薄胶水；领域问题退原owner，报告写域`tasks/M6-integration.md`。root拥有当前EXPERIMENT草案初始化，W2尚未启动，后续完整协议/工具交该轨并在V1后锁定。

`internal/attention/{domain,app}`默认全部只读；EX-01唯一例外归W0：`internal/attention/app/sleep_service.go`仅Run输入选择/不存在topic出生门禁，`tests/validation/sleep/**`及`tasks/M6-EX01.md`。旧期待首RED后补充测试写域`tests/attention-engine/M3_pipeline_test.go`、`tests/attention-engine/M6_week_test.go`、`cmd/sleep/main_test.go`，仅既有主题fixture预播种与tagged禁生期待，不弱化其他断言、不改M2。先查配置、报告最小类型/函数接缝后实施，数学/权重/去噪/fluid/crystallized/共享ConsolidateTopicInTx/merge保持原样。W1不得触及EX-01写域；unassigned标记与消费Handoff由两owner协调，禁止伪标签绕门禁。其他业务路径全部只读。跨轨只交Handoff；共享合约/入口若必需，由root登记唯一owner有限扩展写域，不互改。V0两个写域先并行诊断/边界复现，映射与种子实施顺序由诊断结果决定。各阶段结束普通commit+push并记录精确SHA/时间/测试数。最终单集成Agent只合精确报告与薄胶水，root独立验收/push。

本地证据放工作树外 `C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/`，raw不进Git/远程/终端正文，机器输出只含聚合或derived。未达数量、空源、未知权重和失败都记录；禁止复制/重复事件、靠挑样实现90%、把30天回放冒称七天前瞻。V1用户人工确认和原因评分只能等用户；预测要真实未来行为。

任务书冲突以真实代码诊断后集中提问，其他已裁决项不再讨论：raw字段任意值零出现与derived重用同ts/scope存在字面冲突；Sleep当前无法消费无topic且出生路径与冻结令不一致。参数必须显式记录，未确认前只做独立工作，不跑正式V1。B臂最近七天完整用户/assistant文本在预算内最大装载，保持原文，不摘要/截断单条，不默默换弱模型。P0 raw本地令与B臂模型处理必须同时成立，模型路由在V3前落实。
