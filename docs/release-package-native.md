# Windows ZIP 原生验收客户端

N 独占 `tests/release-package-native/**` 和本文；P 继续负责 `release/**` 的构建、启动、认知快照和包说明。客户端沿用仓库 Go module、`encoding/json`、`archive/zip` 与现有 `modernc.org/sqlite`，产物与原始响应仅写入当前工作树已忽略的 `.verification`。

## 调用入口

在仓库根目录编译：

```powershell
$env:GOMAXPROCS = '2'
go build -p=1 -mod=readonly -o .verification/native-client.exe ./tests/release-package-native
go vet -p=1 -mod=readonly ./tests/release-package-native
```

P 启动实际 ZIP 解压目录中的 `astrocyte.exe` 后调用 `tests/release-package-native/Invoke-NativeClient.ps1`。参数 `Client` 是编译产物，`ServerProcessId` 是实际服务进程，`Executable` 是该进程对应的解压文件，`Zip` 是实际 P ZIP，`PackageSource` 是 P 声明的完整 40 位源码提交，`Base` 是独立 `http://127.0.0.1:端口`，`VerificationRoot` 是独立 `.verification`，`Output` 是该目录下尚不存在的本轮动作目录。

PowerShell 入口核查实际进程文件与该端口的监听者。Go 客户端通过 ZIP 库读取 `Astrocyte-v0.1-windows-x64/astrocyte.exe`，与实际解压文件逐字节比较；每次调用均重新检查。`PackageSource` 记录 P 的来源声明，代码没有从 ZIP 内容独立推导源码提交。保留这些限制，不将来源声明当作构建追溯的独立证据。

服务启动与停止继续由 P 负责。N 客户端通过实际 HTTP intent/confirm 请求服务控制器，CLI 命令与 Agent 注册由 H 的准确配置提供。17321、17342、17343 被入口明确拒绝。

## 可复用动作

| Mode | 额外参数 | 必须通过的检查 |
|---|---|---|
| `baseline` | 无 | 同 URL 的真实 session bootstrap；注册 Agent 为零；任务和已派发作品均为零 |
| `registry` | 无 | 同 URL 的注册 Agent 恰为一个；任务和已派发作品均为零，注册保持数据操作 |
| `idle` | 无 | 任务和已派发作品均为零 |
| `task-state` | `Task`、`Project` | 真实任务恰为该 owned task、绑定一致、单一派发作品和一条 accepted dispatch 回执；输出标识符供停止及重启前后核查 |
| `request` | `Method`、`Path`；写入另传 `BodyFile`；可设 `ExpectStatus` | 真实 Cookie/CSRF、Origin、Idempotency-Key；指定状态；完整响应仅本机封存 |
| `events` | `Database`、`RelayDatabase`、`Task`、`Project`；可设 `MinimumEvents` | 分页选定任务；HTTP 每条精确六字段；每个 event_id 在实际 SQL 中定位并比较字段、毫秒时间和项目范围；核查 owned binding 与自产原件范围 |

`request` 只允许读取 Relay 状态、local-projects、registered 元数据及 Works，或写入既有项目注册、项目 settings、Works refresh/delivery/handoff、Relay intent/confirm/stop。输入 JSON 文件必须放在独立 `.verification`。普通 command envelope 由调用者按实际 API 提供；存在 `request_id` 时，客户端将同一 UUID 用作 `Idempotency-Key`。Relay command 不添加普通 command envelope。

每次动作先执行 `/api/v1/health` 和 `/api/v1/auth/session`，以真实 CookieContainer 和返回的 CSRF token 授权。客户端拒绝 HTTP redirect，禁用 proxy 与连接复用，无自动写入重试。写入发送前保留 operation UUID；传输失败或异常 HTTP 状态保留 UNKNOWN，调用者必须读取原回执并检查本机记录，不得自动再次派发。原始 bootstrap 含 CSRF，只保存在本机证据目录，禁止提交或复制入包。

只读 `events` 每页最多 100 条、最多 1000 页；读取期间 total 变化立即失败，调用者可在原生窗口关闭后显式再读取。数据库必须是独立 `.verification` 内已经存在的文件，使用 `mode=ro` 和 `PRAGMA query_only=1`；mutable live 数据库没有 `immutable`。SQL 查询使用绑定的 event_id，读取 `attention_engine_events` 的对应行，在进程内比较派生 JSON，stdout 只导出安全计数、coverage 与本机指针。原始 SQLite 内容不进入 stdout。

`RelayDatabase` 是同一独立数据目录中的实际 `relay.sqlite`。客户端仅选择指定任务的 owned binding、私有路径元数据和任务元数据，核查 task/project/terminal/worktree、owned、独立公开 Git/worktree 边界。随后复用现有 `events.ReadRelaySession` 读取 binding 精确选定的 Codex 自产新会话文件；不搜索历史、不复制会话、数据库或 CLI 鉴权。文件修改时间不得早于该实际 intent，header 必须匹配实际 native ID 和工作目录。archive 来源的每条 SQL origin 必须对应该原件中实际解析的派生记录；HTTP 的六字段必须一致，HTTP archive 事件总数必须与该观察范围投影数一致。原件 unknown、errors、tail、remaining、snapshot_complete、session_complete、archive append/final 原值进入安全汇总。当前该动作要求实际 archive，`native_notifications` 原件范围比较尚未实现，会立即失败并保留真实限制。

