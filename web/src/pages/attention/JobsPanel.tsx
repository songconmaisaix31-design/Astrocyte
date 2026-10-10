import { useEffect, useState } from 'react';
import type { components } from '../../api/schema';
import { attentionApi } from '../../api/client';
import { useJobs, useDistillations } from '../../hooks/useAttention';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SectionCard } from '../../components/SectionCard';
import { StatusBadge } from '../../components/StatusBadge';
import { formatDateTime } from '../../utils/format';
import { jobPresentation } from './jobPresentation';
import styles from './AttentionPage.module.css';

type Job = components['schemas']['JobV1'];
export function JobsPanel({ fixture, refreshToken, onChanged, onMaterial }: { fixture: boolean; refreshToken: number; onChanged: () => void; onMaterial: (id: string) => void }) {
  const state = useJobs(!fixture);
  const completedMultiple = state.data?.items.filter(job => job.distillation_id && !job.material_id) ?? [];
  const results = useDistillations(!fixture && completedMultiple.length > 0);
  const retryResults = results.retry;
  useEffect(() => { retryResults(); }, [state.data, retryResults]);
  const { retry } = state;
  const active = !!state.data?.items.some(job => job.status === 'queued' || job.status === 'running');
  useEffect(() => { if (refreshToken) retry(); }, [refreshToken, retry]);
  useEffect(() => {
    if (!active || fixture) return;
    // Bound refresh window; job deadlines and retry limits are backend facts.
    let ticks = 0;
    const timer = window.setInterval(() => { if (++ticks > 150) { window.clearInterval(timer); return; } retry(); onChanged(); }, 2000);
    return () => window.clearInterval(timer);
  }, [active, fixture, retry, onChanged]);
  return <SectionCard title="导入与沉淀队列" tabs={['overview']} padded>
    <p className={styles.note}>作业由本地服务持久保存；退出网页不停止已提交作业。进行中最多自动刷新五分钟，之后可手动刷新；刷新不会记录人工关注。</p>
    {fixture ? <p role="note">示例模式不读取真实作业，重试和取消尚未启用。</p> : <><button type="button" className="ac-button secondary compact" onClick={state.retry} disabled={state.loading}>刷新队列</button><QueryState state={state} empty={data => !data.items.length} emptyTitle="暂无作业">{data => <ul className={styles.timeline}>{data.items.map(job => <JobRow key={job.job_id} job={job} onChanged={() => { state.retry(); onChanged(); }} onMaterial={onMaterial} disabled={state.stale} />)}</ul>}</QueryState></>}
    {!fixture && completedMultiple.length > 0 && <section aria-label="多输入作业沉淀结果"><h4>多输入作业来源</h4><QueryState state={results}>{data => completedMultiple.map(job => {
      const record = data.items.find(entry => entry.id === job.distillation_id);
      return <div className={styles.record} key={job.job_id}><p>作业 {job.job_id} · 记录 {job.distillation_id}</p>{record ? <><p>{record.question}</p><div className={styles.actions}>{record.input_refs.map(ref => <button key={`${ref.material_id}@${ref.revision}`} type="button" className="ac-button secondary compact" onClick={() => onMaterial(ref.material_id)}>查看沉淀来源 · {ref.material_id} · v{ref.revision}</button>)}</div></> : <p>服务暂未返回对应沉淀记录，请刷新队列核对；不猜测代表来源。</p>}</div>;
    })}</QueryState></section>}
  </SectionCard>;
}

function JobRow({ job, onChanged, onMaterial, disabled }: { job: Job; onChanged: () => void; onMaterial: (id: string) => void; disabled: boolean }) {
  const command = useCommand(disabled);
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const timer = window.setTimeout(() => setNow(Date.now()), Math.max(0, Math.min(2147483647, Date.parse(job.deadline_at) - Date.now() + 1)));
    return () => window.clearTimeout(timer);
  }, [job.deadline_at]);
  const presentation = jobPresentation(job, now);
  const { active, retryable, unknown } = presentation;
  return <li>
    <div className={styles.itemHeader}><strong>{job.kind === 'import' ? '正文导入' : job.kind === 'distillation' ? '资料整理' : job.kind} · {job.job_id}</strong><StatusBadge value={unknown || job.status === 'failed' ? 'unknown' : job.status} label={presentation.label} /></div>
    {job.kind === 'import' && job.status === 'succeeded' && !unknown && <p className={styles.note}>已完成导入；模型整理请查看资料详情中的独立记录。</p>}
    <p className={styles.note}>尝试 {job.attempts}/{job.max_attempts} · 更新 {formatDateTime(job.updated_at)} · 截止 {formatDateTime(job.deadline_at)}</p>
    {presentation.message && <p role="status">{presentation.message}</p>}
    {presentation.limit && <p className={styles.note}>{presentation.limit}</p>}
    {job.error && <details><summary>处理详情</summary><p>{job.error.message}</p><p>服务记录：{job.error.code} · {job.error.required_action} · 请求 {job.error.request_id}</p></details>}
    {job.cancel_requested && <p role="status">取消已请求，等待服务确认。</p>}
    {job.distillation_id && <p>已保存沉淀记录 · {job.distillation_id}</p>}
    <div className={styles.actions}>
      {job.material_id && <button className="ac-button secondary compact" type="button" onClick={() => onMaterial(job.material_id!)}>查看作业资料</button>}
      <button className="ac-button secondary compact" type="button" disabled={disabled || command.pending || !retryable} title={retryable ? '重试失败作业' : '仅在失败可重试、次数和截止期限内、结果已知时可重试'} onClick={() => { const request = command.prepare({ expected_version: job.version }); void command.run(() => attentionApi.retryJob(job.job_id, request.body, request.key), onChanged, '已请求重试'); }}>重试作业</button>
      <button className="ac-button secondary compact" type="button" disabled={disabled || command.pending || !active || job.cancel_requested} onClick={() => { const request = command.prepare({ expected_version: job.version }); void command.run(() => attentionApi.cancelJob(job.job_id, request.body, request.key), onChanged, '已请求取消；以作业最终状态为准'); }}>取消作业</button>
    </div>
    <CommandState {...command} />
  </li>;
}
