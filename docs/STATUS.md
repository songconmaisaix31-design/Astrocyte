# 验证阶段状态

2026-10-14：首日 [采集日志与评分报告](experiment/DAY0_REPORT_2026-10-13.md) 已整理，统计截至 03:13:43。84 条事件、12 份答卷、6 条事实评分；两题事实宏平均 F1 为 A=5/6、B=0、C=0。独立 judge 已完成，具体评分在独立人评前封存；人评 0/7，预测待 10-17 起实际闭日对账。报告只收派生数字与状态；只读 SQL、12 条运行记录、6 条评分、3 个链接与 diff 检查 PASS，协议全文 SHA 不变，新增业务测试 0、业务构建未执行。

2026-10-14T03:13:43+08:00：V4 D0（10-13）用户授权手动恢复completed，84 commits（Astro65/McPet19）/1生产Sleep/不可变日快照；4题12臂OK、6事实机判与独立judge完成。A聚合误差0/1280引用可溯/raw0，B各完整1003消息；正式库3189事件/32Sleep，V1R原件与协议SHA不变。原01:18启动补调度的missing-project skip保留；Orca rendererReady未等待项目加载，原配置已核对恢复，下次10-15 00:20定时实跑待观察。授权客户端目录最新10-10，原生事件0，未扩源。人评待用户填写、首个预测10-17闭日成熟、v0.2仍冻结。CLI实跑与独立SQL/预算/隐私检查PASS，新增业务测试0；证据及最终Git SHA见[根STATUS](../STATUS.md)与[首日恢复](../tasks/M6-V4-first-day-recovery.md)。

2026-10-13：v0.2 R01–R12已全裁决并入[任务书v1.1](../tasks/v0.2-discovery-split.md)，评审RESOLVED，四轨[精确写域](../tasks/v0.2-discovery-split-plan.md)已登记；SPEC_FINALIZED_IMPLEMENTATION_FROZEN/D0未开始。纯文档，V4协议SHA/日程/业务不改，末D+3及人评/对账实际完成且冻结解除才开工。下面REVIEW_PENDING为历史，不重问。

2026-10-13：v0.2任务书已作冻结期文档评审，12项实码问题见 [评审](../tasks/v0.2-discovery-split-review.md)，状态REVIEW_PENDING_DECISIONS；未实现D0、未解除V4冻结或修改协议SHA/日程。V4真实七日继续，v0.2等末D+3实际完成后开工。具体待裁决读QUESTIONS最新及根STATUS。

2026-10-13T04:22:31.789146+08:00：V3实际root锁完成，day0=2026-10-13。EXPERIMENT.md全文SHA256=`5aa238599b6dd59e28d9ca30f71b6b9d547822b46d8dadc477092a46c09ab1b7`；外部权威锁v4-w2/locked-protocol-20261013-first.json，正文锁前UNLOCKED快照不再修改。SOURCE4727f7a新CI38300750305双平台完整SUCCESS，各226PASS/30条件SKIP；最终普通组合d749ecb已push，真实满人口B28.859787秒PASS、A聚合误差0/2006引用/raw0。内置日程已启用，首10-14 00:20，末预测闭日10-23；七日实际答、人评与D+3未产生，不计bridge/启动验收为正式数据。全部完整SHA、测试数、原RED和冻结记录读根STATUS最新顶部。

2026-10-13：用户对基线3e1f6dae终审PASS，V1（含V1R）已关账；轨迹与权重人工交叉全部通过，McPet主样本确认。当前放行V2/V3并行，协议仍未锁，day0未开始；下方PENDING为历史。

本文件使用同一状态源：[根 STATUS.md](../STATUS.md)。验证期冻结令、各阶段SHA/时间戳/测试数、诊断结论及schema豁免记录均在该文件顶部。当前任务见 [M6 v3.0](../tasks/M6-validation-v3.md)，人口与参数冻结见 [v2.2](../tasks/M6-validation-v2.2.md)。

2026-10-12 23:22：V1R技术交付SOURCEeba5c055已完成，精确W1源码7d57e06双平台CI通过；最终真实回放replay-002，2855事件/30Sleep/2186更新点独立误差0。用户轨迹与新50权重終审PENDING；V1未关账、EXPERIMENT未锁、七日V4未开。
