# S1 同步与本地 Agent 前端 W3

## 用户决定后续作（2026-10-10）

当前 Dispatch `ctx_bf18050e3ee7`，Task `task_bb04bd91e57a`，固定分支/写域不变。普通合入主控 `ca7ba6e80ab256a33465c5bdd959400375569a0e`，未 reset/rebase/force/clean；QUESTIONS 顶部覆盖历史 pending。实际客户端仍为 Orca Codex，本轮未另选模型。

先实现现有作业恢复显示：次数与截止期均来自持久作业，恢复入口同时检查两项；未知外部结果（delivery_unknown 标志或错误码）禁重发、不显示已完成。有限次数/时限耗尽给“暂未完成”及下一步，原始错误保留在可展开处理详情。提交错误区保留中文服务动作和原输入，响应丢失要求先核对原记录，不自动重发。候选表单说明固定版本的主题关联或明确待查问题门槛；酝酿中候选提供补充关联/问题的路径，最终状态由 W1 决定。

`pnpm --dir web typecheck`、`pnpm --dir web lint`、`pnpm --dir web test`（49/49）、`pnpm --dir web build`、`git diff --check` PASS；安全性单测覆盖 unknown 优先级、次数/截止边界、网络响应未知和实际中文动作提示。浏览器/live 未执行，未占测试槽。W0 generated API 尚待发布，已给 W0/W1/W2 交接 UI 所需字段和行为，不自行新增 DTO 或 HTTP。账号真实非空选择、项目注册/权限/原生操作、AT01–04 与重启验收仍需后续完成；不将此阶段检查替代整任务验收。

### 真实项目首轮失败与恢复

真实项目浏览器 SOURCE `9009259b8121849a660f9738b532b7a8e9b4fb96`，主控单槽 `pnpm --dir web exec playwright test s1-projects.spec.ts --project=chromium-1920 --workers=1 --output=test-results/s1-project-first`：1 FAIL，14.8 秒，exit1。实际创建空间与登记项目成功，登记 HTTP200 返回 `history_roots:null` 后 UI 遍历崩溃；B/C 保存、文件上下文和重启尚未执行。首败截图、error-context 与 trace 保留在该 output 子目录，未调用原生控制/模型/媒体。API/Vite/浏览器命令退出后相关端口无监听。

前端兼容缺失历史映射，显示未提供且禁止历史发现；此兼容不授权任何根。W2 修复实际 DTO 空映射/切片，已分别交接 W0/W2；首次 RED 保留，后续定向复验单独输出，不替代历史失败或全任务验收。

修复 SOURCE `711a3cde30830496d7b0cc832f45068e798db8da` 已 push，同命令输出 `test-results/s1-project-null-fix`：1 PASS，25.9 秒，exit0。真实 UI 创建空间、手动登记临时公开根、独立 B 开/C 关、保存明确 README 路径、预览实际文本、SQLite 设置比对和 API 重启持久化均通过，原生会话计数为 0。相同实际状态 resize1280/1920 截图已目视，无横向溢出；服务退出后15173/18787/55302/55303无监听，无相关 node/server，已明确向主控释放唯一 live 槽。个人根和原生/model/media 仍未测试。

根据主控冷未知操作复核，已登记驱动与已观察支持分别处理：准确项目 CLI 与动作许可下，人类可明确首次验证 unknown 的 start/resume/send/observe，unsupported 仍禁用。未知结果/待定操作禁止 send/resume；只允许控制 owned 匹配会话。人类 stop 独立命令不受上下文读取失败或能力缓存过期阻塞，仍需动作许可且服务最终核对停止；无自动派发或 paid 探测。静态首次检查因新增单测误写 observed_at 而 typecheck FAIL，改为契约 checked_at 后 typecheck/build 和三项权限门槛单测 PASS；lint PASS。原生正向浏览器未执行，不把单测当原生接续证据。

