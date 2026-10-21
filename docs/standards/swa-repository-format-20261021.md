# Swarm CLI 的 REA 仓库格式参考

用户要求“继续构建swarm cli，格式上用rea的仓库格式”。参考对象为官方 [morluto/rea](https://github.com/morluto/rea)，2026-10-21 读取 main 后通过 GitHub API 取得源码标识 `70835d3d5f030deadf606e1c79c70a60883e50c9`。这是仓库组织参考，本机已安装 REA 的版本及既有 Orca 逆向证据沿[原记录](swa-orca-cli-reference-20261021.md)保留。

## 已读取的来源

- [README](https://github.com/morluto/rea/blob/70835d3d5f030deadf606e1c79c70a60883e50c9/README.md)：产品用途、快速开始、终端使用、文档及贡献入口。
- [架构图](https://github.com/morluto/rea/blob/70835d3d5f030deadf606e1c79c70a60883e50c9/docs/architecture.mermaid)：CLI 入口、组装、应用、契约、领域逻辑及平台能力的职责。
- [package.json](https://github.com/morluto/rea/blob/70835d3d5f030deadf606e1c79c70a60883e50c9/package.json)：命令入口、发布文件清单及分开的构建、类型、包检查命令。
- [贡献指南](https://github.com/morluto/rea/blob/70835d3d5f030deadf606e1c79c70a60883e50c9/CONTRIBUTING.md)：开发环境、生成文件归属、实际验证与发布过程。

## 与当前项目的对应关系

下表描述已有职责及当前实现路径。CLI 实现位于 `swa-cli-headless-20261020` 分支，尚未集成到 main，静态源码不代表安装可用。

| 职责 | REA 的组织位置 | Swarm 当前实现位置 |
| --- | --- | --- |
| 用户入口 | 根 README、CLI launcher | 根 README、`cmd/swa` |
| 运行组装与原生适配 | `src/composition`、各 provider | `runtime/swa-orca/src`、`internal/adapters/orca` |
| 应用与领域逻辑 | `src/application`、`src/domain` | `internal/workspace/app`、既有领域模块 |
| 输入输出契约 | `src/contracts` | `contracts`、runtime 契约及已有 Go 接口 |
| 平台能力与第三方来源 | `native/windows`、`third_party`、许可文件 | `third_party/orca-v1.4.230`、随包来源与许可 |
| 构建与交付 | `scripts`、显式发布文件清单 | `scripts/build-swa-*`、安装与闭包检查入口 |
| 验证和说明 | `tests`、`docs`、CONTRIBUTING | `tests/swa-cli-headless`、`docs`、CONTRIBUTING |

## 待确认范围与实际限制

已向用户询问“仓库格式”是否仅覆盖目录、README 与开发文档，或者同时覆盖 CLI 命令/输出及构建、安装、发布包结构。尚未收到答案；公开命令、协议、发布渠道和模块迁移保持未决定，已确认的首版构建继续。

REA 使用的 npm 分发、MCP、网站及测试方法各有自身用途。Swarm 的当前技术、权限与验收以本项目规格和用户确认为准，仓库参考没有改变本机 Orca 隔离、真实 Worker 验证或旧 S3 的独立记录。

当前八次完整构建均未形成可安装候选。新开发类型声明的真实锁定/冻结安装和原契约生成已通过；严格类型原1GiB OOM失败保留，用户现已解除资源限制，后继使用8GiB集中取得完整诊断。准确`ada4089de6a9fcaf72f3a91c0b0331ced11343d9`的第八次构建已实际生成短目录分析XML，越过C1258后因FTK1011无法写链接跟踪日志而失败。原Worker集中处理全部中间路径及四个原生模块，复用原校验输入分模块复编，再最终完整打包；精简依赖的实际范围、安装和同一安装候选的双Worker验收仍未通过。当前服务与执行安排以 [STATUS](../../STATUS.md) 为准。
