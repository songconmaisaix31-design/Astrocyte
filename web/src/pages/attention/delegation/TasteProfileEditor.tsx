import { useState, useCallback } from 'react';
import { useReadApi } from '../../../hooks/useReadApi';
import { judgmentApi, type TasteProfile, type TasteRubric, type TasteExample, type TasteDelegation, type QuestionType } from './judgmentApi';
import { TextField, SelectField } from '../FormControls';

function emptyRubric(id: string): TasteRubric {
  return { id, question_type: 'choice', criteria: '', options: [{ id: 'opt-1', description: '' }], levels: [] };
}

function emptyExample(id: string, rubricId: string): TasteExample {
  return { id, rubric_id: rubricId, material_id: '', revision: 1, content_digest: '', expected: '' };
}

function emptyDelegation(ruleId: string): TasteDelegation {
  return { rule_id: ruleId, revision: 1, workflow_id: '' };
}

function RubricEditor({ rubric, onChange, onRemove }: { rubric: TasteRubric; onChange: (r: TasteRubric) => void; onRemove: () => void }) {
  const updateOption = (index: number, field: 'id' | 'description', value: string) => {
    const options = [...rubric.options];
    options[index] = { ...options[index], [field]: value };
    onChange({ ...rubric, options });
  };
  return <fieldset className="ac-taste-rubric">
    <legend>评分规则 {rubric.id}</legend>
    <SelectField label="题型" value={rubric.question_type} onChange={(v: string) => onChange({ ...rubric, question_type: v as QuestionType })} options={[{ value: 'choice', label: '选择' }, { value: 'score', label: '评分' }, { value: 'truth', label: '确定性' }]} />
    <TextField label="评判标准" value={rubric.criteria} onChange={(v: string) => onChange({ ...rubric, criteria: v })} />
    <div role="group" aria-label="候选选项">
      <p>候选选项</p>
      {rubric.options.map((opt, i) => <div key={i} className="ac-inline-pair">
        <TextField label={`标识 ${i + 1}`} value={opt.id} onChange={(v: string) => updateOption(i, 'id', v)} />
        <TextField label="说明" value={opt.description} onChange={(v: string) => updateOption(i, 'description', v)} />
      </div>)}
      <button type="button" className="ac-button secondary compact" onClick={() => onChange({ ...rubric, options: [...rubric.options, { id: `opt-${rubric.options.length + 1}`, description: '' }] })}>添加选项</button>
    </div>
    <TextField label="等级（逗号分隔）" value={rubric.levels.join(', ')} onChange={(v: string) => onChange({ ...rubric, levels: v.split(',').map((s: string) => s.trim()).filter(Boolean) })} />
    <button type="button" className="ac-button secondary compact" onClick={onRemove}>移除此规则</button>
  </fieldset>;
}

