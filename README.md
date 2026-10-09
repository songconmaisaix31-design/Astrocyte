# Astrocyte

个人研究与创造工作台。Go 模块化单体、React + TypeScript + Vite、SQLite；开发入口见 [AGENTS.md](AGENTS.md)，当前范围见 [STATUS.md](STATUS.md) 与 [S0 计划](tasks/S0-plan.md)。

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
pnpm --dir web exec playwright install chromium
pnpm test:e2e          # 1920×1080 与 1280×720，独立数据与端口
pnpm build             # dist/astrocyte[.exe] 与 web/dist
pnpm start             # 构建后启动本地服务，ASTROCYTE_WEB_DIR 指向 web/dist
```

Playwright 在隔离临时数据目录运行，默认 API `18787`、网页 `15173`；可用 `ASTROCYTE_E2E_API_PORT`、`ASTROCYTE_E2E_WEB_PORT` 覆盖。它不会复用其他工作树已启动的服务。报告与失败 trace 在 `web/playwright-report`、`web/test-results`。Linux CI 另运行 race 检查；Windows 本地不以缺少 C 工具链阻止纯 Go SQLite 验证。

`pnpm check` 与 `pnpm build` 对缺少后端或页面入口明确失败；只有契约/客户端先到位的工作树，仍需合入对应 S0 实现后才能验收完整程序。

## HTTP v1 与 fixture 边界

[contracts/openapi.yaml](contracts/openapi.yaml) 是唯一 HTTP 源。全部规格 §13 路径、版本化请求/响应/错误及示例在此维护；[web/src/api/schema.d.ts](web/src/api/schema.d.ts) 自动生成，typed fetch 使用 `openapi-fetch` 的生成路径类型。业务组件调用 [client.ts](web/src/api/client.ts) 的 `createReadApi`、读取助手或 `createApiClient`。

S0 实现健康、基座与集合只读查询；真实空数据库返回空集合。导入、准入、批准、原生接续、交接、认领、成果提交/采用和 SSE 在 S0 返回 **501 `unsupported_capability`**；请求不产生这些未来副作用。错误携带 `code/message/retryable/request_id/required_action`。客户端不会自动重试外部动作，也不会把连接失败转换为成功数据。

UI fixture 必须显式进入；默认页面使用真实 API。协议中的 `fixture-*` 示例用于验证显示与数据结构，不能证明真实导入、模型能力、任务执行或成果采用。S1–S6 与 AT01–AT16 的实际验收由后续切片完成。

## 分层

`internal/{attention,workspace,swarm}/{domain,app}` 保持两层；适配器在 `internal/adapters`，组装在 `cmd/server`。导入检查禁止领域 I/O、领域依赖应用/适配器、跨上下文直接引用与 HTTP/MCP 访问 SQL；应用使用消费方端口，组装层负责桥接。
