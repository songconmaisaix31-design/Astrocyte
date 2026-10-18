import { expect, test } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';
import { readFile, readdir, realpath, rename, writeFile } from 'node:fs/promises';
import { join, sep } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, importMaterial } from '../../tests/s1/api.mjs';
import { registerCurrentProject, repositoryRoot, writeWorks } from './realService';
import type { components } from '../src/api/schema';

type S = components['schemas'];
type API = Awaited<ReturnType<typeof humanAPI>>;

async function importFixedREADME(api: API) {
  const sourceCommit = '657ca1464517a171ed50ddc8fcc787b237c44ad9';
  const text = execFileSync('git', ['show', `${sourceCommit}:README.md`], { cwd: repositoryRoot, encoding: 'utf8' });
  return importMaterial(api, { kind: 'text', adapter: 'manual', source_key: '', content_digest: '', source_locator: `https://github.com/songconmaisaix31-design/Astrocyte/blob/${sourceCommit}/README.md`, export_text: text, title: 'README.md', collection_reason: null, source_spans: ['whole document'] });
}

async function actualMission(api: API, project: S['LocalProjectV1'], idleSeconds: number) {
  const { goal } = await api.write('/goals', { expected_version: 0, goal: { project_id: project.id, title: '核查公开项目动态任务DAG', success_criteria: ['任务与来源、真实停止记录可回查'], priority: 1, status: 'active' } }) as S['S2GoalResultV1'];
  const content: S['S2ProposalContentV1'] = {
    title: '人工发起动态DAG核查', goal: { id: goal.id, revision: goal.revision }, inputs: [],
    scope: { project_id: project.id, settings_revision: project.settings.revision, roots: [project.root], write_paths: ['internal/swarm'], actions: ['create_mission', 'manage_work_items', 'read_swarm', 'recover_mission'], tools: [], external_targets: [], max_concurrency: 1 },
    budget: { calls: 3, tokens: null, money_minor: null, currency: '' }, deliverables: ['公开项目实际任务和事件'], success_criteria: ['保留真实依赖与停止历史'],
    stops: { deadline: new Date(Date.now() + 3600_000).toISOString(), conditions: ['达到批准的无进展时间后停止'], max_work_items: 4, max_attempts: 2, no_progress_seconds: idleSeconds, max_no_progress_rounds: 3 },
  };
  const { proposal } = await api.write('/proposals', { expected_version: 0, content }) as S['S2ProposalResultV1'];
  const { proposal: review } = await api.write(`/proposals/${proposal.id}/submit`, { expected_version: proposal.version }) as S['S2ProposalResultV1'];
  const { approval } = await api.write(`/proposals/${proposal.id}/approvals`, { expected_version: review.version, expected_revision: review.revision, expires_at: content.stops.deadline }) as S['S2ApprovalResultV1'];
  const { mission } = await writeWorks(api, '/missions', { expected_version: 0, grant_id: approval.id, epoch: approval.epoch, proposal_revision: approval.proposal_revision, project_id: project.id, root: project.root }) as S['SwarmMissionEnvelopeV1'];
  return mission;
}

