# Astrocyte v0.1 Windows x64

完整解压 ZIP 到可写的新目录，双击 `Start-Astrocyte.cmd`，打开 http://127.0.0.1:8787。保持控制台打开，按 Ctrl+C 停止应用。Windows x64、Windows PowerShell 5.1 为基本运行条件；浏览器使用本机回环地址。基本产品包含 Go server、MCP、Sleep、手动快照工具与已编译的 web，运行时无需 Go、Node、pnpm、Python 或开发源码。

修改端口或数据目录：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Start-Astrocyte.ps1 -Port 17887 -DataDir 'D:\AstrocyteData'
```

默认数据目录在解压目录的 `data`。SQLite 与对象位于 `data/local/astrocyte-workbench/v0.2-access/app-state`；Relay 注册与本地工作目录在同层 `v0.2-access`。保留整个数据目录，沿用相同 `-DataDir` 重启可继续读取。升级前停止应用并备份整个数据目录，新包解压到另一目录后明确指定原数据目录。不要将正式实验库或认知快照目录作为应用数据目录。

启动脚本为当前进程设置独立 `LOCALAPPDATA` 并清理继承的 `ASTROCYTE_*` 配置。Workbench 使用该目录下 `astrocyte-workbench/v0.1` 的既有只读快照协议；包不携带快照，没有发布快照时返回真实不可用状态。Relay 默认启用，空注册表不会启动 Agent；手动配置使用 `data/local/astrocyte-workbench/v0.2-access/agents.json`。Agent CLI 的安装、登录、控制能力和项目授权需分别成立；原生链路成功以实际控制器及事件证据为准。没有配置的状态、未知交付和生命周期缺口保留原含义。默认启动只运行应用，不安装自动任务、不执行 Sleep、不调用模型。

Relay 会话来源可显式指定 `-RelaySessionSource native_notifications`，使用原生控制器的实际通知；默认 `archive` 使用既有受控归档读取。活性窗口保持未配置，可由用户明确提供 `-RelayLiveSeconds`（1 到 86400 秒），不会自动选择数值。注册表沿用现有 JSON 数组格式：`name`、可信直接 argv 的 `launch_cmd`、`drive_mode`（structured、terminal、both）、实际观测的 `capabilities`，可选 `input_mode`（line、stdin_eof）。不要用脚本或命令行字符串替代直接 argv，不凭已安装或帮助输出填充成功能力。配置后仍需产品内的人工项目授权，当前未授权范围继续拒绝派发。

手动发布自己的实际认知 SQLite：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Publish-WorkbenchSnapshot.ps1 -Source 'D:\AstrocyteData\local\astrocyte-workbench\v0.2-access\app-state\state.sqlite' -DataDir 'D:\AstrocyteData'
```

Source 必须是用户明确选择的、包含现有认知表的实际 SQLite 绝对文件。可选 `-History` 指向用户自己的实际既有历史文件；未指定时不编造过去的历史。Source 与快照根必须独立。工具只读备份真实库并原子发布 `current.json`，应用可读取发布代的数据，空的认知表仍显示真实空状态。此操作无需 Python 或开发源码；启动脚本不自动执行，不调用旧正式实验快照脚本。

论文正文与视频提取是可选能力。启用前安装当前源码锁定的 `@steipete/summarize` 0.25.1 及其真实依赖，并提供绝对文件路径：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Start-Astrocyte.ps1 -SummarizeCli 'D:\tools\summarize\dist\cli.js' -NodeExecutable 'C:\Program Files\nodejs\node.exe'
```

Node 的构建基线为 24.16.0。视频按实际来源需要 `-YtDlp`、`-FFmpeg`、`-WhisperBinary` 和 `-WhisperModel`；列表导入可配置 `-ListingPython` 为已安装 yt-dlp 环境的 Python。PDF 提取需要 `-UVXExecutable` 指向实际 uvx 及其 markitdown 运行依赖。参数均为已存在的绝对文件。包不包含这些 CLI、模型、凭据、字幕或用户文件；缺依赖和无法取得真实内容时按原服务报告失败。上述命令不启用 Codex 自动沉淀；原生 Agent 使用产品中的人工项目授权流程。

`astrocyte-mcp.exe -h` 与 `astrocyte-sleep.exe -h` 列出原生参数。MCP 需要明确的项目、会话、调用者与 `signal-strength`；Sleep 需要明确日期、时区、衰减、初始置信度及事实投影。使用时明确传入应用的 `-data-dir`，未确定的参数由使用者决定。包中的启动脚本不调用这两个程序。

发布人员从完整的精确合并 SOURCE 构建：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Build-Windows.ps1 -Source <40位SOURCE> -Label final
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Verify-Windows.ps1 -Zip <ZIP绝对路径> -RequireCombined
```

`-RequireCombined` 检查新目录中的空配置、真实 session/CSRF/SQLite 写读、手动快照及进程重启，结果是阶段运行检查。实际非空认知验证还需明确提供 `-Source`（包完整 SOURCE）、`-CognitionSourceSha`（已消费的认知工具源码）、`-CognitionSource`（获准的真实 SQLite）、`-CognitionHistory`、`-CognitionPublicationRoot` 和 `-CognitionSafeDir`。安全输入目录包含对应实际 Sleep 后状态的 `sql-expected.json`、`events-six-fields.json` 与 `admission-safe.json`；脚本核查当前发布代、十项 SQL 数值、全部六字段事件、排名、实际历史、分页与重启。验收数据只放在自己的验证目录，完整默认浏览器和原生控制器正向验收须分别通过才能发布。

构建复用现有 `pnpm install --frozen-lockfile` 和 `pnpm build`，并编译现有 `cmd/workbench-snapshot`，需要 Git、Go 1.27.2、Node 24.16.0、pnpm 11.27.0 与依赖下载网络。精确源码归档、构建输出、ZIP 和验证数据只写入发布工作树 `.verification`。ZIP 只包含四个 Go 程序、生产静态资源、启动与手动快照脚本、此说明和 LICENSE；不包含数据库、开发依赖、测试原件、日志、私人配置、用户资料或正式自动任务。候选包演练与最终合并源码验收分别记录，发布渠道、tag 和 GitHub Release 由主控在最终验收后处理。
