import { attentionApi } from '../../api/client';
import type { useImportDraft } from './useImportDraft';
import { CommandState } from '../../components/CommandState';
import { TextField, SelectField } from './FormControls';
import styles from './AttentionPage.module.css';

export function ImportForm({ fixture, draft, onImported }: { fixture: boolean; draft: ReturnType<typeof useImportDraft>; onImported: (jobId: string) => void }) {
  const { adapter, setAdapter, locator, setLocator, reason, setReason, title, setTitle, exportText, setExportText, fileError, setFileError, reading, setReading, command } = draft;
  return <form className={styles.form} onSubmit={event => {
    event.preventDefault();
    const request = command.prepare({ expected_version: 1, adapter, source_locator: locator.trim(), source_key: '', content_digest: '', kind: adapter === 'arxiv' ? 'paper' as const : 'video' as const, collection_reason: reason.trim() || null, title: title.trim(), ...(adapter === 'summarize' ? { export_text: exportText } : {}) });
    void command.run(() => attentionApi.importMaterial(request.body, request.key), result => onImported(result.job_id), '已提交正文导入；在持久化队列查看实际进度，可关闭表单继续查看');
  }}>
    <p className={styles.note}>导入保存正文与来源版本。首次模型整理需在资料详情中明确提交；导入成功不表示已完成整理。</p>
    {fixture && <p role="note">示例模式：写操作尚未启用，请退出示例模式连接真实 API。</p>}
    <fieldset disabled={fixture || command.pending || reading}>
      <legend>导入来源</legend>
      <SelectField label="导入方式" value={adapter} onChange={value => setAdapter(value as typeof adapter)} options={[{ value: 'arxiv', label: 'arXiv 论文' }, { value: 'summarize_url', label: '视频链接 · summarize 获取正文' }, { value: 'summarize', label: 'summarize 既有导出' }]} />
      <TextField label={adapter === 'arxiv' ? 'arXiv 来源' : adapter === 'summarize_url' ? '视频链接' : '视频原始来源'} value={locator} onChange={setLocator} required type={adapter === 'summarize_url' ? 'url' : 'text'} hint={adapter === 'arxiv' ? '输入 arXiv 论文 ID 或公开链接。后端获取真实内容，错误会保留在导入队列。' : adapter === 'summarize_url' ? '填写公开视频 URL。后台用 summarize 获取正文或字幕，无需自行制作导出；缺字幕或需要登录时会说明下一步。' : '填写导出对应的真实视频 URL；已有摘要若无字幕位置，将明确显示未提供。'} />
      <TextField label="标题（可选）" value={title} onChange={setTitle} hint="获取内容时保留来源标题；纯文本导出可使用此标题。" />
      <TextField label="收藏理由（可选）" value={reason} onChange={setReason} multiline hint="留空会显示未提供，不补写为你的观点。导入后可以固定到收藏。" />
      {adapter === 'summarize' && <>
        <label>选择既有导出文件<input type="file" accept=".json,.txt,.md,application/json,text/plain,text/markdown" onChange={async event => {
          const file = event.target.files?.[0];
          if (!file) return;
          setReading(true);
          setFileError(null);
          try { setExportText(await file.text()); } catch { setFileError('无法读取导出文件，请重新选择或粘贴内容。'); } finally { setReading(false); }
        }} /></label>
        <TextField label="summarize 导出内容" value={exportText} onChange={setExportText} multiline required hint="粘贴已有 JSON、Markdown 或纯文本，保留导出中的原文、摘要和真实时间区间。" />
      </>}
    </fieldset>
    {reading && <p role="status">正在读取本地导出…</p>}
    {fileError && <p role="alert">{fileError}</p>}
    <CommandState {...command} />
    {command.error && adapter === 'summarize_url' && <p className={styles.note}>视频链接和输入已保留。按错误提示处理后再明确提交；也可切换为既有导出。登录内容的获取范围尚待确认。</p>}
    <button className="ac-button" type="submit" disabled={fixture || command.pending || reading}>导入资料</button>
  </form>;
}
