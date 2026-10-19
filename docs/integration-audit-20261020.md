# 集成版本、任务与独有成果（2026-10-20）

本页保存接管时的完整快照。当前交付候选已推进为 `works-relay-integration@32ab1eec808a5848aabad40f11133735a1214e55`，实际程序与Web同为S `ed543c12`；当前结果见 [交付任务](../tasks/delivery-candidate-20261020.md) 和 [STATUS](../STATUS.md)。以下版本、Task和服务字段按原核查时间阅读。

本次选定 **PR #2的 `5725c1840a760105518d273eaaa28d7367f369ed` 为优先收敛的产品候选**，继续使用原 `works-relay-integration` 作为产品集成目标。该候选已经汇集现有产品基线、新版界面、动态DAG、发送租约校验及发布检查改进，具备继续集中收尾的源码、Windows包和独立验证记录。完整产品、S3和正式发布仍按原验收判断。

本次计划见 [任务文件](../tasks/integration-convergence-20261020.md)。接管根HEAD为 `9e07d3c69f4f183b7c5bd73cf2568de13c7abc22`，本次只提交主控文档。时间为2026-10-20 Asia/Shanghai；完整fleet采样完成于UTC `2026-10-19T17:45:35.4051955Z`，Git核查确认各来源保持。

## 准确版本与用途

