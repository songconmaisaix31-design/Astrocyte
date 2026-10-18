# Windows Relay Interpreter 显式配置

2026-10-18，按根 `tasks/ddd-review-remediation-20261018.md` 的 17:17 合并授权，由同一 Worker 在 `ddd-domain-20261018` 分支完成原 P 写域。基线为 `8c8510426b828359f4c161f93c1edf1083fd5e88`，本阶段 SOURCE 为 `b8d89fa2f9c4cdf725a740873db424bbe249cb51`，客户端为 Orca / Codex，沿原终端模型配置。

## 启动参数

`release/Start-Astrocyte.ps1` 支持四个成组参数：

| 参数 | 启动入口校验 | 映射到既有环境变量 |
|---|---|---|
| `RelaySessionDeadlineSeconds` | 显式提供 1–1800 秒 | `ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS` |
| `RelayPublicGitRoot` | 已存在的绝对目录 | `ASTROCYTE_RELAY_PUBLIC_GIT_ROOT` |
| `RelayPublicGitCommit` | 小写 40 位 Git commit | `ASTROCYTE_RELAY_PUBLIC_GIT_COMMIT` |
| `RelayPublicContextFiles` | 标准 JSON 数组，1–32 个非空字符串 | `ASTROCYTE_RELAY_PUBLIC_CONTEXT_FILES` |

任意一个参数存在时，其余三个必须同时提供；未提供整组时不注入 Interpreter 配置。入口继续清理继承的 `ASTROCYTE_*`，显式参数在清理后赋值，退出时沿原 `finally` 恢复环境。JSON 使用 Windows PowerShell 自带 `ConvertFrom-Json`；完整 JSON 校验与数组属性解析兼容 PowerShell 5.1 的单元素数组行为。

应用继续验证固定根位于自身 `verification` 目录、唯一相对文件路径、固定 Git 身份与原权限。新入口没有选择业务 deadline、文件或授权默认值。`Start-Astrocyte.cmd` 的既有 `%*` 参数转发已满足要求，文件未修改；实际 CMD JSON 转义方式已写入 `docs/release/windows.md` 并执行通过。

## 验证来源

全部中间结果位于本树已忽略的 `.verification/ddd-20261018/`。重操作使用 BelowNormal、`GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`，逐项串行，最多一个实际 HTTP listener；启动前两次间隔 5 秒核 CPU 低于 70%、可用内存至少 4 GiB。资源条件拒绝的尝试保留为未执行，不计入功能失败或通过。

- 实际服务由本 SOURCE 执行 `go build -mod=readonly -p=1 -o .verification/ddd-20261018/launch/astrocyte.exe ./cmd/server` 产生，EXIT 0；Go 1.27.2、Windows amd64、CGO 关闭。
- `go version -m` 记录 revision 为本 SOURCE，`vcs.modified=true`；构建时仅 D/T 两份报告草稿未跟踪，代码没有未提交修改。该产物属于启动验证程序，没有声明为干净正式包。
- 生产静态网页复用既有 `v1017c` 解压目录，来源 `41cefd00706ae284dbb45f857c2c49d2aebd807f`；未重新构建前端、模型或整个 runtime。
- 验证公开项目仅复制当前公开仓库的 `README.md`、`LICENSE`，在本次隔离目录中实际执行 Git 初始化与提交，commit 为 `fe71eed9edde2a52a12891aae9ac1c2ae88819fb`。实际启动显式选择 `README.md`；没有接入用户作品或私人会话。
- 15 项入口参数拒绝使用同一 SOURCE 的启动脚本，以及既有 `41cefd00` 实际程序与网页作为目录前置；拒绝发生在程序执行前，验证数据目录没有创建。随后三项路径拒绝及三种成功启动均使用本次编译程序。

来源和绝对路径保存在 `release-input.json`、`parameter-input.json`，二进制元数据保存在 `release-binary-buildinfo.log`。实际版本为 Windows PowerShell `5.1.26100.9711`；其标准 AST 解析通过，记录在 `powershell-syntax.log`。

## 实际参数与路径拒绝

执行命令：

```powershell
& ./.verification/ddd-20261018/verify-release-negative.ps1 -InputFile ./.verification/ddd-20261018/parameter-input.json -ParameterOnly
& ./.verification/ddd-20261018/verify-release-negative.ps1 -CaseName @('duplicate-path','parent-path','outside-root')
```

两项命令均 EXIT 0，共 18 PASS、0 SKIP：

