# S1 Attention 一页执行计划

2026-10-09。基线 `d6c1bf2`（S0 + 已交付新布局），主控分支 `s1/attention-materials-20261009`。本轮用户明确授权 S1 完整功能、多 Agent 并行开发、主控验收和最终 push。

目标：真实 arXiv / summarize 结果导入、来源与内容版本、内容/主题/项目沉淀记录、关联与待查问题、依据和下一步齐备的候选、人工反馈、区分人类关注和机器使用、持久化作业及失败恢复。保留 Go/TS/SQLite 与新布局；只做 Attention，不扩展 Workspace/Swarm 执行。无 Python 默认链路；不新增调度框架或完成证明系统。

| 轨道 | 独占 write_paths | 依赖与验收 | 停止点 |
|---|---|---|---|
| W0 契约/传输/集成 | `contracts/`、`internal/attention/app/contracts.go`、`internal/adapters/httpapi/`、`internal/foundation/`、`cmd/`、`migrations/`、`web/src/api/`、根依赖与锁、`scripts/`、`README.md`、`tasks/S1-contract.md`、`tasks/S1-W0.md` | 先发布 DTO/端口/HTTP 契约；迁移与依赖唯一所有者。最后负责普通合并和少量组装胶水；HTTP/契约/构建检查 | 领域缺陷交 W1，存储/导入缺陷交 W2，页面缺陷交 W3 |
| W1 领域/应用 | `internal/attention/domain/`、`internal/attention/app/`（排除 `contracts.go`）、`tasks/S1-W1.md` | W0 契约；纯领域状态、三层沉淀及复用、候选版本/反馈、四维 unknown、人机关注、持久化队列用例；有意义的规则测试 | 公共契约/存储/入口更改交 W0/W2；不实现 S2 批准或执行 |
| W2 存储/导入/整理适配 | `internal/adapters/sqlite/`、`internal/adapters/importers/`、`internal/adapters/objects/`、`internal/adapters/distillers/`、`tasks/S1-W2.md` | W0 端口、W1 规则；SQLite 原子去重/CAS/作业恢复、对象发布、真实 arXiv 和 summarize 导出；用户已选择本机 Codex CLI 整理与原生读取隔离；重启与错误测试 | 迁移仅由 W0 落地；不新建视频下载/转录器，不调用 Python，不修改全局 CLI 配置 |
| W3 Attention 前端 | `web/src/`（排除 `api/`、`pages/workspace/`、`pages/swarm/`）、`web/e2e/`（排除 `s1.spec.ts`）、`tasks/S1-W3.md` | W0 生成客户端；真实导入/作业恢复/继续沉淀/候选反馈和详情来源链；两尺寸、键盘、异步状态；保留现有布局 | 不同时改三页；公共 API/依赖交 W0 |
| W4 验收 | `web/e2e/s1.spec.ts`、`tests/s1/`、`docs/acceptance/S1.md`、`tasks/S1-W4.md` | 发布契约后开始；可重复 AT01–AT04，真实网络材料另验；API/持久化/浏览器相互核对 | 未执行、失败、缺真实材料明确记录，不用 mock 代替真实导入 |

组织：每轨固定 1 Orca Worker + 1 worktree + 1 branch，独占写域并持续返修；完成 commit + push。W0 先发布契约，之后 W1/W2/W3/W4 并行；W0 在独立集成 worktree 普通合并，领域问题退原 Worker。主控只写本计划、决策、STATUS/HANDOFF，最终独立 check/build/E2E/真实验收后提交并 push。实际客户端和模型按启动与 Worker 回执记录，不猜测。

最终验收：AT01 真实论文/视频多轮沉淀有版本且没有 Mission；AT02 同源重复和摘要重试复用或产生新版本、没有重复工作；AT03 Agent 读取不增加人类关注；AT04 later 不丢资料且不是拒绝样本。`pnpm check`、`pnpm build`、`pnpm test:e2e` 必须通过。

用户新增明确要求：术语是“域”，算法排序；人主动分类、@文件；只有进入项目顶层空间后 Agent 才可主动认领读取。热度与权限分别处理，不采用“高层云自动开放”解释。真实材料确定为 arXiv `2504.16054` 与 B 站 `BV1PReT6EEqR`。沿用 W0 契约/迁移、W1 用例、W2 存储、W3 Attention 页面、W4 真实验收的原有写域，权限默认拒绝，不为此大改三页或扩展 Mission 执行。

已确认纳入动作：人在项目空间中 @文件，建立引用并纳入该空间，原資料留在原分类。沿用 Attention 写域实现人类分类、项目空间文件引用与排序；不移动原件、不扩展 Workspace 执行业务。

自动整理先接本机 Codex CLI，沿现有适配器/持久化作业实现，不新增开发框架。仅传明确选定资料正文与版本，不直接交付数据库；先验证当前 Windows 原生读取隔离。

