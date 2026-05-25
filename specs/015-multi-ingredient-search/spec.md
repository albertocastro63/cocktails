# Feature Specification: Multi-Ingredient Search

**Feature Branch**: `015-multi-ingredient-search`
**Created**: 2026-05-24
**Status**: Draft
**Input**: User description: "Add the capability to search by two or more ingredients"

## User Scenarios & Testing

### User Story 1 — AND Search for Multiple Ingredients (Priority: P1)

A user browsing the all-recipes page wants to find every cocktail that contains a specific set of
ingredients. They type a compound query into the existing search bar using natural language
(`Gin and Lemon Juice`) or a shorthand (`Gin + Lemon + Sugar`). The list updates to show only
recipes whose ingredient list includes **all** of the named items.

**Why this priority**: This is the entire feature. No supporting stories are needed — the search
bar already exists and results are already displayed in a grid. The only new behaviour is the
multi-ingredient parsing and AND-matching logic.

**Independent Test**: Navigate to the all-recipes page, type `Gin and Lemon` in the search bar,
and verify that every returned recipe contains both "gin" and "lemon" in its ingredient list, and
that recipes containing only one of the two are absent.

**Acceptance Scenarios**:

1. **Given** recipes A (gin, lemon), B (gin, sugar), C (lemon, sugar) exist,
   **When** the user searches `gin and lemon`,
   **Then** only recipe A is returned.

2. **Given** the same recipes,
   **When** the user searches `gin + lemon + sugar`,
   **Then** no recipes are returned (none has all three).

3. **Given** the same recipes,
   **When** the user searches `gin`,
   **Then** recipes A and B are returned (single-term search keeps existing behaviour).

4. **Given** the same recipes,
   **When** the user searches `Gin AND Lemon` (uppercase),
   **Then** recipe A is returned (matching is case-insensitive).

5. **Given** the user types a compound query with extra whitespace (`gin  +  lemon`),
   **Then** parsing is tolerant and recipe A is returned.

6. **Given** the user clears the search bar,
   **Then** all recipes are shown (existing empty-query behaviour is unchanged).

---

### Edge Cases

- What if one ingredient term in the compound query is empty after trimming (e.g., `gin + + lemon`)? Empty tokens are silently discarded; remaining valid tokens are used.
- What if all tokens are empty after trimming? Falls back to full listing (same as empty query).
- What if only one non-empty token remains after parsing? Falls back to existing single-term search (searches name, ingredients, steps, properties via FTS/substring).
- What if an ingredient name contains the word "and" (e.g., "salt and pepper")? The split is word-boundary-aware: ` and ` (with surrounding spaces) is the delimiter, so `salt and pepper` as a single token is safe if the user types it as one ingredient. However, if they type `salt and pepper and lime`, it splits to `["salt", "pepper", "lime"]`.

## Requirements

### Functional Requirements

- **FR-001**: The system MUST parse the `q` query parameter for multi-ingredient syntax by splitting on ` and ` (case-insensitive, requires surrounding spaces) and ` + ` (with optional surrounding whitespace).
- **FR-002**: When two or more ingredient tokens are detected, the system MUST return only recipes whose ingredient list contains **all** tokens (case-insensitive substring match on ingredient name).
- **FR-003**: When one token is detected, the system MUST fall back to the existing single-term search behaviour (full-text search across name, ingredients, steps, properties).
- **FR-004**: When zero tokens are detected (empty query), the system MUST return all recipes (existing listing behaviour).
- **FR-005**: The search UI MUST display a hint near the search bar informing users of the multi-ingredient syntax (e.g., `Use "and" or "+" to filter by multiple ingredients`).
- **FR-006**: Multi-ingredient matching MUST be case-insensitive and substring-based (searching for "gin" matches "Gin", "Sloe Gin", "Gin Fizz Base").
- **FR-007**: The existing pagination and sort behaviour MUST remain unchanged for multi-ingredient search results.

### Key Entities

- **IngredientQuery**: A parsed representation of a compound ingredient search — a list of one or more trimmed, lowercased ingredient name tokens derived from the raw query string.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A user can find all cocktails containing a given set of two or more ingredients with a single search action, without navigating away from the recipe list.
- **SC-002**: Multi-ingredient search results are returned within the same response-time budget as single-term search (p95 ≤ 200 ms on the backend).
- **SC-003**: Single-term search behaviour is fully preserved — existing searches return identical results before and after this feature is deployed.
- **SC-004**: The search syntax hint is visible without scrolling on all screen sizes where the search bar is visible.

## Assumptions

- The search bar component (`SearchBar`) is already present on the all-recipes page; no new UI widget is needed.
- Multi-ingredient search applies only to the all-recipes page (`GET /api/v1/recipes?q=...`). The My Recipes page is out of scope for this feature.
- The `and` keyword split requires surrounding whitespace to avoid false splits on ingredient names containing "and" (e.g., "Salt and Pepper" used as a single ingredient name will be split if typed verbatim; this is documented as known behaviour and acceptable for v1).
- DynamoDB performance: the current implementation performs a full table scan for all text searches; multi-ingredient search does not worsen this — it adds an in-memory filter. A future GSI-based approach is out of scope.
- No changes to the data model or storage schema are required.

## Clarifications

### Session 2026-05-24

- Q: Should the `+` delimiter require surrounding spaces, or should `Gin+Lemon` (no spaces) also split? → A: Both `Gin + Lemon` and `Gin+Lemon` must split (whitespace around `+` is optional).
