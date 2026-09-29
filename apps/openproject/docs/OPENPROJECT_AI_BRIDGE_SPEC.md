# OpenProject AI Bridge — Technical & Product Specification

> **Status:** Handover-ready\
> **Target audience:** Cody / contributors\
> **Repository:** standalone GitHub repository\
> **Suggested repository name:** `openproject-ai-bridge`\
> **Initial deployment target:** existing self-hosted Infra stack\
> **License recommendation:** Apache-2.0 or MIT\
> **Project type:** independent, unofficial, open-source community integration\
> **Primary purpose:** connect AI clients and automation tools to OpenProject Community Edition through the public OpenProject API v3

---

## 1. Executive Summary

`openproject-ai-bridge` is an independent open-source integration layer between AI clients and OpenProject.

The bridge must **not** be tightly coupled to one project, one AI vendor, one OpenProject installation, or one infrastructure stack.

The intended architecture is:

```text
ChatGPT ─┐
Codex   ─┤
Cody    ─┼──── MCP / REST / CLI ────> OpenProject AI Bridge ────> OpenProject API v3
other AI ─┤
CI/CD   ─┤
CLI     ─┘
```

The first real-world installation will run inside the existing Infra stack and connect to the local OpenProject instance.

The first practical use cases will come from:

- Bewusstseins-Campus
- Corporate Book / further book projects
- workshops
- Sprach-A-Lyzer
- Attention-Hackrs
- Infra projects
- AVSoP modernization
- future standalone projects

However, **none of those domains belong in the bridge repository itself**.

The repository is a generic integration product.

---

# 2. Guiding Principles

The implementation SHALL follow these principles.

## 2.1 OpenProject remains the project system of record

The bridge does not replace OpenProject.

It translates controlled AI actions into normal OpenProject operations.

```text
AI client
    ↓
Bridge
    ↓
OpenProject
```

OpenProject remains authoritative for projects, work packages, statuses, priorities, versions/milestones, comments/journals, attachments, project membership and authorization.

## 2.2 Domain systems remain authoritative for their artifacts

OpenProject is the project steering and handover layer, but not automatically the permanent source of truth for every artifact.

| Artifact | Source of truth |
|---|---|
| software source code | Git repository |
| technical documentation after integration | Git repository |
| project state | OpenProject |
| decisions / blockers / risks | OpenProject |
| transient handover documents | OpenProject attachment |
| book manuscript | dedicated manuscript/document repository |
| workshop assets | dedicated document repository |
| conceptual idea before approval | AI conversation / idea pool |
| approved implementation concept | OpenProject + final target repository |

Rule:

> **OpenProject is the steering and handover system. The appropriate domain system remains the final source of truth.**

## 2.3 Do not synchronize conversations

The bridge SHALL NOT blindly write entire ChatGPT/Cody conversations into OpenProject.

Only durable project deltas should be persisted:

- confirmed decisions
- relevant progress
- milestones
- risks
- blockers
- approved ideas
- open questions
- next actions
- handover-ready artifacts

## 2.4 Human agency first

Default write behavior:

```text
READ_ONLY
ASK_BEFORE_WRITE   <- default
CONTROLLED_WRITE
FULL_PROJECT_ACCESS
```

`ASK_BEFORE_WRITE` SHALL be the recommended/default operational mode.

## 2.5 Vendor-neutral core

The core bridge SHALL NOT depend on ChatGPT-specific behavior.

Core:
- OpenProject API Client
- Domain Service
- Permission Layer
- Audit Layer
- Artifact/Handover Service
- Project Delta Service

Adapters:
- MCP
- REST
- CLI
- ChatGPT/Codex plugin package
- future clients

---

# 3. Scope

## 3.1 MVP Scope

The MVP SHOULD support:

### OpenProject connectivity
- instance discovery / health check
- authentication
- project/workspace discovery
- work package discovery
- work package creation
- work package update
- comments
- statuses
- priorities
- types
- versions / milestones
- relations where supported/needed
- attachments
- attachment download
- project delta submission
- audit trail

### AI-facing interfaces
- MCP server
- REST API
- CLI

### Artifact workflows
- upload Markdown artifact
- attach artifact to work package
- inspect/list attachments
- download artifact
- mark handover state
- register target repository/path
- record integration result

