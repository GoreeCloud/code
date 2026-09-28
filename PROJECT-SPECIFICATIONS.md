# GoreeCloud Code — Project Specifications

**Document Type:** Repository-Native Project Specification  
**Status:** Active specification / Development  
**Project:** GoreeCloud Code  
**Repository:** GoreeCloud/code  
**Authority:** Repository-local project specification  
**Last Updated:** 2026-09-27

## 1. Purpose

GoreeCloud Code is GoreeCloud's first-party developer and source-control platform for repositories, collaboration, CI/CD, packages, security, governed automation, AI-assisted development, and GoreeCloud platform integrations.

Standard Git is the interoperability foundation.

Forgejo is the preferred initial, replaceable infrastructure foundation. It is not the permanent GoreeCloud product boundary. GoreeCloud Code owns the user experience, provider-neutral service contracts, governance model, authorization model, evidence model, integration architecture, migration process, and long-term developer-platform behavior.

## 2. Current accepted implementation boundary

The authoritative default branch is Development.

Current accepted main includes the Milestone 0 source foundation and an in-progress Milestone 1 read-oriented Forgejo connectivity foundation.

Accepted current capabilities include, at source level:

- provider-neutral repository contracts;
- a replaceable ForgeProvider boundary;
- Forgejo adapter foundation;
- GoreeCloud-owned API boundary;
- Glaze-oriented web application shell and repository dashboard;
- provider health/version reporting;
- repository discovery and detail retrieval;
- branch and commit reads;
- issue and pull-request reads;
- CI/build/test scaffolding;
- Forgejo validation tooling;
- repository-native implemented/planned/changelog feature records.

A real-instance Forgejo connectivity validation remains required before Milestone 1 can be considered complete.

Current accepted main does not establish production deployment, Stable status, completed platform-system integration, live governed writes, authoritative migration away from GitHub, or target-environment acceptance.

## 3. Candidate-state boundary

Two material pull requests are open and must remain distinct from accepted main.

### PR #1 — M2 governed branch creation

Draft PR #1 is an unmerged Development candidate at exact head 7521a44afffac6ea674d02750c43b55e66d3288c.

It develops the first bounded provider write and related audit/idempotency/reconciliation evidence model.

The candidate has passed source-level CI at documented checkpoints, including exact head 7521a44afffac6ea674d02750c43b55e66d3288c with CI run 34160056343.

This candidate does not establish production authorization, live Forgejo write acceptance, distributed reconciliation, platform-system acceptance, Release Candidate, Stable, or production authority.

### PR #2 — repository visibility model and CI repair

PR #2 is an unmerged candidate at exact head 95838d462110ec2fa244ef6bef7ee5e4204bc9a3.

Its exact-head GitHub Actions run 36348974548 passed install, check, test, and build.

It documents a mixed public/private repository hosting model and repairs pre-existing validation defects.

That candidate is not accepted main and must not be treated as current repository policy until merged and read back from the authoritative branch.

## 4. Architecture

The primary boundary is:

GoreeCloud clients and Glaze UI  
→ GoreeCloud Code API  
→ provider-neutral ForgeProvider contract  
→ initial ForgejoProvider / future interoperable providers

The browser must not receive Forgejo administrative credentials or depend directly on provider-specific APIs when a GoreeCloud-owned API boundary can mediate the operation.

Provider-specific details must remain isolated so Forgejo can be upgraded, replaced, or supplemented without redefining the GoreeCloud Code product.

## 5. Repository model

GoreeCloud Code must support the normal Git repository model while providing GoreeCloud-owned governance and application behavior.

Repository capabilities may include:

- discovery and metadata;
- branches and refs;
- commits and history;
- issues;
- pull requests and reviews;
- releases;
- packages/registries;
- CI/CD workflows;
- repository settings;
- access policy;
- security findings;
- migration/import/export.

Each capability must be implemented through a bounded contract and verified before being represented as available.

Repository service visibility and individual repository visibility are separate policy dimensions. Public service reachability must not automatically make all hosted repositories public.

