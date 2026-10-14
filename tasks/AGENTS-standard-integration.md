# 按当前 AGENTS.md 整合仓库

用户要求：用新标准整合仓库。执行标准为根 [AGENTS.md](../AGENTS.md)；业务事实仍来自对应任务和最新用户决定。接管基线为 validation `c4faaab44c2a8d50423250f7c7864c294a6b6b3f`、Workbench 已验工程 `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`。当前阶段：规则与进度入口整理，随后独立源码整改。

## 范围与边界

- 统一当前规则、规格、规划和接管入口；历史首失败、未知交付和任务要求保留的验收记录原样保存。
- 整改已确认的解析异常继续执行、Python 文件头部 docstring / 英文注释、测试 mock 问题，并盘点同类路径；保留实际功能与有效验收，不通过删除断言、跳过检查或替换成伪造服务取得通过。
- validation 正式实验、锁定 EXPERIMENT、源库 / 快照、schema、引擎与权重仍冻结；运行日程不变。业务变更需要在独立整改分支完成，不混入正式实验运行树。
- 原 Works Relay R/H/F/I 写入范围与返修人员保留；不修改其正在使用的工作树，不合入未完成源码，不重新派发同一开发轨道。
- 执行假设：以当前仓库提交为整改起点，保留 V4 冻结和所有现有开发分支，最终业务提交只进入独立整改分支。是否将 Workbench 合入主开发分支、是否解除冻结均保持未决定，不据此操作。

## 计划与写入范围

主控独占 `HANDOFF.md`、`STATUS.md`、`docs/SPEC.md`、`docs/TASKS.md`、本文件及共享 `.gitignore`，只维护计划、状态、决策与验收。规则入口文档先独立提交并推送；源码轨道从该精确提交建立独立工作树。

| 轨道 | 固定工作树 / 分支 | 独占 write_paths | 验收 |
|---|---|---|---|
| B 后端与导入 | standards-backend / 同名独立分支 | `internal/**`、`cmd/**`、`extensions/**`、`tests/s1/**`、`tests/attention-engine/**`、`scripts/*.mjs`，不含任何 `.py`；`go.mod`、`go.sum`、`dependencies.lock.json` 的必要 parser 依赖；`docs/standards/backend.md` | 实际导入与权限 / 错误路径、适用 Go / Node 测试；冻结引擎 / schema / 权重保持字节不变 |
| P Python | standards-python / 同名独立分支 | `events/*.py`、`scripts/validation/*.py`、`scripts/evidence/*.py`、`scripts/experiment/*.py`、`internal/adapters/importers/*.py`、`tests/validation/**/*.py`、`requirements-validation.txt` 的必要 Python 解析依赖；`docs/standards/python.md` | 文件头部与注释规范、实际 parser / 文件 / SQLite 操作；正式实验脚本只调整注释和 docstring，执行逻辑保持不变 |
| F 前端 | standards-frontend / Orca 同名独立分支 | `web/src/**`、`web/e2e/**`；`docs/standards/frontend.md` | 移除 mock / 伪造 API 响应，保留实际功能与等效断言，使用实际生产 HTTP / 存储完成相关 E2E |

每轨只派发一个 Orca Codex Agent，实际 Task / Dispatch / 工作树 / 分支在收到开工回执后登记。开发、测试、报告及返修均归原轨人员。Go / TS / 依赖 / OpenAPI / 迁移共享变更只提交 Handoff，主控记录后由唯一人员实施。源码轨道完成后另派一个集成 Agent，普通 no-ff 合精确交付并执行组合检查；领域失败退原轨。暂不合 main 或 Workbench 当前开发分支。

共享 Handoff 已处理：开发与 S1 验证的输出目录由 B 改为本树 `.verification/`，失败原件保留；真实服务需要显式限定资料发现根，由 B 在既有配置与组装范围实现并提供早期 SOURCE 给 F。B 已确认 HTML 和 JavaScript 自写 parser；必要的成熟 parser 依赖三文件交 B 单独持有，其余轨道不得修改。用户既有模型调用与获准项目推断授权继续有效：本轮独立技术验证只可使用当前工作树已跟踪资料或明确公开输入，官方 CLI 自行鉴权，不直接读取凭据或私人会话，不启动正式 V4 runner。

