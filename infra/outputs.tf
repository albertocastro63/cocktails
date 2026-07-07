output "api_url" {
  description = "HTTPS endpoint of the API Gateway — use for all backend API requests"
  value       = module.api_gateway.api_endpoint
}

output "frontend_url" {
  description = "HTTPS URL of the CloudFront distribution serving the frontend SPA"
  value       = "https://${module.cdn.cloudfront_distribution_domain_name}"
}

output "custom_domain_url" {
  description = "Custom domain URL for the app (available after terraform apply with Cloudflare credentials)"
  value       = "https://${var.domain_name}"
}

output "recipes_table_name" {
  description = "DynamoDB table name for recipes"
  value       = module.recipes_table.dynamodb_table_id
}

output "users_table_name" {
  description = "DynamoDB table name for users"
  value       = module.users_table.dynamodb_table_id
}

output "preview_lambda_role_arn" {
  description = "ARN of the shared IAM execution role for PR preview Lambda functions"
  value       = aws_iam_role.preview_lambda.arn
}

output "api_gateway_id" {
  description = "HTTP API Gateway ID — set as the API_GATEWAY_ID GitHub Actions variable for preview routing"
  value       = module.api_gateway.api_id
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID — set as the CLOUDFRONT_DISTRIBUTION_ID GitHub Actions variable"
  value       = module.cdn.cloudfront_distribution_id
}
