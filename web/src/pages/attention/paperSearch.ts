/** 论文检索的纯显示类型；公开元数据、来源身份和可得状态均来自服务。 */

/** 原始可得状态；元数据索引只说明摘要与元数据可得。 */
export type PaperContentState = 'readable_fulltext' | 'abstract_only' | 'paywall' | 'restricted' | 'unknown';

/** 对应生成接口 PaperSearchHitV1 的显示类型。 */
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

/** 对应生成接口 PaperSearchResultV1 的显示类型。 */
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

/** 使用服务给出的 source_key 识别检索结果。 */
export const paperSearchKey = (hit: PaperSearchHit): string => hit.source_key || hit.locator;

function isArxiv(hit: PaperSearchHit): boolean {
  return !!hit.arxiv_id || hit.provider === 'arxiv' || /(^|\.)arxiv\.org$/i.test(hit.locator || '');
}

/** arXiv 来源使用官方导入器；其他来源使用 paper_url。 */
export function paperImportAdapter(hit: PaperSearchHit): 'arxiv' | 'paper_url' {
  return isArxiv(hit) ? 'arxiv' : 'paper_url';
}

/** 付费和受限来源禁用选择；摘要与未知来源的实际正文由人工提交后的作业核实。 */
export function canSelectPaper(hit: PaperSearchHit): boolean {
  return hit.content_state !== 'paywall' && hit.content_state !== 'restricted';
}

/** 只切换人工明确选择的结果。 */
export function togglePaper(selected: string[], hit: PaperSearchHit): string[] {
  return selected.includes(paperSearchKey(hit)) ? selected.filter(id => id !== paperSearchKey(hit)) : [...selected, paperSearchKey(hit)];
}

export function clearPaperSelection(): string[] {
  return [];
}

/** 取得人工已选择的结果。 */
export function pickedPapers(hits: PaperSearchHit[], selected: string[]): PaperSearchHit[] {
  return hits.filter(hit => selected.includes(paperSearchKey(hit)));
}

/** 取得当前允许选择的结果。 */
export function eligiblePapers(hits: PaperSearchHit[]): PaperSearchHit[] {
  return hits.filter(canSelectPaper);
}

/** 判断人工是否已选择全部可选结果。 */
export function isAllEligiblePicked(hits: PaperSearchHit[], selected: string[]): boolean {
  const eligible = eligiblePapers(hits);
  return eligible.length > 0 && pickedPapers(hits, selected).length === eligible.length;
}
