/**
 * Fixture mode detection and data.
 *
 * Fixture mode is ONLY activated by ?fixture=1 in the URL.
 * It provides realistic sample data to exercise long Chinese text,
 * unknown fields, edge cases, and various statuses.
 *
 * API failures NEVER silently fall back to fixtures.
 *
 * Types match the V1 generated schema from web/src/api/schema.d.ts.
 */
import type { components } from '../api/schema';

type Material       = components['schemas']['MaterialV1'];
type Opportunity    = components['schemas']['OpportunityV1'];
type Project        = components['schemas']['ProjectV1'];
type Proposal       = components['schemas']['ProposalV1'];
type Session        = components['schemas']['SessionV1'];
type Mission        = components['schemas']['MissionV1'];
type WorkItem       = components['schemas']['WorkItemV1'];
type SourceRef      = components['schemas']['SourceRefV1'];
type DimScore       = components['schemas']['DimensionScoreV1'];

export function isFixtureMode(query: URLSearchParams): boolean {
  return query.get('fixture') === '1';
}

// ── SourceRefs for evidence ──
const ref1: SourceRef = { material_id: 'mat-001', revision: 1, locator: 'arxiv:2024.astro.12345', span: null };
const ref2: SourceRef = { material_id: 'mat-002', revision: 2, locator: 'pubmed:38901234', span: '§3.2' };
const ref3: SourceRef = { material_id: 'mat-004', revision: 1, locator: 'biorxiv:10.1101/2024.06.15.600123', span: null };

// ── Helper for dimension scores ──
function dim(value: number | null, reason: string): DimScore {
  return { value, reason };
}

// ── Materials ──
export const fixtureMaterials: Material[] = [
  {
    id: 'mat-001',
    source_locator: 'arxiv:2024.astro.12345',
    kind: 'paper',
    current_revision: 1,
    lifecycle: 'active',
    title: '基于图神经网络的星形胶质细胞钙信号动力学建模与仿真：一种多尺度计算方法的研究框架与初步验证',
    collection_reason: '与AD早期钙信号研究高度相关',
    source_spans: ['§1 Introduction', '§3 Methods'],
    import_status: 'succeeded',
    human_usage_count: 5,
    agent_usage_count: 12,
  },
  {
    id: 'mat-002',
    source_locator: 'pubmed:38901234',
    kind: 'paper',
    current_revision: 2,
    lifecycle: 'active',
    title: 'Astrocyte-mediated synaptic modulation and network-level computation in the prefrontal cortex during working memory tasks',
    collection_reason: null,
    source_spans: undefined,
    import_status: 'succeeded',
    human_usage_count: 0,
    agent_usage_count: 3,
  },
  {
    id: 'mat-003',
    source_locator: 'user:upload-77',
    kind: 'file',
    current_revision: 1,
    lifecycle: 'active',
    title: '超短文本',
    collection_reason: '手动上传测试',
    source_spans: undefined,
    import_status: 'failed',
    human_usage_count: 0,
    agent_usage_count: 0,
  },
  {
    id: 'mat-004',
    source_locator: 'biorxiv:10.1101/2024.06.15.600123',
    kind: 'paper',
    current_revision: 1,
    lifecycle: 'archived',
    title: '单细胞转录组学揭示星形胶质细胞异质性在神经退行性疾病中的功能分化与区域特异性表达模式——附补充数据集分析',
    collection_reason: '胶质细胞异质性关键文献',
    source_spans: ['Abstract', '§4 Results', 'Supplementary Table S2'],
    import_status: 'succeeded',
    human_usage_count: 2,
    agent_usage_count: 8,
  },
  {
    id: 'mat-005',
    source_locator: 'youtube:astrocyte-overview-2024',
    kind: 'video',
    current_revision: 1,
    lifecycle: 'active',
    title: '星形胶质细胞研究进展综述报告（2024年国际神经科学大会主题演讲录像）',
    collection_reason: '视频导入方向: summarize',
    source_spans: undefined,
    import_status: 'queued',
    human_usage_count: 0,
    agent_usage_count: 0,
  },
];

