# S4 仓库上下文

本轨基线为 `da79452d1735f8fceca103aaf9d10bb53182f1f5`，分支 `songconmaisaix31-design/works-repository-context-s4-20261015`。客户端为 Codex API；本会话没有可独立核实的精确模型标识。按模块确认，仅消费 S2 的 `4eec31ddfde39fa3f548691ec6e7f2f34470351a`、`9c5b1619710b576f291d42a5aa641b28639d5a57`、`5d0ba735533f68259b91128dc404c8e4c2f2a535`，以及 B 的单表迁移 `98bb9236c68f45ca4ab4114d46eafba5c90cdedf`，均保留 `cherry-pick -x` 来源。

## 设计与接入

`repositorycontext.New(projectReader, indexStore, Options)` 组装真实 Git、官方 CodeGraph 和 AOCI 观察适配。`sqlite.DB` 实现两个存储端口；H 负责服务与 HTTP 薄组装，S4 不修改 server、OpenAPI、依赖或迁移。S2 使用原 `ProjectSnapshotProvider`，没有复制其 Domain 类型。

Git 从注册根核对 worktree 根、common-dir、git-dir 与真实 HEAD；授权范围的 stage 清单和实际文件内容分别形成 index/worktree digest。文件清单包括未忽略的未跟踪文件和已删除记录。项目环境、仓库、worktree、项目、权限 revision 和准确相对 roots 共同决定版本。根目录只在内部观察和 S2 本地快照中使用，仓库查询投影不序列化根目录。

每次读取重新加载真实项目与 Agent grant，检查当前撤销、过期、`read_context`、目录开关和允许子目录。用途、文件数量、读取字节数、结果数量和展开深度必须显式提供，受可信启动配置上限约束。密钥目录、环境文件、依赖、构建目录和符号链接不进入源码投影。查询前后重新捕获真实版本；变化返回 `index_stale`，不发布旧关系。

`CaptureProjectSnapshot` 保存不可变的 Git 观察和真实授权范围；`ValidateProjectSnapshot` 重新核对当前 Git、项目许可和保存的版本，不信任客户端的 Root 或 digest。保存使用 `repository_context_indexes` 的独立 `git-snapshot:UUID` 后端键。SQLite adapter 在已有 Workspace 事务 context 内调用 `WorkspaceTransaction(ctx, db)` 复用其连接，避免单连接事务的二次借用死锁。

CodeGraph 使用官方 npm `@colbymchenry/codegraph` **1.0.1**，MIT，源码标签 `v1.0.1` 对应 `a1489f77a6d69538bfe29020b8366ee034f90860`。可信配置提供 Node、SDK 和本轨 `codegraph.mjs` 的实际路径。仅子进程设置 `CODEGRAPH_TELEMETRY=0`、`CODEGRAPH_NO_DOWNLOAD=1`；不运行安装器、不修改全局 MCP 或 Agent 配置。失败原始诊断只保存在本平台 ignored 投影目录的 `astrocyte-error-*.log`，HTTP 错误保持安全固定原因。

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
| AOCI 源码构建、实际版本、官方 MCP 九工具 | PASS |
| 模型作者六对象 FRAS → 官方整批 apply → Verify/Check/Guide | PASS：applied=6、aligned=true、complete=true、next_action=none |
| 保存 AOCI 地图、无 AOCI 配置的 service 读取地图 | PASS：六条语义和相对引用；GET 没有 AOCI 进程或 ledger 写入，`.aoci` 文件数量、大小、mtime 相同 |
| 无初始化的 B 地图 | PASS：真实 `evidence_missing`，没有伪造语义 |
| 真实 Git 暂存区变化 | PASS：A 的地图和图均 `index_stale`，B 查询保持当前；显式 sync/refresh 恢复 |
| 查询结果数量预算 | PASS：图与六条地图超限均拒绝，没有静默截断 |
| SemanticIndex、HTTP/前端/最终包组合 | NOT_RUN |

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

HTTP SOURCE `896a7b430c62f9d91aaf5434b11c7653a3d00f36` 已普通 push，准确注册函数、五个端口字段和 DTO 已分别交模块负责人、H 与 UI 负责人。`go build -mod=readonly ./internal/adapters/httpapi`、`go vet -mod=readonly ./internal/adapters/httpapi`、`git diff --check` PASS；真实 HTTP/认证/SQLite/重启组合仍 NOT_RUN，等 H 的准确生产薄组装 SOURCE 后由本轨继续验证。没有用编译通过替代实际接口通过。

模块独立审读 `msg_1d21efa0bebb` 指出 GET 查询解析应拒绝畸形编码。本轨使用成熟 `net/url.ParseQuery(r.URL.RawQuery)` 并立即拒绝解析错误，避免静默忽略畸形未知参数；HTTP 验收脚本已准备畸形 percent、semicolon、重复标量与未知参数场景，实际执行仍待生产组装，不使用自研 parser。

查询解析修复 SOURCE `293af633570d6a4f4cf4bbf9c5b651821ab977e7` 已普通 push；当前 HTTP 包 build/vet 和差分检查 PASS。真实 HTTP 验收脚本仅在准备目录中，未宣称实际运行成功；H 生产注册/认证准确 SOURCE 是其当前依赖。模块及 H 已收到准确 Services、注册函数、DTO 和权限规则，F/UI 仅按 H 唯一契约生成类型消费。
