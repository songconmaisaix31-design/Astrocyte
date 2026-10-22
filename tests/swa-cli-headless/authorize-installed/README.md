# 安装物授权工具

这是外部验收工具。`prepare`和`verify`只做本地安装、源码、公开输入和服务身份原件核对；二者不会发HTTP、创建来源或取得human会话。`execute-authorized`会执行人工授权写入和验收驱动，必须先取得用户对准确执行包及清理范围的明确委托。命令名本身不表示已有委托。

在仓库根目录构建，把实际二进制和同版本的纯公开消费者一起放在新的独占工具目录中：

```sh
go build -mod=readonly -o "$TOOL_ROOT/authorize-installed" ./tests/swa-cli-headless/authorize-installed
cp tests/swa-cli-headless/authorize-installed/public-case-preflight.mjs "$TOOL_ROOT/public-case-preflight.mjs"
"$TOOL_ROOT/authorize-installed" -h
"$TOOL_ROOT/authorize-installed" -mode prepare -package "$PRIVATE_PACKAGE"
```

本工具只支持Linux x64、Go 1.27.2。工具目录使用本次新建的0700普通目录，包和证据为0600普通文件；临时目录放在仓库的ignored `.verification`中。consumer从真实二进制旁查找，并将实际SHA写入准备包；不能依赖原ignored开发目录中的副本。包中安装VERSION、源码commit、driver各依赖、consumer、二进制及装内Node/swa均需准确固定。`prepare`另写`.prepared.json`，不能覆盖已有包或证据。

`profile`仅支持两种范围：

| 包字段 | 默认Codex：不填profile | StepFun Pi：profile为stepfun-pi-n3 |
| --- | --- | --- |
| caller | swa-codex-n3-20261022 | swa-stepfun-pi-n3-20261023 |
| agent / model | codex / gpt-5.5 | pi / stepfun/step-3.5-flash |
| workers | 3 | 3 |
| 来源目录字段 | codex_api_key_home | pi_source_home |
| 项目名 | SWA Codex N3 public stats | SWA StepFun Pi N3 public stats |
| 空间名 | SWA Codex N3 public stats 20261022 | SWA StepFun Pi N3 public stats 20261023 |

两种包均只接受既有公开项目`/workspace/swa-public-case-20261022/public-project`及原四文件、固定task/input/followup。未知profile、错caller/agent/model/N或另一profile的非空来源字段在HTTP之前拒绝。来源必须为显式、独立、当前用户0700普通目录；工具只检查其路径元数据，不读取auth、key、models或settings正文，不创建或修复来源。缺少来源、真实启动记录或服务身份时包保持INCOMPLETE，不能执行。Pi的API为openai-completions，endpoint为https://api.stepfun.com/v1，ToolsOff合同为pi-cli-1.0.1，新账户kind为pi-native-api-key-per-session；这项固定目标不表示账号、推理、原生ToolsOff或N已通过。

原serve owner须在真实spawn时保存非秘密启动记录并提供原ready/proof与准确SHA/runtime。Codex记录严格为installation、source_revision、data_root、endpoint、serve_pid、codex_api_key_home六键；Pi记录最后一键为pi_source_home。工具不生成或补造记录。Pi驱动只接收已核记录的`--pi-launch-record`和`--pi-launch-sha256`，从中独立核对SOURCE。真实serve捕获来源的入口是PI_CODING_AGENT_DIR，由原owner负责；工具不启动serve或调整PATH。

Codex固定原六个执行依赖；Pi在原六源基础上增加pi-event-view.mjs与pi-source-observer-entry.ts，全部逐安装VERSION的Git blob和实际SHA核对。缺少新Pi依赖的历史安装不能作为Pi执行候选。源码测试、编译与help只能证明离线范围。

未来委托的效果包括本次human会话、实际项目/空间登记、observe及execute两阶段settings CAS/Grant/token、原驱动业务请求和明确清理。cookie/CSRF/token只保存在内存，token只通过子进程环境传递；key不进入本工具。human写结果未知时停写且不重放，不能用CLI请求账本宣称human写已结算。业务未知只查原R。撤销与human未知后的撤销分别由包中明确委托控制；服务由原serve owner正常EOF并取得全树关闭证据。

driver的外部审阅快照和私有证据保存在EvidenceRoot。公开项目下另有固定公开正文副本及私有result/checked-result导出；只有批准的四原文正文可进入模型，私有结果不作为后继输入或上传。来源默认设置与每Worker三项设置分开。原生metadata中的cost=0仅为客户端默认值，不表示远端账单为零。

退出码：0表示相应工具步骤结束，真实业务仍需独立审阅；2表示参数或本地前置拒绝；3表示缺少必要前置；执行期间真实失败保留原UNKNOWN和安全错误分类。当前实际HTTP、人工授权、原生启动、模型及N验收结果以各自独立报告为准。
