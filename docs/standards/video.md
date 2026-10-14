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

本轮原件、凭据和测试证据仅保留于本机忽略目录及对象存储，不进入 Git 或 release 包。原生 whisper JSON 只允许本机对象读取，HTTP 拒绝；现有公开 summarize 附件保持原有读取权限。正式材料响应沿用现有派生 SourceSpans 契约。

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

私有原件边界已普通消费 B 精确 SOURCE，并完成下方组合 HTTP 验证。组合产品完整默认检查、浏览器和 release 包验收由原模块人员继续。

## 默认真实输入阶段

SOURCE：`a54fe8e238a2e52db98c801012c2ccdfd64edbd5`，已普通 push。

保留既有测试名称和来源、版本、摘要分离、时间、canonical identity、digest、本地文件权限及取消拒绝断言。默认视频输入从实际 CLI / media 原件逐字节读取，按生产桥接字段解析；独立读取原生 whisper JSON，对全部 952 段逐项核实毫秒 from/to、文本归属与公开 SourceSpans。页面原件实际拒绝；已有纯文本导出保持缺少可核时间的状态。非法 JSON 字段仅用于拒绝检查。

论文来源正向断言使用实际固定 arXiv `2501.12948v1` 的 summarize CLI 原输出；原 PDF / metadata 断言使用官方实际 `2501.12948v2`，未发布 v999999 与真实 HTTP cancellation 直接失败。显式代理还在构造时拒绝 0、超出 65535 和非法端口。

默认视频证据必须显式提供：

```text
ASTROCYTE_TEST_VIDEO_BASELINE_DIR=C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/.verification/video-source-third
ASTROCYTE_TEST_VIDEO_TIMED_DIR=C:/Users/DW/orca/workspaces/Astrocyte/product-video-20261014/.verification/video-timed-first
go test -mod=readonly ./internal/adapters/importers -count=1 -v
node --test internal/adapters/importers/paper-dom.test.mjs
```

缺少真实原件配置直接 FAIL。没有新增 Skip，没有把任何已有默认测试移入 video_live 标签。默认 importer 全包在真实输入整改后 EXIT 0，26.182 秒，仍有 9 个其他既有 opt-in SKIP；V 专属原件、时间、代理及原测试改写均执行。随后固定论文 CLI 原件正向补验 8.940 秒 PASS，最终 Reader / 代理端口补验 6.770 秒 PASS。Node 9 PASS / 0 SKIP，Go vet、diff 与格式检查 PASS。

首编译 RED（未使用的 fmt import）在 `.verification/video-default-real-first.log` 保留；后续结果为 `.verification/video-default-real-second.log`、`.verification/importers-default-real-final.log`、`.verification/paper-export-real-final.log`、`.verification/video-default-final-targets.log` 和 `.verification/importers-node-final.log`。

真实证据的具体范围：固定 BV 的原 CLI JSON 只有页面输出，CLI transcriptSegments 专用正向分支及 summary-only JSON 分支仍没有真实原件，证据 NOT_RUN；时间成功通过实际 CLI + media/native 合并解析链和生产刷新链核实。公开旧纯文本导出正向与缺时间状态已经核实。共享 `roundTripFunc` 定义仍由其他 B 测试文件依赖，V 自身没有使用替代 HTTP Transport；其他 B 文件的替代来源整改由 B 持续负责。

已按模块准确授权，普通消费 B 私有附件 SOURCE `cc46b9855775adf70bce74242d08ef84ff4f573b` / REPORT `ee4eeb11dfd7c2e598734ab8ea3da6efd837afd4`；没有手改 B 生产域。

## 组合源码与私有原件 HTTP 验证

组合 SOURCE：`08b3594fe4a3a02d06fead00ee6ead98b9e38bbc`，普通 merge 无冲突，已 push。它包含 V 全部 SOURCE / REPORT，以及 B 对私有附件和显式配置入口的准确交付。

```text
ASTROCYTE_TEST_VIDEO_PUBLICATION=C:\Users\DW\orca\workspaces\Astrocyte\product-video-20261014\.verification\video-timed-first
node tests/s1/private-video-attachment.mjs
go test -mod=readonly ./internal/adapters/importers -count=1 -v
go vet -mod=readonly ./...
go test -mod=readonly -tags video_live ./... -run '^$'
go build -mod=readonly -o .verification/video-combined-server.exe ./cmd/server
node --check internal/adapters/importers/summarize-media.mjs
git diff --check
```

