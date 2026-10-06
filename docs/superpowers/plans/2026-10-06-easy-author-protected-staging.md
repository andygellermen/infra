# EasyAuthor Protected Staging Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a locally verified, repeatable and fully Basic-Auth-protected staging deployment for EasyAuthor at `author.geller.men`, including production containers, Ansible, helper scripts, backup/restore and authenticated smoke checks.

**Architecture:** A production nginx web container serves the React bundle and proxies `/api/` over a private Docker network to the Go API. Only the web container joins Traefik; an EasyAuthor-specific Ansible role owns its data directory, containers, networks and labels, while strict helper scripts create ignored secrets, build immutable images, back up SQLite/library data and verify the complete protected route.

**Tech Stack:** Bash, Docker multi-stage builds, nginx, Go 1.23, Node 20/Vite, Ansible `community.docker`, Traefik, SQLite, curl, bcrypt via `htpasswd`.

**Spec:** `docs/superpowers/specs/2026-10-06-easy-author-protected-staging-design.md`

## Global Constraints

- The only staging domain is `author.geller.men`; no public production release is implied.
- Basic Auth protects UI, API and health routes, and Traefik removes the forwarded `Authorization` header.
- The API publishes no host port and joins only the instance-private network.
- The web container runs a production Vite build under nginx, never the Vite development server.
- Server data lives under `/srv/easy-author/<domain>/data`; local demo or book data is never uploaded implicitly.
- Plaintext passwords, manuscripts and complete hostvars never appear in Git or command output.
- Every redeploy with existing data requires a verified pre-deploy backup; a failed backup stops deployment.
- Restore never overwrites current data in place and requires explicit confirmation.
- EasyReader, AddToBook, PostgreSQL and public account management remain out of scope.

## Review Focus

- A bcrypt hash contains `$` characters: hostvars generation and Ansible labels must preserve it verbatim without shell or Compose interpolation.
- A failed or interrupted backup must leave no archive that a restore command can mistake for complete.
- A restore archive with traversal paths or the wrong instance must be rejected before extraction touches the live data directory.
- A redeploy against a first-time empty instance must not require a nonexistent backup, while every later redeploy must require one.
- API restart persistence checks must compare stable project identity, not merely accept any syntactically valid response.

---

## Planned File Structure

- `apps/easy-author/frontend/Dockerfile`: production build and nginx runtime.
- `apps/easy-author/frontend/nginx.conf`: SPA fallback and private `/api/` proxy.
- `apps/easy-author/docker-compose.yml`: retain the explicitly named local development target.
- `apps/easy-author/backend/Dockerfile`: existing static Go runtime plus healthcheck-compatible image metadata.
- `ansible/playbooks/deploy-easy-author.yml`: localhost deployment entry point.
- `ansible/playbooks/roles/easy-author/tasks/main.yml`: validate configuration and deploy networks, directories and containers.
- `ansible/hostvars/templates/easy-author-hostvars.j2`: secret-free operator template.
- `scripts/easy-author-add.sh`: create ignored mode-0600 hostvars.
- `scripts/easy-author-redeploy.sh`: build immutable images, require backup and invoke Ansible/smoke.
- `scripts/easy-author-smoke-check.sh`: protected external and restart-persistence checks.
- `scripts/easy-author-backup.sh`: consistent, atomic instance backup.
- `scripts/easy-author-restore.sh`: validated staged restore with safety backup.
- `scripts/easy-author-rotate-secrets.sh`: atomic Basic-Auth hash rotation.
- `scripts/verify-easy-author-deployment.sh`: dependency-free static and mocked behavioral regression suite.
- `apps/easy-author/docs/PROTECTED-STAGING.md`: setup, deployment, operation and recovery runbook.

### Task 1: Production Web and API Container Contract

