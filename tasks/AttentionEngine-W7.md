# Attention Engine W7 · 本地作品聚合（AI 作品寻回册）+ 链1事件源

分支 `songconmaisaix31-design/attention-engine-w7-works`，基线 `4d200df`（未 merge `origin/attention-engine/v1-20261011`，M0 契约未就绪，见 §5）。写域仅本任务所列，不改既有文件；端点/迁移/HTTP/主程序接线按 `AttentionEngine-plan.md` 交 W0 统一落。

## 1. 交付内容（五能力 + 事件源）

| 能力 | 落点 |
|---|---|
| ① 平台聚合与统计 | `domain/works.go` `AggregateWorksStats`（作品/文件夹/接力过/最近在动计数），前端 `WorksAlbumPanel` 平台统计区 |
| ② 作品卡片与文件核实 | `adapters/worksscan` 有界扫描读 `mtime`/`file_count` 置 `verified`（"从文件核实"徽标）；JSON 投递不置 `verified` |
| ③ 交接手册 | `WorkHandoff{content,updated_at,source}`，`SetWorkHandoff` 记录更新时间，`source=human|generated` |
| ④ 态度与视图 | `Attitude{continue|park|done}` + 平铺/按平台/按分组/时间线 + 隐藏 + 排序（`FilterWorks`/`SortWorks`） |
| ⑤ 扫描导入与同步 | `worksscan.Scanner.ScanWorks`（有界）、`worksscan.Delivery.ParseDelivery`（AI 投递 JSON）、`RefreshWorks` 同步最近改动 |

事件源：`app.WorksService` 经 `WorksEventSink` 发 `domain.WorkEvent{work_id,kind,source,scope,topic,at,payload}`；kind 词汇 `work_imported|work_attitude_set|work_handoff_set|work_hidden|work_synced|work_verified`。SQLite `works_album` 持久化可替换观察、`works_metadata` 持久化人工元数据（态度/分组/隐藏/手册，revision CAS），冷重启一致。

新增文件：

- `internal/workspace/domain/works.go` `works_test.go`
- `internal/workspace/app/works.go` `works_test.go`（端口 `WorksRepository/WorksScanSource/WorksDeliverySource/WorksEventSink` 定义在 works.go，不改 contracts.go）
- `internal/adapters/worksscan/{scanner,json}.go` + 测试
- `internal/adapters/sqlite/works_store.go` `works_store_test.go`
- `web/src/pages/workspace/{WorksAlbumPanel.tsx,WorksAlbum.module.css,worksAlbumPresentation.ts,worksAlbumPresentation.test.ts}` + `WorkspacePage.tsx` 最小接线（W7 单一 owner）
- `web/e2e/works-album.spec.ts`、`tests/attention-engine/w7_works_test.go`

## 2. 平台勘察（本机实际，有界）

PATH 实测安装：codex、claude、opencode、pi、grok、kimi、qwen、cursor。配置目录存在（仅目录名，未读内容/凭据）：`.claude` `.codex` `.cursor` `.trae` `.trae-cn` `.openclaude` `.workbuddy` `.workbuddy-ai` `WorkBuddy` `WorkBuddy AI`。任务书列名的 灵犀/DeepSeek Harness 未在 PATH 命中。平台清单为开放词汇（`ValidWorkPlatform` ≤64），按扫描标记或 JSON 投递实际取值，不硬编码封闭清单。

## 3. 端点契约（W0 Handoff，落 `contracts/openapi.yaml` + `internal/adapters/httpapi/` + `cmd/server/main.go`）

人类会话 + CSRF，写操作带 `Idempotency-Key`。DTO 字段与 `domain.Work`/`WorksSnapshot` 对齐。

