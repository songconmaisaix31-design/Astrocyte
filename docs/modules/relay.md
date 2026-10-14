# Relay 模块管理与验收

业务事实源为根 AGENTS.md、HANDOFF.md、STATUS.md、docs/SPEC.md、docs/TASKS.md、docs/QUESTIONS.md 和 [发布任务](../../tasks/Release-v0.1-20261014.md)。2026-10-14 用户已确认三层组织；本模块负责人只修改本文件，业务实现、测试、文档和返修由原 R/H 开发人员持续完成。

## 身份与分工

- 总负责人：`term_494acbb9-67a5-4360-82dc-8514ce124990`，顶层 Run `run_973808dd6a2f`。
- 模块负责人：`term_3413269b-9c86-4a2d-804b-790396bf130a`，Task `task_e8c1d01d67ec` / Dispatch `ctx_e6328a370d56`。
- 模块子 Run：`run_8bbda1c3974c`，2026-10-14T12:51:14Z 创建，未绑定或夺取总负责人 Run。
- 模块工作树：`module-relay-20261014`；分支 `songconmaisaix31-design/module-relay-20261014`；基线 `dd60de11bca85662f5bb4a5b281cc7b42e6ffd2f`。

| 开发人员 | 固定工作树 / 分支 | 独占写域 | 调度状态 |
|---|---|---|---|
| R | `works-relay-runtime` / `songconmaisaix31-design/works-relay-runtime` | `internal/workspace/{domain,app}/relay*.go`、`internal/adapters/{sqlite,agents}/relay*.go`、`internal/adapters/relayworktree/**`、`agents/native.go`、`agents/native_process.go`、`docs/access/runtime.md` | 当前 `task_d480ceaf0e15/ctx_1f88d9e6f4f4`，复用 `term_6ffbc768-d919-4bcf-bb28-d43e4e250e42`；actual started 并持续实现，公开观察 live / working / exactWorker；首 start_unknown 保留 |
| H | `works-relay-http` / `songconmaisaix31-design/works-relay-http` | `internal/adapters/events/relay*.go`、`cmd/server/relay*.go`、`internal/adapters/httpapi/relay*.go`、`internal/workbench/app/relay*.go`、`internal/adapters/sqlite/access_session*.go`、`contracts/openapi.yaml` 与 relay examples、`docs/access/http.md`；必要 `cmd/server/main.go` / `internal/adapters/httpapi/server.go` 组装 | `task_cfb7c9bde07d/ctx_312d983772a9` 沿原 Run 活跃；已发直接 Handoff，要求同时向模块 Dispatch 报告 |

