# S3 十轮任务产品记录

## 第 01 轮范围与基线

主控任务为 `C:/Users/DW/orca/Astrocyte/tasks/s3-paper-ten-cycles-20261019.md`。本轨 Task `task_51276f272ff3`、Dispatch `ctx_e49721d7ac8d`，沿原工作树、分支和 Draft PR #2。精确基线 `7ce3a5efe312f4cdac3491293e002f5559f076b0`；实际客户端 Codex 0.162.0，Orca 原终端显示 `GPT-6.1-Sol xhigh fast`，继承配置，没有模型覆盖或子 Agent。

第 01 轮修复实际失败的初始化与迁移检查，第 03 轮补齐发送前租约/用途的当前产品集成，两轮均已获独立限定接受。当前按主控明确方案准备当日 Windows Release；第 05/07/09 轮等待主控逐轮明确。S3 调用单位、执行权限证据覆盖和受限 Codex 三项答复仍待定；Claim 拒绝与 Known=false 保持，未启动原生控制器或模型输入。

## 实际 CI 错误与修复方案

准确业务源码 `982333b871520e606c74d553e448c5049f81abde` 的 push run `39155403418`、PR run `39155406045` 最终均 cancelled；取消前 Ubuntu job `117757474469`、`117757483047` 的全仓 `go test -mod=readonly ./...` 已 failure。日志位于原 Reviewer 树 `.verification/dynamic-dag/ci-982-{push-job117757474469,pr-job117757483047}-failed.log`。此前 gofmt、全仓 vet 通过；后续检查未执行成功。

| 原失败 | 准确来源与处理 |
| --- | --- |
| `TestExplicitProjectDiscoveryConfiguration` | PR 的真实 Git HEAD 可读，detached HEAD 没有 symbolic branch。测试初始化改为从当前真实公开 Astrocyte 仓库克隆到忽略目录，显式建立独立分支；保留 Branch 非空、HEAD、提交时间、Dirty 和所有拒绝断言，并核对准确 HEAD/分支。 |
| `TestTasteProfileActualMigrationUpgrade` | 当前嵌入 27 份真实 SQL，旧断言固定 26。源文件精确固定 27 及 026/027 身份；升级并重复运行后，精确核对当前全部迁移 ID 和记录数量，保留原备份、两张品味表和零虚构 profile 断言。 |
| `TestRelayPersistentRealHTTPReceiptsMutexAndRestart` | 原手写 driver/workspace/sink，设置缺少 reconcile，实际返回 403。原断言保留；没有为使替身通过而扩大权限，未作为真实原生验证运行。 |
| `TestLocalProjectPermissionsTokensAndNativeReceiptsSurviveSQLiteRestart`、`TestConcurrentObservationCannotEraseAcceptedDeliveryReceipt` | 原 localNative 计数替身缺真实 message 仓库、controller reader 和 object store，现有生产入口先拒绝 unsupported；日志中的 OR 断言不能证明实际重发。原测试及断言保留，真实 unknown 生命周期仍待合法原生前置。 |
| 明确输出、公开原件、CLI、认知快照缺失 | 按各测试的精确依赖保留阻塞；不伪造来源、不新增 Skip，不据缺前置降低核心路径验收。 |

## 验证与剩余限制

本树 `.verification/run-s3-cycle01.ps1` 串行执行各 phase，记录准确源码、输出和资源双采样。每项 CPU <70%、可用内存至少 4 GiB，间隔 3 秒；BelowNormal，Go GOMAXPROCS=1/-p=1/GOMEMLIMIT=256MiB，Node512MiB，最多一个实际 HTTP listener。运行期峰值尚未测量，内存限制为软限制。

主控通过 `msg_ff3c34073765` 明确认可两项修复和本轮范围。以下结果由同一 Worker 实际执行，原件均在本树 `.verification/s3-cycle01/`。before 为准确 7ce3 的独立真实 Git clone、detached HEAD；after 为本树基线 7ce3 加本轮两个测试文件的最终修改。测试、vet、格式通过后没有再次修改这两个文件。候选准确 Commit 的独立检查由 Reviewer 执行，本轮接受状态不由 Worker 判定。

| 实际命令 | 结果与原件 |
| --- | --- |
| `go test -mod=readonly ./cmd/server -run '^TestExplicitProjectDiscoveryConfiguration$' -count=1 -timeout=2m -v`，before | EXIT1；Branch=nil，函数 0.21 秒。`discovery-20261019-103740-532.log`。 |
| `go test -mod=readonly ./internal/adapters/sqlite -run '^TestTasteProfileActualMigrationUpgrade$' -count=1 -timeout=2m -v`，before | EXIT1；真实迁移 27，旧断言 26，函数 0.15 秒。`migration-20261019-103638-463.log`。 |
| 上述发现命令，after | EXIT0；函数 1.12 秒，包 2.417 秒。真实 clone 保存原 HEAD，明确分支 `discovery-validation`，观察结果为准确 HEAD/分支及 clean。`discovery-20261019-103852-902.log`。 |
| 上述迁移命令，after | EXIT0；函数 0.15 秒，包 1.468 秒。完整旧迁移升级、真实备份、精确 27 条 ID、重复运行、两张表及零 profile 均通过。`migration-20261019-103818-195.log`。 |
| `go test -mod=readonly ./internal/adapters/httpapi -run '^(TestSwarmActualHTTPHumanAuthorizationTimelineAndRevocation\|TestDynamicDAGActualHTTPGrowthProvenanceAndScopedRead)$' -count=1 -timeout=2m -v` | EXIT0，2 函数，包 1.925 秒。生产路由/session guard、实际 HTTP listener/SQLite，授权/CSRF/撤销、固定输入，HTTP timeline 的 event_id 和 DAG 节点的创建 event_id 逐项真实 SQL 回查。`s3-http-20261019-103936-584.log`。 |
| `go test -mod=readonly ./tests/s3 -run '^TestActualApprovalMissionConcurrentIdempotencyRevokeAndRestart$' -count=1 -timeout=2m -v` | EXIT0，1 函数，包 6.067 秒。双实际连接按同一批准并发创建一个 Mission、幂等、重开及撤销；没有认领 WorkItem，不计 AT08。`s3-ledger-20261019-104006-759.log`。 |
| `go test -mod=readonly ./internal/adapters/sqlite -run '^(TestNativeMessageActualTransactionBoundaries\|TestPreparedNativeMessageMissingOriginalReaderDoesNotPublish\|TestSwarmActualMemberPreflightRejectsEnclosingWriteTransactions)$' -count=1 -timeout=2m -v` | EXIT0，3 函数，包 12.539 秒。实际事务/跨库/取消、缺真实 reader 时拒绝准备与发送、无输入对象和原生会话；未启动控制器。`native-preconditions-20261019-104122-849.log`。 |
| `go vet -mod=readonly ./cmd/server ./internal/adapters/sqlite` | EXIT0。`vet-20261019-104424-102.log`。 |
| `gofmt -d cmd/server/discovery_config_test.go internal/adapters/sqlite/taste_profile_test.go`、`git diff --check` | EXIT0，无差异。`format-20261019-104325-625.log` 及实际命令回执。 |

