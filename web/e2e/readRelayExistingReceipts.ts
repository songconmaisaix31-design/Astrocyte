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
const previous = JSON.parse(await readFile(resolve(directory, 'existing-transition-ui-safe.json'), 'utf8')) as { task_id: string; observed_at: string; actions: { operation_id: string }[] };
expect(previous.task_id).toBe(taskID);
const operations = new Set(previous.actions.map(action => action.operation_id));
const context = await request.newContext({ baseURL: origin });
try {
  expect((await context.get('/api/v1/auth/session')).status()).toBe(200);
  const task = await context.get(`/api/v1/relay/tasks/${taskID}`);
  expect(task.status()).toBe(200);
  const { view } = await task.json() as S['RelayTaskResultV1'];
  const response = await context.get(`/api/v1/relay/tasks/${taskID}/receipts`);
  expect(response.status()).toBe(200);
  const { receipts } = await response.json() as S['RelayReceiptsV1'];
  const selected = receipts.filter(receipt => operations.has(receipt.operation_id)).map(receipt => ({ action: receipt.action, operation_id: receipt.operation_id, status: receipt.status, created_at: receipt.created_at, finished_at: receipt.finished_at, error_code: receipt.error?.code, required_action: receipt.error?.required_action }));
  expect(selected.map(receipt => receipt.operation_id).sort()).toEqual([...operations].sort());
  await writeFile(resolve(directory, 'existing-transition-current-derived.json'), JSON.stringify({ task_id: taskID, observed_at: new Date().toISOString(), first_ui_started_at: previous.observed_at, state: view.task.state, condition: view.task.condition, revision: view.task.project_settings_revision, liveness: view.terminal?.liveness, exited_at: view.terminal?.exited_at, transitions: view.transitions.map(item => ({ from: item.from, to: item.to, ts: item.ts })), receipts: selected, pending_count: receipts.filter(receipt => receipt.status === 'pending').length }, null, 2), { flag: 'wx' });
  console.log('F_EXISTING_GET_OBSERVED: 实际GET状态、准确迁移和收据时刻已保存；没有原生read或其他写入。');
} finally { await context.dispose(); }