**Files:**
- Modify: `apps/easy-author/frontend/Dockerfile`
- Create: `apps/easy-author/frontend/nginx.conf`
- Modify: `apps/easy-author/backend/Dockerfile`
- Modify: `apps/easy-author/docker-compose.yml`
- Create: `scripts/verify-easy-author-deployment.sh`

**Interfaces:**
- Produces: named frontend targets `development` and `production`; images `easy-author-web:<git-sha>` on port `8080` and `easy-author-api:<git-sha>` on port `8086`; nginx routes `/api/` to `http://easy-author-api:8086` and all other unknown paths to `/index.html`.
- Consumes: existing frontend build command and Go server environment variables.

- [ ] **Step 1: Create a failing deployment verifier section** that asserts the frontend Dockerfile has named `development`, Node `build` and nginx `production` stages, that the final production stage contains no `npm run dev`, copies `dist`, runs as an unprivileged nginx user, and installs `nginx.conf`; assert nginx has SPA fallback, security headers and the exact private API upstream.
- [ ] **Step 2: Add failing assertions** that local Compose explicitly selects the `development` target and retains ports 5173/8086, while the backend image exposes only `8086`, contains no host data and adds no package solely for health probing.
- [ ] **Step 3: Run `bash scripts/verify-easy-author-deployment.sh containers`**; expect failure on the current Vite development image and absent nginx configuration.
- [ ] **Step 4: Implement the named development/build/production frontend stages and nginx proxy contract**; use nginx unprivileged on `8080`, cache hashed assets, disable caching for `index.html`, and add `X-Content-Type-Options`, `Referrer-Policy`, and a staging-safe CSP compatible with the current app.
- [ ] **Step 5: Point local Compose explicitly at the development target and keep API readiness in Ansible via an internal network probe, preserving the package-minimal backend runtime.**
- [ ] **Step 6: Run the verifier, frontend tests/build and backend tests**; expect all pass.
- [ ] **Step 7: Build both Docker images locally and inspect exposed ports and commands**; if the Docker daemon is unavailable, record the environment limit but retain static verification.
- [ ] **Step 8: Commit** with `git commit -m "build: add production EasyAuthor containers"`.

### Task 2: Safe Hostvars Creation and Secret Rotation

**Files:**
- Create: `scripts/easy-author-add.sh`
- Create: `scripts/easy-author-rotate-secrets.sh`
- Create: `ansible/hostvars/templates/easy-author-hostvars.j2`
- Modify: `scripts/verify-easy-author-deployment.sh`

**Interfaces:**
- Produces: ignored `ansible/hostvars/<domain>.yml` containing `easy_author_enabled`, image repository names, bcrypt username/hash, realm, paths, retention and smoke settings; rotation replaces only the auth hash atomically.
- Consumes: `verify_domain_resolves_to_host_ipv4(domain, host_ip)` from `scripts/lib/dns-check.sh`.

- [ ] **Step 1: Add failing sandboxed verifier cases** for invalid domains/usernames, duplicate hostvars, missing tools, mode `0600`, DNS-check invocation, absence of plaintext password, preserved bcrypt `$`, and cleanup after interrupted generation.
- [ ] **Step 2: Run `bash scripts/verify-easy-author-deployment.sh hostvars`**; expect failure because both scripts are absent.
- [ ] **Step 3: Implement `easy-author-add.sh <domain> [--username=<name>] [--skip-dns-check]`** with hidden double-entry through `htpasswd`, `umask 077`, temporary-file-plus-rename, fixed staging defaults and no secret output.
- [ ] **Step 4: Implement `easy-author-rotate-secrets.sh <domain> [--username=<name>]`** to validate an existing EasyAuthor hostvars file, preserve unrelated values, atomically replace username/hash, and restore the old file on failure.
- [ ] **Step 5: Add the commented secret-free hostvars template** with `author.geller.men` examples but no valid hash or password.
- [ ] **Step 6: Run hostvars verifier cases and `bash -n` on both scripts**; expect all pass.
- [ ] **Step 7: Commit** with `git commit -m "feat: add EasyAuthor staging secret helpers"`.

