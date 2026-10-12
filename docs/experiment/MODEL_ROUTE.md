# V3 实际本机路由与容量核查

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
