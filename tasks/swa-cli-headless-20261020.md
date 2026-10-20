# 首版 Swarm CLI：隔离后台、身份与真实测试计划

## 当前任务

| 项目 | 准确事实 |
| --- | --- |
| 交付目标 | 完成方案返修和独立审核后，由同一Worker继续产品实现、全部真实测试及安装验收 |
| 最新范围 | [首版任务](swa-cli-v1-20261020.md)、[用户决定](../docs/QUESTIONS.md)首版章节；产品自带隔离后台，不妨碍本机Orca |
| 基线 | 从 `81812c32a6d345362655c6e569f680aab8fca877` 普通纳入e149及main `4e6e8ed4fa52521c2232dd770eb98297d8713785` |
| 工作树 / 分支 | `C:/Users/DW/orca/workspaces/Astrocyte/swa-cli-headless-20261020` / `swa-cli-headless-20261020` |
| 当前Run / Task / Dispatch | `run_62ada397a104` / `task_9f062888e786` / `ctx_0d069056baff`；原静态交付为`task_4501441a2f59 / ctx_dc2fe5ba5526` |
| Worker / 主控 | `term_f5235f4d-4de8-4f6e-8b12-a04e22c55500` / `term_3209681a-faa5-4161-a2f3-9563e1451a33` |
| 实际客户端 / 模型 | Orca中Codex；本次 `terminal show` 的实际预览为 `GPT-6.1-Sol xhigh fast` |
| write_paths | main f7c97d36的[首版任务](swa-cli-v1-20261020.md)“已授权的实施写域”，对应报告§5.3准确清单；公共事实源由主控维护 |
| 原独立Reviewer | `task_1cd8014750f9 / ctx_980a3282c339`，本次不另起Agent |
| 实施状态 | `TEST_PLAN_ACCEPTED / IMPLEMENTING / FUNCTIONAL_VALIDATION_NOT_RUN` |

接管已读取main的HANDOFF/STATUS及本树AGENTS/CODING/RESOURCES、最新首版任务和QUESTIONS相关决定，完整读取主控brief。已核准确上游 `stablyai/orca v1.4.230@b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`，来源工作树干净；只读目录为 `C:/Users/DW/orca/Astrocyte/.verification/orca-source-v1.4.230`。准确30b的两份来源报告保留其原调查与旧方案适用范围，本任务按最新首版范围提出方案。

## 问题、建议与主控决定

完整方案及测试命令见[报告](../docs/standards/swa-cli-headless-20261020.md)。抽取原生Node runtime、SQLite、Run/Task/Dispatch、真实会话与进程控制，使用独立 `swa` 入口、安装根和私有数据根；产品composition不组装全局凭据迁移、全局hooks更新、账户迁移、自动化、云推送及桌面入口。必须保留真实控制端与Worker的身份、持久性、输入观察、输出和退出语义。

主控通过本次Orca `ask` 明确批准：**产品私有Orca原生Run/Task/Dispatch为本首版唯一执行账本**，SWA Workspace复用适用项目、授权和固定上下文；旧Mission/Claim/Relay保持独立记录，同一首版任务不双写。主控在准确方案独立审核后更新SPEC，本Worker不修改公共事实源、既有接口或旧门禁。

主控明确禁止使用当前Claude的Aliyun配置或修改本机Agent配置，允许静态核查现有非AliyunCodex认证和Pi的兼容性。Python3.13的 `tomllib` 与标准JSON解析确认当前Codex两处配置均为默认OpenAI、`gpt-6.1-sol`、`auth_mode=chatgpt`，存在OAuth tokens、无API key；检查只输出白名单非秘密字段。已安装Pi1.0.1的源码包含相同原生provider/模型及CredentialStore接口。建议采用真实Pi SDK和其原生RPC，凭据只读适配在刷新回调之前拒绝修改，全部工具/扩展/上下文发现关闭。主控已接受此静态兼容调查方向；认证成功及受限正向均未验证。

原Reviewer对5523退回的F01—F03已经按准确396e修订并独立接受：真实SDK factory及定向provider刷新、controller公开请求到原生receipt/PTY/Run的映射，以及当前权限下的写入消费者、真实路径/端点/进程和本机Orca定向保全组合。计划不要求WPR、提权、性能组或驱动；独立接受与实施授权见后文。

