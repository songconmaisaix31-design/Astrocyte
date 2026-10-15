# V4 首日恢复（2026-10-14）

用户授权：重新启动采集，检查自动采集未执行的原因。

## 写域与流程

- 主控：自动任务设置核对、独立只读验收、根 STATUS/HANDOFF 与本执行单。正式库和实验协议不由主控改写。
- 既有 automation 的日运行 Agent：独占本次正式库运行及外部 capture/day 产物；使用原生产 CLI 和锁定工具链，不改仓库业务代码或协议。
- 两轨并行：日运行 Agent 执行 10-13 闭日流程；主控诊断 Orca 启动与派发路径。沿用既有工作树和日程，不新增开发分支或调度器。

## 保持项

原自动任务 `f5e3ee7a-ddfe-4abf-b81c-b55a042ab27b`、Asia/Shanghai、`20 0 14-23 10 *`、existing workspace、reuse-session、正式根 `v4-w1/formal-20261013-first` 保持。day0=2026-10-13，锁定全文 SHA 为 `5aa238599b6dd59e28d9ca30f71b6b9d547822b46d8dadc477092a46c09ab1b7`。

不 init、不重锁、不改引擎/schema/参数/权重/题库/模型/预算/评分，不覆盖原 skip/失败/unknown，不重复已送达请求。只在当前合法昨日窗口运行。人评不代填，预测等待真实 D+3 闭日。

## 首次失败与恢复

原 run `df055da3-8b4c-4e02-8470-e53d176a59ce` 原定 00:20，01:18:12 被补调度，状态 `skipped_unavailable`，未派发；错误为 `The target project is no longer available.`。

原 Orca daemon 01:18:09.482 启动、01:18:09.829 ready。安装版 1.4.220 的 `qq` 从界面 state.repos 查仓库，`Qq` 查不到即登记上述 skip；`$q` 注册派发后直接调用 rendererReady，没有等待项目加载的门禁。当前同一仓库/工作树可读，同一目标手动派发成功。错误来自启动补调度时的项目可用性判定；仓库没有删除。

顶层仓库 UUID 与 runContext 中 GitHub project ID 属不同兼容字段，不能把两者不同直接判成绑定失效。公开 CLI 修复探查的失败与临时模式切换记录保留；最终恢复原 existing/reuse/enabled 设置。原提示词、日程与目标在恢复后逐项核对。

手动恢复 run `59891a05-a832-4c13-8436-9b597726da8d`，03:01:12 创建，随后实际派发。验收查正式 capture/Sleep/不可变日快照、十二臂终态、独立 judge、B 全量装载和原协议 SHA；派发本身不计验收通过。

## 采集源现状

固定授权 `C:/Users/DW/.codex/sessions` 共 1714 文件，最新修改 10-10 21:56；`C:/Users/DW/.claude/projects` 共 7 文件，最新修改 10-10 15:12。两目录 10-13 修改文件数均为 0；10-13 新增认知信号全部来自三仓 git 中的 Astrocyte/McPet。此记录不授权扩展源或修改锁定协议。

## 外部证据

设置和恢复回执：`root-v4-start/automation-recovery-20261014-*-local.json`。日采集：`v4-w1/formal-20261013-first/capture-20261013-first/`。实验日：`v4-w2/day-20261013-first/`。raw、答卷和用户评分表只保留本机。

后续 00:20 真正定时运行仍待观察；机器与 Orca 应在日程前运行。未修补 Orca 安装包中的启动门禁，不以手动恢复代替后续定时验收。

## 实际验收（03:13）

手动run completed。新增84 commits（Astro65/McPet19）、1生产Sleep，capture_complete=true；失败/重复/late/source_errors均0。4题12臂OK、6事实评分、1独立judge完成，人评待填，首个预测10-17闭日成熟。B四次各完整1003消息，A四次聚合误差0/raw0/1280引用可溯；所有请求实际usage、输入和1800秒时限均通过。主控独立SQL确认84条s=1/w=.6、正式库/快照四表标量全等，V1R与协议原SHA未变。新增业务测试0；Git只收根交接及本执行单，完整派生证据保留外部原目录。

运行生成未跟踪 `scripts/validation/__pycache__/diagnose_native.cpython-313.pyc`。Worker首清理被自动审批审查拒绝；主控改为单文件路径/类型/长度/创建时间核查后清理，仍被审查以 `blocked by policy` 拒绝。文件保留，不提交，不再尝试其他删除路径。
