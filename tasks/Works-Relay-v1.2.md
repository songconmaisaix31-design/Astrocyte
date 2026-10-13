# Works Relay 任务书 v1.2 定稿（Agent 接入层原型 · 能力板首切片）

本文档为 v1.0 与 v1.1 修订的合并自包含定稿，取代此前所有版本。
v1.1 已登记于 validation 分支（commit 421a076）作为评审记录保留。

来源：用户2026-10-14本轮定稿。本文为 validation 纯文档登记；立即开工授权与实施范围按 §0、§7 生效，业务只在 Workbench 线。此前 v1.0 缺失与实施窗口待答已由本定稿关闭，不重问。

## 0. 定位与体验目标

一个人三分钟体验闭环：打开寻回册 → 看见全部 AI 作品与可用 Agent →
选一件作品 → 交接给一个 Agent 接力 → 看见它开始工作 → 它的活动
出现在认知层 Today 视图。

资产底座（已存在，本任务是组装不是开发）：works 域五能力（平台
统计/卡片/文件核实/交接手册/扫描同步）、agents 域（CLI 清点/项目
发现/权限 A/B/C）、原生会话域（send/read/resume/stop）、events 域
（native 观察→事件→认知）、workbench 前端底座。

- 分支：workbench/v0.2-access 从 workbench/v0.1-workbench 拉出
  （验收基线 d46f9f4，CI 绿）
- 冻结兼容：不触 validation 分支业务路径、V4 正式库、引擎、schema、
  七类权重；快照管道与只读红线不变；本文档登记件留在 validation
  纯文档，业务实现在 workbench 线

## 1. 控制纪律（Orca 源码级转译，五原则 + 身份纪律）

依据 stablyai/orca v1.4.x 架构分析（MIT）：

- 单一权威 store：terminal/task/worktree 全部状态唯一写入方为
  接入层服务；Web UI 与任何读者只做展示，无直改路径
- 三值活性裁决：终端状态 ∈ {live, unverifiable, exited}；
  live = 注册表近期事件证据；exited = 进程退出证据；
  unverifiable = 两者皆无。UI 如实显示三值，unverifiable 不显示
  为 online
- 类型化 Agent 注册表：AgentSpec {name, launch_cmd, capabilities,
  drive_mode: structured | terminal | both}；新增 CLI agent =
  追加一条注册数据，调度代码零改动
- 终端身份：terminal_id 持久，不绑 PID；exited 后身份保留，
  不因重启新建
- 任务状态机：ready → accepted → working → done → released，
  每次迁移落库（task_id, from, to, ts, operator）；exited 终端的
  未完成任务显示 terminal_exited，不自动重派
- 人工门禁：intent 发起仅由页面人工触发；接入层主机决定的是执行
  方式（怎么跑），不是执行本身（跑不跑）——决策权边界写死；
  无自动派发、无定时派发、无链式派发
- 互斥写域：每个中性 worktree 同时最多挂一个活跃 Agent；
  重复派发被拒绝并提示现有任务
- 权限：项目级 A/B/C 沿用 agents 域实现，默认 A；派发确认页
  必须显示权限范围，无跳过路径
- 收据：派发/发送/读取/停止每个动作一条 receipt 落库；失败首 RED
  保留不覆盖
- 恢复：同 task 不重复派发；终端断线按原生会话域既有 resume，
  不发明新机制

## 2. 寻回册总览页（按参考图实现）

- 统计头：AI 平台数 / 作品数 / 文件夹数 / 接力过作品数 / 最近改动
  作品数；平台标签数据驱动（来自扫描结果，禁止硬编）
- 筛选：平台 / 分组 / 时间线 / 显示隐藏项目
- 作品卡列表：平台标签、描述、最近改动时间、关联文件数、三个
  动作——「从文件核实」「交接手册」（works 域既有能力 UI 化）
  「接力」→ §3

## 3. 接力流程（worktree 中性化）

作品卡点「接力」→ 声明 intent（任务描述 + 偏好 Agent + 权限确认）→
接入层主机执行：

1. 中性 worktree：从作品当前状态创建/复用工作区，创建过程剥离
   Agent 字段——工作区先于 Agent、独立于 Agent 存在
2. 挂接 Agent：按 intent 偏好 + 注册表可用性决定；同 worktree 换
   Agent 接力不重建工作区——"接力"的字面实现
3. launch 管道双路径（降级即正常路径，非异常处理）：
   - structured：原生会话驱动（send/read/resume/stop），事件流
     来自会话事件
   - terminal：worktree 内直接起 CLI；事件采集走 git 观察
     （generic watcher 既有路径），活性走三值裁决
   - AgentSpec.drive_mode 声明默认路径；structured 不可用时降级
     terminal，状态记 degraded_launch，不算失败
4. 首条输入 = 交接手册 + 任务描述

## 4. 工作实时视图

- 活跃任务卡：task_id、Agent、作品、分支、状态机当前态、已持续
  时间
- 会话事件流：该任务产生的 native 事件按时间倒序（derived 字段，
  零 raw）
- 动作：发送输入 / 停止（走原生会话域，留 receipt）
- 任务事件同时进入认知事件流（events 域既有路径）——Today 视图
  可见该 Agent 的工作，闭环的可视证明

## 5. 验收

三分钟演示脚本（全程一条 URL）：①打开寻回册 5s 内看到统计与卡片
②筛选某平台 ③点开一件作品看交接手册 ④接力给一个 Agent ⑤实时
视图看到它工作 ⑥切 Today 看到事件出现。

工程验收：
- 三值裁决实测：停掉终端进程 → exited；静默终端 → unverifiable；
  活跃会话 → live，三者 UI 区分可见
- Agent 换手：同一 worktree 上 Codex 接力 Claude 的作品，工作区
  状态保留，无重建
- 注册表扩展：注册表加一个新 CLI 条目，不改调度代码即出现在
  可选列表
- 降级路径：一个 terminal-only Agent 完成完整接力流程，git 观察
  事件进认知流
- 互斥写域拒绝 / 权限拒绝路径 / 状态机迁移落库，各有 E2E
- 零写审计：接入层对 V4 正式库零写（Procmon 惯例）
- 00:30 自动快照不受影响
- 平台列表、作品计数全部来自真实扫描

## 6. 红线

- 无自动派发（QUESTIONS Q6 纪律，澄清前不开启）
- 接入层不写引擎、不写认知状态——只产生事件，认知由引擎更新
- Agent 任务活动如实进事件流（含本原型开发本身）——实验期数据
  如实记录，不过滤
- 派发必须过权限确认页，无跳过路径
- 寻回册平台列表禁止硬编

## 7. 实现窗口与排期

窗口：本任务书确认后立即开工，三天。
D1 终端注册表 + 三值裁决 + 中性 worktree 服务 + 寻回册总览页
D2 接力流程 + 工作实时视图
D3 打磨 + E2E + 演示脚本 + 常驻部署
与 Workbench v0.1 收尾三项（Procmon/两日快照记录）并行，无依赖。

## 8. 文风纪律

存活标准：推进实现或传递信息。以下句式出现即删：
"为了稳妥起见"、"这在某些情况下可能会有问题"、"仅供参考"、
任何不改变下一步动作的免责声明。
