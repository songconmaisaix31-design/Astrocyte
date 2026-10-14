/// <reference types="node" />
import { chromium, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL;
const taskID = process.env.ASTROCYTE_TEST_RELAY_CURRENT_TASK_ID;
if (!origin || !/^http:\/\/127\.0\.0\.1:\d+$/.test(origin) || !taskID || process.env.ASTROCYTE_TEST_RELAY_MANUAL_TRANSITION_ONLY !== '1') throw new Error('需要明确获准的旧公开任务、当前URL与人工迁移功能验证范围');
if (taskID !== '0c6522bc-b232-4040-8407-745885644791') throw new Error('固定历史任务脚本禁止用于其他任务');
const directory = resolve('../.verification/l1-live');
await mkdir(directory, { recursive: true });
const path = resolve(directory, 'existing-transition-ui-safe.json');
const actions: { action: string; operation_id: string; status: number; receipt_status: string; state?: string }[] = [];
const evidence = { task_id: taskID, authorization: 'M-relay msg_dc22a69bafd0', model_content_goal: 'unknown', source_identity_gate: 'failed_nlink2', actions, stage: 'before_actions', observed_at: new Date().toISOString() };
await writeFile(path, JSON.stringify(evidence, null, 2), { flag: 'wx' });
const save = () => writeFile(path, JSON.stringify(evidence, null, 2));
const browser = await chromium.launch();
try {
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();
  expect((await page.request.get(`${origin}/api/v1/auth/session`)).status()).toBe(200);
  await page.goto(`${origin}/workspace/works`);
  const readTask = async () => {
    const response = await page.request.get(`${origin}/api/v1/relay/tasks/${taskID}`);
    expect(response.status()).toBe(200);
    return (await response.json() as S['RelayTaskResultV1']).view;
  };
  const readReceipt = async (operationID: string) => {
    const response = await page.request.get(`${origin}/api/v1/relay/tasks/${taskID}/receipts`);
    expect(response.status()).toBe(200);
    const { receipts } = await response.json() as S['RelayReceiptsV1'];
    const receipt = receipts.find(item => item.operation_id === operationID);
    expect(receipt).toBeDefined();
    return receipt!;
  };
  const before = await readTask();
  expect(before.task.state).toBe('working');
  expect(before.task.project_settings_revision).toBe(2);
  expect(before.terminal?.liveness).toBe('exited');
  expect(before.terminal?.exited_at).toBeTruthy();
  await page.locator('.relay-tasks > li').filter({ hasText: taskID }).getByRole('button', { name: '查看任务、事件和收据', exact: true }).click();
  const dialog = page.getByRole('dialog');
  const controls = dialog.getByRole('region', { name: '人工完成与交接' });
  await expect(controls.getByRole('checkbox')).toBeEnabled();
  await expect(dialog.getByRole('button', { name: '核对事件接入', exact: true })).toBeDisabled();
  await controls.getByRole('checkbox').check();
  await controls.getByRole('button', { name: '取消确认', exact: true }).click();
  await expect(controls.getByRole('checkbox')).not.toBeChecked();

  for (const next of ['done', 'released'] as const) {
    evidence.stage = `before_${next}`;
    await save();
    await controls.getByRole('checkbox').check();
    const transitioned = page.waitForResponse(response => response.url().endsWith(`/tasks/${taskID}/transition`) && response.request().method() === 'POST');
    await controls.getByRole('button', { name: next === 'done' ? '确认任务完成' : '确认释放工作区', exact: true }).click();
    const response = await transitioned;
    const body = await response.json() as S['RelayTaskResultV1'];
    const operationID = response.request().headers()['idempotency-key'];
    actions.push({ action: `transition_${next}`, operation_id: operationID, status: response.status(), receipt_status: 'not_read', state: body.view?.task.state });
    await save();
    expect(response.status()).toBe(200);
    const receipt = await readReceipt(operationID);
    actions.at(-1)!.receipt_status = receipt.status;
    await save();
    expect(receipt.status).toBe('accepted');
    expect(body.view.task.state).toBe(next);
    expect((await readTask()).task.state).toBe(next);
    await expect(dialog).toContainText(next === 'done' ? '任务已完成 · done' : '已释放 · released');
    if (next === 'done') {
      await expect(dialog.getByRole('button', { name: '发送输入', exact: true })).toBeDisabled();
      await expect(dialog.getByRole('button', { name: '停止 Agent', exact: true })).toBeDisabled();
      evidence.stage = 'before_read';
      await save();
      const reading = page.waitForResponse(response => response.url().endsWith(`/tasks/${taskID}/read`) && response.request().method() === 'POST');
      await dialog.getByRole('button', { name: '读取会话并留收据', exact: true }).click();
      const read = await reading;
      const result = await read.json() as S['RelayActionResultV1'];
      actions.push({ action: 'read', operation_id: read.request().headers()['idempotency-key'], status: read.status(), receipt_status: result.receipt?.status || 'missing' });
      await save();
      expect(read.status()).toBe(200);
      expect(result.receipt.status).toBe('accepted');
    }
  }
  evidence.stage = 'released_observed';
  await save();
  console.log('F_MANUAL_TRANSITION_OBSERVED: 实际done/read/released已记录，原模型内容目标保持unknown。');
} catch (cause) {
  evidence.stage = `failed_${evidence.stage}`;
  await save();
  throw cause;
} finally { await browser.close(); }
