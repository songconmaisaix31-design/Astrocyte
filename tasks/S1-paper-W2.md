# W2 项目汇总与原生输入扩容

## 2026-10-10 最新活动 Git 时间补漏（独立 follow-up）

Task `task_5531618a1ae6` / Dispatch `ctx_a12876d7d525`，原工作树/分支 `s1-local-agents-1010`，客户端 Orca Codex、模型继承当前运行环境，未切换 provider。基线原 SOURCE `4201c899f5aa0285a0d19442b674327237967739` / REPORT `9e47c455ba07d83a66696b5bfeccf46ffff4f33c` 保留，本次业务 SOURCE `cd201962997e69bb4772d73a6d8752874085c3ef`；本段随后独立提交为 REPORT，不修改原回执。

聚合最新活动加入每个成员目录非空的 `Git.LastCommitAt`，与既有 Orca 活动取最大值。提交时间在适配器的 dirty 状态查询前独立取得，因此总体 Git status 为 unknown 但非空提交时间仍参与；nil 仍未知。原生 header 的创建/观察时间不作为最新活动，创建字段保留。只修改领域聚合、领域测试、所属应用测试和本报告，没有修改契约、路由、UI、来源采集或模型/媒体。

| 验证命令/行为 | 结果与原件（`%TEMP%`） |
|---|---|
| `go test ./internal/workspace/domain -run 'TestProjectBoardLatest' -count=1 -v` 修复前 | RED，复现 native-only、Git 比 Orca 更新、Git 部分状态失败、跨成员目录遗漏；`astrocyte-w2-git-activity-first-red.log` 保留 |
| `go test ./internal/workspace/domain ./internal/workspace/app -count=1 -v` | PASS；3领域测试含8边界子例，缓存列表应用测试PASS，未提供缓存路径时实际缓存例明确SKIP；`astrocyte-w2-git-activity-targeted.log` |
| 首实际缓存 opt-in | RED，Windows SQLite URI 错把盘符作 authority；测试 URI 修正为 absolute file URI，原件 `astrocyte-w2-git-activity-existing-cache.log` 保留 |
| `$env:ASTROCYTE_TEST_PROJECT_ACTIVITY_CACHE=Join-Path $env:LOCALAPPDATA 'Temp/astrocyte-s1-meg3Pi/data/state.sqlite'; go test ./internal/workspace/domain ./internal/workspace/app -count=1 -v` | PASS/exit0；5顶层测试含上述8子例，`astrocyte-w2-git-activity-cache-retest.log` |
| `go build -mod=readonly ./...` | PASS/exit0；`astrocyte-w2-git-activity-go-build.log` |
| `gofmt` / `git diff --check` | PASS |

实际缓存验证用 SQLite `mode=ro` 读取先前已观察 JSON，再调用真实 `ListRegisteredProjects` 应用服务；注入缓存 repository 与记录调用次数的 source，复用 `withProjectMetadata` 重投影。111目录/71组，旧7组有活动时间，新17组，恢复10组已有 Git 提交时间；source调用0、repository写入0，持久化 JSON 查询前后完全一致。项目/来源统计/失败/观测时间不变；人工字段回归使用测试注入的备注/归档/revision，未宣称读取或改动实际人工字段。

限制：剩余54组最新活动仍未知，不能由会话创建或目录时间补造。此次只验证缓存应用服务路径，未启动HTTP服务器、更新个人预览或跑浏览器；W0普通合并/预览更新与W3后续界面检查由对应owner处理。未全量重新扫描本地项目、调用模型/媒体、重发UNKNOWN、运行全套后端发现或 `pnpm check`，此前真实与完整验收原件继续保留。源码与报告均在原分支普通commit后push，root/W0/W3接收精确SOURCE/REPORT作集成依据。

分支 `s1-local-agents-1010`。本轮 Task `task_bade38476a47` / Dispatch `ctx_4434022e51cd`，客户端 Orca Codex，模型由运行环境继承。基线 `087127d30fc505f8fe25c52142a174f5842e2bef` 普通合并为 `0ca476d`；W0、W1、W3 依赖只普通合入已发布 SHA，未修改其他写域。

