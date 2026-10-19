# 交付候选、验证与运行切换计划

## 范围与准确来源

本次依据主控 `tasks/delivery-candidate-20261020.md`、当前根目录 AGENTS、CODING、RESOURCES、QUESTIONS，以及原 frontend-observatory / dynamic-dag 任务执行。目标为同一准确源码的产品、构建、验证和实际服务对应；独立 Reviewer 决定本次接受范围，主控推进原集成分支并负责原服务切换。

| 项目 | 当前事实 |
| --- | --- |
| Worker / Run | `task_6f09f37378a8` / `ctx_1e0467a3f24a` / `run_1187287582ea` |
| 工作树 / 分支 | `C:/Users/DW/orca/workspaces/Astrocyte/frontend-observatory-20261018` / `songconmaisaix31-design/frontend-observatory-20261018` |
| 接管 HEAD | `5725c1840a760105518d273eaaa28d7367f369ed`，接管时 tracked 干净 |
| 最新业务检查源码 | `22cde9efc540553c433463af5b9e955790d57a99`；至接管 HEAD 只改 README 与产品报告 |
| 集成目标 | `songconmaisaix31-design/works-relay-integration@f17a491cdd6392829094c95a3e5efb6011362bc0` |
| 原 Windows 包 | Source15 `15d74da3edcee26550a71120370fda9c86230e04`，不能代表当前候选 |
| 原运行程序 / Web | `c5b3b5715f2e93b82365720914be24c73642bd1e` / `172bb445597193d55f1563ee5d21ac2ec881798d` |
| 实际客户端 / 模型 | Orca Codex；本次 `orca terminal show` 的终端信息为 `GPT-6.1-Sol xhigh fast`，未覆盖模型、未派生子 Agent |

阶段 A 只读取现有文件、Git 和已授权元数据并提交本文；没有运行编译、测试、安装、服务或浏览器。阶段 B 须同时收到计划接受和重操作窗口授权。后续业务缺陷先交具体文件与补充测试计划，经审核后由本 Worker 在原模块修复。本文、`.verification/**`、`dist/**`、`web/dist/**` 为当前写入范围；主控管理文档和独立审查报告由各自负责人维护。

## 默认检查入口与已核原件

默认 workflow 为 `.github/workflows/check.yml`，矩阵 `ubuntu-24.04` / `windows-2025`。顺序为 Go 模块下载、冻结 pnpm 安装、`pnpm check`、Linux `go test -mod=readonly -race ./...`、`pnpm build`、Chromium 安装与默认 Playwright。Windows 使用 `--workers=1`，Linux 使用 `pnpm test:e2e`。

`pnpm check` 调用 `scripts/check.mjs`：先通过 `createCheckOutputDirectory` 分配实际输出根，再执行 gofmt、全仓 vet、全仓 Go 测试、模块验证、依赖与架构检查、stop-process 测试、OpenAPI lint / 示例 / 生成一致性、S1 acceptance、tsc、eslint、Vitest 和 Git 空白检查。该顺序和全部原断言保持。

历史准确源码 `982333b871520e606c74d553e448c5049f81abde` 的 push run `39155403418` 与 PR run `39155406045` 被主控取消；Ubuntu job `117757474469` / `117757483047` 的 `pnpm check` 在取消前已 FAIL，全仓 Go 测试实际执行。Windows 及后续未完成阶段不能登记通过。原始失败日志在独立旧审查树：

- `C:/Users/DW/orca/workspaces/Astrocyte/dynamic-dag-review-20261019/.verification/dynamic-dag/ci-982-push-job117757474469-failed.log`
- 同目录 `ci-982-pr-job117757483047-failed.log`。

静态核对所得分类如下；分类只说明可定位的原失败，不将所有失败归因于环境或新回归。

| 原件事实 | 当前解释与处理 |
| --- | --- |
| OpenViking HTTP、ProgressPreflight 及其他实际检查缺 `ASTROCYTE_TEST_OUTPUT_ROOT` | Source22 已修默认 check 进程输出前置；已实际验证的函数范围见下一表，其他函数不推定通过 |
| TasteProfile 迁移数量预期 26、实际 27；PR job 的 discovery 在 detached / dirty 条件失败 | 第 01 轮已有准确来源的定向修复与独立接受；复用其原条件结果，不扩展为默认 CI PASS |
| Relay HTTP reconcile 请求为403，原用例期望200 | 原批准缺少该动作；没有降低权限或改成200，交原模块处理 |
| `relayControlled`、`localNative` 等旧替身及原生用例 | 当前任务禁止执行；不运行含这些用例的整套默认测试来制造新结果 |
| 公开论文 / 视频 / Q 快照 / SQL 期望 / 原生回执或 OpenCode CLI 缺失 | 真实来源或能力前置缺失，逐项保留；输出根修复不能补齐这些来源 |

