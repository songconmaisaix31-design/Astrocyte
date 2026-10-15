# V4 B定向豁免、完整长应答与实际边界

用户2026-10-13真实批准B-only最近七日完整聊天送固定gpt-6.1-sol；唯一root记录scope精确 `V4 B full seven-day Claude/Codex chat context only`，B收集结束失效。A derived/C仅问题不借授权，产品事件库/MCP/API仍禁B原文或响应。旧ABC扩大scope、无授权和不同endpoint均拒绝；root记录未改。

正式候选是既有Windows native codex-cli0.162.0，high/priority、OpenAI Responses/既有ChatGPT认证、context872000、output reserve4096、总1800秒/request-stream retries0/no compact/no tools、固定空cwd及ignore-user-config/rules。不读取凭据、不走app-server继承配置。开发终端型号与实验身份分开；native模型echo、provider wire及精确preflight tokenizer未提供，actual请求配置/完整终止/usage可独立审查。

输入接缝锁前修正史：首全B stdin写全1632692字节，native累计字符cap1048576/actual1505850前置拒绝、4.854573秒、模型0/无usage，首unknown保留后追加not_sent。[官方0.162 protocol常量](https://github.com/openai/codex/blob/c1382380de69521303b416720a52f42d51af6248/codex-rs/protocol/src/user_input.rs#L10)与[累计检查](https://github.com/openai/codex/blob/c1382380de69521303b416720a52f42d51af6248/codex-rs/app-server/src/request_processors/turn_processor.rs#L514)确认拆多个text仍总量受限。官方源码clone首checkout因无关TUI Windows长路径RED保留，所需文件完整、binary未改。

root准许[model_instructions_file初始上下文](https://github.com/openai/codex/blob/c1382380de69521303b416720a52f42d51af6248/codex-rs/core/src/config/mod.rs#L4011)：单文件112万字节dense与低token spaces均Bad Request/无usage；1382字节短控制与801349字节spaces控制完整通过。不推定精确provider cap。root再准许两个官方上下文通道：完整compact JSON前900000 UTF8字节（退到完整codepoint）置初始文件、其余全部置stdin，然后同题。两个fragment逐字原序拼接恢复原messages，内容/类型/顺序不变，无压缩编码/裁剪/摘要/多轮预读。ABC外层SYSTEM/固定框/schema/splitalgo相同，只材料不同；judge外层JUDGE_SYSTEM及匿名完成回答，不附聊天。

公开同量dense双通道完整通过：1121534总字节、input164531/output41、15.779590秒、首尾public标记正确。随后唯一实际模型B全1003/未装0、原packet1632617、初始900603+stdin733053=总1633656字节；actual input491479/cached3072/output495(reasoning202)、28.859787秒。首process0.426338/首agent_message24.429915秒，非provider TTFT；完整stdin关闭、turn.completed、最终output与流全等、no compact/no tools，input+4096≤872000/output≤4096。

首次parent schema RED因harness额外要求explanation字符串，题目explanation:...未规定该限制；修复为非空str或dict，原结构化解释完整保留。同一封存响应离线重验PASS、新模型调用0，原FAILED/unknown不覆盖；v4-w2/b-native-context-split-first/{native/result-safe.json,recheck-001-safe.json}。该次只是启动门禁，不是day0答案或原因真值。

两条真实pre-turn item.error已按官方源码分类为初始化Warning（exec JSONL将Warning转item.completed(error)，status仍Running）：skip_host_skill_discovery开发期提示、code-mode host禁用后fail closed提示；与四个公开成功控制逐字全等，不是model/context/metadata/认证错误。精确白名单仅这两条且只允许turn.started之前；未知item.error、请求开始后error、top-level error/turn.failed/compact或任何工具item失败。原提示与root首严格白名单RED保留，未suppression或重发。安全分类证据v4-w2/pre-turn-warning-classification-first-safe.json。

ABC共同字节上界（root已独立接受）由实际通道导出：900000+共同SYSTEM/框603+4*1048576=5094907；不把今日1.63M人口观测当未来低限制。三臂四题全包字节构造在split-all-arms-budget-first-safe.json：B最大1633702、A最大37448、C最大1698，仅构造验证不是A/C推理。root msg_108213fb5c73已独立接受实际预算/身份，W2已填写终稿字段；新SOURCE CI、最终普通集成/工具验收/root锁仍待完成，day0/SHA空。

下方保留02:15及V3历史，旧权限待答与stdin75前缀被本轮真实裁决/双通道替代，旧失败不删。SOURCE5e的CI38289430526已root独验双平台SUCCESS；本轮新源码CI另列。

# V4 实际路由与预算核查（未接受）

用户2026-10-13决定ABC统一gpt-6.1-sol、高reasoning、priority，使用分层A。原Qwen/本地only候选是历史，不再重问该选择。开发Codex --yolo继承模型保持gpt-6.1-sol、没有子Agent；这一型号与实验响应身份分开。

实际处理器为既有Windows native codex-cli 0.162.0，沿用released distiller隔离参数且强制同一空working directory、通用answer_json schema和75字节输入前缀。metadata account/read仅记录type=chatgpt，未读或输出任何credentials；原生默认边界推断为 `https://chatgpt.com/backend-api/codex/responses`，是外部推理，**不是本地模型或response endpoint echo**。运行器默认auth自行处理、重试0、无私有代理/密钥环境变量扩展。没有用户定向raw授权就不会启动私人请求。具体授权接口见OPERATIONS，root已经向用户提问，未答不能猜批准。

[OpenAI官方gpt-6.1-sol资料](https://developers.openai.com/api/docs/models/gpt-6.1-sol)声明1050000 context和128000最大输出；本机两个已知Codex模型catalog刷新后仍为default272000/max872000/effective95%。二者分列，不把官方极限当当前账户已接受。公开请求使用872000配置，不削弱B到272000，但短请求也不能证明真正872000长输入成立。tiktoken0.13.0不认识该精确模型；没有猜o200k等编码，没有外发私人tokenizer探针，没有不存在的count命令。当前原生只能全人口输入、完整完成后核查实际usage，边界选条尚不可用。

公开原生真实调用：首配置错误（内建provider覆盖非法）保留public-native-first，模型请求未发；改正后public-native-second成功14.196671秒/input8416/output40；固定cwd成功12.114199秒/input8418/output42；最终durable CLI成功11.503839秒/input8428/output42。三次完整output与agent_message匹配、turn.completed actual usage；最后stdin180字节完整写入并关闭。first_process_byte=2.794060秒、first_agent_message_event=7.947629秒，明确不是首token或provider首字节。实际response model echo为空，极限context/精确tokenizer/硬输出cap和统一byte ceiling仍PENDING；配置型号不是身份接受。

全七日[Oct6,Oct13)本机1003消息（Claude321user/619assistant、Codex32user/31assistant）原正文人口未缩小。四题JSON UTF8分别1632638/1632511/1632617/1632601，native共同前缀后各加75字节。root独立按12个真正含窗口消息的原JSONL物理行核对全部content/role/phase/ts及四prompt顺序全等、0删减，root-v4-start/b-full-independent-001-safe.json。这是人口/全文构造验收，不是模型推理或token/byte预算接受。

真实分层A九工具通过：每项目全6状态；Morph887 refs/15组/27931字节，Astro1096 refs/5组/33353字节，McPet23 refs/2组/1927字节，全部approved经验0。SQLite整数聚合误差0、全部引用回溯、2855原事件raw audit0；root独立full_state/header/2006refs审计一致（root-v4-start/layered-a-independent-002-safe.json）。A四题最大完整JSON分别30650/36384/2762字节，均小于全B最小观测1632511；仍要按最终同一预算验收，不截断证据。

full B首隐私RED在v4-w2/b-native-privacy-first/result-safe.json：完整1003/未装0，reason prompt JSON1632617/native stdin1632692字节，PRIVATE_EGRESS_BLOCKED/not_sent/model_calls0；没有私人模型回答、没有实际usage。拟总context872000与输出预留4096、system/question/template预留16384/2048/4096仅是候选策略，实际长输入与全部系统/schema开销待验。原600秒Qwen timeout/unknown保留，不是本次1800秒外部B失败。

# V3 历史本机路由与容量核查

开发客户端：Codex `--yolo`，Orca Dispatch 实际 provider model=gpt-6.1-sol，终端显示 GPT-6.1-Sol high fast。继承模型不变，无子Agent。它不处理raw，不是ABC实验路由。

实机：RAM约31.3GiB；RTX5060 Laptop GPU8151MiB显存、首次空闲7801MiB；NVIDIA驱动582.05。C盘首次剩余49,068,453,888字节。PATH未发现ollama/lms/llama-server；127.0.0.1的11434/1234/8080/8000首次均不可用（各2秒，URLError），原记录v3-w2/route-probe-first.json。不遍历凭据或其他私人目录。

可审查候选为[Qwen官方Qwen3-32B-GGUF](https://huggingface.co/Qwen/Qwen3-32B-GGUF)，固定revision `938a7432affaec9157f883a87164e2646ae17555`、Q4_K_M、19,762,149,024字节、官方LFS SHA256 `efd971561896866f0e910cce52761ca77b1b138090c7f15fe284676d57d1f689`。官方声明原生32,768，YaRN可扩131,072；此声明不能代替本机长prompt验收。依赖固定[llama.cpp b11574](https://github.com/ggml-org/llama.cpp/releases/tag/b11574) Windows Vulkan x64公开资产，32,112,117字节，官方asset digest匹配 `73387123e22bd2a01a66194f02bd899189542188e8ca87d542b463d75bedc164`。Vulkan实际设备Vulkan0为AMD610M、Vulkan1为NVIDIA5060，应显式选Vulkan1，避免误用共享显存。下载/运行文件只在外部独占candidate-qwen32b。

128K目标需要实际weights、KV/计算缓存与CPU offload容量/速度核实。Q4模型文件约18.4GiB，不能全放8GiB GPU；系统RAM/可用RAM与长KV限制另测，不把两种内存简单加总声称容量足够。不下载8B/20B代替最强B，不偷降模型；若128K失败，原失败保留，可审查较短context必须显式记录未达128K，不能冒充强B通过。

更高容量候选是[OpenAI官方gpt-oss-120b](https://huggingface.co/openai/gpt-oss-120b)，官方给出单80GB级GPU的MXFP4运行配置；本机现有8GiB不能按该配置运行。这是路由/容量差距，不是宣称已验证该模型或要求raw外发。本轮不送raw到云/GPU出租服务。

本地HTTP传输仅numeric-loopback，禁proxy和redirect；需要实际models/props的alias、model_path、template、n_ctx。apply-template/tokenize与实际completion usage须完全吻合；special-token和thinking/JSON配置需真实核实。响应模型身份变化、输入截断、输出length、无效JSON和HTTP错误均落原本机回执。正式强B路由还要求用户模型选择、root本机进程/无外发与完整预算验收，候选smoke不得自行赋予ACCEPTED。

当前实际native七日清点（截至2026-10-12，10-06至10-13半开）：授权Claude根7个JSONL，Codex根1714个；Claude user321/assistant619，Codex user32/assistant31，原content JSON字符711583/502845。全根kind-blind inventory与loader交叉一致、缺失0，窗口无event_msg user_message/agent_message-only形状；没有按正文相同去重真实消息，也未扩读其他root。总1003完整消息明显需要实际tokenizer计数与预算装载，字符数不是token数。原文只LOCAL。

2026-10-13候选实测：官方32B GGUF下载1817.559秒、runtime下载4.699秒，大小与官方SHA均核对通过。显式Vulkan1、14层GPU offload、131072 context、单slot、Q4 KV、YaRN scale4/original32768，禁context shift；实际models/props/template及公开JSON短问答通过，tokenizer计数与completion usage一致。该结果是候选烟测，不是最强B验收。

实际tokenizer预算首轮RED：系统/题目/模板基础171 token，输出预留4096，prompt上限126976。A三工具完整历史快照证据221700 token，超上限94724；B七日1003完整消息471591 token，不能全部同时装入131072。原始输入和计数仅在v3-w2/token-budget-first-local.json，本机可公开的数值在token-budget-first-safe.json。完整A当前不可执行，不能删证据、改问句或伪装成功；已向root提交确定性无损表示与更大受支持本地容量两种审查方向。B继续按既定完整单条、最大装载规则测真实推理，保留未加载条数和预算原因。

实际B装载数量、回答/失败状态在M6-W2报告登记。当前没有权威模型选择/协议锁/day0，不能把下载、context配置或短公开prompt推理当全文最强B成立。


最终未锁route草案：v4-w2/route-draft-native-split-first-safe.json，status=DRAFT；root已接受实际gpt-6.1-sol/872000/4096/5094907/完整B/正式生产bridge边界。provider echo/wire/精确tokenizer仍false，不宣称完整极限context已验；SOURCE CI/最终集成仍待root核实。EXPERIMENT实际规范已据Handoff独占填写，W2没有实际lock、day0、七日答案或终局报告。
