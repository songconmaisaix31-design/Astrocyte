# N 客户端非模型阶段验收记录

2026-10-14，Task `task_bfc0bedf1c26`，Dispatch `ctx_44647082a057`，发布模块 Run `run_ab36c13bc549`。

用户最新时序先修改 v1.4 的投递阶段、原生会话归属、输入就绪、恢复身份，2026-10-15 再运行真实 CLI、三分钟 Today 和 L1 结算。协调者已通过本 Dispatch 的原问题回复确认：本阶段按客户端准备与已实际完成的早期基础 HTTP 交付结算，今晚停止新的 Relay 注册、派发、模型和 Today 验收。发布模块完整任务仍在进行；本文没有将最终包或原生验收登记为完成。

## SOURCE 与修改范围

- 分支：`songconmaisaix31-design/release-native-verification-20261014`。
- 精确 SOURCE：`7202f07c9c4122f08d760afa88472fb32852b38e`，普通 push 成功。
- 本树起点：`c54bcf35162918dab1c55cfa2f0ba6d97ce931ed`。
- 实际早期 HTTP 修复与运行对应客户端 SOURCE：`0b69dc0049d5e76cebe4f078a330d50a4bbea255`。后续 SOURCE 保留该 request 实现，增加原件失败安全计数、本机数据库指针、保留端口及获准阶段说明。
- 代码与调用说明仅位于 `tests/release-package-native/**`、`docs/release-package-native.md`，所有产物和原始响应位于已忽略的本树 `.verification`。P 的 `release/**`、认知客户端、业务、依赖、配置、正式 V4 运行树和日程均未修改。

客户端具备真实 bootstrap、单次 HTTP 请求、空注册/无派发检查、单一任务状态读取、分页六字段核查的实现；`events` 代码通过实际 `mode=ro/query_only` 连接选择 owned binding，复用现有 JSONL 读取器，逐 event_id 对比实际 SQL、HTTP 和选定 archive 观察范围。原件守卫、未知计数及覆盖范围保持现有规则。调用的实际进程文件、端口监听者和 ZIP 内 exe 字节由入口核查。

## 已运行命令和结果

| 检查 | 实际结果 |
|---|---|
| `GOMAXPROCS=2; go build -p=1 -mod=readonly -o .verification/native-client.exe ./tests/release-package-native` | PASS，退出 0 |
| `GOMAXPROCS=2; go vet -p=1 -mod=readonly ./tests/release-package-native` | PASS，退出 0 |
| `git diff --check` | PASS，退出 0 |
| 实际 Windows PowerShell `5.1.26100.9711` 解析两个 `.ps1` | PASS，退出 0；脚本源码 UTF-8 BOM |
| 实际 PowerShell 5.1 调用 `Verify-BasicClient.ps1`，指定下述早期 ZIP/SOURCE/17896 | PASS，退出 0；调用原包 `Start-Astrocyte.ps1` 和 N 客户端 |

从仓库根目录运行的等价命令为 `powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File tests/release-package-native/Verify-BasicClient.ps1 -Zip '..\release-package-20261014\.verification\package-rehearsal-20261014-202942\Astrocyte-v0.1-windows-x64-c7a70be963c4.zip' -PackageSource c7a70be963c432bb72d0fa355705ce2907821120 -Port 17896`。实际执行的 `Zip` 参数使用该文件的绝对路径，原文保留在本机安全汇总。

## 实际包与安全聚合

- 包 SOURCE：`c7a70be963c432bb72d0fa355705ce2907821120`，由协调者指定为早期基础包。
- 实际 ZIP：P 工作树 `.verification/package-rehearsal-20261014-202942/Astrocyte-v0.1-windows-x64-c7a70be963c4.zip`，26,309,211 bytes。
- 实际运行 exe：本树 `.verification/basic-http-20261014-223940/extracted/Astrocyte-v0.1-windows-x64/astrocyte.exe`。
- 成功轮次实际服务 PID：41464，HTTP 为 `http://127.0.0.1:17896`；进程文件与监听者核查通过，exe 与 ZIP 成员逐字节一致。
- 已知私人路径扫描：匹配文件 0、匹配数 0。范围为当前用户目录和已知正式来源常量；该结果不构成最终组合包隐私验收。
- 成功轮次 3 个 GET：health 200、human bootstrap 200、受保护 local-projects 200。真实 Cookie 和返回 CSRF token 的读取通过；没有执行 API 写入验证。
- 全阶段写入请求 0、原生模型调用 0、Agent 注册动作 0、派发动作 0。
- 三轮 owned 服务进程和端口均已关闭。未读取、复制或建立 H 新旧原档的链接，未读取 H 数据库、旧会话或 CLI 鉴权配置。

包 SOURCE 是 P 的构建来源声明。客户端独立证明实际 exe 和 ZIP 成员一致，没有从 exe 内容独立推导 Git SHA。

## 本机证据

成功轮次 `.verification/basic-http-20261014-223940/basic-safe.json`、`protected-read/summary.json`。原始 bootstrap、API 响应、server 日志位于该轮次本机目录，未写入 stdout、Git 或包。

本任务要求保留的首失败分别位于 `.verification/basic-http-20261014-223748` 和 `.verification/basic-http-20261014-223837` 的安全汇总与原件。失败代码分别为 `zip_executable_count`、`human_bootstrap`，两轮 owned 进程与端口均已关闭，没有模型调用或写入请求。后续真实运行通过，没有跳过成功断言。

SOURCE `317b7efa814403e28fea47b6e33e7daf17d68639` 首次普通 push 的 TLS `SSL_ERROR_SYSCALL` 保留；随后同一普通 push 成功。没有 force push、rebase、reset 或公共历史覆盖。

## NOT_RUN 与后继输入

最终 I/P 新组合 ZIP、同 URL 的 Relay baseline/registry、实际 native intent/confirm、每 event_id 的 SQL/HTTP/原件核查、原生停止及重启、三分钟 Today、DOM 均为 `NOT_RUN`。本树没有运行 `events` 或 `task-state` 模式；该早期包不含 Relay。客户端实现和编译结果没有代替这些实际验收。

目前 `events` 实现要求实际 Codex archive；`native_notifications` 的原件范围比较尚未实现，相关调用会失败。当前 task/receipt 语义属于已读接口记录，v1.4 最终 SOURCE 到达后须重新核查。完整 SessionComplete、ArchiveAppendOnly、PrefixComplete 或原生全部工具关闭尚未由 N 验收，保持未知和未执行。

明日同原 terminal/worktree/branch 用新的 Task/Dispatch 续验需要：v1.4 精确最终 API/source；P fresh ZIP/完整 package SOURCE/真实 exe；独立 N 数据、公开 Git、端口与输出目录；H/F 原窗口结束的实际状态；固定公开任务与真实 CLI argv；官方 CLI 自行鉴权且零复制；文件、命令、MCP、skills、视觉关闭的实际证据；明确 live 秒数；实际 session source 与单一 owned 原件范围、每 event_id 的安全 SQL locator；适用停止/重启与无自动派发的验收窗口。SessionDeadline 等配置仅按最终源码支持字段消费，当前没有自行增加 ENV。

P 的新非 Relay 认知窗口由 P 先验。本阶段没有执行完整仓库构建、默认浏览器、视觉或正式 V4 检查；完整产品的检查与发布验收继续由发布模块负责。
