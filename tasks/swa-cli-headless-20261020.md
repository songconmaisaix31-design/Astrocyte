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
| write_paths | 本文件、`docs/standards/swa-cli-headless-20261020.md`、自有 `.verification/swa-cli-headless-20261020/**` |
| 原独立Reviewer | `task_1cd8014750f9 / ctx_980a3282c339`，本次不另起Agent |
| 实施状态 | `DESIGN_REVISION / TEST_PLAN_PENDING_REVIEW / IMPLEMENTATION_NOT_STARTED` |

接管已读取main的HANDOFF/STATUS及本树AGENTS/CODING/RESOURCES、最新首版任务和QUESTIONS相关决定，完整读取主控brief。已核准确上游 `stablyai/orca v1.4.230@b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`，来源工作树干净；只读目录为 `C:/Users/DW/orca/Astrocyte/.verification/orca-source-v1.4.230`。准确30b的两份来源报告保留其原调查与旧方案适用范围，本任务按最新首版范围提出方案。

## 问题、建议与主控决定

完整方案及测试命令见[报告](../docs/standards/swa-cli-headless-20261020.md)。抽取原生Node runtime、SQLite、Run/Task/Dispatch、真实会话与进程控制，使用独立 `swa` 入口、安装根和私有数据根；产品composition不组装全局凭据迁移、全局hooks更新、账户迁移、自动化、云推送及桌面入口。必须保留真实控制端与Worker的身份、持久性、输入观察、输出和退出语义。

主控通过本次Orca `ask` 明确批准：**产品私有Orca原生Run/Task/Dispatch为本首版唯一执行账本**，SWA Workspace复用适用项目、授权和固定上下文；旧Mission/Claim/Relay保持独立记录，同一首版任务不双写。主控在准确方案独立审核后更新SPEC，本Worker不修改公共事实源、既有接口或旧门禁。

主控明确禁止使用当前Claude的Aliyun配置或修改本机Agent配置，允许静态核查现有非AliyunCodex认证和Pi的兼容性。Python3.13的 `tomllib` 与标准JSON解析确认当前Codex两处配置均为默认OpenAI、`gpt-6.1-sol`、`auth_mode=chatgpt`，存在OAuth tokens、无API key；检查只输出白名单非秘密字段。已安装Pi1.0.1的源码包含相同原生provider/模型及CredentialStore接口。建议采用真实Pi SDK和其原生RPC，凭据只读适配在刷新回调之前拒绝修改，全部工具/扩展/上下文发现关闭。主控已接受此静态兼容调查方向；认证成功及受限正向均未验证。

原Reviewer准确`142869e5ddb7b06a1a1f45be978dbd1c595a98d1`对`5523564f2d76ee992b8ffe00d590dbc3b24a4f85`退回。本次按F01—F03修订真实SDK factory及定向provider刷新、controller公开请求到原生receipt/PTY/Run的映射，并采用当前权限下的写入消费者、真实路径/端点/进程和本机Orca定向保全组合。计划不要求WPR、提权、性能组或驱动；准确返修仍须原Reviewer接受，程序写域和动态窗口仍未授予。

报告§4.3明确Pi过滤前的package-manager home消费者及仅在SDK子进程设置的私有home/agent/session/package路径；§5.3列出逐文件的程序写域申请，新增模块限定必要目录。该清单尚未授权；已只读核main后继`b26a2fab9b5cad853c777793dd17593fb4a509a2`的HANDOFF/STATUS及SPEC §21.1，本树尚未纳入该后继。

## 验收与执行顺序

| 顺序 | 结果及停止条件 |
| --- | --- |
| 本次静态返修 | 两份文档完整描述准确来源、依赖、隔离消费者、身份、CLI和T00—T15；来源/链接检查后普通commit/push给主控与原Reviewer，不据此结束整体产品任务 |
| 独立审核 | 原Reviewer自行推导条件并审核准确Commit；未接受不写程序。新增边界、命令和Pi只读认证均属于待审核方案 |
| 实施准入 | 主控更新相应SPEC并授予程序/依赖/测试写域。动态窗口单独授予；本次没有窗口 |
| 实现与自测 | 同一Worker实施真实native后台、控制端、Pi SDK适配及CLI，运行已审核全部检查；失败时返修 |
| 安装验收 | 在同一准确候选完成安装和一次含至少两个真实Worker的任务：实际模型执行重叠、分别观察、指定停止A、B继续、获取成果、本机Orca保全 |
| 收尾 | 提交实际命令、结果、原件位置和限制，主控及原Reviewer核验；独立接受不自动授予合并/发布 |

首版最终结果不能用本次开发会话、启动回执、两个空闲进程、串行模型输入或不同候选的局部结果代替。公开固定任务由Agent创建的公开Git项目提供；私有用户作品和自由工具保持原范围。Codex0.162.0受限原生工具要求保持拒绝，Pi结果不改写Codex或旧S3的验收。

## 资源与本次检查

当前原S服务PID52040占用唯一窗口，本次未操作该服务、原库或本机Orca进程。本次未运行install/build/model/service/实际产品CLI任务，未调用视觉功能。纯文档和配置结构读取使用BelowNormal；Node文档检查heap上限512MiB。后续构建、安装及真实验证按RESOURCES串行，由主控先安排原服务保全与窗口；同一任务双Worker并发在该次唯一验证窗口内执行。

已执行的检查与实际结果在报告§9登记。程序构建、功能测试、模型调用、安装、双Worker并发/停止/成果和本机Orca安装前后保全均为 `NOT_RUN`。当前限制为方案/测试计划待独立审核、程序写域与动态窗口未授予、Pi受限认证和实际依赖闭包尚未动态验证；不将这些状态描述为功能完成。
