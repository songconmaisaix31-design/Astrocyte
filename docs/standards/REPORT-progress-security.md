# 独立 REPORT：进度安全验收补全

## 任务标识

- **Task**: `task_7b4e93c1a3c1`
- **Dispatch**: `ctx_19150385309b`
- **分支**: `standards-backend`
- **基线**: `4ec19cddb24068e09597357b0aabc6ecb9420930`

## SHAs

| 用途 | SHA | 说明 |
|---|---|---|
| 首 SOURCE | `7e3a4fbb5c67bf49f70ac60cfd80a77a7e682562` | 缓存 GET 核心实现 + 3 项真实测试 |
| 后继 SOURCE | `3ddcf473d9391a2a71f1d261db904fec03be7022` | 5 项安全检查 + backend.md 更新 + REPORT 归并 |
| REPORT | 本提交 | 独立验证报告 |

## 八项真实 HTTP/SQLite 测试（全部 PASS）

| # | 测试名 | 验证内容 | 结果 |
|---|---|---|---|
| 1 | TestProgressCacheGetFillsAuthorizedProjectIdentity | 缓存不存在时 GET 返回 identity，不写入行 | PASS |
| 2 | TestProgressHumanSetAndCacheRead | 非人工/CSRF/验证拒绝，人工写入后 SQLite 持久化 | PASS |
| 3 | TestProgressAgentReadUsesGrant | 无 grant 拒绝，有 grant+token 接受，撤销后拒绝 | PASS |
| 4 | TestProgressCrossProjectGetRejected | 项目 A token 访问项目 B 端点 403，数据不受影响 | PASS |
| 5 | TestProgressValidGrantWithoutReadContextRejected | 合法 Grant 仅含 observe 不含 read_context 时 403 | PASS |
| 6 | TestProgressExpiredGrantGetRejected | 到期前成功，SQLite UPDATE expires_at 到过去后 403 | PASS |
| 7 | TestProgressAgentInferRejected | Agent POST /infer 和 PUT 均被 scopedAgentRoute 拒绝 | PASS |
| 8 | TestProgressRestartPreservesCacheAndPermissions | 同 SQLite 关闭重开后缓存、grant、token 均持久化 | PASS |

测试环境：`httptest.Server`、真实 `sqlite.DB`、真实迁移、cookie jar、CSRF token、Agent bearer token、`ErrorEnvelope` 验证。

| 实际命令 | 结果 |
|---|---|
| `go test -mod=readonly -p=2 ./internal/workspace/app -run '8项测试' -v -count=1` | EXIT 0，8 PASS，包 4.351s |
| `go build -mod=readonly -p=2 ./...` | EXIT 0 |
| `go vet -mod=readonly -p=2 ./...` | EXIT 0 |

## 八项内存替身测试（保留覆盖，诚实记录不计 PASS）

使用 `progressProjects`、`progressStore`、`progressProcessor`、`progressFiles`、`progressRegistry` 替身，不能验证真实 HTTP 路由、SQLite CAS 持久化、session cookie/CSRF 边界、Agent bearer token 认证和 `scopedAgentRoute` 授权。保留为推理生命周期回归覆盖。

## 未做事项

- 未运行全量 Go 测试套件（前次全仓 RED 状态未改变）
- 未修改 `internal/workspace/app/progress.go` 领域逻辑
- 未验证真实产品实例端到端验收（待 H/F 原实例）
- 未运行 paper-dom.mjs 验证（已在 2644083 完成）
- PUT/infer 仍为人工-only，未做 Agent 写入支持
- 8 项内存替身测试未迁移到真实 SQLite/HTTP 边界
