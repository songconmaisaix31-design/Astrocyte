# Swarm 发送前租约校验与单开发 Agent PR

用户 2026-10-18 要求后续由一个开发 Agent 提 PR、主 Agent review，并继续功能开发。管理基线 `2424e6ba9fc61280b07570e245f92cd960bb5e99`；产品基线 `f17a491cdd6392829094c95a3e5efb6011362bc0`，接管时均干净。唯一开发 Agent 使用新工作树和分支 `swarm-send-lease-20261018`，从产品基线创建；主 Agent 维护本计划、状态和 PR review，不修改业务代码，不安排第二个开发或集成 Agent。

## 功能目标

沿 SPEC §7.2、§10.4、§12.5–12.6、§17.2–17.3，补充 Swarm 发送前对当前租约、fence、成员绑定、原准备身份和 operation 的校验。租约丢失、过期、已停止、授权撤销或绑定变化时，发送必须在副作用前失败；未知交付不重发。消费方端口属于 app，SQLite 与外部事实读取属于 adapter，启动入口仅做组装。独立 Works Relay 与 Swarm 的权限规则保持各自已确认范围。

接管基线的 `SendNativeMessage` 消费固定准备，尚未消费 S3 租约。开发 Agent 已交付完整方案、精确接口及写域，经主 Agent review 后实施。最终实现明确准备用途，在 Claim 同事务首次固定原租约，并在发送占用事务核对当前 S3 身份、授权、租约和 fence；外部观察与发送位于事务外。生产准备/发送依赖尚未启用。不能从缺少租约推断为独立发送，不能把历史租约当当前执行权限。

## 写入范围

- `internal/workspace/app/native_messages.go` 及同目录新增 `native_message_*.go`，需要修改 `projects.go` / `contracts.go` 时先交准确 Handoff。
- `internal/swarm/app/**`、`internal/swarm/domain/**` 内与发送授权相关文件。
- `internal/adapters/sqlite/swarm*.go`、`internal/adapters/sqlite/local_native_message*.go`；新增授权 adapter 必须位于该目录并以 `swarm_` 或 `local_native_message_` 命名。
- `internal/adapters/agents/relay*.go`、`cmd/server/relay*.go` 内必要组装及相应真实测试。
- 产品 `AGENTS.md` 同步用户最新单开发 Agent / PR / 主控 review 流程；`docs/standards/swarm-send-lease-20261018.md` 记录方案、实际实现、命令及剩余限制。
- 新 HTTP/OpenAPI 契约、DDL、`cmd/server/main.go` 或以上范围外修改先交 Handoff，不自行实施。

## 验收与执行约束

实现和验证禁止 mock、伪造原生身份/会话/API/CLI、制造模型输出、新增 Skip 或移除原断言。仅使用 Agent 创建的公开 Git 项目、真实 SQLite、实际 HTTP。Codex 文件/命令工具关闭条件仍不满足，禁止原生初始化和模型输入；S3 调用预算、能力证据范围、archive 可信完整性及具体经验批准保持未决定，不添加默认、Known=true 或隐式授权。实现不能据此打开生产发送。

允许范围内必须执行真实缺来源、事务、取消、权限拒绝及重启检查；正向原生来源不足的部分如实未验。运行精确受影响测试、Go vet/build 和 DDD 架构检查；不重复与变更无关的既有发布包验收，不运行含替身的完整旧测试集合。适用检查失败由同一个开发 Agent 修复。

当前资源限制为 `GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`、BelowNormal，编译、测试、实际服务串行，最多一个临时服务。每项重操作前两次采样 CPU <70%、可用内存 >=4 GiB，间隔 2–5 秒；不满足时不启动，继续静态代码或文档。软内存目标不作为硬峰值保证。不关闭其他应用、用户接管终端或历史服务。

开发 Agent 提交并普通 push，使用结构化正文或 `gh pr create --body-file` 创建目标为 `songconmaisaix31-design/works-relay-integration` 的 PR；标题与说明描述最终功能和实际验证，保留未验项。主 Agent 在 PR 创建后独立 review，有问题交回同一开发 Agent，必要复验仅覆盖变更及未解决问题。PR review 与实际验收完成情况分别记录；没有通过的检查不计通过。当前任务不授权正式 Release、共享产品部署或正式 V4。

