# 文件归属与交接

开发分支与准确输入见[上下文](00-context.md)。本轮当前隔离 checkout 由各卡按文件互斥修改，主控统一安排提交。共同规则及实施依据见[执行计划](../../../tasks/swa-task-cards-20261022.md)；已有历史工作树不代表当前云机有对应进程或仍在写入。

| 卡片 | 唯一写入范围 | 直接输入与输出 |
| --- | --- | --- |
| 01 共享合同 | `contracts/`；生成的 `web/src/api/schema.d.ts`；`docs/SPEC.md` §21.1；`docs/QUESTIONS.md` 当前整合记录；`docs/TASKS.md` 整合入口；本目录 `00-context.md`、`01-contract.md`、`source-reuse.md`、`ownership.md` | 输出实际命令/身份/状态/权限/数量/成果规则，不写执行实现 |
| 02 运行时 | `runtime/swa-orca/src/`（含其可执行 `contracts.ts`）；`internal/adapters/orca/`；`internal/workspace/app/orca_execution.go` 及专属执行测试；必要 `internal/adapters/httpapi/orca_cli.go`、`internal/adapters/sqlite/orca_execution.go`、`cmd/server/orca_runtime.go` 直接消费者；运行时专属回归测试 | 复用01合同，改正批准额度/原生执行/输入/退出/成果；公共JSON Schema请求交01 |
| 03 CLI | `cmd/swa/`；`runtime/swa-orca/USAGE.md`；CLI专属参数、输出和生命周期测试 | 消费真实App服务；不直接读数据库/调用原生Agent；组合serve静态目录与05协作 |
| 04 GUI | `web/` 页面、组件、样式、前端API消费者和GUI测试；排除 `web/src/api/schema.d.ts`、`web/package.json` | 消费现有OpenAPI/human session，展示同Run/Dispatch；需要新API先交01/02 |
| 05 交付 | 根及前端依赖声明、`pnpm-lock.yaml`、`pnpm-workspace.yaml`、`dependencies.lock.json`；`scripts/build-swa-*.mjs`、安装/闭包与native脚本、`.github/workflows/`；许可说明；本目录 `05-release.md` | CLI候选依赖01/02/03；组合候选再含04，准确同源码产物交06 |
| 06 最终验收 | `tests/swa-product-acceptance/`、`docs/validation/swa-product/` | 消费真实安装物和当前合法授权；记录PASS/FAIL/NOT_RUN/未知并将缺陷交原负责人 |
| 主控 | `tasks/swa-task-cards-20261022.md`、本轮HANDOFF/STATUS接续、集成提交与分支管理 | 维护优先级、接收审查结果、处理跨卡文件请求与外部条件 |

共享路径是模块索引，开始写某个既有测试文件前还须确认实际原负责人。`tests/swa-cli-headless/` 中既有真实安装和模型脚本由06读取复用；02/03不与06同时修改同文件，新增回归须明确归属。发现范围重叠时暂停该文件写入，由主控分配，其他互斥工作继续。

## 后续卡的具体合同输入

- 02：`01-contract.md` 的原生身份、原请求、批准额度与完整退出；`max_workers` 取批准task spec，不从前端数目或模型文字推断。
- 03：原有 `controller/task/worker/request`、错误/退出6与130、单JSON结果及导出边界；不照抄参考项目的6或自动重发策略。
- 04：OpenAPI现有Orca命令路径与envelope、human session/CSRF、原生result按动作校验；无证据不制造角色、状态、事件或审批。
- 05：CLI与CLI+GUI两阶段、同后台/安装根、固定工具与许可来源；GUI不阻塞CLI候选。
- 06：当前可配置N=3的真实场景、明确`max_workers=N`时占满后的N+1席拒绝、同session追加、指定停止与其余继续、同安装跨入口一致性及具体未验前置。

每次交接给准确提交与文件位置、实际命令和退出码、失败/未验项及下一消费者。历史Windows候选、当前云编译及新安装验收分开报告；审查接受不自动授予远端合并、模型费用、部署或发布。
