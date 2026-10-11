# Attention Engine 接管基线

分支 `attention-engine/v1-20261011`，业务基线 `dab8fea`，接管文档 `0e14c7d9d4f3c43bf473a25e0f6290e81f14be9a`。客户端 Codex 主控；原八轨 OpenCode / DeepSeek V4 Pro（终端观察）。

2026-10-11 21:43，`GOFLAGS=-p=1 pnpm check` 首场 FAIL：`TestPaperDOMInstalledPinAndIsolation` 无法解析已锁定依赖 `entities/lib/index.js`（MODULE_NOT_FOUND）；其余已运行 Go 包通过，尚未运行后续契约/TS等阶段。没有业务源码变更。

执行 `pnpm install --frozen-lockfile` 和同锁 `pnpm install --frozen-lockfile --force`（均 exit0）。随后 `GOFLAGS=-p=1 GOMAXPROCS=2 pnpm check` 的 importer 包通过（6.244 秒），Go、依赖核验、架构检查和 OpenAPI lint 通过；全场在 `validate-examples` 因 Node `ZoneAllocationFailed/OutOfMemory` exit134 再次 FAIL，后续 TS/lint/Vitest/API 阶段未运行。限制 `NODE_OPTIONS=--max-old-space-size=256` 后单独执行 `node scripts/validate-examples.mjs`，241 个示例 PASS。两次整场 FAIL 均保留，定向恢复不等于整场通过。

W0 恢复时默认并行 `go build ./...` 遇到 Windows paging file too small；已要求各轨单包编译并限制进程并发，未修改分页文件或系统设置。

第三场 `GOFLAGS=-p=1 GOMAXPROCS=2 NODE_OPTIONS=--max-old-space-size=256 pnpm check`：Go/vet、依赖、架构、契约及 S1 API 14/14 PASS；TypeScript 因 256 MB 堆上限 exit134 FAIL，后续 lint/Vitest 未运行。在 `web` 工作目录以 512 MB 堆单独执行 `node node_modules/typescript/bin/tsc -b` PASS；整场第三次失败仍保留。

本次读取 W7 猜测的参考图路径 `C:/Users/DW/Downloads/个人作品蜂群接入后.png` 失败（文件不存在）。已要求复用现有看板及任务明确的五能力，不能声称对该图完成精确复刻。

## 波 A 独立检查（2026-10-11 21:58）

W2 `acb4561` 的 30 天 fixture，由 root 以独立 Node 公式重算：30 点 A/M/confidence 最大绝对误差为 0，10 点触底、5 点触顶。相同工作树执行 `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain ./tests/attention-engine` PASS。该 fixture 使用显式 λ=0.1，仅验证纯数学；尚未覆盖事件持久化、重复消费、主题生命周期及冷重启，M2 未完整通过。fixture 的原始一次性生成使用 Python（Worker 如实记录）；用户在 QUESTIONS 后续已允许 Python，不新增 Python 项目依赖或默认验收步骤。
