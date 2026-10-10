/**
 * Paper search client seam.
 *
 * The generated endpoint type does not exist yet (W0 owns `web/src/api/` and the
 * OpenAPI contract). This module calls the proposed `GET /papers/search` route
 * directly with a typed local shape; it never fabricates results and surfaces
 * every failure (404/501/network) as an explicit, retryable error so the panel
 * renders an honest "unavailable" state until the endpoint is wired.
 */
import type { PaperSearchHit, PaperSearchResult } from './paperSearch';

export const PAPER_SEARCH_PATH = '/api/v1/papers/search';

export class PaperSearchUnavailable extends Error {
  constructor(readonly status: number) {
    super(status === 404 || status === 501 ? '论文检索尚未接入（服务未提供该能力）' : `论文检索暂不可用（HTTP ${status}）`);
    this.name = 'PaperSearchUnavailable';
  }
}

interface RawHit {
  id?: unknown; title?: unknown; authors?: unknown; year?: unknown; venue?: unknown;
  source_type?: unknown; arxiv_id?: unknown; doi?: unknown; locator?: unknown;
  abstract?: unknown; availability?: unknown; already_imported?: unknown; import_material_id?: unknown;
}

function asString(value: unknown): string | null {
  return typeof value === 'string' ? value : null;
}
function asStringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : [];
}

function mapHit(raw: RawHit): PaperSearchHit {
  const availability = (raw.availability && typeof raw.availability === 'object' ? raw.availability : {}) as { status?: unknown; detail?: unknown };
  const status = availability.status;
  return {
    id: asString(raw.id) ?? '',
    title: asString(raw.title) ?? '标题未提供',
    authors: asStringArray(raw.authors),
    year: typeof raw.year === 'number' ? raw.year : null,
    venue: asString(raw.venue),
    source_type: asString(raw.source_type) ?? 'unknown',
    arxiv_id: asString(raw.arxiv_id),
    doi: asString(raw.doi),
    locator: asString(raw.locator) ?? '',
    abstract: asString(raw.abstract),
    availability: {
      status: status === 'full_text' || status === 'metadata_only' || status === 'restricted' || status === 'unknown' ? status : 'unknown',
      detail: asString(availability.detail),
    },
    already_imported: raw.already_imported === true,
    import_material_id: asString(raw.import_material_id),
  };
}

export async function searchPapers(query: string, signal?: AbortSignal): Promise<PaperSearchResult> {
  const params = new URLSearchParams({ q: query });
  const response = await fetch(`${PAPER_SEARCH_PATH}?${params}`, { credentials: 'same-origin', signal });
  if (!response.ok) throw new PaperSearchUnavailable(response.status);
  const body = (await response.json().catch(() => undefined)) as { query?: unknown; items?: unknown; next_cursor?: unknown; has_more?: unknown; warnings?: unknown } | undefined;
  const items = Array.isArray(body?.items) ? (body.items as RawHit[]).map(mapHit) : [];
  return {
    query: asString(body?.query) ?? query,
    items,
    next_cursor: asString(body?.next_cursor),
    has_more: body?.has_more === true,
    warnings: asStringArray(body?.warnings),
  };
}

/** Proposed batch import route; write never replays on failure (idempotency key is caller-owned). */
export const PAPER_IMPORT_PATH = '/api/v1/papers/import';

export interface PaperImportResult {
  job_id: string;
}

/**
 * Human-session batch import. Bootstraps the same-origin session for CSRF and
 * sends the caller's idempotency key exactly once. Any non-2xx (including a not
 * yet implemented 404/501) is surfaced verbatim so the panel never claims the
 * import succeeded.
 */
export async function importPapers(body: unknown, key: string, signal?: AbortSignal): Promise<PaperImportResult> {
  const sessionResponse = await fetch('/api/v1/auth/session', { credentials: 'same-origin', signal });
  if (!sessionResponse.ok) throw new Error('本地会话不可用，无法提交批量入库');
  const session = (await sessionResponse.json().catch(() => undefined)) as { csrf_token?: string } | undefined;
  const response = await fetch(PAPER_IMPORT_PATH, {
    method: 'POST',
    credentials: 'same-origin',
    signal,
    headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key, 'X-CSRF-Token': session?.csrf_token ?? '' },
    body: JSON.stringify(body),
  });
  if (!response.ok) throw new PaperSearchUnavailable(response.status);
  return (await response.json().catch(() => undefined)) as PaperImportResult;
}