// ── Opportunities ──
export const fixtureOpportunities: Opportunity[] = [
  {
    id: 'opp-001',
    revision: 3,
    state: 'incubating',
    title: '胶质细胞-神经元代谢偶联在阿尔茨海默病早期诊断中的潜在生物标志物价值评估',
    evidence_refs: [ref1, ref2],
    goal_refs: ['goal-calcium-signaling'],
    dimensions: {
      goal_progress: dim(0.65, '已有初步证据支持'),
      current_interest: dim(0.82, '近期文献增长明显'),
      project_improvement: dim(0.4, '需要更多验证'),
      originality: dim(0.7, '方向较新'),
    },
    next_step: '收集更多关于血液生物标志物的临床证据',
    missing_evidence: ['临床队列数据', '纵向跟踪研究'],
  },
  {
    id: 'opp-002',
    revision: 1,
    state: 'deferred',
    title: 'Targeting astrocytic connexin 43 hemichannels for neuroprotection in acute ischemic stroke — translational gap analysis',
    evidence_refs: [],
    goal_refs: undefined,
    dimensions: {
      goal_progress: dim(null, 'unknown'),
      current_interest: dim(null, 'unknown'),
      project_improvement: dim(null, 'unknown'),
      originality: dim(0.3, '已有大量同类综述'),
    },
    next_step: '等待新证据出现后重新评估',
    missing_evidence: undefined,
  },
  {
    id: 'opp-003',
    revision: 5,
    state: 'ready_for_review',
    title: '光遗传学工具在活体星形胶质细胞功能研究中的新应用方向',
    evidence_refs: [ref1, ref3],
    goal_refs: ['goal-optogenetics'],
    dimensions: {
      goal_progress: dim(0.91, '多条证据链汇聚'),
      current_interest: dim(0.88, '领域热点'),
      project_improvement: dim(0.75, '直接支撑实验设计'),
      originality: dim(0.6, '方法已有先例'),
    },
    next_step: '提交审核，建议纳入研究规划',
    missing_evidence: undefined,
  },
  {
    id: 'opp-004',
    revision: 2,
    state: 'rejected',
    title: '星形胶质细胞反应性表型转换的表观遗传调控机制',
    evidence_refs: [ref2],
    goal_refs: undefined,
    dimensions: {
      goal_progress: dim(0.2, '证据不足'),
      current_interest: dim(0.1, '领域关注度下降'),
      project_improvement: dim(0.05, '与当前项目关联度低'),
      originality: dim(0.4, '已有类似研究'),
    },
    next_step: '无需后续行动',
    missing_evidence: ['关键机制验证数据'],
  },
];

// ── Projects ──
export const fixtureProjects: Project[] = [
  {
    id: 'proj-001',
    name: '阿尔茨海默病星形胶质细胞钙信号异常与β-淀粉样蛋白清除效率的多尺度建模研究项目（2025年度重点课题）',
    environment_id: 'env-ad-research',
    root_path: '/projects/ad-astrocyte-calcium',
    repo_id: 'repo-001',
    worktree_id: null,
  },
  {
    id: 'proj-002',
    name: 'Reactive astrogliosis heterogeneity mapping across neurodegenerative contexts',
    environment_id: 'env-neuro-map',
    root_path: '/projects/reactive-gliosis-map',
    repo_id: null,
    worktree_id: null,
  },
  {
    id: 'proj-003',
    name: '血脑屏障与胶质界膜的相互作用：文献计量与趋势分析',
    environment_id: 'env-bbb-review',
    root_path: '/projects/bbb-glia-interaction',
    repo_id: 'repo-003',
    worktree_id: 'wt-003',
  },
];

// ── Proposals ──
export const fixtureProposals: Proposal[] = [
  {
    id: 'prop-001',
    revision: 2,
    status: 'in_review',
    title: '系统性综述：星形胶质细胞A1/A2极化谱系在帕金森病与肌萎缩侧索硬化中的差异表达分析',
    goal: '识别A1/A2极化在不同神经退行性疾病中的特异性表达模式',
    scope: {
      project_ids: ['proj-001'],
      allowed_roots: ['/projects/ad-astrocyte-calcium'],
      allowed_actions: ['search', 'extract', 'synthesize'],
      data_egress_rules: [],
    },
    deliverables: ['系统综述报告', '偏倚风险评估矩阵', 'PRISMA流程图'],
    budget: { calls: null, tokens: 500000, money: null, currency: null },
    stop_conditions: ['检索文献超过500篇', '时间超过72小时'],
    candidate_refs: [ref1, ref3],
  },
  {
    id: 'prop-002',
    revision: 1,
    status: 'approved',
    title: 'Gap analysis: astrocyte-targeted drug delivery strategies crossing the blood-brain barrier',
    goal: 'Identify translational gaps in astrocyte-targeted nanomedicine',
    scope: {
      project_ids: ['proj-002'],
      allowed_roots: ['/projects/reactive-gliosis-map'],
      allowed_actions: ['search', 'extract', 'review'],
      data_egress_rules: [],
    },
    deliverables: ['Gap analysis report', 'Evidence summary'],
    budget: { calls: 100, tokens: null, money: 50, currency: 'USD' },
    stop_conditions: ['Review complete'],
    candidate_refs: undefined,
  },
  {
    id: 'prop-003',
    revision: 1,
    status: 'declined',
    title: '基于知识图谱的胶质细胞-突触相互作用网络构建与可视化分析平台设计',
    goal: '构建交互式知识图谱可视化平台',
    scope: {
      project_ids: [],
      allowed_roots: [],
      allowed_actions: [],
      data_egress_rules: [],
    },
    deliverables: ['原型设计', '技术方案文档'],
    budget: { calls: null, tokens: null, money: null, currency: null },
    stop_conditions: [],
    candidate_refs: undefined,
  },
];

