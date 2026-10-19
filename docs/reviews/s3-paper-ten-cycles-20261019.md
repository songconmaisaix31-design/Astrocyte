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

实际读取主控 AGENTS、HANDOFF、STATUS、十轮任务、QUESTIONS、SPEC §7–8/10/12/17/19–20、TASKS P3 及研究产品决定；完整读取用户定位稿 v3 和预注册开发草案 v0.1。读取并应用 `astrocyte-research` skill 的证据分级、原始来源、完整配对块、因果识别、隔离和成本审查要求；本轮尚未审查新的论文候选。使用 orchestration/orca-cli 技能处理真实生命周期、消息、CR 通知与主控回复。

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

本阶段报告普通 push 后，按当前 Run/Task/Dispatch、实际终端、完整候选 SHA 和本报告非空绝对路径执行既有 review-decision hook，并交还唯一重操作窗口。hook 的真实消息和主控读取回复在后续阶段登记；入队本身不计作唤起确认，本 Task 保持运行，不提前发送 worker_done。