| 方法与路径 | 用途 |
|---|---|
| `GET /works` | 读缓存相册，返回 `WorksSnapshotResultV1{snapshot:{works,platforms,status,observed_at,failures}}` |
| `POST /works/refresh` | 有界扫描（body `WorksRefreshRequestV1{roots[]}`），返回同上 |
| `POST /works/delivery` | AI 投递 JSON（body `WorksDeliveryRequestV1{content:string}`，≤1MiB），返回同上 |
| `PUT /works/{work_id}/attitude` | body `WorkAttitudeRequestV1{expected_revision,attitude}` |
| `PUT /works/{work_id}/group` | body `WorkGroupRequestV1{expected_revision,group}` |
| `PUT /works/{work_id}/handoff` | body `WorkHandoffRequestV1{expected_revision,content}`（≤64KiB，utf8） |
| `PUT /works/{work_id}/hidden` | body `WorkHiddenRequestV1{expected_revision,hidden}` |

迁移 `014_works.sql`（W0 owner）：

```sql
CREATE TABLE works_album (
    key  TEXT PRIMARY KEY,
    data TEXT NOT NULL
);
CREATE TABLE works_metadata (
    work_id  TEXT PRIMARY KEY,
    data     TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision >= 1)
);
UPDATE foundation_schema SET version = 7 WHERE context = 'workspace';
```

主程序接线（`cmd/server/main.go`，W0 owner）：

```go
worksService := workspaceapp.NewWorksService(db, worksscan.NewScanner(), worksscan.Delivery{}, worksEventSink, workspaceapp.WorksServiceOptions{})
services.Works = worksService
```

`worksEventSink` 指向链1事件落库（见 §4）；`RegisteredProjects` 已可作为 `RefreshWorks` 默认扫描根来源（W0 把 `localProjects` 的登记根传入 `WorksRefreshRequestV1.roots`，或接一个 `WorksRoots` 提供者）。HTTP 层只解析/校验 DTO，鉴权沿 `HumanSession` + CSRF。

## 4. 链1事件源映射（W1 Handoff，`WorkEvent` → §4.1 `Event`）

| WorkEvent.kind | §4.1 kind | source |
|---|---|---|
| `work_imported` | `collect` | manual |
| `work_attitude_set`(continue/done) | `user_approve` | manual |
| `work_attitude_set`(park) | `bookmark` | manual |
| `work_hidden` | `user_veto` | manual |
| `work_handoff_set` | `agent_task`(generated) / `manual`(human) | 按 `handoff.source` |
| `work_synced` | 系统刷新，机器信号（不增人类关注） | 建议丢弃或零权 |
| `work_verified` | 文件核实（机器） | 建议零权 |

`WorkEvent.scope`=作品根，`topic`=platform 种子（睡眠聚类细化），`payload`=工作 JSON（审计）。`tokens` 无（作品活动非 token 计费），W1 归一化时置 0/缺省。事件落库（`event_store`）由 W1 接 `WorksEventSink` 实现；W7 只定义端口与载荷。

## 5. 依赖与剩余限制

- **M0 Event 契约未就绪**：`origin/attention-engine/v1-20261011` 仍为派发记录 `94c6b5d`，各 worker 分支停在 `4d200df`。主控通知后 `git fetch origin && git merge origin/attention-engine/v1-20261011` 再接入 Event 类型；当前事件经 `WorksEventSink` 端口 + §4 映射表，不影响五能力本身。
- 端点/迁移/HTTP/主程序接线全部 W0 owner，W7 未改 `contracts/openapi.yaml`、`internal/adapters/httpapi/`、`migrations/`、`cmd/`。前端 `WorksAlbumPanel` 真机数据待端点接入后切换（当前 fixture 演示 + 真机空态）。
- 有界扫描仅 `worksscan` 给定根；不整盘扫描、不读凭据/私密正文；JSON 投递不读路径。默认扫描根（登记项目根）的接线属 W0（`RegisteredProjects` 复用）。

## 6. 验证

| 命令 | 结果 |
|---|---|
| `go test ./internal/workspace/... ./internal/adapters/worksscan/ ./internal/adapters/sqlite/ -run 'Works|W7'` | PASS |
| `go test ./tests/attention-engine/ -run TestW7` | PASS（真实临时目录扫描 + 冷重启一致 + 五类事件） |
| `go build ./...` | PASS |
| `pnpm --dir web typecheck` | PASS（exit0） |
| `pnpm --dir web exec vitest run src/pages/workspace/worksAlbumPresentation.test.ts` | 7 PASS |
