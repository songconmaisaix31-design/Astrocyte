# 重启应用并继续产品与科研（2026-10-19）

用户要求继续开发 Astrocyte、重启应用，并保持科研与产品并行推进。本页接续 `s3-paper-ten-cycles-20261019.md`，原十轮与发布门禁保持；已接受01/02/03/04/06/08共6轮，不重复计数。

## 当前基线与建议

主控分支 `validation/v1-20261012`，接管源码 `28a4b5c4`，工作区干净。产品报告 `57a91f4601b4292df4f55ba6cac541bad5caab3f`，实际Source `15d74da3edcee26550a71120370fda9c86230e04`；科研技术报告 `50e51fb68d51bc1ff950fe346a7540a9e167e145`，科学正文仍为已接受a664233a；审查报告 `54627b067608f0c25e9212ae39b6a5882d05e102`。三个原工作树tracked干净。

接管时64505/17887没有监听，Orca已将原三个Dispatch记为failed/abandoned（terminal_missing），当前仅主控会话；不将中断视作功能失败。原Run `run_29c7e66b4136`已用正式run-use绑定当前主控 `term_2495c17e-98de-4b86-8393-a3992c369aba`。沿原Task及工作树重启中断执行，不建立新开发轨道。

实际psutil阻塞CPU两次样本UTC11:10:06/11:10:09，CPU11.3%/6.6%，可用14.022/14.035GiB，采样结束间隔约3秒。此采样仅作接管基线，各实际重操作仍须重新执行启动守卫。

## 轨道与执行顺序

| 负责人 | 问题、建议与预期效果 | write_paths与验收 |
| --- | --- | --- |
| 原产品Task task_51276f272ff3 | 先静态核原64505服务的真实程序、数据根、启动参数和恢复入口，给出恢复方案；待验证窗口结束后重启同一应用并保持可访问。继续定位已有完整默认CI的具体产品问题，权限待定项不代选 | 沿原产品写域和原PR #2；追加允许README.md更新当前已实测入口。原共享数据只按既有启动配置读取/使用，不迁移、不重建、不用新空库替换。服务启动另获主控窗口后执行，核实际PID/程序/端口、首页与health、真实SQLite和重启前后关键数据。先提交具体可审结果，原Worker负责修复 |
| 原科研Task task_f824c0088743 | 静态核已验TeX Live安装入口与完整来源流程，沿原方案完成Install→Build→Validate，提高可编译来源交付的实际证据；根据准确产品事实准备研究反馈 | 沿原论文写域及PR #3，实际应用astrocyte-research skill。初始没有重操作窗口，独立验包退出后再获明确授权。保留08科学正文/Bib不变，先完成技术阶段；实际PDF/引用/全文与来源检查通过后交独立Reviewer。不能自报投稿就绪，未知作者/许可/正式结果保持未知 |
| 原Reviewer Task task_2923223271be | 先沿已有准确Source15二进制、C工作区新q15-03输出方案完成独立只读inspect，再按原条件完整Verify与DOM；关闭自有服务浏览器后交还窗口。随后审科研技术阶段及产品恢复结果 | 沿原报告和忽略输出写域，沿原PR发布准确Commit CR及review-decision hook。保持原公开Git/SQLite、真实日期、Reader/SQL全部断言；不覆盖旧输出、不写Worker源码。准确候选发生变化先核差异。每个决定明确范围，主控实际读报告和回复hook |

初始唯一重操作窗口授Reviewer独立验包；产品和科研只静态阅读/编辑。首个完整Verify退出后先完成原服务缺失依赖恢复，再由Reviewer按文档Windows PowerShell5.1完成同源Verify/DOM，随后授科研安装与文稿检查；最后产品启动共享服务。需要产品返修时由原Worker完成，再由Reviewer复验。服务运行期间不另起第二服务或其他重操作，必要验证先协调服务窗口，不终止其他应用。

## 约束与停止条件

所有既有重操作CPU<85%、可用≥4GiB、间隔2–5秒的真实双采样、BelowNormal、Go4/-p4/512MiB、Node1024MiB继续执行；固定inspect可用既有3GiB/Go1/256MiB专用额度。单窗口、单服务、单浏览器，禁止视觉和Aliyun；本次不新增子Agent。资源拒绝保留原件，不自动重试同因失败。

原S3预算单位、权限证据与Codex文件/命令关闭选择未答，Claim拒绝和Known=false保持；产品恢复不运行受阻原生输入。完整S3、独立验包、默认CI分别记录。预览首发范围未答，未满足原门禁不发布Release；作者/机构/许可未答，不上传arXiv。正式V4、用户原件和其他应用不操作。

阶段完成须有实际源码/报告Commit普通push、相应检查结果及独立CR/hook；主控维护HANDOFF/STATUS并核真实HTTP与当前监听。可完成工作持续推进，遇具体权限或外部阻塞保留未验项和原始错误，不降低验收。

## 实际恢复记录

原三个failed Dispatch已逐项正式worker-release，回执均released/processAction=none/archive unavailable，未宣称取得不存在的终端转录。原Task恢复时状态为ready，不满足retry-of的failed Task前置；核原准确失败Dispatch后将Task据实际运行时中断记failed，再沿原Task/树显式retry-of启动。

