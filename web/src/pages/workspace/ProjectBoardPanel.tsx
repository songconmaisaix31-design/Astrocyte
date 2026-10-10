import { useEffect, useState } from 'react';
import type { components } from '../../api/schema';
import { localProjectsApi } from '../../api/s1';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SectionCard } from '../../components/SectionCard';
import { DetailPanel, Field } from '../../components/DetailPanel';
import { BrandIcon } from '../../components/BrandIcon';
import { formatDateTime } from '../../utils/format';
import { ProjectBoardView } from './ProjectBoardView';
import { ProjectHumanForm, type ProjectHumanFields } from './ProjectHumanForm';
import { projectSourceLabel, type BoardProject } from './projectBoardPresentation';
import styles from './ProjectBoard.module.css';

type Summary = components['schemas']['ProjectSummaryV1'];
function present(item: Summary): BoardProject {
  const activitySources = new Set([
    ...item.contributors.filter(source => source.activity_at === item.last_activity_at).map(source => source.source),
    ...item.observations.filter(source => source.activity.last_activity_at === item.last_activity_at && item.last_activity_at).map(source => source.activity.source),
    ...item.observations.filter(source => source.git.last_commit_at === item.last_activity_at && item.last_activity_at).map(() => 'Git 提交'),
  ]);
  return { id: item.id, name: item.name, folders: item.roots, clients: [...new Set(item.contributors.map(source => source.cli).filter(Boolean))], sources: [...new Set([...item.observations.map(source => source.source), ...item.contributors.map(source => source.source)])], activityAt: item.last_activity_at, activitySource: [...activitySources].filter(Boolean).map(projectSourceLabel).join(' · '), activityStatus: item.last_activity_at ? '有活动记录' : '未知', ...item.human };
}

export function ProjectBoardPanel({ query, refreshToken, onManage }: { query: string; refreshToken: number; onManage: (root: string, name: string) => void }) {
  const state = useReadApi(signal => localProjectsApi.listRegisteredProjects({ signal }));
  const retry = state.retry;
  useEffect(() => { if (refreshToken) void Promise.resolve().then(retry); }, [refreshToken, retry]);
  const [selectedID, setSelectedID] = useState('');
  const refresh = useCommand(state.loading || state.stale || !state.data);
  const selected = state.data?.snapshot.board?.find(item => item.id === selectedID);
  return <SectionCard title="项目总览" tabs={['overview']} padded actions={<button type="button" className="ac-button secondary compact" onClick={state.retry} disabled={state.loading}>重载项目看板</button>}>
    <div className={styles.board}>
      <p className={styles.note}>各客户端参与的项目汇在这里。先看下一步，再打开项目继续；记录与归档由你决定。</p>
      <div className={styles.actions}>
        <button type="button" className="ac-button" disabled={state.loading || state.stale || !state.data || refresh.pending} onClick={() => { const request = refresh.prepare({ expected_version: 1 }); void refresh.run(() => localProjectsApi.refreshRegisteredProjects(request.body, request.key), state.retry, '已更新项目来源，人工记录保留'); }}>同步最近改动</button>
        <button type="button" className="ac-button secondary" onClick={() => onManage('', '')}>接入项目</button>
      </div>
      <CommandState {...refresh} />
      <QueryState state={state}>{({ snapshot }) => <>
        <p role="status" className={styles.note}>观察时间 · {snapshot.observed_at ? formatDateTime(snapshot.observed_at) : '未知'} · {snapshot.status === 'complete' ? '登记范围已读取' : snapshot.status === 'partial' ? '部分来源受限' : snapshot.status === 'stale' ? '清单已过期，保留上次结果' : '尚未取得清单'}</p>
        {snapshot.board && snapshot.status !== 'unknown' ? <ProjectBoardView items={snapshot.board.map(present)} query={query} complete={snapshot.status === 'complete'} onSelect={item => setSelectedID(item.id)} /> : <p role="status" className={styles.empty}>项目汇总尚未生成，项目、文件夹与客户端计数未知。可同步最近改动；已有手动登记与项目权限仍可使用。</p>}
        {!!snapshot.sources?.length && <details><summary>客户端来源与读取限制</summary><ul className={styles.provenance}>{snapshot.sources.map((source, index) => <li key={`${source.cli}:${source.source}:${index}`}><BrandIcon name={source.cli} size={18} /><b>{source.cli || '客户端未提供'}</b><span>{source.status} · {source.source}</span><p>{source.reason || '没有提供额外说明'}</p></li>)}</ul></details>}
        {!!snapshot.failures.length && <details><summary>{snapshot.failures.length} 个目录读取未完成</summary>{snapshot.failures.map((failure, index) => <p key={index} className={styles.note}>{failure.root || '目录清单'} · {failure.reason}</p>)}</details>}
      </>}</QueryState>
    </div>
    {selected && <DetailPanel title={`项目 · ${selected.name}`} onClose={() => setSelectedID('')}><ProjectBoardDetail key={selected.id} item={selected} disabled={state.loading || state.stale || state.data?.snapshot.status === 'stale'} onSaved={state.retry} onManage={(root, name) => { setSelectedID(''); onManage(root, name); }} /></DetailPanel>}
  </SectionCard>;
}

