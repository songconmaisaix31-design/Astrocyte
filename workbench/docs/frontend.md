# Frontend v0.1 — 旧规格封存，未验收

状态：`scope_replaced_by_user`。当前 Dispatch `ctx_446537d9f0c1` 的独立前端工程任务未完成，不应合入新的现有 web 扩展方案。

## 规格替换

原事实源为 `tasks/workbench-v01-plan.md`、`tasks/workbench-api.md`，基线 `3e0eb3d`，独占写域 `workbench/frontend/**` 与本报告。协调者于 2026-10-13 16:12 上海时间通过 `msg_9cb8a56583fa` 转达用户替换规格：从 `attention-engine/v1-20261011` CI 绿基线 `100f00cb` 开始，扩展现有 `web`，增加认知一级导航与四视图子路由，复用 token/component/auth/build，并将新认知 spec 放入原 Playwright；不新建前端工程。

收到消息时已进入中的工具批次完成后，停止原开发，普通 WIP commit/push 封存。未 reset、rebase、删除任何贡献；未修改 backend、scripts、root 任务、validation、正式数据库、原 web 或引擎。

## 已保留内容

独立 React 19.3.0 / Vite 8.3.4 / TypeScript 6.0.3 工程骨架，以及今日、焦点、来源、证据四视图的未验证代码。包含六字段事件表、API 请求、分页、筛选、原生 SVG 实测散点与证据抽屉。开发端口原拟 `127.0.0.1:5178`，构建目录原拟 `workbench/frontend/dist`；初次 Handoff 的 8787 已由后端通知修正为 17321。以上均属于已替换方案，不能视为运行事实或新方案的接缝要求。

分支：`songconmaisaix31-design/workbench-v01-frontend`。

WIP SOURCE：`f9477f46f96b78388cc1a46a8860d981b66ec9ab`，已 push。REPORT 为独立后续 commit，其 SHA 由完成消息给出。

## 验证记录

| 命令 / 验收 | 结果 |
|---|---|
| `git diff --cached --check` | PASS，仅 WIP 补丁空白检查 |
| `git push -u origin songconmaisaix31-design/workbench-v01-frontend` | PASS，WIP SOURCE 已推送 |
| `npm install --no-audit --no-fund`（workbench/frontend） | PARTIAL / INTERRUPTED，规格替换后主动中止，exit 1，无安装成功结论 |
| `npm run typecheck` | NOT_RUN，规格替换后停止 |
| `npm run build` | NOT_RUN，规格替换后停止 |
| 四路由真实 browser / 过滤 / 分页 / drawer | NOT_RUN |
| 十个真实 UI 显示值与 root SQL 比对 | NOT_RUN |

首次安装中止记录保留；不存在任何真实浏览器、mock 浏览器或构建通过证据。代码与声明中的 `test:browser` 入口仍为 WIP，不代表测试实现完成。未读取或发布私有快照、数据库事件载荷、浏览器证据、日志或构建产物。

## 剩余操作

原任务按用户替换规格停止，`worker_done` 必须标记 `failed`。由 root 核对新基线后，在同 worktree、同 terminal 复用原 owner 派发新 Task；此报告不批准复用旧工程或替代新方案验收。新方案 typecheck/build、原 Playwright 中四视图真实验证与十值 SQL 核对尚待执行。