报告§4.3明确Pi过滤前的package-manager home消费者及仅在SDK子进程设置的私有home/agent/session/package路径；§5.3列出逐文件程序写域，新增模块限定必要目录。原Reviewer准确`01fd4bc1ef40a0d9494ed99a649c00ec5487ca5f`独立接受396e方案和T00—T15，主控在`msg_4359842084e5`授予该清单及028迁移，并要求普通纳入main。本树已从干净396e普通快进至`f7c97d36af03309868269a7ee3d8fa791a570974`，重新读取SPEC §21.1和准确授权，开始同一Worker静态实现。

## 验收与执行顺序

| 顺序 | 结果及停止条件 |
| --- | --- |
| 本次静态返修 | 两份文档完整描述准确来源、依赖、隔离消费者、身份、CLI和T00—T15；来源/链接检查后普通commit/push给主控与原Reviewer，不据此结束整体产品任务 |
| 独立审核 | 原Reviewer01fd4bc1已接受准确396e的方案与T00—T15；实现和真实结果继续独立CR |
| 实施准入 | 主控f7已更新SPEC并授予程序/依赖/测试准确写域；依赖窗口已执行并交回，构建/测试/服务/模型窗口待交付 |
| 实现与自测 | 同一Worker实施真实native后台、控制端、Pi SDK适配及CLI，运行已审核全部检查；失败时返修 |
| 安装验收 | 在同一准确候选完成安装和一次含至少两个真实Worker的任务：实际模型执行重叠、分别观察、指定停止A、B继续、获取成果、本机Orca保全 |
| 收尾 | 提交实际命令、结果、原件位置和限制，主控及原Reviewer核验；独立接受不自动授予合并/发布 |

首版最终结果不能用本次开发会话、启动回执、两个空闲进程、串行模型输入或不同候选的局部结果代替。公开固定任务由Agent创建的公开Git项目提供；私有用户作品和自由工具保持原范围。Codex0.162.0受限原生工具要求保持拒绝，Pi结果不改写Codex或旧S3的验收。

## 资源与本次检查

主控已保全并停止原S52040，仅为本Worker交付两条依赖命令窗口；本Worker没有操作原服务、原库或本机Orca进程。首次PowerShell未签名脚本在正文前被拒绝；主控随后明确允许本次进程的ExecutionPolicy参数，原失败保留且用户/系统策略未改。两次新资源双采样通过，已串行完成`pnpm install --lockfile-only --ignore-scripts`及`pnpm install --frozen-lockfile --ignore-scripts`，均EXIT0、BelowNormal、Node heap1024MiB和私有temp/cache/store；完整原件位于自有`dependency-01`。安装日志含一次镜像下载超时及pnpm内建重试后成功，没有另发安装命令。窗口已通过`msg_9dca8338f570`交回主控。

主控在`msg_655f03dd1e98`确认原S已同源恢复为PID12800，原库/环境/Q/媒体/CodeGraph保全，原Orca同runtime仍ready；本Worker未接管原服务。入口exec会话37535实际EXIT0，两次PNPM同步返回均为0；定向CIM未发现本项安装脚本或PNPM store参数进程残留。本次没有逐时记录全部短命子PID，对该观察范围保持限制。

已执行的检查与实际结果在报告§9登记。准确b287源码及MIT许可已通过Git archive导入获准来源副本；App/关联存储、HTTP及服务组装、原生调用适配、Pi只读SDK入口和native controller composition已形成阶段源码。主控msg_eaaf01e60d77允许阶段Commit供原Reviewer静态CR；本阶段仍有CLI、打包安装、保护检查及真实测试实施待办。产品构建、功能测试、模型调用、产品安装、双Worker并发/停止/成果和本机Orca安装前后保全均为`NOT_RUN`；当前窗口归原S12800，Pi受限认证和可安装闭包尚未验证，阶段提交不表示功能完成。

阶段`847ec2687857b06430284aecdf81a6887f2f996f`已普通push、EXIT0；原Reviewer`80ed37f6af51f6c38df4f811d8e6c8f2376c089e`退回F04—F07，当前同一Worker返修原生单任务/两席原子准入、真实资源字段和human安全停止。报告§5.4给出父runtime分别持有A/B专有Win32 Job、模型前原生IPC准入、禁止breakaway、句柄所属/最后关闭、全树查询及独立进程退出观察；T03/T06/T08/T09/T12/T13补相应真实条件。此项必要差异先交原Reviewer，相关依赖/native构建仍待接受与明确窗口，其他CLI和打包静态工作继续。