### Security
- per-instance credentials
- permission modes
- explicit destructive-action control
- secrets outside repository
- audit log
- webhook signature verification where used

## 3.2 Explicit Non-Goals for MVP

Do NOT initially implement:
- OpenProject UI replacement
- generic document management platform
- repository hosting
- Git synchronization engine
- autonomous project manager
- full OpenProject API coverage
- bypass/reimplementation of proprietary OpenProject Enterprise code
- cloning OpenProject's official MCP implementation
- automatic migration of entire chat histories
- automatic execution of every AI suggestion
- automatic deletion of work packages
- arbitrary shell execution
- arbitrary repository write access through the bridge
- user/password storage for OpenProject

---

# 4. Fairness & OpenProject Relationship

This project is an **unofficial community integration**.

README and public metadata SHALL include a statement equivalent to:

> `openproject-ai-bridge` is an independent open-source community project. It is not affiliated with, sponsored by, or endorsed by OpenProject GmbH.

Avoid:
- “Official OpenProject AI Plugin”
- official-looking OpenProject branding
- OpenProject logo as the bridge product logo
- claims that the bridge is OpenProject's own MCP implementation

The bridge uses public interfaces, primarily OpenProject API v3.

If the project reaches stable public quality, contributors SHOULD contact OpenProject and offer it for inclusion in their community integrations list.

---

# 5. Technology Recommendation

## 5.1 Language

Preferred: **Go**

Reasons:
- fit with existing ecosystem
- single binary
- low runtime footprint
- easy containerization
- excellent CLI support
- straightforward concurrency

A TypeScript MCP adapter is acceptable if the current official MCP SDK significantly reduces compatibility risk.

Preferred architecture if required:

```text
Go core/service
   +
small TypeScript MCP adapter
```

Before splitting languages, evaluate whether a stable Go MCP implementation meets requirements.

Keep the core protocol-independent.

---

# 6. High-Level Architecture

```text
┌────────────────────────────────────────────────────────────┐
│                         Clients                            │
│ ChatGPT | Codex | Cody | CLI | CI/CD | future AI clients  │
└──────────────────────────────┬─────────────────────────────┘
                               │
                 MCP / REST / CLI
                               │
                               ▼
┌────────────────────────────────────────────────────────────┐
│                  OpenProject AI Bridge                     │
│  Interface adapters: MCP | REST | CLI                      │
│  Application services:                                    │
│    Project / WorkPackage / Delta / Artifact / Handover     │
│    Mapping / Permission / Audit                            │
│  OpenProject API Client (HAL/HATEOAS aware)                │
└──────────────────────────────┬─────────────────────────────┘
                               │ HTTPS
                               ▼
┌────────────────────────────────────────────────────────────┐
│                    OpenProject API v3                      │
└────────────────────────────────────────────────────────────┘
```

---

# 7. OpenProject API Integration

## 7.1 API model

OpenProject API v3 is HAL/HATEOAS based.

The client SHOULD:
- respect `_links`
- avoid assuming every endpoint forever
- use schema/form endpoints where practical
- handle HAL collections correctly
- tolerate additive API fields
- validate server capabilities

## 7.2 API specification discovery

Useful runtime capability:

```text
GET /api/v3/spec.json
GET /api/v3/spec.yml
```

Optional bridge command:

```text
opai instance inspect
```

## 7.3 Work package creation

Prefer the current supported general/workspace route.

Important implementation note:

Older clients may use:

```text
POST /api/v3/projects/{id}/work_packages
```

Current documentation marks that project-scoped creation endpoint as deprecated in favor of:

```text
POST /api/v3/workspaces/{id}/work_packages
```

The client SHALL NOT make the deprecated project-scoped endpoint the architectural default.

A compatibility fallback MAY be implemented for older supported OpenProject installations.

## 7.4 Authentication

Supported bridge auth modes SHOULD include:

### Development / private single-instance MVP
API token as Bearer token:

```http
Authorization: Bearer <OPENPROJECT_API_TOKEN>
```

### Compatibility
API token via Basic Auth:

```text
username: apikey
password: <OPENPROJECT_API_TOKEN>
```

