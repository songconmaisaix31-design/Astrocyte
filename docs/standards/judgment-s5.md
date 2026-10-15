# S5 判断实施与实际验证

固定工作树 `release-judgment-20261015`，分支 `songconmaisaix31-design/release-judgment-20261015`，初始 SOURCE `da79452d1735f8fceca103aaf9d10bb53182f1f5`。2026-10-16 当前恢复 Task `task_1d3e15a66b2a` / Dispatch `ctx_693200d34a48`，接管 HEAD `b125268d11322c69c652bce677665001bd47e344` 和原 B 普通合并的两件 Taste 冲突。本人停止写入、原 B 临时处理归还及模块独核后，协调者 `msg_5796ede098a5` 放行继续，普通合并为 `158e0a3d4361d410c4d6a8664a48c3abb6741b85`。实际 Orca `terminal show` 为 Codex / GPT-6.1-Sol xhigh fast；原机制、委托、模型与各阶段精确来源见对应 REPORT。

仅修改 `internal/judgment/**`、`internal/adapters/judgment/**`、`internal/adapters/sqlite/judgment_*.go`、`internal/adapters/httpapi/judgment_*.go`、`tests/s5/**` 与本文。Go/Node 共享依赖、迁移归 B；公共 HTTP/OpenAPI/main 装配归 H，Web 与生成客户端归 F。H/B 已本人确认并归还判断 handler 前缀，原 S5 持续负责；新增公共字段和完整 Taste 路由已经 H 本人确认。正式 V4 与认知冻结域保持原授权边界。

## 官方输入与资产

