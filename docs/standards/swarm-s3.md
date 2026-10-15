# Swarm S3 / S6 实施与实际验证

## 第十一阶段：恢复检查与固定工作树输入

2026-10-15，Task `task_da9aebf03f83` / 当前 Dispatch `ctx_a34aff42e4d2` 对模块 Task `task_b0197ca7ebdf` / 当前 Dispatch `ctx_c2889e0e477c` 交付 SOURCE `71af7fde6a0a6d0e2f6f23813626d44e26e144e8`。实际客户端与模型由 Orca 公开 worker projection 核实为 Codex / `gpt-6.1-sol`。恢复基线为 `a5d00e764ffa0266bf34067190f72f3463fe9852`，接管时 clean；没有创建原生会话、发送模型输入或开始登记窗口。

前置 SOURCE `a5d00e764ffa0266bf34067190f72f3463fe9852` 增加 `swarm_lease.go` 与真实 SQLite 测试：私有原 native session、adapter、controller 与 startup operation 引用保存在既有租约 JSON 的私有存储结构中，关闭并重开数据库后保持原值。`SwarmRepository.LeaseForOperation(ctx, operationID, sessionID)` 只供受信任 ownership 桥读取原租约；其他 session 被拒绝。公开 Lease JSON、幂等回执和 outbox 继续不序列化这些私有字段。没有迁移或新增表，没有把参数测试写为原生执行成功。

本阶段认领检查复用实际 Git 的固定快照验证：独立且同仓库的 worktree 必须具有指定完整 HEAD、没有已跟踪变化且没有未跟踪输入。新真实 Git 回归测试先确认未跟踪文件会被接受，保留首 RED；修复后该输入返回 `context_stale`，移除本测试创建的该文件后，同一 worktree 与原 baseline 再次通过。实际项目 test/build 使用相同检查，错误摘要只描述固定快照状态。

| 本阶段实际命令 | 结果 |
|---|---|
| `go test ./internal/swarm/... ./tests/s3 -count=1 -timeout=120s` | PASS；领域、真实 S2 批准、并发唯一 Mission、异输入冲突、撤销与重启 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActualGitWorktreeObjectsAndCommittedScope$' -count=1 -timeout=120s` | 首 RED：未跟踪输入获得固定 baseline；没有减少断言或放宽限制 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1 -timeout=180s` | 修复后 PASS，7.123 秒；实际两个 SQLite connection、无半预算、旧 fence、私有 tuple 重启和真实 Git/Go 对象检查 |
| `go test ./internal/adapters/httpapi -run '^TestSwarmActual' -count=1 -timeout=180s` | PASS，2.470 秒；本域生产 router、真实 listener、认证、SQLite 与逐事件核对 |
| `go vet ./internal/swarm/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/s3` | PASS |
| `go build -o D:/Astrocyte-agent-assets/20261015/product/s3/astrocyte.exe ./cmd/server` | PASS |
| `node scripts/check-architecture.mjs`、`git diff --check`、三个修改 Go 文件的 `gofmt -l` | PASS，33 包；格式检查无输出 |
| `git push origin HEAD`、`git ls-remote --heads origin songconmaisaix31-design/product-swarm-20261015` | SOURCE 已普通 push，远端完整 SHA 相同 |

