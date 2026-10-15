# Works Relay HTTP 与 L1 会话采集

## 2026-10-16 真实输出与宿主目标检查的生产读取

SOURCE `b22395c6eb1eed0be9e87957ae21111b5ce8f782`，分支 `songconmaisaix31-design/works-relay-http`。当前实际客户端为 Codex，具体模型未获得本轮可独立核实的资料。完整普通历史包含发布指定、实际构建通过的 I 阶段 `35cc64faac271d43e5f63103a810081b0eb7ac62`；没有复制未提交集成内容。

生产 `assembleRelay` 使用同一个原 ObjectStore、EventService、EventRepository 和 SQLite 注入既有 `RelayGoalVerificationReader`。它读取原已提交的检查关系，重新核当前项目权限、settings revision、批准目录身份、实际应用文件字节和本次准确 owned 输出来源。返回原 task / worktree / terminal / controller / startup / session / native / input operation / turn 关系及实际 SQL 输出 ID、检查事件引用。R 继续核原 accepted dispatch / send 收据，完成仍由 F 原界面的人工动作执行，资源释放独立审计。

`ReadRelayGoalOutput` 只读取原控制器正选 Codex 档案中准确 turn 的 `task_complete` / `turn_complete` 和 `last_agent_message`。来源须通过原单链接、同文件身份、header / CWD 和完整当前快照检查；没有输出、多个匹配输出、文件替换或时间关系不符均拒绝。读取不推进生产游标，也不生成 started、live 或 session_complete 事实。`ReadCommittedRelayGoalOutput` 按准确 origin 定位真实事件库，逐字段核原投影和 session / project，再返回数据库分配的 ID。

原生输出对象、H 使用文件编辑工具实际应用的 `stats.go` / `stats_test.go` 对象，以及宿主实际检查的 stdout / stderr 对象分别保存。关系沿现有 StoredEvent 保存；失败检查保留失败事件，读取者只接受完整成功结果。没有新增迁移、独立证据数据库或 HTTP 结果写入入口。模型正文、原生身份与档案路径只留在本机；HTTP 仅接收原安全 evidence 引用，session-events 保持精确六字段。读取接口不执行检查命令。

实际目标保持原公开 baseline `da3bdddb085808863090a05f60cc1c64415039cf` 和四个批准文件，H 独占已登记的 D 盘 `verification/public-project` checkout，仅允许修改两个源码 / 测试文件。首输入前固定的独立宿主测试涵盖全部 Mean 回归，以及 Median 奇偶长度、负数、重复值、输入保持、空输入、NaN / Inf 和大有限中间值。Go 原生 overlay 只增加宿主测试文件，不替换四个目标文件；固定命令仍为 `go test ./...`、`go build ./...`、`git diff --check`。该 baseline 的原测试、构建和差异检查实际通过，完整新增目标检查实际因缺少 Median 拒绝；未预先实现目标。

本次实际目标检查沿原 `TestRelaySessionActualCommittedEvidence` 入口调用独立 helper。在明确的本次 configured-access 范围内，缺受信应用确认、合法原输出、准确已接受输入关系或真实检查结果即失败。没有新增 Skip 或静默通过入口。实际窗口尚未执行该核心路径。

| 验证命令 / 操作 | 实际结果 |
|---|---|
| `go test -mod=readonly -p=1 ./cmd/server ./internal/adapters/events ./internal/adapters/sqlite -run '^(TestRelayGoal.*\|TestRelayPublicContextActualGitSQLite\|TestRelayTrustedRegistrationActualSQLiteHTTP\|TestRelayFactsActualHTTPSQLiteReadOnlyAndPermission\|TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown)$' -count=1 -v -timeout=120s` | PASS，8 项，零 Skip；三个 package 分别 35.324 / 9.584 / 16.592 秒 |
| `go vet -mod=readonly -p=1 ./cmd/server ./internal/adapters/events ./internal/adapters/sqlite ./internal/adapters/httpapi` | EXIT 0 |
| `go build -mod=readonly -p=1 -o <本机 H 产物目录>/astrocyte-goal.exe ./cmd/server` | EXIT 0，Windows 生产程序已生成，未据此启动原生会话 |
| 原 `TestRelaySessionActualService` 在已登记 D 配置根启动并正常关闭 | PASS，实际检查 3.29 秒，package 15.431 秒；实际 SQLite task / registered agent / native session 均 0 |
| 官方 YAML 2.9.1 严格解析与修复前后完整语义比较 | PASS，139 个唯一路由；保留全部正确路由与 schema 含义 |
| `node scripts/validate-examples.mjs` | EXIT 0，273 个契约示例及能力边界 |
| 原 Redocly CLI，经 `pnpm exec` 和原 `bin/cli.js` 直接运行 | 两次均在输出契约有效、13 warnings 后发生 Windows libuv assertion，直接运行退出码 `-1073740791`；命令保持 FAIL |

使用 `GOMAXPROCS=2`、Go `-p=1`、Node 512 MB，全部新检查日志、临时文件、对象与产物使用 D 盘。日志指针为 `D:/Astrocyte-agent-assets/20261015/h-http-goal/task-3004b636a3ab/`，最终八项结果为 `goal-production-final-tests.log`，构建和 vet 分别为 `goal-production-build.log` / `goal-production-vet.log`，CLI 原件为 `goal-production-openapi.log` / `goal-production-openapi-direct.log`。首 Token 类型接缝编译失败和首新增实际入口 SKIP 分别保留在 `goal-production-combination-first.log` / `goal-production-combination-second.log`；后继实际入口已按当前要求组织。

临时处理唯一 I 的 OpenAPI 单文件时，仅以文件编辑工具删除 7 组完全相同的第二份路由与 1 个完全相同的第二份 schema；严格 YAML 和全部解析语义相等检查通过。未操作该树的 stage、commit 或 merge，已明确停止写入并归还；本 H 消费 35cc 后执行同样的必要契约修复。

真实剩余限制：本次没有登记控制器、初始化原生会话或提交模型输入；合法自产原档、工具关闭、目标正向验证、F 唯一人工完成、180 秒 Today 与独立释放仍待授权窗口。当前 Goal 读取要求实际 archive SQL 记录；现有 native_notifications 采集选择不同时提交 archive，不能据通知通过登记档案或 Goal 通过。旧 nlink2 来源 FAIL、append-only unknown、raw 恢复 gap、session_complete=false 和首次计时失败保持。Redocly 的实际进程退出问题尚未解决。完整新组合、后继 S2 / S4 / S5 装配与最终包验收分别继续，不以本阶段结果替代。

## 2026-10-16 固定批准 Git 上下文生产入口

SOURCE `c0526d444e4a6fe798cdd1514b72a303de10b5b4`，分支 `songconmaisaix31-design/works-relay-http`。开发使用 Orca Codex 0.162.0 / GPT-6.1-Sol xhigh fast；本轮沿原七个 H 路径提交，没有改 B 配置、依赖或领域实现。

主入口和实际服务入口通过 `relayPublicLaunchContexts` 注入既有 `RelayLaunchContextReader`。可信本机配置须同时提供 `ASTROCYTE_RELAY_PUBLIC_GIT_ROOT`、完整小写 commit `ASTROCYTE_RELAY_PUBLIC_GIT_COMMIT` 和 JSON 文件数组 `ASTROCYTE_RELAY_PUBLIC_CONTEXT_FILES`。文件清单最多 32 项，要求唯一相对路径；配置不进入 HTTP 请求。根限接入目录的 verification 范围，并核规范路径和真实目录身份。读取继续经过原 `LocalProjects.ReadProjectContext`，核当前项目、settings revision、目录授权、固定文件版本及正文与原 commit 的字节一致；读取前后均重核 HEAD。权限撤销、旧 revision、文件或 HEAD 改变立即拒绝，错误只返回固定安全原因。

