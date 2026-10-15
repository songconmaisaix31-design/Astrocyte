# 发布组件许可与再分发边界记录（2026-10-15）

## 1. 范围、写入边界与来源

本文由根 `msg_2b7cd779bafa` 授权并按其后的返修指令维护，当前 Task `task_f0b1dca0e6fc`（承接 `task_b68c997b0289`）。写入范围只有本文件；公开核查原件保存在本树已忽略的 `.verification/components-20261015/`。本文没有修改旧科研报告、业务代码、依赖锁或冻结域，没有读取私人凭据、原会话或正式数据库。本许可工作没有发起产品原生 controller，也没有运行额外的验证模型探针；开发客户端为 OpenCode 1.18.35 / `alibaba-cn/deepseek-v4.1-flash`，单独如实记录，当前 TUI 显示不冒称 provider wire 实证。

本文只记录固定组件名、固定版本与一手来源、适用许可、必需运行文件和真正待定的分发项。本文不复制 S4、S5 owner 的动态测试数字或状态表；其功能验收由 M-release 按各 owner 报告管理。本文不提供法律意见，不决定分发方式，不替根选择尚未决定的组件。

来源（只读，owner 报告按当前状态链接）：

| 来源 | 用途 |
|---|---|
| P 工具 SOURCE `1a1c14d6fdd02e7824efb82a618848a1c315dedb`、候选包 SOURCE `15f27f4a8e1e2300b8b1152cfd5b8e5f722db744` | `release/Build-Windows.ps1`、`docs/release/windows.md`、`docs/release/package-report.md` 给出候选包实际内容 |
| S4 标准 [docs/standards/repository-context-s4.md](https://github.com/songconmaisaix31-design/Astrocyte/blob/6ab17193bae62cb74a3ab2871b736d81c35e2791/docs/standards/repository-context-s4.md)，SOURCE `f407337f2199d5e49f1ff49862063d964a738a9c` / 报告 `0bd2aabf58076960597c1eefca526252857583c7` | S4 固定组件、版本与生产运行文件 |
| S5 标准 [docs/standards/judgment-s5.md](https://github.com/songconmaisaix31-design/Astrocyte/blob/00f8165401c522352bc19dedd204e4409f12cb25/docs/standards/judgment-s5.md)，SOURCE `bc5ffc910b660d6ae9084234ffc8ced8e265cbc5` / 独立 REPORT `cd8cb2d36d4640f1d313de67d402cf00779236a2` | S5 固定 SDK、runtime、模型与 Jev 边界 |
| 各官方上游（见 §3 表内链接） | 许可全文与版本标签的独立核查 |

## 2. 候选包（15f）当前实际内容

`release/Build-Windows.ps1`（候选 SOURCE `15f27f4a8e1e2300b8b1152cfd5b8e5f722db744`）只复制下列确定的产物：四个 Go 程序 `astrocyte.exe`、`astrocyte-mcp.exe`、`astrocyte-sleep.exe`、`workbench-snapshot.exe`；`web/dist` 生产静态资源；三个脚本 `Start-Astrocyte.ps1`、`Start-Astrocyte.cmd`、`Publish-WorkbenchSnapshot.ps1`；`docs/release/windows.md`（重命名为包内 `README.md`）；以及项目根 `LICENSE`。

`docs/release/package-report.md` 记录的候选 ZIP 为 39 项、33,768,080 字节，数据库、日志与凭据检查为零。包内 `LICENSE` 是项目的 `GNU AFFERO GENERAL PUBLIC LICENSE Version 3`（AGPL-3.0），包内没有第三方 `NOTICE` 或第三方许可文本。

结论：候选 15f 只携带项目自身 AGPL-3.0 许可；§3 表中的 S4、S5 固定组件全部不在候选包内。候选包的 15f 源码不是 S4、S5 的组合源码，S4、S5 仍在各自工作树。

## 3. 固定组件许可清单

下表逐项给出官方主链接处核到的许可名称、版权归属与固定版本。`是否在 15f 内`一列按 §2 的候选包实际内容判定。

| 组件 | 固定版本 / 修订 | 官方主链接 | 许可 | 版权归属 | 是否在 15f 内 |
|---|---|---|---|---|---|
| Laya SDK（laya-ts） | Git `NandhaKishorM/laya` @ `432312b36714d8889eba405507b37eda25257d9c`，子包 `laya-ts@0.1.0` | https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/LICENSE | Apache-2.0 | Convai Innovations | 否 |
| Laya SDK 根 | 同上仓库根 | https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/LICENSE | Apache-2.0 | Convai Innovations | 否 |
| Laya Python 组件（仅官方 export 准备） | PyPI `laya` 0.4.7 | https://pypi.org/project/laya/0.4.7/ | Apache-2.0 | Convai Innovations | 否（仅准备环境） |
| Laya 模型（English，根 checkpoint） | HF `convaiinnovations/laya` @ `7b928d828b7b0e022f929d9bd2e44165aa270148` | https://huggingface.co/convaiinnovations/laya/tree/7b928d828b7b0e022f929d9bd2e44165aa270148 | apache-2.0（模型卡 `license: apache-2.0`） | Convai Innovations | 否 |
| Laya 模型（multilingual，同仓库 `multilingual` 子目录，mmBERT-base） | 同 HF 修订，subfolder `multilingual` | https://huggingface.co/convaiinnovations/laya/tree/7b928d828b7b0e022f929d9bd2e44165aa270148/multilingual | apache-2.0 | Convai Innovations | 否 |
| Node.js | 24.16.0 | https://github.com/nodejs/node/blob/v24.16.0/LICENSE | MIT（Node.js 本体）及其链接内列出的捆绑第三方 | Node.js contributors / OpenJS Foundation | 否（运行前置） |
| onnxruntime-node | npm 1.30.0 | https://www.npmjs.com/package/onnxruntime-node/v/1.30.0 | MIT | Microsoft Corporation | 否 |
| ONNX Runtime 原生库（onnxruntime.dll，随 onnxruntime-node 1.30.0 附带） | 1.30.0 | https://github.com/microsoft/onnxruntime/blob/v1.30.0/LICENSE | MIT；NOTICE `onnxruntime-ThirdPartyNotices.txt`（捆绑依赖声明） | Microsoft Corporation | 否 |
| Microsoft DirectML（DirectML.dll，随 onnxruntime-node 1.30.0 附带） | NuGet `Microsoft.AI.DirectML` 1.15.4；DLL 文件版本 `1.15.4+241025-1615.1.dml-1.15.fac7597` | https://www.nuget.org/packages/Microsoft.AI.DirectML/1.15.4 | Microsoft Software License Terms（二进制）`DirectML-LICENSE.txt`；NuGet 源代码 `LICENSE-CODE.txt` 为 MIT；`DirectML-ThirdPartyNotices.txt` | Microsoft Corporation | 否 |
| DirectXShaderCompiler（dxcompiler.dll，随 onnxruntime-node 1.30.0 附带） | `1.8.2502.8`；官方 tag `v1.8.2502` SOURCE `b4711839eb9a87da7c3436d9b212e0492359fbbd` | https://github.com/microsoft/DirectXShaderCompiler/blob/b4711839eb9a87da7c3436d9b212e0492359fbbd/LICENSE.TXT | LLVM Release License（University of Illinois/NCSA）与复合第三方声明 | University of Illinois at Urbana-Champaign / Microsoft | 否 |
| vscode-jsonrpc | npm 8.2.1 | https://www.npmjs.com/package/vscode-jsonrpc/v/8.2.1 | MIT | Microsoft Corporation | 否 |
| sourcegraph/jsonrpc2（Go） | Go 模块 `v0.2.1`，官方 tag SOURCE `2cc94179e136b565cb9df7d9d0bff9fc06912ac9` | https://github.com/sourcegraph/jsonrpc2/blob/v0.2.1/LICENSE | MIT | Copyright (c) 2016 Sourcegraph Inc | 否 |
| CodeGraph | npm `@colbymchenry/codegraph` 1.0.1，源码标签 `v1.0.1` 对应 `a1489f77a6d69538bfe29020b8366ee034f90860` | https://github.com/colbymchenry/codegraph/tree/v1.0.1 | MIT | Colby McHenry | 否 |
| AOCI | `v0.1.0-rc20`，标签对象 `7c59b0718780f5fccb65d9f0a3445b0b98fea177`，实际源码提交 `45990a1dcfc4415efdc6fbfa13bc159a415f127c` | https://github.com/aoci-spec/aoci-code/blob/45990a1dcfc4415efdc6fbfa13bc159a415f127c/LICENSE | FSL-1.1-MIT（Functional Source License 1.1, MIT Future License） | Copyright 2026 Liu JinShi | 否 |
| Jev | `jev-1.13.0` | https://docs.typesafe.ai/introduction | 专有云服务条款，见 https://typesafe.ai/legal/terms | TypeSafe AI, Inc. | 否（云 REST 服务） |

补充事实：

- Laya SDK 的 npm registry 实测 E404；固定组件是官方 Git SOURCE `432312b36714d8889eba405507b37eda25257d9c` 的 `laya-ts` 子目录真实 build，不能把不存在的发布包登记成依赖。根 `LICENSE` 与 `laya-ts/LICENSE` 在实际下载中字节相同，均为 Apache-2.0。
- Laya 模型卡在固定修订处声明 `license: apache-2.0`，并写明「Apache 2.0 · Convai Innovations」。S5 的 ONNX 文件由官方 export 脚本从该 Apache-2.0 权重生成，属同许可的派生资产。
- Node.js 根 `LICENSE` 为 MIT，同时按 `deps/` 逐项列出捆绑第三方库（例如 ICU 为 Unicode-3.0、libuv、c-ares、OpenSSL、V8、zlib、nghttp2、brotli、zstd、undici、corepack、npm 等）。分发 Node 运行时应整体保留该 `LICENSE`。
- `sourcegraph/jsonrpc2@v0.2.1` 的实际 Go sum 为 `h1:2GtljixMQYUYCmIg7W9aF2dFmniq/mOr2T9tFRh6zSQ=`（S5 标准记录）。
- AOCI 在 GitHub 的自动识别许可为 `other`/`NOASSERTION`；实际 `LICENSE` 文件是 FSL-1.1-MIT，授予在「非竞争用途」下使用、修改、再分发的权利，并在发布满两周年后转为 MIT。
- `onnxruntime-node` 的 npm 主包为 MIT，但其附带的原生二进制分属不同许可：onnxruntime 原生库为 MIT（`onnxruntime-LICENSE`）；`DirectML.dll` 适用 Microsoft Software License Terms（`DirectML-LICENSE.txt`），NuGet 源代码 `DirectML-LICENSE-CODE.txt` 的 MIT 只针对源码；`dxcompiler.dll` 适用 LLVM NCSA 许可（`DirectXShaderCompiler-LICENSE.txt`）。不能把全部附属 DLL 统称 npm 主包装 MIT。
- `DirectML-LICENSE.txt` 第 1(a) 条允许在面向 Windows 与 Xbox 的应用/服务中复制与分发该二进制，第 3(c) 条要求保留 Microsoft 声明，第 3(e) 条限制作为独立产品分发；`DirectML-ThirdPartyNotices.txt` 含 Microsoft/GSL、half、Zstandard 声明。精确条款与文件已 Handoff P 与协调端，本文只登记，不批准分发。
- `DirectXShaderCompiler-LICENSE.txt` 为复合文件：LLVM Release License（University of Illinois/NCSA）加逐文件第三方声明，并单列「仅构建与测试依赖」段（含 GPL-2.0 及 Autoconf 例外）。

### 3.1 Jev 云服务边界

Jev 是 TypeSafe AI, Inc. 的专有云服务，`jev-1.13.0` 是版本化 API，调用入口为 `POST https://api.typesafe.ai/v1/systemone`，作为既有授权的判断 backend 使用。它按 Site 条款只授予个人、非独占、不可转让、可撤销的有限使用许可，并禁止反向工程与自动抓取；其权重不公开，是云 REST 服务，不是随包分发的本地模型。S5 当前缺少 `TYPESAFE_API_KEY`，真实云调用未通过。Jev 的条款与价格以 https://typesafe.ai/legal/terms 与 https://docs.typesafe.ai/introduction 为准。

## 4. S4 与 S5 的用途、构造与许可输入

本文前一版本引用的 S4 SOURCE `1a55e1630502642324500c3b64853af882962fa4` / REPORT `a393627f8f090525fdf4e3d5bf0e320477964c74` 与 S5 SOURCE `48abd1fe134b88cfe8122d2aee5a35090fc6d62b` / REPORT `95871f3f3d77d27136e3f3195742e6290c587b4a` 已被当前 S4 SOURCE `f407337f2199d5e49f1ff49862063d964a738a9c` 与当前 S5 SOURCE `bc5ffc910b660d6ae9084234ffc8ced8e265cbc5` / 独立 REPORT `cd8cb2d36d4640f1d313de67d402cf00779236a2` 取代；以下按当前状态登记。

### 4.1 S4 仓库上下文

用途：为已注册项目与 worktree 提供真实 Git 观察、官方 CodeGraph 索引与 AOCI 治理地图，并作为 S2 的 `ProjectSnapshotProvider` 与五个 HTTP 能力端口（Map、Search、Impact、Indexes、MapCommands）的同一 adapter。构造：`repositorycontext.New(db, db, repositorycontext.Options{...})`，其中 Git 从注册根核对 worktree/common-dir/git-dir 与 HEAD；`Graph` 需要 Node 可执行程序、官方 CodeGraph SDK 目录与本轨 `codegraph.mjs`；`AOCI` 需要 AOCI 可执行程序路径；`CacheRoot`、`SnapshotBudget`、`Git.EnvironmentID` 由可信启动配置提供。许可输入：CodeGraph 1.0.1（MIT）、AOCI `v0.1.0-rc20`（FSL-1.1-MIT）、Node 24.16.0（运行前置）。报告：[S4 标准文档](https://github.com/songconmaisaix31-design/Astrocyte/blob/6ab17193bae62cb74a3ab2871b736d81c35e2791/docs/standards/repository-context-s4.md)，SOURCE `f407337f2199d5e49f1ff49862063d964a738a9c`，报告 `0bd2aabf58076960597c1eefca526252857583c7`。

### 4.2 S5 判断

用途：对显式题型与语言做可判断的规则判断，规则可确定时不调用模型；模型按单后端路由调用，Jev 用云 REST，Laya 用 CPU ONNX 常驻 Node worker。构造：`judgment.NewService(Repository, Authorizer, ...Model)`、`judgmentadapter.NewJev(apiKey)`、`judgmentadapter.NewLaya(ctx, LayaConfig)`；`LayaConfig` 必须提供绝对路径的 Node / Worker / ModelDir / AssetsPath / SDKEntry / RPCEntry，以及明确的 Language / ModelRevision / Stderr 与启动、退出预算，不默认选择模型或语言。许可输入：Laya SDK 与 Laya 模型（Apache-2.0）、Node 24.16.0（MIT）、onnxruntime-node 1.30.0（MIT）、vscode-jsonrpc 8.2.1（MIT）、sourcegraph/jsonrpc2 v0.2.1（MIT）。报告：[S5 标准文档](https://github.com/songconmaisaix31-design/Astrocyte/blob/00f8165401c522352bc19dedd204e4409f12cb25/docs/standards/judgment-s5.md)，SOURCE `bc5ffc910b660d6ae9084234ffc8ced8e265cbc5`，独立 REPORT `cd8cb2d36d4640f1d313de67d402cf00779236a2`。

### 4.3 现有记录

论文扩展已由 [extensions/paper/THIRD_PARTY.md](../../extensions/paper/THIRD_PARTY.md) 登记 `@steipete/summarize` 0.25.1（MIT）与 Readability 0.6.0（Apache-2.0）的复用来源与许可全文位置。该记录属既有域，不在本次 S4、S5 固定组件范围内。

## 5. 仍未决定

- SemanticIndex 的具体模型、运行时和向量库仍待用户决定。本文不选默认模型或接口；当前 S4 的 CodeSearch 明确报告 `semantic_embedding_not_available`。
- AOCI 的公开包分发方式用户尚未答复。候选 15f 未包含 AOCI；根已独立核查其 FSL 许可并已向用户询问。本文只登记旧候选未包含这一事实，不对二进制是否携带或是否必须外置下结论。

## 6. 交给 P 的必需运行文件与许可输入

许可保留要求：包内已复制项目 `LICENSE`（AGPL-3.0），当前没有第三方许可文本。每个上游许可要求再分发时保留许可副本与版权声明：MIT 与 Apache-2.0 要求随分发携带许可全文（Apache-2.0 还要求保留源码中的版权、专利、商标与归属声明，若有 `NOTICE` 则须一并保留），FSL-1.1-MIT 要求附带条款副本或链接且不得移除版权声明。凡是把 §3 组件并入最终包的构建，都须把这些组件的许可与版权文本一并纳入包；具体打包方案由 P 按已获授权实施。若最终包含 Node 运行时，须整体保留 Node 根 `LICENSE`（含其逐项捆绑第三方）。

必需运行文件（S4 owner 生产配置）：`runtime/node/node.exe`、`runtime/sdk/node_modules/@colbymchenry/codegraph`（SDK）、`runtime/sdk/astrocyte/repository-context/codegraph.mjs`（adapter 入口）。Node 固定 24.16.0；CodeGraph 固定 1.0.1 与源码 `a1489f77a6d69538bfe29020b8366ee034f90860`。CodeGraph 的实际 SDK 由 P 按官方 Git 源码构建并 npm pack；注册的 npm 包是 `npm-sdk.js` 可选平台包装，可能访问宿主缓存。具体合法安装与打包由 P 按 S4 与 P 报告实施，本文不固定平台包或宿主缓存机制、不臆造新 ENV。

S5 运行时通过 `onnxruntime-node@1.30.0` 携带其原生库与 `DirectML.dll`、`dxcompiler.dll`；相应许可文本随分发携带（P 的 `legal/` 与包内 `runtime/sdk/node_modules/onnxruntime-node`）。

S5 factory 输入：`judgment.NewService(Repository, Authorizer, ...Model)`、`judgmentadapter.NewJev(apiKey)`、`judgmentadapter.NewLaya(ctx, LayaConfig)`；`LayaConfig` 的绝对路径与显式语言、模型修订、预算见 §4.2。Jev 依赖运行环境提供 `TYPESAFE_API_KEY`。

owner 报告链接：[S4 标准文档](https://github.com/songconmaisaix31-design/Astrocyte/blob/6ab17193bae62cb74a3ab2871b736d81c35e2791/docs/standards/repository-context-s4.md)、[S5 标准文档](https://github.com/songconmaisaix31-design/Astrocyte/blob/00f8165401c522352bc19dedd204e4409f12cb25/docs/standards/judgment-s5.md)。

## 7. 验证记录

- 许可与元数据原件（14 个文件）保存在本树已忽略 `.verification/components-20261015/upstream/`，逐个记录 SHA-256 前 16 位与字节数。
- ONNX 附属 DLL 许可原件（6 个文件）取自 P 现成真实原件 `D:/Astrocyte-agent-assets/20261015/release-package/onnx-legal-second/legal/`，副本与 SHA-256 前 16 位保存于 `.verification/components-20261015/onnx-legal-second/`：`DirectML-LICENSE.txt`（A05138E3A085FF60）、`DirectML-LICENSE-CODE.txt`（903DF5512F7D0260）、`DirectML-ThirdPartyNotices.txt`（2C95795C13FF48A5）、`DirectXShaderCompiler-LICENSE.txt`（27A49E35D1DA96EB）、`onnxruntime-LICENSE`（2F07C72751AED997）、`onnxruntime-ThirdPartyNotices.txt`（143764B952FDB1A7）。来源、固定版本与 NuGet nuspec 读取过程见 P 原树 `release/Prepare-OnnxLegal.ps1` 与 `Prepare-Runtime.ps1`。
- 上游核查：Laya 根与 `laya-ts` 的 `LICENSE` @ `432312b3` 均为 Apache-2.0；HF `convaiinnovations/laya` @ `7b928d82` 模型卡为 `apache-2.0`；Node `v24.16.0` 根 `LICENSE` 为 MIT 加逐项第三方；`sourcegraph/jsonrpc2` `v0.2.1` 为 MIT；AOCI `45990a1d` 为 FSL-1.1-MIT；`codegraph` 1.0.1 npm `license: MIT` 且 `gitHead` 为 `a1489f77a6d69538bfe29020b8366ee034f90860`；`onnxruntime-node` 1.30.0 与 `vscode-jsonrpc` 8.2.1 的 npm `license: MIT`；PyPI `laya` 0.4.7 的 `license: Apache-2.0`。
- 版本标签核查：CodeGraph `v1.0.1` 标签指向提交 `a1489f77a6d69538bfe29020b8366ee034f90860`；AOCI `v0.1.0-rc20` 为标签对象 `7c59b0718780f5fccb65d9f0a3445b0b98fea177`；Laya 提交 `432312b36714d8889eba405507b37eda25257d9c`（2026-10-14T05:25:53Z）存在；DirectXShaderCompiler 标签 `v1.8.2502` 指向提交 `b4711839eb9a87da7c3436d9b212e0492359fbbd`。
- 候选包内容核查：按 `release/Build-Windows.ps1` 与 `docs/release/package-report.md` 逐项判定的 39 项目录，仅含项目 `LICENSE`，不含 §3 任一组件。
- 来源核查：S4 SOURCE `f407337f2199d5e49f1ff49862063d964a738a9c`、报告 `0bd2aabf58076960597c1eefca526252857583c7` 与 S5 SOURCE `bc5ffc910b660d6ae9084234ffc8ced8e265cbc5`、独立 REPORT `cd8cb2d36d4640f1d313de67d402cf00779236a2` 均位于各自已 push 的远端分支。
- 文档校核：`git diff --check` 与写入范围检查。

未执行：产品构建、Go/前端/浏览器/模型/原生控制器验收、正式采集与发布。本文只做许可与来源核对，不重跑 S4、S5 的领域测试。