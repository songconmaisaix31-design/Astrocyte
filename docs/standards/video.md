# 视频导入实现与实际验证

## 执行范围

第三层 V，Task `task_af564cbc003b`，Dispatch `ctx_1b382b548db7`，模块 Run `run_dbf1a68bc692`。实际客户端由模块公开 worker-show 核实为 Codex / GPT-6.1-Sol high fast。工作树 `product-video-20261014`，分支 `songconmaisaix31-design/product-video-20261014`，接管基线 `03f92127a74e4d8b49cf5c4bb1010b43349d6d43`，接管时 clean。

写入范围：`extractor.go`、`extractor_test.go`、`summarize.go`、`summarize-media.mjs`、`importers_test.go`，同目录新增 `video_*_test.go` 和本报告。共享配置及依赖由 B 负责。正式 V4、engine、schema、mapper、权重、日程及 17321 部署保持冻结。

## 显式匿名代理阶段

SOURCE：`06372a6a82c2942e02c4a46b69cc3f518f4ec5ae`。

`SummarizeOptions.ProxyURL` 为 string，默认空。构造时验证匿名 HTTP(S) origin，拒绝 userinfo、非 HTTP(S) scheme、query、fragment 和非根路径。显式配置后，独立子进程环境加入 `HTTP_PROXY`、`HTTPS_PROXY`、`NODE_USE_ENV_PROXY=1`、`NO_PROXY=localhost,127.0.0.1,::1`。未显式提供时不读取宿主 proxy、cookies 或凭据。B 的配置入口为 `ASTROCYTE_SUMMARIZE_PROXY_URL`。

