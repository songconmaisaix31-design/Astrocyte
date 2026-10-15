# 基础后端与导入整改

本轮状态：阶段源码已提交，整轨验收未完成。分支 `standards-backend`，基线 `4ec19cddb24068e09597357b0aabc6ecb9420930`，本次恢复 HEAD `ddfae290e3b0cf468af9aeaa98c5bdf39ea1604d`。完成收据阶段 SOURCE `c715342cff4ccfa1b3198275f981385296fe6641` 已普通 push，实际业务来源、模型解释与人工批准后使用仍未验收。第一阶段 SOURCE `2644083960649d7c50262ec69fff730b411241a2` / REPORT `6a4415faaa1e092ae1eab42cffed4f3ffa090434`，第二阶段 SOURCE `7f58adecc41070fbec4f88962cfb6fb1723679c6` / REPORT `a3f093ddd3488df9c71f2ad295724dba06e92efb`，依赖 SOURCE `2a6f03605ebb87f36f7bd56c577aaf4853cb6671`，第三阶段 SOURCE `6cfadb36d94ab372eee1ef1f1f687947cbb771fa` / REPORT `12b82723509d45e2dfd712a74386523e5c38014f`，HTTP 命令及错误编码 SOURCE `b470ecd736af261a10e4c4822bdcd3b02d0da454`，Workbench Reader SOURCE `f6491131d5e5e6eeb889ec82362caf9f1db04e40` 均已普通 push。早期生产服务隔离 SOURCE `86d8f910eb5d6641f84bcc248d86444fb7bd81cb` 与作用域契约 SOURCE `e1d0f28f75a3751da5fa39b7d69f9e0ac4763f39` 保留。

后续独立 REPORT `94658c3788fc94db8215df403bea6aab4da1cfbb`、`96618df9b62ed0018182b5cbc1aa626c5bcb8362`，文本目录及实际 Works 发现 SOURCE `01753cf4fb455f35aa4953bf3eaa9a6388c4be5a`，公开生产认知及祖先路径检查 SOURCE `54ad5d71d20bdd3a1d2e9c6ae55b67e784700344`，默认时钟/实际 Sleep/真实 nullable 及公开论文输入 SOURCE `e914ff9bb845d1e7b3ab13d1b1721e1598bafbe7` 均已普通 push。

后续 SOURCE `3384a89aac3be6d5670d756eb2d44263784d0cb8`、`03f92127a74e4d8b49cf5c4bb1010b43349d6d43`、`e05c013a08ccf372bb884dae5eeb3345f92173af`、`5622bfdaec376e23a696cc110a082c72c9d43f3f` 均已普通 push；前一独立 REPORT 为 `15c738f88822e553eb228c12c0fdc71ad07b7531`。整轨及实际模型验收继续保持未完成。

## 当前生产接口与验证

- 按唯一 R owner 的零 WIP 授权，精确消费 `0de4e464e3dab342a06fdb37d76ea58a8d5f1f32` 中单文件 `relay_execution.go`；原 identity、进程和实际 launch 参数检查保持，effectiveToolSet 为 unknown。临时单文件写域已归还 R。
- 真实论文输入具有生成契约类型；`loadPublicPaperSnapshots` 经实际 Node DOM 返回完整 canonical snapshot，包括原 schema、作者、来源及版本。PLOS 完整正文与 ACL 摘要经实际生产 HTTP/SQLite 导入，附件原 bytes 和投影相等。`.verification/public-snapshot-api-first.log` 保留第一次字段选择错误，第二次实际检查通过，证据 `.verification/public-snapshot-api-second.log`。
- `node scripts/check-s1-project-api.mjs` 使用两个实际公开 arXiv 版本，实际接口、固定引用、权限、撤销、调用键复用和重启通过，证据 `.verification/public-project-api-first.log`。类型声明的 web TypeScript 检查、11 个 Node 测试、31 包架构检查通过；证据分别为 `.verification/public-paper-types-second.log`、`.verification/public-snapshot-types-first.log`、`.verification/node-comment-and-parser-first.log`、`.verification/architecture-module-json-first.log`。类型检查首命令错误仍保留。
- P/F 明确授权的三文件 Reader 临时冲突处理均只编辑并 stage 原 e914 不可变内容，三文件逐字相等且无 unmerged entries；其 merge/commit 归原 owner。P Reader 十项实际检查通过，P 旧装配常量导致首 build RED；F 对应实际组合 server build 通过。证据 `.verification/p-conflict-reader-first.log`、`.verification/p-conflict-build-first.log`、`.verification/f-conflict-build-first.log`。两项临时授权均已结束。
- 文本处理原调用处把 project/job/operation 关联到当前精确 controller 身份、原 native_id、PID 和实际 Send turn_id；官方 `thread/read(includeTurns:false)` 元数据及实际原件窗口只保存在已有本机 stateRoot，产品 DB/HTTP 不增加 raw 字段。首次 turn 前仅记录来源路径回执；Send 尝试与受理分别登记，停止前按原 header 的 native_id/cwd 核查并复制实际窗口，窗口完整度保持未核实。
- `go test ./internal/adapters/agents -run 'TestSelectedText|TestOwnedTextExecution' -v` 实际无模型探测第二次 4 PASS，9.708s，原 controller StopConfirmed，model turns 0，effectiveToolSet unknown；证据 `.verification/text-source-actual-second.log`。首检查错误地要求首次 turn 前原件已创建，原 RED 保留于 `.verification/text-source-actual-first.log`。`go build -o .verification/text-source-server.exe ./cmd/server` 与定向 `go vet` 均 EXIT 0，证据 `.verification/text-source-build-third.log`、`.verification/text-source-vet-first.log`。编译首 RED 与更正后通过均保留。
- 用户经协调者决定将实际 selected-text 模型/原生身份完整核查窗口移至 2026-10-15；今晚不再启动原生 controller。当前 Send 后原件窗口、原生文件 alias 守卫接缝和 v1.4 身份契约仍待验证，未登记 CLI 完整验收 PASS。
- 视频固定 owner V 已接收 extractor.go、extractor_test.go、summarize.go、summarize-media.mjs、importers_test.go 五整文件，B 持续停写。T 测试域保持原已交接范围。
- 当前 Windows 已配置匿名 loopback 代理实际指向 `127.0.0.1:7890`；经明确授权只读取得实际配置，原 arXiv 官方 API 通过该代理实际 HTTP 200，13717 bytes，证据 `.verification/arxiv-configured-proxy-first.atom.xml`。未修改系统网络、SSL 或 DNS；生产显式代理配置仍待消费 V 精确 SOURCE。

上述定向通过未覆盖当前全包 RED、尚存替代实现及 favorites 最终容量验收。视频真实时间区间及私有附件读取结果分别由 V 和下节实际检查登记，阶段报告继续保留其他限制。

## 原生视频附件权限与显式代理

SOURCE `cc46b9855775adf70bce74242d08ef84ff4f573b` 已普通 push，包含普通 `--no-ff` 消费的 V SOURCE `06372a6a82c2942e02c4a46b69cc3f518f4ec5ae` 与 `d48a29b6e1744465666edbfa28be919531ca39a4`。前者实际匿名代理接口四种环境隔离检查，后者已保存实际同公开视频 whisper 原生时间区间；其历史 REPORT `478e8909fe6e2836b434b5d528348bf388372bfe` 随原普通祖先保留，未覆盖 V 写域。

