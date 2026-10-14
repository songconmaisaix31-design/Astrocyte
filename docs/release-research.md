# 合并 release 功能与科研原文审阅

审阅日期：2026-10-14。Owner：S；工作树 `release-research-20261014`；分支 `songconmaisaix31-design/release-research-20261014`；起点 `38fc8ff4d5ae48cd5d1af02ac80b0d5eebeb44a2`。只修改本文，公开中间资料保留在本树已忽略的 `.verification/`。本次没有修改业务、其他 owner 文件、正式协议或日程，没有读取私人凭据、私人会话或正式数据库，没有调用模型、启动实验、使用视觉或派发其他 Agent。

2026-10-15 续审：§1–§6 保留 10-14 首轮授权、观察与首失败事实；当前授权、实现状态和新研究以 §7–§12 为准。用户已明确允许按 SPEC、skills、开源、REA 和前沿论文并行推进 S0–S6。S2/S3 新功能归 M-product 新 owner，Relay v1.4 归原 R/H/F；working 关键规则仍待答，正式 V4/v0.2 冻结保持。本轮仅公开研究与文档检查，没有启动产品、原生控制器、模型或实验。

## 1. 当前授权与事实来源

主控 mailbox `msg_a04d0bdc9c04`、`msg_a655cb31ec20` 已传达用户新确认：全部功能指资料导入与人工沉淀、项目权限与进度、认知查询与快照、Relay 人工派发、原生会话采集与 Today、Windows release 包。验收完成后通过仓库既有 GitHub Releases 发布已获授权。范围与渠道均已决定；当前任务文件仍有此前集中询问的文字，以这两条最新确认继续。

已读取主控 `AGENTS.md`、`HANDOFF.md`、`STATUS.md`、[release 任务](../tasks/Release-v0.1-20261014.md)、[规格全部章节](SPEC.md)、[开发规划](TASKS.md)、[决定记录](QUESTIONS.md)、[L1–L4 规划](../tasks/cognitive-stack-L1-L4.md)、[L1 采集计划](../tasks/L1-session-capture-plan.md)、[L2 实验草案](../tasks/L2-work-experiment-draft.md)、[v0.2 定稿](../tasks/v0.2-discovery-split.md)、[v0.3 草案](../tasks/v0.3-capability-board.md)、[科研叙事](narrative/research-alignment.md)。规格中的旧无 Python 条款已由用户允许媒体依赖的决定覆盖；产品继续使用 Go、React/TypeScript 与 SQLite。TASKS 的历史 fixture、日期与轨道人数不成为当前实现或验收要求。

领域事实来自各自工作树的 `docs/standards/backend.md`、`docs/standards/frontend.md`、`docs/access/runtime.md`、`docs/access/http.md`、`docs/access/frontend.md`、`docs/workbench/backend.md`、`docs/workbench/integration.md`、`docs/release/package-report.md`、`docs/release/windows.md`，并只读核查相关代码。报告结果标为领域 owner 已执行；本审阅没有重跑其测试。活动工作树继续变化，下列 SHA 是明确来源，最终发布输入须由主控另行确认。

| 领域 | 本次明确来源 | 已有事实与验收范围 |
|---|---|---|
| B | 第一阶段 SOURCE `2644083960649d7c50262ec69fff730b411241a2`；第二阶段 `7f58adecc41070fbec4f88962cfb6fb1723679c6` | 成熟 AST/HTML/JSON parser 与失败传播已有实现；报告仍登记全包 RED、真实网络失败及未完成整改 |
| F | 阶段 SOURCE `f534298`；Workbench/Relay 已普通消费，后续观察 HEAD `2a5c93110dabf3d52c5569987199c978eac41ee7` 含 B 生产验证 helper | 阶段 typecheck/lint/build 与 112 项 Vitest PASS；默认完整 Playwright PARTIAL；旧阶段报告的“尚未合并”不代表后续实码 |
| R | SOURCE `fbfaa65c2c01094305424ac40d2d39d271eab51b`；观察 REPORT HEAD `53231b19ddd19a65c5a47c91469a0e34d73cf600` | 真实 Codex 0.162.0、SQLite、发送/停止/恢复生命周期 PASS；恢复 raw events、其他 CLI、永久追加性仍有具体限制 |
| H | 报告 SOURCE `99a250d0b9069ffb32205420f63f82b372de5607`；观察 HEAD `094a3fee5b54b9ed4534ab4d6ec7404f2cc592b0` | 实际服务空注册基线、构建/契约检查 PASS；自产原档逐事件、同 URL 新登记与首次 DOM 正在执行 |
| Workbench | 接受的组合 SOURCE `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`；历史静态部署 `d46f9f4cf2fd6e13c8362234746105330356c160` | 既有只读查询与静态部署有工程证据；后续 L1、当前 release 组合、Procmon 与真实两次自动快照不继承 PASS |
| Q | 任务限定 `cmd/workbench-snapshot/**`、`tests/release-cognition/**`、`docs/release-cognition.md`；当前命令已有未交付代码 | 显式 source/root/history 可分发快照正在实现；本次尚未取得最终 SOURCE/REPORT 或非空认知验收结果 |
| P | 工具 SOURCE `da45adfe72a5b7da721bd16be0a1029f8fa05e66`；候选包输入 `c7a70be963c432bb72d0fa355705ce2907821120` | 候选构建和新目录 28 项检查 PASS，工具复验 24 项 PASS；当前用户说明已有手动快照入口，最终精确组合与认知/Relay 新断言仍待运行 |

## 2. release 当前可实施功能与原 owner Handoff

下列工作延续原 owner。S 不分配新写域；跨域契约或依赖变更由主控交现有持有人。源码存在、单域测试与最终组合可用性分别验收。

