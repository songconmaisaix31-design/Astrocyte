# 第02轮实际验证

2026-10-19。同一科研Worker，Codex CLI 0.162.0 / Orca `codex / gpt-6.1-sol`，终端显示GPT-6.1-Sol xhigh fast，继承既有模型及effort配置。准确研究基线为 `cf867ce6e9bcf156fd3e036f0ac48a4834113b6b`，既有Draft PR #3。初稿的引用与方法/协议审查已经接受；本轮负责实际PDF与完整来源，独立审查仍待主控安排。

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

## 真实剩余限制

一处Underfull段落间距警告保留，文本与页面边界检查通过。没有执行视觉检查，未验收视觉排版或arXiv站点处理。完整来源尚未在arXiv受支持的TeX Live处理器上重编译；本地Tectonic成功不能代称本站兼容性通过。

这份PDF仍为方法/协议稿。完整统计程序、条件/时钟/归因建议、盲评/成本/样本规划和最新S3证据由后续明确轮次处理；当前不存在正式八配置、成员扰动或自主迭代结果。工程记录不能证明这些效果，完整产品CI也没有因文稿构建而取得通过。

作者、机构、ORCID、正式模型/任务/阈值/预算/重复次数及注册仍未确认；未填造身份或结果。原创贡献与真实研究证据尚未达到本轮可宣布投稿就绪的状态，论文许可、作者授权、AI使用披露及实际公开提交仍待主控与真人作者处理。没有修改产品、用户原件、正式V4、根治理或workflow，没有运行产品服务、本地模型、正式实验或子Agent。

本Worker交付准确科研候选并普通push原Draft PR #3，随后冻结等待独立Reviewer的真实CR、hook与主控接受。未代替Reviewer发布CR或执行其hook；本Task覆盖后续获准轮次及返修，本轮不结算worker_done。
