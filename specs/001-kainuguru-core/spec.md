# Feature Specification: Kainuguru Grocery Flyer Aggregation System

**Feature Branch**: `001-kainuguru-core`
**Created**: 2025-11-04
**Status**: Draft
**Input**: User description: "Build a Lithuanian grocery flyer aggregation system 'Kainuguru'"

## Clarifications

### Session 2025-11-04

- Q: When new weekly flyers become available, what should happen to the previous week's data? → A: Archive previous flyers for price history (keep indefinitely) but remove images, leave minimal necessary data
- Q: Should unregistered users be able to browse flyers and search products, or is registration required for all features? → A: Anonymous browsing/search allowed, registration for lists only
- Q: When a product in a shopping list is no longer available in any current flyers, how should the system handle it? → A: Keep as text with "no longer available" indicator but aim to match similar products by tags since exact products change weekly due to naming discrepancies
- Q: When automated product extraction from a flyer page fails, what should be the immediate system response? → A: Display page as-is with manual review flag
- Q: Which additional stores should be prioritized for future expansion beyond IKI, Maxima, and Rimi? → A: Add all stores immediately

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse Weekly Grocery Flyers (Priority: P1)

As a Lithuanian shopper (registered or anonymous), I want to browse current weekly flyers from all major grocery stores in one place so I can quickly see what's on sale without visiting multiple websites or collecting paper flyers.

**Why this priority**: This is the core value proposition - aggregating flyer information saves users significant time and enables price comparison. Without this, the system has no value.

**Independent Test**: Can be fully tested by accessing the system and verifying that current week's flyers from at least one store are viewable with product information visible.

**Acceptance Scenarios**:

1. **Given** I am a user on the platform, **When** I access the flyers section, **Then** I see a list of current weekly flyers from Lithuanian grocery stores
2. **Given** current weekly flyers are available, **When** I select a specific store's flyer, **Then** I can view all products and prices from that flyer
3. **Given** I am viewing a flyer, **When** I browse through pages, **Then** I can see product names, prices, and any promotional information in Lithuanian

---

### User Story 2 - Search for Products Across All Flyers (Priority: P1)

As a Lithuanian shopper (registered or anonymous), I want to search for specific products across all current flyers so I can find the best price for items I need to buy.

**Why this priority**: Search functionality is essential for users to find specific products quickly without manually browsing through multiple flyers. This directly helps users save money.

**Independent Test**: Can be tested by searching for common grocery items and verifying that relevant results appear from available flyers with accurate pricing.

**Acceptance Scenarios**:

1. **Given** multiple flyers with products are available, **When** I search for a product name in Lithuanian, **Then** I see all matching products from all stores with their prices
2. **Given** products have varying Lithuanian spellings/names, **When** I search with partial or approximate text, **Then** I still find relevant products
3. **Given** I search for a product, **When** results are displayed, **Then** I can see which store has the best price

---

### User Story 3 - Create and Manage Shopping Lists (Priority: P2)

As a Lithuanian shopper, I want to create shopping lists that persist across weekly flyer updates so I can track items I need to buy even when flyers change.

**Why this priority**: Shopping lists provide continuity for users and enhance the platform's utility beyond just browsing deals. This creates user stickiness and repeat usage.

**Independent Test**: Can be tested by creating a shopping list, adding items, and verifying the list persists when returning to the system later, even after flyer updates.

**Acceptance Scenarios**:

1. **Given** I am a registered user, **When** I create a shopping list and add items, **Then** my list is saved and accessible when I return
2. **Given** I have a shopping list with items, **When** weekly flyers update, **Then** my list items remain accessible
3. **Given** I have items in my shopping list, **When** I view my list, **Then** I can see current prices from available flyers for those items

---

### User Story 4 - Track Price History for Products (Priority: P3)

As a Lithuanian shopper, I want to see historical price trends for products so I can determine if current prices are genuinely good deals.

**Why this priority**: Price history adds analytical value but is not essential for MVP. Users can still save money without this feature by comparing current prices.

**Independent Test**: Can be tested by viewing a product and checking if historical price data is displayed for previous weeks.

**Acceptance Scenarios**:

1. **Given** a product has been in multiple weekly flyers, **When** I view the product details, **Then** I can see its price history over time
2. **Given** I am viewing price history, **When** I look at the data, **Then** I can identify price trends and lowest historical prices

---

### User Story 5 - User Registration and Authentication (Priority: P2)

