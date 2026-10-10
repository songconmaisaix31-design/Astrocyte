import { useState } from 'react';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { TextField } from './FormControls';
import {
  paperAvailabilityLabel, paperSearchKey, pickedPapers, canSelectPaper, isAllEligiblePicked,
  firstUnknownHit, togglePaper, clearPaperSelection, type PaperSearchHit, type PaperSearchResult,
} from './paperSearch';
import { searchPapers, importPapers } from './paperSearchClient';
import { fixturePaperSearchResults } from './paperSearchFixtures';
import styles from './AttentionPage.module.css';

export function PaperSearchPanel({ fixture, onImported, onMaterial }: { fixture: boolean; onImported: () => void; onMaterial: (id: string) => void }) {
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
  const unknownHit = firstUnknownHit(hits);
  const hasSearched = fixture || !!submitted.trim();
  const hasQuery = query.trim().length > 0;

  const runSearch = () => {
    const next = query.trim();
    if (!next) return;
    setSelected([]);
    setSubmitted(next);
  };

  return <section aria-label="论文检索" className={styles.record}>
    <p className={styles.note}>按标题、作者、arXiv/DOI 检索公开论文，查看原文可得状态，勾选后批量入库，再继续沉淀。检索只返回公开元数据，正文在勾选入库时获取；不会自动全量入库。</p>
    <div className={styles.form}>
      <TextField label="检索论文" value={query} onChange={setQuery} hint="输入关键词、arXiv 编号或 DOI；结果仅展示公开元数据与原文可得状态。" />
      <button className="ac-button" type="button" disabled={fixture ? false : !hasQuery} onClick={runSearch}>{fixture ? '查看示例检索结果' : '检索论文'}</button>
    </div>
    {fixture && <p role="status" className={styles.note}>示例模式展示固定样本；真实检索与批量入库均未执行。</p>}
    {!fixture && !hasSearched && <p className={styles.note}>尚未检索。提交关键词后显示结果；未连接或服务未提供检索能力时，这里会明确提示，不会用样本冒充。</p>}
    {!fixture && hasSearched && state.loading && !state.data && <p role="status">检索中…</p>}
    {!fixture && hasSearched && state.error && !state.data && <p role="status">检索未完成：{state.error}</p>}
    {hasSearched && (state.data || fixture) && result && (
      <>
        {result.warnings.length > 0 && <details><summary>检索说明</summary>{result.warnings.map((warning, index) => <p key={index}>{warning}</p>)}</details>}
        {!hits.length && <p role="status">本次检索没有返回可展示的论文；不视为“无结果”。可更换关键词或稍后重试。</p>}
        {hits.length > 0 && <ul className={styles.timeline}>{hits.map(hit => {
          const selectable = canSelectPaper(hit);
          const checked = selected.includes(paperSearchKey(hit));
          return <li key={paperSearchKey(hit)}>
            <h4>{hit.title}</h4>
            <p>{hit.authors.join('、') || '作者未提供'}{hit.year != null ? ` · ${hit.year}` : ''}{hit.venue ? ` · ${hit.venue}` : ''}</p>
            <p className={styles.note}>{hit.arxiv_id ? `arXiv ${hit.arxiv_id}` : ''}{hit.doi ? `${hit.arxiv_id ? ' · ' : ''}DOI ${hit.doi}` : ''}</p>
            {hit.abstract && <p>{hit.abstract}</p>}
            {hit.locator && <a href={hit.locator} target="_blank" rel="noreferrer">查看原始来源</a>}
            <p className={styles.note}>原文可得状态 · {paperAvailabilityLabel(hit.availability.status)}</p>
            {hit.availability.detail && <p role="status">受限说明 · {hit.availability.detail}</p>}
            {hit.already_imported && hit.import_material_id && <button className="ac-button secondary compact" type="button" onClick={() => onMaterial(hit.import_material_id!)}>查看已入库资料</button>}
            {!selectable && !hit.already_imported && <p role="status">原文可得性未知，暂不能选择；请重试或更换来源。</p>}
            <label><input type="checkbox" style={{ width: 'auto' }} checked={checked} disabled={!selectable} onChange={() => setSelected(togglePaper(selected, hit))} /> {hit.already_imported ? '已入库' : '勾选入库'}</label>
          </li>;
        })}</ul>}
        {result.has_more && <p className={styles.note}>还有未加载的检索结果，当前列表仅覆盖本页。</p>}
      </>
    )}
    {unknownHit && !fixture && <p role="status">存在原文可得性未知的结果，请确认或重试后再勾选入库。</p>}
    <div className={styles.form}>
      <TextField label="所选论文收藏理由（可选）" value={reason} onChange={setReason} multiline disabled={fixture || command.pending} />
      {picked.length > 0 && <p className={styles.note}>已勾选 {picked.length} 条可入库论文{isAllEligiblePicked(hits, selected) ? '（当前全部可入库项）' : ''}；批量入库由你明确提交，可随时取消或重试，不会自动全量入库。</p>}
      <div className={styles.actions}>
        <button className="ac-button" type="button" disabled={fixture || command.pending || !picked.length} onClick={() => { const request = command.prepare({ items: picked.map(hit => ({ id: hit.id, source_type: hit.source_type })), collection_reason: reason.trim() || null }); void command.run(() => importPapers(request.body, request.key), () => { setSelected(clearPaperSelection()); onImported(); }, '已提交所选论文的入库作业；未选论文不会自动入库'); }}>批量入库 {picked.length} 条</button>
        <button className="ac-button secondary compact" type="button" disabled={!selected.length} onClick={() => setSelected(clearPaperSelection())}>取消全部勾选</button>
        {!fixture && <button className="ac-button secondary compact" type="button" disabled={state.loading || !submitted.trim()} onClick={state.retry}>重试检索</button>}
      </div>
    </div>
    <CommandState {...command} />
  </section>;
}

/** Keep the hit type exported for callers that need the local shape. */
export type { PaperSearchHit };
