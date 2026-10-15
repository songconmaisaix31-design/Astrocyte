# W2 V3：七天实验工具链与协议（待派发）

用户2026-10-13已终审PASS并关账V1，基线3e1f6dae，明确V2/V3可并行；本轨实现不等待V2资产，最终锁定前验收V2/V3。新固定W2工作树`validation-v1-experiment`、分支`songconmaisaix31-design/validation-v1-experiment`，Codex`--yolo`；实际终端/Task/Dispatch待创建，不预填。root维护计划、决策、独立验收及最终Git；源码/测试/报告由W2持续完成。

write_paths：`scripts/experiment/**`、`tests/validation/experiment/**`、`docs/EXPERIMENT.md`、`docs/EXPERIMENT_REPORT.md`、`docs/experiment/**`、`tasks/M6-W2.md`。未派发前EXPERIMENT草案仍归root；派发后root仅STATUS/HANDOFF/计划，不与W2共写。生产接入CLI涉及W1写域，先Handoff给W1做限定胶水；MCP边界问题退原W0。引擎/schema/参数/权重/UI不改，不新建产品表。实验问答/评分用独立本地数据文件，遵守已有JSONL格式，不建设调度/Attempt/Manifest/证明系统。

## 交付

1. 当日生产采集的真实事件→固定seed题集，每日3–5题，同日同题ABC，事实/原因/预测全部覆盖；题目抽样规则与模板day0冻结。不可用mapper失败/未来事件调整难度。
2. A组合真实get_attention_state/explain_focus/get_experiences；B读取滚动七天完整原生user/assistant消息，保留原文，在同等冻结模型预算内最大装载，禁止截单条/摘要/静默换弱模型；C同模型同问句无记忆。冻结模型身份、上下文总预算、系统/题目/输出预留及装载次序，记录实际装载/未装载数量与原因。B预算不足时按预声明规则取最大可装内容，不能为A胜利删改B输入。
3. 原始聊天与工具参数仍本地，禁止送Git/stdout/云模型。先核实现有本地执行模型及预算能满足最强B；做可审查路由与输入构造，再报告真实缺口。不得以占位回答或弱小模型声称最强baseline成立。LLM judge独立输入/表，不复用用户评分或自评充当ground truth。
4. 事实确定性审计依真实事件库；原因用户本人裁决，用户与LLM judge两份独立表；预测D+3对照D+1至D+3实际事件活跃top-K，top-K和活跃定义锁定。种子floor不能算未来活跃，未到期不填预测正确。
5. live两客户端+三仓git增量经公共mapper/单生产Recorder，与V1R时钟路径一致；重复ID不重复贡献，每日生产Sleep。生产入口复用必要胶水由W1实施；不在Python另算引擎或批量入库后补公式。定时使用已有系统/Orca能力，不自建调度器。
6. 实验协议记录v3.0所有已裁决观察及历史修正，题库/表/成功线/模型/预算/评分全部锁定后打印全文SHA-256入STATUS。day0取真实锁定日，锁后修改判INVALID；不预填实验起止日或历史数据充前瞻。
7. 第七日按真实成熟数据产出三维统计及引用，最后三日预测尚未到D+3列PENDING，到期全对账后报告终局成功线。原因A胜B≥2/3、事实A≥B成立即首次科学验证成立；否则按维度/差距量存失败数据。两者均可完成，篡改输入/题库/评分不计结果。

## 验收与发布

仅针对上述题集同题覆盖、B全文最大预算装载、真实三臂落库/错误路径、评分三权、到期预测/协议锁定进行必要复现测试。fixture证明工具契约，不能当真实七日答案/人评。真实模型/客户端运行单列。首RED/未知回执/未执行项保存外部新独占`v3-w2/`目录，raw不出机器。SOURCE与REPORT分别commit+push，唯一原集成owner普通合；root独立复验后登记V3，用户关账缺失时禁止锁定或开V4。
