import { useState } from 'react';
import { useCommand } from '../../hooks/useCommand';

/** 抽屉关闭后保留页面草稿及失败命令身份。 */
export function useImportDraft() {
  const [adapter, setAdapter] = useState<'arxiv' | 'summarize_url' | 'summarize'>('arxiv');
  const [locator, setLocator] = useState('');
  const [reason, setReason] = useState('');
  const [title, setTitle] = useState('');
  const [exportText, setExportText] = useState('');
  const [exportKind, setExportKind] = useState<'video' | 'paper'>('video');
  const [refresh, setRefresh] = useState(false);
  const [fileError, setFileError] = useState<string | null>(null);
  const [reading, setReading] = useState(false);
  const command = useCommand();
  return { adapter, setAdapter, locator, setLocator, reason, setReason, title, setTitle, exportText, setExportText, exportKind, setExportKind, refresh, setRefresh, fileError, setFileError, reading, setReading, command };
}
