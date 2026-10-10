import type { PaperSearchResult } from './paperSearch';

/** Sample-only data for fixture mode; never presented as real search results. */
export const fixturePaperSearchResults: PaperSearchResult = {
  schema_version: 1,
  query: '示例检索',
  items: [
    { source_key: 'arxiv:2504.16054', provider: 'arxiv', title: '示例论文：公开 arXiv 文献（用于验证检索与勾选流程）', authors: ['示例作者甲', '示例作者乙'], year: 2024, venue: 'arXiv', arxiv_id: '2504.16054', abstract: '示例摘要文字；检索仅返回公开元数据。', locator: 'https://arxiv.org/abs/2504.16054', content_state: 'abstract_only', pdf_urls: [] },
    { source_key: 'doi:10.0000/example', provider: 'crossref', title: '示例期刊论文：经 DOI 元数据索引', authors: ['示例作者丙'], year: 2023, doi: '10.0000/example', abstract: '示例摘要文字。', locator: 'https://doi.org/10.0000/example', content_state: 'abstract_only', pdf_urls: ['https://example.com/example.pdf'] },
  ],
  next_cursor: null,
  has_more: false,
  warnings: [],
};
