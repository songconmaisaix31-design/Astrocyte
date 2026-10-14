# Windows ZIP 原生验收客户端

N 独占 `tests/release-package-native/**` 和本文；P 继续负责 `release/**` 的构建、启动、认知快照和包说明。客户端沿用仓库 Go module、`encoding/json`、`archive/zip` 与现有 `modernc.org/sqlite`，产物与原始响应仅写入当前工作树已忽略的 `.verification`。

## 调用入口

在仓库根目录编译：

```powershell
go build -mod=readonly -o .verification/native-client.exe ./tests/release-package-native
go vet -mod=readonly ./tests/release-package-native
```

P 启动实际 ZIP 解压目录中的 `astrocyte.exe` 后调用 `tests/release-package-native/Invoke-NativeClient.ps1`。参数 `Client` 是编译产物，`ServerProcessId` 是实际服务进程，`Executable` 是该进程对应的解压文件，`Zip` 是实际 P ZIP，`PackageSource` 是 P 声明的完整 40 位源码提交，`Base` 是独立 `http://127.0.0.1:端口`，`VerificationRoot` 是独立 `.verification`，`Output` 是该目录下尚不存在的本轮动作目录。

PowerShell 入口核查实际进程文件与该端口的监听者。Go 客户端通过 ZIP 库读取 `Astrocyte-v0.1-windows-x64/astrocyte.exe`，与实际解压文件逐字节比较；每次调用均重新检查。`PackageSource` 记录 P 的来源声明，代码没有从 ZIP 内容独立推导源码提交。保留这些限制，不将来源声明当作构建追溯的独立证据。

服务启动与停止继续由 P 负责。N 客户端不启动 Go 源码服务、不启动模型、不选择 CLI 命令、不修改 Agent 注册、不发现私人会话。17321、17342、17343 被入口明确拒绝。

## 可复用动作

| Mode | 额外参数 | 必须通过的检查 |
|---|---|---|
| `baseline` | 无 | 同 URL 的真实 session bootstrap；注册 Agent 为零；任务和已派发作品均为零 |
| `registry` | 无 | 同 URL 的注册 Agent 恰为一个；任务和已派发作品均为零，注册保持数据操作 |
| `idle` | 无 | 任务和已派发作品均为零 |
| `request` | `Method`、`Path`；写入另传 `BodyFile`；可设 `ExpectStatus` | 真实 Cookie/CSRF、Origin、Idempotency-Key；指定状态；完整响应仅本机封存 |
| `events` | `Database`、`Task`、`Project`；可设 `MinimumEvents` | 分页选定任务；HTTP 每条精确六字段；每个 event_id 在实际 SQL 中定位并比较字段、毫秒时间和项目范围 |

`request` 只允许读取 Relay 状态、local-projects、registered 元数据及 Works，或写入既有项目注册、项目 settings、Works refresh/delivery、Relay intent/confirm/stop。输入 JSON 文件必须放在独立 `.verification`。普通 command envelope 由调用者按实际 API 提供；存在 `request_id` 时，客户端将同一 UUID 用作 `Idempotency-Key`。Relay command 不添加普通 command envelope。

每次动作先执行 `/api/v1/health` 和 `/api/v1/auth/session`，以真实 CookieContainer 和返回的 CSRF token 授权。客户端拒绝 HTTP redirect，禁用 proxy 与连接复用，无自动写入重试。写入发送前保留 operation UUID；传输失败或异常 HTTP 状态保留 UNKNOWN，调用者必须读取原回执并检查本机记录，不得自动再次派发。原始 bootstrap 含 CSRF，只保存在本机证据目录，禁止提交或复制入包。

只读 `events` 每页最多 100 条、最多 1000 页；读取期间 total 变化立即失败，调用者可在原生窗口关闭后显式再读取。数据库必须是独立 `.verification` 内已经存在的文件，使用 `mode=ro` 和 `PRAGMA query_only=1`；mutable live 数据库没有 `immutable`。SQL 查询使用绑定的 event_id，读取 `attention_engine_events` 的对应行，在进程内比较派生 JSON，stdout 只导出安全计数、coverage 与本机指针。原始 SQLite 内容不进入 stdout。

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

## 当前验证记录

2026-10-14：`go build -mod=readonly -o .verification/native-client.exe ./tests/release-package-native`、`go vet -mod=readonly ./tests/release-package-native`、`git diff --check` 均退出 0。没有新增依赖或修改 P 的客户端。基础实际 HTTP、最终 ZIP 原生动作、原件范围、停止与重启验收均为 `NOT_RUN`；最终 I 源码、P fresh ZIP 和协调者明确的原生窗口尚未到达。

旧候选 c54 privacy scan 的 server 三项匹配 FAIL、原生首确认 503 与 accepted 后 working/exited 的 UNKNOWN 继续由原 owner 的报告保留。本文的编译结果没有改变这些实际验收结论。
