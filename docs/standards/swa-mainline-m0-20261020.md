# SWA M0 来源整合与验证

## 范围与准确来源

用户要求以 `C:/Users/DW/Desktop/re.md` 为准重构主线；[META_PLAN](../META_PLAN.md) 保存 v2.1 完整正文，仅移除外围代码围栏。本任务按[已独立接受的 D01—D08](../../tasks/swa-m0-20261020.md)实施文档部分，产品程序、依赖、契约、迁移、数据和冻结实验保持原样。

用户最新首版目标为抽取 Orca CLI 对应的执行能力，交付从原有 Agent 或终端调用的可控 Swarm CLI；主控经 `msg_8143bd0c98ad` 转达后，当前入口与 QUESTIONS 已记录适用关系。M2—M6 继续作为后续规划，MCP 与完整 S3 不自动成为该首版前置；原权限条件保持，当前任务不定稿执行接口。

| 来源 | 准确值与用途 |
| --- | --- |
| 规范仓库 | 真实 GitHub API 仓库 ID `1410836101`，`songconmaisaix31-design/swarm-anything`，默认 `main` |
| 默认 main 快照 | `3a02ce38481d1885adee2e051a1b524d78376897`；产品32ab祖先 |
| 产品基线 | `32ab1eec808a5848aabad40f11133735a1214e55`，产品和原前端分支；M0 工作树从此创建 |
| 已验程序源码 S | `ed543c12106c2dbb3e181cb9bc1647b626455179`，既有构建与真实运行验证；本轮复用原限定结果 |
| 主控接管快照 | `c8df672770ddfd5e9f53ea64fdcd932d2e4bff76` |
| 主控接续来源 | `cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`；包含本轮规划、最新授权与主控独有文档 |
| 产品与主控共同祖先 | `38fc8ff4d5ae48cd5d1af02ac80b0d5eebeb44a2` |
| 当前分支及目标 | `swa-m0-20261020` → `songconmaisaix31-design/works-relay-integration`；main 由主控随后普通集成 |

上表为 2026-10-20 的实际 API、远端和准确 Git 核查快照。改名已完成；本轮不修改 Go module、应用配置与数据目录。主控相对共同祖先没有独有产品程序变更，按内容接续文档；不以主控旧源码覆盖产品。

本轮再次读取真实 API/远端：仓库 ID、名称、默认分支和 main/产品/前端 SHA 保持上表值；主控新增 `2fefbe79075bb94660059fcedd88b8008931b92f`，仅追加 [M2 建议任务](../../tasks/swa-m2-public-contracts-20261020.md)，已依主控 `msg_fdd7705d3f50` 原样纳入。该任务尚未实施，首版范围仍按最新用户决定。

主控随后提供 `9140cf8aa618c832b3a099222a46991b7c2b3ddd` 的最新六份文档。已重新读取并将首版 CLI 任务、M2 延后说明、QUESTIONS 的最新决定及独立运行待答、总任务和 STATUS/HANDOFF 的接续内容纳入当前入口，保留本轮已修订的规则、来源和导航。首版是否独立运行或依赖已安装 Orca 尚未回答，执行接口保持未决定。

| 追加来源 | 当前目标与消费者 | 验证 |
| --- | --- | --- |
| `9140cf8a:tasks/swa-cli-v1-20261020.md` | [首版 CLI 事实源](../../tasks/swa-cli-v1-20261020.md)，README、HANDOFF、STATUS、TASKS及文档目录 | 完整来源文本比较 |
| `9140cf8a:tasks/swa-m2-public-contracts-20261020.md` | [后续 M2 建议](../../tasks/swa-m2-public-contracts-20261020.md)，原2fef内容及最新延后说明 | 完整来源文本比较 |
| `9140cf8a:docs/QUESTIONS.md`、`tasks/swa-mainline-20261020.md`、`STATUS.md`、`HANDOFF.md` | 对应当前文件中最新首版目标、待答边界、范围与接续；保留已完成的M0整合 | 完整读取最新文件并逐项核对新决定及直接消费者 |

## §7.5 的现行规则

