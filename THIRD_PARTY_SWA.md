# SWA 随包来源

SWA 产品代码沿仓库根 LICENSE。Orca 来源为 `stablyai/orca v1.4.230`，准确 Commit `b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`，MIT 许可完整保留于 `third_party/orca-v1.4.230/LICENSE`，安装候选另附 `ORCA_LICENSE`。来源子树为该 Git archive 的导出视图及产品适配，包含上游 export-ignore 省略和 Windows 文件执行位变化；具体边界及产品适配见任务报告。

原始 `Copyright (c) 2026 Lovecast` 保留在来源 LICENSE 中。已安装元数据确认 Pi SDK1.0.1、jose6.2.3、node-pty1.1.0及windows-process-tree0.8.0均为MIT；其使用位置分别为真实SDK会话/RPC、只读JWT检查、原生PTY/Job，以及实际进程/创建时间观察。原始MIT许可全文和各包notices继续按实际包内容随候选保留。

产品复用 Orca 原生 runtime、持久 SQLite、Run/Task/Dispatch、请求 receipt、会话 transcript/journal、PTY/daemon 和已有进程控制。产品自己的组装入口、受限 Pi SDK 与授权适配位于 `runtime/swa-orca`，Go CLI/App/HTTP 与关联存储沿现有目录。安装候选的 VERSION 和 `swa --version` 指向实际构建源码 Commit；版本标识不代替功能验收。

随包 Node 为上游固定的官方 Windows x64 `24.21.0`。发行输入及真实 node.exe 均核上游固定 SHA256，完整 Node LICENSE 位于安装包 `node/LICENSE`。native 模块使用该版本的 headers/node.lib 编译；没有使用本机 Orca 的 Node、native 二进制或可写状态。

Pi SDK 为 `@earendil-works/pi-coding-agent 1.0.1` 及同版本 Pi family。产品固定 `openai-codex/gpt-6.1-sol` 与 `xhigh`，使用显式已配置 OAuth 文件的只读适配；该文件不会进入安装包。工具、扩展、项目资源发现和认证刷新约束须按当前测试计划获得真实验证。

运行依赖和构建工具的准确输入记录在 `runtime/swa-orca/package.json`、`pnpm-workspace.yaml`、`pnpm-lock.yaml` 及 `dependencies.lock.json`。pnpm deploy 按锁导出包含 workspace 的独立依赖闭包，安装复制实际内容；每个第三方包的许可和 notices 随其原始 package 内容保留。node-pty、windows-process-tree 及相关上游补丁属于获准来源适配，SDK Worker 的停止保护须使用该候选实际编译的 native 能力。node-gyp12.3.0为构建工具，沿准确上游锁采用，不进入运行依赖清单。

安装使用明确的新目录及私有当前用户 ACL，不修改 PATH、注册表、全局 Agent 配置或本机 Orca 安装。后台和会话数据由启动时明确的独立 data-dir 拥有；认证来源、受保护目录和真实当前项目授权分别核查。

当前构建及安装入口尚未运行；实际安装闭包、全部第三方许可、native二进制装载、SDK认证和功能验收均须按任务计划执行后登记。