before 与 after 的差异限于测试前置及其准确结果检查，没有改变生产代码或初始化来源的业务含义。after 的调用工作树为 attached；其中被观察的独立 clone 明确建立分支。Reviewer 需用准确候选的真实 Git clone 在 detached 条件下独立复验；纯 Git archive 缺少该测试原有要求的实际 Git 元数据，不是可用前置。

可用前置为本公开仓库的真实 Git 历史、已配置 Go/Node、现有模块缓存、本树忽略的自有绝对输出目录、实际 SQLite 迁移与 HTTP listener。实际 HTTP 数据库保存在 `.verification/dynamic-dag-http/actual-2280340916/state.sqlite` 和 `.verification/s3-http/actual-2437749593/state.sqlite`，服务均由原测试关闭。双连接数据库在 `.verification/s3-approved/actual-3231713368/state.sqlite`；原生拒绝检查数据库在本轮输出内 `accepted-experience-actual-4214265031`、`accepted-experience-actual-3705556120`。未复制用户业务库或正式实验库。

CPU73%、99%、90% 的三次尝试分别在 migration、discovery、vet 的 Go 调用前拒绝；后续相应 phase 在资源恢复后通过。唯一 build 尝试两次采样为 63%/71%，在调用 Go 前拒绝，没有生产构建结果。主控 `msg_8ebeef255306` 确认本轮只有两处测试修改，定向 go test 已编译相应包，无需扩展全仓 build 或反复采样；该 build 保持未执行。没有运行新 SQL 检查、浏览器或全仓测试集合。

完整默认 CI 保持未通过，其余真实前置及旧替身失败未解决；没有新增 Skip、删除失败测试、放宽权限或捕获错误继续。原生正向、AT08/09/13、用户采用与完整 S3 闭环未通过。workflow、正式 V4、根治理、论文保持只读。

## 后续交接边界

PR #1 已审业务 `0642efdf7748a64c36e07e466b5f507b9c04a004` 不是第 01 轮基线祖先，该基线尚未包含其发送前租约/用途检查。该实际业务缺口已交主控，并由 `msg_6ccfb3e220ab` 确认在第 01 轮接受后另行实施第 03 轮；第 01 轮没有自动合并旧治理或业务。

第 01 轮沿原 PR #2 普通 push，提交使用 `[skip ci]`，保留默认 CI 原有失败状态和全部断言。本轨 Task 保持运行，尚未发送 worker_done。

两项测试的源码候选为 `b21f5b66852a7fa8692185ebfaa34ddaae1f2e00`，已普通 push，`git ls-remote` 核对一致；后继报告提交只补准确来源及校正验证范围措辞。该源码除本轮两项测试与记录外，没有生产变化。

独立 Reviewer 报告 `0bd44e135a0b415b431433a6cbd2075e35a4d131` 接受第 01 轮两项修复；在独立真实 Git clone 的相同 detached 条件下重现 before 两项失败、after 两项通过，定向 vet/格式/空白通过。真实 [CR 5566519259](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5566519259) 为 COMMENTED，准确指向 `dacc742e7ea9f0cc7ff011f0df8343caa976a8a6`。hook `msg_9137a34ddfd6` 已到达，主控完整读取、实际核 CR 并以 `msg_0bd7db71f620` 回复；接受轮数由主控登记为 1/10。该接受不扩展到完整 CI、原生正向或 S3 闭环。

## 第 03 轮：发送租约业务集成

主控 `msg_1b27066e82f6` 在第 01 轮接受后授权本轮。实施基线为 `dacc742e7ea9f0cc7ff011f0df8343caa976a8a6`；来源为已审 PR #1 最终 `75cadf95073d5608fb67f0a4b5be5d65e9b1d8ee`，其业务源码为 `0642efdf7748a64c36e07e466b5f507b9c04a004`，共同基线 `f17a491cdd6392829094c95a3e5efb6011362bc0`。0642 至 75cadf 的差异只有旧标准记录；本轮使用文件编辑工具移入 16 个业务/测试路径的必要差异，不导入该分支 AGENTS 或旧报告，不合并公开 PR 或主分支。

| 范围 | 实际改动 |
| --- | --- |
| `cmd/server/relay_swarm_assembly.go`、`relay_swarm_assembly_test.go` | 在现有同库 Swarm/Workspace 装配中注入实际 authorizer，保留缺准备依赖的真实 HTTP 拒绝。 |
| `internal/adapters/agents/relay_prepared_member.go` | 认领来源必须为原始显式 Swarm 准备，核固定 input/scope/binding，首次认领前不得已有租约。 |
| `internal/adapters/sqlite/local_native_message_{lease,scope}.go`、`local_native_messages.go`、`swarm_native.go`、`swarm_native_message_authorization{,_test}.go` | 私有准备保存用途与首次租约引用；Claim 同事务绑定、最终发送授权复查、历史绑定/批准用途核查、同库事务/取消/缺来源/撤销重开检查。 |
| `internal/swarm/app/{commands,native_send,prepared_member}.go`、`internal/swarm/domain/native_send.go` | 认领要求实际 binder，并在原事务绑定；发送逐次核当前 Mission/WorkItem/Grant/holder/fence/operation/预算预留/写域/截止时间。 |
| `internal/workspace/app/native_message_authorization{,_test}.go`、`native_messages.go` | 显式 project/swarm 准备用途，当前授权和单次 send_attempted 占用同事务；外部 Send 在事务外并携带最早绝对截止时间，未知保持原 operation、不重发。 |

静态来源比较结果：16 个路径相对 0642 的差异只剩当前产品的两个既有贡献。`commands.go` 保留 `CreatedBy/CreatorKind` 和 `ValidateDAG`；`swarm_native.go` 保留 `ReadWorkerParticipation`、独立退出观察及明确缺失分类。其余 14 个路径与已审来源一致，领域、生命周期、事件与前端的既有 DAG 改动保持。没有更改 DDL、OpenAPI、原生程序或调用计量/权限的待定选择。

