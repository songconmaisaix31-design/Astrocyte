# S1 论文、CLI、进度与记忆收尾 W0（最终集成续）

当前 Dispatch `ctx_601d4fe7909f` / Task `task_7c7a06d13948`；开发 Worker OpenCode / `deepseek/deepseek-v4-pro`。独占共享契约、HTTP、入口、迁移 013+、生成 API、脚本、根依赖与所有锁；领域业务退原 owner。承接上一轮契约里程碑（`e39f8f1` 单一权威 GET/DTO 已接受），本轮完成 paper_snapshot 摄取、W2/W3 最终 SOURCE/REPORT 普通合并、进度契约收敛与真实联合链验收。

## 基线与合并

- 沿 `s1-sync-contract-1010` 普通合并（保留 owner 历史，未 squash/rebase/reset/force）：W1 精确 SOURCE+REPORT 合一 `0b189e5f`（paper_snapshot 离线摄取适配器 + IPv6 SSRF 收紧）、W2 精确 SOURCE `d8a2562`（进度 domain 收敛 + OpenCode loopback driver + durable publication）、W3 SOURCE `3206111` + REPORT `6111fd1`（UI 对齐 source_key/content_state/pdf_urls + paper_snapshot）。
- 最终续合并：W2 精确 SOURCE `a6e21b3`（session permission 严格 trailing `*/deny` 不跳过后置 specific-allow + 进度缓存 GET 缺省补齐授权 projectID，`37a7fa3`）、W1 精确 SOURCE `93d3c82` + REPORT `1ec6530`（精确 DOI/arXiv 标识路由到 registry 单条查询，只 metadata/reader 修复无契约改，`019db1e`）。

## 契约单一发布（W0 唯一 owner）

- **`ImportMaterialRequestV1.adapter`** 增 `paper_snapshot`（`662887e`）：人类复核的浏览器插件快照 JSON 作为 `export_text` 摄入，`source_locator=source_url`、`source_key` 留空由 reader 重推导（arxiv_id/ACL URL/DOI），body/provenance 原件 objects 保留，无网络/无 model。
- **`PaperSearchHitV1`** 维持 `source_key/provider/content_state/pdf_urls` 既定接口（`e39f8f1`），未再改换。
- **`InferProgressRequestV1`** 删除 `operation_id`（`cc211ce`）：对齐 W2 权威 domain，操作身份来自 `Idempotency-Key` header（`Caller.OperationID`），不进 body；`ProgressCommand{Files}` 单字段。`/papers/import` 死 501 入口已删；`ProjectProgressV1` 不泄 `operations/pending_operation`，`observed_at/source` 缺失显式 null，`status=unknown`。

## 组装与运输

- httpapi `paper.go`/`progress.go` 投影沿用；`progress.go` `mapProgress` 零值→`status=unknown/observed_at=null/source=null/evidence=[]`，权限识别与错误透传由 W2 domain 负责。
- `cmd/server`：`UVXPath` 接 `ASTROCYTE_UVX_PATH`；`PaperSearch`/`Progress` 组装沿用（未重定义 W1/W2 端口签名）。
- `scripts/paper-snapshot-capture.mjs`（`5152b30`）：隔离 Chromium（临时 profile，非个人 Chrome）加载 MV3 扩展，抓取真实公开页快照并存原始 JSON。

## 真实联合链验收（本 Task 新增）

