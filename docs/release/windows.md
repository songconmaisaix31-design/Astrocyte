# Astrocyte v0.1 Windows x64

完整解压 ZIP 到可写的新目录，双击 `Start-Astrocyte.cmd`，打开 http://127.0.0.1:8787。保持控制台打开，按 Ctrl+C 停止应用。Windows x64、Windows PowerShell 5.1 为基本运行条件；浏览器使用本机回环地址。产品包含 Go server、MCP、Sleep、手动快照工具与已编译的 web。S4/S5 发布准备包含独立 Node 和生产 SDK；用户无需预装这些 SDK 的开发环境。最终集成源码及生产配置尚未交付，本轮准备结果见 [发布记录](package-report.md)。

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
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Build-Windows.ps1 -Source <最终I的40位SOURCE> -Label final -Runtime 'D:\Astrocyte-agent-assets\20261015\release-package\runtime-prepared'
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Verify-Windows.ps1 -Zip <ZIP绝对路径> -RequireCombined
```

`-RequireCombined` 检查新目录中的空配置、真实 session/CSRF/SQLite 写读、手动快照及进程重启，结果是阶段运行检查。实际非空认知验证还需明确提供 `-Source`（包完整 SOURCE）、`-CognitionSourceSha`（已消费的认知工具源码）、`-CognitionSource`（获准的真实 SQLite）、`-CognitionHistory`、`-CognitionPublicationRoot` 和 `-CognitionSafeDir`。安全输入目录包含对应实际 Sleep 后状态的 `sql-expected.json`、`events-six-fields.json` 与 `admission-safe.json`；脚本核查当前发布代、十项 SQL 数值、全部六字段事件、排名、实际历史、分页与重启。验收数据只放在自己的验证目录，完整默认浏览器和原生控制器正向验收须分别通过才能发布。

构建复用现有 `pnpm install --frozen-lockfile` 和 `pnpm build`，并编译现有 `cmd/workbench-snapshot`，需要 Git、Go 1.27.2、Node 24.16.0、pnpm 11.27.0 与依赖下载网络。所有新下载、构建、归档、解压和验证目录都在 `D:/Astrocyte-agent-assets/20261015/release-package/`，`-Output` 可指定其中一个尚不存在的子目录；原 `.verification` 证据保留。构建前后检查 C 盘至少 10 GiB 空余。Go 父进程及子进程仅使用 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend` 和 `GOTMPDIR=D:/Astrocyte-agent-assets/20261015/go-temp/release-package`；pnpm store 与 npm cache 也在本次 D 盘目录，不改全局配置。

包允许四个 Go 程序、生产静态资源、启动与快照脚本、说明、LICENSE/NOTICE、legal 许可目录，以及 runtime 下的 Node、生产 SDK、准确适配脚本和两个固定模型目录。只有 `runtime/sdk/node_modules` 允许生产 npm 依赖；应用开发 node_modules、数据库、日志、私人配置、用户资料与正式自动任务继续拒绝。标准 .NET ZIP 支持这些大型资产；内容扫描逐块读取全部文件的 UTF8/UTF16LE 字符串，不将大型 sidecar 整体载入内存。扫描只输出相对文件名与检查范围，不输出匹配的私人字符串。最终 I 的适用 check/build、默认浏览器、原生控制器和新包实际核心路径验收仍须全部完成。

## 固定 S4/S5 资产准备

