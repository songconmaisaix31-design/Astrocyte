# Works Relay v1.2 一页执行计划

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

排期：D1注册/活性/中性工作区/总览，D2接力/实时，D3 E2E/实测/演示/部署。Procmon与00:30真实两日保持各自真实时间验收，v0.1未收尾不阻塞本任务，未发生不预填PASS。
