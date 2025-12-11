# UX/UI Research Analysis: Backend Reality vs. Frontend Critique

## Document Purpose
This document challenges the critique in `response.md` against the **actual backend implementation**. The goal is to inform the final frontend implementation plan with accurate data.

---

## 1. Big Picture Challenges

### 1.1 "Version 3 as Version 1" Critique

**Their Point:**
> "For an MVP, this is too much surface area... aggressively cut"

**My Challenge - Backend Reality:**

The backend **already supports** all these features. Looking at `schema.graphql`:

| Feature | Backend Status | Effort to Remove from FE |
|---------|---------------|--------------------------|
| Search + product list | ✅ Fully implemented with `searchProducts`, facets | Core - can't cut |
| Flyer browsing | ✅ `flyers`, `flyerPages`, `products` | Core - can't cut |
| Shopping lists | ✅ Full CRUD + categories + items | Core - can't cut |
| Migration wizard | ✅ Full schema in `wizard.graphql` | Cutting wastes invested backend work |
| Price alerts | ✅ Full CRUD + `myPriceAlerts` | Can defer UI, but data exists |
| User account | ✅ `me`, `User`, sessions | Core - needed for auth |

**Counter-recommendation:**

Don't cut features arbitrarily. Instead, **layer complexity**:

```
Phase 1 (MVP):
├── Search (full facets - backend supports it)
├── Flyers (list + detail with products)
├── Shopping list (single list, basic items)
└── Auth (login/register/profile)

Phase 2:
├── Multi-list support
├── Wizard v1 (simple flow, no bulk)
├── Price alerts (simple create from product)
└── Mobile optimization

Phase 3:
├── Wizard v2 (bulk accept, preferences)
├── Advanced price alerts
├── Migration history
└── Push notifications
```

**Key insight:** The backend investment is already made. The critique treats this as a greenfield project - it's not.

---

### 1.2 Home Page "Everything at Once" Critique

**Their Point:**
> "This is way too many competing CTAs... first-time users should: 1. Search for a product or 2. Open a flyer or 3. View their main list"

**My Challenge:**

This is actually **correct but incomplete**. Lithuanian grocery shopping behavior differs from general e-commerce:

**User research context (Lithuanian market):**
- Users check flyers weekly (Thursday/Friday releases)
- Price comparison is the core value prop (not just browsing)
- "Akcijos" (sales) culture is strong - users actively hunt deals

**Revised Home Page Hierarchy:**

```
Priority 1: Search (primary action)
Priority 2: This week's flyers by store (immediate value)
Priority 3: Hot deals grid (engagement hook)
---
Deprioritize for MVP:
- Price comparison spotlight → move to search/product detail
- Shopping list overview → just show icon in header with badge
- Wizard promo → only show in list detail when expired items exist
- Newsletter → footer or deferred modal
```

**Counter-recommendation:**

Home v1:
```
┌─────────────────────────────────────────────────────────────┐
│  [Logo] ─────── [ Search for products... 🔍 ] ─── [👤] [📋]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Raskite geriausias kainas Lietuvos parduotuvėse     │  │
│  │  [ Search bar - large, prominent ]                    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  Šios savaitės leidiniai                      [Visi →]     │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐              │
│  │ MAXIMA │ │  LIDL  │ │  IKI   │ │ NORFA  │ ...          │
│  └────────┘ └────────┘ └────────┘ └────────┘              │
│                                                             │
│  Geriausi pasiūlymai šiandien                 [Visi →]     │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐            │
│  │-40%  │ │-35%  │ │-50%  │ │-25%  │ │-30%  │            │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Only 3 sections: Hero/Search + Flyers + Deals. Everything else moves elsewhere.

---

### 1.3 Navigation & Store Filter Critique

**Their Point:**
> "The store filter appears in multiple places... It's not clear what the source-of-truth store selection is."

**My Challenge:**

This is **valid criticism**. Looking at the backend:

```graphql
# Backend has no user preference for "default store"
type User {
  preferredLanguage: String!  # ← only this preference exists
  # No preferredStoreID!
}
```

**The backend doesn't persist store preference**. This means:

1. Store filter must be **frontend state only** (localStorage/sessionStorage)
2. OR we need a backend change to add `preferredStoreIDs: [Int!]` to User

**Counter-recommendation:**

For MVP, implement simple frontend-only approach:

```typescript
// useStoreFilter.ts
interface StoreFilterState {
  selectedStoreIds: number[];
  source: 'header' | 'page' | 'none';
}

