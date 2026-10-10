import { expect, it } from 'vitest';
import { canImportSnapshot, importAdapterFor, parsePluginSnapshot, snapshotCaveat } from './pluginSnapshot';

it('parses a minimal snapshot and requires a source_url', () => {
  const review = parsePluginSnapshot(JSON.stringify({ source_url: 'https://arxiv.org/abs/2504.16054', title: '标题' }));
  expect(review.source_url).toBe('https://arxiv.org/abs/2504.16054');
  expect(review.title).toBe('标题');
  expect(review.content_state).toBe('abstract_only');
  expect(() => parsePluginSnapshot('{}')).toThrow(/缺少来源链接/);
  expect(() => parsePluginSnapshot('not json')).toThrow(/不是有效的 JSON/);
});

it('parses the real extension snapshot shape (host_family + content_state)', () => {
  const review = parsePluginSnapshot(JSON.stringify({
    schema_version: 1, source_url: 'https://journals.plos.org/plosone/article?id=x', host_family: 'plos',
    source_key: 'doi:10.1/example', title: 'PLOS 论文', authors: ['甲'], doi: '10.1/example',
    content_state: 'readable_fulltext', warning: '', truncated: false,
  }));
  expect(review.host_family).toBe('plos');
  expect(review.content_state).toBe('readable_fulltext');
  expect(review.source_key).toBe('doi:10.1/example');
});

it('never treats an abstract or body field as full text', () => {
  const review = parsePluginSnapshot(JSON.stringify({ source_url: 'https://example.com/p', abstract: '只是摘要', text: '不该被当作正文' }));
  expect(review.abstract).toBe('只是摘要');
  expect((review as unknown as { text?: string }).text).toBeUndefined();
  expect(snapshotCaveat).toContain('不是正文');
});

it('routes arXiv to the arxiv adapter and every other source to paper_url', () => {
  expect(importAdapterFor(parsePluginSnapshot(JSON.stringify({ source_url: 'https://arxiv.org/abs/2504.16054' })))).toBe('arxiv');
  expect(importAdapterFor(parsePluginSnapshot(JSON.stringify({ source_url: 'https://example.com/p', host_family: 'arxiv', arxiv_id: '2504.16054' })))).toBe('arxiv');
  expect(importAdapterFor(parsePluginSnapshot(JSON.stringify({ source_url: 'https://journals.plos.org/p', host_family: 'plos' })))).toBe('paper_url');
  expect(importAdapterFor(parsePluginSnapshot(JSON.stringify({ source_url: 'https://example.com/p', host_family: 'pmc' })))).toBe('paper_url');
});

it('gates import on non-restricted content state', () => {
  expect(canImportSnapshot(parsePluginSnapshot(JSON.stringify({ source_url: 'https://example.com/p', content_state: 'readable_fulltext' })))).toBe(true);
  expect(canImportSnapshot(parsePluginSnapshot(JSON.stringify({ source_url: 'https://example.com/p', content_state: 'abstract_only' })))).toBe(true);
  expect(canImportSnapshot(parsePluginSnapshot(JSON.stringify({ source_url: 'https://example.com/p', content_state: 'paywall' })))).toBe(false);
  expect(canImportSnapshot(parsePluginSnapshot(JSON.stringify({ source_url: 'https://example.com/p', content_state: 'restricted' })))).toBe(false);
});

it('does not crash on a locator that is not an absolute URL', () => {
  expect(importAdapterFor(parsePluginSnapshot(JSON.stringify({ source_url: 'not a url', host_family: 'generic' })))).toBe('paper_url');
});
