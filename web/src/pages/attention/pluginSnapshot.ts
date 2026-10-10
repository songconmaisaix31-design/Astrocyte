/**
 * Plugin snapshot paste-review presentation (pure, no I/O).
 *
 * W1's MV3 extension copies a paper-page JSON snapshot (see extensions/paper/
 * extract.js) to the clipboard. Its shape is: `source_url`, `host_family`,
 * `source_key`, `title`, `abstract`, `authors`, `doi`, `arxiv_id`,
 * `observed_version`, `pdf_urls`, `content_state`, `text`, `warning`,
 * `truncated`, `provenance`. The human pastes it into this same-origin review;
 * the app only reads metadata and the canonical locator. The snapshot (or its
 * abstract) is NEVER trusted as the full body — the service re-extracts the body
 * from the locator on import.
 */

import type { PaperContentState, PaperImportAdapter } from './paperSearch';

export interface PluginSnapshotReview {
  source_url: string;
  host_family: string;
  source_key: string;
  title: string | null;
  authors: string[];
  doi: string | null;
  arxiv_id: string | null;
  abstract: string | null;
  observed_version: string | null;
  pdf_urls: string[];
  content_state: PaperContentState;
  warning: string | null;
  truncated: boolean;
}

interface RawSnapshot {
  source_url?: unknown; host_family?: unknown; source_key?: unknown; title?: unknown;
  authors?: unknown; doi?: unknown; arxiv_id?: unknown; abstract?: unknown;
  observed_version?: unknown; pdf_urls?: unknown; content_state?: unknown;
  warning?: unknown; truncated?: unknown;
}

function asString(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value.trim() : null;
}
function asStringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string' && item.trim().length > 0) : [];
}
const contentStates: PaperContentState[] = ['readable_fulltext', 'abstract_only', 'paywall', 'restricted'];
function asContentState(value: unknown): PaperContentState {
  return (contentStates as unknown[]).includes(value) ? (value as PaperContentState) : 'abstract_only';
}

/** Parse a pasted snapshot. Throws with a readable message on invalid input. */
export function parsePluginSnapshot(text: string): PluginSnapshotReview {
  let raw: unknown;
  try { raw = JSON.parse(text); } catch { throw new Error('无法解析插件快照：不是有效的 JSON。请重新粘贴扩展复制的完整快照。'); }
  const snapshot = (raw && typeof raw === 'object' ? raw : {}) as RawSnapshot;
  const sourceUrl = asString(snapshot.source_url);
  if (!sourceUrl) throw new Error('插件快照缺少来源链接（source_url）。请粘贴扩展复制的完整快照，不要只粘贴摘要文字。');
  return {
    source_url: sourceUrl,
    host_family: asString(snapshot.host_family) ?? 'generic',
    source_key: asString(snapshot.source_key) ?? sourceUrl,
    title: asString(snapshot.title),
    authors: asStringArray(snapshot.authors),
    doi: asString(snapshot.doi),
    arxiv_id: asString(snapshot.arxiv_id),
    abstract: asString(snapshot.abstract),
    observed_version: asString(snapshot.observed_version),
    pdf_urls: asStringArray(snapshot.pdf_urls),
    content_state: asContentState(snapshot.content_state),
    warning: asString(snapshot.warning),
    truncated: snapshot.truncated === true,
  };
}

function locatorHostname(locator: string): string | null {
  try { return new URL(locator).hostname; } catch { return null; }
}

/**
 * Which existing import adapter can consume this snapshot. arXiv snapshots use
 * the `arxiv` adapter (official Atom/PDF); every other public HTTPS page uses
 * W1's `paper_url` body-extraction adapter. Returns null only for an unrecognized
 * source so the UI can honestly gate it.
 */
export function importAdapterFor(review: PluginSnapshotReview): PaperImportAdapter | null {
  if (review.host_family === 'arxiv' || review.arxiv_id) return 'arxiv';
  const hostname = locatorHostname(review.source_url);
  if (hostname && /(^|\.)arxiv\.org$/i.test(hostname)) return 'arxiv';
  return 'paper_url';
}

/** A snapshot whose page is paywalled/restricted cannot be imported without bypass. */
export function canImportSnapshot(review: PluginSnapshotReview): boolean {
  return review.content_state === 'readable_fulltext' || review.content_state === 'abstract_only';
}

/** Always-true caveat: the snapshot/abstract is never treated as the body. */
export const snapshotCaveat = '插件快照（含摘要）不是正文。导入时由服务按来源重新获取正文与版本，快照仅用于人工复核与选择。';
