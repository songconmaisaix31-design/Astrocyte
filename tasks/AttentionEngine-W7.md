# Attention Engine W7 · 本地作品聚合（AI 作品寻回册）+ 链1事件源

> 文档用途：需求与历史记录。保留本文件独有业务条款、来源、实际检查和未验限制；原日期的人员分工、资源额度与操作安排只适用于当时阶段。后续执行读取 [AGENTS](../AGENTS.md)、[RESOURCES](../docs/RESOURCES.md) 和 [当前任务目录](../docs/TASKS.md)，独立 Reviewer 按准确来源验收；归档与入口调整不改变功能完成状态。

## Windows CI 定向返修（2026-10-11，本轮新 Dispatch）

任务 `task_46b94cb91958` / `ctx_fe6ac3090f92`，继续原 Codex --yolo Worker，不创建 Agent。唯一写域 `internal/adapters/worksscan/**` 与本报告；W0 同时独占 agents 返修。本轨先普通合入 root 最终树 `0e49301734549a02b9b1b9078aaef96fd6422742`，merge `82aa0d01a6c0039f42cbfe23de03a21ca7cd80e7`，没有 cherry-pick/squash/force。SOURCE `38c5bdeeb5fcec7b2f0a3097ef3730ab1a30d9c4` 已 push 至 `songconmaisaix31-design/attention-engine-w7-works`。

第一远程 RED 原件 `C:/Users/DW/AppData/Local/Temp/astrocyte-ci-38150022874-first-failed.log` 未改：Windows CI38150022874 的 os.TempDir 返回 `C:\Users\RUNNER~1\...`，EvalSymlinks 展开为 runneradmin，旧 EqualFold 比较把名称别名误判为 symlink；scanner 返回 root_unavailable，marker 测试下标 panic，下游 HTTP/Works 聚合失败。自有合成 TempDir 通过 Windows GetShortPathName API 实际获得 `TESTSC~1/APPROV~1`，修复前根+祖先短名和仅祖先短名两场均同样 RED，不伪造短名或改 TEMP 掩盖。

修复：Windows GetLongPathName 展开目录条目名称作比较；root/每级 ancestor 先 Lstat，拒绝 ModeSymlink 及 Windows reparse point，再独立 EvalSymlinks 校验。保留输入已批准 root 拼写/WorkIdentity，不把它替换为父目录或其他路径；不以 SameFile 单独准入。marker 测试读取 works[0] 前断言 works/failures，失败输出原原因。maxRoots=256（HTTP 既有 roots≤64）、maxDepth=8、maxFiles=10000、marker≤64KiB、私密/依赖目录排除及 HTTP 登记授权均未改变。

额外保留本地 junction 首 RED：当前 Go 将真实 junction 根暴露为 ModeIrregular，ModeSymlink/EvalSymlinks 组合会错误接受它；实际 junction 祖先先被拒绝。新增 Windows reparse 属性检查后根/祖先/短名祖先全部拒绝。测试中 SameFile 明确证实 junction 与 target 为同一对象，但生产代码不以此接受重定向。此发现已向 root Handoff 给 W0 agents owner，未改其文件。

| 本轮命令（GOFLAGS=-p=1，GOMAXPROCS=2） | 结果与范围 |
|---|---|
| `go test -count=1 -v ./internal/adapters/worksscan` | 9 个顶层 PASS，2 SKIP；真实 8.3 两子例、真实 junction 三子例均 PASS；保留 depth、node_modules 排除和 marker 测试 |
| `go test -count=1 ./internal/adapters/httpapi ./internal/workspace/... ./internal/adapters/sqlite ./tests/attention-engine -run 'Works\|W7WorksAggregationAcrossRestart'` | 五包 PASS；既有真实 SQLite/HTTP human+CSRF/登记根边界、请求重放、restart 与 W7 临时目录聚合；没有跑批准的实际作品根 opt-in |
| `GOOS=linux GOARCH=amd64 go test -c ./internal/adapters/worksscan -o C:/Users/DW/AppData/Local/Temp/astrocyte-w7-ci-scanner-linux.test` | PASS，仅跨平台编译，未执行 Linux 二进制 |
| `gofmt` 所改 Go 文件；`git diff --check` / staged check | PASS |