function ProjectBoardDetail({ item, disabled, onSaved, onManage }: { item: Summary; disabled: boolean; onSaved: () => void; onManage: (root: string, name: string) => void }) {
  const command = useCommand(disabled);
  function save(fields: ProjectHumanFields) {
    const request = command.prepare({ expected_version: 1, notes: fields.notes, review: fields.review, group: fields.group, intent: fields.intent, archived: fields.archived, revision: fields.revision });
    void command.run(() => localProjectsApi.setRegisteredProjectMetadata(item.id, request.body, request.key), onSaved, '已保存项目记录，可在看板筛选归档并恢复');
  }
  return <>
    <Field label="项目完成度">未知；没有已批准的进度依据，不从活动时间推断。</Field>
    <Field label="最近记录活动">{item.last_activity_at ? formatDateTime(item.last_activity_at) : '未知'}</Field>
    <ProjectHumanForm human={item.human} disabled={disabled || command.pending} onSave={save} />
    <CommandState {...command} />
    <section className={styles.detailSection}><h4>在项目中继续</h4><p className={styles.note}>选择明确目录，查看资料权限与原生 Agent 操作。</p>{item.roots.map(root => <button key={root} type="button" className={`ac-button secondary ${styles.rootButton}`} onClick={() => onManage(root, item.name)}><span>设置与原生操作 →</span><code>{root}</code></button>)}</section>
    <section className={styles.detailSection}><h4>已记录的参与客户端</h4>{item.contributors.length ? <ul className={styles.provenance}>{item.contributors.map((source, index) => <li key={`${source.cli}:${source.session_id}:${index}`}><BrandIcon name={source.cli} size={18} /><b>{source.cli}</b><code>{source.root}</code><p>会话创建 · {source.created_at ? formatDateTime(source.created_at) : '未知'} · 最近活动 · {source.activity_at ? formatDateTime(source.activity_at) : '未知'}</p><details><summary>来源详情</summary>{source.source}</details></li>)}</ul> : <p className={styles.note}>当前只知道登记时选择的客户端，尚无对应项目工作记录。</p>}</section>
    <details><summary>目录、提交与来源依据</summary>{item.observations.map(observation => <div key={observation.root} className={styles.detailSection}><code className={styles.path}>{observation.root}</code><p className={styles.note}>来源 · {observation.source} · 登记时客户端 · {observation.activity.created_with_cli || '未知'}</p><p className={styles.note}>分支 · {observation.git.branch || '未知'} · 最近提交 · {observation.git.last_commit_at ? formatDateTime(observation.git.last_commit_at) : '未知'} · 工作区改动 · {observation.git.dirty === null ? '未知' : observation.git.dirty ? '有未提交更改' : '无未提交更改'}</p></div>)}{item.limitations.map((limitation, index) => <p key={index} className={styles.note}>{limitation}</p>)}</details>
  </>;
}
