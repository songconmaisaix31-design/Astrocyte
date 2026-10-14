# 前端标准整改与合并产品验证

## 已交付阶段

分支 `standards-frontend`，当前 SOURCE `205d2d2ce91091e6add67d637cb65aae33946292`，已 push。执行客户端为本次 Orca 派发的 Codex；运行参数按派发使用既有 `--yolo`，本次 Worker 的实际模型身份等待主控运行观测确认。本轨没有另派 Agent。

前端移除 `vi.fn` / `vi.mock` API 模拟、`page.route` / `route.fulfill` 伪造响应和 fixture 事实展示。实际页面读取复用会话、生成客户端、生产 Go 服务与独立 SQLite。历史 `fixture=1` 仅作为查询参数回归输入，仍读取实际服务。纯函数和静态组件测试明确标记参数输入，保留 unknown、权限、完整路径和状态显示断言。

真实测试使用既有 `tests/s1/server.mjs`，本轨创建的服务、SQLite、原文对象及首失败仅位于本树 `.verification/`。资料输入使用本轨 Git 已跟踪协议、README、实际捕获的公共文档导出或公开 arXiv 版本。Q 交付的实际公开 Git 来源快照和 H 交付的安全派生事实仅按明确授权目录读取；没有读取正式库、私人会话、原生会话原件或凭据。失败场景使用实际权限拒绝、输入校验、请求取消及网络或服务断开。

已普通合并 Workbench `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`、Relay 前端 `65ba685ea23e796c1810ff5ec13280dd0dda8f93`、H `99a250d0b9069ffb32205420f63f82b372de5607`、R `ce3a519efa5317e642d7488a776b850eebca1093` 及 B `b470ecd736af261a10e4c4822bdcd3b02d0da454`。根文档和 `go.mod` 冲突由对应 Owner 处理。

Relay 写操作在传输失败、存在 accepted 结果及错误、HTTP 5xx（明确未实现的 501 除外）时保存未知交付锁。`sessionStorage` 仅保存 operation UUID，关闭、重开和重新加载继续禁止重复写入。Today 会话区分实际 archive 与 native notifications，保留 unknown、null tokens、分页和未完整确认状态。

## 阶段验证

- `npm run typecheck`、`npm run lint`、`npm run build`、生成 schema 检查和 `git diff --check`：已退出 0。构建保留现有大于 500 kB 的产物提示。
- `npm run test`：26 文件、147 项通过，日志 `.verification/vitest-combined-all-first.log`；包含实际生产 HTTP、SQLite、公开快照和 H 实际六字段事件查询，以及明确输入的纯函数测试。
- Playwright 实际无模型场景筛选运行（`--workers=1`）：两种窗口共 12 项通过，日志 `.verification/playwright-relay-owned-second.log`。覆盖权限拒绝、撤销、revision、网络断开后未知锁及嵌套目录授权。
- `npx playwright test e2e/cognition.spec.ts e2e/cognition-live.spec.ts --workers=1`：最近完整复验 13 项通过、1 项因实际 `ECONNRESET` 失败，日志 `.verification/playwright-cognition-latest-final.log`。失败发生于 1920 窗口独立 SQL 比较的会话 bootstrap，不能视为 14 项通过。
- Cognition 实际非空过滤输入修正后，两种窗口过滤用例 2 项通过，日志 `.verification/playwright-cognition-filter-fixed.log`。保留原先 12 通过 / 2 失败及独立诊断记录。
- `npx playwright test e2e/cognition-live.spec.ts --workers=1`：实际十个 SQL 值在两种窗口独立复验 2 项通过，日志 `.verification/playwright-cognition-sql-current.log`。没有代码修改、自动重试或断言放宽，前次连接重置保持 FAIL 记录；服务启动连接原因已提交 B 排查。
- `npx playwright test e2e/relay-session-readonly.spec.ts --workers=1`：两种窗口共 2 项通过，日志 `.verification/playwright-relay-session-readonly-first.log`。H 实际 18 个事件逐一匹配来源 ID，保留 5 个 unknown、17 个 null tokens、分页、真实 Origin 403 隐藏旧数据及恢复。
- 完整默认 Playwright：初次在范围变更后中断，日志 `.verification/playwright-dispatch-current.log`，保持 PARTIAL；合并产品后的完整默认验收尚未完成。

