# V4 日运行计划（root 执行；2026-10-13）

使用 Orca 内置 automation 接续已有 CLI，不新增调度器、Agent 开发轨或业务代码。仅 root 文档写域可修改；领域故障 Handoff 原 owner。权威锁预定为外部 v4-w2/locked-protocol-20261013-first.json，实际 lock 成功前不得采正式日或请求三臂。

## 每次启动的门禁

工作树 C:/Users/DW/orca/Astrocyte，分支 validation/v1-20261012。先读 HANDOFF、STATUS、锁记录与本计划，核对 EXPERIMENT 全文件 SHA、锁 authority=root/reviewed、CLI --version=0.162.0；配置或源码出现未接受改动则停止报告，不修改协议、预算、模型或评分。Asia/Shanghai 昨日为 D，day0 取真实锁，不按本计划日期猜。只在 day0≤D≤day0+9 执行；其他日期不外发 B。

## 日流程

1. 用原生产 live CLI 采集昨日完整自然日，formal 状态固定 v4-w1/formal-20261013-first/state.sqlite。输出独占 capture-D-first，带 --live-phase formal、真实锁 SHA 和 --live-day0；两客户端原授权根与 Morphogenesis/Astrocyte/mcpet 三仓路径沿 OPERATIONS，不加 --live-init。逐事件原生产路径和每日 Sleep，成功后只读日快照 day-state.sqlite。保留覆盖失败、late、unknown 回执，不盲重试或补写漏日。
2. D 属 day0 至 day0+6 才采集三臂：select-native 新输出保存当时完整两客户端选择；run-day 用本日不可变快照、coverage-safe.jsonl、真实锁与现有生产 MCP exe，输出独占 v4-w2/day-D-first。完整人口 B、原注册双通道、ABC 同模型/预算/问题；不截断、不摘要、不降模。检查真实十二臂状态，失败/STARTED 未明不自动重发。只有 paired 原答齐备后调用 judge；不给 judge 用户分数。
3. 用户另存实际 user-sheet 后才 import-user。空表不代填，匿名映射不展示给评分人。每日报告新表位置和缺失人评，不代用户原因裁决。
4. 汇总正式 capture 的原 coverage 行到本机新 coverage-through-D-safe.jsonl，仅用于现有预测脚本。对 D_question+3 已闭日的真实题调用 score-predictions，以 formal 状态及完整三日覆盖对账。D7–D9 只采集、Sleep 和成熟预测，不生成额外实验题或 B 请求。
5. 日志、输入、响应、用户表、模型流、原始错误只在外部已有证据根。不打印原聊天/答案/raw，不入产品库/MCP/Git。root STATUS 仅写日期、计数、状态、缺失与派生证据路径；正常文档提交加 skip ci、普通 push，不占 W2/协议写域。

## 结束

第七日现有 summary 只给 PROVISIONAL；末 D+3 成熟、84 答/7 人评/7 judge/全部预测齐全才 FINAL。缺失保留，不能缩分母或代写结论。终局报告由原 W2 写 docs/EXPERIMENT_REPORT.md；root 发 Handoff，唯一原集成普通合。day0+9 处理结束禁用本 automation，B 豁免在七日 B 收集结束失效，本机原库/回执归档。

若实际锁日为 10-13，首闭 D0 10-14 00:20 运行，末闭 D6 10-20，预测尾日闭 D9 10-23。机器与 Orca 需运行；平台未实际执行的日程不能记作已采集。
