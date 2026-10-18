# Swarm 发送前租约校验与单开发 Agent PR

用户 2026-10-18 要求后续由一个开发 Agent 提 PR、主 Agent review，并继续功能开发。管理基线 `2424e6ba9fc61280b07570e245f92cd960bb5e99`；产品基线 `f17a491cdd6392829094c95a3e5efb6011362bc0`，接管时均干净。唯一开发 Agent 使用新工作树和分支 `swarm-send-lease-20261018`，从产品基线创建；主 Agent 维护本计划、状态和 PR review，不修改业务代码，不安排第二个开发或集成 Agent。

## 功能目标

沿 SPEC §7.2、§10.4、§12.5–12.6、§17.2–17.3，补充 Swarm 发送前对当前租约、fence、成员绑定、原准备身份和 operation 的校验。租约丢失、过期、已停止、授权撤销或绑定变化时，发送必须在副作用前失败；未知交付不重发。消费方端口属于 app，SQLite 与外部事实读取属于 adapter，启动入口仅做组装。独立 Works Relay 与 Swarm 的权限规则保持各自已确认范围。

当前 `SendNativeMessage` 消费固定准备，但尚未消费 S3 租约；生产准备/发送依赖尚未启用。开发 Agent 先读实际发送与认领链，向主 Agent 交接完整方案、精确接口及写域，明确如何识别 Swarm 发送且防止绕过、如何避免事务内外部调用、如何处理授权与发送的竞态。主 Agent 依据既有规格 review 后实施。不能从缺少租约推断为独立发送，不能把历史租约当当前执行权限。

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

Orca Run 沿用 `run_73578b70c449`，主控 `term_66e5a559-ee45-4d6f-8a1a-8bffa7996608`。新 Task / Dispatch、客户端实际模型、工作树和 PR 地址在启动后登记。

当前 Task `task_1795bc720ae3` / Dispatch `ctx_7e6f00eb7fdb`；工作树 `C:/Users/DW/orca/workspaces/Astrocyte/swarm-send-lease-20261018`、同名分支；终端 `term_8eb15cea-cc6b-4b7f-8c33-ddf81cf3fcb1`。Orca 1.4.220 / Codex 0.162.0，终端显示实际 `GPT-6.1-Sol xhigh fast`，沿用户已有配置启动，无模型覆盖。启动回执为 turn_start_unobserved，已实际读取到任务停在 composer，随后只补发一次 Enter，没有重复任务。PR 尚未创建。

18:16 开发 Agent 交付方案 `msg_c3b02ba4e1f7`，主控 review 后同意在原写域实施：明确 project/swarm 准备入口并持久化用途；Claim 仅消费 Swarm 准备，同事务首次固定原 fence；发送在当前 Grant/绑定/Mission/WorkItem/Lease 授权事务中 CAS `send_attempted`，所有外部观察和 Send 位于事务外。用途未知不默认、历史 Swarm 标记不因租约消失而删除或降为 project，私有字段不公开。发送占用提交作为动作开始点，提交后撤销属于在途，不声称跨 SQLite/原生控制器的原子提交；发送前传递最早绝对截止时间与取消，失败仍保留单次占用，未知不重发。完整方案保存在开发树 `docs/standards/swarm-send-lease-20261018.md`；生产依赖不启用，原生正向未验时创建 draft PR，主控 review 后由同一开发 Agent 返修。

20:39 接续时 Orca runtime 已变为 `335b14a9-75ca-4611-83f6-27a2ac91dbcd`，当前主控为 `term_d7578194-fd20-42b4-9668-b0b0e8683571`，同 Run generation3。原 Dispatch 由 execution_host 明确记录 failed / process_exited，14 件未提交修改与原验证脚本保留，没有 PR 或已完成测试日志。相同 Task 在同一工作树、分支恢复为 `ctx_3bdd3825f01c` / `term_6b6909b0-97b7-4e35-a45a-601f887c15d3`，继续已批准方案，不重复设计决定。旧终端官方 release 两次返回 `release_unknown`，记录终端已无法解析；保留该不确定清理状态，不扩大关闭范围。
