import { useState } from 'react';
import { attentionApi } from '../../api/client';
import type { components } from '../../api/schema';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { TextField } from './FormControls';
import { paperContentStateLabel } from './paperSearch';
import { canImportSnapshot, importAdapterFor, parsePluginSnapshot, snapshotCaveat, type PluginSnapshotReview } from './pluginSnapshot';
import styles from './AttentionPage.module.css';

const toImportAdapter = (adapter: 'arxiv' | 'paper_url'): components['schemas']['ImportMaterialRequestV1']['adapter'] =>
  adapter as components['schemas']['ImportMaterialRequestV1']['adapter'];

export function PluginSnapshotReview({ fixture, onImported }: { fixture: boolean; onImported: () => void }) {
  const [text, setText] = useState('');
  const [review, setReview] = useState<PluginSnapshotReview | null>(null);
  const [parseError, setParseError] = useState<string | null>(null);
  const command = useCommand(fixture);
  const adapter = review ? importAdapterFor(review) : null;
  const importable = review ? canImportSnapshot(review) : false;

  const parse = () => {
    setParseError(null);
    try { setReview(parsePluginSnapshot(text)); } catch (error) { setReview(null); setParseError(error instanceof Error ? error.message : '无法解析插件快照'); }
  };

  const submit = () => {
    if (!review || !adapter || !importable) return;
    const request = command.prepare({ expected_version: 1, adapter: toImportAdapter(adapter), source_locator: review.source_url, source_key: '', content_digest: '', kind: 'paper' as const, collection_reason: null, title: review.title ?? '', refresh: false });
    void command.run(() => attentionApi.importMaterial(request.body, request.key), () => { setReview(null); setText(''); onImported(); }, '已提交复核来源的正文导入；正文由服务按来源获取，快照摘要未当作正文');
  };

  return <section aria-label="插件快照复核导入" className={styles.record}>
    <p className={styles.note}>从论文扩展复制的 JSON 快照在此复核：只读取元数据与来源链接，不读取登录凭据、不做全站读取。{snapshotCaveat}</p>
    {fixture && <p className={styles.note}>示例模式：写操作尚未启用，请退出示例模式连接真实 API。</p>}
    <form className={styles.form} onSubmit={event => { event.preventDefault(); parse(); }}>
      <TextField label="粘贴插件快照 JSON" value={text} onChange={value => { setText(value); setReview(null); setParseError(null); }} multiline hint="粘贴浏览器论文扩展复制的完整快照；仅用于人工复核，不自动批准。" />
      <button className="ac-button secondary compact" type="submit" disabled={command.pending || !text.trim()}>复核快照</button>
    </form>
    {parseError && <p role="alert">{parseError}</p>}
    {review && <>
      <div className={styles.record}>
        <h4>{review.title || '标题未提供'}</h4>
        <p>{review.authors.join('、') || '作者未提供'}</p>
        <p className={styles.note}>{review.arxiv_id ? `arXiv ${review.arxiv_id}` : ''}{review.doi ? `${review.arxiv_id ? ' · ' : ''}DOI ${review.doi}` : ''}{review.host_family ? ` · 来源家族 ${review.host_family}` : ''}</p>
        <p className={styles.note}>来源链接 · {review.source_url}</p>
        {review.abstract && <><p>{review.abstract}</p><p className={styles.note}>以上是快照摘要，不是正文。</p></>}
        <p className={styles.note}>快照原文可得状态 · {paperContentStateLabel(review.content_state)}</p>
        {review.warning && <p role="status">受限说明 · {review.warning}</p>}
        {review.truncated && <p role="status">快照被截断，不能据此判断全文；导入时由服务重新获取。</p>}
      </div>
      {adapter && importable ? <button className="ac-button" type="button" disabled={fixture || command.pending} onClick={submit}>按来源导入正文（{adapter === 'arxiv' ? 'arXiv' : 'paper_url'}）</button> : <p role="status">{review.content_state === 'paywall' || review.content_state === 'restricted' ? '此来源受限或付费，正文不可得，无绕过；快照仅用于人工复核，未当作正文。' : '此来源的正文提取尚未接入，快照仅用于人工复核，未当作正文。'}</p>}
    </>}
    <CommandState {...command} />
  </section>;
}