本轮业务源码固定为 `4c4fa1eab003c1daece0ed9e180e2f39a30f5704`，代码与原测试通过文件编辑工具移入后普通 Commit，没有执行 Git 回滚或业务分支合并。主控 `msg_747413ea9005` 明确允许以 today 下真实 Git clone/detached 副本验证；`msg_7b1d47474027` 在科研第 02 轮实际退出后授予本轮唯一窗口。验证副本为 `D:/Astrocyte-agent-assets/20261015/release-package/today-20261019/s3-cycle03/source-4c4fa1eab003c1daece0ed9e180e2f39a30f5704`，每个阶段核完整 HEAD、真实 detached 和干净状态。没有为输出路径修改原测试。

### 实际命令与结果

执行入口为本树忽略的 `.verification/run-s3-cycle03.ps1`，新日志和资源样本统一位于 `D:/Astrocyte-agent-assets/20261015/release-package/today-20261019/s3-cycle03/logs/`。主控 `msg_7578d18f5d2a` 明确允许复用此前实际使用的 `D:/Astrocyte-agent-assets/20261015/go-build/backend` 编译缓存；没有复制缓存或为了轮次冷编译。新 temp、显式数据库输出及副本内相对输出全部位于本轮 today 子目录。

以下 Go 测试均实际以 `go test -mod=readonly <package> -run '<pattern>' -count=1 -timeout=2m -v` 执行，每次只运行一个 package；函数数目不含嵌套子项。

| package 与 pattern | 结果 | 原件 |
| --- | --- | --- |
| `./internal/workspace/app`，`^TestNativeMessageExplicit` | EXIT0，4 函数，包 1.760 秒；明确用途、私有 JSON、原准备身份和调用者规则。 | `rules-20261019-114601-225.log` |
| `./internal/adapters/sqlite`，`^(TestNativeMessageActualTransactionBoundaries\|TestPreparedNativeMessageMissingOriginalReaderDoesNotPublish\|TestNativeMessageExplicitJSONRules\|TestSwarmNativeMessageActualAuthorizationLifecycle)$` | EXIT0，4 函数，包 2.646 秒；实际同库/跨库/取消、缺原来源、成员授权/撤销和同库重开，未生成原生来源或输入对象。 | `sqlite-20261019-114651-443.log` |
| `./internal/adapters/agents`，`^TestRelayPreparedMemberRejectsMissingOriginalSources$` | EXIT0，1 函数，包 1.004 秒；真实未打开 Registry 的缺来源及取消拒绝。 | `registry-20261019-114803-681.log` |
| `./cmd/server`，`^TestRelaySwarmActualPreparedAssemblySQLiteHTTP$` | EXIT0，1 函数，包 13.464 秒；实际装配/HTTP、身份拒绝、缺原准备的发送 501 和无部分执行记录。 | `http-20261019-114857-326.log` |
| `./internal/swarm/domain`，`^TestDynamicDAG` | EXIT0，2 函数，包 5.113 秒；依赖结构和原停止边界。 | `dag-domain-20261019-115021-106.log` |
| `./internal/swarm/app`，`^TestDynamicDAG` | EXIT0，2 函数，包 4.715 秒；明确缺失/结构错误分类和 unknown 用量边界。 | `dag-app-20261019-115123-183.log` |
| `./internal/adapters/sqlite`，`^TestDynamicDAG` | EXIT0，2 函数，包 13.111 秒；原编码语义及真实缺 Registry 分类。 | `dag-sqlite-20261019-115215-350.log` |
| `./internal/adapters/httpapi`，`^TestDynamicDAG` | EXIT0，3 函数，包 18.668 秒；真实 HTTP/SQLite 增长/权限/来源、无 Worker 停止和过期后人类读取。 | `dag-http-20261019-115304-870.log` |

| 其他实际命令 | 结果 | 原件 |
| --- | --- | --- |
| `go vet -mod=readonly ./...` | EXIT0，无诊断。 | `vet-20261019-115410-697.log` |
| `go build -mod=readonly ./...` | EXIT0，全部 Go 生产包编译通过；没有运行模型或原生控制器。 | `build-20261019-115526-882.log` |
| `node scripts/check-architecture.mjs` | EXIT0，43 包满足 Domain/app/context/transport 边界。 | `architecture-20261019-115621-250.log` |
| `gofmt -d` 本轮 16 个 Go 文件 | EXIT0，无格式差异；路径由准确 dacc→本轮的 Go 差异取得。 | `format-20261019-115703-702.log` |
| `git diff --check`、验证副本 `git status --short` | EXIT0，无输出。 | 实际命令回执 |

所有 19 个定向函数通过。13 个阶段的 26 个启动样本最高 CPU80%、最低可用内存 6.5458GiB；Stopwatch 记录单调间隔，同时保存 UTC。所有启动间隔满足 2–5 秒，CPU<85%、可用≥4GiB、BelowNormal、Go GOMAXPROCS=4/-p=4/GOMEMLIMIT=512MiB、Node heap=1024MiB。采样不代表运行期峰值，内存配置为软限制；没有据包耗时宣称性能收益。

实际 SQLite 拒绝/授权数据库位于本轮 `evidence/accepted-experience-actual-{1641144588,1133109175,3604946547}/`。三个 DAG HTTP 数据库位于准确副本 `.verification/dynamic-dag-http/actual-{4262388596,3397770852,211109139}/state.sqlite`。全部为真实测试服务产生的独立库，没有复制用户库或正式 V4。所有检查调用与自有 HTTP listener 已退出，副本保持干净，无浏览器或原生控制器；已用 `msg_8df78aed44b4` 明确归还唯一重操作窗口，后续只更新报告、普通 push 和 PR。

### 接受范围与剩余限制

本轮实现补齐已审发送租约业务在当前产品中的实际集成缺口，适用拒绝/事务/装配与 DAG 回归、Go 类型和生产编译已由 Worker 运行。独立 Reviewer 报告 `b410c9bd87349b7f13f3d913ccc48d06996c5839` 和 COMMENTED [CR 5567006632](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5567006632) 对准确候选 `3267249922082cb1808b6529a2f04084d40c1f92` 限定接受；独立 19 函数、7 包 vet、43 包架构、格式与 7 个真实 event_id 逐项 SQL 核查通过。hook `msg_d9e88ad0d21c` 已实际到达，主控读取并以 `msg_91414a0f0023` 回复，第 03 轮接受登记完成。真实正向 Prepare→Claim→Send、AT08 两成员争用及旧 fence 拒绝、AT09 原生接续/交接、AT13 unknown 对账、真实进程退出与释放、用户采用和经验再次使用仍缺相应合法前置，未计通过。

## 当日 Windows Release 子任务

主控 `msg_962269134dd8` 授权从准确 `3267249922082cb1808b6529a2f04084d40c1f92` 静态修改 `release/**`、`docs/release/**` 和本记录；同一 Worker、工作树、分支与 PR #2 继续负责，正式发布由主控执行。产品 05/07/09 门禁保持，包构建和重复验证不自动计为十轮改进。

