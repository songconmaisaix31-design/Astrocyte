/**
 * Paper search client seam.
 *
 * W0 published `POST /api/v1/paper/search` (see contracts/openapi.yaml); the
 * generated `attentionApi.searchPapers` wrapper is added by W0 during final
 * integration because `web/src/api/` is W0-owned. Until then this module calls
 * the published route directly with the exact request/response shapes of W0's
 * `PaperSearchRequestV1`/`PaperSearchResultV1` and never fabricates results.
 * Every failure (404/501/network) surfaces as an explicit, retryable error.
 */
import { attentionApi } from '../../api/client';
import type { components } from '../../api/schema';
import { paperImportAdapter, type PaperContentState, type PaperImportAdapter, type PaperMetadata, type PaperSearchResult } from './paperSearch';

export const PAPER_SEARCH_PATH = '/api/v1/paper/search';

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
const contentStates: PaperContentState[] = ['readable_fulltext', 'abstract_only', 'paywall', 'restricted'];
function asContentState(value: unknown): PaperContentState {
  return (contentStates as unknown[]).includes(value) ? (value as PaperContentState) : 'abstract_only';
}

interface RawHit {
  site?: unknown; source_key?: unknown; doi?: unknown; arxiv_id?: unknown; title?: unknown;
  authors?: unknown; abstract?: unknown; published_at?: unknown; locator?: unknown;
  content_state?: unknown; license?: unknown; warning?: unknown;
}

function mapHit(raw: RawHit): PaperMetadata {
  return {
    site: asString(raw.site) ?? 'unknown',
    source_key: asString(raw.source_key) ?? asString(raw.locator) ?? '',
    doi: asString(raw.doi) ?? undefined,
    arxiv_id: asString(raw.arxiv_id) ?? undefined,
    title: asString(raw.title) ?? '标题未提供',
    authors: asStringArray(raw.authors),
    abstract: asString(raw.abstract) ?? undefined,
    published_at: asString(raw.published_at) ?? undefined,
    locator: asString(raw.locator) ?? '',
    content_state: asContentState(raw.content_state),
    license: asString(raw.license) ?? undefined,
    warning: asString(raw.warning) ?? undefined,
  };
}

/** Deterministic idempotency key so a retry of the same query reuses it. */
function stableKey(query: string): string {
  let h = 0x811c9dc5;
  for (let i = 0; i < query.length; i++) {
    h ^= query.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return `paper-search-${(h >>> 0).toString(16)}`;
}

export async function searchPapers(query: string, signal?: AbortSignal): Promise<PaperSearchResult> {
  const key = stableKey(query);
  const sessionResponse = await fetch('/api/v1/auth/session', { credentials: 'same-origin', signal });
  if (!sessionResponse.ok) throw new PaperSearchUnavailable(sessionResponse.status);
  const session = (await sessionResponse.json().catch(() => undefined)) as { csrf_token?: string } | undefined;
  const response = await fetch(PAPER_SEARCH_PATH, {
    method: 'POST',
    credentials: 'same-origin',
    signal,
    headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key, 'X-CSRF-Token': session?.csrf_token ?? '' },
    body: JSON.stringify({ schema_version: 1, request_id: key, expected_version: 1, query }),
  });
  if (!response.ok) throw new PaperSearchUnavailable(response.status);
  const body = (await response.json().catch(() => undefined)) as { schema_version?: unknown; items?: unknown; warnings?: unknown } | undefined;
  const items = Array.isArray(body?.items) ? (body.items as RawHit[]).map(mapHit) : [];
  return { schema_version: 1, items, warnings: asStringArray(body?.warnings) };
}

/**
 * A selected hit is imported through the existing ImportMaterial path (owned by
 * W0). `paper_url` is W1's real body-extraction adapter; it is not yet in W0's
 * generated adapter enum, so it is cast at this single seam for W0 to confirm
 * during final integration.
 */
const toImportAdapter = (adapter: PaperImportAdapter): components['schemas']['ImportMaterialRequestV1']['adapter'] =>
  adapter as components['schemas']['ImportMaterialRequestV1']['adapter'];

export async function importPaperHit(hit: PaperMetadata, reason: string | null, key: string) {
  return attentionApi.importMaterial({
    schema_version: 1,
    request_id: key,
    expected_version: 1,
    adapter: toImportAdapter(paperImportAdapter(hit)),
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
 * its own idempotency key. Every import is attempted; failures are surfaced as a
 * single actionable error and the successful jobs remain queryable in the job
 * queue. Never imports an unselected hit.
 */
export async function importPaperBatch(hits: PaperMetadata[], reason: string | null): Promise<void> {
  const settled = await Promise.allSettled(hits.map(hit => importPaperHit(hit, reason, crypto.randomUUID())));
  const rejected = settled.filter((result): result is PromiseRejectedResult => result.status === 'rejected');
  if (!rejected.length) return;
  const ok = settled.length - rejected.length;
  const first = rejected[0].reason;
  const detail = first instanceof Error ? first.message : String(first);
  throw new Error(ok ? `已提交 ${ok} 条，${rejected.length} 条入库失败（详情见处理队列）：${detail}` : `批量入库失败（详情见处理队列）：${detail}`);
}
