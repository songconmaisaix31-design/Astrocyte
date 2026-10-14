# Works Relay v1.2 一页执行计划

## 当前 L1 继续开发（2026-10-14）

执行标准为当前根 [AGENTS.md](C:/Users/DW/orca/Astrocyte/AGENTS.md)。业务目标仍为 `workbench/v0.2-access`；本轮实施和验收见 [L1 计划](C:/Users/DW/orca/workspaces/Astrocyte/cognitive-stack-planning-20261014/tasks/L1-session-capture-plan.md)。主控 `term_4a28ac76-bd04-4a8b-acb2-b3114ca74ee4` 持有原 Run `run_d577f0d3fee1`，负责计划、决策、验收和最终 Git。

原 R 工程 SOURCE `62c7d99f1d6b180ef19f90a2112e801702d9fffb` 已交付，R 在原工作树继续 Task `task_a51fab762c01` / Dispatch `ctx_1415e8205d75`，负责实际原生控制器、来源覆盖与生命周期。H 在 `works-relay-http` / `songconmaisaix31-design/works-relay-http` 继续 Task `task_cfb7c9bde07d` / Dispatch `ctx_976e732672da`；F 在 `works-relay-frontend` / `songconmaisaix31-design/works-relay-frontend` 继续 Task `task_3c33a99685aa` / Dispatch `ctx_1db29418e335`。各轨保持既有独占文件范围；必要 L1 解析依赖的 `go.mod/go.sum` 由 H 独占修改。

原生归档保留策略的覆盖缺口需要现有控制器提供事件观察端口。R 本轮必要修改可包括 `internal/adapters/agents/native.go/native_process.go` 的最少可选端口与协议处理，保持默认行为；只进入 Workbench，validation / V4 对应文件不变。R 先发布具体接口、来源依据与验收方式 Handoff，再交 H 消费精确 SOURCE。

H 完成实际原生 Codex 控制器、会话来源、独立数据库与 HTTP 验证，错误立即返回，成熟格式使用库解析。F 连接 H 实际服务，在现有两个 viewport 做 DOM 验证，新增会话测试不使用 mock 或 HTTP 回放。注册数据新增、人工派发与三分钟 Today 可见时间须来自同一实际演示。SOURCE / REPORT 分开提交并 push；原唯一 I 在两轨完成后普通合并精确交付，只处理必要接口衔接。

本轮只使用自己持有的公开项目，原生文件和命令工具保持关闭。实际用户作品、工具执行、原件留存及产品活性窗口的未决项保持未决定。V4 正式材料、协议、引擎、mapper、认知 schema / migration、权重、快照流程及 17321 部署保持既定边界。中间文件放当前目录已忽略的专用目录；最终报告描述正确实现、实际验证结果和真实限制。

## 此前阶段记录

2026-10-14 L1续作写域：仍为原R/H/F及唯一同I，不重复采集owner。H当前task_f215e7452d70额外拥有既有cmd/server/relay_assembly.go、relay_observer.go、relay_config.go，新增cmd/server/relay_session*.go、httpapi/relay_session*.go、workbench/app/relay_session*.go及sqlite/access_session*.go（避开R的relay前缀），原server.go/OpenAPI/relay-session例子和自身报告。装配沿owned lifecycle→SessionCapture→独立既有EventService/EventRepo→六字段只读HTTP，GET不采集；旧SnapshotReader、引擎/mapper/schema/00:30不变。允许原应用状态目录单个原子替换游标JSON，仅位置/绑定UUID/文件身份/派生覆盖元数据；无新表/迁移/raw副本，不借认知状态存游标。事件库仍是事件权威，HTTP ID须可回查已提交receipt，静默无变化不保存游标。

原R task_3303613a7ff7继续owned只读引用与实际原生metadata provider，不读私有Session、不启动第二app-server。原F新Task task_33717da2d50e / Dispatch ctx_673d0bdbc496于同terminal/worktree/branch observed，消费H525d5df只读契约、既有generator并实现Today独立派生面板，原快照数值与来源不变。父CognitionPage的快照only footer留同I最终一行标签胶水；不扩大F写域。早期R7e28222/H3e22986/F38035bc只是可评审SOURCE，非最终集成或旧5dd CI覆盖。真实CLI工具/源作品/原件留存等未答问题仍未答；公开fixture及独立工程不等待这些答案。

事实源：validation 分支 `a429a4d` 的 [v1.2 定稿](C:/Users/DW/orca/Astrocyte/tasks/Works-Relay-v1.2.md)，取代所有旧版；旧评审421a076保留。目标分支 `workbench/v0.2-access`，从 `workbench/v0.1-workbench` 的 `9d78835bc2b00d3ed0fec493e3fa040704754abe` 建立；与用户验收基线 `d46f9f4cf2fd6e13c8362234746105330356c160` 仅 HANDOFF/STATUS/集成报告三项Markdown不同，业务全等。

