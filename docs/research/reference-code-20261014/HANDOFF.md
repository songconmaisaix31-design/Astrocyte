# 固定逆向 Agent 首轮开发参考 Handoff

接收方：M-product `run_dbf1a68bc692`；抄送 ROOT `run_973808dd6a2f`。发送方保持 Codex 会话 `01a13ae1-1399-78d2-b329-0ea4fc5eb255` / Orca `term_79eeba93-c810-4a04-a1a0-e27ce0767ab5`。本阶段沿单次用户任务使用普通 Handoff，没有正式 Task/Dispatch 生命周期。

工作树：`C:/Users/DW/orca/workspaces/Astrocyte/reverse-reference-20261014`。分支：`songconmaisaix31-design/reverse-reference-20261014`。精确基线：`1c026c3e7f9e91c53d1dced6479099668c9162cb`。正式写域仅 `docs/research/reference-code-20261014/**`；下载和 REA 输出仅本工作树 `.verification/reference-code-20261014/**`，已忽略。

研究依据：[SOURCE](SOURCE.md)、[资产代码索引](CSGHUB-ASSETS.md)、[运行与研究流程分析](CSGHUB-RUNTIME-RESEARCH.md)。现有负责人先审阅，再按当前任务文件决定是否采用。此 Handoff 不创建新的实现任务、不解除冻结、不替换业务负责人。

## 按现有负责人分配的可执行检查

| 接收负责人 | 本地入口（精确基线中已有） | 检查内容与建议顺序 | 真实验收条件 |
| --- | --- | --- | --- |
| M-product → B | `internal/workspace/domain/repositories.go`；`internal/adapters/sqlite/local_agents_repositories.go:53`；`internal/attention/app/materials.go:49`、`:86` | 优先核查材料来源版本、附件 revision、Git metadata revision、placement revision、实际 clone Head。参考上游镜像数据库事务边界及模型执行 commit。针对数据集快照/执行 revision 不同的风险，要求输入记录能关联实际读取版本。 | 使用公开真实来源修改后，旧材料内容仍能按原 revision 读取；无权修改/读取拒绝；实际下载错误保留；版本未知不伪造；数据库记录和完成状态与实际文件一致。已有流程若满足要求，记录检查结论，不添加重复系统。 |
| M-product → F | 现有 `web/` 资料、项目和 Agent 状态页面；后端 `internal/adapters/httpapi/session.go` | 检查 UI 是否把请求提交、已启动、已完成、停止确认、未知状态分别显示；参考 RepoDetailStore 的权限/镜像状态呈现和 REA 已提取 API 调用。所有可操作权限由已有后端结果决定。 | 在实际 HTTP 与浏览器中呈现真实状态和错误；后端拒绝时 UI 不显示成功；CSRF/origin 与 session 要求继续通过。由现有 F 按任务范围选择页面。 |
| M-product 转 M-relay → R | `internal/adapters/agents/native.go`；`internal/adapters/agents/native_process.go:179`、`:452`、`:464`；`internal/workspace/domain/projects.go:109` | 核查 reply timeout/退出后 DeliveryUnknown、resume/stop/observe/reconcile 的原生结果，及旧操作迟到结果的归属。参考上游 unknown 继续阻止同会话重复执行、恢复先核实身份的行为。 | 真实控制器 reply 未确认时不得自动重发；停止未确认不得显示 stopped；reconcile 必须读取当前真实控制器/会话；原生能力不能由配置存在推断。保留原控制器与已有字段，不复制 allocation/lease 调度体系。 |
| M-product 转 M-relay → H | `internal/adapters/events/native_reader.go`；`internal/adapters/events/native_record.go`；`internal/attention/app/event_service.go` | 核查实际 session、operation、event 的关联与重复/乱序处理。参考上游 turn 明确关联的目的；Astrocyte 继续读取控制器自产记录并生成已允许派生字段。 | 每个 event_id 在真实 SQLite 中存在；保持真实输入/操作/结果关联；未知 outcome 不转换为失败或成功；原始会话和私有内容不进入交付。 |
| M-product 转 M-release → Q/S，涉及知识域先查冻结 | `internal/attention/app/experience_ingest.go`；`internal/attention/app/experience_service.go:71`、`:198`、`:229`；`internal/attention/domain/experience.go` | 阅读上游 feedback 与 memory adapter 边界，审阅当前 Event → quarantine → 人工批准 → 明确分享的证据关系。研究评分、用户 like 和批准权限保持各自语义。 | 只从真实已入库 Event 捕获；unknown outcome 不生成失败候选；人工批准校验 caller/project/expected version；跨项目仅明确分享可见；经验 evidence 可逐条查到。不打开 V4 自动任务，不引入外部 MemMachine 服务。 |

## 实现前需要保留的决定边界

- CSGHub 两仓库分别固定 SHA；它们是否构成官方兼容发布组合未知。第三方代码移植、依赖安装和服务部署尚未执行。
- 公开 CE 中 sandbox-v2 接口与 deploy watcher 不提供完整运行控制器；不得按接口名称制作假的原生能力。
- 上游归档 commit 解析失败后的继续下载、数据集执行 revision 与记录 commit 不同、runner 错误加 hard timeout 推断失败，都必须按本项目实际验收要求处理。
- autoscientist 准确仓库由用户决定；SOURCE 只列官方候选。核心实现未公开或许可证未核实的部分继续标为未知。
- 运行验证继续限定已授权的 Agent 创建公开项目、真实控制器/SQLite/HTTP/自产会话、原生文件与命令工具关闭；实际用户作品和自由工具保持待定。本研究没有运行该链路，也没有扩大授权。

## 验证与移交

本轮执行 `rea analyze-javascript-application` 与 `rea trace-application-feature`，均退出码 0；coverage partial，七个 JS 解析失败 0。两仓库 SHA、根许可证和 clean 状态已核实。执行 `node .verification/reference-code-20261014/verify-references.cjs`，退出码 0：55 个不同的固定源码链接的仓库、SHA、文件与非空行号均匹配本地公开副本；REA 追踪引用原始证据，真实 seed 匹配数为 1。结果保存在忽略目录 `reference-check.json`。此检查验证引用位置和证据关系，功能结论仍按上述源码阅读范围解释。正式修改执行 `git diff --check`。本轮没有业务代码修改，未重复运行与文档变化无关的业务测试。

所有阶段完成后沿独立分支普通 push，发送最终 SHA 与 clean 状态。原会话在本轮结束后进入下一任务的安全接管窗口；现有终端是否 idle 由 Orca 实际状态核实后再建立合法新 Dispatch，不能用工作完成的文字替代终端状态。
