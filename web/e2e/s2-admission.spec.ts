import { test, expect, type Page } from '@playwright/test';
import { humanAPI } from '../../tests/s1/api.mjs';
import { realService, registerCurrentProject, trackedCandidate } from './realService';
import type { AdmissionRequest, AdmissionResult } from '../src/api/admissions';
import type { components } from '../src/api/schema';

async function openCandidate(page: Page, url: string) {
  await page.goto(`${url}/attention`);
  await page.locator('main [role="button"]').filter({ hasText: '前端标准整改核查' }).click();
  const dialog = page.getByRole('dialog', { name: '机会详情', exact: true });
  await expect(dialog.getByRole('region', { name: '资料人工准入', exact: true })).toBeVisible();
  return dialog;
}

test('实际准入固定候选与项目资料，原事件和草稿在撤销及服务重启后保留', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    const candidate = await trackedCandidate(api);
    const ref = candidate.evidence_refs[0];
    await api.write(`/project-spaces/${project.space_id}/references`, { material_id: ref.material_id, revision: ref.revision });
    let dialog = await openCandidate(page, server.webURL);
    const form = dialog.getByRole('region', { name: '资料人工准入', exact: true });
    await expect(form.getByRole('button', { name: '准备人工准入请求', exact: true })).toBeDisabled();
    await form.getByLabel('准入后的项目归属', { exact: true }).selectOption(project.id);
    await expect(form.getByText('本候选全部固定版本已在项目空间主动引用', { exact: true })).toBeVisible();
    await form.getByRole('button', { name: '准备人工准入请求', exact: true }).click();
    const submit = form.getByRole('button', { name: '确认资料准入并生成草稿', exact: true });
    await expect(submit).toBeDisabled();
    await form.getByRole('checkbox', { name: '我已核对本次固定候选、资料版本和项目归属，确认人工准入', exact: true }).check();
    const expectedBody: AdmissionRequest = { schema_version: 1, request_id: expect.any(String), expected_version: candidate.version, expected_revision: candidate.revision, project_id: project.id };
    const writes: string[] = [];
    page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') writes.push(new URL(request.url()).pathname); });
    const admittedHTTP = page.waitForResponse(response => response.request().method() === 'POST' && response.url().endsWith(`/opportunities/${candidate.id}/admissions`));
    await submit.click();
    const response = await admittedHTTP;
    expect(response.status()).toBe(201);
    expect(response.request().postDataJSON()).toEqual(expectedBody);
    expect(response.request().headers()['idempotency-key']).toEqual(expect.any(String));
    const saved = await response.json() as AdmissionResult;
    expect(saved.admission).toMatchObject({ opportunity_id: candidate.id, opportunity_revision: candidate.revision, opportunity_version: candidate.version, project_id: project.id, authority: 'human', status: 'active', version: 1, inputs: [{ kind: 'material', id: ref.material_id, revision: ref.revision }] });
    expect(saved.draft?.id).toBeTruthy();
    expect(server.query('SELECT data FROM attention_admissions WHERE id=?', saved.admission.id).map(row => {
      if (typeof row.data !== 'string') throw new Error('真实准入数据列没有返回 JSON 文本');
      return JSON.parse(row.data);
    })).toEqual([saved.admission]);
    expect(server.query('SELECT id,type,correlation_id FROM attention_outbox WHERE id=?', saved.admission.admitted_event_id)).toEqual([{ id: saved.admission.admitted_event_id, type: 'attention.opportunity_admitted', correlation_id: expectedBody.request_id }]);
    expect(server.query("SELECT id,json_extract(data,'$.status') AS status,json_extract(data,'$.content.scope.project_id') AS project_id FROM workspace_proposals WHERE id=?", saved.draft!.id)).toEqual([{ id: saved.draft!.id, status: 'draft', project_id: project.id }]);
    await expect(form.getByRole('region', { name: '已保存准入记录', exact: true })).toContainText(saved.admission.id);
    await expect(submit).toBeDisabled();
    await form.getByLabel('撤销准入理由', { exact: true }).fill('隔离功能验证已结束，撤销本次资料准入。');
    const revokedHTTP = page.waitForResponse(result => result.request().method() === 'POST' && result.url().endsWith(`/admissions/${saved.admission.id}/revocations`));
    await form.getByRole('button', { name: '确认撤销资料准入', exact: true }).click();
    expect((await revokedHTTP).status()).toBe(200);
    const revoked = await api.get(`/admissions/${saved.admission.id}`) as AdmissionResult;
    expect(revoked.admission).toMatchObject({ ...saved.admission, version: 2, status: 'revoked', updated_at: expect.any(String), revoked_event_id: expect.any(String), reason: '隔离功能验证已结束，撤销本次资料准入。' });
    expect(server.query('SELECT id,type FROM attention_outbox WHERE id=?', revoked.admission.revoked_event_id!)).toEqual([{ id: revoked.admission.revoked_event_id, type: 'attention.admission_revoked' }]);
    expect(server.query('SELECT id FROM workspace_proposals')).toEqual([{ id: saved.draft!.id }]);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_approvals')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
    const currentProjects = await api.get('/local-projects') as components['schemas']['LocalProjectListV1'];
    expect(currentProjects.items[0].settings).toEqual(project.settings);
    await server.restart();
    dialog = await openCandidate(page, server.webURL);
    await dialog.getByLabel('准入记录 ID', { exact: true }).fill(saved.admission.id);
    await dialog.getByRole('button', { name: '读取准入记录', exact: true }).click();
    await expect(dialog.getByRole('region', { name: '已保存准入记录', exact: true })).toContainText(revoked.admission.revoked_event_id!);
    expect(((await (await humanAPI(server.apiURL)).get(`/admissions/${saved.admission.id}`)) as AdmissionResult).admission).toEqual(revoked.admission);
    expect(writes).toEqual([`/api/v1/opportunities/${candidate.id}/admissions`, `/api/v1/admissions/${saved.admission.id}/revocations`]);
  } finally { await server.close({ preserveData: true }); }
});

