# Astrocyte

个人研究与创造工作台。Go 模块化单体、React + TypeScript + Vite、SQLite；开发入口见 [AGENTS.md](AGENTS.md)，当前范围见 [STATUS.md](STATUS.md)、[界面替换计划](tasks/UI-preview-plan.md) 与 [S0 计划](tasks/S0-plan.md)。

三页界面采用用户提供的 `Astrocyte-preview.html` 设计，迁入现有 React 应用。顶部搜索筛选当前页已加载数据；资料类型筛选、标签页历史与键盘切换、详情抽屉可操作。默认读取真实 API，`?fixture=1` 才进入显式示例模式。Attention 的 S1 接入见 [公共契约](tasks/S1-contract.md)；批准、接续与执行仍属后续切片。研究路线、固定动态与拓扑示例仅用于设计展示，不代表真实会话或实验结果。

## 环境

- Node **24.16.0**、pnpm **11.27.0**、Go **1.27.2**。直接依赖使用精确版本，完整锁定见 `pnpm-lock.yaml`、`go.mod`、`go.sum`；清单见 [dependencies.lock.json](dependencies.lock.json)。
- Windows/PowerShell 是本地开发主路径；Linux/WSL 使用独立环境与项目路径。无需 Python、Docker 或模型下载。
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

S1 的本地浏览器读取先通过 `GET /api/v1/auth/session` 建立 HttpOnly/SameSiteStrict 会话；写入同时携带返回的 `csrf_token`（`X-CSRF-Token`）和稳定的 `Idempotency-Key`。正文不能设置 actor 或权限。服务重启后重新建立会话，调用者 ID 保持稳定，原幂等回执继续可用。`createAttentionApi` 提供生成类型的写入助手；连接失败保留原幂等键，不自动重发命令。可选 `ASTROCYTE_AGENT_TOKEN` 只提供 Bearer 读取身份，不能建立人类会话或写入。此凭据边界不提供同机操作系统进程隔离。

S1 配置由入口显式读取：`ASTROCYTE_IMPORT_ROOTS` 使用平台路径分隔符（Windows 分号、Linux 冒号）列出可读资料目录，默认为空，拒绝本地文件读取。网页可直接上传/粘贴既有 summarize JSON/Markdown；导入器保留真实工具版本和来源，缺字幕或片段时不补造时间戳。arXiv 保存固定版本 PDF 和来源元数据，摘要不标为全文提取；资料版本的受控附件端点提供原始字节下载。`source_key` 和 `content_digest` 传空字符串表示由后端根据真实来源计算，非空值由适配器核验。

队列配置为 `ASTROCYTE_JOB_CONCURRENCY`、`ASTROCYTE_JOB_MAX_ATTEMPTS`、`ASTROCYTE_JOB_TIMEOUT_SECONDS`；设置边界由服务入口校验。开发网页 Origin 默认来自 `ASTROCYTE_WEB_PORT`，额外本地 Origin 使用 `ASTROCYTE_ALLOWED_ORIGINS` 逗号分隔并列出完整 scheme/host/port。会话与 CSRF 不接受任意 loopback Origin。三层人工整理明确记录 manual 来源；自动处理后端另待用户决定，不从人工记录推断自动能力。

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