export function TasteProfileEditor() {
  const tasteState = useReadApi(signal => judgmentApi.taste(signal));
  const [draft, setDraft] = useState<{ enabled: boolean; rubrics: TasteRubric[]; examples: TasteExample[]; delegations: TasteDelegation[] } | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [unknown, setUnknown] = useState<string | null>(null);

  const current: TasteProfile | null = tasteState.data?.profile ?? null;
  const configured = tasteState.data?.configured ?? false;
  const version = current?.version ?? 0;

  const initDraft = useCallback(() => {
    if (current) {
      setDraft({ enabled: current.enabled, rubrics: [...current.rubrics], examples: [...current.examples], delegations: [...current.delegations] });
    } else {
      setDraft({ enabled: true, rubrics: [], examples: [], delegations: [] });
    }
    setError(null);
    setUnknown(null);
  }, [current]);

  const handleSave = useCallback(async () => {
    if (!draft) return;
    setSaving(true);
    setError(null);
    setUnknown(null);
    const requestId = crypto.randomUUID();
    try {
      await judgmentApi.saveTaste({
        schema_version: 1,
        request_id: requestId,
        expected_version: version,
        enabled: draft.enabled,
        rubrics: draft.rubrics,
        examples: draft.examples,
        delegations: draft.delegations,
      });
      tasteState.retry();
      setDraft(null);
    } catch (cause) {
      const message = cause instanceof Error ? cause.message : String(cause);
      if (message.includes('unknown') || message.includes('网络') || !navigator.onLine) {
        setUnknown(`保存请求已发送但结果未知（${requestId}）。请核查当前配置后重试。`);
      } else {
        setError(message);
      }
    } finally {
      setSaving(false);
    }
  }, [draft, version, tasteState]);

  if (tasteState.loading && !tasteState.data) return <div role="status">加载 Taste 配置…</div>;
  if (tasteState.error && !tasteState.data) return <div role="alert">无法读取 Taste 配置：{tasteState.error}<button type="button" className="ac-button secondary compact" onClick={tasteState.retry}>重试</button></div>;

  const editing = draft !== null;
  return <section className="ac-taste-editor" aria-label="Taste 配置">
    <h3>判断偏好配置（Taste）</h3>
    <p>当前状态：{configured ? `已配置（版本 ${version}）` : '未配置'}{current?.enabled === false ? ' · 已停用' : ''}</p>

    {unknown && <div role="status" className="ac-notice">{unknown}</div>}
    {error && <div role="alert" className="ac-error">{error}</div>}

    {!editing && <button type="button" className="ac-button" onClick={initDraft}>{configured ? '编辑配置' : '创建配置'}</button>}

    {editing && draft && <>
      <label className="ac-toggle">
        <input type="checkbox" checked={draft.enabled} onChange={e => setDraft({ ...draft, enabled: e.target.checked })} />
        启用判断偏好
      </label>

      <div role="group" aria-label="评分规则">
        <h4>评分规则（{draft.rubrics.length}）</h4>
        {draft.rubrics.map((r, i) => <RubricEditor key={i} rubric={r} onChange={updated => { const rubrics = [...draft.rubrics]; rubrics[i] = updated; setDraft({ ...draft, rubrics }); }} onRemove={() => setDraft({ ...draft, rubrics: draft.rubrics.filter((_, j) => j !== i) })} />)}
        <button type="button" className="ac-button secondary" onClick={() => setDraft({ ...draft, rubrics: [...draft.rubrics, emptyRubric(`rubric-${draft.rubrics.length + 1}`)] })}>添加评分规则</button>
      </div>

      <div role="group" aria-label="示例">
        <h4>示例（{draft.examples.length}）</h4>
        {draft.examples.map((ex, i) => <fieldset key={i} className="ac-taste-example">
          <legend>示例 {i + 1}</legend>
          <SelectField label="关联规则" value={ex.rubric_id} onChange={(v: string) => { const examples = [...draft.examples]; examples[i] = { ...ex, rubric_id: v }; setDraft({ ...draft, examples }); }} options={draft.rubrics.map(r => ({ value: r.id, label: r.id }))} />
          <TextField label="资料 ID" value={ex.material_id} onChange={(v: string) => { const examples = [...draft.examples]; examples[i] = { ...ex, material_id: v }; setDraft({ ...draft, examples }); }} />
          <TextField label="版本" value={String(ex.revision)} onChange={(v: string) => { const examples = [...draft.examples]; examples[i] = { ...ex, revision: Number(v) || 1 }; setDraft({ ...draft, examples }); }} />
          <TextField label="预期结果" value={ex.expected} onChange={(v: string) => { const examples = [...draft.examples]; examples[i] = { ...ex, expected: v }; setDraft({ ...draft, examples }); }} />
          <button type="button" className="ac-button secondary compact" onClick={() => setDraft({ ...draft, examples: draft.examples.filter((_, j) => j !== i) })}>移除</button>
        </fieldset>)}
        <button type="button" className="ac-button secondary" onClick={() => setDraft({ ...draft, examples: [...draft.examples, emptyExample(`example-${draft.examples.length + 1}`, draft.rubrics[0]?.id ?? '')] })} disabled={!draft.rubrics.length}>添加示例</button>
      </div>

      <div role="group" aria-label="委托">
        <h4>委托规则引用（{draft.delegations.length}）</h4>
        {draft.delegations.map((d, i) => <fieldset key={i} className="ac-taste-delegation">
          <legend>委托 {i + 1}</legend>
          <TextField label="规则 ID" value={d.rule_id} onChange={(v: string) => { const delegations = [...draft.delegations]; delegations[i] = { ...d, rule_id: v }; setDraft({ ...draft, delegations }); }} />
          <TextField label="版本" value={String(d.revision)} onChange={(v: string) => { const delegations = [...draft.delegations]; delegations[i] = { ...d, revision: Number(v) || 1 }; setDraft({ ...draft, delegations }); }} />
          <TextField label="工作流 ID" value={d.workflow_id} onChange={(v: string) => { const delegations = [...draft.delegations]; delegations[i] = { ...d, workflow_id: v }; setDraft({ ...draft, delegations }); }} />
          <button type="button" className="ac-button secondary compact" onClick={() => setDraft({ ...draft, delegations: draft.delegations.filter((_, j) => j !== i) })}>移除</button>
        </fieldset>)}
        <button type="button" className="ac-button secondary" onClick={() => setDraft({ ...draft, delegations: [...draft.delegations, emptyDelegation(`rule-${draft.delegations.length + 1}`)] })}>添加委托引用</button>
      </div>

      <div className="ac-form-actions">
        <button type="button" className="ac-button" onClick={handleSave} disabled={saving}>{saving ? '保存中…' : '保存配置'}</button>
        <button type="button" className="ac-button secondary" onClick={() => { setDraft(null); setError(null); setUnknown(null); }}>取消</button>
        {tasteState.stale && <span role="status">数据可能已过期，保存前建议取消并重新读取。</span>}
      </div>
    </>}
  </section>;
}