| 规划要求 | 规则拥有者及直接消费者 | 实施结果 |
| --- | --- | --- |
| 依赖方向 | SPEC §3.1 → AGENTS、CODING、角色、CONTRIBUTING | Domain 纯逻辑；App 声明端口；Adapters 实现；Composition 注入；协议入口共用 App |
| Python 现行决定 | QUESTIONS → SPEC §9/§11/§15/AT14/§21、CODING、AGENTS | 允许明确需要的工具链，保留核心结构；研究运行时独立 |
| 协作按实际写域 | AGENTS → 主 Agent/Worker、TASKS | 去除固定重复百分比，跨模块契约和实际 write_paths 决定协作 |
| 决策分级 | AGENTS → 角色、CODING | 普通命名及局部实现由负责人处理；真实目标、权限、验收、预算、外发和不可逆问题保留待答 |
| 验证按变化 | AGENTS → SPEC、角色、CONTRIBUTING | 独立测试计划审核及必要既有检查保持；不机械复制全部测试层 |
| 模块与任务分支 | AGENTS → SPEC、角色、CONTRIBUTING | 模块责任连续，任务分支交付后集成并结束 |
| 文档事实所属 | STATUS / RESOURCES / QUESTIONS → HANDOFF、README、docs/README | 当前事实、资源参数、用户决定分别维护；docs/STATUS 只作入口 |
| 在规则边界修正 | SPEC / CODING → AGENTS、角色 | 修改负责解释的边界，并更新直接消费者 |
| 稳定规则与临时状态 | AGENTS → 当前任务、QUESTIONS、STATUS | 临时派发、授权和实际模型登记在任务；AGENTS 只留稳定执行规则 |

真实 Codex 控制器、SQLite、HTTP、实际产生的会话文件及逐个 event_id 定位要求保留；原生文件和命令工具关闭、公开固定 Agent 项目范围、未知不重发、禁止主动视觉和用户自由工具未决定等边界保持。S3 预算、权限证据、固定资料准入、完成证据读取、经验注入、Laya 语言、首发范围、TeX 并发及作者许可继续按 QUESTIONS 原记录处理。

## 逐项纳入主控独有文档

以下每行的原提交均为完整 `cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`，原路径与目标路径相同。归档内容保持源文件完整文本；成熟 Markdown 解析器检查本次新增 Markdown 导航，归档 txt 按原文保存。

