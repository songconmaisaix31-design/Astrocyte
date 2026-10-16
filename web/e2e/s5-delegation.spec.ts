import { expect, test, type Locator, type Page } from '@playwright/test';
import { realService, registerCurrentProject, importTrackedDocument, repositoryRoot } from './realService';
import { humanAPI, importMaterial } from '../../tests/s1/api.mjs';
import { execFileSync } from 'node:child_process';
import { hostname } from 'node:os';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const tastePath = '/attention-taste-profile';

async function setup() {
  const { server, api } = await realService(true, { ASTROCYTE_REPOSITORY_ENVIRONMENT_ID: hostname() });
  const registered = await registerCurrentProject(api);
  const configured = await api.write(`/local-projects/${registered.id}/settings`, {
    expected_revision: registered.settings.revision,
    settings: { allow_directory: true, allowed_subdirs: ['STATUS.md', 'AGENTS.md'], allowed_actions: ['read_context'], allowed_tools: [], external_model_cli: '', allow_agent_control: false },
  }, { method: 'PUT' }) as S['LocalProjectResultV1'];
  const project = configured.project;
  expect(project.settings.revision).toBeGreaterThan(registered.settings.revision);
  const imported = await importTrackedDocument(api);
  const material = imported.detail.material as S['MaterialV1'];
  const revision = (imported.detail.revisions as S['MaterialRevisionV1'][]).find(item => item.revision === material.current_revision);
  if (!revision || !revision.content_digest) throw new Error('实际导入资料缺少当前固定版本或摘要');
  const { space } = await api.get(`/project-spaces/${project.space_id}`) as S['ProjectSpaceResultV1'];
  await api.write(`/project-spaces/${project.space_id}/references`, { expected_version: space.version, material_id: material.id, revision: revision.revision });
  return { server, api, project, material, revision };
}

async function openTaste(page: Page, webURL: string) {
  await page.goto(`${webURL}/attention`);
  const section = page.getByRole('region', { name: 'Taste 配置', exact: true });
  await expect(section.getByText('当前状态：未配置', { exact: true })).toBeVisible();
  await section.getByRole('button', { name: '创建配置', exact: true }).click();
  await expect(section.getByLabel('启用状态', { exact: true })).toHaveValue('');
  return section;
}

async function saveTaste(page: Page, section: Locator) {
  const response = page.waitForResponse(response => response.url().endsWith(`/api/v1${tastePath}`) && response.request().method() === 'PUT');
  await section.getByRole('button', { name: '保存配置', exact: true }).click();
  const result = await response;
  expect(result.status(), await result.text()).toBe(200);
  const data = await result.json() as S['TasteProfileResultV1'];
  if (!data.profile || data.profile.version < 1) throw new Error('Taste 保存响应缺少真实版本');
  await expect(section.getByRole('button', { name: '编辑配置', exact: true })).toBeEnabled();
  return data.profile;
}

