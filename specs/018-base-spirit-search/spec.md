# Feature Specification: Base Spirit Search Filter

**Feature Branch**: `018-base-spirit-search`
**Created**: 2026-05-25
**Status**: Draft
**Input**: User description: "Add search by base spirit by writing in the search field 'base spirit is ...' or 'base spirit = ...'. If that is included in the search filter by that condition, for example you can say 'search for a cocktail that uses absynthe where the base spirit is rye whiskey'. Finally, allow alternative spellings for whiskey or whisky."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Search by Base Spirit (Priority: P1)

A user browsing the recipe list can refine results by naming the base spirit directly in the search field using the syntax `base spirit is <name>` or `base spirit = <name>`. The filter returns only recipes whose flagged base-spirit ingredient matches the given name. This filter can be combined with a regular ingredient search so the user can, for example, find every absinthe cocktail whose base spirit is rye whiskey.

**Why this priority**: Users with specific spirit preferences (e.g. tequila-drinkers, rye fans) currently have no way to filter by base spirit without knowing all recipe names. This makes recipe discovery faster and more intentional.

**Independent Test**: Type `base spirit is gin` in the search field on the recipe list page. Confirm only recipes that have gin marked as their base spirit are returned. Type `base spirit = Gin` (different capitalisation) and confirm the same results appear.

**Acceptance Scenarios**:

1. **Given** a user types `base spirit is rye whiskey` in the search field, **When** results load, **Then** only recipes whose base spirit is rye whiskey (or rye whisky) are shown.
2. **Given** a user types `base spirit = bourbon`, **When** results load, **Then** only recipes whose base spirit is bourbon are shown; the equals-sign syntax is treated identically to the `is` syntax.
3. **Given** a user types `absinthe base spirit is rye whiskey`, **When** results load, **Then** only recipes containing absinthe as an ingredient AND having rye whiskey as their base spirit are returned (combined filter).
4. **Given** a user types `base spirit is rye whisky` (alternative spelling), **When** results load, **Then** the same results are returned as `base spirit is rye whiskey` (both spellings match the same ingredient).
5. **Given** a user types `base spirit is ` (empty value), **When** results load, **Then** the filter is ignored and normal search behaviour applies.
6. **Given** no recipes have the specified base spirit, **When** results load, **Then** the empty-state message is shown, consistent with any other search that yields no results.

---

### User Story 2 — Whiskey / Whisky Spelling Normalisation (Priority: P1)

The application treats `whiskey` and `whisky` as equivalent in all ingredient searches, not just base-spirit searches. A user searching for `rye whisky` finds the same recipes as `rye whiskey`, regardless of how the recipe creator spelled the ingredient when adding it.

**Why this priority**: Spirit names frequently have variant spellings (Irish/American "whiskey" vs. Scotch/Japanese "whisky"). Forcing users to know which spelling a recipe creator used creates a poor search experience.

**Independent Test**: Add a recipe with an ingredient named "rye whisky". Search for "rye whiskey". Confirm the recipe appears. Search for "rye whisky". Confirm the same recipe appears.

**Acceptance Scenarios**:

1. **Given** a recipe was created with ingredient "scotch whisky", **When** a user searches "scotch whiskey", **Then** that recipe appears in results.
2. **Given** a recipe was created with ingredient "bourbon whiskey", **When** a user searches "bourbon whisky", **Then** that recipe appears in results.
3. **Given** a user searches "whisky sour", **When** results load, **Then** recipes with "whiskey sour" in their name or ingredients are also returned.
4. **Given** spelling normalisation is active, **When** it is applied, **Then** it does not affect non-whiskey/whisky terms — other ingredient names are unaffected.

---

### Edge Cases

- **Base spirit filter alone**: `base spirit is vodka` with no other terms is valid and returns all recipes whose base spirit is vodka.
- **Multiple `base spirit is` clauses**: Only the first occurrence is used; additional occurrences are silently ignored.
- **Case insensitivity**: The `base spirit is` / `base spirit =` keywords and the spirit name are matched case-insensitively.
- **Partial spirit name**: `base spirit is rye` returns recipes whose base spirit name contains "rye" (e.g. "rye whiskey", "rye whisky") — substring matching consistent with existing ingredient search.
- **No base spirit flagged**: Recipes with no ingredient flagged as base spirit are excluded from base-spirit-filtered results.
- **Whiskey/whisky in base spirit filter**: `base spirit is rye whiskey` and `base spirit is rye whisky` return identical results.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The recipe search field MUST recognise the syntax `base spirit is <value>` and `base spirit = <value>` as a base-spirit filter directive (case-insensitive).
- **FR-002**: When a base-spirit filter is present, results MUST include only recipes that have at least one ingredient flagged as the base spirit whose name matches the filter value.
- **FR-003**: The base-spirit filter MUST be combinable with regular ingredient search terms; the non-filter portion of the query is applied as a normal ingredient search alongside the base-spirit constraint.
- **FR-004**: The spirit name value in a base-spirit filter MUST use substring matching, consistent with the existing ingredient search behaviour.
- **FR-005**: The application MUST treat `whiskey` and `whisky` as equivalent in all ingredient searches — both the regular ingredient search and the base-spirit filter.
- **FR-006**: An empty or whitespace-only base-spirit value MUST be silently ignored; the remainder of the query is processed normally.
- **FR-007**: The search field hint text MUST be updated to communicate that `base spirit is <name>` syntax is available.
- **FR-008**: If multiple base-spirit filter clauses appear in one query, only the first is applied; no error is shown to the user.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A search for `base spirit is <X>` returns only recipes whose base spirit matches `<X>`; zero recipes without a matching base spirit appear in results.
- **SC-002**: The `is` and `=` syntaxes for the base-spirit filter return identical result sets for the same value.
- **SC-003**: A combined query (ingredient term + base spirit filter) returns the intersection of both constraints — no recipe missing either condition appears in results.
- **SC-004**: Searching for `rye whiskey` and `rye whisky` returns identical result sets in all search contexts (regular and base-spirit filter).
- **SC-005**: The base-spirit filter syntax is visible in the search hint text so users can discover it without external documentation.

## Assumptions

- The base-spirit flag already exists on recipe ingredients (introduced in feature 011); this feature adds new search/filter behaviour on top of existing data — no data model changes are required.
- Whiskey/whisky normalisation is applied to the search query text before it is sent; stored ingredient names are not modified.
- The `base spirit is` / `base spirit =` clause is parsed on the client side before the search request is dispatched, consistent with how the existing multi-ingredient AND-search (`+`) works.
- The feature applies to the recipe list page search only; recipe detail, admin, and other views are out of scope.
- Only one base-spirit filter clause per query is expected; the first occurrence wins.
- The existing backend search API already supports filtering by ingredient attributes including `is_base_spirit`; if not, the backend extension is in scope for this feature.
