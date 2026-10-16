# Works Relay integration report

## 2026-10-16 根接管集成

按用户“继续集成，集成交给你了”，根在原 `works-relay-integration`、原 `songconmaisaix31-design/works-relay-integration` 分支直接负责普通合并、检查与提交。原 I 两次 Codex 启动在 `account/read` 路由发现超时后实际退出，失败原件保留；没有替代 Task / Run 或活跃 I 写入者。领域业务修改继续由原作者完成。本节为当前集成入口，后续章节保留各阶段准确来源和结果。

从干净 `38fb289160ef85f75062f917e861488b95e2a309` 普通合并以下准确提交，均无冲突：T REPORT `149b5261f52e5a33a2172241ce2089ae12fbbcb3`（SOURCE `5d701bdd7f4d6696a1f730f21c47f84d222dd6cd`）、H REPORT `f40680bca1391f7f00b937d8f3bfb43aa0cda1a4`、H 运行恢复 REPORT `59e114a9aa5732a41e1ed9f5773452dac0ecd0cf`、B REPORT `9fa97efe8208de85a2251f1ea01008c79ad04945`（SOURCE `dba9f94d31ab1827fa90f5923e68b77144e8b3ec`）。领域未提交修改均未复制。

| 准确组合与命令 | 实际结果 |
|---|---|
| `7120a207ee6da0923262bbc6347661d92ac7b285`：`go test -mod=readonly -p=1 ./tests/s1 -run '^TestProductionCompletionHTTPMissingSourceSurvivesProcessRestart$' -count=1 -v -timeout=15m` | EXIT0，真实生产入口、HTTP、SQLite、实际关闭和重启；来源缺失返回404，原数据库与 Relay store 保持、原生会话和经验记录为零。单项20.73秒，包21.679秒；原件 `D:/Astrocyte-agent-assets/20261015/integration/root-completion-7120a20/completion.log`。仅验证真实拒绝和持久化，不作为 Interpreter 正向结果。 |
| `f9453c455e13b412e7af252b60e313552d17ec6f`：`go test -mod=readonly -p=1 -tags=repository_context_live ./cmd/server -run '^(TestListingPythonExplicitInputsFailAtConfiguration\|TestListingPythonUnconfiguredCapabilityRemainsAbsent\|TestVideoSourceConfigActualInstalledTools)$' -count=1 -v -timeout=3m` | EXIT0，三主项、七项显式无效配置子项通过，包2.000秒。使用原 B 真实媒体配置构造生产 Reader、核实际 Python 路径和显式禁用状态，未执行提取、ASR、模型或共享业务操作。原件 `D:/Astrocyte-agent-assets/20261015/integration/root-media-f9453c4/media.log`。 |

两次检查前后源码相同，子进程缓存、空身份目录和输出均在 D 盘；`git diff --check` 通过。实际共享服务仍使用 H 的准确 `38fb289` 稳定副本、原 SQLite 和 `http://127.0.0.1:64505/`，上述后继检查不声明运行实例已更新。原 `38fb289` 完整构建成功记录保持；本阶段尚未执行后继完整构建、完整适用检查和最终发布包。

P 的准确准备 REPORT `b208ab8ac696fea43ad192648a2e273392f5c99a` 已完成只读合并预检查，实际冲突为 `HANDOFF.md`、`STATUS.md`，尚未实际合并。发布入口的 S4 / S5 显式配置、C 盘空间守卫及 Q 源码祖先仍须处理。Native 权限、语言、具体候选批准及其他用户待定项按根当前交接与 QUESTIONS 保持。

## 2026-10-16 普通产品集成与构建

本阶段沿根 04:04 HANDOFF / STATUS、发布模块 04:08 计划及当前经理消息执行。唯一 I 为 Task `task_2ee44933df72` / Dispatch `ctx_4a4234c26f8c` / 终端 `term_9f0b076a-c973-4c86-b073-2c17640750e9`，公开终端核实客户端 Codex / GPT-6.1-Sol xhigh fast。固定工作树 `works-relay-integration`、分支 `songconmaisaix31-design/works-relay-integration`，接管准确 HEAD `8b0eb882969d972f3f419602c12b5e20f00b79e2`，工作区干净。当前优先同一普通产品运行及核心操作；完整默认、L1 / L2 和最终发布继续按实际结果验收。

### 准确来源与普通合并

仅消费 M-release 已确认的提交，全部普通 `git merge --no-ff --no-edit` 无冲突完成，没有复制原作者 WIP。直接文件编辑仅本报告；领域源码和检查脚本均来自原作者提交。

