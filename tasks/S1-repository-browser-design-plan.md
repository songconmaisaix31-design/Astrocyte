# S1 仓库、浏览器收藏与交互重设计

基线：`c3bdb42128081ad3d97b758cd669203880792806`。本轮用户确认：GitHub 先同步仓库信息，人工放进最顶层蜂群开发空间才克隆代码；安装浏览器插件接入已登录抖音的选定收藏夹。原默认项目权限 A、按动作授权、模型许可独立、固定程序同步和去重保持。首批 GitHub 链接/私有范围待答，先交付公开手动入口，不扫描私人仓库或凭据。

前端按用户新要求整体重设计，保持 Go/React/SQLite/OpenAPI 底座和真实 API：增加留白、立体层次与容器高度，将资料、来源和项目操作分成明确工作流；高级权限收进小设置，平台和 Agent 使用可核实的官方图标，不虚构连接和能力。此次视觉写域覆盖三页，业务仍按纵向切片实现，不增加 Mission 执行框架。

| 轨 | 固定工作树/分支 | 独占 write_paths | 交付 |
|---|---|---|---|
| W0 契约与集成 | s1-sync-contract-1010 | contracts/；internal/{attention,workspace}/app/contracts.go；internal/httpapi/；internal/foundation/；cmd/；migrations/；web/src/api/；依赖与锁；scripts/；tests/s1/；tasks/S1-sync-W0.md | 先发布必要端口、唯一迁移序列/入口，最后普通合并三轨，仅补组装胶水 |
| W1 收藏同步 | s1-sync-attention-1010 | internal/attention/{domain,app}/（除 contracts.go）；internal/adapters/{importers,distillers,objects}/；internal/adapters/sqlite/（除 local_agents*.go）；tasks/S1-sync-W1.md | 复用 OpenCLI 已登录浏览器桥，在明确选定收藏夹中读取元数据，持久化、重复同步与人工选取；保留 summarize 正文链路 |
| W2 仓库与空间 | s1-local-agents-1010 | internal/workspace/{domain,app}/（除 contracts.go）；internal/adapters/agents/；internal/adapters/sqlite/local_agents*.go；docs/probes/local-agents.md；tasks/S1-sync-W2.md | 公开 GitHub 信息同步；人工纳入蜂群空间才固定程序克隆，持久化实际根/HEAD/状态，权限不自动扩大 |
| W3 交互与官方图标 | s1-sync-ui-1010 | web/src/（除 api/）；web/public/；web/e2e/；docs/acceptance/S1-sync.md；tasks/S1-sync-W3.md | 整体交互重设计、官方资产、真实 API 入口；1280/1920/窄屏浏览器验收 |

主控只维护计划、决定、状态、Git 与独立验收；四轨复用原 worktree/branch，各自先普通合入此基线，完成 commit + push。实际客户端/模型由运行回执核实，不指定未经用户指定的模型。浏览器插件安装由主控操作；实际个人浏览器仅一个所有者，W1 得到连接交接后读取指定抖音页，不读取私信、观看历史、无关标签或 Cookie 原文。

验收：真实 API → SQLite/对象存储 → 重启可查询；GitHub 信息同步没有代码目录，纳入蜂群空间有真实 clone/HEAD，重复纳入不重复克隆，不启动 Agent 或创建 Mission；收藏元数据同步不调用模型，推荐后用户选择才入库；原 AT01–04 保持回归，旧未知模型作业不重放。核心适用测试/check/build 和真实浏览器通过后才接受功能；原首个失败保留，新修复结果单独记录。原论文 DNS 例外、第二自动整理 CLI 与输入容量问题仍未答，不据插件许可扩大。
