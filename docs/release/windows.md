# Astrocyte v0.1 Windows x64

完整解压 ZIP 到可写的新目录，双击 `Start-Astrocyte.cmd`，打开 http://127.0.0.1:8787。保持控制台打开，按 Ctrl+C 停止应用。Windows x64、Windows PowerShell 5.1 为基本运行条件；浏览器使用本机回环地址。基本产品包含 Go server、MCP、Sleep 与已编译的 web，运行时无需 Go、Node、pnpm 或开发源码。

修改端口或数据目录：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Start-Astrocyte.ps1 -Port 17887 -DataDir 'D:\AstrocyteData'
```

默认数据目录在解压目录的 `data`。SQLite 与对象位于 `data/local/astrocyte-workbench/v0.2-access/app-state`；Relay 注册与本地工作目录在同层 `v0.2-access`。保留整个数据目录，沿用相同 `-DataDir` 重启可继续读取。升级前停止应用并备份整个数据目录，新包解压到另一目录后明确指定原数据目录。不要将正式实验库或认知快照目录作为应用数据目录。

启动脚本为当前进程设置独立 `LOCALAPPDATA` 并清理继承的 `ASTROCYTE_*` 配置。Workbench 使用该目录下 `astrocyte-workbench/v0.1` 的既有只读快照协议；包不携带快照，没有发布快照时返回真实不可用状态。Relay 默认启用，空注册表不会启动 Agent；手动配置使用 `data/local/astrocyte-workbench/v0.2-access/agents.json`。Agent CLI 的安装、登录、控制能力和项目授权需分别成立；原生链路成功以实际控制器及事件证据为准。没有配置的状态、未知交付和生命周期缺口保留原含义。默认启动只运行应用，不安装自动任务、不执行 Sleep、不调用模型。

论文正文与视频提取是可选能力。启用前安装当前源码锁定的 `@steipete/summarize` 0.25.1 及其真实依赖，并提供绝对文件路径：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Start-Astrocyte.ps1 -SummarizeCli 'D:\tools\summarize\dist\cli.js' -NodeExecutable 'C:\Program Files\nodejs\node.exe'
```

Node 的构建基线为 24.16.0。视频按实际来源需要 `-YtDlp`、`-FFmpeg`、`-WhisperBinary` 和 `-WhisperModel`；列表导入可配置 `-ListingPython` 为已安装 yt-dlp 环境的 Python。参数均为已存在的绝对文件。包不包含这些 CLI、模型、凭据、字幕或用户文件；缺依赖和无法取得真实内容时按原服务报告失败。上述命令不启用 Codex 自动沉淀；原生 Agent 使用产品中的人工项目授权流程。

`astrocyte-mcp.exe -h` 与 `astrocyte-sleep.exe -h` 列出原生参数。MCP 需要明确的项目、会话、调用者与 `signal-strength`；Sleep 需要明确日期、时区、衰减、初始置信度及事实投影。使用时明确传入应用的 `-data-dir`，未确定的参数由使用者决定。包中的启动脚本不调用这两个程序。

发布人员从完整的精确合并 SOURCE 构建：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Build-Windows.ps1 -Source <40位SOURCE> -Label final
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Verify-Windows.ps1 -Zip <ZIP绝对路径> -RequireCombined
```

构建复用现有 `pnpm install --frozen-lockfile` 和 `pnpm build`，需要 Git、Go 1.27.2、Node 24.16.0、pnpm 11.27.0 与依赖下载网络。精确源码归档、构建输出、ZIP 和验证数据只写入发布工作树 `.verification`。ZIP 只包含三个 Go 程序、生产静态资源、启动脚本、此说明和 LICENSE；不包含数据库、开发依赖、测试原件、日志、私人配置、用户资料或正式自动任务。候选包演练与最终合并源码验收分别记录，发布渠道、tag 和 GitHub Release 由主控在最终验收后处理。