### 已发布客户端上的实际表单

普通合入 W0 typed client `b47a3fc63e5ff07b825ecb672454ed5005743650`，再合 `2e5ef61982bbb55278ea9fe566b5cda72d9df27e`（公开收藏列表、历史根、身份临时凭据）。Attention 现可绑定多个公开创作者/收藏夹，所选 B站 UID 默认填入但不自动绑定；抖音 self 显示公开 URL 请求并禁提交。清单重载为 GET，标题同步为明确确认的 POST100，人工正文选择须实际反馈成功且元数据版本相同；running/unknown 不重发。收藏列表由实际 API 返回供人选择，不自动绑定所有列表。普通单资料导入默认复用，刷新正文需明确勾选。

共同工作区局部接入真实 /local-projects，与 S0 projection 分开：手动绝对根、所选目录发现子项目、明确登记/关联空间；项目 A 默认，B/C 独立，准确 CLI 模型处理许可、动作授权与显式历史根；人类可授予/撤销项目身份并按需生成临时 token（仅页面内存、不日志、不 storage）。只交付所选固定资料与 B 文件，原生 start/resume/send/stop/observe 和真正新会话交接分别呈现。未知能力暂禁用；W0 将发布实际 registered adapter 字段后加入明确验证入口，不能由 installed 推导支持。外部历史只观察，不冒充可控制会话。

本阶段首次 typecheck/build 因新单测 provenance.model=null 与契约 string? 不符 FAIL（TS2322），保留为首败；修正为未提供 model 后两命令 PASS。lint PASS、Vitest51/51 PASS、diff PASS。仍未占浏览器槽或调用模型、媒体、原生控制；实际 HTTP 组装与 W2 验证尚在推进，不把编译表单当作 live 验收。

2026-10-10；基线 `20437708e43201e352d6c6926902e1363fd2ad3e`；分支 `s1-sync-ui-1010`。本机 Orca 原生 Codex，主控通过 worker-read 确认实际模型 `gpt-6.1-sol`。写域为 web/src（排除 api 与 pages/swarm）、web/e2e、本报告与 docs/acceptance/S1-sync.md。

## 阶段一

已读取 AGENTS、HANDOFF、STATUS、S1-sync-local-agent-plan、SPEC §6/§19、QUESTIONS，并用 view_image 检查指定 JPG。保留既有三页 shell/布局；只在 Attention 加入账号待接入区，把 Workspace 项目区改为参考图结构的暖白概览、平台/绑定状态/分组筛选与网格卡。卡显示服务注册目录，详情沿用原抽屉。统计基于已加载项目/会话，平台仅由匹配 project_id 的会话聚合，不推导已安装或正在运行；API 未提供分组、文件数、交接内容、活动时间，保持未提供，最近活动排序禁用。

账号 B站/抖音绑定方式待确定，其余平台明确待接入，无伪连接按钮。没有新增连接、扫描、执行或摘要请求；未接入数据不得等同零更新。W0 的客户端/契约发布后由 W3 消费，不修改 API 或锁。

已执行 `pnpm --dir web typecheck`、`pnpm --dir web lint`、`pnpm --dir web test`（43/43）、`pnpm --dir web build`、`git diff --check`，PASS。新增单测验证项目关联隔离、筛选组合、活动排序未知/无效值置后且不改输入。浏览器由主控授予单槽，准备执行现有 ephemeral helper 上的 targeted spec/navigation/preview-design；尚未执行，不能声明浏览器通过。

## 待定用户旅程

这是讨论草案，不冻结接口：确认绑定来源 → 获取更新列表 → 展示来源/状态与 Agent 反馈 → 人工选择 → 复用现有导入/作业反馈 → 回查资料版本。反馈前是否预取正文仍 pending；账号登录或公开绑定、真实目标账号、本地操作范围与读取根待用户。遇到挑战页/需登录/不可用应显示具体原因，保留先前清单，不把异常转换为空更新。项目导入和扫描入口待指定目录授权；安装/配置/可启动/原生能力分别显示，观察会话不等于控制。

