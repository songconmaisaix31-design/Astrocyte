# V0 规格修正与参数定案（v2.1，2026-10-12）

> 版本关系：本文保留 v2.1 的隐私定义和参数来源；人口分母、topic 条件、数量目标和参数绑定按 v2.2，补源与正式阶段按 v3.0 和原锁定协议执行。相应旧条件仅供追溯。

用户主控裁决，取代[M6 v2.0](M6-validation-v2.md)相应条款。原100条真实事件eligible≥90%废止：全部会话行比例测量行为密度，不测接入正确性。规格修正与实现失败分别记录，原100条首RED不改写。

## V0 验收

- 最近30天两客户端全部认知工具事件与两仓已捕获1893 commits，成功归一化为认知kind且topic非空的比例≥95%。
- 窗口内随机抽200条非认知行，误标为认知事件比例=0。
- 五工具/explain_focus任意响应，包括backend错误和已提交回执错误，不含raw_unique任何长度≥8的字符串值。

eligible = kind属于认知值域 ∧ topic属于种子∪git派生∪手工播种 ∧ strength为权重表值；正常路径不产生nil strength。种子五主题沿既有匹配规则；git Morphogenesis仓→Morphogenesis，Astrocyte仓→Astrocyte/Attention Engine，commit message关键词附加；手工来源沿SeedTopic。association规则沿W1，不新增。无标事件留存、计数进入unassigned，不硬塞、不丢弃。

## raw 审计

不新增表，store raw列原状。exposure序列化前强制投影{event_id, kind, topic, tokens, scope, ts}，白名单外不进响应。raw_unique为每事件raw载荷中除derived投影外字段值，包括消息全文、工具参数全文、base64附件和session原始元数据；逐响应字符串级断言，不将derived同值算泄漏。schema零豁免。

## 正式参数（V1–V4 冻结）

λ=0.1，α=0.3，A_floor=0.05，momentum门限=0.15。

权重：隐式深度1.0、隐式常规0.6、隐式轻度0.3、显式正向1.5、显式负向-0.8、agent成功0.8、agent失败-0.1。

λ_r=0.05；access_boost初始1.0，每次+0.1，上限2.0，24h冷却回退。禁止为历史吻合调参。既有代码单位、强度与权重字段若存在语义差异，记录实际代码和接入绑定，不改公式或重复乘权重。

## 主题来源与无标桶

v0.1验证期不做自动无标出生，发现冻结至v0.2。主题来源=种子+git派生+手工播种。unassigned留存、计数、参与Sleep输入，不产生主题。此项取代EX-01正向无标出生要求，tagged分裂继续禁用，既有更新/去噪/fluid/merge不改。EXPERIMENT尚未锁SHA，预注册草案依此合法修正，不再等待候选provider。

## V1 门槛与执行序列

先盘点两客户端全部工具调用+已捕获1893 commits，得到真实总量N。V1全量N入库；N≥5000达成数量目标，N<5000如实报告原因，不注水。按ts排序逐条走现有引擎路径，内部时钟随事件ts，真实Δt，30次模拟日Sleep；导出A(t)、每日top-5、50条权重抽检材料，真实数据重跑M2复算，焦点历史吻合由用户确认。

执行：①Codex工具映射/header继承修复；②全量盘点N；③raw白名单投影与raw_unique含错误审计；④参数冻结记录；⑤V0三指标；⑥V1全量回放/30日/轨迹/M2。V2–V4其余实验纪律沿v2.0，协议仅在V1之后锁SHA，真实七日与D+3不可用回放代替。