本阶段所有 Go 命令仅对自身进程设置 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`、`GOTMPDIR=D:/Astrocyte-agent-assets/20261015/go-temp/product-s3`。首次协调消息用了 CLI 不支持的 `--type progress`，明确拒绝且未入队；读取实际协议后改用 `status`，消息已持久入队。权威模块与发布文件在模块工作树及根工作树读取，没有复制或改写其他负责人的管理文件。

根当前已确认完成双通道、started 事实独立和资源释放独立审计；执行后端单 Mission 独占的抽象归 v0.3，自动路由归 L3，本域没有自行变更这些架构或替代真实 started。Calls 单位与执行工具/外发证据覆盖仍待用户回答，现有 `Known=false`、nullable 实际用量和 Claim 拒绝保持。H 主服务与 OpenAPI 组合、R 同 Registry 事实适配、两个实际 Agent、实际接续/交接、人类采用与第二任务经验引用仍未通过；这些组件 PASS 不代表完整 AT08/09/13/16 或最终 release 通过。Task 继续，阶段 SOURCE 与本独立 REPORT 供唯一集成人员普通消费。

## H 公共装配的最小业务提交与接口

2026-10-15，Task `task_da9aebf03f83` / Dispatch `ctx_2e078edf896b` 对模块 Task `task_b0197ca7ebdf` / 当前 Dispatch `ctx_9ce12d181768`（原 `ctx_b1cd2e0a39aa`）交付以下装配说明。H `ctx_e7abbaf902da` 的核对基线为 `53d9bb107440d27fc74b70609ecddc3c299574ac`。只读比较确认该基线中的 `migrations/019_workspace_approvals.sql`、`020_swarm.sql`、`internal/workspace/app/proposal_ports.go`、`internal/workspace/domain/proposals.go`、`approvals.go`、`internal/adapters/sqlite/workspace_proposals.go` 与本域已验证版本逐文件相同。H 已有真实 `S2Service.ReadProposal`、`ReadApproval`、`WithAuthorizedSessionBindings` 和 `WorkspaceTransaction` 公开端口。原 SHA 不是其祖先，不能由此推断缺少这些文件，也无需重复消费原 S2/B 全部历史。

该 H 基线还需 B 的 `c110201774dd61b8d7e5280526364786c6af4cc2`：只修改 `internal/workspace/app/project_policy.go` 及 `project_policy_test.go`，将 `reject_artifact` 加入已知操作并要求人类身份。019 原始来源为 B `da79452d1735f8fceca103aaf9d10bb53182f1f5`，020 原始来源为 B `da48a114d91181075bd32b62996beadd9c2c0a25`；所需 S2 公开端口原始来源依次为 `4eec31ddfde39fa3f548691ec6e7f2f34470351a`、`5d0ba735533f68259b91128dc404c8e4c2f2a535`、`507be52f4e31759dc8c3765aeea964f1f2efe013`、`80248d2c7177823d4535e7730453481426a27305`。这些前置已经存在于核对过的 H 文件内容中。

随后按下表顺序消费九个已 push 的纯业务提交。每行 `{a,b}.go` 分别表示同一路径下的 `a.go` 与 `b.go`；`git show --format= --name-only <完整 SHA>` 返回该提交实际文件列表。九个提交全部只修改 Swarm、Swarm SQLite/HTTP 与 S3 测试，不含 report-only、共享依赖、冻结领域或其他业务历史。

| 顺序与准确 SOURCE | 实际修改路径 |
|---|---|
| 1 `b220ab27c4e57d0bfdfaef224a9d03c1ae92366a` | `internal/swarm/app/{commands,contracts,lifecycle}.go`；`internal/swarm/domain/mission.go`；`internal/adapters/sqlite/swarm_{approvals,evidence,store,store_test}.go` |
| 2 `3b78fd04dbf758e745e9880297ee1c2d6e58bfc2` | `internal/swarm/app/{commands,contracts,lifecycle,queries}.go`；`internal/swarm/domain/{mission,policy,policy_test}.go`；`internal/adapters/sqlite/swarm_{approvals,evidence,evidence_test,store}.go`；`internal/adapters/httpapi/swarm_{missions,work_items}.go`；`tests/s3/approval_test.go` |
| 3 `e18c0965708e091b547ca2108579a4b3ca81b72c` | `internal/swarm/app/{binding_context,commands,contracts,experience,queries,review}.go`；`internal/swarm/domain/mission.go`；`internal/adapters/sqlite/swarm_{approvals,evidence,go_checks,go_checks_test,history}.go`；`internal/adapters/httpapi/swarm_{missions_test,work_items}.go`；`tests/s3/testdata/public-stats/{go.mod,stats.go,stats_test.go}` |
| 4 `3c568c7ec8bb13a88b0abf95d44cbb13fc6422b5` | `internal/swarm/app/{commands,contracts,lifecycle,queries,review}.go`；`internal/swarm/domain/mission.go`；`internal/adapters/sqlite/swarm_{approvals,store}.go`；`internal/adapters/httpapi/swarm_{missions,work_items}.go`；`tests/s3/approval_test.go` |
| 5 `ec653ddd90a9c91c83af28a8909fb41741116160` | `internal/swarm/app/{commands,contracts,experience}.go`；`internal/swarm/domain/mission.go`；`internal/adapters/sqlite/swarm_history.go` |
| 6 `93d1f6ba3ebb6ab029ced8ebe7c268a62d5ed2b0` | `internal/swarm/app/{commands,contracts,lifecycle,queries,usage_test}.go`；`internal/swarm/domain/{mission,usage,usage_test}.go`；`internal/adapters/sqlite/swarm_{approvals,store,store_test}.go` |
| 7 `2bd5d26ea6f932d8511afd869ff41f2146c0f94a` | `internal/swarm/domain/{policy,policy_test}.go`；`internal/adapters/sqlite/swarm_{history,store,store_test}.go` |
| 8 `10fa36cdd68b3247a70da9230453b74c0f5a04fa` | `internal/swarm/app/lifecycle.go`；`internal/swarm/domain/{mission,recovery_test}.go`；`tests/s3/approval_test.go` |
| 9 `70e6e5a03fca9a19aa0a70e44b23b78337d34714` | `internal/swarm/app/{binding_context,binding_context_test,commands,contracts,queries}.go`；`internal/adapters/sqlite/swarm_approvals.go` |

完整工厂为 `swarmapp.NewSwarmService(repo, grants, members, evidence, ownership, options...)`，五个必需依赖均不能为 nil。`repo` 使用 `sqlite.NewSwarmRepository(db)`；`grants` 使用 `sqlite.NewSwarmApprovalBridge(db, actualS2)`，必须为相同 `*DB`；`members` 同时提供 `MemberAuthorizer.CheckMember` 与 `ExecutionDescriptorReader.DescribeExecution`；`evidence` 使用 `sqlite.NewSwarmGitEvidence(actualObjects, actualProjectChecker)`；`ownership` 提供原 operation/session 的 `Reconcile` 和 `Stop`。`ServiceOptions{ProjectMaxActive: n}` 是唯一可选配置，只接受正数，缺省为 2，同时遵守 Grant 更小的上限。

固定获准 Go 项目的实际检查工厂为 `sqlite.NewSwarmGoProjectChecks(canonicalRoot, baseline, actualObjects)`，返回值作为上述 actualProjectChecker。只绑定已获准的 Median 项目原根与完整 baseline `da3bdddb085808863090a05f60cc1c64415039cf`；该检查器拒绝其他仓库，按实际进程结果保存固定 `go test ./...` 与 `go build ./...` 对象回执。缺少实际项目检查时，提交不能获得 `ApplicableChecksVerified`。

`Member` 的准确字段为 Purpose、BindingID、OperationID、HolderID、SessionID、ProjectID、Worktree、Version。Purpose 分别为 `claim_work_item` 或 `submit_artifact`；Version 为实际绑定版本的十进制字符串。执行描述必须给出实际 `Capabilities`、`Tools`、`ExternalTargets` 和 `Known`，未知不能写为已知空集。`ExecutionFact` 返回相同原 OperationID/SessionID、StopConfirmed、Outcome、nullable Usage、CallsKnown/TokensKnown 和允许公开的 FailureCode。真实所有权适配缺失时，已实施的 `NewSwarmLedger(repo, grants)` 可提供实际批准和 SQL 账本；缺少原生能力的入口返回 `unsupported_capability`，不登记为原生验收成功。

H 在同一生产认证 router 内调用 `h.registerSwarmS3(mux, service)`，移除同路径的 S0 占位注册，继续使用现有 session/CSRF 与 Agent token 守卫。Principal 来自 `localCaller(r)`，不能从 JSON 注入。全部写入 Command 的顶层元数据为 `schema_version=1`、`request_id`、`expected_version`，`Idempotency-Key` 只从请求头读取。路径 ID 与正文 ID 不同时拒绝；CreateMission 的 `event_id` 只供内部受信任 consumer，HTTP 传入非空值时拒绝。

| 实际 HTTP 路由（均以 `/api/v1` 开头） | 请求 DTO、预期版本与实际响应 |
|---|---|
| `GET /missions` | 经过当前项目权限读取；既有 ListResult |
| `GET /missions/{id}` | `schema_version` 及展开的 MissionDetail：`max_active_executors`、`mission`、`work_items`、`leases`、`artifacts`、`timeline`、`executions` |
| `GET /missions/{id}/work-items`、`/timeline` | 当前 Mission 权限；既有 ListResult，空集合为数组 |
| `POST /missions` | CreateMissionCommand：元数据、grant_id、epoch、proposal_revision、project_id、root；expected_version=0；返回 `{schema_version:1,mission}` |
| `POST /missions/{id}/work-items` | AddWorkItemCommand：元数据、mission_id、title、inputs、dependencies、write_paths、required_capabilities、artifact_inputs；expected_version 为 Mission.version；返回 `{schema_version:1,work_item}` |
| `POST /missions/{id}/pause`、`/cancel`、`/recover` | TransitionCommand：元数据、mission_id、reason；expected_version 为 Mission.version；返回 `{schema_version:1,mission}` |
| `POST /work-items/{id}/claims` | ClaimCommand：元数据、mission_id、work_item_id、binding_id、session_id、member_version、worktree、operation_id、resources、expires_at；expected_version 为 WorkItem.version；返回 `{schema_version:1,lease}` |
| `POST /work-items/{id}/artifacts` | SubmitCommand：元数据、mission_id、work_item_id、fence、commit、object_ref、checks_ref、changed_paths、inputs、usage；expected_version 为 WorkItem.version；返回 `{schema_version:1,artifact}` |
| `GET /artifacts/{id}?mission_id=...&version=...` | mission_id 必需；version 可省略，提供时必须是正整数；返回 `{schema_version:1,artifact}` |
| `GET /artifacts/{id}/diff?mission_id=...&version=...` | 同一读取权限及版本规则；返回 schema_version、artifact_id、commit、base_commit、diff |
| `POST /artifacts/{id}/acceptance` | AcceptCommand：元数据、mission_id、artifact_id；expected_version 为 Artifact.version；返回 `{schema_version:1,artifact}` |
| `POST /artifacts/{id}/rejection` | RejectCommand：元数据、mission_id、artifact_id、reason；expected_version 为 Artifact.version；返回 `{schema_version:1,artifact}` |

`inputs` 为 `{id,revision}` 固定来源数组；`artifact_inputs` 为 `{id,version,commit,object_ref}` 的已实际采用快照数组。Resources 为 `{calls,tokens,money,currency}`，calls/tokens 是非负整数，金额未知为 null；Usage 的 calls/tokens/money 分别使用 nullable 整数，未知保持 null。`expires_at` 使用 Go `time.Time` 的 RFC3339 JSON 格式。Claim 的 operation_id 必须是实际已准备的 input operation UUID；不能使用 startup operation 或新造身份。

本域 `swarmResult` 返回 200 成功；NotFound→404；ValidationFailed/EvidenceMissing→400；ScopeDenied/ApprovalRequired/ApprovalRevoked→403；VersionConflict/ContextStale/IndexStale/DeliveryUnknown/LeaseLost/BudgetExhausted→409；UnsupportedCapability→501；ProviderUnavailable→503；其他错误→500，并沿现有 `writeError` 错误 DTO 返回安全摘要。实际源码为 `internal/adapters/httpapi/swarm_missions.go` 和 `swarm_work_items.go`，H 保持 OpenAPI、主 router 与装配的唯一写入范围。

上述清单已先发 H、M-relay 和本模块；当前文档不表示 H 新组合已经测试通过。原生依赖继续由 R 提供，已知两阶段准确 SOURCE 为 `cc326be57a9a5f3bdd45687b0ac23b972adb8cef` / REPORT `41b4d71452f711f972d3c9836e0484fa94fdb854`，不依赖其后续 L Interpreter WIP。真实成员、执行描述、ownership 组合与完整 S3 验收仍待实际运行。

## 第十阶段：当前成员绑定与原生事实桥接

Task `task_da9aebf03f83` / Dispatch `ctx_2e078edf896b` 对模块 Task `task_b0197ca7ebdf` / 当前 Dispatch `ctx_9ce12d181768`（原 `ctx_b1cd2e0a39aa`）继续实施。模块 `msg_df9afffad744` 已明确：R 只提供原控制器事实，S2 核绑定、epoch、Grant 和当前权限，S3 在自身 `swarm_*` 范围映射 Member/执行描述/ownership，H 只做工厂装配。没有扩大写入范围或派发其他人员。

SOURCE `70e6e5a03fca9a19aa0a70e44b23b78337d34714` 已普通提交并 push，远端同值。Claim、Submit 和局部任务选择在读取实际执行描述前，经 `executionForGrant` 在同一真实授权事务中按 `claim_work_item` / `submit_artifact` 用途重新检查 S2 绑定。该检查使用现有实际账本累计用量；最终认领或提交仍另行重核完整资源、当前工具、授权与 CAS。

`SwarmApprovalBridge` 使用实际 `WithAuthorizedSessionBindings` 和 `ReadContextPacket` 读取授权绑定及不可变上下文包；核对 Packet/Project/Grant/epoch/Worktree 后，向私有 `AuthorizedBinding` context 注入用途、binding ID/version、packet ID、project/member、session/native session、adapter、worktree ID 和固定 Root。`AuthorizedBindingForMember` 将调用方与 Member 逐项匹配；缺少真实授权上下文、用途变化、身份或版本变化、工作树变化和不完整原身份均立即拒绝。请求正文不能填写或覆盖该 context。参数测试仅验证这些规则，没有构造原生 API、Registry 或模型成功。

| 本阶段实际命令 | 结果 |
|---|---|
| `go test ./internal/swarm/... ./tests/s3 -count=1 -timeout=120s` | PASS，包含新的绑定参数检查与真实 S2/SQLite 链 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1 -timeout=180s` | 最终 PASS，7.865 秒；首次完整运行的失败另行保留 |
| `go test ./internal/adapters/httpapi -run '^TestSwarmActual' -count=1 -timeout=180s` | PASS，16.054 秒 |
| `go vet ./internal/swarm/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/s3` | PASS |
| `go build -o D:/Astrocyte-agent-assets/20261015/product/s3/astrocyte.exe ./cmd/server` | PASS |
| `node scripts/check-architecture.mjs`、`git diff --check` | PASS，33 包边界保持 |

