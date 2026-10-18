# 2026-10-18 DDD 开发与审查整改

用户要求按照开发规范继续代码开发，采用既有 DDD 分层。基线：管理分支 `validation/v1-20261012` / `734c9b0d0ed357c159bd2127bb63b79e4c6667a9`；产品集成分支 `songconmaisaix31-design/works-relay-integration` / `8c8510426b828359f4c161f93c1edf1083fd5e88`，均干净。本轮 review 原件在根 `.verification/review-20261018/review.md`。当前 64505 无监听，旧服务运行记录按历史保留。正式实验与未答决定依根 QUESTIONS。

本轮使用三个互斥开发轨道。历史开发树和未接受提交保留；为避免历史未结算责任与分支差异影响本次源码，在准确产品基线上创建三个独立工作树。每轨固定同一 Agent 完成实现、真实验证、报告与返修。主 Agent 只维护本计划、状态、决定及验收；开发完成后由唯一集成 Agent 在原产品集成树普通合并，主 Agent 最终检查、提交与普通 push。

| 轨道 | 固定工作树 / 分支 | write_paths 与结果 |
|---|---|---|
| D 领域与接入 | ddd-domain-20261018 / 创建后登记准确分支 | `internal/swarm/**`、`internal/workspace/{app,domain}/relay*.go`、`internal/adapters/sqlite/swarm*.go`、`internal/adapters/agents/relay*.go`、`internal/adapters/events/relay*.go`、`cmd/server/relay*.go`、`docs/standards/ddd-domain-20261018.md`。沿既有 prepared input 与成员端口修正首次 Claim；检查 archive 可信完整性来源和传递，只有来源可证时实施相应接线。消费方端口归 app，外部事实与持久化归 adapters，启动只做组装。未知事实和未答预算不补默认值。 |
| T 检查入口与真实测试 | ddd-validation-20261018 / 创建后登记准确分支 | `.github/workflows/**`、`scripts/check*.mjs`、`docs/standards/ddd-validation-20261018.md`；需要修改其他测试先交准确 Handoff。使 DDD 检查覆盖当前业务上下文，核默认检查所需真实前置；修正可独立验证的工程问题，保持全部原断言、来源范围与验收要求。Native 和模型来源不足时如实失败，不加入 skip、mock 或伪造素材。 |
| P 发布配置 | ddd-release-20261018 / 创建后登记准确分支 | `release/**`、`docs/release/windows.md`、`docs/release/ddd-release-20261018.md`。标准启动入口接通既有 Interpreter deadline 与固定公开上下文配置，成组校验并保持启动时权限和来源检查；使用真实程序核参数拒绝、实际启动及正常关闭。不得启动原生控制器或模型，不重新打包整个模型资产。 |

所有轨道先读根 HANDOFF / STATUS / QUESTIONS、本计划、产品 AGENTS、SPEC §3/§7/§8/§12/§14/§17 和相关原任务。现有 Go/React/SQLite、目录及消费方接口优先复用；禁止新调度器、跨 app 依赖、传输层访问 SQLite、手写成熟格式 parser、程序化编辑代码和视觉操作。原生文件/命令工具关闭要求保持；没有获准真实来源时不运行原生正向测试。Agent 创建的公开项目及实际自有 SQLite/HTTP 可用于其他检查。

每轨提交 SOURCE、普通 push 并记录实际检查及剩余缺口。新增依赖、DDL、HTTP 契约或跨轨修改先交 Handoff，由主 Agent 核既有授权和唯一写入者；不得借工程接线决定未答权限、archive 证明语义或业务默认。输出使用各树已忽略 `.verification/ddd-20261018/`；可复用 D 现有 Go 缓存，禁止读写 `/tmp`、私人会话和正式实验数据。

验收为精确合并源码的 Go vet/build、架构检查、各变更实际 SQLite/HTTP/发布入口验证及适用前端检查。完整默认 CI、L1/L2 正向、最终包和正式 Release 的限制分别登记。当前已确认 native 工具边界、S3 预算与完整能力证据、具体经验批准等保持未答；本轮不以子集通过宣称整体交付完成。实际 Orca Run、Task、Dispatch、客户端及有效模型在启动回执后补充。
