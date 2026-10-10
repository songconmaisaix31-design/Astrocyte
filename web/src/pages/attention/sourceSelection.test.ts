import { expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { canRecommendSourceItem, canSelectSourceItem, sourceItemKey } from './sourceSelection';

const item: components['schemas']['SourceItemV1'] = {
  source_id: 's1', external_id: 'BV1public', revision: 2, stale: false, selected: false, import_job_id: null, material_id: null,
  metadata: { external_id: 'BV1public', locator: 'https://www.bilibili.com/video/BV1public', title: '实际标题', description: '实际简介', author: '公开作者', cover: '', published_at: 0, provider_status: null, unavailable_reason: '' },
  recommendation: { status: 'succeeded', text: '可供人工继续核对', reason: '', metadata_revision: 2, configuration_id: 'approved-cli', error: null, provenance: { processor: 'cli', version: '1', source: 'metadata', mode: 'model' } },
};

it('requires completed feedback on the exact metadata revision before human import selection', () => {
  expect(canSelectSourceItem(item)).toBe(true);
  for (const status of ['pending', 'running', 'failed', 'unknown', 'unavailable'] as const) expect(canSelectSourceItem({ ...item, recommendation: { ...item.recommendation!, status } })).toBe(false);
  expect(canSelectSourceItem({ ...item, recommendation: { ...item.recommendation!, metadata_revision: 1 } })).toBe(false);
  expect(canSelectSourceItem({ ...item, stale: true })).toBe(false);
  expect(canSelectSourceItem({ ...item, selected: true })).toBe(false);
  expect(sourceItemKey(item)).not.toBe(sourceItemKey({ ...item, revision: 3 }));
});

it('does not resend running or unknown recommendation requests, including an unknown error code', () => {
  expect(canRecommendSourceItem({ ...item, recommendation: null })).toBe(true);
  expect(canRecommendSourceItem(item)).toBe(false);
  for (const status of ['pending', 'running', 'unknown'] as const) expect(canRecommendSourceItem({ ...item, recommendation: { ...item.recommendation!, status } })).toBe(false);
  expect(canRecommendSourceItem({ ...item, recommendation: { ...item.recommendation!, status: 'failed', error: { code: 'delivery_unknown', message: 'unknown', request_id: 'r', retryable: true, required_action: 'reconcile' } } })).toBe(false);
});
