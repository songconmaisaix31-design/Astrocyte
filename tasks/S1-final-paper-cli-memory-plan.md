# S1 论文、CLI、进度与记忆收尾

2026-10-11 恢复：Orca runtime 已变为 `90510c9d-1184-4e26-8291-3b372b161199`，root 新终端 `term_ab71d476-33a6-42ec-b5dd-5bb9ac628055` 已重新绑定原 Run，不建立平行 Run。原 W1 SOURCE `11791ce` 已发布（论文检索/HTML/可构建 MV3），原 W2 SOURCE `90c18d4` / REPORT `f1ace50` 仅部分成果（进度域与工具规划），都不表示整体完成。W3 重试于 23:25 异常退出 code1073807364，已提交 UI SOURCE `add30be` / REPORT `97090d5`，另有插件复核未提交改动保留；W0 runtime恢复后 abandoned/terminal_missing，未提交契约与迁移013保留。恢复仍按原四写域：W0 完成真实接口与最终普通集成；W1 补公共 PDF 提取和实际插件验证；W2 修进度错误/结果恢复并验证 OpenCode 子进程配置隔离与原生回路；W3 对齐实际 DTO 完成插件复核、真实搜索/批选/进度浏览器验收。CLI全覆盖、记忆可见范围仍待用户，不据“继续”代选。已安装/源码接口/本地测试不等于原生接入与真实端到端。

23:03 W3 首次 OpenCode 在读取计划后退出，无业务改动，未结算完成。执行主机核验其专属 pwsh 59940 启动目录为 UI 工作树且没有子 OpenCode，原终端只读 `$PID` 返回 59940 和 PowerShell 提示符；依据这一实际退出事实停止旧 Dispatch `ctx_15e81ca1f271`，同 Task/Worktree/Branch 用 OpenCode 重试为 `ctx_cc2d02cac1ac`，新终端 `term_488b1577-e950-4f29-a5d0-948262c849eb`。旧失败保留；没有并行第二编辑者，其他三轨未停止。新 Worker 已实际读取任务，TUI 为 DeepSeek V4 Pro。

22:42 四轨已在原工作树启动新Orca Run `run_8c1696bb815a`：W0 Task task_3097e734e4d0 / Dispatch ctx_5cc390be9fb4；W1 task_ff677991837d / ctx_b97954c9d788；W2 task_11167b05140a / ctx_fda75109968c；W3 task_88f06e405600 / ctx_15e81ca1f271。launch.effective.agent全部opencode，模型参数未覆盖；已从实际TUI观察DeepSeek V4 Pro/DeepSeek，turn-start自动协议unsupported但terminal显示活动，不据缺少协议证明重发输入。原本机个人preview73199保持，CLI完整覆盖和记忆范围两项新问题待答。

基线 `d9acb43d4d617034f0865f7c8b7ecea377162058`，主控分支 `s1/attention-materials-20261009`。用户确认论文搜索＋正文提取、项目授权范围内批量读取、有依据的Agent进度推断、新Worker全部OpenCode、完成S1；新增蜂群工具层规划与Claude-mem式长期记忆。复用Go/React/SQLite/OpenAPI、summarize0.25.1、四原工作树和权限入口，主控不写业务代码。CLI覆盖验收与记忆存储/可见范围已问待答；不因此停下独立论文/插件、OpenCode协议、进度与工具规划。

| 轨 | 固定worktree/branch | 独占write_paths | 目标与验收 |
|---|---|---|---|
| W0 契约/HTTP/集成 | s1-sync-contract-1010 | contracts/；internal/{attention,workspace}/app/contracts.go；internal/adapters/httpapi/；internal/foundation/；cmd/；migrations/013+；web/src/api/；根依赖/所有锁；scripts/；tests/s1/；README；tasks/S1-final-W0.md | 先与领域owner形成最小可消费契约与存储迁移，论文检索/选择提取/插件审阅与批量真实同源身份、CLI/进度/记忆接口；不扩大CORS、不装未认证daemon、不代选记忆策略；最后普通合并三轨，适用check/build/真实API及S1验收 |
| W1 论文/独立插件 | s1-sync-attention-1010 | internal/attention/{domain,app}/除contracts.go；internal/adapters/{importers,distillers,objects}/；internal/adapters/sqlite/除local_agents*.go；extensions/paper/除锁；docs/acceptance/S1-paper.md；tasks/S1-final-W1.md | 公开学术检索→来源版本→选择正文→既有ImportMaterial/job去重与重试；通用论文元数据/HTML/PDF提取复用summarize，受限/paywall明确，无绕过；独立可安装MV3，当前页点击及显式批量选项，不全站自动读取；真实多站点与隔离浏览器插件→应用持久化验收 |
| W2 CLI/进度/记忆 | s1-local-agents-1010 | internal/workspace/{domain,app}/除contracts.go；internal/adapters/agents/；internal/adapters/sqlite/local_agents*.go；docs/probes/local-agents.md；docs/architecture/swarm-tools-memory.md；tasks/S1-final-W2.md | OpenCode真实正式协议适配，安装/配置/可启动/八能力分别记录；已批准项目启动/读取/发送/接续/停止/reconcile及真实跨会话交接，旧会话未确认停则禁止双写；进度只读取获准固定TASK/STATUS文件，模型推断附来源/版本/新鲜度，非批准；研究Claude-mem并规划工具层，记忆权限待答后再接口与落地；不扫凭据/私密历史/整盘、不覆盖用户CLI配置 |
| W3 前端/插件流程验收 | s1-sync-ui-1010 | web/src/除api；web/public/；web/e2e/；docs/acceptance/S1-paper-board.md；tasks/S1-final-W3.md | 延续已验布局：论文搜索→勾选→正文/受限状态→批量入库→继续沉淀；CLI项目范围/原生操作/交接可见，进度依据与记忆访问分层；真实API三尺寸/键盘/空错误/权限/SQLite刷新，不虚构进度/安装状态 |

每轨新Task/Dispatch、固定原Worktree/Branch、Agent为OpenCode，模型沿其现有配置并观察实际值，不能声称继承Codex模型。跨域提议仅Handoff，OpenAPI/迁移序号/依赖锁/入口W0唯一owner。Worker持续开发测试文档返修，完成commit+push；W0作为最终集成Agent只胶水，领域错误退原owner，root验收及最终提交push。未答关键选项不冻结接口、不缩小验收冒充完成；不造调度器/完成证明系统。

验收：S1 AT01真实论文/视频版本与多轮无Mission；AT02重复/内容更新/摘要重试无重复工作；AT03 Agent读不增人类关注；AT04延期不丢资料且非拒绝。既有成功媒体/模型和UNKNOWN不重发，用既有实际结果＋新增链路必要运行。新增论文多站点搜索/HTML/PDF/插件加载与真实持久化、批量选取不全量、撤销边界；CLI原生同ID接续和跨会话交接分别真实观察，关键权限/停止/未知恢复；进度与记忆新API进程后可查且项目隔离。完成适用Go/前端/OpenAPI/check/build及浏览器检查；首次失败保留，最终只报告实际覆盖和剩余限制。个人5173/8787在最终通过前保持基线，测试服务不能盲停个人/队友进程；最终迁移前一致SQLite备份，不复制验收库进个人库。
