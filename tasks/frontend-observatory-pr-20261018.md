# 新版前端、官方图标与在线协作 DAG

用户 2026-10-18 提供 `C:/Users/DW/Downloads/astrocyte-frontend-submission (2).zip`，要求使用新版前端、媒体与 Agent 使用真实 logo、协作网络由 Agent 在线生成 DAG、OpenViking 可选配置。当前采用一个开发 Agent 提 PR、主 Agent review；代码、测试、文档及返修由同一开发 Agent 完成，主 Agent 不修改产品业务代码。

2026-10-19 后续开发按 [动态 DAG 与独立 Reviewer 循环](dynamic-dag-loop-20261019.md) 继续：用户已确认动态 DAG 与 Worker 消亡机制，主控建议、Worker 实现、独立 Reviewer 验证并决定是否接受。本文件以下保留前轮657基线的交付和验证记录。

## 基线与目标

管理分支 `validation/v1-20261012` 接管基线 `9d63adee51d6d938e682947adf2833ac27b51052`；产品基线 `f17a491cdd6392829094c95a3e5efb6011362bc0`。前轮发送租约 Draft PR #1 为独立任务，不合并或覆盖。工作树固定为 `frontend-observatory-20261018`，实际分支为 `songconmaisaix31-design/frontend-observatory-20261018`，从准确产品基线创建，PR 目标为 `songconmaisaix31-design/works-relay-integration`。主控交接期间另有已提交的 V4 日运行记录 `8c3a45fd`，保持其原有内容，不改变正式实验。

压缩包已使用系统 ZIP 库解压到主控 `.verification/frontend-submission-20261018/`，原文件保持。包内31件源码/说明和补丁源自旧 `workbench/v0.2-access`，属于独立示例，包含沙盒事件、固定三 Agent 协作与 localStorage 业务状态。正式接入采用其新版布局、样式、交互和 SVG DAG 构件，保留现有生产路由、统一 API 客户端、人类批准、真实来源与服务端状态；不得运行或迁入沙盒业务。实际源码按文件编辑工具迁移，不直接运行包内脚本或盲目应用旧基线补丁。

图标优先复用 `web/public/brands` 及 `provenance.json` 已取得的官方资产，核查页面实际使用的品牌映射、文件格式和缺项。新增资产只从品牌官方站点/官方仓库取得，记录实际来源与获取结果；界面不得把通用图形或首字母显示成品牌 logo。

OpenViking 为可选集成，未配置时主程序与核心页面正常运行，显示明确未配置状态，仅禁用远端操作。保持既有 Go 服务与 `/api/v1` 权限边界，不引入示例 Node API 作为第二业务后端；凭据仅在服务端，资料外发依现有授权。配置不完整或上游失败立即返回准确错误，未知写入不重放。先读取包内适配及官方接口，提交准确 Go/app/adapter/HTTP 设计 Handoff 后实施新契约。

“Agent 在线生成 DAG”已向用户询问：Agent 主动规划任务依赖并关联实际执行事件，或仅把现有真实任务/事件实时组成 DAG。答复前保持未决定，不冻结新规划接口、不选择调用后端、不扩大原生工具权限。可以继续新版页面和图标接入、现有真实数据读取及 OpenViking 设计。收到答复后由主控传达准确范围。图节点与边须追溯真实来源、执行者、任务及事件；空数据为空图，不创建固定流程、伪造 Agent 状态或计划完成事实。DAG 无环与缺端点在错误位置拒绝，Agent 生成草稿与已批准执行事实必须明确区分。

## 唯一开发轨道与写入范围

- 固定 Agent / Worktree：一个 Codex 开发 Agent，`C:/Users/DW/orca/workspaces/Astrocyte/frontend-observatory-20261018`。Orca 实际创建分支为 `songconmaisaix31-design/frontend-observatory-20261018`，沿该分支持续开发。Run 沿用 `run_73578b70c449`，主控 `term_d7578194-fd20-42b4-9668-b0b0e8683571`；Task `task_006c2800099b` / Dispatch `ctx_06a3baf3e8db`，终端 `term_18b3ab1d-5471-40ad-8bfa-78fe847c5668`。实际客户端 Codex 0.162.0，模型 GPT-6.1-Sol xhigh fast，沿已有配置。启动回执 ready；实际读取确认任务仍在 composer 后只补一次 Enter，没有重复任务。
- 允许 `web/**`、依赖所需的根 `pnpm-lock.yaml`、`dependencies.lock.json`，以及 `docs/standards/frontend-observatory-20261018.md`、产品 `AGENTS.md` 的最新执行流程同步。
- 首次任务先检查当前生产 UI、全部提交包源码、可复用真实 API/品牌资产，并交接完整页面迁移和后端设计。OpenViking 的具体新增 Go/app/adapter/HTTP、OpenAPI/客户端生成路径，以及 DAG 规划相关后端写入范围，经主控 review 后追加到本计划。答复和写域批准前不修改这些业务文件。
- 不修改前轮租约源码、正式数据库、实验数据、原用户配置、共享服务、发布包或默认 CI。不得删除无法解释的旧功能、测试或队友修改；业务能力须在新版页面可达。