首 RED 保留：首次 SQLite 组合运行中，`TestSwarmActualFixedGoChecksApprovedPublicBaseline` 在原两分钟 Lease 窗口内未完成，141.63 秒返回 `context deadline exceeded`，包运行 172.417 秒失败。没有增加 Lease、测试超时、skip 或减少断言。随后相同源与限制的该测试单独运行 6.108 秒 PASS，再运行完整六项组合 7.865 秒 PASS；此前超时的具体原因尚未确认。新 Go 命令已按根当前指令仅对自身进程使用 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`，`GOTMPDIR` 继续为本轨独占 `D:/Astrocyte-agent-assets/20261015/go-temp/product-s3`，没有清理或复制缓存。

R 新 SOURCE `9807d2b6d7f0042bcd59d36ded0b332d74333dc2` 提供 `Registry.ReadOwnedInputOperation(ctx, sessionID, inputOperationID, purpose)` 与 `StopOwnedInputOperation`。已只读核其两用途语义：认领要求原 live controller，提交要求原 input 实际已受理并允许原持有退出 tuple；Tools/Targets/Calls/Tokens Known 目前均为 false。该文件依赖完整 cc326 的 Relay 基础，而当前 S3 起点没有整套 Relay 类型与实现；已向 R 交独立可编译通用事实契约和准确最小来源请求，避免重复引入无关历史。未知事实保持未知，不能据此放行认领或释放有上限的预算。

根 `msg_cfd1d9b6090a` 经模块 `msg_ec54aa029bae` 确认已安装 Codex 0.162 原生 schema 实际不提供完整 Tools/Targets 集合和全部尝试调用分母；对应 Known 保持 false。原 `rawResponse/completed` 只支持去重后的已完成 response 计数，nullable token usage 只能按原 accepted turn/response 在私有来源中去重。没有把缺证据写成零或完整成功率，没有修改公开六字段。当前 Swarm 的原生执行守卫仍拒绝未知完整执行描述，有上限的实际 calls 未知时继续保留预算；已将授权政策验证与这些观测指标的区分作为待明确项交模块，不自行变更预算含义。

S2 当前绑定实现对所有操作无条件拒绝 `session.StopConfirmed`，已将这一真实提交阻断交 S2 `ctx_b9a5305d8480`：认领应核当前 live，提交应保留同一原持有退出 tuple 的来源与当前批准；新绑定的 live 条件继续独立核实。S3 没有修改 S2 文件或绕过其授权。完整原生事实桥、两个真实 Agent、实际提交与人类采用、接续/交接和第二任务经验引用尚未通过，本阶段没有模型或原生启动，Task 继续执行。

## 第九阶段：固定项目实际公开

Task `task_da9aebf03f83` / Dispatch `ctx_2e078edf896b`，模块 Task `task_b0197ca7ebdf` / 当前 Dispatch `ctx_9ce12d181768`，原模块 Dispatch `ctx_b1cd2e0a39aa`。根决定 `msg_5691824b3258` 经模块 `msg_d264ae90eaf9` 授权本 owner 将必要的固定任务代码与原历史实际公开；M-relay `msg_c43ffea07b79` 确认沿唯一原 Root，没有其他 remote 或竞争项目。

公开仓库为 [astrocyte-s3-public-stats](https://github.com/songconmaisaix31-design/astrocyte-s3-public-stats)，origin 为 `https://github.com/songconmaisaix31-design/astrocyte-s3-public-stats.git`，默认分支 `main`，准确 SOURCE/baseline 为 `da3bdddb085808863090a05f60cc1c64415039cf`。本地根继续使用 `C:/Users/DW/orca/workspaces/Astrocyte/product-swarm-20261015/.verification/s3-public-stats`。只有 `README.md`、`go.mod`、`stats.go`、`stats_test.go` 四个原固定文件、一个原提交；没有增加文件、改变原历史或预先实现 Median。

