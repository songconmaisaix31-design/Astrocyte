# SWA 任务卡执行

## 本次会话的最新决定

用户在首轮开发提交后明确：目标同时覆盖 Windows 和 Linux，Windows 架构为 x64，适配大部分 Agent；真实模型测试预算不设上限。允许人类按照当前项目权限独立授权后向 CLI Agent 创建的任务追加输入。这项新增允许仅对应跨创建者的 `worker.send`，不自动开放跨创建者 `worker.start`；人类自己创建任务的原有启动权限不变。

当前交互式环境为 Linux 云机，另有 Windows-2025 GitHub CI 用于真实原生检查；Windows目标架构为x64，没有提供可交互的Windows主机。不能把交叉编译或历史 Windows 报告记为实际验证。后继实现先推进 Linux 完整运行链和现有 Agent 适配器的双平台复用，逐项列明发现、配置、启动、关闭工具、输入、停止、成果的实际覆盖。上游 45 个 TUI 模板与 7 个 structured 原生引擎分别列出，不能把名称发现计为执行适配通过，也不自行将“大部分”冻结为某个人为数量。

用户新增执行要求：每次失败先交代最小复现、根因证据和局部验证结果，再启动下一次完整构建。局部探测与完整构建分别记录，根因未知时继续局部诊断；跳过完整构建不能记为通过。

预算授权不扩大原生工具、私有材料或账号范围；继续使用获准公开固定输入及 ToolsOff。实际只读检查显示本机 Codex 0.163.0 已报告登录，但未据此声称其产品原生执行链可用，未打印或复制凭据。新增人类准入由02实现、01维护合同、06独立审核，CLI Linux所属生命周期由03、双平台安装和依赖由05负责，04只消费真实后端权限和结果。

## 目标与基线

执行用户在本次会话提供的 `SWA-Agent-Task-Cards.zip`：CLI 优先，GUI 共用服务，最终多 Worker 验收默认 N=3、数量可配置。附件中的历史授权与运行结果作为待核材料，不作为当前云机的凭据或付费调用授权。

接管 main 为 `6445149358f5799fc5a786eb59ff494cddbec72e`，PR #5 实际 head 为 `6778a66faf6d1ca7b50658f23966db3e406f38c6`。工作树原为干净状态。集成分支 `integration/swa-task-cards` 在当前隔离云机使用原 checkout，保留 main 的最新交接文档并本地合入 PR #5 实现；不合并远端 PR，不发布产品。

## 负责人和写入范围

开发与独立审查使用当前 Codex 协作工具；本机未发现 Orca CLI，不声称已经通过 Orca 派发。各卡使用同一隔离 checkout 的互斥文件范围，提交由主控串行安排，不创建额外 worktree。

| 卡片 | 唯一实现范围 | 前置 |
| --- | --- | --- |
| 01 契约 | docs/integration/openmausbot、SPEC 对应章节、共同决定、适用 Schema 和生成类型 | 准确 PR 基线和参考源码 |
| 02 运行时 | runtime/swa-orca/src、internal/adapters/orca、Workspace 执行用例及其直接服务消费者和运行时测试 | 01 合同；既有明确缺陷可先核查 |
| 03 CLI | cmd/swa、runtime/swa-orca/USAGE.md、专属 CLI 测试 | 01 合同、02 服务 |
| 04 GUI | web 页面、API 消费者、组件、样式、GUI 测试；不修改生成 schema 和 package | 01 公共接口、02 服务 |
| 05 交付 | scripts/build-swa-*、安装与闭包脚本、依赖及锁、CI、交付说明 | CLI 候选依赖 01/02/03，组合候选再依赖 04 |
| 06 验收 | tests/swa-product-acceptance、docs/validation/swa-product | 05 实际安装物与合法当前授权 |

## 验证计划和停止条件

先由独立审查者审核各卡具体测试计划，再实现。运行 frozen 安装、Go module 校验、适用合同和类型检查、CLI 编译及相关真实测试、GUI 构建。检查必须保持真实服务、SQLite、请求回执和来源，不制造模型执行或通过结果。

本机 Linux x86_64 可运行 Go/Node 开发检查；Windows x64 安装、DPAPI、完整 Job 退出必须在实际 Windows 环境验证。旧机 PID、账号和过期 Grant 不移植。本机仅发现 Codex 0.163.0，尚未验证其通过产品的受限启动链；模型预算按本页最新用户决定，不扩大工具或私有资料范围。

正式验收要求同产品任务至少 N=3 不同 Dispatch/原生 session 的真实执行重叠，指定停止一个后其余 N−1 同身份继续，原 request 的幂等和效果未知正确处理，同 session 追加输入、完整成果、权限拒绝、完整所属进程退出。组合候选还要求 CLI/GUI 同一身份和数据的双向一致性。缺少外部平台、合法当前授权或真实输入时保留 NOT_RUN，并继续独立开发工作。

