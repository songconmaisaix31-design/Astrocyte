# S5 判断实施与实际验证

固定工作树 `release-judgment-20261015`，分支 `songconmaisaix31-design/release-judgment-20261015`，接管 SOURCE `da79452d1735f8fceca103aaf9d10bb53182f1f5`，接管工作树 clean。实际 Orca 终端为 Codex / GPT-6.1-Sol high fast；Task `task_1d3e15a66b2a` / Dispatch `ctx_f403dc7813bc`。

仅修改 `internal/judgment/**`、`internal/adapters/judgment/**`、`internal/adapters/sqlite/judgment_*.go`、`internal/adapters/httpapi/judgment_*.go`、`tests/s5/**` 与本文。Go/Node 共享依赖、迁移交 B；公共 HTTP/OpenAPI/main 装配交 H，Web 与生成客户端交 F。S2 当前未交稳定权限端口，相关集成仍未决定；正式 V4 与认知冻结域保持原授权边界。

## 官方输入与资产

完整读取 [Jev Introduction](https://docs.typesafe.ai/introduction)、[Confidence](https://docs.typesafe.ai/confidence)、[API](https://docs.typesafe.ai/api.md)、[Models](https://docs.typesafe.ai/models.md)、[Quickstart](https://docs.typesafe.ai/introduction/quickstart.md) 与 [Laya README](https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/README.md)。Markdown 页面网页工具首件失败：API/models 不可访问、quickstart `Unsupported content-type: text/markdown`；随后真实 `curl.exe --fail --location` 完整取得官方原件，保存在本树已忽略 `.verification/s5/upstream/`。

Jev 使用 `POST https://api.typesafe.ai/v1/systemone`，固定 `jev-1.13.0`，核实际 response `model` 和 `usage`。Noul 没有 provider confidence，保存 null。Go REST 不自动重试。当前父进程 `TYPESAFE_API_KEY` 存在性 false，真实调用 NOT_RUN；已通过协调消息 `msg_f6860f13bbc2` 请求根安排临时 secret 注入，不读取私人凭据文件。

Laya 源码准确 revision `432312b36714d8889eba405507b37eda25257d9c`，`laya-ts@0.1.0`，Python 准备组件 `laya@0.4.7`，根 LICENSE Apache-2.0。模型 [convaiinnovations/laya](https://huggingface.co/convaiinnovations/laya/tree/7b928d828b7b0e022f929d9bd2e44165aa270148) 固定 revision `7b928d828b7b0e022f929d9bd2e44165aa270148`，公开未 gated，模型卡 Apache-2.0；根模型为 English ModernBERT-large。官方文件列表没有 ONNX。通过固定 revision 下载 safetensors/config/tokenizer，先严格加载检查权重与源码架构，再调用该准确源码的官方 export，不传 `--no-verify`；导出未成功前不登记模型可用。实际 Python/ONNX 与 Node 版本在安装和运行后登记。

运行路径为明确 CPU 的常驻 Node `Agent`，不用 Router 自动选择语言或后端。源码 CPU 错误直接抛出；GPU fallback 分支不会进入。配置必须明确允许语言与题型，不把模型卡作者 benchmark 当本项目准确率。

## 交 B 的迁移设计

以下 DDL 是待 B 安排的迁移内容；本文不执行产品启动迁移、不修改迁移文件。结果与请求为不可变记录，缓存只引用同一记录；反馈保存实际 human 和固定结果版本。

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

`ProjectAuthorization` 使用既有 `LocalProjectRepository` 与 `ProjectReferences.ReadSelected`。Agent 每次核实际 project 身份、read_context、revoked/expired；human 才能改策略或反馈。请求不保存 resolved source text，结果不保存原始 provider 回答、prompt 或凭据；产品 usage 只保留规范字段，真实原件仅忽略目录。S2 Grant WIP 不消费。公共 HTTP 字段/路由仍由 H 确认，当前只有私有契约。

判断适配器读取实际 Git HEAD 和既有 `ProjectFiles` 权限过滤的明确 `snapshot_files`，保存实际文件内容 digest 及 mtime/size 版本。worktree 身份绑定真实 canonical root，snapshot 绑定完整原始 HEAD 输出和选定文件版本；每次 resolve、模型结果使用和缓存命中均重新读取。没有 Git 文件格式 parser。S4 的 RepositoryMap/CodeSearch generation 仍须消费其精确端口，本阶段不替代 S4 索引验收。

历史记录读取和反馈单独核当前 caller/project/材料访问与固定结果版本；旧 goal/model/policy 的人工历史反馈可以保留。缓存和后续结果消费仍核全部当前版本，旧版本不会获得执行或准入权限。

## 实际准备环境与首件记录

Python `3.13.13`、torch `2.14.1+cpu`、transformers `5.19.0`、ONNX `1.23.2`、onnxruntime `1.31.0`、onnxscript `0.7.2`、laya `0.4.7`；Node `24.16.0`、TypeScript `7.0.2`、onnxruntime-node `1.30.0`。依赖来自实际安装结果，源码和模型固定为上文精确 revision。

English 权重 strict=True 匹配，官方 export 三次真实 CPU torch/ONNX parity PASS，最大 logits 绝对差为 `1.2e-7 / 2.4e-7 / 1.2e-7`，act probabilities 差均为 0。六个实际文件为 encoder/head ONNX、两份 external data sidecar、tokenizer/config。首次导出在 PyTorch Unicode 状态打印遇 Windows GBK 编码失败，原件 `laya-export-first.log`；UTF-8 复验 `laya-export-utf8.log` 使用同一模型与官方导出，未降低 parity 要求。

直接 SDK CPU 真实三题型/全文截断/候选覆盖检验已执行。首两次 options 前置假设 FAIL 原件保留，候选 ID 先前各自短前缀可区分；真实文档段落的完整标签构成第三次输入，实际 `total=90 / distinct=1 / tokens_per_option=4`。SDK actual model `laya-rl-agent`，local load 的 SDK reported revision 为 null；资产 revision 由严格匹配和已加载文件版本分别保存，不伪造供应商 echo。English 的 `choice:11+` 原配置温度 `0.10058280825614929` 被 SDK 警告为 uncalibrated；产品在推理前拒绝受影响 bucket。实际 SDK 诊断原件 `laya-live-first.log`、`laya-live-options.log`、`laya-live-collapsed.log` 与 `laya-live.json` 在忽略目录。

multilingual 同精确 HF revision 真实下载并 strict=True 匹配，首次官方导出 `OSError: [Errno 28] No space left on device`，原件 `laya-export-multilingual-first.log`。只移除了本树可重新下载的 English 源权重，先保存实际 digest；已通过的 English ONNX、原日志与配置保留。multilingual 官方复验三次 parity PASS，最大 logits 绝对差 `1.4e-6 / 9.5e-7 / 2.1e-6`，act probabilities 差均为 0，实际六项 ONNX 资产已发布到本树忽略目录。multilingual 正式配置的 temperature 三项均为 1，per-bucket 配置为空；实际 Node 推理与语言覆盖仍待验证。

## 验收状态

| 项目 | 当前结果 |
|---|---|
| 精确 B 基线与 clean | PASS |
| 官方文档完整读取与原件 | PASS；网页首错误保留 |
| SDK 源码许可与固定模型卡 | PASS；分发仍须保留适用 LICENSE/NOTICE |
| 官方 SDK build | PASS，忽略目录真实 npm build |
| English ONNX 导出与 parity | PASS，官方三次 CPU 数值检查；首编码 FAIL 保留 |
| multilingual ONNX 导出与 parity | PASS，官方三次 CPU 数值检查；首空间 FAIL 保留 |
| SDK 三题型本地 CPU / 长截断 / options不足 | PASS；正常短输入未截断，长输入和候选覆盖损失实际记录；首前置 FAIL 保留 |
| Node/Go 常驻 RPC / 重启 | PASS，English/en、multilingual/en、zh；独占 D 资产实际重载/重启 PASS |
| 断网 | NOT_RUN，隔离范围裁决待答 |
| Jev 真 API/model/usage | NOT_RUN，缺本进程 secret |
| 真实 SQLite/HTTP/撤销/缓存/反馈重启 | NOT_RUN |
| AT10/AT11/AT12 | NOT_RUN |

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

当前支持语言验收仅 English/en 和 multilingual/en、zh；其他语言未宣布支持。实际 OS 完全断网尚未执行，独立问题 `msg_d85e334bfdc4` 等待隔离范围裁决；真实 Jev key、B 生产迁移/锁、H 公共字段/装配和同事务委托消费端口继续按准确交接处理。

## 第二阶段独立 REPORT

SOURCE `30ddaa76b19807e78bf6f56421b4461a6b3cd3f4` 已普通提交/push。本阶段记录 Go/Node 真实成熟 RPC、三题型、两个 checkpoint 与三条明确语言路由的实际重启、真实 Git snapshot 校核和 D 资产重载；没有运行 HTTP/SQLite 完整业务或 Jev 真 API，没有改变公共接口或 S1/S2 批准。

`go test -modfile=.verification/s5/go.mod -mod=mod ./internal/judgment ./internal/adapters/judgment` PASS，私有核心/适配器编译；`go vet -modfile=.verification/s5/go.mod ./internal/judgment ./internal/adapters/judgment ./tests/s5` PASS；真实 model 测试命令和结果见上节。`git diff --check` / `git diff --cached --check` PASS，root `go.mod/go.sum` 零变动。

root 正常入口首检 `go test -mod=readonly ./internal/adapters/judgment` 为真实 FAIL：`github.com/olivere/ndjson@v1.0.1` 要求 `github.com/google/go-cmp@v0.5.6` 的 go.mod sum 缺失。原件 `.verification/s5/root-dependency-first.log`、安全结果 `root-dependency-first-safe.json` 保留；sourcegraph SDK 同时尚待 B root 依赖集成。该结果已准确交 B/协调者，忽略 module 的组件 PASS 不替代 root 默认 build。公共路由、三表生产迁移、同事务委托消费和实际断网按未决/准确来源继续，不降低验收或宣称整个 S5 完成。
