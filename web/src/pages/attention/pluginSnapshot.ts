/**
 * Plugin snapshot paste-review presentation (pure, no I/O).
 *
 * W1's MV3 extension copies a paper-page JSON snapshot (see extensions/paper/
 * extract.js and W1 `PaperSnapshot`). Its shape is: `source_url`, `host_family`,
 * `source_key`, `title`, `abstract`, `authors`, `doi`, `arxiv_id`,
 * `observed_version`, `pdf_urls`, `content_state`, `text`, `warning`,
 * `truncated`, `provenance`.
 *
 * Unlike an earlier revision, the raw `text` (the browser-extracted full body)
 * is PRESERVED: a readable full-text snapshot is imported through the minimal
 * `paper_snapshot` ImportMaterial entry carrying the whole reviewed snapshot
 * JSON as `export_text`, so the service does NOT re-fetch the page and never
 * trusts the snapshot's self-reported source key (it re-derives it from the
 * observed DOI/arXiv/URL identity). A `pdf_urls` entry imports through the
 * `paper_pdf` adapter (SSRF-safe public PDF fetch, not arxiv-only). The human
 * is offered an explicit HTML-fulltext vs public-PDF choice; the abstract, a
 * truncated body and a paywall are never treated as full text, and no personal
 * browser or history is read automatically.
 */

export type SnapshotContentState = 'readable_fulltext' | 'abstract_only' | 'paywall' | 'restricted';

/** Import adapters the existing ImportMaterial entry can consume. */
export type SnapshotImportAdapter = 'arxiv' | 'paper_pdf' | 'paper_snapshot';

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
  content_state: SnapshotContentState;
  text: string | null;
  warning: string | null;
  truncated: boolean;
}

/** One human-selectable import path for a reviewed snapshot. */
export interface SnapshotImportOption {
  adapter: SnapshotImportAdapter;
  /** What the existing ImportMaterial entry receives as source_locator. */
  source_locator: string;
  /** Raw reviewed snapshot JSON for paper_snapshot; null for PDF/arxiv. */
  export_text: string | null;
  label: string;
}

interface RawSnapshot {
  source_url?: unknown; host_family?: unknown; source_key?: unknown; title?: unknown;
  authors?: unknown; doi?: unknown; arxiv_id?: unknown; abstract?: unknown;
  observed_version?: unknown; pdf_urls?: unknown; content_state?: unknown;
  text?: unknown; warning?: unknown; truncated?: unknown;
}

function asString(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value.trim() : null;
}
function asStringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string' && item.trim().length > 0) : [];
}
const contentStates: SnapshotContentState[] = ['readable_fulltext', 'abstract_only', 'paywall', 'restricted'];
function asContentState(value: unknown): SnapshotContentState {
  return (contentStates as unknown[]).includes(value) ? (value as SnapshotContentState) : 'abstract_only';
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
    text: asString(snapshot.text),
    warning: asString(snapshot.warning),
    truncated: snapshot.truncated === true,
  };
}

function isArxiv(review: PluginSnapshotReview): boolean {
  if (review.host_family === 'arxiv' || review.arxiv_id) return true;
  try { return /(^|\.)arxiv\.org$/i.test(new URL(review.source_url).hostname); } catch { return false; }
}

/** Human label for a snapshot content state (domain vocabulary). */
export function snapshotContentStateLabel(state: SnapshotContentState): string {
  switch (state) {
    case 'readable_fulltext': return '原文可得';
    case 'abstract_only': return '仅摘要/元数据';
    case 'paywall': return '付费墙';
    case 'restricted': return '受限';
  }
}

/** A body is usable as full text only when full-text, untruncated and non-empty. */
function hasFullText(review: PluginSnapshotReview): boolean {
  return review.content_state === 'readable_fulltext' && !review.truncated && !!review.text;
}

/** Restricted/paywalled content never yields an importable full text. */
function isRestricted(review: PluginSnapshotReview): boolean {
  return review.content_state === 'paywall' || review.content_state === 'restricted';
}

/**
 * The human-selectable import paths for a reviewed snapshot: the preserved
 * HTML full body (`paper_snapshot`, no re-fetch, whole snapshot JSON as
 * `export_text`) and/or each public PDF (`paper_pdf`, SSRF-safe fetch, not
 * arxiv-only). An arxiv snapshot with neither falls back to the official
 * `arxiv` adapter. A paywall/restricted snapshot yields nothing (no bypass),
 * and an abstract-only snapshot without a PDF is not full text, so it yields
 * nothing either. `rawJson` is the exact reviewed snapshot JSON.
 */
export function snapshotImportOptions(review: PluginSnapshotReview, rawJson: string): SnapshotImportOption[] {
  const options: SnapshotImportOption[] = [];
  if (hasFullText(review)) {
    options.push({ adapter: 'paper_snapshot', source_locator: review.source_url, export_text: rawJson, label: '按快照正文导入（HTML 正文）' });
  }
  if (!isRestricted(review)) {
    for (const pdf of review.pdf_urls) {
      options.push({ adapter: 'paper_pdf', source_locator: pdf, export_text: null, label: '按公共 PDF 导入' });
    }
  }
  if (!options.length && isArxiv(review) && !isRestricted(review)) {
    options.push({ adapter: 'arxiv', source_locator: review.source_url, export_text: null, label: '按 arXiv 官方导入' });
  }
  return options;
}

/** Whether at least one non-bypassing import path exists for this snapshot. */
export function canImportSnapshot(review: PluginSnapshotReview, rawJson: string): boolean {
  return snapshotImportOptions(review, rawJson).length > 0;
}

/** Caveat: the abstract/truncated/paywall is never treated as the body. */
export const snapshotCaveat = '插件快照（含摘要）不是正文；只有快照携带的完整正文或公共 PDF 才能入库，正文由你复核后写入，服务不会重新抓取网页或读取你的浏览器。';
