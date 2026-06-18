import { createContext } from 'react';

export type ToastVariant = 'default' | 'error' | 'success';

export interface ShowToastInput {
  title?: string;
  description: string;
  variant?: ToastVariant;
}

export interface ToastApi {
  show: (toast: ShowToastInput) => void;
}

export const ToastContext = createContext<ToastApi | null>(null);