// Rule: Header selection is "global default"
// Page-level filters can override but don't persist
// URL params override everything: ?stores=1,2,3
```

**Do NOT add backend preference for v1** - it adds complexity. Users can re-select per session.

---

## 2. Technical Schema Challenges

### 2.1 ID Type Inconsistency

**Their Point:**
> "You use both Int! and ID! for IDs... That's a frontend trap"

**Backend Reality:**

Looking at actual schema:
```graphql
type Product { id: Int! }           # Integer
type User { id: ID! }               # String/UUID
type PriceHistory { id: ID! }       # String/UUID
type ShoppingList { id: Int! }      # Integer
type ShoppingListItem { id: Int! }  # Integer
```

This is **intentional** - User uses UUID (ID!), everything else uses Int auto-increment.

**Counter-recommendation:**

Don't "fix" this in schema - it's correct. Instead, frontend should:

```typescript
// types/ids.ts
type IntId = number;          // For Product, Flyer, ShoppingList, etc.
type UuidId = string;         // For User, PriceHistory, PriceAlert

// Helper for cache keys
function cacheKey(type: string, id: IntId | UuidId): string {
  return `${type}:${id}`;
}
```

**Accept the dual-ID pattern** - it's a reasonable backend choice.

---

### 2.2 DateTime vs String Dates

**Their Point:**
> "Dates as String guarantee you'll scatter ad-hoc parsing all over the UI"

**Backend Reality:**

```graphql
scalar DateTime  # ← Defined but...

type Product {
  validFrom: String!   # ← Used as String
  validTo: String!
  createdAt: String!
}

type WizardSession {
  startedAt: DateTime!  # ← Used correctly in wizard
  expiresAt: DateTime!
}
```

**The inconsistency is real but limited to the main schema vs wizard schema.**

**Counter-recommendation:**

Frontend should normalize this:

```typescript
// utils/dates.ts
import { parseISO, format, isAfter, isBefore } from 'date-fns';
import { lt } from 'date-fns/locale';

function parseApiDate(date: string | Date): Date {
  return typeof date === 'string' ? parseISO(date) : date;
}

function formatLithuanian(date: string | Date): string {
  return format(parseApiDate(date), 'yyyy-MM-dd', { locale: lt });
}

// All API dates go through these helpers - never raw parsing
```

This is a **frontend responsibility** - don't wait for backend schema changes.

---

### 2.3 Overloaded Types

**Their Point:**
> "Product and ShoppingListItem types are overloaded... enforce fragments per use-case"

**My Agreement + Enhancement:**

This is **correct**. The backend returns rich types; frontend must be selective.

**Required fragments for MVP:**

```graphql
# Product list (search results, flyer products, deals)
fragment ProductListItem on Product {
  id
  name
  brand
  price { current original discount discountPercent }
  imageURL
  store { id code name }
  isOnSale
  validFrom
  validTo
}

# Product detail
fragment ProductDetail on Product {
  ...ProductListItem
  description
  category
  unitSize
  unitPrice
  productMaster { id canonicalName }
  priceHistory { price recordedAt store { name } }
}

# Shopping list item (list view)
fragment ShoppingListItemRow on ShoppingListItem {
  id
  description
  quantity
  unit
  isChecked
  category
  estimatedPrice
  linkedProduct { id name price { current } store { code } }
}

