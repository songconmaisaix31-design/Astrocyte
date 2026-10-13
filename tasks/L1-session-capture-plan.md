# Works Relay v1.3 · L1 会话采集增补与一页执行计划

事实源：[认知栈 L1–L4 原文](cognitive-stack-L1-L4.md) §2、§7；用户指定接入层主控终端 `term_d7f7e173-a64c-43fe-981a-dbe57124de87`。本会话主控 `term_23fa4b37-e887-4848-8d1e-c52dd88cab52` 负责增补计划、验收和最终 Git；原 Works Relay 的 R/H/F/I owner 保持，不接管其 Run 或重派任务。

## 基线与范围

- 原 Works Relay Run：`run_d577f0d3fee1`；目标 `workbench/v0.2-access`。可消费工程 SOURCE 为 `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`，原集成分支 `songconmaisaix31-design/works-relay-integration`；REPORT `2a4c58c20bb77f535fb2b3af2901af8d1a627aa4` 与其业务全等。Handoff：`msg_60ab2f9d21ed`。CI 38469944551 在 Handoff 时仍运行，未以旧工程检查替代最终验收。
- 文档基线：validation `98a13b43e76240436fb17d3246d02507bf959417`。原文登记与本计划、[L2 协议草案](L2-work-experiment-draft.md)在独立文档工作树完成；validation 只入 Markdown。
- 冻结期业务只在 workbench 线：L1 完整原生 Session 纳入接入层自有采集管道，输出统一派生 Event 到独立 workbench 库，并让 Today 可见。L2 实验和 L3/L4 实施不在本轮启动。
- “完整”指获准、接入层持有的会话中事件覆盖完整并可对账；不把未采集、截断、未知格式或失败记录排除后宣称完整。不展示原始思考、prompt、工具参数、stdout 或私人路径。

## 互斥轨道与接缝

| 轨 | 固定 owner 与候选 write_paths | 交付与依赖 |
|---|---|---|
| R：owned-session 接缝 | 原 R owner、原 `works-relay-runtime` 工作树 / `songconmaisaix31-design/works-relay-runtime`；既有 `internal/workspace/{domain,app}/relay*.go`、`internal/adapters/{sqlite,agents}/relay*.go` 与自身报告 | 只读 Service 端口暴露内部 owned task/session 绑定供采集器使用；不绕过唯一 store writer，不将私有 HistoryPath 变为 HTTP 字段 |
| H：Session 采集与消费组装 | 原 H owner、原 `works-relay-http` 工作树 / `songconmaisaix31-design/works-relay-http`；新增 `internal/adapters/events/relay_session*.go`，既有 `cmd/server/relay_assembly.go`、observer/config、workbench Reader / HTTP 写域与 `docs/access/http.md` | 完整行读取、会话/CWD 校验、派生投影、稳定事件身份、绑定到既有观察生命周期；复用独立应用库事件仓储，Today 读取接缝；不改 mapper / 引擎 / schema |
| F：真实体验验收 | 原 F owner、原 worktree/branch；其既有 Today / Relay UI 与新增 `relay*.spec.ts` 写域 | 注册→人工派发→实时事件→Today，三值显示与真实对账；只在实际必要时修改 UI |

原主控已明确采集与组装合为原 H 一轨，不另建 C owner；`relay_session*` 前缀不让出。R/H/F 每轨仍一 Agent、一工作树、一分支，跨域只交 Handoff；只有原主控为 retained owner 建立新的 Task/Dispatch 后才续作，不向已结算的 Dispatch 投递新开发。所有开发、测试、文档和返修留在同一 owner。唯一集成由原主控的集成 owner 普通 no-ff 合精确 SOURCE/REPORT；领域失败退原轨。本会话只写新三份任务 / 协议文档。

2026-10-14 开工回执：原 H 新 Task `task_f215e7452d70` / Dispatch `ctx_2d4b642fbac5`，复用终端 `term_b93b8ad2-cd2c-4f23-8981-1e0a13e0bd15`；`ready / input_accepted / turnStart observed`，实际模型 `gpt-6.1-sol`。首阶段独占仅新 `internal/adapters/events/relay_session*.go`（含测试）与 `docs/access/http.md`，先做公开受控样本下读取 / 身份绑定 / 投影 / 幂等；assembly、HTTP、Today 在接缝及用户决定齐备后由原 owner 续作。活动回执不代表功能完成。

跨线消费：原文与计划从精确文档提交读取；如需在业务线登记三份 Markdown，以同内容新文档提交保留出处，不把整个 validation 父历史合入 workbench。业务 SOURCE/REPORT 则在原 workbench 线普通合并，保持 v1.2 当前 CI 候选不混入新采集代码。

