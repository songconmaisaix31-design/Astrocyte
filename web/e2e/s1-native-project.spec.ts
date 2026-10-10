/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { randomUUID } from 'node:crypto';
import process from 'node:process';
import type { components } from '../src/api/schema';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';

type S = components['schemas'];

// Separately assigned after public-material acceptance. At most three short
// actual Codex turns, one approved public file, no tool permission or private roots.
test('human observes, stops, resumes the same actual native session and sends one scoped message', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_REAL_NATIVE_UI !== '1', 'Requires separately declared coordinator-assigned native slot.');
  test.setTimeout(360_000);
  const server = await startS1Server({ browser: true });
  const api = await humanAPI(server.apiURL);
  const root = join(server.temporary, 'approved-native-public-project');
  await mkdir(root);
  await writeFile(join(root, 'README.md'), 'The approved public project marker is Astrocyte-W3-public-native.\n');
  const results: unknown[] = [];
  let completed = false;
  let projectID: string | undefined;
  try {
    await page.goto(`${server.webURL}/workspace`);
    const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
    await panel.getByText('登记项目 / 发现子项目', { exact: true }).click();
    await panel.getByLabel('新项目顶层空间名称', { exact: true }).fill('原生公开文件空间');
    await panel.getByRole('button', { name: '创建空间用于此项目', exact: true }).click();
    await expect(panel.getByLabel('关联项目顶层空间', { exact: true })).not.toHaveValue('');
    await panel.getByLabel('项目名称', { exact: true }).fill('原生公开文件项目');
    await panel.getByLabel('项目绝对目录', { exact: true }).fill(root);
    const registering = page.waitForResponse(r => r.url().endsWith('/local-projects') && r.request().method() === 'POST');
    await panel.getByRole('button', { name: '登记此项目', exact: true }).click();
    const registered = await registering;
    expect(registered.status()).toBe(200);
    const project = (await registered.json() as S['LocalProjectResultV1']).project;
    projectID = project.id;
    const manage = panel.getByRole('region', { name: '已登记项目管理' });
    await manage.getByText('项目小权限与模型处理许可', { exact: true }).click();
    await manage.getByRole('checkbox', { name: 'B：允许读取明确目录 / 文件' }).check();
    await manage.getByLabel('允许的项目子目录 / 文件（每行一项）', { exact: true }).fill('README.md');
    await expect(manage.getByRole('checkbox', { name: 'C：允许展开获准资料的引用' })).not.toBeChecked();
    await manage.getByLabel('此项目允许模型处理的 CLI', { exact: true }).selectOption('codex');
    for (const action of ['读取上下文', '启动', '停止', '原生接续', '发送输入', '观察与日志']) await manage.getByRole('checkbox', { name: action, exact: true }).check();
    const saving = page.waitForResponse(r => r.url().endsWith(`/local-projects/${project.id}/settings`) && r.request().method() === 'PUT');
    await manage.getByRole('button', { name: '保存此项目许可', exact: true }).click();
    const saved = await saving;
    expect(saved.status()).toBe(200);
    const settings = (await saved.json() as S['LocalProjectResultV1']).project.settings;
    expect(settings.allowed_tools).toEqual([]);
    expect(settings.allow_agent_control).toBe(false);
    const native = manage.getByRole('region', { name: '项目原生 Agent 操作' });
    await native.getByLabel('项目操作客户端', { exact: true }).selectOption('codex');
    await native.getByLabel('本次读取的获准文件（每行一项）', { exact: true }).fill('README.md');
    await native.getByLabel('给 Agent 的本次消息', { exact: true }).fill('只根据已交付README原文返回公开项目marker，不使用工具或其他资料，回答仅一行。');
    const starting = page.waitForResponse(r => r.url().endsWith('/sessions/start') && r.request().method() === 'POST');
    await native.getByRole('button', { name: /^(验证并)?启动获准原生会话$/ }).click();
    const started = await starting;
    expect(started.status()).toBe(200);
    const session = (await started.json() as S['NativeSessionResultV1']).session;
    expect(session.ownership).toBe('owned');
    expect(session.native_id).toBeTruthy();
    results.push(session);
    const row = native.locator('li').filter({ hasText: `原生 ID ${session.native_id}` });
    async function observeUI() {
      const observing = page.waitForResponse(r => r.url().endsWith(`/sessions/${session.id}/observe`));
      await row.getByRole('button', { name: /^(验证并)?观察状态与日志$/ }).click();
      const response = await observing;
      expect(response.status()).toBe(200);
      const result = await response.json() as S['NativeObservationResultV1'];
      results.push(result);
      expect(['unknown', 'blocked', 'interrupted']).not.toContain(result.observation.status);
      return result.observation;
    }
    async function settledOutput(minimumEvents: number) {
      let observation: S['NativeObservationV1'] | undefined;
      await expect.poll(async () => {
        observation = (await api.get(`/local-projects/${project.id}/sessions/${session.id}/observe`) as S['NativeObservationResultV1']).observation;
        if (['unknown', 'blocked', 'interrupted'].includes(observation.status)) throw new Error(`Actual native outcome ${observation.status}; no resend allowed`);
        return observation.status === 'completed' && observation.events.length > minimumEvents && observation.events.filter(event => event.kind === 'text').map(event => event.text).join('').includes('Astrocyte-W3-public-native');
      }, { timeout: 120_000, intervals: [500] }).toBe(true);
      return observation!;
    }
    await observeUI(); // First unknown-capability human observe uses actual driver.
    await settledOutput(0);
    await observeUI();
    await expect(row).toContainText('Astrocyte-W3-public-native');
    async function stopUI() {
      const stopping = page.waitForResponse(r => r.url().endsWith(`/sessions/${session.id}/stop`) && r.request().method() === 'POST');
      await row.getByRole('button', { name: /^(验证并)?停止原生会话$/ }).click();
      const response = await stopping;
      expect(response.status()).toBe(200);
      const result = await response.json() as S['NativeObservationResultV1'];
      expect(result.observation.stop_confirmed).toBe(true);
      results.push(result);
      await expect(row).toContainText('原执行停止 已确认');
    }
    await stopUI();
    await native.getByLabel('给 Agent 的本次消息', { exact: true }).fill('原生接续：请沿用此前已交付的同一README，再返回同一marker一行，不使用工具。');
    const resuming = page.waitForResponse(r => r.url().endsWith(`/sessions/${session.id}/resume`) && r.request().method() === 'POST');
    await row.getByRole('button', { name: /^(验证并)?原生接续$/ }).click();
    const resumedResponse = await resuming;
    expect(resumedResponse.status()).toBe(200);
    const resumed = (await resumedResponse.json() as S['NativeSessionResultV1']).session;
    expect(resumed.id).toBe(session.id);
    expect(resumed.native_id).toBe(session.native_id);
    expect(resumed.mode).not.toBe('context_handoff');
    results.push(resumed);
    const second = await settledOutput(0);
    await observeUI();
    await native.getByLabel('给 Agent 的本次消息', { exact: true }).fill('明确第三次输入：只返回同一已交付README的marker，不使用工具或读取其他文件。');
    const sending = page.waitForResponse(r => r.url().endsWith(`/sessions/${session.id}/send`) && r.request().method() === 'POST');
    await row.getByRole('button', { name: /^(验证并)?发送本次消息$/ }).click();
    const sent = await sending;
    expect(sent.status()).toBe(200);
    results.push(await sent.json());
    await settledOutput(second.events.length);
    await observeUI();
    await stopUI();
    expect(server.query('SELECT COUNT(*) AS n FROM local_agent_sessions')[0].n).toBe(1);
    expect(server.query('SELECT COUNT(*) AS n FROM attention_materials')[0].n).toBe(0);
    for (const width of [1280, 1920]) {
      await page.setViewportSize({ width, height: width === 1280 ? 720 : 1080 });
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await native.screenshot({ path: testInfo.outputPath(`actual-native-${width}.png`) });
    }
    completed = true;
  } finally {
    if (!completed && projectID) {
      const actual = await api.get(`/local-projects/${projectID}/sessions`) as S['NativeSessionListV1'];
      for (const session of actual.items.filter(item => item.ownership === 'owned' && !item.stop_confirmed)) {
        // Explicit human cleanup stop, never model resend or history takeover.
        const key = randomUUID();
        const response = await fetch(`${server.apiURL}/api/v1/local-projects/${projectID}/sessions/${session.id}/stop`, { method: 'POST', headers: { Cookie: api.cookie, Origin: server.apiURL, 'Content-Type': 'application/json', 'X-CSRF-Token': api.session.csrf_token, 'Idempotency-Key': key }, body: JSON.stringify({ schema_version: 1, request_id: key, expected_version: 2 }) });
        results.push({ cleanupStopStatus: response.status, receipt: await response.json() as unknown });
      }
    }
    await testInfo.attach('actual-native-receipts', { body: JSON.stringify(results, null, 2), contentType: 'application/json' });
    if (!completed) await testInfo.attach('actual-retained-store-path', { body: JSON.stringify({ temporary: server.temporary, dataDir: server.dataDir }, null, 2), contentType: 'application/json' });
    await server.close({ preserveData: !completed });
  }
});
