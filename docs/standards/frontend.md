# 前端标准整改与合并产品验证

## 已交付阶段

分支 `standards-frontend`，整改 SOURCE `f534298`。执行客户端为本次 Orca 派发的 Codex；运行参数按派发使用既有 `--yolo`。实际模型身份等待主控本次运行观测确认，本轨没有发起其他 Agent 或模型调用。

前端移除 `vi.fn` / `vi.mock` API 模拟、`page.route` / `route.fulfill` 伪造响应和 fixture 事实展示。实际页面读取复用会话、生成客户端、生产 Go 服务与独立 SQLite。历史 `fixture=1` 仅作为查询参数回归输入，仍读取实际服务。纯函数和静态组件测试明确标记参数输入，保留 unknown、权限、完整路径和状态显示断言。

真实测试使用既有 `tests/s1/server.mjs`，服务、SQLite、原文对象及首失败仅位于本树 `.verification/`。资料输入使用本轨 Git 已跟踪协议、README、实际捕获的公共文档导出或公开 arXiv 版本；没有读取正式库、私人会话或凭据。失败场景使用实际权限拒绝、输入校验、请求取消及网络或服务断开。

## 阶段验证

- `npm run typecheck`：退出 0。
- `npm run lint`：退出 0。
- `npm run test`：16 文件、112 项通过，包含生产 HTTP 客户端与 SQLite 检查；日志 `.verification/vitest-dispatch-current-node.log`。
- `npm run build`：退出 0。
- `git diff --check`：退出 0。
- 完整默认 Playwright：本次阶段运行在收到合并产品范围变更后主动中断，日志 `.verification/playwright-dispatch-current.log`；结果保持 PARTIAL，不能视为通过。

前次纯函数、显示契约、typecheck、lint 首失败见 `.verification/first-failures.md`。前次真实浏览器日志及产物保留于 `.verification/playwright-first-*`、`playwright-materials-first-*`、`playwright-resumed-*`，包含未完成运行与失败；本次结果不覆盖首失败。

## 后续范围与真实限制

主控已授权合并产品前端由本轨独占，普通消费 Workbench `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b` 与 Relay 前端 `65ba685ea23e796c1810ff5ec13280dd0dda8f93`。当前尚未合并这两个输入；合并后在稳定 SOURCE 重新运行适用检查及完整默认 Playwright。

真实 Codex 原生进度、Relay controller / SQLite / HTTP / 每个 event_id / 三分钟链路等待 R/H 实际公开项目交付。本轨未执行真实 CLI / 模型，保持 NOT_RUN。现有需明确许可的原生、浏览器、媒体或模型用例保留原条件门禁；未获许可的路径不计为通过。Mission 创建接口仍受当前后端能力与授权限制。
