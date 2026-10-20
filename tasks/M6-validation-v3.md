# Astrocyte 验证期 v3.0（2026-10-12）

> 适用范围：本文保留验证期 v3.0 的独有业务要求；锁前准备与用户确认的历史状态以原记录日期理解，实际锁、day0 和后续日流程见 M6-V4-daily-operation。开发角色和资源按当前 AGENTS、RESOURCES 执行。

用户裁决：McPet仍在开发，是实验窗口内活跃源；授权maimai-monster仓git log。沿`validation/v1-20261012`继续，冻结引擎、schema、v2.2参数及权重表。首次V1产物保留，不覆盖。

## V1R 补源重放

- 窗口`[2026-09-12T00:00:00+08:00,2026-10-12T00:00:00+08:00)`，补采McPet仓全部commit，记录捕获HEAD。原两仓1893事实不重新采集。
- git common-dir/remote归属第三条：maimai-monster及全部worktree→McPet；原会话集合按同规则重归类，报告849无标中实际转入数量。kind语义表不变，s=1/w=冻结表、λ=.1/天、A_floor=.05、种子confidence=.5。
- 逐ts单事件走生产Recorder，30次生产Sleep；M2独立复算。产出Morphogenesis/Astrocyte/McPet三线图、主题覆盖表、事件计数表，McPet commit与会话分开计数。
- 沿首次固定抽样算法及seed对新版人口重抽50条，机查s/w/贡献并给人工表；未覆盖McPet则追加专项10条。不把机查当人工正确率。
- root独立复核后，用户终审轨迹与权重，STATUS登记真实确认，再关账V1。关账后的重跑另登记，不覆盖本次。

## V2 真实库证据链

Morphogenesis、Astrocyte/Attention Engine、McPet各输出当前attention/momentum/confidence；按ts排列ts/kind/w/event_id及分kind计数；关联节点与强度。每条event_id能本地回溯raw，Markdown存`docs/evidence/`；对外仍仅derived。沿用现有轨迹展示，不新增UI。旧P2新增React/ECharts任务撤销。

## V3 工具链与预注册

当日真实事件随机题集每日3–5题，覆盖事实/原因/预测；同题三臂A接口组合get_attention_state/explain_focus/get_experiences、B七天完整用户/assistant记录在预算内最大装载、C无记忆。事实审计脚本、独立LLM judge提示词和用户表、D+3实际事件活跃top-K预测对账。

`docs/EXPERIMENT.md`逐项记录：三臂与三权分立评分（事实事件库、原因用户、预测未来行为）；原因A胜B≥2/3且事实A不低于B；EX-01及无自动发现/分裂；两次人口规格修正（100行eligible口径、全部工具认知口径）；Agent Swarm仓根折叠、Agent Memory采集源边界、MCP关键词仅标签导致零信号；项目级粒度；A(t)饱和及agent-origin权重为v0.2观察候选；McPet用户10-12确认开发中。零信号主题不补信号。模型/预算/题目/评分在day0锁定，B不得弱化。raw本地边界持续，模型执行路由必须先落实。

V1用户关账后记录EXPERIMENT全文SHA-256入STATUS；day0为锁定日，锁后修改即实验作废。锁前准备不称协议已锁或正式运行。

## V4 七天实跑

每日live采集两客户端原生目录及三仓增量git→同生产路径归一化入库→Sleep→题集→三臂→评分，D+3自动对账。第七日报告三维度胜负、成功线、原始数据本地引用；最后三天预测到期后完成终局统计。未达线如实记录差距及失败数据，作为v0.2输入。禁止以历史回放/fixture替七天前瞻；未到期预测不得填完成。

执行序：W1补源/归属→V1R→root复核/用户终审关账→MCP owner V2→新W2工具链与协议→锁SHA→七日运行/报告。冻结令其他条款不变。