test('真实依赖增长、事件回查、键盘、断线重连、停止历史和撤销清除', async ({ page }) => {
  test.setTimeout(180_000);
  const server = await startS1Server({ env: { ASTROCYTE_WEB_DIR: fileURLToPath(new URL('../dist/', import.meta.url)), OPENVIKING_BASE_URL: '', OPENVIKING_API_KEY: '', OPENVIKING_ACCOUNT: '', OPENVIKING_USER: '', OPENVIKING_ROOT_URI: '' } });
  try {
    let api = await humanAPI(server.apiURL);
    const project = await registerCurrentProject(api);
    const { project: permitted } = await api.write(`/local-projects/${project.id}/settings`, { expected_revision: project.settings.revision, settings: { allow_directory: true, allowed_subdirs: ['.'], allowed_actions: ['create_mission', 'manage_work_items', 'read_swarm', 'recover_mission'], allowed_tools: [], external_model_cli: '', allow_agent_control: false } }, { method: 'PUT' }) as S['LocalProjectResultV1'];
    await importFixedREADME(api);
    const mission = await actualMission(api, permitted, 3600);
    const errors: string[] = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto(`${server.apiURL}/swarm`);
    const region = page.getByRole('region', { name: '动态任务依赖', exact: true });
    await region.getByLabel('DAG任务', { exact: true }).selectOption(mission.id);
    await expect(region.getByText('当前任务尚无工作项，依赖图为空。')).toBeVisible();
    const firstTitle = '读取公开SPEC并核查动态任务依赖、创建主体与原始事件来源';
    const { work_item: first } = await writeWorks(api, `/missions/${mission.id}/work-items`, { expected_version: mission.version, title: firstTitle, inputs: mission.grant.inputs, dependencies: [], write_paths: ['internal/swarm'] }) as { work_item: S['SwarmWorkItemV1'] };
    const firstButton = region.getByRole('button', { name: `${firstTitle} 开放`, exact: true });
    await expect(firstButton).toBeVisible({ timeout: 12_000 });
    await firstButton.focus();
    await page.keyboard.press('Enter');
    const details = region.getByRole('region', { name: '依赖节点详情', exact: true });
    await expect(details).toContainText(first.id);
    await expect(details).toContainText('人工 local-human');
    const { work_item: second } = await writeWorks(api, `/missions/${mission.id}/work-items`, { expected_version: mission.version + 1, title: '逐项回查真实事件并记录本轮验证结果', inputs: mission.grant.inputs, dependencies: [first.id], write_paths: ['internal/swarm'] }) as { work_item: S['SwarmWorkItemV1'] };
    await expect(region.getByRole('status').filter({ hasText: '工作项 2 · 依赖 1' })).toBeVisible({ timeout: 12_000 });
    await expect(details).toContainText(first.id);
    await region.getByLabel('依赖布局方向', { exact: true }).selectOption('TB');
    await region.getByLabel('依赖布局缩放', { exact: true }).selectOption('0.8');
    await expect(firstButton).toBeVisible();
    const started = performance.now();
    const snapshot = await api.get(`/missions/${mission.id}/dag`) as S['SwarmMissionDAGV1'];
    await writeFile(join(server.temporary, 'dag-before-disconnect.json'), JSON.stringify(snapshot, null, 2), { flag: 'wx' });
    console.log(JSON.stringify({ measurement: 'actual_dag_get', nodes: snapshot.nodes.length, edges: snapshot.edges.length, elapsed_ms: performance.now() - started }));
    expect(snapshot.workers).toEqual([]);
    expect(snapshot.edges).toHaveLength(1);
    expect(snapshot.edges[0]).toMatchObject({ from: first.id, to: second.id, required: true });
    for (const node of snapshot.nodes) {
      expect(node.work_item.creator_kind).toBe('human');
      expect(node.creation_event).not.toBeNull();
      if (!node.creation_event) throw new Error('真实节点缺少创建事件');
      expect(server.query('SELECT event_id FROM foundation_outbox WHERE event_id=? AND aggregate_id=?', node.creation_event.event_id, node.work_item.id)).toHaveLength(1);
    }
    expect(server.query('SELECT event_id FROM foundation_outbox WHERE event_id=?', snapshot.edges[0].event_id)).toHaveLength(1);
    for (const event of snapshot.events) expect(server.query('SELECT event_id FROM foundation_outbox WHERE event_id=? AND event_type=? AND aggregate_id=?', event.event_id, event.type, event.aggregate_id)).toHaveLength(1);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);

    await server.stopAPI();
    await expect(region.getByRole('alert').filter({ hasText: '数据可能已过期' })).toBeVisible({ timeout: 15_000 });
    await expect(firstButton).toBeVisible();
    await server.startAPI();
    api = await humanAPI(server.apiURL);
    await expect(region.getByRole('alert').filter({ hasText: '数据可能已过期' })).toHaveCount(0, { timeout: 15_000 });
    await expect(region.getByRole('status').filter({ hasText: '工作项 2 · 依赖 1' })).toBeVisible();
    expect((await api.get(`/missions/${mission.id}/dag`) as S['SwarmMissionDAGV1']).nodes.map(node => node.work_item.id).sort()).toEqual([first.id, second.id].sort());

    const stopped = await actualMission(api, permitted, 1);
    await region.getByRole('button', { name: '刷新任务依赖', exact: true }).click();
    await region.getByLabel('DAG任务', { exact: true }).selectOption(stopped.id);
    await expect(region).toContainText('no_progress_time_limit', { timeout: 15_000 });
    const stopSnapshot = await api.get(`/missions/${stopped.id}/dag`) as S['SwarmMissionDAGV1'];
    await writeFile(join(server.temporary, 'dag-actual-idle-stop.json'), JSON.stringify(stopSnapshot, null, 2), { flag: 'wx' });
    expect(stopSnapshot.mission_state).toBe('blocked');
    expect(stopSnapshot.workers).toEqual([]);
    expect(stopSnapshot.events.filter(event => event.type === 'swarm.mission_stop_requested')).toHaveLength(1);
    const stopEvent = stopSnapshot.events.find(event => event.type === 'swarm.mission_stop_requested');
    if (!stopEvent) throw new Error('实际停止事件缺失');
    expect(server.query('SELECT event_id FROM foundation_outbox WHERE event_id=? AND json_extract(payload,\'$.source\')=?', stopEvent.event_id, 'system:swarm-lifecycle')).toHaveLength(1);
    for (const event of stopSnapshot.events) expect(server.query('SELECT event_id FROM foundation_outbox WHERE event_id=? AND event_type=? AND aggregate_id=?', event.event_id, event.type, event.aggregate_id)).toHaveLength(1);

    await region.getByLabel('DAG任务', { exact: true }).selectOption(mission.id);
    await expect(firstButton).toBeVisible();
    await api.write(`/local-projects/${project.id}/settings`, { expected_revision: permitted.settings.revision, settings: { allow_directory: true, allowed_subdirs: ['.'], allowed_actions: ['read_context'], allowed_tools: [], external_model_cli: '', allow_agent_control: false } }, { method: 'PUT' });
    await expect(region.getByRole('region', { name: '工作项依赖画布', exact: true })).toHaveCount(0, { timeout: 12_000 });
    await expect(region).toContainText('API 403');
    let subsequentDAGRequests = 0;
    page.on('request', request => { if (new URL(request.url()).pathname.endsWith('/dag')) subsequentDAGRequests++; });
    await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '共同工作区', exact: true }).click();
    await expect(page).toHaveURL(/\/workspace$/);
    const countAfterNavigation = subsequentDAGRequests;
    await page.waitForTimeout(5500);
    expect(subsequentDAGRequests).toBe(countAfterNavigation);
    expect(errors).toEqual([]);
  } finally { await server.close({ preserveData: true }); }
});

