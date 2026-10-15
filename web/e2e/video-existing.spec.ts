/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { createHash } from 'node:crypto';
import type { components } from '../src/api/schema';

type S = components['schemas'];
test('既有实际视频两版原文和952时间位置只读显示，私有原生附件没有HTTP入口', async ({ page }, info) => {
  const origin = process.env.ASTROCYTE_TEST_VIDEO_PUBLIC_URL;
  const id = process.env.ASTROCYTE_TEST_VIDEO_MATERIAL_ID;
  expect(origin, '需要B交付既有实际视频服务的只读页面URL').toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  expect(id).toMatch(/^[A-Z2-7]{26}$/);
  const commands: string[] = [];
  page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') commands.push(request.method()); });
  expect((await page.request.get(`${origin}/api/v1/auth/session`)).status()).toBe(200);
  const response = await page.request.get(`${origin}/api/v1/materials/${id}`);
  expect(response.status()).toBe(200);
  const detail = await response.json() as S['MaterialDetailV1'];
  expect(detail.material.kind).toBe('video');
  expect(detail.material.source_locator).toBe('https://www.bilibili.com/video/BV1PReT6EEqR/');
  expect(detail.material.current_revision).toBe(2);
  expect(detail.revisions.map(item => item.revision).sort()).toEqual([1, 2]);
  expect(detail.material.source_spans).toHaveLength(952);
  const privateRefs = new Set(detail.revisions.flatMap(item => item.attachments ?? []).filter(item => item.media_type === 'application/vnd.astrocyte.private-native+json' || item.name === 'whisper-original.json').map(item => item.object_ref));
  expect(privateRefs.size).toBeGreaterThan(0);
  const actualCounts = detail.revisions.map(item => ({ revision: item.revision, positions: item.source_spans.length, private_count: (item.attachments ?? []).filter(attachment => privateRefs.has(attachment.object_ref)).length, public_count: (item.attachments ?? []).filter(attachment => !privateRefs.has(attachment.object_ref)).length }));
  expect(actualCounts.find(item => item.revision === 2)).toEqual({ revision: 2, positions: 952, private_count: 1, public_count: 3 });
  await info.attach('actual-revision-counts', { body: JSON.stringify(actualCounts), contentType: 'application/json' });
  await page.goto(`${origin}/attention`);
  await page.locator('main li[role="button"]').filter({ has: page.getByText(detail.material.title!, { exact: true }) }).click();
  const dialog = page.getByRole('dialog');
  const source = dialog.getByRole('region', { name: '来源与原文版本', exact: true });
  const versions: { revision: number; positions: number; public_attachments: number; private_attachments: number; text_length: number }[] = [];
  for (const revision of detail.revisions) {
    const contentResponse = await page.request.get(`${origin}/api/v1/materials/${id}/revisions/${revision.revision}/content`);
    expect(contentResponse.status()).toBe(200);
    const content = await contentResponse.json() as S['ContentV1'];
    expect(content.text.length).toBeGreaterThan(1000);
    expect(content.content_digest).toBe(revision.content_digest);
    await dialog.getByLabel('原文版本', { exact: true }).selectOption(String(revision.revision));
    await expect(source.locator('pre')).toHaveText(content.text);
    await expect(source.getByText(revision.source_spans.length ? revision.source_spans.join('；') : '未提供真实位置', { exact: true })).toBeVisible();
    let publicCount = 0, privateCount = 0;
    for (const attachment of revision.attachments ?? []) {
      if (privateRefs.has(attachment.object_ref)) {
        privateCount++;
        await expect(source.getByText(`${attachment.name} · 仅保存在本机`, { exact: true })).toBeVisible();
        await expect(source.getByRole('link').filter({ hasText: attachment.name })).toHaveCount(0);
      } else {
        publicCount++;
        const link = source.getByRole('link').filter({ hasText: attachment.name });
        const path = `/api/v1/materials/${id}/revisions/${revision.revision}/attachments/${encodeURIComponent(attachment.name)}`;
        await expect(link).toHaveAttribute('href', path);
        const original = await page.request.get(`${origin}${path}`);
        expect(original.status()).toBe(200);
        expect(createHash('sha256').update(await original.body()).digest('hex')).toBe(attachment.object_ref);
      }
    }
    const expectedCounts = actualCounts.find(item => item.revision === revision.revision)!;
    expect(privateCount).toBe(expectedCounts.private_count);
    expect(publicCount).toBe(expectedCounts.public_count);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    versions.push({ revision: revision.revision, positions: revision.source_spans.length, public_attachments: publicCount, private_attachments: privateCount, text_length: content.text.length });
  }
  const finalResponse = await page.request.get(`${origin}/api/v1/materials/${id}`);
  expect(finalResponse.status()).toBe(200);
  expect(await finalResponse.json()).toEqual(detail);
  expect(commands).toEqual([]);
  const directory = resolve('../.verification/video-existing');
  await mkdir(directory, { recursive: true });
  await writeFile(resolve(directory, `${info.project.name}-derived.json`), JSON.stringify({ material_id: id, versions, read_only: true, new_asr_calls: 0, new_model_turns: 0, private_original_requested: false }, null, 2));
});
