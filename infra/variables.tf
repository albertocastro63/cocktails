variable "project_name" {
  description = "Short identifier prefix applied to all resource names"
  type        = string
  default     = "cocktails"
}

variable "environment" {
  description = "Deployment environment label (e.g. prod, staging)"
  type        = string
  default     = "prod"
}

variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "lambda_binary_path" {
  description = "Path to the pre-compiled Go binary zip (GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda && zip bootstrap.zip bootstrap)"
  type        = string
}

variable "jwt_secret" {
  description = "JWT signing secret for the API backend"
  type        = string
  sensitive   = true
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token with DNS edit permissions for the albertomcastro.com zone"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID for albertomcastro.com (static value from Cloudflare dashboard → Zone Overview)"
  type        = string
}

variable "domain_name" {
  description = "Custom domain name to attach to the CloudFront distribution"
  type        = string
  default     = "cocktails.albertomcastro.com"
}