async function fillTaste(page: Page, section: Locator, materialID: string, revision: number) {
  await section.getByLabel('启用状态', { exact: true }).selectOption('true');
  for (const kind of ['truth', 'choice', 'score'] as const) {
    await section.getByRole('button', { name: '添加评判准则', exact: true }).click();
    const rubric = section.locator('.ac-taste-rubric').last();
    await expect(rubric.getByLabel('准则题型', { exact: true })).toHaveValue('');
    await rubric.getByLabel('准则题型', { exact: true }).selectOption(kind);
    await rubric.getByLabel('评判标准', { exact: true }).fill(`人工核查公开执行协议的固定来源 · ${kind}`);
    if (kind === 'choice') {
      for (const option of [{ id: 'review', description: '审阅固定来源' }, { id: 'verify', description: '补充验证证据' }]) {
        await rubric.getByRole('button', { name: '添加选项', exact: true }).click();
        const field = rubric.getByRole('group', { name: /^选项 \d+$/ }).last();
        await field.getByLabel('选项标识', { exact: true }).fill(option.id);
        await field.getByLabel('选项说明', { exact: true }).fill(option.description);
      }
    }
    if (kind === 'score') {
      for (const [i, description] of ['补充来源', '审阅固定来源', '完成实际验证'].entries()) {
        await rubric.getByRole('button', { name: '添加等级', exact: true }).click();
        await rubric.getByLabel(`等级 ${i}`, { exact: true }).fill(description);
      }
    }
    await section.getByRole('button', { name: '添加固定资料示例', exact: true }).click();
    const example = section.locator('.ac-taste-example').last();
    await example.getByLabel('示例关联准则', { exact: true }).selectOption({ index: (await section.locator('.ac-taste-rubric').count()) });
    await example.getByLabel('示例资料', { exact: true }).selectOption(materialID);
    await expect(example.getByLabel('示例固定版本', { exact: true })).toBeEnabled();
    await expect(example.getByLabel('示例固定版本', { exact: true })).toHaveValue('');
    await example.getByLabel('示例固定版本', { exact: true }).selectOption(String(revision));
    if (kind === 'choice') await example.getByLabel('示例预期选项', { exact: true }).selectOption('review');
    else await example.getByLabel('示例预期数值', { exact: true }).fill('1');
  }
  return saveTaste(page, section);
}

async function createGoal(page: Page, webURL: string, projectID: string) {
  await page.goto(`${webURL}/workspace`);
  await page.getByRole('tab', { name: '提案与批准', exact: true }).click();
  await page.getByRole('button', { name: '创建目标', exact: true }).click();
  const dialog = page.getByRole('dialog', { name: '创建目标', exact: true });
  await dialog.getByLabel('目标所属项目', { exact: true }).selectOption(projectID);
  await dialog.getByLabel('目标名称', { exact: true }).fill('核查公开执行协议的判断与委托记录');
  await dialog.getByLabel('目标成功条件', { exact: true }).fill('固定来源和人工反馈在实际服务重启后保持');
  await dialog.getByLabel('目标优先级', { exact: true }).fill('1');
  const response = page.waitForResponse(response => response.url().endsWith('/api/v1/goals') && response.request().method() === 'POST');
  await dialog.getByRole('button', { name: '保存目标', exact: true }).click();
  const saved = await response;
  expect(saved.status(), await saved.text()).toBe(200);
  const result = await saved.json() as S['S2GoalResultV1'];
  if (result.goal.revision < 1) throw new Error('真实目标版本缺失');
  return result.goal;
}

