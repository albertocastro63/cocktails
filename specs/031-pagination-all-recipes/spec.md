# Feature Specification: Pagination for All Recipes

**Feature Branch**: `031-pagination-all-recipes`  
**Created**: 2026-10-24  
**Status**: Draft  
**Input**: User description: "Add proper pagination to the All Recipes listing so that all cocktails in the catalog are reachable, not just the first 20. Currently the frontend requests the paginated recipes endpoint (which defaults to 20 per page) but only ever shows page 1, so recipes beyond the first 20 are silently dropped. The feature should let users browse the entire catalog via page navigation controls, showing total count and current page, and keep working with the existing search, sort, base-spirit, and ingredient filters."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse the entire catalog (Priority: P1)

A visitor opens the "All Recipes" page. The catalog contains more recipes than fit on a single page. The visitor can move forward and backward through pages of recipes until they have seen every recipe in the catalog.

**Why this priority**: This is the core defect the feature exists to fix. Today, any recipe beyond the first page is unreachable from the listing, which means the catalog silently hides content from users. Without this, the feature delivers no value.

**Independent Test**: Seed the catalog with more recipes than one page holds (e.g., 30 recipes with a page size of 20). Open the listing, confirm the first page shows the page-size number of recipes, navigate to the next page, and confirm the remaining recipes appear and that every recipe is reachable across pages.

**Acceptance Scenarios**:

1. **Given** a catalog with 30 recipes and a page size of 20, **When** the visitor opens the All Recipes page, **Then** the first 20 recipes are shown along with navigation to reach the rest.
2. **Given** the visitor is on page 1 of a multi-page catalog, **When** they choose to go to the next page, **Then** the following set of recipes is shown and the controls reflect that they are now on page 2.
3. **Given** the visitor is on the last page, **When** they view the navigation controls, **Then** the "next" action is unavailable (or disabled) because there are no further pages.
4. **Given** the visitor is on page 1, **When** they view the navigation controls, **Then** the "previous" action is unavailable (or disabled) because there is no earlier page.

---

### User Story 2 - See where I am in the catalog (Priority: P2)

While browsing, the visitor can see the total number of recipes in the catalog and which page they are currently viewing, so they understand the size of the catalog and their position within it.

**Why this priority**: Orientation makes pagination usable and trustworthy. It is strongly desirable but the catalog is still fully browsable (P1) even without an explicit position indicator.

**Independent Test**: Open the listing against a known catalog size and confirm the displayed total matches the true count and the current-page indicator updates as pages change.

**Acceptance Scenarios**:

1. **Given** a catalog of 30 recipes, **When** the visitor opens the listing, **Then** the page shows the total count of recipes in the catalog.
2. **Given** the visitor navigates between pages, **When** the page changes, **Then** the current-page indicator updates to match the page being viewed.

---

### User Story 3 - Pagination works within filtered and sorted results (Priority: P2)

The visitor applies a search term, an ingredient filter, or a base-spirit filter, or changes the sort order. Pagination continues to work: the total count and page controls reflect the filtered result set, not the whole catalog, and every matching recipe remains reachable across pages.

**Why this priority**: The existing listing already supports search, ingredient, base-spirit filtering, and sorting. Pagination must not break or bypass those capabilities, otherwise the fix would regress existing behavior. It is P2 because the core catalog-browsing fix (P1) is the primary deliverable, but filtered browsing must not regress.

**Independent Test**: Apply a filter that matches more recipes than one page holds, confirm the total reflects the number of matches, page through the matches, and confirm all matching recipes are reachable and no non-matching recipes appear.

**Acceptance Scenarios**:

1. **Given** a search or filter that matches more recipes than fit on one page, **When** the visitor views the results, **Then** the total count reflects the number of matching recipes and the controls allow reaching all of them.
2. **Given** the visitor is viewing a later page of results, **When** they change the search term or a filter, **Then** the results reset to the first page of the new result set.
3. **Given** a sort order is applied, **When** the visitor pages through results, **Then** the ordering is consistent and uninterrupted across page boundaries (no duplicates, no gaps).

