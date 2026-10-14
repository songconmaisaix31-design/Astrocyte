# Attention 测试整改记录

本轨工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/product-attention-tests-20261014`，分支为 `songconmaisaix31-design/product-attention-tests-20261014`，接管基线为 `b470ecd736af261a10e4c4822bdcd3b02d0da454`。Orca 终端 `term_c1d90d02-51be-4443-9afe-ddce0615acb4`，Task `task_5f0a79b6a17a`，Dispatch `ctx_8674c456f49f`；实际终端显示 Codex / GPT-6.1-Sol high fast。

本轨唯一写入范围为 `internal/attention/app/*_test.go`、`tests/s1/automatic_test.go` 和本文件；生产、依赖、native/relay 与其他测试归原人员。原始错误、实际来源、模型响应和会话证据只保留本树已忽略 `.verification`，不纳入 Git 或产品接口。

## 已提交源码阶段

SOURCE `a1e1195464c2f989083780a7b2bc0ade0cc07ab6` 已普通提交并 push：候选门槛、人工使用动作、固定版本项目引用、分类和空间、profile CAS 使用外部测试包与生产 SQLite / Reader / ObjectStore。新增两个本目录专用测试文件，生产代码没有修改。纯排名参数测试继续直接检查生产纯函数。

第二阶段 SOURCE `03af9854dac57f69fc206d4d647ad0c518b98b20` 已普通提交并 push。经模块明确授权，普通 `--no-ff` 消费 B 的前置 SOURCE `01753cf4fb455f35aa4953bf3eaa9a6388c4be5a`，没有手工修改 B 的生产文件。T 本阶段实际修改 9 个 App 测试文件，新增专用 `attention_service_test.go`、`imports_service_test.go`、`jobs_service_test.go` 和 `tracking_production_test.go`；新增测试均在约定的 `*_test.go` 写入范围内。

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

## 仍在执行

剩余 App 共享 MemoryRepo / MemoryObjects、替代模型、provider 目录及 S1 unknown 生命周期继续整改；本轨未完成，未发送 worker_done。已有成功和失败断言、未知状态继续保留，不新增 skip、不删除断言、不替换为伪外部成功。

生产 `agents/text.go` 的 selected-text 目录已通过 B 前置 SOURCE 改为 registry owned root。本轨后续 Go 命令的 `GOTMPDIR / TEMP / TMP` 显式指向 `.verification/go-temp`；S1 子服务另显式使用其 owned evidence directory 内的 native 临时目录。

S1 `automatic_test.go` 当前是尚未提交的测试 WIP：使用实际编译的生产服务、真实 HTTP/session bootstrap、官方 Atom/PDF、项目许可、取消请求及同库重启。首次运行在第一次导入请求的身份构造前 RED，8.36s，没有原生 probe 或模型请求；原件 `.verification/s1-native-first.log` 与 `.verification/s1-native-1981875207/server.log` 保留。请求身份已修正，尚未继续运行到模型。模块已要求真实模型前先由 B/R 保留原 registry 中精确 session、自产源指针和实际运行配置，T 等待该生产口，不以 `ExternalStarted` 或工具调用零单独证明实际 Send / 工具关闭，不扫描原生 home、不重发 B 的旧成功请求。

完整成功、权限拒绝、未知交付、重启和整改后相关全包测试尚未完成；本文件所列定向结果不构成全部产品验收。剩余旧共享 helper / 自动模型 / tracking 外部替代测试仍需整改，目录容量不足及模型证据前置分别记录；未用旧替代测试的 PASS 填成本轨完整通过。
