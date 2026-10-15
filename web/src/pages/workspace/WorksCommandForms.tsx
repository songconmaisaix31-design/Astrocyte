import { useState } from 'react';
import { attitudeOptions, type WorksAttitude, type WorksWork } from './worksAlbumPresentation';
import styles from './WorksAlbum.module.css';

export type WorksMetadataEdit =
  | { kind: 'attitude'; value: WorksAttitude }
  | { kind: 'group'; value: string }
  | { kind: 'handoff'; value: string }
  | { kind: 'hidden'; value: boolean };

// 命令失败时保留草稿；调用方负责传输、版本校验和服务返回结果。
export function WorksMetadataForm({ item, busy, save }: {
  item: WorksWork;
  busy: boolean;
  save: (edit: WorksMetadataEdit) => Promise<void>;
}) {
  const [attitude, setAttitude] = useState(item.attitude);
  const [group, setGroup] = useState(item.group);
  const [handoff, setHandoff] = useState(item.handoff.content);
  return <div className={styles.commandForms}>
    <form onSubmit={event => { event.preventDefault(); if (attitude) void save({ kind: 'attitude', value: attitude }); }}>
      <label>记录态度<select value={attitude} onChange={event => setAttitude(event.target.value as WorksWork['attitude'])} disabled={busy} required>
        <option value="" disabled>请选择态度</option>{attitudeOptions.map(option => <option key={option.value} value={option.value}>{option.label}</option>)}
      </select></label>
      <button type="submit" className="ac-button compact" disabled={busy || !attitude}>保存态度</button>
    </form>
    <form onSubmit={event => { event.preventDefault(); void save({ kind: 'group', value: group }); }}>
      <label>编辑分组<input value={group} maxLength={200} onChange={event => setGroup(event.target.value)} disabled={busy} /></label>
      <button type="submit" className="ac-button compact" disabled={busy}>保存分组</button>
    </form>
    <form onSubmit={event => { event.preventDefault(); void save({ kind: 'handoff', value: handoff }); }}>
      <label>编辑交接手册<textarea value={handoff} rows={6} onChange={event => setHandoff(event.target.value)} disabled={busy} /></label>
      <p className={styles.note}>写下现状和下一步，保存为人工编写。上限 64 KiB。</p>
      <button type="submit" className="ac-button compact" disabled={busy || new TextEncoder().encode(handoff).length > 65536}>保存交接手册</button>
    </form>
    <button type="button" className="ac-button secondary compact" disabled={busy} onClick={() => { void save({ kind: 'hidden', value: !item.hidden }); }}>{item.hidden ? '取消隐藏作品' : '隐藏作品'}</button>
  </div>;
}

export function WorksDeliveryForm({ busy, submit }: { busy: boolean; submit: (content: string) => Promise<void> }) {
  const [content, setContent] = useState('');
  const tooLarge = new TextEncoder().encode(content).length > 1048576;
  return <form className={styles.commandForms} onSubmit={event => { event.preventDefault(); void submit(content); }}>
    <label>AI 投递 JSON 内容<textarea value={content} rows={10} onChange={event => setContent(event.target.value)} disabled={busy} required /></label>
    <p className={styles.note}>提交 works 数组，每项含 name、platform、path，可附 kind、group、attitude、handoff。路径只保存为线索；文件核实需另行扫描已批准项目。上限 1 MiB。</p>
    {tooLarge && <p role="alert">内容超过 1 MiB，请缩小投递范围。</p>}
    <button type="submit" className="ac-button compact" disabled={busy || tooLarge || !content.trim()}>保存 JSON 投递</button>
  </form>;
}

export function WorksScanForm({ roots, busy, submit }: { roots: string[]; busy: boolean; submit: (roots: string[]) => Promise<void> }) {
  const [selected, setSelected] = useState<string[]>([]);
  return <form className={styles.commandForms} onSubmit={event => { event.preventDefault(); void submit(selected); }}>
    <p className={styles.note}>仅扫描已批准的登记项目元数据；读取目录名、文件数量与修改时间。</p>
    <fieldset disabled={busy}><legend>选择登记项目</legend>{roots.map(root => <label className={styles.rootChoice} key={root}><input type="checkbox" checked={selected.includes(root)} onChange={event => setSelected(previous => event.target.checked ? [...previous, root] : previous.filter(value => value !== root))} /><code>{root}</code></label>)}</fieldset>
    {!roots.length && <p role="status">尚无可扫描的已批准登记项目。</p>}
    <button type="submit" className="ac-button compact" disabled={busy || !selected.length || selected.length > 64}>扫描所选项目</button>
  </form>;
}
