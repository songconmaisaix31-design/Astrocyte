import type { PaperSearchResult } from './paperSearch';

/** Sample-only data for fixture mode; never presented as real search results. */
export const fixturePaperSearchResults: PaperSearchResult = {
  schema_version: 1,
  items: [
    { site: 'arxiv', source_key: 'arxiv:2504.16054', arxiv_id: '2504.16054', title: '示例论文：公开 arXiv 文献（用于验证检索与勾选流程）', authors: ['示例作者甲', '示例作者乙'], abstract: '示例摘要文字；检索仅返回公开元数据。', published_at: '2024-01-01', locator: 'https://arxiv.org/abs/2504.16054', content_state: 'abstract_only' },
    { site: 'crossref', source_key: 'doi:10.0000/example', doi: '10.0000/example', title: '示例期刊论文：经 DOI 元数据索引', authors: ['示例作者丙'], abstract: '示例摘要文字。', published_at: '2023-01-01', locator: 'https://doi.org/10.0000/example', content_state: 'abstract_only' },
  ],
  warnings: [],
};
