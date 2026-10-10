/**
 * Paper search client seam.
 *
 * W0 published `GET /api/v1/papers/search?q=&provider=&limit=` returning
 * `PaperSearchResultV1` (contracts/openapi.yaml); the generated
 * `attentionApi.searchPapers` wrapper is added by W0 during final integration
 * because `web/src/api/` is W0-owned. Until then this module calls the
 * published route directly with the exact wire shape and never fabricates
 * results. Every failure (404/501/network) surfaces as an explicit, retryable
 * error.
 *
 * Selection is imported per-item through the existing ImportMaterial path
 * (kind=paper, adapter arxiv|paper_url, source_key dedup, new revision). Each
 * item carries its own freshly generated Idempotency-Key; the backend's
 * source/content dedup (not a client hash) prevents a duplicate revision when
 * the same paper is submitted again, and a changed collection_reason is a new
 * body with a new key, never a 409. A settled submission only proves the job
 * was accepted, never that it succeeded; jobs stay queryable in the existing
 * processing queue.
 */
import { attentionApi } from '../../api/client';
import { paperImportAdapter, type PaperContentState, type PaperSearchHit, type PaperSearchResult } from './paperSearch';

export const PAPER_SEARCH_PATH = '/api/v1/papers/search';

export class PaperSearchUnavailable extends Error {
  constructor(readonly status: number) {
    super(status === 404 || status === 501 ? '论文检索尚未接入（服务未提供该能力）' : `论文检索暂不可用（HTTP ${status}）`);
    this.name = 'PaperSearchUnavailable';
  }
}

function asString(value: unknown): string | null {
  return typeof value === 'string' ? value : null;
}
function asStringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : [];
}
const contentStates: PaperContentState[] = ['readable_fulltext', 'abstract_only', 'paywall', 'restricted', 'unknown'];
function asContentState(value: unknown): PaperContentState {
  return (contentStates as unknown[]).includes(value) ? (value as PaperContentState) : 'abstract_only';
}

interface RawHit {
  source_key?: unknown; provider?: unknown; title?: unknown; authors?: unknown; year?: unknown;
  venue?: unknown; arxiv_id?: unknown; doi?: unknown; locator?: unknown;
  abstract?: unknown; content_state?: unknown; pdf_urls?: unknown;
}

function mapHit(raw: RawHit): PaperSearchHit {
  return {
    source_key: asString(raw.source_key) ?? '',
    provider: asString(raw.provider) ?? 'unknown',
    title: asString(raw.title) ?? '标题未提供',
    authors: asStringArray(raw.authors),
    year: typeof raw.year === 'number' ? raw.year : 0,
    venue: asString(raw.venue) ?? undefined,
    arxiv_id: asString(raw.arxiv_id) ?? undefined,
    doi: asString(raw.doi) ?? undefined,
    locator: asString(raw.locator) ?? '',
    abstract: asString(raw.abstract) ?? undefined,
    content_state: asContentState(raw.content_state),
    pdf_urls: asStringArray(raw.pdf_urls),
  };
}

export async function searchPapers(query: string, signal?: AbortSignal): Promise<PaperSearchResult> {
  const params = new URLSearchParams({ q: query });
  const response = await fetch(`${PAPER_SEARCH_PATH}?${params.toString()}`, { credentials: 'same-origin', signal });
  if (!response.ok) throw new PaperSearchUnavailable(response.status);
  const body = (await response.json().catch(() => undefined)) as { schema_version?: unknown; query?: unknown; items?: unknown; next_cursor?: unknown; has_more?: unknown; warnings?: unknown } | undefined;
  const items = Array.isArray(body?.items) ? (body.items as RawHit[]).map(mapHit) : [];
  return {
    schema_version: 1,
    query: asString(body?.query) ?? query,
    items,
    next_cursor: asString(body?.next_cursor),
    has_more: body?.has_more === true,
    warnings: asStringArray(body?.warnings),
  };
}

/**
 * A selected hit is imported through the existing ImportMaterial path with a
 * fresh idempotency key per item. The backend's source_key dedup keeps a
 * duplicate revision from being created on a retry; a changed reason is a new
 * body with a new key.
 */
export async function importPaperHit(hit: PaperSearchHit, reason: string | null) {
  const adapter = paperImportAdapter(hit);
  const key = crypto.randomUUID();
  return attentionApi.importMaterial({
    schema_version: 1,
    request_id: key,
    expected_version: 1,
    adapter,
    source_locator: hit.locator,
    source_key: hit.source_key,
    kind: 'paper',
    content_digest: '',
    collection_reason: reason,
    title: hit.title,
    refresh: false,
  }, key);
}

/**
 * Human-session batch import: one ImportMaterial job per selected hit, each with
 * its own idempotency key. Every selection is attempted; a settled submission
 * only means the job was accepted (not that it succeeded). Failures are surfaced
 * as a single actionable error, and accepted jobs remain queryable in the
 * existing processing queue. Never imports an unselected hit.
 */
export async function importPaperBatch(hits: PaperSearchHit[], reason: string | null): Promise<void> {
  const settled = await Promise.allSettled(hits.map(hit => importPaperHit(hit, reason)));
  const rejected = settled.filter((result): result is PromiseRejectedResult => result.status === 'rejected');
  if (!rejected.length) return;
  const accepted = settled.length - rejected.length;
  const first = rejected[0].reason;
  const detail = first instanceof Error ? first.message : String(first);
  throw new Error(accepted ? `已提交 ${accepted} 条入库作业，${rejected.length} 条提交失败（详情见处理队列）：${detail}` : `批量入库提交失败（详情见处理队列）：${detail}`);
}