Orca Run 沿用 `run_73578b70c449`，启动时主控为 `term_66e5a559-ee45-4d6f-8a1a-8bffa7996608`，重启后为 `term_d7578194-fd20-42b4-9668-b0b0e8683571`。

Task `task_1795bc720ae3` 的初始 Dispatch 为 `ctx_7e6f00eb7fdb`；工作树 `C:/Users/DW/orca/workspaces/Astrocyte/swarm-send-lease-20261018`、同名分支；初始终端 `term_8eb15cea-cc6b-4b7f-8c33-ddf81cf3fcb1`。Orca 1.4.220 / Codex 0.162.0，终端显示实际 `GPT-6.1-Sol xhigh fast`，沿用户已有配置启动，无模型覆盖。启动回执为 turn_start_unobserved，实际读取到任务停在 composer 后只补发一次 Enter，没有重复任务。

18:16 开发 Agent 交付方案 `msg_c3b02ba4e1f7`，主控 review 后同意在原写域实施：明确 project/swarm 准备入口并持久化用途；Claim 仅消费 Swarm 准备，同事务首次固定原 fence；发送在当前 Grant/绑定/Mission/WorkItem/Lease 授权事务中 CAS `send_attempted`，所有外部观察和 Send 位于事务外。用途未知不默认、历史 Swarm 标记不因租约消失而删除或降为 project，私有字段不公开。发送占用提交作为动作开始点，提交后撤销属于在途，不声称跨 SQLite/原生控制器的原子提交；发送前传递最早绝对截止时间与取消，失败仍保留单次占用，未知不重发。完整方案保存在开发树 `docs/standards/swarm-send-lease-20261018.md`；生产依赖不启用，原生正向未验时创建 draft PR，主控 review 后由同一开发 Agent 返修。

20:39 接续时 Orca runtime 已变为 `335b14a9-75ca-4611-83f6-27a2ac91dbcd`，当前主控为 `term_d7578194-fd20-42b4-9668-b0b0e8683571`，同 Run generation3。原 Dispatch 由 execution_host 明确记录 failed / process_exited，14 件未提交修改与原验证脚本保留，没有 PR 或已完成测试日志。相同 Task 在同一工作树、分支恢复为 `ctx_3bdd3825f01c` / `term_6b6909b0-97b7-4e35-a45a-601f887c15d3`，继续已批准方案，不重复设计决定。旧终端官方 release 两次返回 `release_unknown`，记录终端已无法解析；保留该不确定清理状态，不扩大关闭范围。

## PR 与当前 review

源码 `3abe39be8f09689012ddd818e19141a1d480ea25` 已普通 push，Draft PR 为 <https://github.com/songconmaisaix31-design/Astrocyte/pull/1>，目标分支为原产品集成分支。推送 TLS 错误与创建 API EOF 后先读状态再重试，没有重复 PR。现有默认 workflow 会运行尚未整改的替身集合及缺少真实来源的全套检查，本轮提交使用 `[skip ci]`，完整默认 CI 保持未执行、未通过；本任务的真实适用测试与构建单独实际执行，不改变默认 workflow、断言或测试 Skip。

主控已静态读取该提交全部业务差异和新增测试。21:02 原作者 Handoff `msg_8010fdb44759` 确认：现有 project 检查仅查询准备标记与租约，尚未覆盖先绑定 S2、后准备的会话。主控批准在同一最终事务读取绑定的可信批准动作及 Mission 关联，发现 `claim_work_item` / `submit_artifact` 或既有 Swarm Mission 即拒绝 project，不按 active/status 丢弃历史依据；通用 Workspace 绑定不一概归为 Swarm。来源缺失、损坏或无法证明用途直接失败，不据错误回退 project，不新建用途默认、DDL 或公开接口。修复与返修检查由同一开发 Agent 追加 commit、普通 push 并更新同一 PR。真实原生绑定正向与原 S3 未答项仍未解决，PR 保持 Draft。

## 最终代码交付与检查

业务修复 `0642efdf7748a64c36e07e466b5f507b9c04a004`、最终报告 `75cadf95073d5608fb67f0a4b5be5d65e9b1d8ee` 已普通 push。主控实际读取最终报告、检查日志和修复差异，并通过 `git ls-remote` 核对开发分支远端等于最终 HEAD，工作树干净。产品分支仍为基线 `f17a491cdd6392829094c95a3e5efb6011362bc0`。

