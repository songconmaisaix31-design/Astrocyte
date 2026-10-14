# Relay 模块管理与验收

业务事实源为根 AGENTS.md、HANDOFF.md、STATUS.md、docs/SPEC.md、docs/TASKS.md、docs/QUESTIONS.md 和 [发布任务](../../tasks/Release-v0.1-20261014.md)。2026-10-14 用户已确认三层组织；本模块负责人只修改本文件，业务实现、测试、文档和返修由原 R/H 开发人员持续完成。

## 身份与分工

- 总负责人：`term_494acbb9-67a5-4360-82dc-8514ce124990`，顶层 Run `run_973808dd6a2f`。
- 模块负责人：`term_3413269b-9c86-4a2d-804b-790396bf130a`，Task `task_e8c1d01d67ec` / Dispatch `ctx_e6328a370d56`。
- 模块子 Run：`run_8bbda1c3974c`，2026-10-14T12:51:14Z 创建，未绑定或夺取总负责人 Run。
- 模块工作树：`module-relay-20261014`；分支 `songconmaisaix31-design/module-relay-20261014`；基线 `dd60de11bca85662f5bb4a5b281cc7b42e6ffd2f`。

| 开发人员 | 固定工作树 / 分支 | 独占写域 | 调度状态 |
|---|---|---|---|
| R | `works-relay-runtime` / `songconmaisaix31-design/works-relay-runtime` | `internal/workspace/{domain,app}/relay*.go`、`internal/adapters/{sqlite,agents}/relay*.go`、`internal/adapters/relayworktree/**`、`agents/native.go`、`agents/native_process.go`、`docs/access/runtime.md` | 原阶段 completed/succeeded 且总负责人接收并精确释放旧终端；后续 `task_b833b3a06c1e/ctx_216bf731aaf8` 已在本模块子 Run 实际启动，ready / turnStart observed |
| H | `works-relay-http` / `songconmaisaix31-design/works-relay-http` | `internal/adapters/events/relay*.go`、`cmd/server/relay*.go`、`internal/adapters/httpapi/relay*.go`、`internal/workbench/app/relay*.go`、`internal/adapters/sqlite/access_session*.go`、`contracts/openapi.yaml` 与 relay examples、`docs/access/http.md`；必要 `cmd/server/main.go` / `internal/adapters/httpapi/server.go` 组装 | `task_cfb7c9bde07d/ctx_312d983772a9` 沿原 Run 活跃；已发直接 Handoff，要求同时向模块 Dispatch 报告 |

依赖单归 B，H 不修改 `cmd/server/config.go`。F `ctx_5bac511ada08` 是当前实际产品页面唯一操作人；B `ctx_7217777c7bf5` 接收真实公开来源、生产入口及普通库存测试迁移 Handoff。产品模块负责人 `ctx_b1cd2e0a39aa`、集成发布模块负责人 `ctx_376d222c774b` 已由总负责人确认，跨模块依赖直接 Handoff 其精确 Dispatch。

## 连续执行计划

1. R 修复 CreateIntent 的项目权限和版本绑定：经既有领域端口读取实际项目 settings，保存真实 `settings.revision` 与 `permissionlevel`。确认必须再次验证当前权限和版本；保留前端比较。真实 SQLite/HTTP 正常确认与之后权限变化阻断均须验证。
2. R 普通 push 阶段 SOURCE 和独立 REPORT，将精确 SHA 交 H；H 普通消费并核实自身实际服务组合。R 另交 B 实际 `actualpublicsources.json` 本机指针与生产入口，作品进度 infer 使用实际公开 README/手册及既有领域端口。
3. F 在同一 URL 建立无 Codex 登记基线并启动 DOM 观察；H 仅收到 ready 后真实注册、重启并记录原始开始时间。F 经产品 HTTP 执行 confirm/send/read/stop/release/409；H 独立核控制器自产文件、真实 SQL 与每个 event_id。
4. H 交精确 SOURCE/REPORT、服务配置参数、安全聚合与本机证据指针；模块负责人逐项核验并向集成发布负责人交接。阶段通过后继续支持组合、Windows x64 包与返修，最终发布支持未完成时模块 Dispatch 保持活跃。

