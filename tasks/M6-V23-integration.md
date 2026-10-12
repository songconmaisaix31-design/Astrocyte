# M6 V2/V3 工具链唯一集成执行单

用户2026-10-13终审PASS关账V1/V1R，基线3e1f6dae；本轮不重跑或覆盖原历史，不再询V1。原三个领域owner完成并普通commit/push后，沿用唯一原集成owner。root维护本执行单、STATUS和独立验收。

| 固定轨 | 普通合并交付 | 接口与验收 |
|---|---|---|
| W0 V2 | SOURCE93030768b10549542546178a006507615cdd8ba8；REPORTd7044ee1c77ae7fc9980181795b21d203a35de12 | 三主题完整2006条Markdown；生产MCP/原库回溯与零泄漏已独立PASS |
| W1 live | SOURCE4fbea9d7a32ed7b7f0e0eaf382cd7dc9fe7091a8；REPORT101ee708fa835c25a733a82257855749748be180 | 两native根/三git仓闭日、singleton生产Record/Sleep、只读日快照、phase边界；prelock250新事件/重复零贡献已独立PASS |
| W2 V3 | SOURCE4a179cef4735ee6bfa1dc87e0bc4624774f16fc0；REPORT8e092164d4761fffda1ea68419ab374dd45d10b0；生命周期结算后才启动集成 | 题集/ABC全文预算/事实审计/独立双评分/未来预测/协议检查；无损A候选未激活、模型未定、EXPERIMENT未锁、day0不存在 |

固定集成工作树`C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-final-integration`、分支`songconmaisaix31-design/validation-v1-integration`、原Codex --yolo terminal `term_2b81572f-6ce6-407e-8835-acaaef4d4acf`。先读HANDOFF/STATUS和本单，普通合root最新精确文档SHA，再按表合三个精确REPORT。所有SOURCE、REPORT及首RED普通祖先保留，零force/rebase/squash，最终SOURCE/REPORT各普通commit+push。

集成write_paths仅`tasks/M6-integration.md`及领域owner完成后明确必要的少量导入/路径胶水。不直接改领域实现、EXPERIMENT或模型/评分规则；冲突中的领域内容交原owner，不用ours/theirs整文件覆盖。引擎、schema、params、weights、公共semantic mapper及UI相对V1基线零变更。不得锁SHA、启动模型、运行formal/day0、伪造未来或用户评分。没有必要胶水则源码merge commit为SOURCE。

适用检查：三组Python验证全通过；Go M2、M5证据/隐私与cmd/validation定向通过；kind映射生成一致、diff check通过；核对三个交付SOURCE/REPORT普通祖先、冻结写域零diff与资产齐全。原库只读，不重复已通过的全历史扫描。报告明确fixture/actual/independent和UNLOCKED/NOT_RUN；root合并后负责最终整仓check/build和push，不借旧CI归给新SHA。报告只记录实际命令/测试数/源码SHA/UTC或+08时间/剩余门槛。

当前真实门槛：A原JSON221700超过126976 prompt限；可读无损候选72890，完整往返通过；B1003全文逐条装载38条至126970，首真实长推理600秒Timeout/delivery_unknown保留，不自动重试；仅本任务owned服务已核对后停止，模型文件/首证据保留。用户模型+表示选择未回答，保持未决定。双评缺失却可FINAL的源码漏洞已由原W2复现修复，最终4a179ce独立Python35 PASS；最终集成必须包含修复与测试。V3工具交付与协议锁定分列，V4需要七个真实日及最后D+3全部成熟。
