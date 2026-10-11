# M6 P0 限定集成交接

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
