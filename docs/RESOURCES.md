# 资源限制

本文件集中保存当前已确认的资源额度。2026-10-19 用户允许提高资源使用；适用依据为主控[十轮任务](https://github.com/songconmaisaix31-design/Astrocyte/blob/42570f89d38e5fe7ef8d63067de4301f6e5df381/tasks/s3-paper-ten-cycles-20261019.md)和[重启任务](https://github.com/songconmaisaix31-design/Astrocyte/blob/42570f89d38e5fe7ef8d63067de4301f6e5df381/tasks/restart-product-research-20261019.md)。仅执行固定只读inspect的专用额度及静态审查范围承接清理前[AGENTS原文](https://github.com/songconmaisaix31-design/Astrocyte/blob/42570f89d38e5fe7ef8d63067de4301f6e5df381/docs/archive/20261020/AGENTS.txt)。新授权由QUESTIONS和对应任务记录后更新本文件。

## 全局运行规则

- 编译、测试、实际服务、浏览器、文稿构建和依赖安装按任务串行安排；全局只有一个重操作窗口、一个浏览器和一个服务。
- 重操作使用 BelowNormal 优先级。只在当前任务及其子进程设置额度，不修改系统全局配置，不终止其他用户进程。
- 已启动检查保留实际使用的额度和结果，不为改变额度重复运行。资源不足时保留未执行状态，继续静态工作。
- 资源拒绝、失败及未知结果保留原件；不缓存采样，不自动重试。启动时满足额度只表示当时通过检查，不代表运行峰值保证。
- 安装与持续服务并行的例外尚未确认；当前服务占用运行窗口时，继续静态工作。服务停止或程序切换由对应任务安排。

## 重操作额度

| 项目 | 要求 |
| --- | --- |
| CPU | 两次实际启动采样均小于 85% |
| 可用内存 | 两次均不少于 4 GiB |
| 采样间隔 | 以单调计时确认两次采样间隔为 2–5 秒，并分别记录 UTC |
| 优先级 | BelowNormal |
| Go | `GOMAXPROCS=4`、`-p=4`、`GOMEMLIMIT=512MiB`，内存参数为软限制 |
| Node | heap 不超过 1024 MiB |

编译、功能测试、Go vet、完整 Verify、服务、浏览器、TeX 构建和依赖安装均遵守本节。具体任务已采用更小额度的检查沿其真实记录，不补写为本节额度。

## 固定只读来源核查额度

仅适用于既有 `cmd/workbench-snapshot/verify --mode inspect`，来源固定为获准公开 Git 和真实 SQLite，原源码、全部来源、Reader、SQL 断言与真实日期保持。

| 项目 | 要求 |
| --- | --- |
| CPU / 可用内存 | 两次实际采样均 CPU 小于 85%、可用内存不少于 3 GiB |
| 采样方式 | 使用已安装 psutil 的实际阻塞 CPU 采样及可用内存；以单调计时确认间隔 2–5 秒，保留两个 UTC 样本与所有拒绝 |
| 优先级 | BelowNormal |
| Go | 单并发、256 MiB 软限制 |
| Node | heap 不超过 512 MiB |
| 运行窗口 | 全局独占 |

不使用首次非阻塞 CPU 零值，不缓存采样或自动重试。该额度不适用于 collect、编译、功能服务、完整 Verify、浏览器、TeX 构建或依赖安装。

## 静态审查

下列只读检查按 BelowNormal、Go 单并发、Node heap 不超过 512 MiB 执行，不套用重操作的 CPU 启动门限：

- `gofmt -d`。
- `git diff --check`。
- PowerShell 5.1 / 7 的 `Language.Parser.ParseFile`，只解析脚本，不执行正文。
- 现有 `scripts/check-architecture.mjs` 中只执行 `go list -mod=readonly` 和依赖边界规则的路径。

逐项串行确认实际退出，不修改检查规则，不安装或下载依赖；缺少真实依赖立即失败。Go vet、构建和运行测试继续使用重操作额度。

纯文档检查可以使用已安装的成熟 Markdown 解析器核对本地链接、文件内容和 Git 原文，采用 BelowNormal、Node heap 不超过 512 MiB，逐项串行执行；不下载依赖，不触发业务脚本、Go、浏览器或服务。
