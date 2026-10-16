import { useEffect, useState } from 'react';
import { admissionsApi, type AdmissionRequest, type AdmissionResult } from '../../api/admissions';
import { localProjectsApi } from '../../api/s1';
import { useReadApi } from '../../hooks/useReadApi';
import { useProjectSpace } from '../../hooks/useAttention';
import { usePersistentCommand } from '../../hooks/usePersistentCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { Field } from '../../components/DetailPanel';
import { SelectField, TextField } from './FormControls';
import { SourceRefs } from './SourceFields';
import type { Material, Opportunity } from './model';
import styles from './AttentionPage.module.css';

export function OpportunityAdmission({ item, materials, disabled, onChanged, onBlockedChange }: {
  item: Opportunity; materials: Material[]; disabled: boolean; onChanged: () => void; onBlockedChange: (blocked: boolean) => void;
}) {
  const projects = useReadApi(signal => localProjectsApi.listProjects({ signal }));
  const [projectID, setProjectID] = useState('');
  const project = projects.data?.items.find(entry => entry.id === projectID);
  const space = useProjectSpace(project?.space_id ?? '', !!project);
  const [prepared, setPrepared] = useState<{ body: AdmissionRequest; permissionRevision?: number } | null>(null);
  const [confirmed, setConfirmed] = useState(false);
  const [recordInput, setRecordInput] = useState('');
  const [recordID, setRecordID] = useState('');
  const [reason, setReason] = useState('');
  const record = useReadApi(signal => admissionsApi.read(recordID, signal), { key: recordID, enabled: !!recordID });
  const command = usePersistentCommand<AdmissionResult>(`opportunity:${item.id}:admission`, () => ({ status: 'succeeded', message: '已保存人工准入和提案草稿' }));
  const revoke = usePersistentCommand<AdmissionResult>(`opportunity:${item.id}:admission-revocation`, () => ({ status: 'succeeded', message: '已撤销资料准入' }));
  const writing = command.pending || command.uncertain || revoke.pending || revoke.uncertain;
  useEffect(() => { onBlockedChange(writing); }, [writing, onBlockedChange]);
  const projectReady = projectID === 'unassigned' || !!project && !projects.loading && !projects.stale && !!space.data && !space.loading && !space.stale && item.evidence_refs.every(ref => space.data!.space.material_refs.some(allowed => allowed.material_id === ref.material_id && allowed.revision === ref.revision));
  const changed = !!prepared && (prepared.body.expected_version !== item.version || prepared.body.expected_revision !== item.revision || (prepared.body.project_id ?? '') !== (projectID === 'unassigned' ? '' : projectID) || prepared.permissionRevision !== project?.settings.revision);
  const blocked = disabled || writing;
  const canPrepare = !blocked && item.state === 'ready_for_review' && !!item.version && projectReady;
  const receipt = command.result?.admission;
  return <section className={styles.record} aria-label="资料人工准入">
    <h4>资料人工准入</h4>
    <p className={styles.note}>人工准入保存固定候选及资料版本，并生成提案草稿。任务批准、执行权限和启动操作仍须分别确认。</p>
    <button type="button" className="ac-button secondary compact" disabled={projects.loading || writing} onClick={() => { projects.retry(); if (project) space.retry(); }}>刷新准入项目权限</button>
    <QueryState state={projects}>{list => <SelectField label="准入后的项目归属" value={projectID} disabled={blocked} onChange={value => { setProjectID(value); setConfirmed(false); }} options={[
      { value: '', label: '请明确选择项目归属…' }, { value: 'unassigned', label: '暂不分配项目，仅生成未分配草稿' }, ...list.items.map(entry => ({ value: entry.id, label: entry.name })),
    ]} />}</QueryState>
    {project && <><Field label="准入项目">{project.name} · {project.id}</Field><Field label="当前项目权限">r{project.settings.revision} · 目录读取 {project.settings.allow_directory ? '已允许' : '未允许'} · 子路径 {project.settings.allowed_subdirs.join('；') || '无'} · 动作 {project.settings.allowed_actions.join('；') || '无'} · 工具 {project.settings.allowed_tools.join('；') || '无'} · 模型客户端 {project.settings.external_model_cli || '未许可'} · Agent 控制 {project.settings.allow_agent_control ? '已允许' : '未允许'}</Field>
      <QueryState state={space}>{data => <Field label="项目固定资料引用">{item.evidence_refs.every(ref => data.space.material_refs.some(allowed => allowed.material_id === ref.material_id && allowed.revision === ref.revision)) ? '本候选全部固定版本已在项目空间主动引用' : '项目空间缺少本候选的固定资料版本，请先主动引用'}</Field>}</QueryState></>}
    {item.state !== 'ready_for_review' && <p role="note">当前候选状态为 {item.state}，仅待审核候选可提交人工准入。</p>}
    <Field label="本次准入的固定候选">{item.title} · {item.id} · r{item.revision} · v{item.version ?? '未知'}</Field>
    <Field label="本次准入的资料">{item.evidence_refs.map(ref => <p key={`${ref.material_id}:${ref.revision}`}>{materials.find(material => material.id === ref.material_id)?.title || ref.locator} · {ref.material_id} · revision {ref.revision}</p>)}<SourceRefs refs={item.evidence_refs} /></Field>
    <button type="button" className="ac-button secondary" disabled={!canPrepare} onClick={() => {
      setPrepared({ body: { schema_version: 1, request_id: crypto.randomUUID(), expected_version: item.version!, expected_revision: item.revision, ...(project ? { project_id: project.id } : {}) }, permissionRevision: project?.settings.revision }); setConfirmed(false);
    }}>准备人工准入请求</button>
    {prepared && <form className={styles.form} aria-label="确认人工准入" onSubmit={event => {
      event.preventDefault(); if (!canPrepare || changed || !confirmed) return;
      const body = prepared.body;
      void command.run(async key => {
        const response = await admissionsApi.admit(item.id, body, key);
        const admission = response.admission;
        if (admission.opportunity_id !== item.id || admission.opportunity_revision !== body.expected_revision || admission.opportunity_version !== body.expected_version || (admission.project_id ?? '') !== (body.project_id ?? '') || admission.authority !== 'human' || admission.status !== 'active' || !response.draft?.id || JSON.stringify(admission.inputs) !== JSON.stringify(item.evidence_refs.map(ref => ({ kind: 'material', id: ref.material_id, revision: ref.revision })))) throw new Error('准入结果与本次固定输入不一致，请核查原请求');
        return response;
      }, response => { setRecordInput(response.admission.id); setRecordID(response.admission.id); setConfirmed(false); onChanged(); }, body.request_id);
    }}>
      <Field label="准入请求路径">POST /api/v1/opportunities/{item.id}/admissions</Field>
      <pre aria-label="准入请求预览">{JSON.stringify(prepared.body, null, 2)}</pre>
      <p className={styles.note}>Idempotency-Key 与请求中的 request_id 相同。</p>
      {changed && <p role="alert">候选版本、项目选择或当前权限已有变化，原请求保留；请重新准备并审阅。</p>}
      <label><input type="checkbox" checked={confirmed} disabled={!canPrepare || changed} onChange={event => setConfirmed(event.target.checked)} />我已核对本次固定候选、资料版本和项目归属，确认人工准入</label>
      <button type="submit" className="ac-button" disabled={!canPrepare || changed || !confirmed}>确认资料准入并生成草稿</button>
    </form>}
    <CommandState {...command} />
    {command.uncertain && <p role="alert">准入请求 {command.operationID} 的结果未知，关闭或刷新后保持锁定；读取候选或准入记录不会解除原请求锁。</p>}
    {receipt && <><Field label="准入记录标识">{receipt.id}</Field><Field label="准入事件">{receipt.admitted_event_id}</Field>{command.result?.draft && <Field label="生成的提案草稿">{command.result.draft.id} · v{command.result.draft.version} · <a href="/workspace">前往共同工作区审阅</a></Field>}</>}
    <form className={styles.form} aria-label="查询准入记录" onSubmit={event => { event.preventDefault(); if (!recordInput.trim()) return; if (recordID === recordInput.trim()) record.retry(); else setRecordID(recordInput.trim()); }}>
      <TextField label="准入记录 ID" value={recordInput} onChange={setRecordInput} required disabled={revoke.pending} />
      <button type="submit" className="ac-button secondary compact" disabled={!recordInput.trim() || record.loading || revoke.pending}>读取准入记录</button>
    </form>
    {recordID && <QueryState state={record}>{data => <>
      <AdmissionFields result={data} />
      {data.admission.opportunity_id !== item.id ? <p role="alert">此准入属于其他候选，当前页面不提供撤销。</p> : <form className={styles.form} aria-label="撤销资料准入" onSubmit={event => {
        event.preventDefault(); if (blocked || record.loading || record.stale || data.admission.status !== 'active' || !reason.trim()) return;
        const original = data.admission;
        void revoke.run(async key => {
          const response = await admissionsApi.revoke(original.id, { schema_version: 1, request_id: key, expected_version: original.version, reason: reason.trim() }, key);
          if (response.admission.id !== original.id || response.admission.status !== 'revoked' || response.admission.version !== original.version + 1) throw new Error('撤销结果与原准入记录不一致');
          return response;
        }, () => { record.retry(); onChanged(); });
      }}><TextField label="撤销准入理由" value={reason} onChange={setReason} required multiline disabled={blocked || data.admission.status !== 'active'} /><button type="submit" className="ac-button secondary" disabled={blocked || record.loading || record.stale || data.admission.status !== 'active' || !reason.trim()}>确认撤销资料准入</button><p className={styles.note}>撤销后保留原准入事件和草稿；执行批准由共同工作区独立管理。</p></form>}
    </>}</QueryState>}
    <CommandState {...revoke} />
    {revoke.uncertain && <p role="alert">撤销请求 {revoke.operationID} 的结果未知，原请求保持锁定，请只读核查。</p>}
  </section>;
}

function AdmissionFields({ result }: { result: AdmissionResult }) {
  const value = result.admission;
  return <section aria-label="已保存准入记录"><Field label="已保存准入 ID">{value.id}</Field><Field label="已保存候选版本">{value.opportunity_id} · r{value.opportunity_revision} · v{value.opportunity_version}</Field><Field label="已保存准入状态">{value.status} · v{value.version}</Field><Field label="已保存项目归属">{value.project_id || '未分配项目'}</Field><Field label="准入权限来源">{value.authority} · {value.actor_id} · {value.owner_namespace}</Field><Field label="已保存固定资料">{value.inputs.map(ref => <p key={`${ref.kind}:${ref.id}:${ref.revision}`}>{ref.kind} · {ref.id} · revision {ref.revision}</p>)}</Field><Field label="已保存准入事件">{value.admitted_event_id}</Field>{value.revoked_event_id && <Field label="已保存撤销事件">{value.revoked_event_id} · {value.reason}</Field>}</section>;
}
