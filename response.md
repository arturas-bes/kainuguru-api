Big‑picture: what’s wrong with your approach

1.1 You’ve designed “version 3” as your version 1

Your spec covers:
•	Full design system
•	10+ page types (home, flyers, flyer detail, product detail, search, lists, list detail, wizard, alerts, account, auth, etc.)
•	Migration wizard with preferences, bulk actions, history, stats, subscriptions
•	Price alerts with advanced filters
•	Detailed empty, error, and loading states for almost everything

For an MVP, this is too much surface area:
•	More screens → more state → more bugs → slower delivery.
•	Each feature touches multiple layers: UI → GraphQL queries → server logic.
•	You are locking the frontend into a large mental model before you’ve validated real user behavior.

Blunt recommendation:
For v1, aggressively cut. Focus on:
1.	Search + product list
2.	Basic flyer browsing
3.	Single simple shopping list with “add item” and “add from product/flyer”
4.	Very soft version of wizard (even manual “suggest alternatives” dialog)

Everything else (price alerts, migration history, wizard preferences, migration statistics, advanced account settings) is nice-to-have, not core for day one.

⸻

1.2 Your home page tries to be everything at once

Your home page currently wants to show:
•	Hero with search & value prop
•	Current flyers section
•	Hot deals carousel
•	Price comparison spotlight
•	Shopping lists overview
•	Price alerts summary
•	Migration wizard promo
•	Newsletter
•	Footer

This is way too many competing CTAs.

From a behavior standpoint, first-time users should:
1.	Search for a product or
2.	Open a flyer or
3.	View their main list

All the other modules just add noise.

Recommendation:
•	For home v1, pick two core sections under the hero:
•	“Current Flyers” (by store)
•	“Popular Deals / Recommendations” (a simple productsOnSale grid)
•	Push:
•	Price alerts into the user account area only.
•	Wizard promo into shopping list detail only.
•	Shopping list overview into either a small card or even just a simple link: “Peržiūrėti pirkinių sąrašą”.

⸻

1.3 Navigation & mental model is heavy

You have:
•	Header: store dropdown, global search, cart/lists, profile
•	Per-page filters (stores, categories, etc.)
•	Separate pages for flyers, products, lists, alerts, wizard, etc.

This is fine long term, but for implementation:
•	The store filter appears in multiple places (header dropdown + per-page tabs/filters).
•	It’s not clear what the source-of-truth store selection is.

Recommendation:
•	Decide one global rule:
“Global store selection lives in header and is persisted in app state. Feature pages (flyers, search, lists) default to it but can override locally.”
•	Implement it as a single piece of state on the frontend:
•	e.g. useStoreFilter() hook backed by GraphQL user preferences if logged in.
•	Don’t re-invent store filters in every page; reuse the same controlled component.

2.2 Inconsistent types will bite your frontend

In schema:
•	You use both Int! and ID! for IDs (e.g. userID: ID! vs id: Int!, shoppingListID: Int!).
•	You define scalar DateTime but many fields are String! dates like validFrom, validTo, createdAt, etc.

That’s a frontend trap:
•	ID type mismatches are annoying in TypeScript and cache keys (Apollo/urql/TanStack Query).
•	Dates as String guarantee you’ll scatter ad-hoc parsing all over the UI.

Concrete improvements:
•	Make all IDs ID! unless there’s a very strong reason.
•	Use DateTime for:
•	validFrom, validTo, saleStartDate, saleEndDate, createdAt, updatedAt, expiryDate, checkedAt, etc.
•	If the backend can’t handle real DateTime yet, still define it as DateTime and internally map to ISO strings; you’ll thank yourself on the frontend.

2.3 Your Product and ShoppingList types are overloaded

Product contains:
•	Full pricing breakdown
•	Flyer + page + position stuff
•	Store context
•	Computed fields (isCurrentlyOnSale, isValid, isExpired)

ShoppingListItem contains:
•	Core list fields (description, quantity)
•	Full linkage to Product, Store, Flyer, ProductMaster
•	Suggestion/matching metadata (suggestionSource, matchingConfidence, availabilityStatus)

As a result, every query returning these types can easily overfetch.

Better approach:
•	You don’t need a separate GraphQL type, but you should enforce fragments per use-case on the frontend:
•	ProductListItemFragment
•	ProductDetailFragment
•	ShoppingListItemRowFragment
•	WizardSuggestionFragment

This is a frontend responsibility: define fragments and use them across queries so you don’t accidentally use the heavy “detail” selection on list pages.

If you want to be stricter, you can create lighter types like SearchProduct, but fragments are usually enough.

⸻

3. Screen‑by‑screen critique & upgrades

3.1 Home page