P 提供 Zip/PackageSource/Executable/实际 ServerProcessId/Base、独立数据目录、确定的公开任务与 CLI/窗口参数；N 负责在自己的 `.verification` 内运行实际测试与返修。N 从本次 bootstrap 得到 human Caller，P 不提供 Cookie、CSRF 或 CLI credential。项目、Works 与任务标识取自本次真实 HTTP 响应；输入文件中的 manual revision 取自实际项目状态。

## 已读取的现有接口

基线读取源码为 `c54bcf35162918dab1c55cfa2f0ba6d97ce931ed`。Agent 登记由 `cmd/server/relay_config.go` 在启动时读取独立 access root 的 `agents.json`，HTTP 不提供该登记的写接口。支持环境变量为 `ASTROCYTE_RELAY`、`ASTROCYTE_RELAY_PUBLIC_GIT_ROOT`、`ASTROCYTE_RELAY_SESSION_SOURCE`、`ASTROCYTE_RELAY_LIVE_SECONDS`；CLI argv、公开固定任务、来源模式和实际 live 秒数等待 H 的精确安全参数。`SessionDeadlineSeconds` 是 trusted assembly 字段，其配置不能通过本文增加 ENV。

| 操作 | 现有 API |
|---|---|
| 真实人类 session | `GET /api/v1/auth/session` |
| 项目注册 | `POST /api/v1/local-projects` |
| 人工项目权限 | `PUT /api/v1/local-projects/{id}/settings` |
| 读取注册和项目权限 | `GET /api/v1/relay/agents`、`GET /api/v1/relay/projects` |
| 创建人工 intent | `POST /api/v1/relay/intents`，`work_id/project_id/agent/task` |
| 人工确认 | `POST /api/v1/relay/intents/{id}/confirm`，`permission_confirmed/permission_level/project_settings_revision` |
| 读取任务和原回执 | `GET /api/v1/relay/tasks/{id}`、`GET /api/v1/relay/tasks/{id}/receipts` |
| 手工停止 | `POST /api/v1/relay/tasks/{id}/stop`，空 JSON 对象 |
| 已提交派生会话事件 | `GET /api/v1/relay/session-events?task_id=…&page=…&page_size=100` |

事件公开字段严格为 `event_id/kind/topic/tokens/scope/ts`。客户端保留 coverage 的 `partial/unknown`、gap、错误与未知计数、`snapshot_complete`、`session_complete` 原值。HTTP/SQL 字段一致不能证明 DOM、原件全范围或完整原生会话；这些结果分别保留 `NOT_RUN`，直至实际 H/F 窗口和原件核查完成。

H 已交固定公开正文：`Public fixed context L1_PUBLIC_CANARY_976E. This public project checks native session event capture. Explain in three concise sentences why a newline boundary and an owned CWD matter. Do not use tools, read files, execute commands, search, or inspect any history. Use only this explicit public context.` 新原生窗口的 effectiveToolSet/MCP/skills/视觉关闭证据、CLI actualargv、最终源码与来源范围仍等待 H 的实际闭环结果。

## 当前验证记录

2026-10-14：Go build、Go vet、`git diff --check` 均退出 0；实际 PowerShell `5.1.26100.9711` 解析通过。PowerShell 源文件使用 UTF-8 BOM，N 本人通过该版本实际调用客户端。没有新增依赖或修改 P 的客户端。

协调者明确提供早期基础包 `c7a70be963c432bb72d0fa355705ce2907821120` 进行无模型客户端验证。`Verify-BasicClient.ps1` 使用该 ZIP 原有 `Start-Astrocyte.ps1`、空身份目录、系统 PATH、N 独立数据和端口 17896。实际 ZIP 为 P 工作树 `.verification/package-rehearsal-20261014-202942/Astrocyte-v0.1-windows-x64-c7a70be963c4.zip`，26,309,211 bytes。

本机 `.verification/basic-http-20261014-223940/basic-safe.json` 和其 `protected-read/summary.json` 记录 PASS：已知私人路径匹配 0，运行的解压 `astrocyte.exe` 与真实 ZIP 成员逐字节一致，实际进程与监听端口一致，health 200、human session bootstrap 200、真实 Cookie/CSRF、受保护 `GET /api/v1/local-projects` 200。实际服务进程和端口均已关闭，写入请求和模型调用均为 0。仅保存安全聚合及本机指针，原始 bootstrap 与 server 日志留在忽略目录。

本任务要求保留的客户端首失败原件分别位于 `.verification/basic-http-20261014-223748` 与 `.verification/basic-http-20261014-223837`；两轮均已关闭 owned 进程和端口。该早期 ZIP 不含 Relay，Relay baseline、registry、事件 SQL/原件范围、原生模型、原生停止及重启、DOM 仍为 `NOT_RUN`。最终 I 源码、P fresh ZIP 和协调者明确的原生窗口尚未到达。

旧候选 c54 privacy scan 的 server 三项匹配 FAIL、原生首确认 503 与 accepted 后 working/exited 的 UNKNOWN 继续由原 owner 的报告保留。本文的编译结果没有改变这些实际验收结论。