test('任务依赖按需加载，真实分块缺失时局部报错并由人工重新加载恢复', async ({ page }) => {
  test.setTimeout(120_000);
  const dist = await realpath(fileURLToPath(new URL('../dist/', import.meta.url)));
  expect(dist).toBe(join(await realpath(repositoryRoot), 'web', 'dist'));
  const assets = await realpath(join(dist, 'assets'));
  expect(assets.startsWith(`${dist}${sep}`)).toBe(true);
  const entries = await readdir(assets);
  const chunks = entries.filter(name => name.startsWith('MissionDAGView-') && name.endsWith('.js'));
  expect(chunks).toHaveLength(1);
  const chunk = await realpath(join(assets, chunks[0]));
  const held = join(assets, `${chunks[0]}.verification-held`);
  expect(chunk.startsWith(`${assets}${sep}`)).toBe(true);
  expect(held.startsWith(`${assets}${sep}`)).toBe(true);
  expect(entries).not.toContain(`${chunks[0]}.verification-held`);
  const server = await startS1Server({ env: { ASTROCYTE_WEB_DIR: dist, OPENVIKING_BASE_URL: '', OPENVIKING_API_KEY: '', OPENVIKING_ACCOUNT: '', OPENVIKING_USER: '', OPENVIKING_ROOT_URI: '' } });
  let moved = false;
  try {
    const api = await humanAPI(server.apiURL);
    await importFixedREADME(api);
    const index = await readFile(join(dist, 'index.html'));
    const routes: { path: string; status: number; content_type: string }[] = [];
    for (const path of ['/assets', '/assets/', '/cognition/evidence']) {
      const route = await page.request.get(`${server.apiURL}${path}`);
      expect(route.status()).toBe(200);
      expect(route.headers()['content-type']).toMatch(/^text\/html/);
      expect(await route.body()).toEqual(index);
      routes.push({ path, status: route.status(), content_type: route.headers()['content-type'] });
    }
    const health = await page.request.get(`${server.apiURL}/api/v1/health`);
    expect(health.status()).toBe(200);
    expect(health.headers()['content-type']).toMatch(/^application\/json/);
    expect(await health.json()).toMatchObject({ status: 'ok' });
    const unknownAPI = await page.request.get(`${server.apiURL}/api/v1/dag-loading-missing`, { headers: { Cookie: api.cookie } });
    expect(unknownAPI.status()).toBe(404);
    expect(unknownAPI.headers()['content-type']).toMatch(/^application\/json/);
    await writeFile(join(server.temporary, 'actual-static-routes.json'), JSON.stringify({ routes, api: [{ path: '/api/v1/health', status: health.status() }, { path: '/api/v1/dag-loading-missing', status: unknownAPI.status() }] }, null, 2), { flag: 'wx' });
    const requests: string[] = [];
    const failedRequests: { path: string; error: string | null }[] = [];
    page.on('request', request => { if (new URL(request.url()).pathname.endsWith('.js')) requests.push(new URL(request.url()).pathname); });
    page.on('requestfailed', request => failedRequests.push({ path: new URL(request.url()).pathname, error: request.failure()?.errorText ?? null }));
    await rename(chunk, held);
    moved = true;
    try {
      await page.goto(`${server.apiURL}/attention`);
      await expect(page.locator('h1')).toHaveText('资料沉淀');
      await expect(page.locator('main li[role="button"]').filter({ hasText: 'README.md' })).toBeVisible();
      await page.waitForLoadState('networkidle');
      expect(requests.some(path => path.endsWith(chunks[0]))).toBe(false);
      const missingResponse = page.waitForResponse(response => new URL(response.url()).pathname === `/assets/${chunks[0]}`);
      await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '蜂群空间', exact: true }).click();
      const response = await missingResponse;
      expect(response.status()).toBe(404);
      expect(response.headers()['cache-control']).toBe('no-store');
      expect(response.headers()['content-type']).toMatch(/^text\/plain/);
      const failure = page.getByRole('region', { name: '任务依赖加载错误', exact: true });
      await expect(failure.getByRole('alert')).toContainText('任务依赖视图无法显示：');
      await expect(failure.getByRole('alert')).toContainText(chunks[0]);
      await expect(failure.getByRole('button', { name: '重新加载页面', exact: true })).toBeVisible();
      await expect(page.getByRole('region', { name: '动态任务依赖', exact: true })).toHaveCount(0);
      await expect(page.locator('h1')).toHaveText('蜂群空间');
      await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '共同工作区', exact: true }).click();
      await expect(page.locator('h1')).toHaveText('共同工作区');
      await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '蜂群空间', exact: true }).click();
      await expect(failure).toBeVisible();
      const countBeforeWait = requests.filter(path => path.endsWith(chunks[0])).length;
      await page.waitForTimeout(5500);
      expect(requests.filter(path => path.endsWith(chunks[0]))).toHaveLength(countBeforeWait);
      await writeFile(join(server.temporary, 'actual-missing-dag-chunk.json'), JSON.stringify({ chunk, requests, failedRequests, response: { status: response.status(), content_type: response.headers()['content-type'], cache_control: response.headers()['cache-control'], body_bytes: (await response.body()).length }, actual_error: await failure.getByRole('alert').textContent() }, null, 2), { flag: 'wx' });
    } finally {
      await rename(held, chunk);
      moved = false;
    }
    const reload = page.getByRole('button', { name: '重新加载页面', exact: true });
    await reload.focus();
    const recoveredResponse = page.waitForResponse(response => new URL(response.url()).pathname === `/assets/${chunks[0]}`);
    await Promise.all([page.waitForEvent('load'), page.keyboard.press('Enter')]);
    const recovered = await recoveredResponse;
    expect(recovered.status()).toBe(200);
    expect(recovered.headers()['content-type']).toMatch(/javascript/);
    const recoveredBytes = (await recovered.body()).length;
    expect(recoveredBytes).toBeGreaterThan(0);
    await expect(page.getByRole('region', { name: '动态任务依赖', exact: true })).toBeVisible();
    await expect(page.getByRole('region', { name: '任务依赖加载错误', exact: true })).toHaveCount(0);
    await writeFile(join(server.temporary, 'actual-recovered-dag-chunk.json'), JSON.stringify({ chunk, status: recovered.status(), content_type: recovered.headers()['content-type'], body_bytes: recoveredBytes }, null, 2), { flag: 'wx' });
    expect(await realpath(chunk)).toBe(chunk);
  } finally {
    try { if (moved) await rename(held, chunk); }
    finally { await server.close({ preserveData: true }); }
  }
});