test('实际候选版本冲突保留原请求和确认输入，错误不产生准入或草稿', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const candidate = await trackedCandidate(api);
    const dialog = await openCandidate(page, server.webURL);
    const form = dialog.getByRole('region', { name: '资料人工准入', exact: true });
    await form.getByLabel('准入后的项目归属', { exact: true }).selectOption('unassigned');
    await form.getByRole('button', { name: '准备人工准入请求', exact: true }).click();
    await form.getByRole('checkbox').check();
    await api.write(`/opportunities/${candidate.id}/reviews`, { expected_version: candidate.version, feedback: 'later', reason: '实际候选已由另一人工操作暂缓。' }, { status: 201 });
    const failedHTTP = page.waitForResponse(response => response.request().method() === 'POST' && response.url().endsWith(`/opportunities/${candidate.id}/admissions`));
    await form.getByRole('button', { name: '确认资料准入并生成草稿', exact: true }).click();
    expect((await failedHTTP).status()).toBe(409);
    await expect(form.getByRole('alert')).toContainText('409');
    await expect(form.getByRole('checkbox')).toBeChecked();
    await dialog.getByRole('button', { name: '刷新候选详情', exact: true }).click();
    await expect(form.getByRole('button', { name: '确认资料准入并生成草稿', exact: true })).toBeDisabled();
    expect(server.query('SELECT COUNT(*) AS count FROM attention_admissions')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_proposals')).toEqual([{ count: 0 }]);
  } finally { await server.close({ preserveData: true }); }
});

