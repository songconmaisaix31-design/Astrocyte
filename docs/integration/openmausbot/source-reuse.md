# 来源、采用方式与验证条件

SWA 输入为 `main@6445149358f5799fc5a786eb59ff494cddbec72e` 与 PR #5 `6778a66faf6d1ca7b50658f23966db3e406f38c6` 的本地集成 `6195f4cfb6b32aedd43845443c029ecc11d5578d`。下列 SWA 路径均在该基线实际存在；后续变更由[负责人表](ownership.md)指定负责人验证。

本轮新增人类跨创建者`worker.send`独立当前准入、Windows x64/Linux和更广Agent适配的来源与实现缺口见[Agent矩阵](02-agent-matrix.md)。该矩阵以真实45项TUI配置与7项结构化runtime catalog为输入，逐项区分模板/原生adapter/产品政策/实际执行；OpenMausBot交互参考不扩大SWA的Agent账号、工具或平台能力。

OpenMausBot 所有引用固定到 [`dbb73a73ae2cba48fdf608e9e22d1d5d91608532`](https://github.com/songconmaisaix31-design/OpenMausBot/tree/dbb73a73ae2cba48fdf608e9e22d1d5d91608532)。本轮通过 HTTPS Git 读取该准确对象，并核对引用路径。本轮采用方式为既有 SWA 代码直接复用或改造、OpenMausBot 结构/交互参考；本表不表示复制其实现。

| 现有 SWA 能力与路径 | 固定参考路径 | 采用方式与负责人 | 验证条件 |
| --- | --- | --- | --- |
| `cmd/swa/main.go`、`output.go`；`runtime/swa-orca/USAGE.md` | `docs/headless-cli.md`、`server/headless/client.ts`、`output.ts` | 直接复用 SWA 明确目标/单 JSON/退出码；参考无 GUI 使用体验，03负责 | 实际帮助、合法服务请求、JSON唯一输出、错误/130/6与原请求关联 |
| `runtime/swa-orca/src/agent-discovery.ts`、`agent-settings.ts`、`agent-launch-admission.ts` | `server/headless/engine.ts` 的所在位置仅作入口索引 | 改造 SWA 真实注册与稳定选择，02负责；不导入参考项目 engine/账号存储 | 分开核注册、安装、配置、可启动、控制能力和实际已选模型 |
| `internal/workspace/app/orca_execution.go`、`runtime/swa-orca/src/client.ts`、`controller-helper.ts`、`native-methods.ts` | `server/headless/run.ts`、`client.ts` | 直接复用 Run/Task/Dispatch 与 mutation receipts；参考整项任务的观察体验；02/03分工 | N 个不同真实身份、输入不串帧、同 request 不多执行、未知效果只查原件 |
| `internal/adapters/httpapi/orca_cli.go`、`contracts/openapi.yaml`、`contracts/swa-cli.schema.json` | `server/headless/client.ts` | 01维护已有 HTTP/App 合同；CLI/GUI共用实际服务；不导入参考项目“本机即owner”规则 | path/caller/项目/Grant一致，当前设置和CSRF校验，失效与越权拒绝 |
| `runtime/swa-orca/src/task-result.ts`、`cmd/swa/main.go` 的 exportResult | `server/headless/export.ts`、`output.ts` | 复用 SWA 真实 journal/transcript 完整导出；参考成果入口；02/03负责 | 原生成果身份、大小/一致性/秘密检查；预览截断不能作为完整成功 |
| `web/src/pages/swarm/SwarmPage.tsx` 及现有 typed API、DAG 组件 | `src/App.tsx` 的 `Sidebar`、`ChatView`、`ActivityPanel`、`FilesView`、`TeamMapPage` 组织 | 沿用 SWA React 工作台；参考侧栏、聊天与可展开观察面板，04负责 | 实际API/身份绑定、任意已核Worker数量、无证据不画协作边、断线恢复与部分失败 |
| 现有人类授权与 `web/src/api/client.ts` | `src/components/ApprovalCard.tsx` 的所在位置仅作入口索引 | 审批交互参考由04按实际源文件继续核；批准由原SWA后端裁定 | 浏览器不保存模型密钥，不签发执行Grant，不显示未支持审批按钮为可用 |
| `scripts/build-swa-package.mjs`、`build-swa-runtime.mjs`、`build-swa-native.mjs`、`install-swa.ps1` | `package.json`/Electron相关文件仅为产品形态背景 | 改造SWA CLI候选与组合候选，05负责；不引入Electron前置 | 准确源码、依赖/许可、ABI、仓库外安装、同服务资产、所属退出 |
| `tests/swa-cli-headless/installed-flow.mjs`、`public-case.mjs` 和既有生命周期检查 | `server/headless/run.ts` 中完整job观察方式 | 原历史两席证据保持；06另写可配置N的真实验收 | 同安装N=3真实重叠、指定停止、其余继续、同session输入、真实成果与跨入口一致性 |

上述参考中，已完整读取 `docs/headless-cli.md`、`server/headless/client.ts`、`run.ts`、`output.ts`、`export.ts`、`src/App.tsx` 与 `LICENSING.md`。索引列中的其他组件/engine文件仅核路径存在，由实际采用者读取其内容；不把路径存在等同于已审实现或已运行。

## 许可与排除范围

固定提交的 [LICENSING.md](https://github.com/songconmaisaix31-design/OpenMausBot/blob/dbb73a73ae2cba48fdf608e9e22d1d5d91608532/LICENSING.md) 说明：`enterprise/` 是另有许可的 source-available 范围，不属于 Apache-2.0；其他源码沿 [LICENSE](https://github.com/songconmaisaix31-design/OpenMausBot/blob/dbb73a73ae2cba48fdf608e9e22d1d5d91608532/LICENSE)。本轮不采用 enterprise 实现，不通过删除付费层绕过已配置的认证主体。

OpenMausBot/MausBot 名称和吉祥物为其权利人的商标；SWA 使用自身品牌。第三方软件保留各自许可，实际复制任何源码或资产前须继续核该文件及其 `NOTICE`/`third_party` 来源，由05收集必要许可与声明。当前只参考结构和交互，不把 Apache-2.0 概括成全仓、全部资产均可任意复制。

上传的 `Grok-Bot-0.75.0-REPORT.md` 为静态分析材料。附件给出的 SHA-256 为 `ca3f3b1366f5b9d1ad6215dc6c32f427f7e90d106252d7eb655c2f9aaf491377`；其内容不作为闭源发行包代码许可、动态测试通过或本次新权限授权。本轮不移植闭源代码、付费认证、云电脑、语音或连接器市场。
