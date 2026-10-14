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
| B 后端与导入 | standards-backend / Orca 同名独立分支 | `internal/**`、`cmd/**`、`extensions/**`、`tests/s1/**`、`tests/attention-engine/**`、`scripts/*.mjs`，不含任何 `.py`；`docs/standards/backend.md` | 实际导入与权限 / 错误路径、适用 Go / Node 测试；冻结引擎 / schema / 权重保持字节不变 |
| P Python | standards-python / Orca 同名独立分支 | `events/*.py`、`scripts/validation/*.py`、`scripts/evidence/*.py`、`scripts/experiment/*.py`、`internal/adapters/importers/*.py`、`tests/validation/**/*.py`；`docs/standards/python.md` | 文件头部与注释规范、实际 parser / 文件 / SQLite 操作；正式实验脚本只调整注释和 docstring，执行逻辑保持不变 |
| F 前端 | standards-frontend / Orca 同名独立分支 | `web/src/**`、`web/e2e/**`；`docs/standards/frontend.md` | 移除 mock / 伪造 API 响应，保留实际功能与等效断言，使用实际生产 HTTP / 存储完成相关 E2E |

每轨只派发一个 Orca Codex Agent，实际 Task / Dispatch / 工作树 / 分支在收到开工回执后登记。开发、测试、报告及返修均归原轨人员。Go / TS / 依赖 / OpenAPI / 迁移共享变更只提交 Handoff，主控记录后由唯一人员实施。源码轨道完成后另派一个集成 Agent，普通 no-ff 合精确交付并执行组合检查；领域失败退原轨。暂不合 main 或 Workbench 当前开发分支。

## 验收

1. 当前规范没有允许 mock 或虚假数据的冲突条款；入口链接指向实际存在的文件。
2. 每项业务整改执行相关实际路径、适用测试和构建；没有完成的项目保持 NOT_RUN / BLOCKED，不用文档检查代替。
3. 记录各轨 SOURCE / REPORT 与实际命令结果；保持普通 Git 历史，最终分支推送，不 force push。
4. 最终报告只记录完成内容、分支 / SHA、验证结果、真实限制和未执行操作。

## 当前状态

已读取当前交接 / 状态 / 规格 / 规划、原 Workbench 计划及 Orca 编排记录。已核实原 R/H/F/I 轨道存在；本轮未接管或变更其生命周期。规则和入口链接校核 PASS，源码、组合测试与部署尚未执行。中间结果仅放在各自工作树 `.verification/`；正式原件不读取、不复制、不输出。