| 本阶段实际命令 | 结果 |
|---|---|
| 固定项目 `git status --short`、`git log --all`、`git ls-files -s`、`git remote -v` | 发布前 clean、仅原提交与四文件、无 remote；发布后 clean、仅公开 origin |
| 固定提交 `git grep -l` 的凭据、私人路径标记检查 | 没有命中；没有输出文件正文 |
| `gh api user --jq .login` | 当前账户 `songconmaisaix31-design` |
| `gh repo create songconmaisaix31-design/astrocyte-s3-public-stats --public ...` | PASS，实际建立空公开仓库 |
| 固定项目 `git push origin da3bdddb085808863090a05f60cc1c64415039cf:refs/heads/main` | PASS，普通 push 原提交 |
| `gh api repos/songconmaisaix31-design/astrocyte-s3-public-stats`，仅提取公开 metadata | PASS：`visibility=public`、`private=false`、`default_branch=main` |
| 固定项目 `git ls-remote origin refs/heads/main` | PASS，远端 SHA 精确等于上述 baseline |

创建前的指定仓库 GET 实际返回 404，随后执行授权的创建。创建后首次只读 visibility 查询遇到 `TLS handshake timeout`，保留失败；后一次真实官方 GET 成功，没有重复创建或改写该失败。固定代码与已验证基线完全一致，本阶段没有重复运行应用测试。前阶段业务 SOURCE `10fa36cdd68b3247a70da9230453b74c0f5a04fa` / 独立 REPORT `58f847cac85a05e794d7ad1d19198712a96f256a` 已普通 push，远端一致。

公开事实独立 REPORT `5b2d4acd65fc14c3637964c92ae123a400f51c4f` 首次 push 被 GitHub 以 `fatal error in commit_refs` 拒绝，远端只读核实仍为 `58f847c`。随后普通 push 相同提交成功，没有 force、改写历史或重建仓库。该失败与后续成功均保留。

实际 remote、visibility 与 baseline 已交 H `ctx_e7abbaf902da`、F `ctx_4a385bf65390`、M-relay `ctx_bc51f7008918` 和本模块。真实原生动作仍由 R/H/F 按同 controller 空 Message 初始化、实际 source/tool 事实、observer ready、唯一登记和输入的顺序完成。本阶段没有模型输入、原生启动或登记；完整 S3、人工采用和第二任务经验引用继续 `NOT_RUN`，本 Task 尚未结算。

## 第八阶段当前结果

Task `task_da9aebf03f83` / Dispatch `ctx_2e078edf896b` 继续原分支与工作树，恢复基线为 `fb57f0f31c4e667b050cb26891a891190eaf4c14`。当前模块 Task `task_b0197ca7ebdf` / Dispatch `ctx_9ce12d181768`，原模块 Dispatch `ctx_b1cd2e0a39aa` 保留为历史。模块已明确恢复基线优先于最初的 `391c617` 起点；只读模块准确文档，不将管理文档合入业务分支。公开 `worker-show` 核当前客户端为 Codex、实际模型为 `gpt-6.1-sol`，本阶段外部模型调用数为零。

SOURCE `10fa36cdd68b3247a70da9230453b74c0f5a04fa` 已普通提交并 push，`git ls-remote` 核远端同值。RecoverMission 在外部对账前记录 Mission 版本，每完成一个原执行者的原子对账只增加一个预期版本。最终恢复事务核对该版本，并拒绝从 `completed` 或 `cancelled` 恢复执行；期间发生的暂停、取消或其他更新保留当前状态。截止时间与无进展停止条件仍会使恢复后的 Mission 保持 `blocked`。取消后的对账保留 `cancelled`，在当前项目读取权限内保存幂等回执，不要求重新启用已经撤销的执行批准。

真实 S2 → SQLite 测试实际创建项目空间、LocalProject、Goal、Proposal 和人类 Grant，经两个连接创建唯一 Mission。追加暂停后显式恢复、回执重放、同键异输入冲突，以及撤销后取消、取消后对账、关闭重开后的回执重放和异输入拒绝；没有创建原生会话或租约。Domain 参数检查覆盖较新版本的暂停与其他控制、两个终态、有效当前恢复及停止条件。这些结果不替代在途原生对账验收。

| 本阶段实际命令 | 结果 |
|---|---|
| `go test ./internal/swarm/... ./tests/s3 -count=1 -timeout=120s` | PASS；Domain、App 与实际 S2/SQLite 链 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1 -timeout=180s` | PASS；六项实际 SQLite、Git、固定 Go 检查组件 |
| `go test ./internal/adapters/httpapi -run '^TestSwarmActual' -count=1 -timeout=180s` | PASS；真实 listener、认证、SQLite 与本域生产路由 |
| `go vet ./internal/swarm/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/s3` | PASS |
| `go build -o D:/Astrocyte-agent-assets/20261015/product/s3/astrocyte.exe ./cmd/server` | PASS |
| `node scripts/check-architecture.mjs` | PASS，33 包 |
| `git diff --check`、普通 push、远端 SHA 核对 | PASS |

首 RED 保留：新增真实取消对账测试在 `approval_test.go:270` 得到 `wanted version_conflict, got <nil>`，没有删除或降低断言，修复后通过。首次 Go 运行因指定的 D 盘 `GOTMPDIR` 目录不存在退出 1；创建本轨目录后运行上述检查。Go 进程使用独立 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/product-s3` 和 `GOTMPDIR=D:/Astrocyte-agent-assets/20261015/go-temp/product-s3`，没有清理共享缓存。一次只读探查指定了不存在的 `workspace_transaction.go`，实际事务实现在 `workspace_proposals.go`，未修改其他人员文件。

