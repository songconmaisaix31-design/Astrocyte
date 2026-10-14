# 基础后端与导入整改

本轮状态：阶段源码已提交，整轨验收未完成。分支 `standards-backend`，基线 `4ec19cddb24068e09597357b0aabc6ecb9420930`，接管 HEAD `e1d0f28f75a3751da5fa39b7d69f9e0ac4763f39`。第一阶段 SOURCE `2644083960649d7c50262ec69fff730b411241a2` / REPORT `6a4415faaa1e092ae1eab42cffed4f3ffa090434`，第二阶段 SOURCE `7f58adecc41070fbec4f88962cfb6fb1723679c6` / REPORT `a3f093ddd3488df9c71f2ad295724dba06e92efb`，依赖 SOURCE `2a6f03605ebb87f36f7bd56c577aaf4853cb6671`，第三阶段 SOURCE `6cfadb36d94ab372eee1ef1f1f687947cbb771fa` / REPORT `12b82723509d45e2dfd712a74386523e5c38014f`，HTTP 命令及错误编码 SOURCE `b470ecd736af261a10e4c4822bdcd3b02d0da454`，Workbench Reader SOURCE `f6491131d5e5e6eeb889ec82362caf9f1db04e40` 均已普通 push。早期生产服务隔离 SOURCE `86d8f910eb5d6641f84bcc248d86444fb7bd81cb` 与作用域契约 SOURCE `e1d0f28f75a3751da5fa39b7d69f9e0ac4763f39` 保留。

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
| `go test -mod=readonly ./internal/adapters/httpapi -run 'TestHealth\|TestHumanAttentionHTTPFailure\|TestJSONEncodingFailureReturnsErrorV1' -v -count=1` | EXIT 0，实际服务失败与明确不可编码值保留完整 ErrorV1；`.verification/http-json-error-first.log` |
| `node --test tests/s1/server-lifecycle.test.mjs`，命令身份修正后 | EXIT 0，两个实际布局各验证成功刷新及显式错误 request_id 400；2 PASS、0 SKIP，`.verification/helper-command-identity-first.log` |
| `go test -mod=readonly ./internal/adapters/sqlite -run TestWorkbench -v -count=1` | EXIT 0，六项实际 SQLite 检查：投影/过滤/日期/分页、数据不完整拒绝、缓存切换、同源及硬链接拒绝、独立源可读取、实际 Windows junction 拒绝；`.verification/workbench-reader-second.log`。输入认知行明确为人工输入，不冒称 native/Sleep 来源 |
| Reader SOURCE 的 `go build -mod=readonly ./...` / `go vet -mod=readonly ./...` | 全部 EXIT 0，`.verification/workbench-{build,vet}-first.log`；架构检查 31 包通过 |
| 实际 Reader 测试二进制扫描 | EXIT 0，旧个人路径常量不在实际编译二进制，21141504 bytes；`.verification/workbench-binary-first.log`。验证从不可变原始代码 AST 提取字符串，只比较二进制 bytes，未打开该原始路径 |

公开 DOM 来源为 `https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514` 与 `https://aclanthology.org/2024.acl-long.1/`。本报告只给聚合与本机证据指针；原件保留在 `.verification/`。

## 真实剩余限制

- 当前全包检查为 RED，公开网络存在 TLS `EOF`、`fetch failed`；源请求失败未用本地伪造响应替换，未加 skip 或放宽断言。
- 原有 attention 应用测试仍有内存 Repository/ObjectStore 与替代模型。经双方零 WIP checkpoint 确认，协调者已把 `internal/attention/app/*_test.go` 与 `tests/s1/automatic_test.go` 共十四个既有文件及同目录新测试交给唯一第三层 T，精确起点 `b470ecd736af261a10e4c4822bdcd3b02d0da454`；B 从此排除这些测试写域，生产需求由 T 提交 Handoff。
- B 本写域仍有 HTTP works 的替代注册来源；OpenCode 和注册项目测试的替代来源；M5 MCP 与 delivery 测试的替代 backend、失败存储和确认丢失；SQLite `local_agents*_test.go` 的替代原生控制器与引用读取。后续各项整改继续保留完整覆盖。
- `tests/s1/fixtures.mjs`、默认 acceptance 与部分 web E2E 仍依赖虚构论文/视频来源。B 继续整改共享输入，web 改动交 F，不能把现有默认套件通过记为真实来源验收。公开视频 whisper 转写无时间区间，不能伪造已有 positive 时间区间覆盖所需原件。
- 全写域英文注释清理尚未完成。原生格式 helper 的解码路径及其他同类异常需要继续核查；已增加格式门禁不等于全域零静默解析。
- 实际 Codex controller / SQLite / HTTP / 自产会话逐 event_id 验证归 R/H，当前 B 没有运行该组合验收。发布完整 E2E、Workbench/Relay 集成与 ZIP 解压验收尚未运行。
- B README 一次实际 inference 已核对 HTTP、模型返回、原 native_id、SQLite 与重启；当次进程启动 argv 尚未独立留存。生产代码关闭工具的配置与零工具调用不能补写成当次 argv 的实际观测。B 不重复发送已有成功推断，运行期工具关闭证据继续交 R/H 核查。
- Workbench Reader 六项检查仅覆盖 B 新加入 Reader 及自身实际 SQLite；H 需要在既有路由接入该精确源码，并移除其启动配置中的旧常量引用。发布 HTTP、完整二进制与两次实际 00:30 触发验收尚未运行。
- 当前 arXiv DNS 实际响应 `198.18.0.71`，系统网络配置不在 B 授权写域。完整 Go 检查失败已提交模块负责人；公开 source 与 CLI positive 断言保持，未添加 skip 或伪造响应。
- Mission 创建仍使用生产 501 状态，未获业务扩展授权。V4、人评、正式库、冻结脚本、引擎、schema、迁移、权重和日程保持原边界。

## 未执行操作与后续交付

部分 SOURCE/REPORT 普通 push 遇到 GitHub `SSL_ERROR_SYSCALL`，原件保留；上述已列源码及报告均已普通 push，未 force push。Reader 前四次握手失败保留在 `.verification/workbench-source-push-{first,second,third,fourth}.log`，第五次正常推送成功。本报告作为独立 REPORT 追加提交并推送。阶段 SOURCE/REPORT 不结算整轨，后续源码与报告继续追加；最终精确 SHA 和剩余限制由最后一次报告登记。
