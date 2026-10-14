# 公开参考代码来源与分析记录

获取开始于 2026-10-14，首轮交付日期为 2026-10-15（Asia/Shanghai）。分析基线为 Astrocyte `1c026c3e7f9e91c53d1dced6479099668c9162cb`。本次只获取、静态分析公开代码；没有启动上游服务、自治科研实验或模型请求。

## 官方来源与精确版本

| 项目 | 官方代码与当前获取 SHA | 许可证 | 本地副本 |
| --- | --- | --- | --- |
| OpenCSGs/csghub | [官方仓库](https://github.com/OpenCSGs/csghub)，`bbb9bf4b346a199a387090de76fcd6c1fffe98f6` | [根 LICENSE：Apache-2.0](https://github.com/OpenCSGs/csghub/blob/bbb9bf4b346a199a387090de76fcd6c1fffe98f6/LICENSE#L1) | `.verification/reference-code-20261014/csghub-git/` |
| OpenCSGs/csghub-server | [官方仓库](https://github.com/OpenCSGs/csghub-server)，`a67b3024a9553aecf3379c9f6bccfcdada2034f5` | [根 LICENSE：Apache-2.0](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/LICENSE#L1) | `.verification/reference-code-20261014/csghub-server/` |

两者属于 OpenCSGs 官方组织；[服务端 README 第 3 行](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/README.md#L3)直接声明其为 CSGHub 的服务端组成部分。上述 SHA 是分别获取的两个仓库当前版本，官方兼容发布组合未知。许可证核实对象为各仓库根文件；没有逐一核查全部第三方依赖或资产许可证。

全部第三方代码、REA 输出和临时脚本位于本工作树已忽略目录；不会提交到 Astrocyte。正式写入范围为 `docs/research/reference-code-20261014/**`。

## 实际获取与错误记录

服务端 `git clone --depth 1 https://github.com/OpenCSGs/csghub-server.git` 成功。前端首次普通 clone 失败，退出码 1，错误为 `OpenSSL SSL_connect: SSL_ERROR_SYSCALL in connection to github.com:443`；后续 PowerShell GitHub API 请求遇到连接意外 EOF，未取得归档。最终使用按命令设置 `git -c http.version=HTTP/1.1 clone --depth 1 https://github.com/OpenCSGs/csghub.git` 成功。TLS 验证保持启用，未更改 Git 全局配置或认证配置。

两份本地副本 `git rev-parse HEAD` 与表中 SHA 一致，`git status --short` 均为空。后续所有代码引用使用 SHA 固定链接。

## REA 6.5.0 的实际证据

客户端为现有 Codex 会话 `01a13ae1-1399-78d2-b329-0ea4fc5eb255`，Orca 终端 `term_79eeba93-c810-4a04-a1a0-e27ce0767ab5`；终端实际显示模型 `GPT-6.1-Sol high fast`。使用 `reverse-engineer-anything` 第 42 版技能和已安装 REA 6.5.0 CLI。Codex 的 REA MCP 保持关闭。

输入为前端源码目录，指定 `package.json`、`vite.config.js`、`src/main.js`、`src/packs/useFetchApi.js`、`src/packs/csrfFetch.js`、`src/packs/jwtFetch.js`、`src/stores/RepoDetailStore.js`、`src/packs/contextMessage.js` 八个文件。命令为：

```powershell
rea analyze-javascript-application <工作树>/.verification/reference-code-20261014/csghub-git/frontend --source-path package.json --source-path vite.config.js --source-path src/main.js --source-path src/packs/useFetchApi.js --source-path src/packs/csrfFetch.js --source-path src/packs/jwtFetch.js --source-path src/stores/RepoDetailStore.js --source-path src/packs/contextMessage.js --json
rea trace-application-feature <工作树>/.verification/reference-code-20261014/csghub-fetch-trace-input.json --json
```

提取成功，退出码 0：清点相关文件 273 个，指定八个文件包含七个 JavaScript 文件，解析七个 JavaScript 文件，解析失败 0，读取文本 32,529 bytes，访问 AST 节点 3,539 个，findings 270 个。`statistics.modules=0` 是原输出统计值，不能解释为不存在模块；实际 graph 含 `javascript-module` 节点。

- 主证据：`ev_4b9c27fd9c9f3ae4cde0b4b634bea21bc5912a4f7d1641eb826b1ebfd7ea51f9`，文件 `csghub-javascript-evidence.json`。
- 调用关系追踪：`ev_2a7888994782ec8c9a7a4ae9a1052277738959058d9a9db02c07a066991c608d`，文件 `csghub-fetch-trace.json`。种子为 `useFetchApi.js` 的真实 module 节点 `jag_node_ae74e21068a78551b4c773133314d8c68ac86897ce4a8374b0fde005ca07f436`，方向 outgoing，匹配一个节点。
- 两项覆盖均为 `partial`：265 个清点文件未选入；未选择的 Vue 文件、外部依赖、动态绑定和真实服务运行关系保持未知。
- REA semantic graph 实际记录 `useFetchApi.js:79` 对 `popupReloginDialog` 的本地 `calls/resolved`，以及第 54、59 行对 `acceptLanguage` 的本地调用。它们是静态语法解析结果；没有运行调用或 HTTP 验证。

本次没有获取生产发布包。源码树的 package/入口/import/调用提取不能代表官方发布包结构。Go 的接口、事务和状态转移来自常规源码阅读，均标为源码证据；没有把 Go 阅读结果写成 REA 证据。

## autoscientist 待确认

用户准确仓库尚未确认。本轮仅保留名称相符的官方候选 [Adaption AutoScientist 产品页](https://adaptionlabs.ai/auto-scientist)和 [Adaption 官方文档](https://docs.adaptionlabs.ai/autoscientist/overview/)。官方文档描述数据优化、训练、评估和下一轮参数调整的服务流程。该候选的核心公开实现仓库、精确 SHA 和代码许可证尚未核实，均为未知。没有替用户选择项目，没有克隆候选、运行其服务或推断内部实现。

## 交付顺序

第一阶段：[资产、权限与版本代码索引](CSGHUB-ASSETS.md)，Commit `85c7fcf87e4033a5826fd5af10a36cf2c38f5ee5` 已普通 push 并向 M-product、ROOT 发送 Handoff。第二阶段：[Agent 状态、恢复、研究步骤、评审与经验存储](CSGHUB-RUNTIME-RESEARCH.md)以及 [可执行 Handoff](HANDOFF.md)。待确认仓库不会阻塞 CSGHub 交付；开发范围由 M-product 和现有任务负责人决定。
