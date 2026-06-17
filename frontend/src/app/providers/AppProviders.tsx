import { type ReactNode } from 'react';
import { ApolloProvider } from '@apollo/client';
import { makeApolloClient } from '@/api/apollo/client';
import { ToastProvider } from '@/components/ui/Toast/ToastProvider';
import { AuthProvider } from '@/features/auth/AuthProvider';
import { CartProvider } from '@/features/cart/CartProvider';
import { ErrorBoundary } from '@/components/feedback/ErrorBoundary';

// One client for the app's lifetime.
const client = makeApolloClient();

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary>
      <ApolloProvider client={client}>
        <AuthProvider>
          <ToastProvider>
            <CartProvider>{children}</CartProvider>
          </ToastProvider>
        </AuthProvider>
      </ApolloProvider>
    </ErrorBoundary>
  );
}