| 原提交与路径 | 当前目标 | 消费者 | 验证 |
| --- | --- | --- | --- |
| `cf22:.agents/skills/astrocyte-research/SKILL.md` | [.agents/skills/astrocyte-research/SKILL.md](../../.agents/skills/astrocyte-research/SKILL.md) | 科研角色、原研究任务及产品反馈 | 完整文本比较 |
| `cf22:docs/META_PLAN.md` | [docs/META_PLAN.md](../../docs/META_PLAN.md) | 当前规则与M0—M6任务 | 完整文本比较 |
| `cf22:docs/archive/20261020/AGENTS.txt` | [docs/archive/20261020/AGENTS.txt](../../docs/archive/20261020/AGENTS.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/20261020/HANDOFF.txt` | [docs/archive/20261020/HANDOFF.txt](../../docs/archive/20261020/HANDOFF.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/20261020/README.md` | [docs/archive/20261020/README.md](../../docs/archive/20261020/README.md) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/20261020/STATUS.txt` | [docs/archive/20261020/STATUS.txt](../../docs/archive/20261020/STATUS.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/20261020/docs-STATUS.txt` | [docs/archive/20261020/docs-STATUS.txt](../../docs/archive/20261020/docs-STATUS.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/README.md` | [docs/archive/superseded-20261020/README.md](../../docs/archive/superseded-20261020/README.md) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/README.txt` | [docs/archive/superseded-20261020/README.txt](../../docs/archive/superseded-20261020/README.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/docs/SPEC.txt` | [docs/archive/superseded-20261020/docs/SPEC.txt](../../docs/archive/superseded-20261020/docs/SPEC.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/docs/TASKS.txt` | [docs/archive/superseded-20261020/docs/TASKS.txt](../../docs/archive/superseded-20261020/docs/TASKS.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/B0-baseline.txt` | [docs/archive/superseded-20261020/tasks/B0-baseline.txt](../../docs/archive/superseded-20261020/tasks/B0-baseline.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/M6-V1R-integration.txt` | [docs/archive/superseded-20261020/tasks/M6-V1R-integration.txt](../../docs/archive/superseded-20261020/tasks/M6-V1R-integration.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/M6-V23-integration.txt` | [docs/archive/superseded-20261020/tasks/M6-V23-integration.txt](../../docs/archive/superseded-20261020/tasks/M6-V23-integration.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/M6-V4-ci-repair.txt` | [docs/archive/superseded-20261020/tasks/M6-V4-ci-repair.txt](../../docs/archive/superseded-20261020/tasks/M6-V4-ci-repair.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/M6-V4-integration.txt` | [docs/archive/superseded-20261020/tasks/M6-V4-integration.txt](../../docs/archive/superseded-20261020/tasks/M6-V4-integration.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/M6-V4-start.txt` | [docs/archive/superseded-20261020/tasks/M6-V4-start.txt](../../docs/archive/superseded-20261020/tasks/M6-V4-start.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/M6-validation-plan.txt` | [docs/archive/superseded-20261020/tasks/M6-validation-plan.txt](../../docs/archive/superseded-20261020/tasks/M6-validation-plan.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/S0-plan.txt` | [docs/archive/superseded-20261020/tasks/S0-plan.txt](../../docs/archive/superseded-20261020/tasks/S0-plan.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/UI-preview-plan.txt` | [docs/archive/superseded-20261020/tasks/UI-preview-plan.txt](../../docs/archive/superseded-20261020/tasks/UI-preview-plan.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/Works-Relay-v1.1.txt` | [docs/archive/superseded-20261020/tasks/Works-Relay-v1.1.txt](../../docs/archive/superseded-20261020/tasks/Works-Relay-v1.1.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/ddd-review-remediation-20261018.txt` | [docs/archive/superseded-20261020/tasks/ddd-review-remediation-20261018.txt](../../docs/archive/superseded-20261020/tasks/ddd-review-remediation-20261018.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/integration-cleanup-20261017.txt` | [docs/archive/superseded-20261020/tasks/integration-cleanup-20261017.txt](../../docs/archive/superseded-20261020/tasks/integration-cleanup-20261017.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/tasks/swarm-send-lease-pr-20261018.txt` | [docs/archive/superseded-20261020/tasks/swarm-send-lease-pr-20261018.txt](../../docs/archive/superseded-20261020/tasks/swarm-send-lease-pr-20261018.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/archive/superseded-20261020/web/UI-preview-report.txt` | [docs/archive/superseded-20261020/web/UI-preview-report.txt](../../docs/archive/superseded-20261020/web/UI-preview-report.txt) | 历史任务、决定及原始状态读取 | 完整文本比较 |
| `cf22:docs/business/ASTROCYTE-BP.md` | [docs/business/ASTROCYTE-BP.md](../../docs/business/ASTROCYTE-BP.md) | 原商业提案查询 | 完整文本比较 |
| `cf22:docs/documentation-audit-20261020.md` | [docs/documentation-audit-20261020.md](../../docs/documentation-audit-20261020.md) | 文档导航与主控接续 | 完整文本比较 |
| `cf22:docs/experiment/README.md` | [docs/experiment/README.md](../../docs/experiment/README.md) | 正式V4接续 | 完整文本比较 |
| `cf22:docs/integration-audit-20261020.md` | [docs/integration-audit-20261020.md](../../docs/integration-audit-20261020.md) | 文档导航与主控接续 | 保留原正文并增加M0接续；人工复核 |
| `cf22:docs/modules/daily-operation-20261015.md` | [docs/modules/daily-operation-20261015.md](../../docs/modules/daily-operation-20261015.md) | 模块交接及已实现接口查询 | 完整文本比较 |
| `cf22:docs/modules/opencode-model-20261015.md` | [docs/modules/opencode-model-20261015.md](../../docs/modules/opencode-model-20261015.md) | 模块交接及已实现接口查询 | 完整文本比较 |
| `cf22:docs/modules/product.md` | [docs/modules/product.md](../../docs/modules/product.md) | 模块交接及已实现接口查询 | 完整文本比较 |
| `cf22:docs/modules/relay.md` | [docs/modules/relay.md](../../docs/modules/relay.md) | 模块交接及已实现接口查询 | 完整文本比较 |
| `cf22:docs/modules/release.md` | [docs/modules/release.md](../../docs/modules/release.md) | 模块交接及已实现接口查询 | 完整文本比较 |
| `cf22:docs/research/orca-rea-20261014.md` | [docs/research/orca-rea-20261014.md](../../docs/research/orca-rea-20261014.md) | 科研角色、原研究任务及产品反馈 | 完整文本比较 |
| `cf22:docs/research/research-product-decisions.md` | [docs/research/research-product-decisions.md](../../docs/research/research-product-decisions.md) | 科研角色、原研究任务及产品反馈 | 完整文本比较 |
| `cf22:tasks/Works-Relay-v1.4.md` | [tasks/Works-Relay-v1.4.md](../../tasks/Works-Relay-v1.4.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/ddd-review-remediation-20261018.md` | [tasks/ddd-review-remediation-20261018.md](../../tasks/ddd-review-remediation-20261018.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/delivery-candidate-20261020.md` | [tasks/delivery-candidate-20261020.md](../../tasks/delivery-candidate-20261020.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/documentation-cleanup-20261020.md` | [tasks/documentation-cleanup-20261020.md](../../tasks/documentation-cleanup-20261020.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/documentation-conflicts-20261020.md` | [tasks/documentation-conflicts-20261020.md](../../tasks/documentation-conflicts-20261020.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/dynamic-dag-loop-20261019.md` | [tasks/dynamic-dag-loop-20261019.md](../../tasks/dynamic-dag-loop-20261019.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/frontend-observatory-pr-20261018.md` | [tasks/frontend-observatory-pr-20261018.md](../../tasks/frontend-observatory-pr-20261018.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/integration-cleanup-20261017.md` | [tasks/integration-cleanup-20261017.md](../../tasks/integration-cleanup-20261017.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/integration-convergence-20261020.md` | [tasks/integration-convergence-20261020.md](../../tasks/integration-convergence-20261020.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/research-preprint-20261019.md` | [tasks/research-preprint-20261019.md](../../tasks/research-preprint-20261019.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/research-preprint-review-20261019.md` | [tasks/research-preprint-review-20261019.md](../../tasks/research-preprint-review-20261019.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/restart-product-research-20261019.md` | [tasks/restart-product-research-20261019.md](../../tasks/restart-product-research-20261019.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/s3-paper-ten-cycles-20261019.md` | [tasks/s3-paper-ten-cycles-20261019.md](../../tasks/s3-paper-ten-cycles-20261019.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |
| `cf22:tasks/swa-mainline-20261020.md` | [tasks/swa-mainline-20261020.md](../../tasks/swa-mainline-20261020.md) | TASKS、当前接续及历史任务引用 | 保留原正文并增加M0接续；人工复核 |
| `cf22:tasks/swarm-send-lease-pr-20261018.md` | [tasks/swarm-send-lease-pr-20261018.md](../../tasks/swarm-send-lease-pr-20261018.md) | TASKS、当前接续及历史任务引用 | 完整文本比较 |

## 既有任务的有效主控记录

下表接续主控中的适用范围、来源或归档入口，当前任务及实现状态由 TASKS/STATUS 判断。旧文本及原命令在所链接归档或准确 Git 历史保留，原人员、日期与额度不自动成为新派发。

| 当前目标 | 来源 | 消费者与处置 | 验证 |
| --- | --- | --- | --- |
| [tasks/AGENTS-standard-integration.md](../../tasks/AGENTS-standard-integration.md) | `cf22:tasks/AGENTS-standard-integration.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W0.md](../../tasks/AttentionEngine-W0.md) | `cf22:tasks/AttentionEngine-W0.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W1.md](../../tasks/AttentionEngine-W1.md) | `cf22:tasks/AttentionEngine-W1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W2.md](../../tasks/AttentionEngine-W2.md) | `cf22:tasks/AttentionEngine-W2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W3.md](../../tasks/AttentionEngine-W3.md) | `cf22:tasks/AttentionEngine-W3.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W4.md](../../tasks/AttentionEngine-W4.md) | `cf22:tasks/AttentionEngine-W4.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W5.md](../../tasks/AttentionEngine-W5.md) | `cf22:tasks/AttentionEngine-W5.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W6.md](../../tasks/AttentionEngine-W6.md) | `cf22:tasks/AttentionEngine-W6.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-W7.md](../../tasks/AttentionEngine-W7.md) | `cf22:tasks/AttentionEngine-W7.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-integration.md](../../tasks/AttentionEngine-integration.md) | `cf22:tasks/AttentionEngine-integration.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/AttentionEngine-plan.md](../../tasks/AttentionEngine-plan.md) | `cf22:tasks/AttentionEngine-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/B0-baseline.md](../../tasks/B0-baseline.md) | `cf22:tasks/B0-baseline.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/L1-session-capture-plan.md](../../tasks/L1-session-capture-plan.md) | `cf22:tasks/L1-session-capture-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/L2-work-experiment-draft.md](../../tasks/L2-work-experiment-draft.md) | `cf22:tasks/L2-work-experiment-draft.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-EX01-decision.md](../../tasks/M6-EX01-decision.md) | `cf22:tasks/M6-EX01-decision.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-EX01.md](../../tasks/M6-EX01.md) | `cf22:tasks/M6-EX01.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V1-replay.md](../../tasks/M6-V1-replay.md) | `cf22:tasks/M6-V1-replay.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V1R-integration.md](../../tasks/M6-V1R-integration.md) | `cf22:tasks/M6-V1R-integration.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V1R-replay.md](../../tasks/M6-V1R-replay.md) | `cf22:tasks/M6-V1R-replay.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V2-evidence.md](../../tasks/M6-V2-evidence.md) | `cf22:tasks/M6-V2-evidence.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V23-integration.md](../../tasks/M6-V23-integration.md) | `cf22:tasks/M6-V23-integration.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V3-experiment.md](../../tasks/M6-V3-experiment.md) | `cf22:tasks/M6-V3-experiment.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V4-ci-repair.md](../../tasks/M6-V4-ci-repair.md) | `cf22:tasks/M6-V4-ci-repair.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V4-first-day-recovery.md](../../tasks/M6-V4-first-day-recovery.md) | `cf22:tasks/M6-V4-first-day-recovery.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V4-integration.md](../../tasks/M6-V4-integration.md) | `cf22:tasks/M6-V4-integration.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V4-live.md](../../tasks/M6-V4-live.md) | `cf22:tasks/M6-V4-live.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V4-raw-exemption.md](../../tasks/M6-V4-raw-exemption.md) | `cf22:tasks/M6-V4-raw-exemption.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-V4-start.md](../../tasks/M6-V4-start.md) | `cf22:tasks/M6-V4-start.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-W0.md](../../tasks/M6-W0.md) | `cf22:tasks/M6-W0.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-W1.md](../../tasks/M6-W1.md) | `cf22:tasks/M6-W1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-W2.md](../../tasks/M6-W2.md) | `cf22:tasks/M6-W2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-integration.md](../../tasks/M6-integration.md) | `cf22:tasks/M6-integration.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-validation-plan.md](../../tasks/M6-validation-plan.md) | `cf22:tasks/M6-validation-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-validation-v2.1.md](../../tasks/M6-validation-v2.1.md) | `cf22:tasks/M6-validation-v2.1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-validation-v2.2.md](../../tasks/M6-validation-v2.2.md) | `cf22:tasks/M6-validation-v2.2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-validation-v2.md](../../tasks/M6-validation-v2.md) | `cf22:tasks/M6-validation-v2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/M6-validation-v3.md](../../tasks/M6-validation-v3.md) | `cf22:tasks/M6-validation-v3.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/Release-v0.1-20261014.md](../../tasks/Release-v0.1-20261014.md) | `cf22:tasks/Release-v0.1-20261014.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S0-plan.md](../../tasks/S0-plan.md) | `cf22:tasks/S0-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-W0.md](../../tasks/S1-W0.md) | `cf22:tasks/S1-W0.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-W1.md](../../tasks/S1-W1.md) | `cf22:tasks/S1-W1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-W2.md](../../tasks/S1-W2.md) | `cf22:tasks/S1-W2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-W3.md](../../tasks/S1-W3.md) | `cf22:tasks/S1-W3.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-W4.md](../../tasks/S1-W4.md) | `cf22:tasks/S1-W4.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-contract.md](../../tasks/S1-contract.md) | `cf22:tasks/S1-contract.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-final-W0.md](../../tasks/S1-final-W0.md) | `cf22:tasks/S1-final-W0.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-final-W1.md](../../tasks/S1-final-W1.md) | `cf22:tasks/S1-final-W1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-final-W3.md](../../tasks/S1-final-W3.md) | `cf22:tasks/S1-final-W3.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-final-paper-cli-memory-plan.md](../../tasks/S1-final-paper-cli-memory-plan.md) | `cf22:tasks/S1-final-paper-cli-memory-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-paper-W0.md](../../tasks/S1-paper-W0.md) | `cf22:tasks/S1-paper-W0.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-paper-W1.md](../../tasks/S1-paper-W1.md) | `cf22:tasks/S1-paper-W1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-paper-W2.md](../../tasks/S1-paper-W2.md) | `cf22:tasks/S1-paper-W2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-paper-W3.md](../../tasks/S1-paper-W3.md) | `cf22:tasks/S1-paper-W3.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-paper-project-board-plan.md](../../tasks/S1-paper-project-board-plan.md) | `cf22:tasks/S1-paper-project-board-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-plan.md](../../tasks/S1-plan.md) | `cf22:tasks/S1-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-repository-browser-design-plan.md](../../tasks/S1-repository-browser-design-plan.md) | `cf22:tasks/S1-repository-browser-design-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-summarize-plan.md](../../tasks/S1-summarize-plan.md) | `cf22:tasks/S1-summarize-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-sync-W0.md](../../tasks/S1-sync-W0.md) | `cf22:tasks/S1-sync-W0.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-sync-W1.md](../../tasks/S1-sync-W1.md) | `cf22:tasks/S1-sync-W1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-sync-W2.md](../../tasks/S1-sync-W2.md) | `cf22:tasks/S1-sync-W2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-sync-W3.md](../../tasks/S1-sync-W3.md) | `cf22:tasks/S1-sync-W3.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-sync-local-agent-plan.md](../../tasks/S1-sync-local-agent-plan.md) | `cf22:tasks/S1-sync-local-agent-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/S1-video-fix-plan.md](../../tasks/S1-video-fix-plan.md) | `cf22:tasks/S1-video-fix-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/UI-preview-plan.md](../../tasks/UI-preview-plan.md) | `cf22:tasks/UI-preview-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/Workbench-v01.md](../../tasks/Workbench-v01.md) | `cf22:tasks/Workbench-v01.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/Works-Relay-v1.1.md](../../tasks/Works-Relay-v1.1.md) | `cf22:tasks/Works-Relay-v1.1.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/Works-Relay-v1.2.md](../../tasks/Works-Relay-v1.2.md) | `cf22:tasks/Works-Relay-v1.2.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/cognitive-stack-L1-L4.md](../../tasks/cognitive-stack-L1-L4.md) | `cf22:tasks/cognitive-stack-L1-L4.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/v0.2-discovery-split-plan.md](../../tasks/v0.2-discovery-split-plan.md) | `cf22:tasks/v0.2-discovery-split-plan.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/v0.2-discovery-split-review.md](../../tasks/v0.2-discovery-split-review.md) | `cf22:tasks/v0.2-discovery-split-review.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |
| [tasks/v0.2-discovery-split.md](../../tasks/v0.2-discovery-split.md) | `cf22:tasks/v0.2-discovery-split.md` | 原任务调用者；来源、归档入口或有效续行记录 | 完整文本比较 |