B 当前 Dispatch `ctx_1fd66a8adc28` 已收到 `ReadAcceptedArtifactEvidence` 的准确 SOURCE、factory、同数据库事务和真实 outbox 来源要求。采用事实保存 `pending`，实际候选完成后回链 `candidate_created`；没有同步调用经验录入或制造采用事件。M-relay 当前 `ctx_bc51f7008918` 已交 R SOURCE `ce0302835cb7477eb606107351a9ab46dd8d4714` / REPORT `5d145cb148d1cf8d7053e2b2daf650eb32a95ef8`；本阶段只读核接口，未将该源码合入。R 当前 `ctx_99b96d053325` 已收到 Member Purpose、原完整 session/input operation、实际工具/目标与 Reconcile/Stop 适配请求；H 当前 `ctx_e7abbaf902da` 的主装配和实际运行结果尚未交付。

唯一公开改进目标保持本地受控 Go 项目 `Median`，baseline `da3bdddb085808863090a05f60cc1c64415039cf`，保留已有 Mean；当前没有互联网 remote，不宣称互联网发布。本阶段未预先实现 Median、未另起 CLI、未创建原生会话。完整两 Agent、原会话接续、显式交接、在途崩溃对账、人工采用及第二任务真实经验引用仍 `NOT_RUN`；等待原负责人满足有效工具与原生所有权事实并交付实际装配，本 Task 继续执行。

## 第七阶段当前结果

Task `task_da9aebf03f83` / Dispatch `ctx_d4bf65ca06bf` 对模块 Task `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa` 的后继 SOURCE 为 `2bd5d26ea6f932d8511afd869ff41f2146c0f94a`，前置 SOURCE 为 `93d1f6ba3ebb6ab029ced8ebe7c268a62d5ed2b0`。第六阶段 REPORT `23367b448d28c9780ec84d770548ccf870b4d8dd` 已普通 push，远端完整 SHA 一致。

成功经验来源的全部 Swarm 查询复用当前实际连接：Load、List、Timeline、ExecutionHistory、ReadAcceptedArtifactEvent 都先检查同数据库 Swarm/Workspace 事务。真实单连接测试在未提交 Mission 内执行各读取，缺少采用事实仍返回真实 NotFound；外层失败后 Mission 整体回滚。没有构造 accepted 事件或经验候选补充正向验收。B 的实际解释、候选写入和当前权限复核继续由原负责人完成。

LocalPolicy 在固定上下文排序前过滤预算耗尽、有限维度实际用量未知和缺少实际 token_budget/money_budget 能力。实际已用与当前预留共同占满 calls 上限时不推荐新工作项；减少预留后恢复可选任务。领域检查覆盖这些预算状态，没有执行原生会话或虚构绑定。

`go test ./internal/swarm/... ./tests/s3 -count=1`、`go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1`、`go test ./internal/adapters/httpapi -run '^TestSwarmActual' -count=1`、`go vet ./internal/swarm/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/s3`、`go build -o .verification/s3-build/astrocyte.exe ./cmd/server`、33 包架构检查和 `git diff --check` 均 PASS。实际采用、H 主装配、原生生命周期、两实际 Agent 和 S6 后任务经验引用仍未通过；这些组件结果不替代完整验收。当前客户端/模型仍为 codex / gpt-6.1-sol，本 Task 外部模型调用数为零。

## 第六阶段记录

Task `task_da9aebf03f83` / Dispatch `ctx_d4bf65ca06bf`，模块 Task `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa`。SOURCE `93d1f6ba3ebb6ab029ced8ebe7c268a62d5ed2b0` 使用 codex / gpt-6.1-sol 实施。本 Task 没有调用外部模型，没有创建原生会话。

预留 `Resources` 与实际 `Usage` 分开。实际 calls、tokens、money 分别使用 nullable 数量；没有证据的维度在 HTTP、Artifact 和持久化中保留 null。实际已知零保留为零，累计时未知部分继续保持未知。`ExecutionFact.Usage` 和提交正文 usage 现在使用 `domain.Usage`；原 ownership 适配必须逐项提供真实证据，known 标记与缺失数量冲突时立即拒绝。预算没有约束的未知维度不阻止对账；有上限的维度必须有真实数量且满足原预留与累计预算，否则保留租约、预算和写域。有限 tokens/money 预算在认领时还要求原执行描述提供真实的 token_budget/money_budget 执行能力，缺少能力时拒绝。

项目活跃执行者默认上限为 2，同时受 Grant 更小上限约束。完整构造函数仍接受五个实际依赖，新增可选 `ServiceOptions{ProjectMaxActive: n}`，只接受正数；详情返回实际有效的 max_active_executors。Claim 和 Submit 用各自真实 worktree 调用 S2 根范围检查，不能根据 Git 同仓库关系扩大权限。Member.Purpose 明确为 claim_work_item 或 submit_artifact，供原负责人分别核验存活执行者和已停止的原始绑定；该字段不能代替原生事实。原生 Start 的空 Message 不调用 send，可由原 R/H 沿既有实现准备控制器，再完成真实绑定和预算租约后发送新 input operation；不能事后补造租约。

SwarmRepository.Within 复用实际 WorkspaceTransaction 的同数据库连接。内部 Swarm 事务和 Workspace 事务都拒绝跨数据库复用；借用连接不自行提交、回滚或关闭。真实单连接测试覆盖同事务成功提交、失败整体回滚、错误数据库回调不执行。自有连接关闭错误与原错误共同返回。

