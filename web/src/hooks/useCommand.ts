import { useCallback, useEffect, useRef, useState } from 'react';
import { formatError } from './useReadApi';

/** Commands only run after an explicit action. Never automatically retry writes. */
export function useCommand(disabled = false) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const busy = useRef(false);
  const mounted = useRef(false);
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

  return { pending, error, notice, run };
}