### Task 3: Atomic Backup and Explicit Restore

**Files:**
- Create: `scripts/easy-author-backup.sh`
- Create: `scripts/easy-author-restore.sh`
- Modify: `scripts/verify-easy-author-deployment.sh`

**Interfaces:**
- Produces: `easy-author-backup.sh <domain> [--retention=<n>] [--allow-empty] -> absolute archive path`; `easy-author-restore.sh <domain> <archive> --confirm=<domain>` creates a safety backup and atomically swaps validated data.
- Consumes: instance root `/srv/easy-author/<domain>`, container names `easy-author-<domain-id>-api` and `easy-author-<domain-id>-web`.

- [ ] **Step 1: Add failing mocked-Docker verifier cases** for first deploy with no data, existing-data API stop/start, timestamped temporary archive, archive readability, retention preserving newest success, and cleanup/restart on tar failure.
- [ ] **Step 2: Add failing restore cases** for missing confirmation, mismatched domain, archive outside the expected naming contract, absolute or `..` members, missing required data root, extraction failure and safety-backup failure.
- [ ] **Step 3: Run `bash scripts/verify-easy-author-deployment.sh backup` and `... restore`**; expect failure because scripts are absent.
- [ ] **Step 4: Implement backup** with exact path validation, a trap that restarts an API it stopped, archive creation to `.partial`, `tar -tf` verification, atomic rename, mode `0600`, and post-success retention.
- [ ] **Step 5: Implement restore** with exact confirmation, archive-member validation before extraction, an automatic safety backup, extraction into a sibling temporary directory, required SQLite/library layout validation, stopped containers during atomic directory rename, and rollback on failure.
- [ ] **Step 6: Run verifier cases and `bash -n`**; expect all pass.
- [ ] **Step 7: Commit** with `git commit -m "feat: add EasyAuthor backup and restore helpers"`.

### Task 4: Idempotent Ansible Deployment and Redeploy Orchestration

**Files:**
- Create: `ansible/playbooks/deploy-easy-author.yml`
- Create: `ansible/playbooks/roles/easy-author/tasks/main.yml`
- Create: `scripts/easy-author-redeploy.sh`
- Modify: `scripts/verify-easy-author-deployment.sh`

**Interfaces:**
- Produces: `easy-author-redeploy.sh <domain> [--check-only] [--build-only] [--skip-auth-smoke]`; deploy role accepts `target_domain`, all `easy_author_*` hostvars, and required per-run `easy_author_deploy_api_image`/`easy_author_deploy_web_image` values tagged with the full Git SHA.
- Consumes: Task 1 image tags, Task 2 hostvars and Task 3 backup command.

- [ ] **Step 1: Add failing static Ansible assertions** for exact domain selection, `no_log`, Basic-Auth validation, bcrypt preservation, private and Traefik networks, directory modes, no API published port, web-only Traefik labels, HTTPS redirect, cert resolver and auth middleware on all paths.
- [ ] **Step 2: Add failing redeploy mock cases** for dirty tree rejection, immutable SHA tags, build failure causing no backup/deploy, first deploy allowing empty data, existing deploy requiring successful backup, check/build-only behavior and smoke invocation.
- [ ] **Step 3: Run `bash scripts/verify-easy-author-deployment.sh ansible` and `... redeploy`**; expect failure because role, playbook and helper are absent.
- [ ] **Step 4: Implement the role** to collect only enabled EasyAuthor hostvars, require and validate the two full-SHA `easy_author_deploy_*_image` values passed by the redeploy helper, create networks/directories, deploy API with internal readiness, then deploy nginx with complete Traefik labels.
- [ ] **Step 5: Implement the playbook** for localhost, privilege escalation and only the EasyAuthor role.
- [ ] **Step 6: Implement redeploy orchestration** with clean tracked tree requirement, full commit SHA tags, both image builds before backup, conditional first-deploy handling, mandatory backup thereafter, Ansible call and authenticated smoke.
- [ ] **Step 7: Run verifier cases, `bash -n`, `ansible-playbook --syntax-check`, and `--check` using a temporary non-secret fixture**; expect no mutation and exit 0.
- [ ] **Step 8: Commit** with `git commit -m "feat: add protected EasyAuthor Ansible deployment"`.

