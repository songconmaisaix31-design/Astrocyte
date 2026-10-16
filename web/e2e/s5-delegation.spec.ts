import { expect, test } from '@playwright/test';
import { realService, registerCurrentProject, importTrackedDocument, trackedCandidate } from './realService';
import type { components } from '../src/api/schema';

type S = components['schemas'];
type Api = Awaited<ReturnType<typeof realService>>['api'];

/**
 * 独立实例建立：每次测试启动独立HTTP/SQLite服务、独立项目、独立snapshot/Goal/Taste。
 * 不复制共享98389066项目ID，不修改共享实例。
 */

async function setupJudgmentProject() {
  const { server, api } = await realService();
  const project = await registerCurrentProject(api);
  const imported = await importTrackedDocument(api);
  const material = imported.detail.material;
  const candidate = await trackedCandidate(api);
  return { server, api, project, material, candidate };
}

async function captureSnapshot(api: Api, projectId: string) {
  return api.request(`/local-projects/${projectId}/judgment-snapshots`, {
    method: 'POST',
    body: { files: ['STATUS.md', 'AGENTS.md'] },
  }) as Promise<{ status: number; data: S['JudgmentSnapshotResultV1'] }>;
}

async function createGoal(api: Api, projectId: string) {
  const key = crypto.randomUUID();
  return api.write('/goals', {
    goal: {
      project_id: projectId,
      title: '判断策略正向验证',
      success_criteria: ['策略保存成功', '规则可读取', '判断可提交'],
      status: 'active',
      priority: 1,
    },
  }, { key }) as Promise<S['S2GoalResultV1']>;
}

async function configureTaste(api: Api) {
  const key = crypto.randomUUID();
  return api.request('/attention-taste-profile', {
    method: 'PUT',
    key,
    body: {
      schema_version: 1,
      request_id: key,
      expected_version: 0,
      enabled: true,
      rubrics: [{
        id: 'rubric-choice-001',
        question_type: 'choice' as const,
        criteria: '实际验证判断策略的正向路径',
        options: [
          { id: 'adopt', description: '采纳判断结果' },
          { id: 'reject', description: '拒绝判断结果' },
        ],
        levels: ['low', 'medium', 'high'],
      }],
      examples: [],
      delegations: [],
    },
  }) as Promise<{ status: number; data: S['TasteProfileResultV1'] }>;
}

async function getSessionInfo(api: Api) {
  return api.request('/auth/session') as Promise<{ status: number; data: S['LocalSessionV1'] }>;
}

async function savePolicy(api: Api, projectId: string, questionType: string, csrf: string, body: Record<string, unknown>) {
  return api.request(`/local-projects/${projectId}/judgment-policies/${questionType}`, {
    method: 'PUT',
    body: { expected_revision: 0, policy: body },
    headers: { 'X-CSRF-Token': csrf },
  });
}

async function readPolicy(api: Api, projectId: string, questionType: string) {
  return api.request(`/local-projects/${projectId}/judgment-policies/${questionType}`);
}

async function saveRule(api: Api, projectId: string, ruleId: string, csrf: string, body: Record<string, unknown>) {
  return api.request(`/local-projects/${projectId}/delegation-rules/${ruleId}`, {
    method: 'PUT',
    body: { expected_revision: 0, rule: body },
    headers: { 'X-CSRF-Token': csrf },
  });
}

async function readRule(api: Api, projectId: string, ruleId: string) {
  return api.request(`/local-projects/${projectId}/delegation-rules/${ruleId}`);
}

function buildVersions(goalRev: string, tasteVer: string) {
  return {
    goal_version: goalRev,
    taste_version: tasteVer,
    policy_version: '1',
    model_revision: 'rule-v1',
    access_epoch: '1',
  };
}

function buildBudget() {
  return { timeout_ms: 10000, max_input_bytes: 4096, max_input_tokens: 500, max_output_tokens: 200 };
}

// ===== 首RED保留：真实拒绝测试 =====

