# 预印本首轮独立审查

## 目标与所有权

这是 `tasks/research-preprint-20261019.md` 的审查任务。科研Worker完成完整初稿、普通push和Draft PR后，主控通过Orca CLI派发一个独立Reviewer。研究候选完整SHA与PR编号由主控在派发时提供；未取得候选前不开始CR，不把本文件视为已完成审查。

Reviewer使用独立Worktree/Branch，按仓库默认Git基线建立，只修改 `docs/reviews/research-preprint-20261019.md`、自身 `.gitignore` 中的 `.verification/` 条目和自身已忽略 `.verification/`。研究稿从科研树准确Commit只读；需要构建时可用成熟Git归档/复制工具把准确候选放进自身忽略目录，在该目录生成真实构建输出，不能改写科研Worker树。正文、引文、skill和研究反馈保持只读，返修交原科研Worker。读取主控AGENTS、HANDOFF、STATUS、研究任务/产品决定和两份Downloads原件；读取研究树已配置的 `.agents/skills/astrocyte-research/SKILL.md`，记录实际应用情况。禁止子Agent和Aliyun Agent。

## 必须核查的内容

1. 完整正文是否明确属于提出机制与计划实验的稿件，摘要、方法、讨论和结论是否一致；作者、机构、ORCID、正式注册、样本量、参数和结果的未知项是否诚实保留。
2. 真实读取原始来源，核查标题、作者、日期、版本和对应正文支持。逐项检查文献记录，特别是MAAGL作者索引差异和ASA原链接新闻稿；引用中的未核信息不作为已知事实。核查正文/BibTeX条目是否一致，下载全文与缓存是否留在忽略目录。
3. 经验更新×关系更新×决策位置的八配置、共同初始状态、固定模型、共用评分与信息/候选/权限/并发控制，是否忠实于两份原稿。同步回放的实现一致性与实际部署的位置效应是否分别定义，是否允许零或负结果。
4. 完整八配置配对序列的重复单位、任务集与初始化层次、主要交互估计、三阶交互、位置差异、盲评隔离、预算耗尽的分母、快照成本、区间/实用阈值/等效性/多重比较及样本量计划，是否存在泄漏、伪重复或因果混淆。
5. 主张—证据表及中文反馈是否准确区分拟议、已实现、工程验证与正式结果；研究策略与产品 `7ce3a5efe312f4cdac3491293e002f5559f076b0` 的 `swarm.fixed.v1`、经验人工批准、DAG展示和原生权限限制是否一致。正式V4及工程场景不能被改称新研究结果。
6. 实际执行适用的文稿构建与引用检查，记录准确命令、环境、结果和限制；Markdown、LaTeX及BibTeX如果共存，需要核查内容一致性。禁止伪造编译或把人工清单称为程序检查。论文技能通过真实稿件和本次CR验证，不另建设自我证明测试或视觉报告。

## 决策、hook与资源

报告按严重程度说明具体位置、证据、影响、可复现检查和必要返修。使用 `accept`、`revise` 或 `insufficient_evidence`，判定仅覆盖首轮完整稿件，不表示产品、正式实验或arXiv投稿已通过。把正式研究前置条件与本次稿件错误分开，不能降低原有验收。

在准确候选SHA上发布真实GitHub CR；同一GitHub账户无法提交原生APPROVE时，以COMMENTED公开具体审查决定并如实说明。随后调用已有真实hook：`C:/Users/DW/orca/workspaces/Astrocyte/frontend-observatory-20261018/scripts/review-decision-hook.mjs`，传入实际Run/Task/Dispatch/终端、40位候选SHA、决定、非空报告绝对路径和原因；Windows Orca入口为 `C:/Users/DW/AppData/Local/Programs/orca/resources/bin/orca.exe`。主控真实读取、回复后才算唤起已确认。

只执行本次文稿所需的检查，不启动产品服务、本地模型、正式V4或八配置实验，不修改模型和权限配置。重操作必须取得主控串行窗口、通过CPU<70%/可用内存≥4GiB双采样，BelowNormal运行，Node上限512MiB。禁止视觉、伪造数据、Git回滚和force push。报告提交、普通push后保持原Task处理复查；主控明确结束后再正式完成与释放。

## 实际候选与派发

候选为 `cf867ce6e9bcf156fd3e036f0ac48a4834113b6b`，Draft PR [#3](https://github.com/songconmaisaix31-design/Astrocyte/pull/3)，研究Branch `songconmaisaix31-design/research-preprint-20261019`，目标main，派发前主控实际核对远端HEAD。Run `run_d59e41c6d2f1`；Reviewer Task `task_a0c647db8962`，Dispatch `ctx_8bc7f5be8bbe`，终端 `term_a494186b-4a9c-467d-a366-0f0184ef4e80`。

Reviewer Worktree为 `C:/Users/DW/orca/workspaces/Astrocyte/research-preprint-review-20261019`，Branch `songconmaisaix31-design/research-preprint-review-20261019`，独立默认基线。启动回执turn_start_unobserved后，主控读取实际终端，确认任务仍在Codex输入框，使用原生Enter提交既有内容，没有重发任务或重建Agent；真实Codex会话随后产生assistant与工具记录，provider为gpt-6.1-sol、activity为working。科研Worker已交还唯一重操作窗口并冻结候选等待CR。

## 实际结论与结算

决定accept，仅覆盖首轮完整方法/协议稿、来源、主张证据分级及中文反馈，没有本轮返修缺陷。独立报告分支最终Commit为 `cab4233829b5caa1d66f8b05ce458d8596ef5289`，仅新增报告和必要忽略条目；首份报告68288c4f之后的提交只补实际CR/hook/主控回复，主控已读取全部差异。准确研究候选始终为cf867ce6。

Reviewer实际发布 [CR 5566157313](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5566157313)，COMMENTED、commit_id准确；独立18原始响应、准确Git归档副本、实际Pandoc构建和文稿引用/八配置检查通过。启动样本最高CPU22%、最低可用5.961GiB，串行BelowNormal；这些是启动采样，未声称运行峰值。没有TeX引擎，未编译PDF，正式统计/实验与产品链路不在接受范围。

实际hook消息 `msg_3de99dd43db5` 已由主控完整读取，主控核GitHub API后真实回复 `msg_7de6cb67c399`。Reviewer按明确结算指令发送成功worker_done `msg_9b6807907c29`，主控接受后、ack前对准确Dispatch官方release，回执released / closed_agent_terminal / transcript captured。审查Task已完成，报告与Worktree保留，自有验证调用均退出，没有后台服务或本地模型。
