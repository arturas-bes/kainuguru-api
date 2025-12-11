# Kainuguru Frontend Implementation Plan

> Comprehensive development roadmap from UX design to production deployment

**Version:** 1.0
**Date:** 2025-12-10
**Status:** Approved for Development

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Technology Stack](#2-technology-stack)
3. [Project Structure](#3-project-structure)
4. [Phase 1: MVP Core](#4-phase-1-mvp-core)
5. [Phase 2: Enhanced Features](#5-phase-2-enhanced-features)
6. [Phase 3: Advanced Features](#6-phase-3-advanced-features)
7. [GraphQL Integration](#7-graphql-integration)
8. [Component Library](#8-component-library)
9. [State Management](#9-state-management)
10. [Testing Strategy](#10-testing-strategy)
11. [Performance Requirements](#11-performance-requirements)
12. [Accessibility Requirements](#12-accessibility-requirements)
13. [Deployment Strategy](#13-deployment-strategy)
14. [Risk Mitigation](#14-risk-mitigation)

---

## 1. Executive Summary

### 1.1 Project Overview

Build a modern, mobile-first frontend for Kainuguru - a Lithuanian grocery price comparison platform. The frontend will consume an existing GraphQL API that provides:

- Store and flyer management
- Product search with faceted filtering
- Shopping list management
- Migration wizard for expired deals
- Price alerts and history
- User authentication

### 1.2 Development Phases

| Phase | Duration | Focus | Deliverables |
|-------|----------|-------|--------------|
| Phase 1 (MVP) | 5-6 weeks | Core functionality | Home, Search, Flyers, Lists, Auth |
| Phase 2 | 3-4 weeks | Enhanced features | Wizard, Alerts, Mobile Nav |
| Phase 3 | 3-4 weeks | Advanced features | PWA, History, Preferences |

### 1.3 Key Decisions from UX Research

| Decision | Rationale |
|----------|-----------|
| **Include bounding boxes in MVP** | Backend data exists; 2-3 hours CSS work |
| **Backend owns wizard state** | `WizardSession` manages all wizard state server-side |
| **Frontend-only store filter** | No backend preference storage; use localStorage |
| **Layer features, don't cut** | Backend investment already made; phase complexity |
| **Simplified home page** | 3 sections: Hero/Search + Flyers + Deals |

---

## 2. Technology Stack

### 2.1 Core Framework

```
Framework:      Next.js 14+ (App Router)
Language:       TypeScript 5.x (strict mode)
Styling:        Tailwind CSS 3.x + CSS Modules for complex components
State:          React Query (TanStack Query) for server state
                Zustand for client state
GraphQL:        urql or Apollo Client 3.x
Forms:          React Hook Form + Zod validation
```

### 2.2 Supporting Libraries

```
UI Components:  Radix UI primitives (accessible, unstyled)
Icons:          Lucide React
Charts:         Recharts (price history)
Date:           date-fns + date-fns/locale/lt
Images:         Next.js Image optimization
Animation:      Framer Motion (minimal, purposeful)
Testing:        Vitest + React Testing Library + Playwright
```

### 2.3 Development Tools

```
Package Manager:  pnpm
Linting:          ESLint + Prettier
Type Generation:  GraphQL Code Generator
Git Hooks:        Husky + lint-staged
CI/CD:            GitHub Actions
Deployment:       Vercel (recommended) or Docker
```

### 2.4 Design Token Integration

```typescript
// tailwind.config.ts
import type { Config } from 'tailwindcss';

const config: Config = {
  theme: {
    extend: {
      colors: {
        brand: {
          primary: '#2E7D32',
          'primary-light': '#4CAF50',
          'primary-dark': '#1B5E20',
        },
        store: {
          maxima: '#E31E24',
          lidl: '#0050AA',
          iki: '#E30613',
          rimi: '#D52B1E',
          norfa: '#009639',
        },
        sale: {
          DEFAULT: '#FF6D00',
          light: '#FF9E40',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
};
```

---

## 3. Project Structure

```
/kainuguru-frontend
├── /app                          # Next.js App Router
│   ├── /(public)                 # Public routes group
│   │   ├── /page.tsx             # Home
│   │   ├── /search/page.tsx      # Search results
│   │   ├── /flyers/page.tsx      # Flyer listing
│   │   ├── /flyers/[id]/page.tsx # Flyer detail
│   │   ├── /products/[id]/page.tsx # Product detail
│   │   ├── /login/page.tsx       # Login
│   │   └── /register/page.tsx    # Register
│   ├── /(authenticated)          # Protected routes group
│   │   ├── /lists/page.tsx       # Shopping lists
│   │   ├── /lists/[id]/page.tsx  # List detail
│   │   ├── /account/page.tsx     # User account
│   │   └── /alerts/page.tsx      # Price alerts (Phase 2)
│   ├── /api                      # API routes (if needed)
│   ├── layout.tsx                # Root layout
│   ├── error.tsx                 # Error boundary
│   ├── loading.tsx               # Loading state
│   └── not-found.tsx             # 404 page
│
├── /components
│   ├── /ui                       # Design system primitives
│   │   ├── Button.tsx
│   │   ├── Card.tsx
│   │   ├── Input.tsx
│   │   ├── Select.tsx
│   │   ├── Modal.tsx
│   │   ├── Badge.tsx
│   │   ├── Skeleton.tsx
│   │   ├── Toast.tsx
│   │   └── index.ts
│   ├── /layout
│   │   ├── Header.tsx
│   │   ├── Footer.tsx
│   │   ├── MobileNav.tsx
│   │   ├── PageShell.tsx
│   │   └── index.ts
│   └── /shared                   # Shared business components
│       ├── ProductCard.tsx
│       ├── StoreFilter.tsx
│       ├── PriceDisplay.tsx
│       ├── SaleBadge.tsx
│       └── index.ts
│
├── /features                     # Feature modules
│   ├── /home
│   │   ├── HomePage.tsx
│   │   ├── HeroSection.tsx
│   │   ├── FlyerCarousel.tsx
│   │   ├── DealsGrid.tsx
│   │   └── hooks/
│   │       └── useHomeSummary.ts
│   ├── /flyers
│   │   ├── FlyersPage.tsx
│   │   ├── FlyerDetailPage.tsx
│   │   ├── FlyerCard.tsx
│   │   ├── FlyerPageViewer.tsx
│   │   ├── ProductHotspot.tsx
│   │   └── hooks/
│   │       ├── useFlyers.ts
│   │       └── useFlyerDetail.ts
│   ├── /search
│   │   ├── SearchPage.tsx
│   │   ├── SearchFilters.tsx
│   │   ├── ProductGrid.tsx
│   │   ├── FacetSidebar.tsx
│   │   └── hooks/
│   │       └── useProductSearch.ts
│   ├── /shopping-list
│   │   ├── ShoppingListsPage.tsx
│   │   ├── ShoppingListDetailPage.tsx
│   │   ├── ListCard.tsx
│   │   ├── ListItem.tsx
│   │   ├── AddItemBar.tsx
│   │   └── hooks/
│   │       ├── useShoppingLists.ts
│   │       └── useShoppingList.ts
│   ├── /wizard                   # Phase 2
│   │   ├── WizardOverlay.tsx
│   │   ├── WizardItemCard.tsx
│   │   ├── SuggestionCard.tsx
│   │   ├── WizardSummary.tsx
│   │   └── hooks/
│   │       └── useWizard.ts
│   ├── /auth
│   │   ├── LoginForm.tsx
│   │   ├── RegisterForm.tsx
│   │   └── hooks/
│   │       └── useAuth.ts
│   └── /account
│       ├── AccountPage.tsx
│       ├── ProfileForm.tsx
│       ├── AlertsList.tsx        # Phase 2
│       └── hooks/
│           └── useAccount.ts
│
├── /graphql
│   ├── /fragments
│   │   ├── product.ts
│   │   ├── flyer.ts
│   │   ├── shoppingList.ts
│   │   └── wizard.ts
│   ├── /queries
│   │   ├── home.ts
│   │   ├── flyers.ts
│   │   ├── search.ts
│   │   ├── shoppingLists.ts
│   │   └── priceAlerts.ts
│   ├── /mutations
│   │   ├── auth.ts
│   │   ├── shoppingLists.ts
│   │   ├── wizard.ts
│   │   └── priceAlerts.ts
│   └── client.ts                 # GraphQL client setup
│
├── /hooks                        # Global hooks
│   ├── useStoreFilter.ts
│   ├── useLocalStorage.ts
│   ├── useMediaQuery.ts
│   └── useDebounce.ts
│
├── /lib
│   ├── dates.ts                  # Date utilities
│   ├── prices.ts                 # Price formatting
│   ├── confidence.ts             # Confidence score mapping
│   └── api.ts                    # API helpers
│
├── /stores                       # Zustand stores
│   ├── authStore.ts
│   ├── storeFilterStore.ts
│   └── uiStore.ts
│
├── /types
│   ├── generated.ts              # GraphQL codegen output
│   └── index.ts                  # Custom types
│
├── /public
│   ├── /images
│   ├── /icons
│   └── manifest.json             # PWA manifest (Phase 3)
│
└── /tests
    ├── /unit
    ├── /integration
    └── /e2e
```

---

## 4. Phase 1: MVP Core

### 4.1 Sprint 1: Foundation (Week 1-2)

#### Tasks

| Task | Priority | Estimate | Dependencies |
|------|----------|----------|--------------|
| Project setup (Next.js, TypeScript, Tailwind) | P0 | 4h | None |
| GraphQL client setup + codegen | P0 | 4h | None |
| Design token integration | P0 | 4h | None |
| UI component library (Button, Card, Input, Badge) | P0 | 16h | Design tokens |
| Layout components (Header, Footer, PageShell) | P0 | 8h | UI components |
| Auth store + context | P0 | 8h | GraphQL client |
| Store filter hook (localStorage) | P1 | 4h | None |

#### Deliverables
- [ ] Running Next.js app with TypeScript
- [ ] GraphQL client connected to backend
- [ ] Core UI components documented
- [ ] Layout with responsive header/footer
- [ ] Auth context (not wired to pages yet)

---

### 4.2 Sprint 2: Home + Flyers (Week 2-3)

#### Tasks

| Task | Priority | Estimate | Dependencies |
|------|----------|----------|--------------|
| Home page layout | P0 | 4h | Layout |
| Hero section with search | P0 | 8h | UI components |
| Flyer carousel component | P0 | 8h | GraphQL |
| Deals grid component | P0 | 8h | GraphQL |
| Flyers listing page | P0 | 8h | Layout |
| Flyer card component | P0 | 4h | UI components |
| Flyers filters (store, date) | P1 | 8h | Store filter |

#### GraphQL Queries Required

```graphql
# Home page
query HomeSummary($first: Int) {
  currentFlyers(first: $first) {
    edges {
      node {
        ...FlyerCard
      }
    }
  }
  productsOnSale(first: 12) {
    edges {
      node {
        ...ProductListItem
      }
    }
  }
}

# Flyers listing
query FlyersList($filters: FlyerFilters, $first: Int, $after: String) {
  flyers(filters: $filters, first: $first, after: $after) {
    edges {
      node {
        ...FlyerCard
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

#### Deliverables
- [ ] Home page with hero, flyer carousel, deals grid
- [ ] Flyers listing with filters and pagination
- [ ] All components mobile-responsive

---

### 4.3 Sprint 3: Flyer Detail + Search (Week 3-4)

#### Tasks

| Task | Priority | Estimate | Dependencies |
|------|----------|----------|--------------|
| Flyer detail page layout | P0 | 4h | Layout |
| Flyer page viewer (image + thumbnails) | P0 | 12h | None |
| Product hotspot overlay | P0 | 8h | Page viewer |
| Products list for flyer | P0 | 8h | Product card |
| Search results page | P0 | 8h | Layout |
| Faceted filter sidebar | P0 | 12h | GraphQL |
| Product grid with pagination | P0 | 8h | Product card |
| Search URL sync | P1 | 4h | Search |

#### Flyer Page Viewer Implementation

```typescript
// FlyerPageViewer.tsx
interface FlyerPageViewerProps {
  flyerPage: FlyerPage;
  products: Product[];
  onProductClick: (productId: number) => void;
}

export function FlyerPageViewer({ flyerPage, products, onProductClick }: FlyerPageViewerProps) {
  const [zoom, setZoom] = useState(1);
  const containerRef = useRef<HTMLDivElement>(null);

  return (
    <div ref={containerRef} className="relative overflow-hidden">
      {/* Main image */}
      <Image
        src={flyerPage.imageURL}
        alt={`Page ${flyerPage.pageNumber}`}
        width={flyerPage.imageWidth}
        height={flyerPage.imageHeight}
        className="w-full h-auto"
        style={{ transform: `scale(${zoom})` }}
      />

      {/* Product hotspots */}
      {products.map((product) => (
        product.boundingBox && (
          <ProductHotspot
            key={product.id}
            product={product}
            imageWidth={flyerPage.imageWidth}
            imageHeight={flyerPage.imageHeight}
            onClick={() => onProductClick(product.id)}
          />
        )
      ))}

      {/* Zoom controls */}
      <div className="absolute bottom-4 right-4 flex gap-2">
        <Button size="sm" onClick={() => setZoom(z => Math.max(1, z - 0.25))}>-</Button>
        <Button size="sm" onClick={() => setZoom(z => Math.min(3, z + 0.25))}>+</Button>
      </div>
    </div>
  );
}

// ProductHotspot.tsx
function ProductHotspot({ product, imageWidth, imageHeight, onClick }) {
  const { x, y, width, height } = product.boundingBox;

  return (
    <button
      className="absolute border-2 border-brand-primary/60 hover:border-brand-primary
                 hover:bg-brand-primary/10 transition-colors cursor-pointer
                 focus:outline-none focus:ring-2 focus:ring-brand-primary"
      style={{
        left: `${(x / imageWidth) * 100}%`,
        top: `${(y / imageHeight) * 100}%`,
        width: `${(width / imageWidth) * 100}%`,
        height: `${(height / imageHeight) * 100}%`,
      }}
      onClick={onClick}
      aria-label={`View ${product.name}`}
    >
      {/* Optional: Show price badge on hover */}
      <span className="absolute -bottom-6 left-1/2 -translate-x-1/2
                       bg-brand-primary text-white px-2 py-1 rounded text-xs
                       opacity-0 group-hover:opacity-100 transition-opacity">
        {formatPrice(product.price.current)}
      </span>
    </button>
  );
}
```

#### Deliverables
- [ ] Flyer detail with page viewer and bounding boxes
- [ ] Search with faceted filters
- [ ] URL-synced search state
- [ ] Mobile-optimized views

---

### 4.4 Sprint 4: Auth + Shopping Lists (Week 4-5)

#### Tasks

| Task | Priority | Estimate | Dependencies |
|------|----------|----------|--------------|
| Login page + form | P0 | 8h | Auth store |
| Register page + form | P0 | 8h | Auth store |
| Auth middleware/guards | P0 | 4h | Auth store |
| Token refresh logic | P0 | 4h | GraphQL client |
| Shopping lists page | P0 | 8h | Auth |
| List card component | P0 | 4h | UI components |
| Create list modal | P0 | 4h | Modal |
| Empty states | P1 | 4h | UI components |

#### Auth Implementation

```typescript
// stores/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: Date | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (credentials: LoginInput) => Promise<void>;
  register: (input: RegisterInput) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
      isAuthenticated: false,
      isLoading: false,

      login: async (credentials) => {
        set({ isLoading: true });
        try {
          const result = await loginMutation(credentials);
          set({
            user: result.user,
            accessToken: result.accessToken,
            refreshToken: result.refreshToken,
            expiresAt: new Date(result.expiresAt),
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: () => {
        logoutMutation();
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          expiresAt: null,
          isAuthenticated: false,
        });
      },

      refreshAuth: async () => {
        const { refreshToken } = get();
        if (!refreshToken) return;

        try {
          const result = await refreshTokenMutation();
          set({
            accessToken: result.accessToken,
            expiresAt: new Date(result.expiresAt),
          });
        } catch {
          get().logout();
        }
      },
    }),
    {
      name: 'kainuguru-auth',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: state.user,
      }),
    }
  )
);
```

#### Deliverables
- [ ] Login/register flow working
- [ ] Protected routes with middleware
- [ ] Token refresh on expiry
- [ ] Shopping lists overview page

---

### 4.5 Sprint 5: Shopping List Detail + Product Detail (Week 5-6)

#### Tasks

| Task | Priority | Estimate | Dependencies |
|------|----------|----------|--------------|
| Shopping list detail page | P0 | 8h | Lists |
| List item component | P0 | 8h | UI |
| Add item bar (free text) | P0 | 4h | Forms |
| Check/uncheck items | P0 | 4h | Mutations |
| Product detail page | P0 | 8h | Layout |
| Price display component | P0 | 4h | UI |
| Add to list from product | P0 | 4h | Modal |
| Category grouping (basic) | P1 | 4h | List detail |

#### List Item States

```typescript
// ListItem.tsx
interface ListItemProps {
  item: ShoppingListItem;
  onCheck: () => void;
  onDelete: () => void;
  onEdit: () => void;
}

export function ListItem({ item, onCheck, onDelete, onEdit }: ListItemProps) {
  const isExpired = item.linkedProduct?.isExpired;

  return (
    <div className={cn(
      "flex items-center gap-3 p-3 rounded-lg border",
      item.isChecked && "bg-gray-50 opacity-60",
      isExpired && "border-warning bg-warning/5"
    )}>
      {/* Checkbox */}
      <button
        onClick={onCheck}
        className={cn(
          "w-6 h-6 rounded-full border-2 flex items-center justify-center",
          item.isChecked ? "bg-brand-primary border-brand-primary" : "border-gray-300"
        )}
        aria-label={item.isChecked ? "Uncheck item" : "Check item"}
      >
        {item.isChecked && <CheckIcon className="w-4 h-4 text-white" />}
      </button>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <p className={cn(
          "font-medium truncate",
          item.isChecked && "line-through text-gray-500"
        )}>
          {item.description}
        </p>

        {item.linkedProduct && (
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <StoreBadge store={item.linkedProduct.store} size="sm" />
            <span>{formatPrice(item.linkedProduct.price.current)}</span>
            {isExpired && (
              <Badge variant="warning" size="sm">Pasibaigė</Badge>
            )}
          </div>
        )}
      </div>

      {/* Quantity */}
      <span className="text-sm text-gray-500">
        {item.quantity} {item.unit || 'vnt'}
      </span>

      {/* Actions */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm">
            <MoreVerticalIcon className="w-4 h-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem onClick={onEdit}>Redaguoti</DropdownMenuItem>
          <DropdownMenuItem onClick={onDelete} className="text-error">
            Ištrinti
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
```

#### Deliverables
- [ ] Shopping list detail with items
- [ ] Add/edit/delete/check items
- [ ] Product detail page with price info
- [ ] Add to shopping list from product

---

### 4.6 Phase 1 Completion Checklist

```
Pages:
[ ] Home page (hero + flyers + deals)
[ ] Search results with facets
[ ] Flyers listing with filters
[ ] Flyer detail with page viewer + bounding boxes
[ ] Product detail
[ ] Login
[ ] Register
[ ] Shopping lists overview
[ ] Shopping list detail
[ ] 404 page
[ ] Error page

Components:
[ ] Full UI component library
[ ] Responsive header/footer
[ ] Product card
[ ] Flyer card
[ ] List item
[ ] Store badge
[ ] Sale badge
[ ] Price display
[ ] Skeleton loaders
[ ] Toast notifications

Features:
[ ] GraphQL integration with codegen
[ ] Auth flow (login/register/logout)
[ ] Token refresh
[ ] Store filter (frontend state)
[ ] Search with URL sync
[ ] Pagination (cursor-based)
[ ] Mobile responsive
```

---

## 5. Phase 2: Enhanced Features

### 5.1 Sprint 6: Migration Wizard (Week 7-8)

#### Tasks

| Task | Priority | Estimate | Dependencies |
|------|----------|----------|--------------|
| Wizard overlay component | P0 | 8h | Modal |
| Expired items banner in list | P0 | 4h | List detail |
| Wizard item card | P0 | 8h | UI |
| Suggestion card with scoring | P0 | 8h | UI |
| Decision flow (replace/skip/remove) | P0 | 8h | Mutations |
| Wizard summary | P0 | 4h | UI |
| useWizard hook | P0 | 8h | GraphQL |

#### Wizard Hook Implementation

```typescript
// hooks/useWizard.ts
import { useState, useCallback } from 'react';
import { useMutation, useQuery } from 'urql';

interface UseWizardOptions {
  shoppingListId: number;
  onComplete?: (result: WizardResult) => void;
}

export function useWizard({ shoppingListId, onComplete }: UseWizardOptions) {
  const [sessionId, setSessionId] = useState<string | null>(null);

  // Query active session
  const [{ data: sessionData, fetching }] = useQuery({
    query: WIZARD_SESSION_QUERY,
    variables: { id: sessionId },
    pause: !sessionId,
  });

  const session = sessionData?.wizardSession;

  // Mutations
  const [, startWizardMutation] = useMutation(START_WIZARD);
  const [, recordDecisionMutation] = useMutation(RECORD_DECISION);
  const [, completeWizardMutation] = useMutation(COMPLETE_WIZARD);
  const [, cancelWizardMutation] = useMutation(CANCEL_WIZARD);

  // Start wizard
  const start = useCallback(async (preferredStoreIds?: number[]) => {
    const result = await startWizardMutation({
      input: { shoppingListId, preferredStoreIds },
    });
    if (result.data?.startWizard) {
      setSessionId(result.data.startWizard.id);
    }
    return result;
  }, [shoppingListId, startWizardMutation]);

  // Record decision
  const decide = useCallback(async (
    itemId: string,
    decision: DecisionType,
    suggestionId?: string
  ) => {
    if (!sessionId) return;

    const result = await recordDecisionMutation({
      input: { sessionId, itemId, decision, suggestionId },
    });
    return result;
  }, [sessionId, recordDecisionMutation]);

  // Complete wizard
  const complete = useCallback(async (applyChanges: boolean) => {
    if (!sessionId) return;

    const result = await completeWizardMutation({
      input: { sessionId, applyChanges },
    });

    if (result.data?.completeWizard) {
      onComplete?.(result.data.completeWizard);
      setSessionId(null);
    }
    return result;
  }, [sessionId, completeWizardMutation, onComplete]);

  // Cancel wizard
  const cancel = useCallback(async () => {
    if (!sessionId) return;
    await cancelWizardMutation({ sessionId });
    setSessionId(null);
  }, [sessionId, cancelWizardMutation]);

  // Derived state
  const currentItem = session?.expiredItems?.[session.currentItemIndex];
  const progress = session?.progress;
  const isComplete = session?.status === 'COMPLETED';
  const isActive = session?.status === 'ACTIVE';

  return {
    session,
    currentItem,
    progress,
    isComplete,
    isActive,
    isLoading: fetching,
    start,
    decide,
    complete,
    cancel,
  };
}
```

#### Confidence Score Display

```typescript
// lib/confidence.ts
export function getConfidenceLabel(confidence: number): {
  label: string;
  variant: 'success' | 'warning' | 'default';
} {
  if (confidence >= 0.8) {
    return { label: 'Puikus atitikimas', variant: 'success' };
  }
  if (confidence >= 0.6) {
    return { label: 'Geras atitikimas', variant: 'default' };
  }
  if (confidence >= 0.4) {
    return { label: 'Vidutinis atitikimas', variant: 'warning' };
  }
  return { label: 'Silpnas atitikimas', variant: 'warning' };
}

// SuggestionCard.tsx
function SuggestionCard({ suggestion, onSelect, isSelected }: SuggestionCardProps) {
  const { label, variant } = getConfidenceLabel(suggestion.confidence);
  const priceDiff = suggestion.priceDifference;

  return (
    <Card
      className={cn(
        "cursor-pointer transition-all",
        isSelected && "ring-2 ring-brand-primary"
      )}
      onClick={onSelect}
    >
      <div className="flex gap-4">
        {/* Product image */}
        <div className="w-20 h-20 relative">
          <Image
            src={suggestion.product.imageURL}
            alt={suggestion.product.name}
            fill
            className="object-contain"
          />
        </div>

        {/* Content */}
        <div className="flex-1">
          <div className="flex items-start justify-between">
            <div>
              <h4 className="font-medium">{suggestion.product.name}</h4>
              <StoreBadge store={suggestion.product.store} size="sm" />
            </div>
            <Badge variant={variant}>{label}</Badge>
          </div>

          {/* Price comparison */}
          <div className="mt-2 flex items-center gap-2">
            <span className="text-lg font-bold">
              {formatPrice(suggestion.product.price.current)}
            </span>
            <span className={cn(
              "text-sm",
              priceDiff < 0 ? "text-success" : priceDiff > 0 ? "text-error" : "text-gray-500"
            )}>
              {priceDiff < 0 ? '' : '+'}{formatPrice(priceDiff)}
            </span>
          </div>

          {/* Score breakdown (expandable) */}
          <Collapsible>
            <CollapsibleTrigger className="text-sm text-gray-500">
              Rodyti įvertinimą
            </CollapsibleTrigger>
            <CollapsibleContent>
              <ScoreBreakdown breakdown={suggestion.scoreBreakdown} />
            </CollapsibleContent>
          </Collapsible>
        </div>
      </div>
    </Card>
  );
}
```

---

### 5.2 Sprint 7: Price Alerts + Mobile (Week 9-10)

#### Tasks

| Task | Priority | Estimate | Dependencies |
|------|----------|----------|--------------|
| Create alert modal (from product) | P0 | 8h | Modal |
| Alerts list in account | P0 | 8h | Account page |
| Alert card component | P0 | 4h | UI |
| Activate/deactivate alerts | P0 | 4h | Mutations |
| Mobile bottom navigation | P0 | 8h | Layout |
| Mobile filter sheet | P0 | 8h | Sheet component |
| Mobile search improvements | P1 | 4h | Search |

#### Mobile Bottom Navigation

```typescript
// MobileNav.tsx
const navItems = [
  { href: '/', icon: HomeIcon, label: 'Pradžia' },
  { href: '/search', icon: SearchIcon, label: 'Paieška' },
  { href: '/flyers', icon: NewspaperIcon, label: 'Leidiniai' },
  { href: '/lists', icon: ListIcon, label: 'Sąrašai', requiresAuth: true },
  { href: '/account', icon: UserIcon, label: 'Paskyra', requiresAuth: true },
];

export function MobileNav() {
  const pathname = usePathname();
  const { isAuthenticated } = useAuthStore();

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-white border-t
                    md:hidden safe-area-inset-bottom">
      <div className="flex items-center justify-around h-16">
        {navItems.map((item) => {
          if (item.requiresAuth && !isAuthenticated) return null;

          const isActive = pathname === item.href;
          const Icon = item.icon;

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-col items-center justify-center flex-1 h-full",
                "text-gray-500 hover:text-brand-primary transition-colors",
                isActive && "text-brand-primary"
              )}
            >
              <Icon className="w-6 h-6" />
              <span className="text-xs mt-1">{item.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
```

---

## 6. Phase 3: Advanced Features

### 6.1 Sprint 8: History + Preferences (Week 11-12)

| Task | Priority | Estimate |
|------|----------|----------|
| Price history chart | P1 | 12h |
| Migration history page | P2 | 8h |
| User preferences form | P2 | 8h |
| Bulk wizard operations | P2 | 8h |

### 6.2 Sprint 9: PWA + Polish (Week 13-14)

| Task | Priority | Estimate |
|------|----------|----------|
| PWA manifest + service worker | P1 | 8h |
| Offline fallback pages | P1 | 4h |
| Push notification setup | P2 | 8h |
| Performance optimization | P1 | 8h |
| Accessibility audit fixes | P0 | 8h |

---

## 7. GraphQL Integration

### 7.1 Fragment Definitions

```typescript
// graphql/fragments/product.ts
import { graphql } from '../generated';

export const ProductListItemFragment = graphql(`
  fragment ProductListItem on Product {
    id
    name
    brand
    imageURL
    isOnSale
    validFrom
    validTo
    isExpired
    price {
      current
      original
      discount
      discountPercent
      currency
    }
    store {
      id
      code
      name
    }
  }
`);

export const ProductDetailFragment = graphql(`
  fragment ProductDetail on Product {
    ...ProductListItem
    description
    category
    subcategory
    unitSize
    unitType
    unitPrice
    boundingBox {
      x
      y
      width
      height
    }
    productMaster {
      id
      canonicalName
    }
  }
`);

// graphql/fragments/shoppingList.ts
export const ShoppingListItemRowFragment = graphql(`
  fragment ShoppingListItemRow on ShoppingListItem {
    id
    description
    quantity
    unit
    isChecked
    category
    estimatedPrice
    availabilityStatus
    linkedProduct {
      id
      name
      isExpired
      price {
        current
      }
      store {
        code
        name
      }
    }
  }
`);

// graphql/fragments/wizard.ts
export const WizardSuggestionFragment = graphql(`
  fragment WizardSuggestion on Suggestion {
    id
    score
    confidence
    priceDifference
    explanation
    matchedFields
    scoreBreakdown {
      brandScore
      storeScore
      sizeScore
      priceScore
      totalScore
    }
    product {
      id
      name
      brand
      imageURL
      price {
        current
        original
      }
      store {
        code
        name
      }
    }
  }
`);
```

### 7.2 Query Definitions

```typescript
// graphql/queries/search.ts
export const SEARCH_PRODUCTS = graphql(`
  query SearchProducts($input: SearchInput!) {
    searchProducts(input: $input) {
      products {
        product {
          ...ProductListItem
        }
        searchScore
        matchType
        highlights
      }
      totalCount
      hasMore
      queryString
      suggestions
      pagination {
        currentPage
        totalPages
        itemsPerPage
      }
      facets {
        stores {
          options {
            value
            count
            name
          }
          activeValue
        }
        categories {
          options {
            value
            count
          }
          activeValue
        }
        brands {
          options {
            value
            count
          }
          activeValue
        }
        priceRanges {
          options {
            value
            count
          }
          activeValue
        }
      }
    }
  }
`);
```

### 7.3 Codegen Configuration

```typescript
// codegen.ts
import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: 'http://localhost:8080/graphql',
  documents: ['src/**/*.{ts,tsx}'],
  generates: {
    './src/types/generated.ts': {
      plugins: [
        'typescript',
        'typescript-operations',
        'typescript-urql',
      ],
      config: {
        skipTypename: false,
        withHooks: true,
        withHOC: false,
        withComponent: false,
      },
    },
  },
  ignoreNoDocuments: true,
};

export default config;
```

---

## 8. Component Library

### 8.1 Core Components

| Component | Props | Variants | A11y |
|-----------|-------|----------|------|
| Button | onClick, disabled, loading, href | primary, secondary, ghost, danger | aria-label, disabled state |
| Card | children, onClick, variant | elevated, flat, outlined | role="article" |
| Input | value, onChange, error, label | default, search | aria-invalid, aria-describedby |
| Select | value, onChange, options | default, multi | aria-expanded, listbox |
| Modal | open, onClose, title | default, fullscreen | focus trap, aria-modal |
| Badge | children, variant | store, sale, status | aria-label |
| Skeleton | variant | text, card, image | aria-busy |
| Toast | message, type | success, error, info | role="alert" |

### 8.2 Component Documentation (Storybook)

```typescript
// Button.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta: Meta<typeof Button> = {
  title: 'UI/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'ghost', 'danger'],
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    children: 'Patvirtinti',
    variant: 'primary',
  },
};

export const Loading: Story = {
  args: {
    children: 'Siunčiama...',
    loading: true,
  },
};
```

---

## 9. State Management

### 9.1 State Categories

| Category | Tool | Persistence | Examples |
|----------|------|-------------|----------|
| Server state | React Query / urql | Cache | Products, flyers, lists |
| Auth state | Zustand + persist | localStorage | User, tokens |
| UI state | Zustand | Memory | Modals, toasts |
| Filter state | Zustand + URL | URL params | Search filters, store filter |

### 9.2 Store Filter Implementation

```typescript
// stores/storeFilterStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface StoreFilterState {
  selectedStoreIds: number[];
  setStores: (ids: number[]) => void;
  toggleStore: (id: number) => void;
  clearStores: () => void;
}

export const useStoreFilterStore = create<StoreFilterState>()(
  persist(
    (set) => ({
      selectedStoreIds: [],

      setStores: (ids) => set({ selectedStoreIds: ids }),

      toggleStore: (id) => set((state) => ({
        selectedStoreIds: state.selectedStoreIds.includes(id)
          ? state.selectedStoreIds.filter((s) => s !== id)
          : [...state.selectedStoreIds, id],
      })),

      clearStores: () => set({ selectedStoreIds: [] }),
    }),
    {
      name: 'kainuguru-store-filter',
    }
  )
);

// Hook with URL sync
export function useStoreFilter() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const store = useStoreFilterStore();

  // URL takes precedence
  const urlStores = searchParams.get('stores')?.split(',').map(Number) || [];
  const activeStores = urlStores.length > 0 ? urlStores : store.selectedStoreIds;

  const setStores = (ids: number[]) => {
    store.setStores(ids);

    // Update URL
    const params = new URLSearchParams(searchParams);
    if (ids.length > 0) {
      params.set('stores', ids.join(','));
    } else {
      params.delete('stores');
    }
    router.push(`?${params.toString()}`);
  };

  return {
    selectedStoreIds: activeStores,
    setStores,
    toggleStore: store.toggleStore,
    clearStores: store.clearStores,
  };
}
```

---

## 10. Testing Strategy

### 10.1 Test Pyramid

```
                    ┌─────────────┐
                    │    E2E      │  5-10 critical flows
                    │  Playwright │
                  ┌─┴─────────────┴─┐
                  │  Integration    │  Component + API
                  │  RTL + MSW      │
                ┌─┴─────────────────┴─┐
                │      Unit Tests     │  Hooks, utils, logic
                │        Vitest       │
                └─────────────────────┘
```

### 10.2 Critical E2E Flows

```typescript
// tests/e2e/shopping-list.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Shopping List Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto('/login');
    await page.fill('[name="email"]', 'test@example.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('can create list and add item', async ({ page }) => {
    // Navigate to lists
    await page.goto('/lists');

    // Create new list
    await page.click('text=Sukurti sąrašą');
    await page.fill('[name="name"]', 'Savaitės pirkiniai');
    await page.click('text=Sukurti');

    // Verify created
    await expect(page.locator('text=Savaitės pirkiniai')).toBeVisible();

    // Open list
    await page.click('text=Savaitės pirkiniai');

    // Add item
    await page.fill('[placeholder="Pridėti prekę..."]', 'Pienas 2L');
    await page.keyboard.press('Enter');

    // Verify item added
    await expect(page.locator('text=Pienas 2L')).toBeVisible();
  });

  test('can check off items', async ({ page }) => {
    await page.goto('/lists/1'); // Assuming list exists

    const firstItem = page.locator('[data-testid="list-item"]').first();
    const checkbox = firstItem.locator('[role="checkbox"]');

    await checkbox.click();
    await expect(checkbox).toHaveAttribute('aria-checked', 'true');
    await expect(firstItem).toHaveClass(/checked/);
  });
});
```

### 10.3 Unit Test Examples

```typescript
// lib/dates.test.ts
import { describe, it, expect } from 'vitest';
import { parseApiDate, isExpired, formatLithuanian } from './dates';

describe('dates', () => {
  describe('parseApiDate', () => {
    it('parses ISO string', () => {
      const result = parseApiDate('2025-01-15T00:00:00Z');
      expect(result).toBeInstanceOf(Date);
      expect(result.getFullYear()).toBe(2025);
    });

    it('handles Date input', () => {
      const date = new Date('2025-01-15');
      expect(parseApiDate(date)).toBe(date);
    });
  });

  describe('isExpired', () => {
    it('returns true for past dates', () => {
      expect(isExpired('2020-01-01')).toBe(true);
    });

    it('returns false for future dates', () => {
      expect(isExpired('2030-01-01')).toBe(false);
    });
  });

  describe('formatLithuanian', () => {
    it('formats date in Lithuanian locale', () => {
      const result = formatLithuanian('2025-01-15');
      expect(result).toBe('2025-01-15'); // Or localized format
    });
  });
});

// hooks/useWizard.test.ts
import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { useWizard } from './useWizard';

describe('useWizard', () => {
  it('starts wizard and sets session', async () => {
    const { result } = renderHook(() => useWizard({ shoppingListId: 1 }));

    expect(result.current.session).toBeNull();

    await act(async () => {
      await result.current.start();
    });

    expect(result.current.session).not.toBeNull();
    expect(result.current.isActive).toBe(true);
  });
});
```

---

## 11. Performance Requirements

### 11.1 Core Web Vitals Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| LCP (Largest Contentful Paint) | < 2.5s | Homepage hero load |
| FID (First Input Delay) | < 100ms | Search interaction |
| CLS (Cumulative Layout Shift) | < 0.1 | No layout jumps |
| TTFB (Time to First Byte) | < 600ms | Server response |

### 11.2 Optimization Strategies

```typescript
// Image optimization
<Image
  src={product.imageURL}
  alt={product.name}
  width={300}
  height={300}
  loading="lazy"  // Below fold
  priority={index < 4}  // Above fold
  placeholder="blur"
  blurDataURL={shimmerPlaceholder}
/>

// Code splitting
const WizardOverlay = dynamic(
  () => import('@/features/wizard/WizardOverlay'),
  { loading: () => <Skeleton variant="fullscreen" /> }
);

// Prefetching
<Link href={`/products/${product.id}`} prefetch={true}>
  View Product
</Link>

// GraphQL fragment colocation
// Only request fields needed for current view
query ProductList {
  products {
    ...ProductListItem  // Not ProductDetail
  }
}
```

---

## 12. Accessibility Requirements

### 12.1 WCAG 2.1 AA Compliance

| Requirement | Implementation |
|-------------|----------------|
| **Keyboard Navigation** | All interactive elements focusable, logical tab order |
| **Focus Visible** | Custom focus ring: `ring-2 ring-brand-primary ring-offset-2` |
| **Color Contrast** | Minimum 4.5:1 for text, 3:1 for large text |
| **Screen Reader** | ARIA labels, landmarks, live regions |
| **Reduced Motion** | Respect `prefers-reduced-motion` |
| **Form Labels** | All inputs have visible labels or aria-label |

### 12.2 Component A11y Patterns

```typescript
// Modal with focus trap
export function Modal({ open, onClose, title, children }: ModalProps) {
  const closeRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (open) {
      closeRef.current?.focus();
    }
  }, [open]);

  return (
    <Dialog.Root open={open} onOpenChange={onClose}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/50" />
        <Dialog.Content
          className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2
                     bg-white rounded-lg p-6 max-w-md w-full"
          aria-describedby={undefined}
        >
          <Dialog.Title className="text-xl font-semibold">
            {title}
          </Dialog.Title>

          {children}

          <Dialog.Close asChild>
            <button
              ref={closeRef}
              className="absolute top-4 right-4"
              aria-label="Uždaryti"
            >
              <XIcon className="w-5 h-5" />
            </button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

// Screen reader announcements
function SearchResults({ results, loading }: SearchResultsProps) {
  return (
    <>
      <div role="status" aria-live="polite" className="sr-only">
        {loading
          ? 'Ieškoma...'
          : `Rasta ${results.totalCount} rezultatų`
        }
      </div>

      <div role="region" aria-label="Paieškos rezultatai">
        {/* Results grid */}
      </div>
    </>
  );
}
```

---

## 13. Deployment Strategy

### 13.1 Environments

| Environment | URL | Purpose |
|-------------|-----|---------|
| Development | localhost:3000 | Local development |
| Staging | staging.kainuguru.lt | QA, client review |
| Production | kainuguru.lt | Live site |

### 13.2 CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v2
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
      - run: pnpm install
      - run: pnpm lint
      - run: pnpm type-check

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v2
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
      - run: pnpm install
      - run: pnpm test

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v2
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
      - run: pnpm install
      - run: pnpm exec playwright install --with-deps
      - run: pnpm e2e

  deploy-preview:
    needs: [lint, test]
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}

  deploy-production:
    needs: [lint, test, e2e]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'
```

---

## 14. Risk Mitigation

### 14.1 Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| GraphQL schema changes | High | Use codegen, version API |
| Performance on mobile | Medium | Lazy load, skeleton UI, image optimization |
| Auth token expiry | High | Proactive refresh, graceful logout |
| Wizard session timeout | Medium | Show countdown, auto-save progress |
| Offline usage | Low | PWA Phase 3, basic offline page |

### 14.2 Dependency on Backend

| Frontend Feature | Backend Dependency | Fallback |
|-----------------|-------------------|----------|
| Store filter | None (frontend state) | N/A |
| Search facets | `SearchFacets` type | Hide facets if null |
| Bounding boxes | `ProductBoundingBox` | Show products list only |
| Wizard | Full wizard API | Disable wizard CTA |
| Price history | `priceHistory` query | Hide chart section |

---

## Appendix: Quick Reference

### API Endpoints

```
GraphQL:    https://api.kainuguru.lt/graphql
Health:     https://api.kainuguru.lt/health
```

### Key GraphQL Operations

```graphql
# Public
currentFlyers, flyers, flyer
products, productsOnSale, searchProducts
stores, store
priceHistory

# Authenticated
me
shoppingLists, shoppingList, myDefaultShoppingList
createShoppingList, createShoppingListItem
checkShoppingListItem, uncheckShoppingListItem
startWizard, recordDecision, completeWizard
priceAlerts, createPriceAlert
```

### Design Tokens (Tailwind Classes)

```
Colors:     brand-primary, brand-primary-light, store-maxima, sale
Spacing:    space-xs(4), space-sm(8), space-md(16), space-lg(24)
Radius:     rounded-sm(4), rounded-md(8), rounded-lg(12)
Shadows:    shadow-sm, shadow-md, shadow-lg
```

---

*End of Implementation Plan*

**Document version:** 1.0
**Next review:** After Phase 1 completion
