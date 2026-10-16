# 可分发认知快照

构建：`go build -trimpath -mod=readonly -o workbench-snapshot.exe ./cmd/workbench-snapshot`。命令属于现有 Go module，使用已有 modernc.org/sqlite 与 golang.org/x/sys，无 Python 或 Node 运行依赖。

```powershell
./workbench-snapshot.exe --source C:/explicit/app-state/state.sqlite --root C:/explicit/local/astrocyte-workbench/v0.1 --timeout 120s
```

`--source`、`--root`、正数 `--timeout` 必填。可选 `--history` 只接收用户显式选择的实际既有历史 JSON；省略时使用本根已发布历史与本次真实节点。历史格式为 `{ "实际主题": [{ "ts": 真实状态毫秒时间, "attention": 实际状态数值 }] }`，时间必须大于零，attention 必须有限且在 0 到 1 之间。未知历史不添加时间点、不插值。

成功 stdout 为 `{ "status":"ok", "generation":"snapshot-...", "snapshot_at":"RFC3339Nano UTC" }`。失败退出码为 1，stderr 固定 `{ "status":"error", "code":"SNAPSHOT_UNAVAILABLE" }`。错误不输出底层 SQL、来源内容或文件路径。

目的根使用既有 `snapshot.lock`、`snapshots/snapshot-*/snapshot.sqlite`、`history.json`、`current.json`。来源通过 SQLite `mode=ro`、`query_only(1)` 和固定读事务在线备份，包含 WAL 中已提交的数据。新副本通过 quick_check 与生产查询所需字段检查后，单独转换为 DELETE journal。历史优先保留显式历史，再继承本根已发布点，最后按实际 node.updated_at 添加缺失点；同一主题与时间确定性去重。来源不运行迁移、更新、checkpoint 或 journal_mode 变更。

拒绝来源位于目的根内、目的目录符号链接或 junction、目的文件与来源/sidecar/显式历史的 hardlink 身份重叠、目的锁冲突及来源身份变化。Windows 使用 LockFileEx 与 MoveFileEx(REPLACE_EXISTING | WRITE_THROUGH)；Unix 使用 flock 与同目录 rename。每阶段检查 deadline，失败不替换 current.json，未发布 generation 留在本机诊断，旧 generation 保留供 immutable Reader 使用。目的根应由同一用户独占写入；目录检查和 OS 锁不能限制其他账户在检查后替换文件。

