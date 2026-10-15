# S4 仓库上下文

本轨基线为 `da79452d1735f8fceca103aaf9d10bb53182f1f5`，分支 `songconmaisaix31-design/works-repository-context-s4-20261015`。客户端为 Codex API；本会话没有可独立核实的精确模型标识。按模块确认，仅消费 S2 的 `4eec31ddfde39fa3f548691ec6e7f2f34470351a`、`9c5b1619710b576f291d42a5aa641b28639d5a57`、`5d0ba735533f68259b91128dc404c8e4c2f2a535`，以及 B 的单表迁移 `98bb9236c68f45ca4ab4114d46eafba5c90cdedf`，均保留 `cherry-pick -x` 来源。

## 设计与接入

`repositorycontext.New(projectReader, indexStore, Options)` 组装真实 Git、官方 CodeGraph 和 AOCI 观察适配。`sqlite.DB` 实现两个存储端口；H 负责服务与 HTTP 薄组装，S4 不修改 server、OpenAPI、依赖或迁移。S2 使用原 `ProjectSnapshotProvider`，没有复制其 Domain 类型。

Git 从注册根核对 worktree 根、common-dir、git-dir 与真实 HEAD；授权范围的 stage 清单和实际文件内容分别形成 index/worktree digest。文件清单包括未忽略的未跟踪文件和已删除记录。项目环境、仓库、worktree、项目、权限 revision 和准确相对 roots 共同决定版本。根目录只在内部观察和 S2 本地快照中使用，仓库查询投影不序列化根目录。

每次读取重新加载真实项目与 Agent grant，检查当前撤销、过期、`read_context`、目录开关和允许子目录。用途、文件数量、读取字节数、结果数量和展开深度必须显式提供，受可信启动配置上限约束。密钥目录、环境文件、依赖、构建目录和符号链接不进入源码投影。查询前后重新捕获真实版本；变化返回 `index_stale`，不发布旧关系。

`CaptureProjectSnapshot` 保存不可变的 Git 观察和真实授权范围；`ValidateProjectSnapshot` 重新核对当前 Git、项目许可和保存的版本，不信任客户端的 Root 或 digest。保存使用 `repository_context_indexes` 的独立 `git-snapshot:UUID` 后端键。SQLite adapter 在已有 Workspace 事务 context 内调用 `WorkspaceTransaction(ctx, db)` 复用其连接，避免单连接事务的二次借用死锁。

CodeGraph 使用官方 npm `@colbymchenry/codegraph` **1.0.1**，MIT，源码标签 `v1.0.1` 对应 `a1489f77a6d69538bfe29020b8366ee034f90860`。可信配置提供 Node、SDK 和本轨 `codegraph.mjs` 的实际路径。仅子进程设置 `CODEGRAPH_TELEMETRY=0`；不运行安装器、不修改全局 MCP 或 Agent 配置。

显式 `index/sync/rebuild` 只把批准范围内的实际源码复制到本平台拥有的独立 generation 目录。版本仍指原 Git repo/worktree；`source_span` 对应原相对文件和行号。sync 复制上次已关闭、已 checkpoint 的真实 SQLite 图，再由官方 SDK 增量识别新增、修改和删除；失败不替换当前已发布 generation，成功后用真实 SQLite generation CAS 发布。受损投影经显式 rebuild 从原源码重建。

CodeGraph 1.0.1 的 `CodeGraph.open` 实现没有使用其声明的 `readOnly` 参数。查询改用官方导出的 `QueryBuilder` 和 Node `DatabaseSync(..., {readOnly:true})`，不触发迁移、自动同步或索引写入。符号搜索使用其真实 FTS/符号检索；关系按官方保存的调用、导入、继承和引用边进行有界展开，标记 `codegraph_static_heuristic` 和提取器版本。动态调用与测试覆盖始终报告缺口。FTS 不标为 embedding。