当前规则已更新至 `bc5c6aa04e6c92e98091732903328d20a5ac6788`：真实 Codex 验证对象进一步限定为 Agent 创建的公开 Git 项目，原生文件和命令工具关闭，必须使用真实控制器、SQLite / HTTP 及自产会话文件，每个 event_id 逐条定位；全部轨道已收到重读要求。显式目录发现的真实结果被原 OpenAPI source 枚举拒绝，B 追加独占 `contracts/openapi.yaml` 及相关 examples，仅登记真实 `explicit_directory` 来源类型；F 消费该精确契约后在自身写入范围生成 `web/src/api/schema.d.ts`。不修改认知 Event 六字段、数据库 schema、引擎或权重。

P 的非冻结 Git 输出路径保护仍解析 porcelain。该问题继续由原 P 返修：使用成熟 pygit2 仓库 / worktree API，保留所有关联工作树的输出排除，不以仅检查 `.git` 标记缩减布局覆盖。Python 依赖文件只归 P；`dependencies.lock.json` 登记仍只归 B，双方 Handoff。P 已有 SOURCE `55ca989` / REPORT `c592b6a` 保留，下一次源码和报告追加提交。

Orca Run `run_7b2ce96bcd13`，主控自有终端 `term_66b7a627-900a-4604-93e4-3ad1dddf079e`。三轨工作树均从 `4ec19cddb24068e09597357b0aabc6ecb9420930` 建立，位于 `C:/Users/DW/orca/workspaces/Astrocyte/`；Git 分支分别为 `standards-backend`、`standards-python`、`standards-frontend`。三次回执均 ready / input_accepted / turnStart observed，没有重复派发。

- B：Task `task_7b4e93c1a3c1` / Dispatch `ctx_88275b88b389`，终端 `term_b79cd91c-47fa-47b2-ba09-270fb101a22f`；Codex 0.162.0，终端显示 GPT-6.1-Sol high fast / YOLO。
- P：Task `task_bfa0d177ae9a` / Dispatch `ctx_f386e3cc44ac`，终端 `term_0bc0e215-a73c-4333-99c0-3208601e2307`；Orca 观测模型 gpt-6.1-sol。
- F：Task `task_7cbe30017760` / Dispatch `ctx_0b17703feb2c`，终端 `term_63ec6974-000a-4064-b4c9-4eb308f61a3e`；Orca 观测模型 gpt-6.1-sol。

早期交付：B SOURCE `86d8f910eb5d6641f84bcc248d86444fb7bd81cb` 已 push，提供显式发现目录和实际 HTTP / SQLite 验证生命周期，已 Handoff 给 F 普通消费；B 继续 parser 和同类整改。P SOURCE `55ca9893c74fd25b153a0560f52428a322722ebe` / REPORT `c592b6ac8cec5a634f16df3e3c18bd2761499cc3` 已 push，69 项适用验证通过、14 个冻结脚本 AST 独立复核全等。P 新 Task `task_7ea3f8d0332b` / Dispatch `ctx_71b0ea451645` 已在同一终端 / 工作树 / 分支续作 Git 解析返修，未增加第二人员。

真实范围限制：Mission 生产创建接口仍返回 501，原模拟运行态缺少实际来源；已向用户询问是否扩展到真实创建与执行，未答不新增该业务。冻结实验测试中 28 项完整模型 / 外部来源路径尚未执行，原替换与手写 SSE / Markdown parser 的行为整改仍受冻结约束。相关限制不能由适用子集测试填为 PASS。

## 验收

