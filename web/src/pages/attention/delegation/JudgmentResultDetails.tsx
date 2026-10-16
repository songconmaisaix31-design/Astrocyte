import { Field } from '../../../components/DetailPanel';
import type { JudgmentResult } from './judgmentApi';
import styles from './DelegationWorkspace.module.css';

const decisions = { recommend: '建议采用', abstain: '保留判断', need_evidence: '需要补充依据' };
const questionTypes = { choice: '选项判断', score: '分级评分', truth: '真实性判断' };
const providers = { rule: '确定性规则', jev: '云端 Jev', laya: '本地 Laya' };

export function JudgmentResultDetails({ result }: { result: JudgmentResult }) {
  return <section className={styles.result} aria-label="固定判断结果">
    <Field label="判断结果">{decisions[result.decision]}</Field>
    <Field label="题目与固定版本"><code>{result.question_id}</code> · {questionTypes[result.question_type]} · r{result.result_version}</Field>
    <Field label="判断记录标识"><code>{result.decision_id}</code></Field>
    <Field label="返回选项">{result.selected ?? '未提供选项'}</Field>
    <Field label="判断数值">{result.value === null ? '未知' : result.value}</Field>
    <Field label="模型置信信息">{result.provider_confidence === null ? '未知' : result.provider_confidence}</Field>
    <Field label="判断来源">{providers[result.provider]} · {result.provider_model ?? '模型未提供'} · {result.model_revision ?? '版本未提供'}</Field>
    <Field label="输入截断">{result.truncated ? '已截断，请复核完整输入' : '未截断'}</Field>
    <Field label="缓存命中">{result.cache_hit ? '使用已保存的同输入结果' : '本次实际判断'}</Field>
    {result.distribution !== null && <details><summary>查看返回分布</summary><dl>{Object.entries(result.distribution).map(([option, value]) => <div key={option}><dt>{option}</dt><dd>{value}</dd></div>)}</dl></details>}
    <details><summary>调用用量与覆盖信息</summary>{result.usage === null ? <p>调用用量未知。</p> : <>
      <Field label="输入 Token">{result.usage.input_tokens}</Field><Field label="输出 Token">{result.usage.output_tokens}</Field>
      <Field label="状态 Token">{result.usage.state_tokens ?? '未知'}</Field><Field label="丢弃的状态 Token">{result.usage.state_tokens_dropped ?? '未知'}</Field>
      <Field label="后端截断信息">{result.usage.truncated === null ? '未知' : result.usage.truncated ? '已截断' : '未截断'}</Field>
      <Field label="被截断的题目">{result.usage.truncated_questions === null ? '未知' : result.usage.truncated_questions.length ? result.usage.truncated_questions.join('；') : '无'}</Field>
    </>}{result.options === null ? <p>选项覆盖信息未知。</p> : <ul>{Object.entries(result.options).map(([id, coverage]) => <li key={id}>{id}：总量 {coverage.total} · 不同项 {coverage.distinct} · 每项 Token {coverage.tokens_per_option}</li>)}</ul>}</details>
    <details><summary>固定来源与版本</summary>
      <Field label="目标版本标识">{result.versions.goal_version}</Field><Field label="Taste 版本标识">{result.versions.taste_version}</Field>
      <Field label="策略版本标识">{result.versions.policy_version}</Field><Field label="模型版本标识">{result.versions.model_revision}</Field><Field label="授权版本标识">{result.versions.access_epoch}</Field>
      {result.sources && <><Field label="固定目标"><code>{result.sources.goal.id}</code> · r{result.sources.goal.revision}</Field><Field label="完整 Taste 来源"><code>{result.sources.taste.id}</code> · v{result.sources.taste.version}</Field>{result.sources.ranking && <Field label="排序来源"><code>{result.sources.ranking.id}</code> · v{result.sources.ranking.version}</Field>}</>}
      <ul>{(result.evidence_refs ?? []).map(reference => <li key={`${reference.material_id}:${reference.revision}`}><code>{reference.material_id}</code> · r{reference.revision}<br /><code>{reference.digest}</code></li>)}</ul>
      <Field label="输入摘要标识"><code>{result.input_digest}</code></Field><Field label="返回原因标识">{result.reason_code}</Field><Field label="记录时间">{result.created_at}</Field>
    </details>
    {result.context && <details><summary>委托与候选版本</summary><Field label="委托规则"><code>{result.context.rule_id}</code> · r{result.context.rule_version}</Field><Field label="工作流"><code>{result.context.workflow_id}</code></Field><Field label="固定候选">{result.context.opportunity.title} · r{result.context.opportunity.revision} · v{result.context.opportunity.version}</Field><Field label="候选状态">{result.context.opportunity.state}</Field><Field label="下一步">{result.context.opportunity.next_step}</Field><Field label="缺少依据">{result.context.opportunity.missing_evidence?.join('；') || '无'}</Field></details>}
    <p>模型置信信息与判断数值分别记录。资料准入、人工采用与任务批准使用各自的确认入口。</p>
  </section>;
}
