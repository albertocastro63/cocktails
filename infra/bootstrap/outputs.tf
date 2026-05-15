output "state_bucket_name" {
  description = "Name of the S3 bucket created for Terraform remote state storage. Copy this value into infra/backend.tf before running terraform init in infra/."
  value       = aws_s3_bucket.state.bucket
}
