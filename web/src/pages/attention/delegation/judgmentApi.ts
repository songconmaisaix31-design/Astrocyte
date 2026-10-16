import { api, ApiError, type ApiErrorBody, type components, type paths } from '../../../api/client';

export type JudgmentResult = components['schemas']['JudgmentResultV1'];
export type JudgmentPolicy = components['schemas']['JudgmentPolicyV1'];
export type JudgmentSnapshot = components['schemas']['JudgmentSnapshotV1'];
export type EvidenceReference = components['schemas']['JudgmentEvidenceRefV1'];
export type DelegationRule = components['schemas']['DelegationRuleV1'];
export type QuestionType = components['schemas']['JudgmentQuestionV1']['question_type'];
export type PolicyCommand = paths['/local-projects/{project_id}/judgment-policies/{question_type}']['put']['requestBody']['content']['application/json'];
export type RuleCommand = paths['/local-projects/{project_id}/delegation-rules/{rule_id}']['put']['requestBody']['content']['application/json'];
export type RuleStateCommand = paths['/local-projects/{project_id}/delegation-rules/{rule_id}/state']['post']['requestBody']['content']['application/json'];
export type RuleEvaluationCommand = paths['/local-projects/{project_id}/delegation-rules/{rule_id}/judgments']['post']['requestBody']['content']['application/json'];
export type AdmissionCommand = paths['/local-projects/{project_id}/judgments/{decision_id}/admissions']['post']['requestBody']['content']['application/json'];
export type JudgmentCommand = Omit<paths['/local-projects/{project_id}/judgments']['post']['requestBody']['content']['application/json'], 'context'>;

export type TasteCommand = paths['/attention-taste-profile']['put']['requestBody']['content']['application/json'];
export type TasteProfile = components['schemas']['TasteProfileV1'];
export type TasteRubric = components['schemas']['TasteRubricV1'];
export type TasteExample = components['schemas']['TasteExampleV1'];
export type TasteDelegation = components['schemas']['TasteDelegationV1'];

async function unwrap<T>(request: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
  const { data, error, response } = await request;
  if (error) throw new ApiError(response.status, error);
  if (!data) throw new Error('判断服务未返回数据');
  return data;
}

async function headers() {
  const session = await unwrap(api.GET('/auth/session'));
  return { 'X-CSRF-Token': session.csrf_token };
}

// 复用已有公开客户端与会话；命令只执行一次，未知结果由页面保留原请求身份。
export const judgmentApi = {
  session: () => unwrap(api.GET('/auth/session')),
  taste: (signal?: AbortSignal) => unwrap(api.GET('/attention-taste-profile', { signal })),
  saveTaste: async (body: TasteCommand) => {
    const key = body.request_id;
    return unwrap(api.PUT('/attention-taste-profile', { params: { header: { ...await headers(), 'Idempotency-Key': key } }, body }));
  },
  snapshot: async (project_id: string, files: string[]) => unwrap(api.POST('/local-projects/{project_id}/judgment-snapshots', { params: { path: { project_id }, header: await headers() }, body: { files } })),
  policy: (project_id: string, question_type: QuestionType, signal?: AbortSignal) => unwrap(api.GET('/local-projects/{project_id}/judgment-policies/{question_type}', { params: { path: { project_id, question_type } }, signal })),
  savePolicy: async (project_id: string, question_type: QuestionType, body: PolicyCommand) => unwrap(api.PUT('/local-projects/{project_id}/judgment-policies/{question_type}', { params: { path: { project_id, question_type }, header: await headers() }, body })),
  rules: (project_id: string, signal?: AbortSignal) => unwrap(api.GET('/local-projects/{project_id}/delegation-rules', { params: { path: { project_id } }, signal })),
  rule: (project_id: string, rule_id: string, signal?: AbortSignal) => unwrap(api.GET('/local-projects/{project_id}/delegation-rules/{rule_id}', { params: { path: { project_id, rule_id } }, signal })),
  saveRule: async (project_id: string, rule_id: string, body: RuleCommand) => unwrap(api.PUT('/local-projects/{project_id}/delegation-rules/{rule_id}', { params: { path: { project_id, rule_id }, header: await headers() }, body })),
  setState: async (project_id: string, rule_id: string, body: RuleStateCommand) => unwrap(api.POST('/local-projects/{project_id}/delegation-rules/{rule_id}/state', { params: { path: { project_id, rule_id }, header: await headers() }, body })),
  evaluate: async (project_id: string, rule_id: string, body: RuleEvaluationCommand) => unwrap(api.POST('/local-projects/{project_id}/delegation-rules/{rule_id}/judgments', { params: { path: { project_id, rule_id }, header: await headers() }, body })),
  judge: async (project_id: string, body: JudgmentCommand) => unwrap(api.POST('/local-projects/{project_id}/judgments', { params: { path: { project_id }, header: await headers() }, body })),
  result: (project_id: string, decision_id: string, signal?: AbortSignal) => unwrap(api.GET('/local-projects/{project_id}/judgments/{decision_id}', { params: { path: { project_id, decision_id } }, signal })),
  feedback: (project_id: string, decision_id: string, signal?: AbortSignal) => unwrap(api.GET('/local-projects/{project_id}/judgments/{decision_id}/feedback', { params: { path: { project_id, decision_id } }, signal })),
  recordFeedback: async (project_id: string, decision_id: string, body: components['schemas']['JudgmentFeedbackRequestV1']) => unwrap(api.POST('/local-projects/{project_id}/judgments/{decision_id}/feedback', { params: { path: { project_id, decision_id }, header: await headers() }, body })),
  admit: async (project_id: string, decision_id: string, body: AdmissionCommand, key: string) => {
    if (!key || key !== body.request_id) throw new Error('准入请求身份必须与原命令一致');
    return unwrap(api.POST('/local-projects/{project_id}/judgments/{decision_id}/admissions', { params: { path: { project_id, decision_id }, header: { ...await headers(), 'Idempotency-Key': key } }, body }));
  },
};
