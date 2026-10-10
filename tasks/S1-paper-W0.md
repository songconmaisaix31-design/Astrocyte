# S1 论文与项目看板 W0

当前 Dispatch `ctx_b162c766fc84` / Task `task_95839accebf3`；开发客户端 Orca Codex，执行环境未暴露可核实的精确开发模型标识，未自行切换模型。实际联合蒸馏客户端/模型另记为Codex0.162.0/gpt-6.1-sol。独占共享契约、HTTP、入口、迁移012+、生成API、脚本及依赖锁；其他领域由原owner交付，最终普通合并精确已push提交。

最终 W0 SOURCE `3b5a4343ff6fe08a56b89c24642050e932d97f9f` 已push，随后独立提交本报告为REPORT并通过Orca交接精确SHA。共享契约/HTTP/迁移/测试及普通集成范围已完成；Search/插件/进度产品决定仍pending，不宣称完整S1。下文历史各场结果保持各自精确源码，最终检查见末段。

## 基线与边界

- 干净工作树 `s1-sync-contract-1010` 普通合入已发布根基线 `087127d30fc505f8fe25c52142a174f5842e2bef`，合并提交 `54336fcabecfd59dac9c9ce34cd830bb1258d9b0`。
- 已读 HANDOFF、STATUS、QUESTIONS 与 `S1-paper-project-board-plan.md`。512KiB UTF-8 是选定资料、引用和指令的总模型输入；输出、原生历史和权限不扩容。
- 搜索/当前页提取、插件授权范围、人工进度/Agent推断仍 ASKED PENDING。未冻结这些接口或选择产品行为。
- 复用现有 same-origin HttpOnly/SameSiteStrict 人类会话与 CSRF；已发现/已安装不推导权限。没有启动summarize daemon、占用8787、改个人5173/8787或读取个人Chrome。

## 本轮实施与验证

已发布 `f9cdec80265f55efbfbdf6a9380ed684e46eee2d`：看板/来源/贡献者、人工notes/review/group/intent/archived/revision/updated_at契约、生成API与012迁移。`pnpm generate`、`pnpm check:contracts` PASS，237示例；旧EventV1未使用警告保持。

普通合入W2领域前置 `35edf6b2aff980ed69e8d8998908cf99f9aa9dc5`，再发布 `7374f40a7f54d2a8427c84209c6909b86e0d0253`：独立metadata repository/service端口与严格人类HTTP命令。已有发现接口可独立工作；未接元数据领域时明确501。请求的revision是人工字段CAS版本（初始0），共用expected_version仍>=1；不接收updated_at、actor、progress或权限字段。

`go test -mod=readonly ./internal/adapters/httpapi ./internal/workspace/domain` PASS。运输定向验证无会话、过期会话、缺CSRF、Agent bearer、外站Origin以及伪造时间/身份/进度/权限不会进入领域；这仅证明运输边界，不替代实际存储。

已普通合入W2中间SOURCE `88b2bcc4b77050b50a56afd068767ceac82c5fa4`，当前业务组装 `25949b3639e8229fb9ff0b94308f75b4837417ef`。此时W2还在修正原生header创建时间与最近活动的区别，未将中间版看成最终时间语义通过。

`ASTROCYTE_TEST_PROJECT_BOARD=1 node --test tests/s1/project-board.test.mjs` 首轮 **1 PASS / 0 SKIP / 31.88秒 / exit0**。真实Orca登记54个观察目录聚合14组，partial与16个原来源限制保留；Codex header来源partial，其余来源unknown明确保留。真实API人工备注/复盘/分组/意图/归档保存、陈旧CAS409、非法身份/权限/时间拒绝、刷新保留、同SQLite的所属API进程冷重启及旧会话403均通过；local project列表及grant数量未变。没有模拟看板种子、模型/媒体调用或个人Chrome。原库保留 `%LOCALAPPDATA%/Temp/astrocyte-s1-bLFn1Q`，命令原日志 `%LOCALAPPDATA%/Temp/astrocyte-w0-project-board-first-20261010.log`，所属进程已由helper退出。

当前组装首轮 `pnpm check` **PASS / exit0**：Go test/vet/mod、架构与依赖、进程停止、237契约例/生成漂移、14项真实HTTP contract_local测试、TS/lint、58前端测试及diff。原日志 `%LOCALAPPDATA%/Temp/astrocyte-w0-paper-check-first-20261010.log`。测试中合成Attention资料属于contract_local，不将这些例子混算真实论文/视频。

SQLite/foundation/cmd定向Go测试亦PASS。未启用opt-in时项目看板测试为 **0 PASS / 1 SKIP**，这只是默认关闭结果，不替代上述实际运行。

