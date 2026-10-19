# P0 限界修复裁决（冻结豁免 EX-01）

> 版本关系：这是 EX-01 原始裁决。后续 v2.1 已将自动无标出生冻结至 v0.2，v2.2/v3.0 继续相应约束；本文件不单独授权当前正向出生或恢复旧待答。

来源：用户2026-10-12裁决。允许两处限定修复，定性P0 unassigned契约补全（v2 §1.3消费端本属P0），非引擎能力变更。

1. 桶读取：Sleep输入=tagged ∪ unassigned，只改输入选择。
2. 出生门禁：cluster→birth仅unassigned来源簇生效；tagged不创建主题，继续去噪/fluid概念更新/merge。

不触碰A(t)/M(t)公式、权重表、λ、去噪规则、fluid写入、crystallized追加。若merge共用出生路径，先在STATUS报路径，用收窄门禁保留merge；无法收窄停止上报。已有配置能达成优先零代码。

STATUS冻结令下记录EX-01、触及函数、保持不变、行为变化（已标事件出生实为分裂，本验证期禁用，推迟v0.2）、验证测试、回滚commit。

V0增补：生产Sleep门禁干跑tagged-only→0新主题；unassigned→出生只源于unassigned；M2重跑绿。

EXPERIMENT.md预注册：本期主题发现只覆盖unassigned出生，分裂禁用。回放自主发现的未播种主题如McPet/MorphBench按实际记录，为V1加号证据。
