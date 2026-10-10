# W3 论文流程与项目看板

## 侧栏全局添加资料返修（2026-10-10）

Dispatch `ctx_b423c20821bc` / task `task_bf953532256e`；沿原 `s1-sync-ui-1010` 工作树/分支，实际 Orca Codex / 默认 GPT-6.1-Sol 继承，无模型 override、无新增 Agent。独占编辑仅 `web/src/components/Nav.tsx`、`web/e2e/navigation.spec.ts`、本文件。

- 普通消费 ROOT 精确基线 `4b733a6e4484c8b38ff6892c890e45cc6a33623c`（含 W0 最终 `42e1941`），没有冲突；非本写域变化仅来自 owner 历史的普通合并。
- SOURCE `3b9b491a611d43cef4cef5819231b166763ee4fb` 已 commit/push。删除 `canImport` 的 Attention 页面位置条件；保留 `!fixture`、服务 `capabilities.imports` 与 `!foundation.stale`。复用原 `/attention?import=1` 导航、表单及关闭行为，没有增加权限或 API。
- 当前 `navigation.spec.ts` 没有可替换的旧添加资料断言，新增 8 个定向用例：workspace/swarm 各自实际能力下打开/关闭既有表单、删除旧 tab 参数且无写请求；fixture、能力关闭和首次能力读取失败仍禁用。能力关闭/未知场景明确为浏览器 GET 拦截。
- `pnpm --dir web typecheck`、`pnpm --dir web exec eslint src/components/Nav.tsx e2e/navigation.spec.ts`、`pnpm build`（Go + TS + Vite，92 模块）、`git diff --check` 均 PASS / exit0；未跑全套检查或默认 E2E 服务。
- ROOT 已普通合入 SOURCE 并确认 5173 HMR 读取新条件。随后 PowerShell stdin `node --input-type=module` 调用既有 `@playwright/test` 的 Chromium/expect，在原个人 5173/8787 执行上述定向浏览器操作：1920×1080 与 1280×720、workspace/swarm、实际/fixture/能力关闭/能力未知共 **16 PASS / exit0**。实际正向 4 项取得 foundation 200 且 imports=true，跨页打开既有“添加资料”dialog、可见导入方式、Escape 关闭及移除 import 参数全部通过；fixture 4 项与 GET 拦截负向能力 8 项单列，不能当成后端实际能力证据。
- 该场首运行即通过，日志 `web/test-results/s1-import-navigation-first-20261010.log`、4 张 `web/test-results/import-navigation-{workspace,swarm}-{1920,1280}-20261010.png` 保留；写请求 0、外网请求 0、新服务 0。没有 POST、保存、发现刷新、正文获取、扫描、媒体或模型，也没有默认 Playwright suite / webServer 启动或全套通过声明。stale 条件在源码保留，本次未另造 foundation 刷新失败场景。
- REPORT 为本文件随后独立报告提交，精确 SHA 通过本 Dispatch 的最终 worker_done 交接；SOURCE 与报告区分，原看板/模型/活动等历史验证不重跑或改记。本次最终集成及主控验收由 ROOT 承担，未合 main、未运行远程 CI。

本次只修复全局入口，不扩大论文/插件/进度等仍待决定的范围；真实提交与导入链路不因打开表单而被宣称通过。

本轮 Dispatch `ctx_d742848ca3e0` / task `task_11fedd3807d6`。固定工作树与分支 `s1-sync-ui-1010`，仅写 `web/src/`（排除 `api/`）、`web/public/`、`web/e2e/`、本文件与 `docs/acceptance/S1-paper-board.md`。实际客户端为当前 Orca Codex 会话，不另开 Agent。

## 范围与顺序

1. 已读 HANDOFF、STATUS、QUESTIONS、SPEC/TASKS 相关项目章节及本轮计划；干净工作树普通合并精确基线 `087127d30fc505f8fe25c52142a174f5842e2bef`。已实际查看用户提供的 `44fd1cdada350aac2c5bdef56f394fe7.jpg`，复用现有 React、查询层、抽屉、表单和官方图标。
2. 先交付独立展示层：项目总览、项目/文件夹/客户端分开统计、平铺/平台组合/分组/时间线，筛选、来源和实际活动。多客户端同一后端项目身份只一张卡；创建客户端不是贡献客户端。人工备注、复盘、分组、继续意愿、可恢复归档与未知完成度区分。当前没有冻结新的 HTTP 接口。
3. 消费 W0 发布的精确生成契约和 W2 真实聚合接口后接入主入口及持久化写入；现有项目权限、Agent 原生和仓库工作流保持可达。论文搜索/当前页/插件授权/进度选择仍为 ASKED PENDING；未获根决定前不接相关新写入。
4. 在协调确认的独占 18787/15173 槽使用隔离数据与浏览器，验证真实 API、人工字段保存、刷新与服务冷重启、键盘和 1280/1920/390 页面无横溢。插件验收依 W1 交接；不触碰个人 Chrome/全局配置。至多一次新容量联合调用须先完成 W1/W2 组装和根所有权交接，旧成功与 UNKNOWN 不重放。

## 当前检查

