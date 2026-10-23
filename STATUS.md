# 当前状态

当前任务是[产品拆分与完整任务交付](tasks/swa-evolution-20261023.md)。用户已确认 Astrocyte、Attention Mirror 拆为两个独立项目/仓库，SWA 只保留必要集成；本轮实际并发目标为 N=4、N=6。开发分支 `integration/swa-task-cards`。清理、前后端及默认检查已集成；安装候选来源为 `50d099ddf9787a283573a477fb5f29017aef6c89`。后继 `214ab146c33680504c39216d7ad41863c834177c` 只修正外部验收工具的空间回执校验，不改变该安装物。

## 当前实现与验证

| 范围 | 已取得的结果 | 剩余验收 |
| --- | --- | --- |
| Workspace 与 SWA 服务 | 唯一空间存储责任、原表/ID/资料引用/同连接事务保留；独立 swa-server、swahttp/swastore/projectfiles 闭包通过；真实 SQLite/HTTP、原生产服务生成的公开旧库兼容、开发进程生命周期及 Windows 交编通过，源码限定接受 | 后继控制器修复来源的完整安装 |
| Astrocyte | 独立本地仓库 `/workspace/astrocyte`，`426723b5525ec36b1e0c93dd4b6d9eb6e42287f1`；认知 MCP/sleep、快照/只读 HTTP、GUI 和显式单文件事件导入均已独立审查。Go 测试、Linux 构建、真实 MCP→快照→七 API→浏览器 DOM/原库不变/正常退出通过；真实 Git→实际导入二进制→SQLite 幂等与冲突回滚通过；Windows x64 交编通过 | 自动来源和外部完成证据仍为显式集成边界；Windows 原生运行未验 |
| Attention Mirror | 独立仓库 `/workspace/attention-mirror`，`58bb6e4454e03db135cabe6de2dec8dfc481aee6`；独立资料服务/GUI、显式提取器配置、资产根限制及原适配器测试已独立审查。真实公开原件导入/重放/人工沉淀/Taste/撤回保全、Go 检查、前端 67 项规则、DOM 9 项规则、实际安装启动通过；Windows 交编通过 | 公开 ACL URL 实测因 DNS 解析失败；模型、PDF/媒体、OpenViking 当前空间权限集成与 Windows 原生仍未验 |
| run/wait/export | CLI 与 GUI 接同一 Run；稳定请求 ID、递归 strict、分离退出码、完整导出已实现；CLI 局部测试、Web 类型检查/ESLint、观察规则 9 项、真实无原生运行时 DOM 2 项通过 | 完整真实 Run、断线恢复、实际递归失败和模型结果 |
| Go 组合提交 | 原 controller/task/worker 三阶段、输入完整内容绑定、已完成任务关联恢复；真实文件同大小同时间变更拒绝测试通过 | 本轮安装中的原生幂等恢复 |
| 能力与角色 | 实际观察层级、送达/恢复状态、角色/讨论界面已接入；能力、角色和实际 SQLite CRUD/CAS 规则通过，源码接受 | 实际模型及账号归属、父任务恢复 |
| 原生 Run 生命周期 | 原 Task/Dispatch、journal、成果和原资源释放路径已实现并获独立源码接受；完成邮件仅经明确 worker.send 恢复，notify 只存记录，成果保持 unreviewed。严格 TS 与实际 SQLite 规则检查通过 | 实际子成果唤醒父任务、完整资源回收、最终递归聚合 |
| N4/N6 | 同 Run/角色复用/实际孙任务/父恢复/停止隔离/自动释放驱动已审查；实际 N4 已通过来源、授权和原生依赖检查，随后在控制器创建时拒绝 Linux shellOverride，未启动 Worker | 修复并局部验证真实控制器，再运行新安装的 N4/N6；并发、停止隔离、继续输出与完整业务结算未通过 |

旧混合 Go/SQL 与配套资源已按独立审查清单移出，公开 testdata 和研究数据保留；完整保存点与归档见 [LEGACY](docs/LEGACY.md)。清理后的 `pnpm check`、`pnpm check:race`、`pnpm build` 与实际默认启动检查均退出0，同产物页面/JS/CSS、独立 SQLite、健康/会话/项目及正常退出通过。Go 依赖清理后双平台含测试图、17包架构、空 tidy 差分与模块校验通过。

S50 Linux 组合包构建、归档、解包、严格安装及装内 D12（12通过、0失败、0跳过）均通过，安装根为 `/workspace/swa-linux-installs/swa-linux-combined-50d099dd-n46-01`。首次 N4 在空间 POST 201 后被工具误套 UUID 条件拒绝；原数据库只有该空间，项目和 Grant 均为零，原服务正常退出。空间 ID 在拆分前就为不透明字符串；工具修复的真实 SQLite/HTTP 创建、查询和项目登记局部测试通过，原请求没有重放。