当前 Build/Verify 具备重操作阶段的真实 CPU/内存双采样、BelowNormal、Go4/`-p=4`/512MiB 和 Node1024MiB，并恢复原进程设置。新 temp、下载和输出进入 `D:/Astrocyte-agent-assets/20261015/release-package/today-20261019/` 下的独立目录；工具缓存复用既有获准 GOCACHE。Verify 必须接收准确 Source，保留原 ZIP/许可/完整隐私与真实 HTTP/SQLite/认知断言，另核实际未配置 OpenViking。旧 runtime-sanitized-third、公开来源及旧 ZIP 只读。准确源码 `15d74da3edcee26550a71120370fda9c86230e04` 已完成下述新包构建与130项解压运行；DOM检查结果和待验范围分别记录。

主控 `msg_351fa9bca129` 明确要求包验证按实际监听归属保护隔离数据。当前 Verify 每次启动前重新确认端口空闲，使用 Windows `MSFT_NetTCPConnection`、`Win32_Process` 和持有的 `Diagnostics.Process` 核对监听 PID、父子进程及创建时间；根进程和监听均须仍存在。首个 health 与所有后续 HTTP 请求前后复查，无法确认或异主直接失败，不探测其他服务、不自动换端口。成功启动只记录本次 PID/归属；原重启与数据断言保持。未构建的源码 `21a92dc6b97fa5635cb12bb9eb93cd3f643e966b` 保留历史，新包以该修复后的准确源码为准。

按主控 `msg_d3afdea3cd29` 的静态检查授权，PowerShell `5.1.26100.9711` 和 `7.6.6` 分别运行 `.verification/check-release-syntax.ps1`。两者对 Build、Common、Verify、Start 和 Publish 五个实际脚本的 `Language.Parser.ParseFile` 全部 EXIT0；原件位于本日资产根 `syntax-release-ps51/syntax.log` 与 `syntax-release-ps7/syntax.log`。`git diff --check` 无输出。静态解析与实际运行分别记录，后续实际失败和修复范围见下文。

监听归属记录使用标准 `List<int>.ToArray()` 保存实际父子 PID 数组。最终脚本在两个 PowerShell 版本分别再次解析 EXIT0，原件为本日 `syntax-listener-ps51/syntax.log` 与 `syntax-listener-ps7/syntax.log`；没有执行待验脚本正文。

首次真实 OS socket 检查使用源码 `3e52cc13caa934d2514e74d7ffcc83a350381531`，资源 6%/16%、5.2537/5.1360GiB、2.686 秒、BelowNormal。进入实际 socket 场景后，PowerShell 5.1 返回 `VariableIsUndefined: listeners`；原件 `l3e52cc13/listener.log` 和资源 JSON 保留，后续 HTTP/Build/inspect/Verify/浏览器均未执行。全部自有 socket/PS 对象退出，17887 无监听，已通过 `msg_67b42dcc2d3c` 归还窗口，主控 `msg_d30c1e16b804` 确认。

实际标准 PowerShell AST 核查定位到 Common 的文件解码：5.1 的 `ParseFile` 将中文注释与下一行 listeners 赋值一起读取为 Comment，而显式 UTF-8 的 `ParseInput` 保留赋值；两者均没有语法错误。Common 补充 UTF-8 BOM，与其他四个原脚本的既有编码一致，StrictMode、实际监听与全部业务断言保持。静态语法检查同时比较实际 `ParseFile` 与标准 UTF-8 `ParseInput` 的完整源码，避免仅凭零语法错误判断源码被完整读取。

修复后 PowerShell 5.1 的两个标准解析入口均识别第 39 行 listeners 赋值，中文注释仅占第 38 行。两个 PowerShell 版本的五文件语法与完整解码比较全部 EXIT0，原件为 `syntax-utf8-ps51/syntax.log` 与 `syntax-utf8-ps7/syntax.log`。以下实际复验使用包含该修复的准确源码。

### 实际依赖恢复与新包构建

本节全部原件相对于 `D:/Astrocyte-agent-assets/20261015/release-package/today-20261019/`。主控 `msg_ae3f52e051cd` 授予准确 `15d74da3edcee26550a71120370fda9c86230e04` 的实际检查窗口；依赖缺失后的恢复窗口由 `msg_b81911d2b818` 明确授权。没有改变源码、固定版本、断言、内存配置或资源门限。

| 实际命令/阶段 | 结果与原件 |
| --- | --- |
| `powershell.exe -NoProfile -ExecutionPolicy Bypass -File .verification/check-release-listener.ps1 -Output .../l15d74da3 -Port 17887` | EXIT0；真实持有 OS socket 被接受，另一个实际 PS 进程树被拒绝，HTTP 请求 0。`l15d74da3/listener-result.json`、`listener.log` 与资源 JSON。两个持有进程均退出；此项不替代真实 Go 子进程的包启动验收。 |
| 首次 Build，Runtime=`.../runtime-sanitized-third`、Output=`.../p15d74da3` | EXIT1，真实 `runtime-check` 发现 `runtime/sdk/node_modules/laya-ts/package.json` 缺失；Node24.16.0。资源 31%/15%、4.8321/4.9361GiB 允许后，校验直接失败。尚未源码归档、安装、编译或生成 ZIP。`p15d74da3/build.log`、`resources-runtime-check.json` 保留，未移用历史 runtime 通过结果。 |
| 只读旧 ZIP 元数据和固定锁核对 | 标准 `.NET ZipArchive` 读取旧 `b1017c/Astrocyte-v0.1-windows-x64-41cefd00706a.zip`，SDK1370文件/377214131B；实际 package309B、lock20726B 与保留 runtime 相同。五项 SDK 版本、Node、十二项模型及两个 descriptor、许可和 NOTICE 原件均存在。首次 PS5.1 读取 npm lock 空键的 JSON 失败保留；固定 lock 改为完整原文比较，仅普通 package 元数据交标准 JSON 解析，实际 EXIT0。 |
| `.verification/restore-release-runtime.ps1 -Output .../r15d74da3` | EXIT0；成熟 `.NET ZIP` 仅恢复旧包的 runtime/legal/NOTICE，共1400文件/3484759644B，逐项允许项与新目标包含性检查，不覆盖旧输入。实际 package/lock 全等；固定 Node/ONNX binding/Laya/CodeGraph/RPC/SQLite、十二项模型身份与 descriptor 检查通过，没有执行推理。runtime 全扫描1388文件/3484313396B，private_matches=0。`r15d74da3/restore.log` 与三阶段资源 JSON。 |
| `powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Build-Windows.ps1 -Source 15d74da3edcee26550a71120370fda9c86230e04 -Label first -Runtime .../r15d74da3 -Output .../p15d74da3-2` | EXIT0；准确源码归档，官方 registry 冻结安装、前端 lint、三 Go 程序、tsc/Vite、snapshot 第四 Go 程序编译通过。准确源码 web、两个 adapter、启动与快照脚本、LICENSE/NOTICE/说明进入新包，旧二进制和旧 web 未复用。全包扫描1442文件/3548860672B，private_matches=0；全部 ZIP 成员与空间检查通过。`p15d74da3-2/build.log` 及13阶段资源 JSON。Vite 大分块提示保留。 |

