import { api, ApiError, attentionApi } from '../../api/client';
import { paperImportAdapter, type PaperSearchHit, type PaperSearchResult } from './paperSearch';

export async function searchPapers(query: string, signal?: AbortSignal): Promise<PaperSearchResult> {
  const { data, error, response } = await api.GET('/papers/search', { params: { query: { q: query } }, signal });
  if (error) throw new ApiError(response.status, error);
  if (!data || !Array.isArray(data.items)) throw new Error('论文检索响应缺少 items');
  return data;
}

/** 每份人工选定的资料单独提交，作业结果由处理队列提供。 */
export async function importPaperHit(hit: PaperSearchHit, reason: string | null) {
  const key = crypto.randomUUID();
  return attentionApi.importMaterial({
    schema_version: 1, request_id: key, expected_version: 1,
    adapter: paperImportAdapter(hit), source_locator: hit.locator, source_key: hit.source_key,
    kind: 'paper', content_digest: '', collection_reason: reason, title: hit.title, refresh: false,
  }, key);
}

/** 保留已接受的作业并报告全部提交失败，禁止重发结果未知的命令。 */
export async function importPaperBatch(hits: PaperSearchHit[], reason: string | null): Promise<void> {
  const settled = await Promise.allSettled(hits.map(hit => importPaperHit(hit, reason)));
  const rejected = settled.filter((result): result is PromiseRejectedResult => result.status === 'rejected');
  if (!rejected.length) return;
  const accepted = settled.length - rejected.length;
  const first = rejected[0].reason;
  const detail = first instanceof Error ? first.message : String(first);
  throw new Error(accepted ? `已提交 ${accepted} 条入库作业，${rejected.length} 条提交失败（详情见处理队列）：${detail}` : `批量入库提交失败（详情见处理队列）：${detail}`);
}
