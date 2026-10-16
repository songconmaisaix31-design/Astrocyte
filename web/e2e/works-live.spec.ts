/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const rootKey = (root: string) => root[1] === ':' ? root.split(String.fromCharCode(92)).join('/').toLowerCase() : root;
const rawText = (value: unknown) => {
  if (!(value instanceof Uint8Array)) throw new Error('Works raw observation must be a SQLite BLOB');
  return Buffer.from(value).toString('utf8');
};

test('actual approved works UI commands persist in SQLite after owned API restart', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_W7_LIVE !== '1', 'Requires an explicit approved metadata root and isolated browser slot.');
  const approvedRoot = process.env.ASTROCYTE_TEST_W7_APPROVED_ROOT;
  expect(approvedRoot).toBeTruthy();
  test.setTimeout(600_000);
  page.setDefaultTimeout(30_000);
  const server = await startS1Server({ browser: true, apiPort: 18847, webPort: 15247 });
  let api = await humanAPI(server.apiURL);
  let primaryError: unknown;
  const errors: string[] = [];
  const commands: { path: string; key: string; body: Record<string, unknown> }[] = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    if (request.method() === 'GET' || !new URL(request.url()).pathname.startsWith('/api/v1/works')) return;
    const body = request.postDataJSON() as Record<string, unknown>;
    const headers = request.headers();
    commands.push({ path: new URL(request.url()).pathname, key: headers['idempotency-key'], body });
    expect(headers['x-csrf-token']).toBeTruthy();
    expect(body.request_id).toBe(headers['idempotency-key']);
    expect(body.schema_version).toBe(1);
    expect(body.expected_version).toBe(1);
  });
  const read = async () => ((await api.get('/works')) as S['WorksSnapshotResultV1']).snapshot;
  console.log(JSON.stringify({ scope: 'task_live_works', directory: server.temporary, approved_root: approvedRoot, api: server.apiURL, web: server.webURL }));
  try {
    await expect.poll(async () => ((await api.get('/local-projects/registered')) as S['RegisteredProjectDiscoveryResultV1']).snapshot.projects.some(project => rootKey(project.root) === rootKey(resolve(approvedRoot!))), { timeout: 130_000 }).toBe(true);
    const initial = await read();
    expect(initial.status).toBe('unknown');
    expect(initial.observed_at).toBeNull();
    await page.goto(`${server.webURL}/workspace`);
    const album = page.getByRole('region', { name: '作品寻回册', exact: true });
    await expect(album.getByRole('button', { name: '扫描导入项目', exact: true })).toBeEnabled();
    await expect(album.getByLabel('作品统计').locator('dd')).toHaveText(['未知', '未知', '未知', '未知', '未知']);
    await expect(album.getByLabel('作品统计').locator('dt')).toHaveText(['AI 平台', '作品', '文件夹', '接力过', '最近在动']);
    expect(server.query('SELECT COUNT(*) AS n FROM works_album')[0].n).toBe(0);

    await album.getByRole('button', { name: '扫描导入项目', exact: true }).click();
    let dialog = page.getByRole('dialog');
    const choices = dialog.locator('label').filter({ has: page.locator('code') });
    const selected = choices.filter({ hasText: resolve(approvedRoot!) });
    await expect(selected).toHaveCount(1);
    await selected.getByRole('checkbox').check();
    const scanResponse = page.waitForResponse(response => response.url().endsWith('/works/refresh') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '扫描所选项目', exact: true }).click();
    expect((await scanResponse).status()).toBe(200);
    await expect(dialog.getByRole('status')).toContainText('已保存文件观察');
    await page.keyboard.press('Escape');
    let snapshot = await read();
    expect(snapshot.works).toHaveLength(1);
    const work = snapshot.works[0];
    expect(rootKey(work.root)).toBe(rootKey(resolve(approvedRoot!)));
    expect(work.files.verified).toBe(true);
    expect(work.files.file_count).toBeGreaterThan(0);
    expect(work.platform).toBe('unknown');
    const card = album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true });
    await expect(card).toContainText('平台未知');
    await expect(card).toContainText('从文件核实');
    await card.focus();
    await page.keyboard.press('Enter');
    dialog = page.getByRole('dialog');
    await dialog.getByLabel('记录态度').selectOption('continue');
    await dialog.getByRole('button', { name: '保存态度', exact: true }).click();
    await expect(dialog.getByRole('status')).toContainText('人工记录已保存');
    await dialog.getByLabel('编辑分组').fill('W7 验收');
    await dialog.getByRole('button', { name: '保存分组', exact: true }).click();
    await expect(dialog.getByRole('button', { name: '保存分组', exact: true })).toBeEnabled();
    await dialog.getByLabel('编辑交接手册').fill('实际 W7 工作树：下一步等待链1接线，不推断完整完成。');
    await dialog.getByRole('button', { name: '保存交接手册', exact: true }).click();
    await expect(dialog.getByRole('button', { name: '保存交接手册', exact: true })).toBeEnabled();
    await expect(dialog).toContainText('人工编写');
    await dialog.getByRole('button', { name: '隐藏作品', exact: true }).click();
    await expect(dialog.getByRole('button', { name: '取消隐藏作品', exact: true })).toBeEnabled();
    await page.keyboard.press('Escape');
    await expect(card).toBeHidden();
    await album.getByLabel('显示已隐藏').check();
    await card.click();
    await page.getByRole('dialog').getByRole('button', { name: '取消隐藏作品', exact: true }).click();
    await expect(page.getByRole('dialog').getByRole('button', { name: '隐藏作品', exact: true })).toBeEnabled();
    await page.keyboard.press('Escape');

    // 未登记路径只保存为投递线索，不读取文件。
    await album.getByRole('button', { name: 'AI 投递 JSON', exact: true }).click();
    dialog = page.getByRole('dialog');
    await dialog.getByLabel('AI 投递 JSON 内容').fill(JSON.stringify({ works: [{ name: '明确投递线索', platform: 'codex', path: `${server.temporary}/unregistered-clue`, handoff: '生成手册只是线索，未执行任务。' }] }));
    await dialog.getByRole('button', { name: '保存 JSON 投递', exact: true }).click();
    await expect(dialog.getByRole('status')).toContainText('JSON 投递已保存');
    await page.keyboard.press('Escape');
    await expect(album.getByRole('button', { name: '查看作品 明确投递线索', exact: true })).toContainText('文件数量未知');
    await expect(album.getByLabel('作品统计').locator('div').filter({ has: page.getByText('最近在动', { exact: true }) }).locator('dd')).toHaveText('未知');

    for (const view of ['平铺', '按平台', '按分组', '时间线']) {
      await album.getByRole('button', { name: view, exact: true }).click();
      await expect(album.getByRole('button', { name: /^查看作品 / })).toHaveCount(2);
    }
    const beforeSync = await read();
    const syncResponse = page.waitForResponse(response => response.url().endsWith('/works/refresh') && response.request().method() === 'POST');
    await album.getByRole('button', { name: '同步最近改动', exact: true }).click();
    expect((await syncResponse).status()).toBe(200);
    snapshot = await read();
    expect(snapshot.works.find(item => item.id === work.id)?.handoff).toEqual(beforeSync.works.find(item => item.id === work.id)?.handoff);
    expect(snapshot.works).toHaveLength(2);
    expect(commands.filter(command => command.path.endsWith('/refresh')).map(command => (command.body.roots as string[]).map(rootKey))).toEqual([[rootKey(resolve(approvedRoot!))], [rootKey(resolve(approvedRoot!))]]);
    const beforeRestart = await read();
    const metadata = server.query('SELECT work_id,data,revision FROM works_metadata ORDER BY work_id');
    const requests = server.query('SELECT caller_id,command,request_key,status,typeof(input) AS input_type,typeof(result) AS result_type FROM works_requests ORDER BY request_key');
    expect(requests).toHaveLength(commands.length);
    for (const receipt of requests) expect(receipt).toMatchObject({ status: 'accepted', input_type: 'blob', result_type: 'blob' });
    const albumBytes = server.query('SELECT data FROM works_album');
    // HTTP 拒绝缺少 CSRF 或越界目录的输入，存储保持原样。
    const rejected = await api.request('/works/refresh', { method: 'POST', body: { schema_version: 1, request_id: 'scope-rejected', expected_version: 1, roots: [server.temporary] }, key: 'scope-rejected' });
    expect(rejected.status).toBe(403);
    const csrf = await api.request('/works/refresh', { method: 'POST', body: { schema_version: 1, request_id: 'csrf-rejected', expected_version: 1, roots: [work.root] }, key: 'csrf-rejected', headers: { 'X-CSRF-Token': '' } });
    expect(csrf.status).toBe(403);
    expect(server.query('SELECT data FROM works_album')).toEqual(albumBytes);
    await server.restart();
    api = await humanAPI(server.apiURL);
    expect(await read()).toEqual(beforeRestart);
    expect(server.query('SELECT work_id,data,revision FROM works_metadata ORDER BY work_id')).toEqual(metadata);
    expect(server.query('SELECT data FROM works_album')).toEqual(albumBytes);
    await page.reload();
    await expect(card).toContainText('实际 W7 工作树');
    await expect(card).toContainText('W7 验收');
    for (const width of [1920, 1280, 390]) {
      await page.setViewportSize({ width, height: width === 390 ? 844 : width === 1280 ? 720 : 1080 });
      await expect(card).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
      await card.focus(); await page.keyboard.press('Enter');
      await expect(page.getByRole('dialog')).toContainText('人工编写');
      await page.keyboard.press('Escape');
      await expect(card).toBeFocused();
    }
    expect(errors).toEqual([]);
    expect(new Set(commands.map(command => command.key)).size).toBe(commands.length);
    await writeFile(testInfo.outputPath('actual-works-evidence.json'), JSON.stringify({ snapshot: beforeRestart, metadata, requests, commands, errors, chain1: 'NOT_RUN_nil_sink_pending_W1_W0' }, null, 2));
  } catch (error) { primaryError = error; throw error; }
  finally { await server.close({ preserveData: true, primaryError }); }
});

