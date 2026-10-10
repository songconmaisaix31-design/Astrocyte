# S1 W3 项目总览与人类流程验收

本报告只覆盖本轮已授权、已实现的项目看板与现有入口。论文搜索、当前页提取、插件权限与进度推断仍为 **ASKED PENDING**；没有据此选择产品方案、冻结接口或宣称完整 S1 交付。

## 分支与源码

- 分支 `s1-sync-ui-1010`，最终 UI/测试 SOURCE `2aa57e264aa56a7d7f271c5e087bc540f3f862ab`，已 push。REPORT 为提交本文件及 W3 任务记录的随后报告提交，精确 SHA 通过 Orca 交接。
- 精确根基线 `087127d30fc505f8fe25c52142a174f5842e2bef` 普通合并；W0 契约/HTTP/helper 与 W2 聚合/存储按发布 SHA 普通消费，最后已消费 W2 `be3cbfdbf536053f196ae5c48afad5fbf48138c2`、W0 组装 `967ed09eb50c988fbd9e04c99f48e91a4c023f47`。没有 rebase/squash/reset/force push 或越域业务编辑。
- 固定开发客户端为本 Orca Codex 会话，无额外 Agent。未自行切换模型；会话未暴露可核实的精确模型标识，不推测记录。

## 完成内容

工作区实际入口改为真实项目总览。后端 canonical ID 决定卡片身份，同一项目多客户端/多目录不增加重复卡。项目、来源文件夹、已记录客户端、已知活动时间与归档分别统计；支持平铺、平台组合、分组、时间线，按名称/目录/备注、自由人工意愿、客户端、组、来源、活动状态与归档筛选。页顶与看板搜索取交集，“清除筛选”同时清除两处输入，搜索文案按当前页实际数据范围显示。

项目详情优先呈现备注/下一步、复盘、分组、自由人工意愿与可恢复归档。SQLite 人工记录使用真实 API 与 revision；保存后采用服务器版本，外部更新保留未保存输入并明确提示。意愿不转成完成率，来源活动不转成人类意愿，完成度保持未知。根目录、来源、创建时间、真实记录活动、贡献客户端和仓库观察可回查；权限/CLI/原生/源码仓库工作流仍在“项目接入、权限与原生操作”入口，未增加自动控制或授权。

默认无 fixture 回填；加载、空、错误、旧缓存与 partial 来源明确显示。布局复用现有 Go/React/SQLite/OpenAPI、查询层、抽屉与官方图标，实际查看了用户参考图片；手机详情长目录换行有界，键盘 Enter 打开、Escape 关闭并恢复卡片焦点。

## 验证与范围

| 验证 | 精确范围与结果 |
|---|---|
| `pnpm --dir web typecheck` | 首次展示层 PASS；最终布局 `7373922` PASS；`acbfc43` build 含 tsc PASS |
| `pnpm --dir web test` | `c07fb11` 前端 61 PASS，其中项目展示模型 3 例验证去重、活动/意愿分离与筛选非破坏性；后续仅布局及导航测试变更，没有重复不变单测 |
| `pnpm --dir web build` | `c07fb11` 与最终生产布局 `acbfc43` PASS；`acbfc43..2aa57e2 -- web/src web/public` 无差异 |
| 定向 `eslint` | 已修改 UI、呈现模型、表单与浏览器测试 PASS；最后测试修正是标题选择器的精确修复 |
| 实际浏览器 `--config=e2e/project-board.config.ts` | **固定 `c07fb112ffd2497d4a913540f6836793e60a13db` 1 PASS，47.4秒（测试46.2秒）**。真实本机来源、真实 API、SQLite 保存/复盘/分组/意愿/归档/恢复、刷新后保留、同库新进程冷重启、四视图无重复、键盘焦点与 1920/1280/390 无横溢；没有模型/媒体/克隆或原生运行 |
| 最终布局实际预览 | **固定 `acbfc432aceec14fcafe3ad37d3845b994f450ef` 1 PASS，24.2秒**，1920×1080、1280×720、390×844；实际项目 API、非 mock，仅 preview GET，未重放人工字段/模型任务。首卡顶部桌面604.7px、手机769.6px；桌面明确断言首卡顶部至少80px处于初始视口。来源详情真实 partial，无虚构完成率 |
| 定向导航/错误传播 | **固定 `2aa57e2` 4 PASS，22.2秒**，1920/1280。工作区项目总览→接入设置→Agent会话；错误注入 `missions` 后任务、工作项、产物引用各自在实际面板/标签页显示错误。此项是受控接口失败回归，不冒充实际服务故障或 task_live |

