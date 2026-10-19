# 科研 Agent 与 arXiv 预印本首轮

## 目标、基线与范围

用户提供《Astrocyte_论文定位稿_2026-10-18 (2).md》和《Astrocyte_实验预注册草案_v0.1_2026-10-19.md》，要求启动一个科研 Agent、配置科研 skill、开始写 arXiv 预印本，并由主控统筹科研反馈和产品需求。两份原件位于 Downloads，只读；前者实际内容为2026-10-19定位稿v3，后者为开发/试运行准备草案，未正式注册、未生成正式结果。

管理基线 `validation/v1-20261012` / `bcf4e0e3b4cfd6ac4d5d629edd3eada6e3d41ae5`；产品候选 `7ce3a5efe312f4cdac3491293e002f5559f076b0`，业务982333b8；产品集成仍f17a491c。上一轮只接受人工DAG与前端构件；原生生成/Worker退出、OpenViking正向和完整CI未通过。正式V4原外部证据根不可用，实验与automation不属于本任务写域。

本轮建议：围绕经验更新×关系更新的交互收益，写明决策位置的次要研究问题，忠实保留八配置设计；建立拟议机制、当前实现和可检验证据的对应关系，完成可审阅预印本初稿和中文科研反馈。先按英文正文/中文反馈准备，语言与署名已异步询问，作者/机构/ORCID未答保持待填；不阻塞来源核查和方法草稿，不提交 arXiv 或正式注册。

## 所有权与交付

只启动一个长期科研 Worker，Orca CLI 独立 Worktree/Branch，项目专用 skill 为 `.agents/skills/astrocyte-research/SKILL.md`。启动后显式读取主控树该 skill，并在自身树用文件编辑工具配置同名项目 skill，报告实际读取与使用。独立研究分支从仓库默认基线建立，通过主控给定的准确产品树/Commit和交接了解最新状态，不擅自合并产品分支。

Worker write_paths：`papers/astrocyte/**`、`docs/research/astrocyte-research-feedback.md`、`.agents/skills/astrocyte-research/**` 及自身已忽略 `.verification/`。默认基线仅LICENSE，追加允许本研究树 `.gitignore` 仅登记 `.verification/` 与 `papers/astrocyte/build/`，避免把下载全文和构建缓存提交。可读取全仓和用户两份原件；不能修改产品代码、根AGENTS/HANDOFF/STATUS、正式实验、workflow或其他任务。主控维护本计划、状态、用户决定及 `docs/research/research-product-decisions.md`；科研反馈不能直接触发产品修改。

首轮产物：完整英文草稿（Markdown为可审阅正文，提供适用的LaTeX/BibTeX来源或明确构建限制）、原始文献核查记录、主张—证据表、实现缺口/投稿条件、中文研究反馈及文稿构建说明。用户两份原件按实际内容保留来源和版本，不改写原件；所有未知实验结果、身份、正式阈值、样本量和模型清单明确标记。

Worker完成后提交清晰Commit、普通push并创建Draft PR。后续安排一个独立Reviewer，只读论文与证据、写评审报告，负责准确候选的CR与hook；返修仍交科研Worker。主控审查研究反馈对产品可靠性、权限、资源及研究有效性的影响，决定接受/延后/需要用户决定；新产品需求必须形成具体任务后交开发Worker。

## 验收、资源与停止条件

引用需由真实原始文献正文支持，标题/作者/版本/日期与内容逐项核验；打不开或不匹配的文献保持待核。工程检查不得作为八配置或因果收益结果。主要交互估计、位置差异、预算和独立重复定义与用户草案一致；方法可批评，但待定选择不能自行冻结。feedback/evaluation隔离、泄漏、混淆与伪重复需有明确审查。

科研Agent实际使用该skill，输出完整初稿和反馈，并对来源、主张与构建执行适用检查；只提供提纲不算首轮交付。PDF构建用成熟工具，按实际工具可用性报告；允许必要依赖，避免安装无关组件。首轮不执行八配置正式实验、不造结果、不修改产品或正式V4、不发布/注册。初稿PR和独立CR完成后保存后续研究/产品决定及真实阻塞。

沿已有资源守卫：重操作全局串行，CPU<70%、可用内存≥4GiB的双采样、BelowNormal，Node512MiB、Go单并发/256MiB软限制。初轮主要为阅读与写作，不启动本地模型、GPU或产品服务。禁止Aliyun Agent；首次只用一个科研Worker，不为skill自我证明另开A/B测试或浏览器评价。使用真实初稿与独立来源审查验证skill的适用性。

## 实际派发

Run `run_d59e41c6d2f1` / 主控 `term_d7578194-fd20-42b4-9668-b0b0e8683571`；科研Task `task_2b1b744c0b74`。独立Worktree `C:/Users/DW/orca/workspaces/Astrocyte/research-preprint-20261019`，Branch `songconmaisaix31-design/research-preprint-20261019`，默认Git基线 `3a02ce38481d1885adee2e051a1b524d78376897`；最新产品仍按原产品树7ce3读取，未合并业务。

