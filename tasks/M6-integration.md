# M6 V1 真实回放组合树集成交接

2026-10-12；Task `task_0fff678cdbb0` / Dispatch `ctx_ffc79455aad8`，Run `run_d09ff3f95ebc`。原唯一集成owner继续工作树 `C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-final-integration`、分支 `songconmaisaix31-design/validation-v1-integration`，基线 `2524e9e48b23aa6f0ee4b6277a55539009d669dc`，客户端沿派发Codex `--yolo`。未启动其他Agent或变更模型；唯一自主修改为本报告，业务胶水为0。

指定普通合并、适用组合树回归及SOURCE发布已完成。root已批准V0，本轮真实V1回放与独立M2为技术PASS；**用户焦点确认与50权重人工核对保持PENDING，V1完整人验未批准，EXPERIMENT未锁，V2/V3/V4未启动**。事实源为 [V1任务](M6-V1-replay.md)、最新HANDOFF/STATUS及 [W1报告](M6-W1.md)。本任务不重复抓源、真实全回放或未变化的raw100审计，不以合成fixture替代人验。

旧Dispatch `ctx_3083dbc2d717` 因Orca重启为 **failed/terminal_missing**，已发布报告 `2524e9e48b23aa6f0ee4b6277a55539009d669dc` 和原测试证据保留；下方原报告仅是已完成工作的历史记录，不是accepted worker_done。本Task只使用新Dispatch回执，不追补旧假回执；所有首RED、unknown及原始载荷保持。

## 精确普通历史与SOURCE发布

开始时 `git status --porcelain=v1` 为空、分支匹配。两次 `git merge --no-ff` 均无冲突，未force/squash/rebase/cherry-pick或覆盖队友贡献：

| 精确输入 | 普通merge SHA / commit时间（+08:00） |
|---|---|
| root `7b739334d2e53d228e10fe3fd1223461078b0e1b` | `613606ec1933286c51ca707c92a2a4d41519b302` / 16:50:01 |
| W1 REPORT `b4243e41fd66d730f27db15e47fd85e6006ad92c` | **组合SOURCE `4d5665cbc058001c04466beb2590ee7796de78c5`** / 16:50:01 |

两精确输入、W1业务SOURCE `67cc8f6430ec3f274ac77d28ac7a4cd8c1d084e4` 及原报告2524e9e逐一 `git merge-base --is-ancestor <SHA> HEAD` EXIT0。相对root合并613606ec，W1恰好六个变化文件：`cmd/validation/main.go`、`cmd/validation/replay.go`、`cmd/validation/replay_test.go`、`scripts/validation/verify_v1.py`、`tests/validation/ingest/test_v1_reference.py`、`tasks/M6-W1.md`，全部位于W1写域。

相对本Task基线2524e9e，`git diff --exit-code` 核对全部app/domain、SQLite、attentionruntime、migrations/schema、contracts、web、Go/Node依赖和锁均零差异，EXIT0。其他变化为经授权普通merge消费的root任务/状态文档；集成owner未修改业务、参数或期待。

全部下列检查EXIT0后，组合SOURCE于 **16:51:50.227 +08:00** 普通push；`git ls-remote origin refs/heads/songconmaisaix31-design/validation-v1-integration` 确认精确 `4d5665cbc058001c04466beb2590ee7796de78c5`。本报告随后单独REPORT commit/push，精确SHA和时间通过本Dispatch worker_done提供。合并与报告使用 `[skip ci]`，root随后独立最终check/build/push/CI，不将旧CI归到新SOURCE。

## 本Task组合树实际验证

Windows / PowerShell；测试进程清除继承 `ASTROCYTE_*`，`GOFLAGS=-p=1`、`GOMAXPROCS=2`、Go `-mod=readonly`，`PYTHONDONTWRITEBYTECODE=1`。未开启真实源/隐私opt-in、未读个人实例、未并行执行pnpm或前端生成。

```powershell
go test -mod=readonly ./cmd/validation ./internal/adapters/events -count=1 -json
go test -mod=readonly ./tests/attention-engine -run '^TestM2' -count=1 -json
go test -mod=readonly ./tests/validation/sleep -count=1 -json
python -m unittest discover -s tests/validation/ingest -p 'test_*.py' -v
python scripts/validation/generate_kind_map.py --check
go vet -mod=readonly ./cmd/validation ./internal/adapters/events ./tests/attention-engine ./tests/validation/sleep
git diff --check
```

