import { useCallback, useEffect, useRef, useState } from 'react';
import { formatError } from './useReadApi';

/** Commands only run after an explicit action. Never automatically retry writes. */
export function useCommand(disabled = false) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const busy = useRef(false);
  const mounted = useRef(false);
  const lastRequest = useRef<{ signature: string; key: string } | null>(null);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);

  const run = useCallback(async <T,>(command: () => Promise<T>, onSuccess?: (result: T) => void, message = '已保存') => {
    if (disabled || busy.current) return;
    busy.current = true;
    setPending(true);
    setError(null);
    setNotice(null);
    try {
      const result = await command();
      if (mounted.current) {
        lastRequest.current = null;
        onSuccess?.(result);
        setNotice(message);
      }
    } catch (err) {
      if (mounted.current) setError(formatError(err));
    } finally {
      busy.current = false;
      if (mounted.current) setPending(false);
    }
  }, [disabled]);

  const prepare = <T,>(payload: T) => {
    const signature = JSON.stringify(payload);
    if (lastRequest.current?.signature !== signature) lastRequest.current = { signature, key: crypto.randomUUID() };
    const key = lastRequest.current.key;
    return { body: { ...payload, schema_version: 1 as const, request_id: key }, key };
  };
  return { pending, error, notice, run, prepare };
}
