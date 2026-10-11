# Attention Engine 最终集成交接

本次限定集成已完成；整体产品与真实七日科学验证未完成。Task `task_9c5f23440239` / Dispatch `ctx_7eb57b08004b`，实际客户端 `codex --yolo`。隔离工作树 `C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-final-integration`，分支 `songconmaisaix31-design/attention-engine-final-integration`。

## 精确集成源

基线 root `dd62ec98b263e84b50124b2106ab1ce887ce6942`。普通 `git merge --no-ff` 合入 W0 最终 REPORT，生成 SOURCE `a806425ab1477cab9ebd9fdfecf0c0c96eea94d0`；双亲为 root 基线与 W0 REPORT。合并无冲突，未修改领域代码或添加胶水。其余七轨最终报告已在 W0 普通历史中，逐项 `git merge-base --is-ancestor <SHA> HEAD` 均 exit0，无需重复合并。没有消费未提交内容，没有 squash、force push、删除历史或改 root 独占文档。

| Owner | 精确最终 REPORT SHA |
|---|---|
| W0 | `5c1aa0dc41358db29c3cc12c9a250f19b7fc69b0` |
| W1 | `360730c05a739f8c28177041795c856e64dba0c2` |
| W2 | `2c148c1034ec28b5d787309f1320198706d97e4a` |
| W3 | `39cb8fec12b5fedcdc82578d9b678c3857688d72` |
| W4 | `11899c92dd4687f319fb084bfad7218662545df3` |
| W5 | `737a4b04cd90953077662928a99c814d870dc2d4` |
| W6 | `a1ee679aed7c09fc35bb47d1885861c83bbdf484` |
| W7 | `d4dbc8bca11c01af2bb4480bccb4b2d07be1b676` |

W0 强化 M6 SOURCE `f4bf0257b3f2e93e2ce67197bfbb0764a68332a3`；W4 推荐包装解码修复 `abff94dbb8b614e4489e467e9095935fd0815ba9`，W3 字面事实投影 `61a196ba3e3c24606a72d6fecb40427281a23ae5`，亦保留在普通历史。报告提交只新增本文件，准确 REPORT SHA 随 worker_done 交 root。

## 本集成 Agent 实际验证

环境：Windows / PowerShell；Go 使用 `%LOCALAPPDATA%/Programs/go/bin`，`GOFLAGS=-p=1`、`GOMAXPROCS=2`、`NODE_OPTIONS=--max-old-space-size=512`。依赖未安装，实际执行 `pnpm install --frozen-lockfile` PASS，pnpm 11.27.0，锁与依赖版本未变。

| 命令 / 操作 | 实际结果 |
|---|---|
| `git diff --check` | PASS / exit0，无诊断 |
| `go test -count=1 ./tests/attention-engine -run '^TestM6' -v` | PASS / exit0，包 1.670s；主用例 1 PASS / 0.76s，冷进程 helper 独立顶层 1 SKIP，但父用例实际派生调用并验证成功 |
| PowerShell 直接解析新 JSON，以图中 A、首次创建年龄、人工 reference 审计独立复算推荐 | PASS；最大绝对误差 `2.220446049250313e-16`，两项 A 与图匹配且得分非零 |
| 八轨最终 SHA 祖先检查 | 8/8 exit0；root 基线也保留 |

本轮 M6 通过正式 Sleep factory 注入合成 Clock，调用真实生产服务、SQLite、HTTP 人工关口和新的冷 Go 子进程。30 条原始观察、7 个日周期、1 条冲突、1 条批准经验；第 1–4 天 focus=research，第 5–7 天 focus=implementation。确认/分享/引用、只追加固化版本、失败仅降 confidence、pending 不生成 Signal、原 raw BLOB 冷重载一致和 query_only 推荐均由本次主用例验证。

| 最终 topic | 图中 A = 推荐输入 A | 实际 score | 独立 score |
|---|---|---|---|
| implementation | 0.9552011491162884 | 1.0480013210594767 | 1.0480013210594765 |
| research | 0.33674673723047327 | 0.3350055591708292 | 0.3350055591708292 |

独立公式为 `A × exp(-1e-8 × age_seconds) × (1 + 0.1 × human_reference_count)`，本查询无 topic/text 约束，relevance=1；显式年龄基准 created，最终合成时间 `1791414000002`。这不是默认政策决定或仅字母序检查。

本任务独占原始输出目录：`C:/Users/DW/AppData/Local/Temp/astrocyte-attention-final-integration-task_9c5f23440239`。保留 `M6-integration.log`、`M6-exit.txt`、`diff-check.log`、`recommendation-independent.json` 与 `M6/M6-week-result.json`（270881 bytes）；`ASTROCYTE_M6_REPORT_DIR` 指向该独立 M6 子目录。测试 SQLite 使用既有 `t.TempDir()` 生命周期，冷进程验证实际完成，但本轮未另留数据库副本。

首次全零推荐原件继续保留于 `C:/Users/DW/AppData/Local/Temp/astrocyte-M6-W0-20261011/M6-first-zero-ranking.json` 与 `M6-week-first-zero-attention.json`；本轮不覆盖旧文件。W4 首 node-wrapper RED、W0 paging/ENAMETOOLONG/summarize 探测超时及其他 owner 首错仍沿各原报告保存。

## 已有 owner 验证与下一步边界

W0 REPORT 已记录该 owner 实际 `pnpm check` 和 `pnpm build` PASS（27 包架构、275 契约例、S1 API14、前端90项），实际 HTTP/stdIO/Sleep 单日命令与重启通过；构建产物包括 server/MCP/Sleep。以上是 owner 原验证，本集成任务未重复全 check/build 或实际 Sleep CLI。正式 factory 七日合成验证与 actual-built Sleep 单日/重启分别成立，没有添加历史产品 CLI 回放。

W1 原 Dispatch 因真实 Claude Code/Codex metadata 范围未答以 failed 结算；本地实现可合，M1 真实两客户端验收 NOT_DONE。原生来源读取、工程默认参数、知识版本再确认与检索年龄政策仍未决定，运行要求显式配置；数据保留、项目隔离和人工显式分享沿既有实现，不代选默认。sqlite-vec CUT，检索明确使用 `sqlite_vec_cut:topic_keyword`。

官方 MCP client/native Agent 互操作、task_live 和连续七日真实使用对基线比较均 NOT_RUN；没有私人来源、模型、媒体请求或旧 UNKNOWN 重发。W7 真实作品 UI/批准根/留存库链1 pending observation 验证沿 owner 报告，不因本轮合成 M6 成为真人七日认知证据。

root 下一步普通合入本 REPORT，在交付分支独立执行最终 `pnpm check`、`pnpm build` 与适用 E2E，并完成主验收文档和最终 push。本任务不执行 main 合并、远端 CI、个人预览升级或用户主观验收；本分支 SOURCE 与文档 REPORT commit+push 后 bounded succeeded 交接。
