# DDD 整改普通合并记录（2026-10-18）

## 分支与准确来源

本次唯一集成 Agent 使用 Orca / Codex，在现有 `C:/Users/DW/orca/workspaces/Astrocyte/works-relay-integration` 工作树执行任务 `task_ce7309910dab` / Dispatch `ctx_f87e16719af4`，所属 Run 为 `run_73578b70c449`。接管时工作树干净，分支为 `songconmaisaix31-design/works-relay-integration`，HEAD 与指定基线完全相同。

| 提交用途 | 完整 SHA |
|---|---|
| 产品集成基线 | `8c8510426b828359f4c161f93c1edf1083fd5e88` |
| D SOURCE | `0b585160541cc8f8e29eaffea699c29690b50d03` |
| T SOURCE | `0ce1161de0de2a344a6f33a8442b9874ca5768bd` |
| P SOURCE | `b8d89fa2f9c4cdf725a740873db424bbe249cb51` |
| 开发分支 `ddd-domain-20261018` 报告 HEAD | `f7007d2d132d3b0585ebe9777b2a377174e78d7c` |
| 本次普通合并提交 | `1674075a803225e82f69436a6dc363d7b8d570ab` |

四个开发提交位于同一仓库，按 D → T → P → 报告依次继承产品基线。已核对完整差异、来源分支和三份报告；本地开发分支与其远端跟踪引用均指向上述报告 HEAD。

实际执行 `git merge --no-ff f7007d2d132d3b0585ebe9777b2a377174e78d7c -m 'merge: integrate reviewed DDD remediation 20261018'`，EXIT 0，无冲突。合并提交的第一父提交为产品基线，第二父提交为指定报告 HEAD；合并后的文件内容与该报告 HEAD 完全相同。本次集成追加的文件仅为本报告。

## 合入的变更

来源分支相对产品基线变更 13 件文件，增加 550 行、删除 12 行：

- D：Swarm app 增加 `PreparedMemberInputReader` 与准备身份校验，agents adapter 映射原准备身份，SQLite bridge 在首次缺租约时消费该端口，服务入口组装同一 Registry 和实际 Workspace 准备读取端口；同步三件测试文件，共七件 Go 文件。
- T：`scripts/check-architecture.mjs` 扩展至 Workbench Domain/app 和共享 judgment，保持现有传输层、I/O 与跨业务依赖限制。
- P：`release/Start-Astrocyte.ps1` 接受四项成组 Interpreter 显式配置，`docs/release/windows.md` 记录参数、路径和 CMD JSON 传参方式。
- 三份来源报告：[D 领域报告](ddd-domain-20261018.md)、[T 验证报告](ddd-validation-20261018.md)、[P 发布配置报告](../release/ddd-release-20261018.md)。

## 作者已经验证的结果

以下结果来自上述三份作者报告；主控已审查其真实验证原件。本集成 Agent 没有重复执行这些命令。

| 作者实际命令或操作 | 来源与结果 |
|---|---|
| `go test -mod=readonly -p=1 ./internal/adapters/agents -run '^TestRelayPreparedMemberRejectsMissingOriginalSources$' -v -count=1` | D SOURCE；EXIT 0，1 PASS / 0 SKIP |
| `go test -mod=readonly -p=1 ./cmd/server -run '^TestRelaySwarmActualPreparedAssemblySQLiteHTTP$' -v -count=1` | D SOURCE；EXIT 0，1 PASS / 0 SKIP；真实 SQLite、HTTP 和缺来源拒绝 |
| `go vet -mod=readonly -p=1 ./...` | D SOURCE；EXIT 0 |
| `go build -mod=readonly -p=1 ./...` | P SOURCE；EXIT 0，其 Go 代码与 D SOURCE 相同 |
| `node scripts/check-architecture.mjs` | P SOURCE；EXIT 0，42 包、7 个 Domain/app、1 个共享 judgment；脚本与 T SOURCE 相同 |
| `node --check scripts/check-architecture.mjs`、七件 Go 文件的 `gofmt -l`、`git diff --check` | 作者报告均为 EXIT 0 |
| Windows PowerShell 5.1 AST 解析 | 通过 |
| 发布入口参数及实际应用路径拒绝 | 18 PASS / 0 SKIP；拒绝路径环境恢复通过 |
| PowerShell 显式参数、CMD 显式参数、默认参数三种实际启动 | HTTP health、首页、session/CSRF、空 Mission/Relay 与真实 SQLite 检查通过；服务全部停止 |
| 实际服务关闭 | CMD/default 服务退出码为 0；PowerShell 正常关闭日志、进程退出和端口释放已核，服务退出码未知 |