### Recommended public/multi-user mode
OAuth 2.0.

Architecture rule:

> Start MVP with API token support, but model credentials behind an interface so OAuth2 can be added without rewriting application services.

Example:

```go
type CredentialProvider interface {
    AuthorizationHeader(ctx context.Context) (string, error)
}
```

Implementations:
- APITokenProvider
- OAuth2Provider
- FutureExternalOIDCProvider

---

# 8. Identity & Agent Separation

Each automated actor SHOULD use a distinct OpenProject identity.

Example:

```text
ai-chatgpt
ai-cody
ai-ci
ai-local
```

Do not share one all-powerful administrator credential.

The bridge itself SHALL NOT require an OpenProject administrator account for normal project operations.

---

# 9. Permission Modes

## READ_ONLY
Allowed:
- list/read project data
- list/read work packages
- read comments
- read metadata
- list/download permitted attachments

Forbidden:
- all writes

## ASK_BEFORE_WRITE
Default.

Reads are allowed.

Write request returns a structured pending action where client/user confirmation is expected.

## CONTROLLED_WRITE
Allow a configured low-risk write allowlist.

Example allowed operations:
- add comment
- create idea
- create AI worklog
- update non-sensitive progress metadata

## FULL_PROJECT_ACCESS
Advanced opt-in only.

Still do not silently perform destructive operations.

---

# 10. Destructive Action Policy

MVP SHOULD NOT expose delete operations to AI tools.

Avoid generic tools such as:

```text
execute_openproject_request(method, path, payload)
```

Expose specific semantic tools instead.

---

# 11. MCP Tool Catalog

## Instance / discovery
```text
get_instance_info
list_projects
get_project
list_work_package_types
list_statuses
list_priorities
list_versions
```

## Work packages
```text
list_work_packages
get_work_package
create_work_package
update_work_package
add_work_package_comment
list_work_package_relations
create_work_package_relation
```

## Attachments / artifacts
```text
list_attachments
get_attachment_metadata
download_attachment
attach_artifact
```

Later:
```text
prepare_direct_upload
```

## Project Delta
```text
preview_project_delta
apply_project_delta
```

## Handover
```text
prepare_artifact_handover
get_handover
mark_handover_integrated
link_repository_artifact
```

---

# 12. MCP Tool Metadata

For public ChatGPT/Codex plugin submission, every MCP tool SHALL be annotated accurately.

At minimum consider:
- `readOnlyHint`
- `openWorldHint`
- `destructiveHint`

Never mark a write operation as read-only.

---

# 13. REST API

Suggested prefix:

```text
/api/v1
```

Possible routes:

```text
GET  /health
GET  /instances
GET  /instances/{id}/projects

POST /deltas/preview
POST /deltas/apply

POST /work-packages
PATCH /work-packages/{id}
POST /work-packages/{id}/comments

POST /work-packages/{id}/artifacts
GET  /work-packages/{id}/artifacts

POST /handovers
GET  /handovers/{id}
POST /handovers/{id}/integrated
```

The bridge REST API is semantic, versioned and stable.

---

# 14. CLI

Binary suggestion:

```text
opai
```

Examples:

```bash
opai instance test
opai project list
opai wp list --project corporate-book
opai wp get BOOK-030
opai delta preview project-delta.json
opai delta apply project-delta.json
opai artifact attach BOOK-030 docs/chapter-30-concept.md
opai handover inspect BOOK-030
opai handover integrated BOOK-030 --repo openproject-ai-bridge --path docs/architecture.md --commit abc1234
```

---

# 15. PROJECT DELTA Specification

Purpose:

> Capture durable changes resulting from a creative/AI work session without storing the whole conversation.

Delta categories:
- progress
- decision
- risk
- blocker
- open_question
- idea
- milestone
- next_step
- artifact
- handover

Example:

```json
{
  "schema_version": "1.0",
  "instance": "infra",
  "project": "corporate-book",
  "work_package": "BOOK-030",
  "source": {
    "type": "chatgpt",
    "actor": "ai-chatgpt",
    "session_ref": null
  },
  "summary": "Concept for chapter 30 refined",
  "changes": [
    {"type": "progress", "text": "Chapter concept defined"},
    {"type": "decision", "text": "Separate linguistic history from spiritual interpretation"},
    {"type": "open_question", "text": "Determine final chapter placement"},
    {"type": "next_step", "text": "Create full chapter draft"}
  ],
  "artifacts": [],
  "requested_updates": {
    "status": "in-progress"
  }
}
```