官方依据：[Node 环境代理](https://nodejs.org/api/cli.html#node_use_env_proxy1)说明启用后 Node 在启动时读取 HTTP_PROXY、HTTPS_PROXY 和 NO_PROXY。实际本机 Node 为 24.16.0，summarize 为项目锁定的 0.25.1。

验证命令：

```text
pnpm install --frozen-lockfile
go test -mod=readonly ./internal/adapters/importers -run '^TestVideoExplicitAnonymousProxyIsolation$' -count=1 -v
git diff --check
gofmt -l internal/adapters/importers/extractor.go internal/adapters/importers/video_proxy_test.go
```

结果：依赖安装 PASS，锁文件未修改；代理隔离及无效输入拒绝 PASS，实际 CLI version 四次检查 PASS，包运行 7.694 秒；diff 和格式检查 PASS。首 RED 原件 `.verification/video-proxy-first-red.txt` 保留。SOURCE 的直连 push 和显式代理 push 均首次返回 `SSL_ERROR_SYSCALL`，schannel 单次握手失败也保留；23:23 后普通 `git push origin HEAD` 成功，代理 SOURCE/REPORT 与下述时间 SOURCE 均已 push。

## 视频时间证据基线

公开固定来源为 `https://www.bilibili.com/video/BV1PReT6EEqR/`。B 固定原件目录为 `C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/.verification/video-source-third`。媒体原件的正文为 32908 bytes，`provider=whisper.cpp`、`error=null`、`segments=null`；CLI 原件为 `strategy=html`、`textProvided=false`、`transcriptSource=null`。已有生产导入与重启正文验证 PASS，时间 segments/spans 为 0，时间验收未通过。

实际安装的 summarize-core 0.25.1 的 `transcription/whisper/whisper-cpp.js` 使用 `--no-timestamps` 和 `--output-txt`，返回 text/provider/error/notes，缺少 segments。`content/transcript/providers/youtube/yt-dlp.js` 将缺少的 segments 保留为 null。Go 已读取返回的 segments。

模块已批准在授权桥接文件内复用上游 `acquireYtDlpMedia`、cleanup 和 `spawnTracked`，读取同一本机 whisper-cli 的原生 JSON 时间输出。官方 [whisper.cpp 1.9.2 CLI](https://github.com/ggml-org/whisper.cpp/blob/v1.9.2/examples/cli/cli.cpp)定义 `--output-json`，JSON 的 offsets.from/to 为毫秒。桥接验证整数、非负、单调、不重叠时间及正文归属。原生程序一次生成 TXT 和 JSON；原生 JSON 从固定本机文件交 Go 读取，不进入桥接 stdout。

原始转写、原模型输出、凭据和测试原件仅保留于本机忽略目录及对象存储，不进入 Git、HTTP 或 release 包。正式公开响应沿用现有派生 SourceSpans 契约。

## 真实时间、刷新与重启阶段

SOURCE：`d48a29b6e1744465666edbfa28be919531ca39a4`，已普通 push。

```text
go test -mod=readonly -tags video_live ./internal/adapters/importers -run '^TestVideoProductionHTTPRefreshTimelineRestart$' -count=1 -v -timeout 35m
go test -mod=readonly ./internal/adapters/importers -count=1 -v
go vet -mod=readonly ./internal/adapters/importers
node --check internal/adapters/importers/summarize-media.mjs
go build -mod=readonly -o .verification/video-server.exe ./cmd/server
git diff --check
```

时间核心首次实际结果 PASS，测试 249.72 秒、包 251.500 秒。实际 HTTP 提交旧原件 local_file_ref 后，再显式提交相同公开 BV 的 summarize_url refresh=true；生产 Reader/Service 自然绑定同一 SourceKey 和 MaterialID，两版正文为 32908 / 35807 bytes，current_revision=2。新版本保留 952 个实际时间片段和 4 个附件，原生 JSON 段数与 SourceSpans 全等；原标题与来源从实际 CLI 保留，未提供模型摘要。旧成功正文与 time0 未通过原件保持。

普通 tracking URL 别名提交复用同一新作业，实际 HTTP 服务关闭、新服务启动与 SQLite 关闭重开后，两版正文对象、原件附件和材料投影逐字节 / 字段全等。真实 SQLite 行数为材料 1、版本 2、作业 2、沉淀 0、候选 0、Mission 事件 0；人工 / Agent 使用行数均为 0。本次实际 ASR 使用 whisper.cpp 1.9.2 / ggml-base multilingual，Agent 沉淀模型调用为 0。首次原始安全日志末行字段 `no_model_calls` 的检查对象为零 Agent 沉淀，最终 SOURCE 日志字段为 `no_agent_model_calls` 并明确记录 ASR 模型；未重发成功 ASR。

实际工具绝对路径由模块准确核实既有安装及 B 随后确认：`C:/Users/DW/AppData/Local/Astrocyte/media` 下的 `yt-dlp-2026.08.19/Scripts/yt-dlp.exe`、`ffmpeg-9.0.2/ffmpeg-9.0.2-essentials_build/bin/ffmpeg.exe`、`whisper-1.9.2/Release/whisper-cli.exe`、`models/ggml-base.bin`。本次显式 ProxyURL 为 `http://127.0.0.1:7890`；未读取系统代理。

本机证据目录：`C:/Users/DW/orca/workspaces/Astrocyte/product-video-20261014/.verification/video-timed-first`。准确原件为 `state.sqlite`、`objects`、`summarize-upstream-media.json`、`summarize-cli-original.json`、`summarize-cli.stderr.txt`、`whisper-original.json`、`video-original-text.txt`；安全运行日志为 `.verification/video-timed-first.log`。首新作业 `KICSCKJUSDQE4PKYKZDFIWC64R`，材料 `534HWFW7E5GEF75ZY4F7ETPLQ2`。

默认 importer 测试在本次运行 EXIT 0，27.576 秒，9 个既有条件 SKIP；此运行尚包含既有替代来源，不能作为全部默认真实来源整改完成。Node 语法、Go vet、实际服务二进制构建与 diff / 格式检查 PASS。

当前剩余边界：原生 `whisper-original.json` 在既有附件服务中仍可被读取，V 未请求其 HTTP 原件；生产私有读取拒绝已 Handoff 给 B，需精确 SOURCE 的实际 HTTP 拒绝验证后才通过完整视频验收。默认测试的全部真实原件整改仍继续；组合产品和 release 包验收由原模块人员继续。
