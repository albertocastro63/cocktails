locals {
  name_prefix = "${var.project_name}-${var.environment}"

  # Frontend assets path — operator must run: cd frontend && npm run build
  frontend_dist_path = "${path.root}/../frontend/dist"

  # MIME type map keyed by file extension (without leading dot)
  content_type_map = {
    "html"  = "text/html; charset=utf-8"
    "css"   = "text/css"
    "js"    = "application/javascript"
    "json"  = "application/json"
    "svg"   = "image/svg+xml"
    "png"   = "image/png"
    "jpg"   = "image/jpeg"
    "jpeg"  = "image/jpeg"
    "ico"   = "image/x-icon"
    "woff"  = "font/woff"
    "woff2" = "font/woff2"
    "ttf"   = "font/ttf"
    "webp"  = "image/webp"
    "txt"   = "text/plain"
    "xml"   = "application/xml"
  }
}

# ─────────────────────────────────────────────
# T009 — Artifact S3 Bucket
# Stores the compiled Lambda deployment zip before function creation.
# ─────────────────────────────────────────────

module "artifact_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "~> 5.0"

  bucket = "${local.name_prefix}-artifacts"

  versioning = {
    enabled = true
  }

  server_side_encryption_configuration = {
    rule = {
      apply_server_side_encryption_by_default = {
        sse_algorithm = "AES256"
      }
      bucket_key_enabled = true
    }
  }

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true

  lifecycle_rule = [
    {
      id      = "expire-old-versions"
      enabled = true

      noncurrent_version_expiration = {
        newer_noncurrent_versions = 3
        noncurrent_days           = 30
      }
    }
  ]
}

# ─────────────────────────────────────────────
# T010 — DynamoDB: Recipes Table
# ─────────────────────────────────────────────

module "recipes_table" {
  source  = "terraform-aws-modules/dynamodb-table/aws"
  version = "~> 5.0"

  name         = "${var.project_name}-recipes"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attributes = [
    {
      name = "id"
      type = "S"
    }
  ]

  point_in_time_recovery_enabled = true
  server_side_encryption_enabled = true
}

# ─────────────────────────────────────────────
# T011 — DynamoDB: Users Table
# ─────────────────────────────────────────────

module "users_table" {
  source  = "terraform-aws-modules/dynamodb-table/aws"
  version = "~> 5.0"

  name         = "${var.project_name}-users"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  # username-index (feature 029): declares the GSI login queries via
  # GetByUsername. The live table already carries this index (it was created
  # out-of-band), so this reconciles the definition with reality — Terraform
  # refreshes the GSI into state from the live table, making apply a no-op.
  attributes = [
    { name = "id", type = "S" },
    { name = "username", type = "S" },
  ]

  global_secondary_indexes = [
    {
      name            = "username-index"
      hash_key        = "username"
      projection_type = "ALL"
    }
  ]

  point_in_time_recovery_enabled = true
  server_side_encryption_enabled = true
}

# ─────────────────────────────────────────────
# Feature 014 — DynamoDB: Favorites Table
# Composite primary key: user_id (hash) + recipe_id (range).
# GSI on recipe_id supports future CountByRecipe queries.
# ─────────────────────────────────────────────

module "favorites_table" {
  source  = "terraform-aws-modules/dynamodb-table/aws"
  version = "~> 5.0"

  name         = "${var.project_name}-favorites"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "user_id"
  range_key    = "recipe_id"

  attributes = [
    { name = "user_id", type = "S" },
    { name = "recipe_id", type = "S" },
  ]

  global_secondary_indexes = [
    {
      name            = "recipe_id-index"
      hash_key        = "recipe_id"
      projection_type = "ALL"
    }
  ]

  point_in_time_recovery_enabled = true
  server_side_encryption_enabled = true
}

# ─────────────────────────────────────────────
# T012 — CloudWatch Log Group
# Created explicitly before Lambda to enforce the 14-day retention policy.
# Without this, Lambda auto-creates the log group with infinite retention.
# ─────────────────────────────────────────────

resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${local.name_prefix}-api"
  retention_in_days = 14

  tags = {
    project     = var.project_name
    environment = var.environment
    managed-by  = "terraform"
  }
}

