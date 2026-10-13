/// <reference types="node" />
import { test, expect, type Page } from '@playwright/test';
import { readFile, realpath } from 'node:fs/promises';
import { join, relative, isAbsolute } from 'node:path';
import type { CognitionToday } from '../src/api/cognitionTypes';
import type { RelaySessionEvents, RelaySessionCoverage } from '../src/api/relaySession';

const taskID = 'cf1497b0-e092-43c0-80a9-33449bb57a40';
const coverage: RelaySessionCoverage = { task_id: taskID, terminal_id: 'c48d6c05-3b58-4bdc-8a7b-46a471612a0b', worktree_id: '0623c675-ddb7-4d62-bd02-831f6cb6d51d', project_id: 'public-project', status: 'partial', gap: 'prefix_unverified', lines: 3, projected: 2, duplicates: 1, unknown: 1, errors: 1, unknown_tokens: 2, remaining_bytes: 4, tail_bytes: 3, snapshot_complete: true, session_complete: false };
const local: RelaySessionEvents = { events: [{ event_id: 'public-session-event', kind: 'edit_file', topic: 'public-session', tokens: null, scope: 'public-project', ts: '2026-10-14T01:00:01Z' }], coverage: [coverage], total: 51, page: 1, page_size: 50 };
const today: CognitionToday = { day: '2026-10-14', focus: [{ topic: '公开快照甲', attention: 3.25, momentum: 0.4, confidence: 0.8, attention_delta: null }, { topic: '公开快照乙', attention: 1.5, momentum: -0.1, confidence: 0.6, attention_delta: null }], activity: [], sleep: { day: '2026-10-13', status: 'completed', completed_at: '2026-10-14T00:30:00Z' }, daily_counts: [{ day: '2026-10-12', count: 0 }, { day: '2026-10-13', count: 5 }, { day: '2026-10-14', count: 10 }], sources: [] };

async function setup(page: Page, initial = structuredClone(local)) {
  const state = { body: initial, failure: 0, snapshotFailure: false, requests: [] as URL[], writes: [] as string[] };
  page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') state.writes.push(request.method()); });
  await page.route('**/api/v1/cognition/**', async route => {
    expect(route.request().method()).toBe('GET');
    const suffix = new URL(route.request().url()).pathname.split('/').at(-1);
    const body = suffix === 'health' ? { status: 'ok', snapshot_at: '2026-10-14T00:35:00Z', timezone: 'Asia/Shanghai' } : suffix === 'filters' ? { topics: today.focus.map(topic => topic.topic), kinds: [], sources: [] } : today;
    await route.fulfill(state.snapshotFailure && suffix === 'today' ? { status: 503, json: { schema_version: 1, error: { code: 'provider_unavailable', message: 'Snapshot unavailable', request_id: 'public-snapshot', retryable: false, required_action: 'read_published_snapshot' } } } : { json: body });
  });
  await page.route('**/api/v1/relay/session-events**', async route => {
    const request = route.request(), url = new URL(request.url()); state.requests.push(url);
    expect(request.method()).toBe('GET');
    if (state.failure) {
      await route.fulfill({ status: state.failure, json: { schema_version: 1, error: { code: state.failure === 403 ? 'scope_denied' : 'provider_unavailable', message: state.failure === 403 ? 'Current human project permission required' : 'Relay provider unavailable', required_action: state.failure === 403 ? 'review_project_permissions' : 'inspect_local_relay_provider', retryable: false, request_id: 'public-read' } } });
      return;
    }
    await route.fulfill({ json: { ...state.body, page: Number(url.searchParams.get('page') || state.body.page), events: state.body.events.map(event => ({ ...event, raw: 'PRIVATE_RAW_CANARY', prompt: 'PRIVATE_RAW_CANARY', thought: 'PRIVATE_RAW_CANARY', body: 'PRIVATE_RAW_CANARY', command: 'PRIVATE_RAW_CANARY', path: 'PRIVATE_RAW_CANARY', native_id: 'PRIVATE_NATIVE_ID' })) } });
  });
  return state;
}
async function snapshotTen(page: Page) {
  for (const [name, value] of [['today-a-公开快照甲', '3.25'], ['today-m-公开快照甲', '0.4'], ['today-c-公开快照甲', '0.8'], ['today-a-公开快照乙', '1.5'], ['today-m-公开快照乙', '-0.1'], ['today-c-公开快照乙', '0.6'], ['daily-2026-10-12', '0'], ['daily-2026-10-13', '5'], ['daily-2026-10-14', '10'], ['sleep-status', '已完成']]) await expect(page.getByTestId(name)).toHaveText(value);
  await expect(page.getByTestId('snapshot-time')).toContainText('2026/10/14');
  await expect(page.getByTestId('today-delta-公开快照甲')).toContainText('缺少昨日确证值');
}