不把构建通过称为最终功能通过；不公开 Release、不部署、不 force push、不读取或复制凭据值。完成独立可运行工作后报告准确提交和剩余外部条件。

## 当前云机验证

- 集成基线提交：`6195f4cfb6b32aedd43845443c029ecc11d5578d`，保留两个父提交的来源。
- Go 1.27.2 模块下载与 `go mod verify` 实际通过；此前 `storage.googleapis.com` 代理拒绝已解除。
- Linux 实际安装采用四个开发目标的 filtered frozen 命令，690 条供应链策略通过；完整六工作区的首次安装因 Windows-native 工作区在 Linux 运行 node-gyp 而失败，原失败日志保留，未关闭安全校验。
- Linux CLI、标准 HTTP 服务构建成功。使用仓库外 `/workspace/astrocyte-cloud-data` 数据目录启动标准服务，实际检查 health、human session、SQLite 项目发现缓存含本 checkout、已构建前端页面，之后正常退出 0。此检查未配置 SWA 原生安装，不证明原生 Worker 执行成功。
- 独立运行 `go test -mod=readonly -json ./...` 退出 1，包含子测试的事件数为通过 1142、失败 154、跳过 61。完整原件位于 `.verification/cloud-setup/go-full-tests.jsonl`，逐项失败位于同目录 `go-failures.json`。该检查在各卡实现前段运行，不代表最终改动的完整回归结果，也不与历史 Windows/Linux 的失败名称数量混同。
- 失败包括缺少实际 OpenCode、公开 DOM/SQLite/receipt/publication 输入、原生启动提前退出及公开提供方不可达。直接 Go 命令未初始化标准 runner 的 `ASTROCYTE_TEST_OUTPUT_ROOT`，部分 S1 测试因此缺前置。独立审查另确认 SQLite 测试 `TestLocalProjectPermissionsTokensAndNativeReceiptsSurviveSQLiteRestart` 缺当前原生消息服务依赖，在未发生发送时触发旧断言；这不是重复执行的证据，也不证明真实幂等通过。
- 各卡的最终适用检查与剩余条件由其交付文档记录，最终安装验收入口见 [真实验收计划](../docs/validation/swa-product/README.md)。
- 使用标准 `pnpm check` 再次验证：gofmt、`go vet -mod=readonly ./...` 通过，设置标准输出目录后的 Go 测试仍退出 1，文本报告有 94 条失败事件，其名称均已在前段独立全量检查中出现。其后的检查因 runner fast-fail 未执行，适用的合同、strict、前端及 CLI 检查由各卡单独运行；不把失败名称相同当作原因相同或完整回归通过。原件为 `.verification/cloud-setup/standard-check.log`。
- 已实际运行开发脚本，验证 API、Vite 页面、Vite API 代理和正常所属进程退出，全部通过；复跑幂等安装脚本也退出 0。可复用环境草稿的安装及启动说明已更新。
- Playwright固定浏览器下载受到 `cdn.playwright.dev` 代理403拒绝；所需域名已在草稿中增量保存，不等于当前策略已应用。后续实际使用系统Chromium154执行页面检查，按真实浏览器与Playwright版本分别记录。用户已确认Windows x64、无限模型预算及人类独立追加输入语义，见本页最新决定；未提供可访问Windows主机。
- 本轮完整 `pnpm build` 退出 0，实际构建三个 Go 入口和前端；`pnpm check:architecture`、`pnpm check:dependencies` 均退出 0（45 个 Go 包的既有分层规则、精确依赖清单及 summarize CLI 版本一致）。前端体积警告保留。

## 分批提交与接受范围

实现使用当前 Codex 协作会话的单卡负责人和独立验收审查者；以下为本机开发与审查结论，不替代 Windows 安装及模型验收。

| 卡片 | 本地提交 | 当前可交付范围 |
| --- | --- | --- |
| 01 | `633d03a6784fc49198d4f69d11858a2cdbbb4cd3` | 共享任务文件合同、来源许可、身份状态和文件归属；264 例及生成一致性、17 输入边界通过 |
| 02 | `7830bdaaed14ee55a70dad2dbf848b440151497a` | 参数化总席位、未知预留、完成与启动互斥、可信组合网页目录；4 条生产规则、Go/文件系统/HTTP 检查及 strict 通过 |
| 03 | `18b57e289b62aeccb2c5aea4660f44f1c738325f` | 本地命令帮助、准确 JSON/退出语义、显式同服务网页参数与 N=3 调用说明；CLI 本机测试及 Windows 交编通过，真实安装测试跳过 |
| 04 | `b7447f4a1af3ef80a2fe921d561c70d86d46fe9e` | 共享真实 API 工作台、任务与 Worker 发起/观察/输入/停止/成果入口、原请求恢复；11 项 API/解析检查和系统 Chromium 页面边界 1 项通过，真实 Worker 场景未验 |
| 05 | `731d4da315171cee7e7dd7bf4e3250767629b635` | 显式 Python、strict 入口、CLI/combined 构建与真实资产/许可证闭包检查；Windows 候选尚未构建 |
| 06 | `46f7bb7ad1cacda7d63045257ced732e1b5f3a4a` | 真实验收计划、只读前置检查和既有失败分类；里程碑 A/B 均 NOT_RUN |

