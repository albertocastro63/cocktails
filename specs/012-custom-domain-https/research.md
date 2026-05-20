# Research: Custom Domain with HTTPS

**Feature**: 012-custom-domain-https  
**Date**: 2026-05-20

---

## Cloudflare Terraform Provider

**Decision**: Use `cloudflare/cloudflare` provider version `~> 5.0`  
**Rationale**: Version 5.x is the current stable major release with a stable schema. The `~> 5.0` constraint allows patch/minor upgrades while pinning the major.  
**Alternatives considered**: v4 (older, fewer features), data source lookup for Zone ID (rejected per clarification — static variable preferred).

Key resources needed:
- `cloudflare_dns_record` — creates CNAME records in a Cloudflare zone
- Credentials: `CLOUDFLARE_API_TOKEN` environment variable or `api_token` provider argument (sensitive variable)

---

## ACM Certificate for CloudFront

**Decision**: Use `aws_acm_certificate` with `dns` validation in `us-east-1`  
**Rationale**: CloudFront requires ACM certificates to be in `us-east-1` regardless of where other resources reside. DNS validation is automated (no email interaction, auto-renews) as long as the CNAME record remains in Cloudflare.  
**Pattern**:
1. `aws_acm_certificate` requests the cert and produces `domain_validation_options` (a set of CNAME records to add)
2. `cloudflare_dns_record` creates the validation CNAME derived from `domain_validation_options`
3. `aws_acm_certificate_validation` blocks Terraform apply until ACM confirms issuance (waits up to 45 min by default)

**Key consideration**: The `domain_validation_options` set must be iterated with `for_each` using `{ for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.domain_name => dvo }`. Both the `aws_acm_certificate_validation` `validation_record_fqdns` and the `cloudflare_dns_record` reference this map.

---

## CloudFront Module — Alias and Certificate Wiring

**Current state** (confirmed by reading `infra/main.tf`):
- `viewer_certificate = { cloudfront_default_certificate = true }` — uses the default `*.cloudfront.net` cert
- No `aliases` block — distribution does not accept the custom domain

**Required changes**:
- Add `aliases = ["cocktails.albertomcastro.com"]` to the `module "cdn"` block
- Replace `viewer_certificate` with:
  ```hcl
  viewer_certificate = {
    acm_certificate_arn      = aws_acm_certificate_validation.cert.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
  ```
- The `module "cdn"` block must depend (implicitly or explicitly) on `aws_acm_certificate_validation.cert` so that the cert is fully issued before the distribution is updated.

**Viewer protocol policies**:
- `default_cache_behavior.viewer_protocol_policy = "redirect-to-https"` ✅ already correct
- `ordered_cache_behavior[0].viewer_protocol_policy = "https-only"` ⚠ must be changed to `"redirect-to-https"` to comply with FR-006 (both behaviors must redirect HTTP to HTTPS, not reject with 403)

---

## Cloudflare DNS Records

Two records are needed:

| Record | Type | Name | Target | Proxy |
|--------|------|------|--------|-------|
| ACM validation | CNAME | `_xxxx.cocktails.albertomcastro.com` (from ACM) | `_xxxx.acm-validations.aws` (from ACM) | DNS-only (proxied = false) |
| Routing | CNAME | `cocktails` | `d3uf8lx6eccqzj.cloudfront.net` | DNS-only (proxied = false) |

Both records must be `proxied = false` (DNS-only / grey cloud). CloudFront SNI matching requires the request to reach CloudFront without Cloudflare proxying, and ACM validation fails with proxied records.

---

## Terraform State

No new state backend is needed. All new resources (`aws_acm_certificate`, `aws_acm_certificate_validation`, `cloudflare_dns_record.*`) are added to the existing state in `s3://cocktails-tf-state-689595418365`.

---

## Sensitive Variable Handling

- `cloudflare_api_token`: `sensitive = true`, never written to tfstate in plaintext (Terraform marks it sensitive), never committed to VCS
- Supply at apply time via `TF_VAR_cloudflare_api_token` environment variable or `-var` flag
- `cloudflare_zone_id`: Not sensitive (it's a public identifier from the Cloudflare dashboard), but kept as a variable for flexibility