Go开始/结束 **16:50:51.698–16:51:11.846 +08:00**，三命令均EXIT0，共 **52顶层PASS / 0 FAIL / 2顶层SKIP**；包含子例81 PASS/5 SKIP，不称为81独立测试。

| 检查 | 顶层PASS / FAIL / SKIP |
|---|---:|
| cmd/validation（含两项生产V1测试） | 12 / 0 / 0 |
| events | 30 / 0 / 1 opt-in；含子例57 PASS/4 SKIP |
| 原M2 | 7 / 0 / 1 helper入口 |
| EX-01（validation/sleep） | 3 / 0 / 0；含子例5 PASS |

`TestV1ProductionSingleEventEmptyTopicClockAndThirtySleeps`、`TestV1RejectsDuplicateOutOfWindowAndOverwritingEvidence` 实际PASS：生产Recorder逐事件、无标持久化、真实30日Sleep/时钟及独立Python参考检查执行。它们是合成生产路径回归，不重复真实2832人口回放，也不证明人工焦点或权重验收。

events顶层SKIP为 `TestLocalV22CookedCommandByteParity` 本机真实源opt-in未开启；另三symlink子例因Windows权限条件SKIP。原M2 `TestM2ProcessReload` 是child helper独立入口SKIP，父测试实际执行进程重启；不将SKIP算PASS或本轮真实字节审计。Python开始/结束 **16:50:50.749–16:50:53.039 +08:00**，**13 PASS / 0 FAIL，.440s**，canonical map `--check` EXIT0。适用vet **16:51:18.610–16:51:19.750 +08:00** EXIT0；diff与冻结域检查EXIT0。

## 已完成真实证据的引用与限制

以下沿本轮最新root独立验收和W1报告引用，**不是本Task重新执行**：W1精确SOURCE67cc8f6的外部 `v1-w1/replay-002/` 有2832次单事件Record、2832信号、849无标、1893唯一commit、30生产Sleep含10空日，raw全BLOB留存。W1独立2169状态点含6seed；root从事前冻结V0源生成独立期望，对2163更新点及30日排名比较A/M/C误差0，并只读核对SQLite与机器权重样本50/50。原首回放和所有编译/断言/审计脚本首RED保持，不改expected或参数救验收。

模型焦点重建为09-22至10-08 Morphogenesis第一、10-09至10-11 Astrocyte第一；前10个空日全种子floor并列，不推断人类焦点。10-08包含当日前二3/4，严格当日后3/3。用户焦点确认与固定seed2026101201的50权重人工核对均 **PENDING**，机器50/50不是人工正确率。未配置facts/解释器，不补事实、不调用模型；Agent Swarm实际信号0，不补假链。

本Task所有新原件独占外部目录：`C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/integration-v1/task_0fff678cdbb0-ctx_ffc79455aad8/`，创建前核对不存在。含两merge日志、base/start/祖先、全变更与W1六文件清单、冻结域diff、三份Go JSONL/counts/skips、Python/generator/vet/diff日志、exit与起止时间、SOURCE/REPORT push/remote记录。raw、历史命令全文和数据库不stdout、不Git、不网络；不覆盖旧原件、不动个人库。

提交前附加身份核对首RED保留为 `w1-source-parity.log`：集成脚本误将W1报告文件与业务SOURCE67cc8f6一起要求零差异，而报告本来在独立REPORT b4243e4更新。随后分别比较五个业务文件对精确SOURCE、W1报告对精确REPORT，两个 `git diff --exit-code` 均EXIT0，原日志不覆写；这是核对脚本口径错误，不是业务/回归失败，没有修改任何SOURCE、REPORT或验收期待。

root后续独立整仓check/build和最终Git/CI；本Task未执行整仓pnpm、未重跑raw100/全人口/真实V1、浏览器E2E、官方客户端native task_live、人验、EXPERIMENT锁定、V2/V3/V4/七日真实三臂、main合并、部署、个人实例升级或rollback。组合树适用检查没有发现需薄胶水或退原W1的问题；本轮交付不代表完整V1人验或终局科学验收。

---

# M6 v2.2 普通集成与限定验证交接（历史保留）

2026-10-12；actual Task `task_63080055d561` / Dispatch `ctx_3083dbc2d717`，Run `run_d09ff3f95ebc`。原唯一集成owner、工作树 `C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-final-integration`、分支 `songconmaisaix31-design/validation-v1-integration`，本轮起始 `ffc7c0df8418918b85682269dc398080c5d00d08`。沿既有Codex `--yolo`，Orca terminal show实际preview **GPT-6.1-Sol high fast**；未新建Agent/worktree或更换模型。

