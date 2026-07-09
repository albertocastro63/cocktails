#!/usr/bin/env bash
set -euo pipefail

# Required env vars (injected by preview-deploy.yml):
#   PR_NUMBER, AWS_REGION, LAMBDA_ROLE_ARN, API_GATEWAY_ID,
#   FRONTEND_BUCKET, CLOUDFRONT_DISTRIBUTION_ID, JWT_SECRET,
#   PROD_RECIPES_TABLE, PROD_USERS_TABLE
# Optional:
#   PREVIEW_SEED_USERNAMES (default "admin alberto") — the prod usernames whose
#   records are copied into the preview users table so it is testable via login.

# Logs go to stderr so the only thing on stdout is the final preview URL,
# which the workflow captures. Mixing logs into stdout previously caused the
# PR comment to show a log line instead of the URL.
log() { echo "[preview-deploy] $*" >&2; }

PR_ID="pr-${PR_NUMBER}"
FUNCTION_NAME="cocktails-${PR_ID}-api"
RECIPES_TABLE="cocktails-${PR_ID}-recipes"
USERS_TABLE="cocktails-${PR_ID}-users"
FAVORITES_TABLE="cocktails-${PR_ID}-favorites"

# ── Step 1: DynamoDB tables ──────────────────────────────────────────────────

seed_table() {
  local SRC="$1" DST="$2"
  log "Seeding ${DST} from ${SRC}..."

  local ITEMS
  ITEMS=$(aws dynamodb scan --table-name "${SRC}" --region "${AWS_REGION}" \
    --query 'Items' --output json)

  local COUNT
  COUNT=$(echo "${ITEMS}" | python3 -c "import sys, json; print(len(json.load(sys.stdin)))")
  if [[ "${COUNT}" -eq 0 ]]; then
    log "  ${SRC} is empty, nothing to seed."
    return
  fi

  echo "${ITEMS}" | _SEED_DST="${DST}" python3 -c "
import sys, json, subprocess, os

items = json.load(sys.stdin)
dst   = os.environ['_SEED_DST']
region = os.environ['AWS_REGION']
batch = []

def write_items(request_items):
    # NOTE: the CLI --request-items value IS the {TableName: [...]} map itself,
    # NOT wrapped in a {'RequestItems': ...} object (that wrapper is only for
    # the boto3/API shape and causes a ParamValidation error via the CLI).
    p = subprocess.run(
        ['aws', 'dynamodb', 'batch-write-item',
         '--request-items', json.dumps(request_items),
         '--region', region],
        capture_output=True, text=True
    )
    if p.returncode != 0:
        sys.stderr.write(p.stderr)
        raise SystemExit(f'batch-write-item failed (exit {p.returncode})')
    return json.loads(p.stdout or '{}').get('UnprocessedItems') or {}

def flush(b):
    unprocessed = write_items({dst: [{'PutRequest': {'Item': i}} for i in b]})
    # Retry any throttled/unprocessed items a few times before giving up.
    for _ in range(5):
        if not unprocessed:
            break
        unprocessed = write_items(unprocessed)
    if unprocessed:
        raise SystemExit('batch-write-item left unprocessed items after retries')

for item in items:
    batch.append(item)
    if len(batch) == 25:
        flush(batch)
        batch = []
if batch:
    flush(batch)
"
}

# Copy a fixed set of production users (by username) into the preview users
# table so the preview is testable via login. Uses the production username-index
# GSI to look each account up, then writes it verbatim (keeping the same password
# hash, so the account's real password works). Missing usernames are skipped.
seed_users() {
  local names="${PREVIEW_SEED_USERNAMES:-admin alberto}"
  log "Seeding preview users (${names}) from ${PROD_USERS_TABLE}..."
  for uname in ${names}; do
    local item
    item=$(aws dynamodb query \
      --table-name "${PROD_USERS_TABLE}" \
      --index-name username-index \
      --key-condition-expression 'username = :u' \
      --expression-attribute-values "{\":u\":{\"S\":\"${uname}\"}}" \
      --region "${AWS_REGION}" \
      --query 'Items[0]' --output json)
    if [[ "${item}" == "null" || -z "${item}" ]]; then
      log "  user '${uname}' not found in ${PROD_USERS_TABLE}, skipping."
      continue
    fi
    aws dynamodb put-item \
      --table-name "${USERS_TABLE}" \
      --item "${item}" \
      --region "${AWS_REGION}"
    log "  seeded user '${uname}'."
  done
}