Processing:

```text
Receive delta
   ↓
Validate schema
   ↓
Resolve instance/project/work package
   ↓
Check permissions
   ↓
Generate preview
   ↓
Human confirmation (default)
   ↓
Apply
   ↓
Write OpenProject journal/comment
   ↓
Write bridge audit event
```

---

# 16. Work Package Mapping

Avoid hardcoding project-specific type IDs.

Use logical bridge types mapped per OpenProject instance/project.

Example:

```yaml
mapping:
  work_package_types:
    idea: "Idea"
    decision: "Decision"
    risk: "Risk"
    milestone: "Milestone"
    implementation: "Task"
    ai_worklog: "AI Worklog"
```

If target type does not exist:
- do not silently create admin-level configuration
- return a clear configuration error
- optionally fall back to configured generic Task type

---

# 17. Recommended OpenProject Work Package Types

For our first installation:

```text
Idea
Decision
Open Question
Risk
Experiment
Milestone
Implementation
AI Worklog
Artifact Handover
```

These are installation-specific recommendations, NOT requirements of the bridge.

---

# 18. Recommended Status Model

Suggested general statuses:

```text
New
Exploring
Planned
In Progress
Review
Blocked
Done
Discarded
```

Artifact state:

```text
DRAFT
REVIEW_READY
HANDOVER_READY
INTEGRATED
SUPERSEDED
```

Do not depend on OpenProject custom fields for MVP unless current API support is confirmed for the installed version.

---

# 19. AI Worklog

`AI Worklog` is an optional work package type.

Purpose:
- summarize meaningful AI-supported work sessions
- preserve decision context
- avoid raw chat transcript storage

Do not create one for every trivial operation.

---

# 20. Artifact Model

An artifact is a meaningful generated file.

Typical formats:

```text
.md
.txt
.json
.yaml
.pdf
.docx
.png
```

MVP priority:
```text
Markdown
JSON/YAML
```

Recommended metadata:

```yaml
title: "OpenProject AI Bridge – Architecture"
artifact_type: "concept"
source: "chatgpt"
status: "handover-ready"
target_repository: "openproject-ai-bridge"
target_path: "docs/architecture.md"
related_work_package: "OPAI-012"
```

---

# 21. Artifact Lifecycle

```text
DRAFT
   ↓
REVIEW_READY
   ↓
HANDOVER_READY
   ↓
INTEGRATED
   ↓
SUPERSEDED
```

Rules:
- DRAFT is not eligible for automatic repository integration.
- HANDOVER_READY explicitly signals that Cody or another implementation agent may take ownership.
- INTEGRATED records final repository/path/commit.
- OpenProject attachment remains historical evidence.
- Git becomes source of truth for integrated technical documentation.

---

# 22. OpenProject Attachment Integration

MVP SHALL support standard multipart upload to:

```text
POST /api/v3/work_packages/{id}/attachments
```

Bridge should support:
- upload
- list
- metadata
- download

OpenProject also supports a prepare/direct-upload workflow:

```text
POST /api/v3/work_packages/{id}/attachments/prepare
```

Do not require direct upload for MVP.

Fallback must remain ordinary multipart upload.

---

# 23. Artifact Handover Workflow

```text
ChatGPT / AI
   ↓
creates concept.md
   ↓
Bridge attaches file to OpenProject work package
   ↓
artifact status = HANDOVER_READY
   ↓
Cody reads attachment
   ↓
Cody evaluates repository context
   ↓
Cody adapts if necessary
   ↓
Cody writes file into target repo
   ↓
commit / PR
   ↓
Cody calls mark_handover_integrated
   ↓
OpenProject comment + bridge audit record
```

Cody MUST NOT blindly copy documentation.