1. 当前规范没有允许 mock 或虚假数据的冲突条款；入口链接指向实际存在的文件。
2. 每项业务整改执行相关实际路径、适用测试和构建；没有完成的项目保持 NOT_RUN / BLOCKED，不用文档检查代替。
3. 记录各轨 SOURCE / REPORT 与实际命令结果；保持普通 Git 历史，最终分支推送，不 force push。
4. 最终报告只记录完成内容、分支 / SHA、验证结果、真实限制和未执行操作。

## 当前状态

2026-10-14 网络中断恢复：继续原 Run `run_7b2ce96bcd13`，主控自己的终端经当前屏幕命令及执行进程身份核实为 `term_c26f9d91-dde4-49ea-a94c-4d472b19929f`，Run 已绑定 generation 3。运行时将 B / F 原 Dispatch 登记为 abandoned / failed、Task 恢复为 ready；后续核查确认原 B 在 `term_1a398364-d5de-42cc-a63c-6fe34c4bc749` 继续运行，其 Orca 工作树元数据为根目录，实际执行目录为 standards-backend。新增 B 已 accepted worker_done failed，并 release / closed_agent_terminal；双方修改完整保留。原 B 同 Task 新 Dispatch `ctx_5a0a4d492434` 继续唯一写入，注入首次停在输入框而 turn_start_unobserved，主控确认输入框后仅补按 Enter，原 B 已实际收到新 preamble 并开始执行。F 在原工作树继续独立开发。两轨完成后，由唯一集成人员普通合并精确 SOURCE / REPORT，执行组合检查并推送独立整改分支。

恢复回执：F Dispatch `ctx_a69aaadfe5b6` / 终端 `term_54201688-6be7-42d0-a42a-15e6ea3b6631`，ready / input_accepted / turnStart observed。新增 B `ctx_e0356fbca4e1` / 终端 `term_37307630-9a25-480e-b6d4-2bc830dbc3d1` 同样实际启动，发现原 B 持续写入后停止修改；其新增 native 格式校验及错误传播仍保留在原工作树，原 B 后续统一审阅，Go 整包首次 RED、Node 12 PASS、events / cmd/validation PASS 和公共来源网络失败证据均保留。P 已结算，但旧终端清理为 `release_unknown`；按回执检查并重试后仍无法证明旧进程状态，未使用批量终端关闭。

P 已完成返修：SOURCE `a1089f94a13306e3cd250ed8facbb438d6e75a19` / REPORT `91e3e59dcfc8024cd2e4955ac72f627e65c18e71`。原 `worker_done` 已被编排接受；报告记载本次 29 项证据测试与 25 项安全测试 PASS、36 文件规范检查及 14 冻结 AST 全等。主控将核对远端、修改路径与适用独立验证，不将该子集计作完整实验验收。Mission 扩展决定仍未收到；V4 / Discovery & Split 冻结、正式数据和日程继续按既有边界处理。

主控已独立核对 P REPORT 远端一致及修改路径在原授权范围；重新执行 `python -m unittest discover -s tests/validation/evidence -v`，29 PASS（16.906 秒），`python tests/validation/experiment/run_safe_tests.py`，25 PASS（0.301 秒），`python tests/validation/experiment/check_python.py 4ec19cddb24068e09597357b0aabc6ecb9420930`，36 文件 / 14 冻结 AST PASS。后续只因新修改、失败或组合疑点追加检查；原完整实验 28 项及真实模型路径仍 NOT_RUN。F 已回执实际终端模型 GPT-6.1-Sol high fast，继续消费 B 精确契约与真实服务，原浏览器首次失败证据保留。

已读取当前交接 / 状态 / 规格 / 规划、原 Workbench 计划及 Orca 编排记录。已核实原 R/H/F/I 轨道存在；本轮未接管或变更其生命周期。规则与入口阶段提交 `4ec19cd` 已 push，链接和 diff 检查 PASS；三条整改轨道已接受任务，源码与组合验收尚未完成。中间结果仅放在各自工作树 `.verification/`；正式原件不读取、不复制、不输出。
