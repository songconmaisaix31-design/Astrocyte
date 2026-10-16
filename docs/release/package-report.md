# Windows x64 发布包记录

## 2026-10-16 Build 环境作用域修正、pwsh 功能验证与 S4 正向构造

本阶段 Task `task_3347b755071e`，Dispatch `ctx_1874978728b8`。固定原树和分支 `songconmaisaix31-design/release-package-20261014`。本轮三项修正：Build-Windows.ps1 环境作用域、pwsh 功能验证的准确进程级报告、一次真实 S4 完整参数正向构造。

### Build-Windows.ps1 Node 检查纳入 D 环境作用域

Check-Runtime.mjs 和 Inspect-Package.mjs 两个真实 Node 调用从 try 块前移至 try 块内，在 D 盘环境变量（GOCACHE/GOTMPDIR/TEMP/TMP/npm_config_cache/NODE_COMPILE_CACHE/ASTROCYTE_TEST_OUTPUT_ROOT/PATH）设置之后、source archive 之前执行。完整保存/设置/try-finally 还原作用域覆盖所有 Node 调用，确保构建环境隔离完整。Output 不复用、路径/space/privacy/fast-fail 守卫保持。本轮只做 Parser 适用检查，不执行 Build/3GB 扫描/最终包。

```diff
-& (Join-Path $Runtime 'runtime\node\node.exe') (Join-Path $PSScriptRoot 'Check-Runtime.mjs') (Join-Path $Runtime 'runtime')
-if ($LASTEXITCODE -ne 0) { throw 'Prepared real runtime is invalid.' }
-& (Join-Path $Runtime 'runtime\node\node.exe') (Join-Path $PSScriptRoot 'Inspect-Package.mjs') (Join-Path $Runtime 'runtime')
-if ($LASTEXITCODE -ne 0) { throw 'Prepared runtime privacy check failed before source archive.' }
 ...
 try {
     ... (D env setup) ...
+    & (Join-Path $Runtime 'runtime\node\node.exe') (Join-Path $PSScriptRoot 'Check-Runtime.mjs') (Join-Path $Runtime 'runtime')
+    if ($LASTEXITCODE -ne 0) { throw 'Prepared real runtime is invalid.' }
+    & (Join-Path $Runtime 'runtime\node\node.exe') (Join-Path $PSScriptRoot 'Inspect-Package.mjs') (Join-Path $Runtime 'runtime')
+    if ($LASTEXITCODE -ne 0) { throw 'Prepared runtime privacy check failed before source archive.' }
     Push-Location $repo
```

### pwsh 功能验证与 Go 子进程退出发现

新增 pwsh (PS7) 功能验证脚本 `functional-test-pwsh.ps1`，独立于既有 PS5.1 七项测试。实际进程级验证：

| 项目 | 实际结果 |
|---|---|
| 父 Shell | pwsh 7.6.6 PID 44048 |
| 子 Shell (wrapper) | pwsh.exe (PS7) PID 49240 |
| Go 子进程 | astrocyte.exe PID 46912 |
| 进程树 | pwsh -> pwsh -> astrocyte.exe（实际验证） |
| 默认无模型启动 | PASS，端口 19882 监听 |
| HTTP health | PASS，200 |
| SQLite 创建 | PASS，4096 bytes |
| 端口释放 | PASS |

**关键发现**：当 pwsh wrapper 被 Kill 时，Go 子进程 astrocyte.exe **不会自动退出**。实际测试中 wrapper PID 49240 退出（code=-1），但 Go 子进程 PID 46912 仍在运行，必须显式 force-kill。这证明旧 harness 的 Stop-TestApp 空 catch 没有证明 Go child 退出。新 harness 在 wrapper kill 后显式检查 Go 子进程状态，若仍在运行则 force-kill 并报告。

新 harness 改进：
1. Stop-TestApp 不再吞错误（空 catch 改为实际错误报告）
2. wrapper kill 后通过 Win32_Process CIM 查询确认 Go 子进程是否存在
3. Go 子进程未退出时显式 force-kill 并记录
4. 准确报告父/子 Shell 版本和进程树

既有 PS5.1 七项功能测试（默认启动、五拒绝用例、再次验证空默认）不重复。pwsh 测试仅覆盖默认启动/HTTP/SQLite/owned 退出，不与 PS5.1 重复拒绝用例。所有后继运行保持完整进程级 D Go/cache/temp/output-root/npm/Node cache 与 GOMAXPROCS=2/GOFLAGS=-p=1。既有进程 secret 只继承不读取打印复制，不新增 CLI 明文 key。

### S4 CodeGraph 完整参数正向构造

按根 `msg_1a8b11891180` 和模块 `msg_50b7b037d3da` 批准，使用 H 非秘密指针 `s4-config-public.json` 的四项现用值，在新 D staging 独占数据目录仅执行一次 S4 完整参数正向构造：

| 参数 | 值来源 |
|---|---|
| `CodeGraphNode` | `runtime-sanitized-third/runtime/node/node.exe`（Node 24.16.0） |
| `CodeGraphSDK` | `runtime-sanitized-third/runtime/sdk/node_modules/@colbymchenry/codegraph`（1.0.1） |
| `CodeGraphScript` | `runtime-sanitized-third/runtime/adapters/codegraph.mjs`（已提交同源） |
| `RepositoryEnvironmentID` | `H-shared-product-20261016-ad8413` |

实际进程级结果：

| 项目 | 结果 |
|---|---|
| 父 Shell | pwsh 7.6.6 PID 51016 |
| 子 Shell (wrapper) | pwsh.exe PID 29888 |
| Go 子进程 | astrocyte.exe PID 12656 |
| 进程树 | pwsh -> pwsh -> astrocyte.exe |
| 端口监听 | PASS，19883 |
| HTTP health | PASS，200 |
| SQLite 创建 | PASS，4096 bytes |
| **repository-context/runtime** | **PASS，实际创建（appdata/localappdata/profile/temp）** |
| owned 退出 | PASS，Go child 显式 force-kill |

**关键生产配置证明**：`repository-context/runtime/` 目录及其四个子目录（appdata、localappdata、profile、temp）的实际创建，证明 `resolveRepositoryContextConfig` 仅在完整有效 S4 CodeGraph 参数组加有效 RepositoryEnvironmentID 时才执行。无模型配置的默认启动不创建此目录；此效果不可由字符串匹配或模拟替代。AOCI 未配置（`aoci_configured=false`），不影响 S4 路径。

未索引仓库、未初始化 Laya/任何模型、未执行 native/ASR/下载/3GB 扫描/最终包。Node/Go/PS 全 D 环境，原共享 64505 和库不变。真实 runtime SOURCE 38fb2891 仅用于启动参数兼容与 S4 正向检查，不记为最终同源包。

本轮实际修改文件：`release/Build-Windows.ps1`（Node 检查移入 try 块）、`docs/release/package-report.md`（本章节）。pwsh 与 S4 功能验证脚本和日志保留在 D 盘独占 staging `staging-38fb-first`。未执行最终 Build/Verify/archive/Release/tag。严禁新 Laya 初始化/模型/ASR/OS/采集/Sleep/native/资产扫描下载。

