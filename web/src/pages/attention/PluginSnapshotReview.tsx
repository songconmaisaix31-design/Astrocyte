import { useState } from 'react';
import { attentionApi } from '../../api/client';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { TextField } from './FormControls';
import { canImportSnapshot, parsePluginSnapshot, snapshotContentStateLabel, snapshotCaveat, snapshotImportOptions, type PluginSnapshotReview, type SnapshotImportOption } from './pluginSnapshot';
import styles from './AttentionPage.module.css';

export function PluginSnapshotReview({ fixture, onImported }: { fixture: boolean; onImported: () => void }) {
  const [text, setText] = useState('');
  const [review, setReview] = useState<PluginSnapshotReview | null>(null);
  const [parseError, setParseError] = useState<string | null>(null);
  const command = useCommand(fixture);
  const options = review ? snapshotImportOptions(review, text) : [];
  const importable = review ? canImportSnapshot(review, text) : false;

  const parse = () => {
    setParseError(null);
    try { setReview(parsePluginSnapshot(text)); } catch (error) { setReview(null); setParseError(error instanceof Error ? error.message : '无法解析插件快照'); }
  };

  const submitOption = (option: SnapshotImportOption) => {
    if (!review) return;
    // The idempotency key is derived by useCommand.prepare from the FULL payload
    // signature: a failed retry of the same body keeps the key, while any body
    // change (a new snapshot text at the same URL, a different PDF link) yields
    // a new key so a new revision is created instead of a 409. No custom hash.
    const request = command.prepare({
      expected_version: 1,
      adapter: option.adapter,
      source_locator: option.source_locator,
      source_key: '',
      kind: 'paper' as const,
      content_digest: '',
      collection_reason: null,
      title: review.title ?? '',
      refresh: false,
      ...(option.export_text ? { export_text: option.export_text } : {}),
    });
    const message = option.adapter === 'paper_snapshot'
      ? '已提交快照正文入库作业；正文由你复核后写入，服务不会重新抓取网页'
      : option.adapter === 'paper_pdf'
        ? '已提交公共 PDF 入库作业；由服务安全获取该公共 PDF，不会读取你的浏览器'
        : '已提交 arXiv 官方来源入库作业';
    void command.run(() => attentionApi.importMaterial(request.body, request.key), () => { setReview(null); setText(''); onImported(); }, message);
  };

  return <section aria-label="插件快照复核导入" className={styles.record}>
    <p className={styles.note}>从论文扩展复制的 JSON 快照在此复核：保留快照正文与公共 PDF 选项，只读取你粘贴的内容，不读取登录凭据、不做全站读取。{snapshotCaveat}</p>
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
        <p className={styles.note}>快照原文可得状态 · {snapshotContentStateLabel(review.content_state)}</p>
        {review.text && <p className={styles.note}>快照包含浏览器提取的正文（{review.text.length} 字符），将作为正文写入，不再重新抓取网页。</p>}
        {review.pdf_urls.length > 0 && <p className={styles.note}>快照提供 {review.pdf_urls.length} 个公共 PDF 链接，可择一按 PDF 入库。</p>}
        {review.warning && <p role="status">受限说明 · {review.warning}</p>}
        {review.truncated && <p role="status">快照正文被截断，不能当作全文；将不会按快照正文入库。</p>}
      </div>
      {options.length > 0 ? <>{options.map(option => <button key={option.label + option.source_locator} className="ac-button" type="button" disabled={fixture || command.pending} onClick={() => submitOption(option)}>{option.label}</button>)}</> : <p role="status">{review.content_state === 'paywall' || review.content_state === 'restricted' ? '此来源受限或付费，正文不可得，无绕过；快照仅用于人工复核，未当作正文。' : '此快照未携带完整正文或公共 PDF，仅用于人工复核，未当作正文。'}</p>}
      {importable && <p className={styles.note}>以上选项由你明确选择后入库；未选择不会自动写入。</p>}
    </>}
    <CommandState {...command} />
  </section>;
}
