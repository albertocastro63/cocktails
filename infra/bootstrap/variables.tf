variable "project_name" {
  description = "Short identifier prefix for the state bucket name"
  type        = string
  default     = "cocktails"
}

variable "aws_region" {
  description = "AWS region where the state bucket is created"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment label applied to resource tags"
  type        = string
  default     = "bootstrap"
}
