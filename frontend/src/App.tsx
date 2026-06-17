import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { AppShell } from '@/app/AppShell';
import { StoreRoute } from '@/routes/StoreRoute';
import { RootRedirect } from '@/routes/RootRedirect';
import { NotFoundRoute } from '@/routes/NotFoundRoute';

// Open storefront (ADR-0005): no auth gate — the gateway authenticates requests.
const router = createBrowserRouter([
  { path: '/', element: <RootRedirect /> },
  {
    path: '/store/:storeId',
    element: <AppShell />,
    children: [{ index: true, element: <StoreRoute /> }],
  },
  { path: '*', element: <NotFoundRoute /> },
]);

export function App() {
  return <RouterProvider router={router} />;
}
