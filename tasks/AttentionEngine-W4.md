# AttentionEngine W4 · Knowledge 两级存储 + 检索重排（M4）

事实源：任务书 §4.4 / §5.6 / §8 M4。基线 `attention-engine/v1-20261011`，实际分支 `songconmaisaix31-design/attention-engine-w4-knowledge`。

## 阶段状态（交接 Codex --yolo）

### Codex 接续检查（2026-10-11）

- 当前 Dispatch：`task_da85be3296c1` / `ctx_062b79f07e46`；客户端 Codex，模型以运行配置为准，未从客户端名称推断。
- 基线 `13a80ef`：`GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain/ ./tests/attention-engine/` 与同两包 `go vet` PASS，仅为纯逻辑证据。
- 独立修正：追加固化快照时复制所有 JSON 字节，调用方与返回值修改不能改变既有版本；补充两组显式 λr 下独立复算、排序变化与固定注意力验证。
- 修正后 `go test -count=1`、同两包 `go vet`、`git diff --check` PASS；SQLite 重启/引用关口/人工分享审计仍 NOT_RUN。
- 已通知协调者发布精确 M0 SHA、公共存储基础与 sqlite-vec spike 决策；发布前不实现依赖 M0 的 app/store，也不自建迁移或公共类型。
- 协调者回复：M0 尚未发布完整稳定 SHA，scope/version/share audit 需求已转 W0；sqlite-vec 结论待 owner 报告，不将未知当作真相似度。
- 独立阶段源码 `47a8950d475b612c0cc0e4e5e03ed2ef32131124` 已 push；既有 `go test ./internal/attention/app ./internal/adapters/sqlite` 与 `go build ./...` PASS（同限流环境），不包含新增 M4 持久化路径。

### 待 M0 后实际验收顺序

1. 在临时文件 SQLite 经正式迁移建立库，调用真实 knowledge application/store 创建 inbox 条目和来源引用；从另一项目读不到，inbox 引用被拒。
2. 同一条目两次 sleep 写入：fluid 只保留第二次，crystallized 两版内容与各自时间戳都存在；关闭连接、重新 Open，query_only 直接核查正式表，不以内存返回值替代。
3. 人工 confirm 后可固定引用；Agent confirm/share 被拒；拥有项目人工显式分享后目标项目可读，第三项目仍不可读，并核对审计中的 actor、源/目标项目、entry 与时间。
4. 给定固定注意力表、查询时间和显式 λr，从落库候选读取 relevance、age、boost 并逐项独立重算排序；查询前后直接比对 attention 状态与关联记录不变。
5. 真相似度或退化 relevance 只采用 W0 正式 spike 结论；共享 HTTP/cmd 接线与迁移修正请求提交 W0 Handoff，不越写域。

- **已完成**：两级存储纯规则（`domain/knowledge.go`：`KnowledgeStatus`/`ConfirmKnowledge`/`ReferencableKnowledge`/`CrystallizedVersion`/`AppendCrystallized`/`KnowledgeVisible`）+ 检索纯函数（`domain/retrieval.go`：`RetrievalScore`/`RankRetrieval`，λr 必传、无冻结默认）+ 验收脚手架 `tests/attention-engine/M4_retrieval_test.go`。commit `ab7d381`、`7294260` 已 push。
- **未完成 / NOT_RUN**：`app/knowledge_*.go`、`adapters/sqlite/{knowledge_store,vec}*.go`、端到端 `M4_*_test.go`（fluid 重写/crystallized 只追加/confirmed 引用/跨项目过滤+分享审计）。全部依赖 M0。
- **FAIL / UNKNOWN**：无本轨失败；M0 契约/迁移/sqlite-vec 结论状态 UNKNOWN（集成分支 `77712f3` 仍 docs-only，W0 分支未 push）。
- **下一步接口要求**：见下方 Handoff——`domain.KnowledgeStatus` 引用、`knowledge_*` 四张表 DDL、W0 存储基础（Open/ExecMigration/事务助手）、sqlite-vec 结论。M0 SHA 发布后实现 store/app/vec。

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
- `λr`：检索独立衰减，与状态更新 `λ` 无关；**可配置参数，无冻结默认值**（调用方显式传入）。
- `access_boost`：人工使用/提及/项目复用信号，基准 1.0。

本轨已落纯函数（`domain/retrieval.go`）：`RetrievalScore`（纯公式）、`RankRetrieval`（可复算、稳定排序、ID 决胜）。检索任何路径不写 `A(t)`；`A(t)` 演化只归 W2。

## 波 A 已交付（不依赖 M0 类型）