Issues:
•	Too many sliders/sections = visual noise.
•	The “price comparison spotlight” is visually cool but may steal focus from search.
•	Migration wizard promo on the home page is premature. That’s a post‑adoption feature.

Recommendations:
•	Above-the-fold: Search bar + very simple 2–3 shortcuts:
•	“Peržiūrėti akcijų leidinius”
•	“Atidaryti mano pirkinių sąrašą”
•	Below:
•	Section 1: “Šios savaitės leidiniai” (cards for current flyers)
•	Section 2: “Geriausi pasiūlymai šiandien” (simple grid of deals)
•	Move wizard & alerts deeper into the app (shopping list & account).

3.2 Flyers (list + detail)

You have:
•	Flyers list page with filters, sort, store tabs.
•	Flyer detail page with:
•	Page thumbnails, big image viewer, bounding boxes for products
•	Extracted products list with filters.

Risks:
•	Implementing a custom page viewer with bounding boxes, zooming, scroll, and click → non-trivial.
•	If you do heavy interaction here, you’ll burn a lot of time before users even care.

Recommendations:
•	v1: keep flyer detail dumb:
•	Static image per page.
•	Simple “products extracted from this flyer” list under/next to it.
•	Clicking a product row scrolls to the product card; clicking card opens product detail.
•	Add bounding boxes as a v2 enhancement only after base flows are stable.

3.3 Search results

You’ve designed a classic faceted search page: facets left, results in center.

Checklist for improvement:
•	Make search results strongly consistent with searchProducts input:
•	Each facet in UI must map to some GraphQL filters field.
•	Avoid full-page reload on filter changes:
•	Use a debounced searchProducts query with query state mirrored to URL (?q=milk&store=IKI&discounted=true).
•	Add a clear “Reset filters” pattern; this always gets missed.

From frontend architecture: treat search as a single feature with:
•	useProductSearch hook handling:
•	Query
•	Pagination
•	Facets
•	URL sync

Do not sprinkle GraphQL calls directly in the page component.

3.4 Shopping lists (list + detail)

This is one of your main value props, but you’re mixing “fancy” with “v1”.

Spec includes:
•	Overview page with stats, list cards, “default list”, etc.
•	Detail page with:
•	Category groups
•	Expired badges
•	Per item links to products/flyers
•	Summary price stats
•	Wizard button
•	Share code, etc.

Issues:
•	You’re going to spend a lot of time on list management details instead of price comparison.
•	Category handling is unclear (static categories vs user-defined?).

Recommendations:

v1 detail page:
•	Flat list with optional lightweight grouping (e.g. category label above an item).
•	Simple capabilities:
•	Add item (free text)
•	“Link to product offer” (if available)
•	Checkbox to mark done
•	Show “approximate total price” using estimatedTotalPrice
•	Wizard: just a single CTA above the list:
"Rasta X prekių su pasibaigusiomis akcijomis – Atnaujinti"
Don’t embed wizard logic into every row; keep it in the wizard feature.

Also, make sure your UI design matches the schema:
•	Use ShoppingList.hasActiveWizardSession, expiredItemCount for banners/cards, not extra queries when the list loads.
•	Items: rely on ShoppingListItem and its linkedProduct/productMaster instead of re-querying product for each row.

3.5 Migration Wizard

This is the most complex frontend flow. You’ve:
•	Made a multi-step wizard flow spec.
•	Designed a rich GraphQL model:
•	WizardSession, ExpiredItem, Suggestion, MigrationItemDecision, stats, preferences, history.
•	Queries: activeWizardSession, wizardSession, getItemSuggestions, migrationHistory, etc.
•	Mutations: startWizard, recordDecision, bulkAcceptSuggestions, completeWizard, cancelWizard, updateMigrationPreferences.
•	Subscriptions: wizardSessionUpdates, expiredItemNotifications.

From a frontend perspective, this screams state machine, but your spec treats it like a linear form.

Concrete improvements:
1.	Make the frontend wizard a dedicated state machine
Something like:
•	States: idle → loadingSession → reviewingItem → applyingBulk → summary → error
•	Events: START, NEXT_ITEM, PREV_ITEM, ACCEPT, SKIP, REMOVE, COMPLETE, CANCEL.
Whether you implement this with XState or plain reducer/hooks is your call, but you should design it as a state machine.
2.	Minimize round trips:
•	startWizard should return:
•	WizardSession + ExpiredItem[] (without suggestions or with top-N suggestions if cheap).
•	For each item, either:
•	Preload suggestions via ExpiredItem.suggestions, or
•	Lazy-load them via getItemSuggestions as the user hits “Peržiūrėti alternatyvas”.
•	recordDecision should return the updated WizardSession including new currentItemIndex and updated item status.
3.	Kill subscriptions in v1:
•	wizardSessionUpdates and expiredItemNotifications are nice, but not necessary for a single-user flow in the browser.
•	Implement wizard purely via queries + mutations.
•	Add subscriptions only if multi-device or background detection becomes real requirements.
4.	UI vs schema alignment:
•	Your UI wants:
•	Old vs new price
•	Store logos
•	Confidence scores/labels (e.g., “Rekomenduojama”, “Vidutinis atitikimas”)
•	You have these fields already:
•	Suggestion.score, Suggestion.confidence, priceDifference, ScoreBreakdown, matchedFields.
•	Frontend should:
•	Implement a mapping:
e.g. confidence > 0.8 → "Labai geras atitikimas"; 0.6–0.8 → “Vidutinis”, etc.
•	Create a single WizardSuggestionCard component consuming this.