指定普通合并及适用检查已完成；**自写业务胶水为0，唯一自主修改为本报告**。本任务不自批完整V0，不启动正式V1。业务事实沿 [v2.2裁决](M6-validation-v2.2.md) 和root最新HANDOFF/STATUS，下方v2.1/P0及所有首RED原件保留。原参数/主题待答已由v2.2裁决，不能沿下方历史继续报未答；6016全部工具人口及N≥5000/topic门槛已废止，规格修正与旧实现RED分别保留。

## 精确普通历史与源码发布

开始时工作树clean且分支匹配，依次普通 `git merge --no-ff`，无冲突：

| 精确输入 | 普通merge SHA / commit时间（+08:00） |
|---|---|
| root文档 `14daa596c48236e954c9d5677ad51b3cd7ad4366` | `9404ee6b26f163575063d99227975d93d1eaf906` / 15:56:42 |
| W1 REPORT `2b8137de8974c26d2da67c7319421ffe08a3d954` | **SOURCE `b4b5a388539a2428432cfa94133378dd9cf0d53a`** / 15:56:43 |

测试前完整组合树已存在；上述两精确输入及W1 business SOURCE `0b42bb97157b0b324c2aaa55c4badc08a488ed7b`、UTF8 SOURCE `867118002daa6d0c3897e6085b4de82f265d301a`、final SOURCE `c0f6936f2751cd22f0bc5f63c68211a37caa2436` 均经 `git merge-base --is-ancestor <SHA> HEAD` EXIT0。未force/squash/rebase/cherry-pick或删除未知文件，既有SOURCE/REPORT普通祖先完整。

全部适用测试EXIT0后，SOURCE于 **15:58:59.047 +08:00** 普通push，`git ls-remote`核实远端为精确 `b4b5a388539a2428432cfa94133378dd9cf0d53a`。报告随后独立REPORT commit/push，精确SHA/时间经本Dispatch完成消息回root；合并与报告使用 `[skip ci]`，最终Git/CI归root独立验收。

## 对原root基线的冻结域实际diff

固定比较基线 **`4538be918ec34eda573ccbbbd687ba41f8954d1b`**。`git diff --exit-code` 核实 domain、SQLite/store、migrations、contracts、web、Go/Node依赖与锁、全部其他app及EX-01生产Sleep/测试均原样，EXIT0，schema零豁免。

唯一app差异为 `internal/attention/app/knowledge_retrieval.go`：`KnowledgeService.RetrieveKnowledge`调用同文件新增纯计算 `knowledgeReferenceBoost`，复用已有human reference `ActionAudit.TS`，排序复制时间序列、cap2、24h无访问回退1，读路径不新增持久化写入；import新增sort供helper使用。M4两fixture实际diff只加非正式超限输入说明并将独立乘积/AccessBoost期待3→2，**原ReferenceBoost=2输入不变**；正式绑定+.1，无领域公式/权重/λ/Sleep接缝变化。diff原件在外部日志，详细业务实现归 [W1报告](M6-W1.md)；集成owner无业务或测试期待返修。

## 组合树适用验证

Windows / PowerShell；`GOFLAGS=-p=1`、`GOMAXPROCS=2`、`-mod=readonly`、Python `PYTHONDONTWRITEBYTECODE=1`。Go测试进程先清继承 `ASTROCYTE_*`，显式设置本机 `ASTROCYTE_V22_COOKED_LOCAL=C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/root-v22/v227-sample-shell-LOCAL.json` 与 `ASTROCYTE_V22_FROZEN_LOCAL=C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v21-w1/tool-records-local.jsonl`，检查两文件存在后执行。没有运行全人口diagnostic/replay，也没有并行pnpm/frontend生成。

```powershell
go test -mod=readonly -count=1 -json ./cmd/validation ./internal/adapters/events ./internal/attention/app ./tests/attention-engine ./tests/validation/sleep ./tests/validation/privacy ./internal/adapters/mcp
python -m unittest discover -s tests/validation/ingest -p 'test_*.py' -v
python scripts/validation/generate_kind_map.py --check
go vet -mod=readonly ./cmd/validation ./internal/adapters/events ./internal/attention/app ./tests/attention-engine ./tests/validation/sleep ./tests/validation/privacy ./internal/adapters/mcp
git diff --check
```

Go开始/结束 **15:57:17.669–15:57:47.138 +08:00**，七包EXIT0，**184顶层PASS / 0 FAIL / 12 SKIP**；含中间节点及子例317 PASS/15 SKIP，不称为317独立测试。