## 2026-10-16 Helper 错误处理返修

本阶段 Task `task_3347b755071e`，Dispatch `ctx_39329742a5fc`。固定原树和分支，SOURCE `9e3464ed49f10e4e8e3ebc1d2271cf987850dd9d` 已普通 push，远端同 SHA。本 Dispatch 仅返修两个 D 盘外部 helper 脚本的错误处理，不修改任何 Git 跟踪的生产代码，不重跑 S4/新启动/模型/构造/索引/资产扫描。

### 旧 Helper 实际观察与末尾清理错误

两次历史运行的实际日志保留不变：

**pwsh 测试**（`functional-test-pwsh.log`）：pwsh 7.6.6 PID 44048 → wrapper PID 49240 → Go PID 46912，端口 19882 监听/health 200/SQLite 4096 bytes 全部 PASS。杀 wrapper 后 Go 未退出，显式 force-kill。PASS 后 finally 块重复调用 Stop-TestApp，因 `$childProcess = $null` 仅创建函数局部变量而未清除脚本级引用，导致第二次 Stop-TestApp 在已 Dispose 的 Process 对象上执行 Kill/WaitForExit，末尾出现 `Wrapper kill error PID :` / `Wrapper wait error PID :` / `WARNING: Wrapper PID  did not exit within 5s` 三行错误。这些错误来自 finally 重复清理，不代表测试本身失败。

**S4 测试**（`functional-test-s4-positive.log`）：wrapper PID 29888 → Go PID 12656，端口 19883/health/SQLite/repository-context/runtime 四子目录全部 PASS。同样的 finally 重复清理问题，末尾 `Wrapper kill error PID :` / `WARNING: Wrapper PID  did not exit within 5s`。

两次测试的核心 PASS 结果（启动、健康、SQLite、生产配置效果、owned 退出）均为真实观察，末尾清理错误不影响这些结果的有效性。但旧 helper 的干净完成度不准确——PASS 输出后仍有未捕获的异常行。

### 修订后的 Helper 改进（仅 Parser 验证，未重新运行）

两个 helper 已修订，根禁止重跑 S4/新启动，仅做 PS5.1 和 PS7 parser 检查：

| 检查 | PS5.1 | PS7 |
|---|---|---|
| `functional-test-pwsh.ps1` | PASS | PASS |
| `functional-test-s4-positive.ps1` | PASS | PASS |

修正内容：
1. **`$script:childProcess` 作用域**：Stop-TestApp 内使用 `$script:childProcess = $null` 清除脚本级引用，finally 块不再重复执行清理
2. **移除空 `catch {}`**：所有 Kill/WaitForExit/Get-Process/Dispose 错误不再静默吞掉；Kill 和 WaitForExit 失败直接 throw，Dispose 错误显式报告
3. **已退出进程的合法状态**：通过 `$proc.HasExited` 分支显式处理，不依赖空 catch 忽略异常
4. **进程身份核验**：Go 子进程的 Path 和 StartTime 在查询时记录
5. **数据目录不复用**：新执行根 `test-data-pwsh-v2` / `test-data-s4-positive-v2` 已存在即 throw，保留旧证据不被覆盖
6. **中文注释**：全部改为中文
7. **端口占用改为 throw**：旧 WARNING 改为 throw，SQLite 缺失同样 throw

### 只读端口检查

使用 .NET `GetActiveTcpListeners()` 检查四端口当前状态，不杀任何进程：

| 端口 | 状态 |
|---|---|
| 19880 | free |
| 19881 | free |
| 19882 | free |
| 19883 | free |

此前末次四端口 shell 检查实际为 ParserError（PowerShell 字符串变量作用域语法），不能作为验证。本次使用 `-f` 格式化避免了该问题。

### 准确声明范围

本 Dispatch 未重新运行修订后的 helper，因此报告不能声称修订后的 helper 已实际执行通过。历史运行数据（两次 PASS + 末尾清理错误）为旧 helper 的实际观察。修订后的 helper 仅通过 PS5.1/PS7 parser 检查，实际行为验证留待后续授权执行。SOURCE `9e3464ed49f10e4e8e3ebc1d2271cf987850dd9d` 仍是代码来源，未修改任何 Git 跟踪文件。Build-Windows.ps1 Node 检查在 D 作用域内已确认正确。未声称新最终包或原生/默认全绿。

## 2026-10-16 S4/S5/Jev 显式启动参数映射

本阶段 Task `task_3347b755071e`，Dispatch `ctx_70c076b66d81`。固定原树和分支 `songconmaisaix31-design/release-package-20261014`。普通消费 H/B 已交付的精确 SOURCE `fc294bee3d4a513d90cca6768d74fd4b90088394` 中 `cmd/server/config.go` 与 `relay_judgment.go` 的完整真实配置。

### 首轮提交 7bec269/3c0f5b2 返修

根 `msg_37028cb014d5` 明确五项返修要求，本轮逐项修复：

1. **删除 TypesafeApiKey 明文 CLI 参数**：移除 `TypesafeApiKey` 参数和 `TYPESAFE_API_KEY` 环境变量写入。Jev 进程 secret 由既有可信环境提供，启动脚本不读取/打印/复制 secret。原写入后未在 finally 还原违反边界。
2. **修复 StrictMode 下 null.Count**：`($layaPaths | Where-Object { $_ -ne '' }).Count` 在默认空参数下返回 null，StrictMode 访问 null 的属性抛异常。改为 `@($layaPaths | Where-Object { $_ -ne '' }).Count` 确保数组强制。真实默认启动（无模型配置）在 PS5.1 和 PS7 均实际通过。
3. **新代码注释改用中文**：所有新增注释改为中文。
4. **Build-Windows.ps1 环境变量完整设置**：`$priorGo` 新增 `TEMP`/`TMP`/`npm_config_cache`/`NODE_COMPILE_CACHE`/`ASTROCYTE_TEST_OUTPUT_ROOT`，均指向 Output 独占 D 子目录；完整保存原环境值，try/finally 恢复。`GOCACHE` 共享 `go-build/backend`，`GOTMPDIR` 独占 D，`GOMAXPROCS=2`/`GOFLAGS=-p=1` 保持。
5. **真实功能验证**：使用既有准确 D 盘 `relay/H/shared-product-ad8413/versions/38fb2891/astrocyte.exe`（runtime SOURCE 38fb，仅用于启动参数兼容检查，不记为最终同源包）和完整 web 复制到独占 D staging 目录。实际启动默认无模型配置，真实 HTTP 200 health、SQLite 创建、进程停止均通过；不完整 S4/S5 组和非法路径在正确位置拒绝；PS5.1 和 PS7 各 7 项实际执行全部 PASS。严禁假 exe/伪 CLI/mock 或校验脚本直接重复源码字符串作行为通过。

