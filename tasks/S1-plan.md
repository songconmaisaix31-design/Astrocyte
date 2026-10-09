# S1 Attention 一页执行计划

2026-10-09。基线 `d6c1bf2`（S0 + 已交付新布局），主控分支 `s1/attention-materials-20261009`。本轮用户明确授权 S1 完整功能、多 Agent 并行开发、主控验收和最终 push。

目标：真实 arXiv / summarize 结果导入、来源与内容版本、内容/主题/项目沉淀记录、关联与待查问题、依据和下一步齐备的候选、人工反馈、区分人类关注和机器使用、持久化作业及失败恢复。保留 Go/TS/SQLite 与新布局；只做 Attention，不扩展 Workspace/Swarm 执行。无 Python 默认链路；不新增调度框架或完成证明系统。

| 轨道 | 独占 write_paths | 依赖与验收 | 停止点 |
|---|---|---|---|
| W0 契约/传输/集成 | `contracts/`、`internal/attention/app/contracts.go`、`internal/adapters/httpapi/`、`internal/foundation/`、`cmd/`、`migrations/`、`web/src/api/`、根依赖与锁、`scripts/`、`README.md`、`tasks/S1-contract.md`、`tasks/S1-W0.md` | 先发布 DTO/端口/HTTP 契约；迁移与依赖唯一所有者。最后负责普通合并和少量组装胶水；HTTP/契约/构建检查 | 领域缺陷交 W1，存储/导入缺陷交 W2，页面缺陷交 W3 |
| W1 领域/应用 | `internal/attention/domain/`、`internal/attention/app/`（排除 `contracts.go`）、`tasks/S1-W1.md` | W0 契约；纯领域状态、三层沉淀及复用、候选版本/反馈、四维 unknown、人机关注、持久化队列用例；有意义的规则测试 | 公共契约/存储/入口更改交 W0/W2；不实现 S2 批准或执行 |
| W2 存储/导入/整理适配 | `internal/adapters/sqlite/`、`internal/adapters/importers/`、`internal/adapters/objects/`、`internal/adapters/distillers/`、`tasks/S1-W2.md` | W0 端口、W1 规则；SQLite 原子去重/CAS/作业恢复、对象发布、真实 arXiv 和 summarize 导出；用户已选择本机 Codex CLI 整理与原生读取隔离；重启与错误测试 | 迁移仅由 W0 落地；不新建视频下载/转录器，不调用 Python，不修改全局 CLI 配置 |
| W3 Attention 前端 | `web/src/`（排除 `api/`、`pages/workspace/`、`pages/swarm/`）、`web/e2e/`（排除 `s1.spec.ts`）、`tasks/S1-W3.md` | W0 生成客户端；真实导入/作业恢复/继续沉淀/候选反馈和详情来源链；两尺寸、键盘、异步状态；保留现有布局 | 不同时改三页；公共 API/依赖交 W0 |
| W4 验收 | `web/e2e/s1.spec.ts`、`tests/s1/`、`docs/acceptance/S1.md`、`tasks/S1-W4.md` | 发布契约后开始；可重复 AT01–AT04，真实网络材料另验；API/持久化/浏览器相互核对 | 未执行、失败、缺真实材料明确记录，不用 mock 代替真实导入 |

组织：每轨固定 1 Orca Worker + 1 worktree + 1 branch，独占写域并持续返修；完成 commit + push。W0 先发布契约，之后 W1/W2/W3/W4 并行；W0 在独立集成 worktree 普通合并，领域问题退原 Worker。主控只写本计划、决策、STATUS/HANDOFF，最终独立 check/build/E2E/真实验收后提交并 push。实际客户端和模型按启动与 Worker 回执记录，不猜测。

最终验收：AT01 真实论文/视频多轮沉淀有版本且没有 Mission；AT02 同源重复和摘要重试复用或产生新版本、没有重复工作；AT03 Agent 读取不增加人类关注；AT04 later 不丢资料且不是拒绝样本。`pnpm check`、`pnpm build`、`pnpm test:e2e` 必须通过。

用户新增明确要求：术语是“域”，算法排序；人主动分类、@文件；只有进入项目顶层空间后 Agent 才可主动认领读取。热度与权限分别处理，不采用“高层云自动开放”解释。真实材料确定为 arXiv `2504.16054` 与 B 站 `BV1PReT6EEqR`。沿用 W0 契约/迁移、W1 用例、W2 存储、W3 Attention 页面、W4 真实验收的原有写域，权限默认拒绝，不为此大改三页或扩展 Mission 执行。

已确认纳入动作：人在项目空间中 @文件，建立引用并纳入该空间，原資料留在原分类。沿用 Attention 写域实现人类分类、项目空间文件引用与排序；不移动原件、不扩展 Workspace 执行业务。

自动整理先接本机 Codex CLI，沿现有适配器/持久化作业实现，不新增开发框架。仅传明确选定资料正文与版本，不直接交付数据库；先验证当前 Windows 原生读取隔离。

待答：Agent 可读范围、私有资料外发范围、候选 readiness 关联要求。答复前继续真实公开材料获取、分类/@引用、Codex 整理及现有 S1 功能，不冻结未决定的共享接口，不外发私有资料。

主控收口：普通集成源 `e51676bca9508d698bd9968ca1f3e4777a7469a0`，最终报告普通合并 `01e3b0aa4947935fd9b31e06227d234d1277afb6`（只变文档）。`pnpm check` PASS（API14/14、单测40/40）、`pnpm build` PASS、完整 `pnpm test:e2e`144/144 PASS。真实论文人工两轮整理、候选 later 的浏览器/API/SQLite/重载及两尺寸目视核对通过；指定视频缺 summarize 真字幕/总结、授权 Agent 阅读范围和最新 Codex UNKNOWN 仍阻止完整四项收口。全部 settled Workers 输出已归档并释放终端，五条工作树/分支保留；仍可从原轨恢复，主控没有写业务代码。使用 Orca CLI 调度本机 Codex，实际开发模型 gpt-6.1-sol，权限运行探针与模型整理结果另见 W2，不从开发客户端推断原生能力。

用户最后重申视频复用 summarize：使用其既有导出与真实 provenance，无 Python、不另建下载/转录器。完整完成内容、真实限制和本机预览重启见 STATUS 顶部。
