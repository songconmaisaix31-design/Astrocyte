# V4 B窄豁免操作与数据契约（未锁；完整B门禁已完成）

2026-10-13定案优先：ABC固定gpt-6.1-sol，正式A为全量状态排序、完整explain_focus数值头、kind×自然日整数聚合、全部event_id引用和approved经验。开发终端不是实验响应模型证明。用户B窄例外已批准；无B专用授权或旧ABC宽scope仍启动前拒绝not_sent。A derived/C无记忆不借豁免，judge显式匿名评分用途并从完成的paired answers重建校验。全部新证据写入外部 `C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/` 独占新目录，旧v3-w2不覆盖。stdout只派生数字与状态，源码/报告不得含raw。

从仓库根执行下列实际入口（参数路径必须是外部既有根内真实文件、新输出不得已存在）：

```text
python scripts/experiment/experiment.py v4-a-smoke --binary C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/astrocyte-mcp.exe --out C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/layered-a-cli-first
python scripts/experiment/experiment.py v4-public-probe --out C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/public-native-cli-first
python scripts/experiment/experiment.py v4-prepare-b --day 2026-10-12 --selection C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v3-w2/native-selection-first.json --out C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/b-preparation-new
python scripts/experiment/experiment.py v4-b-smoke --candidate C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/b-full-population-local.json --out C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/b-native-privacy-first
python -m unittest discover -s tests/validation/experiment -v
```

前两条已实测，最后B命令首轮EXIT1、PRIVATE_EGRESS_BLOCKED/not_sent/0模型调用，原回执保留；上述已存在输出不能重跑覆盖。prepare-b入口复用原selection/loader与题库，读取只读真实库当日事件，保存全部原content及四题完整prompt；1003人口先前构造和root独立逐物理行审计已通过，并非该命令运行的伪证据。每次新烟测选不同新目录，unknown请求不因新目录而盲重试。

root已按真实裁决释放外部root-v4-start/b-raw-user-authorization-20261013.json，B命令加 `--authorization`。文件必须含authority=user、status=APPROVED、model=gpt-6.1-sol、scope=`V4 B full seven-day Claude/Codex chat context only`、endpoint_boundary=`native OpenAI Responses provider; external provider inference, CLI-owned existing authentication`和真实decision_reference。W2不改root记录、不代用户生成APPROVED；B收集结束即失效，B输入/响应只本机harness封存，禁止进入Astrocyte库/MCP/API。隐私block仅not_sent，真正已启动错误/超时保留unknown_no_auto_retry。启动B一次完整回答和usage即使通过，也只是门禁烟测，不计正式七日。

原生CLI隔离配置：固定空cwd、ignore-user-config/rules、schema、禁tools/apps/skills/MCP/子Agent；请求gpt-6.1-sol/high/priority、provider默认采样、872000 context、禁止compact、request/stream retries0、总1800秒。三臂同SYSTEM/固定双通道数据框/完整JSON包/末片段后同题/通用answer_json schema；judge同配置但独立JUDGE_SYSTEM。model_context_split前900000 UTF8字节退到完整codepoint置model_instructions_file，余下全文置stdin；逐字重组，不能截断/编码压缩/摘要或多轮读入。所有可控初始文件+stdin字节共计；原native stdin共同75字节前缀仅历史。

原生只支持**全人口输入+完整终止实际usage**，不支持近似tokenizer边缘选条。全人口fit即最大B，不padding；不fit整臂失败。后续请求归档实际stdin-local.txt、原messages、initial-context-local.txt、argv、schema、stdout/stderr、流事件时间和完整output。JSONL first_process_byte与first_agent_message_event非provider TTFT，model echo/provider wire/精确tokenizer/输出硬cap未露出；实际usage后验input+4096≤872000、output≤4096，总elapsed≤1800，失败保留完整证据、unknown不自动重试。

本轮唯一实际B已完成，命令（输出已存在，禁止再执行覆盖）：

```text
python scripts/experiment/experiment.py v4-b-smoke --input-transport model_context_split --candidate C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/b-full-population-local.json --authorization C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/root-v4-start/b-raw-user-authorization-20261013.json --out C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w2/b-native-context-split-first
```

