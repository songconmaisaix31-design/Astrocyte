# M6 V2：真实证据链与一张注意力轨迹图

依赖V1产物与root独立验收；本执行单尚未派发。继续原W0 `validation-v1-boundary` worktree/branch及Codex `--yolo`，开发、测试、报告与返修归同owner；root不写业务代码。

## 实码差异与复用

现仓为React19/Vite，已有PageFrame/App/原生History router；没有ECharts依赖或可复用图表。依v2.0只增加P2指定ECharts图及必要依赖/路由，记录该差异，其他UI不优化。现有生产MCP的explain_focus已返回event_id、真实ts、kind、s/w、衰减及贡献，confidence从同次get_attention_state读出组合展示，不改引擎/共享schema或发明证据。

## 固定写域

`web/src/pages/validation/**`、`web/src/App.tsx`、`web/package.json`、`pnpm-lock.yaml`、`scripts/evidence/**`、`tests/validation/evidence/**`、`web/e2e/m6-validation.spec.ts`、`tasks/M6-W0.md`。router既有History功能可复用，不新增无必要抽象；其他路径只交Handoff。W1同时拥有cmd/validation及scripts/validation，不写入这些路径。internal/domain/app/SQLite/store/migrations/contracts及其他UI冻结。

## 交付与验收

1. 从V1的新本机SQLite调用现有生产MCP入口，导出真实Agent Swarm及实际有证据topic的完整explain_focus；每条都含event_id/ts/kind/权重及可审计引用。消息或工具参数全文不得进入响应/导出/Git/网络。confidence与A/M由既有只读状态取得，不新增推断能力。
2. 只展示V1的30日top-5注意力与排序变化，一张ECharts时间图，通过本机选择回放derived JSON一键加载；topic、日期、A和排名可查看。无需新增HTTP端点、后台扫描、上传或个人库写入。加载前将输入投影为本图所需派生字段，排除raw/metadata。
3. Agent Swarm若真实链为空，明确输出0条；禁止复制Morphogenesis证据、按关键词另造贡献、补合成论文或假用户标注。报告V2真实证据缺口，测试或面板渲染通过不能替代真实Agent Swarm完整链验收。
4. 测试只覆盖安全投影、真实文件加载/图上30日与排序、生产MCP证据引用及raw_unique边界。浏览器实测本机回放产物，不将fixture截图当真实渲染；类型/lint/适用测试/build通过后SOURCE与REPORT分别commit/push。
5. 本机日志/证据在新独占 `v2-w0/` 子目录。原RED和原始载荷不覆盖，不将原文送模型、Git或远端。用户焦点确认保持其真实状态，未完成的V1/V2验收不改为PASS。

V3可与本轨并行准备，题库/评分表/成功线锁SHA必须晚于V1。七天三臂与D+3真实对账另验。
