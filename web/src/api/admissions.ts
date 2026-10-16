import { api, ApiError, type ApiErrorBody } from './client';
import type { components } from './schema';

type S = components['schemas'];
export type AdmissionRequest = S['AdmissionRequestV1'];
export type AdmissionResult = S['AdmissionResultV1'];

export function createAdmissionsApi(client = api) {
  async function unwrap<T>(request: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('准入服务未返回数据');
    return data;
  }
  async function headers(key: string, requestID: string) {
    if (!key.trim() || key !== requestID) throw new Error('准入请求身份必须与原命令标识相同');
    const session = await unwrap(client.GET('/auth/session'));
    return { 'Idempotency-Key': key, 'X-CSRF-Token': session.csrf_token };
  }
  return {
    admit: async (id: string, body: AdmissionRequest, key: string) => unwrap(client.POST('/opportunities/{id}/admissions', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    read: (id: string, signal?: AbortSignal) => unwrap(client.GET('/admissions/{id}', { params: { path: { id } }, signal })),
    revoke: async (id: string, body: S['RevokeAdmissionRequestV1'], key: string) => unwrap(client.POST('/admissions/{id}/revocations', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
  };
}

export const admissionsApi = createAdmissionsApi();
