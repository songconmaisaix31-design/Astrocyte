# S1 同步与本地 Agent 前端 W3

2026-10-10；基线 `20437708e43201e352d6c6926902e1363fd2ad3e`；分支 `s1-sync-ui-1010`。本机 Orca 原生 Codex，主控通过 worker-read 确认实际模型 `gpt-6.1-sol`。写域为 web/src（排除 api 与 pages/swarm）、web/e2e、本报告与 docs/acceptance/S1-sync.md。

## 阶段一

已读取 AGENTS、HANDOFF、STATUS、S1-sync-local-agent-plan、SPEC §6/§19、QUESTIONS，并用 view_image 检查指定 JPG。保留既有三页 shell/布局；只在 Attention 加入账号待接入区，把 Workspace 项目区改为参考图结构的暖白概览、平台/绑定状态/分组筛选与网格卡。卡显示服务注册目录，详情沿用原抽屉。统计基于已加载项目/会话，平台仅由匹配 project_id 的会话聚合，不推导已安装或正在运行；API 未提供分组、文件数、交接内容、活动时间，保持未提供，最近活动排序禁用。

账号 B站/抖音绑定方式待确定，其余平台明确待接入，无伪连接按钮。没有新增连接、扫描、执行或摘要请求；未接入数据不得等同零更新。W0 的客户端/契约发布后由 W3 消费，不修改 API 或锁。

已执行 `pnpm --dir web typecheck`、`pnpm --dir web lint`、`pnpm --dir web test`（43/43）、`pnpm --dir web build`、`git diff --check`，PASS。新增单测验证项目关联隔离、筛选组合、活动排序未知/无效值置后且不改输入。浏览器由主控授予单槽，准备执行现有 ephemeral helper 上的 targeted spec/navigation/preview-design；尚未执行，不能声明浏览器通过。

## 待定用户旅程

这是讨论草案，不冻结接口：确认绑定来源 → 获取更新列表 → 展示来源/状态与 Agent 反馈 → 人工选择 → 复用现有导入/作业反馈 → 回查资料版本。反馈前是否预取正文仍 pending；账号登录或公开绑定、真实目标账号、本地操作范围与读取根待用户。遇到挑战页/需登录/不可用应显示具体原因，保留先前清单，不把异常转换为空更新。项目导入和扫描入口待指定目录授权；安装/配置/可启动/原生能力分别显示，观察会话不等于控制。

完整账号集成、真实同步、AT01–04、新模型整理、授权 Agent 读取、两尺寸最终真实验收尚未完成。旧 UNKNOWN 未重发；未导入个人库，未读凭据或扫描磁盘。

## 阶段一首轮浏览器（原始失败保留）

SOURCE `04a67c54a6ccdbba55dc08e5ec44c07a739d79e7` 已 push。总控单槽安排 `pnpm --dir web exec playwright test s1-sync.spec.ts navigation.spec.ts preview-design.spec.ts --workers=1`：120 PASS /2 FAIL，2.3分钟，exit1。两失败均为新 empty-state 用例中对原生 `<option>` 使用 `toBeDisabled()`，Playwright 展示 `<option disabled>` 与可访问快照 disabled 却返回 enabled；改核对 `toHaveJSProperty('disabled', true)`，应用代码不改。占位账号文本/零按钮测试按主控复核移除，后续应测试实际绑定/选择行为。首次不是绿，不覆盖。

首轮 artifacts：`web/test-results/s1-sync-local-overview-kee-02fb9-acts-explicit-at-both-sizes-chromium-{1920,1280}/` 的 error-context、test-failed-1.png、trace.zip。两尺寸 `web/test-results/workspace-{1920x1080,1280x720}.png` 与卡片截图已目视检查：保留 shell，宽屏三列/1280两列、筛选可换行、目录可读、无横溢。示例图是布局证据，不是本地项目接入证据。

临时目录 `%LOCALAPPDATA%/Temp/astrocyte-dev-qQGXqS`，个人库未使用；API86296/Vite88652，父72516，创建时间03:04:56。命令结束后所属三进程均不在，15173/18787无监听，Test-Path仍True，目录保留不自行清理；已向主控释放单槽。仅失败两尺寸用例的修复验证待主控重新安排；不重跑已过的导航回归。

## 阶段二：已发布清单客户端

普通合并 W0 已推送 `bb912efd6f28f9f3909d87ea9fc1f4acb04c9098`（含 W2 DTO），没有手改 generated API。新增 useLocalAgents 与 LocalAgentsPanel，分别显示安装/配置/可启动观察和八项原生能力的状态、原因、核实时间。计数仅来自该接口已加载记录，版本缺失保持未提供，fixture 模式不查询真实清单。清单 GET 只展示已保存探测结果；没有控制按钮或命令请求。

阶段二 `pnpm --dir web typecheck`、lint、test（44/44，含 W0 新客户端单测）、build PASS。新增浏览器用例比较真实 GET local-agents 数据与页面逐项原因及八能力、刷新只发 GET；等待 W0 组装/service 源和主控单槽，尚未运行。不能从 DTO 发布或 UI 编译成功推断实际已接入，更不推断启动/接续/停止能力。

W0 真实组装源 `86502c63765c19a086770d8646030791ca6c39da` 已发布并普通合并（本地合并 `6aede6d72ed2fa6aedf0f3a3e1b2ca173e46b288`）；原生版本/help 启动探测总10秒、每次5秒，GET只读缓存。主控复核后将已知原因译为中文，未知原因保留诊断详情，保持未知与不支持区别。新增实际清单 E2E 和原失败修复准备一起在单槽运行，首次 artifacts 不覆盖。
