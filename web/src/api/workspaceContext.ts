import { api, ApiError, type ApiErrorBody } from './client';
import type { components } from './schema';

type S = components['schemas'];
export type WorkspaceContextPacket = S['S2ContextPacketV1'];
export type WorkspaceContextPacketRequest = S['S2ContextPacketRequestV1'];
export type WorkspaceContextPacketResult = S['S2ContextPacketResultV1'];
export type WorkspaceSessionBinding = S['S2SessionBindingV1'];
export type WorkspaceSessionBindingRequest = S['S2SessionBindingRequestV1'];
export type WorkspaceSessionBindingResult = S['S2SessionBindingResultV1'];

/** 固定上下文及会话绑定沿准确安全契约读取，不补充本机私有身份。 */
export function createWorkspaceContextApi(client = api) {
  async function unwrap<T>(request: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('工作区上下文服务未返回数据');
    return data;
  }
  async function headers(key: string, requestID: string, signal?: AbortSignal) {
    if (!key.trim() || key !== requestID) throw new Error('工作区上下文请求身份必须与原命令标识相同');
    const session = await unwrap(client.GET('/auth/session', { signal }));
    return { 'Idempotency-Key': key, 'X-CSRF-Token': session.csrf_token };
  }
  return {
    packet: (id: string, signal?: AbortSignal) => unwrap(client.GET('/context-packets/{id}', { params: { path: { id } }, signal })),
    buildPacket: async (body: WorkspaceContextPacketRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/context-packets', { body, params: { header: await headers(key, body.request_id, signal) }, signal })),
    bindSession: async (body: WorkspaceSessionBindingRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/session-bindings', { body, params: { header: await headers(key, body.request_id, signal) }, signal })),
  };
}

export const workspaceContextApi = createWorkspaceContextApi();
