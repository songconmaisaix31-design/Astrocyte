import { api, ApiError, type ApiErrorBody } from './client';
import type { components } from './schema';

type S = components['schemas'];
export type Goal = S['S2GoalV1'];
export type Proposal = S['ProposalV1'];
export type ProposalContent = S['S2ProposalContentV1'];
export type ApprovalGrant = S['S2ApprovalGrantV1'];
export type GoalRequest = S['S2GoalRequestV1'];
export type ProposalRequest = S['S2ProposalRequestV1'];
export type ProposalTransitionRequest = S['SubmitProposalRequestV1'];
export type ProposalApprovalRequest = S['ApproveProposalRequestV1'];
export type ApprovalChangeRequest = S['RevokeApprovalRequestV1'];

export function createProposalsApi(client = api) {
  async function unwrap<T>(request: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('提案服务未返回数据');
    return data;
  }
  async function headers(key: string, requestID: string) {
    if (!key.trim() || key !== requestID) throw new Error('提案请求身份必须与原命令标识相同');
    const session = await unwrap(client.GET('/auth/session'));
    return { 'Idempotency-Key': key, 'X-CSRF-Token': session.csrf_token };
  }
  return {
    goals: (signal?: AbortSignal) => unwrap(client.GET('/goals', { signal })),
    goal: (id: string, revision?: number, signal?: AbortSignal) => unwrap(client.GET('/goals/{id}', { params: { path: { id }, query: revision === undefined ? {} : { revision } }, signal })),
    createGoal: async (body: GoalRequest, key: string) => unwrap(client.POST('/goals', { body, params: { header: await headers(key, body.request_id) } })),
    reviseGoal: async (id: string, body: GoalRequest, key: string) => unwrap(client.PUT('/goals/{id}', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    proposals: (signal?: AbortSignal) => unwrap(client.GET('/proposals', { signal })),
    proposal: (id: string, revision?: number, signal?: AbortSignal) => unwrap(client.GET('/proposals/{id}', { params: { path: { id }, query: revision === undefined ? {} : { revision } }, signal })),
    createProposal: async (body: ProposalRequest, key: string) => unwrap(client.POST('/proposals', { body, params: { header: await headers(key, body.request_id) } })),
    reviseProposal: async (id: string, body: ProposalRequest, key: string) => unwrap(client.PUT('/proposals/{id}', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    submitProposal: async (id: string, body: ProposalTransitionRequest, key: string) => unwrap(client.POST('/proposals/{id}/submit', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    declineProposal: async (id: string, body: ProposalTransitionRequest, key: string) => unwrap(client.POST('/proposals/{id}/decline', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    approveProposal: async (id: string, body: ProposalApprovalRequest, key: string) => unwrap(client.POST('/proposals/{id}/approvals', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    approval: (id: string, signal?: AbortSignal) => unwrap(client.GET('/approvals/{id}', { params: { path: { id } }, signal })),
    revokeApproval: async (id: string, body: ApprovalChangeRequest, key: string) => unwrap(client.POST('/approvals/{id}/revoke', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    completeApproval: async (id: string, body: ApprovalChangeRequest, key: string) => unwrap(client.POST('/approvals/{id}/complete', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
    expireApproval: async (id: string, body: ApprovalChangeRequest, key: string) => unwrap(client.POST('/approvals/{id}/expire', { body, params: { path: { id }, header: await headers(key, body.request_id) } })),
  };
}

export const proposalsApi = createProposalsApi();
