# AttentionEngine W4 · Knowledge 两级存储 + 检索重排（M4）

事实源：任务书 §4.4 / §5.6 / §8 M4。基线 `attention-engine/v1-20261011`，实际分支 `songconmaisaix31-design/attention-engine-w4-knowledge`。

## 范围与分层

- 链2 末端：Sleep 巩固 → Knowledge Graph → 用户确认 → Project Workspace。
- 本轨交付：`KnowledgeEntry` 两级存储语义 + 检索侧重排（§5.6）。KnowledgeEntry 结构体属 W0 契约（`*_contract*.go`），本轨只写 `domain/{knowledge,retrieval}*.go` 内的纯规则与 `app/knowledge_*.go`、`adapters/sqlite/{knowledge_store,vec}*.go` 的落库/检索。

## 两级存储语义（§4.4）

- **fluid**：流动摘要，即时推理用，随时被 sleep 重写。单值、可覆盖、无历史。
- **crystallized**：固化快照，结构化事实，**只追加带时间戳版本**。事实演化历史不丢（Zep 时序图简化）。落库为独立版本表，`crystallized` 字段指向最新版本。
- **status**：`inbox → confirmed`。`confirmed` 才可被 Project Workspace 引用。确认是人工关口动作，`ConfirmKnowledge` 幂等，其余状态拒绝。
- **可见性**：默认项目隔离，`KnowledgeVisible` 判定 owner == project 或显式 `knowledge_shares` 命中；人工跨项目分享带审计。

## 检索侧重排（§5.6，与状态更新严格分离）

`score = relevance × e^(-λr·Δt) × A(topic) × access_boost`

- `relevance`：sqlite-vec 相似度（W0 spike 定接入/退化）。
- `A(topic)`：条目挂靠主题当前注意力强度，纯读不写。
- `λr`：检索独立衰减（`DefaultRetrievalLambdaR = 1/(7d)`），与状态更新 `λ` 无关。
- `access_boost`：人工使用/提及/项目复用信号，基准 1.0。

本轨波 A 已落纯函数（`domain/retrieval.go`）：`RetrievalScore`（纯公式）、`RankRetrieval`（可复算、稳定排序、ID 决胜）。检索任何路径不写 `A(t)`；`A(t)` 演化只归 W2。

## 波 A 已交付（不依赖 M0 类型）

- `internal/attention/domain/retrieval.go`：`RetrievalScore`、`RankRetrieval`、`RetrievalCandidate`、`RetrievalRanking`、`DefaultRetrievalLambdaR`。
- `internal/attention/domain/knowledge.go`：`KnowledgeStatus`、`ConfirmKnowledge`、`ReferencableKnowledge`、`CrystallizedVersion`、`AppendCrystallized`、`KnowledgeVisible`。
- 测试 `retrieval_test.go` / `knowledge_test.go`：公式复算、排序可复算、候选不可变、追加不覆盖、时间戳单调拒绝、隔离可见性。

验证：`go test ./internal/attention/domain/` PASS（GOFLAGS=-p=1 GOMAXPROCS=2）；`go vet` 同包 exit 0。

## Handoff → W0（M0 契约/迁移/存储基础）

1. **KnowledgeEntry 状态类型**：本轨已定义 `domain.KnowledgeStatus`（`inbox|confirmed`）。建议 `KnowledgeEntry.Status` 直接引用该类型，避免重复枚举。
2. **crystallized 落库不是单 JSON 列**：需迁移一张只追加版本表，`KnowledgeEntry.crystallized` 指向最新版本。请求 schema 包含：
   - `knowledge_entry(id TEXT PK, topic TEXT, fluid TEXT, status TEXT, owner_project TEXT, created_at INTEGER)`；`fluid` 可 UPDATE，其余列不覆盖。
   - `knowledge_crystallized(entry_id, version INTEGER, content JSON, recorded_at INTEGER, PRIMARY KEY(entry_id, version))` 只 INSERT。
   - `knowledge_sources(entry_id, event_id, weight_at_use REAL, ts INTEGER)`（`EvidenceRef` 视图，§4.5）。
   - `knowledge_shares(entry_id, project, shared_at INTEGER)`（人工跨项目分享审计）。
3. **Store 接口**：本轨 `knowledge_store*.go` 需要 W0 的连接/迁移执行/`engine_common*.go` 共享基础设施（Open/ExecMigration/事务助手），接口形状按 W0 发布后对齐；本轨不造 placeholder Store。
4. **sqlite-vec**：`relevance` 由 W0 spike 决定接入或砍。接不上则本轨退化 `relevance` 为确定性标量（topic 精确命中 / 仅 recency+access），重排公式不变，退化在报告记录。

## 波 B（M0 合入后）

- `app/knowledge_*.go`：confirm 端点胶水、confirm 前禁止引用、跨项目分享入口。
- `adapters/sqlite/knowledge_store*.go`：fluid 覆盖 / crystallized 追加 / 确认 / 隔离可见列表，全部只经 W0 存储基础。
- `adapters/sqlite/vec*.go`：sqlite-vec 相似度（或退化标量）→ `relevance`。
- `tests/attention-engine/M4*`：端到端验收——同一注意力状态排序可复算；fluid 可重写、crystallized 只追加（SQLite 只读核查）。

## 验收（M4）

同一注意力状态下检索排序可复算；fluid 可重写、crystallized 只追加（SQLite 只读核查）。
