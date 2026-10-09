# Astrocyte

个人研究与创造工作台。Go 模块化单体、React + TypeScript + Vite、SQLite；开发入口见 [AGENTS.md](AGENTS.md)，当前范围见 [STATUS.md](STATUS.md)、[界面替换计划](tasks/UI-preview-plan.md) 与 [S0 计划](tasks/S0-plan.md)。

三页界面采用用户提供的 `Astrocyte-preview.html` 设计，迁入现有 React 应用。顶部搜索筛选当前页已加载数据；资料类型筛选、标签页历史与键盘切换、详情抽屉可操作。默认读取真实 API，`?fixture=1` 才进入显式示例模式。Attention 的 S1 接入见 [公共契约](tasks/S1-contract.md)；批准、接续与执行仍属后续切片。研究路线、固定动态与拓扑示例仅用于设计展示，不代表真实会话或实验结果。

## 环境

- Node **24.16.0**、pnpm **11.27.0**、Go **1.27.2**。直接依赖使用精确版本，完整锁定见 `pnpm-lock.yaml`、`go.mod`、`go.sum`；清单见 [dependencies.lock.json](dependencies.lock.json)。
- Windows/PowerShell 是本地开发主路径；Linux/WSL 使用独立环境与项目路径。Go/前端构建无需 Python 或 Docker；视频音轨回退复用 summarize 的 yt-dlp/本地转写依赖，用户已允许 Python。外部媒体工具与模型需独立安装和验证，npm 安装不代表视频可处理。
- Node 脚本依次检查 `ASTROCYTE_GO`、当前 `PATH`、Windows `%LOCALAPPDATA%/Programs/go/bin/go.exe`。设置了无效 override 时明确失败，避免静默切换工具链。

PowerShell：

```powershell
node --version
pnpm --version
$env:ASTROCYTE_GO = Join-Path $env:LOCALAPPDATA 'Programs/go/bin/go.exe'
& $env:ASTROCYTE_GO version
& $env:ASTROCYTE_GO mod download
pnpm install --frozen-lockfile
# 显式指定本地数据目录；不要放在仓库或 worktree 中。
$env:ASTROCYTE_DATA_DIR = Join-Path $env:LOCALAPPDATA 'Astrocyte/data'
pnpm dev
```

Linux / WSL：

```sh
node --version
pnpm --version
go version
go mod download
pnpm install --frozen-lockfile
export ASTROCYTE_DATA_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/astrocyte"
pnpm dev
```

打开 `http://127.0.0.1:5173`；API 监听 `127.0.0.1:8787`，Vite 代理 `/api`。`ASTROCYTE_PORT` 与 `ASTROCYTE_WEB_PORT` 可覆盖端口。按 Ctrl+C 停止本次启动的 API 和 Vite；只关闭浏览器不会停止服务。开发脚本构建临时后端可执行文件；临时目录只在子进程确认退出后清理。业务 SQLite 保留在数据目录中。

## 检查与构建

```sh
pnpm generate          # 从唯一 OpenAPI 源重新生成类型
pnpm check:contracts   # 校验规格、示例与生成漂移
pnpm check             # gofmt、vet、Go 测试、导入边界、TS、ESLint、Vitest
# 浏览器安装命令与共享开发机的保护设置见下文。
pnpm test:e2e          # 1920×1080 与 1280×720，独立数据与端口
pnpm build             # dist/astrocyte[.exe] 与 web/dist
pnpm start             # 构建后启动本地服务，ASTROCYTE_WEB_DIR 指向 web/dist
```

Playwright 在隔离临时数据目录运行，默认 API `18787`、网页 `15173`；可用 `ASTROCYTE_E2E_API_PORT`、`ASTROCYTE_E2E_WEB_PORT` 覆盖。它不会复用其他工作树已启动的服务。报告与失败 trace 在 `web/playwright-report`、`web/test-results`。Linux CI 另运行 race 检查；Windows 本地不以缺少 C 工具链阻止纯 Go SQLite 验证。

