import type { JudgmentPolicyDraft } from './JudgmentPolicyFields';
import type { JudgmentPolicy } from './judgmentApi';

export function policyToDraft(policy?: JudgmentPolicy): JudgmentPolicyDraft {
  return {
    backend: policy?.backend ?? '', languages: policy?.languages?.join('\n') ?? '', modelRevision: policy?.versions.model_revision ?? '', policyVersion: policy?.versions.policy_version ?? '', externalScope: policy?.external_scope ?? '',
    timeoutMS: policy ? String(policy.budget.timeout_ms) : '', maxInputBytes: policy ? String(policy.budget.max_input_bytes) : '', maxInputTokens: policy ? String(policy.budget.max_input_tokens) : '', maxOutputTokens: policy?.budget.max_output_tokens === undefined ? '' : String(policy.budget.max_output_tokens),
    minConfidence: policy?.min_confidence == null ? '' : String(policy.min_confidence), truthAcceptMin: policy?.truth_accept_min == null ? '' : String(policy.truth_accept_min), truthAcceptMax: policy?.truth_accept_max == null ? '' : String(policy.truth_accept_max),
  };
}

export function requiredSafeInteger(value: string, label: string, minimum: number): number {
  const parsed = Number(value);
  if (!value.trim() || !Number.isSafeInteger(parsed) || parsed < minimum) throw new Error(`${label}需要填写${minimum === 0 ? '非负' : '正'}安全整数`);
  return parsed;
}
