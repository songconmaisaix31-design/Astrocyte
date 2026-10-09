# S1 summarize 正式接入草案

2026-10-09，基线 `cad91f0f684f9bd222968ffd8e02ba34f1f921f3`。用户要求把 <https://github.com/steipete/summarize> 植入 Astrocyte，有问题先对接。随后明确允许Python，原无Python门槛不再适用；本轮优先复用summarize/yt-dlp/转写依赖，沿用 Attention、Go/TS/SQLite、现有作业/对象存储与本机 Codex CLI。本文用于后续原轨恢复，未回答的产品选择不冻结为接口或组件。

接管基线：arXiv 调用全局 summarize 0.21.8 的 extract-only 路径；视频仅接收既有 JSON/Markdown。W0 在 `0b7ec9642fc3b37d1b485b0b64eec2a30a05c0ee` 发布项目依赖0.25.1、默认启动与 `summarize_url`；最终组装 `d8f29070073e6f2e71ae6fcca5fa259205208b8f` 已普通合入主控源 `81e92ae521fadf6def68b771e005bd77cc3f7ed3`。全局0.21.8未升级。MIT、Node>=24，与当前Node24环境相容；真实媒体工具可用性和所选视频成功分别验收。

建议纵向切片：人在 Astrocyte 提交链接 → summarize 获取正文/真实字幕 → 首次整理 → 保存固定版本来源、原输出和整理记录 → 资料域中查看并人工分类/引用。使用上游项目依赖，具体选择CLI或core由适配核查决定，不复制整个仓库并维护分叉。首次整理与后续主题/项目关联分开，排序不扩大Agent读取权；旧Codex未知调用不自动重发。

## 已询问、待答

1. 本轮先完成Astrocyte内的链接处理，还是同时纳入summarize浏览器扩展？
2. 指定B站公开字幕要求登录时，是否允许利用用户登录浏览器取得选定视频内容，并对缺字幕音轨做本地转写？或只用可获得字幕，缺失时报错？登录/配对由人完成，不读取凭据原件。
3. 同一无版本视频重复导入：普通导入复用已有结果、主动“刷新来源”才重新获取，或每次新提交都重新获取检查变化？现有作业对这类可变URL按新命令重新读取，资料版本去重仍可能重复下载/转写；未答不把命令重放通过写成新表单重复通过，也不自行冻结刷新行为。
4. 本机代理DNS将所选B站域名返回198.18.0.160、arxiv.org返回198.18.0.40；上游0.25.1网络检查拒绝。已询问用户是否配置公开来源真实DNS例外；未答不改宿主代理、不移除检查、不自动回退旧版本，真实正文正路径保持阻塞。

## 实际约束