| 原作者输入 | 准确 SOURCE / REPORT | 组合 SOURCE |
|---|---|---|
| B 项目设置提交后唤醒清理，与既有 S2 AfterWorkspaceCommit 配对 | SOURCE `b68c3537e9bd9ffe3a06d2d8d513408f89d14de3` / REPORT `0e6139f3236c32787719fd1ce23e834ed68723c3` | 先普通消费 B，再消费下一行 F，得到 `cffac3671f21dbf90df9eaa2804a2eb5e070d332` |
| F 固定上下文预算及真实事件身份断言 | SOURCE `90e69e703f79b81653ea5ccc1e6c9eca4bd781f9` / REPORT `a166310a3c7d7b22806ec74aa4125e9f6c6a2222` | `cffac3671f21dbf90df9eaa2804a2eb5e070d332` |
| F 同一产品检查辅助脚本，七件 web/e2e 文件 | SOURCE `8bd599c8a34661dc9f4a47398212b387278a483f` | `79b2de5e318e900269230bfc5b12c1a5aea9f16b` |
| F 分离旧记录与新输出目录，增加只读 Stored 核查 | SOURCE `6503077885aee63dd97f9642deab66ba57167925` | `76030ec7846d41b709a0f21d6d3761ba4d7bcae0` |
| F 当前持久化恢复的独立报告 | REPORT `da692e8dd1a29e5621d127a4bea4fbc9ef7ab21a` | `deec6c7b27510f2139a910508a268f6d2585dff2` |
| B 生产启动检查的阶段与子进程原始错误 | SOURCE `fbb32d6520f501d8b1cd5ea38d978e56ac16d41d` / REPORT `c563ff9498d4e54f095a1bddf3c46d2cd869ea85` | `1c27477272c8d6e8aeb6dac5302bacb65fef614a` |
| H 清理生命周期、完成来源转发和独立认知根 | SOURCE `3569f5da9d4df423916e3cff618f301a070d87cc` / REPORT `b94f2bef2e529de23aaf89f34704d342ffe5c8d9` | `b73dc79e7c59a9192ec2bfbd3625a620d8304efb` |
| R 原控制器缺失时保留 delivery_unknown | SOURCE `ce541794d58560747124fea2cf9d08222700587e` / REPORT `4f7a041e489ac6d3c56918314ba24679e50e4d3f` | `b36c420c51d80cbc2f0d3bd188d6355e6c1467b2` |
| B OpenCode 明确输出目录与 Works 实际 CanonicalRoot | SOURCE `a8d361b1cb4133713e25f1d6c20ea4bfd22d844d` / REPORT `8179cf9c5074ab26a053fe7f1bfd691b153b4084` | `7401b302275c7ef6c8dc1dcca133ace263527a2f` |
| H 安全 HTTP 失败诊断 | SOURCE `a77f9b12e84147b8de95f9179d9e03aa170b27e6` / REPORT `a4a9ce37464d1e12880dc65b4e5348c50504131e` | `3edd1f3cb9c3a95ee2c06446af9667e1f3f9a24f` |
| B 完成经验 HTTP bridge 与固定请求身份 | SOURCE `faee82ec7a3f4e0c832bb198d1c42393ae88ea02` / REPORT `921e30b56683118a0d22a29cb63161b599395d3a` | `2128fd3177ff4d567eb1f0871945d3ec1cdea121` |
| F 过期 Grant 与安全固定上下文检查 | SOURCE `1e242e1573b72046adf572bb0874c6f870730636` / REPORT `38acb723bfa04fcb7eed8a69347b8e6c958f6cf7` | `9686801b0bc9843a268d3e8c8f816c16e991011a` |
| UI 明确判断参数、原 Goal 查看与固定引用 | SOURCE `f0d27271c15d131aea50012658232f7fa41a142b` / REPORT `7467ddb17ab204172d9656e47c5f82ec0a97d9b2`，含已核前阶段 `0ad83a96bc50747867146a38a80b7a90ecbedc5f` | `b2333d18858a55f008953d69ecda9154ebd4a641` |
| B Origin 标准库解析 | SOURCE `bd18df4c57f41b3c880bb47fba98c0deec5135ce` / REPORT `517b3073c803f88c2d7798e3da4ceb876fba03f5` | `5322e7de66945067601868a82b4b7d790f236fef` |
| H 三项真实检查使用明确输出根 | SOURCE `237a95719d18b69063def0323cf35cc3a93b8232` / REPORT `a5c05cffdb27c8baba33337bb2dd0c5345ea817b` | `6f613692b8ec3de5f33b9ff56244734bcb24d56f` |
| S4 真实生产清理检查与最终准确 b73 结果 | SOURCE `6e441f7a6c14a11492d188a7259106445715700a` / REPORT `7f9c7fa9fc3b23153eb462931924a7782775cacb` | `5996a3facbded36a6c2d5c5afff68579806eb1a1` |
| F 同代认知与全部事件只读核查 | SOURCE `12a17f7bf7a9fbef273e872516961aa9262b7110` / REPORT `94d827a34f8cc659de8ab3e051483eb194480d76` | `48793bac010d9b160106488040491556d49d63c7` |
| B 人工准入的当前项目权限组件 | SOURCE `19ff699f304961d3c9fad6a336163848c923057b` / REPORT `bf8fda3c45b9198a7e7537355d5ba419f9bc87f8` | `8c09088caacaf010c7e541e6c5f432a5b74e5442` |
| T 完成来源拒绝与实际收据事务检查 | SOURCE `fa5c6e6a2cedcf34b22bf760c24ac6b81c3a7241` / REPORT `41cfefc24792470bd00d25b80b3235142d68e3ee` | SOURCE 合入为 `1284563dff2ff1a28a2ec32e60792b9ea2c48919`，REPORT 合入为 `3c72fa1a9c9f27b881571b7cf82da3edfda6f172` |
| F 三段原 UI 内容相容消费 | SOURCE `a871d62c6b5ed6b708dfc1b70f938741f0b44081` / REPORT `a4e3dc39350911268fd5cde4f177d2362928e765` | `8ac0e6b7eaf868d198d75071f73820224b1e68d4`；普通合并实际仅增加原 F 报告 13 行，业务差异为零，未重复目标或构建 |
| B 当前固定资料引用权限与原 PDF 处理 | SOURCE `089a3d276fa186cf41cc8384fb2bc9dad91d9383` / REPORT `c0bfe00456329249e279a4353abe634381656806`，含 SOURCE `25a4ae879ca0a4cd059540bfa5aeddf8cf713dcc` / REPORT `ce4e12aca2f272f7ed5edb48b9cf76fdccddad5e` | `f1fdc35d1b5386749cd10b2f636fa47a5581846c`；两个 agents 文件仅中文注释，其行为差异为零 |
| B 历史公开进度的实际存储与人工 CAS 检查 | SOURCE `82426bab2c405e1a1a2184801a60fa4e98985c2b` / REPORT `c0a7f837fb06f9858e996597e45900dc3c4c4a47` | `fd07ac5f809e9cba8db7efe568112d84da9a61a3`；生产文件仅中文注释，实际新增行为为原测试的真实来源消费 |
| H 准入、完成经验生产入口与完整公开 Schema | SOURCE `25b341dd5a6b74cc2588d12544eb955235f9274c` / REPORT `c7be470c4e72fa36128a8db9c230d324dfe8c1f8` | SOURCE 合入为 `609993b25543a74e64d89a31c99bf3c08b81b44b`，REPORT 合入为 `5ac02b0eca9193ee869081e909cb21e117ee56c6` |
| F 真实候选已保存操作的恢复检查 | SOURCE `fd7db97d26991ec04937bfb7ecae3208ad339706` / REPORT `3320e0a18b17edc6b6bdba9e4b08ccef6d4966af` | `ecb29c9593348d409514cc1ddff05ff04affcd87`；三脚本和原报告均普通消费，I 没有执行候选 mutation |
| B Windows junction 路径辅助与真实身份检查 | SOURCE `d8067aae16acdec89e2bb8cb91ec153f45bf32f4` / REPORT `e850414e1b97b91a2e690cfdf62492577dd9cf00` | `b2673ffb4fc5f0d2a921593b0a4bcc503ea7ba56` |
| F 人工进度表单、当前权限与实际保存检查 | SOURCE `696828478eeaf128309079957234bb78aafbaaf9` / REPORT `963d880e30df4fff0923623c5efa711e6adfed2a` | `49e84c83764ebb35e08062dea37321f095b93508`；I 未重放原 F 已保存写入 |
| F 原 generator 生成准确客户端 | SOURCE `4e3c384032ab1671e226a819009d95c366d80361` | `fc172eb6e81ed398f8a21414221c5ef9bd66e3aa`；仅 schema.d.ts 的真实生成输出 |
| H 实际 Workbench 链接目录的明确输出与原路径检查 | SOURCE `c1750e867fb0dbe31dcf083909b0f84c8fca9c83` / REPORT `cd616d392396978a855da78495942c0b27f93350` | `3eed9fea2955e71fe0c0ddffdc74eb5408847dd2`；原 symlink 断言与 OS1314 失败保留 |
| V / S2 同原 Task 的旧视频只读审计报告 | SOURCE `d58ce7d5c724e5f165c545a2a1bb1cb9ebfee796` / REPORT `a640c181a6a1cc056be021385c23e71ab6913b6c` | `f73cd7c93f2d6e2b367a6c7b93222ae8b6a0a871`；闭包含同原作者 S2 文档登记祖先 `81eb88df11f2c69b71290d5bb9ad5b2e2e0f0d9f`，只有两份文档变化 |
| B MCP 收据序列化的直接失败与准确整数 | SOURCE `895ae6a3a879e159e5de2c8629eed3423398cf68` / REPORT `cb2b8297ae891ce79b59d8cd2648654440e02f89`，原 B 直接确认交付 | `019e732f54c5c22d3f49b32776b3471c092f6a78` |