## 6. Governed write model

Repository writes require a stricter boundary than reads.

The planned governed-write model includes:

- explicit application authorization distinct from provider permission;
- pre-provider audit/evidence creation;
- durable idempotency reservation;
- data-minimized operation context;
- safe completed replay;
- conflict detection;
- fail-closed unresolved-state handling;
- provider-neutral status inspection;
- read-only reconciliation assessment;
- authoritative mutation/reconciliation only through separately approved capabilities;
- rollback/repair expectations appropriate to the operation.

A provider token that can perform an operation is not, by itself, GoreeCloud Code authorization.

## 7. Audit and evidence foundation

Security- and governance-relevant mutations must produce durable, attributable evidence appropriate to lifecycle.

Evidence should support:

- operation identity;
- actor/service identity where available;
- target repository and operation;
- requested source/target refs;
- authorization decision;
- provider result;
- idempotency state;
- timestamps;
- reconciliation state;
- validation/acceptance evidence.

Evidence structures must fail closed on malformed, ambiguous, contradictory, stale, or improperly bound inputs.

Evidence alone must not become an automatic authorization mechanism unless a separate governed capability explicitly establishes that behavior.

## 8. Idempotency and reconciliation

Bounded write operations should support idempotency where retries can otherwise create duplicate or ambiguous provider effects.

Raw idempotency keys should not be persisted when a one-way derived representation is sufficient.

Reconciliation assessment must remain observational unless separately authorized.

For uncertain provider outcomes, reconciliation should distinguish states such as:

- operation already durably succeeded;
- legacy/insufficient operation context;
- provider target present;
- provider target absent;
- provider result ambiguous;
- provider observation unavailable.

Exact source-revision comparison may be useful evidence for manual review, but equality or inequality must not automatically authorize repair or mutation.

## 9. Authentication and authorization

GoreeCloud Identity should establish authenticated identity and relevant claims where applicable.

GoreeCloud Code remains responsible for application authorization.

Provider credentials must remain server-side, narrowly scoped, protected, and separated from browser/client code.

Development-only shared-secret or bearer mechanisms must not be represented as accepted production identity architecture.

Administrative authority, repository write authority, workflow authority, package authority, and migration authority should be separately scopeable.

## 10. Wardveil Security

Wardveil Security is responsible for applicable evidence-backed security controls and security-state presentation.

Relevant domains include:

- repository and provider trust;
- dependency and supply-chain state;
- runner/workflow security;
- artifact provenance;
- protected-branch and review policy;
- high-risk repository operations;
- deployment-security evidence.

A “secure,” “protected,” “verified,” or similar state must derive from evidence, not a static label.

## 11. Privacy Shield

Privacy requirements include data minimization, explicit telemetry behavior, retention control, and separation between repository metadata and unnecessary user/private content.

Repository contents, issue data, pull-request data, private repository inventory, secrets, logs, and operational evidence must not be exposed beyond authorized need.

Provider errors and diagnostics should be sanitized before being returned to less-trusted clients.

## 12. Everkeep and recovery

GoreeCloud Code must support recovery and portability for authoritative repository and developer-platform state.

Git history is necessary but not sufficient for complete recovery.

Migration/recovery planning must explicitly account for non-Git metadata such as:

- repository visibility;
- collaborators and teams;
- branch protections;
- review requirements;
- issues and pull requests where preservation is required;
- release/package metadata;
- CI/CD configuration and secrets;
- webhooks/integrations;
- provider settings;
- audit/evidence records;
- migration state.

Backup configuration alone is not recovery proof. Representative restoration and validation are required for acceptance claims.

## 13. GoreeCloud Mesh

Mesh may support coordination, events, service discovery, and governed communication between GoreeCloud Code and other first-party systems.

Mesh connectivity must not grant repository authorization by itself.

Event contracts must preserve authoritative source ownership and avoid leaking protected repository information.

## 14. Glaze UI