| 验证 | 结果 |
|---|---|
| 真实快照抓取（隔离 Chromium + MV3 扩展，非个人 Chrome） | PLOS fulltext `readable_fulltext` 63934 字符 `doi:10.1371/journal.pdig.0000514`、ACL `abstract_only` `doi:10.18653/v1/2024.acl-long.1`，存 `%TEMP%/astrocyte-paper-snapshot/` |
| `node tests/s1/paper-snapshot-live.mjs` | PASS：真实 fulltext/abstract 快照经 `ImportMaterial(adapter=paper_snapshot)` → material/content/attachment byte 级保留、去重复用同一 material、冷 API 重启后 digest 一致；无 fixture/网络/model |
| `node tests/s1/progress-live.mjs` | PASS：真实 project-space→local-project 注册、缺席 GET→`status=unknown/observed_at=null/source=null`、人工 PUT→持久化 `source=human/percent=40`、冷 API 重启重查一致、无 model consent 的 infer→403 结构化错误（非 200 unknown 掩盖） |
| `GOFLAGS=-p=1 pnpm check` | PASS/exit0：gofmt/vet/mod、19 包架构、依赖、进程退出、Go 全测试、14 项 S1 acceptance、241 契约例、TS/lint、82 前端单测、diff |
| `GOFLAGS=-p=1 pnpm build` | PASS/exit0：Go 程序 + Vite 生产构建（101 模块） |

### 一次实际 Codex 进度推断（获准真实操作，`125697f`）

`node tests/s1/progress-infer-live.mjs`：在隔离验收库注册真实 Astrocyte 工作树（`root=CWD`），设置 `allow_directory + allowed_subdirs=["tasks"] + external_model_cli="codex"`，探针 `sessions/probe{cli:codex}` 观察 Codex 原生配置，提交一次 `POST .../progress/infer`（`Idempotency-Key` 为 UUID 操作身份，`files=["tasks/S1-final-W2.md"]`）。真实模型结果：`source=agent_inferred`、`model=gpt-6.1-sol`、`native_id=01a1270a-0483-7bb2-9bfa-b7f7367142ea`、`status=开发收尾，待功能补齐与集成验收`、证据 `tasks/S1-final-W2.md`（mtime+size 版本）、`observed_at` 真实、`cost=null` 未知未伪造；冷 API 重启后 freshAPI 重查一致、Mission=0。advisory 非 human-approved，仅这一次，未重发旧模型/UNKNOWN。

未重发旧 AT 成功媒体/模型与 UNKNOWN、未启用 paid/media、未读个人 Chrome/5173/8787、未安装插件、未改 host/DNS。真实 `paper_url`/`paper_pdf` 公网下载 vertical 仍被本机 fakeIP 阻断（公网校验正确拒绝），不标在线抓取成功。

## 给 owner 的 Handoff（已发 root `msg_e487a5379557`）

- **W3**（续接 `ctx_55f6713c28e7`，进行中）：`web/src/pages/workspace/progressClient.ts` + `ProjectProgressPanel.tsx` 仍在 infer body 送 `operation_id` 并维护独立 `operationId` ref；收敛 domain 用 `DisallowUnknownFields` 拒绝，浏览器 infer 流会 403/400。W3 owner 需删除 body `operation_id`，改用 `command.prepare` 生成的 `Idempotency-Key`（已是 UUID）作为操作身份，并修同 URL 更新 409 + PDF source_key。paper search/plugin snapshot UI 已对齐，无需再改。W3 由自有 `startS1Server({browser:true})` 做实际 UI + 快照并保留库。
- **W2**：`d8a2562` + 最终 `a6e21b3` 已普通合并并 release；进度 domain 错误退 owner。

## 剩余限制与未执行

- 三尺寸键盘/浏览器 search+已选快照批量端到端 e2e 属 W3（受其 operation_id 返修阻塞，W3 自行真实浏览器闭环）；W3 最终 SOURCE/REPORT 发布后由本轨做最后胶水合并与受影响 checks/build 一次。
- 真实进度 Agentinfer 已完成一次（见上表）；OpenCode textProcessor 仍 unsupported、不静默 switch provider。
- CLI 完整覆盖、长期记忆共享、公开源代理策略用户未答，未代选、未默认权限。未合 main/远程 CI、未升级个人预览；个人库备份已 root 完成 `%APPDATA%/astrocyte/backups/pre-s1-paper-cli-20261011-013712/`。