真实限制：本机无 os.Symlink 创建权限，root/ancestor 普通 symlink 两项 SKIP，不能称该两项实际运行通过；unprivileged 真实 junction 的 root/ancestor/短名重定向则已执行通过。真实 8.3 已执行，没有短名 NOT_RUN；其他文件系统不生成短名时测试明确 NOT_RUN/SKIP。远程 CI 首次失败仍 RED，新最终树 CI 与最终合入由 root 验证，本轨未启动/重跑远程 CI、全套检查、UI 写入、媒体、真实作品扫描或私人 native 记录。下文上一 Dispatch 的实际五能力/链1报告为历史有效证据，本轮仅解决上述 CI 路径兼容故障，不宣称整套 CI 已绿。

## 当前返修状态（2026-10-11）

**W7 本轨验收通过**：ea942c4 的旧交付仅领域、存储、本地服务和 fixture；返修已完成真实五能力 UI、实际登记根有界扫描、持久化 HTTP 命令、重启、三尺寸键盘以及链1原始 pending observation 验收。普通消费 W0 P2 `5d356370c4157e74d96388afd098d5fea14dc799` 和最终 startup `396aec1d6bee777dd380ff06fc5d821d0f7f5b3f`，依赖 owner 接线，未修改共享 HTTP/契约/迁移/generated/cmd。当前 Dispatch task_355d5b638062 / ctx_cf173bcd2418，实际客户端 Codex（当前已启动的 --yolo 会话），不新建 Agent。根节点独占协调；本 Worker 保留原写域。以下早期 NOT_RUN/NOT_DONE/RED 是当时记录，保留原始失败；最终证据见下文“最终链1阶段”。本结论不表示整个 Attention Engine 或用户主观视觉验收完成。

第一阶段纠正：未知缓存的数量显示未知；未核实作品的文件数与最近活动显示未知；未知扫描平台使用 unknown，不推断客户端；扫描最多 8 层/10000 文件，截断/读取失败明确未完整核实，不跟随 root symlink。WorkEvent topic 未知保持空，sync/verify 来源 machine，分组使用 work_group_set，导入/同步仅发本次涉及作品。

主控通过 msg ask 答复批准当前工作树 `C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-w7-works` 为 metadata-only 根。`ASTROCYTE_TEST_W7_APPROVED_ROOT` 定向 Go 验收实际读目录名/stat/mtime/count，538 文件、unknown 平台、verified=true、0 failures；这是实际扫描证据，**不等于真实 HTTP/浏览器/链1验收**。可信登记根解析、HTTP 鉴权/CSRF、schema/迁移/启动仍属 W0，EventSink 归 W1；未发布前不冻结传输契约或私自接线。

首类型检查 RED（展示 helper 替换误用了变量），修正后 typecheck PASS；首局部 ESLint RED（effect 同步 setState），改为派生搜索后局部 ESLint PASS。Go `-p=1` / GOMAXPROCS=2 定向 Works/W7/Scanner/Delivery PASS；Vitest 展示 7 PASS；`pnpm --dir web build` PASS；`git diff --check` PASS。Node 限制 512MiB。真实 API UI 命令、SQLite HTTP 后重启复核和三尺寸键盘仍 NOT_RUN。

分支 `songconmaisaix31-design/attention-engine-w7-works`，基线 `4d200df`（未 merge `origin/attention-engine/v1-20261011`，M0 契约未就绪，见 §5）。写域仅本任务所列，不改既有文件；端点/迁移/HTTP/主程序接线按 `AttentionEngine-plan.md` 交 W0 统一落。

阶段 SOURCE `bbb1268` 已 push；独立 fixture Playwright `e2e/works-album.spec.ts --workers=1` 使用隔离端口 15247/18847，14 PASS / 24.9s（两项目重复包含 1920/1280/390 与键盘），不计为真实 API 验收。共享依赖依主控 follow-up 仍 pending；不消费 partial b6a5794 作为完整 M0。未接线的 `WorksCommandForms.tsx` 仅准备有界登记根选择与人工元数据/JSON 表单，无传输契约猜测。新库缓存明确 status=unknown、observed_at=null。

后续扫描纠正：无标记平台的扫描更新同根已有投递作品文件事实，保留明确投递平台、名称、来源和人工元数据，避免重复未知卡；根读取失败使缓存文件事实未核实并显示原因。定向 Works app 验证 PASS。上述改动不改变已发布 Go 公共 API；ca0c255 可供 W0 接线编译，最终精确源另行报告。

已普通合入正式 M0 source `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0`，merge `b19bcd7` 已 push。合后定向首 RED：新共享 harness 用 attention_engine_test，而旧 W7 为 attentionengine_test；仅本轨 w7_works_test.go 改为共享 package 名，随后 `go test -count=1 ./tests/attention-engine -run TestW7` PASS。同轮 domain/app/scanner/sqlite 定向 PASS、typecheck PASS。HTTP Works P2 SHA 仍待 W0，不扩展全仓检查。

