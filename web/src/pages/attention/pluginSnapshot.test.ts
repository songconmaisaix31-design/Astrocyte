import { expect, it } from 'vitest';
import { canImportSnapshot, parsePluginSnapshot, snapshotContentStateLabel, snapshotCaveat, snapshotImportOptions } from './pluginSnapshot';

const snapshot = (fields: Record<string, unknown>): string => JSON.stringify({ source_url: 'https://example.com/p', ...fields });

it('parses a minimal snapshot and requires a source_url', () => {
  const review = parsePluginSnapshot(JSON.stringify({ source_url: 'https://arxiv.org/abs/2504.16054', title: '标题' }));
  expect(review.source_url).toBe('https://arxiv.org/abs/2504.16054');
  expect(review.title).toBe('标题');
  expect(review.content_state).toBe('abstract_only');
  expect(() => parsePluginSnapshot('{}')).toThrow(/缺少来源链接/);
  expect(() => parsePluginSnapshot('not json')).toThrow(/不是有效的 JSON/);
});

it('parses the real extension snapshot shape and preserves the raw text', () => {
  const review = parsePluginSnapshot(snapshot({ host_family: 'plos', source_key: 'doi:10.1/example', title: 'PLOS 论文', authors: ['甲'], doi: '10.1/example', content_state: 'readable_fulltext', text: '这是浏览器提取的完整正文，不能被丢弃。' }));
  expect(review.host_family).toBe('plos');
  expect(review.content_state).toBe('readable_fulltext');
  expect(review.text).toBe('这是浏览器提取的完整正文，不能被丢弃。');
});

it('offers the preserved full body as paper_snapshot carrying the whole snapshot JSON (no re-fetch)', () => {
  const raw = snapshot({ content_state: 'readable_fulltext', text: '完整正文', source_key: 'doi:10.1/x' });
  const options = snapshotImportOptions(parsePluginSnapshot(raw), raw);
  expect(options.map(o => o.adapter)).toEqual(['paper_snapshot']);
  expect(options[0].source_locator).toBe('https://example.com/p');
  expect(options[0].export_text).toBe(raw);
});

it('offers public PDF options via paper_pdf for any host, not only arxiv', () => {
  const raw = snapshot({ content_state: 'readable_fulltext', pdf_urls: ['https://journals.plos.org/p/file.pdf'], text: '' });
  const options = snapshotImportOptions(parsePluginSnapshot(raw), raw);
  expect(options.map(o => o.adapter)).toEqual(['paper_pdf']);
  expect(options[0].source_locator).toBe('https://journals.plos.org/p/file.pdf');
  expect(options[0].export_text).toBeNull();
});

it('offers both HTML full text and public PDF when both are present', () => {
  const raw = snapshot({ content_state: 'readable_fulltext', text: '正文', pdf_urls: ['https://example.com/p.pdf'] });
  const options = snapshotImportOptions(parsePluginSnapshot(raw), raw);
  expect(options.map(o => o.adapter)).toEqual(['paper_snapshot', 'paper_pdf']);
});

it('falls back to the arxiv adapter only for an arxiv snapshot with no body or PDF', () => {
  const arxiv = snapshot({ source_url: 'https://arxiv.org/abs/2504.16054', arxiv_id: '2504.16054', content_state: 'abstract_only' });
  expect(snapshotImportOptions(parsePluginSnapshot(arxiv), arxiv).map(o => o.adapter)).toEqual(['arxiv']);
  const nonArxiv = snapshot({ content_state: 'abstract_only' });
  expect(snapshotImportOptions(parsePluginSnapshot(nonArxiv), nonArxiv)).toEqual([]);
});

it('never treats an abstract, truncated body or paywall as full text', () => {
  const truncated = snapshot({ content_state: 'readable_fulltext', text: '被截断的正文', truncated: true });
  expect(snapshotImportOptions(parsePluginSnapshot(truncated), truncated)).toEqual([]);
  const paywalled = snapshot({ content_state: 'paywall', text: '不该被当作正文', pdf_urls: ['https://example.com/p.pdf'] });
  expect(snapshotImportOptions(parsePluginSnapshot(paywalled), paywalled)).toEqual([]);
  expect(snapshotCaveat).toContain('不是正文');
});

it('gates import on a non-bypassing import path', () => {
  const full = snapshot({ content_state: 'readable_fulltext', text: '正文' });
  expect(canImportSnapshot(parsePluginSnapshot(full), full)).toBe(true);
  const pdf = snapshot({ content_state: 'abstract_only', pdf_urls: ['https://example.com/p.pdf'] });
  expect(canImportSnapshot(parsePluginSnapshot(pdf), pdf)).toBe(true);
  const restricted = snapshot({ content_state: 'restricted' });
  expect(canImportSnapshot(parsePluginSnapshot(restricted), restricted)).toBe(false);
  const abstractOnly = snapshot({ content_state: 'abstract_only' });
  expect(canImportSnapshot(parsePluginSnapshot(abstractOnly), abstractOnly)).toBe(false);
});

it('labels snapshot content states without guessing', () => {
  expect(snapshotContentStateLabel('readable_fulltext')).toBe('原文可得');
  expect(snapshotContentStateLabel('abstract_only')).toBe('仅摘要/元数据');
  expect(snapshotContentStateLabel('paywall')).toBe('付费墙');
  expect(snapshotContentStateLabel('restricted')).toBe('受限');
});

it('does not crash on a locator that is not an absolute URL', () => {
  const raw = snapshot({ source_url: 'not a url', host_family: 'generic', content_state: 'abstract_only' });
  expect(snapshotImportOptions(parsePluginSnapshot(raw), raw)).toEqual([]);
});