来源子树是准确b287的Git archive导出视图与产品适配，包含导出省略和Windows执行位变化；实际边界见报告，不声明为整个原树逐项相同。阶段源码及剩余实现没有构建、模型、安装或双Worker验收通过结论。

原Reviewer msg_08a9c85e2399确认准确6168的F04—F07静态修复成立，F08要求将关闭和未物化的创建放入同一原生事务屏障。当前返修关闭准入、保留pending/unknown及已完成关闭屏障，并补T06/T11真实竞争；准确后继仍待独立核查。CLI、原生成果分页、项目允许目录导出、私有HTTP生命周期和构建脚本继续实施；node-gyp12.3.0沿准确上游锁声明为必要构建依赖，尚未更新锁或安装，不把这项静态声明写成已可构建。

原Reviewer准确6874db16正式接受6168四项静态修复及父持Job方案/测试补充，主控msg_5ca526aeec7c授权继续必要native/runtime实现；已读取main13974的SPEC21.1更新，原S12800和无动态窗口保持。旧self Job工作原型已手动移除，当前按父持方案实现独立native句柄及完整查询、真实Node IPC bootstrap和全树退出后的生命周期发布。CLI、原生成果、独立服务、构建/安装入口已形成静态源码，后继准确CR与全部真实验收仍待执行；锁及OpenAPI生成类型尚未同步。

准确f0ba550352ea28b63b008c12542e83dc47adbdf3已普通push，包含阶段源码10deb及获准main13974；后继继续编写真实普通后代fixture和CLI字段拒绝测试，均NOT_RUN。私有next-window.ps1已列dependencies、generate和build三个独立Stage及资源参数；窗口申请通过Orca ask等待主控明确交付，本Worker没有执行这些Stage，没有操作原S或启动模型/服务。主控msg_aeb9a96e921e明确可以先构建再继续真实测试，源码CR与窗口授权分别处理。

后继53ad6e1479ef5cf3099372d2855087cf002513bb已普通push；原Reviewer对f0接受F08静态修复并退回F09未分配空Job结算路径。当前返修实际root成功分配/成员见证、真实processless无树证明及相应普通Node负向测试，同步报告§3.1为程序实际平铺安装布局，待准确后继复审。主控ask已返回prepare-02资源拒绝，可用3.529GiB不足4GiB，S12800仍运行；新依赖/生成/构建、SDK、服务及安装均未执行，继续静态源码和真实测试准备，不退出整体产品任务。

准确f8c69161a6a3bb030c730cdc0c9ff6229b8c95e6已普通push，F09返修等待原Reviewer准确审核。报告§5.5补充完整成果读取与T07/T13真实用例，解决原生workerRead单block1200字符的预览边界；在独立接受前保持新增读取路径未实施。安装候选继续补充实际安装脚本、准确使用说明和T00/T02/T13/T15检查，不将缺少真实项目Grant或token自动转为human权限。

原Reviewer cc4d1e6接受准确f8的F09静态修复；99d084da接受准确d7完整成果计划和76e9dafd安装入口/说明、真实注册target及SDK空闲abort测试源码。主控msg_db8583a97ed2已实际读取回复并授予实施§5.5，新增严格成熟JSONL和native journal只读成果、真实Worker source观察、每次Worker创建前实际资源双采样及私有SDK事件日志正在同一写域实施。T00公开来源、T02实际安装、T13服务/授权拒绝入口和T15定向Orca保全脚本已形成源码，双Worker完整驱动和后继真实用例继续编写；缺少脚本/授权/动态窗口均明确失败，不能登记通过。全部JS/Go功能、SDK、构建、安装和模型检查仍未运行。

准确0c6及普通纳入main4b后的f2已push，主控已实际读取；原Reviewer提出F10—F12后，本Worker修正五项真实事件订阅的取得/存活检查/消费/清理、实际ownerProcess命名及按Pi1.0.1角色/内容块/entry的严格Zod校验。T13真实双Worker驱动、T07实际SDK原件的自有故障副本检查、本机Orca并发期间定向观察及服务所有后代退出检查已编写，后继仍待独立CR和真实运行。候选补齐原node-pty实际ConPTY文件及SDK实际xhigh核验；所有修改只完成git diff --check与PowerShell语法读取，JS/Go程序、SDK、安装、模型和资源窗口均未运行或授予。