# Wizard suggestion
fragment WizardSuggestion on Suggestion {
  id
  score
  confidence
  priceDifference
  product { id name price { current original } store { code name } imageURL }
  scoreBreakdown { brandScore storeScore sizeScore priceScore }
}
```

**Counter-recommendation:**

Create a `/graphql/fragments.ts` file as part of frontend setup. This is **mandatory architecture**, not optional.

---

## 3. Screen-by-Screen Re-evaluation

### 3.1 Flyer Detail - Bounding Boxes

**Their Point:**
> "v1: keep flyer detail dumb: Static image per page... Add bounding boxes as a v2 enhancement"

**Backend Reality:**

```graphql
type Product {
  boundingBox: ProductBoundingBox  # ← Already exists
  pagePosition: ProductPosition    # ← Row/column/zone
}

type ProductBoundingBox {
  x: Float!
  y: Float!
  width: Float!
  height: Float!
}
```

The bounding box data **already comes from the backend** with every product query.

**Counter-recommendation:**

Implement bounding boxes in v1, but **progressively enhance**:

```
v1.0: Static image + product list below (clickable)
v1.1: Add simple hotspot overlay (CSS positioned divs)
v1.2: Add zoom/pan for mobile (Pinch-zoom library)
```

The bounding box display is **2-3 hours of CSS work** if data already exists. Don't defer it.

```typescript
// FlyerPageViewer.tsx
function ProductHotspot({ product, imageWidth, imageHeight }) {
  const { x, y, width, height } = product.boundingBox;
  return (
    <div
      className="absolute border-2 border-green-500 hover:bg-green-500/20 cursor-pointer"
      style={{
        left: `${(x / imageWidth) * 100}%`,
        top: `${(y / imageHeight) * 100}%`,
        width: `${(width / imageWidth) * 100}%`,
        height: `${(height / imageHeight) * 100}%`,
      }}
      onClick={() => openProductDetail(product.id)}
    />
  );
}
```

---

### 3.2 Migration Wizard - Complexity

**Their Point:**
> "This screams state machine... Make the frontend wizard a dedicated state machine"

**Backend Reality (wizard.graphql):**

```graphql
type WizardSession {
  id: ID!
  status: WizardStatus!           # ACTIVE | COMPLETED | EXPIRED | CANCELLED
  currentItemIndex: Int!           # ← Backend tracks position
  progress: WizardProgress!        # ← Backend tracks stats
  expiredItems: [ExpiredItem!]!    # ← All items upfront
}

# Mutations return updated session
mutation recordDecision(...): WizardSession!
mutation completeWizard(...): WizardResult!
```

**The backend IS the state machine.** Frontend just reflects backend state.

**Counter-recommendation:**

Don't over-engineer with XState. Use simple React state:

```typescript
// useWizard.ts
function useWizard(shoppingListId: number) {
  const [session, setSession] = useState<WizardSession | null>(null);

  // Start
  const start = async () => {
    const result = await startWizardMutation({ shoppingListId });
    setSession(result);
  };

  // Record decision - backend returns updated session
  const decide = async (itemId: string, decision: DecisionType, suggestionId?: string) => {
    const result = await recordDecisionMutation({
      sessionId: session.id,
      itemId,
      decision,
      suggestionId
    });
    setSession(result.session); // Backend state is source of truth
  };

  // Derived state
  const currentItem = session?.expiredItems[session.currentItemIndex];
  const isComplete = session?.status === 'COMPLETED';

  return { session, currentItem, start, decide, isComplete };
}
```

**Key insight:** The critique assumes frontend owns wizard state. It doesn't. Backend owns it via `WizardSession`.

---

### 3.3 Price Alerts - Scope

**Their Point:**
> "v1 UI: only basic list + simple 'create alert' dialog for one product"

**My Agreement:**

This is **correct**. The backend supports:
- `createPriceAlert` (product-specific)
- `myPriceAlerts` (list all)
- `AlertType`: PRICE_DROP | TARGET_PRICE | PERCENTAGE_DROP

**v1 Implementation:**

```
Create Alert Flow:
ProductDetail → "Pranešti apie kainą" button → Modal:
  ┌─────────────────────────────────────────┐
  │  Sukurti kainos įspėjimą                │
  │                                          │
  │  Produktas: Coca-Cola 2L                │
  │                                          │
  │  Pranešti, kai:                         │
  │  ○ Kaina nukris žemiau: [___] €         │
  │  ○ Kaina nukris N%: [___] %             │
  │  ○ Bet koks kainos kritimas             │
  │                                          │
  │  [Atšaukti]      [Sukurti įspėjimą]     │
  └─────────────────────────────────────────┘

