# Data Model: Custom Domain with HTTPS

**Feature**: 012-custom-domain-https  
**Date**: 2026-05-20

This feature adds no application-level data model changes. All entities are Terraform-managed infrastructure resources.

---

## Terraform Resource Inventory

### New Resources

| Resource | Name | Purpose |
|----------|------|---------|
| `aws_acm_certificate` | `cert` | Public TLS cert for `cocktails.albertomcastro.com`, DNS-validated, us-east-1 |
| `aws_acm_certificate_validation` | `cert` | Blocks apply until ACM confirms the cert is issued |
| `cloudflare_dns_record` | `acm_validation` | CNAME record for ACM DNS ownership proof; DNS-only |
| `cloudflare_dns_record` | `routing` | CNAME `cocktails → d3uf8lx6eccqzj.cloudfront.net`; DNS-only |

### Modified Resources

| Resource | Change |
|----------|--------|
| `module "cdn"` (CloudFront) | Add `aliases = ["cocktails.albertomcastro.com"]` |
| `module "cdn"` (CloudFront) | Replace `viewer_certificate.cloudfront_default_certificate = true` with `acm_certificate_arn`, `ssl_support_method = "sni-only"`, `minimum_protocol_version = "TLSv1.2_2021"` |

### New Variables

| Variable | Type | Sensitive | Source |
|----------|------|-----------|--------|
| `cloudflare_api_token` | `string` | yes | `TF_VAR_cloudflare_api_token` or `-var` at apply time |
| `cloudflare_zone_id` | `string` | no | Static value from Cloudflare dashboard → Zone Overview |
| `domain_name` | `string` | no | Default `"cocktails.albertomcastro.com"` |

### New Provider

| Provider | Source | Version |
|----------|--------|---------|
| `cloudflare/cloudflare` | `registry.terraform.io/cloudflare/cloudflare` | `~> 5.0` |

---

## Resource Relationships

```
aws_acm_certificate.cert
  └── domain_validation_options (set)
        └── cloudflare_dns_record.acm_validation  (DNS-only CNAME, validates ownership)
              └── aws_acm_certificate_validation.cert  (waits for issuance)
                    └── module.cdn  (viewer_certificate.acm_certificate_arn)

cloudflare_dns_record.routing  (DNS-only CNAME: cocktails → cloudfront.net)
  └── depends on module.cdn outputs (distribution domain name)
```

---

## Existing Resources (unchanged)

- `module.cdn` cache behaviors — no changes to viewer protocol policies (already correct)
- `module.frontend_bucket`, `module.recipes_table`, `module.users_table`, Lambda resources, API Gateway — all unchanged
