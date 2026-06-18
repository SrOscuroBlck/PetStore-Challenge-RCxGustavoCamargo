/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Dev-only proxy target for /graphql and /pictures (see vite.config.ts). */
  readonly VITE_DEV_API_PROXY_TARGET?: string;
  /** Optional dev convenience: store id to redirect a bare "/" to. */
  readonly VITE_DEFAULT_STORE_ID?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