私有原件实际 HTTP 测试 EXIT 0：只读原真实 SQLite，经 Node SQLite backup 建立自有副本 `.verification/astrocyte-s1-y3xYnU`，只按实际材料对象引用复制原件。真实人类 / Agent 身份、普通 / 编码名称、启动及重启后共 8 个 `403 scope_denied`；其他 3 个已有公开附件共 6 个 HTTP 200，字节与本机原件全等。952 段原生 JSON 本机对象逐字节保持，拒绝没有增加机器使用记录，实际 API 进程退出和新进程重启通过，本次新 ASR 为 0。安全日志 `.verification/video-private-http-final.log`。

组合默认 importer 全包首场 EXIT 1，39.208 秒，唯一失败为 B 的 `TestScholarExactDOIRouting` 请求真实 Crossref 返回 EOF；原件 `.verification/importers-default-combined-final.log` 保留。实际定向恢复 PASS，2.767 秒；同精确 SOURCE 后续全包 EXIT 0，37.468 秒，9 个既有条件 SKIP，日志 `.verification/importers-default-combined-recovery.log`。本轮没有代理 / DNS / SSL 全局修改，没有更改失败断言。

全仓 Go vet、带 video_live 的全部 Go 测试编译、服务构建、Node 桥接语法与 diff 检查 PASS。`-run '^$'` 只证明全部包及测试文件编译通过，没有执行其余全仓业务测试；日志 `.verification/video-all-compile-final.log`。

当前真实剩余限制：上述 9 个原有其他范围 opt-in 未执行；CLI 自身时间专用 / summary-only JSON 的实际原件证据不足；共享 Transport helper 的其他 B 使用尚待其 owner 整改；转写进行中的视频取消尚未实际验证。已验证的时间、自然同源版本、重复复用、重启与私有原件 HTTP 权限保持 PASS，未重新运行成功 ASR。完整产品 check / 浏览器 E2E / Windows ZIP 解压和发布由所属模块继续，V 保持模块功能及包支持。

## 真实原生进程启动后取消

SOURCE：`e9e5f52511ca44a947d39fff61fac098c6fbad74`，已普通 push。模块明确授权固定公开 BV 的新取消故障路径。

`SummarizeOptions.ObserveMediaProcess` 是可选的本机回调，默认 nil；它复用固定上游 `setProcessObserver` / `spawnTracked` 的真实 PID 与退出事件，提供安全 kind/phase/pid/exitCode，以及原观察器实际提供的 progress 值，不改变公开 HTTP DTO。Go 使用标准 JSON Decoder，环境、命令参数、原始 stderr 和正文不进入回调。

```text
go test -mod=readonly -tags video_live ./internal/adapters/importers -run '^TestVideoNativeStartupCancellationHTTPProcessRestart$' -count=1 -v -timeout 8m
go vet -mod=readonly ./internal/adapters/importers
node --check internal/adapters/importers/summarize-media.mjs
git diff --check
```

实际首次取消验证 PASS，18.73 秒 / 包 20.371 秒。专属 Go 子进程运行生产 Service、HTTP、SQLite、ObjectStore 和 Reader；在原上游观察器报告 whisper.cpp 实际 PID 启动后，父验证器通过原 HTTP CancelJob 取消真实 running 作业。真实进程记录为 Go 服务 `59100`、Node 桥接 `48580`、字幕 Node `54116`、yt-dlp `30252`、whisper `34604`；whisper 退出 1，Node 桥接退出 130。Windows 实际进程句柄 Wait 核实所有上述所属进程退出。

独立新 Go 服务进程 `48172` 重开同库并运行原 worker / RecoverJobs；作业保持 cancelled / attempts=1 / 同 operation_id，没有材料或版本，普通重复导入复用同一取消作业，重启没有新提取进程。两次 Go HTTP 进程均实际退出，真实 SQLite 材料 / 版本 / 沉淀 / 候选为 0，作业为 1。没有 Agent 沉淀调用，也没有重做已成功转写。

本机安全证据为 `.verification/video-cancel-first.log` 及 `.verification/video-cancel-first/owned-process-events.jsonl`、`go-service-1.log`、`go-service-2.log`、`state.sqlite` / `objects`；此前所有成功和首失败原件保持。

转写中途进度证据仍 NOT_RUN：本机 whisper-cli 1.9.2 官方 `--print-progress` 回调通过 fprintf 输出文本；当前时间桥接没有可复用的结构化 native progress 出口。没有自行解析 stderr 或把定时器 / 任意输出标作转写开始，准确验收范围为实际 whisper 进程启动后的取消、清理和重启不续发。