test('saved actual works render at three viewports with keyboard and no writes', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_W7_VIEWS !== '1', 'Requires the explicitly owned successful Works database.');
  test.setTimeout(300_000);
  const owned = process.env.ASTROCYTE_W7_REUSE_OWNED_TEMP;
  expect(owned).toBeTruthy();
  const server = await startS1Server({ browser: true, apiPort: 18847, webPort: 15247, reuseOwnedTemporary: { path: owned!, ownedRoot: owned! } });
  const api = await humanAPI(server.apiURL);
  const writes: string[] = [];
  const errors: string[] = [];
  page.on('request', request => { if (request.method() !== 'GET' && new URL(request.url()).pathname.startsWith('/api/v1/works')) writes.push(request.url()); });
  page.on('pageerror', error => errors.push(error.message));
  let primaryError: unknown;
  try {
    const before = server.query('SELECT data FROM works_album');
    const metadata = server.query('SELECT work_id,data,revision FROM works_metadata ORDER BY work_id');
    const snapshot = ((await api.get('/works')) as S['WorksSnapshotResultV1']).snapshot;
    const observations = server.query("SELECT raw_bytes,mapping_status FROM attention_engine_events WHERE origin_key LIKE 'works:%'");
    for (const observation of observations) {
      expect(observation.mapping_status).toBe('pending');
      expect(JSON.parse(rawText(observation.raw_bytes))).toMatchObject({ topic: '' });
    }
    const actual = snapshot.works.find(work => work.source === 'scan');
    expect(actual).toBeTruthy();
    await page.goto(`${server.webURL}/workspace`);
    const album = page.getByRole('region', { name: '作品寻回册', exact: true });
    const card = album.getByRole('button', { name: `查看作品 ${actual!.name}`, exact: true });
    for (const width of [1920, 1280, 390]) {
      await page.setViewportSize({ width, height: width === 390 ? 844 : width === 1280 ? 720 : 1080 });
      await expect(card).toContainText('从文件核实');
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
      await card.focus(); await page.keyboard.press('Enter');
      const dialog = page.getByRole('dialog');
      await expect(dialog).toContainText('人工编写');
      await expect(dialog.getByLabel('编辑交接手册')).toHaveValue(actual!.handoff.content);
      await expect(dialog.getByRole('button', { name: '保存交接手册', exact: true })).toBeEnabled();
      await page.keyboard.press('Tab');
      expect(await dialog.evaluate(element => element.contains(document.activeElement))).toBe(true);
      await page.keyboard.press('Escape');
      await expect(card).toBeFocused();
    }
    expect(writes).toEqual([]);
    expect(errors).toEqual([]);
    expect(server.query('SELECT data FROM works_album')).toEqual(before);
    expect(server.query('SELECT work_id,data,revision FROM works_metadata ORDER BY work_id')).toEqual(metadata);
    await writeFile(testInfo.outputPath('saved-works-views-evidence.json'), JSON.stringify({ owned, writes, errors, viewports: [1920, 1280, 390], snapshot }, null, 2));
  } catch (error) { primaryError = error; throw error; }
  finally { await server.close({ preserveData: true, primaryError }); }
});

