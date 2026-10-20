# M1 Windows 契约命令正常退出

## 目标与完成条件

依照 `C:/Users/DW/Desktop/re.md`、主控 `docs/META_PLAN.md` §5.1、§9.1 M1、§12 A03 及 `tasks/swa-mainline-20261020.md`，修复真实 Windows 契约命令的进程退出故障。原 `pnpm check:contracts` 必须实际完成 Redocly、264 项契约示例和生成类型逐字漂移检查，退出 0；错误输入继续返回真实非零状态。详细依据与测试计划见[根因调查与测试计划](../docs/standards/swa-m1-contracts-20261020.md)。

## 基线、归属与依赖

| 项目 | 实际记录 |
| --- | --- |
| 实施工作树 | `C:/Users/DW/orca/workspaces/Astrocyte/swa-m1-contracts-implementation-20261020` |
| 分支 / 接管 HEAD | `swa-m1-contracts-implementation-20261020` / `4132f28ecda414f2a1bd040d9eef50bbe1087cb6`，接管时干净 |
| 原 M1 材料 | `80d4f8d1e323982ef73d880ce9927f4bc5c58340`，通过普通合并 `4f49d6ea3c0cfcfa6bdf4143d576239b158fe878` 保留两份文档及历史；原产品树只读 |
| 原实际运行源码 S | `ed543c12106c2dbb3e181cb9bc1647b626455179`；原服务和数据继续由主控管理 |
| Orca | Run `run_62ada397a104`；Task `task_ceb3b3bb76be`；Dispatch `ctx_8d6a6e1c0145` |
| Worker / 主控 | `term_e9753d8f-70d9-4acd-8cc4-b98a61328b1b` / `term_3209681a-faa5-4161-a2f3-9563e1451a33` |
| 实际客户端与模型 | Orca 1.4.230 / Codex；本 Dispatch 的 `worker-show` 实际 projection 为 `gpt-6.1-sol`，实际终端预览为 `GPT-6.1-Sol xhigh fast`；请求参数中的 model 为 null，没有子 Agent |
| 前置条件 | 准确 `5cd8dbdfa59a0cabe8191e9343ca43122c8afa51` 计划已接受，M0/A01 已接受，实现写域已授权；当前只允许静态准备，重操作窗口尚未交付 |

原调查的规格与来源保留在报告。本次已读取本树 HANDOFF、STATUS、AGENTS、CODING、RESOURCES、相关规格、根主控当前首版 CLI 任务与 QUESTIONS、两份准确 M1 Git 原文和独立审核报告。实际 Git 比较确认本树 package、workspace、lock、contracts、scripts 和生成类型与 `32ab1eec808a5848aabad40f11133735a1214e55`、S 均无差异；`80d4` 相对于 `32ab` 仅新增本任务和报告，`5cd8` 后没有修改已接受的测试表。首版独立运行问题仍未回答，本任务不选择执行组件或扩大 S3 工具、预算与数据权限。

## 根因判断与一个方案

原入口与同版本 CLI 直调均在 Redocly 宣告校验有效、输出 13 项警告后发生 `UV_HANDLE_CLOSING` 断言，退出 `-1073740791`（无符号 `3221226505`）。本机已安装 `@redocly/cli@2.60.0` 的命令包装器在 `beforeExit` 中执行 `process.exit(code)`；官方对应源码相同。结合 Node 官方退出说明和 Windows 同类故障记录，目前判断是该强制退出边界触发运行时退出故障。实际异步句柄的创建方尚未定位，因果关系等待审核后同条件前后运行验证。

方案：使用现有 pnpm 的 `patchedDependencies` 为精确 `@redocly/cli@2.60.0` 保存项目依赖补丁，仅把命令包装器既有 `beforeExit` 回调中的 `process.exit(code)` 改为 `process.exitCode = code`，保留 0 / 1 / 2 的计算与错误路径。原 package 命令、校验规则、示例、生成器、运行时版本与遥测行为保持原内容。具体包补丁在已授权的副本中通过文件编辑工具编写，再由 pnpm 登记、生成锁文件及安装。

## write_paths

本 Dispatch 的文档与临时写入范围：

- `tasks/swa-m1-contracts-20261020.md`。
- `docs/standards/swa-m1-contracts-20261020.md`。
- 自有 `.verification/swa-m1-contracts-20261020/**`，已被根 `.gitignore` 排除。

本 Dispatch 已授权、等待窗口后实施的程序路径：