主控/W0 已确认新增请求去重在本轨范围，migration016 的 works_requests 字段已对齐。复用 Caller.OperationID（由 HTTP Idempotency-Key 给定）；独立 WorksRequestRepository 原子 reserve pending 后才做效果，相同已接受输入返回原结果，输入不同 version_conflict，pending/unknown 不重放。metadata canonical input 含 work id、expected_revision 和业务值；JSON 投递保留原 content；操作无 key 的旧内部调用只用于 local-only 服务测试，不代表 HTTP。

事件 sink 失败现在显式 delivery_unknown（本地作品/metadata 可能已保存），元数据播种读写错误不再静默继续。SQLite 定向 TestWorksRequestReceiptsPersistAndPreventUnknownReplay PASS：实际 reserve/result 持久化、重复请求只发一次事件、CAS 重复返回原 revision、sink 未知效果不重发、pending/accepted/unknown 真数据库关闭后再打开均不重放。五包 Works/W7/Scanner/Delivery `-count=1` PASS。当前仍为 contract_local DDL，不等于 W0 migration016 或真实 HTTP 验收。

## 真实界面阶段（2026-10-11 22:36–22:42）

`WorksAlbumPanel` 使用真实 GET/POST/PUT 与 generated Work/WorksSnapshot 类型；HTTP/CSRF/会话复用既有 api 客户端。每次人工提交单一 UUID 同时作为 request_id/Idempotency-Key，不重试命令；独立 expected_revision CAS；失败留草稿，结果未知时先读缓存。fixture 只展示且扫描/保存禁用。新 owns `worksAlbumApi.ts`/`WorksCommandForms.tsx`，无共享客户端或 S1 写域修改。

- 真实批准根：当前 W7 工作树；首次有界扫描和同步只发这个登记根，实际 557 文件、platform=unknown。JSON 提交为明确新线索，不读 unregistered-clue 路径，files.verified=false，数量/最近活动未知。
- `ASTROCYTE_TEST_W7_LIVE=1 ... playwright test --config e2e/works-live.config.ts`：1 PASS / 1.5m；8 真实命令（扫描、态度、分组、手册、隐藏、取消隐藏、JSON、同步），会话+CSRF+同 UUID；SQLite 全部 accepted，input/result typeof blob。人工 revision5、手册 human 与生成投递手册 generated 分开；越界根/缺 CSRF 均403且数据库不变；API 真停止/重启后相册原字节、metadata 和浏览器重读一致。
- 首场 viewport/键盘 DOM PASS，但页面截图误落相邻面板，不声称截图视觉通过。原件保存在 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-aijSTJ/first-browser-evidence`，真实数据库同目录 data/state.sqlite，所属 API/Vite 已关闭。
- 复用原 owned 临时库，仅 GET、0 Works 写请求的 `ASTROCYTE_TEST_W7_VIEWS=1 ... --grep 'saved actual works'`：1 PASS / 10.6s；1920/1280/390、Enter/Tab/Escape/焦点返回、无横向溢出/页面错误。局部作品/详情截图已逐张查看，原件保存在同库 `read-only-views-evidence`；截图整段截取包含浮动页头叠加，不冒充用户主观视觉验收。
- fixture/mock 分开：`playwright test e2e/works-album.spec.ts --workers=1` 16 PASS / 29.8s（原 fixture14 + unknown 写失败草稿/读缓存2），不计真实入库。
- `pnpm --dir web typecheck`、所改 TS/TSX/spec/config 局部 ESLint、`pnpm --dir web build` PASS，Node512；真实 Go receipt 在 canonical migration016 的定向 restart 测试 PASS。

W1/W4 合后 receipt harness 首 RED（重复 CREATE works_requests）已保留；本轨即时 `a3eab2c6d0df026e056a5a873a0d63bf58282688` 移除 CREATE、仅使用实际 W0 migration016，定向 PASS 并广播 owner。W1 原 sink-error RED 在消费 `b166848` 后两项定向 GREEN 已报告，保留旧源 RED。W1 pending-only sink source `b73bf6575621ea0c46fcf63df9bf01fe2726f900` 已交 W0，待 W0 发布启动 assembly 精确源后，复用真实库验证原始 observation 落库、topic-empty/pending/noSignals 与重启；当前 nil sink 的真实 UI 场不是链1已接线证据。

## 最终链1阶段（2026-10-11 22:47 起）

W0 最终启动 `396aec1d6bee777dd380ff06fc5d821d0f7f5b3f` 普通 merge 为 `07c580a95fe4a67bce05e0e34c5b294bba656939`：实际 startup 将 Works 交给 W1 `NewWorksEventSink(NewPendingEventService(sqlite.NewEventRepository(db)), nil)`。复用原成功 owned SQLite 库，不重跑首场八命令来凑事件数；仅新增一个人工分组与一个批准根机器同步，其余为 GET/已接受请求相同 key 重放。

- `ASTROCYTE_TEST_W7_SOURCE=1 ASTROCYTE_W7_REUSE_OWNED_TEMP=... playwright test --config e2e/works-live.config.ts --grep 'same owned works API'` PASS（`.last-run.json` passed，成功生成独立证据）。新增恰好两条 Works raw BLOB：人工 `manual/work_group_set` 和机器 `machine/work_synced`，均 `mapping_status=pending`、topic 空、Event data 为 null；不把 platform 当 topic，也不把 raw kind 当闭集 Event.Kind。
- 两次 accepted 人工 key 重放（重启前/后）均200且事件集合不变。API 真停止/重启后相册、人工 metadata、receipt input/result 与 raw observation 字节一致；Signal 数一直0、节点集合仍空，机器更新没有增加注意力。人工 metadata revision6；同步真实文件数657，与先前557是不同 checkout 时点的实际值，不冻结或补造计数。
- 原始证据保存在 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-aijSTJ/final-chain1-evidence/actual-works-chain1-evidence.json`；只含本次 Works 自有事件，不读取其他会话正文。此库和三轮独立证据均保留；所属服务关闭。
- 最终 source 合后类型首 RED：新版生成类型要求 expected_version 字面量1，旧 helper 推断 number；新增 raw BLOB 测试遇到既有 harness 声明不包含 Uint8Array。仅本轨修复 `1 as const` 和显式运行时 BLOB 检查；保留此 RED，后续 targeted GREEN 单列。Go 定向五包 Works/W7/Scanner/Delivery PASS，`go build ./...` PASS；Vitest 展示7 PASS。