| 包 | 顶层PASS / FAIL / SKIP | 包elapsed |
|---|---:|---:|
| cmd/validation | 10 / 0 / 0 | .945s |
| events | 31 / 0 / 0（含子例58 PASS/3 SKIP） | 2.146s |
| attention/app | 49 / 0 / 0 | .671s |
| attention-engine | 67 / 0 / 6 | 7.131s |
| validation/sleep | 3 / 0 / 0 | 1.120s |
| privacy默认 | 11 / 0 / 6 | 3.716s |
| MCP adapter | 13 / 0 / 0 | .314s |

| 要求分列的回归 | 顶层PASS / FAIL / SKIP |
|---|---:|
| 原M2 | 7 / 0 / 1 helper入口 |
| M5 MCP | 6 / 0 / 1 helper入口 |
| EX-01 | 3 / 0 / 0 |
| M4（含两原超限fixture） | 11 / 0 / 0 |

`TestLocalV22CookedCommandByteParity` **实际PASS/.99s**，核对5对真实源/独立cooked字节及原生AST桥；`TestNativeShellSyntaxBridgePreservesUTF8`实际PASS，默认无本机输入SKIP不能代表此结果。固定解析器只消费stdin数据并ParseInput，不执行历史JS/命令，raw与命令全文不打印。参数生产测试通过既有SQLite/restart/query_only核对六种子confidence=.5、s=1/w冻结表、λ_r=.05/86400秒⁻¹及初始1/+.1/cap2/24h回退1。

引擎六顶层SKIP为两个live metadata与四个child helper入口，child通过父测试实际运行；events三symlink子例因Windows权限条件未执行。privacy默认六opt-in SKIP不算审计PASS，本轮实际独立执行下方real100，未另跑旧SQLite opt-in或旧四canary。Python **8 PASS / 0 FAIL，.133s**；generator `--check` EXIT0；适用七包vet EXIT0；diff check EXIT0。Python/vet开始/结束 **15:57:17.886–15:57:26.147 +08:00**。

## 原100 raw_unique实际生产审计

独立进程清继承 `ASTROCYTE_*` 后显式 `ASTROCYTE_V21_REAL100_AUDIT=1`，执行 `go test -mod=readonly ./tests/validation/privacy -run '^TestReal100V21ProductionExposure$' -count=1 -v`，结束清除opt-in。开始/结束 **15:58:13.626–15:58:26.511 +08:00**，EXIT0，测试11.53s/包12.014s，**1顶层/100 event-case PASS、0 FAIL/0 SKIP**。

原Codex50/Claude50、**1,101逐事件去重raw_unique字符串**、**2,100实际生产MCP.Serve响应**严格违规0。分布：success500/backend error500/decoder error500、report/veto capture200/dynamics200/unknown delivery200。actual native probe accept100/kind mapped4/topic assigned3/unmapped96/association incomplete1；另100个 **`real_raw_contract_local`** 合成信号只为原raw穿越真实SQLite/五工具/ExplainFocus错误路径提供本机审计，不改变真实semantic population，不代表native eligible、task_live、自然失败、正式V1或完整V0。原字节只读、原RED不覆盖，没有替换冻结100或自制响应绕过生产Serve。

## 外部原件与后续边界

本Task独占新目录（创建前确认不存在）：`C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/integration-v22/task_63080055d561-ctx_3083dbc2d717/`。含两merge日志、base/start/ancestor、changed-paths/frozen-domain-diff/knowledge-retrieval.diff/m4-fixtures.diff、`go-targeted.jsonl`、counts/skips/selected/local-bridge计数、Python/generator/vet日志、`real100-audit.log`、全部exit/起止时间及SOURCE/REPORT push/remote记录。所有新测试日志只落该外部目录；raw/参数全文不stdout、不Git/远端，未移动/覆盖任何旧RED或原始载荷，未改个人库。

沿root已完成独立源join和桶审查引用，不重复全人口：**原6016父 / 8914语义单位；2832/2964=95.5466%数值门槛；132结构失败保留；5950非认知；原200误标0；topic1983/2832=70.0212%，849无标**。每桶20共160源定位和新增22 nonsearch调用/50候选命令由root独立核对。topic只报告，无标kind成功保留；结构失败不挪出分母。数值门槛与本次审计/单元回归不自批完整V0。

