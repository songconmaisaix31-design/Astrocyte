import { test, expect } from '@playwright/test';
import { realService, importTrackedDocument, importPublicPaperVersions, registerCurrentProject } from './realService';

test('实际输入校验失败保留输入和同一命令身份', async ({ page }) => {
  const writes: { key: string; body: unknown }[] = [];
  page.on('request', request => {
    if (request.method() === 'POST' && request.url().endsWith('/materials/imports')) writes.push({ key: request.headers()['idempotency-key'], body: request.postDataJSON() });
  });
  await page.goto('/attention');
  const add = page.locator('main').getByRole('button', { name: '添加资料', exact: true });
  await add.click();
  const dialog = page.getByRole('dialog', { name: '添加资料' });
  await dialog.getByLabel('arXiv 来源').fill(' ');
  const reason = '核查输入拒绝后保留人工草稿'.repeat(8);
  await dialog.getByLabel('收藏理由（可选）').fill(reason);
  for (let index = 0; index < 2; index++) {
    const response = page.waitForResponse(value => value.url().endsWith('/materials/imports') && value.request().method() === 'POST');
    await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
    expect((await response).status()).toBe(400);
    await expect(dialog.getByRole('alert')).toBeVisible();
    await expect(dialog.getByLabel('arXiv 来源')).toHaveValue(' ');
    await expect(dialog.getByLabel('收藏理由（可选）')).toHaveValue(reason);
    if (index === 0) {
      await page.keyboard.press('Escape');
      await expect(add).toBeFocused();
      await page.keyboard.press('Enter');
    }
  }
  expect(writes).toHaveLength(2);
  expect(writes[1]).toEqual(writes[0]);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});

test('实际公开论文版本切换丢弃取消请求，读取不增加人工使用次数', async ({ page }) => {
  test.setTimeout(300_000);
  const { server, api } = await realService();
  try {
    const { first, second } = await importPublicPaperVersions(api);
    const id = second.detail.material.id;
    const latest = await api.get(`/materials/${id}/revisions/2/content`);
    const previous = await api.get(`/materials/${id}/revisions/1/content`);
    expect(latest.content_digest).not.toBe(previous.content_digest);
    const writes: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) writes.push(request.url()); });
    await page.goto(`${server.webURL}/attention`);
    await page.locator('main [role="button"]').filter({ hasText: second.detail.material.title! }).click();
    const dialog = page.getByRole('dialog', { name: '素材详情' });
    await expect(dialog.locator('pre').filter({ hasText: latest.text })).toHaveText(latest.text);
    await dialog.getByLabel('原文版本', { exact: true }).selectOption('1');
    await dialog.getByLabel('原文版本', { exact: true }).selectOption('2');
    await expect(dialog.locator('pre').filter({ hasText: latest.text })).toHaveText(latest.text);
    await dialog.getByRole('button', { name: '刷新资料详情' }).click();
    await expect(dialog.getByRole('button', { name: '主动重读' })).toBeEnabled();
    expect((await api.get(`/materials/${id}`)).material.human_usage_count).toBe(first.detail.material.human_usage_count);
    expect(writes).toEqual([]);
  } finally { await server.close({ preserveData: true }); }
});

test('实际网络断开保留详情快照并禁止写入，重连后恢复', async ({ page, context }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const imported = await importTrackedDocument(api);
    const content = await api.get(`/materials/${imported.detail.material.id}/revisions/1/content`);
    await page.goto(`${server.webURL}/attention`);
    const trigger = page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' });
    await trigger.focus();
    await page.keyboard.press('Enter');
    const dialog = page.getByRole('dialog');
    await expect(dialog.locator('pre').filter({ hasText: content.text })).toHaveText(content.text);
    await context.setOffline(true);
    await dialog.getByRole('button', { name: '刷新资料详情' }).click();
    await expect(dialog.getByRole('alert')).toContainText('数据可能已过期');
    await expect(dialog.locator('pre').filter({ hasText: content.text })).toHaveText(content.text);
    await expect(dialog.getByRole('button', { name: '继续沉淀', exact: true })).toBeDisabled();
    await context.setOffline(false);
    await dialog.getByRole('button', { name: '重试加载' }).click();
    await expect(dialog.getByRole('button', { name: '继续沉淀', exact: true })).toBeEnabled();
    await page.keyboard.press('Escape');
    await expect(trigger).toBeFocused();
  } finally { await context.setOffline(false); await server.close({ preserveData: true }); }
});

test('实际未配置自动处理和未纳入项目范围均禁用提交', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    await importTrackedDocument(api);
    const project = await registerCurrentProject(api);
    const writes: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) writes.push(request.url()); });
    await page.goto(`${server.webURL}/attention`);
    await page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' }).click();
    const dialog = page.getByRole('dialog', { name: '素材详情' });
    await dialog.getByRole('button', { name: '继续沉淀', exact: true }).click();
    const automatic = dialog.getByRole('region', { name: '自动沉淀' });
    await expect(automatic.getByRole('button', { name: '提交自动沉淀', exact: true })).toBeDisabled();
    await automatic.getByLabel('自动沉淀的项目模型许可', { exact: true }).selectOption(project.id);
    await expect(automatic).toContainText('主动 @ 纳入本次全部固定版本');
    await expect(automatic.getByRole('button', { name: '提交自动沉淀', exact: true })).toBeDisabled();
    expect(writes).toEqual([]);
  } finally { await server.close({ preserveData: true }); }
});

test('实际人工前轮记录和固定来源保留，权限拒绝不创建模型作业', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const imported = await importTrackedDocument(api);
    const material = imported.detail.material;
    const content = await api.get(`/materials/${material.id}/revisions/1/content`);
    const refs = [{ material_id: material.id, revision: 1, locator: material.source_locator, span: null }];
    const prior = await api.write('/distillations', {
      input_refs: refs, stage: 'content', output_text: content.text.split('\n').slice(0, 3).join('\n'),
      question: '摘录本轨执行协议标题', processing_config: 'manual:v1',
    });
    expect(prior).toHaveProperty('distillation');
    await page.goto(`${server.webURL}/attention`);
    await page.locator('main [role="button"]').filter({ hasText: 'AGENTS.md' }).click();
    const dialog = page.getByRole('dialog', { name: '素材详情' });
    await dialog.getByRole('button', { name: '继续沉淀', exact: true }).click();
    const automatic = dialog.getByRole('region', { name: '自动沉淀' });
    await expect(automatic).toContainText('摘录本轨执行协议标题');
    const key = crypto.randomUUID();
    const body = { schema_version: 1, request_id: key, expected_version: 1, input_refs: refs, stage: 'topic', question: '检查固定引用', processing_config: 'manual:v1', prior_distillation_ids: [] };
    const first = await api.request('/distillations/jobs', { method: 'POST', body, key });
    const second = await api.request('/distillations/jobs', { method: 'POST', body, key });
    expect(first.status).toBe(403);
    expect(first.data).toMatchObject({ error: { code: 'scope_denied', required_action: 'choose_authorized_public_source' } });
    expect(second.status).toBe(first.status);
    expect((await api.get(`/materials/${material.id}`)).distillations).toHaveLength(1);
    expect((await api.get('/jobs')).items.filter(job => job.kind === 'distillation')).toHaveLength(0);
  } finally { await server.close({ preserveData: true }); }
});