# ─────────────────────────────────────────────
# T013 — Lambda Function
# Runtime: provided.al2023 (Go custom runtime), arm64 (Graviton2).
# Pre-built binary zip is uploaded to the artifact bucket by the module.
# ─────────────────────────────────────────────

module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "~> 8.0"

  function_name = "${local.name_prefix}-api"
  description   = "Cocktails API backend (Go, arm64)"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]

  memory_size = 256
  timeout     = 30

  # Use a pre-built zip deployed directly from the local filesystem.
  create_package         = false
  local_existing_package = var.lambda_binary_path

  # Use the explicit log group created above (14-day retention enforced).
  use_existing_cloudwatch_log_group = true
  logging_log_group                 = aws_cloudwatch_log_group.lambda_logs.name

  environment_variables = {
    STORE_BACKEND   = "dynamodb"
    RECIPES_TABLE   = module.recipes_table.dynamodb_table_id
    USERS_TABLE     = module.users_table.dynamodb_table_id
    FAVORITES_TABLE = module.favorites_table.dynamodb_table_id
    JWT_SECRET      = var.jwt_secret
    # Session lifetime; app falls back to 24h if unset. Override per-environment
    # (e.g. a short value on a preview) to exercise the auto-logout flow.
    JWT_TOKEN_DURATION = var.jwt_token_duration
    # Feature 026 — transactional email (password recovery)
    MAIL_FROM    = local.ses_sender_address
    APP_BASE_URL = "https://${var.domain_name}"
    # Feature 027 — backend logging: production defaults to warnings + errors.
    LOG_LEVEL = "warn"
  }

  # IAM: least-privilege access to DynamoDB tables and CloudWatch logs.
  create_role              = true
  attach_policy_statements = true

  policy_statements = {
    dynamodb = {
      effect = "Allow"
      actions = [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:BatchGetItem",
        "dynamodb:BatchWriteItem",
      ]
      resources = [
        module.recipes_table.dynamodb_table_arn,
        "${module.recipes_table.dynamodb_table_arn}/index/*",
        module.users_table.dynamodb_table_arn,
        "${module.users_table.dynamodb_table_arn}/index/*",
        module.favorites_table.dynamodb_table_arn,
        "${module.favorites_table.dynamodb_table_arn}/index/*",
      ]
    }

    cloudwatch_logs = {
      effect = "Allow"
      actions = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
      resources = [
        aws_cloudwatch_log_group.lambda_logs.arn,
        "${aws_cloudwatch_log_group.lambda_logs.arn}:*",
      ]
    }

    # Feature 026 — send password-recovery email via SES, scoped to our identity.
    ses = {
      effect = "Allow"
      actions = [
        "ses:SendEmail",
        "ses:SendRawEmail",
      ]
      resources = [aws_sesv2_email_identity.domain.arn]
    }
  }

  depends_on = [aws_cloudwatch_log_group.lambda_logs]
}

