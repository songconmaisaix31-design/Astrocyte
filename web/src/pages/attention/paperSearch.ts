/**
 * Paper search presentation model (pure, no I/O).
 *
 * The paper search endpoint and its generated DTO are owned by W0/W1 and are
 * not yet present in `web/src/api/schema.d.ts`. Until then the UI consumes the
 * local view shape below; the panel maps the real endpoint onto it once W0
 * publishes the generated client. No fixture or guessed value is ever
 * presented as a real search result.
 */

export type PaperAvailability = 'full_text' | 'metadata_only' | 'restricted' | 'unknown';

/** Local view shape for one search hit; mirrors the proposed W0/W1 DTO. */
export interface PaperSearchHit {
  /** Stable source identity (arXiv/DOI-derived); used as the selection key. */
  id: string;
  title: string;
  authors: string[];
  year: number | null;
  venue: string | null;
  source_type: string;
  arxiv_id: string | null;
  doi: string | null;
  locator: string;
  abstract: string | null;
  availability: { status: PaperAvailability; detail: string | null };
  /** Already stored as a material; cannot be re-imported from this list. */
  already_imported: boolean;
  import_material_id: string | null;
}

export interface PaperSearchResult {
  query: string;
  items: PaperSearchHit[];
  next_cursor: string | null;
  has_more: boolean;
  warnings: string[];
}

export const paperAvailabilityLabel = (status: PaperAvailability): string => {
  switch (status) {
    case 'full_text': return '原文可得';
    case 'metadata_only': return '仅元数据';
    case 'restricted': return '受限';
    case 'unknown': return '可得性未知';
  }
};

export const paperSearchKey = (hit: PaperSearchHit): string => hit.id;

/**
 * A hit can be picked for import only when it is not already stored and its
 * availability is not unknown. Restricted and metadata-only hits remain
 * selectable so the human can see the restriction note and decide; the server
 * is the authority on what a restricted hit actually imports.
 */
export function canSelectPaper(hit: PaperSearchHit): boolean {
  return !hit.already_imported && hit.availability.status !== 'unknown';
}

/** Explicit, human-only selection. Never selects a hit automatically. */
export function togglePaper(selected: string[], hit: PaperSearchHit): string[] {
  return selected.includes(hit.id) ? selected.filter(id => id !== hit.id) : [...selected, hit.id];
}

export function clearPaperSelection(): string[] {
  return [];
}

/** Hits the human has actually picked and that remain eligible. */
export function pickedPapers(hits: PaperSearchHit[], selected: string[]): PaperSearchHit[] {
  return hits.filter(hit => selected.includes(hit.id) && canSelectPaper(hit));
}

/** Eligible hits that are not yet picked; used to avoid "import everything" silently. */
export function eligiblePapers(hits: PaperSearchHit[]): PaperSearchHit[] {
  return hits.filter(canSelectPaper);
}

/** True when the human picked every currently eligible hit. */
export function isAllEligiblePicked(hits: PaperSearchHit[], selected: string[]): boolean {
  const eligible = eligiblePapers(hits);
  return eligible.length > 0 && pickedPapers(hits, selected).length === eligible.length;
}

/** First unavailable (unknown) hit, if any; shown so the human can wait/retry. */
export function firstUnknownHit(hits: PaperSearchHit[]): PaperSearchHit | undefined {
  return hits.find(hit => hit.availability.status === 'unknown');
}