第二次 N4 使用新的独立服务。原 Run 请求 `52573151-edbc-409c-9210-21efdad053f0` 在 `pty-spawn` 阶段停为 pending：SWA 将 `/bin/bash` 传给仅支持 Windows 的 `shellOverride`，上游明确拒绝创建终端。项目 Grant 撤销已确认，无 Worker 成果。关闭时服务因未决原生请求退出6；观察器退出0，连续两次观察所属进程为空。资源退出与业务 UNKNOWN 分别保留，不能将该次判为成功。准确原件在 `/workspace/swa-pi-n4-packet-50d099dd-20261023-02/evidence` 和对应 `swa-pi-owned-50d099dd-n4-20261023-02/evidence`。

同 S50 Windows run `39900064207` 在 build 阶段退出1。两次只读诊断 `39901968797`、`39902560190` 均完成，固定原产物证明冻结安装、Go 下载与校验均退出0；原冻结安装日志仍筛选旧工作区名称并报告 `No projects matched the filters "astrocyte, @astrocyte/web"`，Web 命令随后因 `'tsc' is not recognized as an internal or external command` 失败。修复 `67c3d63de12ce87f65ea14b429a8845116f5a288` 仅改为根目录与 Web 路径选择器；固定 pnpm 实际选择从2恢复为4个工作区，固定 Node 下实际 Web 工具版本和 PowerShell Parser 检查均通过，已独立审查，未重试 Windows 完整构建。

Linux 控制器改为省略仅适用 Windows 的 shell override。严格 TypeScript 和真实本地 PTY→原 helper 两次同 PID 识别→正常 EOF/进程消失的局部检查通过，细节见[修复记录](tasks/swa-controller-linux-startup-20261023.md)。该局部检查不代替完整 RPC、Run 或模型验收。每次失败均先交代最小复现、根因证据和局部验证，再启动后继完整构建。

两份独立仓库另有可恢复完整 main 历史的 `/workspace/astrocyte-426723b5.bundle` 和 `/workspace/attention-mirror-58bb6e4.bundle`，均通过 `git bundle verify`。它们不包含 ignored 数据或账号；远端仓库尚未创建。

## 保留的安装基线

- Linux 组合安装来源 `aefe6c4be30ec3ba0447fed9163eeff5a0fb04ad`，根为 `/workspace/swa-linux-installs/swa-linux-combined-aefe6c4b-pi-01`。构建、归档、严格安装、装内 D12 与组合服务正常 EOF 检查通过。旧 held 服务已经正常关闭，serve/observer 均退出 0，最终两次完整所属进程观察为空。
- Windows 标准 CLI 来源 `1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da`，CI run `39858402174` / job `120102466532`，构建、ZIP、严格安装、readiness-eof 服务及正常退出通过。该结果不包含真实模型或 N4/N6，也不代验本轮源码。
- 两平台基线来源不同。过期 N=3 执行包永久失效，不得执行或借用旧身份。
- 已清理四份可从保留归档恢复的旧解包副本，实际释放 3,967,369,216 字节；原归档、SHA、独立安装、数据库、账号与证据均保留。记录见 `.verification/cloud-setup/release-cleanup-20261023-01/operation-01/`。
- 本轮另核对并清理 `559c621b-02`、`f71eec93-03`、`aefe6c4b-pi-01`、`50d099dd-n46-01` 的四份 packages 解包副本，实际可用空间增加 3,957,583,872 字节。完整成员内容与硬链接核验、当前用户进程引用核查、删除后原归档校验通过；dist 候选、独立安装及所有数据和证据保留。系统 Docker/containerd 进程资源不可读的范围单独保留，未宣称核查所有系统进程。原件为 `.verification/cloud-setup/release-cleanup-n46-20261023-02/operation-01/`。

## 当前边界与下一步

当前仅连接 Linux 环境；Windows x64 可用原生 CI，但真实模型验收尚无连接环境。真实模型继续使用已配置 StepFun 来源、固定公开四文件项目及 ToolsOff；密钥不进入仓库、日志或导出。此次明确的 N>3 请求授权其必要验证，不逐 SHA 重复请求同范围批准。

正式科研实验的原证据根与权威锁尚未恢复，本轮未运行依赖这些原件的实验；既有科研与用户数据保全。

Linux 控制器与 Windows 安装选择器的修复、局部验证和独立审查均已完成。下一步固定整合后的来源，构建新候选，继续 N4/N6、同 Run GUI 和 Windows 适用检查。正式发布、远端新仓库创建和 main 合并尚未执行。云环境配置草稿 revision 9 已保存并读回确认，安装选择器与启动说明已同步；仍需在环境设置保存并发布才生效。草稿仓库成员仍只有 SWA，两份没有远端的独立本地仓库在新任务中的恢复尚未验证。

历史阶段原件保留在[准确旧状态](https://github.com/songconmaisaix31-design/swarm-anything/blob/11619d72f08c163f159a4f4befe270806b1b67ea/STATUS.md)，其过期服务、请求和待答说明不作为当前执行条件。