待答：Agent 可读范围、私有资料外发范围、候选 readiness 关联要求。答复前继续真实公开材料获取、分类/@引用、Codex 整理及现有 S1 功能，不冻结未决定的共享接口，不外发私有资料。

主控收口：普通集成源 `e51676bca9508d698bd9968ca1f3e4777a7469a0`，最终报告普通合并 `01e3b0aa4947935fd9b31e06227d234d1277afb6`（只变文档）。`pnpm check` PASS（API14/14、单测40/40）、`pnpm build` PASS、完整 `pnpm test:e2e`144/144 PASS。真实论文人工两轮整理、候选 later 的浏览器/API/SQLite/重载及两尺寸目视核对通过；指定视频缺 summarize 真字幕/总结、授权 Agent 阅读范围和最新 Codex UNKNOWN 仍阻止完整四项收口。全部 settled Workers 输出已归档并释放终端，五条工作树/分支保留；仍可从原轨恢复，主控没有写业务代码。使用 Orca CLI 调度本机 Codex，实际开发模型 gpt-6.1-sol，权限运行探针与模型整理结果另见 W2，不从开发客户端推断原生能力。

用户最后重申视频复用 summarize：使用其既有导出与真实 provenance，无 Python、不另建下载/转录器。完整完成内容、真实限制和本机预览重启见 STATUS 顶部。

## 第一步发布：契约与写入口

面向 W0–W4 的统一索引，核对源为已发布 `af90eb77a1527c0ebebb46dc54a9c8d10da4ab8f`；不重建现有接口，也不把下列待定项视为已批准。详细边界见 [W0 契约](S1-contract.md)，HTTP 唯一事实源为 [OpenAPI](../contracts/openapi.yaml)，Go 端口为 [contracts.go](../internal/attention/app/contracts.go)，前端读取 [生成类型](../web/src/api/schema.d.ts)。接口、迁移序列、锁与程序入口继续只有 W0 一个所有者。

写请求共同字段：正文 `schema_version=1`、`request_id`、`expected_version`；请求头 `Content-Type: application/json`、`Idempotency-Key`、人类会话 Cookie 与 `X-CSRF-Token`，通过 Origin 校验。拒绝未知正文属性，单个 JSON 正文最多 8 MiB。HTTP 创建的 expected_version=1，更新使用当前聚合 version；存储内部 CAS 的 expectedVersion=0 才表示新增。聚合 version 与内容 revision 分开，不把资料元数据修改变成来源新版本。

以下路径统一带 `/api/v1`；每个写命令包含上述共同字段，字段细节以对应 OpenAPI schema 为准。

| 操作 / 输入重点 | 写入口 | 成功输出 |
|---|---|---|
| 导入：source_locator、source_key、kind、content_digest，及 adapter/export_text/local_file_ref | POST `/materials/imports` | 202 ImportJobV1；只表示已受理，按 job_id 查询结果 |
| 资料元数据：pinned、lifecycle、collection_reason | PATCH `/materials/{id}` | 200 MaterialDetailV1 |
| 主动人工使用：action，含 mention / project_reuse | POST `/materials/{id}/uses` | 200 VersionResultV1 |
| 人工沉淀：固定 input_refs、stage、question、processing_config、output_text 及该层关联/待查字段 | POST `/distillations` | 200 DistillationResultV1，含 reused |
| Codex 整理：固定 input_refs、stage、question、processing_config、可选 prior_distillation_ids | POST `/distillations/jobs` | 202 ImportJobV1；完成 Job.distillation_id 指向真实输出 |
| 候选：title、evidence_refs、purpose、next_step、dimensions、distillation_ids 等 | POST `/opportunities`；POST `/opportunities/{id}/revisions` | 创建201 / 修订200 OpportunityDetailV1 |
| 人工反馈：feedback、reason、可选 dimensions；作用于当前候选聚合版本 | POST `/opportunities/{id}/reviews` | 201 VersionResultV1；later→deferred，原资料保留，不作拒绝样本 |
| 人工分类域：title、description；详情字段见 MaterialDomainRequestV1 | POST `/material-domains`；PATCH `/material-domains/{id}` | 创建201 / 修订200 MaterialDomainResultV1 |
| 分类成员：domain_ids；空数组清除该资料分类成员关系 | PUT `/materials/{id}/domains` | 200 MaterialDetailV1，来源版本保持 |
| 项目顶层空间：ProjectSpaceRequestV1 | POST `/project-spaces` | 201 ProjectSpaceResultV1 |
| @引用：SourceRef；移除引用：RemoveMaterialReferenceRequestV1 | POST `/project-spaces/{id}/references`；POST `/project-spaces/{id}/references/remove` | 200 ProjectSpaceResultV1，原资料/原分类保留 |
| 人工四维排序配置：enabled、完整 weights | PUT `/attention-ranking-profile` | 200 RankingProfileDetailV1，当前配置与历史版本 |
| 作业重试 / 取消：共同字段，expected_version 为当前 Job.version | POST `/jobs/{id}/retry`；POST `/jobs/{id}/cancel` | 200 JobV1；UNKNOWN 重试409拒绝 |

