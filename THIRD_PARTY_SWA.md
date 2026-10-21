# SWA 随包来源

SWA 产品代码沿仓库根 LICENSE。Orca 来源为 `stablyai/orca v1.4.230`，准确 Commit `b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`，MIT 许可完整保留于 `third_party/orca-v1.4.230/LICENSE`，安装候选另附 `ORCA_LICENSE`。来源子树为该 Git archive 的导出视图及产品适配，包含上游 export-ignore 省略和 Windows 文件执行位变化；具体边界及产品适配见任务报告。

原始 `Copyright (c) 2026 Lovecast` 保留在来源 LICENSE 中。已安装元数据确认 Pi1.0.1、jose6.2.3、node-pty1.1.0及windows-process-tree0.8.0均为MIT；原始MIT许可全文和各包notices继续按实际包内容随候选保留。各运行依赖的必要消费者沿实际源码和构建metafile核查，依赖声明或包名不能代替完整闭包验证。

产品复用 Orca 原生 runtime、持久 SQLite、Run/Task/Dispatch、请求 receipt、会话 transcript/journal、PTY/daemon 和已有进程控制。产品自己的组装入口、原生CLI账号/启动政策与授权适配位于 `runtime/swa-orca`，Go CLI/App/HTTP 与关联存储沿现有目录。安装候选的 VERSION 和 `swa --version` 指向实际构建源码 Commit；版本标识不代替功能验收。

随包 Node 为上游固定的官方 Windows x64 `24.21.0`。发行输入及真实 node.exe 均核上游固定 SHA256，完整 Node LICENSE 位于安装包 `node/LICENSE`。native 模块使用该版本的 headers/node.lib 编译；没有使用本机 Orca 的 Node、native 二进制或可写状态。

本阶段Worker使用实际安装的 `@earendil-works/pi-coding-agent 1.0.1` 原生CLI和其真实provider实现，Agent及provider/model必须明确选择。已审政策限定既有原生API-key来源及三类已核API，原auth/models只读，所选有效材料在私有DataRoot中按受保护ACL独占创建；原账号不会进入安装包。固定RPC/关闭工具参数、原生状态、来源及当前批准在效果前核查；认证、工具关闭及实际模型仍须按当前测试计划获得真实验证。

运行依赖和构建工具的准确输入记录在 `runtime/swa-orca/package.json`、`pnpm-workspace.yaml`、`pnpm-lock.yaml` 及 `dependencies.lock.json`。pnpm deploy 按锁导出包含 workspace 的独立依赖闭包，安装复制实际内容；每个第三方包的许可和 notices 随其原始 package 内容保留。node-pty、windows-process-tree 及Windows安全模块的相关补丁属于获准来源适配，Worker及服务Job保护须使用该候选实际编译的native能力。node-gyp12.3.0和TypeScript6.0.3为构建/检查工具；Electron43.7.8与新增@types包为开发类型输入，安装使用ignore-scripts且prod deploy保持原规则。真实声明、锁与安装已经同步。

构建从实际 esbuild 内联输入所属包复制原始许可全文至安装根目录 `third-party-licenses/<包名>@<版本>/`，作用域包名中的 `/` 使用 `__`。该目录保留 zod4.6.5、yaml2.9.1、diff8.0.4、smol-toml1.9.1、jsonc-parser3.3.1、@streamparser/json0.0.26、tweetnacl1.0.3与0.14.5、proper-lockfile4.1.2 的完整 LICENSE/LICENCE/COPYING/NOTICE；文件名及原版权按实际包内容保留。八项直接依赖作为构建输入安装，间接生产依赖由 pnpm 的实际生产图保留。

开发类型输入另含真实@types/express5.0.6，运行资源补充上游同版emojibase-data17.0.0；准确声明及锁同步保留。TypeScript6.0.3严格完整检查已通过，四个native模块已在固定Node24.21.0下真实编译及装载；完整候选仍须核相同源码及安装闭包。

安装使用明确的新目录及私有当前用户 ACL，不修改 PATH、注册表、全局 Agent 配置或本机 Orca 安装。后台和会话数据由启动时明确的独立 data-dir 拥有；认证来源、受保护目录和真实当前项目授权分别核查。

准确1209完整构建已完成native、Go及VERSION，候选内闭包和装载通过；独立安装后的依赖解析失败原件保留。后继采用成熟pnpm hoisted部署生成普通目录及文件，安装复制前拒绝全部reparse point；实际局部搬迁和动态装载已通过。完整安装、原生CLI认证及功能验收继续按任务计划执行，当前没有完整安装验收或功能通过结论。

后继来源适配将实际服务Job归属从daemon入口经startDaemon完整传递给DaemonServer，并在原SQLite适配器打开/备份边界使用Node成熟toNamespacedPath。原生身份、PID record及内核成员的严格核验继续保留；这些修复的实际限定结果见任务报告§31。
