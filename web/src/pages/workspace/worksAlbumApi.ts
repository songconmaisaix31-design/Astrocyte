import { api, ApiError, type ApiErrorBody } from '../../api/client';
import type { WorksMetadataEdit } from './WorksCommandForms';
import type { WorksWork } from './worksAlbumPresentation';

async function unwrap<T>(result: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
  const { data, error, response } = await result;
  if (error) throw new ApiError(response.status, error);
  if (!data) throw new Error('作品 API 未返回数据');
  return data;
}

async function headers(key: string) {
  const session = await unwrap(api.GET('/auth/session'));
  return { 'Idempotency-Key': key, 'X-CSRF-Token': session.csrf_token };
}
const command = (key: string) => ({ schema_version: 1 as const, request_id: key, expected_version: 1 as const });

// 人工提交具有唯一操作身份；读取复用既有会话，失败或结果未知的写入不自动重发。
export const worksAlbumApi = {
  list: (signal?: AbortSignal) => unwrap(api.GET('/works', { signal })),
  refresh: async (roots: string[], key: string) => unwrap(api.POST('/works/refresh', { body: { ...command(key), roots }, params: { header: await headers(key) } })),
  delivery: async (content: string, key: string) => unwrap(api.POST('/works/delivery', { body: { ...command(key), content }, params: { header: await headers(key) } })),
  async edit(work: WorksWork, edit: WorksMetadataEdit, key: string) {
    const params = { path: { work_id: work.id }, header: await headers(key) };
    const body = { ...command(key), expected_revision: work.revision };
    switch (edit.kind) {
      case 'attitude': return unwrap(api.PUT('/works/{work_id}/attitude', { params, body: { ...body, attitude: edit.value } }));
      case 'group': return unwrap(api.PUT('/works/{work_id}/group', { params, body: { ...body, group: edit.value } }));
      case 'handoff': return unwrap(api.PUT('/works/{work_id}/handoff', { params, body: { ...body, content: edit.value } }));
      case 'hidden': return unwrap(api.PUT('/works/{work_id}/hidden', { params, body: { ...body, hidden: edit.value } }));
    }
  },
};
