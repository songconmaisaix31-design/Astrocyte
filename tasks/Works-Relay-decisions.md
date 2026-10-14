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

## 工程阶段收敛（用户最终验收不变）

- R SOURCE `a3f47921f4627a8b3e48e44c7cabe0d2b021b617` / REPORT `2e8112669deffd9b85017fc81cc665e7863bb030` 已push、clean且remote一致。owner报告20个有测试Go包全套、架构/vet/build PASS；受控真实CLI贯穿Service/SQLite/neutral Git/derived sink，源文件未改。首10秒context超时RED保留，不以单项诊断PASS覆盖。真实Agent和正式认知闭环仍NOT_RUN。
- H SOURCE `8d353e3381cd620834ee64aea357611267f0d27d` 普通消费R `cf11b1a`；observer和shutdown仅已有task、cancellation join与关闭顺序，14项真实HTTP响应AJV PASS。首CI `38455064422` exact `3e9dc58` 为Linux无LOCALAPPDATA测试fixture失败、Windows生成schema漂移；原run不重跑覆盖。H只修测试根和缺失环境负例，生产路径约束保持，远端新CI结果尚待。
- F收到实际receipt枚举、exact-root项目匹配、权限摘要与done/released UI接缝，仍原owner修。A/B/C不是新的scope降级接口；新项目默认A，已获准B/C准确展示并确认。
- 正式认知流缺口与Today展示分别验证：本机pending observation接入不等于normalized event/engine更新或正式快照发布；旁路展示不能冒充完整闭环。H给可复用正式采集方案，不改mapper/引擎/权重。
- 可独立工程完成后，交唯一集成owner合精确SOURCE/REPORT并验构建和接口。阶段交付不宣称v1.2完成，不降低用户工程验收；未答自由tools、来源、活性窗及Today仍保持门禁。原固定领域owner持续负责返修。

## 后续联调与原owner返修

- F首工程阶段Task `task_fae5b6564156` / Dispatch `ctx_a7f2a326ff24` 已报告succeeded；SOURCE `2cdcf9dc49722c63aad69ee59f54da36bc17c4f2` / REPORT `ab3d8e27d182d2c28e368454fa9e66af6e8272d5` 已push，owner报告generator/typecheck/lint/build、103 Vitest及42项适用E2E PASS。真实公开样本扫描2项PASS（平台codex、1作品、2文件、手册重载），不替真实业务全部作品或Agent验收。Root看过公开样本1920/390截图，五指标及动作可见，无横溢；接力任务服务未启用时如实显示未知。
- F已由Root立即复用同一已证明terminal/worktree/branch，后续Task `task_12347e40c49f` / Dispatch `ctx_52ef7f772849`；旧Dispatch已inactive，cleanup ownership已转移，禁止向旧mailbox派发。返修仅安全显示互斥现有task提示和适用UI验核。
- H互斥安全提示SOURCE `7e167026593cd510f2f6e62aa4b3d737531ade64`：仅VersionConflict、canonical非零UUID且同human Caller经runtime公开GetTask确认时保留view_existing_task提示；原receipt及重载一致。任意provider action/path/raw、未知UUID和无关冲突不透传。F消费此源码。
- Windows目录短名firstCI `38457889686` / `38458172696` RED保留。R SOURCE `af6e76d1083c73b3fb3d28d0e3c915e99ee014cb` 改以真实目录identity比较，H SOURCE `0c21364218eb4dc0ae379a5a66942532e986b460` 普通合入；本机真实8.3 alias及不同目录拒绝PASS。后来remote Go包检查通过至schema漂移处，不是完整CI绿；race/build/E2E尚未跑。
- Root新增必要审查退R：真实新Git登记事件应计入活性，重复观察不续命、多来源LastEvent不倒退、exited优先；同Agent新task复用terminal_id时应创建新native session，未完成task resume不重派。另要求复核operation与target task绑定：原实现已把TaskID封入输入，Root最初疑点未成立；跨task负例属于验证补充，不记录为原缺陷或首RED。
- R仅以本机公开help/version核实真实CLI stdin/EOF/TTY兼容性，未运行自由工具；H据独立能力证据准备初始注册数据，注册不授权、不派发、未知不宣称可用。
- H核实正式collector为validation的runLiveCLI/captureLive，要求已授权native根、时间窗、cwd/session与真实Git身份；不会消费relay.sqlite或独立pending观察。正式路径覆盖与真实启动证据未核实，认知闭环仍BLOCKED/NOT_RUN，validation代码/mapper/权重不改。

## 最终工程候选与集成门

- R最终SOURCE `b19d83635fe10cde9d42642d7444f3771d02f513` / REPORT `4848a98eef45656f30326054f3da8298f61044b1`，原Task已按独立工程阶段succeeded结算。新task的新native session、Git证据单调、structured默认安全降级、可选stdin_eof及无变化观察零Update均有适用回归；完整Go/架构/vet/build PASS。真实Codex/Claude、terminal-only Agent coding和认知闭环未验收。
- F最终SOURCE `b48a6e5008b4da087a5d19c0a091b487197e7494` / REPORT `fdcf21b25dc3e8c84d2aa39b772a381374b50135`，同owner返修Task已succeeded。安全显示真实互斥task提示，terminal stdin_eof明确禁用发送；6项适用Vitest、16项适用E2E及生成/typecheck/lint/build PASS。E2E包含真实HTTP公开响应回放与mock，不是实际模型启动。
- R与F由Root明确retained：依据用户“开发、测试、文档和返修由同一个Worker持续完成”，保留同terminal/worktree/branch用于后续领域返修。新工作需新Task/Dispatch，不给已完成Dispatch发送；未结算或静默不能释放。
- H最终SOURCE `c7c3255d0530b93312bb541a477b9ec47e0afc15` / REPORT `6e211dd3e3d92cecadd6a513f7bb3b41075f661c`已push、clean/remote一致，原Task独立工程阶段succeeded结算并按同一固定owner要求retained。实际TCP/SQLite/受控EOF进程回归、289契约例子、26实际HTTP JSON独立AJV、vet/build/架构PASS。注册表当前启动时读取，追加条目需重启服务才能显示，不能宣称热更新。独立公开样本应验证同一测试URL重启后列表扩展且调度代码不变。
- 三轨结算后唯一集成Worker从root最新分支创建，普通no-ff合精确REPORT（保留SOURCE祖先）；只允许少量入口/路由/类型/配置胶水及自己的集成报告，领域问题退原owner。先完成可独立的全工程检查，再由Root复验、提交与push；用户待答语义不自行配置，17321已验实例不提前替换。
