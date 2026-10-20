# SWA 当前状态

本页统一维护当前版本、已验范围和实际限制。资源参数只在 [RESOURCES](docs/RESOURCES.md) 维护，授权和待答项只在 [QUESTIONS](docs/QUESTIONS.md) 维护，接管操作见 [HANDOFF](HANDOFF.md)。

## 当前主线与 M0

| 项目 | 当前事实 |
| --- | --- |
| 用户首版目标 | 抽取 Orca CLI 对应执行能力，交付用户从原有 Agent 或终端调用的可控 Swarm CLI；[首版任务](tasks/swa-cli-v1-20261020.md)记录具体范围与验收，程序实现中，尚未通过验收 |
| 首版执行与验收 | 已确认 Orca 无界面执行及控制端身份；固定抽取范围后定义 CLI 规则，完成多 Worker、失败/资源及安装包。最终使用安装后的同一候选完成一次至少两个真实 Worker 的任务，验证并发、指定停止及成果获取；产品采用独立后台，须验证不妨碍本机Orca CLI及其原配置/状态/进程 |
| 规范仓库 / 默认分支 | GitHub 仓库 ID `1410836101`，`songconmaisaix31-design/swarm-anything` / `main` |
| main 的产品与治理集成来源 | `6eb7c7f640a9cf1b05ea365b3afdf6883cd01765`；已从初始 main 普通快进并 push，完整包含产品32ab历史与接受后的M0文档；后继当前事实记录沿 main 维护 |
| 产品基线 | `songconmaisaix31-design/works-relay-integration@32ab1eec808a5848aabad40f11133735a1214e55`；原 PR #2 已普通合入 |
| M0 任务分支 | `swa-m0-20261020@6eb7c7f640a9cf1b05ea365b3afdf6883cd01765`；独立 [COMMENT review 5581860208](https://github.com/songconmaisaix31-design/swarm-anything/pull/4#pullrequestreview-5581860208) 接受，原hook已由主控读取并回复；[PR #4](https://github.com/songconmaisaix31-design/swarm-anything/pull/4) 已实际合入产品分支，同源快进 main |
| 主控内容来源 | `validation/v1-20261012@cf22f118193d5dcc3f2d2ec436bd4f06e9a22c9b`；独有决定、归档、任务、实验入口与研究文档按[整合清单](docs/integration-audit-20261020.md)纳入 |
| 已审调查记录集成 | `fac75c7f963d16c7f112e8e90ad30e36510d4738` 已独立接受并普通快进/push main；只新增准确30b来源调查、b112契约记录和34cfe独立报告共八份文档，各自接受范围保持，程序/依赖未变 |
| 后继工作 | 同一无界面Worker以Task9f/ctx0d接续。准确c29普通CLI、f08锁/真实类型及c0所属EOF/Job/native关闭已获独立6d693df7局部静态接受；主控完整读取两报告并分别回复四项hooks。首次build-01在mkdir处ENOENT/EXIT1，无部署或实际编译/候选；准确`bc217de88f4bb4b8af1ab63f3c281f4c2f649aea`修正受控dist父目录，独立97a04f01已局部静态接受。后继`897531d5d2978e3a51a043063a26e254b3d0f766`仅两文档提交T12/strict caller完整提案及R01—R09，待独立审核和SPEC接口定稿，未实施；两提交Worker报告普通push EXIT0，主控已完整读差异并通过空白检查。当前仅交一次干净897531的build-02，实际结果待回执；C01—C09/helper与T12仍未完，T11范围待用户回答。功能、安装、模型及双Worker/隔离动态验收均未通过 |
| 当前独立审查归属 | 同一Reviewer Task `task_1cd8014750f9`以原生retry-of恢复到Dispatch `ctx_6339d8beeb73`、terminal `term_9e9193d2-733a-4ab1-8220-22a5dc640c0f`，同一`swa-mainline-review-20261020`工作树/分支，两份原报告写域保持。旧ctx980的前台Agent已返回PowerShell，提醒被shell拒绝；原生精确stop确认PTY退出后才重启，新preamble实际turn_started。既有27f报告保留；新邮件已入队，实际读取确认及新结论待回执，不将启动当成审查通过 |

当前 main 工作树为 `C:/Users/DW/orca/workspaces/Astrocyte/main-20261020`；原根目录的 `validation/v1-20261012` 与正式实验材料保留。Orca 的新工作树默认基准已实际改为 `origin/main`。M0 的接受仅覆盖文档整合，首版 CLI 程序实现中，尚未通过验收。

## 已有产品验证与最近运行观察

程序和 Web 的已验源码 S 为 `ed543c12106c2dbb3e181cb9bc1647b626455179`。两次CLI依赖窗口均已完成原S保全、停止和同源恢复。最近原件位于根忽略目录`.verification/swa-cli-v1-control/prepare-03/`、`stop-03/`及`restore-03/`；首次原件`prepare-01/stop-01/restore-01`继续保留。原件含实际环境、Q/媒体/CodeGraph配置、原库一致备份及停止区间内容相同结果；备份包含真实数据，不进入Git或安装包。M1原件保留原位置，其已验范围见[产品交付报告](docs/standards/delivery-candidate-20261020.md)及[交付任务](tasks/delivery-candidate-20261020.md)。

| 范围 | 记录的结果 |
| --- | --- |
| 当前服务与窗口 | 原S45948已按prepare-05保全，于UTC `2026-10-20T14:26:46.673512Z`精确退出15，当前停止；stop-05新资源检查/停止后一致备份及window-ready通过。唯一窗口只交原Worker对干净897531执行一次build-02，完成后使用本次prepare-05/stop-05同一原库/配置恢复准确S。停止后原绝对Orca CLI实际公开status仍ready/reachable/connected、graph ready，app PID38652及runtime `49da37c6-51eb-4061-b225-00ce0e551430`保持；本次没有启动产品候选服务 |
| 同源产物 | 程序和 Web 同一 S；`D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/versions/ed543c12/` |
| 原库与配置 | 原运行根 `D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/` 下 `local/astrocyte-workbench/v0.2-access/app-state/state.sqlite`；当前备份完整性/外键及停止期间内容一致性通过，恢复使用该原库和原配置 |
| Windows 普通交付 | `dist/delivery-candidate-20261020/`，三个 Go 程序、Web 与启动入口；完整新便携模型包未验 |
| 已通过的限定检查 | 构建、DAG9函数、OpenViking4函数、前端53项、DAG双视口4项、页面8项、架构/依赖、264项契约示例及生成一致性 |
| 独立运行核验 | 准确 S 的49项 GET（含5项403）、真实资料/项目 ID 与原 SQLite 对应、33件 Web 字节、原业务和 Q 保全 |
| 原始失败与未执行 | 完整默认 CI：NOT_RUN；原 `pnpm check:contracts` 与 Redocly 直调：进程崩溃 FAIL；lint/vet 的实际串行偏差和退出保留在原报告 |
| 本次 M1 原入口 | 准确 `987811f417e92cd6f15961e60c279f616e3b21a8`，原依赖冻结安装与未补丁的 `pnpm check:contracts` 均 EXIT0；13项警告、264项示例及类型一致均实际完成。Node24.16.0、pnpm11.27.0、CLI2.60.0；历史崩溃未复现，未应用补丁，T02—T08未执行，完整修复因果证据不足 |
| 首版CLI依赖准备 | pnpm11.27.0的`install --lockfile-only --ignore-scripts`及`install --frozen-lockfile --ignore-scripts`均EXIT0，私有store/TEMP/TMP/cache，两次新资源门限通过。实际入口签名拒绝、peer警告及镜像超时/原命令内建重试记录保留；尚未构建或调用模型，不能计为安装包和CLI功能通过 |
| 第二次窗口准备 | 根忽略目录`prepare-02/backup-resources.json`记录UTC11:00:16、CPU7.2%、可用内存3.529GiB；低于4GiB门槛，prepare命令EXIT1，未备份完成、未停止原S、未交窗口。原件保留于`.verification/swa-cli-v1-control/prepare-02/`；后续依赖/生成结果见本次窗口，候选构建未执行 |
| 第三次窗口准备 | UTC13:22:33/36的prepare-03两次真实阻塞采样为CPU19.0%/5.2%、可用15.719/15.746GiB，单调间隔3.005秒、BelowNormal；命令EXIT0，准确S12800及实际环境/配置核实、原库一致备份完整性/外键和prepared标记通过；随后精确停止并交付两Stage，本机Orca保持原运行 |
| 本次依赖/生成窗口 | stop-03实际EXIT0，S12800于UTC13:29:19退出15；停止后双样本CPU43.2%/16.6%、可用12.918/12.808GiB、间隔3.003秒/BelowNormal。原Worker串行执行`next-window.ps1 -Stage dependencies -EvidenceName dependency-02`和`-Stage generate -EvidenceName generate-01`，锁同步/冻结安装/真实类型生成及各自资源检查均EXIT0；pnpm原命令内建镜像超时重试及peer警告保留。仅锁/生成类型提交f08，实际原件在Worker自有忽略目录。没有构建、功能测试、安装或模型调用 |
| 本次同源恢复 | restore-03实际EXIT0，启动前双样本CPU18.7%/6.8%、可用10.369/10.485GiB、间隔3.002秒/BelowNormal；原库integrity为ok、外键为空、stopped_interval_comparison为equal，恢复使用同一原库。实际S43068及原Orca核查通过。控制脚本两项实际进程身份更新SHA256为`8e06639aa6c0bf635b037fb82a8b745540ed0dd3a754b15e8220d098652b606d`，已获准确报告ebd7880d独立接受，主控完整读取并实际回复hook；没有据此授予产品功能通过 |
| 第四次构建窗口 | prepare-04实际EXIT0：UTC14:09:55/58 CPU16.4%/49.2%、可用10.706/10.627GiB、间隔3.004秒/BelowNormal；准确环境/配置及原库一致备份通过。stop-04 EXIT0，S43068于UTC14:10:08退出15，停止后双样本及备份/window-ready通过。干净c0仅执行一次`next-window.ps1 -Stage build -EvidenceName build-01`；资源EXIT0，构建EXIT1/ENOENT。mkdir候选目录时缺dist父目录，未执行部署/Node下载/native/TS/Go编译，无VERSION；Node48860实际退出1、同步版本命令已返回及两次定向进程核查完成后交回窗口。原失败与MODULE_TYPELESS警告保留，未自动重试 |
| 第四次同源恢复 | restore-04 EXIT0，UTC14:14:43/46 CPU4.8%/9.8%、可用10.709/10.650GiB、间隔3.002秒/BelowNormal；实际同源S45948和原库/环境/配置/health/Web及本机Orca核查通过。控制脚本两身份字段SHA256为`6d7bed3576ade7b9522738ba741a55eade309f9c8b9de93a869b3150562d6870`，独立97a04f01已接受，主控完整读报告并实际回复；Reviewer创建时间显示精度检查两次失败及实际epoch数值通过原件保留，没有修改控制脚本/原件 |
| 第五次构建窗口 | prepare-05 EXIT0：UTC14:26:20/23 CPU14.8%/10.1%、可用10.887/10.931GiB、间隔3.003秒/BelowNormal，真实环境/配置/一致备份通过。stop-05 EXIT0，停止后UTC14:26:48/51 CPU17.5%/18.0%、可用10.930/10.979GiB、间隔3.004秒。仅一次`next-window.ps1 -Stage build -EvidenceName build-02`交原Worker，准确897531干净、build-02及candidate01不存在已核；tracked和Git冻结，T12草案只在互斥忽略目录准备。成功后仅实际版本核对，其他测试/安装/服务/信号/模型未授权；当前build结果未验 |
| 后继只读资源观察 | UTC11:12:58的系统只读查询显示可用2.49GiB；UTC11:15:17的进程查询显示StartMenuExperienceHost/PID15952工作集11.62GiB。这两项为单次观察，未代替双样本准入，也未再次执行prepare或操作该非所属进程；原服务及本机Orca保持运行 |

完整默认浏览器、真实原生/S3、带 Worker 停止退出、未知用量对账和释放、两 Agent 闭环、成果采用、OpenViking 真实远端正向均未通过完整验收。正式 Release 未创建。原生控制器工具关闭、真实事件逐项定位、预算与权限证据门禁保持。

## 科研与正式实验

| 范围 | 已有结果及限制 |
| --- | --- |
| 科研预印本 | 独立分支 `songconmaisaix31-design/research-preprint-20261019@fcb1e0a70541e15de1f52507a43edadb9801e368`；方法稿已有接受范围，完整 TeX 安装/构建/验证、新引擎产物及独立重建未完成；正式实验结果、作者/机构/许可未确认，原 [PR #3](https://github.com/songconmaisaix31-design/swarm-anything/pull/3) 的证据不足范围保持 |
| 研究与产品边界 | [研究反馈](docs/research/research-product-decisions.md)与[科研 Skill](.agents/skills/astrocyte-research/SKILL.md)已保留；研究运行时、正式实验代码和论文原件仍在原位置，不作为产品执行权威 |
| 正式 V4 | 上次接管核查中 `C:/Users/DW/AppData/Local/astrocyte-validation/` 的原证据根、外部权威锁和正式 SQLite 不存在，无法核实锁/day0/路由/授权；本轮没有重新核查或补采。实际人评与终局未完成，按[日运行任务](tasks/M6-V4-daily-operation.md)和[实验入口](docs/experiment/README.md)接续 |

完整历史状态见[20261020归档](docs/archive/20261020/README.md)，本轮状态入口迁移的产品原文见[M0归档](docs/archive/swa-m0-20261020/README.md)。历史任务中的旧进程、派发和额度按原核查时间阅读。