| 本阶段实际命令 | 结果 |
|---|---|
| `go test ./internal/swarm/... ./tests/s3 -count=1` | PASS；新增未知用量/已知零/矛盾证据/累计用量检查，真实 S2 批准链继续通过 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1` | PASS；六个真实 SQLite/Git/Go 组件，含新增同连接事务复用、回滚与错误数据库拒绝 |
| `go test ./internal/adapters/httpapi -run '^TestSwarmActual' -count=1` | PASS；真实 listener、session guard、SQLite 与本域生产路由 |
| `go vet ./internal/swarm/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/s3` | PASS |
| `go build -o .verification/s3-build/astrocyte.exe ./cmd/server` | PASS |
| `node scripts/check-architecture.mjs` | PASS，33 包 |
| `git diff --check` | PASS |

本阶段没有新测试 RED。读取命令首次使用 PowerShell 不展开的 rg 路径通配符和不存在的 attention_tx.go，读取失败后改用实际目录与文件；未改变任何代码或验收断言。全部先前 RED、UNKNOWN 和 NOT_RUN 在历史记录保留原结果。

根已允许真实 Codex turn/started 公共固定任务正向验证；无 started 的 human-done 分支保持未决定。公开 Go 基线与 Median 固定任务指针已交 M-relay/H/F 原负责人，实际模型候选尚未返回，本 Task 没有预先实现 Median。R/S2 原 Registry 的 full session、准确 input operation、用途区分、有效工具事实和实际 worktree/context 映射仍需原负责人精确 SOURCE；H 主服务装配、真实两 Agent、人工采用、成功经验候选和后任务引用仍 NOT_RUN。上述组件 PASS 不表示完整 AT08/09/13 或 S6 通过，本 Task 继续执行。

## 第五阶段记录

SOURCE `ec653ddd90a9c91c83af28a8909fb41741116160` 已普通 push 并核实远端同值。Task `task_da9aebf03f83` / `ctx_d4bf65ca06bf` 对模块 `task_b0197ca7ebdf` / `ctx_b1cd2e0a39aa` 交付成功采用来源端口：`AcceptedArtifactEvidenceReader.ReadAcceptedArtifactEvidence(ctx, principal, localProjectID, eventID)`。它读取实际不可变 artifact_accepted 事件，核对注册 LocalProjectID，再通过当前项目权限回调授权；返回实际 actor_id、actor_kind、occurred_at 及固定 Artifact/version/source/commit/object/检查引用。actor_kind 来自采用时实际保存的人类 Principal，旧事件缺少该事实时拒绝补造。

候选回链 `ExperienceCandidateSource` 明确区分 LocalProjectID 与 OwnerProjectID；前者是注册项目 ID，后者由 B 的真实 LocalProject→AttentionSpace 映射提供。S3 不根据名称猜测映射，不读取 Attention 内部表。实际候选录入、批准、分享和正向回链仍 NOT_RUN；此阶段的真实批准/HTTP组件测试与 Domain 编译继续 PASS，新增来源读取正路径等待真实采用事实。

## 第四阶段记录

Task `task_da9aebf03f83` / Dispatch `ctx_d4bf65ca06bf`，模块 Task `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa`。SOURCE `3c568c7ec8bb13a88b0abf95d44cbb13fc6422b5` 已普通 push，远端完整 SHA 一致；Orca 当前公开 worker 配置记录为客户端 codex、模型 gpt-6.1-sol，本 Task 外部模型调用数仍为零。

Grant 的固定输入现由实际 S2 `ReadProposal` 提供，包含批准提案 revision、Goal revision 和实际已批准资料 revision。AddWorkItem 对普通输入执行白名单检查；真实批准测试追加了未批准输入拒绝及事务没有留下部分任务/计数的断言。相同 Grant 用新幂等键重新投递时复用原 Mission，并保持真实 mission_created 事件只有一条。

根据根允许的 Swarm 域内成果继承，AddWorkItem 增加 artifact_inputs，字段为 id、version、commit、object_ref。服务读取既有 outbox 的不可变 Artifact 版本，要求该版本真实 accepted、接受者引用存在、commit/对象完全匹配，并检查同项目和当前 read_swarm 权限；没有跨项目默认分享。普通资料输入、结果引用与执行权限分别检查。GET Artifact 和 diff 支持可选正整数 version，以读取实际保存的旧版本；没有根据最新 metadata 猜测旧版本。

原执行结果必须属于 completed、failed、cancelled、stopped、not_started 的已知事实；空值和未知结果保持占用，not_started 无法作为成果提交。Artifact 保存实际 execution_outcome 与经过允许列表校验的 execution_failure_code，原始失败文本没有公开输出。调用/tokens 的 known 检查与未知费用规则保持原含义。

本阶段 `go test ./internal/swarm/... ./tests/s3 -count=1`、`go test ./internal/adapters/httpapi -run '^TestSwarmActual' -count=1`、`go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1` 均 PASS；scope 测试、真实 HTTP/SQLite 认证与回执、五个真实 SQL/Git/Go 组件继续通过。`go vet ./internal/swarm/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/s3`、当前服务构建、33 包架构检查、`git diff --check` 通过。

成果引用接口已经实现，实际 acceptedArtifact 来源和真实第二任务执行仍 NOT_RUN；没有使用合成 accepted 记录填补原生依赖。当前独立实现与适用组件检查已交接 H/B/F/UI，后续需要 H 主服务装配精确 SOURCE、R/S2 同原 Registry 的实际绑定/原 operation 生命周期及有效工具/目标事实、B 实际成功经验候选适配。原生合法条件仍按根待答保持未决定。完整 AT08/09/13、实际采用和 AT16 没有降低验收，本 Task 尚未结算。

## 第三阶段记录

Task `task_da9aebf03f83` / Dispatch `ctx_d4bf65ca06bf`，模块 Task `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa`。SOURCE `e18c0965708e091b547ca2108579a4b3ca81b72c` 已普通 push，远端完整 SHA 一致。普通消费 S2 当前归属检查 SOURCE `80248d2c7177823d4535e7730453481426a27305` 和 B 的明确退回 policy SOURCE `c110201774dd61b8d7e5280526364786c6af4cc2`；直接编辑仍在本 Task 独占路径。以下各历史阶段保留当时的实际证据与限制。

新增 `GET /api/v1/artifacts/{id}/diff?mission_id=...`，返回 schema_version、artifact_id、commit、base_commit、diff。只读取已提交 Artifact 的真实 Git diff 对象，检查对象 digest 与 UTF-8，按当前项目权限读取。新增 `POST /api/v1/artifacts/{id}/rejection`，正文为 metadata、mission_id、reason；使用独立 `reject_artifact` 人类权限、当前 Grant 和 Artifact CAS。人工退回保存真实审阅者、时间、原因，保留原 commit、对象、检查、采用以外的提交事实，进入 revision_needed 并计入无进展轮数。

Mission 详情新增 executions，只从已存在 outbox 的 work_item_claimed、artifact_submitted、execution_blocked、execution_reconciled、artifact_revision_requested、artifact_accepted 构造原 operation 引用与事件链；没有创建新 Attempt 身份或调度器。绑定版本和不可变 context_packet_id 来自同事务的 S2 已授权绑定，保存在 Lease 与 Artifact。尚无真实工具事件适配时 tool_events_status 为 unsupported，空事件列表不表示实际使用工具为零。

根据模块已确认的事务边界，AcceptArtifact 现在提交人类采用、WorkItem 状态及真实 artifact_accepted outbox，experience_id 为空，experience_status 为 pending。实际成功候选录入交 B 原经验应用负责人。受信任 consumer 可通过 `ReadAcceptedArtifactEvent` 读取真实不可变事件，`LinkAcceptedExperience` 必须使用真实 ExperienceCandidateReader 核候选版本、拥有项目、来源事件与 Artifact 后，幂等保存 candidate_created 引用；此链接不批准经验。上述两个内部端口没有 HTTP 构造入口，实际候选录入和后任务 approved 经验仍 NOT_RUN。完整构造函数现为 `NewSwarmService(repo, grants, members, evidence, ownership)`，五个依赖均须实际实现。

根批准的公开 Go 验证项目已建在 `.verification/s3-public-stats`，独立本地受控 Git baseline 为 `da3bdddb085808863090a05f60cc1c64415039cf`，没有 GitHub remote。正确 Mean 基线支持空值/非有限值拒绝、正负/重复/极大有限输入与输入保持，实际 go test、go build、git diff --check 通过。目标为新增 Median 的真实有用功能，尚未实现 Median、没有启动两原生会话。本地首次在已指定工作目录后仍使用重复相对路径，gofmt 与 git init 路径错误；使用准确路径后运行通过。原生工具仍关闭，实际有效工具集合、外发目标和 working 条件缺证保持未决定/unknown。

`NewSwarmGoProjectChecks(root, baseline, objects)` 只对明确根与 baseline 所属仓库执行固定 go test/build，执行前后重核当前 commit、跟踪与未跟踪输入；使用实际进程退出结果和实际起止时间，原 stdout/stderr 只通过对象引用保存。对其他仓库直接拒绝。提交同时保存 project_checks_ref；Git diff --check 与项目适用检查分别验证，原执行停止与分项 usage 必须先确认。可重现组件测试使用 `tests/s3/testdata/public-stats` 中相同的正确公开 Mean 源码、真实独立 Git 与实际 Go 工具，没有替代工具协议、模型或 API，也没有构造故意失败基线。

| 第三阶段实际命令 | 结果与覆盖 |
|---|---|
| `go test ./internal/swarm/... ./tests/s3 -count=1` | PASS；固定策略和实际 S2 批准/两个连接 Mission 唯一、异输入冲突、重启、撤销拒绝 |
| `go test ./internal/adapters/httpapi -run '^TestSwarmActual' -count=1` | PASS；生产 registerSwarmS3、生产 session guard、实际 HTTP listener、真实 SQLite；未认证/无 CSRF 拒绝，真实 human 创建与重放、伪造 actor_type 拒绝、每个 timeline event_id 逐条定位真实 outbox、撤销启动拒绝、撤销后人类安全暂停，租约数保持零 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1` | PASS，5 项；原 4 项与真实固定 Go 检查回执、独立仓库拒绝 |
| `go vet ./internal/swarm/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/s3` | PASS |
| `go build -o .verification/s3-build/astrocyte.exe ./cmd/server` | PASS；当前 Go 服务真实构建 |
| `node scripts/check-architecture.mjs` | PASS，33 包 |
| `git diff --check` / push / ls-remote | PASS |

