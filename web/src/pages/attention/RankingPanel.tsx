import { useEffect, useState } from 'react';
import type { components } from '../../api/schema';
import { attentionApi } from '../../api/client';
import { useRankingProfile } from '../../hooks/useAttention';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SectionCard } from '../../components/SectionCard';
import { TextField } from './FormControls';
import { dimensionLabels } from './model';
import styles from './AttentionPage.module.css';

type Profile = components['schemas']['RankingProfileV1'];
export function RankingPanel({ fixture, refreshToken, onChanged }: { fixture: boolean; refreshToken: number; onChanged: () => void }) {
  const state = useRankingProfile(!fixture);
  const { retry } = state;
  useEffect(() => { if (refreshToken) retry(); }, [refreshToken, retry]);
  return <SectionCard title="候选四维排序" tabs={['overview', 'opportunities']} padded>
    <p className={styles.note}>未配置时保留四维与人工顺序。由你填写完整权重后才启用综合排序；目标进展首要，其余三项必须有效。排序不授权读取或执行。</p>
    {fixture ? <p role="note">示例模式无真实排序配置，修改尚未启用。</p> : <QueryState state={state}>{data => <>
      <p>{data.configured && data.profile ? `已保存配置 v${data.profile.version} · ${data.profile.enabled ? '启用' : '停用'}` : '尚未配置综合排序，权重留空'}</p>
      <RankingForm initial={data.profile} disabled={state.loading || state.stale} onSaved={() => { retry(); onChanged(); }} />
      <details><summary>排序配置版本历史</summary><ul className={styles.timeline}>{data.versions.map(profile => <li key={profile.version}>v{profile.version} · {profile.enabled ? '启用' : '停用'} · {Object.entries(profile.weights).map(([key, value]) => `${dimensionLabels[key as keyof typeof dimensionLabels]} ${value}`).join('；')}</li>)}</ul>{!data.versions.length && <p>暂无已保存配置</p>}</details>
    </>}</QueryState>}
  </SectionCard>;
}
function RankingForm({ initial, disabled, onSaved }: { initial: Profile | null; disabled: boolean; onSaved: () => void }) {
  const [expectedVersion, setExpectedVersion] = useState(initial?.version ?? 1);
  const [enabled, setEnabled] = useState(initial?.enabled ?? false);
  const [weights, setWeights] = useState<Record<keyof Profile['weights'], string>>({ goal_progress: initial ? String(initial.weights.goal_progress) : '', current_interest: initial ? String(initial.weights.current_interest) : '', project_improvement: initial ? String(initial.weights.project_improvement) : '', originality: initial ? String(initial.weights.originality) : '' });
  const command = useCommand(disabled);
  return <form className={styles.form} onSubmit={event => { event.preventDefault(); const request = command.prepare({ expected_version: expectedVersion, enabled, weights: { goal_progress: Number(weights.goal_progress), current_interest: Number(weights.current_interest), project_improvement: Number(weights.project_improvement), originality: Number(weights.originality) } }); void command.run(() => attentionApi.updateRankingProfile(request.body, request.key), result => { setExpectedVersion(result.profile?.version ?? expectedVersion); onSaved(); }, '已保存排序配置；未知维度仍保持未知'); }}>
    {initial && initial.version !== expectedVersion && <p role="note">配置已更新，当前输入保留。<button type="button" className="ac-button secondary compact" onClick={() => { setExpectedVersion(initial.version); setEnabled(initial.enabled); setWeights({ goal_progress: String(initial.weights.goal_progress), current_interest: String(initial.weights.current_interest), project_improvement: String(initial.weights.project_improvement), originality: String(initial.weights.originality) }); }}>载入最新排序配置</button></p>}
    <fieldset disabled={disabled || command.pending}><legend>人工设置排序权重</legend>
      <label><span><input type="checkbox" style={{ width: 'auto' }} checked={enabled} onChange={event => setEnabled(event.target.checked)} /> 启用综合排序</span></label>
      {(Object.keys(dimensionLabels) as (keyof Profile['weights'])[]).map(key => <TextField key={key} label={`${dimensionLabels[key]}权重`} type="number" required value={weights[key]} numberRange={{ min: 0.000001, step: 'any' }} onChange={value => setWeights({ ...weights, [key]: value })} />)}
    </fieldset><CommandState {...command} /><button className="ac-button secondary compact" type="submit" disabled={disabled || command.pending}>保存排序配置</button>
  </form>;
}