test('same owned works API persists raw pending chain1 observations without signals', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_W7_SOURCE !== '1', 'Requires exact W0 pending-only startup and the owned successful Works store.');
  test.setTimeout(300_000);
  const owned = process.env.ASTROCYTE_W7_REUSE_OWNED_TEMP;
  expect(owned).toBeTruthy();
  const server = await startS1Server({ browser: true, apiPort: 18847, webPort: 15247, reuseOwnedTemporary: { path: owned!, ownedRoot: owned! } });
  let api = await humanAPI(server.apiURL);
  let primaryError: unknown;
  const read = async () => ((await api.get('/works')) as S['WorksSnapshotResultV1']).snapshot;
  const rows = () => server.query("SELECT id,origin_key,source,kind,topic,data,raw_bytes,observation,mapping_status,typeof(raw_bytes) AS raw_type FROM attention_engine_events WHERE origin_key LIKE 'works:%' ORDER BY origin_key");
  let captured: { body: Record<string, unknown>; key: string; path: string } | undefined;
  page.on('request', request => {
    if (request.method() === 'PUT' && new URL(request.url()).pathname.endsWith('/group')) captured = { body: request.postDataJSON() as Record<string, unknown>, key: request.headers()['idempotency-key'], path: new URL(request.url()).pathname.replace('/api/v1', '') };
  });
  try {
    const initial = await read();
    const actual = initial.works.find(work => work.source === 'scan');
    expect(actual).toBeTruthy();
    const before = rows();
    const signals = server.query('SELECT COUNT(*) AS n FROM attention_engine_signals');
    const nodes = server.query('SELECT id,attention,momentum,confidence,last_active,updated_at FROM attention_engine_nodes ORDER BY id');
    await page.goto(`${server.webURL}/workspace`);
    const album = page.getByRole('region', { name: '作品寻回册', exact: true });
    const card = album.getByRole('button', { name: `查看作品 ${actual!.name}`, exact: true });
    await card.click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('编辑分组').fill('W7 链1验收');
    const saved = page.waitForResponse(response => response.url().endsWith('/group') && response.request().method() === 'PUT');
    await dialog.getByRole('button', { name: '保存分组', exact: true }).click();
    expect((await saved).status()).toBe(200);
    await expect(dialog.getByRole('status')).toContainText('人工记录已保存');
    expect(captured).toBeTruthy();
    const afterHuman = rows();
    expect(afterHuman).toHaveLength(before.length + 1);
    const human = afterHuman.find(row => !before.some(old => old.id === row.id))!;
    expect(human).toMatchObject({ source: 'manual', kind: 'work_group_set', topic: '', mapping_status: 'pending', raw_type: 'blob', data: 'null' });
    const rawHuman = JSON.parse(rawText(human.raw_bytes)) as { work_id: string; kind: string; source: string; topic: string; payload: S['WorkV1'] };
    expect(rawHuman).toMatchObject({ work_id: actual!.id, kind: 'work_group_set', source: 'manual', topic: '', payload: { group: 'W7 链1验收', revision: actual!.revision + 1 } });
    const replay = await api.request(captured!.path, { method: 'PUT', body: captured!.body, key: captured!.key });
    expect(replay.status).toBe(200);
    expect(rows()).toEqual(afterHuman);
    await page.keyboard.press('Escape');
    const synced = page.waitForResponse(response => response.url().endsWith('/works/refresh') && response.request().method() === 'POST');
    await album.getByRole('button', { name: '同步最近改动', exact: true }).click();
    expect((await synced).status()).toBe(200);
    const afterMachine = rows();
    expect(afterMachine).toHaveLength(afterHuman.length + 1);
    const machine = afterMachine.find(row => !afterHuman.some(old => old.id === row.id))!;
    expect(machine).toMatchObject({ source: 'machine', kind: 'work_synced', topic: '', mapping_status: 'pending', raw_type: 'blob', data: 'null' });
    expect(server.query('SELECT COUNT(*) AS n FROM attention_engine_signals')).toEqual(signals);
    expect(server.query('SELECT id,attention,momentum,confidence,last_active,updated_at FROM attention_engine_nodes ORDER BY id')).toEqual(nodes);
    const beforeRestart = await read();
    const receiptBytes = server.query('SELECT request_key,status,input,result FROM works_requests ORDER BY request_key');
    await server.restart();
    api = await humanAPI(server.apiURL);
    expect(await read()).toEqual(beforeRestart);
    expect(rows()).toEqual(afterMachine);
    expect(server.query('SELECT request_key,status,input,result FROM works_requests ORDER BY request_key')).toEqual(receiptBytes);
    const replayAfterRestart = await api.request(captured!.path, { method: 'PUT', body: captured!.body, key: captured!.key });
    expect(replayAfterRestart.status).toBe(200);
    expect(rows()).toEqual(afterMachine);
    expect(server.query('SELECT COUNT(*) AS n FROM attention_engine_signals')).toEqual(signals);
    expect(server.query('SELECT id,attention,momentum,confidence,last_active,updated_at FROM attention_engine_nodes ORDER BY id')).toEqual(nodes);
    await page.reload();
    await expect(card).toContainText('W7 链1验收');
    await writeFile(testInfo.outputPath('actual-works-chain1-evidence.json'), JSON.stringify({ startup: '396aec1', root: actual!.root, before_receipts: before.length, after_receipts: afterMachine.length, human: { ...human, raw_bytes: rawText(human.raw_bytes) }, machine: { ...machine, raw_bytes: rawText(machine.raw_bytes) }, signals, nodes, request: captured, restart_equal: true }, null, 2));
  } catch (error) { primaryError = error; throw error; }
  finally { await server.close({ preserveData: true, primaryError }); }
});