构建使用实际 Go1.27.2、Node24.16.0、pnpm11.27.0。13 个构建阶段的26次启动采样最高 CPU27%、最低可用4.5950927734375GiB；各阶段 CPU<85%、可用≥4GiB、Stopwatch间隔2–5秒、BelowNormal、Go4/`-p=4`/512MiB、Node1024MiB。没有测量运行期峰值或宣称速度收益。

准确新 ZIP 为 `p15d74da3-2/Astrocyte-v0.1-windows-x64-15d74da3edce.zip`，1949417822 bytes；源码保持 `15d74da3edcee26550a71120370fda9c86230e04`，本记录属于后继报告，不改变包来源。主控 `msg_6d5dd47e93cc` 安排 Build 后插入科研06，Worker 在所有构建/扫描/压缩调用退出后以 `msg_a481f570c6b0` 立即交回窗口；主控实际读取 Build 日志、独立核 ZIP 字节后以 `msg_80bb71a45604` 确认。构建阶段报告后继 `b022d0e3bf32a224893b46a7d1d9dfe2f69a2455` 仅更新本记录。

### 实际历史认知与解压运行

主控 `msg_bf4bd434ca58` 授予续验窗口，旧工具失败后通过 `msg_475b6d983b19` 明确授权准确源码的既有只读工具及后续原完整 Verify。没有修改产品、旧库、原 history/current、真实日期、Clock 或原断言。

本节原输入根为 `C:/Users/DW/orca/workspaces/Astrocyte/release-cognition-20261014/.verification/cross-day-20261015-0228`：CognitionSource 为其 `app-state/state.sqlite`，CognitionPublicationRoot 为其 `local/astrocyte-workbench/v0.1`，CognitionHistory 为该发布根的 `snapshots/snapshot-1711995690/history.json`。公开 Git 为输入根的 `public-source.git`，固定 ref `b789466ff62f3c9b0e08ec392729f5bbfc5cfaa1`。下表 `.../` 输出均指本日 D 盘根，工具原始命令及参数在对应 transcript 中。

| 实际命令/阶段 | 结果与原件 |
| --- | --- |
| 原 `actual-public-cognition-final.exe --mode inspect`，输出 `q15d74da3` | EXIT1，原输出为 `PUBLIC_SOURCE_VERIFICATION_FAILED`，没有生成 SQL 预期/六字段/准入安全输入或诊断文件。`q15d74da3/inspect.log` 与资源 JSON 保留，后续立即停止并以 `msg_1dd66c4e238e` 交回窗口。只读 `go version -m` 显示该旧程序为6ad1dfa1387d、modified=true，不能从当前源码反推具体失败原因。 |
| 原公开来源的成熟 Git/SQLite 只读元数据 | 标准 Node `DatabaseSync(readOnly:true)`、query_only=1，app-state与current指定snapshot均有1316事件/1316信号/2节点/1次Sleep。Git实际origin为获准公开URL，`rev-parse --verify b789...^{commit}`与`rev-list --count`分别得到准确固定commit与1316。此项只确认元数据，不代完整Reader/来源验收。 |
| `.verification/build-release-inspector.ps1 -SourceRoot .../p15d74da3-2/source -Output .../q-tool15` | EXIT0；准确Source15归档中的既有 `cmd/workbench-snapshot/verify`，Go4/`-p=4`/512MiB和原缓存。`q-tool15/build-inspector.log`、资源JSON；没有源码修改或新解析器。 |
| `.verification/run-release-inspect.ps1 -Output .../q15d74da3-2 -Inspector .../q-tool15/actual-public-cognition-inspect.exe` | EXIT0；原app-state/root/public-source.git固定ref，mode仅inspect，SQLite mode=ro/query_only。完整Git原始来源、生产Reader、SQL10、1316逐事件六字段、排名、分页和原历史断言通过，安全输入实际生成。真实day2026-10-19、today_events=0、minimum_today=0、closed_day为空，没有采集/Sleep。实际验证工具来源为Source15；历史Q组件40880b67e4a668ec0e19d00eb4f77a76fdf210ab另列，不用旧工具来源替代当前工具。 |
| `release/Verify-Windows.ps1 -Zip .../p15d74da3-2/Astrocyte-v0.1-windows-x64-15d74da3edce.zip -Source 15d74da3edcee26550a71120370fda9c86230e04 -Port 17887 -RequireCombined -CognitionSource <原app-state> -CognitionHistory <原snapshot-1711995690/history.json> -CognitionPublicationRoot <原v0.1> -CognitionSafeDir .../q15d74da3-2 -CognitionSourceSha 15d74da3edcee26550a71120370fda9c86230e04 -Output .../v15d74da3` | EXIT0，130项PASS、0项未通过；`v15d74da3/result.json` scope=`candidate-package-cognition-runtime`，准确Source15。ZIP允许项、全量隐私/实际固定runtime、SPA及资产、session/CSRF/跨来源拒绝、真实SQLite写读、OpenViking未配置、空快照/非空历史的包内发布、SQL10/排名/历史、全部1316事件六字段、20/50同主题分页、实际进程停止/重启及持久读取均通过。原130断言完整执行，旧包125项未移用。 |

Verify初始实际监听进程链为45788→35948，重启为47912→48736，均通过Windows实际PID/创建时间核对，每个HTTP请求前后复查。两个根进程/Go子进程及监听实际停止。八个阶段16次启动样本最高CPU48%、最低可用5.303668975830078GiB，满足原资源和单调间隔条件；运行期峰值未测。解压后的全包隐私扫描1442文件/3548860672B，private_matches=0。

所有自有调用退出、17887监听0后，以 `msg_d91f5b3b1caa` 立即交回窗口，主控读取实际result/日志/q安全输入后以 `msg_82499cb58bec` 确认。按 `msg_38c22357ec0d` 插入科研08；独立包CR、完整默认浏览器、原生/S3及正式发布门禁仍保持未通过。已通过工具/inspect/Build/Verify不重复执行。

### 首次实际 DOM 检查

