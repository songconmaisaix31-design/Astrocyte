# 实际验证与交付边界

2026-10-19。执行者为同一科研Worker，Codex CLI 0.162.0 / Orca `gpt-6.1-sol`，未覆盖模型或effort参数。科研分支从仅含LICENSE的 `3a02ce38481d1885adee2e051a1b524d78376897` 开始；产品读取准确7ce3a5e，独立报告读取准确7f781dee，未合并产品源码。

## 已执行的验证

| 命令或检查 | 实际结果 | 适用范围 |
|---|---|---|
| 真实网页打开八个原链接；`.verification/run-research.ps1 -Phase sources` | HTTP原始来源均可访问；六篇固定版本全文、EvoMap整页、ASA三页已读取 | 不复现论文实验，不读取图像。逐项URL和时间见sources.md。 |
| `.verification/run-research.ps1 -Phase extra` | MAAGL实际PDF首页与HTML同为四人；DGM明确v1元数据可读 | MAAGL抽象页索引差异仍待核，DGM未混用v3正文。 |
| `git -C <产品树> rev-parse HEAD` 与 `git status --short` | 产品准确 `7ce3a5efe312f4cdac3491293e002f5559f076b0`，读取时干净 | 只读源码与报告，没有重跑产品测试。 |
| `python -m pip install --target .verification/python pypandoc_binary pybtex` | EXIT0；pypandoc_binary1.17、Pandoc3.9、Pybtex0.26.1及其依赖只安装到本树忽略目录 | pip显示本机Paddle对PyYAML版本的依赖提示；本次目标目录隔离，未升级系统安装。BeautifulSoup4.14.3和PyMuPDF1.28.2原已可用。 |
| `& .\papers\astrocyte\build.ps1 -Pandoc (Resolve-Path .\.verification\python\pypandoc\files\pandoc.exe).Path` | EXIT0；HTML、完整LaTeX来源、JSON AST生成通过；citeproc与`--fail-if-warnings`未报错 | 生成LaTeX不等于PDF编译，不包含版式或视觉验收。 |
| `.verification/run-research.ps1 -Phase validate`，内部运行`python -X utf8 papers/astrocyte/verify.py` | EXIT0；9引用对应9条BibTeX和9条HTML参考文献；8配置行正确；52个MathML节点；10个本地文档链接存在；正文约4850个空白分词；正文无本机绝对路径或凭据模式 | 检查真实文稿结构；没有实验数据或研究效果检验。 |
| `Get-Command latexmk,pdflatex,xelatex,lualatex,tectonic -ErrorAction SilentlyContinue` | PATH无上述工具 | PDF编译未执行，不能报告成功。 |
| `git check-ignore -v .verification/sources/gene-full.html papers/astrocyte/build/manuscript.html` | 两个目录分别被获准的.gitignore条目忽略 | 他人完整论文、依赖和生成文件不进入公开PR。 |
| `git diff --cached --check` | EXIT0，无空白错误；暂存11件获准文件 | 没有暂存下载全文、生成文件、依赖或产品改动。 |

全部重操作串行，实际下载、依赖安装、构建及最终验证启动前均执行间隔约3.6秒的CPU/内存双采样，阈值为CPU<70%、可用≥4GiB。各阶段最高启动CPU分别为来源8%、补查13%、依赖15%、构建24%、末次验证19%；对应最少可用内存分别为6.902、6.301、6.429、6.423、6.753GiB。进程使用BelowNormal，Node堆512MiB；这些是启动采样，未声明运行峰值上限。

构建脚本和验证脚本均通过文件编辑工具编写。下载/提取脚本只处理公开来源和本树忽略输出，没有修改用户原件、产品、正式V4或根状态。项目skill通过文件编辑工具配置并用于实际来源与主张审查。没有额外skill A/B试验或子Agent。

## 真实剩余限制

PDF编译和排版检查未执行；当前成果是可构建的Markdown、BibTeX及实际生成的HTML/LaTeX来源。arXiv兼容性与正式投稿未验收。作者、机构、ORCID待填，注册平台/编号为空。

MAAGL索引差异待核；ASA正式声明文章本轮未单独读取。已有技术文献支撑相关工作定位，正式统计程序仍需确认技术依据、零假设校准、区间与等价/多重比较细节。

八配置运行、模型实验、成员扰动和A/B/C迭代案例均未执行；正式预算、阈值、模型/任务清单与独立重复次数未冻结。原生工具关闭、S3两项待答、真实Worker生命周期、OpenViking正向及作品采用/复用边界保持。完整产品CI有实际失败和取消，未通过；本任务没有修改原断言或workflow。V4原证据根不可用的状态来自主控交接，本任务没有操作正式实验。

本Worker不发布独立CR或执行Reviewer hook。普通push和Draft PR提供准确科研候选，后续由主控安排独立Reviewer；最终接受范围以其CR为准，同一科研Worker负责返修。
