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

结果：依赖安装 PASS，锁文件未修改；代理隔离及无效输入拒绝 PASS，实际 CLI version 四次检查 PASS，包运行 7.694 秒；diff 和格式检查 PASS。首 RED 原件 `.verification/video-proxy-first-red.txt` 保留。SOURCE 的直连 push 和显式代理 push 均首次返回 `SSL_ERROR_SYSCALL`，远端发布尚未成功。

## 视频时间证据基线

公开固定来源为 `https://www.bilibili.com/video/BV1PReT6EEqR/`。B 固定原件目录为 `C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/.verification/video-source-third`。媒体原件的正文为 32908 bytes，`provider=whisper.cpp`、`error=null`、`segments=null`；CLI 原件为 `strategy=html`、`textProvided=false`、`transcriptSource=null`。已有生产导入与重启正文验证 PASS，时间 segments/spans 为 0，时间验收未通过。

实际安装的 summarize-core 0.25.1 的 `transcription/whisper/whisper-cpp.js` 使用 `--no-timestamps` 和 `--output-txt`，返回 text/provider/error/notes，缺少 segments。`content/transcript/providers/youtube/yt-dlp.js` 将缺少的 segments 保留为 null。Go 已读取返回的 segments。

模块已批准在授权桥接文件内复用上游 `acquireYtDlpMedia`、cleanup 和 `spawnTracked`，读取同一本机 whisper-cli 的原生 JSON 时间输出。官方 [whisper.cpp 1.8.3 CLI](https://github.com/ggml-org/whisper.cpp/blob/v1.8.3/examples/cli/cli.cpp)定义 `--output-json`，JSON 的 offsets.from/to 为毫秒。实现及真实时间验证当前 NOT_RUN，真实同源两个变化版本当前 NOT_RUN；不制造正文变化或时间。

原始转写、原模型输出、凭据和测试原件仅保留于本机忽略目录及对象存储，不进入 Git、HTTP 或 release 包。正式公开响应沿用现有派生 SourceSpans 契约。
