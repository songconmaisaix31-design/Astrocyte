# 预印本首轮独立审查

## 目标与所有权

这是 `tasks/research-preprint-20261019.md` 的审查任务。科研Worker完成完整初稿、普通push和Draft PR后，主控通过Orca CLI派发一个独立Reviewer。研究候选完整SHA与PR编号由主控在派发时提供；未取得候选前不开始CR，不把本文件视为已完成审查。

Reviewer使用独立Worktree/Branch，只修改 `docs/reviews/research-preprint-20261019.md` 和自身已忽略 `.verification/`。允许将已推送、已指定的研究候选作为审查Git基线初始化；正文、引文、skill和研究反馈保持只读，返修交原科研Worker。读取主控AGENTS、HANDOFF、STATUS、研究任务/产品决定和两份Downloads原件；读取研究树已配置的 `.agents/skills/astrocyte-research/SKILL.md`，记录实际应用情况。禁止子Agent和Aliyun Agent。

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