test('实际项目读取拒绝后禁用准入，恢复权限重新核查固定资料', async ({ page, context }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    const candidate = await trackedCandidate(api);
    const ref = candidate.evidence_refs[0];
    const dialog = await openCandidate(page, server.webURL);
    const form = dialog.getByRole('region', { name: '资料人工准入', exact: true });
    await form.getByLabel('准入后的项目归属', { exact: true }).selectOption(project.id);
    await expect(form.getByText('项目空间缺少本候选的固定资料版本，请先主动引用', { exact: true })).toBeVisible();
    await expect(form.getByRole('button', { name: '准备人工准入请求', exact: true })).toBeDisabled();
    await api.write(`/project-spaces/${project.space_id}/references`, { material_id: ref.material_id, revision: ref.revision });
    await form.getByRole('button', { name: '刷新准入项目权限', exact: true }).click();
    await expect(form.getByRole('button', { name: '准备人工准入请求', exact: true })).toBeEnabled();
    await context.setExtraHTTPHeaders({ Authorization: 'Bearer invalid' });
    const denied = page.waitForResponse(response => response.url().endsWith('/local-projects'));
    await form.getByRole('button', { name: '刷新准入项目权限', exact: true }).click();
    expect((await denied).status()).toBe(403);
    await expect(form.getByRole('button', { name: '准备人工准入请求', exact: true })).toBeDisabled();
    await expect(form.getByRole('alert').first()).toContainText('过期');
    await context.setExtraHTTPHeaders({});
    await form.getByRole('button', { name: '刷新准入项目权限', exact: true }).click();
    await expect(form.getByRole('button', { name: '准备人工准入请求', exact: true })).toBeEnabled();
    expect(server.query('SELECT COUNT(*) AS count FROM attention_admissions')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  } finally { await context.setExtraHTTPHeaders({}); await server.close({ preserveData: true }); }
});

test('真实离线准入保留未知请求锁，刷新和重新打开不重发且锁定候选编辑', async ({ page, context }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    await trackedCandidate(api);
    let dialog = await openCandidate(page, server.webURL);
    let form = dialog.getByRole('region', { name: '资料人工准入', exact: true });
    await form.getByLabel('准入后的项目归属', { exact: true }).selectOption('unassigned');
    await form.getByRole('button', { name: '准备人工准入请求', exact: true }).click();
    await form.getByRole('checkbox').check();
    const writes: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && new URL(request.url()).pathname.startsWith('/api/v1/')) writes.push(request.url()); });
    await context.setOffline(true);
    await form.getByRole('button', { name: '确认资料准入并生成草稿', exact: true }).click();
    const lockAlert = form.getByRole('alert').filter({ hasText: '关闭或刷新后保持锁定' });
    await expect(lockAlert).toBeVisible();
    const lockRequestID = (await lockAlert.textContent() ?? '').match(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/)?.[0] ?? '';
    expect(lockRequestID).toMatch(/^[0-9a-f-]{36}$/);
    await expect(lockAlert).toContainText(lockRequestID);
    await expect(dialog.getByRole('button', { name: '编辑候选', exact: true })).toBeDisabled();
    await expect(dialog.getByRole('button', { name: '保存人工反馈', exact: true })).toBeDisabled();
    await context.setOffline(false);
    await dialog.getByRole('button', { name: '刷新候选详情', exact: true }).click();
    await expect(form.getByRole('button', { name: '准备人工准入请求', exact: true })).toBeDisabled();
    await page.reload();
    dialog = await openCandidate(page, server.webURL);
    form = dialog.getByRole('region', { name: '资料人工准入', exact: true });
    await expect(form.getByRole('alert').filter({ hasText: '关闭或刷新后保持锁定' })).toContainText(lockRequestID);
    await expect(form.getByRole('button', { name: '准备人工准入请求', exact: true })).toBeDisabled();
    await expect(dialog.getByRole('button', { name: '编辑候选', exact: true })).toBeDisabled();
    expect(writes).toEqual([]);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_admissions')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_proposals')).toEqual([{ count: 0 }]);
  } finally { await context.setOffline(false); await server.close({ preserveData: true }); }
});
