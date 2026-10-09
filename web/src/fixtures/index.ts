/**
 * Fixture mode detection and data.
 *
 * Fixture mode is ONLY activated by ?fixture=1 in the URL.
 * Provides neutral sample data exercising long Chinese text, unknown fields,
 * edge cases, and various statuses.
 *
 * API failures NEVER silently fall back to fixtures.
 *
 * Types match the V1 generated schema from web/src/api/schema.d.ts.
 * All source locators are explicitly fixture-prefixed (not real arXiv/PubMed/DOI).
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
const ref1: SourceRef = { material_id: 'mat-001', revision: 1, locator: 'fixture-source-alpha', span: null };
const ref2: SourceRef = { material_id: 'mat-002', revision: 2, locator: 'fixture-source-beta', span: '§3.2' };
const ref3: SourceRef = { material_id: 'mat-004', revision: 1, locator: 'fixture-source-delta', span: null };

// ── Helper for dimension scores ──
function dim(value: number | null, reason: string): DimScore {
  return { value, reason };
}

// ── Materials ──
export const fixtureMaterials: Material[] = [
  {
    id: 'mat-001',
    source_locator: 'fixture:paper-long-chinese-title',
    kind: 'paper',
    current_revision: 1,
    lifecycle: 'active',
    title: '示例论文：这是一个非常长的中文素材标题，用于测试文本截断和多行显示行为是否在各种视口宽度下正常工作',
    collection_reason: '示例收集原因',
    source_spans: ['§1 引言', '§3 方法'],
    import_status: 'succeeded',
    human_usage_count: 5,
    agent_usage_count: 12,
  },
  {
    id: 'mat-002',
    source_locator: 'fixture:paper-english-medium',
    kind: 'paper',
    current_revision: 2,
    lifecycle: 'active',
    title: 'Fixture Sample Paper: Medium-length English Title for Layout Testing',
    collection_reason: null,
    source_spans: undefined,
    import_status: 'succeeded',
    human_usage_count: 0,
    agent_usage_count: 3,
  },
  {
    id: 'mat-003',
    source_locator: 'fixture:file-short-title',
    kind: 'file',
    current_revision: 1,
    lifecycle: 'active',
    title: '短标题',
    collection_reason: '手动上传测试',
    source_spans: undefined,
    import_status: 'failed',
    human_usage_count: 0,
    agent_usage_count: 0,
  },
  {
    id: 'mat-004',
    source_locator: 'fixture:paper-archived-long',
    kind: 'paper',
    current_revision: 1,
    lifecycle: 'archived',
    title: '示例已归档论文：此标题故意设计得较长，以验证归档状态标签和长标题截断在同一行内的显示效果——附补充说明文字',
    collection_reason: '示例关键素材',
    source_spans: ['摘要', '§4 结果', '附表 S2'],
    import_status: 'succeeded',
    human_usage_count: 2,
    agent_usage_count: 8,
  },
  {
    id: 'mat-005',
    source_locator: 'fixture:video-summarize',
    kind: 'video',
    current_revision: 1,
    lifecycle: 'active',
    title: '示例视频素材（视频导入方向: summarize）',
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
    title: '示例机会：长标题用于测试孵化状态下的维度分数显示和证据引用数量',
    evidence_refs: [ref1, ref2],
    goal_refs: ['goal-example-01'],
    dimensions: {
      goal_progress: dim(0.65, '已有初步证据支持'),
      current_interest: dim(0.82, '近期关注增长明显'),
      project_improvement: dim(0.4, '需要更多验证'),
      originality: dim(0.7, '方向较新'),
    },
    next_step: '收集更多支撑证据',
    missing_evidence: ['补充数据 A', '补充数据 B'],
  },
  {
    id: 'opp-002',
    revision: 1,
    state: 'deferred',
    title: 'Fixture Deferred Opportunity: All Dimensions Unknown for Edge Case Testing',
    evidence_refs: [],
    goal_refs: undefined,
    dimensions: {
      goal_progress: dim(null, 'unknown'),
      current_interest: dim(null, 'unknown'),
      project_improvement: dim(null, 'unknown'),
      originality: dim(0.3, '已有大量同类方向'),
    },
    next_step: '等待新证据出现后重新评估',
    missing_evidence: undefined,
  },
  {
    id: 'opp-003',
    revision: 5,
    state: 'ready_for_review',
    title: '示例待审核机会：高维度分数用于验证高亮显示',
    evidence_refs: [ref1, ref3],
    goal_refs: ['goal-example-02'],
    dimensions: {
      goal_progress: dim(0.91, '多条证据链汇聚'),
      current_interest: dim(0.88, '关注度高'),
      project_improvement: dim(0.75, '直接支撑当前项目'),
      originality: dim(0.6, '方法已有先例'),
    },
    next_step: '提交审核',
    missing_evidence: undefined,
  },
  {
    id: 'opp-004',
    revision: 2,
    state: 'rejected',
    title: '示例已拒绝机会',
    evidence_refs: [ref2],
    goal_refs: undefined,
    dimensions: {
      goal_progress: dim(0.2, '证据不足'),
      current_interest: dim(0.1, '关注度低'),
      project_improvement: dim(0.05, '关联度低'),
      originality: dim(0.4, '已有类似方向'),
    },
    next_step: '无需后续行动',
    missing_evidence: ['关键验证数据'],
  },
];

// ── Projects ──
export const fixtureProjects: Project[] = [
  {
    id: 'proj-001',
    name: '示例项目：这是一个故意设计的非常长的中文项目名称，用于测试项目名称在卡片布局中的截断和换行行为（含年份标记 2025）',
    environment_id: 'fixture-env-alpha',
    root_path: '/fixture/projects/alpha',
    repo_id: 'fixture-repo-001',
    worktree_id: null,
  },
  {
    id: 'proj-002',
    name: 'Fixture Project Beta',
    environment_id: 'fixture-env-beta',
    root_path: '/fixture/projects/beta',
    repo_id: null,
    worktree_id: null,
  },
  {
    id: 'proj-003',
    name: '示例项目丙',
    environment_id: 'fixture-env-gamma',
    root_path: '/fixture/projects/gamma',
    repo_id: 'fixture-repo-003',
    worktree_id: 'fixture-wt-003',
  },
];

// ── Proposals ──
export const fixtureProposals: Proposal[] = [
  {
    id: 'prop-001',
    revision: 2,
    status: 'in_review',
    title: '示例提案：系统性综述方向的差异表达分析',
    goal: '识别不同条件下的特异性表达模式',
    scope: {
      project_ids: ['proj-001'],
      allowed_roots: ['/fixture/projects/alpha'],
      allowed_actions: ['search', 'extract', 'synthesize'],
      data_egress_rules: [],
    },
    deliverables: ['综述报告', '风险评估矩阵', '流程图'],
    budget: { calls: null, tokens: 500000, money: null, currency: null },
    stop_conditions: ['文献超过500篇', '时间超过72小时'],
    candidate_refs: [ref1, ref3],
  },
  {
    id: 'prop-002',
    revision: 1,
    status: 'approved',
    title: 'Fixture Approved Proposal: Gap Analysis',
    goal: 'Identify gaps in the current approach',
    scope: {
      project_ids: ['proj-002'],
      allowed_roots: ['/fixture/projects/beta'],
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
    title: '示例已拒绝提案',
    goal: '构建可视化分析平台',
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
    adapter: 'fixture-adapter-a',
    native_session_id: 'fixture-native-001',
    project_id: 'proj-001',
    worktree_id: null,
    binding_status: 'bound',
    capabilities: { native_resume: false, handoff: false },
    context_packet_id: 'fixture-ctx-pkt-001',
    context_state: 'current',
  },
  {
    id: 'sess-002',
    adapter: 'fixture-adapter-b',
    native_session_id: null,
    project_id: 'proj-002',
    worktree_id: null,
    binding_status: 'unavailable',
    capabilities: { native_resume: null, handoff: null },
    context_packet_id: 'fixture-ctx-pkt-002',
    context_state: 'stale',
  },
  {
    id: 'sess-003',
    adapter: 'fixture-adapter-c',
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
    grant_id: 'fixture-grant-001',
    goal: '示例运行中任务：对系统筛选与质量评估流程进行验证（含多个工作项）',
    status: 'running',
    work_items: [
      { id: 'wi-001', version: 1, status: 'in_progress', holder_id: 'fixture-holder-01', context_packet_id: 'fixture-ctx-pkt-001', artifact_ids: [] },
      { id: 'wi-002', version: 1, status: 'open', holder_id: null, context_packet_id: null, artifact_ids: [] },
    ],
    artifact_ids: [],
    blockers: undefined,
  },
  {
    id: 'miss-002',
    version: 1,
    grant_id: 'fixture-grant-002',
    goal: 'Fixture Pending Mission: No Work Items',
    status: 'pending',
    work_items: undefined,
    artifact_ids: undefined,
    blockers: undefined,
  },
  {
    id: 'miss-003',
    version: 3,
    grant_id: 'fixture-grant-003',
    goal: '示例已完成任务',
    status: 'completed',
    work_items: [
      { id: 'wi-003', version: 2, status: 'accepted', holder_id: 'fixture-holder-02', context_packet_id: 'fixture-ctx-pkt-003', artifact_ids: ['fixture-art-001', 'fixture-art-002'] },
      { id: 'wi-005', version: 1, status: 'accepted', holder_id: 'fixture-holder-03', context_packet_id: null, artifact_ids: ['fixture-art-003'] },
    ],
    artifact_ids: ['fixture-art-001', 'fixture-art-002', 'fixture-art-003'],
    blockers: undefined,
  },
  {
    id: 'miss-004',
    version: 1,
    grant_id: 'fixture-grant-004',
    goal: '示例失败任务——服务器超时导致中断',
    status: 'failed',
    work_items: [
      { id: 'wi-004', version: 1, status: 'cancelled', holder_id: null, context_packet_id: null, artifact_ids: [] },
    ],
    artifact_ids: [],
    blockers: ['服务器连接超时'],
  },
  {
    id: 'miss-005',
    version: 1,
    grant_id: 'fixture-grant-005',
    goal: '示例被阻塞任务',
    status: 'blocked',
    work_items: undefined,
    artifact_ids: undefined,
    blockers: ['外部数据授权未到位', '审查进行中'],
  },
];

// ── Flatten work items and artifacts from missions for display ──

export function flattenWorkItems(missions: Mission[]): WorkItem[] {
  return missions.flatMap(m => m.work_items ?? []);
}

export function flattenArtifactIds(missions: Mission[]): string[] {
  return missions.flatMap(m => m.artifact_ids ?? []);
}