## 已交付

- selected-text 最终 UTF-8 输入512KiB（含指令/信封），启动前拒绝超限、无截断；输出128KiB和原生历史128KiB分别保留。导出校验/准确字节测量，实际发送前输出仅长度的 `selected_text_input` 指标；调用方/项目许可不变。
- Orca 登记与固定 Git common-dir 身份聚合，原生 Codex/Pi/Claude 有界路径/会话元数据采集；独立元数据 cwd 只核固定项目标记，未登记也可发现，不扫描父级/磁盘或读源码内容。其他客户端逐个 unknown。
- 目录观察、聚合项目、历史客户端关联、创建时间、Orca 活动和 Git 提交分别保存；创建客户端不冒充贡献者，会话创建不冒充最新活动。保留来源/真实错误及样本上限，移除每项目通用证明式限制文字。
- 独立 SQLite 人工 notes/review/group/intent/archived，revision CAS 和服务时间，拒绝 Agent/匿名/伪时间写入；反复刷新保留字段，可逆归档，无代码删除或 grant/native control 变化。缓存 GET 不扫描、不写入。
- W0 单一拥有契约、迁移012、HTTP保护、生成客户端；W3单一拥有前端。跨域问题均通过 Orca Handoff，由原owner返修。

业务 SOURCE：`be3cbfdbf536053f196ae5c48afad5fbf48138c2`（包含 `fc49bfa` 采样/来源计数和 `2ef7328` 原生实际输入长度指标）。最终验证且已push的 checkout SOURCE：`4201c899f5aa0285a0d19442b674327237967739`，普通合入 W0 `a681d54` / W3 `dfc4db8`；REPORT 是随后独立的报告提交，不改变业务代码。

## 验证与首失败

| 命令/实际行为 | 结果 | 原始记录（`%TEMP%`） |
|---|---|---|
| `go test ./internal/workspace/... ./internal/adapters/agents ./internal/adapters/sqlite` | 首领域/适配/SQLite场 PASS | `astrocyte-w2-paper-first-tests.log` |
| `go test ... -run 'TestNativeMetadata\|TestClaude\|TestProjectMetadata\|TestProjectBoard\|TestSelectedText' -count=1 -v` | 最新8个重点回归 PASS | `astrocyte-w2-paper-final-sampling-focused.log` |
| 初始 Claude 元数据测试 | RED：Windows 不接受 `/approved` 为 absolute cwd；修正为真实临时根后 PASS，原日志保留 | `astrocyte-w2-paper-final-focused.log` / `astrocyte-w2-paper-final-focused-retest.log` |
| SQLite人工字段/刷新/Agent拒绝/CAS/独立子进程冷打开/反归档 | PASS；子进程另开同库，模拟源测试不冒充真实项目采集 | `astrocyte-w2-paper-restart.log` |
| `$env:ASTROCYTE_TEST_REGISTERED_BOARD='1'; go test ./internal/adapters/sqlite -run '^TestActualRegisteredBoardSQLite$' -count=1 -v` 首已登记场 | PASS 51.08秒：54目录/14组/26 Codex历史关联，16根/Git限制 | `astrocyte-w2-paper-actual-board.log` |
| 同上，补入实际 Claude 元数据 | PASS 18.11秒：54目录/14组/27关联，Claude真实匹配1条；不是全量客户端覆盖 | `astrocyte-w2-paper-actual-claude-board.log` |
| 同上，扩展独立原生根后的首场 | PASS 171.90秒：73目录/33组/19原生独立根，Codex256条上限提前停止，原限制保留 | `astrocyte-w2-paper-actual-standalone-board.log` |
| 同上，按根/客户端4样本继续扫描后 | PASS 37.86秒：111目录/71组/57原生独立根，4多目录组、5实际多客户端组，125保留关联样本；第二次真实刷新保留备注/复盘/归档，Agent被拒 | `astrocyte-w2-paper-actual-sampled-board.log` |
| 本轨首完整 `pnpm check` / `pnpm build` | PASS：237契约例、14合成输入真实API例、58前端例；真实付费/媒体关闭 | `astrocyte-w2-paper-check.log` / `astrocyte-w2-paper-build.log` |
| 新 source enum 后首次 `pnpm build` | RED：`RegisteredProjectsPanel.tsx:44` TS7053；已 Handoff W3，原owner修复 `dfc4db88771e82d8cd999f9a097d62939852071e`，普通合入 | `astrocyte-w2-paper-final-build.log` |
| 最终 `pnpm check` | PASS/exit0：Go test/vet/mod、19包架构、进程退出回归、237契约例/生成一致、14合成输入真实API例、61前端例、TS/lint/diff | `astrocyte-w2-paper-final-check.log` |
| 最终 `pnpm build` 返修复验 | PASS/exit0：Go可执行程序、TypeScript、Vite92模块；未覆盖前一场RED日志 | `astrocyte-w2-paper-final-build-retest.log` |
| `$env:ASTROCYTE_TEST_PROJECT_BOARD='1'; node --test tests/s1/project-board.test.mjs`，最终SOURCE真实API/契约/SQLite/自有进程冷启动 | 1 PASS / 44.87秒 / exit0：111目录、71组；人工字段/重复刷新/403/CSRF/CAS、无grant改变、旧cookie冷启动失效、新会话读回同备注/归档/SQLite原数据；无模型/媒体 | `astrocyte-w2-paper-actual-api-restart.log`；保留库 `astrocyte-s1-meg3Pi/data/state.sqlite` |

