import { useEffect, useRef, useState } from 'react';
import { ApiError } from '../../api/client';
import { formatError } from '../../hooks/useReadApi';
import type { RelayActionResult, RelayTaskResult } from '../../api/relay';
import { relayExistingTaskID } from './relayPresentation';

export function relayCommandError(cause: unknown) {
  // ErrorV1 使用服务端过滤后的错误；传输异常使用公开恢复说明。
  return cause instanceof ApiError ? formatError(cause) : '接力请求结果未知，请只读取任务与收据核对。';
}

/** 请求结果未知时保留任务和操作身份，只允许读取核对。 */
export function useRelayCommand<T extends RelayActionResult | RelayTaskResult = RelayActionResult>(scope: string) {
  const storageKey = `astrocyte:relay:pending:${scope}`;
  const savedOperation = typeof window === 'undefined' ? '' : window.sessionStorage.getItem(storageKey) ?? '';
  const [pending, setPending] = useState(false);
  const [uncertain, setUncertain] = useState(!!savedOperation);
  const [error, setError] = useState('');
  const [existingTaskID, setExistingTaskID] = useState('');
  const [result, setResult] = useState<T | null>(null);
  const [operationID, setOperationID] = useState(savedOperation);
  const busy = useRef(false), mounted = useRef(false), blocked = useRef(!!savedOperation);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const run = async (action: (key: string) => Promise<T>, received: (result: T) => void) => {
    if (busy.current || blocked.current) return;
    const key = crypto.randomUUID();
    // 仅保存操作身份；重开面板或刷新后保持未知锁，不保存输入或响应原文。
    window.sessionStorage.setItem(storageKey, key);
    busy.current = true; setPending(true); setError(''); setExistingTaskID(''); setOperationID(key);
    try {
      const response = await action(key);
      const unresolved = 'receipt' in response && ['pending', 'delivery_unknown'].includes(response.receipt.status);
      if (!unresolved) window.sessionStorage.removeItem(storageKey);
      if (!mounted.current) return;
      setResult(response);
      if ('receipt' in response) {
        if (unresolved) { blocked.current = true; setUncertain(true); }
        if (response.receipt.error) setError(`${response.receipt.error.code} · ${response.receipt.error.message} · ${response.receipt.error.required_action}`);
        setExistingTaskID(response.receipt.status === 'rejected' ? relayExistingTaskID(response.receipt.error) : '');
      }
      received(response);
    } catch (cause) {
      const unknown = !(cause instanceof ApiError) || ['delivery_unknown', 'internal_error'].includes(cause.code);
      if (!unknown) window.sessionStorage.removeItem(storageKey);
      if (!mounted.current) return;
      if (unknown) { blocked.current = true; setUncertain(true); }
      setExistingTaskID(cause instanceof ApiError && cause.status === 409 ? relayExistingTaskID(cause.detail.error) : '');
      setError(relayCommandError(cause));
    } finally { busy.current = false; if (mounted.current) setPending(false); }
  };
  return { pending, uncertain, error, existingTaskID, result, operationID, run };
}