B 的准确 REPORT 祖先同时带入已提交的 private prepared input 契约消费、token 撤销争用检查和 cleanup batch 配置；上述范围已核实际合并差异。B 原 native WIP 保持原树。F 的 inspect 写入脚本保留其原成功记录，本阶段没有执行这些脚本；同实例实际 HTTP / DOM / SQLite 操作仍由原 F 独占。每个组合 SOURCE 均已立即普通 push；`deec6c7` 的远端分支已独立读回同值。

### 实际验证

| 准确来源 / 命令 | 实际结果与原件 |
|---|---|
| 原 `8b0eb882` 完整 `pnpm check` 原件核查 | EXIT UNKNOWN。`c16-product-check-8b0eb882/check.log` 有 gofmt / vet 完成及四个 Go 包失败，最后输出为 httpapi FAIL；没有 `result.json`。后续包和契约、TypeScript、lint、Vitest 阶段无完成记录，未重新运行已知未修全套。 |
| `cffac367`：`pwsh -NoProfile -File D:/Astrocyte-agent-assets/20261015/integration/run-validation.ps1 -Step build -RunName c16-product-build-cffac367` | 原 `pnpm build` EXIT0 / 49.3755472 秒；server / MCP / Sleep / TypeScript / Vite 完成，source 与 source_after 精确同值。原件为该 RunName 目录的 `result.json`、`environment.json`、`build.log`，原 bundle 体积提示保留。 |
| `cffac367`：`go version -m D:/Astrocyte-agent-assets/20261015/integration/build/dist/astrocyte.exe` | 实际 revision 为 `cffac3671f21dbf90df9eaa2804a2eb5e070d332`，`vcs.modified=false`。 |
| `79b2de5`：`pnpm --dir web typecheck`、`pnpm --dir web lint` | 两命令均实际 EXIT0；日志为 `c16-product-static-79b2de5-typecheck.log`、`c16-product-static-79b2de5-lint.log`。 |
| `76030ec`：`pnpm --dir web typecheck`、`pnpm --dir web lint` | 新增四件后另行检查，两命令均实际 EXIT0，检查前后 SOURCE 同值；日志为 `c16-product-static-76030ec-typecheck.log`、`c16-product-static-76030ec-lint.log`。 |
| `1c274772`：原 runner `-Step build -RunName c16-product-build-1c274772` | 原 `pnpm build` EXIT0 / 31.0071399 秒，source / source_after 同值；server / MCP / Sleep / TypeScript / Vite 完成，实际 Go revision 同值、modified=false。原件为该 RunName 目录 `result.json`、`environment.json`、`build.log`。 |
| `1c274772`：`go test -mod=readonly -p=2 -tags=repository_context_live ./cmd/server -run '^TestRepository(ContextConfig(ExplicitAndUnavailable\|RejectsAmbiguousConfiguration\|RejectsActualUnavailableDependencies\|ActualNodeSyntaxRefusal)\|CleanupBatchTrustedConfiguration)$' -v -count=1` | 实际 EXIT0，5 项主检查及 8 项配置子检查，包 4.016 秒；实际 Node / SDK 后以原 Go 文件触发 Node 格式拒绝，阶段与原错误断言均通过。检查前后准确 SOURCE 同值，日志 `c16-product-config-1c274772.log`。 |
| `b73dc79`：原 runner `-Step build -RunName c16-product-build-b73dc79` | 原 `pnpm build` EXIT0 / 20.8431586 秒，source / source_after 同值；三个 Go 程序、TypeScript 和 Vite 完成。原件为该 RunName 目录 `result.json`、`build.log`。 |
| `b73dc79`：`go test -mod=readonly -p=1 ./cmd/server -run '^(TestRelayRepositoryCleanupActualSQLiteLifecycle\|TestRelayJudgmentProductionAssemblyActualSQLiteHTTP\|TestRelayReaderRootActualDirectoriesAndDatabaseAliases)$' -count=1 -v -timeout=120s` | 实际 EXIT0，三主项与两清理子项，包 4.358 秒；真实 SQLite / HTTP、同主库事务、清理取消等待与实际错误、双快照 hardlink 拒绝通过，原生会话 0。日志 `c16-product-assembly-b73dc79.log`。 |
| `7401b30`：agents 三项原控制器拒绝 / 未知保留 / 受限工具入场，httpapi 实际 Works 重开，events 实际来源身份目标 | 实际总 EXIT1；agents 三主项 PASS / 0.916 秒，Works 主项 PASS / 3.256 秒；events 合法单链接目标 FAIL / 1.530 秒，错误 `relay session capture: workspace_unverifiable`。日志 `c16-product-target-7401b30.log`，已退原 H / R。 |
| `3edd1f3`：原 runner `-Step build -RunName c16-product-build-3edd1f3` | 原 `pnpm build` EXIT0 / 37.1029359 秒，source / source_after 同值；原件为该 RunName 目录 `result.json`、`build.log`。 |
| `3edd1f3`：`go test -mod=readonly -p=1 ./cmd/server -run '^(TestRelayFactsActualHTTPSQLiteReadOnlyAndPermission\|TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown)$' -count=1 -v` | 新诊断实际 EXIT1 / 包 4.535 秒；facts 第76行为 `scope_denied: project root is unavailable`，listener 第198行为 HTTP403 / 期望200 / `public_code=scope_denied`。日志 `c16-product-http-diagnostic-3edd1f3.log`，已退原 H。 |
| `2128fd3`：`go test -mod=readonly -p=1 ./cmd/server -run '^TestCompletionExperienceHTTPActualProjectBoundariesAndRestart$' -count=1 -v` | 实际 EXIT0 / 包 2.336 秒；真实 HTTP / SQLite 的项目权限、缺失来源拒绝与重启通过。日志 `c16-product-completion-2128fd3.log`。没有正向完成来源、Interpreter 或模型结果。 |
| `9686801`：`pnpm --dir web typecheck` 与定向 ESLint `e2e/readProductProposal.ts` | 均实际 EXIT0，前后 SOURCE 同值；日志 `c16-product-static-9686801-typecheck.log`、`c16-product-static-9686801-lint.log`。 |
| `b2333d1`：TypeScript、全 web ESLint、`vitest run` 两个新增 policy / question 参数检查文件 | 三命令均实际 EXIT0；2 文件 / 7 测试 PASS，明确参数检查不作为真实模型或实际保存。日志 `c16-product-ui-b2333d1-{typecheck,lint,vitest}.log`。UI 普通合并实际 17 件，均为已确认 REPORT 的原作者祖先，新带入正式 validation 历史为 0。 |
| `5322e7d`：`go test -mod=readonly -p=1 ./internal/adapters/httpapi -run '^(TestLoopback\|TestIsLoopbackHost\|TestRequestIDMiddleware)' -v -count=1` | 实际 EXIT0 / 包 1.781 秒，6 主项与 16 个新增真实 HTTP Origin 输入通过；日志 `c16-product-origin-5322e7d.log`。 |
| `5322e7d`：原 runner `-Step build -RunName c16-product-build-5322e7d` | 原 `pnpm build` EXIT0 / 34.1603979 秒，source / source_after 同值；server / MCP / Sleep / TypeScript / Vite 全部完成。原件为该 RunName 目录 `result.json`、`build.log`。 |
| `6f61369`：原 D 环境只复验 facts / listener / source identity 三项首次失败 | 实际总 EXIT0，cmd/server 两项 PASS / 包 2.376 秒，events 一项 PASS / 包 1.585 秒；原生产 guard、全部断言与 Skip 保持。日志 `c16-product-root-repair-6f61369.log`，前后 SOURCE 同值；`7401b30` / `3edd1f3` 首次 RED 仍保留。 |
| `48793ba`：TypeScript 与定向 ESLint `e2e/readProductCognition.ts` | 均实际 EXIT0，前后 SOURCE 同值；日志 `c16-product-static-48793ba-{typecheck,lint}.log`。没有再次执行 F 的实际认知读取。 |
| `8c09088`：`go test -mod=readonly -p=1 ./internal/adapters/admissionbridge -run '^TestHumanAdmissionActualProjectAndTransaction$' -v -count=1` | 实际 EXIT0 / 包 1.585 秒，当前项目、同主库事务零额外连接、错误库事务、取消、角色及 ID 拒绝与零业务记录通过；日志 `c16-product-admission-8c09088.log`。 |
| `1284563`：`go test -mod=readonly -p=1 ./internal/attention/app -run '^(TestCompletionCaptureActualProjectMissingSourceAndReopen\|TestCompletionRepositorySharesActualCommandReceiptRollbackAndReopen)$' -v -count=1` | 实际 EXIT0 / 包 2.194 秒，两项真实项目 / 来源缺失拒绝、人工收据事务 / 约束回滚 / 重开通过；日志 `c16-product-completion-boundary-1284563.log`。没有将人工命令收据作为实际 completion 正向。 |
| `f1fdc35`：更新后的 `TestHumanAdmissionActualProjectAndTransaction` | 显式提供已确认公开 `paper-dom-actual-recovery.snapshot.json`，只核该原件存在及 59,717 字节元数据后执行原目标；实际 EXIT0 / 包 1.174 秒，当前 fixed refs、revision / lifecycle / 撤销、同事务权限通过。日志 `c16-product-admission-refs-f1fdc35.log`，前后 SOURCE 同值。 |
| `fd07ac5`：`go test -mod=readonly -p=1 ./internal/adapters/sqlite -run '^TestProgressStorageRoundTripAndCAS$' -v -count=1` | 明确提供已批准的两项公开输入，实际 EXIT0 / 包 2.489 秒；只读原公开验证库的 progress / project 派生记录，在自有 SQLite 验证真实迁移、完整往返、CAS 拒绝和人工 service 更新。日志 `c16-product-progress-fd07ac5.log`。历史 accepted receipt 保留来源，未作为本轮 infer / 模型结果。 |
| `609993b`：原 runner `-Step build -RunName c16-product-build-609993b` | 原 `pnpm build` EXIT0 / 30.1152367 秒，source / source_after 同值；server / MCP / Sleep / TypeScript / Vite 全部完成。实际 Go revision 同值、modified=false；原件为该 RunName 目录 `result.json`、`build.log`。 |
| `609993b`：`pnpm check:contracts` | 实际 EXIT1：Redocly 成功且保留原 13 warnings，264 examples 通过；最后原 `generate-client.mjs --check` 报 `Generated client schema drift`。日志 `c16-product-contracts-609993b.log`，已交 M-product / 原 F 按准确公开契约生成类型，I 未生成或手改客户端。 |
| `609993b`：`go test -mod=readonly -p=1 ./cmd/server -run '^(TestRelayJudgmentProductionAssemblyActualSQLiteHTTP\|TestRelayHumanAdmissionsWithoutJudgmentSnapshotsActualSQLiteHTTP\|TestRelayStaticActualMissingAndInvalidPaths)$' -v -count=1` | 实际 EXIT0 / 包 4.761 秒，新增两模式真实准入、当前权限拒绝、未分配草稿、读取与撤销、实际空完成源404及静态路径错误通过；日志 `c16-product-main-609993b.log`，前后 SOURCE 同值、没有原生或模型输入。 |
| `ecb29c9`：TypeScript 与三个新增候选脚本的定向 ESLint | 均实际 EXIT0，前后 SOURCE 同值；日志 `c16-product-static-ecb29c9-{typecheck,lint}.log`，I 未重放成功或 unknown 写入。 |
| `b2673ff`：`go test -mod=readonly -p=1 ./internal/adapters/events -run '^TestReplayJunctionUsesActualWindowsPaths$' -v -count=1` | 实际 EXIT0 / 包 1.410 秒，带空格目录的同一 helper 实际创建 junction 并核文件身份；日志 `c16-product-junction-b2673ff.log`。没有执行旧 replayFixture 或将其当原生回放通过。 |
| `f73cd7c`：`pnpm check:contracts` | 原完整链实际 EXIT0，Redocly 通过且保留原 13 warnings，264 examples 及原 generator 漂移检查通过；日志 `c16-product-contracts-f73cd7c.log`，前后 SOURCE 同值。`609993b` 的首次 drift FAIL 保持。 |
| `f73cd7c`：`go test -mod=readonly -p=1 ./cmd/server -run '^TestRelayWorkbenchActualLinkedGenerationRejected/junction$' -v -count=1` | 实际 EXIT0 / 包 2.949 秒，实际空 SQLite 迁移、hardlink / junction 与受保护目录拒绝通过；日志 `c16-product-workbench-junction-f73cd7c.log`。原 symlink 子项的 OS1314 仍未通过。 |
| `019e732`：`go test -mod=readonly -p=1 ./internal/adapters/mcp -run '^(TestDeliveryErrorFormatRules\|TestAstrocyteRegistersFiveTools\|TestAstrocyteDispatchGetAttentionState)' -v -count=1` | 实际 EXIT0 / 包 1.540 秒，4 PASS / 0 SKIP；两项明确格式规则及真实 SQLite / 生产人工主题 / OS pipe 的工具注册与关注查询通过，日志 `c16-product-mcp-019e732.log`。没有原生或模型输入，未当作 postcommit 实际异常通过。 |
| 普通 push、`git ls-remote origin refs/heads/songconmaisaix31-design/works-relay-integration`、工作区与 diff 核查 | 当前组合 `019e732f54c5c22d3f49b32776b3471c092f6a78` 普通 push 实际 EXIT0，独立远端同值；各轮消费结束无未合并路径，适用 `git diff --check` 通过。独立报告提交另在交接消息登记。 |