既有两阶段领域入口在初始化和首次输入前分别消费该读取器。main 的 Swarm 薄装配使用同一 SQLite、S2 批准服务、Registry 和原 object store，转发已消费的原生证据桥；独立 Relay L1 的生产调用链保持不经过 S3 Claim。session-events 精确六字段保持现有契约。

| 验证命令 | 实际结果 |
|---|---|
| `go test -mod=readonly -p=1 ./tests/s2 -run '^TestActualSQLiteApprovalRestartAndRollback$' -count=1 -timeout=120s` | PASS，18.395 秒；原十秒 HTTP client 时限、真实 Git / SQLite / Context 安全 DTO 与批准重启检查保持 |
| `go test -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' ./cmd/server -run '^(TestRelayPublicContextActualGitSQLite\|TestRelayTrustedRegistrationActualSQLiteHTTP\|TestRelayFactsActualHTTPSQLiteReadOnlyAndPermission\|TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown)$' -count=1 -timeout=120s` | PASS，3.585 秒；本仓库四个公开文件建立独立 Git / SQLite 项目，实际授权、撤销、revision、正文和 HEAD 变更拒绝；原三项 HTTP / SQLite / Origin / CSRF / 同服务登记与关闭检查保持 |
| `go vet -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' ./cmd/server ./internal/adapters/httpapi` | EXIT 0 |
| `go build -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' -o D:/Astrocyte-agent-assets/20261015/relay/H/current/context-server.exe ./cmd/server` | EXIT 0；产物未启动 |
| `git diff --check` / 暂存范围检查 / 后继普通 push | PASS；SOURCE 七个 H 路径 |

检查使用 `GOMAXPROCS=2 / -p=1`、共享 D 盘 Go cache 与 H 独占 D 盘 TEMP / TMP / GOTMPDIR。日志和产物在 `D:/Astrocyte-agent-assets/20261015/relay/H/current/`，分别为 `context-first.log`、`public-context-and-http-second.log`、`context-vet.log`、`context-build.log`。新增检查首场 SQL 表名错误保存为 `public-context-first.log`；SOURCE 首次 push TLS 失败保留，后继普通 push 成功。原 Context 十秒超时、旧 Reader 单树 build FAIL、P579 五文件只读 overlay 的适用范围、旧来源 nlink2 FAIL、raw 恢复 gap、session_complete=false 和旧计时失败保持各自记录。

本次新检查的实际 `local_agent_sessions` 数量为零，没有原生初始化或模型输入。真实批准目标四文件的原生输入、控制器自产来源、工具关闭、逐 event_id SQLite、六字段 HTTP、Today 180 秒和完整产品验收仍待唯一 I 准确组合及原 F / H 授权窗口；本次定向检查只验证固定上下文读取和现有无模型组合。

## 2026-10-15 安全 Context、两阶段 HTTP 与同实例可信登记

SOURCE `fbf2a4a4752ef2842900b3d3a851f80986734ce9`，分支 `songconmaisaix31-design/works-relay-http`。本轮沿原 H 工作树恢复，原 S2 按限定四路径编辑和 stage 后明确停止写入，产品模块独立核 U0 后仅 H 继续原消费序列。准确消费 S2 `aa445adf834c3a397a0817bec0593fcf1039420a` 为 `c84cec7`、`02d16dd05bfb4f944c480ba844dbc2d3a79d63e4` 为 `1e6ee32`，S3 `70e6e5a03fca9a19aa0a70e44b23b78337d34714` 为 `688ebb8`，R 登记 SOURCE `4d93a52028e5ea72c705c5c6e7f30d0cb7f016bb` 为 `525288a`。两阶段领域来源仍为 `cc326be57a9a5f3bdd45687b0ac23b972adb8cef` / REPORT `41b4d71452f711f972d3c9836e0484fa94fdb854`。

Context HTTP 只投影安全 DTO：snapshot 省略私有绝对 Root，scope.roots 使用该固定项目内的相对路径，session binding 省略 native / controller / startup operation 身份及能力原因原文。生产 S2 配置使用同一 S4 snapshot provider、原 object store 和原 Registry 的只读 `ReadOwnedNativeSession` 桥；冷 Registry 没有原持有对象时明确拒绝。内部仍保存原始归属与工作区身份。新的 owned 原生正向绑定尚未执行。

`POST /api/v1/relay/intents/{id}/prepare` 使用既有 `RelayConfirmRequestV1`、人类 session、Origin、CSRF 和 Idempotency-Key，转发原 `RelayIntentPreparer`。它表示空 Message 初始化；首次输入仍由原 Confirm 操作提交。生产 driver 已接入 `BeforeFirstInput: sourceGuard.Check`；实际来源检查、工具关闭与固定上下文、同控制器首次输入的正向链路尚未运行。session-events 保持精确六字段。

可信登记使用同一 Service、Registry、store 和原 observer。`relaySessionRuntime.ReloadTrustedAgents` 只读取启动时固定根下的 `agents.json`，重新核根文件身份和既定活性窗口，再显式转发原 `RelayTrustedAgentRegistrar`。原 observer 每轮观察前执行该重载；GET 只读取已发布快照。完整快照由 R 校验、复制并原子发布，在途身份 / 策略冲突按原 store 规则拒绝；解析、根身份或启动策略变化立即返回错误并终止该次观察。没有新增 HTTP 配置入口。

| 验证命令 | 实际结果 |
|---|---|
| `go test -mod=readonly -p=1 ./tests/s2 -run '^TestActualSQLiteApprovalRestartAndRollback$' -count=1 -timeout=120s` | PASS，5.304 秒；真实 Git / SQLite / 批准重启与 Context HTTP，含人类和项目 Agent 的安全相对路径投影；`recovery-s2-context-first.log` |
| `go test -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' ./cmd/server -run '^(TestRelayTrustedRegistrationActualSQLiteHTTP\|TestRelayFactsActualHTTPSQLiteReadOnlyAndPermission\|TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown)$' -count=1 -timeout=120s` | PASS，2.255 秒；同服务空登记、原 observer 激活、GET 无重载、八路并发、畸形登记拒绝、启动策略保护、Prepare 当前权限 403、真实 SQLite / HTTP / Origin / CSRF / 关闭；`recovery-http-registration-second.log` |
| `go build -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' -o D:/Astrocyte-agent-assets/20261015/relay/H/recovery-context-registration-server.exe ./cmd/server` | EXIT 0；产物未启动，`recovery-context-registration-build.log` |
| `go vet -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' ./cmd/server ./internal/adapters/httpapi` | EXIT 0；`recovery-context-registration-vet.log` |
| 官方 Redocly lint / `node scripts/validate-examples.mjs` | EXIT 0，13 项已有 unused component 警告 / PASS，273 个示例 |
| `node scripts/check-architecture.mjs` / 十个 Go 文件 `gofmt -l` / `git diff --check` | PASS，34 个包边界 / 无输出 / EXIT 0 |

本机检查日志与构建产物位于 `D:/Astrocyte-agent-assets/20261015/relay/H`；本次实际登记检查的自有数据库和安全聚合位于忽略目录 `dist/relay-registration/owned-*`。检查使用已安装的真实 Codex 启动入口作为登记值，没有启动控制器；实际任务、收据、原生会话、terminal、worktree 和模型输入均为零。本轮开发客户端为 Orca Codex，公开运行状态显示 `gpt-6.1-sol`；Go `GOMAXPROCS=2 / -p=1`，Node 512 MiB。