**root仍需新独占全人口、完整pnpm check/build、独立隐私及最终Git/CI，三指标独立通过并明确批准后才开V1**。本任务未执行正式入库/逐ts回放/30Sleep/真实人口M2复算/A(t)/焦点人工确认；EXPERIMENT未锁，V2–V4/真实七日前瞻、官方客户端/native task_live、部署/个人实例升级/main合并/rollback均未执行。旧failed/first RED/delivery_unknown保持，未编造人工确认。本轮未发现需要集成业务胶水或退W1的失败。

---

# M6 v2.1 完整组合树集成交接（历史保留）

2026-10-12；Task `task_60592bb5f8de` / Dispatch `ctx_50eb89deca8f`，Run `run_d09ff3f95ebc`。原唯一集成 owner、原工作树 `attention-engine-final-integration`、分支 `songconmaisaix31-design/validation-v1-integration`，基线 `17d77c443d44ec260d86de7322292874eff30059`。沿原 Codex `--yolo`，实际 Orca terminal show preview 为 **GPT-6.1-Sol high fast**；未另开 Agent或更换模型。

本轮指定普通合并、组合树定向测试及原100条生产隐私审计完成并通过；**V0整体仍 RED，正式V1未执行**。唯一自主修改为本文件，零薄胶水/业务代码修改。下方P0历史和所有原RED保持，以本段及 [v2.1规格](M6-validation-v2.1.md) 覆盖旧的待provider出生/raw侧表方案；v0.1无标自动出生冻结至v0.2，schema零豁免、raw仍本地原store保留。

## 普通合并与发布

开始时 `git status --porcelain=v1` 为空、分支匹配。按指定顺序普通 `git merge --no-ff <SHA>`，三次均无冲突，未squash/rebase/cherry-pick/force或覆盖队友贡献。

| 合并输入 | 生成merge SHA |
|---|---|
| root计划/文档 `2f7b0c2975f3faf465bb1637b6bf1bf49ecca68d` | `c48f85a501fd9e794a6dcc710efe6854f2ba1f7a` |
| W1 REPORT `3020bfd7a47e1b83f765ab3c1ad0437841746af6` | `b44038a589fddcd4656ba0656dbfd5229faf8956` |
| W0 REPORT `88268a05a756f108f76422e1279e734851e4c324` | **SOURCE `25457b27d1bca7fb3d1042e91dc6c7eaabf53f1f`** |

测试前完整组合树已存在。上述三个精确输入及 W1 SOURCE `cadd33af594eee8f7d246bb9ba2b292241dfdca2`、W0隐私SOURCE `6e7de7e340606f6f13fad79f00a814b455daed72`、W0 M5 fixture SOURCE `b69ec6fffc109bf1ed8242a59699511aca967060`，逐一 `git merge-base --is-ancestor <SHA> HEAD` 均EXIT0。两轨全部SOURCE/REPORT普通历史保留。

相对基线 `git diff --exit-code` 核对 `internal/attention/app`、`internal/attention/domain`、`migrations`、`contracts`、Go/Node依赖与锁、`web` 均零差异。W1变化为canonical映射/生产工具块与wrapper入口/全人口诊断及对应测试，W0变化为MCP曝光投影/错误回执和对应隐私、CLI及M5 fixture测试；owner详细语义及首错见 [W1报告](M6-W1.md)、[W0报告](M6-W0.md)。root文档仅通过授权merge消费。

SOURCE commit时间 **04:52:52 +08:00**，在所有下列测试EXIT0后 **04:54:52.573 +08:00** 普通push，`git ls-remote origin refs/heads/songconmaisaix31-design/validation-v1-integration`确认精确 `25457b27d1bca7fb3d1042e91dc6c7eaabf53f1f`。本文件随后单独REPORT commit/push，精确REPORT SHA和时间由本Dispatch完成消息提供；各合并/报告使用 `[skip ci]`，root负责最终Git/CI验收。

## 组合树实际测试

Windows / PowerShell，Go `-mod=readonly`，`GOFLAGS=-p=1`、`GOMAXPROCS=2`，Python `PYTHONDONTWRITEBYTECODE=1`。测试进程先清除继承 `ASTROCYTE_*`，仅独立real100进程显式开启指定audit并在结束清除。未调任何业务参数。

root自然检查点消息 `msg_0fec78ba2bbd` 明确修正：`tests/validation/ingest`无Go文件，移除其Go positional package，以Python discover单独执行，其他验收不变。实际命令：

