# Attention 测试整改记录

本轨工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/product-attention-tests-20261014`，分支为 `songconmaisaix31-design/product-attention-tests-20261014`，首次接管基线为 `b470ecd736af261a10e4c4822bdcd3b02d0da454`。2026-10-15 恢复基线为 `5b45b0c2a4c6ca8c33a0a1ae8478ec4481877e04`，接管时工作树干净；当前 Orca 终端为 `term_e39f5608-91dd-4c07-856a-9ebe7b9bb555`，Task `task_5f0a79b6a17a`，Dispatch `ctx_491ad741b3cd`。公开 terminal 元数据和模块 worker-list 均核实实际客户端为 Codex、模型 `gpt-6.1-sol`，终端显示 xhigh fast。

本轨唯一写入范围为 `internal/attention/app/*_test.go`、`tests/s1/automatic_test.go` 和本文件；生产、依赖、native/relay 与其他测试归原人员。原始错误、实际来源、模型响应和会话证据只保留本树已忽略 `.verification`，不纳入 Git 或产品接口。

## 已提交源码阶段

SOURCE `a1e1195464c2f989083780a7b2bc0ade0cc07ab6` 已普通提交并 push：候选门槛、人工使用动作、固定版本项目引用、分类和空间、profile CAS 使用外部测试包与生产 SQLite / Reader / ObjectStore。新增两个本目录专用测试文件，生产代码没有修改。纯排名参数测试继续直接检查生产纯函数。

第二阶段 SOURCE `03af9854dac57f69fc206d4d647ad0c518b98b20` 已普通提交并 push。经模块明确授权，普通 `--no-ff` 消费 B 的前置 SOURCE `01753cf4fb455f35aa4953bf3eaa9a6388c4be5a`，没有手工修改 B 的生产文件。T 本阶段实际修改 9 个 App 测试文件，新增专用 `attention_service_test.go`、`imports_service_test.go`、`jobs_service_test.go` 和 `tracking_production_test.go`；新增测试均在约定的 `*_test.go` 写入范围内。

第三阶段 SOURCE `8525843d458f2d3a5d272b239ab76ac50c3223a5` 已普通提交并 push。迁移生产 factory 的全局可用性状态、真实保存页面后的 SQLite 关闭 / 重开 / 恢复 / 无 Reader 完成，以及启动恢复的两种后续作业情形；没有手工创建 completed 页面、成功作业、metadata 或时间。实际 API 最后一页提供 1 项，原 `Limit=100`、1 项保存页面、成功 / LastSuccessAt / nonstale 与旧失败不重写的断言继续保留。目录请求与未完成事务由真实生产实现执行。

第四阶段 SOURCE `bee27ca1f710b6b077e51f78d5ec744f24d0ded6` 已普通提交并 push，唯一业务测试改动为 `tests/s1/automatic_test.go`。删除替代 processor，改用实际构建的 server / HTTP / session bootstrap / SQLite / objects；实际三类权限拒绝和进程终止 / 两秒真实预算 / 同库重启路径已通过。模型未知交付分支保留全部原输入、版本、幂等、payload / caller、无结果 / 无候选 / 无重发断言，仍未执行，也没有把 `ExternalStarted` 提升为原生 Send 证据。

第五阶段 SOURCE `aa7257e66bff29140a826afdaab93f36e35958d9` 已普通提交并 push，只修改 `video_reuse_test.go`。原替代 SourceReader 已移除，使用 V 明确交付的同 BV 两个 immutable 实际文本导出和生产 Reader / SQLite / ObjectStore，保留 alias 首读 1 次、显式 B refresh 与 normal latest 2 次、返回 A 第 3 次读取 / 3 个作业 / 2 个 revision / B head / 旧 A revision1 / 同 material，以及 multipart 单独身份的全部断言。p2 只提交其真实 URL，没有借用 p1 原件；生产 `summarize_url` 对不支持的 multipart 返回实际 validation_failed，零新资料。

第六阶段 SOURCE `61a012464a22d60ef421ea1fdaa5541ef65f1f7e` 已普通提交并 push，修改两个 App tracking 测试文件。经模块 `msg_64b573bcf0a3` 授权，普通消费 B 准确 cursor SOURCE `52bd568811cbcdf59540ba18eee5a37c9d32002b`；同原获准公开收藏的真实 100 条 App / SQLite 验证通过。原 110 条 prefix 缓存断言迁移为真实 `100 + 10 + 前缀100` 三窗口，全部旧记录不标 stale；另新增原 partial 场景的实际只读半段：真实末两页各 1 项、原 cap100、第二次读取发生真实 context 终止、生产 Run 自动有界重试同 cursor，3 次实际 reader 调用 / 2 attempts / 同 operation 与 deadline / 2 条非 stale 记录均通过。观察 helper 只委托生产 Reader，不提供替代页或错误；原模型 unknown 半段继续等待实际模型窗口。

第七阶段 SOURCE `a7370bf1c1086e3cacd612435e15f3f26feb58d2` 已普通提交并 push，修改 `automatic_test.go`，新增本目录专用 `automatic_permissions_service_test.go`。原 Agent 拒绝、空来源许可和未配置处理器断言迁到生产 Service / 官方固定 arXiv Reader / SQLite / ObjectStore；原 `TestAutomaticPriorScopeCannotExpandSelectedSources` 测试名保留并迁到外部包，先通过实际人工 RecordDistillation 保存带有未选上下文的记录，再验证原请求范围拒绝。真实 manual 来源即使其实际 key 已在许可列表，也按真实 provenance 被拒绝；未更改来源 provenance、写入假成功记录或提供替代模型。旧配置变更与无效模型引用断言仍保留等待原生门槛。

第八阶段 SOURCE `870b6561472a978bb684f4c1c83b8e3fae40911f` 已普通提交并 push，只修改 `tracking_production_test.go` 和 `tests/s1/automatic_test.go`。真实原公开 Creator 的现有可用 metadata 支持推荐前正文选择拒绝、未配置推荐失败及 Agent 拒绝检查，没有重复收藏100验证或将有限目录填成容量通过。实际 HTTP 生产服务验证固定公开项目的未引用资料拒绝、未观察原生配置的明确 failed / nonunknown / ExternalStarted=false，以及撤销模型许可后的新请求拒绝；本机 SQLite 中 session / distillation / opportunity 均零，只有一项真实未配置失败的 distillation 作业。原 S1 unknown 作业种类查询同样使用 SQLite `json_extract(data,'$.kind')`，该模型路径仍 NOT_RUN。

第九阶段 SOURCE `ea8e7e35412e94e10526728fd68c67fffa517469` 已普通提交并 push，新增本目录专用 `admissions_service_test.go`。按模块 `msg_ae8e7fcc175f` 和 SPEC 7.1，使用实际官方固定论文导入、真实对象与 SQLite、明确人工待查问题，经生产 CreateOpportunity 生成 `ready_for_review` 的 revision1 前置；固定引用保存不变，GoalRefs 为空，GoalProgress 与 CompositeScore 保持未知。创建候选只产生一项真实 `attention.opportunity_created` 事件，没有对象发布、准入事件或原生会话；没有 fakeJudge 或伪造事件。该阶段覆盖准入前置；新生产入口的准确消费和实际验证见第十阶段，前置结果不等于准入或 S2 / S3 完整验收。

第十阶段 SOURCE `1d134eff5e4b1f4c1e173f242b8fbe9a6b7952e6` 已普通提交并 push，唯一测试改动为本目录 `admissions_service_test.go`。按 B `msg_d75f623c74d3`，准确 SOURCE `e96f8d78f686f6db0f0a9edb5e15cc57ddb8aa1a` 经普通 `--no-ff` 消费到 `8337780a55cb45c870d14dd207b89a239fd6d7d4`，没有手工修改生产或其他 owner 文件。实际 AdmissionService / AdmissionRepository / admissionbridge.Connect / S2Service 使用同一 SQLite；普通 Agent、无判断授权、真实 incubating 候选、实际新 revision 后旧 revision 和真实 outbox SQL 故障回滚均通过。随后人工准入在实际 S2 consumer 的 authority / 固定来源 guard 失败，原 human PolicyVersion 为空，未补造 policy 或判断；完整事务回滚后 Admission / Proposal 为零，事件 / 回执不变，原 Ready version 保留，八项 execution / session 表为零。人工准入成功后幂等、草稿事件重放、namespace 隔离、撤销、来源撤回和零执行断言已写入，因前置真实失败保持 NOT_RUN。

第十一阶段 SOURCE `41f88230997382ea0972652ab3f15135c75a90f1` 已普通提交并 push，远端分支同值，实际只修改 `admissions_service_test.go`、`automatic_permissions_service_test.go`、`http_evidence_test.go`、`production_fixture_test.go`。恢复基线已包含 S2 准确 SOURCE `55b48b755d7d9790d5f6f58f53724a0e4f59ce7e` 的普通合并；没有手工修改生产文件。

按模块 `msg_f14bea6c12a5`，原 `TestActualAdmissionServiceAndDraftLifecycle` 继续以官方固定论文运行全部准入生命周期断言；新增独立 `TestActualPublicGitAdmissionServiceAndDraftLifecycle`。两个测试共用完整断言，保留所有原成功和失败条件，并增加消费者回执、逐 event_id 定位以及两次实际关闭 / 重开 SQLite 的检查。固定 Git 场景从本获准公开 Astrocyte 仓库的实际 HEAD 读取原始 README，经生产 Reader 的显式 `manual` 导入形成固定 revision1，实际来源和错误仅存本机忽略目录。其 provenance 保持 manual，没有外部读取或模型成功声明，也没有网络失败后改换来源的执行分支。

准确 SOURCE 的固定 Git/manual 完整链已通过：Agent / 未授权判断 / incubating / 旧 revision / 实际 outbox SQL 故障拒绝、人工准入、原请求幂等、真实草稿事件重放、其他所有者拒绝、撤销及撤销幂等、来源撤回后的停止分发。原始 admitted / revoked event_id 均在真实 outbox 逐一定位；两次数据库关闭重开后，准入、草稿、撤销和双方回执保持完整 JSON 契约及全部时间精度。人工 PolicyVersion / Model 继续为空、GoalRefs 与综合分保持未知，八项 execution / session 表始终为零。原官方论文完整路径遇到实际 Atom `HTTP 429`，仍未完成该来源的准入成功验证；不以 manual 成功替代论文 AT01 或官方源成功。

本阶段的 HTTP 观察 helper 继续委托实际 Transport，附加保存实际状态 / 请求定位和传输错误的本机 JSON；不复制认证头或 Cookie、不替换响应。三次来源前置失败和所有首 RED 均保留；定位到官方 Atom 429 后没有继续重复请求、修改 DNS / SSL 或使用其他网络绕行方式。

本测试要求显式绝对路径 `ASTROCYTE_TEST_VIDEO_EXPORT_A` 与 `ASTROCYTE_TEST_VIDEO_EXPORT_B`，没有个人路径源码默认值、fallback 或 skip。此次 A 路径为 `C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/.verification/video-source-third/video-original-text.txt`，B 为 `C:/Users/DW/orca/workspaces/Astrocyte/product-video-20261014/.verification/video-timed-first/video-original-text.txt`，均按模块与 V `msg_43e318f2b008` 的精确只读授权消费。production `summarize_markdown` loader 保留已有真实正文；该范围没有宣称 Markdown 自身证明 952 个时间片段，也没有调用新提取 / ASR / 模型 / controller。

经模块允许，普通 `--no-ff` 消费 B 前置 `3384a89aac3be6d5670d756eb2d44263784d0cb8` 的共同 owned execution 接口。后续消费的 `52bd568811cbcdf59540ba18eee5a37c9d32002b` 已包含 selected-text 的 ProjectID / JobID / OperationID / SessionID / NativeID、原 controller `thread/read` 指针、Send 阶段和原件保留代码；本轨只读确认代码存在，尚未运行或验收这条映射。T 没有手工修改生产输入。

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
| 显式两实际视频导出及现系统获准匿名 loopback 代理后，`go test -mod=readonly ./internal/attention/app -count=1` | RED，23.041s；唯一失败为原 Creator 上传目录 3 项未达到 100 条门槛，其他当前测试通过；仍含待整改的旧替代模型，未作为真实模型验收 | `.verification/app-default-stage6.log` |
| `go test -mod=readonly ./internal/attention/app ./tests/s1 -list '^Test'` 与 `go build -mod=readonly -o .verification/astrocyte-stage6.exe ./cmd/server` | EXIT0；当前两包编译清单及实际生产服务构建，未运行 S1 模型路径 | `.verification/coverage-stage6.log`、`.verification/build-stage6.log` |
| 自动权限迁移首次编译 | RED：新增用例拆分后的变量声明未完成；原件保留 | `.verification/automatic-permissions-production-first.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestAutomatic(PermissionsAndUnconfiguredProcessorUseActualSource\|PriorScopeCannotExpandSelectedSources)$' -count=1`，沿用获准现系统匿名 loopback 代理 | EXIT0，19.707s；实际官方固定 arXiv、人工附加上下文、真实 SQLite 和对象，五项原拒绝 / 未配置断言组保留，零模型 / controller | `.verification/automatic-permissions-production-second.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestActualPublicMetadataRequiresRecommendationBeforeBodySelection$' -count=1` | EXIT0，3.329s；实际公开 metadata、推荐前选择拒绝、未配置推荐失败与 Agent 拒绝，正文读操作及对象发布零 | `.verification/tracking-selection-gate-production-first.log` |
| 选定项目实际 HTTP 权限首次验证 | RED，9.811s；作业种类查询不符合真实表结构，前置权限断言已执行，首记录保留 | `.verification/s1-selected-permissions-production-first.log` |
| `go test -mod=readonly ./tests/s1 -run '^TestSelectedProjectMembershipAndRevocationBeforeNativeStart$' -count=1` | EXIT0，9.087s；实际服务构建 / HTTP / session / SQLite / ObjectStore、固定引用和许可撤销，原生配置未观察保持明确失败，无 probe 或 controller 启动 | `.verification/s1-selected-permissions-production-second.log` |
| `go test -mod=readonly ./internal/attention/app ./tests/s1 -list '^Test'` | EXIT0；第七 / 八阶段当前两包编译清单，未运行模型路径 | `.verification/coverage-stage7.log`、`.verification/coverage-stage8.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestActualFixedPaperReadyCandidateDoesNotAdmitOrExecute$' -count=1`，沿用获准现系统匿名 loopback 代理 | EXIT0，5.246s；真实官方固定论文 / 人工待查问题 / Ready revision1 / 真实创建事件，准入及原生会话零；仅准入前置 | `.verification/admission-ready-prerequisite-first.log` |
| 准入生命周期首次编译 | RED：测试领域引用字段与实际 InputRef.ID 不一致；首记录保留 | `.verification/admission-service-first.log` |
| `go test -mod=readonly ./internal/attention/app -run '^TestActualAdmissionServiceAndDraftLifecycle$' -count=1 -v`，沿用获准现系统匿名 loopback 代理 | RED，8.965s；四项拒绝 / 回滚子例 PASS，实际人工 draft consumer 阻断；事务没有留下部分准入、草稿、事件或回执，后续成功路径 NOT_RUN | `.verification/admission-service-second.log`、`.verification/attention-3161444415/state.sqlite` |
| 同准入定向命令，补充实际错误本机保留 | RED，12.062s；同四项子例 PASS，实际 consumer authority guard 再次阻断，派生 guard 匹配为 true；只写本机原件及指针，不打印原错误 | `.verification/admission-service-third.log`、`.verification/attention-2261260671/admission-consumer-error.json` |
| 恢复基线 `5b45b0c`，`go test -mod=readonly ./internal/attention/app -run '^TestActualAdmissionServiceAndDraftLifecycle$' -count=1 -v` | RED：官方来源前置 35.76s 失败，未进入准入 consumer；SQLite 为 known `provider_unavailable` | `.verification/admission-resume-first.log`、`.verification/attention-1455398200/state.sqlite` |
| 同官方命令，显式既有获准匿名 loopback 代理 | RED：官方来源前置 24.00s 失败，未进入准入 consumer；没有绕过 DNS / SSL | `.verification/admission-resume-existing-proxy.log`、`.verification/attention-1115979219/state.sqlite` |
| 同官方命令，保存实际传输证据 | RED，6.641s；官方 Atom 实际 HTTP429，尚未请求 PDF 或进入准入 consumer | `.verification/admission-resume-evidence-first.log`、`.verification/public-http-3065916626/response-1.json`、`.verification/attention-89735024/state.sqlite` |
| 固定 Git/manual 完整准入场景首次运行 | RED，4.558s；四项拒绝 / 回滚及人工准入成功，原回执比较未通过；原库和日志保留 | `.verification/admission-public-git-first.log`、`.verification/attention-1157505328/state.sqlite` |
| 固定 Git/manual 完整准入场景，完整 JSON 回执比较 | EXIT0，2.548s；全部成功 / 拒绝 / 撤销 / 两次 SQLite 重开断言通过；原始实际输入为基线 `5b45b0c` 的 README | `.verification/admission-public-git-second.log`、`.verification/attention-4008557781/state.sqlite` |
| `go test -mod=readonly ./internal/attention/app -run '^TestActualPublicGitAdmissionServiceAndDraftLifecycle$' -count=1 -v`，两个来源入口分开后 | EXIT0；仅独立固定 Git/manual 完整场景，官方完整场景及断言保留 | `.verification/admission-public-git-separated.log` |
| 下列十二项适用回归命令首次运行 | RED，36.770s；8 项真实导入作业未在原 3 秒观测期内全部完成，原库保留 5 succeeded / 2 running / 1 queued | `.verification/admission-stage11-appropriate-regression.log`、`.verification/attention-2255690632/state.sqlite` |
| 同十二项命令，保持原断言和全部观测期限 | EXIT0，12 项 PASS，3.992s；没有修改并发数、作业数或期限 | `.verification/admission-stage11-appropriate-regression-second.log` |
| 准确 SOURCE `41f8823`，同十二项命令 | EXIT0，12 项 PASS，3.595s；固定 Git 原始输入为该 SOURCE 的实际 README | `.verification/admission-stage11-source41f8823.log` |
| `go build -mod=readonly -o D:/Astrocyte-agent-assets/20261015/product/t/astrocyte-stage11.exe ./cmd/server` | EXIT0；生产服务构建，未执行原生模型或完整产品链 | `.verification/build-stage11.log` |

第十一阶段十二项适用回归的实际命令：

```powershell
go test -mod=readonly ./internal/attention/app -run '^(TestActualPublicGitAdmissionServiceAndDraftLifecycle|TestImportReceiptsContentVersionsAndTransactionIO|TestHumanQueriesMachineReadsPinAndValueAreIndependent|TestConcurrentImportCommandsShareDurableJob|TestOutboxFailureRollsBackAggregateAndReceipt|TestFeedbackCASHistoryAndUnknownDimensions|TestCandidateReadyRequiresRelevantAssociationOrOpenQuestion|TestSpaceAndClassificationFailureRollbackSignals|TestProfileCASHistoryReceiptAndRollback|TestRunningCancelStopsOwnedContextAndDoesNotCommitMaterial|TestCancellationReceiptCannotCancelLaterRetry|TestWorkerPoolFiniteConcurrencyAndGracefulStop)$' -count=1 -v
```

## 仍在执行

剩余 App 共享 MemoryRepo / MemoryObjects、替代模型、provider 目录及 S1 unknown 生命周期继续整改；本轨未完成，未发送 worker_done。已有成功和失败断言、未知状态继续保留，不新增 skip、不删除断言、不替换为伪外部成功。

生产 `agents/text.go` 的 selected-text 目录已通过 B 前置 SOURCE 改为 registry owned root。2026-10-15 本恢复阶段的 Go 命令仅为自身进程设置 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/product-t`、`GOTMPDIR / TEMP / TMP=D:/Astrocyte-agent-assets/20261015/go-temp/product-t`；日志、来源和数据库仍在本树 `.verification`。模块 `msg_cb2c75691b3a` 的后续要求在当前检查结束后生效：新 Go 命令仅进程改用已预热共享 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`，GOTMPDIR 继续本轨独占 D 目录，不复制或清理缓存。S1 子服务仍显式使用其 owned evidence directory 内的 native 临时目录。

S1 首次运行在第一次导入请求的身份构造前 RED，8.36s，没有原生 probe 或模型请求；原件 `.verification/s1-native-first.log` 与 `.verification/s1-native-1981875207/server.log` 保留。模型前置问题 `msg_18fa60d90459` 的回复与后续 SOURCE 分别记录；当前代码已有上述 selected-text 映射，真实运行核查按模块延期决定保持 NOT_RUN。T 不以 `ExternalStarted` 或工具调用零单独证明实际 Send / 工具关闭，不扫描原生 home、不重发 B 的旧成功请求。

收藏夹真实响应原件位于 `.verification/public-http-1166636830/response-1.bin` 至 `response-9.bin`。只读派生为资源页 `20 / 20 / 19 / 0` 项、均 `code=0`、`has_more=true`；目录元数据为 459 项。生产 `PublicListing.ReadPage` 按 returned observed 推进窗口，而 provider 页位置按 requested page size 计算，第三页不足 20 后下一请求落到空槽。这一具体问题已直接 Handoff B `ctx_7217777c7bf5`；T 没有修改 provider、补造行数 / metadata、绕过网络权限或降低 cap。creator3 的有限列表观察、收藏459的目录前置、100 截断失败和最后一页缓存成功分别保留。

根 `msg_8d2dcf0e1841` 批准另选足量、匿名公开可访问 Creator 作为工程 QA 来源；正式用户案例与科研成绩不随之改变。T 对三个现有公开发布者候选 `49683148`、`1408168321`、`946974` 使用原生产 `bilibili-listing.py` 与已安装固定 yt-dlp 查询首 metadata 页，实际均 HTTP412，stdout 原件为零，真实发布总数仍 UNKNOWN。原错误仅在 `.verification/creator-<UID>-first.error.log`，不进 stdout / Git；没有登录、Cookie、凭据、源外下载或生成数量。三个同类拒绝已向模块及 B 报告，不继续批量盲扫。

## 后续实际验收条件

模块最新 `msg_0efc7351d1c1` 与 `msg_fc21587f0442` 已替代按日期暂停的指令，根已授权当前语义 / CLI 与 SPEC S0–S6 实施。本轨继续原独占测试整改；实际 selected-text 模型恢复条件为精确 R/H 新语义通过且根 `working/currentowned` 具体规则答复明确，不能仅凭日期或一般实施授权启动 paid controller。目前实际模型、原生 controller 启动均为零；包含该隐式路径的 `go test -mod=readonly ./internal/attention/app ./tests/s1 -count=1` 保持 NOT_RUN，不新增 skip、不删除 case 或隐藏默认门槛。

继续验收需要按上述具体门槛核查已消费 text / bridge 代码的原 operation → 原 registry / session / 正选原档 / 实际 argv / Send 受理 / Stop 证据，以及可访问的足量公开 Creator。根已裁决 100 预算为实际 metadata，允许跨缺失槽到第 100 条，真实 Observed19 保留，100 即停止、超时及无进展失败继续保持。准确 cursor 已消费，收藏100 / 110缓存 / partial只读半段均有本轨独立实际结果；Creator 原3与三个候选412仍保持原范围失败，未由收藏结果替代。

T 已将实际依赖精确交付 B `ctx_7217777c7bf5`：`msg_602d3cc2641f` 说明 production Import Reader 尚无已找到的真实 DeliveryUnknown 路径，现 OpenCLI unknown 只在选中目录桥接且无本轨用户来源授权；`msg_8a48fc5def0a` 说明 S1 须在同一 Job / operation 的实际 `send_accepted` / TurnID 后取消，当前新 UUID controller 的本机 source 指针消费方式仍需准确衔接。现 `OwnedRelayExecution.EffectiveToolSet` 明确保留 unknown，实际 argv 与 LaunchPolicy 不能单独证明有效工具关闭。T 不扫描 home / 私有历史，不生成新索引或 manifest，不替代 B / R / H 写入域；原 Import / listing unknown 断言和模型成功断言均保留，依赖答复与实际原生门槛尚未满足。

模块新增资料准入测试支持由本 owner 承担。B 已取得本轨实际 helper 签名和来源环境 `msg_de49a3b81a9a`；其准确生产 SOURCE 已按上述普通消费并独立执行。真实人工 consumer 首失败已精确 Handoff B `msg_98795e92f7b5`，原 S2 准确修复 SOURCE 已普通消费；当前固定 Git/manual 的完整成功、幂等、撤销、来源生命周期和两次数据库重开均已实际通过。官方论文完整生命周期继续受实际 Atom429 阻断，原断言与首失败保留；没有补造 PolicyVersion / 判断授权或事件，没有修改 S2 / B 的生产域。判断委托的正向路径和模型仍须实际获准入口，GoalRef 不自行补充。当前真实 selected-text / Import unknown 生产入口的依赖已重新交 B `msg_b818207d89f7` 和模块 `msg_75ba5cdb74f7`；仅依赖待交付时，本 Task 保持 check --wait 和检查点心跳，继续准确 SOURCE 的组合与返修支持，不按日期暂停或结算任务。

模块 `msg_1c993379cac1` 通知实际 C 盘容量限制，本轨已确认保留所有首失败与当前证据，不重复构建 / 下载 / ASR、不清理其他 owner 或私有资料；必要新字段和核心检查继续执行。当前未发生本轨 NoSpace，若实际发生则按准确 owned 路径和所需容量向模块与 M-release 交付，不以 skip 或删除断言处理资源不足。

V 的两真实文本导出已支持本轨版本复用验证，完整时间 / 默认提取验收仍归 V 的真实生产路径，不借本轨 Markdown 版本复用结果填其全模块通过。原生只提供固定公开输入，工具配置由真实运行证据核对；未确定部分保持 UNKNOWN。完整 App 自动处理 / tracking 及 shared helper 的替代实现整改仍未完成，本 owner 持续负责同一写入范围、功能验证与最终组合支持。T 没有发送 worker_done 或发布产品。

完整成功、权限拒绝、未知交付、重启和整改后相关全包测试尚未完成；本文件所列定向结果不构成全部产品验收。剩余旧共享 helper / 自动模型 / tracking 外部替代测试仍需整改，目录容量不足及模型证据前置分别记录；未用旧替代测试的 PASS 填成本轨完整通过。
