/** Failure injection for UI recovery only. S1 real-service acceptance belongs to W4. */
import { test, expect } from '@playwright/test';
import type { components } from '../src/api/schema';
import { fixtureMaterials } from '../src/fixtures';

const material = { ...fixtureMaterials[0], version: 1, pinned: false };
const provenance: components['schemas']['ProvenanceV1'] = { processor: 'test-injection', version: '1', mode: 'manual', source: 'Injected transport response' };
const detail: components['schemas']['MaterialDetailV1'] = { schema_version: 1, material: { ...material, current_revision: 2 }, revisions: [1, 2].map(revision => ({ material_id: material.id, revision, source_key: 'injected-source', source_locator: material.source_locator, content_digest: `injected-${revision}`, object_ref: `injected-${revision}`, source_spans: [], provenance, created_at: '2026-10-09T00:00:00Z' })), distillations: [], uses: [] };
const error = { schema_version: 1, error: { code: 'provider_unavailable', message: '注入网络故障', retryable: true, request_id: 'injected-request', required_action: '恢复服务后重试' } };

test.beforeEach(async ({ page }) => {
  for (const endpoint of ['material-domains', 'project-spaces']) await page.route(`**/api/v1/${endpoint}`, route => route.fulfill({ json: { schema_version: 1, items: [] } }));
});

test('import failure preserves input and retries the exact command identity', async ({ page }) => {
  await page.route('**/api/v1/auth/session', route => route.fulfill({ json: { schema_version: 1, actor_id: 'test', actor_kind: 'human', csrf_token: 'test-csrf' } }));
  await page.route('**/api/v1/jobs', route => route.fulfill({ json: { schema_version: 1, items: [], next_cursor: null } }));
  const writes: { key: string; body: unknown }[] = [];
  await page.route('**/api/v1/materials/imports', route => {
    writes.push({ key: route.request().headers()['idempotency-key'], body: route.request().postDataJSON() });
    return route.fulfill({ status: 503, json: error });
  });
  await page.goto('/attention');
  await page.locator('main').getByRole('button', { name: '添加资料', exact: true }).click();
  const dialog = page.getByRole('dialog', { name: '添加资料' });
  await dialog.getByLabel('arXiv 来源').fill('https://arxiv.org/abs/2504.16054');
  await dialog.getByLabel('收藏理由（可选）').fill('失败仍保留的长中文收藏理由'.repeat(8));
  await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
  await expect(dialog.getByRole('alert')).toContainText('恢复服务后重试');
  await expect(dialog.getByLabel('arXiv 来源')).toHaveValue('https://arxiv.org/abs/2504.16054');
  await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
  await expect.poll(() => writes.length).toBe(2);
  expect(writes[1]).toEqual(writes[0]);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: `test-results/attention-import-error-${page.viewportSize()?.width}.png`, fullPage: true });
});

test('source version switching discards delayed content and reads never record human activity', async ({ page }) => {
  const writes: string[] = [];
  page.on('request', request => { if (request.method() !== 'GET') writes.push(request.url()); });
  await page.route('**/api/v1/materials', route => route.fulfill({ json: { schema_version: 1, items: [detail.material], next_cursor: null } }));
  await page.route(`**/api/v1/materials/${material.id}`, route => route.fulfill({ json: detail }));
  await page.route(`**/api/v1/materials/${material.id}/revisions/*/content`, async route => {
    const revision = Number(route.request().url().match(/revisions\/(\d+)\//)?.[1]);
    if (revision === 1) await new Promise(resolve => setTimeout(resolve, 500));
    await route.fulfill({ json: { schema_version: 1, material_id: material.id, revision, content_digest: 'injected', text: `注入原文版本 ${revision}`, provenance } }).catch(() => {});
  });
  await page.goto('/attention');
  await page.locator('main [role="button"]').filter({ hasText: material.title! }).click();
  const dialog = page.getByRole('dialog', { name: '素材详情' });
  await expect(dialog).toContainText('注入原文版本 2');
  await dialog.getByLabel('原文版本', { exact: true }).selectOption('1');
  await dialog.getByLabel('原文版本', { exact: true }).selectOption('2');
  await expect(dialog).toContainText('注入原文版本 2');
  await expect(dialog).not.toContainText('注入原文版本 1');
  await dialog.getByRole('button', { name: '刷新资料详情' }).click();
  await expect(dialog.getByRole('button', { name: '主动重读' })).toBeEnabled();
  expect(writes).toEqual([]);
});

test('detail refresh failure keeps the snapshot and disables writes until recovery', async ({ page }) => {
  let fail = false;
  await page.route('**/api/v1/materials', route => route.fulfill({ json: { schema_version: 1, items: [detail.material], next_cursor: null } }));
  await page.route(`**/api/v1/materials/${material.id}`, route => route.fulfill(fail ? { status: 503, json: error } : { json: detail }));
  await page.route(`**/api/v1/materials/${material.id}/revisions/*/content`, route => route.fulfill({ json: { schema_version: 1, material_id: material.id, revision: 2, content_digest: 'injected', text: '注入的旧快照原文', provenance } }));
  await page.goto('/attention');
  const trigger = page.locator('main [role="button"]').filter({ hasText: material.title! });
  await trigger.focus();
  await page.keyboard.press('Enter');
  const dialog = page.getByRole('dialog');
  await expect(dialog).toContainText('注入的旧快照原文');
  fail = true;
  await dialog.getByRole('button', { name: '刷新资料详情' }).click();
  await expect(dialog.getByRole('alert')).toContainText('数据可能已过期');
  await expect(dialog).toContainText('注入的旧快照原文');
  await expect(dialog.getByRole('button', { name: '继续沉淀', exact: true })).toBeDisabled();
  fail = false;
  await dialog.getByRole('button', { name: '重试加载' }).click();
  await expect(dialog.getByRole('button', { name: '继续沉淀', exact: true })).toBeEnabled();
  await page.keyboard.press('Escape');
  await expect(trigger).toBeFocused();
});

test('automatic processing stays disabled when unavailable or the selected source is outside approved scope', async ({ page }) => {
  let available = false;
  const writes: string[] = [];
  page.on('request', request => { if (request.method() !== 'GET') writes.push(request.url()); });
  await page.route('**/api/v1/materials', route => route.fulfill({ json: { schema_version: 1, items: [detail.material], next_cursor: null } }));
  await page.route(`**/api/v1/materials/${material.id}`, route => route.fulfill({ json: detail }));
  await page.route('**/api/v1/distillations/processor', route => route.fulfill({ json: { schema_version: 1, available, processor: 'injected-status', configuration_id: available ? 'injected-config' : null, model: null, reason: '注入隔离检查未通过', required_action: '先修复实际处理环境', allowed_source_keys: [] } }));
  await page.goto('/attention');
  await page.locator('main [role="button"]').filter({ hasText: material.title! }).click();
  const dialog = page.getByRole('dialog', { name: '素材详情' });
  await dialog.getByRole('button', { name: '继续沉淀', exact: true }).click();
  const automatic = dialog.getByRole('region', { name: '自动沉淀' });
  await expect(automatic).toContainText('注入隔离检查未通过');
  await expect(automatic.getByRole('button', { name: '提交自动沉淀', exact: true })).toBeDisabled();
  available = true;
  await automatic.getByRole('button', { name: '检查自动处理服务' }).click();
  await expect(automatic).toContainText('此来源未纳入已批准的自动处理范围');
  await expect(automatic.getByRole('button', { name: '提交自动沉淀', exact: true })).toBeDisabled();
  expect(writes).toEqual([]);
});