前次纯函数、显示契约、typecheck、lint 首失败见 `.verification/first-failures.md`。前次真实浏览器日志及产物保留于 `.verification/playwright-first-*`、`playwright-materials-first-*`、`playwright-resumed-*`。Works 首次实际 400 导致 12 项失败保留于 `.verification/playwright-relay-owned-first.log`；请求 key / revision 修复后的通过记录单独保存。原生注册三分钟失败记录和后续观测分别保存，本次结果不覆盖首失败。

## 后续范围与真实限制

Q 最新公开 Git 来源快照 `snapshot-892856014` 实际包含 1339 个可定位提交、两个主题各三条实际轨迹和一次实际 Sleep。实际最高 A 主题 MCP 有 20 条证据；独立分页断言使用实际 134 条证据的 Astrocyte / Attention Engine。最高 A 排序与大于 20 条证据的完整分页分别验证。最终验收使用 Sleep 后的当前来源，不使用此前来源替代。

获准固定公开任务通过生产 GUI 仅提交一次，任务 `0c6522bc-b232-4040-8407-745885644791`。实际输入被接受后返回 HTTP 503 / `provider_unavailable`，交付保持未知并禁止自动重试。H 独立检查确认实际 `gpt-6.1-sol`、一个 assistant、一个 task_complete、8827 tokens、观察到零次工具调用；旧任务的有效工具集合及 owned PID 未完整确认，工具关闭验收不能标记 PASS。

原始注册时间 `2026-10-14T13:23:34.0366105Z` 后三分钟窗口失败。两个窗口后续首次可见时间为 `13:33:25.152Z` / `13:33:25.153Z`，距离原注册约 591 秒，距离实际 dispatch 约 14.516 / 14.517 秒；该后续延迟只作实际观测。`.verification/l1-live/late-first-visible.json` 与 H 独立首 DOM / SQLite 派生记录保留，不能替代三分钟验收。

旧任务为 `working / terminal_exited / exited`。本轨通过 GUI 成功执行一次获准读取，operation `ede29910-9f7a-4f5e-97a9-9293f7b87a25`，实际 HTTP 200 / accepted，记录 `.verification/l1-live/existing-ui-safe.json`。旧 revision 2 的授权没有 reconcile，当前人工完成及释放控件禁用；R 正按主控新裁决实现 observe 下人工完成 / 释放，待准确 SOURCE、H 实际组装和本轨 GUI 复验。

| 验收范围 | 当前状态 |
| --- | --- |
| 实际 HTTP 客户端、公开 cognition 查询、六字段原生事件、纯函数输入 | Vitest PASS 147 项 |
| Works 权限、撤销、revision、嵌套目录、未知锁关闭 / 重开 / reload | 实际 Playwright PASS 12 项 |
| 最新 cognition 四视图、排序、实际轨迹、过滤、日期、分页、十个 SQL 值 | 最近完整 13 PASS / 1 FAIL，继续排查连接重置 |
| Today 实际事件 ID、null / unknown、两页及实际 403 | 实际 Playwright PASS 2 项 |
| 注册后原生三分钟首次可见 | FAIL，首窗口不重置 |
| 旧任务人工完成 / 释放、完整工具关闭证据 | BLOCKED / 未完整确认 |
| 实际 EOF、结构化 EOF、generic 409、publication error、第二 CLI | 所需真实条件未完整交付，未 PASS |
| 资料版本 / 去重 / 显式刷新、人工内容与问题、later / reject / revise、四维 unknown、A/B/C 与独立模型参数 | 完整默认用例保留，合并后的最终完整运行未完成 |
| 人工进度与一次真实生产 infer 的缓存页面验证 | B 已报告实际调用，等待获准缓存输入及 GUI 复验，未 PASS |
| 实际带时间段的视频来源 | 当前公开来源零个时间段，正向时间断言缺少真实输入，未 PASS |

原生场景保留准确获准任务 ID 输入要求；不添加 skip，不删除成功覆盖，不伪造来源、错误或执行结果。B 新公开输入缓存及 Reader、R 新执行 / 权限源码、H 配套组装尚待准确交付及消费。Mission 创建接口仍受当前后端能力与授权限制。
