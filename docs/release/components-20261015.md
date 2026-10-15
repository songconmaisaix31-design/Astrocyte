# 发布组件许可与再分发边界记录（2026-10-15）

## 1. 范围、写入边界与来源

本文由根 `msg_2b7cd779bafa` 授权的组件许可简单任务产生，Task `task_08d2ca49a825`。写入范围只有本文件；公开核查原件保存在本树已忽略的 `.verification/components-20261015/`。本文没有修改旧科研报告、业务代码、依赖锁或冻结域，没有读取私人凭据、原会话或正式数据库，没有调用模型或启动实验。

本文只记录固定组件的实际许可名称、版权归属、官方主链接，以及这些组件当前是否已经进入发布候选包。本文不提供法律意见，不决定分发方式，不替根选择尚未决定的组件，也不替代任何运行验收。

来源（只读）：

| 来源 | 用途 |
|---|---|
| P 工具 SOURCE `1a1c14d6fdd02e7824efb82a618848a1c315dedb`、候选包 SOURCE `15f27f4a8e1e2300b8b1152cfd5b8e5f722db744` | `release/Build-Windows.ps1`、`docs/release/windows.md`、`docs/release/package-report.md` 给出候选包实际内容 |
| S4 标准 [docs/standards/repository-context-s4.md](https://github.com/songconmaisaix31-design/Astrocyte/blob/1a55e1630502642324500c3b64853af882962fa4/docs/standards/repository-context-s4.md)，SOURCE `1a55e1630502642324500c3b64853af882962fa4` / REPORT `a393627f8f090525fdf4e3d5bf0e320477964c74` | S4 实际固定组件与版本 |
| S5 标准 [docs/standards/judgment-s5.md](https://github.com/songconmaisaix31-design/Astrocyte/blob/48abd1fe134b88cfe8122d2aee5a35090fc6d62b/docs/standards/judgment-s5.md)，SOURCE `48abd1fe134b88cfe8122d2aee5a35090fc6d62b` / REPORT `95871f3f3d77d27136e3f3195742e6290c587b4a` | S5 实际固定 SDK、runtime、模型与 Jev 边界 |
| 各官方上游（见 §3 表内链接） | 许可全文与版本标签的独立核查 |

## 2. 候选包（15f）当前实际内容

`release/Build-Windows.ps1`（候选 SOURCE `15f27f4a8e1e2300b8b1152cfd5b8e5f722db744`）只复制下列确定的产物：四个 Go 程序 `astrocyte.exe`、`astrocyte-mcp.exe`、`astrocyte-sleep.exe`、`workbench-snapshot.exe`；`web/dist` 生产静态资源；三个脚本 `Start-Astrocyte.ps1`、`Start-Astrocyte.cmd`、`Publish-WorkbenchSnapshot.ps1`；`docs/release/windows.md`（重命名为包内 `README.md`）；以及项目根 `LICENSE`。

`docs/release/package-report.md` 记录的候选 ZIP 为 39 项、33,768,080 字节，数据库、日志与凭据检查为零。包内 `LICENSE` 是项目的 `GNU AFFERO GENERAL PUBLIC LICENSE Version 3`（AGPL-3.0），包内没有第三方 `NOTICE` 或第三方许可文本。

结论：候选 15f 只携带项目自身 AGPL-3.0 许可；§3 表中的 S4、S5 固定组件全部不在候选包内。候选包的 15f 源码不是 S4、S5 的组合源码，S4、S5 仍在各自工作树，Go/npm 依赖集成归 B。

## 3. 固定组件许可清单

下表逐项给出官方主链接处核到的许可名称、版权归属与固定版本。`是否在 15f 内`一列按 §2 的候选包实际内容判定。

| 组件 | 固定版本 / 修订 | 官方主链接 | 许可 | 版权归属 | 是否在 15f 内 |
|---|---|---|---|---|---|
| Laya SDK（laya-ts） | Git `NandhaKishorM/laya` @ `432312b36714d8889eba405507b37eda25257d9c`，子包 `laya-ts@0.1.0` | https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/LICENSE | Apache-2.0 | Convai Innovations | 否 |
| Laya SDK 根 | 同上仓库根 | https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/LICENSE | Apache-2.0 | Convai Innovations | 否 |
| Laya Python 组件（仅官方 export 准备） | PyPI `laya` 0.4.7 | https://pypi.org/project/laya/0.4.7/ | Apache-2.0 | Convai Innovations | 否（仅准备环境，不进入包） |
| Laya 模型（English，根 checkpoint） | HF `convaiinnovations/laya` @ `7b928d828b7b0e022f929d9bd2e44165aa270148` | https://huggingface.co/convaiinnovations/laya/tree/7b928d828b7b0e022f929d9bd2e44165aa270148 | apache-2.0（模型卡 `license: apache-2.0`） | Convai Innovations | 否 |
| Laya 模型（multilingual，同仓库 `multilingual` 子目录，mmBERT-base） | 同 HF 修订，subfolder `multilingual` | https://huggingface.co/convaiinnovations/laya/tree/7b928d828b7b0e022f929d9bd2e44165aa270148/multilingual | apache-2.0 | Convai Innovations | 否 |
| Node.js | 24.16.0 | https://github.com/nodejs/node/blob/v24.16.0/LICENSE | MIT（Node.js 本体）及其链接内列出的捆绑第三方 | Node.js contributors / OpenJS Foundation | 否（运行前置，按绝对路径配置） |
| onnxruntime-node | npm 1.30.0 | https://www.npmjs.com/package/onnxruntime-node/v/1.30.0 | MIT | Microsoft Corporation | 否 |
| vscode-jsonrpc | npm 8.2.1 | https://www.npmjs.com/package/vscode-jsonrpc/v/8.2.1 | MIT | Microsoft Corporation | 否 |
| sourcegraph/jsonrpc2（Go） | Go 模块 `v0.2.1`，官方 tag SOURCE `2cc94179e136b565cb9df7d9d0bff9fc06912ac9` | https://github.com/sourcegraph/jsonrpc2/blob/v0.2.1/LICENSE | MIT | Copyright (c) 2016 Sourcegraph Inc | 否 |
| CodeGraph | npm `@colbymchenry/codegraph` 1.0.1，源码标签 `v1.0.1` 对应 `a1489f77a6d69538bfe29020b8366ee034f90860` | https://github.com/colbymchenry/codegraph/tree/v1.0.1 | MIT | Colby McHenry | 否 |
| AOCI | `v0.1.0-rc20`，标签对象 `7c59b0718780f5fccb65d9f0a3445b0b98fea177`，实际源码提交 `45990a1dcfc4415efdc6fbfa13bc159a415f127c` | https://github.com/aoci-spec/aoci-code/blob/45990a1dcfc4415efdc6fbfa13bc159a415f127c/LICENSE | FSL-1.1-MIT（Functional Source License 1.1, MIT Future License） | Copyright 2026 Liu JinShi | 否（本轮只本机适配，二进制不携带） |
| Jev | `jev-1.13.0` | https://docs.typesafe.ai/introduction | 专有云服务条款，见 https://typesafe.ai/legal/terms | TypeSafe AI, Inc. | 否（仅云 REST 调用） |

补充事实：

- Laya SDK 的 npm registry 实测 E404；固定组件是官方 Git SOURCE `432312b36714d8889eba405507b37eda25257d9c` 的 `laya-ts` 子目录真实 build，不能把不存在的发布包登记成依赖。根 `LICENSE` 与 `laya-ts/LICENSE` 在实际下载中字节相同，均为 Apache-2.0。
- Laya 模型卡在固定修订处声明 `license: apache-2.0`，并写明「Apache 2.0 · Convai Innovations」。S5 的 ONNX 文件由官方 export 脚本从该 Apache-2.0 权重生成，属同许可的派生资产。
- Node.js 根 `LICENSE` 为 MIT，同时按 `deps/` 逐项列出捆绑第三方库（例如 ICU 为 Unicode-3.0、libuv、c-ares、OpenSSL、V8、zlib、nghttp2、brotli、zstd、undici、corepack、npm 等）。分发 Node 运行时应整体保留该 `LICENSE`。
- `sourcegraph/jsonrpc2@v0.2.1` 的实际 Go sum 为 `h1:2GtljixMQYUYCmIg7W9aF2dFmniq/mOr2T9tFRh6zSQ=`（S5 标准记录）。
- AOCI 在 GitHub 的自动识别许可为 `other`/`NOASSERTION`；实际 `LICENSE` 文件是 FSL-1.1-MIT，授予在「非竞争用途」下使用、修改、再分发的权利，并在发布满两周年后转为 MIT。

### 3.1 Jev 云服务边界

Jev 是 TypeSafe AI, Inc. 的专有云服务，`jev-1.13.0` 是版本化 API，调用入口为 `POST https://api.typesafe.ai/v1/systemone`。它的权重不公开，条款只授予对 Site 的个人、非独占、不可转让、可撤销的有限使用许可，并禁止反向工程与自动抓取。Jev 不是可分发本地模型：不得把它当作本地 model 打包进发布包。S5 当前父进程缺少 `TYPESAFE_API_KEY`，真实调用为 `NOT_RUN`；`backend` 配置不授予私有资料外发许可。Jev 的第三方条款与价格以 https://typesafe.ai/legal/terms 与 https://docs.typesafe.ai/introduction 为准。

## 4. S4 与 S5 接入状态

### 4.1 S4 仓库上下文

S4 工作树 `works-repository-context-s4-20261015`，标准 `docs/standards/repository-context-s4.md`，SOURCE `1a55e1630502642324500c3b64853af882962fa4` / REPORT `a393627f8f090525fdf4e3d5bf0e320477964c74`。已固定的组件是 CodeGraph `1.0.1`（MIT）与 AOCI `v0.1.0-rc20`（FSL-1.1-MIT）；Node 24.16.0 作为运行前置。真实 Git、两棵官方 CodeGraph 索引、AOCI 作者地图与只读查询在本机通过；HTTP、前端与最终包组合为 `NOT_RUN`。

### 4.2 S5 判断

S5 工作树 `release-judgment-20261015`，标准 `docs/standards/judgment-s5.md`，SOURCE `48abd1fe134b88cfe8122d2aee5a35090fc6d62b` / REPORT `95871f3f3d77d27136e3f3195742e6290c587b4a`。已固定并实测的组件是 Laya SDK + Laya 模型 + Node 24.16.0 + onnxruntime-node 1.30.0 + vscode-jsonrpc 8.2.1 + sourcegraph/jsonrpc2 v0.2.1；Jev 保持云调用。English 与 multilingual 的官方 CPU torch↔ONNX parity 三次通过，Node 与 Go 成熟 RPC2 的常驻、三类题型、长截断与独立重启通过；Jev 真 API、真实 SQLite/HTTP/撤销/缓存/反馈重启与 AT10/AT11/AT12 为 `NOT_RUN`。

### 4.3 现有记录

论文扩展已由 [extensions/paper/THIRD_PARTY.md](../../extensions/paper/THIRD_PARTY.md) 登记 `@steipete/summarize` 0.25.1（MIT）与 Readability 0.6.0（Apache-2.0）的复用来源与许可全文位置。该记录属既有域，不在本次 S4、S5 固定组件范围内。

## 5. 仍未决定与不代替验收

- SemanticIndex 的具体模型、运行时和向量库仍待用户决定。本文不选默认模型或接口；当前 S4 的 CodeSearch 明确报告 `semantic_embedding_not_available`。
- AOCI 的 FSL 包分发未获决定，本轮二进制只在本机工具目录构建，不携带进包。
- Jev 的 `TYPESAFE_API_KEY` 缺失，真实云调用 `NOT_RUN`；OS 完整断网仍 `NOT_RUN`。
- 上述组件均未进入候选 15f；本文只记录许可事实，不构成运行验收或发布放行。

## 6. 交给 P 的配置与打包 Handoff

给 P 的输入：

- 包内已复制项目 `LICENSE`（AGPL-3.0），当前没有第三方许可文本。每个上游许可都要求再分发时保留许可副本与版权声明：MIT 与 Apache-2.0 要求随分发携带许可全文（Apache-2.0 还要求保留源码中的版权、专利、商标与归属声明，若有 `NOTICE` 则须一并保留），FSL-1.1-MIT 要求附带条款副本或链接且不得移除版权声明。因此，凡是把 §3 组件并入最终包的构建，都必须把这些组件的许可与版权文本一并纳入包；采用何种打包方式由 P 与根决定，本文不指定。
- 若最终包包含 Node 运行时，须整体保留 Node 的根 `LICENSE`（含其逐项捆绑第三方）。若最终包不捆绑 Node、只在用户机上按绝对路径配置，则不需要在包内复制 Node 许可。
- S5 运行时组件来源：Node 24.16.0、`onnxruntime-node@1.30.0`、`vscode-jsonrpc@8.2.1`、`sourcegraph/jsonrpc2@v0.2.1`，以及 `laya-ts` 的 Git SOURCE 构建产物与两个 Apache-2.0 模型派生 ONNX 资产。
- S4 运行时组件来源：`@colbymchenry/codegraph@1.0.1`（MIT）与 AOCI `v0.1.0-rc20`（FSL-1.1-MIT），Node 24.16.0 为前置。

配置接口（信任启动参数提供，非本文冻结）：

- S4 构造 `repositorycontext.New(db, db, repositorycontext.Options{...})`，其中 `Graph` 需要 Node 可执行程序、官方 CodeGraph SDK 目录与本轨 `codegraph.mjs`；`AOCI` 需要 AOCI 可执行程序路径；`CacheRoot`、`SnapshotBudget`、`Git.EnvironmentID` 由可信启动配置给出。仓库查询投影不序列化根目录。
- S5 模型资产根为 `D:/Astrocyte-agent-assets/20261015/judgment-s5`（English 与 multilingual 共 12 项，3,017,906,443 字节），进程通过 `ASTROCYTE_S5_ASSET_ROOT` 指向该根。
- Jev 只通过 `POST https://api.typesafe.ai/v1/systemone` 调用，依赖运行环境提供 `TYPESAFE_API_KEY`；它不是可打包的本地模型。

需要根或用户决定：SemanticIndex 组件选择、AOCI 的 FSL 分发方式、Jev 是否作为可选云能力保留、Node 与各 CLI 是随包捆绑还是由用户按绝对路径提供。本文不代选。

## 7. 验证记录

- 许可与元数据原件（14 个文件）保存在本树已忽略 `.verification/components-20261015/upstream/`，逐个记录 SHA-256 前 16 位与字节数。
- 上游核查：Laya 根与 `laya-ts` 的 `LICENSE` @ `432312b3` 均为 Apache-2.0；HF `convaiinnovations/laya` @ `7b928d82` 模型卡为 `apache-2.0`；Node `v24.16.0` 根 `LICENSE` 为 MIT 加逐项第三方；`sourcegraph/jsonrpc2` `v0.2.1` 为 MIT；AOCI `45990a1d` 为 FSL-1.1-MIT；`codegraph` 1.0.1 npm `license: MIT` 且 `gitHead` 为 `a1489f77a6d69538bfe29020b8366ee034f90860`；`onnxruntime-node` 1.30.0 与 `vscode-jsonrpc` 8.2.1 的 npm `license: MIT`；PyPI `laya` 0.4.7 的 `license: Apache-2.0`。
- 版本标签核查：CodeGraph `v1.0.1` 标签指向提交 `a1489f77a6d69538bfe29020b8366ee034f90860`；AOCI `v0.1.0-rc20` 为标签对象 `7c59b0718780f5fccb65d9f0a3445b0b98fea177`；Laya 提交 `432312b36714d8889eba405507b37eda25257d9c`（2026-10-14T05:25:53Z）存在。
- 候选包内容核查：按 `release/Build-Windows.ps1` 与 `docs/release/package-report.md` 逐项判定的 39 项目录，仅含项目 `LICENSE`，不含 §3 任一组件。
- 文档校核：`git diff --check` 与写入范围检查。

未执行：产品构建、Go/前端/浏览器/模型/原生控制器验收、正式采集与发布。本文只做许可与来源核对，不重跑 S4、S5 的领域测试。

## 8. 独立 REPORT

本轮 SOURCE `9be28549ae038f19fe951a249a1b54acf6fa281e` 已普通 push，远端 `git ls-remote origin refs/heads/songconmaisaix31-design/release-research-20261014` 与实际同完整值。本节是该 SOURCE 之后的独立 REPORT，只修改本文件；REPORT 自身的完整 SHA 由实际 Git 提交与 push 回执提供，不在文件内循环引用尚未产生的自身 SHA。

独立核查结果：

- 写入范围：相对基线 `de9ef941adcb550b997b98fa819ba05bcfccfead` 只新增 `docs/release/components-20261015.md`，业务代码、`go.mod`、`go.sum`、`package.json`、`pnpm-lock.yaml`、`dependencies.lock.json` 与冻结域零改动；`.verification/components-20261015/` 已由 `.gitignore` 忽略。
- 上游许可与版本：Laya、Node、jsonrpc2、CodeGraph、AOCI 与 Jev 的许可名称、版权归属和版本标签均按 §3、§7 的官方主链接核对；S4 SOURCE `1a55e163` / REPORT `a393627f` 与 S5 SOURCE `48abd1fe` / REPORT `95871f3f` 均确认位于各自已 push 的远端分支。
- 候选包：15f 的 39 项目录只含项目 AGPL-3.0 `LICENSE`，不含任何 S4、S5 固定组件；本文未据此放行发布。

真实剩余限制：SemanticIndex 选择、AOCI 的 FSL 分发方式、Jev 是否作为可选云能力保留、Node 与各 CLI 的捆绑方式均需根或用户决定；这些组件尚未进入最终包，最终打包方式由 P 与根决定。