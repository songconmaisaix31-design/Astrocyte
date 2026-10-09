# 视频导入修复与公开收藏夹后续切片

2026-10-10；基线 b2a7bbcfff8bb43c43b18314d6d6c7cf2527ca2f，沿用 s1/attention-materials-20261009。用户要求先修好所选视频导入；随后支持链接、账号公开收藏夹更新与批量选择，Agent提供反馈，最终由人挑选入库。用户已明确允许修改所选来源的代理DNS配置。Python允许，Go/TS/SQLite与summarize0.25.1继续沿用。

第一切片终点：所选 BV1PReT6EEqR 从实际页面提交，经 Service 获取真实字幕或上游本地音轨转写，SQLite/对象保存来源和正文，真实重启后可读；不创建Mission、不触发应用LLM、不读取cookies，不用网页推荐文字替代视频。先解决Clash Party/Mihomo fake-IP；仅增加B站公开来源真实DNS例外，保留其他代理规则和上游检查。若后续实际失败，退对应原Worker修复。原未知Codex调用不重放。

| 轨道 | 原worktree/branch与独占write_paths | 当前任务 |
|---|---|---|
| W2 导入适配 | s1-storage-1009；internal/adapters/importers、objects、sqlite、distillers、tasks/S1-W2.md | 消费准确基线，修实际字幕/音轨/转写路径和provenance；真实公开视频验收，保留实际失败；不得修改配置/依赖/入口/领域/前端或宿主DNS |
| W0 入口与集成 | s1-contract-1009；cmd、scripts、根依赖/锁、contracts、app/contracts.go、httpapi、foundation、migrations、web/src/api、README、tasks/S1-W0.md | 检查既有运行配置、真实媒体耗时与作业期限/取消；只修实际必要入口问题，公共文件唯一owner；W2完成后普通合并，不重写领域 |
| W3 真实页面验收 | s1-attention-ui-1009；web/e2e/attention-video-import.spec.ts、web/e2e/summarize-env.d.ts、tasks/S1-W3.md；仅实际页面缺陷时改Attention组件 | 原成功URL opt-in场景使用强制禁用提取的ephemeral服务，不能运行真实成功链路；按既有helper明确开启所选公开源、关闭Codex，仅修真实验收入口；W2完成后再唯一运行，不与媒体任务并发 |

主控负责本计划/决定/状态、用户已授权的宿主DNS例外、发布准确源码、独立页面/API/持久化验收、最终check/build/适用浏览器回归和commit/push。使用本机Orca Codex，模型以实际运行观测记录。两个Worker不得并行跑浏览器或同时下载转写同一个视频；首先由W2跑媒体正路径，主控最终页面验收。无自研调度器、Manifest或完成证明代码，公共契约未决定不冻结。

后续收藏夹切片已收到需求，尚待对齐：首批是否B站公开UID/收藏夹而无登录；先同步标题/简介待选列表供Agent反馈再勾选提取，还是先取字幕再勾选；实际绑定账号/收藏夹与同步频率。同步发现不是全量资料入库，用户决定选择哪些；不从收藏夹绑定推导全库读取、登录凭据或自动Mission授权。未答前不写账号/同步/Agent筛选新契约或大改页面。

实际启动：Run run_75f5663450b7；W2 task_fbf46b3309c2 / ctx_55b0e7395a05 / term_f23619a6-1520-4c43-99b4-ac75ce926099，W0 task_6c5e48d158b5 / ctx_93524ed1e8dd / term_9969bda4-0a33-4bb5-ba7b-65ce9164849f；均本机Codex，实际模型gpt-6.1-sol。W0首启动turn未观测，读屏明确原prompt仍在composer，主控仅补Enter，随后原dispatch有真实执行/心跳；未重复派发。W0已复现Go直接启动pnpm逻辑CLI路径不能解析core，入口owner作最小realpath修复。

DNS实际操作01:07：备份并修改 `%APPDATA%/mihomo-party/mihomo.yaml` 与 `work/config.yaml`，仅blacklist追加 `+.bilibili.com`，原其余解析/路由保留；控制接口热加载204，所选B站公网解析与锁定上游检查PASS，arxiv.org未调整。备份后缀 `.astrocyte-before-video-2026-10-09T17-07-13-140Z.bak`。Clash Party图形窗口恢复未成功，client缓存未通过设置UI刷新；后续设置操作可能覆盖，应核验该限制，不伪称GUI保存通过。W2于network ready后开始唯一实际媒体验收。

01:19 主控决定：真实媒体桥01:08:37开始，音轨下载已结束，01:08:52启动的同一whisper进程持续CPU工作，实际转写已超过原300秒。W0仅将已有Reader与Service共享作业默认期限调整为1800秒，保留显式环境配置与86400秒上限；Codex执行上限仍为独立180秒，旧UNKNOWN不重放。不新增按业务域分派期限的契约或调度层。W3 task_82c3c3a4b5a7 / ctx_ea2d95222765 / term_82cfbb11-2a26-4a15-bc5b-ab35567dfdf3 已实际执行，先修正向验收服务与静态检查，W2结束前不启动第二媒体任务。
