# M1 Windows 契约命令正常退出

## 目标与完成条件

依照 `C:/Users/DW/Desktop/re.md`、主控 `docs/META_PLAN.md` §5.1、§9.1 M1、§12 A03 及 `tasks/swa-mainline-20261020.md`，修复真实 Windows 契约命令的进程退出故障。原 `pnpm check:contracts` 必须实际完成 Redocly、264 项契约示例和生成类型逐字漂移检查，退出 0；错误输入继续返回真实非零状态。详细依据与测试计划见[根因调查与测试计划](../docs/standards/swa-m1-contracts-20261020.md)。

## 基线、归属与依赖

| 项目 | 实际记录 |
| --- | --- |
| 产品工作树 | `C:/Users/DW/orca/workspaces/Astrocyte/frontend-observatory-20261018` |
| 分支 / 接管 HEAD | `songconmaisaix31-design/frontend-observatory-20261018` / `32ab1eec808a5848aabad40f11133735a1214e55`，接管时干净 |
| 原实际运行源码 S | `ed543c12106c2dbb3e181cb9bc1647b626455179`；原服务和数据继续由主控管理 |
| Orca | Run `run_62ada397a104`；Task `task_261b4dab6d89`；Dispatch `ctx_958a2bf0ae27` |
| Worker / 主控 | `term_dd3bd86f-4d58-4e59-b59a-8b643ac53be5` / `term_3209681a-faa5-4161-a2f3-9563e1451a33` |
| 实际客户端与模型 | Orca / Codex；`worker-show` 的实际 projection 为 `gpt-6.1-sol`，终端显示 `GPT-6.1-Sol xhigh fast`；没有覆盖模型或使用子 Agent |
| 前置条件 | 根因方案和测试计划由独立 Reviewer 接受；M0 依赖完成；主控明确实现写域和唯一重操作窗口 |

完整读取桌面 `re.md`；主控 `META_PLAN.md` 全文与它的逐行比较只差外层 Markdown 围栏。已读取当前 HANDOFF、STATUS、AGENTS、CODING、RESOURCES、TASKS、QUESTIONS、SPEC §20、主控交付任务、产品交付报告及两份实际失败日志和结果。M0 在独立文档树实施，本轨不修改原产品治理文件。

## 根因判断与一个方案

原入口与同版本 CLI 直调均在 Redocly 宣告校验有效、输出 13 项警告后发生 `UV_HANDLE_CLOSING` 断言，退出 `-1073740791`（无符号 `3221226505`）。本机已安装 `@redocly/cli@2.60.0` 的命令包装器在 `beforeExit` 中执行 `process.exit(code)`；官方对应源码相同。结合 Node 官方退出说明和 Windows 同类故障记录，目前判断是该强制退出边界触发运行时退出故障。实际异步句柄的创建方尚未定位，因果关系等待审核后同条件前后运行验证。

方案：使用现有 pnpm 的 `patchedDependencies` 为精确 `@redocly/cli@2.60.0` 保存项目依赖补丁，仅把命令包装器既有 `beforeExit` 回调中的 `process.exit(code)` 改为 `process.exitCode = code`，保留 0 / 1 / 2 的计算与错误路径。原 package 命令、校验规则、示例、生成器、运行时版本与遥测行为保持原内容。具体包补丁在已授权的副本中通过文件编辑工具编写，再由 pnpm 登记、生成锁文件及安装。

## write_paths

当前仅允许：

- `tasks/swa-m1-contracts-20261020.md`。
- `docs/standards/swa-m1-contracts-20261020.md`。
- 自有 `.verification/swa-m1-contracts-20261020/**`，已被根 `.gitignore` 排除。

请求主控在独立接受并完成 M0 后确认的实现路径：

- `pnpm-workspace.yaml`：登记精确包补丁。
- `pnpm-lock.yaml`：由当前 pnpm 生成补丁来源及依赖解析记录。
- `patches/@redocly__cli@2.60.0.patch`：只修正实际发布包 `lib/index.js` 中对应的退出操作。
- 上述两份任务文档及自有验证目录。

生成类型仅在自有隔离验证副本中使用真实历史版本形成漂移；产品 `contracts/openapi.yaml`、`contracts/redocly.yaml`、`scripts/validate-examples.mjs`、`scripts/generate-client.mjs` 和 `web/src/api/schema.d.ts` 不需要修改。主控尚未授权上述实现路径，当前没有修改它们。

## 实施、验证与停止条件

独立审核通过后，先由主控处理原 64505 服务、授予唯一窗口，再依序执行详细计划 T01—T08：同条件原入口基线、精确依赖补丁登记与冻结安装、原入口三阶段完成、真实规则错误、真实配置缺失、真实历史类型漂移、独立副本冻结安装及原入口、适用静态检查。所有重操作逐项两次实际采样，CPU 均小于 85%、可用内存均至少 4 GiB，单调间隔 2—5 秒，BelowNormal、Node heap 不超过 1024 MiB，原件保存于自有验证目录。静态检查额度引用 [RESOURCES](../docs/RESOURCES.md)。

每项保存命令、准确源码、实际工具版本、stdout/stderr、退出、耗时及采样。改前和改后保持同一 Node、pnpm、契约、规则和输出方式。保留原失败；不自动重试。运行时崩溃、未知退出、缺少阶段输出、非预期依赖变化或资源拒绝立即记录并向主控报告。资源拒绝的检查保持未执行；根因被否定时先修订方案并重新审核。

实现、自测和全部返修由同一 Worker 继续负责。完成后通过 Orca 提交准确 Commit 与审查材料；本次派发禁止向 GitHub 等其他渠道发消息，PR 创建或更新由主控明确安排。普通 push 与主线集成按主控安排执行，不 force push。M1 不改变 S3 权限、预算、准入、采用或发布条件。

## 当前结果与接续

当前完成只读来源调查和待审测试计划；未运行复现、安装、编译、测试或业务脚本，未启动或停止服务，未修改源码、依赖或系统环境。历史两次崩溃继续为 FAIL，历史示例和生成检查的通过只属于原 S。

下一步通过 Orca ask 提交计划、申请独立审核和实现写域，等待 Reviewer 决策与 M0 完成。此时未创建实现 Commit，未普通 push，未创建或更新 PR。
