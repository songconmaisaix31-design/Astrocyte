# DDD 检查入口与真实验证前置

2026-10-18，按根 `tasks/ddd-review-remediation-20261018.md` 的 17:17 合并决定，由同一 Worker 在 `ddd-domain-20261018` 分支完成原 T 写域。基线为 `8c8510426b828359f4c161f93c1edf1083fd5e88`；本阶段 SOURCE 为 `0ce1161de0de2a344a6f33a8442b9874ca5768bd`，之前的 D 源码为 `0b585160541cc8f8e29eaffea699c29690b50d03`。客户端为 Orca / Codex，沿原终端模型配置，不派子 Agent。

## 检查边界

`scripts/check-architecture.mjs` 读取当前实际 Go import graph，按以下范围检查：

- `attention`、`workspace`、`swarm`、`workbench` 的 Domain/app 禁止 I/O、adapter 和跨业务上下文导入；Domain 保留原本上下文和纯标准库约束。
- 共享 `judgment` 禁止 I/O，内部依赖仅允许自身、`apierrors` 和其既有契约实际使用的 `workspace/domain` 值类型；不允许依赖其他上下文应用服务或 adapter。
- HTTP/MCP 继续禁止访问 SQL、SQLite adapter 与 modernc 驱动。
- Go 自身的包加载失败与循环依赖错误直接终止；输出实际检查过的 Domain/app 与共享 judgment 包，便于核查当前覆盖范围。

没有修改 `scripts/check.mjs` 的完整检查列表或 `.github/workflows/check.yml` 的双平台流程，也没有新增替身、Skip、条件移除或降低已有断言。此阶段没有新增业务组件、依赖或 HTTP 契约。

## 实际检查

| 命令 | 结果 |
|---|---|
| `node scripts/check-architecture.mjs` | EXIT 0；42 包、7 个 Domain/app 包、1 个共享 judgment 包 |
| `node --check scripts/check-architecture.mjs` | EXIT 0 |
| `git diff --check` | EXIT 0，无输出 |

实际执行 SOURCE 为同分支后继 `b8d89fa2f9c4cdf725a740873db424bbe249cb51`，架构脚本与 T SOURCE 完全相同；后继只增加发布配置。输出保存在本树 `.verification/ddd-20261018/architecture.log` 与 `architecture.stdout.log`。按主控 17:27 的当前资源要求使用 `GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`、BelowNormal，重操作串行，每项启动前两次核 CPU 低于 70% 且可用内存至少 4 GiB。软内存目标不作为实际峰值或硬限制保证。

## 完整默认检查仍缺少的真实前置

`internal/adapters/sqlite/workbench_test.go` 要求 `ASTROCYTE_TEST_COGNITION_LOCALAPPDATA` 指向已获准的实际公开快照；从实际 publication 读取并验证时间、事件与 SQL 结果。当前 CI 没有准备该来源。

`web/src/api/relaySession.test.ts` 要求 `ASTROCYTE_TEST_RELAY_PUBLIC_URL` 与 `ASTROCYTE_TEST_RELAY_L1_GETEVENT_PUBLIC_FILE`，连接实际 L1 HTTP 服务并读取真实 GetEvent 事实。当前原生文件/命令工具关闭前置不满足，不能制造非空会话、逐 event_id 结果或用空文件替代。

根 review 还记录实际 OpenCode 配置不足，以及 `cmd/server/relay_runtime_http_test.go` 的 reconcile 权限与原 200 断言冲突。该测试现有实际/替身混合范围需要原负责方完成真实来源整改；本阶段没有运行该旧测试，也没有仅为通过断言添加授权。

`internal/workspace/app/progress_test.go`、`internal/adapters/distillers/selected_text_test.go`、`cmd/server/relay_session_poll_test.go` 的旧替身覆盖仍在默认集合。相关真实模型输出、发布恢复、非空 owned poll 需要其对应真实前置和验收，不能由 D 的缺来源拒绝检查替代。

因此本轮完整 `pnpm check`、双平台 CI、默认浏览器和 L1/L2 均未执行通过。需要原领域负责人完成真实来源和适用测试整改，再由集成负责人运行准确合并源码的完整验收；此次架构检查通过也只代表当前静态依赖边界。