完整读取 [Jev Introduction](https://docs.typesafe.ai/introduction)、[Confidence](https://docs.typesafe.ai/confidence)、[API](https://docs.typesafe.ai/api.md)、[Models](https://docs.typesafe.ai/models.md)、[Quickstart](https://docs.typesafe.ai/introduction/quickstart.md) 与 [Laya README](https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/README.md)。Markdown 页面网页工具首件失败：API/models 不可访问、quickstart `Unsupported content-type: text/markdown`；随后真实 `curl.exe --fail --location` 完整取得官方原件，保存在本树已忽略 `.verification/s5/upstream/`。

Jev 使用 `POST https://api.typesafe.ai/v1/systemone`，固定 `jev-1.13.0`，核实际 response `model` 和 `usage`。Noul 没有 provider confidence，保存 null。Go REST 不自动重试。当前父进程 `TYPESAFE_API_KEY` 存在性 false，真实调用 NOT_RUN；已通过协调消息 `msg_f6860f13bbc2` 请求根安排临时 secret 注入，不读取私人凭据文件。

Laya 源码准确 revision `432312b36714d8889eba405507b37eda25257d9c`，`laya-ts@0.1.0`，Python 准备组件 `laya@0.4.7`，根 LICENSE Apache-2.0。模型 [convaiinnovations/laya](https://huggingface.co/convaiinnovations/laya/tree/7b928d828b7b0e022f929d9bd2e44165aa270148) 固定 revision `7b928d828b7b0e022f929d9bd2e44165aa270148`，公开未 gated，模型卡 Apache-2.0；根模型为 English ModernBERT-large。官方文件列表没有 ONNX。通过固定 revision 下载 safetensors/config/tokenizer，先严格加载检查权重与源码架构，再调用该准确源码的官方 export，不传 `--no-verify`；导出未成功前不登记模型可用。实际 Python/ONNX 与 Node 版本在安装和运行后登记。

运行路径为明确 CPU 的常驻 Node `Agent`，不用 Router 自动选择语言或后端。源码 CPU 错误直接抛出；GPU fallback 分支不会进入。配置必须明确允许语言与题型，不把模型卡作者 benchmark 当本项目准确率。

## 交 B 的迁移设计

以下三表已经由 B 的正式 `024_judgment.sql` 创建；S5 只调用原 migrations 执行器，不修改迁移文件。结果与请求为不可变记录，缓存只引用同一记录；反馈保存实际 human 和固定结果版本。

```sql
CREATE TABLE judgment_results (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    actor_kind TEXT NOT NULL CHECK(actor_kind IN ('human','agent')),
    result_version INTEGER NOT NULL CHECK(result_version = 1),
    cache_key TEXT NOT NULL,
    cacheable INTEGER NOT NULL CHECK(cacheable IN (0,1)),
    request_json TEXT NOT NULL CHECK(json_valid(request_json)),
    result_json TEXT NOT NULL CHECK(json_valid(result_json)),
    created_at TEXT NOT NULL
);
CREATE INDEX judgment_results_cache ON judgment_results(cache_key, cacheable, created_at DESC);
CREATE TABLE judgment_feedback (
    id TEXT PRIMARY KEY,
    result_id TEXT NOT NULL REFERENCES judgment_results(id),
    result_version INTEGER NOT NULL CHECK(result_version = 1),
    actor_id TEXT NOT NULL,
    actor_kind TEXT NOT NULL CHECK(actor_kind = 'human'),
    kind TEXT NOT NULL CHECK(kind IN ('adopt','later','reject','revise','already_solved')),
    reason TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX judgment_feedback_result ON judgment_feedback(result_id, created_at);
CREATE TABLE judgment_policies (
    project_id TEXT NOT NULL,
    question_type TEXT NOT NULL CHECK(question_type IN ('choice','score','truth')),
    version INTEGER NOT NULL CHECK(version >= 1),
    data TEXT NOT NULL CHECK(json_valid(data)),
    PRIMARY KEY(project_id, question_type)
);
```

缓存键只用于 SPEC §11.3 的判断复用，包含问题 schema/候选顺序/实际证据 digest、project/worktree/snapshot、goal/taste/policy/model 版本、access epoch 与 actor。每次命中和反馈前重新核验现行授权；调用开始与结果使用之间版本变化直接拒绝使用。没有未配置阈值的数字默认，不因反馈建立训练或批准。

## 内部契约、策略与 IPC

正式问题 `msg_590b73803414` 已获协调者答复：使用 `sourcegraph/jsonrpc2@v0.2.1` 与 `vscode-jsonrpc@8.2.1`，Content-Length framing 由库负责；不切换 JSON-RPC 版本，不手写流 parser。共享锁由 B 维护。`laya-ts` 的 npm registry 实测 E404；固定官方 Git SOURCE 的 `laya-ts` 子目录真实 build，不能把不存在的发布包填成依赖。

内部 `JudgmentRequest/Result`、题型规则、单后端路由、完整版本键、使用前复核和 SQLite 记录已经实现。`judgment_policies` 保存实际 human 配置，含 monotonic `revision`、固定 project settings revision、worktree/snapshot、每题型单一 backend/model、明确语言、预算及可空阈值。Jev 外发另外逐条核 `external_evidence` 的 fixed material revision/digest；配置 backend 不扩大资料外发许可。判断服务不产生准入或任务批准。

`ProjectAuthorization` 使用既有 `RepositoryProjectReader` 与 `ProjectReferences.ReadSelected`。Agent 每次核实际 project 身份、read_context、revoked/expired；human 才能改策略或反馈。请求不保存 resolved source text，结果不保存原始 provider 回答、prompt 或凭据；产品 usage 只保留规范字段，真实原件仅忽略目录。公共 HTTP 字段/路由仍由 H 确认。

判断适配器通过 S4 `ProjectSnapshotProvider` 捕获并持久化真实 canonical root、worktree、HEAD、index/worktree digest 与 generation。策略保存该完整快照，每次 resolve、模型结果使用和缓存命中均调用现行 S4 provider 复核。SQLite 权限与结果写入共用 `WorkspaceTransaction`；没有本域重复的 Git snapshot 实现。本阶段验证其判断接入，不替代 S4 RepositoryMap/CodeSearch 索引验收。

历史记录读取和反馈单独核当前 caller/project/材料访问与固定结果版本；旧 goal/model/policy 的人工历史反馈可以保留。缓存和后续结果消费仍核全部当前版本，旧版本不会获得执行或准入权限。

## 实际准备环境与首件记录

Python `3.13.13`、torch `2.14.1+cpu`、transformers `5.19.0`、ONNX `1.23.2`、onnxruntime `1.31.0`、onnxscript `0.7.2`、laya `0.4.7`；Node `24.16.0`、TypeScript `7.0.2`、onnxruntime-node `1.30.0`。依赖来自实际安装结果，源码和模型固定为上文精确 revision。

English 权重 strict=True 匹配，官方 export 三次真实 CPU torch/ONNX parity PASS，最大 logits 绝对差为 `1.2e-7 / 2.4e-7 / 1.2e-7`，act probabilities 差均为 0。六个实际文件为 encoder/head ONNX、两份 external data sidecar、tokenizer/config。首次导出在 PyTorch Unicode 状态打印遇 Windows GBK 编码失败，原件 `laya-export-first.log`；UTF-8 复验 `laya-export-utf8.log` 使用同一模型与官方导出，未降低 parity 要求。

直接 SDK CPU 真实三题型/全文截断/候选覆盖检验已执行。首两次 options 前置假设 FAIL 原件保留，候选 ID 先前各自短前缀可区分；真实文档段落的完整标签构成第三次输入，实际 `total=90 / distinct=1 / tokens_per_option=4`。SDK actual model `laya-rl-agent`，local load 的 SDK reported revision 为 null；资产 revision 由严格匹配和已加载文件版本分别保存，不伪造供应商 echo。English 的 `choice:11+` 原配置温度 `0.10058280825614929` 被 SDK 警告为 uncalibrated；产品在推理前拒绝受影响 bucket。实际 SDK 诊断原件 `laya-live-first.log`、`laya-live-options.log`、`laya-live-collapsed.log` 与 `laya-live.json` 在忽略目录。

multilingual 同精确 HF revision 真实下载并 strict=True 匹配，首次官方导出 `OSError: [Errno 28] No space left on device`，原件 `laya-export-multilingual-first.log`。只移除了本树可重新下载的 English 源权重，先保存实际 digest；已通过的 English ONNX、原日志与配置保留。multilingual 官方复验三次 parity PASS，最大 logits 绝对差 `1.4e-6 / 9.5e-7 / 2.1e-6`，act probabilities 差均为 0，实际六项 ONNX 资产已发布到本树忽略目录。multilingual 正式配置的 temperature 三项均为 1，per-bucket 配置为空；实际 Node 推理与语言覆盖仍待验证。

## 验收状态

| 项目 | 当前结果 |
|---|---|
| 初始精确 B 基线与共享合并 | PASS；S4 普通 merge `512fa692585e3f2b6b848802a356705818ec3c7d` 已推送 |
| 官方文档完整读取与原件 | PASS；网页首错误保留 |
| SDK 源码许可与固定模型卡 | PASS；分发仍须保留适用 LICENSE/NOTICE |
| 官方 SDK build | PASS，忽略目录真实 npm build |
| English ONNX 导出与 parity | PASS，官方三次 CPU 数值检查；首编码 FAIL 保留 |
| multilingual ONNX 导出与 parity | PASS，官方三次 CPU 数值检查；首空间 FAIL 保留 |
| SDK 三题型本地 CPU / 长截断 / options不足 | PASS；正常短输入未截断，长输入和候选覆盖损失实际记录；首前置 FAIL 保留 |
| Node/Go 常驻 RPC / 重启 | PASS，English/en、multilingual/en、zh；独占 D 资产实际重载/重启 PASS |
| 断网 | 进程 HTTP 出口拒绝及专用程序 OS 出站 Block 实测 PASS；English 三题型、双 PID 重启和规则/进程/环境清理 PASS |
| Jev 真 API/model/usage | NOT_RUN，缺本进程 secret |
| 真实 SQLite/撤销/缓存/反馈重启 | PASS；原 HTTP 会话、资料导入和判断 handler 实际运行；H 生产组装 NOT_RUN |
| AT10 | 判断接入 PASS：两个实际公开 worktree 与未提交 README 变化隔离；不替代 S4 索引全项 |
| AT11 | 既有真实 Laya 三题型、截断/options、低置信、缺证据、未授权外发、worker 故障不缓存 PASS；完整 Sources 的实际模型输入及专用程序 OS 出站隔离 PASS；Jev 和 H 生产组装按真实剩余分别登记 |
| AT12 | 实际 Goal、完整 Taste 的 rubric/example/delegation、被引用规则和 Ranking 各自变化后，旧 cache/proof 在 Policy 更新前拒绝 PASS；完整 Sources 的真实模型输入、HTTP 往返/独立失效与 SQLite 重启 PASS；H 生产组装 NOT_RUN |

SOURCE 和独立 REPORT 由各阶段普通提交/push 后登记完整 SHA；当前不能宣称 S5 完成。

## 第一阶段独立 REPORT

SOURCE `48abd1fe134b88cfe8122d2aee5a35090fc6d62b` 已普通 push，`git ls-remote origin refs/heads/songconmaisaix31-design/release-judgment-20261015` 实际同完整 SHA。SOURCE 包含 12 个允许路径文件；私有契约、规则、权限、Jev REST 和 SQLite query/INSERT 已编译，Node worker 的 syntax 检查通过；共享依赖、产品 migrations 与公共路由装配未修改。

实际命令：`go test ./internal/judgment ./internal/adapters/judgment` PASS，但当前无 Go 测试文件；`go test ./internal/adapters/sqlite -run '^TestJudgment' -count=1` PASS，仅编译、无匹配测试；`node --check internal/adapters/judgment/laya-worker.mjs` PASS；`git diff --check` 与 `git diff --cached --check` PASS。这些结果不代表运行持久化或完整 S5 验收。

真实资产命令：`.verification/s5/python/Scripts/python.exe tests/s5/prepare_laya.py --source-dir .verification/s5/upstream/laya --output-dir .verification/s5/assets` 在 UTF-8 环境完整 PASS；同命令 `--output-dir .verification/s5/assets-multilingual --subfolder multilingual` 在本树可解释的空间回收后完整 PASS。CPU/长截断/options 的实际 SDK 命令 `node tests/s5/laya_live.mjs` 已真实执行；正常短输入复验、RPC2 与重启、完整 HTTP/SQLite/撤销/全缓存版本、真实 Jev 与 AT10/11/12 按上表继续，不结算整个 Task。

## 常驻 RPC 实现与第二阶段结果

Go `Laya` 使用准确 `sourcegraph/jsonrpc2@v0.2.1`，官方 tag SOURCE `2cc94179e136b565cb9df7d9d0bff9fc06912ac9`，MIT，真实 Go sum `h1:2GtljixMQYUYCmIg7W9aF2dFmniq/mOr2T9tFRh6zSQ=`；Node `vscode-jsonrpc@8.2.1`，MIT。双方分别使用 `VSCodeObjectCodec` 和 `StreamMessageReader/Writer`。CPU checkpoint/语言显式单一配置，startup/shutdown/time/token/byte budget 均显式；未知或超时不会自动建立新 worker。每次实际回复核 provider model、question type、全部概率候选、实际 token usage、截断/候选覆盖与加载后的模型版本。model revision 包含精确 HF/source/checkpoint 与六个实际模型文件的联合版本，涵盖 external data/config/tokenizer；SDK local reported revision 保持 null。

`node tests/s5/rpc_live.mjs` 真实第一轮 PASS：English/en、multilingual/en、multilingual/zh 各三题型、正常短输入未截断、自然长输入截断、常驻 PID 保持与两个独立进程重启，共六个进程退出 0。随后新增 Go 适配器的实际同资产检查 `ASTROCYTE_S5_LIVE=1 go test -modfile=.verification/s5/go.mod -mod=mod ./tests/s5 -run '^TestLayaActual' -count=1 -v` PASS，216.60 秒，三个路由均完整运行。共享 root mod/lock 尚待 B，所以使用忽略目录准备 module 文件引用真实下载的固定官方 Go SDK；root `go.mod/go.sum` 无改动，此检查不表示 root 默认 build 已集成该依赖。

正常短输入的直接 SDK 复验 PASS，实际 usage 为 input 677、output 0、state tokens 187、dropped 0、truncated false；真实全文 state tokens 4023、dropped 3543、truncated true；候选覆盖 total 90、distinct 1。诊断对 English 无效的 choice:11+ 保留原 SDK uncalibrated 警告，产品在该 bucket 推理前拒绝；其他语言/backend 没有 fallback。

## 根授权的独占 D 资产与实际重载

根 `msg_63b51b955ab1` 明确授权唯一 S5 owner 使用 `D:/Astrocyte-agent-assets/20261015/judgment-s5` 作为模型与可重建环境根。模型进程全部退出后检查该根为空且非 reparse；原生 PowerShell 逐 LiteralPath 复制 English 六项和 multilingual 六项，12 项共 `3,017,906,443` bytes，两端实际 SHA256 逐项一致。源码/首 FAIL/输入输出/日志/metadata 保持本树，原件不迁移；没有重跑 export 或更改全局 HOME/TEMP/auth。

任务 process 显式 `ASTROCYTE_S5_ASSET_ROOT=D:/Astrocyte-agent-assets/20261015/judgment-s5`，同一 Go-RPC 三组每组两次真实加载/重启 PASS，92.65 秒，总命令 93.853 秒；全部 owned 模型关闭。随后检查最终绝对路径仍在本树 `.verification/s5`，逐项删除 C 的已验证可重建六+六模型文件，未递归删除，实际释放 `3,017,906,443` bytes；当次 C 可用 `7,389,450,240` bytes。D 六项模型/配置继续完整保留。实际 copy/reload/release 证据为 `.verification/s5/assets-copy-d-safe.json`、`go-laya-d-reload.log`、`assets-d-released-safe.json`。

当前支持语言验收仅 English/en 和 multilingual/en、zh。根已批准先验证进程 HTTP 出口拒绝，实际结果见后文；OS 隔离尚未获执行或提权授权。真实 Jev key、共享合并、H 公共字段/装配和同事务委托消费端口继续按准确交接处理。

## 第二阶段独立 REPORT

SOURCE `30ddaa76b19807e78bf6f56421b4461a6b3cd3f4` 已普通提交/push。本阶段记录 Go/Node 真实成熟 RPC、三题型、两个 checkpoint 与三条明确语言路由的实际重启、真实 Git snapshot 校核和 D 资产重载；没有运行 HTTP/SQLite 完整业务或 Jev 真 API，没有改变公共接口或 S1/S2 批准。

`go test -modfile=.verification/s5/go.mod -mod=mod ./internal/judgment ./internal/adapters/judgment` PASS，私有核心/适配器编译；`go vet -modfile=.verification/s5/go.mod ./internal/judgment ./internal/adapters/judgment ./tests/s5` PASS；真实 model 测试命令和结果见上节。`git diff --check` / `git diff --cached --check` PASS，root `go.mod/go.sum` 零变动。

root 正常入口首检 `go test -mod=readonly ./internal/adapters/judgment` 为真实 FAIL：`github.com/olivere/ndjson@v1.0.1` 要求 `github.com/google/go-cmp@v0.5.6` 的 go.mod sum 缺失。原件 `.verification/s5/root-dependency-first.log`、安全结果 `root-dependency-first-safe.json` 保留；sourcegraph SDK 同时尚待 B root 依赖集成。该结果已准确交 B/协调者，忽略 module 的组件 PASS 不替代 root 默认 build。公共路由、三表生产迁移、同事务委托消费和实际断网按未决/准确来源继续，不降低验收或宣称整个 S5 完成。

## 恢复阶段：进程出口验证与 OS 清理方案

已封存的 `.verification/s5/offline-http-proxy-safe.json` 记录两个真实 CPU 进程 PID `57796`、`23948`，均对固定官方公开模型 URL 取得 `ECONNREFUSED`，随后完成 English 三题型并退出 0。实际回答、usage、stderr 和网络错误各自保留；该范围为两个进程的 HTTP 出口，OS 全部网络隔离保持 `NOT_RUN`。本轮只读取原件，没有重复模型调用或导出。

当前 `tests/s5/os-offline.ps1` 默认仅生成方案。专用程序为本树 `.verification/s5/os-offline/node.exe`，真实 Node `v24.16.0`，SHA256 `B3094D0B49F9AD602262A9921551737BB97637C05DD357A06AE98188D7290AA3`，与 `C:/Program Files/nodejs/node.exe` 相同。规则名称为 `Astrocyte-S5-ctx_0a7797464f02-owned-node-outbound`，创建时附加本次唯一 Group，限制该专用程序的 Outbound / Block / Any profile / Any protocol。共享 Node 和 Orca 不在该程序规则范围。每次方案及执行证据使用独立文件前缀，保留先前原件。

执行分支要求显式 `-Execute` 与已有 Administrator 身份；不启动提权。嵌套 `finally` 分别尝试清理专用路径全部进程、删除本次规则、恢复每个保存的 Process 环境变量。终止前重查 PID、ExecutablePath 和 CreationDate，等待真实退出；删除前重查 Name、唯一 Group、Program、Direction 和 Action；恢复后逐变量读取确认。每个清理阶段发生错误仍进入其余必要清理阶段，最终保存实际状态与错误并抛出所有异常；原执行异常同时保留。未能确认的清理状态不会写成成功。

Windows PowerShell `5.1` 的第一次读取出现真实解析 FAIL，原件 `.verification/s5/os-offline-ctx_0a7797464f02-ps51-first.log` 保留。当前脚本的默认预览已由 `powershell.exe -NoProfile -ExecutionPolicy Bypass -File tests/s5/os-offline.ps1` 实际通过，只复制或核验专用 runtime 并保存方案，没有创建、删除防火墙规则或调用模型。PowerShell 7 parser、Node 两个 offline 脚本语法检查与私域 `git diff --check` 通过；清理执行分支仍未运行，不宣称 OS 验收通过。当前可审阅方案为 `.verification/s5/os-offline-ctx_0a7797464f02-dad09c7b1de149f1827fc985d22db55d-plan.json`。

本人停止写入确认 `msg_12626069e7c3` 已交 M-release，`msg_512dc4296921` 已交 M-product。原 B/S2 按临时范围处理并归还后，M-release `msg_cc73bde2dce8` 明确放行；独立核 `git ls-files -u`、两类 diff-check 均无错误，普通 merge 为 `512fa692585e3f2b6b848802a356705818ec3c7d`，已 push 并核远端同值。缓存维护权限由 `msg_9083f709e503` 撤回，本执行者没有开始 `go clean` 或删除缓存；后续构建按进程使用 D 盘，最新安全检查点改用共享 warm `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build/backend`，`GOTMPDIR=D:/Astrocyte-agent-assets/20261015/go-tmp/judgment-s5` 保持独占。Jev key 与 OS 防火墙/提权仍待根处理，当前 Task 尚未完成。

## 第三阶段独立 REPORT

SOURCE `bc5ffc910b660d6ae9084234ffc8ced8e265cbc5` 已普通提交/push，远端 `refs/heads/songconmaisaix31-design/release-judgment-20261015` 核对为同完整 SHA。该 SOURCE 为 S4 准确快照接入、真实事务复用、权限复核/缓存/反馈与安全方案，共 11 件 S5 允许文件；S4 普通合并独立保留。所有未运行项均保留原验收要求。

实际 `ASTROCYTE_S5_LIVE=1 go test -mod=readonly ./tests/s5 -run '^TestActual' -count=1 -v -timeout 4m` 首轮 PASS，`23.942s`。三个测试使用获准公开 Git 项目的两个真实 worktree、固定官方文档、原 SQLite migrations、原 HTTP human session 与原资料导入服务。已实际核规则结果/cache 命中，goal/taste/policy/model/access 五次持久化策略变更与旧请求拒绝，五类 fixed-version human feedback、错误版本拒绝和关闭/重新打开同一 SQLite，真实 Agent token 撤销后判断/cache/历史拒绝，同 `WorkspaceTransaction` 的真实 SQLite 主键错误完整回滚且没有借用第二个连接，真实未提交 README 变化不影响另一 worktree。模型没有参与这些规则测试，私有作品没有进入输入。

日志为 `.verification/s5/actual-service-ctx_0a-first.log`。适用 `go vet -mod=readonly ./internal/judgment ./internal/adapters/judgment ./tests/s5` PASS，日志 `actual-service-ctx_0a-vet.log`；`gofmt -d tests/s5/service_test.go`、`git diff --check`、`git diff --cached --check` 无输出。此前 root readonly 私域/适配器/fixture 编译 PASS，日志 `root-combination-ctx_0a-first.log`；该编译不计功能测试。

供 H 精确消费的现有构造函数为 `judgment.NewService(Repository, Authorizer, ...Model)`、`judgmentadapter.NewJev(apiKey)`、`judgmentadapter.NewLaya(ctx, LayaConfig)`。`ProjectAuthorization` 使用真实 `RepositoryProjectReader`、`ProjectReferences`、`PolicyRepository` 与 `ProjectSnapshotProvider`；`Snapshot` 返回 `domain.ProjectContextSnapshot`，`SetPolicy` 需要实际 human 与 expected revision。请求、结果、策略、usage、feedback 的准确 JSON tag 均在本 SOURCE 的 `internal/judgment/contracts.go`。Laya 必须提供绝对路径 Node/Worker/ModelDir/AssetsPath/SDKEntry/RPCEntry、明确 Language/ModelRevision/Stderr 和启动/退出预算，不默认选择模型或语言；资产与成熟 SDK 的准确准备路径见前文及真实 `tests/s5/laya_test.go`。

本阶段 SOURCE 没有公共判断 HTTP 路由，没有执行 OS 防火墙、提权或 Jev 真 API，没有授予自动准入。本 Task 继续后续真实 Laya 服务验证与已授权 HTTP 接入。根已通过 `msg_91a31e8a9410` 明确要求按 SPEC 实现版本化 DelegationRule、实际绑定和配置入口；实际用户逐项开启，范围、backend、语言、预算、复核条件及必要阈值均由运行时用户显式配置，没有配置即拒绝。

## DelegationRule 持久化 Handoff

使用现有 `WorkspaceTransaction` 保存实际 human 创建的版本化规则与不可变判断绑定。规则按 project/workflow 唯一，OwnerID 从实际鉴权调用者取得并保持，JSON 保存明确状态、题型/问题、policy revision、固定允许资料、admit 候选或数值范围、外发许可与复核条件。绑定保存实际判断结果、规则版本、workflow、固定 Opportunity revision、owner 和完整来源引用；来源由现行 Attention 读取，HTTP 不接收替代 proof。暂停、撤销、策略/模型/目标/品味/访问变化后，受信任适配器在 B 同事务 Context 中重新核查并拒绝旧授权。`JudgmentResult.decision` 仍只有 `recommend/abstain/need_evidence`，`opportunity_admission` 只作 purpose。

新增迁移只由 B 编写，准确表结构如下。没有 SQL 默认开启值或阈值。

```sql
CREATE TABLE judgment_delegation_rules (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    workflow_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision >= 1),
    state TEXT NOT NULL CHECK(state IN ('disabled','active','paused','revoked')),
    data TEXT NOT NULL CHECK(json_valid(data)),
    UNIQUE(project_id, workflow_id)
);
CREATE TABLE judgment_admission_bindings (
    result_id TEXT PRIMARY KEY REFERENCES judgment_results(id),
    rule_id TEXT NOT NULL REFERENCES judgment_delegation_rules(id),
    rule_revision INTEGER NOT NULL CHECK(rule_revision >= 1),
    project_id TEXT NOT NULL,
    workflow_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    opportunity_id TEXT NOT NULL,
    opportunity_revision INTEGER NOT NULL CHECK(opportunity_revision >= 1),
    data TEXT NOT NULL CHECK(json_valid(data)),
    created_at TEXT NOT NULL
);
CREATE INDEX judgment_admission_bindings_opportunity
    ON judgment_admission_bindings(project_id, opportunity_id, opportunity_revision);
```

## 第四阶段独立 REPORT

机制 SOURCE `d73507a8a26ce49b92ec8504649738f336bcdfd1` 含完整规则 DTO、固定机会 Context、同事务可信 authorizer、已经 H 确认的判断/配置/准入 handler 及真实 Laya 服务测试。第一次普通 push 为实际 `exit 128 / Recv failure: Connection was reset`，保留原工具回执；之后只读确认远端仍为 `5d5922e22ba7d8be3def0f4616a7f576d00162c4`，普通 push 成功，`git ls-remote` 返回准确 d735 全 SHA，模块也已独立核实。没有 force push、变更鉴权或改 TLS。

B 精确两表来源 `491e7f606d8265e9f7e9694f547a91682b305e74` 已普通合为 `c1f5d99c46082dfe6053583faa0f8bfc992a4f5b`，没有冲突。该合并包含十件 B 祖先文件和正式 023/025；B 后续明确确认没有已知必须撤出的不兼容，保留普通历史。新表实际由同一 SQLite 原 migrations 应用，未使用临时 DDL。

`ASTROCYTE_S5_LIVE=1 go test -mod=readonly ./tests/s5 -run '^TestActualDelegationBindingsAndAdmission$' -count=1 -v -timeout 6m` 首轮 `14.984s PASS`，补充人工复核和机会修订后 `13.852s PASS`，最终测试源码为 `1e89ef832b77ae31d75d9db920e7186773b75c94`，已普通 push 且远端全 SHA 一致。原 HTTP human session、固定官方文档、原手工沉淀/Opportunity、原 S1/S2 bridge 均实际运行；没有替代服务或模型回答。验证 disabled 拒绝、逐项开启、同一绑定缓存命中、不同机会和 workflow 隔离、机会 revision 变化失效、暂停旧 proof 拒绝、规则复开产生新绑定、`human_adopt` 缺失拒绝/实际采用后允许/再拒绝后阻止、策略变更失效、过期策略仍能暂停撤销及 terminal revoked。实际 S1 事务保存 delegated admission 和 S2 draft，重复命令保留同一回执，SQL 逐项核 `authority/decision_id/purpose/question_type`。日志 `actual-delegation-ctx_0a-{first,review}.log`。

此前四项真实组件组合 `ASTROCYTE_S5_LIVE=1 go test -mod=readonly ./tests/s5 -run '^TestActual' -count=1 -v -timeout 6m` 为 `290.250s PASS`。Laya 三题型实际 input token 为 `205/220/220`，平台 Noul confidence 为 null；全文实际 input `1024`、丢弃 `3006` state token，90 个公开原文候选的 distinct 为 1。缺证据、外发未授权、真实 worker 关闭后的错误持久化及不缓存均通过；该检查使用实际成熟 SDK / ONNX，未训练权重。原 `actual-model-service-ctx_0a-{first,noul,fixed-space}.log` 首 FAIL 与通过的 `reselected`、`actual-combination-ctx_0a-first.log` 保留。

新增低置信和绑定候选消费测试的首次启动为实际 60 秒 startup 超时，随后扩展场景在原三分钟总测试 Context 中到期；两件原日志为 `actual-model-delegation-ctx_0a-first.log`、`actual-model-sanitized-ctx_0a-first.log`。新增场景使用五分钟测试总 Context，产品每次调用预算、startup 60 秒和 shutdown 10 秒保持原值，最终结果见第五阶段。

HTTP handler 当前有真实 SOURCE 和编译/vet 结果，生产注册、原 guard 白名单与 `cmd/server` factory 继续由 H 交付；本阶段没有将它们登记为完整 HTTP 验收。Jev API secret 与 OS 单程序隔离许可仍待根处理，未调用云端、未创建防火墙或提权。实际 Goal/taste 来源问题已提交协调者：当前五版本测试作用于真实持久化 Policy；Goal 或 profile 在自身服务变更而 Policy 未更新的情形尚未验收。

## 已授权 ONNX 元数据隐私衍生资产

P 原完整包扫描在四图 `onnx.StringStringEntryProto.value` 检出 current_home 类路径。根确认由唯一 S5 模型 owner 用成熟 ONNX/Protobuf 仅处理这些必要元数据；原十二项资产、导出记录及首 FAIL 保留。新目录为 `D:/Astrocyte-agent-assets/20261015/judgment-s5/sanitized-ctx_0a`。English 删除 encoder `1633` / head `155` 项，multilingual 删除 encoder `1289` / head `155` 项，与 P 独立安全计数准确相同；没有输出匹配值。

两套模型的 `onnx.checker.check_model(..., full_check=True)` 均通过；sidecar、tokenizer、config 实际 SHA 与原件相同，所有其余 Protobuf 字段保存/重新读取后序列相同。实际公开 README 的 64/192 token 输入分别通过 CPU encoder 与三题型 head，原件/新件所有输出 tensor `np.array_equal`，各最大绝对差为 0。安全记录 `.verification/s5/onnx-sanitized-ctx_0a-safe.json`，过程日志 `onnx-sanitized-ctx_0a-first.log`。

新 descriptor 明确 `artifact_kind=astrocyte_privacy_sanitized_derivative` 与自身 `artifact_revision=metadata-clean-20261015-ctx_0a7797464f02`，`derived_from` 保留准确原版与原文件 digest；官方 HF/source revision 仅作来源。新字节、有效 revision 与描述均独立，不能当作官方原件。P 使用原 scanner 复验，通过 14 项共 `3,014,649,292` bytes，privacy 为 0；协调者回执 `msg_c891c370dab6`。本项不替代最终集成包的完整验收。

## 第五阶段独立 REPORT

SOURCE `50d9cb299954fa41e381bd5d2d3ce758d00e353b` 已普通提交和 push，远端全 SHA 一致。六件允许文件包含成熟 ONNX 元数据处理与数值等价检查脚本、衍生资产准确版本读取、流式完整性检查、常驻 worker 有界关闭和真实模型消费测试。原始资产和首 FAIL 均保持。

衍生资产实际命令为 `ASTROCYTE_S5_LIVE=1 ASTROCYTE_S5_ASSET_ROOT=<sanitized-ctx_0a> ASTROCYTE_S5_METADATA_ROOT=<sanitized-ctx_0a> go test -mod=readonly ./tests/s5 -run '^TestActualLayaJudgmentServiceSafety$' -count=1 -v -timeout 6m`，`35.412s PASS`，日志 `actual-model-sanitized-ctx_0a-scope.log`。实际返回版本为 `metadata-clean-20261015-ctx_0a7797464f02/7b928d828b7b0e022f929d9bd2e44165aa270148/multilingual/432312b36714d8889eba405507b37eda25257d9c/779825135bb7a2663f6635f75ec7c6a9bc0754633912d0804adbeb01f908c767`。实际模型 confidence 经真实 human 提交显式测试阈值后返回 `low_provider_confidence/abstain`，完整 Context 绑定后的真实 90 个公开原文候选覆盖不足也返回 abstain，两者均被可信准入 authorizer 拒绝。没有产品默认数字阈值。

同资产 `go test -mod=readonly ./tests/s5 -run '^TestLayaActualCPUThreeTypesAndRestart$' -count=1 -v -timeout 8m` 最终 `147.849s PASS`，三条明确路由分别 `43.02/60.08/41.42s`，六次进程启动、三题型、自然长输入截断和显式重启全部通过。日志 `actual-sanitized-restart-ctx_0a-bounded.log`。首轮 `actual-sanitized-restart-ctx_0a-first.log` 的真实启动与退出等待超时保留；关闭路径超出预算时请求终止并明确返回退出未确认，禁止无界等待或把请求终止当作退出成功。

`go vet ./internal/judgment ./internal/adapters/judgment ./internal/adapters/httpapi ./internal/adapters/sqlite ./tests/s5` PASS，`go test -mod=readonly` 同五包 `-run '^$' -count=1` 编译 PASS，Node syntax、gofmt diff、Git diff-check 均通过。上述 Go 命令逐进程使用已授权 D 盘共享 warm GOCACHE 和本任务独占 GOTMPDIR。

AT12 按根 `msg_e1415e3ad447` 继续：固定真实 Goal ID/revision 经 S2 读取固定及当前版本；四维 RankingProfile 使用真实 ID/version/configured/enabled，完整 TasteProfile 的实际 rubric/example/delegation 来源由原 B 补齐；全部实际参与内容进入既有缓存并同事务重核。当前 Policy 五版本测试不算完整 AT12。新依赖 DTO 已交 H 确认，Jev secret、OS 单程序隔离执行许可及 H 生产 HTTP 装配仍待原责任人；没有扩大授权或宣称 S5 完成。

## Goal 与排序来源的真实读取

本次恢复保留三件原 WIP 并完成实际检查。`GoalReader.ReadGoal(context.Context, domain.Caller, string, int) (domain.Goal, error)` 使用原 S2 服务读取明确固定 revision 和 revision=0 的当前版本；项目、内容或 active 状态变化均阻止旧引用使用。`RankingReader.GetRankingProfile(context.Context, attention.Principal) (attention.RankingProfileDetail, error)` 使用原 Attention 服务，未配置、disabled 和旧 ID/version 分别拒绝；没有生成默认权重或完整品味版本。

`ASTROCYTE_S5_LIVE=1 go test -mod=readonly ./tests/s5 -run '^TestActualGoalAndRankingSources$' -count=1 -v -timeout 2m` 实际 PASS，测试 `0.15s` / 包 `2.049s`。真实 HTTP human session、原资料导入、原 SQLite migrations 和两个公开项目均由现有 fixture 建立；Goal 通过原 S2 写入真实 revision 1/2，Ranking 通过原 Attention 写入真实 version 1/2/3。检查固定旧 Goal 内容不可变、跨项目拒绝、当前 Goal 改变使旧引用冲突、排序权重改变使旧 version 冲突和 disabled 拒绝。同一 `WithJudgmentTx` 内读取两个真实来源，没有新增 SQLite 连接等待。原件 `.verification/s5/actual-sources-ctx_8ff-first.log` 保留。

当前两个 reader 尚未进入 JudgmentRequest/Policy 的最终 Sources DTO；完整 TasteProfile reader/DTO 正与 B 核实，新公共字段已交 H 确认，未据候选冻结接口。完整内容进入模型输入、既有缓存及同事务重核和 AT12 仍须后续实现与实际运行。没有重复导出或已通过的模型验证，没有使用模型回答替代本节的领域来源检查。

## 第六阶段独立 REPORT

来源读取 SOURCE `32e701e98604b5c3139c3da61a5cfbe4a63d220e` 已普通提交并 push，`git ls-remote origin refs/heads/songconmaisaix31-design/release-judgment-20261015` 读回同完整 SHA。SOURCE 包含两个原域 adapter、一个真实来源测试和本文，共四件允许路径；gofmt、工作区和 staged diff-check 均通过。本节为随后独立报告提交。

H `msg_17b68cf74737` 确认真实 Goal/Taste 及显式 Ranking 引用的语义符合根已定要求，S5 在原域交精确类型 SOURCE 后由 H 同步公共 OpenAPI。H 同时确认当前没有可交付的 Judgment 生产注册 SOURCE，公共 HTTP 完整验证尚未通过。B 完整 Taste 与 S2 新 Goal 来源待准确提交后消费，Jev secret、OS 管理员执行仍未决定；当前 Task 持续，未结算 S5。

## 判断 HTTP handler 的真实运行

`TestJudgmentActualHTTPRulesFeedbackAndRestart` 启动真实 loopback HTTP listener，注册原 Judgment/Attention/LocalProject handler，使用原 session/CSRF/loopback/request-ID middleware、原 SQLite migrations 和公开 Laya README。测试实际运行导入、快照捕获、策略保存与读取、规则判断、同结果缓存命中、五类 fixed-version 人工反馈、错误版本与 CSRF 拒绝，以及服务/SQLite 关闭后重新打开、重新鉴权并读取完整结果和五项反馈。模型调用为零。该测试的注册由包内测试调用原注册函数；H 的生产 `NewServer` 和 `cmd/server` 组装仍须独立消费与验证。

HTTP `snapshot.root` 保持现有字段，值固定为空字符串。PUT 策略拒绝非空 `snapshot.root`；内部策略从已鉴权的实际注册项目取得 Root，再由原 S4 provider 校验完整固定快照。所有测试中的快照、策略、结果和反馈响应逐字节检查实际项目绝对路径没有进入响应；保存的内部快照继续含真实路径。H `msg_fd4aaf71c092` 已确认该边界和字段兼容处理，准确 SOURCE 后同步 schema。

首轮 handler 实际 `2.12s PASS`，安全快照回传策略复验 `1.84s PASS`；两件原日志 `.verification/s5/actual-http-handler-ctx_8ff-first.log`、`actual-http-safe-snapshot-ctx_8ff-first.log` 保留。后续同源码组合命令 `ASTROCYTE_S5_LIVE=1 go test -mod=readonly ./internal/adapters/httpapi ./tests/s5 -run '^(TestJudgmentActualHTTPRulesFeedbackAndRestart|TestActualGoalAndRankingSources|TestActualDelegationBindingsAndAdmission)$' -count=1 -v -timeout 2m` 全部 PASS，handler `1.93s`、委托 `10.02s`、来源 `0.15s`，日志 `actual-handler-core-ctx_8ff-first.log`。这些规则结果与原完整品味来源的 AT12 分别登记；完整 Sources 接入继续。

## 完整 Sources、领域来源与实际失效检查

HTTP handler SOURCE `c26c3873f216d75f06c32dca576dab263c3aed68` 已普通 push 并核远端同值；包含 PUT 策略非空 Root 拒绝的最终 handler 复验 `2.63s PASS`，原件 `actual-http-root-reject-ctx_8ff-first.log`。其后精确消费 S2 SOURCE `ee7889cf04383162e160ae30544eb324257ed45c` 为本树 `7d83003`，唯一 `tests/s2/approval_test.go` 冲突由原 S2 在模块批准的单文件范围编辑/stage；其本人停止写入归还、模块独核 U0 后，`msg_3f895002c9db` 放行本人唯一 continue。B SOURCE `41602e6f2a14a75e3765ef0a72f7afadf1856bbb` 消费为 `03e2e51`，新增正式 026 两表与完整 TasteProfile 的领域/存储。没有改写 B/S2 源码、依赖或迁移。

请求、策略与结果增加可空 `sources`：明确 Goal `{id, revision}`、完整 Taste `{id, version}`，以及仅在参与判断时显式提供的 Ranking `{id, version}`。`GoalReader` 使用 S2 的固定历史读取及 `CurrentGoalReader.ReadCurrentGoalSnapshot`，核当前 active Goal 和实际项目 settings revision。`TasteReader.ReadTasteProfile` 读取 B 的固定与当前不可变版本，核 configured/enabled、实际 rubric、固定示例资料和 delegation 引用；当前判断只接受 `choice/score/truth`。原 Ranking reader 使用真实权重及 enabled/version。未配置完整 Goal/Taste 的策略保留人工判断入口，active Delegation 的配置与使用均拒绝。

完整 Goal、Taste JSON、实际 Ranking、S5 权威规则内容和有权限的固定示例正文进入原模型 state、input_digest 与缓存键。没有保存第二份委托权威；`ReadDelegationSource` 从原 S5 表读取当前固定 rule/workflow/revision 和实际 owner/project，disabled/paused 状态作为来源事实保留，读取来源不会产生准入权限。自动使用继续逐项检查绑定规则的当前 active 状态及原授权。所有缓存命中、结果保存和准入在原 WorkspaceTransaction 中重新读取完整内容；使用已有记录时重新计算实际 input_digest，内容不同即拒绝。

Jev 的 `external_sources` 必须与所选完整 Sources 引用逐项相等，固定示例还必须逐条列入 `external_evidence`。模型 backend 配置没有附带 Goal、品味或示例的外发许可。历史结果和人工反馈允许保留原 Goal/Taste 版本，同时重新验证固定 Goal 的项目权限与当时示例的当前资料权限；示例撤回后，缓存、准入、历史读取和新增反馈均拒绝该资料。

真实新 `TestActualCompleteSourcesInvalidateCacheAndAdmission` 依次通过原领域服务更新 Goal、rubric、固定示例、Taste 的 delegation 选择、被引用规则状态和 Ranking 权重。每项更新均在 Policy 版本保持原值时检验旧请求与已绑定准入拒绝，再由真实 human 配置当前版本并检验新的 input_digest 和有效绑定。同一 SQLite 事务授权没有增加连接等待。实际 Goal 最终 revision=2，Taste version=5，Ranking version=2，被引用规则 revision=2；最终撤回示例后四种使用全部拒绝。模型调用为零。

首次完整新测试三分钟期限到达时正在真实 S4 Git 快照校验，`actual-complete-sources-ctx_8ff-first.log` 保留全部 panic/堆栈，未计通过。当前来源错误先于昂贵的 Git 快照检查立即失败；新测试只保留必要的初次同源缓存命中，六组新增场景的测试总 Context 为六分钟，产品预算、S4 快照检查和全部独立失效断言保持。该场景复验 `240.67s PASS`，日志 `actual-complete-sources-ctx_8ff-recheck.log`。

增加历史示例权限/撤回和外发许可检查后的五项组合实际命令为 `ASTROCYTE_S5_LIVE=1 go test -mod=readonly ./tests/s5 -run '^TestActual(CompleteSourcesInvalidateCacheAndAdmission|AutomaticSourcesAndExternalPermission|DelegationBindingsAndAdmission|GoalAndRankingSources|DelegationSourceSnapshot)$' -count=1 -v -timeout 8m`，完整 `389.387s PASS`，日志 `actual-sources-final-ctx_8ff-first.log`。原 Delegation 全链 `120.97s`、完整来源 `217.81s`、未配置自动使用与独立外发许可 `19.76s`、Goal/Ranking `1.72s`、规则来源 `11.42s`，零 skip。该结果属于规则、领域服务与真实 SQL 来源；新 Sources 进入真实 Laya 模型的必要检查和 H 生产组装分别验收。

完整 Taste 的有限选项、评分等级和实际 human expected 对应已向 B 交具体方案 `msg_f8ce1c42a2fa` / `msg_f12f0dc54eed`，准确后继源码待原作者交付；既有存储接受的任意题型不会直接进入 S5 模型。完整 Taste 的 HTTP 人工配置入口已获 H `msg_f17902dcd7ec` 核定。Jev secret、OS 隔离执行仍未得到所需输入/许可。两次普通 push 依赖消费提交发生真实 `OpenSSL SSL_connect: SSL_ERROR_SYSCALL`；随后公开 GitHub HTTPS 的 curl 也握手失败，远端当时仍为 c26 全 SHA，未将本地提交登记为已推送。

## 当前 Sources 输入、HTTP 与验收交接

接续原 `TestActualLayaCompleteSourceInput` 与实际原件：`.verification/s5/actual-laya-complete-sources-ctx_8ff-lifecycle.log` 为测试 `8.40s` / 包 `10.290s PASS`。同一真实 CPU 调用包含完整 Goal、Taste 与固定示例，实际 `input_tokens=981 / state_tokens=942 / truncated=false / confidence=null / decision=abstain`，同源再次请求命中 SQLite 缓存，原领域服务变更 Goal 后拒绝旧请求。`t.Cleanup(cancel)` 注册在模型清理之前，真实 `Close` 完成后才取消 Context；成功原件对应当前代码。本次恢复沿用该成功推理，模型调用为零。原 `first` 的启动/关闭超时和 `quiet` 的实际推理后退出 1 原件均保留。

判断请求中的过期 Policy/快照版本在实际授权读取处立即返回 `version_conflict`，随后当前来源与 S4 快照仍逐一核验。原 90 秒规则测试期限保持；五版本更新、固定结果与五类反馈关闭重开后全部通过，重开阶段 `0.01s`。首回归的 `provider_unavailable` 原件 `.verification/s5/actual-sources-http-regression-ctx_8ff-first.log` 保留。

本次 HTTP 使用原 SQLite migrations、loopback listener、session/CSRF 和原领域服务，实际保存完整 Taste，创建固定 Goal，通过 Policy/Request/Result 的 `sources` 往返、同源缓存、五类反馈及 SQLite/HTTP 重启。实际 Taste 更新独立返回 `version_conflict`，实际 Goal 更新独立返回 `context_stale`；旧人工历史仍能按当前资料权限读取，全部响应继续检查项目绝对路径没有泄漏。H 的 `NewServer` 与 `cmd/server` 生产组装单独验收。

H 本人 `msg_f17902dcd7ec` 经协调者 `msg_e30c8662ea05` 确认 `GET/PUT /api/v1/attention-taste-profile` 与原 ranking 路径并存，采用原 `TasteProfileDetail` 和 `UpdateTasteProfileCommand` 的 session/CSRF/Idempotency-Key/CAS 语义；`JudgmentServices.Taste` 使用原 `attention.TasteProfileService`。核心提交为 `cabeb3bdce74ba52b958f2c6835ee324550766ea`；同阶段 HTTP 源码另有普通提交，完整来源在随后独立 REPORT 登记。

新增运行入口 `tests/s5/verify-sources.ps1` 显式使用本任务独占 `D:/Astrocyte-agent-assets/20261015/judgment-s5/verification-ctx_fc76cd365979`，其 `runs` 保存真实 SQLite/objects/projections，`temp` 同时供 `GOTMPDIR/TEMP/TMP` 使用；共享 warm `GOCACHE=D:/Astrocyte-agent-assets/20261015/go-build`。Go 子进程 stdout/stderr 分别原样保存，非零退出立即失败，`finally` 恢复 Process 环境。S5 模型测试的 stderr 同样要求显式 `ASTROCYTE_S5_RUN_ROOT`。

实际命令均在当前固定工作树执行：

| 命令 | 结果与准确原件 |
|---|---|
| `powershell -NoProfile -ExecutionPolicy Bypass -File tests/s5/verify-sources.ps1 -EvidenceRoot D:/Astrocyte-agent-assets/20261015/judgment-s5/verification-ctx_fc76cd365979 -Phase tests` | 八项实际非模型场景全部 PASS、零 skip，包 `41.260s`；`tests-20261015T1605544395993.stdout.log`，stderr 空 |
| 同入口 `-Phase http` | 判断 handler 完整 Sources、明确来源错误码、缓存/反馈/重启全部 PASS，测试 `4.43s` / 包 `6.415s`；`http-20261015T1607008656898.stdout.log`，stderr 空 |
| 同入口 `-Phase checks` | 五个适用包 `go vet -mod=readonly` 与 `node --check internal/adapters/judgment/laya-worker.mjs` EXIT 0；`checks-20261015T1607407049704.{stdout,stderr}.log` 均空 |

本次验证入口首次使用的 .NET 方法在 Windows PowerShell 5.1 不可用，真实工具回执保留；当前入口使用其支持的标准路径方法。首次组合另有实际 unused-import 编译失败，完整原件为 `tests-20261015T1605127891197.{stdout,stderr}.log`；当前八项复验通过。测试全部成功断言、原 90 秒 Context、产品 60 秒 startup 和 10 秒 shutdown 预算保持原值。

交 H 的装配输入为 `ProjectAuthorization.Goals` 接实际 S2 `ReadGoal/ReadCurrentGoalSnapshot`，`Taste/Ranking` 接实际 Attention service，`Rules` 接同一 DB；`JudgmentServices.Taste` 使用该 Attention service。公共 `sources` 含固定 `goal{id,revision}`、`taste{id,version}` 和可选 `ranking{id,version}`，Policy 的 `external_sources` 为独立显式外发许可。现有 JSON 列保存这些字段，本阶段没有新增依赖或 DDL 请求。

真实剩余为 H 的 OpenAPI 与生产组装、B 的完整 Taste 有限候选/评分等级后继、实际 Jev 回执和 OS 全断网。当前父进程 `TYPESAFE_API_KEY` 存在性仍为 false；Jev 为 NOT_RUN。OS 管理员执行尚未授权，既有进程 HTTP 出口拒绝范围单独保留。S5 完整验收与最终组合/发布尚未完成。

## 第七阶段独立 REPORT

完整 Sources 与已确认 HTTP SOURCE `688587ffcccafaf80ffe420e08cdaf5848eb36b1` 已普通 push，实际 `git ls-remote origin refs/heads/songconmaisaix31-design/release-judgment-20261015` 返回同完整 SHA。其前一核心提交为 `cabeb3bdce74ba52b958f2c6835ee324550766ea`；相对本次接管 `03e2e511de290454d3f21a7803f291bbfe7614a7` 共十八件允许文件。本文 REPORT 单独提交，原固定工作树的业务源码已经没有未提交修改。

核心准确源码使用 `git archive --format=tar --output=<D 盘证据根>/core-cabeb3b.tar cabeb3bdce74ba52b958f2c6835ee324550766ea`，由标准 tar 解压到同证据根 `core-cabeb3b`。只补固定官方 README/API 原件副本及两份 Agent 创建的公开 Astrocyte Git clone；该独立目录没有候选 HTTP 工作区修改。首次 `git clone --local` 跨 C/D 卷创建默认 hardlink 实际失败 `Improper link`，原工具回执保留；标准 `git clone --local --no-hardlinks` 实际复制成功，两份公开输入均为 `da79452d1735f8fceca103aaf9d10bb53182f1f5`。

准确 archive 中执行 `powershell -NoProfile -ExecutionPolicy Bypass -File <archive>/tests/s5/verify-sources.ps1 -EvidenceRoot D:/Astrocyte-agent-assets/20261015/judgment-s5/verification-ctx_fc76cd365979/exact-core-cabeb3b -Phase tests`，八项全部 PASS、零 skip，包 `328.320s`。真实委托 `92.63s`、规则五版本与反馈重启 `51.48s`、撤销与事务回滚 `21.76s`、双公开 worktree `27.31s`、完整 Sources 独立失效 `111.96s`、自动使用/外发许可 `10.55s`、Goal/Ranking reader `0.54s`、委托来源 `4.78s`。原件在该 `exact-core-cabeb3b` 目录 `tests-20261015T1611453019333.{stdout,stderr}.log`，stderr 空，原期限和预算保持。

完整 HTTP 源码沿用上节实际 `6.415s PASS`，核心工作区八项 `41.260s PASS`、五包 vet、Node syntax、gofmt 与两种 Git diff-check 各按原准确范围保留。独立 archive 的成功单独证明分拆核心可以实际编译运行，没有将 HTTP 工作区结果替代该检查。本轮模型调用为零，既有完整 Sources 的真实单次 truth 推理按原件保留；新增完整 Sources 与 choice/score 的组合、Sources 绑定后的低置信/候选覆盖场景尚未重新运行，前阶段实际模型结果保持各自范围。

H 本人已确认字段、路由和原服务语义，准确源码与 `registerJudgment(*http.ServeMux, JudgmentServices)` / `registerJudgmentTaste(*http.ServeMux, JudgmentServices)` 已交协调者供 H 消费；生产注册与 OpenAPI 仍由 H 完成。B 的完整 Taste 后继、Jev secret 与 OS 执行许可仍需对应实际输入，整个 S5 Task 持续。

## 完整 Taste 的有限题型配置

原 B 精确 SOURCE `6d1d59bf07b008a0969efefceb911e49b6b33db6` 已经通过上述普通合并消费，026 两表、不可变完整 Taste 内容及严格题型校验保持原领域实现。S5 的实际来源准备显式映射 choice 的有限 Options 和 score 的有序 Levels，truth 明确提供两个空数组；所有 expected 由实际 human 提交，choice 对应候选 ID，score 和 truth 使用原 JSON 数值范围校验。没有新增依赖或 DDL 请求。

真实 HTTP 保存 truth、choice、score 三项 rubric 及其固定公开原文示例，拒绝布尔 truth、未知 choice、越界 score 和 null score 后，实际版本与配置保持不变。完整候选、评分等级、示例、Goal/Taste 来源、同源缓存及五类反馈在 SQLite/HTTP 关闭重开后逐项核查。命令 `powershell -NoProfile -ExecutionPolicy Bypass -File tests/s5/verify-sources.ps1 -EvidenceRoot D:/Astrocyte-agent-assets/20261015/judgment-s5/verification-ctx_693200d34a48 -Phase http` 实际测试 `23.27s` / 包 `37.014s PASS`，原件 `http-20261015T1646564925526.{stdout,stderr}.log`，stderr 空。合并后的首次实际 PUT 400 / 包 `2.022s FAIL` 保存在 `http-20261015T1645356259282.stdout.log`。

同一入口 `-Phase tests` 八项真实非模型场景全部 PASS、零 skip，包 `169.713s`。完整 Sources 场景 `21.09s PASS`，在实际配置中加入有限 choice/score 及固定 human expected，再独立更新 choice 候选描述和 score 等级；每次均在 Policy 更新前拒绝旧缓存和旧准入依据，实际 human 更新引用后产生新的 input_digest。Goal、rubric、示例、delegation、被引用规则、Ranking 及示例撤回检查同时通过，最终 Goal revision=2、Taste version=8、Ranking version=2、来源规则 revision=2。原委托 `59.86s`、五版本与反馈重启 `38.07s`、撤销/回滚 `17.35s`、双公开 worktree `23.42s`、自动使用与外发许可 `1.79s`、Goal/Ranking reader `0.14s`、委托来源 `0.86s`，原件 `tests-20261015T1648346546164.{stdout,stderr}.log`，stderr 空。

同一入口 `-Phase checks` 五个适用包 `go vet -mod=readonly` 与 Node worker syntax 均 EXIT 0，`checks-20261015T1651589199092.{stdout,stderr}.log` 均空。`gofmt -l` 对两件 Go 修改无输出，普通与 staged Git diff-check 通过。本阶段只修改两件 S5 真实测试和本文；公共装配、共享依赖、迁移及其他领域由各原负责人保持。

本轮沿用既有完整 Sources 的真实 Laya 推理与成功 Close 原件，没有重复模型调用。H 的生产注册与 OpenAPI、实际 Jev API 回执和 OS 全断网仍按各自实际验收要求保留，当前没有宣称完整 S5 通过。

## 第八阶段独立 REPORT

SOURCE `f955fe3c3450200342e8cbf0b0c152e8065a882e` 已普通提交并 push，独立 `git ls-remote origin refs/heads/songconmaisaix31-design/release-judgment-20261015` 返回同完整 SHA，工作区无修改。原 B SOURCE `6d1d59bf07b008a0969efefceb911e49b6b33db6` 经普通 merge `158e0a3d4361d410c4d6a8664a48c3abb6741b85` 保持祖先关系。本人本阶段仅修改 `tests/s5/source_policy_test.go`、`internal/adapters/httpapi/judgment_http_test.go` 和本文；B 两件临时冲突域在归还后结束，没有越域修改领域实现。

准确源码对应上述三题型真实 HTTP `37.014s PASS`、八项真实非模型场景 `169.713s PASS / 零 skip`、五包 vet、Node syntax、gofmt 和两类 diff-check。当前 Task 没有增加模型调用，已有固定模型、CPU 推理、usage 与 Close 原件继续按其准确范围保留。合并后首 HTTP 400 / FAIL 原件独立保存，完整默认产品检查、H 生产注册、Jev 和 OS 隔离不计为本阶段通过。

准确源码和装配签名已通过 `msg_f112b8e776ad` 交协调者供 H/I/T 消费；当前 Process 的 `TYPESAFE_API_KEY` 和 `ASTROCYTE_LAYA_MODEL_DIR` 存在性均为 false，没有读取凭据文件。Jev 临时 secret、OS 执行许可及 H 生产源码仍为后续实际输入，完整 S5 Task 持续。本 REPORT 作为随后单文档普通提交。

## 已授权的专用程序 OS 出站隔离

根 `msg_063936932ea3` 转达用户授权，协调者 `msg_b17c189f1a09` 明确本 owner 单次运行原 `tests/s5/os-offline.ps1 -Execute`，允许用户确认 UAC。`msg_7ae67fbed787` 确认在原 Node 验证程序补启动/调用 60 秒、关闭 10 秒的分阶段计时检查；原 PowerShell 参数、规则作用、固定模型/公开输入和每轮 120 秒整体期限保持。没有新增普通模型预检或第二次 Execute。

实际唯一命令为 `Start-Process -FilePath powershell.exe -ArgumentList @('-NoProfile','-ExecutionPolicy','Bypass','-File','"C:\Users\DW\orca\workspaces\Astrocyte\release-judgment-20261015\tests\s5\os-offline.ps1"','-Execute') -Verb RunAs -WindowStyle Hidden -PassThru`，随后等待同一进程退出。唯一管理员 PID `54324`，UTC `2026-10-15T17:04:14.8945115Z` 至 `17:04:49.7144505Z`，实际 EXIT 0。原专用程序为 `.verification/s5/os-offline/node.exe`，`92,279,112 bytes`、Node `v24.16.0`，SHA256 `B3094D0B49F9AD602262A9921551737BB97637C05DD357A06AE98188D7290AA3` 与配置的 Node 一致。执行前独占目录非 reparse、专用进程和同名规则均为零，三个防火墙 profile 已启用。

沿用已审阅的规则 Name `Astrocyte-S5-ctx_0a7797464f02-owned-node-outbound`，本次唯一 Group `Astrocyte-S5-ctx_0a7797464f02-ea5eba8bfaf24c769e033503a2071a93`，Program 始终为该专用 exe。实际规则/Program 读回与真实网络错误原件保留；两个 CPU PID `47140`、`53160` 的固定官方 HTTP 请求均真实返回 `EACCES`。两轮各完成 choice/truth/score，实际 input_tokens 分别 `212/224/224`，全部未截断、provider model 为 `laya-rl-agent`、SDK revision 为 null，资产有效 revision 为 `7b928d828b7b0e022f929d9bd2e44165aa270148/english/432312b36714d8889eba405507b37eda25257d9c/196b372122a5fec9c701abe4de8660e86c7a184fc1d40ff2e1f8b6a7b58a9e05`。

启动实际 `7470.7873/7632.2908 ms`，关闭 `95.8965/126.0599 ms`；六次推理均在 60 秒内，两个不同 PID 均 EXIT 0。`.verification/s5/offline-os-program-block-{1,2}.json` 保存实际回答/usage/时间，`offline-os-program-block-safe.json` 保存该专用程序 OS 出站检查 PASS。原 PowerShell 运行的 plan/rule/first/cleanup 文件共同前缀为 `os-offline-ctx_0a7797464f02-ea5eba8bfaf24c769e033503a2071a93`；清理原件记录准确规则已删除、专用进程已退出、全部 Process 环境已恢复，cleanup_failures 为空。本人在原进程结束后独立查询准确 Name/Group 规则数、专用程序进程数及管理员 PID 数均为零，三个 profile 仍启用。本阶段新增证据为这些小型本树记录，模型继续只读使用原获准 D 盘资产。

Jev 安全本机注入入口按根 `msg_5ec15ce2eb5a` 仅准备，实际 key 调用仍未授权/未执行；H 生产装配准确源码仍在途。完整 S5 不因本次 OS 检查通过而结算成功。

## 第九阶段独立 REPORT

OS SOURCE `ef20179eb60d986a3d4d93a5232cffa0970382fd` 已普通 push，独立 `git ls-remote` 返回同完整 SHA；该提交仅含 `tests/s5/offline-live.mjs` 与本文。原 PowerShell 文件和准确命令参数保持原值，Node syntax 与两类 diff-check 通过，实际单次管理员进程 EXIT 0 和上述两轮 EACCES / CPU 三题型 / 重启 / 分阶段预算 / 清理原件共同构成本阶段验收；没有另行重复 Execute 或成功模型预检。

本阶段实际结果已通过 `msg_75dcca8abe7d` 交协调者。追加独立查询准确 Name/Group 规则数、专用程序进程数、原管理员 PID 数均为零，全部 profile 仍启用；Process 环境在原脚本内逐项恢复并读回。通用网络修复已由用户取消，本执行者没有开展该事项、变更网络/代理或全局防火墙。Jev 本机安全入口继续准备，实际云端回执与 H 生产组合仍未通过，完整 Task 保持持续。

## Jev 本机安全输入与固定公开任务入口

根 `msg_5ec15ce2eb5a` 经协调者 `msg_585470199343` 授权准备可审阅入口，实际云端调用仍待用户明确授权和本机输入。`tests/s5/verify-jev.ps1` 默认 `Preview`，`Prepare` 只核官方公开 source revision/README 内容并编译真实 Go 验收程序，均发送零 provider 请求。手动验收文件使用显式 `astrocyte_jev_live` build constraint，Prepare/Execute 编译参数明确选择该标记，默认 Go 测试不包含该文件，没有新增 skip；手动执行缺少授权标记或 key 时立即失败。`Execute` 支持本机 `Read-Host -AsSecureString`，或显式使用既有 Process 的 `TYPESAFE_API_KEY`；两条路径均只向固定任务测试子进程传入同名环境变量，编译/Git 子进程移除该变量，调用方 Process 环境保持原值。

固定任务使用原官方 Laya `432312b36714d8889eba405507b37eda25257d9c` 的完整 `laya-ts/README.md`，语言 `en`，依次提问导出格式 choice、数值验证说明 score 与默认 torch/ONNX 校验 truth。仅请求 `POST https://api.typesafe.ai/v1/systemone`、固定 `jev-1.13.0`，至多三次请求、每次 60 秒，完整输入上限 131072 bytes / 8192 tokens、输出上限 256 tokens，没有自动重试。首失败立即终止剩余题目，现有证据目录拒绝覆盖。

`internal/adapters/judgment/jev_live_test.go` 使用生产 Jev 适配器和实际标准 HTTP transport，保存真实公开请求正文、原始响应正文、HTTP 状态及规范化 model/usage/nullable confidence。记录器不保存请求头；写入回执及打印/保存子进程 stdout/stderr 前检查密钥，检测到密钥即拒绝输出或持久化。密钥通过子进程环境传递，未进入命令行、任务计划、Git 或证据；本机安全输入不要求发送私密值到协调消息或读取凭据文件。

已实际运行 `powershell -NoProfile -ExecutionPolicy Bypass -File tests/s5/verify-jev.ps1 -EvidenceRoot D:/Astrocyte-agent-assets/20261015/judgment-s5/jev-prepare-ctx_693200d34a48-003 -Phase Prepare`，EXIT 0、provider_requests=0，官方 source revision 和 README 一致，`go test -c -tags=astrocyte_jev_live -trimpath -mod=readonly` 真实编译成功。该目录的 compile stdout/stderr 与 source-content stdout/stderr 均空，source-revision stdout 为固定全 SHA；先前 `-001/-002` 准备原件保留。实际 `go list -json` 核默认文件集合排除 `jev_live_test.go`，带该标记时进入 TestGoFiles；`go vet -tags=astrocyte_jev_live -mod=readonly`、gofmt 和 diff-check 通过。默认 Preview 与 PowerShell parser 已通过；没有运行 `Execute`、输入实际 key 或调用 Jev。原已结束的 OS 验收保持其准确结果，后续日志的 nullable 字段使用显式空值变量，原件保持。

根确认本次三个公开请求并由用户在本机输入后，准确命令为：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "C:\Users\DW\orca\workspaces\Astrocyte\release-judgment-20261015\tests\s5\verify-jev.ps1" -EvidenceRoot "D:\Astrocyte-agent-assets\20261015\judgment-s5\jev-public-authorized-001" -Phase Execute -PromptForKey
```

该命令在准备完成后显示安全输入提示，API key 由用户直接输入本机终端。证据根必须是全新目录；执行失败保留原件并交根处理，不自动重复调用。真实 Jev 三题型 model/usage 回执、安全输入全流程及 H 生产装配仍为 NOT_RUN，准备通过不替代这些验收。

## 第十阶段独立 REPORT

Jev 入口 SOURCE `6c7b37a670830f82f35bcd85a13b7fa79758ba19` 已普通 push，独立 `git ls-remote` 返回同完整 SHA。该源码包含本机安全输入脚本、真实标准 HTTP 验收测试、后续 nullable 日志输出及本文，共四件允许路径；没有修改共享依赖、迁移、公共契约或生产装配。准确源码的 Preview / Prepare / Go 编译 / adapter vet / PowerShell parser / gofmt / 两类 diff-check 均通过，Prepare 原件在上述 `-002` 目录，实际 provider 请求数为零。

完整可审阅入口、固定公开任务、准确命令和待用户完成的本机操作已通过 `msg_177fbe489bc0` 交协调者供根处理。用户确认后在本机独立 PowerShell 输入密钥，不向 Agent 消息或命令行参数发送密钥；本执行者没有运行 `Execute`、请求用户在本窗口输入或读取私密凭据。OS 验收由根 `msg_9506521581ee` 接纳并关闭，原件保持，没有新增 OS 检查或模型调用。当前继续原 Dispatch 等待 H 生产准确 SOURCE 和 Jev 后继确认，完整 S5 尚未通过。