| 测试 | PS5.1 | PS7 |
|---|---|---|
| 默认无模型启动/HTTP health/SQLite/停止 | PASS | PASS |
| 不完整 CodeGraph 组拒绝 | PASS | PASS |
| 不完整 Laya 组拒绝 | PASS | PASS |
| 不存在 CodeGraphNode 路径拒绝 | PASS | PASS |
| 不存在 LayaNode 路径拒绝 | PASS | PASS |
| Laya 缺少语言拒绝 | PASS | PASS |
| 再次验证空默认启动 | PASS | PASS |

本轮实际修改文件：`release/Start-Astrocyte.ps1`（删除 TypesafeApiKey、@() 修复、中文注释）、`release/Build-Windows.ps1`（TEMP/TMP/npm_config_cache/NODE_COMPILE_CACHE/ASTROCYTE_TEST_OUTPUT_ROOT 环境设置和还原）、`docs/release/windows.md`（删除 TypesafeApiKey 使用说明）、`docs/release/package-report.md`（本章节）。

| 参数 | 环境变量 | 来源 |
|---|---|---|
| `CodeGraphNode` | `ASTROCYTE_CODEGRAPH_NODE` | config.go 官方 Node 24.16.0 绝对路径 |
| `CodeGraphSDK` | `ASTROCYTE_CODEGRAPH_SDK` | config.go `@colbymchenry/codegraph@1.0.1` 生产目录 |
| `CodeGraphScript` | `ASTROCYTE_CODEGRAPH_SCRIPT` | config.go 已提交 `codegraph.mjs` |
| `RepositoryEnvironmentID` | `ASTROCYTE_REPOSITORY_ENVIRONMENT_ID` | config.go 显式环境身份 |
| `AOCIExecutable` | `ASTROCYTE_AOCI_EXECUTABLE` | config.go AOCI v0.1.0-rc20 |
| `LayaNode` | `ASTROCYTE_LAYA_NODE` | relay_judgment.go 官方 Node |
| `LayaWorker` | `ASTROCYTE_LAYA_WORKER` | relay_judgment.go `laya-worker.mjs` |
| `LayaModelDir` | `ASTROCYTE_LAYA_MODEL_DIR` | relay_judgment.go 衍生资产目录 |
| `LayaAssets` | `ASTROCYTE_LAYA_ASSETS` | relay_judgment.go `assets.json` descriptor |
| `LayaSDKEntry` | `ASTROCYTE_LAYA_SDK_ENTRY` | relay_judgment.go laya-ts 生产入口 |
| `LayaRPCEntry` | `ASTROCYTE_LAYA_RPC_ENTRY` | relay_judgment.go vscode-jsonrpc 入口 |
| `LayaLanguage` | `ASTROCYTE_LAYA_LANGUAGE` | relay_judgment.go 显式语言 |
| `LayaStderrDir` | `ASTROCYTE_LAYA_STDERR_DIR` | relay_judgment.go 私有 stderr 目录 |
| `LayaStartupTimeoutMS` | `ASTROCYTE_LAYA_STARTUP_TIMEOUT_MS` | relay_judgment.go 正整数毫秒预算 |
| `LayaShutdownTimeoutMS` | `ASTROCYTE_LAYA_SHUTDOWN_TIMEOUT_MS` | relay_judgment.go 正整数毫秒预算 |

验证逻辑：S4 CodeGraph 三参数不完整组直接失败；S5 Laya 所有路径、语言、stderr 目录和生命周期预算不完整组直接失败；所有路径必须为已存在的绝对路径；语言、路径和生命周期预算由显式参数输入，不设未答默认。`LayaStderrDir` 不存在时自动创建。Jev 仅 cloud REST key，由既有进程环境提供，启动脚本不接受。脚本 parser 和实际功能验证（PS5.1/PS7 各 7 项）全部 PASS。`windows.md` 同步更新使用说明。未修改业务代码、依赖锁、迁移、原生控制器或组件文档。

SOURCE `4106fbf` 已普通 push，远端同 SHA。本 REPORT 随后单独追加提交。实际修改四个文件：`release/Start-Astrocyte.ps1`（删除 TypesafeApiKey、@() StrictMode 修复、中文注释）、`release/Build-Windows.ps1`（TEMP/TMP/npm_config_cache/NODE_COMPILE_CACHE/ASTROCYTE_TEST_OUTPUT_ROOT 完整保存和恢复）、`docs/release/windows.md`（删除 TypesafeApiKey 使用说明，Jev 改为既有进程环境）、`docs/release/package-report.md`（本章节）。D 盘独占 staging `staging-38fb-first` 保留功能测试日志。真实 runtime SOURCE 38fb2891 仅用于启动参数兼容检查，不记为最终同源包。C 盘空余 `9,056,583,680 bytes`（低于 10 GiB 阈值，仅 commit/push 操作）。未执行最终 Build/Verify/archive/Release/tag。严禁新 Laya 初始化/模型/ASR/OS/采集/Sleep/native/资产扫描下载。真实剩余：最终 I 精确 SOURCE 的完整 Build/Verify/archive、完整默认浏览器和原生控制器正向验收、C 盘空余恢复至 10 GiB 以上、tag 和 GitHub Releases 发布。

## 2026-10-15 固定运行依赖与打包准备

本阶段 Task `task_3347b755071e`，当前恢复 Dispatch `ctx_6cf3f6a0c68f`，固定原树和分支 `songconmaisaix31-design/release-package-20261014`。原派发 `ctx_d20c14245ccd` 已由运行时记录 abandoned / failed，协调者确认无其他写入者。本次接管 HEAD 为 `a4abbab5a0b095822437237c9216e96b0ffd5be8`，保留本报告原有未提交内容及四件未推送提交；当前终端 `term_c5c89219-2808-4a0d-99d1-893b1bbd9981` 的公开状态显示 Orca Codex / `GPT-6.1-Sol xhigh fast`。没有派发下级 Agent。原 P 工具 SOURCE `8e62c7f99b513115a19fc22569e33587bddcc300`、许可消费组合 `2790488346c0e0b9b28051793ed9d5cc318231dd` 及原来源历史保留。

本次准备 SOURCE 为 `7431db6451fa759638dedc4aad7687a217d3a6d9`，追加修改四个发布检查/构建脚本和 Windows 说明；父链包含原固定依赖、NOTICE、标准 adapter 准备与组件文档普通合并。未修改业务、根依赖锁、迁移、原生控制器或正式运行树。所有新增下载、源码构建、模型复制、标准 ZIP、解压和日志都在 `D:/Astrocyte-agent-assets/20261015/release-package/`，原 `.verification` 与 15f 候选保留。Go 父进程仅使用共享标准 `go-build/backend` 和独占 `go-temp/release-package` 的 D 盘进程配置；没有清理缓存或修改全局 HOME/auth/TEMP。大型操作的 C 盘空余检查保持至少 10 GiB，PS7 本次末次记录 `12,330,934,272 bytes`。

SOURCE 已普通 push，独立 `git ls-remote origin refs/heads/songconmaisaix31-design/release-package-20261014` 返回相同完整 SHA。首次只读远端 TLS 失败及首次 push 的 `remote rejected (failed)` / EXIT 1 保留，服务端没有提供更多原因；先核远端仍为 `2790488346c0e0b9b28051793ed9d5cc318231dd`，再一次普通 push 成功，没有更改鉴权、TLS、代理、全局配置或公共历史。本 REPORT 随后单独追加提交。

