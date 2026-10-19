# M6 V1：冻结人口的30天逐事件回放

> 文档用途：需求与历史记录。保留本文件独有业务条款、来源、实际检查和未验限制；原日期的人员分工、资源额度与操作安排只适用于当时阶段。后续执行读取 [AGENTS](../AGENTS.md)、[RESOURCES](../docs/RESOURCES.md) 和 [当前任务目录](../docs/TASKS.md)，独立 Reviewer 按准确来源验收；归档与入口调整不改变功能完成状态。

事实源为 `M6-validation-v2.2.md`；仅在root明确批准新V0后启动。固定W1、原worktree `validation-v1-ingest`、原分支 `songconmaisaix31-design/validation-v1-ingest`，客户端Codex `--yolo`。root维护计划、决策和独立验收；业务实现及返修仍归W1。

## 写域与冻结

仅 `events/**`、`internal/adapters/events/**`、`cmd/validation/**`、`scripts/validation/**`、`tests/validation/ingest/**`、`tasks/M6-W1.md`。本轮不再修改boost函数或M4 fixture；cmd入口可直接复用现有runtime包。其他路径需要Handoff说明必要性，未登记前不写。schema、domain/app动力学、权重、λ、去噪、fluid、crystallized、EX-01和UI不改。无自动出生或分裂，不新增认知能力。

## 输入与运行

1. 只读既定外部证据根 `C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012` 中冻结的 `v21-w1/tool-records-local.jsonl` 与 `diagnosis-w1/p0-git/git-log-local.jsonl`。6016父单位/8914语义单位并行保留；沿生产v22.9公共semantic mapper，不另造回放映射、重新采集git或扩大私有来源。
2. 窗口 `[2026-09-12T00:00:00+08:00,2026-10-12T00:00:00+08:00)`。预计2832 kind成功：1893 commit、539 edit_file、400 search，其中849 topic空。全部成功事件入独立LOCAL SQLite，raw仍本地既有列；132结构失败与5950非认知另报，不伪造映射或注水。
3. 六种子主题通过既有SeedTopic，InitialConfidence=.5；signal s=1、w=冻结表。λ=.1、α=.3、A_floor=.05、momentum门限=.15；既定其他参数保持。没有agent成功事件时不替未定的成功confidence政策赋值。源没有raw.facts/解释器契约时沿现有未配置路径，不生成事实或调用模型。
4. ts升序，等ts按稳定event_id排序。每次仅一个StoredEvent经过实际 `cmd/internal/attentionruntime.EventRecorder.Record`，禁止先批量入库再回算。EventService、DynamicsEngine、Sleep共用模拟时钟，时钟推进到事件真实ts，内层调用继承父ts，不发明时间。
5. 每个模拟日末调用实际 `cmd/attentionruntime.NewSleepService(...).Run`，共30次，包括10个无事件日。记录day、模拟起止ts、真实输入数量、结果/回执；未结束前不得将后日事件喂入。无标进入Sleep输入并保持pending，不生成topic。

## 导出与验收

- 本轮所有DB、原始载荷、命令全文和运行日志只落新独占外部 `v1-w1/` 子目录，不覆盖首RED/旧回放，不操作个人实例。网络/Git仅传源码和安全聚合报告。
- 导出每topic A(t)/M(t)/confidence时间序列、30日top-5排序及变化、30次生产Sleep运行记录，JSON可直接供P2展示。安全事件引用遵守既有六字段exposure投影；权重核对表单独标注s/w/贡献。
- M2在真实2832事件上独立复算：独立 `exp(-.1*Δt)`、clamp、momentum递推，不能调用生产公式生成expected或把生产输出当expected。包括每事件更新、日末去重信号与零信号衰减，以及等ts逐事件顺序；逐点误差和实测数量入报告。运行原M2套件，SKIP单列。
- 确认库中2832成功事件/信号、849无标、1893唯一commit、无双计数和实际30 Sleep。随机50条权重样本采用预声明固定seed，安全引用供人工核对；用户核对未完成保持PENDING。
- 焦点锚点如实报告：九月Morphogenesis第一，10-08后Astrocyte前二。单列模型重建结果与用户确认；没有确认不写V1人验PASS，禁止调参使其吻合。
- 适用测试须实测生产入口、按ts推进、30日（含空日）、无标持久化和独立复算；不扩展无关测试。源码SOURCE与自己的报告REPORT分别commit/push，普通历史完整，通知root精确SHA、起止时间、测试PASS/SKIP、外部产物路径和真实限制。

V1结果经root独立核对后推进V2。EXPERIMENT保持未锁，不能将回放当真实七日前瞻实验或生成终局胜负。
