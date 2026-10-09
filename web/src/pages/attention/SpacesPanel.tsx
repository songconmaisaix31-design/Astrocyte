import { useEffect, useState } from 'react';
import { attentionApi } from '../../api/client';
import { useProjectSpaces, useProjectSpace, useMaterialDetail } from '../../hooks/useAttention';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SectionCard } from '../../components/SectionCard';
import { TextField, SelectField } from './FormControls';
import { SourceRefs } from './SourceFields';
import type { Material } from './model';
import styles from './AttentionPage.module.css';

export function SpacesPanel({ fixture, materials, refreshToken, onChanged }: { fixture: boolean; materials: Material[]; refreshToken: number; onChanged: () => void }) {
  const state = useProjectSpaces(!fixture);
  const { retry } = state;
  useEffect(() => { if (refreshToken) retry(); }, [refreshToken, retry]);
  const [selected, setSelected] = useState('');
  const [title, setTitle] = useState('');
  const command = useCommand(fixture || state.stale || state.loading || !!state.error);
  return <SectionCard title="项目顶层空间" tabs={['overview']} padded>
    <p className={styles.note}>在空间中主动 @ 文件建立固定版本引用，原资料留在原分类。纳入与移除都由你操作，排序不扩大权限，不创建 Mission。</p>
    {fixture ? <p role="note">示例模式未提供真实项目空间，引用写操作尚未启用。</p> : <>
      <QueryState state={state} empty={data => !data.items.length} emptyTitle="暂无项目顶层空间">{data => <div className={styles.form}><SelectField label="当前项目顶层空间" value={selected} onChange={setSelected} options={[{ value: '', label: '选择空间…' }, ...data.items.map(space => ({ value: space.id, label: `${space.title} · v${space.version}` }))]} /></div>}</QueryState>
      <details><summary>创建项目顶层空间</summary><form className={styles.form} onSubmit={event => { event.preventDefault(); const request = command.prepare({ expected_version: 1, title: title.trim() }); void command.run(() => attentionApi.createProjectSpace(request.body, request.key), result => { state.retry(); setSelected(result.space.id); setTitle(''); }, '已创建项目顶层空间'); }}><TextField label="空间名称" value={title} onChange={setTitle} required disabled={command.pending} /><CommandState {...command} /><button className="ac-button secondary compact" type="submit" disabled={fixture || state.stale || state.loading || !!state.error || command.pending}>创建空间</button></form></details>
      {selected && <SpaceContent key={selected} id={selected} materials={materials} refreshToken={refreshToken} onChanged={() => { state.retry(); onChanged(); }} />}
    </>}
  </SectionCard>;
}
function SpaceContent({ id, materials, refreshToken, onChanged }: { id: string; materials: Material[]; refreshToken: number; onChanged: () => void }) {
  const state = useProjectSpace(id);
  const { retry } = state;
  useEffect(() => { if (refreshToken) retry(); }, [refreshToken, retry]);
  const [materialId, setMaterialId] = useState('');
  const [revision, setRevision] = useState('');
  const source = useMaterialDetail(materialId, !!materialId);
  const command = useCommand(state.stale || state.loading || !state.data);
  return <QueryState state={state}>{data => <section className={styles.record} aria-label="空间资料引用">
    <h4>{data.space.title} · v{data.space.version}</h4>
    <SourceRefs refs={data.space.material_refs} />
    <div className={styles.actions}>{data.space.material_refs.map(ref => <button key={ref.material_id} type="button" className="ac-button secondary compact" disabled={state.stale || state.loading || command.pending} onClick={() => { const request = command.prepare({ expected_version: data.space.version, material_id: ref.material_id }); void command.run(() => attentionApi.removeMaterialReference(id, request.body, request.key), () => { state.retry(); onChanged(); }, '已移除空间引用，原资料仍保留'); }}>移除引用 · {materials.find(material => material.id === ref.material_id)?.title || ref.material_id}</button>)}</div>
    <form className={styles.form} onSubmit={event => { event.preventDefault(); if (!revision) return; const request = command.prepare({ expected_version: data.space.version, material_id: materialId, revision: Number(revision) }); void command.run(() => attentionApi.referenceMaterial(id, request.body, request.key), () => { state.retry(); onChanged(); }, '已 @ 引用到项目顶层空间；原分类保持不变'); }}>
      <fieldset disabled={state.stale || state.loading || command.pending}><legend>主动 @ 文件</legend>
        <SelectField label="@ 资料" value={materialId} onChange={value => { setMaterialId(value); setRevision(''); }} options={[{ value: '', label: '选择资料…' }, ...materials.filter(material => material.lifecycle !== 'withdrawn').map(material => ({ value: material.id, label: material.title || material.source_locator }))]} />
        {materialId && <QueryState state={source}>{detail => <SelectField label="@ 固定内容版本" value={revision} onChange={setRevision} disabled={source.loading || source.stale} options={[{ value: '', label: '选择版本…' }, ...detail.revisions.map(entry => ({ value: String(entry.revision), label: `v${entry.revision} · ${entry.source_locator}` }))]} />}</QueryState>}
      </fieldset><CommandState {...command} /><button className="ac-button" type="submit" disabled={state.stale || state.loading || command.pending || source.loading || source.stale || !revision || !source.data}>@ 引用文件</button>
    </form>
    <p className={styles.note}>Agent 可读范围尚未配置；不自动纳入间接引用或项目全目录。</p>
  </section>}</QueryState>;
}