### Task 5: Authenticated Smoke, Persistence Check and Operations Runbook

**Files:**
- Create: `scripts/easy-author-smoke-check.sh`
- Create: `apps/easy-author/docs/PROTECTED-STAGING.md`
- Modify: `apps/easy-author/README.md`
- Modify: `scripts/verify-easy-author-deployment.sh`

**Interfaces:**
- Produces: `easy-author-smoke-check.sh <domain> --username=<name> [--password-stdin] [--skip-restart]`; documented operator flow add → deploy → smoke → backup/restore → rotate.
- Consumes: deployed Basic Auth, `/api/health`, `/api/projects`, `/api/kanban`, exact API container name and Docker access for persistence verification.

- [ ] **Step 1: Add failing smoke verifier cases** for HTTP redirect, unauthenticated `401`, authenticated UI `200`, health JSON, projects JSON, five Kanban phases, no API published ports, captured stable project ID, API restart, same ID after restart, hidden password input and secret cleanup.
- [ ] **Step 2: Run `bash scripts/verify-easy-author-deployment.sh smoke`**; expect failure because the smoke helper is absent.
- [ ] **Step 3: Implement the smoke helper** with curl status/body separation, JSON validation using Python 3 or a validated available parser, exact container inspection and bounded readiness retries after restart.
- [ ] **Step 4: Write the protected staging runbook** covering DNS, dependencies, hostvars, deployment, review URL, logs, backup listing, explicit restore, rotation, rollback, known first-deploy behavior and prohibition on committing secrets/local data.
- [ ] **Step 5: Link the runbook from the EasyAuthor README** without changing the local development workflow.
- [ ] **Step 6: Run all verifier groups, all shell syntax checks, Ansible syntax/check, backend/frontend tests, production build and `docker compose config --quiet`**; expect all pass.
- [ ] **Step 7: If Docker is available, run isolated production containers on a temporary network and prove UI→nginx→API plus persistence restart; otherwise record the exact daemon limitation in the runbook validation section.**
- [ ] **Step 8: Commit** with `git commit -m "docs: complete EasyAuthor staging operations"`.

### Task 6: Final Deployment Readiness Audit

**Files:**
- Modify: `apps/easy-author/docs/PROTECTED-STAGING.md`

**Interfaces:**
- Consumes: all prior task interfaces.
- Produces: a concise readiness record for local preparation; no real DNS, hostvars or server mutation.

- [ ] **Step 1: Run the complete dependency-free deployment verifier** and record group counts and result.
- [ ] **Step 2: Run `shellcheck` on new scripts when available; otherwise record that exact environment boundary and retain `bash -n` plus behavioral tests.**
- [ ] **Step 3: Run Ansible syntax/check mode, both application test suites, both production image builds when Docker is available, and config validation.**
- [ ] **Step 4: Confirm with `git grep` that no bcrypt hash fixture usable for login, plaintext password, generated hostvars, SQLite file, library content or backup archive is tracked.**
- [ ] **Step 5: Confirm the API image/container has no published host port and every HTTPS route uses Basic Auth in the rendered Ansible model.**
- [ ] **Step 6: Add the dated command/result matrix and remaining environment limits to the runbook.**
- [ ] **Step 7: Run `git diff --check` and inspect `git status --short`; only intended readiness documentation may remain.**
- [ ] **Step 8: Commit** with `git commit -m "test: validate EasyAuthor protected staging"`.
