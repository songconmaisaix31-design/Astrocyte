# S1 同步 / 本地 Agent：W0 契约与集成

2026-10-10；工作树 `s1-sync-contract-1010`，分支 `s1-sync-contract-1010`，基线 `2043770`（已 push）。客户端为当前原生 Codex，实际模型尚未从本终端观测；不读取私有会话或凭据来核实。写域沿主计划；W1/W2/W3 的领域实现只通过普通 exact `--no-ff` 合并接入，缺陷退原 owner。

## 已有接口盘点

- S1 已组装资料列表/详情、各版正文与附件、人工整理、自动整理作业与处理器状态、作业列表/详情/重试/取消、候选、域、项目空间与固定版本引用、排序配置。
- `/projects`、`/sessions`、`/proposals` 仍为空读；`/sessions/{id}/resume` 和 `/handoff` 为 501。项目空间引用不授予 Agent 访问权。
- 现有人类会话 + CSRF + loopback/Origin 保护可复用；Agent bearer 仅识别身份，待定 scope 的受保护读取仍拒绝。不能从 token、CLI 安装或热度推导授权。
- 迁移已有 001–005；SQLite/objects 位于仓库外，本轮检查只用临时数据目录，未读取/修改个人库、宿主 DNS 或代理。

## 第一阶段：基线与契约漂移修复

发现 `FoundationV1.capabilities.imports` 固定 false，与既有组装 S1 返回 true 不一致。移除错误固定值，补实际 S1 结构示例并重新生成 TS；其他执行/批准/接续标记仍 false。`stage=S1` 只表示已组装切片，不代表 AT01–04 完成。

| 实际命令 | 结果 |
|---|---|
| `pnpm generate` | PASS；安装现有锁定依赖并生成 `schema.d.ts`，summarize 仍为 0.25.1 |
| `pnpm check` | PASS，exit0；Go 格式/vet/test/mod、18 包边界、227 OpenAPI 示例/生成漂移、contract_local API 14/14、TS/lint、前端 40/40；既有 EventV1 unused 警告保留 |
| `pnpm build` | PASS，exit0；Go 程序与生产网页 |

这些是 contract_local/离线检查，未调用应用模型、媒体转写或浏览器验收；旧 UNKNOWN 未重发。

## 第二阶段：公共发现契约（W2 已对齐）

`GET /local-agents` 使用既有人类会话，只读缓存；GET 不执行 CLI。W2 adapter 在受控启动阶段执行有时限的 PATH version/help 发现，配置/可启动未测保持 unknown。安装/配置/可启动观察与八项原生能力分别报告，原生 enum 沿 SPEC §8.1 `supported/unsupported/unknown`。不暴露本地路径、凭据、私有会话或项目内容。W0 发布 OpenAPI、生成客户端和 app port；W2 所有者实现领域/adapter/service。

W2 domain commit `1fe4269` 已通过普通 `--no-ff` 合入，merge `6a2c791`；W0 HTTP 仅读取 service 缓存，缺失服务明确 501，不伪装空库。仅 LocalAgents 组装也启用原会话保护；无会话、有效/无效 Agent bearer、外部 Origin 均在调用 service 前拒绝，安装 available 不把 configured/startable/八项原生 unknown 改成支持。非结构化内部错误脱敏。

验证：`go test -mod=readonly ./internal/adapters/httpapi ./internal/workspace/app` PASS；`pnpm check:contracts` PASS（232 示例、生成漂移，原 EventV1 警告）；`pnpm --dir web exec tsc -b` PASS。这些是 transport/contract_local 检查，真实 adapter 及 server 组装尚待 W2 后续。

## 待决定契约草案（未冻结）

已通过 Orca ask 向总控提交：来源候选 identity 可考虑 provider + source_kind + external_id；列表项保留来源 item identity/metadata revision，未人工选取前不变成 Material；选取复用现有 imports/jobs，不自动创建 Mission。公开/登录绑定、推荐与正文读取先后、刷新策略、明确 Agent 读取根和原生操作范围均等待用户决定；本轮不发布这些输入 schema、不新增授权或迁移。

## 剩余 / 未执行

共同工作区发现 DTO/HTTP 已实现，真实 adapter/service/启动组装尚待 W2；账号同步和四轨最终集成未完成。真实 AT01–04、应用模型、媒体和浏览器需总控单次调度；没有合入 main，没有本轮远端 CI 结果。总控已明确本 Dispatch 持续等待决定/集成，当前不发送 worker_done。
