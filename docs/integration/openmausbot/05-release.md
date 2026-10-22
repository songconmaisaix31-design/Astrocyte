# SWA-05 安装交付与 CI

本轮开发基线为 `6195f4cfb6b32aedd43845443c029ecc11d5578d`，包含 PR #5 的 CLI 实现。当前云机器为 Linux x64。Windows x64 安装候选、Windows ACL、原生 Job 和真实模型验收尚未执行。源码与工作目录由主控统一提交，不能把当前未提交工作树打成可发布候选。

## 开发依赖与检查

工具使用 Node 24.16.0、pnpm 11.27.0 和 Go 1.27.2。只安装产品使用的六个工作区中可跨平台开发的四个目标：

```sh
pnpm --filter astrocyte --filter @astrocyte/web --filter @swa/orca-runtime --filter @orca/process-host install --frozen-lockfile
pnpm check:contracts
pnpm typecheck:swa
go test -count=1 -mod=readonly -v ./cmd/swa
```

根 `pnpm-workspace.yaml` 保留原有 `allowBuilds` 与补丁，未关闭未知安装脚本检查。Windows 原生工作区保留依赖链接，原生构建由候选构建脚本按随包 Node 的 ABI 显式执行。直接安装全部工作区会触发两个 Windows 专属工作区的隐式 node-gyp 安装，不是 Linux 开发依赖的必要步骤。

`pnpm typecheck:swa` 使用 8192MiB heap，继续保留原 TypeScript 的 `strict`、`noEmit` 和完整诊断。候选构建也调用该入口。CI 的独立 `swa-foundation` 工作在 Ubuntu 与 Windows 上执行冻结安装、Go 模块校验、契约检查、完整 strict、CLI 边界测试，以及按准确 Git SHA 编译 CLI 并核对 `--version`。这些检查不调用模型。原有应用检查工作保留全部检查，不掩盖历史失败。

## Windows x64 候选

构建需要已有 Python、MSVC、PowerShell 7、固定 pnpm JavaScript 入口，以及干净的准确源码提交。Python 必须显式提供绝对路径，优先级为 `--python`、`SWA_BUILD_PYTHON`；显式无效值立即失败，不切换解释器。脚本核查实际解释器文件、`sys.executable` 与版本。固定 node-gyp 12.3.0 的 `lib/find-python.js` 支持范围为 `>=3.6.0`；本入口明确接受 Python 3.6 及以上的 Python 3，不承诺未来主版本。编译器选择继续使用现有 MSVC 2022 配置，不修改全局 PATH。

```powershell
$env:SWA_PNPM_CLI = '已安装 pnpm 11.27.0 的绝对 JavaScript 入口'
$python = '已安装 Python 3 的绝对可执行文件路径'
pnpm build:swa --output '本仓库 dist 下尚不存在的绝对候选目录' --python $python --edition cli
```

CLI 是默认形态，不需要前端构建或浏览器。组合形态使用同一源码，通过 `--edition combined` 执行真实 `web` 构建并生成 Vite manifest，复制 `web/dist` 到候选的 `web`，同时保存前端依赖的原许可证。根与前端 package、依赖清单、锁文件保存在候选 `metadata`，不会覆盖私有 runtime 的 package。

两个形态都写入准确源码 `VERSION` 与 `EDITION`。`EDITION` 仅接受 `cli` 或 `combined`；旧候选没有该文件时仅按原 CLI 闭包核查，不声明 GUI 能力。组合候选缺少 index、manifest、实际资源文件、manifest 内部导入，或资源经符号链接逃逸时立即失败。manifest 检查只证明引用闭包；实际页面与 API 行为仍须独立 HTTP、浏览器与原生执行验收。

组合候选沿相同 `swa serve` 服务，通过显式 `--web-dir` 绑定本安装的 `web`。该参数与私有服务环境由 02/03 实现，05 只生成资产，不启动第二个 runtime。服务接口完成并经真实候选验证前，不声明组合形态已经运行通过。

```powershell
pnpm build:swa --output '本仓库 dist 下另一个尚不存在的绝对候选目录' --python $python --edition combined
# 使用实际安装目录与已核实的数据、端口、受保护根；本命令保持所属 stdin。
& $swa serve --data-dir $dataRoot --port $port --protected-roots ($protectedRoots -join ';') --web-dir (Join-Path $installRoot 'web') --json
```

候选仍沿原安装脚本安装到仓库外不存在的新目录。安装、版本、私有 ACL、实际 native 装载、服务及所属资源退出必须在 Windows 中核实；Linux 上的编译不能代替这些检查。原始校验和、TLS、原生复用来源及许可证检查保持启用。CLI 候选先交06验收；组合候选形成后在该准确安装重新完成跨入口与可配置 N Worker 验收。此任务不创建公开 Release 或发布环境。

## 本次实际证据

日志保存在本机忽略目录 `.verification/cloud-setup/`，不包含凭据。

| 命令或行为 | 本次结果 | 日志 |
|---|---|---|
| 全六工作区 frozen install | FAIL，两个 Windows 工作区隐式 node-gyp 安装失败，原件保留 | `frozen-pr5.log` |
| 四目标 filtered frozen install，690 条供应链策略 | PASS，退出 0，锁文件未变 | `frozen-linux-pr5.log` |
| `pnpm typecheck:swa`，8192MiB，原 strict/noEmit | PASS，退出 0 | `swa-strict-8192.log` |
| `pnpm typecheck:swa`，02/03源码整合后 | PASS，退出 0，完整 strict 诊断为空 | `swa-strict-integrated.log` |
| `pnpm --dir web build --manifest` | PASS，退出 0；保留既有 chunk 体积警告 | `web-manifest-build.log` |
| 实际 Python 选择、缺失/相对/无效参数拒绝、真实 Vite manifest 与实际生产依赖许可复制 | PASS，退出 0；复制七个生产依赖许可证，四个直接依赖许可原字节一致 | `release-config-check-final.log` |
| `pnpm check:contracts` | PASS，退出 0；264 个示例和生成类型一致，保留20项警告 | `release-contracts.log` |
| `go test -mod=readonly -v ./cmd/swa` 首次 | FAIL，捕获02/03共享 helper 尚未写完的中间状态；等待文件完成后重试 | `cli-boundary-tests.log` |
| `go test -count=1 -mod=readonly -v ./cmd/swa`，02/03接口完成后 | PASS，退出 0；五个非付费顶层测试实际通过，一个真实安装测试显式跳过 | `cli-boundary-tests-uncached.log` |
| `node --test tests/swa-cli-headless/task-budget.test.mjs` | PASS，退出 0，三个测试通过、零跳过 | `task-budget-ci.log` |
| Windows 候选/native/安装/ACL、真实模型、多 Worker | NOT_RUN，当前机器没有 Windows 执行环境 | 无 |

独立配置检查使用真实已安装 Python、真实前端构建输出和已安装依赖许可证，不创建替身服务或假安装候选。最终提交、接口完成后的 CLI 测试与实际候选由主控记录准确来源。
