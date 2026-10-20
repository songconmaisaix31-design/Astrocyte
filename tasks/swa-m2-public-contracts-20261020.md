# M2 目标级公共契约与能力目录

用户随后明确首版为抽取 Orca CLI 执行能力的可控 Swarm CLI，当前实施入口转为 [首版任务](swa-cli-v1-20261020.md)。本页保留整体规划的公共 CLI/MCP 契约调查，尚未派发或冻结接口；其范围由后续任务接续。

## 目标与完成条件

依据 [META_PLAN](../docs/META_PLAN.md) §5、§6、M2 和 A02/A03/A07/A08，为 M3 明确服务到 CLI/MCP 的映射，并让能力目录能够由真实契约和注册检查得出。实际实现前由 Worker 补充测试计划，经独立审核后执行。本页当前记录主 Agent 的建议与已读取事实，接口尚未定稿。

## 基线与复用

产品基线为 `32ab1eec808a5848aabad40f11133735a1214e55`，程序来源为 `ed543c12106c2dbb3e181cb9bc1647b626455179`；实施从已接受的 M0/main 接续，准确版本届时记录。使用的技术保持 Go、TypeScript 和 SQLite；资源引用 [RESOURCES](../docs/RESOURCES.md)，有效授权和未决项引用 [QUESTIONS](../docs/QUESTIONS.md)。

主 Agent 已静态读取以下真实实现，Worker 应从当前源码复核：

- `internal/swarm/app/commands.go` 的 `CreateMission` 检查现有 Grant、固定 scope 和幂等回执；调用主体限定 human 或受信 consumer。每个 Grant 的 Mission 使用现有存储查询和唯一性约束。
- `internal/swarm/app/lifecycle.go` 的取消要求已认证 human，任务状态、原生停止、未知结果和资源保留由该服务处理。
- `internal/workspace/app/project_policy.go` 保留人类控制的创建 Mission、取消、批准和采用动作；现有 Agent token 不获得这些动作。
- `internal/adapters/httpapi/session.go` 与 `relay_scoped_workspace.go` 校验真实项目令牌和允许路由；现有 Agent 可以准备和提交 Proposal，不能通过传入主体字段变成人类。
- `internal/workspace/app/s2_service.go` 的批准保存 `proposal_approved` 事件。SPEC 描述 Swarm 按批准事件建立 Mission；当前有限源码搜索尚未找到实际消费组装，不能直接宣称该自动链路存在。
- `internal/swarm/app/queries.go`、HTTP Mission/Artifact 路由和同库组装可复用状态、产物、执行历史及授权查询。
- `cmd/mcp/main.go` 当前组装五项认知工具，目标级任务工具尚未注册。`go.mod` 已声明官方 `github.com/modelcontextprotocol/go-sdk v1.6.1`；新协议入口沿既有 SDK 要求核实实际对外 Schema。

## 建议与写入范围

先核实提交目标、已批准范围、待批准 Proposal、Mission 标识和取消请求的实际业务关系，给出一套成立的映射；用户本人批准与采用继续通过既有可信入口。若完成目标确需改变现有人类专属动作的授权语义，明确差异并提交有效决定，在决定前继续不依赖该部分的查询和目录工作。

公共字段以现有 OpenAPI 为来源；CLI/MCP 保留协议专属注册并按明确名称绑定。目录分别表示软件支持、当前运行环境观察和具体任务事实，未知项有来源和缺失条件。共同业务对象继续由 Workspace/Swarm 拥有；客户端连接结束不结束后台任务。

拟由一名入口模块 Worker 负责契约、目录检查、相关适配和文档，长期维护该模块。准确 write_paths 在派发前按实际方案确定，优先包括本任务、对应独立报告、受影响 SPEC 章节、公共契约、现有生成器与必要注册来源；保持与 M1 的依赖补丁及其报告互斥。任何跨模块业务语义变更先提交主 Agent。

## 验证与接续

Worker 从需求推导必要覆盖：实际注册与目录差异、公共 Schema、未知能力、名称到处理函数的绑定、真实身份拒绝、现有数据读取、幂等和版本边界。检查使用真实 SDK、真实 HTTP 服务和 SQLite；具体哪些场景属于 M2、哪些由 M3 实际入口验收，由测试计划说明。继承已有效的源码证据；契约命令修复由 M1 负责，不改变原规则。

当前 M0 正在实施，M1 测试计划正在独立复验；M2 尚未派发、未冻结接口、未运行检查。独立于 S3 预算和受限原生工具决定的部分继续推进，M4/M6 的真实待决项不因目录生成而获得默认答案。
