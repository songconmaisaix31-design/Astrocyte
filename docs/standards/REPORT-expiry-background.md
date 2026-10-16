# 独立 REPORT：projects.go 恢复与自然到期后台验证

## 任务标识

- **Task**: `task_7b4e93c1a3c1`
- **Dispatch**: `ctx_7230f56d8f19`（续）
- **正式决定**: `msg_ef2ae7cc0b04`

## SHAs

| 用途 | SHA |
|---|---|
| 本轮 SOURCE | 本提交前一个 |
| REPORT | 本提交 |

## 本轮变更

### projects.go 恢复原语义

- `GrantProjectAgent`：恢复 `g.ExpiresAt = nil`（清空调用者指定值）
- `IssueProjectAgentToken`：恢复 `expiry := time.Now().UTC().Add(time.Hour)`（固定 1h）
- 追加提交保历史，未使用 Git 回滚

### 测试变更

- 移除 `TestProgressExpiredGrantGetRejected`（短期 ExpiresAt 未授权）
- 新增 `TestProgressTokenIssuedWithExpiry`：验证 1h token 签发和到期前可用性
- 新增 `progress_expiry_bg_test.go`：`TestProgressExpiryBackgroundVerification`（ASTROCYTE_VERIFY_EXPIRY=1 激活）

### 后台到期验证

- PID: 664
- 状态: **RUNNING**
- 签发真实 1h token，读取 ExpiresAt，等待自然到期后验证 HTTP/服务层拒绝
- 不改 DB/Clock/模型
- 完成后结果写入 `.verification/expiry-bg-result.json`

## 测试结果

| 测试 | 状态 |
|---|---|
| TestProgressCacheGetFillsAuthorizedProjectIdentity | PASS |
| TestProgressHumanSetAndCacheRead | PASS |
| TestProgressAgentReadUsesGrant | PASS |
| TestProgressCrossProjectGetRejected | PASS |
| TestProgressValidGrantWithoutReadContextRejected | PASS |
| TestProgressTokenIssuedWithExpiry | PASS |
| TestProgressAgentInferRejected | PASS |
| TestProgressRestartPreservesCacheAndPermissions | PASS |
| TestProgressExpiryBackgroundVerification | **RUNNING**（后台 PID 664） |

## 剩余

- 后台到期验证 RUNNING，完成后计 PASS
- 8 项内存替身测试保留覆盖不计 PASS
- 三个 REPORT Markdown 已归并到 backend.md 并移除
