# Astrocyte Workbench v0.1 · 修订执行单

> 文档用途：需求与历史记录。保留本文件独有业务条款、来源、实际检查和未验限制；原日期的人员分工、资源额度与操作安排只适用于当时阶段。后续执行读取 [AGENTS](../AGENTS.md)、[RESOURCES](../docs/RESOURCES.md) 和 [当前任务目录](../docs/TASKS.md)，独立 Reviewer 按准确来源验收；归档与入口调整不改变功能完成状态。

事实源：用户2026-10-13只读工作台任务书与分支/结构修订。修订取代独立workbench/frontend工程及独立Go服务方案，旧workbench/v0.1和两轨WIP保留为历史，不合入本目标。

目标分支workbench/v0.1-workbench从attention-engine/v1-20261011建立，实际起点6b2b91086d87e2173122c3c9f53a36332acffff2；该点相对CI绿基线100f00cb6945c96a7f3bd3f100f7ff65efc61197仅四份交接/验收文档变动，业务源码完全一致。既有React/Vite web、server、鉴权、构建与Playwright复用；Nav新增认知一级页，四视图子路由/cognition/today、/cognition/focus、/cognition/agents、/cognition/evidence。

## 一页计划与独占写域

root负责本单、HTTP契约、状态、独立验收和最终Git。复用原两个Codex owner、原独占worktree及terminal；旧WIP先普通commit/push，旧Dispatch以规格替换结算，再同worktree创建基于新起点的开发分支，不rebase/reset/删贡献。两轨完成后一个集成owner普通no-ff合SOURCE/REPORT，只修少量装配胶水，领域失败退原owner。

| 轨 | 精确write_paths | 交付 |
|---|---|---|
| backend | 新internal/workbench/app/**；新internal/adapters/sqlite/workbench*.go；新internal/adapters/httpapi/workbench*.go；internal/adapters/httpapi/server.go；cmd/server/main.go及新cmd/server/workbench*.go；scripts/workbench/**；contracts/openapi.yaml及新contracts/examples/workbench*.json；docs/workbench/backend.md | 同server只读SQL端点，六字段投影及隐私/鉴权错误测试，在线快照/本机部署/00:30日程 |
| frontend | web/src/App.tsx；web/src/components/Nav.tsx；web/src/api/client.ts、schema.d.ts及新cognition*.ts；新web/src/pages/cognition/**；新web/src/styles/cognition.css；新web/e2e/cognition*.spec.ts；新web/src/**/cognition*.test.ts或tsx；docs/workbench/frontend.md | 复用token/组件/鉴权的四子路由，图表/抽屉/筛选/分页，新spec进原Playwright |

