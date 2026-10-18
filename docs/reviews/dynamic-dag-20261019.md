# 动态 DAG 独立评审

2026-10-19。Reviewer使用Codex CLI 0.162.0，Orca终端元数据为GPT-6.1-Sol xhigh fast，继承原配置；Go实际版本1.27.2。独立分支：`songconmaisaix31-design/dynamic-dag-review-20261019`。

## 决定与准确提交

| 范围 | 决定 | 独立证据 |
|---|---|---|
| 依赖DAG、人工任务生长与来源、当前读取权限、双视口键盘、断线重连与历史 | 接受已验证构件 | 657真实基线9场景、b0真实17场景、最终982业务再次独立17场景；生产HTTP/SQLite/DOM和逐项事件回查 |
| 局部按需加载、首屏JS减少、真实缺分块与人工重新加载 | 接受 | 同一准确982 Go程序及同SQLite先后提供b0和候选真实产物，四个冷启动context全部JS实测，两个视口真实缺文件与同context恢复 |
| 完整模型生成及带Worker的停止、退出、结算与释放 | 证据不足 | 实际工作项全部human创建，原生会话及租约0；原生工具关闭及S3门禁保持 |

本报告实际验证的DAG和加载构件没有要求继续返修的已确认未解决缺陷。接受范围不能作为全部动态DAG、完整CI或真实Worker生命周期完成的结论。

