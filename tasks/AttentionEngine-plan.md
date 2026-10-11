# Attention Engine 认知内核开发计划（v1.1）

事实源：[tabbit-code (2).md](../../../Downloads/tabbit-code%20(2).md)（Astrocyte Attention Engine 开发任务书 v1.1 · 认知内核）。
基线：`dab8fea`，集成分支 `attention-engine/v1-20261011`；s1 分支族只读不改。本轮目标 M0–M6（三条链全部），按互斥写域解耦、无并发上限并行。

## 主控接管（2026-10-11 21:41）

2026-10-11 23:25 Windows CI 限定返修：最终普通合并 `0e49301734549a02b9b1b9078aaef96fd6422742` 的 root `pnpm check` / `pnpm build` 已 PASS，默认浏览器回归运行中。已观察自动 CI `38150022874`（root `dd62ec9`）及 `38150373657`（root `7c199aa`）Windows FAIL / Ubuntu PASS；首次完整失败日志保留 `%TEMP%/astrocyte-ci-38150022874-first-failed.log`。原因是 Windows `RUNNER~1` 短路径经 EvalSymlinks 展开成 runneradmin，合法同目录被判 redirected / outside scope。追加两条互斥限定返修轨，沿用原 Codex --yolo / gpt-6.1-sol：W0 原终端/工作树/分支独占 `internal/adapters/agents/**` 与自己的报告，修授权目录、注册项目及既有元数据路径比较；W7 原终端/工作树/分支独占 `internal/adapters/worksscan/**` 与自己的报告，修作品扫描路径比较。不得互改、不增共享调度或证明设施、不放宽符号链接/遍历/私有目录禁读边界，不读取个人原生记录。使用公开 fixture 和短路径实测，保留未具备实际 8.3 路径时的 NOT_RUN；定向回归、commit + push 后交同一集成 Agent 普通合并，root 完整检查后 push 并核实新的自动 CI。旧远程结论仅属于其精确 SHA，不能冒充最终树的结果。

2026-10-11 23:12 波 C：原八轨及W3/W4/W6限定返修全部结算；W1真实metadata回放因权限范围未答failed，其本地实现可合但M1真实验收未通过，其余限定实现succeeded。W0最终 SOURCE `f4bf0257b3f2e93e2ce67197bfbb0764a68332a3` / REPORT `5c1aa0dc41358db29c3cc12c9a250f19b7fc69b0` 完整check/build PASS。唯一集成轨：Codex--yolo / gpt-6.1-sol，worktree `C:/Users/DW/orca/workspaces/Astrocyte/attention-engine-final-integration`、branch `songconmaisaix31-design/attention-engine-final-integration`、Task `task_9c5f23440239` / Dispatch `ctx_7eb57b08004b` / terminal `term_28cadb7a-4147-4dec-9f7a-87ecc545576d`。基线 root `dd62ec9`；write_paths仅 `tasks/AttentionEngine-integration.md` 及必要路由/导入/配置/类型薄胶水，领域返修退原owner。普通合精确最终报告，定向M6和diff复验；root随后完整check/build/适用E2E及最终push。首次注入turn_start_unobserved，检查后仅bareEnter，已实际观察开始阅读/工作，未重新派发或重复整条prompt。

2026-10-11 22:00 用户更新：后续统一 `codex --yolo`，覆盖此前 OpenCode 客户端决定。已核 Codex CLI 0.162.0 接受该参数；按轨先提交当前阶段、结算并归档旧 Worker，再于原工作树启动 Codex，写域及 Git 历史保留。八轨已实际接续为 Codex / gpt-6.1-sol（实际 runtime projection），有效 Task/Dispatch 见 STATUS 最新表。W5 旧终端由运行时标记 user_takeover，未强制关闭，已结算 idle。

