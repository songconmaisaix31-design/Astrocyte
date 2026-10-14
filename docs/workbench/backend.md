# Workbench v0.1 后端交付与本机运行

## 精确来源

- 分支：`workbench/v0.1-backend`，已设置 `origin/workbench/v0.1-backend` upstream。
- 修订基线：`a88fd6e`，来自 `attention-engine/v1-20261011`；业务源码与 CI 基线 `100f00cb` 一致。
- OpenAPI SOURCE：`1ef8f751a56a6f94dd26ba0474da25b65c5e68ac`。
- 实现 SOURCE：`8635895e40eb022e92a86870861b4d09fd168147`。
- 本文件为独立 REPORT 提交；不把 REPORT 的文档变更计作业务源码。
- 客户端：本机 Codex / GPT-6，沿当前 Dispatch 执行，未新增 Worker。

## 实现范围

复用现有 Go server、session/loopback/origin、统一 ErrorV1。新增 `/api/v1/cognition/health`、`today`、`focus`、`topics/evidence`、`agents`、`events`、`filters` 的 GET。其他动词不进入 Reader；HEAD 返回固定 405。应用契约在 `internal/workbench/app`，SQL 在 `internal/adapters/sqlite/workbench.go`，HTTP 不导入 SQL 或 SQLite。

事件只有 `event_id/kind/topic/tokens/scope/ts` 六字段。tokens 缺失返回 null；source 只作绑定过滤参数和聚合组名。Reader 直接选择标量字段，不先反序列化完整 data/raw/observation。查询使用参数绑定，上海自然日包含 start/end 两日，排序为 ts 倒序、id 正序，page_size 最大 100。Evidence 总数和每日 kind 聚合覆盖完整证据，逐页可回溯所有事件。底层 SQL、载荷、路径错误收敛为固定不可用错误，无副本不回退正式源或 fixture。

每个请求读取本机 `current.json`，固定一个已完成的 generation。该代次下 DB/history 同时有效；后续请求在原子更换指针后重开新代次。SQLite 使用 `mode=ro`、`immutable=1`、`query_only(1)`，不运行迁移、不创建副本 WAL。旧代次保留，避免正在执行的 Reader 失去文件。

正式源固定为 `C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w1/formal-20261013-first/state.sqlite`，只有 snapshot.py 用只读连接执行 SQLite online backup；源不可由配置替换。备份固定读事务，模拟并发 WAL writer 可继续提交；只在新副本执行 quick_check 和 journal_mode=DELETE。失败不更新 current，人工和计划刷新通过本机 OS 文件锁互斥。

server 原应用数据库使用独立 app-state。进入现有迁移前拒绝正式源目录、snapshots 目录及其子目录作为 dataDir，并检查解析后的 DB 路径与源/副本文件身份。测试覆盖 snapshot hardlink。普通 server 的迁移仍只针对自己的独立库；Workbench Reader 从未调用迁移构造器。

Focus 读取固定 V1R safe 历史：`v1r-w1/replay-002/topic-timeseries-safe.json`，不是不存在的 docs/demo-data。每次成功快照将节点 `updated_at` 与 `attention` 的确证点追加到本机 history，按 topic+ts 去重。2192 个 V1R 点加 6 个节点点形成 2198 点；6 个节点值与本代次 SQL 一致。没有以备份时刻重标 A、插值或动力学复算；前日确证点或今日状态不足时 attention_delta 为 null。

引擎、events、Sleep、权重、公式、迁移、V4/validation、依赖和既有前端/spec/config 零改。旧 WIP 分支历史未合入本目标。

## 本机命令

仓库根目录运行（默认构建既有 web；`-WebDir` 可指定已构建的集成前端）：

```powershell
powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/workbench/start.ps1 -Port 17321
```

该命令构建现有 server/web、发布快照、Hidden 启动、验证 session/health，并注册 Windows 内置任务。默认 8787；端口占用会失败，不停止其他服务。已有同路径 Workbench 进程会拒绝重建；停止前必须核对 server.pid 与 ExecutablePath。`-SkipSnapshot` 只用于已有合法副本的部署，避免改变正在联调的代次。

所有 bin/config/PID/log/snapshot/派生 history 在 `%LOCALAPPDATA%/astrocyte-workbench/v0.1`，不进 Git/MCP。默认 app-state 位于该目录。`ASTROCYTE_WORKBENCH=1` 且未指定 `ASTROCYTE_DATA_DIR` 使用 app-state；原 Playwright ephemeral server 自己的临时 dataDir 仍可用，只要与源和副本独立。snapshot root 来自继承的 LOCALAPPDATA，无需额外挂正式源路径。