HTTP 结果属于本域生产路由组件与真实认证/存储组合；H 的 server/main 主服务装配、共享 Agent route policy、OpenAPI 与整个部署仍需 H 精确源码及实际验证。真实人工退回、diff 页面与执行历史的完整正路径需要真实提交；现阶段仅代码/编译完成，不能由 SQL 参数或空数组代替实际成果。两实际 Agent 的 Median 改进、原会话 resume、显式交接、未知在途对账、人工采用、真实候选与实际第二任务引用仍未通过。本 Task 继续执行，未发送 worker_done。

## 第二阶段记录

开发仍为 `task_da9aebf03f83` / `ctx_d4bf65ca06bf`，模块仍为 `task_b0197ca7ebdf` / `ctx_b1cd2e0a39aa`。第二阶段 SOURCE `3b78fd04dbf758e745e9880297ee1c2d6e58bfc2` 已普通 push，远端同值；准确消费共享 policy SOURCE `4725f0b1c6907ddefb3b1e1dac0f49840031db59`、S2 SOURCE `5d0ba735533f68259b91128dc404c8e4c2f2a535` 和 `507be52f4e31759dc8c3765aeea964f1f2efe013`。以下第一阶段记录保留当时结果。

真实 S2 → Swarm 链已通过：Attention 项目空间、真实 LocalProject、Goal、Proposal、人工 ApproveProposal，经两个独立 SQLite 连接并发 CreateMission，得到同一 Mission；同键不同输入冲突、关闭重开后的回执重放与撤销后拒绝启动均通过。未启动原生执行者，没有使用批准 reader 替代实现。首轮因 S2 的零并发限制返回 scope_denied；原 S2 修改允许真实零活跃数后，第二轮发现回调替换 context 导致另取同 DB 连接等待，已终止该次实际测试。Swarm bridge 在相同连接重新附加自身事务 context 后回归通过；两次失败保持原含义。

授权 bridge 现在在公开 Workspace 写事务中查询自身 Swarm 账本，传递实际活跃租约数、当前操作与累计 Used/Reserved/本次资源、可信 Caller、实际工具和外发目标。Claim 和 Submit 需要准确 `binding_id`，同事务消费 S2 的 `WithAuthorizedSessionBindings`，并核对真实成员；未知执行策略拒绝认领。调用和 tokens 使用独立 known 标记，没有完整实际用量证据时不释放占用。原 R 明确当前真实有效工具集合与目标 endpoint 仍 unknown，未以请求关闭配置或零 tool_calls 填为已核实。

`httpapi/swarm_missions.go` 和 `swarm_work_items.go` 提供 `registerSwarmS3(mux, *swarmapp.Service)`，Principal 从 `localCaller` 注入。Domain RuleError 在本域 HTTP 层转换为既有 apierrors，scope/approval 拒绝为 403，version/fence/budget/unknown 为 409。路径与正文 ID 冲突拒绝；外部请求不能写 trusted consumer 的 event_id。

公开路由为 GET missions、missions/{id}、missions/{id}/work-items、missions/{id}/timeline、artifacts/{id}?mission_id=；POST missions、missions/{id}/work-items、missions/{id}/pause、cancel、recover、work-items/{id}/claims、work-items/{id}/artifacts、artifacts/{id}/acceptance，均使用 `/api/v1/` 前缀。Mission 详情返回 schema_version、mission、work_items、leases、artifacts、timeline；时间线仅返回真实 outbox 的 event_id、type、aggregate_id、aggregate_version、occurred_at，不返回原始模型或控制器内容。H 负责主服务路由替换、鉴权规则、OpenAPI 与装配；本阶段没有修改这些共享文件。

`NewSwarmLedger(repo, actualGrantAuthorizer)` 可提供真实账本与批准操作，执行能力缺失时明确返回 unsupported_capability。完整 `NewSwarmService` 需要真实成员、原执行者、成果检查和经验管道。Git/object 实际集成测试已通过；Git diff --check 的结果单独保存，Submit 同时要求实际项目适用检查的 `ApplicableChecksVerified`。缺少真实项目检查时无法进入可采用提交。固定策略 `swarm.fixed.v1` 先筛当前权限、依赖、能力、占用与停止条件，再按相同固定输入匹配排序；策略版本保存在 Lease。

| 第二阶段验证命令 | 实际结果 |
|---|---|
| `go test ./internal/swarm/... ./tests/s3 -count=1` | PASS；固定 LocalPolicy 安全筛选与真实批准/两连接 Mission/幂等/重启/撤销链 |
| `go test ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1` | PASS，4 项；新增真实 Git worktree、commit、diff/check 对象正路径及无关仓库/脏工作树/错误对象/遗漏路径拒绝 |
| `go test ./internal/adapters/httpapi -run '^$'` | PASS；真实领域路由编译，尚未当作实际 HTTP 验收 |
| `node scripts/check-architecture.mjs` | PASS，33 包；原 S2 Domain 越层首 RED 由其原负责人修复 |
| `git diff --check` / push / ls-remote | PASS；第二 SOURCE 远端完整 SHA 一致 |

新增首编译失败记录：Recover 的重放检查增加局部 err 后重复声明，编译拒绝；删除该重复声明后通过。第二阶段没有新增 skip、替代原生或外部 API 实现，也没有减少断言。

