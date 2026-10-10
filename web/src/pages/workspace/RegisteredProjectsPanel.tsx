import { useState } from 'react';
import type { components } from '../../api/schema';
import { localProjectsApi } from '../../api/s1';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { FilterChips } from '../../components/CollectionOverview';
import { matchesQuery } from '../../utils/search';
import { formatDateTime } from '../../utils/format';
import styles from './LocalProjectsPanel.module.css';

type Candidate = components['schemas']['RegisteredProjectV1'];
const sourceLabels = { orca_registered: 'Orca 登记目录', subproject: '目录内子项目', git_worktree: 'Git 工作树', native_project_metadata: '客户端项目记录' };
const snapshotLabels = { unknown: '尚未取得目录清单', complete: '已读取登记范围', partial: '部分目录暂不可读', stale: '本次读取未完成，保留上次清单' };

export function RegisteredProjectsPanel({ query, registeredRoots, onSelect }: { query: string; registeredRoots: string[]; onSelect: (candidate: Candidate) => void }) {
  const state = useReadApi(signal => localProjectsApi.listRegisteredProjects({ signal }));
  const command = useCommand(state.loading || state.stale || !state.data);
  const [client, setClient] = useState('all');
  const [sort, setSort] = useState('source');
  return <section aria-label="Orca 登记目录发现">
    <h3>Orca 已登记目录</h3>
    <p className={styles.note}>从已登记的项目目录发现候选，由固定程序读取目录与 Git 信息。选择后仍需关联空间并登记；读取目录不授予写入、执行或模型处理许可。</p>
    <button className="ac-button secondary compact" type="button" disabled={state.loading || command.pending} onClick={state.retry}>重载已保存目录清单</button>{' '}
    <button className="ac-button secondary compact" type="button" disabled={state.loading || state.stale || command.pending || !state.data} onClick={() => { const request = command.prepare({ expected_version: 1 }); void command.run(() => localProjectsApi.refreshRegisteredProjects(request.body, request.key), state.retry, '已请求重新读取 Orca 登记范围；查看清单中的实际读取结果'); }}>重新读取登记目录</button>
    <CommandState {...command} />
    {command.pending && <p role="status" className={styles.note}>正在重新读取登记目录，请稍候。期间保留上次清单，不改变项目许可。</p>}
    {state.error && !state.data && <details><summary>登记目录连接详情</summary><p>{state.error}</p></details>}
    <QueryState state={{ ...state, error: state.error ? '暂未取得登记目录，请检查服务连接后重新加载，也可手动填写明确目录。' : null }}>{({ snapshot }) => {
      const clients = [...new Set(snapshot.projects.flatMap(item => item.activity.created_with_cli ? [item.activity.created_with_cli] : []))];
      const visible = snapshot.projects.filter(item => (client === 'all' || (client === 'unknown' ? !item.activity.created_with_cli : item.activity.created_with_cli === client)) && matchesQuery(query, [item.name, item.root, item.parent_root, item.activity.created_with_cli ?? '']));
      if (sort === 'name') visible.sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'));
      if (sort === 'activity') visible.sort((left, right) => (right.activity.last_activity_at ? Date.parse(right.activity.last_activity_at) : 0) - (left.activity.last_activity_at ? Date.parse(left.activity.last_activity_at) : 0));
      const unavailable = state.loading || state.stale || snapshot.status === 'stale' || command.pending;
      return <>
        <p role="status" className={styles.note}>{snapshotLabels[snapshot.status]} · 观察时间 {snapshot.observed_at ? formatDateTime(snapshot.observed_at) : '未知'} · 已保存目录 {snapshot.status === 'unknown' ? '未知' : snapshot.projects.length}</p>
        <FilterChips label="登记客户端" value={client} onChange={setClient} options={[{ value: 'all', label: '全部' }, ...clients.map(value => ({ value, label: value })), { value: 'unknown', label: '客户端未记录' }]} />
        <div className={styles.controls}><label>排序<select aria-label="发现目录排序" value={sort} onChange={event => setSort(event.target.value)}><option value="source">服务返回顺序</option><option value="name">项目名称</option><option value="activity" disabled={!snapshot.projects.some(item => item.activity.last_activity_at)}>最近记录活动 · 仅已知时间</option></select></label></div>
        {!visible.length && <p className={styles.note}>{snapshot.status === 'unknown' ? '清单尚未就绪，可重新读取登记目录，或手动填写明确目录。' : snapshot.projects.length ? '没有匹配目录，可调整客户端筛选或搜索。' : '登记范围内没有返回候选目录，可手动填写明确目录。'}</p>}
        <ul className={styles.cards}>{visible.map(item => {
          const registered = registeredRoots.includes(item.root);
          return <li key={item.root} className={styles.card}>
            <h4>{item.name}</h4><div className={styles.tags}><span>{sourceLabels[item.source]}</span><span>{registered ? '已登记到 Astrocyte' : '待选择登记'}</span></div>
            <code className={styles.path}>{item.root}</code>
            {item.parent_root && <p>所属目录 · <span className={styles.path}>{item.parent_root}</span></p>}
            <p>分支 · {item.git.branch ?? '未知'} · 工作区更改 · {item.git.dirty === null ? '未知' : item.git.dirty ? '有未提交更改' : '无未提交更改'}</p>
            <p>最近提交 · {item.git.last_commit_at ? formatDateTime(item.git.last_commit_at) : '未知'}</p>
            <p>登记时客户端 · {item.activity.created_with_cli ?? '未记录'} · 最近记录活动 · {item.activity.last_activity_at ? formatDateTime(item.activity.last_activity_at) : '未知'}</p>
            <button className="ac-button secondary compact" type="button" disabled={unavailable || registered} onClick={() => onSelect(item)}>{registered ? '已登记此目录' : '选择此目录登记'}</button>
            <details><summary>目录来源与读取说明</summary><p>活动来源 · {item.activity.source || '未提供'}；Git 观察 · {item.git.reason || item.git.status}</p>{item.git.head && <p className={styles.path}>提交 · {item.git.head}</p>}{item.limitations.map((limitation, index) => <p key={index}>{limitation}</p>)}</details>
          </li>;
        })}</ul>
        {snapshot.failures.length > 0 && <details><summary>{snapshot.failures.length} 个目录读取未完成</summary>{snapshot.failures.map((failure, index) => <p key={index}><span className={styles.path}>{failure.root || '登记目录清单'}</span>{failure.reason}</p>)}</details>}
      </>;
    }}</QueryState>
  </section>;
}