1003全部/0未装，总1633656字节、input491479/output495(reasoning202)、28.859787秒、完整stdin/终止/usage/no compact/tools。parent首schema RED保留（题目允许结构化explanation，startup错误要求str），同一原答离线重验recheck-001-safe.json PASS，新增请求0；不是重新调用。该次stdin额外归档仅配置重构stdin-reconstructed-local.txt，明确非原始wire捕获，原写入关闭回执仍在。

root msg_108213fb5c73已独立接受共同byte ceiling，W2终稿登记：固定初始上界900603，加turn/start每scalar最多4UTF8×1048576，导出5094907，不以今日人口观测人为压低未来预算。正式route草案只供root审查，不代写ACCEPTED，不实际lock；root另验新SOURCE CI、普通最终集成与工具后才权威锁SHA/day0。native lock要求外部runtime实证+B窄授权、actual identity/full_population_usage/native complete input/byte ceiling/strongB/budget/live与root reviewed，formal_live_root沿v4-w1原生产目录。

下面保留V3操作历史；涉及Qwen候选和本地only接受条件不是V4模型选择或外发授权。run-day、人评、judge、预测成熟和summary规则继续适用；推理seed在native未暴露，只有题库seed固定。V4 run-day实际调用分层A和native full population，错误/输入/答案仍本机JSONL；任何锁后协议字节变化INVALID。

# V3 历史本机操作与数据契约

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

## 本次真实scope及初始context修复证据

旧授权已严格收窄为B-only，root记录未改。真实唯一authorized stdin调用b-native-authorized-first全1003消息/1632692字节完整写入关闭，native turn/start在model IO前拒绝：code -32602/input_too_large、max_chars1048576、actual_chars1505850，4.854573秒，无usage。首unknown保留，独立诊断后追加not_sent而非盲重试；result-safe.json只派生字段。官方0.162累计检查连多text项也不能绕过，官方源码commit c1382380de69521303b416720a52f42d51af6248。

root明确允许官方exec model_instructions_file初始context接缝；保留ignore-user-config/rules与既有认证、短stdin同题、全原messages JSON可逐字/类型/顺序恢复，预算计文件与stdin总UTF8，禁tools/no compact。app-server缺ignore-user-config且空mcp_servers不会清继承，因此不使用它发private。context候选命令 `v4-public-context-probe --out <外部新目录>` 与短控制 `--public-filler-repetitions 4`，私人B仅在长公开探针与root验收通过后可使用 `v4-b-smoke --input-transport model_instructions_file --authorization <root记录> ...`；不能因该参数存在就当路由已通过。

公开long public-native-context-first：initial1121119+stdin235=1121354字节，12.055640秒，turn.started后固定Bad Request/turn.failed，无完整终止/usage，原unknown保留；small-control-first：initial1147+stdin235=1382字节，12.048045秒/input4493/output65，完整终止与首尾public标记正确，no tools/no compact。长输入原因边界未暴露，不把small通过写成full B、872000 context或正式byte ceiling通过。新raw例外与模板候选均在未锁时预注册，day0不填。


两条真实pre-turn item.error已按官方源码分类为初始化Warning（exec JSONL将Warning转item.completed(error)，status仍Running）：skip_host_skill_discovery开发期提示、code-mode host禁用后fail closed提示；与四个公开成功控制逐字全等，不是model/context/metadata/认证错误。精确白名单仅这两条且只允许turn.started之前；未知item.error、请求开始后error、top-level error/turn.failed/compact或任何工具item失败。原提示与root首严格白名单RED保留，未suppression或重发。安全分类证据v4-w2/pre-turn-warning-classification-first-safe.json。


最终未锁route草案：v4-w2/route-draft-native-split-first-safe.json，status=DRAFT；root已接受实际gpt-6.1-sol/872000/4096/5094907/完整B/正式生产bridge边界。provider echo/wire/精确tokenizer仍false，不宣称完整极限context已验；SOURCE CI/最终集成仍待root核实。EXPERIMENT实际规范已据Handoff独占填写，W2没有实际lock、day0、七日答案或终局报告。
