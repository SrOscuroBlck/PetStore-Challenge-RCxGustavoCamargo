import { useEffect, useRef } from 'react';

interface InfiniteScrollOptions {
  /** Called when the sentinel scrolls into view. Guard re-entrancy in the caller. */
  onLoadMore: () => void;
  /** Observe only while true (e.g. hasNextPage && !loadingMore). */
  enabled: boolean;
  rootMargin?: string;
}

/**
 * Returns a ref to attach to a sentinel element. Auto-loads when it nears the viewport.
 * Feature-detected — if IntersectionObserver is absent, the caller's "Load more" button
 * is the fallback. Disconnects while disabled so a mid-load fetch can't double-fire.
 */
export function useInfiniteScroll<T extends HTMLElement>({
  onLoadMore,
  enabled,
  rootMargin = '300px',
}: InfiniteScrollOptions) {
  const sentinelRef = useRef<T>(null);
  const onLoadMoreRef = useRef(onLoadMore);
  onLoadMoreRef.current = onLoadMore;

  useEffect(() => {
    if (!enabled) return;
    const el = sentinelRef.current;
    if (!el || typeof IntersectionObserver === 'undefined') return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) onLoadMoreRef.current();
      },
      { rootMargin },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [enabled, rootMargin]);

  return sentinelRef;
}