Expected behavior:
1. read attachment
2. inspect repository
3. find existing related documentation
4. merge rather than duplicate where appropriate
5. adapt links, paths and technical details
6. preserve conceptual intent
7. write resulting documentation
8. test/validate if applicable
9. commit or prepare PR
10. report final repo path and commit/PR
11. mark handover integrated

---

# 24. Handover Record

```json
{
  "schema_version": "1.0",
  "work_package": "OPAI-012",
  "artifact": {
    "attachment_id": 773,
    "filename": "architecture.md",
    "status": "HANDOVER_READY"
  },
  "target": {
    "repository": "openproject-ai-bridge",
    "path": "docs/architecture.md"
  },
  "integration": {
    "status": "INTEGRATED",
    "commit": "abc1234",
    "pull_request": null
  }
}
```

---

# 25. Repository Linking

The bridge SHALL initially store repository references only as metadata/links.

It SHALL NOT require direct GitHub write access for MVP.

Future providers MAY include:
- GitHub
- GitLab
- Forgejo/Gitea

---

# 26. Webhooks

The bridge SHOULD support OpenProject webhook consumption after the basic write path works.

Recommended uses:
- invalidate caches
- update handover state
- notify workflows about changes
- prevent stale writes
- detect human changes performed directly in OpenProject

Webhook security:
- require configured signature secret
- verify signature
- reject invalid payloads
- protect against replay where practical
- log webhook event IDs/digests

Suggested endpoint:

```text
POST /webhooks/openproject/{instance}
```

---

# 27. Concurrency / Stale Update Protection

AI clients must not blindly overwrite newer human changes.

Bridge logic SHOULD:
1. fetch current work package
2. compare last-known state/version
3. apply only requested fields
4. reject conflict if target changed materially
5. return a structured conflict

Avoid full-object overwrite behavior.

---

# 28. Instance Configuration

Support multiple OpenProject instances even if MVP begins with one.

```yaml
instances:
  infra:
    base_url: "https://openproject.example.org"
    auth:
      type: "api_token"
      token_env: "OPENPROJECT_INFRA_TOKEN"
    default_permission_mode: "ask_before_write"
```

Never store plaintext secrets in committed config.

---

# 29. Secrets

MVP:
- environment variables
- Docker secrets

Never:
- committed `.env`
- API token in logs
- API token in OpenProject comments
- API token in AI tool responses

Provide `.env.example` with placeholders only.

---

# 30. Local State / Database

Recommended MVP: **SQLite**

Store only bridge-specific metadata:
- instance definitions without raw secrets
- handover state
- external IDs
- idempotency keys
- audit records
- cached mappings
- optional pending approvals

OpenProject remains authoritative project database.

---

# 31. Idempotency

Write operations SHOULD accept an idempotency key.

Repeated request:
- returns original result
- does not create duplicate work package/comment/artifact

Critical for:
- create_work_package
- apply_project_delta
- attach_artifact
- mark_handover_integrated

---

# 32. Audit Log

Every write operation SHALL create a bridge audit entry.

Fields:
- timestamp
- instance
- actor/client
- operation
- target
- request_id
- idempotency_key
- confirmation_mode
- result
- OpenProject resource ID
- error if any

Do not store secrets.

---

# 33. Observability

Implement:
- structured logs
- request IDs
- health endpoint
- readiness endpoint

```text
GET /healthz
GET /readyz
```

---

# 34. Rate Limiting

Requirements:
- exponential backoff for transient 429/5xx
- respect Retry-After if present
- bounded retry count
- no automatic retry on validation 4xx
- idempotency for retried writes

---

# 35. Error Model

Example:

```json
{
  "code": "OPENPROJECT_VALIDATION_ERROR",
  "message": "The requested status transition is not allowed.",
  "details": {},
  "retryable": false
}
```

Suggested codes:
- AUTHENTICATION_FAILED
- AUTHORIZATION_FAILED
- RESOURCE_NOT_FOUND
- CONFIGURATION_ERROR
- VALIDATION_ERROR
- CONFLICT
- RATE_LIMITED
- OPENPROJECT_UNAVAILABLE
- ATTACHMENT_UPLOAD_FAILED
- HANDOVER_STATE_ERROR
- CONFIRMATION_REQUIRED
- UNSUPPORTED_CAPABILITY

---

# 36. Repository Structure

