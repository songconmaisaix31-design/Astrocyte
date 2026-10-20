# Astrocyte M6 验证期任务书（v2.0 · 从代码到证据）

> 版本关系：本文保留 v2.0 的独有需求和历史验收。人口、隐私、参数与出生条件按 v2.1、v2.2、v3.0 及原锁定 EXPERIMENT 的后续对应条款执行；旧 100 行 eligible 门槛、数量目标和 schema 豁免不作为当前门禁。

来源：用户2026-10-12下发任务；以下保留任务要求与验收，不变更裁决。工程期结束，验证期开始；M0–M5骨架保持现状，不新增认知能力。北极星是持续演化的个人认知模型，本阶段产出证据：证明Astrocyte比普通Agent Memory更懂一个长期使用者。已裁决项不重新讨论。从 `attention-engine/v1-20261011` 拉 `validation/v1-20261012`。

## P0：真实数据接入

先诊断再修复，30分钟内取100条真实事件（Claude Code/Codex各50）过adapter，记录topic无值域、kind非认知值域、metadata授权各过滤层，结论写STATUS。诊断决定kind映射与种子实施顺序，两者都做。

确定性硬编码映射放 `events/kind_map.py`，一行一规则，注释只写规则：

| 来源 | 规则 | kind |
|---|---|---|
| Claude Code | tool_use(name=Edit或Write) | edit_file |
| Claude Code | tool_use(name=Grep或WebSearch) | search |
| Claude Code | assistant消息含arxiv/pdf，阈值规则 | read_paper |
| Codex | rollout按同构规则 | 同上 |
| git log | commit | commit |

种子主题5个：Morphogenesis、Agent Swarm、Astrocyte/Attention Engine、Agent Memory、MCP。确定性关键词+scope路径匹配。无标事件不丢弃、不硬塞，进入unassigned桶，仅Sleep消费；新主题只能从unassigned桶诞生。

RED三层：raw原始载荷在本地store独立表，仅审计路径可读，不出机器；derived为kind/topic/tokens/scope/ts/event_id，引擎与检索可用；exposure的MCP只暴露derived引用+event_id，任何响应无raw。report_event外部上报与对外查询共用exposure边界。

V0：100真实事件全链eligible≥90%；任何MCP响应序列化文本不出现raw载荷的任何字段值，由审计测试证明。

## P1：30天回放

来源：`~/.claude/projects/`（含subagents）、`~/.codex/sessions/`（新旧布局）、Morphogenesis/Astrocyte git log。最近30天，约09-12至10-12，覆盖Morphogenesis后期、比赛冲刺、Astrocyte立项。

按ts排序逐条喂入，禁止按摄入时间批量灌入。内部Clock跟随事件ts、delta t真实间隔。按模拟日触发30次Sleep，产出30条运行记录。回放与实时同代码路径，无回放后门。

交付：≥10000 eligible入库；每topic一条A(t) JSON序列；真实数据重跑M2可复算验收；每日top-5排序迁移；抽检50条权重由人核对正确率。

V1：Morphogenesis九月后期第一焦点，10月8日后Astrocyte/Attention爬升前二；是否吻合由用户人工确认，这是重建研究迁移的证据。

## P2：证据链与轨迹展示

explain_focus(topic)输出Topic/Attention/Momentum/Confidence、逐条带时间戳/kind/说明/weight/event_id的Evidence、Related topic及强度；每个event_id可审计回溯原事件。任务示例Agent Swarm关注来源为09-28论文、10-02代码、10-08提交、10-09user_mark、10-10agent_task，不把示例伪造成真实事件。

复用现有驾驶舱React+ECharts，只一张30天×top-5 topic注意力轨迹（时间轴+排名），无其他UI。V2：回放产物一键渲染；真实explain_focus("Agent Swarm")完整时间证据链。不可解释的attention等于不存在。

## P3：七天三臂前瞻实验

同日同题paired：A为get_attention_state+explain_focus+get_experiences组合；B为最强普通agent baseline，用最近7天JSONL提取完整用户/assistant消息，预算内能塞多少塞多少，不截断不摘要；C为无记忆floor校准难度。弱化B唯一不可撤销作弊，赢弱baseline无效。

每日生成器从当日事件流随机构造3–5题，覆盖：

| 维度 | 题型 | 第一评分人 |
|---|---|---|
| 事实准确 | 最近项目/论文 | 事件库确定性审计脚本 |
| 原因解释 | 为什么长期关注某topic | 用户本人独裁 |
| 下一步预测 | 未来3天深入topic | D提问/D+3真实行为 |

预测命中为D+1至D+3事件流活跃top-K，ground truth采集真实行为，零人工标注。用户+LLM judge独立双评，分歧按维度第一评分人，禁止模型自评充当裁决。

day0锁成功线：原因解释A对B胜率≥2/3；事实准确A不低于B。跑完才可改。

V3交付：当日题集生成器、同题三臂收集器/回答入库、事实审计脚本、LLM judge提示词和独立用户评分表、D+3预测对账脚本；`docs/EXPERIMENT.md`锁题库/评分表/成功线，打印SHA进STATUS；之后任何修改实验作废。V4：7天×3臂×每日题集完整落库，全部预测对账，`docs/EXPERIMENT_REPORT.md`三维度胜负统计。

## 冻结令与纪律

V0–V4：schema零改动，P0接缝确需schema逐项在STATUS记录豁免，否则revert；不新增feature、不接Morphogenesis（接口v1.1预留）、不优化UI（仅P2图例外）；bug修复附复现测试。冻结令置顶STATUS，对后续会话生效。

V0→V1→V2→V3→V4；V3可与V2并行准备，但锁晚于V1，题库依真实分布。阶段结束 `docs/STATUS.md` 记SHA+时间戳+测试数。仓库不符任务书以真实代码为准，在STATUS记偏差。

红线：不新增认知能力、一行引擎逻辑不改（P0映射/种子仅接入层）；不训练/调权重（校准冻到实验结束）；不弱化B；按ts逐条，不批量灌；不做不存在规模的抽象，测试只覆盖验收。注释/commit/STATUS/报告只推进实现或传递信息，删除无动作的免责声明与套话；报告只写结论、数字、胜负。

完成：报告达到成功线则第一次科学验证成立，为论文/产品提供基础；未达线如实记录差距维度/量，失败数据入库存档供校准/v0.2。两种结果都完成；篡改任何臂输入、题库、评分不算结果。
