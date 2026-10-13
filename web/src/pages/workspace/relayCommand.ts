import { useEffect, useRef, useState } from 'react';
import { ApiError } from '../../api/client';
import { formatError } from '../../hooks/useReadApi';
import type { RelayActionResult, RelayTaskResult } from '../../api/relay';
import { relayExistingTaskID } from './relayPresentation';

export function relayCommandError(cause: unknown) {
  // ErrorV1 has passed H's closed sanitizer; arbitrary transport errors have not.
  return cause instanceof ApiError ? formatError(cause) : '接力请求结果未知，请只读取任务与收据核对。';
}

/** The UI never replays a write after uncertain delivery or invents a task state. */
export function useRelayCommand<T extends RelayActionResult | RelayTaskResult = RelayActionResult>() {
  const [pending, setPending] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [error, setError] = useState('');
  const [existingTaskID, setExistingTaskID] = useState('');
  const [result, setResult] = useState<T | null>(null);
  const [operationID, setOperationID] = useState('');
  const busy = useRef(false), mounted = useRef(false), blocked = useRef(false);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const run = async (action: (key: string) => Promise<T>, received: (result: T) => void) => {
    if (busy.current || blocked.current) return;
    const key = crypto.randomUUID();
    busy.current = true; setPending(true); setError(''); setExistingTaskID(''); setOperationID(key);
    try {
      const response = await action(key);
      if (!mounted.current) return;
      setResult(response);
      if ('receipt' in response) {
        if (['pending', 'delivery_unknown'].includes(response.receipt.status)) { blocked.current = true; setUncertain(true); }
        if (response.receipt.error) setError(`${response.receipt.error.code} · ${response.receipt.error.message} · ${response.receipt.error.required_action}`);
        setExistingTaskID(response.receipt.status === 'rejected' ? relayExistingTaskID(response.receipt.error) : '');
      }
      received(response);
    } catch (cause) {
      if (!mounted.current) return;
      if (!(cause instanceof ApiError) || ['delivery_unknown', 'internal_error'].includes(cause.code)) { blocked.current = true; setUncertain(true); }
      setExistingTaskID(cause instanceof ApiError && cause.status === 409 ? relayExistingTaskID(cause.detail.error) : '');
      setError(relayCommandError(cause));
    } finally { busy.current = false; if (mounted.current) setPending(false); }
  };
  return { pending, uncertain, error, existingTaskID, result, operationID, run };
}
