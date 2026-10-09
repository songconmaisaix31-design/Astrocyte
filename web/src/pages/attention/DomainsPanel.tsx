import { useState } from 'react';
import type { components } from '../../api/schema';
import { attentionApi } from '../../api/client';
import { useCommand } from '../../hooks/useCommand';
import { useDomains } from '../../hooks/useAttention';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SectionCard } from '../../components/SectionCard';
import { TextField } from './FormControls';
import type { Material } from './model';
import styles from './AttentionPage.module.css';

type Domain = components['schemas']['MaterialDomainV1'];
export function DomainsPanel({ state, fixture }: { state: ReturnType<typeof useDomains>; fixture: boolean }) {
  return <SectionCard title="我的域" tabs={['overview', 'saved']} padded>
    <p className={styles.note}>由你主动分类，例如日常学习与当前研究。分类和算法排序不授予 Agent 读取权限。</p>
    {fixture ? <p role="note">示例模式未提供真实域，分类写操作尚未启用。</p> : <>
      <QueryState state={state} empty={data => !data.items.length} emptyTitle="暂无资料域">{data => <ul className={styles.timeline}>{data.items.map(domain => <li key={domain.id}><DomainEditor initial={domain} disabled={state.stale || state.loading} onSaved={state.retry} /></li>)}</ul>}</QueryState>
      <details><summary>新建资料域</summary><DomainEditor disabled={state.stale || state.loading || !!state.error} onSaved={state.retry} /></details>
    </>}
  </SectionCard>;
}
function DomainEditor({ initial, disabled, onSaved }: { initial?: Domain; disabled: boolean; onSaved: () => void }) {
  const [expectedVersion, setExpectedVersion] = useState(initial?.version ?? 1);
  const [title, setTitle] = useState(initial?.title ?? '');
  const [description, setDescription] = useState(initial?.description ?? '');
  const command = useCommand(disabled);
  return <form className={styles.form} onSubmit={event => { event.preventDefault(); const request = command.prepare({ expected_version: expectedVersion, title: title.trim(), description: description.trim() }); void command.run(() => initial ? attentionApi.reviseMaterialDomain(initial.id, request.body, request.key) : attentionApi.createMaterialDomain(request.body, request.key), result => { if (initial) setExpectedVersion(result.domain.version); onSaved(); if (!initial) { setTitle(''); setDescription(''); } }, '已保存资料域'); }}>
    {initial && initial.version !== expectedVersion && <p role="note">域已更新，当前输入保留。<button type="button" className="ac-button secondary compact" onClick={() => { setTitle(initial.title); setDescription(initial.description); setExpectedVersion(initial.version); }}>载入最新域信息</button></p>}
    <fieldset disabled={disabled || command.pending}><legend>{initial ? `资料域 · v${initial.version}` : '新建资料域'}</legend><TextField label="域名称" value={title} onChange={setTitle} required /><TextField label="域说明" value={description} onChange={setDescription} multiline /></fieldset>
    <CommandState {...command} /><button className="ac-button secondary compact" type="submit" disabled={disabled || command.pending}>{initial ? '保存域信息' : '创建资料域'}</button>
  </form>;
}

export function MaterialDomainForm({ material, state, disabled, onSaved }: { material: Material; state: ReturnType<typeof useDomains>; disabled: boolean; onSaved: () => void }) {
  const [expectedVersion, setExpectedVersion] = useState(material.version);
  const [ids, setIds] = useState(material.domain_ids ?? []);
  const command = useCommand(disabled || !material.version || material.domain_ids == null);
  return <section className={styles.record}><h4>人工分类</h4><QueryState state={state}>{data => <form className={styles.form} onSubmit={event => { event.preventDefault(); const request = command.prepare({ expected_version: expectedVersion!, domain_ids: ids }); void command.run(() => attentionApi.setMaterialDomains(material.id, request.body, request.key), result => { setExpectedVersion(result.material.version); onSaved(); }, '已保存人工分类；原文与内容版本保持可回查'); }}>
    {material.version !== expectedVersion && <p role="note">资料已更新，当前分类输入保留。<button type="button" className="ac-button secondary compact" onClick={() => { setIds(material.domain_ids ?? []); setExpectedVersion(material.version); }}>载入最新分类</button></p>}
    <fieldset disabled={disabled || state.stale || state.loading || command.pending || material.domain_ids == null}><legend>此资料所属域</legend>
      {data.items.map(domain => <label key={domain.id}><span><input type="checkbox" style={{ width: 'auto' }} checked={ids.includes(domain.id)} onChange={event => setIds(event.target.checked ? [...ids, domain.id] : ids.filter(id => id !== domain.id))} /> {domain.title}</span></label>)}
      {!data.items.length && <p>尚无资料域，请在资料页新建。</p>}
      {ids.filter(id => !data.items.some(domain => domain.id === id)).map(id => <p key={id}>未加载的域 {id}（保留已有分类）</p>)}
      {material.domain_ids == null && <p>服务未提供分类状态，保存禁用，请重新连接服务。</p>}
    </fieldset><CommandState {...command} /><button className="ac-button secondary compact" type="submit" disabled={disabled || state.stale || state.loading || command.pending || material.domain_ids == null}>保存人工分类</button>
  </form>}</QueryState></section>;
}
