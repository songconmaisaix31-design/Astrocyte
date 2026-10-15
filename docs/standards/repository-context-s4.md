# S4 仓库上下文

本轨基线为 `da79452d1735f8fceca103aaf9d10bb53182f1f5`，分支 `songconmaisaix31-design/works-repository-context-s4-20261015`。客户端为 Orca Codex，Orca 当前 provider 投影记录模型 `gpt-6.1-sol`。按模块确认，仅消费 S2 的 `4eec31ddfde39fa3f548691ec6e7f2f34470351a`、`9c5b1619710b576f291d42a5aa641b28639d5a57`、`5d0ba735533f68259b91128dc404c8e4c2f2a535`，以及 B 的单表迁移 `98bb9236c68f45ca4ab4114d46eafba5c90cdedf`，均保留 `cherry-pick -x` 来源。

## 设计与接入

`repositorycontext.New(projectReader, indexStore, Options)` 组装真实 Git、官方 CodeGraph 和 AOCI 观察适配。`sqlite.DB` 实现两个存储端口；H 负责服务与 HTTP 薄组装，S4 不修改 server、OpenAPI、依赖或迁移。S2 使用原 `ProjectSnapshotProvider`，没有复制其 Domain 类型。

Git 从注册根核对 worktree 根、common-dir、git-dir 与真实 HEAD；授权范围的 stage 清单和实际文件内容分别形成 index/worktree digest。文件清单包括未忽略的未跟踪文件和已删除记录。项目环境、仓库、worktree、项目、权限 revision 和准确相对 roots 共同决定版本。根目录只在内部观察和 S2 本地快照中使用，仓库查询投影不序列化根目录。

每次读取重新加载真实项目与 Agent grant，检查当前撤销、过期、`read_context`、目录开关和允许子目录。用途、文件数量、读取字节数、结果数量和展开深度必须显式提供，受可信启动配置上限约束。密钥目录、环境文件、依赖、构建目录和符号链接不进入源码投影。查询前后重新捕获真实版本；变化返回 `index_stale`，不发布旧关系。

`CaptureProjectSnapshot` 保存不可变的 Git 观察和真实授权范围；`ValidateProjectSnapshot` 重新核对当前 Git、项目许可和保存的版本，不信任客户端的 Root 或 digest。保存使用 `repository_context_indexes` 的独立 `git-snapshot:UUID` 后端键。SQLite adapter 在已有 Workspace 事务 context 内调用 `WorkspaceTransaction(ctx, db)` 复用其连接，避免单连接事务的二次借用死锁。

CodeGraph 使用官方 npm `@colbymchenry/codegraph` **1.0.1**，MIT，源码标签 `v1.0.1` 对应 `a1489f77a6d69538bfe29020b8366ee034f90860`。可信配置提供 Node、SDK 和本轨 `codegraph.mjs` 的实际路径。仅子进程设置 `CODEGRAPH_TELEMETRY=0`、`CODEGRAPH_NO_DOWNLOAD=1`；不运行安装器、不修改全局 MCP 或 Agent 配置。失败原始诊断只保存在本平台 ignored 投影目录的 `astrocyte-error-*.log`，HTTP 错误保持安全固定原因。

显式 `index/sync/rebuild` 只把批准范围内的实际源码复制到本平台拥有的独立 generation 目录。版本仍指原 Git repo/worktree；`source_span` 对应原相对文件和行号。sync 复制上次已关闭、已 checkpoint 的真实 SQLite 图，再由官方 SDK 增量识别新增、修改和删除；失败不替换当前已发布 generation，成功后用真实 SQLite generation CAS 发布。受损投影经显式 rebuild 从原源码重建。

