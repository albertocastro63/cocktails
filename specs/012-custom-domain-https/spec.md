# Feature Specification: Custom Domain with HTTPS for Cocktails App

**Feature Branch**: `012-custom-domain-https`
**Created**: 2026-05-20
**Status**: Draft
**Input**: Configure cocktails.albertomcastro.com to point to the CloudFront distribution with HTTPS via ACM certificate

## Clarifications

### Session 2026-05-20

- Q: Does the HTTP → HTTPS redirect (FR-006) apply to both the frontend (`/*`) and API (`/api/*`) CloudFront cache behaviors, or only the frontend? → A: Both behaviors — frontend and API — must redirect HTTP to HTTPS.
- Q: How should the Cloudflare Zone ID be supplied to Terraform — as a direct variable or looked up via data source? → A: Supplied directly as a Terraform variable (static value from the Cloudflare dashboard).

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Access App via Custom Domain over HTTPS (Priority: P1)

A visitor navigates to `https://cocktails.albertomcastro.com` in their browser and the cocktails application loads normally — same content, same speed, with a valid padlock in the address bar.

**Why this priority**: This is the core deliverable. Without a working HTTPS custom domain, all other stories are moot. Delivers immediate value: a memorable, professional URL for the app.

**Independent Test**: Open `https://cocktails.albertomcastro.com` in a browser. The page loads, the browser shows a valid SSL certificate issued to `cocktails.albertomcastro.com`, and the certificate is trusted without any security warnings.

**Acceptance Scenarios**:

1. **Given** the DNS and CloudFront configuration is in place, **When** a user visits `https://cocktails.albertomcastro.com`, **Then** the cocktails app loads without any SSL/TLS errors or certificate warnings.
2. **Given** the CloudFront distribution has the custom domain configured, **When** the SSL certificate details are inspected, **Then** the certificate subject shows `cocktails.albertomcastro.com` and is issued by a trusted CA.
3. **Given** the app is live, **When** the response is checked, **Then** it returns HTTP 200 and the full app HTML from the existing CloudFront distribution.

---

### User Story 2 — HTTP Requests Automatically Redirect to HTTPS (Priority: P2)

A visitor (or a link) pointing to `http://cocktails.albertomcastro.com` (plain HTTP) is automatically and transparently redirected to the secure `https://` version without any user intervention.

**Why this priority**: Ensures no traffic is served over unencrypted HTTP. Protects users and prevents mixed-content warnings. Standard security baseline.

**Independent Test**: Make an HTTP request to `http://cocktails.albertomcastro.com`. The response must be a redirect (301 or 302) to `https://cocktails.albertomcastro.com`, and following that redirect loads the app over HTTPS.

**Acceptance Scenarios**:

1. **Given** the viewer protocol policy is set to redirect HTTP to HTTPS, **When** a request is sent to `http://cocktails.albertomcastro.com`, **Then** the response is a redirect to `https://cocktails.albertomcastro.com`.
2. **Given** a redirect is in place, **When** the redirect is followed, **Then** the app loads over HTTPS with a valid certificate.

---

### User Story 3 — DNS Records Are Managed as Code (Priority: P2)

The two Cloudflare DNS records required for this feature (the ACM validation CNAME and the routing CNAME) are provisioned and managed by the same Terraform configuration that manages the rest of the infrastructure. Running `terraform apply` creates all records automatically; no manual steps in the Cloudflare dashboard are required.

