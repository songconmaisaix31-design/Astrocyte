# 安装物授权工具

这是外部验收工具。`prepare`和`verify`只做本地安装、源码、公开输入和服务身份原件核对；二者不会发HTTP、创建来源或取得human会话。`execute-authorized`会执行已委托的人工授权写入和验收驱动，执行范围须有明确用户授权。2026-10-23 用户开始测试N>3已覆盖同公开项目、StepFun、ToolsOff的N4/N6及原测试所需授权、收尾；新包只作技术来源/身份核查，不逐SHA重复索取同意。命令名或包内bool本身不产生授权，其他范围按原规则处理。

在仓库根目录构建，把实际二进制和同版本的纯公开消费者一起放在新的独占工具目录中：

```sh
go build -mod=readonly -o "$TOOL_ROOT/authorize-installed" ./tests/swa-cli-headless/authorize-installed
cp tests/swa-cli-headless/authorize-installed/public-case-preflight.mjs "$TOOL_ROOT/public-case-preflight.mjs"
"$TOOL_ROOT/authorize-installed" -h
"$TOOL_ROOT/authorize-installed" -mode prepare -package "$PRIVATE_PACKAGE"
```

本工具只支持Linux x64、Go 1.27.2。工具目录使用本次新建的0700普通目录，包和证据为0600普通文件；临时目录放在仓库的ignored `.verification`中。consumer从真实二进制旁查找，并将实际SHA写入准备包；不能依赖原ignored开发目录中的副本。包中安装VERSION、源码commit、driver各依赖、consumer、二进制及装内Node/swa均需准确固定。`prepare`另写`.prepared.json`，不能覆盖已有包或证据。

`profile`仅支持以下固定范围：

| 包字段 | 默认Codex：不填profile | StepFun Pi：profile为stepfun-pi-n3 |
| --- | --- | --- |
| caller | swa-codex-n3-20261022 | swa-stepfun-pi-n3-20261023 |
| agent / model | codex / gpt-5.5 | pi / stepfun/step-3.5-flash |
| workers | 3 | 3 |
| 来源目录字段 | codex_api_key_home | pi_source_home |
| 项目名 | SWA Codex N3 public stats | SWA StepFun Pi N3 public stats |
| 空间名 | SWA Codex N3 public stats 20261022 | SWA StepFun Pi N3 public stats 20261023 |

Pi新增profile `stepfun-pi-n4`和`stepfun-pi-n6`，workers分别只能4、6；caller分别为`swa-stepfun-pi-n4-20261023`、`swa-stepfun-pi-n6-20261023`，项目/空间标题的N同步变为N4/N6。各自公开项目固定为`/workspace/swa-public-case-20261023-n4/public-project`、`/workspace/swa-public-case-20261023-n6/public-project`。Agent、model、来源字段和ToolsOff仍沿Pi表，不允许用其他N替换profile。N6在N4实际完整结算之后执行，不能同时抢用同一driver输入或owner。

旧Codex/Pi N3包仍只接受既有公开项目`/workspace/swa-public-case-20261022/public-project`。所有profile核同公开revision四原文及固定task/input/followup；N4/N6使用六种区别明确的公开审阅职责中的前四/全部六项，显式initial_workers=N，经默认等待run创建lead后补同Run N−1；槽位4为lead的真实孙任务。两批均主动停非lead槽位2；N4只在lead首轮idle且孙任务仍运行时追加lead，N6用子成果原mail/callback产生lead第二轮，不另加手工模型轮。四个实际项目Role和Conversation、完整递归wait/export与自动release沿原回执/原生来源核对；提示职责不代替持久Role，stored通知不当provider送达。未知profile、错caller/agent/model/N/项目或另一profile的非空来源字段在HTTP之前拒绝。来源必须为显式、独立、当前用户0700普通目录；工具只检查其路径元数据，不读取auth、key、models或settings正文，不创建或修复来源。缺少来源、真实启动记录或服务身份时包保持INCOMPLETE，不能执行。Pi的API为openai-completions，endpoint为https://api.stepfun.com/v1，ToolsOff合同为pi-cli-1.0.1，新账户kind为pi-native-api-key-per-session；这项固定目标不表示账号、推理、原生ToolsOff或N已通过。