| 用途 | 分支/位置 | 准确版本与状态 |
| --- | --- | --- |
| 优先收敛的产品候选 | `songconmaisaix31-design/frontend-observatory-20261018`，[Draft PR #2](https://github.com/songconmaisaix31-design/Astrocyte/pull/2) | `5725c1840a760105518d273eaaa28d7367f369ed`，本地/远端/PR一致、工作区干净；开发检查源码为`22cde9efc540553c433463af5b9e955790d57a99`，后继仅README/本轨报告 |
| 已普通集成的产品基线及PR目标 | `songconmaisaix31-design/works-relay-integration` | `f17a491cdd6392829094c95a3e5efb6011362bc0`，本地/远端一致、工作区干净；是候选祖先，相差26提交；本次没有推进该分支 |
| 根目录主控和正式实验线 | `validation/v1-20261012` | 接管`9e07d3c69f4f183b7c5bd73cf2568de13c7abc22`；保留V4运行代码、协议、主控任务及决定；本次追加核查文档 |
| 默认远端分支 | `main` | `3a02ce38481d1885adee2e051a1b524d78376897`，仍为Initial commit；当前产品未合入main |
| 已独立验包来源 | 原产品轨Source15 | `15d74da3edcee26550a71120370fda9c86230e04`；ZIP、130项及人工DOM按此来源验收，Source22尚未重建包 |
| 实际运行应用 | `http://127.0.0.1:64505/` | 程序`c5b3b5715f2e93b82365720914be24c73642bd1e`，Web`172bb445597193d55f1563ee5d21ac2ec881798d`；PID45676，启动UTC`2026-10-19T12:09:30.7340265Z`，BelowNormal；运行源码早于候选 |
| 科研候选 | `songconmaisaix31-design/research-preprint-20261019`，[Draft PR #3](https://github.com/songconmaisaix31-design/Astrocyte/pull/3) | `fcb1e0a70541e15de1f52507a43edadb9801e368`；科学正文仍以`a664233a`为已接受方法稿，后续为TeX技术交付 |
| 最新独立审查记录 | `songconmaisaix31-design/research-preprint-review-20261019` | `13bba19f5fa4c7fc2980ebc840c70ea9765bd4b8`；产品与科研分别限定结论，审查成果保存在独立分支 |

PR #2当前GitHub `MERGEABLE/CLEAN`，head的`statusCheckRollup`为空。准确982源码的两个CI run [39155403418](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/39155403418)、[39155406045](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/39155406045)均为cancelled；历史原件记录Ubuntu全仓Go测试先已失败。没有当前完整CI通过证据，MERGEABLE只说明当前Git合并状态。

## 当前需要继续的目标

最新三Task属于Run `run_29c7e66b4136`。本次实际查询均为failed；相应最新Dispatch均released/exited，保留已交付阶段与未完成目标。原十轮01/02/03/04/06/08共6轮限定接受，05/07/09/10待验。

| 目标与原Task | 独有成果/已接受范围 | 当前阻塞 |
| --- | --- | --- |
| 产品，`task_51276f272ff3` / `ctx_013843bce872` | 原PR #2；新版界面、26品牌、人工动态DAG与按需加载、发送租约/用途集成；Source15独立130项/双视口DOM；Source22检查输出目录修复及独立三项真实HTTP/SQLite；原c5同库恢复 | 完整CI和默认浏览器、真实原生接续、带Worker停止/unknown对账/释放、两Agent闭环与实际采用未通过；预算单位、权限证据和文件/命令工具关闭的具体选择未答 |
| 科研，`task_f824c0088743` / `ctx_1fe65b3eaadd` | 原PR #3；27页/14引用方法及统计提案、已有PDF流程、证据与实现分级；固定TeX Live2025验签和入口错误传播修订 | 完整TeX安装/Build/Validate、新引擎PDF/来源ZIP及独立重建未完成；服务与安装并发例外、作者/机构/许可未答；正式研究数据缺失，10依赖09 |
| 独立审查，`task_2923223271be` / `ctx_a8916e612c35` | 原审查分支；PR #2包/入口/恢复的三次限定accept，PR #3准确fcb为insufficient_evidence，真实CR/hook/主控读取回复已保留 | 等待对应Worker合法可验结果；完整十轮、原生/S3、科研技术交付未验范围保留 |
| V4日运行，原`task_e538790026ad`已completed | 根实验线、既有automation与历史采集/评分记录独立保留；automation `f5e3ee7a-ddfe-4abf-b81c-b55a042ab27b`当前enabled，next为2026-10-21 00:20 Asia/Shanghai | 原外部根`C:/Users/DW/AppData/Local/astrocyte-validation/`、权威锁和正式SQLite本次仍不存在，无法核验锁/day0/路由/授权；实际人评及终局未完成，未启动补采或模型请求 |

当前普通服务终端`term_d303f609-6d36-42fd-b1ca-54f970beee1c`独立于Worker；终端清单将它标为orphaned，实际OS进程仍在原路径监听且HTTP正常，保留运行。另有三个连接的根目录Codex终端（原主控、本次核查、V4），终端连接状态与受监督Worker活性分别记录。

## 原系统中尚未结算的Task

全部相关Run实际task-list共184个Task：149 completed、21 failed、9 ready、3 blocked、2 dispatched。下列14项为Task状态保留项，全部相关Dispatch的workerState均已结算；`unverifiable`保留未知，不能当作进程退出或开发继续的证明。本次没有重试、释放或变更这些历史记录。

| Task | 实际状态 | 对应范围/保留入口 | 续行条件 |
| --- | --- | --- | --- |
| `task_1d3e15a66b2a` | ready | S5，`release-judgment-20261015` | Laya语言、Jev配置及真实判断矩阵；原尝试abandoned/unverifiable |
| `task_2ee44933df72` | blocked | 原四树集成，`works-relay-integration` | 已有集成推进至f17，后续集中接入PR #2；原尝试stopped/exited |
| `task_b0197ca7ebdf` | ready | 旧产品模块负责人，`module-product-20261014` | 使用已汇集的当前产品候选和原报告，恢复前核原权威；原尝试abandoned/unverifiable |
| `task_f082d8b36471` | blocked | Relay R，`works-relay-runtime` | 受限原生、真实身份/完成/unknown证据；原尝试failed/exited，release_unknown |
| `task_3004b636a3ab` | ready | Relay H，`works-relay-http` | 已有HTTP源码已在产品历史；原生正向和组合范围待验；原尝试abandoned/unverifiable |
| `task_bd209566be37` | blocked | S4，`works-repository-context-s4-20261015` | 已有仓库上下文源码在产品历史，完整独立范围待验；原尝试failed/exited，release_unknown |
| `task_7b4e93c1a3c1` | ready | 后端规范，`standards-backend` | 已有成果在产品历史，原尝试abandoned/unverifiable，另有旧reclaimable |
| `task_7cbe30017760` | dispatched | 前端规范，`standards-frontend` | 记录中的替换尝试均已failed或abandoned；当前没有live Dispatch，Task标签待原owner核实 |
| `task_5f0a79b6a17a` | ready | Attention真实检查，`product-attention-tests-20261014` | 代码历史已覆盖；原尝试failed/exited，reclaimable，完整准确组合检查待验 |
| `task_6846cdb90c46` | ready | S2，`product-video-20261014` | 已有S2源码已覆盖；真实批准、执行及采用分别验收；原尝试abandoned/unverifiable |
| `task_da9aebf03f83` | ready | S3，`product-swarm-20261015` | 当前S3目标由最新产品轨继续；原尝试abandoned/unverifiable，原生/预算/采用前置保持 |
| `task_d0cee5a583bb` | dispatched | 旧S2/S3 Web，`product-ui-20261015` | 已有源码在产品历史；替换尝试均已failed或abandoned，当前没有live Dispatch |
| `task_3c33a99685aa` | ready | L1 Today，`works-relay-frontend` | 原生正向未验；原尝试abandoned/unverifiable，报告另有未提交修改 |
| `task_63080055d561` | ready | 旧M6集成，`attention-engine-final-integration` | 现有源码已由后续历史覆盖，正式V4仍遵守真实锁；原尝试abandoned/unverifiable |

全局fleet实际读取15页，以resource与projection中的workspace识别357条本项目Dispatch：160 succeeded、44 failed、24 stopped、129 abandoned，运行时无live记录；132 exited、225 unverifiable。终端状态字段为114 released、232 retained、9 release_unknown、2 reclaimable，均无active；其中36条没有materialized resource，只保留context/projection，不能将其retained字段当作真实保留终端。以上历史数量不代表正在运行的Agent数量，也不自动授权清理。

## 分支保留的独有内容

74个本地分支中，56个分支HEAD已由产品候选的Git历史覆盖；其余18个分支保留下表内容。比较同时使用可达关系与`git log --cherry-pick --right-only --no-merges`，分叉提交数只用于定位来源；最终是否缺少功能按具体文件及当前实现核实。

| 分支及HEAD | 独有成果与集成处置 |
| --- | --- |
| `validation/v1-20261012@9e07d3c6` | 正式实验/协议、日运行记录、主控决定和任务；保持独立，产品候选接续采用这些当前决定 |
| `standards-python@91e3e59d` | fail-fast整改、条件import检查及pygit2工作树保护；其中`a1089f94`等源码未进入候选，涉及冻结实验代码和依赖，需按原任务/冻结范围单独评审 |
| `songconmaisaix31-design/product-swarm-20261015@51b720f4` | 原prepared-input校验和报告；后续DDD整改已将读取移到app端口、候选继续扩展租约与观察，旧实现/旧测试不直接覆盖当前实现，报告保留 |
| `swarm-send-lease-20261018@75cadf95`，PR #1 | 发送租约/fence/用途校验已由`4c4fa1ea`按16个业务/测试路径移入PR #2并获限定接受；当前14文件内容与PR #1一致，另外两文件包含后续DAG/观察扩展。原Commit和独立报告保留；PR #1仍Draft，避免重复引入代码 |
| `songconmaisaix31-design/release-native-verification-20261014@b74b97a4` | `tests/release-package-native/`原HTTP/SQL客户端、受限archive校验与基础Windows验包工具；候选未包含这四文件及对应说明，真实原生正向仍缺许可/证据。后续仅在原生验收需要时按原owner评审接入 |
| `songconmaisaix31-design/cognitive-stack-planning-20261014@00220725` | L1当前规则及原生Session事件/生命周期范围的两份规划更新，作为接续要求保留 |
| `songconmaisaix31-design/module-product-20261014@d64a52c4` | 产品模块、双源和超限真实验收记录，未进入候选的提交为模块文档 |
| `songconmaisaix31-design/module-relay-20261014@f602bce0` | Relay、原生/HTTP和经验交接边界及未验范围的模块文档 |
| `songconmaisaix31-design/module-release-20261014@1a0c5ed3` | 原发布协调、实际41ce包和准确集成交接文档 |
| `songconmaisaix31-design/release-research-20261014@61cad05f` | CodeGraph适配器入口核查报告 |
| `songconmaisaix31-design/reverse-reference-20261014@153b8534` | 原始参考、CSGHub与记忆研究索引、带来源商业计划及独立附件；没有据此替代产品实现或正式研究结果 |
| `songconmaisaix31-design/research-preprint-20261019@fcb1e0a7` | 完整论文、Bib/来源/证据表、研究反馈、TeX构建入口；保持PR #3独立交付 |
| `songconmaisaix31-design/dynamic-dag-review-20261019@7f781dee` | 独立DAG审查报告；保留准确源码、实际加载比较、事件SQL与未验范围 |
| `songconmaisaix31-design/research-preprint-review-20261019@13bba19f` | 科研/十轮/产品附加独立审查、原件索引和实际CR/hook记录 |
| `workbench/v0.1@72639435` | 早期独立工作台owner与范围文档，已被后续工作台/Relay方向接续 |
| `songconmaisaix31-design/workbench-v01-backend@3b123fcd` | 标明superseded的后端WIP源码及未验记录，保持原件 |
| `songconmaisaix31-design/workbench-v01-frontend@7335ce54` | 标明superseded的前端WIP源码及未验记录，保持原件 |
| `workbench/v0.2-access@8f94d502` | 原接入协调及L1实际服务范围报告；当前运行产品和后续集成已采用独立开发历史，旧协调分支保留来源 |

当前远端共有70分支，均有同名本地分支且HEAD相同。4个仅本地分支为`ddd-release-20261018`、`ddd-validation-20261018`、`s2-s3-product-ui-implementation`、`songconmaisaix31-design/charybdis`；它们的HEAD均已在候选历史中。存在9个未配置upstream的本地分支，其中5个实际有同名同SHA远端，不能将缺upstream判成未push。

接管时62个工作树中，3个存在工作区变化；本次主控文档另按上方写域提交：

| 工作树 | 实际变化 | 处置 |
| --- | --- | --- |
| `works-relay-frontend` | tracked `docs/access/frontend.md`未提交，包含L1 Today实际服务/事件读取记录，当前分支HEAD`65ba685e` | 保留独有文稿，后续交原owner核实提交；本次未替他人暂存或覆盖 |
| `s1-acceptance-1009` | 五个未跟踪日志/浏览器输出入口 | 保留验证原件，不计为未提交业务实现 |
| `validation-v1-ingest` | 两个未跟踪`__pycache__/`目录 | 保留原目录，不做清理 |

## 优先收敛顺序与停止边界

1. **PR #2优先。** 保持原树、分支和Worker负责，将当前源码作为唯一产品收尾来源。先核原CI失败和Source22已修三项，区分剩余旧替身/来源/迁移/独立race前置；仅针对真实问题提出具体返修。准确Source15的130项验包和人工DOM已有独立accept，按原条件复用记录。需要新源码验收或重建包时，用同来源真实数据及独立Reviewer核准确Commit。
2. **同一产品集成目标。** PR #2当前可无Git冲突进入原f17目标，但完整CI、业务验收和原批准边界继续逐项核实。当前实际服务保持c5；切换程序、合并PR和正式发布在各自授权内另行执行。本次只选定来源并记录顺序。
3. **科研按已有PR继续。** 静态来源和文稿工作可接续；完整安装/编译需要满足资源窗口及服务并发决定。完整TeX交付与作者/许可和研究实证分别验收。正式V4原锁/库恢复和人评沿原独立任务处理。
4. **独有成果保留。** 原生验包工具、Python冻结线整改、独立报告和未提交L1文稿逐项按需要接入。14个旧Task和历史资源记录由原权威核实后处理，当前记录不足以将它们认作并行开发人员；本次未新增Agent、Run、Task或Dispatch。

后续完整默认CI、原生AT09/AT13、S3和两Agent成果实际采用为仍未通过项；S3预算单位、权限证据覆盖、受限Codex、具体经验准入/采用、首发范围、TeX与服务共存、作者/机构/许可继续待答。用户允许费用与下载的授权保持，未答问题仍不代选。当前保留64505单服务，未启动编译、测试、TeX安装、浏览器或模型输入。

## 本次实际核查

| 命令/检查 | 结果 |
| --- | --- |
| `git status --short --branch`、`git worktree list --porcelain`，每树`git status --short`及HEAD | 完成62树清点；接管根和当前产品/科研/审查树干净，3树变化如上 |
| `git for-each-ref`、`git rev-list --left-right --count`、`git log --cherry-pick --right-only --no-merges`及具体文件diff | 完成74分支可达/独有提交核查；f17到5725为0/26，18分支独有历史按内容分类 |
| `git ls-remote --heads origin`、`gh pr list`、`gh pr view 1/2`、`gh run list --branch …` | 70个远端HEAD与同名本地一致；PR #1/#2/#3均Draft；PR #2目标/HEAD及Git合并状态已核，完整CI未通过 |
| `orca worktree list`、`orca orchestration worker-list --include-remote --limit 100`逐页读取、相关Run的`task-list`、`orca terminal list` | 15页全局fleet；本项目357 Dispatch（含36条无materialized resource的context记录）、184 Task、14未结Task、4终端；最新两个产品相关Run的active均0 |
| `orca automations list --json`与权威路径`Test-Path` | V4计划仍enabled；原根/锁/SQLite均false，未运行实验 |
| `Get-NetTCPConnection`、`Get-Process`、GET health及首页 | 唯一64505监听仍为原PID45676/路径/创建时间/BelowNormal，health为ok、首页200/562bytes |
| `git diff --check`、PowerShell标准Parser及本次10个本地文档链接目标检查 | 均PASS；仅静态/文档检查 |

完整只读核查入口与JSON位于根忽略目录`.verification/integration-audit-20261020.ps1`及`.verification/integration-audit-20261020/`。本次核查没有重新验证产品功能、SQLite逐表计数、正式实验或论文构建；这些范围沿准确原报告保留。本次未合并PR、切换服务、创建tag/Release、投稿或清理历史资源。