首次Dispatch `ctx_71821af30379` 启动在Codex account/read workspace routing discovery超时后明确退出至PowerShell，主控只读终端核实后官方stop/release，无业务编辑。原Task/树重试为 `ctx_d72d923f0117`，终端 `term_afc3f5b3-3549-41da-b0cd-8b2aaae7aa52`；真实回执ready/turn_started，无模型参数覆盖。以科研Worker首个实际Handoff继续核实际客户端/模型、skill使用和证据范围。

科研Worker首个真实Handoff `msg_84d9a392c059` 已确认完整读取两份原件、主控AGENTS/任务/决定和科研skill，并用文件编辑工具配置自身同名项目skill，开始真实文献核查及完整初稿。实际Codex0.162.0，Orca provider元数据为gpt-6.1-sol，继承默认模型，无Aliyun或子Agent。主控已回复科研重点和产品边界；本研究树缺少既有.gitignore的必要配置写域已明确追加。

科研Worker提交普通push、Draft PR [#3](https://github.com/songconmaisaix31-design/Astrocyte/pull/3)，候选 `cf867ce6e9bcf156fd3e036f0ac48a4834113b6b`，PR目标main。真实Handoff `msg_9505de237e6e` 交还重操作窗口，保持原Task等独立CR和返修。主控已实际读取完整稿件、来源、主张表、中文反馈及构建脚本，核对PR仍Draft、准确远端HEAD；未将工程报告当作正式研究结果。

独立审查按 `tasks/research-preprint-review-20261019.md`，Reviewer只写自身 `docs/reviews/research-preprint-20261019.md`、必要 `.gitignore` 的 `.verification/` 条目及忽略输出；论文只读，构建只在自身准确候选副本执行。Reviewer取得唯一重操作窗口前仍须资源双采样。原科研Worker不进行并行检查或修改，待准确CR后承接返修。

Reviewer Task `task_a0c647db8962` / Dispatch `ctx_8bc7f5be8bbe`，独立树 `C:/Users/DW/orca/workspaces/Astrocyte/research-preprint-review-20261019`，实际Codex0.162.0 / gpt-6.1-sol默认配置。主控已通过真实会话记录确认其读取审查依据；候选、审查写域、资源窗口和hook规则见审查任务。最终独立CR接受首轮稿件，范围与未验事项如下。

## 首轮验收与后续入口

科研分支最终为 `cf867ce6e9bcf156fd3e036f0ac48a4834113b6b`。Reviewer独立下载18个真实原始响应，检查准确候选副本，Pandoc3.9生成HTML、完整LaTeX和AST通过，`verify.py`的9引用/9条目、8配置、52MathML及10本地链接检查通过。首份报告为68288c4f，最终补充实际通知记录的报告为 `cab4233829b5caa1d66f8b05ce458d8596ef5289`，分支 `songconmaisaix31-design/research-preprint-review-20261019`，普通push与干净状态已核。

真实 [CR 5566157313](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5566157313) 为COMMENTED、准确指向cf867ce6，决定accept，仅覆盖完整方法/协议初稿、来源、证据分级和科研反馈，没有本轮返修要求。同一GitHub账号不能原生APPROVE，未把COMMENTED改称平台批准。Reviewer实际hook `msg_3de99dd43db5` 已到达；主控完整读报告并经GitHub API核commit_id后，实际回复 `msg_7de6cb67c399`，确认唤起和接受范围。

科研Task正式成功worker_done `msg_91bda7a8d6ab`、Reviewer Task正式成功worker_done `msg_9b6807907c29` 均已核当前准确Dispatch。主控在确认后、ack前分别官方释放 `ctx_d72d923f0117` 和 `ctx_8bc7f5be8bbe`，回执均为released / closed_agent_terminal / transcript captured。自有下载/构建/验证调用已退出，未启动产品服务、浏览器或本地模型，没有关闭其他用户终端。

首轮科研写作及独立审查已完成；PR保持Draft，没有合并、部署、注册或投稿。PDF/模板/arXiv兼容性未验证，作者/机构/ORCID待填，MAAGL索引差异原因、统计技术依据及正式分析细节未齐；正式模型、任务、预算、阈值和重复次数未冻结。共用策略、配置隔离、条件与归因、盲评和完整计量需要产品实现及真实验收；原生/S3、完整CI、实际采用复用与正式V4证据根限制保持。

后续科研反馈复用现有研究树、分支及 `.agents/skills/astrocyte-research/SKILL.md`，由主控明确下一轮Task和写域后通过Orca CLI安排。主控按 `docs/research/research-product-decisions.md` 管理科研与产品先后，不让论文提案直接成为产品接口或执行授权。Reviewer提醒：`8rT`包含预算耗尽后的计划任务位置，真实尝试数、重试和实际消耗须分别记录；该要求纳入后续计量。
