import { useState, type FormEvent } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useParams } from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import { TextField } from '@/components/ui/TextField';
import { useAuth } from './useAuth';

/**
 * Sign-in modal opened via `ensureSignedIn()` when placing an order. Validates by login-by-probe
 * against the routed store; on success the AuthProvider closes it and resumes the pending action.
 */
export function LoginDialog() {
  const { isLoginOpen, closeLogin, login } = useAuth();
  const { storeId } = useParams();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!storeId) return;
    setSubmitting(true);
    setError(null);
    const result = await login(email, password, storeId);
    setSubmitting(false);
    if (!result.ok) setError(result.errorMessage ?? 'Sign in failed.');
  };

  return (
    <Dialog.Root
      open={isLoginOpen}
      onOpenChange={(open) => {
        if (!open) closeLogin();
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[60] bg-black/50" />
        <Dialog.Content className="fixed left-1/2 top-1/2 z-[61] w-full max-w-sm -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-card p-6 shadow-xl focus:outline-none">
          <Dialog.Title className="font-display text-xl font-bold">Sign in to buy</Dialog.Title>
          <Dialog.Description className="mt-1 text-sm text-muted">
            Browsing is open — sign in to place your order.
          </Dialog.Description>
          <form onSubmit={(e) => void handleSubmit(e)} className="mt-5 flex flex-col gap-4" noValidate>
            <TextField
              label="Email"
              type="email"
              autoComplete="username"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={submitting}
            />
            <TextField
              label="Password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={submitting}
            />
            {error ? (
              <p role="alert" className="text-sm text-danger">
                {error}
              </p>
            ) : null}
            <div className="mt-1 flex justify-end gap-2">
              <Dialog.Close asChild>
                <Button variant="secondary" type="button">
                  Cancel
                </Button>
              </Dialog.Close>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Signing in…' : 'Sign in'}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
