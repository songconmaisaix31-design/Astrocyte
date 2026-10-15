import { chromium, expect, type Locator, type Page } from '@playwright/test';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import { formatDate, formatNumber } from '../src/pages/cognition/format.ts';
import type { CognitionAgents, CognitionEvent, CognitionEvidence, CognitionFocus, CognitionHealth, CognitionToday } from '../src/api/cognitionTypes';

type SQLValue = { route: string; metric: string; topic?: string; source?: string; day?: string; value: number };
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const source = process.env.ASTROCYTE_TEST_PRODUCT_SOURCE;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (origin !== 'http://127.0.0.1:64505' || !source || !/^[0-9a-f]{40}$/.test(source) || !output) throw new Error('需要H当前同一产品入口、准确SOURCE和独占输出目录');
const snapshotRoot = 'C:/Users/DW/orca/workspaces/Astrocyte/release-cognition-20261014/.verification/cross-day-20261015-0228/local/astrocyte-workbench/v0.1';
const pointer = JSON.parse(await readFile(resolve(snapshotRoot, 'current.json'), 'utf8')) as { generation: string; snapshot_at: string };
expect(pointer.generation).toBe('snapshot-1711995690');
const expected = JSON.parse(await readFile(resolve(snapshotRoot, '../../../main-safe-final/sql-expected.json'), 'utf8')) as { snapshot_at: string; day: string; values: SQLValue[] };
expect(expected.values).toHaveLength(10);
expect(Date.parse(pointer.snapshot_at)).toBe(Date.parse(expected.snapshot_at));
const directory = resolve(output);
await mkdir(directory, { recursive: true });
const path = resolve(directory, 'product-cognition-readonly.json');
const evidence: {
  source: string; generation: string; stage: string; started_at: string; completed_at?: string;
  snapshot_at: string; original_sql_day: string; actual_day?: string;
  metrics: (SQLValue & { actual_value: number; text: string })[];
  topics: { topic: string; points: number; events: number; pages: number }[];
  event_ids: string[]; writes: string[]; page_errors: string[];
} = { source, generation: pointer.generation, stage: 'started', started_at: new Date().toISOString(), snapshot_at: pointer.snapshot_at,
  original_sql_day: expected.day, metrics: [], topics: [], event_ids: [], writes: [], page_errors: [] };
