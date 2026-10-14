# Windows x64 发布包记录

发布 owner：`release-package-20261014`，分支 `songconmaisaix31-design/release-package-20261014`。只修改 `release/**` 和 `docs/release/**`，本机输出在自己的 `.verification`。实际执行客户端为 Orca Codex；当前终端 `term_e67c7493-cef5-439e-9687-3197880820d7` 的公开运行信息显示 `GPT-6.1-Sol high fast`。未读取私人凭据、原会话、正式数据库，未使用视觉功能，未发布 tag 或 GitHub Release。

## 发布工具 SOURCE

最新阶段 SOURCE `830faaea4f5e19bdf213079f428dc9e0c6995d2e` 已 push，普通消费 F 修复 `9ccb198fd1af1bd6ceb2d98b25b0531802543983`、Q 最终 SOURCE `1d07c5b2ff78db5ffe7e90b4d903c618b5a8ac50` 与 REPORT `6ad1dfa1387d5e6a0b81ac1b2ad2887359780658`。包括精确 Git 源码归档、复用生产构建、标准 ZIP、用户启动与手动快照脚本、真实 HTTP/session/CSRF/SQLite/重启验证、生产路径扫描、明确输入的真实认知核查及用户说明。TS2305 修复后的 c54 候选构建通过，但解压私有路径扫描失败，尚未通过最终验收。此前全部普通追加历史保留。

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

最新验证工具将 `-RequireCombined` 静态路径设置为现有产品的 `/cognition/{today,focus,agents,events}` 和 `/workspace/works`，并增加空 Relay 注册、空任务、零派发、零会话事件与未知覆盖状态的具体断言。候选模式的四个基本 SPA 路径、其余真实运行路径在同一候选 ZIP、另一新目录 `.verification/verify-20261014-203557` 复验 EXIT 0，24 项 PASS；认知和 Relay 新断言仍待最终合并源码运行。此工具变更不改变候选包的三个程序和用户启动脚本。

## 首次失败与修复

- Windows PowerShell 5.1 首次加载 UTF-8 无 BOM 的中文注释脚本，报告 `MissingCatchOrFinally` / `UnexpectedToken`，未进入构建。三份脚本增加 BOM 后，5.1 实际 parser 与执行通过。
- 首次以 `$env:OS` 判定宿主，当前 Orca 进程没有该环境变量，报告 `Windows x64 build host required`。改为系统 API 的 `Win32NT` 与 64 位操作系统判定后通过。
- 首次 `git archive --output=(...)` 在 PowerShell 原生命令参数处理中将 ZIP 路径当作对象名，报告 `not a valid object name` / `git archive failed`。证据保留在 `.verification/package-rehearsal-20261014-202917/build.log`；改为一个完整 `--output=` 参数，另建演练目录通过。

以上失败继续保留，后续通过不覆盖失败记录。未自动重发任何模型调用。

## 可选能力与当前限制

基本包实际启动无需开发源码和 node_modules。默认关闭 summarize 和 Codex 自动沉淀，只运行应用；可选 summarize 0.25.1、Node、yt-dlp、FFmpeg、Whisper 二进制/模型及已安装 Python 的绝对路径参数见 [用户说明](windows.md)。MCP 和 Sleep 沿用原程序与原未确定参数，启动脚本不调用、不安装正式日程。

Workbench 原 `scripts/workbench/snapshot.py` 的 main 使用硬编码正式 SOURCE/HISTORY，未纳入本包、未执行。该入口缺口已交 Q owner 补充；Q 早期 SOURCE `8001a76646f5892bc18d8ccb9e5cd065b51425a6` 提供 Go 原生的显式 source/root/history/timeout 命令。初始候选 ZIP 仍未包含该程序；没有快照时的 503 只证明不可用状态如实返回。