View Alerts:
Account → "Mano įspėjimai" → Simple list with delete option
```

**Defer:**
- Category-wide alerts
- Multi-store tracking
- Complex conditions

---

## 4. Valid Critiques to Accept

### 4.1 Accessibility (Accept Fully)

**Their Point:**
> "You've visually spec'd components but almost no a11y behavior"

This is **completely valid**. The FRONTEND_DESIGN_SPEC.md lacks:
- Focus states
- Keyboard navigation
- ARIA labels
- Screen reader considerations

**Required additions to design spec:**

```
Accessibility Requirements (WCAG 2.1 AA):

Focus Management:
- Visible focus ring: 2px solid var(--brand-primary)
- Focus-visible for keyboard only
- Tab order follows visual order

Keyboard Navigation:
- Carousel: Arrow keys navigate, Enter selects
- Modal: Trap focus, Escape closes
- Wizard: Tab between suggestions, Enter to accept

Screen Reader:
- All images: alt text from product name
- Sale badges: aria-label="Nuolaida 30 procentų"
- Store logos: aria-label="Maxima parduotuvė"
- Progress: aria-live="polite" for wizard progress
```

---

### 4.2 Design Tokens (Accept Fully)

**Their Point:**
> "You haven't explicitly mapped them to code-level tokens"

**Add to design spec:**

```typescript
// design-tokens.ts
export const tokens = {
  colors: {
    brand: {
      primary: '#2E7D32',
      primaryLight: '#4CAF50',
      primaryDark: '#1B5E20',
      secondary: '#FF9800',
    },
    store: {
      maxima: '#E31E24',
      lidl: '#00529B',
      iki: '#E30613',
      norfa: '#00A651',
      rimi: '#C8102E',
    },
    neutral: {
      50: '#FAFAFA',
      100: '#F5F5F5',
      // ...
      900: '#212121',
    },
    semantic: {
      success: '#4CAF50',
      warning: '#FF9800',
      error: '#F44336',
      info: '#2196F3',
    },
  },
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
    xxl: '48px',
  },
  typography: {
    fontFamily: {
      sans: '"Inter", -apple-system, sans-serif',
      mono: '"JetBrains Mono", monospace',
    },
    fontSize: {
      xs: '12px',
      sm: '14px',
      base: '16px',
      lg: '18px',
      xl: '20px',
      '2xl': '24px',
      '3xl': '30px',
      '4xl': '36px',
    },
  },
  breakpoints: {
    sm: '640px',
    md: '768px',
    lg: '1024px',
    xl: '1280px',
  },
};
```

---

## 5. Revised MVP Scope Recommendation

### Core MVP (4-6 weeks frontend)

```
├── /                    Home (search + flyers + deals)
├── /search              Search results with facets
├── /flyers              Flyer listing
├── /flyers/:id          Flyer detail with products + bounding boxes
├── /products/:id        Product detail (price, store, basic info)
├── /login               Auth
├── /register            Auth
├── /lists               Shopping lists overview (auth required)
├── /lists/:id           List detail with items (auth required)
└── /account             Basic profile (auth required)
```

### Phase 2 (2-3 weeks)

```
├── Wizard integration in /lists/:id
├── /alerts              Price alerts list
├── Create alert from product detail
└── Mobile bottom navigation
```

### Phase 3 (2-3 weeks)

```
├── Bulk wizard operations
├── Migration history
├── User preferences
├── PWA / push notifications
└── Advanced price history charts
```

---

## 6. Frontend Architecture Agreement

**Accept the proposed structure with modifications:**

```
/src
  /features
    /home
      HomePage.tsx
      components/
        FlyerCarousel.tsx
        DealsGrid.tsx
    /flyers
      FlyersPage.tsx
      FlyerDetailPage.tsx
      components/
        FlyerCard.tsx
        FlyerPageViewer.tsx       # With bounding box support
        ProductHotspot.tsx
    /search
      SearchPage.tsx
      hooks/
        useProductSearch.ts       # All search logic here
      components/
        SearchFilters.tsx
        ProductGrid.tsx
        Pagination.tsx
    /shoppingList
      ShoppingListsPage.tsx
      ShoppingListDetailPage.tsx
      hooks/
        useShoppingList.ts
        useWizard.ts              # Wizard as hook, not full feature
      components/
        ListCard.tsx
        ListItem.tsx
        WizardOverlay.tsx         # Simple overlay, not separate route
    /auth
      LoginPage.tsx
      RegisterPage.tsx
      hooks/
        useAuth.ts
    /account
      AccountPage.tsx
      components/
        ProfileForm.tsx
        AlertsList.tsx            # Simple list in account

  /components
    /ui
      Button.tsx
      Card.tsx
      Input.tsx
      Modal.tsx
      Badge.tsx
      Skeleton.tsx
    /layout
      Header.tsx
      Footer.tsx
      PageShell.tsx
      MobileNav.tsx

  /graphql
    /fragments
      product.ts
      shoppingList.ts
      wizard.ts
    /queries
      home.ts
      flyers.ts
      search.ts
      shoppingLists.ts
    /mutations
      auth.ts
      shoppingLists.ts
      wizard.ts

  /hooks
    useStoreFilter.ts             # Global store selection
    useAuth.ts                    # Auth context

  /lib
    apollo-client.ts              # or urql/tanstack
    dates.ts                      # Date utilities
    tokens.ts                     # Design tokens

  /types
    generated.ts                  # GraphQL codegen
    ids.ts                        # ID type helpers
