# A 无损列式输入候选（未激活）

协调方2026-10-13要求准备具体预注册选择。现行 EXPERIMENT 和 runner 仍用完整原JSON；本候选不被正式collector导入，不锁协议，不改变B/ABC模型预算、题目或任何生产工具。源代码为 scripts/experiment/lossless_candidate.py。

表示为可读JSON节点：scalar保留原JSON标量（包括null、bool、整数/浮点）；object为原顺序的 `[key,node]` 对；array保留原节点顺序。同构对象行可表示为table。table有精确row_count、嵌套有序schema、columns；schema整数为叶列索引，其余为有序 `[key,schema]`，完整保留嵌套对象和字段顺序。每列name为完整键路径，例如 `["event","ts"]`。

列值只有四种：constant为每行同值节点；same_as引用此前逐行完全相同的列；values为按原行顺序排列的节点；dictionary和indices分别为精确节点字典和每行字典索引。整数/浮点和bool以精确JSON序列化比较，不能把0、0.0、false合并。字典仅在字符体积更小时使用。异构schema/字段顺序保留原array。模型前置明文包含全部解码规则，可沿schema/列名读取原字段；没有base64、摘要、选优、认知计算或遗漏。

真实三项目通过现有生产MCP stdin工具各三次，ordered JSON往返完全相同：Morphogenesis887 evidence行，原173620 token→62735；Astrocyte1096行，221700→72890；McPet23行，5002→2944。全部包含attention_state、explain_focus所有字段/related/related_projects和完整experiences。计数包含相同问题/系统、表示明文与实际模板，Qwen3-32B-Q4_K_M actual tokenizer、131072 context、4096输出预留；此处只计数，模型回答数0，不等于模型理解/长问答验收。

另对原库全部2855成功事件投影（含849 unassigned）做精确ordered JSON往返，没有把未命中项目的原库事件假称为三项目工具返回。原SQLite前后SHA256一致 `279ffd78ad8cf083b1e279fb8c4ed32ae53a662b52d8a0f220f96d98c9816d4e`、分布不变。首次候选和增加全2855投影校验后的第二次都保留在外部v3-w2/lossless-a-first及lossless-a-second；原MCP/事件/编码仅LOCAL，result-safe.json只有计数、公开topic和状态。

实际命令：`python scripts/experiment/lossless_candidate.py --binary <本机生产MCPexe> --endpoint http://127.0.0.1:18080 --model Qwen3-32B-Q4_K_M --context-tokens 131072 --out <v3-w2下新目录>`。新增可逆性测试覆盖异构字段顺序、重复列、常量、字典、nested/related/experiences及类型区别；实际往返是独立于fixture的证据。

尚需root向用户提交模型和输入表示选择、取得选择后原W2修改正式规范/collector并做实际长A问答验收；目前不可自行激活。当前原库fit不保证七日live增长后仍fit，超预算须保留失败并报告，不能事后丢证据。此候选解决的是输入体积，不能证明用户原因胜率、事实准确或七日实验成功。
