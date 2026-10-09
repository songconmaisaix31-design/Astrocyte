# S1 summarize 正式接入草案

2026-10-09，基线 `cad91f0f684f9bd222968ffd8e02ba34f1f921f3`。用户要求把 <https://github.com/steipete/summarize> 植入 Astrocyte，有问题先对接。沿用 Attention、Go/TS/SQLite、现有作业/对象存储与本机 Codex CLI；默认链路无 Python。本文用于后续原轨恢复，未回答的产品选择不冻结为接口或组件。

当前实际接入：arXiv 调用全局 summarize 0.21.8 的 extract-only 路径；视频仅接收既有 JSON/Markdown。根依赖没有 summarize，页面没有视频链接自动处理入口。2026-10-09 官方 npm latest 与上游 package.json 均为0.25.1；该版本只是接入候选，尚未安装/升级或核验其实际输出。MIT、Node>=24，与当前Node24环境相容；不能从此推导Windows媒体/模型路径可用。

建议纵向切片：人在 Astrocyte 提交链接 → summarize 获取正文/真实字幕 → 首次整理 → 保存固定版本来源、原输出和整理记录 → 资料域中查看并人工分类/引用。使用上游项目依赖，具体选择CLI或core由适配核查决定，不复制整个仓库并维护分叉。首次整理与后续主题/项目关联分开，排序不扩大Agent读取权；旧Codex未知调用不自动重发。

## 已询问、待答

1. 本轮先完成Astrocyte内的链接处理，还是同时纳入summarize浏览器扩展？
2. 指定B站公开字幕要求登录时，是否允许利用用户登录浏览器取得选定视频内容，并对缺字幕音轨做本地转写？或只用可获得字幕，缺失时报错？登录/配对由人完成，不读取凭据原件。

## 实际约束

- [上游媒体文档](https://github.com/steipete/summarize/blob/main/docs/media.md)：字幕优先，通用媒体fallback可能调用yt-dlp；[yt-dlp依赖](https://github.com/yt-dlp/yt-dlp#dependencies)是Python。不得默认启用这条fallback，也不另建视频下载/转录系统。
- 上游Chrome扩展有浏览器WebCodecs/Whisper转写，但登录资源支持需实测；CLI文档明确嵌入媒体不处理auth/cookie。不能从浏览器已登录推断指定视频可采集。
- [扩展文档](https://github.com/steipete/summarize/blob/main/apps/chrome-extension/README.md)：扩展/daemon有安装和配对步骤，Windows npm CLI缺打包native-host exe时daemon模式受限；默认8787与Astrocyte API相同，若选择该模式必须独立配置端口。
- 本机PATH有Node和summarize，ffmpeg来自其他软件目录；无ffprobe、whisper-cli或sherpa-onnx。这只是文件发现，不是媒体可用验收，不自动把其他软件的ffmpeg作为项目依赖。
- 本轮没有启动模型、安装扩展/daemon、读取登录凭据、升级全局summarize或改变已有资料权限。

## 原轨写域与验收

| 轨道 | 固定原worktree/branch与write_paths | 工作 |
|---|---|---|
| W0 | s1-contract-1009；contracts/、app/contracts.go、httpapi/、foundation/、cmd/、migrations/、web/src/api/、scripts/、根依赖锁及README、自己的契约/报告 | 依赖版本/安装入口、公共配置和契约唯一所有者；最后普通集成；不重写领域 |
| W2 | s1-storage-1009；sqlite/、importers/、objects/、distillers/、自己的报告 | summarize实际调用/版本/provenance，真实字幕与缺失错误、取消和持久化；跨轨只交Handoff |
| W1（仅需用例变更时） | s1-domain-1009；attention/domain/、app/排除contracts.go、自己的报告 | 复用现有作业完成首次整理关联，固定输入及去重；不新增调度框架 |
| W3 | s1-attention-ui-1009；web/src/排除api与Workspace/Swarm、原自己的e2e/报告写域 | 链接提交、实际能力/作业反馈及原文/首次整理展示；不同时改三页 |

决定收到后按必要写域恢复最少Workers，先由W0发布变动契约，Worker持续开发/测试/返修/commit/push；公共接口、迁移序列、依赖锁与程序入口只有W0写。主控只维护本计划/决定/状态、独立验收、最终提交与push。原W4验收文件仍归W4，需要改测试才恢复，不抢写。

验收使用页面真实提交、Service、SQLite和对象存储及实际重启；重复链接/固定版本不重复工作，实际内容更新保留新旧版本；保留已有arXiv多版本通过结果并在新源回归。视频必须是所选视频的真实字幕/转写/总结，网页推荐文字不能代替。首次模型整理必须有真实输出、固定输入和实际客户端/模型记录；缺失/未知不补造。执行适用检查、构建与浏览器操作后再标完成，范围仍是S1 Attention。