原serve owner须在真实spawn时保存非秘密启动记录并提供原ready/proof与准确SHA/runtime。Codex记录严格为installation、source_revision、data_root、endpoint、serve_pid、codex_api_key_home六键；Pi记录最后一键为pi_source_home。工具不生成或补造记录。Linux Codex/Pi驱动均接收已核原ready的`--service-ready-record`和`--service-ready-sha256`；驱动绑定实际status/runtimeId、装内Node与原daemon→runtime→serve身份链，使用独立runtime托管者核Worker custody。daemon原身份与认证检查保留。Pi另接收已核记录的`--pi-launch-record`和`--pi-launch-sha256`，从中独立核对SOURCE。真实serve捕获来源的入口是PI_CODING_AGENT_DIR，由原owner负责；工具不启动serve或调整PATH。

Codex/Pi共同固定driver-failure.mjs、codex-event-view.mjs、prepare-source-observer.mjs、source-observer-entry.ts、extract-observer-source.py和pi-event-view.mjs六个执行依赖；Pi再增加pi-source-observer-entry.ts。pi-event-view的私有FD读取器也被Codex用于ready文件，Codex仍使用已有JSON解析，不引入Pi protocol运行依赖。全部依赖逐安装VERSION的Git blob和实际SHA核对。缺少依赖的历史安装不能作为执行候选。源码测试、编译与help只能证明离线范围。

委托效果包括本次human会话、实际项目/空间登记、observe及execute两阶段settings CAS/Grant/token、原驱动业务请求和明确清理。cookie/CSRF/token只保存在内存，token只通过子进程环境传递；key不进入本工具。human写结果未知时停写且不重放，不能用CLI请求账本宣称human写已结算。业务未知只查原R。撤销与human未知后的撤销分别由包中明确委托控制；服务由原serve owner正常EOF并取得全树关闭证据。过期055b及其旧owned服务不得复用为新执行包。

driver的外部审阅快照和私有证据保存在EvidenceRoot。公开项目下另有固定公开正文副本及私有result/checked-result导出；只有批准的四原文正文可进入模型，私有结果不作为后继输入或上传。来源默认设置与每Worker三项设置分开。原生metadata中的cost=0仅为客户端默认值，不表示远端账单为零。

退出码：0表示相应工具步骤结束，真实业务仍需独立审阅；2表示参数或本地前置拒绝；3表示缺少必要前置；执行期间真实失败保留原UNKNOWN和安全错误分类。当前实际HTTP、人工授权、原生启动、模型及N验收结果以各自独立报告为准。

空间回执的 `id` 是不透明字符串，不能套用项目 ID 的 UUID 校验。拆分前的空间服务与当前 Workspace 服务都由 `crypto/rand.Text` 产生空间 ID；工具核对 schema、非空且无首尾空白的 ID、版本 1、原提交标题、非 null 的空资料引用和有效创建时间。项目 ID 仍须是 UUID，且继续核对实际目录、原空间和默认关闭的权限。

2026-10-23 首个 S50 N4 包在空间 POST 201 后被工具的旧 UUID 校验误拒为 `actual_space_receipt_missing`；原请求、原 UNKNOWN 和正常关闭的服务证据保留，修复不重放该请求，也不改产品安装。`TestActualSpaceReceiptAndProjectRegistration` 使用新的临时 SQLite 与真实 SWA HTTP listener、human 会话和 CSRF，只创建一次新空间、只读核对原 HTTP/SQLite 回执并登记实际公开测试目录；不执行 Grant、原生运行时或模型。错误字段和 UUID 兼容仅作为真实回执的纯解析变体验证。该局部测试不代替重新固定来源与服务身份的新验收包或真实 N4/N6。