已完整读取并通过普通 `--no-ff` 合并消费 Flash 许可 SOURCE `da5bbf28eaa4776b6fa267c22bbf3bdf309df224` 和最新路径说明 SOURCE `e5d52481fa4a66d66d25497a9c80cc26a82441f1`；对应独立 REPORT 为 `2341b353cdc29ebeae71aa996b5fc1c8d31f859c`、`61cad05f13d56d82a1ac916a5f0027b64c429554`。合并无冲突，未代改 [组件文档](components-20261015.md)，当前组件文件与 e5d 的 Git blob 完全一致。路径明确为 `runtime/adapters/codegraph.mjs`，H/B 按实际文件和其准确生产配置接入。

### 实际准备与准确来源

当前完整准备目录为 `D:/Astrocyte-agent-assets/20261015/release-package/runtime-sanitized-third`，18:00:59 的 `prepare.log` 明确准备成功；Node/生产 SDK 显式复用本轨 `runtime-fourth`，模型使用原 S5 的 `sanitized-ctx_0a` 新 descriptor 和实际资产。原始模型分发仍被隐私检查拒绝，原件目录保留。

| 组件 | 固定来源与实际结果 |
|---|---|
| Node | 官方 `node-v24.16.0-win-x64.zip`，与官方 SHASUMS256 相符，实际运行版本 v24.16.0 |
| Laya SDK | 官方 Git `432312b36714d8889eba405507b37eda25257d9c`，laya-ts 0.1.0 子目录固定锁安装、真实 build、标准 npm pack 与生产安装通过 |
| CodeGraph SDK | 官方 Git `a1489f77a6d69538bfe29020b8366ee034f90860`，固定锁安装、真实 build/npm pack；包根 main=dist/index.js，生产 JS/WASM 和全部 S4 现用 API 导出核查通过 |
| ONNX / RPC | onnxruntime-node/common 1.30.0、vscode-jsonrpc 8.2.1；真实 native binding、SDK/RPC 导入和 Node SQLite 检查通过；Go jsonrpc2 v0.2.1 对应 `2cc94179e136b565cb9df7d9d0bff9fc06912ac9` 许可准确保留 |
| 模型 | HF `convaiinnovations/laya` 的 `7b928d828b7b0e022f929d9bd2e44165aa270148` 与上述官方源码 export 为原始来源；本次实际消费 S5 独立衍生资产，十二文件及两个 descriptor 共 `3,014,649,292 bytes`，身份与原 S5 新 descriptor 逐项一致；未重导出或调用推理 |
| 许可 | 原 Node LICENSE、Laya Apache-2.0、固定模型卡、CodeGraph/Go RPC MIT、ONNX v1.30.0 LICENSE/完整 ThirdPartyNotices 和标准 npm CycloneDX SBOM 已保留 |
| 附属 DLL | DirectML 实际版本 1.15.4，官方 NuGet 的 LICENSE、LICENSE-CODE、ThirdPartyNotices 通过标准 ZIP/XML 读取；DXC 1.8.2502.8 对应 `b4711839eb9a87da7c3436d9b212e0492359fbbd`，完整复合通知已保留 |

DirectML 二进制适用其 Microsoft 条款，§1(a) 允许其随 Windows/Xbox 的 ML 应用和服务分发，§3(c)/(e) 保留通知并限制独立分发；LICENSE-CODE 单独适用于 NuGet 的代码部分。实际厂商文件完整保留，Laya 明确 CPU 配置不变。许可原件在 `runtime-fourth/legal`，补充读取原件在 `onnx-legal-second/legal`，官方 nupkg 在 `onnx-legal-first/downloads`。精确归属和链接见 [NOTICE](../../release/NOTICE.md)。

当前 SDK 根为 `runtime-sanitized-third/runtime/sdk/node_modules/@colbymchenry/codegraph`，Node 为 `runtime-sanitized-third/runtime/node/node.exe`。S4 已只读消费同源 SDK，准确 SOURCE `faab2daf83f6a578773b61077140aff73af9ef9b` / REPORT `9907404b8df7f9f4a76dc7c90372e967fd2b33cd` 记录独立 D 盘普通路径下从零建立两图、实际 Search/Impact、只读查询、WAL 拒绝和重启通过。255 字符图数据库路径的 SQLite WAL 失败保留，191 字符普通路径实测通过；当前没有长路径支持保证。完整产品配置、前端与最终 ZIP 仍须后续准确组合验收。

两个已准备 adapter 逐项对应准确已提交源码：`runtime/adapters/codegraph.mjs` 与 S4 faab 的 Git blob `d35fff9daa0a20354b7f891706846935ef85823c` 相同，`runtime/adapters/laya-worker.mjs` 与 S5 `50d9cb299954fa41e381bd5d2d3ce758d00e353b` 的 blob `d630058470347ebed9decd3299192aa28c616e2f` 相同。`Prepare-Adapter.ps1` 使用标准单文件 Git archive / ZIP 解压和 Node syntax，不覆盖已有版本；最终应用包仍从最终 I 的准确源码复制两个 adapter。

实际衍生准备命令如下，EXIT 0；完整 argv 和结果保存在该目录 `prepare.log`。现入口要求显式 `-ModelRoot`，不使用原模型目录作为默认输入。

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Prepare-Runtime.ps1 -Output D:/Astrocyte-agent-assets/20261015/release-package/runtime-sanitized-third -RuntimeSource D:/Astrocyte-agent-assets/20261015/release-package/runtime-fourth -EnglishAssets D:/Astrocyte-agent-assets/20261015/judgment-s5/sanitized-ctx_0a/laya-english-assets.json -MultilingualAssets D:/Astrocyte-agent-assets/20261015/judgment-s5/sanitized-ctx_0a/laya-multilingual-assets.json -ModelRoot D:/Astrocyte-agent-assets/20261015/judgment-s5/sanitized-ctx_0a
```

`Prepare-OnnxLegal.ps1` 在 `onnx-legal-first` 完成官方下载，并在 `onnx-legal-second` 显式使用同一官方 nupkg 完成全部通知读取，均 EXIT 0。本次只读核对衍生准备目录的六份 ONNX/DirectML/DXC 许可，与 Flash 原文档记录的全部 SHA256 前缀相符；没有重新下载或替换许可。

### 前阶段依赖 ZIP 实际验证

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Test-ReleaseTools.ps1 -Runtime 'D:/Astrocyte-agent-assets/20261015/release-package/runtime-fourth' -Output 'D:/Astrocyte-agent-assets/20261015/release-package/probe-ps51-fourth'
pwsh -NoProfile -ExecutionPolicy Bypass -File release/Test-ReleaseTools.ps1 -Runtime 'D:/Astrocyte-agent-assets/20261015/release-package/runtime-fourth' -Output 'D:/Astrocyte-agent-assets/20261015/release-package/probe-ps7-first'
```

