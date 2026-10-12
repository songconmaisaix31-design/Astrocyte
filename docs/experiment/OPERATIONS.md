# V3 本机操作与数据契约

全部命令从仓库根执行。开发客户端是 Codex `--yolo` / GPT-6.1-Sol high fast；它不接收聊天原文，也不是 ABC 实验模型。候选模型部署、fixture、历史真实 MCP、前瞻模型回答分别记账。

输出只能在 `C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v3-w2/` 的新文件/目录。此处文件不进 Git、stdout 或云；Git 只收工具、协议、空表模板与聚合报告。日志创建独占文件；原始失败、STARTED/FAILED 和未知送达保留，不以重跑覆盖。

## 准备与烟测

`python scripts/experiment/experiment.py baseline --out <v3-w2新文件>` 只读核对固定 V1R。`preview --day YYYY-MM-DD --out <新文件>` 是历史题集构造，不是 day0、不生成实验答案。

`probe --out <新文件>` 仅检查 numeric-loopback `/v1/models`，没有 raw 请求。`select-native --out <新文件>` 遍历仅获准的 `.claude/projects` 与 `.codex/sessions`，包括 subagents 和新旧布局，列出全部 JSONL；随后每条完整原生消息按滚动七天时间筛选。不能拿 W1 当日认知命中文件代替七天全文。缺 timestamp、损坏/不可读或缺 content 进入覆盖缺口，正式 B 失败；不把未知归属猜进当天。

`go build -mod=readonly -o <v3-w2/astrocyte-mcp.exe> ./cmd/mcp` 构建现有生产入口。`mcp-smoke --binary <该exe> --out <新目录>` 从冻结 V1R `mode=ro/query_only` SQLite backup 到独占本机副本，再通过真实 stdin JSON-RPC initialize + 三个 tools/call 访问副本；记录 MCP 成功/错误原回执及源库前后 SHA。副本允许现有 CLI 的启动 migration，源库从不开 writable。真实模型调用数为零，不能冒称 ABC PASS。

`python scripts/experiment/download_candidate.py` 下载固定官方 Qwen3-32B Q4_K_M 与 llama.cpp Windows Vulkan 运行器到独占外部目录，核对公开资产大小与 SHA；不是自动改实验模型。模型/运行器候选信息见 [MODEL_ROUTE.md](MODEL_ROUTE.md)。运行器必须以 numeric-loopback 绑定、禁止联网代理，并经本机进程/模型身份核实。

`model-smoke --endpoint http://127.0.0.1:18080 --model <实际alias> --context-tokens <实际n_ctx> --out <新目录>` 先调用真实 models/props/apply-template/tokenize，再进行无私有数据的 JSON 推理。加 `--selection <七日完整选择文件> --day YYYY-MM-DD` 才构造真实七天最强 B 原文并做一次候选推理；只记录候选运行，不冻结身份或赋予 strong-B acceptance。context 配置值不等于实际长上下文推理通过。

## W1 生产接缝

W1 独占 `cmd/validation/**`、events 原 mapper/接入域、scripts/validation 及自己的测试/报告。W2 不复制 mapper/动力学，不直接 INSERT 生产表。W1 接口：

```text
go run ./cmd/validation --live-state <v4-w1/live-name/state.sqlite> --live-out <新capture目录> --live-start RFC3339 --live-end exclusiveRFC3339 --live-claude-root C:/Users/DW/.claude/projects --live-codex-root C:/Users/DW/.codex/sessions --live-morph-root C:/Users/DW/orca/Morphogenesis --live-astro-root C:/Users/DW/orca/Astrocyte --live-mcpet-root C:/Users/DW/orca/mcpet
```

可选 `--live-init` 只读备份冻结 V1R 到新 live 库。prelock smoke 与正式锁后 day0 使用不同目录。正式需 --live-phase formal --live-protocol-sha <root锁SHA> --live-day0 <真实day0>，coverage记录phase/protocol_sha256/day0；非正式禁止携带锁。成功capture提供snapshot_file=day-state.sqlite只读backup，W2读此不可变快照。三仓增量 git 与两客户端经过公共 mapper、逐条 Recorder、原事件时钟、stable ID 幂等和每日生产 Sleep。失败/未知保留 event/receipt IDs，不新 ID 盲重试。

