/**
 * Paper search presentation model (pure, no I/O).
 *
 * The local shape below mirrors W0's published `PaperMetadataV1`/`PaperSearchResultV1`
 * contract (see contracts/openapi.yaml, owned by W0). Search is ephemeral public
 * metadata only: providers are metadata indexes, so every hit arrives with
 * `content_state === "abstract_only"` and full-text availability is only resolved
 * when the human selects a hit and imports it through the existing ImportMaterial
 * path (adapter per site, source_key dedup, new revision). No fixture or guessed
 * value is ever presented as a real search result.
 */

export type PaperContentState = 'readable_fulltext' | 'abstract_only' | 'paywall' | 'restricted';

/** Local view shape mirroring W0 `PaperMetadataV1`. */
export interface PaperMetadata {
  site: string;
  source_key: string;
  doi?: string;
  arxiv_id?: string;
  title: string;
  authors?: string[];
  abstract?: string;
  published_at?: string;
  locator: string;
  content_state: PaperContentState;
  license?: string;
  warning?: string;
}

/** Local view shape mirroring W0 `PaperSearchResultV1`. */
export interface PaperSearchResult {
  schema_version: 1;
  items: PaperMetadata[];
  warnings?: string[];
}

export const paperContentStateLabel = (state: PaperContentState): string => {
  switch (state) {
    case 'readable_fulltext': return '原文可得';
    case 'abstract_only': return '仅摘要/元数据';
    case 'paywall': return '付费墙';
    case 'restricted': return '受限';
  }
};

export const paperSearchKey = (hit: PaperMetadata): string => hit.source_key || hit.locator;

/** Import adapter for a hit; non-arXiv sites use W1's `paper_url` body extraction. */
export type PaperImportAdapter = 'arxiv' | 'paper_url';
export function paperImportAdapter(hit: PaperMetadata): PaperImportAdapter {
  return hit.arxiv_id || hit.site === 'arxiv' ? 'arxiv' : 'paper_url';
}

/**
 * A search hit is selectable unless its content state is paywall/restricted.
 * W1's search providers are metadata indexes and always return abstract_only, so
 * this is effectively always true today; the check stays defensive for any
 * future provider that reports a restricted state. The human explicitly picks
 * each hit; nothing is auto-selected.
 */
export function canSelectPaper(hit: PaperMetadata): boolean {
  return hit.content_state === 'readable_fulltext' || hit.content_state === 'abstract_only';
}

/** Explicit, human-only selection. Never selects a hit automatically. */
export function togglePaper(selected: string[], hit: PaperMetadata): string[] {
  return selected.includes(paperSearchKey(hit)) ? selected.filter(id => id !== paperSearchKey(hit)) : [...selected, paperSearchKey(hit)];
}

export function clearPaperSelection(): string[] {
  return [];
}

/** Hits the human has actually picked. */
export function pickedPapers(hits: PaperMetadata[], selected: string[]): PaperMetadata[] {
  return hits.filter(hit => selected.includes(paperSearchKey(hit)));
}

/** All selectable hits. */
export function eligiblePapers(hits: PaperMetadata[]): PaperMetadata[] {
  return hits.filter(canSelectPaper);
}

/** True when the human picked every currently selectable hit. */
export function isAllEligiblePicked(hits: PaperMetadata[], selected: string[]): boolean {
  const eligible = eligiblePapers(hits);
  return eligible.length > 0 && pickedPapers(hits, selected).length === eligible.length;
}
