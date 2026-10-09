# 用户界面替换

2026-10-09。来源：`C:\Users\DW\Downloads\Astrocyte-preview.html`。基线：`649d4297d5bd34ebd849944379cb07bf909d7acb`；交付分支：`ui/preview-replacement-20261009`。

目标：把用户提供的三页界面迁入现有 React/TypeScript 前端，复用布局、样式、图标和交互；保持 Go/SQLite 与 OpenAPI 契约。仅做本次界面替换，不展开 S1–S6。

待用户确认：默认显示真实 API 数据还是 HTML 示例内容。问题未回答前不决定默认数据模式；可先迁移视觉组件。

单一开发轨（共享布局与页面写域重叠，不拆并行）：

| 负责人 | 写域 | 依赖 | 验收 | 停止点 |
|---|---|---|---|---|
| 前端 Worker，Orca Codex，模型以启动回执为准 | `web/`，不改 `web/src/api/schema.d.ts` 和依赖清单 | 上述 HTML、SPEC §16/§19、已冻结 API | 类型、lint、单测、浏览器；1920×1080/1280×720 视觉与键盘；API 错误和空态；示例标记；构建 | 默认数据方式待定；新增后端能力、契约或依赖需求先交接 |
| 主 Agent | 本计划、STATUS、HANDOFF；最终检查、普通合并、提交和 push | Worker 源码 | `pnpm check`、`pnpm build`、`pnpm test:e2e`；实际预览 | 核心检查失败则返修 |

执行：Worker 固定一个子 worktree 与分支，负责开发、测试和返修；主 Agent 最终收口。复用生成客户端与查询层，真实状态与示例状态分清；未实现的写操作继续显示禁用原因。开发完成后保存源码提交，普通合并到交付分支，最终 push，不覆盖历史。
