# Astrocyte 当前状态

日期：2026-10-10。当前任务：[S1 Attention 计划](tasks/S1-plan.md) 与 [summarize 接入](tasks/S1-summarize-plan.md)。完整真实验收尚未通过。用户决定见 [docs/QUESTIONS.md](docs/QUESTIONS.md)。下面论文第二步、首轮界面与 S0 记录为历史结果，不代表新 summarize 全部通过。

视频修复正在进行，当前切片与互斥写域见 [修复计划](tasks/S1-video-fix-plan.md)。用户已授权所选B站DNS例外，01:07实际热加载204与上游网络检查PASS。W2实际Reader/Service导入成功730.29秒，正文32908bytes、三附件与SQLite/objects重开一致，人类关注/整理/Mission0；已验收释放。W0修正Windows项目CLI真实路径与既有默认期限1800秒，Codex独立180秒不变。统一源fd9937b42842a93a9b63d417a27725bd68b48389普通合入root为017c133ec4322a6d1011874eee25e148d814fd9c，主控独立pnpm check/build PASS（API14/前端40/契约226等）。W3已01:33:18从空库真实页面提交202，唯一转写仍在进行，浏览器正向/实际服务器进程重启尚未通过。以下未解决DNS与四轨释放等是前轮快照。收藏夹账号、先发现再选择的反馈范围和同步频率尚未决定，本轮先修指定单视频。

当前预览01:35:52重启为017c133业务源，5173/8787均200；个人SQLite材料/作业仍0，验收资料未灌入，Codex关闭。API PID86716、Vite56584、父64976，程序目录`%LOCALAPPDATA%/Temp/astrocyte-dev-QtmjWR`；停止前应重新核对。旧进程退出再次出现既有清理竞态，旧PID/监听实际消失后才启动新服务。无会话直接GET资料被正常403拒绝，未绕过；随后只读SQLite计数核实为空。本轮完整浏览器回归待W3媒体结束后执行。

用户已允许Python，原无Python限制撤销。正式接入沿用原 W0/W2/W3/W4 工作树和互斥写域，本机 Orca Codex / gpt-6.1-sol；主控没有写业务代码。当前普通合并源 `81e92ae521fadf6def68b771e005bd77cc3f7ed3`，含 W0 最终 `d8f29070073e6f2e71ae6fcca5fa259205208b8f`、W2 `777f8559d5ba3afa17c01681397b5dc2bbc0be8f`、W3 最终 `e8cc1e001cdf3468187c901f7fa1e262ebc40359` 与 W4 最终 `1a77d16c83f063c94e3df6d2a5bcaa785523b058`。各轨已 push、终端已结算释放，工作树与历史保留。W3 原 native 内存分配失败后，按同任务 retry-of 恢复原轨；没有重建调度器。

已接入项目锁定 summarize/core 0.25.1、真实视频 URL 表单、上游字幕优先和本地媒体转写薄适配、真实输出/provenance、作业失败提示及同页输入恢复。Python yt-dlp 2026.08.19 使用独立环境，ffmpeg 9.0.2、whisper.cpp 1.9.2、多语言 ggml-base 在项目独立目录；冻结安装、真实工具启动与模型加载通过。全局 summarize 0.21.8 与其他软件工具未覆盖。首次模型整理仍沿既有显式 Codex 作业；本轮未发应用 LLM 请求或重放旧 UNKNOWN。

当前网络将 `www.bilibili.com` 解析为 `198.18.0.160`、`arxiv.org` 为 `198.18.0.40`；锁定上游的网络检查拒绝。真实 B 站 URL 经页面→HTTP202→Service→SQLite 保存为失败作业，action `configure_public_source_network`，页面中文提示可见；同命令重放与实际数据库重启保留原 job/attempt/error，无资料、沉淀、Mission或人类关注。W3 两尺寸实际页面及恢复 4 PASS / 2 SKIP；这是失败持久化通过，**不是视频正文成功、新表单去重或首次整理通过**。主控此前 Orca 内嵌页面曾 Failed to fetch 且无POST，保留该观察，不用独立 Chromium 成功解释其原因。

