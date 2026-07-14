# Phase 1 Data Model: Local DynamoDB Emulator

This feature introduces **no new application entities**. It changes where the existing entities are stored (DynamoDB everywhere) and defines the **table/index schema** the local emulator must provision to match production. The item attributes below already exist in `internal/model` and `internal/store/dynamo`; they are restated here as the schema contract `EnsureSchema` must create.

## Tables

### `recipes`

- **Primary key**: `id` (S) — hash only.
- **GSIs**: none.
- **Item attributes** (existing): `id`, `name`, `ingredients`, `steps`, `properties`, `notes`, `garnishes`, `related_ids`, `creator_id`, `created_at`, `updated_at`.
- **Access patterns**: GetByID (key); List/Search/Random/ListAll/ListByCreator/ExistsByName/ingredient & base-spirit search (Scan with filters — small dataset); SetRelated (Update + counterpart updates).

### `users`

- **Primary key**: `id` (S) — hash only.
- **GSI**: `username-index` — hash `username` (S), projection ALL.
- **Item attributes** (existing): `id`, `username`, `password_hash`, `is_admin`, `first_name`, `last_name`, `email`, `token_version`, `created_at`, and reset fields (`reset_token_hash`, `reset_token_expires`, `reset_window_start`, `reset_request_count`).
- **Access patterns**: GetByID (key); GetByUsername (**Query on `username-index`** — login path); GetByEmail & List & Count (Scan with filter); Create/Update/Delete.
- **Note**: The `username-index` GSI is mandatory — login fails without it. It must be created locally even though it is currently absent from the Terraform `users_table` module (see research D3).

### `favorites`

- **Primary key**: `user_id` (S, hash) + `recipe_id` (S, range) — composite.
- **GSI**: `recipe_id-index` — hash `recipe_id` (S), projection ALL (reserved for future count-by-recipe).
- **Item attributes** (existing): `user_id`, `recipe_id`, `created_at`.
- **Access patterns**: Add (PutItem), Remove (DeleteItem), Check (GetItem by composite key), List (Query on `user_id`).

## Provisioning entity: `TableNames`

A small config struct/value passed to `EnsureSchema`, carrying the three resolved table names (from `RECIPES_TABLE` / `USERS_TABLE` / `FAVORITES_TABLE`, with local defaults). It is the seam that lets tests provision uniquely-named tables while the server provisions the conventional ones.

- **Fields**: `Recipes string`, `Users string`, `Favorites string`.
- **Validation**: all three non-empty before any `CreateTable` call.

## Billing / options

- `BillingMode = PAY_PER_REQUEST` (matches production; DynamoDB Local ignores capacity anyway).
- No point-in-time recovery / SSE settings needed locally (managed-service-only concerns).

## State / lifecycle

- **Local/test tables** are ephemeral: created on demand (server startup or test setup), destroyed when the emulator container is removed (server) or by `t.Cleanup` (tests). No migration or data carry-over from SQLite (out of scope).
- **Provisioning is idempotent**: re-running against existing tables is a no-op (ResourceInUseException treated as success).
