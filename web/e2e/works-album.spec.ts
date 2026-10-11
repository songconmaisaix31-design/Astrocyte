import { test, expect } from '@playwright/test';

/**
 * W7 works album ("AI 作品寻回册") e2e. The durable scan/import/attitude/handoff
 * commands are W0-owned HTTP endpoints; until they land this spec exercises the
 * fixture path and asserts the three-viewport / keyboard / no-pageerror / no
 * horizontal-overflow acceptance that does not depend on the endpoint.
 */

const sizes = [
  { width: 1920, height: 1080 },
  { width: 1280, height: 720 },
  { width: 390, height: 844 },
];

test.describe('Works album fixture', () => {
  test('renders platform stats and verified work cards', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('[aria-label="平台统计"]')).toBeVisible();
    await expect(page.getByRole('button', { name: '查看作品 个人知识库' })).toBeVisible();
    await expect(page.locator('text=从文件核实').first()).toBeVisible();
  });

  test('switches views and filters by attitude', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: '按分组' }).click();
    await expect(page.getByRole('heading', { name: '研究' })).toBeVisible();
    await page.getByLabel('态度').selectOption('park');
    await expect(page.getByRole('button', { name: '查看作品 短视频脚本生成器' })).toBeVisible();
    await expect(page.getByRole('button', { name: '查看作品 个人知识库' })).toBeHidden();
  });

  test('opens detail with file facts and handoff', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: '查看作品 个人知识库' }).click();
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('文件核实');
    await expect(dialog).toContainText('交接手册');
    await page.keyboard.press('Escape');
    await expect(dialog).not.toBeVisible();
  });

  test('Enter opens detail from a card (keyboard)', async ({ page }) => {
    await page.goto('/workspace?fixture=1');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: '查看作品 个人知识库' }).focus();
    await page.keyboard.press('Enter');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
  });
});

test.describe('Works album viewport and page errors', () => {
  for (const size of sizes) {
    test(`no horizontal overflow or pageerror at ${size.width}x${size.height}`, async ({ page }) => {
      const pageErrors: Error[] = [];
      page.on('pageerror', error => pageErrors.push(error));
      await page.setViewportSize(size);
      await page.goto('/workspace?fixture=1');
      await page.waitForLoadState('networkidle');
      await expect(page.getByRole('button', { name: '查看作品 个人知识库' })).toBeVisible();
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth + 1);
      expect(overflow).toBe(false);
      expect(pageErrors).toEqual([]);
    });
  }
});