依赖单归 B，H 不修改 `cmd/server/config.go`。总负责人 2026-10-14T13:13:29Z 明确增加 H 的 `cmd/server/workbench_config.go` 单文件薄接口写域，只消费 B 已交精确 Reader 接口，默认无私人正式源路径、保留只读和路径保护；Reader / contracts / config.go 均仍由 B 独占。F `ctx_5bac511ada08` 是当前实际产品页面唯一操作人；B `ctx_7217777c7bf5` 接收真实公开来源、生产入口及普通库存测试迁移 Handoff。产品模块负责人 `ctx_b1cd2e0a39aa`、集成发布模块负责人 `ctx_376d222c774b` 已由总负责人确认，跨模块依赖直接 Handoff 其精确 Dispatch。

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
- H 已通过自己 `relay_session_live_test.go` 一行复用生产 `NewProjectProgressService(db, db, agents.ProjectFiles{}, registry, registry)`；本地 SOURCE `12be99f`，后续普通合 R9719 的 H HEAD 为 `eb13789f830949e016d9124ed5113754418c0e22`。H 独立实际 A/B/C SQLite/HTTP 拒绝测试 PASS 10.95s，当前无注册/模型/计时，等待 F 新 session / DOM ready；具体远端最新核验仍由其继续报告。
- R Native fast-fail 追加 SOURCE `d6d4da6ac00a196e6bbcf592cfdc0a2a2731dbe6` / 独立 REPORT `07449a65ea3bd2465d207cd8771aceda6fe50a9e` 已普通 push，模块独立远端 REPORT、SOURCE 普通祖先与 clean 核对 PASS。模块独立 `go test -mod=readonly ./internal/adapters/agents -run '^TestRelayNativeDecodeRejectsMalformedValuesWithoutRawLeak$' -count=1 -v -timeout=90s` PASS，9 个纯 JSON 错误子例，包 1.333s；第一选择器无匹配为 NOT_RUN，正确名称实际运行另列。错误保持安全 delivery_unknown，不创建控制器或成功协议替身，H 已收到首个真实注册前串行消费要求。
- 子 Run 2026-10-14T13:19:18Z 收到 R 本阶段合法 `worker_done=succeeded`，模块按精确 SHA / 写域 / 实际独立检查接收；`worker-retain` 返回 retained / user_requested / processAction none，固定原终端等待发布返修。该阶段结果已交总负责人接受和后续终端交接，模块当前 Dispatch 继续活跃，H/F 与最终包真实验收未完成。
- 默认 `relay-runtime` 浏览器场景除三分钟 flow 外仍需要每尺寸的真实 EOF、互斥/重复/通用冲突、working、unknown send、complete/release、alternate 和 event publication 条件；已交 F/H 逐项真实场景与缺口，不制造 unknown / terminalMode / completed、不增加 skip 或删除成功断言。最终 ZIP 必须对实际 packaged server 独立端口执行原生正向与 event_id / SQL / HTTP 核查；H 已获要求将安全启动参数和本机指针交集成发布模块/P，包不含原 DB/会话/凭据。
- 总负责人 13:21:27Z 明确由模块自主接受子任务和决定同终端后续，不再等待最高层重复接受；暂无具体工作时释放合法结算 owned terminal，不能只因可能发布返修 retained。模块已给 R 具体默认生命周期场景支持任务。原 completed 终端再次复用返回 agent_unconfigured、request `7f0d9995-ccfd-4d9d-970f-aee96e2abc48`，未新生任务；精确 `worker-release ctx_216bf731aaf8` 返回 closed_agent_terminal / transcript captured，原工作树 terminal list 为空。
- 随后支持 Task `task_6586f574ed0b` 首 Dispatch `ctx_2125bf78ac50` 为 turn_start_unobserved / outcome_unknown；公开 terminal read 明确 account/read routing discovery timeout 后返回 PowerShell。模块据实际退出证据精确 worker-stop / release，原始 startup UNKNOWN 保留，不据 missing_status 推断退出。以同 Task / `--retry-of ctx_2125bf78ac50` / 原 worktree 和 branch 恢复，当前 `ctx_1210a540a63b` / `term_6ffbc768-d919-4bcf-bb28-d43e4e250e42` 为 ready / turnStart observed，13:26:51Z 原 owner actual started；无重叠写入者、第四层或 H/F 任务代执行。
- F 13:20:00Z 新 observer session23343 实际两 viewport auth200 / 空 registry ready。H 完成串行部署 R d6d4da6 / REPORT07449 后真实 agents.json 仅一次追加，固定 `registered_at=2026-10-14T13:23:34.0366105Z`，首三分钟截止 `13:26:34.0366105Z`；准确固定公开 task 已直接给唯一 F。模块尚待原 firstDOM / accepted 时间结果，不重记起点、不据消息缺失判模型失败或完成。
- 工具关闭验收必须核本次实际 owned PID / 启动禁用配置及 controller 自产 turn_context 或对应 source 安全聚合；tool_calls=0 只证明未调用，不能单独证明禁用。已交 H/R 核本次正选来源，不能重发旧成功补证或读私人历史/凭据。
- 首注册三分钟实际 FAIL：H 13:27:05.9397126Z 只读 GET tasks=0 / dispatched=0，F 明确未 intent/confirm/模型，原 observer 已超时且后续 registry 非空。起点与截止均不修改。固定输入交接最初只有发送回执、没有 F 读取确认；H 后来不附发送者 lifecycle 字段重新发 `msg_67709f2c6dfc`，模块公开 worker transcript 核准确既定正文后原样转 F 当前 Dispatch，回执 `msg_bf915b5dafbc`。F 13:35:42Z 明确实际收到并 ACK，交付异常已解除，不假定旧消息误发。
- F 首次迟到实际 UI task `0c6522bc-b232-4040-8407-745885644791`，人工 A/rev2 确认一次，operation `906dd04e-6956-4539-af2b-694d33f286c0`，13:33:19.092Z HTTP503。模块独立只读 HTTP 收据为 accepted + provider_unavailable / inspect_local_relay_provider，实际 task working、terminal_exited / exited、event publication published；503 不解释为未投递，不重发原输入。
- H 后续取得原 controller 正选 archive，模块 read-only/query_only SQLite 独立核 source_status available、gap null、has_archive_path / has_native_id true、archive_finalized unknown。H 对此原来源执行实际 `TestRelaySessionActualCommittedEvidence` PASS 0.14s：15 个完整行 / 105564 bytes、15 projected、unknown5 / errors0，逐 origin / SQL / GetEvent ID、HTTP 六字段/分页/四403/GET静默计数/原 canary 排除/持久游标与当前文件一致均通过；自产 assistant1 / task_complete1 / model gpt-6.1-sol / tokens8827 / tool0。current snapshot complete 与 session complete false 分列，不覆盖来源缺口或首503。
- 模块独立 HTTP session-events 对原 task 返回18事件，字段集合精确六项，其中17 tokens null；只读 SQLite 的 attention_engine_events 逐18 ID 均存在，missing0。首 SQLite CLI 多行 JSON 收集错误导致比较未执行，已修正收集方式实际重验，不当作事件缺失。真实 archive coverage 为 lines15/projected15/unknown5/errors0/snapshot_complete true/session_complete false，gap unknown_records；已给 B/H 真实 nullable 样本本机指针范围，不插假数据或新增模型凑样本。
- H 原本机白名单 WARN 保留 method thread/read、protocol_code -32603、reason thread_rejected、实际 Shanghai21:33:19；已给原 R 继续处理首错误安全保留，不猜内部拒绝原因。原调用模型成功与首 metadata 503 分列；旧 owned PID / argv 未保存，工具有效集合证据不足，不由 tool0 或新配置反填。
- R 新窗口支持 SOURCE `6ed9ae1301dc2e91f0d17a335300697fae3e253f` 已普通 push 且模块远端核验：显式 SessionDeadlineSeconds（零仍180秒）、真实已装 Claude 固定 stdin_eof argv / 单公开根 LaunchPolicy，以及原 owned session 的本机 PID / deadline / argv 策略安全证据口。后续 `055a57a18c732ffbbc55d49c605f65c655b8edea` 核通知模式真实 argv 参数，`5c37c261ae1f98fbdab6926415712ef4a5181581` 移除新增 Skip 并保留缺入口拒绝验证；R 继续同域首 RPC 错误安全摘要。模块实际三个门禁测试 PASS（package8.190s），后接 ls-remote TLS失败分列，不声明最新远端核验完成。
- H 原 Dispatch 的跨 Run `ask --to run:run_8bbda1c3974c` 实际被所属 Run 校验拒绝；模块 reply 原问题也返回 consumer_fenced。日常领域问题通过普通 Handoff 直接给模块，必要的正式 ask 留原 owning Run，由总负责人回复。原问题 `msg_f4c73bd8170c` 已由总负责人合法回复并获 H 实际确认；worker_done 仍留原 owning Run。模块不夺取绑定、不读取 Orca 内部 DB。
- R 默认场景阶段 SOURCE `87a54b762aef0c6a175eab1470508e76aa6ace45` / REPORT `63b55d770a00724795196417af2ef4f3ba922ed0` 已由 R 普通 push 并核远端，13:55:21Z worker_done 接收。模块独立四个真实拒绝及首错误安全保留测试 PASS，package 1.215s；最新独立远端读取 TLS 失败另列。该阶段保留默认180秒，显式可信组装可配置900秒；既有 EOF helper 只接受准确公开根、实际已装 Claude 固定 argv 和关闭工具策略，effective tool set 仍需实际窗口证据。
- 当前 R Task 初次启动 turn_start_unobserved，公开文本确认任务停留在 composer 后只提交原输入一次；14:03:56Z R actual started。14:31Z 公开 worker-show 为 owned / live / working / exactWorker；首次 startup UNKNOWN 不改为首次成功，无重复人员或输入。
- R 可移植内部证据口 SOURCE `0de4e464e3dab342a06fdb37d76ea58a8d5f1f32` / 独立 REPORT `d8cb13a8b1a392a2105bba9e826daa1364e70caa` 只给 B 原样消费 `internal/adapters/agents/relay_execution.go`，R metadata 保留独立文件。B SOURCE `3384a89aac3be6d5670d756eb2d44263784d0cb8` 已普通 push，B 和产品模块双 ACK 内容逐行大小写全等、无其他 R 文件消费；14:31Z 模块正式归还该文件写域给 R。B 后续只修改自身 text / bridge，对缺失 metadata 保持 unknown，不以 NativeID 重建原 process 所有权。
- 总负责人正式 reply `msg_0c1795a84006` 回答 `msg_ac65fadbbe5e`：纯人工 done / released Transition 核现有 observe 许可；保留人工身份、原项目 revision / A-B-C / 对象所有权、grant 撤销和失效、completion_confirmed、真实退出与 pending 拒绝。真正事件 publication reconciliation 继续核 reconcile。R 正在原写域实现；旧 task revision2、原输入、收据与开始时间不修改，F 是唯一实际 UI 操作人。
- R 新权限真实 SQLite / HTTP A-B-C、vet / build / architecture 已报告 PASS；全 Go 首 RED 为 H `TestRelayPersistentRealHTTPReceiptsMutexAndRestart` 成功 reconcile 场景未预先授权 reconcile，实际403 approval_required。首日志 `dist/relay-l1/human-transition-full-go-test.log` 保留，已交 H 原 owner 补真实权限前置和缺许可403；无新增 skip 或删除成功断言。修复 SOURCE 与完整组合通过仍待实际核验。
- R 人工权限 SOURCE `d14b08dc50014baf9ee636543564b040112cf3b5` / 独立 REPORT `fae304b80cb3288b9edd08f10500c6e5af7e4210` 已普通 push，模块独立远端与普通祖先核验 PASS，REPORT 仅修改 `docs/access/runtime.md`。源码修改只在 R 三个 relay 文件；模块独立 `go test -mod=readonly ./internal/workspace/app -run '^TestRelayHumanTransitionProductionPermissionGates$' -count=1 -v -timeout=90s` PASS，真实 A/B/C Git / production SQLite / HTTP 拒绝检查1.13秒、package3.073秒，零模型。该检查不制造 working / exit，也不替代旧 task 实际成功释放；H 已收到串行部署与 F 唯一操作交接。
- H Foundation / Projects 实际组装 SOURCE `3016a5369393166ac3a40f520138ae352d2edf92` / REPORT `82d014dd78ac732b86e1a562f633909c75d4941e` 已普通 push；显式受保护来源 guard SOURCE `24b6390bcad60a2a20b671dcb60725e77eb3a7a4` 已 push。薄接口等待 B 精确窄 Reader 提交或唯一集成人员实际组装；不能合 B 含冻结历史的整个分支进入 H，不能由 H 修改 Reader。
- F 首 native DOM 实际为13:33:25.152Z / .153Z；H 逐 SQL 核两个自产 archive ID，第三个 relay_task 事件不计 native。原注册至 DOM591116 / 591117毫秒 FAIL；dispatch 创建至 DOM14.516 / 14.517秒 PASS，两种起点分列。原 task 单次 read 返回200 / accepted，controller 已实际退出，人工 done / released 等上述权限修复投入和 F 实际操作。
- 总负责人已授权先合法释放旧 task、SQL 工作区解锁和 pending0，再创建独立明确公开 scope / DB / follow-up；新窗口使用既有显式900秒并提前取得 F observer ready。新模型工具关闭须有本次实际 PID / argv / controller 来源证据；旧第一 FAIL、503 accepted 和 unknown 不覆盖。PUBLICATION_ERROR 尚未执行；H 正核新独占 SQLite 的真实 writer 争用是否能只使事件 sink busy，不能假造 sink、增加测试开关或使用旧 scope / 正式数据。
- 14:37:50Z H 从原 relay.sqlite 精确 binding 选定本次 controller 原档，实际 regular / bytes105564 / nlink2 / source_status available。Windows 来源身份 guard 要求单链接，当前 `source_identity_unverifiable_or_aliased`、coverage errors1 / snapshot_complete false，当前覆盖 FAIL；此前15行 / 18个 HTTP ID 通过保留为当时结果。H 声明原档未建链接，其保护测试只链接独立新 SQLite。模块已交 R、产品 F/B、发布 Q/N/P 核准确本次消费者和自有链接创建清理证据，禁止扫描私人历史、删除未知第二链接、放宽 guard 或重复旧成功模型填证据。