Source22 原件根为本树 `.verification/restart-product-20261019/`。已读取 result.json 与实际 check.log：

| 准确来源 / 原件 | 实际结果与接受范围 |
| --- | --- |
| before `2a05bd0d78cbeb5e69ba09728923ea2f2d1d1b65`；`ci-before-01/before.stdout.log` | 原 OpenViking 函数 EXIT1，创建实际 HTTP / SQLite 前因缺输出目录失败 |
| Source22；`ci-openviking-01/{result.json,check.log,resources.json}` | 原 OpenViking 函数 EXIT0，1函数，0 Skip；无参数分配副本默认 `.verification/check-DOIgoT` |
| Source22；`ci-progress-01/{result.json,check.log,resources.json}` | 原 ProgressPreflight 两函数 EXIT0，0 Skip；显式父根 `evidence/check-04Syyp`，真实配置拒绝、unknown保留、同库重启拒绝重发及 SQLite READONLY / HTTP500 |
| Source22；`ci-invalid-roots-01/{result.json,check.log,resources.json}` | 相对路径及实际普通文件立即拒绝，Go 启动0 |

`git diff 22cde9ef..5725c184` 已核只有文档。上述输出前置结果在准确未变条件内复用，不再重复 before / after 克隆及相同拒绝。最新候选的必要真实功能检查和新包验证仍须执行。

本次最新候选完整默认 CI 当前为 **NOT_RUN**：默认集合包含当前禁止的替身和未满足门禁的原生路径。历史准确982的 `pnpm check` 为 **FAIL**，两场 run 最终为 **cancelled**。不会改 workflow、添加测试 skip、降低断言、伪造来源或把选集执行称为默认通过。发布门禁保持。主控已明确允许普通 push 及原 PR #2 正文更新；纯文档后继 Commit 使用已有 `[skip ci]` 约定，不 amend 既有 Commit，不触发远程禁止集合。

## 必要检查与需求对应

全部命令针对阶段 B 确定的准确候选 SHA，记录工作目录、命令、源码、真实退出码、stdout / stderr、资源采样和原件路径。每项失败立即停止该阶段并报告；其他独立检查仅在明确窗口内继续。没有源码变化的已接受性能比较沿原来源复用，不重复进行前后优化测量。

| 层次与需求依据 | 计划命令 / 操作 | 必须核对的结果 |
| --- | --- | --- |
| 单元：动态任务完整性、SPEC §10.1 / §10.4 | `go test -mod=readonly ./internal/swarm/domain ./internal/swarm/app ./internal/adapters/sqlite ./internal/adapters/httpapi -run '^TestDynamicDAG' -count=1 -timeout=2m -v` | 当前9函数：空图、分支汇合、环、重复节点 / 边、自依赖、缺端点、跨 Mission；未知用量及已知超限分开；停止错误类型、历史确认编码及真实 Registry 缺失原错误保持 |
| 集成：同命令中的实际 SQLite / HTTP，原动态 DAG 验收 | 同上使用原真实生产装配、公开 Git 与实际 human 会话 | 人工0→2节点 / 1依赖、每个 event_id 回查 foundation_outbox；缺依赖拒绝且不部分写入；read_swarm、跨项目及当前撤权；无 Worker 的停止事件只产生一次，批准过期仍可按原规则读取历史 |
| 集成：可选 OpenViking、SPEC §17.1 | `go test -mod=readonly ./internal/attention/domain ./internal/adapters/openviking ./internal/attention/app ./internal/adapters/sqlite ./internal/adapters/httpapi ./cmd/server -run '^TestOpenViking' -count=1 -timeout=2m -v` | URI与绑定 / 配置参数，实际未配置200、远端动作503、真实连接拒绝、空集合、Origin / CSRF、缺失 / null level 拒绝；零远端提交。无匹配函数的包只计编译 |
| 单元：界面显示与错误状态，SPEC §16.4 | `pnpm --dir web exec vitest run --maxWorkers=1 src/components/displayContract.test.tsx src/utils/__tests__/format.test.ts src/utils/__tests__/commandError.test.ts` | 现有参数与显示规则，不计 API / Native 正向 |
| 生产端到端：原前端和 DAG 验收，SPEC §16.5 / §17.3 | `pnpm --dir web exec playwright test --config dynamic-dag.config.ts --workers=1`；随后 `pnpm --dir web exec playwright test --config observatory.config.ts --workers=1 e2e/observatory.check.ts` | 真实 HTTP / SQLite，1920×1080及1280×720，人工增长和逐事件SQL、方向 / 缩放 / 键盘、长标题、断线stale / 同库重启恢复、撤权清图和离页停止；真实移走自身dist分块、原context普通reload恢复；其他生产页面、批准 / 作品入口、搜索与26品牌HTTP / decode。截图 / 视频 / trace关闭 |
| 静态和契约：现有公开接口与依赖边界 | `pnpm check:contracts`、`node scripts/check-architecture.mjs`、`node scripts/check-dependencies.mjs`、`pnpm --dir web typecheck`、`pnpm --dir web lint`、`git diff --check` | 保留真实错误与既有警告；不生成新模块或新接口，不为本任务扩充架构规则 |
| 编译检查：当前整体装配 | `go vet -mod=readonly ./...`；下节同来源生产构建 | 当前源码诊断和构建成功；不把编译计为真实原生可用 |