| 范围 | 实际检查 |
|---|---|
| 参数组 | 分别缺 deadline、root、commit、files，四项均拒绝 |
| deadline | 0、1801 均由参数约束拒绝 |
| JSON | 语法错误、非数组、空数组、数字成员、空字符串成员、33 个成员均拒绝 |
| Git root/commit | 大写 commit、相对 root、不存在目录均拒绝 |
| 实际应用路径约束 | 重复路径、`../README.md` 均返回 `scope_denied: fixed public context requires unique relative file selections` |
| 实际应用隔离根 | 公开根位于隔离范围外时返回 `scope_denied: fixed public context root is outside isolated verification state` |

每项都核实预期拒绝原因，并核实全部原 `ASTROCYTE_*` 与 `LOCALAPPDATA` 环境恢复。实际应用三项拒绝以进程退出码 1 终止，外层验证将其作为预期结果。完整输出与汇总分别使用 `20261018T092816283-*` 和 `20261018T093049008-*` 文件名。

初次运行的辅助脚本在 Windows PowerShell 5.1 解析 UTF-8 无 BOM 的中文字符串时失败，产品程序没有执行；该记录保留在不带时间前缀的 `missing-deadline-*` 文件。当前辅助脚本使用 ASCII 错误字符串，实际通过结果保存在上述独立文件。

## 实际启动、HTTP、SQLite 与关闭

执行入口分别为：

```powershell
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File ./.verification/ddd-20261018/start-live.ps1 -Mode powershell
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File ./.verification/ddd-20261018/start-live.ps1 -Mode cmd
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File ./.verification/ddd-20261018/start-live.ps1 -Mode default
```

`powershell` 与 `cmd` 两种模式显式提供 deadline 30 秒、公开 Git root、准确 commit 和单元素 JSON 文件列表。`default` 模式只传原有 Port/DataDir。三种模式都先设置四个无效继承环境变量；实际应用成功启动，证明显式参数覆盖及默认清理生效。

每次启动后执行 `probe-live.ps1 -Mode <实际模式>`，通过真实 `http://127.0.0.1:17921` 检查 health 200、生产首页 200、实际人工 session/CSRF、空 Mission 列表和空 Relay 事件及 coverage。分别记录在 `powershell-http.json`、`cmd-http.json`、`default-http.json`。

`node ./.verification/ddd-20261018/read-launch-state.mjs` 使用标准 `node:sqlite` 只读实际数据库；`swarm_missions`、`swarm_leases`、`local_agent_sessions`、`workspace_session_bindings`、`attention_engine_events` 均为零行，结果保存在 `launch-sqlite-counts.json`。这些空值只验证本次未产生原生会话或租约，不能代表非空 Interpreter 路径通过。

| 启动模式 | 实际关闭结果 |
|---|---|
| PowerShell 显式参数 | Ctrl+C 后服务记录 `shutdown signal received`、`HTTP server shutting down`、`server stopped`；PID 41136 退出且端口释放。没有预先持有进程句柄，未取得该服务退出码；宿主终端因中断退出为 1 |
| CMD 显式参数 | Ctrl+C 后服务正常关闭；独立持有的进程句柄确认 PID 41944 退出码 0，记录在 `cmd-process-exit.json`；CMD 中断提示处理后宿主退出为 1 |
| 默认参数 | HTTP 检查通过后，对已核准自身程序路径、父子关系及全部附加 console 进程的控制台发送实际 `CTRL_C_EVENT`；PID 26524 退出码 0，记录在 `default-process-exit.json`；端口释放 |

正向启动的宿主环境恢复检查没有取得返回标记，未声明该项已通过；18 项拒绝路径的环境恢复均已实际通过。所有验证服务已关闭。

## 真实剩余限制

成功启动时仍出现 `registered project discovery failed; prior cache retained` 警告，自动项目发现未验收。既有启动流程执行公开 CLI version/help 探测，验证目录中可能保留这些 CLI 产生的运行文件；本轮没有发起原生会话初始化、模型推理或私人会话读取。

本次仅完成启动配置、真实参数拒绝、空 HTTP/SQLite 和正常关闭检查。首次真实 Claim、Prepare → Claim → Send、archive 完整会话、L1/L2、完整默认 CI、默认浏览器、当前源码前端及完整模型包验收均未完成；本轮没有运行正式 V4 或发布 Release。对应原生工具关闭、S3 预算与能力、可信 archive 来源前置继续按 D/T 报告和根 QUESTIONS 处理。

最终合并源码的组合验收由集成负责人执行；正式包仍须从准确最终 SOURCE 构建并按既有完整发布条件验证。
