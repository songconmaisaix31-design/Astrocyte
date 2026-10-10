import { useEffect, useState } from 'react';
import { trackingApi, localProjectsApi } from '../../api/s1';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { QueryState } from '../../components/QueryState';
import { SelectField, TextField } from './FormControls';
import { ProvenanceFields } from './SourceFields';
import { formatDateTime } from '../../utils/format';
import styles from './AttentionPage.module.css';
import { sourceItemKey as itemKey, canSelectSourceItem, canRecommendSourceItem } from './sourceSelection';

export function SourceReview({ id, onChanged, onMaterial }: { id: string; onChanged: () => void; onMaterial: (id: string) => void }) {
  const state = useReadApi(signal => trackingApi.getSource(id, { signal }), { key: id });
  const projects = useReadApi(signal => localProjectsApi.listProjects({ signal }));
  const [projectID, setProjectID] = useState('');
  const [selected, setSelected] = useState<string[]>([]);
  const [reason, setReason] = useState('');
  const [confirmSync, setConfirmSync] = useState(false);
  const command = useCommand(state.loading || state.stale || !state.data);
  const active = !!state.data && (['syncing', 'recommending', 'running', 'queued'].includes(state.data.source.status) || state.data.items.some(item => ['pending', 'running'].includes(item.recommendation?.status ?? '')));
  const reload = state.retry;
  useEffect(() => {
    if (!active) return;
    let ticks = 0;
    const timer = window.setInterval(() => { if (++ticks > 150) { window.clearInterval(timer); return; } reload(); }, 2000);
    return () => window.clearInterval(timer);
  }, [active, reload]);
  const project = projects.data?.items.find(item => item.id === projectID);
  const cli = project?.settings.external_model_cli;
  const changed = () => { state.retry(); onChanged(); };
  return <section aria-label="公开来源更新清单" className={styles.record}>
    <button type="button" className="ac-button secondary compact" disabled={state.loading} onClick={state.retry}>重载已保存清单</button>
    {active && <p className={styles.note}>处理进度最多自动读取五分钟，之后可手动重载；读取不会重发模型请求或获取正文。</p>}
    <QueryState state={state}>{data => {
      const eligible = data.items.filter(canSelectSourceItem);
      const picked = eligible.filter(item => selected.includes(itemKey(item)));
      const pendingRecommendation = data.items.filter(canRecommendSourceItem).slice(0, 100);
      const blocked = state.loading || state.stale || command.pending;
      return <>
        <h4>{data.source.title || data.source.external_id}</h4><p>{data.source.locator}</p>
        <p className={styles.note}>上次成功同步 · {data.source.last_success_at ? formatDateTime(data.source.last_success_at) : '未完成'} · 已保存标题 {data.items.length} 条{data.has_more ? '，还有未加载记录' : ''}</p>
        {data.source.last_error && <><p role="status">暂未取得新清单，请检查公开链接与平台访问限制；已有记录保留。</p><details><summary>同步详情</summary><p>{data.source.last_error.message} · {data.source.last_error.required_action}</p></details></>}
        {data.warnings.length > 0 && <details><summary>来源说明</summary>{data.warnings.map((warning, index) => <p key={index}>{warning}</p>)}</details>}
        <button className="ac-button secondary compact" type="button" disabled={blocked} onClick={() => setConfirmSync(!confirmSync)}>同步新标题与简介</button>
        {confirmSync && <div className={styles.record}><p>将重新访问此公开来源，最多获取100条标题与简介；正文和模型建议不会随同步自动执行。</p><button className="ac-button" type="button" disabled={blocked} onClick={() => { const request = command.prepare({ expected_version: data.source.version, limit: 100 }); void command.run(() => trackingApi.syncSource(id, request.body, request.key), () => { setConfirmSync(false); setSelected([]); changed(); }, '同步请求已完成；清单状态以服务保存结果为准'); }}>确认同步标题清单</button>{data.has_more && data.next_cursor && <button className="ac-button secondary compact" type="button" disabled={blocked} onClick={() => { const request = command.prepare({ expected_version: data.source.version, limit: 100, cursor: data.next_cursor! }); void command.run(() => trackingApi.syncSource(id, request.body, request.key), () => { setConfirmSync(false); changed(); }, '已获取下一页标题'); }}>获取下一页标题</button>}</div>}
        {!data.items.length && <p>{data.source.last_error || !data.source.last_success_at ? '清单尚未成功获取，不视为空更新。' : '公开来源本次没有返回可展示内容，可提供另一个公开来源。'}</p>}
        <div className={styles.form}><QueryState state={projects}>{list => <SelectField label="建议使用的项目许可" value={projectID} onChange={setProjectID} options={[{ value: '', label: '选择已批准模型处理的项目…' }, ...list.items.map(item => ({ value: item.id, label: item.name }))]} />}</QueryState>
          <p className={styles.note}>处理客户端 · {cli || '尚未设置模型许可，请在共同工作区配置'}。仅发送本页列出的标题与简介，不读取正文；本机 CLI 可能调用外部模型。</p>
          <button className="ac-button secondary compact" type="button" disabled={blocked || projects.loading || projects.stale || !cli || !pendingRecommendation.length} onClick={() => { const request = command.prepare({ expected_version: data.source.version, project_id: projectID, cli: cli!, items: pendingRecommendation.map(item => ({ external_id: item.external_id, revision: item.revision })) }); void command.run(() => trackingApi.recommendItems(id, request.body, request.key), changed, '已请求标题与简介建议；完成前不能勾选入库'); }}>获取本页 Agent 建议</button>
        </div>
        <ul className={styles.timeline}>{data.items.map(item => <li key={itemKey(item)}>
          <h4>{item.metadata.title || '标题未提供'}</h4><p>{item.metadata.description || '简介未提供'}</p><p className={styles.note}>{item.metadata.author || '作者未提供'} · 标题版本 {item.revision} · {item.stale ? '旧记录，需先同步核对' : '已保存元数据'}</p><a href={item.metadata.locator} target="_blank" rel="noreferrer">查看公开原始来源</a>
          {item.recommendation?.status === 'succeeded' && item.recommendation.error?.code !== 'delivery_unknown' ? <><p>Agent 建议 · {item.recommendation.text || '建议文本未提供'}</p><ProvenanceFields value={item.recommendation.provenance} />{item.recommendation.metadata_revision !== item.revision && <p>建议对应旧标题版本，需重新取得建议。</p>}</> : <p role="status">{item.recommendation?.status === 'unknown' || item.recommendation?.error?.code === 'delivery_unknown' ? '建议结果未知，请先核对原处理；不会自动重发。' : item.recommendation?.status === 'running' || item.recommendation?.status === 'pending' ? '建议正在处理，完成后再选择。' : item.recommendation?.status === 'failed' ? '建议暂未完成，请在处理队列核对原因；恢复配置后可明确重试。' : '尚未取得此标题版本的 Agent 建议。'}</p>}
          {item.recommendation?.error && <details><summary>建议处理详情</summary><p>{item.recommendation.error.message} · {item.recommendation.error.required_action}</p></details>}
          <label><input type="checkbox" style={{ width: 'auto' }} checked={picked.some(entry => itemKey(entry) === itemKey(item))} disabled={blocked || !canSelectSourceItem(item)} onChange={event => setSelected(event.target.checked ? [...selected, itemKey(item)] : selected.filter(key => key !== itemKey(item)))} /> {item.selected ? '已选择入库' : '选择此内容获取正文'}</label>
          {item.material_id && <button className="ac-button secondary compact" type="button" onClick={() => onMaterial(item.material_id!)}>查看已入库资料</button>}{item.import_job_id && <p className={styles.note}>正文处理已提交，请在导入队列查看进度。</p>}
        </li>)}</ul>
        <form className={styles.form} onSubmit={event => { event.preventDefault(); if (!picked.length) return; const request = command.prepare({ expected_version: data.source.version, items: picked.map(item => ({ external_id: item.external_id, revision: item.revision })), collection_reason: reason.trim() || null }); void command.run(() => trackingApi.selectItems(id, request.body, request.key), () => { setSelected([]); changed(); }, '已提交选中内容的正文作业；未选内容不会自动入库'); }}><TextField label="所选内容收藏理由（可选）" value={reason} onChange={setReason} multiline disabled={command.pending} /><button className="ac-button" type="submit" disabled={blocked || !picked.length}>获取所选 {picked.length} 条正文</button></form>
        <CommandState {...command} />
      </>;
    }}</QueryState>
  </section>;
}

