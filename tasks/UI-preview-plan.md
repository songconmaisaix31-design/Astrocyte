# 用户界面替换

2026-10-09。来源：`C:\Users\DW\Downloads\Astrocyte-preview.html`。基线：`649d4297d5bd34ebd849944379cb07bf909d7acb`；交付分支：`ui/preview-replacement-20261009`。

目标：把用户提供的三页界面迁入现有 React/TypeScript 前端，复用布局、样式、图标和交互；保持 Go/SQLite 与 OpenAPI 契约。仅做本次界面替换，不展开 S1–S6。

待用户确认：是否改为默认显示 HTML 示例内容。问题未回答前不做这一新决定；界面替换沿用 S0 既有数据约定（默认真实 API，显式 `?fixture=1` 示例），不阻塞模式无关的工作。

单一开发轨（共享布局与页面写域重叠，不拆并行）：

| 负责人 | 写域 | 依赖 | 验收 | 停止点 |
|---|---|---|---|---|
| 前端 Worker，Orca Codex，模型以启动回执为准 | `web/`，不改 `web/src/api/schema.d.ts` 和依赖清单 | 上述 HTML、SPEC §16/§19、已冻结 API | 类型、lint、单测、浏览器；1920×1080/1280×720 视觉与键盘；API 错误和空态；示例标记；构建 | 默认数据方式待定；新增后端能力、契约或依赖需求先交接 |
| 主 Agent | 本计划、STATUS、HANDOFF；最终检查、普通合并、提交和 push | Worker 源码 | `pnpm check`、`pnpm build`、`pnpm test:e2e`；实际预览 | 核心检查失败则返修 |

执行：单轨 Worker 固定当前 worktree 与交付分支，独占 `web/`；主 Agent 仅写文档，负责最终收口与源码提交，无需额外合并。复用生成客户端与查询层，真实状态与示例状态分清；未实现的写操作继续显示禁用原因。最终 push，不覆盖历史。

实际客户端/模型：Orca Codex / `gpt-6.1-sol`。1920×1080 与 1280×720 的源 HTML 三页截图已检查；新开发预览终端为 `term_af3ba99d-1301-41be-be0c-860189357aaa`，网页与代理 health 已就绪。

结果：界面替换 PASS。源码 `5c1517d652afef8273cf14dab5c28021db648cfa` 已由主 Agent 提交与 push；`pnpm check`（37/37）、`pnpm build`、`pnpm test:e2e`（112/112，51.0 秒）及两尺寸三页真实/示例预览均由主 Agent 独立复验通过。开发失败与限制见 [前端交接](../web/UI-preview-report.md)，本轮 CI 状态及后续入口见 [STATUS](../STATUS.md)；未合入 main。最终收口只改文档，保留单轨源码与历史。