# Lambda permission for API Gateway — separate to avoid circular dependency between
# the lambda and api_gateway modules (each references the other's output).
resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_function.lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.api_gateway.api_execution_arn}/*/*"
}

# ─────────────────────────────────────────────
# T014 — HTTP API Gateway (v2)
# Single $default route proxies all requests to the Lambda function.
# ─────────────────────────────────────────────

module "api_gateway" {
  source  = "terraform-aws-modules/apigateway-v2/aws"
  version = "~> 6.0"

  name          = "${local.name_prefix}-api"
  protocol_type = "HTTP"

  create_domain_name = false

  cors_configuration = {
    allow_headers = ["content-type", "authorization"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_origins = ["*"]
    max_age       = 300
  }

  # v6.x uses 'routes' (not 'integrations'). Each route embeds its integration config.
  routes = {
    # Feature 026 — coarse gateway throttle on the password-recovery request
    # endpoint to blunt unauthenticated blanket abuse of the email-resolving
    # Scan (mitigation M2), independent of the per-user 6/hour application limit.
    # ~1 req/s steady with a burst of 5 is far above legitimate use.
    "POST /api/v1/auth/forgot-password" = {
      throttling_rate_limit  = 1
      throttling_burst_limit = 5
      integration = {
        uri                    = module.lambda_function.lambda_function_arn
        type                   = "AWS_PROXY"
        payload_format_version = "2.0"
      }
    }

    "$default" = {
      integration = {
        uri                    = module.lambda_function.lambda_function_arn
        type                   = "AWS_PROXY"
        payload_format_version = "2.0"
      }
    }
  }
}

# ─────────────────────────────────────────────
# T015 — Frontend Origin S3 Bucket
# Private bucket — only CloudFront OAC can read objects (see T017).
# ─────────────────────────────────────────────

module "frontend_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "~> 5.0"

  bucket = "${local.name_prefix}-frontend"

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ─────────────────────────────────────────────
# T016 — CloudFront Distribution
# Serves the SPA with HTTPS; OAC-only access to the S3 origin.
# SPA routing: 404 responses from S3 are rewritten to index.html with HTTP 200.
# ─────────────────────────────────────────────

module "cdn" {
  source  = "terraform-aws-modules/cloudfront/aws"
  version = "~> 4.0"

  create_origin_access_control = true

  origin_access_control = {
    "s3_oac" = {
      description      = "CloudFront OAC for ${local.name_prefix} frontend bucket"
      origin_type      = "s3"
      signing_behavior = "always"
      signing_protocol = "sigv4"
    }
  }

  # v4.x uses 'origin' (singular) as a map of origin objects.
  origin = {
    frontend = {
      domain_name           = module.frontend_bucket.s3_bucket_bucket_regional_domain_name
      origin_access_control = "s3_oac"
    }
    api = {
      domain_name = replace(module.api_gateway.api_endpoint, "https://", "")
      custom_origin_config = {
        http_port              = 80
        https_port             = 443
        origin_protocol_policy = "https-only"
        origin_ssl_protocols   = ["TLSv1.2"]
      }
    }
  }

  default_root_object = "index.html"
  price_class         = "PriceClass_100"
  http_version        = "http2"

  default_cache_behavior = {
    target_origin_id       = "frontend"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    use_forwarded_values   = false
    cache_policy_id        = "658327ea-f89d-4fab-a63d-7e88639e58f6" # CachingOptimized (AWS managed)

    function_association = {
      viewer-request = {
        function_arn = aws_cloudfront_function.spa_pr_routing.arn
      }
    }
  }

  ordered_cache_behavior = [
    {
      path_pattern             = "/api/*"
      target_origin_id         = "api"
      viewer_protocol_policy   = "redirect-to-https"
      allowed_methods          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods           = ["GET", "HEAD"]
      use_forwarded_values     = false
      cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingDisabled (AWS managed)
      origin_request_policy_id = "b689b0a8-53d0-40ab-baf2-68738e2966ac" # AllViewerExceptHostHeader (AWS managed)
    },
    {
      # Preview API traffic: /pr-<n>/api/* is forwarded to the shared API Gateway
      # origin. This behavior is matched before the default (S3) behavior, so the
      # SPA viewer-request function never rewrites preview API calls to index.html.
      path_pattern             = "/pr-*/api/*"
      target_origin_id         = "api"
      viewer_protocol_policy   = "redirect-to-https"
      allowed_methods          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods           = ["GET", "HEAD"]
      use_forwarded_values     = false
      cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingDisabled (AWS managed)
      origin_request_policy_id = "b689b0a8-53d0-40ab-baf2-68738e2966ac" # AllViewerExceptHostHeader (AWS managed)
    }
  ]

  # Missing frontend objects come back from S3 as 403 (OAC grants GetObject only,
  # not ListBucket) — including torn-down preview paths and any unknown path.
  # Serve the branded /404.html with a real 404 status. Origin 404s (from the API)
  # are intentionally NOT remapped, so API JSON error responses pass through.
  custom_error_response = [
    {
      error_code            = 403
      response_code         = 404
      response_page_path    = "/404.html"
      error_caching_min_ttl = 0
    },
  ]

  aliases = [var.domain_name]

  viewer_certificate = {
    acm_certificate_arn      = aws_acm_certificate_validation.cert.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  tags = {
    project     = var.project_name
    environment = var.environment
    managed-by  = "terraform"
  }
}

# ─────────────────────────────────────────────
# T017 — Frontend S3 Bucket Policy (CloudFront OAC)
# Grants the CloudFront distribution read-only access to the frontend bucket.
# The AWS:SourceArn condition restricts access to this specific distribution.
# ─────────────────────────────────────────────

data "aws_iam_policy_document" "frontend_oac" {
  statement {
    sid    = "AllowCloudFrontServicePrincipal"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    actions   = ["s3:GetObject"]
    resources = ["${module.frontend_bucket.s3_bucket_arn}/*"]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [module.cdn.cloudfront_distribution_arn]
    }
  }
}

resource "aws_s3_bucket_policy" "frontend_oac" {
  bucket = module.frontend_bucket.s3_bucket_id
  policy = data.aws_iam_policy_document.frontend_oac.json
}

# ─────────────────────────────────────────────
# T018 — Frontend Asset Upload
# Uploads all files from frontend/dist/ to the frontend S3 bucket.
# Prerequisites: operator must run `cd frontend && npm run build` first.
# ─────────────────────────────────────────────

resource "aws_s3_object" "frontend_assets" {
  for_each = fileset(local.frontend_dist_path, "**")

  bucket = module.frontend_bucket.s3_bucket_id
  key    = each.value
  source = "${local.frontend_dist_path}/${each.value}"
  etag   = filemd5("${local.frontend_dist_path}/${each.value}")

  # Derive content-type from the file extension (last segment after splitting on ".")
  content_type = lookup(
    local.content_type_map,
    reverse(split(".", each.value))[0],
    "application/octet-stream"
  )
}

# ─────────────────────────────────────────────
# T019 — CloudFront Cache Invalidation
# Triggers a "/*" invalidation whenever any frontend asset ETag changes.
# Requires AWS CLI installed and credentials configured on the operator's machine.
# ─────────────────────────────────────────────

resource "null_resource" "cdn_invalidation" {
  triggers = {
    etags = md5(join(",", [for o in aws_s3_object.frontend_assets : o.etag]))
  }

  provisioner "local-exec" {
    command = "aws cloudfront create-invalidation --distribution-id ${module.cdn.cloudfront_distribution_id} --paths '/*'"
  }

  depends_on = [aws_s3_object.frontend_assets]
}

# ─────────────────────────────────────────────
# T021 — SAM Metadata (Phase 4 / US2)
# Allows SAM CLI to discover and locally invoke the Lambda function
# without deploying to AWS. Run: sam local start-api --hook-name terraform
# ─────────────────────────────────────────────

resource "null_resource" "sam_metadata_lambda_function" {
  triggers = {
    # SAM CLI reads these trigger keys from the terraform plan output to locate
    # the function source and built artifact for local emulation.
    resource_name        = "module.lambda_function.aws_lambda_function.this[0]"
    resource_type        = "ZIP_LAMBDA_FUNCTION"
    original_source_code = "${path.root}/../backend"
    built_output_path    = "${path.root}/../backend/bootstrap.zip"
  }

  depends_on = [module.lambda_function]
}

# ─────────────────────────────────────────────
# Feature 021 — PR Preview: CloudFront Function for SPA routing
# Rewrites /pr-{N} and /pr-{N}/deep/link to /pr-{N}/index.html so the SPA
# loads correctly on first visit and on browser refresh at any sub-path.
# Paths with a file extension (assets) pass through unchanged.
# ─────────────────────────────────────────────

resource "aws_cloudfront_function" "spa_pr_routing" {
  name    = "${var.project_name}-spa-pr-routing"
  runtime = "cloudfront-js-2.0"
  comment = "Rewrite PR preview paths to index.html for SPA routing"
  publish = true
  code    = file("${path.module}/cloudfront-function-spa.js")
}

# ─────────────────────────────────────────────
# Feature 021 — PR Preview: Shared Lambda Execution Role
# All preview Lambdas share this role. DynamoDB access is scoped to
# cocktails-pr-* table ARNs; CloudWatch Logs scoped to /aws/lambda/cocktails-pr-*.
# ─────────────────────────────────────────────

data "aws_iam_policy_document" "preview_lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "preview_lambda_policy" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:UpdateItem",
      "dynamodb:DeleteItem", "dynamodb:Query", "dynamodb:Scan",
      "dynamodb:BatchGetItem", "dynamodb:BatchWriteItem",
    ]
    resources = ["arn:aws:dynamodb:*:*:table/${var.project_name}-pr-*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["arn:aws:logs:*:*:log-group:/aws/lambda/${var.project_name}-pr-*:*"]
  }

  # Feature 026 — preview Lambdas may also send password-recovery email via SES
  # (scoped to our verified identity). Previews without MAIL_FROM fall back to
  # the stub sender, so this permission is only exercised when SES is wired in.
  statement {
    effect    = "Allow"
    actions   = ["ses:SendEmail", "ses:SendRawEmail"]
    resources = [aws_sesv2_email_identity.domain.arn]
  }
}

resource "aws_iam_role" "preview_lambda" {
  name               = "${var.project_name}-preview-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.preview_lambda_assume.json

  tags = {
    project    = var.project_name
    managed-by = "terraform"
  }
}

resource "aws_iam_role_policy" "preview_lambda" {
  name   = "${var.project_name}-preview-lambda-policy"
  role   = aws_iam_role.preview_lambda.id
  policy = data.aws_iam_policy_document.preview_lambda_policy.json
}

# ─────────────────────────────────────────────
# Feature 012 — Custom Domain with HTTPS
# ACM certificate, Cloudflare DNS records, and CloudFront custom domain wiring.
# ─────────────────────────────────────────────

resource "aws_acm_certificate" "cert" {
  domain_name       = var.domain_name
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

resource "cloudflare_dns_record" "acm_validation" {
  for_each = {
    for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.domain_name => dvo
  }

  zone_id = var.cloudflare_zone_id
  name    = each.value.resource_record_name
  type    = each.value.resource_record_type
  content = each.value.resource_record_value
  proxied = false
  ttl     = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn         = aws_acm_certificate.cert.arn
  validation_record_fqdns = [for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.resource_record_name]
}

resource "cloudflare_dns_record" "routing" {
  zone_id = var.cloudflare_zone_id
  name    = "cocktails"
  type    = "CNAME"
  content = module.cdn.cloudfront_distribution_domain_name
  proxied = false
  ttl     = 300
}

# ─────────────────────────────────────────────
# Feature 026 — Transactional email (Amazon SES v2)
# Domain identity + Easy DKIM + custom MAIL FROM for the password-recovery
# emails. DKIM/SPF/MX records are published via the existing Cloudflare
# provider so the domain (and thus DMARC alignment) verifies automatically.
# NOTE: SES starts in the sandbox; production access must be requested
# separately (feature 026 task T029) before real users receive email.
# ─────────────────────────────────────────────

locals {
  # Sender address for password-recovery email; consumed by the Lambda as MAIL_FROM.
  ses_sender_address   = "no-reply@${var.domain_name}"
  ses_mail_from_domain = "mail.${var.domain_name}"
}

resource "aws_sesv2_email_identity" "domain" {
  email_identity = var.domain_name

  dkim_signing_attributes {
    next_signing_key_length = "RSA_2048_BIT"
  }
}

# Publish the three Easy DKIM CNAME tokens so SES can verify the domain.
resource "cloudflare_dns_record" "ses_dkim" {
  count = 3

  zone_id = var.cloudflare_zone_id
  name    = "${aws_sesv2_email_identity.domain.dkim_signing_attributes[0].tokens[count.index]}._domainkey.${var.domain_name}"
  type    = "CNAME"
  content = "${aws_sesv2_email_identity.domain.dkim_signing_attributes[0].tokens[count.index]}.dkim.amazonses.com"
  proxied = false
  ttl     = 300
}

# Custom MAIL FROM domain aligns the Return-Path with our domain (improves
# deliverability / DMARC). Falls back to the amazonses.com default on MX failure.
resource "aws_sesv2_email_identity_mail_from_attributes" "domain" {
  email_identity         = aws_sesv2_email_identity.domain.email_identity
  mail_from_domain       = local.ses_mail_from_domain
  behavior_on_mx_failure = "USE_DEFAULT_VALUE"
}

resource "cloudflare_dns_record" "ses_mail_from_mx" {
  zone_id  = var.cloudflare_zone_id
  name     = local.ses_mail_from_domain
  type     = "MX"
  content  = "feedback-smtp.${var.aws_region}.amazonses.com"
  priority = 10
  proxied  = false
  ttl      = 300
}

resource "cloudflare_dns_record" "ses_mail_from_spf" {
  zone_id = var.cloudflare_zone_id
  name    = local.ses_mail_from_domain
  type    = "TXT"
  content = "v=spf1 include:amazonses.com ~all"
  proxied = false
  ttl     = 300
}
