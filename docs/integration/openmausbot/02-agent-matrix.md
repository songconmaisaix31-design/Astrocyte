# Agent 来源、双平台目标与实际能力

本次用户要求适配大部分Agent，产品目标为Windows x64和Linux。当前依据是本地集成树`b7447f4a`中的Orca 1.4.230源码。模板、结构化引擎、安装、登录、可启动、工具政策、完整成果和所属退出是不同层次；本表不按名称数量宣称已适配多数，也不把需求改写成某个供应商数量或凭据数量。

## 实际来源

来源根`U = third_party/orca-v1.4.230/src`。以下路径均实际读取或由原TypeScript编译器解析，身份与命令取源文件内容：

- `U/shared/tui-agent-display-names.ts`导出的`ALL_TUI_AGENTS`含45个ID；`U/shared/tui-agent-config.ts`的`TUI_AGENT_CONFIG_SOURCE`也有45项。
- `U/main/runtime/structured-agent-runtime-metadata.ts`和`structured-agent-runtime-registrations.ts`注册7个结构化引擎：Pi、Codex、Claude，以及`U/main/acp/acp-launch-specs.ts`中的Grok、OpenCode、OMP、Antigravity。
- SWA基线`runtime/swa-orca/src/agent-discovery.ts`只枚举这7个结构化注册；配置观察分支为Codex/Claude/Pi/OpenCode，获审NativeLaunch联合及关闭工具准入当时仅Pi/OpenCode。TUI模板名称不因此自动成为SWA可执行引擎。

本轮先行复用上述7个实际原生适配器，并逐项评估其他TUI模板的可复用启动、输入、取消、会话和成果入口。缺少结构化接口的Agent继续按其真实终端能力评估，不为凑支持数量另造任务权威或把终端预览当完整成果。

## 45个TUI模板清单

表中命令是上游检测标识，保持aliases与ID分开；不是本轮直接执行的命令建议。`未声明禁用`仅指该模板没有`detectUnsupportedRuntimes`，不表示Windows/Linux已运行或上游二进制兼容性已验证。`结构化`列表示Orca已有相应注册，其余为当前TUI来源候选。

| ID | 检测命令 / aliases | 额外启动或检测条件 | 原生输入模式 | 结构化 | Windows / Linux模板声明 |
| --- | --- | --- | --- | --- | --- |
| claude | claude | 原CLI与账户隔离另核 | argv | 有 | 未声明禁用 / 未声明禁用 |
| claude-agent-teams | orca / orca-dev、orca-ide | 还需claude；launch为orca claude-teams，平台命令名由源码处理 | stdin-after-start | 无 | 明确禁用win32及WSL / 未声明禁用 |
| codebuddy | codebuddy / cbc | 原CLI | argv | 无 | 未声明禁用 / 未声明禁用 |
| openclaude | openclaude | 原CLI，草稿prefill | argv | 无 | 未声明禁用 / 未声明禁用 |
| codex | codex | 原CLI与账户隔离另核 | argv | 有 | 未声明禁用 / 未声明禁用 |
| autohand | autohand | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| ante | ante | bare TUI；--prompt是一次性模式 | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| trae | traecli | 使用TRAE CN标识，prompt前有-- | argv | 无 | 未声明禁用 / 未声明禁用 |
| opencode | opencode | 输入就绪信号与ACP分别核查 | flag-prompt | 有 | 未声明禁用 / 未声明禁用 |
| opencode2 | opencode2 | launch追加--standalone；不能冒充opencode注册 | flag-prompt | 无 | 未声明禁用 / 未声明禁用 |
| mimo-code | mimo | 输入信号沿上游模板，未独立验证 | flag-prompt | 无 | 未声明禁用 / 未声明禁用 |
| pi | pi | ORCA_PI_PREFILL草稿入口 | argv | 有 | 未声明禁用 / 未声明禁用 |
| omp | omp | ORCA_OMP_PREFILL草稿入口 | argv | 有 | 未声明禁用 / 未声明禁用 |
| prime-agent | prime-agent | prompt前有-- | argv | 无 | 未声明禁用 / 未声明禁用 |
| qoder | qodercli / qoder | 原CLI与trust另核 | flag-prompt-interactive | 无 | 未声明禁用 / 未声明禁用 |
| qoder-cn | qoderclicn / qodercn | 原CLI与trust另核 | flag-prompt-interactive | 无 | 未声明禁用 / 未声明禁用 |
| gemini | gemini | 原CLI | flag-prompt-interactive | 无 | 未声明禁用 / 未声明禁用 |
| antigravity | agy | 此为TUI，不是下表ACP服务器 | flag-prompt-interactive | 有，使用不同ACP命令 | 未声明禁用 / 未声明禁用 |
| aider | aider | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| goose | goose | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| amp | amp | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| kilo | kilo | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| kiro | kiro-cli | launch为kiro-cli chat --tui | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| crush | crush | 显示名Charm | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| aug | auggie | ID aug与命令auggie不同 | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| cline | cline | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| freebuff | freebuff | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| codebuff | codebuff | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| command-code | command-code | launch追加--trust；不能与cmd.exe检测混淆 | argv | 无 | 未声明禁用 / 未声明禁用 |
| continue | cn | continue是shell关键字，检测使用cn | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| cursor | cursor-agent | 原CLI与trust另核 | argv | 无 | 未声明禁用 / 未声明禁用 |
| droid | droid | 原CLI | argv | 无 | 未声明禁用 / 未声明禁用 |
| kimi | kimi / kimi-code | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| mistral-vibe | vibe / mistral-vibe | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| qwen-code | qwen | Windows paste后Enter有原生重试延迟 | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| rovo | acli | launch为acli rovodev run | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| hermes | hermes | launch追加--tui | hermes-query | 无 | 未声明禁用 / 未声明禁用 |
| openclaw | openclaw | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| copilot | copilot | 互动模式与一次性prompt不同 | flag-interactive | 无 | 未声明禁用 / 未声明禁用 |
| grok | grok | prompt前有--；ACP使用agent stdio | argv | 有 | 未声明禁用 / 未声明禁用 |
| muse | muse | launch追加--trust-workspace | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| dsh | dsh-tui / dst | 还需dsh；launch为dsh-tui . | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| zcode | zcode | expectedProcess为zcode-cli | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| devin | devin | bare REPL；argv prompt会直接提交 | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |
| jcode | jcode | 原CLI | stdin-after-start | 无 | 未声明禁用 / 未声明禁用 |

