# Astrocyte Workbench v0.1

用户2026-10-13任务书：三天只读个人工作台，消费已验收资产；React/Vite四视图，薄Go SQL查询与静态服务。分支workbench/v0.1从main建立。main基线3a02ce38481d1885adee2e051a1b524d78376897仅LICENSE；独立模块置workbench/，不合validation，不改web/。docs/demo-data当前不存在，历史输入采用V1R已验收导出件同源本机文件，不提交快照/raw。

## 范围与数据

正式源：C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w1/formal-20261013-first/state.sqlite。历史：同根v1r-w1/replay-002/topic-timeseries-safe.json、daily-top5-safe.json。快照/本机配置/日志/二进制默认存LOCALAPPDATA/astrocyte-workbench/v0.1，不进Git。仅在线备份程序可读正式源；Go服务只挂快照mode=ro，不接受正式库作为服务数据库。

每日00:30在线备份，原子发布完整副本，支持手动刷新。正式库及其WAL/SHM不写；并发生产写入时备份需能完成且不阻塞生产。API事件严格投影event_id/kind/topic/tokens/scope/ts；source只用于SQL过滤和分组聚合，不透传raw或observation。A/M/C、环比、Sleep状态、轨迹、kind×日统计是允许的聚合。Today以Asia/Shanghai真实自然日；空日/尚未产出Sleep显示实际状态，不挪历史事件冒充今日。

## 两条互斥开发轨

| 轨 | 写域 | 交付 |
|---|---|---|
| backend | workbench/backend/**、scripts/workbench/**、workbench/docs/backend.md | 独立Go模块、在线快照、SQL白名单API/静态服务、启动/常驻/00:30任务、只读与raw审计 |
| frontend | workbench/frontend/**、workbench/docs/frontend.md | 独立React/Vite四路由、图表/抽屉/过滤/分页、真实API浏览器验收 |

每轨Codex --yolo、一个Agent/独立worktree/branch；开发测试返修报告同owner，SOURCE与REPORT普通commit/push。root仅任务/契约/状态/独立验收。两轨完成后一个集成owner普通no-ff合并，仅补启动/路由/配置接缝；领域问题退原owner。所有原validation路径、V4协议/参数/权重/引擎、正式库、web/与原E2E不写。不得构造自研调度/完成证明系统。

## 视图

Today：A/M/C焦点卡片及日环比，source过滤倒序活动，昨夜Sleep徽章，7日事件计数迷你图。Focus：多主题历史与实验期拼接A(t)，按A排序列表，点击打开kind×日统计和六字段证据列表抽屉。Agents：source分组今日/7日/30日计数及最近事件。Evidence：topic/kind/日期/source过滤、分页六字段表。只读无写按钮；手动刷新走本机快照脚本。

## 验收与时间

Go适用测试、前端typecheck/build、四路由真实浏览器渲染；固定10条UI数值独立SQL核对；完整正常/错误API零raw审计；实际进程文件系统监控证明正式DB/WAL/SHM零写。一条命令启动及本机后台常驻，真实00:30日程安装。连续两日自动成功只能在实际两次跨日触发后确认，本轮实现/手动运行不得替代。阶段保留清晰commit并push；首次RED与未运行验收保留。

Orca Run run_a75ccac871c9。当前已创建协调worktree并从main建立目标分支，开发轨待派发。v0.2仍冻结；本任务替换当前开发目标，不解除V4冻结。