```powershell
go test -mod=readonly ./internal/adapters/events ./cmd/validation ./internal/adapters/mcp ./tests/validation/privacy ./cmd/mcp ./tests/validation/sleep ./tests/attention-engine -count=1 -json
python -m unittest discover -s tests/validation/ingest -p 'test_*.py' -v
python scripts/validation/generate_kind_map.py --check
$env:ASTROCYTE_V21_REAL100_AUDIT='1'
go test -mod=readonly ./tests/validation/privacy -run '^TestReal100V21ProductionExposure$' -count=1 -v
```

Go开始/结束 **04:53:23.173–04:54:02.361 +08:00**，EXIT0。七包合计 **122顶层PASS / 12顶层SKIP / 0 FAIL**，计入中间节点和子例为230 PASS/15 SKIP；后者不称为230独立测试。

| 包 | 顶层PASS / SKIP | 包elapsed |
|---|---:|---:|
| events | 19 / 0（含子例46 PASS/3 SKIP） | 1.337s |
| cmd/validation | 8 / 0 | .557s |
| MCP adapter | 13 / 0 | .510s |
| privacy默认 | 11 / 6 | 5.146s |
| cmd/mcp（实际构建/运行CLI） | 1 / 0 | 5.527s |
| validation/sleep（EX-01） | 3 / 0 | 1.135s |
| attention-engine | 67 / 6 | 9.882s |

原M2 **7 PASS/1 helper SKIP**，M5 MCP **6 PASS/1 helper SKIP**；实际stdio/SQLite重启与错误提交断言执行。EX-01三项通过，保留tagged禁生、unknown收据与已有topic更新/merge断言。六个引擎顶层SKIP为两个live metadata入口及四个child helper入口，child通过父测试实际执行；events三个symlink子例因Windows privilege不足未执行。privacy默认六个opt-in SKIP不计为隐私通过；本轮只单独实际执行下面real100，不另跑旧SQLite opt-in或四项旧canary审计。

Python开始/结束 **04:53:22.768–04:53:24.504 +08:00**，**8 PASS / 0 FAIL，.204s**；generation `--check` EXIT0，canonical Python与Go生产表无漂移。`git diff --check`和冻结域/精确祖先核对EXIT0。

## 原100生产曝光隐私审计

开始/结束 **04:54:17.774–04:54:34.485 +08:00**，EXIT0；测试14.86s/包15.531s，**1顶层/100 event-case PASS、0 FAIL、0 SKIP**。严格逐字符串审计 **1,101个逐事件去重raw_unique值**与 **2,100实际生产MCP响应**，违规0；原冻结Codex50/Claude50原始字节及observed association只读，未重抽样、重复全人口扫描或改原RED。

组合树实际native probe：accept100、kind mapped **4**、topic assigned **3**、unmapped **96**、association incomplete1。相较W0单轨旧mapped3/unmapped97，此组合消费W1新增wrapper规则，与W1同冻结100复验一致；不是改分母或新增认知资格。另100个 `real_raw_contract_local` 合成信号使用原raw测试真实SQLite/生产MCP曝光路径，**不等于native eligible或task_live**。它们不把96个unknown重新标成真实认知、不证明官方客户端实际调用或正式回放。

响应分布：五工具success500/backend error500/decoder error500；report/veto capture error200/dynamics error200/unknown-delivery error200。错误注入是本地受控传播，不当成观察到的native失败。完整raw BLOB本地逐字节保留；报告、终端、Git与协调消息均无raw正文/参数/session/base64。

## 外部日志与真实剩余项

全部新原件目录：`C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/integration-v21/task_60592bb5f8de-ctx_50eb89deca8f/`。该目录创建前确认不存在，未覆盖任何历史证据。包含三份merge日志、base/start/祖先记录、changed paths/frozen diff、`go-targeted.jsonl`、`test-counts.json`/`test-skips.json`/`selected-test-counts.json`、Python测试和生成日志、`real100-audit.log`、命令exit/起止时间、push/remote记录。测试使用既有隔离临时数据库，不改个人库；raw样本只按已授权本地audit读取，不外发、不stdout、不入Git。

root此前独立全人口结果沿原证据引用，**本任务未重复盘点**：N6016（≥5000）、complete1971/6016=**32.76% RED**（<95%）、negative200 false0、unassigned3497。隐私定向通过不改变V0另外门槛，也不代表全人口raw_unique/HTTP/UI全边界通过。unassigned继续留存计数，不缩分母、不伪造主题，不恢复v0.1自动出生。