create_and_seed_tables() {
  log "Creating DynamoDB tables for ${PR_ID}..."

  aws dynamodb create-table \
    --table-name "${RECIPES_TABLE}" \
    --billing-mode PAY_PER_REQUEST \
    --attribute-definitions AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --region "${AWS_REGION}" > /dev/null

  # The users table needs the same username-index GSI as production: the backend
  # looks up users by username via that index (login/auth). Without it, any
  # seeded login fails with "Invalid Credentials".
  aws dynamodb create-table \
    --table-name "${USERS_TABLE}" \
    --billing-mode PAY_PER_REQUEST \
    --attribute-definitions AttributeName=id,AttributeType=S AttributeName=username,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --global-secondary-indexes '[{"IndexName":"username-index","KeySchema":[{"AttributeName":"username","KeyType":"HASH"}],"Projection":{"ProjectionType":"ALL"}}]' \
    --region "${AWS_REGION}" > /dev/null

  aws dynamodb create-table \
    --table-name "${FAVORITES_TABLE}" \
    --billing-mode PAY_PER_REQUEST \
    --attribute-definitions AttributeName=user_id,AttributeType=S AttributeName=recipe_id,AttributeType=S \
    --key-schema AttributeName=user_id,KeyType=HASH AttributeName=recipe_id,KeyType=RANGE \
    --region "${AWS_REGION}" > /dev/null

  log "Waiting for tables to become ACTIVE..."
  aws dynamodb wait table-exists --table-name "${RECIPES_TABLE}"   --region "${AWS_REGION}"
  aws dynamodb wait table-exists --table-name "${USERS_TABLE}"     --region "${AWS_REGION}"
  aws dynamodb wait table-exists --table-name "${FAVORITES_TABLE}" --region "${AWS_REGION}"

  # Recipes are seeded in full. For users, only a fixed, small set of accounts
  # (PREVIEW_SEED_USERNAMES, default "admin alberto") is copied so the preview is
  # testable via login. NOTE: this intentionally copies real user records (incl.
  # bcrypt password hashes) into a publicly accessible preview — scoped to these
  # test accounts only. The favorites table is left empty.
  seed_table "${PROD_RECIPES_TABLE}" "${RECIPES_TABLE}"
  seed_users
}

maybe_create_tables() {
  if aws dynamodb describe-table --table-name "${RECIPES_TABLE}" --region "${AWS_REGION}" \
      --query 'Table.TableStatus' --output text 2>/dev/null | grep -q "ACTIVE"; then
    log "Tables already exist for ${PR_ID}, skipping creation and seeding."
  else
    create_and_seed_tables
  fi
}

# ── Step 2: Lambda function ──────────────────────────────────────────────────

provision_lambda() {
  # Accept the zip at the CI-standard path or at repo root
  local ZIP="backend/bin/bootstrap.zip"
  [[ -f "${ZIP}" ]] || ZIP="bootstrap.zip"

  # Preview Lambda environment. MAIL_FROM enables the real SES sender so
  # previews exercise the full password-recovery email flow (the preview Lambda
  # role already has scoped ses:SendEmail); APP_BASE_URL makes the reset link in
  # the email point back to this preview (no trailing slash — the handler adds
  # the /#/reset path).
  local ENV_VARS="Variables={STORE_BACKEND=dynamodb,RECIPES_TABLE=${RECIPES_TABLE},USERS_TABLE=${USERS_TABLE},FAVORITES_TABLE=${FAVORITES_TABLE},JWT_SECRET=${JWT_SECRET},STRIP_PATH_PREFIX=/${PR_ID},MAIL_FROM=no-reply@cocktails.albertomcastro.com,APP_BASE_URL=https://cocktails.albertomcastro.com/${PR_ID}}"

  if aws lambda get-function --function-name "${FUNCTION_NAME}" \
      --region "${AWS_REGION}" > /dev/null 2>&1; then
    log "Updating Lambda code for ${FUNCTION_NAME}..."
    aws lambda update-function-code \
      --function-name "${FUNCTION_NAME}" \
      --zip-file "fileb://${ZIP}" \
      --region "${AWS_REGION}" > /dev/null

    # update-function-code does not touch configuration, so an already-existing
    # preview Lambda would keep stale env vars (e.g. no MAIL_FROM). Wait for the
    # code update to settle, then sync the environment explicitly.
    log "Syncing Lambda environment for ${FUNCTION_NAME}..."
    aws lambda wait function-updated \
      --function-name "${FUNCTION_NAME}" --region "${AWS_REGION}"
    aws lambda update-function-configuration \
      --function-name "${FUNCTION_NAME}" \
      --environment "${ENV_VARS}" \
      --region "${AWS_REGION}" > /dev/null
  else
    log "Creating Lambda function ${FUNCTION_NAME}..."
    aws lambda create-function \
      --function-name "${FUNCTION_NAME}" \
      --runtime provided.al2023 \
      --architectures arm64 \
      --role "${LAMBDA_ROLE_ARN}" \
      --handler bootstrap \
      --zip-file "fileb://${ZIP}" \
      --environment "${ENV_VARS}" \
      --region "${AWS_REGION}" > /dev/null
  fi
}

