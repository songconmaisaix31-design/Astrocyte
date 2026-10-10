import { expect, it } from 'vitest';
import {
  canSelectPaper, clearPaperSelection, eligiblePapers, isAllEligiblePicked,
  paperContentStateLabel, paperImportAdapter, paperSearchKey, pickedPapers, togglePaper,
  type PaperMetadata,
} from './paperSearch';

const hit = (overrides: Partial<PaperMetadata> = {}): PaperMetadata => ({
  site: 'arxiv', source_key: 'arxiv:2504.16054', arxiv_id: '2504.16054', title: '示例论文标题',
  authors: ['甲', '乙'], abstract: '摘要文字', published_at: '2024-01-01',
  locator: 'https://arxiv.org/abs/2504.16054', content_state: 'abstract_only',
  ...overrides,
});

it('maps content states to readable labels', () => {
  expect(paperContentStateLabel('readable_fulltext')).toBe('原文可得');
  expect(paperContentStateLabel('abstract_only')).toBe('仅摘要/元数据');
  expect(paperContentStateLabel('paywall')).toBe('付费墙');
  expect(paperContentStateLabel('restricted')).toBe('受限');
});

it('keys a hit by its canonical source key, falling back to locator', () => {
  expect(paperSearchKey(hit())).toBe('arxiv:2504.16054');
  expect(paperSearchKey(hit({ source_key: '', locator: 'https://example.com/p' }))).toBe('https://example.com/p');
});

it('routes import to arxiv only for arXiv identity, paper_url otherwise', () => {
  expect(paperImportAdapter(hit())).toBe('arxiv');
  expect(paperImportAdapter(hit({ arxiv_id: undefined, site: 'crossref', doi: '10.1/x' }))).toBe('paper_url');
  expect(paperImportAdapter(hit({ arxiv_id: undefined, site: 'unknown', locator: 'https://journals.plos.org/x' }))).toBe('paper_url');
});

it('treats metadata hits as selectable; restricted states stay non-selectable', () => {
  expect(canSelectPaper(hit())).toBe(true);
  expect(canSelectPaper(hit({ content_state: 'readable_fulltext' }))).toBe(true);
  expect(canSelectPaper(hit({ content_state: 'paywall' }))).toBe(false);
  expect(canSelectPaper(hit({ content_state: 'restricted' }))).toBe(false);
});

it('toggles selection explicitly without ever auto-selecting', () => {
  const a = hit({ source_key: 'a' });
  const b = hit({ source_key: 'b' });
  expect(togglePaper([], a)).toEqual(['a']);
  expect(togglePaper(['a'], b)).toEqual(['a', 'b']);
  expect(togglePaper(['a', 'b'], a)).toEqual(['b']);
  expect(clearPaperSelection()).toEqual([]);
});

it('returns only eligible picked hits and detects full eligibility', () => {
  const a = hit({ source_key: 'a' });
  const b = hit({ source_key: 'b' });
  expect(pickedPapers([a, b], ['a'])).toEqual([a]);
  expect(eligiblePapers([a, b])).toEqual([a, b]);
  expect(isAllEligiblePicked([a, b], ['a'])).toBe(false);
  expect(isAllEligiblePicked([a, b], ['a', 'b'])).toBe(true);
  expect(isAllEligiblePicked([], [])).toBe(false);
});