主控已独立核对原批准 Scope 不变、原提案 revision 保留及固定 packet 来源；最终绑定检查读取全部历史状态，三个 Scope、存储键、版本与 epoch 必须一致。P1 已修复，当前没有追加业务修改要求。[最终 review](https://github.com/songconmaisaix31-design/Astrocyte/pull/1#pullrequestreview-5563152542) 为 COMMENTED，指向准确最终 HEAD；发布成功回执 ID `5563152542`。初始 review POST 曾发生 TLS 超时，随后实际查询为空；最终只发布当前结果，没有重复旧问题评论。

以下由原开发 Agent 在资源守卫下串行执行，主控读取实际原件；`GOFLAGS=-p=1`、`GOMAXPROCS=1`、`GOMEMLIMIT=256MiB`、BelowNormal。完整命令和原件索引位于开发树 `docs/standards/swarm-send-lease-20261018.md`。

| 命令 | 结果与范围 |
|---|---|
| `go test -mod=readonly -p=1 ./internal/workspace/app -run '^TestNativeMessageExplicit' -count=1 -timeout=120s -v` | 四项参数规则通过，最终源码，原件 `rules-20261018T130624.log`。 |
| `go test -mod=readonly -p=1 ./internal/adapters/sqlite -run '^(TestNativeMessageActualTransactionBoundaries\|TestPreparedNativeMessageMissingOriginalReaderDoesNotPublish\|TestNativeMessageExplicitJSONRules\|TestSwarmNativeMessageActual)' -count=1 -timeout=120s -v` | 四项通过、零 Skip；含真实 SQLite 事务、取消、跨库、撤销与重开，原件 `sqlite-20261018T130652.log`。 |
| `go test -mod=readonly -p=1 ./internal/adapters/agents -run '^TestRelayPreparedMemberRejectsMissingOriginalSources$' -count=1 -timeout=120s -v` | 一项通过、零 Skip；最终 agents 实现相同，原件 `agents-20261018T124356.log`。 |
| `go test -mod=readonly -p=1 ./cmd/server -run '^TestRelaySwarmActualPreparedAssemblySQLiteHTTP$' -count=1 -timeout=120s -v` | 一项通过、零 Skip；实际 HTTP 与生产发送依赖拒绝，原件 `http-20261018T130723.log`。 |
| `go vet -mod=readonly -p=1 ./...` | 最终源码 EXIT 0；终端 `PHASE=vet EXIT=0`、资源采样 `vet-20261018T131035-resources.json`；标准输出为空，没有 stdout 日志文件。 |
| `go build -mod=readonly -p=1 ./...` | 最终源码 EXIT 0；退出记录 `build-20261018T131246.log`。 |
| `node scripts/check-architecture.mjs` | 最终源码 EXIT 0、42包，原件 `architecture-20261018T131517.log`。 |
| 本次 Go 文件 `gofmt -l`、`git diff --check` | EXIT 0，无输出。 |

CPU 72% /100% 的最终 vet 尝试均在调用 Go 前拒绝；后续两次采样为69% /62%、可用内存约7.59 /7.57 GiB时实际通过。最初失败、资源守卫拒绝和最终通过的原件分别保留。参数规则与空来源拒绝不计为真实原生正向验收。

实际验收仍缺正向 Prepare → Claim → 同 operation Send、真实已有 S2 绑定用途分流、原租约绑定正向 CAS、停止/撤销/换绑定/fence 与发送的竞态、租约消失后的持久化用途拒绝和未知交付对账。合法原生来源与文件/命令工具关闭条件未满足，没有初始化原生控制器或输入模型；原 S3 预算/能力证据、archive 与具体经验批准继续待答。完整默认 CI、L1/L2 和正式发布未执行，PR 保持 Draft，没有合并、部署或打包。

开发 Agent 的 `worker_done` 为 `msg_c8350121a6c6`，准确 Task/Dispatch 已核；当前 `ctx_3bdd3825f01c` 官方 release 为 `released / closed_agent_terminal / transcript captured`。原 `ctx_7e6f00eb7fdb` 的进程退出事实与 `release_unknown` 清理限制单独保留。
