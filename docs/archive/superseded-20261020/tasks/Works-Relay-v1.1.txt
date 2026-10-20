# Works Relay 任务书 v1.1 修订（参考 stablyai/orca 实现方案）

登记日期：2026-10-14。来源：用户本轮修订正文。状态：REVISION_REGISTERED_REVIEW_PENDING；尚未启动实现。

依据：Orca v1.4.x 源码架构分析（MIT，TypeScript）五原则 + worktree 中性创建。以下替换 v1.0 对应章节，其余条款不变。

版本边界：当前仓库任务与规格目录未找到 Works Relay v1.0，Git 提交标题检索也未找到对应记录；已请求用户提供路径或正文。本文完整保存本轮替换条款，不补写 v1.0 未提供的条款，不宣称已形成完整合并版。v1.0 对照和实现窗口未裁决前，不冻结新增接口或启动业务实现。

## 1. 控制纪律（Orca 源码级转译）

- 单一权威 store：terminal/task/worktree 全部状态唯一写入方为接入层服务；Web UI 与任何读者只做展示，无直改路径。
- 三值活性裁决：终端状态 ∈ {live, unverifiable, exited}；live = 注册表近期事件证据；exited = 进程退出证据；unverifiable = 两者皆无。UI 如实显示三值，unverifiable 不显示为 online（诚实纪律与 Astrocyte 事件采集同构）。
- 类型化 Agent 注册表：AgentSpec {name, launch_cmd, capabilities, drive_mode: structured | terminal | both}；新增 CLI agent = 追加一条注册数据；调度代码零改动。
- 终端身份：terminal_id 持久，不绑 PID；exited 后身份保留，不因重启新建（沿用）。
- 任务状态机：ready → accepted → working → done → released，迁移落库（沿用）；exited 终端的未完成任务显示 terminal_exited，不自动重派。
- 人工门禁：intent 发起仅由页面人工触发；主机决策的是执行方式（怎么跑），不是执行本身（跑不跑）——决策权边界写死。
- 互斥写域：中性 worktree 同时最多挂一个活跃 Agent（沿用）。

## 2. 接力流程（worktree 中性化修订）

作品卡点「接力」→ 声明 intent（任务描述 + 偏好 Agent + 权限确认）→ 接入层主机执行：

1. 中性 worktree：从作品当前状态创建/复用工作区，创建过程剥离 Agent 字段——工作区先于 Agent、独立于 Agent 存在。
2. 挂接 Agent：按 intent 偏好 + 注册表可用性决定；同 worktree 换 Agent 接力不重建工作区——这是“接力”的字面实现。
3. launch 管道双路径（降级即正常）：
   - structured：原生会话驱动（send/read/resume/stop），事件流来自会话事件。
   - terminal：worktree 内直接起 CLI；事件采集走 git 观察（generic watcher 既有路径），activity 三值裁决。
   - AgentSpec.drive_mode 声明默认路径；structured 不可用时降级 terminal，状态记录 degraded_launch，不算失败。
4. 首条输入 = 交接手册 + 任务描述（沿用）。

## 3. 验收增补

- 三值裁决实测：人为停掉一个终端进程 → exited；静默终端 → unverifiable；活跃会话 → live，三者 UI 区分可见。
- Agent 换手：同一 worktree 上 Codex 接力 Claude 的作品，工作区状态保留，无重建。
- 注册表扩展：注册表加一个新 CLI 条目，不改调度代码即可出现在可选列表。
- 降级路径：一个 terminal-only Agent 完成完整接力流程，git 观察事件进认知流。

## 4. 实码基线与复用评审（非新增业务条款）

基线：`validation/v1-20261012` / `47e83cfada27b020e61da22679a3fb690c28b898`；登记前工作区干净，远端分支同 SHA。现有技术栈为 Go、React / TypeScript / Vite、SQLite、OpenAPI；保持现有目录与 DDD 两层结构。

已读取 [开发规范](../AGENTS.md)、[交接](../HANDOFF.md)、[状态](../STATUS.md)、[业务规格](../docs/SPEC.md) §3、§6–8、§12，以及 [开发规划](../docs/TASKS.md)、[待定项](../docs/QUESTIONS.md)。本轮使用 Codex API 主 Agent（GPT-6）；文档任务单 Agent 完成，无 worker 派发。

