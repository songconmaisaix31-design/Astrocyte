# 第02轮实际验证

2026-10-19。同一科研Worker，Codex CLI 0.162.0 / Orca `codex / gpt-6.1-sol`，终端显示GPT-6.1-Sol xhigh fast，继承既有模型及effort配置。第02轮准确研究基线为 `cf867ce6e9bcf156fd3e036f0ac48a4834113b6b`，既有Draft PR #3。02/04/06文稿与完整来源已获独立限定接受，逐轮记录保留；当前第08轮记录在本文件末尾。

主控任务为十轮计划的02。本轮重新读取两份Downloads原件、当前主控任务/决定与独立报告cab4233829b5caa1d66f8b05ce458d8596ef5289，实际应用本树astrocyte-research skill。正文及九条BibTeX保持cf基线内容，统计程序、具体机制和真实S3证据更新分别等待后续轮次。主控 `msg_c7d5680e477c` 已授予本轮唯一重操作窗口。

## 命令与实际结果

| 命令或检查 | 实际结果与范围 |
| --- | --- |
| `Get-Command xelatex,pdflatex,tectonic,latexmk -ErrorAction SilentlyContinue` 及常见TeX目录轻量检查 | 未找到可复用安装；没有进行全盘扫描。 |
| `& .\.verification\prepare-tex.ps1` | 官方GitHub release API固定Tectonic0.17.0，唯一Windows MSVC ZIP与API长度一致；下载/解压/实际`--version`均成功。工具仅置于本树忽略目录。 |
| `& .\.verification\run-research.ps1 -Phase build`，内部调用下列完整构建命令 | Pandoc3.9的HTML、完整LaTeX、JSON AST生成成功；citeproc和`--fail-if-warnings`通过。Tectonic0.17.0实际完成TeX及xdvipdfmx，生成完整13页PDF。 |
| 首次冷构建诊断及配置修复后的实际构建 | 首次控制台提示Fontconfig缺少默认配置；添加本项目标准Fontconfig配置并重编译后该诊断消失。TeX日志保留一处`Underfull \hbox (badness 2922)`，对应§8含源码路径的段落；没有提高警告阈值或隐藏诊断。无Overfull、Missing character、未定义引用或致命TeX错误。 |
| `& .\.verification\run-research.ps1 -Phase validate`，内部运行`python -X utf8 papers/astrocyte/verify.py` | 实际通过：9引用/9条BibTeX/9条HTML参考文献、八配置行、52个MathML节点；PDF13页、26个章节标题、9条参考文献标题，提取38566字符。9个字体均嵌入且无Type3；全部提取文字span位于页面内，无替换字符，PDF未加密且无需修复。 |
| 完整阅读`build/manuscript-pdf.txt`，对照Markdown、生成TeX及TeX诊断 | 已读677行提取文本，覆盖摘要至九条参考文献；折扣关系式、softmax、八配置表、成本式、三个估计量、Holm式及样本规划式存在。正文、无结果状态和待填署名完整。提取的数学上下标阅读顺序不等同视觉排版。 |
| `git diff --check` 与交付范围核查 | 无空白错误。只包含获准科研路径；依赖、缓存、他人全文及中间日志不进入公开PR。 |

实际完整构建命令：

```powershell
& .\papers\astrocyte\build.ps1 `
  -Pandoc (Resolve-Path .\.verification\python\pypandoc\files\pandoc.exe).Path `
  -Tectonic (Resolve-Path .\.verification\tectonic\tectonic.exe).Path
```

完整[LaTeX来源](preprint/manuscript.tex)由同一[Markdown正文](manuscript.md)和[BibTeX](references.bib)生成，参考文献已经内联，不依赖运行时BibTeX或独立删节稿。[PDF](preprint/manuscript.pdf)由该完整来源实际编译。构建日志、AST、HTML和完整PDF提取文本保留在忽略的`build/`；字体配置为[fontconfig.conf](fontconfig.conf)。

## 资源与实际失败记录

下载、构建和验证串行，进程BelowNormal、Node1024MiB。Tectonic缓存、Fontconfig缓存与临时文件均配置在本树`build/`或`.verification/`，没有系统安装、全局配置修改或常驻服务。资源数值为启动采样，不能证明运行峰值。