实际代码问项沿W1报告：Signal.Strength非负s、Weight为signed w，实际s*w，probe s=1不能把s=w后再乘；λ_r用AgeSeconds，需要确认用户.05的单位与显式绑定；retrieval boost虽可显式+.1，现有cap2/24h冷却缺失，冻结核心下未实现；InitialConfidence、成功delta/conflict threshold、retrieval age basis等正式运行完整绑定仍需root裁决。v2.1已定参数值不擅自改写，未为这些差异调公式或启动正式V1。

root后续独立完整 `pnpm check`/`pnpm build`、组合树隐私审计与最终Git/CI；本任务未执行全仓check/build、浏览器/E2E、全人口正式入库、V1/30Sleep/真实M2复算/焦点人工确认、V2–V4/真实七日前瞻、官方客户端/native task_live、paid调用、main合并、个人实例升级、部署或rollback。旧first RED/failed/delivery_unknown保持。本轮不宣称V0整体、真实M2/V1或科学验证完成。

---

# M6 P0 限定集成交接（历史保留）

2026-10-12；Task `task_948e5c3f0432` / Dispatch `ctx_7b2d6d2c5891`，Run `run_d09ff3f95ebc`。沿原唯一 Codex `--yolo` 终端与工作树，实际模型沿客户端配置，本任务没有更换或额外启动 Agent。分支 `songconmaisaix31-design/validation-v1-integration`，基线 `ffd889db7156d580212f914a76a8ab8fb19a1643`；原工程分支保留。

本次普通合并与指定定向检查已完成；V0/V1 及正式实验未通过。唯一自主写入为本文件，无薄胶水或业务修补。

## 精确普通历史

依次 `git merge --no-ff` 两位 owner 的最终 REPORT，均无冲突。源码合并 SOURCE **`43746b74cdb1411efb29646e1e7306191abfd714`** 已独立 commit+push；本文件随后独立报告 commit+push，准确 REPORT SHA 随 worker_done 回 root。两次 merge 使用 `[skip ci]`，最终交付远端检查归 root；无 squash/cherry-pick/force、历史删除或未发布内容消费。

| Owner | 精确 SOURCE / REPORT / 保留诊断 |
|---|---|
| W1 | SOURCE `902b3b1f204d5197ed9850126a53d88293ba5068`；REPORT `a227ca5298d6d7b69ad58708493b310969ef2963`；诊断 `b745f06dcebdda44ff919582380ca81d5abb573d` |
| W0 | EX-01 SOURCE `2947454ff697fb759b32a0d62aea8c8fccf04b4d`；REPORT `9f654e6fa122064626b1edcec3bf1af34df0e79a`；privacy 诊断 `3c6454651063539150169d334b8ea2d7fa9fe7c6` |

以上六个精确 SHA 均通过 `git merge-base --is-ancestor <SHA> HEAD` / exit0，普通历史完整。

## 冻结域核对

相对基线的引擎生产差异只有 `internal/attention/app/sleep_service.go`，两个 diff hunk 均在 **SleepService.Run**：

1. 已存 mapped 空 Topic 与 Event=nil 记录不再无记录跳过，保留原 event_id 与真实观测时间的 `unassigned_pending` / `mapping_pending` 候选收据；未知 bucket/时间仍显式未知，不补标签或 ReceivedAt 代替事件时间。
2. tagged 簇缺节点时记录 `tagged_birth_disabled_v0_2` 并跳过出生分支；已有主题的信号、证据、邻接、fluid 和独立 association merge 路径保持。

`internal/attention/domain/**`、其他 app 文件、A/M 数学、固定权重、λ/去噪/fluid/crystallized、`sleepBirthSupport`、共享 `ConsolidateTopicInTx`、`observeSleepAssociations`、M2 文件、SQLite/MCP/schema/migrations/contracts 均未改。root 的 STATUS/HANDOFF/plan/EX-01决定/EXPERIMENT草案也未改。已将冻结域相对基线 `git diff --exit-code` 与全引擎变更文件清单核实为 exit0 / 仅 Run 文件；diff 原件单独保留。

W1 的生产变化限 events 接入规则/observed association/固定 seed 匹配和实际 git-log adapter；Python literal 规则是唯一规则源，生成 Go 表。本轮核对 `MapGitLogRecord` 要求真实 log record 的完整 commit ID 与 TS；shell 中调用 git commit 不形成已提交事实。关联缺失、unknown、unassigned 不补造。其他变化是诊断命令/脚本、合成测试和 owner 报告；未扩展引擎/schema。

## 本集成任务实际检查

