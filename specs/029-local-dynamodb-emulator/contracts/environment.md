# Contract: Environment & Store Selection

This feature has no new HTTP/API surface. Its "contract" is the **environment-variable contract** that selects and configures the store, plus the **schema contract** (see `data-model.md`) the emulator must satisfy. Downstream tasks and docs must honor this.

## Store selection contract

| Condition | Behavior |
|---|---|
| `DYNAMODB_ENDPOINT` set | Build the SDK client against that endpoint with static dummy credentials + region. Used by local dev and tests (the emulator). |
| `DYNAMODB_ENDPOINT` unset | Build the SDK client from the default AWS config chain (IAM role/credentials). Used by production (Lambda). |
| (any) | Store is **always** DynamoDB. There is no SQLite path and no `STORE_BACKEND=sqlite`. |

- `STORE_BACKEND` is **removed** as a selector (it only ever chose sqlite vs dynamodb). If retained transitionally, only `dynamodb` is valid; any other value is an error, never a silent fallback.
- `DB_PATH` is **removed** entirely.

## Environment variables (backend)

| Variable | Local (emulator) | Production | Notes |
|---|---|---|---|
| `DYNAMODB_ENDPOINT` | `http://localhost:8000` | unset | Presence = "use emulator". |
| `AWS_REGION` | `us-east-1` (any) | set by Lambda | Required by the SDK; dummy locally. |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | `test` / `test` | via role | Dummy locally; DynamoDB Local ignores their value but the SDK requires them. |
| `RECIPES_TABLE` / `USERS_TABLE` / `FAVORITES_TABLE` | local defaults (e.g. `cocktails-recipes`, …) | Terraform-set names | Passed to `EnsureSchema` and the stores. |
| `ADMIN_BOOTSTRAP_PASSWORD` | optional | set out-of-band | Unchanged bootstrap behavior. |

## Provisioning contract (`EnsureSchema`)

- Signature (indicative): `EnsureSchema(ctx context.Context, client *dynamodb.Client, names TableNames) error`.
- MUST create exactly the tables/GSIs in `data-model.md` when absent, wait until `ACTIVE`, and treat "already exists" as success (idempotent).
- MUST be invoked by the local server on startup when `DYNAMODB_ENDPOINT` is set, and MUST NOT run in production (tables are Terraform-managed there).
- MUST be the single provisioning path used by tests (replacing the ad-hoc `createTable` helper).

## Failure contract

- If `DYNAMODB_ENDPOINT` is set but the emulator is unreachable, the server MUST exit with a clear, actionable message (start the emulator) — not a stack trace, not a hang.
- If store/integration tests run without `DYNAMODB_ENDPOINT`, they MUST fail with an actionable message, not skip and not fall back.

## CI contract

- The `backend` job keeps the `amazon/dynamodb-local` service and `DYNAMODB_ENDPOINT=http://localhost:8000`.
- `go test -p 1 -coverprofile=… -coverpkg=./internal/... ./...` MUST pass with total coverage ≥ 75% after SQLite removal.
