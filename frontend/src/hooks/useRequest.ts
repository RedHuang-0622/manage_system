import { useEffect, useRef, useCallback } from 'react';

/**
 * React StrictMode fires every effect twice (mount→unmount→mount).
 * This doubles every API call, saturates Chrome's 6-connection-per-origin
 * limit, and causes 8-second axios timeouts after ~3 page navigations.
 *
 * useRequest deduplicates: the first mount's call is cancelled via
 * setTimeout(0), only the second mount's call survives.
 * In production (no StrictMode double-fire) it behaves identically
 * to a normal useEffect with the same deps.
 */
export function useRequest(effect: () => void, deps: React.DependencyList) {
  const ranOnce = useRef(false);

  useEffect(() => {
    // StrictMode: first mount → cleanup kills the timer.
    // Second mount → fires normally. Net result: 1 call instead of 2.
    if (!ranOnce.current) {
      ranOnce.current = true;
      const id = setTimeout(effect, 0);
      return () => clearTimeout(id);
    }
    effect();
  }, deps); // eslint-disable-line react-hooks/exhaustive-deps
}

/**
 * Like useRequest but returns a refetch function.
 */
export function useRequestWithRefetch(
  fetchFn: () => Promise<void>,
  deps: React.DependencyList,
) {
  const ranOnce = useRef(false);

  useEffect(() => {
    if (!ranOnce.current) {
      ranOnce.current = true;
      const id = setTimeout(fetchFn, 0);
      return () => clearTimeout(id);
    }
    fetchFn();
  }, deps); // eslint-disable-line react-hooks/exhaustive-deps

  return useCallback(() => { fetchFn(); }, [fetchFn]);
}
