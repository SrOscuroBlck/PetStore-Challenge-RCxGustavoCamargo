import { Link } from 'react-router-dom';
import { Logo } from '@/components/brand/Logo';

export function NotFoundRoute() {
  return (
    <main className="flex min-h-dvh flex-col items-center justify-center gap-4 bg-bg p-6 text-center">
      <Logo />
      <p className="eyebrow mt-4">Error 404</p>
      <h1 className="text-3xl font-bold">Page not found</h1>
      <p className="max-w-sm text-muted">The page you’re looking for doesn’t exist or has moved.</p>
      <Link
        to="/"
        className="mt-2 rounded-sm bg-primary px-5 py-2.5 font-display text-sm font-semibold uppercase tracking-wide text-primary-fg transition-colors hover:bg-primary-strong"
      >
        Go home
      </Link>
    </main>
  );
}
