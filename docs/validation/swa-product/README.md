# SWA 产品验收

本目录由第06卡负责。开发基线为 `6195f4cfb6b32aedd43845443c029ecc11d5578d`，具体计划见[本轮计划](plan-20261022.md)。历史安装和两 Worker 结果保留其原范围。

`tests/swa-product-acceptance/preflight.mjs` 只读观察当前平台、明确选择的安装根、目录、Agent 文件及变量存在状态。支持观察 Windows x64 和 Linux x64 glibc；新候选的 `PLATFORM` 为 `win32-x64` 或 `linux-x64-glibc`，缺失时仅兼容原 Windows 候选，不能将其当成 Linux 安装。它不申请 Grant，不读取凭据文件，不调用模型，不执行 Worker，也不确认产品验收通过。`SWA_PROJECT_TOKEN` 已设置仅证明变量存在。

在仓库根目录运行：

```sh
node tests/swa-product-acceptance/preflight.mjs --workers 3
```

当前平台具有实际候选和合法当前授权后，可以指定全部路径。`--agent` 是明确选择的原生可执行文件绝对路径，`--model` 是对应已批准模型名称。私有数据根、安装根和公开测试项目必须分离；Linux 数据根还要求当前 uid 所有、0700。凭据通过受保护运行环境注入，不放入命令参数、报告或 Git。

```powershell
node tests/swa-product-acceptance/preflight.mjs --candidate $InstallRoot --data-root $DataRoot --project $PublicProject --agent $AgentExecutable --model $ApprovedModel --workers 3
```

退出码 `2` 表示参数无效；`1` 表示明确选定的实际输入检查失败；`3` 表示验收尚未运行或前置未完整验证。该脚本始终保留真实产品场景为 `NOT_RUN`，不运行其他平台的候选。主机平台已观察不代表已有对应候选。最小文件及版本观察不能代替 `check:swa` 闭包、权限、实际完整所属树和模型验证。

## 固定公开输入准备

用户已确认本轮真实模型测试费用不设上限。ToolsOff、固定公开输入、当前项目/模型/外发批准、原请求去重及资源退出要求继续适用；预算决定不会创建账号或项目 token，也不扩大私有资料范围。跨创建者的新许可仅为人类独立当前权限下的 `worker.send`，不包括 `worker.start`。

准备器通过现有 HTTPS Git 代理读取固定公开 stats 项目的准确提交，核对实际四文件，不执行 hook、子模块或项目脚本。选择仓库外尚不存在的专用目录：

```sh
node tests/swa-product-acceptance/prepare-public-case.mjs --root /workspace/swa-public-case-20261022 --workers 3
```

实际生成任务规格、N份不同职责的公开输入、同会话后续输入及来源记录；这些均是验收输入，不是原生身份、事件、状态或模型结果。目录已存在时拒绝覆盖。此准备器按当前完整成果100项观察边界接受 N=3..100；这不是产品 `max_workers` 的硬上限。更大的完整验收需要先解决已有分页/完整库存限制。

当前实际公开案例为 `/workspace/swa-public-case-20261022/public-project`，提交 `da3bdddb085808863090a05f60cc1c64415039cf`。2026-10-22实际准备退出0，已有目录拒绝退出1，N=2/3.5/101拒绝退出2，未调用模型。Linux后继前置观察退出3，主机平台匹配但没有据此声明安装候选或里程碑通过。日志在 `.verification/swa-product-acceptance/`。

现有真实验收入口为 `tests/swa-cli-headless/installed-flow.mjs`、`public-case.mjs` 和 `scripts/verify-swa-cli-headless.ps1`。它们当前具有 Windows 专用路径和两 Worker 条件，完整多 Worker 驱动须在公共契约及实际候选完成后由本卡接续，不能把现有脚本直接记为 N=3 通过。
