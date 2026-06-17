import { MockedProvider, type MockedResponse } from '@apollo/client/testing';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { AvailablePetsDocument, PetStatus, Species } from '@/__generated__/graphql';
import { ToastProvider } from '@/components/ui/Toast/ToastProvider';
import { AuthProvider } from '@/features/auth/AuthProvider';
import { CartProvider } from '@/features/cart/CartProvider';
import { CatalogSection } from '../CatalogSection';
import { PAGE_SIZE } from '../useAvailablePets';

const STORE = 'store-1';

// CatalogSection renders cart/purchase/login UI; it needs the Auth/Toast/Cart providers and a
// router (the login dialog reads the store id from the route).
function renderCatalog(mocks: MockedResponse[]) {
  return render(
    <MockedProvider mocks={mocks}>
      <AuthProvider>
        <ToastProvider>
          <CartProvider>
            <MemoryRouter initialEntries={[`/store/${STORE}`]}>
              <Routes>
                <Route path="/store/:storeId" element={<CatalogSection storeId={STORE} />} />
              </Routes>
            </MemoryRouter>
          </CartProvider>
        </ToastProvider>
      </AuthProvider>
    </MockedProvider>,
  );
}
const variables = { storeId: STORE, species: null, first: PAGE_SIZE };

const populated: MockedResponse = {
  request: { query: AvailablePetsDocument, variables },
  result: {
    data: {
      availablePets: {
        __typename: 'PublicPetConnection',
        edges: [
          {
            __typename: 'PublicPetEdge',
            cursor: 'c1',
            node: {
              __typename: 'PublicPet',
              id: '1',
              name: 'Whiskers',
              species: Species.Cat,
              ageYears: 2,
              description: 'A calm cat.',
              pictureUrl: '/pictures/1',
              status: PetStatus.Available,
              createdAt: '2026-01-01T00:00:00Z',
            },
          },
          {
            __typename: 'PublicPetEdge',
            cursor: 'c2',
            node: {
              __typename: 'PublicPet',
              id: '2',
              name: 'Rex',
              species: Species.Dog,
              ageYears: 1,
              description: 'A playful dog.',
              pictureUrl: '/pictures/2',
              status: PetStatus.Available,
              createdAt: '2026-01-02T00:00:00Z',
            },
          },
        ],
        pageInfo: { __typename: 'PageInfo', hasNextPage: false, endCursor: 'c2' },
      },
    },
  },
};

const empty: MockedResponse = {
  request: { query: AvailablePetsDocument, variables },
  result: {
    data: {
      availablePets: {
        __typename: 'PublicPetConnection',
        edges: [],
        pageInfo: { __typename: 'PageInfo', hasNextPage: false, endCursor: null },
      },
    },
  },
};

describe('CatalogSection', () => {
  it('shows a skeleton on first load (AC3)', () => {
    renderCatalog([populated]);
    expect(screen.getByTestId('catalog-skeleton')).toBeInTheDocument();
  });

  it('renders pets with name, species and age once loaded (AC1)', async () => {
    renderCatalog([populated]);
    expect(await screen.findByText('Whiskers')).toBeInTheDocument();
    expect(screen.getByText('Rex')).toBeInTheDocument();
    expect(screen.getByText('2 yrs')).toBeInTheDocument(); // Whiskers
    expect(screen.getByText('1 yr')).toBeInTheDocument(); // Rex (singular)
  });

  it('shows the friendly empty state when there are no pets (AC4/5)', async () => {
    renderCatalog([empty]);
    expect(await screen.findByText(/no pets available/i)).toBeInTheDocument();
  });
});
