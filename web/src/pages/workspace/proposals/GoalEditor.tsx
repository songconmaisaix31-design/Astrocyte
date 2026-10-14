import { useState } from 'react';
import { proposalsApi, type Goal } from '../../../api/proposals';
import { localProjectsApi } from '../../../api/s1';
import type { components } from '../../../api/schema';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { commandError } from '../../../utils/commandError';
import { SelectField, TextField } from '../../attention/FormControls';
import { lines } from './proposalDraft';
import styles from './ProposalWorkspace.module.css';

export function GoalEditor({ onSaved }: { onSaved: (goal: Goal) => void }) {
  const [projectId, setProjectId] = useState('');
  const [title, setTitle] = useState('');
  const [criteria, setCriteria] = useState('');
  const [priority, setPriority] = useState('');
  const [validation, setValidation] = useState('');
  const projects = useReadApi(signal => localProjectsApi.listProjects({ signal }), { key: 'goal-editor-projects' });
  const command = usePersistentCommand<components['schemas']['S2GoalResultV1']>('goal:create', () => ({ status: 'succeeded', message: '目标已保存' }), { failureMessage: cause => commandError(cause).message });
  const disabled = command.pending || command.uncertain || projects.loading || projects.stale || !projects.data;
  return <form className={styles.formFields} onSubmit={event => {
    event.preventDefault();
    if (disabled) return;
    if (!Number.isSafeInteger(Number(priority)) || !priority.trim() || !lines(criteria).length) { setValidation('请填写整数优先级和至少一项成功条件'); return; }
    setValidation('');
    void command.run(key => proposalsApi.createGoal({ schema_version: 1, request_id: key, expected_version: 0, goal: { project_id: projectId, title: title.trim(), success_criteria: lines(criteria), priority: Number(priority), status: 'active' } }, key), (result, outcome) => { if (outcome.status === 'succeeded') onSaved(result.goal); });
  }}>
    <QueryState state={projects}>{data => <SelectField label="目标所属项目" value={projectId} onChange={setProjectId} disabled={disabled} options={[{ value: '', label: '请选择项目' }, ...data.items.map(project => ({ value: project.id, label: project.name }))]} />}</QueryState>
    <fieldset disabled={disabled}><legend>目标与成功条件</legend><TextField label="目标名称" value={title} onChange={setTitle} required /><TextField label="目标成功条件" value={criteria} onChange={setCriteria} required multiline hint="每行一项可检查的成功条件。" /><TextField label="目标优先级" value={priority} onChange={setPriority} required type="number" numberRange={{ min: -Number.MAX_SAFE_INTEGER, max: Number.MAX_SAFE_INTEGER, step: '1' }} /></fieldset>
    {validation && <p role="alert">{validation}</p>}<CommandState {...command} />
    {command.uncertain && <p role="alert">目标保存请求 {command.operationID} 的结果未知，禁止重复创建；请重新读取目标列表核对。</p>}
    <button className="ac-button" type="submit" disabled={disabled || !projectId}>保存目标</button>
  </form>;
}
