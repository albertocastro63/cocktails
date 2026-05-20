# Implementation Plan: Custom Domain with HTTPS

**Branch**: `012-custom-domain-https` | **Date**: 2026-05-20 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/012-custom-domain-https/spec.md`

## Summary

Extend the existing Terraform configuration to serve the cocktails app at `https://cocktails.albertomcastro.com`. This requires: (1) an ACM TLS certificate in `us-east-1` validated via DNS, (2) two Cloudflare DNS records managed by a new Cloudflare Terraform provider, and (3) wiring the certificate and custom domain alias into the existing CloudFront distribution.

## Technical Context

**Language/Version**: HCL / Terraform >= 1.10  
**Primary Dependencies**: AWS provider `~> 6.0`, Cloudflare provider `~> 5.0`, `terraform-aws-modules/cloudfront/aws ~> 4.0` (existing)  
**Storage**: Terraform state in S3 (`cocktails-tf-state-689595418365`) — no application storage changes  
**Testing**: Manual verification via `curl` and browser; quickstart.md scenarios  
**Target Platform**: AWS (us-east-1) + Cloudflare DNS  
**Project Type**: Infrastructure-as-code (Terraform)  
**Performance Goals**: Custom domain response time within 200 ms of baseline CloudFront URL (SC-005)  
**Constraints**: ACM certificate must be in `us-east-1` (hard CloudFront requirement); Cloudflare records must be DNS-only (not proxied); apply time ~10–45 min for cert issuance  
**Scale/Scope**: Single domain, two DNS records, one certificate, one CloudFront distribution

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ Each Terraform resource has a single purpose |
| II. Test-First | Are failing tests written before implementation begins? | ✅ N/A — infrastructure changes verified via manual scenarios in quickstart.md |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ N/A — no UI changes |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ SC-005 explicitly checks custom domain adds ≤ 200 ms overhead |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ N/A — Terraform changes; no backend/frontend code changes |

## Project Structure

### Documentation (this feature)

```text
specs/012-custom-domain-https/
├── plan.md              # This file
├── research.md          # Provider version, ACM pattern, Cloudflare DNS analysis
├── data-model.md        # Terraform resource inventory + new variables
├── quickstart.md        # Operator verification steps (6 scenarios)
└── checklists/
    └── requirements.md  # Spec quality checklist (all items pass)
```

### Source Code

```text
infra/
├── versions.tf          # ADD: cloudflare/cloudflare ~> 5.0 provider + cloudflare provider block
├── variables.tf         # ADD: cloudflare_api_token (sensitive), cloudflare_zone_id, domain_name
└── main.tf              # ADD: aws_acm_certificate, aws_acm_certificate_validation,
                         #      cloudflare_dns_record.acm_validation, cloudflare_dns_record.routing
                         # MODIFY: module "cdn" — add aliases, replace viewer_certificate
                         # MODIFY: ordered_cache_behavior[0] — viewer_protocol_policy https-only → redirect-to-https (FR-006)
```

No changes to `backend/`, `frontend/`, or any other directory.

**Structure Decision**: Single project. All changes confined to `infra/`. Three files modified/extended.

## Complexity Tracking

No constitution violations. All principles pass without justification needed.