完整账号集成、真实同步、AT01–04、新模型整理、授权 Agent 读取、两尺寸最终真实验收尚未完成。旧 UNKNOWN 未重发；未导入个人库，未读凭据或扫描磁盘。

## 阶段一首轮浏览器（原始失败保留）

SOURCE `04a67c54a6ccdbba55dc08e5ec44c07a739d79e7` 已 push。总控单槽安排 `pnpm --dir web exec playwright test s1-sync.spec.ts navigation.spec.ts preview-design.spec.ts --workers=1`：120 PASS /2 FAIL，2.3分钟，exit1。两失败均为新 empty-state 用例中对原生 `<option>` 使用 `toBeDisabled()`，Playwright 展示 `<option disabled>` 与可访问快照 disabled 却返回 enabled；改核对 `toHaveJSProperty('disabled', true)`，应用代码不改。占位账号文本/零按钮测试按主控复核移除，后续应测试实际绑定/选择行为。首次不是绿，不覆盖。

首轮 artifacts：`web/test-results/s1-sync-local-overview-kee-02fb9-acts-explicit-at-both-sizes-chromium-{1920,1280}/` 的 error-context、test-failed-1.png、trace.zip。两尺寸 `web/test-results/workspace-{1920x1080,1280x720}.png` 与卡片截图已目视检查：保留 shell，宽屏三列/1280两列、筛选可换行、目录可读、无横溢。示例图是布局证据，不是本地项目接入证据。

临时目录 `%LOCALAPPDATA%/Temp/astrocyte-dev-qQGXqS`，个人库未使用；API86296/Vite88652，父72516，创建时间03:04:56。命令结束后所属三进程均不在，15173/18787无监听，Test-Path仍True，目录保留不自行清理；已向主控释放单槽。仅失败两尺寸用例的修复验证待主控重新安排；不重跑已过的导航回归。

## 阶段二：已发布清单客户端

普通合并 W0 已推送 `bb912efd6f28f9f3909d87ea9fc1f4acb04c9098`（含 W2 DTO），没有手改 generated API。新增 useLocalAgents 与 LocalAgentsPanel，分别显示安装/配置/可启动观察和八项原生能力的状态、原因、核实时间。计数仅来自该接口已加载记录，版本缺失保持未提供，fixture 模式不查询真实清单。清单 GET 只展示已保存探测结果；没有控制按钮或命令请求。

阶段二 `pnpm --dir web typecheck`、lint、test（44/44，含 W0 新客户端单测）、build PASS。新增浏览器用例比较真实 GET local-agents 数据与页面逐项原因及八能力、刷新只发 GET；等待 W0 组装/service 源和主控单槽，尚未运行。不能从 DTO 发布或 UI 编译成功推断实际已接入，更不推断启动/接续/停止能力。

W0 真实组装源 `86502c63765c19a086770d8646030791ca6c39da` 已发布并普通合并（本地合并 `6aede6d72ed2fa6aedf0f3a3e1b2ca173e46b288`）；原生版本/help 启动探测总10秒、每次5秒，GET只读缓存。主控复核后将已知原因译为中文，未知原因保留诊断详情，保持未知与不支持区别。新增实际清单 E2E 和原失败修复准备一起在单槽运行，首次 artifacts 不覆盖。

## 阶段二真实浏览器（新启动失败保留）

测试前再普通合并 W0 最新 `343070262c6990f82f8ee26817ae75d29867f5f0`（包含 W1 d0e717d 和 W2 最终47772035），合并 `4c4101e86adafafd56d50f3be701fc2ba63859f4`。实际 SOURCE `e2aa6dcf1dcd29339aacc541ace76991b3263632` 已 push。

