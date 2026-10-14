# Workbench v0.1 集成执行单

事实源：用户修订版任务书、tasks/Workbench-v01.md、tasks/Workbench-api.md。目标workbench/v0.1-workbench；validation和main不改。两轨结算后仅一个集成owner在独立worktree/branch执行，root负责最终Git与独立验收。

精确业务SOURCE：backend8635895e40eb022e92a86870861b4d09fd168147；frontend8ec40f7633236b80bf12e34aa6468633f96d4072（已普通消费backend SOURCE）。精确REPORT由root收到worker_done并核对remote/clean后补充Handoff，禁止合未知移动分支HEAD或未提交文件。普通no-ff合各REPORT，保留SOURCE/REPORT祖先；冲突若为领域行为，退原owner，集成只允许装配胶水。

独占写域：docs/workbench/integration.md；仅真实必要薄胶水允许cmd/server/main.go、internal/adapters/httpapi/server.go、web/src/App.tsx、web/src/api/schema.d.ts；涉及业务变化先Handoff。禁止引擎/events/Sleep/迁移/参数/依赖/旧spec/Playwright配置修改。不得把raw、快照、配置、私人路径载荷、截图或真实UI取证JSON放进Git。领域SOURCE路径须与最终树一致，REPORT只变对应报告文件。

集成owner执行git diff --check、普通祖先与冻结域检查、generate --check；整树pnpm check/build和完整默认E2E由root在最终目标执行，不重复昂贵全套。已有两视口认知12、Nav104、真实live2和十值独立SQL通过，保留首RED。没有新胶水时业务树不变。

部署：复用scripts/workbench/start.ps1，最终认知web/dist由同集成树构建后绑定本机config.web；先查LOCALAPPDATA/astrocyte-workbench/v0.1/bin/astrocyte-workbench.exe的PID和17321监听，只有完全同路径的已拥有Workbench实例可停止后升级。不得停止其他个人/V4实例。使用-Port17321 -SkipSnapshot保留已核代际；配置/日志/bin/PID仅本机。复核/cognition/today静态页、共享session和认知GET健康；确认Windows00:30快照/登录任务保持Ready。提交报告给出精确SOURCE/REPORT、命令/结果、ownedPID/地址及剩余限制，commit/push后worker_done。

只读监控NOT_RUN：现有非管理员token不能进行Procmon驱动捕获；官方签名工具和scripts/workbench/monitor.ps1已提供，只能在用户管理员shell实测，不以静态/hash代替。连续两日00:30需真实跨日记录，当前安装和手动刷新不算两日通过。最终实施可交付，这两项全验收保持未完成。