- 可复用看板展示层：定向 Vitest 3 PASS，验证多客户端卡不重复、实际活动不推导意愿、未知时间排序、来源筛选、归档可恢复且不修改输入。
- 初次展示层 TypeScript 与 lint PASS。第一次接入真实调用方后 TypeScript RED：Windows 大小写相近的 `ProjectBoardView.tsx` / `projectBoardView.ts` 导致模块解析冲突；改用明确不同文件名 `projectBoardPresentation.ts`，后续定向重验单列，保留此次 RED。
- 已普通消费 W0 `f9cdec80265f55efbfbdf6a9380ed684e46eee2d` 生成 API。W2明确 human intent 为自由文本，界面改为自由填写、三种可选中文建议，按实际文本筛选；没有固化完成阶段或进度推断。
- 真实浏览器槽、接口发布时间与产品问题回复已向协调者交接；等待时继续独立授权工作。
- W0端口 helper 的声明初次缺 `apiPort/webPort`，TypeScript RED 已退唯一 owner，`bbe512699d3e54d04ecd3280d8490f604dc1cd9c` 修正后定向 typecheck PASS。W0/W2 created_at 与活动分离已消费；后端最新原生元数据来源 `c177dc92a35c5e15c5c3fd8e7507dfbfb045e4d6` 与对应生成契约 `c6639f8dd61458da793f333da2eeedb35e20cf4b` 均普通合入。
- 首实际布局预览 PASS（21.1秒）：原隔离库 `C:/Users/DW/AppData/Local/Temp/astrocyte-s1-7wY4sg`，API18787/Vite15173，54目录/14项目为中间观察，不当作最终覆盖。主控已看1920/1280/390截图，要求更明显项目卡和简化手机筛选；返修后的viewport预览 PASS（1.3分钟）。仅安全GET、无人工字段写入/模型/媒体/克隆。
- 当前主入口支持自由人工意愿、真实来源统计、四视图、主筛选与明确更多筛选；卡片已增加独立表面和打开入口，权限/原生操作为次级可达。61项前端单测及前端生产构建此前 PASS；最终实际字段写入、刷新、冷重启与宽度/键盘正在验证。
- 首实际字段场次 `web/test-results/s1-board-live-20261010` **FAIL**（4.5分钟）。精确名称查询发现自定义意愿/筛选 label 混入提示和选项文字；当场 HMR 修正继续完成了保存/归档/刷新/真实进程冷重启/恢复，然后390详情断言发现长目录按钮横溢。该场非冻结源码验收，完整原trace、截图、失败与同一SQLite保留，不覆盖为PASS。修复为明确 aria-label/description 和目录按钮正常换行/有界网格；后续精确commit重验另记。
- 主控第二次布局返修要求已实施：去掉重复可视标题、压缩手机统计到一排、保留卡片内留白、来源限制简短提示和详情说明；顶部搜索按项目/目录/备注（会话、提案页分别按实际字段）展示。看板“清除筛选”同时清除页顶查询，两个查询仍取交集。W2最终采样 `be3cbfdbf536053f196ae5c48afad5fbf48138c2` 和 W0组装 `967ed09eb50c988fbd9e04c99f48e91a4c023f47` 已普通消费。
- 固定 `c07fb112ffd2497d4a913540f6836793e60a13db` 真实完整看板流程 **1 PASS /47.4秒**，111目录/71项目/43限制partial；SQLite人工字段、刷新、同库新进程、四视图、键盘/宽度与权限/作业不变均核实。最终生产布局 `acbfc432aceec14fcafe3ad37d3845b994f450ef` 真实预览 **1 PASS /24.2秒**，1920×1080/1280×720/390×844无横溢，首卡top604.7/604.7/769.6，含桌面至少80px首卡可达断言，未重复字段/模型任务。
- 最终 UI/测试 SOURCE `2aa57e264aa56a7d7f271c5e087bc540f3f862ab` 已push，与 `acbfc43` 生产前端完全相同。最终定向导航/三处错误传播 **4 PASS /22.2秒**；W0原4RED、W3首修2RED和标题选择器漏字2RED保留独立原件，不称整套重跑。构建/TS、定向lint PASS，61前端单测先前PASS保持独立范围。
- 完整范围与原件见 `docs/acceptance/S1-paper-board.md`。根新发现Git提交未进汇总活动，W2原owner独占返修中；依根消息暂缓最终worker_done，等待修复及根验收交接。论文/插件/进度仍未决定，不自行补齐。
- 22:16收到 W2 `cd201962997e69bb4772d73a6d8752874085c3ef` / REPORT `9319374f4a55cf88901e2ed77b439f45aaea1b89`，普通消费为最终 SOURCE `90cee4eda3c5f069325751064ecc944a303f6f77`，已push；UI/public与最终接受布局无差异。W2真实来源缓存服务重投影7→17已知时间（metadata adapter模拟边界明确），未扫盘或写缓存。根22:17明确W0负责后续真实缓存HTTP GET/人类字段不变与最终集成，W3可结算独立UI；根已看并接受1280/390截图。W3不重复启动发现或付费行为，最终报告已补记该分工及未执行边界。
- 协调问题 `msg_de7eb50d3762` 已正式回复并关闭：同意W3独立UI结算、无需再启动服务；root负责必要个人预览升级后的实际GET与活动显示确认。所有待答论文/插件/进度选项仍PENDING，源码/报告push后按原Dispatch发送唯一worker_done并停止。