主控单槽 `pnpm --dir web exec playwright test s1-sync.spec.ts --workers=1 --output=test-results/s1-sync-phase2`：5 PASS /1 FAIL，19.1秒，exit1。真实 inventory 在1920/1280均 GET200，10客户端身份/8已核实安装；其余2个入口未在PATH找到。所有配置、可启动、八项原生能力均 unknown，已核实配置/启动计数0不代表不支持。页面逐项匹配API版本、原因和八个unknown状态；刷新GET200，无非GET业务请求。标准 Playwright 附件 `local-agents-response` 保留真实JSON；截图 `web/test-results/s1-local-inventory-{1920,1280}.png` 已目视，两列/一列可读，无虚构活动或能力。所有能力在测试中展开便于复核，日常默认折叠。

新失败仅1920首个空状态：Vite已HTTP就绪，API仍在启动版本/help探测，03:14:23 auth/session和foundation代理 ECONNREFUSED，API03:14:26.576才开始监听。真实页面显示错误，不把失败改为0数据；W0拥有scripts，已发送修复启动就绪顺序Handoff，无测试sleep/retry掩盖。1920失败迹在 `web/test-results/s1-sync-phase2/s1-sync-local-overview-kee-02fb9-acts-explicit-at-both-sizes-chromium-1920/`。原option断言修复在1280正常通过；1920修复验证被新的启动失败阻断。

独立临时 `%LOCALAPPDATA%/Temp/astrocyte-dev-Cj4TbI`，个人库未使用。命令退出后15173/18787无监听、Vite86036不在，无node/server命令行匹配本工作树或该临时目录，目录仍True保留；已释放单槽。待W0脚本修复后仅复验失败1920用例，再交主控最终独立完整检查。账号绑定/同步持久化/真实项目导入或操作/AT01–04仍未运行，不声明本轮任务完成。

## 启动修复后唯一失败项复验与最终交接

普通合并 W0 `23028df474d0a9d5f53bce4e2386c361026c2d98`，已 push 的应用验收源 `d87604e1e3d732dc12f2c8349958d797d8324607`。W0只改既有dev启动顺序：API健康后启动Vite、启动失败不放行前端；没有业务重试/测试sleep。主控单槽 `pnpm --dir web exec playwright test s1-sync.spec.ts --project=chromium-1920 --grep 'local overview keeps real empty' --workers=1 --output=test-results/s1-sync-readiness-fix`：1 PASS，9.7秒、exit0。API03:19:02.302监听，health200先于Vite；首屏auth/session、foundation、projects、local-agents均200，空状态与disabled DOM属性正常。只是失败用例复验，不宣称整套重跑通过；首次120/2与阶段二5/1历史和trace保留。

临时 `%LOCALAPPDATA%/Temp/astrocyte-dev-hJ3JOt`，Vite84548/父84984；退出后两进程不存在，无本工作树/临时目录所属server/node，15173/18787无监听。目录仍True保留，未绕过Windows helper退出后的清理限制。修复截图 `web/test-results/s1-local-empty-1920.png` 可回查；首次阶段二失败目录与readiness-fix独立，不覆盖。

最终为**部分交付**：公共概览/筛选/项目卡与实际本机安装清单已实现并验证；默认真实API、UNKNOWN、未支持操作与无样本回退保持。原任务不能记成功：账号绑定方式、清单反馈/正文顺序、授权本地操作/读取根待用户，真实账号/清单/人工勾选入库及重启、真实项目活动/交接、AT01–04尚未完成；旧UNKNOWN未重发。主控已指示本Dispatch按outcome failed/partial收口，同分支/写域保留供后续确认后续作。

交W0普通合并最终报告后，由主控独立最终check/build/E2E。W3不重复已过广泛回归、不调用模型/媒体/原生控制、不使用个人库。当前应用源只比阶段二增加启动脚本修复，最终报告提交只改本报告与acceptance；分支已push，无待提交文件。