test.describe('L1 contract-local UI evidence; not native CLI or production acceptance', () => {
  test('complete claim is qualified when unknown records, errors or lifecycle gaps remain', async ({ page }) => {
    const state = await setup(page, { ...structuredClone(local), coverage: [{ ...coverage, status: 'complete', gap: '', unknown: 0, errors: 0, remaining_bytes: 0, tail_bytes: 0, session_complete: true }] });
    await page.goto('/cognition/today');
    const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
    await expect(region).toContainText('完整（服务报告） · complete');
    await expect(region).toContainText('完整会话生命周期：服务报告已核实');
    state.body.coverage[0] = { ...state.body.coverage[0], unknown: 1, errors: 1, tail_bytes: 3, session_complete: false };
    await region.getByRole('button', { name: '只读刷新会话事件', exact: true }).click();
    await expect(region).toContainText('完整主张尚未核实 · complete');
    await expect(region).toContainText('不能据此宣称会话完整');
    await expect(region).toContainText('完整会话生命周期：尚未核实');
    await expect(region).not.toContainText('完整（服务报告）');
    await snapshotTen(page);
    expect(state.writes).toEqual([]);
  });

  test('six-field admissions preserve snapshot ten values, missing tokens, gaps, GET paging and narrow layout', async ({ page }) => {
    const state = await setup(page); await page.goto('/cognition/today');
    const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
    await expect(region.locator('tbody tr')).toHaveCount(1);
    await expect(region.locator('thead th')).toHaveCount(6);
    await expect(region.locator('tbody')).toContainText('暂无');
    await expect(region).toContainText('当前文件快照：已读完当前冻结字节范围');
    await expect(region).toContainText('完整会话生命周期：尚未核实');
    for (const label of ['未知格式', '错误', '剩余字节', '未完成尾部字节']) await expect(region.locator('dl div').filter({ hasText: label }).locator('dd')).not.toHaveText('0');
    await snapshotTen(page);
    await region.getByRole('button', { name: '会话下一页', exact: true }).click();
    await expect.poll(() => state.requests.some(url => url.searchParams.get('page') === '2')).toBe(true);
    state.body.events.push({ ...state.body.events[0], event_id: 'public-new-admission', tokens: 99999, ts: '2026-10-14T01:00:02Z' });
    await region.getByRole('button', { name: '只读刷新会话事件', exact: true }).click();
    await expect(region.locator('tbody tr')).toHaveCount(2);
    await expect(region.locator('tbody tr').first()).toContainText('public-new-admission');
    await snapshotTen(page);
    await expect(page.locator('body')).not.toContainText('PRIVATE_RAW_CANARY');
    await expect(page.locator('body')).not.toContainText('PRIVATE_NATIVE_ID');
    await page.setViewportSize({ width: 390, height: 844 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    expect(state.writes).toEqual([]);
  });

  for (const status of [403, 503]) {
    test(`HTTP ${status} makes current coverage unknown and hides previous cache during retry`, async ({ page }) => {
      const state = await setup(page); await page.goto('/cognition/today');
      const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
      await expect(region.locator('tbody tr')).toHaveCount(1);
      state.failure = status;
      await region.getByRole('button', { name: '只读刷新会话事件', exact: true }).click();
      await expect(region).toContainText('当前会话覆盖未知');
      await expect(region.locator('tbody tr')).toHaveCount(0);
      await expect(region).not.toContainText('public-session-event');
      await snapshotTen(page);
      expect(state.writes).toEqual([]);
    });
  }

  test('empty and first capture failure retain unknown counts without raw gap text or completeness promotion', async ({ page }) => {
    const state = await setup(page, { events: [], coverage: [{ ...coverage, status: 'unknown', gap: 'PRIVATE_RAW_CANARY /native/path', lines: 0, projected: 0, unknown: 0, errors: 1, snapshot_complete: false }], total: 0, page: 1, page_size: 50 });
    await page.goto('/cognition/today');
    const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
    await expect(region).toContainText('未知 · unknown');
    await expect(region).toContainText('存在未识别的覆盖缺口');
    await expect(region).toContainText('空结果不证明会话完整');
    await expect(region.locator('dl div').filter({ hasText: '错误' }).locator('dd')).toHaveText('1');
    await expect(region).not.toContainText('PRIVATE_RAW_CANARY');
    state.body.coverage = [];
    await region.getByRole('button', { name: '只读刷新会话事件', exact: true }).click();
    await expect(region).toContainText('尚无会话覆盖证据 · unknown');
    await snapshotTen(page);
    expect(state.writes).toEqual([]);
  });

  test('manual strict filters stay GET and invalid input is not submitted', async ({ page }) => {
    const state = await setup(page); await page.goto('/cognition/today');
    const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
    await region.getByLabel('会话任务 ID', { exact: true }).fill('/private/path');
    await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
    await expect(region).toContainText('查询未提交');
    expect(state.requests.every(url => !url.searchParams.has('task_id'))).toBe(true);
    await region.getByLabel('会话任务 ID', { exact: true }).fill(taskID);
    await region.getByLabel('会话开始时间', { exact: true }).fill('2026-10-14T00:00:00Z');
    await region.getByLabel('会话结束时间', { exact: true }).fill('2026-10-15T00:00:00Z');
    await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
    await expect.poll(() => state.requests.some(url => url.searchParams.get('task_id') === taskID && url.searchParams.get('start') === '2026-10-14T00:00:00Z' && url.searchParams.get('end') === '2026-10-15T00:00:00Z' && url.searchParams.get('page') === '1')).toBe(true);
    expect(state.writes).toEqual([]);
  });

  test('session admissions remain independently readable when the snapshot is unavailable', async ({ page }) => {
    const state = await setup(page); state.snapshotFailure = true; await page.goto('/cognition/today');
    await expect(page.getByRole('region', { name: '本机 Session 派生事件', exact: true }).locator('tbody tr')).toHaveCount(1);
    await expect(page.getByText('API 503', { exact: false })).toBeVisible();
    await expect(page.getByTestId('today-a-公开快照甲')).toHaveCount(0);
    expect(state.writes).toEqual([]);
  });
});

test.describe('H actual HTTP public DTO replay; not live browser/service acceptance', () => {
  for (const index of [0, 1, 2]) {
    test(`recorded session response ${index + 1} preserves coverage and snapshot separation`, async ({ page }) => {
      const samplePath = process.env.ASTROCYTE_TEST_RELAY_L1_HTTP_SAMPLE;
      test.skip(!samplePath, 'Requires the exact H-owned derived public HTTP sample path.');
      const root = await realpath(join(process.env.LOCALAPPDATA!, 'astrocyte-workbench/v0.2-access/verification/session-http'));
      const selected = await realpath(samplePath!);
      const rel = relative(root, selected);
      expect(!rel.startsWith('..') && !isAbsolute(rel)).toBe(true);
      const samples: { schema: string; body: RelaySessionEvents }[] = JSON.parse(await readFile(selected, 'utf8'));
      expect(samples).toHaveLength(3);
      expect(samples[index].schema).toBe('RelaySessionEventsV1');
      const body = samples[index].body;
      const state = await setup(page, body); await page.goto('/cognition/today');
      const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
      if (body.coverage.length) {
        await expect(region).toContainText(`· ${body.coverage[0].status}`);
        await expect(region).toContainText('完整会话生命周期：尚未核实');
      } else await expect(region).toContainText('尚无会话覆盖证据 · unknown');
      await expect(region.locator('tbody tr')).toHaveCount(body.events.length);
      await snapshotTen(page);
      await expect(page.locator('body')).not.toContainText('PRIVATE_RAW_CANARY');
      expect(state.writes).toEqual([]);
    });
  }
});