```

---

## 7. Summary: What to Accept vs. Challenge

| Critique | Decision | Reason |
|----------|----------|--------|
| "Cut MVP scope" | **Partially accept** | Layer features, don't cut arbitrarily |
| "Home page too busy" | **Accept** | Reduce to search + flyers + deals |
| "Store filter unclear" | **Accept** | Define single source of truth (frontend state) |
| "ID type inconsistency" | **Challenge** | It's intentional (UUID vs Int) |
| "DateTime as String" | **Accept** | Normalize in frontend utils |
| "Overloaded types" | **Accept** | Define fragments per use-case |
| "Defer bounding boxes" | **Challenge** | Data exists, 2-3 hours work |
| "Wizard needs state machine" | **Challenge** | Backend IS state machine |
| "Price alerts minimal" | **Accept** | Simple create + list only |
| "Missing accessibility" | **Accept** | Add to design spec |
| "Missing design tokens" | **Accept** | Add code-level mapping |
| "Kill subscriptions v1" | **Accept** | Not needed for single-user browser |

---

## 8. Next Steps for UX/UI Research Phase

1. **User flow validation** - Map actual Lithuanian grocery user journeys
2. **Competitor analysis** - How do Barbora, Pigu, existing flyer apps work?
3. **Mobile-first review** - 70%+ traffic likely mobile
4. **Accessibility audit** - Add WCAG requirements to spec
5. **Design token finalization** - Create Tailwind/CSS-in-JS config
6. **Prototype key flows** - Wizard, search, list management

---

*Document version: 1.0*
*Based on: FRONTEND_DESIGN_SPEC.md + response.md + actual backend schemas*
*Date: 2025-12-10*
