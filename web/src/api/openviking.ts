import { api, ApiError, type ApiErrorBody, type components } from './client';

type S = components['schemas'];
export type VikingSubmission = S['OpenVikingSubmissionV1'];
export type VikingHit = S['OpenVikingHitV1'];
export function createOpenVikingApi(client = api) {
  async function unwrap<T>(request: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('OpenViking 服务未返回数据');
    return data;
  }
  async function headers() {
    const session = await unwrap(client.GET('/auth/session'));
    return { 'X-CSRF-Token': session.csrf_token };
  }
  return {
    status: (signal?: AbortSignal) => unwrap(client.GET('/openviking/status', { signal })),
    list: (id: string, signal?: AbortSignal) => unwrap(client.GET('/project-spaces/{id}/openviking/submissions', { params: { path: { id } }, signal })),
    submit: async (id: string, body: S['OpenVikingSubmitRequestV1'], key: string) => {
      if (body.request_id !== key) throw new Error('外发请求身份必须与幂等键一致');
      return unwrap(client.POST('/project-spaces/{id}/openviking/submissions', { params: { path: { id }, header: { ...await headers(), 'Idempotency-Key': key } }, body }));
    },
    refresh: async (id: string) => unwrap(client.POST('/openviking/submissions/{id}/refresh', { params: { path: { id }, header: await headers() } })),
    search: async (id: string, query: string) => unwrap(client.POST('/project-spaces/{id}/openviking/search', { params: { path: { id }, header: await headers() }, body: { query } })),
    content: async (id: string, uri: string, level: 0 | 1 | 2) => unwrap(client.POST('/project-spaces/{id}/openviking/content', { params: { path: { id }, header: await headers() }, body: { uri, level } })),
  };
}
export const openVikingApi = createOpenVikingApi();
