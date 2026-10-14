/** 复核人工粘贴的插件快照；保留完整 JSON 与正文，明确选择 HTML、公共 PDF 或 arXiv 导入。 */

export type SnapshotContentState = 'readable_fulltext' | 'abstract_only' | 'paywall' | 'restricted' | 'unknown';

/** 既有 ImportMaterial 接受的导入器。 */
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

/** 人工可选择的快照导入路径。 */
export interface SnapshotImportOption {
  adapter: SnapshotImportAdapter;
  /** 传给既有接口的 source_locator。 */
  source_locator: string;
  /** paper_snapshot 保留已复核 JSON；PDF 和 arXiv 使用 null。 */
  export_text: string | null;
  label: string;
}

interface RawSnapshot {
  schema_version?: unknown;
  source_url?: unknown; host_family?: unknown; source_key?: unknown; title?: unknown;
  authors?: unknown; doi?: unknown; arxiv_id?: unknown; abstract?: unknown;
  observed_version?: unknown; pdf_urls?: unknown; content_state?: unknown;
  text?: unknown; warning?: unknown; truncated?: unknown;
}

function asString(value: unknown): string | null {
  if (value === null || value === undefined) return null;
  if (typeof value !== 'string') throw new Error('插件快照的文本字段类型无效。');
  return value.trim() || null;
}
function asStringArray(value: unknown): string[] {
  if (value === null || value === undefined) return [];
  if (!Array.isArray(value) || value.some(item => typeof item !== 'string' || !item.trim())) throw new Error('插件快照的文本列表字段类型无效。');
  return value;
}
const contentStates: SnapshotContentState[] = ['readable_fulltext', 'abstract_only', 'paywall', 'restricted', 'unknown'];
function asContentState(value: unknown): SnapshotContentState {
  if (value === null || value === undefined) return 'unknown';
  if (!(contentStates as unknown[]).includes(value)) throw new Error('插件快照的原文可得状态无效。');
  return value as SnapshotContentState;
}

/** 解析粘贴快照；JSON 无效或缺少来源时立即报告错误。 */
export function parsePluginSnapshot(text: string): PluginSnapshotReview {
  let raw: unknown;
  try { raw = JSON.parse(text); } catch { throw new Error('无法解析插件快照：不是有效的 JSON。请重新粘贴扩展复制的完整快照。'); }
  const snapshot = (raw && typeof raw === 'object' ? raw : {}) as RawSnapshot;
  const sourceUrl = asString(snapshot.source_url);
  if (!sourceUrl) throw new Error('插件快照缺少来源链接（source_url）。请粘贴扩展复制的完整快照，不要只粘贴摘要文字。');
  if (snapshot.schema_version !== 1) throw new Error('插件快照协议版本无效（schema_version 必须为 1）。请粘贴扩展复制的完整快照。');
  if (snapshot.truncated !== undefined && typeof snapshot.truncated !== 'boolean') throw new Error('插件快照的截断标识类型无效。');
  if (snapshot.text !== undefined && snapshot.text !== null && typeof snapshot.text !== 'string') throw new Error('插件快照正文类型无效。');
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
    text: typeof snapshot.text === 'string' && snapshot.text.trim() ? snapshot.text : null,
    warning: asString(snapshot.warning),
    truncated: snapshot.truncated === true,
  };
}

function isArxiv(review: PluginSnapshotReview): boolean {
  if (review.host_family === 'arxiv' || review.arxiv_id) return true;
  return URL.canParse(review.source_url) && /(^|\.)arxiv\.org$/i.test(new URL(review.source_url).hostname);
}

/** 可得状态缺少观察时保留未知。 */
export function snapshotContentStateLabel(state: SnapshotContentState): string {
  switch (state) {
    case 'readable_fulltext': return '原文可得';
    case 'abstract_only': return '仅摘要/元数据';
    case 'paywall': return '付费墙';
    case 'restricted': return '受限';
    case 'unknown': return '可得性未知';
  }
}

/** 仅接受明确可读、完整且未截断的正文。 */
function hasFullText(review: PluginSnapshotReview): boolean {
  return review.content_state === 'readable_fulltext' && !review.truncated && !!review.text;
}

/** 付费与受限来源不生成完整正文导入选项。 */
function isRestricted(review: PluginSnapshotReview): boolean {
  return review.content_state === 'paywall' || review.content_state === 'restricted';
}

/** 提供人工明确选择的导入器；受限来源不生成选项，完整 JSON 保留为输入。 */
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

/** 判断快照是否具有可选择的导入路径。 */
export function canImportSnapshot(review: PluginSnapshotReview, rawJson: string): boolean {
  return snapshotImportOptions(review, rawJson).length > 0;
}

/** 摘要、截断正文与付费页面不构成完整正文。 */
export const snapshotCaveat = '插件快照（含摘要）不是正文；只有快照携带的完整正文或公共 PDF 才能入库，正文由你复核后写入，服务不会重新抓取网页或读取你的浏览器。';
