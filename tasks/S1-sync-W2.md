# S1 同步 W2：本地项目与原生会话续作

2026-10-10 当前开发源码 `d5ee4a11826a4f654562a19c8f8ac41009d04e5f`，分支 `s1-local-agents-1010`，已 push；完整 context 实现起点 `3637de70a74b2c3e143a2bacfea6c37286e83d12`，报告提交 SHA 另由 Orca Handoff 提供。先普通合入指定 `ca7ba6e80ab256a33465c5bdd959400375569a0e`，并普通消费 W0 公共端口/迁移；未 reset/rebase/clean/force。开发客户端仍为 Orca Codex，协调者明确提供的开发模型为 `gpt-6.1-sol`。下方库存旧报告保留为历史阶段，旧待答结论不代表当前授权。

## 当前完成及所有权

- 人类显式登记项目根，默认 A 固定版本选择引用；B 目录/子目录及 C 关联引用分别 opt-in、动态撤销。目录拒绝越界、符号链接、凭据/隐藏客户端目录；显式文件与固定引用合计 128KiB、每类最多32，C 仅一跳加 cycle guard。候选发现仅所选根，深度3/2000项/100项目；不回退 cwd、不扫描私有根。
- 人类修改设置/授权并签发1小时 project-scoped Agent token，Agent 无法自授权或充当 human；每次重新核实项目、动作、授权及固定引用。模型外发只覆盖人类指定的当前 CLI。全部 native filesystem/command tools 关闭；非空工具 allowlist 返回 unsupported，未把协作式 CLI 称为 OS sandbox。
- Codex app-server、Pi RPC、Claude stream-json/control 原生适配器与通用 registry / selected-text 消费端口，安装/协议启动/模型成功分开缓存。GET 不启动客户端或更新旧观察时间。其他安装客户端没有已验证 driver，操作返回明确 unsupported；不是“没有原生协议”的结论。
- 原生 start/resume/send/stop/observe/read_context 有真实协议实现。Same native ID 恢复保留累计 B/C 引用；权限收窄后 send/resume/read/observe 输出均拒绝，human owned stop 保持可用。每项目一 writer、registry全局4进程、1–1800秒 lifetime、128KiB/256事件输出、每 session64条投递回执；部分写/失联/未确认停止不重发、不恢复。确认未启动的失败为 unstarted/failed，可另发新操作；已投递 UNKNOWN 保留。
- 后续发现并修正：owned read_context曾只返回附着事件。现在Codex先thread/read校验原ID/cwd，再thread/items/list固定32项/最多8页/重复cursor拒绝；initialize明确协商experimentalApi，协议返回-32601保持unsupported。Pi get_messages读取原会话当前conversation，过滤仅human/assistant文字。超过128KiB/256条不返回伪完整部分。原生读取非模型调用，不自动启动已停止进程；Claude完整历史读取明确unsupported，observe仍供当前附着输出。此次完整读取仅离线协议检查，安装协议真实验收仍NOT_RUN。
- `NativePrelaunchFailure`只标记能证明新操作未启动的失败。已确认停止的原会话resume遇到此失败时，保留原owned/nativeID/正向停止事实，将本次回执标记failed，允许新显式操作重试；不把旧会话改成unstarted，也不清除实际投递UNKNOWN。实际进程可能已启动的ownership错误仍保持unknown。
- SQLite 实际持久化项目、设置、grant/token digest、原生 ID、最小范围和投递回执，不存上下文正文。旧空 map/slices 正常化；历史无时间为 null。外部历史只支持所选根的 Codex/Pi header 过滤，read-only/external_observed；不接管外部进程，restart需重新发现。Claude 外部历史格式未验证，返回 unsupported。
- 仅修改 W2 write_paths；公共端口/HTTP/入口/迁移由 W0、引用消费由 W1、UI/browser 由 W3，未跨 app import。主要提交 `022585c`、`3e08b0f`、`9316c89`、`45aca51`、`f7fefeb`、`1bc9741`、`b0645f1`、`e839fba`、`5491fd3`、`3637de7`、`5599350`、`5219421`、`d5ee4a1` 均已 push。

## 验证及首次失败