[能力板任务](../../tasks/v0.3-capability-board.md)保留产品32ab中已确认的六步定义，仅增加适用范围说明；主控旧“六步尚待给出”措辞不作为现行状态。[QUESTIONS](../QUESTIONS.md)接续主控完整决定并增加 §7.5 替代关系；产品32ab中的权限及未验摘要均由完整原记录覆盖，Windows契约任务继续在 TASKS 中登记。

## 保留与未迁移范围

- 产品独有的 docs/access、docs/release、docs/standards、docs/workbench、发布说明及任务全部保留，准确32ab程序和所有非文档文件保持无差异。
- 主控对旧验收、探针、实验操作文档及 web/UI-preview-report 的范围提示不覆盖产品原件；当前状态由唯一入口说明。docs/EXPERIMENT.md、现有 docs/experiment 原件和 docs/RESOURCES.md 保持准确32ab内容。
- 原主控归档的来源说明和 txt 全部接续；docs/STATUS 的产品32ab原文另存 [M0归档](../archive/swa-m0-20261020/README.md)。AGENTS/CODING/SPEC/角色的前态仍由准确32ab Git 历史完整保存。
- 研究论文、Morphogenesis 运行时、正式实验程序与数据库保留原位置，仅接续任务、反馈与科研 Skill。原独立审查分支的报告不伪装成本分支自产结果。
- 其他工作树未提交作品、外部数据库与本地验证原件保持原位置；本轮没有复制私有原件、改动资源参数、启动服务或创建发布。

