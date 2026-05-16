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
# T013 — Lambda Package Upload
# Uploads the pre-built bootstrap.zip to the artifact bucket so that
# the lambda module can reference it via s3_existing_package (v8 requires
# the upload and the function config to be separate resources).
# ─────────────────────────────────────────────

resource "aws_s3_object" "lambda_package" {
  bucket = module.artifact_bucket.s3_bucket_id
  key    = "lambda/bootstrap.zip"
  source = var.lambda_binary_path
  etag   = filemd5(var.lambda_binary_path)
}

# ─────────────────────────────────────────────
# T013 — Lambda Function
# Runtime: provided.al2023 (Go custom runtime), arm64 (Graviton2).
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

  # Reference the zip already uploaded above; v8 requires s3_existing_package
  # instead of local_existing_package + store_on_s3 (those two now conflict).
  create_package      = false
  s3_existing_package = {
    bucket     = module.artifact_bucket.s3_bucket_id
    key        = aws_s3_object.lambda_package.key
    version_id = aws_s3_object.lambda_package.version_id
  }

  # Use the explicit log group created above (14-day retention enforced).
  use_existing_cloudwatch_log_group = true
  logging_log_group                 = aws_cloudwatch_log_group.lambda_logs.name

  environment_variables = {
    STORE_BACKEND  = "dynamodb"
    RECIPES_TABLE  = module.recipes_table.dynamodb_table_id
    USERS_TABLE    = module.users_table.dynamodb_table_id
    JWT_SECRET     = var.jwt_secret
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

    cache_policy_id      = "658327ea-f89d-4fab-a63d-7e88639e58f6" # CachingOptimized (AWS managed)
    use_forwarded_values = false # required when cache_policy_id is set
  }

  custom_error_response = [
    {
      error_code            = 404
      response_code         = 200
      response_page_path    = "/index.html"
      error_caching_min_ttl = 0
    },
    {
      error_code            = 403
      response_code         = 200
      response_page_path    = "/index.html"
      error_caching_min_ttl = 0
    },
  ]

  viewer_certificate = {
    cloudfront_default_certificate = true
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
    resource_name       = "module.lambda_function.aws_lambda_function.this[0]"
    resource_type       = "ZIP_LAMBDA_FUNCTION"
    original_source_code = "${path.root}/../backend"
    built_output_path   = "${path.root}/../backend/bootstrap.zip"
  }

  depends_on = [module.lambda_function]
}
