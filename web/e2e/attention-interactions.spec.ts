import { test, expect } from '@playwright/test';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, importMaterial } from '../../tests/s1/api.mjs';
import { paperImport } from '../../tests/s1/fixtures.mjs';
import type { components } from '../src/api/schema';
import { seedCandidate } from '../../tests/s1/scenarios.mjs';

type Schemas = components['schemas'];

// Actual HTTP/SQLite with deliberately synthetic contract_local input. No network/model calls.
// W4 owns AT01–04; this checks the additional editable domain/reference/manual forms.
test('human classification, pinned references, manual layers and candidate edits use real service state', async ({ page }, testInfo) => {
  test.setTimeout(120_000);
  const server = await startS1Server({ browser: true });
  try {
    const api = await humanAPI(server.apiURL);
    const imported = await importMaterial(api, { ...paperImport('https://example.invalid/contract-local/w3-editing'), title: 'contract_local 长中文交互材料' });
    const id = imported.detail.material.id;
    await page.goto(`${server.webURL}/attention`);
    const main = page.locator('main');
    await main.locator('summary').filter({ hasText: /^新建资料域$/ }).click();
    const domainForm = main.locator('form').filter({ has: page.getByRole('button', { name: '创建资料域', exact: true }) });
    await domainForm.getByLabel('域名称', { exact: true }).fill('当前研究 · contract_local');
    await domainForm.getByLabel('域说明', { exact: true }).fill('人工分类不赋予 Agent 权限'.repeat(12));
    await domainForm.getByRole('button', { name: '创建资料域', exact: true }).click();
    await expect.poll(async () => (await api.get('/material-domains') as Schemas['MaterialDomainListV1']).items.length).toBe(1);
    await main.locator('[role="button"]').filter({ hasText: imported.detail.material.title! }).click();
    let dialog = page.getByRole('dialog', { name: '素材详情' });
    await expect(dialog.getByRole('checkbox', { name: '当前研究 · contract_local', exact: true })).toBeEnabled();
    await dialog.getByRole('checkbox', { name: '当前研究 · contract_local', exact: true }).check();
    await dialog.getByRole('button', { name: '保存人工分类', exact: true }).click();
    await expect.poll(async () => (await api.get(`/materials/${id}`)).material.domain_ids?.length).toBe(1);
    const classifiedBaseline = await api.get(`/materials/${id}`);

    await dialog.getByRole('button', { name: '继续沉淀', exact: true }).click();
    for (const stage of ['content', 'topic', 'project']) {
      await dialog.getByLabel('沉淀层次', { exact: true }).selectOption(stage);
      await dialog.getByLabel('本轮问题', { exact: true }).fill(`contract_local ${stage} 问题`);
      await dialog.getByLabel('人工整理结果', { exact: true }).fill(`contract_local ${stage} 人工结果；没有自动模型调用`);
      await dialog.getByLabel('下一轮问题（可选）', { exact: true }).fill('继续核对来源版本和缺失依据');
      if (stage !== 'content') await dialog.getByLabel('待查问题（每行一项）', { exact: true }).fill('真实科学收益尚未测量');
      if (stage === 'project') {
        await dialog.getByLabel('关联目标（每行一项）', { exact: true }).fill('contract_local-goal');
        await dialog.getByLabel('现有资产（每行一项）', { exact: true }).fill('已有来源版本');
        await dialog.getByLabel('预期改进', { exact: true }).fill('保留明确人工来源');
        await dialog.getByLabel('最小成果', { exact: true }).fill('人工对照记录');
      }
      await dialog.getByRole('button', { name: '保存人工沉淀', exact: true }).click();
      await expect.poll(async () => (await api.get(`/materials/${id}`)).distillations.length).toBe(['content', 'topic', 'project'].indexOf(stage) + 1);
      await expect(dialog.getByRole('button', { name: '保存人工沉淀', exact: true })).toBeEnabled();
    }
    const detail = await api.get(`/materials/${id}`);
    expect(detail.distillations.map((record: { stage: string }) => record.stage).sort()).toEqual(['content', 'project', 'topic']);
    expect(detail.material.human_usage_count).toBe(classifiedBaseline.material.human_usage_count);
    await dialog.getByRole('button', { name: '记录本次项目复用', exact: true }).click();
    await expect.poll(async () => (await api.get(`/materials/${id}`)).material.human_usage_count).toBe(classifiedBaseline.material.human_usage_count! + 1);
    expect((await api.get(`/materials/${id}`)).uses.filter((use: { action: string }) => use.action === 'project_reuse')).toHaveLength(1);
    await dialog.getByRole('button', { name: '形成候选', exact: true }).click();
    const candidateForm = dialog.locator('form').filter({ has: page.getByRole('button', { name: '保存候选', exact: true }) });
    await candidateForm.getByLabel('候选标题', { exact: true }).fill('contract_local 候选编辑');
    await candidateForm.getByLabel('候选用途 / 为什么值得做', { exact: true }).fill('检验真实保存与版本保留');
    await candidateForm.getByLabel('最小下一步', { exact: true }).fill('人工对照来源版本');
    await candidateForm.getByRole('checkbox').last().check();
    await candidateForm.getByRole('button', { name: '保存候选', exact: true }).click();
    dialog = page.getByRole('dialog', { name: '机会详情' });
    await expect(dialog).toBeVisible();
    await dialog.getByRole('button', { name: '编辑候选', exact: true }).click();
    await dialog.getByLabel('候选标题', { exact: true }).fill('contract_local 候选新版本');
    await dialog.getByRole('button', { name: '保存候选新版本', exact: true }).click();
    await expect.poll(async () => (await api.get('/opportunities')).items[0].revision).toBe(2);
    const candidateID = (await api.get('/opportunities')).items[0].id;
    expect((await api.get(`/opportunities/${candidateID}`)).revisions).toHaveLength(2);
    await dialog.getByRole('button', { name: '关闭', exact: true }).click();

    await main.locator('summary').filter({ hasText: /^创建项目顶层空间$/ }).click();
    await main.getByLabel('空间名称', { exact: true }).fill('contract_local 项目空间');
    await main.getByRole('button', { name: '创建空间', exact: true }).click();
    await expect(main.getByLabel('@ 资料', { exact: true })).toBeVisible();
    await main.getByLabel('@ 资料', { exact: true }).selectOption(id);
    await main.getByLabel('@ 固定内容版本', { exact: true }).selectOption('1');
    await main.getByRole('button', { name: '@ 引用文件', exact: true }).click();
    const spaceID = (await api.get('/project-spaces') as Schemas['ProjectSpaceListV1']).items[0].id;
    await expect.poll(async () => (await api.get(`/project-spaces/${spaceID}`) as Schemas['ProjectSpaceResultV1']).space.material_refs.length).toBe(1);
    const classified = await api.get(`/materials/${id}`);
    expect(classified.material.domain_ids).toEqual(detail.material.domain_ids);
    await main.getByRole('button', { name: /^移除引用/ }).click();
    await expect.poll(async () => (await api.get(`/project-spaces/${spaceID}`) as Schemas['ProjectSpaceResultV1']).space.material_refs.length).toBe(0);
    expect((await api.get(`/materials/${id}`)).material.lifecycle).toBe('active');
    expect((await api.get('/missions')).items).toHaveLength(0);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.screenshot({ path: testInfo.outputPath('attention-classification-space.png'), fullPage: true });
  } finally { await server.close(); }
});