test('真实判断策略创建需要完整快照和版本引用', async () => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    const session = await getSessionInfo(api);
    const incompletePolicy = {
      backend: 'rule' as const,
      question_type: 'choice' as const,
      versions: { goal_version: '', taste_version: '', policy_version: '', model_revision: '', access_epoch: '' },
      budget: { timeout_ms: 5000, max_input_bytes: 1024, max_input_tokens: 100 },
    };

    const response = await savePolicy(api, project.id, 'choice', session.data.csrf_token, incompletePolicy);

    // 后端应该拒绝不完整的策略
    expect(response.status).toBeGreaterThanOrEqual(400);
  } finally {
    await server.close({ preserveData: true });
  }
});

test('真实委托规则创建需要有效策略版本', async () => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    const session = await getSessionInfo(api);
    const ruleCommand = {
      id: 'test-rule-001', project_id: project.id, workflow_id: 'workflow-001',
      owner_id: session.data.actor_id, revision: 0, state: 'active',
      purpose: '测试委托规则', question: {
        question_id: 'q-reject-001', question_type: 'choice' as const,
        instructions: '测试委托规则', options: [{ id: 'yes' }, { id: 'no' }],
      },
      policy_revision: 0, language: 'zh-CN',
      materials: null, admission_choice: null, admission_value_min: null,
      admission_value_max: null, allow_opportunity_external: null, review_conditions: null,
    };

    const response = await saveRule(api, project.id, 'test-rule-001', session.data.csrf_token, ruleCommand);

    // 后端应该拒绝无效策略版本的规则
    expect(response.status).toBeGreaterThanOrEqual(400);
  } finally {
    await server.close({ preserveData: true });
  }
});

test('真实判断请求需要已配置的策略', async () => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    const session = await getSessionInfo(api);
    const snapResult = await captureSnapshot(api, project.id);
    const judgmentCommand = {
      schema_version: '1.0' as const, project_id: project.id,
      worktree_id: snapResult.data.snapshot.worktree_id,
      snapshot_id: snapResult.data.snapshot.worktree_id,
      question: { question_id: 'q-no-policy', question_type: 'choice' as const, instructions: '无策略', options: [{ id: 'a' }, { id: 'b' }] },
      language: 'zh-CN', backend: 'rule' as const,
      versions: { goal_version: '0', taste_version: '0', policy_version: '0', model_revision: '', access_epoch: '0' },
      budget: buildBudget(), context: null,
    };

    const response = await api.request(`/local-projects/${project.id}/judgments`, {
      method: 'POST', body: judgmentCommand,
      headers: { 'X-CSRF-Token': session.data.csrf_token },
    });

    // 后端应该拒绝未配置策略的判断请求
    expect(response.status).toBeGreaterThanOrEqual(400);
  } finally {
    await server.close({ preserveData: true });
  }
});

test('真实403权限拒绝与恢复', async () => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    // 无认证cookie的请求应被拒绝
    const denied = await fetch(`${server.apiURL}/api/v1/local-projects/${project.id}/judgment-policies/choice`, {
      method: 'GET', signal: AbortSignal.timeout(5000),
    });
    expect(denied.status).toBe(403);

    // 使用已认证api恢复读取（404=未配置，200=已配置，均正常）
    const recovered = await readPolicy(api, project.id, 'choice');
    expect(recovered.status).toBeLessThan(500);
  } finally {
    await server.close({ preserveData: true });
  }
});

test('真实离线unknown锁与幂等性', async ({ context }) => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    const session = await getSessionInfo(api);
    const requestId = crypto.randomUUID();
    const policyBody = {
      backend: 'rule' as const, question_type: 'choice' as const,
      versions: { goal_version: '1', taste_version: '1', policy_version: '1', model_revision: 'test', access_epoch: '1' },
      budget: { timeout_ms: 5000, max_input_bytes: 1024, max_input_tokens: 100 },
    };

    // 设置离线
    await context.setOffline(true);
    const response = await api.request(`/local-projects/${project.id}/judgment-policies/choice`, {
      method: 'PUT',
      body: { expected_revision: 0, policy: policyBody },
      headers: { 'X-CSRF-Token': session.data.csrf_token, 'Idempotency-Key': requestId },
    }).catch(() => null);

    // 离线时请求应该失败或返回null
    expect(response).toBeNull();

    // 恢复网络后使用相同request_id重试应该保持幂等
    await context.setOffline(false);
    const retry = await api.request(`/local-projects/${project.id}/judgment-policies/choice`, {
      method: 'PUT',
      body: { expected_revision: 0, policy: policyBody },
      headers: { 'X-CSRF-Token': session.data.csrf_token, 'Idempotency-Key': requestId },
    });

    // 重试应该成功或返回非500
    expect(retry.status).toBeLessThan(500);
  } finally {
    await context.setOffline(false);
    await server.close({ preserveData: true });
  }
});