Playwright helper 会编译并启动实际 Go 服务；其编译、服务及浏览器均占用唯一窗口，每次串行，绝不与64505共存运行。若旧入口内额度小于当前上限，保存实际额度，不为提高额度重跑。`observatory-focus.check.ts` 的既有历史 Q 检查依赖原只读来源；没有变化的 Reader、历史与性能结果继续沿原准确来源和限定条件复用，不能当作今天在线活动。切换时核原Q指针及实际接口历史，不重复采集、Sleep或完整认知验包。

不运行模型、Claim / Send、带 Worker 的停止退出 / 用量 / 释放、未知副作用重发、实际采用或自由工具验收。原生工具关闭、S3两项待答、OpenViking无实际配置、经验资料准入与采用的条件保持；这些范围独立登记未验。

## 同源码构建与 Windows 交付

主控已确认本次必要交付为普通 Go / Web 构建及同库服务。阶段 B 在计划接受后冻结准确源码 S，tracked 干净后构建；业务返修导致 S 变化时，更新来源和受影响检查，不能复用旧程序。后继报告 Commit 可以只追加本文，必须以 `git diff S..<报告HEAD>` 明确业务与构建输入未变，直接指向 S 和实际产物；不为报告重复构建。

1. 从本树准确 S 执行现有 `pnpm build`，由 `scripts/build.mjs` 串行编译 `dist/astrocyte.exe`、`astrocyte-mcp.exe`、`astrocyte-sleep.exe`，再以 tsc / Vite 构建完整 `web/dist`。复用已有真实依赖缓存；缺少依赖时立即报告，安装另经窗口安排。旧 Source15 程序和 Web 不作为当前构建产物。
2. `go version -m` 必须记录实际 `vcs.revision=S`、`vcs.modified=false` 与 Go 版本；记录构建前后Git状态。构建日志、资源原件和工具版本保存在 `.verification/delivery-candidate-20261020/`。既有普通构建具有 Git 版本来源，不使用 archive 代替本次构建。
3. 创建本树全新 `dist/delivery-candidate-20261020/` Windows 交付目录，完整复制本次三程序、Web、现有启动脚本和使用说明。保存实际目录及全部复制来源，保留原始 `dist` 与 `web/dist`；启动和专项核查使用同一批产物。原已有固定模型 / SDK / 媒体依赖继续按原可信配置使用，不复制完整模型runtime或打包用户SQLite / 配置。该目录是普通产品候选，完整便携模型包尚未构建或验收。
4. 新交付目录实际端到端通过后，独立 Reviewer 核准确 S、对应产物和本报告，决定接受范围；主控普通推进原集成分支。原PR #2正文通过 `.verification/delivery-candidate-20261020/pr-body.md` 和 `gh pr edit --body-file` 更新，列明实际命令、范围和限制；不自动合并、创建 Release / tag。

Source15 完整 Windows ZIP、130项解压运行和独立验包保持原准确来源的历史结果。本次没有新完整便携包；Build-Windows / Verify-Windows / 模型复制 / 全量认知验包均不作为本次必要前置，发布状态保持。新增本树验证入口只编排实际工具、参数与日志，不制造来源、响应、CLI或原生结果。若普通构建暴露超出已授权范围的缺陷，先交具体文件与补充测试计划。

## 原真实库与服务切换 / 恢复

只读 `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/running.json` 已确认原配置指向：

- URL `http://127.0.0.1:64505/`，记录 PID45676，创建UTC `2026-10-19T12:09:30.7340265Z`。
- 程序 `versions/c5b3b571/astrocyte.exe`，Web `versions/172bb445/web`。
- 原数据库 `local/astrocyte-workbench/v0.2-access/app-state/state.sqlite`，完整绝对路径位于上述共享运行根。
- Q 为 `release-cognition-20261014/.verification/cross-day-20261015-0228/local/astrocyte-workbench/v0.1`，generation=`snapshot-1711995690`。
- 媒体配置为原 `standards-backend/.verification/video-trusted-config-first/environment.json`；仅读取既有白名单配置，不输出凭据。Laya / Jev未配置，原生输入0的历史事实不替代当前核查。

