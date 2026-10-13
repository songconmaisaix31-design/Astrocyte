/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { realpath } from 'node:fs/promises';
import { join, relative, isAbsolute } from 'node:path';

test('real scanner reads only the explicitly owned public sample and saves the handbook', async ({ page }) => {
  test.skip(process.env.ASTROCYTE_TEST_RELAY_PUBLIC_SCAN !== '1', 'Requires the owned public sample and an already-ready isolated server with empty native metadata roots.');
  const origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL;
  const sample = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_ROOT;
  expect(origin).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  expect(sample).toBeTruthy();
  const evidenceRoot = await realpath(join(process.env.LOCALAPPDATA!, 'astrocyte-workbench/v0.2-access'));
  const sampleRoot = await realpath(sample!);
  const rel = relative(evidenceRoot, sampleRoot);
  expect(rel !== '' && !rel.startsWith('..') && !isAbsolute(rel)).toBe(true);
  const scans: string[][] = [], errors: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    if (request.method() === 'POST' && new URL(request.url()).pathname.endsWith('/works/refresh')) scans.push(request.postDataJSON().roots as string[]);
  });
  await page.goto(`${origin}/workspace/works`);
  const album = page.getByRole('region', { name: '作品寻回册', exact: true });
  await album.getByRole('button', { name: '扫描导入项目', exact: true }).click();
  let dialog = page.getByRole('dialog');
  await dialog.locator('label').filter({ hasText: sampleRoot }).getByRole('checkbox').check();
  await dialog.getByRole('button', { name: '扫描所选项目', exact: true }).click();
  await expect(dialog.getByRole('status')).toContainText('已保存文件观察');
  await page.keyboard.press('Escape');
  await expect(album.getByLabel('作品统计').locator('dd')).toHaveText(['1', '1', '1', '未知', '1']);
  const card = album.getByRole('button', { name: '查看作品 Relay public scan sample', exact: true });
  await expect(card).toContainText('codex');
  await expect(card).toContainText('从文件核实');
  await expect(card).toContainText('2 个文件');
  await album.getByLabel('平台', { exact: true }).selectOption('codex');
  await card.click();
  dialog = page.getByRole('dialog');
  await dialog.getByLabel('编辑交接手册').fill('Public sample handoff: inspect README only; no Agent launch requested.');
  await dialog.getByRole('button', { name: '保存交接手册', exact: true }).click();
  await expect(dialog.getByRole('status')).toContainText('人工记录已保存');
  await page.keyboard.press('Escape');
  await page.reload();
  await expect(card).toContainText('Public sample handoff');
  expect(scans).toEqual([[sampleRoot]]);
  expect(errors).toEqual([]);
});