两命令均 EXIT 0。两个新目录均实际执行固定模型身份、native/SDK/RPC/SQLite 检查、全部 PS 脚本 parser、生产依赖复制、允许项校验、完整 UTF8/UTF16LE 字符串与凭据文件名检查、标准 .NET ZIP、新目录解压、解压 Node 和 S4 API、再扫描与逐文件字节核对。每个依赖包 **1384 文件 / 470,098,559 bytes**，私人匹配为零；模型本体没有进入这两个小范围 ZIP。此处的模型身份通过不代表模型隐私通过。

| 环境 | 实际 ZIP 字节数 | 证据 |
|---|---:|---|
| Windows PowerShell 5.1.26100.9711 | 155,404,809 | `probe-ps51-fourth/probe.log`、`runtime-dependencies-probe.zip`、`extracted/` |
| PowerShell 7.6.6 | 159,525,724 | `probe-ps7-first/probe.log`、`runtime-dependencies-probe.zip`、`extracted/` |

`Check-CodeGraph.mjs` 实际核对静态 init/open、indexAll/sync/getChangedFiles/getStats/close、QueryBuilder searchNodes/getNodeById/getIncomingEdges、setLogger/silentLogger。`node --check`、两个 PS parser、`git diff --check` / staged diff 检查通过。扫描实际既有含个人路径的旧发布报告时如实拒绝，安全日志 `privacy-existing-report-rejection.log` 未输出匹配内容。

### 本次恢复的完整 runtime 与依赖 ZIP 检查

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Test-ReleaseTools.ps1 -Runtime D:/Astrocyte-agent-assets/20261015/release-package/runtime-sanitized-third -Output D:/Astrocyte-agent-assets/20261015/release-package/probe-ps51-recovered-first
pwsh -NoProfile -ExecutionPolicy Bypass -File release/Test-ReleaseTools.ps1 -Runtime D:/Astrocyte-agent-assets/20261015/release-package/runtime-sanitized-third -Output D:/Astrocyte-agent-assets/20261015/release-package/probe-ps7-recovered-first
```

两环境均实际 PASS，PowerShell 7 命令实际 EXIT 0；Windows PowerShell 5.1 的原执行输出未取回 exit code，其 `probe.log` 在 21:56:37 记录全部检查 PASS 及正常 transcript 结束。本次核对 Node 24.16.0、真实 native ONNX binding、Laya/CodeGraph/RPC 导入和 SQLite；全部 PS 文件 parser 与两个准确 adapter 的 Node syntax 通过。模型只执行 descriptor/实际字节身份检查，要求衍生来源、修改图身份、未修改 sidecar/tokenizer/config 的原身份和全部文件集准确，没有推理。

完整 `runtime` 扫描实际覆盖 **1,388 文件 / 3,484,313,396 bytes / private_matches=0**，包含十四项衍生模型文件及两个 adapter。每个依赖包实际 **1,386 文件 / 470,110,352 bytes**，包含 Node、生产 SDK、两个 adapter、全部许可与 NOTICE；归档前允许项/隐私扫描、标准 .NET ZIP、新目录解压、实际解压 Node/CodeGraph API、再次扫描和逐文件字节比较全部通过。模型本体没有进入此小范围 ZIP。

| 环境 | 实际 ZIP 字节数 | 本任务 D 盘证据目录 |
|---|---:|---|
| Windows PowerShell 5.1.26100.9711 | 155,409,804 | `probe-ps51-recovered-first/` |
| PowerShell 7.6.6 | 159,530,734 | `probe-ps7-recovered-first/` |

每个目录保留 `probe.log`、`runtime-dependencies-probe.zip`、归档内容和 `extracted/`。两个 ZIP 字节大小因 .NET 压缩器不同而不同，解压内容按实际每文件比较相同。构建 NOTICE 现从最终 I 的准确源码读取；这条复制路径通过两个 PS parser，本阶段没有执行最终 Build 或应用 archive。`node --check release/Check-Runtime.mjs`、`git diff --check` 和 SOURCE 的 staged diff 检查通过；文档报告变更不重复产品测试。

### 模型隐私首失败与当前衍生来源

完整 `runtime-fourth/runtime` 扫描实际 EXIT 1，首文件 `models/english/encoder.onnx`。`Inspect-OnnxMetadata.py` 使用已安装官方 ONNX/Protobuf 库、`load_external_data=False` 和 `-B` 只读定位；四图包含当前用户 home 的字段均为 `onnx.StringStringEntryProto.value`、类型 string。安全派生结果 `onnx-metadata-privacy.json` 只包含类别、字段和数量，不包含实际值：

| 图 | 包含私人路径的字段字符串数量 |
|---|---:|
| english/encoder.onnx | 1633 |
| english/head.onnx | 155 |
| multilingual/encoder.onnx | 1289 |
| multilingual/head.onnx | 155 |

已通过 `msg_f4b08b2936cc` / `msg_4b3359bd4728` 交模块负责人和原 S5 `ctx_0a7797464f02`。根确认原 S5 通过成熟 ONNX 工具生成必要 metadata 净化的独立新产物，原十二资产/导出/首失败保留。当前已接收准确 SOURCE `50d9cb299954fa41e381bd5d2d3ce758d00e353b` / REPORT `d85fa5e80350ca2d45ec67e6bbc5f7f4c47bbb73` 和新 descriptor，明确 `artifact_kind=astrocyte_privacy_sanitized_derivative` / `artifact_revision=metadata-clean-20261015-ctx_0a7797464f02`，`derived_from` 保存原资产来源与身份。原 owner 的两 checkpoint、64/192 token encoder/三题型 head 输出逐项相等及三路六 PID 重启结果按准确报告复用，本轨没有重复推理。P 未修改模型、descriptor 或私人路径规则，没有豁免扫描。准备/构建入口都执行完整 runtime 隐私门禁，Build 在应用源码 archive 前拒绝不合格输入。

H 的 `31256d5d6758daf25892d55df73feffde7a7e573` / `89d84d55a822e2f0506e18e93adaa1ec02a8342e` 仍属其阶段交付；完整生产配置与最终 I 未交付，本轨未用 WIP 冻结参数或增加虚构 ENV。S5 原 CPU 验证没有代表 OS 隔离；Jev 没有 key 的真实 API 仍为 NOT_RUN。

### 首失败原件与当前范围

准备 `runtime-first` 的 TypeScript Windows 编译器依赖错误、`runtime-second` 官方 npm ECONNRESET、`runtime-third` npm CodeGraph 包装入口失败均保留各自 transcript。当前使用固定源码实际 SDK，不调用 CLI 安装器或宿主缓存。PS5 首三个 probe 原件保留：首轮 ZIP 成员反斜杠、下一轮原生库格式标签匹配、第三轮压缩程序集未加载；最终 UTF8 BOM、显式标准压缩程序集、前向斜杠成员名与文本材料匹配已在两个实际环境通过。没有跳过厂商文件或改变私人路径匹配条件。

衍生准备 `runtime-sanitized-first` 在固定 HF 模型卡下载失败后立即终止，`runtime-sanitized-second` 在复制已有 legal 目录时报错；两个原目录/日志保留。`runtime-sanitized-third` 使用显式已构建 runtime 来源并完成全部真实准备。现有 runtime-source 分支逐个复制许可成员到新目录，缺失直接失败，没有更换上游或自动下载替代。

本阶段未执行应用源码 archive、15f 重建、最终包、默认浏览器、原生控制器、采集、Sleep、Jev 或其他模型推理。原 `runtime-fourth` 的完整模型隐私失败继续有效，当前完整衍生 runtime 与旧原件分别保留。

### 最终组合的当前准确认知输入

协调者 `msg_f61e0255b359` 已通知唯一 I 启动，尚未交付准确组合 SOURCE。新 Q SOURCE `40880b67e4a668ec0e19d00eb4f77a76fdf210ab` / REPORT `f1110523c584aa7aa59ad07a05fc70cf88e138b7` 为后续同代验证输入。已只读其完整当前报告和 main 的 `current.json` / 安全 SQL / admission：Q 原树 `.verification/cross-day-20261015-0228/app-state/state.sqlite` 为来源，发布根为同目录 `local/astrocyte-workbench/v0.1`，实际 generation `snapshot-1711995690` / `2026-10-14T18:48:42.2097174Z`，History 为该代 `history.json`，安全期望为同目录 `main-safe-final`。

当前安全输入绑定 2026-10-15，十项 SQL、1,316 条六字段、实际 Sleep 1、Today 52 / 前日 74、MCP 最高排名及实际 21 事件、Astrocyte 134 条分页和两主题各三历史点。这里只登记 Q 交付值及现有指针，没有重新运行 Q 验收、读取原始会话或采集/Sleep。P helper 按输入数量与实际同代动态核查，没有固定旧 1,339 或旧 generation；旧候选结果继续仅属于下面的历史章节。若实际验收跨日，现有日期检查会拒绝过期期望，交原 Q owner 提供准确日期的同代 SQL 期望，不借旧样本通过。

实际剩余：H/B 的准确完整生产配置及启动参数接入；S5 完整 Goal/TasteProfile AT12 和生产 HTTP；AOCI FSL-1.1-MIT 的用户用途/分发决定；OS 执行授权及 Jev key；S4 较长 Windows 投影路径限制、SemanticIndex 未决定与其他原域缺项；唯一 I、完整默认检查、真实原生/Today、同源码新包和实际核心路径验收。AOCI 没有进入资产或 ZIP，显式合法外置配置仍为已提交方案；Jev 仅固定 `jev-1.13.0` 云 REST，缺 key 的真实 API 为 NOT_RUN。未发布 tag/Release，没有人工代测要求，领域失败均通过协调者交原 owner。独立可消费准备阶段不等待只能在最终组合上执行的检查。

## 2026-10-14 候选包历史

当日发布 owner 为 `release-package-20261014`，分支相同，输出在自己的 `.verification`。当时 Orca Codex 终端 `term_e67c7493-cef5-439e-9687-3197880820d7` 公开运行信息显示 `GPT-6.1-Sol high fast`。未读取私人凭据、原会话、正式数据库，未使用视觉功能，未发布 tag 或 GitHub Release。

## 发布工具 SOURCE

最新验证工具 SOURCE `1a1c14d6fdd02e7824efb82a618848a1c315dedb` 已普通 push，远端同 SHA；包的精确 SOURCE 为 `15f27f4a8e1e2300b8b1152cfd5b8e5f722db744`。此前普通消费 Q 最终 SOURCE `1d07c5b2ff78db5ffe7e90b4d903c618b5a8ac50` / REPORT `6ad1dfa1387d5e6a0b81ac1b2ad2887359780658`、B SOURCE `e914ff9bb845d1e7b3ab13d1b1721e1598bafbe7` / REPORT `15c738f88822e553eb228c12c0fdc71ad07b7531`、F SOURCE `205d2d2ce91091e6add67d637cb65aae33946292`、H SOURCE `b247dfb12b938739f09acc73e6b7582376951a0c` / REPORT `a8334b79aeb0c3a1b1038d45964d6092dc6cc5de`。包含精确 Git 归档、生产构建、标准 ZIP、启动与手动快照脚本、真实运行验证和用户说明。当前候选构建及运行 124 项检查通过；此前全部普通追加历史和首次失败证据保留。

## 今晚候选包阶段交付

主控传达用户新授权：今晚交付已通过的候选包供试用，阶段 REPORT 完成后结算；今晚停止新增 Relay 注册、派发和模型调用，真实 CLI → 三分钟 Today → L1 验收安排在 2026-10-15。当前阶段完成，不计为唯一集成人员 I 的最终源码验收。明日沿用当前 worktree/branch，收到 I 精确 SOURCE 后同源码重建、重新解压并验证。

实际构建命令，EXIT 0：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Build-Windows.ps1 -Source 15f27f4a8e1e2300b8b1152cfd5b8e5f722db744 -Label cognition-paired
```

