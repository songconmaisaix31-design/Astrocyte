# Astrocyte M6 前瞻实验协议

## v2.2 预注册草案修正（2026-10-12）

协议未锁SHA，day 0未开始。第二次V0规格修正：认知人口按工具语义预声明，结构不可解析留分母失败；kind成功含unassigned接入，topic只报告覆盖率，N≥5000废止。6016全量与旧RED并行保留，不改历史结果。

主题来源=原五种子+McPet种子+git派生+SeedTopic手工播种；逐唯一cwd以git common-dir/remote识别Astrocyte及Morphogenesis全部worktree、maimai-monster→McPet。InitialConfidence统一.5。无自动无标出生、无已标分裂，冻结至v0.2；unassigned留桶计数参与Sleep。commit附加关键词沿原实现，不新增association规则。

冻结绑定：Signal s=1/w=按kind冻结表值；λ_r=.05/86400秒⁻¹；boost初始1、访问+.1、上限2、24h无访问回退1；A/M/λ/权重不变。新V0过线后真实ts逐事件回放、30次模拟日及真实M2复算；V1分布与人工锚点验收后才补齐题库、评分表、工具链并锁完整SHA。本修正发生在锁定前，属于合法规格记录。

状态：预注册草案。V1尚未通过；题库、评分表和工具链尚未交付，未锁定SHA，未开始day 0。事实源：[M6 v2.0](../tasks/M6-validation-v2.md)、[v2.1补丁](../tasks/M6-validation-v2.1.md)。

## v2.1：主题来源与发现冻结

主题来源=固定五种子+git派生+SeedTopic手工播种。v0.1无自动发现，无标出生与已标分裂均冻结至v0.2。tagged事件继续参与去噪、fluid概念更新和merge；unassigned桶留存、计数、参与Sleep输入，不产生主题。

修正原因：V0阶段规格修正，协议尚未锁SHA，合法记录。此声明取代EX-01无标自主出生及其V1加号证据条款，旧首RED保留。固定参数见[STATUS](../STATUS.md)和v2.1，不为历史吻合调参。

现有EX-01接线保留输入/pending与禁生门禁，不新增候选provider。实施范围与原回滚对象记在STATUS。

## 已裁决实验设计

连续七天，同日同题三臂paired，每日3–5题覆盖事实准确、原因解释和下一步预测。A使用get_attention_state、explain_focus、get_experiences；B使用最近七天完整user/assistant消息，在上下文预算内最大装载，不摘要、不截断单条；C无记忆。B臂不得弱化。

事实由事件库确定性审计；原因解释由用户裁决；预测在D+3按D+1至D+3真实事件流活跃top-K对账。用户和LLM judge独立评分，分歧按该维度第一评分人裁决。成功线：原因解释A对B胜率≥2/3，事实准确A不低于B。

V1之后由实验owner依据真实分布补齐题库、两份独立评分表、上下文预算/模型路由、top-K与工具链，再锁定完整协议SHA到STATUS。锁定后任何修改使实验作废。七天真实回答和末日预测D+3对账完成后才产出终局EXPERIMENT_REPORT，不以历史回放替代前瞻实验。
