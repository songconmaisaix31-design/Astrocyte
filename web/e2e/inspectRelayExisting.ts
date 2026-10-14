/// <reference types="node" />
import { chromium, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL;
const taskID = process.env.ASTROCYTE_TEST_RELAY_CURRENT_TASK_ID;
if (!origin || !/^http:\/\/127\.0\.0\.1:\d+$/.test(origin) || !taskID) throw new Error('需要H当前固定公开task/URL');
const browser = await chromium.launch();
try {
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();
  expect((await page.request.get(`${origin}/api/v1/auth/session`)).status()).toBe(200);
  await page.goto(`${origin}/workspace/works`);
  const response = await page.request.get(`${origin}/api/v1/relay/tasks/${taskID}`);
  expect(response.status()).toBe(200);
  const view = (await response.json() as S['RelayTaskResultV1']).view;
  const projectsResponse = await page.request.get(`${origin}/api/v1/relay/projects`);
  expect(projectsResponse.status()).toBe(200);
  const permission = (await projectsResponse.json() as S['RelayProjectsV1']).projects.find(project => project.project_id === view.task.project_id);
  expect(permission).toBeDefined();
  await page.locator('.relay-tasks > li').filter({ hasText: taskID }).getByRole('button', { name: '查看任务、事件和收据', exact: true }).click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toContainText(taskID);
  const controls = dialog.getByRole('region', { name: '人工完成与交接' });
  const before = { state: view.task.state, condition: view.task.condition, liveness: view.terminal?.liveness, allowed_actions: permission!.allowed_actions, task_revision: view.task.project_settings_revision, current_revision: permission!.project_settings_revision, read_disabled: await dialog.getByRole('button', { name: '读取会话并留收据' }).isDisabled(), stop_disabled: await dialog.getByRole('button', { name: '停止 Agent', exact: true }).isDisabled(), completion_checkbox_disabled: await controls.getByRole('checkbox').isDisabled() };
  let readStatus: number | null = null;
  if (!before.read_disabled) {
    const read = page.waitForResponse(response => response.url().endsWith('/read') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '读取会话并留收据', exact: true }).click();
    readStatus = (await read).status();
  }
  const receiptsResponse = await page.request.get(`${origin}/api/v1/relay/tasks/${taskID}/receipts`);
  expect(receiptsResponse.status()).toBe(200);
  const receipts = (await receiptsResponse.json() as S['RelayReceiptsV1']).receipts.map(receipt => ({ action: receipt.action, status: receipt.status, error_code: receipt.error?.code, required_action: receipt.error?.required_action, operation_id: receipt.operation_id }));
  const directory = resolve('../.verification/l1-live');
  await mkdir(directory, { recursive: true });
  await writeFile(resolve(directory, 'existing-ui-safe.json'), JSON.stringify({ task_id: taskID, observed_at: new Date().toISOString(), before, read_status: readStatus, receipts }, null, 2));
  console.log('F_EXISTING_UI_OBSERVED: 当前真实权限、禁用状态与一次read已保存；没有confirm/send/resume/重派。');
} finally { await browser.close(); }