首命令的未完整引用 overlay 参数被 PowerShell 截断，原始失败保留 `recovery-http-registration-first.log`，该命令没有进入编译或实际检查。源码首次 push 成功，独立远端核对首场 TLS 失败与后继检查分别保存。既有 H 单树旧 Reader build FAIL、P579 五文件 readonly overlay 组合 PASS、S4 首 search unknown、旧 SQL 表名 FAIL、nlink2 来源 FAIL、raw 恢复 gap、session_complete=false、旧三分钟 FAIL 和未知交付均保留。

F 继续独占产品界面与客户端生成。Calls / Tools / Targets Known 的待定范围保持原 Claim 拒绝，本次目标登记 t0、原生初始化和模型输入仍为零。双完成通道由原 R 实施后交准确 SOURCE，started 独立保存，资源释放保持原审计。实际固定上下文输入、合法来源正向核验、逐 event_id SQLite、真实 CLI / Today、S5 生产注册、完整默认浏览器和最终包验收尚未完成；本阶段交付支持唯一 I 准确组合。

## Context 与 S3 依赖及既有检查

本阶段 SOURCE 为 `59a7150b1143680b67fe53fdf831fcca5f328660`，分支为 `songconmaisaix31-design/works-relay-http`。准确消费 S2 `b7e3b1c5c9d56c5bf68f67a2fe5b70d8970f191d`、S3 的 `b220ab2`、`3b78fd0`、`e18c096`、`3c568c7`、`ec653dd`、`93d1f6b`、`2bd5d26`、`10fa36c` 八项源码，以及 B 的 action 提交 `c110201` 和可信配置 `e346c40cc9edcbd5b4588e47585f3c0b616b0770`。配置冲突仅由原 B 在模块限定范围内编辑并 stage，归还后 H 核实 U=0、完整 index blob `2d7a7bdc1a31b8f97545baed2f74806c2168d796` 和仅两件 staged 文件，再普通继续为 `b535216`；没有消费其他 B 历史。

生产入口将明确的环境身份、工具路径、缓存目录和预算传给同一个 S4 provider，同时把该对象、原 object store 和原 Registry 的私有 owned 读取桥配置给当前 S2 服务。环境身份为空时保留明确的缺配置结果。HTTP 只提供当前批准下的固定 Context Packet 构建、读取和原生绑定命令；原生 ID 不进入绑定 DTO。OpenAPI 使用现有固定引用、Goal、Scope 和 Stops，闭合描述安全快照、后继关系、六组上下文记录、能力观察和错误状态。

生产 `SwarmS3` 通过原 SQLite repository、同库批准桥和当前 S2 服务构造原 `NewSwarmLedger`，注册原 S3 全部处理器；原生执行、成员与证据端口缺失时仍由现有领域拒绝。Agent 白名单只增加已鉴权的任务读取、work item、claim、submit 和 diff，人工任务创建、暂停、取消、恢复、验收和拒绝继续经过人类入口。OpenAPI 的任务、lease、artifact、时间线、执行引用、未知用量与命令对应原服务实际字段和 200 响应。

私有 `VerifyRelaySessionArchive` 使用原 NDJSON 库和既有 header 身份 / 工作目录检查，在同一文件句柄上先后检查正选来源与单链接身份。该函数只返回本机文件身份，不产生事件或游标；`relaySourceGuard.Check` 已采用此检查。当前尚未接入新的两阶段实际窗口，新的真实原档正向预查仍为 NOT_RUN，实际调用前不能据编译或文件身份测试宣称来源通过。

本机安全日志在 `D:/Astrocyte-agent-assets/20261015/relay/H`，构建及临时目录使用前述 D 盘专用进程环境；Go `GOMAXPROCS=2 / -p=1`、Node 512 MiB。当前 Orca Codex 客户端的具体模型没有独立回显，本阶段模型调用为零。

| 验证命令 | 实际结果 |
|---|---|
| `go test -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' ./cmd/server -run '^(TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown\|TestRepositoryContextConfig.*)$' -count=1 -timeout=180s` | PASS，1.998 秒；实际主服务、同库 ledger、空任务读取、缺授权 / CSRF 拒绝、五项缺配置、配置三项和 native sessions=0；`v14-context-s3-main-actual.log` |
| 原 S3 实际 HTTP 授权、时间线和撤销检查 | PASS，3.319 秒；此前同组 H atomic 检查 PASS，2.407 秒；`v14-s3-http-first.log` |
| `go test -mod=readonly -p=1 ./tests/s3 -run '^TestActualApprovalMissionConcurrentIdempotencyRevokeAndRestart$' -count=1 -timeout=180s` | PASS，6.830 秒；`v14-s3-approval-actual.log` |
| `go test -mod=readonly -p=1 ./tests/s2 -run '^TestActualSQLiteApprovalRestartAndRollback$' -count=1 -timeout=180s` | PASS，23.068 秒；真实 SQLite、固定 Git 快照、批准、重启、后继与事务回退；`v14-context-actual.log` |
| `go test -mod=readonly -p=1 ./internal/adapters/events -run '^TestRelaySourceIdentityActualFileAndAliasRejection$' -count=1 -timeout=90s` | PASS，0.935 秒；实际普通文件身份与 hardlink 拒绝，未使用伪造原生文件；`v14-source-archive-identity.log` |
| `go vet` 当前 cmd/server、HTTP、events、repositorycontext / `go build` 当前 cmd/server，均带准确 P579 readonly overlay | EXIT 0 / EXIT 0；`v14-context-s3-final-vet.log`、`v14-context-s3-final-build.log`；构建产物未启动 |
| 官方 Redocly CLI lint / `node scripts/validate-examples.mjs` | EXIT 0，13 项 unused 等警告 / PASS，273 个现有示例；`v14-context-s3-openapi.log`、`v14-context-s3-examples.log` |
| `node scripts/check-architecture.mjs` / `git diff --check` | PASS，34 个包边界 / EXIT 0 |

此前第一次 S3 正则在 `tests/s3` 未命中测试，保持 NOT_RUN；随后已用上表准确测试名运行实际并发批准检查。原 B 在本树进行配置检查时的旧 Reader 变参编译 RED、H 单树旧 Reader FAIL、S4 首 search 超时 unknown、旧 SQL 表名 FAIL、旧 nlink2 来源 FAIL、旧三分钟 FAIL 和未知交付保留实际记录。当前 build / vet 使用的 overlay 仍为既有 P `579a1ce` 五件准确 Reader 源，尚未将 B 新 Reader 闭包写入本树。

F 是唯一实际 GUI 操作人，生成客户端与当前完整配置 HTTP 正向组合交原 owner 消费准确 SOURCE。Calls、Tools、Targets 的 Known 完整性继续等待根转交的用户决定；Claim 保持原拒绝。本次目标原生登记、controller、模型输入和新 t0 均为零；S5 薄装配、真实 CLI / Today / 逐 event_id SQL 及最终产品验收继续，不发送阶段 worker_done。

## 2026-10-15 仓库上下文公共装配阶段

