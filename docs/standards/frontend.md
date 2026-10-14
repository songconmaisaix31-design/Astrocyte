# 前端标准整改与合并产品验证

分支 `standards-frontend`，当前本域 SOURCE `fcf5f55012ba0b05535fe851606cbf3c5a9f3f87`，已普通 push。客户端实际为 `codex-cli 0.162.0`，本次 Orca 终端观测 `GPT-6.1-Sol high fast`；沿派发既有 `--yolo` 配置。本轨没有另派 Agent，没有读取凭据、私人原件或正式实验库。

## 完成内容

前端移除 `vi.fn` / `vi.mock` API 模拟、`page.route` / `route.fulfill` 伪造响应及 fixture 事实展示。历史 `fixture=1` 仅作为回归查询参数，页面继续读取实际服务。纯函数测试明确参数输入，不声明检索、模型或原生操作已经执行。读取、命令、生成客户端和会话复用现有实现。

实际验证使用既有生产 Go 服务、真实 SQLite 和真实 HTTP。失败条件使用权限拒绝、输入校验、取消、网络断开、进程退出及实际系统端口拒绝。F 创建的产物保留在本树 `.verification/`，不会覆盖首失败。

插件输入直接采用 B 的生产 `paperDOM` 对获准公开 HTML 生成的完整对象，保留 schema、来源、正文和原始 JSON 附件。缺少或非法 schema version、非法文本 / 列表 / 可得状态 / 截断字段立即拒绝，正文首尾空白完整保留。公开论文两个实际版本经生产手工导入，来源、不同正文、版本切换、固定引用及人工计数仍完整验证；该路径不声明本次重新进行网络提取。

作品分组断言完整包含分组名称和数量；进度重启后重新打开实际折叠区，继续比较原摘要。资料来源位置从实际 revision DTO 获取，内容 digest、正文、DOM 和原始附件保留。私有原生附件依现有 MediaType、旧 `whisper-original.json` 标记和跨 revision 相同 ObjectRef 仅显示本机保留，不生成 HTTP 链接；公开附件保留原入口。

Relay 人工迁移按已确认规则使用 `observe`，事件接入核对使用 `reconcile`。写操作在传输失败、accepted 与错误并存、HTTP 5xx（明确未实现的 501 除外）时保留未知交付锁；关闭、重开和重新加载继续禁止重复写入。`sessionStorage` 仅保存 operation UUID。Today 区分 archive 与 native notifications，完整性、unknown、null tokens 和原始窗口失败分别保存。

## 精确源码消费

普通消费 Workbench `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`、Relay 前端 `65ba685ea23e796c1810ff5ec13280dd0dda8f93`、B 实际输入 `e05c013a08ccf372bb884dae5eeb3345f92173af`、私有附件 `cc46b9855775adf70bce74242d08ef84ff4f573b`、真实 URL helper `072628536d325eedbb6b53ffe9ab8fc31c38beeb`、R 人工迁移 `d14b08dc50014baf9ee636543564b040112cf3b5`、H Reader 装配 `b247dfb12b938739f09acc73e6b7582376951a0c` 和真实 listener `1aadeebaba2330bb8b786989f577415da5208fd5`。

H / B 端口方案由 Go 一次持有真实 port 0 listener，helper 使用自产实际 URL 后再核 health，Origin 使用实际地址，Vite 使用严格端口。共享 Go 冲突由模块明确授权的原 Owner 使用文件编辑工具处理并 stage，F 完成普通 merge；F 没有编辑共享 Go、根配置、依赖、OpenAPI 或迁移。

## 验证命令与结果

| 命令 / 实际验证 | 结果 | 本机记录 |
| --- | --- | --- |
| `npm run typecheck`、`npm run lint`、`npm run build` | 退出 0；保留现有约 515 kB 产物提示 | `.verification/typecheck-source-final.log`、`lint-source-final.log`、`web-build-source-final.log` |
| `node scripts/generate-client.mjs --check`、`git diff --check` | 退出 0 | 生成 schema 与共享源相等，未手工编辑 |
| `npx vitest run src/api/client.test.ts src/pages/attention/pluginSnapshot.test.ts src/pages/attention/sourceSelection.test.ts` | 3 文件 21 项 PASS；包含真实生产 HTTP / SQLite | `.verification/vitest-affected-http-final.log` |
| 合并产品之前完整 `npm run test` | 26 文件 147 项 PASS，限定当时 SOURCE | `.verification/vitest-combined-all-first.log` |
| 完整默认 `npx playwright test --workers=1` | 280 项：203 PASS、41 FAIL、32 既有门禁 SKIP、4 NOT_RUN，退出 1 | `.verification/playwright-combined-default-first.log`，逐项清单 `combined-default-first-failures.md` |
| 本域五个受影响用例，两种窗口 | 首复验 9 PASS / 1 Windows bind FAIL | `.verification/playwright-affected-ten-first.log` |
| `npx playwright test e2e/s1.spec.ts --grep 'browser imports a summarize' --project=chromium-1920 --workers=1` | 实际 port 0 helper 修复后单例 PASS；旧 49680 bind 拒绝保留 | `.verification/playwright-s1-port0-single.log` |
| 实际两版论文取消、人工重读和分类 / 固定引用 | 两种窗口 6 PASS | `.verification/playwright-public-versions-current.log` |
| `npx playwright test e2e/progress-existing.spec.ts --workers=1` | 两种窗口 2 PASS，全部 GET，新的模型调用 0 | `.verification/playwright-progress-existing-contract.log` |
| `npx playwright test e2e/video-existing.spec.ts --workers=1` | 两种窗口 2 PASS，两版正文 / 位置 / 私有入口 / 公开原件内容均核实 | `.verification/playwright-video-existing-bytes.log` |
| 先前真实 Works 无模型权限 / revision / 撤销 / 嵌套目录 / 未知锁 | 两种窗口 12 PASS | `.verification/playwright-relay-owned-second.log` |
| 最新公开 cognition 四视图及十个独立 SQL 值 | 原完整复验 13 PASS / 1 ECONNRESET FAIL；无修改的独立 SQL 两窗口 2 PASS | `.verification/playwright-cognition-latest-final.log`、`playwright-cognition-sql-current.log` |

