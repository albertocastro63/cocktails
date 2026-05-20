# Tasks: Custom Domain with HTTPS

**Input**: Design documents from `specs/012-custom-domain-https/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, quickstart.md ✅

**Organization**: All tasks are pure Terraform changes in `infra/`. US1 drives all implementation — US2, US3, and US4 are verified as side effects once the US1 infrastructure is in place and `terraform apply` completes.

---

## Phase 1: Setup

**Purpose**: Add the Cloudflare Terraform provider declaration so `terraform init` can download it.

- [ ] T001 Add `cloudflare/cloudflare ~> 5.0` to `required_providers` in `infra/versions.tf` and add a `provider "cloudflare"` block with `api_token = var.cloudflare_api_token`
- [ ] T002 Run `terraform init` from `infra/` to download the new Cloudflare provider plugin — verify no errors

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Declare the new input variables and the ACM certificate resource that all user story tasks depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T003 [P] Add `cloudflare_api_token` (sensitive string), `cloudflare_zone_id` (string), and `domain_name` (string, default `"cocktails.albertomcastro.com"`) variables to `infra/variables.tf`
- [ ] T004 [P] Add `resource "aws_acm_certificate" "cert"` to `infra/main.tf`: domain_name = `var.domain_name`, validation_method = `"DNS"`, with a `lifecycle { create_before_destroy = true }` block

**Checkpoint**: `terraform plan` shows the new ACM certificate resource. Variables are accessible to subsequent resources.

---

## Phase 3: User Story 1 — Access App via Custom Domain over HTTPS (Priority: P1) 🎯 MVP

**Goal**: `https://cocktails.albertomcastro.com` loads the cocktails app with a valid, browser-trusted TLS certificate. All four user stories (including US2 HTTP redirect, US3 DNS-as-code, US4 auto-renewal) converge on the same Terraform resources implemented in this phase.

**Independent Test**: After `terraform apply`, open `https://cocktails.albertomcastro.com` in a browser — the app loads with a valid padlock and no security warnings. `curl -sI http://cocktails.albertomcastro.com` returns a 301 redirect to the HTTPS URL.

### Implementation for User Story 1

- [ ] T005 [US1] Add `resource "cloudflare_dns_record" "acm_validation"` to `infra/main.tf` using `for_each` over `{ for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.domain_name => dvo }`: type = `"CNAME"`, name = `each.value.resource_record_name`, content = `each.value.resource_record_value`, zone_id = `var.cloudflare_zone_id`, proxied = `false` — note: Cloudflare provider v5 uses `content`, not `value`; this is the ACM DNS ownership proof record (US3: DNS-as-code; US4: kept in place for auto-renewal)
- [ ] T006 [US1] Add `resource "aws_acm_certificate_validation" "cert"` to `infra/main.tf`: certificate_arn = `aws_acm_certificate.cert.arn`, validation_record_fqdns = `[for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.resource_record_name]` — derives FQDNs directly from ACM (source of truth), avoiding reliance on a Cloudflare provider computed attribute that may not exist in v5; this blocks apply until ACM confirms the certificate is issued
- [ ] T007 [US1] Add `resource "cloudflare_dns_record" "routing"` to `infra/main.tf`: type = `"CNAME"`, name = `"cocktails"`, value = `module.cdn.cloudfront_distribution_domain_name`, zone_id = `var.cloudflare_zone_id`, proxied = `false` — routes user traffic to CloudFront (US3: DNS-as-code)
- [ ] T008 [US1] Update `module "cdn"` in `infra/main.tf`: add `aliases = [var.domain_name]` and replace the `viewer_certificate` block (`cloudfront_default_certificate = true`) with `{ acm_certificate_arn = aws_acm_certificate_validation.cert.certificate_arn, ssl_support_method = "sni-only", minimum_protocol_version = "TLSv1.2_2021" }`
- [ ] T009 [US1] Update `ordered_cache_behavior[0]` in `infra/main.tf`: change `viewer_protocol_policy` from `"https-only"` to `"redirect-to-https"` — required by FR-006 (both behaviors must redirect HTTP to HTTPS, not return 403)

