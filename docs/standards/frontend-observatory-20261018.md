# 新版前端与可选 OpenViking 接入

## 身份与基线

- 日期：2026-10-18；开发客户端：Orca Codex；本会话模型：GPT-6。
- Run：`run_73578b70c449`；Task：`task_006c2800099b`；Dispatch：`ctx_06a3baf3e8db`。
- 工作树：`frontend-observatory-20261018`；分支：`songconmaisaix31-design/frontend-observatory-20261018`。
- 产品基线：`f17a491cdd6392829094c95a3e5efb6011362bc0`；PR 目标：`songconmaisaix31-design/works-relay-integration`。
- 当前已授权写入：`web/**`、必要的根依赖锁文件、本报告和产品 `AGENTS.md` 的执行流程同步。后端设计待主控 review 追加具体写入路径。

## 提交包与生产映射

已读取提交包的页面、来源表单、DAG 构建和 SVG、OpenViking 浏览器与 Node 适配、路由、配置、测试及说明。示例使用旧基线 `8f94d502763239f2c3e1e6167ed390830459ff45`，生产使用上述产品基线。

| 提交包构件 | 正式接入位置 | 真实数据与保留能力 |
|---|---|---|
| 左侧品牌、工作空间、导航、页头搜索 | `web/src/App.tsx`、`components/Nav.tsx`、限定在 `.ac-observatory` 内的样式 | 保留 History API 路由、加载内容筛选、键盘导航、系统能力与真实项目列表 |
| 注意力流的指标、深色画布、分布、右侧详情、活动列表 | `pages/cognition/FocusView.tsx` 与现有认知子页面 | 继续读取 `cognitionApi.focus` 的 A/M/C 与历史点，真实时间与证据入口保留，不将其换成演示百分比 |
| 资料来源的分类、列表、详情与仓库状态 | `pages/attention/AttentionPage.tsx` 和既有导入、媒体、论文组件 | `/attention` 与 `/attention/sources` 保留资料、收藏、研究机会、分类、正文、附件与授权操作 |
| Agent 卡片与交接面板 | `pages/workspace`、认知来源页面 | 本地清点、发现、注册、项目许可、原生会话、进度、提案与人类批准全部可达；卡片显示服务端观察 |
| SVG DAG 的方向切换、缩放、键盘节点、详情与导出 | `pages/swarm` 内独立显示组件 | 只展示可追溯的真实记录；新规划契约及调用后端待用户决定，不迁入 `buildAgentDag` 的固定流程 |
| 事件筛选、表格与证据详情 | 现有 `cognition/EvidenceView.tsx`、Mission 事件面板 | 保留事件 ID、来源、任务、执行者和原始证据入口，显示读取失败和过期状态 |
| 作品列表 | 保留 `/workspace/works` 及其现有子页面 | Works Relay 的登记、完成通道、许可、租约与交付能力继续使用原实现 |

页面采用提交包的固定左侧栏、72px 页头、浅绿色背景、清晰面板层次、深色图画布和右侧检查区；字号适配现有正文及双视口可读性。生产页面仍使用现有组件和 API，不复制示例 localStorage、事件场景、固定主题或三 Agent 推进器。已有默认路由保持，增加直接导航到焦点、资料来源和证据的入口。

品牌组件集中映射现有 `web/public/brands` 官方文件。需要核查 `claude-code`、`gemini-cli`、`qwen-code`、`cursor-agent` 等实际适配器标识；未知品牌显示明确的“官方图标未取得”文字。Kimi 和以 PNG 扩展名保存的 ICO 文件通过浏览器 `decode()` 核实实际可用性，新增资源记录官方来源、HTTP 获取结果及实际类型。

## OpenViking 官方接口依据

2026-10-18 读取官方仓库以下完整接口文档，本树本机原件位于 `.verification/official-openviking/`：