```text
openproject-ai-bridge/
├── cmd/
│   ├── server/
│   └── opai/
├── internal/
│   ├── app/
│   ├── auth/
│   ├── config/
│   ├── openproject/
│   ├── projects/
│   ├── workpackages/
│   ├── delta/
│   ├── artifacts/
│   ├── handover/
│   ├── permissions/
│   ├── audit/
│   ├── webhook/
│   └── storage/
├── mcp/
│   ├── server/
│   └── tools/
├── api/
│   └── openapi/
├── plugin/
│   ├── plugin.json
│   ├── mcp.json
│   ├── skills/
│   │   ├── project-status/SKILL.md
│   │   ├── project-delta/SKILL.md
│   │   └── artifact-handover/SKILL.md
│   └── assets/
├── docs/
│   ├── architecture.md
│   ├── security.md
│   ├── configuration.md
│   ├── project-delta.md
│   ├── artifact-handover.md
│   ├── mcp-tools.md
│   └── contributing.md
├── examples/
│   ├── compose.yaml
│   └── config.example.yaml
├── migrations/
├── test/
│   ├── integration/
│   └── fixtures/
├── Dockerfile
├── compose.yaml
├── .env.example
├── go.mod
├── go.sum
├── LICENSE
├── CONTRIBUTING.md
├── SECURITY.md
└── README.md
```

---

# 37. ChatGPT / Codex Plugin Packaging

Public plugin support is a secondary distribution layer.

The core project MUST work without ChatGPT.

Current plugin architecture supports:
- remote MCP server
- skills
- or both

Portable plugin package can use:
- `plugin.json`
- `mcp.json`
- `skills/`
- `assets/`

Public submission must NOT be an MVP blocker.

Milestones:

```text
MVP local/self-hosted
↓
stable MCP
↓
plugin package
↓
local/private testing
↓
security review
↓
optional public submission
```

---

# 38. Skills

Potential packaged skills:

## project-status
Retrieve OpenProject state, summarize progress, identify blockers/decisions, never mutate project.

## project-delta
Transform session results into a Project Delta, preview changes, request confirmation, apply approved delta.

## artifact-handover
Locate approved HANDOVER_READY artifact, inspect metadata, guide takeover, record integration result.

---

# 39. Infra Stack Deployment

The bridge remains an independent repository.

Infra repository only consumes a released image/configuration.

Preferred:

```text
ghcr.io/<org>/openproject-ai-bridge:<version>
```

Example:

```yaml
services:
  openproject-ai-bridge:
    image: ghcr.io/<org>/openproject-ai-bridge:0.1.0
    restart: unless-stopped
    environment:
      OPAI_CONFIG: /config/config.yaml
      OPAI_DB: /data/opai.db
      OPENPROJECT_INFRA_TOKEN: ${OPENPROJECT_INFRA_TOKEN}
    volumes:
      - ./config/openproject-ai-bridge:/config:ro
      - openproject_ai_bridge_data:/data
    networks:
      - proxy
      - openproject
```

Expose through Traefik only if remote MCP/REST access is required.

---

# 40. Container Image

Requirements:
- multi-stage build
- non-root runtime user
- minimal base image
- read-only root filesystem where practical
- healthcheck
- SBOM in CI if possible
- signed releases/container images later

---

# 41. CI/CD

GitHub Actions should run:
- format
- lint
- unit tests
- integration tests
- build
- container build
- vulnerability scan

On tagged release:
- build binaries
- build OCI image
- publish GHCR image
- create GitHub release
- attach checksums

---

# 42. Testing Strategy

## Unit
- delta validation
- permission policy
- mapping
- handover state machine
- idempotency
- error translation

## Integration
Test:
- auth
- list project
- create work package
- update work package
- add comment
- attach Markdown
- retrieve attachment
- webhook verification
- conflict handling

## MCP
At minimum:
- tool schemas
- read/write annotations
- positive use cases
- negative/safety cases

---

# 43. Security Test Cases

1. read-only actor attempts update -> denied
2. wrong project scope -> denied
3. delete request -> unsupported in MVP
4. malicious attachment filename -> sanitized/rejected
5. oversized attachment -> rejected/configured limit
6. webhook invalid signature -> rejected
7. duplicate idempotency key -> no duplicate action
8. stale work package update -> conflict
9. secret supplied in Project Delta -> redact/reject
10. path traversal in target_path -> reject