CodeGraph 1.0.1 的 `CodeGraph.open` 实现没有使用其声明的 `readOnly` 参数。查询使用官方导出的 `QueryBuilder` 和 Node `DatabaseSync`，以 `mode=ro&immutable=1` 的 SQLite URI 打开已发布的独立 generation，先拒绝非空 WAL，不触发迁移、自动同步、索引或 shm 写入。URI 由标准 `pathToFileURL` 构造；[Node 24.16.0 源码](https://github.com/nodejs/node/blob/v24.16.0/src/node_sqlite.cc) 显式启用 `SQLITE_OPEN_URI`，[SQLite 官方文档](https://www.sqlite.org/uri.html) 规定 immutable 的适用条件。每个 generation 在维护完成、SDK 关闭后发布，后续维护写入新目录。符号搜索使用其真实 FTS/符号检索；关系按官方保存的调用、导入、继承和引用边进行有界展开，标记 `codegraph_static_heuristic` 和提取器版本。动态调用与测试覆盖始终报告缺口。FTS 不标为 embedding。

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
| 失效投影清理、同 worktree 的其他项目保护、分批删除与重启 | PASS；内部 pass 与专用生命周期，B/H 生产接线待交 |
| AOCI 源码构建、实际版本、官方 MCP 九工具 | PASS |
| 模型作者六对象 FRAS → 官方整批 apply → Verify/Check/Guide | PASS：applied=6、aligned=true、complete=true、next_action=none |
| 保存 AOCI 地图、无 AOCI 配置的 service 读取地图 | PASS：六条语义和相对引用；GET 没有 AOCI 进程或 ledger 写入，`.aoci` 文件数量、大小、mtime 相同 |
| 无初始化的 B 地图 | PASS：真实 `evidence_missing`，没有伪造语义 |
| 真实 Git 暂存区变化 | PASS：A 的地图和图均 `index_stale`，B 查询保持当前；显式 sync/refresh 恢复 |
| 查询结果数量预算 | PASS：图与六条地图超限均拒绝，没有静默截断 |
| H `31256d5` 实际完整配置 HTTP / SQLite / 官方源码 SDK 组合 | PASS；首轮搜索 timeout 原件保留，原因 UNKNOWN |
| SemanticIndex、生产 main 配置接线、前端/最终包组合 | NOT_RUN |

适用命令：`go test ./internal/workspace/... ./internal/adapters/repositorycontext ./tests/s4/...`、`go vet ./internal/adapters/repositorycontext ./tests/s4/...`、`go build ./...` 已通过。真实操作入口为 `tests/s4/verify.ps1`，各阶段独立进程；本机结果位于 `.verification/s4/actual-first/` 和 `dist/s4/initial-first.log`。`deleted-failure.txt` 保存首 RED，后续 `deleted-check.log` 为准确文件删除验收，不覆盖原件。

首次额外 `go test ./internal/adapters/sqlite` 保留 FAIL：九项既有 Workbench 实际来源测试未配置 `ASTROCYTE_TEST_COGNITION_LOCALAPPDATA` / `ASTROCYTE_TEST_COGNITION_NATIVE_LOCALAPPDATA`。没有新增 skip、修改其断言或声明全包通过。完整默认 check/build、HTTP、前端和包级验证由最终精确组合接续执行。

## 首个可消费交付

SOURCE `9eff2edb34058826d302d3a465b7fe7e74797be5` 已普通 push。消费方可独立使用 S2 的真实 `ProjectSnapshotProvider`：`repositorycontext.New(db, db, Options)` 后把返回的 service 交给原 S2 构造接口；没有 HTTP 自证快照，也不需要第二 SQLite 连接。查询的 `Root` 不会进入仓库上下文 JSON。H 仍须对 S2 现有快照字段执行其既定安全 DTO 审核。

本次源码适用 Go 测试、vet、完整 Go build 与 `git diff --check` 均 PASS；最后一次独立进程 restart 再次核两 worktree 的真实 Count 符号和调用关系 PASS。AOCI 缺口、SemanticIndex 未决定及完整 HTTP/前端/包验证保持上表状态。本阶段继续同 Agent、工作树和分支开发，不作为 S4 或发布包最终结算。

## AOCI 真实作者与边界验收

两个公开 worktree 的完整 HEAD 均为 `4317d610cbe80aa6c4e22ea4da5e8dde013de4ed`。AOCI 只在 A 内执行官方 init/scan，不写平台 AGENTS、全局 MCP 或已有用户配置。源码和工具存于 `dist/s4`。当前模型完整读取六个候选对象、关联的统计测试与正式 Meta，分别根据实际职责和约束编写条目，通过官方 `aoci_update_entry` 保存 source/candidate/code batch 绑定；没有从名称或 AST 模板生成语义。六个管理对象是 `.gitattributes`、`AGENTS.md`、`README.md`、`cmd/textstats/main.go`、`go.mod`、`textstats/stats.go`；测试属于 AOCI 既有 observe 范围，平台没有降低其治理要求。

`dist/s4/aoci-authored-request.json`、`aoci-authored-first-result.json`、`aoci-verify-first.json`、`aoci-check-first.json`、`aoci-guide-final.json` 是本机原件，全部 ignored。机器治理通过只证明条目格式、来源绑定和对齐条件，语义正确性仍须依据源码和实际行为判断。AOCI build date unknown、发布签名验证 NOT_RUN、FSL 包分发未决定均保留。

首地图查询 FAIL 是 UI 协议对象类型 `file` 的接入判断，原件 `.verification/s4/actual-first/map-failure.txt` 保留；后续按官方 `EntryView` 执行并通过。首本机额外 Go build 曾把 ignored 临时 MCP 客户端文件纳入根 `./...`，缺少该临时工具的依赖 checksum；其记录见 `.verification/s4/actual-first/checkpoint-errors.md`。临时工具在 Go 明确忽略的自有目录中运行，平台正式 MCP 验证入口为 `tests/s4/aoci`，依赖按 B 的准确 SOURCE `6b78f641ecb2a3ae53b11df9aeeb837e6c72c646` 消费，后续 `go build -mod=readonly ./...` PASS。没有修改 AOCI 源码或其 go.mod。

`tests/s4/aoci` 使用官方 Go MCP SDK，拒绝其他 remote，输出只显示工具名、错误状态和数量；原响应以独占文件写入指定 ignored 目录。新验证日志和错误文件均带本次实际时间，保存各轮结果。`tests/s4/verify.ps1 -Phase map-readonly/map-missing/result-bounds/expect-stale` 对应上述实际边界；没有请求新的 native 模型窗口、浏览器或发布操作。

## AOCI 后继交付

SOURCE `263bc7069682ede35af60426140ee6f744b0d472` 已普通 push，父链含 B 的精确 SDK/依赖 SOURCE。该源码的 `node --check internal/adapters/repositorycontext/codegraph.mjs`、`go test -mod=readonly ./internal/workspace/... ./internal/adapters/repositorycontext ./tests/s4/...`、`go vet -mod=readonly ./internal/adapters/repositorycontext ./tests/s4/...`、`go build -mod=readonly ./...`、`git diff --check` 全部 PASS。源码提交前最后一次独立进程 restart、map-readonly、result-bounds 均 PASS，正式 MCP 验证客户端实际列出官方九工具。

模块负责人在自有新 SQLite 和投影目录独立运行 `verify -phase initial`，实际 6.181 秒 PASS：同 HEAD 的独立 worktree 身份、同事务快照校验、篡改拒绝、两棵真实官方图、Count 与调用关系及安全相对 span。执行来源包含当时已读后继修改，模块没有把该结果冒称首个 9eff 源码的精确验收；消息 `msg_98a7649970a4` 保留范围。公开源码与原验证证据没有被独立检查覆盖。

当前真实限制：SemanticIndex 的用户决定未到；HTTP/前端/最终 Windows 包组合未执行；AOCI 完整系统理解未作额外全量认知声称，FSL 分发仍未决定。公开项目目录和 SDK/工具是本机依赖，发布方须使用 B 的依赖清单和可信实际路径配置。阶段 SOURCE/REPORT 供 S2/H 消费，当前 S4 继续承担后续接入与返修。

Git 子进程移除宿主 `GIT_*` 环境变量，防止外部 `GIT_INDEX_FILE/GIT_WORK_TREE` 重定向注册仓库的实际观察。`verify -phase ambient-git` 实际把宿主 index 指向 B 的真实暂存区，A 的 IndexDigest、WorktreeDigest 和 WorktreeID 保持准确，原快照 Validate 仍通过；随后独立 restart 两图查询 PASS。仓库和 worktree 身份计算方式保持现有记录兼容。

环境约束 SOURCE `1a55e1630502642324500c3b64853af882962fa4` 已普通 push；`go vet -mod=readonly ./internal/adapters/repositorycontext ./tests/s4/...`、`git diff --check` 与上述实际两阶段验收 PASS。HTTP 处理器与正式互斥交接见下一节。

## HTTP 业务处理器

模块管理 SOURCE `95782646bb6181ba8fe97acc2a3f0d4b7b9455f9` 和 `msg_589549037e1d` 已追加本轨唯一 `internal/adapters/httpapi/repository_context*.go` 写域，产品/B 两个原 owner 已确认零 WIP 并停止写入。H 继续只写共享 main/server/session/OpenAPI 薄组装；B 继续配置、依赖与迁移。本轨导出 `RepositoryContextServices`，五个字段分别为 Map、Search、Impact、Indexes、MapCommands，类型使用既有独立 app 端口；`RegisterRepositoryContextRoutes(mux, services)` 在生产认证 router 内注册，不创建额外认证服务。

路由前缀为 `/api/v1/local-projects/{id}/repository-context`：

| 方法与后缀 | 请求 | 行为 |
|---|---|---|
| GET `/map` | purpose、重复 root、max_files/max_bytes/max_results/max_depth 均必需 | 仅读保存的安全地图和当前授权 Git |
| POST `/search` | schema_version=1、purpose、roots、budget、query | 当前授权符号/关键词检索 |
| POST `/impact` | 同上，使用 symbol_id | 当前快照的有界关系展开 |
| POST `/index` | 同上，action 为 index/sync/rebuild | 显式人类索引维护 |
| POST `/map/refresh` | schema_version=1、purpose、roots、budget | 显式人类地图刷新 |

所有 HTTP roots 必须为安全相对路径；ProjectID 只取 route，Caller 只取既有认证 context。严格 JSON 解码拒绝未知的身份、Root、HEAD、digest 和 generation。GET 拒绝未知参数和重复标量。读取仍由 Service 重新验证当前项目授权，维护双重核实 human 身份。H 的 scoped Agent 白名单只增加 map GET、search/impact POST；Agent 维护路由继续拒绝。

成功响应为 `{schema_version:1, repository_context:...}`。IndexStale、ContextStale、BudgetExhausted、VersionConflict 均使用既有 S3 的 HTTP 409，ErrorV1.code 分别保留；未知内部错误只显示固定安全原因。CodeGraph score 是普通排序得分，实际可能大于 1，不能解释为概率；向量能力仍明确缺失。处理器已通过当前真实包 build/vet 和差分检查，实际 HTTP 验证待 H 的准确生产注册及认证薄组装 SOURCE，未使用替代认证或虚假 HTTP。

另已执行 `verify -phase same-project-worktrees`：同一个真实项目 ID 在两个同 HEAD worktree 之间更换注册根，B 的图不会复用 A 的内容；返回 A 时权限 revision 变化要求重新构建，随后准确恢复自己的实现和地图。实际 PASS，原源码/HEAD/暂存区均未修改。

HTTP SOURCE `896a7b430c62f9d91aaf5434b11c7653a3d00f36` 已普通 push，准确注册函数、五个端口字段和 DTO 已分别交模块负责人、H 与 UI 负责人。`go build -mod=readonly ./internal/adapters/httpapi`、`go vet -mod=readonly ./internal/adapters/httpapi`、`git diff --check` PASS；后继真实 HTTP、认证、SQLite 与重启结果见下文生产组合记录。

模块独立审读 `msg_1d21efa0bebb` 指出 GET 查询解析应拒绝畸形编码。本轨使用成熟 `net/url.ParseQuery(r.URL.RawQuery)` 并立即拒绝解析错误，避免静默忽略畸形未知参数；HTTP 验收脚本已准备畸形 percent、semicolon、重复标量与未知参数场景，实际执行仍待生产组装，不使用自研 parser。

查询解析修复 SOURCE `293af633570d6a4f4cf4bbf9c5b651821ab977e7` 已普通 push；当前 HTTP 包 build/vet 和差分检查 PASS。真实 HTTP 验收脚本仅在准备目录中，未宣称实际运行成功；H 生产注册/认证准确 SOURCE 是其当前依赖。模块及 H 已收到准确 Services、注册函数、DTO 和权限规则，F/UI 仅按 H 唯一契约生成类型消费。

## 生产组合的准确输入

本次恢复使用同一 Task `task_bd209566be37`、Dispatch `ctx_09391d7a7e02`、工作树与分支；前次 `ctx_12ae8872df87` 已由恢复流程记录 failed / terminal_missing。Orca 当前 provider 投影记录客户端 `codex`、模型 `gpt-6.1-sol`，本次实际开始已向模块确认。新的 Go 及调用 Go 的父进程仅在自身环境中使用 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`，`GOTMPDIR`、`TEMP`、`TMP` 均使用本轨独占的 `D:/Astrocyte-agent-assets/20261015/go-temp/S4`。原始数据库、日志与验证源码继续保留。

H 按顺序精确消费以下 SOURCE；仅含报告的提交无须进入其业务组合：

| SOURCE | 修改范围 |
|---|---|
| `9eff2edb34058826d302d3a465b7fe7e74797be5` | repositorycontext 五个 Go 文件与 codegraph.mjs、sqlite/repository_context.go、workspace app/domain repository_context.go、tests/s4/verify 与 verify.ps1、本标准 |
| `263bc7069682ede35af60426140ee6f744b0d472` | aoci.go、codegraph.go、app/repository_context.go、tests/s4/aoci、verify、本标准 |
| `1a55e1630502642324500c3b64853af882962fa4` | git.go、verify、本标准 |
| `896a7b430c62f9d91aaf5434b11c7653a3d00f36` | httpapi/repository_context.go、verify、本标准 |
| `293af633570d6a4f4cf4bbf9c5b651821ab977e7` | httpapi/repository_context.go、本标准 |

必要依赖来源分别为 B 的 `98bb9236c68f45ca4ab4114d46eafba5c90cdedf`（仅 migrations/021_repository_context.sql）和 `6b78f641ecb2a3ae53b11df9aeeb837e6c72c646`（仅 dependencies.lock.json、go.mod、go.sum）；本树对应保留来源的提交为 `f25b838738d981f73380f042c2f7940dc3fd3a7e`、`b4e288503868d4e5d1034026f82025444d5b13b9`。S2 接口来源为其原始 `4eec31ddfde39fa3f548691ec6e7f2f34470351a` 与 `5d0ba735533f68259b91128dc404c8e4c2f2a535`；由消费方核实已存在的接口及自身依赖，不合入其他历史。

实际构造为 `repositorycontext.New(db, db, repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: ...}, Graph: repositorycontext.CodeGraph{Node: ..., SDK: ..., Script: ...}, AOCI: ..., CacheRoot: ..., SnapshotBudget: ...})`。Node 可执行程序、官方 SDK 目录、本轨 codegraph.mjs、AOCI 可执行程序与投影目录均由可信启动配置提供。生产服务使用同一 adapter 实例作为 S2 快照端口和五个 HTTP 能力端口。

`tests/s4/_http/main.go` 是 SOURCE `faab2daf83f6a578773b61077140aff73af9ef9b` 的实际 HTTP 验证入口，使用 H 的生产 `Services`、认证白名单和 server。`query-readonly` 阶段核对真实 Search、Impact 与保存地图查询前后所有自有图投影和 AOCI 运行文件的数量、大小及 mtime，实际结果见下文。

恢复后在 D 盘进程缓存配置下执行 `tests/s4/verify.ps1 -Phase restart` PASS。`query-readonly` 首次真实运行 FAIL，原件为 `.verification/s4/actual-first/query-readonly-failure-20261015T030418.352154300.txt` 及 `query-readonly-20261015T0304098327094.log`。随后保持同一数量、大小和 mtime 断言，真实两图符号与关系查询、六条作者地图读取全部 PASS，日志为 `query-readonly-20261015T0305549291173.log`；当前查询使用上述 immutable URI。

`uncheckpointed-wal` 也实际 PASS：使用 Node 官方 SQLite 在自有真实图上执行保持原符号名称的 UPDATE，真实事务提交后保持 writer 连接，确认实际 WAL 非空；A 查询返回 `index_stale`，B 查询正常。关闭 writer 完成 checkpoint 后，A 的实际 Count 符号与调用关系再次查询通过。`tests/s4/wal.mjs` 只用于这个 SQLite 并发边界，不创建替代 CodeGraph、API、源码对象或符号。

只读修复 SOURCE `f407337f2199d5e49f1ff49862063d964a738a9c` 已普通 push，并直接交 H 与模块负责人消费。该精确源码的 Node 语法检查、`go test -mod=readonly ./internal/workspace/... ./internal/adapters/repositorycontext ./tests/s4/...`、`go vet -mod=readonly ./internal/adapters/repositorycontext ./tests/s4/...`、`go build -mod=readonly ./...`、gofmt 检查与 `git diff --check` 均 PASS。Go 日志为 `D:/Astrocyte-agent-assets/20261015/relay/S4/readonly-source-go-{test,vet,build}-final.log`；实际 WAL 验收日志为 `.verification/s4/actual-first/uncheckpointed-wal-20261015T0310375483343.log`。

H 的准确注册、认证与实际 Service 构造已按下文完成 HTTP 验证。SemanticIndex 用户决定、B 生产配置与 main 接线、前端和最终 Windows 包组合及 AOCI 分发条件分别保留待定或 NOT_RUN。

SPEC §9.5 的内部磁盘清理 pass 已按末节实现和真实验证；生产异步调用仍待 B 的既有项目授权动作与 worker 接入。当前 Agent grant 撤销立即拒绝该主体读取，源码删除后的 sync 移除当前图中的旧文件与符号。单个 Agent 的授权撤销保留其他主体仍可使用的共享项目图，GET 查询保持只读。

## 生产配置与发布资产交接

模块 `msg_ba1ff43b111f` 已确认由 B 的可信启动配置显式提供 `ASTROCYTE_REPOSITORY_ENVIRONMENT_ID`，贯穿同一 S4 / S2 快照 provider。身份必须非空并在重启后保持稳定；缺失时相应快照与索引能力明确 unavailable，其他服务可以启动。测试值 `s4-public-windows` 只用于本轨公开项目验证，不作为生产默认。H 负责薄组装，B 负责 config.go，S4 不修改共享入口。

生产 `CacheRoot` 使用实际应用数据目录下的 `repository-context`，独立于源码和发布包。可信预算允许的已有上限为 MaxFiles=20000、MaxBytes=134217728、MaxResults=200、MaxDepth=5；每次 HTTP 查询仍必须提供全部显式预算。Graph 的 Node / SDK / Script 须成组完整，部分配置立即失败；AOCI 可独立显式配置，保存地图的读取不需要其二进制程序。

发布路径为 `runtime/node/node.exe`、`runtime/sdk/node_modules/@colbymchenry/codegraph` 和 `runtime/sdk/astrocyte/repository-context/codegraph.mjs`。P 从固定官方 Git `a1489f77a6d69538bfe29020b8366ee034f90860` 实际执行 npm ci、build、标准 npm pack 和生产安装，交付 `version=1.0.1`、`main=dist/index.js` 的 SDK 包。Node 固定 24.16.0。Graph.SDK 指向这个包根目录，真实导出 CodeGraph、QueryBuilder、setLogger 和 silentLogger，包含其生产依赖与 WASM 资产；这份源码 SDK 不需要 npm-wrapper 的平台包。发布方继续承担第三方许可、归档与解压验收。

AOCI 只接受显式外部本机绝对路径，所选版本与源码标识见前文；分发条件未决定前不把其二进制程序放入发布包。`.verification` 与 `dist/s4` 的工具、数据库和日志均不作为生产默认或包内容。上述生产配置、版本和资产约束已直接交 B、H、P 和产品负责人。

## H 生产 HTTP 与发布 SDK 的实际组合

准确生产源码为 H 的 `31256d5d6758daf25892d55df73feffde7a7e573`，独立报告 `89d84d55a822e2f0506e18e93adaa1ec02a8342e`。按模块 `msg_67fc82a35ed3`，使用该精确 Git 归档和成熟 Go overlay，只映射本轨 `tests/s4/_http/main.go` 为验证入口；生产 NewServer、认证 guard、路由与业务源码保持原样。没有在 H 活动工作树编译或合入其旧历史。`tests/s4/http.ps1` 将真实关闭后的本轨 SQLite 和 CodeGraph 投影复制到独立自有目录，核对无非空 WAL，再通过实际 TCP 服务执行操作。验证仅使用原两棵 Agent 创建的公开 Git worktree，没有原生控制器或新模型调用。

实际资产为 P `msg_c77e7d591a08` 明确 ready 的 `D:/Astrocyte-agent-assets/20261015/release-package/runtime-fourth/runtime/node/node.exe` 与同 runtime 的 `sdk/node_modules/@colbymchenry/codegraph`。版本、入口和实际 API 已读取；SDK require 实测 151.627 毫秒，已有两图的真实符号、调用关系及 adapter 重启读取 PASS。原始 SDK 和 Node 资产只读消费。

首次实际 HTTP 搜索在既有 60 秒 client timeout 下失败，原件为 `D:/Astrocyte-agent-assets/20261015/relay/S4/http-31256d5-first/http-first.log` 和其 `evidence/http-failure-20261015T040514.511460400.txt`，原因仍 UNKNOWN。随后保持原 timeout 和断言，添加仅写本机的 10 秒等待堆栈诊断；完整第二轮 PASS，等待诊断触发数为 0。第二轮日志为 `D:/Astrocyte-agent-assets/20261015/relay/S4/http-31256d5-diagnostics-second/http-first.log`，没有把首失败改记成功或宣称修复其未知原因。

实际通过：缺配置时仍需会话与 CSRF，合法请求明确 503；真实 map/search/impact 及无私有路径投影；未知 DTO、空用途、缺少/超长 query 或 symbol、绝对/逃逸 root、结果预算拒绝；畸形 percent/semicolon、重复标量与未知参数拒绝；真实签发 token、Agent 三种读取、跨项目拒绝、维护仅人类与授权撤销；实际暂存已有功能源文件后 A 的图/地图 stale、B 保持当前，显式 sync/refresh 恢复；HTTP / SQLite 关闭重开后 generation 与地图保持。这个单次状态验证需要尚未暂存的真实 `textstats/stats.go` 功能修改，不制造或重置 Git 状态来重复取得通过。

P 源码 SDK 在 `D:/Astrocyte-agent-assets/20261015/relay/S4/sdk-first` 新建真实 SQLite 与投影，分别从零建立两棵图，实际同 commit / 不同 worktree、同 SQLite 事务快照校验、篡改拒绝及两图真实符号和关系全部 PASS。其图数据库路径长度为 191。此前 `.verification/s4/packaged-source-sdk-first` 的 255 字符图数据库路径在启用 WAL 时失败；标准 Windows 命名空间路径可建库但 SDK 扫描返回空图，两次原件均保留。当前桥使用普通路径，较长 Windows 投影路径仍是已观察到的组件限制；没有修改 SDK 或缩短原公开仓库路径掩盖失败。

本 HTTP 组合验证覆盖真实已配置 Service / HTTP；B 生产配置解析与 main 接线、前端、最终 Windows ZIP、SemanticIndex 仍待各自准确组合验收。旧 generation 清理的独立实际结果见末节。

验证工具 SOURCE `faab2daf83f6a578773b61077140aff73af9ef9b` 已普通 push。实际适用命令 `go test -mod=readonly ./internal/workspace/... ./internal/adapters/repositorycontext ./tests/s4/...`、`go vet -mod=readonly ./internal/adapters/repositorycontext ./tests/s4/...`、`go build -mod=readonly ./...`、对 H 精确归档和相同 overlay 的 `go vet ./tests/s4/http`、两个 PowerShell 文件 parser 检查、gofmt 检查和 `git diff --check` 全部 PASS。未知验证阶段实际返回非零，原件为 `D:/Astrocyte-agent-assets/20261015/relay/S4/unsupported-phase-first.log`，没有静默打印阶段成功。

发布 SDK 的 `initial`、`map`、`query-readonly`、`uncheckpointed-wal` 实际日志保存在 `D:/Astrocyte-agent-assets/20261015/relay/S4/sdk-first`；失败的 C 盘长路径目录及全部旧阶段证据保留。此次代码变更仅包含本轨验证入口及文档，不修改 H 的生产源码、P 的 SDK / Node 资产或任何冻结领域。

## 失效投影清理与既有 worker 衔接

内部端口为 `app.RepositoryProjectionCleaner.CleanupRepositoryContext(ctx, maxGenerations)`，结果只含 `RetiredIndexes / DeletedGenerations / PreservedEntries / HasMore`。调用者必须是持有同一个 `repositorycontext.Service` 的既有后台 worker，且在请求事务之外调用。该方法复用索引维护和图查询的互斥锁；一个 CacheRoot 只能由该单一服务实例维护，不能由另一个进程并发构建或清理。

SQLite 先读取并关闭完整索引 / 项目查询，再于 SQL 写事务之外通过真实 Git 核当前注册根的 WorktreeID 与稳定 EnvironmentID。随后在短 `WithWorkspaceTx / BEGIN IMMEDIATE` 中重新核对当前 project JSON、完整 index JSON 与 generation；任何变化立即返回 `version_conflict` 并回滚。目录授权撤回、allowed_subdirs 为空、权限 revision 不符或项目已删除时，按准确项目 / worktree / backend / generation 撤下失效元数据。Git 观察错误立即返回，临时不可读取不作为来源撤回。Root 不通过公开 JSON 获取，也不写回外部提供的自证值。

元数据提交后，清理只检查可信 CacheRoot 的目录。所有保留索引共同保护其 generation，包括同一 worktree 注册的其他项目；单个 Agent grant 撤销不会撤下项目索引。合法名称必须是 64 位小写十六进制 worktree 身份和标准 UUID generation，其他条目保留并计数。目录链接被拒绝。每次递归删除前验证解析后的绝对目标在 CacheRoot 内，再使用 `os.Root.RemoveAll` 约束操作；源项目、AOCI 原文件和发布资产不作为删除对象。

每次最多删除 `maxGenerations` 个目录，范围为 1–1000；还有旧 generation 时返回 `HasMore=true`。清理未发布或已被替代的旧 generation，保留当前图供增量 sync 使用。事务已提交而磁盘清理失败时，不重新开放旧结果；下一次 pass 根据真实元数据继续清理。现有表已足够保存事实，此变更没有新增 DDL。

根授权 `msg_0f397f12b04e` 与模块 `msg_40c63d059b3a` 已明确专用 Go 清理生命周期，由本轨在同一 Service 实现。B 在项目目录设置 / 注册根变更或项目删除成功提交后窄唤醒，H 注入同一 adapter 并负责启动、错误上送、取消等待和随后关闭 SQLite。该 B/H 生产接线尚待准确 SOURCE；HTTP 不提供清理命令。

实际命令为 `tests/s4/cleanup.ps1 -WorkRoot D:/Astrocyte-agent-assets/20261015/relay/S4/cleanup-second -SdkPath D:/Astrocyte-agent-assets/20261015/release-package/runtime-fourth/runtime/sdk/node_modules/@colbymchenry/codegraph -NodePath D:/Astrocyte-agent-assets/20261015/release-package/runtime-fourth/runtime/node/node.exe`。在原两个公开 Git 根与全新自有 SQLite / 投影中，官方 SDK 实际建图，真实 Agent grant 撤销后人类仍可查询；项目授权收窄后图和六条实际作者地图立即拒读。真实 Windows junction 导致安全拒绝，失效元数据已经提交、全部图文件仍在；移除该自有 junction 后，每次删除一个目录并明确报告剩余，最后当前 A 投影从磁盘移除，同 worktree 另一项目与 B 的图和关系保持可查。独立进程重新打开 SQLite 后清理幂等，剩余投影的数量、大小和 mtime 不变。

原件为 `cleanup-second/{initial-20261015T1404321108530.log,cleanup-20261015T1404424765055.log,cleanup-restart-20261015T1405033435118.log}`。首相对根命令真实被拒，原件 `cleanup-first/initial-failure-20261015T140359.991362100.txt` 保留；后续改用既有 verify.ps1 的绝对公开根，没有改变作用域断言。首包含完整 sqlite 包的 Go 命令因九个 Workbench 测试缺少其独立实际来源环境变量退出 1，详情保留在本树 `.verification/s4/cleanup-first/checkpoint-failures.md`；未给这些测试提供私人来源、移除断言或新增 skip，该九项不计通过。此次真实清理、适用 workspace / adapter 检查和构建结果分别登记。

清理 SOURCE `4ebd777b7b340c717c5311d78a0f26174374bbbf` 已普通提交和 push。该精确源码的 `go test -mod=readonly ./internal/workspace/... ./internal/adapters/repositorycontext ./tests/s4/...`、`go vet -mod=readonly ./internal/adapters/repositorycontext ./internal/adapters/sqlite ./tests/s4/...`、`go build -mod=readonly ./...` 全部退出 0，日志为上述 `cleanup-second/exact-source-{test,vet,build}.log`；PowerShell parser、gofmt 和 `git diff --check` 通过。最后一次独立 `cleanup-restart` 同时实际核验请求事务内调用 production Service 的拒绝与保留图重读通过。首 push 及随后 `ls-remote` 曾返回 `github.com:443 SSL_ERROR_SYSCALL`，之后原 `git push origin HEAD` 成功，未改 TLS 或代理配置。

H 后继主入口 SOURCE `fbf2a4a4752ef2842900b3d3a851f80986734ce9` 的既有 S4 production 文件与本次 SOURCE 的父版本一致。已读取其 `cmd/server/relay_repository_context.go`：原 S2、Map、Search、Impact 和显式索引维护复用同一个 provider；新增清理扩展必须完整消费 adapter、SQLite 和 app/domain 文件。当前源码和报告均交模块、H、B 与发布负责人，生产 worker 接线、SemanticIndex 决定、AOCI 分发及最终 UI / ZIP 仍未完成。

## 专用清理生命周期

`app.RepositoryCleanupNotifier` 仅提供 `WakeRepositoryCleanup()`；`app.RepositoryCleanupWorker` 另外提供 `RunRepositoryCleanup(ctx, maxGenerations) error`。同一 Service 同时只允许一个 Run，第二次调用立即返回 `version_conflict`。Run 一开始就完整扫描，不依赖之前的通知；每个 pass 仍与前台图查询和显式索引维护共用同一互斥锁。显式索引维护创建自有目录后，无论成功或失败都会在离开临界区时通知 worker，当前正在构建的目录受同一锁保护。

通知使用容量为 1 的 channel 合并。持久待清理事实由当前项目权限 / revision / root、保存的索引引用和磁盘目录共同表示：已提交撤回使旧索引失配，元数据已撤下而目录仍存在时形成无引用目录，后继启动均可重新观察。目录实际删除且没有后续条目时，本次 pass 返回 `HasMore=false`；没有单独的完成标记或持久请求表。通知丢失会延迟本进程的唤醒，重启的全量扫描会恢复实际剩余；不能把该 channel 描述为持久队列。

只有 `HasMore=true` 且确有已删除 generation 时 Run 立即继续下一批，否则等待下一次通知或 context 取消。Git、SQLite、权限冲突、路径安全或文件系统失败立即从 Run 返回，由 H 上送到原主生命周期并结束服务；本轨不自动重试。取消时等待当前互斥区和文件系统操作结束，pass 在 Git / SQL 操作和条目之间检查 context，H 必须等待 Run 返回后才关闭数据库。

删除无引用旧图之前，实际 `astrocyte-error-UUID.log` 原件使用 `os.Root.Rename` 移至同一私有 CacheRoot 的 `diagnostics` 目录。日志身份、常规文件及目标目录均严格检查，目标已存在时失败；不覆盖原件。完成移动而后续删除失败时，下次扫描继续删除原目录，日志继续保留。该目录不进入 HTTP 或包内容。

实际运行 `tests/s4/cleanup.ps1` 的新独立目录为 `D:/Astrocyte-agent-assets/20261015/relay/S4/cleanup-worker-third`，Node / SDK 参数与前文相同。真实通过：无通知的冷启动按每轮一个目录删除两个官方图的旧 generation；第二个 Run 被拒绝；实际损坏自有 SQLite 后读取拒绝、正式 rebuild 自动唤醒、原错误日志逐字节保留；实际目录授权撤销后立即拒读，合并一千次通知后清理并取消等待；已提交授权撤回而不发送通知，下一独立进程从真实元数据和目录恢复。B 的真实符号与关系保持可查。原件为该目录的 `cleanup-worker-20261015T1437145478368.log` 与 `cleanup-worker-restart-20261015T1439272899271.log`；原 `initial / cleanup / cleanup-restart` 日志也保留。

`cleanup-second/cleanup-sql-boundary-20261015T1436121380348.log` 实际通过 SQL 边界：官方 `git hash-object --stdin` 在真实输入管道上等待期间，另一 SQLite 写者在既定十秒上限内提交真实项目目录撤销，读取立即拒绝；Git 释放后，重新核对项目变化，清理提交返回 `version_conflict` 且原索引 generation 保持。另一次取消真实 Git 观察后 SQLite 与当前图仍可读取。Git 使用不写对象的实际命令，未替换程序、输出或协议，也没有新增提交或事件。

Go 生命周期测试使用真实新 SQLite 和已有迁移，验证两个并发 Run 仅一个取得执行权、一千次合并通知、取消等待与实际不合法缓存文件的错误返回。没有合成索引、虚假 SDK 或条件 skip；测试目录位于本树 ignored `.verification/s4/cleanup-worker-tests`。B/H 的生产后继源码仍须消费本次完整窄闭包再验证原主生命周期，未把本组件的运行结果作为整体产品验收。

生命周期与短事务 SOURCE `054cc5dba765ee211a8e383a09355719d7ecaf7a` 已普通 push；前置为清理 pass SOURCE `4ebd777b7b340c717c5311d78a0f26174374bbbf`。准确十一文件闭包仅在本轨原独占域，包含两个 worker 文件、清理 / Git / Service、SQLite 清理、app 窄端口、实际检查与本文档。当前 `go test -mod=readonly ./internal/workspace/... ./internal/adapters/repositorycontext ./tests/s4/...`、`go vet -mod=readonly ./internal/adapters/repositorycontext ./internal/adapters/sqlite ./tests/s4/...`、`go build -mod=readonly ./...` 均退出 0，原件为 `cleanup-worker-third/final-{test,vet,build}.log`；PowerShell parser、gofmt 与差分检查通过。完整 sqlite 包中九个独立 Workbench 来源环境缺失的首次失败仍按前文保留，未记为通过。

本次准确窄契约与源码已直接交 B、H 和模块：B 负责原项目事务成功后的 notifier，H 负责同 provider 的主生命周期接线和实际退出；根的 SemanticIndex / AOCI 分发决定、生产完整组合、前端与最终 Windows 包仍待实际结果。没有启动新的原生会话、模型调用或登记 t0，本 Task 按模块要求继续承担后继集成验证与返修。

## 短事务快照引用核验

SOURCE `55fd3853a8230da580c01e3bc427a4074ada5553` 在原 Service 增加 `ValidateProjectSnapshotReference(context.Context, domain.Caller, domain.ProjectContextSnapshot) error`，并实现 `app.ProjectSnapshotReferenceValidator`。该方法只读取保存的固定 `git-snapshot:<generation>` 索引和当前项目、Agent grant；SQLite reader 复用 `WorkspaceTransaction(ctx, db)` 的同一连接，其他数据库的事务 context 立即被拒绝。方法不观察 Git、文件或进程，不创建事务，不写入数据库。

引用校验核对 generation、`ready/git-v1`、项目 / 环境 / repo / worktree 身份、HEAD、index / worktree digest 和准确 ObservedAt；服务 EnvironmentID 也必须一致。SQLite 同时核数据库 generation 列与 metadata 的值。当前真实项目 Root 必须精确等于快照 Root，当前 settings revision 必须等于保存的 SettingsRevision；固定 roots 必须规范且仍在当前 AllowedSubdirs 范围内。每次重新读取目录开关、grant 的项目 / Agent、撤销、过期和 read_context，不缓存权限事实。

RepositoryVersion.Root 继续从 JSON 隐藏，保存的 RepositoryIndex 不携带该根路径；此引用方法直接使用同事务当前 LocalProject.Root 核对快照字段。现有 Domain 字段、存储 JSON 和迁移保持原格式，旧快照使用相同核验流程。S2 在事务外保留 Capture 的双观察及提交前真实 Validate，在最终短事务中调用 Reference 后保存 packet / binding。源码实际变更后，Reference 仍只证明保存引用及当前数据库权限；完整 Validate 会重新观察 Git 并拒绝过期快照。

本轮实际检查入口为 `go run -mod=readonly -p=1 ./tests/s4/verify -phase snapshot-reference -a (Resolve-Path 'dist/s4/worktree-a').Path -b (Resolve-Path 'dist/s4/worktree-b').Path -evidence D:/Astrocyte-agent-assets/20261015/relay/S4/reference-checks/verify-first`。入口先检查两个原根的准确公开 remote，再从原 B 本地 clone 到独占 D 盘目录，运行真实 Git、全新 SQLite 和已有迁移；全部 stdout 仅包含安全 PASS 标签。未启动 CodeGraph / AOCI / 模型或原生控制器。

实际九组检查均 PASS：保存引用重启和私有 Root；临时移走自有 clone、PATH 无 Git 且同连接 query_only 的引用读取；逐项快照字段和服务环境篡改拒绝；当前 grant 撤销 / 过期 / action / scope；同事务尚未提交的 grant 撤销及 Root 变化；其他数据库事务拒绝；保存版本 / 状态 / roots / revision 及 generation 列不一致；当前项目目录开关 / 子目录 / revision / 注册根变化；真实源码新增文件后外部 Git Validate 返回 context_stale。用例直接损坏自有 SQLite 元数据以验证拒绝路径，不伪造成功来源或 CLI。

全部 Go 父进程使用 `GOMAXPROCS=2`，Go 参数 `-p=1`，共享 `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`；TEMP / TMP / GOTMPDIR 仅在进程中指向本轨 `reference-checks/temp`。最终 `go test -mod=readonly -p=1 -count=1 ./internal/workspace/... ./tests/s4/...`、`go vet -mod=readonly -p=1 ./internal/adapters/repositorycontext ./internal/adapters/sqlite ./internal/workspace/... ./tests/s4/...`、`go build -mod=readonly -p=1 ./...`、gofmt 和 `git diff --check` 全部通过。原件为 `reference-checks/{verify-final-first,final-test,final-vet,final-build}.log`；SOURCE 已普通 push，独立 ls-remote 精确核同值。

本轮组合检查首次的 workspace 两包通过，S4 检查进程经精确身份核实后由本 Agent 终止，最终 `exit status 0xffffffff` / 171.749 秒 / EXIT 1 保留在 `reference-checks/test-final.log`，不记为完整通过。后继准确 Root 语义的独立检查实际 22.003 秒通过，原件 `test-reference-final-first.log` 保留；当前源码的正式入口九组结果另存 `verify-final-first.log`。旧 first RED / UNKNOWN / NOT_RUN 记录继续保留。

本次 SOURCE 已直接交原 S2、H、唯一 I 和模块负责人，未合入其 WIP。S2 的 packet / binding 事务消费、准确组合与原十秒 Context 预算、完整 HTTP / UI / Windows 包仍待其实际验收；本轮未执行这些操作，也未执行新增模型调用或登记原生 t0。