**Checkpoint**: `terraform plan` shows: 1 ACM certificate created, 2 Cloudflare DNS records created, 1 ACM validation resource created, CloudFront distribution updated (aliases + viewer_certificate + API behavior redirect policy). No other resources changed.

---

## Phase 4: Polish & Verification

**Purpose**: Validate syntax, apply the changes, and verify all six success criteria from the spec.

- [ ] T010 Run `terraform validate && terraform fmt --check` from `infra/` — confirms provider schema compliance and HCL formatting before committing a 10–45 min apply
- [ ] T011 Run `terraform apply` from `infra/` with `cloudflare_api_token` and `cloudflare_zone_id` supplied (note: expect 10–45 min wait for ACM cert issuance)
- [ ] T012 Run quickstart.md verification scenarios 1–6 against the live infrastructure:
  - SC-001: `https://cocktails.albertomcastro.com` loads with valid cert (browser + curl)
  - SC-002: `curl -sI http://cocktails.albertomcastro.com` returns 301 → HTTPS (US2); `curl -sI http://cocktails.albertomcastro.com/api/v1/recipes` also returns 301 (FR-006 API path)
  - SC-003: `dig _*.cocktails.albertomcastro.com CNAME` resolves to `*.acm-validations.aws.` (US3/US4)
  - SC-004: ACM certificate status is `ISSUED` in AWS console or via `aws acm list-certificates`
  - SC-005: Custom domain response time ≤ baseline + 200 ms (curl timing comparison)
  - SC-006: Cloudflare dashboard shows both DNS records present with DNS-only proxy status (US3)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately; T001 then T002
- **Foundational (Phase 2)**: Depends on Phase 1 complete; T003 and T004 are parallel (different files)
- **US1 (Phase 3)**: Depends on Phase 2 complete; tasks T005→T006→T007→T008→T009 are sequential (same file: `main.tf`)
- **Polish (Phase 4)**: Depends on Phase 3 complete; T010→T011→T012 sequential

### Task Sequencing within Phase 3

T005 must precede T006 (certificate validation references ACM domain_validation_options).  
T006 must precede T008 (CloudFront viewer_certificate references `aws_acm_certificate_validation.cert.certificate_arn`).  
T007 and T009 can run in parallel with T005–T006 (different resources in same file — but serialized for simplicity).

### Parallel Opportunities

```bash
# Phase 2 — both run in parallel (different files):
T003  # variables.tf
T004  # main.tf (ACM cert resource)

# Phase 3 — sequential (all edit main.tf):
T005 → T006 → T007 → T008 → T009

# Phase 4 — sequential:
T010 (validate) → T011 (apply) → T012 (verify)
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1: Setup (T001–T002)
2. Complete Phase 2: Foundational (T003–T004)
3. Complete Phase 3: US1 (T005–T009)
4. **STOP and VALIDATE**: `terraform plan` shows exactly the expected diff before applying
5. Complete Phase 4: Validate + Apply + verify all 6 scenarios (T010–T012)

### Incremental Delivery

This feature is a single atomic infrastructure change — all resources must be applied together for the custom domain to work. There is no partial delivery milestone (you can't test HTTPS on the custom domain until cert + DNS + CloudFront alias are all in place).

---

## Notes

- `aws_acm_certificate_validation` causes Terraform to wait (up to 45 min) at apply time until ACM confirms cert issuance — this is expected and required. Do not interrupt the apply.
- US2 (HTTP → HTTPS redirect): The default behavior already uses `redirect-to-https` (T008 wires the alias + cert). T009 additionally fixes the API cache behavior from `https-only` to `redirect-to-https` to fully satisfy FR-006 for both behaviors.
- Both Cloudflare DNS records must be `proxied = false` (DNS-only / grey cloud). Proxied mode would break ACM validation and SNI matching.
- `cloudflare_api_token` is sensitive — supply via `TF_VAR_cloudflare_api_token` or `-var` flag, never in tfvars files committed to VCS.
- The `cloudflare_dns_record.routing` value should use `module.cdn.cloudfront_distribution_domain_name` (Terraform output) rather than the hardcoded domain, so it updates automatically if the distribution is recreated.