重点边界：接受 >128KiB 与准确512KiB最终信封，拒绝多一个UTF-8字节及多字节溢出，并在空Registry中核对拒绝早于原生选择/启动。采集回归验证只读元数据、未匹配私密内容不保留、16记录/64KiB、独立标记根、重复身份、后续不同项目不被单根历史挤掉。没有为采集重发旧模型、媒体、克隆或未知作业。

实际采样场的 Codex 检查1801目录项、保留119条样本；Claude检查21目录项、保留6条样本，Pi源 unknown。其日志旧 `matched` 是该场保留样本；最终字段已拆分匹配元数据与保留样本，不能把119/6解释为总会话或当前活动。来源一直 partial；消失/无Git/重定向/非项目/样本和深度限制保留。

最终实际API场的来源统计：Codex `entries_examined=1802`、`headers_examined=1714`、`matched_headers=896`、`matched_roots=63`、`retained_associations=119`；Claude相应为21/7/6/6/6。只在这些已知来源/本次上限内匹配，不能把896当机器全量会话。实际API服务已由既有测试helper结束，原库和原日志保留；没有个人预览更新。

## 剩余范围与交接

- 进度人工阶段或Agent推断为 ASKED PENDING，没有模型推断、自动读取TASK/STATUS或新增特权接口；intent 是自由人工文字。
- 全部本地客户端覆盖仍受来源限制：Pi known-source不存在，另外7种客户端暂无可靠元数据适配；本轮真实多客户端仅 Codex/Claude。
- 目录/历史关联与固定标记不表示研究成果、活跃进程、完成率、原生可用性或已配置。未知Git提交/工作状态维持未知，独立克隆不凭同名/同HEAD合并。
- root已明确“OSrestart”指自有 API 进程终止后用同SQLite/对象目录启动新进程并真实API/浏览器查询，不需要重启Windows。本轨已用最终SOURCE跑过真实API冷启动；W3浏览器使用指定18787/15173端口，由其唯一owner执行并向root报告。本轨未使用它的端口或伪装浏览器已通过。W0首次旧SOURCE `88b2bcc` 的1 PASS/31.88秒另保留，不合算最终场。
- 新的完整原论文+视频联合模型调用由root授权后仅W0执行一次；W2没有模型调用。不重发旧UNKNOWN，不能用512KiB合成边界替代新增完整模型结果。