- [上游媒体文档](https://github.com/steipete/summarize/blob/main/docs/media.md)：字幕优先，通用媒体fallback可能调用yt-dlp；[yt-dlp依赖](https://github.com/yt-dlp/yt-dlp#dependencies)是Python。用户现已允许，可纳入实际依赖与可用性检查；仍复用上游，不另建视频下载/转录系统。
- 上游Chrome扩展有浏览器WebCodecs/Whisper转写，但登录资源支持需实测；CLI文档明确嵌入媒体不处理auth/cookie。不能从浏览器已登录推断指定视频可采集。
- [扩展文档](https://github.com/steipete/summarize/blob/main/apps/chrome-extension/README.md)：扩展/daemon有安装和配对步骤，Windows npm CLI缺打包native-host exe时daemon模式受限；默认8787与Astrocyte API相同，若选择该模式必须独立配置端口。
- 初始本机只有其他软件的ffmpeg，不能作为项目默认依赖。当前独立安装 yt-dlp/ffmpeg/ffprobe/whisper.cpp/多语言模型，版本与安装入口见README；本地工具启动/模型加载通过，所选视频真实音轨尚未执行成功。
- 本轮没有发应用 LLM 请求、安装扩展/daemon、读取登录凭据、升级全局summarize或改变已有资料权限。媒体依赖已独立安装，whisper.cpp 与多语言模型已实际加载；本地语音运行检查不代表指定视频转写成功。

## 原轨写域与验收

| 轨道 | 固定原worktree/branch与write_paths | 工作 |
|---|---|---|
| W0 | s1-contract-1009；contracts/、app/contracts.go、httpapi/、foundation/、cmd/、migrations/、web/src/api/、scripts/、根依赖锁（含pnpm-workspace.yaml）及README、自己的契约/报告 | 依赖版本/安装入口、公共配置和契约唯一所有者；最后普通集成；不重写领域 |
| W2 | s1-storage-1009；sqlite/、importers/、objects/、distillers/、自己的报告 | summarize实际调用/版本/provenance，真实字幕与缺失错误、取消和持久化；跨轨只交Handoff |
| W1（仅需用例变更时） | s1-domain-1009；attention/domain/、app/排除contracts.go、自己的报告 | 复用现有作业完成首次整理关联，固定输入及去重；不新增调度框架 |
| W3 | s1-attention-ui-1009；web/src/排除api与Workspace/Swarm、原自己的e2e/报告写域 | 链接提交、实际能力/作业反馈及原文/首次整理展示；不同时改三页 |
| W4 | s1-acceptance-1009；tests/s1/、web/e2e/s1.spec.ts、docs/acceptance/S1.md、自己的报告 | 隔离测试默认禁用外部提取，显式真实验收开启；API/持久化/重启检查 |

当前先恢复W0/W2，开发两种产品选择都需要的项目内公开链接处理，使用原worktree/branch和互斥写域。W0维护项目依赖/启动配置，W2实现既有summarize适配；登录/扩展分支继续待答。先发布必要契约再恢复W1/W3；Worker持续开发/测试/返修/commit/push。公共接口、迁移序列、依赖锁与程序入口只有W0写。主控只维护本计划/决定/状态、独立验收、最终提交与push。原W4验收文件仍归W4，需要改测试才恢复，不抢写。

当前实际恢复W0 `ctx_5534a77ba1f3`、W2 `ctx_538fb321f712`、W3 `ctx_fbe66c4a9ebd`，Orca本地Codex，运行时观测模型gpt-6.1-sol。W3先准备现有接口上的链接/既有导出模式，owner交接后组装；不实现待答扩展/登录分支。W0首次pnpm安装因忽略构建策略退出，pnpm-workspace.yaml按原根依赖职责明确纳入W0写域，经必要性与实际加载核验后决定脚本策略，不新增并行配置owner。

W4 恢复 `ctx_ec1a29e5ef2c`，沿用原 worktree/branch/write_paths，先消费 W0 入口变更，修正普通测试禁用提取与显式 live 开启，之后验收集成源。公开 B 站 yt-dlp 元数据已取得，无 cookies；实际音轨下载/转写仍由 W2 核验。B站通用CLI不能直接识别嵌入音轨时，允许W2薄适配调用锁定上游媒体实现，分别保留原CLI输出与上游真实转写来源；不伪装为CLI总结。首次整理仍需实际模型结果，旧未知调用不重发。

验收使用页面真实提交、Service、SQLite和对象存储及实际重启；重复链接/固定版本不重复工作，实际内容更新保留新旧版本；保留已有arXiv多版本通过结果并在新源回归。视频必须是所选视频的真实字幕/转写/总结，网页推荐文字不能代替。首次模型整理必须有真实输出、固定输入和实际客户端/模型记录；缺失/未知不补造。执行适用检查、构建与浏览器操作后再标完成，范围仍是S1 Attention。

## 2026-10-10 收口

W0/W2/W3/W4 已推送并普通集成，当前四轨终端均结算释放；原工作树/分支保留。W3 原 ctx_fbe66c4a9ebd 在 native 内存分配失败后，主控确认进程退出、停止原 dispatch，按同任务 retry-of 恢复 ctx_fa699bc72681；沿原 UI 写域完成中文失败下一步与两尺寸真实页面检查，不新增 Agent 或重做业务。实际客户端 Orca Codex / gpt-6.1-sol。

真实 B 站页面提交保存为网络失败，原命令重放/实际 SQLite 重启不重复提取；不是成功视频、新表单复用或首次模型整理。依赖加载、主控 check/build通过，准确源81e92ae完整 `pnpm test:e2e --workers=1` 为146 PASS/4 SKIP、4.2分钟，4个跳过的正/负live场景分别保留，不能替代成功视频。主控目视复核 W3 1280实际失败/草稿恢复截图。最新5173/8787预览已受控重启，原个人库保持空资料/作业/候选/Mission、处理器available=false。旧 native UNKNOWN未重试。尚待上述用户产品/网络决定、真实字幕与升级arXiv正路径、最新模型结果和授权Agent读取正路径；完整S1仍未通过，详细结果见STATUS顶部。
