# Workbench v0.1 前端交付

本报告对应修订后的既有 `web` / 同一 `server` 方案，事实源为 [执行单](../../tasks/Workbench-v01.md) 与 [HTTP 契约](../../tasks/Workbench-api.md)。旧独立前端方案及其已推送提交保留为历史，未合入本目标。

分支：`workbench/v0.1-frontend`。SOURCE：`8ec40f7633236b80bf12e34aa6468633f96d4072`。基线 `a88fd6e`；普通 no-ff 消费契约 `1ef8f751a56a6f94dd26ba0474da25b65c5e68ac` 与后端 SOURCE `8635895e40eb022e92a86870861b4d09fd168147`。本次执行客户端为 Codex；具体模型由 Orca launcher 管理，本 owner 未独立核实。报告独立提交，SOURCE 不含本报告。

## 完成内容

- 原 Nav 新增“认知”，进入 `/cognition/today`；今日、焦点、来源、证据四子路由保留直接访问、重载、键盘切换与原 shell。
- 今日读取真实 A/M/C、nullable 日环比、来源活动、昨日 Sleep 和七日计数；今日无事件与缺少前日确证值如实显示。
- 焦点显示 API 提供的历史 safe 与实验期确证 A 多主题轨迹，不计算新 A、不生成补点，缺失长区间不连线；主题按 A 排序，可开完整证据抽屉，含日/kind 聚合与完整分页。
- 来源按存储 source 展示今日/近七日/近三十日及最近事件。`manual` 保留存储分组含义，不标成“人类行为”。
- 证据提供主题、kind、source 和上海自然日日期范围筛选，20/50/100 条分页，筛选重置页码，完整回溯。
- 事件表显式仅渲染 `event_id/kind/topic/tokens/scope/ts`；source 只用于查询/聚合组标签。认知读取经原 `createApiClient` session bootstrap，修正原公开 health 匹配误把认知 health 当公开端点的问题；无裸 fetch、新认证、依赖或独立项目。
- 原生成脚本实际目标为 `web/src/api/schema.d.ts`，已按精确发布 OpenAPI 生成并核实零漂移；主控 Handoff 确认此实际路径。

## 验证

环境：Windows / PowerShell；原 Playwright Chromium 1920×1080、1280×720，两视口均执行。所有日志、截图、trace 和 UI 数值 JSON 存于 `%LOCALAPPDATA%/astrocyte-workbench/v0.1/verification/frontend`，不入 Git。临时 server 使用独立 ephemeral 应用库，工作台 Reader 读取外部快照；未以部署中的独立 17321 服务代替同 server 装配验证。

| 命令 | 结果 |
| --- | --- |
| `pnpm check` | PASS / EXIT 0；包含 Go、架构/依赖、契约/examples、生成漂移、TS、lint、14 项 S1 acceptance、96 项 Vitest |
| `pnpm build` | PASS / EXIT 0；保留 Vite 大于 500 kB chunk 提示 |
| `pnpm --dir web typecheck` / `lint` / `test` | PASS；96/96 单测 |
| `node scripts/generate-client.mjs --check` / `git diff --check` | PASS |
| `pnpm --dir web test:e2e cognition.spec.ts navigation.spec.ts --workers=1` | PASS；最终认知契约与原 Nav 回归 116/116，两视口 |
| `ASTROCYTE_WORKBENCH=1 WORKBENCH_REAL_ACCEPTANCE=1 pnpm --dir web test:e2e cognition-live.spec.ts --workers=1` | PASS；真实四视图两视口 2/2 |
| 三份 spec 最终组合：`pnpm --dir web test:e2e cognition.spec.ts cognition-live.spec.ts navigation.spec.ts --workers=1`（同上两个环境变量） | PASS / EXIT 0；118/118，2.6 分钟，最终 SOURCE 同 server 与两视口 |

实际 PowerShell 用 `$env:ASTROCYTE_WORKBENCH='1'; $env:WORKBENCH_REAL_ACCEPTANCE='1'` 设置环境；Playwright 的 `--output` 与 `PLAYWRIGHT_HTML_OUTPUT_DIR` 指向上述外部取证目录，不修改原配置、端口或 timeout。

真实代际 `snapshot-hxfipcpr`，`snapshot_at=2026-10-13T17:24:01.253839Z`，上海当前日 `2026-10-14`。按 root 独立 SQL 导出冻结十项：三个 daily_count、最高主题 A/M/C、三个来源 days30、全事件 total。每视口各十项 `data-value` 与 SQL 数值精确一致，最大绝对误差 0；可见文本等于 SQL 数值以 zh-CN、最多四位小数格式化后的文本，未将舍入后的显示数声称为未舍入标量。两份 `ui-values-chromium-*.json` 留存比较结果，root 期望文件为外部 `root-sql-expected-2026-10-14.json`；root 另用独立 Python 逐项复核两个视口数值与文本均 PASS，取证在 `root-visible-ui-sql.json`，也已独立复核真实 Today / Focus 截图。

真实交互包含今日来源 GET、A 排序、主题抽屉第二页、六列事件、Evidence 第二页/筛选重置/自然日空集/清除筛选、所有四视图无水平整页溢出、全程 API 请求仅 GET、页面错误为零及开始/结束快照时间一致。四视图与抽屉截图已本机目视复核；不把 fixture 契约测试作为真实数据验收。

## 首次失败记录

- 初次定向 Vitest 为 12 PASS / 1 FAIL：测试 mock 对 bootstrap 与事件请求复用同一 Response，正文被提前消费。改为每次构造新 Response 后通过；原 `first-vitest-red.txt` 保留。
- 初次认知 + Nav 浏览器为 114 PASS / 2 FAIL：新增 spec 在抽屉最后一页预期 Shift+Tab 聚焦已禁用的“下一页”。增加末页“下一页”禁用断言，并验证焦点回到最后一个可用按钮“上一页”；保留六字段、页数、隐私、Escape 和 trigger focus restoration 断言，未修改共享组件或旧 spec。原 `contract-browser-first.log`、首轮 results/trace 留存，不改记 PASS。

## 剩余限制与未执行

- 本 owner 已验证前端同 server 真实读取；主控独立最终验收、唯一集成 owner 的正式静态 web 装配/部署及远端 CI 结果由其后续核实。本报告不宣称整个 Workbench 已最终验收，也未合 main。
- 管理员 Procmon 正式 DB/WAL/SHM 零写审计与两个真实日期 00:30 自动快照成功属于 root/backend 的时间与权限验收，不因前端只发 GET、手动刷新或本次浏览器 PASS 而解除。
- 当前快照 10-13/10-14 无事件，10-13 Sleep 未产出，日环比缺前日确证值；这些空值保持原样。默认 CI 跳过 opt-in live spec，真实验收必须明确配置外部快照与独立 SQL 期望文件。
- 未执行完整既有 Playwright 全套，仅执行本轨新增认知 spec、真实 spec 与适用的原 Nav 回归；未修改 V4、validation、引擎、正式数据或现有 E2E 配置。