本轮全部新增验证原件位于 `D:/Astrocyte-agent-assets/20261015/integration/`。构建沿用原 dist、web/dist、.verification 到 D 盘的 junction，Go 缓存与临时编译目录为已确认的 D 盘位置；只设置验证子进程环境。没有改变全局 HOME、身份配置或 TEMP。

原 `8b0eb882` 的失败逐包保留并已交 M-release 分派原作者：

- `cmd/server`：公开项目注册失败；HTTP 403 与预期 200；`approval_required`；原 `delivery_unknown`；真实 symlink OS1314。
- `internal/adapters/agents`：四项 OpenCode redirected ancestor；原控制器拒绝结果预期不符；selected-text 自产来源身份或启动观察不符。
- `internal/adapters/events`：Windows junction 命令 `Invalid switch`；合法单链接文件被拒绝。
- `internal/adapters/httpapi`：实际 Works 发现为 `no_project_git_marker`，未取得预期 Git 来源。

没有改动这些断言、skip、retry、timeout 或原 UNKNOWN 记录。原 F 完整 320 项中断记录仍为 204 PASS / 55 FAIL / 43 原 SKIP / 18 无结果、EXIT UNKNOWN，归原 F 的准确来源；本阶段未新运行完整默认 E2E。

### 运行交接与真实剩余项

