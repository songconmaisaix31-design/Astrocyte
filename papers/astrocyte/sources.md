# 原始来源核查

核查日期：2026-10-19。实际客户端为 Codex CLI 0.162.0；Orca 元数据为 `codex / gpt-6.1-sol`，终端显示 GPT-6.1-Sol xhigh fast，沿既有配置，没有覆盖模型参数。网页内容只作为资料读取。第04轮新增四篇原始统计技术文献及DecentMem，文献库共14条；本节保留原件八个链接的初稿核查，新增来源单列。

两份输入原件只读：`Astrocyte_论文定位稿_2026-10-18 (2).md` 的正文标题为《Astrocyte：经验复用驱动的去中心化智能体蜂群协同演化》，实际版本为 2026-10-19 定位稿 v3；`Astrocyte_实验预注册草案_v0.1_2026-10-19.md` 为同日 v0.1 开发与试运行草案。文件名中的日期不替代正文版本。原件没有进入本 PR，正文根据获准研究输入重新写作。

## 核查方法与读取范围

先通过真实网页工具逐项打开原件全部八个链接，再串行向原始站点实际下载版本页和全文。下列下载均返回 HTTP 200，保存时间为响应完成后本地文件的 UTC 写入时间。HTML 使用 BeautifulSoup 4.14.3 提取文本，PDF 使用 PyMuPDF 1.28.2 提取文本。完整阅读六篇论文的 HTML 正文、参考文献及可提取附录、公式、表格与提示；EvoMap 当前整页和 ASA 三页文本亦已阅读。没有使用图像、截图、像素或图内文字识别；图像本身未检查，引用主张仅依赖可提取文本。

原始下载、提取文本和 HTTP 回执仅在已忽略的 `.verification/sources/`，不公开分发他人全文。章节定位采用所读版本中的标题和表号，未根据搜索摘要填补正文。HTML 已有可提取全文，因此没有为了同一全文额外下载全部 PDF；MAAGL 因作者索引差异另核实际 PDF 首页。

