# AttentionEngine-W1 · 事件流与归一化（M1）

事实源：`C:/Users/DW/Downloads/tabbit-code (2).md` §4.1/§4.5/§5.2/§8 M1。
分支：`songconmaisaix31-design/attention-engine-w1-events`，基线 `dab8fea`。
当前 Task `task_13ac51983837` / Dispatch `ctx_62837e7857f0`；客户端 Codex，沿启动配置执行，没有启动其他 Agent。
协调者确认运行配置为 Codex / gpt-6.1-sol / YOLO。

## 当前状态（2026-10-11）

M1 NOT_DONE；独立解析阶段已 push `bd33fdaf2810ce07a378b2b4d6e133af89d79c16`。
该阶段落 `internal/adapters/events/native_record.go`、`native_reader.go`、`s1_record.go`，
仅适配器内部观测解析，不定义跨轨共享 Event/Signal，不消费未发布 M0。
新增单项目 S1 缓存元数据读侧 `s1_metadata.go`，不启动 discovery、不读消息正文。
新增 canonical outbox 读侧 `s1_reader.go`：每页最多 128 条，稳定 ID 游标，完整 envelope 字段和原 payload 字节；不修改 s1 回执。
有界 snapshot 分页以 ID 排序，不把游标当增量时间水位；下一次完整回放从空游标开始并靠 event_id 去重。

M0 仍等待协调者精确稳定 SHA；W0 未提交代码只读，不复制或合并。
旧波 A 映射文档有错，本文件已重写；旧 SHA `1321b23`、`05065f3` 保留于普通 Git 历史，不能作为实现规格。

## 当前有效映射边界

- 每条源记录保留原载荷；映射未知的记录留 adapter pending，不能创建 core `unknown` kind 或流入 Dynamics。
- 导入/蒸馏 job 保留原 source kind/outcome；无论 succeeded/failed，都不是 agent_task 闭环或失败经验。
- NativeOperation accepted 只说明命令回执，不证明 Agent 任务成功；unknown 不等于 failure。
- source/session 时间为 Unix 毫秒；correlation_id 是请求关联，不冒充原生会话。
- token 源字段缺失/null 保持未知，不猜 0；普通 Read 不证明长时阅读，视频类型不证明暂停/回放/倍速。
- topic 未观测留空；缺少 project_id 不用 aggregate_id 猜项目 scope。
- later/deferred 不当 veto；用户明确 reject/revise/withdrawn 才可映射否决。
- material_imported/source_items_selected/project_space_referenced 可映射 collect。
- material_updated 的 pinned=true 可映射用户标记；原生明确 Edit/Write/MultiEdit/apply_patch 可映射 edit_file。
- Grep/Glob/WebSearch/WebFetch 可映射 search；单个明确 git commit 命令可映射 commit。
- 单条原生记录有多种工具行为时保留 pending，不猜单一类型。

真实 Claude/Codex 回放只允许本轮获准 Astrocyte 项目或原批准 scope 的既有 metadata-only 产出。
外部原生对话记录选取范围仍待用户答复；没有读取消息正文、提示词、文件正文、凭据。
原生选定文件 reader 已实现最多 128 条/1MiB、头部项目/session 匹配、符号链接/越界拒绝，返回截断状态；
此 reader 尚未用于读取私密原文件，不把本地 fixture 当真实回放。

## §5.2 固定权重与强度

| 信号 | 权重 |
|---|---:|
| 深度行为有直接证据 | 1.0 |
| edit_file/commit/search | 0.6 |
| bookmark/collect | 0.3 |
| 用户标记重要 | 1.5 |
| 用户否决/打回 | -0.8 |
| 明确 Agent 任务成功闭环 | 0.8 |
| 明确 Agent 任务失败 | -0.1，仅 confidence |

权重不在线学习。强度产品初值尚未决定；保持显式参数，测试 fixture 可显式传 1，不冻结默认、不按 tokens 缩放。
缺少 outcome/depth 的 observation 不推断对应信号。共享 Signal 失败路径由 feedback_failure 分类承载，仅供 confidence。

## M0 Handoff（已发协调者，等待 W0 统一）

1. 稳定发布 SHA 与 Event/Signal 最终代码。
2. §4 Event 字段保持原规格；另统一 adapter observation wrapper：稳定 receipt/event_id、源 kind、可空 token、unknown outcome/depth、可选映射 Event。
3. event 同一行存 rawbytes，避免 encoding/json.Marshal 压缩 RawMessage 丢原字节；不新增 evidence 表。
4. pending 未映射观测能入库，不能造新核心 kind；signal 去重键支持 replay 幂等。
5. 现有 s1 outbox 为 canonical 事件入口；不重复把 Usage/outbox 同一行为算两次。
6. 明确 replay 元数据来源与 provenance；缺失真实来源保留 NOT_RUN/RED，不造样本。

## 验证与首失败

- `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/adapters/events` PASS，3 项：未知语义、逐条原字节与有界回放、s1 job/deferred 不造 Agent 结果。
- 后续独立 source 阶段同命令 PASS，5 项；新增缓存项目范围/截断和 outbox 原 payload 字节/分页收据对账。
- `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine -run TestM1ExistingS1ActionOutputPath -v` PASS：既有 s1 Service.UpdateMaterial→事务 outbox→新读侧→SQLite 冷重开，ID/envelope/payload 字节一致；新 Engine Event/Signal 段仍待 P1。
- `ASTROCYTE_M1_METADATA_DB=C:/Users/DW/AppData/Roaming/astrocyte/state.sqlite ASTROCYTE_M1_PROJECT_ROOT=C:\Users\DW\orca\Astrocyte GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tests/attention-engine -run TestM1CachedNativeMetadataLive -v` 首次 FAIL：claude=0/codex=0，partial=false。
- 该失败只表示精确选定缓存/项目没有可用来源，不表示本机总会话为 0；原库只读，不重新发现或扩大范围。
- root 另行只读确认该项目缓存存在、contributors 为空、cacheStatus=partial；初次输出 partial=false 只描述本次 record-bound 没有截断，不代表全局缓存完整。
- Windows 范围键现在规范 slash/case；Unix 仍按路径大小写精确比较。不重跑已确认空场。
- contract_local 已验证独立解析；真实 Event→Signal 与 SQLite 冷重启、幂等回放验收尚 NOT_RUN。
- M1 不宣称完成；等待不是 worker_done。

## 后续验收

稳定 M0 合入后：Event→Signal 落码、event_store、s1 outbox 经真实应用路径消费，
原始 rawbytes 冷重启一致、稳定 ID 重放不重复、真实 Claude/Codex 元数据回放 provenance 和未知字段保留，
§5.2 各类固定权重抽检及 failure confidence-only。

## 独占写域

`internal/attention/domain/{event,signal}*.go` 与 `internal/attention/app/{event,signal}*.go`，排除所有 `*_contract*.go`；
`internal/adapters/events/**`；`internal/adapters/sqlite/event_store*.go`；`tests/attention-engine/M1*`；本文件。
契约/迁移/入口交 W0；状态、规划、最终集成由 root 管理。
