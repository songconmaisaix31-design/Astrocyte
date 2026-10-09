import { useState } from 'react';
import { attentionApi } from '../../api/client';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { TextField, SelectField } from './FormControls';
import styles from './AttentionPage.module.css';

export function ImportForm({ fixture, onImported }: { fixture: boolean; onImported: (jobId: string) => void }) {
  const [adapter, setAdapter] = useState<'arxiv' | 'summarize'>('arxiv');
  const [locator, setLocator] = useState('');
  const [reason, setReason] = useState('');
  const [title, setTitle] = useState('');
  const [exportText, setExportText] = useState('');
  const [fileError, setFileError] = useState<string | null>(null);
  const [reading, setReading] = useState(false);
  const command = useCommand(fixture);
  return <form className={styles.form} onSubmit={event => {
    event.preventDefault();
    const request = command.prepare({ expected_version: 1, adapter, source_locator: locator.trim(), source_key: '', content_digest: '', kind: adapter === 'arxiv' ? 'paper' as const : 'video' as const, collection_reason: reason.trim() || null, title: title.trim(), ...(adapter === 'summarize' ? { export_text: exportText } : {}) });
    void command.run(() => attentionApi.importMaterial(request.body, request.key), result => onImported(result.job_id), '已进入持久化导入队列；可关闭页面后继续查看');
  }}>
    <p className={styles.note}>导入保存原始来源和版本。重复内容由后端去重，资料不会自动启动任务。</p>
    {fixture && <p role="note">示例模式：写操作尚未启用，请退出示例模式连接真实 API。</p>}
    <fieldset disabled={fixture || command.pending || reading}>
      <legend>导入来源</legend>
      <SelectField label="导入方式" value={adapter} onChange={value => setAdapter(value as typeof adapter)} options={[{ value: 'arxiv', label: 'arXiv 论文' }, { value: 'summarize', label: 'summarize 既有导出' }]} />
      <TextField label={adapter === 'arxiv' ? 'arXiv 来源' : '视频原始来源'} value={locator} onChange={setLocator} required hint={adapter === 'arxiv' ? '输入 arXiv 论文 ID 或公开链接。后端获取真实内容，错误会保留在导入队列。' : '填写导出对应的真实视频 URL；已有摘要若无字幕位置，将明确显示未提供。'} />
      <TextField label="标题（可选）" value={title} onChange={setTitle} hint="arXiv 和 JSON 导出保留原始标题；纯文本导出可使用此标题。" />
      <TextField label="收藏理由（可选）" value={reason} onChange={setReason} multiline hint="留空会显示未提供，不补写为你的观点。导入后可以固定到收藏。" />
      {adapter === 'summarize' && <>
        <label>选择既有导出文件<input type="file" accept=".json,.txt,.md,application/json,text/plain,text/markdown" onChange={async event => {
          const file = event.target.files?.[0];
          if (!file) return;
          setReading(true);
          setFileError(null);
          try { setExportText(await file.text()); } catch { setFileError('无法读取导出文件，请重新选择或粘贴内容。'); } finally { setReading(false); }
        }} /></label>
        <TextField label="summarize 导出内容" value={exportText} onChange={setExportText} multiline required hint="粘贴既有 JSON 导出或纯文本；保留工具版本、原文和真实时间区间。不在页面下载视频或生成摘要。" />
      </>}
    </fieldset>
    {reading && <p role="status">正在读取本地导出…</p>}
    {fileError && <p role="alert">{fileError}</p>}
    <CommandState {...command} />
    <button className="ac-button" type="submit" disabled={fixture || command.pending || reading}>导入资料</button>
  </form>;
}