`cffac367` 构建产物已通过 M-release 交原 H。H 本人确认复制到其稳定 `shared-product-ad8413/versions/cffac367`，复制前后实际 Go revision 同值、modified=false、index 存在，随后停止读取并归还 I 的产物窗口。后继 F 脚本和文档没有引起重复产品构建，`cffac367` 的 build 结果仅归该准确来源。

B 后继准确 REPORT `c563ff9` 已无冲突普通消费，其祖先包含原作者已提交 fixed native input SOURCE `759764900bc45ed4c58efaad79ae9d20d421b834` / REPORT `9a96c365021c826e90a6271d7a3f8571fb0c780f`。实际 12 件合并差异已核，不含原四件 experience WIP。原生 Prepare / 单次发送和正向模型没有因消费这些已提交代码取得实际通过记录。

`1c274772` 第一轮普通 push 的退出未捕获，结束后远端仍为 `7da8d4c`，该轮保持 EXIT UNKNOWN。确认没有在途同命令进程后，第二轮普通 push 实际 EXIT0，日志 `c16-product-push-1c274772-second.log`；随后的独立远端读回与本地精确同值。`1c274772` 必要生产构建通过后，已交 H 读取同一 I 产物目录，归还前保持不覆盖。以上配置检查显式使用 P 已准备的 `runtime-sanitized-third/runtime` 下 Node、官方 SDK 和桥脚本，没有启动额外模型、登记控制器或修改原依赖。

`b2333d1` 普通 push 的在途状态单独登记，随后取得同一 session 的实际 EXIT0。`5322e7d` 第一轮普通 push 实际 EXIT1，输出为 `remote rejected` / `failed`，独立查询远端仍为 `b2333d1`；第二轮普通 push 实际 EXIT0，日志 `c16-product-push-5322e7d-second.log`，随后的独立远端精确同值。没有修改网络、全局 Git 或采用强制推送。

同一产品 64505 的 Node / SDK probe 首 EXIT1 原件和原因 UNKNOWN 保留，三个同 Dir / 白名单环境 probe 后继 EXIT0 由原 H 独立登记。H 已在 F / UI 安全 ACK 后更新同一 D 盘 SQLite 为准确 `b73dc79`，回执为 PID4676 / 启动 UTC `2026-10-15T20:22:58.8651960Z`，Q 的原 `snapshot-1711995690` 独立根已配置；M-release 独立核服务 ready。F 保持唯一实际 HTTP / DOM 操作人，已在该来源核全部认知页及真实事件库的 155 个 event_id 六字段。该结果归原 F 准确报告，I 没有重复读取实际库或页面。H 已完成并归还 `b73dc79`、`3edd1f3`、`5322e7d` 和 `609993b` 产物窗口，I 当前可覆盖后继必要构建。准确后继客户端已普通消费并取得原契约全链通过，当前等待 F 的两个 Relay 标签源码后固定必要构建。

原 H 已提交明确输出根修复 `237a957`，I 在准确 `6f61369` 与原 D 环境复验三项实际通过。三项生产 guard、全部断言与 Skip 保持，原首次失败准确分列。OpenCode 四项真实控制器检查只保留原 B 报告的实际结果，本阶段 I 未额外启动这些控制器。原 runtime_http 的 approval_required、session_git 的 delivery_unknown、workbench symlink OS1314、events junction 命令 Invalid switch、selected-text 自产身份仍无 I 新通过记录。M-release 明确受限控制器与重复探针暂停，完整 check / 默认保持未通过或未执行，等待根明确可执行范围或实际条件变化；没有修改断言、skip 或原生程序。

H 已在原 main / Services 注入 Relay 完成来源 Reader、真实同数据库准入与完整公共契约。有 snapshots 的准入保留 H 的 selected refs / delegated owner 检查；无 snapshots 模式使用 B `089a3d2` 的当前项目与每条固定引用权限。原 `completionInterpreter` 仍未配置，`provider_unavailable` 和缺失来源拒绝保持，L2 正向与六步未运行。F 已在保护原人工进度修改后消费准确 `ecb29c9`，使用原 generator 交付 `4e3c384`，I 已普通合入并复验完整契约链通过。F 原全局类型检查首 EXIT2 为 Relay 标签缺少 `resources_released` / `input_result`；原 F 正在完成准确修复，I 未改类型、标签或重复已知失败的全局 typecheck / build。实际候选处于 ready_for_review，具体准入及批准仍由用户决定，原 F 创建 / 修改阶段的真实首次失败和成功写入各自保留。Jev 密钥、Laya 语言、Agent completion action 等待定项保持原决定状态；受限 Codex 工具关闭要求和 unknown 不重发继续适用。

