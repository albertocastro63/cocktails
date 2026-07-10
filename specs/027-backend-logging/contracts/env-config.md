# Contract: Environment Configuration

## `LOG_LEVEL`

The single variable controlling log verbosity (FR-002, FR-004).

| Property | Value |
|----------|-------|
| Name | `LOG_LEVEL` |
| Accepted values | `debug`, `info`, `warn` (alias `warning`), `error` (case-insensitive) |
| Missing / invalid | Falls back to `error` only; one ERROR line records the fallback (FR-005) |
| Production default | `warn` — set in `infra/main.tf` Lambda `environment_variables` |
| Preview default | `debug` — set in `.github/scripts/preview-deploy.sh` Lambda env |
| Runtime change | Editable in the AWS Lambda console; applies to subsequent invocations without redeploy (FR-004) |

### Terraform (production) — excerpt to add

```hcl
environment_variables = {
  # ...existing (STORE_BACKEND, *_TABLE, JWT_SECRET, MAIL_FROM, APP_BASE_URL)...
  LOG_LEVEL = "warn"
}
```

### Preview deploy script — excerpt to add

Append `LOG_LEVEL=debug` to the `Variables={...}` set used by `create-function` / `update-function-configuration` for the preview Lambda.

## Verification

- `aws lambda get-function-configuration --function-name <fn> --query 'Environment.Variables.LOG_LEVEL'` returns the expected value per environment.
- Changing the value in the console and issuing a new request produces log lines at the new verbosity within one minute (SC-003).
