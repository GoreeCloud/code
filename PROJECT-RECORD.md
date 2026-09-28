# GoreeCloud Code — Project Record

**Document Type:** Repository-Native Project Record  
**Status:** Active  
**Project:** GoreeCloud Code  
**Repository:** GoreeCloud/code  
**Authority:** Repository-local project record  
**Last Updated:** 2026-09-27

## 2026-09-27 — Project specification migration staged

GoreeCloud Code's project specification and significant project history are being migrated from transitional Drive sources into PROJECT-SPECIFICATIONS.md and PROJECT-RECORD.md.

The migration reconciles the historical repository name GoreeCloud/goreecloud-code to the live GoreeCloud/code repository.

Current GitHub main is authoritative for accepted implementation state. The Drive sources remain requirement/history inputs and do not override newer repository evidence.

Both Drive project-specification sources remain protected until this migration is reviewed, accepted, merged, read back from main, and reconciled without remaining discrepancies.

## 2026-09-27 — Feature-state migration accepted

Pull request #5, "Migrate Code feature tracking from Drive," was merged to main as 9c199f58b2585d404d540606cb1097204d4cd6ef.

Its exact head was 4e0af9c0496d3d13c9b376d1cff98e5263b64263.

The migration established repository-native IMPLEMENTED-FEATURES.md, PLANNED-FEATURES.md, and CHANGELOGS.md.

Milestone 0 remained a completed source foundation and Milestone 1 remained in progress. No lifecycle promotion was implied.

## Current accepted main boundary

At the start of the project-specification migration, authoritative main was 9c199f58b2585d404d540606cb1097204d4cd6ef.

Accepted main records:

- M0 source foundation complete;
- M1 Forgejo connectivity in progress;
- provider-neutral repository/API contracts;
- Forgejo adapter foundation;
- GoreeCloud-owned API and web shell;
- provider health/repository/branch/commit/issue/pull-request reads;
- Glaze-oriented repository dashboard;
- Forgejo validation tooling.

A real-instance validation run remains required before M1 completion.

Accepted main does not establish production deployment, Stable status, live governed writes, complete platform-system integration, or migration of repository authority away from GitHub.

## 2026-09-27 — PR #2 visibility/CI candidate

PR #2, "Document repository visibility model and restore CI," remains open and unmerged.

Exact head: 95838d462110ec2fa244ef6bef7ee5e4204bc9a3.

GitHub Actions run 36348974548 passed setup, pnpm install, pnpm check, pnpm test, and pnpm build on that exact candidate.

The PR documents the intended separation between service visibility and per-repository visibility and repairs pre-existing CI defects.

Because it remains unmerged, those changes are candidate state rather than accepted main.

Its documentation also identified the stale GoreeCloud/goreecloud-code repository name and Drive-roadmap references; this project-record migration resolves the project-specification authority and repository-name portion on its own candidate branch without treating PR #2 as merged.

## 2026-09-07 — PR #1 governed-write candidate

Draft PR #1, "Start M2 governed branch creation," remains open and unmerged.

Current exact head at the migration audit: 7521a44afffac6ea674d02750c43b55e66d3288c.

CI run 34160056343 passed workspace TypeScript checks, the full test suite, and build on that exact head.

The candidate develops:

- bounded provider-neutral branch creation;
- separate provider permission and GoreeCloud Code application authorization;
- mandatory pre-provider audit evidence;
- durable idempotency;
- data-minimized operation context;
- read-only governed-write status;
- observation-only reconciliation assessment;
- fail-closed ambiguity handling;
- exact source-revision comparison evidence;
- evidence binding/provenance structures.

The candidate deliberately keeps reconciliation assessment non-authorizing. It does not establish automatic repair, provider write retry, journal mutation, production identity, Wardveil/Privacy Shield/Everkeep/Mesh acceptance, target-environment validation, RC, Stable, or production authority.

## Product direction preserved from Drive

The Drive project specification established the enduring product direction that GoreeCloud Code is the GoreeCloud-owned developer/source-control platform while Forgejo remains replaceable infrastructure.

The specification also established:

- provider-neutral architecture;
- governed writes with separate application authorization;
- audit/evidence requirements;
- durable idempotency and reconciliation;
- Glaze UI developer experience;
- Wardveil Security;
- Privacy Shield;
- Everkeep recovery/portability;
- GoreeCloud Identity;
- GoreeCloud Mesh;
- validation and production gates;
- explicit deferral of broad write/admin capabilities until their safety and acceptance requirements are met.

These requirements are preserved in PROJECT-SPECIFICATIONS.md.

## Authority transition

After the migration is accepted and verified on main:

- PROJECT-SPECIFICATIONS.md is the canonical Code project specification;
- PROJECT-RECORD.md is the canonical significant project history/evidence record;
- IMPLEMENTED-FEATURES.md and PLANNED-FEATURES.md own feature lifecycle state;
- CHANGELOGS.md owns release/repository change history;
- Google Drive no longer remains a parallel authoritative project specification.

The Drive sources become permanently deletable only after accepted default-branch readback and final reconciliation.
