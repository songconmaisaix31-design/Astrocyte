# S2/S3 产品页面开发记录

当前任务 `task_d0cee5a583bb` / Dispatch `ctx_e6aafcccc10a`，模块负责人任务 `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa`，子 Run `run_dbf1a68bc692`。工作树 `product-ui-20261015`，分支 `songconmaisaix31-design/product-ui-20261015`。实际客户端 Orca Codex，模型 `gpt-6.1-sol`。接管基线 `6f51df100299274f01740fcac83c8f2cf91d0751`，工作区 clean。

只修改 `web/src/pages/workspace/proposals/**`、`web/src/pages/swarm/SwarmPage.tsx` 与其 CSS、`web/src/pages/swarm/missions/**`、`web/e2e/s2-*.spec.ts`、`web/e2e/s3-*.spec.ts` 和本文件。Workspace 入口、公共客户端、生成类型、查询 Hook、认证、共享组件与真实服务 helper 归 F；OpenAPI 与服务器装配归 H；S2/S3 领域与存储分别归对应开发人员。

## 首阶段源码

SOURCE `27e9fd86f4a4d4f7ec29013a14670914b9f82c7a` 已普通 push，`git ls-remote` 确认远端同 SHA。

`ProposalWorkspace` 使用现有真实 `listProposals` 和 `useReadApi`，覆盖 loading、empty、error、stale 与搜索。稳定入口为 `ProposalWorkspace({ query, onClearQuery, refreshToken, onChanged })`，F 负责在 Workspace 提案位置接入。详情展示现有生成 `ProposalV1` 的全部项目、根目录、动作、外发规则、调用及 Token 上限、费用未知、产出、停止条件、候选固定版本。`ProposalRevisionDiff` 接收同一提案的两个实际版本，逐字段显示执行内容变化。

本阶段是查询与显示组件，不构成 S2/S3 整功能通过。创建、人工批准、拒绝、撤销、Grant 状态、Context、Mission 时间线、工作项命令与 Artifact 验收等待准确公开契约及生产接口。已直接向 S2 `ctx_f32b6d92119a`、S3 `ctx_d4bf65ca06bf`、H `ctx_e1a8ebeeea83` 与 F `ctx_5bac511ada08` 交接；不手写未经确认的公共 DTO。

## 执行内容表单阶段

SOURCE `d6be4017a54797fddaf40f9d18f1acbee951c3eb` 已普通 push，远端 SHA 一致。`ProposalContentFields` 按 S2 精确 SOURCE `4eec31ddfde39fa3f548691ec6e7f2f34470351a` 的 `ProposalContent` 提供完整编辑字段：固定目标版本、项目、根目录、动作、工具、外发目标、并发限制、调用/Token/最小货币单位预算、币种、产出、成功条件、截止时间及停止条件和数量/尝试/无进展限制。表单字符串状态属于页面编辑内容，生产命令仍等待公共生成类型与 F 客户端绑定。该阶段 `pnpm --dir web typecheck` 和 `pnpm --dir web lint` PASS。

已交 S2 补充批准记录的持久查询关联：现有提案读取未提供 Grant ID，重载后需通过真实关联或过滤列表取得有效 Grant 和撤销入口。Goal 创建后可使用实际返回的固定版本，列表需求也已交 S2。未知命令锁的公共复用入口交 F；现有 `useCommand` 不提供跨重开持久锁。

## 实际检查与首次失败

| 命令 | 实际结果 |
|---|---|
| `pnpm install --frozen-lockfile` | PASS，本树独立依赖安装 |
| `pnpm --dir web typecheck` | PASS |
| `pnpm --dir web lint` | PASS |
| `git diff --check` | PASS |
| `pnpm build` | PASS：Go 可执行文件与生产 Web 资源 |
| `pnpm check` | FAIL：`cmd/server` 的 `TestRelayPersistentRealHTTPReceiptsMutexAndRestart`，`relay_runtime_http_test.go:297` 实际 HTTP 403 `approval_required`，预期 200；另 8 项 importers 测试缺明确视频/PDF目录，9 项 Workbench 测试缺授权实际 publication 环境。日志保留于本机 `.verification/product-ui/check-stage1.log`，已交回原负责人并请求准确公开测试输入指针 |

未执行真实 S2/S3 交互、两尺寸验收、撤销后新动作拒绝、未知命令持久化与重启验收。Relay 实际 GUI 由 F 独占，本任务没有注册、确认、读取或停止其原生任务。原始输出与凭据不进入 Git 或 HTTP。
