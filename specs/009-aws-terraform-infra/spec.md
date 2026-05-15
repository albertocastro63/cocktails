# Feature Specification: AWS Infrastructure Deployment

**Feature Branch**: `009-aws-terraform-infra`
**Created**: 2026-05-15
**Status**: Draft
**Input**: User description: "Deploy all application resources to AWS using Terraform with SAM support, targeting us-east-1, using serverless.tf modules where applicable."

## Clarifications

### Session 2026-05-15

- Q: Should the Terraform code deploy the frontend static assets alongside the backend API, or is this deployment backend-only? → A: Deploy both — backend API (Lambda + API Gateway + DynamoDB) and frontend static site (S3 + CloudFront) in a single Terraform workspace.
- Q: How should the remote state backend (S3 bucket + DynamoDB lock table) be provisioned? → A: A separate bootstrap Terraform config (or shell script) in the repo provisions it as a one-time step before the first deploy.
- Q: What level of observability is required for the deployed functions? → A: Structured logs forwarded to a managed log group with a defined retention period (14 days); no additional tracing or metrics dashboard required.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Deploy Application to AWS (Priority: P1)

An operator with valid AWS credentials runs a single command to provision all cloud infrastructure required to serve the cocktails application in AWS. After the command completes successfully, the application is reachable at a public URL.

**Why this priority**: Without a working deployment, all other infrastructure stories have no foundation. This is the minimum deliverable: a live application.

**Independent Test**: Run `terraform apply` against an empty AWS account. Confirm the command completes without errors, and that the application URL returned by the output responds with HTTP 200 to a recipe list request.

**Acceptance Scenarios**:

1. **Given** AWS credentials are configured on the local machine, **When** the operator runs the deploy command, **Then** all required cloud resources are created in `us-east-1`, the backend API is accessible at the API output URL, and the frontend is accessible at the CDN output URL.
2. **Given** the infrastructure was previously deployed, **When** the operator runs the deploy command again, **Then** the command succeeds idempotently (no duplicate resources, no errors).
3. **Given** the deploy command fails partway through, **When** the operator retries, **Then** the command recovers cleanly and reaches a consistent deployed state.

---

### User Story 2 — Test Infrastructure Locally Before Deploying (Priority: P2)

An operator uses SAM CLI integration to invoke the application's serverless functions locally against the Terraform-defined infrastructure configuration, catching bugs before incurring AWS costs.

**Why this priority**: Local testing with SAM reduces deploy-debug cycles. It is valuable but depends on US1 infrastructure definitions existing first.

**Independent Test**: Run `sam local start-api` (or equivalent SAM local command) against the Terraform project. Confirm that a recipe list request returns a valid response without hitting real AWS endpoints.

**Acceptance Scenarios**:

1. **Given** the Terraform project is initialized, **When** the operator starts the SAM local environment, **Then** the API functions respond to requests using local emulation.
2. **Given** the SAM local environment is running, **When** the operator makes a recipe CRUD request, **Then** the response is equivalent to what the deployed application would return.

---

### User Story 3 — Tear Down Infrastructure (Priority: P3)

An operator destroys all AWS resources created by the deployment in a single command, leaving the account clean with no orphaned resources and no ongoing charges.

**Why this priority**: Cost control and clean-up are essential for a dev/test deployment workflow. Lower priority than provisioning because it is the inverse of US1.

**Independent Test**: After a successful US1 deployment, run the destroy command. Confirm the application URL is no longer reachable and the AWS account contains no resources from this project.

**Acceptance Scenarios**:

1. **Given** the infrastructure is fully deployed, **When** the operator runs the destroy command, **Then** all project resources are removed from `us-east-1` and the account incurs no further charges.
2. **Given** the destroy command is run against an account with no deployed infrastructure, **Then** the command exits cleanly with no errors.

---

### Edge Cases