S4 最终 REPORT 的生产清理 PASS 准确归 `b73dc79`：真实 main / SQLite / CodeGraph / AOCI、两个公开 worktree、授权撤销、两次重启、十秒清理 A 且保留 B 原代均有原作者实际结果及经理独立核查。先前 f027 清理失败和孤立 H 缺 S2 提交后通知依赖的启动失败仍保留；I 的 b73 已含该准确 S2 依赖。准入的两模式与原完整权限均以 H `25b341d` 的实际装配及上述准确组合检查为准。

最终完整 check / 默认 E2E、原生 L1、L2 六步和最终同源发布仍未通过。本次来源组合和构建属于普通产品阶段，正式 V4 / v0.2、17321、原锁、协议、日程、D1 / Q 及成功 OS / 模型结果保持原范围。

## 先前已结束的 Works Relay v1.2 集成记录

This report covers the integration engineering task only. Real model handover, terminal-only coding, the three-minute Today loop and formal zero-write/scheduled-date acceptance remain separate gates.

Owner I: Codex, Orca terminal `term_55146ecc-fe9f-40b4-99c0-31d12b8eb5aa`, observed client `GPT-6.1-Sol high fast`. Run `run_d577f0d3fee1`, Task `task_db3080f0a933`, Dispatch `ctx_ace29bd7baa1`. Fixed branch `songconmaisaix31-design/works-relay-integration`; starting root documentation commit `8d9de2733713675c975b2920e71e21f07da7dcdc`. No additional Agent was launched.

## Exact source and history

Early integration SOURCE `84bb3f95fc3bbd2d373a561336578ead706bb4b1` was pushed and handed to Root before broad validation. Integration needed no additional application entrypoint, route, import or generated-type glue. After the first Windows CI browser resource failure, Root explicitly added `.github/workflows/check.yml` to this same owner's write domain in `msg_316344b26348`. CI-resource SOURCE `b25f43af865ad07398c00cc2a9281367c1e0760e` was separately committed, pushed and handed off immediately. H's first fixture repair integrated as `bdfbbdbbd6cbd6567e57ef3ed06ac2b3a860e97a`; its subsequent Linux evidence-directory repair integrated as final code SOURCE `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`, again pushed and handed off before validation. H's final pure REPORT then merged as `2a4c58c20bb77f535fb2b3af2901af8d1a627aa4` with `[skip ci]`, leaving every non-Markdown file equal to `5dd4b72`. This owner's only direct tracked writes are that three-line workflow addition and this report; H repaired its own fixture and report.

The workflow uses `pnpm --dir web exec playwright test --workers=1` only on Windows, and the original `pnpm test:e2e` elsewhere. The entire resulting workflow matches existing reference `4727f7a02a9b0fdfa1f12917f388d08da1c5ed84`; that validation history was not merged and is not an ancestor. No spec, viewport, timeout, retry, skip or fixture was changed by I. From early SOURCE to `b25f43a`, the only diff is authorized CI resource configuration. From `b25f43a` to final SOURCE, the only changes are H's EOF fixture layout and H report; application/Web/old E2E trees are unchanged.

| Track | SOURCE | REPORT | Ordinary merge |
| --- | --- | --- | --- |
| R | `b19d83635fe10cde9d42642d7444f3771d02f513` | `4848a98eef45656f30326054f3da8298f61044b1` | `bfe03bb` |
| H | `c7c3255d0530b93312bb541a477b9ec47e0afc15` | `6e211dd3e3d92cecadd6a513f7bb3b41075f661c` | `4475988` |
| F | `b48a6e5008b4da087a5d19c0a091b487197e7494` | `fdcf21b25dc3e8c84d2aa39b772a381374b50135` | `84bb3f9` |
| H fixture repair | `19c83e39b65e9a0cebc5bad20f7dba3f6c82cd2a` | `91d27e5a25ce676988ab90ef52bb81d647f324c6` | `bdfbbdb` |
| H Linux evidence repair | `67c7d2a5a205f4df8cc824b8d9c8c20ba60a5821` | `9238e101be10a0f52392e25cf1177adef2bdd56b` | SOURCE `5dd4b72`; REPORT `2a4c58c` |

All ten exact commits are ancestors, including H's original SOURCE/REPORT and both repairs. Initial R/H/F business parity passed for 14/16/24 files. Final parity again passed for R, repaired H and F (14/16/24 files) against their current SOURCE; each REPORT differs from its SOURCE only in its own report, and the latest three owner reports match the final tree. The original H report remains an ancestor with only H's authorized appended repair sections in the latest file. No squash, rebase, force push or history replacement occurred.

Against accepted baseline `d46f9f4cf2fd6e13c8362234746105330356c160`, attention/engine/weights, cognitive schema and mapper, migrations, validation, snapshot scripts, `docs/EXPERIMENT.md`, dependency locks and Go modules have zero diff. Existing Playwright configuration is unchanged; the later CI workflow exception is exactly the resource glue authorized above. The only old E2E change is the expressly authorized five-value expectation and platform-label assertion in `works-live.spec.ts`; timeout, retry, skip and remaining assertions are unchanged. Independent parsed OpenAPI comparison also preserved all 101 old paths and 209 old schemas; 14 Relay paths were added. The shared Workbench data-dir helper is H's authorized extraction with the existing protection body retained.

## Validation and exact source attribution

Commands ran from the fixed worktree. Early local full check/build/E2E and actual registry experiments belong to `84bb3f9`; they were not relabeled after the CI configuration change. After H's Go fixture changes, `bdfbbdb` and final SOURCE `5dd4b72` each have separately recorded fresh isolated complete checks; final owner/frozen parity and the new full remote CI belong to `5dd4b72` and its report-only descendant. Web is unchanged, so its complete local E2E was not repeated. `GOFLAGS=-p=1`, `GOMAXPROCS=2`, `NODE_OPTIONS=--max-old-space-size=512` bounded local resources; these settings do not change checks or assertions.