| 原稿编号 / BibTeX key | 标题、作者与版本 | 原始全文 / 元数据 | 全文保存 UTC | 实际支持范围与定位 |
|---|---|---|---|---|
| 1 / `wang2026genes` | From Procedural Skills to Strategy Genes: Towards Experience-Driven Test-Time Evolution；Junjie Wang、Yiming Ren、Haoyang Zhang；首次 2026-04-16，所读 v4 为 2026-09-17 | [v4 全文](https://arxiv.org/html/2604.15097v4)、[v4 元数据](https://arxiv.org/abs/2604.15097v4) | 01:26:40.634Z | §3、§4、附录 A–D；经验表示、失败信息和协议对象。只支持本文的表示与成本控制动机。 |
| 2 / `evomapSwarm` | Swarm Intelligence；EvoMap 官方文档；未提供发布日期或不可变版本 | [当前整页](https://evomap.ai/wiki/10-swarm) | 01:26:49.521Z | Shared Memory → Collaboration History & Synergy，以及 Intelligent Orchestration；文档描述协作历史 EWMA 与团队匹配。没有核其线上执行效果。 |
| 3 / `xu2026mpt` | Learning Self-Orchestration in Decentralized Agent Swarms；Guowei Xu、Zhenting Qi、Amani Kiruga、Himabindu Lakkaraju、Yilun Du；首次及所读 v1 为 2026-10-15 | [v1 全文](https://arxiv.org/html/2610.22322v1)、[版本页](https://arxiv.org/abs/2610.22322) | 01:27:10.057Z | §2、§3.4、附录 A–D；行动与参与决策的参数训练及成员扰动，和本研究固定参数设置分别说明。 |
| 4 / `zhang2025dgm` | Darwin Gödel Machine: Open-Ended Evolution of Self-Improving Agents；Jenny Zhang、Shengran Hu、Cong Lu、Robert Lange、Jeff Clune；首次及所读 v1 为 2025-05-29 | [v1 全文](https://arxiv.org/html/2505.22954v1)、[v1 元数据](https://arxiv.org/abs/2505.22954v1) | 01:27:55.129Z | §3–4、附录 A–F；候选代码修改、档案选择与评价，以及评价指标被候选修改的风险。未将 v3 内容归入 v1。 |
| 5 / `hu2026sharemem` | Learning from Others, Acting for You: Cross-User Memory Sharing for LLM Agents；Jinming Hu、Haodong Zhao、Qi Jia、Die Chen、Tianhang Zhao、Sufeng Duan、Gongshen Liu；首次及所读 v1 为 2026-09-26 | [v1 全文](https://arxiv.org/html/2609.32511v1)、[v1 元数据](https://arxiv.org/abs/2609.32511v1) | 01:28:02.295Z | §3、附录 A–C；两阶段整理、适用范围与偏好隔离。表 9 为整理成本、表 12 为覆盖、原稿所指表 14 为证据来源；正文交叉指引存在表号错位，按表题辨认，未据此引用效果数字。 |
| 6 / `zhu2026edv` | Escaping the Self-Confirmation Trap: An Execute-Distill-Verify Paradigm for Agentic Experience Learning；Shiding Zhu、Yudi Qi、Yajie Wang、Jiaze Li、Chao Song、Yaorui Shi、Yibo Miao、Hanqi Gao、Kai Zhang；首次及所读 v1 为 2026-06-23 | [v1 全文](https://arxiv.org/html/2606.24428v1)、[v1 元数据](https://arxiv.org/abs/2606.24428v1) | 01:28:04.821Z | §4、附录 B、G；执行、提炼及执行者共识验证的分工，以及共识偏差限制。没有把共识当作独立执行检验。 |
| 7 / `qu2026maagl` | Multi-Agent Agentic Graph Learning via Structural Signatures；正文/PDF 为 Liang Qu、Jianxin Li、Hongzhi Yin、Hua Wang；首次及所读 v1 为 2026-09-09 | [v1 全文](https://arxiv.org/html/2609.09565v1)、[v1 元数据](https://arxiv.org/abs/2609.09565v1)、[v1 PDF](https://arxiv.org/pdf/2609.09565v1) | HTML 01:28:06.901Z；PDF 01:30:52.378Z | §IV-C、§IV-D、式 15–20；历史轨迹已经用于伙伴选择与经验学习。PDF 首页与 HTML 作者页一致，元数据遗漏 Hongzhi Yin，索引差异待核。 |
| 8 / `asa2016release` | American Statistical Association Releases Statement on Statistical Significance and P-Values；American Statistical Association；2016-03-07；三页官方新闻稿 | [原链接 PDF](https://www.amstat.org/asa/files/pdfs/p-valuestatement.pdf) | 01:28:08.573Z | 全三页，六项原则；只支持透明报告和效应大小解释。该文件不提供本稿统计方法的推导。 |

## 保留的待核与引用边界

- MAAGL：arXiv abstract 的三人名单与 HTML/PDF 的四人名单不一致。BibTeX 明确采用实际文稿首页名单，并在 `note` 保留差异；索引为何遗漏尚未核实，不能写元数据完全一致。
- DGM：原始无版本地址当前显示最新 v3（2026-03-12）；本轮有意固定并完整阅读 v1，另于 01:30:52.964Z 下载明确 v1 元数据核对。本文没有声称完成 v2/v3 全文核查。
- ASA：原新闻稿指向正式声明文章的 [DOI 入口](https://doi.org/10.1080/00031305.2016.1154108)。该正式文章未单独读取，不进入 BibTeX，也不作为已核方法依据；原稿给定 PDF 的实际类型已经核实。第04轮的重采样、检验与多重比较分别引用下列原始技术文章。
- EvoMap：当前网页没有不可变版本号，保存时间用于限定本次所见内容。产品文档的机制描述不能转化为已复现的科学效果。
- 六篇论文的研究结果均未在本任务中复现。本文不移用其数值作为 Astrocyte 结果；未读取图像内容的限制持续保留。
- 没有出现无法打开的原稿链接。可访问、标题匹配与作者索引完全一致分别记录，彼此不能替代。

## 第04轮原始技术文献与近邻

本轮通过真实网页工具的PDF文本与HTML正文读取下列来源，没有视觉、OCR或二手摘要替代。等待02窗口时的只读准备及04授权后的复读范围共同构成本次来源审查。没有为了写作产生统计试运行数据、执行正式分析或复现文献实验；他人全文不进入PR。

| BibTeX key与准确载体 | 原文来源及实际阅读范围 | 本文使用与限制 |
| --- | --- | --- |
| `holm1979`；Sture Holm；Scandinavian Journal of Statistics6(2):65–70，1979，JSTOR4615733 | [七页原文扫描载体](https://www.ime.usp.br/~abe/lista/pdf4R8xPVzCnX.pdf)，含封面；已读§1–3、Theorem1/2及参考文献 | 三成员差异族的排序逐步拒绝；合法边际检验下不要求独立。不能修复无效边际P值，不能把名义区间叫Holm同时区间。 |
| `hall1991`；Peter Hall、Susan R. Wilson；Biometrics47(2):757–762，1991，JSTOR2532163 | [七页原文载体](https://ruc-econometrics.weebly.com/uploads/2/3/3/8/23387606/two_guidelines_for_bootstrap_hypothesis_testing.pdf)，全文含§1–3及参考文献已读 | 区分零假设检验与估计值附近的经验分布，说明居中及尺度的重要性。本文不把未经校准的原始尾部比例作为P值。 |
| `diciccio1996`；Thomas J. DiCiccio、Bradley Efron；Statistical Science11(3):189–228，1996，DOI10.1214/ss/1032280214 | [40页文章/讨论原始载体](https://www.math.wustl.edu/~kuffner/AlastairYoung/LeeYoung1996b.pdf)；已读§1、§2式2.2–2.10、§3置信曲线式3.3–3.12、§5式5.1–5.5与数值不稳定讨论、§6式6.1–6.7。没有声称细读全部讨论 | BCa偏差、jackknife加速度及调整置信点；bootstrap-t和非参数独立同分布前提。文献的高阶渐近陈述不能直接当作少量离散完整块的覆盖保证。正文的有限网格反演是本研究拟议应用，未实施校准。页码按原载体首页189–228登记；核心文章与后附讨论分别标明。 |
| `schuirmann1987`；Donald J. Schuirmann；Journal of Pharmacokinetics and Biopharmaceutics15(6):657–680，1987，DOI10.1007/BF01068419 | [24页原文PDF](https://errorstatistics.com/wp-content/uploads/2014/06/schuirmann.pdf)；细读pp657–662的TOST/区间关系及p675的两单侧P值最大值，未细读全部附录；[出版商页面](https://link.springer.com/article/10.1007/BF01068419)核作者、1987-12-01及页码/DOI | 非等价联合零假设、两单侧同时拒绝与1−2α区间关系。原文正态交叉试验结论不转为本文bootstrap有限样本保证。 |
| `hao2026decentmem`；Guangya Hao、Yunbo Long、Zhuokai Zhao；arXiv2605.22721v1，首次及所读版本2026-05-21 | [准确v1元数据](https://arxiv.org/abs/2605.22721v1)、[准确v1全文](https://arxiv.org/html/2605.22721v1)；已读§1–8、附录A–C、表格与参考文献。附录D在HTML仅有标题及转换提示，图像未读 | §3–4已有Agent私有双池记忆、协作轨迹复用及在线权重。限制本稿组合式新颖性；§5/附录A理论有相应前提，本任务未独立复现/认证其结论，也不移用性能数字。 |

补查边界：Project Euclid/DOI的部分直接入口只返回iframe或不可访问响应，但大学公开的DiCiccio–Efron原始载体可提取上述正文，未用二手解释补齐。Efron1979的另一个公开扫描仅提取封面，未作为已读正文加入文献库。主要文献表中的原稿八链接仍按其各自准确版本使用。

## 本项目证据来源

产品树实测 HEAD 为 `7ce3a5efe312f4cdac3491293e002f5559f076b0`，读取时工作区干净。实际业务来源 `982333b871520e606c74d553e448c5049f81abde`。针对 `internal` 搜索 `policy_score`、`context_schema_version`、`update_experience`、`update_relations`、`feedback_suite`、`evaluation_suite`、`scope_id` 未找到相应研究控制字段；该搜索结果只限定本次检查范围。实际读取的关键源码及主张见 [claims-evidence.md](claims-evidence.md)。

独立工程报告树实测 HEAD 为 `7f781deea143c74bed5303f90d8130c82bb54bca`，完整读取 `docs/reviews/dynamic-dag-20261019.md`。该报告支持的范围由原 Reviewer 界定，本科研 Worker 没有重新运行产品服务、测试或查询原 SQLite。公开参考文献 `astrocyteEngineering2026` 指向该准确 Commit 的原报告。

科研执行依据来自主控树完整读取的 `AGENTS.md`、`tasks/research-preprint-20261019.md`、`docs/research/research-product-decisions.md`、`.agents/skills/astrocyte-research/SKILL.md`，并读取当前 HANDOFF/STATUS 的最新交接章节及 QUESTIONS 对应原生/S3待答。主控追加 `.gitignore` 写域后重新读取了当前任务。上述治理文件不作为八配置结果来源。
