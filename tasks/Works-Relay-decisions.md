# Works Relay 实施接缝与决策

事实源为validation登记 `a429a4d` 的v1.2。v1.0补文与实施窗口已关闭，不重问；业务线独立，validation只文档。

| 事项 | 状态 / 依据 | 下一步 |
|---|---|---|
| 实施起点 | 已核实workbench/v0.1-workbench 9d78835；相对验收d46仅3个Markdown不同；GitHub CI38440379678 head=d46、completed/success | 使用workbench/v0.2-access，不借该CI当新代码验收 |
| 持久接入store | root按v1.2状态/动作receipt落库要求准予R在独占relay_store.go建立独立relay.sqlite两表relay_state/relay_receipts；只接入状态，不修改任何既有认知schema/migrations/源/副本 | 只Service写；误配正式/副本路径拒绝、事务CAS与receipt测试由R完成 |
| 权限 | A/B/C复用既有ProjectSettings/ProjectGrant，只是确认摘要，不建立新授权系统 | 每动作重新核查；人工确认页必经；安装/已配置/能力/可用分别显示 |
| structured工具限制 | R发现既有原生会话对工具有限制；优先调查已有NativeRequest能力，未支持任务所需工具时按v1.2走terminal并记录degraded_launch | 不改native共享实现，不假称无工具会话完成接力；未知启动不降级成第二写者 |
| workspace执行权限 | root与R实码核对：既有A/B/C只资料，Settings拒绝非空AllowedTools，Codex/Claude/Pi工具禁用。已向人问是否在必经confirm页独立显示确认该中性工作区读写/执行并保持A/B/C原义；未答 | 不以start许可推导CLI自由工具，不改existing projects/native；store/无工具驱动与受控CLI预验继续，真实coding接力NOT_RUN |
| 五统计旧回归 | root据v1.2五指标要求准予F独占修改works-live.spec.ts的初始四未知预期为五未知，增加平台标签核验 | 其余旧断言/超时/skip/执行路径保持，不派额外owner |
| DataDir组装 | 已核对旧guard在WORKBENCH=1与DATA_DIR空时回写v0.1默认；root准H独占workbench_config.go抽现有explicit目录guard供relay main复用 | 旧默认与所有正式/副本/硬链接保护不变，新增relay_config回归，不改快照管道 |
| Today实时接缝 | 已向用户问：现有快照每日发布，是否可追加本地derived实时事件且保留快照数值 / 边界；尚未回答 | H调查事件接入，F先做不依赖此项的UI |
| 作品当前状态 | 已问获准未提交/未跟踪文件及非Git/单文件作品处理；尚未回答 | R先做无Agent工作区端口与互斥，不暗中缩小验收范围 |
| 参考图 / 活性窗 | 本轮未附图；已问沿既有视觉及可配置60秒近期窗口；尚未回答 | 不冻结默认窗或新视觉；用户答后同owner落实 |

## 固定owner

Orca运行 `run_d577f0d3fee1`，coordinator `term_d7f7e173-a64c-43fe-981a-dbe57124de87`。三次worker-start均 `ready` / `turnStart=observed`，未出现unknown，不重复派发。

| 轨 | 工作树 / 分支 | Task / Dispatch | 客户端模型 |
|---|---|---|---|
| R | works-relay-runtime / songconmaisaix31-design/works-relay-runtime | task_f4a5f33cc891 / ctx_fd41ff32f67d | Codex；owner核对面板GPT-6.1-Sol high fast |
| H | works-relay-http / songconmaisaix31-design/works-relay-http | task_65836fdd6887 / ctx_bcc3677fe29c | Codex；2026-10-14 Orca worker-list provider.model=gpt-6.1-sol |
| F | works-relay-frontend / songconmaisaix31-design/works-relay-frontend | task_fae5b6564156 / ctx_a7f2a326ff24 | Codex；2026-10-14 Orca worker-list provider.model=gpt-6.1-sol |

三轨固定write_paths见一页计划。早期Go/API Handoff不等于实现完成；后续SOURCE/REPORT、真实换手/terminal-only/Procmon/自动快照及最终部署各自核验。首失败与未知结果保留，仅derived汇总进Git。

## 早期发布（仍非完成）

- R接口SOURCE `3430302eea9892f1cb138dc8ee67c41f61a7078a` 已push；root发现Domain跨层依赖并退原owner，修正SOURCE `d2cf9d3c13be32369bfe353ef9907c4aa2e4cf6c` 已push，owner architecture检查PASS。store重载/互斥与真实执行尚待。
- H API候选SOURCE `f9ab2a5390b09bf4fe23df1c87cf76afb9cb4fb8` 已push；owner lint/287 examples PASS，root核对变动仅契约/examples/本轨报告、diff检查及remote一致。F已普通合候选并使用现有generator；尚无组装功能验收。
- F D1 SOURCE `f6433c6` / REPORT `9d370e0` 已push；owner typecheck、3项指标测试、双视口新旧album E2E18项PASS，均fixture/mock展示而非真实扫描/派发；首typecheck RED保留。真实换手/terminal-only/Today仍NOT_RUN。

## 跨轨复核与后续修正

- R运行时早期SOURCE `bcc60c0681f92855fd334008b95a7c73a2d7c509`、H路由SOURCE `3e9dc58d4dd89bb352eab9d662c025bd3d9e53d7`、F人工确认SOURCE `9465f3b` 已发布，仍非最终交付。H/F通过普通merge消费接口，未替换历史。
- Root发现收据成功枚举为 `accepted`，H卡片接力计数却比较 `succeeded`；退原owner。H已修正候选并报告真实HTTP + R服务 + SQLite计数/重载及12项AJV响应验核PASS，待精确修正SOURCE。
- Root要求接入store在打开SQLite写连接前以只读方式拒绝外域数据库；误置文件和硬链接测试只能使用受控测试库，不能拿V4正式库做破坏性负例。R正在修正，首问题保留，不用事后table检查冒充零写。
- Root要求sink失败保留原已接受launch收据，并从store公开单独的事件publication状态/计数；未发布不显示已进入认知。人工reconcile仅同ID derived事件接入，不能重跑Agent；Today策略仍待用户决定。
- Root准许cmd生命周期内cancellable observer只观察人工已启动的owned task，沿现有观察端口；GET仅缓存读，观察不launch/send/resume，不自动重试明确拒绝/未知动作或失败发布。内部采集cadence不决定尚未确认的活性窗口。不新增调度器。
- Native事件域无原始时间戳；首次新增sequence的观察时间是登记证据，重复缓存及重载不得刷新LastEvent。Git活动只证明变化，不证明完成；未跟踪文件同名内容二次变化尚未完整覆盖，owner记录限制。
- F新runtime mock E2E首RED为卡片高度导致接力按钮被后续标题遮挡，原owner修布局；既有断言/超时未改。owner报告104项Vitest与typecheck PASS，均不替代真实CLI闭环。
