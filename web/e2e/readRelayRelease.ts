/// <reference types="node" />
import { request, expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL;
const taskID = process.env.ASTROCYTE_TEST_RELAY_CURRENT_TASK_ID;
if (!origin || !/^http:\/\/127\.0\.0\.1:\d+$/.test(origin) || !taskID) throw new Error('需要准确公开任务和当前本机HTTP入口');
const directory = resolve('../.verification/l1-live');
const original = JSON.parse(await readFile(resolve(directory, 'existing-release-ui-safe.json'), 'utf8')) as { task_id: string; operation_id: string; status: number };
expect(original.task_id).toBe(taskID);
expect(original.status).toBe(200);
const context = await request.newContext({ baseURL: origin });
try {
  expect((await context.get('/api/v1/auth/session')).status()).toBe(200);
  const task = await context.get(`/api/v1/relay/tasks/${taskID}`);
  expect(task.status()).toBe(200);
  const { view } = await task.json() as S['RelayTaskResultV1'];
  const response = await context.get(`/api/v1/relay/tasks/${taskID}/receipts`);
  expect(response.status()).toBe(200);
  const { receipts } = await response.json() as S['RelayReceiptsV1'];
  const release = receipts.find(receipt => receipt.operation_id === original.operation_id);
  expect(release?.status).toBe('accepted');
  expect(view.task.state).toBe('released');
  expect(view.worktree?.active_task_id).toBe('');
  expect(view.worktree?.prepared).toBe(true);
  expect(receipts.find(receipt => receipt.operation_id === '5f99133e-390f-4987-b8f2-2889716921ff')?.status).toBe('delivery_unknown');
  expect(receipts.filter(receipt => receipt.status === 'pending')).toHaveLength(0);
  await writeFile(resolve(directory, 'existing-release-current-derived.json'), JSON.stringify({ task_id: taskID, observed_at: new Date().toISOString(), state: view.task.state, condition: view.task.condition, revision: view.task.project_settings_revision, liveness: view.terminal?.liveness, exited_at: view.terminal?.exited_at, prepared: view.worktree?.prepared, active_task_id: view.worktree?.active_task_id, transitions: view.transitions.map(item => ({ from: item.from, to: item.to, ts: item.ts })), receipts: receipts.filter(item => item.operation_id === original.operation_id || item.operation_id === '5f99133e-390f-4987-b8f2-2889716921ff').map(item => ({ operation_id: item.operation_id, action: item.action, status: item.status, created_at: item.created_at, finished_at: item.finished_at, error_code: item.error?.code, required_action: item.error?.required_action })), pending_count: 0, goal_acceptance: 'failed_before_goal_check', source_identity_gate: 'failed_nlink2' }, null, 2), { flag: 'wx' });
  console.log('F_RELEASE_GET_OBSERVED: 实际released、空占用、prepared目录和原read未知通过只读GET核实。');
} finally { await context.dispose(); }
