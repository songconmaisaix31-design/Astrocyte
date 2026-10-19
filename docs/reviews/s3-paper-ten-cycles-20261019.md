# S3 与论文十轮独立审查

2026-10-19。按主控逐轮冻结的准确候选审查；本报告从本次第 01 轮开始，上一轮研究稿的接受不计入本次。每轮只有具体改变、适用真实检查、准确 GitHub CR 和主控实际读取回复全部具备后，才能由主控登记接受轮数。

## 对象、职责与当前边界

| 项目 | 实际值 |
| --- | --- |
| 审查基线 | `cab4233829b5caa1d66f8b05ce458d8596ef5289` |
| 独立分支 | `songconmaisaix31-design/research-preprint-review-20261019` |
| 生命周期 | Run `run_29c7e66b4136`；Task `task_2923223271be`；Dispatch `ctx_568047699698`；终端 `term_447ff4b9-bd02-4409-aa43-4b86f1f74556` |
| 实际客户端 | Codex CLI 0.162.0；Orca 终端 `GPT-6.1-Sol xhigh fast`；继承配置，无模型覆盖 |
| 写域 | 本报告及自身已忽略 `.verification/`；两个 Worker 树保持只读，修复交原 Worker |
| 权限 | 预算计数、权限证据覆盖及受限 Codex 三项未答；现有 Claim 拒绝、各 `Known=false`、原生文件/命令工具关闭要求保持 |

实际读取主控 AGENTS、HANDOFF、STATUS、十轮任务、QUESTIONS、SPEC §7–8/10/12/17/19–20、TASKS P3 及研究产品决定；完整读取用户定位稿 v3 和预注册开发草案 v0.1。逐论文候选应用 `astrocyte-research` skill 的证据分级、原始来源、完整配对块、因果识别、隔离和成本审查要求。使用 orchestration/orca-cli 技能处理真实生命周期、消息、CR 通知与主控回复；第 02 轮另读取 PDF skill，按用户要求只作文本与结构核查，不实施视觉检查。

用户后续要求初步跑通产品并发布首版 Release，以及符合发布标准的 arXiv 预印本。后续相应候选仍需独立验收包与准确源码的一致性、实际可用性及官方投稿要求；当前接受范围没有覆盖这些后续结果。作者身份与三项执行选择继续待答。

## 第 01 轮：测试前置与迁移断言

决定：`accept`，接受真实 Git 测试前置和准确迁移身份/幂等断言修复。当前没有需要原 Worker 返修的本轮缺陷；接受范围仅限以下两项测试修复。

