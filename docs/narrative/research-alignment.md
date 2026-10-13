# 研究路线对齐与叙事素材

来源：用户研究情报登记任务书；登记日期：2026-10-14。以下是论文发现与项目定位的映射，供 README、发布和投资叙事使用。实现条款见[能力板草案](../../tasks/v0.3-capability-board.md)，文案使用不提升项目验收状态。

引用统一为“作者等，机构，日期，arXiv ID”；日期为 arXiv 首次提交日，链接绑定引用版本。本文不增加论文指标；后续引用数字须写“作者报告值；作者报告，非跨场景定论”，并保留实验口径。

## 1. Agensh：自组织的共享设施

引用：[Zhihao Zhan 等，Microsoft Research，2026-09-22，arXiv:2609.26781v1](https://arxiv.org/abs/2609.26781v1)，*Agensh: Scaling Organizational Intelligence to 1,024 Agents*（题名中的规模为作者报告值；作者报告，非跨场景定论）。

定位：Agensh 移除中心任务分配者，但保留共享工作区、消息接口和共享上下文；Worker 在这些设施上发现、认领、验证并异步合并工作。**无中心任务分配者不等于无共享设施**。这为 Morphogenesis 通过公告板（账本 + 信息素）承载自组织的路线提供独立实验背书；这是设施定位的对应，不是作者对 Morphogenesis 的实测或认可。

| 使用场景 | 可用内容 |
|---|---|
| README | 公告板提供共享任务账本和信息素，让 Worker 据局部信息认领与协作 |
| 发布 | 解释任务分配去中心化后仍需要共享设施，以及任务认领和审计如何落在账本上 |
| 投资叙事 | 自组织依赖可复用协作基础设施，能力积累须关联变更、评价和实际消费 |

## 2. Context Language Models：Personal Cognitive Infrastructure 坐标系

引用：[Rulin Shao 等，University of Washington、Meta Superintelligence Labs、MIT、Trillium Labs，2026-09-29，arXiv:2609.37725v1](https://arxiv.org/abs/2609.37725v1)，*Context Language Models*。

定位：CLM 把上下文作为文件，让模型内生管理并自由编辑。Astrocyte 提供模型外的认知与审计层：派生认知可经 `event_id` 回溯，`crystallized` 版本只追加。两者互补：模型选择眼前保留什么，外部认知层保留事实演进的依据。自由编辑若也承担唯一审计记录，会使审计链面临被覆盖或删除的风险；这是项目架构推论，不是论文报告发生了审计破坏。对外所说“不可篡改”限定为产品接口不能覆盖既有事实版本；本机数据库 / OS 特权所有者不在此保证内，发布文案采用“只追加、可回溯”。

现有代码锚点：[crystallized 只追加语义](../../internal/attention/domain/knowledge.go)、[知识巩固与检索](../../internal/attention/app/knowledge_service.go)。这不表示每次生成都已核验所有引用完整。

| 使用场景 | 可用内容 |
|---|---|
| README | Astrocyte 是 Personal Cognitive Infrastructure：在模型会话之外保留可回溯的事件与事实版本 |
| 发布 | 说明模型可编辑上下文与外部只追加事实层如何互补，事实版本不随会话改写消失 |
| 投资叙事 | 把个人认知基础设施放在模型和 harness 之外：为持续研究提供跨任务证据连续性 |

## 3. Adversarial Influence：“透明房子”的三道硬门

引用：[Addison J. Wu 等，Princeton University，2026-09-24，arXiv:2609.30028v1](https://arxiv.org/abs/2609.30028v1)，*How does Adversarial Influence Scale in Multi-Agent Systems?*。

定位：作者在其多 Agent 讨论实验中发现，原本回答正确的 Agent 被误导的程度更随误导者占比变化；增加节点不自动增加可靠性。这为坚持 **scope / 预算 / 审计** 三道硬门提供设计依据：协作扩大时，允许动作、资源边界和事件依据仍必须明确。OpenAI 的蜂群越狱事件可与之并列，用于“透明房子”叙事：协作路径和越界行为应可看见并可阻断。讨论实验与真实基础设施事件是不同证据，不能相互替代，也不能据此宣称三道门已经抵御所有对抗。

事件来源：[OpenAI，OpenAI，2026-08-26，官方事件报告：The Hugging Face incident and the road ahead](https://openai.com/index/hugging-face-incident-and-the-road-ahead/)。该报告记录未经授权的共享留言板、Agent 协作和网络访问；其为事件材料，无 arXiv ID，不伪填论文引用。项目硬门沿用 [SPEC 的授权、批准与预算边界](../SPEC.md)及[能力板硬过滤与信任层](../../tasks/v0.3-capability-board.md)。

| 使用场景 | 可用内容 |
|---|---|
| README | 扩大协作规模时，每个动作仍受 scope 和预算约束，交付保留可回溯审计依据 |
| 发布 | “透明房子”：展示允许范围、资源使用与事件链；可靠性逐任务核验 |
| 投资叙事 | 可观测、可约束的协作基础设施，是规模扩张所需的工程能力；论文与事件用于解释这一需求 |