| 功能 | 实际实现或缺口 | 原 owner 与允许写域 | 适用真实验收与停止点 | 科研机制的适用范围 |
|---|---|---|---|---|
| 论文导入、正文/附件、重复与版本 | B 已有公开论文查询、真实 DOM 提取和 SQLite/objects；arXiv summarize 曾 fetch failed，全包尚 RED | B：既有 importers、Attention 应用、tests/s1、原 scripts；依赖仅 B | 实际官方公共来源→生产 Service→SQLite/对象→HTTP→F DOM；同源复用、真实新版本与重启保持引用；网络失败保留，不改 DNS/放行规则补成功 | 可回溯来源为检索与经验提供依据；导入成功没有证明模型工作更强 |
| 视频导入与来源定位 | 报告已有公开视频正文与附件、重启一致；实际时间区间为 0 | B 同域，F 界面域 | 复用锁定 summarize 0.25.1 及真实依赖；字幕/转写/正文来源逐项核验，缺时间位置显示缺失；新来源失败不换成固定内容 | 保留精确证据防止摘要丢失事实；缺失定位不借论文指标宣布完成 |
| 人工沉淀、分类、候选与反馈 | 既有 RecordDistillation/Review 及人工入口；部分测试替代模型/来源仍在整改 | B 原 Attention/HTTP 域；F 原页面域 | 实际导入版本→人工整理/关联或待查问题→later/reject 等反馈→真实持久化；关注与机器使用分开，零自动 Mission | 反思与长期记忆研究解释记录价值；人工反馈不自动晋级经验或扩大权限 |
| 项目 A/B/C、访问撤销、真实作品目录 | 既有项目设置与授权；发现、读取、模型处理、启动分别检查 | B 原 workspace/foundation 域，R 只核运行时权限，H 只组装，F 只展示 | 实际设置 revision、来源授权与 CSRF；撤销后拒绝新访问/启动，旧回执保留；仅获准公开 Git 项目用于本次原生验收 | Adversarial Influence 与官方事件支持明确动作边界；论文没有认证本项目全部对抗安全 |
| 项目进度与手册 | `workspace/app/progress.go` 有 InferProjectProgress、manual/agent_inferred、来源与回执；真实公开文档模型闭环尚须 B/H 独立证明 | B 生产进度域；H 仅实际服务组装；F 展示 | 同一获准公开 TASK/STATUS/README 固定版本，经真实 Native、Service、HTTP、SQLite、重启核对；percent 缺依据为空；活动时间和 R 固定回复不能代替进度 | CLM 的上下文选择与来源连续性可作设计参照；客户端存活不证明工作质量 |
| Workbench Today/Focus/Agents/Events 与证据分页 | 已有 GET Reader、六字段、nullable 历史/数值、会话鉴权；新 release 当前认知成功源需 Q | 既有 Reader 不擅改；H 原消费/HTTP 域，F 原四页面域，Q 只命令/自身测试/报告 | 使用 Q 本次实际公开生产来源快照，独立 SQL 核 A/M/C、真实时间点、分页数量和每个 event_id；无数据按空/null，样本不足保留限制；无快照 503 单列 | 认知查询是可观察功能；L2 工作收益须另行预注册实验 |
| 用户显式选择源的可分发快照 | 旧部署 snapshot.py main 绑定正式源；Q 新 Go 命令正在补用户源入口，P 已增加手动脚本说明 | Q 独占 cmd/workbench-snapshot/**、tests/release-cognition/** 与报告；P 仅打包/脚本消费；go.mod/go.sum 仍 B | source/root/history 明确、源目标分离与文件身份校验、真实 WAL 一致备份、源只读、失败不换 current.json、旧代 Reader 可继续读；不迁移正式源或编造历史 | SQLite 官方备份/URI 契约直接适用；不可将 mutable 源按 immutable 打开，也不把原子指针等同所有磁盘故障保证 |
| Relay 人工 intent/confirm、任务/终端/收据 | R 单一 store writer，CreateIntent 不启动，Confirm 重核权限；GET 读缓存，unknown 不重发 | R 原 relay* domain/app/sqlite/agents、relayworktree 与 native 观察口；H 原组装/HTTP 域 | 真实 controller 与 SQLite；人工确认、拒权/撤销、重复操作、真实停止/恢复；ready/accepted/working/done/released 各有实际证据 | Agensh 共享设施支持可持续协作；当前产品人工确认语义继续，不据此新增中心调度或自动认领 |
| 可信会话来源与完整采集 | R metadata-only owned port；H 原档与通知两种源；Codex durable records 不覆盖全部 transient 通知 | R 提供内部来源；H events/relay_session*、access_session*、workbench/HTTP、contracts 原域 | 原控制器给出精确 session/CWD/path；actual archive→逐 origin/event_id→真实库；末尾行/错误/未知保留分母，增量与重读去重；恢复 raw gap 单列 | AgentEvolver 的版本—评价—消费者关系支持逐来源核验；记录进入库尚未证明被后任务利用 |
| 新 CLI 注册→派发→Today 三分钟 | 受信任 AgentSpec 通过 encoding/json 启动加载；H 空注册服务已运行，F 正在实际 DOM | H 注册/服务接缝，R 既有原生适配器，F 既有 Today/Relay DOM | 从未登记的真实 CLI，仅注册数据新增；同 URL 注册前基线、人工接受、自产事件、首次 DOM 时间；文件/命令工具关闭；每条六字段与真实 SQLite 对账；不制造事件补分页 | abtop/AgentSight 仅有任务命名，当前尚无固定原文证据；三分钟来自用户验收要求 |
| 三值活性与生命周期显示 | R 用新序号实际接收时间，缓存重读不刷新；H 产品活性窗口无默认值，未配为 unverifiable | R 活性事实，H DTO，F 状态显示 | 新真实活动 live、正证据退出 exited、静默/联系缺失 unverifiable；R 的 10 秒测试窗口不写成产品默认；live 不提升为 done | MeshHeal 明确“可响应”与持续质量退化分离；当前不导入其委员会、检测器或恢复策略 |
| 合并前端完整默认验收 | F 已消费两线；阶段 112 项单测通过，默认浏览器 PARTIAL；新增旧模拟测试需继续整改 | F 独占 web/src/**、web/e2e/**、docs/standards/frontend.md | 稳定精确 SOURCE、现有会话/生成客户端/真实 Go/SQLite；完整默认浏览器、权限/空/null/unknown/重启及受控正向路径；禁止 route.fulfill 等代替来源 | 独立最终评估防止开发反馈自证；主控独立验收与论文自评不同 |
| Windows ZIP 与真正用户启动 | P 候选包已实际解压启动/session/CSRF/写读/重启；最终四程序与两线功能尚待组合 | P 独占 release/**、docs/release/**；唯一集成人员只补必要胶水 | 精确合并 SOURCE 构建、新目录解压、开发工具不在 PATH、静态资源/session/持久化、Q 非空认知与 Relay 权限/零自动派发；完整浏览器与原生链路分别交接 | 发布可用性由实际运行证明；科研收益不随 ZIP 构建产生 |

已发送原 owner 的具体 Handoff：B `msg_3d5835024dec`；F `msg_bf2c7fb51bc5`；H `msg_3ac6bb18f381`；P `msg_fbfaa8964984`。R 当前 `ctx_b4fcb141ca00` 已 completed，CLI 明确拒绝旧 mailbox 投递；同内容转主控 `msg_3b5846ec3e34`，如需返修由主控重启原 owner 的新 Dispatch。快速清单 `msg_2da34fd809d4` 与后续范围修订 `msg_5e5e89ee2a4b` 已发主 Run；科研机制修订 `msg_c5093326d558` 已发主控。发送成功只证明 durably enqueued，不证明收件人已接受或完成。

## 3. 真正未决定与明确冻结

| 条目 | 当前状态 | 对本次 release 的具体影响 |
|---|---|---|
| 实际用户作品、自由文件/命令工具、dirty/untracked/non-Git 中性来源 | 未决定 | 当前仅用 Agent 自建公开 Git 与关闭工具的固定上下文；产品如实显示权限/unsupported，不能代用户选根目录或开放操作 |
| 产品活性窗口、证据冲突与无原生归档 CLI 的完整覆盖策略 | 未决定或缺具体原生端口 | 保留 unverifiable/unsupported 与来源 gap，不把测试窗口成为默认，不伪造绑定 |
| Today 实时派生事件与每日认知确证值的额外融合语义、完整原件留存范围 | 尚无本次新裁决 | 已确认六字段实时展示可继续；采集层只产生事件，快照数值沿现有确证源；不生成新公式、权限或原文 exposure |
| Mission 扩展与 S2–S6 执行权 | 未获本轮范围扩大授权；现有 Mission 501 | 不将六组 release 功能确认解释为自动蜂群、批准/认领/采用全链；需主控另获具体业务授权 |
| v0.2 Discovery/Split | SPEC_FINALIZED_IMPLEMENTATION_FROZEN；R01–R12 全已裁决 | 不重问已定参数；实际 V4 末次 D+3、人评/对账完整与解除冻结之前，不实现 origin 折扣、FilterSplit、URL 新生人口、provider、schema 或校准 |
| L2 经验流水线与纵向三臂 | 草案 DRAFT/UNLOCKED/NOT_RUN；正式实现/实验门禁未解除 | 可以登记当前既有经验/知识端口和缺口；不把 release 人工沉淀升级为 Relay 自动经验抽取，不启动实验、并库或发模型请求 |
| v0.3 能力板、L3 组织与 L4 四通道 | DRAFT_IMPLEMENTATION_FROZEN | 八类资产、anti-crowding、双时标信任、scope 共现图与承诺状态仅登记；不裁决 f/g、窗口、权重、审查触发或租约参数 |
| 训练/蒸馏、MASS、Mirage | 训练红线保持；MASS 不做；Mirage 后期再讨论 | 当前不训练模型、蒸馏路由或注入幻影拓扑 |
| V4/正式验证/17321 部署/00:30 日程 | 保持原 owner 与原流程 | S 不读取库或运行材料，不移动源，不解除冻结；Procmon 与两次实际自动快照原 NOT_RUN 不变 |

L2 锁前仍需用户裁决真实任务与数量、各臂相同模型/工具/预算/顺序/污染控制、传统组资料外发、人工量表、轮次、撞坑机会、巩固保留精度/存储缩减定义与时间窗、经验可见性/撤销/冲突及未知请求对账。本文没有代填。

认知规划的 V4 独立性指科学命题不依赖胜负；v0.2/v0.3 和 L2–L4 的实际实施/实验仍服从各自明文门禁。日期到达、首日机判或当前 release 可运行均不解除门禁。

## 4. 官方科研原文、版本与适用限制

结论对应下列官方原文方法及限制章节。完整网页/PDF 获取与具体条款核验分开记录；本次未复现实验、逐附录运行代码或审核作者原始数据。原文的成绩均属于作者报告，本文不把其转成本项目指标或成功线。未核实的来源不作为已验证实施依据。

| 官方来源与引用版本 | 核实的原文机制 | 对当前功能/规划的作用与限制 |
|---|---|---|
| [Agensh，Zhihao Zhan 等，Microsoft Research，2026-09-22，2609.26781v1](https://arxiv.org/html/2609.26781v1)，§2.1–2.3、§3 | 私有 Git checkout、共享工作区、消息与只追加共享上下文支持异步自组织 | 支持已有独占写域与事实交接的设计解释；本产品仍人工派发。ProgramBench 离线/限时复现条件与个人工作台不同，不能由规模报告推导本项目收益 |
| [Context Language Models，Rulin Shao 等，UW/Meta/MIT/Trillium，2026-09-29，2609.37725v1](https://arxiv.org/html/2609.37725v1)，§1–2 | 模型自由编辑可同步到实时上下文的文件 | 用于说明外部可回溯记录与可编辑上下文的互补。产品 crystallized 只追加是自身接口语义；不宣称 OS 所有者不可修改，不把 CLM 自治/训练/SCR 导入 release |
| [Adversarial Influence，Addison J. Wu 等，Princeton，2026-09-24，2609.30028v1](https://arxiv.org/html/2609.30028v1)，§2–3.1 | 匿名共享板讨论、初始正确后转错的 defection 指标；按误导者比例与人数作受控比较 | 支持规模扩张仍核权限/预算/来源的风险判断；HLE 讨论结果不直接验证工具安全或三道硬门。具体安全靠真实拒权与隔离验收 |
| [Homogenization，Prakhar Ganesh 等，McGill/Mila/UW/Apple，2026-10-07，2610.09824v1](https://arxiv.org/html/2610.09824v1)，§4.1–4.2、§5.1 | 多轮评分趋同与共享安全盲点，在不同模型配置也观察到 | 已确认“异构客户端不保证观点异构”的设计约束；anti-crowding 是项目待验设计，原文未验证本文 C(scope) 公式。未核模型保持 unknown |
| [MeshHeal，Keru Chen 等，ASU/UH/OSU/JHU，2026-09-24，2609.29015v1](https://arxiv.org/html/2609.29015v1)，§3–4.2、§6 | 当前输出同行审查与跨任务退化检测/路由调整双时标；可响应同时可能质量退化 | 当前只支持活性与成果质量分开验收；v0.3 审查条款仍冻结。原文准确题名为 *MeshHeal: Self-Healing Decentralized LLM Agent Networks under Gray Failures*；不搬其阈值、委员会或恢复探测 |
| [AgentEvolver，Wentao Zhang 等，NTU/SMU，2026-10-08，2610.11613v1 PDF](https://arxiv.org/pdf/2610.11613v1)，§3.2–3.5、§4.1、§5、C.1–C.2 | 八类资产的版本身份、评价、留存与后续消费者分别关联；最终任务另验 | v0.3 已采纳身份连续性可以登记；保持人工晋级和现有状态。论文自评/引用检查不担保测试有效，外机逐任务原件/实际配置/完整费用未提供；编码比较配置仅一个 coordinator，不据聚合净差宣称组织收益 |
| [MAScope，Xinwen Liu 等，Beihang，2026-10-07，2610.10126v1](https://arxiv.org/html/2610.10126v1)，§3–4.1、§5 | 从带身份的消息证据恢复控制/use/feedback/pool_read 等关系，再用拓扑辅助失败诊断 | 后层可登记诊断线索。现有六字段只有 scope 共现，不能自动判为通信、因果或重复劳动；现阶段不新增图服务或 LLM 诊断 |
| [Waggle，Mingxi Zou 等，SAIS/Fudan/Yunnan，2026-09-28，2609.34136v1](https://arxiv.org/html/2609.34136v1)，§3.1–3.2、附录 P | 本地视野的承诺及反证后更新，由学习策略与固定 substrate 分工 | 维持/修改/撤销可作既有租约的后续设计输入，不能把论文状态等同 fencing 安全。匿名局部策略与 SCD 训练未授权 |
| [Generative Agents，Joon Sung Park 等，2023，2304.03442v2 PDF](https://arxiv.org/pdf/2304.03442v2)，§4.1–4.2 | recency、importance、relevance 经归一化后加权求和；reflection 引用所依据 memory records | 三维来源与解释适用；认知计划的乘号只能作维度简称，不能宣称论文使用乘法。虚拟角色模拟、embedding 与模型打分不授权修改当前冻结引擎或参数 |
| [AgeMem，Yi Yu 等，Alibaba/Wuhan，2026-01-05，2601.01885v1](https://arxiv.org/html/2601.01885v1)，§3.2–3.4、A.1 | Add/Update/Delete 管长期记忆，Retrieve/Summary/Filter 管上下文，共六操作；三阶段 RL 学习管理 | 当前既有显式经验/知识端口可参考职责划分；产品操作名称须沿自己的端口。自主 update/delete 与训练收益不替代人工批准或 crystallized 只追加 |
| [Human-Inspired Memory，Doga Kerestecioglu 等，Microsoft，2026-05-08，2605.08538v1](https://arxiv.org/html/2605.08538v1)，§3–6、§9、§11；[官方出版页](https://www.microsoft.com/en-us/research/publication/human-inspired-memory-architecture-for-llm-agents/) | 批次巩固、去重、遗忘、成熟、检索；保留精度与存储缩减分别测量 | L2 可借双指标提问。VSCode 是未来引用保留代理指标，非任务 QA；成熟/再巩固缺当前 benchmark 激活条件。六小时日程、TTL、权重及合成校准不搬进产品 |
| [MAGMA，Dongming Jiang 等，2601.03236v1](https://arxiv.org/html/2601.03236v1)，§3.1–3.4 | 语义、时间、因果、实体四种关系与查询路径；向量库和异步巩固参与实现 | 经验关系的类型可作后层设计参照；scope/时间共现没有证明因果。没有直接迁入向量库、多图数据库、路由器或新增 schema |
| [Cognitive Scaffold，Qiuyuan Ai 等，ACL 2026，2026.acl-long.1170](https://aclanthology.org/2026.acl-long.1170.pdf)，§3.1–3.3 | Fluid Working Context 与持久 Knowledge Graph 分开，结构化 event snapshot；RFT 训练压缩模型 | 已有 fluid/crystallized 术语可说明存储职责；论文采用 KG/RFT，产品只追加事实规则仍需自己的测试。训练、模型与压缩成绩不自动适用于当前接口 |
| [SRD，2610.08077v1 PDF](https://arxiv.org/pdf/2610.08077v1)，§2–3.2 | 成功轨迹提供 KNOWLEDGE，失败轨迹提供 PITFALL，事后监督事前预见的蒸馏目标 | 支持同时保留成败信息，不能由此断言所有失败轨迹信息密度最高。当前可记录既有失败候选；不实施训练或从 RL 结果推导外部经验检索收益；LEAFE 仍缺唯一引用 |
| [MASS，Hyunin Lee 等，UC Berkeley/Sakana，2026-10-08，2610.12176v1](https://arxiv.org/html/2610.12176v1)，§2–3 | 工作流优化与模型 fine-tuning 交替，多角色轨迹产生训练监督 | 用户已裁决不做；固定模型产品不继承其训练收益。本文没有新增训练或用生成 token 比例推算 release 费用 |
| [Mirage，Longzhu He 等，2026-09-29，2609.37567v1](https://arxiv.org/html/2609.37567v1)，§2–4 | 针对黑盒拓扑推断的 phantom topology 与双视图保护 | 当前继续按用户裁决留后期，不改变审计事件或生成幻影拓扑；私有 topology 威胁模型与本地六字段查询不同 |
| [OpenAI 官方事件报告，2026-08-26](https://openai.com/index/hugging-face-incident-and-the-road-ahead/)，未经授权通信及评估边界章节 | 不同评估任务通过未授权共享板形成跨任务协作，评估激励与越界行为关联 | release 要核实际权限、禁止把已答题/已生成当受控链路完成；这属于事件材料，不能称论文、填 arXiv ID 或宣称项目已防住全部此类行为 |

既有代码的实际职责也需区分：`attention/app/experience_service.go` 已有 CaptureFailure、人工 Approve/Demote/Share、approved-only GetExperiences 与 evidence resolver；`attention/domain/knowledge.go` 有 ConfirmKnowledge、AppendCrystallized 与项目可见性。本次只读核实端口，不查询 approved 实际数量。L1–L4 计划登记 approved=0 与 Sleep 次数属于规划时点记录，本次没有确认其当前数据库值。Relay 收据→Interpreter 候选→quarantine→人工批准→后任务检索/首条注入→真实消费结果的整条 L2 仍无本次交付证据。

## 5. 官方技术资料与具体实现约束

| 官方原文 | 实际核查 | 对 owner 的约束 |
|---|---|---|
| [Codex policy.rs，0.162.0 对应固定 SHA c1382380de69521303b416720a52f42d51af6248](https://raw.githubusercontent.com/openai/codex/c1382380de69521303b416720a52f42d51af6248/codex-rs/rollout/src/policy.rs) | durable persisted_rollout_item 按类型保留；raw-response、error、warning、delta、ShutdownComplete 等 transient 记录不持久化；部分持久输出有截断 | R/H 分别报告原档 retained range 与通知范围，不将 JSONL 当前完整读完称为全原生流完整。未知/错误/截断不从分母删除 |
| [Codex thread.rs，同一固定 SHA](https://raw.githubusercontent.com/openai/codex/c1382380de69521303b416720a52f42d51af6248/codex-rs/app-server-protocol/src/protocol/v2/thread.rs) | 已取得官方协议文件；本次没有逐字段重新核验全部 1948 行。R 已按同版本交付 thread/start、metadata read 和恢复限制 | 具体运行正证据沿 R 实际 controller 测试；仅 metadata path 不证明前缀或永久追加。本审阅不替 R 补签完整协议支持 |
| [SQLite Online Backup](https://sqlite.org/backup.html)，页面标明 2025-11-13 更新；[API 契约](https://sqlite.org/c3ref/backup_finish.html) | backup init/step/finish、锁、并发源更新可触发重启与一致性；持续更新可能使无界备份无法结束 | Q 保持显式来源、可取消/超时、源只读、目标独立、完成前不发布指针；验真实 WAL，不直接在线复制主文件遗漏 WAL |
| [SQLite URI](https://sqlite.org/uri.html)，§3.3 | mode=ro 控制只读；immutable 跳过锁与变更检测，对仍变化的文件可能产生错误结果/SQLITE_CORRUPT | 只有完成且此后不修改的独立代次供 Reader immutable 读取；源库不可使用 immutable 冒充只读一致性 |

本次不新增 AOCI/CodeGraph/Jev/Laya/MCP 等适配器，没有据 SPEC 附录链接冻结新版本、许可或接口。相应 S4/S5 功能需独立范围与真实资产验收；当前 release 只复用已在代码中的组件。

## 6. 获取失败、未核实来源与最终验收边界

- Firecrawl 裸命令未安装；经 `npx --yes firecrawl-cli` 可用，状态探测报告 account info fetch failed，普通官方页面/PDF scrape 随后成功。公开文档只保存 `.verification/`，没有输出或复制鉴权信息。
- 并行 requests 首次取得 Agensh、Adversarial Influence、Homogenization、MASS、Mirage、AgeMem 全部 HTML 文件；其余多个源出现 ConnectionResetError，AgentEvolver HTML 为 404。逐项首次结果保留 `.verification/primary-fetch.json`，下载成功本身不算条款核验。
- 官方 web 读取 AgentEvolver HTML / SRD HTML 失败，组合 PDF 调用曾出现 connection failed；随后 Firecrawl 成功取得 AgentEvolver 官方 PDF，SRD 官方 PDF 可读。Generative Agents v1 请求失败后，官方无版本 PDF 实际是 v2，引用按真实 v2 登记，未伪称读到 v1。
- 已取得正文的主要论文核查范围为表中方法/评估/限制章节；完整附录的每项算法、数值及作者原始数据没有逐项复现或认证。科研机制映射不作为本项目效果 PASS。
- LEAFE 尚无唯一题名/ID/原文绑定；abtop、AgentSight 与 Orca v1.4 仅有规划命名，尚未取得本任务指定的固定官方仓库版本/原文。继续登记未核实，不借标题、模型记忆或二手介绍补结论。Morphogenesis 四通道目前是自有计划引用，本次没有检查其私人运行状态或重新开启执行。
- `research-alignment.md`、v0.3 等其他 owner 文档未修改；题名、公式、六操作和训练/消融范围的准确表述写入本文，并已 Handoff 主控。主控可在其文档域消费这些结果。

S 的适用验证仅为文档链接/路径、来源版本与内容核查、`git diff --check`、唯一写域和 commit/push 后远端一致性。没有运行 Go、前端、浏览器、模型或正式实验验证；B 的 RED、F 的 PARTIAL、H/Q 的未完成与 P 的候选演练范围均保留。

最终发布仍需原 owner 的精确 SOURCE/REPORT、唯一集成人员普通 `--no-ff` 合并、适用 check/build 与完整默认浏览器、实际 Codex→自产来源→SQLite→HTTP→Today、Q 真实认知来源及 Windows 新目录最终 ZIP 验收。成果由主控在这些实际结果成立后发布；S 的文档提交不替代任何功能或科研验收。

## 7. 2026-10-15 当前研究范围与可复查基线

本轮 Task `task_ec1f2cb096b1` / Dispatch `ctx_980d42612042`，实际终端 `term_1f61ef16-79ad-49c8-b804-9f149c4fc544`；协调端为 M-release。本树接管 HEAD `7d030fc83120052a8f32ea74c6007ff9be6878ca`、指定分支、clean 均已核实。实际客户端为 Codex，`orca terminal show --terminal … --json` 显示 `GPT-6.1-Sol high fast`；M-release 消息 `msg_4794f63b988d` 另核 projection `gpt-6.1-sol`、depth 2。本轮未派发第四层。

已读取本树实际 AGENTS、HANDOFF/STATUS 当前入口、SPEC 全文、TASKS、QUESTIONS、L1–L4 和 L2 协议草案、原完整报告，以及下列当前管理与研究材料。下表是只读观察时点，不宣称已合入本树；活动 owner 后续提交由其管理者更新。

| 来源 | 观察版本与用途 |
|---|---|
| M-release `docs/modules/release.md` | 观察 HEAD `0b0b0ce750b70867c959d3327b45e19c9705b6fd`；最新 S0–S6 授权、Q/P/N 阶段范围、最终 I 尚未派发及正式冻结 |
| M-product `docs/modules/product.md` | 观察 HEAD `1b696fff3b95d8e4a6bb18365efa0977d6c97a67`；B/F/T/V 当前写域与验收。此版仍保留六组旧范围；最新 Task 明确的 S0–S6 授权优先。新功能计划路径已通过 M-release 请求，尚未收到，不据此给新 owner 增列独占文件 |
| M-relay `tasks/Works-Relay-v1.4.md` | 观察 HEAD `8c93a29ffe7df1f8c110634b3dc1b3c4c63062e9`；已读投递事实、归属、就绪、new/resume/fork/handoff 全文；working 与无 started 的人工完成仍待决定 |
| REA/Orca 既有报告 | `docs/research/orca-rea-20261014.md`，在 M-relay 树读取；复用其 1.4.220 选择范围静态结果，不重读安装包或私人状态，不重新启用 REA MCP |
| CSGHub 既有研究四文件 | X 树 `docs/research/reference-code-20261014/{SOURCE,CSGHUB-ASSETS,CSGHUB-RUNTIME-RESEARCH,HANDOFF}.md`，观察 HEAD `153b8534a8be46e4ce20c8f77a43c6bca730529e`；复用公开 SHA、许可和 REA 覆盖，并在已有公开副本复核本文采用的权限、事务及 unknown 函数 |

当前代码中，`internal/workspace/app/service.go` 的 `emptyProposalService` 与 `internal/swarm/app/service.go` 的 `emptyMissionService` 仍为空查询；`internal/adapters/httpapi/server.go:116–128` 的准入、批准/撤销、Mission、认领、提交与采用写入口仍返回 501。该事实与当前 Task 提供的 S2/S3 empty 一致。已有项目权限、Relay、选定资料读取和 Experience 服务不能替这些核心用例计完成。

`internal/workspace/domain/projects.go:71` 已有名为 `ContextPacket` 的 selected-context DTO；`app/projects.go:291` 固定读取材料 revision，检查引用与文件权限，`checkPacket:394` 在后续使用时复核访问。此 DTO 尚未具备 SPEC §6.3/§12.5 要求的持久不可变包、批准引用、完整项目快照和 successor 链。新 owner 应复用已有权限与来源读取，完成缺少的业务语义；不能仅因同名类型存在就登记 S2 Context 完成。

## 8. 当前一手论文：方法、实验条件与限制

检索及读取日期均为 **2026-10-15（Asia/Shanghai）**。使用 Firecrawl CLI `1.26.3` 的 research index 按 2026-09-01 至 10-15 查询，另用公开网页搜索定位来源；机制只依据作者 arXiv 正文与附录。两篇完整 HTML 转 Markdown 保存在 `.verification/*-full-20261015.md`；已核方法、主实验、对照、消融、限制及相关附录，不依赖摘要或搜索片段写机制。本轮没有运行作者实验、复算其成绩或调用其 reader/verifier/judge。

### 8.1 记忆使用资格：权限、生命周期与相关性分别检查

来源：[Zi Wang、Xingqiao Wang、Emmanuel Addai、Devika Ambekar、Xiaowei Xu，*The Right Memory in the Wrong Context: Verifying Retrieval Admissibility in Long-Term Agent Memory*，2610.07309v1，2026-10-05](https://arxiv.org/html/2610.07309v1)。方法见 §3，实验见 §4–§5，限制见附录 B，复现范围见附录 N。

作者把每个 memory/query 对的 scope、policy、lifecycle 分别标为成立、不成立、未解决；任一已知违规即可判不可用，信息不足继续保留未解决。namespace 来自可信身份，按身份先限定候选，再做语义排序。当前状态查询和历史变化查询分别处理旧版本；历史检索允许所需旧事实参与，当前权限仍需独立满足。存储、返回检索结果、进入实际 prompt、答案披露分别记录，prompt 暴露本身不证明模型因果使用。

作者实际条件为两个公开来源 RHELM/MemOps，共 3,767 个查询、87 个 namespace groups；top-20 是冻结排序的事后资格重评分。可信 namespace 完整、组内分数不变、必需依据位于自身 namespace 是关键前提。embedding 为固定 Qwen3-Embedding-8B revision、4,096 维、最大 512 tokens、float32、确定 ID 平分顺序。文本 verifier 另用 24 个校准与 72 个分析案例；1,523 个 paired reader 案例与 16 个构造暴露场景分别分析，不能合并分母。

作者报告 top-20 必需依据召回从 0.432 到 0.533；这属于作者数据。reader 的结果属于观察关联，召回也同时改变，不能归因为权限筛选单一因素。构造暴露试验中四个 reader 只有一个相关不可用信息的披露区间排除零。policy 标签的人评一致性弱，RHELM 评估只有七组；硬件记录不足以把候选计算减少换成延迟收益。本文不将这些结果写成本项目准确率、安全认证或阈值。

对当前代码的可实现机制是：在 `ReadProjectContext`、后续持久包分发与 approved 经验检索入口，先按受信任 caller/project、现行授权和显式分享筛选；再核版本、使用目的及适用条件。文本相似不能生成授权；旧包中的批准引用不能越过当前撤销检查。权限未知时沿既有拒绝/待补证据行为处理。论文评估中的 unresolved 与实验臂的保留策略不决定产品 fail-fast 策略，也不决定 Relay working 状态。

### 8.2 经验共享：可复用做法与接收者约束分开

来源：[Jinming Hu、Haodong Zhao、Qi Jia、Die Chen、Tianhang Zhao、Sufeng Duan、Gongshen Liu，*Learning from Others, Acting for You: Cross-User Memory Sharing for LLM Agents*，2609.32511v1，2026-09-26](https://arxiv.org/html/2609.32511v1)。ShareMem 方法见 §3，实验/限制见 §4、伦理说明和附录 A–C。

方法先在用户本地把轨迹和邻近经验归纳为编辑，再由共享管理者依据这些编辑和共享邻居独立处理共享经验。具体个人值被抽象为所需约束的语义位置；执行时从接收用户自己的偏好存储解析具体值。共享编辑由作者 LLM manager 接受，与 Astrocyte 的人工批准含义不同。scope-first 在同一优先组内合并本地/共享候选，用共同条目预算排序；低优先组只填剩余位置。这里的 scope 表示任务适用领域，不能替代 Astrocyte 的项目访问权限。

实际条件：四个 backbone 为 Qwen 3.6-27B、DeepSeek v4-flash、Gemma 4-26B、Gemini 3 Flash，检索使用 BGE-M3 和 BGE-Reranker-v2-M3；作者设经验总条目 K=5、邻居数 5、偏好数 10、每集合候选 25。Mind2Web 使用 20 个模拟用户、252 个任务的已记录网页状态预测，主配置让训练示范与评估网站归不同用户；MemoryCode 使用三个各 20-dialogue 子集，分别 395/413/420 个任务，按 AST/regex 规则评分；两者经验在评估期间固定。VitaBench 2.0 使用 56 用户、771 subtasks，时间顺序交错执行，只从 full-reward 任务在线更新，三次 trial 各自清空经验起点。

作者发现共享收益随本地同领域经验覆盖提高而缩小；近完全覆盖时主增益消失。较差轨迹来源会损害共享效果；active retrieval 消融同时改变信息和交互机会，不能只归因为检索频次。editable 与 append-only 对照还改变了抽取时是否看到旧池，不能据此否定本项目 crystallized 只追加。跨用户偏好误用审计是模型确认的已暴露失败案例，未测漏检，不能当总体泄漏率。VitaBench 的部分 full-context/rewrite 对照优于 ShareMem，说明论文没有证明记忆必胜完整上下文。

当前可借用的设计是：经验描述适用条件、做法、失败条件和所需约束；后任务使用自己的目标、批准范围和资料版本填入具体值。人工批准和跨项目显式分享继续使用已有治理入口。K=5 及模型/embedding 均为作者实验设置，本报告不选为产品默认，不引入共享原始聊天或自动改写已固化知识。

## 9. 真实开源代码入口与复用门禁

### 9.1 新论文作者官方仓库

2610.07309v1 正文直接链接 [作者仓库](https://github.com/ziwang11112/right-memory-wrong-context)。本轮普通 shallow clone、`git rev-parse HEAD` 和公开 `ls-remote HEAD` 同为 `1fca7d1554865fc3129d9869981c6af770152475`，副本只在 `.verification/admissibility-code-20261015/`。已读 README、LICENSE、pyproject、第三方声明、代码到论文映射及下列实际实现。

| 固定代码入口 | 实际机制与复用限制 |
|---|---|
| [admissibility.py:16](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/src/verify_agent_memory/admissibility.py#L16) | 三值合取；`:24` 的 lifecycle 函数在 state 或 intent 未知时返回 None，`:33` 分别合并 scope/policy/lifecycle。可参照纯规则职责，不把它接成产品授权后端 |
| [retrieval.py:467](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/src/verify_agent_memory/retrieval.py#L467)、[route:546](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/src/verify_agent_memory/retrieval.py#L546) | 可信 namespace 相等后才做本地 dense 排序；policy 试验函数保留缺少 decision 的记录，历史意图可保留旧事实。这些是评估臂，不能复制成权限缺失时允许处理的产品规则 |
| [metrics.py:146](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/src/verify_agent_memory/metrics.py#L146)、[support_controls.py:84](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/src/verify_agent_memory/support_controls.py#L84) | 通用 `score_route` 从 `usable_status=True` 推导 anchors；`score_matched_recall` 则显式接收固定 anchor_ids。若后续独立评估要保证不能因过滤改变标准答案分母，应使用明确固定的依据集合；不直接搬通用函数的分母 |
| [run_frozen_natural_support_expansion.py:160](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/scripts/run_frozen_natural_support_expansion.py#L160) | 实际自然来源评分入口读取 `query.required_memory_ids`；与通用演示路径分别核查。上游 README 的 quick start 使用 synthetic fixtures，本项目不运行或把该结果记为真实验收 |
| [analysis/exposure.py:10](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/src/verify_agent_memory/analysis/exposure.py#L10) | 区分暴露和输出，缺任一分组立即失败；函数明确无因果解释。可用来理解未来独立研究的观测单位，不能证明某条 approved 经验已经改善任务 |

根 [LICENSE](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/LICENSE) 为 MIT；[pyproject.toml](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/pyproject.toml) 要求 Python ≥3.11、基础 dependencies 为空，dev/plots 另有 pytest/PyYAML/ruff/matplotlib。无需为借鉴少量纯规则引入其完整研究环境。若实际复制代码，依赖 owner B 核固定版本与通知，发布 owner P 保留适用声明；本轮未复制代码或安装该环境。

[THIRD_PARTY_NOTICES](https://github.com/ziwang11112/right-memory-wrong-context/blob/1fca7d1554865fc3129d9869981c6af770152475/THIRD_PARTY_NOTICES.md) 和 README 明确不分发完整自然语料排序、历史 embeddings 或原 provider 响应，原始 benchmark 材料另有条款。论文人评一致性数字没有对应本轮可独立验证的完整公开人评包，代码映射文档也将其列在 validated release 范围外。上游 hash/manifest/claim 工具只属于其研究材料，本项目不引入完成证明系统。

ShareMem v1 正文与 reproducibility statement 给出方法、协议和 prompts，本轮未绑定作者公开实现仓库、固定 SHA 或代码许可证。其 arXiv 非独占发布许可不自动授予代码或模型资产分发权。因此本文只给出独立实现机制与既有本地入口，不安装同名包、不复制论文 prompts、不声称已有可直接集成库。

### 9.2 复用 REA/CSGHub 的具体结果

CSGHub server 固定 `a67b3024a9553aecf3379c9f6bccfcdada2034f5`，前端固定 `bbb9bf4b346a199a387090de76fcd6c1fffe98f6`。既有 X 已核根 Apache-2.0；本轮复核 server LICENSE 及下面源码。源码静态结论不继承上游服务实测通过。

| 公开真实入口 | 对本项目的具体采用方式与限制 |
|---|---|
| [CheckUserRepoPermission，component/repo.go:2334](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L2334) | read/write/admin 分别核资源权限，私有拒绝清缓存。可用于审阅 S2 每次动作查现行权限的位置；其 ReBAC/管理员路径不等于 Astrocyte 的人类批准，不直接增设外部授权服务 |
| [CreateMirrorRepoRecords，mirror_repo.go:91](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/mirror_repo.go#L91) | 同一事务保存 repository/mirror/task/current task及可选 job。可参照 SPEC 中 Grant/幂等回执/outbox 原子边界；外部 Git、后续权限关系与数据库没有同一原子提交保证。只复用边界思想，不引入其队列或调度 |
| [reconcileStaleExecutionLease，agent_orchestrator.go:981](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L981)、[checkSessionExecutionLease:1035](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L1035) | running 失去心跳后记录 unknown，原 session lease 继续阻止新执行。其容量 hold 到期并不确认旧执行停止；本项目复用 R 既有 operation/store 和真实停止证据，不引入 allocation/generation 新体系或自动恢复 |
| [memory handler:84](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/handler/memory.go#L84) 与 [MemMachine adapter:21](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/memory/memmachine/adapter.go#L21) | 既有研究已确认它们是输入校验与外部服务适配；scope 字段和已登录不足以证明跨项目授权，未核外部记忆批准。Astrocyte 继续使用自身人工批准/分享端口；不把外部 memory HTTP 写入当成果采用 |

许可/依赖门禁：按根 Apache-2.0 和实际文件声明核保留许可、修改声明与适用 NOTICE；不能推导模型、数据或 EE/saas 控制器许可。上游 go.mod 声明 Go 1.26.0，包含 Bun/pgx、OpenFGA 等依赖；既有报告的公开 CE sandbox/watcher 空实现、外部 MemMachine 与 Temporal/EvalScope 边界均保留。本文不引入这些服务，不把公开代码可读写成完整自治科研平台可运行。

REA/Orca 报告中的输入接受、回合开始、目标完成、归属和恢复关系，已经进入 v1.4 原 owner 规格。本轮不复制安装包代码；该包的静态分析许可不等于公开开源复用许可。autoscientist 准确项目仍待用户决定，不替用户选择同名服务。

## 10. S0–S6 与 L 系列的机制、owner 和具体验收

下表是交给当前领域 owner 的可实施机制与验收要求，依据已授权 SPEC；本轮所有新增业务验收均 **NOT_RUN**。只列已确认的行为，不固定用户尚未决定的接口字段、状态迁移、默认模型或参数。新的 S2/S3 owner 由 M-product 在安全移交后指定；本报告不接管 B/R/H/F 的文件。

| 切片/层 | 已实现基础与当前缺口 | 当前授权内可实施机制及真实代码入口 | 领域 owner 与具体验收 |
|---|---|---|---|
| S0/S1 | 基座、导入/版本、人工反馈和项目许可已有实码；完整当前组合仍有失败记录 | 复用 `attention/app/materials.go`、`distillations.go`、`workspace/app/projects.go`；论文候选和经验引用固定来源 revision，不把机器读取当人类关注 | M-product → B/V/T/F；真实公开导入→版本/附件→HTTP/SQLite→重启，旧版本仍可读；AT01–04，零自动 Mission；既有失败不覆盖 |
| S2 准入→草稿 | admission/proposal 主入口 501 | SPEC §7.1：固定 ready_for_review revision→Admission 与既有 outbox 同事务→幂等草稿；不启动 Agent。`httpapi/server.go:116/120`、`workspace/app/service.go` 为补全位置；CSGHub 事务仅提供实现参照 | M-product 新 S2/S3 owner，B 提供既有存储/依赖，F 消费精确契约；重复同候选版本只一个草稿，撤回/旧 revision 请求失败，进程启动数为零 |
| S2 批准/撤销 | 项目 grant 与 SettingsRevision 已有；Proposal 的 ApprovalGrant 尚缺 | SPEC §7/§12：受信任 human→固定提案 revision→Grant/幂等回执/outbox 同事务；每次保护动作复查现行授权。`httpapi/session.go` 与 `projects.go:authorize` 复用身份/项目边界，不把 ProjectGrant 直接当 Proposal 批准 | 新 owner + 原 B/F；AT05–07：实际 Agent token 无批准权、旧提案返回 version_conflict、批准后启动前撤销拒绝、重复启动事件同 grant 最多一个 Mission；SQLite/HTTP/重启分别核 |
| S2/S3 Context | selected_context 已有，持久不可变 successor/批准/项目快照尚缺 | 在现有 `ContextPacket`/`ReadProjectContext` 业务入口实现 SPEC §6.3/§12.5；组包前后核来源/文件变化，保存固定包引用，分发时再核权限。借用 §8.1 的检索与实际输入分别核，不新建审计账本 | 新 owner；原 R/H 核原生交付，F 展示来源与模式；旧 revision 可追溯、变更返回 context_stale、撤销阻断再次分发；两 worktree 不混用；真实原生 resume 与新 session handoff 分别通过 AT09/10 |
| S3 Mission/WorkItem | emptyMissionService，认领/提交/采用 501 | 补 SPEC 已定账本：有效 Grant→幂等 Mission，WorkItem 条件更新/单调 fence/预算预留同事务；核当前持有者、写域和实际成果再提交；用户采用生成结果回流。延用已有操作/存储，不新增自研调度或 Attempt 管理系统 | 新 owner；原 R 只提供现有执行能力，H/F 按精确接口消费；AT08/13/15：两个真实执行者仅一有效认领，旧 fence 拒绝，重启未知动作不重发，提交与采用分别持久化；来源/成果/检查可回查 |
| S3/L1 Relay v1.4 | R/H 正在实现，当前链路未完成 | §9.2 与既有 REA 支持输入接受/开始/人工完成分开、归属和未知交付保持；继续现有 `relay.go`、`relay_lifecycle.go`、`relay_store.go`、`relay_session*.go` | M-relay → R/H，F 唯一 UI 操作者；真实 Codex/SQLite/HTTP/自产来源、工具关闭、event_id 逐一定位、同次 firstDOM 三分钟。working/无 started 完成路径待用户答复，不能以论文决定 |
| S4 | RepositoryMap/CodeSearch/ImpactQuery 为 SPEC 适配方向，本轮无新验收 | 将检索结果绑定实际 repo/worktree/snapshot，权限筛选先于语义检索，旧索引标 index_stale。`internal/adapters/indexing` 是 SPEC 规划路径，本树尚不存在；应由领域 owner 补消费方端口和适配实现。AOCI/CodeGraph 的具体所选版本、许可和执行输入仍需核定 | M-product 指定对应 owner，依赖交 B；真实两 worktree 不串结果、修改未提交源码后返回过期、撤销移出可检索集合；不因论文 namespace 实验宣布索引隔离通过 |
| S5 | SPEC 已定义 Judgment 契约；本树 `internal/judgment` 只有 `doc.go`，未取得 Jev/Laya 实现、资产或运行结果 | 沿既有目录补用例与适配器，权限/题型/允许后端先检查，拒答、截断或资产缺失保留实际状态。论文生成式 verifier 不替代确定性批准；不把 BGE/Qwen 参数锁成新后端 | M-product 指定判断 owner，B 管依赖；AT11/12：拒答/截断不自动准入，模型/目标/策略变动缓存失效，实际 Laya 资产许可/断网/重启逐项验收；缺资产保持未完成 |
| S6/L2 功能 | `CaptureFailure`、人工批准/分享、approved-only 检索已有；Relay→候选→批准→后任务消费链缺本次证据 | `experience_ingest.go` 与 `experience_service.go:71/198/229`、`experience_store.go` 复用；只用实际已提交 Event，unknown 不造失败候选；按 ShareMem 分离做法与接收项目当前约束；L2 适配输入先交 owner 审查冻结边界 | 新功能 owner 联原经验域 owner，M-release 路由；真实失败→quarantine、Agent 无批准权、人工批准后第二 Agent 可检索并在实际任务使用；证据逐 event_id 核。未分享项目不可见，批准/撤销版本冲突真实返回；AT16 不以“已检索”替代成果使用 |
| L2 巩固/研究 | Sleep/Knowledge 已有；经验接入与独立三臂仍为草案/未运行 | `sleep_knowledge.go`、`knowledge_{fluid,crystallized,retrieval}.go` 仅登记入口；保留 fluid/crystallized 各自语义。固定后任务适用机会、输入条目与结果；保留精度/存储缩减分开定义 | 认知原 owner 与根；允许范围内设计可继续，正式认知 schema/engine/权重与 V4/v0.2 不改。独立协议裁决/锁定/实际解冻后才真实实验；人类质量评分、客观检查、真实 usage 与撞坑回溯分别验收 |
| L3/L4 | 既有规划提供组织/群体命题，当前无本轮实现证据 | §4 已核八类资产连续性、双时标、异构趋同及 scope/预算/审计边界继续作为研究输入；需要将任务适用性与权限分开，不从共现推断协作或因果 | 由根按现有 L 任务指定 owner；本报告只给设计依据。能力板、策略/训练、跨库合并和正式实验仍服从其冻结/未决门禁；不启动共享原始聊天、自动认领或新的控制器 |

具体区分：用户对 S0–S6 系列的新授权允许相关独立产品设计与实现推进；正式 V4/v0.2 的冻结依然有效。L2–L4 涉及冻结核心、并库或实验的部分，须满足原任务门禁。产品功能可以接受真实业务验收，科研效果还需独立实验。论文中的“自动接受编辑”“历史可检索”与“scope优先”均不产生批准、资料外发或扩大工具的权利。

## 11. 已交领域缺口与仍待决定

只向当前 M-release 交接，由其路由 M-product/M-relay 原 owner；没有联系旧 Dispatch、Slack 或 GitHub 消息渠道。

| 交接事项 | 接收与下一步 |
|---|---|
| 新计划准确路径/SOURCE 与安全写域 | `msg_83abd45e4b67` 已送 M-release；其 `msg_4794f63b988d` 确认已请求 M-product，研究按 SPEC 独立继续。未收到准确计划前不代定新 owner 文件集合 |
| S2/S3 empty 与同名 ContextPacket 的真实覆盖 | `msg_bb9a8a5e146a` 已交 M-release：转新 S2/S3 owner 及原 B/F/R/H，按 §10 实施/验收；零新业务修改 |
| 论文代码的权限缺省策略与分母 | 同消息明确 generic score 与固定 source anchor 的不同入口；后续独立 L2 评估设计使用自己的真实任务依据，不搬作者合成 fixture/manifest 系统或 0.8 阈值 |

working 严格开始证据、无 started 客户端的人工完成路径、实际用户作品与自由工具范围仍未决定。L2 真实任务集合、预算/模型/工具/顺序/污染控制、传统输入外发、人工评分、机会分母、巩固双指标及统计方法继续以原草案待答项为准。本报告不填默认值，不把 V4 B 臂豁免扩给新实验。

## 12. 本轮检查、提交范围与未执行项

本轮允许写入仅 `docs/release-research.md`。公开 research 检索、arXiv 全文和作者代码副本保留于已忽略 `.verification/`；未把其他树材料、第三方源码、数据库、会话、凭据或研究环境打入提交。原报告的首段与首失败事实保留。

适用检查为：报告本地相对链接/新增固定代码链接对应公开副本及行号、`git diff --check`、相对接管 HEAD 的单文件写域、首段保留、普通 SOURCE/独立 REPORT 提交及 push 后完整远端 SHA。准确 SOURCE 与检查结果在随后独立 REPORT 追加；任何检查未执行不预填 PASS。

未执行：产品 Go/前端/浏览器检查与构建、作者测试或实验、真实模型/原生控制器/SQLite/HTTP 验收、正式采集/Sleep/judge、部署和发布。没有读取私人会话/凭据/正式库；实际 approved 数量、正式认知状态和当日科研结果未核，不从规划历史值推断当前值。新功能计划准确来源、ShareMem 官方代码/许可与模型/数据资产分发许可尚未核定；对应实现或分发仍由领域 owner 在使用前完成检查。