精确命令环境统一 `GOFLAGS=-p=1 GOMAXPROCS=2 NODE_OPTIONS=--max-old-space-size=512`。仅新增 source/只读浏览器检查，首场 nil-sink 和 fixture/mock 证据保持分开；没有自动重试未知写，也没有改 shared owner 文件。

类型修正后 `pnpm --dir web typecheck`、所改 TS/TSX/spec/config 的局部 ESLint、`pnpm --dir web build` 全 PASS（106 modules）。同库最终 `ASTROCYTE_TEST_W7_VIEWS=1 ... --grep 'saved actual works'` 1 PASS / 12.2s：1920/1280/390、Enter/Tab/Escape/焦点返回、0 Works 写请求、0 页面错误、无横向溢出、相册和 metadata 不变；额外实际解码两条 retained raw BLOB，均 pending/topic-empty。最终截图与只读报告存于 owned 临时目录 `final-read-only-views-evidence`。这轮只读结果是最终 source 的独立 GREEN，没有重发链1人工/机器命令。

## 1. 交付内容（五能力 + 事件源）

| 能力 | 落点 |
|---|---|
| ① 平台聚合与统计 | `domain/works.go` `AggregateWorksStats`（作品/文件夹/接力过/最近在动计数），前端 `WorksAlbumPanel` 平台统计区 |
| ② 作品卡片与文件核实 | `adapters/worksscan` 有界扫描读 `mtime`/`file_count` 置 `verified`（"从文件核实"徽标）；JSON 投递不置 `verified` |
| ③ 交接手册 | `WorkHandoff{content,updated_at,source}`，`SetWorkHandoff` 记录更新时间，`source=human|generated` |
| ④ 态度与视图 | `Attitude{continue|park|done}` + 平铺/按平台/按分组/时间线 + 隐藏 + 排序（`FilterWorks`/`SortWorks`） |
| ⑤ 扫描导入与同步 | `worksscan.Scanner.ScanWorks`（有界）、`worksscan.Delivery.ParseDelivery`（AI 投递 JSON）、`RefreshWorks` 同步最近改动 |

