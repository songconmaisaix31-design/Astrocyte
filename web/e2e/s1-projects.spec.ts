/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import type { components } from '../src/api/schema';

// Actual temporary project/API/SQLite operations, no synthetic sessions or page.route.
// This tests project registration/scoping; public-material AT03 is separate.
test('human registers a specified project, independently saves B/C, reads only approved files and persists across restart', async ({ page }, testInfo) => {
  test.setTimeout(120_000);
  const server = await startS1Server({ browser: true });
  const projectRoot = join(server.temporary, 'approved-public-project');
  const filename = 'README.md';
  const text = 'Astrocyte temporary public project for approved-context UI verification.\n';
  await mkdir(projectRoot);
  await writeFile(join(projectRoot, filename), text);
  try {
    await page.goto(`${server.webURL}/workspace`);
    await page.getByRole('button', { name: '接入项目', exact: true }).click();
    const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
    await panel.getByText('登记项目 / 发现子项目', { exact: true }).click();
    await panel.getByLabel('新项目顶层空间名称', { exact: true }).fill('临时公开项目空间');
    await panel.getByRole('button', { name: '创建空间用于此项目', exact: true }).click();
    await expect(panel.getByLabel('关联项目顶层空间', { exact: true })).not.toHaveValue('');
    await panel.getByLabel('项目名称', { exact: true }).fill('临时公开项目');
    await panel.getByLabel('项目绝对目录', { exact: true }).fill(projectRoot);
    const registerResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/local-projects') && response.request().method() === 'POST');
    await panel.getByRole('button', { name: '登记此项目', exact: true }).click();
    const registered = await registerResponse;
    expect(registered.status()).toBe(200);
    const receipt = await registered.json() as components['schemas']['LocalProjectResultV1'];
    expect(receipt.project.root).toBe(projectRoot);
    expect(receipt.project.settings).toMatchObject({ allow_directory: false, expand_references: false, external_model_cli: '', allow_agent_control: false });
    await expect(panel.getByRole('heading', { name: '临时公开项目', exact: true }).last()).toBeVisible();
    const manage = panel.getByRole('region', { name: '已登记项目管理' });
    await manage.getByText('项目小权限与模型处理许可', { exact: true }).click();
    const toggleB = manage.getByRole('checkbox', { name: 'B：允许读取明确目录 / 文件' });
    const toggleC = manage.getByRole('checkbox', { name: 'C：允许展开获准资料的引用' });
    await expect(toggleB).not.toBeChecked();
    await expect(toggleC).not.toBeChecked();
    await toggleB.check();
    await expect(toggleC).not.toBeChecked();
    await manage.getByLabel('允许的项目子目录 / 文件（每行一项）', { exact: true }).fill(filename);
    const settingsResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${receipt.project.id}/settings`) && response.request().method() === 'PUT');
    await manage.getByRole('button', { name: '保存此项目许可', exact: true }).click();
    const settings = await settingsResponse;
    expect(settings.status()).toBe(200);
    const saved = await settings.json() as components['schemas']['LocalProjectResultV1'];
    expect(saved.project.settings).toMatchObject({ allow_directory: true, expand_references: false, allowed_subdirs: [filename], allow_agent_control: false });
    const native = manage.getByRole('region', { name: '项目原生 Agent 操作' });
    await native.getByLabel('本次读取的获准文件（每行一项）', { exact: true }).fill(filename);
    const contextResponse = page.waitForResponse(response => response.url().endsWith(`/local-projects/${receipt.project.id}/context`) && response.request().method() === 'POST');
    await native.getByRole('button', { name: '预览本次获准上下文', exact: true }).click();
    const contextResult = await contextResponse;
    expect(contextResult.status()).toBe(200);
    const context = await contextResult.json() as components['schemas']['LocalContextPacketV1'];
    expect(context.files).toHaveLength(1);
    expect(context.files).toMatchObject([{ path: filename, text }]);
    expect(context.materials).toEqual([]);
    await native.getByText(/^实际上下文 ·/).click();
    await expect(native).toContainText(text.trim());
    await expect(native.getByRole('button', { name: '启动获准原生会话', exact: true })).toBeDisabled();
    const before = server.query('SELECT data FROM local_agent_projects WHERE id=?', receipt.project.id);
    expect(before).toHaveLength(1);
    const stored = JSON.parse(String(before[0].data)) as components['schemas']['LocalProjectV1'];
    expect(stored.settings).toEqual(saved.project.settings);
    await testInfo.attach('actual-project-settings', { body: JSON.stringify(saved.project, null, 2), contentType: 'application/json' });
    await server.restart();
    await page.reload();
    await page.getByRole('button', { name: '接入项目', exact: true }).click();
    await panel.getByRole('button', { name: /临时公开项目.*查看项目与权限/ }).click();
    await manage.getByText('项目小权限与模型处理许可', { exact: true }).click();
    await expect(toggleB).toBeChecked();
    await expect(toggleC).not.toBeChecked();
    await expect(manage.getByLabel('允许的项目子目录 / 文件（每行一项）', { exact: true })).toHaveValue(filename);
    expect(server.query('SELECT data FROM local_agent_projects WHERE id=?', receipt.project.id)).toEqual(before);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')[0].count).toBe(0);
    for (const width of [1280, 1920]) {
      await page.setViewportSize({ width, height: width === 1280 ? 720 : 1080 });
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await panel.screenshot({ path: testInfo.outputPath(`actual-project-${width}.png`) });
    }
  } finally { await server.close(); }
});
