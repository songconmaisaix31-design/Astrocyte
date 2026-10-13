import { useEffect, useRef, useState } from 'react';
import { ApiError } from '../../api/client';
import { formatError } from '../../hooks/useReadApi';
import type { RelayActionResult, RelayTaskResult } from '../../api/relay';

/** The UI never replays a write after uncertain delivery or invents a task state. */
export function useRelayCommand<T extends RelayActionResult | RelayTaskResult = RelayActionResult>() {
  const [pending, setPending] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<T | null>(null);
  const [operationID, setOperationID] = useState('');
  const busy = useRef(false), mounted = useRef(false), blocked = useRef(false);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const run = async (action: (key: string) => Promise<T>, received: (result: T) => void) => {
    if (busy.current || blocked.current) return;
    const key = crypto.randomUUID();
    busy.current = true; setPending(true); setError(''); setOperationID(key);
    try {
      const response = await action(key);
      if (!mounted.current) return;
      setResult(response);
      if ('receipt' in response) {
        if (['pending', 'delivery_unknown'].includes(response.receipt.status)) { blocked.current = true; setUncertain(true); }
        if (response.receipt.error) setError(`${response.receipt.error.code} · ${response.receipt.error.message} · ${response.receipt.error.required_action}`);
      }
      received(response);
    } catch (cause) {
      if (!mounted.current) return;
      if (!(cause instanceof ApiError) || ['delivery_unknown', 'internal_error'].includes(cause.code)) { blocked.current = true; setUncertain(true); }
      setError(formatError(cause));
    } finally { busy.current = false; if (mounted.current) setPending(false); }
  };
  return { pending, uncertain, error, result, operationID, run };
}
