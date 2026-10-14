/// <reference types="node" />
import { chromium, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import type { components } from '../src/api/schema';

const origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL;
const requestedAgent = process.env.ASTROCYTE_TEST_RELAY_L1_AGENT;
if (!origin || !/^http:\/\/127\.0\.0\.1:\d+$/.test(origin)) throw new Error('需要H交付实际页面URL');
const browser = await chromium.launch();
try {
  const viewports = [{ name: 'chromium-1920', width: 1920, height: 1080 }, { name: 'chromium-1280', width: 1280, height: 720 }];
  const contexts = await Promise.all(viewports.map(viewport => browser.newContext({ viewport: { width: viewport.width, height: viewport.height } })));
  const pages = await Promise.all(contexts.map(context => context.newPage()));
  await Promise.all(pages.map(async page => {
    const session = await page.request.get(`${origin}/api/v1/auth/session`);
    expect(session.status()).toBe(200);
  }));
  await Promise.all(pages.map(page => page.goto(`${origin}/cognition/today`)));
  const registry = await pages[0].request.get(`${origin}/api/v1/relay/agents`);
  expect(registry.status()).toBe(200);
  const registryBefore = (await registry.json() as components['schemas']['RelayAgentsV1']).agents.map(item => item.name);
  expect(registryBefore).toEqual([]);
  if (requestedAgent) expect(registryBefore).not.toContain(requestedAgent);
  console.log('F_BROWSER_READY: 两个原viewport已打开实际Today，等待H注册并派发获准公开任务。');
  const observations = await Promise.all(pages.map(async (page, index) => {
    let taskID = '', observedAgent = '', eventIDs: string[] = [];
    const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
    await expect.poll(async () => {
      const tasksResponse = await page.request.get(`${origin}/api/v1/relay/tasks`);
      expect(tasksResponse.status()).toBe(200);
      const tasks = await tasksResponse.json() as components['schemas']['RelayTasksV1'];
      const task = tasks.tasks.find(item => requestedAgent ? item.task.agent === requestedAgent : !registryBefore.includes(item.task.agent));
      if (!task) return false;
      taskID = task.task.task_id;
      observedAgent = task.task.agent;
      await region.getByLabel('会话任务 ID', { exact: true }).fill(taskID);
      await region.getByLabel('会话每页条数', { exact: true }).selectOption('100');
      await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
      const rows = region.locator('tbody tr');
      if (!await rows.count()) return false;
      eventIDs = await rows.locator('td:first-child').allTextContents();
      return eventIDs.length > 0;
    }, { timeout: 180_000, intervals: [1000, 2000, 5000] }).toBe(true);
    return { name: viewports[index].name, first_visible_at: new Date().toISOString(), event_ids: eventIDs, task_id: taskID, agent_name: observedAgent };
  }));
  const directory = resolve('../.verification/l1-live');
  await mkdir(directory, { recursive: true });
  expect(observations[0].agent_name).toBe(observations[1].agent_name);
  await writeFile(resolve(directory, 'first-visible.json'), JSON.stringify({ origin, agent_name: observations[0].agent_name, registry_before: registryBefore, viewports: observations }, null, 2));
  console.log('F_BROWSER_FIRST_VISIBLE: 两个原viewport实际首次可见记录已保存到.verification/l1-live/first-visible.json。');
} finally { await browser.close(); }