| 实际命令/范围 | 结果 |
|---|---|
| `go test ./...` | PASS；默认跳过 opt-in CLI/model，不能将 SKIP 当 native 实测 |
| `go vet ./internal/adapters/agents ./internal/workspace/... ./internal/adapters/sqlite`、`go build ./...` | PASS |
| `GOOS=linux go build ./internal/adapters/agents ./internal/workspace/...` | PASS，仅交叉编译 |
| `git diff --check` | PASS |
| SQLite 临时库关闭重开 | PASS：默认拒绝/B-C范围/身份跨项目/动态撤销/累计恢复范围/未知投递不重发/未确认停止不恢复/并发Observe不覆盖已接受回执/并发token issuance不能覆盖revoke/旧DTO正常化；原生 adapter fixture 明确离线 |
| 真实临时 helper parent+descendant | PASS：owned stop确认父退出且后代未写标记；不是安装 Agent/model 实测。部分写UNKNOWN、乱序RPC相关、超限取消、共享并发与过期缓存检查 PASS |
| `ASTROCYTE_TEST_NATIVE_HANDSHAKE=1 go test ./internal/adapters/agents -run '^TestNativeHandshakeLiveNoInference$' -count=1 -v` | 单槽授权下 Codex/Pi/Claude启动协议、观察idle、owned stop PASS；零模型 turn |
| `ASTROCYTE_TEST_NATIVE_MODEL=1 go test ./internal/adapters/agents -run '^TestNativeLiveTurnAndOriginalSessionResume$/codex$' -count=1 -v` | 两真实公开marker turns、相同原生 ID resume与正向 stop PASS，16.63秒；实际 model `gpt-6.1-sol` / provider `openai` |
| 同上 `/claude$` | 两真实公开marker turns、相同UUID resume与正向 stop PASS，14.23秒；Claude2.1.238实际 model `qwen3.7-max`，provider端点未证明，仅标记 `configured_cli_transport` |
| Pi首次真实turn及 root 单次有限诊断 | FAIL：无正文、实际 `claude-opus-4-8`/`anthropic`，有限原因 `authentication_rejected`，owned stop确认；不改认证、不重试、不伪造成功 |

最新源码 `d5ee4a1` 的 `go test ./...`、`go build ./...`及`git diff --check` PASS；新增未启动resume连续重试保留原所有权测试PASS。表中原生握手及真实turn结果是较早版本的独立证据；`5599350`增加的Codex/Pi完整context和恢复后原用户marker断言尚未执行，不能从此前PASS推导新增断言通过。

保留首次 RED：W0未到位前 Caller 参数编译不匹配，合入真实端口后 PASS；Codex thread/start sandbox enum首次 `-32600`（文档与安装schema差异），按本机公开生成 schema 修正 `read-only` 后 PASS；Pi冷启动version5秒超时，有限15秒probe后握手 PASS；Claude包入口首次缺少旧 cli.js，按公开 npm wrapper 改用实际 bin/claude.exe后 PASS。首次 Codex completed 后 interrupt `-32600` 与正向 owned exit分开记录，已改为仅 running 时 interrupt。开发中两次局部编译错误修正后重测通过。W3/W0登记history_roots=null首RED由各 owner保留，W2新旧DTO修复不改写历史。

## 当前限制与未执行

真实个人项目/历史根尚未提供；只用非私有临时项目，不声称实际用户历史接入。未执行跨会话 context_handoff 的真实模型验收；同一原生 ID resume不是该证据。Pi模型当前认证拒绝；OpenCode/Grok/Kimi/Qwen/Cursor native driver未实现/验收，所有安装不推断原生支持。Claude generic selected-text 路径因协议在 paid turn前无法新观察当前model而明确 EvidenceMissing；workspace直接运行已验证，不能用旧 cache假装当前配置相同。原生usage/cost尚未投影，保持 null。

选定材料只读工具关闭；CLI进程仍可自行读取其现有配置/认证，宿主不读凭据，协作模式无 OS读隔离保证。未运行旧sandbox setup、用户ACL/firewall、seat安装或全局SDK迁移；未写用户私有项目。Linux仅build。UI/API/SQLite浏览器验收归W0/W3，W2不冒充已执行；需协调者最后集成验收。真实单槽已明确归还，未再调用native/model/media/browser；未提交 worker_done，等待 root 判断原始范围和真实阻塞。

OpenCode1.18.35官方版本源码与本机公开包进一步核实：native serve协议成熟，但OPENCODE_PURE只抑制外部plugin，不阻止继承的instruction读取；config.instructions数组合并，[]不能清空，MCP按合并后enabled配置自动连接。project-disable不能取消全局/home instruction。没有为通过而启动不安全server或改变当前provider；普通native缓存/依赖初始化已由root说明在用户正常CLI授权内，不再把所有nativecache写当额外审批。第二通用backend仍等待用户选择Pi正常登录/配置或Claude显式per-project模型许可，未冻结共享API。详见探测文档 primary links。

---

## 历史：本地 Agent 库存阶段

2026-10-10；分支 `s1-local-agents-1010`，基线 `20437708e43201e352d6c6926902e1363fd2ad3e`，工作树为同名 Orca 工作树。先读 AGENTS、HANDOFF、STATUS、S1-sync-local-agent-plan 与 SPEC §8.1。客户端为本机原生 Orca Codex；有效模型 `gpt-6.1-sol` 由主控 worker-read projection 明确观察后通过本 Dispatch 告知，Worker 没有读取私有配置或会话历史来反查。

## 已完成和接口交接