- What happens when the AWS credentials have insufficient permissions? → Deployment fails with a clear error message listing the missing permission.
- What happens when a resource already exists (e.g., S3 bucket name collision)? → The tool reports the conflict and the operator can resolve it without losing data.
- What happens when the deployment is interrupted mid-way (e.g., network loss)? → Re-running the deploy command recovers to a consistent state.
- What happens when `us-east-1` is experiencing a partial outage? → The failure is reported clearly; no partial resources are left in an inconsistent state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The infrastructure code MUST provision all resources required to run the cocktails application on AWS in the `us-east-1` region, including compute, storage, and networking.
- **FR-002**: The infrastructure code MUST support serverless compute (functions) as the primary runtime for the application backend.
- **FR-003**: The infrastructure code MUST provision a persistent data store for recipe and user data.
- **FR-004**: The infrastructure code MUST expose the application via a publicly accessible HTTP endpoint.
- **FR-005**: The infrastructure code MUST be organized so that deploying requires a single command after credentials are configured.
- **FR-006**: The infrastructure code MUST support SAM CLI local invocation for testing functions without deploying to AWS.
- **FR-007**: The infrastructure code MUST support tearing down all provisioned resources via a single command.
- **FR-008**: The infrastructure MUST use reusable, community-maintained modules where they cover the required functionality, to reduce maintenance burden.
- **FR-009**: The infrastructure MUST store its state remotely so multiple operators can manage the same deployment safely.
- **FR-010**: The infrastructure code MUST produce an output containing the application's public URL after a successful deployment.
- **FR-011**: The infrastructure code MUST deploy the frontend static assets (HTML, JS, CSS) to a content delivery network and serve them at a publicly accessible URL in the same deployment as the backend API.
- **FR-012**: The repository MUST include a separate bootstrap configuration that provisions the remote state S3 bucket as a one-time step, distinct from the main deployment workspace. State locking MUST use the storage backend's native file lock mechanism — no additional locking service is required.
- **FR-013**: The infrastructure MUST configure structured log groups for all compute functions with a retention period of 14 days.

### Key Entities

- **Compute Function**: The serverless unit that handles HTTP requests (maps to the existing Go backend handler).
- **HTTP API**: The public-facing gateway that routes HTTP requests to compute functions.
- **Data Store**: The persistence layer for recipes and users (DynamoDB, consistent with the existing store backend).
- **Artifact Storage**: The location where the compiled function binary is uploaded before deployment.
- **Remote State Backend**: The storage location for infrastructure state shared across operators.
- **IAM Role**: The permission boundary that grants compute functions the minimum access needed to read/write the data store.
- **Frontend Distribution**: The content delivery network and origin storage that serves the static frontend assets globally.
- **Bootstrap Configuration**: A separate, one-time Terraform config that creates the remote state S3 bucket before the main workspace can be initialised. State locking is handled by the S3 backend's native file lock (Terraform ≥ 1.10); no additional resource is required.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A full deploy completes in under 10 minutes on a standard internet connection.
- **SC-002**: The deployed application responds to a recipe list request within 3 seconds on cold start. Warm (non-cold-start) invocations MUST complete at p95 ≤ 200 ms per the project performance baseline.
- **SC-003**: A full destroy leaves zero project-tagged resources in the AWS account, verifiable via the AWS console.
- **SC-004**: An operator unfamiliar with the project can deploy from scratch by following the README instructions in under 30 minutes.
- **SC-005**: SAM local test environment starts in under 60 seconds and handles at least one successful API request before connecting to AWS.
- **SC-006**: After a failed API request, the corresponding error log entry is visible in the managed log group within 60 seconds, with sufficient detail to identify the failure cause.

## Assumptions

- AWS credentials are already configured on the operator's local machine (e.g., via `~/.aws/credentials`, environment variables, or an SSO profile); credential setup is out of scope.
- The target AWS account has sufficient service limits to provision the required resources (Lambda, DynamoDB, API Gateway, S3).
- The application backend is compiled to a Linux/amd64 binary (or arm64) before the deploy command is run; the infrastructure code does not build the binary itself.
- The DynamoDB backend (`STORE_BACKEND=dynamodb`) is the deployment target; SQLite is only for local development.
- Remote state storage (S3 bucket for Terraform state) is provisioned by a dedicated bootstrap configuration in the repo; operators run this once before the first deploy. State locking uses the S3 backend's native file lock — no DynamoDB table is needed.
- A single deployment environment (no separate staging/production separation) is in scope for this feature; multi-environment support is out of scope.
- The frontend is served as a static site; its hosting (e.g., S3 + CloudFront) is in scope alongside the backend API.
- SAM CLI must be installed on the operator's machine; installation instructions may be referenced but are not automated.
