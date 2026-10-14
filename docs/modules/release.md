# 集成发布模块执行记录

事实源：[Release v0.1](../../tasks/Release-v0.1-20261014.md)、[AGENTS.md](../../AGENTS.md)、[HANDOFF.md](../../HANDOFF.md)、[STATUS.md](../../STATUS.md)。基线 `dd60de11bca85662f5bb4a5b281cc7b42e6ffd2f`；模块分支 `songconmaisaix31-design/module-release-20261014`。

## 一页执行计划

总负责人只委派模块负责人并最终验收。M-release 管理 Q / P / S 和唯一 I，只修改本文；功能开发人员不派发第四层。模块 Task `task_fc57d8d54d31` / Dispatch `ctx_376d222c774b` / 终端 `term_5cf00e8f-ad9f-4429-8823-79faa4866c6f`；子 Run `run_ab36c13bc549`。使用 Orca CLI 与 Codex；运行时所见模块、Q / P / S 均为 `gpt-6.1-sol`，未据请求参数推断模型结果。

| 长期轨道 | 固定 Agent / 工作树 / 分支 | write_paths | 当前阶段与下一交付 |
|---|---|---|---|
| Q 认知快照 | `term_ae3e598d-82b7-4d0a-97b1-830e2e75ba51`；`release-cognition-20261014`；`songconmaisaix31-design/release-cognition-20261014` | `cmd/workbench-snapshot/**`、`tests/release-cognition/**`、`docs/release-cognition.md` | 原 `task_1e8ffb3cbbd9` / `ctx_89ea0fc5b202` 继续。可分发命令 SOURCE 已交付；真实来源认知、独立 SQL、分页与历史仍待完成。 |
| P Windows 包 | `term_e67c7493-cef5-439e-9687-3197880820d7`；`release-package-20261014`；`songconmaisaix31-design/release-package-20261014` | `release/**`、`docs/release/**` | 原 `task_76c43934b15c` / `ctx_6d251e04179d` 继续。完成 Q binary / 显式手动快照包装，最终对 I 精确 SOURCE 重建 ZIP 并新目录验收。 |
| S 科研核验 | `term_131cc610-a2b4-4b1b-8335-b5c4e9d9512a`；`release-research-20261014`；`songconmaisaix31-design/release-research-20261014` | `docs/release-research.md` | 原 `task_c2ab97d9ed86` / `ctx_8df7a990a423` 已 succeeded / completed；总负责人已接受合法 `worker_done` 并确认 released / transcript captured。精确文档作为 I 输入保留，未重派同域人员。 |
| I 唯一集成 | 领域终稿可验后，独立 release 集成工作树 / 分支 / 第三层 Agent | 必要路由、导入、配置和类型衔接；集成报告 | 尚未派发。普通 `--no-ff` 精确消费各 SOURCE / REPORT 与管理文档，领域失败回原 owner；完整检查后 push SOURCE，交 P 同源重建。 |

原活跃 Dispatch 保持原 Run / 终端 / 工作树 / 分支 / 写域，直接向模块 Dispatch 或子 Run 汇报；原 `worker_done` 仍由其当前 Run 接收。只有总负责人明确接收结算并转交终端，才在子 Run 给同一 owner 创建后续 Task。没有重新绑定总负责人旧 Run。

依赖顺序：Q 命令 SOURCE → P binary 与脚本；Q 真实认知来源 / SQL expected / 独立快照根 → F 实际查询、分页与历史；B / F 和 R / H 的最终同源 SOURCE / REPORT → I 精确集成 / 全部适用验证 → P 最终 archive / ZIP 解压验收 → 总负责人最终验收与 GitHub Release 发布。S 的缺口直接交原 owner 和对应模块负责人。

## 精确交付与验收状态

| 输入 | 精确 SHA / 来源 | 证据范围与状态 |
|---|---|---|
| Q 命令 SOURCE | `8001a76646f5892bc18d8ccb9e5cd065b51425a6` | 已核对远端一致；Q 已明确接管回执，REPORT 尚未提交。Q 报告真实空库、七个生产 Reader、失败不发布、旧 immutable 代次检查 PASS；尚未独立签认最终组合与非空认知。 |
| P 工具 SOURCE / REPORT | `da45adfe72a5b7da721bd16be0a1029f8fa05e66` / `e5e4d26bdc5ff64560c8f55aa633c54273161b91` | P 报告候选基础 28 项、工具复验 24 项 PASS；候选输入 `c7a70be963c432bb72d0fa355705ce2907821120`，ZIP 26,309,211 bytes。候选路径待 P 精确补充；最终认知 / Relay / 组合验收未完成。 |
| P 快照包装阶段 SOURCE | `d0fb7ff0bb1e791583cc405f098eb684affce35e` | 已核对远端一致；P 已明确接管回执。PS5 参数默认值首次失败保留，本域修复后实际发布通过但修复尚未 commit；独立 Q 8001 archive 编译及候选真实 SQLite 发布 status:ok。Q 普通合并的 HANDOFF.md 冲突转根文档 owner。 |
| P 默认目录修复 SOURCE | `27e65dd80436030b0157edb6f811c83320dda001` | P 报告已 push，Q 普通合并 `4b15b70` 保留；正在从此精确 SOURCE 重新构建候选，阶段 REPORT 仍待。该候选继续等待 I 最终同源输入。 |
| S 科研报告 | `7d030fc83120052a8f32ea74c6007ff9be6878ca` | 已核对远端一致与唯一文档写域；本模块只读审阅六组范围、主来源方法 / 限制、具体 owner Handoff 与未核实资料。S 未重跑产品、模型或正式实验；论文作者成绩保持归属作者。 |
| Python 标准整改 SOURCE / REPORT | `a1089f94a13306e3cd250ed8facbb438d6e75a19` / `91e3e59dcfc8024cd2e4955ac72f627e65c18e71` | 总负责人交接：独立 29 + 25 测试、36 文件 / 14 冻结 AST PASS；其余 28 项实验 NOT_RUN 保留。 |
| Relay 阶段 SOURCE / REPORT | `fbfaa65c2c01094305424ac40d2d39d271eab51b` / `53231b19ddd19a65c5a47c91469a0e34d73cf600` | R 阶段生命周期验证；CreateIntent 需原 owner 修复 SettingsRevision。H 观察 SOURCE `7dd7d1b91128a268c035fcfa49da5f9f8667078d`，实际原生三分钟 / 每 event_id / SQL / HTTP / DOM 验收未完成。 |

