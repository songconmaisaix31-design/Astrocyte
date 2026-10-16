/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import { cognitionTestServer } from '../src/api/realHttpTest';
import type { CognitionHealth } from '../src/api/cognitionTypes';
import { formatNumber } from '../src/pages/cognition/format';

type SQLValue = { route: string; metric: string; topic?: string; source?: string; day?: string; value: number; query: Record<string, string> };
type SQLExpected = { snapshot_at: string; day: string; values: SQLValue[] };

function actualSQLValues(database: DatabaseSync, expected: SQLExpected, day: string) {
  const start = Date.parse(`${day}T00:00:00+08:00`), end = start + 86400_000;
  const count = (sql: string, ...parameters: (string | number)[]) => (database.prepare(sql).get(...parameters) as { count: number }).count;
  return expected.values.map(item => {
    let value: number;
    if (item.route === '/cognition/today' && item.metric === 'daily_count' && item.day) {
      const lower = Date.parse(`${item.day}T00:00:00+08:00`);
      value = count('SELECT COUNT(*) AS count FROM attention_engine_events WHERE ts>=? AND ts<?', lower, lower + 86400_000);
      expect(value).toBe(item.value);
    } else if (item.route === '/cognition/focus' && ['attention', 'momentum', 'confidence'].includes(item.metric) && item.topic) {
      value = count(`SELECT ${item.metric} AS count FROM attention_engine_nodes WHERE topic=?`, item.topic);
      expect(value).toBe(item.value);
    } else if (item.route === '/cognition/agents' && ['today', 'days7', 'days30'].includes(item.metric) && item.source) {
      const lower = start - ({ today: 0, days7: 6, days30: 29 }[item.metric as 'today' | 'days7' | 'days30']) * 86400_000;
      value = count('SELECT COUNT(*) AS count FROM attention_engine_events WHERE source=? AND ts>=? AND ts<?', item.source, lower, end);
    } else if (item.route === '/cognition/evidence' && item.metric === 'total') {
      value = count('SELECT COUNT(*) AS count FROM attention_engine_events');
      expect(value).toBe(item.value);
    } else throw new Error('独立 SQL 预期包含不支持的真实指标');
    return { ...item, value, original_value: item.value };
  });
}

test('真实公开来源快照：四个视图只读交互与十个独立SQL值逐项比较', async ({ page }, info) => {
  test.setTimeout(120_000);
  const expectedPath = process.env.ASTROCYTE_TEST_COGNITION_SQL_EXPECTED_FILE;
  expect(expectedPath, '需要Q从本次公开来源快照独立导出的SQL预期').toBeTruthy();
  expect(expectedPath).toMatch(/[\\/]\.verification[\\/]/);
  const expected = JSON.parse(await readFile(expectedPath!, 'utf8')) as SQLExpected;
  expect(expected.values).toHaveLength(10);
  const localAppData = process.env.ASTROCYTE_TEST_COGNITION_LOCALAPPDATA;
  const outputRoot = process.env.ASTROCYTE_TEST_OUTPUT_ROOT;
  if (!localAppData || !outputRoot) throw new Error('需要明确的 Q 公开快照目录和本次验证输出目录');
  const snapshotRoot = resolve(localAppData, 'astrocyte-workbench/v0.1');
  const pointer = JSON.parse(await readFile(resolve(snapshotRoot, 'current.json'), 'utf8')) as { generation: string; snapshot_at: string };
  expect(pointer.generation).toBe('snapshot-1711995690');
  expect(Date.parse(pointer.snapshot_at)).toBe(Date.parse(expected.snapshot_at));
  const server = await cognitionTestServer(true);
  const database = new DatabaseSync(resolve(snapshotRoot, 'snapshots', pointer.generation, 'snapshot.sqlite'), { readOnly: true });
  try {
    const api = await humanAPI(server.apiURL);
    const health = await api.get('/cognition/health') as CognitionHealth;
    expect(Date.parse(health.snapshot_at)).toBe(Date.parse(expected.snapshot_at));
    expect(health.timezone).toBe('Asia/Shanghai');
    const actualDay = new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
    expect((await api.get('/cognition/today') as { day: string }).day).toBe(actualDay);
    const actual = actualSQLValues(database, expected, actualDay);
    const writes: string[] = [], errors: string[] = [], values: (SQLValue & { text: string })[] = [];
    page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') writes.push(request.method()); });
    page.on('pageerror', error => errors.push(error.name));
    for (const item of actual) {
      expect(['/cognition/today', '/cognition/focus', '/cognition/agents', '/cognition/evidence']).toContain(item.route);
      expect(Object.entries(item.query).every(([key, value]) => value === '' || key === 'page' && value === '1' || key === 'page_size' && value === '20')).toBe(true);
      await page.goto(`${server.webURL}${item.route}`);
      await expect(page.getByTestId('snapshot-time')).toBeVisible();
      let id: string;
      if (item.route === '/cognition/today' && item.metric === 'daily_count') id = `daily-${item.day}`;
      else if (item.route === '/cognition/focus' && ['attention', 'momentum', 'confidence'].includes(item.metric)) id = `focus-${{ attention: 'a', momentum: 'm', confidence: 'c' }[item.metric as 'attention' | 'momentum' | 'confidence']}-${item.topic}`;
      else if (item.route === '/cognition/agents' && ['today', 'days7', 'days30'].includes(item.metric)) id = `agent-${item.source}-${item.metric}`;
      else if (item.route === '/cognition/evidence' && item.metric === 'total') id = 'events-total';
      else throw new Error('独立SQL预期包含尚无显示入口的指标');
      const locator = page.getByTestId(id);
      await expect(locator).toBeVisible();
      expect(Number(await locator.getAttribute('data-value'))).toBe(item.value);
      const text = (await locator.innerText()).trim();
      expect(text).toBe(formatNumber(item.value));
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      values.push({ ...item, text });
    }
    expect(new Set(values.map(value => value.route)).size).toBe(4);
    expect((await api.get('/cognition/today') as { day: string }).day).toBe(actualDay);
    expect((await api.get('/cognition/health') as CognitionHealth).snapshot_at).toBe(health.snapshot_at);
    expect(writes).toEqual([]);
    expect(errors).toEqual([]);
    const evidenceRoot = resolve(outputRoot, 'cognition-ui');
    await mkdir(evidenceRoot, { recursive: true });
    await writeFile(resolve(evidenceRoot, `${info.project.name}-sql-values.json`), JSON.stringify({ snapshot_at: health.snapshot_at, timezone: health.timezone, day: actualDay, original_sql_day: expected.day, original_values: expected.values, values, sql_comparisons: values.length, maximum_absolute_error: 0, read_only: true }, null, 2));
  } finally { database.close(); await server.close(); }
});