## v1.3 采集条款与可复用端口

- `RelayState.Sessions` 在独立 relay.sqlite 中按 terminal_id 持久化；不得假设 `ProjectRepository.ListSessions` 可取得 Relay 持有会话。R 通过 Service 提供只读内部绑定，H 不直接写 store。
- 沿用 `NativeAdapter` 的身份 / 鉴权边界与 `Observe`、`RelayLifecycle.ObserveOwned / Shutdown` 生命周期；GET 只读缓存，生命周期仅观察人工已接受的任务，取消并 join 后关闭存储。不得自动 launch、send、resume 或重新派发。
- 现有 `ReadContext` 是有条数 / 字节 / 页数上限的 user/assistant 文本投影，不是完整事件归档；不能当 L1 全事件采集，也不能为了采集放宽旧接口。H 在新 `relay_session*` 适配器中实现完整行增量读取，分别记录游标、身份与覆盖缺口；`Observe` 的活动摘要同样不替代完整归档。
- 新 task 使用新 native session；同未完成 task resume 保留原 session；terminal_id 不绑 PID。terminal-only 需真实原生归档或既有端口的正证据，不凭终端有输出就认定 Session 覆盖完整，不扫描家目录找身份。
- 原 `internal/adapters/events/relay.go` 到 `NewPendingEventService` 的通道可复用，但 published 只表示本机 admission；另行核对完整 Session 覆盖与 Today 展示，不能偷换为正式认知更新。
- 采集仅产生事件；现有权重与规范化规则不改，不擅自把未知 native_activity 映射为 edit/commit 或赋固定信号强度。无映射保持未知，完整性统计仍包括它。
- Today 沿现有只读 API 的六字段事件投影消费本机独立库；与快照确证数值并置的具体语义待用户决定。正式采集根、每日快照、认知状态、schema 与协议均不变。

## 核心验收

1. 完整性：超长 JSONL 行、末尾未完成行、连续追加、重启后读取、重复事件、格式未知与错误记录均可核对；只读获准会话，不遍历其他账户或历史根。
2. 身份：terminal_id 与 task_id、native session、原 CWD 对应；换 Agent 不换工作区，新任务不偷用旧 session。原生会话 metadata 不得由 PID 或客户端名猜测。
3. 存储：复用既有 EventService / EventRepository；稳定 origin 去重，失败保留首证据，不自动重发模型或换身份。当前旧 sink 的 published 仅表示 pending observation，不能拿它替代此次事件覆盖与 Today 对账。
4. 展示：沿六字段 `event_id/kind/topic/tokens/scope/ts`；新增会话事件自动出现在同 URL Today。快照 A/M/C 与 Sleep 仍按原确证来源，采集层不计算或写认知状态。
5. 真实演示：一个此前未登记的真实 CLI，仅增加注册数据，经人工确认启动，三分钟内在 Today 看到它的真实事件；演示不得由 fixture/mock 或人工采集替代。
6. 活性：活动事件证明 live、真实进程退出证明 exited，静默/联系丢失为 unverifiable；不因重新读取缓存而续命。
7. 隔离：V4 正式库与采集根、锁定协议、mapper、引擎、schema/迁移、七类权重和 00:30 快照管道零改动。零写审计保持原 Procmon 验收，缺管理员捕获不填 PASS。

## 决策与停止点

- 新任务明确要求 workbench 独立库和自动 Session 事件进入 Today；具体读侧接缝交原 H owner，不修改每日正式快照策略。
- 原主控既有未决项：CLI 中性工作区读写/执行授权、脏/未跟踪/非 Git 作品范围、活性窗口；先继承其最新裁决，不重复询问已答事项。未答时相关真实验收保持未运行，不能缩减为无工具聊天。
- Session 绑定/来源、独立库消费权限、未知 Agent 原生格式如仍无法从事实源和已有公开端口确定，集中向用户询问；可先实现不依赖决定的读取/投影/幂等单元。
- 本会话已集中询问：确认页的中性工作区工具范围、完整原件留存与 Today 派生边界、未提交源作品处理。未回答不激活 `LaunchPolicy` 或 `AllowCommittedGit`，不默选验收范围。
- 当前状态：精确 SOURCE 与互斥写域已确认，原 H 的完整读取 / 投影 / 幂等工程续作已观察到开工；SOURCE/REPORT 与适用验证待交付。真实 CLI / Today 完整闭环仍 NOT_RUN。文档准备和开工回执不是 L1 完成证据。

客户端：本会话 Codex，面板实际显示 GPT-6.1-Sol ultra fast；Worker 使用现有配置并登记实际模型，不凭客户端名推断。
