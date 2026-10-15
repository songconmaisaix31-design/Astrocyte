import type { RepositoryQueryRequest } from '../../../api/repositoryContext';
import type { RepositoryQueryDraft } from './RepositoryQueryFields';

export function newRepositoryDraft(): RepositoryQueryDraft {
  return { purpose: '', roots: '', query: '', symbolId: '', maxFiles: '', maxBytes: '', maxResults: '', maxDepth: '' };
}

export function repositoryRequest(draft: RepositoryQueryDraft): RepositoryQueryRequest {
  const count = (value: string, label: string, maximum: number) => {
    const parsed = Number(value);
    if (!value.trim() || !Number.isSafeInteger(parsed) || parsed < 1 || parsed > maximum) throw new Error(`${label}必须为 1 至 ${maximum} 的整数`);
    return parsed;
  };
  const purpose = draft.purpose.trim();
  if (!purpose || new TextEncoder().encode(purpose).length > 512) throw new Error('请填写不超过 512 UTF-8 字节的查询用途');
  const roots = draft.roots.split(/\r?\n/).map(value => value.trim()).filter(Boolean);
  if (!roots.length || roots.length > 32) throw new Error('请填写 1 至 32 项明确的项目相对路径');
  return {
    schema_version: 1, purpose, roots,
    budget: { max_files: count(draft.maxFiles, '最多读取文件数', 20000), max_bytes: count(draft.maxBytes, '最多读取字节数', 134217728), max_results: count(draft.maxResults, '最多返回结果数', 200), max_depth: count(draft.maxDepth, '关系展开深度', 5) },
  };
}

export function boundedQuery(value: string, label: string, maximum: number) {
  const text = value.trim();
  if (!text || new TextEncoder().encode(text).length > maximum) throw new Error(`请填写不超过 ${maximum} UTF-8 字节的${label}`);
  return text;
}
