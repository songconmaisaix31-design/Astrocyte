/**
 * Paper search presentation model (pure, no I/O).
 *
 * The local shapes below mirror W0's published `PaperSearchResultV1` /
 * `PaperSearchHitV1` contract (contracts/openapi.yaml, owned by W0, converged at
 * `e39f8f1`). Search is ephemeral public metadata only: `GET /api/v1/papers/search`
 * returns hits keyed by a canonical `source_key`, a `provider` index, a raw
 * `content_state` and candidate `pdf_urls`. Full-text availability is only
 * resolved when the human selects a hit and imports it through the existing
 * ImportMaterial path (adapter per site, source_key dedup, new revision). No
 * fixture or guessed value is ever presented as a real result.
 */

/** Raw availability vocabulary; metadata indexes always report `abstract_only`. */
export type PaperContentState = 'readable_fulltext' | 'abstract_only' | 'paywall' | 'restricted' | 'unknown';

/** Local view shape mirroring W0 `PaperSearchHitV1`. */
export interface PaperSearchHit {
  source_key: string;
  provider: string;
  title: string;
  authors: string[];
  year: number;
  venue?: string;
  arxiv_id?: string;
  doi?: string;
  locator: string;
  abstract?: string;
  content_state: PaperContentState;
  pdf_urls: string[];
}

/** Local view shape mirroring W0 `PaperSearchResultV1`. */
export interface PaperSearchResult {
  schema_version: 1;
  query: string;
  items: PaperSearchHit[];
  next_cursor: string | null;
  has_more: boolean;
  warnings: string[];
}

export const paperContentStateLabel = (state: PaperContentState): string => {
  switch (state) {
    case 'readable_fulltext': return '原文可得';
    case 'abstract_only': return '仅摘要/元数据';
    case 'paywall': return '付费墙';
    case 'restricted': return '受限';
    case 'unknown': return '可得性未知';
  }
};

/** Canonical key of a hit is its contract `source_key`; fall back to locator. */
export const paperSearchKey = (hit: PaperSearchHit): string => hit.source_key || hit.locator;

function isArxiv(hit: PaperSearchHit): boolean {
  return !!hit.arxiv_id || hit.provider === 'arxiv' || /(^|\.)arxiv\.org$/i.test(hit.locator || '');
}

/** Default import adapter for a search hit; non-arXiv sites use `paper_url` HTML extraction. */
export function paperImportAdapter(hit: PaperSearchHit): 'arxiv' | 'paper_url' {
  return isArxiv(hit) ? 'arxiv' : 'paper_url';
}

/**
 * A search hit is selectable unless its content state is paywall/restricted.
 * `abstract_only` and `unknown` stay selectable: the import job resolves real
 * availability and the source_key dedup keeps a retry from creating a
 * duplicate. Nothing is auto-selected.
 */
export function canSelectPaper(hit: PaperSearchHit): boolean {
  return hit.content_state !== 'paywall' && hit.content_state !== 'restricted';
}

/** Explicit, human-only selection. Never selects a hit automatically. */
export function togglePaper(selected: string[], hit: PaperSearchHit): string[] {
  return selected.includes(paperSearchKey(hit)) ? selected.filter(id => id !== paperSearchKey(hit)) : [...selected, paperSearchKey(hit)];
}

export function clearPaperSelection(): string[] {
  return [];
}

/** Hits the human has actually picked. */
export function pickedPapers(hits: PaperSearchHit[], selected: string[]): PaperSearchHit[] {
  return hits.filter(hit => selected.includes(paperSearchKey(hit)));
}

/** All selectable hits. */
export function eligiblePapers(hits: PaperSearchHit[]): PaperSearchHit[] {
  return hits.filter(canSelectPaper);
}

/** True when the human picked every currently selectable hit. */
export function isAllEligiblePicked(hits: PaperSearchHit[], selected: string[]): boolean {
  const eligible = eligiblePapers(hits);
  return eligible.length > 0 && pickedPapers(hits, selected).length === eligible.length;
}
