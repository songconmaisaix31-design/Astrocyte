# W3 论文检索、进度与前端 DTO 交接

Branch `s1-sync-ui-1010`；Task `task_88f06e405600` / Dispatch `ctx_cc2d02cac1ac`，Run `run_8c1696bb815a`。固定工作树与分支，客户端 OpenCode（当前会话实际模型由运行环境配置，未覆盖模型参数，不声称继承 Codex 模型）。独占写域 `web/src/`（除 `api/`）、`web/public/`、`web/e2e/`、`docs/acceptance/S1-paper-board.md` 与本文件。

## 基线与边界

- 干净工作树普通合入 ROOT 精确基线 `8277667`，随后快进至 `origin/s1/attention-materials-20261009`（保留原 owner 历史，无 squash/rebase/reset/force，未越写域编辑）。
- 已读 HANDOFF、STATUS、QUESTIONS、SPEC §16/§17/§19 与本轮计划 `S1-final-paper-cli-memory-plan.md`。三项已答（论文搜索＋正文、允许批量读取、Agent 从获准 TASK/STATUS 推断进度）可落实；新两项（CLI 完整覆盖范围、记忆隔离/全局/仅规划）仍 PENDING，未代选、未建 UI。
- 未改契约/OpenAPI/生成客户端/迁移/锁/入口；检索与批量入库端点尚未发布，前端按真实 API 语义诚实呈现，不 fixture 冒充。

## 完成内容

Attention 新增「检索论文」入口：顶部流程条与「来源与整理」折叠面板均可进入。面板展示结果元数据与**原文可得状态**（原文可得/仅元数据/受限/可得性未知）、受限说明、已入库去重；勾选纯人工、可取消、批量入库由人显式提交且不全量；未知可得性不可勾选、提示等待重试。真实模式只走真实 API（未连接/501/失败等待与重试）；示例模式标注固定样本并禁止写入。

进度展示层为纯函数（`progressPresentation.ts`）：`percent=null` 呈现「未知」、绝不伪造；阶段/TASK-STATUS 来源引用/新鲜度并列展示，无依据呈现「未知」。CLI 原生操作沿用既有 `ManagedProjectsPanel`/`NativeProjectPanel`，权限判断不变，看板不据发现授权。

## DTO 交接（供 W0/W1/W2）

见 `docs/acceptance/S1-paper-board.md` 顶部「DTO 交接」：论文检索 `GET /papers/search`、批量入库 `POST /papers/import`、进度 `GET /local-projects/{id}/progress`、原生 `reconcile`、记忆（待答后）。本地形状 `paperSearch.ts`/`progressPresentation.ts` 已实现并被 UI 消费；集成时由 W0 用生成类型替换本地形状，UI 结构不变。

## 验证与首失败

| 命令/行为 | 结果 |
|---|---|
| `pnpm --dir web typecheck` | PASS/exit0 |
| `pnpm --dir web test` | 72 PASS（论文检索 7 + 进度 4，既有 61 保持） |
| 定向 `eslint`（新增/改动文件） | PASS/exit0 |
| `pnpm --dir web build` | PASS/exit0（96 模块） |
| `pnpm --dir web exec playwright test paper-search.spec.ts` | **8 PASS / 1.2 分钟**（4 例 × 1920/1280）：示例结果/可得状态/受限说明/选择/取消且 0 写请求；未知与已入库不可勾选；真实 501 呈现「检索未完成」且无样本冒充 |

首失败保留：首次 e2e 场 4 FAIL/4 PASS——「取消全部勾选」在 fixture 下被误置 disabled（纯 UI 动作不应受 fixture 门控），且测试用 `getByText('原文可得')` 命中多条导致严格模式冲突。分别改为 `disabled={!selected.length}` 与 `.first()`/更精确文案后 8 PASS；原 RED 不覆盖为通过。

## 未完成与真实限制

- 检索/批量入库/进度/记忆真实端点尚未由 W0/W1/W2 发布；端到端真实检索、批量入库、进度读取待契约落地后由 W0 集成（本轮 UI 已按真实语义诚实呈现未连接）。
- CLI 完整覆盖、记忆范围仍 PENDING，未代选；记忆与工具层仅规划，不建 UI 伪工具、不默认全局或原对话采集。
- 未运行个人 5173/8787、未安装插件、未读个人 Chrome、未合 main/远程 CI；三尺寸真实数据截图待端到端后由 root 复核（本轮 e2e 覆盖 1920/1280）。
- SOURCE 为 UI/测试精确提交，REPORT 为本文件与接受文档的随后提交；最终集成与主控验收由 ROOT 承担。