## 实际验证

静态文档检查使用已安装 `marked@18.1.0`，BelowNormal，Node heap 512 MiB，逐项串行；临时检查文件位于已获准且忽略的 `.verification/swa-m0-20261020/`。仅检查文件、链接与 Git 内容，不触发 Go、业务测试、浏览器或模型。

| 检查 | 实际命令与结果 |
| --- | --- |
| D01 真实来源 | `gh api repos/songconmaisaix31-design/swarm-anything --jq '{id,full_name,default_branch}'`、`git ls-remote origin refs/heads/main refs/heads/songconmaisaix31-design/works-relay-integration refs/heads/songconmaisaix31-design/frontend-observatory-20261018 refs/heads/validation/v1-20261012`：退出0，与上表来源一致；`git merge-base` 和祖先查询确认历史关系 |
| D02 原文 | `node --max-old-space-size=512 .verification/swa-m0-20261020/check-documents.mjs`：退出0，re.md移除外围围栏后与 META_PLAN 完整一致 |
| D03 规则 | 逐项读取 AGENTS/CODING/SPEC/四角色，结合旧分层、固定百分比、Python绝对禁令及全部三层测试要求的搜索核对：九项修订均有当前负责位置及直接消费者；Python文件格式和禁止脚本改写的规则继续有效 |
| D04 决定 | 完整读取 cf22 QUESTIONS、RESOURCES、状态及对应任务；保留原决定正文、六步已确认定义与未决索引；新增首版决定明确替代原计划的首版范围和顺序，资源文件不变 |
| D05/D08 保全与边界 | 同一静态检查：产品1446个路径全部保留，非文档修改为0；现有实验、资源参数和产品独有报告无差异；51份cf22独有文档及9140首版CLI/M2任务已纳入，121份来源文本比较通过 |
| D06 链接与格式 | UTC `2026-10-20T07:04:03.677Z` 同一检查：123份变更Markdown、791个本地链接、34项引用定义、57个代码围栏，错误0；本次文档没有本地标题锚点引用，锚点数量为0 |
| D07 接管阅读 | README → CONTRIBUTING/AGENTS → HANDOFF/STATUS → TASKS/总任务/M0 → SPEC/QUESTIONS/RESOURCES，以及研究、V4和归档入口逐项读取；已实现能力、规划与原 FAIL/NOT_RUN/BLOCKED 没有混写 |
| D08 格式与PR | `git diff --check`：退出0；首份提交普通push、远端HEAD一致；[PR #4](https://github.com/songconmaisaix31-design/swarm-anything/pull/4)的base为产品集成分支、head为swa-m0-20261020；后续来源接续普通提交后核最新HEAD，准确实现独立CR另行记录 |

检查结果原件为忽略目录中的 `result.json`。纯文档工作不将既有产品 FAIL、NOT_RUN 或 BLOCKED 改记为通过。D01/D03/D04/D07 的语义核对与 D02/D05/D06/D08 的机器检查共同支持 M0 文档结论。准确实现的独立 CR 与最终 main 集成尚待对应流程。

## 真实剩余限制

Windows原契约命令仍 FAIL，完整默认CI仍 NOT_RUN，真实原生/S3、两 Agent 闭环及实际采用未通过完整验收；研究技术交付、正式数据与许可缺项及 V4 原根/权威锁限制保持。M0 不构成 M1—M6 或正式 Release 的验收。main 普通集成由主控执行，当前任务只交付文档 PR。
