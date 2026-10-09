# 视频导入修复与公开收藏夹后续切片

2026-10-10；基线 b2a7bbcfff8bb43c43b18314d6d6c7cf2527ca2f，沿用 s1/attention-materials-20261009。用户要求先修好所选视频导入；随后支持链接、账号公开收藏夹更新与批量选择，Agent提供反馈，最终由人挑选入库。用户已明确允许修改所选来源的代理DNS配置。Python允许，Go/TS/SQLite与summarize0.25.1继续沿用。

第一切片终点：所选 BV1PReT6EEqR 从实际页面提交，经 Service 获取真实字幕或上游本地音轨转写，SQLite/对象保存来源和正文，真实重启后可读；不创建Mission、不触发应用LLM、不读取cookies，不用网页推荐文字替代视频。先解决Clash Party/Mihomo fake-IP；仅增加B站公开来源真实DNS例外，保留其他代理规则和上游检查。若后续实际失败，退对应原Worker修复。原未知Codex调用不重放。

| 轨道 | 原worktree/branch与独占write_paths | 当前任务 |
|---|---|---|
| W2 导入适配 | s1-storage-1009；internal/adapters/importers、objects、sqlite、distillers、tasks/S1-W2.md | 消费准确基线，修实际字幕/音轨/转写路径和provenance；真实公开视频验收，保留实际失败；不得修改配置/依赖/入口/领域/前端或宿主DNS |
| W0 入口与集成 | s1-contract-1009；cmd、scripts、根依赖/锁、contracts、app/contracts.go、httpapi、foundation、migrations、web/src/api、README、tasks/S1-W0.md | 检查既有运行配置、真实媒体耗时与作业期限/取消；只修实际必要入口问题，公共文件唯一owner；W2完成后普通合并，不重写领域 |
| W3 真实页面验收 | s1-attention-ui-1009；web/e2e/attention-video-import.spec.ts、web/e2e/summarize-env.d.ts、tasks/S1-W3.md；仅实际页面缺陷时改Attention组件 | 原成功URL opt-in场景使用强制禁用提取的ephemeral服务，不能运行真实成功链路；按既有helper明确开启所选公开源、关闭Codex，仅修真实验收入口；W2完成后再唯一运行，不与媒体任务并发 |

主控仅维护计划/决定/状态、获准的宿主DNS、独立验收与最终check/build/commit/push，不写业务代码。W2媒体完成后W3才运行唯一页面任务，另一尺寸复用结果；公共契约/迁移/锁/入口由W0唯一持有并普通集成。沿用现有框架，无自研调度器、Manifest或完成证明代码。

后续收藏夹切片已收到需求，尚待对齐：首批是否B站公开UID/收藏夹而无登录；先同步标题/简介待选列表供Agent反馈再勾选提取，还是先取字幕再勾选；实际绑定账号/收藏夹与同步频率。同步发现不是全量资料入库，用户决定选择哪些；不从收藏夹绑定推导全库读取、登录凭据或自动Mission授权。未答前不写账号/同步/Agent筛选新契约或大改页面。

执行：本机Orca Codex / 实际gpt-6.1-sol；Run run_75f5663450b7，W2 ctx_55b0e7395a05、W0 ctx_93524ed1e8dd、W3 ctx_ea2d95222765，三个原worktree/branch保留，完成后均commit/push并验收释放。入口首修RED与各轮结果保留在原owner报告，不重复派发或覆盖历史。

DNS操作：01:07备份 `%APPDATA%/mihomo-party/mihomo.yaml` 与 `work/config.yaml`，后缀 `.astrocyte-before-video-2026-10-09T17-07-13-140Z.bak`；仅blacklist追加 `+.bilibili.com`、热加载204，上游检查PASS，arXiv/其他路由不改。设置UI缓存未刷新，后续Clash设置操作可能覆盖例外。

期限决定：首次真实转写超过300秒，W0仅把现有Reader/Service共享默认改1800秒，保留显式1–86400秒设置；Codex执行上限独立180秒、旧UNKNOWN不重发，不新增领域预算接口。

交付：指定单视频从空库真实页面提交、原文/三附件、实际进程重启和同一结果两尺寸PASS；W2首轮730.29秒、W3作业275.98秒/用例1 PASS，正文同为32908bytes。主控业务源017c133ec4322a6d1011874eee25e148d814fd9c，check/build PASS，完整E2E146 PASS/4 SKIP（4.7分钟、exit0），最后报告合并a1a2d5213d2a17ede83c04ecbce95e308d687080不改代码。预览已更新，个人库不写验收资料、Codex关闭。

本切片终点达成，完整S1及收藏夹尚未完成。准确各轨提交、原始输出、ASR/无时间轴、未实测取消、待答权限/刷新与人工清理见 [STATUS](../STATUS.md)、[QUESTIONS](../docs/QUESTIONS.md) 和 [W0](S1-W0.md) / [W2](S1-W2.md) / [W3](S1-W3.md)。