2026-10-11 22:22 正式 P1 基础 `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0` 已 root 普通合入 `2268f4920973128068f46a182dfddfb040a1dbc2`、精确源五包复验 PASS 并 push，波 B 已解锁。W0 P2 消费 W7 已发布稳定 Go 阶段 `0e38f94fa516f5b9f8ec4f48f6f76fbad425e5e1` 作 HTTP/cmd 编译依赖；不作最终集成验收。领域缺口继续原轨返修，最终集成仍在全部轨 settle 后。sqlite-vec CUT，relevance 使用显式可复算退化，不能称向量检索已验。

用户要求 Codex 接替原 OpenCode 主控。原主控终端 `term_462ee860-933e-43d7-94cb-491d01617fcf` 已中断；原 Run `run_19a0b1268e5a` 绑定到 Codex `term_1adf40a6-ef79-4efa-b602-7b55173624e6`（consumer_generation=2）。八轨原 Task/Dispatch/工作树保留。八个 Worker 均卡在读取用户指定任务书的权限弹窗，单次读取已放行；此前 ready/input_accepted 不代表开发已开始。

实际 Worker 分支使用 Orca 自动名称 `songconmaisaix31-design/attention-engine-w0-core`、`-w1-events`、`-w2-engine`、`-w3-sleep`、`-w4-knowledge`、`-w5-experience`、`-w6-mcp`、`-w7-works`。下表 `attention-engine/wN-*` 为原计划名称，消费提交时以实际分支和 owner 发布的精确 SHA 为准，不重建分支。

W2/W5 原 Task 提前以设计交付结束：`8a4db07` / `22fef44` 只算波 A 文档，M2/M5 尚未实现。原终端已复用为 W2 `task_1efa13b6e981` / `ctx_ad864daccaa9`、W5 `task_57f40b160754` / `ctx_3f9743f7e989`，原已结算 Dispatch 不再使用。

W3/W4 原 Task 亦仅交付波 A 核心：`c3e8395` / `ab7d381`。原终端已复用为 W3 `task_e5487ee2214c` / `ctx_91c890736e4a`、W4 `task_da85be3296c1` / `ctx_8ae58c0f5840`。W3 `3f6cc6f` 已修正多主题误自动合并；W2 `acb4561` 已通过纯公式复算；上述均不等于存储、应用层或完整里程碑交付。当前共同依赖 M0 精确提交。

写域补充：所有 `*_contract*.go` 归 W0，W1 event/signal 与 W2 engine 通配符明确排除契约文件。阶段设计交付用 status/merge_ready，业务验收完成后才 worker_done。Go 编译使用进程环境 `GOFLAGS=-p=1 GOMAXPROCS=2`，不修改宿主分页设置。

参数待答：衰减 λ、信号强度、主题密度/分裂及 confidence 冲突阈值的工程初值已一次集中询问，未答时只实现参数化内核，不冻结默认值。α=0.3、A_floor=0.05、θ=0.15、合并关联>0.8、90 天冷归档沿任务书。momentum 不额外改写 A(t) 公式。

## 北极星（一切裁决）
核心算法是持续演化的个人认知模型：观察行为→注意力；注意力→组织知识；经验→提升 Agent 能力。每行代码须能回答"这行代码让 'Astrocyte 如何知道什么值得关注' 更可回答了吗"。

## 三条链与里程碑（任务书 §8）
| 里程碑 | 内容 | 验收标准 |
|---|---|---|
| M0 | §4 全部 schema 落码 + SQLite 迁移 + 引擎接口定义 | schema 测试通过；现有 watcher 事件可无损映射到新 Event；sqlite-vec spike 结论 |
| M1 | 事件归一化：Event→Signal，权重表硬编码 | 真实 Claude Code/Codex 会话回放；权重标注抽检 |
| M2 | Attention Dynamics：§5.1–5.5 全量 | 30 天伪造流 A(t) 与手工复算逐点一致 |
| M3 | Sleep Consolidation 流水线 + topic 生命周期 | 双指标可测；冲突记录；无 LLM 决策入库 |
| M4 | KnowledgeEntry 两级存储 + 检索重排（§5.6） | 排序可复算；fluid 可重写、crystallized 只追加 |
| M5 | Experience Memory + MCP 五操作 | 失败轨迹捕获；explain_focus 返回完整证据链 |
| M6 | 端到端闭环 | 一周混合流 → 焦点迁移 + 推荐 + 可解释报告 |