test('ranking waits for explicit complete human weights, keeps unknown scores and saves immutable profile versions', async ({ page }, testInfo) => {
  test.setTimeout(120_000);
  const server = await startS1Server({ browser: true });
  try {
    let api = await humanAPI(server.apiURL);
    await seedCandidate(api, 'https://example.invalid/contract-local/w3-ranking');
    const getProfile = async () => await api.get('/attention-ranking-profile') as Schemas['RankingProfileDetailV1'];
    expect((await getProfile()).configured).toBe(false);
    await page.goto(`${server.webURL}/attention`);
    const section = page.locator('section').filter({ has: page.getByRole('heading', { name: '候选四维排序', exact: true }) });
    for (const label of ['目标进展', '当前兴趣', '项目改善', '创新性']) await expect(section.getByLabel(`${label}权重`, { exact: true })).toHaveValue('');
    await expect(section.getByRole('checkbox', { name: '启用综合排序', exact: true })).not.toBeChecked();
    await section.getByLabel('目标进展权重', { exact: true }).fill('0.5');
    await section.getByLabel('当前兴趣权重', { exact: true }).fill('0.2');
    await section.getByLabel('项目改善权重', { exact: true }).fill('0.2');
    await section.getByLabel('创新性权重', { exact: true }).fill('0.1');
    await section.getByRole('checkbox', { name: '启用综合排序', exact: true }).check();
    await section.getByRole('button', { name: '保存排序配置', exact: true }).click();
    await expect.poll(async () => (await getProfile()).profile?.version).toBe(1);
    await expect(section.getByRole('button', { name: '保存排序配置', exact: true })).toBeEnabled();
    const ranked = (await api.get('/opportunities')).items[0];
    expect(ranked.composite_score == null).toBe(true);
    expect(Object.values(ranked.dimensions).every(dimension => dimension.value === null)).toBe(true);
    await section.getByLabel('目标进展权重', { exact: true }).fill('0.6');
    await section.getByRole('button', { name: '保存排序配置', exact: true }).click();
    await expect.poll(async () => (await getProfile()).profile?.version).toBe(2);
    expect((await getProfile()).versions).toHaveLength(2);
    expect((await getProfile()).versions.find(version => version.version === 1)?.weights.goal_progress).toBe(0.5);
    await server.restart();
    api = await humanAPI(server.apiURL);
    await page.reload();
    await expect(section.getByLabel('目标进展权重', { exact: true })).toHaveValue('0.6');
    await expect(section.getByRole('checkbox', { name: '启用综合排序', exact: true })).toBeChecked();
    expect((await getProfile()).profile?.version).toBe(2);
    expect((await api.get('/missions')).items).toHaveLength(0);
    await section.scrollIntoViewIfNeeded();
    await page.screenshot({ path: testInfo.outputPath('attention-ranking-profile.png') });
  } finally { await server.close(); }
});
