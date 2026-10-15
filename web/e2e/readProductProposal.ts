import { chromium, expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
const records = process.env.ASTROCYTE_TEST_PRODUCT_RECORDS_DIR;
const source = process.env.ASTROCYTE_TEST_PRODUCT_SOURCE;
if (origin !== 'http://127.0.0.1:64505' || !output || !records || resolve(records) === resolve(output) || !source || !/^[0-9a-f]{40}$/.test(source)) throw new Error('需要同一产品入口、分开的原记录和输出目录及H归还的准确服务SOURCE');
const directory = resolve(output);
const original = JSON.parse(await readFile(resolve(records, 'product-proposal-human.json'), 'utf8')) as { project_id: string; material_id: string; goal_id: string; proposal_id: string; grant_id: string; context_id: string; context_status: number };
expect(original.context_status).toBe(200);
const api = await humanAPI(origin);
const { proposal } = await api.get(`/proposals/${original.proposal_id}`) as S['S2ProposalResultV1'];
const { approval } = await api.get(`/approvals/${original.grant_id}`) as S['S2ApprovalResultV1'];
const { context_packet: packet } = await api.get(`/context-packets/${original.context_id}`) as S['S2ContextPacketResultV1'];
const goal = (await api.get('/goals') as S['S2GoalListV1']).items.find(item => item.id === original.goal_id);
expect(goal?.id).toBe(original.goal_id);
expect(proposal.status).toBe('approved');
expect(proposal.grant_id).toBe(original.grant_id);
expect(approval.scope.actions).toEqual(['read_context']);
expect(approval.scope.tools).toEqual([]);
expect(packet.project_id).toBe(original.project_id);
expect(packet.grant_id).toBe(original.grant_id);
expect(packet.goal.id).toBe(original.goal_id);
expect(packet.sources.map(item => item.reference)).toEqual([{ material_id: original.material_id, revision: 1 }]);
expect(packet.snapshot.head).toMatch(/^[0-9a-f]{40}$/);
const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
const events: { event_id: string; event_type: string; aggregate_id: string }[] = [];
try {
  for (const [table, id, expected] of [
    ['workspace_proposals', proposal.id, proposal], ['workspace_approvals', approval.id, approval], ['workspace_context_packets', packet.id, packet],
  ] as const) {
    const row = database.prepare(`SELECT data FROM ${table} WHERE id=?`).get(id) as { data: string };
    expect(JSON.parse(row.data)).toEqual(expected);
  }
  for (const [id, type] of [[proposal.id, 'workspace.proposal_approved'], [packet.id, 'workspace.context_packet_built']]) {
    const rows = database.prepare('SELECT event_id,event_type,aggregate_id FROM foundation_outbox WHERE aggregate_id=? AND event_type=?').all(id, type) as typeof events;
    expect(rows).toHaveLength(1);
    for (const event of rows) expect(database.prepare('SELECT event_type,aggregate_id FROM foundation_outbox WHERE event_id=?').get(event.event_id)).toEqual({ event_type: type, aggregate_id: id });
    events.push(...rows);
  }
  const approved = database.prepare('SELECT payload FROM foundation_outbox WHERE event_id=?').get(events[0].event_id) as { payload: string };
  expect((JSON.parse(approved.payload) as { grant_id: string }).grant_id).toBe(approval.id);
  expect(database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions').get()).toEqual({ count: 0 });
} finally { database.close(); }
const browser = await chromium.launch();
const commands: string[] = [];
try {
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  page.on('request', request => {
    if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') commands.push(request.method());
  });
  await page.goto(`${origin}/workspace?tab=proposals`);
  await page.getByRole('button', { name: `查看提案 ${proposal.content.title}`, exact: true }).click();
  const dialog = page.getByRole('dialog', { name: '提案详情', exact: true });
  const panel = dialog.getByRole('region', { name: '批准范围内的固定上下文', exact: true });
  await panel.getByLabel('读取固定上下文包 ID', { exact: true }).fill(packet.id);
  await panel.getByRole('button', { name: '读取指定上下文包', exact: true }).click();
  const reference = panel.getByRole('region', { name: `固定上下文包 ${packet.id}`, exact: true });
  await reference.getByRole('button', { name: '展开固定上下文', exact: true }).click();
  await expect(reference).toContainText(packet.snapshot.head);
  await expect(reference).toContainText(packet.sources[0].object_ref);
  await expect(reference).toContainText(original.grant_id);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  expect(commands).toEqual([]);
  const result = { source, goal_id: original.goal_id, proposal_id: proposal.id, grant_id: approval.id, context_id: packet.id, events, native_sessions: 0, commands, completed_at: new Date().toISOString() };
  await writeFile(resolve(directory, 'product-proposal-readonly.json'), JSON.stringify(result, null, 2), { flag: 'wx' });
  console.log(JSON.stringify(result));
} finally { await browser.close(); }