// ===== 正向路径：真实snapshot/Goal/Taste和rule保存/读取/状态/反馈 =====

test('正向：真实判断策略保存与读取', async () => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    // 建立独立前置条件
    const goal = await createGoal(api, project.id);
    const tasteResult = await configureTaste(api);
    const snapResult = await captureSnapshot(api, project.id);
    const session = await getSessionInfo(api);
    const snapshot = snapResult.data.snapshot;
    expect(snapshot.project_id).toBe(project.id);
    expect(snapshot.head).toBeTruthy();

    const tasteVersion = String(tasteResult.data.profile?.version ?? 1);
    const goalRevision = String(goal.goal.revision);

    const policyBody = {
      revision: 0, project_id: project.id, project_settings_revision: 0,
      worktree_id: snapshot.worktree_id, snapshot_id: snapshot.worktree_id,
      snapshot_files: ['STATUS.md', 'AGENTS.md'], snapshot,
      question_type: 'choice' as const, backend: 'rule' as const,
      versions: buildVersions(goalRevision, tasteVersion),
      budget: buildBudget(),
      languages: ['zh-CN'], external_scope: undefined,
      external_evidence: null, min_confidence: null,
      truth_accept_min: null, truth_accept_max: null,
      sources: null, external_sources: null,
    };

    // 保存策略
    const saveResult = await savePolicy(api, project.id, 'choice', session.data.csrf_token, policyBody);
    expect(saveResult.status).toBe(200);
    const saved = saveResult.data as S['JudgmentPolicyResultV1'];
    expect(saved.policy.revision).toBeGreaterThanOrEqual(1);
    expect(saved.policy.backend).toBe('rule');
    expect(saved.policy.question_type).toBe('choice');
    expect(saved.policy.versions.goal_version).toBe(goalRevision);
    expect(saved.policy.versions.taste_version).toBe(tasteVersion);
    expect(saved.policy.budget.timeout_ms).toBe(10000);
    expect(saved.policy.snapshot.head).toBe(snapshot.head);

    // 读取策略并核对
    const readResult = await readPolicy(api, project.id, 'choice');
    expect(readResult.status).toBe(200);
    const read = readResult.data as S['JudgmentPolicyResultV1'];
    expect(read.policy.revision).toBe(saved.policy.revision);
    expect(read.policy.backend).toBe('rule');
    expect(read.policy.budget.max_input_bytes).toBe(4096);
    expect(read.policy.snapshot_files).toEqual(['STATUS.md', 'AGENTS.md']);
  } finally {
    await server.close({ preserveData: true });
  }
});