技术用法已核查 [SQLite Online Backup](https://www.sqlite.org/backup.html)、[modernc.org/sqlite Backup](https://pkg.go.dev/modernc.org/sqlite@v1.60.1#Backup)、[database/sql Conn.Raw](https://pkg.go.dev/database/sql#Conn.Raw)、[Windows LockFileEx](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-lockfileex)、[MoveFileEx](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw)、[GetFileAttributes](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getfileattributesw) 与 [reparse 属性](https://learn.microsoft.com/en-us/windows/win32/fileio/file-attribute-constants)，并读取当前依赖版本的 Backup 实现。

生产 Workbench 接口继续投影 event_id、kind、topic、tokens、scope、ts 六个字段，source 只用于过滤和来源聚合。快照为本机完整来源副本，可能包含原始事件，必须放在用户自身数据目录，不能加入 Git 或发布包。测试产物只位于当前工作树 `.verification/`。

公开来源认知、并发 WAL 备份与原生产 Reader 的实际结果见下列 REPORT。前端与发布包的组合验收归 F/P/I；正式 V4、旧私人历史、管理员 Procmon、两次真实跨日触发没有在本任务运行。

## REPORT：2026-10-16 directory-alias fast-fail 与测试输出路径整改

分支 `songconmaisaix31-design/release-cognition-20261014`，基线 `f1110523c584aa7aa59ad07a05fc70cf88e138b7`。实际客户端 Claude Code / `qwen3.7-max`。根 `msg_e0be901882a5` 授权窄返修；经理 `msg_72aa5148db8b` 确认首改 directory-alias，后续根指令扩展处理集成 `b02d1635828eae98ac6b5dac89cd9383d021257d`（普通合 f19）的真实首次 publish 失败。不执行采集/Sleep/重放/模型或历史重算。

### 修改 1：directory-alias fast-fail

`directory-alias` 子测试按 `runtime.GOOS` 选择实际机制：Windows 直接通过 PowerShell 创建 Junction，非 Windows 直接调用 `os.Symlink`；每个创建错误在位置直接 `t.Fatal`/`t.Fatalf`。生产 `rejectAlias`（`files_windows.go`）通过 `GetFileAttributes` 检查 `FILE_ATTRIBUTE_REPARSE_POINT` 拒绝所有目的祖先的 Windows junction；`files_unix.go` 通过 `os.Lstat` 检查 `os.ModeSymlink` 拒绝 Unix symlink。本轮未修改生产文件。

### 修改 2：ASTROCYTE_TEST_OUTPUT_ROOT

根集成 `b02d163` 的 `TestServerInitializedSnapshot` 首次 publish（原 main_test.go:105）报 `The system cannot find the path specified`，EXIT1/16.563s，七子项未开始；原件 `D:/Astrocyte-agent-assets/20261015/integration/root-snapshot-b02d163/snapshot.log`。根因：集成树的 `repo/.verification` 为 D: junction，测试硬编码输出路径 `repo/.verification/snapshot-tests` 落在 junction 上，生产 `rejectAlias` 正确拒绝该目的根。

修正为支持显式 `ASTROCYTE_TEST_OUTPUT_ROOT` 环境变量：设置时测试输出基目录使用该值的 `snapshot-tests` 子目录，不设时保持原 `repo/.verification/snapshot-tests` 默认。`repo` 仍用于构建 server binary（`build.Dir = repo`），不受影响。没有 fallback、skip、提升权限、削弱断言，不改生产保护、引擎、Clock、原数据或 parser。

### 命令与结果

```powershell
# 首两条命令未设置 D: 环境变量，go vet 使用了默认 C: go-build 缓存
gofmt -l cmd/workbench-snapshot/main_test.go   # 空输出，格式通过
go vet ./cmd/workbench-snapshot/...             # 无错误

# 以下测试命令完整设置 D: 环境：
# GOCACHE=D:/.../go-build/backend
# GOTMPDIR/TEMP/TMP=D:/.../cognition-Q-alias/go-temp/cognition-Q-alias
# GOMAXPROCS=2 GOFLAGS=-p=1

# 首轮（f19 源码，修改 2 前）：定向 directory-alias
go test -mod=readonly -count=1 -v -run TestServerInitializedSnapshot/directory-alias ./cmd/workbench-snapshot/...
# 15.392s 目标包，directory-alias 子项 0.37s

# 首轮全包
go test -mod=readonly -count=1 -v ./cmd/workbench-snapshot/...
# 13.840s，7 子测试全部 PASS

# 第二轮（修改 2 后）：显式 ASTROCYTE_TEST_OUTPUT_ROOT=D:/.../test-output
go test -mod=readonly -count=1 -v -run ^TestServerInitializedSnapshot$ ./cmd/workbench-snapshot/... -timeout=3m
# 18.053s，7 子测试全部 PASS

# 第二轮默认（不设 ASTROCYTE_TEST_OUTPUT_ROOT）
go test -mod=readonly -count=1 -v -run ^TestServerInitializedSnapshot$ ./cmd/workbench-snapshot/... -timeout=3m
# 11.688s，7 子测试全部 PASS
```

7 个子测试：missing-source、source-in-root、missing-history、hardlink、directory-alias（Windows Junction 创建并被生产代码拒绝）、expired、lock。测试日志 `D:/Astrocyte-agent-assets/20261015/cognition-Q-alias/logs/`。

```powershell
# 有效 D: build：GOCACHE/GOTMPDIR/TEMP/TMP 完整 D:，GOMAXPROCS=2, GOFLAGS=-p=1
go build -trimpath -mod=readonly -o D:/.../cognition-Q-alias/build/workbench-snapshot.exe ./cmd/workbench-snapshot
# exit 0；产物 11,432,448 bytes；go version 报告 go1.27.2
```

### 实际限制

- 首条 `go vet` 未设 D: 环境变量，C: `go-build` 缓存因此增长约 314.8 MB；后续测试与有效 build 均已完整设置 D: 环境。
- 独立 symlink OS1314 原未通过保留，Windows 真实 Junction 拒绝通过不计 symlink 通过。

### 根集成首 FAIL 保留

集成 `b02d163` 首次 `TestServerInitializedSnapshot` EXIT1/16.563s，`main_test.go:105` 报 `The system cannot find the path specified`，七子项未开始。原件 `D:/Astrocyte-agent-assets/20261015/integration/root-snapshot-b02d163/snapshot.log` 保留。本 Q 修复后两轮测试（显式 output root 与默认）均 7/7 PASS，根集成复验待根确认。

旧 Task 的跨日 collect、采集、重放、Sleep、模型调用和历史重算本轮一律不执行。原五个 pub 指针、旧 DB、preSleep、历史和首 FAIL 原件保留不改。

## REPORT：2026-10-15 公开工程跨日更新

实际客户端为 Orca Codex / `gpt-6.1-sol`。本轮基线 `6ad1dfa1387d5e6a0b81ac1b2ad2887359780658`，分支 `songconmaisaix31-design/release-cognition-20261014`；SOURCE `40880b67e4a668ec0e19d00eb4f77a76fdf210ab` 已普通 push，独立 `git ls-remote` 返回相同完整 SHA。本 REPORT 单独追加文档提交。

根 `msg_55f7b29dede5` 授权公开工程跨日更新；本派发协调者在实际 ask 回复中确认完整真实采集 → preSleep → 原 Sleep 关闭 14 日 → closed/main 的顺序。获取时公开 `standards-backend` 最新实际提交为 `b789466ff62f3c9b0e08ec392729f5bbfc5cfaa1`。本次新 bare checkout 共享已存在公开对象，再从获准公开 origin 获取该 ref，没有复制大型来源目录；原 checkout 与五份旧发布产物保留。

新库由真实 server 的 HTTP health 200 确认初始化，随后由原 MapGitLogRecord → EventRecorder → EventService/NormalizeSignal → Dynamics 按原始 commit.ts 完整采集该 ref 的 1,316 个唯一可达提交，产生 1,316 个真实 events 与 signals、两个主题。独立 inspector 逐 event_id 定位 SQL，逐 commit 与真实 Git 可达人口及 commit.ts 核对；没有遗漏、重复或人口外事件。上海 10 月 13 日 65 条、14 日 74 条、15 日 52 条。主题 Astrocyte/Attention Engine 134 条、MCP 21 条；其余 1,161 条保持原 mapper 的未分类结果。strength=1、decay=0.1、initial-confidence=0.5，历史 SeedTopic 初始化沿此前已批准的各主题最早真实 Git 时间。

原生产 Sleep 仅选择上海 `2026-10-14` 的事件窗口，实际于 `2026-10-15T02:48:42.101+08:00` 开始、`2026-10-15T02:48:42.110+08:00` 完成；SQL 开始/完成毫秒值分别为 `1792003722101` / `1792003722110`。新库恰有一条 Sleep，Today 的 yesterday status 与完成时间逐项匹配 SQL。closed 与 main 包含完整已采今日事件，保存同一真实终态；closed 的名称表示已完成 14 日闭日处理，其内容不能作为 14 日纯截面。没有更改 Clock、来源时间、engine、mapper、schema、迁移或权重。

实际 writer PID 66000；WAL 在线备份期间已有 2,101,232 bytes，采集进度从 33 条增至 36 条。实际发布如下，所有路径相对于本 Q 工作树 `.verification/cross-day-20261015-0228/`，数据库及原始日志只留本机 ignored 目录。

| 阶段 | LOCALAPPDATA 子目录 | generation | snapshot_at（UTC） | 每主题实际历史点 |
|---|---|---|---|---|
| 空 server 库 | local（保留旧 generation） | snapshot-390490354 | 2026-10-14T18:44:37.6712322Z | 0 |
| 并发 WAL | local（保留旧 generation） | snapshot-1429914530 | 2026-10-14T18:46:48.9033717Z | 1 |
| preSleep | preSleep-local | snapshot-1373353207 | 2026-10-14T18:47:35.9453898Z | 2 |
| main | local | snapshot-1711995690 | 2026-10-14T18:48:42.2097174Z | 3 |
| closed | closed-local | snapshot-3544892468 | 2026-10-14T18:48:42.3242026Z | 3 |

### 准确来源与消费者输入

- 实际 source：上述新目录 `app-state/state.sqlite`，2,555,904 bytes。
- 每个发布根：对应 LOCALAPPDATA 子目录的 `astrocyte-workbench/v0.1/`；pub 为其中 `current.json`，明确指向表中 generation。
- 不可变副本与真实 history：对应发布根 `snapshots/<表中generation>/snapshot.sqlite` 和 `history.json`；P 可以只读选择实际 source 或明确阶段的不可变副本。
- 最终 main 的独立 SQL expected：`main-safe-final/sql-expected.json`；preSleep 与 closed 分别为 `preSleep-safe-final/sql-expected.json`、`closed-safe-final/sql-expected.json`。各文件严格绑定自身 snapshot_at，恰有十值、覆盖四路由，数值直接与生产 Reader 比较全等。
- 对应 safe-final 目录均提供 `actual-counts.json`、`events-six-fields.json`、`origin-safe.json`、`privacy-safe.json`、`production-reader-safe.json`、`cross-day-safe.json`、`admission-safe.json`。三个最终 inspector 均 PASS。

全部 1,316 条六字段逐行/逐页/末尾空页与 SQL 全等，两个主题的全部 A/M/C 与真实排名核对通过，Astrocyte 134 条按 20/50 页大小完整分页通过。两个主题的实际历史时间严格递增、至少两点、最新节点 updated_at/attention 与 SQL 全等；历史继承本次实际并发、完整采集和 Sleep 后发布状态。main/closed 最高 attention 为 MCP，A=`0.9956816876459585`、M=`-0.0030228186348930475`、C=`0.5`，其 21 条真实事件独立计数；preSleep 最高排名为 Astrocyte。1,316 个原始公开 subject 在安全 DTO 中泄漏计数为零。生产 Git `tokens=0` 保留原字段，不表示 model usage；验收进程的模型/native/judge 调用全部为零。

### 实际命令与验证结果

```powershell
./tests/release-cognition/verify-wal.ps1 -Repository .verification/cross-day-20261015-0228/public-source.git -PublicRef b789466ff62f3c9b0e08ec392729f5bbfc5cfaa1 -ActualScope .verification/cross-day-20261015-0228/public-source.git -SleepDay 2026-10-14 -Timezone Asia/Shanghai -Strength 1 -Decay 0.1 -InitialConfidence 0.5 -AuthorizedPublicReplay -CrossDay -OutputDirectory .verification/cross-day-20261015-0228
./.verification/cross-day-20261015-0228/inspect-final.ps1
go test -mod=readonly -count=1 -v ./cmd/workbench-snapshot/...
go vet ./cmd/workbench-snapshot/...
```

跨日脚本在 PowerShell 7 实际运行 EXIT0，server/collector/snapshot/Sleep 四项构建、原生产采集与闭日全部成功；运行日志 `run-first.log`。最终 inspector 重新构建后，对三个对应不可变 SQLite 副本各自只读验证，`inspect-final-first.log` 记录 preSleep/closed/main 三项 PASS；没有重采事件或再次执行 Sleep。`go-test-source-final.log` 全部 PASS，含真实 server 初始化/Reader、副本保护、路径/锁/超时拒绝；`go-vet-final.log` 无错误。`gofmt -l` 与 `git diff --check` 通过。

Windows PowerShell 5 的首次直接文件解析发现 UTF-8 无 BOM 的中文注释解码问题，原 `powershell5-parser-first.log` 保留。最终脚本采用 UTF-8 BOM，Windows PowerShell 5 `Parser.ParseFile` 解析 PASS，记录 `powershell5-parser-final.log`；完整跨日运行使用 PowerShell 7，不将语法解析登记为 PowerShell 5 完整运行。

原五个 pub 指针只读核对为 5/5 与任务基线一致：Q `snapshot-892856014` / `snapshot-1030107153`，B empty/closed/native 分别 `snapshot-2498450736` / `snapshot-2249990550` / `snapshot-3888376845`。旧 DB、preSleep、历史和首 FAIL 原件保留；B native 的 H 固定公开任务 25 个 null 事件没有修改。准确三阶段来源/SQL/历史指针已通过 Orca Handoff 交协调者转 B/F/UI/P。

本轮完成公开工程认知的跨日工程验证；完整产品 HTTP/UI 组合、最终 I 同源 ZIP 与原生链路由其原 owner 继续验收。正式 V4、自动跨日日程、管理员 Procmon、模型/native/judge、GitHub Release 均未执行；没有锁定或修改任何日程，没有清理共享目录或他人文件。

## 本次公开来源验证入口

`tests/release-cognition/verify-wal.ps1` 只对新建 `.verification/public-cognition-*` 运行真实 server 初始化，显式关闭可选 summarize 与模型处理；其后启动独立公共 Git writer、在真实 WAL 上执行在线备份，再调用原 `cmd/sleep` 和生产 WorkbenchReader。`cmd/workbench-snapshot/verify` 为该验收入口，不进入可分发 snapshot binary。

公开来源限定主控确认由 Agent 创建的 [Astrocyte](https://github.com/songconmaisaix31-design/Astrocyte)，证据为 [S0 四轨记录](../tasks/S0-plan.md) 与其中精确 SOURCE。读取 Git 采用官方 [rev-list](https://git-scm.com/docs/git-rev-list)、[show](https://git-scm.com/docs/git-show) 命令，JSON 由 encoding/json 编解码，不读取凭据、私人会话或正式库。

2026-10-14 主控在 `msg_2cd6790ae313` 回复批准独立公开 Git 历史生产重放。调用既有 SeedTopic 人工入口时，以每个实际分类主题的最早真实 Git commit.ts 作为历史重放初始化时间，显式 strength=1、decay=0.1、initial-confidence=0.5；该调用由验收 Agent 执行，不代表当前用户人工建主题或人工评分。后续 MapGitLogRecord → EventRecorder → EventService/NormalizeSignal → Dynamics 保留每个 commit 的原始时间、origin_key 和实际 event_id。Sleep 与 snapshot 使用实际运行时刻，历史点只来自实际已持久化状态和已有发布点；不为分页额外创建 commit。该验证属于公开历史生产重放，独立实际 native/HTTP/DOM 链路与正式 V4 科研验收仍分别登记。

```powershell
git clone --bare https://github.com/songconmaisaix31-design/Astrocyte .verification/public-source.git
./tests/release-cognition/verify-wal.ps1 -Repository .verification/public-source.git -PublicRef 6f61d87ab043fd9801781fba09504120e32b3ab4 -ActualScope .verification/public-source.git -SleepDay 2026-10-14 -Timezone Asia/Shanghai -Strength 1 -Decay 0.1 -InitialConfidence 0.5 -AuthorizedPublicReplay
```

ActualScope 是本次实际 bare repository 的相对路径；主题仅由既有 mapper 的真实 subject 匹配产生。显式 PublicRef 必须来自已取得公开 Git，不消费其他本地历史。原始公开 Git 收据只保存在本次独立 SQLite，本地 safe 目录仅保存聚合、原生产 Reader 安全 DTO、六字段事件和覆盖四视图的十项 SQL 期望；发布包不能包含这些数据库或验收原件。

## REPORT：2026-10-14 实际结果

分支 `songconmaisaix31-design/release-cognition-20261014`，基线 `6f61d87ab043fd9801781fba09504120e32b3ab4`。最终 SOURCE 为 `1d07c5b2ff78db5ffe7e90b4d903c618b5a8ac50`，普通提交继承 `8001a76646f5892bc18d8ccb9e5cd065b51425a6` → `88bab5fe524e55d40b3799de0ff5fc6d45d08f6b` → `72b3ad224276e205c823585971b8aa4ff9e79738`。SOURCE 已 push，`git ls-remote origin refs/heads/songconmaisaix31-design/release-cognition-20261014` 实际返回相同完整 SHA；本 REPORT 独立追加提交，不修改业务代码。

本次 Q 使用 Codex/Orca；验收进程只运行现有 Go server、Git、SQLite、原 Sleep 与 Reader，没有向模型发起验收请求。所有代码变更只在 Q 独占路径，没有修改 engine、mapper、schema、迁移、权重、正式脚本或 B/R/H/F/P 的文件；依赖没有变化。

实际生产来源为已克隆公开 Astrocyte 中精确 `6f61d87` 可达的 1,339 个唯一真实提交。每个记录保留 `git:<实际commit>` origin_key、原 commit.ts、独立真实 event_id，所有 1,339 个 signal 均可关联真实 Git 收据。两个实际分类主题为 Astrocyte/Attention Engine（134 条）与 MCP（20 条）；1,185 条主题未分类，保持原 mapper 结果。manual 来源只表示现有 Git mapper 的存储类别，不代表人类身份。

新库由真实 server 初始化；没有人工 INSERT 或假来源库。历史初始化按批准调用既有 SeedTopic 入口，来源时间分别为 Astrocyte/Attention Engine `1791547734000`、MCP `1791727883000` 毫秒。实际 writer PID 43216，在线备份时已提交 WAL 为 2,027,072 字节，真实采集进度从 32 条增至 34 条。备份与 writer 是两个独立进程；后续完成全部 1,339 条，再运行原 `cmd/sleep`，显式 day=2026-10-14、timezone=Asia/Shanghai、decay=0.1、initial-confidence=0.5、fact-projection=none，实际持久化一条 Sleep run。

| 实际阶段 | generation | snapshot_at（UTC） |
|---|---|---|
| server 初始化空库 | snapshot-1037689488 | 2026-10-14T13:03:19.3878806Z |
| writer 运行中的 WAL 副本 | snapshot-2590329101 | 2026-10-14T13:05:01.4003976Z |
| 完整采集、Sleep 前 | snapshot-976935757 | 2026-10-14T13:05:44.9642893Z |
| 最新实际 Sleep 后 | snapshot-892856014 | 2026-10-14T13:05:45.4183032Z |

最新实际状态的 MCP 为最高 attention，A=`0.8298035126507481`、M=`-0.1191375411013564`、C=`0.5`；Astrocyte/Attention Engine 的 A=`0.7509206988621383`、M=`-0.1743555107965032`、C=`0.5`。最高排名与全部节点标量独立按 SQL 核对。两个主题各有三个实际已发布历史点；没有插值，没有用 snapshot_at 复制节点 attention。Workbench Today 查询前一天 2026-10-13 的 Sleep 保持 `not_produced`；本次实际运行日为 2026-10-14，没有补前日结果。缺少确证前日状态的 delta 保持 null。

### 命令与结果

- `go test -mod=readonly -count=1 -v ./cmd/workbench-snapshot/...`：真实 server 初始化、生产 Reader 七视图、缺失输入/重叠/hardlink/junction/超时/锁失败不发布、源主 DB 字节不变、旧 immutable 连接、真实历史导入、CLI 输出通过。日志 `.verification/final-tests.log`。
- `go vet ./cmd/workbench-snapshot/...`：通过。
- `go build -trimpath -mod=readonly -o .verification/workbench-snapshot-final.exe ./cmd/workbench-snapshot`：通过；`rg -a -l 'formal-20261013-first|astrocyte-validation|topic-timeseries-safe' .verification/workbench-snapshot-final.exe` 返回 1，没有对应私人快照常量。
- 上文 `verify-wal.ps1` 命令：实际 server 初始化、独立公共 writer、WAL 备份、四次发布、原 Sleep 均实际运行；首轮末尾样本前置失败保留。最终 SOURCE 的 inspector 在同一最新实际库与 generation 上复验通过，未增加来源事件或改变原始数据。
- `.verification/actual-public-cognition.exe --mode inspect --source .verification/public-cognition-6f63200932f14a10ad4d9fb91c78d749/app-state/state.sqlite --root .verification/public-cognition-6f63200932f14a10ad4d9fb91c78d749/local/astrocyte-workbench/v0.1 --output .verification/cognition-latest-safe`：生产 Reader 七视图、1,339 个 event_id/Git origin/source ts、全部六字段逐条 SQL、每页长度/总数/末尾空页、全部节点排序/标量通过；实际 Astrocyte 主题 134 条按 20/50 页大小完整分页通过。1,339 个原始 public subject 字符串在安全 DTO 中零泄漏，输出四路由十项独立 SQL 期望。

### 明确交付给 F/P 的本机指针

所有以下相对路径均位于当前 Q 工作树，不进入 Git 或 ZIP。

- 验收目录：`.verification/public-cognition-6f63200932f14a10ad4d9fb91c78d749/`。
- P 可只读选择的实际来源：上述目录的 `app-state/state.sqlite`。
- F `ASTROCYTE_TEST_COGNITION_LOCALAPPDATA`：上述目录的 `local/`；其 `astrocyte-workbench/v0.1/current.json` 指向最新 `snapshot-892856014`。
- P 可显式消费的真实历史：上述目录 `local/astrocyte-workbench/v0.1/snapshots/snapshot-892856014/history.json`。
- F `ASTROCYTE_TEST_COGNITION_SQL_EXPECTED_FILE`：`.verification/cognition-latest-safe/sql-expected.json`。
- 同 safe 目录提供 `actual-counts.json`、`events-six-fields.json`、`production-reader-safe.json`、`admission-safe.json`、`privacy-safe.json`、`origin-safe.json`；均为安全派生值。
- 原完整 Sleep 前 generation 与 `.verification/cognition-before-sleep-local/` 只用于已发布旧代读取与阶段 UI 核查，最终组合使用最新 Sleep 后 generation。

### 首次失败与当前限制

首次空库启动测试缺少显式关闭可选 summarize，产生 readiness RED，实际错误为 `project summarize 0.25.1 is missing or mismatched`。首次执行回执与失败目录 `.verification/snapshot-tests/actual-561268267/` 保留；该首次测试使用进程内日志，没有独立 server.log。后续真实启动显式设置 `ASTROCYTE_ENABLE_SUMMARIZE=false`，无需开发 node_modules。

Windows junction 实测曾被 Go EvalSymlinks 检查接受，产生目录别名/旧指针保护 RED；失败目录保留。最终 SOURCE 检查每个目的路径祖先的 Windows reparse 属性，实际 junction 拒绝与旧指针保护复验通过。普通 symlink 创建缺少 OS 特权，实际使用无需该特权的 junction 验证别名拒绝。

首轮最新 Sleep 状态的样本前置错误地绑定最高 A 主题超过 20 条，MCP 恰有 20 条，脚本实际退出 1；原 `.verification/cognition-after-sleep-safe/admission-safe.json` 与首轮 safe 诊断保持。主控在 `msg_ecbf07cfedce` 明确批准按原 Workbench 规格独立选择实际超过 20 条的主题进行分页，并单独验证最高 A 真实排名。最终 SOURCE 选实际 Astrocyte 134 条完整分页，MCP 的最高排名和 20 条计数完整保留，没有删除成功断言、缩减来源、改 topic/scope/ts/权重/数值或为样本制造提交。初次辅助 inspector 的相对 root 不适合旧 Reader，失败输出保留，最终 inspector 显式解析绝对根。

公开远端元数据查询与数次 Git 网络操作曾发生 EOF/SSL_ERROR_SYSCALL，失败保留。正常 clone、SOURCE push 与远端 SHA 核对均已实际完成。后续公开 refs 刷新实见 1,451 个可达提交，原基线外 112 条没有新增 MCP subject；没有把新提交填入本次 1,339 来源 cohort。

这是独立公开历史生产重放和工程验证，不能表述为当前用户真实人工评分、正式 V4 完成、科研收益或实际 native/HTTP/DOM 全链路验收。H 本次来源仍为 Pending observation、Event=nil、signals/nodes 为零，没有写入 H 库或借此补信号。完整前端默认浏览器、真实原生 controller 链路、最终集成与 ZIP 解压验收由 F/R/H/I/P 原 owner 继续；Q 未替它们填为通过。

目的根要求用户独占写入，路径检查和 snapshot.lock 不限制其他账户在检查后替换文件。管理员 Procmon、两次真实跨日快照和正式 V4 没有在本任务执行。Linux/其他平台运行与 fsync 崩溃恢复不在 Windows x64 实测范围。原 server/Reader 的私人保护常量整改已 Handoff B/M-release，Q 没有修改该旧文件；可分发 snapshot CLI 不引用它。
