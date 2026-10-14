# Windows x64 发布包记录

发布 owner：`release-package-20261014`，分支 `songconmaisaix31-design/release-package-20261014`。只修改 `release/**` 和 `docs/release/**`，本机输出在自己的 `.verification`。实际执行客户端为 Orca Codex；当前终端 `term_e67c7493-cef5-439e-9687-3197880820d7` 的公开运行信息显示 `GPT-6.1-Sol high fast`。未读取私人凭据、原会话、正式数据库，未使用视觉功能，未发布 tag 或 GitHub Release。

## 发布工具 SOURCE

`c7a70be963c432bb72d0fa355705ce2907821120` 已 push。包括精确 Git 源码归档、复用生产构建、标准 ZIP、用户启动脚本、解压真实 HTTP/session/CSRF/SQLite/重启验证及用户说明。前三个阶段提交 `885cd595f811330590b1270464789d3c8142dcff`、`4e02520a0148c735faacd711e8e7cf57719d814e`、`2d54a7bc8657ffce2502cf90affb580308bc18a7` 保留普通追加历史。

## 候选包演练

演练输入 SOURCE：`c7a70be963c432bb72d0fa355705ce2907821120`。该基线没有最终 Workbench/Relay 组合；演练不计最终产品验收。

实际命令：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Build-Windows.ps1 -Source c7a70be963c432bb72d0fa355705ce2907821120 -Label rehearsal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Verify-Windows.ps1 -Zip 'C:/Users/DW/orca/workspaces/Astrocyte/release-package-20261014/.verification/package-rehearsal-20261014-202942/Astrocyte-v0.1-windows-x64-c7a70be963c4.zip'
git diff --check
```

构建 EXIT 0：Go 1.27.2/windows/amd64、CGO_ENABLED=0；Node 24.16.0；pnpm 11.27.0。冻结锁安装与原 `pnpm build` 完成 server/MCP/Sleep 三个程序、TypeScript 与 Vite 生产资源。依赖缓存复用标准 Go/pnpm 缓存，候选源码安装和全部交付输出位于 `.verification`。

ZIP：`C:/Users/DW/orca/workspaces/Astrocyte/release-package-20261014/.verification/package-rehearsal-20261014-202942/Astrocyte-v0.1-windows-x64-c7a70be963c4.zip`，26,309,211 字节。只包含三个程序、web 生产资源、两份启动脚本、README 与 LICENSE。

解压验证 EXIT 0，28 项检查通过。新目录 `.verification/verify-20261014-203035/extracted/Astrocyte-v0.1-windows-x64`，空的验收身份目录，系统 PATH 不含 Go/Node/pnpm/Orca/Agent CLI。实际健康 HTTP 200；八个 SPA 路径与两份静态资源 HTTP 200；无 session 的读与无 CSRF 的写 HTTP 403；bootstrap HTTP 200；带实际 cookie/CSRF 的分类创建 HTTP 201、读取 HTTP 200；跨来源写 HTTP 403。应用创建真实独立 SQLite，经两次实际 Windows 进程树启动/终止与新 session bootstrap，重启后的分类 ID 可读取，最终监听均已退出。

证据：`.verification/package-rehearsal-20261014-202942/build.log`、`.verification/verify-20261014-203035/{verify.log,result.json,initial-stdout.log,initial-stderr.log,restart-stdout.log,restart-stderr.log}`。静态 HTTP 检查不包含浏览器渲染；完整默认浏览器与原生控制器验收仍归最终组合验收。

## 首次失败与修复

- Windows PowerShell 5.1 首次加载 UTF-8 无 BOM 的中文注释脚本，报告 `MissingCatchOrFinally` / `UnexpectedToken`，未进入构建。三份脚本增加 BOM 后，5.1 实际 parser 与执行通过。
- 首次以 `$env:OS` 判定宿主，当前 Orca 进程没有该环境变量，报告 `Windows x64 build host required`。改为系统 API 的 `Win32NT` 与 64 位操作系统判定后通过。
- 首次 `git archive --output=(...)` 在 PowerShell 原生命令参数处理中将 ZIP 路径当作对象名，报告 `not a valid object name` / `git archive failed`。证据保留在 `.verification/package-rehearsal-20261014-202917/build.log`；改为一个完整 `--output=` 参数，另建演练目录通过。

以上失败继续保留，后续通过不覆盖失败记录。未自动重发任何模型调用。

## 可选能力与当前限制

基本包实际启动无需开发源码和 node_modules。默认关闭 summarize 和 Codex 自动沉淀，只运行应用；可选 summarize 0.25.1、Node、yt-dlp、FFmpeg、Whisper 二进制/模型及已安装 Python 的绝对路径参数见 [用户说明](windows.md)。MCP 和 Sleep 沿用原程序与原未确定参数，启动脚本不调用、不安装正式日程。

Workbench 已发布 `scripts/workbench/snapshot.py` 的 main 使用硬编码正式 SOURCE/HISTORY，未纳入本包、未执行。用户自有数据的安全快照生成入口存在缺口，已通过 Orca Handoff 通知主控并交 B/H 处理。`publish(root, source, history)` 可复用，但当前包没有可分发的用户显式 source/history/root 模块。没有快照时的 503 只证明不可用状态如实返回。

最终合并源码、认知公开自产来源、Relay 注册和真实 CLI 输入待唯一集成人员交付。接收其精确 SOURCE 后普通消费、重新构建、新目录解压并实际验证，随后追加独立 REPORT。完整默认 E2E、原生 event_id 闭环、最终发布与人工待定参数尚未因本候选包演练成为通过；正式 V4 数据、协议和日程保持原状态。