`startS1Server` 增加可选apiPort/webPort供W3使用主控分配18787/15173；默认随机保持，5173/8787与占用端口拒绝，不停止其他owner服务。该阶段等待W1/W2/W3最新精确SOURCE/REPORT普通合并及最终check/build/browser，当时未完整完成。

## 插件路线研究（选项，未决定）

[Chrome activeTab官方说明](https://developer.chrome.com/docs/extensions/develop/concepts/activeTab)明确：用户调用后当前标签临时授权，离开该来源/关闭标签后失效；不是全站读取许可。[官方消息机制](https://developer.chrome.com/docs/extensions/develop/concepts/messaging)提供应用页面向扩展的externally_connectable通路，需显式匹配应用地址、验证发送者和严格消息；content script内容按不可信输入处理。

现有服务只接受HTTP(S)同源/明确允许Origin与Strict人类cookie+CSRF，chrome-extension直接跨源请求不属于该身份。可选路线为应用同源人工审阅后提交，或主控确认后的资源范围配对+当前权限复核/撤销；未选择或写入任何配对、宽CORS接口。[summarize公开上游](https://github.com/steipete/summarize)有可选daemon，当前已锁0.25.1不因上游main变化升级或安装daemon；W1研究对应锁定源码/许可。

主控消息 `msg_aa4e0d8a97be` 已确认现有同源人类会话/CSRF作为基本审阅handoff的技术身份，不需要新增配对、宽CORS或daemon；这只批准运输准备。插件标签访问、Agent批量/自动读取产品选择仍pending，不据此启用读取或入库。

首失败、UNKNOWN与未执行项将保留在本报告，不以局部修后通过替代初始记录或整体完成。

新增首RED：W3在本轮真实看板浏览器规格的TypeScript编译发现 `startS1Server` 的 `.d.mts` 未同步新apiPort/webPort选项（主控转交消息 `msg_a18fd833fcd7`）。这是W0遗漏，保留首失败；已补声明，后续W3类型/浏览器验证是独立复验，不改写首RED。没有用cast或降低类型检查绕过。

## 实际完整联合输入（只执行一次）

主控 `msg_8e127ba9a790` 明确批准从既有人工允许的资料库制作一致隔离副本，进行一个新的完整论文+视频联合蒸馏。测试helper使用SQLite online backup和普通digest对象复制；原库只读，原成功任务、旧UNKNOWN、固定材料引用和全文保持。没有重抓论文/媒体、重克隆或新增Mission。新增运行器 `tests/s1/joint-capacity.mjs` 默认关闭，不因验收重跑模型。

- 实际运行精确 SOURCE `52ff9564c1b40109987668a95ed4ad35946d6847`；`ASTROCYTE_RUN_JOINT_CAPACITY=1 node tests/s1/joint-capacity.mjs` **PASS / exit0**。
- 原库 `%LOCALAPPDATA%/Temp/astrocyte-s1-4pmMWO/data`；一致副本及保存产物 `%LOCALAPPDATA%/Temp/astrocyte-s1-7NybyZ`，包括 `joint-request.json`、`joint-result.json`、SQLite和对象。原命令日志 `%LOCALAPPDATA%/Temp/astrocyte-w0-joint-capacity-first-20261010.log`。
- 论文 `IJNSLKOAR6LH6XT2KDZXJE6CCC` revision1全文95314 UTF-8 bytes，固定对象 `05acdc763e444fa7b9762df6dc2c879867d6a2e38d42884dcf3b80e07cf89d64`；视频 `3766OCH7SX3O5AVIDRPWYPCL3G` revision2全文32908 bytes，固定对象 `eb1106b97e3db4ddb6465ae066d1efe29affb809aa4841d52014ede3ab45a3fc`。
- 实际选择器在交付前测得prompt135376 bytes，原生SDK发送前测得含包装wire135754 bytes，低于授权524288；输出上限仍131072，未裁剪全文/引用/指令、未扩容历史。预检测量135045/135423属于另一次无模型assembly，不冒充本次实际交付计数。
- 实际 Codex0.162.0 / gpt-6.1-sol，job `ZWZSJOBRPFWWVZ2HL4TAVNBSAW` / operation `FGAEAIIOCQVPKEWW54PID54NTI` / result `N3PHOEWQXWAAA4MBXP6YF23YZ6`；21:56:05.135提交、21:56:53.793成功，attempts1、delivery_unknown=false。用量/成本仍unknown/null。
- 产出对象 `f562092fd236c6f5d46235e1a767814267472bec3ea791f0942545da6720e570` 实际保存，API内容与对象一致；新所属API进程同库冷重启后结果和两份全文仍一致，原库全部job行、旧UNKNOWN `4GQ7GKDXHZIND7CAB7GIZU3W4V` 和Mission均未改。

未触发本地发布故障，因此没有执行保存结果恢复，也没有重试模型。此证据只证明已选材料的一次实际完整联合输入与持久化；不证明全站Search、插件安装、科学结论或模型费用。

## 中间集成与本轮最终检查的范围

已普通合入已发布W1 SOURCE `172d278f72cc13e8f39ee2ea2f14278e42277616` / REPORT `b91e7fff46e20c706845148879708ac151abf3a7`，W2已发布领域 `be3cbfdbf536053f196ae5c48afad5fbf48138c2`（其前置`fc49bfa61d9399e6c3c31e2b7f239614bf4d6cdd`包含实际SDK测量），W3中间源码 `dfc4db88771e82d8cd999f9a097d62939852071e`，组装为967ed09。W1以整体未完成的failed生命周期报告交付有效局部成果；未将它改写为全论文Search/真实已安装插件完成。

967ed09：`pnpm check` **PASS / exit0**（Go/vet/mod/架构/进程、237契约示例与漂移、14真实HTTP contract_local验收、61前端测试、TS/lint/diff）；`pnpm build` **PASS / exit0**（Go与TS/Vite）。日志分别 `%LOCALAPPDATA%/Temp/astrocyte-w0-paper-check-final-20261010.log`、`astrocyte-w0-paper-build-final-20261010.log`。

967ed09：实际项目板 `ASTROCYTE_TEST_PROJECT_BOARD=1 node --test tests/s1/project-board.test.mjs` **1 PASS / 0 SKIP / 87.01秒 / exit0**。真实111根聚合71组；Codex扫描1802项/1714header/896匹配header/63根/119保留关联，Claude21项/7header/6匹配/6根/6保留，来源partial原因仍保留。其余来源unknown，不宣称完整CLI覆盖。人工字段安全、CAS、存储、刷新、冷重启、创建时间区别于活动、权限不变通过，保留库 `%LOCALAPPDATA%/Temp/astrocyte-s1-rdqrMn` 与日志 `astrocyte-w0-project-board-final-20261010.log`。

这轮默认广泛浏览器回归使用W0所属19787/16173，不开启paid/media opt-in；开始后主控通知W3仍返修，结果保留为中间源码回归，等待W3最终SOURCE及主控视口审阅后才发起适用最终检查。个人5173/8787未被替换或停止。

主控已接受W2最终切片：checkout SOURCE `4201c899f5aa0285a0d19442b674327237967739` / REPORT `9e47c455ba07d83a66696b5bfeccf46ffff4f33c` 已push，其最终实际API/SQLite/缓存/冷启动1 PASS（44.87秒），库 `astrocyte-s1-meg3Pi` 保留。业务源码与上述W0已集成W2部分一致；后续复用该实际证据，不再次刷新采集以制造额外绿灯。

967ed09的默认浏览器回归首轮完成：`pnpm test:e2e --workers=1 --output=test-results/s1-w0-paper-final-20261010` **158 PASS / 4 FAIL / 22 SKIP / 7.7分钟 / exit1**。四个首RED为1920/1280两尺寸各一次 `navigation.spec.ts:31` 旧工作区标题不存在，以及 `:513` 三个Mission错误alert处于折叠记录面板内不可见；原日志、截图、trace与error-context保留，不宣称全套通过。W3收到 `msg_f2668ac7d1d8`，确认调整为项目总览/接入设置导航，并打开已有任务记录后仍验证三个错误传播，不移除断言或改用fallback。等待最终owner源码后只跑受影响用例及适用最终check/build，不重跑无关全套。

本场停止后已普通合入W2上述SOURCE与REPORT，push组装 `38d0b4f3ef89abc23412fe6891f03d9625dd0bbb`；没有新业务变化。所有W0本场所属测试进程按helper退出，个人预览不动。

新增真实领域遗漏（主控 `msg_4bf502e148a8`）：已有 `Git.LastCommitAt` 未进入聚合 `LastActivityAt`，原生独立项目可能误显活动unknown（中间观察只有7/71有已知活动）。Git详情证据存在不等于汇总正确；原owner W2重开定向域/测试任务修复，W0不越域修改。会话创建仍不能作为最近活动；最终完成等待此修复及现有缓存GET验证，避免重复发现扫描。

## 最终集成、复验与主控接受

按已发布精确提交普通合入：W1 SOURCE `172d278f72cc13e8f39ee2ea2f14278e42277616` / REPORT `b91e7fff46e20c706845148879708ac151abf3a7`；W2定向修复 SOURCE `cd201962997e69bb4772d73a6d8752874085c3ef` / REPORT `9319374f4a55cf88901e2ed77b439f45aaea1b89`；W3 UI/测试 `2aa57e264aa56a7d7f271c5e087bc540f3f862ab`、最终消费W2的SOURCE `90cee4eda3c5f069325751064ecc944a303f6f77` / REPORT `1c7dfa84e8ac927cf5485e28b494219504083dd2`。最后一个W3报告普通合并为最终W0 SOURCE3b5a434，无越域业务胶水或历史覆盖。

主控已查看接受最终紧凑1920/1280/390布局，W3实际字段流固定c07、最终布局固定acbfc的证据保持独立，详见 `docs/acceptance/S1-paper-board.md`。W2修复仅5行领域逻辑：取非空Git提交与既有活动最大值，header创建仍独立；原领域RED、Windows SQLite URI首RED与复验分开保留。W2既有缓存应用服务7→17已知活动、0扫描/0写入的证据保留，原人工metadata adapter为模拟，未将该证据冒充实际人类记录/API。

| 本轮最终命令 | 精确源码、结果及原件 |
|---|---|
| `pnpm check` | 组装`0e55f141d2c75eb48808002df6e047ff235f7aab` PASS/exit0：Go/vet/mod/架构/进程、237契约示例/生成漂移、14真实HTTP contract_local场、61前端例、TS/lint/diff；`%TEMP%/astrocyte-w0-assembled-check-20261010.log` |
| `pnpm build` | 同0e55f14 PASS/exit0：Go可执行程序、TS/Vite；`astrocyte-w0-assembled-build-20261010.log` |
| `pnpm test:e2e --workers=1 --grep 'navigates to 共同工作区 page\|swarm propagates mission error' --output=test-results/s1-w0-paper-owner-retest-20261010` | 同0e55f14 **4 PASS / 33.1秒 / exit0**，1920/1280导航与三处可见错误传播，属于受控接口失败回归；`astrocyte-w0-assembled-browser-retest-20261010.log`。中间全套4RED不改写，没有重跑无关全套 |
| `$env:ASTROCYTE_TEST_PROJECT_ACTIVITY_HTTP=Join-Path $env:LOCALAPPDATA 'Temp/astrocyte-s1-rdqrMn/data/state.sqlite'; go test -mod=readonly ./tests/s1 -run '^TestActualCachedProjectActivityHTTP$' -count=1 -v` | 新HTTP测试SOURCE `da515b826a4d21c0bff872bafa3bb32120d99418` **1 PASS / exit0**。实际111根/71组、17已知活动、1实际人类记录；`astrocyte-w0-cached-http-retest-20261010.log` |
| 新HTTP测试首场 | SOURCE `eebaf2820966623fe5e012d110039e66d09260bb` **RED**：预查询误按grants不存在的id列排序，在HTTP请求前失败；改按首主键列排序，原日志 `astrocyte-w0-cached-http-first-20261010.log` 独立保留 |
| 加入HTTP测试后的 `pnpm check` | da515b8 **PASS/exit0**：完整适用检查、237契约例、14 contract_local例、61前端例、TS/lint/diff；默认无路径的actual-cache例明确SKIP，不替代上述显式实际PASS；`astrocyte-w0-final-source-check-20261010.log` |

新测试 `tests/s1/project_activity_http_test.go` 复用正常SQLite仓库、真实application服务、HTTP/session中间件及真实loopback端口。已有批准临时SQLite连接设置query_only，不启动main或任何attention/native/discovery worker，配置无source的真实仓库，执行正常human bootstrap和一次缓存GET。响应实际人类notes/review/group/intent/archive/revision与各自SQLite记录逐项相同；查询前后发现JSON/人工记录/grants/jobs行完全一致。没有模拟metadata adapter、重复源采集或模型。所属HTTP listener在测试结束Shutdown完成。

测试/产品源码与文档映射：0e55f14→da515b8仅增加并修复上述HTTP测试；da515b8→3b5a434仅W3最终报告/任务文档变化，运行代码和所有测试未变。因此build/4项浏览器复验沿用0e55精确产品代码，最终check与实际缓存HTTP沿用da515精确测试代码，不为文档合并重复运行。主控消息 `msg_17c41d57fb80` 已接受实际缓存HTTP17/71与原SQL RED，并要求消费最终W3报告后结算；个人预览升级的实例验收由root负责。

## 剩余限制与未执行操作

- Search范围/当前页提取、插件访问与Agent批量行为、进度推断仍待产品决定；未冻结新接口、扩大CORS/配对或从发现授予权限。W1已有DOM/元数据适配和许可材料不等于已安装可用MV3插件、全站搜索或PDF/所有网站支持。
- 采集仅已批准已知元数据目录、有界header与固定项目标记；来源partial/unknown保留，不证明全部CLI、当前活跃Agent、配置/可启动/原生能力或科学完成。
- 一次实际完整联合模型调用成功，但费用/原生usage未知；旧UNKNOWN保留。未人为注入发布故障，因此未实际验证本次模型结果的恢复重试路径。
- 不安装插件，不更新个人5173/8787运行实例，不读取个人Chrome，不重启Windows；安装与个人实例切换由root负责。人类最终产品选择和验收不由代码测试替代。