本阶段 SOURCE 为 `31256d5d6758daf25892d55df73feffde7a7e573`。沿原作者准确消费 S4 的 `9eff2ed`、`263bc70`、`1a55e16`、`896a7b4`、`293af63`、`f407337` 六个源码提交，以及 B 的 `98bb923` 迁移和 `6b78f64` 依赖提交。三件依赖冲突由原 B 通过文件编辑工具解决并明确归还；S4 报告冲突先由原 owner 处理，后经模块 `msg_d8d09a25bb0e` 临时授权 H 仅处理该报告的后续冲突。H 已在 `msg_1ef84752900b` 明确停止并归还该单文件范围，S4 业务文件保持准确源码内容。

生产 `httpapi.Services.RepositoryContext` 现在在原 `NewServer` 和原 session guard 中注册五个 S4 处理器。现有项目 Agent 白名单只放行 map GET 与 search / impact POST；索引和地图维护仍限人类。OpenAPI 闭合描述全部请求、预算、版本、结果和缺口；Search 必需 query，Impact 必需 symbol_id，分别注明实际 1024 / 512 UTF-8 字节限制。查询不接受调用方提供的原始根、Caller、HEAD 或 generation。

主入口显式保留空能力集合，因此缺可信配置时实际接口返回 `provider_unavailable / 503`。私有 `assembleRepositoryContext` 使用原 SQLite 和准确 S4 Options 构造同一个 provider，返回供后续 S2 快照装配使用。生产 EnvironmentID / 工具 / 投影 / 预算配置仍由 B 精确交付；当前没有把测试路径或测试 EnvironmentID 作为生产默认。配置完整的实际图 / 地图 HTTP 正向检查仍由原 S4 消费本 SOURCE 后执行，本阶段没有将其填为通过。

B 的成熟依赖检查脚本经产品模块授权，在 H 树内由原 B 用文件编辑工具精确消费 `efa137eb36177f69643c750c26ad8eb59b446741:scripts/check-dependencies.mjs`。本人停止写入与提交权归还见 `msg_7f13d6c5cd8c`；H 和模块核得该文件完整内容相等。脚本使用 Go 的 JSON 元数据检查准确依赖，没有由 H 修改其逻辑或合入 B 的其他历史。

新增私有 `relaySourceGuard.Check` 已编译：它读取原 Registry 持有的 controller / startup operation、真实进程启动策略和权限，再核正选文件与单链接身份；预查和后续输入可复用同一内存身份。缺少正选 metadata 或文件尚未产生时保持固定 NotYetAvailable 原因。该回调尚未接入真实两阶段模型窗口，不能据本阶段构建声称实际工具关闭或来源通过。采集器的私有 archive-only 选择也尚未用于新的实际窗口。

新命令统一使用进程环境 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`，`GOTMPDIR/TEMP/TMP` 为 D 盘 H 独占目录；`GOMAXPROCS=2`、Go `-p=1`、Node 512 MiB。安全日志位于本机 `D:/Astrocyte-agent-assets/20261015/relay/H`。

| 验证命令 | 实际结果 |
|---|---|
| 原两项真实 HTTP / SQLite 测试，准确 P579 Reader readonly overlay，`-count=1 -timeout=120s` | PASS，15.927 秒；五项缺配置接口实际 503、CSRF 403、畸形编码 / 重复标量 / 未知参数 400，安全错误无本机原路径，项目不变，实际 native sessions=0；`v14-s4-http-first.log` |
| `go vet -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' ./cmd/server ./internal/adapters/repositorycontext` | EXIT 0；`v14-s4-vet.log` |
| `go build -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' -o D:/Astrocyte-agent-assets/20261015/relay/H/public-s4-server.exe ./cmd/server` | EXIT 0，产物未启动；`v14-s4-build.log` |
| `node scripts/check-architecture.mjs` | PASS，33 个包边界 |
| 同一官方 Redocly CLI 直接 lint / `node scripts/validate-examples.mjs` | EXIT 0，4 项既有警告 / 284 个示例；`v14-s4-openapi-reviewed.log`、`v14-s4-examples-reviewed.log` |
| `node scripts/check-dependencies.mjs` / `git diff --check` | EXIT 0，准确依赖与实际安装一致 / EXIT 0；`v14-s4-dependencies.log` |

本阶段首次新增路由缺少既有 validator 要求的 supported 标记，契约检查失败原件为 `v14-s4-examples-first.log`；最终标记与实际实现一致，没有改 validator 或减少检查。先前 H 单树 Reader 构建 FAIL、首 SQL 表名 FAIL、旧 nlink2 来源 FAIL、旧三分钟 FAIL 及所有未知交付保持原结果。

当前没有登记 Agent、初始化原生控制器或发送模型输入；F 仍是唯一实际 GUI 操作人。Context 的安全 DTO / 原生绑定、S3 准确八提交闭包、B 可信配置和 R 两阶段 HTTP 是后续准确装配内容；真实 CLI、Today、逐 event_id / 六字段 SQL、固定目标核查及产品最终验收均未因本阶段通过而完成。当前 Task 继续，不发送阶段 worker_done。

## 2026-10-15 当前来源与安全契约阶段

本阶段 SOURCE 为 `37156a31478c22480abb7ae64a5f62f20d620b1f`，分支 `songconmaisaix31-design/works-relay-http`；当前 Task `task_3004b636a3ab` / Dispatch `ctx_e7abbaf902da`。已普通合入 R SOURCE `ce0302835cb7477eb606107351a9ab46dd8d4714` / REPORT `5d145cb148d1cf8d7053e2b2daf650eb32a95ef8`，原四项未提交修改已完整保留并完成适用检查。开发使用当前 Orca Codex 客户端，具体运行模型未独立回显；本轮外部模型调用为零。

`events.VerifyRelaySessionSource` 复用采集器既有的授权根、原工作目录和路径保护，只检查控制器已选中的文件及其单链接身份，不读取正文或创建采集游标。返回的文件身份只用于本机后续核对。多名称来源保持 `source_identity_unverifiable_or_aliased` 拒绝；该检查不能证明档案已完成、永久追加、模型消费或目标完成。实际采集仍须核原绑定、header、CWD、native identity 与真实提交事件。

原服务包装器现转发准确 `ListOwnedTurnBindings` 私有端口。公共 `facts.current_ownership` 接受 R 的 `confirmed / controller_exited`：它只表示原已退出的 controller 对象仍被持有，不能据此执行 Read 或 Send。session-events 继续精确六字段，原生身份、文件身份、路径和原始输出没有进入公开 DTO。

真实服务验证入口支持从明确本机公开 Git 来源以 `--no-hardlinks` 取得准确 baseline，并在启动时重核 HEAD。实际目标仍为获准 S3 公开 Go 项目 baseline `da3bdddb085808863090a05f60cc1c64415039cf`；没有在本轮登记、初始化控制器或发送模型输入。