当前剩余：H 主 HTTP/鉴权实际服务尚未装配验收；真实成员绑定的原 full Start session / 同 Registry / controller / startup 与 input operation resolver 待 R/S2 原负责人闭合；真实有效工具集合、目标与分项 usage 缺证保持 unknown。实际公开小改进项目、获准固定测试/构建由模块向根确认，项目原生自由工具权限保持关闭。实际两 Agent、一次显式交接、人工采用、经验候选真实录入和后任务 approved 经验实际使用均 NOT_RUN。AT07 的真实服务批准、并发唯一 Mission、重启/撤销组件通过；完整 AT08/09/13、S6 实际第二任务、最终组合 check/build 仍未通过，本 Task 继续执行。

## 第一阶段记录

开发 Task `task_da9aebf03f83` / Dispatch `ctx_d4bf65ca06bf`，模块 Task `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa`。使用当前 Orca Codex 客户端和会话所配置模型；本阶段未调用外部模型，也未根据客户端名称推断模型。独立分支 `songconmaisaix31-design/product-swarm-20261015`，接管源码 `391c6176145044b099c13a05a926aef2096db170`。

第一阶段 SOURCE `b220ab27c4e57d0bfdfaef224a9d03c1ae92366a` 已普通 push，`git ls-remote` 核实远端同值。通过普通合并消费 B 的五表迁移 SOURCE `da48a114d91181075bd32b62996beadd9c2c0a25` 和 S2 的 Grant 公开端口 SOURCE `4eec31ddfde39fa3f548691ec6e7f2f34470351a`。全部直接编辑位于本 Task 独占路径；共享迁移、依赖、Workspace 业务与 Web 由各原负责人维护。

## 实现与接口

`internal/swarm/domain/mission.go` 定义 Mission、WorkItem、Lease、Artifact、固定来源版本、写域、资源与停止条件。货币使用明确币种和最小单位，空费用保持未知；预算上限与已知资源数量分开，空 calls/tokens 上限保留原值。

`internal/swarm/app/contracts.go` 和 `commands.go` 提供 CreateMission、AddWorkItem、ClaimWorkItem、SubmitArtifact、AcceptArtifact。Command 的 schema_version、request_id、expected_version 位于 JSON 顶层，Idempotency-Key 来自受信任入口。Principal 不能从命令正文反序列化。采用要求人类身份，事务内保存 `accepted` Artifact、WorkItem 状态及真实 `swarm.artifact_accepted` outbox，`experience_id` 为空、`experience_status` 为 `pending`。B 的实际 consumer 读取该不可变事件并录入经验候选，Swarm 使用 `LinkAcceptedExperience` 核对真实候选来源后保存 `candidate_created` 引用；经验批准与分享继续由原治理接口处理。

认领在授权写事务内重新检查成员、依赖、当前任务版本、项目活跃执行者上限、写域占用与预算，然后保存工作项、单调 fence、租约、资源预留、幂等回执和 outbox。过期租约继续占用资源，删除租约必须使用匹配 fence。提交保存独立 Artifact；提交与采用使用不同命令。Git merge、push 和 release 未作为采用副作用。

`lifecycle.go` 提供暂停、取消和显式恢复：先持久化停止新认领的状态，再调用原执行者的停止或对账端口。原 operation/session 或用量未知时保留租约、预算和写域，标记 blocked；恢复只核对原身份，不创建执行者。新认领仍要求显式请求和当前批准。

`sqlite.NewSwarmRepository(db)` 使用真实 SQLite BEGIN IMMEDIATE；`NewSwarmApprovalBridge(db, actualWorkspaceGrantAuthorizer)` 消费 S2 的 `WorkspaceTransaction(ctx, db)`，复用同一数据库连接。桥接不读取 Workspace 内部表。`NewSwarmGitEvidence(objects)` 核验独立 Git worktree、共同仓库、固定 commit、真实 diff 对象和固定 git diff --check 记录；路径判定由 Git 完成，无自写 Git 格式 parser。

## 已运行验证

| 命令 | 实际结果与覆盖 |
|---|---|
| `go test ./internal/swarm/... ./internal/adapters/sqlite -run '^TestSwarmActual' -count=1` | PASS；三个真实 SQLite 组件测试。两个独立 DB 连接并发仅一个认领成功；失败事务没有半预算；版本/fence/尝试计数、回执与 outbox 一致；关闭重开后过期未知租约保留；grant 唯一；消费事件去重；旧 fence 无法删除租约 |
| `go test ./internal/adapters/sqlite -run '^$'` | PASS；适配层编译验证 |
| `node scripts/check-architecture.mjs` | 在消费 S2 前 PASS，31 包；消费 `4eec31d` 后 FAIL，Workspace Domain 引入 apierrors，已交 S2 原负责人返修；Swarm Domain 自有 RuleError，未越层引用 |
| `git diff --check` | PASS |
| `git push` / `git ls-remote` | 第一 SOURCE push 成功，远端完整 SHA 一致 |

SQLite 测试中的明确参数仅验证真实数据库与领域规则，未声称存在真实批准、原生会话或 Agent。该组件结果不等于 AT08 两真实执行者完整通过。

首失败保留：初次 SQLite 测试编译误用了不存在的 `DB.Migrate`，实际方法为 `RunMigrations`，修改测试调用后上述三项通过；首次 PowerShell gofmt 命令把 `swarm_*.go` 作为字面路径传入而失败，随后使用准确文件路径完成格式化；首次元数据补丁因当前格式与预期文本不同拒绝应用，读取当前文件后重新应用。没有降低断言或新增 skip。

## 验收与依赖

| 场景 | 当前结论 |
|---|---|
| AT07 真实 S2 批准、重复启动、撤销启动拒绝 | NOT_RUN；S2 019 迁移与实际业务验证尚待原负责人 SOURCE。当前只有同 grant 最多一个 Mission 的真实 SQL 约束组件通过 |
| AT08 完整授权/成员/两个实际 Agent/HTTP 认领与成果 | NOT_RUN；SQLite 两连接 CAS、无半预算与旧 fence 组件已通过，真实成员与原生原 ownership 端口尚待接入 |
| AT09 原会话真实 resume 与一次显式交接 | NOT_RUN；R SOURCE `79d93132967de20fb1fda3d8d74b4edff386178c` 的安全 current ownership 端口已通知，实际适配与合法原生条件待完成 |
| AT13 在途崩溃与原 operation 对账 | NOT_RUN；未知占用保留的 SQLite 重启组件通过，原控制器实际在途场景待完成 |
| S6 LocalPolicy、无进展停止、第二实际任务引用 approved 经验 | NOT_RUN；本 Task 继续负责实现，经验治理/查询走原公开接口，不修改冻结经验内部 |

本域 HTTP 路由、真实 S2 批准和 SQLite/Git/object 组件已经实施并运行；H 主服务、F/UI 与 R 实际成员及生命周期组合以各原负责人精确 SOURCE 和本轮实际结果为准。Artifact 人类采用与后续经验候选分别提交；经验候选来源从真实采用事件读取。完整原生提交重放、生命周期、两 Agent 改进、人工采用及第二任务经验引用尚未通过，本 Task 继续返修和集成。
