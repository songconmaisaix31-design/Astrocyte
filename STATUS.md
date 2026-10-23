# 当前状态

当前任务是[产品拆分与完整任务交付](tasks/swa-evolution-20261023.md)。用户已确认 Astrocyte、Attention Mirror 拆为两个独立项目/仓库，SWA 只保留必要集成；本轮实际并发目标为 N=4、N=6。开发分支 `integration/swa-task-cards`，最近已推送提交 `644bbe8701ae1b1c30b54f18621c28e03dbf6082` 包含已审查的 SWA 后端、CLI 与原生生命周期。前端、旧源移出和默认检查正在共享工作树收口。

## 当前实现与验证

| 范围 | 已取得的结果 | 剩余验收 |
| --- | --- | --- |
| Workspace 与 SWA 服务 | 唯一空间存储责任、原表/ID/资料引用/同连接事务保留；独立 swa-server、swahttp/swastore/projectfiles 闭包通过；真实 SQLite/HTTP、原生产服务生成的公开旧库兼容、开发进程生命周期及 Windows 交编通过，源码限定接受 | 同提交完整安装 |
| Astrocyte | 独立本地仓库 `/workspace/astrocyte`，`426723b5525ec36b1e0c93dd4b6d9eb6e42287f1`；认知 MCP/sleep、快照/只读 HTTP、GUI 和显式单文件事件导入均已独立审查。Go 测试、Linux 构建、真实 MCP→快照→七 API→浏览器 DOM/原库不变/正常退出通过；真实 Git→实际导入二进制→SQLite 幂等与冲突回滚通过；Windows x64 交编通过 | 自动来源和外部完成证据仍为显式集成边界；Windows 原生运行未验 |
| Attention Mirror | 独立仓库 `/workspace/attention-mirror`，`58bb6e4454e03db135cabe6de2dec8dfc481aee6`；独立资料服务/GUI、显式提取器配置、资产根限制及原适配器测试已独立审查。真实公开原件导入/重放/人工沉淀/Taste/撤回保全、Go 检查、前端 67 项规则、DOM 9 项规则、实际安装启动通过；Windows 交编通过 | 公开 ACL URL 实测因 DNS 解析失败；模型、PDF/媒体、OpenViking 当前空间权限集成与 Windows 原生仍未验 |
| run/wait/export | CLI 与 GUI 接同一 Run；稳定请求 ID、递归 strict、分离退出码、完整导出已实现；CLI 局部测试、Web 类型检查/ESLint、观察规则 9 项、真实无原生运行时 DOM 2 项通过 | 完整真实 Run、断线恢复、实际递归失败和模型结果 |
| Go 组合提交 | 原 controller/task/worker 三阶段、输入完整内容绑定、已完成任务关联恢复；真实文件同大小同时间变更拒绝测试通过 | 本轮安装中的原生幂等恢复 |
| 能力与角色 | 实际观察层级、送达/恢复状态、角色/讨论界面已接入；能力、角色和实际 SQLite CRUD/CAS 规则通过，源码接受 | 实际模型及账号归属、父任务恢复 |
| 原生 Run 生命周期 | 原 Task/Dispatch、journal、成果和原资源释放路径已实现并获独立源码接受；完成邮件仅经明确 worker.send 恢复，notify 只存记录，成果保持 unreviewed。严格 TS 与实际 SQLite 规则检查通过 | 实际子成果唤醒父任务、完整资源回收、最终递归聚合 |
| N4/N6 | 同 Run/角色复用/实际孙任务/父恢复/停止隔离/自动释放验收驱动已实现并通过独立源码审查，固定公开输入局部检查通过 | 真实 StepFun 并发、停止隔离、继续输出和最终所属进程释放尚未运行 |

以上局部检查针对当时源码，不能替代最终集成 Commit 的安装验收。旧混合 Go/SQL 与配套资源已按独立审查清单移出，公开 testdata 和研究数据保留；完整保存点与归档见 [LEGACY](docs/LEGACY.md)。清理后的 `pnpm check`、`pnpm check:race`、`pnpm build` 与实际默认启动检查均退出0，同产物页面/JS/CSS、独立 SQLite、健康/会话/项目及正常退出通过。Go 依赖清理后双平台含测试图、17包架构、空 tidy 差分与模块校验通过。尚未启动本轮完整安装候选构建。实际失败及最小复现、根因和局部复验保存在各任务文件。

## 保留的安装基线

- Linux 组合安装来源 `aefe6c4be30ec3ba0447fed9163eeff5a0fb04ad`，根为 `/workspace/swa-linux-installs/swa-linux-combined-aefe6c4b-pi-01`。构建、归档、严格安装、装内 D12 与组合服务正常 EOF 检查通过。旧 held 服务已经正常关闭，serve/observer 均退出 0，最终两次完整所属进程观察为空。
- Windows 标准 CLI 来源 `1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da`，CI run `39858402174` / job `120102466532`，构建、ZIP、严格安装、readiness-eof 服务及正常退出通过。该结果不包含真实模型或 N4/N6，也不代验本轮源码。
- 两平台基线来源不同。过期 N=3 执行包永久失效，不得执行或借用旧身份。
- 已清理四份可从保留归档恢复的旧解包副本，实际释放 3,967,369,216 字节；原归档、SHA、独立安装、数据库、账号与证据均保留。记录见 `.verification/cloud-setup/release-cleanup-20261023-01/operation-01/`。

## 当前边界与下一步

当前仅连接 Linux 环境；Windows x64 可用原生 CI，但真实模型验收尚无连接环境。真实模型继续使用已配置 StepFun 来源、固定公开四文件项目及 ToolsOff；密钥不进入仓库、日志或导出。此次明确的 N>3 请求授权其必要验证，不逐 SHA 重复请求同范围批准。

正式科研实验的原证据根与权威锁尚未恢复，本轮未运行依赖这些原件的实验；既有科研与用户数据保全。

完成 Windows 组合包验证入口，固定已审查的集成来源，再构建新候选、运行 Linux N4/N6、同 Run GUI 观察和 Windows 适用检查。正式发布、远端新仓库创建和 main 合并尚未执行。云环境配置草稿 revision 9 已保存并读回确认，安装选择器与启动说明已同步；仍需在环境设置保存并发布才生效。草稿仓库成员仍只有 SWA，两份没有远端的独立本地仓库在新任务中的恢复尚未验证。

历史阶段原件保留在[准确旧状态](https://github.com/songconmaisaix31-design/swarm-anything/blob/11619d72f08c163f159a4f4befe270806b1b67ea/STATUS.md)，其过期服务、请求和待答说明不作为当前执行条件。