## 验收、资源与停止条件

全程遵循当前主控仓库 `AGENTS.md`、DDD、fast-fail、文件编辑工具、无 mock/假数据/伪造原生或 CLI、无新增 Skip/移除断言。禁止视觉功能，浏览器验证使用文本/DOM/网络事实。合法真实来源不足时记录未验，不放宽原生文件/命令工具关闭、S3 预算/能力证据等原有条件。

构建、测试和服务串行，最多一个验证服务；Go `GOMAXPROCS=1`、`GOFLAGS=-p=1`、`GOMEMLIMIT=256MiB`、BelowNormal；Node/浏览器使用单 worker、适度堆限制与 BelowNormal，不新增高并发。每项重操作前两次采样 CPU <70%、可用内存 >=4 GiB，间隔2–5秒；不足时不启动重操作，继续静态开发与审查。复用 D 现有缓存，临时结果放本树已忽略 `.verification/`。不关闭其他应用或用户终端。

验收包括前端 typecheck/lint/build、适用无替身测试、真实 HTTP/SQLite 空与拒绝路径、1920×1080及1280×720 DOM/键盘交互、真实品牌资源响应和图来源追溯；DAG 与 OpenViking 按最终批准设计做适用实际验证。只运行必要检查，完整旧替身集合不计通过。编译或核心路径失败由同一开发 Agent 迭代修复。用户未答或原生来源不足只停止依赖该前提的实施/验收，独立部分继续。

每阶段清晰 Commit，普通 push，同一 PR。提交保持 `[skip ci]`，完整默认 CI 单列未通过，不改 workflow/断言/Skip；主控发布当前准确 HEAD 的 review。禁止 force push、合并、共享部署、打包或正式发布。最终提供文件清单、Commit SHA、PR、实际命令/结果/原件及剩余限制；交付后正式 worker_done，由主控明确复用或官方释放。

## 2026-10-19 设计 review 与追加写域

开发 Agent 已普通 push 设计提交 `99b48e53b9fb3a02b419177715e3df4fce906064`，Handoff `msg_bb4fad1f9456`。主控完整读取报告、页面映射和六条候选 HTTP 接口，批准按该设计实施页面及可选 OpenViking；DAG 主动规划/事件投影的提问仍待答。本次资料外发沿 SPEC §17.1 的显式授权：配置仅提供连接能力，每次上传由真实人类会话明确提交指定项目、固定资料 revision 和当前目的地，并持久化该操作的主体、来源、版本与目的地。该确认仅覆盖本次 OpenViking 外发，不授予 Agent、原生控制、S2 执行或自动同步权限；不使用 ExternalModelCLI 许可替代。验证继续限定 Agent 创建的公开 Git 项目和获准公开来源。

追加允许写入：`cmd/server/openviking*.go`、`cmd/server/main.go` 的必要装配；`internal/attention/domain/openviking*.go`、`internal/attention/app/openviking*.go`；`internal/adapters/openviking/**`；`internal/adapters/sqlite/openviking*.go`；`migrations/027_openviking.sql`；`internal/adapters/httpapi/openviking*.go`、`internal/adapters/httpapi/server.go` 与 `internal/adapters/httpapi/session.go` 的必要注册/分类；`contracts/openapi.yaml` 与 `contracts/examples/openviking*`；前述范围内相邻真实测试。若还需修改已有 app Contracts/Options/构造器或其他源文件，先交准确路径 Handoff。现有 SQL embed 自动包含新迁移，保持已有迁移文件。

