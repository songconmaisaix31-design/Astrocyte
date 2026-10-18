import { useState } from 'react';
import { openVikingApi, type VikingHit } from '../../api/openviking';
import { attentionApi, type components } from '../../api/client';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SectionCard } from '../../components/SectionCard';
import { Icon } from '../../components/DesignIcons';
import { formatDateTime } from '../../utils/format';

type S = components['schemas'];
const states: Record<string, string> = { submitting: '正在提交', submitted: '已提交，待查询任务', pending: '等待处理', running: '正在处理', cancelling: '正在取消', completed: '已入库', failed: '处理失败', cancelled: '已取消', unknown: '结果未知，禁止重发' };

export function OpenVikingPanel() {
  const status = useReadApi(signal => openVikingApi.status(signal));
  const spaces = useReadApi(signal => attentionApi.listProjectSpaces({ signal }));
  const [spaceID, setSpaceID] = useState('');
  const space = spaces.data?.items.find(item => item.id === spaceID);
  return <section className="ac-viking-panel" aria-label="OpenViking 资料仓库">
    <div className="ac-viking-heading"><span className="ac-viking-symbol"><Icon name="library" size={25} /></span><div><span className="ac-overline">ATTENTION REPOSITORY</span><h2>OpenViking <span className="ac-viking-state">{status.error ? '核验失败' : status.loading ? '核验中' : status.data?.configured ? status.data.connected ? '已连接' : '未连接' : '未配置'}</span></h2></div><button type="button" className="ac-button secondary compact" disabled={status.loading} onClick={status.retry}>检查连接</button></div>
    <QueryState state={status}>{data => <><p>{data.message}</p>{data.configured ? <code>{data.base_url} · {data.account} / {data.user}<br />{data.root_uri}</code> : <p className="ac-viking-note">可选资料仓库。配置完成后，可由你选择固定资料版本并确认外发。</p>}</>}</QueryState>
    <QueryState state={spaces}>{data => <label className="ac-viking-project">项目空间<select aria-label="OpenViking 项目空间" value={spaceID} onChange={event => setSpaceID(event.target.value)}><option value="">选择项目空间</option>{data.items.map(item => <option key={item.id} value={item.id}>{item.title}</option>)}</select></label>}</QueryState>
    {!spaces.loading && !spaces.error && !spaces.data?.items.length && <p>暂无项目空间，可在“管理资料域、排序与空间引用”中创建并纳入固定资料。</p>}
    {space && <ProjectRepository key={`${space.id}:${space.version}:${status.data?.base_url}:${status.data?.account}:${status.data?.user}:${status.data?.root_uri}`} space={space} status={status.data ?? undefined} disabled={status.loading || status.stale || !!status.error || spaces.stale || !!spaces.error} />}
    {!space && <button type="button" className="ac-button secondary compact" disabled>选择项目后提交或检索</button>}
  </section>;
}