原 `works-relay-http/dist/product-recovery-20261016/restore-38fb.ps1` 可参数化核准确程序、原数据和媒体 / Q；其全表计数完全相同断言无法直接覆盖本候选的027迁移。主控已要求本树提供可审查的显式阶段切换脚本，原脚本及门禁保持。脚本实现须在本计划接受后进行，Worker不执行外部部署。

| 阶段 / 执行归属 | 计划与失败边界 |
| --- | --- |
| 切换前 / 主控 | 重新核PID、创建时间、程序路径、唯一监听与原SQLite实际路径；任何身份不同立即停止，不结束未知进程。主控停止已确认旧服务，确认退出及64505无监听后才交出重操作窗口 |
| 保全 / 主控审批脚本 | Python标准 `sqlite3.Connection.backup` 从 `mode=ro` 原库创建一致备份，明确 `query_only`；保留原对象目录及配置 / running.json / Q指针。记录 integrity_check、foreign_key_check、sqlite_schema、准确 `_migrations.id` 和所有既有业务表行数；完整数据仅保存在私有忽略目录。备份和源库逐表内容通过SQLite双向EXCEPT比较，不能只有行数。禁止单独复制活跃WAL主文件冒充一致备份 |
| 迁移预演 / Worker获准副本 | 对主控提供的自有副本使用候选实际程序启动迁移；独立端口和隔离原生配置，禁止扫描用户会话。核027准确新增表 `attention_openviking_submissions` / 唯一索引与 `_migrations` 新记录；原迁移ID保持、原业务表schema及内容保持、完整性与外键检查通过。除准确预期新增外出现变化立即停止。原对象保持只读；不把副本预演计为原服务已切换 |
| 切换 / 主控 | 审核本树脚本与独立结果后，部署同一候选程序和Web到新版本目录；64505、原数据库、原公开项目 / import roots、Q、媒体配置保持明确参数。Go4 / -p=4 / 512MiB、Node1024MiB、BelowNormal；关闭CodeX沉淀、保留archive来源，Laya / Jev / OpenViking不凭空配置。新启动只一次，失败不自动重试或恢复 |
| 同库核查 / 主控 | 每次HTTP前后核本次持有进程、实际PID / 创建时间和唯一监听；health、首页及全部本地assets与本次构建内容对应，human session、只读核心列表 / DAG / 可选状态。SQLite原业务数据与备份逐表比较，027新表应为空，迁移记录预期增加；若已有待运行任务可能触发生命周期修改，先核条件并交主控，不忽略数据差异。核Q与媒体参数保持，不触发原生 / 模型 |
| 恢复 / 主控显式决定 | 新程序失败立即报告并保存原件；主控确认新进程身份和退出后保全失败数据库 / WAL，再从一致备份恢复原库及旧c5 / 172程序目录与配置。不能直接用旧程序打开已迁移库冒充恢复成功；不能覆盖候选运行后新产生的数据。备份后若有合法用户写入，先保全差异并交主控决定。旧入口可用于恢复后的准确旧来源，资源守卫及同库 / HTTP核查照常执行 |

## 资源、停止条件与交付记录

阶段 B 全局只有一个重操作窗口、一个服务、一个浏览器。每项重操作两次真实CPU / 可用内存采样，均CPU<85%、可用内存≥4GiB，单调间隔2–5秒并分别保存UTC；BelowNormal，Go `GOMAXPROCS=4` / `GOFLAGS=-p=4` / `GOMEMLIMIT=512MiB`，Node heap≤1024MiB。保留全部拒绝，不自动重试；不终止其他用户进程。只读 inspect 只有满足 RESOURCES 专用定义才使用对应额度。静态检查按该文件例外串行执行，不安装依赖。TEMP / TMP / GOTMPDIR和本次日志均在已忽略专用目录；禁止 `/tmp`。

当前停止点是等待独立计划审核及主控窗口。后续每个失败提交真实文件、错误与最小必要受影响验证范围给主控；业务选择未答时只继续无依赖工作。准确候选默认CI状态、必要检查、构建、独立结论、主控实际运行来源和数据保全全部有可核查记录后，主控决定本次收敛完成；完整CI / S3 / 原生 / 采用 / 发布的真实限制保留。

阶段 B 报告需追加：准确源码与报告Commit、默认CI各阶段PASS / FAIL / NOT_RUN及原因、完整命令和原件、资源样本、Go / Web / Windows目录、实际SQLite与备份路径、运行参数 / PID / 地址 / 程序路径、主控切换结果和独立Reviewer结论。下一项真实产品任务由主控依据本次证据与 QUESTIONS 明确可执行的六步 / 原生前置；未决业务条件由主控确认后再形成独立任务，本次不自行开始。
