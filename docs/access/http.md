# Works Relay HTTP 与 L1 会话采集

H 分支：`songconmaisaix31-design/works-relay-http`。源码：`3016a5369393166ac3a40f520138ae352d2edf92`，通过普通合并消费 R 的 `fbfaa65/53231b19`、意图版本 `ce3a519/9719e69`、原生解码 `d6d4da6/07449a6`。本次实际服务的业务构建为 `099ce0b`，报告合并 `b73f749` 的业务相同；后续 H 改动是原件策略核验和既有 Foundation/Projects 构造器组装，尚未替换原运行服务。客户端为 Codex API，工作会话没有暴露精确部署模型名称；实际原生控制器为 Codex `0.162.0`，本次 `turn_context.model` 为 `gpt-6.1-sol`。

## 当前实现

所有 Relay HTTP 使用既有 Go 服务、loopback、human session、Origin、CSRF 与 ErrorV1。GET 读取已持有任务、派生事件和原收据，不启动 CLI、不发送输入、不恢复会话、不读取原文归档。人工创建 intent 后必须明确确认权限；运行时 Service 是 task、terminal、worktree 与收据的唯一写入者。未知输入交付结果保持 `delivery_unknown`，不能自动重发。

`agents.json` 使用 `encoding/json` 解析受信任的 `AgentSpec` 数组，在启动时加载。条目包含 `name/launch_cmd/capabilities/drive_mode`，可选 `input_mode` 为 `line` 或 `stdin_eof`。新增条目后正常关闭并重启同一服务，调度按注册数据选择既有原生适配器；注册本身不授予启动或工具权限。

`cmd/server/relay_assembly.go` 复用实际 Works、既有项目权限与 NativeRegistry、独立 relay.sqlite、既有 PendingEventService 和 EventRepository。observer 只观察人工已接受的任务，取消并 join 后关闭自有进程和存储。静默重读不更新活动时间或持久化未变化的游标。

`GET /api/v1/relay/session-events` 按当前 owned binding 与权限读取已经提交的真实事件。事件只有 `event_id/kind/topic/tokens/scope/ts` 六字段，时间倒序，ID 升序，分页最多 100 条。未知 tokens 保持 null。任务级事件使用仓储提交 ID，归档记录使用稳定 origin 查得的提交 ID；公开查询没有 prompt、reasoning 原文、工具参数、stdout、归档路径或 SQL 错误。

## 实际来源与覆盖

`ASTROCYTE_RELAY_SESSION_SOURCE` 未配置或为 `archive` 时采集控制器提供的准确归档路径。H 通过 `github.com/olivere/ndjson` 读取该实际 JSONL，校验 owned task、terminal、session、原生 ID、原 CWD 和文件身份；增量位置只在事件 admission 成功后持久化。未知记录仍计入覆盖，格式错误立即失败，未完成尾行保留待追加范围。

选择 `native_notifications` 时显式启用 R 的原生观察端口，从控制器持有的通知序号派生事件。其覆盖标为 `source=native_notifications`；归档覆盖标为 `source=archive`。通知和归档分别保留来源语义，通知的前缀或关闭证据不证明归档文件永久追加。

`snapshot_complete` 只表示当前选定范围的完整行已覆盖。归档的完整前缀、永久追加性和关闭证据不足时，`session_complete=false`，公开状态保留 partial 与具体 gap。R 的真实原生结果保留 `ArchiveAppendOnly=unknown`：当前接口没有永久 append-only 保证，原生会话允许 rollback/revert/压缩。恢复后的 raw 通知订阅缺口为 `native_resume_raw_events_unavailable`。进程退出、输入接受、任务完成和模型成功分别解释，不互相替代。

## 本机配置

接入状态位于 `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access`，注册文件为其中 `agents.json`，应用状态为 `app-state/state.sqlite`。启用接入模式用 `ASTROCYTE_RELAY=1`。`ASTROCYTE_RELAY_LIVE_SECONDS` 没有默认值；未配置时活动三值保持 unverifiable，不自行决定活性窗口。

`ASTROCYTE_RELAY_PUBLIC_GIT_ROOT` 只用于明确获准的 Agent 自建公开 Git 验证项目。目录必须位于独立接入根的 `verification` 内，每次 Prepare 都以 `os.SameFile` 核对该唯一目录；其他目录立即拒绝。未配置时公开项目启动权限保持关闭。既有中性 worktree 仍核对干净、已提交、正确 Git 根和原身份，不修改用户作品或开放自由工具。实际 Codex 使用现有控制器关闭文件、命令、MCP、skills、浏览器、视觉和其他工具，只提供固定公开上下文。

实际验收服务使用生产 `assembleRelay`、真实 SQLite、真实 HTTP 和本次控制器自己产生的会话文件。API 为 `http://127.0.0.1:17342`，F 网页为 `http://127.0.0.1:17343`。本次数据根为 `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/h-production-current-20261014`；项目、worktree、数据库和进程均由本次验证持有。所有脚本、中间文件和安全聚合位于 H 工作目录内被忽略的 `dist/l1-live/current-20261014`。原件不进入 Git、HTTP、stdout 或 release 包。官方 CLI 自行鉴权，不读取或复制凭据；归档只读取控制器给出的本次 thread 路径，不枚举历史。

## 验证命令与当前结果