主控：计划/状态/决策、独立验收、最终Git，不写业务代码。三条长期轨各一个Codex、一个Orca worktree/branch，读取全仓但只能写自己的路径；跨轨Handoff，原owner测试/文档/返修。使用现有客户端配置，记录实际模型，不指定未核实provider。最后一个集成owner普通no-ff合精确SOURCE/REPORT，仅少量装配胶水，领域问题退原轨。

| 轨道 | 独占 write_paths | 交付 / 依赖 |
|---|---|---|
| R-runtime | 新 `internal/workspace/domain/relay*.go`、`internal/workspace/app/relay*.go`、`internal/adapters/sqlite/relay*.go`、`internal/adapters/agents/relay*.go`、新 `internal/adapters/relayworktree/**`、`docs/access/runtime.md` | 持久单store/注册数据/三值/状态迁移及动作receipt；中性工作区、structured/terminal复用、互斥与权限端口；早期发布精确Go接口Handoff给H，不写共享contracts/入口 |
| H-http | `contracts/openapi.yaml`、新 `contracts/examples/relay*.json`、新 `internal/adapters/httpapi/relay*.go`、`internal/adapters/httpapi/server.go`、`cmd/server/main.go`、新 `cmd/server/relay*.go`、新 `internal/adapters/events/relay*.go`、`internal/workbench/app/contracts.go`、新 `internal/workbench/app/relay*.go`、现有 `internal/adapters/httpapi/workbench*.go`、现有 `internal/adapters/sqlite/workbench*.go`、`docs/access/http.md` | 同server/session/origin/ErrorV1 HTTP与组装，事件只走既有EventService端口，不改映射/引擎；早期发布OpenAPI精确SHA给F；Today实时接缝等待用户裁决，snapshot Reader仍只读 |
| F-ui | `web/src/App.tsx`、`web/src/components/Nav.tsx`、`web/src/api/schema.d.ts`（现有生成器）、新 `web/src/api/relay*.ts`、现有 `web/src/pages/workspace/Works*`、`web/src/pages/workspace/worksAlbum*`、新 `web/src/pages/workspace/relay*`、`web/src/pages/cognition/TodayView.tsx`、新 `web/src/pages/cognition/relay*`、新 `web/src/styles/relay*`、新 `web/e2e/relay*.spec.ts`、新 `web/src/**/relay*.test.ts(x)`、`docs/access/frontend.md` | 真实扫描总览、权限确认/手册/接力、三值任务卡/derived事件流/发送停止、E2E和三分钟脚本；消费H精确契约；用户未给图/活性窗未定前不冻结相应体验选择 |

上述 glob 限于明确新前缀和既有指定文件；三轨不可共写。当前Root仅本计划/HANDOFF/STATUS及 `tasks/Works-Relay-decisions.md`。root与业务worker不得改validation工作区任何文件；原有未提交V4恢复任务与Python缓存保留。

2026-10-14独占补充：F获准修改 `web/e2e/works-live.spec.ts` 的初始未知统计预期，从四个未知改为任务明确要求的五个未知，并核对新增平台指标标签；保留其余原断言、超时、skip与执行路径。此为新五指标行为的精确回归，不降低验收；禁止修改其他旧spec。

2026-10-14组装补充：H独占 `cmd/server/workbench_config.go`，仅抽出既有formal/snapshot/hardlink前置检查为显式目录helper，防relay模式被v0.1默认目录覆盖；旧Workbench默认行为与全部防护保持。新增回归写H已拥有的relay_config_test.go，不修改原workbench测试、快照脚本、Reader或00:30任务。

串行门：R早期Go接口→H组装；H早期OpenAPI→F生成schema；三轨完成→唯一集成→root完整check/build/E2E、真实CLI换手和terminal-only、三值、权限/互斥/持久迁移及receipt→常驻同URL部署。依赖前独立工作立即开工，不为等接口停整轨。

禁止：现有引擎/events mapper/Sleep/认知schema/migrations/权重/正式库与V4日程、`scripts/workbench/**`快照管道、现有依赖锁/Go模块、旧E2E断言/超时/重试配置；不自动派发/定时/链式派发；不新造调度器/Attempt/Manifest/Hash/证明系统。不硬编平台，不以手册统计假称真实接力。

运行时只写独立接入层本机状态，沿已有应用状态与权限服务；不对任何认知源/副本迁移或写入。三值只用事件/退出证据，静默非退出；未知启动不重派，降级前须确认无第二写者。所有raw/数据库/配置/浏览器及Procmon原件仅 `LOCALAPPDATA/astrocyte-workbench/v0.2-access`，不stdout/Git；报告只有derived与汇总，保留首RED/NOT_RUN/delivery_unknown。开发活动不从正式采集过滤。

用户待答接缝：Today实时本地事件追加或快照及时发布；当前未提交/未跟踪和非Git/单文件作品的中性工作区范围；参考图缺失时沿用既有视觉与可配置60秒窗口。独立模块可推进，相关语义未答不自行冻结。v1.0和实施窗口已定，不能重问。

