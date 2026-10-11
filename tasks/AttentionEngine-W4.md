# AttentionEngine W4 · Knowledge 两级存储与检索（M4）

事实源：用户任务书 `tabbit-code (2).md` §4.4 / §5.6 / §8 M4；分支 `songconmaisaix31-design/attention-engine-w4-knowledge`。本次 Dispatch `task_da85be3296c1` / `ctx_062b79f07e46`，协调者确认客户端/模型为 Codex / gpt-6.1-sol。

## 当前结果（2026-10-11 22:43 +08:00）

M4 的 10 个实际验收例 PASS；SQLite 全包首次 RED 保留。已普通消费原 W7 owner 精确修复 `a3eab2c6d0df026e056a5a873a0d63bf58282688`，对应失败例定向复验 PASS；没有宣称 SQLite 全包重新跑绿。

- 最终 SOURCE：`82e4340e50943d355ecd81a2a47d65b7e3929f5a` 已 push；实现阶段 `73ba73b29ce3694dce7624a942fb26bff887c2eb` / `74ff62d95d9c44a95144eb08793762c9e41cb3e1` 已 push。
- 验证组合：本轨实现 + W0 `875980708d11cdf3ab392eb8ee3c386aa2617866` + W7 `a3eab2c6d0df026e056a5a873a0d63bf58282688`，全部普通合并；最后增加协调者要求的验收库保留入口。
- 共享契约、迁移017/018、只读事务、HTTP治理与会话 guard 均由 W0 发布精确 SHA 后普通合并；本轨未修改公共契约/迁移/HTTP/cmd。

## 完成内容

- 实际 `KnowledgeService` / `KnowledgeRepository` 复用 Go + modernc SQLite 和 `EngineRepository`，提供来源校验、fluid 重写、固化版本追加及人工治理。
- fluid 是单行当前摘要，覆盖后关闭连接、重新 Open，再以 SQLite query_only 直接核对第二次值；同事务失败会回滚。
- crystallized 是独立不可变时间戳版本表；新事实追加两版并保留原版字节，直接 UPDATE/DELETE 被正式 append-only trigger 拒绝。最新指针与旧历史都可重启重查。
- confirm/demote/share/reference 使用受信 Principal、CommandMeta、乐观版本和真实审计；既有 `attention_receipts` 按 `Knowledge.*` 命名空间保存完整重试结果。新的 tracing request ID 不重复写入。
- 默认项目隔离，只有数据库中匹配 owner/target 的人工分享才扩大范围；第三项目仍拒绝。Agent 无授予项目范围时拒绝，人工确认/分享操作拒绝 Agent。
- confirmed 才可引用；正式 reference 行记录项目、知识版本、固定 crystallized 版本与 human actor。后续追加第二版后，已引用第一版的行保持不变。
- 检索读取正式 SQLite confirmed 候选、固定注意力、内容年龄和 audited reference boost，在 deferred read snapshot 中重排；query_only 测试证明没有注意力写入，data/attention/updated_at/version 前后相同。
- W0 sqlite-vec spike 已 CUT：采用 topic 精确命中/持久化邻接/literal 关键词回退，结果明确携带 `sqlite_vec_cut:topic_keyword`，没有新的向量引擎或真相似度宣称。

## 配置保持显式

重新确认策略与年龄起点由协调者询问用户，尚未选择产品默认。

- `KnowledgeCrystallizedUpdate.Policy` 必须显式选择 `KnowledgeReconfirmVersion` 或 `KnowledgeKeepConfirmation`；缺失会拒绝。fluid-only 重写保留既有确认状态和固化历史。
- `KnowledgeQuery.AgeBasis` 必须显式选择 `KnowledgeAgeCreated` 或 `KnowledgeAgeConsolidated`；两种路径各自独立复算通过。
- `KnowledgeQuery.LambdaR` 为必传指针，单位 seconds^-1，nil 不执行检索，显式零可用，无默认 λr。`ReferenceBoost` 是调用方显式给出的 audited reference 增益，零保持基准 1。
- `consolidated_at` 只在 fluid/crystallized 写入时更新，人工动作仅更新 `updated_at`；旧 NULL 保持未知，不能从确认/分享时间伪造内容时间。

## 验证命令与结果

环境统一 `GOFLAGS=-p=1 GOMAXPROCS=2`。

