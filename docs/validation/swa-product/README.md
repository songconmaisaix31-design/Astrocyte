# SWA 产品验收

本目录由第06卡负责。开发基线为 `6195f4cfb6b32aedd43845443c029ecc11d5578d`，具体计划见[本轮计划](plan-20261022.md)。历史安装和两 Worker 结果保留其原范围。

`tests/swa-product-acceptance/preflight.mjs` 只读观察当前平台、明确选择的安装根、目录、Agent 文件及变量存在状态。它不申请 Grant，不读取凭据文件，不调用模型，不执行 Worker，也不确认产品验收通过。`SWA_PROJECT_TOKEN` 已设置仅证明变量存在。

在仓库根目录运行：

```sh
node tests/swa-product-acceptance/preflight.mjs --workers 3
```

Windows x64 具有实际候选和合法当前授权后，可以指定全部路径。`--agent` 是明确选择的原生可执行文件绝对路径，`--model` 是对应已批准模型名称。私有数据根、安装根和公开测试项目必须分离。凭据通过受保护运行环境注入，不放入命令参数、报告或 Git。

```powershell
node tests/swa-product-acceptance/preflight.mjs --candidate $InstallRoot --data-root $DataRoot --project $PublicProject --agent $AgentExecutable --model $ApprovedModel --workers 3
```

退出码 `2` 表示参数无效；`1` 表示明确选定的实际输入检查失败；`3` 表示验收尚未运行或前置未完整验证。该脚本始终保留真实产品场景为 `NOT_RUN`。Linux 不执行 Windows 候选，也不尝试读取 Windows 私有目录。Windows 的最小文件及版本观察不能代替原 `check:swa` 闭包、ACL、完整 Job 和模型验证。

现有真实验收入口为 `tests/swa-cli-headless/installed-flow.mjs`、`public-case.mjs` 和 `scripts/verify-swa-cli-headless.ps1`。它们当前具有 Windows 专用路径和两 Worker 条件，完整多 Worker 驱动须在公共契约及实际候选完成后由本卡接续，不能把现有脚本直接记为 N=3 通过。