## 七个结构化引擎的实际复用入口

下列路径以`U/main/`为前缀。`startSession/send/cancelTurn/closeSession`等是源码适配器入口，不是SWA已对所有引擎开放的公共动作。各引擎都要分别获得当前项目批准、真实账号、关闭工具政策、同session输入/取消及完整成果验证。

| 引擎 / 原生传输 | 启动与输入/取消入口 | 账号与ToolsOff现状 | 会话与完整成果入口 | 平台源证据 |
| --- | --- | --- | --- | --- |
| Pi / jsonl-rpc | `pi/rpc-runtime-registration.ts`、`rpc-launch-resolution.ts`、`rpc-session-adapter.ts`；cancelTurn发abort，stop结束session | `PI_CODING_AGENT_DIR`；SWA已有固定1.0.1的`native-pi-policy.ts`/`private-pi-account.ts`，需真实账号/API/模型匹配 | `pi/pi-transcript-import.ts`、`pi-session-history.ts`及SWA `task-result.ts`的自产JSONL/journal检查 | `supportsSupervisedProviderChildLocation`允许原生local非WSL；Windows另需进程创建时间proof；SWA完整custody平台实现/实测仍各自核 |
| Codex / app-server | `codex/codex-structured-launch-resolution.ts`、`codex-structured-session-adapter.ts`及`codex-structured-session-teardown.ts` | `CODEX_HOME`；原`codex-structured-permission-policy.ts`的workspace-write/approval并非ToolsOff；SWA基线未有可执行NativeLaunch政策，需02补审 | `codex/codex-cli-conversations.ts`、原生thread/journal翻译与SWA成果消费者待接入 | `supportsCodexStructuredLocation`允许原生local非WSL；实际Windows/Linux账号和完整退出独立检查 |
| Claude / Claude原CLI的structured SDK桥 | `claude/claude-structured-launch-resolution.ts`、`claude-structured-session-adapter.ts`；cancel/close由原adapter调用真实CLI | `CLAUDE_CONFIG_DIR`；权限模式/canUseTool回调不自动证明模型请求无工具；SWA基线政策与账号私有化待接入 | `claude/claude-cli-conversations.ts`、原journal与SWA成果消费者待接入；不能拿预览冒充完整 | `supportsClaudeStructuredLocation`允许原生local非WSL；不以当前机器无CLI推断平台不支持 |
| Grok / ACP | `acp/acp-launch-specs.ts`使用grok agent stdio；`acp/acp-structured-session-adapter.ts`的输入/interruptAcpTurn/closeSession | `GROK_HOME`；native spec可选择API key或cached_token；SWA基线配置、私有账号、ToolsOff政策未完成 | `acp/acp-dialects/grok-dialect.ts`、`grok/grok-transcript-import.ts`、`grok/grok-session-history.ts`；实际SWA结果政策待接 | ACP local非WSL且Windows创建时间proof；原spec未作Windows/Linux排除，安装/启动仍待实测 |
| OpenCode / ACP | `acp/acp-launch-specs.ts`使用opencode acp；原线1.x>=1.18.31或2.x>=2.0.14；输入/取消共用ACP adapter | SWA已有固定2.0.31/DeepSeek-pro的native-opencode-policy/private-opencode-account/opencode-policy-plugin；不能推广为任意版本/提供商已验 | `acp/acp-dialects/opencode-dialect.ts`、原生私有SQLite/journal与SWA `native-opencode-import.ts`/`task-result.ts` | ACP local非WSL，Windows创建时间proof；SWA基线launcher仍Windows专用，Linux后继需独立custody与实际结果 |
| OMP / ACP | `acp/acp-launch-specs.ts`使用omp acp；版本>=17.0.5；输入/取消共用ACP adapter | `PI_CODING_AGENT_DIR`默认.omp/agent；模型仅session内发现；SWA配置、隔离和ToolsOff政策待完成 | `acp/acp-dialects/omp-dialect.ts`、`omp/omp-transcript-import.ts`、`omp/omp-session-history.ts`；SWA完整成果待接 | ACP local非WSL，Windows创建时间proof；未声明平台禁止不等于本次通过 |
| Antigravity / ACP | `acp/acp-launch-specs.ts`使用独立agy_acp_server.par；Linux传--uid=；不是agy TUI | `antigravity/antigravity-chat-account.ts`的专用OAuth目录，与agy登录不同；ToolsOff、模型发现和SWA政策待核 | spec的transcriptImport为null；sessionHistory不从agy参数恢复；完整成果/重启恢复不能据原有CLI路径推定 | `antigravity/antigravity-acp-runtime.ts`固定1.3.0校验pin有Linux x64/arm64及macOS，明确没有Windows pin，当前Windows ACP不支持 |