主控在上述准确源 `pnpm check` PASS（Go/vet/mod、18包边界、226契约例子/生成、API14/14、TS/lint、前端40/40），`pnpm build` PASS；完整 `pnpm test:e2e --workers=1` 为 **146 PASS /4 SKIP、4.2分钟、exit0**，两尺寸均完成。4个跳过是两个视口的真实视频成功导入与当前网络专用负例，后者由W3显式运行并通过，前者仍未通过；本次回归未发模型请求。W0默认并发首轮完整浏览器为130 FAIL /4 PASS /2 SKIP /12 NOT_RUN，第二轮 workers=2 为142 PASS /4 FAIL /2 SKIP；后续定向8/8 PASS分别记录，原RED未抹除。运行时出现过Windows pagefile不足/native内存失败，主控最终回归没有与其他浏览器并发。测试服务退出后15173/18787监听已消失。

未解决：指定视频字幕/中文转写及升级 arXiv 真正成功、最新真实自动多轮/摘要重试、授权 Agent 多次读旧资料正路径。AT04 历史真实论文 later 已通过，完整 AT01/02/03 仍未通过。需要用户决定公开来源真实DNS例外、普通视频导入复用/主动刷新行为、Agent可读范围与新模型作业；登录/扩展不从Python授权推导。不移除上游网络检查、不读取cookies、不扩大Agent权限、不自动回退旧CLI。

