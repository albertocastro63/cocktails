#!/usr/bin/env bash
set -euo pipefail

# Required env vars (injected by preview-deploy.yml):
#   PR_NUMBER, AWS_REGION, LAMBDA_ROLE_ARN, API_GATEWAY_ID,
#   FRONTEND_BUCKET, CLOUDFRONT_DISTRIBUTION_ID, JWT_SECRET,
#   PROD_RECIPES_TABLE

log() { echo "[preview-deploy] $*"; }

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

def flush(b):
    req = {'RequestItems': {dst: [{'PutRequest': {'Item': i}} for i in b]}}
    subprocess.run(
        ['aws', 'dynamodb', 'batch-write-item',
         '--request-items', json.dumps(req),
         '--region', region],
        check=True, capture_output=True
    )

for item in items:
    batch.append(item)
    if len(batch) == 25:
        flush(batch)
        batch = []
if batch:
    flush(batch)
"
}

create_and_seed_tables() {
  log "Creating DynamoDB tables for ${PR_ID}..."

  aws dynamodb create-table \
    --table-name "${RECIPES_TABLE}" \
    --billing-mode PAY_PER_REQUEST \
    --attribute-definitions AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --region "${AWS_REGION}" > /dev/null

  aws dynamodb create-table \
    --table-name "${USERS_TABLE}" \
    --billing-mode PAY_PER_REQUEST \
    --attribute-definitions AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
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

  # Only the recipes table is seeded from production. The users and favorites
  # tables are created empty: previews are publicly accessible (FR-011), so
  # production user records and their favorites (PII) are never copied into them.
  seed_table "${PROD_RECIPES_TABLE}" "${RECIPES_TABLE}"
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

  if aws lambda get-function --function-name "${FUNCTION_NAME}" \
      --region "${AWS_REGION}" > /dev/null 2>&1; then
    log "Updating Lambda code for ${FUNCTION_NAME}..."
    aws lambda update-function-code \
      --function-name "${FUNCTION_NAME}" \
      --zip-file "fileb://${ZIP}" \
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
      --environment "Variables={STORE_BACKEND=dynamodb,RECIPES_TABLE=${RECIPES_TABLE},USERS_TABLE=${USERS_TABLE},FAVORITES_TABLE=${FAVORITES_TABLE},JWT_SECRET=${JWT_SECRET},STRIP_PATH_PREFIX=/${PR_ID}}" \
      --region "${AWS_REGION}" > /dev/null
  fi
}

# ── Step 3: API Gateway route ────────────────────────────────────────────────

provision_api_route() {
  local ROUTE_KEY="ANY /${PR_ID}/api/{proxy+}"

  local LAMBDA_ARN
  LAMBDA_ARN=$(aws lambda get-function --function-name "${FUNCTION_NAME}" \
    --region "${AWS_REGION}" --query 'Configuration.FunctionArn' --output text)

  # Grant API Gateway permission to invoke the Lambda (idempotent via || true)
  aws lambda add-permission \
    --function-name "${FUNCTION_NAME}" \
    --statement-id "apigw-${PR_ID}" \
    --action lambda:InvokeFunction \
    --principal apigateway.amazonaws.com \
    --source-arn "arn:aws:execute-api:${AWS_REGION}:*:${API_GATEWAY_ID}/*/*" \
    --region "${AWS_REGION}" > /dev/null 2>&1 || true

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
  aws s3 sync frontend/dist/ "s3://${FRONTEND_BUCKET}/${PR_ID}/" \
    --delete \
    --region "${AWS_REGION}"
}

# ── Main ─────────────────────────────────────────────────────────────────────

maybe_create_tables
provision_lambda
provision_api_route
upload_frontend

PREVIEW_URL="https://cocktails.albertomcastro.com/${PR_ID}/"
log "✓ Preview environment ready: ${PREVIEW_URL}"
echo "${PREVIEW_URL}"
