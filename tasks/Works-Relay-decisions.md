# Works Relay 实施接缝与决策

事实源为validation登记 `a429a4d` 的v1.2。v1.0补文与实施窗口已关闭，不重问；业务线独立，validation只文档。

| 事项 | 状态 / 依据 | 下一步 |
|---|---|---|
| 实施起点 | 已核实workbench/v0.1-workbench 9d78835；相对验收d46仅3个Markdown不同；GitHub CI38440379678 head=d46、completed/success | 使用workbench/v0.2-access，不借该CI当新代码验收 |
| 持久接入store | root按v1.2状态/动作receipt落库要求准予R在独占relay_store.go建立独立relay.sqlite两表relay_state/relay_receipts；只接入状态，不修改任何既有认知schema/migrations/源/副本 | 只Service写；误配正式/副本路径拒绝、事务CAS与receipt测试由R完成 |
| 权限 | A/B/C复用既有ProjectSettings/ProjectGrant，只是确认摘要，不建立新授权系统 | 每动作重新核查；人工确认页必经；安装/已配置/能力/可用分别显示 |
| structured工具限制 | R发现既有原生会话对工具有限制；优先调查已有NativeRequest能力，未支持任务所需工具时按v1.2走terminal并记录degraded_launch | 不改native共享实现，不假称无工具会话完成接力；未知启动不降级成第二写者 |
| Today实时接缝 | 已向用户问：现有快照每日发布，是否可追加本地derived实时事件且保留快照数值 / 边界；尚未回答 | H调查事件接入，F先做不依赖此项的UI |
| 作品当前状态 | 已问获准未提交/未跟踪文件及非Git/单文件作品处理；尚未回答 | R先做无Agent工作区端口与互斥，不暗中缩小验收范围 |
| 参考图 / 活性窗 | 本轮未附图；已问沿既有视觉及可配置60秒近期窗口；尚未回答 | 不冻结默认窗或新视觉；用户答后同owner落实 |

## 固定owner

Orca运行 `run_d577f0d3fee1`，coordinator `term_d7f7e173-a64c-43fe-981a-dbe57124de87`。三次worker-start均 `ready` / `turnStart=observed`，未出现unknown，不重复派发。

| 轨 | 工作树 / 分支 | Task / Dispatch | 客户端模型 |
|---|---|---|---|
| R | works-relay-runtime / songconmaisaix31-design/works-relay-runtime | task_f4a5f33cc891 / ctx_fd41ff32f67d | Codex；owner核对面板GPT-6.1-Sol high fast |
| H | works-relay-http / songconmaisaix31-design/works-relay-http | task_65836fdd6887 / ctx_bcc3677fe29c | Codex；provider模型待owner实证，launch requested/effective model均null |
| F | works-relay-frontend / songconmaisaix31-design/works-relay-frontend | task_fae5b6564156 / ctx_a7f2a326ff24 | Codex；provider模型待owner实证，launch requested/effective model均null |

三轨固定write_paths见一页计划。早期Go/API Handoff不等于实现完成；后续SOURCE/REPORT、真实换手/terminal-only/Procmon/自动快照及最终部署各自核验。首失败与未知结果保留，仅derived汇总进Git。