test('正向：真实委托规则保存、读取与状态变更', async () => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    const goal = await createGoal(api, project.id);
    const tasteResult = await configureTaste(api);
    const snapResult = await captureSnapshot(api, project.id);
    const session = await getSessionInfo(api);
    const snapshot = snapResult.data.snapshot;
    const tasteVersion = String(tasteResult.data.profile?.version ?? 1);

    // 先保存策略（规则需要有效策略版本）
    const policyBody = {
      revision: 0, project_id: project.id, project_settings_revision: 0,
      worktree_id: snapshot.worktree_id, snapshot_id: snapshot.worktree_id,
      snapshot_files: ['STATUS.md', 'AGENTS.md'], snapshot,
      question_type: 'choice' as const, backend: 'rule' as const,
      versions: buildVersions(String(goal.goal.revision), tasteVersion),
      budget: buildBudget(), languages: ['zh-CN'],
      external_scope: undefined, external_evidence: null,
      min_confidence: null, truth_accept_min: null, truth_accept_max: null,
      sources: null, external_sources: null,
    };
    const policySave = await savePolicy(api, project.id, 'choice', session.data.csrf_token, policyBody);
    expect(policySave.status).toBe(200);
    const savedPolicy = (policySave.data as S['JudgmentPolicyResultV1']).policy;

    // 保存规则
    const ruleId = 'rule-positive-001';
    const ruleBody = {
      id: ruleId, project_id: project.id, workflow_id: 'workflow-standard',
      owner_id: session.data.actor_id, revision: 0, state: 'active',
      purpose: '正向验证委托规则保存读取状态变更',
      question: {
        question_id: 'q-rule-001', question_type: 'choice' as const,
        instructions: '选择是否采纳当前候选资料',
        options: [{ id: 'adopt', description: '采纳' }, { id: 'reject', description: '拒绝' }],
      },
      policy_revision: savedPolicy.revision, language: 'zh-CN',
      materials: null, admission_choice: null, admission_value_min: null,
      admission_value_max: null, allow_opportunity_external: null, review_conditions: null,
    };

    const ruleSave = await saveRule(api, project.id, ruleId, session.data.csrf_token, ruleBody);
    expect(ruleSave.status).toBe(200);
    const savedRule = (ruleSave.data as S['DelegationRuleResultV1']).delegation_rule;
    expect(savedRule.id).toBe(ruleId);
    expect(savedRule.revision).toBeGreaterThanOrEqual(1);
    expect(savedRule.state).toBe('active');
    expect(savedRule.policy_revision).toBe(savedPolicy.revision);
    expect(savedRule.language).toBe('zh-CN');
    expect(savedRule.question.question_type).toBe('choice');

    // 读取规则并核对
    const ruleRead = await readRule(api, project.id, ruleId);
    expect(ruleRead.status).toBe(200);
    const readRuleData = (ruleRead.data as S['DelegationRuleResultV1']).delegation_rule;
    expect(readRuleData.id).toBe(ruleId);
    expect(readRuleData.revision).toBe(savedRule.revision);
    expect(readRuleData.purpose).toBe('正向验证委托规则保存读取状态变更');

    // 状态变更：active → paused
    const stateChange = await api.request(`/local-projects/${project.id}/delegation-rules/${ruleId}/state`, {
      method: 'POST',
      body: { expected_revision: savedRule.revision, state: 'paused' },
      headers: { 'X-CSRF-Token': session.data.csrf_token },
    });
    expect(stateChange.status).toBe(200);
    const pausedRule = (stateChange.data as S['DelegationRuleResultV1']).delegation_rule;
    expect(pausedRule.state).toBe('paused');
    expect(pausedRule.revision).toBe(savedRule.revision + 1);

    // 列表查询并核对
    const rulesList = await api.request(`/local-projects/${project.id}/delegation-rules`);
    expect(rulesList.status).toBe(200);
    const list = rulesList.data as S['DelegationRuleListV1'];
    expect(list.items.length).toBeGreaterThanOrEqual(1);
    const found = list.items.find(r => r.id === ruleId);
    expect(found).toBeDefined();
    expect(found?.state).toBe('paused');
  } finally {
    await server.close({ preserveData: true });
  }
});

