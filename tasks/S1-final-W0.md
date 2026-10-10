# S1 论文、CLI、进度与记忆收尾 W0

当前 Dispatch `ctx_a785bfe63b44` / Task `task_3097e734e4d0`；开发 Worker 为 OpenCode，执行环境观察到的模型标识 `deepseek/deepseek-v4-pro`（DeepSeek V4 Pro），与 root 在计划中记录的 TUI 观察一致。独占共享契约、HTTP、入口、迁移 013+、生成 API、脚本、根依赖与所有锁；其他领域由原 owner 交付，最后普通合并精确已 push 提交。本轮先完成契约里程碑并推送给 W1/W2 领域 owner，统一集成在收到 W1/W2/W3 精确 SOURCE/REPORT 后进行。

## 基线与边界

- 工作树 `s1-sync-contract-1010` 已在 HEAD `1b8dfd1` 普通合入 root 精确基线 `8277667`（合并提交保留原 owner 历史），未 squash/rebase/reset/force，未删除工作树贡献。
- 已读 AGENTS/HANDOFF/STATUS、`S1-final-paper-cli-memory-plan.md` 与 QUESTIONS 最新段。上一 W0 因 runtime 恢复遗留的未提交契约与迁移 013 保留在本次工作树中，本场验证后提交。
- 论文搜索＋正文、允许显式批量、允许获准项目 Agent 依据 TASK/STATUS 推断进度，均已由用户确认。新增完整 CLI 覆盖与长期记忆可见范围两项仍 PENDING，未代选、不据此降低验收或冻结接口。
- 复用现有 same-origin HttpOnly/SameSiteStrict 人类会话与 CSRF；不扩大 CORS、不安装未认证 daemon、不改个人 5173/8787、不读取个人 Chrome 或凭据。记忆策略未定，故本轮不新增记忆接口与迁移。

## 契约里程碑（本轮交付）

已交付最小可消费契约与存储迁移，供 W1/W2 在同一 Service 上实现领域而无需重构服务或并行为此改架构：

- **论文检索（attention，供 W1 实现）**：`attention/app/contracts.go` 新增 `PaperSearchService.SearchPapers`、`PaperMetadata`、`PaperSearchCommand`、`PaperSearchResult`。检索是临时公开元数据，不导入、不取全文、不调模型；`content_state` 明确 `readable_fulltext/abstract_only/paywall/restricted`，受限/付费不被绕过。人选来源后沿既有 `ImportMaterial`（kind=paper、按站点 adapter、source_key 去重、新 revision）与 job 重试机制入库，不新增第二条导入路径。
- **进度（workspace，供 W2 实现）**：`workspace/app/contracts.go` 新增 `ProgressService.GetProjectProgress/InferProjectProgress`、`ProjectProgress`、`ProgressEvidence`、`InferProgressCommand`。进度是模型推断观察，只读获准固定 TASK/STATUS 文件并附来源/版本/新鲜度依据；`inferred=true`，非人类批准、非项目状态、非授权。
- **迁移 013**：`migrations/013_project_progress.sql` 新建 `local_project_progress` 表（project_id 主键、data、revision>=1），`foundation_schema` workspace context 5→6。仅新增，不重建既有表；`_migrations` 幂等记录。
- **HTTP 路由**：`internal/adapters/httpapi/paper.go` 注册 `POST /api/v1/paper/search`，`progress.go` 注册 `GET /api/v1/local-projects/{id}/progress` 与 `POST /api/v1/local-projects/{id}/progress/infer`。领域服务未组装时明确返回 501 `unsupported_capability`，而非空成功。
- **OpenAPI 与生成客户端**：`contracts/openapi.yaml` 新增 `/paper/search`、`/local-projects/{id}/progress`、`/local-projects/{id}/progress/infer` 及 `PaperSearchRequestV1/ResultV1/MetadataV1`、`InferProgressRequestV1`、`ProjectProgressV1/ResultV1`、`ProgressEvidenceV1` schema，均 `additionalProperties:false`、`x-s1-supported:true`、`HumanSession` 安全；`pnpm generate` 重新生成 `web/src/api/schema.d.ts`。
- 未引入新 SDK、依赖或锁；全部复用稳定 deps。

## 验证

| 命令 | 结果 |
|---|---|
| `pnpm check` | **PASS / exit0**：Go fmt/vet/test/mod、依赖与架构、进程停止、redocly lint、示例校验、生成客户端一致、14 项真实 HTTP contract_local 场、TS/lint、61 前端单测、diff；新迁移在测试服务进程实际 applied |
| `pnpm build` | **PASS / exit0**：Go 可执行程序与 TS/Vite 生产构建 |

契约与迁移只建立运输/存储骨架，不替代 W1 真实多站点检索→正文、W2 真实进度推断与 OpenCode 原生接入的领域验收。旧 joint 成功与 UNKNOWN 未重发，未启用 paid/media opt-in，未读个人 Chrome。

## 给领域 owner 的最小接口/迁移交接（Handoff）

- **W1**：实现 `PaperSearchService` 于 `internal/attention/app`（除 contracts.go）与 `internal/adapters/importers`，沿既有 `ImportMaterial` 去重/revision；插件同源人工 review 与 CSRF 或受管批量授权复用现有会话身份，协商真正可用的浏览器 handoff（非仅 DOM 资产），不 wildcard CORS、不装 daemon。
- **W2**：实现 `ProgressService` 于 `internal/workspace/app`（除 contracts.go）与 `internal/adapters/agents`，读写 `local_project_progress` 表；进度只读获准 TASK/STATUS 并附依据，OpenCode 原生协议沿用既有项目身份/接续/ContextPacket 的同一 Service。记忆接口待用户答复可见范围后，再经本轨新增 014+ 迁移。
- 任何新增迁移序号（014+）、OpenAPI 变更、依赖锁更新均回归 W0 单写。

## 剩余限制与未执行操作

- CLI 完整覆盖范围（三客户端还是全部已安装）与记忆隔离/全局/仅规划两项仍 PENDING，未代选，未据此冻结或缩验收。
- 统一集成未进行：等待 W1/W2/W3 精确 SOURCE/REPORT 后普通合并，领域缺陷退原 owner，只修少量路由/导入/配置胶水。
- 未标全 CLI/大部分付费全文已支持；本轮不进行论文全站搜索、插件安装或真实进度推断验收。