```powershell
# 已部署脚本的手动刷新与 API 审计
powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "$env:LOCALAPPDATA/astrocyte-workbench/v0.1/scripts/refresh.ps1"
python scripts/workbench/audit.py
# 计划任务复核，不把手动触发当作每日自动验收
Get-ScheduledTask -TaskName 'Astrocyte-Workbench-v01-*'
Get-ScheduledTaskInfo -TaskName 'Astrocyte-Workbench-v01-Snapshot'
```

两个任务均使用当前用户 Interactive、Limited：Snapshot 每日上海 00:30，Start 用户登录启动，StartWhenAvailable、IgnoreNew、5 分钟上限。机器时区必须 China Standard Time，用户需处于登录会话。快照任务最近执行 `2026-10-14 01:23:58 +08:00`、Result=0，下一次 `2026-10-15 00:30 +08:00`；登录任务最近 `01:16:58`、Result=0。01:24 快照不是准点 00:30 验收，可能为 StartWhenAvailable 补跑；Operational 日志未启用，触发起因未确证。两次真实跨日准点自动成功保持 **NOT_RUN**。

## 验证结果（SOURCE 8635895e）

| 命令/现场 | 结果 |
| --- | --- |
| `go test -mod=readonly -count=1 ./internal/workbench/app ./internal/adapters/sqlite ./internal/adapters/httpapi ./cmd/server` | PASS，SQL/HTTP/server 三包；app 为类型契约，无测试文件 |
| `go vet -mod=readonly ./cmd/server ./internal/workbench/app ./internal/adapters/httpapi ./internal/adapters/sqlite` | PASS |
| `python scripts/workbench/test_snapshot.py` | 2 PASS：模拟源并发 WAL writer、失败发布保留指针、确证历史保留；正式源未写 |
| `python scripts/workbench/test_http.py --server <本机构建 exe>` | PASS：隔离模拟副本、实际 server 13 请求、raw_unique 1、泄漏 0；正常/SQL失败/缺副本/鉴权均覆盖 |
| `python scripts/workbench/audit.py` | PASS：真实副本 50 请求，3105 唯一事件完整分页，3220 六字段投影；10 数值与独立 SQL 一致 |
| 真实 raw_unique | **NOT_APPLICABLE**，实际数据中 0 个哨兵；不把空遍历称 canary PASS |
| 真实 raw 字符串审计 | PASS：7248 个非派生字符串（长度至少 8），50 请求的 JSON 字符串值中命中 0；原始字符串只留进程内存 |
| `node node_modules/@redocly/cli/bin/cli.js lint contracts/openapi.yaml --config contracts/redocly.yaml` | PASS，1 个既有 EventV1 unused warning |
| `node scripts/validate-examples.mjs` | PASS，277 例 |
| `node scripts/check-architecture.mjs` | PASS，28 包，HTTP 不依赖持久化 |
| `node scripts/build.mjs` | PASS，server/MCP/Sleep 构建、原 web TS/Vite 构建 |
| `start.ps1 -Port 17321 -SkipSnapshot` | PASS，最终 SOURCE 本机构建/Hidden 部署/session health/任务 Ready |
| PowerShell parser / Python compile / `git diff --check` | PASS |

本次部署后 17321 owned PID=`15652`，ExecutablePath=`%LOCALAPPDATA%/astrocyte-workbench/v0.1/bin/astrocyte-workbench.exe`。此 PID 是核实时的现场值，后续使用前重核。current=`snapshot-hxfipcpr`，snapshot_at=`2026-10-13T17:24:01.253839+00:00`；3105 事件、31 Sleep、事件截至上海 10-12，source 总数 claude_code=93、codex=846、manual=2166。manual 包含 git 来源，不推断为纯人类；上海 10-14 今日活动真实为空，不造今日事件。

派生证据在本机 logs：`source-8635895-go-tests.log`、`source-8635895-backup-tests.log`、`source-8635895-http-canary.log`、`source-8635895-api-audit.log`、`api-audit.json`、`source-8635895-tasks-final.json`、`source-8635895-data-summary.json`。api-audit.json 含 10 项独立 SQL 与期望值，可供 frontend/root 核 UI；浏览器四视图验收属于 frontend/root，不从后端 HTTP 结果推定 UI PASS。