在共享开发机安装 Chromium 前设置 `PLAYWRIGHT_SKIP_BROWSER_GC=1`，保留其他项目可能使用的浏览器版本；该设置只影响安装时的旧版本清理，不跳过下载或测试。PowerShell：

```powershell
$env:PLAYWRIGHT_SKIP_BROWSER_GC = '1'
pnpm --dir web exec playwright install chromium
```

Linux / WSL：

```sh
PLAYWRIGHT_SKIP_BROWSER_GC=1 pnpm --dir web exec playwright install chromium
```

`pnpm start` 在 API 端口同时提供构建后的网页，默认打开 `http://127.0.0.1:8787/attention`；修改 `ASTROCYTE_PORT` 时使用对应端口。`ASTROCYTE_WEB_DIR` 可覆盖静态资源目录，`/api` 路径始终交给 API。

`pnpm check` 与 `pnpm build` 对缺少后端或页面入口明确失败；只有契约/客户端先到位的工作树，仍需合入对应 S0 实现后才能验收完整程序。

## HTTP v1 与 fixture 边界

S1 的本地浏览器读取先通过 `GET /api/v1/auth/session` 建立 HttpOnly/SameSiteStrict 会话；写入同时携带返回的 `csrf_token`（`X-CSRF-Token`）和稳定的 `Idempotency-Key`。正文不能设置 actor 或权限。服务重启后重新建立会话，调用者 ID 保持稳定，原幂等回执继续可用。`createAttentionApi` 提供生成类型的写入助手；连接失败保留原幂等键，不自动重发命令。可选 `ASTROCYTE_AGENT_TOKEN` 只标识 Bearer 身份，不授予全库读取权限，不能建立人类会话或写入；用户资料勾选/规则授权契约待定期间，Agent 资料与派生记录读取全部返回 403。此凭据边界不提供同机操作系统进程隔离。

S1 配置由入口显式读取：`ASTROCYTE_IMPORT_ROOTS` 使用平台路径分隔符（Windows 分号、Linux 冒号）列出可读资料目录，默认为空，拒绝本地文件读取。网页可直接上传/粘贴既有 summarize JSON/Markdown；导入器保留真实工具版本和来源，缺字幕或片段时不补造时间戳。arXiv 保存固定版本 PDF 和来源元数据，摘要不标为全文提取；资料版本的受控附件端点提供原始字节下载。`source_key` 和 `content_digest` 传空字符串表示由后端根据真实来源计算，非空值由适配器核验。