候选包精确路径：`C:/Users/DW/orca/workspaces/Astrocyte/release-package-20261014/.verification/package-rehearsal-20261014-202942/Astrocyte-v0.1-windows-x64-c7a70be963c4.zip`。该包用于阶段演练，最终 I 同源 ZIP 尚未生成。

模块联络：M-product `ctx_b1cd2e0a39aa`；M-relay `ctx_e6328a370d56` / 子 Run `run_8bbda1c3974c`。双方最终同源交付尚待接收。

## 实际阻塞与边界

- Q 已 clone 获准公开 Astrocyte 来源，报告当前实际 1339 commits。生产 Engine 对当前 time.Now seed 后导入更早 Git 时间的首失败保留。总负责人已于原 ask `msg_2cd6790ae313` 回复 `msg_81e3f3984acb`：允许新的隔离公开 Git 历史生产重放 DB，每 topic 以最早真实 commit.ts 作历史重放初始化起点；所有源 ts / event / Git / 数值保持原样、engine 与既定参数不改。该明确授权已交 Q/F/P，时间阻塞解除；只验认知查询 / 分页与可分发快照，不能称本次 live 主题建立时间、用户亲自评分、正式 V4 或科研收益。
- F 需要 `ASTROCYTE_TEST_COGNITION_LOCALAPPDATA` 独立快照根、独立 SQL expected、两个真实 topic、同 topic 超过 20 个真实 events 及两个实际 historypoints；样本不足继续采集并保持实际状态。H 的 app-state 不能作为已参数化 / Sleep 完成的认知来源。
- P 普通合并 HANDOFF.md 首冲突保留。总负责人 `msg_0a8a4dd18ac6` 临时授予本模块该单文件文档冲突域；已用 apply_patch 保留当前 P/Q 消费事实与两线历史并单文件 stage，diff 检查通过且无未合并文件，通知原 P 完成普通 merge。临时域随此处理结束，其他业务文件与 P 未提交修改保持原 owner 管理。
- Q 发现既有 Reader 的 `WorkbenchFormalSource` 个人路径常量会进入 server binary，Reader open 还会对其执行 stat；本模块只核源码引用，没有读取正式源。已向总负责人升级并 Handoff M-product/B 决定发布线的原 owner 修复，Q/P 不修改 Reader。`-trimpath` 只能除构建路径，最终包须独立核查此业务常量，不能据 trimpath 填为通过。
- M-relay 首次第三层派发遭 `nested_worker_depth_exceeded` / max1 保留。总负责人消息 `msg_498d887dea6a` 已确认通过官方设置把深度改为 2，并独立 readback Value2；第三层后续任务可派发，开发人员仍禁止派发第四层。唯一 I 及本模块后续 Task 尚未派发。
- 原始会话、凭据、SQLite、私人来源和测试原件不进入 stdout / 产品 HTTP / Git / 包。默认包启动不读取个人正式来源、不运行快照 / Sleep / 自动日程。
- 正式 V4、engine / mapper / schema / weights、日程与 17321 保持冻结。ArchiveAppendOnly unknown、PrefixComplete / SessionComplete 不足、恢复 raw gap、首 RED / UNKNOWN / NOT_RUN / partial 按真实范围保留。
- 实际用户作品、自由工具、非 Git 来源和产品活性窗口保持未决定。最终 GitHub Release 由总负责人验收后发布。

## 发布完成门禁

1. 各原 owner 交精确 SOURCE / 独立 REPORT，含首失败、适用验证、实际限制及远端一致性。
2. I 普通精确合并、`pnpm check` / `pnpm build` / 完整默认 E2E（workers=1）及适用原生链路；事件库、SQL / API / UI 交叉核查，失败回原 owner。
3. P 从 I 精确 SOURCE 用标准 Git archive 构建 server / MCP / Sleep / snapshot 与既有 Vite web，保留 license / 必需依赖说明；ZIP 不带开发依赖、数据库、私人配置、日志、secret 或测试原件。
4. P 在新目录解压，以无开发工具的运行环境实测静态资源、session / CSRF、真实读写 / 重启、非空认知正向与 Relay 原生路径，报告完整 sourceSHA / 路径 / 字节大小 / 验证结果。
5. 测试修复导致 SOURCE 变化时复验适用影响，再交总负责人最终接受与发布。本文当前状态为模块管理进行中，尚未达到发布完成条件。

## 模块文档验证

管理计划提交 `e1d1246d40fd1caf19a34ef5e6a10272eea5668c` 已普通 push，并用 `git ls-remote` 核对完整 SHA。文档本地链接 4 PASS，`git diff --cached --check` PASS；未因管理文档变更运行产品测试。首次 push / 一次远端读取的 TLS `SSL_ERROR_SYSCALL` 保留，后续普通 push / 远端读取成功。Q/P/S 的远端 SHA 已独立核对，业务检查仍按上表原 owner 实测范围登记。
