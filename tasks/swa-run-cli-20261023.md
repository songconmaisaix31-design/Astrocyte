# SWA 一次提交与完整任务 CLI

## 范围、基线与来源

负责人 `run_cli`，共享工作树 `/workspace/swarm-anything`，分支 `integration/swa-task-cards`，起点 `11619d72f08c163f159a4f4befe270806b1b67ea`。本任务落实用户 2026-10-23 的第一项要求。实现前需主控冻结跨模块合同并由独立 Reviewer 接受本测试计划。

已完整读取 [OpenMausBot 固定版本 headless CLI 文档](https://github.com/milind-soni/OpenMausBot/blob/30b62aaf18adc22099c24a7181abf5e9ddead980/docs/headless-cli.md)。原文保存在 ignored `.verification/swa-run-cli-20261023/openmausbot-headless-cli.md`，仅作为设计参考。采用一次提交、重连、递归等待、严格失败、独立等待退出状态、稳定请求 ID 和完整导出；权限继续使用 SWA 当前项目 Grant，不能采用原文中回环服务自动成为 owner 的规则。

现有生产路径：`cmd/swa/main.go` → `POST /api/v1/local-projects/{id}/orca/commands` → `OrcaExecutionService.ExecuteOrca` → runtime `client.ts`/`native-methods.ts` → 原生 Orca Run/Task/Dispatch。已有规范 UUID、原请求查询、实际选择及身份、一次 stdout JSON 和独占目录 `task.result` 导出。原生 `task.create` 已通过稳定子请求 ID 创建 Run 与根 Task。

现有缺口：CLI 只有单次低层命令；controller helper 不是执行主 Agent；根任务仍需显式 `task.complete`；`task.result` 只接受根的直接子任务且完整库存上限 100。CLI 不能把 Worker 启动成功、任务静止或 HTTP 成功解释为任务交付。完整任务状态、递归关系和成果来源由本轮后端负责人统一提供。

## 已确认合同

```text
swa run --project P --workdir W --spec-file F --input-file I
        --agent A --model M [--name lead] --request-id UUID
        [--role ROLE --role-revision REV] [--conversation CONVERSATION]
        [--timeout 30m] [--on-input exit|wait]
        [--strict] [--export ABS] [--json]
swa wait --controller C --task ROOT --run R
         [--timeout 30m] [--on-input exit|wait]
         [--strict] [--export ABS] [--json]
swa export --controller C --task ROOT --run R --output-dir ABS [--json]
```

`run` 默认等待整个递归任务。每次输出保留原 Run、根 Task、controller 和原 request ID，重连不查找“最新任务”。等待期间 timeout 与 Ctrl-C 仅取消当前 CLI 的等待，原任务继续运行；不会隐式 stop/release。需要用户输入默认返回单独退出码；`--on-input wait` 继续等待同一任务。

HTTP 动作为 `run.submit`、`run.show`、`run.export`，沿当前 OrcaCommand/OrcaEnvelope；OrcaCommand 增加 `run_id`。后端共享快照位于 `result.native.run_view`，schema 为 `swa.run-view.v1`，包含 `controller/run_id/root_task_id/status/root/tasks/dispatches/questions/blockers/failures/deliverables/complete/observed_at`；status 为 `active/waiting_input/blocked/completed/failed/unknown`。GUI 与 CLI 使用同一快照，CLI 不通过任务数量计算完成。提交可先返回准确原身份，CLI 随即读取 Run。导出位于 `result.native={run_view,result}`，其中 result 是既有完整 task.result 原件。

问题回答增补使用 `question.answer`，公开命令固定 controller/run_id/root task_id/dispatch_id/question_id/request_id 和 input_file 或 text 二选一；CLI使用 `swa question answer` 及文件，GUI使用同 Run 表单。新增 `answered_questions` 由原问题和答案消息完整派生，pending questions 含义保留。原文最多32000 UTF-8字节，原 request 未确认时只能查询；准确 completed parent/child 均明确 rejected/effects_applied:false 后才允许用户显式修正并新建请求。人类按当前项目独立send授权，Agent仍绑定原Grant。原 reply/ask 与 db.answerQuestion 是唯一答复语义，不另发worker.send。

该增补统一由 `remaining_delivery_audit` 维护，经 Root 接受及独审计划 `dacf0594235054725dd6215b8ecd107d56bbe2b99bf854ea9a2f0b66a872e4b3` 后实施；业务/契约/GUI写域另含直接角色schema消费者和Vitest显式清单。当前产品没有合法生产 ask 入口，provider question journal也没有转换到原问题表；真实模型提问、原ask等待/返回和完整CLI/GUI效果链继续未验，属于具体后继功能及运行范围，不能以数据库局部检查代验。准确局部证据与已知失败修复写入 `.verification/cloud-setup/run-question-answer-implementation.md`。

`run.submit` 由后端复用当前 controller/task/worker 的实际生产调用和原生请求账本。一个外部 UUID 派生固定阶段请求 ID，重复或重连不随机重造提交身份；同 ID 改变内容须沿原幂等规则拒绝。任何阶段已送出但结果未知时停止写入并查询原请求；客户端不自动重发提交。读请求可在暂时断线后继续重连同 Run；明确权限拒绝立即结束。

主控确认本批 `run.show` 立即只读返回，不增加尚未实现的 `wait_ms`。CLI 在自身 timeout/context 内按有界间隔重读同一身份，不读取或确认 Agent 信箱，不建立后台轮询任务、调度器或另一套完成状态存储。

保留低层命令现有退出语义，完整任务命令使用：0 成功；2 参数/文件错误；3 权限/批准不足；4 能力不可用；5 资源拒绝；6 效果未知；7 任务失败、严格子失败或导出失败；8 等待超时；9 需要用户回答；130 当前等待取消。`--strict` 使当前 Run 完整递归任务集合中的任一子任务失败返回 7；实际创建 Task 的 `consult` 同样计入。单向 `notify` 没有 Task，不持有任务完成条件。

`--json` stdout 始终只有一个最终 JSON，stderr 输出进度 JSON 与可直接重连的原身份命令。进度不包含 token、API key 或原生秘密配置。JSON 的 `ok` 仍表示当前命令结果；任务业务状态单独来自后端快照，不能把提交成功当成交付成功。

`--export` 在任何提交之前检查目标为绝对且独占的新目录，现有目录拒绝；本轮不提供覆盖文件的 `--force`。`export` 复用后端完整原生结果读取及当前项目文件权限，保存任务记录、实际结果与可准入的产物。任意 Markdown 链接不能直接成为文件读取授权；拒绝越界、符号链接、任意目录递归和不完整成果。后端返回固定导出合同后，CLI 只负责安全写入。

## 写入归属与依赖

独占写入：`cmd/swa/main.go`、`cmd/swa/output.go`、新增 `cmd/swa/run_flow.go` 及对应 `cmd/swa/*test.go`；`runtime/swa-orca/USAGE.md` 的新增完整任务 CLI 章节；本任务文件。保留 serve 和原低层命令语义。没有授权修改原生 Orca、Go app/HTTP、共享 JSON/TS 契约、GUI、任务状态库。

`collaboration_model` 负责 Run 提交、递归快照、交接与成果的后端语义；Root 负责 Go/schema 公共合同与 GUI；独立 Reviewer 负责计划和源码/结果审查。准确DTO冻结前不写依赖该DTO的产品代码。

## 验证计划

1. 本地真实 CLI 参数与文件路径：run/wait/export help 不访问服务；未知参数、缺失 UUID、非法时限、非绝对/已存在导出目录、非法等待组合在传输前失败；已有 controller/task/worker/request 命令准入保持。
2. 正常真实服务链：同一安装与当前项目 Grant 下通过 run 提交、取得原生 Run/root/主 Dispatch；CLI 与 GUI 读取同一递归集合，等待至真实交付并导出完整结果。实际 N=4、N=6 与递归孙任务由总验收驱动覆盖，不以 N=3 或仅顺序完成代替。
3. 幂等与断线：同一稳定请求 ID 重复提交只产生一份原生工作；内容变化拒绝。客户端断开后用打印出的准确身份 wait，运行期间或已结束时均返回原任务；暂时服务断线只重读；未知送达后不产生第二次写请求。根据真实请求账本和 Dispatch 库核对数量。
4. 独立退出：实际需要用户回答、实际等待超时、实际失败和 Ctrl-C 分别产生 9、8、7、130，stdout 为一个完整 JSON，原任务继续存活；权限撤销/过期为 3 且停止查询。失败路径必须通过真实生产服务操作或已存在原始失败记录验证，不能伪造 HTTP、CLI、Agent 或数据库执行事件。
5. 递归严格语义：主任务成功且一个真实子任务失败时，普通结果保留失败事实，strict 返回 7，包含实际创建 Task 的 consult；孙任务未完成必须继续等待；notify 不持有完成。动态创建的孙任务与主任务收尾须通过后端一致快照/事务门判断，不能以暂时空集合判为完成。跨项目、跨 Run、不同 controller/root 或未验证递归集合失败关闭。
6. 导出：真实完整结果与后端原件匹配，结果文件只出现一次；缺失/增长中的原生结果、非法/已存在输出、实际越界/符号链接和写入失败均保留错误，不能返回截断的成功成果。输出错误时只写 stderr 和已知原身份，不尝试第二个 stdout JSON。
7. 局部检查：使用已配置 Go 工具链运行 `go test -mod=readonly ./cmd/swa` 适用测试及 `gofmt` 检查；真实安装检查归入统一验收，不自行启动完整 build 或 Windows CI。跨合同静态检查由主控安排一次整合运行。

测试前将 TMPDIR 指向 ignored 工作目录，避免使用 `/tmp`。保持真实最小复现、原日志和精确原因；任何失败在修复并局部验证后先交代结果，随后才可启动下一完整构建。没有真实模型或平台证据的场景继续写为未运行。

## 当前结果

已完成固定来源和现有生产路径调查。主控已确认 RunView、动作、Go RunID 字段、CLI 自身有界只读轮询及本模块写域。独立 Reviewer 已接受整体计划并要求 strict 覆盖所有实际递归子任务、动态孙任务通过后端一致快照判断；本文件已修正。进入 CLI 实现，真实验收尚未开始。

首轮局部命令 `go test -mod=readonly ./cmd/swa -run 'Test(RunCLI|RunExport|OrcaCLIRejects|OrcaCLIMissing|OrcaCLIHelp)' -count=1` 在共享 Go 接线仍处编辑中时执行，退出 1，未运行 CLI 用例。原日志 `.verification/swa-run-cli-20261023/local-cli-test-01.log` 的全部诊断是 `OrcaRunSubmission` 与 `submitOrcaRun` 尚未定义，对应主控正在原子补充的 `orca_run.go`。停止复跑，等待主控声明该接线就绪后再做同一局部验证；未启动完整构建。

主控补齐 Go 接线并完成其局部编译后，同命令第二轮实际退出 0，`cmd/swa` 用时 0.011s，原日志 `.verification/swa-run-cli-20261023/local-cli-test-02.log`。此轮覆盖新旧本地帮助、参数拒绝、缺少当前 Grant，以及真实文件系统的新/已有/相对/缺父目录和符号链接导出前置。`git diff --check` 通过。没有用假 HTTP、原生协议或 Agent 事件测试成功路径；同安装真实 Run、超时/问题/严格失败、完整导出与 N>3 仍待整合验收。

当前 CLI 初稿包含单次 `run.submit`、准确身份的只读 `run.show` 等待、终态后的 `run.export`，stdout 沿原类只发布一次最终封套，stderr 发布进度及重连命令。导出写入独占 `run.json`、`result.json`、真实 Task 结果文本和对应 Task ID 索引；不读取结果中的任意文件链接。独立源码审查进行中。

## GUI 同一 Run 接线

主控后续授权本负责人追加 `web/src/pages/swarm/SwaRunWorkspace.tsx`、`swaRunView.ts`、`SwaWorkspace.tsx` 的必要接线及 `swaObservation.ts` 消费者扩展、共享样式中的复选框尺寸、`web/e2e/swa-workspace.spec.ts` 对应真实 DOM 检查。一次提交表单复用当前 `swaApi`、项目能力准入和稳定请求保存；未知时仅查询同一原请求及其已记录阶段，不自动重发。准确 controller/root/Run 重连通过 `useReadApi` 读取同一 `run_view`，展示完整任务集合、阻塞/问题、严格失败事实与成果入口；原 Worker 操作仍连接同一个根任务。长期角色面板由 capabilities 负责人独占实现，本模块负责接入回调与准确角色版本、讨论会话 ID 透传。

GUI 验证包含 TypeScript 与 ESLint、原观察/准入规则测试，以及真实服务上的无截图 DOM 操作：缺权限/未知请求阻止重复提交、项目切换和过期快照禁用操作、实际 Run 与 CLI 相同、重连只读、成果入口只接受后端完整结果。浏览器不采用 mock 路由、伪造 Agent 或截图。后端尚未启用时只验证实际不可用/拒绝状态，不能把它记作 Run 成功验收。

既有真实服务不可用 DOM 用例增加完整任务提交与重连的禁用断言。允许 observe 授权后的 role.list/conversation.list 真实只读查询，保留准确项目路径和无写动作断言，原项目控制许可往返、撤销及版本冲突用例继续运行。

局部 Web 类型检查第二轮实际退出 2：仅 `SwaRoles.tsx` 的四处 `role.skills/role.knowledge` 可为 null，却直接调用 join。最小复现 `pnpm --dir web exec tsc -b --pretty false`，完整日志 `.verification/swa-run-cli-20261023/web-typecheck-02.log`。capabilities 负责人依共享合同把明确 null 归一为空引用数组，修复四处消费者后同命令第三轮退出 0（`web-typecheck-03.log`）；未启动完整候选构建。

最终 CLI 局部第四轮 `go test -mod=readonly ./cmd/swa -run 'Test(RunCLI|RunExport|OrcaCLIRejects|OrcaCLIMissing|OrcaCLIHelp)' -count=1` 退出 0（`local-cli-test-04.log`）。指定 GUI 源及 DOM spec 的 ESLint 第二轮退出 0（`web-lint-02.log`）。`pnpm --dir web exec vitest run src/pages/swarm/swaObservation.test.ts` 实际 9 项全部通过（`web-observation-test-01.log`）。这些结果只覆盖局部类型、静态规则、参数、文件边界及既有观察纯规则；不代替同一实际安装 Run 的模型成功证据。

DOM 第一轮退出 1，业务用例尚未执行：Chromium 报 `Socket path too long`，准确失败路径为本次长 TMPDIR 中的 `org.chromium.Chromium.*/SingletonSocket`（`web-dom-01.log`）。改用仓库 ignored `.verification/run-tmp` 后，单独实际 Chromium launch/close 检查退出 0（`browser-start-local-01.log`），随后同一 DOM 命令第二轮实际退出 0，2 项通过，用时 14.0s（`web-dom-02.log`）。命令为 `ASTROCYTE_TEST_BROWSER_EXECUTABLE=/usr/bin/chromium pnpm --dir web exec playwright test --config e2e/swa-workspace.config.ts --project chromium-1280 --reporter list`，TMPDIR 与输出根均显式使用本仓 ignored 目录；实际 Go HTTP、SQLite、Vite、浏览器产生的原日志保存在 `dom-r2/`。没有使用截图、mock 路由或虚假原生数据。

同一 DOM 检查确认完整任务提交/重连在运行时不可用时禁用，角色/讨论读取失败时新建入口禁用且编辑/归档/追加入口不存在；撤销 observe 后这些入口全部消失，发出的命令仅 status/role.list/conversation.list 且项目路径准确。原权限保存、撤销与陈旧版本冲突同样通过。该环境未启用 native，无法触发真实 Run 或未知写入，相关成功/未知场景仍待最终集成。

等待期间完成后的 export 读取若恰逢 deadline，也保持 8 超时与 130 用户取消的区分。此修正后的 CLI 局部第五轮实际通过（`local-cli-test-05.log`，0.010s）；`run_flow.go` SHA256 为 `7fe0928d7f456f08c581827a79afef7b7d97410eada5417b9527ab9deecae0bd`。最终指定 8 个 GUI/DOM 文件 ESLint 第三轮退出 0（`web-lint-03.log`），包含 capabilities 负责人最终角色三个 TypeScript 文件。

最终 Web `pnpm --dir web exec tsc -b --pretty false` 第四轮退出 0（exec session 89841 完成记录，`web-typecheck-04.log`）；指定 CLI/GUI/DOM/USAGE/任务文件 `git diff --check` 退出 0。未创建 Commit 或发起完整候选构建，准确分支仍为 `integration/swa-task-cards`；同一安装完整 Run/N4/N6/Windows 与其失败分支由总验收继续。

后继 Root 冻结 `run_view.acceptance: unreviewed` 与 `handoffs` 实际原生消息数组。GUI 区分执行完成与成果未审，展示真实 from/to/text/read 和原 record；观察不确认消息已读，恢复排队不表示父已处理。CLI 进度保留 acceptance 与交接数量，要求同版本完整集合，成功退出不宣称人工已经验收。新增字段消费后的 run_flow SHA256 为 `62631bccf05d93d57ca6a16aaf35c3372be8c84750fff814a646262588b2645d`，替代此前摘要；该字段消费后的局部 CLI R6、Web tsc/lint R3 与实际 Vite 构建均通过，准确日志见 [独立前端验证](swa-web-separation-20261023.md)。后继独立服务的真实不可用 DOM 两例也通过；真实交接消费与 N4/N6 模型结果由总验收继续。