As a user, I want to create an account and securely log in so I can save my shopping lists and preferences for future visits.

**Why this priority**: User accounts enable personalization and data persistence, crucial for shopping list functionality and building user loyalty.

**Independent Test**: Can be tested by registering a new account, logging out, and logging back in to verify access to saved data.

**Acceptance Scenarios**:

1. **Given** I am a new user, **When** I register with my email, **Then** I can create an account and receive confirmation
2. **Given** I have an account, **When** I log in with my credentials, **Then** I can access my saved shopping lists and preferences
3. **Given** I forgot my password, **When** I request a reset, **Then** I receive instructions to create a new password

---

### Edge Cases

- What happens when a store's flyer is temporarily unavailable? System displays previous available flyer with outdated warning
- How does the system handle products with multiple size/quantity variants?
- What occurs when product names change between weekly flyer rotations? Product Master uses tags to maintain continuity
- How are promotional bundles (buy 2 get 1 free) represented?
- What happens when searching in mixed Lithuanian/English text?
- When extraction fails for a page, display page image with manual review flag and queue for admin attention

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST collect and display current weekly flyers from all major Lithuanian grocery stores (IKI, Maxima, Rimi, Lidl, Norfa, and other chains)
- **FR-002**: System MUST extract individual product information including names, prices, and quantities from flyers
- **FR-003**: System MUST provide full-text search functionality supporting Lithuanian language with diacritics (ą, č, ę, ė, į, š, ų, ū, ž)
- **FR-004**: System MUST support fuzzy/approximate search to handle spelling variations and typos
- **FR-005**: Users MUST be able to create persistent shopping lists that survive weekly data updates
- **FR-006**: System MUST automatically update flyer data weekly as stores publish new promotional materials, archiving previous flyers with minimal data (no images) for price history
- **FR-007**: System MUST handle product matching across different flyers despite naming variations
- **FR-008**: System MUST support user registration and secure authentication, with anonymous access allowed for browsing and searching
- **FR-009**: System MUST display prices and product information in Lithuanian
- **FR-010**: System MUST provide product search results within 500 milliseconds for responsive user experience
- **FR-011**: System MUST continue functioning when individual flyer extractions fail, displaying pages as-is with manual review flags
- **FR-012**: System MUST support at least 100 concurrent users during peak shopping planning times
- **FR-013**: Shopping lists MUST intelligently adapt when products are no longer available in new flyers, keeping items with unavailable status and suggesting similar products by tag matching

### Key Entities *(include if feature involves data)*

- **Store**: Represents a grocery store chain (name, location information, update schedule)
- **Flyer**: Weekly promotional material from a store (valid dates, page count, store reference, archival status indicating if images removed)
- **Product**: Individual item from a flyer (name, price, quantity/size, promotional details)
- **Shopping List**: User-created collection of desired products (name, owner, items)
- **Shopping List Item**: Individual entry in a shopping list (product reference or text description, quantity needed, availability status, suggested similar products by tags)
- **User**: Registered platform user (authentication credentials, preferences, owned lists)
- **Product Master**: Normalized product identity across different flyer iterations (uses tags for matching similar products when exact names vary)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can view current weekly flyers from all major Lithuanian grocery stores within 3 clicks of accessing the platform
- **SC-002**: Product search returns relevant results in under 500 milliseconds for 95% of queries
- **SC-003**: Shopping lists maintain 90% or higher item recognition rate when transitioning between weekly flyer updates
- **SC-004**: System successfully extracts and displays at least 80% of products from collected flyers
- **SC-005**: Users can complete a product price comparison across all available stores in under 30 seconds
- **SC-006**: Platform maintains 99% uptime during standard shopping hours (8 AM - 10 PM Lithuanian time)
- **SC-007**: 90% of users can successfully create and retrieve a shopping list on their first attempt
- **SC-008**: System processes and publishes new weekly flyers within 4 hours of availability
- **SC-009**: Search functionality correctly handles Lithuanian diacritics in 100% of cases
- **SC-010**: Platform supports at least 10,000 registered users with acceptable performance

## Assumptions

- Grocery stores publish flyers on a weekly basis with predictable timing
- Users primarily shop at major chain stores rather than small local shops
- Product information in flyers is primarily in Lithuanian language
- Users have internet access and basic digital literacy to use web platforms
- Store flyer formats remain relatively consistent week-to-week
- Users value price comparison and deal-finding over other shopping factors