## 红线（任务书 §9，每条都是资源裁决）
不自研向量检索（sqlite-vec，1 天接不完即砍）；不训练/不微调 embedding；LLM 只做解释器（入库靠规则+用户关口，capture→structure→approve）；不做 Agent 编排/任务调度；不做防御性工程；测试只覆盖 M0–M6 验收本身。文风纪律（§10）：注释/commit/报告只保留"推进实现或传递信息"的句子。

## 决定（本轮已与用户确认，记入 docs/QUESTIONS.md）
1. 范围三条链全部（M0–M6）；在现有 `internal/attention` 内演进（任务书 `attention-engine/` 模块树映射到 `internal/attention/{domain,app}` + `internal/adapters/{events,sqlite,mcp,worksscan}` + `migrations/014+`）。
2. 信号来源按任务书 §4 清单：claude_code|codex|manual|mcp + 现有 s1 产出经适配器归一化；不读浏览器历史/不整盘监视。
3. 人工关口（Knowledge confirm / Experience approve / 冲突裁决）经新增 HTTP 端点；本轮 attention engine 本体不新增前端。
4. 记忆可见范围：Experience/KnowledgeEntry 默认**项目隔离** + 人工显式跨项目分享（带审计）；Attention Graph 本体个人级。
5. 新增 W7 本地作品聚合轨（"AI 作品寻回册"级别，设计参考无源码）：平台聚合与统计 / 作品卡片与文件核实 / 交接手册 / 态度与视图 / 扫描导入与同步，并作为链1真实事件源。

## 轨道与互斥写域（1 Agent + 1 Worktree + 1 Branch，已切到 Codex --yolo）
| 轨 | 分支 | 独占 write_paths | 里程碑 |
|---|---|---|---|
| W0 契约·迁移·HTTP·存储基础·集成 | attention-engine/w0-core | `contracts/`、`web/src/api/`、`migrations/014+`、M0 契约文件（`internal/attention/domain/engine_contract.go`、`internal/attention/app/engine_contract.go`、`*_contract*.go`：Event/Signal+权重表/AttentionNode/Experience/KnowledgeEntry/Evidence/引擎接口）、`internal/adapters/httpapi/`、`internal/adapters/sqlite/` 共享基础设施（连接/迁移执行/`engine_common*.go`）、`cmd/`、`go.mod`/`go.sum`、`scripts/`、`tests/attention-engine/` 共享 harness | M0 + 存储基础 + HTTP 关口 + 最终集成 + M6 |
| W1 事件流与归一化 | attention-engine/w1-events | `internal/attention/domain/{event,signal}*.go`、`internal/attention/app/{event,signal}*.go`、`internal/adapters/events/**`（新包，消费现有 s1 产出）、`internal/adapters/sqlite/event_store*.go`、`tests/attention-engine/M1*`、`tasks/AttentionEngine-W1.md` | M1 |
| W2 动力学引擎 | attention-engine/w2-engine | `internal/attention/domain/{dynamics,lifecycle,topic,antimatthew,conflict}*.go`、`internal/attention/app/engine_*.go`（状态更新 + 注意力/解释读侧）、`internal/adapters/sqlite/engine_store*.go`、`tests/attention-engine/M2*`、`tasks/AttentionEngine-W2.md` | M2 |
| W3 Sleep 巩固 | attention-engine/w3-sleep | `internal/attention/domain/sleep*.go`、`internal/attention/app/sleep_*.go`、`internal/adapters/sqlite/sleep_store*.go`、`tests/attention-engine/M3*`、`tasks/AttentionEngine-W3.md` | M3 |
| W4 Knowledge 两级存储+检索 | attention-engine/w4-knowledge | `internal/attention/domain/{knowledge,retrieval}*.go`、`internal/attention/app/knowledge_*.go`、`internal/adapters/sqlite/{knowledge_store,vec}*.go`、`tests/attention-engine/M4*`、`tasks/AttentionEngine-W4.md` | M4 |
| W5 Experience 记忆 | attention-engine/w5-experience | `internal/attention/domain/experience*.go`、`internal/attention/app/experience_*.go`、`internal/adapters/sqlite/experience_store*.go`、`tests/attention-engine/M5-experience*`、`tasks/AttentionEngine-W5.md` | M5（经验侧） |
| W6 MCP 认知接口 | attention-engine/w6-mcp | `internal/adapters/mcp/**`（新包，stdio MCP 五操作，项目 scope 过滤）、`internal/attention/app/mcp_*.go`（薄胶水）、`tests/attention-engine/M5-mcp*`、`tasks/AttentionEngine-W6.md` | M5（接口侧） |
| W7 本地作品聚合 | attention-engine/w7-works | `internal/workspace/{domain,app}/` 新文件（works 聚合，不改既有）、`internal/adapters/worksscan/**`（新包，有界扫描/导入）、`internal/adapters/sqlite/works_store*.go`、`web/src/pages/workspace/` 新页面 + `WorkspacePage.tsx` 最小接线（该文件 W7 单一 owner）、`web/e2e/` 新 spec、`tests/attention-engine/W7*`、`tasks/AttentionEngine-W7.md` | 作品聚合五能力 + 链1事件源接入 |

