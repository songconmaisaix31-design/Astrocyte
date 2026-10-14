# Attention 测试整改记录

本轨工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/product-attention-tests-20261014`，分支为 `songconmaisaix31-design/product-attention-tests-20261014`，接管基线为 `b470ecd736af261a10e4c4822bdcd3b02d0da454`。Orca 终端 `term_c1d90d02-51be-4443-9afe-ddce0615acb4`，Task `task_5f0a79b6a17a`，Dispatch `ctx_8674c456f49f`；实际终端显示 Codex / GPT-6.1-Sol high fast。

本轨唯一写入范围为 `internal/attention/app/*_test.go`、`tests/s1/automatic_test.go` 和本文件；生产、依赖、native/relay 与其他测试归原人员。原始错误、实际来源、模型响应和会话证据只保留本树已忽略 `.verification`，不纳入 Git 或产品接口。

## 已提交源码阶段

SOURCE `a1e1195464c2f989083780a7b2bc0ade0cc07ab6` 已普通提交并 push：候选门槛、人工使用动作、固定版本项目引用、分类和空间、profile CAS 使用外部测试包与生产 SQLite / Reader / ObjectStore。新增两个本目录专用测试文件，生产代码没有修改。纯排名参数测试继续直接检查生产纯函数。

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

## 仍在执行

剩余 App 共享 MemoryRepo / MemoryObjects、替代模型、provider 目录及 S1 unknown 生命周期继续整改；本轨未完成，未发送 worker_done。已有成功和失败断言、未知状态继续保留，不新增 skip、不删除断言、不替换为伪外部成功。

生产 `agents/text.go` 的 selected-text 临时目录需要原 owner B 改为 owned root；M-product 已向 B 协调。真实 Native 启动、配置与取消入口由 R 交接，采用本轨独立固定公开输入与新 session，不能重复 B 已成功 infer 或使用 H/F 活跃窗口。完整成功、权限拒绝、未知交付、重启和整改后相关全包测试尚未完成；本文件所列定向结果不构成全部产品验收。
