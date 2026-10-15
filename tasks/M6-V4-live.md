# W1 V4：生产增量采集接缝

V1用户已关账，W2实际检查确认cmd/validation仅诊断/冻结回放，缺少每日live入口。本轨复用原W1工作树、分支、Codex --yolo终端，不重开V1或修改历史。固定write_paths：`cmd/validation/**`、`events/**`、`internal/adapters/events/**`、`scripts/validation/**`、`tests/validation/ingest/**`、`tasks/M6-W1.md`。优先仅CLI/脚本/测试胶水；公共语义表、引擎/schema/参数/权重零改。

用户授权原生源`~/.claude/projects/`含subagents、`~/.codex/sessions/`两布局、Morphogenesis/Astrocyte/McPet三仓git。显式源根、窗口和输出目录，原生header继承及wrapper分类复用公共mapper；git提交事实以仓git log为准。全源捕获后按真实ts及固定稳定ID逐条生产Recorder，s=1/w按kind绑定、种子confidence=.5，去重不重复贡献；每日生产Sleep，849历史无标仍按原冻结规则。迟到事件及错误如实计数，不静默丢弃或重写历史。

live独立state.sqlite，可由冻结V1R只读SQLite backup初始化一次以延续认知模型，源库不得写入，保护原SHA；非正式prelock验证与正式day0目录分开。不能Python动力学/生产表直接写入/重复创建主题。保留原Sleep日界及时钟推进路径，30次旧回放不重跑。完整生产副作用/回执未知保留原ID，不盲目重试。

W2边界：live库只读消费，安全coverage JSONL至少day/window_start/window_end/capture_complete/client及git计数/认知成功/无标/结构失败/真实Sleep回执；实际事件ID及ts单列安全回执。协议用既有本地文件，不建设调度/Attempt/Manifest/证明系统。具体字段/命令尽早Handoff W2并稳定，禁止跨写实验工具或协议。真实native payload只在本机新独占`v4-w1/`，不进Git/stdout/云模型。

2026-10-13 root批准W2具体Handoff：`--live-phase prelock|bridge|formal`默认prelock，formal显式传已锁protocol_sha256/day0，W2按权威EXPERIMENT锁核对；每成功capture只读backup为独占`day-state.sqlite`，供截止D日末的不可变视图。snapshot/回执失败capture_complete=false并保留已提交IDs，不接受D+1事件/Sleep污染。当前未锁，不能formal；所有字段只在CLI/本地文件，不是schema豁免。

必要复现测试覆盖公共mapper继承与窗口、跨源排序、稳定ID重复摄入无双贡献、空日Sleep、错误/未知回执；限定真实源prelock验证单列，不当正式七日。SOURCE/REPORT分别commit+push，报告精确SHA/时间/测试与真实缺口；唯一原集成owner后续普通合。root验收完整工具/模型/协议后才锁SHA/day0，当前不开始正式V4。