await writeFile(path, JSON.stringify(evidence, null, 2), { flag: 'wx' });
const save = () => writeFile(path, JSON.stringify(evidence, null, 2));
const database = new DatabaseSync(resolve(snapshotRoot, 'snapshots', pointer.generation, 'snapshot.sqlite'), { readOnly: true });
const browser = await chromium.launch();
try {
  const api = await humanAPI(origin);
  const health = await api.get('/cognition/health') as CognitionHealth;
  expect(Date.parse(health.snapshot_at)).toBe(Date.parse(pointer.snapshot_at));
  expect(health.timezone).toBe('Asia/Shanghai');
  const today = await api.get('/cognition/today') as CognitionToday;
  const actualDay = new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
  expect(today.day).toBe(actualDay);
  evidence.actual_day = actualDay;
  const focus = await api.get('/cognition/focus') as CognitionFocus;
  const agents = await api.get('/cognition/agents') as CognitionAgents;
  const nodes = database.prepare('SELECT topic,attention,momentum,confidence FROM attention_engine_nodes ORDER BY attention DESC,topic ASC').all();
  expect(focus.topics).toEqual(nodes);
  expect(focus.topics).toHaveLength(2);
  const history = JSON.parse(await readFile(resolve(snapshotRoot, 'snapshots', pointer.generation, 'history.json'), 'utf8')) as Record<string, { ts: number; attention: number }[]>;
  for (const line of focus.series) {
    expect(line.points.map(point => ({ ts: Date.parse(point.ts), attention: point.attention }))).toEqual(history[line.topic]);
    expect(line.points.length).toBeGreaterThanOrEqual(2);
  }
  expect(focus.series.map(line => line.topic).sort()).toEqual(Object.keys(history).sort());
  const start = Date.parse(`${actualDay}T00:00:00+08:00`), end = start + 86400_000;
  const count = (sql: string, ...args: (number | string)[]) => (database.prepare(sql).get(...args) as { count: number }).count;
  const context = await browser.newContext({ viewport: { width: 1280, height: 720 } });
  const page = await context.newPage();
  page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') evidence.writes.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  page.on('pageerror', error => evidence.page_errors.push(error.name));
  for (const item of expected.values) {
    evidence.stage = `metric:${item.route}:${item.metric}`;
    await save();
    let value: number, id: string;
    if (item.route === '/cognition/today' && item.metric === 'daily_count' && item.day) {
      const dayStart = Date.parse(`${item.day}T00:00:00+08:00`);
      value = count('SELECT COUNT(*) AS count FROM attention_engine_events WHERE ts>=? AND ts<?', dayStart, dayStart + 86400_000);
      expect(today.daily_counts.find(day => day.day === item.day)?.count).toBe(value);
      expect(value).toBe(item.value);
      id = `daily-${item.day}`;
    } else if (item.route === '/cognition/focus' && ['attention', 'momentum', 'confidence'].includes(item.metric)) {
      value = (database.prepare(`SELECT ${item.metric} AS count FROM attention_engine_nodes WHERE topic=?`).get(item.topic!) as { count: number }).count;
      expect(value).toBe(item.value);
      id = `focus-${{ attention: 'a', momentum: 'm', confidence: 'c' }[item.metric as 'attention' | 'momentum' | 'confidence']}-${item.topic}`;
    } else if (item.route === '/cognition/agents' && ['today', 'days7', 'days30'].includes(item.metric)) {
      const lower = start - ({ today: 0, days7: 6, days30: 29 }[item.metric as 'today' | 'days7' | 'days30']) * 86400_000;
      value = count('SELECT COUNT(*) AS count FROM attention_engine_events WHERE source=? AND ts>=? AND ts<?', item.source!, lower, end);
      expect(agents.groups.find(group => group.source === item.source)?.[item.metric as 'today' | 'days7' | 'days30']).toBe(value);
      id = `agent-${item.source}-${item.metric}`;
    } else if (item.route === '/cognition/evidence' && item.metric === 'total') {
      value = count('SELECT COUNT(*) AS count FROM attention_engine_events');
      expect(value).toBe(item.value);
      id = 'events-total';
    } else throw new Error('独立SQL预期包含未支持的实际指标');
    await page.goto(`${origin}${item.route}`);
    await expect(page.getByTestId('snapshot-time')).toBeVisible();
    const metric = page.getByTestId(id);
    await expect(metric).toHaveAttribute('data-value', String(value));
    await expect(metric).toHaveText(formatNumber(value));
    evidence.metrics.push({ ...item, actual_value: value, text: (await metric.innerText()).trim() });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  }
  const locate = (events: CognitionEvent[]) => {
    for (const event of events) {
      expect(Object.keys(event).sort()).toEqual(['event_id', 'kind', 'scope', 'tokens', 'topic', 'ts']);
      const row = database.prepare("SELECT id,kind,topic,CASE WHEN json_type(data,'$.tokens')='integer' THEN json_extract(data,'$.tokens') END AS tokens,CASE WHEN json_type(data,'$.scope')='text' THEN json_extract(data,'$.scope') ELSE '' END AS scope,ts FROM attention_engine_events WHERE id=?").get(event.event_id) as { id: string; kind: string; topic: string; tokens: number | null; scope: string; ts: number };
      expect(row).toEqual({ id: event.event_id, kind: event.kind, topic: event.topic, tokens: event.tokens, scope: event.scope, ts: Date.parse(event.ts) });
      evidence.event_ids.push(event.event_id);
    }
  };
  const table = async (container: Locator | Page, events: CognitionEvent[]) => {
    const rows = container.locator('.cg-events tbody tr');
    await expect(rows).toHaveCount(events.length);
    await expect(rows.first().locator('td').first()).toHaveText(events[0].event_id);
    expect(await rows.evaluateAll(items => items.map(row => [...row.querySelectorAll('td')].map(cell => cell.textContent?.trim())))).toEqual(events.map(event => [event.event_id, event.kind, event.topic || '未标注', formatNumber(event.tokens), event.scope || '未标注', formatDate(event.ts)]));
  };
  for (const topic of focus.topics) {
    evidence.stage = `topic:${topic.topic}`;
    await save();
    await page.goto(`${origin}/cognition/focus`);
    await expect(page.getByRole('img', { name: '多主题注意力历史与实验轨迹' })).toBeVisible();
    const trigger = page.locator('.cg-topic-row').filter({ has: page.getByTestId(`focus-a-${topic.topic}`) });
    const first = await api.get(`/cognition/topics/evidence?${new URLSearchParams({ topic: topic.topic, page: '1', page_size: '20' })}`) as CognitionEvidence;
    const total = count('SELECT COUNT(*) AS count FROM attention_engine_events WHERE topic=?', topic.topic);
    expect(first.total).toBe(total);
    expect(total).toBeGreaterThan(20);
    await trigger.click();
    const dialog = page.getByRole('dialog', { name: `${topic.topic} · 主题证据`, exact: true });
    const ids = new Set<string>();
    let result = first;
    const pages = Math.ceil(total / 20);
    for (let number = 1; number <= pages; number++) {
      expect(result.page).toBe(number);
      expect(result.total).toBe(total);
      expect(result.events).toHaveLength(Math.min(20, total - (number - 1) * 20));
      locate(result.events);
      for (const event of result.events) { expect(ids.has(event.event_id)).toBe(false); ids.add(event.event_id); }
      await table(dialog, result.events);
      if (number < pages) {
        const response = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/cognition/topics/evidence') && new URL(response.url()).searchParams.get('topic') === topic.topic && new URL(response.url()).searchParams.get('page') === String(number + 1));
        await dialog.getByRole('button', { name: '下一页', exact: true }).click();
        const next = await response;
        expect(next.status()).toBe(200);
        result = await next.json() as CognitionEvidence;
      }
    }
    expect(ids.size).toBe(total);
    await expect(dialog.getByRole('button', { name: '下一页', exact: true })).toBeDisabled();
    await page.keyboard.press('Escape');
    await expect(trigger).toBeFocused();
    evidence.topics.push({ topic: topic.topic, points: history[topic.topic].length, events: total, pages });
    await save();
  }
  expect(evidence.writes).toEqual([]);
  expect(evidence.page_errors).toEqual([]);
  expect(JSON.parse(await readFile(resolve(snapshotRoot, 'current.json'), 'utf8'))).toEqual(pointer);
  evidence.stage = 'readonly_completed';
  evidence.completed_at = new Date().toISOString();
  await save();
  console.log(JSON.stringify({ ...evidence, event_ids: { observed: evidence.event_ids.length, unique: new Set(evidence.event_ids).size } }));
} finally { await save(); await browser.close(); database.close(); }