主控 `msg_53137a57405c` 授予同一准确Source15新ZIP的DOM及逐事件SQL检查窗口。实际命令为 `powershell.exe -NoProfile -ExecutionPolicy Bypass -File .verification/run-release-dom.ps1 -PackageRoot .../v15d74da3/extracted/Astrocyte-v0.1-windows-x64 -SourceRoot .../p15d74da3-2/source -Output .../dom15d74da3 -Source 15d74da3edcee26550a71120370fda9c86230e04 -Port 17887`，输出位于本日资产根。检查使用真实公开Git clone、包内程序、HTTP与SQLite；隔离HOME和原生配置，可选OpenViking未配置；没有模型输入、替代响应、截图、trace或视频。

该调用在 `2026-10-19T05:59:35Z` EXIT1。首个失败为检查工具逐HTTP请求执行的 `assertOwnedListener` 子进程 `spawnSync powershell.exe ETIMEDOUT`，实际硬截止时间10000ms；触发于 `verify-release-dom.mjs:105` 的浏览器请求处理，尚未执行该请求的 `route.continue()`。包日志记录人工项目/Mission创建和attention/workspace实际请求，未产生人工工作项、`dom-result.json`或逐事件SQL通过记录。失败没有形成业务断言结论，不能计DOM通过。

本次root为58796、实际开始时间 `2026-10-19T05:55:24.8092192Z`。launcher的finally关闭本次服务树，随后实际查询无本次root、Node、Playwright Chromium且17887监听为0，以 `msg_2ac2442ea32c` 立即交回窗口供08独立复验。原件为 `dom15d74da3/dom.log`、包stdout/stderr、三阶段资源JSON及统一终端session7527实际错误摘录 `failure-observed.md`。没有覆盖失败目录或重试。静态修复建议通过 `msg_35a4ba98c87d` 交主控：只调整自身忽略检查工具的同步执行方式、可观测截止时间和异常传递，保持每HTTP真实归属及全部业务断言；产品Source、成功ZIP及130项结果不变。

主控 `msg_de9cede26a3b` 确认先以一个本次专用PowerShell检查子进程、标准JSON/readline异步串行调用原Common，每次10000ms和原断言保持。两个PowerShell版本的四个相关工具Parser/完整UTF8解码，以及三个Node文件语法检查实际EXIT0，原件为 `dom-syntax-ps51-2` 和 `dom-syntax-ps7-2`。新 `guard15d74da3-2` 真实socket调用EXIT1：检查PS52580在3220.2142ms返回ready，但首个正向请求从 `06:12:48.188Z` 到 `06:12:58.195Z` 耗时10007.6809ms仍超时，检查通过数0。实际 `positive-observations.jsonl` 与stderr均0B，现有证据尚不能定位stdin、Common入口或具体CIM等待，不能认定复用已解决原因。CPU63%/47%、可用6.0552/6.0925GiB、Stopwatch间隔3933.2792ms、BelowNormal的启动采样允许该调用；不代表运行期峰值。

该次没有HTTP或DOM，两个socket根、检查PS和Node全部退出，实际17887监听0后以 `msg_b4a39692ae71` 立即归还窗口。完整客户端时序、实际timeout错误和退出SIGTERM记录保留，不覆盖首次DOM失败。主控随后要求先静态核已装psutil7.2.2与官方对应API及源码；具体等价条件、ISO精度与每次新Process/系统监听方案在本树忽略的 `.verification/psutil-listener-plan.md`，以 `msg_f0b81d7240a0` 提交。第二次失败不提供产品页面、人工DAG或逐事件SQL通过结论。

主控 `msg_f4844bdbfcfc` 全文读取方案后批准实施及唯一窗口，仅修改自身忽略工具。一个本次Python子进程直接import已装7.2.2，每次使用fresh Process与系统TCP表，保留root原ISO小于1ms、首次实际创建时间身份、完整父链时间顺序/循环拒绝和二次监听；没有缓存OS结果或增加超时。正式产品Common及Source15没有更改。Python标准AST、三个Node语法、PS5.1/7四脚本完整解码/Parser均实际EXIT0，原件 `dom-syntax-node-python-3`、`dom-syntax-ps51-3`、`dom-syntax-ps7-3`。

新 `guard15d74da3-3` 实际EXIT0：同一个真实socket的两次正向检查与实际异进程树拒绝通过，HTTP=0。正向checker11448、root58544，真实创建时间ISO差0.00095367431640625ms；客户端每次实测30.6593/26.8913ms，checker正常退出0。异树checker58688返回原时间顺序拒绝并退出1，检查工具按准确异常断言接受该拒绝，没有继续放行；所有socket/检查对象均退出。`guard-result.json`、每次OS阶段/创建时间/父链/完整异常和客户端时序记录保留。该观测仅说明该环境下实际调用通过，不推断旧超时的具体原因或产品性能。

随后 `dom15d74da3-2` 在首项 `public-git-clone` 资源守卫EXIT1：CPU100%/37%、可用5.8228/5.1281GiB、Stopwatch间隔3399.2431ms、BelowNormal。真实CPU样本触发原85%门限，尚未Git clone、包启动、HTTP、浏览器或SQL；没有自动重试。全部调用退出且实际17887监听0，以 `msg_18f497450acd` 立即交回窗口。原资源JSON和transcript保留，等待主控明确续验；成功包和130项结果不变。

### 实际双视口 DOM 与事件来源

主控 `msg_a5b4eb6b6b4c` 在其他轨道全部调用退出后授予一次续验窗口。实际命令仍为 `.verification/run-release-dom.ps1`，PackageRoot=`.../v15d74da3/extracted/Astrocyte-v0.1-windows-x64`、SourceRoot=`.../p15d74da3-2/source`、Source=`15d74da3edcee26550a71120370fda9c86230e04`、Port=17887，仅Output为新 `.../dom15d74da3-3`。同一成功ZIP的包内程序/启动脚本、原完整DOM断言和每HTTP实际监听归属核查保持，没有重建包或重复130项。

实际EXIT0，`dom15d74da3-3/dom-result.json`与`dom.log`记录通过。一个实际Chromium、一个实际包服务，两视口顺序执行，检查完成后全部退出。

