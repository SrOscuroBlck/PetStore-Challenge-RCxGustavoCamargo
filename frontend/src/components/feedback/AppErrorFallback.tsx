import { Button } from '@/components/ui/Button';
import { Logo } from '@/components/brand/Logo';

interface AppErrorFallbackProps {
  onReload: () => void;
}

/** Shown when the top-level ErrorBoundary catches a render crash. No detail leaks. */
export function AppErrorFallback({ onReload }: AppErrorFallbackProps) {
  return (
    <main className="flex min-h-dvh flex-col items-center justify-center gap-4 bg-ink p-6 text-center text-ink-fg">
      <Logo />
      <p className="eyebrow mt-4">Something broke</p>
      <h1 className="text-2xl font-bold">An unexpected error occurred</h1>
      <p className="max-w-md text-ink-fg/70">Reloading usually fixes it.</p>
      <Button variant="primary" onClick={onReload} className="mt-2">
        Reload
      </Button>
    </main>
  );
}