项目精确依赖 [`@steipete/summarize` **0.25.1**](https://github.com/steipete/summarize/blob/main/package.json)。运行 `pnpm install --frozen-lockfile` 后，`pnpm dev` / `pnpm start` 自动解析项目内 CLI 与当前 Node，不需设置全局 summarize 路径；不升级或回退全局 CLI。直接运行构建后的二进制时，从工作目录或二进制旁的已安装 checkout 查找项目依赖；将二进制单独复制到其他目录需要显式配置。`ASTROCYTE_SUMMARIZE_CLI`、`ASTROCYTE_NODE` 可覆盖为绝对路径；缺失/错误配置明确失败。入口会将 CLI 的 pnpm 链接解析为真实文件路径，使媒体桥能找到同一安装中的 summarize-core；显式覆盖 CLI 路径也执行此解析。`ASTROCYTE_ENABLE_SUMMARIZE=false` 显式关闭提取，保留 arXiv 元数据/PDF 与既有导出导入；普通临时浏览器服务也显式关闭提取。

公开视频链接使用现有 `POST /materials/imports`，`adapter=summarize_url`、`kind=video`，不传 `export_text` / `local_file_ref`，`source_key` / `content_digest` 传空字符串由后端核验。`summarize`、`summarize_json`、`summarize_markdown` 继续导入既有输出，`arxiv` 继续固定版本论文全文。提取保存真实原输出；字幕、转写与位置缺失不补造，网页文字不等于视频内容。提取成功不表示 Codex 首次整理成功；模型仍由独立显式授权入口控制。

本地音轨回退的启动配置使用绝对路径：`ASTROCYTE_YT_DLP_PATH` 为 yt-dlp 可执行文件；`ASTROCYTE_FFMPEG_PATH` 为 ffmpeg（同目录需要 ffprobe）；`ASTROCYTE_WHISPER_BINARY` 为 whisper.cpp CLI，`ASTROCYTE_WHISPER_MODEL` 为其本地模型。提取进程只传这些指定工具，隔离 HOME/配置目录，不继承提供商、cookie 或浏览器认证。缺工具时只使用实际上游能取得的公开字幕，失败明确返回缺证据/依赖错误，不自动调用云端转写。CPU 转写受现有 `ASTROCYTE_JOB_TIMEOUT_SECONDS` 限制；安装文件、版本命令成功和所选视频转写成功需分别核实。扩展/daemon 与登录浏览器权限仍待用户回答。

Windows x64 可运行 `pwsh -NoProfile -File scripts/install-media.ps1`（需要已安装 Python >=3.10；`-PythonExecutable` 可指定其绝对路径）。脚本读取依赖清单中的固定 yt-dlp、ffmpeg/ffprobe、whisper.cpp release 和固定上游 revision 的多语言 base 模型，安装至 `%LOCALAPPDATA%/Astrocyte/media`；`pnpm dev/start` 自动发现其中实际存在的工具。`ASTROCYTE_MEDIA_DIR` 或安装参数 `-MediaRoot` 可改为专用绝对目录，单项路径覆盖优先。安装仅使用版本目录/venv，不更新全局 CLI 或 Python 包。Linux/macOS 目前需自行安装对应上游工具并配置单项路径，Windows安装脚本不适用。

summarize 0.25.1 保留上游网络保护：所选公开来源应解析为实际可访问的公开地址。若系统代理的 Fake-IP DNS 把 arXiv/Bilibili 返回为 `198.18.0.0/15`，上游会拒绝，安装媒体依赖不能修复这一网络错误。需由用户决定代理/DNS配置，允许时把选定来源域名加入 Fake-IP 排除并使用其正常 DNS；项目不自动改变宿主代理、绕过 guard 或降级上游。配置完成后仍须重新验证所选材料的真实提取。

队列配置为 `ASTROCYTE_JOB_CONCURRENCY`、`ASTROCYTE_JOB_MAX_ATTEMPTS`、`ASTROCYTE_JOB_TIMEOUT_SECONDS`；设置边界由服务入口校验。作业期限和视频提取器共享 `ASTROCYTE_JOB_TIMEOUT_SECONDS`，正式入口默认 **1800 秒**，可显式配置 1–86400 秒。所选视频的真实本地转写已超过原 300 秒默认值，因此调整这一既有共享默认；它同时影响导入和自动整理作业的期限，Codex 自身执行上限仍为下述独立 180 秒。上游音轨下载另有自身时限，作业期限不覆盖上游限制；取消沿用现有机制。开发网页 Origin 默认来自 `ASTROCYTE_WEB_PORT`，额外本地 Origin 使用 `ASTROCYTE_ALLOWED_ORIGINS` 逗号分隔并列出完整 scheme/host/port。会话与 CSRF 不接受任意 loopback Origin。三层人工整理明确记录 manual 来源。本地 Codex 端口、适配、持久化队列和入口已组装；最新真实候选 schema 调用响应丢失、效果和费用未知，自动三层正路径尚未验收，不从人工记录推断自动成功。

自动处理配置使用 `ASTROCYTE_ENABLE_CODEX_DISTILLATION=true` 显式启用（默认关闭）、`ASTROCYTE_CODEX_EXECUTABLE` 原生 exe 绝对路径、`ASTROCYTE_CODEX_MODEL` 明确模型名称及 `ASTROCYTE_PROCESSING_SOURCE_KEYS` JSON 数组授权范围；没有任何内置来源白名单。`ASTROCYTE_CODEX_TIMEOUT_SECONDS` 默认 180 秒。普通浏览器临时服务关闭模型和外部提取。当前 native 政策验证的是 **Windows Codex 0.162.0**，使用已授权的纯文本推理模式，运行时拒绝工具执行；不是成功的操作系统文件读取隔离。模型推理会向其提供商发送选中原文，只配置用户授权的来源，本次仅授权所给公开论文与视频；私有库未授权。CLI 自己使用既有登录，不复制或显示凭据；代码不修改全局 CLI 权限配置。

`GET /api/v1/distillations/processor` 只核实原生版本与当前配置，返回实际配置模型、schema 身份及授权来源；不会调用模型，available 不代表提供商当前可连接。自动请求使用固定原文版本、显式前轮记录、问题与真实处理配置复用，完成后保留适配器来源和可空候选建议。人工确认再创建或修订候选。响应丢失的 operation 不自动或人工盲重发；已收到输出但本地保存失败的重试复用原输出。域分类和项目空间 @ 引用保留原件且不授予 Agent 访问。人工重读、提及与项目复用形成关注信号，刷新和机器读取不冒充人工行为；候选排序须先设置完整的版本化四维权重，未知维度不填零。

[contracts/openapi.yaml](contracts/openapi.yaml) 是唯一 HTTP 源。全部规格 §13 路径、版本化请求/响应/错误及示例在此维护；[web/src/api/schema.d.ts](web/src/api/schema.d.ts) 自动生成，typed fetch 使用 `openapi-fetch` 的生成路径类型。业务组件调用 [client.ts](web/src/api/client.ts) 的 `createReadApi`、读取助手或 `createApiClient`。

S0 实现健康、基座与集合只读查询；真实空数据库返回空集合。导入、准入、批准、原生接续、交接、认领、成果提交/采用和 SSE 在 S0 返回 **501 `unsupported_capability`**；请求不产生这些未来副作用。错误携带 `code/message/retryable/request_id/required_action`。客户端不会自动重试外部动作，也不会把连接失败转换为成功数据。

UI fixture 必须显式进入；默认页面使用真实 API。协议中的 `fixture-*` 示例用于验证显示与数据结构，不能证明真实导入、模型能力、任务执行或成果采用。S1–S6 与 AT01–AT16 的实际验收由后续切片完成。

开发服务下的三个 fixture 入口为：

- 资料沉淀：`http://127.0.0.1:5173/attention?fixture=1`
- 共同工作区：`http://127.0.0.1:5173/workspace?fixture=1`
- 蜂群执行：`http://127.0.0.1:5173/swarm?fixture=1`

构建后的服务使用同一路径和查询参数，默认端口为 `8787`。页面显示 fixture 标记；点击退出 fixture 或删除 `?fixture=1` 返回真实 API。后端数据库仍保持独立，示例不会写入业务数据。

## 数据与升级备份

业务数据库为 `ASTROCYTE_DATA_DIR/state.sqlite`。未指定数据目录时，后端使用 Go `os.UserConfigDir()` 下的 `astrocyte` 目录；上述启动示例显式指定仓库外路径。

已有迁移记录的数据库遇到待执行迁移时，先通过 SQLite **`VACUUM INTO`** 创建一致快照，包括 WAL 中已提交的数据。备份位于同一数据目录的 `backups/backup_<UTC日期时间>_<微秒>_v<当前schema版本>.sqlite`；例如 `backups/backup_20261009_060000_123456_v1.sqlite`。新建数据库及没有待执行迁移的重启不创建备份。

备份失败时启动终止，待执行迁移不会应用；迁移逐个事务提交，失败时回滚该迁移并终止启动。程序保留备份，不自动清理或自动恢复。需要恢复时先停止所有访问该数据目录的服务，保留当前数据库与 WAL 文件，再用备份恢复到匹配的程序版本；人工恢复不属于 S0 自动操作。

## 分层

`internal/{attention,workspace,swarm}/{domain,app}` 保持两层；适配器在 `internal/adapters`，组装在 `cmd/server`。导入检查禁止领域 I/O、领域依赖应用/适配器、跨上下文直接引用与 HTTP/MCP 访问 SQL；应用使用消费方端口，组装层负责桥接。