# ── Step 3: API Gateway route ────────────────────────────────────────────────

provision_api_route() {
  local ROUTE_KEY="ANY /${PR_ID}/api/{proxy+}"

  local LAMBDA_ARN
  LAMBDA_ARN=$(aws lambda get-function --function-name "${FUNCTION_NAME}" \
    --region "${AWS_REGION}" --query 'Configuration.FunctionArn' --output text)

  # Account ID is required in the source ARN — a wildcard (*) in the account
  # field fails ARN validation, which silently left API Gateway unable to
  # invoke the Lambda (500 Internal Server Error). Derive it from the ARN.
  local ACCOUNT_ID
  ACCOUNT_ID=$(echo "${LAMBDA_ARN}" | cut -d: -f5)

  # Grant API Gateway permission to invoke the Lambda. Remove any stale
  # statement first (idempotent), then add without masking real failures.
  aws lambda remove-permission \
    --function-name "${FUNCTION_NAME}" \
    --statement-id "apigw-${PR_ID}" \
    --region "${AWS_REGION}" > /dev/null 2>&1 || true
  aws lambda add-permission \
    --function-name "${FUNCTION_NAME}" \
    --statement-id "apigw-${PR_ID}" \
    --action lambda:InvokeFunction \
    --principal apigateway.amazonaws.com \
    --source-arn "arn:aws:execute-api:${AWS_REGION}:${ACCOUNT_ID}:${API_GATEWAY_ID}/*/*" \
    --region "${AWS_REGION}" > /dev/null

  local EXISTING_ROUTE_ID
  EXISTING_ROUTE_ID=$(aws apigatewayv2 get-routes \
    --api-id "${API_GATEWAY_ID}" \
    --region "${AWS_REGION}" \
    --query "Items[?RouteKey=='${ROUTE_KEY}'].RouteId | [0]" \
    --output text)

  if [[ "${EXISTING_ROUTE_ID}" != "None" && -n "${EXISTING_ROUTE_ID}" ]]; then
    log "API Gateway route ${ROUTE_KEY} already exists."
    return
  fi

  log "Adding API Gateway route ${ROUTE_KEY}..."
  local INTEGRATION_ID
  INTEGRATION_ID=$(aws apigatewayv2 create-integration \
    --api-id "${API_GATEWAY_ID}" \
    --integration-type AWS_PROXY \
    --integration-uri "${LAMBDA_ARN}" \
    --payload-format-version "2.0" \
    --region "${AWS_REGION}" \
    --query 'IntegrationId' --output text)

  aws apigatewayv2 create-route \
    --api-id "${API_GATEWAY_ID}" \
    --route-key "${ROUTE_KEY}" \
    --target "integrations/${INTEGRATION_ID}" \
    --region "${AWS_REGION}" > /dev/null
}

# ── Step 4: Frontend upload ──────────────────────────────────────────────────

upload_frontend() {
  log "Uploading frontend assets to s3://${FRONTEND_BUCKET}/${PR_ID}/..."
  # Send sync progress to stderr so stdout stays clean for the preview URL.
  aws s3 sync frontend/dist/ "s3://${FRONTEND_BUCKET}/${PR_ID}/" \
    --delete \
    --region "${AWS_REGION}" >&2
}

# ── Main ─────────────────────────────────────────────────────────────────────

maybe_create_tables
provision_lambda
provision_api_route
upload_frontend

PREVIEW_URL="https://cocktails.albertomcastro.com/${PR_ID}/"
log "✓ Preview environment ready: ${PREVIEW_URL}"
echo "${PREVIEW_URL}"
