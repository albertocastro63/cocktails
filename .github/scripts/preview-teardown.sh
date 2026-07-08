#!/usr/bin/env bash
set -euo pipefail

# Required env vars (injected by preview-teardown.yml):
#   PR_NUMBER, AWS_REGION, API_GATEWAY_ID,
#   FRONTEND_BUCKET, CLOUDFRONT_DISTRIBUTION_ID

log() { echo "[preview-teardown] $*"; }

PR_ID="pr-${PR_NUMBER}"
FUNCTION_NAME="cocktails-${PR_ID}-api"
RECIPES_TABLE="cocktails-${PR_ID}-recipes"
USERS_TABLE="cocktails-${PR_ID}-users"
FAVORITES_TABLE="cocktails-${PR_ID}-favorites"

# ── Step 1: Lambda function ──────────────────────────────────────────────────

delete_lambda() {
  log "Deleting Lambda ${FUNCTION_NAME}..."
  aws lambda delete-function \
    --function-name "${FUNCTION_NAME}" \
    --region "${AWS_REGION}" > /dev/null 2>&1 || true
}

# ── Step 2: API Gateway route + integration ──────────────────────────────────

delete_api_route() {
  local ROUTE_KEY="ANY /${PR_ID}/api/{proxy+}"
  log "Removing API Gateway route ${ROUTE_KEY}..."

  local ROUTE_ID
  ROUTE_ID=$(aws apigatewayv2 get-routes \
    --api-id "${API_GATEWAY_ID}" \
    --region "${AWS_REGION}" \
    --query "Items[?RouteKey=='${ROUTE_KEY}'].RouteId | [0]" \
    --output text 2>/dev/null || echo "None")

  if [[ "${ROUTE_ID}" == "None" || -z "${ROUTE_ID}" ]]; then
    log "  Route not found, skipping."
    return
  fi

  local INTEGRATION_ID
  INTEGRATION_ID=$(aws apigatewayv2 get-route \
    --api-id "${API_GATEWAY_ID}" \
    --route-id "${ROUTE_ID}" \
    --region "${AWS_REGION}" \
    --query 'Target' --output text | sed 's|integrations/||')

  aws apigatewayv2 delete-route \
    --api-id "${API_GATEWAY_ID}" \
    --route-id "${ROUTE_ID}" \
    --region "${AWS_REGION}" > /dev/null 2>&1 || true

  if [[ -n "${INTEGRATION_ID}" && "${INTEGRATION_ID}" != "None" ]]; then
    aws apigatewayv2 delete-integration \
      --api-id "${API_GATEWAY_ID}" \
      --integration-id "${INTEGRATION_ID}" \
      --region "${AWS_REGION}" > /dev/null 2>&1 || true
  fi
}

# ── Step 3: DynamoDB tables + log group ─────────────────────────────────────

delete_dynamo_tables() {
  log "Deleting DynamoDB tables for ${PR_ID}..."
  for TABLE in "${RECIPES_TABLE}" "${USERS_TABLE}" "${FAVORITES_TABLE}"; do
    aws dynamodb delete-table \
      --table-name "${TABLE}" \
      --region "${AWS_REGION}" > /dev/null 2>&1 || true
  done

  for TABLE in "${RECIPES_TABLE}" "${USERS_TABLE}" "${FAVORITES_TABLE}"; do
    aws dynamodb wait table-not-exists \
      --table-name "${TABLE}" \
      --region "${AWS_REGION}" 2>/dev/null || true
  done

  log "Deleting Lambda log group for ${FUNCTION_NAME}..."
  aws logs delete-log-group \
    --log-group-name "/aws/lambda/${FUNCTION_NAME}" \
    --region "${AWS_REGION}" > /dev/null 2>&1 || true
}

# ── Step 4: S3 frontend assets ───────────────────────────────────────────────

delete_s3_assets() {
  log "Deleting S3 objects at s3://${FRONTEND_BUCKET}/${PR_ID}/..."
  aws s3 rm "s3://${FRONTEND_BUCKET}/${PR_ID}/" \
    --recursive \
    --region "${AWS_REGION}" > /dev/null 2>&1 || true
}

# ── Step 5: CloudFront invalidation ─────────────────────────────────────────

invalidate_cloudfront() {
  log "Invalidating CloudFront cache for /${PR_ID}/*..."
  aws cloudfront create-invalidation \
    --distribution-id "${CLOUDFRONT_DISTRIBUTION_ID}" \
    --paths "/${PR_ID}/*" > /dev/null
}

# ── Main ─────────────────────────────────────────────────────────────────────

delete_lambda
delete_api_route
delete_dynamo_tables
delete_s3_assets
invalidate_cloudfront

log "✓ Preview environment ${PR_ID} torn down."
