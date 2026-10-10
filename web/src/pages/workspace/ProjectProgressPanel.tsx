import { useRef } from 'react';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { getProjectProgress, inferProjectProgress } from './progressClient';
import { hasProgress, progressKindLabel, progressSourceLabel, progressSourceSummary, progressStatusLabel } from './progressPresentation';
import { formatDateTime } from '../../utils/format';
import type { components } from '../../api/schema';
import styles from '../attention/AttentionPage.module.css';

type Project = components['schemas']['LocalProjectV1'];

const INFER_FILES = ['TASK.md', 'STATUS.md'];

export function ProjectProgressPanel({ project, disabled }: { project: Project; disabled: boolean }) {
  const state = useReadApi(signal => getProjectProgress(project.id, signal), { key: project.id });
  const command = useCommand(disabled);
  const consented = !!project.settings.external_model_cli;
  const progress = state.data?.progress;
  const operationId = useRef<string | null>(null);

  const infer = () => {
    if (!operationId.current) operationId.current = crypto.randomUUID();
    const request = command.prepare({ expected_version: 1, operation_id: operationId.current, files: INFER_FILES });
    void command.run(
      () => inferProjectProgress(project.id, request.key, operationId.current!, INFER_FILES),
      () => { operationId.current = null; state.retry(); },
      '已发起进度推断；结果按获准文件实际返回，未伪造百分比',
    );
  };

  return <section aria-label="项目进度推断" className={styles.record}>
    <h4>项目进度（人工记录 / Agent 推断）</h4>
    <p className={styles.note}>进度只从获准的固定 TASK/STATUS 文件推断，附来源、版本与新鲜度；不是项目状态、许可或人类批准。未知时如实显示未知，不伪造百分比。</p>
    {state.loading && !state.data ? <p role="status">进度加载中…</p> : state.error && !state.data ? <p role="status">进度未获取：{state.error}</p> : progress && <>
      <p className={styles.note}>状态 · {progressStatusLabel(progress.status)}{progress.source ? ` · ${progressSourceLabel(progress.source)}` : ''}</p>
      {progress.percent != null && <p className={styles.note}>完成度 · {progress.percent}%（人工记录）</p>}
      {progress.summary && <p>{progress.summary}</p>}
      <p className={styles.note}>依据来源 · {progressSourceSummary(progress.evidence)}</p>
      {progress.evidence.length > 0 && <ul className={styles.timeline}>{progress.evidence.map((item, index) => <li key={index}><h5>{item.source_path}</h5><p className={styles.note}>{progressKindLabel(item.kind)} · 版本 {item.version || '未知'}</p>{item.excerpt && <pre>{item.excerpt}</pre>}</li>)}</ul>}
      <p className={styles.note}>{progress.observed_at ? `观察时间 · ${formatDateTime(progress.observed_at)}` : '观察时间未知'}{progress.native_id ? ` · 原生会话 ${progress.native_id}` : ''}{progress.model ? ` · 模型 ${progress.model}` : ''}{progress.revision ? ` · 版本 r${progress.revision}` : ''}</p>
      {progress.warning && <p role="status">{progress.warning}</p>}
    </>}
    {!hasProgress(progress ?? { project_id: project.id, status: 'unknown', evidence: [], revision: 0 }) && !state.loading && !state.error && <p role="status">尚无进度记录；可明确发起一次推断。</p>}
    <div className={styles.actions}>
      <button className="ac-button secondary compact" type="button" disabled={state.loading} onClick={state.retry}>重载进度</button>
      <button className="ac-button" type="button" disabled={disabled || command.pending || !consented} onClick={infer}>{consented ? '发起进度推断' : '需先许可项目模型客户端'}</button>
    </div>
    {!consented && <p className={styles.note}>此项目尚未许可模型客户端，不能发起进度推断；请在项目接入设置中许可后再试。</p>}
    <CommandState {...command} />
  </section>;
}