- [Resources](https://github.com/volcengine/OpenViking/blob/main/docs/en/api/02-resources.md)
- [Retrieval](https://github.com/volcengine/OpenViking/blob/main/docs/en/api/06-retrieval.md)
- [Content](https://github.com/volcengine/OpenViking/blob/main/docs/en/api/12-content.md)
- [Background Tasks](https://github.com/volcengine/OpenViking/blob/main/docs/en/api/17-tasks.md)
- [System](https://github.com/volcengine/OpenViking/blob/main/docs/en/api/07-system.md)

采用官方 HTTP 接口和 Go 标准 `net/http`、`encoding/json`、`mime/multipart`、`net/url`。适配器只覆盖本次五类操作，不建设通用代理。健康成功需另行核实认证身份；资源提交响应与任务完成分开保存。摘要和概览读取目录，正文读取文件；结果可能直接为字符串。检索返回的 score 按原值显示，不能转换成完成度或限定为 0..1。

## Go / app / adapter / HTTP 设计 Handoff（待 review）

### 可信配置与可选状态

`cmd/server/openviking_config.go` 读取服务端 `OPENVIKING_BASE_URL`、`OPENVIKING_API_KEY`、`OPENVIKING_ACCOUNT`、`OPENVIKING_USER`、`OPENVIKING_ROOT_URI`。全部缺省时装配明确的未配置能力，主程序及核心页面正常启动，状态读取返回 `configured=false`。仅配置了部分字段、非法 HTTP(S) URL、URL 凭据、query/fragment 或不完整 account/user 对时，在配置位置返回准确错误。基础地址不可由请求正文设置。

默认根为 `viking://~/resources/astrocyte`；别名通过带配置凭据的 `/api/v1/system/status` 解析真实 user，再生成规范根。禁止重定向；设置请求超时和响应大小上限；不记录 key、正文或认证头。认证头由适配器固定装配，前端仅得到状态、当前目的地和安全错误。

### 应用边界与持久化

`internal/attention/app/openviking.go` 提供状态、提交、查询提交、检索和分层读取用例；`openviking_contracts.go` 定义用途明确的 `OpenVikingGateway`、`OpenVikingRepository`。领域记录放在 `internal/attention/domain/openviking.go`，包含本地提交 ID、request ID、项目空间、资料 ID/固定 revision、真实来源、明确目的地、人工授权主体/时间、提交阶段、上游 task ID/root URI、最后观察及错误。

资料正文由既有 MaterialRevision/对象存储读取，并在外发前重新核查当前项目引用和权限。请求只接收现有资料固定版本，不能接收服务器文件路径、任意 URL 或自报已批准状态。项目根位于配置根下的 `/projects/{project_space_id}`，每个项目只检索自身已批准外发范围。新元提示词使用已有人工资料导入，保留其资料身份及正文版本。

人工提交在现有人类会话、CSRF、来源检查与幂等保护内完成，确认界面展示固定版本、目的地及将外发的内容。现有 `ExternalModelCLI` 许可针对 CLI；它不能直接代表 OpenViking 外发许可。新增外发确认的具体业务授权含义需主控确认后实施，不能仅凭配置 key 自动同步私人资料。

`internal/adapters/sqlite/openviking.go` 保存提交记录；独立迁移接入现有迁移序列。发起上游动作前持久化 submitting；相同 request ID 不产生第二次副作用。保存返回 task ID 后进入 submitted；刷新真实任务状态后更新 pending/running/cancelling/completed/failed/cancelled。进程重启后无 task ID 的在途记录显示 unknown；超时、响应损坏或提交后的本地持久化失败保留 unknown，禁止自动重发。temp_upload 与 resources 分别记住实际阶段，上传成功不等于资源已入库。

提交不启动原生 Agent，不创建 Mission，不改引擎和注意力值。只有实际 completed 及有效来源才可显示“已入库”。任务缺失或过期显示实际错误，保留原记录。

### 上游适配

`internal/adapters/openviking/client.go` 实现上述 Gateway：

| 用例 | 上游接口与约束 |
|---|---|
| 状态 | `GET /health` 与认证 `GET /api/v1/system/status`，分别核健康、initialized 和实际 user |
| 提交固定资料 | `POST /api/v1/resources/temp_upload` 上传既有正文快照，再 `POST /api/v1/resources`，指定项目 parent、create_parent、wait=false，不启用 watch |
| 查询任务 | `GET /api/v1/tasks/{task_id}?include_events=true`，只接受本地记录关联的真实 ID，保留历史缺失、dropped_count、started_mid_task |
| 检索 | `POST /api/v1/search/find`，固定项目 target_uri、context_type=resource、limit=10、include_provenance=true；不附带会话或扩大到全库 |
| 读取 | 依据已校验的检索结果层级，目录调用 abstract/overview，文件调用 read；返回 `{content: string}` |

所有 URI 使用标准 URL 解析并检查规范路径与项目边界；结果越界立即拒绝，不显示后再过滤。无凭据、无本地授权、资料版本变化、未知状态、上游 HTTP/业务错误在发生处返回。未知副作用不以 retryable=true 宣称可重发。

### HTTP 与客户端

`internal/adapters/httpapi/openviking.go` 注册在现有 Go 服务，`server.go` 的 Services 增加对应 app 用例，并使用既有 session guard / loopback / Origin / CSRF。候选契约：

| 方法与路径 | 行为 |
|---|---|
| `GET /api/v1/openviking/status` | 可选配置、目的地、连接核查与明确未配置状态 |
| `GET /api/v1/project-spaces/{id}/openviking/submissions` | 本项目真实提交记录 |
| `POST /api/v1/project-spaces/{id}/openviking/submissions` | 人类明确批准固定资料版本外发，带 request_id 与幂等键 |
| `POST /api/v1/openviking/submissions/{id}/refresh` | 读取本地关联的上游 task 并持久化观察，不重发原提交 |
| `POST /api/v1/project-spaces/{id}/openviking/search` | 人类对指定项目进行远端语义检索，携带 CSRF |
| `POST /api/v1/project-spaces/{id}/openviking/content` | 核查当前目的地和结果 URI 范围后分层读取 |

未配置远端动作返回标准 `ErrorV1` / HTTP503，状态仍为200；Agent 凭据无上述人类操作权限；方法错误及跨源请求保留既有拒绝语义。新契约写入 `contracts/openapi.yaml` 与必要 `contracts/examples/`，通过 `scripts/generate-client.mjs` 生成 `web/src/api/schema.d.ts`，`web/src/api/openviking.ts` 使用现有统一 API 客户端。生成文件不手写。

### 待主控追加的后端写入范围

`cmd/server/openviking_config.go`、必要的 `cmd/server/main.go` 装配；`internal/attention/domain/openviking.go`；`internal/attention/app/openviking*.go`；`internal/adapters/openviking/**`；`internal/adapters/sqlite/openviking*.go` 与该树实际迁移入口；`internal/adapters/httpapi/openviking*.go`、`server.go` 和必要 session guard 路由分类；`contracts/openapi.yaml` 与必要新契约示例；上述功能相邻的真实测试。新增业务依赖仅在确有需求时更新对应锁文件。

## DAG 的决定边界

用户尚未决定在线生成代表主动规划，还是实时组织已有真实任务/事件。现阶段只迁入 SVG 显示结构和现有真实读取，不声明规划已完成，不定义新规划端点或选择模型后端。

无论最终方式如何，节点与边都需指向真实任务、执行者和事件。已批准事实、执行观察及 Agent 草稿必须分开显示；无数据返回空图。重复 ID、缺端点、环在数据进入布局前拒绝。方向、缩放、过滤、导出与键盘选择为显示行为，不改变执行状态。Dagre 只负责图布局，新增版本在实际安装前核官方包信息。

## 验证计划与当前结果

构建、测试与服务串行；重操作前两次采样 CPU<70% 且可用内存≥4GiB，间隔2–5秒。Go 使用 GOMAXPROCS=1 / GOFLAGS=-p=1 / GOMEMLIMIT=256MiB，Node 与浏览器单 worker、适度堆限制和 BelowNormal；本树 `.verification/` 保存原件并复用既有 D 缓存。

必要验收为前端 typecheck/lint/build、适用无替身测试、独立真实 HTTP/SQLite 空与拒绝路径、1920×1080 和1280×720 DOM/键盘、官方品牌响应与 decode，以及实际来源与事件追溯。OpenViking 未配置路径独立验收；真实远端成功需合法实例与外发范围，未取得时保持未验。DAG 按用户最终决定及获准原生来源验收。

当前仅完成静态调查与设计；构建、业务测试、浏览器、真实 OpenViking 和在线规划均未运行。完整默认 CI 保持未通过，提交使用 `[skip ci]`，不修改 workflow/断言/Skip。每阶段普通提交与 push，同一 PR，主控负责 review；本任务不合并、共享部署、打包或发布。