权限待答补充：既有A/B/C是材料读取范围，不能代替CLI工具执行权限。当前原生驱动沿用禁止工具行为；用户尚未确认中性工作区read/write/execute执行范围，真实自由CLI候选未激活。无OS沙箱证据不宣称机器隔离。

工程候选：R SOURCE b19d836 / REPORT 4848a98；F SOURCE b48a6e5 / REPORT fdcf21b；H SOURCE c7c3255 / REPORT 6e211dd。三轨结算仅独立工程阶段，按用户固定owner要求保留返修。最后唯一集成Worker合精确提交、完整check/build/E2E，再由Root验核；真实CLI/Today/Procmon/自动两日未完成，不降低验收。

唯一集成I从root `8d9de2733713675c975b2920e71e21f07da7dcdc`创建works-relay-integration；Task task_db3080f0a933 / Dispatch ctx_ace29bd7baa1 / terminal term_55146ecc-fe9f-40b4-99c0-31d12b8eb5aa，Codex现有配置，turnStart observed。只写docs/access/integration.md及必要少量main/relay_assembly/httpapi server/App/Nav/现有生成schema胶水，领域返修退原owner。普通no-ff三REPORT无冲突，早期集成SOURCE `84bb3f95fc3bbd2d373a561336578ead706bb4b1`，仍待完整检查。临时同URL真实agents.json追加重启验列表，不dispatch模型；不替换17321，不启用未答门禁。

CI资源例外（精确首失败后）：I额外独占 `.github/workflows/check.yml`，只把Windows完整Playwright执行设为workers=1，沿已有4727f7a的两OS条件步骤；Linux原命令不变。首集成CI38463553347 Windows旧s1.spec.ts:178 page.reload报ERR_NO_BUFFER_SPACE，267 PASS/1 FAIL/38 SKIP；本机1worker完整268/38、Linux2workers完整成功，不是领域断言问题。原run保留，不改任何spec、viewport、fixture、timeout、retry、skip或业务；新SOURCE精确CI验证资源修复。

排期：D1注册/活性/中性工作区/总览，D2接力/实时，D3 E2E/实测/演示/部署。Procmon与00:30真实两日保持各自真实时间验收，v0.1未收尾不阻塞本任务，未发生不预填PASS。

工程验核更新：Root已普通合业务并push，4252db7完整check PASS、ff7c551 build PASS，I84bb完整默认E2E268 PASS/38 SKIP。b25双平台CI38466602701成功；后续bdf CI38468510081 Linux仅EOF测试末尾证据目录缺失RED，H同ownerTask task_c0577095b4cb / Dispatch ctx_a01ed6996519返修。允许路径仅cmd/server/relay_eof_http_test.go与自身报告，保持短Windows cwd、全部业务与断言/timeout/skip不变。最终CI及I报告待完成；原首失败均保留，用户门禁和正式验收不降低。

后次返修已结算：H SOURCE67c7d2a / REPORT9238e10只补显式证据目录，Windows深与空LOCALAPPDATA及23项cmd/server/vet通过；原owner retained。I普通合最终SOURCE5dd4b729并push，完整check113.70s PASS，CI38469944551待结果；Root449bbce独立两个定向均PASS，已合SOURCE祖先及H报告。新L1–L4事实源协作登记由独立文档owner完成，完整Session采集仍原H一轨，不增owner；R运行时与F Today写域保持，具体v1.3写域与零raw派生投影须登记。

v1.3 L1工程续作已开工：事实源精确DOC `913db499233720c369f7f5bf24965d9bc6a8ba54` 的tasks/L1-session-capture-plan.md（初始登记54f28ab，仅三新Markdown），原H同terminal/worktree/branch新Task `task_f215e7452d70` / Dispatch `ctx_2d4b642fbac5` observed。本阶段独占新internal/adapters/events/relay_session*.go及测试、自身docs/access/http.md；完整行/尾行/增量重启/owned身份/覆盖缺口/稳定origin/零raw派生与EventService受控工程立即推进。R只读Service端口先Handoff再同R续作；既有assembly/HTTP/Today暂不修改，R/F写域不转交。新采集源码与I当前5dd v1.2候选分开，I最终CI不被打断；后续仍唯一同I集成。精确读取DOC，不带入其validation父业务历史。未答项只阻塞依赖它的真实launch、原件复制及Today并置语义，不阻塞独立解析开发。

H早期接口Handoff已转原R：新Task `task_3303613a7ff7` / Dispatch `ctx_98f86f7facfd` 在原terminal/process/worktree/branch observed。仅原R relay contracts/service/tests及自身报告最小只读绑定，核grant/身份与closed尾部/重启gap；私有来源不进HTTP、不扫描home，不造proof/schema。首次start在输入前agent_unconfigured拒绝，request absent及Task列表无新R任务核对后switch原pane刷新识别才接受，原process incarnation保持，未重派未知操作。R/H跨轨只Handoff，不互写；旧I仍仅5dd候选。
