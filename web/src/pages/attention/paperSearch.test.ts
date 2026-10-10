import { expect, it } from 'vitest';
import {
  canSelectPaper, clearPaperSelection, eligiblePapers, firstUnknownHit,
  isAllEligiblePicked, paperAvailabilityLabel, paperSearchKey, pickedPapers, togglePaper,
  type PaperSearchHit,
} from './paperSearch';

const hit = (overrides: Partial<PaperSearchHit> = {}): PaperSearchHit => ({
  id: 'arxiv:2504.16054v1', title: '示例论文标题', authors: ['甲', '乙'], year: 2024,
  venue: 'arXiv', source_type: 'arxiv', arxiv_id: '2504.16054', doi: null,
  locator: 'https://arxiv.org/abs/2504.16054', abstract: '摘要文字',
  availability: { status: 'full_text', detail: null }, already_imported: false, import_material_id: null,
  ...overrides,
});

it('maps availability states to readable labels', () => {
  expect(paperAvailabilityLabel('full_text')).toBe('原文可得');
  expect(paperAvailabilityLabel('metadata_only')).toBe('仅元数据');
  expect(paperAvailabilityLabel('restricted')).toBe('受限');
  expect(paperAvailabilityLabel('unknown')).toBe('可得性未知');
});

it('keys a hit by its stable source identity', () => {
  expect(paperSearchKey(hit())).toBe('arxiv:2504.16054v1');
  expect(paperSearchKey(hit({ id: 'doi:10.1/2' }))).not.toBe(paperSearchKey(hit()));
});

it('only allows selecting hits that are not already imported and not unknown', () => {
  expect(canSelectPaper(hit())).toBe(true);
  expect(canSelectPaper(hit({ already_imported: true }))).toBe(false);
  expect(canSelectPaper(hit({ availability: { status: 'unknown', detail: null } }))).toBe(false);
  // restricted and metadata-only remain selectable so the human can read the note
  expect(canSelectPaper(hit({ availability: { status: 'restricted', detail: '付费墙' } }))).toBe(true);
  expect(canSelectPaper(hit({ availability: { status: 'metadata_only', detail: null } }))).toBe(true);
});

it('toggles selection explicitly without ever auto-selecting', () => {
  const a = hit({ id: 'a' });
  const b = hit({ id: 'b' });
  expect(togglePaper([], a)).toEqual(['a']);
  expect(togglePaper(['a'], b)).toEqual(['a', 'b']);
  expect(togglePaper(['a', 'b'], a)).toEqual(['b']);
  expect(clearPaperSelection()).toEqual([]);
});

it('returns only eligible picked hits', () => {
  const a = hit({ id: 'a' });
  const b = hit({ id: 'b', already_imported: true });
  const c = hit({ id: 'c', availability: { status: 'unknown', detail: null } });
  const hits = [a, b, c];
  expect(pickedPapers(hits, ['a', 'b', 'c'])).toEqual([a]);
  expect(eligiblePapers(hits)).toEqual([a]);
});

it('detects whether every eligible hit is picked', () => {
  const a = hit({ id: 'a' });
  const b = hit({ id: 'b' });
  expect(isAllEligiblePicked([a, b], ['a'])).toBe(false);
  expect(isAllEligiblePicked([a, b], ['a', 'b'])).toBe(true);
  expect(isAllEligiblePicked([], [])).toBe(false);
  expect(isAllEligiblePicked([a], [])).toBe(false);
});

it('surfaces the first unknown-availability hit', () => {
  const a = hit({ id: 'a' });
  const u = hit({ id: 'u', availability: { status: 'unknown', detail: null } });
  expect(firstUnknownHit([a, u])).toEqual(u);
  expect(firstUnknownHit([a])).toBeUndefined();
});
