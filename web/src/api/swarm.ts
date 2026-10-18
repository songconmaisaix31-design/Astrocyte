import { api, ApiError, type ApiErrorBody } from './client';
import type { components, operations } from './schema';

type S = components['schemas'];
export type SwarmMission = S['SwarmMissionV1'];
export type SwarmMissionDetail = S['SwarmMissionDetailV1'];
export type SwarmMissionDAG = S['SwarmMissionDAGV1'];
export type SwarmWorkItem = S['SwarmWorkItemV1'];
export type SwarmArtifact = S['SwarmArtifactV1'];
export type SwarmLease = S['SwarmLeaseV1'];
export type SwarmTimelineEvent = S['SwarmTimelineEventV1'];
export type SwarmExecution = S['SwarmExecutionV1'];
export type SwarmCreateRequest = S['SwarmCreateRequestV1'];
export type SwarmWorkItemRequest = S['SwarmWorkItemRequestV1'];
export type SwarmTransitionRequest = S['SwarmTransitionRequestV1'];
export type SwarmStopRequest = S['SwarmStopRequestV1'];
export type SwarmAcceptRequest = S['SwarmAcceptRequestV1'];
export type SwarmRejectRequest = S['SwarmRejectRequestV1'];
export type SwarmArtifactQuery = NonNullable<operations['readSwarmArtifact']['parameters']['query']>;

/** 命令完整保留版本、授权和原请求身份，读取不产生执行或采用事实。 */
export function createSwarmApi(client = api) {
  async function unwrap<T>(request: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('蜂群服务未返回数据');
    return data;
  }
  async function headers(key: string, requestID: string, signal?: AbortSignal) {
    if (!key.trim() || key !== requestID) throw new Error('蜂群请求身份必须与原命令标识相同');
    const session = await unwrap(client.GET('/auth/session', { signal }));
    return { 'Idempotency-Key': key, 'X-CSRF-Token': session.csrf_token };
  }
  return {
    missions: (signal?: AbortSignal) => unwrap(client.GET('/missions', { signal })),
    mission: (id: string, signal?: AbortSignal) => unwrap(client.GET('/missions/{id}', { params: { path: { id } }, signal })),
    dag: (id: string, signal?: AbortSignal) => unwrap(client.GET('/missions/{id}/dag', { params: { path: { id } }, signal })),
    createMission: async (body: SwarmCreateRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/missions', { body, params: { header: await headers(key, body.request_id, signal) }, signal })),
    workItems: (id: string, signal?: AbortSignal) => unwrap(client.GET('/missions/{id}/work-items', { params: { path: { id } }, signal })),
    addWorkItem: async (id: string, body: SwarmWorkItemRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/missions/{id}/work-items', { body, params: { path: { id }, header: await headers(key, body.request_id, signal) }, signal })),
    timeline: (id: string, signal?: AbortSignal) => unwrap(client.GET('/missions/{id}/timeline', { params: { path: { id } }, signal })),
    pause: async (id: string, body: SwarmStopRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/missions/{id}/pause', { body, params: { path: { id }, header: await headers(key, body.request_id, signal) }, signal })),
    cancel: async (id: string, body: SwarmStopRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/missions/{id}/cancel', { body, params: { path: { id }, header: await headers(key, body.request_id, signal) }, signal })),
    recover: async (id: string, body: SwarmTransitionRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/missions/{id}/recover', { body, params: { path: { id }, header: await headers(key, body.request_id, signal) }, signal })),
    artifact: (id: string, query: SwarmArtifactQuery, signal?: AbortSignal) => unwrap(client.GET('/artifacts/{id}', { params: { path: { id }, query }, signal })),
    artifactDiff: (id: string, query: SwarmArtifactQuery, signal?: AbortSignal) => unwrap(client.GET('/artifacts/{id}/diff', { params: { path: { id }, query }, signal })),
    acceptArtifact: async (id: string, body: SwarmAcceptRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/artifacts/{id}/acceptance', { body, params: { path: { id }, header: await headers(key, body.request_id, signal) }, signal })),
    rejectArtifact: async (id: string, body: SwarmRejectRequest, key: string, signal?: AbortSignal) => unwrap(client.POST('/artifacts/{id}/rejection', { body, params: { path: { id }, header: await headers(key, body.request_id, signal) }, signal })),
  };
}

export const swarmApi = createSwarmApi();