Windows / PowerShell；Go PATH 为 `%LOCALAPPDATA%/Programs/go/bin`，`GOFLAGS=-p=1`、`GOMAXPROCS=2`，`PYTHONDONTWRITEBYTECODE=1`。测试进程清除继承的 `ASTROCYTE_*` 参数，live/私人样本 opt-in 未开启；没有目录扫描、原始 100 条读取或模型调用。

`go test -mod=readonly ./cmd/validation ./internal/adapters/events ./tests/attention-engine ./tests/validation/sleep ./cmd/sleep -count=1 -json` **PASS / exit0**：

| 包 | 顶层 PASS / SKIP | 包 elapsed |
|---|---|---|
| cmd/validation | 7 / 0 | .331s |
| events | 14 / 0；计入子例为 41 PASS / 3 symlink SKIP | .696s |
| attention-engine | 67 / 6；计入子例为 87 PASS / 6 SKIP | 6.450s |
| validation/sleep | 3 / 0；含子例5 PASS | .809s |
| cmd/sleep | 1 / 0，实际 built Sleep CLI/日幂等及重启 | 2.476s |

共 **92 顶层 PASS / 6 顶层 SKIP**。六个顶层 SKIP 为既有 live metadata/W7 metadata 与 M2/M3/M6/MCP 子进程入口；父测试实际调用的 child helper 不视为未执行。events 三个普通 symlink 子例因 Windows privilege 不足为 NOT_RUN/SKIP。M2 7 PASS/1 helper SKIP、M3 10 PASS/1 helper SKIP、M6合成周1 PASS/1 helper SKIP、EX-01三顶层 PASS；不从合成 fixture 推断真实无标出生或真实回放。

- `python scripts/validation/generate_kind_map.py --check`：PASS / exit0，canonical Python 与生产 Go 表无漂移。
- `python -m unittest discover -s tests/validation/ingest -p test_sampling.py -v`：5 PASS / exit0，.044s。
- `git diff --check`、精确祖先及限定变更路径核对：PASS / exit0。

外部独占原件目录：`C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/evidence/integration-p0/task_948e5c3f0432`。保留 `go-targeted.jsonl`、`test-counts.json`、`python-generation.log`、`python-ingest.log`、各命令 `*-exit.txt`、`source-changed-paths.txt`、`sleep-run.diff`、`diff-check.log`；只含合成测试/聚合检查输出，不含真实 native raw。测试目录按既有 helper 生命周期清理，本轮未另留测试 SQLite。

## 仍未通过与 root 下一步

双方原诊断及 root 基线已记录：同一冻结100条 retained100 / kind-mapped3 / unknown97 / topic+association候选3，**formal eligible0**。W1 实际1893条 git-log commit 与原100分母分开，不能拿记录数或 scope assigned 数当认知 eligible。本集成任务未重复扫描/抽样/同100复测，root 将在新树独立复测；不能声明V0≥90%或V1≥10000/30Sleep/真实焦点迁移通过。

EX-01 的无标真实候选 label/provider 缺失，**正向 unassigned discovery/birth BLOCKED**。当前 pending 保留与 tagged 零出生仅是限定门禁证据，不能替代正向出生验收或伪造自主发现主题。原 M3/CLI/M6 与路径测试首 RED 沿 [EX-01报告](M6-EX01.md) / [W1报告](M6-W1.md) 保留，fixture 预播种使用既有 human SeedTopic，不改策略追旧值。

**raw 泄漏诊断仍 RED**：W0 原最终 opt-in诊断5顶层 FAIL/4 PASS，含通用事件读、explain_focus 的原文/base64及错误回显路径；本轮没有重跑该审计，也没有因其 opt-in默认不执行而声称隐私通过。019 schema/derived隔离/MCP投影仅是 [W0方案](M6-W0.md)，当前没有schema豁免实施。raw与derived同值审计口径及正式运行参数仍待裁决；不选默认、不修冻结域绕过阻塞。

EXPERIMENT.md 包含 EX-01 的**未锁定草案**，V1未过；不启动day0、不锁V3、不生成终局 EXPERIMENT_REPORT、不把历史回放当真实七日前瞻。原工程M1 failed/首RED历史保持；本轮授权的最近30天本地读取范围沿 root 决定，但集成任务没有再次执行真实读取。

root 下一步普通合本 REPORT，执行新树完整 pnpm check/build、同100独立复测与最终交付远端检查；本任务未重复整仓UI/媒体/paid、未升级个人实例、未合main、未执行rollback。EX-01 rollback对象仍为 `2947454ff697fb759b32a0d62aea8c8fccf04b4d`，只记录原 owner 可 review 的普通 revert 对象，没有执行或补造回滚 SHA。