| 关注点 | 当前可复用入口 | 已核实差距 / 验收边界 |
|---|---|---|
| 作品与交接手册 | [作品领域](../internal/workspace/domain/works.go)、[应用服务](../internal/workspace/app/works.go)、[SQLite 存储](../internal/adapters/sqlite/works_store.go) | Work 当前是作品索引，不是执行绑定；“接力过”仅统计非空手册，不能算真实 relay |
| Works 页面与命令 | [页面](../web/src/pages/workspace/WorksAlbumPanel.tsx)、[HTTP](../internal/adapters/httpapi/works.go)、[API 契约](../contracts/openapi.yaml) | 已有人工元数据与请求回执；本轮检索未发现 Relay intent、terminal_exited、degraded_launch 或三值终端 UI 实现 |
| 原生 Agent 驱动 | [注册表](../internal/adapters/agents/native.go)、[进程驱动](../internal/adapters/agents/native_process.go) | 注册表仍固定列出 Codex / Pi / Claude / OpenCode；已有进程与会话能力，不等于数据注册新 CLI 的 terminal 路径已支持 |
| git 观察入认知流 | [Git 记录映射](../internal/adapters/events/git_record.go)、[仓库观察](../internal/adapters/events/repository_observation.go)、[live 采集](../cmd/validation/live_capture.go) | 已有 git 记录采集和映射；generic watcher 在 Relay 中的运行、worktree 绑定及活性证据来源尚需对照 v1.0，不能将现有闭日采集视作接力验收 |
| 人工授权与复用边界 | [会话与 origin](../internal/adapters/httpapi/session.go)、SPEC §7–8 | 复用现有人工控制面、授权根与错误契约；意图、授权、任务状态、终端活性保持各自语义，作品态度不触发执行 |

上游参考：[stablyai/orca](https://github.com/stablyai/orca)。本轮读取其公开仓库入口，可确认 MIT 标识及 CLI / worktree 方向；未锁定 v1.4.x 的具体 tag / commit，也未独立完成源码级五原则核验。§1–3 是用户提供的项目要求，不把入口阅读冒称为上游源码验证。

### 4.1 待裁决事项

以下问题仅列出，不据此选择组件、改写验收或生成接口：

1. v1.0 的准确路径 / 正文，以及本修订对应章节。缺失时只保留修订，不能补全“其余条款”。
2. 实现窗口：保持 V4 验证期冻结、先评审，还是用户明确允许独立分支 / worktree 实现。当前等待答复，未解除冻结；[Workbench v0.1](Workbench-v01.md) 的独立实施授权不自动转给 Relay。
3. 活性“近期”窗口、证据同时存在时的裁决顺序、服务重启后退出证据的恢复口径；静默不推断退出。
4. 非 Git / 单文件作品、未提交修改的“作品当前状态”如何进入中性 worktree，以及复用选择规则；现有 Works 同时支持 folder / file，不能擅自缩为已提交 Git 仓库。
5. structured 不可用与启动结果未知的分界。已有 [错误契约](../docs/SPEC.md) 的 delivery_unknown 不允许自动重试；降级前需避免已启动进程造成第二写者，不能把未知结果直接作为安全降级依据。
6. terminal-only 首条输入与任务 accepted / working / done / released 的确认依据，以及持久 terminal_id 重启 / 换手的绑定规则；git 活动本身不足以证明 done。
7. 上游源码分析的具体版本和引用路径，以及既有 generic watcher 的准确入口；不凭名称新造采集器或 Orca 调度器。

前两项已向用户集中发问；其余待 v1.0 对照后消除已有答案，仅对仍未决定的关键项请求裁决。

### 4.2 实施与验证纪律

- 本轮只登记 Markdown。§3 四项验收全部 NOT_RUN；未创建实现分支 / worktree，未启动、停止或切换任何 Agent 终端。
- `docs/EXPERIMENT.md` 已在本轮核对 SHA-256：`5aa238599b6dd59e28d9ca30f71b6b9d547822b46d8dadc477092a46c09ab1b7`。协议、日程、正式库、引擎、schema、参数与业务源码不变；本轮不读取新增 raw。
- 若后续允许实现，先补齐 v1.0 对照和关键裁决，再按确切互斥文件写域组织最少开发轨；公共契约、迁移、入口由一个 owner 持有，主控负责独立验收与最终提交，普通合并并 push。不能用粗目录重叠、额外证明系统或 Mock 代替可运行流程。
- 当前检查仅限修订覆盖、相对链接、Git diff、协议未变和 push 后远端 SHA；业务测试 / 构建 / 四项真实接力验收本轮未执行，文档检查不能提升功能完成状态。