---

### Edge Cases

- **Empty catalog or no matches**: When the catalog is empty or a filter matches nothing, the listing shows the existing empty-state message and no page navigation controls.
- **Single page**: When all recipes (or all matches) fit on one page, the listing shows them without requiring navigation; forward/back actions are unavailable.
- **Out-of-range page**: If a page beyond the available range is requested (e.g., via a stale link or manual navigation), the system returns an empty result with the correct total rather than an error, and the visitor can return to a valid page.
- **Catalog changes between pages**: If recipes are added or removed while browsing, the total and page boundaries may shift; the visitor still sees a valid page of results without an error.
- **Catalog larger than one retrieval batch**: The reported total and the set of reachable recipes must remain correct even when the catalog grows large enough that the underlying data store cannot return all items in a single retrieval (i.e., the total must count every stored recipe, not just the first batch).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The All Recipes listing MUST make every recipe in the catalog reachable through page navigation, not just the first page.
- **FR-002**: The listing MUST display recipes in pages of a defined, consistent page size.
- **FR-003**: Users MUST be able to navigate to the next and previous pages of results.
- **FR-004**: The system MUST disable or hide the "next" action when the current page is the last page, and the "previous" action when on the first page.
- **FR-005**: The listing MUST display the total number of recipes in the current result set (whole catalog when unfiltered, matching recipes when filtered).
- **FR-006**: The listing MUST indicate which page is currently being viewed.
- **FR-007**: Pagination MUST apply to filtered and searched results, with the total and navigation reflecting the matching result set rather than the whole catalog.
- **FR-008**: When the search term or any filter changes, the listing MUST reset to the first page of the new result set.
- **FR-009**: The reported total count MUST reflect every stored recipe in the result set, even when the catalog is too large to retrieve in a single underlying batch.
- **FR-010**: When there are no recipes in the result set, the listing MUST show the existing empty-state message and MUST NOT show page navigation controls.
- **FR-011**: Requesting a page beyond the available range MUST return an empty page with the correct total and MUST NOT produce an error.
- **FR-012**: Sort order MUST remain consistent across page boundaries so that paging produces no duplicate and no skipped recipes within a stable catalog.

### Key Entities *(include if feature involves data)*

- **Recipe**: An individual cocktail entry in the catalog. Only its presence in and position within the result set matter for this feature; its internal attributes are unchanged.
- **Result Page**: A bounded slice of the result set defined by a page number and page size, accompanied by the total count of items in the full result set.
- **Result Set**: The ordered collection of recipes being browsed — the entire catalog when unfiltered, or the matching recipes when a search/filter is applied.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With a catalog of 30 recipes, a user can reach and view all 30 recipes through the listing (previously only 20 were reachable).
- **SC-002**: The total count shown on the listing matches the actual number of recipes in the result set 100% of the time across unfiltered, searched, and filtered views.
- **SC-003**: A user can move from the first recipe to the last recipe in a 30-recipe catalog using only the on-page navigation controls, without manually editing URLs or inputs.
- **SC-004**: Applying or changing a search or filter returns the user to the first page of results in 100% of cases.
- **SC-005**: No recipe appears twice and none is skipped when paging through a stable catalog in a given sort order.

## Assumptions

- The existing recipes endpoint already supports page and page-size parameters and returns a total count; this feature primarily closes the gap between that capability and the listing UI, plus ensures the total is accurate for large catalogs.
- A default page size of 20 recipes is retained unless a reason to change it emerges during planning; the exact value is a tuning detail, not a scope driver.
- Classic page navigation (previous / next, with current page and total shown) is the intended interaction model, as stated in the request, rather than infinite scroll or a "load more" button.
- The feature targets the public "All Recipes" listing. Other listings (e.g., a user's own recipes, favorites) are out of scope unless they share the same component and benefit automatically.
- Deep-linking to a specific page via a shareable URL is desirable but not required for the initial version; it may be addressed during planning if low-cost.
- Existing search, sort, ingredient, and base-spirit filtering behavior is reused and must not regress.