---

# 44. Attachment Safety

At minimum:
- sanitize filenames
- configurable max size
- content type inspection
- prevent path traversal
- never execute uploaded artifacts
- Markdown is treated as data
- follow OpenProject attachment restrictions
- support OpenProject-side virus scanning if configured

---

# 45. Markdown Handover Front Matter

```yaml
---
title: "OpenProject AI Bridge – Architecture"
artifact_type: "concept"
project: "openproject-ai-bridge"
source: "chatgpt"
status: "handover-ready"
created_at: "2026-09-29"
target_repository: "openproject-ai-bridge"
target_path: "docs/architecture.md"
related_work_package: "OPAI-012"
---
```

---

# 46. Source Control Strategy

Recommended:
- protected `main`
- feature branches / PRs
- conventional commits optional

Examples:

```text
feat(mcp): add work package lookup
feat(delta): add preview/apply flow
feat(artifact): support markdown attachments
fix(auth): refresh oauth token
docs(handover): define Cody takeover workflow
```

---

# 47. Versioning

Use Semantic Versioning.

```text
0.1.0  MVP foundation
0.2.0  Project Delta
0.3.0  Artifact handover
0.4.0  MCP hardening
0.5.0  OAuth/multi-user
1.0.0  stable public API/tool contracts
```

---

# 48. Suggested Development Milestones

## M0 — Repository bootstrap
Repository, license, README, Go project, Dockerfile, CI, config model, health endpoint.

## M1 — OpenProject client
API token auth, instance health, projects/workspaces, work package read/create/update/comment, tests.

## M2 — Project Delta
Schema, validation, preview, apply, idempotency, audit.

## M3 — Artifacts
Upload attachment, list/download, front matter parser, artifact states.

## M4 — Handover
Handover model, HANDOVER_READY, target repo/path, integration result, comment, CLI.

## M5 — MCP
MCP transport, semantic tools, annotations, read/write policies, Cody/ChatGPT testing.

## M6 — Webhooks
Signature verification, selected events, cache invalidation, stale-state protection.

## M7 — Public plugin package
Plugin manifest, skills, remote MCP deployment pattern, privacy/security docs, test cases, optional submission.

---

# 49. Suggested First Use Case

Dogfood the bridge on itself.

OpenProject project:

```text
OpenProject AI Bridge
```

Initial work packages:

```text
OPAI-001 Repository bootstrap
OPAI-002 Architecture
OPAI-003 OpenProject API client
OPAI-004 Authentication
OPAI-005 Work package service
OPAI-006 Project Delta schema
OPAI-007 Approval / permission model
OPAI-008 Audit log
OPAI-009 Attachment service
OPAI-010 Artifact metadata
OPAI-011 Handover workflow
OPAI-012 MCP server
OPAI-013 CLI
OPAI-014 REST API
OPAI-015 Webhook receiver
OPAI-016 Docker image
OPAI-017 Infra integration example
OPAI-018 Security hardening
OPAI-019 Plugin package
OPAI-020 Public documentation
```

---

# 50. Open Questions

The following questions remain, but **none blocks repository bootstrap**.

## Q1 — Final language stack for MCP
Preferred core: Go.

Decide during M5:
- MCP also in Go
- or thin TypeScript MCP adapter around Go service

## Q2 — Persistent state
Recommendation: **SQLite first.**

## Q3 — Authentication beyond private deployment
MVP: API token.
Public/general deployment: OAuth 2.0 strongly preferred.

## Q4 — Remote MCP public hosting
Keep private initially. Public ChatGPT/Codex plugin needs stable HTTPS endpoint and hardened auth/security.

## Q5 — Exact OpenProject versions supported
Document minimum/latest tested versions after integration testing.

## Q6 — Custom fields
Do not make mandatory in MVP. Add optional mapping after testing current instance/API support.

## Q7 — Repository provider integration
No direct GitHub write capability in bridge MVP.

## Q8 — Naming
Current recommendation:

