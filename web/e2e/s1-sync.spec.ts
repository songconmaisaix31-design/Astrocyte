import { test, expect } from '@playwright/test';

test('account platform placeholders never pretend to be connected', async ({ page }) => {
  await page.goto('/attention');
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '账号与更新清单', exact: true }) }).first();
  await expect(panel).toContainText('已绑定账号');
  await expect(panel).toContainText('未接入');
  await expect(panel).toContainText('绑定方式待确定');
  await expect(panel.getByRole('button')).toHaveCount(0);
  await expect(panel).toContainText('入库由人工选择');
});

test('local overview keeps real empty facts explicit at both sizes', async ({ page }) => {
  await page.goto('/workspace');
  await expect(page.getByRole('status', { name: '暂无本地项目' })).toBeVisible();
  const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
  await expect(panel).toContainText('未提供');
  await expect(panel.getByRole('combobox', { name: '项目排序' }).locator('option[value="activity"]')).toBeDisabled();
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