读入口包括资料/候选/沉淀/作业列表和详情、固定 revision 的 content/attachments、分类域、项目空间、排序 profile，以及 GET `/distillations/processor` 的实际可用配置。查询/刷新不产生人工关注。SourceRef 固定 material_id + revision + locator，可选 span；stage 的传输枚举为 `content / topic / project`。四维缺证据保留 null/unknown；模型候选建议由人保存，不自动批准或生成 Mission。错误响应沿用 ErrorV1；非法输入400、未授权403、缺对象404、CAS/幂等冲突/未知外部结果409，未实现能力501、处理器不可用503。

数据表由 [003](../migrations/003_attention.sql)、[004](../migrations/004_attention_spaces.sql)、[005](../migrations/005_attention_ranking_profile.sql) 顺序管理；当前 Attention schema version=3。

| 迁移 | 表及持久化边界 |
|---|---|
| 003 | attention_materials（source_key唯一、聚合version）；attention_material_revisions（material/revision主键、material/content_digest唯一、不可变来源）；attention_distillations（reuse_key唯一）；attention_opportunities / attention_opportunity_revisions（当前候选与不可变修订，反馈保存在聚合data.reviews） |
| 003 | attention_jobs（dedupe_key唯一、version/status/data/payload/caller）；attention_receipts（caller/command/key联合主键、输入digest与完整回执）；attention_outbox（事件与聚合版本） |
| 004 | attention_domains、attention_project_spaces（各自version与聚合data；分类成员在Material.domain_ids，空间引用在空间data；均不授予权限） |
| 005 | attention_ranking_profiles、attention_ranking_profile_revisions（当前配置与不可变历史，无预置四维权重） |

聚合修改、回执、outbox 在同一事务内提交；CAS冲突全部回滚。同caller/command/key且语义输入相同返回旧回执，输入不同409。来源以 source_key + 实际字节 content_digest 去重；source_key/content_digest 字段仍必传，可传空字符串让适配器计算。对象存储先发布真实附件/正文，再引用到数据库；不在 SQL 事务中做外部下载/模型调用。

作业 status 的实际枚举仅为 `queued / running / succeeded / failed / cancelled`，HTTP 与003 CHECK约束一致。UNKNOWN 是独立 `delivery_unknown=true`，伴随 error.code=delivery_unknown，**不是第六种status**。CAS确保一个queued作业只被一个worker认领；尝试计数、max_attempts、deadline、operation_id持久保存。重启将遗留running记为failed；若外部调用已开始且没有已持久化结果，则UNKNOWN，禁止自动重发和原作业显式重试。已知失败仅在剩余次数/原截止时间内显式重试，保留operation_id、payload、caller；已保存的模型结果可用于本地发布恢复，不再次调用模型。取消不保证撤销已发生的外部效果。

身份由 [session.go](../internal/adapters/httpapi/session.go) 注入可信 Principal，不接受请求正文的 actor。GET `/auth/session` 创建12小时 HttpOnly/SameSiteStrict 本地人类会话，actor_id=local-human、actor_kind=human；写入还需CSRF。可配置Bearer只识别Agent，当前权限未决定期间拒绝受保护读写及人类会话引导，不能把有效token或高热度当访问授权。此本地会话信任本机所有者，不提供任意同用户进程/CLI的OS身份隔离。托管Codex仅收到明确选定固定版本正文，数据库和任意对象路径不作为模型工具；私有外发未授权。

各轨消费：W1用Go端口维护状态/CAS/人机信号；W2实现端口及上述表/对象约束，迁移变更交W0；W3只调用生成客户端并显示作业标记/真实缺失，权限不从前端字段推断；W4使用OpenAPI与实际API/SQLite核对，不把accepted或403当业务正路径通过。所有跨轨字段/状态调整先交W0，不手改生成类型或抢写迁移。

第一步本次核验：`pnpm check:contracts` PASS（226示例、生成类型一致；保留既有EventV1未引用警告）；`go test -mod=readonly ./internal/adapters/httpapi ./internal/attention/domain` PASS（缓存结果，非重新执行真实模型/视频）。本次只发布主控索引，不改业务代码，不重启或重放已结束作业。完整Agent可读范围、私有外发、ready门槛、A→B→A head以及Codex新作业仍见 [QUESTIONS](../docs/QUESTIONS.md)，未冻结为规则。

发布到共享分支，并通过 Orca 将 W0–W4 共同交接发布到原项目 Run。原 Dispatch 已 completed，直接投递返回 dispatch_inactive，随后按运行时指引改投 Run；没有为了通知而重新启动 Workers，不声称原 Workers 已读。后续恢复每轨时将本节和 W0 契约作为共同入口。
