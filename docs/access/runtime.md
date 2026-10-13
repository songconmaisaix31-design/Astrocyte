# Works Relay runtime

事实源：`Works-Relay-v1.2.md`；基线 `6ebf523ad4abf4f80d78ee4207fa30168f299ba6`。独占路径为新 relay 前缀 Go 文件、relayworktree 目录及本报告。实际客户端 Codex；Orca terminal show 核实面板模型 `GPT-6.1-Sol high fast`。

早期接缝：`app.RelayRuntime` 提供 CreateIntent（不启动）与 ConfirmIntent（人类身份、权确认及当前项目 revision 必经），其余 List/Get/Send/Read/Stop/Resume/Transition 都由接入 Service 维护唯一状态。`RelayEventSink.EmitRelayEvent` 只传 derived 活动，H-http 通过现有 EventService 装配，事件 mapper 和认知引擎保持冻结。

主控已通过 Orca ask 授权新 `internal/adapters/sqlite/relay_store.go` 中的独立 `relay_state` / `relay_receipts` 两表，文件固定在 LOCALAPPDATA/astrocyte-workbench/v0.2-access/relay.sqlite，不改任何既有迁移，不写认知源或副本。A/B/C 是现有 ProjectSettings/ProjectGrant 的确认摘要，不另造授权系统。

实码调查：现有 NativeAdapter 提供 Start/Send/Observe/Resume/Stop，Codex/Claude/Pi 的既有实现禁用文件与命令工具。主控确认：需要写代码的接力不可把无工具会话当成功；安全确定 structured 不可用时才正常降级 terminal，未确定的 delivery_unknown 保留占用且禁止第二写者。未找到运行中的 generic Git watcher；已有 registeredReadCommand 的 fixed Git 元数据观察可复用为接入端口。

用户尚未回答：Today 实时与快照接缝；未提交/未跟踪及非 Git 单文件中性化范围；活性窗配置。接口不设隐含 60 秒窗口，准备策略保持显式配置，不将未回答变成最终产品决定。

当前验证与完整实现待补；早期接口提交不代表功能完成。真实 native/terminal-only 交接、界面验收、Procmon 与两日快照 NOT_RUN。
