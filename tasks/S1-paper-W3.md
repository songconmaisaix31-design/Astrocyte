# W3 论文流程与项目看板

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
