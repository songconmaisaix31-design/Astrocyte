# Workbench v0.1 精确集成与静态部署

## 完成内容与来源

本轨复用协调者已经建立的独立 Orca worktree `workbench-v01-integration`，起点为 root `438f50ec318888badcab6febf23c0ca71851afe2`，来自 `workbench/v0.1-workbench`。交付分支 `songconmaisaix31-design/workbench-v01-integration`；客户端 Codex，未创建其他 Worker。

依次普通 `git merge --no-ff` 两个精确 REPORT，无冲突，无装配胶水或领域变更：

| 来源 | SOURCE | REPORT |
| --- | --- | --- |
| backend | `8635895e40eb022e92a86870861b4d09fd168147` | `51a2102f566718280669eda774d8bb442fb2f1c9` |
| frontend | `8ec40f7633236b80bf12e34aa6468633f96d4072` | `ca1bcde736706e1f0ce8a80c6ece65ff2bd9fea1` |
| 本次集成 | `c5da66724d22455fbbf376c2516f91c4bef63d5c` | 本文件的独立提交；精确 SHA 随 worker_done 回传 |

backend 合并提交为 `28942f77d08ebdf0f0a5ec7cbfb6f1ab9c70806d`，是最终 SOURCE 第一父提交。最终 SOURCE 已先行 push 并 Handoff root 执行整树最终检查。本 REPORT 只新增本文件；业务、配置、测试与 SOURCE 完全相同。旧独立工程、移动未知 HEAD 和未提交文件均未消费。

## 验证命令与结果

| 命令或检查 | 结果 |
| --- | --- |
| `git merge-base --is-ancestor <SHA> c5da667` | PASS：root、backend SOURCE/REPORT、frontend SOURCE/REPORT 五个精确祖先 |
| `git diff --name-status <owner SOURCE> <owner REPORT>` | PASS：分别只新增 backend.md、frontend.md |
| `git diff --exit-code <backend SOURCE> c5da667 -- <backend cmd/contracts/internal/scripts 变更路径>` | PASS，后端源码路径完全相等 |
| `git diff --exit-code <frontend SOURCE> c5da667 -- <frontend web 变更路径>` | PASS，前端源码路径完全相等 |
| `git diff --name-only <frontend SOURCE> c5da667` | 仅两份 owner 报告与 root 原有五份协调文档不同；所有业务、测试、依赖、配置完全相等 |
| `git diff --exit-code 438f50e c5da667 -- HANDOFF.md STATUS.md tasks/Workbench-api.md tasks/Workbench-integration.md tasks/Workbench-v01.md` | PASS，root 文档原样保留 |
| `git diff --name-only 438f50e c5da667` 与两轨源码路径及两份报告集合比较 | PASS，无额外路径；引擎/events/Sleep/schema migrations/参数/依赖/旧 spec/Playwright 配置冻结域零变更 |
| `git diff --check 438f50e c5da667` / `git diff --check` | PASS |
| `pnpm install --frozen-lockfile` | PASS，pnpm 11.27.0，锁文件未改 |
| `node scripts/generate-client.mjs --check` | PASS，schema.d.ts 零漂移 |
| `pnpm --dir web build` | PASS，TypeScript 与 Vite；保留既有 513.09 kB chunk 提示 |
| `powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/workbench/start.ps1 -Port 17321 -SkipSnapshot -WebDir <本集成 web/dist>` | PASS，构建本集成 server、Hidden 启动、session/认知 health 与内置任务 Ready |
| `node "$env:LOCALAPPDATA/astrocyte-workbench/v0.1/verification/integration/verify-static.cjs"` | PASS，实际静态页与两个资源字节等于本集成 dist；1920×1080 四子页直接加载、正确选中 tab、无整页横向溢出 |
| 同一静态浏览器的 session 与认知请求 | PASS：初始无 session 认知 health=403；原 auth/session=200；共享 session 后原 health=200、认知 health=200；自动 bootstrap=200；认知请求仅 GET，失败请求和 pageerror 均零 |
| owned PID/ExecutablePath/监听/config.web/快照指针复核 | PASS，部署结果见下；current.json 部署前后原文完全相等 |
| `Get-ScheduledTask` / `Get-ScheduledTaskInfo` / `Get-TimeZone` | PASS，Snapshot/Start 均 Ready；China Standard Time；Snapshot 每日 00:30，下一次 2026-10-15 00:30 +08:00 |

本机复核脚本与派生摘要只在 `%LOCALAPPDATA%/astrocyte-workbench/v0.1/verification/integration`：`verify-static.cjs`、`static-verification.json`、`deployment-summary.json`、`pointer-before.json`。未发布 cookie、session token、配置、数据库、原始事件、截图或 UI 数据到 Git。

## 部署现场

地址：`http://127.0.0.1:17321/cognition/today`。核实时 owned PID=`42472`，ExecutablePath 与 `%LOCALAPPDATA%/astrocyte-workbench/v0.1/bin/astrocyte-workbench.exe` 完全相等；17321 监听属于该 PID，config.web 完全匹配本工作树 `web/dist`。

旧 server.pid 记录 `15652`，但部署前两次核查均确认 PID 已不存在，同 ownedexe 无实例且 17321 无监听。因此本轨停止进程数为零，未停止个人/V4/其他进程；没有猜测杀进程。

保持 `snapshot-hxfipcpr`，snapshot_at=`2026-10-14T01:24:01.253839+08:00`，与 `2026-10-13T17:24:01.253839Z` 同一时刻。部署使用 `-SkipSnapshot`，未执行 refresh，未读取 V4 正式源。Windows Snapshot 和登录 Start 任务重装后均 Ready，最近结果均为 0；登录任务没有 NextRunTime 属正常无固定下一次时间，未推断为失败或跨日通过。

## 首次诊断与真实限制

- 首次整业务树比较脚本把 root 原有 HANDOFF/STATUS/三个任务文档差异误报为业务漂移并退出；之后明确检查这些文档与 root 完全相等，再检查剩余所有路径相等，PASS。没有为通过检查改动文档或业务。
- 首次部署探测尝试读取并不存在的 auth.go，且使用 ErrorAction Stop 查询已空的 17321 监听而退出；后续改为有界、空集有效的 ownedexe/PID/端口联合复核，确认旧 PID 已退出。首次任务摘要对登录任务 null NextRunTime 调用方法而退出；保存 nullable 值后通过。以上均为本轨复核脚本诊断，不改写先前 owner 的首 RED。
- 整树 `pnpm check`、`pnpm build`、完整默认 E2E 和远端 CI 由 root 在已发布 SOURCE 后续核实，本轨未重复，也未从 frontend 118 PASS/真实十值 PASS 推定全套通过。
- 实际管理员 Procmon 正式 DB/WAL/SHM 零写监控 **NOT_RUN**；真实连续两个日期 00:30 自动成功 **NOT_RUN**。任务 Ready、历史 LastTaskResult=0、只读 API 与静态部署不解除这两个 pending；执行步骤沿 backend 报告。
- 当前快照没有上海 10-14 今日事件，nullable/空集保持真实语义；本轨不刷新快照、不新增数据或重算动力学。
- 本次只交付技术集成与静态部署；main 合并、validation 解冻与正式科学/人工验收未执行，V4 正式源及日程零改动。