只有总负责人明确接收旧 worker_done 并交接已核实终端，才在子 Run 复用同一开发人员、终端、工作树和分支派下一任务。不得中断活跃人员、重复创建同写域写入者或向 completed Dispatch 投递新开发任务。

## 当前证据与第一失败

- R 已接收阶段 SOURCE `fbfaa65c2c01094305424ac40d2d39d271eab51b` / REPORT `53231b19ddd19a65c5a47c91469a0e34d73cf600`；本机只读核实 HEAD 为 REPORT、工作树 clean、原分支正确。实际 Codex/SQLite 生命周期结果由 R 报告，HTTP/DOM/三分钟尚未据此通过。
- H 用户派发输入 SOURCE 为 `99a250d0b9069ffb32205420f63f82b372de5607`；2026-10-14T12:52Z 本机读取其当前 HEAD 为 `7dd7d1b91128a268c035fcfa49da5f9f8667078d`，`docs/access/http.md` 有活跃未提交修改。当前文件属于 H，模块负责人未修改或接纳为最终报告。
- 实际 API `http://127.0.0.1:17342`、F 页面 `http://127.0.0.1:17343`；H 数据根 `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/h-production-current-20261014`，安全中间结果位于 H 忽略目录 `dist/l1-live/current-20261014`。
- 已知产品首 RED：CreateIntent 保存的 revision 为 0，UI 显示 changed 并阻止确认，等待 R 修复。旧失败证据不覆盖。
- R 返修首次 `worker-start --spec ... --worktree path:C:/Users/DW/orca/workspaces/Astrocyte/works-relay-runtime --terminal term_e0e48c51-2379-4c5e-b0e0-9206b29fcffe --run run_8bbda1c3974c` 返回 `agent_unconfigured`；request `889a4d41-f37f-43f5-8204-82003c437fd7`。子 Run 的公开 worker-list 和 task-list 均为空，未生成新 Task/Dispatch。
- 随后公开 `worker-show ctx_b4fcb141ca00` 返回 completed/succeeded、`projection.liveness=live`、`agentIdentity=codex`、`exactWorker=true`；terminal read 显示 GPT-6.1-Sol high fast 空闲提示。CLI 启动识别与现有进程证据不一致，已向总负责人 escalation 并追加状态，未重试、关闭、释放或另派写入者。
- 总负责人 2026-10-14T12:54:17Z 报告已对旧 R 精确 `worker-release`：`released/processAction=closed_agent_terminal`、transcript captured；公开 worker-read 同样返回 resource released / liveness exited，原工作树 terminal list 为空。随后已授权在原工作树/分支启动唯一后续 Codex 开发人员，保留全部提交。
- 按该授权 `worker-start --spec ... --agent codex --run run_8bbda1c3974c` 返回 `nested_worker_depth_exceeded`：深度 2、当前最大深度 1，`effectsApplied=false`，request `a1c9069c-8e54-454e-a175-64ffcfb48e73`。尚无新 Task/Dispatch/Agent；已向总负责人报告需在 Orca Settings → Orchestration 将 Nested worker depth 设置为至少 2，不能用新 Run/低级 dispatch 绕过，也不能由模块负责人代写领域代码。
- 总负责人 2026-10-14T12:57:50Z 报告通过官方 Settings 的纯文本 accessibility 将 nested depth 改为 2 并读回，未改内部 DB/源码。配置状态改变后正常 `worker-start` 成功：`task_b833b3a06c1e/ctx_216bf731aaf8`、终端 `term_1394f6f0-7f05-486f-b0ee-246a193105a4`，`ready/input_accepted/turnStart=observed`，prompt observation supported，process incarnation `ff19643c-de7a-44ad-b85a-7f69709b68bb`，request `1fbe74a4-7c89-4bc8-9c9f-c4d2ab24630a`。原工作树复用、默认 Codex 配置继承，未指定未经核实模型；第一失败仍保留。
- H/F 已收到精确 Dispatch Handoff，登记 durable enqueue；尚未把 enqueue 解释为实际开始或接收。
- F 2026-10-14T12:53:45Z 报告在同 URL 实际两 viewport session200 与空 registry 基线，观察器打印 `F_BROWSER_READY`。其初次 registry 403 竞态原件 `.verification/l1-first-visible-current.log` 保留。H 12:54:45Z 确认注册未开始、模型派发为零、计时未启动，待 R 精确 SOURCE 后正常消费/编译/重启并协调 F 观察窗口。
- 已向 B 提供现存实际来源文件本机指针 `C:/Users/DW/orca/workspaces/Astrocyte/works-relay-runtime/dist/relay-l1/actualpublicsources.json`，只读文件列表确认存在，未读取或输出源内容；生产入口与进度上下文要求由 R/H 持续交接。
- R 2026-10-14T12:59:57Z 报告实际 started：原 tree/branch clean、HEAD 为原 REPORT，已读取规则和当前任务；不启动额外 CLI/模型，沿既有权限领域端口返修。公开 worker-list 进一步给出当前 `ctx_216bf731aaf8` 为 active/live/working、实际 provider.model `gpt-6.1-sol`；该状态仅证明开发任务活跃。
- H 阶段 REPORT `fad2b48a9ff7ed0d8d5d6c202caf8efe9076e880` 已本地提交，仅 `docs/access/http.md`，SOURCE `7dd7d1b91128a268c035fcfa49da5f9f8667078d` 为普通祖先。普通 push 因 `SSL_ERROR_SYSCALL` 未成功；模块独立只读 ls-remote 同样 TLS 失败，已交总负责人共享网络阻塞，不声称 REPORT 已 push。
- B 报告自身生产 HTTP/Service 对公开 README 的一次真实 progress infer 已成功，并核 SQLite、引用 mtime/size 与新 API 进程；该路径与 H 同一受控项目的组合验收分列，成功调用不重复。普通库存生命周期迁移所需实际 NativeRegistry 输入和工具关闭入口已交 R 继续提供。
- R 阶段 SOURCE `ce3a519efa5317e642d7488a776b850eebca1093` 已普通 push，模块独立 ls-remote 同 SHA、diff 写域核对 PASS。生产逻辑保存真实项目 revision / permissionlevel，Confirm 在原意图与当前权限不同时返回 ApprovalRevoked，保留请求版本的 VersionConflict 校验；不覆写原意图。H 已收到精确源码消费和复启 Handoff，真实成功确认与权限变化拒绝仍待执行。
- 独立源码审查发现该阶段新 `relay_intent_test.go` 使用原 contract fixture 的模拟 driver 成功和项目手改，已退原 R 在同写域普通追加真实适配器 / SQLite 验证；保留阶段 Git 历史，不接纳替身成功为当前验收。最终 SOURCE / REPORT 等待该返修。
- 总负责人已将共享网络排查交集成发布模块。其只读诊断为 TCP443 可达、TLS 间歇失败；现有系统代理为 `127.0.0.1:7890`，仅命令级 proxy/schannel/HTTP1.1 曾读成功，后续 push 仍可能失败。已交 R/H 有界普通 push 与独立远端核查，禁止全局 TLS/代理修改；一次成功不登记网络永久恢复。
- R 追加真实测试 SOURCE `9719e6970927caee3e0b08d4020708e383881880` / 独立 REPORT `392bd81e822148cb2a7cb8f7bd7a83699d0f7d44` 已普通 push，模块独立核远端 REPORT 与 SOURCE 普通祖先。相对 ce3a519 只替换新测试，生产逻辑全等；模块独立 `go test -mod=readonly ./internal/workspace/app -run '^TestRelayIntentBindsSQLiteSettingsAcrossReload$' -count=1 -v -timeout=90s` PASS，A/B/C 共 0.96s / package 2.756s，真实 production adapters / Git / SQLite / HTTP 权限版本保存重载及旧意图 403 拒绝，零模型派发。
- H 已普通 no-ff 合 R ce3a519，merge `afb4993e3316d946ec2727686fdc397a2fc1fd82`，追加真实 assistant / task_complete / tool0 / nullable token 检查 SOURCE `9b0df7f2251cd16f0176035c8ab31bb07b0bca4e`。模块独立远端 SHA 一致，旧 H REPORT fad2b48 为普通祖先。H 报告 API17342 正常 shutdown/recompile/restart 为本次 owned PID12552；F 新 ready 前 registry/model/clock 仍为零，不据源码合入推断 DOM 通过。
- 产品模块发现 R `native.go` 原生 response decode 可能忽略 json.Unmarshal 错误，已交原 R 同一活跃 Dispatch 核查 fast-fail。R 13:10:55Z 确认尚未 worker_done，继续原任务纯解析错误检查与普通追加 SOURCE/REPORT，不重发实际任务。
- B 普通 Reader 新 API 预告为 `NewWorkbenchReader(root string, protectedSources ...string)`，默认不 stat 私人正式源，显式保护源保留同文件/alias 拒绝与 ro/immutable/query_only。`workbench_config.go` 旧常量衔接已请求总负责人明确 H 写域或交唯一集成人员；接口未有精确 SOURCE 前不消费 WIP。H 当前服务 Progress 未配置，已交产品模块/B 给现有生产构造口，只由 H 自身 relay 验证组装文件补必要导入，不代写进度领域。