- `pnpm-workspace.yaml`：登记精确包补丁。
- `pnpm-lock.yaml`：由当前 pnpm 生成补丁来源及依赖解析记录。
- `patches/@redocly__cli@2.60.0.patch`：只修正实际发布包 `lib/index.js` 中对应的退出操作。
- 上述两份任务文档及自有验证目录。

生成类型仅在自有隔离验证副本中使用真实历史版本形成漂移；产品 `contracts/openapi.yaml`、`contracts/redocly.yaml`、`scripts/validate-examples.mjs`、`scripts/generate-client.mjs` 和 `web/src/api/schema.d.ts` 保持原内容。其他文件或依赖变化先提交 Handoff。

## 实施、验证与停止条件

独立审核通过后，先完成详细计划 T00 的当前服务保全和停止，由主控授予唯一窗口，再依序执行 T01—T08（含 T05A）：同条件原入口基线、精确依赖补丁登记与冻结安装、原入口三阶段完成、真实规则错误、真实配置缺失、原示例验证器的 supported capability 拒绝、真实历史类型漂移、独立副本冻结安装及原入口、适用依赖/文档静态检查。T05A 只在隔离契约副本中将真实 `GET /events` 的公开 403 example 原样用于其 501 example，通过原命令取得第二阶段真实错误及后继未执行证据。检查结束或失败后交还窗口，由主控执行 T09 恢复同一 S 与当前原库。所有重操作逐项两次实际采样，CPU 均小于 85%、可用内存均至少 4 GiB，单调间隔 2—5 秒，BelowNormal、Node heap 不超过 1024 MiB，原件保存于自有验证目录。静态检查额度引用 [RESOURCES](../docs/RESOURCES.md)。

主控 `msg_656f36f44f3e` 已核当前 PID45956、UTC 创建时间 `2026-10-19T19:47:04.6572166Z`、S 程序和唯一 loopback 监听。实际 `running.json` 记录同一 S。现成 `candidate-service.py start` 和旧恢复脚本仍绑定 c5 prior，不能直接承担本次恢复。需要的控制工具只在自有验证目录内、审核后编写，再交主控审阅和执行；按当前真实记录保全 SQLite、Q、媒体和 CodeGraph 配置及当前进程环境，停止前重新核实身份，恢复同一 S 和原库。原工具和原件保持，不编造 stage 元数据，不直接写原数据库业务值。

每项保存命令、准确源码、实际工具版本、stdout/stderr、退出、耗时及采样。改前和改后保持同一 Node、pnpm、契约、规则和输出方式。保留原失败；不自动重试。运行时崩溃、未知退出、缺少阶段输出、非预期依赖变化或资源拒绝立即记录并向主控报告。资源拒绝的检查保持未执行；根因被否定时先修订方案并重新审核。

实现、自测和全部返修由同一 Worker 继续负责。主控在对 Orca ask `msg_55a6311b4b99` 的回复中明确授权：补丁验证完成后普通 push、创建和更新 PR，无需重复询问该权限。通过 Orca 提交准确 Commit 与审查材料；主线集成仍由主控安排，不 force push。M1 不改变 S3 权限、预算、准入、采用或发布条件。

## 当前结果与接续

独立 Reviewer 已接受准确 `5cd8dbdfa59a0cabe8191e9343ca43122c8afa51` 的方案与 T00—T09（含 T05A）；接受 hook `msg_38a94c047d3d` 已由主控完整读取并回复 `msg_298a8710f923`。本次主控已确认 M0/A01 接受并授权程序写域，当前尚未交付重操作窗口。

本树自有 `E=.verification/swa-m1-contracts-20261020/` 准备 `run-check.ps1`、`runtime-info.mjs`、`resources.py` 与 `CONTROL.md`。已完整读取原树同名检查脚本及其资源脚本，使用文件编辑工具按本树路径准备；原树服务控制脚本只由主控执行，本树不复制或执行服务控制。旧 Task `task_261b4dab6d89 / ctx_958a2bf0ae27` 已结算，原件继续保留。历史两次崩溃保持 FAIL；本树实际安装和修复结果均为 NOT_RUN。

本树初始没有 node_modules。窗口交付后先用原锁文件冻结安装到本树，再执行 T01；该准备不改变依赖或已接受输入及断言。T02 后保存准确补丁源码 Commit，再继续冻结安装及 T03—T08，T07 使用准确程序 Commit 的干净归档。每条命令采用新的独立输出目录，命令与目录映射见报告。所有原命令、故意拒绝输入与失败原件保持。通过后普通提交、push 并创建针对 main 的新 M1 PR，交原独立 Reviewer；原 PR #2/#4 保持原状态，主控负责最终恢复和集成。