全部新 Go 命令使用进程环境 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/H`、`GOTMPDIR=D:/Astrocyte-agent-assets/20261015/go-temp/H`、`GOMAXPROCS=2` 和 `-p=1`。安全日志保存在本机 `D:/Astrocyte-agent-assets/20261015/relay/H`，没有修改全局环境。

| 验证命令 | 实际结果 |
|---|---|
| `go test -mod=readonly -p=1 '-overlay=dist/l1-live/reader-overlay/overlay.json' ./cmd/server -run '^TestRelay(FactsActualHTTPSQLiteReadOnlyAndPermission\|HTTPAtomicListenerActualSQLiteSessionAndShutdown)$' -count=1 -timeout=120s` | PASS，2.012 秒；真实 HTTP / SQLite、权限、GET 只读、实际 listener 与正常关闭。overlay 仍为已登记 P579 的准确 Reader 五文件；没有宣称 H 单树或全仓通过 |
| `go test -mod=readonly -p=1 ./internal/adapters/events -run '^TestRelaySourceIdentityActualFileAndAliasRejection$' -count=1 -timeout=90s` | PASS，1.050 秒；实际公开普通文件、稳定身份、检查前后文件不变、未授权根、取消与实际硬链接拒绝 |
| `node scripts/validate-examples.mjs` | PASS，284 个当前契约示例 |
| `node node_modules/@redocly/cli/bin/cli.js lint contracts/openapi.yaml --config contracts/redocly.yaml` | EXIT 0，4 项既有警告；`v14-openapi-direct.log` |
| `node scripts/check-architecture.mjs` / `git diff --check` | PASS，30 个包边界 / EXIT 0 |

首 overlay 参数没有整体引号，Go 将参数路径截断，测试未启动；原件保留 `v14-baseline-composition.log`。`pnpm exec redocly` 完成校验后宿主进程出现 `UV_HANDLE_CLOSING` 并异常退出，保留 `v14-openapi.log`；上表直接运行同一官方 CLI 的 EXIT 0 单独记录。旧单树 Reader 构建 FAIL、旧 nlink2 来源 FAIL 和旧三分钟 FAIL 保持原结果。

ROOT 经模块当前消息 `msg_1320601dbef0` 明确选择由原 R 补充 Relay 同 controller 两阶段入口：空 Message 初始化不触发模型，后续首输入保持原 controller、权限 revision、唯一 operation、source guard 和工具关闭检查。准确 SOURCE 尚待原 R 交付；当前没有用独立 LocalProject 会话代替 Relay / Today。原 F 继续唯一 GUI 操作，后续仍须空注册基线、observer ready、来源实际就绪和模块单窗口授权。新真实 CLI、Today 两尺寸、逐 event_id / 六字段 SQL 核对、合法目标核查及人工完成均未执行，当前 Task 继续。

H 分支：`songconmaisaix31-design/works-relay-http`。当前非原生依赖 SOURCE 为 `4f45a3ab50e6a2863b940915874374f58e65f1da`，生产 listener 接口为 `1aadeebaba2330bb8b786989f577415da5208fd5`；前阶段原生验证准备 SOURCE 为 `b6c67e91555522b3d7515d405c4c69a93d0e0bc8`。主装配保护源接线为 `b247dfb12b938739f09acc73e6b7582376951a0c`，真实 SQLite 配置测试为 `78903848199b915a21e95e42f52cbed64c6d606c`。R 的人工完成权限 SOURCE `d14b08dc50014baf9ee636543564b040112cf3b5` 与 REPORT `fae304b80cb3288b9edd08f10500c6e5af7e4210` 已普通合并。已关闭旧服务的业务版本为 `e5714bb868a6519ec70520bca1c0a1b6c4678fb8`。客户端为 Codex API，工作会话没有暴露精确部署模型名称；旧实际原生控制器为 Codex `0.162.0`，原 `turn_context.model` 为 `gpt-6.1-sol`。

## 端口 0 与精确配置交接

本轮 Task `task_abf1f8b1e12e` / Dispatch `ctx_d7d8c2828fba` 只完成非原生依赖。`ASTROCYTE_PORT=0` 由 Go 一次调用 `net.Listen("tcp", "127.0.0.1:0")` 原子选择端口，主装配一直持有该 listener，并以实际端口构建 HTTP 服务。`httpapi.Server.Serve(net.Listener)` 直接使用同一个已绑定的 loopback TCP listener，关闭时返回 `http.ErrServerClosed`。无第二次绑定、端口探测后释放、换端口或重启 fallback。

主装配在路由构建完成、listener 已绑定且仍被持有时，向 stdout 编码一行 JSON：`{"type":"http_listening","api_url":"http://127.0.0.1:<actual-port>"}`。仅此安全地址记录进入该流；它表示绑定完成，HTTP 响应 readiness 仍由消费者对该 URL 的真实 health GET 确认。消费者不得把公告先于 Serve 当成 HTTP 已可响应。每次进程启动重新消费其实际公告，正常关闭不补发地址。B 的 `tests/s1/server.mjs` 消费此接口，API 不再通过 Node 先占用、释放再交 Go 重绑；配置的 Web Origin 仍由既有 Origin 白名单处理，API 同源以真实请求地址核对。

`cmd/server/workbench_config.go` 已核对与精确 SOURCE `24b6390bcad60a2a20b671dcb60725e77eb3a7a4` 无 diff，并在 B 复制期间冻结。B ACK SOURCE `4a6076c402e0b5d31982154a31a905bee6127477`：内容逐行一致、实际 SQLite 两项测试和 build / vet 通过，文件写入范围已归还。H 本轮没有修改 B 的 `workbench_config_test.go`、`config.go`、helper 或依赖。listener 三文件向 B 定向交付精确 `1aadeeb` changedblocks 与新 helper；消费者不复制额外 H / Relay 历史。

| 本轮命令或真实检查 | 结果 |
|---|---|
| `go test -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json ./cmd/server -run=^TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown$ -v` | PASS，0.11 秒；真实 Go HTTP、实际 SQLite 迁移和项目读写，无原生会话 |
| `go build -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json -o dist/http-listener/server.exe ./cmd/server` | PASS；只读 overlay 为 P `579a1cede98b47b08a3126f5d3974da938348929` 的五个精确 Reader 文件 |
| `go vet -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json ./cmd/server ./internal/adapters/httpapi` | PASS |
| H 单树 `go build -mod=readonly -p=1 ./cmd/server` | FAIL：旧单参数 Reader 与主装配可变参数接线未组合；没有整体合入 B 分支 |
| `gofmt -l` 四个本轮 Go 文件、`git diff --check` | PASS |
| 原子绑定、实际地址公告、同 socket 持有 | PASS，公告实际端口非 0，重复绑定该自有端口立即失败，服务地址保持 |
| Origin / CSRF | 实际 API bootstrap 与 CSRF 写入 PASS；配置的独立 Web Origin 读写 PASS；未授权 session、端口 0 Origin 和缺 CSRF 均 403，拒绝后 SQLite 项目数仍 0 |
| 实际 SQLite 与关闭 | 真实项目创建及设置 revision 增加可通过仓储核对；独立 query_only SQL 的 native session 数为 0；正常 Shutdown 后 Serve 返回关闭结果，原地址连接被拒绝 |

本机忽略目录 `dist/http-listener` 保存命令日志与每次独立 `owned-*` SQLite、公开项目目录和 `safe-result.json`。`actual-first.log` 保留首 SQL 表名检查失败；`actual-final.log` 为当前通过结果，不覆盖首失败。测试没有替代 API、driver、会话或模型，没有修改 schema / engine / mapper / weights、V4、日程和 17321。完整默认浏览器与 B helper 的实际组合消费由 B / F 继续验证，本轮通过范围不替代旧 L1 失败或新 v1.4 原生验收。

## 当前实现

当前 v1.4 验证准备 SOURCE 为 `261fb73921d2a1b929f40bf4d714de5356b9024d`。实际服务测试与归档 / 通知证据检查都消费本次 listener 和 `service-public.json` 保存的真实 URL，通过 `net/url` / `encoding/json` 校验明确 IPv4 loopback、已绑定端口、无凭据或额外 URL 内容；不再以固定端口连接其他服务。实际验证服务使用既有 `Server.Serve(listener)`；`ASTROCYTE_L1_API_PORT=0` 要求显式 `ASTROCYTE_L1_WEB_PORT`，两端口必须独立。公共记录解析拒绝尾随 JSON。

薄配置 SOURCE `25d843783d65ada98c0c51766a8d5b35b8f5f941` 对已有受保护目录、应用库、来源与快照的异常身份 IO 立即返回固定安全错误；首次尚未创建的目标保持原初始化语义。`os.ReadDir` 枚举出版条目后先核对 canonical 和实际身份，再判定目录；redirected generation 立即拒绝。真实 SQLite / source hardlink 保护与实际 Junction 拒绝通过；symlink 创建在本机因 OS error 1314 失败，其后保护语义未验证。首失败及 Junction 通过分别保留 `dist/http-listener/linked-generations-os.log` 与 `linked-generation-junction.log`，没有修改系统权限或增加 skip。该薄文件已交 B 精确消费，在 B ACK 前保持冻结。

上述准备通过真实 HTTP / SQLite 的 `TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown`（0.10 秒），精确 P579 Reader overlay build / vet 与格式检查通过。首正常关闭失败保留 `dist/http-listener/v14-address-first-failure.log`，当前通过日志为 `v14-address-second.log`；本次客户端主动释放自己的 idle 连接后正常关闭，没有修改原生状态或重试业务请求。新的实际会话、模型和 Today 窗口尚未启动。working 严格证据门禁仍待答；人工 done 已确认双通道，权限成立时 accepted / working 均可由操作者明确确认目标完成，started 保持独立且不回填。

安全事实 HTTP SOURCE 为 `074cf55de4f2d87812a87cde77780c67b76c8fe5`，普通消费 R 的 `1a4f120beff6a9a8d6033c0eed9c0a748f7f3f7e` 与 `79d93132967de20fb1fda3d8d74b4edff386178c`。`GET /api/v1/relay/tasks/{id}/facts` 返回闭合 `facts/session_relations/current_ownership`。事实中的 receipt_id 保持原业务字符串身份，逐项核对该任务可读收据；不会改变原收据 ID。历史关系与当前控制器归属独立，后者只读取原 Registry 持有的 controller 缓存并重新检查权限，冷 Registry 或缺原绑定返回 unknown，不执行 RPC、启动、刷新 live 或写入 store。事实类型、来源、状态和原因使用闭合词表，内部 controller/native identity、CWD、原始输入输出不进入响应；会话事件仍为原六字段。

`RelayIntentRequestV1` 明确提供 `handoff_from_task_id` 与 `handoff_context_confirmed`；人明确选择原任务并确认上下文，R 验证同项目 / Work / 原工作区、不同 Agent、已释放和退出且无 pending。声明的 handoff 关系保持 unknown，实际新 session 证据到达后独立追加关系；省略来源保持 ordinary intent，不自动推断上下文交接。

私有通知游标 SOURCE `ae2160a4a1be77314570a83cc149e019f584ba57` 在原来源 UUID 下使用准确 controller / startup operation 区分连接，恢复的新连接从自身序号范围开始，不套用旧序号。旧 checkpoint 保留在本机 history；GET 读取旧已提交 origin 前通过 R 的历史端口重核同原 session、native identity、provider、workspace 和原 operation。缺旧连接字段的记录保留原来源和 unknown，不倒造连接或当前所有权。归档仍按原文件身份和 offset 增量采集，通知完整性和 raw 恢复 gap 保持独立。精确 Reader overlay build / vet 与实际无控制器 HTTP / SQLite 两项组合 PASS；新的真实通知连接、恢复、历史 origin 正向仍 NOT_RUN，代码提交不替代该验收。

当前公开固定 JSON 目标已准备，实际 assistant 核验必须在来源 identityguard 接纳原档后，准确关联本次 dispatch / send operation。R 已供给原控制器身份和 `RelayOwnedTurnBindingReader` 私有对应端口，H 依照原绑定读取 operation 与实际 turn_id，时间邻近不用于推断对应。新目标核验保持 NOT_RUN，来源被拒绝时保持真实 FAIL，归档与通知分别验收。人工 done 如实保存操作者确认及实际完成通道，不记录为目标独立核验通过。

`TestRelayFactsActualHTTPSQLiteReadOnlyAndPermission` 使用本次独占公开 Git 项目、生产 Registry / worktree / runtime、真实 Go listener 和 SQLite，在没有控制器启动的情况下生成真实 ready 与拒绝收据。当前 PASS 0.30 秒，HTTP 每个 fact_id 精确对应实际 store，闭合响应字段检查、GET 前后全状态一致、跨 Origin 拒绝、grant 改变后权限拒绝、原生 session / terminal / worktree 为零。首失败保留 `dist/http-listener/facts-http-first.log`，当前为 `facts-http-final.log`；OpenAPI lint、精确 Reader overlay build / vet、格式检查通过。F 独占生成客户端与产品 UI 消费，新真实窗口仍需模块明确授权。

所有 Relay HTTP 使用既有 Go 服务、loopback、human session、Origin、CSRF 与 ErrorV1。GET 读取已持有任务、派生事件和原收据，不启动 CLI、不发送输入、不恢复会话、不读取原文归档。人工创建 intent 后必须明确确认权限；运行时 Service 是 task、terminal、worktree 与收据的唯一写入者。未知输入交付结果保持 `delivery_unknown`，不能自动重发。

`agents.json` 使用 `encoding/json` 解析受信任的 `AgentSpec` 数组，在启动及原 observer 的可信重载入口读取。条目包含 `name/launch_cmd/capabilities/drive_mode`，可选 `input_mode` 为 `line` 或 `stdin_eof`。完整有效登记由同一服务原子发布，调度按该登记选择既有原生适配器；注册本身不授予启动或工具权限。

`cmd/server/relay_assembly.go` 复用实际 Works、既有项目权限与 NativeRegistry、独立 relay.sqlite、既有 PendingEventService 和 EventRepository。observer 只观察人工已接受的任务，取消并 join 后关闭自有进程和存储。静默重读不更新活动时间或持久化未变化的游标。

`GET /api/v1/relay/session-events` 按当前 owned binding 与权限读取已经提交的真实事件。事件只有 `event_id/kind/topic/tokens/scope/ts` 六字段，时间倒序，ID 升序，分页最多 100 条。未知 tokens 保持 null。任务级事件使用仓储提交 ID，归档记录使用稳定 origin 查得的提交 ID；公开查询没有 prompt、reasoning 原文、工具参数、stdout、归档路径或 SQL 错误。

## 实际来源与覆盖

`ASTROCYTE_RELAY_SESSION_SOURCE` 未配置或为 `archive` 时采集控制器提供的准确归档路径。H 通过 `github.com/olivere/ndjson` 读取该实际 JSONL，校验 owned task、terminal、session、原生 ID、原 CWD 和文件身份；增量位置只在事件 admission 成功后持久化。未知记录仍计入覆盖，格式错误立即失败，未完成尾行保留待追加范围。

选择 `native_notifications` 时显式启用 R 的原生观察端口，从控制器持有的通知序号派生事件。其覆盖标为 `source=native_notifications`；归档覆盖标为 `source=archive`。通知和归档分别保留来源语义，通知的前缀或关闭证据不证明归档文件永久追加。

`snapshot_complete` 只表示当前选定范围的完整行已覆盖。归档的完整前缀、永久追加性和关闭证据不足时，`session_complete=false`，公开状态保留 partial 与具体 gap。R 的真实原生结果保留 `ArchiveAppendOnly=unknown`：当前接口没有永久 append-only 保证，原生会话允许 rollback/revert/压缩。恢复后的 raw 通知订阅缺口为 `native_resume_raw_events_unavailable`。进程退出、输入接受、任务完成和模型成功分别解释，不互相替代。

## 本机配置

接入状态位于 `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access`，注册文件为其中 `agents.json`，应用状态为 `app-state/state.sqlite`。启用接入模式用 `ASTROCYTE_RELAY=1`。`ASTROCYTE_RELAY_LIVE_SECONDS` 没有默认值；未配置时活动三值保持 unverifiable，不自行决定活性窗口。

`ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS` 是启动时的可信配置，显式值范围为 1–1800 秒，无值时传递 NativeDriver 原默认策略。配置错误在创建 Relay store 前立即返回。已授权独立验证使用 900 秒；该寿命不会改变活动三值或既有任务权限。实际验证程序通过 `ASTROCYTE_L1_API_PORT` 选择独占端口，新建项目的初始 actions 包含 `reconcile`。原项目和旧收据身份保持。

`ASTROCYTE_RELAY_PUBLIC_GIT_ROOT` 只用于明确获准的 Agent 自建公开 Git 验证项目。目录必须位于独立接入根的 `verification` 内，每次 Prepare 都以 `os.SameFile` 核对该唯一目录；其他目录立即拒绝。未配置时公开项目启动权限保持关闭。既有中性 worktree 仍核对干净、已提交、正确 Git 根和原身份，不修改用户作品或开放自由工具。实际 Codex 使用现有控制器关闭文件、命令、MCP、skills、浏览器、视觉和其他工具，只提供固定公开上下文。

实际验收服务使用生产 `assembleRelay`、真实 SQLite、真实 HTTP 和本次控制器自己产生的会话文件。API 为 `http://127.0.0.1:17342`，F 网页为 `http://127.0.0.1:17343`。本次数据根为 `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/h-production-current-20261014`；项目、worktree、数据库和进程均由本次验证持有。所有脚本、中间文件和安全聚合位于 H 工作目录内被忽略的 `dist/l1-live/current-20261014`。原件不进入 Git、HTTP、stdout 或 release 包。官方 CLI 自行鉴权，不读取或复制凭据；归档只读取控制器给出的本次 thread 路径，不枚举历史。