先显式运行联网准备命令；应用启动时没有隐式下载：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Prepare-Runtime.ps1 -Output 'D:\Astrocyte-agent-assets\20261015\release-package\runtime-prepared' -ModelRoot 'D:\Astrocyte-agent-assets\20261015\judgment-s5\sanitized-ctx_0a' -EnglishAssets 'D:\Astrocyte-agent-assets\20261015\judgment-s5\sanitized-ctx_0a\laya-english-assets.json' -MultilingualAssets 'D:\Astrocyte-agent-assets\20261015\judgment-s5\sanitized-ctx_0a\laya-multilingual-assets.json'
```

`-ModelRoot` 必须显式指定。当前获准的衍生资产根为 `D:/Astrocyte-agent-assets/20261015/judgment-s5/sanitized-ctx_0a`，两个 descriptor 为该根的 `laya-english-assets.json` 和 `laya-multilingual-assets.json`。准备程序核对两个 checkpoint 的十二项资产和两个 descriptor，复制后再次核对，保持输入不变。它不重新下载权重、不导出模型、不调用推理。每个模型目录保留 encoder/head ONNX、两份 `.onnx.data`、tokenizer、rl_agent_config 和 S5 worker 的 `assets.json`。S5 原官方 export 的 strict 加载和 CPU parity 记录是这些资产的来源；模型并非上游直接提供的 ONNX。

原始四份 ONNX 图的 metadata 含导出环境私人路径，原资产完整 runtime 隐私扫描实际失败并保留。当前输入为 S5 SOURCE `50d9cb299954fa41e381bd5d2d3ce758d00e353b` / REPORT `d85fa5e80350ca2d45ec67e6bbc5f7f4c47bbb73` 的独立新产物，十四项共 `3,014,649,292 bytes`，声明 `artifact_kind=astrocyte_privacy_sanitized_derivative`、`artifact_revision=metadata-clean-20261015-ctx_0a7797464f02`，带新文件身份和 `derived_from` 来源。检查要求这两个标识、准确上游 revision、六文件集合、四个未修改输入的原身份与全部实际字节相符；P 对该实际目录的原扫描为零匹配。S5 已交数值等价、三路六 PID 重启的实际验证，P 复用该结果；生产 HTTP、AT12 与最终应用组合仍按各 owner 的实际交付验收。准备与构建入口均执行完整 runtime 扫描，构建入口在应用源码归档前拒绝不合格输入。原始资产与首失败保留，净化后的图按 NOTICE 记作衍生文件。

可以显式传入本任务已取得的 `-NodeArchive` / `-NodeChecksums`、`-NpmCache` 和 `-DirectMLPackage`，复用同版本官方 ZIP、标准缓存和 NuGet；这些路径仍须在自身 D 盘范围，校验继续执行。未提供时只从脚本列出的固定官方来源联网准备，失败立即终止，没有自动更换来源或版本。

`-RuntimeSource` 显式指定本任务已经记录过来源的固定构建目录时，复制其中 Node、生产 SDK、标准 npm 锁与许可/SBOM，模型从本次明确的 `-ModelRoot` 和两个 descriptor 装配。复制后重新核对全部实际版本、native binding、十二文件身份和完整 runtime 隐私；该模式使用已提供的固定产物，目录缺失或检查失败立即终止。

| 路径 | 实际组件及配置用途 |
|---|---|
| `runtime/node/node.exe` | 官方 Node 24.16.0 Windows x64；官方 ZIP 的 SHASUMS256 校验 |
| `runtime/sdk/node_modules/laya-ts/dist/index.js` | 官方 Git `432312b36714d8889eba405507b37eda25257d9c` 的 laya-ts 0.1.0 子目录实际 build、标准 npm pack |
| `runtime/sdk/node_modules/onnxruntime-node` | 固定 1.30.0 CPU 原生依赖；安装禁用 lifecycle，实际加载原生 binding 检查 |
| `runtime/sdk/node_modules/vscode-jsonrpc/lib/node/main.js` | 固定 8.2.1，Go 对端为 sourcegraph/jsonrpc2 v0.2.1 / `2cc94179e136b565cb9df7d9d0bff9fc06912ac9` |
| `runtime/sdk/node_modules/@colbymchenry/codegraph` | 固定 1.0.1 / `a1489f77a6d69538bfe29020b8366ee034f90860` 的源码实际 build/npm pack，main 为 `dist/index.js`，包含生产 JS/WASM；实际 SQLite 能力检查；静态图检索不表示 embedding |
| `runtime/models/english`、`runtime/models/multilingual` | HF `convaiinnovations/laya` revision `7b928d828b7b0e022f929d9bd2e44165aa270148` 的实际官方转换资产 |
| `runtime/adapters/laya-worker.mjs`、`codegraph.mjs` | 最终精确应用 SOURCE 中的生产适配脚本，不能复制开发工作树 WIP |

上游 Laya 构建使用它自己的固定 package-lock 和真实 Windows TypeScript 编译器；构建目录不会进入包。生产 SDK 使用独立 package.json 和标准 npm package-lock，固定直接版本，完整传递依赖保留其许可文件，并输出标准 CycloneDX SBOM。Node LICENSE 保留其 bundled third-party notices，Laya Apache-2.0、模型卡、CodeGraph MIT 和 Go JSON-RPC MIT 原件放入 legal。模型格式转换的归属与改变说明在包内 NOTICE。许可文档不替代实际资产加载检查。

CodeGraph 同样使用准确源码的 package-lock 实际构建；发布 SDK 根目录直接导出 S4 使用的 API。ONNX 原 npm 包完整保留；其主项目 MIT、固定 v1.30.0 ThirdPartyNotices、随附 DirectML 1.15.4 的 Microsoft 二进制条款/源码许可/第三方通知，以及 DXC 1.8.2502.8 的完整 NCSA 复合通知分别放入 legal。DirectML 条款 §1(a) 允许其随 Windows/Xbox 的 ML 应用和服务分发，§3(c)/(e) 保留通知并限制独立分发；本包保留条款，不把其 DLL 归为 MIT。该 DLL 的存在不启用 GPU，实际 Laya 配置仍为明确 CPU。

S5 `NewLaya` 需要绝对 Node/Worker/ModelDir/AssetsPath/SDKEntry/RPCEntry 路径、明确 Language/ModelRevision/Stderr 和启动/退出预算。S4 `repositorycontext.New` 需要 Git 身份、Node/SDK/Script、投影目录、读取预算和适用 AOCI 路径。语言、模型、预算和准入规则由现有产品策略与原 owner 的生产配置确定；P 不填默认值。H/B 的准确生产配置 SOURCE 未交付前，启动脚本没有新增虚构环境变量，这些字段尚不能通过本包启动参数启用。

配置前可从已提交 owner SOURCE 准备准确 adapter 资产：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Prepare-Adapter.ps1 -Source faab2daf83f6a578773b61077140aff73af9ef9b -Component repositorycontext -Runtime 'D:\Astrocyte-agent-assets\20261015\release-package\runtime-prepared' -Output 'D:\Astrocyte-agent-assets\20261015\release-package\adapter-s4'
```