| 对象 | 准确来源 |
| --- | --- |
| PR | [产品 PR #2](https://github.com/songconmaisaix31-design/Astrocyte/pull/2)，实际读取时 open、Draft、HEAD 与候选一致 |
| 候选 | `dacc742e7ea9f0cc7ff011f0df8343caa976a8a6` |
| 测试源码提交 | `b21f5b66852a7fa8692185ebfaa34ddaae1f2e00`；至候选的后继差异仅标准记录 |
| before | `7ce3a5efe312f4cdac3491293e002f5559f076b0` |
| 原 CI 业务源码 | `982333b871520e606c74d553e448c5049f81abde` |
| 唯一窗口 | 主控消息 `msg_83e4b62f0f28` 明确授予第 01 轮，限定本次测试修改的适用检查 |

准确 before→after 差异只有 `cmd/server/discovery_config_test.go`、`internal/adapters/sqlite/taste_profile_test.go` 和当前标准记录。生产代码、workflow、原生配置、预算与授权逻辑均未变化。

### 原失败及具体改变

实际读取原 Reviewer 保存的两份原始 Ubuntu job 日志：push run `39155403418` / job `117757474469`，PR run `39155406045` / job `117757483047`。两次 run 最终 cancelled，取消前全仓 Go test 已经 failure。PR 日志第 14–15 行记录真实 detached HEAD 的 `Branch=nil`；两份日志分别第 128–130、130–132 行记录迁移数量 27 与旧断言 26 不符。其他失败仍存在，本轮没有将 cancelled 或局部通过改称完整 CI 通过。

| 文件 | 真实改变与静态判定 |
| --- | --- |
| [discovery_config_test.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/dacc742e7ea9f0cc7ff011f0df8343caa976a8a6/cmd/server/discovery_config_test.go) | 从真实当前公开仓库 clone，读取真实 HEAD，建立 `discovery-validation` 分支。保留原 Branch 非空、HEAD、提交时间、Dirty 以及无效配置、重复根和已删除根的拒绝断言；增加 HEAD/分支精确相等及 clean 检查。没有把 detached HEAD 伪报为 attached，也没有伪造 Git/CLI。初始化使用实际 Git、禁用 hooks、清除继承的 `GIT_*` 并设置 30 秒超时，失败直接终止。 |
| [taste_profile_test.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/dacc742e7ea9f0cc7ff011f0df8343caa976a8a6/internal/adapters/sqlite/taste_profile_test.go) | 固定当前真实 27 份迁移及 026/027 身份；重复运行后核对数量和全部迁移 ID。保留旧版本 SQL 升级、真实备份、两张品味表及零虚构 profile 断言。结果检查更准确，没有删除失败分支或放宽生产权限。 |

### 独立实际运行

依据主控明确补充的真实 Git clone 方式，在自身 `.verification/ten-cycles/01/{before,after}` 分别执行 `git clone --shared --no-checkout` 和首次 `git switch --detach <完整SHA>`。核对各自 HEAD、实际 origin、detached 和干净状态，结果保存在 `before-identity.json`、`after-identity.json`。两份源码均来自真实公开产品仓库；没有在无 `.git` 归档中向上误读审查树身份，没有写入产品 Worker 树。

执行入口为自身 `.verification/ten-cycles/review-cycle01.ps1`。以下命令均在对应准确克隆中运行，原始日志、资源样本、旧 CI 日志副本及真实 SQLite 保存在 `.verification/ten-cycles/01/`。

| 命令与条件 | 实际结果 |
| --- | --- |
| before：`go test -mod=readonly ./cmd/server ./internal/adapters/sqlite -run '^(TestExplicitProjectDiscoveryConfiguration\|TestTasteProfileActualMigrationUpgrade)$' -count=1 -timeout=2m -v` | EXIT1，两项均真实失败。发现测试 `Branch=nil`；迁移测试返回 27。函数分别 2.14 秒、0.62 秒。原件 `before.log`。 |
| after：相同测试命令，准确候选、detached 调用条件 | EXIT0，两项均 PASS。发现函数 11.10 秒，真实内层 clone 的 HEAD 精确为候选、分支为 `discovery-validation`，clean 和全部拒绝断言通过。迁移函数 0.76 秒，实际备份、27 条迁移 ID、重复执行、两表及零 profile 全部通过。原件 `after.log`。 |
| `go vet -mod=readonly ./cmd/server ./internal/adapters/sqlite` | EXIT0，原件 `vet.log`。 |
| `gofmt -d cmd/server/discovery_config_test.go internal/adapters/sqlite/taste_profile_test.go` | EXIT0，无输出差异。 |
| `git diff --check` 及准确 before→after 的 `git diff --check` | EXIT0，原件 `format.log` 及实际命令回执。 |

检查改善了两个原本不符合实际前置/迁移数量的测试：在真实 detached 调用条件下，它们现在验证所声明的 attached Git 观察和当前迁移完整性。没有删除原断言、使用 mock、虚假数据或仅为通过而伪造原生能力。未改变产品行为，未据这些测试推断科研效果。

启动时 CPU/内存双采样、BelowNormal 和自身输出目录均实际生效。snapshot 与 before 在用户新资源决定前使用 Go1/`-p=1`/256MiB、Node512MiB；已启动 before 保留原额度。主控消息 `msg_955e8b1a5580` 后，after/vet/format 使用 CPU<85%、可用≥4GiB、Go4/`-p=4`/512MiB、Node1024MiB。实际 Go 阶段样本间隔约 4 秒；所有记录样本 CPU 44%–67%，最低可用内存 6.637GiB。快照采样间隔 5.035 秒，包含当时 3 秒等待与 CIM 开销；后续采用 2 秒等待，保留原始记录，没有重跑快照补证明。记录是启动样本，内存设置为软限制，未监测并宣称峰值。前后资源额度和编译缓存状态不同，因此不进行速度或资源收益比较。

本轮仅运行上述两项测试和定向检查。Worker 另报的 HTTP/SQLite 回归没有在本轮独立重复，不扩张本次接受范围。全部自有验证调用已退出，未启动后台服务、浏览器、本地模型、产品原生输入或正式实验；当前没有待关闭的自有后台进程。

### 当前缺口与真实 CR

完整默认 CI 仍未通过；全仓 build 未运行，本轮未补跑此前被资源守卫拒绝的 Worker build。原生正向、AT08/09/13、真实两 Agent 工作、具体用户采用、经验再次使用及完整 S3 均未获本轮验收。PR #2 尚待第 03 轮处理其与已审 PR #1 的发送前租约/用途检查差距；本轮没有自动合并其他来源。Release 包、论文 PDF、arXiv 标准及正式八配置研究也未由本轮通过。

Reviewer 于 UTC `2026-10-19T03:11:13Z` 发布 [真实 CR 5566519259](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5566519259)。GitHub 原始回执为 `COMMENTED`、`commit_id=dacc742e7ea9f0cc7ff011f0df8343caa976a8a6`，正文明确第 01 轮、接受范围和全部关键未验项。同一账号使用 COMMENT，未伪称 GitHub APPROVE；原件为 `github-review.json` 和 `github-review-receipt.json`。

报告 `0bd44e135a0b415b431433a6cbd2075e35a4d131` 已普通 push、远端一致。实际执行既有 hook，UTC `03:14:22Z` 入队 `msg_9137a34ddfd6`，原件 `01/review-hook-receipt.log`。主控 UTC `03:15:18Z` 回复 `msg_0bd7db71f620`，thread 指向该 hook，确认完整读取报告、核对准确 GitHub CR 和远端，并登记第 01 轮限定接受、累计 1/10。主控已收回该窗口，真实读取回复确认本次唤起；本 Task 继续后续审查。

## 第 02 轮：完整 PDF 与单一正文构建

决定：`accept`，接受完整 PDF 交付和 Markdown/BibTeX→完整 TeX/PDF 的实际构建流程。当前没有需要原 Worker 返修的本轮构建或交付缺陷；投稿质量、视觉排版、站点兼容及研究效果分别保持未验。

| 对象 | 准确来源 |
| --- | --- |
| PR / 候选 | [研究 PR #3](https://github.com/songconmaisaix31-design/Astrocyte/pull/3)；`618c82f270b4f35282dce86ef341b568a9690979`，真实 API 核 Draft、open、HEAD 一致 |
| before | `cf867ce6e9bcf156fd3e036f0ac48a4834113b6b`；已有完整 Markdown、BibTeX 和 LaTeX 生成入口，没有 PDF 交付 |
| 当前交付 | [完整 TeX](https://github.com/songconmaisaix31-design/Astrocyte/blob/618c82f270b4f35282dce86ef341b568a9690979/papers/astrocyte/preprint/manuscript.tex) 1018 行及 [PDF](https://github.com/songconmaisaix31-design/Astrocyte/blob/618c82f270b4f35282dce86ef341b568a9690979/papers/astrocyte/preprint/manuscript.pdf)；正文与 9 条 BibTeX 相对 before 无差异 |
| 窗口与重试 | `msg_4a38206e392f` 授予唯一窗口；第一次 CPU87% 拒绝后，`msg_8cf1331d8f00` 明确允许至少两分钟后的单次 snapshot 重试 |

准确差异为 9 个获准科研文件：构建/校验、Fontconfig、完整 TeX/PDF、README、validation、主张记录及 skill 资源额度。构建使用成熟 Pandoc citeproc 与 Tectonic，不维护独立删节稿；Pandoc 警告和进程失败直接终止，TeX 诊断保留。Fontconfig/cache/temp 均由进程变量指向本次副本，没有修改全局配置或两个 Worker 树。

### 独立构建与已提交文件核对

执行自身 `.verification/ten-cycles/review-cycle02.ps1 -Phase snapshot|build|verify`，准确 Git archive 及全部输出在 `.verification/ten-cycles/02/`。构建前将归档中的已提交 TeX/PDF 保留至 `submitted/`；可执行文件、成熟 Python 依赖和已下载 TeX 缓存按主控允许只读复用或复制，实际编译进程、缓存和输出完全属于 Reviewer 副本。

| 实际命令 / 检查 | 实际结果与原件 |
| --- | --- |
| 第一次 `-Phase snapshot` | EXIT1；首样本 CPU87%、可用 6.803215GiB，在归档和工具/缓存复制前停止，没有 TeX/Python 调用。原始工具输出抄录在 `snapshot-first-attempt-error.txt`；该次 UTC 未保存，未补造。 |
| 按主控允许的延时单次 snapshot 重试 | EXIT0；准确 Git 归档、原提交文件保留及工具/缓存复制成功。CPU48%/54%、内存7.046/7.053GiB，单调间隔4.219秒；`resources-snapshot.json`、`candidate.zip`。 |
| 候选 `build.ps1 -Pandoc <自身Pandoc3.9> -Tectonic <自身复制的Tectonic0.17.0>` | EXIT0；完整 HTML、AST、TeX、TeX/xdvipdfmx PDF 编译成功，原件 `build.log`、候选副本 `build/manuscript.log`。实际使用候选内层双采样守卫，没有额外外层构建采样。 |
| 候选 `python -X utf8 .../papers/astrocyte/verify.py` | EXIT0；9引用/9条BibTeX/9条HTML参考文献、8配置、52MathML、26个PDF标题、9个参考文献标题、19个本地链接。PDF13页、38566字符、9个嵌入字体，无Type3、替换字符、加密或修复，所查文字span均在页面内；`verify.log`。 |
| 自身 `python -X utf8 .../compare-cycle02.py` | EXIT0；已提交与独立重建的两份 PDF 均13页、38566字符、9嵌入字体，并分别检查页外文字/字体。两份完整提取文本完全相同；提交 TeX 与单一正文实际重建 TeX 逐字节相同。PyMuPDF1.28.2；`compare.log`、`submitted-pdf.txt`、`rebuilt-pdf.txt`。 |
| 完整提取文本与来源、诊断静态核对 | 已读完整提取文本及完整TeX，覆盖摘要至9条参考文献。折扣关系式、softmax、八配置表、成本式、三项估计量、Holm式和样本规划式完整保留；结果未知及作者待填状态仍明确。 |

实际最终 TeX 日志在第 874 行保留一处 `Underfull \hbox (badness 2922)`，对应来源858–868行、§8源码路径段落；控制台因两轮 TeX 处理重复报告同一位置。错误/警告定位没有发现 Overfull、Missing character、未定义引用或致命 TeX 错误。数学上下标的文本提取顺序不等同视觉排版，文字边界检查也不证明视觉质量。

全部重操作串行、BelowNormal、Node1024MiB；通过阶段 CPU2%–54%、最低可用6.526GiB。build内层两样本为8%/8%，verify为2%/16%，均符合85%/4GiB门限和2–5秒间隔。首次拒绝及主控单次重试指令保留，没有循环采样或提高阈值；这些是启动观察，未报告峰值或速度改善。全部编译/Python调用已退出，无后台服务、浏览器、模型或产品原生输入；窗口通过 `msg_2b06952a1870` 提前交回，主控 `msg_9af8fdef8781` 已确认收回。

本轮新增的实际效果为：完整正文和引文现在具有可独立重建、与提交来源及PDF文字一致的13页交付。未将该结果当作正式八配置、适应效果、因果收益或完整产品验证。未对未变的9条文献重新下载或重复首轮研究接受。

### 官方要求与当前剩余限制

本次使用只读网页工具实际读取官方当前相关正文和处理器章节，访问日期2026-10-19。Firecrawl skill 已读，但 CLI 不在 PATH；没有为静态阅读安装依赖。来源和对应限制如下。

| 原始官方来源 | 与当前交付的关系 |
| --- | --- |
| [TeX submission requirements](https://info.arxiv.org/help/submit_tex.html)、[TeX Live at arXiv](https://info.arxiv.org/help/faq/texlive.html) | 当前支持TeX Live2025/2023及XeLaTeX；TeX稿件应上传必要源码并由作者检查站点PDF。本地Tectonic构建和审阅用 `preprint/` 仍未验证站点处理，也不是已批准的上传包。 |
| [Moderation policy](https://info.arxiv.org/help/moderation/index.html) | 官方要求原创、实质研究和显著AI使用披露，真人作者承担内容责任；研究提案可能被拒绝。当前方法/协议稿的研究贡献、完整定义与证据仍待后续审查，PDF构建不能证明投稿就绪。 |
| [Submission agreement](https://info.arxiv.org/help/policies/submission_agreement.html) | 实际提交须由具备作者/共同作者授权及材料权利者完成，并确定适用许可。作者、相关授权和论文许可继续待定，未代替真人接受协议或上传。 |

统计技术来源与完整分析程序、条件/时钟/归因、盲评/成本/样本规划及最新S3证据由后续轮次处理。正式八配置/扰动/自主迭代结果尚未采集。完整CI、原生/S3门禁、实际用户采用及正式V4边界保持，未由本轮通过。视觉检查未执行，Underfull保留；arXiv受支持环境重编译、本站处理与作者预览未执行。

UTC `2026-10-19T04:07:48Z` 实际发布 [CR 5566782506](https://github.com/songconmaisaix31-design/Astrocyte/pull/3#pullrequestreview-5566782506)，原始回执为 `COMMENTED`、`commit_id=618c82f270b4f35282dce86ef341b568a9690979`。同账号未作 GitHub APPROVE；实际请求/回执在 `02/github-review{,-receipt}.json`。本阶段报告普通 push 后执行当前生命周期的既有 hook，主控实际读取回复后登记本轮接受；本 Task 保持后续审查，不提前结算。
