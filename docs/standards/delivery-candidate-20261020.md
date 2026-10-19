# 当前产品候选与真实运行

## 准确来源与交付范围

运行与验证源码 **S=`ed543c12106c2dbb3e181cb9bc1647b626455179`**。三个Go程序和Web从同一干净Git源码执行普通构建，实际Go元数据均为完整S、`vcs.modified=false`。主控已将S程序与Web部署到原64505，保留原SQLite、Q与媒体配置；独立Reviewer已接受同源产物、迁移副本和实际原库运行对应。

| 项目 | 准确来源 |
| --- | --- |
| Worker / Run | `task_6f09f37378a8` / `ctx_1e0467a3f24a` / `run_1187287582ea` |
| 工作树 / 分支 | `C:/Users/DW/orca/workspaces/Astrocyte/frontend-observatory-20261018` / `songconmaisaix31-design/frontend-observatory-20261018` |
| 接管基线 | `5725c1840a760105518d273eaaa28d7367f369ed`，当时tracked干净 |
| S与计划审核 | 准确S的[计划原文](https://github.com/songconmaisaix31-design/Astrocyte/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/docs/standards/delivery-candidate-20261020.md)经独立接受后冻结；诊断与迁移观察时间补充计划分别独立接受 |
| 产品主线 / PR | `songconmaisaix31-design/works-relay-integration` / [原PR #2](https://github.com/songconmaisaix31-design/Astrocyte/pull/2)，由主控普通推进 |
| 当前文档后继 | S之后仅本报告及获准治理文档；以`git diff S..<报告HEAD> --name-only`核查业务、契约、锁和构建输入未变，不为报告重建 |
| 客户端 / 模型 | Orca Codex；本次实际终端信息为`GPT-6.1-Sol xhigh fast`，无模型覆盖或子Agent |
| 主控事实源 | [42570f89](https://github.com/songconmaisaix31-design/Astrocyte/tree/42570f89d38e5fe7ef8d63067de4301f6e5df381)及[当前交付任务](https://github.com/songconmaisaix31-design/Astrocyte/blob/validation/v1-20261012/tasks/delivery-candidate-20261020.md) |

本次交付为普通产品候选。完整便携模型包、完整默认CI、原生/S3、实际采用和正式发布分别保留原条件。正式实验、论文、原生工具授权及发布门禁没有修改。准确报告Commit的COMMENT见[原PR审查记录](https://github.com/songconmaisaix31-design/Astrocyte/pull/2)，普通合入见主控交付任务。

## 默认检查的准确状态

`.github/workflows/check.yml`默认矩阵为Ubuntu24.04和Windows2025：Go模块下载、冻结pnpm安装、`pnpm check`、Linux全仓race、`pnpm build`、Chromium安装及对应默认Playwright。`scripts/check.mjs`先分配真实输出根，再依次执行gofmt、全仓vet/Go测试、模块校验、依赖/架构、stop-process测试、契约/示例/生成一致性、S1、tsc、eslint、Vitest及Git空白检查，首个失败终止后续步骤。

| 范围 | 状态与原因 |
| --- | --- |
| 当前S完整默认CI / `pnpm check` / 默认浏览器矩阵 | **NOT_RUN**；默认集合含当前禁止的旧替身及未满足权限/真实来源前置的原生路径，没有启动该集合 |
| 当前`pnpm check:contracts` | **FAIL**；Redocly输出API有效与13警告后发生Node `UV_HANDLE_CLOSING`断言，退出`-1073740791`（无符号3221226505）；该次后两个`&&`阶段未运行 |
| 经独立审核的同版本实际Redocly直调 | **FAIL**；相同断言；未定位完整根因，没有重试原入口、升级依赖或修改规则 |
| 原契约示例 / 生成一致性独立直调 | 分别**PASS**，264项示例与生成一致性实际完成；不替换原入口失败 |
| 历史982默认检查 | 准确`982333b871520e606c74d553e448c5049f81abde`的push run39155403418 / PR run39155406045最终cancelled；两个Ubuntu job的全仓Go测试在取消前已FAIL，后续未完成阶段不计通过 |

历史失败原件为旧独立审查树`dynamic-dag-review-20261019/.verification/dynamic-dag/ci-982-{push-job117757474469,pr-job117757483047}-failed.log`。已逐类核查输出根、26/27迁移、discovery调用条件、Relay reconcile批准缺动作、旧替身及公开资料/CLI/Q/原生来源缺失；没有将全部失败归因于环境或新回归。

Source22 `22cde9efc540553c433463af5b9e955790d57a99`的输出前置修复至接管HEAD只含文档后继。原OpenViking1函数与ProgressPreflight2函数、无参数默认根/显式父根、相对路径/实际普通文件拒绝结果在原准确条件内复用，原件为本树`.verification/restart-product-20261019/ci-{before,openviking,progress,invalid-roots}-01/`；不重复before/after克隆。该修复不表示workflow另行race已分配输出根，也不补齐其他真实来源。

本次没有改workflow、业务、依赖、原断言或添加测试Skip。纯文档后继沿主控已确认的`[skip ci]`提交约定普通push，保留当前默认状态。

治理文档通过已安装`marked 18.1.0`核查：15份Markdown、56个本地文件链接PASS，当前根AGENTS/CODING与四角色原文及角色目录内容一致。实际命令为`node E/check-documents.mjs document-check-final.json`，BelowNormal，`NODE_OPTIONS=--max-old-space-size=448 --max-semi-space-size=16`，实际V8 heap上限496MiB；原件`E/document-check-final.json`。`git diff --cached --check`无空白错误。历史STATUS/HANDOFF等以准确S的Git原文保留，缺失管理引用连接主控42570f89或validation分支，不复制全仓管理历史。

## 实际命令与结果

以下均针对S。Worker原件根记为 **E=`本树/.verification/delivery-candidate-20261020/`**。实际入口`pwsh -NoProfile -File E/run-check.ps1 -Phase <阶段>`固定S与干净Git状态，每项保存`run.log`、`check.log`（有输出时）、`result.json`及资源原件。没有业务改动的历史性能与Source15完整包结果沿原准确来源复用。

| 实际命令 | 结果与需求覆盖 | E下原件 |
| --- | --- | --- |
| `pnpm build` | EXIT0；普通server/mcp/sleep三程序、完整tsc与Vite；入口719784B、DAG分块57129B，既有500kB提示保持 | `run-build/` |
| `go version -m`三个程序；`stage-artifacts.ps1` | 三者Go1.27.2、完整S、modified=false；38项程序/Web/启动入口复制逐文件完整字节一致 | `all-go-metadata.json`、`artifact-copies.json`、`artifact-resources.json` |
| `go test -mod=readonly ./internal/swarm/domain ./internal/swarm/app ./internal/adapters/sqlite ./internal/adapters/httpapi -run '^TestDynamicDAG' -count=1 -timeout=2m -v` | 9函数PASS；空图/环/重复/缺端点/跨Mission、未知用量与已知超限、错误分类和历史确认编码；真实HTTP/SQLite人工增长、逐事件SQL、read_swarm/跨项目/撤权、无Worker停止及过期批准历史 | `run-dag/`，实际HTTP库在本树`.verification/dynamic-dag-http/actual-*` |
| `go test -mod=readonly ./internal/attention/domain ./internal/adapters/openviking ./internal/attention/app ./internal/adapters/sqlite ./internal/adapters/httpapi ./cmd/server -run '^TestOpenViking' -count=1 -timeout=2m -v` | 4函数PASS；URI/绑定/配置规则，实际HTTP/SQLite未配置/连接拒绝/空集合/Origin/CSRF/缺失与null层级拒绝；无匹配函数的包只计编译，零远端提交 | `run-openviking/` |
| `pnpm --dir web exec vitest run --maxWorkers=1 src/components/displayContract.test.tsx src/utils/__tests__/format.test.ts src/utils/__tests__/commandError.test.ts` | 3文件53项PASS，显示/格式/错误规则，不计API或原生正向 | `run-display/` |
| `node scripts/check-architecture.mjs`、`node scripts/check-dependencies.mjs` | 分别EXIT0，43包既有边界及实际依赖清单；架构使用Go1/-p1、Node512专用额度 | `run-architecture/`、`run-dependencies/` |
| `pnpm --dir web lint`、`go vet -mod=readonly ./...` | 分别EXIT0；下节记录实际重叠，不能登记为符合串行条件的独立运行 | `run-lint/`、`run-vet/` |
| `pnpm check:contracts` | FAIL；真实工具退出断言，未执行其后两个链式阶段 | `run-contracts/` |
| `node node_modules/@redocly/cli/bin/cli.js lint contracts/openapi.yaml --config contracts/redocly.yaml` | FAIL；同版本实际直调仍崩溃，13项警告原样保留 | `contracts-diagnosis.md`、`run-redocly/` |
| `node scripts/validate-examples.mjs`、`node scripts/generate-client.mjs --check` | 各自EXIT0，264项及生成一致性完成 | `run-examples/`、`run-client/` |
| `pnpm --dir web exec playwright test --config dynamic-dag.config.ts --workers=1` | 4/4 PASS，2.0m；1920×1080与1280×720，真实增长/事件SQL/键盘方向缩放/长标题、断线stale与同库重启、撤权清图/离页停止、真实缺分块及原context人工reload恢复 | `run-browser/`及`evidence/`，Playwright输出在本树`.verification/dynamic-dag-browser-*` |
| `pnpm --dir web exec playwright test --config observatory.config.ts --workers=1 e2e/observatory.check.ts` | 8/8 PASS，20.7s；双视口页面/批准与作品入口/搜索键盘、固定正文与可选未配置、26品牌真实HTTP/decode | `run-regression/`及`evidence/` |
| `sqlite-data.py backup`、`candidate-service.py rehearse --output E/rehearsal-01` | 成熟SQLite一致备份通过；真实候选进程PID4052执行44次HTTP与全部Web字节比较、Q指针保持后停止。首次全值比较因既有启动刷新observed_at而EXIT1，原件保留 | `run-backup/`、`rehearsal-01.log`、`rehearsal-01/` |
| 经独立接受的`sqlite-data.py compare --migration <准确027> --launch E/rehearsal-01/launch.json` | 只读复核既存副本EXIT0；准确027增量和所有原业务值/类型/重复数保持；仅registered观察时间在真实进程至复核区间更新，其余JSON含数字类型全等。全值相等标记保持false | `migration-observation-plan.md`、`discovery-difference.json`、`rehearsal-01/migration-reviewed.json`、`migration-reviewed-resources.json` |

上述用例来自原前端/DAG任务、SPEC §10/16/17及当前保全目标。实际人工增长与无Worker停止不会登记为模型生成或Worker退出。浏览器截图、视频和trace关闭；没有伪造CLI、响应、会话、Claim或租约，没有调用新模型。

## 资源与窗口

Worker42个实际启动样本最高CPU46.4%、最低可用9.770252GiB，单调间隔均2–5秒；Go4/-p4/512MiB、Node≤1024MiB、BelowNormal，架构单独使用Go1/-p1/Node512。没有资源拒绝，没有测运行期峰值。全部原件保留。

执行偏差：lint未退出时启动vet；vet于UTC19:31:37–19:31:45运行，lint于19:32:03.790退出，不符合全局串行要求。两项实际退出和额度保留，已报告主控并由Reviewer核查，不为删除偏差重跑；此后逐项确认前项退出再启动。该偏差不计为符合资源规则的证据。

Worker于UTC19:38:24.506的`state-before-handoff.json`确认自有server/astrocyte/浏览器0、不可读同名进程0、64505/17887无监听；随后只有只读SQL复核和Parser，无新服务。`msg_bf9a9e856ef9`正式交窗，主控`msg_fba490d216e3`实际接收。原64505由主控随后启动，独立Reviewer复用同一实际服务，未开第二服务。

主控stage/启动4个样本最高CPU42%、最低可用10.553741GiB；`original-01/{prepare-resources,startup-resources}.json`记录实际UTC及约3秒间隔，服务为BelowNormal/Go4/-p4/512MiB/Node1024。独立Reviewer的样本及更小读取额度分别在其原件中记录，不移用为Worker结果。

## Windows产物、原库与实际服务

本树普通交付目录为`dist/delivery-candidate-20261020/`：`astrocyte.exe`26579456B、`astrocyte-mcp.exe`12374528B、`astrocyte-sleep.exe`12489728B、完整Web、原Windows启动入口及本候选说明。33件Web与本次构建完整字节一致。原`dist`和`web/dist`保留，已忽略目录没有用户数据库或凭据进入Git。完整新便携模型包未构建或验收；Source15 ZIP与130项结果仍只属于原准确Source15，见[既有产品记录](s3-ten-cycles-20261019.md)。

| 运行项 | 实际结果 |
| --- | --- |
| URL / PID / 创建时间 | `http://127.0.0.1:64505/` / `45956` / UTC`2026-10-19T19:47:04.657217Z`；唯一`127.0.0.1:64505`监听、BelowNormal |
| 程序 / Web | `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/versions/ed543c12/astrocyte.exe`及同目录`web`，来源均S |
| 原SQLite | `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite` |
| 一致备份 | E下`original-01/original.sqlite`；Python标准`sqlite3.Connection.backup`从mode=ro/query_only原库创建；启动前与原库全列/类型/重复数量相同 |
| 迁移与保全 | 原68表/93行/26迁移；准确新增`attention_openviking_submissions`空表/唯一索引与027迁移记录。所有原业务值保持，integrity/foreign_key通过。唯一registered `data.observed_at`由`12:09:40.7714069Z`更新为`19:47:23.1062105Z`，处于实际启动/核查区间，其余JSON严格保持 |
| Q与媒体 | Q完整指针`generation=snapshot-1711995690`、`snapshot_at=2026-10-14T18:48:42.2097174Z`保持；原12项可信媒体配置沿用，外发/资料根未扩大 |
| 运行参数 | `ASTROCYTE_PORT=64505`、Relay=1、session source=archive、原data目录、同源web目录、Workbench=1、原Q reader root、原公开项目/import roots、原repository environment；Codex沉淀关闭，Laya/Jev/OpenViking未新增配置。服务继承原可信媒体路径；没有输出凭据 |
| 主控命令与原件 | `C:/Python313/python.exe -I -B E/candidate-service.py stage --output E/original-01 --execute-original`，随后同参数`start`，均EXIT0；`original-01/{stage.json,launch.json,go-version.txt,backup.json,before-start.json,http.json,migration.json}`及原运行根`running.json` |
| 持续运行归属 | 独立服务终端`term_c043b5a9-d4f4-41d0-b387-866ed544ca6d`保持；Worker结算不关闭服务 |

原c5程序、172 Web与配置保留。恢复入口为经审查的`candidate-service.py stop/restore-data --execute-original`及原c5恢复入口；必须先核新进程身份和退出、保全失败数据库，确认没有其他业务变化，再由主控显式恢复一致备份和旧程序。没有自动恢复或忽略差异，恢复阶段本次未执行；如果候选运行后产生合法新数据，须先保全并交主控决定。

## 独立结论与接续

独立审查树`research-preprint-review-20261019/docs/standards/delivery-candidate-review-20261020.md`记录完整依据：`predeploy-01/result.json`于UTC19:44:14.576通过同源产物与迁移副本核查；`live-01/result.json`于19:49:22.543通过实际原服务核查。live实际命令为`pwsh -NoLogo -NoProfile -File .verification/delivery-candidate-20261020/review-live.ps1 -ReportSHA S -ExpectedPID 45956 -StageEvidence <本树E/original-01>`。

独立live核查49项逻辑GET（含无会话、非法Host、跨Origin会话、跨Origin资料、非法Agent凭据5项403）、33件Web全部字节、实际materials2与local-projects1的ID逐项对应原SQLite、完整Go来源/PID/路径/监听、027与原业务/Q保全。Reviewer只有GET与只读SQLite，没有启动服务/浏览器、编译或模型。实际运行接受hook`msg_1ec86dc76c63`已经主控实际读取并回复；准确Commit的正式COMMENT与结论保存在原PR审查记录，合入结果由主控交付任务记录。

本次未通过范围：完整默认CI与默认浏览器、Redocly正常退出、lint/vet原串行条件、真实原生/S3、模型规划/Claim/Send、带Worker停止退出/未知用量/资源释放、两Agent闭环、真实成果采用与经验复用、OpenViking真实远端、完整新便携模型包及正式Release。原S3预算/权限证据、受限Codex、具体资料/经验准入与采用等未答条件保持；Worker未合并分支、创建tag/Release或修改正式实验，产品主线由主控普通推进。

下一项真实产品任务为**修复Windows契约检查进程崩溃，使原`pnpm check:contracts`完成**。依据本次两个实际FAIL，入口为根package命令、真实Redocly、`validate-examples.mjs`与`generate-client.mjs`。原Worker先提交根因方案、具体写域和测试计划独立审核；验收为原入口EXIT0、三个原组成检查全部实际完成及Reviewer正常/失败退出复验。当前只登记在[TASKS](../TASKS.md)，不预选依赖版本或开始新实现。