| Command / check | Result |
| --- | --- |
| `git fetch origin`; ordinary `git merge --no-ff <exact REPORT>` three times | PASS; no conflict or additional glue |
| `git merge-base --is-ancestor <each SOURCE/REPORT> HEAD`; per-owner business/report diff | PASS; six ancestors and all owner file parity |
| Frozen-domain diff; old E2E diff; parsed old OpenAPI parity | PASS |
| `git diff --check` | PASS |
| `pnpm install --frozen-lockfile` | PASS, exit 0 |
| `pnpm check` | PASS, exit 0, 180.16 seconds; all 20 Go packages containing tests, vet/module verification, 29-package architecture, dependency checks, 289 contract examples, generator drift check, 14 S1 acceptance tests, TypeScript/lint and 107 Vitest tests |
| `pnpm build` | PASS, exit 0, 15.61 seconds; server/MCP/Sleep binaries and existing web build |
| `pnpm test:e2e --workers=1 --output <external-evidence>/e2e-first-results` | PASS, exit 0, 440.23 seconds; complete default suite, 268 PASS / 38 conditional SKIP, two unchanged 1920/1280 projects, no integration change to skip/retry/assertions |
| `node <external-evidence>/registry-main-second.mjs` | PASS; actual built main/server, real registry JSON, real TCP HTTP and UI; details below |
| `node <external-evidence>/registry-main-stable-fifth.mjs` | PASS; same real registry experiment plus stable dialog/control geometry at both viewports, original bounds retained; earlier supplemental REDs retained below |
| Local Windows race | NOT_APPLICABLE to the existing workflow: Windows/amd64, `CGO_ENABLED=0`; workflow runs `go test -mod=readonly -race ./...` on Linux only |
| First remote CI `38463553347` | RED on exact SOURCE `84bb3f95fc3bbd2d373a561336578ead706bb4b1`; Linux full check/race/build/E2E PASS (268/38), Windows check/build PASS but E2E 267 PASS / 1 FAIL / 38 SKIP; preserved without rerunning |
| Root diagnostic CI `38465806043` | CANCELLED/PARTIAL by Root as obsolete after resource integration; SOURCE `3387019f7c349957ef8f4c75d98ecd7fbbe42200` has verified ancestor and non-Markdown parity against `84bb3f9`, but this diagnostic is not PASS and does not replace final SOURCE CI |
| Resource SOURCE CI `38466602701` | PASS, completed/success on exact resource-fix SOURCE `b25f43af865ad07398c00cc2a9281367c1e0760e`; both platforms full check/build/E2E 268 PASS / 38 SKIP (Linux 4.3m / Windows 8.1m), Linux race PASS; Windows log confirms one worker. Separate pre-fixture-repair result, not final-source acceptance |
| Fixture SOURCE isolated `pnpm check` | PASS, exit 0 / 115.31 seconds, on `bdfbbdbbd6cbd6567e57ef3ed06ac2b3a860e97a`; fresh owned child LOCALAPPDATA and empty CLI metadata roots, compiler cache separately pinned; all 20 Go test packages and complete subsequent checks |
| Fixture SOURCE CI `38468510081` | RED, completed/failure on exact SOURCE `bdfbbdbbd6cbd6567e57ef3ed06ac2b3a860e97a`; Linux check failed at EOF fixture line 193 evidence write, later Linux race/build/E2E NOT_RUN; Windows full check/build/E2E PASS, 268/38, 8.5m, one worker. H's later correction remains independently attributed |
| Final SOURCE isolated `pnpm check` | PASS, exit 0 / 113.70 seconds, on `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`; fresh owned child state and empty CLI metadata roots; all 20 Go test packages and complete subsequent checks |
| Final SOURCE [CI38469944551](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/38469944551) | PASS, completed/success, exact SOURCE `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`; Ubuntu/Windows full frozen install/check/build/default E2E each 268 PASS / 38 SKIP, 4.9m / 9.5m; Linux race PASS, Windows race conditional SKIP; actual logs confirm 2 Linux / 1 Windows workers |
| Root independent `pnpm check` on `3387019f7c349957ef8f4c75d98ecd7fbbe42200` | First RED, exit 1 / 316.5 seconds: `cmd/server` EOF HTTP fixture `TestRelayRealHTTPEOFFirstInputAndUnsupportedReceipt`; 19 other Go packages passed, later generator/Vitest stages NOT_RUN. Root returned diagnosis to the same H owner; I's earlier local PASS and any later remote PASS do not overwrite this result |
| Root later complete `pnpm check` | Root reported PASS on exact `4252db7575496b872317b750895ef20f6d414f65` in the same deep environment; I independently fetched and verified every non-Markdown file equals `bdfbbdb`. This separate PASS does not erase Root's first RED or the new Linux evidence-write failure |

The full check uses the existing generator's `--check`; F's checked-in generated types already match the combined contract. No generated types were handwritten or edited by I. Existing unused-contract and bundle-size warnings do not prevent the checks/build from passing.

The default skip count is explicitly 32 pre-existing baseline cases plus six new Relay opt-in cases: four initial/store-reopened HTTP DTO replay cases and two public scan cases across the two viewports. Log parsing confirmed those six as `relay-runtime.spec.ts` (4) and `relay-scan-live.spec.ts` (2), with all other skipped specs totaling 32. Their owners' targeted results and I's actual public scan/registry experiment are separate evidence; the complete default result remains 268/38 and is not full real-user acceptance.

The first exact CI's only Windows failure is the unchanged `web/e2e/s1.spec.ts:123` human-classification/reference test at `chromium-1280`: line 178 `page.reload` reports `net::ERR_NO_BUFFER_SPACE`. This is a browser networking resource error at navigation, not a failed expected-value assertion or timeout. The hosted run used two workers; I's complete local run used the requested one worker and passed. That difference was diagnostic evidence; Root then authorized the narrow execution-resource repair, whose result is recorded independently. No old spec, retry, timeout or skip was changed by I. The exact first RED and full logs were handed to Root and retained. A first log-download request made before the entire run finished was unavailable; completed-run downloads succeeded and were saved separately.

The new `bdfbbdb` Linux first failure is separately retained: `TestRelayRealHTTPEOFFirstInputAndUnsupportedReceipt`, 0.10 seconds, `relay_eof_http_test.go:193`, `EOF public response evidence unavailable`. Earlier actual HTTP/input/receipt assertions completed; the final external public response write failed. I observed that the shorter fixture initialized `Dir(evidence)` while the final write still used `evidence/http-eof-contract-samples.json`; the old recursive layout had implicitly created that verification directory. Root restored H, which explicitly initialized `evidence` in SOURCE `67c7d2a`; the shorter process cwd, all functional assertions and timeout remain unchanged. H's deep target PASS / 3.45 seconds, missing-LOCALAPPDATA fresh fallback target PASS / 3.06 seconds, entire cmd/server 23 PASS / 0 FAIL / 3.47 seconds and vet PASS are separate Windows results. H has no local WSL Go, so native Linux remains for the new exact CI, not inferred from that fallback. Root independently repeated deep/fallback targets with PASS / 7.17 and 3.14 seconds on `449bbce`. The first API log request failed during transport and the whole-run `gh --log` request was not ready; a second direct job-log download succeeded. Those acquisition results and the test RED remain distinct.