AOCI 固定 **v0.1.0-rc20**，标签对象 `7c59b0718780f5fccb65d9f0a3445b0b98fea177` 的实际源码提交为 `45990a1dcfc4415efdc6fbfa13bc159a415f127c`。已完整读取官方 README、安装文档、所选版本 LICENSE 与实际 JSON/CLI 源码，并在本机 ignored 工具目录构建；实际二进制报告 `v0.1.0-rc20`、上述 commit、build date unknown。许可为 FSL-1.1-MIT，本轮仅本机适配验证，最终包分发尚未获决定，不携带该 binary。

`RefreshRepositoryMap` 是显式人类动作，要求完整项目目录授权。它启动官方 AOCI 仅 loopback、`--discover=false`、随机端口的临时只读服务；从成熟 `/api/state`、`/api/entries` JSON 读取实际治理状态与模型作者维护的 F/R/A/S，并只保存安全相对引用和语义字段。缺少合法完整索引时返回 `evidence_missing/index_stale`。`ReadRepositoryMap` 只读已保存的安全投影和当前 Git，不启动 AOCI，不运行会写 ledger 的 verify/check，不解析 FRAS 文件、不从文件名或 AST 生成语义。

SemanticIndex 的具体模型、运行时和向量库仍待用户决定。没有冻结默认模型或接口，没有零向量、虚假模型成功或替代 embedding。当前 CodeSearch 明确报告 `semantic_embedding_not_available`。

## 真实验证记录

验证对象为 Agent 创建的[公开 Unicode 统计项目](https://github.com/songconmaisaix31-design/astrocyte-s4-public-context-20261015)，原始实际提交 `4317d61`。只在自有 `dist/s4` 项目和 `.verification/s4` 投影、SQLite 与日志中运行，不读取用户作品、凭据或原生历史，不启动新模型会话。公开源码实现 Unicode 字符、单词、行与段落统计，修改前后实际运行 `go test ./...`、`go build ./...`。

| 验证 | 实际结果 |
|---|---|
| 同 HEAD 两 worktree，真实 repo/worktree 身份 | PASS：共享 RepoID、不同 WorktreeID |
| S2 快照、篡改拒绝、同一真实 SQLite 事务 Validate | PASS |
| 两个 worktree 分别建立官方 CodeGraph 索引 | PASS |
| Count 符号、真实调用关系、相对 span 和来源标签 | PASS |
| A 实际新增段落功能，B 保持不变；旧索引 stale；真实 sync | PASS |
| 删除独立段落文件并把功能保留在原文件；真实 sync | PASS；首断言错误期待大小写不敏感查询为空，原 RED 保留；改核被删除的准确 source path 与保留功能后 PASS |
| 关闭后重新打开真实 SQLite/adapter | PASS |
| 实际破坏自有图 SQLite；拒绝查询；显式 rebuild | PASS |
| 跨项目、逃逸根、私有路径、空用途、读取预算拒绝 | PASS |
| SQLite Agent grant 撤销、目录权限撤销、revision 变化后重建 | PASS |
| AOCI 源码构建与实际版本 | PASS；模型作者完整地图正向链路尚未执行 |
| SemanticIndex、HTTP/前端/最终包组合 | NOT_RUN |

适用命令：`go test ./internal/workspace/... ./internal/adapters/repositorycontext ./tests/s4/...`、`go vet ./internal/adapters/repositorycontext ./tests/s4/...`、`go build ./...` 已通过。真实操作入口为 `tests/s4/verify.ps1`，各阶段独立进程；本机结果位于 `.verification/s4/actual-first/` 和 `dist/s4/initial-first.log`。`deleted-failure.txt` 保存首 RED，后续 `deleted-check.log` 为准确文件删除验收，不覆盖原件。

首次额外 `go test ./internal/adapters/sqlite` 保留 FAIL：九项既有 Workbench 实际来源测试未配置 `ASTROCYTE_TEST_COGNITION_LOCALAPPDATA` / `ASTROCYTE_TEST_COGNITION_NATIVE_LOCALAPPDATA`。没有新增 skip、修改其断言或声明全包通过。完整默认 check/build、HTTP、前端和包级验证由最终精确组合接续执行。