主控通过当前 Dispatch 回传：frontend SOURCE `8ec40f7` 已普通消费后端 `8635895e`，原 Playwright 双 viewport 四视图真实联调及 10 个 UI 数值 PASS，root 独立 10/10×2 复核 PASS。这是 frontend/root 的现场结果，本轨未修改或自行重跑前端代码。最终组合完整 check/build/defaultE2E/CI 仍由集成/root 执行。

## 首次失败与验收边界

- 最初 Windows SQLite URI 用 `file://C:/...`，adapter 报 invalid uri authority；首 RED 在 `backend-first-tests.log`。现为正确 `/C:/...` URL Path，回归已 PASS，原日志保留。
- 旧全量 Go 首轮 importer `TestPaperDOMInstalledPinAndIsolation` 30 秒 deadline，原 `go-all-first.log` 保留；历史 `go test -p 1` 重跑 PASS。该历史结果不能冒称本次全量默认 check 通过。
- 旧 `pnpm check` 在 importer/version probe/Windows 进程资源错误后中断，`pnpm-check-first.log` 保持 **PARTIAL/FAIL**，未完成后续阶段。本次只声称适用定向检查和完整 build 通过；最终组合 check 与新 cognition Playwright 由 frontend/集成/root 完成。契约生成器实际输出 `web/src/api/schema.d.ts`，由 frontend owner 消费精确契约生成，backend 不跨写。
- 字符串审计开发首轮把毫秒格式不同的派生 ts、合法 scope 子串和 schema key 当命中，诊断 RED 留在原会话与本机回顾摘要；修正的是审计对派生值和 JSON 字符串值的识别，未裁剪业务 scope 或改数据投影。
- Microsoft portable Procmon64 签名 Valid，本机 token 非管理员；实际文件系统监控 **NOT_RUN**。静态边界、query_only 测试、只读 backup 调用和 API PASS 均不能替代正式 DB/WAL/SHM 写入监控。
- 两日真实自动执行 **NOT_RUN**。不修改 V4 日程、协议、正式库，未合 main；v0.2 保持冻结。

## 管理员文件监控待执行步骤

仅在用户已经打开的管理员终端执行；脚本不会弹 UAC、全局安装、升级权限或停止无关 Procmon。当前工具：`%LOCALAPPDATA%/astrocyte-workbench/v0.1/tools/ProcessMonitor/Procmon64.exe`。

1. 在官方 Procmon Filter 中仅保存下列三个 Path **is / Include** 规则，删除无关规则，保存为 `%LOCALAPPDATA%/astrocyte-workbench/v0.1/logs/filesystem/formal-source.pmc`。记录阶段不提前限定 PID，以免漏掉稍后启动的 backup 子进程。

```powershell
New-Item -ItemType Directory -Force "$env:LOCALAPPDATA/astrocyte-workbench/v0.1/logs/filesystem" | Out-Null
```

```text
C:\Users\DW\AppData\Local\astrocyte-validation\v1-20261012\v4-w1\formal-20261013-first\state.sqlite
C:\Users\DW\AppData\Local\astrocyte-validation\v1-20261012\v4-w1\formal-20261013-first\state.sqlite-wal
C:\Users\DW\AppData\Local\astrocyte-validation\v1-20261012\v4-w1\formal-20261013-first\state.sqlite-shm
```

2. 执行下列命令。monitored-refresh 记录实际 server PID 和新 Python backup PID 到 owned-pids.txt，并在 capture 期间执行快照与真实 API 审计。snapshot 与 server 的 stdout/stderr 只留本机 logs。

```powershell
$wbRuntime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "$wbRuntime/scripts/monitor.ps1"
powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "$wbRuntime/scripts/monitored-refresh.ps1"
powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "$wbRuntime/scripts/monitor.ps1" -StopAndAudit
```

3. StopAndAudit 验证自身 Procmon PID/ExecutablePath 后才终止 capture，保存 PML/CSV。审计按三文件与两具体 PID 过滤，检查 WriteFile、截断、分配、删除、重命名、元数据设置，以及 CreateFile/mapping 中可写访问。零事件只能 REVIEW_REQUIRED；有观察且零写操作/零可写 opens/maps 才报 ZERO_WRITE_OBSERVED，仍需人工复核过滤器、捕获时间和 PML。SQLite 只读连接对 WAL/SHM 的实际行为尚未经过本机管理员监控，不预先宣称零写。

原始 PML/CSV 和任何私有载荷不上传、不进 Git/MCP；仅发布派生计数和明确验收状态。两日自动验收需保存真实 00:30 触发记录、任务结果及对应 snapshot_at；补跑和手动刷新独立标记。