## 验证命令与当前结果

| 命令 | 结果 |
|---|---|
| 当前 H 原样 `go build -mod=readonly -p=1 ./cmd/server` | FAIL：本分支 Reader 仍为单参数，main 接线需组合消费 B 精确可变参数 Reader |
| 精确发布组合 `go test -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json -c ./cmd/server` | PASS，实际服务测试二进制已编译；使用下面登记的精确五文件 overlay |
| 精确发布组合 `go vet -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json ./cmd/server` | PASS |
| `node scripts/check-architecture.mjs` | PASS，29 个 Go 包边界 |
| `pnpm exec redocly lint contracts/openapi.yaml --config contracts/redocly.yaml` | PASS，保留既有 EventV1 未引用警告 |
| `node scripts/validate-examples.mjs` | PASS，291 个契约示例 |
| `git diff --check` | PASS |
| 实际服务无 Codex 注册基线 | PASS，17342 空注册表，无模型派发 |
| F 注册前真实浏览器基线 | F 于 `2026-10-14T12:53:45Z` 报告两尺寸 session 200、空 registry 与浏览器 ready |
| 实际注册前 HTTP 权限检查 | PASS，无 session、跨 Origin、未授权 bearer、缺 CSRF 均 403 |
| 原生派发前 SQLite 独立只读计数 | 实际 events=3、signals=0、nodes=0；项目 / Works 观测 |
| 同 URL 追加登记与正常重启 | 实际执行一次，原始 `registered_at=2026-10-14T13:23:34.0366105Z` |
| 注册起点 → 原生事件首次 DOM | FAIL：F 两尺寸首显示 `13:33:25.152Z/.153Z`，591116/591117 ms，大于 180000 ms |
| 原 dispatch 收据创建 → 原生事件首次 DOM | PASS：创建 `13:33:10.6363218Z`，约 14.516/14.517 秒；独立于注册起点 FAIL |
| 首次实际原件与提交 ID | 首次检查 PASS：15 个完整行、105564 字节、15 个原档 origin 对应提交 ID；全部 18 个 HTTP 事件 GetEvent 一致 |
| 首次实际文件范围 | `snapshot_complete=true`、尾行 / 剩余字节 0、unknown 5、errors 0、`session_complete=false` |
| 当前原件身份与采集范围 | FAIL：实际 nlink=2，`source_identity_unverifiable_or_aliased`、`snapshot_complete=false`；现有已提交 ID 保留 |
| 实际回答与结束轮次 | `gpt-6.1-sol`，assistant 1、task_complete 1、已报总 tokens 8827；没有重复原输入 |
| 实际权限、分页与 GET 只读 | PASS：四项 HTTP 403、五项实际文件身份 / 权限拒绝、全分页、canary 排除、静默 GET 事件数不变、持久游标相符 |
| 实际原件中的工具与策略 | 当前文件工具调用 0、`never/read-only`；旧 owned PID / 启动参数未当时采集，effective tool set 保持 unknown，不能仅用 tool 0 填禁用证据 PASS |
| 连续追加 / 明确 Stop 后范围 | 未执行，原控制器已在默认 180 秒后实际退出，旧输入没有自动重发 |
| 当前 Foundation / Projects | 实际服务 GET 均 200；运行源码包含 R `d14b08dc50014baf9ee636543564b040112cf3b5` |
| 旧任务人工 done | 实际 200，operation `32c7f5a0-0cb0-4de2-91aa-3e7b9bc0ad07`，收据 accepted；目标验收 FAIL |
| done 后实际 read | 首次 409，operation `5f99133e-390f-4987-b8f2-2889716921ff`，收据 `delivery_unknown`；没有重试 |
| 旧任务单次合法 release | 实际 200 / accepted，operation `ccf54824-6194-4161-81c0-0d6985ee5171`；独立 SQL 核 `released`、`ActiveTaskID` 空、Prepared 保持 true、pending action / receipt 0 |
| 旧实际服务正常关闭 | PASS，1909.08 秒 / exit 0；原真实进程退出、read unknown 和收据身份保持 |
| 主装配显式保护源 | SOURCE `b247dfb12b938739f09acc73e6b7582376951a0c`；解析在应用目录和迁移前执行，参数传给 Reader |
| 精确发布组合构建 | `go build -mod=readonly -p=1 -overlay=dist/l1-live/reader-overlay/overlay.json -o dist/l1-live/reader-overlay/server.exe ./cmd/server` PASS；overlay 为 P `579a1cede98b47b08a3126f5d3974da938348929` 的五个精确 Reader 文件 |
| 保护源与寿命配置目标测试 | 四项 PASS：真实 SQLite 显式来源、受保护目录、实际 snapshot hardlink、启动寿命边界；原编译首 FAIL 保留本机 `first-test-compile-failure.log` |
| 新独立服务 17344 | 实际启动和正常关闭 PASS，60.32 秒 / exit 0；真实 GET agents=0、tasks=0，无注册文件、无原生会话文件 |
| 新 scope 原生 / Today | 原停止窗口没有注册或派发；当前已授权语义通过后执行，实际注册、模型、浏览器与三分钟验收仍 NOT_RUN，待全部前置条件和模块单窗口授权 |

