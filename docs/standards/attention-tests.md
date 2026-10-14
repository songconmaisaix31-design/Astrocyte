# Attention 测试整改记录

本轨工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/product-attention-tests-20261014`，分支为 `songconmaisaix31-design/product-attention-tests-20261014`，接管基线为 `b470ecd736af261a10e4c4822bdcd3b02d0da454`。Orca 终端 `term_c1d90d02-51be-4443-9afe-ddce0615acb4`，Task `task_5f0a79b6a17a`，Dispatch `ctx_8674c456f49f`；实际终端显示 Codex / GPT-6.1-Sol high fast。

本轨唯一写入范围为 `internal/attention/app/*_test.go`、`tests/s1/automatic_test.go` 和本文件；生产、依赖、native/relay 与其他测试归原人员。原始错误、实际来源、模型响应和会话证据只保留本树已忽略 `.verification`，不纳入 Git 或产品接口。

## 已提交源码阶段

SOURCE `a1e1195464c2f989083780a7b2bc0ade0cc07ab6` 已普通提交并 push：候选门槛、人工使用动作、固定版本项目引用、分类和空间、profile CAS 使用外部测试包与生产 SQLite / Reader / ObjectStore。新增两个本目录专用测试文件，生产代码没有修改。纯排名参数测试继续直接检查生产纯函数。

第二阶段 SOURCE `03af9854dac57f69fc206d4d647ad0c518b98b20` 已普通提交并 push。经模块明确授权，普通 `--no-ff` 消费 B 的前置 SOURCE `01753cf4fb455f35aa4953bf3eaa9a6388c4be5a`，没有手工修改 B 的生产文件。T 本阶段实际修改 9 个 App 测试文件，新增专用 `attention_service_test.go`、`imports_service_test.go`、`jobs_service_test.go` 和 `tracking_production_test.go`；新增测试均在约定的 `*_test.go` 写入范围内。

第三阶段 SOURCE `8525843d458f2d3a5d272b239ab76ac50c3223a5` 已普通提交并 push。迁移生产 factory 的全局可用性状态、真实保存页面后的 SQLite 关闭 / 重开 / 恢复 / 无 Reader 完成，以及启动恢复的两种后续作业情形；没有手工创建 completed 页面、成功作业、metadata 或时间。实际 API 最后一页提供 1 项，原 `Limit=100`、1 项保存页面、成功 / LastSuccessAt / nonstale 与旧失败不重写的断言继续保留。目录请求与未完成事务由真实生产实现执行。

第四阶段 SOURCE `bee27ca1f710b6b077e51f78d5ec744f24d0ded6` 已普通提交并 push，唯一业务测试改动为 `tests/s1/automatic_test.go`。删除替代 processor，改用实际构建的 server / HTTP / session bootstrap / SQLite / objects；实际三类权限拒绝和进程终止 / 两秒真实预算 / 同库重启路径已通过。模型未知交付分支保留全部原输入、版本、幂等、payload / caller、无结果 / 无候选 / 无重发断言，仍未执行，也没有把 `ExternalStarted` 提升为原生 Send 证据。

第五阶段 SOURCE `aa7257e66bff29140a826afdaab93f36e35958d9` 已普通提交并 push，只修改 `video_reuse_test.go`。原替代 SourceReader 已移除，使用 V 明确交付的同 BV 两个 immutable 实际文本导出和生产 Reader / SQLite / ObjectStore，保留 alias 首读 1 次、显式 B refresh 与 normal latest 2 次、返回 A 第 3 次读取 / 3 个作业 / 2 个 revision / B head / 旧 A revision1 / 同 material，以及 multipart 单独身份的全部断言。p2 只提交其真实 URL，没有借用 p1 原件；生产 `summarize_url` 对不支持的 multipart 返回实际 validation_failed，零新资料。

第六阶段 SOURCE `61a012464a22d60ef421ea1fdaa5541ef65f1f7e` 已普通提交并 push，修改两个 App tracking 测试文件。经模块 `msg_64b573bcf0a3` 授权，普通消费 B 准确 cursor SOURCE `52bd568811cbcdf59540ba18eee5a37c9d32002b`；同原获准公开收藏的真实 100 条 App / SQLite 验证通过。原 110 条 prefix 缓存断言迁移为真实 `100 + 10 + 前缀100` 三窗口，全部旧记录不标 stale；另新增原 partial 场景的实际只读半段：真实末两页各 1 项、原 cap100、第二次读取发生真实 context 终止、生产 Run 自动有界重试同 cursor，3 次实际 reader 调用 / 2 attempts / 同 operation 与 deadline / 2 条非 stale 记录均通过。观察 helper 只委托生产 Reader，不提供替代页或错误；原模型 unknown 半段继续等待实际模型窗口。

本测试要求显式绝对路径 `ASTROCYTE_TEST_VIDEO_EXPORT_A` 与 `ASTROCYTE_TEST_VIDEO_EXPORT_B`，没有个人路径源码默认值、fallback 或 skip。此次 A 路径为 `C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/.verification/video-source-third/video-original-text.txt`，B 为 `C:/Users/DW/orca/workspaces/Astrocyte/product-video-20261014/.verification/video-timed-first/video-original-text.txt`，均按模块与 V `msg_43e318f2b008` 的精确只读授权消费。production `summarize_markdown` loader 保留已有真实正文；该范围没有宣称 Markdown 自身证明 952 个时间片段，也没有调用新提取 / ASR / 模型 / controller。

经模块允许，普通 `--no-ff` 消费 B 前置 `3384a89aac3be6d5670d756eb2d44263784d0cb8`；其中只有共同 owned execution 口，text / bridge 的原 registry、作业 / operation、活会话和原档 / Send 映射尚未交付。T 没有手工修改该生产输入。

新增 `http_evidence_test.go` 只委托真实 HTTP Transport，并通过 `io.TeeReader` 将其实际响应原件保存到本机忽略目录；没有替代请求 / 响应、认证 Cookie 或凭据，也没有把原件写入 Git / HTTP / 产品字段。它仅用于定位真实 provider / pagination 问题。

已迁移共享人工业务的全部 7 项断言组、5 项读操作作业断言组，以及来源绑定、官方 arXiv 两版和别名去重。保留原成功、失败、回执、12 个并发请求、2 个 Worker / 8 个作业、旧正文与 PDF、100 条截断要求、事务外 I/O、人机使用分离、未知维度及反馈历史。观察 helper 只记录或同步生产调用边界，所有来源与对象结果均委托生产 Reader / Store；实际 SQLite 负责事务及回执。

mutable 参数与真实可变文件使用明确 manual 导入；不声称远端 arXiv 或视频发生了修改。原输入分类的 21 个参数只验证生产纯函数，另以相同 21 个参数在真实 SQLite 验证队列与回执去重，未将参数分类填成网络成功。

人工导入正文是显式 manual 请求参数，调用生产 Reader 完成读取、校验和对象持久化；此阶段不声称外部网络或模型成功。对象调用计数器只委托生产 Store，没有提供替代对象响应。失败事务通过测试独占 SQLite 中的真实 outbox SQL 缺失关系失败，保留原失败、回滚和回执断言。

## 实际验证

| 命令 | 结果 | 本机原件 |
|---|---|---|
| `go test -mod=readonly ./internal/attention/app ./tests/s1 -count=1`，接管基线 | EXIT0：App 6.041s，S1 60.962s；含旧替代实现，不代表整改后真实模型验收 | `.verification/baseline.log` |
| 首次迁移 App 全包 | RED：未使用变量造成编译失败 | `.verification/migration-first.log` |
| 第二次迁移 App 全包 | RED：Crossref 请求 EOF；outbox SQLite constraint 错误分类与旧替代错误预期不同 | `.verification/migration-second.log` |
| `go test -mod=readonly ./internal/attention/app -run 'TestCandidateReady\|TestExplicitMention\|TestScopedAgent\|TestOptionalLinked\|TestClassification\|TestSpaceAnd\|TestMaterialList\|TestImportReturning' -count=1` | EXIT0，8 项真实持久化测试 | `.verification/migration-targeted.log` |
| 上述定向命令加 `TestProfileCAS` | EXIT0，9 项；仅定向覆盖 | `.verification/migration-profile.log` |
| 首次 `go build` 包含 `./tests/s1` | RED：该目录只有测试文件，不能作为 build 目标 | `.verification/build-first.log` |
| `go build -mod=readonly -o .verification/astrocyte-stage1.exe ./cmd/server` | EXIT0 | `.verification/build-stage1.log` |
| `go test -mod=readonly ./internal/attention/app -run 'TestFixedArxiv\|TestMutableImport\|TestImportFixedArxivBoundary' -count=1` | EXIT0，45.208s；实际官方 Atom/PDF、旧版本和别名复用；另有纯分类与人工导出范围 | `.verification/import-production-first.log` |
| `go test -mod=readonly ./internal/attention/app -run 'TestImportReceipts\|TestConcurrentImport\|TestDistillationLayers\|TestFeedbackCAS\|TestOutboxFailure\|TestHumanQueries\|TestWithdrawnSource' -count=1` | EXIT0，7 项实际持久化业务测试 | `.verification/manual-production-second.log` |
| 读操作首次实际失败场景 | RED：不存在文件是不可重试输入错误，不能当作临时网络失败的错误文案 | `.verification/jobs-production-first.log` |
| `go test -mod=readonly ./internal/attention/app -run 'TestReadOnlyFailure\|TestRunningCancel\|TestCancellationReceipt\|TestWorkerPoolFinite\|TestCanonicalContract' -count=1` | EXIT0，5 项；实际 HTTP Client 期限、生产 Reader、SQLite、有限作业与取消回执 | `.verification/jobs-production-fourth.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestSourceIdentityAndPublicOnlyBinding$' -count=1` | EXIT0；人工绑定参数及 SQLite，不声明对应外部目录成功 | `.verification/tracking-bind-production.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestPublicTrackingSourceCapacity$' -count=1` | RED：实际获准 UID `3494358764489275` 上传目录仅观察到 3 项，`has_more=false`，无法验证 100 条截断；未降低门槛 | `.verification/listing-production-first.log` |
| `go test -mod=readonly ./internal/attention/app ./tests/s1 -list '^Test'` | EXIT0，当前编译及测试清单；没有运行实际模型，不代表全包通过 | `.verification/coverage-current.log` |
| `go build -mod=readonly -o .verification/astrocyte-stage2.exe ./cmd/server` | EXIT0 | `.verification/build-stage2.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestGlobalSelectedProcessorStatusDoesNotInventAvailability$' -count=1` | EXIT0，使用生产 SelectedTextFactory 和真实 SQLite，保持未选项目的配置 / 模型未知 | `.verification/processor-status-production.log` |
| `go test -mod=readonly ./tests/s1 -run '^TestAutomaticHTTPRejectsUntrustedCallerWithoutSideEffects$' -count=1` | EXIT0，9.258s；实际生产服务的无 session、无效 Agent 与跨来源请求均 403，SQLite 作业 / 回执 / 沉淀 / 资料均零 | `.verification/s1-http-denied-first.log` |
| 保存页面首次及后续真实前置验证 | RED：完成事务限制的首次配置错误、creator offset 请求未取得页面、真实事务错误向调用者立即返回；全部原件保留 | `.verification/tracking-final-page-first.log`、`-second.log`、`-third.log`、`-favorites.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestSavedFinalListingPageCompletesWithoutReader$' -count=1` | EXIT0，2.021s；同获准 UID 收藏夹实际最后一页、完整 cap100、SQLite 实际关闭重开后无 Reader 完成 | `.verification/tracking-final-page-reopen.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestStartupResumesCurrentFailureButNotSupersededFailure$' -count=1` | EXIT0，3.652s；两子例均以真实目录、失败完成事务和实际数据库恢复创建前置，后续成功作业来自实际请求 | `.verification/tracking-startup-production.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestPublicFavoriteSourceCapacity$' -count=1` | RED：同已批准 UID 匿名目录观察 2 个收藏夹，选中夹报告 459 项；分页实际 provider_unavailable，尚未验证 100 条截断 | `.verification/favorites-capacity-first.log`、`-second.log`、`-capture.log` |
| 当前相关包 `-list '^Test'` | EXIT0；编译及测试清单，不代表全包或模型通过 | `.verification/coverage-stage3.log` |
| 使用现系统匿名 loopback 代理的公开 paper search 定向命令 | EXIT0；仅受控 Go 子进程显式 HTTPS_PROXY / HTTP_PROXY、NO_PROXY loopback；系统 / DNS / SSL / 凭据未修改 | `.verification/paper-search-existing-proxy.log` |
| `go test -mod=readonly ./internal/attention/app -count=1` | RED：固定官方 arXiv 导入、Crossref 请求、creator3 无法验证100、favorites459 实际分页失败；旧替代模型的结果仍不作真实模型验收 | `.verification/app-default-stage3.log` |
| `go test -mod=readonly ./tests/s1 -run '^TestReadOnlyImportHTTPProcessExitPreservesDeadline$' -count=1` | EXIT0，20.021s；实际 running 作业中终止 owned server，实际预算到期后重启，保留 failed / nonunknown、原 operation / deadline / attempts=1、payload / caller；到期重试实际 400 | `.verification/s1-readonly-exit-first.log` |
| `go test -mod=readonly ./tests/s1 -run 'TestTrackingIdentityMigrationPreservesLegacyItems\|TestPublicArxivAliasesPinOriginalVersion' -count=1` | EXIT0；已有迁移参数范围与官方原件 / 别名 / 已发布版本 / 未发布版本失败，未调用模型 | `.verification/s1-nonmodel-public-first.log` |
| 相关两包 `-list '^Test'` 与 `go build -mod=readonly -o .verification/astrocyte-stage4.exe ./cmd/server` | EXIT0；编译清单与生产服务构建 | `.verification/coverage-stage4.log`、`.verification/build-stage4.log` |
| 实际两视频导出首次 A/B/A 验证 | RED：新增 p2 失败断言的预期错误类型与生产拒绝规则不同；A/B/A 原断言已通过，首记录保留 | `.verification/video-reuse-actual-exports-first.log` |
| 显式 `ASTROCYTE_TEST_VIDEO_EXPORT_A/B` 后，`go test -mod=readonly ./internal/attention/app -run '^TestNormalVideoReuseExplicitRefreshAndABAHead$' -count=1` | EXIT0；实际已有两版导出、alias、refresh、A/B/A、SQLite / objects 与 p2 实际拒绝 | `.verification/video-reuse-actual-exports-explicit.log` |
| 消费准确 cursor SOURCE 后，`go test -mod=readonly ./internal/attention/app -run '^TestPublicFavoriteSourceCapacity$' -count=1` | EXIT0；同原获准匿名公开夹，实际 metadata100 / has_more / SQLite，资料零 | `.verification/favorites-capacity-fixed-cursor.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestBoundedPrefixDoesNotMarkUnobservedOlderCacheStale$' -count=1` | EXIT0；真实三窗口，保留原 110 条记录 / has_more / 全部 nonstale 断言 | `.verification/tracking-prefix-actual-first.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestActualPartialListingFailureRetainsCursorAndBudget$' -count=1` | EXIT0；实际 1 项缓存 / stale 失败、同 cursor 自动安全重试、总 2 项 / attempts2 / 3 次真实调用 / 原预算 | `.verification/tracking-partial-actual-first.log` |

## 仍在执行

剩余 App 共享 MemoryRepo / MemoryObjects、替代模型、provider 目录及 S1 unknown 生命周期继续整改；本轨未完成，未发送 worker_done。已有成功和失败断言、未知状态继续保留，不新增 skip、不删除断言、不替换为伪外部成功。

生产 `agents/text.go` 的 selected-text 目录已通过 B 前置 SOURCE 改为 registry owned root。本轨后续 Go 命令的 `GOTMPDIR / TEMP / TMP` 显式指向 `.verification/go-temp`；S1 子服务另显式使用其 owned evidence directory 内的 native 临时目录。

S1 首次运行在第一次导入请求的身份构造前 RED，8.36s，没有原生 probe 或模型请求；原件 `.verification/s1-native-first.log` 与 `.verification/s1-native-1981875207/server.log` 保留。模型前置问题 `msg_18fa60d90459` 已由模块回复：只可消费 B 的共同 execution 接口，text / bridge 映射未交付。T 不以 `ExternalStarted` 或工具调用零单独证明实际 Send / 工具关闭，不扫描原生 home、不重发 B 的旧成功请求。

收藏夹真实响应原件位于 `.verification/public-http-1166636830/response-1.bin` 至 `response-9.bin`。只读派生为资源页 `20 / 20 / 19 / 0` 项、均 `code=0`、`has_more=true`；目录元数据为 459 项。生产 `PublicListing.ReadPage` 按 returned observed 推进窗口，而 provider 页位置按 requested page size 计算，第三页不足 20 后下一请求落到空槽。这一具体问题已直接 Handoff B `ctx_7217777c7bf5`；T 没有修改 provider、补造行数 / metadata、绕过网络权限或降低 cap。creator3 的有限列表观察、收藏459的目录前置、100 截断失败和最后一页缓存成功分别保留。

根 `msg_8d2dcf0e1841` 批准另选足量、匿名公开可访问 Creator 作为工程 QA 来源；正式用户案例与科研成绩不随之改变。T 对三个现有公开发布者候选 `49683148`、`1408168321`、`946974` 使用原生产 `bilibili-listing.py` 与已安装固定 yt-dlp 查询首 metadata 页，实际均 HTTP412，stdout 原件为零，真实发布总数仍 UNKNOWN。原错误仅在 `.verification/creator-<UID>-first.error.log`，不进 stdout / Git；没有登录、Cookie、凭据、源外下载或生成数量。三个同类拒绝已向模块及 B 报告，不继续批量盲扫。

## 后续实际验收条件

根 `msg_4da181ce8f60` 明确 selected-text 原生身份与模型验证延期至明天 v1.4 语义稳定后的共同身份 / 目标核查窗口。今晚保持实际模型、原生 controller 启动均为零；包含该隐式路径的 `go test -mod=readonly ./internal/attention/app ./tests/s1 -count=1` 保持 NOT_RUN，不新增 skip、不删除 case 或隐藏默认门槛。

继续验收需要 B 的精确 text / bridge SOURCE 及原 operation → 原 registry / session / 正选原档 / 实际 argv / Send 受理 / Stop 证据，以及可访问的足量公开 Creator。根已裁决 100 预算为实际 metadata，允许跨缺失槽到第 100 条，真实 Observed19 保留，100 即停止、超时及无进展失败继续保持。准确 cursor 已消费，收藏100 / 110缓存 / partial只读半段均有本轨独立实际结果；Creator 原3与三个候选412仍保持原范围失败，未由收藏结果替代。

V 的两真实文本导出已支持本轨版本复用验证，完整时间 / 默认提取验收仍归 V 的真实生产路径，不借本轨 Markdown 版本复用结果填其全模块通过。原生只提供固定公开输入，工具配置由真实运行证据核对；未确定部分保持 UNKNOWN。完整 App 自动处理 / tracking 及 shared helper 的替代实现整改仍未完成，明天由本 owner 继续同一写入范围。T 没有发送 worker_done 或发布产品。

完整成功、权限拒绝、未知交付、重启和整改后相关全包测试尚未完成；本文件所列定向结果不构成全部产品验收。剩余旧共享 helper / 自动模型 / tracking 外部替代测试仍需整改，目录容量不足及模型证据前置分别记录；未用旧替代测试的 PASS 填成本轨完整通过。