`coverage-safe.jsonl` 每日包含 day、window_start/window_end、capture_complete、client_counts、git_counts、cognitive_success、unassigned、structure_failed、late、source_errors、sleep_receipt_id、sleep_status。W2 只接受精确毫秒完整本地自然日窗口、capture_complete/selection_complete=true、sleep_status=completed或owner确认无新/迟到的already_completed、sleep_receipt_id非空、late/source_errors为0；未完成覆盖不评分。event-receipts-safe.jsonl 用原 event_id/ts/状态对账；raw-local.jsonl 本机独占，不发云模型。W1 最终字面契约与精确源码到达后由 W2 完成适配，不能将尚未实现的 Handoff 当验收。

## 锁定和每日收集

root 在V3普通集成、工具验收、用户模型选择、真实 local-runtime/强 B 全文预算及 live 接缝通过后，给W2原owner Handoff，由W2独占将实际 model、revision、quantization、model_path、chat_template、endpoint、n_ctx 和所有预留写入 EXPERIMENT 规则段。V3不等待V2。工具 `lock --protocol docs/EXPERIMENT.md --route <已验收本机route.json> --root-reviewed --out <新lock.json>` 仅供 root 最终使用，取当前本地日为 day0，写全文 SHA 到 STATUS。W2 不实际锁定。

route记录需真实model/context_tokens/model_path/chat_template/endpoint与规范一致，status=ACCEPTED、local_runtime_verified/strong_b_accepted/budget_smoke_passed/production_live_accepted为root实际验收结果，并指定独立v4-w1/formal_live_root。这些字段当前未填，不由smoke自动给通过。W2规范含完整用户评分字段、两组匿名标签与judge提示，引用模板仅是镜像。CLI门禁错误也追加cli-failures-local.jsonl；锁后修改输出INVALID，原错误不覆盖。

`run-day --day <刚结束的昨日> --db <锁记录formal_live_root下W1不可变day-state.sqlite> --selection <完整七日选择> --coverage <真实coverage> --lock <权威lock> --binary <生产MCPexe> --out <新day目录>` 只接受锁后七个自然日内刚结束的一日；漏日不补写为前瞻。先持久化 STARTED、只读 snapshot，当日题集/原输入落库，再分别执行 ABC；同题/同模型/同 seed，固定轮转平衡顺序。A 三个真实工具缺一失败；B 原文不完整失败；超预算不截单条；C不附记忆。返回 DAY_COLLECTED 需要查看各臂状态，不等于全臂通过。超时/出错保存 FAILED 和 unknown_no_auto_retry，无自动调度或重试。

answers.jsonl、inputs.jsonl、mcp-receipts.jsonl、receipts.jsonl、questions.jsonl 保留本机原输入/输出/错误。fact-scores.jsonl 是独立事件库mapped且已有signal的成功事件确定性分数。user-sheet.json 原字段和两组实际0–2数值为空，只供用户编辑另存；blind-mapping.json 不给评分人。judge-input.json 仅匿名题目与回答，不读用户分数。

`import-user --source <用户自行另存文件> --day-dir <该日目录>` 原样验证并导入 user-scores.jsonl，不自动补人评。用户选择 ab_preference 的 P/Q/tie，三臂 winner 单独观察。`judge --day-dir <该日目录> --lock <权威lock>` 使用锁定本地模型将匿名输入落独立 judge-scores.jsonl；judge 没有事实、用户原因或预测真值裁决权。

`score-predictions --day-dir <该日目录> --db <live库> --coverage <截至真实今天的coverage> --lock <权威lock>` 未过 D+3 自然日末或未来三天采集/Sleep 缺失均 PENDING。到期重检查追加新行，原 PENDING 保留。种子floor、unassigned及无真实事件节点不进入活跃榜。

`summary --day-dirs <实际七日目录...> --lock <权威lock> --out <新聚合文件>` 不生成或补齐答案。第七日只给真实成熟 provisional；末三日 D+3 全到期后才 FINAL，按固定28题/84答、人评7题及独立judge7份OK/合法数值回执检查完整性。缺失/失败judge保持PROVISIONAL；judge偏好不改变用户成功线。按匿名映射还原用户三臂和AB胜率、独立judge三臂/均分/分歧数，并输出三维派生原始数字与本机JSONL引用。失败不删除，不缩分母，不填终局 Markdown。终局报告格式见 [REPORT_TEMPLATE.md](REPORT_TEMPLATE.md)。
