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

实际验证记录将在独立 REPORT 提交追加；公开来源认知验收、并发 WAL 采集及前端验收尚未完成。正式 V4、旧私人历史、管理员 Procmon、两次真实跨日触发没有在本任务运行。

## 本次公开来源验证入口

`tests/release-cognition/verify-wal.ps1` 只对新建 `.verification/public-cognition-*` 运行真实 server 初始化，显式关闭可选 summarize 与模型处理；其后启动独立公共 Git writer、在真实 WAL 上执行在线备份，再调用原 `cmd/sleep` 和生产 WorkbenchReader。`cmd/workbench-snapshot/verify` 为该验收入口，不进入可分发 snapshot binary。

公开来源限定主控确认由 Agent 创建的 [Astrocyte](https://github.com/songconmaisaix31-design/Astrocyte)，证据为 [S0 四轨记录](../tasks/S0-plan.md) 与其中精确 SOURCE。读取 Git 采用官方 [rev-list](https://git-scm.com/docs/git-rev-list)、[show](https://git-scm.com/docs/git-show) 命令，JSON 由 encoding/json 编解码，不读取凭据、私人会话或正式库。

2026-10-14 主控在 `msg_2cd6790ae313` 回复批准独立公开 Git 历史生产重放。调用既有 SeedTopic 人工入口时，以每个实际分类主题的最早真实 Git commit.ts 作为历史重放初始化时间，显式 strength=1、decay=0.1、initial-confidence=0.5；该调用由验收 Agent 执行，不代表当前用户人工建主题或人工评分。后续 MapGitLogRecord → EventRecorder → EventService/NormalizeSignal → Dynamics 保留每个 commit 的原始时间、origin_key 和实际 event_id。Sleep 与 snapshot 使用实际运行时刻，历史点只来自实际已持久化状态和已有发布点；不为分页额外创建 commit。该验证属于公开历史生产重放，独立实际 native/HTTP/DOM 链路与正式 V4 科研验收仍分别登记。

```powershell
git clone --bare https://github.com/songconmaisaix31-design/Astrocyte .verification/public-source.git
./tests/release-cognition/verify-wal.ps1 -Repository .verification/public-source.git -PublicRef 6f61d87ab043fd9801781fba09504120e32b3ab4 -ActualScope .verification/public-source.git -SleepDay 2026-10-14 -Timezone Asia/Shanghai -Strength 1 -Decay 0.1 -InitialConfidence 0.5 -AuthorizedPublicReplay
```

ActualScope 是本次实际 bare repository 的相对路径；主题仅由既有 mapper 的真实 subject 匹配产生。显式 PublicRef 必须来自已取得公开 Git，不消费其他本地历史。原始公开 Git 收据只保存在本次独立 SQLite，本地 safe 目录仅保存聚合、原生产 Reader 安全 DTO、六字段事件和覆盖四视图的十项 SQL 期望；发布包不能包含这些数据库或验收原件。
