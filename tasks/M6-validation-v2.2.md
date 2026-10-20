# V0 认知人口重定义与参数定案（v2.2，2026-10-12）

> 版本关系：本文件的语义人口、隐私继承和冻结参数仍为原 V4 依据；v3.0 补充 McPet 来源与后续阶段，实际日运行还须核原权威锁。新协作规范不改变本文件的业务参数。

主控裁决，覆盖v2.1相应条款。第二次规格修正：废止两客户端全部工具调用均为认知人口；第一次废止100行eligible≥90%。规格修正与实现失败分别记录，6016全量留存计数及旧RED不改写。

## 1. 预声明语义人口

| 归类 | 全量统一规则 | 处理 |
|---|---|---|
| 认知 | Edit/Write/apply_patch→edit_file；Grep/Glob/rg/find/web search→search；arxiv/论文读取→read_paper；git log adapter→commit | 计入分母，须映射成功 |
| 非认知 | Read普通文件、ls/tree、git status/diff/log命令、测试/构建/安装、进程/任务控制、callback、计划/待办、Bash非搜索非提交 | 剔出人口，单独计数留存 |
| 结构不可解析 | 剥不开的wrapper、动态参数等 | 留分母，计为失败 |

Bash git commit不映射commit：命令不等于提交事实，提交仅用已冻结git adapter1893条，禁止双计数。2520混合桶必须剥开内层按语义分类，不能将解析失败伪装非认知。直接工具612报告工具名分布，值域内认知工具漏映射须修映射表。禁止按mapper结果划分或微调规则；规则版本化，修改后全量重跑。W1报告各桶计数、每桶20条样本，交root抽查。

## 2. V0

归一化完整率=认知语义事件成功映射kind且按§4绑定strength / 认知语义事件总数，门槛≥95%；唯一允许残余为结构不可解析。topic不再是完整率条件。主题覆盖率=kind成功且topic非空 / kind成功，仅报告、不设门槛。unassigned的kind成功即接入成功，留桶计数。

随机200非认知行误标0、raw_unique零泄漏维持；原PASS证据保留，新口径全量验收重跑。raw、schema及exposure边界沿v2.1。

## 3. topic

每唯一cwd解析git仓根，以remote/common dir识别：Astrocyte及全部worktree→Astrocyte/Attention Engine；Morphogenesis及其worktree→Morphogenesis。新增McPet种子，仓根maimai-monster。原五种子与commit关键词附加Agent Swarm/MCP沿用。所有播种InitialConfidence=.5。

## 4. 冻结参数绑定

Signal贡献=s×w；接入s=1，w按kind查冻结权重表，禁止s=w=表值。λ_r=.05/天，秒⁻¹绑定为.05/86400。access_boost初始1，每次访问+.1，上限2，24h无访问回退1。现有仅累加无cap/冷却属于v2.1参数未落地，W1本轮补齐并逐函数登记STATUS；不动A(t)/M(t)/权重/λ，schema零豁免。

## 5. V1

入库全部kind成功事件，含topic空的unassigned。原N≥5000废止，认知总数如实报告。按ts排序逐条喂入同一生产路径，引擎时钟跟随事件ts，30次模拟日Sleep，真实M2复算重跑。焦点锚点：九月Morphogenesis第一；10-08后Astrocyte前二。吻合由用户确认，禁止为历史吻合调参。

## 6. 执行序列

1. W1 wrapper嵌套解析、612拆桶、预声明语义规则全量应用，报告各桶。
2. W1 git仓根归属与McPet种子，重分类unassigned。
3. W1参数四处绑定，含boost cap/冷却及复现测试。
4. root抽查每桶20条；发现规则错误须版本化全量重跑。
5. V0新口径完整率≥95%、误标0、泄漏0。
6. 过线启动V1。

其他冻结条款继续：schema零豁免、引擎动力学零改动，无自动出生或分裂，不提前锁实验协议。