**Why this priority**: Eliminates manual, error-prone DNS steps. Keeps infrastructure fully reproducible — a fresh `terraform apply` on a new environment sets up DNS, certificate, and CloudFront wiring in one operation. Ties closely to P1 (you can't test P1 without DNS being correct).

**Independent Test**: Starting from a clean state (no DNS records for `cocktails.albertomcastro.com`), run `terraform apply`. Both Cloudflare DNS records appear in the Cloudflare dashboard without any manual action. The ACM certificate transitions to "Issued" automatically.

**Acceptance Scenarios**:

1. **Given** valid Cloudflare API credentials are configured, **When** `terraform apply` is run, **Then** the ACM validation CNAME record is created in Cloudflare automatically with DNS-only mode.
2. **Given** `terraform apply` completes successfully, **When** the Cloudflare DNS dashboard is checked, **Then** the routing CNAME (`cocktails → <distribution>.cloudfront.net`) is present in DNS-only mode.
3. **Given** both DNS records are managed by Terraform, **When** `terraform destroy` is run, **Then** both DNS records are removed from Cloudflare cleanly.

---

### User Story 4 — Certificate Auto-Renews Without Manual Intervention (Priority: P3)

The SSL certificate for `cocktails.albertomcastro.com` automatically renews before expiry. The site owner is never required to manually rotate or reissue the certificate.

**Why this priority**: Prevents the site from going down due to an expired certificate. Maintenance-free once set up, reducing operational burden.

**Independent Test**: Confirm that the ACM DNS validation CNAME record (`_xxxxx.cocktails.albertomcastro.com`) is present in Cloudflare with DNS-only mode (not proxied). ACM uses this record for automated renewal; its presence guarantees renewals will succeed indefinitely.

**Acceptance Scenarios**:

1. **Given** the ACM validation CNAME is managed by Terraform and present in Cloudflare in DNS-only mode, **When** ACM performs its renewal check, **Then** the certificate status remains "Issued" and does not expire.
2. **Given** the certificate was issued with DNS validation, **When** the certificate approaches its expiry date, **Then** ACM automatically issues a renewal without any manual action.

---

### Edge Cases

- What happens if the Cloudflare CNAME for `cocktails` is accidentally set to "Proxied" (orange cloud) instead of "DNS only"? CloudFront will reject the request because its SNI will not match. The site will fail with a certificate error. Since the record is managed by Terraform, the fix is to update the Terraform resource to DNS-only and re-apply.
- What happens if the Cloudflare API token is revoked or expires? Terraform will fail to manage DNS records on the next apply. A new token must be generated and the Terraform variable updated. The DNS records themselves remain in place and the site continues to function; only future Terraform operations are affected.
- What happens if the ACM validation CNAME is deleted from Cloudflare? ACM will eventually fail to renew the certificate and the cert will expire after its validity period (~13 months). The site will show a security error until the cert is reissued.
- What happens if the ACM certificate is requested in the wrong AWS region (not us-east-1)? CloudFront cannot attach it; the certificate will not appear in the CloudFront custom SSL dropdown. The fix is to request a new certificate in us-east-1.
- What happens if `cocktails.albertomcastro.com` is not added as an Alternate Domain Name (CNAME) in the CloudFront distribution? CloudFront will return a "Bad Request" (400) or SNI mismatch error for requests to the custom domain.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The application MUST be reachable at `cocktails.albertomcastro.com` via HTTPS with no browser security warnings.
- **FR-002**: A valid SSL/TLS certificate MUST be issued for `cocktails.albertomcastro.com` and presented to browsers automatically.
- **FR-003**: The SSL/TLS certificate MUST be validated via DNS (a CNAME record in Cloudflare).
- **FR-004**: The DNS validation CNAME record MUST be configured in Cloudflare as DNS-only (not proxied) to allow ACM to verify ownership.
- **FR-005**: `cocktails.albertomcastro.com` MUST be registered as an alternate domain name on the existing CloudFront distribution.
- **FR-006**: The CloudFront distribution MUST redirect all HTTP requests to HTTPS automatically — this applies to both the frontend cache behavior (`/*`) and the API cache behavior (`/api/*`).
- **FR-007**: The Cloudflare CNAME record pointing `cocktails` to the CloudFront domain MUST be configured as DNS-only (not proxied).
- **FR-008**: The SSL certificate MUST be capable of automatic renewal without manual intervention, as long as the DNS validation record remains in place.
- **FR-009**: Both Cloudflare DNS records (validation CNAME and routing CNAME) MUST be provisioned and managed by the project's infrastructure-as-code configuration — no manual DNS steps required.
- **FR-010**: A Cloudflare API token with DNS edit permissions for the `albertomcastro.com` zone MUST be the only credential required to manage DNS; the token MUST be supplied via a configuration variable and never stored in version control.
- **FR-011**: The Cloudflare Zone ID for `albertomcastro.com` MUST be supplied as a static Terraform variable; Terraform MUST NOT perform a live API lookup to resolve it.

### Key Entities

- **ACM Certificate**: Issued for `cocktails.albertomcastro.com` in us-east-1; validated via DNS; attached to the CloudFront distribution; auto-renews via the validation CNAME.
- **Cloudflare DNS Record (validation)**: A CNAME record generated by ACM (`_xxxxx.cocktails → _xxxxx.acm-validations.aws`); DNS-only; managed by infrastructure-as-code; must remain in place permanently for auto-renewal.
- **Cloudflare DNS Record (routing)**: A CNAME record pointing `cocktails → <distribution>.cloudfront.net`; DNS-only; managed by infrastructure-as-code; routes user traffic to CloudFront.
- **CloudFront Distribution**: The existing distribution; extended with the custom alternate domain name and the ACM certificate; viewer protocol policy set to redirect HTTP to HTTPS.
- **Cloudflare API Token**: A scoped credential granting DNS edit access to the `albertomcastro.com` zone; used exclusively by the infrastructure-as-code configuration; never committed to version control.
- **Cloudflare Zone ID**: The static identifier for the `albertomcastro.com` zone in Cloudflare; supplied as a Terraform variable; found in the Cloudflare dashboard under the zone's Overview page.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `https://cocktails.albertomcastro.com` loads the cocktails application with a valid, browser-trusted SSL certificate and no security warnings — verifiable in any major browser within 10 minutes of `terraform apply` completing.
- **SC-002**: `http://cocktails.albertomcastro.com` responds with an HTTP redirect (3xx) to the HTTPS URL — verifiable with a single curl command.
- **SC-003**: The ACM DNS validation CNAME record is present in Cloudflare in DNS-only mode — verifiable via DNS lookup (`dig _xxxxx.cocktails.albertomcastro.com CNAME`).
- **SC-004**: The ACM certificate status shows "Issued" in the certificate management console — confirming ownership validation succeeded.
- **SC-005**: The custom domain response time is within 200 ms of the baseline CloudFront URL response time — confirming no performance degradation from the domain change.
- **SC-006**: A single `terraform apply` (with valid credentials) creates all DNS records, provisions the certificate, and wires up the custom domain end-to-end — no manual steps required in the Cloudflare dashboard.

## Assumptions

- The existing CloudFront distribution (`d3uf8lx6eccqzj.cloudfront.net`) is already deployed and serving the cocktails app correctly.
- The domain `albertomcastro.com` is registered in Cloudflare; the operator has access to generate a scoped API token with DNS edit permissions for the zone.
- Cloudflare is used as a DNS resolver only — no Cloudflare proxy (no orange cloud) for either DNS record. Cloudflare features such as WAF, DDoS protection, and caching are not in scope for this feature.
- The ACM certificate will be provisioned in the `us-east-1` AWS region, which is a hard requirement for CloudFront-attached certificates.
- Infrastructure changes are managed via Terraform; the ACM certificate, CloudFront alternate domain name, and Cloudflare DNS records will all be added to the existing Terraform configuration.
- The Cloudflare Terraform provider will be added to the project; it requires the Cloudflare Zone ID for `albertomcastro.com` (a static string supplied directly as a Terraform variable) and a scoped API token — both supplied as Terraform variables and never committed to version control.
- The existing CloudFront viewer protocol policy may need updating to "redirect HTTP to HTTPS".
- Importing existing manually-created Cloudflare DNS records (if any) into Terraform state is out of scope; the assumption is no conflicting records exist for `cocktails.albertomcastro.com`.
