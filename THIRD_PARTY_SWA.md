# SWA 无界面运行时来源与许可

产品运行时的源码来源为 `stablyai/orca` 的 `v1.4.230`，准确提交 `b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`。来源副本位于 `third_party/orca-v1.4.230/`，保留原始 [MIT 许可](third_party/orca-v1.4.230/LICENSE)及 `Copyright (c) 2026 Lovecast`；产品接入修改沿本仓库普通 Git 历史记录。

运行时依赖在 `runtime/swa-orca/package.json` 显式声明，由根 `pnpm-lock.yaml` 固定实际依赖图。上游 native 补丁来自同一准确来源的 `config/patches/`，通过工作区 `patchedDependencies` 应用。当前没有生成或验收安装包，安装包的实际依赖闭包、原生二进制来源及全部许可清单仍需在打包阶段逐项保留和检查。

已读取的实际安装元数据：

| 组件 | 版本 | 许可 | 使用位置 |
| --- | --- | --- | --- |
| `@earendil-works/pi-coding-agent` | 1.0.1 | MIT | 原生 ModelRuntime、AgentSession、资源加载、真实会话及 RPC |
| `jose` | 6.2.3 | MIT | 只读 OAuth 来源的 JWT 有效期及账户字段检查 |
| `node-pty` | 1.1.0 | MIT | 原生 daemon PTY |
| `@vscode/windows-process-tree` | 0.8.0 | MIT | Windows 进程及创建时间观察，需验证实际补丁二进制 |

产品自有适配代码位于 `runtime/swa-orca/`、`internal/adapters/orca/`、Workspace App及对应组合入口。Codex OAuth 来源仅以明确路径只读访问，不进入源码、依赖库、安装包、日志或许可证清单。Node、Pi 及 native 的实际装载、第三方许可随包交付和安装后验证尚未完成；本文件不表示安装包已通过验收。
