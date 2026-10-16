# 独立 REPORT：进度缓存 GET 读取与 per-request Grant 检查

## 任务标识

- **Task**: `task_7b4e93c1a3c1`
- **Dispatch**: `ctx_e85083f7cd56`
- **分支**: `standards-backend`
- **基线**: `4ec19cddb24068e09597357b0aabc6ecb9420930`
- **模块**: B (Backend)

## SHAs

| 用途 | SHA | 说明 |
|---|---|---|
| SOURCE | `7e3a4fbb5c67bf49f70ac60cfd80a77a7e682562` | 进度缓存 GET 核心实现 + 真实测试 |
| DOCS | `f40aad77be6e9681c0d390fca01629ae0040d7bd` | backend.md 文档更新 |
| REPORT | 本提交 | 独立验证报告 |

## 变更文件（SOURCE `7e3a4fb`）

| 文件 | 变更类型 | 说明 |
|---|---|---|
| `internal/adapters/httpapi/progress.go` | 修改 | HTTP 适配层注释翻译中文，缓存 GET 读取不写入/不触发模型 |
| `internal/adapters/httpapi/session.go` | 修改 | `scopedAgentRoute` 增加 `progress` GET 路由允许 Agent token |
| `internal/workspace/app/progress_test.go` | 修改 | 9 块英文注释翻译中文，移除 3 项测试到 actual 文件 |
| `internal/workspace/app/progress_actual_test.go` | 新增 | 3 项真实 HTTP/SQLite 测试 |

## 实际验证结果

### 三项真实测试（计 PASS）

| 测试名 | 验证内容 | 结果 |
|---|---|---|
| `TestProgressCacheGetFillsAuthorizedProjectIdentity` | 缓存不存在时 GET 返回 identity（status=unknown），不写入行 | PASS |
| `TestProgressHumanSetAndCacheRead` | 非人工 PUT 拒绝，CSRF 缺失拒绝，验证失败拒绝，人工写入后 SQLite 持久化并 GET 读回 | PASS |
| `TestProgressAgentReadUsesGrant` | 无 grant Agent 拒绝，有 grant+token 接受，撤销后再次拒绝 | PASS |

测试环境：`httptest.Server`、真实 `sqlite.DB`、真实迁移、cookie jar、CSRF token、Agent bearer token、`ErrorEnvelope` 验证。

### 八项内存替身测试（保留覆盖，诚实记录不计 PASS）

`TestProgressInferRecordsEvidenceAndIsIdempotent`、`TestProgressInferRejectsInvalidModelOutput`、`TestProgressInferPersistsPendingReceiptBeforeModel`、`TestProgressInferRechecksConsentBeforePublication`、`TestProgressHumanSetPreservesOperationReceipts`、`TestProgressPublicationRecoversFromCASConflictWithoutNewModel`、`TestProgressPublicationFailureIsNotReportedSuccess`、`TestProgressComputedResultRecoveredAfterPublicationFailure`

限制原因：使用 `progressProjects`、`progressStore`、`progressProcessor`、`progressFiles`、`progressRegistry` 内存替身，不能验证真实 HTTP 路由、SQLite CAS 持久化、session cookie/CSRF 边界、Agent bearer token 认证和 `scopedAgentRoute` 授权。保留为推理生命周期回归覆盖，待迁移到真实原生条件后计 PASS。

### 编译与静态检查

| 命令 | 结果 |
|---|---|
| `go build -mod=readonly -p=2 ./...` | EXIT 0 |
| `go vet -mod=readonly -p=2 ./...` | EXIT 0 |

### paper-dom.mjs

catch 块移除已在先前 SOURCE `2644083960649d7c50262ec69fff730b411241a2` 完成（9/9 + 2/2 + 4/4 PASS），不在本轮 dispatch 范围内。

## 实现要点

1. **缓存 GET 读取**：`GetProjectProgress` 每次请求通过 `progressProject()` 逐次检查 `read_context` 的 grant/action/expiry/revoke，不触发写入或模型调用
2. **Agent 路由授权**：`scopedAgentRoute` 允许 `parts[4] == "progress" && r.Method == http.MethodGet`，PUT/infer 仍仅限人工
3. **跨项目拒绝**：Agent token 的项目范围与请求路径项目 ID 逐次比对，不匹配直接 403
4. **缺失状态映射**：`mapProgress` 中缺失状态为 `"unknown"`，缺失 source/observed_at 为 `null`

## 未做事项（诚实记录）

- 未运行全量 Go 测试套件（前次全仓 RED 状态未改变）
- 未修改 `internal/workspace/app/progress.go` 领域逻辑（本轮仅改 HTTP 适配层和测试）
- 未验证真实产品实例端到端验收（待 H/F 原实例）
- 未重复 paper-dom.mjs 验证（已在 2644083 完成）
- 未运行广域审计（前轮 dispatch 已完成）
- PUT/infer 仍为人工-only，未做 Agent 写入支持
- 8 项内存替身测试未迁移到真实 SQLite/HTTP 边界
