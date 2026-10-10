import { useState } from 'react';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { TextField } from './FormControls';
import {
  paperContentStateLabel, paperSearchKey, pickedPapers, canSelectPaper, isAllEligiblePicked,
  togglePaper, clearPaperSelection, type PaperSearchHit, type PaperSearchResult,
} from './paperSearch';
import { searchPapers, importPaperBatch } from './paperSearchClient';
import { fixturePaperSearchResults } from './paperSearchFixtures';
import styles from './AttentionPage.module.css';

export function PaperSearchPanel({ fixture, onImported }: { fixture: boolean; onImported: () => void }) {
  const [query, setQuery] = useState('');
  const [submitted, setSubmitted] = useState('');
  const [selected, setSelected] = useState<string[]>([]);
  const [reason, setReason] = useState('');
  const command = useCommand(fixture);
  const state = useReadApi(
    signal => searchPapers(submitted, signal),
    { enabled: !fixture && !!submitted.trim(), key: submitted.trim() },
  );
  const result: PaperSearchResult | null = fixture ? fixturePaperSearchResults : state.data;
  const hits = result?.items ?? [];
  const picked = pickedPapers(hits, selected);
  const hasSearched = fixture || !!submitted.trim();
  const hasQuery = query.trim().length > 0;

  const runSearch = () => {
    const next = query.trim();
    if (!next) return;
    setSelected([]);
    setSubmitted(next);
  };

  const submitBatch = () => {
    void command.run(
      () => importPaperBatch(picked, reason.trim() || null),
      () => { setSelected(clearPaperSelection()); onImported(); },
      `已提交 ${picked.length} 条论文入库作业；提交成功不等于入库成功，请在处理队列查看，未选论文不会自动入库`,
    );
  };

  return <section aria-label="论文检索" className={styles.record}>
    <p className={styles.note}>按标题、作者、arXiv/DOI 检索公开论文。检索只返回公开元数据；原文可得性（全文/付费墙/受限）在勾选入库后由作业确定，失败可在处理队列查看。不会自动全量入库。</p>
    <div className={styles.form}>
      <TextField label="检索论文" value={query} onChange={setQuery} hint="输入关键词、arXiv 编号或 DOI；结果仅展示公开元数据。" />
      <button className="ac-button" type="button" disabled={fixture ? false : !hasQuery} onClick={runSearch}>{fixture ? '查看示例检索结果' : '检索论文'}</button>
    </div>
    {fixture && <p role="status" className={styles.note}>示例模式展示固定样本；真实检索与批量入库均未执行。</p>}
    {!fixture && !hasSearched && <p className={styles.note}>尚未检索。提交关键词后显示结果；未连接或服务未提供检索能力时，这里会明确提示，不会用样本冒充。</p>}
    {!fixture && hasSearched && state.loading && !state.data && <p role="status">检索中…</p>}
    {!fixture && hasSearched && state.error && !state.data && <p role="status">检索未完成：{state.error}</p>}
    {hasSearched && (state.data || fixture) && result && (
      <>
        {result.warnings && result.warnings.length > 0 && <details><summary>检索说明</summary>{result.warnings.map((warning, index) => <p key={index}>{warning}</p>)}</details>}
        {!hits.length && <p role="status">本次检索没有返回可展示的论文；不视为“无结果”。可更换关键词或稍后重试。</p>}
        {hits.length > 0 && <ul className={styles.timeline}>{hits.map(hit => {
          const selectable = canSelectPaper(hit);
          const checked = selected.includes(paperSearchKey(hit));
          return <li key={paperSearchKey(hit)}>
            <h4>{hit.title}</h4>
            <p>{hit.authors.join('、') || '作者未提供'}{hit.year ? ` · ${hit.year}` : ''}{hit.venue ? ` · ${hit.venue}` : ''}{hit.provider ? ` · 索引 ${hit.provider}` : ''}</p>
            <p className={styles.note}>{hit.arxiv_id ? `arXiv ${hit.arxiv_id}` : ''}{hit.doi ? `${hit.arxiv_id ? ' · ' : ''}DOI ${hit.doi}` : ''}</p>
            {hit.abstract && <p>{hit.abstract}</p>}
            {hit.locator && <a href={hit.locator} target="_blank" rel="noreferrer">查看原始来源</a>}
            {hit.pdf_urls.length > 0 && <p className={styles.note}>公共 PDF · {hit.pdf_urls.join(' · ')}（元数据不当作正文）</p>}
            <p className={styles.note}>原文可得状态 · {paperContentStateLabel(hit.content_state)}（入库时确定）</p>
            <label><input type="checkbox" style={{ width: 'auto' }} checked={checked} disabled={!selectable} onChange={() => setSelected(togglePaper(selected, hit))} /> 勾选入库</label>
          </li>;
        })}</ul>}
      </>
    )}
    <div className={styles.form}>
      <TextField label="所选论文收藏理由（可选）" value={reason} onChange={setReason} multiline disabled={fixture || command.pending} />
      {picked.length > 0 && <p className={styles.note}>已勾选 {picked.length} 条论文{isAllEligiblePicked(hits, selected) ? '（当前全部检索结果）' : ''}；批量入库由你明确提交，逐条生成作业并保留各自幂等键，可随时取消或重试，不会自动全量入库。</p>}
      <div className={styles.actions}>
        <button className="ac-button" type="button" disabled={fixture || command.pending || !picked.length} onClick={submitBatch}>批量入库 {picked.length} 条</button>
        <button className="ac-button secondary compact" type="button" disabled={!selected.length} onClick={() => setSelected(clearPaperSelection())}>取消全部勾选</button>
        {!fixture && <button className="ac-button secondary compact" type="button" disabled={state.loading || !submitted.trim()} onClick={state.retry}>重试检索</button>}
      </div>
    </div>
    <CommandState {...command} />
  </section>;
}

/** Keep the hit type exported for callers that need the local shape. */
export type { PaperSearchHit };