test('完整真实DOM链：Taste三题型、启停、Goal、Snapshot、Policy、Rule、判断反馈及重启', async ({ page }) => {
  test.setTimeout(240_000);
  const { server, api, project, material, revision } = await setup();
  try {
    const session = await api.get('/auth/session') as S['LocalSessionV1'];
    let section = await openTaste(page, server.webURL);
    let taste = await fillTaste(page, section, material.id, revision.revision);
    expect(taste.version).toBe(1);
    expect(taste.configured_by).toBe(session.actor_id);
    expect(taste.rubrics.map(rubric => [rubric.question_type, rubric.options.length, rubric.levels.length])).toEqual([['truth', 0, 0], ['choice', 2, 0], ['score', 0, 3]]);
    expect(taste.examples.map(example => example.expected)).toEqual(['1', 'review', '1']);
    expect(taste.examples.every(example => example.revision === revision.revision && example.content_digest === revision.content_digest)).toBe(true);
    await section.getByText('查看完整已保存配置', { exact: true }).click();
    await expect(section.getByText('人工核查公开执行协议的固定来源 · truth', { exact: true })).toBeVisible();
    for (const enabled of [false, true]) {
      await section.getByRole('button', { name: '编辑配置', exact: true }).click();
      await section.getByLabel('启用状态', { exact: true }).selectOption(String(enabled));
      const next = await saveTaste(page, section);
      expect(next.version).toBe(taste.version + 1);
      expect(next.enabled).toBe(enabled);
      expect(next.rubrics).toEqual(taste.rubrics);
      taste = next;
    }
    expect((await api.get(tastePath) as S['TasteProfileResultV1']).profile).toEqual(taste);
    const goal = await createGoal(page, server.webURL, project.id);
    await page.goto(`${server.webURL}/attention`);
    await page.getByLabel('判断与委托的项目', { exact: true }).selectOption(project.id);
    await page.getByLabel('读取或配置策略的题型', { exact: true }).selectOption('truth');
    await page.getByRole('button', { name: '创建此题型的新策略', exact: true }).click();
    let dialog = page.getByRole('dialog', { name: '判断策略', exact: true });
    await dialog.getByLabel('判断固定目标', { exact: true }).selectOption(`${goal.id}@${goal.revision}`);
    await dialog.getByLabel('判断完整 Taste 来源', { exact: true }).selectOption(`${taste.id}@${taste.version}`);
    await dialog.getByLabel('判断快照文件范围', { exact: true }).fill('STATUS.md\nAGENTS.md');
    const snapshotResponse = page.waitForResponse(response => response.url().endsWith('/judgment-snapshots') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '读取当前真实项目快照', exact: true }).click();
    const captured = await snapshotResponse;
    expect(captured.status(), await captured.text()).toBe(200);
    const { snapshot } = await captured.json() as S['JudgmentSnapshotResultV1'];
    expect(snapshot.project_id).toBe(project.id);
    expect(snapshot.generation).not.toBe('');
    expect(snapshot.root).toBe('');
    await dialog.getByLabel('判断后端', { exact: true }).selectOption('rule');
    await dialog.getByLabel('允许的判断语言', { exact: true }).fill('en');
    const values: [string, string][] = [
      ['判断策略版本标识', 'public-dom-v1'], ['判断模型或规则版本', 'rule:all_evidence_present/v1'],
      ['判断超时上限（毫秒）', '60000'], ['判断输入字节上限', '131072'], ['判断输入 Token 上限', '8192'], ['判断输出 Token 上限', '0'],
      ['目标版本标识', `${goal.id}@${goal.revision}`], ['Taste 版本标识', `${taste.id}@${taste.version}`], ['授权版本标识', String(project.settings.revision)],
    ];
    for (const [label, value] of values) await dialog.getByLabel(label, { exact: true }).fill(value);
    const policyResponse = page.waitForResponse(response => response.url().endsWith('/judgment-policies/truth') && response.request().method() === 'PUT');
    await dialog.getByRole('button', { name: '保存当前题型的判断策略', exact: true }).click();
    const policySaved = await policyResponse;
    expect(policySaved.status(), await policySaved.text()).toBe(200);
    const { policy } = await policySaved.json() as S['JudgmentPolicyResultV1'];
    expect(policy.snapshot_id).toBe(snapshot.generation);
    expect(policy.worktree_id).toBe(snapshot.worktree_id);
    expect(policy.project_settings_revision).toBe(project.settings.revision);
    expect(policy.sources).toEqual({ goal: { id: goal.id, revision: goal.revision }, taste: { id: taste.id, version: taste.version } });
    expect((await api.get(`/local-projects/${project.id}/judgment-policies/truth`) as S['JudgmentPolicyResultV1']).policy).toEqual(policy);
    await page.getByRole('button', { name: '创建未开启的委托规则', exact: true }).click();
    dialog = page.getByRole('dialog', { name: '委托规则编辑', exact: true });
    await dialog.getByLabel('委托工作流标识', { exact: true }).fill('public-protocol-review');
    await dialog.getByLabel('判断题目标识', { exact: true }).fill('public-protocol-evidence');
    await dialog.getByLabel('判断题型', { exact: true }).selectOption('truth');
    await dialog.getByLabel('完整判断题目与要求', { exact: true }).fill('Is the selected fixed public protocol evidence available?');
    await dialog.getByLabel('确定性判断规则', { exact: true }).selectOption('all_evidence_present');
    await dialog.getByRole('button', { name: `使用已核对的策略 r${policy.revision}`, exact: true }).click();
    await dialog.getByLabel('本规则允许的判断语言', { exact: true }).selectOption('en');
    await dialog.getByRole('group', { name: '本规则的固定资料范围', exact: true }).getByRole('checkbox').check();
    await dialog.getByLabel('资料准入数值下限', { exact: true }).fill('1');
    await dialog.getByLabel('资料准入数值上限', { exact: true }).fill('1');
    await dialog.getByLabel('候选元数据外发许可', { exact: true }).selectOption('deny');
    await dialog.getByRole('checkbox', { name: '准入前重新核对固定资料、候选版本和当前权限（必选）', exact: true }).check();
    const ruleResponse = page.waitForResponse(response => /\/delegation-rules\/[^/]+$/.test(response.url()) && response.request().method() === 'PUT');
    await dialog.getByRole('button', { name: '保存为未开启的委托规则', exact: true }).click();
    const ruleSaved = await ruleResponse;
    expect(ruleSaved.status(), await ruleSaved.text()).toBe(200);
    let { delegation_rule: rule } = await ruleSaved.json() as S['DelegationRuleResultV1'];
    expect(rule.state).toBe('disabled');
    expect(rule.owner_id).toBe(session.actor_id);
    expect(rule.materials).toEqual([{ material_id: material.id, revision: revision.revision, digest: revision.content_digest }]);
    dialog = page.getByRole('dialog', { name: '委托规则详情', exact: true });
    for (const state of ['active', 'paused']) {
      await dialog.getByLabel('新的委托状态', { exact: true }).selectOption(state);
      const response = page.waitForResponse(response => response.url().endsWith(`/delegation-rules/${rule.id}/state`) && response.request().method() === 'POST');
      await dialog.getByRole('button', { name: '确认变更当前委托状态', exact: true }).click();
      const changed = await response;
      expect(changed.status(), await changed.text()).toBe(200);
      const next = (await changed.json() as S['DelegationRuleResultV1']).delegation_rule;
      expect(next.revision).toBe(rule.revision + 1);
      expect(next.state).toBe(state);
      rule = next;
      await expect(dialog.getByRole('button', { name: '确认变更当前委托状态', exact: true })).toBeDisabled();
    }
    await dialog.getByRole('button', { name: '关闭', exact: true }).click();
    await page.getByRole('button', { name: '人工发起一次判断', exact: true }).click();
    dialog = page.getByRole('dialog', { name: '人工判断', exact: true });
    await dialog.getByLabel('判断题目标识', { exact: true }).fill('public-protocol-manual');
    await dialog.getByLabel('判断题型', { exact: true }).selectOption('truth');
    await dialog.getByLabel('完整判断题目与要求', { exact: true }).fill('Is the selected fixed public protocol evidence available?');
    await dialog.getByLabel('确定性判断规则', { exact: true }).selectOption('all_evidence_present');
    await dialog.getByRole('button', { name: '确认本次使用该固定策略', exact: true }).click();
    await dialog.getByLabel('本次判断语言', { exact: true }).selectOption('en');
    await dialog.getByRole('group', { name: '本次判断的固定资料', exact: true }).getByRole('checkbox').check();
    const judgmentResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${project.id}/judgments`) && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '执行这一次明确的判断', exact: true }).click();
    const judged = await judgmentResponse;
    expect(judged.status(), await judged.text()).toBe(200);
    const { judgment } = await judged.json() as S['JudgmentResultEnvelopeV1'];
    expect(judgment.decision).toBe('recommend');
    expect(judgment.value).toBe(1);
    expect(judgment.provider).toBe('rule');
    expect(judgment.provider_confidence).toBeNull();
    dialog = page.getByRole('dialog', { name: '判断结果与反馈', exact: true });
    await dialog.getByLabel('判断反馈类型', { exact: true }).selectOption('adopt');
    await dialog.getByLabel('判断反馈理由', { exact: true }).fill('人工核查固定公开协议的规则结果');
    const feedbackResponse = page.waitForResponse(response => response.url().endsWith(`/judgments/${judgment.decision_id}/feedback`) && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '保存该固定版本的人工反馈', exact: true }).click();
    const feedback = await feedbackResponse;
    expect(feedback.status(), await feedback.text()).toBe(200);
    expect(server.query('SELECT COUNT(*) AS count FROM judgment_results')).toEqual([{ count: 1 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM judgment_feedback')).toEqual([{ count: 1 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_taste_profile_revisions')).toEqual([{ count: 3 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
    await server.restart();
    const reopened = await humanAPI(server.apiURL);
    expect((await reopened.get(tastePath) as S['TasteProfileResultV1']).profile).toEqual(taste);
    expect((await reopened.get(`/local-projects/${project.id}/delegation-rules/${rule.id}`) as S['DelegationRuleResultV1']).delegation_rule).toEqual(rule);
    expect((await reopened.get(`/local-projects/${project.id}/judgments/${judgment.decision_id}`) as S['JudgmentResultEnvelopeV1']).judgment).toEqual(judgment);
    expect((await reopened.get(`/local-projects/${project.id}/judgments/${judgment.decision_id}/feedback`) as S['JudgmentFeedbackListV1']).items).toHaveLength(1);
    await page.reload();
    section = page.getByRole('region', { name: 'Taste 配置', exact: true });
    await expect(section.getByText(`当前状态：已配置（版本 ${taste.version}） · 已启用`, { exact: true })).toBeVisible();
  } finally { await server.close(); }
});

test('Taste实际409保留草稿，人工只读核查版本后明确重新准备', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService(true);
  try {
    const section = await openTaste(page, server.webURL);
    await section.getByLabel('启用状态', { exact: true }).selectOption('false');
    const first = await saveTaste(page, section);
    await section.getByRole('button', { name: '编辑配置', exact: true }).click();
    await section.getByRole('button', { name: '添加评判准则', exact: true }).click();
    await section.getByLabel('准则题型', { exact: true }).selectOption('truth');
    await section.getByLabel('评判标准', { exact: true }).fill('保留人工草稿并核查当前来源');
    const second = await api.write(tastePath, { expected_version: first.version, enabled: false, rubrics: [], examples: [], delegations: [] }, { method: 'PUT' }) as S['TasteProfileResultV1'];
    const response = page.waitForResponse(response => response.url().endsWith(`/api/v1${tastePath}`) && response.request().method() === 'PUT');
    await section.getByRole('button', { name: '保存配置', exact: true }).click();
    const conflict = await response;
    expect(conflict.status()).toBe(409);
    expect((await conflict.json() as S['ErrorV1']).error.code).toBe('version_conflict');
    await expect(section.getByLabel('评判标准', { exact: true })).toHaveValue('保留人工草稿并核查当前来源');
    await expect(section.getByRole('button', { name: '保存配置', exact: true })).toBeDisabled();
    await section.getByRole('button', { name: '只读刷新 Taste 配置', exact: true }).click();
    await section.getByRole('button', { name: '读取最新配置以核查原草稿', exact: true }).click();
    const version = second.profile?.version;
    if (!version) throw new Error('实际并发更新缺少版本');
    await section.getByRole('button', { name: `已核查当前配置版本 ${version}`, exact: true }).click();
    await section.getByRole('button', { name: '使用已核查版本重新准备', exact: true }).click();
    const third = await saveTaste(page, section);
    expect(third.version).toBe(version + 1);
    expect(third.rubrics[0].criteria).toBe('保留人工草稿并核查当前来源');
    expect(server.query('SELECT COUNT(*) AS count FROM attention_taste_profile_revisions')).toEqual([{ count: 3 }]);
  } finally { await server.close(); }
});

test('Taste真实浏览器离线unknown跨关闭重开刷新保持原身份且不重发', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService(true);
  try {
    let section = await openTaste(page, server.webURL);
    await section.getByLabel('启用状态', { exact: true }).selectOption('false');
    const writes: string[] = [];
    page.on('request', request => { if (request.method() === 'PUT' && request.url().endsWith(`/api/v1${tastePath}`)) writes.push(request.url()); });
    await context.setOffline(true);
    await section.getByRole('button', { name: '保存配置', exact: true }).click();
    await expect(section.getByRole('alert').filter({ hasText: '当前接口没有按请求身份读取收据' })).toBeVisible();
    await expect(section.getByRole('button', { name: '保存配置', exact: true })).toBeDisabled();
    await expect(section.getByRole('button', { name: '取消编辑', exact: true })).toBeDisabled();
    const storageKey = 'astrocyte:command:pending:attention:taste';
    const original = await page.evaluate(key => sessionStorage.getItem(key), storageKey);
    expect(original).toMatch(/^[0-9a-f-]{36}$/);
    // Node HTTP 未断网，用实际库核查浏览器未送达的命令没有产生记录。
    expect((await api.get(tastePath) as S['TasteProfileResultV1']).configured).toBe(false);
    expect(server.query("SELECT COUNT(*) AS count FROM attention_receipts WHERE command='UpdateTasteProfile'")).toEqual([{ count: 0 }]);
    await context.setOffline(false);
    await page.goto(`${server.webURL}/workspace`);
    await page.goto(`${server.webURL}/attention`);
    section = page.getByRole('region', { name: 'Taste 配置', exact: true });
    await expect(section.getByRole('button', { name: '创建配置', exact: true })).toBeDisabled();
    await section.getByRole('button', { name: '只读刷新 Taste 配置', exact: true }).click();
    await expect(section.getByText('当前状态：未配置', { exact: true })).toBeVisible();
    await page.reload();
    await expect(section.getByRole('button', { name: '创建配置', exact: true })).toBeDisabled();
    expect(await page.evaluate(key => sessionStorage.getItem(key), storageKey)).toBe(original);
    expect(writes).toHaveLength(0);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_taste_profile_revisions')).toEqual([{ count: 0 }]);
  } finally { await context.setOffline(false); await server.close(); }
});

test('Taste真实403读取过期保留数据并禁写，恢复后人工重新读取', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server } = await realService(true);
  try {
    const section = await openTaste(page, server.webURL);
    await section.getByLabel('启用状态', { exact: true }).selectOption('false');
    const profile = await saveTaste(page, section);
    await section.getByRole('button', { name: '编辑配置', exact: true }).click();
    await context.setExtraHTTPHeaders({ Authorization: 'Bearer invalid' });
    const response = page.waitForResponse(response => response.url().endsWith(`/api/v1${tastePath}`) && response.request().method() === 'GET');
    await section.getByRole('button', { name: '只读刷新 Taste 配置', exact: true }).click();
    expect((await response).status()).toBe(403);
    await expect(section.getByRole('alert').filter({ hasText: '数据可能已过期' })).toBeVisible();
    await expect(section.getByText(`当前状态：已配置（版本 ${profile.version}） · 已停用`, { exact: true })).toBeVisible();
    await expect(section.getByRole('button', { name: '保存配置', exact: true })).toBeDisabled();
    await context.setExtraHTTPHeaders({});
    await section.getByRole('button', { name: '只读刷新 Taste 配置', exact: true }).click();
    await expect(section.getByRole('button', { name: '保存配置', exact: true })).toBeEnabled();
    expect(server.query('SELECT COUNT(*) AS count FROM attention_taste_profile_revisions')).toEqual([{ count: 1 }]);
  } finally { await context.setExtraHTTPHeaders({}); await server.close(); }
});

test('Taste选择实际较早固定版本，换材料与移除示例后保持草稿，撤回来源阻止保存', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService(true);
  try {
    const commits = execFileSync('git', ['log', '-n', '2', '--format=%H', '--', 'AGENTS.md'], { cwd: repositoryRoot, encoding: 'utf8' }).trim().split(/\r?\n/).reverse();
    if (commits.length !== 2) throw new Error('实际公开协议缺少两份历史版本');
    const originals = commits.map(commit => ({ commit, text: execFileSync('git', ['show', `${commit}:AGENTS.md`], { cwd: repositoryRoot, encoding: 'utf8' }) }));
    if (originals[0].text === originals[1].text) throw new Error('实际公开协议的两份历史正文相同');
    const versions = [];
    for (const original of originals) {
      versions.push(await importMaterial(api, {
        kind: 'text', adapter: 'manual', source_key: '', content_digest: '',
        source_locator: 'https://github.com/songconmaisaix31-design/Astrocyte/blob/main/AGENTS.md',
        export_text: original.text, title: '实际公开执行协议历史', collection_reason: null,
        source_spans: [`https://github.com/songconmaisaix31-design/Astrocyte/blob/${original.commit}/AGENTS.md · whole document`],
      }));
    }
    const first = versions[0].detail as S['MaterialDetailV1'];
    const second = versions[1].detail as S['MaterialDetailV1'];
    expect(first.material.id).toBe(second.material.id);
    expect(second.material.current_revision).toBe(2);
    const fixed = second.revisions.find(revision => revision.revision === 1);
    if (!fixed) throw new Error('实际旧版本缺失');
    const other = await importTrackedDocument(api, 'STATUS.md');
    const section = await openTaste(page, server.webURL);
    const profile = await fillTaste(page, section, second.material.id, fixed.revision);
    expect(profile.examples.every(example => example.revision === 1 && example.content_digest === fixed.content_digest)).toBe(true);
    await section.getByRole('button', { name: '编辑配置', exact: true }).click();
    const example = section.locator('.ac-taste-example').first();
    await example.getByLabel('示例资料', { exact: true }).selectOption(other.detail.material.id);
    await example.getByLabel('示例资料', { exact: true }).selectOption(second.material.id);
    await expect(example.getByLabel('示例固定版本', { exact: true })).toBeEnabled();
    await expect(example.getByLabel('示例固定版本', { exact: true })).toHaveValue('');
    await example.getByLabel('示例固定版本', { exact: true }).selectOption('1');
    await example.getByRole('button', { name: '移除示例', exact: true }).click();
    await expect(section.locator('.ac-taste-example')).toHaveCount(2);
    const updated = await saveTaste(page, section);
    expect(updated.examples).toEqual(profile.examples.slice(1));
    await section.getByRole('button', { name: '编辑配置', exact: true }).click();
    const current = await api.get(`/materials/${second.material.id}`) as S['MaterialDetailV1'];
    await api.write(`/materials/${second.material.id}`, { expected_version: current.material.version, lifecycle: 'withdrawn' }, { method: 'PATCH' });
    const writes: string[] = [];
    page.on('request', request => { if (request.method() === 'PUT' && request.url().endsWith(`/api/v1${tastePath}`)) writes.push(request.url()); });
    await section.getByRole('button', { name: '保存配置', exact: true }).click();
    await expect(section.getByRole('alert').filter({ hasText: '示例资料当前未处于 active 状态' })).toBeVisible();
    expect(writes).toHaveLength(0);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_taste_profile_revisions')).toEqual([{ count: 2 }]);
  } finally { await server.close(); }
});
