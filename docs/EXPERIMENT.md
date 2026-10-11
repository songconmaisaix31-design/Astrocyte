# Astrocyte M6 前瞻实验协议

状态：预注册草案。V1尚未通过；题库、评分表和工具链尚未交付，未锁定SHA，未开始day 0。事实源：[M6 v2.0](../tasks/M6-validation-v2.md)、[EX-01裁决](../tasks/M6-EX01-decision.md)。

## EX-01：主题发现范围

本验证期主题发现机制仅覆盖unassigned出生；分裂禁用。从已标事件出生的分裂机制推迟v0.2。tagged事件继续参与去噪、fluid概念更新和merge。

回放中从unassigned自主发现的未播种主题如实记录，作为V1加号证据。McPet、MorphBench只是用户给出的例子，不预填为已发现主题。

当前实现：tagged出生已禁用，无标输入保留pending引用；真实unassigned候选来源/消费provider缺失，正向出生尚未接通。不得以0出生或测试提供候选标签宣称该项通过。实施范围与回滚对象记在[STATUS](../STATUS.md)。

## 已裁决实验设计

连续七天，同日同题三臂paired，每日3–5题覆盖事实准确、原因解释和下一步预测。A使用get_attention_state、explain_focus、get_experiences；B使用最近七天完整user/assistant消息，在上下文预算内最大装载，不摘要、不截断单条；C无记忆。B臂不得弱化。

事实由事件库确定性审计；原因解释由用户裁决；预测在D+3按D+1至D+3真实事件流活跃top-K对账。用户和LLM judge独立评分，分歧按该维度第一评分人裁决。成功线：原因解释A对B胜率≥2/3，事实准确A不低于B。

V1之后由实验owner依据真实分布补齐题库、两份独立评分表、上下文预算/模型路由、top-K与工具链，再锁定完整协议SHA到STATUS。锁定后任何修改使实验作废。七天真实回答和末日预测D+3对账完成后才产出终局EXPERIMENT_REPORT，不以历史回放替代前瞻实验。
