# Contract: `GET /v1/recipes` (listing with pagination & sort)

Status: existing endpoint, extended with a `sort` parameter. Response shape
unchanged. This contract documents the behavior the frontend relies on.

## Request

```
GET {API_PREFIX}/v1/recipes?page={n}&limit={m}&q={query}&base_spirit={spirit}&sort={order}
```

All query parameters are optional.

| Param         | Type    | Default | Notes |
|---------------|---------|---------|-------|
| `page`        | integer | 1       | 1-based page index. |
| `limit`       | integer | 20      | Page size; server caps at 100. |
| `q`           | string  | —       | Search term / ingredient tokens (`and` / `+` for multiple). Unchanged. |
| `base_spirit` | string  | —       | Base-spirit filter. Unchanged. |
| `sort`        | string  | —       | `name_asc` or `name_desc`. Unknown/empty ⇒ default order. **New.** |

No authentication required (public listing). An optional bearer token continues to
be used only to decorate favorites and is out of scope here.

## Response — 200 OK

```json
{
  "data": [ { "id": "…", "name": "…", "ingredients": [ … ], "steps": [ … ], "...": "…" } ],
  "total": 30,
  "page": 1,
  "limit": 20
}
```

- `data`: recipes on the requested page, after filtering and sorting. May be `[]`.
- `total`: total number of recipes in the filtered result set across all pages.
  MUST count every stored recipe, not only those in the first storage scan batch.
- `page`, `limit`: echoed back.

## Behavioral guarantees

1. **Completeness**: Summed over `page = 1..ceil(total/limit)`, `data` contains every
   recipe in the result set exactly once (stable catalog, fixed `sort`).
2. **Total accuracy**: `total` equals the true count of matching recipes regardless of
   underlying storage batch limits (FR-009).
3. **Sort**: With `sort=name_asc` / `name_desc`, the full result set is ordered by name
   (case-insensitive) before the page is sliced, so order is consistent across pages
   (FR-012). Without `sort`, a deterministic default order is used.
4. **Out-of-range page**: A `page` beyond the last page returns `200` with `data: []`
   and the correct `total` (never an error) (FR-011).
5. **Empty result set**: `total: 0`, `data: []`.
6. **Filtering**: `sort` composes with `q` and `base_spirit`; `total` reflects the
   filtered set, not the whole catalog (FR-007).

## Errors

| Status | Code             | When |
|--------|------------------|------|
| 500    | `INTERNAL_ERROR` | Storage read failed. Body: `{"error":{"code":"INTERNAL_ERROR","message":"failed to retrieve recipes"}}`. |

Malformed `page`/`limit`/`sort` values degrade to defaults rather than returning 4xx,
consistent with the endpoint's current lenient parsing.

## Contract tests (to be added)

- `page`/`limit` slice correctly; `total` independent of returned slice size.
- Out-of-range page ⇒ empty `data`, correct `total`.
- `sort=name_asc` / `name_desc` orders the whole set before paging (page 1 last item
  sorts before page 2 first item).
- Unknown `sort` ⇒ default order, no error.
- `sort` combined with `q` / `base_spirit` reflects filtered totals.
