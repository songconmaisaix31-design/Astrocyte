# Works Relay v1.3 · L1 会话采集执行计划

事实源：[L1–L4 原文](cognitive-stack-L1-L4.md) §2、§7；执行标准为当前根 [AGENTS.md](C:/Users/DW/orca/Astrocyte/AGENTS.md)。本轮只继续 L1；[L2 协议](L2-work-experiment-draft.md)保持草案，L2 实验与 L3/L4 实施等待既定开工条件。

## 目标与基线

新 Agent 通过注册数据接入，经人工派发后，其原生会话事件自动进入 Workbench 独立数据库，在同 URL 的 Today 页面可见。注册到可见事件的实际演示须在三分钟内完成，活动状态使用 `live / unverifiable / exited`。

业务目标分支为 `workbench/v0.2-access`。已集成 v1.2 SOURCE `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b` 的 CI 38469944551 双平台通过；本轮 L1 需要独立验证。

| 轨道 | 当前工程 SOURCE | REPORT |
|---|---|---|
| R 运行时 | `62c7d99f1d6b180ef19f90a2112e801702d9fffb` | `0b4ec501721361e8104a4e6e00781d73656157a4` |
| H 采集与 HTTP | `e3b63ab30151c372ae8bc3e6594227513ef8cea6` | `e12ce161d300c0d043bd8e3dccee87d489737db3` |
| F Today | `a500c417377a4b1404d8d4529dfa322a876810e4` | `cf0250f7e6106f6c2963e45e5be977093ed337a7` |

以上提交均已推送。实际原生控制器、完整采集与 Today 组合验收进行中；现有受控协议和 HTTP 回放结果不满足当前实际服务验收。

## 固定开发人员与文件责任

Run 为 `run_d577f0d3fee1`，当前主控为 `term_4a28ac76-bd04-4a8b-acb2-b3114ca74ee4`。主控负责计划、状态、决策、独立检查与最终提交；业务代码由原 R/H/F 开发人员负责。每轨保持原工作树和分支，跨域修改交 Handoff。

| 轨道 | 工作树与分支 | 独占 write_paths / 当前任务 |
|---|---|---|
| R | `works-relay-runtime` / `songconmaisaix31-design/works-relay-runtime` | `internal/workspace/{domain,app}/relay*.go`、`internal/adapters/{sqlite,agents}/relay*.go`、`internal/adapters/relayworktree/**`、`docs/access/runtime.md`；必要的 `agents/native.go/native_process.go` 可选观察端口同归 R，默认行为保持；Task `task_a51fab762c01` / Dispatch `ctx_1415e8205d75`，继续实际原生控制器与生命周期验证 |
| H | `works-relay-http` / `songconmaisaix31-design/works-relay-http` | `events/relay_session*.go`、`cmd/server/relay*.go` 与既有组装文件、`httpapi/relay*.go` 与 server、`workbench/app/relay*.go`、`sqlite/access_session*.go`、OpenAPI 与 relay examples、自身 `http.md`；解析依赖需要时独占 `go.mod/go.sum`；Task `task_cfb7c9bde07d` / Dispatch `ctx_976e732672da` |
| F | `works-relay-frontend` / `songconmaisaix31-design/works-relay-frontend` | 原 Relay UI 与 API、生成 schema、`TodayView.tsx`、cognition/relay 文件、新 `relay*.spec.ts` 与测试、自身 `frontend.md`；Task `task_3c33a99685aa` / Dispatch `ctx_1db29418e335` |

R 当前终端为 `term_fa1e374d-7982-4f44-988b-59c6b979f7a5`。H 当前终端为 `term_6d682528-790c-470d-a203-afe228214f50`，执行目录必须明确指定原 `works-relay-http`；其终端工作区元数据指向根目录。F 终端为 `term_480eeb0d-b66b-4cd4-8a9f-f3eb2750e0d4`。使用现有 Codex 配置，实际模型记录为 `gpt-6.1-sol`。

R/H/F 完成后，由原唯一 I 集成人员在 `works-relay-integration` 普通 `--no-ff` 合并精确交付，只处理必要路由、导入、配置、生成类型和 Today 父级来源说明；领域问题交回原轨。开发人员完成开发、测试和报告，SOURCE / REPORT 分开提交并 push。文档从精确提交读取后以同内容登记到业务线；validation 的业务父历史不合入 Workbench。