发布启动程序由 P SOURCE 构建，`vcs.modified=true`，构建时仅两份报告草稿未跟踪；静态网页复用既有 `41cefd00706ae284dbb45f857c2c49d2aebd807f` 来源。正向启动的宿主环境恢复没有取得返回标记，自动项目发现仍有警告；这两项没有登记通过。作者原件位置及准确执行入口保留在三份来源报告。

## 本次集成检查

| 实际命令 | 结果 |
|---|---|
| `git merge-base --is-ancestor 8c8510426b828359f4c161f93c1edf1083fd5e88 f7007d2d132d3b0585ebe9777b2a377174e78d7c` | EXIT 0 |
| `git diff --check 8c8510426b828359f4c161f93c1edf1083fd5e88 1674075a803225e82f69436a6dc363d7b8d570ab` | EXIT 0，无输出 |
| `git diff --exit-code f7007d2d132d3b0585ebe9777b2a377174e78d7c 1674075a803225e82f69436a6dc363d7b8d570ab` | EXIT 0，无输出 |
| `git status --porcelain=v1`（普通合并后、编写本报告前） | EXIT 0，工作树干净 |
| `git diff --cached --check`（本报告）、`git diff --check` | EXIT 0，无输出 |

## 待主控执行的最终集成验证

本集成阶段未运行 Go 编译、架构包加载、实际服务、模型、原生控制器、前端构建、完整 CI 或模型打包。主控将在本 Worker 完成并释放后，对包含本报告的准确最终 HEAD 执行管理树既有入口：

```powershell
& 'C:/Users/DW/orca/Astrocyte/.verification/ddd-20261018/verify-integration.ps1' -Phase all
```

该入口依次执行全 Go vet/build、扩展架构检查及以下八项现有真实检查；本报告中全部保持未执行：

```text
TestRelaySessionActualEmptyRegistrationPollAndReopen
TestRelaySessionHTTPActualCollectorGuardsAndClosedStore
TestRelaySessionHTTPActualUnconfiguredProviderGuards
TestRelayCompletionExperienceActualAssemblyAndRequestBoundaries
TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown
TestRelaySwarmActualPreparedAssemblySQLiteHTTP
TestRelayPreparedMemberRejectsMissingOriginalSources
TestRelayMemberExecutionRejectsUnownedMetadata
```

最终组合检查使用 `GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB` 和 BelowNormal；每项重操作前两次核 CPU 低于 70%、可用内存至少 4 GiB，间隔 2–5 秒，串行执行。资源不足时保持对应命令未执行。作者通过结果不替代准确集成 HEAD 的复验。

## 真实剩余限制与未执行操作

- 生产 LocalProject 准备/发送依赖没有启用；真实首次 Claim、准备身份正向匹配、租约创建、已有租约原生重放和 Prepare → Claim → Send 未验。发送仍须补充当前 S3 租约及绑定校验。
- archive 没有可信 `PrefixComplete` / `AppendOnly` 来源，完整会话验收未完成。
- Codex 原生文件/命令工具关闭条件尚未满足，保持禁止原生初始化与模型输入；S3 预算、能力证据范围和具体经验批准仍待用户回答。
- 完整默认 `pnpm check`、双平台 CI、默认浏览器、L1/L2、当前源码完整前端及模型包、最终包与正式 Release 均未完成。正式实验未运行。
- 当前共享 64505 服务没有恢复。作者的公开固定 Git、SQLite 和 HTTP 检查只覆盖报告中的实际范围；非空原生会话、首次 Claim、发送和后续经验复用仍未通过。

本报告记录普通合并与来源核查结果；最终报告提交 SHA、普通 push 及工作树状态随本 Dispatch 的 `worker_done` 交付主控。
