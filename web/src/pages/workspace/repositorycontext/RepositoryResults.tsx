import type { RepositoryIndex, RepositoryMapEntry, RepositoryResult, RepositorySymbol } from '../../../api/repositoryContext';
import { Field, FieldRow, MutedValue } from '../../../components/DetailPanel';
import styles from './RepositoryContext.module.css';

function Values({ items }: { items: string[] | null }) {
  if (items === null) return <MutedValue>未提供</MutedValue>;
  return items.length ? <ul>{items.map((value, index) => <li key={`${index}:${value}`}>{value}</li>)}</ul> : <MutedValue>无</MutedValue>;
}

function Entries({ entries }: { entries: RepositoryMapEntry[] | undefined }) {
  if (entries === undefined) return <p>服务未返回仓库说明。</p>;
  return entries.length ? <ul className={styles.list}>{entries.map((entry, index) => <li className={styles.item} key={`${index}:${entry.path}`}>
    <Field label="文件或目录"><code>{entry.path}</code></Field><Field label="职责">{entry.responsibility}</Field><Field label="关系说明">{entry.relationships}</Field><Field label="接口说明">{entry.api}</Field><Field label="约束">{entry.constraints}</Field><Field label="说明来源">{entry.evidence_kind}</Field>
  </li>)}</ul> : <p>当前快照没有仓库说明。</p>;
}

function Span({ span }: { span: RepositorySymbol['source_span'] }) {
  return <><code>{span.path}</code>{span.start_line > 0 && span.end_line >= span.start_line ? ` · 第 ${span.start_line}–${span.end_line} 行` : ' · 未提供行号'}</>;
}

export function RepositoryIndexView({ index, entries = true }: { index: RepositoryIndex; entries?: boolean }) {
  const snapshot = index.snapshot;
  return <div className={styles.result}>
    <FieldRow><Field label="当前索引状态">{index.state}</Field><Field label="索引生成版本"><code>{index.generation}</code></Field></FieldRow>
    <Field label="实际查询后端">{index.backend} · {index.version}</Field><Field label="索引建立时间">{index.built_at}</Field>
    <details><summary>展开实际项目快照与来源</summary><div className={styles.provenance}>
      <Field label="项目 ID"><code>{snapshot.project_id}</code></Field><Field label="仓库 ID"><code>{snapshot.repo_id}</code></Field><Field label="工作树 ID"><code>{snapshot.worktree_id}</code></Field><Field label="环境 ID"><code>{snapshot.environment_id}</code></Field>
      <Field label="HEAD"><code>{snapshot.head}</code></Field><Field label="Index digest"><code>{snapshot.index_digest}</code></Field><Field label="工作目录 digest"><code>{snapshot.worktree_digest}</code></Field>
      <Field label="项目权限版本">r{snapshot.settings_revision}</Field><Field label="快照查询范围"><Values items={snapshot.roots} /></Field><Field label="存在未提交改动">{snapshot.dirty ? '是' : '否'}</Field><Field label="快照观察时间">{snapshot.observed_at}</Field>
    </div></details>
    {entries && <Entries entries={index.map_entries} />}
  </div>;
}

export function RepositoryQueryResultView({ result, mode, disabled, onSelectSymbol }: { result: RepositoryResult; mode: 'map' | 'search' | 'impact'; disabled: boolean; onSelectSymbol: (id: string) => void }) {
  return <>
    <RepositoryIndexView index={result.index} entries={false} />
    <Field label="实际覆盖范围"><Values items={result.coverage} /></Field><Field label="缺口与限制"><Values items={result.gaps} /></Field>
    {mode !== 'map' && <><h4>符号结果</h4>
    {result.symbols === null ? <p>服务未提供符号列表。</p> : !result.symbols.length ? <p>当前范围没有匹配符号。</p> : <ul className={styles.list}>{result.symbols.map(symbol => <li className={styles.item} key={symbol.id}>
      <Field label="符号名称">{symbol.name}</Field><Field label="符号类型">{symbol.kind}</Field><Field label="源码位置"><Span span={symbol.source_span} /></Field><Field label="匹配分值">{symbol.score}</Field><Field label="符号 ID"><code>{symbol.id}</code></Field>
      <button className="ac-button secondary compact" type="button" disabled={disabled} onClick={() => onSelectSymbol(symbol.id)}>选择此符号查询影响</button>
    </li>)}</ul>}</>}
    <h4>关系证据</h4>
    {result.relations === null ? <p>服务未提供关系列表。</p> : !result.relations.length ? <p>当前范围没有返回关系证据。</p> : <ul className={styles.list}>{result.relations.map((relation, index) => <li className={styles.item} key={`${index}:${relation.source}:${relation.target}`}>
      <Field label="关系"><code>{relation.source}</code> → <code>{relation.target}</code></Field><Field label="关系类型">{relation.kind}</Field><Field label="证据类型">{relation.evidence_kind}</Field><Field label="证据位置"><Span span={relation.source_span} /></Field><Field label="提取器版本">{relation.extractor_version}</Field>
    </li>)}</ul>}
    <h4>仓库说明</h4><Entries entries={result.map_entries} />
  </>;
}