事件源：`app.WorksService` 经 `WorksEventSink` 发 `domain.WorkEvent{work_id,kind,source,scope,topic,at,payload}`；kind 词汇 `work_imported|work_attitude_set|work_group_set|work_handoff_set|work_hidden|work_synced|work_verified`。SQLite `works_album` 持久化可替换观察、`works_metadata` 持久化人工元数据（态度/分组/隐藏/手册，revision CAS），冷重启一致。`works_requests` 保存 caller+command+key 的 pending/accepted/unknown/rejected 收据，未知效果不重复执行。

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

W0 P2 已发布七端点与生成类型；人类会话 + CSRF，写操作带 `schema_version=1/request_id/expected_version=1` 与 `Idempotency-Key`。metadata 的 expected_revision 独立；DTO 字段与 domain.Work/WorksSnapshot 对齐。

| 方法与路径 | 用途 |
|---|---|
| `GET /works` | 读缓存相册，返回 `WorksSnapshotResultV1{snapshot:{works,platforms,status,observed_at,failures}}` |
| `POST /works/refresh` | 有界扫描（body `WorksRefreshRequestV1{roots[]}`），返回同上 |
| `POST /works/delivery` | AI 投递 JSON（body `WorksDeliveryRequestV1{content:string}`，≤1MiB），返回同上 |
| `PUT /works/{work_id}/attitude` | body `WorkAttitudeRequestV1{expected_revision,attitude}` |
| `PUT /works/{work_id}/group` | body `WorkGroupRequestV1{expected_revision,group}` |
| `PUT /works/{work_id}/handoff` | body `WorkHandoffRequestV1{expected_revision,content}`（≤64KiB，utf8） |
| `PUT /works/{work_id}/hidden` | body `WorkHiddenRequestV1{expected_revision,hidden}` |

实际迁移 `016_works_album.sql`（W0 owner；含 works_requests）：

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

## 4. 链1事件源映射（纠正后 W1 Handoff）

W7 保留原始 WorkEvent 词汇，W1 独占闭集 Event.Kind 的合法转换。没有认知 topic 的事件保留空 topic，不能使用 platform 代替。

| 原始 kind | actor/source | W1 边界 |
|---|---|---|
| work_imported | manual（人提交 JSON） | collect；仅本次导入项，生成文本不表示任务成功 |
| work_attitude_set | manual | 显式态度可按 user_approve/bookmark 规则映射 |
| work_group_set | manual | 分组元数据，不表示态度审批 |
| work_hidden | manual | 隐藏/取消隐藏需按实际 hidden 值区分，不统一 user_veto |
| work_handoff_set | manual | 保存人工手册，不是闭集 manual kind，也不是 agent_task |
| work_synced/work_verified | machine | 系统观察，禁止增加人类注意力；不能转成 manual 关注信号 |

Payload 为作品活动记录，Scope 为作品根；生成手册不推断 agent_task 成功。默认空 topic 的认知处理和未知事件归 W1，不在 W7 冻结策略。sink 错误已显式返回 delivery_unknown；最终实际 W1/W0 接线和重启已核验 raw pending 落库，不能沿 sink collector 测试推断认知映射或 Signal 完成。

## 5. 依赖与剩余限制

- 正式 M0 `47f47f3`、P2 `5d35637`、最终 startup `396aec1` 已普通消费；W0 仍唯一 owner：契约、HTTP、generated 客户端、migration016、主程序。实际 assembly 已验证，本轨未自行改 cmd。
- W1 owns 原始事件归一化，topic 未明确保持 pending，不发 Signals；machine sync/verify 不增人类关注。用户待定 signal scale/真实会话/算法参数不在 W7 代选。
- 有界项目 metadata 扫描，不覆盖整盘、私密正文、凭据；最多64 roots、8层、10000 files，截断/读取失败显示 unknown/partial，不跟随 symlink 根。默认/请求根均由 W0 可信登记缓存验证。
- 真实浏览器与模型/执行/跨平台全文无关；人工态度不等于测试通过，生成手册不等于 Agent 任务成功。整体验收和个人预览升级归 root，未运行远程 CI/未合 main。

## 6. 验证

| 命令 | 结果 |
|---|---|
| `go test ./internal/workspace/... ./internal/adapters/worksscan/ ./internal/adapters/sqlite/ -run 'Works|W7'` | PASS |
| `go test ./tests/attention-engine/ -run TestW7` | PASS（真实临时目录扫描 + 冷重启一致 + 五类事件） |
| `go build ./...` | PASS |
| `pnpm --dir web typecheck` | PASS（exit0） |
| `pnpm --dir web exec vitest run src/pages/workspace/worksAlbumPresentation.test.ts` | 7 PASS |
