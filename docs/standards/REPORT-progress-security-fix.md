# 独立 REPORT：进度安全验收修正

## 任务标识

- **Task**: `task_7b4e93c1a3c1`
- **Dispatch**: `ctx_7230f56d8f19`
- **分支**: `standards-backend`

## SHAs

| 用途 | SHA | 说明 |
|---|---|---|
| 首 SOURCE | `7e3a4fbb5c67bf49f70ac60cfd80a77a7e682562` | 缓存 GET + 3 项真实测试 |
| 二 SOURCE | `3ddcf473d9391a2a71f1d261db904fec03be7022` | 5 项安全检查 |
| 本轮 SOURCE | 本提交前一个 | 自然到期 + HTTP 重启修正 |
| REPORT | 本提交 | 独立验证报告 |

## 本轮修正

### 1. 自然到期测试（TestProgressExpiredGrantGetRejected）

**问题**：前版用 SQLite UPDATE 把 `expires_at` 改到过去，不是真实到期。
**修正**：`GrantProjectAgent` 不再强制 `g.ExpiresAt = nil`，保留调用者指定的到期时间。`IssueProjectAgentToken` 保留已有到期，仅未指定时默认一小时。测试用 2 秒短到期 + 实际 `time.Sleep(3s)` 等待自然到期后验证 HTTP 和服务层均拒绝。不修改系统时钟，不用 SQL UPDATE。

### 2. HTTP 重启测试（TestProgressRestartPreservesCacheAndPermissions）

**问题**：前版只重开 DB 未通过 HTTP1 验证缓存和 token。
**修正**：第一轮真正启动 httptest.Server HTTP1，通过 HTTP GET session bootstrap → 人工缓存读取 → Agent token 验证均经 HTTP1 完成；然后关闭 HTTP1（server.Close）和 DB；第二轮重新打开同一 SQLite 文件启动 HTTP2，验证缓存和原 token 权限均持久化。

## 八项真实测试（全部 PASS）

| 实际命令 | 结果 |
|---|---|
| `go test -mod=readonly -p=2 ./internal/workspace/app -run '8项测试' -v -count=1` | EXIT 0，8 PASS，包 20.217s |
| `go build -mod=readonly -p=2 ./...` | EXIT 0 |
| `go vet -mod=readonly -p=2 ./...` | EXIT 0 |

## 剩余限制

- 8 项内存替身测试保留覆盖不计 PASS
- 未运行全量 Go 测试套件
- 未验证真实产品实例端到端验收
- PUT/infer 仍为人工-only