| 启动阶段 | CPU两次 | 可用内存两次GiB | 实际间隔 |
| --- | --- | --- | --- |
| Tectonic下载 | 70%、40% | 7.530、7.417 | 5.028秒 |
| 首次冷构建，内层实际构建检查 | 73%、78% | 7.260、7.149 | 5.025秒 |
| 首次冷构建，外层检查 | 77%、77% | 7.377、7.283 | 5.170秒 |
| 配置修复后的构建，内层 | 16%、14% | 6.632、6.602 | 2.524秒 |
| 配置修复后的构建，外层 | 9%、2% | 6.589、6.577 | 2.584秒 |
| PDF首次实际验证 | 14%、16% | 6.290、6.248 | 2.626秒 |
| 最终构建，内层 | 75%、83% | 6.443、6.409 | 4.066秒 |
| 最终构建，外层 | 55%、70% | 6.567、6.476 | 4.181秒 |
| 最终PDF及文档验证 | 41%、55% | 6.454、6.436 | 4.135秒 |

首次三条间隔叠加CIM采集耗时后略超2–5秒，已向主控报告并保留原始UTC回执；主控要求继续本次构建，不为该历史记录重新下载或冷构建。当前脚本等待2秒，使用Stopwatch单调计时检查实际2–5秒间隔，日志独立保存UTC时间。一次验证启动因字符串时间本地化造成错误间隔而提前终止，当次未调用Python，修复后实际验证通过。

## 第02轮验收边界

一处Underfull段落间距警告保留，文本与页面边界检查通过。没有执行视觉检查，未验收视觉排版或arXiv站点处理。完整来源尚未在arXiv受支持的TeX Live处理器上重编译；本地Tectonic成功不能代称本站兼容性通过。

这份PDF仍为方法/协议稿。完整统计程序、条件/时钟/归因建议、盲评/成本/样本规划和最新S3证据由后续明确轮次处理；当前不存在正式八配置、成员扰动或自主迭代结果。工程记录不能证明这些效果，完整产品CI也没有因文稿构建而取得通过。

作者、机构、ORCID、正式模型/任务/阈值/预算/重复次数及注册仍未确认；未填造身份或结果。原创贡献与真实研究证据尚未达到本轮可宣布投稿就绪的状态，论文许可、作者授权、AI使用披露及实际公开提交仍待主控与真人作者处理。没有修改产品、用户原件、正式V4、根治理或workflow，没有运行产品服务、本地模型、正式实验或子Agent。

第02轮准确候选 `618c82f270b4f35282dce86ef341b568a9690979` 普通push原Draft PR #3。[独立报告e192829f](https://github.com/songconmaisaix31-design/Astrocyte/blob/e192829f0d937e70d9f72de8d2dac27c88762ca1/docs/reviews/s3-paper-ten-cycles-20261019.md)及[真实CR5566782506](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5566782506)限定接受完整PDF与同源构建；主控已收到实际hook并回复，02计入十轮。未代替Reviewer发布CR或执行其hook；本Task覆盖后续获准轮次及返修，没有提前结算。

# 第04轮分析定义与原始依据

准确基线为已接受的 `618c82f270b4f35282dce86ef341b568a9690979`。主控 `msg_abefe8c756fd` 授权静态写作，`msg_4962c4dec2b0` 在Reviewer03实际退出后授予本轮唯一构建窗口。本轮使用已有成熟工具与自身缓存，没有额外安装。完整来源与PDF现为同一0.2版正文实际生成的交付。

具体修改为：完整八配置块向量与等权对比、整块共同抽样的BCa端点/退化检查、同一曲线有限网格反演的边际P上界、原三比较Holm族、TOST及明确标为建议的三比较等价族/次要预算均分、条件推断与真实pilot前提。四篇原始统计文献和DecentMem已按实际范围实读并加入BibTeX/来源表；正文收紧原创性表述，披露真实Codex工具使用。引用一致性检查改为读取实际文献集合，保留完整匹配与空集合拒绝。