模块管理阶段 SOURCE `e404da04f7b10296bed8133589c292766c5398cf` 已普通 push，`git diff --check` 和 `git ls-remote` 精确核对 PASS；该结果仅包含本模块管理文档与派发核查，业务修复和产品验收仍按下面记录进行。

## 验收记录

| 核心路径 | 当前结果 | 必需证据与负责人员 |
|---|---|---|
| CreateIntent 保存权限 revision；正常确认；权限变化阻断 | 首 RED 保留；9719e69 实际 SQLite/HTTP 保存与拒绝独立 PASS，真实产品成功确认等待 H/F | H 实际 SQLite/HTTP，F 实际页面，保留前端比较 |
| 控制器 start/send/read/stop/resume、持久身份与三值状态 | R 阶段报告 PASS，组合未验 | 当前真实 controller 与 SQLite、明确活动窗口语义；H/F 产品链路 |
| 控制器自产会话逐 event_id 定位、六字段 HTTP、分页与 GET 只读 | 进行中 | H 独立原档 / SQL / EventRepository / HTTP 比较 |
| 来源覆盖、追加、停止和恢复边界 | PARTIAL / UNKNOWN 保留 | `ArchiveAppendOnly=unknown`，文件 PrefixComplete / SessionComplete 不足，恢复 raw gap 独立记录 |
| 同 URL 未登记基线、真实登记、首次 Today DOM 三分钟 | NOT_RUN / 等待 F-H 协作 | 开始与 firstDOM 原始时间，不能事后重记 |
| 注册表扩展、中性 worktree 换手、terminal-only、互斥与权限拒绝 | 最终组合未验 | 对照 v1.2 与当前 release 任务；每项真实结果独立报告 |
| 默认检查、构建、完整浏览器、Windows 解压启动读写及重启持久化 | 模块未验 | 各原领域人员、唯一集成人员和发布包人员精确组合 SOURCE |
| Procmon、两次真实 00:30、用户自由工具与实际作品 | NOT_RUN / 未决定 | 不降低门禁、不扩大授权 |

## 发布交接约束

原始源、凭据、模型输出、会话文件和正式数据不进入 HTTP/Git/release 包/stdout；只传安全派生和本机证据指针。旧正式 V4、锁定协议、认知 schema、engine、mapper、七类权重、日程和既有 17321 服务保持冻结。每阶段 SOURCE 与独立 REPORT 普通提交并 push，保留第一 RED/UNKNOWN/NOT_RUN；不 force、rebase、reset 或 squash。最终 GitHub Release 由总负责人接收具体可审查结果后发布，本模块不发布。