| 实际范围 | 结果 |
| --- | --- |
| 1920×1080与1280×720的attention/workspace/swarm/cognition/focus/cognition/evidence | 两组均通过真实标题/导航、无横向溢出、无pageerror；前两路由未加载DAG chunk，进入Swarm后实际按需加载。 |
| 26件品牌资源与10个实际客户端清单条目 | 真实HTTP200、image MIME、非空内容及浏览器解码尺寸通过；UI原资产映射通过，cursor-agent使用cursor.svg。没有执行CLI probe或原生能力检查。 |
| 固定公开README、本地检索键盘与可选OpenViking | 实际公开Git clone/detached/clean为准确Source15，真实人工导入及固定版本正文一致；Ctrl+K/Escape通过。OpenViking未配置且外发选项禁用，本地页面与资料读取通过。 |
| 人工DAG逐步增长及交互 | 实际Mission空图→一个工作项→两个工作项/一条依赖，workers始终空；键盘详情、TB方向、0.8缩放、显式刷新与原API节点/边来源通过。 |
| 当前权限撤销 | 第二视口实际收窄项目授权后旧画布清除，显示API403；实际DAG GET返回403。没有Claim或Send。 |
| 真实SQLite逐事件 | 三次来源文件分别为`dag-sql-1920.json`、`dag-sql-1280.json`、`dag-sql-after-revocation.json`；当前DAG的每个event_id逐项查询原foundation_outbox，类型/aggregate/payload、Mission项目和工作项原id/title/mission/dependencies/created_by/creator_kind与实际API返回一致，节点创建与边事件均存在。 |

本次项目 `f00349c8-9a7a-4af3-a6c3-f80b7ee42285`，Mission `8b7fa023-519f-4343-b234-bd99823cfaf8`。实际事件如下，两个工作项的创建主体均为`local-human`、`creator_kind=human`，没有Worker贡献声明。

| event_id | 原事件类型 | aggregate_id |
| --- | --- | --- |
| `9f11b1d2-d224-4875-a7c7-efc89d376b0c` | `swarm.mission_created` | `8b7fa023-519f-4343-b234-bd99823cfaf8` |
| `7b7e7416-286b-40d6-a3bc-2c750480c400` | `swarm.work_item_created` | `529b725e-6e26-4047-81b4-834a02be6208` |
| `58185773-2413-420a-9b60-711bb390a347` | `swarm.work_item_created` | `2247754f-3cdb-4c67-85e9-fefe655ee4c7` |

430次逐HTTP实际归属检查完成，检查Python49060正常退出0，stderr为0B。实际OS阶段、root/listener创建时间与父链记录为`dom-listener-observations.jsonl`，标准JSON客户端时序和关闭记录为`dom-listener-client.jsonl`。三阶段六次启动采样最高CPU19%、最低可用5.26092529296875GiB，原CPU<85%/可用≥4GiB/单调间隔2–5秒及BelowNormal均满足；没有测量运行期峰值或据此声称速度收益。

所有浏览器/Node/Python检查/PS与自有服务实际退出，17887监听0后，以 `msg_d5bb645587e2` 立即交回窗口，随后仅报告、普通push与PR更新。先前两个检查工具超时及CPU拒绝原件保持；本次限定DOM通过不替代完整默认浏览器或独立包CR。

三项 S3/原生选择仍待答，Claim 拒绝及 `Known=false` 保持；没有初始化原生控制器、发送模型输入、伪造会话/claim/lease、重发 unknown 或新增 Skip。完整默认 CI 仍未通过，未执行其含旧替身的全套测试。前端、OpenAPI、workflow、正式 V4、根治理和论文没有本轮修改。新包构建、限定130项解压运行与上述DOM通过；完整默认浏览器、今日实际新增非空、原生控制器、完整 S3、独立包审查和正式发布尚未通过。历史非空与今日新增分别记录，旧125项没有移作本轮证据。

## 2026-10-19 原应用恢复：依赖准备

本次接续主控 `tasks/restart-product-research-20261019.md`，沿原 Task `task_51276f272ff3`，当前 Dispatch `ctx_013843bce872`、终端 `term_78b3371d-5de3-4a8b-8213-85f280c81054`。实际客户端 Codex CLI0.162.0，Orca 原生终端显示 `GPT-6.1-Sol xhigh fast`，没有模型覆盖或子 Agent。接管 HEAD 为 `57a91f4601b4292df4f55ba6cac541bad5caab3f`，tracked 干净；01/03 已接受，不重复实施或计数。

原应用稳定程序仍为 `c5b3b5715f2e93b82365720914be24c73642bd1e`、modified=false、26187264 bytes，Web 为原 `172bb445597193d55f1563ee5d21ac2ec881798d` 的30文件/1058897 bytes。原运行根为 `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413`，SQLite 位于其 `local/astrocyte-workbench/v0.2-access/app-state/state.sqlite`，接管时1277952 bytes；历史40804进程不存在，64505/17887没有监听。原 Q current 为 `snapshot-1711995690` / `2026-10-14T18:48:42.2097174Z`。接管核查只读文件元数据，未查询原应用 SQLite、请求应用 HTTP 或执行原恢复正文。

原媒体配置引用的 standards-backend/node_modules 已缺失，原 runtime-sanitized-third/runtime/sdk/node_modules 亦缺失。原 H `dist/product-recovery-20261016/restore-38fb.ps1` 保留缺媒体依赖立即拒绝，未在缺失状态运行。主控完整读取自身恢复方案后，通过 `msg_9b016eeda64c` 明确允许恢复两个原路径，四份原应用配置、SDK package/lock、H脚本和历史日志只读。实际唯一依赖窗口为 `msg_95f72d12cd7b`，应用服务启动另行等待窗口。

以下命令使用本树忽略入口 `.verification/restart-product-20261019/restore-dependencies.ps1`，输出均在同目录的独立子目录。两个 PowerShell 入口通过5.1/7标准 Parser；资源入口通过 Python 标准 AST。原 H 脚本未编辑。

| 实际命令与范围 | 结果与原件 |
| --- | --- |
| `pwsh -NoProfile -File .verification/restart-product-20261019/restore-dependencies.ps1 -Phase summarize -Output <本树忽略目录/summarize-01>` | EXIT0。原 standards-backend 的 pnpm11.27.0 执行 `install --filter astrocyte --fail-if-no-match --frozen-lockfile --child-concurrency=1 --network-concurrency=1 --store-dir <本次输出/pnpm-store>`；378项锁文件供应链政策检查和187依赖安装通过，耗时3m52.9s。原 package.json/pnpm-lock.yaml/pnpm-workspace.yaml/.npmrc 完整字节相同，web/node_modules 仍不存在。原精确媒体配置路径的实际 CLI `--version` 为0.25.1。`summarize-01/restore.log`、`pnpm.log`、两组资源JSON。 |
| `pwsh -NoProfile -File .verification/restart-product-20261019/restore-dependencies.ps1 -Phase sdk -Output <本树忽略目录/sdk-01>` | EXIT0。成熟 .NET ZipArchive 读取准确 Source15 ZIP，先逐字节核原 SDK package/lock，再完整恢复其 node_modules 的1368文件/377193096 bytes；全部目标先验证包含关系、重复与未占用，不覆盖已有文件。原 `release/Check-Runtime.mjs` 实际通过 Node/native ONNX binding/Laya/CodeGraph/RPC/SQLite、十二固定资产及两个来源 descriptor；没有执行推理。`sdk-01/restore.log`、两组资源JSON及实际终端退出记录。 |