没有真实pilot或正式结果，本轮没有执行数值分析、生成测试数据或声称区间覆盖/功效通过。B/J、网格、有效尾部秩、正式r及模型/任务/预算/阈值继续未确认；三成员等价族和误差预算为研究建议。正文的网格反演是同一置信曲线的保守计算定义，仍依赖BCa采样校准，不把它或Holm写成无条件有限样本保证。

## 第04轮实际命令与内容核查

| 命令或检查 | 实际结果 |
| --- | --- |
| `& .\papers\astrocyte\build.ps1 -Pandoc (Resolve-Path .\.verification\python\pypandoc\files\pandoc.exe).Path -Tectonic (Resolve-Path .\.verification\tectonic\tectonic.exe).Path` | EXIT0；Pandoc3.9引用解析与fail-if-warnings通过，Tectonic0.17.0完整编译17页PDF；本轮直接使用构建入口的单次资源守卫。 |
| 首次`& .\.verification\run-research.ps1 -Phase validate` | EXIT1；原始源码正则在索引27450匹配数学集合定义中的`x:\`，误认为Windows绝对路径，尚未进入PDF检查。原始失败文本保留在忽略目录，未修改数学正文规避。 |
| 文件编辑工具修复privacy检查，随后只重跑`-Phase validate` | EXIT0；用已有Pandoc AST检查正文、代码、原始块和链接目标，并检查实际PDF提取文本。驱动器模式增加标识符边界，防止将URL协议尾部误当驱动器；保留Unix路径和Bearer模式。没有删除隐私检查或把任意匹配视为通过。 |
| 本轮成熟文稿/PDF验证输出 | 14引用=14条BibTeX=14条HTML参考文献；八配置、125个MathML节点、19本地链接。PDF17页、30标题、14参考文献标题、49930字符；9字体全部嵌入、无Type3/替换字符/加密/修复，全部文字span在页内。 |
| 完整提取文本、来源和TeX诊断静态阅读 | 已读876行，覆盖摘要至14条参考文献。新增完整块向量、BCa偏差/加速度/端点、两侧P上界、Holm、TOST与原样本规划式完整；AI披露和无结果/待填作者状态保留。来源阅读范围与BibTeX对应，未把未细读的讨论或附录D写为已读。 |

实际生成TeX共1274行、63239bytes，PDF125646bytes。TeX日志保留两处Underfull段落间距警告：源694–702行、§6.2完整块向量段落的badness1708；源1076–1086行、§8源码路径段落的badness2922。两次TeX处理重复报告相同两处。无Overfull、Missing character、未定义引用或致命TeX错误；没有提高诊断阈值或隐藏警告。没有视觉、OCR或站点TeX Live检查。

| 实际阶段 | UTC启动时间 | CPU两次 | 可用GiB两次 | UTC记录间隔 |
| --- | --- | --- | --- | --- |
| 完整构建 | 04:40:22.818 / 04:40:25.346 | 4%、7% | 6.135、6.112 | 2.528秒 |
| 首次验证 | 04:41:04.067 / 04:41:06.605 | 10%、16% | 6.220、6.233 | 2.538秒 |
| 修复后的验证 | 04:42:14.063 / 04:42:17.417 | 39%、53% | 6.157、6.141 | 3.354秒 |

全部重操作串行、BelowNormal、Node1024MiB，实际间隔由Stopwatch守卫，UTC另存；启动数值不说明运行峰值。构建/首次错误/最终验证回执及完整日志保留在本树忽略目录。首次验证发现的是真实当前稿数学表达式误报，没有生成替身数据或自我证明测试。所有Pandoc/Tectonic/Python调用退出后，`msg_0359f670ec90`立即归还窗口，主控`msg_55d4adc4dca3`确认收回；其后只做完整文本阅读和静态报告/提交。

## 第04轮剩余与交接

本轮完成的是可审阅的分析程序提案、真实原始来源和更新后的完整来源/PDF。BCa与网格反演尚未用真实pilot校准，原生执行/正式研究参数未获确认，没有八配置或其他效果结果。条件/时间/归因完整机制在下述06轮继续；盲评/全成本/样本规划及最新S3证据等待后续明确轮次。arXiv实质贡献、受支持TeX Live/站点处理、真实作者及其科学核验/许可/批准继续未完成。

实际AI使用已在正文披露，没有虚称未知作者完成人工核查。产品、用户原件、根治理、workflow和正式V4未修改；无产品服务、本地模型、正式实验、视觉或子Agent。准确候选 `afe471edc25305537f11ee4453a93c7b3be790cf` 已普通push原PR #3；[独立报告3b6883dd](https://github.com/songconmaisaix31-design/Astrocyte/blob/3b6883dd32f91d471403239f0878bd5ad883e9bb/docs/reviews/s3-paper-ten-cycles-20261019.md)及[真实CR5567098640](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5567098640)限定接受04，实际hook `msg_d8dbadf2a00a` 由主控读取并以 `msg_3b165bee72a4` 回复，04计入十轮。本Task保持后续轮次与返修，不提前worker_done。

# 第06轮条件、时钟与归因

准确基线为已接受的 `afe471edc25305537f11ee4453a93c7b3be790cf`。主控 `msg_56215e0da433` 授权06静态实施，`msg_d272e39c466f` 明确主控绝对任务路径；本Worker完整读取该文件06方案、当前决定及独立02/04报告相应章节。具体条件、时钟、义务与修订方案保持研究建议，未冻结产品接口或正式参数。

正文0.3补充六类决策前条件、逐项缺失与精确分组、关系/义务类与具体目标、每任务/键唯一结果、固定封闭目标、验证修订链、既定位置时钟、严格先前可见性、真实历史初始化及关闭通道的衰减/缓存规则。给出五项完整条件证明：规范化/重复不变性、关系分数敏感度、固定候选softmax总变差、同信息与统一随机数选择一致、关闭通道新历史不干涉。构建前独立WIP预读及主控明确方案要求，已补充M细化/softmax组合的同位置/时钟/λ开启投影前提，并保留原通用L1界；另把权威首次可用a与实际V/H集合、父链闭合、已知撤销/缺口、未知陈旧性和本次计龄t完整区分。WIP预读不代替准确候选CR。统计§6、十四条文献和八配置保留。

有限只读复核产品32672499与当前3e52cc13的完整policy文件，相同路径差异为空，固定排序与拟议研究policy之间的差距仍在。没有将该有限复读扩展为完整S3或任何研究运行结果。来源、证据表与中文反馈已同步，具体真实实现验收要求在反馈06专节。

主控 `msg_41cf65ed3bd9` 在产品调用全部退出并归还后，明确授予06唯一构建窗口。本轮使用既有工具与缓存，实际命令和结果如下。

## 第06轮实际构建与核查

| 命令或检查 | 实际结果 |
| --- | --- |
| `& .\papers\astrocyte\build.ps1 -Pandoc (Resolve-Path .\.verification\python\pypandoc\files\pandoc.exe).Path -Tectonic (Resolve-Path .\.verification\tectonic\tectonic.exe).Path` | EXIT0；Pandoc3.9完整HTML/JSON AST/LaTeX及引用解析通过，Tectonic0.17.0实际编译正文0.3，生成23页PDF。 |
| `& .\.verification\run-research.ps1 -Phase validate` | EXIT0；14引用=14条BibTeX=14条HTML参考文献、八配置、214个MathML、19本地链接。PDF23页、39标题、14参考文献标题、69863字符；9字体均嵌入且无Type3，全部提取文字span在页面内，无替换字符、加密或修复。 |
| 完整阅读`build/manuscript-pdf.txt`及相关TeX/诊断 | 已读全部1186行，覆盖摘要至14条参考文献；条件目录、关系单位、b/a/V/H定义、按本次t计龄与初始化、五项证明及限定、八配置、完整§6与无结果/待填作者/AI披露均存在。对照源文及生成TeX，未采用视觉或OCR。 |
| §6与准确afe基线逐段字符串比较；`git diff --check` | 分析章节完全相同，空白检查EXIT0。仅科研获准路径及忽略验证目录变化。Git提示工作区LF将转换为CRLF，未报告空白错误。 |

完整生成TeX共1681行、85509bytes，PDF155840bytes；交付TeX与本次build输出一致。实际TeX仅保留两处Underfull段落间距警告：源1098–1106行、§6.2完整块向量的badness1708；源1480–1490行、§8源码路径的badness2922。两次TeX处理重复报告同两处；无Overfull、Missing character、未定义引用或致命错误，没有调高阈值或隐藏诊断。公式提取顺序和文字边界不证明视觉排版质量。

| 阶段 | 两次UTC采样 | CPU两次 | 可用GiB两次 | UTC记录间隔 |
| --- | --- | --- | --- | --- |
| 完整构建 | 05:27:27.780 / 05:27:30.527 | 14%、35% | 5.243、5.142 | 2.746秒 |
| 成熟文稿/PDF验证 | 05:27:47.871 / 05:27:50.566 | 16%、23% | 5.245、5.219 | 2.695秒 |

进程BelowNormal、Node1024MiB、串行，启动CPU<85%/可用≥4GiB及Stopwatch实际2–5秒守卫通过；启动采样不代表峰值保证。本轮构建和验证均首次通过，未重跑或安装新依赖。资源与完整诊断实际副本保留在忽略目录的round06文件。所有自有Pandoc/Tectonic/Python退出后，`msg_ac1da31fa417`已立即归还唯一窗口；之后只有静态全文阅读、报告与普通提交。

## 第06轮范围与交接

本轮交付完整机制提案和明确前提的证明及真实同源PDF，没有运行产品、真实事件链、模型、pilot、正式分析或视觉检查。条件目录、时钟、义务封闭及视图方案未成为已批准产品接口；统计参数、正式模型/任务/预算/阈值/r及研究授权保持未决定。当前固定policy与完整研究实现仍有差距，研究效果没有数据。

受支持TeX Live/站点处理、原创实质贡献、真实研究证据、作者元数据及其科学核验/许可/批准继续未完成；Tectonic本地成功不代表arXiv就绪。盲评/全成本/样本规划在下述08轮继续，完整最新S3证据等待主控明确10方案。准确候选 `f27289c9dbc7820e1b3bcde8330d5aa4f79b4908` 已普通push原Draft PR #3。[独立报告c1ffa8e7](https://github.com/songconmaisaix31-design/Astrocyte/blob/c1ffa8e7a3f1fc657725f31cee780bdfe33934fb/docs/reviews/s3-paper-ten-cycles-20261019.md)及[真实CR5567321949](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5567321949)限定接受方法提案/证明与文稿，主控实际读取并以 `msg_9d3e3d5bcb39` 回复真实hook `msg_fa885052b983`，06计入十轮。本Task继续，没有提前结算。

# 第08轮封存权限、全成本与试运行规划

准确研究基线为已接受 `f27289c9dbc7820e1b3bcde8330d5aa4f79b4908`。主控 `msg_274a0ed955fd` 明确08三段方案及静态授权，本Worker完整读取主控绝对任务文件的新方案、当前QUESTIONS和研究产品决定，以及独立报告c1ffa8e7的全部06章节。原Task/科研写域/分支/PR #3与实际工具配置保持。

正文0.4补充执行/更新/独立评价/分析四角色权限矩阵，固定最终作品与评价盲化、全部主实验计划块完成前封存、确认回执及各类日志/索引/上下文/错误泄漏路径的真实验证要求。成本补充原task/operation来源、平台提交与实际尝试、累计订正/父子包含的非重复记账、原用量对账不重发、支付/资源/墙钟/重叠分列，以及真实P一次记录、计划8r分摊和逐次恢复/执行/评价费用。§6.6只补合法前置、实现/隔离/计量、预先规划的独立完整八配置pilot、真实方差/依赖/退化/精度/校准依据，再正式决定参数的顺序。

封存释放要求固定实际生成的作品及各计划位置有来源的终局状态；已观察无作品算法失败按预声明规则计失败，未知评价/损坏/不完整块保持原停止规则，不填零未知或补造作品。pilot结果前须版本化模型/提示/推理/工具/policy/任务分布/T/历史/预算及拟规划目标，正式冻结逐项核对应；实质改变须有旧方差适用依据，否则先预规划修订目标独立pilot，再定r；开发数据/费用不并入正式块。中文关闭投影统一说明冻结研究位置零的初始非空计数/权重且不衰减，当前权限即时检查。

以上均为完整研究建议；权限矩阵不证明隔离，成本公式没有原始费用数值或预算单位决定，pilot顺序没有生成数据或正式r。06的英文条件/时钟/实际视图/限定证明以及04的主要估计量/BCa/反演/Holm/TOST程序保持，十四条文献不变。主控 `msg_936bb5b26301` 在产品实际Verify/服务全部退出并归还后授予08唯一构建窗口，实际结果如下。

## 第08轮实际构建与核查

| 命令或检查 | 实际结果 |
| --- | --- |
| `& .\papers\astrocyte\build.ps1 -Pandoc (Resolve-Path .\.verification\python\pypandoc\files\pandoc.exe).Path -Tectonic (Resolve-Path .\.verification\tectonic\tectonic.exe).Path` | EXIT0；Pandoc3.9完整HTML/JSON AST/LaTeX及引用解析通过，Tectonic0.17.0实际编译正文0.4，生成27页PDF。 |
| `& .\.verification\run-research.ps1 -Phase validate` | EXIT0；14引用=14条BibTeX=14条HTML参考文献、八配置、233个MathML、19本地链接。PDF27页、44标题、14参考文献标题、84405字符；9字体均嵌入且无Type3，全部提取文字span在页面内，无替换字符、加密或修复。 |
| 完整阅读`build/manuscript-pdf.txt`，对照正文/完整TeX与诊断 | 已读全部1440行，从摘要至14条参考文献。四角色六类资料矩阵、实际作品与无作品失败、未知缺测、全部块封存、全成本包含/订正、P/(8r)分项式、pilot目标对应及既有方法/统计公式完整。 |
| 准确f272基线逐段比较及`git diff --check` | 06英文Setting/Proposed method和§6.1–6.5完全相同，空白检查EXIT0；Git保留LF→CRLF提示，无空白错误。 |

完整生成TeX1978行、101986bytes，PDF172215bytes；交付TeX与本次build输出一致。TeX日志仅两处原有Underfull：源1312–1320行、§6.2完整块向量badness1708；源1777–1787行、§8源码路径badness2922。两次TeX处理重复同两处，无Overfull、Missing character、未定义引用或致命错误；没有隐藏诊断或调整阈值。完整提取文本/页内边界不证明视觉排版质量。

| 阶段 | 两次UTC采样 | CPU两次 | 可用GiB两次 | UTC记录间隔 |
| --- | --- | --- | --- | --- |
| 完整构建 | 05:48:35.766 / 05:48:38.445 | 8%、11% | 5.723、5.689 | 2.679秒 |
| 成熟文稿/PDF验证 | 05:48:50.887 / 05:48:53.639 | 11%、14% | 5.881、5.890 | 2.752秒 |

BelowNormal、Node1024MiB、串行，CPU<85%/可用≥4GiB及Stopwatch实际2–5秒守卫通过，启动数值不代表峰值。本轮两调用首次均通过，没有新安装或重跑；完整日志/资源副本保留自身忽略目录round08文件。全部Pandoc/Tectonic/Python退出后，`msg_340c9dd3225a`已立即交回唯一窗口；之后仅静态完整文本阅读、文档和普通提交。

## 第08轮范围与交接

未运行产品、原生输入、模型、模拟、pilot、正式实验、视觉或TeX Live。原生/S3三项待答、正式预算单位、模型/任务/阈值/样本量等仍未确定。没有真实隔离或成本匹配/覆盖/功效/效果结论；历史P未知没有变成虚构数值，位置零初始非空计数与权重保持冻结语义。

原创实质贡献、真实研究数据、受支持TeX Live/站点兼容、作者元数据及人工科学核验/许可/批准仍未完成，没有注册或上传arXiv。最新完整S3证据待主控明确10方案。准确候选 `a664233a1276003a69f3d6850e4ecc04c08f4370` 已普通push原Draft PR #3；[独立报告c36f8990](https://github.com/songconmaisaix31-design/Astrocyte/blob/c36f899007bef3d1a90599fa884ceaa48fe28ceb/docs/reviews/s3-paper-ten-cycles-20261019.md)和[真实CR5567552099](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5567552099)限定接受08完整提案与同源交付，真实hook `msg_814e1b18d6a8` 经主控完整读取并以 `msg_7cac54d820d1` 回复，08计入十轮。本Task继续负责返修与后续获准工作。

# TeX Live 2025技术准备

主控 `msg_74297159b44d` 先授权只读官方手册与自身忽略目录具体方案；方案已实写 `.verification/texlive2025-plan-20261019.md` 并由 `msg_c54b361cf68d` 交接。08接受后，主控 `msg_c0d16ba47b6d` 明确授权build-texlive.ps1、verify.py显式PDF参数、README/validation及来源ZIP的技术实现，科学正文/十四条文献/原TeX-PDF保持冻结。该技术准备不计额外轮次，也不启动依赖09的10。

技术入口与自身忽略目录完整profile已实现。官方2025 final目录和同发行install-tl的POD手册按实际范围阅读，Windows wrapper实际完整HTTP文本已读。网页工具读取TUG HTML返回403；其后直接HTTP完整读取[官方签名验证页](https://www.tug.org/texlive/verify.html)，核实官方主密钥fingerprint。完整scheme-full/windows、doc/src、七树和临时/缓存/GnuPG限定自身忽略目录，portable关闭宿主PATH/菜单/文件关联。源码ZIP仅包含实际编译需要的完整根manuscript.tex；Build干净解包并经XeLaTeX两遍及诊断通过后生成来源交付，再由Validate执行完整文稿/链接/PDF检查。

启动采样逐次先写自身忽略目录的实际UTC、Stopwatch elapsed/interval及全部拒绝原因，再立即抛出不合门限的错误；文件名保留真实采样时间，后续显式调用不覆盖旧拒绝记录。没有改变85%/4GiB/2–5秒门限或增加重试。

| 静态命令 | 实际结果 |
| --- | --- |
| `& .\.verification\check-texlive-syntax.ps1`，内部仅Language.Parser.ParseFile | PowerShell7.6.6 Parser EXIT0，没有执行目标脚本正文。技术入口要求PowerShell7.2+并带UTF-8 BOM。 |
| `powershell.exe -NoProfile -File .verification/check-texlive-syntax.ps1` | EXIT1，当前签名策略拒绝未签名辅助脚本，解析器未运行。未修改执行策略或把此项记为通过。 |
| Python `ast.parse`读取verify.py源码 | EXIT0，未import或运行文稿验证正文。 |
| `git diff --check`及准确a664的科学正文/Bib/原TeX-PDF差异检查 | EXIT0；冻结制品未改变，原PDF所有检查仅新增明确输入路径，默认行为和断言保留。 |

## Download首次实际结果

主控 `msg_cfebfcce0e34` 授予仅Download窗口。实际ZIP36160153bytes、SHA512文件145bytes、签名488bytes、官方公钥5231bytes下载完成，GNU sha512sum返回`install-tl.zip: OK`。随后GPG公开密钥导入EXIT2，实际日志报告`can't connect to the gpg-agent: Invalid value passed to IPC`及`no gpg-agent running`；入口随即EXIT1，没有导出、签名验证、解压或安装。原GNUPGHOME、下载和signing-key-import.log全部保留。所有前台调用退出后，`msg_38a119c710ad`立即归还窗口，主控 `msg_853cd37e7ffe` 确认收回。

