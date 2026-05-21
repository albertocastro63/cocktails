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
