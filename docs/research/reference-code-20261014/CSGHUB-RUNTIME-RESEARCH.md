# CSGHub Agent 状态、恢复、研究步骤与经验存储

来源和许可证见 [SOURCE.md](SOURCE.md)。本报告的 Go、Python、Shell 结论均来自精确 SHA `a67b3024a9553aecf3379c9f6bccfcdada2034f5` 的常规源码阅读。REA 证据仅覆盖 SOURCE 中指定的前端 JavaScript。没有执行服务端、数据库迁移、容器、评估脚本或模型调用。

## 公开实现的运行边界

公开仓库含运行配置和完整的数据层函数，但默认 CE 的 [component/sandbox_v2_ce.go:1](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/sandbox_v2_ce.go#L1)将 `SandboxV2Component` 定义为空接口，并注明 adapter 依赖 EE/saas；[component/callback/agent_deploy_watcher_ce.go:11](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/callback/agent_deploy_watcher_ce.go#L11)的 `WatchAgentChange` 直接返回 nil。公开代码可以支持下述存储行为分析；从这些函数无法确认 CE 完整部署/执行/恢复链路。EE/saas 相应控制器在本次公开代码中未核实，其实现和许可证为未知。

[component/agent_runtime_profile.go:23](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/agent_runtime_profile.go#L23)定义 agent type、version、image、port、command、readiness probe、环境变量和 UI 配置；`:53` 从配置目录同步，`:100` 用标准 JSON 解析并校验。`content_sha` 用于其配置更新判断。该配置存在不能证明本机安装、认证、启动或 resume 已通过。Astrocyte 已有各项能力独立观察的设计，继续使用原生控制器实际结果；无需引入新的 Manifest、Hash 或调度系统。

## Agent 状态与恢复的数据层

| 实现 | 固定文件和行号 | 可确认的行为与限制 |
| --- | --- | --- |
| 身份与运行状态分离 | [builder/store/database/agent_orchestrator.go:39](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L39) | `AgentAllocation` 分别保存 actor、runtime principal、desired state、observed state、runtime generation、accepted generation、last error 与结束时间。请求身份、目标状态和观测结果是不同字段。 |
| 执行准入 | [agent_orchestrator.go:888](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L888)、[agent_orchestrator.go:1078](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L1078) | 事务锁定 allocation；actor、runtime principal、仍有效的 binding、desired running、observed ready、accepted generation、请求 pinned generation 必须满足要求。检查同 session 的 lease 和容量后创建 reserved admission。 |
| 超时后的未知状态 | [agent_orchestrator.go:981](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L981)、[agent_orchestrator.go:1035](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L1035) | stale reserved 取消并结束 lease；stale running 改成 unknown，保留 lease，设置容量占用截止时间。同一 session 仍返回 `ErrAgentExecutionRecoveryRequired`。容量占用到期不等于旧执行已经停止。 |
| 完成和迟到结果 | [agent_orchestrator.go:1124](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L1124) | 完成更新检查实际 affected rows。unknown 的迟到 completed/failed 仅在 lease 尚未结束且 allocation 仍接受 admission 的 generation 时可以更新；恢复后的旧 generation 不能沿该条件更新。 |
| 显式恢复 | [agent_orchestrator.go:1369](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L1369) | 校验 actor 和当前 lifecycle；复用已有活动 restart operation。新恢复在同事务递增 generation/operation sequence，将旧 reserved/running/unknown 标为 interrupted、结束 lease，创建 pending restart。该事务表示恢复记录建立，尚未证明外部 runtime 已重启。 |
| 恢复运行与完成 | [agent_orchestrator.go:1434](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L1434)、[agent_orchestrator.go:1458](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_orchestrator.go#L1458) | 启动 restart 需要 dispatched；完成需要 running 和有效 deploy ID；两者检查 operation generation 是否仍为当前 generation。函数返回成功只能按各自具体状态解释。 |

值得复用的行为是：请求已发送、运行已观察、停止已确认分别记录；回复超时保留 unknown；恢复前核实当前运行身份和状态；旧操作不能覆盖新状态。上游 generation/lease/allocation 表及控制器不直接引入 Astrocyte。

## 会话与反馈记录

[builder/store/database/agent_instance_session.go:16](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_instance_session.go#L16)保存 session UUID、instance、user、type 和 last turn；`:28` 的 history 保存 request/response、turn、content、feedback、feedback reason 和 rewritten 标志。

[History Create:299](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/agent_instance_session.go#L299)在事务内锁定 session 行：请求增加 turn；响应要求当前 turn 已有请求，且没有已有响应；history 插入与 last turn 更新在事务中完成。可参考其显式关联和并发检查，不能把数据库中存在 history 解释为原生会话可恢复，也不能将存储原始会话内容引入 Astrocyte 的公开派生事件。

反馈是 `none/like/dislike` 和 reason；源码未在这里实现研究结论审核、经验批准或跨项目经验分享。反馈字段不能充当 Astrocyte 的人工批准状态。

## 研究步骤与评估结果的公开实现

CSGHub 的已核实执行流程是模型/Agent 评估。没有核实到一个完整的自动提出假设、科研计划、审稿、改进并形成已批准经验的公开实现。下表列出实际可读部分，不从产品名称推断额外步骤。

| 位置 | 已核实步骤 | 供本项目参考的关系 |
| --- | --- | --- |
| [component/evaluation.go:108](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/evaluation.go#L108) | 当前用户和目标 namespace write 检查；启用的 framework；解析模型、数据集；规范化 framework config；选择资源和实际镜像；提交任务。 | 一次操作的权限、实际输入版本、实际配置和结果应能关联。执行配置和 UI 默认值不同，应记录执行配置。 |
| [builder/deploy/deployer.go:937](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/deploy/deployer.go#L937) | 建立 evaluation template 和真实 workflow 请求，传递 env、模型版本、数据集 revision、hardware；保存模型 commit、数据集 commit、配置和资源快照。 | 模型按 commit 执行；数据集执行端按 revision 下载，快照另记 commit。后者不能作为完整可重复执行的证据。 |
| [docker/evaluation/evalscope/start.sh:101](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/docker/evaluation/evalscope/start.sh#L101)、[start.sh:183](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/docker/evaluation/evalscope/start.sh#L183)、[start.sh:209](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/docker/evaluation/evalscope/start.sh#L209) | 下载数据集；按模型列表循环下载指定版本，将各次结果放进独立目录；调用 evalscope；收集报告并汇总上传。 | 保存输入与具体结果的关系能防止两个同名仓库或同一模型多个版本混淆。脚本存在备用来源、默认值与继续执行行为，需按本项目立即失败要求审阅，不能原样复制。 |
| [docker/evaluation/evalscope/upload_files.py:148](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/docker/evaluation/evalscope/upload_files.py#L148)、[upload_files.py:242](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/docker/evaluation/evalscope/upload_files.py#L242) | Python 标准 JSON 读取真实评估报告；按运行目录关联 repo ID/revision，将 metric/score 形成 JSON 和 pandas Excel；`:85` 可经 bearer token storage gateway 上传。 | 结果记录关联具体输入，使用成熟格式库输出。该脚本的部分降级与异常处理不符合本项目要求。结果地址的下载权限仍需独立核实。 |
| [builder/store/database/argo_workflow.go:50](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/argo_workflow.go#L50)、[component/evaluation.go:464](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/evaluation.go#L464) | 保存 task ID、phase、reason、时间、镜像、版本/配置快照和结果 URL；读取结果检查任务类型与 namespace read。 | 评分、运行成功、用户可读和人工认可分别处理。实际报告对象下载鉴权不能从结果查询鉴权推断。 |

[api/workflow/deploy_reconcile.go:10](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/workflow/deploy_reconcile.go#L10)使用 Temporal activity 的 10 分钟超时和最多三次尝试。公开 [ReconcileWorkflowStatus，deploy_reconcile_activity.go:357](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/workflow/activity/deploy_reconcile_activity.go#L357)会查询 runner phase；`:442` 的分支在 runner 查询错误且超过 hard timeout 时将 workflow 标成 failed。因此它不能支持“所有无法观察状态都保留 unknown”的结论。Astrocyte 仍须按原生控制器实际信息保留未知状态；不新增 Temporal 或自研调度层。

## 经验/记忆的存储边界

[api/router/api.go:1351](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/router/api.go#L1351)的 `/api/v1/memory` 接口要求登录，提供项目创建/查询/列表/删除和 memories 写入/search/list/delete、health。

[common/types/memory.go:39](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/common/types/memory.go#L39)定义 episodic/semantic 类型及 agent/session/org/project scope。`user_id`、content、metadata 与 scope 是传输字段。[api/handler/memory.go:84](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/handler/memory.go#L84)检查非空消息、非空 content、合法类型，并拒绝单条消息覆写请求级 scope。[SearchMemories:158](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/handler/memory.go#L158)处理分页后转交 component。所读 handler 没有展示 caller 与请求 project/org scope 的 ReBAC 检查；仅登录、scope 字段存在不足以证明跨项目隔离。

[component/memory.go:33](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/memory.go#L33)根据配置创建 backend adapter；[builder/memory/memmachine/adapter.go:21](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/memory/memmachine/adapter.go#L21)注册 MemMachine，[client.go:112](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/memory/memmachine/client.go#L112)把 memories 写入外部 HTTP 服务，[client.go:152](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/memory/memmachine/client.go#L152)调用外部 search。实际记忆持久化、检索、模型生成和经验批准不在这些 adapter 中实现；外部服务内部实现、许可证及运行结果本轮未知。

Astrocyte 已有 `ExperienceService.CaptureFailure` 从真实 Event 捕获 quarantine、人工 approve/demote、明确分享记录和 evidence 解析。沿现有 `Experience`/`Knowledge` 类型核实来源、版本和批准权限即可。用户自由工具、真实用户作品、V4 自动科研范围仍未授权；没有把 CSGHub conversation 或 semantic memory 自动转换成已批准经验。

## 本轮覆盖与剩余未知

已覆盖资产导入、权限、版本，Agent 数据层状态/恢复，会话关联/反馈，评估步骤/结果归属和 memory adapter。可执行检查交给 [HANDOFF.md](HANDOFF.md)。公开 CE 完整 Agent runtime、EE/saas 控制器、外部记忆持久化、自治科研/审稿/经验批准实现、autoscientist 准确仓库保持未知。未进行上游构建或部署测试，因此报告不能用于声称其产品功能已在本机通过。
