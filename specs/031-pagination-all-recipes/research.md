# Phase 0 Research: Pagination for All Recipes

## Decision 1: Interaction model — classic page controls

- **Decision**: Use classic pagination controls — Previous / Next buttons, a "Page X of Y" indicator, and a total-results count — rather than infinite scroll or a "Load more" button.
- **Rationale**: The feature request explicitly asks for "page navigation controls, showing total count and current page." Classic controls are the least ambiguous, are trivially testable (deterministic page boundaries), and match the backend's existing `page`/`limit`/`total` response shape.
- **Alternatives considered**:
  - *Infinite scroll*: better for very large media feeds, but hides total/position, is harder to test deterministically, and contradicts the request.
  - *Load-more*: simpler but still does not give an explicit page position, and accumulates all loaded items in the DOM.

## Decision 2: Keep server-side pagination (do not fetch-all client-side)

- **Decision**: Continue to let the backend paginate and return one page plus `total`; the frontend requests one page at a time.
- **Rationale**: The endpoint already returns `{data, total, page, limit}`. Honoring it keeps payloads small and page boundaries authoritative. Fetching the whole catalog to paginate in the browser would defeat the endpoint's design and scale poorly as the catalog grows.
- **Alternatives considered**: Client fetches every page then paginates in the UI — rejected; reintroduces the "load everything" problem the endpoint exists to avoid and makes `total` redundant.

## Decision 3: Move sort to the server (`sort` query param)

- **Decision**: Add an optional `sort` parameter to `GET /v1/recipes` with values `name_asc` and `name_desc`. The store applies a stable sort over the full result set *before* slicing the requested page. The frontend stops sorting client-side and instead passes `sort` and resets to page 1 when the sort changes. When no sort is selected, the existing scan order is used (deterministic for a stable table).
- **Rationale**: Today `RecipeList.sortRecipes` sorts only the data currently loaded. With pagination, client-side sort would reorder a single page, so A→Z on page 1 and page 2 would each be locally sorted but globally wrong — a direct violation of FR-012 (consistent order across page boundaries). Sorting server-side over the full set before paging is the only way to keep ordering globally consistent while still shipping one page at a time.
- **Alternatives considered**:
  - *Keep client-side sort*: rejected — breaks FR-012 once more than one page exists.
  - *Add a DynamoDB GSI / sort key*: rejected as overkill for a small catalog; an in-memory stable sort over the already-scanned items is sufficient and avoids schema/index changes.

## Decision 4: Fix Scan pagination with a shared `scanAll` helper

- **Decision**: Introduce `scanAll(ctx, client, input) ([]map[string]types.AttributeValue, error)` in the dynamo package that loops on `LastEvaluatedKey` / `ExclusiveStartKey` until the scan is exhausted, and route every store method that currently issues a single `Scan` (`List`, `Search`, `Random`, `ExistsByName`, `ListByCreator`, `ListAll`, `SearchByIngredients`, `SearchByBaseSpirit`, `SearchByBaseSpiritAndIngredients`) through it.
- **Rationale**: A DynamoDB `Scan` returns at most 1 MB per call. Every store method currently reads only the first batch, so once the recipes table exceeds ~1 MB both `total` and the reachable set silently under-count — exactly the class of "silently dropped recipes" bug this feature fixes, one layer down (FR-009). A single helper removes the duplicated single-Scan pattern (Code Quality principle) and fixes all call sites at once.
- **Alternatives considered**:
  - *Use `dynamodb.NewScanPaginator`*: viable, but a small explicit loop keeps the helper dependency-light, uniform with the existing hand-written client code, and easy to unit-test against DynamoDB Local.
  - *Fix only `List`*: rejected — leaves the same latent bug in Search/ListAll/etc.

## Decision 5: Page size and bounds

- **Decision**: Keep the default `limit=20` and the existing cap of `limit ≤ 100`. Requesting a page beyond the available range returns an empty `data` array with the correct `total` (no error), which the frontend renders as "no results on this page" while keeping controls usable to go back.
- **Rationale**: Preserves current behavior and the contract; the handler already computes `total` independently of the returned slice, so an out-of-range page naturally yields an empty slice with a correct total (FR-011). Default 20 is a reasonable grid size for the 3-column layout.
- **Alternatives considered**: Larger default (50/100) — rejected; it only postpones the problem and this feature makes every page reachable regardless of size.

## Open questions

- None. All spec `[NEEDS CLARIFICATION]` markers were resolved during specification; no unknowns remain in Technical Context.
