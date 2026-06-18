/// <reference types="vitest" />
import { fileURLToPath, URL } from 'node:url';
import { loadEnv, type ProxyOptions } from 'vite';
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  // Dev only: where /graphql and /pictures are proxied. Same-origin in prod (ADR-0003).
  const target = env.VITE_DEV_API_PROXY_TARGET || 'https://localhost:8443';

  // Dev gateway (ADR-0006): inject the AMBIENT browse credential server-side ONLY when the
  // request has no Authorization, so a signed-in client's own credential passes through for
  // orders. Read from a gitignored .env — NOT a VITE_ var, so it never reaches the bundle.
  const email = env.DEV_CUSTOMER_EMAIL;
  const password = env.DEV_CUSTOMER_PASSWORD;
  const ambient =
    email && password
      ? `Basic ${Buffer.from(`${email}:${password}`).toString('base64')}`
      : undefined;

  const base = { target, changeOrigin: true, secure: false }; // secure:false: self-signed dev TLS
  const graphqlProxy: ProxyOptions = {
    ...base,
    configure: (proxy) => {
      proxy.on('proxyReq', (proxyReq, req) => {
        if (ambient && !req.headers.authorization) proxyReq.setHeader('Authorization', ambient);
      });
    },
  };

  return {
    base: '/',
    plugins: [react()],
    resolve: {
      alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
    },
    server: {
      proxy: {
        // Only /graphql needs auth; /pictures is public catalog content.
        '/graphql': graphqlProxy,
        '/pictures': base,
      },
    },
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: ['./vitest.setup.ts'],
      css: true,
    },
  };
});
