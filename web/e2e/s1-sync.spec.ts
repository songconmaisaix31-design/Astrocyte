import { test, expect } from '@playwright/test';
import type { components } from '../src/api/schema';

test('local overview keeps real empty facts explicit at both sizes', async ({ page }) => {
  await page.goto('/workspace');
  await expect(page.getByRole('status', { name: '暂无本地项目' })).toBeVisible();
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
  await expect(panel).toContainText('未提供');
  await expect(panel.getByRole('combobox', { name: '项目排序' }).locator('option[value="activity"]')).toHaveJSProperty('disabled', true);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await panel.screenshot({ path: `test-results/s1-local-empty-${page.viewportSize()?.width}.png`, animations: 'disabled' });
});

test('fixture local cards filter by actual recorded association and preserve details', async ({ page }) => {
  await page.goto('/workspace?fixture=1');
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
  await expect(panel).toContainText('已加载项目');
  await expect(panel).toContainText('来源目录');
  await expect(panel).toContainText('交接手册 · 未提供');
  await panel.screenshot({ path: `test-results/s1-local-cards-${page.viewportSize()?.width}.png`, animations: 'disabled' });
  await panel.getByRole('group', { name: '平台', exact: true }).getByRole('button', { name: '未关联平台', exact: true }).click();
  await expect(panel).toContainText('示例项目丙');
  await expect(panel.getByRole('button', { name: /Fixture Project Beta.*查看项目详情/ })).toHaveCount(0);
  await panel.getByRole('group', { name: '平台', exact: true }).getByRole('button', { name: '全部', exact: true }).click();
  await panel.getByRole('combobox', { name: '项目分组' }).selectOption('unknown');
  await panel.getByRole('combobox', { name: '项目排序' }).selectOption('name');
  await panel.getByRole('button', { name: /Fixture Project Beta.*查看项目详情/ }).click();
  await expect(page.getByRole('dialog')).toContainText('/fixture/projects/beta');
  await expect(page.getByRole('dialog').getByRole('button', { name: '编辑项目' })).toBeDisabled();
});

test('real local inventory displays saved independent observations and refreshes without commands', async ({ page }) => {
  const commands: string[] = [];
  page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) commands.push(`${request.method()} ${request.url()}`); });
  const responsePromise = page.waitForResponse(response => response.url().endsWith('/api/v1/local-agents'));
  await page.goto('/workspace');
  const response = await responsePromise;
  expect(response.status()).toBe(200);
  const payload = await response.json() as components['schemas']['LocalAgentListV1'];
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本机 Agent 清单', exact: true }) }).first();
  await expect(panel.locator('li')).toHaveCount(payload.items.length);
  for (const agent of payload.items) {
    const card = panel.locator('li').filter({ has: page.getByRole('heading', { name: agent.display_name, exact: true }) });
    await expect(card).toContainText(agent.version ?? '未提供');
    const readinessReasons = await card.locator(':scope > dl dd > small[title]').evaluateAll(elements => elements.map(element => element.getAttribute('title')));
    for (const observation of [agent.installed, agent.configured, agent.startable]) {
      expect(readinessReasons).toContain(observation.reason);
      if (observation.status === 'unknown') await expect(card).toContainText('未知');
    }
    const native = card.locator('details').filter({ has: page.locator('summary').filter({ hasText: /^原生能力/ }) }).first();
    await native.locator(':scope > summary').click();
    await expect(native.locator('dl > div')).toHaveCount(8);
    const nativeReasons = await native.locator('dd > small[title]').evaluateAll(elements => elements.map(element => element.getAttribute('title')));
    for (const observation of Object.values(agent.capabilities)) expect(nativeReasons).toContain(observation.reason);
  }
  const refresh = page.waitForResponse(response => response.url().endsWith('/api/v1/local-agents'));
  await panel.getByRole('button', { name: '刷新清单', exact: true }).click();
  expect((await refresh).status()).toBe(200);
  expect(commands).toEqual([]);
  await panel.screenshot({ path: `test-results/s1-local-inventory-${page.viewportSize()?.width}.png`, animations: 'disabled' });
});
