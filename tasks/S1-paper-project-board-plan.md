# S1 扩容、论文与项目总览

基线 `b8308e9bdb06b059079961cbd1abd1ad9e937064`，主控分支 `s1/attention-materials-20261009`。本轮用户要求：扩容此前超128KiB的论文/视频联合输入、优化人类交互、开发论文Agent Search及基于summarize的独立插件、按提供截图自动汇总各Agent项目并提供统计、进度、管理、备注、复盘与归档。沿Go/React/SQLite/OpenAPI和原四工作树，不另建调度框架；主控只计划、决定、验收和最终提交。

容量按此前待答方案提升联合模型输入至512KiB UTF-8（包含指令/引用），拒绝超限，不截断；模型输出、原生历史和权限上限分别保持现有值，不因容量授权扩大目录或外发。重要选择已问：论文搜索/当前页提取/两者、插件逐次或站点授权、人工进度或有依据的Agent推断；未答项不冻结接口或代选。当前能独立推进容量、项目识别和人工备注/复盘/归档，以及论文站点与上游研究。

| 轨 | 固定 worktree/branch | 独占 write_paths | 目标与验收 |
|---|---|---|---|
| W0 契约与最终集成 | s1-sync-contract-1010 | contracts/；internal/{attention,workspace}/app/contracts.go；internal/adapters/httpapi/；internal/foundation/；cmd/；migrations/012+；web/src/api/；根依赖/所有锁；scripts/；tests/s1/；README；tasks/S1-paper-W0.md | 先给其他轨可消费的必要契约/迁移/安全身份；未答问题保持未决定；最后普通合并三轨，仅组装胶水；真实API权限、持久化/冷启动、契约漂移与check/build |
| W1 论文与独立插件 | s1-sync-attention-1010 | internal/attention/{domain,app}/除contracts.go；internal/adapters/{importers,distillers,objects}/；internal/adapters/sqlite/除local_agents*.go；extensions/paper/除依赖锁；docs/acceptance/S1-paper.md；tasks/S1-paper-W1.md | summarize上游代码/许可证与纸页提取复用；大部分论文站点的元数据/正文分开、DOI/arXiv去重与明确受限；独立插件接应用，不另存Provider凭据；512KiB指令输入边界；与W0协调搜索/插件权限后再冻结相关接口；真实多站点和安装加载/应用入库检查 |
| W2 项目汇总与原生容量 | s1-local-agents-1010 | internal/workspace/{domain,app}/除contracts.go；internal/adapters/agents/；internal/adapters/sqlite/local_agents*.go；docs/probes/local-agents.md；tasks/S1-paper-W2.md | 原生输入512KiB；复用Orca登记范围和明确CLI项目索引/会话路径元数据，自动聚合同项目多目录/多客户端；稳定来源、真实活动/提交、人工备注/复盘/归档，刷新不丢人工字段；人类身份写入，缓存GET不扫盘；无整盘/凭据/无关对话读取，无自动Agent控制；持久化/重启/边界验收 |
| W3 人类交互与看板 | s1-sync-ui-1010 | web/src/除api/；web/public/；web/e2e/；docs/acceptance/S1-paper-board.md；tasks/S1-paper-W3.md | 项目总览成为工作区主入口：统计、平铺/平台/分组/时间线、筛选/排序、项目详情备注/复盘/归档；进度与活动区分、未知明确；搜索论文→看来源→选提取/入库→继续整理按用户选择接真实API；现有入口/真实空错误/权限可达；1920/1280/390键盘与实际浏览器 |

各Worker先普通合入基线，完成和返修持续由原owner做，commit+push精确SOURCE/REPORT。W0单一拥有共享契约、迁移序号、锁和程序入口；跨写域只Handoff。研究上游以公开官方文档及代码为准，保留来源和许可证；独立插件先完成可安装构建和真实隔离浏览器验收，个人浏览器安装由主控最终操作，避免与原OpenCLI范围冲突。个人Chrome当前仍只批准求职夹，不据本轮论文功能读取所有个人标签或消息。

21:35 四轨实际开始，Run `run_dcc0b937ae90`，实际Orca Codex0.162.0 / GPT-6.1-Sol high fast。W0 `ctx_b162c766fc84`、W1 `ctx_aefe41ec20f2`、W2 `ctx_4434022e51cd`、W3 `ctx_d742848ca3e0`。首W0/W1/W3就绪超时未运行任务的失败保留；同终端识别后用原Task重派，清理责任转移新Dispatch。四轨输入初留composer，读取确认后各仅一次Enter，已观察实际工作，没有重复创建终端/编辑者。W2先交项目汇总/人工字段端口提议给W0；论文访问/进度问题仍未答，按范围继续独立开发。

验收终点：512KiB以上/以下边界、真实论文/视频联合完整输入且无截断；多论文站点搜索/选取/全文提取/来源版本/重复导入行为；插件真实加载、项目许可与应用持久化入库；真实本机项目去重聚合、人工备注/复盘/归档、刷新和API服务进程重启保留、统计来源明确；真实API加浏览器交互和适用check/build。重启指停止受控API进程后以同一存储启动新的API进程并重新查询，不要求重启Windows或用户电脑。付费/媒体只执行本轮新增必要行为，旧成功和UNKNOWN不重发；fixture不能替代实际验收，原失败保留，不制造完整率或后台权限。