生产配置读取 `ASTROCYTE_SUMMARIZE_PROXY_URL`，默认空；固定 Node/summarize 启动环境只接收明确批准的匿名 HTTP(S) origin，继承的代理、凭据和 `NODE_OPTIONS` 不进入提取进程。实际 Node 24.16.0 的安装帮助和 [官方 CLI 文档](https://nodejs.org/api/cli.html#node_use_env_proxy1) 已核对 `NODE_USE_ENV_PROXY=1` 与 HTTP/HTTPS/NO_PROXY 的启动语义。`go test ./internal/adapters/importers -run TestVideoExplicitAnonymousProxyIsolation -v -count=1` EXIT 0，7.856s；证据 `.verification/proxy-isolation-consumed-first.log`。

原生 whisper JSON 使用已有 MediaType 字段的 `application/vnd.astrocyte.private-native+json` 类别，本机 ObjectStore 保留原件。`GetAttachment` 检查所有 revision 中同一不可变 ObjectRef 的私有类别，已有实际 `whisper-original.json` 作为既存原件识别；同一对象的名称或 revision 别名不能开放 HTTP 读取。冻结 schema、领域及迁移没有改动。

`ASTROCYTE_TEST_VIDEO_PUBLICATION=<V 明确实际出版目录> node tests/s1/private-video-attachment.mjs` 第三次 EXIT 0，实际媒体 id `534HWFW7E5GEF75ZY4F7ETPLQ2`、revision 2、952 个原生 transcription segments。只读 V SQLite 经成熟 Node backup 创建 B 自有副本，按真实 revision 中六个 hash 引用逐件 `COPYFILE_EXCL` 复制对象；没有新 ASR。实际人类及配置 Agent 身份、普通及 URL 编码路径、冷热进程累计 8 次 HTTP 403 `scope_denied`；其他三个公开附件累计 6 次 HTTP 200 且 bytes 全等；本机原生 JSON bytes 完整保留，拒绝读取没有生成成功 machine use，真实 HTTP 进程重启通过。证据 `.verification/private-video-http-third.log` 与 `.verification/astrocyte-s1-FAkZe5/`，全部原件仅在本机。

首复制检查在已有目标目录失败、第二次 ErrorV1 字段选择断言失败均保留于 `.verification/private-video-http-{first,second}.log`；没有放宽断言。当前 `go build -mod=readonly ./...`、`go vet -mod=readonly ./...` 均 EXIT 0，证据 `.verification/private-source-{build,vet}-first.log`。实际完整 Go 套件仍为前次 RED，未因定向权限通过而升级；新生产 private MediaType 的独立后续消费由 V/T 原 owner 继续检查。

旧 README 实际结果使用原 `.verification/astrocyte-s1-opvXPA/data/state.sqlite` 启动当前生产 HTTP，只运行 GET session/progress；结果保持旧 native_id、model 与 evidence，新增模型 turns 0，旧 effectiveToolSet/source completeness 仍 unknown。F 已收到只读 URL 与项目 id，专属服务等待其检查后停止；证据 `.verification/existing-progress-http-first.log`。该服务不接续原 native 会话。

F 已完成旧结果两个 viewport 的只读 UI 检查，2 PASS；专属 53863 服务已停止，实际 health 请求 `ECONNREFUSED` 确认，`.verification/existing-progress-stop-first.log`。旧实际 DB 的进度及 accepted receipt 都无 summary；当前 prompt/parser 仅读取 status/percent，summary 在现契约中可选。只读聚合 `.verification/existing-progress-shape-first.log` 确认没有字段丢失，未补造摘要或重发模型；原 frontend 404 与要求非空摘要的首 RED 由 F 保留。

## 实际临时目录与 PDF 格式检查

SOURCE `d74c007d31dbcc73b9c86335525953760b0490b7` 与 `9292620ec4c6b442f598be6702000ce0894f6c0e` 已普通 push；其间普通合入 V 精确 default SOURCE `a54fe8e238a2e52db98c801012c2ccdfd64edbd5`，merge `31a7e10` 保留原历史，未代改 V 的五个整文件。

`startS1Server` 为子进程显式设置本工作树专属 `temporary/tool-temp` 的 TEMP/TMP。`node --test tests/s1/server-lifecycle.test.mjs` EXIT 0，两个实际普通/Relay 目录布局分别运行 HTTP、SQLite、权限、停止、重新启动与原目录再打开，2 PASS、0 SKIP，43.648s；`.verification/helper-owned-temp-first.log`。V 默认输入依赖使用原明确 `.verification/video-source-third` 与 V `.verification/video-timed-first`，七项实际原件解析与 952 个原生时间区间逐段核查 EXIT 0，1.221s，`.verification/video-default-consumed-first.log`；没有新 ASR 或模型。

PDF MIME 校验复用标准 `net/http.DetectContentType`，PDF 正文仍由现 summarize/markitdown 解析；没有实现 PDF parser。FetchPDF 保持原 ValidationFailed 消息。真实已完整读取的 ACL HTML 47355 bytes 经该生产本机规则拒绝，`go test ./internal/adapters/importers -run TestPaperPDFRequiresPDFMagic -v -count=1` EXIT 0，1.144s，`.verification/paper-pdf-real-format-second.log`；显式环境 `ASTROCYTE_TEST_PUBLIC_PAPER_DIR` 选择获准原件目录，缺失直接失败。原替代 HTTP Transport 已移除，V 可由其原 owner 清理 shared helper。

实际新 ACL HTTP 被 DNS `198.18.0.98` 的原 SSRF 守卫拒绝，`.verification/paper-pdf-real-html-first.log` EXIT 1，未到达 PDF 格式路径。可信代理 DNS 仍未授权，没有修改来源 IP 守卫、OS、DNS 或 TLS；本机格式通过不升级该 HTTP 结果。必要生产 NewArxiv collector 第一轮在官方 PDF GET 出现 EOF，`.verification/public-pdf-collector-first.log`，没有生成或伪造原件。第二次使用明确新目录完成生产官方来源下载与实际 CLI 提取，结果登记如下。

当前全仓 build/vet EXIT 0，`.verification/helper-video-combined-{build,vet}-first.log`，MIME 阶段全仓 vet EXIT 0，`.verification/paper-mime-vet-first.log`。原子 port0 监听及 H 普通 Workbench guard 测试的精确薄装配依赖已交当前协调者；旧 H/R dispatch 已结束，未向其追加执行或跨域编辑。实际 controller/模型窗口仍按根决定暂停新增，未把原先定向通过记为完整 Go、Relay 或最终产品验收。

## 实际 PDF 原件与 H 明确源测试

SOURCE `4a6076c402e0b5d31982154a31a905bee6127477` 已普通 push；前一 REPORT 为 `b01519924a10128314fd04279738332c14597a0f`。生产 `NewArxiv.Read` 完整取得固定官方 `https://arxiv.org/pdf/2501.12948v1`，1312189 PDF bytes；实际 `ExtractPDF` 使用原 summarize/uvx/markitdown 隔离进程，取得 59636 text bytes 与 64089 原 CLI JSON bytes，0 模型调用。证据 `.verification/public-pdf-collector-second.log`，原件仅在 `.verification/public-pdf-current-second/`；首 EOF 保留。

`TestPaperPDFSourceNeverPresentsURLAsFullText` 使用该实际 PDF、官方源 metadata、原 CLI JSON 及正文，全字段来源和原附件 bytes 保持断言。格式、来源 key、无提取配置及错误输入四项检查 EXIT 0，1.132s，`.verification/paper-pdf-actual-source-first.log`；显式环境 `ASTROCYTE_TEST_PDF_PUBLICATION` 选择上述获准目录，缺失直接失败。该检查是生产来源构造函数的真实输入验证，没有宣称新 `paper_pdf` HTTP 导入已通过。

PDF CLI 原 JSON 包含本机输入路径，构造函数使用已有 private-native MediaType 类别保留在本机对象。跨 revision ObjectRef 权限规则同时识别已有 `summarize-pdf-original.json`，防止既存名字别名开放原生元数据。新的 PDF HTTP 导入受前述原网络守卫限制，实际正向 HTTP 与私有 PDF 原件拒绝尚未运行；不以视频对应检查替代该结论。

协调者正式新建 H 非原生任务 `ctx_d7d8c2828fba` 后，H 冻结 `cmd/server/workbench_config.go`，B 仅原样消费不可变 `24b6390bcad60a2a20b671dcb60725e77eb3a7a4` 单文件，逐行 case-sensitive 内容相等。普通测试保留 H `78903848199b915a21e95e42f52cbed64c6d606c` 全部原断言，唯一修订为本机专属 `.verification/workbench-guard` 根；真实 SQLite/migrations、显式自有 sources、普通 app-state 与实际 snapshot hardlink 拒绝两项检查 EXIT 0，1.832s，`.verification/workbench-config-actual-first.log`。实际服务构建及全仓 build/vet EXIT 0，`.verification/workbench-config-build-first.log`、`.verification/public-pdf-config-{build,vet}-first.log`。该 helper 原写域已经 ACK 归还 H，源中的通用注释与异常处理整改交原 owner，未编辑其 main 装配。

F 只通过 GET 检查原真实视频副本 `.verification/astrocyte-s1-FAkZe5` 的两版资料、952 spans 和安全附件信息，两个 viewport 2 PASS。实际 revision 1 为 spans 0、public 1、private 0；revision 2 为 spans 952、public 3、private 1；安全只读聚合 `.verification/existing-video-version-counts-first.log`，原首错误计数断言由 F 保留。专属 API 61280 已停止，实际 health `ECONNREFUSED` 确认，`.verification/existing-video-stop-first.log`；0 新 ASR/模型，没有人工 SQL 修改资料。

## 原子监听与真实 DOM 输入

SOURCE `072628536d325eedbb6b53ffe9ab8fc31c38beeb`、`bd3c3ee3628e61bc99f800232fc17550843af026` 均已普通 push。经 H owner 与两个模块明确临时授权，B 仅移植 `1aadeebaba2330bb8b786989f577415da5208fd5` 相对其父的 main/server changed blocks，以及整文件 `relay_http_listener.go`，新增文件逐行 case-sensitive 相等；未消费其他 H 装配或整段历史。三个临时 H 文件域已在 SOURCE 后明确 ACK 归还，B 后续停止修改。

Go 一次绑定并持有实际 port0 listener；helper 使用成熟 readline、JSON、URL 读取自产 `http_listening/api_url`，随后实际 GET health。API 没有 Node 先占后释放的探测；明确重启绑定原已观测端口，Vite 使用该准确 API port，`--strictPort` 在端口失败时直接终止。公告和 HTTP 可响应分别核验，启动失败保留原 cause，不重绑或换端口。

`go test ./cmd/server -run 'TestResolvePort|TestWorkbench' -v -count=1` EXIT 0，四项测试及全部原有效/无效参数覆盖通过，2.300s，`.verification/atomic-listener-main-first.log`。端口 0 按明确新语义移至成功参数，其余拒绝断言保留。实际服务 build、全仓 vet 和 gofmt/diff 检查通过，`.verification/atomic-listener-build-first.log`、`.verification/atomic-listener-vet-first.log`。

`node --test tests/s1/server-lifecycle.test.mjs` 当前完整三项 EXIT 0，3 PASS、0 SKIP，64.936s，`.verification/atomic-listener-helper-final.log`；覆盖实际两个布局/SQLite/重开、Vite 代理/同源身份/CSRF、第二真实 Go 进程重复绑定必失败，以及原端口进程重启后的真实 Works 写入与独立 SQL 完整性。首 Vite 检查在缺公告事件的文字断言失败，`.verification/atomic-listener-web-first.log`；生产 helper 对同一公告失效语义保留不同实际事件 cause，第二次完整检查 23.971s 通过，`.verification/atomic-listener-web-second.log`。原 F 49680 Windows bind 首 RED 仍保留，F 独立实际用例验收不由该 helper 结果替代。

论文来源工厂及隔离检查移除虚构的论文 HTML/title/abstract/text，使用完整实际 ACL 与 PLOS HTML，经现生产 Node DOM 解析，正文和附件原 bytes、source key、标题、摘要和状态保持断言。restricted 是明确拒绝规则参数，未声明观察真实付费来源；IP/URL 表格属于明确纯函数参数。`go test ./internal/adapters/importers -run 'TestPaperSource|TestPaperDOMInstalledPinAndIsolation|TestIsPublicIP|TestValidatePublicPaperURL' -v -count=1` EXIT 0，六项 PASS，2.188s，`.verification/paper-source-actual-first.log`。必要 Node/固定 CLI 缺失现在直接失败，没有新增 skip；既有独立 opt-in public captures 保留，未假设其已运行。

当前根已授权继续语义、CLI 与 L/S 的已确认工作，但实际 CLI/selected-text/Today 仍须先相关 R/H 语义通过和根 working/done 决定。没有仅因跨日启动 controller 或模型，Source alias 合法端口及全写域残余替代实现仍待后续原 owner 完成。

## 当前默认来源检查与失败传播

SOURCE merge `0d9640ecb8142f643c4565aaac28b1c8b6b4b956` 普通消费 V 精确 `eead5308823804dc744723ad8e27550f91107c5b`，含其原 observer/cancellation SOURCE `e9e5f52511ca44a947d39fff61fac098c6fbad74` 和 REPORT `c2833a5de92644602868abc4200c1774186356c3` 祖先；已正常 push。ExtractPDF 在 dependency/CLI 之前复用成熟 MIME 规则，shared 替代 Transport helper 定义由原 V owner 移除；B 未修改 V 五文件。原实际 native progress 是否可观测保持 V 的 NOT_RUN，未用自写 parser 补时间或进度。

明确提供实际 video baseline/timed、public paper dir、PDF publication、批准匿名 proxy 及公开搜索开关后，`go test -mod=readonly ./internal/adapters/importers -json -count=1` 当前 EXIT 0，76 个测试/子测试 PASS、0 FAIL、6 个既有 opt-in SKIP，package 44.930s；`.verification/importers-default-current-sixth.jsonl`。SKIP 为 public Bilibili 独立检查、独立 arXiv live、本地 PDF live、两个跨站捕获以及 paper URL live，不能当作通过。此前默认公开网络的首 RED 保留，当前 package 通过不替代全仓或整轨验收。对应全仓 build/vet EXIT 0，`.verification/video-mime-combined-{build,vet}-first.log`。

F 再次明确授权的唯一 `cmd/server/workbench_config_test.go` 冲突已按不可变 `4a6076c402e0b5d31982154a31a905bee6127477` 完整源经文件编辑工具处理并仅 stage，逐行大小写相等，unmerged entries 空；其他 staged 依赖和 web WIP 全部保留。实际 F 组合服务 build EXIT 0，`.verification/f-workbench-config-build-first.log`；没有在 F commit/merge，单文件临时授权已归还。

SOURCE `391c6176145044b099c13a05a926aef2096db170` 已普通 push，进度失败保存与未知结果读取分别保留原语义错误和实际存储异常。真实 SQLite/migrations 经实际 context cancellation 同时触发生产写入和读取失败，原错误及 `context.Canceled` 都可定位，数据库零进度记录；`go test ./internal/workspace/app -run TestProgressFailureKeepsActualSQLiteCancellation -v -count=1` EXIT 0，0.974s，`.verification/progress-sqlite-cancel-first.log`。测试只调用相同生产函数，没有替代 repository；语义错误是明确函数参数，未声明模型调用或交付。进度代码注释改为中文，定向 vet 和全仓 build EXIT 0，`.verification/progress-error-vet-second.log`、`.verification/progress-error-build-first.log`。旧 progress memory/native 替代测试仍待完整整改，未用新增检查覆盖该限制。

协调者按最新用户决定要求 S2/S3 实现。B 已确认下列域在精确已 push `0d9640e` 无 WIP，并停止写入后移交固定 owner：S2 goals/proposals/approvals/session_bindings/proposal_context/proposal_ports 的指定 workspace/domain/app 文件、既有 app/service.go、指定 sqlite/httpapi workspace 文件、tests/s2 及 workspace-s2 报告；S3 internal/swarm、指定 sqlite/httpapi swarm 文件及 tests/s3。B 保留 projects/contracts/native_context/deps/migrations/scripts/config 的共享边界，跨域需求只提交 Handoff。

OpenCode 实际安装版本和 npm manifest 都为 1.18.35。固定版本来源、权限、server/CLI 文档与源码已核实；按根授权，版本检查和 serve 均使用完整自有 HOME/profile、XDG config/data/state/cache、temp、managed、数据库与 Windows appdata/programdata 子环境。原宿主 auth、配置与历史保持不读取、不复制，环境仅继承明确允许的系统路径与 locale，provider 凭据和启动注入项不继承。实际无模型生命周期已验证；当前模型和正向 Send 继续保持未核实。

注册项目缓存 SOURCE `bee569f0b30fe2496dc933b026989d955efdbaf3` 已普通 push。应用测试使用本工作树实际 Git、生产 `NewScopedRegisteredProjects`、真实 SQLite 和正常 metadata 接口；缓存读取断开发现来源并设置 `PRAGMA query_only=ON`，实际观察、人工备注、revision/updated_at 和持久化 JSON 全部保持。只清空可重建的 Board 活动投影以验证重新聚合，原 Git 来源不变；创建时间与活动时间规则另以明确聚合函数输入验证，不声明原生会话事实。两个原测试名及活动、元数据、零扫描/写入、原缓存不变覆盖保留，内存 repository 和替代扫描回调已移除。

`go test -mod=readonly ./internal/workspace/app -run '^TestRegisteredProjects(CachedGitActivity|ExistingSQLiteCache)$' -v -count=1` EXIT 0，1.237s，`.verification/registered-cache-actual-first.log`；实际 roots=1/groups=1、投影活动 0→1、恢复 1。定向 vet、全仓 build、gofmt 和 diff 检查 EXIT 0，`.verification/registered-cache-{vet,build}-first.log`。S2 当前唯一 owner 为 `ctx_f32b6d92119a`，使用新增 s2_service.go；既有 app/service.go 归 B 薄装配，在准确构造 SOURCE 前继续暂停修改。

注册发现生命周期 SOURCE `2a6a6d0ae20cc02f9160f6a1c4fdf5d1066d7ad7` 已普通 push。刷新错误在保存 stale/unknown 旧缓存后继续返回原错误，保存或 metadata 读取错误经 `errors.Join` 一并保留。Scoped 发现通过实际固定 Git `--git-common-dir` 命令取得同一 worktree 家族身份；已知 Git 身份读取失败立即失败，保留未知 Git 观察。SQLite discovery/metadata 测试移除替代 source、原生 controller 和引用 reader，观察包装只计数并委托生产来源。

目录创建、移除、恢复均发生在本树自有目录；实际刷新、权限拒绝、冷重启、unknown/stale、零项目/授权创建均保持覆盖。metadata 测试建立实际 Git 仓库和两个 worktree，通过正常 metadata 接口保存备注；实际 worktree move 后名称更新，移走自有 `.git` 后身份未知，人工备注及已有身份保持，新的 OS 测试进程继续读取同一 SQLite。并发取消由实际事务占用生产单连接触发，首刷新占有刷新锁并等待 SQLite；取消等待者和首刷新均及时返回 `context.Canceled`，没有额外连接等待或来源调用。该检查不声明取消了实际运行中的 Git 子进程。

`go test -mod=readonly ./internal/adapters/sqlite -run '^Test(Registered(ProjectCacheRestartStaleAndNoAuthority|SourceUnavailableBeforeFirstObservation|RefreshWaitCancellationDoesNotBlockShutdown)|ProjectMetadataPersistsThroughRefreshArchiveAndProcessRestart)$' -v -count=1` 第四轮 EXIT 0，4 项，3.806s，`.verification/registered-sqlite-lifecycle-fourth.log`。第一轮共享 helper 编译错误、第二轮被生产抑制的错误、第三轮错误类型期望错误均保留；最终精确断言生产 `ScopeDenied`、固定消息及 action。更新后应用缓存两项、相关三包 vet 和全仓 build EXIT 0，`.verification/registered-cache-current-second.log` 与 `.verification/registered-lifecycle-{vet,build}-first.log`。

按协调者新增共享 DDL 唯一 owner 分配，S3 请求的独立 SOURCE `da48a114d91181075bd32b62996beadd9c2c0a25` 已普通 push，仅增加 `migrations/020_swarm.sql` 五表及 Swarm 版本登记。Mission grant_id、租约 work_item_id/operation_id、消费者 consumer_id/event_id 唯一约束及所请求外键保持，复用既有幂等/outbox 定义；001–018 内容没有修改，019 留给 S2 准确表需求。上述四项实际 SQLite 检查在新迁移下再次 EXIT 0，3.403s，`.verification/swarm-migration-actual-first.log`，包含正常迁移和重新打开。Swarm 业务 CAS、授权同事务重查及两执行者闭环由 `ctx_d4bf65ca06bf` 原 owner 执行，B 尚未运行其验收。

当前协调者最终划分为既有 app/service.go 归 B 薄装配、新增 s2_service.go 归 S2；S2 已交准确构造 SOURCE，既有 service.go 无需修改。B 已普通 no-ff 消费 S2 `507be52f4e31759dc8c3765aeea964f1f2efe013`，同数据库 Attention adapter 衔接与实际验证登记如下。H 单 helper 修正已按下述精确 SOURCE 消费并归还临时写域。

S2 精确请求的新增迁移 SOURCE `da79452d1735f8fceca103aaf9d10bb53182f1f5` 已普通 push，仅增加 `migrations/019_workspace_approvals.sql`。七表保存目标/提案固定版本、ApprovalGrant、ContextPacket 和 SessionBinding，复合版本键、指定唯一约束与外键均按原 owner 请求实现；实际原会话表为 `local_agent_sessions`。新增 foundation_idempotency.result_json 使用 `TEXT NOT NULL DEFAULT 'null'`，既有 result_code/outbox 和 001–018 文件保持原内容；workspace schema 版本由 6 增至 7。上述四项实际 SQLite 迁移/重启/metadata OS 进程/权限/取消检查 EXIT 0，7.487s，`.verification/workspace-migration-actual-first.log`。S2 已发布源码已消费，HTTP 及完整上下文接续验收仍由原 owner 与 H 执行。

H 唯一 helper 精确消费 SOURCE `9f7891ecabd51cffa645a61cfa861eeef670b046` 已普通 push，`cmd/server/workbench_config.go` 与 H `25d843783d65ada98c0c51766a8d5b35b8f5f941` 全文件 case-sensitive 比对相等；只使用文件编辑工具，没有合入其他 H 历史。Stat/ReadDir、generation canonical 与 identity 错误立即拒绝，首次尚未创建目录保持原初始化语义。B 既有真实 SQLite 目录和 hardlink 两项 EXIT 0，1.874s，`.verification/workbench-helper-25d-second.log`，临时单文件写域已归还。H 自己的实际 junction 检查通过和 symlink OS 1314 首次失败分别保留，B 没有将该 symlink 语义记为通过。

共享 policy SOURCE `4725f0b1c6907ddefb3b1e1dac0f49840031db59` 已普通 push。`ValidateProjectActionScope` 只接收事务已加载的项目与要求，强制当前 settings revision、明确业务 action、tools、CLI 原名外发以及目录/AllowedSubdirs；`claim_work_item` 还需 AgentControl。`AuthorizeProjectActionCaller` 接收同事务已加载的 ProjectGrant，检查 Agent 归属、到期、撤销与 action，拒绝 Agent 的五项人类专属业务；人类安全停止在身份验证后通过原 owned 控制器处理，不依赖已撤销执行批准。S3 九项业务名称逐一加入既有 validActions，没有 native 别名映射。两项纯函数参数检查及 11 个子项 EXIT 0，0.929s，`.verification/project-policy-ports-first.log`；这些输入不声明真实批准或来源，S2/Swarm 实际接口与并发验收仍由原 owner 执行。

OpenCode SOURCE `a3fe68f5420f439c51ffa5e67e034543a01358cf` 已普通 push，移除全部替代 HTTP 服务。实际安装的 1.18.35 原生进程返回原 session ID，实际 HTTP 核对自有路径、空 MCP、空 external plugin、末尾通配 deny 权限；创建、空历史读取、确认进程停止、重新启动同 ID 恢复、错误 Basic auth、未知 session、四种不安全权限均通过真实原生接口。模型调用为 0；正文/回复投影继续使用明确 JSON 函数参数，不冒称实际模型回复。paid Send、真实非空上下文及跨会话交接保持 NOT_RUN，原 opt-in paid 测试继续保留，未新增 skip。

启动必须先取得固定版本在 Server.listen 成功后输出的实际 listener URL，经标准 `net/url` 核对精确 endpoint 后，只发送一次健康请求；原 20 秒启动总预算保持。官方来源为 [固定版本 serve](https://github.com/anomalyco/opencode/blob/v1.18.35/packages/opencode/src/cli/cmd/serve.ts)、[固定版本 listener](https://github.com/anomalyco/opencode/blob/v1.18.35/packages/opencode/src/server/server.ts) 与 [server 文档](https://opencode.ai/docs/server/)。HTTP/JSON/响应关闭/清理错误明确返回，停止仍完成实际进程树处理；输出超限拒绝。实际 raw CLI 输出、原生数据库与原生响应只留在本树 `.verification/opencode-owned-*`，不进入产品数据库或 Git。

`go test -mod=readonly ./internal/adapters/agents -run '^TestOpencode(IsolationEnv|SessionLifecycleOverHTTP|SessionDenyPermissionFailClosed|RejectsUnauthenticatedAndUnknownReply|ReplyProjectionInputs|WaitHealth|AdapterSnapshotRegistered)$' -v -count=1` 最终 EXIT 0，7 项，12.280s，`.verification/opencode-owned-announced-first.log`。实际四个保留目录为 `opencode-owned-1049957266`、`opencode-owned-425773741`、`opencode-owned-2171616221`、`opencode-owned-4283419979`；首个目录记录同 ID 创建/停止/恢复。前面的 Windows 错误类型与首次健康请求期限失败仍保留于 `opencode-owned-no-model-{final,corrected,budget}.log`，未重标为 PASS。最终相关包 vet、全仓 build、gofmt/diff EXIT 0，`.verification/opencode-announced-{vet,build}-first.log`。协作式 CLI 的 OS 读取隔离、原生来源完整度、有效工具集合及实际 Codex event_id 链验收均未由这些结果提升。

S2 精确普通合并为 `20bdbc55850e0a5e9d5297aa9c27d399f8a2e696`，Attention 共享事务 SOURCE `ce281d003c093ec26e19dfec78e912f20a014787` 已普通 push。`AttentionRepository.WithTx` 明确区分缺席、同数据库和另一数据库；同数据库复用现连接，由原 Workspace 事务提交或回滚，独立事务则保留原处理并传播回滚/关闭异常。实际 `Service.CreateProjectSpace` 同事务提交，重复同一空间 INSERT 的真实主键约束使空间、attention_outbox、attention_receipts 全部回滚；另一数据库拒绝回调，缺席上下文可正常开启独立事务，连接等待计数保持原值。两个测试含四条路径 EXIT 0，1.839s，`.verification/attention-workspace-transaction-second.log`，第一轮查错 foundation 表的断言失败保留。S2 原实际批准/撤销/幂等/SQLite 重启用例在 B 组合源码 EXIT 0，1.295s，`.verification/workspace-s2-507-actual-first.log`；相关 vet/full build EXIT 0。源码合并第一次远端拒绝，核 remote 仍在前一 REPORT 后第二次正常 push 成功。

按 M/H 明确授权及 H 零 WIP 停写 ACK，B 在 H 工作树仅使用文件编辑工具复制 019 原件并只 stage 该文件。H index blob 与原 da794 文件精确同为 `352ecc0e5e2f12aa52bf3871e7304f15292ba667`，001–018 无 diff；H 的其他 WIP、源码提交、实际服务和数据库没有由 B 操作。临时范围已归还 H，未复制 020 或整段 B 历史。

S4 新增投影迁移 SOURCE `98bb9236c68f45ca4ab4114d46eafba5c90cdedf` 与 S3 人工退回 policy SOURCE `c110201774dd61b8d7e5280526364786c6af4cc2` 已普通 push。`021_repository_context.sql` 只增加所请求安全元数据五列、三字段主键及 workspace 版本 8；S4 实际 generation/CAS/索引语义归其原 owner。`reject_artifact` 逐名加入 validActions 并仅允许人类，Agent 即使持有该动作字符串仍拒绝。新增迁移下实际 SQLite 四路径 EXIT 0，2.069s，policy 两项含 11 子项 EXIT 0，0.906s，`.verification/repository-context-migration-first.log`；定向 vet/full build/gofmt/diff EXIT 0。CodeGraph 1.0.1 官方元数据和实际 SDK 依赖已按下述结果登记，实际 S4 索引与 MCP 能力继续由原 owner 验证。

当前已向新 UI 与 M 提供实际 GetItem 核实的五个 Reader 环境变量指针：Q main `public-cognition-6f63200932f14a10ad4d9fb91c78d749/local`、Q `cognition-before-sleep-local`，以及 B `reader-actual-public/{empty,closed,native}/local`。这些出版物均为本次公开来源工程派生；本机已跨上海 10-15，旧出版物和闭日 Sleep 执行时间仍为 10-14，默认实际时钟检查的今日来源前提需要当前公开出版物。未改变时钟、断言、正式资料或实验日程；整套 Reader 新一轮验收保持未完成。

依赖 SOURCE `6b78f641ecb2a3ae53b11df9aeeb837e6c72c646` 已普通 push，`go.mod`、`go.sum` 与现有 `dependencies.lock.json` 增加官方 `github.com/modelcontextprotocol/go-sdk v1.6.1` 及其实际运行依赖，现有 x/net/x/sys 版本保持。官方固定版本 go.mod 要求 Go 1.25.0，当前 1.27.2 满足；[固定版本许可](https://github.com/modelcontextprotocol/go-sdk/blob/v1.6.1/LICENSE) 按贡献适用 Apache-2.0/MIT，非规范文档为 CC-BY-4.0，未统一标为单一 MIT。

CodeGraph 顶层和实际 Windows bundle 均为 `1.0.1`，本机 bundled Node 实际 `--version` 为 `v24.16.0`，SDK 要求 `>=20.0.0 <25.0.0`。`lib/package.json` 的十项实际 runtime 版本与许可已经逐一读取并登记；tree-sitter-wasms `0.1.13` 使用 Unlicense，其余九项元数据为 MIT。Node 使用上游 MIT 及[第三方许可](https://github.com/nodejs/node/blob/v24.16.0/LICENSE)。顶层/Windows bundle 许可为 package metadata 的 MIT，顶层没有独立 LICENSE 文件；实际依赖 LICENSE 保留在安装目录。子环境要求 `CODEGRAPH_TELEMETRY=0`、`CODEGRAPH_NO_DOWNLOAD=1`，没有修改 global、root package.json 或 pnpm lock，也没有在 B 启动索引/daemon。实际清单 `.verification/codegraph-actual-dependencies-first.json`；AOCI 完整构建身份等待 S4 原 owner 的准确指针，未自行补造。

`go build -mod=readonly github.com/modelcontextprotocol/go-sdk/mcp`、`go list -mod=readonly -m -json all`、`node scripts/check-dependencies.mjs`、全仓 `go build -mod=readonly ./...` 和 `go vet -mod=readonly ./...` 均 EXIT 0。证据为 `.verification/mcp-go-sdk-build-first.log`、`mcp-go-sdk-module-graph-third.json`、`mcp-sdk-dependency-inventory-first.log` 与 `mcp-sdk-combined-{build,vet}-first.log`。模块全图前两次缺既有 ndjson 传递元数据 checksum 的失败保留，校验值仅从实际官方 Go download 结果通过文件编辑工具追加。SDK 包编译与库存通过不等于 AOCI/CodeGraph 实际查询验收。

按 H 当前路径 freeze 和 M 必要依赖授权，B 又在 H 树仅精确复制并 stage `072628536d325eedbb6b53ffe9ab8fc31c38beeb:tests/s1/server.mjs` 与 `da48a114d91181075bd32b62996beadd9c2c0a25:migrations/020_swarm.sql`，对应 index blob 为 `2bf5614179269c87e3570e79dbbb19a0e086f726`、`298034448fe71fa2d74c8307c20393affaaf4792`，均与不可变原件相等。H 001–018 无 diff，Node --check EXIT 0；H 其他 WIP 没有修改，源码提交、实际 S2 HTTP 与首次失败证据由 H 完成，临时范围已归还。

收藏第三页实际仅返回 19 条元数据。协调者转根正式决定：初次 100 指实际返回的资料元数据，允许同一已获准公开来源跨越缺失请求槽位，取得第 100 条即停止；Observed 保持实际条数，cursor 按请求窗口推进，原 timeout、显式 limit 和无进展失败继续执行。

收藏 cursor SOURCE `52bd568811cbcdf59540ba18eee5a37c9d32002b` 已普通 push。`go test ./internal/adapters/importers -run TestPublicListingActualWindowCursorWithUnavailableSlot -v -count=1` EXIT 0，实际生产集合列举选出公开 folder `2356677875`，真实 Transport 始终访问上游 HTTP；三次请求位置 0/20/40，各请求 20，实际返回共 59，下一 cursor 为 4。观察包装只记录实际 query，不替代 HTTP 响应；没有把 19 改成 20。证据 `.verification/public-listing-cursor-first.log`，定向 vet EXIT 0，`.verification/public-listing-cursor-vet-first.log`。

容量测试 SOURCE `c28f6b15ae0032df2a21e0625f0985546759610d` 已普通 push，独立前段 REPORT 为 `580b4a37d0bd3decce491a772fe5ee86fd64d445`。同一 folder 属于原已获准公开 UID `3494358764489275`，来源许可为 `docs/QUESTIONS.md` 2026-10-10 用户决定；T 明确原件目录 `public-http-1166636830` 中 response 3/5/7/9 的成熟 JSON 聚合分别为相同 id、media_count 459、20/20/19/0，确认与 T 原容量来源相同，没有新增收藏范围。

原 `TestPublicListingBoundsActualFavoriteRequestsIncludingSkippedRows` 完整覆盖改为生产公开 HTTP，计数和剩余预算保持断言；六次实际窗口位置 0/20/40/60/80/100，尺寸 20/20/20/20/20/1，观察恰 100、实际资料 100、请求槽位 101、next cursor `6:1`。三项定向检查 EXIT 0，2.273s，`.verification/public-listing-capacity-first.log`；最后十行 uploads JSON 属于明确格式函数参数和期望值，未声明为真实上传来源或容量验收。两个 SOURCE push 首次 TLS 失败保留，第三次普通 push 成功。T 的独立应用/SQLite 100 条验收继续由其原 owner 执行。

## 完成内容

- 保留接管时全部 B 未提交贡献。`paper-dom.mjs` 和插件提取模块使用 `JSON.parse`、`URL` 直接失败，错误 JSON-LD 与 PDF URL 各有断言。
- JavaScript 结构由 `goja` AST 解析；Bash 由 `mvdan.cc/sh/v3` AST 解析；Windows PowerShell 使用实际 `Parser.ParseInput`。结构分类保留未知状态和解析失败信息，既有循环展开规则继续执行。
- 新增原生 JSON 字段格式验证，错误 timestamp、content、usage 和工具参数在映射之前返回错误；`cmd/validation/semantic.go` 传播原生工具读取错误。
- Crossref 摘要标签和字符引用使用 `golang.org/x/net/html` tokenizer；年份、出版时间、arXiv 标识符解析错误立即返回。Go 模块元数据用于依赖库存验证。
- summarize 字幕 CLI 失败、JSON 失败及来源不匹配立即失败；只有实际明确无字幕时执行同一公开视频的上游本地转写。时间区间必须来自上游输出，缺失保持缺失。
- GitHub 元数据测试使用实际公开查询和独立格式输入；MCP 注册及关注状态测试使用生产 SQLite、引擎和 MCP 服务；论文应用查询测试调用实际 Crossref。HTTP 500 完整 ErrorV1 断言通过真实请求取消触发生产 foundation 服务错误。
- 实际文件回放测试明确创建 Astrocyte scope，并另外以纯函数输入逐项断言未知 scope 的主题、token、outcome 与 depth 保持未知；冻结分类规则未改。
- 部分导入及 foundation 注释改为中文；复制快照失败显示实际错误并抛出。验证进程和资料保留在本工作树 `.verification/`。
- 原生 record 改为 typed JSON 解码，timestamp、content 和 usage 处理直接传播错误；整数 token 累加溢出明确失败，缺失分量保持未知。
- S1 论文来源移除本地伪造 Atom/PDF 服务，使用真实 arXiv 接口，四种公开别名保持同一固定版本，两个实际版本保留不同原件，未发布版本必须拒绝。
- HTTP attention 测试使用实际应用、SQLite 与官方论文 PDF；鉴权拒绝时数据库没有作业，正确身份、调用键复用、下载原件 bytes、反馈 201 与无效 revision 400 全部保留断言。HTTP 项目权限测试使用实际注册、签发与撤销 token，观察包装只记录参数并委托生产实现。
- `startS1Server({ relay: true, browser: true })` 显式采用 `<temporary>/local-appdata/astrocyte-workbench/v0.2-access/app-state`，返回 `localAppData` 与 `relayRoot`；所有目录位于本树 `.verification/`。注册表默认不存在，生产 Relay 按空列表处理。重用原目录时继续传 `relay: true`。该 helper 只准备目录和启动实际源码，本树尚无 Relay 路由，Relay 业务验收由 F/H 在实际组合源码执行。
- Works 原有十项测试全部改为实际 SQLite、生产 Scanner/Delivery/WorksEventSink 与磁盘资料；保留平台计数、人工字段、版本冲突、投递来源、事件归属、零额外认知 Signal、实际根失效的覆盖，并增加格式错误保留 stale 缓存的测试。
- 作品标记 JSON 格式、文件类型和遍历异常直接失败。Workspace 保存 stale 缓存后返回原失败，HTTP 不会把失败刷新报告为成功；深度/数量限制仍明确报告未核实。
- 进度人工输入验证不再创建虚构 TASK/STATUS 或 OS 临时根。实际项目读取授权仅选 `README.md`，Codex inference 使用已有配置且不指定模型；独立 `stat` 核对实际 `mtime;size` 引用，SQLite 与新 HTTP 进程复核同一结果，零 Mission。
- 根据 H 不可变 `99a250d0b9069ffb32205420f63f82b372de5607` 的实际 NDJSON 使用，依赖源追加 `github.com/olivere/ndjson v1.0.1` 与精确 checksum。保留 B parser 依赖、`x/sys v0.49.0` 与直接 `uuid`，模块和库存验证通过。后续协调者明确临时授权 B 编辑并 stage F 唯一冲突 `go.mod`，实际内容与 B `2a6f03605ebb87f36f7bd56c577aaf4853cb6671` 全等；其他 F 文件及 merge commit 均归 F，单文件授权已经结束。
- HumanAPI helper 使用实际 Idempotency-Key 作为未显式提供的 request_id，显式错误身份仍由生产 400 拒绝。HTTP 响应先完成 JSON 编码再写状态；编码失败返回完整 ErrorV1 500，响应写入失败终止当前 HTTP 处理。非 JSON HTTP 原件及解析错误只保存在 `.verification/`，异常文本仅公开证据指针。
- 按明确授权，复用不可变 Workbench SOURCE `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b` 中 `internal/workbench/app/contracts.go`、`internal/adapters/sqlite/workbench.go` 和对应测试。保持既有六字段事件投影、日期范围、分页、过滤和 null 语义；未导入其他 owner 文件。
- Reader 使用 `NewWorkbenchReader(root string, protectedSources ...string)`。受保护源由可信启动调用者显式指定，默认仅读取发布根；不含或探测个人正式库常量。显式源同文件及硬链接通过 `os.SameFile` 拒绝，发布文件及路径拒绝 symlink/junction，SQLite 采用 `mode=ro`、`immutable=1`、`query_only(1)`，解析和关闭异常传播。上海时区使用嵌入 tzdata。
- 文本处理目录限定到现有 Registry 的绝对 `stateRoot/selected-text`；相对根与真实文件类型异常直接失败。HTTP Works 通过生产注册服务对明确选择的实际目录执行一次发现，缓存、权限、刷新、重复请求及 SQLite 重启由生产服务处理；观察包装只统计真实刷新调用。
- Reader 查询使用 Q 不可变 SOURCE `1d07c5b2ff78db5ffe7e90b4d903c618b5a8ac50` / REPORT `6ad1dfa1387d5e6a0b81ac1b2ad2887359780658` 的实际公开 Git 生产回放出版物。1,339 个实际 event_id 逐一核对 Git commit origin、原 ts 与公开仓库，全部六字段、日期、分页、聚合、节点数值、实际历史、current 切换与只读检查保持完整断言。
- Windows 每个路径祖先均使用现有 `x/sys/windows.GetFileAttributes` 核查 reparse；实际根和 generation 的 junction，以及同文件/hardlink 全部拒绝。历史去重仅在内存中重复同一实际点作为明确规则输入；JSON null/整数投影作为明确 SQL 参数输入，两者均不写来源事件、节点或 Sleep。
- 默认 Reader 与 Sleep 均使用实际 `time.Now`。在新的自有公开副本中，实际生产 `cmd/sleep` 手动处理明确闭日 2026-10-13，开始 `1791986176576`、完成 `1791986176583` 毫秒，均为今天实际执行时间。成熟 Q CLI 生成 `snapshot-2249990550`，默认 Today 正确查询该前日完成记录；此操作只属于本次工程验证，正式实验及 Q 最终发布代保持原件。
- H 本轮自有实际公开数据库通过其 `dist/l1-live/current-20261014/service-private.json` 中正选 `db_path` 使用 Q 只读在线备份，B 出版代 `snapshot-3888376845`。独立 SQL 实际 25 个事件全部逐 id 和字段核对，Reader tokens 投影 25 个 null。该 Workbench 投影与先前 Relay HTTP 18/17 的投影及取得时刻分别登记；原生会话完整度及工具集合的未知状态不改变。
- `loadPublicPaperInputs(absoluteDirectory)` 使用实际已经完整取得的 arXiv v1/v2 与 PLOS HTML，经现有 linkedom/Readability/paperDOM 解析。固定来源身份与版本全等、两版真实正文不同、完整正文状态通过；导入字段只包含实际正文、标题及 whole-document 引用。该入口尚待默认 S1 与 F 消费。

## 实际验证

| 命令或实际路径 | 结果与证据 |
|---|---|
| `go build -mod=readonly ./...` | EXIT 0，`.verification/build-current-all.log` |
| `go vet -mod=readonly ./...` | EXIT 0，`.verification/vet-current.log` |
| `node scripts/check-dependencies.mjs` | EXIT 0，固定依赖、安装与锁清单一致，实际 summarize CLI 0.25.1；`.verification/dependencies-current.log` |
| `node scripts/check-architecture.mjs` | EXIT 0，30 包满足生产导入边界；`.verification/architecture-current.log` |
| `node --test internal/adapters/importers/paper-dom.test.mjs scripts/stop-process.test.mjs` | EXIT 0，11 PASS、0 SKIP；`.verification/node-current-first.log` |
| `node tests/s1/paper-dom-public.mjs` | EXIT 0，读取前轮完整下载的真实 PLOS 与 ACL 页面；PLOS 260555 HTML bytes / 63806 正文字符 / 15 作者，ACL 47355 bytes / 摘要元数据 / 8 作者；Node 与插件结果全等；`.verification/dom-public-current.log` |
| `go test -mod=readonly ./... -count=1`，`ASTROCYTE_PAPER_SEARCH_LIVE=1` | 两次全包运行均 EXIT 1，首失败保留。第一次 M1 路径包含 Astrocyte 导致原未知断言失败；第二次 M1 与其余包通过，importers 的 arXiv/Crossref/Europe PMC 网络失败；`.verification/go-current-all-{first,second}.log` |
| `go test -mod=readonly ./internal/adapters/importers -v -count=1`，公开查询开启 | EXIT 1，Crossref/Europe PMC/arXiv 实际查询均通过，arXiv summarize 提取出现 `fetch failed`；`.verification/importers-current-third.log` |
| 前轮公开视频实际 Service / SQLite / objects / 重启 | 公开 `BV1PReT6EEqR` 正文 32908 bytes，上游 whisper.cpp，3 附件、同原作业复用、重启内容全等、零自动沉淀/人工关注/Mission；`.verification/video-live-fourth.log` 与 `.verification/video-source-third/`。时间区间 0，不能记为时间区间验收通过 |
| `gofmt -l` 变更 Go 文件、`git diff --check` | 无输出，EXIT 0 |
| 相对基线检查 `internal/attention/domain`、`migrations`、`docs/EXPERIMENT.md` | 零 diff，正式实验目录、数据与日程未读取或修改 |
| `go test -mod=readonly ./tests/s1 -v -count=1` | EXIT 0，3 PASS、原有受控库路径 opt-in 1 SKIP；两个 S1 测试实际官方 Atom/PDF。自动未知作业测试的模型错误仍是替代实现，不能算真实模型未知交付；`.verification/s1-real-source-first.log` |
| `go test -mod=readonly ./internal/adapters/httpapi -v -count=1` | EXIT 0，公开论文实际导入/PDF/反馈通过；`.verification/http-real-attention-first.log` |
| `go test -mod=readonly ./internal/adapters/httpapi -run 'TestScopedAgent\|TestLocalProjectWrites' -v -count=1` | EXIT 0，两项真实 token/权限测试 PASS；首次包装接口缺少生产鉴权方法的 RED 保留在 `.verification/http-real-project-first.log`，最终 `.verification/http-real-project-second.log` |
| `go test -mod=readonly ./internal/adapters/events ./cmd/validation ./tests/attention-engine -count=1` | EXIT 0，typed native 解码与原生格式/语义/回放全部适用检查通过；`.verification/native-typed-first.log` |
| `node --test tests/s1/server-lifecycle.test.mjs` | EXIT 0，普通与 Relay 目录布局各 1 PASS、0 SKIP，实际 HTTP、session、SQLite、退出/再启动与新 helper 重开通过；`.verification/server-relay-layout-first.log` |
| 第二阶段 `go vet -mod=readonly ./...` / `go build -mod=readonly ./...` / `node --check tests/s1/server.mjs` | 全部 EXIT 0，`.verification/{vet,build}-second-source.log`；helper syntax EXIT 0 |
| `go test -mod=readonly ./internal/workspace/app ./internal/adapters/worksscan ./internal/adapters/sqlite -count=1` | EXIT 0，真实 Works 测试、错误传播与既有存储回归通过；`.verification/works-real-first.log` |
| `node tests/s1/progress-infer-live.mjs` | EXIT 0，一次实际生产 HTTP/Codex/SQLite/重启；实际 native_id `01a13a73-1de2-7761-aaa3-5cbd8fb433cc`，配置模型 `gpt-6.1-sol`，README 26056 bytes，独立引用 `mtime:1791968415369282600;size:26056`；`.verification/progress-infer-readme-first.log` 与 `.verification/astrocyte-s1-opvXPA/` |
| `node tests/s1/progress-live.mjs` | EXIT 0，实际人工输入、未知初态、重启与无许可 infer 403；`.verification/progress-manual-current.log` 与 `.verification/astrocyte-s1-lxfFRc/` |
| `go mod verify` / `node scripts/check-dependencies.mjs` | EXIT 0，实际 Relay parser 依赖 union 验证；`.verification/relay-dependency-union-first.log` |
| 第三阶段 `go vet -mod=readonly ./...` / `go build -mod=readonly ./...` | 全部 EXIT 0，`.verification/{vet,build}-third-source.log` |
| `go test -mod=readonly ./... -json -count=1`，公开查询开启 | EXIT 1，七项实际公开请求失败，网络 RED 保留；`.verification/go-current-all-third.jsonl`。不以之前定向 PASS 覆盖当前全包结论 |
| `go test -mod=readonly ./internal/adapters/httpapi -run 'TestHealthEndpoint\|TestHTTP500_ErrorV1_AllFieldsPresent\|TestJSONEncodingFailureDoesNotSendSuccess' -v -count=1` | EXIT 0，实际服务失败与明确不可编码值保留完整 ErrorV1；`.verification/http-json-error-first.log` |
| `node --test tests/s1/server-lifecycle.test.mjs`，命令身份修正后 | EXIT 0，两个实际布局各验证成功刷新及显式错误 request_id 400；2 PASS、0 SKIP，`.verification/helper-command-identity-first.log` |
| 早期 Reader 物理机制检查 | `.verification/workbench-reader-second.log` 保留，仅属于早期物理 SQLite/机制候选；当前真实来源查询结论使用下方完整生产来源记录 |
| Reader SOURCE 的 `go build -mod=readonly ./...` / `go vet -mod=readonly ./...` | 全部 EXIT 0，`.verification/workbench-{build,vet}-first.log`；架构检查 31 包通过 |
| 实际 Reader 测试二进制扫描 | EXIT 0，旧个人路径常量不在实际编译二进制，21141504 bytes；`.verification/workbench-binary-first.log`。验证从不可变原始代码 AST 提取字符串，只比较二进制 bytes，未打开该原始路径 |
| `go test -mod=readonly ./internal/adapters/agents -run TestSelectedText -v -count=1` | EXIT 0，两项完整 UTF-8 与实际受控目录检查；TEMP/TMP 为本树 `.verification/go-temp`，`.verification/selected-text-owned-second.log` |
| `go test -mod=readonly ./internal/adapters/httpapi -run 'TestWorksHTTPActualStoresRegisteredBoundsAndRestart\|TestEngineIndependentServicesActivateSessionGuard' -v -count=1` | EXIT 0，两项真实发现/SQLite/HTTP/权限/重启检查；`.verification/http-works-real-second.log` |
| `go test -mod=readonly ./internal/adapters/sqlite -run TestWorkbench -v -count=1`，明确选择下方 Q/B/H 的实际出版物 | EXIT 0，10 PASS、0 SKIP；`.verification/workbench-real-source-fifth.log`。Git cohort 1339 origins 全匹配，实际整数投影 1339/null 0；另一次实际 native 副本 25/null 25。整数投影不冒称模型 usage |
| 当前 Reader SOURCE 的 build / vet / architecture | `go build -mod=readonly ./...`、`go vet -mod=readonly ./...` 全 EXIT 0；`.verification/workbench-real-{build,vet}-first.log`，架构 31 包通过 |
| 当前 Reader 测试二进制扫描 | EXIT 0，旧个人路径常量不存在，21196288 bytes；`.verification/workbench-real-binary-first.log`，只读 Git 对象及本机二进制 |
| `go run ./cmd/sleep --data-dir .verification/reader-actual-public/closed/app-state --day 2026-10-13 --timezone Asia/Shanghai --decay 0.1 --initial-confidence 0.5 --fact-projection none` | EXIT 0，真实闭日处理完成且非 replay；`.verification/reader-closed-day-sleep-first.log`，没有改 Clock 或源时间 |
| Q 实际 `workbench-snapshot-final.exe --source <本轮明确自有源> --root <本树出版根> --timeout 120s` | 闭日、实际空库、H 实际 native 三次均 EXIT 0；`.verification/reader-{closed,empty,native}-publication-first.log` |
| `node tests/s1/public-paper-inputs.mjs .verification` | EXIT 0，arXiv v1/v2 正文分别 57465/229400 bytes、PLOS 64008 bytes，实际版本与 source_key 全等；`.verification/public-paper-inputs-first.log` |
| `node scripts/verify-import-inputs.mjs .verification/public-inputs-youtube-first youtube` | EXIT 1，实际完整网页取得后 summarize 子进程 `ETIMEDOUT`；`.verification/public-youtube-input-first.log`，未伪造字幕或补时间区间 |
| `go test -mod=readonly ./... -json -count=1`，全部当次实际来源与公开查询环境明确提供 | EXIT 1，arXiv 原件 HTTP、summarize、精确查询、公开搜索四个相关 case 失败；`.verification/go-current-all-fourth.jsonl`，网络首 RED 保留 |

公开 DOM 来源为 `https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514` 与 `https://aclanthology.org/2024.acl-long.1/`。本报告只给聚合与本机证据指针；原件保留在 `.verification/`。

Reader 验证环境须明确提供下列实际本机原件，不提供时直接失败，不加 skip：

- `TEMP`、`TMP`：本工作树 `.verification/go-temp`。
- `ASTROCYTE_TEST_COGNITION_LOCALAPPDATA`：Q 工作树 `.verification/public-cognition-6f63200932f14a10ad4d9fb91c78d749/local`。
- `ASTROCYTE_TEST_COGNITION_PRE_SLEEP_LOCALAPPDATA`：Q 工作树 `.verification/cognition-before-sleep-local`。
- `ASTROCYTE_TEST_COGNITION_EMPTY_LOCALAPPDATA`：本工作树 `.verification/reader-actual-public/empty/local`。
- `ASTROCYTE_TEST_COGNITION_CLOSED_DAY_LOCALAPPDATA`：本工作树 `.verification/reader-actual-public/closed/local`。
- `ASTROCYTE_TEST_COGNITION_NATIVE_LOCALAPPDATA`：本工作树 `.verification/reader-actual-public/native/local`。

Q 工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/release-cognition-20261014`。全部副本和出版物只在本机，不加入 Git 或发布包；H 原源、Q 原 source/current、正式实验均未写入。

## 原生操作身份检查与准入开发状态

SOURCE `dfdd75ed4b245fe462df2ff916728afdae7361cc` 已普通 push，保留现有生产 `NativeRequest.Session.PendingOperation` 调用链，新增实际 SQLite、实际公开项目注册、当前项目设置与真实 Registry 的拒绝检查。无效操作 UUID 在启动前返回 `validation_failed`，实际会话表为空，Registry 的启动观察保持 `unknown`；没有新模型调用。全部中间原件留在本树 `.verification/native-operation-actual-3426206494/`。

`go test -mod=readonly ./internal/workspace/app -run TestNativeOperationInvalidIdentityDoesNotLaunch -count=1 -v` EXIT 0，1 PASS、0 SKIP，2.712s；证据 `.verification/native-operation-identity-owned.log`。早期实际检查 `.verification/native-operation-identity-first.log` 与 `.verification/native-operation-identity-current.log` 保留。

准确 S2 SOURCE `bcc80d08e4bef7f93027a5fd0eb84f2d8f1749d9` 经普通 merge `cc3ed6887ed6834d691988c040e3b9600b52277c` 消费。准入阶段 SOURCE `e96f8d78f686f6db0f0a9edb5e15cc57ddb8aa1a` 已普通 push，包含独立领域文件、应用服务、SQLite repository、新业务迁移 `022_attention_admissions.sql`、实际 S2 bridge、HTTP 路由注册函数及实际来源测试；旧领域文件和迁移 001–018没有修改。`go build -mod=readonly ./...`、`go vet -mod=readonly ./...` 与 `node scripts/check-architecture.mjs` 均 EXIT 0，架构检查覆盖 33 个包；证据 `.verification/admission-current-build.log`、`.verification/admission-current-vet.log`、`.verification/admission-architecture-first.log`。编译及静态检查未作为准入真实验收结果。

实际准入验证先经历两次公开 arXiv 来源 `provider_unavailable`，原失败保留于 `.verification/admission-actual-first.log`、`.verification/admission-actual-second.log`。显式使用已经核实的现有匿名系统代理后，实际固定论文导入、人工待审记录和候选创建成功；第三次验证在 S2 消费者因人工准入缺少委托策略版本返回 `evidence_missing`，证据 `.verification/admission-actual-proxy-third.log`。该约束已交原 S2 owner 修正，人工记录不填充委托模型或策略版本，真实事务正向仍为未通过。独立准入领域规则检查 EXIT 0，证据 `.verification/admission-domain-current.log`，不代表真实委托判断或业务组合验收。

`admissionbridge.Connect` 在启动前连接真实 Attention core、`NewAdmissionRepository(db)` 和同一实际数据库的 S2 service，配置实际事件读取和草稿消费。准入记录固定候选 revision、输入、真实操作者及所有者空间；准入 outbox、候选状态、生产者回执和 S2 草稿共用一个 SQLite 事务。撤销保留原固定输入和准入事件 ID，执行批准由既有 Workspace 单独管理。真实委托入口需要实际决策记录、当前问题委托、固定输入、明确拒答与截断状态、模型和策略版本、当前权限及预算；未配置时返回 `approval_required`。对应决策源尚未装配，未执行真实委托正向。

HTTP 新文件提供 `registerAdmissions(mux, service)`，沿用现有严格 JSON、请求元数据、认证主体和错误映射；共享 `server.go` 的 Services 字段与路由调用、`main.go` 的生产装配交 H 原 owner。当前 B 未宣称默认服务已开放准入路由，实际 HTTP 组合和 S2 准入正向仍待完成。S2 消费者、原生控制器和共享装配文件均未由 B 代改。

## S5 共享迁移与成熟 RPC 依赖

SOURCE `b502f5c3dbf0b57f84a937d314df09d94d164f70` 已普通 push。新业务迁移 `024_judgment.sql` 完整采用 S5 SOURCE `48abd1fe134b88cfe8122d2aee5a35090fc6d62b` / REPORT `95871f3f3d77d27136e3f3195742e6290c587b4a` 的 `judgment_results`、`judgment_feedback`、`judgment_policies` 设计、索引及 JSON、角色、固定版本和外键约束；新增 `foundation_schema` 的 judgment 版本 1。迁移 001–018没有修改，该迁移不依赖尚未提交的成功经验来源元数据。

共享 Go 依赖固定 `github.com/sourcegraph/jsonrpc2@v0.2.1`，官方实际 revision `2cc94179e136b565cb9df7d9d0bff9fc06912ac9`，MIT；下载返回的真实 module/GoMod sums 已由文件编辑工具登记，完整 Go module graph 所需 `gorilla/websocket@v1.4.1/go.mod` sum 同样来自实际下载。`go build -mod=readonly github.com/sourcegraph/jsonrpc2`、`go list -mod=readonly -m -json all`、`node scripts/check-dependencies.mjs` 全部 EXIT 0；证据 `.verification/judgment-jsonrpc2-first-build.log`、`.verification/judgment-go-graph-first.json`、`.verification/judgment-dependencies-first.log`。下载首命令参数解析错误保留，修正后的准确下载及 sums 在 `.verification/judgment-jsonrpc2-download-second.log` 和 `.verification/judgment-jsonrpc2-transitive-download.log`。

实际 S5 公共源码安装目录中的 `vscode-jsonrpc@8.2.1`、`onnxruntime-node@1.30.0` 已直接载入，真实 RPC framing API 和 `InferenceSession.create` 存在，runtime 报告 common/node 均为 `1.30.0`；`.verification/judgment-node-actual-import.log`，模型推理 0。完整读取的实际 package/license 元数据登记了官方 Laya SOURCE `432312b36714d8889eba405507b37eda25257d9c` 的 `laya-ts@0.1.0`、固定模型 revision、MIT runtime 及 `adm-zip@0.6.1` MIT、`global-agent@4.1.3` BSD-3-Clause。Laya 使用准确官方源码 build，未登记不存在的 npm 发布包；未改动 S5 安装目录、Python、root package 或 lock。

`go test -mod=readonly ./internal/workspace/app -run TestNativeOperationInvalidIdentityDoesNotLaunch -count=1 -v` 在新建的实际 SQLite 上执行全部当前迁移及生产身份拒绝流程，EXIT 0，1 PASS、0 SKIP，2.583s；`.verification/judgment-migrations-actual-second.log`，自有数据库在 `.verification/native-operation-actual-3851963715/`。该结果证明新迁移可以实际应用，不代表 S5 判断、缓存、反馈、撤销或 HTTP 已验收；这些实际业务检查继续由 S5 原 owner 完成。全部共享依赖和迁移已交付准确 SOURCE，未等待独立报告阻断其消费。

## 完成收据、经验候选与原操作身份

SOURCE `c715342cff4ccfa1b3198275f981385296fe6641` 已普通 push，包含恢复时保留的五个 B 文件。`AcceptedExperienceService.CaptureCompletion` 只接收实际业务收据引用，通过 `ExperienceCompletionReader` 重查项目、操作者、完成结果和固定成果，再由 `CompletionExperienceInterpreter` 返回结构化解释。候选始终为 `quarantine`；人工批准、降级、共享和批准后检索复用既有 `ExperienceService`。

实际业务来源保存在新增 `023_experience_completion_sources.sql` 的独立元数据表，与经验候选同一 SQLite 事务提交。业务收据保留 `swarm_artifact_accepted` 或 `relay_task_completed` 类型，原生事件外键留空；没有生成替代原生 Event。来源读取、候选读取、CAS 更新和批准列表均核对不可变的 owner、event、时间与来源类型，查询行读取和关闭错误直接返回。

`CompletionExperienceRequest` 提供 `OperationID`、`Evidence`、`ResumeOnly`。服务沿既有 `attention_receipts`，在调用 Interpreter 前事务保存同 owner/event 的唯一 UUID 和固定输入摘要；该 `prepared` 收据只记录已准备的请求身份。再次请求、并发请求和重启后的请求带同一 UUID 及 `ResumeOnly=true`，适配器只能读取和核对原操作；返回解释必须具有相同的 `Provenance.OperationID`。该接口已交原 R 负责人，实际原生并发、受理后未知与重启对账仍为 NOT_RUN。

本轮所有直接 Go 检查均使用进程级 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`、`GOTMPDIR=D:/Astrocyte-agent-assets/20261015/go-temp/backend`，测试 TEMP/TMP 为本树 `.verification/go-temp`；没有修改全局环境。

| 实际命令 | 结果 |
|---|---|
| `go test -mod=readonly -p=1 ./internal/adapters/sqlite -run '^TestAcceptedExperience' -v -count=1` | EXIT 0，3 项检查及提交/实际约束回滚两子项通过，最终 2.418s；`.verification/accepted-experience-operation-current-third.log` |
| `go build -mod=readonly -p=1 ./...` | EXIT 0；`.verification/accepted-experience-operation-build-second.log` |
| `go vet -mod=readonly -p=1 ./internal/attention/app ./internal/adapters/sqlite` | EXIT 0；`.verification/accepted-experience-operation-vet-second.log` |
| 同 D 盘 Go 环境下 `node scripts/check-architecture.mjs` | EXIT 0，33 包；`.verification/accepted-experience-current-architecture-second.log`。首次父 Node 未配置 D 环境的实际标准库错误保留在 `accepted-experience-current-architecture-first.log` |
| `node --test internal/adapters/importers/paper-dom.test.mjs` | EXIT 0，9 PASS、0 SKIP；`.verification/paper-dom-recovery-current.log` |
| `node tests/s1/paper-dom-public.mjs` | EXIT 0，实际已保存 PLOS/ACL 完整 HTML 共 307910 bytes；正文状态、来源标识和 Node/插件提取全等；`.verification/paper-dom-public-recovery-current.log` |
| 变更 Go 文件 `gofmt -l` 与 `git diff --check` | 无输出，EXIT 0 |

实际 SQLite 检查覆盖事务提交、真实主键约束引发完整回滚、同数据库连接复用、错误数据库拒绝、关闭存储拒绝及取消后零证据写入。最终数据库分别保留在 `.verification/accepted-experience-actual-{1075054943,1488676106,1864100285,3037321422}/`；这些检查没有创建完成来源、模型输出或经验候选，不作为完成收据正向验收。

经 M-product 明确临时放行，B 在 `release-judgment-20261015` 仅使用文件编辑工具消除 `dependencies.lock.json`、`go.sum` 冲突并 stage 两件。JSON 实际解析及两项 diff 检查 EXIT 0，index blob 分别为 `64c1432f721a6371a5c8e4badce719cd1bbbbd92`、`def74ba382fec897f8f9bc9a0d40f5bb6e1bb88f`；现有 MCP、RPC 和 Laya 依赖内容保留。其余 16 项 S2 冲突及 S5 其他 WIP 未代改，没有执行 merge 生命周期操作；两件临时范围已明确停止写入并归还。

实际 S3 来源桥接、原 R Interpreter、HTTP 装配、原生请求并发与未知对账、人工批准后第二任务引用均未完成，继续由原负责人串行衔接。完整 Go 与产品组合验收继续保持此前未通过状态；没有把阶段源码、静态检查或数据库机制通过登记为业务完成。

## 就绪请求诊断与原生 Go 依赖元数据

SOURCE `efa137eb36177f69643c750c26ad8eb59b446741` 已普通 push，仅修改 `tests/s1/server.mjs`、`scripts/check-dependencies.mjs`。就绪 GET 失败保留实际 URL、自有 PID、`exit_code`、`signal_code` 与原 cause，成功响应完整消费正文；现有请求期限、总期限、启动等待、端口选择与全部生命周期断言保持。Go 依赖检查通过 `go mod edit -json` 比较全部 direct requirements 和工具链，通过 `go list -m -json` 检查实际选择的版本、间接属性和 replace，不解析 go.mod 字符串。

`node --test tests/s1/server-lifecycle.test.mjs` 在原 helper 和当前 helper 上分别 EXIT 0，3 PASS、0 SKIP，164.304s / 87.586s。实际检查包括普通/Relay 自有目录、HTTP、SQLite、session、停止重开、Vite 代理、重复绑定拒绝与原端口重启。证据为 `.verification/server-readiness-recovery-first.log`、`.verification/server-readiness-context-second.log`；本树没有复现 F 的原 `read ECONNRESET`。F 的 `TsIlGq` 日志顺序显示 helper 在启动内部清理、尚未发出会话请求，网页就绪 GET 为当前排查位置，原错误具体请求 URL 和当时子进程退出状态未封存，根因保持 UNKNOWN。新增诊断及响应消费不作为该故障已修复的结论。

`node scripts/check-dependencies.mjs` 在同 D 盘 Go 环境下 EXIT 0，实际依赖、锁文件、安装与 summarize 0.25.1 CLI 版本通过，日志 `.verification/go-dependency-native-json-current.log`。两个变更文件 `node --check` 与 `git diff --check` EXIT 0。

按 M-product 已 push 的 `bbc64c3` 临时计划，B 仅在 H `works-relay-http` 编辑并 stage `go.mod`、`go.sum`、`dependencies.lock.json`。实际全模块 JSON、`go mod verify` 及 8 项 direct pin 原生 JSON 比较均 EXIT 0，diff 检查通过、未合并项为零；原件 `.verification/h-sdk-current-{module-graph.json,verify.log,pins.log}`。index blob 依次为 `e5fe6c301d7f4cb7f62dc46cbe31273bdb1f9433`、`50cb9ecc67890807e03828a3683dc4f6098ce98d`、`3b40e4f8bc1e699e5bb470f7e854a6e8c13db99a`。H 旧检查器对 grouped require 的首 RED 保留于 `.verification/h-sdk-current-dependency.log`；当前成熟元数据检查器的准确 SOURCE 已交 H，由 H 消费。没有修改 H 其他源码或操作其 cherry-pick 生命周期；三件临时范围已停止写入并归还。

根已将 `internal/adapters/httpapi/judgment*.go` 及同前缀测试交唯一原 S5；B 本人确认该前缀无文件/WIP，并已停止写入。B 继续共享依赖、迁移及 `config.go`，S4 可信配置等待其原负责人准确接口与资产交接。冷 Registry 原生会话测试整改等待实际自有项目和原会话来源，真实 Interpreter、人工批准后使用及全部组合验收仍未完成；没有以机制通过替代这些结果。

## 模型身份与独立版本的实际来源

SOURCE `f76e25b84b29495994f2e5877b586d4c22669d3f` 已普通 push。`ExperienceInterpretationProvenance` 新增 `ModelIdentitySource` 和 `ModelVersionSource`；模型身份明确区分 `configured_route` 与 `provider_reported`，独立版本不可观测时两个版本字段均保存 `unknown`。已知独立版本仅接受实际 provider 报告，禁止重复模型 ID。完整响应、明确非拒绝/非截断、原 `OperationID`、policy 与真实完成来源要求保持，候选仍进入 `quarantine`。

`go test -mod=readonly -p=1 ./internal/attention/app ./internal/adapters/sqlite -run '^(TestCompletionInterpretationRules|TestAcceptedExperience)' -v -count=1` EXIT 0，规则检查及三项实际 SQLite 检查通过，日志 `.verification/completion-model-identity-rule-first.log`；SQLite 最终 16.237s。新增规则检查仅使用明确输入和期望值，没有模型调用、控制器或来源收据。`go build -mod=readonly -p=1 ./...`、目标包 `go vet`、两文件 `gofmt -l` 和 `git diff --check` 均 EXIT 0，无格式输出；全部 Go 命令沿进程级 D 盘缓存与临时目录。实际 R Interpreter 与来源桥接仍未运行，不据此宣称原生验收或不可变模型复现。

## 真实剩余限制

- 当前全包检查为 RED，公开网络存在 TLS `EOF`、`fetch failed`；源请求失败未用本地伪造响应替换，未加 skip 或放宽断言。
- 原有 attention 应用测试仍有内存 Repository/ObjectStore 与替代模型。经双方零 WIP checkpoint 确认，协调者已把 `internal/attention/app/*_test.go` 与 `tests/s1/automatic_test.go` 共十四个既有文件及同目录新测试交给唯一第三层 T，精确起点 `b470ecd736af261a10e4c4822bdcd3b02d0da454`；B 从此排除这些测试写域，生产需求由 T 提交 Handoff。
- B 本写域仍有注册项目 adapter 测试的替代来源；M5 MCP 与 delivery 测试的替代 backend、失败存储和确认丢失；SQLite 其他 `local_agents*_test.go` 的替代原生控制器与引用读取。后续各项整改继续保留完整覆盖。
- `tests/s1/fixtures.mjs`、默认 acceptance 与部分 web E2E 仍依赖虚构论文/视频来源。B 继续整改共享输入，web 改动交 F，不能把现有默认套件通过记为真实来源验收。公开视频 whisper 转写无时间区间，不能伪造已有 positive 时间区间覆盖所需原件。
- 全写域英文注释清理尚未完成。原生格式 helper 的解码路径及其他同类异常需要继续核查；已增加格式门禁不等于全域零静默解析。
- 实际 Codex controller / SQLite / HTTP / 自产会话逐 event_id 验证归 R/H，当前 B 没有运行该组合验收。发布完整 E2E、Workbench/Relay 集成与 ZIP 解压验收尚未运行。
- B README 一次实际 inference 已核对 HTTP、模型返回、原 native_id、SQLite 与重启；当次进程启动 argv 未独立留存。R 原 owner 已核实原 controller Stop 后无 capture pointer、新 Registry 无 ownership，不能按旧 native_id 重建控制器或扫描私人历史。该旧结果的有效工具集合保持未知，不重复成功推断；新文本处理的正选证据接口需模块确认精确 R 文件消费方式后继续 B/R 互斥接入。
- Workbench Reader 十项检查覆盖实际公开生产回放、实际 Sleep 和当前自产 native 副本。H/I 仍需在既有路由组合精确源码；发布完整二进制、浏览器 E2E 与两次实际 00:30 触发验收尚未运行。H 原会话部分完整度未知不能因 Reader 查询通过而提升。
- 当前 arXiv DNS 实际响应 `198.18.0.71`，系统网络配置不在 B 授权写域。完整 Go 检查失败已提交模块负责人；公开 source 与 CLI positive 断言保持，未添加 skip 或伪造响应。
- 当前 B 源码的 Mission 创建仍返回生产 501；最新用户授权的 S2/S3 实际业务实现已移交固定 owner，尚未消费其 SOURCE 或运行该功能验收。V4、人评、正式库、冻结脚本、引擎、现有 schema、迁移、权重和日程保持原边界，新增共享 DDL 仅按明确跨域需求实施。

## 未执行操作与后续交付

部分 SOURCE/REPORT 普通 push 遇到 GitHub `SSL_ERROR_SYSCALL`，原件保留；上述已列源码及报告均已普通 push，未 force push。Reader 前四次握手失败保留在 `.verification/workbench-source-push-{first,second,third,fourth}.log`，第五次正常推送成功。本报告作为独立 REPORT 追加提交并推送。阶段 SOURCE/REPORT 不结算整轨，后续源码与报告继续追加；最终精确 SHA 和剩余限制由最后一次报告登记。
