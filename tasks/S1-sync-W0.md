# S1 同步 / 本地 Agent：W0 契约与集成

2026-10-10；工作树 `s1-sync-contract-1010`，分支 `s1-sync-contract-1010`，基线 `2043770`（已 push）。客户端为当前原生 Codex，实际模型由总控 worker-read projection 独立核实为 `gpt-6.1-sol`；本终端不读取私有会话或凭据来核实。写域沿主计划；W1/W2/W3 的领域实现只通过普通 exact `--no-ff` 合并接入，缺陷退原 owner。

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

## 第三阶段：真实 CLI / HTTP 组装

W2 `763e183df35951a3ba919f593a51168d22fee522` 普通 exact merge `73af8a6`；W0 只组装其 `agents.NewInventory` / `RefreshCLI` / `NewLocalAgentService`。启动在既有信号 context 下仅探测 version/help，总限额 10 秒、每条 5 秒；超时保留部分缓存，不把安装当 native readiness，GET 无探测。信号在启动阶段也可取消探测，发现服务本身不改变资料权限或开启 native 控制。

新增可复用 `node scripts/check-local-agents.mjs`，沿现有 S1 helper 启动独立临时库，不访问个人库，不运行浏览器/模型/媒体。真实 API 按 OpenAPI 校验：Foundation S1 imports=true、其他执行标记 false；10 个稳定 CLI identity，8 个 version/help 成功（Codex 0.162.0、Claude 2.1.238、OpenCode 1.18.35、Pi 1.0.1、Grok 1.0.34、Kimi 2.1.1、Qwen 0.24.6、Cursor 3.23.23），Gemini/Cursor Agent 在当前 PATH 未找到，不推断全机未安装。配置/可启动/八项原生全部 unknown。

该命令 PASS/exit0：4 次 GET 完整缓存一致；无会话/Agent bearer/外部 Origin 403；人类 CSRF 原生 resume 501；实际停止/新进程重启后重建会话、CLI identities 与版本可查询；materials/jobs/distillations/outbox 仍 0。启动探测实际约 4.02 秒与重启后 4.63 秒。临时服务与数据由既有 helper 清理；这些是 installed CLI/interface 证据，不是原生接续、项目连接或 S1 完整验收。

`go test -mod=readonly ./cmd/server ./internal/adapters/httpapi ./internal/workspace/... ./internal/adapters/agents` PASS（live CLI 专用 opt-in 未在单测中开启；真实入口已由 smoke 实测）；本阶段 `pnpm build` PASS。

## 待决定契约草案（未冻结）

已通过 Orca ask 向总控提交：来源候选 identity 可考虑 provider + source_kind + external_id；列表项保留来源 item identity/metadata revision，未人工选取前不变成 Material；选取复用现有 imports/jobs，不自动创建 Mission。公开/登录绑定、推荐与正文读取先后、刷新策略、明确 Agent 读取根和原生操作范围均等待用户决定；本轮不发布这些输入 schema、不新增授权或迁移。

## 第四阶段：安全阶段普通集成

已按 owner 正式 Handoff ordinary exact `--no-ff` 合并，无冲突、无领域重写：

| Owner source | W0 merge | 范围 |
|---|---|---|
| W2 `47772035b39ee6ce445e6fe90bd6120b0c87e200` | `1d4e711` | 部分探测状态、取消测试、报告，沿既有未知边界 |
| W3 `b49c1f92035ee42a875d446d3852d5d3dd44e5aa` | `9163412` | 真实发现 API 面板、项目概览、待决定账号提示与针对性浏览器用例 |
| W1 `d0e717dd404d9ef917a31985461fc59c33e92756` | `2f70282` | 公开 B站列表适配与抖音页面解析、稳定 identity/变更比较；未组装账号同步或选择入库接口 |

本阶段业务/检查源 `2f70282ae414105431832102821cefe26f677d07`：`pnpm check` PASS/exit0（19 包边界、232 示例、API contract_local 14/14、TS/lint、前端 44/44，既有 EventV1 警告保留）；`pnpm build` PASS/exit0。检查使用临时库并默认关闭模型/媒体，首次完整 S1 验收仍未成立。W3 由总控授予目标浏览器 slot；W0 不运行浏览器，待接收其最后报告/返修结果。

## 第五阶段：保留首败，修复开发入口启动竞态

W3 正式 Handoff 报告首次 UI `04a67c5` 目标集合 120 PASS/2 matcher FAIL，修正 matcher 后 `e2aa6dc` 六项 5 PASS/1 startup FAIL：第一个 1920 空页请求发生在 API CLI 启动探测约 4.3 秒内，Vite 已可访问但 API 未监听，代理 `ECONNREFUSED`。其后五项包括真实库存两尺寸通过，API 显示 10 identities/8 安装、配置/可启动/原生全部未知；这些晚期 PASS 不覆盖原首败。完整原件位置见 W3 报告。

W0 修正拥有的 `scripts/dev.mjs`：先启动 API，在既有 owned-child/关闭机制下以总计 30 秒 health readiness 检查等待，再启动 Vite。API 启动失败退出或关闭信号停止等待和网页启动，按已有 helper 清理；不修改用例等待、不添加业务重试/调度框架。W3 最后 UI/report `3ec0d8331da08cb1e1127efb787768fd72cccb6c` 已普通 exact merge `eaa1829`（此前 e2aa6dc 单独 merge 只引入中文提示与实际 API 比较）。

`node --check scripts/dev.mjs`、`git diff --check`、合入最终 UI 后 `pnpm --dir web exec tsc -b` PASS。实际故障分支：设置 `ASTROCYTE_PORT=invalid`、模型/媒体 false，运行 `node scripts/dev.mjs --ephemeral`，API 明确拒绝非法端口且 dev **预期 exit1**，没有启动 Vite；临时目录 `astrocyte-dev-GMdTz1` 自动清理后 `Test-Path` 为 False。正向首载与浏览器修复验收待总控调度 W3 原失败项，本轮不把语法/负例当成正向已验。

## 剩余 / 未执行

共同工作区发现 DTO/HTTP/adapter/service/启动组装通过真实 CLI/API 检查，四轨安全阶段已普通合并；首载启动修复正向浏览器与总控独立最终检查仍待。账号同步、原生控制/项目接入及完整 S1 未完成。真实 AT01–04、应用模型、媒体和浏览器需总控单次调度；没有合入 main，没有本轮远端 CI 结果。未答权限/认证/刷新保持待定。W1/W2 原任务已由总控按完整范围未完成结算，安全成果保留；W0 当前继续等待最后返修验收与总控收口指令。
