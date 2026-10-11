# M6 v2.1 完整组合树集成交接

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
