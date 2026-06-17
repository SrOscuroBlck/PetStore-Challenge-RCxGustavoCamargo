import { Navigate } from 'react-router-dom';
import { env } from '@/lib/env';
import { Logo } from '@/components/brand/Logo';

/**
 * Customers arrive via a store URL (/store/:id). Bare "/" goes to the dev default store if
 * configured; otherwise a branded landing (this path is not part of the customer journey).
 */
export function RootRedirect() {
  if (env.defaultStoreId) {
    return <Navigate to={`/store/${env.defaultStoreId}`} replace />;
  }
  return (
    <main className="flex min-h-dvh flex-col items-center justify-center gap-3 bg-ink p-6 text-center text-ink-fg">
      <Logo />
      <p className="eyebrow mt-4">Welcome</p>
      <h1 className="text-2xl font-bold">Open your store link to start browsing</h1>
      <p className="max-w-sm text-ink-fg/70">
        Each store has its own address, e.g. <span className="font-mono text-accent">/store/&lt;id&gt;</span>.
      </p>
    </main>
  );
}