// ── Sessions ──
export const fixtureSessions: Session[] = [
  {
    id: 'sess-001',
    adapter: 'claude-code',
    native_session_id: 'native-001',
    project_id: 'proj-001',
    worktree_id: null,
    binding_status: 'bound',
    capabilities: { native_resume: false, handoff: false },
    context_packet_id: 'ctx-pkt-001',
    context_state: 'current',
  },
  {
    id: 'sess-002',
    adapter: 'codex',
    native_session_id: null,
    project_id: 'proj-002',
    worktree_id: null,
    binding_status: 'unavailable',
    capabilities: { native_resume: false, handoff: false },
    context_packet_id: 'ctx-pkt-002',
    context_state: 'stale',
  },
  {
    id: 'sess-003',
    adapter: 'fixture',
    native_session_id: null,
    project_id: 'proj-001',
    worktree_id: null,
    binding_status: 'observed',
    capabilities: { native_resume: false, handoff: false },
    context_packet_id: null,
    context_state: 'unknown',
  },
];

// ── Missions ──
export const fixtureMissions: Mission[] = [
  {
    id: 'miss-001',
    version: 2,
    grant_id: 'grant-ad-calcium-2025',
    goal: '对AD早期星形胶质细胞钙信号相关文献进行系统筛选与质量评估，输出PRISMA流程图与偏倚风险矩阵',
    status: 'running',
    work_items: [
      { id: 'wi-001', version: 1, status: 'in_progress', holder_id: 'agent-atlas-01', context_packet_id: 'ctx-pkt-001', artifact_ids: [] },
      { id: 'wi-002', version: 1, status: 'open', holder_id: null, context_packet_id: null, artifact_ids: [] },
    ],
    artifact_ids: [],
    blockers: undefined,
  },
  {
    id: 'miss-002',
    version: 1,
    grant_id: 'grant-nano-delivery',
    goal: 'Generate a comprehensive gap analysis report on astrocyte-targeted nanomedicine delivery approaches for BBB crossing',
    status: 'pending',
    work_items: undefined,
    artifact_ids: undefined,
    blockers: undefined,
  },
  {
    id: 'miss-003',
    version: 3,
    grant_id: 'grant-kg-construction',
    goal: '构建胶质细胞钙信号与Aβ清除效率的因果关系知识图谱',
    status: 'completed',
    work_items: [
      { id: 'wi-003', version: 2, status: 'accepted', holder_id: 'agent-synthesizer', context_packet_id: 'ctx-pkt-003', artifact_ids: ['art-001', 'art-002'] },
      { id: 'wi-005', version: 1, status: 'accepted', holder_id: 'agent-reviewer', context_packet_id: null, artifact_ids: ['art-003'] },
    ],
    artifact_ids: ['art-001', 'art-002', 'art-003'],
    blockers: undefined,
  },
  {
    id: 'miss-004',
    version: 1,
    grant_id: 'grant-import-batch',
    goal: '失败的文献导入任务——服务器超时导致中断（含500条CSV记录的批量导入）',
    status: 'failed',
    work_items: [
      { id: 'wi-004', version: 1, status: 'cancelled', holder_id: null, context_packet_id: null, artifact_ids: [] },
    ],
    artifact_ids: [],
    blockers: ['服务器连接超时，需要重试'],
  },
  {
    id: 'miss-005',
    version: 1,
    grant_id: 'grant-blocked-mission',
    goal: '被阻塞的跨机构合作研究任务——等待外部数据授权',
    status: 'blocked',
    work_items: undefined,
    artifact_ids: undefined,
    blockers: ['外部数据库授权未到位', '合规审查进行中'],
  },
];

// ── Flatten work items and artifacts from missions for display ──

export function flattenWorkItems(missions: Mission[]): WorkItem[] {
  return missions.flatMap(m => m.work_items ?? []);
}

export function flattenArtifactIds(missions: Mission[]): string[] {
  return missions.flatMap(m => m.artifact_ids ?? []);
}