- `internal/attention/domain/retrieval.go`：`RetrievalScore`、`RankRetrieval`、`RetrievalCandidate`、`RetrievalRanking`（λr 为必传参数，无默认常量）。
- `internal/attention/domain/knowledge.go`：`KnowledgeStatus`、`ConfirmKnowledge`、`ReferencableKnowledge`、`CrystallizedVersion`、`AppendCrystallized`、`KnowledgeVisible`。
- 测试 `retrieval_test.go` / `knowledge_test.go`：公式复算、排序可复算、候选不可变、追加不覆盖、时间戳单调拒绝、隔离可见性。
- 验收脚手架 `tests/attention-engine/M4_retrieval_test.go`：固定注意力下排序独立复算（手工重算公式逐项核对）、检索不写回注意力。

验证：`go test ./internal/attention/domain/ ./tests/attention-engine/` PASS；`go vet` 两包 exit 0（GOFLAGS=-p=1 GOMAXPROCS=2）。

## Handoff → W0（M0 契约/迁移/存储基础）

1. **KnowledgeEntry 状态类型**：本轨已定义 `domain.KnowledgeStatus`（`inbox|confirmed`）。建议 `KnowledgeEntry.Status` 直接引用该类型，避免重复枚举。
2. **crystallized 落库不是单 JSON 列**：需迁移一张只追加版本表，`KnowledgeEntry.crystallized` 指向最新版本。请求 schema（DDL 提案，W0 写入 014+ 迁移）：

```sql
CREATE TABLE knowledge_entry (
    id            TEXT PRIMARY KEY,
    topic         TEXT NOT NULL,
    fluid         TEXT NOT NULL DEFAULT '',      -- 可 UPDATE，sleep 覆盖
    status        TEXT NOT NULL DEFAULT 'inbox', -- inbox | confirmed
    owner_project TEXT NOT NULL DEFAULT '',      -- '' = 个人作用域
    created_at    INTEGER NOT NULL
);

CREATE TABLE knowledge_crystallized (            -- 只 INSERT，不 UPDATE/DELETE
    entry_id    TEXT NOT NULL REFERENCES knowledge_entry(id),
    version     INTEGER NOT NULL,
    content     TEXT NOT NULL,                   -- JSON
    recorded_at INTEGER NOT NULL,
    PRIMARY KEY (entry_id, version)
);

CREATE TABLE knowledge_sources (                 -- EvidenceRef 视图 §4.5
    entry_id      TEXT NOT NULL REFERENCES knowledge_entry(id),
    event_id      TEXT NOT NULL,
    weight_at_use REAL NOT NULL,
    ts            INTEGER NOT NULL
);

CREATE TABLE knowledge_shares (                  -- 人工跨项目分享，审计
    entry_id  TEXT NOT NULL REFERENCES knowledge_entry(id),
    project   TEXT NOT NULL,
    shared_at INTEGER NOT NULL,
    PRIMARY KEY (entry_id, project)
);
```

3. **Store 接口**：本轨 `knowledge_store*.go` 需要 W0 的连接/迁移执行/`engine_common*.go` 共享基础设施（Open/ExecMigration/事务助手）。本轨将实现的接口形状（签名按 W0 类型对齐后落地）：

```go
type KnowledgeStore interface {
    // fluid 覆盖（可重写），crystallized 追加（不可变），confirm（inbox→confirmed），
    // 隔离可见列表，人工分享审计，检索候选装配（relevance + A + boost + age）。
}
```

4. **sqlite-vec**：`relevance` 由 W0 spike 决定接入或砍。接不上则本轨退化 `relevance` 为确定性标量（topic 精确命中 / 仅 recency+access），重排公式不变，退化在报告记录。
5. **依赖顺序**：本轨 store/app/vec 全部依赖 M0 发布（契约类型 + 迁移 + 存储基础 + sqlite-vec 结论）。M0 SHA 发布前只做纯函数与独立验证，不造 placeholder Store 或 public 类型。

## 波 B（M0 合入后）

- `app/knowledge_*.go`：confirm 端点胶水、confirm 前禁止引用、跨项目分享入口（人工关口）。
- `adapters/sqlite/knowledge_store*.go`：fluid 覆盖 / crystallized 追加 / 确认 / 隔离可见列表 / 分享审计，全部只经 W0 存储基础。
- `adapters/sqlite/vec*.go`：sqlite-vec 相似度（或退化标量）→ `relevance`。
- `tests/attention-engine/M4_*_test.go` 补齐：端到端验收——fluid 可重写、crystallized 只追加（SQLite 只读核查）、confirmed 才可引用、跨项目过滤与分享审计。

## 验收（M4）

同一注意力状态下检索排序可复算；fluid 可重写、crystallized 只追加（SQLite 只读核查）；confirmed 才可被 Project Workspace 引用；跨项目过滤 + 人工分享审计。
