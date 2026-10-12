# 真实成熟报告格式（空模板，不是实验结果）

锁定后按真实数据填：protocol SHA、day0、七日范围、实际model/revision/量化/template/context、source/live coverage；每日题数、ABC成功/失败/未知、B原人口/装载/未装载理由和token预算。

事实：三臂14题F1/precision/recall/漏项/虚构及事件库引用。原因：七道用户AB偏好、独立三臂winner、用户原表引用，judge独立分数/人机分歧，不覆盖用户。预测：每天D+3 due、真实top-K、score/exact_order，未到期/覆盖缺失PENDING。

第七日PROVISIONAL，最后三天预测全到期且覆盖完整才FINAL。原因A胜B至少2/3且事实A不低B才达成功线；未满足列差距与失败，缺失不缩分母。限制与未执行另列，不以fixture/回放/候选smoke替代真实验证。不附原文、工具参数、session或raw。