安装使用实际原有 registry.npmmirror.com，慢请求及包管理器原生 ETIMEDOUT 重试保留；安装进程最终成功，没有脚本级重试或换源。Corepack/pnpm store、npm cache、XDG cache/state、TEMP/TMP 和 Node 编译缓存均显式位于本次自身忽略输出，没有改写宿主设置。SDK来自原准确 Source15 ZIP，包内程序和 Web 未用于更新原应用。

四个阶段的八次启动样本最高 CPU19.8%、最低可用12.798637390136719GiB；实际阻塞CPU采样，结束点单调间隔约3.002秒，均满足原85%/4GiB条件，父子进程 BelowNormal。Go4/`-p=4`/512MiB、Node1024MiB；没有运行 Go 编译，未测量运行期峰值。

全部依赖安装和检查调用实际退出后，以 `msg_5dce3320d058` 立即交回唯一窗口。当前原应用尚未启动；原库逐表计数前后比较、实际服务PID/唯一监听/health/首页及独立恢复审查保持未执行。产品 Source15、ZIP与此前130项/DOM结果保持，新的独立包审查由 Reviewer 分别记录。原生/S3三项待答与Claim拒绝保持，完整默认CI、S3、发布和正式实验没有新增通过结论。

## 开发检查入口的真实输出目录

主控 `msg_b48869ca0bea` 批准本项静态修改，`msg_e35ee1a69e70` 批准本地固定源码，`msg_e6fb63d3ce4a` 授予实际检查窗口。before 为 `2a05bd0d78cbeb5e69ba09728923ea2f2d1d1b65`，after 源码为 `22cde9efc540553c433463af5b9e955790d57a99`；两者通过真实 Git clone/detached 保留准确身份，各阶段前后 tracked 干净。当前任务、客户端、工作树与分支保持上节记录。

原准确982 CI的 OpenViking HTTP 和两个 ProgressPreflight 检查在创建HTTP/SQLite前因缺 `ASTROCYTE_TEST_OUTPUT_ROOT` 失败。`scripts/check.mjs` 原先直接运行 Go 检查，没有设置这项真实输出前置。现在 `scripts/process.mjs` 的 `createCheckOutputDirectory` 复用现有 `verificationOutputRoot`，以 Node 标准 mkdir/mkdtemp 在所选根新建 `check-*`；check.mjs在寻找Go前明确设置本次进程的环境，原子进程继续使用现有继承规则。相对根立即拒绝，目录错误在Go启动前抛出。所有原检查项目、顺序、命令、测试、断言及workflow保持。

实际入口为本树忽略的 `.verification/restart-product-20261019/run-output-check.ps1` 与 `verify-output-root.mjs`。下表目录均相对于该忽略根；两个源码副本分别为 `source-before-2a05` 与 `source-after-22cde9ef`。每个命令显式提供准确 `-Source`、副本 `-SourceRoot` 和新的 `-Output`。Go命令通过既有 `findGo/run` 和实际安装的Go1.27.2执行，Node使用标准模块直接调用同一新增辅助函数。

| 实际阶段与命令 | 结果与原件 |
| --- | --- |
| `run-output-check.ps1 -Phase clone`，before/after各一次 | 两次EXIT0，实际clone后checkout准确提交、逐次核HEAD与clean。`ci-clone-before-01`、`ci-clone-after-01` 的clone.log/run.log及资源JSON。 |
| before，`go test -mod=readonly ./internal/adapters/httpapi -run '^TestOpenVikingActualHTTPUnconfiguredAndRefused$' -count=1 -timeout=2m -v` | 原Go进程EXIT1，精确显示“需要明确的本次验证输出目录”；负向检查确认该预期前置失败，包2.069秒。没有将本项记为功能通过。`ci-before-01/before.stdout.log`、stderr、result.json与check.log。 |
| after，清除输出变量后通过无参数辅助函数分配默认根，再执行同一原OpenViking命令 | EXIT0，原1函数通过、0 Skip，函数0.13秒/包2.415秒。实际输出为准确副本 `.verification/check-DOIgoT`。真实HTTP/SQLite、未配置、真实连接拒绝、空集合、Origin/CSRF、缺内容层级和零远端提交断言保持。`ci-openviking-01/result.json`、check.log及资源JSON。 |
| after，明确本次父根并通过同一无参数辅助函数继承，`go test -mod=readonly ./internal/workspace/app -run '^(TestProgressPreflightActualColdConfigurationAndRestart\|TestProgressPreflightActualSQLiteReadOnlyRefusesPending)$' -count=1 -timeout=2m -v` | EXIT0，原2函数通过、0 Skip，各0.16/0.12秒，包2.342秒。实际输出为 `ci-progress-01/evidence/check-04Syyp`；真实生产配置前置拒绝、保守unknown回执、人工保存、原operation/同库重启拒绝重发，以及实际query_only/READONLY/HTTP500与零回执保持。未Probe Registry、零原生会话和来源README只读守卫通过。`ci-progress-01/result.json`、check.log及资源JSON。 |
| after，`run-output-check.ps1 -Phase invalid-roots` | EXIT0，相对路径和本次实际创建的普通文件均由真实文件系统/原校验拒绝，Go启动0。`ci-invalid-roots-01/result.json`、check.log及资源JSON。 |
| 两个生产Node脚本与验证Node入口 `node --check`，PS5.1/7标准Parser，`git diff --check` | 全部EXIT0；仅静态解析，没有执行待检查脚本正文。 |

六个实际阶段的12次启动样本最高CPU9.4%、最低可用15.298702239990234GiB，实际阻塞CPU采样、单调结束点间隔3.0019–3.0091秒、BelowNormal；Go4/`-p=4`/512MiB软限制、Node1024MiB、真实服务串行。没有运行浏览器或Native，没有测量运行期峰值或宣称速度收益。全部调用实际退出后以 `msg_59a9a6d7962c` 立即交回窗口，随后仅报告、普通提交及push。

本项修复开发检查入口的输出前置。完整默认CI仍未通过；原公开论文/视频/历史SQL和真实原生前置未补造，旧relayControlled/localNative等替身未运行，原生/S3三项选择未代答。原workflow另行直接执行的race命令没有由本项分配输出根，范围单独保留。未运行全套pnpm check，未重建原Source15 ZIP或启动共享64505服务。当前等待独立Reviewer准确复验与CR，不自行宣布接受或增加十轮次数。
