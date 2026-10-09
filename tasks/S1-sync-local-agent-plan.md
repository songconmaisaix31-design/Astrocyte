# S1 账号同步与本地 Agent 接入执行计划

2026-10-10；基线 `c46c9ad2ec42390965b9f687e0d26627e8de5f63`，主控分支 `s1/attention-materials-20261009`。本轮用户要求多 Agent 并行：先实现 B站、抖音账号追踪同步，其他主流平台占位；参考所附“AI 作品寻回册”实现本地 Agent 接入，并完成 S1。沿用 Go/React/TypeScript/SQLite、现有 summarize 0.25.1 与本地媒体依赖，Python 已获允许。不引入开发调度框架，不重做三页，不自动创建 Mission。

| 轨道 | 固定工作树/分支 | 独占 write_paths | 交付与验收 |
|---|---|---|---|
| W0 契约、HTTP、组装、集成 | `s1-sync-contract-1010` | `contracts/`、`internal/attention/app/contracts.go`、`internal/workspace/app/contracts.go`、`internal/adapters/httpapi/`、`internal/foundation/`、`cmd/`、`migrations/`、`web/src/api/`、根依赖/锁、`scripts/`、`README.md`、`tasks/S1-sync-W0.md` | 公共接口/迁移/依赖/入口唯一所有者；先发布已确定契约。其余轨完成后由此 Agent 普通合并，只补少量胶水，领域返修退原轨。契约、HTTP、构建检查。 |
| W1 账号同步与 Attention 后端 | `s1-sync-attention-1010` | `internal/attention/domain/`、`internal/attention/app/`（排除 `contracts.go`）、`internal/adapters/importers/`、`internal/adapters/distillers/`、`internal/adapters/objects/`、`internal/adapters/sqlite/`（排除 `local_agents*.go`）、`tasks/S1-sync-W1.md` | B站/抖音真实列表适配、账号/收藏夹更新去重与选择入库、取消/错误/重启；复用 summarize 单视频处理。补齐 S1 缺陷并保持 UNKNOWN 不自动重发。读取主流平台正式文档或上游源码，不凭旧 API 印象实现。 |
| W2 本地 Agent 与项目适配 | `s1-local-agents-1010` | `internal/workspace/domain/`、`internal/workspace/app/`（排除 `contracts.go`）、`internal/adapters/agents/`、`internal/adapters/sqlite/local_agents*.go`、`docs/probes/local-agents.md`、`tasks/S1-sync-W2.md` | 重新核实本机 CLI；安装/配置/可启动/八项原生能力分别呈现。按用户确定范围接入指定项目与活动、原生操作；不能把读取日志当作原生接续，不能扫描全盘或读取凭据。不接 Swarm/Mission。 |
| W3 前端及浏览器验收 | `s1-sync-ui-1010` | `web/src/`（排除 `api/`、`pages/swarm/`）、`web/e2e/`、`docs/acceptance/S1-sync.md`、`tasks/S1-sync-W3.md` | Attention 账号绑定/同步清单/人工勾选/作业反馈；共同工作区局部加入参考图的真实平台/分组/活动/项目卡。真实 API 默认；占位不伪装已连接。覆盖 S1 AT01–04 与账号同步，核对 API、SQLite/对象、浏览器及重启。 |

每轨一个本机 Orca Codex Agent、一个隔离 worktree、一条分支；同一 Worker 持续实现、测试、返修、commit + push。默认继承本机客户端模型配置，启动回执记录实际值。W0 发布契约后各轨更新；跨轨只发 Handoff。主控只维护计划/决定/状态，最终独立 `pnpm check`、`pnpm build`、`pnpm test:e2e --workers=1` 并核验真实四项验收后提交 push。全机同时只跑一个真实媒体/模型/浏览器验收，避免重复下载和资源争抢。

待用户决定：公开或浏览器登录绑定方式及实际账号/收藏夹；同步清单先建议后勾选的入库顺序；本地接入是否包含启动/接续/停止/观察；Agent 只读主动引用文件或明确项目目录。未答项不冻结接口、选择组件、降级验收或开放目录。先并行开发已授权公共 CLI 发现、公开列表可行性、现有 S1 缺口核查与界面公共组件。原单视频成功不代表账号同步或 S1 已通过。

验收终点：AT01 真实论文/视频多轮沉淀有来源版本且无 Mission；AT02 重复导入与摘要重试无重复工作、变化产生新版本；AT03 授权 Agent 多次读取旧资料不增加人类关注；AT04 以后再做保留资料且不是拒绝样本。账号同步必须真实来源列表、变化检测、人工选取、持久化和重启可查。本地 Agent 能力逐项记录真实结果；未执行/未知保持明确。旧 Codex UNKNOWN 不重发；私有外发、整盘扫描和自动执行不能由“全面接入”推导。

状态：四条 Orca 开发轨已启动且均实际 `turn_started`，Run `run_dc19393eafd5`；原开发客户端模型从 Worker projection 确认为 `gpt-6.1-sol`。W0 `ctx_76f4c72a4454`，W1 `ctx_3dec09df71bd`，W2 `ctx_661d633eeaf6`，W3 `ctx_491aa39c4018`。工作树均以含本计划的 `2043770` 为基线，分支名与表中工作树名一致。关键问题仍待用户；W0/W2 先协调不依赖私有读取或执行授权的真实 CLI 清点端口。主控不写业务代码。

03:26 主控验收：`bbf99ac7ce823a73170b519b014a4bea9984651e` 的业务代码与 `23028df474d0a9d5f53bce4e2386c361026c2d98` 一致；check、build PASS，完整 E2E 152 PASS / 4 SKIP（4.1分钟、exit0），测试服务监听已退出。原验收失败保留在各轨报告。W1/W2/W3 已结算为原任务未完成并释放；公共部分通过不等于原账号同步、原生操作与完整 S1 完成。

临时写域移交：W3 已释放，测试已结束。仅将 `web/src/pages/attention/AccountsPanel.tsx` 的平台名称常量数组移交 W0 集成 Agent，补齐其他主流平台的“待接入”标签；不改行为、领域、接口或测试。W0 完成后 typecheck/build、push；主控在实际预览复核布局。完整 E2E 只覆盖上述补标签前的源，最终名称差异单独记录。