| 轨道 | 当前Dispatch | 当前终端 | 实际启动 |
| --- | --- | --- | --- |
| 产品 | ctx_013843bce872 | term_78b3371d-5de3-4a8b-8213-85f280c81054 | ready/input_accepted/turn_started |
| 科研 | ctx_1fe65b3eaadd | term_87a9c8c3-4457-4f0c-8b5b-5416bff7f447 | ready/input_accepted/turn_started |
| Reviewer | ctx_a8916e612c35 | term_45b72208-38a3-4f83-8b76-d43f5684de7d | ready/input_accepted/turn_started |

三者均使用Codex默认启动配置，未覆盖模型。实际模型/客户端待各自Handoff核实。正式新指令分别msg_d5ec8bbfc581、msg_c0d2f7541542、msg_b5a2896a5ff7；排队回执不作为已读证明。

共享服务恢复只读定位到原H树 `dist/product-recovery-20261016/restore-38fb.ps1`，运行根 `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413`。追加允许产品Worker在该既有运行根生成本次日志并由原恢复入口更新真实running.json；原程序、H脚本、配置与历史日志不改，原SQLite仅按既有服务使用，不重建或替换。先核实际文件和配置再启动，原记录40804仅为历史PID。

实际Handoff核原c5程序、SQLite和Q均存在，summarize CLI及CodeGraph SDK原依赖目录缺失。主控完整读取产品自身plan.md后接受原路径恢复：原standards-backend四份配置/锁不变，pnpm11.27.0冻结安装并核summarize0.25.1；原runtime/sdk package与lock逐字节核等于准确Source15 ZIP后，成熟ZipArchive只恢复完整node_modules，原Check-Runtime实际import/资产核查，不执行推理。追加产品写域仅 `C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/node_modules`及必要忽略缓存、`D:/Astrocyte-agent-assets/20261015/release-package/runtime-sanitized-third/runtime/sdk/node_modules`、原H `process-20261016`缓存。原配置/锁/脚本不改，目标包含检查、原目录不存在、不覆盖既有文件、原完整性与失败立即停止保持。依赖恢复尚未运行。

Reviewer实际HEAD为ffbc5762f578ab39fdacd9a9cc31b970d5beead6，相对54627b06仅追加主控收到原hook的两行记录，已只读核实。新C q15-03的真实inspect于UTC11:17:48 EXIT0，origin-safe逐项匹配1316个真实Git来源与event_id、可达Commit SQL通过；三份自身实际安全JSON已移交D新q15-03。完整验包与DOM尚待Reviewer最终结果，不提前接受。

首次独立完整Verify使用PowerShell7.6.6，在解压/隐私/runtime/初始HTTP/真实快照发布与health通过后因JSON日期成为DateTime、原snapshot_at.Length访问失败而EXIT1，原新v15-01保留，服务正常退出且17889无监听。主控核发布文档第3/78行明确基本条件与命令为Windows PowerShell5.1，下一次独立比较采用相同5.1环境及原Source15/ZIP/参数，新v15-02，成功后新DOM。保留7.6失败，不新增Core兼容或修改原断言/ZIP；不是产品功能失败的判定。阶段审查报告511cdce8f3bbea03f72d293668c334e9d92dff41已普通push，最终CR/hook待实际结果。

产品原冻结依赖安装实际pnpm11.27.0完成187项/378锁政策，原精确CLI0.25.1、四配置完整字节相同。SDK仅从原Source15 ZIP恢复1368文件/377193096bytes，package/lock一致，原完整Check-Runtime EXIT0且未推理；保留原镜像超时与成熟工具原生重试日志，未执行外层重试或改源。四阶段八样本最高CPU19.8%、最低可用12.798GiB、间隔约3.002秒。全部调用退出后msg_5dce3320d058交回，root读原日志确认。准确报告2a05bd0d78cbeb5e69ba09728923ea2f2d1d1b65普通push/远端一致/clean，相对57仅本轨报告21行，Source15/ZIP不变。

独立Reviewer在Windows PowerShell5.1实际完成新v15-02原完整Verify EXIT0，准确Source15/candidate-package-cognition-runtime，130 PASS/0 nonpass，root实际标准JSON逐项结果分组核计数。原SQL/六字段、历史、HTTP与真实进程重启/持久状态通过；随后新dom15-01原完整工具EXIT0，两视口/26品牌/人工DAG增量/撤权清图，3个DAG事件逐项SQL核查。root读原result/DOM/log并核17889无监听，msg_6e503bdda22e实际交回窗口后回复msg_771e61e58629；最终准确报告候选2a05的综合CR/hook仍待Reviewer，不增加轮数或宣布完整S3/发布。

科研已获msg_0b8bc1f9dc5a唯一窗口，直接既有Install→成功才Build→成功才Validate，科学正文/Bib/原制品冻结，实际结果待交接。研究技术结果由独立Reviewer另用自身输出重建核查，之后才产品同库应用启动。
