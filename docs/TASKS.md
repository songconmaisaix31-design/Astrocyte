# 产品开发任务

当前主线为 `songconmaisaix31-design/works-relay-integration`，模块开发、PR和独立CR按 [AGENTS](../AGENTS.md) 及[角色说明](agents/README.md)执行。运行源码S与报告后继分列；当前来源、验证和服务见[交付报告](standards/delivery-candidate-20261020.md)及根 [STATUS](../STATUS.md)。

## 当前候选收敛

目标为准确S的程序、Web、真实检查与原64505服务可对应。构建和必要自测原件已保存；主控使用原库启动S，独立Reviewer已接受同源产物、迁移副本和实际原库运行范围。准确Commit审查见[原PR #2](https://github.com/songconmaisaix31-design/Astrocyte/pull/2)，普通合入见[主控交付记录](https://github.com/songconmaisaix31-design/Astrocyte/blob/validation/v1-20261012/tasks/delivery-candidate-20261020.md)。业务与构建输入保持S，不为报告重复构建。

完整默认CI、原生/S3、两Agent闭环、实际采用、完整新便携模型包和正式Release仍按各自条件判断；本次有限候选不改变这些门禁。

## 下一项：修复Windows契约检查进程崩溃

依据为 `.verification/delivery-candidate-20261020/run-contracts/` 与 `run-redocly/` 的真实FAIL：Redocly输出valid与13警告后Node在 `src/win/async.c:94` 抛出 `UV_HANDLE_CLOSING` 断言，退出 `-1073740791`。单独264项示例验证与生成一致性已通过，原链式入口仍失败。

目标为定位真实退出原因，修复实际项目入口或必要项目工具依赖，使原 `pnpm check:contracts` 完成。保留Redocly全部原规则、264项示例、生成一致性、fast-fail及真实退出；不改业务/API/S3门禁或系统全局配置。

入口为根 `package.json` 的 `check:contracts`、实际 `@redocly/cli`、`scripts/validate-examples.mjs`、`scripts/generate-client.mjs` 及上述原件。原模块Worker先提交根因方案、具体可改路径及测试计划，经独立审核后实现；当前不预选版本或开始修改。

验收为原 `pnpm check:contracts` EXIT0、三个原组成检查均实际完成，并由独立Reviewer复验正常与失败退出。默认CI中的旧替身、真实来源和权限前置另外按实际清单接续；原生/六步待答条件保持。

产品原有规格与任务历史见[准确S的开发规划](https://github.com/songconmaisaix31-design/Astrocyte/blob/ed543c12106c2dbb3e181cb9bc1647b626455179/docs/TASKS.md)；主控、科研与正式实验任务见[主控任务入口](https://github.com/songconmaisaix31-design/Astrocyte/blob/validation/v1-20261012/docs/TASKS.md)。