root 独占：`tasks/AttentionEngine-plan.md`、`STATUS.md`、`HANDOFF.md`、`docs/QUESTIONS.md`、`docs/acceptance/AttentionEngine*.md`、`README.md`、最终 Git。

## 依赖与波次
- 契约/迁移/入口/依赖锁仅 W0 可改；跨轨需求走 Handoff。
- 波 A（立即并行）：W0 做 M0（契约+迁移+引擎接口+spike）；W1 事件源勘察与回放数据准备；W6 MCP 骨架（§7 自足）；W7 本机平台/作品来源勘察与 JSON 导入设计；W2/W3/W4/W5 做只读勘察+里程碑设计+验收脚手架（不写依赖 M0 类型的代码）。
- 波 B（M0 合入集成分支后）：W1–W6 全速编码（fetch 集成分支取得契约类型）；里程碑逐个由 root 独立验收。
- 波 C：全部轨 settle 后派 1 个集成 Agent 普通合并八轨（只补路由/导入/配置/类型胶水），M6 演示 + 冷重启；root 全量 check/build 后最终 push。

## 验收（root 独立复验，不冒充）
M0 schema 测试+迁移+spike；M1 真实会话回放与权重抽检；M2 第二实现逐点复算；M3 断 LLM 跑 pipeline+双指标；M4 排序复算+SQLite 只读核查；M5 MCP 五操作实测+证据链；M6 一周流演示报告；W7 真实本机作品+三尺寸/键盘+冷重启。每阶段在 STATUS.md 记录 SHA+时间戳+测试数。

## 偏差记录（任务书 §8 授权，开工即记）
1. 任务书 `docs/STATUS.md` → 仓库根 `STATUS.md`。
2. 任务书 `attention-engine/` 独立模块树 → 映射进现有 `internal/attention` + adapters（用户"现有 attention 内演进"）。
3. "从最新主线拉分支" → main 仅 Initial commit，最新主线=s1 主控 HEAD `dab8fea`，自它拉 `attention-engine/v1-20261011`。
4. sqlite-vec：modernc.org/sqlite 纯 Go 无 CGO，spike ≤1 天；接不上按红线砍、relevance 退化并记录。
5. 任务书 §3 "现有 s1 面板不改" → W7 按"最小接线、不破坏既有行为"执行。
6. §11 七天真实使用对比属 M6 之后用户侧验证，另行排期。
