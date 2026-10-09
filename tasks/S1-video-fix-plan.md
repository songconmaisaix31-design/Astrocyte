# 视频导入修复与公开收藏夹后续切片

2026-10-10；基线 b2a7bbcfff8bb43c43b18314d6d6c7cf2527ca2f，沿用 s1/attention-materials-20261009。用户要求先修好所选视频导入；随后支持链接、账号公开收藏夹更新与批量选择，Agent提供反馈，最终由人挑选入库。用户已明确允许修改所选来源的代理DNS配置。Python允许，Go/TS/SQLite与summarize0.25.1继续沿用。

第一切片终点：所选 BV1PReT6EEqR 从实际页面提交，经 Service 获取真实字幕或上游本地音轨转写，SQLite/对象保存来源和正文，真实重启后可读；不创建Mission、不触发应用LLM、不读取cookies，不用网页推荐文字替代视频。先解决Clash Party/Mihomo fake-IP；仅增加B站公开来源真实DNS例外，保留其他代理规则和上游检查。若后续实际失败，退对应原Worker修复。原未知Codex调用不重放。

| 轨道 | 原worktree/branch与独占write_paths | 当前任务 |
|---|---|---|
| W2 导入适配 | s1-storage-1009；internal/adapters/importers、objects、sqlite、distillers、tasks/S1-W2.md | 消费准确基线，修实际字幕/音轨/转写路径和provenance；真实公开视频验收，保留实际失败；不得修改配置/依赖/入口/领域/前端或宿主DNS |
| W0 入口与集成 | s1-contract-1009；cmd、scripts、根依赖/锁、contracts、app/contracts.go、httpapi、foundation、migrations、web/src/api、README、tasks/S1-W0.md | 检查既有运行配置、真实媒体耗时与作业期限/取消；只修实际必要入口问题，公共文件唯一owner；W2完成后普通合并，不重写领域 |

主控负责本计划/决定/状态、用户已授权的宿主DNS例外、发布准确源码、独立页面/API/持久化验收、最终check/build/适用浏览器回归和commit/push。使用本机Orca Codex，模型以实际运行观测记录。两个Worker不得并行跑浏览器或同时下载转写同一个视频；首先由W2跑媒体正路径，主控最终页面验收。无自研调度器、Manifest或完成证明代码，公共契约未决定不冻结。

后续收藏夹切片已收到需求，尚待对齐：首批是否B站公开UID/收藏夹而无登录；先同步标题/简介待选列表供Agent反馈再勾选提取，还是先取字幕再勾选；实际绑定账号/收藏夹与同步频率。同步发现不是全量资料入库，用户决定选择哪些；不从收藏夹绑定推导全库读取、登录凭据或自动Mission授权。未答前不写账号/同步/Agent筛选新契约或大改页面。
