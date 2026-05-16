terraform {
  backend "s3" {
    # After running: cd infra/bootstrap && terraform init && terraform apply
    # replace the placeholder below with the state_bucket_name output value.
    # Then run: cd infra && terraform init
    bucket = "cocktails-tf-state-689595418365"

    key    = "cocktails/prod/terraform.tfstate"
    region = "us-east-1"

    # S3 native file locking — no DynamoDB table required (Terraform >= 1.10).
    # Terraform writes a .tflock file to the bucket when a lock is acquired.
    use_lockfile = true
  }
}
