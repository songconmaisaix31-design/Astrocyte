# 预印本首轮独立审查

2026-10-19。决定：`accept`，仅接受完整方法与实验协议初稿、文献核查、主张证据分级和中文科研反馈。当前候选没有已确认、需要退回 Writer 修改的初稿缺陷。正式实验、完整统计程序、产品运行、PDF 编译、arXiv 兼容性与投稿均未通过本次审查。

## 准确对象与独立性

| 项目 | 实际值 |
|---|---|
| 研究候选 | `cf867ce6e9bcf156fd3e036f0ac48a4834113b6b` |
| Draft PR | [#3](https://github.com/songconmaisaix31-design/Astrocyte/pull/3)，目标 main，审查时 headRefOid 与候选一致 |
| 产品只读候选 | `7ce3a5efe312f4cdac3491293e002f5559f076b0`，业务来源 `982333b871520e606c74d553e448c5049f81abde` |
| 工程报告 | `7f781deea143c74bed5303f90d8130c82bb54bca` 的 `docs/reviews/dynamic-dag-20261019.md` |
| Reviewer 分支 | `songconmaisaix31-design/research-preprint-review-20261019`，起始 `3a02ce38481d1885adee2e051a1b524d78376897` |
| 实际客户端 | Codex CLI 0.162.0；Orca provider `codex / gpt-6.1-sol`，终端 GPT-6.1-Sol xhigh fast；没有模型参数覆盖 |
| 生命周期 | Run `run_d59e41c6d2f1`，Task `task_a0c647db8962`，Dispatch `ctx_8bc7f5be8bbe`，终端 `term_a494186b-4a9c-467d-a366-0f0184ef4e80` |

实际读取治理树 AGENTS、HANDOFF/STATUS 的当前科研、产品和 V4 交接章节、研究及审查任务、研究产品决定和 QUESTIONS 的原生/S3 条件。完整读取两份 Downloads 原件：文件名为 `Astrocyte_论文定位稿_2026-10-18 (2).md` 的正文实际是 2026-10-19 定位稿 v3；另一份为同日实验预注册开发草案 v0.1。作者身份、正式模型、任务、阈值、预算、样本量及注册继续待定。

读取并实际应用研究树 `.agents/skills/astrocyte-research/SKILL.md`：按原始来源审查论证，区分证据等级，检查因果识别、完整序列重复、预算和信息隔离。另应用 Orca orchestration 技能处理真实交接与决策通知。本次单个独立会话，研究树及产品树只读；写域只有本报告、自身 `.gitignore` 的 `.verification/` 条目及其忽略内容。

## 来源与元数据复核

Reviewer 独立联网取得 18 个原始响应，均为 HTTP 200：六篇指定版本 HTML 全文、六个指定版本元数据页、三个原稿无版本元数据链接、MAAGL PDF、EvoMap 整页和 ASA PDF。首次下载 UTC 01:43:12–01:43:53，原稿无版本链接补查 UTC 01:46:51–01:46:53。使用 Requests 2.33.1、BeautifulSoup 4.14.3 和 PyMuPDF 1.28.2 提取；原始响应、完整提取文本及时间回执只保存在自身 `.verification/sources/`。

逐项比较全部九条 BibTeX 记录。论文的标题、作者和首次提交/所读版本日期与原页对应；MAAGL 的名单差异按各个载体分别记录。细读范围为下表对应方法、正文讨论及附录，核查稿件实际引用的论断；没有复现这些论文的实验，没有读取图像，也不声称通读了与引用无关的全部附录提示、代码清单和参考文献。Firecrawl CLI 在 PATH 不可用，实际下载和解析使用上述工具。

| 条目与版本 | 元数据复核 | 原文支持与引用边界 |
|---|---|---|
| [Strategy Genes v4](https://arxiv.org/html/2604.15097v4) | Junjie Wang、Yiming Ren、Haoyang Zhang；首次 2026-04-16，v4 为 2026-09-17 | §3、§4.3 明确固定模型参数、外部经验表示和失败信息的不同组织方式。支持正文第28行的设计背景，不提供 Astrocyte 的收益证据。 |
| [EvoMap Swarm](https://evomap.ai/wiki/10-swarm) | 官方文档，无可核不可变版本及发布日期；本次访问日期 2026-10-19 | “Collaboration History & Synergy”描述协作历史 EWMA，“Team Formation Algorithm”把历史协作纳入匹配。正文第38行将其限定为设计背景，没有移用文档中的效果保证。 |
| [MPT v1](https://arxiv.org/html/2610.22322v1) | Guowei Xu、Zhenting Qi、Amani Kiruga、Himabindu Lakkaraju、Yilun Du；2026-10-15 | §2 实际训练行动与竞价参数；§3.4研究群体规模和成员不可用。正文第36行正确区分其参数训练与本稿固定参数、外部状态更新。 |
| [DGM v1](https://arxiv.org/html/2505.22954v1) | Jenny Zhang、Shengran Hu、Cong Lu、Robert Lange、Jeff Clune；2025-05-29 | §3和附录F支持代码修改、候选档案及评价记录被候选修改的风险。无版本原链接当前为2026-03-12的v3，稿件有意固定v1并已说明；未把后续版本内容归入v1。 |
| [ShareMem v1](https://arxiv.org/html/2609.32511v1) | Jinming Hu、Haodong Zhao、Qi Jia、Die Chen、Tianhang Zhao、Sufeng Duan、Gongshen Liu；2026-09-26 | §3支持两阶段整理、范围检索和偏好隔离；附录B.1/B.3/B.4分别讨论整理成本、覆盖和来源。正文第30行仅提取相应研究动机。来源中表号交叉指引有错位，按章节和表题核查，没有移用效果数字。 |
| [EDV v1](https://arxiv.org/html/2606.24428v1) | Shiding Zhu、Yudi Qi、Yajie Wang、Jiaze Li、Chao Song、Yaorui Shi、Yibo Miao、Hanqi Gao、Kai Zhang；2026-06-23 | §4确认异构执行、第三方提炼和原执行群体共识；附录G明确共识偏差。正文第32行及证据准入正确保留模型共识与可独立核查证据的区别。 |
| [MAAGL v1](https://arxiv.org/html/2609.09565v1) | 2026-09-09；[摘要页](https://arxiv.org/abs/2609.09565v1)列 Liang Qu、Jianxin Li、Hua Wang；HTML 与实际13页PDF首页另列 Hongzhi Yin | §IV-C式15–16使用历史轨迹估计置信和选择协作者；§IV-D式18–20提炼条件经验并验证。正文第34行准确承认该机制已有先例。BibTeX采用实际文稿四人名单并注明索引差异；差异原因仍未知。 |
| [ASA 原链接](https://www.amstat.org/asa/files/pdfs/p-valuestatement.pdf) | American Statistical Association，2016-03-07，实际三页新闻稿 | 已读取全部三页。正文第232行只用其支持透明报告及效应大小解释；它不推导本稿配对重采样、Holm、等效性或样本量方法。正式声明文章没有在本次另行取得全文，不算已核技术依据。 |
| [准确工程报告](https://github.com/songconmaisaix31-design/Astrocyte/blob/7f781deea143c74bed5303f90d8130c82bb54bca/docs/reviews/dynamic-dag-20261019.md) | 报告Commit、产品候选和既有CR链接与所读实际文件对应；BibTeX没有虚构个人作者 | 完整读取报告；支持人工DAG、有限前端检查及未验范围的转述。本次未重跑产品测试、访问其SQLite或重新验证其中事件。 |

没有发现当前稿件以搜索摘要代替上述细节依据、把模型共识当成客观正确，或声称首次让历史经验参与伙伴选择。作者索引差异及尚缺统计技术依据已同时出现在正文、来源记录、BibTeX说明或反馈中，不需要凭猜测补齐。

## 科学设计与实际实现核查

下列位置均指准确候选的 `papers/astrocyte/manuscript.md`。

| 检查与位置 | 证据与判定 |
|---|---|
| 稿件性质，第7、12、289、303、307行 | 摘要、方法、结果状态、局限及结论一致表述为拟议机制与未注册协议。作者/机构/ORCID待填，无虚构数值、结果图表、注册编号或投稿状态。A/B/C仅为拟议案例。 |
| 八配置及关闭语义，第127–144行 | L/C各00、10、01、11完整。相同非空初始经验和关系快照、隔离状态、配对任务序列；经验关闭冻结内容、验证、使用统计及历史排序，关系关闭冻结计数、权重及衰减；即时状态和权限继续生效。对缓存、索引、摘要和共享服务的间接污染有明确限制。 |
| 公平控制及位置干预，第90、102、144、146行 | 共用评分/softmax/探索、对应成员的信息视图、候选、检索和决策额度、认领及冲突服务、总并发额度均保留。同步回放只检查同状态同随机数的一致选择；实际位置效应依赖状态年龄、等待、提交顺序、冲突与通信机制，并保留零或负结果。部署尚未形成可运行干预。 |
| 关系更新及证据，第66–88行 | 公式与原件一致，q=0不增加成功/失败；更新开启时仍可发生时间衰减，关闭时完全冻结。具体时钟、可见时点和延迟顺序待登记。事件去重、同任务同结果归并、未归因事件保留，任务总成功不能直接给每条边记功；折扣分数未被称为校准后验。 |
| 三种估计量，第178–216行 | 更新交互、位置对交互的影响、位置对共同更新总收益的影响各自保留。代数分解正确；L11–L00大不直接证明交互，交互为正也不保证超过单独更新的实际采用价值。 |
| 重复单位与推断范围，第218–226行 | 重采样单位为完整八配置配对块，保留任务序列与初始化绑定；任务、八个标签、重复检查和多次分析不会增加独立样本量。同任务集/快照的重复只支持条件推断，恢复隔离副本不等于独立抽取历史。云服务漂移和任务依赖被保留为限制。 |
| 区间、实用阈值与多重比较，第222–242行 | 0.03仍是草案实用目标；双侧95%下界对应名义单侧2.5%边界，样本量示例使用z0.975。Holm三项公式正确，但有效边际检验、零假设校准、重采样区间、等效性程序及其检验族保持未定，稿件没有给出调整显著性或等效结论。无试运行方差，不能计算正式r。 |
| 盲评、排除和预算，第152–174行 | 盲评分数对执行及更新过程隔离至完整序列结束，说明反馈可能泄漏同一答案；外部损坏排除配置对称，算法失败保留。超时、预算耗尽和未执行的后续任务仍在计划分母。模型/工具/提炼/验证/检索/协调/重试/初始化计费，支付与资源用量分列，未知不记零。快照实际生成和分摊分列；r=5/T=10/30次调用仅为预算算例。 |
| 机制解释及A/B/C，第244–266行 | 距离是观察变量，过程相关性不能独自证明因果；另行干预与主结果分开。案例允许改变代码但不混入主实验，封存结果不能反馈至后续候选轮次；没有伪造逐轮表。 |
| 主张证据与中文反馈，第268–305行及配套文档 | P/I/E/R分类、R1–R7与P1–P3符合当前研究产品决定。工程DAG、JS字节与没有Worker的停止没有变成自主规划、成员退出、质量或成本收益。完整产品CI有真实失败及取消，当前未通过；V4原件不可用，历史汇总未进入研究结果。 |

实际只读 `policy.go` 的 `SelectWorkItems`，确认版本 `swarm.fixed.v1`，过滤后按 MatchingInputs 和 WorkItem ID 排序；没有据此发现拟议概率策略。读取 `experience.go`、`experience_service.go` 的 human-only治理与approved-only读取，以及 `swarm/app/experience.go`、`experience_accepted.go` 的来源核对与去重，确认验证强度与批准权限必须分别处理。检查 `commands.go` 的 ClaimWorkItem 当前绑定、能力和预算检查。对准确产品 `internal/` 搜索七项研究字段无匹配，只支持所查路径内未找到相应接口，不能推出全仓绝无同类能力。

## 正式研究前置条件，按影响排序

这些事项已在初稿中明确公开，属于正式研究的未验范围；它们没有被本次初稿接受改记为完成，也不产生新产品或权限授权。

| 优先级、具体位置 | 影响与必要完成条件 | 可复核方式 |
|---|---|---|
| 高：第90、127、146、152行；中文反馈R1–R4 | 共用概率策略、更新隔离、条件/时钟/归因和盲评访问尚未实测。缺少时不能把八标签解释为受控干预，位置差异也可能含信息或资源混淆。 | 产品Worker按确认方案实现后，真实检查每种配置的状态读取、候选/概率/随机数、部署及评价访问；先完成等条件回放，再分别核部署机制。 |
| 高：第218–242行；R7 | 尚无可执行的正式区间、有效差异检验、等效检验及重复次数。小r、任务依赖或共享服务相关性可破坏精度与检验能力。 | 获取独立试运行配对块方差，固定抽样总体、区间和校准方法、检验族、r/最大次数及停止规则；以真实试运行核查分析，不能用虚构数据替代。 |
| 高：第160–174、283行；R5与P1 | 计量、原生文件/命令工具关闭和S3预算/权限证据仍有未满足条件。预算公式是计划量；8rT包含预算耗尽后未执行的计划任务位置，实际尝试数和重试必须另记。 | 先取得现有待答的明确决定及合法执行条件；逐运行核真实服务、SQLite、原生会话与每个event_id，再核全部可观测成本、未知成本和快照分摊。 |
| 中：第34、232行；sources.md | MAAGL索引差异原因、统计技术文献与正式方法依据仍未核全。 | 保持固定版本和四人文稿名单的说明；正式注册/投稿前补充原始统计来源及对应实施程序。ASA新闻稿继续仅支持报告原则。 |
| 中：README.md的PDF限制；第303行 | PDF、模板及arXiv兼容性没有证据，身份和正式研究清单待填。 | TeX环境具备后实际编译并处理诊断；作者和主控确认元数据、公开投稿授权及正式清单。当前不能宣称arXiv-ready。 |

## 实际文稿检查与资源

研究树准确候选通过 `git archive --format=zip ... cf867ce6e9bcf156fd3e036f0ac48a4834113b6b` 归档并由 PowerShell Expand-Archive 解到 Reviewer 自身 `.verification/candidate/`。全部生成文件只写该副本。候选正文、README、来源、主张表、中文反馈、构建/校验脚本和skill均已读取；研究树 HEAD 与候选一致，tracked内容相对候选无差异。

| 实际命令 | 结果与边界 |
|---|---|
| `./.verification/review.ps1 -Phase sources`、`-Phase originals` | EXIT0；18个原始HTTP响应与全文提取成功，MAAGL PDF13页、ASA PDF3页；不使用截图或视觉。 |
| `./.verification/review.ps1 -Phase dependencies` | EXIT0；Pypandoc_binary1.17/Pandoc3.9、Pybtex0.26.1及依赖仅装入自身 `.verification/python`。pip提示已有Paddlex的PyYAML版本约束；本次目标目录隔离，未升级系统包。 |
| `./.verification/review.ps1 -Phase archive` | EXIT0；准确候选Git归档与解压成功，没有写入Writer树。 |
| `./.verification/review.ps1 -Phase build` | EXIT0；实际调用候选 `papers/astrocyte/build.ps1 -Pandoc <自身pandoc.exe绝对路径>`，串行生成HTML、完整LaTeX和JSON AST，HTML/LaTeX均使用citeproc与`--fail-if-warnings`。 |
| `./.verification/review.ps1 -Phase verify` | EXIT0；实际运行候选verify.py：9引用=9BibTeX=9HTML参考文献，8配置行，52个MathML节点，10个本地链接，正文约4850空白分词，无所查私人路径模式。 |
| `Get-Command latexmk,pdflatex,xelatex,lualatex,tectonic -ErrorAction SilentlyContinue` | PATH未找到TeX引擎，PDF编译未执行。生成LaTeX不构成排版或投稿兼容性验证。 |
| `git check-ignore -v .verification/candidate/papers/astrocyte/build/manuscript.html .verification/sources/maagl-pdf.pdf` | 两者均被自身唯一`.verification/`条目忽略；没有把他人全文、依赖或构建产物提交。 |

Markdown与BibTeX是单一可编辑正文来源；LaTeX由同一完整正文和同一引文库生成，没有并存的手工维护短稿。独立检查正文和实际生成内容，引用已解析，方法/分析/状态/局限均存在；上述程序检查只验证文档结构，科学判定由本报告的逐项审查给出。

所有下载、安装、归档、构建与校验串行；每阶段启动前CPU/内存双采样，相隔3秒睡眠加查询时间。所记样本最高CPU22%、最低可用内存5.961GiB，全部符合CPU<70%、可用≥4GiB；候选构建脚本内另一次守卫也通过。PowerShell及其子进程BelowNormal，Node堆512MiB。这些是启动样本，未监测并宣称运行峰值。

没有启动产品服务、浏览器、本地模型、正式八配置实验或V4；没有发送产品原生输入、开放工具权限、修改业务/工作流/正式实验，没有子Agent、Aliyun、视觉、伪数据、Git回滚或force push。下载和构建进程均已正常退出，唯一重操作窗口可交回主控。身份、正式设计、来源差异、分析、真实产品链路及PDF限制继续按上表保留。

## 真实CR、hook与主控确认

首份报告提交 `68288c4f72b4ec05a8d45d7e340d0aea0432962b` 已普通push，随后从GitHub实际读取和本地一致的分支SHA。`git diff --cached --check`通过，提交仅包含本报告及自身`.gitignore`。

Reviewer于UTC `2026-10-19T01:48:32Z` 实际发布 [PR #3 review 5566157313](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5566157313)，GitHub读取回执为 `COMMENTED`、`commit_id=cf867ce6e9bcf156fd3e036f0ac48a4834113b6b`。当前GitHub账号是PR作者，因此使用COMMENT公开上述限定接受决定，没有原生APPROVE。

随后实际执行原产品树 `scripts/review-decision-hook.mjs`，传入表中完整Run/Task/Dispatch/终端、40位候选SHA、`accept`、本报告非空绝对路径、限定原因和已安装Windows Orca入口。Node24.16.0以BelowNormal和512MiB堆配置运行，EXIT0；UTC `01:48:46Z` 回执 `ok=true`，消息 `msg_3de99dd43db5` 已持久入队，原始回执保留在自身 `.verification/review-hook-receipt.log`。

主控UTC `01:49:31Z` 真实回复 `msg_7de6cb67c399`，thread_id指向该hook消息：确认已完整读取报告并通过GitHub API核对真实CR和准确源码，接受首轮完整稿件的审查结论，没有本轮返修要求，收回唯一重操作窗口，并明确指示同Task检查最终消息后发送一次成功worker_done。正式研究、统计程序、原生/S3、完整产品CI、PDF和投稿均继续未验收。该实际读取与回复确认了本次唤起，入队回执本身没有被当成确认。

本次Python、Pandoc和Node调用均已完成并退出，没有未结束的验证session，也没有启动自有后台服务。最终报告补充该真实回执后按同分支普通提交与push，不修改研究候选或重新发布决定；主控按准确Dispatch负责后续正式释放。
