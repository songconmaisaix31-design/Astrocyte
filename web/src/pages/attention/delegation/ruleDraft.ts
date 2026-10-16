import type { components } from '../../../api/client';
import type { JudgmentQuestionDraft } from './JudgmentQuestionFields';
import type { DelegationReviewDraft } from './DelegationReviewFields';
import type { DelegationRule, EvidenceReference } from './judgmentApi';
import { questionDraftIssues } from './questionDraftIssues';
import { evidenceKey } from './useProjectEvidence';

export function questionToDraft(question?: DelegationRule['question']): JudgmentQuestionDraft {
  return { id: question?.question_id ?? '', type: question?.question_type ?? '', instructions: question?.instructions ?? '', rule: question ? question.rule || 'none' : '', options: question?.options?.map(option => ({ id: option.id, description: option.description ?? '', evidenceKeys: option.required_evidence?.map(evidenceKey) ?? [] })) ?? [], levels: question?.levels ?? [] };
}

export function reviewToDraft(rule?: DelegationRule): DelegationReviewDraft {
  return { admissionChoice: rule?.admission_choice ?? '', admissionValueMin: rule?.admission_value_min == null ? '' : String(rule.admission_value_min), admissionValueMax: rule?.admission_value_max == null ? '' : String(rule.admission_value_max), externalPermission: rule?.allow_opportunity_external == null ? '' : rule.allow_opportunity_external ? 'allow' : 'deny', fixedInputs: rule?.review_conditions?.includes('fixed_inputs') ?? false, humanAdopt: rule?.review_conditions?.includes('human_adopt') ?? false };
}

export function draftToQuestion(draft: JudgmentQuestionDraft, evidence: EvidenceReference[]): components['schemas']['JudgmentQuestionV1'] {
  if (!draft.type || !draft.rule) throw new Error('请明确选择题型和确定性规则使用方式');
  const issues = questionDraftIssues(draft);
  if (issues.length) throw new Error(issues.join(' '));
  const rules = { choice: 'single_supported_option', score: 'evidence_coverage', truth: 'all_evidence_present' };
  if (draft.rule !== 'none' && draft.rule !== rules[draft.type]) throw new Error('确定性规则与当前题型不匹配，请重新核查');
  return {
    question_id: draft.id.trim(), question_type: draft.type, instructions: draft.instructions, rule: draft.rule === 'none' ? '' : draft.rule,
    options: draft.type === 'choice' ? draft.options.map(option => ({ id: option.id.trim(), description: option.description, required_evidence: option.evidenceKeys.map(key => {
      const reference = evidence.find(value => evidenceKey(value) === key);
      if (!reference) throw new Error('原选项依据当前未被纳入固定资料，请核查版本和权限');
      return reference;
    }) })) : null,
    levels: draft.type === 'score' ? draft.levels : null,
  };
}

export function draftToReview(draft: DelegationReviewDraft, question: components['schemas']['JudgmentQuestionV1']): Pick<DelegationRule, 'admission_choice' | 'admission_value_min' | 'admission_value_max' | 'allow_opportunity_external' | 'review_conditions'> {
  if (!draft.fixedInputs) throw new Error('请确认准入前核对固定资料、候选版本和当前权限');
  const common = { allow_opportunity_external: draft.externalPermission === '' ? null : draft.externalPermission === 'allow', review_conditions: ['fixed_inputs', ...(draft.humanAdopt ? ['human_adopt'] : [])] };
  if (question.question_type === 'choice') {
    const choice = draft.admissionChoice.trim();
    if (!choice || !question.options?.some(option => option.id === choice)) throw new Error('请选择当前题目中的明确准入选项');
    return { ...common, admission_choice: choice, admission_value_min: null, admission_value_max: null };
  }
  const minimum = Number(draft.admissionValueMin), maximum = Number(draft.admissionValueMax);
  const limit = question.question_type === 'truth' ? 1 : (question.levels?.length ?? 0) - 1;
  if (!draft.admissionValueMin.trim() || !draft.admissionValueMax.trim() || !Number.isFinite(minimum) || !Number.isFinite(maximum) || minimum < 0 || maximum > limit || minimum > maximum) throw new Error('准入数值范围需要完整填写、处于当前题型范围内，且下限不得高于上限');
  return { ...common, admission_choice: null, admission_value_min: minimum, admission_value_max: maximum };
}
