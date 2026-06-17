import { useEffect, useRef } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { Logo } from '@/components/brand/Logo';
import { CartButton } from '@/features/cart/CartButton';
import { useAuth } from '@/features/auth/useAuth';

/** Storefront chrome: the RC-navy header (brand + future cart), and the routed content. */
export function AppShell() {
  const location = useLocation();
  const auth = useAuth();
  const mainRef = useRef<HTMLElement>(null);

  // Move focus to the content region on navigation (a11y).
  useEffect(() => {
    mainRef.current?.focus();
  }, [location.pathname]);

  return (
    <div className="flex min-h-dvh flex-col bg-bg">
      <a
        href="#main-content"
        className="sr-only-focusable absolute left-4 top-4 z-50 rounded bg-primary px-3 py-2 text-primary-fg"
      >
        Skip to content
      </a>
      <header className="sticky top-0 z-30 border-b border-white/10 bg-ink text-ink-fg">
        <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-4 sm:px-6">
          <Link to={location.pathname} className="rounded-sm" aria-label="Petstore home">
            <Logo />
          </Link>
          <div className="flex items-center gap-1 sm:gap-3">
            {auth.isSignedIn ? (
              <>
                <span className="hidden text-sm text-ink-fg/70 sm:inline">{auth.email}</span>
                <button
                  type="button"
                  onClick={auth.logout}
                  className="rounded-sm px-2 py-1 text-sm text-ink-fg/80 transition-colors hover:text-ink-fg"
                >
                  Log out
                </button>
              </>
            ) : null}
            <CartButton />
          </div>
        </div>
      </header>
      <main id="main-content" ref={mainRef} tabIndex={-1} className="flex-1 outline-none">
        <Outlet />
      </main>
      <footer className="border-t border-border bg-card">
        <div className="mx-auto w-full max-w-6xl px-4 py-6 text-sm text-muted sm:px-6">
          Pets available for adoption — browse and take one home.
        </div>
      </footer>
    </div>
  );
}