模块管理阶段 SOURCE `e404da04f7b10296bed8133589c292766c5398cf` 已普通 push，`git diff --check` 和 `git ls-remote` 精确核对 PASS；该结果仅包含本模块管理文档与派发核查，业务修复和产品验收仍按下面记录进行。

## 验收记录

| 核心路径 | 当前结果 | 必需证据与负责人员 |
|---|---|---|
| CreateIntent 保存权限 revision；正常确认；权限变化阻断 | 首 RED 保留；9719e69 实际 SQLite/HTTP 保存与拒绝独立 PASS，真实产品成功确认等待 H/F | H 实际 SQLite/HTTP，F 实际页面，保留前端比较 |
| 控制器 start/send/read/stop/resume、持久身份与三值状态 | R 阶段报告 PASS，组合未验 | 当前真实 controller 与 SQLite、明确活动窗口语义；H/F 产品链路 |
| 控制器自产会话逐 event_id 定位、六字段 HTTP、分页与 GET 只读 | 原 task H 独立15行 / 模块18 HTTP ID 实际匹配 PASS | H 原档 / SQL / GetEvent / HTTP 比较；原503与完整会话 partial 独立保留 |
| 来源覆盖、追加、停止和恢复边界 | 当前原档 nlink2 / errors1 / snapshot_complete false 为 FAIL；既有 PARTIAL / UNKNOWN 保留 | 已交原来源消费者核链接所有权；`ArchiveAppendOnly=unknown`，文件 PrefixComplete / SessionComplete 不足，恢复 raw gap 独立记录 |
| 同 URL 未登记基线、真实登记、首次 Today DOM 三分钟 | 首注册至真实 native DOM591116 / 591117毫秒 FAIL；dispatch 至 DOM14.516 / 14.517秒 PASS | 原13:23:34.0366105Z起点不变，两个真实 archive ID 逐 SQL 核实，两种计时分别报告 |
| 人工 done / released 和事件 reconciliation 许可 | observe 人工映射已批准；R 实现中；旧 task 实际完成释放未验 | R 权限与真实适配器，H 投入和 SQL 核验，F 唯一 UI；reconciliation 继续核 reconcile |
| 注册表扩展、中性 worktree 换手、terminal-only、互斥与权限拒绝 | 最终组合未验 | 对照 v1.2 与当前 release 任务；每项真实结果独立报告 |
| 默认检查、构建、完整浏览器、Windows 解压启动读写及重启持久化 | 模块未验 | 各原领域人员、唯一集成人员和发布包人员精确组合 SOURCE |
| Procmon、两次真实 00:30、用户自由工具与实际作品 | NOT_RUN / 未决定 | 不降低门禁、不扩大授权 |

## 发布交接约束

原始源、凭据、模型输出、会话文件和正式数据不进入 HTTP/Git/release 包/stdout；只传安全派生和本机证据指针。旧正式 V4、锁定协议、认知 schema、engine、mapper、七类权重、日程和既有 17321 服务保持冻结。每阶段 SOURCE 与独立 REPORT 普通提交并 push，保留第一 RED/UNKNOWN/NOT_RUN；不 force、rebase、reset 或 squash。最终 GitHub Release 由总负责人接收具体可审查结果后发布，本模块不发布。