实际浏览器执行设置：`ASTROCYTE_TEST_PROJECT_BOARD=1`，`ASTROCYTE_BOARD_REUSE_OWNED_TEMP` 与 `ASTROCYTE_BOARD_REUSE_APPROVED_ROOT` 均为自有 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-7wY4sg`；API18787/Vite15173，helper 只启动/停止自有进程。预览另设 `ASTROCYTE_BOARD_PREVIEW_ONLY=1`；导航重验使用 `ASTROCYTE_E2E_API_PORT=18787`、`ASTROCYTE_E2E_WEB_PORT=15173` 和 `playwright test navigation.spec.ts --grep 'navigates to 共同工作区 page|swarm propagates mission error'`。个人 Chrome、5173/8787、后台 daemon、全局配置与 Provider 凭据均未操作。

固定实际场次看到 **111来源目录 / 71 canonical 项目 / 2已记录客户端 / 43来源读取限制，status=partial**。这不是全盘覆盖、全客户端覆盖或真实项目完成度。实际测试还断言 `attention_jobs`、`local_agent_grants` 和受控项目权限记录在本次人工字段操作前后不变。

### 保留原失败与原件

1. 接入初次 TS RED：Windows 下 `ProjectBoardView.tsx` 与 `projectBoardView.ts` 大小写邻近冲突；后改为 `projectBoardPresentation.ts`。W0 helper 首缺 apiPort/webPort 声明造成另一次 TS RED，交唯一 owner 修正后 PASS；不改写原 RED。
2. `s1-board-live-20261010` 首实际场次 **FAIL / 4.5分钟**：可访问名称混入提示/选项，HMR 修正后继续保存、刷新、冷重启，最终390长目录详情按钮横溢。该场不能算固定源码验收；原 SQLite、日志、截图、`error-context.md`、`trace.zip` 均保留。独立固定 `c07fb11` 的 PASS 作为后续证据。
3. W0 `967ed09` 默认回归 **158 PASS / 4 FAIL / 22 SKIP / 7.7分钟**（W0交接，不冒充 W3 全套执行）：旧导航标题和折叠任务错误面板断言。W3 首修后定向 **2 PASS / 2 FAIL / 13.5秒**，暴露原数3遗漏产物专用标签；改按三处语义检查后的第二场 **2 PASS / 2 FAIL / 23.2秒**，产物标题选择器漏“引用”字样。原失败及 traces 均保留；最终4PASS不覆盖历史、不称整套重新通过。

本工作树本地原件：

- `web/test-results/s1-board-live-20261010/`：首失败 trace/截图/上下文；日志 `%TEMP%/astrocyte-w3-board-live-20261010.log`。
- `web/test-results/s1-board-final-live-20261010/s1-project-board-actual-pr-f19e0-resh-and-owned-cold-restart-chromium-owned-board/`：固定c07实际 `actual-snapshot.json`、`actual-after-restart.json`、三尺寸看板及详情 PNG；日志 `%TEMP%/astrocyte-w3-board-final-live-20261010.log`。
- `web/test-results/s1-board-final-compact-20261010/s1-project-board-actual-pr-f19e0-resh-and-owned-cold-restart-chromium-owned-board/`：最终布局实际来源快照、1920/1280/390截图；日志 `%TEMP%/astrocyte-w3-board-final-compact-20261010.log`。
- `web/test-results/s1-board-navigation-{retest,final,verified}-20261010/` 与 `%TEMP%/astrocyte-w3-board-navigation-{retest,final,verified}-20261010.log`：两次 RED 与最终定向4PASS原件分开。

## 未完成与交接

- 根发现 W2 汇总尚未纳入已观察 Git 提交时间，native-only 项目活动被错误标未知；当前上述快照7/71已知活动是修复前观察。原 owner 正在单独修复，本 W3 尚未宣布最终结算，后续真实汇总观察由精确 W2 发布及根验收补记；不重扫盘或重放付费任务证明它。
- 论文搜索→来源→选提取/入库→继续整理、独立插件真实加载/入库：依用户对搜索/当前页/插件权限的待答决定，**NOT_RUN/PENDING**。W3未自行启用，不把 W1 DOM提取核心视作可安装插件。
- 进度推断依据仍未决定，完成度未知；自由人类意愿仅为已批准的人工记录。
- 冷重启为停止自有 API 进程并使用同库启动新进程，未执行操作系统重启。最终小布局变更没有重跑不变真实人工字段流程；最终产物对应固定版本检查分别列明。
- 未合 main、未运行远程 CI、未更新个人预览/安装插件；根主控负责最终普通集成、全局检查与人类视觉接受。本报告不是完整S1或所有待答目标的接受证明。