- 安装库存独立呈现 installed/configured/startable 与 SPEC 八项 native capability；PATH 发现不代表原生 discover。Fresh CLI 探测确认 Codex 0.162.0、Claude 2.1.238、OpenCode 1.18.35、Pi 1.0.1、Grok 1.0.34、Kimi 2.1.1、Qwen 0.24.6、Cursor GUI 3.23.23。额外 Gemini、cursor-agent 未在 PATH 找到；agent.exe 是 Grok 别名。所有 configured/startable/八项 native 为 unknown / NOT_RUN。
- `agents.NewInventory` PATH-only 初始化；显式 `RefreshCLI` 只允许固定 version/help，5 秒单命令期限、64KiB 保留输出、拥有的进程清理、固定原因码。缓存不会泄漏路径/帮助/原始诊断，不读凭据或会话目录。`Snapshot` 和 app `ListLocalAgents` 不启动命令，未连接 provider 明确 unsupported。取消保留已完成部分并标记未运行或不完整，不制造版本/能力成功。
- W0 已发布保护人类会话的 `GET /local-agents`、OpenAPI/生成客户端和 `InventoryProvider` / `LocalAgentInventory`；W2 通过普通 merge 消费精确 W0 `bb912efd6f28f9f3909d87ea9fc1f4acb04c9098`，不修改他轨文件。W0 负责入口组装，建议启动 refresh 总期限 10–15 秒，查询仍只读缓存；W3 消费同一契约。
- [探测文档](../docs/probes/local-agents.md) 保留旧清点与首次工具错误，记录 Codex app-server/Pi RPC 的本机版本证据、原生接续/送消息/取消/事件流可行路线及具体前置授权要求；没有造 scheduler、Mission、Swarm 或假原生会话。

## 分阶段提交

| 提交 | 内容 |
|---|---|
| `1fe4269` | 独立安装与能力 domain DTO，已 push，供 W0 契约依赖 |
| `331c4c9` | 安全库存适配器、Windows/Unix 探测边界、实测与公开原生接口调查，已 push |
| `763e183df35951a3ba919f593a51168d22fee522` | 消费 W0 契约后提供缓存查询 service、部分探测结果语义，已 push |

本报告提交另包含启动中断边界测试与部分探测明确标记；最终准确 SHA 通过 Orca Handoff 报告，不将文档 SHA 自写入同一提交。

## 实际验证

| 命令 | 结果 |
|---|---|
| `go test ./internal/adapters/agents ./internal/workspace/... ./internal/adapters/httpapi` | PASS；拒绝非 version/help 的 exec/resume/sessions/doctor、命令注入路径；实际临时 helper 超时退出，Windows 带空格 CMD 包装成功且取消后后代未写遗留标记；启动中断停止后续探测，缓存保留部分与未运行差别。HTTP 受保护库存测试由 W0 提供；不是假 native session 验收 |
| `go vet ./internal/adapters/agents ./internal/workspace/...` | PASS |
| `go build ./...` | PASS，包含合入 W0 公共端口后的 Go 项目 |
| `GOOS=linux go build ./internal/adapters/agents ./internal/workspace/...` | PASS（交叉编译；不是 Linux 运行时能力实测） |
| `ASTROCYTE_TEST_LOCAL_AGENT_CLI=1 go test ./internal/adapters/agents -run '^TestInventoryLiveCLI$' -count=1 -v` | PASS；首次适配器 fresh8 CLI version/help 4.54 秒；最终显式部分结果语义实测 5.10 秒（Go 命令总 5.454 秒），8 个版本/帮助通过、2 入口未找到，无模型或原生会话操作 |
| `git diff --check` | PASS |

首个探测失败保留：Python CMD 错误套引号使六个包装脚本 exit1；修正后的打印又遇 GBK UnicodeEncodeError、脚本 exit1。UTF-8/正确 argv 后全8 CLI 探测通过，原失败属于探测器；不得解释为 Agent 原生运行时成功或失败。默认 Go 测试跳过 opt-in CLI 实测；上述显式命令是单独 nonsecret 运行，没有将跳过算通过。

## 尚未授权、未执行和真实限制

用户仍未回答“完整 start/resume/send/stop/observe”或“先库存”、Agent 主动引用文件或具体项目根；W2 尚无可按其授权开发的本地项目/活动绑定范围。没有冻结绑定契约或 SQLite 迁移，没有读取私有历史/auth 配置、扫描整盘、控制既有会话、调用应用模型或真实媒体、运行浏览器验收。旧 UNKNOWN 模型请求未重发。

因此库存阶段已可集成，原 TASK 中项目/活动绑定与八项原生能力验收仍 NOT_RUN，不能据此声称完整本地接入或 S1 完成。未来原生操作需选定明确项目/读取根与目标原生 ID、确认占用交接、工具/外发规则及主控单场运行安排；安装 CLI 与官方协议文档不能证明 native readiness。本次 Windows 临时进程树测试不等于任意原生 daemon 或既有用户会话控制已验证；Linux 仅交叉编译。完整 pnpm check/build/E2E 和真实页面/API/SQLite 重启验收归主控/W0/W3 后续集成，本 Worker 未冒充已执行。

SQLite `adapter.go`、迁移、HTTP、组装、前端、全局 Agent 配置均由既有所有者维护；本轨没有编辑这些业务文件。已向主控和 W0 发送精确源码 SHA、命令结果、启动总期限/部分缓存语义及待答限制，等主控安排后续范围。