## Linux基底与新增权限的实际结果

- `1accc19e`允许当前项目获准human独立执行跨创建者`worker.send`，保留跨创建者`worker.start`与Agent原Grant规则；真实SQLite三组共25子场景及race通过。`1518645b`保留GUI显式控制设置并刷新追加输入权限；实际HTTP/SQLite与系统Chromium检查通过，真人到原生Worker链仍未验。
- Linux CLI所属生命周期为`d2a10dc9`，构建、私有安装与依赖为`3dad4b36`，公开N输入准备为`66a55a77`，运行时身份/私有目录/daemon基础为`1961dfba`。真实proc、pidfd、subreaper、收养后代/僵尸与拒绝检查，以及适用Go、strict、Windows交编分别通过，不外推为每Worker或模型验收。
- Linux候选01在部署闭包因pnpm启动器符号链接退出1；正式普通启动器配置`559c621b`修复。候选02构建0、安装在平台空`.git`误判时退出1；`f71eec93`修复。候选03构建0、安装因fs.cp对新建空目标的语义退出1；`e14800c1`修复逐顶层entry复制。各失败日志及partial均保留，没有改旧候选或放宽拒覆盖、保护目录及无链接规则。
- 准确`e14800c1b207a8c063531fedf014aa0cd2374558`的CLI04标准build、普通归档SHA核查、仓库外解包和严格安装全部退出0。安装根`/workspace/swa-linux-installs/swa-linux-cli-e14800c1-04`；实际Node24.21.0、PTY输入回显/exit0、watcher create及28253路径闭包通过，root0700/文件0600或可执行0700/当前UID。证据在`.verification/swa-linux/evidence/e14800c1b207a8c063531fedf014aa0cd2374558-cli-04/`。
- 相同安装的实际服务证据为`/workspace/swa-cli-smoke-e14800c1-1792661325429/smoke-evidence/result.json`：runtime `746084d0-a595-4690-98d7-d711925c9d2d`，service与process_tree PASS，HTTP health=ok，正常stdin关闭code0。status/controller因缺少当前合法项目Grant为NOT_RUN，不自动取得human会话或签发Grant。
- 每Worker Linux独立完整回收D由`2df6f223`实现，固定node-pty patch/构建/API probe为`9baa2ed5`。真实旧ABI floor失败最小复现、admission观察错误、ACK失败与scope失效的审查发现及修复结果已先报告用户；局部14项全部通过，严格检查通过后才启动后继组合完整构建。详细原件见`.verification/cloud-setup/linux-provider-tree-failure-notes.md`，不把静态发现描述成已动态观测的旧错误释放。
- 准确`6890522e43d7b81e99b09ca94acc27b9d3a5c7dd`的combined01标准构建、归档SHA、仓库外普通解包及严格安装均0，安装根`/workspace/swa-linux-installs/swa-linux-combined-6890522e-01`。原checker检查28307路径闭包、网页清单2入口/4引用及真实PTY/watcher/kernel probe；同安装D检查14PASS0skip，11项使用装内Tree/managed/entry/nodepty原件，3项同SHA诊断，actual bundledNode/VERSION/CLI版本/cleanHEAD守卫均通过。06独立核验4原件candidate/install SHA一致并限定接受安装和普通进程托管；组合serve/网页HTTP、GUI、AgentNativeLaunch/账号/模型/N=3不包含在内。原件在`.verification/swa-linux/evidence/6890522e43d7b81e99b09ca94acc27b9d3a5c7dd-combined-01/`及`.verification/cloud-setup/linux-provider-tree-installed-6890522e-01.log`。
- 同组合安装的03服务首次检查退出1：health与首页200，但ignored脚本将标准`/index.html`→`./`的301误作失败。用户已先收到最小复现、Go固定源码根因和局部首页字节一致结果，再修正检查脚本；未改产品、未重建。局部复验退出0，runtime `ced236d2-b0c3-49a8-ab78-93295c28f7aa`，endpoint18793的health200、首页562bytes/4个JS-CSS资产逐字节及MIME一致、301严格同源；stdin正常关闭code0，内部双owned0/全部pidfd退出，外部4身份退出及两次完整空扫描相隔53.4ms。证据`/workspace/swa-cli-smoke-6890522e-combined-1792664829113/smoke-evidence`，原失败data1792664746641保留；没有human会话、Grant或模型调用，不代表GUI业务或Worker验收。
- Codex0.163完整session结果读取为`0d0cbaf3`；官方私有空HOME无模型metadata04通过，不代表账号绑定或模型ToolsOff执行。原ChatGPT认证不能据当前process-local锁证明多个私有HOME克隆的刷新安全；后继采用官方stdin API-key登录、独立HOME、固定官方模型目标及有效配置检查计划，缺少真实安全配置key时准确保留未配置。真实N验收驱动为`6890522e`，公开固定项目和输入准备、help及非法参数检查通过，真实Grant/原生Worker/模型场景仍NOT_RUN。
- GitHub实际run `39747808189`的准确`faea9d4a`基础matrix在Ubuntu和Windows均通过；完整check仍失败。Windows原生run `39757398493`因preflight失败；`39758336431`的准确`0209d004`诊断取得PATH多Application Source被拼为单命令，`3f4a4ea2`修复三个选择为First1并保留版本要求。后继实际run `39758793164`进入build后失败。只读诊断run `39760430188`（来源`8d02edd5`）成功读取该失败artifact，完整native任务跳过，原node-gyp报VS探测的`ERR_CHILD_PROCESS_STDIO_MAXBUFFER`及无法找到VS，尚不能证明机器未安装VS；05下一步仅局部对比原finder与官方Developer Shell，不运行完整产品构建。此链路没有取得账号或调用模型。
- VS局部probe为`369680ec88d24a33e00d535b6c3eda0041809d1d`，已普通push。实际run39763935502/job119774602903的setup/pnpm成功，probe运行19秒退出1，artifact上传成功；完整native、prior诊断及旧check/foundation跳过。无baseline摘要，不能声称已动态复现VS探测问题。root核原下载helper的47—51行发现probe的RUNNER_TEMP工作目录必违反固定candidate工作根约束，随后还漏查下载exit；06独立确认并接受最小返修计划。用户先收到这一轮最小执行、证据不足和确定静态调用缺陷，再做合法独占目录/下载exit/有界phase诊断的局部修正；原helper、TLS/SHA和完整构建门均不放宽。下一轮局部probe须先实际验证原非法参数的stage=validate/ArgumentException拒绝，再取得原finder与Developer Shell真实结果。
- `bcba96e2`局部run39764771846实际通过上述非法目录负测和合法下载/哈希，随后因查找VS17而失败；真实C#观察只有VS18.10，原builder明确msvs2022，原node-gyp12.3支持VS18→2026/v145。`2c023886`局部run39765889558进一步得到原2022选择false/mismatch=true、仅2026选择true、官方DevShell2026/v145=true；同一node-gyp源码与Node24.21，未升级依赖或改buffer。原VSSetup命令默认execFile在1048576字节报MAXBUFFER，stream实际1451733字节/1694packages，C#78312字节0。该轮job整体仍FAIL，结果文件完整输出，cl/Bv无输入exit2是待核收尾原因，不把实际局部成功等同job成功。05/06继续局部退出分类和产品工具链选择计划，完整Windows构建未重开。
- `4c79625d`实现Codex/Claude显式独立API来源目录，原普通/发现/default账号、产品根、保护根及彼此source双向隔离；仅读路径元数据，定向Go/真实文件拒绝及Windows交编0。`3b32fae6`新增Codex strict NativeLaunch合同，94 Ajv/TS形状边界及264示例/生成一致性0，Go与实际运行消费者另验。`ac668ca8`随包复制3件未修改官方catalog/LICENSE/NOTICE并核既有inventory，局部缺件/重复/缺pin/源不可用检查0；原689安装保持不变，新运行时仍待完整strict/新候选。
- Codex账号模块局部检查首两次bundle因Linux没有Win .node、未指定生产tsconfig失败；只修ignored入口的external/tsconfig设置后，两产品模块bundle与12项真实无key文件负测通过。没有伪造key/auth/native响应，新增空账号预检接口仍待统一strict与最终源码复验。详细`.verification/cloud-setup/codex-policy-review/account-implementation-result.md`，真实登录、账号导出、原生模型与N=3均NOT_RUN。
- 当前无认证HEAD到`https://api.openai.com/v1/models`、`https://api.anthropic.com/v1/models`均在CONNECT阶段403，未发送账号或模型输入。云配置草稿已保留原自定义3域并增量追加`api.openai.com`、`api.anthropic.com`；保存已确认、requires_publish=true，当前策略未应用。新增网络范围不代替真实安全key来源和human项目Grant，不绕过代理或另换目标地址。
