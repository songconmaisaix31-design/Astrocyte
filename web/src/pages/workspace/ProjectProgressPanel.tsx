import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { getProjectProgress, inferProjectProgress } from './progressClient';
import { hasInference, progressKindLabel, progressSourceSummary, progressStatusLabel } from './progressPresentation';
import { formatDateTime } from '../../utils/format';
import type { components } from '../../api/schema';
import styles from '../attention/AttentionPage.module.css';

type Project = components['schemas']['LocalProjectV1'];

export function ProjectProgressPanel({ project, disabled }: { project: Project; disabled: boolean }) {
  const state = useReadApi(signal => getProjectProgress(project.id, signal), { key: project.id });
  const command = useCommand(disabled);
  const consented = !!project.settings.external_model_cli;
  const progress = state.data?.progress;

  return <section aria-label="项目进度推断" className={styles.record}>
    <h4>项目进度（Agent 推断）</h4>
    <p className={styles.note}>进度只从获准的固定 TASK/STATUS 文件推断，附来源、版本与新鲜度；不是项目状态、许可或人类批准。未知时如实显示未知，不伪造百分比。</p>
    {state.loading && !state.data ? <p role="status">进度加载中…</p> : state.error && !state.data ? <p role="status">进度未获取：{state.error}</p> : progress && <>
      <p className={styles.note}>状态 · {progressStatusLabel(progress.status)}{progress.inferred ? ' · 推断自获准文件' : ' · 尚未推断'}</p>
      {progress.summary && <p>{progress.summary}</p>}
      <p className={styles.note}>依据来源 · {progressSourceSummary(progress.evidence)}</p>
      {progress.evidence && progress.evidence.length > 0 && <ul className={styles.timeline}>{progress.evidence.map((item, index) => <li key={index}><h5>{item.source_path}</h5><p className={styles.note}>{progressKindLabel(item.kind)} · 版本 {item.version || '未知'} · 新鲜度 {item.freshness || '未知'}</p>{item.excerpt && <pre>{item.excerpt}</pre>}</li>)}</ul>}
      <p className={styles.note}>{progress.inferred_at ? `推断时间 · ${formatDateTime(progress.inferred_at)}` : '推断时间未知'}{progress.processor ? ` · 处理器 ${progress.processor}` : ''}{progress.model ? ` · 模型 ${progress.model}` : ''}</p>
      {progress.warning && <p role="status">{progress.warning}</p>}
    </>}
    {!hasInference(progress ?? { project_id: project.id, status: 'unknown', inferred: false }) && !state.loading && !state.error && <p role="status">尚无进度推断；可明确发起一次推断。</p>}
    <div className={styles.actions}>
      <button className="ac-button secondary compact" type="button" disabled={state.loading} onClick={state.retry}>重载进度</button>
      <button className="ac-button" type="button" disabled={disabled || command.pending || !consented} onClick={() => { const request = command.prepare({ expected_version: 1 }); void command.run(() => inferProjectProgress(project.id, request.key), state.retry, '已发起进度推断；结果按获准文件实际返回'); }}>{consented ? '发起进度推断' : '需先许可项目模型客户端'}</button>
    </div>
    {!consented && <p className={styles.note}>此项目尚未许可模型客户端，不能发起进度推断；请在项目接入设置中许可后再试。</p>}
    <CommandState {...command} />
  </section>;
}