所有现有引擎/events/Sleep/权重/公式/schema migrations零改；不修改web既有spec/Playwright配置或降低断言/timeout。root package/go.mod/go.sum/dependencies.lock不改，当前依赖够用；如遇实际阻塞Handoff。HTTP适配器不引入SQL，SQL只在sqlite适配器；Reader契约类型位于workbench/app。认知HTTP路径/api/v1/cognition/*，走现有loopback/session/origin鉴权与统一错误封装；无第二套认证或前端工程。

## 真实数据与只读红线

唯一快照源：C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w1/formal-20261013-first/state.sqlite。仅在线备份程序可只读此路径；server只挂LOCALAPPDATA/astrocyte-workbench/v0.1内副本，正式源不可被当作server的ASTROCYTE_DATA_DIR或工作台副本。原应用自身写入库使用独立本机app-state目录，不能向正式源或副本运行迁移。工作台Reader mode=ro/query_only，副本原子更换后重开读侧；不对正式DB/WAL/SHM写入。

V1R历史沿已验收外部v1r-w1/replay-002/topic-timeseries-safe.json；docs/demo-data不存在，源偏差记录而不重造演示数据。本机保存快照、配置、PID、日志、bin、已确证状态点，不进Git/MCP；daily00:30在线备份，手动刷新及一条命令启动/常驻沿scripts/workbench。现有程序写入自己的独立app-state不等于工作台写正式库，文件监控审计需按正式DB/WAL/SHM和进程过滤。

输出事件只event_id/kind/topic/tokens/scope/ts。数据库source仅用于过滤及source聚合组标签；其他显示只聚合数。events.data含raw，禁止整个data/observation/raw_bytes序列化；正常/错误/鉴权路径都用固定错误，raw_unique审计。本机源3105事件、31Sleep截至10-12，三源claude_code93/codex846/manual2166。manual存储源含git，不虚构人类/agent比例。今日空活动如实显示。

Focus拼接历史safe A(t)与每次成功快照取得的确证节点标量点，使用状态updated_at/Sleep时点，不拿执行时刻复制当前A、不把贡献当A、不插值、不新增动力学复算。新快照的派生数值历史仅本机保存；按topic+状态ts去重。日环比缺确证前日值用null。

## 验收与收敛

原pnpm check/build通过；新增认知Playwright spec在两viewport运行，四视图真实数据渲染/交互，抽10个UI数值与独立SQL误差0；GET正常/错误raw_unique零泄漏及六字段投影；实际文件系统监控正式库零写。root已下载Microsoft签名Valid的官方portable Procmon64，本机token非管理员，管理员监控当前NOT_RUN；不得用哈希/静态扫描或普通watcher替代完整监控。两日00:30自动成功待实际跨日，两次手动不算。首次RED/NOT_RUN/unknown保留。

开发周期D1快照/Today、D2Focus/Agents、D3Evidence/部署；可在一次会话推进全部实现，时间验收仍按真实日期。validation分支、V4协议/日程/参数/正式库零修改；v0.2仍冻结。解冻后validation与本分支统一向主线收敛，一次合并；当前不向main合并。

2026-10-13T16:22+08:00：旧backend task634f/ctxe5a和frontend task1ec/ctx446已按scope_replaced_by_user失败结算，WIP SOURCE0fa17aa6/REPORT3b123fcd与SOURCEf9477f46/REPORT7335ce54均已push并独立核对clean；不算实现失败或验收通过，不进入新分支。原owner/工作树/终端复用为backend branch workbench/v0.1-backend task8922bfc4a097/ctx607328b4cef6，frontend branch workbench/v0.1-frontend task0e13b2d2c00e/ctx8e65c15392e9。换branch后首次preflight agent_unconfigured拒绝未建立新Task/Dispatch；再次派发输入已接受但turn_start_unobserved原回执保留。实读屏幕确认文本停在composer，各仅一次bare Enter提交，backend已观察Working；不重发任务或启重复owner。实际生成器目标schema.d.ts已校正写域，旧schema.ts计划名不创建。


## Backlog（研究情报登记，2026-10-14）

本文件既有正文从协调分支 `workbench/v0.1-workbench` 的 `438f50ec318888badcab6febf23c0ca71851afe2` 原样带入本次文档分支；该正文与集成提交 `c5da66724d22455fbbf376c2516f91c4bef63d5c` 中的版本逐字节一致。上文定稿条款保留；本节仅登记未来工作，不启动实现或提升 v0.1 验收状态。后续合流只应用本追加节。

### Agents 视图 v0.2 增强：协作图

- 目标：事件流按 scope 共现恢复 Agent 协作结构，呈现谁与谁在同 scope 工作、重复劳动候选和瓶颈节点候选，供定位协作问题。
- 依据：[Xinwen Liu 等，Beihang University，2026-10-07，arXiv:2610.10126v1](https://arxiv.org/abs/2610.10126v1)，*Know the Shape, Find the Fault: Topology-Conditioned Diagnosis of Multi-Agent LLM Failures*（MAScope）。拓扑是诊断变量，不仅是性能参数。
- 范围：纯消费层扩展、零引擎改动，在 workbench v0.2 窗口实施；复用现有快照、只读 API 与六字段事件投影，不增加采集或放宽隐私边界。
- 证据语义：scope 共现只构成协作线索，不直接等同消息通信、重复劳动或瓶颈事实。图上关联须可回溯到 event_id；重复劳动和瓶颈结论须有额外任务 / 时间证据，缺证据时显示为候选。
- 身份限制：当前 Agents 按 source 分组，事件六字段中没有独立 Worker 身份；不得把 source 类别当成单个 Agent，尤其不能把 manual 当成人类身份。开工前核对已有获准消费数据是否能关联 Worker；不足时保留“身份未知”，真实逐 Worker 协作图验收保持未通过，不新增引擎字段凑图。
- 窗口入口：先核对可用数据和判定口径，再在 v0.2 任务书裁决具体消费写域与验收；本登记不占用 v0.1 backend / frontend / integration 的开发 write_paths。

关联：[能力板任务书草案](v0.3-capability-board.md)、[研究叙事素材](../docs/narrative/research-alignment.md)。本节不使用论文数值作为项目指标。
