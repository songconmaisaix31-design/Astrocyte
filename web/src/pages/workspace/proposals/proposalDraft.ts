import type { components } from '../../../api/schema';
import type { ProposalDraft } from './ProposalContentFields';

type Content = components['schemas']['S2ProposalContentV1'];
type Reference = components['schemas']['FixedReferenceV1'];

export function newProposalDraft(): ProposalDraft {
  return { title: '', goalId: '', goalRevision: '', projectId: '', roots: '', writePaths: '', actions: '', tools: '', externalTargets: '', maxConcurrency: '', calls: '', tokens: '', moneyMinor: '', currency: '', deliverables: '', successCriteria: '', deadline: '', stopConditions: '', maxWorkItems: '', maxAttempts: '', noProgressSeconds: '', maxNoProgressRounds: '' };
}

export function lines(value: string) {
  return value.split('\n').map(line => line.trim()).filter(Boolean);
}

function integer(value: string, label: string, nullable: false): number;
function integer(value: string, label: string, nullable: true): number | null;
function integer(value: string, label: string, nullable: boolean) {
  if (nullable && !value.trim()) return null;
  const number = Number(value);
  if (!value.trim() || !Number.isSafeInteger(number) || number < 0) throw new Error(`${label}需要填写非负安全整数`);
  return number;
}

export function draftToContent(draft: ProposalDraft, inputs: Reference[], settingsRevision: number): Content {
  const deadline = new Date(draft.deadline);
  if (!draft.deadline || !Number.isFinite(deadline.getTime())) throw new Error('请选择有效的截止时间');
  return {
    title: draft.title.trim(), goal: { id: draft.goalId.trim(), revision: integer(draft.goalRevision, '目标版本', false) }, inputs,
    scope: { project_id: draft.projectId.trim(), settings_revision: settingsRevision, roots: lines(draft.roots), write_paths: lines(draft.writePaths), actions: lines(draft.actions), tools: lines(draft.tools), external_targets: lines(draft.externalTargets), max_concurrency: integer(draft.maxConcurrency, '最大并发数', false) },
    budget: { calls: integer(draft.calls, '调用上限', true), tokens: integer(draft.tokens, 'Token 上限', true), money_minor: integer(draft.moneyMinor, '费用上限', true), currency: draft.currency.trim() },
    stops: { deadline: deadline.toISOString(), conditions: lines(draft.stopConditions), max_work_items: integer(draft.maxWorkItems, '最大工作项数', false), max_attempts: integer(draft.maxAttempts, '最大尝试次数', false), no_progress_seconds: integer(draft.noProgressSeconds, '无进展时限', false), max_no_progress_rounds: integer(draft.maxNoProgressRounds, '连续无进展轮数上限', false) },
    deliverables: lines(draft.deliverables), success_criteria: lines(draft.successCriteria),
  };
}

export function contentToDraft(content: Content): ProposalDraft {
  const deadline = new Date(content.stops.deadline);
  const localDeadline = new Date(deadline.getTime() - deadline.getTimezoneOffset() * 60_000).toISOString().slice(0, 16);
  return {
    title: content.title, goalId: content.goal.id, goalRevision: String(content.goal.revision), projectId: content.scope.project_id,
    roots: (content.scope.roots ?? []).join('\n'), writePaths: (content.scope.write_paths ?? []).join('\n'), actions: (content.scope.actions ?? []).join('\n'), tools: (content.scope.tools ?? []).join('\n'), externalTargets: (content.scope.external_targets ?? []).join('\n'), maxConcurrency: String(content.scope.max_concurrency),
    calls: content.budget.calls === null ? '' : String(content.budget.calls), tokens: content.budget.tokens === null ? '' : String(content.budget.tokens), moneyMinor: content.budget.money_minor === null ? '' : String(content.budget.money_minor), currency: content.budget.currency,
    deliverables: (content.deliverables ?? []).join('\n'), successCriteria: (content.success_criteria ?? []).join('\n'), deadline: localDeadline,
    stopConditions: (content.stops.conditions ?? []).join('\n'), maxWorkItems: String(content.stops.max_work_items), maxAttempts: String(content.stops.max_attempts), noProgressSeconds: String(content.stops.no_progress_seconds), maxNoProgressRounds: String(content.stops.max_no_progress_rounds),
  };
}