If you don’t treat wizard as its own, well-encapsulated feature, it will leak weird edge cases across your app.

⸻

3.6 Price alerts

You have full flows: list, create, edit, activation, etc. Schema is also rich (PriceAlert, CreatePriceAlertInput, UpdatePriceAlertInput, filters).

V1 reality: alerts are add-on, not core.

Recommendation:
•	v1 UI: only basic list + simple “create alert” dialog for one product:
•	Trigger from product detail: “Pranešti, kai kaina kris žemiau X €”
•	Save priceAlert tied to product.
•	Hide advanced things like:
•	Category-wide alerts
•	Complex conditions
until you see usage justifying them.

From frontend side, this avoids building an entire alerts management console that nobody uses on day one.

⸻

4. Design system & components

4.1 Design system is decent, but you’re missing implementation thinking

You did:
•	Colors, typography, spacing, breakpoints.
•	Buttons, cards, inputs.
•	Store brand colors.

But you haven’t explicitly mapped them to code-level tokens and components.

Concrete improvement:
•	Define a design tokens layer:

// tokens.ts
export const colors = {
brand: {
primary: '#2E7D32',
primaryLight: '#4CAF50',
primaryDark: '#1B5E20',
// ...
},
store: {
maxima: '#E31E24',
lidl: '#00529B',
// ...
},
} as const;

Then build foundational components:
•	<Button variant="primary" size="md" />
•	<Card variant="elevated|flat">
•	<Badge variant="store" store="MAXIMA">

The goal: the pixel-perfect spec must map to a small set of reusable components, not one-off CSS per page.

4.2 Accessibility is under-specified

You’ve visually spec’d components but almost no a11y behavior:
•	Focus states?
•	Keyboard navigation (for carousels, wizard, modals)?
•	Screen reader labels for icons (e.g., sale badges, store icons)?

You’ll regret ignoring this later.

Bare minimum for v1:
•	Define focus styles for:
•	Buttons, links, chips, facet filters, search input, pagination, wizard controls.
•	Use semantic HTML:
•	<nav> for navigation
•	<main> for content
•	<section> with ARIA labels for major sections
•	Ensure modals/wizard:
•	Trap focus
•	Return focus when closed
•	Are closable via ESC

5. Concrete frontend structure I’d recommend

Assuming React/TypeScript + GraphQL client (Apollo/urql/TanStack Query with custom fetch), I’d structure like this:

/src
/features
/home
HomePage.tsx
useHomeSummary.ts
/flyers
FlyersPage.tsx
FlyerDetailPage.tsx
components/
/search
SearchPage.tsx
useProductSearch.ts
/shoppingList
ShoppingListPage.tsx
ShoppingListDetailPage.tsx
useShoppingList.ts
/wizard
WizardOverlay.tsx
wizardMachine.ts (or useWizard.ts)
api.ts (startWizard, recordDecision, etc.)
/priceAlerts
PriceAlertsPage.tsx (v2)
/auth
LoginPage.tsx
RegisterPage.tsx
/components
/ui (button, card, input, badge, modal, etc.)
/layout (Header, Footer, PageShell)

Key frontend rules:
•	Every feature has:
•	Its own hooks that wrap GraphQL.
•	Local state and UI glued together.
•	Screens use feature hooks, not raw GraphQL clients directly.
•	Shared state like selectedStore and currentUser sit in a core context or separate feature.

6. What I’d do next in your place

If you want this to actually ship instead of becoming a beautiful but dead spec:
1.	Cut the scope:
•	Lock in a brutal MVP feature list.
2.	Add screen-level queries:
•	homeSummary
•	shoppingListDetail (list + items + wizard flags)
3.	Normalize IDs and dates in schema.
4.	Define frontend feature modules and one shared design system layer.
5.	Treat wizard as its own feature with a state machine, and drop subscriptions for now.

You’re clearly thinking in a very structured way, which is good. But right now you’re trying to boil the ocean. Trim this to something buildable, make the frontend architecture boring and predictable, and then you can layer on all the fancy wizard/alerts/history stuff later without everything collapsing under its own weight.