实际证据检查逐原档 origin 查询 `attention_engine_events` 并通过 `EventRepository.GetEvent` 定位提交 ID，独立比较 HTTP 六字段、分页和 canary 排除；核对重读及持久游标、追加与停止后的真实文件范围。实际 HTTP 权限检查包含无 session、跨 Origin、未授权 bearer 和缺 CSRF；静默 GET 前后事件数量必须一致。模型名称与工具调用数量仅从选定实际文件生成安全聚合。

## 剩余限制

本报告为实际已执行范围的阶段报告，H 保持同一活动 Task 继续真实闭环。F 是主任务唯一产品 UI 操作人；H 负责注册文件与正常重启，并只读核对原档、SQLite 和 HTTP。原 task 为 `0c6522bc-b232-4040-8407-745885644791`，仅确认一次。HTTP 503 与 dispatch `accepted + provider_unavailable` 原收据保留；实际安全原因是 `thread/read / -32603 / thread_rejected`，更具体的服务内部原因未知。之后原 observer 实际取得正选归档，原件回答成功不覆盖首 503，不授权重发。

F 首 DOM 三个 ID 均在真实 SQLite 定位，其中 `MBV3V2HETK5753LFVJY4PGVJWL` 和 `HFZSCX6TFXYYMOU5VBUNG65QIV` 分别是原档 `task_started` 与 `native_session_header`，第三个是 task 级 `turn_started`。安全证据为本地忽略目录的 `committed-public.json`、`initial-actual-evidence-derived.json`、`initial-actual-model-derived.json` 和 `first-dom-sql-derived.json`。14 个原档记录的 tokens 为真实 null，未填值。原 controller 文件路径只保存在本次 relay.sqlite 的私有绑定中。

