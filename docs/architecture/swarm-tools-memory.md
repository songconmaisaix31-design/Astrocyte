# 蜂群工具层与长期记忆规划

本文件是 W2 的工具层与长期记忆**规划**，按 SPEC §8.4、§13.2 与用户对 Claude-mem 式长期记忆的要求整理，供 W0（契约/迁移/HTTP 唯一 owner）与后续切片消费。它只描述设计与边界，不落地迁移、接口或调度框架；记忆的存储/可见范围尚未获用户答复，相关结论保持 PENDING，不据此冻结表结构或自动采集。

## 1. 共享工具服务（SPEC §8.4）

平台统一提供资料检索、上下文、代码查询、经验和任务工具，保留每个 Agent 自己的私有会话。工具名按 `attention_search`、`workspace_context`、`code_search`、`code_impact`、`swarm_list_work`、`swarm_claim`、`swarm_submit`、`experience_search` 组织，全部调用同一 Service（§13.2）。

- **共享 Service 是唯一事实源，读写分开授权**：所有工具最终落到同一组 Service 方法，工具层只做参数翻译与权限校验，不复刻业务逻辑或索引数据库。读操作（`attention_search`/`code_search`/`swarm_list_work`/`experience_search` 等）走只读路径；写操作（`swarm_claim`/`swarm_submit` 等）复用 Service 的获准写入路径，**逐次按 action/project 授权**，不是只读旁路，也不因“工具存在”而放宽写入授权。
- **HTTP 与 stdio 双通道**：支持 HTTP 的客户端直连本地 MCP 服务；仅支持 stdio 的客户端使用薄桥接（thin bridge），把 stdio JSON-RPC 转接到同一 Service，不各启动一份索引实例。MCP 接入采用官方 Go SDK（[R02]），不手写协议。
- **按任务最小挂载**：工具配置按 `ToolProfile` 编译成对应客户端格式，只给当前任务挂载它需要的工具与资料，不向所有 Agent 广播全部工具定义和资料（§8.4、§13.2）。

### 1.1 ToolProfile

`ToolProfile` 聚合根包含：工具引用、配置覆盖、`secret_ref`、版本；**不保存明文密钥，不混合 Agent 私有状态**（SPEC §6.1）。`secret_ref` 指向平台管理的凭据引用，工具服务在调用时按引用解析，绝不在 profile 里落明文或把供应商 SDK 类型传进业务层。

- 版本锁定：锁定 Agent、MCP SDK、工具 schema 版本；用真实样本做合同测试，未实现的能力留 `unsupported/unknown`，不冒充支持。
- 来源可见：配置使用用户层、项目层、任务层覆盖，页面显示每个值的来源（§6.4）。

### 1.2 权限与提权边界

- **批准/提权不暴露为 Agent 工具**：批准、提权、增加预算、修改委托不进入任何 Agent 工具清单（§8.4、§13.2）。模型可以提出换路线/扩范围的建议，结果只进待批状态（§7.4）。
- **MCP resource 只回范围内快照**：不能通过猜测 ID 获取其他项目数据（§13.2）。每个受保护动作每次检查当前授权，不只相信启动时的状态（§7.3）。
- **Agent MCP 凭据与用户控制面会话分离**：Agent 工具凭据不可调用批准/扩范围/提预算接口（§7.4）。
- 不引入新的调度器、Attempt/Manifest 或完成证明系统；工具层只做授权 + 翻译 + 调用既有 Service。

## 2. 长期记忆规划（Claude-mem 参考）

参考 [thedotmack/claude-mem](https://github.com/thedotmack/claude-mem)：**Apache-2.0**（README 与 LICENSE 已核对），核心是“生命周期 Hook 采集工具使用观察 → 生成语义摘要 → 未来会话可查”，存储用 SQLite（含 FTS5）与 Chroma 向量索引，检索按三层渐进披露 `search → timeline → get_observations`（先拿紧凑索引，再按需取详情，省 token），隐私用 `<private>` 标签排除敏感内容。

本项目**只参考其分层检索与“会话观察→摘要”的思路，不直接读 Codex 平台 memories 或用户 CLI 全部历史当产品记忆**，也不把私人对话自动入库存档。

### 2.1 已确定的边界（不依赖待答项）

- **许可**：claude-mem 为 Apache-2.0，思路与字段命名可复用；不复制其 worker/Bun 运行时、Chroma 依赖或云同步组件。底座继续 Go/SQLite（既有 FTS5 能力），不引入 Python/向量库依赖。
- **只摘要，不原文归档**：会话观察 → 摘要为产品记忆；原始私有对话不进入产品记忆库。摘要只含获准项目/任务范围内的结论、决定、未决项与依据引用，不含凭据、令牌或无关聊天。
- **分层检索**：`search`（紧凑索引）→ `timeline`（时间上下文）→ `detail`（按 ID 取详情），与 claude-mem 三层一致，避免一次性拉全文。
- **授权门槛**：记忆检索只对被授权项目/调用者开放；`experience_search` 等工具经 §1 的共享 Service 与权限校验，不因“记忆存在”而放宽项目隔离。

### 2.2 待答范围（不冻结，不自动采集）

用户已明确**要实现长期记忆**（不再有“仅规划”选项）。尚未回答的只是记忆的**存储/可见范围**：项目隔离 / 个人全局。未答前：

- 不落地 `memory` 迁移表、服务接口或 FTS5 索引结构；
- 不自动采集任何会话或 CLI 历史；
- 不在代码里默认启用“全局记忆”或“自动捕获”。

答复后由原 owner（W2 或其继任）按回答落地，立即实现原生 SQLite FTS5：

- 若**项目隔离**：新增 workspace 项目记忆表 + 服务，复用既有 SQLite FTS5；记忆按 project_id 隔离，检索按项目授权。
- 若**个人全局**：先明确可见范围与授权模型（仅本人/本人可见等），再落地全局记忆，不与项目记忆混写。

### 2.3 与工具层的关系

`experience_search`（记忆检索）是共享工具服务的一个工具，走 §1 的 ToolProfile / secret_ref / 按任务挂载 / 授权路径；记忆不向所有 Agent 广播，只在获准任务里按需暴露。批准/提权不暴露为工具的原则同样适用于记忆检索与写入。

## 3. 当前落地状态与下一步

- 本轮 W2 交付：OpenCode 1.18.35 loopback Basic-auth 会话驱动（`internal/adapters/agents/opencode.go`，配置隔离 + deny-by-default，见 [local-agents.md](../probes/local-agents.md)）；获准 Agent 进度四项返修（settle 返回 typed error、Set 保留 Operations、accepted retry 返回原始快照、成功结果先 durable 再 publish 且 CAS 冲突重试不重跑模型）及 `ProgressCommand` 与 W0 对齐（files + header operation identity）；本工具层与记忆规划。
- 待 W0：发布 `013_project_progress` 迁移、`ProjectProgressService` 的 HTTP 入口与 OpenAPI 契约，并最终集成三轨；`ProgressCommand` 的操作身份来自传输层 `Idempotency-Key` header（`Caller.OperationID`），body 仅 `files`。
- 待用户：记忆存储/可见范围、CLI 完整覆盖范围；未答复不冻结接口、不自动采集、不降低验收。答复后由原 owner（W2 或其继任）按 §2.2 落地：项目隔离则新增 workspace 项目记忆表 + 服务并复用既有 SQLite FTS5，或仅规划。