六条候选契约按设计报告实施：状态、项目提交列表/创建、既有提交刷新、项目检索、项目分层读取；复用 `/api/v1` 客户端与权限防护。状态与身份缺失不可显示已连接；任务查询仅使用已持久化关联的真实 task_id；URI 解析后核规范路径与项目根，重定向和越界结果立即失败。上传与资源提交分别保留阶段，同 request_id 只允许一次副作用；响应丢失、持久化失败、重启在途与未知状态不能自动重发或登记完成。外部请求不持有 SQL 事务，发送前核当前资料、引用、权限与目的地，配置变化不改写原记录。普通核心页面不依赖可选状态成功。无真实上游来源的正向验收单列未验，契约示例不得伪造完成事实。

## 2026-10-19 PR 交付与最终 review

- 提交：设计 `99b48e53b9fb3a02b419177715e3df4fce906064`，业务 `f48122ba5eb4f469bf13cb0e27539942ceee3841`，同一开发 Agent 返修 `657ca1464517a171ed50ddc8fcc787b237c44ad9`；均普通 push，最终工作树干净。Draft PR #2：<https://github.com/songconmaisaix31-design/Astrocyte/pull/2>，远端 HEAD 已独立核对。
- 完成：新版生产布局和焦点画布，保留真实路由/资料/批准/作品能力；26件官方媒体/Agent 资源及别名；可选 OpenViking 可信配置、DDD 领域与应用用例、Go 上游适配、SQLite 迁移、既有 HTTP/CSRF、生成契约客户端及人工固定资料外发界面。Grok 使用官方原始 SVG；媒体依实际域名显示已有官方资产。
- 主控完整审查初始源码与返修差异。首轮 [review](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5563932069) 的 P1 资源范围改绑和 P2 内容 level 缺失/null 已修复；资源 URI 首次绑定后只核同一规范 URI，数据库 CAS 限制已知 root_uri/task_id。准确最终 HEAD 的 [review](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5563954154) 为 COMMENTED，当前没有追加业务修改要求。
- 开发 Agent 实际检查通过：`pnpm --dir web typecheck`、`pnpm --dir web lint`、`pnpm --dir web build`；单 worker Vitest 三文件53项；Playwright 两视口10项真实 HTTP/DOM/键盘检查；`go test -mod=readonly ./internal/attention/domain ./internal/adapters/openviking ./internal/attention/app ./internal/adapters/sqlite ./internal/adapters/httpapi ./cmd/server -run '^TestOpenViking' -count=1 -timeout=2m`；`go vet -mod=readonly ./...`；`go build -mod=readonly ./...`；`pnpm check:contracts`；`node scripts/check-architecture.mjs`。四个 Go 测试函数覆盖参数规则及实际 HTTP/SQLite 未配置/连接拒绝/权限/必填层级路径，无测试函数的包只计编译。
- 最终原件：开发树 `.verification/go-20261019-005013-674.log`、`vet-20261019-005109-201-resource.json`、`go-build-20261019-005144-904-resource.json`；vet/build 无标准输出，以实际工具 EXIT0 为结果。前端及浏览器完整命令和原件见同树 `docs/standards/frontend-observatory-20261018.md`。早期失败与一次内存守卫拒绝均保留，资源满足后串行复验；没有提高资源上限或关闭其他应用。
- 非空焦点检查只读获准公开 Q 发布目录：`snapshot-1711995690`，时间 `2026-10-14T18:48:42.2097174Z`，公开 Git ref `b789466ff62f3c9b0e08ec392729f5bbfc5cfaa1`。没有新采集、Sleep、原生初始化或模型调用；历史快照不代表在线 Agent。
- 真实剩余：DAG 主动规划/现有事件投影的选择未答，在线 DAG 尚未实施，原拓扑入口保持禁用；没有实际 OpenViking 配置，正向入库、任务完成、未知外部副作用、检索和分层读取未验收；完整默认 CI 未执行、未通过。保留 JS 717.36kB 分块提示和契约13项既有警告。PR #1/#2 未合并，产品集成仍为 `f17a491c`，共享服务未部署，未重新打包或发布。
- 唯一开发 Agent 已收到最终 review；Handoff `msg_bdce00479624` 确认最终命令退出结果与自有 server/Go/Playwright 进程停止。正式 worker_done `msg_224a0d0c68be` 对应原准确 Task/Dispatch，主控立即执行官方释放，回执 `released / closed_agent_terminal / transcript captured`。未关闭其他应用，不再新建开发任务或重复检查。历史旧 Dispatch `ctx_7e6f00eb7fdb` 的 `release_unknown` 与用户保留终端继续沿原记录，不扩大清理范围。