function ProjectRepository({ space, status, disabled }: { space: S['ProjectSpaceV1']; status: S['OpenVikingStatusV1'] | undefined; disabled: boolean }) {
  const records = useReadApi(signal => openVikingApi.list(space.id, signal));
  const command = useCommand(disabled);
  const [selected, setSelected] = useState('');
  const [confirmed, setConfirmed] = useState(false);
  const [query, setQuery] = useState('');
  const [hits, setHits] = useState<VikingHit[] | null>(null);
  const [content, setContent] = useState<string | null>(null);
  const ref = space.material_refs.find(item => `${item.material_id}@${item.revision}` === selected);
  const source = useReadApi(signal => {
    if (!ref) throw new Error('需要选择固定资料版本');
    return attentionApi.getMaterialContent(ref.material_id, ref.revision, { signal });
  }, { key: selected, enabled: !!ref });
  const sourceReady = !!ref && source.data?.material_id === ref.material_id && source.data.revision === ref.revision && !!source.data.text.trim() && !source.loading && !source.stale && !source.error;
  const destination = status?.root_uri ? `${status.root_uri}/projects/${space.id}` : '';
  const existing = records.data?.items.find(row => row.material_id === ref?.material_id && row.revision === ref.revision && row.destination === destination && row.base_url === status?.base_url && row.account === status?.account && row.user === status?.user && row.status !== 'failed');
  const ready = !disabled && !!status?.connected && !!records.data && !records.stale && !records.error;
  async function submit() {
    if (!ref || !status || !confirmed || !ready || !sourceReady || existing) return;
    const prepared = command.prepare({ expected_version: 0 as const, expected_space_version: space.version, material_id: ref.material_id, revision: ref.revision, destination, base_url: status.base_url, account: status.account, user: status.user });
    await command.run(async () => { try { return await openVikingApi.submit(space.id, prepared.body, prepared.key); } finally { records.retry(); } }, () => setConfirmed(false), '外发提交已记录；请查询真实任务状态');
  }
  return <div className="ac-viking-workspace">
    <SectionCard title="人工选择固定资料版本" padded><label>已纳入的资料<select aria-label="OpenViking 固定资料版本" value={selected} onChange={event => { setSelected(event.target.value); setConfirmed(false); }}><option value="">选择固定版本</option>{space.material_refs.map(item => <option key={`${item.material_id}@${item.revision}`} value={`${item.material_id}@${item.revision}`}>{item.locator} · r{item.revision}</option>)}</select></label>
      {ref && <div className="ac-viking-confirm"><p>来源：{ref.locator} · r{ref.revision}</p><p>目的地：{status?.base_url || '未配置'} · {status?.account} / {status?.user}<br /><code>{destination || '未配置'}</code></p><QueryState state={source}>{data => <details><summary>核对实际固定版本正文 · {data.text.length} 字符</summary><pre className="ac-viking-content">{data.text}</pre><p>来源处理：{data.provenance.processor} · {data.provenance.version} · {data.provenance.mode}</p></details>}</QueryState><label><input type="checkbox" checked={confirmed} disabled={!ready || !sourceReady || !!existing} onChange={event => setConfirmed(event.target.checked)} />我同意将此固定版本的完整正文发送到上述目的地</label></div>}
      <button type="button" className="ac-button" disabled={!ready || !sourceReady || !ref || !confirmed || !!existing || command.pending} onClick={submit}>{existing ? states[existing.status] : '确认并提交到 OpenViking'}</button>
    </SectionCard>
    <SectionCard title="真实入库记录" padded actions={<button type="button" className="ac-text-button" onClick={records.retry}>读取记录</button>}><QueryState state={records} empty={data => !data.items.length} emptyTitle="暂无 OpenViking 提交">{data => <ul className="ac-viking-records">{data.items.map(row => <li key={row.id}><strong>{row.title}</strong><span>{states[row.status]} · r{row.revision} · {row.stage}</span><code>{row.id}</code>{row.task_id && <code>task_id: {row.task_id}</code>}{row.root_uri && <code>{row.root_uri}</code>}<small>由 {row.actor_id} 确认 · {formatDateTime(row.approved_at)}</small>{row.error && <p>{row.error}</p>}<button type="button" className="ac-button secondary compact" disabled={!ready || !row.task_id || command.pending} onClick={() => command.run(() => openVikingApi.refresh(row.id), records.retry, '已读取真实任务状态')}>查询原任务</button></li>)}</ul>}</QueryState></SectionCard>
    <SectionCard title="仓库语义检索" padded><form onSubmit={event => { event.preventDefault(); setHits(null); setContent(null); command.run(() => openVikingApi.search(space.id, query), result => setHits(result.items), '已读取检索结果'); }}><label>检索问题<input aria-label="OpenViking 语义检索" value={query} disabled={!ready} onChange={event => setQuery(event.target.value)} maxLength={8000} /></label><button className="ac-button" disabled={!ready || !query.trim() || command.pending}>检索当前项目</button></form>{hits !== null && <div>{hits.length ? hits.map(hit => <article className="ac-viking-hit" key={hit.uri}><p>{hit.abstract || hit.uri}</p><code>{hit.uri}</code><small>L{hit.level} · score {hit.score}</small><div>{(hit.level < 2 ? [0, 1] as const : [2] as const).map(level => <button type="button" className="ac-button secondary compact" key={level} disabled={!ready || command.pending} onClick={() => { setContent(null); command.run(() => openVikingApi.content(space.id, hit.uri, level), result => setContent(result.content), '已读取仓库内容'); }}>L{level} {['摘要', '概览', '正文'][level]}</button>)}</div></article>) : <p role="status">当前项目没有匹配资料</p>}</div>}{content !== null && <pre className="ac-viking-content">{content}</pre>}</SectionCard>
    <CommandState {...command} />
  </div>;
}