## 数据与接口

- `RelayState.Sessions` 和任务绑定由原唯一 Relay store 管理。R 只读端口逐次核查权限、任务、原生 session 和 CWD；私有归档路径不进入 HTTP。
- 沿已持有原生控制器取得具体会话归档来源，核实原生 ID、CWD、文件身份和当前读取范围；禁止扫描家目录、其他账户或历史记录寻找会话。
- Codex 官方 `rust-v0.162.0` 源码提交 `c1382380de69521303b416720a52f42d51af6248` 的归档保留策略会过滤部分原生通知。R 复用已持有控制器补充可选派生事件观察端口，H 消费精确交付；归档覆盖与全会话通知覆盖分别记录。默认 Native 行为保持，业务修改只进入 Workbench。
- 采集器观察人工已接受的任务，使用库解析成熟格式。错误立即返回，未知语义保持明确状态；未完成尾部不能推进游标。原子游标只保存位置、绑定、文件身份和派生覆盖状态，不保存原文。
- 事件复用既有独立应用数据库的 EventService / EventRepository。稳定 origin 去重；HTTP `event_id` 必须是实际已提交收据 ID，可直接回查 GetEvent。无变化时不保存游标。
- GET 只读取现有派生记录和覆盖状态，不启动控制器或采集。生命周期停止先取消并等待观察结束，再关闭存储；未知交付不自动重发。
- Today 事件只包含 `event_id/kind/topic/tokens/scope/ts`。原始思考、prompt、工具参数、stdout 和私有路径不进入产品字段。A/M/C、Sleep 与每日统计继续读取既有快照；采集器只产生事件。
- 当前文件字节范围读完、原生会话全生命周期覆盖和认知更新分别核实。未知格式、错误、未完成记录与来源缺口均计入覆盖检查。

## 实际验收

1. 使用实际已安装配置的 Codex、真实独立 SQLite、真实 HTTP 服务及控制器实际产生的归档。验证实际追加、重启去重、退出和停止读取；逐事件比对数据库收据 ID。
2. 初始测试注册表没有该 CLI，仅增加 `AgentSpec` 数据并正常重启同 URL 服务，经人工派发后记录首次 Today 可见的原生 Session 事件 ID 与时间，从注册和派发两个起点计时。通过真实数据库 origin 区分原生 Session 来源与任务级收据 / Git 事件。全过程没有人工采集动作或新增平台派发代码。
3. F 在两个已有 viewport 连接 H 实际服务做 DOM 检查，验证六字段、时间排序、查询分页、未知 tokens、权限与覆盖缺口；禁止 mock、虚假数据和响应拦截。
4. 快照可用时比对实际读前与读后值；不可用时验证真实服务的不可用状态以及独立会话区域，不制造快照数值。
5. 活动事件证明 live，实际进程退出证明 exited，静默或联系缺失为 unverifiable；缓存重读不刷新活动证据。未决定的产品活性窗口保持未决定。
6. 验证修改范围与冻结域差异，执行适用检查和构建、`git diff --check`，核实精确远端 SHA。最终组合需要独立检查，既有 CI 不代替新 L1 验收。

## 权限与剩余条件

本轮可以在自己创建并持有的公开 Git 项目中验证固定公开任务，调用实际模型；文件和命令工具保持关闭。实际用户作品范围、中性工作区工具权限、原件留存和产品活性窗口尚未裁决，相关验证继续保持未运行。公开任务验证只证明其实际覆盖的接入和会话行为。

业务只进入 Workbench 分支。V4 正式材料、锁定协议、引擎、mapper、认知 schema / migration、权重和快照流程保持既定冻结边界。现有 17321 部署不覆盖。管理员 Procmon 零写审计与两日自动快照按各自真实条件验收，未执行不填写通过。

中间文件放各自当前工作目录已忽略的专用目录；代码通过文件编辑工具修改。禁止 Git 回滚、覆盖公共历史或删除无法解释的文件。最终文档只记录正确实现、实际结果和真实限制。