最终合并源码、认知公开自产来源、Relay 注册和真实 CLI 输入待唯一集成人员交付。接收其精确 SOURCE 后普通消费、重新构建、新目录解压并实际验证，随后追加独立 REPORT。完整默认 E2E、原生 event_id 闭环、最终发布与人工待定参数尚未因本候选包演练成为通过；正式 V4 数据、协议和日程保持原状态。

## Q 程序与手动包装预验

P 追加 SOURCE `d0fb7ff0bb1e791583cc405f098eb684affce35e` 已 push，增加 `workbench-snapshot.exe` 构建和手动 `Publish-WorkbenchSnapshot.ps1`，以及实际快照健康与重启检查。Q SOURCE `8001a76646f5892bc18d8ccb9e5cd065b51425a6` 普通合并首次遇到唯一 `HANDOFF.md` 冲突；该文件超出 P 写域，原现场保留并交根文档 owner。模块负责人获 root 单次文档授权后用文件编辑工具解决并仅 stage 该文件，P 完成普通 merge `4b15b70`。未回滚 Git，未编辑自动合并的业务代码。第一次 push 遇 `OpenSSL SSL_connect: SSL_ERROR_SYSCALL`，普通重试成功且远端 SHA 已核对一致。

独立预验只使用 Q 精确源码归档：`git archive --format=zip --output=<自身.verification/q-snapshot-smoke/source.zip> 8001a76646f5892bc18d8ccb9e5cd065b51425a6`，标准解压后执行 `go build -mod=readonly -trimpath -o <自身.verification/q-snapshot-smoke/workbench-snapshot.exe> ./cmd/workbench-snapshot`，EXIT 0。复制当前 P 手动脚本到同一预验输出目录；来源为此前真实候选应用创建并经实际重启的 `.verification/verify-20261014-203557/extracted/Astrocyte-v0.1-windows-x64/data/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite`。

`powershell.exe -NoProfile -ExecutionPolicy Bypass -File <自身.verification/q-snapshot-smoke/Publish-WorkbenchSnapshot.ps1> -Source <上述实际SQLite>` 首次失败：Windows PowerShell 5.1 参数默认表达式内的 `PSScriptRoot` 为空，`Join-Path` 报 `ParameterArgumentValidationErrorEmptyStringNotAllowed`。显式 `-DataDir <自身.verification/q-snapshot-smoke/data>` 运行 EXIT 0，真实 generation `snapshot-3747436004`；随后将默认目录计算移到脚本体，省略 `-DataDir` 的相同实际命令 EXIT 0，真实 generation `snapshot-3863927605`。默认目录修复同时用于用户启动脚本，已独立提交并 push 为 `27e65dd80436030b0157edb6f811c83320dda001`。

预验没有调用 Python、模型或正式自动任务，没有构造认知事件。来源为真实应用空认知库；此结果只证明 Go 程序与手动包装可实际执行，不替代最终包、真实认知数值或六字段验收。Q 的公开受控来源和安全期望仍待交付。

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

新 P 验证逻辑要求显式传入上述来源、历史、发布根、安全期望目录和完整 package/Q SOURCE。校验 Q 源码祖先、包文件名、当前发布代与 after_sleep，随后实际包快照命令发布、HTTP SQL10、全部事件六字段、最高排序、每个实际历史点、20/50 同主题分页及重启。Windows PowerShell 5.1 parser 通过；此正向逻辑尚未在通过隐私扫描的新包上执行，不能计为通过。

真实 packaged native 正向还需要原 H 的固定公开任务、真实 CLI/自行鉴权、明确窗口、真实 API 注册/授权/确认与每个 event_id 的 SQLite 核查输入。仅准备脚本，P 尚未启动原生模型。R 已明确 SessionDeadlineSeconds 是可信组装参数，当前不属于环境变量，P 没有发明此参数。候选的空 registry / 空会话事件与认知健康不替代实际原生正向；最终仍须唯一 I 的精确 SOURCE 与全部原验收。
