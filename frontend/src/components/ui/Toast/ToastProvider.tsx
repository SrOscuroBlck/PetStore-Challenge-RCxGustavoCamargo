import { useCallback, useMemo, useState, type ReactNode } from 'react';
import * as ToastPrimitive from '@radix-ui/react-toast';
import { cn } from '@/lib/cn';
import {
  ToastContext,
  type ShowToastInput,
  type ToastApi,
  type ToastVariant,
} from './toastContext';

interface ToastMessage extends ShowToastInput {
  id: number;
  variant: ToastVariant;
}

const variantStyles: Record<ToastVariant, string> = {
  default: 'border-primary',
  error: 'border-danger',
  success: 'border-success',
};

let nextId = 0;

/**
 * App-wide toast host built on Radix (accessible live region, focus & swipe handling).
 * Used by the catalog/purchase flows (#2/#3) to surface error-code-mapped messages.
 */
export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const show = useCallback((toast: ShowToastInput) => {
    nextId += 1;
    setToasts((prev) => [...prev, { id: nextId, variant: 'default', ...toast }]);
  }, []);

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  const api = useMemo<ToastApi>(() => ({ show }), [show]);

  return (
    <ToastContext.Provider value={api}>
      <ToastPrimitive.Provider swipeDirection="right">
        {children}
        {toasts.map((toast) => (
          <ToastPrimitive.Root
            key={toast.id}
            duration={6000}
            onOpenChange={(open) => {
              if (!open) remove(toast.id);
            }}
            className={cn(
              'relative rounded-xl border-l-4 bg-card p-4 pr-8 text-sm text-fg shadow-soft',
              'data-[state=open]:animate-in data-[state=closed]:animate-out',
              variantStyles[toast.variant],
            )}
          >
            {toast.title ? (
              <ToastPrimitive.Title className="font-display font-semibold">
                {toast.title}
              </ToastPrimitive.Title>
            ) : null}
            <ToastPrimitive.Description>{toast.description}</ToastPrimitive.Description>
            <ToastPrimitive.Close
              aria-label="Dismiss"
              className="absolute right-2 top-2 text-muted hover:text-fg"
            >
              ✕
            </ToastPrimitive.Close>
          </ToastPrimitive.Root>
        ))}
        <ToastPrimitive.Viewport className="fixed bottom-0 right-0 z-50 m-4 flex w-96 max-w-[100vw] flex-col gap-2 outline-none" />
      </ToastPrimitive.Provider>
    </ToastContext.Provider>
  );
}
