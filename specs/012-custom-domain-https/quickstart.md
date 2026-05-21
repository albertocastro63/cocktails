# Quickstart: Custom Domain with HTTPS

**Feature**: 012-custom-domain-https  
**Date**: 2026-05-20

---

## Prerequisites

Before running `terraform apply`:

1. **Cloudflare API Token**: Generate a scoped token in the Cloudflare dashboard:
   - Permissions: `Zone → DNS → Edit` for zone `albertomcastro.com`
   - Copy the token — it is shown only once

2. **Cloudflare Zone ID**: Find in Cloudflare dashboard → `albertomcastro.com` → Overview → right-hand sidebar → "Zone ID"

3. **No conflicting DNS records**: Verify there is no existing `cocktails` CNAME or A record in Cloudflare for `albertomcastro.com`. If one exists, either delete it manually or import it into Terraform state before applying.

---

## Apply

```bash
cd infra

terraform init  # fetches the new cloudflare provider

terraform plan \
  -var="cloudflare_api_token=<YOUR_TOKEN>" \
  -var="cloudflare_zone_id=<YOUR_ZONE_ID>" \
  -var="jwt_secret=<JWT_SECRET>" \
  -var="lambda_binary_path=<BINARY_PATH>"

terraform apply \
  -var="cloudflare_api_token=<YOUR_TOKEN>" \
  -var="cloudflare_zone_id=<YOUR_ZONE_ID>" \
  -var="jwt_secret=<JWT_SECRET>" \
  -var="lambda_binary_path=<BINARY_PATH>"
```

**Expected apply time**: 10–45 minutes. The `aws_acm_certificate_validation` resource waits until ACM validates the DNS record and issues the certificate. ACM typically validates within 5–10 minutes once the Cloudflare DNS record is in place, but can take up to 45 minutes.

---

## Verification Scenarios

### Scenario 1: HTTPS loads the app (SC-001)

```bash
curl -s -o /dev/null -w "%{http_code}" https://cocktails.albertomcastro.com
# Expected: 200

curl -vI https://cocktails.albertomcastro.com 2>&1 | grep -E "subject:|issuer:|SSL"
# Expected: subject: CN=cocktails.albertomcastro.com
```

Open `https://cocktails.albertomcastro.com` in a browser. The padlock should be green with no security warnings. The certificate should be issued to `cocktails.albertomcastro.com`.

### Scenario 2: HTTP redirects to HTTPS (SC-002)

```bash
curl -sI http://cocktails.albertomcastro.com
# Expected: HTTP/1.1 301 Moved Permanently
# Location: https://cocktails.albertomcastro.com/
```

### Scenario 3: ACM validation CNAME present in DNS (SC-003)

After apply, find the validation CNAME name from the AWS console or Terraform state:
```bash
cd infra && terraform show | grep -A5 "acm_validation"
```

Then verify it resolves:
```bash
dig _<value>.cocktails.albertomcastro.com CNAME +short
# Expected: _<value>.acm-validations.aws.
```

### Scenario 4: Certificate status is Issued (SC-004)

```bash
aws acm list-certificates --region us-east-1 \
  --query "CertificateSummaryList[?DomainName=='cocktails.albertomcastro.com']"
# Expected: Status: ISSUED
```

### Scenario 5: Response time within 200 ms of baseline (SC-005)

```bash
# Baseline (CloudFront default domain)
curl -s -o /dev/null -w "%{time_total}" https://d3uf8lx6eccqzj.cloudfront.net

# Custom domain
curl -s -o /dev/null -w "%{time_total}" https://cocktails.albertomcastro.com
# Expected: custom domain time ≤ baseline + 200 ms
```

### Scenario 6: Single terraform apply creates everything (SC-006)

Starting from a clean state (no records, no cert):
```bash
terraform destroy  # if needed to reset
terraform apply -var="..."
# Expected: apply completes, both DNS records visible in Cloudflare dashboard,
#           cert status ISSUED, custom domain loads the app
```

---

## Cleanup

```bash
terraform destroy \
  -var="cloudflare_api_token=<YOUR_TOKEN>" \
  -var="cloudflare_zone_id=<YOUR_ZONE_ID>" \
  -var="jwt_secret=<JWT_SECRET>" \
  -var="lambda_binary_path=<BINARY_PATH>"
```

Both Cloudflare DNS records are removed automatically. The ACM certificate is deleted. The CloudFront distribution reverts to the default `*.cloudfront.net` certificate (note: `terraform destroy` destroys all resources including the rest of the app stack).