所有通过项限定各自 SOURCE、实际来源和运行范围。完整默认 41 项失败未由子集通过覆盖，下一次完整运行等待 R / H 场景与相应合同具备条件。初始 fixture / 类型 / 显示契约失败、HTTP 400、外部来源 503、bootstrap 连接错误、系统 bind 拒绝和未执行项均保留。

## 既有真实模型进度与视频

B 原实际 README 结果为 `agent_inferred`、模型 `gpt-6.1-sol`、原生 ID `01a13a73-1de2-7761-aaa3-5cbd8fb433cc`、revision 3、一个实际依据。用户决定及 OpenAPI 允许 summary 缺失；B 独立原 SQLite / receipt 确认此记录没有生成摘要，HTTP 没有丢失字段。页面完整比较实际状态、依据 / 版本、模型、原生 ID、观察时间及缺失摘要，人工摘要重启的正向覆盖仍保留。首次不存在 GET 路径的 404 和两次错误 summary 必需断言保留。没有取得 Agent 摘要正向样本，原始来源完整性和有效工具集合仍 UNKNOWN。

既有实际公开视频 `BV1PReT6EEqR`、material `534HWFW7E5GEF75ZY4F7ETPLQ2`：revision 1 原文长度 11746、位置 0、公开附件 1、私有附件 0；revision 2 原文长度 13979、位置 952、公开附件 3、私有附件 1。F 对各版实际 HTTP、完整 DOM、digest 和公开附件与既有不可变对象内容逐项比较，私有原件从未通过 F 请求。每版附件都相同及原件必须非空的首次错误断言保留，实际空文件仍完整保存。新的提取、ASR 和模型调用均为 0。

## 原生链路与当前限制

原固定公开任务 `0c6522bc-b232-4040-8407-745885644791` 仅提交一次。输入被接受后实际 HTTP 503 / provider_unavailable 保留；H 独立观察实际 `gpt-6.1-sol`、assistant 1、task_complete 1、8827 tokens、工具调用 0。原有效工具集合 / owned PID 未完整核实，不能标记工具关闭验收 PASS。

原注册 `2026-10-14T13:23:34.0366105Z` 的三分钟窗口 FAIL。两个窗口后续首次 DOM 为 `13:33:25.152Z` / `13:33:25.153Z`，距离原注册约 591 秒，距离实际 dispatch 约 14.516 / 14.517 秒，单独记录。旧 18 个已提交事件仍逐一匹配 H GetEvent，先前只读两窗口 2 PASS；后续原档实测 nlink 2，当前守卫正确拒绝，coverage errors 1 / snapshot_complete false，本次默认只读两窗口 FAIL。F 没有读取、复制或创建原档链接。

人工 done 实际操作 `32c7f5a0-0cb0-4de2-91aa-3e7b9bc0ad07` 于 `15:00:24.5863561Z` 迁移，HTTP 200 / accepted。在模型内容目标核查前勾选，根明确该目标完成验收 FAIL；实际历史保留。随后唯一 native read `5f99133e-390f-4987-b8f2-2889716921ff` 实际 HTTP 409 / delivery_unknown，未重试。

根随后单独批准符合真实退出与归属条件后的清理释放。H 独立 SQL 确认 done、原退出、Prepared、同项目 / 工作区 / Agent 归属及 pending 0 后，F 单次 GUI release `ccf54824-6194-4161-81c0-0d6985ee5171` 于 `15:19:03.3026899Z` 实际 HTTP 200 / accepted。独立 GET 与 H SQL 核 `released`、ActiveTaskID 空、Prepared true、pending 0，原 read 未知和 condition 未知保留。初始脚本错误 Prepared false 断言保留，未重新执行 mutation。清理不填模型目标或来源验收 PASS。

自有 Vite 17343 / PID 3460、只读进度 17346 / PID 36948、只读视频 17347 / PID 60264 均按实际 listener、父进程配置、Vite 命令和退出核实后准确关闭；没有终止 H / B 的 API 或原生进程。新 17345 观察 / 确认 / 模型操作均为 0。

| 剩余范围 | 当前状态 |
| --- | --- |
| 原三分钟窗口、目标完成及当前原档身份 | FAIL，历史不改 |
| 新 v1.4 投递阶段、原生归属、输入 ready、恢复身份与人工上下文交接 | 已收到成对人工交接决定；准确 R / H SOURCE、OpenAPI 和 working 规则尚待交付，不手写或冻结接口 |
| EOF / structured、互斥与 generic 409、未知 send、第二 CLI、publication error | 原完整默认 22 项缺获准真实场景 FAIL，尚未通过 |
| 新真实 CLI / Today / L1 完整验收 | 用户当前允许继续准备；在语义及根规则满足前未开启新窗口 |
| 完整默认合并产品最终验收 | 尚未通过，203 / 41 / 32 / 4 保留 |
| Agent 检索、Mission 创建等未确定业务范围 | 保持原边界，不扩大权限或实现未确认业务 |
| 生成 schema 中共享源英文说明 | 未自行改变共享源或生成内容；本域普通注释已中文化 |

没有 push main、force push、Git 回滚、私有原件读取、正式实验执行或人工评分补填。后续字段、真实模型 / CLI 窗口和最终组合验收由当前任务及协调者的准确交付继续。