Go 1.27.2、Node 24.16.0、pnpm 11.27.0；Windows amd64、CGO_ENABLED=0，GOFLAGS=-p=1、GOMAXPROCS=2、NODE_OPTIONS=--max-old-space-size=512。冻结锁安装、原 `pnpm build` 和 Q Go 快照程序编译完成。Vite 514.47 kB chunk 大小警告保留。ZIP 共 39 项，四个 Go 程序、生产静态资源、启动与手动快照脚本、README、LICENSE；数据库、日志和凭据名检查为零。

ZIP 绝对路径：`C:/Users/DW/orca/workspaces/Astrocyte/release-package-20261014/.verification/package-cognition-paired-20261014-231405/Astrocyte-v0.1-windows-x64-15f27f4a8e1e.zip`，**33,768,080 字节**。构建原件 `.verification/package-cognition-paired-20261014-231405/build.log`。

实际完整运行验证命令，Windows PowerShell 5.1，默认隔离端口 17887，EXIT 0：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Verify-Windows.ps1 -Zip 'C:/Users/DW/orca/workspaces/Astrocyte/release-package-20261014/.verification/package-cognition-paired-20261014-231405/Astrocyte-v0.1-windows-x64-15f27f4a8e1e.zip' -RequireCombined -Source 15f27f4a8e1e2300b8b1152cfd5b8e5f722db744 -CognitionSourceSha 1d07c5b2ff78db5ffe7e90b4d903c618b5a8ac50 -CognitionSource 'C:/Users/DW/orca/workspaces/Astrocyte/release-cognition-20261014/.verification/public-cognition-6f63200932f14a10ad4d9fb91c78d749/app-state/state.sqlite' -CognitionHistory 'C:/Users/DW/orca/workspaces/Astrocyte/release-cognition-20261014/.verification/public-cognition-6f63200932f14a10ad4d9fb91c78d749/local/astrocyte-workbench/v0.1/snapshots/snapshot-892856014/history.json' -CognitionPublicationRoot 'C:/Users/DW/orca/workspaces/Astrocyte/release-cognition-20261014/.verification/public-cognition-6f63200932f14a10ad4d9fb91c78d749/local/astrocyte-workbench/v0.1' -CognitionSafeDir 'C:/Users/DW/orca/workspaces/Astrocyte/release-cognition-20261014/.verification/cognition-latest-safe'
```

124/124 PASS，scope `candidate-package-cognition-runtime`。新目录为 `.verification/verify-20261014-231846/extracted/Astrocyte-v0.1-windows-x64`；空身份目录、系统 PATH 下实际启动，无开发源码或 node_modules 运行依赖。生产私有路径扫描零匹配。真实健康、九个静态页面路径与资源、session bootstrap、无 session/无 CSRF/错误 Origin 的拒绝、实际分类创建 HTTP 201 和读取均通过。未知认知快照实际返回 503；空 Relay 注册、任务、派发和会话事件保持真实空状态。

先对应用自己创建的真实 SQLite 手动发布空认知代 `snapshot-856439225`，再对 Q 公开自产 SQLite 和真实 History 手动发布非空代 `snapshot-1764716540`；两次均由解压包的实际快照程序执行。健康、Today、Focus、Agents、Topics、Evidence、Filters 和 Events 实际 HTTP 读取与同代时间头一致。十项独立 SQL 数值、最高 MCP 排名、两主题每个历史点的时间和 attention、全部 1339 个事件经 14 页逐一核对六字段及字段集合、Astro134 的 20/50 分页均通过。真实进程终止与重启后新 session、原分类和认知数值保留，最终两次 owned listener 均关闭；当前没有保留应用 PID 或活动 URL。

安全结果：`C:/Users/DW/orca/workspaces/Astrocyte/release-package-20261014/.verification/verify-20261014-231846/result.json` 与同目录 `private-path-scan.json`；原件 `verify.log`、`initial-stdout.log`、`initial-stderr.log`、`restart-stdout.log`、`restart-stderr.log` 均保留在该目录。M-release 已独立核查 result 124 PASS、scan 零匹配与 ZIP 项目数。验收 SQLite 和真实认知输入没有进入 ZIP。

同一完整验证命令首次在 `.verification/verify-20261014-231729/verify.log` EXIT 1：`FAIL: actual SQL10 and nonempty six-field source are complete`。Windows PowerShell 5.1 的 `@(ConvertFrom-Json)` 将 1339 项包装为一项嵌套数组；错误发生在输入校验，应用尚未启动。直接接收标准 JSON 转换结果后，在上述新目录实际通过；修复 SOURCE 为 `1a1c14d6fdd02e7824efb82a618848a1c315dedb`，只改未进入 ZIP 的验证 helper。包仍严格记录 SOURCE 15f；没有为测试修改包、跳过扫描或覆盖首失败。

基本用户运行条件为 Windows x64、Windows PowerShell 5.1 和浏览器，完整解压后执行 `Start-Astrocyte.cmd`。可选 CLI、模型及配置参见 [用户说明](windows.md)。今晚模型调用零；默认浏览器全套、真实原生 Relay/Today/L1、媒体提取仍未验收，I 最终 SOURCE 重建、tag 和 GitHub Releases 发布未执行。root 在全部最终验收后统一发布。独立 SOURCE 与本阶段 REPORT 分别追加 commit 和普通 push。

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

da45 阶段验证工具将 `-RequireCombined` 静态路径设置为现有产品的 `/cognition/{today,focus,agents,events}` 和 `/workspace/works`，并增加空 Relay 注册、空任务、零派发、零会话事件与未知覆盖状态的具体断言。候选模式的四个基本 SPA 路径、其余真实运行路径在同一候选 ZIP、另一新目录 `.verification/verify-20261014-203557` 复验 EXIT 0，24 项 PASS。该次仅检查基本候选；组合认知与空 Relay 断言已在今晚 15f 阶段包实际运行通过。此工具变更不改变初始候选包的三个程序和用户启动脚本。

## 首次失败与修复

- Windows PowerShell 5.1 首次加载 UTF-8 无 BOM 的中文注释脚本，报告 `MissingCatchOrFinally` / `UnexpectedToken`，未进入构建。三份脚本增加 BOM 后，5.1 实际 parser 与执行通过。
- 首次以 `$env:OS` 判定宿主，当前 Orca 进程没有该环境变量，报告 `Windows x64 build host required`。改为系统 API 的 `Win32NT` 与 64 位操作系统判定后通过。
- 首次 `git archive --output=(...)` 在 PowerShell 原生命令参数处理中将 ZIP 路径当作对象名，报告 `not a valid object name` / `git archive failed`。证据保留在 `.verification/package-rehearsal-20261014-202917/build.log`；改为一个完整 `--output=` 参数，另建演练目录通过。

以上失败继续保留，后续通过不覆盖失败记录。未自动重发任何模型调用。

## 可选能力与当前限制

基本包实际启动无需开发源码和 node_modules。默认关闭 summarize 和 Codex 自动沉淀，只运行应用；可选 summarize 0.25.1、Node、yt-dlp、FFmpeg、Whisper 二进制/模型及已安装 Python 的绝对路径参数见 [用户说明](windows.md)。MCP 和 Sleep 沿用原程序与原未确定参数，启动脚本不调用、不安装正式日程。

Workbench 原 `scripts/workbench/snapshot.py` 的 main 使用硬编码正式 SOURCE/HISTORY，未纳入本包、未执行。该入口缺口已交 Q owner 补充；Q 早期 SOURCE `8001a76646f5892bc18d8ccb9e5cd065b51425a6` 提供 Go 原生的显式 source/root/history/timeout 命令。初始候选 ZIP 仍未包含该程序；没有快照时的 503 只证明不可用状态如实返回。

认知公开自产来源已由 Q 交付并完成本阶段候选验证；最终合并 SOURCE 仍待唯一集成人员交付。接收其精确 SOURCE 后普通消费、重新构建、新目录解压并实际验证，随后追加独立 REPORT。完整默认 E2E、原生 event_id 闭环、最终发布与人工待定参数继续保留未完成状态；正式 V4 数据、协议和日程保持原状态。

## Q 程序与手动包装预验

P 追加 SOURCE `d0fb7ff0bb1e791583cc405f098eb684affce35e` 已 push，增加 `workbench-snapshot.exe` 构建和手动 `Publish-WorkbenchSnapshot.ps1`，以及实际快照健康与重启检查。Q SOURCE `8001a76646f5892bc18d8ccb9e5cd065b51425a6` 普通合并首次遇到唯一 `HANDOFF.md` 冲突；该文件超出 P 写域，原现场保留并交根文档 owner。模块负责人获 root 单次文档授权后用文件编辑工具解决并仅 stage 该文件，P 完成普通 merge `4b15b70`。未回滚 Git，未编辑自动合并的业务代码。第一次 push 遇 `OpenSSL SSL_connect: SSL_ERROR_SYSCALL`，普通重试成功且远端 SHA 已核对一致。

独立预验只使用 Q 精确源码归档：`git archive --format=zip --output=<自身.verification/q-snapshot-smoke/source.zip> 8001a76646f5892bc18d8ccb9e5cd065b51425a6`，标准解压后执行 `go build -mod=readonly -trimpath -o <自身.verification/q-snapshot-smoke/workbench-snapshot.exe> ./cmd/workbench-snapshot`，EXIT 0。复制当前 P 手动脚本到同一预验输出目录；来源为此前真实候选应用创建并经实际重启的 `.verification/verify-20261014-203557/extracted/Astrocyte-v0.1-windows-x64/data/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite`。

`powershell.exe -NoProfile -ExecutionPolicy Bypass -File <自身.verification/q-snapshot-smoke/Publish-WorkbenchSnapshot.ps1> -Source <上述实际SQLite>` 首次失败：Windows PowerShell 5.1 参数默认表达式内的 `PSScriptRoot` 为空，`Join-Path` 报 `ParameterArgumentValidationErrorEmptyStringNotAllowed`。显式 `-DataDir <自身.verification/q-snapshot-smoke/data>` 运行 EXIT 0，真实 generation `snapshot-3747436004`；随后将默认目录计算移到脚本体，省略 `-DataDir` 的相同实际命令 EXIT 0，真实 generation `snapshot-3863927605`。默认目录修复同时用于用户启动脚本，已独立提交并 push 为 `27e65dd80436030b0157edb6f811c83320dda001`。

预验没有调用 Python、模型或正式自动任务，没有构造认知事件。来源为真实应用空认知库；此结果只证明 Go 程序与手动包装可实际执行，不替代最终包、真实认知数值或六字段验收。该次预验未使用后来 Q 交付的公开受控来源和安全期望。

用户 CMD 启动入口增加原错误退出码传递：实际 `cmd.exe /d /c 'release\Start-Astrocyte.cmd -Port 0 <NUL'` 返回预期 EXIT 1 与端口范围错误。该负例不启动应用。

## 组合构建首次失败与路径检查

实际执行 `powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Build-Windows.ps1 -Source 27e65dd80436030b0157edb6f811c83320dda001 -Label q-rehearsal`，EXIT 1。冻结锁安装与三个 Go 程序编译完成，原 `pnpm build` 的 TypeScript 阶段报 `web/e2e/relay-runtime.spec.ts(6,10): TS2305`：`worksAlbumPresentation` 没有 `fixtureWorksSnapshot` 导出。未产出新 ZIP，未绕过 TypeScript 或删减检查，原前端 owner 已收到 Handoff。首原件 `.verification/package-q-rehearsal-20261014-210111/build.log` 保留。

生产验证新增 .NET Regex 字符串扫描，对当前用户目录和已知正式源私有常量只输出相对文件与计数，不输出匹配内容。初始 c7a70be ZIP 在另一新目录 `.verification/verify-20261014-210551` 实际扫描零匹配并完成基本运行 25 项 PASS，结果和 scan 聚合记录在该目录；该旧包没有 Workbench 组合，结果不替代新组合。对上述组合已编译 server 的独立只读 `rg --text --count-matches --with-filename` 扫描，安全输出 `astrocyte.exe:3`，确认组合存在私有路径/正式源常量，已交业务 owner 修复。新包最终必须通过生产扫描。

Q 加强保护 SOURCE `88bab5fe524e55d40b3799de0ff5fc6d45d08f6b` 已普通消费为当前阶段 SOURCE。同一精确 Q 归档在 `.verification/q-snapshot-smoke-88bab5fe` 运行 `go build -mod=readonly -trimpath` EXIT 0，再用该 Go 程序和本 P 手动脚本对上述实际候选 SQLite 发布，EXIT 0、generation `snapshot-2185978382`。未重复已失败的完整组合构建；后续候选等待前端精确修复，最终包等待唯一 I 的精确 SOURCE。

## F 修复后新候选与隐私首失败

F 精确修复已普通消费为 `4fd19d83d94d0d8867864d4352c5d07a54ce9986`，P 追加明确的 uvx、既有 Relay 来源模式和用户显式活性窗口启动参数，默认窗口继续未配置；监听检查改用 .NET IPGlobalProperties，网络查询错误立即失败。最新基本验证脚本对初始 c7 ZIP 在 `.verification/verify-20261014-211602` 新目录复验 25 项 PASS；Windows PowerShell 5.1 parser PASS。新的可选参数不启用模型或扩大工具权限。

实际 `powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Build-Windows.ps1 -Source c54bcf35162918dab1c55cfa2f0ba6d97ce931ed -Label fq-rehearsal` EXIT 0：冻结锁安装、原 Go/TypeScript/Vite 生产构建、Q Go 快照程序均完成。Vite 保留 500 kB chunk 大小警告。ZIP 为 `.verification/package-fq-rehearsal-20261014-211834/Astrocyte-v0.1-windows-x64-c54bcf351629.zip`，33,749,040 字节。

实际 `powershell.exe -NoProfile -ExecutionPolicy Bypass -File release/Verify-Windows.ps1 -Zip <上述ZIP> -RequireCombined` 在 `.verification/verify-20261014-212203` 失败，未启动应用：生产 `astrocyte.exe` 的 UTF-8 字符串扫描匹配 3 处当前个人目录或已知正式源常量。安全记录只输出文件名和计数；首原件 `verify.log`、`private-path-scan.json` 保留。未手工删除二进制字符串，未跳过扫描，问题已交 B/H；后续必须消费其精确修复并同源码重建。该 ZIP 不可作为最终发布包。

## 明确的真实认知与原生验证输入

Q 最终 SOURCE/REPORT 已普通消费。Q 明确交付的 latest 输入是独立 `public-cognition-6f63200932f14a10ad4d9fb91c78d749/app-state/state.sqlite`、`local/astrocyte-workbench/v0.1/current.json` 的实际 `snapshot-892856014` 与该代的真实 `history.json`，均位于 Q 自己的 `.verification`。安全期望位于 Q `.verification/cognition-latest-safe`；独立读其 `admission-safe.json` / `sql-expected.json` 与 current 指针，已确认 `after_sleep`、实际 Sleep 1、原代时间一致、SQL10、1339 六字段、最高 MCP20 与分页 Astro134、两主题各3真实历史。没有按含有 final 的旧目录名选择输入，没有读取正式 V4 或用户原会话。

P 验证逻辑要求显式传入上述来源、历史、发布根、安全期望目录和完整 package/Q SOURCE。校验 Q 源码祖先、包文件名、当前发布代与 after_sleep，随后实际包快照命令发布、HTTP SQL10、全部事件六字段、最高排序、每个实际历史点、20/50 同主题分页及重启。Windows PowerShell 5.1 parser 与 15f 候选包实际正向运行均通过，具体命令和阶段范围记录在本报告的今晚交付段落。

真实 packaged native 正向还需要原 H 的固定公开任务、真实 CLI/自行鉴权、明确窗口、真实 API 注册/授权/确认与每个 event_id 的 SQLite 核查输入。仅准备脚本，P 尚未启动原生模型。R 已明确 SessionDeadlineSeconds 是可信组装参数，当前不属于环境变量，P 没有发明此参数。候选的空 registry / 空会话事件与认知健康不替代实际原生正向；最终仍须唯一 I 的精确 SOURCE 与全部原验收。
