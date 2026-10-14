# 合并 release 功能与科研原文审阅

审阅日期：2026-10-14。Owner：S；工作树 `release-research-20261014`；分支 `songconmaisaix31-design/release-research-20261014`；起点 `38fc8ff4d5ae48cd5d1af02ac80b0d5eebeb44a2`。只修改本文，公开中间资料保留在本树已忽略的 `.verification/`。本次没有修改业务、其他 owner 文件、正式协议或日程，没有读取私人凭据、私人会话或正式数据库，没有调用模型、启动实验、使用视觉或派发其他 Agent。

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
