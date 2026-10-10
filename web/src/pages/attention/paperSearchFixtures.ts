import type { PaperSearchResult } from './paperSearch';

/** Sample-only data for fixture mode; never presented as real search results. */
export const fixturePaperSearchResults: PaperSearchResult = {
  query: '示例检索',
  items: [
    { id: 'fixture:paper-full', title: '示例论文：全文可得的公开文献（用于验证检索与勾选流程）', authors: ['示例作者甲', '示例作者乙'], year: 2024, venue: 'arXiv', source_type: 'arxiv', arxiv_id: '2504.00000', doi: null, locator: 'fixture:paper-full', abstract: '示例摘要文字。', availability: { status: 'full_text', detail: null }, already_imported: false, import_material_id: null },
    { id: 'fixture:paper-restricted', title: '示例受限论文：付费墙，仅能获取元数据', authors: ['示例作者丙'], year: 2023, venue: '某期刊', source_type: 'doi', arxiv_id: null, doi: '10.0000/example', locator: 'fixture:paper-restricted', abstract: null, availability: { status: 'restricted', detail: '来源需要订阅或付费，正文不可得，仅保留元数据。' }, already_imported: false, import_material_id: null },
    { id: 'fixture:paper-unknown', title: '示例可得性未知论文：等待确认原文状态', authors: ['示例作者丁'], year: null, venue: null, source_type: 'doi', arxiv_id: null, doi: '10.0000/unknown', locator: 'fixture:paper-unknown', abstract: null, availability: { status: 'unknown', detail: '原文可得性尚未确认，需重试。' }, already_imported: false, import_material_id: null },
    { id: 'fixture:paper-imported', title: '示例已入库论文：不能再从检索重复导入', authors: ['示例作者戊'], year: 2022, venue: 'arXiv', source_type: 'arxiv', arxiv_id: '2201.00000', doi: null, locator: 'fixture:paper-imported', abstract: null, availability: { status: 'full_text', detail: null }, already_imported: true, import_material_id: 'mat-fixture-imported' },
  ],
  next_cursor: null,
  has_more: false,
  warnings: [],
};