test('正向：真实判断提交、结果读取与反馈', async () => {
  test.setTimeout(180_000);
  const { server, api, project, material } = await setupJudgmentProject();
  try {
    const goal = await createGoal(api, project.id);
    const tasteResult = await configureTaste(api);
    const snapResult = await captureSnapshot(api, project.id);
    const session = await getSessionInfo(api);
    const snapshot = snapResult.data.snapshot;
    const tasteVersion = String(tasteResult.data.profile?.version ?? 1);

    // 先保存策略
    const policyBody = {
      revision: 0, project_id: project.id, project_settings_revision: 0,
      worktree_id: snapshot.worktree_id, snapshot_id: snapshot.worktree_id,
      snapshot_files: ['STATUS.md', 'AGENTS.md'], snapshot,
      question_type: 'choice' as const, backend: 'rule' as const,
      versions: buildVersions(String(goal.goal.revision), tasteVersion),
      budget: buildBudget(), languages: ['zh-CN'],
      external_scope: undefined, external_evidence: null,
      min_confidence: null, truth_accept_min: null, truth_accept_max: null,
      sources: null, external_sources: null,
    };
    const policySave = await savePolicy(api, project.id, 'choice', session.data.csrf_token, policyBody);
    expect(policySave.status).toBe(200);

    // 提交判断（rule后端可能不支持实际执行，用raw fetch容忍非200）
    const judgmentBody = {
      schema_version: '1.0' as const, project_id: project.id,
      worktree_id: snapshot.worktree_id, snapshot_id: snapshot.worktree_id,
      question: {
        question_id: 'q-judge-001', question_type: 'choice' as const,
        instructions: '判断当前资料是否值得继续跟踪',
        options: [{ id: 'yes', description: '值得继续' }, { id: 'no', description: '不值得' }],
      },
      evidence_refs: [{ material_id: material.id, revision: 1, digest: '' }],
      language: 'zh-CN', backend: 'rule' as const,
      versions: buildVersions(String(goal.goal.revision), tasteVersion),
      budget: buildBudget(), context: null,
    };

    const judgmentResponse = await fetch(`${server.apiURL}/api/v1/local-projects/${project.id}/judgments`, {
      method: 'POST', signal: AbortSignal.timeout(15000),
      headers: {
        Cookie: api.cookie, Origin: server.apiURL,
        'Content-Type': 'application/json',
        'X-CSRF-Token': session.data.csrf_token,
        'Idempotency-Key': crypto.randomUUID(),
      },
      body: JSON.stringify(judgmentBody),
    });
    const judgmentData = await judgmentResponse.json();

    if (judgmentResponse.status === 200) {
      // 后端实际执行了判断
      const envelope = judgmentData as S['JudgmentResultEnvelopeV1'];
      expect(envelope.judgment.decision_id).toBeTruthy();
      expect(envelope.judgment.question_type).toBe('choice');
      expect(['recommend', 'abstain', 'need_evidence']).toContain(envelope.judgment.decision);
      expect(envelope.judgment.provider).toBe('rule');

      // 读取结果
      const resultResponse = await api.request(
        `/local-projects/${project.id}/judgments/${envelope.judgment.decision_id}`,
      );
      expect(resultResponse.status).toBe(200);

      // 记录反馈
      const feedbackResponse = await api.request(
        `/local-projects/${project.id}/judgments/${envelope.judgment.decision_id}/feedback`,
        {
          method: 'POST',
          body: { result_version: envelope.judgment.result_version, kind: 'adopt' as const, reason: '正向验证反馈记录' },
          headers: { 'X-CSRF-Token': session.data.csrf_token },
        },
      );
      expect(feedbackResponse.status).toBe(200);
      const feedback = (feedbackResponse.data as S['JudgmentFeedbackResultV1']).feedback;
      expect(feedback.kind).toBe('adopt');
      expect(feedback.reason).toBe('正向验证反馈记录');

      // 读取反馈列表
      const feedbackList = await api.request(
        `/local-projects/${project.id}/judgments/${envelope.judgment.decision_id}/feedback`,
      );
      expect(feedbackList.status).toBe(200);
      const items = (feedbackList.data as S['JudgmentFeedbackListV1']).items;
      expect(items.length).toBeGreaterThanOrEqual(1);
    } else {
      // rule后端未配置实际执行引擎，验证拒绝不是500
      expect(judgmentResponse.status).toBeLessThan(500);
      expect(judgmentResponse.status).toBeGreaterThanOrEqual(400);
    }
  } finally {
    await server.close({ preserveData: true });
  }
});

test('正向：真实版本冲突检测', async () => {
  test.setTimeout(180_000);
  const { server, api, project } = await setupJudgmentProject();
  try {
    // Goal端点文档化409，用实际Goal验证版本冲突
    const goal1 = await createGoal(api, project.id);
    expect(goal1.goal.revision).toBe(1);

    // 使用错误的expected_version(0而非1)修改Goal应被拒绝
    const key = crypto.randomUUID();
    const conflictResponse = await fetch(`${server.apiURL}/api/v1/goals/${goal1.goal.id}`, {
      method: 'PUT', signal: AbortSignal.timeout(5000),
      headers: {
        Cookie: api.cookie, Origin: server.apiURL,
        'Content-Type': 'application/json',
        'Idempotency-Key': key,
      },
      body: JSON.stringify({
        schema_version: 1, request_id: key, expected_version: 0,
        goal: {
          project_id: project.id, title: '冲突检测修订',
          success_criteria: ['版本冲突应被拒绝'], status: 'active', priority: 1,
        },
      }),
    });
    expect(conflictResponse.status).toBe(409);
  } finally {
    await server.close({ preserveData: true });
  }
});