## Real registry extension and owned public scan

The external script launched `dist/astrocyte.exe`, the actual server entrypoint built on the integration SOURCE. Child `LOCALAPPDATA` and access state were isolated under this task's external evidence root. `CODEX_HOME`, `CLAUDE_CONFIG_DIR` and `PI_CODING_AGENT_DIR` pointed only to newly created empty public fixtures. Inherited application configuration was removed from this child; no liveness policy was selected. The production `AllowCommittedGit: false` and nil launch policy remain unchanged. Only processes launched by the script were stopped, and their exit events were confirmed.

The corrected second experiment registered an existing installed Node executable with `--version` arguments as public CLI data. Existence was checked, the entry was never dispatched or executed, and runtime capability/readiness remains unknown. Initial JSON held one entry. Adding a second entry while the server was running left the actual `/api/v1/relay/agents` response at one, proving startup-only loading. Stopping the owned server and restarting on the same temporary URL produced two entries in HTTP and in the actual preference dropdown at both 1920 and 1280 pixels. Selecting the new option generated no mutation request or page error. Scheduling code had zero changes between these observations. The isolated service ended with zero tasks and zero action receipts; no model was called.

The F-owned public sample was independently verified through `orca repo show --repo id:8e938f27-488c-4c20-9a15-242364790b60 --json`. A bounded fixture of only that actual public registration was placed in the isolated app's existing registration cache while the owned server was offline. Global registered-root discovery was not invoked, so no other user root or private native history was examined. This cache setup proves a scoped registration input, not complete project discovery. The subsequent actual HTTP `/works/refresh` passed exactly that one sample root to the production scanner: one verified codex work, one folder and two files. Actual UI metrics were `1 / 1 / 1 / 0 / 1`; these sample observations do not establish the user's full work inventory or a real relay event entering cognition.

The first registry trial used nonexistent executable placeholders. Its list/restart mechanics passed, but it does not establish real CLI extension and is retained separately without rewriting its first evidence. Root's correction was applied only to a second external script and separate state/evidence directory. An intervening PowerShell setup invocation failed parsing before any command or write; the corrected setup is recorded separately. These are fixture/setup limitations, not business repairs or a substitute for complete regression checks.

Root noticed that a second-trial screenshot caught the existing dialog's `translateX(100%)` slide-in. A third supplemental stable-frame check waited for animation completion: document horizontal widths and 1920 controls passed, but a full-height bound at 1280 was RED. The fourth independent diagnostic retained that RED and measured the textarea after Playwright nearest scrolling: `y=579.4375`, `height=141.46875`, bottom `720.90625` in a 720-high viewport. Its horizontal extent `x=605`, width `631`, right `1236` stayed within 1280. This was a verification positioning assumption, not demonstrated horizontal product overflow. The fifth independent script used actual instant center scrolling and two animation frames before measuring, without changing the original boundary tolerance or any product/test/animation code. All four key controls were then fully inside the viewport in both before/after frames at 1920/1280; document scrollWidth equaled clientWidth. Stable dialogs were `x=1200`, width `720` at 1920 and `x=560`, width `720` at 1280. Screenshots used `animations: disabled`. The third/fourth REDs remain recorded separately from the fifth PASS; no product layout repair or F acceptance promotion occurred.

## Evidence and remaining gates

I's logs, DBs, configuration, HTTP responses and screenshots remain at `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/integration`. Git contains derived summaries only. Evidence includes initial/current source audits, parsed old-contract and Root-source parity, `install-first.log`, separately named initial/`bdf`/`5dd` complete check logs and result JSON, initial build/E2E logs and external Playwright results/HTML, separately named complete first/resource/fixture/final CI logs and status, and `registry-first`, `registry-second`, `registry-stable-third`, `registry-diagnostic-fourth`, `registry-stable-fifth` with independent summaries. `registry-first-limitations.txt` retains the setup limitation and Root correction. The initial complete check also used the existing Go tests' default public-fixture export location in the adjacent access verification directory; subsequent `bdf`/`5dd` checks used child LOCALAPPDATA under integration to isolate those exports. Existing fixture evidence was not deleted or moved. Root/H-owned diagnosis logs retain their separately stated locations. The external scripts are verification aids, not shipped product infrastructure.

R/H/F reports retain their prior first REDs: native cross-context/check issues, controlled CLI timeout, H Linux fixture and Windows short-root failures, schema drift, and F initial typecheck/layout/locator failures. Later passes do not erase those outcomes. Any newly observed business/test failure must return through Root to the original retained owner; I did not repair any domain or test file.

Root's separate EOF HTTP first failure used a deeper isolated child `LOCALAPPDATA` and empty CLI metadata roots. Root explicitly retained `verification/root-3387019-20261014/check-first.*` and restored H for diagnosis. H subsequently proved that repeated fixture path nesting produced a 281-character cwd and Win32 `ERROR_DIRECTORY 267` before Start; the same public helper at a 126-character cwd completed EOF with exit 0. Root authorized only a shorter internal fixture layout, preserving the same outer isolation, HTTP/service/helper assertions and timeouts. H's exact repair SOURCE `19c83e3` / REPORT `91d27e5` are now ordinary ancestors: same-Root target PASS / 3.50 seconds, entire cmd/server 23 PASS / 0 FAIL / 19.16 seconds, vet PASS, new fixture cwd 233 and complete EOF proof. These are H's separately attributed results; this is not a production launch repair or permission change. The observed Windows long-CWD process restriction remains a limitation. Root's required final checks remain separate from I's engineering evidence.

Unanswered Today append/publication policy, dirty/untracked/non-Git/current-state treatment, visual/liveness policy and CLI workspace tool permission remain undecided. Pending observation admission is not formal collector ingestion, Normalize/engine execution or Today PASS; the formal collector does not consume Relay/application state. Real Codex-to-Claude handover, terminal-only model coding, real three-value CLI UI acceptance and the same-URL three-minute Today loop remain NOT_RUN or BLOCKED. Administrator Procmon formal DB/WAL/SHM zero-write and two real dates at 00:30 remain NOT_RUN. Controlled Go fixtures, default mock E2E, a public scan and the registry experiment do not clear those gates.

No existing service, scheduled task, validation worktree or V4 formal database was changed; port 17321 was not replaced. No deployment or main merge occurred. This engineering integration is complete on the tested final SOURCE; Root retains final target-branch commit/push and product acceptance. Final report commit and remote consistency are supplied in the worker handoff, avoiding a self-referential SHA in this file. The subsequent L1 collaboration is separate work and is not included in this Task or its validation.