启动样本UTC06:29:34.109/06:29:36.686，CPU18%/5%，可用5.404514/5.684055GiB，Stopwatch间隔2.5773639秒，无拒绝；完整原资源JSON保留。此后C盘准确可用55,639,179,264bytes。30GiB磁盘/12GiB网络仍为规划预留，未安装完整发行或测量其总量。

## VerifyInstaller修订与实际结果

主控只授元数据和官方文档静态核查。实际gpg/gpgv均2.4.5-unknown，串行version/dump-options四命令全部EXIT0；gpg支持dearmor/assert-signer/no-autostart/no-default-keyring，gpgv不支持assert-signer。已实读[GnuPG操作文档](https://www.gnupg.org/documentation/manuals/gnupg/Operational-GPG-Commands.html)及[配置文档](https://www.gnupg.org/documentation/manuals/gnupg/GPG-Configuration-Options.html)相应完整条目，静态改为GPG dearmor生成专用公钥ring，再原生assert-signer验证官方fingerprint；不导入本地钥匙库、不启agent、不手解析OpenPGP或让EXIT2继续。

新VerifyInstaller只复用原四份下载，使用新的自身verify-installer目录及日志，完整SHA512/签名通过才解压；不重下ZIP或覆盖原失败GNUPGHOME。成熟ZipArchive还只读TLCrypto.pm和标准公钥ring条目，Install将从已验安装器内的标准tlpkg/gpg核仓库，全部路径仍自身忽略树。PowerShell Parser与空白检查通过。

主控 `msg_5c8deff9e39d` 明确批准且只授VerifyInstaller唯一窗口。实际命令 `& .\papers\astrocyte\build-texlive.ps1 -Phase VerifyInstaller` EXIT0：GNU SHA512复核OK，GPG dearmor EXIT0，签名与原生assert-signer EXIT0，随后Expand-Archive完成。新公钥ring为3803bytes；原生VALIDSIG记录签名子密钥 `D8F2F86057A857E42A88106A4CE1877E19438C70`，ASSERT_SIGNER通过官方主密钥 `C78B82D8C79512F79CC0D7C80D5E5D9106BAB6BC`。原日志保留TRUST_UNDEFINED及未获信任签名认证的标准警告，没有使用always-trust或省略fingerprint断言。此次没有import、启动agent、联网重下或覆盖首次失败证据。

实际ZIP含2610条目，条目展开长度合计119953438bytes；解压目录为自身 `installer/install-tl-20260301`，唯一Windows入口为 `install-tl-windows.bat`。此数量仅描述安装器ZIP，不表示完整TeX Live已安装。启动UTC06:37:25.661/06:37:28.200，CPU18%/6%，可用5.049244/4.992504GiB，Stopwatch间隔2.5384087秒，无拒绝；当时C盘可用55,658,147,840bytes。资源原JSON与三份核验日志完整保留。

全部调用退出后，`msg_26f5055d1935`立即归还窗口，主控 `msg_12d31f37859c`确认收回，只允许静态报告与profile/参数同步。当前没有Install/Build/Validate、新引擎PDF、模型、产品或正式实验。2025 final本地兼容、arXiv2025-08-03精确快照、本站处理、研究贡献与作者许可分别核验；签名通过和安装器解压没有构成投稿就绪。

## 安装前路径核查

主控 `msg_35aa77d76817` 明确要求只读核查Windows路径。宿主LongPathsEnabled实际为1，当前安装前缀98字符。已验发行Config元数据为Perl5.40.1；其[官方Windows源码](https://raw.githubusercontent.com/Perl/perl5/v5.40.1/win32/win32.c)相应stat/rename/mkdir函数仍含A接口或传统路径缓冲。同发行TLUtils.pm的tl_abs_path对不存在路径明确拒绝`//?/`。已读[Microsoft完整长路径说明](https://learn.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation)，注册表为1不足以证明全部应用链路支持。

主控 `msg_17dadf82afd2` 明确批准轻量官方元数据例外，BelowNormal/Node512串行读取，不占Reviewer的重操作窗口。实际下载同官方texlive.tlpdb.xz2704836bytes及checksum144bytes、签名488bytes；已验发行自带xz解压20189437bytes，GNU SHA512和原生GPG签名/官方主密钥assert-signer分别EXIT0。使用发行自带Perl和TeXLive::TLPDB/TLPOBJ原生from_file、expand_dependencies及replace_reloc_prefix读取实际scheme-full/windows依赖，全过程EXIT0，没有手写TLPDB parser。

全库8406包，适用依赖4978包，236376个唯一条目：run181845、doc46601、src6958、windows-bin972。98字符根下最长实际文件248字符，为ualberta包的 `texmf-dist/doc/latex/ualberta/03_References/Reference_PDFs/Aldrich2017-OSM_Classic_an_Optical_Imaging_Technique_for_Accurately_Determining_Strain.pdf`；文件长度≥260或父目录长度>248的条目为0。完整JSON和原始stderr保留自身path-metadata，Perl对继承C.UTF-8不受支持而使用系统默认locale的原生警告原样保留，没有人工重跑隐藏。此项不证明所有生成/临时路径或安装成功，未安装、合成路径测试、修改宿主、创建盘符、缩减scheme或更换原目录。
