output "api_url" {
  description = "HTTPS endpoint of the API Gateway — use for all backend API requests"
  value       = module.api_gateway.api_endpoint
}

output "frontend_url" {
  description = "HTTPS URL of the CloudFront distribution serving the frontend SPA"
  value       = "https://${module.cdn.cloudfront_distribution_domain_name}"
}

output "recipes_table_name" {
  description = "DynamoDB table name for recipes"
  value       = aws_dynamodb_table.recipes.id
}

output "users_table_name" {
  description = "DynamoDB table name for users"
  value       = aws_dynamodb_table.users.id
}