- 最终PR候选：`7ce3a5efe312f4cdac3491293e002f5559f076b0`。
- 实际业务验证源码：`982333b871520e606c74d553e448c5049f81abde`。已完整核查982至7ce3仅标准报告的CI事实修正，业务、测试、锁和AGENTS不变；ff-only更新后未重复相同业务检查。
- 动态DAG前基线：`657ca1464517a171ed50ddc8fcc787b237c44ad9`；首个已实测DAG候选：`b0a21a2cd03c074ba4097012be8f7e7748bb1f3d`。
- 本人实际发布的最终[PR #2 CR 5564387472](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5564387472)：UTC2026-10-18T18:52:14Z，commit_id准确7ce3，state=COMMENTED。当前账号为PR作者，本次COMMENT review没有原生APPROVE。此前准确b0的独立[CR 5564261919](https://github.com/songconmaisaix31-design/Astrocyte/pull/2#pullrequestreview-5564261919)也已实际发布。

## 实际可用性与优化效果

固定来源为已获准的Agent创建公开Astrocyte项目，正文为真实Git对象 `657ca1464517a171ed50ddc8fcc787b237c44ad9:README.md` 的完整原文，经生产manual接口保存，并保留对应公开GitHub精确locator。项目、目标、提案、批准、Mission与工作项均由实际human会话经生产接口创建。没有响应替换、伪API、伪会话、模拟退出或直接写业务表。

657基线DAG接口实际404、入口禁用；b0及982实际200。最终982从两节点/一依赖经外部实际API新增变为三节点/两依赖，原选中节点及DOM身份保留；本次新增显示观察为4956.95ms，与既有5秒读取周期相符，不作速度提升结论。创建主体、原创建事件、依赖来源、任务版本和固定输入可读；执行往返事件与任务依赖分别保留。

最终17场景还覆盖：真实空集合；缺失依赖404且事务无部分写入；CSRF/跨源403；实际签发凭据仅read_context时403，加入read_swarm后200，另一公开项目凭据读取本项目403；1920×1080、1280×720的布局方向/缩放、节点Enter、工作项详情及Escape焦点返回、长中文与页面横向范围；真实停服后过期快照和禁用动作、同库重连、撤销当前read_swarm清图及重新授权恢复。资料、提案批准、作品入口及OpenViking未配置禁用回归通过，页面错误为空。

生产生命周期按已批准的一秒无进展条件将没有Worker的Mission置为blocked/no_progress_time_limit，下一周期无重复事件，同SQLite重启事实相同；该场景没有生成Worker退出事实。环、重复、自依赖、缺端点和跨Mission属于前轮实际Domain参数测试；这些规则检查没有被提升为原生执行的正向证据。

加载比较使用同一准确982生产Go程序、同一独立SQLite，先后提供Reviewer保存的准确b0静态产物及本树候选产物；明确隔离静态资源变化，没有将b0产物阶段称为原b0后端。两个视口分别为attention、workspace创建全新context，记录全部实际JS响应及浏览器资源条目。

| 全量JS测量，单位bytes | b0产物 | 候选产物 |
|---|---:|---:|
| attention/workspace冷首屏全部JS，两个视口四组均相同 | 774329 | 719784 |
| 第一次进入Swarm追加JS | 0 | 57129 |
| 直接进入Swarm的全部JS | 774329 | 776913 |
| 完整生产构建全部JS | 774329 | 776913 |
| 完整构建全部JS gzip，同一zlib默认参数 | 216760 | 217799 |

实际JS响应均200、HTTP编码identity。b0为 `index-DpvBzN0-.js`，候选为 `index-DyOZ-u9Q.js` 和按需 `MissionDAGView-DJuVNnnW.js`；首次Swarm真实请求后者，重复进入已加载模块不再请求JS。冷首屏减少54545bytes，完整JS增加2584bytes、同口径gzip总量增加1039bytes。gzip为离线计算，未充当实际压缩传输量；Vite的gzip提示单独保留。没有全站总量下降、速度提升比例或内存下降结论。

真实缺失场景先核本树生成文件绝对路径，再临时移动实际DAG分块，并在finally原样恢复。两个视口均实际404/Cache-Control:no-store及net::ERR_ABORTED，局部显示含实际URL的导入错误，原资料页继续可用；同一context恢复文件后，用Enter触发普通整页重新加载，真实重新请求200 JavaScript并恢复两节点图。没有清缓存、更换恢复context、修改mtime或拦截响应。恢复文件SHA256为 `d1a063c74b54e84b1156a874fbb2a257e59c1692e329f79d8d3b8bc06bfa6291`。

实际 `/attention`、`/swarm`、`/workspace/works`、`/assets`、`/assets/` 正文与对应完整构建index相等；缺构建文件GET/HEAD均404/no-store，HEAD无正文。真实human cookie访问health为200 JSON、bare API为404 JSON。静态修复仅匹配os.ErrNotExist和/assets/前缀，保留API、存在目录及普通stat错误行为。原断言、默认workflow未修改；按需加载阶段未修改依赖和锁文件。

## 本人执行的命令与原件

以下五个phase均在准确982执行，经本树 `.verification/dynamic-dag/run-review.ps1`，参数为 `-SourceRoot C:/Users/DW/orca/workspaces/Astrocyte/dynamic-dag-review-20261019 -SourceSHA 982333b871520e606c74d553e448c5049f81abde -ExpectDAG`，每阶段核准确HEAD和资源。

| 实际命令 | 结果 | 原件，位于.verification/dynamic-dag |
|---|---|---|
| `pnpm --dir web build` | EXIT0，类型检查和200模块生产构建通过；原大于500kB提示保留 | web-build-20261019-024345-828.log |
| `pnpm --dir web lint` | EXIT0，零诊断 | lint-20261019-024411-898.log |
| `go vet -mod=readonly ./internal/adapters/httpapi` | EXIT0，零诊断 | http-vet-20261019-024428-360-resource.json；命令无输出 |
| `node .verification/dynamic-dag/actual-review.mjs` | EXIT0，17场景；原helper实际生产Go构建/HTTP/SQLite/顺序双视口 | actual-20261019-024446-767.log |
| `node .verification/dynamic-dag/loading-review.mjs` | EXIT0，四组真实加载/资源/缺失恢复场景，含原目录/深层/API | loading-20261019-024548-248.log |
| `gofmt -d internal/adapters/httpapi/server.go` | EXIT0，无差异 | 实际工具回执 |
| `git diff --check b0a21a2cd03c074ba4097012be8f7e7748bb1f3d 7ce3a5efe312f4cdac3491293e002f5559f076b0` | EXIT0 | 实际工具回执 |

此前在准确b0独立执行冻结锁安装、build/lint、四包9个DynamicDAG函数、contracts264例和生成一致性、architecture43包及17真实场景，均EXIT0；contracts保留13项既有警告。原件在本树 `decision-b0a21a2c.md` 及对应日志。未将这些前轮结果称为在最终提交重新执行。

实际17场景原件为 `.verification/dynamic-dag/actual-982333b87152-1792349091168/result.json`，数据库位于其下 `astrocyte-s1-0OYiON/data/state.sqlite`，保留完整服务日志。生长Mission为 `857d2d03-9490-4df9-8338-46cecde27449`，停止Mission为 `beeb415c-6e55-45c5-a990-a5af91d1fa16`。

加载比较原件为 `.verification/dynamic-dag/loading-982333b87152-1792349152708/result.json`，两个阶段使用同一 `astrocyte-s1-4f4hfA/data/state.sqlite`。Mission为 `89426764-3d7f-443a-9b49-2e663a25e24f`，固定资料 `TJL34N4UMHWXJR6TLOGEVBLI5P@1`。原始SQLite、完整JSON与进程日志仅保留在本机忽略目录，没有随报告提交。

以下9个唯一event_id均经真实HTTP返回，并逐项查询foundation_outbox，核类型、aggregate_id、版本、时间及payload；加载比较的同三个事件在前后阶段分别回查，未重复计为新事件。

| event_id | event_type | 对象及版本 |
|---|---|---|
| dd860573-a6c3-48cf-893d-ea5d34ab32cd | swarm.mission_created | 生长Mission v1 |
| e882e57e-1aa6-4d57-ad90-6845ae292d14 | swarm.work_item_created | e0f21fcd-a879-4daf-b837-6b6c119a4f3c v1 |
| b64b8361-2fb5-4fe9-9e05-8ae0d90b17c9 | swarm.work_item_created | 8d32e831-341f-4577-9d6a-5e0ef3db9ea0 v1 |
| e4344d4b-4dfb-4375-9b6b-8ad52a7e76b5 | swarm.work_item_created | a0c2664c-7ec1-4533-8561-d1793b5dbc53 v1 |
| 74ff70b9-31f7-48dd-87a6-9a7e95d7f39c | swarm.mission_created | 停止Mission v1 |
| 7b570951-22b5-4ede-87b4-5d991f1c3c44 | swarm.mission_stop_requested | 停止Mission v2 |
| 6ff0994f-6ae3-42bf-98fb-147bf53cdd34 | swarm.mission_created | 加载比较Mission v1 |
| edaef0f2-88f1-49de-b882-ec750044a7c8 | swarm.work_item_created | 23e8aeaf-396f-475e-b9c8-3519a347480b v1 |
| 2b7bcbef-23f3-480d-8e58-ee0637ad0e9a | swarm.work_item_created | 92cc14de-8d93-4b59-be11-282852c16ab9 v1 |

## 完整CI的实际失败与边界

本机未执行完整默认CI。准确982自动触发[push run39155403418](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/39155403418)和[PR run39155406045](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/39155406045)，安装及pnpm check实际开始；主控按范围停止，Reviewer用gh独立核两个run均completed/cancelled。取消前Ubuntu job117757474469及117757483047的pnpm check已failure，另两个job为cancelled，后续独立race/build/Playwright步骤为skipped。

本人只读失败原日志确认：首个失败命令为 `go test -mod=readonly ./...`，此前gofmt、全仓vet通过，pnpm check内后续契约和前端步骤未到达。日志原件为 `ci-982-push-job117757474469-failed.log` 和 `ci-982-pr-job117757483047-failed.log`。没有重新执行历史fixture或放宽断言。

具体来源核查如下，未将所有错误笼统归为环境或本轮回归：

- push首个失败为Relay旧fixture的events/reconcile期望200，但其AllowedActions缺reconcile，现有permission按动作拒绝403。测试、相关HTTP路由、permission及ReconcileEvents从准确657至最终HEAD未改；测试Config.WebDir为空。
- PR另有Git.Branch实际nil与旧测试强制非nil不符；未据此推断本轮Git功能回归。
- 旧localNative是offline-protocol-fixture计数stub，未注入真实controller reader和object store。从未变源码推断SendPrepared会先被nativeMessageServices拒绝，计数不等于1的OR断言不能证明发生真实重发；未把该测试提升为真实原生退出或重发证据。
- 迁移测试仍断言26，但准确657已含27份SQL和027_openviking；相关测试/迁移未改，属于可定位的既有不一致。
- 其他日志含缺明确自有输出目录、合法公开原件、CLI不在PATH等实际前置失败。尚未逐项审计全部历史测试及失败归属。

上述检查没有发现据此要求本轮DAG/加载返修的已证实新回归；完整CI仍未通过。7ce3仅按主控授权以[skip ci]文档提交修正CI运行事实，未修改workflow、业务或原断言。

## 真实决定通知与主控确认

实际Task为 `task_40fbc675246c`，Dispatch为 `ctx_5a4a6eff08a2`，Run为 `run_73578b70c449`。全部决定使用仓库Node review-decision hook、实际Orca可执行文件与准确身份，均EXIT0；原件位于本树.verification/dynamic-dag。下表保留实际消息，主控读取回复与单纯入队分别记录。

| 决定范围 | 实际hook消息 | 主控实际确认 |
|---|---|---|
| 657静态证据不足 | msg_5e5595bbaf68 | msg_0a59dd232fa7 |
| 16dd通信构件接受 | msg_da0b5fae07e9 | msg_c22963d7c3c0 |
| b0已验证构件接受 | msg_994d0edd71ab | msg_a59990de20ac |
| b0完整链路不足 | msg_e351e28845ca | msg_35162ad7e2a7 |
| 最终7ce3已验证构件接受 | msg_fcc2af9debd9 | msg_744d85cce321，已完整读取本地报告与准确GitHub CR并确认唤起 |
| 最终7ce3完整链路不足 | msg_819601d74f38 | msg_76d5bff6b2fd，已实际读取不足hook、最终报告及CI原日志范围并确认唤起 |

主控msg_744d85cce321确认可验证的前端/动态投影/按需加载规划结束，授权本报告单个Commit和普通push；无新增优化、子Agent或重操作。报告只提交到Reviewer既有分支，不创建业务PR。

## 资源与真实剩余限制

最终五phase共10个启动采样，间隔约3.6秒，最高CPU36%、最低可用内存7.7397GiB；BelowNormal，Go GOMAXPROCS=1/GOFLAGS=-p=1/GOMEMLIMIT=256MiB，Node512MiB，一个浏览器且context顺序执行，最多一个服务，复用D盘缓存。启动采样不代表运行期峰值。

实际17场景初始PID44824工作集76574720bytes、CPU累计0.40625秒；后期已为重启后PID46376，未据此推断内存下降。加载比较初始PID50792/39100均实测BelowNormal，不同进程和初始化阶段的CPU/RSS不作为优化效果。

自有浏览器及服务均已由原helper关闭并等待退出，61314/65342/49321监听0；真实分块原样恢复。msg_c7d8419baa09已明确交还重操作窗口。没有关闭其他应用，没有视觉、截图、视频、trace或像素读取。

- 原生工具关闭及S3调用预算、工具/外发目标证据门禁仍未满足；本轮没有启动产品原生Agent或新模型输入。真实规划/生成、Claim/Send、带Worker停止/进程退出、租约过期、暂时失联、Stop/Recover并发、未知用量结算和资源释放均未验。
- 退出事实去重、旧字段兼容、错误分类和停止/结算原因分离的带Worker分支，只有静态与参数边界证据；未构造假claim、lease、session或退出事实。
- 审批和作品仅回归实际页面与入口；真实执行、完成、采用及后续经验复用未计通过。OpenViking没有合法远端配置，真实提交/任务完成/未知副作用/检索/分层读取仍未验。
- 完整CI未通过，包含实际失败和取消；失败归属未完成全量审计，本机未重跑完整默认CI。
- 没有合并PR、共享部署、打包或发布。剩余原生条件及完整正向路径需要相应用户决定、合法来源与后续明确授权。