准确abdb07d31954b9e72f12bdaafdfe56406deb9004已普通push并经ls-remote核实；初两次TLS失败保留。原Reviewer a211b66接受F10—F12静态修复，F13/F14后继已按真实原生Stop目标turn/fold及成熟stringer正确边界和目标错误断言返修，继续同一模块复核。四包Go真实检查、无服务编译及实际服务执行分开的入口、T02真实拒绝用例和已跟踪资源采样器已编写；所有正文NOT_RUN。最近主控只读可用0.376GiB，未交新窗口，原S12800与本机Orca保持，整体首版没有通过动态验收。

准确39cc967df6a71b2c0c22e3daa520ca08331db17e已普通push并经主控核实，F13/F14继续独立CR。T06发现同R任务创建会重新采样context ID/时间及解析创建后的scope；主控msg_5ff5b86bf277确认按既有语义修复。现有原生指纹规范化保留完整正文、实际context版本/内容/路径/权限、caller/project及scope的project/controller/authority，handler和首次固定原件保持；具体剔除字段及等价绑定见报告§3.3。真实非模型幂等脚本及T13写动作重放/两席拒绝已形成源码，断开/重启及不同R竞争保持待验证；最新资源观察可用0.247GiB，无新动态窗口，全部正文仍NOT_RUN。

准确9741fccb8f705b057b913e617d9ae522536c1e88普通push及Worker ls-remote EXIT0，主控独立远端查询连接失败如实保留，已交原Reviewer静态CR。后继仅在swa方法边界保留实际OrchestrationError原code/data，修正既有CLI身份/资源及pending退出语义，并补T08真实SQLite写锁拒绝源码；输入后写失败及自然低资源仍待验证。gofmt -d、PowerShell ParseFile和git diff --check通过，全部程序正文/SDK/编译/安装/模型仍未执行；原S和Orca保持。

准确2c007de5ec54ce1cdb11a2f6052a116d1dc958c4普通push及Worker远端核对EXIT0。T06后继源码从真实完成receipt及实际PTY进程确认副作用，在未收到stdout且仍为精确创建身份时只终止本项CLI子进程，再查询原R和同一controller；未命中仅记INCONCLUSIVE且不重试，其他输入/重启边界保留未验。PowerShell语法及差异检查通过，所有正文NOT_RUN，无动态窗口，继续同一模块直至真实验收或实际阻塞。

准确72c7b8961a5677c7291967448507d3fc64340a4b普通push及Worker ls-remote EXIT0。后继让request.show在当前项目授权后直接按真实receipt核project/controller，避免pending固定任务提前阻断查询；所属错配保留scope_denied/EXIT3。真实脚本补两种selector等价及另一实际controller拒绝，gofmt -d和差异检查通过，所有正文仍NOT_RUN；主控最新可用0.396GiB，无动态窗口。

准确8220d851dd998639a0ae3fbf4eff0212dd20ed7c普通push及Worker ls-remote EXIT0。主控msg_34756d70dc45、msg_e18403c22437确认CLI本地错误/原R输出及serve未消费自身stdin的缺口，要求先补计划独立审核；报告§7.5列C01—C09、真实ErrorV1来源、网络断开、隐藏console隔离和实际130、启动前中后EOF/已观察后代与unknown条件。当前只修改两份文档，程序保持8220；未发送console信号，所有新增正文NOT_RUN，计划未接受前不实施该补充。

原Reviewer准确2bd10d21接受39cc静态返修及四包Go补充，d7fbcfb0接受准确b3f0e229的§7.5计划；主控msg_d7ccfb662bde已授予同一CLI写域静态实施，无运行窗口。F15按实际mtime/size版本修正真实用例：未变化时成功重放，context变化及恢复字节后的真实新版本都拒绝，原生首次固定context保持；没有伪造旧mtime或丢弃version。程序/SDK/编译/信号/安装/模型正文继续NOT_RUN，后继准确CR及§7.5实施继续由本Worker负责。
