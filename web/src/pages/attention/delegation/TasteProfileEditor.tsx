import { useEffect, useRef, useState } from 'react';
import { localProjectsApi } from '../../../api/s1';
import { useReadApi, formatError } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { CommandState } from '../../../components/CommandState';
import { QueryState } from '../../../components/QueryState';
import { SelectField } from '../FormControls';
import { judgmentApi, type TasteCommand } from './judgmentApi';
import { TasteExampleFields, TasteRubricFields } from './TasteProfileFields';
import { profileToDraft, tasteInput, type TasteDraft } from './tasteDraft';
import styles from './DelegationWorkspace.module.css';

export function TasteProfileEditor() {
  const taste = useReadApi(signal => judgmentApi.taste(signal));
  const materials = useReadApi(signal => judgmentApi.materials(signal));
  const references = useReadApi(async signal => {
    const projects = await localProjectsApi.listProjects({ signal });
    const lists = await Promise.all(projects.items.map(async project => ({ project, rules: await judgmentApi.rules(project.id, signal) })));
    return lists.flatMap(({ project, rules }) => rules.items.map(rule => ({ project, rule })));
  });
  const [draft, setDraft] = useState<TasteDraft | null>(null);
  const [expectedVersion, setExpectedVersion] = useState(1);
  const [validation, setValidation] = useState('');
  const [verifying, setVerifying] = useState(false);
  const verifyingRef = useRef(false);
  const verification = useRef<AbortController | null>(null);
  useEffect(() => () => verification.current?.abort(), []);
  const [reviewedVersion, setReviewedVersion] = useState<number | null>(null);
  const [needsPreparation, setNeedsPreparation] = useState(false);
  const [readLatest, setReadLatest] = useState(false);
  const latest = useReadApi(signal => judgmentApi.taste(signal), { enabled: readLatest });
  const save = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.saveTaste>>>('attention:taste', () => ({ status: 'succeeded', message: 'Taste 配置已保存' }));
  const blocked = verifying || save.pending || save.uncertain;
  const stale = taste.loading || taste.stale || !taste.data;
  const unavailable = blocked || stale || !!draft?.examples.length && (materials.loading || materials.stale || !materials.data);
  const profile = taste.data?.profile ?? null;
  const rules = references.data?.map(item => item.rule) ?? [];
  const referencesUnavailable = references.loading || references.stale || !references.data;
  const update = (action: (draft: TasteDraft) => TasteDraft) => setDraft(current => current ? action(current) : null);

  const submit = async () => {
    if (!draft || unavailable || needsPreparation || verifyingRef.current) return;
    let frozen: Omit<TasteCommand, 'request_id'>;
    try {
      if (draft.delegations.length && referencesUnavailable) throw new Error('委托规则读取不可用，请先只读刷新引用');
      frozen = { ...tasteInput(draft, rules), expected_version: expectedVersion };
    } catch (cause) { setValidation(formatError(cause)); return; }
    verifyingRef.current = true;
    const controller = new AbortController();
    verification.current = controller;
    setVerifying(true);
    setValidation('');
    try {
      // 固定输入在开始核查时冻结；读取失败或摘要变化立即停止，保留原草稿。
      for (const example of frozen.examples) {
        const detail = await judgmentApi.materialDetail(example.material_id, controller.signal);
        if (detail.material.lifecycle !== 'active') throw new Error('示例资料当前未处于 active 状态');
        const revision = detail.revisions.find(item => item.revision === example.revision);
        if (!revision) throw new Error('示例固定版本当前不可读取');
        if (revision.content_digest !== example.content_digest) throw new Error('示例固定内容摘要已变化，请核查原版本');
      }
      for (const reference of frozen.delegations) {
        const selected = references.data?.find(item => item.rule.id === reference.rule_id);
        if (!selected) throw new Error('原委托规则当前不可读取');
        const { delegation_rule: rule } = await judgmentApi.rule(selected.project.id, reference.rule_id, controller.signal);
        if (rule.revision !== reference.revision || rule.workflow_id !== reference.workflow_id) throw new Error('原委托规则版本已变化，请核查固定引用');
      }
      controller.signal.throwIfAborted();
    } catch (cause) {
      setValidation(formatError(cause));
      verifyingRef.current = false;
      verification.current = null;
      setVerifying(false);
      return;
    }
    verification.current = null;
    await save.run(key => judgmentApi.saveTaste({ ...frozen, request_id: key }), (_, outcome) => {
      if (outcome.status === 'succeeded') { taste.retry(); setDraft(null); }
    });
    // 已拒绝的命令也须明确重新准备；未知命令继续由持久锁阻止写入。
    setNeedsPreparation(true);
    verifyingRef.current = false;
    setVerifying(false);
    setReviewedVersion(null);
    setReadLatest(false);
  };

  return <section className={styles.result} aria-label="Taste 配置">
    <h3>判断偏好配置（Taste）</h3>
    <QueryState state={taste}>{data => <>
      <p>当前状态：{data.configured && data.profile ? `已配置（版本 ${data.profile.version}） · ${data.profile.enabled ? '已启用' : '已停用'}` : '未配置'}</p>
      {data.profile && <details><summary>查看完整已保存配置</summary>
        <p>配置版本：{data.profile.version} · 配置人：{data.profile.configured_by} · 保存时间：{data.profile.created_at}</p>
        {data.profile.rubrics.map((rubric, i) => <div key={rubric.id}><h4>准则 {i + 1} · {rubric.question_type}</h4><p>{rubric.criteria}</p>{rubric.options.map(option => <p key={option.id}>{option.id}：{option.description}</p>)}{rubric.levels.map((level, index) => <p key={index}>等级 {index}：{level}</p>)}</div>)}
        {data.profile.examples.map(example => <div key={example.id}><h4>固定示例 {example.id}</h4><p>关联准则：{example.rubric_id} · 资料：{example.material_id} · r{example.revision}</p><p>内容摘要：<code>{example.content_digest}</code></p><p>预期结果：{example.expected}</p></div>)}
        {data.profile.delegations.map(reference => <p key={reference.rule_id}>委托规则：{reference.rule_id} · r{reference.revision} · 工作流：{reference.workflow_id}</p>)}
      </details>}
    </>}</QueryState>
    <button className="ac-button secondary compact" type="button" disabled={save.pending || verifying} onClick={() => { setReviewedVersion(null); taste.retry(); }}>只读刷新 Taste 配置</button>
    {save.uncertain && <p role="alert">保存请求 {save.operationID} 的结果未知，关闭、重开和刷新后继续禁止重发。当前接口没有按请求身份读取收据的能力，读取配置版本无法确认原请求结果，写入保持锁定。</p>}
    {stale && <p role="status">Taste 读取当前不可用或已过期，写入已禁用。</p>}
    {!draft && <button className="ac-button" type="button" disabled={unavailable} onClick={() => {
      setDraft(profileToDraft(profile)); setExpectedVersion(profile ? profile.version : 1); setValidation(''); setNeedsPreparation(false); setReviewedVersion(null);
    }}>{profile ? '编辑配置' : '创建配置'}</button>}
    {draft && <form className={styles.result} onSubmit={event => { event.preventDefault(); void submit(); }}>
      <p>本次保存使用的配置版本：{expectedVersion}</p>
      <fieldset className={styles.fields} disabled={unavailable}><legend>人工判断偏好</legend>
        <SelectField label="启用状态" value={draft.enabled === null ? '' : String(draft.enabled)} onChange={value => update(current => ({ ...current, enabled: value === '' ? null : value === 'true' }))} options={[{ value: '', label: '请明确选择启用或停用' }, { value: 'true', label: '启用判断偏好' }, { value: 'false', label: '停用判断偏好' }]} />
        {draft.rubrics.map((rubric, index) => <TasteRubricFields key={rubric.id} rubric={rubric} index={index} disabled={unavailable} onChange={next => update(current => ({ ...current, rubrics: current.rubrics.map(item => item.id === rubric.id ? next : item), examples: current.examples.map(example => example.rubric_id === rubric.id && rubric.question_type !== next.question_type ? { ...example, expected: '' } : example) }))} onRemove={() => update(current => ({ ...current, rubrics: current.rubrics.filter(item => item.id !== rubric.id), examples: current.examples.filter(example => example.rubric_id !== rubric.id) }))} />)}
        <button className="ac-button secondary compact" type="button" disabled={draft.rubrics.length >= 256} onClick={() => update(current => ({ ...current, rubrics: [...current.rubrics, { id: crypto.randomUUID(), question_type: '', criteria: '', options: [], levels: [] }] }))}>添加评判准则</button>
        <QueryState state={materials}>{data => <>
          {draft.examples.map((example, index) => <TasteExampleFields key={example.id} example={example} index={index} rubrics={draft.rubrics} materials={data.items} disabled={unavailable || materials.loading || materials.stale} onChange={next => update(current => ({ ...current, examples: current.examples.map(item => item.id === example.id ? next : item) }))} onRemove={() => update(current => ({ ...current, examples: current.examples.filter(item => item.id !== example.id) }))} />)}
          <button className="ac-button secondary compact" type="button" disabled={materials.loading || materials.stale || !draft.rubrics.length || draft.examples.length >= 256} onClick={() => update(current => ({ ...current, examples: [...current.examples, { id: crypto.randomUUID(), rubric_id: '', material_id: '', revision: 0, content_digest: '', expected: '' }] }))}>添加固定资料示例</button>
        </>}</QueryState>
        <button className="ac-button secondary compact" type="button" onClick={materials.retry}>只读刷新示例资料</button>
        <div role="group" aria-label="Taste 委托引用"><h4>委托规则引用</h4>
          {draft.delegations.map((reference, index) => <p key={reference.rule_id}>规则 {reference.rule_id} · r{reference.revision} · 工作流 {reference.workflow_id}<button className="ac-button secondary compact" type="button" onClick={() => update(current => ({ ...current, delegations: current.delegations.filter((_, i) => i !== index) }))}>移除委托引用</button></p>)}
          <QueryState state={references}>{items => <SelectField label="添加实际委托规则引用" value="" disabled={referencesUnavailable} onChange={value => {
            if (!value) return;
            const selected = items.find(item => item.rule.id === value);
            if (!selected) throw new Error('所选实际委托规则不可读取');
            const rule = selected.rule;
            update(current => ({ ...current, delegations: [...current.delegations, { rule_id: rule.id, revision: rule.revision, workflow_id: rule.workflow_id }] }));
          }} options={[{ value: '', label: '请选择已有实际规则版本' }, ...items.filter(item => !draft.delegations.some(reference => reference.rule_id === item.rule.id)).map(({ project, rule }) => ({ value: rule.id, label: `${project.name} · ${rule.workflow_id} · r${rule.revision} · ${rule.question.instructions}` }))]} />}</QueryState>
          <button className="ac-button secondary compact" type="button" onClick={references.retry}>只读刷新委托引用</button>
        </div>
      </fieldset>
      {validation && <p role="alert">{validation} 输入已保留。</p>}
      <CommandState {...save} />
      {save.failure?.status === 409 && <p role="alert">配置版本冲突，原草稿及保存版本已保留。请只读刷新配置，核查后明确重新准备。</p>}
      {needsPreparation && !save.uncertain && <div><p>请读取最新配置并核查后，明确使用其版本重新准备原草稿。</p>
        <button className="ac-button secondary compact" type="button" disabled={blocked || latest.loading} onClick={() => { setReviewedVersion(null); if (readLatest) latest.retry(); else setReadLatest(true); }}>读取最新配置以核查原草稿</button>
        {readLatest && <QueryState state={latest}>{data => <>
          <p>本次核查的配置版本：{data.profile ? data.profile.version : '首次配置'} · {data.profile?.enabled ? '已启用' : '未启用'}</p>
          {data.profile?.rubrics.map(rubric => <p key={rubric.id}>{rubric.criteria}</p>)}
          <button className="ac-button secondary compact" type="button" disabled={blocked || latest.loading || latest.stale} onClick={() => setReviewedVersion(data.profile ? data.profile.version : 1)}>已核查当前配置版本 {data.profile ? data.profile.version : 1}</button>
          <button className="ac-button secondary compact" type="button" disabled={blocked || latest.loading || latest.stale || reviewedVersion === null || reviewedVersion !== (data.profile ? data.profile.version : 1)} onClick={() => { if (reviewedVersion === null) return; setExpectedVersion(reviewedVersion); setNeedsPreparation(false); setReviewedVersion(null); }}>使用已核查版本重新准备</button>
        </>}</QueryState>}
      </div>}
      <div className={styles.actions}>
        <button className="ac-button" type="submit" disabled={unavailable || needsPreparation}>保存配置</button>
        <button className="ac-button secondary" type="button" disabled={blocked} onClick={() => { setDraft(null); setValidation(''); }}>取消编辑</button>
      </div>
    </form>}
  </section>;
}
