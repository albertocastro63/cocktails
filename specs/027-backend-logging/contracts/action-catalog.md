# Contract: Action Log Catalog

The complete set of backend actions that MUST emit a log entry (FR-006, SC-001), with the action name, the success level (per FR-015), the handler/site, and the safe context fields. Failures for any action are logged at **ERROR** with `outcome=failure` and a sanitized `error`; auth rejections (bad credentials/expired token) are **WARN**.

| Action name | Success level | Handler (file.method) | Context fields |
|-------------|---------------|-----------------------|----------------|
| `auth.login` | INFO | auth.go · Login | `user_id`, `outcome` |
| `auth.logout` | INFO | auth.go · Logout* | `user_id`, `outcome` |
| `recipe.create` | INFO | recipes.go · Create | `user_id`, `recipe_id` |
| `recipe.update` | INFO | recipes.go · Update | `user_id`, `recipe_id` |
| `recipe.delete` | INFO | recipes.go · Delete | `user_id`, `recipe_id` |
| `recipe.get` | DEBUG | recipes.go · GetByID | `recipe_id` |
| `recipe.list` | DEBUG | recipes.go · List / Mine | `count`, filters |
| `recipe.random` | DEBUG | recipes.go · Random | `recipe_id` |
| `ingredients.list` | DEBUG | recipes.go (ingredients endpoint) | `count` |
| `search.ingredients` | DEBUG | recipes.go · List (ingredient filter) | `count`, query terms |
| `search.base_spirit` | DEBUG | recipes.go · List (base-spirit filter) | `count`, spirit |
| `favorite.add` | INFO | favorites.go · Add | `user_id`, `recipe_id` |
| `favorite.remove` | INFO | favorites.go · Remove | `user_id`, `recipe_id` |
| `favorite.check` | DEBUG | favorites.go · Check | `user_id`, `recipe_id` |
| `favorite.list` | DEBUG | favorites.go · List | `user_id`, `count` |
| `password.reset_request` | INFO | password_reset.go · Forgot | `user_id` (when resolved), `outcome` |
| `password.reset` | INFO | password_reset.go · Reset | `user_id`, `outcome` |
| `admin.user.create` | INFO | admin.go · CreateUser | `user_id` (actor), `target_id` |
| `admin.user.update` | INFO | admin.go · UpdateUser | `user_id`, `target_id` |
| `admin.user.delete` | INFO | admin.go · DeleteUser | `user_id`, `target_id` |
| `admin.user.list` / `get` | DEBUG | admin.go · ListUsers / GetUser | `count` / `target_id` |

\* **Logout note**: The backend is stateless-JWT; "logout" is client-side token disposal. If there is no server logout endpoint, `auth.logout` is emitted where the server does participate in session invalidation (e.g., token-version bump on password reset). Confirm during implementation and either instrument an existing endpoint or record this as N/A in tasks — do not invent an endpoint solely to log.

## Notes for implementation

- Names are stable identifiers for filtering; do not rename casually once shipped.
- Recoverable anomalies use WARN and a descriptive `msg` (e.g., `rate limit exceeded`, `LOG_LEVEL fallback applied`).
- Every failure path that currently calls `writeError(...)` in an instrumented handler should also emit the ERROR action log (with the same sanitized reason surfaced to the client).