ACP共同定义`acp-structured-agent-definitions.ts`的`approvalEnforcement='orca'`表示处理权限请求；不是禁用所有工具。Grok的`--always-approve`仅与fullAccess有关，OMP/OpenCode也有权限请求处理，均不能拿“自动拒绝/批准”替代ToolsOff原生证明。Antigravity的yolo/default模式同样不证明无工具。

`provider-process/supervised-provider-child-location.ts`要求local、非WSL，Windows还核真实进程创建时间。Codex/Claude通过`shared/execution-host.ts`的原生local判定。SWA产品层还必须接入对应平台的真实完整所属树控制，不能把上游adapter的supportsLocation当作产品安装或停止验收。

## 每个引擎的实际覆盖记录

本卡只进行了源码与结构化定义核查。当前云机为Linux x86_64；主控已只读核到Codex 0.163.0及login reported，API_KEY未设置不能据此称缺登录。账号值未读取或复制。Codex产品私有启动政策、ToolsOff、模型调用、session接续、完整成果和所属树退出仍须由02/06实际执行后记录。其他六个引擎本卡没有进行登录或模型调用，不能把它们填为已安装/已认证/已启动。

后续逐次记录每个引擎的`installed/version`、`login/config`、`NativeLaunch/ToolsOff`、`start/send/read/stop/resume`、`完整result`和`process custody`的PASS/FAIL/NOT_RUN或unknown及来源；每项分别标Linux和Windows。用户确认Windows x64是目标体系架构，当前未连接Windows主机，其实际候选与链路均为NOT_RUN，不重复询问架构，也不声称可远程执行。

模型测试预算不设上限解除本轮费用额度限制；它不产生缺少的供应商账号、Windows机器、原生支持、工具权限、私有资料权限或外部发布许可。固定公开输入、ToolsOff、原request幂等、效果未知、超时/并发和机器资源准入继续有效。后续实现缺口交02，依赖/原生构建和平台候选交05，命令体验交03，真实验收交06；本轮不以任何比例宣称已经完成“大部分Agent适配”。