人工 done/release 使用用户确认的 `observe` 权限，实际事件出版重核使用 `reconcile`。当前旧 task 已 released，独立 SQL 检查确认进程真实退出、Prepared 工作区和 terminal 归属一致、pending receipt / action 均为 0，工作区 `ActiveTaskID` 为空，已解除占用。read 的 `delivery_unknown` 收据与任务条件保持；工作区清理不表示目标或来源验收通过。原生回答的目标内容未获合法核实，目标验收 FAIL；当前身份守卫拒绝的文件不绕过读取。F 首 release 脚本对 Prepared 的错误期望保留首 FAIL，本次 SQL 结果核对现有正确释放语义。

R 通过实际 Windows 文件名称接口与 `samefile` 核对两个名称，分别位于默认 Codex sessions 与 Orca 账户 sessions，均不在 H / R 验证目录。创建者未知，镜像解释仅为推断，没有删除链接或放宽守卫。完整路径仅在本机私有证据中，安全聚合为 `source-identity-derived.json` 和 R 的 `owned-archive-link-safe-aggregate.json`。

主装配通过显式 `ASTROCYTE_WORKBENCH_PROTECTED_SOURCES` JSON 数组传递保护来源，默认空数组，不隐式选择私人来源。Reader 使用发布候选精确文件的只读 overlay 验证，未将 B 冻结域历史整体合入。组合构建和适用配置检查通过不替代当前 H 分支完整测试；测试 SOURCE `7890384` 的原域消费由 B 确认。后续 R 寿命、EOF 与 owned 执行策略端口不反填旧进程证据。

新独立验证目录为本机 `verification/h-production-next-20261014-2310`，中间文件位于 H 被忽略的 `dist/l1-live/next-20261014-2310`。用户停止指令收到时未注册或派发，随后正常关闭。安全计数与关闭结果保存在 `closure-derived.json`，不会替代明日实际会话验证。内部执行策略采集复用原 Registry 的 `OwnedRelayExecution` 与 SQLite 实际 session，保存安全派生 PID、启动策略、寿命和首 RPC 摘要，不返回 HTTP；新 scope 没有实际模型，执行元数据采集仍 NOT_RUN，effective tool set 不推定为已验证。

完整默认 check/build/E2E、合并前端、Windows ZIP 解压启动和正式发布由对应 owner 与集成验收负责。H 的静态构建和契约检查不能代替实际 Today DOM、完整会话、终端编码或正式部署。

## v1.4 采集与状态边界

已全文读取 `docs/research/orca-rea-20261014.md`，按当前正式 `Works-Relay-v1.4.md` 消费批准的事实和 handoff 接口。下表记录采集层验收边界；working 严格开始证据门禁仍待用户答案。人工完成已确认 `verified_complete` 与 `user_confirmed_complete` 两通道：权限成立且操作者明确确认目标完成时，accepted / working 均可进入 done；verified 另须本任务合法 owned 输出及固定目标实际检查通过。started 独立保存且永不回填，资源独立核退出、pending、归属和权限后沿 done → released。旧记录缺少完成通道时保持未记录，当前真实双通道 HTTP / UI 正向仍待准确组合验收。

| 方向 | 当前 H 接口与采集边界 | 安全事实与关系 | 真实验收 |
|---|---|---|---|
| 投递阶段 | `events/relay_session_codex.go` 识别带原 turn_id 的开始 / 完成记录；通知保留序号。六字段事件不扩大为 operation DTO，采集层不判定 working | 安全 facts 表达原 operation 的 written / accepted / started / unknown；收据和后到事实各自保持。实际目标尚缺 operation 到原 turn_id 的私有准确对应 | 新窗口未运行；开始规则仍待答。没有开始证据时不从普通活动、Git 或 stdout 推断，未知收据保留占用且不重发 |
| 原生会话归属 | archive binding 保留原 task / terminal / session / worktree / project / native identity / format 与文件身份、位置；当前 alias FAIL 保留 | R 私有 controller / startup operation 与 history 已准确消费；GET facts 独立返回原 Registry 缓存 current_ownership，历史不授当前归属 | 实际无控制器的 unknown / GET 只读与权限拒绝通过；新有会话正向、冷重启仍 NOT_RUN，原档冲突立即失败、不删未知链接 |
| 初始化与消费 | 受信任登记保持 drive_mode / input_mode，原 driver 的初始化、session 确认、管道写入 / 关闭与消费 unknown 分项记录 | 闭合 facts 提供来源、原 operation、状态和时间；GET 不探测 ready，不将 input_accepted 或 ownership 推定为可再送输入 / RPC 可用 | 新真实初始化、消费和工具关闭证据仍 NOT_RUN；EOF 与 stdin 写入不推定模型已消费，缺能力具体报告 |
| 恢复与交接 | 原档以 binding / file identity / offset 去重；通知新连接以原 controller / startup operation 独立 UUID 来源和 sequence 去重，旧 checkpoint 留本机 history | R 提供 new / resume / handoff / unknown 关系，声明与真实 session 确认独立；获准 handoff 字段要求明确来源和人工上下文确认，fork 真实拒绝 | 无新真实窗口；恢复 raw gap 保持 `native_resume_raw_events_unavailable`，session_complete=false，不能用通知验收覆盖 archive FAIL |

归档与通知的完整性分别验收，回合完成不封存整个 Session。当前获准的真实 CLI → 三分钟 Today → L1 验收需要语义 SOURCE 和实际组合通过、新独立 scope、F observer 先 ready、真实工具关闭的启动证据以及模块单窗口授权；当前没有启动该窗口。verified 完成须在本次合法输出和实际目标检查成立后记录，人工完成保存明确确认出处，资源释放另行核验。

实际用户作品、自由工具授权、默认活性窗口仍未决定。原生文件的完整前缀、永久追加性与恢复 raw 通知接口不足时，保持具体 unknown/partial。原始 Native 共享源码、mapper、engine、schema/迁移、七类权重及快照流程没有 H 修改。V4 协议、库和运行资料没有读取或修改。管理员 Procmon、两次实际 00:30、正式认知验收和现有 17321 部署均未由本次 H 验收执行。