该命令只归档准确提交的单个 adapter 文件，标准解压并核对 Git blob、Node syntax 后放入 `runtime/adapters/codegraph.mjs`；不会覆盖已有版本。`-Component judgment` 对应准确 S5 SOURCE 的 `laya-worker.mjs`。最终应用包的两个脚本由 Build 从最终 I 的源码归档复制。S4 已观察到图数据库路径 255 字符时 SQLite WAL 建库失败，191 字符的实际普通路径通过；投影缓存目录须由可信配置在目标环境实际验证，源码目录保持原位，当前没有长路径支持保证。

AOCI 固定 rc20 / `45990a1dcfc4415efdc6fbfa13bc159a415f127c` 的当前许可为 FSL-1.1-MIT；rc20 公布于 2026-10-09，未来 MIT 于 2028-10-09 生效。其 Permitted Purpose 对提供相同或近似功能的商业用途有限制；分发还须保留 LICENSE/NOTICE 等归属。用户尚未决定本产品用途、外置配置或适用作者授权，AOCI 纳包方式保持未决定；当前准备程序没有复制它的二进制。显式配置合法取得的同版本 AOCI 可满足完整地图刷新功能，属于已提交的可行方案，尚未视为用户决定。现有 S4 完整能力验收继续保留。

Jev 仅为官方 `jev-1.13.0` 云 REST 模型，不分发本地权重或凭据。真实调用仍须进程 secret 和逐项证据外发授权；缺少 key 时真实 API 验证保持 NOT_RUN。Laya 本轮支持范围仍按 S5 实际结果：English/en、multilingual/en 和 zh；English 无效校准 bucket 在推理前拒绝，不替换模型或语言。OS 隔离、Jev、SemanticIndex 的未完成或未决定状态由原 owner 记录。

准备完成后运行实际依赖的小范围 ZIP/解压检查：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\release\Test-ReleaseTools.ps1 -Runtime 'D:\Astrocyte-agent-assets\20261015\release-package\runtime-prepared' -Output 'D:\Astrocyte-agent-assets\20261015\release-package\dependency-probe'
```

此命令核全部固定衍生模型身份、两个准确 adapter 的 Node syntax、真实 SDK/native binding/SQLite 和完整 runtime 隐私；归档实际 Node/SDK/adapter/许可并在新目录解压，逐文件检查原字节、允许项与完整 privacy 扫描。模型本体不进入此依赖 ZIP。它不构建应用候选、不运行模型推理，不替代最终 I 同源码包验收。PowerShell 5.1 与 7 的检查分别登记实际结果。包内 NOTICE 从最终 I 的准确源码复制，依赖的许可原件从已核准备目录复制。

构建进程使用 `GOFLAGS=-p=1`、`GOMAXPROCS=2` 和 `NODE_OPTIONS=--max-old-space-size=512`；退出后恢复原进程配置。这些设置仅限制构建资源，不删减适用检查。
