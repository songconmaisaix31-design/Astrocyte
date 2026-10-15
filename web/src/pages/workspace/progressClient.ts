import { api, ApiError, type ApiErrorBody, type components } from '../../api/client';
import type { ProjectProgressResult } from './progressPresentation';

type Result = components['schemas']['ProjectProgressResultV1'];

async function unwrap(result: Promise<{ data?: Result; error?: ApiErrorBody; response: Response }>): Promise<ProjectProgressResult> {
  const { data, error, response } = await result;
  if (error) throw new ApiError(response.status, error);
  if (!data?.progress) throw new Error('项目进度响应缺少 progress');
  return data;
}

export function getProjectProgress(id: string, signal?: AbortSignal): Promise<ProjectProgressResult> {
  return unwrap(api.GET('/local-projects/{id}/progress', { params: { path: { id } }, signal }));
}

/** 人工记录沿既有会话与生成契约保存，不改变项目或模型许可。 */
export async function setProjectProgress(id: string, body: components['schemas']['SetProgressRequestV1'], key: string): Promise<ProjectProgressResult> {
  if (!key.trim() || body.request_id !== key) throw new Error('人工进度请求身份必须与命令标识一致');
  const { data, error, response } = await api.GET('/auth/session');
  if (error) throw new ApiError(response.status, error);
  if (!data?.csrf_token) throw new Error('本地会话响应缺少 CSRF token');
  return unwrap(api.PUT('/local-projects/{id}/progress', {
    params: { path: { id }, header: { 'Idempotency-Key': key, 'X-CSRF-Token': data.csrf_token } }, body,
  }));
}

/** 使用既有会话和生成接口发送一次人工推断命令。 */
export async function inferProjectProgress(id: string, key: string, files: string[], signal?: AbortSignal): Promise<ProjectProgressResult> {
  if (!key.trim()) throw new Error('Idempotency key is required');
  const { data, error, response } = await api.GET('/auth/session', { signal });
  if (error) throw new ApiError(response.status, error);
  if (!data?.csrf_token) throw new Error('本地会话响应缺少 CSRF token');
  return unwrap(api.POST('/local-projects/{id}/progress/infer', {
    params: { path: { id }, header: { 'Idempotency-Key': key, 'X-CSRF-Token': data.csrf_token } },
    body: { schema_version: 1, request_id: key, expected_version: 1, files }, signal,
  }));
}