预览已在00:32重启为当前源：[Attention](http://127.0.0.1:5173/attention)，API8787；页面和健康HTTP200，原个人库沿用且未写验收资料，Codex关闭。终端仍为 `term_af3ba99d-1301-41be-be0c-860189357aaa`；实际新API PID85056、Vite PID69864，父86148，创建2026-10-10 00:32:27，临时程序在 `%LOCALAPPDATA%/Temp/astrocyte-dev-BAxpTt`。下次停止前须重新核对，不使用此历史PID猜测操作。旧预览仍出现“子进程不存在”清理竞态，主控确认旧进程/5173/8787全退出后才启动。当前自动解析项目CLI与独立媒体，不沿用旧全局CLI覆盖。所有验收服务/自建页面已退出；主控临时库 `astrocyte-s1-Te659M` 保留，旧两个目录的自动审批删除拒绝记录见下文。未执行本轮远端CI、main合并、扩展安装或新应用LLM作业。

## 第二步：真实导入、版本与重启

当前业务合并源 `556165c338519e7f35818d9ea05113cc3644f862`，普通合入 W0 `40f0543ddaa99d0dd1a2b99e568d50a47728a87c`，含 W1 固定版本/别名作业复用修复 `09957ee3d43bde2e555173171bf07345bca1698e`。本轮只恢复原 W1/W0 工作树和写域，主控独立验收，没有改业务代码。第一步契约发布和本次操作细节见 [S1 计划](tasks/S1-plan.md)。

| 本轮操作 / 命令 | 结果 |
|---|---|
| 真实页面导入指定 `2504.16054v1` | 官方正文和附件入库；新表单重复提交首次发现 jobs2/downloads2，已退 W1 修复，保留该失败 |
| 修复后页面导入获准的 `2501.12948v1/v2` | PASS：同一资料两个来源版本；v1 正文59,601字符、PDF1,312,189字节，v2 正文190,684字符、PDF5,034,998字节；旧版本不被覆盖 |
| 新表单重复v1、官方PDF别名、论文ID | PASS：同一个job、attempts=1、资料/版本/对象数量不增加；新表单使用不同幂等键 |
| 实际停止后端并沿用SQLite/objects重启 | PASS：详情与重启前一致，两版正文、8个附件的API结果与存储字节相同；失败视频作业仍在；资料1、版本2、作业3、回执6，人工关注0、Mission0 |
| 重载浏览器并切到旧v1 | PASS：1920×1080 / 1280×720，显示旧正文且无横向溢出；截图在本机web/test-results/root-s1-step2-restarted-v1-1920.png / -1280.png |
| `pnpm check` | PASS：Go/vet/格式、18包边界、契约226示例/生成一致、真实API14/14、TS/lint、前端40/40；既有EventV1警告保留 |
| `pnpm build` | PASS：Go程序及生产web资源 |
| `pnpm test:e2e` | PASS：144/144，59.9秒；代表材料测试与上述真实官方导入分开 |

论文导入纵向链路通过；指定视频仍不能标为通过。页面上传指定B站的真实summarize原输出后，作业以 `evidence_missing` 失败，没有创建视频资料；缺少真实字幕/总结的外部条件未解决。提取沿用summarize 0.21.8，无Python、无模型调用、无新增视频下载/转录器。多版本论文授权仅用于导入验收，不扩大模型可处理来源。

兼容限制：保留旧Refresh去重键及回执，旧库升级后的首次新规则导入可能再建一个作业，后续相同固定版本/别名复用；不删除历史重复记录。A→B→A head仍未决定。完整AT01、最新真实自动多轮及授权Agent读取AT03仍未完成，下方首轮边界继续适用。当前5173/8787预览已重新编译为上述业务源，原个人库保留、未导入验收数据、自动Codex关闭。开发预览停止时仍出现Windows“子进程已退出”清理竞态，确认旧进程和监听消失后才启动新服务；没有改写其失败为成功。

验收进程与自建浏览器页已停止/关闭。自动审批审查拒绝递归删除测试临时目录，只有“blocked by policy”原因；未换工具绕过。`%LOCALAPPDATA%/Temp/astrocyte-s1-Uoy9rJ` 和 `astrocyte-s1-OHU8NF` 因此保留，含本轮公开验收资料和测试程序，可由用户人工清理。

## 首轮 S1 主控集成与验收（历史）

主控分支 `s1/attention-materials-20261009`，普通合并源码 `e51676bca9508d698bd9968ca1f3e4777a7469a0`，含 W0 集成源 `ba7bdf6fe806a573e4531bf47cb8d1592d1c7baa`；随后 `01e3b0aa4947935fd9b31e06227d234d1277afb6` 只合入最终任务/验收报告，没有改变业务或测试源码。以下是本次源码的主控结果，后面的阶段记录保留历史含义。

已实现真实导入、不可变来源版本及附件、人工三层沉淀与固定输入复用、关联/待查问题、候选依据/用途/下一步及人工反馈、分类域、算法排序、项目空间 @固定版本引用、排序配置历史、人类关注与机器使用分离、持久化作业/幂等/失败恢复。本机 Codex 0.162.0 适配已组装，未知外部结果不能自动重发；模型只给建议，由人保存候选。只修改 Attention 业务，Workspace/Swarm 执行仍属后续切片。

| 主控命令 / 操作 | 实际结果 |
|---|---|
| `pnpm check` | PASS：Go 格式/vet/测试、锁/模块核验、18 包边界、226 契约示例与生成漂移、14/14 实际 API 检查、TS/lint、40/40 前端单测；既有 EventV1 未引用警告保留 |
| `pnpm build` | PASS：Go 可执行文件与 web/dist |
| `pnpm test:e2e` | PASS：144/144，59.2 秒，1920×1080 / 1280×720；契约代表材料与失败注入不等于指定公开视频或真实模型成功 |
| 指定论文浏览器实际操作 → API → SQLite | PASS：官方 PDF/HTML/Atom/summarize 原始输出，94,712 字符正文包，v1；两轮明确标为 human/manual 的内容与主题待查整理；候选保存后选择 later，重载仍延期，原文/两记录保留，没有 reject，Mission=0 |
| 指定论文候选两尺寸目视复核 | PASS：详情与延期状态正常、无横向溢出；截图位于本机忽略目录 web/test-results/root-s1-real-paper-1920.png / -1280.png |

这次人工验收使用临时库和真实 UI 表单，没有以 API 写入代替浏览器操作。资料 ID `B25L7ELU3T5EOEZYVP4PQSNAT2`、候选 `SZFBKWKNSD3ZM65HJ25PFNERJX`；temporary/source 为 W0 `a84e1ec5b0f61c2ac67eca8662341d480f7dcf21`，与最终集成的业务源码相同，后续只增加测试/报告。主控只读核对首次把 opportunity.state 写成 status、第二次把 material.current_revision 写成 revision，断言失败；按契约字段修正后上述核对 PASS。另一次截图脚本在抽屉动画完成前取矩形/用过窄文字选择器，未作为视觉通过依据；等待真实反馈内容加载后另拍图并目视复核。

| S1 指定验收 | 当前结论 |
|---|---|
| AT01 | 部分通过：真实论文全文和人工多轮有来源且无 Mission；指定 B 站视频没有真实字幕/总结，最新自动模型调用结果未知，完整验收 BLOCKED |
| AT02 | 去重、版本、同请求/并发、固定输入复用、已知失败恢复及未知禁止重发的 API/持久化/浏览器场景 PASS；指定视频与最新真实摘要重试未通过，A→B→A 当前版本选择仍待用户 |
| AT03 | 刷新中性、人机分离及未授权拒绝 PASS；授权 Agent 多次读取旧资料正路径 NOT_RUN，读取范围仍待用户，不能以默认 403 替代此验收 |
| AT04 | PASS：真实论文候选 later，经浏览器、API、SQLite、重载核对，资料仍在，无拒绝反馈/样本，无 Mission |

视频仅复用 summarize 开源项目的真实既有 JSON/Markdown 导出，保留原始输出与已有时间位置；不建设下载/转录器。用户再次明确这一方向。当前 summarize 0.21.8 对指定 B 站 URL 只有推荐网页文字、llm/transcript 为 null，产品拒绝为 evidence_missing，而非成功视频。OpenCLI 的公开字幕调查仅是诊断，不是新增默认导入后端。

仍待决定/人工操作：Agent 读取边界、私有资料外发、候选主题关联门槛、A→B→A 默认 head；指定视频的可用 summarize 真输出；是否允许最新 Codex 超时后另建新作业。旧结构真实论文整理成功不能代替最新结构：后者180秒超时，完整结果/usage/费用未知，发生于独立 adapter 探针，不在产品作业库中，未自动重放。CLI legacy 探针曾触发宿主 ACL 刷新，没有修改前快照，未凭猜测撤销；准确路径/ACE 和恢复限制见 [W2 交接](tasks/S1-W2.md)。固定文本处理参数已实测拒绝执行入口，但不声明 OS 读取隔离通过。

本机预览已重启为本次 S1：http://127.0.0.1:5173/attention（API 8787），终端仍为 `term_af3ba99d-1301-41be-be0c-860189357aaa`。原数据库重启前已备份到 `C:/Users/DW/AppData/Roaming/astrocyte/backups/pre-s1-20261009-2230/state.sqlite`；启动完成迁移，实际 S1 会话/空资料/作业/候选 API PASS。启用现有 summarize 的无模型原文提取，Codex 自动处理继续关闭；没有导入验收候选或外发私有资料。旧预览停止时旧版脚本报告子进程已不存在，监听与对应旧进程均确认退出后才重启，未猜测杀进程或删除遗留临时目录。尚未合入 main；本轮 CI 未查询，不能沿用历史 PASS。

## S1 阶段历史（以以上最终结果为准）

主控分支 `s1/attention-materials-20261009`，基线 `d6c1bf2`。Orca Codex 五条互斥写域轨持续开发与返修，实际模型 `gpt-6.1-sol`；契约、迁移、锁和入口只有 W0 写入，主控只布置、维护决定及最终验收。

已发布阶段：W1 用例 `f7bf7fdb1c6f311e19a64c54cc1334d247b735ab`、W2 存储/导入 `7d27ccdde41e6ef57d8b82faa6d81e74e8fea5db`、W3 页面 `eeb269ec24fe53d5e45210870a83ffc75b23a26c`、W0 可运行后端 `d89f3567ebba07d6e7bc022ae37cbfb75dbeca01`。后续修复和扩展持续普通合并到 `s1-contract-1009`，尚未最终合入主控；阶段绿灯不能沿用到未验的新源码。

实际结果：W0 在 `d89f356` 的 `pnpm check` / `pnpm build` PASS。W4 已运行独立真实 API、SQLite、对象与浏览器场景，首轮 4/8 API 通过，发现数组为 null、topic/theme 阶段不一致、反馈 HTTP 状态等问题并交原轨修复；各轮结果保留于 Worker 报告，S1 最终验收未通过。真实 arXiv `2504.16054v1` 已通过官方导入并取得 16,213,397 字节 PDF；指定 B 站视频 summarize 仅取得推荐网页文字，没有字幕或真实总结，不能作为视频验收成功。

当前新增工作：人类分类的域、算法排序、项目空间 @文件引用（原分类与原文保留）、本机 Codex CLI 自动整理及原生读取隔离。Agent 默认不能读取全库；是否可读取纳入文件的间接引用仍待用户。AT01–AT04 正在验收，自动 CLI、多轮真实视频及授权 Agent 正向阅读尚未通过；S2–S6 未实施。已有 5173/8787 预览仍属于下方 UI 基线，新源码用隔离端口和临时数据库验收，不覆盖现有个人数据。

更新：W0 `c20fc706e81bd3d6a868d834c60c532cfeba2d31` 的正式 arXiv 全文入口已真实运行，保存 94,496 字符正文、官方 HTML、PDF 与原始 summarize 提取结果，重启后仍可回查；提取阶段未调用模型。B 站公开字幕需要登录，page-only 提取现被拒绝为 `evidence_missing`，没有生成伪视频资料。W1 可调排序 `8267aafbb5b12488b31378327131ae6a8b83cfbd` 与 W2 存储合并后，W4 真实 API/持久化检查 14/14 PASS；新域/@浏览器与自动整理仍待完整集成验证。W3 新页面阶段 `27a3f1832c6b75ad543aa89585bb3a37f3ea5b9f` 已 push。

本机 Codex `0.162.0` 曾完成公开论文全文整理，但最新输出结构的真实调用超时，结果未知，不自动重发；用户是否允许新作业仍待答。CLI legacy 沙箱探测曾自动修改部分宿主读取 ACL，没有改动前快照，未凭猜测撤销；后续采用已验证会拒绝工具执行的固定文本处理参数，不宣称 OS 读取隔离通过。以上都是阶段结果，主控最终 S1 check/build/E2E 尚未执行，AT01–AT04 未完成。

## 历史交付：界面替换

分支 `ui/preview-replacement-20261009`。源码提交 `5c1517d652afef8273cf14dab5c28021db648cfa`，已 push；S0 交接基线 `649d4297d5bd34ebd849944379cb07bf909d7acb`。

已把用户提供的 `C:\Users\DW\Downloads\Astrocyte-preview.html` 的白绿布局、品牌、顶部搜索、侧栏、三页标签、封面卡、会话卡、右栏、网络插画和详情抽屉迁入现有 React/TypeScript。保持现有 API、查询层、DTO、依赖与 Go/SQLite。开发使用 Orca Codex / `gpt-6.1-sol`，单轨独占 `web/`，主 Agent 最终检查、提交与 push。开发记录见 [前端交接](web/UI-preview-report.md)。

主 Agent 在上述源码提交、Windows/PowerShell 实际验证：

| 命令或操作 | 结果 |
|---|---|
| `pnpm check` | PASS；Go 格式/vet/测试、依赖、13 包边界、226 契约示例、生成漂移、TS、lint、37/37 单测；保留既有 `EventV1` 未引用警告 |
| `pnpm build` | PASS；`dist/astrocyte.exe` 与 `web/dist` |
| `pnpm test:e2e` | PASS，112/112，51.0 秒；1920×1080、1280×720，独立临时数据库与端口 |
| 三页 × 真实/示例 × 两尺寸预览 | PASS，12 张截图，标题、横向溢出与页面运行错误检查；目视复核布局、长标题、右栏、抽屉与拓扑示例 |

本轮 [CI 37928464780](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37928464780) 针对源码 `5c1517d` 已启动，记录时仍在运行，不沿用下方 S0 的 CI PASS。后续交付记录只改文档，不重复触发应用检查。

开发阶段 E2E 依次为 86/88、108/110、110/112，失败分别是旧标题断言重复匹配、新测试选错预算样本、空预算详情缺少明确未知标记；修正后 Worker 与主 Agent 各自另行取得 112/112。前轮失败保留在前端交接，不改记成功。

真实限制：默认继续读取真实 API，`?fixture=1` 显式示例；改为默认示例尚未决定。搜索只筛选当前页已加载数据。收藏、研究路线与事件接口尚未接入；固定研究路线/动态/拓扑只在显式示例模式出现。写入、批准、执行、原生接续与采用仍未实现，相关按钮禁用；S1–S6、AT01–AT16 未因界面替换变成 PASS。未合入 main。

当前预览：http://127.0.0.1:5173/workspace；显式示例：http://127.0.0.1:5173/workspace?fixture=1。新预览终端 `term_af3ba99d-1301-41be-be0c-860189357aaa`；API 仍为 8787，沿用仓库外数据库。停止预览在该终端按 Ctrl+C。

## S0 基线与历史验收

- 起始代码：只有 AGPL v3 LICENSE，初始提交 3a02ce38481d1885adee2e051a1b524d78376897。
- 文档基线：de433528d360413305a4a65c997e11ef7fbf4841，已 push baseline/workbench-v0.1。该提交的 SPEC/TASKS 与用户下载文件逐字节一致。
- S0 交付分支：s0/workbench-foundation。SPEC/TASKS 后续修订记录用户确认，原文保留于上述基线提交。
- 远端：https://github.com/songconmaisaix31-design/Astrocyte。
- 主工作树：C:\Users\DW\orca\Astrocyte；原有 charybdis 工作树保持原位。

用户决定见 [docs/QUESTIONS.md](docs/QUESTIONS.md)：先交付 S0；主 Agent 最终 check/提交；并发无数量上限；允许依赖/模型下载；模型按任务难度选择；summarize 和 arXiv 为导入方向，案例后定。

## 环境与进度

| 项目 | 当前结果 |
|---|---|
| Node / pnpm | v24.16.0 / 11.27.0 |
| Codex / Claude Code | 0.160.0 / 2.1.238，原生能力尚未验收 |
| Go | go1.27.2 windows/amd64，已安装；纯 Go SQLite，无 CGO 依赖 |
| W0 契约与入口 | 已合入 a1530cd；锁定依赖、生成客户端、开发/检查/构建脚本及 CI |
| W1 Go/SQLite | 已合入 72409eb；含保留记录的迁移、一致性备份、完整错误响应 |
| W2 三页 | 已合入 96cb7f5；含刷新失败保留数据、焦点与请求清理修正 |
| W3 本地工具清点 | 已合入 59fd965；8 个 Agent CLI、summarize 版本与帮助入口已核实 |
| S0 | PASS；主 Agent 本地检查、构建、浏览器与实际启动通过，最终分支 Windows/Linux CI 通过 |
| S1–S6、AT01–AT16 | NOT_RUN |

开发使用 Codex、Claude Code、Pi、OpenCode 四种本地 Agent；CLI 名称与实际配置的模型分别记录，不以客户端名称推断模型。

业务集成源：`856698000fba3271d25ff68434ef4ced08e2d9fc`；Linux 浏览器测试退出修复：`0169ce2c075c94e5c850408c6f651fb0a1718cf5`。最终源码为主分支普通合并提交 `0a0f49c25f4c2e4a7714e80387f98c48dd7441e3`。

首次合并检查：`pnpm check` 在 gofmt 阶段 FAIL（`cmd/server/main.go`）；独立 `pnpm --dir web lint` FAIL（动态 Hook 依赖，另有三项警告）。原负责人修正后，主 Agent 在上述最终合并源码运行 `pnpm check` PASS。OpenAPI 保留一项 `EventV1` 未引用警告；ESLint 零错误、零警告，Vitest 37/37。

## 主 Agent 最终验证

环境：Windows / PowerShell，SOURCE_SHA `fa7b97f659e6773d83d7dd864ddd0b0835ceac77`。

| 实际操作 | 结果 |
|---|---|
| `go mod download` | PASS |
| `pnpm install --frozen-lockfile` | PASS |
| `pnpm check` | PASS；Go 格式/vet/测试、依赖、13 包导入边界、226 契约示例、生成漂移、TS、lint、Vitest |
| `pnpm build` | PASS；`dist/astrocyte.exe` 与 `web/dist` |
| `pnpm test:e2e` | PASS，88/88；1920×1080、1280×720 |
| 两尺寸三页与详情截图检查 | PASS；标题、徽标、布局、示例标记和禁用操作可辨认 |
| 构建产物在仓库外目录启动，按 OpenAPI 验证实际响应 | PASS；24 操作、空集合、501、Origin 拒绝、错误 DTO、4 网页路由 |

退出修复合入后的主 Agent 复验：SOURCE_SHA `0a0f49c25f4c2e4a7714e80387f98c48dd7441e3`，`pnpm --dir web typecheck`、`git diff --check`、`pnpm test:e2e` 均 PASS（88/88）。应用代码未因该三行测试服务配置修改而改变。

## CI 与退出问题

最终分支 SOURCE_SHA `0a0f49c25f4c2e4a7714e80387f98c48dd7441e3`，[运行 37901441478](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37901441478)，整体 success：

| 环境 | 实际命令与结果 |
|---|---|
| Windows-2025 | `go mod download`、冻结锁安装、`pnpm check`、`pnpm build`、Chromium 安装、`pnpm test:e2e` 全部 PASS |
| Ubuntu-24.04 | 上述命令全部 PASS；另 `go test -mod=readonly -race ./...` PASS |

初次集成源的 [运行 37897830295](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37897830295) 保留为 cancelled：Windows PASS；Linux 的 88 项断言成功后服务退出卡住，取得日志时主动取消。退出卡住由 Playwright 默认强制关闭监督进程、留下持有输出管道的独立 API/Vite 进程组造成。原 W0 在 `0169ce2` 加入标准 `gracefulShutdown`（SIGTERM，15 秒），让现有开发脚本关闭自己的子进程组；未改变浏览器断言。修复前主分支 [运行 37900190606](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37900190606) 的 Windows PASS，Linux 仍卡住后取消；修复源 [运行 37901041699](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/37901041699) 两平台 PASS，Linux 显示服务正常关闭，浏览器阶段约 50 秒完成。最终分支上述运行另行通过。

## 本地交付

S0 当时由 Orca 终端 `term_b2100c22-c321-4344-9578-8b9f42601ce2` 启动 `pnpm dev`，就绪检查 PASS：网页、代理后的 health/foundation 均为 200。该旧终端已失效且服务已停止；当前预览终端见本页顶部。

- 网页：http://127.0.0.1:5173；三页显式示例入口在 [README.md](README.md)。
- API：http://127.0.0.1:8787/api/v1/health。
- 数据：`C:\Users\DW\AppData\Roaming\astrocyte\state.sqlite`，位于仓库外。
- 旧预览已停止；当前预览的停止方法见本页顶部。

最终验收记录为纯文档变更，执行本地链接与 `git diff --check` 后提交；不重复触发应用全套 CI。交付分支为 `s0/workbench-foundation`，已 push；GitHub main 尚未合入。

## 后续适配准备

S0 按 SPEC §19.1 的基座与契约验收收口。开发规划 P0 的适配准备不随基座通过自动变为 PASS：

- P0-05：Agent 八项原生能力实测 NOT_RUN；当前报告是版本与帮助清点。
- P0-06：无 Python 向量后端选型与实测 NOT_RUN，尚未决定后端。
- P0-07：summarize 已安装；真实导出样本与字段映射 NOT_RUN。用户要求先搭环境，案例后定。
- P0-08：AOCI/CodeGraph 所选版本、许可和工作树隔离实测 NOT_RUN。
- P0-09：Laya 可直接加载的 ONNX 模型包与断网推理 NOT_RUN；平台未启用该适配器。

以上后端、资产与控制模式的关键决定仍待相关任务明确；arXiv 与 summarize 导入方向已由用户确认。
