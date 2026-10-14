# 基础后端与导入整改

本轮状态：阶段源码已提交，整轨验收未完成。分支 `standards-backend`，基线 `4ec19cddb24068e09597357b0aabc6ecb9420930`，接管 HEAD `e1d0f28f75a3751da5fa39b7d69f9e0ac4763f39`，第一阶段 SOURCE `2644083960649d7c50262ec69fff730b411241a2` / REPORT `6a4415faaa1e092ae1eab42cffed4f3ffa090434` 均已 push。第二阶段 SOURCE `7f58adecc41070fbec4f88962cfb6fb1723679c6`。早期生产服务隔离 SOURCE `86d8f910eb5d6641f84bcc248d86444fb7bd81cb` 与作用域契约 SOURCE `e1d0f28f75a3751da5fa39b7d69f9e0ac4763f39` 保留。

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

公开 DOM 来源为 `https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514` 与 `https://aclanthology.org/2024.acl-long.1/`。本报告只给聚合与本机证据指针；原件保留在 `.verification/`。

## 真实剩余限制

- 当前全包检查为 RED，公开网络存在 TLS `EOF`、`fetch failed`；源请求失败未用本地伪造响应替换，未加 skip 或放宽断言。
- 本写域仍有外部替代实现：`internal/attention/app/attention_test.go` 的内存 Repository/ObjectStore 与 `automatic_test.go`、`project_distiller_test.go` 的替代模型；`internal/workspace/app/works_test.go` 的替代存储/扫描/交付/事件；HTTP works 的替代注册来源；OpenCode 和注册项目测试的替代来源；M5 MCP 与 delivery 测试的替代 backend、失败存储和确认丢失；SQLite `local_agents*_test.go` 的替代原生控制器与引用读取。
- `tests/s1/automatic_test.go` 仍有替代模型错误；`tests/s1/fixtures.mjs`、默认 acceptance 与部分 web E2E 仍依赖虚构纸/视频来源。B 继续整改自身测试与共享输入，web 改动交 F，不能把现有默认套件通过记为真实来源验收。
- 全写域英文注释清理尚未完成。原生格式 helper 的解码路径及其他同类异常需要继续核查；已增加格式门禁不等于全域零静默解析。
- 实际 Codex controller / SQLite / HTTP / 自产会话逐 event_id 验证归 R/H，当前 B 没有运行该组合验收。发布完整 E2E、Workbench/Relay 集成与 ZIP 解压验收尚未运行。
- Mission 创建仍使用生产 501 状态，未获业务扩展授权。V4、人评、正式库、冻结脚本、引擎、schema、迁移、权重和日程保持原边界。

## 未执行操作与后续交付

首次 SOURCE push 因 GitHub `SSL_ERROR_SYSCALL` 失败，随后普通 push 已成功；第二阶段首次 push 同样失败，后续状态以实际 Git 结果登记。不 force push。本报告作为独立 REPORT 提交并推送。阶段 SOURCE/REPORT 不结算整轨，后续源码与报告继续追加；最终精确 SHA 和剩余限制由最后一次报告登记。