User-facing GoreeCloud Code surfaces must use the applicable current Stable Glaze UI contract at implementation and acceptance time.

The developer experience should provide:

- clear repository and branch context;
- visible current/blocked/candidate states;
- explicit confirmation for high-impact actions;
- accessible focus/navigation;
- readable code/repository information;
- clear loading, empty, partial, error, and recovery states;
- evidence-linked security and governance status.

Historical Glaze UI version references are not permanent conformance evidence.

## 15. CI/CD and execution portability

CI behavior should remain portable between GitHub Actions, Forgejo Actions, GoreeCloud-controlled runners, and future GoreeCloud Pipelines where practical.

Workflow execution must preserve:

- least-privilege tokens;
- secret separation;
- artifact provenance;
- exact-revision checkout where evidence requires it;
- isolation appropriate to workload risk;
- reproducible or reviewable dependencies;
- explicit production/deployment authorization.

Source CI success does not establish target-environment production acceptance.

## 16. Migration and interoperability

Migration from an external source-control provider must preserve Git history and exact revision identity.

Before authority moves, the process must separately inventory and recreate all required non-Git state, including visibility, access policy, branch protections, review requirements, CI/CD boundaries, secrets separation, recovery, and provider metadata.

Migration must be reversible until acceptance.

The original authoritative provider remains authoritative until migration completeness, validation, cutover, and rollback gates are satisfied.

## 17. Deferred write capabilities

Beyond governed branch creation, later writes may include:

- issue mutations;
- pull-request creation;
- review workflows;
- repository administration;
- branch deletion;
- package publishing;
- pipeline changes;
- AI-assisted code changes;
- migration/cutover actions.

Each class requires its own authorization, audit, idempotency/reconciliation, recovery, security, privacy, and acceptance design.

No future write capability becomes implemented merely because it appears in this specification.

## 18. Validation

Validation should cover, as applicable:

- provider contract behavior;
- API behavior;
- authorization boundaries;
- audit/evidence creation;
- idempotency;
- retry/replay;
- reconciliation assessment;
- malformed/ambiguous evidence;
- provider failures;
- private/public repository boundaries;
- secrets protection;
- exact revision evidence;
- CI/build/test operation;
- target Forgejo interoperability;
- migration/restore procedures;
- applicable platform-system integration.

A live representative target-environment validation is required for claims that depend on real provider behavior.

## 19. Stable and production-readiness boundary

Stable or production-ready status requires evidence appropriate to the claimed scope, including:

- accepted provider connectivity and required write operations;
- authenticated identity and application authorization;
- accepted Wardveil Security, Privacy Shield, Everkeep, Mesh, and applicable Glaze UI integration;
- secrets and least-privilege controls;
- protected/private repository behavior;
- recovery/restore evidence;
- monitoring and diagnostics;
- migration/cutover/rollback evidence where authority changes;
- accepted CI/CD and artifact provenance behavior;
- representative end-to-end target-environment tests;
- reconciliation of documentation, feature state, project records, and provider state.

Unmerged PR evidence remains candidate evidence only.

## 20. Roadmap

### M0 — Bootstrap

Accepted source foundation: product boundary, provider contracts, Forgejo adapter, web shell, API boundary, CI, and architecture decisions.

### M1 — Forgejo connectivity

In progress on accepted main. Real-instance validation remains open.

### M2 — Governed write operations

Candidate work exists in draft PR #1. No governed-write production acceptance is established.

### M3 — Pipelines, packages, and migration

Planned: portable workflow execution, GoreeCloud-controlled runners, package/OCI registry integration, provider migration tooling, and external mirroring where approved.

## 21. Maintenance

PROJECT-SPECIFICATIONS.md is the canonical project specification.

PROJECT-RECORD.md preserves significant history and evidence.

IMPLEMENTED-FEATURES.md records verified implemented-feature state.

PLANNED-FEATURES.md records planned feature state.

CHANGELOGS.md records release/repository change history.

Google Drive project-specification copies are migration sources only. After this migration is accepted and read back from main, they must be permanently removed.
