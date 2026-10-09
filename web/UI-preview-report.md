# UI-preview-web 开发交接

2026-10-09；Orca dispatched Worker / Codex，启动回执模型 `gpt-6.1-sol`。分支 `ui/preview-replacement-20261009`；本轨源码未 commit/push，按用户要求由主 Agent 最终提交。当前文档 HEAD `4fc6243059fe7243cb6d67f805683610c6206475`，不是本轨源码提交。

## 完成内容与修改路径

- `src/App.tsx`、`components/Nav.tsx`、`PageFrame.tsx`、`activeTab.ts`：迁入白绿布局、品牌 SVG、顶部圆角搜索、侧栏项目、三页 tabs、右栏、页脚；搜索仅筛选当前页已加载数据。支持 `/`、1/2/3、tab 左右/Home/End、历史返回与显式示例模式保留。
- `styles/preview.css` 复用用户 HTML 的 presentation CSS；`styles/variables.css` 集中匹配配色，`public/favicon.svg` 匹配品牌。没有迁入 HTML 或 React16 runtime，没有 iframe、原 HTML 覆盖、ts-nocheck、关闭 lint 或新增依赖。
- `pages/attention/*`、`components/MaterialCover.tsx`：纸张/音波/笔记封面、资料类型筛选、机会依据和来源详情；保留原始字段、人工/Agent 使用计数及未知状态。
- `pages/workspace/*`：首项目 hero 与品牌网络插画、其他项目紧凑卡、三列会话、提案批准 tab；保留路径、绑定、原生能力、上下文版本、停止条件与预算，费用/调用上限未知明确显示。
- `pages/swarm/*`：任务、工作项、产物引用卡；`components/DesignExamples.tsx` 在显式示例模式呈现原 HTML 的研究路线、固定动态、拓扑/列表切换及可检查来源的示例抽屉。
- `components/DetailPanel*`、`FixtureBanner.tsx`、`SectionCard*`、`DesignIcons.tsx`：柔和示例提示、遮罩抽屉、Escape/点击遮罩关闭、焦点约束和恢复、禁用操作与原因。fixture/API 切换通过 page key 重新挂载，不能残留示例选中详情。
- `utils/search.ts`、`e2e/navigation.spec.ts`、`e2e/preview-design.spec.ts`：保留 API 空态/error/stale/retry、路由、详情和示例边界覆盖；按新 tab 更新断言，新增搜索/筛选/历史/键盘/禁用操作/模式切换/横向溢出验证。

未修改生成 schema、依赖清单/锁、查询层、fixture DTO、后端或规格。主控修改的计划文件不属于本轨修改。

## 实际验证

| 命令 | 最终结果 |
|---|---|
| `pnpm --dir web typecheck` | PASS |
| `pnpm --dir web lint` | PASS，0 errors / 0 warnings |
| `pnpm --dir web test` | PASS，37/37 |
| `pnpm test:e2e` | PASS，112/112；1920×1080 和 1280×720；49.1 秒 |
| `pnpm build` | PASS，Go 可执行文件及 React/Vite 产物 |
| `git diff --check` | PASS；仅 Windows CRLF 提示 |

开发失败保留为独立过程事实：首轮 E2E 86/88（旧标题断言同时匹配右栏“会话”）；第二轮 108/110（新增预算测试选中已知预算样本）；第三轮 110/112（正确空预算样本仍只显示破折号）。已分别修正断言/样本及未知预算展示，最终另行获得 112/112，不将前轮结果改记成功。开发中的 TS 字段名/可空字段错误及两条 Fast Refresh lint warning 均在最终验证前修正。

浏览器使用独立端口 15173/18787、临时数据库；最终日志 `test-results/ui-preview-e2e.log`，报告 `playwright-report/index.html`。截图 `test-results/{attention,workspace,swarm,detail}-{1920x1080,1280x720}.png` 与 `swarm-topology-{1920x1080,1280x720}.png` 已实际查看；截图禁用动画并回到页顶，正文、封面、右栏、长标题、抽屉及拓扑示例可辨认，两尺寸无横向溢出。上述生成目录由现有 gitignore 忽略，重跑 Playwright 会更新它们。

## 真实限制与待主控操作

- 延续已交付行为：真实 API 为默认，`fixture=1` 显式示例；默认示例的新决定仍未决定。
- 当前 API 不提供收藏状态、研究路线或事件时间线。真实模式明确未接入，拓扑按钮禁用；原 HTML 的固定研究路线/动态/图只在显式示例模式出现并标记来源。
- 网络 hero 是品牌插画，不能用来推断执行拓扑；示例客户端名称不代表模型或活跃进程。
- S1–S6 写入、批准、认领、接续、交接、暂停/取消/diff/采用均未实现，按钮禁用；产物只有引用 ID，不声明实际内容或已采用。
- 主 Agent 尚需独立最终检查、实际预览验收、提交和 push。本轨没有执行 commit/push 或发布。