| 命令 | 结果 |
|---|---|
| `go build -mod=readonly ./cmd/server` | PASS |
| `go test -mod=readonly -c ./cmd/server` | PASS，实际服务与证据测试二进制已编译 |
| `go vet -mod=readonly ./cmd/server ./internal/adapters/events ./internal/adapters/httpapi ./internal/adapters/sqlite ./internal/workbench/app` | PASS |
| `node scripts/check-architecture.mjs` | PASS，29 个 Go 包边界 |
| `pnpm exec redocly lint contracts/openapi.yaml --config contracts/redocly.yaml` | PASS，保留既有 EventV1 未引用警告 |
| `node scripts/validate-examples.mjs` | PASS，291 个契约示例 |
| `git diff --check` | PASS |
| 实际服务无 Codex 注册基线 | PASS，17342 空注册表，无模型派发 |
| F 注册前真实浏览器基线 | F 于 `2026-10-14T12:53:45Z` 报告两尺寸 session 200、空 registry 与浏览器 ready |
| 实际注册前 HTTP 权限检查 | PASS，无 session、跨 Origin、未授权 bearer、缺 CSRF 均 403 |
| 本轮 SQLite 独立只读计数 | 实际 events=3、signals=0、nodes=0；原生派发前的项目 / Works 观测 |
| 同 URL 追加登记与正常重启 | 实际执行一次，原始 `registered_at=2026-10-14T13:23:34.0366105Z` |
| 注册起点 → 原生事件首次 DOM | FAIL：F 两尺寸首显示 `13:33:25.152Z/.153Z`，591116/591117 ms，大于 180000 ms |
| 原 dispatch 收据创建 → 原生事件首次 DOM | PASS：创建 `13:33:10.6363218Z`，约 14.516/14.517 秒；独立于注册起点 FAIL |
| 实际原件与提交 ID | PASS：15 个完整行、105564 字节、15 个原档 origin 对应提交 ID；全部 18 个 HTTP 事件 GetEvent 一致 |
| 实际文件范围 | `snapshot_complete=true`、尾行 / 剩余字节 0、unknown 5、errors 0、`session_complete=false` |
| 实际回答与结束轮次 | `gpt-6.1-sol`，assistant 1、task_complete 1、已报总 tokens 8827；没有重复原输入 |
| 实际权限、分页与 GET 只读 | PASS：四项 HTTP 403、五项实际文件身份 / 权限拒绝、全分页、canary 排除、静默 GET 事件数不变、持久游标相符 |
| 实际原件中的工具与策略 | 当前文件工具调用 0、`never/read-only`；旧 owned PID / 启动参数未当时采集，effective tool set 保持 unknown，不能仅用 tool 0 填禁用证据 PASS |
| 连续追加 / 明确 Stop 后范围 | 未执行，原控制器已在默认 180 秒后实际退出，旧输入没有自动重发 |

实际证据检查逐原档 origin 查询 `attention_engine_events` 并通过 `EventRepository.GetEvent` 定位提交 ID，独立比较 HTTP 六字段、分页和 canary 排除；核对重读及持久游标、追加与停止后的真实文件范围。实际 HTTP 权限检查包含无 session、跨 Origin、未授权 bearer 和缺 CSRF；静默 GET 前后事件数量必须一致。模型名称与工具调用数量仅从选定实际文件生成安全聚合。

## 剩余限制

本报告为实际已执行范围的阶段报告，H 保持同一活动 Task 继续真实闭环。F 是主任务唯一产品 UI 操作人；H 负责注册文件与正常重启，并只读核对原档、SQLite 和 HTTP。原 task 为 `0c6522bc-b232-4040-8407-745885644791`，仅确认一次。HTTP 503 与 dispatch `accepted + provider_unavailable` 原收据保留；实际安全原因是 `thread/read / -32603 / thread_rejected`，更具体的服务内部原因未知。之后原 observer 实际取得正选归档，原件回答成功不覆盖首 503，不授权重发。

F 首 DOM 三个 ID 均在真实 SQLite 定位，其中 `MBV3V2HETK5753LFVJY4PGVJWL` 和 `HFZSCX6TFXYYMOU5VBUNG65QIV` 分别是原档 `task_started` 与 `native_session_header`，第三个是 task 级 `turn_started`。安全证据为本地忽略目录的 `committed-public.json`、`initial-actual-evidence-derived.json`、`initial-actual-model-derived.json` 和 `first-dom-sql-derived.json`。14 个原档记录的 tokens 为真实 null，未填值。原 controller 文件路径只保存在本次 relay.sqlite 的私有绑定中。

本次旧项目权限没有 `reconcile`，因此 done/release 未获准；不能更新项目后覆盖旧 task 的绑定版本、手改状态或把退出自动当完成。Foundation/旧 Projects GET 的未组装异常已保留本机原日志，源码已补既有生产构造器，原服务尚未串行替换核验。后续 R 的寿命、EOF 与 owned 执行策略提供口没有反填旧进程证据。B 新 Reader 接口需要精确窄接缝或 I 组合消费；H 没有整体合入其冻结域旧历史。

完整默认 check/build/E2E、合并前端、Windows ZIP 解压启动和正式发布由对应 owner 与集成验收负责。H 的静态构建和契约检查不能代替实际 Today DOM、完整会话、终端编码或正式部署。

实际用户作品、自由工具授权、默认活性窗口仍未决定。原生文件的完整前缀、永久追加性与恢复 raw 通知接口不足时，保持具体 unknown/partial。原始 Native 共享源码、mapper、engine、schema/迁移、七类权重及快照流程没有 H 修改。V4 协议、库和运行资料没有读取或修改。管理员 Procmon、两次实际 00:30、正式认知验收和现有 17321 部署均未由本次 H 验收执行。