| 命令 | 实际结果 |
|---|---|
| `go test -count=1 ./tests/attention-engine -run TestM4 -timeout 90s` | PASS，10例，1.428s；按协调者要求保留库后同10例 PASS，1.505s；包含 production HTTP handler + 真实 SQLite + 冷重启 |
| `go test -count=1 ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite ./internal/adapters/httpapi` | domain/app/httpapi PASS；sqlite RED：W7 `TestWorksRequestReceiptsPersistAndPreventUnknownReplay` 重复 CREATE `works_requests` |
| `go test -count=1 ./internal/adapters/sqlite -run TestWorksRequestReceiptsPersistAndPreventUnknownReplay` | 消费原 owner `a3eab2c` 后定向 PASS，1.191s；不是全包重跑 |
| `go vet ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite ./internal/adapters/httpapi ./tests/attention-engine` | PASS，独立命令 exit0 |
| `go build ./internal/attention/... ./internal/adapters/sqlite ./internal/adapters/httpapi ./cmd/server` | PASS，独立命令 exit0 |
| `git diff --check` | PASS |

输入是 contract_local 的受控事件/注意力/策略 fixture；数据库、真实 HTTP 会话/CSRF/服务/网络请求与重启是实际本地执行。没有把受控输入称为真实个人认知流、用户人工验收、M2 动力学或 M6 全链路验收。

## 保留首次失败及修复

1. 真实 fluid 重启/confirm 测试首次 RED：JSON null 被当成固化历史内容，随后发现 action INSERT 7 列用了8占位符。归一化 null、修正 SQL 后定向 PASS。
2. 持久化检索在 query_only 下首次 RED：共享 BEGIN IMMEDIATE 被 SQLite 拒绝。W0 发布 deferred `WithReadTx` 后同例 PASS。
3. standalone KnowledgeGovernance HTTP 首次 RED：session bootstrap404。W0 guard 补治理服务后真实 HTTP/session/CSRF/governance/restart 全例 PASS。
4. 最终受影响 SQLite 包首次 RED：W7 测试重复 CREATE 正式迁移表。未改非写域测试；原 owner `a3eab2c` 修复后单例 PASS，前述定向绿色不替代此历史失败。

## 协调者可独立只读核查的留存库

`ASTROCYTE_M4_ARTIFACT_DIR` 只在明确设置为新绝对目录时保留验收材料，存在同名数据库会拒绝覆盖；正常测试仍使用临时库。此次根目录：

`C:\Users\DW\AppData\Local\Temp\astrocyte-m4-6b799aaebe6c468896492f068d317ad6`

- `TestM4PersistentRetrievalRecomputedAndAttentionUnchanged/knowledge.sqlite` 与 `retrieval.json`：实际 confirmed 候选、未确认和跨项目排除样本、两个固定 attention 节点、审计 reference；JSON包含查询 λr/age/now/boost、返回候选因子/得分及前后 attention SQL 字符串。
- 独立重算输入：Now=10000ms，created_at 年龄，ReferenceBoost=2；a 的 relevance=1、A=.6、boost=3、age=8s；b 的邻接 relevance=.25、A=.9、boost=1、age=6s。λr=0 时 a 首位，λr=1.5 时 b 首位。均为显式受控参数，不是产品默认。
- `TestM4CrystallizedVersionsAndPinnedReferencesAfterRestart/knowledge.sqlite`：两版 `{"fact":"one"}`/`{"fact":"two"}`，时间3000/7000ms，fluid=second；pin-first 固定 knowledge_version=3 / crystallized_version=1。
- `TestM4ActualHTTPGovernanceAndColdStore/knowledge.sqlite`：实际 localhost HTTP 的 local-human 审计、p1→p2 分享、p2 reference，重启后核查。
- 其余4个 M4 存储例也各留 `knowledge.sqlite`，未复制到个人库；读取时请在新连接开启 query_only。

## 集成 Handoff 与真实剩余限制

- W0：组装 `app.NewKnowledgeService(sqlite.NewKnowledgeRepository(engineBase), options)` 到 `httpapi.Services.KnowledgeGovernance`。`GetKnowledge(ctx,p,id,project)` 与公共命名接口一致；主程序接线仍由 W0 owner 完成，单独实际 HTTP 装配已验。
- W3：`sqlite.BindKnowledgeTx(existing EngineTx)` -> `app.KnowledgeTx`；`app.RewriteKnowledgeFluidTx(tx, KnowledgeFluidUpdate)` 加入既有 Sleep 事务，不另开事务。`ListKnowledge(owner,false)` 包含分享项，写入前必须筛选 `StoredKnowledge.OwnerProject == owner`。需要正式迁移018才有独立内容时间。
- 本轨没有 raw-to-fluid 关键事实提取约定；W3 确定性 FactProjector 与不可测指标显式标记由 W3 负责，EventID 召回不能冒充事实召回。矛盾待裁决保持原知识，由 W2/W3 Conflict owner 处理。
- 推荐消费端通过 `RetrieveKnowledge` 使用真实读侧；λr/年龄/版本策略仍需调用方明确配置，产品默认待用户决定。
- 主控综合 check/build、M6、个人预览升级、main 合入与远程 CI 不属于本轨已执行结果；产品默认待答保持显式。W4 可按本报告的真实 M4 验收结果结算，主控仍需独立只读复算和普通集成。