```text
Product: OpenProject AI Bridge
Repo: openproject-ai-bridge
CLI: opai
```

Before public branding, keep “unofficial community project” prominent.

---

# 51. Decisions Already Made

- standalone GitHub repository
- open source
- independent from Infra repository
- consumed by Infra as container/service
- OpenProject Community Edition is a primary target
- API v3 is backend integration
- MCP is first-class AI interface
- REST and CLI are also interfaces
- vendor neutrality
- ChatGPT and Cody should both be usable clients
- human-controlled writes by default
- agent-specific OpenProject identities
- no shared admin token
- no raw conversation synchronization
- Project Delta is core capability
- attachments are first-class
- Markdown artifact handover is first-class
- Cody takeover workflow is first-class
- Git remains source of truth for integrated technical docs
- OpenProject remains source of truth for project state
- no cloning/reimplementation of proprietary OpenProject code
- public plugin submission is optional/later
- clearly labeled unofficial/community

---

# 52. Definition of MVP Done

MVP is complete when:

1. bridge runs as Docker container
2. bridge connects to self-hosted OpenProject using API token
3. configured AI/client identity can list projects
4. client can list/read work packages
5. client can create/update work package
6. client can add comment
7. Project Delta can be validated and previewed
8. approved Project Delta can be applied idempotently
9. Markdown file can be attached
10. attachment can be downloaded
11. artifact can be marked HANDOVER_READY
12. Cody/CLI can read handover metadata
13. integration result can record repo/path/commit
14. all writes produce audit records
15. default permission mode protects writes
16. secrets stay out of repo/log/tool responses
17. unit/integration tests cover core flows
18. Infra consumes released Docker image without copying bridge source

---

# 53. First Implementation Instruction for Cody

Start with repository/bootstrap and architecture, not MCP.

Recommended order:

```text
1. initialize repository
2. define config model
3. define domain interfaces
4. implement OpenProject API client
5. integration-test work packages
6. implement permission/audit layer
7. implement Project Delta
8. implement attachments
9. implement handover
10. add CLI
11. add REST
12. add MCP
13. add webhook synchronization
14. package ChatGPT/Codex plugin
```

Reason:

> MCP is an adapter. The business logic must be stable without it.

---

# 54. Acceptance Philosophy

Prefer a small, boring, auditable core over a clever autonomous agent.

Good:

```text
explicit tool
explicit target
explicit delta
explicit preview
explicit approval
explicit result
```

Avoid:

```text
AI decides everything
generic API proxy
implicit writes
hidden state changes
admin credentials
automatic destructive cleanup
```

The bridge should make AI **useful and visible**, not powerful and mysterious.

---

# 55. Final Product Vision

```text
Human
   ↓
works with AI
   ↓
AI produces meaningful result
   ↓
Project Delta
   ↓
human approves
   ↓
OpenProject reflects durable project state
   ↓
artifact becomes HANDOVER_READY
   ↓
implementation agent takes over
   ↓
repo/document system becomes final source of truth
   ↓
OpenProject records completion and provenance
```

> **Creative conversations remain free. Durable outcomes become transparent, auditable project reality.**

---

# 56. Reference Notes for Implementers

Before coding against OpenProject, verify against the running instance's current API documentation.

Relevant official documentation areas:
- OpenProject API v3 introduction
- Work Packages API
- Attachments API
- API and Webhooks administration
- OAuth applications
- `/api/v3/spec.json`
- `/api/v3/spec.yml`

For ChatGPT/Codex plugin packaging, verify against current OpenAI plugin documentation:
- Plugins overview
- Build an MCP server
- Build skills
- Package your plugin
- Submit plugins

Treat the live API specification and stable vendor documentation as source of truth.

---

# 57. Final Note to Cody

This document is the **initial handover contract**, not an immutable implementation prison.

If repository reality exposes a simpler, safer or more maintainable implementation:

1. preserve the architectural principles,
2. document the deviation,
3. explain the tradeoff,
4. update this specification or successor ADR,
5. keep external tool contracts stable where possible.

Most important constraints:

> **Open source. Independent. Safe by default. Self-host friendly. Vendor-neutral. Human-controlled. Auditable. Useful.**

Build the smallest version that proves this end-to-end.
